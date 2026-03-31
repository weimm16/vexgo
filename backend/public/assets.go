package public

import (
	"encoding/json"
	"io/fs"
	"path/filepath"
	"strings"
)

// AssetManifest holds the mapping of logical asset names to their hashed filenames
type AssetManifest struct {
	CSS map[string]string `json:"css"`
	JS  map[string]string `json:"js"`
}

var manifest AssetManifest

// LoadAssetManifest loads the asset manifest from the embedded filesystem
func LoadAssetManifest() error {
	// Try to load manifest.json from embedded fs
	manifestData, err := staticFS.ReadFile("dist/manifest.json")
	if err != nil {
		// If manifest doesn't exist, build it dynamically
		return buildAssetManifest()
	}

	return json.Unmarshal(manifestData, &manifest)
}

// buildAssetManifest builds the manifest by scanning the assets directory
func buildAssetManifest() error {
	manifest = AssetManifest{
		CSS: make(map[string]string),
		JS:  make(map[string]string),
	}

	// Walk through the embedded assets directory
	err := fs.WalkDir(staticFS, "dist/assets", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			return nil
		}

		// Get just the filename
		filename := filepath.Base(path)

		// Check if it's a CSS or JS file with hash pattern
		if strings.HasPrefix(filename, "index-") && strings.HasSuffix(filename, ".css") {
			manifest.CSS["index"] = "/assets/" + filename
		} else if strings.HasPrefix(filename, "index-") && strings.HasSuffix(filename, ".js") {
			manifest.JS["index"] = "/assets/" + filename
		} else if strings.HasPrefix(filename, "react-vendor-") && strings.HasSuffix(filename, ".js") {
			manifest.JS["react-vendor"] = "/assets/" + filename
		} else if strings.HasPrefix(filename, "ui-vendor-") && strings.HasSuffix(filename, ".js") {
			manifest.JS["ui-vendor"] = "/assets/" + filename
		} else if strings.HasPrefix(filename, "utils-vendor-") && strings.HasSuffix(filename, ".js") {
			manifest.JS["utils-vendor"] = "/assets/" + filename
		}

		return nil
	})

	return err
}

// GetAssetURL returns the actual URL for a logical asset name
func GetAssetURL(assetType, name string) string {
	switch assetType {
	case "css":
		if url, ok := manifest.CSS[name]; ok {
			return url
		}
	case "js":
		if url, ok := manifest.JS[name]; ok {
			return url
		}
	}

	// Fallback to hardcoded values if manifest lookup fails
	// These should match your current build
	switch name {
	case "index":
		if assetType == "css" {
			return "/assets/index-BTvxqpsA.css"
		}
		return "/assets/index-DR2bYJZO.js"
	case "react-vendor":
		return "/assets/react-vendor-BmqGXi6J.js"
	case "ui-vendor":
		return "/assets/ui-vendor-CEsCEvQe.js"
	case "utils-vendor":
		return "/assets/utils-vendor-42ANG6Sg.js"
	}

	return ""
}
