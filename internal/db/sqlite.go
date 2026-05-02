package db

import (
	"os"
	"path/filepath"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func NewConnection(dsn string) (*gorm.DB, error) {
	if err := os.MkdirAll(filepath.Dir(dsn), 0755); err != nil {
		return nil, err
	}
	return gorm.Open(sqlite.Open(dsn), &gorm.Config{})
}

func AutoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&Post{}, &Page{}, &User{}, &Category{},
		&Media{}, &Setting{}, &Plugin{}, &Theme{},
	)
}
