// backend/config/sso.go
package config

import (
	"os"
	"strconv"
	"strings"
)

// SSOProviderConfig holds OAuth2 credentials for a single provider.
type SSOProviderConfig struct {
	ClientID     string
	ClientSecret string
}

// OIDCConfig holds the full configuration for OIDC-based SSO.
// Supports standard OIDC discovery (IssuerURL) as well as manual endpoint
// override (AuthURL / TokenURL) for non-standard providers.
type OIDCConfig struct {
	SSOProviderConfig

	// ── Required ─────────────────────────────────────────────────────────────
	Enabled   bool   // OIDC_ENABLED
	IssuerURL string // OIDC_ISSUER_URL  — used for OIDC discovery (.well-known/openid-configuration)

	// ── Manual endpoint override (only needed when discovery is unavailable) ─
	AuthURL     string // OIDC_AUTH_URL
	TokenURL    string // OIDC_TOKEN_URL
	UserInfoURL string // OIDC_USERINFO_URL

	// ── Scopes ───────────────────────────────────────────────────────────────
	// Default: "openid profile email"
	// Some providers need extra scopes, e.g. "openid profile email groups"
	Scopes []string // OIDC_SCOPES  (space-separated)

	// ── Claims ───────────────────────────────────────────────────────────────
	// Override if your provider uses non-standard claim names.
	EmailClaim string // OIDC_EMAIL_CLAIM  (default: "email")
	NameClaim  string // OIDC_NAME_CLAIM   (default: "name")
	GroupClaim string // OIDC_GROUP_CLAIM  (default: "groups")

	// ── Access control ───────────────────────────────────────────────────────
	// Comma-separated list of groups allowed to log in.
	// Empty = allow all users.
	AllowedGroups []string // OIDC_ALLOWED_GROUPS  (comma-separated)

	// ── UX ───────────────────────────────────────────────────────────────────
	AutoRedirect bool // OIDC_AUTO_REDIRECT  — skip login page, go straight to provider
	VerifyEmail  bool // OIDC_VERIFY_EMAIL   — require email_verified=true in token
}

type ssoConfig struct {
	// BaseURL overrides auto-detected callback host, e.g. https://vexgo.yzlab.de
	BaseURL string

	// Simple OAuth2 providers
	GitHub SSOProviderConfig
	Google SSOProviderConfig

	// Full OIDC (Keycloak / Authentik / Authelia / Okta / Casdoor / university SSO …)
	OIDC OIDCConfig

	// Global option: set false to force SSO-only (disable password login)
	AllowLocalLogin bool // ALLOW_LOCAL_LOGIN (default: true)
}

