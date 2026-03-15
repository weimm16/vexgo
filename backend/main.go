package main

import (
	"fmt"
	"path/filepath"
	"strings"
	"vexgo/backend/cmd"
	"vexgo/backend/config"
	"vexgo/backend/handler"
	"vexgo/backend/middleware"
	"vexgo/backend/public"

	"github.com/gin-gonic/gin"
)

func main() {
	// 1. Parse command line arguments
	cfg := cmd.ParseFlags()

	// 2. Initialize configuration (load JWT secret, etc., support config files and environment variables)
	config.Init(cfg.JWTSecret)

	// 2.1 Load SSO configuration from config file (overrides environment variables)
	config.LoadFromConfig(cfg)

	// Set data directory (for file uploads, only used if S3 is not enabled)
	handler.DataDir = cfg.DataDir

	// 3. Initialize S3 storage if enabled
	if cfg.S3Enabled {
		s3Cfg := &config.S3Config{
			Enabled:                  cfg.S3Enabled,
			Endpoint:                 cfg.S3Endpoint,
			Region:                   cfg.S3Region,
			Bucket:                   cfg.S3Bucket,
			AccessKey:                cfg.S3AccessKey,
			SecretKey:                cfg.S3SecretKey,
			ForcePath:                cfg.S3ForcePath,
			CustomDomain:             cfg.S3CustomDomain,
			DisableBucketInCustomURL: cfg.S3DisableBucketInCustomURL,
		}
		fmt.Printf("S3 Config Loaded: Enabled=%v, Endpoint=%s, Region=%s, Bucket=%s, CustomDomain=%s, DisableBucketInCustomURL=%v\n", s3Cfg.Enabled, s3Cfg.Endpoint, s3Cfg.Region, s3Cfg.Bucket, s3Cfg.CustomDomain, s3Cfg.DisableBucketInCustomURL)
		if err := handler.InitS3(s3Cfg); err != nil {
			panic(err)
		}
		fmt.Println("S3 storage initialized")
	} else {
		fmt.Println("Using local file storage")
	}

	// 4. Initialize database connection (ensure database driver and connection string are configured correctly)
	handler.InitDB(cfg, cfg.DataDir)
	// Set database connection to authentication middleware
	middleware.SetDB(handler.DB())

	// 4. Create Gin engine instance (includes Logger and Recovery middleware by default)
	r := gin.Default()

	// Configure trusted proxies based on environment/configuration
	// If BEHIND_REVERSE_PROXY=true, use TRUSTED_PROXIES list or common defaults
	// If BEHIND_REVERSE_PROXY=false, disable proxy trust (no warning)
	if cfg.BehindReverseProxy {
		if len(cfg.TrustedProxies) > 0 {
			// Use explicitly configured trusted proxies
			r.SetTrustedProxies(cfg.TrustedProxies)
			fmt.Printf("Trusted proxies configured: %v\n", cfg.TrustedProxies)
		} else {
			// Use common defaults: trust all private IP ranges and localhost
			// This is a reasonable default for self-hosted behind reverse proxy
			defaultProxies := []string{
				"127.0.0.1",
				"::1",
				"192.168.0.0/16",
				"10.0.0.0/8",
				"172.16.0.0/12",
			}
			r.SetTrustedProxies(defaultProxies)
			fmt.Printf("Trusted proxies set to common private networks (behind reverse proxy)\n")
		}
	} else {
		// Not behind a reverse proxy, disable trust
		r.SetTrustedProxies(nil)
	}

	// ===================== Core API routing group (all endpoints under /api) =====================
	// All API routing definitions have been moved to handler.RegisterAPIRoutes to avoid cluttering main.go.
	handler.RegisterAPIRoutes(r)

	// ===================== Static file hosting =====================
	// Only serve local uploads if S3 is not enabled
	if !cfg.S3Enabled {
		mediaDir := filepath.Join(cfg.DataDir, "media")
		r.Static("/uploads", mediaDir)
	}

	r.GET("/assets/*filepath", func(c *gin.Context) {
		file := strings.TrimPrefix(c.Param("filepath"), "/")
		content, err := public.ReadAsset("assets/" + file)
		if err != nil {
			c.Status(404)
			return
		}
		ext := filepath.Ext(file)
		switch ext {
		case ".js":
			c.Data(200, "application/javascript", content)
		case ".css":
			c.Data(200, "text/css", content)
		case ".html":
			c.Data(200, "text/html", content)
		case ".json":
			c.Data(200, "application/json", content)
		case ".png":
			c.Data(200, "image/png", content)
		case ".jpg", ".jpeg":
			c.Data(200, "image/jpeg", content)
		case ".gif":
			c.Data(200, "image/gif", content)
		case ".svg":
			c.Data(200, "image/svg+xml", content)
		case ".ico":
			c.Data(200, "image/x-icon", content)
		case ".woff":
			c.Data(200, "font/woff", content)
		case ".woff2":
			c.Data(200, "font/woff2", content)
		default:
			c.Data(200, "application/octet-stream", content)
		}
	})

	r.GET("/", func(c *gin.Context) {
		c.Data(200, "text/html; charset=utf-8", public.GetIndexHTML())
	})

	// ===================== Frontend SPA route compatibility =====================
	r.NoRoute(func(c *gin.Context) {
		if !strings.HasPrefix(c.Request.URL.Path, "/api/") {
			c.Data(200, "text/html; charset=utf-8", public.GetIndexHTML())
			return
		}
		c.JSON(404, gin.H{"error": "Not Found"})
	})

	r.Run(cfg.GetListenAddr())
}
