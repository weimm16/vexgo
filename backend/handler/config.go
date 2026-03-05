package handler

import (
	"net/http"

	"blog-system/backend/model"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// GetSMTPConfig 获取 SMTP 配置
func GetSMTPConfig(c *gin.Context) {
	var config model.SMTPConfig
	if err := db.First(&config).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			// 返回默认配置
			c.JSON(http.StatusOK, model.SMTPConfig{
				Enabled:   false,
				Host:      "",
				Port:      587,
				Username:  "",
				Password:  "",
				FromEmail: "",
				FromName:  "Blog System",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get SMTP config"})
		return
	}

	// 不返回密码字段
	config.Password = ""
	c.JSON(http.StatusOK, config)
}

// UpdateSMTPConfig 更新 SMTP 配置
func UpdateSMTPConfig(c *gin.Context) {
	var req struct {
		Enabled   bool   `json:"enabled"`
		Host      string `json:"host"`
		Port      int    `json:"port"`
		Username  string `json:"username"`
		Password  string `json:"password"` // 如果为空则不更新密码
		FromEmail string `json:"fromEmail"`
		FromName  string `json:"fromName"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 获取现有配置
	var config model.SMTPConfig
	if err := db.First(&config).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			// 创建新配置
			config = model.SMTPConfig{
				Enabled:   req.Enabled,
				Host:      req.Host,
				Port:      req.Port,
				Username:  req.Username,
				Password:  req.Password,
				FromEmail: req.FromEmail,
				FromName:  req.FromName,
			}
			if err := db.Create(&config).Error; err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create SMTP config"})
				return
			}
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get SMTP config"})
			return
		}
	} else {
		// 更新现有配置
		config.Enabled = req.Enabled
		config.Host = req.Host
		config.Port = req.Port
		config.Username = req.Username
		config.FromEmail = req.FromEmail
		config.FromName = req.FromName

		// 只有当提供了新密码时才更新密码
		if req.Password != "" {
			config.Password = req.Password
		}

		if err := db.Save(&config).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update SMTP config"})
			return
		}
	}

	// 返回配置，但不包含密码
	config.Password = ""
	c.JSON(http.StatusOK, config)
}
