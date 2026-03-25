package handler

import (
	"fmt"
	"math"
	"net/http"
	"time"

	"github.com/sirupsen/logrus"

	"vexgo/backend/model"
	"vexgo/backend/utils"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

func Register(c *gin.Context) {
	logrus.Info("User registration attempt started")

	var req struct {
		Email        string `json:"email" binding:"required,email"`
		Password     string `json:"password" binding:"required"`
		Username     string `json:"username" binding:"required"`
		CaptchaID    string `json:"captcha_id"`
		CaptchaToken string `json:"captcha_token"`
		CaptchaX     int    `json:"captcha_x"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		logrus.WithError(err).Warn("Failed to bind registration request JSON")
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	logrus.WithFields(logrus.Fields{
		"email":    req.Email,
		"username": req.Username,
	}).Debug("Registration request parsed successfully")

	// Check if registration is allowed
	var settings model.GeneralSettings
	if err := db.First(&settings).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			// Allow registration by default
			settings.RegistrationEnabled = true
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check registration settings"})
			return
		}
	}

	if !settings.RegistrationEnabled {
		c.JSON(http.StatusForbidden, gin.H{"error": "Registration is disabled, please contact administrator"})
		return
	}

	// Check if captcha verification is enabled
	captchaEnabled, err := IsCaptchaEnabled()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check captcha settings"})
		return
	}

	// If captcha verification is enabled, verify captcha
	if captchaEnabled {
		logrus.Debug("Captcha verification enabled, validating user captcha")
		if req.CaptchaID == "" || req.CaptchaToken == "" || req.CaptchaX == 0 {
			logrus.WithFields(logrus.Fields{
				"email":     req.Email,
				"captchaID": req.CaptchaID,
				"captchaX":  req.CaptchaX,
			}).Warn("Captcha verification failed: missing required fields")
			c.JSON(http.StatusBadRequest, gin.H{"error": "Please complete captcha verification"})
			return
		}
		// Query captcha
		var captcha model.Captcha
		if err := db.Where("id = ? AND token = ?", req.CaptchaID, req.CaptchaToken).First(&captcha).Error; err != nil {
			logrus.WithFields(logrus.Fields{
				"captchaID": req.CaptchaID,
				"email":     req.Email,
			}).Warn("Captcha verification failed: captcha not found or invalid token")
			c.JSON(http.StatusNotFound, gin.H{"error": "Captcha does not exist or has expired"})
			return
		}

		// Check if expired
		if time.Now().After(captcha.ExpiresAt) {
			logrus.WithFields(logrus.Fields{
				"captchaID": req.CaptchaID,
				"expiresAt": captcha.ExpiresAt,
				"email":     req.Email,
			}).Warn("Captcha verification failed: captcha expired")
			c.JSON(http.StatusBadRequest, gin.H{"error": "Captcha has expired"})
			return
		}

		// Verify position (allow certain tolerance)
		tolerance := 5
		if math.Abs(float64(req.CaptchaX-captcha.X)) > float64(tolerance) {
			logrus.WithFields(logrus.Fields{
				"captchaID": req.CaptchaID,
				"userX":     req.CaptchaX,
				"correctX":  captcha.X,
				"tolerance": tolerance,
				"email":     req.Email,
			}).Warn("Captcha verification failed: incorrect position")
			c.JSON(http.StatusBadRequest, gin.H{"error": "Verification failed, please try again"})
			return
		}

		logrus.WithFields(logrus.Fields{
			"captchaID": req.CaptchaID,
			"email":     req.Email,
		}).Debug("Captcha verification passed")

		// If captcha has not been used yet, mark it as used
		if !captcha.Used {
			captcha.Used = true
			if err := db.Save(&captcha).Error; err != nil {
				logrus.WithFields(logrus.Fields{
					"captchaID": req.CaptchaID,
					"email":     req.Email,
				}).WithError(err).Error("Failed to mark captcha as used")
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Captcha verification failed"})
				return
			}
			logrus.WithField("captchaID", req.CaptchaID).Debug("Captcha marked as used")
		}
		// If captcha already used, pre-verification successful, pass directly
	}

	// Check if user already exists
	var existingUser model.User
	if err := db.Where("email = ?", req.Email).First(&existingUser).Error; err == nil {
		logrus.WithField("email", req.Email).Warn("Registration failed: user already exists")
		c.JSON(http.StatusConflict, gin.H{"error": "User already exists"})
		return
	}

	// Encrypt password
	logrus.WithField("email", req.Email).Debug("Starting password hashing")
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"email": req.Email,
		}).WithError(err).Error("Failed to hash password")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
		return
	}
	logrus.WithField("email", req.Email).Debug("Password hashed successfully")

	// Create new user
	newUser := model.User{
		Username:      req.Username,
		Email:         req.Email,
		Password:      string(hashedPassword),
		Role:          model.RoleGuest, // Default role is guest
		EmailVerified: false,
	}

	logrus.WithFields(logrus.Fields{
		"username": req.Username,
		"email":    req.Email,
		"role":     model.RoleGuest,
	}).Info("Creating new user")
	if err := db.Create(&newUser).Error; err != nil {
		logrus.WithFields(logrus.Fields{
			"username": req.Username,
			"email":    req.Email,
		}).WithError(err).Error("Failed to create user in database")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
		return
	}
	logrus.WithFields(logrus.Fields{
		"userID":   newUser.ID,
		"username": req.Username,
		"email":    req.Email,
	}).Info("User created successfully")

	// Check if SMTP is enabled, if so send verification email
	mailer := utils.NewMailer(db)
	enabled, err := mailer.IsEmailEnabled()
	if err != nil {
		logrus.WithField("email", newUser.Email).WithError(err).Warn("Failed to check if SMTP is enabled")
	} else if enabled {
		logrus.WithFields(logrus.Fields{
			"userID":   newUser.ID,
			"username": newUser.Username,
			"email":    newUser.Email,
		}).Info("Email verification enabled, generating verification token")

		// Generate verification token
		token, err := mailer.GenerateVerificationToken(newUser.ID)
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"userID":   newUser.ID,
				"username": newUser.Username,
				"email":    newUser.Email,
			}).WithError(err).Error("Failed to generate verification token")
		} else {
			logrus.WithFields(logrus.Fields{
				"userID":   newUser.ID,
				"username": newUser.Username,
				"email":    newUser.Email,
			}).Debug("Verification token generated successfully")

			// Build verification link - use request protocol and hostname
			protocol := "http"
			if c.Request.TLS != nil || c.GetHeader("X-Forwarded-Proto") == "https" {
				protocol = "https"
			}
			host := c.Request.Host
			verificationLink := fmt.Sprintf("%s://%s/verify-email?token=%s", protocol, host, token)

			logrus.WithFields(logrus.Fields{
				"userID":   newUser.ID,
				"username": newUser.Username,
				"email":    newUser.Email,
				"protocol": protocol,
				"host":     host,
			}).Debug("Sending verification email")

			// Send verification email
			if err := mailer.SendVerificationEmail(newUser.Email, newUser.Username, verificationLink); err != nil {
				logrus.WithFields(logrus.Fields{
					"userID":   newUser.ID,
					"username": newUser.Username,
					"email":    newUser.Email,
				}).WithError(err).Error("Failed to send verification email")
			} else {
				logrus.WithFields(logrus.Fields{
					"userID":   newUser.ID,
					"username": newUser.Username,
					"email":    newUser.Email,
				}).Info("Verification email sent successfully")

				c.JSON(http.StatusCreated, gin.H{
					"message":               "Registration successful! Please verify your email address before logging in. Check your inbox and click the verification link.",
					"user":                  newUser,
					"email_verified":        false,
					"requires_verification": true,
				})
				return
			}
		}
	} else {
		logrus.WithField("email", newUser.Email).Info("SMTP not enabled, skipping email verification")
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":               "Registration successful",
		"user":                  newUser,
		"email_verified":        newUser.EmailVerified,
		"requires_verification": false,
	})
}
