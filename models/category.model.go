package models

import (
	"time"

	"github.com/google/uuid"
)

// Category represents a post category with hierarchical support
type Category struct {
	ID          uuid.UUID `gorm:"type:uuid;default:uuid_generate_v4();primary_key" json:"id"`
	Name        string    `gorm:"type:varchar(255);not null" json:"name"`
	Slug        string    `gorm:"type:varchar(255);uniqueIndex;not null" json:"slug"`
	Description string    `gorm:"type:text" json:"description,omitempty"`
	Image       string    `gorm:"type:varchar(500)" json:"image,omitempty"`

	// Hierarchical support
	ParentID *uuid.UUID `gorm:"type:uuid;index" json:"parent_id,omitempty"`
	Parent   *Category  `gorm:"foreignKey:ParentID;references:ID" json:"parent,omitempty"`
	Children []Category `gorm:"foreignKey:ParentID;references:ID" json:"children,omitempty"`
	Level    int        `gorm:"default:0" json:"level"`
	Path     string     `gorm:"type:varchar(500)" json:"path,omitempty"`

	// Timestamps
	CreatedAt time.Time `gorm:"not null" json:"created_at"`
	UpdatedAt time.Time `gorm:"not null" json:"updated_at"`
}

// TableName specifies the table name for the Category model
func (Category) TableName() string {
	return "categories"
}

// CreateCategoryRequest represents the request payload for creating a category
type CreateCategoryRequest struct {
	Name        string     `json:"name" binding:"required"`
	Slug        string     `json:"slug" binding:"required"`
	Description string     `json:"description"`
	Image       string     `json:"image" binding:"omitempty,url"`
	ParentID    *uuid.UUID `json:"parent_id,omitempty"`
}

// UpdateCategoryRequest represents the request payload for updating a category
type UpdateCategoryRequest struct {
	Name        *string    `json:"name,omitempty"`
	Slug        *string    `json:"slug,omitempty"`
	Description *string    `json:"description,omitempty"`
	Image       *string    `json:"image,omitempty"`
	ParentID    *uuid.UUID `json:"parent_id,omitempty"`
}
