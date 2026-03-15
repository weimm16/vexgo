// backend/handler/sso.go
package handler

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"vexgo/backend/config"
	"vexgo/backend/model"

	"github.com/coreos/go-oidc"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/github"
	"golang.org/x/oauth2/google"
	"gorm.io/gorm"
)

// ─────────────────────────────────────────────
// State cache (CSRF protection, no cookie needed)
// ─────────────────────────────────────────────

const (
	stateLength = 16
	stateExpire = 5 * time.Minute
)

type stateEntry struct {
	ip      string
	method  string
	expires time.Time
}

var (
	stateMu    sync.Mutex
	stateCache = make(map[string]stateEntry) // key: "provider_state"
)

func generateState(provider, ip, method string) string {
	b := make([]byte, stateLength)
	rand.Read(b)
	state := base64.URLEncoding.EncodeToString(b)[:stateLength]
	key := provider + "_" + state
	stateMu.Lock()
	stateCache[key] = stateEntry{ip: ip, method: method, expires: time.Now().Add(stateExpire)}
	stateMu.Unlock()
	return state
}

func verifyState(provider, ip, state string) (method string, ok bool) {
	key := provider + "_" + state
	stateMu.Lock()
	entry, exists := stateCache[key]
	if exists {
		delete(stateCache, key) // one-time use
	}
	stateMu.Unlock()
	if !exists || entry.ip != ip || time.Now().After(entry.expires) {
		return "", false
	}
	return entry.method, true
}

// LocalLoginGuard returns a middleware that rejects password-based login
// when ALLOW_LOCAL_LOGIN=false.
func LocalLoginGuard() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !config.SSOConfig.AllowLocalLogin {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error": "local login is disabled, please use SSO",
			})
			return
		}
		c.Next()
	}
}

// ─────────────────────────────────────────────
// OAuth2 configs (built per-request to allow dynamic redirect URI)
// ─────────────────────────────────────────────

func githubOAuth2Config(redirectURI string) *oauth2.Config {
	return &oauth2.Config{
		ClientID:     config.SSOConfig.GitHub.ClientID,
		ClientSecret: config.SSOConfig.GitHub.ClientSecret,
		Endpoint:     github.Endpoint,
		Scopes:       []string{"read:user", "user:email"},
		RedirectURL:  redirectURI,
	}
}

func googleOAuth2Config(redirectURI string) *oauth2.Config {
	return &oauth2.Config{
		ClientID:     config.SSOConfig.Google.ClientID,
		ClientSecret: config.SSOConfig.Google.ClientSecret,
		Endpoint:     google.Endpoint,
		Scopes:       []string{"openid", "profile", "email"},
		RedirectURL:  redirectURI,
	}
}

// oidcOAuth2Config builds the oauth2.Config for OIDC.
// When IssuerURL is set it uses OIDC discovery to obtain endpoints automatically;
// AuthURL / TokenURL are used as a manual fallback for non-standard providers.
func oidcOAuth2Config(ctx context.Context, redirectURI string) (*oauth2.Config, error) {
	cfg := config.SSOConfig.OIDC

	var endpoint oauth2.Endpoint
	if cfg.IssuerURL != "" {
		provider, err := oidc.NewProvider(ctx, cfg.IssuerURL)
		if err != nil {
			return nil, fmt.Errorf("OIDC discovery failed for %s: %w", cfg.IssuerURL, err)
		}
		endpoint = provider.Endpoint()
	} else if cfg.AuthURL != "" && cfg.TokenURL != "" {
		endpoint = oauth2.Endpoint{
			AuthURL:  cfg.AuthURL,
			TokenURL: cfg.TokenURL,
		}
	} else {
		return nil, errors.New("OIDC: set either OIDC_ISSUER_URL (recommended) or both OIDC_AUTH_URL and OIDC_TOKEN_URL")
	}

	return &oauth2.Config{
		ClientID:     cfg.ClientID,
		ClientSecret: cfg.ClientSecret,
		Endpoint:     endpoint,
		Scopes:       cfg.Scopes,
		RedirectURL:  redirectURI,
	}, nil
}

