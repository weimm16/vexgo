package public

import (
	"embed"
	"io/fs"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
)

//go:embed dist/**/*
var staticFS embed.FS

//go:embed dist/index.html
var indexHTML []byte

// GetStaticFS returns the embedded static filesystem
// It strips the "dist" prefix to provide clean paths
func GetStaticFS() fs.FS {
	sub, err := fs.Sub(staticFS, "dist")
	if err != nil {
		panic(err)
	}
	return sub
}

// GetIndexHTML returns the embedded index.html content
func GetIndexHTML() []byte {
	return indexHTML
}

// ReadAsset reads an asset file from the embedded filesystem
func ReadAsset(path string) ([]byte, error) {
	// Read from the original embed.FS with the dist prefix
	// Use forward slashes for embed.FS compatibility across platforms
	return staticFS.ReadFile("dist/" + path)
}

// AssetExists checks if an asset exists in the embedded filesystem
func AssetExists(path string) bool {
	_, err := staticFS.ReadFile("dist/" + path)
	return err == nil
}

// RegisterStaticRoutes registers all static file routes
// Parameters:
//   - r: Gin router to register routes on
//   - dataDir: directory for local uploads (only used if s3Enabled is false)
//   - s3Enabled: whether S3 storage is enabled (if true, local uploads won't be served)
func RegisterStaticRoutes(r *gin.Engine, dataDir string, s3Enabled bool) {
	// Serve local uploads if S3 is not enabled
	if !s3Enabled {
		mediaDir := filepath.Join(dataDir, "media")
		r.Static("/uploads", mediaDir)
	}

	// Serve embedded assets
	r.GET("/assets/*filepath", func(c *gin.Context) {
		file := strings.TrimPrefix(c.Param("filepath"), "/")
		content, err := ReadAsset("assets/" + file)
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

	// Root route - serve index.html
	r.GET("/", func(c *gin.Context) {
		c.Data(200, "text/html; charset=utf-8", GetIndexHTML())
	})

	// SPA fallback - for any non-API route, serve index.html
	r.NoRoute(func(c *gin.Context) {
		if !strings.HasPrefix(c.Request.URL.Path, "/api/") {
			c.Data(200, "text/html; charset=utf-8", GetIndexHTML())
			return
		}
		c.JSON(404, gin.H{"error": "Not Found"})
	})
}
