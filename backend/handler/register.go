package handler

import (
	"fmt"
	"log"
	"net/http"

	"blog-system/backend/model"
	"blog-system/backend/utils"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

func Register(c *gin.Context) {
	var req struct {
		Email    string `json:"email" binding:"required,email"`
		Password string `json:"password" binding:"required"`
		Username string `json:"username" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 检查用户是否已存在
	var existingUser model.User
	if err := db.Where("email = ?", req.Email).First(&existingUser).Error; err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "User already exists"})
		return
	}

	// 密码加密
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
		return
	}

	// 创建新用户
	newUser := model.User{
		Username:      req.Username,
		Email:         req.Email,
		Password:      string(hashedPassword),
		Role:          "contributor", // 默认角色为投稿者
		EmailVerified: false,
	}

	if err := db.Create(&newUser).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
		return
	}

	// 检查是否启用了 SMTP，如果启用则发送验证邮件
	mailer := utils.NewMailer(db)
	enabled, err := mailer.IsEmailEnabled()
	if err == nil && enabled {
		// 生成验证令牌
		token, err := mailer.GenerateVerificationToken(newUser.ID)
		if err != nil {
			log.Printf("生成验证令牌失败: %v", err)
		} else {
			// 构建验证链接
			verificationLink := fmt.Sprintf("%s/verify-email?token=%s", c.Request.Host, token)

			// 发送验证邮件
			if err := mailer.SendVerificationEmail(newUser.Email, newUser.Username, verificationLink); err != nil {
				log.Printf("发送验证邮件失败: %v", err)
			} else {
				c.JSON(http.StatusCreated, gin.H{
					"message": "注册成功，请检查您的邮箱完成验证",
					"user":    newUser,
				})
				return
			}
		}
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "注册成功",
		"user":    newUser,
	})
}