// callbackURI builds the absolute callback URL for the given provider.
// When BASE_URL is set (e.g. https://vexgo.yzlab.de), it takes priority over
// auto-detection from the request host. This is needed when running behind a
// reverse proxy or when the public domain differs from the listen address.
func callbackURI(c *gin.Context, provider string) string {
	if base := config.SSOConfig.BaseURL; base != "" {
		return fmt.Sprintf("%s/api/sso/%s/callback", strings.TrimRight(base, "/"), provider)
	}
	scheme := "http"
	if c.Request.TLS != nil || c.GetHeader("X-Forwarded-Proto") == "https" {
		scheme = "https"
	}
	return fmt.Sprintf("%s://%s/api/sso/%s/callback", scheme, c.Request.Host, provider)
}

// ─────────────────────────────────────────────
// Route handlers
// ─────────────────────────────────────────────

// SSOProviders returns which SSO providers are currently enabled.
// This is a public endpoint — no authentication required.
//
// GET /api/sso/providers
//
// Response:
//
//	{
//	  "providers": ["github", "google"],   // only enabled ones
//	  "allow_local_login": true
//	}
func SSOProviders(c *gin.Context) {
	enabled := make([]string, 0, 3)
	if config.SSOConfig.GitHub.ClientID != "" {
		enabled = append(enabled, "github")
	}
	if config.SSOConfig.Google.ClientID != "" {
		enabled = append(enabled, "google")
	}
	if config.SSOConfig.OIDC.Enabled && config.SSOConfig.OIDC.ClientID != "" {
		enabled = append(enabled, "oidc")
	}
	c.JSON(http.StatusOK, gin.H{
		"providers":         enabled,
		"allow_local_login": config.SSOConfig.AllowLocalLogin,
	})
}

// SSOLoginRedirect starts the OAuth2 authorization flow.
//
// GET /api/sso/:provider/login?method=sso_get_token|get_sso_id
//
//   - sso_get_token  → full login, issues a JWT on callback
//   - get_sso_id     → only returns the provider-side ID (used to bind SSO
//     to an existing account from the settings page)
func SSOLoginRedirect(c *gin.Context) {
	provider := c.Param("provider")
	method := c.DefaultQuery("method", "sso_get_token")
	if !isValidMethod(method) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid method, use sso_get_token or get_sso_id"})
		return
	}

	redirectURI := callbackURI(c, provider)
	state := generateState(provider, c.ClientIP(), method)

	var authURL string
	switch provider {
	case "github":
		if config.SSOConfig.GitHub.ClientID == "" {
			c.JSON(http.StatusNotImplemented, gin.H{"error": "GitHub SSO not configured"})
			return
		}
		authURL = githubOAuth2Config(redirectURI).AuthCodeURL(state)
	case "google":
		if config.SSOConfig.Google.ClientID == "" {
			c.JSON(http.StatusNotImplemented, gin.H{"error": "Google SSO not configured"})
			return
		}
		authURL = googleOAuth2Config(redirectURI).AuthCodeURL(state)
	case "oidc":
		if !config.SSOConfig.OIDC.Enabled || config.SSOConfig.OIDC.ClientID == "" {
			c.JSON(http.StatusNotImplemented, gin.H{"error": "OIDC SSO not configured"})
			return
		}
		oidcCfg, err := oidcOAuth2Config(c.Request.Context(), redirectURI)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		authURL = oidcCfg.AuthCodeURL(state)
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "unsupported provider: " + provider})
		return
	}

	c.Redirect(http.StatusFound, authURL)
}

// SSOCallback handles the OAuth2 callback for all providers.
// The popup window calls postMessage to pass data back to the opener, then closes.
//
// GET /api/sso/:provider/callback?method=...&code=...&state=...
func SSOCallback(c *gin.Context, db *gorm.DB) {
	provider := c.Param("provider")

	method, ok := verifyState(provider, c.ClientIP(), c.Query("state"))
	if !ok {
		respondError(c, "invalid or expired state parameter")
		return
	}
	if !isValidMethod(method) {
		respondError(c, "invalid method in state")
		return
	}
	if c.Query("code") == "" {
		respondError(c, "no authorization code provided")
		return
	}

	redirectURI := callbackURI(c, provider)
	info, err := exchange(c, provider, c.Query("code"), redirectURI)
	if err != nil {
		respondError(c, err.Error())
		return
	}

	// get_sso_id: just return the provider-scoped ID so the frontend can bind it
	if method == "get_sso_id" {
		respondPostMessage(c, map[string]string{
			"sso_id": provider + ":" + info.providerID,
		})
		return
	}

	// sso_get_token: find or create local user, then issue JWT
	user, err := findOrCreateUser(db, provider, info)
	if err != nil {
		respondError(c, err.Error())
		return
	}
	token, err := issueJWT(user)
	if err != nil {
		respondError(c, "failed to issue token")
		return
	}
	respondPostMessage(c, map[string]string{"token": token})
}

