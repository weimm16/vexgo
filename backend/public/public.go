package public

import (
	"embed"
	"io/fs"
	"path/filepath"
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
	return staticFS.ReadFile(filepath.Join("dist", path))
}

// AssetExists checks if an asset exists in the embedded filesystem
func AssetExists(path string) bool {
	_, err := staticFS.ReadFile(filepath.Join("dist", path))
	return err == nil
}
