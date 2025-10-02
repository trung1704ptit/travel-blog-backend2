package routes

import (
	"app/controllers"
	"app/middleware"

	"github.com/gin-gonic/gin"
)

type PostRouteController struct {
	postController *controllers.PostController
}

func NewRoutePostController(postController *controllers.PostController) PostRouteController {
	return PostRouteController{postController}
}

func (pc *PostRouteController) PostRoute(rg *gin.RouterGroup) {
	router := rg.Group("posts")

	// Public routes (no authentication required)
	router.GET("", pc.postController.FindPosts)
	router.GET("/search", pc.postController.SearchPosts)

	// Most blogs/CMSs use slug as the primary identifier
	router.GET("/:slug", pc.postController.FindPostBySlug)    // Primary: by slug
	router.GET("/id/:postId", pc.postController.FindPostById) // Alternative: by UUID

	// Protected routes (authentication required)
	router.Use(middleware.DeserializeUser())
	router.POST("", pc.postController.CreatePost)
	router.PATCH("/:postId", pc.postController.UpdatePost)
	router.DELETE("/:postId", pc.postController.DeletePost)
	router.POST("/:postId/like", pc.postController.LikePost)
}
