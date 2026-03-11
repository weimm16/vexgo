// backend/router/sso_routes.go
package router

import (
	"net/http"

	"vexgo/backend/config"
	"vexgo/backend/handler"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// RegisterSSORoutes mounts SSO endpoints and optional middleware onto the engine.
// Call this from your main router setup after other routes are registered.
//
//	router.RegisterSSORoutes(r, db)
func RegisterSSORoutes(r *gin.Engine, db *gorm.DB) {
	sso := r.Group("/api/sso")
	{
		// Public: which providers are enabled (used by frontend to show/hide buttons)
		sso.GET("/providers", handler.SSOProviders)

		// Step 1 – open in a popup (or redirect directly when AutoRedirect is on)
		//   /api/sso/github/login?method=sso_get_token
		//   /api/sso/oidc/login?method=get_sso_id
		sso.GET("/:provider/login", handler.SSOLoginRedirect)

		// Step 2 – provider redirects back here, popup sends postMessage → closes
		sso.GET("/:provider/callback", func(c *gin.Context) {
			handler.SSOCallback(c, db)
		})
	}

	// ── OIDC auto-redirect ────────────────────────────────────────────────────
	// When OIDC_AUTO_REDIRECT=true, hitting the login page skips the form and
	// goes straight to the OIDC provider. Wire this up in your login route:
	//
	//   r.GET("/login", func(c *gin.Context) {
	//       if config.SSOConfig.OIDC.AutoRedirect {
	//           c.Redirect(http.StatusFound, "/api/sso/oidc/login?method=sso_get_token")
	//           return
	//       }
	//       // render normal login page
	//   })

	// ── ALLOW_LOCAL_LOGIN=false guard ────────────────────────────────────────
	// Attach this middleware to your password-login endpoint to block it when
	// the operator has disabled local credentials:
	//
	//   authGroup.POST("/login", LocalLoginGuard(), handler.Login)
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

// ── Migration ────────────────────────────────────────────────────────────────
// Add to your existing AutoMigrate block in main.go:
//
//	db.AutoMigrate(&model.SSOBinding{})
//	db.Exec(`CREATE UNIQUE INDEX IF NOT EXISTS idx_sso_provider_id
//	         ON sso_bindings (provider, provider_id)`)
//
// ── go.mod dependencies ──────────────────────────────────────────────────────
//	go get github.com/coreos/go-oidc
//	go get golang.org/x/oauth2
//
// ── Frontend usage ───────────────────────────────────────────────────────────
//
//	const popup = window.open(
//	  '/api/sso/github/login?method=sso_get_token',
//	  'sso_login', 'width=600,height=700'
//	)
//	window.addEventListener('message', (e) => {
//	  if (e.data.token)  loginWithToken(e.data.token)
//	  if (e.data.sso_id) bindAccount(e.data.sso_id)   // get_sso_id flow
//	  if (e.data.error)  showError(e.data.error)
//	})
