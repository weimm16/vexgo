package main

import (
	"embed"
	"io/fs"
	"path/filepath"
)

//go:embed static/**/*
var staticFS embed.FS

//go:embed static/index.html
var indexHTML []byte

// GetStaticFS returns the embedded static filesystem
// It strips the "static" prefix to provide clean paths
func GetStaticFS() fs.FS {
	sub, err := fs.Sub(staticFS, "static")
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
	// Read from the original embed.FS with the static prefix
	return staticFS.ReadFile(filepath.Join("static", path))
}

// AssetExists checks if an asset exists in the embedded filesystem
func AssetExists(path string) bool {
	_, err := staticFS.ReadFile(filepath.Join("static", path))
	return err == nil
}
