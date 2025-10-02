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

type PostService struct {
	DB *gorm.DB
}

func NewPostService(db *gorm.DB) *PostService {
	return &PostService{DB: db}
}

// CreatePost creates a new post with all fields
func (s *PostService) CreatePost(payload *models.CreatePostRequest, userID uuid.UUID) (*models.Post, error) {
	now := time.Now()

	// Parse category IDs
	var categories []models.Category
	if len(payload.CategoryIDs) > 0 {
		for _, catID := range payload.CategoryIDs {
			categoryUUID, err := uuid.Parse(catID)
			if err != nil {
				continue
			}
			var category models.Category
			if err := s.DB.First(&category, "id = ?", categoryUUID).Error; err == nil {
				categories = append(categories, category)
			}
		}
	}

	newPost := models.Post{
		Title:              payload.Title,
		Slug:               payload.Slug,
		Content:            payload.Content,
		Thumbnail:          payload.Thumbnail,
		Image:              payload.Image,
		ShortDescription:   payload.ShortDescription,
		MetaDescription:    payload.MetaDescription,
		Keywords:           payload.Keywords,
		Tags:               payload.Tags,
		Categories:         categories,
		AuthorID:           userID,
		ReadingTimeMinutes: payload.ReadingTimeMinutes,
		Published:          payload.Published,
		CreatedAt:          now,
		UpdatedAt:          now,
	}

	// Set published date if published
	if payload.Published {
		newPost.PublishedAt = &now
	}

	result := s.DB.Create(&newPost)
	if result.Error != nil {
		if strings.Contains(result.Error.Error(), "duplicate key") {
			return nil, errors.New("post with that title or slug already exists")
		}
		return nil, fmt.Errorf("failed to create post: %w", result.Error)
	}

	// Reload with associations
	s.DB.Preload("Author").Preload("Categories").First(&newPost, "id = ?", newPost.ID)

	return &newPost, nil
}

// UpdatePost updates an existing post
func (s *PostService) UpdatePost(postID string, payload *models.UpdatePostRequest, userID uuid.UUID) (*models.Post, error) {
	var existingPost models.Post
	result := s.DB.Preload("Categories").First(&existingPost, "id = ?", postID)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, errors.New("post not found")
		}
		return nil, fmt.Errorf("failed to fetch post: %w", result.Error)
	}

	now := time.Now()

	// Update fields if provided
	if payload.Title != nil {
		existingPost.Title = *payload.Title
	}
	if payload.Slug != nil {
		existingPost.Slug = *payload.Slug
	}
	if payload.Content != nil {
		existingPost.Content = *payload.Content
	}
	if payload.Thumbnail != nil {
		existingPost.Thumbnail = *payload.Thumbnail
	}
	if payload.Image != nil {
		existingPost.Image = *payload.Image
	}
	if payload.ShortDescription != nil {
		existingPost.ShortDescription = *payload.ShortDescription
	}
	if payload.MetaDescription != nil {
		existingPost.MetaDescription = *payload.MetaDescription
	}
	if payload.Keywords != nil {
		existingPost.Keywords = *payload.Keywords
	}
	if payload.Tags != nil {
		existingPost.Tags = *payload.Tags
	}
	if payload.ReadingTimeMinutes != nil {
		existingPost.ReadingTimeMinutes = *payload.ReadingTimeMinutes
	}
	if payload.Views != nil {
		existingPost.Views = *payload.Views
	}
	if payload.Likes != nil {
		existingPost.Likes = *payload.Likes
	}
	if payload.Comments != nil {
		existingPost.Comments = *payload.Comments
	}

	// Handle publishing status
	if payload.Published != nil {
		wasPublished := existingPost.Published
		existingPost.Published = *payload.Published

		// Set published date when first published
		if !wasPublished && existingPost.Published && existingPost.PublishedAt == nil {
			existingPost.PublishedAt = &now
		}
	}

	// Update categories if provided
	if payload.CategoryIDs != nil {
		var categories []models.Category
		for _, catID := range *payload.CategoryIDs {
			categoryUUID, err := uuid.Parse(catID)
			if err != nil {
				continue
			}
			var category models.Category
			if err := s.DB.First(&category, "id = ?", categoryUUID).Error; err == nil {
				categories = append(categories, category)
			}
		}

		// Replace categories
		if err := s.DB.Model(&existingPost).Association("Categories").Replace(categories); err != nil {
			return nil, fmt.Errorf("failed to update categories: %w", err)
		}
	}

	existingPost.UpdatedAt = now

	result = s.DB.Save(&existingPost)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to update post: %w", result.Error)
	}

	// Reload with associations
	s.DB.Preload("Author").Preload("Categories").First(&existingPost, "id = ?", existingPost.ID)

	return &existingPost, nil
}