// ─────────────────────────────────────────────
// Per-provider exchange
// ─────────────────────────────────────────────

type ssoUserInfo struct {
	providerID string
	username   string
	email      string
	avatar     string
}

func exchange(c *gin.Context, provider, code, redirectURI string) (*ssoUserInfo, error) {
	switch provider {
	case "github":
		return exchangeGitHub(c, code, redirectURI)
	case "google":
		return exchangeGoogle(c, code, redirectURI)
	case "oidc":
		return exchangeOIDC(c, code, redirectURI)
	default:
		return nil, errors.New("unsupported provider: " + provider)
	}
}

func exchangeGitHub(c *gin.Context, code, redirectURI string) (*ssoUserInfo, error) {
	tok, err := githubOAuth2Config(redirectURI).Exchange(c.Request.Context(), code)
	if err != nil {
		return nil, fmt.Errorf("token exchange failed: %w", err)
	}

	body, err := apiGet("https://api.github.com/user", tok.AccessToken, "token")
	if err != nil {
		return nil, err
	}

	var data map[string]interface{}
	if err := json.Unmarshal(body, &data); err != nil {
		return nil, err
	}

	info := &ssoUserInfo{}
	if id, ok := data["id"].(float64); ok {
		info.providerID = fmt.Sprintf("%.0f", id)
	}
	info.username, _ = data["login"].(string)
	if name, _ := data["name"].(string); name != "" {
		info.username = name
	}
	info.email, _ = data["email"].(string)
	info.avatar, _ = data["avatar_url"].(string)

	// GitHub may omit email in /user when it is set to private
	if info.email == "" {
		info.email = fetchGitHubPrimaryEmail(tok.AccessToken)
	}
	if info.providerID == "" {
		return nil, errors.New("cannot get user ID from GitHub")
	}
	return info, nil
}

func exchangeGoogle(c *gin.Context, code, redirectURI string) (*ssoUserInfo, error) {
	tok, err := googleOAuth2Config(redirectURI).Exchange(c.Request.Context(), code)
	if err != nil {
		return nil, fmt.Errorf("token exchange failed: %w", err)
	}

	body, err := apiGet("https://www.googleapis.com/oauth2/v2/userinfo", tok.AccessToken, "bearer")
	if err != nil {
		return nil, err
	}

	var data map[string]interface{}
	if err := json.Unmarshal(body, &data); err != nil {
		return nil, err
	}

	info := &ssoUserInfo{}
	info.providerID, _ = data["id"].(string)
	info.username, _ = data["name"].(string)
	info.email, _ = data["email"].(string)
	info.avatar, _ = data["picture"].(string)

	if info.providerID == "" {
		return nil, errors.New("cannot get user ID from Google")
	}
	return info, nil
}

