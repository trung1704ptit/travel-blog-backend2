package services

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"app/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type CategoryService struct {
	DB *gorm.DB
}

func NewCategoryService(db *gorm.DB) *CategoryService {
	return &CategoryService{DB: db}
}

// CreateCategory creates a new category
func (s *CategoryService) CreateCategory(payload *models.CreateCategoryRequest) (*models.Category, error) {
	now := time.Now()

	// Validate parent if provided
	var level int
	var path string
	if payload.ParentID != nil {
		var parent models.Category
		if err := s.DB.First(&parent, "id = ?", payload.ParentID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, errors.New("parent category not found")
			}
			return nil, fmt.Errorf("failed to fetch parent category: %w", err)
		}
		level = parent.Level + 1
		if parent.Path != "" {
			path = parent.Path + "/" + payload.Slug
		} else {
			path = parent.Slug + "/" + payload.Slug
		}
	} else {
		level = 0
		path = payload.Slug
	}

	newCategory := models.Category{
		Name:        payload.Name,
		Slug:        payload.Slug,
		Description: payload.Description,
		Image:       payload.Image,
		ParentID:    payload.ParentID,
		Level:       level,
		Path:        path,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	result := s.DB.Create(&newCategory)
	if result.Error != nil {
		if strings.Contains(result.Error.Error(), "duplicate key") {
			return nil, errors.New("category with that slug already exists")
		}
		return nil, fmt.Errorf("failed to create category: %w", result.Error)
	}

	// Reload with parent
	s.DB.Preload("Parent").First(&newCategory, "id = ?", newCategory.ID)

	return &newCategory, nil
}

// UpdateCategory updates an existing category
func (s *CategoryService) UpdateCategory(categoryID string, payload *models.UpdateCategoryRequest) (*models.Category, error) {
	var existingCategory models.Category
	result := s.DB.First(&existingCategory, "id = ?", categoryID)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, errors.New("category not found")
		}
		return nil, fmt.Errorf("failed to fetch category: %w", result.Error)
	}

	now := time.Now()

	// Update fields if provided
	if payload.Name != nil {
		existingCategory.Name = *payload.Name
	}
	if payload.Slug != nil {
		existingCategory.Slug = *payload.Slug
		// Recalculate path if slug changes
		if existingCategory.ParentID != nil {
			var parent models.Category
			s.DB.First(&parent, "id = ?", existingCategory.ParentID)
			if parent.Path != "" {
				existingCategory.Path = parent.Path + "/" + *payload.Slug
			} else {
				existingCategory.Path = parent.Slug + "/" + *payload.Slug
			}
		} else {
			existingCategory.Path = *payload.Slug
		}
	}
	if payload.Description != nil {
		existingCategory.Description = *payload.Description
	}
	if payload.Image != nil {
		existingCategory.Image = *payload.Image
	}

	// Handle parent change
	if payload.ParentID != nil {
		if *payload.ParentID == uuid.Nil {
			// Remove parent (make root)
			existingCategory.ParentID = nil
			existingCategory.Level = 0
			existingCategory.Path = existingCategory.Slug
		} else {
			// Set new parent
			var parent models.Category
			if err := s.DB.First(&parent, "id = ?", payload.ParentID).Error; err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					return nil, errors.New("parent category not found")
				}
				return nil, fmt.Errorf("failed to fetch parent category: %w", err)
			}

			// Check for circular reference
			if parent.ID == existingCategory.ID {
				return nil, errors.New("category cannot be its own parent")
			}

			existingCategory.ParentID = payload.ParentID
			existingCategory.Level = parent.Level + 1
			if parent.Path != "" {
				existingCategory.Path = parent.Path + "/" + existingCategory.Slug
			} else {
				existingCategory.Path = parent.Slug + "/" + existingCategory.Slug
			}
		}
	}

	existingCategory.UpdatedAt = now

	result = s.DB.Save(&existingCategory)
	if result.Error != nil {
		if strings.Contains(result.Error.Error(), "duplicate key") {
			return nil, errors.New("category with that slug already exists")
		}
		return nil, fmt.Errorf("failed to update category: %w", result.Error)
	}

	// Reload with associations
	s.DB.Preload("Parent").Preload("Children").First(&existingCategory, "id = ?", existingCategory.ID)

	return &existingCategory, nil
}