// FindPostByID retrieves a post by ID with all associations
func (s *PostService) FindPostByID(postID string) (*models.Post, error) {
	var post models.Post
	result := s.DB.Preload("Author").Preload("Categories").First(&post, "id = ?", postID)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, errors.New("post not found")
		}
		return nil, fmt.Errorf("failed to fetch post: %w", result.Error)
	}

	return &post, nil
}

// FindPostBySlug retrieves a post by slug with all associations
func (s *PostService) FindPostBySlug(slug string) (*models.Post, error) {
	var post models.Post
	result := s.DB.Preload("Author").Preload("Categories").First(&post, "slug = ?", slug)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, errors.New("post not found")
		}
		return nil, fmt.Errorf("failed to fetch post: %w", result.Error)
	}

	return &post, nil
}

// FindPosts retrieves a paginated list of posts with filters
func (s *PostService) FindPosts(page, limit int, published *bool, categoryID *string) ([]models.Post, int64, error) {
	offset := (page - 1) * limit

	query := s.DB.Model(&models.Post{}).Preload("Author").Preload("Categories")

	// Filter by published status if provided
	if published != nil {
		query = query.Where("published = ?", *published)
	}

	// Filter by category if provided
	if categoryID != nil && *categoryID != "" {
		query = query.Joins("JOIN post_categories ON posts.id = post_categories.post_id").
			Where("post_categories.category_id = ?", *categoryID)
	}

	// Get total count
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count posts: %w", err)
	}

	// Get posts
	var posts []models.Post
	result := query.Order("created_at DESC").Limit(limit).Offset(offset).Find(&posts)
	if result.Error != nil {
		return nil, 0, fmt.Errorf("failed to fetch posts: %w", result.Error)
	}

	return posts, total, nil
}

// DeletePost deletes a post by ID
func (s *PostService) DeletePost(postID string) error {
	result := s.DB.Delete(&models.Post{}, "id = ?", postID)
	if result.Error != nil {
		return fmt.Errorf("failed to delete post: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return errors.New("post not found")
	}

	return nil
}

// IncrementViews increments the view count for a post
func (s *PostService) IncrementViews(postID string) error {
	result := s.DB.Model(&models.Post{}).Where("id = ?", postID).UpdateColumn("views", gorm.Expr("views + ?", 1))
	if result.Error != nil {
		return fmt.Errorf("failed to increment views: %w", result.Error)
	}
	return nil
}

// IncrementLikes increments the like count for a post
func (s *PostService) IncrementLikes(postID string) error {
	result := s.DB.Model(&models.Post{}).Where("id = ?", postID).UpdateColumn("likes", gorm.Expr("likes + ?", 1))
	if result.Error != nil {
		return fmt.Errorf("failed to increment likes: %w", result.Error)
	}
	return nil
}

// SearchPosts searches posts by title, content, or tags
func (s *PostService) SearchPosts(query string, page, limit int) ([]models.Post, int64, error) {
	offset := (page - 1) * limit
	searchPattern := "%" + query + "%"

	dbQuery := s.DB.Model(&models.Post{}).
		Preload("Author").
		Preload("Categories").
		Where("published = ?", true).
		Where("title ILIKE ? OR content ILIKE ? OR short_description ILIKE ?",
			searchPattern, searchPattern, searchPattern)

	// Get total count
	var total int64
	if err := dbQuery.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count search results: %w", err)
	}

	// Get posts
	var posts []models.Post
	result := dbQuery.Order("created_at DESC").Limit(limit).Offset(offset).Find(&posts)
	if result.Error != nil {
		return nil, 0, fmt.Errorf("failed to search posts: %w", result.Error)
	}

	return posts, total, nil
}
