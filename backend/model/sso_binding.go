// backend/model/sso_binding.go
package model

import "gorm.io/gorm"

// SSOBinding stores the mapping between a local user and an external SSO provider identity.
// One user can have multiple SSO bindings (e.g. both GitHub and Google).
type SSOBinding struct {
	gorm.Model
	UserID     uint   `gorm:"not null;index"`
	Provider   string `gorm:"not null;size:32"`  // "github" | "google" | "oidc"
	ProviderID string `gorm:"not null;size:256"` // Unique ID from the provider
	Email      string `gorm:"size:256"`
	Name       string `gorm:"size:256"`
	Avatar     string `gorm:"size:1024"`

	// Composite unique constraint: one provider account maps to exactly one local user
	// gorm tag below creates a unique index on (provider, provider_id)
}

// TableName sets the table name for SSOBinding
func (SSOBinding) TableName() string {
	return "sso_bindings"
}

// Migrate adds SSOBinding table and the unique index.
// Call this alongside your existing AutoMigrate in main.go or migrations.
//
//	db.AutoMigrate(&model.SSOBinding{})
//	db.Exec("CREATE UNIQUE INDEX IF NOT EXISTS idx_sso_provider_id ON sso_bindings (provider, provider_id)")