// FindCategoryByID retrieves a category by ID
func (s *CategoryService) FindCategoryByID(categoryID string) (*models.Category, error) {
	var category models.Category
	result := s.DB.Preload("Parent").Preload("Children").First(&category, "id = ?", categoryID)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, errors.New("category not found")
		}
		return nil, fmt.Errorf("failed to fetch category: %w", result.Error)
	}

	return &category, nil
}

// FindCategoryBySlug retrieves a category by slug
func (s *CategoryService) FindCategoryBySlug(slug string) (*models.Category, error) {
	var category models.Category
	result := s.DB.Preload("Parent").Preload("Children").First(&category, "slug = ?", slug)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, errors.New("category not found")
		}
		return nil, fmt.Errorf("failed to fetch category: %w", result.Error)
	}

	return &category, nil
}

// FindCategories retrieves all categories with optional filters
func (s *CategoryService) FindCategories(parentID *string, level *int) ([]models.Category, error) {
	query := s.DB.Model(&models.Category{})

	// Filter by parent
	if parentID != nil {
		if *parentID == "null" || *parentID == "" {
			// Root categories (no parent)
			query = query.Where("parent_id IS NULL")
		} else {
			query = query.Where("parent_id = ?", *parentID)
		}
	}

	// Filter by level
	if level != nil {
		query = query.Where("level = ?", *level)
	}

	var categories []models.Category
	result := query.Order("name ASC").Find(&categories)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to fetch categories: %w", result.Error)
	}

	return categories, nil
}

// FindRootCategories retrieves all root categories (no parent)
func (s *CategoryService) FindRootCategories() ([]models.Category, error) {
	var categories []models.Category
	result := s.DB.Where("parent_id IS NULL").Preload("Children").Order("name ASC").Find(&categories)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to fetch root categories: %w", result.Error)
	}

	return categories, nil
}

// FindCategoryTree builds a hierarchical tree of categories
func (s *CategoryService) FindCategoryTree() ([]models.Category, error) {
	var categories []models.Category
	// Get all categories ordered by level and name
	result := s.DB.Order("level ASC, name ASC").Find(&categories)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to fetch categories: %w", result.Error)
	}

	// Build tree structure
	categoryMap := make(map[uuid.UUID]*models.Category)
	var roots []models.Category

	// First pass: create map
	for i := range categories {
		categoryMap[categories[i].ID] = &categories[i]
		categories[i].Children = []models.Category{}
	}

	// Second pass: build tree
	for i := range categories {
		if categories[i].ParentID == nil {
			roots = append(roots, categories[i])
		} else {
			parent := categoryMap[*categories[i].ParentID]
			if parent != nil {
				parent.Children = append(parent.Children, categories[i])
			}
		}
	}

	return roots, nil
}

// DeleteCategory deletes a category by ID
func (s *CategoryService) DeleteCategory(categoryID string, deleteChildren bool) error {
	var category models.Category
	result := s.DB.Preload("Children").First(&category, "id = ?", categoryID)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return errors.New("category not found")
		}
		return fmt.Errorf("failed to fetch category: %w", result.Error)
	}

	// Check if category has children
	if len(category.Children) > 0 {
		if !deleteChildren {
			return errors.New("category has children, cannot delete (use deleteChildren=true to force)")
		}
		// Recursively delete children
		for _, child := range category.Children {
			if err := s.DeleteCategory(child.ID.String(), true); err != nil {
				return err
			}
		}
	}

	result = s.DB.Delete(&category)
	if result.Error != nil {
		return fmt.Errorf("failed to delete category: %w", result.Error)
	}

	return nil
}

// GetCategoryPostCount returns the number of posts in a category
func (s *CategoryService) GetCategoryPostCount(categoryID string) (int64, error) {
	var count int64
	result := s.DB.Table("post_categories").Where("category_id = ?", categoryID).Count(&count)
	if result.Error != nil {
		return 0, fmt.Errorf("failed to count posts: %w", result.Error)
	}
	return count, nil
}

// SearchCategories searches categories by name or description
func (s *CategoryService) SearchCategories(query string) ([]models.Category, error) {
	searchPattern := "%" + query + "%"

	var categories []models.Category
	result := s.DB.Preload("Parent").
		Where("name ILIKE ? OR description ILIKE ?", searchPattern, searchPattern).
		Order("name ASC").
		Find(&categories)

	if result.Error != nil {
		return nil, fmt.Errorf("failed to search categories: %w", result.Error)
	}

	return categories, nil
}
