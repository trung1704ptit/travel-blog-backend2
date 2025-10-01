package routes

import (
	"app/controllers"
	"app/middleware"

	"github.com/gin-gonic/gin"
)

type CategoryRouteController struct {
	categoryController *controllers.CategoryController
}

func NewCategoryRouteController(categoryController *controllers.CategoryController) CategoryRouteController {
	return CategoryRouteController{categoryController}
}

func (cc *CategoryRouteController) CategoryRoute(rg *gin.RouterGroup) {
	router := rg.Group("categories")

	// Public routes (no authentication required)
	router.GET("", cc.categoryController.FindCategories)
	router.GET("/roots", cc.categoryController.FindRootCategories)
	router.GET("/tree", cc.categoryController.FindCategoryTree)
	router.GET("/search", cc.categoryController.SearchCategories)

	// Explicit ID routes (with /id/ prefix to avoid conflicts)
	router.GET("/id/:categoryId", cc.categoryController.FindCategoryById)
	router.GET("/id/:categoryId/count", cc.categoryController.GetCategoryPostCount)

	// Wildcard route MUST be last (catches everything else)
	router.GET("/:slug", cc.categoryController.FindCategoryBySlug)

	// Protected routes (authentication required)
	router.Use(middleware.DeserializeUser())
	router.POST("", cc.categoryController.CreateCategory)
	router.PUT("/:categoryId", cc.categoryController.UpdateCategory)
	router.DELETE("/:categoryId", cc.categoryController.DeleteCategory)
}