// SSOConfig is the global SSO configuration, populated from environment
// variables at process startup. A provider is enabled when its CLIENT_ID
// (or OIDC_ENABLED) is non-empty / true.
//
// ── GitHub ───────────────────────────────────────────────────────────────────
//   GITHUB_CLIENT_ID        GitHub OAuth App Client ID
//   GITHUB_CLIENT_SECRET    GitHub OAuth App Client Secret
//
// ── Google ───────────────────────────────────────────────────────────────────
//   GOOGLE_CLIENT_ID        Google OAuth 2.0 Client ID
//   GOOGLE_CLIENT_SECRET    Google OAuth 2.0 Client Secret
//
// ── OIDC ─────────────────────────────────────────────────────────────────────
//   OIDC_ENABLED            Enable OIDC login (default: false)
//   OIDC_ISSUER_URL         Issuer URL for OIDC discovery
//                           e.g. https://auth.example.com/realms/myrealm
//                           Tip: verify at <issuer>/.well-known/openid-configuration
//   OIDC_CLIENT_ID          Client ID
//   OIDC_CLIENT_SECRET      Client Secret
//
//   Manual endpoint override (only when OIDC discovery is unavailable):
//   OIDC_AUTH_URL           Authorization endpoint
//   OIDC_TOKEN_URL          Token endpoint
//   OIDC_USERINFO_URL       UserInfo endpoint (optional fallback)
//
//   OIDC_SCOPES             Space-separated extra scopes (default: "openid profile email")
//                           e.g. "openid profile email groups"
//   OIDC_EMAIL_CLAIM        Claim name for email    (default: "email")
//   OIDC_NAME_CLAIM         Claim name for username (default: "name")
//   OIDC_GROUP_CLAIM        Claim name for groups   (default: "groups")
//   OIDC_ALLOWED_GROUPS     Comma-separated groups permitted to log in
//                           e.g. "admins,developers"  (empty = allow all)
//   OIDC_AUTO_REDIRECT      Redirect to OIDC provider automatically (default: false)
//   OIDC_VERIFY_EMAIL       Require email_verified=true in token    (default: false)
//
// ── Global ───────────────────────────────────────────────────────────────────
//   ALLOW_LOCAL_LOGIN       Allow password-based login (default: true)
//                           Set false to enforce SSO-only access
var SSOConfig = ssoConfig{
	BaseURL: os.Getenv("BASE_URL"),
	GitHub: SSOProviderConfig{
		ClientID:     os.Getenv("GITHUB_CLIENT_ID"),
		ClientSecret: os.Getenv("GITHUB_CLIENT_SECRET"),
	},
	Google: SSOProviderConfig{
		ClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
		ClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
	},
	OIDC: OIDCConfig{
		Enabled:   parseBool("OIDC_ENABLED", false),
		IssuerURL: os.Getenv("OIDC_ISSUER_URL"),
		SSOProviderConfig: SSOProviderConfig{
			ClientID:     os.Getenv("OIDC_CLIENT_ID"),
			ClientSecret: os.Getenv("OIDC_CLIENT_SECRET"),
		},
		AuthURL:     os.Getenv("OIDC_AUTH_URL"),
		TokenURL:    os.Getenv("OIDC_TOKEN_URL"),
		UserInfoURL: os.Getenv("OIDC_USERINFO_URL"),
		Scopes:        parseScopes("OIDC_SCOPES", []string{"openid", "profile", "email"}),
		EmailClaim:    getEnvOrDefault("OIDC_EMAIL_CLAIM", "email"),
		NameClaim:     getEnvOrDefault("OIDC_NAME_CLAIM", "name"),
		GroupClaim:    getEnvOrDefault("OIDC_GROUP_CLAIM", "groups"),
		AllowedGroups: parseCommaSeparated("OIDC_ALLOWED_GROUPS"),
		AutoRedirect:  parseBool("OIDC_AUTO_REDIRECT", false),
		VerifyEmail:   parseBool("OIDC_VERIFY_EMAIL", false),
	},
	AllowLocalLogin: parseBool("ALLOW_LOCAL_LOGIN", true),
}

// ── helpers ───────────────────────────────────────────────────────────────────

func getEnvOrDefault(key, defaultVal string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultVal
}

func parseBool(key string, defaultVal bool) bool {
	v := os.Getenv(key)
	if v == "" {
		return defaultVal
	}
	b, err := strconv.ParseBool(v)
	if err != nil {
		return defaultVal
	}
	return b
}

// parseScopes parses a space-separated scope string.
// Falls back to defaultScopes when the env var is unset.
func parseScopes(key string, defaultScopes []string) []string {
	v := os.Getenv(key)
	if v == "" {
		return defaultScopes
	}
	return strings.Fields(v)
}

// parseCommaSeparated parses a comma-separated list, trimming whitespace.
// Returns nil (not an empty slice) when the env var is unset, so callers
// can use `len(AllowedGroups) == 0` to mean "allow all".
func parseCommaSeparated(key string) []string {
	v := os.Getenv(key)
	if v == "" {
		return nil
	}
	parts := strings.Split(v, ",")
	result := make([]string, 0, len(parts))
	for _, p := range parts {
		if s := strings.TrimSpace(p); s != "" {
			result = append(result, s)
		}
	}
	return result
}
