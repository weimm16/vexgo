package handler

import (
	"net/http"

	"blog-system/backend/model"
	"blog-system/backend/utils"

	"github.com/gin-gonic/gin"
)

// VerifyEmail 验证邮箱
func VerifyEmail(c *gin.Context) {
	token := c.Query("token")
	if token == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "验证令牌不能为空"})
		return
	}

	mailer := utils.NewMailer(db)
	if err := mailer.VerifyEmail(token); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 重定向到成功页面或返回成功消息
	c.JSON(http.StatusOK, gin.H{
		"message": "邮箱验证成功！您现在可以登录了。",
	})
}

// GetVerificationStatus 获取当前用户的邮箱验证状态
func GetVerificationStatus(c *gin.Context) {
	userContext, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未授权"})
		return
	}

	if userMap, ok := userContext.(map[string]interface{}); ok {
		if userID, ok := userMap["id"].(float64); ok {
			var user model.User
			if err := db.First(&user, uint(userID)).Error; err != nil {
				c.JSON(http.StatusNotFound, gin.H{"error": "用户不存在"})
				return
			}

			c.JSON(http.StatusOK, gin.H{
				"email_verified": user.EmailVerified,
				"email":          user.Email,
			})
			return
		}
	}

	c.JSON(http.StatusInternalServerError, gin.H{"error": "获取用户信息失败"})
}

// ResendVerificationEmail 重新发送验证邮件
func ResendVerificationEmail(c *gin.Context) {
	userContext, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未授权"})
		return
	}

	if userMap, ok := userContext.(map[string]interface{}); ok {
		if userID, ok := userMap["id"].(float64); ok {
			var user model.User
			if err := db.First(&user, uint(userID)).Error; err != nil {
				c.JSON(http.StatusNotFound, gin.H{"error": "用户不存在"})
				return
			}

			if user.EmailVerified {
				c.JSON(http.StatusBadRequest, gin.H{"error": "邮箱已验证"})
				return
			}

			mailer := utils.NewMailer(db)
			enabled, err := mailer.IsEmailEnabled()
			if err != nil || !enabled {
				c.JSON(http.StatusServiceUnavailable, gin.H{"error": "邮件服务未启用"})
				return
			}

			// 生成新的验证令牌
			token, err := mailer.GenerateVerificationToken(user.ID)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "生成验证令牌失败"})
				return
			}

			// 构建验证链接
			verificationLink := c.Request.Host + "/verify-email?token=" + token
			if err := mailer.SendVerificationEmail(user.Email, user.Username, verificationLink); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "发送验证邮件失败"})
				return
			}

			c.JSON(http.StatusOK, gin.H{
				"message": "验证邮件已重新发送，请检查您的邮箱",
			})
			return
		}
	}

	c.JSON(http.StatusInternalServerError, gin.H{"error": "获取用户信息失败"})
}
