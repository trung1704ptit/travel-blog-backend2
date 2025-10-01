package controllers

import (
	"net/http"
	"strconv"

	"app/models"
	"app/services"

	"github.com/gin-gonic/gin"
)

type CategoryController struct {
	categoryService *services.CategoryService
}

func NewCategoryController(categoryService *services.CategoryService) *CategoryController {
	return &CategoryController{categoryService: categoryService}
}

// CreateCategory handles category creation
func (cc *CategoryController) CreateCategory(ctx *gin.Context) {
	var payload *models.CreateCategoryRequest

	if err := ctx.ShouldBindJSON(&payload); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"status": "fail", "message": err.Error()})
		return
	}

	category, err := cc.categoryService.CreateCategory(payload)
	if err != nil {
		switch err.Error() {
		case "category with that slug already exists":
			ctx.JSON(http.StatusConflict, gin.H{"status": "fail", "message": err.Error()})
		case "parent category not found":
			ctx.JSON(http.StatusNotFound, gin.H{"status": "fail", "message": err.Error()})
		default:
			ctx.JSON(http.StatusBadGateway, gin.H{"status": "error", "message": err.Error()})
		}
		return
	}

	ctx.JSON(http.StatusCreated, gin.H{"status": "success", "data": category})
}

// UpdateCategory handles category updates
func (cc *CategoryController) UpdateCategory(ctx *gin.Context) {
	categoryID := ctx.Param("categoryId")

	var payload *models.UpdateCategoryRequest
	if err := ctx.ShouldBindJSON(&payload); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"status": "fail", "message": err.Error()})
		return
	}

	category, err := cc.categoryService.UpdateCategory(categoryID, payload)
	if err != nil {
		switch err.Error() {
		case "category not found":
			ctx.JSON(http.StatusNotFound, gin.H{"status": "fail", "message": err.Error()})
		case "parent category not found":
			ctx.JSON(http.StatusNotFound, gin.H{"status": "fail", "message": err.Error()})
		case "category cannot be its own parent":
			ctx.JSON(http.StatusBadRequest, gin.H{"status": "fail", "message": err.Error()})
		case "category with that slug already exists":
			ctx.JSON(http.StatusConflict, gin.H{"status": "fail", "message": err.Error()})
		default:
			ctx.JSON(http.StatusBadGateway, gin.H{"status": "error", "message": err.Error()})
		}
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"status": "success", "data": category})
}

// FindCategoryById handles retrieving a single category by ID
func (cc *CategoryController) FindCategoryById(ctx *gin.Context) {
	categoryID := ctx.Param("categoryId")

	category, err := cc.categoryService.FindCategoryByID(categoryID)
	if err != nil {
		switch err.Error() {
		case "category not found":
			ctx.JSON(http.StatusNotFound, gin.H{"status": "fail", "message": err.Error()})
		default:
			ctx.JSON(http.StatusBadGateway, gin.H{"status": "error", "message": err.Error()})
		}
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"status": "success", "data": category})
}

// FindCategoryBySlug handles retrieving a single category by slug
func (cc *CategoryController) FindCategoryBySlug(ctx *gin.Context) {
	slug := ctx.Param("slug")

	category, err := cc.categoryService.FindCategoryBySlug(slug)
	if err != nil {
		switch err.Error() {
		case "category not found":
			ctx.JSON(http.StatusNotFound, gin.H{"status": "fail", "message": err.Error()})
		default:
			ctx.JSON(http.StatusBadGateway, gin.H{"status": "error", "message": err.Error()})
		}
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"status": "success", "data": category})
}

// FindCategories handles retrieving all categories with optional filters
func (cc *CategoryController) FindCategories(ctx *gin.Context) {
	// Get filters from query params
	var parentID *string
	if parentIDParam := ctx.Query("parent_id"); parentIDParam != "" {
		parentID = &parentIDParam
	}

	var level *int
	if levelParam := ctx.Query("level"); levelParam != "" {
		levelInt, err := strconv.Atoi(levelParam)
		if err == nil {
			level = &levelInt
		}
	}

	categories, err := cc.categoryService.FindCategories(parentID, level)
	if err != nil {
		ctx.JSON(http.StatusBadGateway, gin.H{"status": "error", "message": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"results": len(categories),
		"data":    categories,
	})
}

// FindRootCategories handles retrieving all root categories
func (cc *CategoryController) FindRootCategories(ctx *gin.Context) {
	categories, err := cc.categoryService.FindRootCategories()
	if err != nil {
		ctx.JSON(http.StatusBadGateway, gin.H{"status": "error", "message": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"results": len(categories),
		"data":    categories,
	})
}

// FindCategoryTree handles retrieving hierarchical category tree
func (cc *CategoryController) FindCategoryTree(ctx *gin.Context) {
	categories, err := cc.categoryService.FindCategoryTree()
	if err != nil {
		ctx.JSON(http.StatusBadGateway, gin.H{"status": "error", "message": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   categories,
	})
}

// DeleteCategory handles category deletion
func (cc *CategoryController) DeleteCategory(ctx *gin.Context) {
	categoryID := ctx.Param("categoryId")

	// Get deleteChildren flag from query
	deleteChildren := ctx.DefaultQuery("deleteChildren", "false") == "true"

	err := cc.categoryService.DeleteCategory(categoryID, deleteChildren)
	if err != nil {
		switch err.Error() {
		case "category not found":
			ctx.JSON(http.StatusNotFound, gin.H{"status": "fail", "message": err.Error()})
		default:
			if err.Error() == "category has children, cannot delete (use deleteChildren=true to force)" {
				ctx.JSON(http.StatusBadRequest, gin.H{"status": "fail", "message": err.Error()})
			} else {
				ctx.JSON(http.StatusBadGateway, gin.H{"status": "error", "message": err.Error()})
			}
		}
		return
	}

	ctx.JSON(http.StatusNoContent, nil)
}

// GetCategoryPostCount handles retrieving post count for a category
func (cc *CategoryController) GetCategoryPostCount(ctx *gin.Context) {
	categoryID := ctx.Param("categoryId")

	count, err := cc.categoryService.GetCategoryPostCount(categoryID)
	if err != nil {
		ctx.JSON(http.StatusBadGateway, gin.H{"status": "error", "message": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data": gin.H{
			"category_id": categoryID,
			"post_count":  count,
		},
	})
}

// SearchCategories handles searching categories
func (cc *CategoryController) SearchCategories(ctx *gin.Context) {
	query := ctx.Query("q")
	if query == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"status": "fail", "message": "search query is required"})
		return
	}

	categories, err := cc.categoryService.SearchCategories(query)
	if err != nil {
		ctx.JSON(http.StatusBadGateway, gin.H{"status": "error", "message": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"results": len(categories),
		"data":    categories,
	})
}
