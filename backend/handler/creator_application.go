package handler

import (
	"fmt"
	"net/http"
	"strconv"

	"vexgo/backend/model"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// ApplyForCreator handles creator application submission
func ApplyForCreator(c *gin.Context) {
	// Get current user information from context
	currentUserInterface, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "No user information provided"})
		return
	}

	// Convert user information from map to model.User
	userMap, ok := currentUserInterface.(map[string]interface{})
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "User information format error"})
		return
	}

	currentUser := model.User{
		ID:       userMap["id"].(uint),
		Username: userMap["username"].(string),
		Role:     userMap["role"].(string),
	}

	// Only guest and contributor users can apply for role upgrade
	if currentUser.Role != model.RoleGuest && currentUser.Role != model.RoleContributor {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Only guest and contributor users can apply for role upgrade"})
		return
	}

	// Parse request parameters
	var req struct {
		Reason string `json:"reason"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Check if user already has a pending application
	var existingApplication model.CreatorApplication
	if err := db.Where("user_id = ? AND status = ?", currentUser.ID, model.CreatorApplicationStatusPending).First(&existingApplication).Error; err == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "You already have a pending application"})
		return
	}

	// Create new application
	application := model.CreatorApplication{
		UserID: currentUser.ID,
		Status: model.CreatorApplicationStatusPending,
		Reason: req.Reason,
	}

	if err := db.Create(&application).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create application"})
		return
	}

	// Determine target role based on current role
	targetRole := ""
	if currentUser.Role == model.RoleGuest {
		targetRole = "contributor"
	} else if currentUser.Role == model.RoleContributor {
		targetRole = "author"
	}

	// Send notification to admins and super admins
	var admins []model.User
	if err := db.Where("role IN ?", []string{model.RoleAdmin, model.RoleSuperAdmin}).Find(&admins).Error; err == nil {
		for _, admin := range admins {
			CreateNotification(
				admin.ID,
				"role",
				"New Role Application",
				fmt.Sprintf("User %s has applied for %s role", currentUser.Username, targetRole),
				fmt.Sprintf("%d", application.ID),
				"creator_application",
			)
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"message":       "Application submitted successfully",
		"applicationId": application.ID,
	})
}

// GetCreatorApplications gets creator applications for admin review
func GetCreatorApplications(c *gin.Context) {
	// Get current user information from context
	currentUserInterface, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "No user information provided"})
		return
	}

	// Convert user information from map to model.User
	userMap, ok := currentUserInterface.(map[string]interface{})
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "User information format error"})
		return
	}

	currentUser := model.User{
		ID:   userMap["id"].(uint),
		Role: userMap["role"].(string),
	}

	// Only admins and super admins can access this endpoint
	if currentUser.Role != model.RoleAdmin && currentUser.Role != model.RoleSuperAdmin {
		c.JSON(http.StatusForbidden, gin.H{"error": "No permission to access creator applications"})
		return
	}

	// Pagination parameters
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	status := c.DefaultQuery("status", "pending")

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}

	var applications []model.CreatorApplication
	var total int64

	// Query applications with user information
	query := db.Model(&model.CreatorApplication{}).Preload("User")

	// Filter by status
	if status != "" {
		query = query.Where("status = ?", status)
	}

	// Count total
	query.Count(&total)

	// Paginated query
	query.Offset((page - 1) * limit).
		Limit(limit).
		Order("created_at DESC").
		Find(&applications)

	totalPages := int((total + int64(limit) - 1) / int64(limit))
	if totalPages == 0 {
		totalPages = 1
	}

	// Format response
	var response []map[string]interface{}
	for _, app := range applications {
		response = append(response, map[string]interface{}{
			"id":          app.ID,
			"userId":      app.UserID,
			"username":    app.User.Username,
			"email":       app.User.Email,
			"currentRole": app.User.Role,
			"status":      app.Status,
			"reason":      app.Reason,
			"createdAt":   app.CreatedAt.Format("2006-01-02T15:04:05Z"),
			"updatedAt":   app.UpdatedAt.Format("2006-01-02T15:04:05Z"),
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"applications": response,
		"pagination": gin.H{
			"total":      total,
			"page":       page,
			"limit":      limit,
			"totalPages": totalPages,
		},
	})
}

// ReviewCreatorApplication handles creator application review
func ReviewCreatorApplication(c *gin.Context) {
	// Get current user information from context
	currentUserInterface, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "No user information provided"})
		return
	}

	// Convert user information from map to model.User
	userMap, ok := currentUserInterface.(map[string]interface{})
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "User information format error"})
		return
	}

	currentUser := model.User{
		ID:   userMap["id"].(uint),
		Role: userMap["role"].(string),
	}

	// Only admins and super admins can review applications
	if currentUser.Role != model.RoleAdmin && currentUser.Role != model.RoleSuperAdmin {
		c.JSON(http.StatusForbidden, gin.H{"error": "No permission to review creator applications"})
		return
	}

	// Get application ID
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid application ID"})
		return
	}

	var application model.CreatorApplication

	// Find application with user information
	if err := db.Preload("User").First(&application, uint(id)).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Application does not exist"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to query application"})
		return
	}

	// Check if application is still pending
	if application.Status != model.CreatorApplicationStatusPending {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Application has already been processed"})
		return
	}

	// Parse request parameters
	var req struct {
		Action string `json:"action" binding:"required"`
		Reason string `json:"reason"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate action
	if req.Action != "approve" && req.Action != "reject" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid action"})
		return
	}

	// Update application status
	if req.Action == "approve" {
		application.Status = model.CreatorApplicationStatusApproved
		// Update user role based on current role
		if application.User.Role == model.RoleGuest {
			application.User.Role = model.RoleContributor
		} else if application.User.Role == model.RoleContributor {
			application.User.Role = model.RoleAuthor
		}
		if err := db.Model(&application.User).Select("Role").Updates(&application.User).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user role"})
			return
		}
	} else {
		application.Status = model.CreatorApplicationStatusRejected
	}

	application.ReviewerID = &currentUser.ID
	application.ReviewReason = req.Reason

	if err := db.Save(&application).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update application"})
		return
	}

	// Send notification to the applicant
	var notificationTitle, notificationContent string
	if req.Action == "approve" {
		if application.User.Role == model.RoleAuthor {
			notificationTitle = "Author Application Approved"
			notificationContent = "Your author application has been approved. You are now an author."
		} else {
			notificationTitle = "Contributor Application Approved"
			notificationContent = "Your contributor application has been approved. You are now a contributor."
		}
	} else {
		notificationTitle = "Role Application Rejected"
		notificationContent = "Your role application has been rejected."
		if req.Reason != "" {
			notificationContent += " Reason: " + req.Reason
		}
	}

	CreateNotification(
		application.UserID,
		"role",
		notificationTitle,
		notificationContent,
		"",
		"",
	)

	c.JSON(http.StatusOK, gin.H{
		"message": "Application reviewed successfully",
	})
}
