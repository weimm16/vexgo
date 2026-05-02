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

		// Extract asset name from filename pattern: {name}-{hash}.{ext}
		// e.g., "index-DR2bYJZO.js" -> name="index", ext="js"
		ext := filepath.Ext(filename) // ".css" or ".js"
		if ext != ".css" && ext != ".js" {
			return nil
		}

		base := filename[:len(filename)-len(ext)] // "index-DR2bYJZO"
		if idx := strings.LastIndex(base, "-"); idx != -1 {
			assetName := base[:idx] // "index"
			assetPath := "/assets/" + filename

			if ext == ".css" {
				manifest.CSS[assetName] = assetPath
			} else {
				manifest.JS[assetName] = assetPath
			}
		}

		return nil
	})

	return err
}

// GetAssetURL returns the actual URL for a logical asset name
func GetAssetURL(assetType, name string) string {
	var manifestMap map[string]string
	switch assetType {
	case "css":
		manifestMap = manifest.CSS
	case "js":
		manifestMap = manifest.JS
	default:
		return ""
	}

	return manifestMap[name]
}
