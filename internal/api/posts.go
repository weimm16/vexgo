package api

import (
	"net/http"
	"strconv"

	"go-cms/internal/cms"

	"github.com/gin-gonic/gin"
)

type PostHandler struct {
	service *cms.ContentService
}

func NewPostHandler(s *cms.ContentService) *PostHandler {
	return &PostHandler{service: s}
}

type CreatePostRequest struct {
	Title      string `json:"title" binding:"required"`
	Content    string `json:"content"`
	Status     string `json:"status"`
	MetaTitle  string `json:"meta_title"`
	MetaDesc   string `json:"meta_description"`
	OGImage    string `json:"og_image"`
	Tags       string `json:"tags"`
	CategoryID JSONUint `json:"category_id"`
}

func (h *PostHandler) Create(c *gin.Context) {
	var req CreatePostRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	input := cms.CreatePostInput{
		Title:      req.Title,
		Content:    req.Content,
		Status:     req.Status,
		MetaTitle:  req.MetaTitle,
		MetaDesc:   req.MetaDesc,
		OGImage:    req.OGImage,
		Tags:       req.Tags,
	}
	if req.CategoryID.Valid {
		input.CategoryID = req.CategoryID.Value
	}
	post, err := h.service.CreatePost(input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, post)
}

func (h *PostHandler) GetBySlug(c *gin.Context) {
	slug := c.Param("slug")
	post, err := h.service.GetPostBySlug(slug)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "post not found"})
		return
	}
	c.JSON(http.StatusOK, post)
}

func (h *PostHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	posts, total, err := h.service.ListPosts(page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": posts, "total": total, "page": page, "page_size": pageSize})
}

func (h *PostHandler) Update(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)
	var req CreatePostRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	input := cms.CreatePostInput{
		Title:      req.Title,
		Content:    req.Content,
		Status:     req.Status,
		MetaTitle:  req.MetaTitle,
		MetaDesc:   req.MetaDesc,
		OGImage:    req.OGImage,
		Tags:       req.Tags,
	}
	if req.CategoryID.Valid {
		input.CategoryID = req.CategoryID.Value
	}
	post, err := h.service.UpdatePost(uint(id), input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, post)
}

func (h *PostHandler) Delete(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)
	if err := h.service.DeletePost(uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}
