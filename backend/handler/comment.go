package handler

import (
    "blog-system/backend/model"
    "net/http"
    "strconv"

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
    // 支持前端传入 postId 为数字或字符串
    var req struct {
        PostID   interface{} `json:"postId" binding:"required"`
        Content  string      `json:"content" binding:"required"`
        ParentID *uint       `json:"parentId"`
    }

    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    // 解析 PostID 为 uint
    var postID uint
    switch v := req.PostID.(type) {
    case float64:
        postID = uint(v)
    case string:
        if id64, err := strconv.ParseUint(v, 10, 64); err == nil {
            postID = uint(id64)
        }
    case int:
        postID = uint(v)
    case uint:
        postID = v
    default:
        // 如果无法解析，返回错误
        c.JSON(http.StatusBadRequest, gin.H{"error": "postId 类型不合法"})
        return
    }

    uid, _ := c.Get("userID")
    userID, ok := uid.(uint)
    if !ok {
        // 拒绝未登录请求
        c.JSON(http.StatusUnauthorized, gin.H{"error": "未登录"})
        return
    }

    comment := model.Comment{
        PostID:   postID,
        Content:  req.Content,
        ParentID: req.ParentID,
        UserID:   userID,
    }

    if err := db.Create(&comment).Error; err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "创建评论失败"})
        return
    }

    // 返回创建的评论及更新后的评论计数
    var count int64
    db.Model(&model.Comment{}).Where("post_id = ?", postID).Count(&count)

    // 预加载作者信息
    db.Preload("User").First(&comment, comment.ID)

    c.JSON(http.StatusCreated, gin.H{"message": "评论创建成功", "comment": comment, "commentsCount": count})
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