func exchangeOIDC(c *gin.Context, code, redirectURI string) (*ssoUserInfo, error) {
	oidcCfg := config.SSOConfig.OIDC

	oauth2Cfg, err := oidcOAuth2Config(c.Request.Context(), redirectURI)
	if err != nil {
		return nil, err
	}
	tok, err := oauth2Cfg.Exchange(c.Request.Context(), code)
	if err != nil {
		return nil, fmt.Errorf("token exchange failed: %w", err)
	}

	var claims map[string]interface{}

	// Prefer id_token claims (avoids an extra round-trip to userinfo)
	if rawIDToken, ok := tok.Extra("id_token").(string); ok {
		if c, err := parseOIDCIDTokenClaims(rawIDToken); err == nil {
			claims = c
		}
	}

	// Fallback: hit userinfo endpoint
	if claims == nil {
		if oidcCfg.UserInfoURL == "" {
			return nil, errors.New("OIDC: id_token missing and OIDC_USERINFO_URL not configured")
		}
		body, err := apiGet(oidcCfg.UserInfoURL, tok.AccessToken, "bearer")
		if err != nil {
			return nil, err
		}
		if err := json.Unmarshal(body, &claims); err != nil {
			return nil, err
		}
	}

	info := claimsToUserInfo(claims)
	if info.providerID == "" {
		return nil, errors.New("cannot get user ID from OIDC provider")
	}

	// ── OIDC_VERIFY_EMAIL ────────────────────────────────────────────────────
	if oidcCfg.VerifyEmail {
		if verified, _ := claims["email_verified"].(bool); !verified {
			return nil, errors.New("email address has not been verified by the OIDC provider")
		}
	}

	// ── OIDC_ALLOWED_GROUPS ──────────────────────────────────────────────────
	if len(oidcCfg.AllowedGroups) > 0 {
		if !isInAllowedGroups(claims, oidcCfg.GroupClaim, oidcCfg.AllowedGroups) {
			return nil, errors.New("you are not in an allowed group")
		}
	}

	return info, nil
}

// isInAllowedGroups checks whether the groups claim contains at least one of the allowed groups.
func isInAllowedGroups(claims map[string]interface{}, groupClaim string, allowed []string) bool {
	raw, ok := claims[groupClaim]
	if !ok {
		return false
	}
	// groups claim is typically []interface{} (JSON array of strings)
	groups, ok := raw.([]interface{})
	if !ok {
		return false
	}
	allowedSet := make(map[string]struct{}, len(allowed))
	for _, g := range allowed {
		allowedSet[g] = struct{}{}
	}
	for _, g := range groups {
		if s, ok := g.(string); ok {
			if _, found := allowedSet[s]; found {
				return true
			}
		}
	}
	return false
}

// parseOIDCIDTokenClaims decodes the payload of the id_token without signature verification.
// Signature verification is skipped here because the token was obtained directly via
// a back-channel code exchange (not supplied by the user), so it is already trusted.
func parseOIDCIDTokenClaims(rawIDToken string) (map[string]interface{}, error) {
	parts := strings.Split(rawIDToken, ".")
	if len(parts) < 2 {
		return nil, errors.New("malformed id_token")
	}
	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, err
	}
	var claims map[string]interface{}
	if err := json.Unmarshal(payload, &claims); err != nil {
		return nil, err
	}
	return claims, nil
}

func claimsToUserInfo(claims map[string]interface{}) *ssoUserInfo {
	cfg := config.SSOConfig.OIDC
	info := &ssoUserInfo{}

	// subject / id
	for _, key := range []string{"sub", "id"} {
		if v, _ := claims[key].(string); v != "" {
			info.providerID = v
			break
		}
	}

	// name — respect OIDC_NAME_CLAIM, fall back to preferred_username
	if name, _ := claims[cfg.NameClaim].(string); name != "" {
		info.username = name
	}
	if info.username == "" {
		info.username, _ = claims["preferred_username"].(string)
	}

	// email — respect OIDC_EMAIL_CLAIM
	info.email, _ = claims[cfg.EmailClaim].(string)
	info.avatar, _ = claims["picture"].(string)
	return info
}

// ─────────────────────────────────────────────
// User find-or-create
// ─────────────────────────────────────────────

func findOrCreateUser(db *gorm.DB, provider string, info *ssoUserInfo) (*model.User, error) {
	// 1. Exact SSO binding match
	var binding model.SSOBinding
	if err := db.Where("provider = ? AND provider_id = ?", provider, info.providerID).First(&binding).Error; err == nil {
		var user model.User
		if err := db.First(&user, binding.UserID).Error; err != nil {
			return nil, errors.New("user account not found")
		}
		return &user, nil
	}

	// 2. Email match → link to existing account
	var user model.User
	if info.email != "" {
		db.Where("email = ?", info.email).First(&user)
	}

	// 3. Auto-register new user
	if user.ID == 0 {
		username := generateUsername(db, info.username, info.email)
		user = model.User{
			Username:        username,
			Email:           info.email,
			Role:            model.RoleGuest,
			PasswordVersion: 0,
			// No password set — this user can only log in via SSO
		}
		if err := db.Create(&user).Error; err != nil {
			return nil, fmt.Errorf("failed to create user: %w", err)
		}
	}

	// Persist binding so future logins skip steps 2-3
	db.Create(&model.SSOBinding{
		UserID:     user.ID,
		Provider:   provider,
		ProviderID: info.providerID,
		Email:      info.email,
		Name:       info.username,
		Avatar:     info.avatar,
	})

	return &user, nil
}

