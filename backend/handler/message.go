package handler

import (
	"net/http"
	"strconv"
	"vexgo/backend/model"

	"github.com/gin-gonic/gin"
)

// GetMessages 获取消息列表
func GetMessages(c *gin.Context) {
	userID, _ := c.Get("userID")
	uid := userID.(uint)

	// 分页参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	offset := (page - 1) * limit

	// 筛选参数
	messageType := c.Query("type")
	isRead := c.Query("is_read")

	// 构建查询
	query := db.Model(&model.Notification{}).Where("user_id = ?", uid)

	// 类型筛选
	if messageType != "" {
		query = query.Where("type = ?", messageType)
	}

	// 已读状态筛选
	if isRead != "" {
		readStatus := isRead == "true"
		query = query.Where("is_read = ?", readStatus)
	}

	// 计算总数
	var total int64
	query.Count(&total)

	// 查询消息
	var notifications []model.Notification
	query.Order("created_at DESC").Offset(offset).Limit(limit).Find(&notifications)

	c.JSON(http.StatusOK, gin.H{
		"notifications": notifications,
		"pagination": gin.H{
			"total":      total,
			"page":       page,
			"limit":      limit,
			"totalPages": (total + int64(limit) - 1) / int64(limit),
		},
	})
}

// MarkAsRead 标记消息为已读
func MarkAsRead(c *gin.Context) {
	userID, _ := c.Get("userID")
	uid := userID.(uint)

	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid message ID"})
		return
	}

	// 直接更新消息状态，避免先查询再更新的问题
	result := db.Model(&model.Notification{}).Where("id = ? AND user_id = ?", id, uid).Update("is_read", true)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to mark message as read"})
		return
	}
	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Message not found or not updated"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Message marked as read"})
}

// MarkAllAsRead 标记所有消息为已读
func MarkAllAsRead(c *gin.Context) {
	userID, _ := c.Get("userID")
	uid := userID.(uint)

	// 标记所有消息为已读
	result := db.Model(&model.Notification{}).Where("user_id = ?", uid).Update("is_read", true)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to mark all messages as read"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "All messages marked as read"})
}

// DeleteMessage 删除消息
func DeleteMessage(c *gin.Context) {
	userID, _ := c.Get("userID")
	uid := userID.(uint)

	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid message ID"})
		return
	}

	// 直接删除消息，避免先查询再删除的问题
	result := db.Where("id = ? AND user_id = ?", id, uid).Delete(&model.Notification{})
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete message"})
		return
	}
	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Message not found or not deleted"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Message deleted"})
}

// GetUnreadCount 获取未读消息数量
func GetUnreadCount(c *gin.Context) {
	userID, _ := c.Get("userID")
	uid := userID.(uint)

	// 计算未读消息数量
	var count int64
	db.Model(&model.Notification{}).Where("user_id = ? AND is_read = ?", uid, false).Count(&count)

	c.JSON(http.StatusOK, gin.H{"unreadCount": count})
}

// CreateNotification 创建通知
func CreateNotification(userID uint, notificationType string, title string, content string, relatedID string, relatedType string) error {
	notification := model.Notification{
		UserID:      userID,
		Type:        notificationType,
		Title:       title,
		Content:     content,
		RelatedID:   relatedID,
		RelatedType: relatedType,
		IsRead:      false,
	}

	return db.Create(&notification).Error
}