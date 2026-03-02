package handler

import (
	"blog-system/backend/model"
	"net/http"

	"github.com/gin-gonic/gin"
)

// 获取某篇文章的评论列表
func GetComments(c *gin.Context) {
    postID := c.Param("id")
    var comments []model.Comment
    db.Where("post_id = ?", postID).
        Preload("User").
        Find(&comments)

    c.JSON(http.StatusOK, gin.H{"comments": comments})
}

// 创建评论（需登录）
func CreateComment(c *gin.Context) {
    var req struct {
        PostID   uint   `json:"postId" binding:"required"`
        Content  string `json:"content" binding:"required"`
        ParentID *uint  `json:"parentId"`
    }
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    uid, _ := c.Get("userID")
    userID := uid.(uint)

    comment := model.Comment{
        PostID:   req.PostID,
        Content:  req.Content,
        ParentID: req.ParentID,
        UserID:   userID,
    }

    if err := db.Create(&comment).Error; err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "创建评论失败"})
        return
    }

    c.JSON(http.StatusCreated, gin.H{"message": "评论创建成功", "comment": comment})
}

// 删除评论（需登录，作者或管理员）
func DeleteComment(c *gin.Context) {
    id := c.Param("id")
    var comment model.Comment
    if err := db.First(&comment, id).Error; err != nil {
        c.JSON(http.StatusNotFound, gin.H{"error": "评论不存在"})
        return
    }

    uid, _ := c.Get("userID")
    userID := uid.(uint)
    var user model.User
    if err := db.First(&user, userID).Error; err == nil {
        if user.Role != "admin" && comment.UserID != userID {
            c.JSON(http.StatusForbidden, gin.H{"error": "无权删除该评论"})
            return
        }
    }

    db.Delete(&comment)
    c.JSON(http.StatusOK, gin.H{"message": "评论已删除"})
}
