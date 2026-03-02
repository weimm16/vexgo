package handler

import (
	"blog-system/backend/model"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// 获取统计信息
func GetStats(c *gin.Context) {
	var postsCount, usersCount, categoriesCount, tagsCount, commentsCount int64

	db.Model(&model.Post{}).Count(&postsCount)
	db.Model(&model.User{}).Count(&usersCount)
	db.Model(&model.Category{}).Count(&categoriesCount)
	db.Model(&model.Tag{}).Count(&tagsCount)
	db.Model(&model.Comment{}).Count(&commentsCount)

	c.JSON(http.StatusOK, gin.H{
		"stats": gin.H{
			"posts":      postsCount,
			"users":      usersCount,
			"comments":   commentsCount,
			"categories": categoriesCount,
			"tags":       tagsCount,
		},
	})
}

// 获取热门文章
func GetPopularPosts(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "5"))
	var posts []model.Post

	db.Where("status = ?", "published").
		Order("created_at DESC").
		Limit(limit).
		Preload("Author").
		Preload("Tags").
		Find(&posts)

	c.JSON(http.StatusOK, gin.H{
		"posts": posts,
	})
}

// 获取最新文章
func GetLatestPosts(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "5"))
	var posts []model.Post

	db.Where("status = ?", "published").
		Order("created_at DESC").
		Limit(limit).
		Preload("Author").
		Preload("Tags").
		Find(&posts)

	c.JSON(http.StatusOK, gin.H{
		"posts": posts,
	})
}

// 获取分类列表
func GetCategories(c *gin.Context) {
	var categories []model.Category

	db.Find(&categories)

	c.JSON(http.StatusOK, gin.H{
		"categories": categories,
	})
}

// 创建分类
func CreateCategory(c *gin.Context) {
	var req struct {
		Name        string `json:"name" binding:"required"`
		Description string `json:"description"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	category := model.Category{
		Name:        req.Name,
		Description: req.Description,
	}

	if err := db.Create(&category).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建分类失败"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":  "分类创建成功",
		"category": category,
	})
}

// 获取标签列表
func GetTags(c *gin.Context) {
	var tags []model.Tag

	db.Find(&tags)

	c.JSON(http.StatusOK, gin.H{
		"tags": tags,
	})
}
