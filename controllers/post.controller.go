package controllers

import (
	"net/http"
	"strconv"

	"app/models"
	"app/services"

	"github.com/gin-gonic/gin"
)

type PostController struct {
	postService *services.PostService
}

func NewPostController(postService *services.PostService) *PostController {
	return &PostController{postService: postService}
}

// CreatePost handles post creation
func (pc *PostController) CreatePost(ctx *gin.Context) {
	currentUser := ctx.MustGet("currentUser").(models.User)
	var payload *models.CreatePostRequest

	if err := ctx.ShouldBindJSON(&payload); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"status": "fail", "message": err.Error()})
		return
	}

	newPost, err := pc.postService.CreatePost(payload, currentUser.ID)
	if err != nil {
		switch err.Error() {
		case "post with that title or slug already exists":
			ctx.JSON(http.StatusConflict, gin.H{"status": "fail", "message": err.Error()})
		default:
			ctx.JSON(http.StatusBadGateway, gin.H{"status": "error", "message": err.Error()})
		}
		return
	}

	ctx.JSON(http.StatusCreated, gin.H{"status": "success", "data": newPost})
}

// UpdatePost handles post updates
func (pc *PostController) UpdatePost(ctx *gin.Context) {
	postID := ctx.Param("postId")
	currentUser := ctx.MustGet("currentUser").(models.User)

	var payload *models.UpdatePostRequest
	if err := ctx.ShouldBindJSON(&payload); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"status": "fail", "message": err.Error()})
		return
	}

	updatedPost, err := pc.postService.UpdatePost(postID, payload, currentUser.ID)
	if err != nil {
		switch err.Error() {
		case "post not found":
			ctx.JSON(http.StatusNotFound, gin.H{"status": "fail", "message": err.Error()})
		default:
			ctx.JSON(http.StatusBadGateway, gin.H{"status": "fail", "message": err.Error()})
		}
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"status": "success", "data": updatedPost})
}

// FindPostById handles retrieving a single post by ID
func (pc *PostController) FindPostById(ctx *gin.Context) {
	postID := ctx.Param("postId")

	post, err := pc.postService.FindPostByID(postID)
	if err != nil {
		switch err.Error() {
		case "post not found":
			ctx.JSON(http.StatusNotFound, gin.H{"status": "fail", "message": err.Error()})
		default:
			ctx.JSON(http.StatusBadGateway, gin.H{"status": "error", "message": err.Error()})
		}
		return
	}

	// Optionally increment views
	go pc.postService.IncrementViews(postID)

	ctx.JSON(http.StatusOK, gin.H{"status": "success", "data": post})
}

// FindPostBySlug handles retrieving a single post by slug
func (pc *PostController) FindPostBySlug(ctx *gin.Context) {
	slug := ctx.Param("slug")

	post, err := pc.postService.FindPostBySlug(slug)
	if err != nil {
		switch err.Error() {
		case "post not found":
			ctx.JSON(http.StatusNotFound, gin.H{"status": "fail", "message": err.Error()})
		default:
			ctx.JSON(http.StatusBadGateway, gin.H{"status": "error", "message": err.Error()})
		}
		return
	}

	// Optionally increment views
	go pc.postService.IncrementViews(post.ID.String())

	ctx.JSON(http.StatusOK, gin.H{"status": "success", "data": post})
}

// FindPosts handles retrieving paginated posts with filters
func (pc *PostController) FindPosts(ctx *gin.Context) {
	page, err := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	if err != nil || page < 1 {
		page = 1
	}

	limit, err := strconv.Atoi(ctx.DefaultQuery("limit", "10"))
	if err != nil || limit < 1 {
		limit = 10
	}

	// Get published filter
	var published *bool
	if publishedStr := ctx.Query("published"); publishedStr != "" {
		val := publishedStr == "true"
		published = &val
	}

	// Get category filter
	var categoryID *string
	if catID := ctx.Query("category_id"); catID != "" {
		categoryID = &catID
	}

	posts, total, err := pc.postService.FindPosts(page, limit, published, categoryID)
	if err != nil {
		ctx.JSON(http.StatusBadGateway, gin.H{"status": "error", "message": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   posts,
		"pagination": gin.H{
			"page":  page,
			"limit": limit,
			"total": total,
		},
	})
}

// SearchPosts handles searching posts
func (pc *PostController) SearchPosts(ctx *gin.Context) {
	query := ctx.Query("q")
	if query == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"status": "fail", "message": "search query is required"})
		return
	}

	page, err := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	if err != nil || page < 1 {
		page = 1
	}

	limit, err := strconv.Atoi(ctx.DefaultQuery("limit", "10"))
	if err != nil || limit < 1 {
		limit = 10
	}

	posts, total, err := pc.postService.SearchPosts(query, page, limit)
	if err != nil {
		ctx.JSON(http.StatusBadGateway, gin.H{"status": "error", "message": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   posts,
		"pagination": gin.H{
			"page":  page,
			"limit": limit,
			"total": total,
		},
	})
}

// DeletePost handles post deletion
func (pc *PostController) DeletePost(ctx *gin.Context) {
	postID := ctx.Param("postId")

	err := pc.postService.DeletePost(postID)
	if err != nil {
		switch err.Error() {
		case "post not found":
			ctx.JSON(http.StatusNotFound, gin.H{"status": "fail", "message": err.Error()})
		default:
			ctx.JSON(http.StatusBadGateway, gin.H{"status": "error", "message": err.Error()})
		}
		return
	}

	ctx.JSON(http.StatusNoContent, nil)
}

// LikePost handles incrementing post likes
func (pc *PostController) LikePost(ctx *gin.Context) {
	postID := ctx.Param("postId")

	err := pc.postService.IncrementLikes(postID)
	if err != nil {
		ctx.JSON(http.StatusBadGateway, gin.H{"status": "error", "message": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"status": "success", "message": "Post liked successfully"})
}
