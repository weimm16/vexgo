package public

import (
	"embed"
	"io/fs"
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
