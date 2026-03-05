package handler

import (
	"blog-system/backend/model"
	"log"

	"golang.org/x/crypto/bcrypt"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var db *gorm.DB

func InitDB() {
	var err error
	db, err = gorm.Open(sqlite.Open("blog.db"), &gorm.Config{})
	if err != nil {
		log.Fatalf("failed to connect database: %v", err)
	}

	// 自动迁移模型
	if err := db.AutoMigrate(
		&model.Post{},
		&model.User{},
		&model.Tag{},
		&model.Category{},
		&model.Comment{},
		&model.Like{},
		&model.MediaFile{},
	); err != nil {
		log.Fatalf("auto migrate failed: %v", err)
	}

	// 创建一个默认超级管理员（如果不存在），使用 bcrypt 存储密码
	var u model.User
	if err := db.Where("username = ?", "admin").First(&u).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			pwHash, err := bcrypt.GenerateFromPassword([]byte("password"), bcrypt.DefaultCost)
			if err != nil {
				log.Printf("failed to hash default admin password: %v", err)
			} else {
				u = model.User{Username: "admin", Email: "admin@example.com", Password: string(pwHash), Role: "super_admin"}
				if err := db.Create(&u).Error; err != nil {
					log.Printf("failed to create default admin: %v", err)
				}
			}
		}
	}

	// 创建一个默认分类（如果不存在）
	var defaultCategory model.Category
	if err := db.Where("name = ?", "Default").First(&defaultCategory).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			defaultCategory = model.Category{
				Name:        "Default",
				Description: "默认分类，用于未指定分类的文章",
			}
			if err := db.Create(&defaultCategory).Error; err != nil {
				log.Printf("failed to create default category: %v", err)
			} else {
				log.Println("default category created: Default")
			}
		}
	}
}

// DB returns the database instance
func DB() *gorm.DB {
	return db
}