// ─────────────────────────────────────────────
// JWT
// ─────────────────────────────────────────────

func issueJWT(user *model.User) (string, error) {
	claims := jwt.MapClaims{
		"user_id":          user.ID,
		"username":         user.Username,
		"role":             user.Role,
		"password_version": user.PasswordVersion,
		"exp":              time.Now().Add(7 * 24 * time.Hour).Unix(),
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(config.JWTSecret)
}

// ─────────────────────────────────────────────
// Utilities
// ─────────────────────────────────────────────

func generateUsername(db *gorm.DB, name, email string) string {
	base := name
	if base == "" {
		if idx := strings.Index(email, "@"); idx > 0 {
			base = email[:idx]
		}
	}
	if base == "" {
		base = "user"
	}
	var sb strings.Builder
	for _, ch := range base {
		if (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') || (ch >= '0' && ch <= '9') || ch == '_' {
			sb.WriteRune(ch)
		}
	}
	if sb.Len() == 0 {
		sb.WriteString("user")
	}
	candidate := sb.String()

	var count int64
	for suffix := 0; ; suffix++ {
		username := candidate
		if suffix > 0 {
			username = fmt.Sprintf("%s%d", candidate, suffix)
		}
		db.Model(&model.User{}).Where("username = ?", username).Count(&count)
		if count == 0 {
			return username
		}
	}
}

func apiGet(url, accessToken, scheme string) ([]byte, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	if strings.ToLower(scheme) == "token" {
		req.Header.Set("Authorization", "token "+accessToken) // GitHub style
	} else {
		req.Header.Set("Authorization", "Bearer "+accessToken)
	}
	req.Header.Set("Accept", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return io.ReadAll(resp.Body)
}

func fetchGitHubPrimaryEmail(accessToken string) string {
	body, err := apiGet("https://api.github.com/user/emails", accessToken, "token")
	if err != nil {
		return ""
	}
	var emails []map[string]interface{}
	if err := json.Unmarshal(body, &emails); err != nil {
		return ""
	}
	for _, e := range emails {
		if primary, _ := e["primary"].(bool); !primary {
			continue
		}
		if verified, _ := e["verified"].(bool); !verified {
			continue
		}
		if email, _ := e["email"].(string); email != "" {
			return email
		}
	}
	return ""
}

func isValidMethod(method string) bool {
	return method == "sso_get_token" || method == "get_sso_id"
}

// respondPostMessage renders an HTML popup page that sends data back to opener via postMessage.
// Pattern mirrors OpenList's implementation.
// SSO_STORAGE_KEY must match the constant in the frontend ssoLogin() helper.
const ssoStorageKey = "sso_callback_result"

// respondPostMessage writes the result to localStorage so the opener window
// can pick it up via the 'storage' event. Using localStorage instead of
// postMessage avoids the window.opener=null issue caused by cross-origin
// redirects during the OAuth2 / OIDC flow.
func respondPostMessage(c *gin.Context, data map[string]string) {
	pairs := make([]string, 0, len(data))
	for k, v := range data {
		pairs = append(pairs, fmt.Sprintf(`%q:%q`, k, v))
	}
	payload := "{" + strings.Join(pairs, ",") + "}"
	html := fmt.Sprintf(`<!DOCTYPE html>
<head></head>
<body>
<script>
try { localStorage.setItem(%q, JSON.stringify(%s)) } catch(e) {}
window.close()
</script>
</body>`, ssoStorageKey, payload)
	c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(html))
}

// respondError writes an error result to localStorage and closes the popup.
func respondError(c *gin.Context, msg string) {
	html := fmt.Sprintf(`<!DOCTYPE html>
<head></head>
<body>
<script>
try { localStorage.setItem(%q, JSON.stringify({"error":%q})) } catch(e) {}
window.close()
</script>
</body>`, ssoStorageKey, msg)
	c.Data(http.StatusBadRequest, "text/html; charset=utf-8", []byte(html))
}
