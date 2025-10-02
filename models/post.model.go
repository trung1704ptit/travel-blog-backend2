package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
)

// JSONStringSlice is a custom type for storing string slices as JSON in the database
type JSONStringSlice []string

// Value implements the driver.Valuer interface
func (j JSONStringSlice) Value() (driver.Value, error) {
	if j == nil {
		return nil, nil
	}
	return json.Marshal(j)
}

// Scan implements the sql.Scanner interface
func (j *JSONStringSlice) Scan(value interface{}) error {
	if value == nil {
		*j = nil
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("failed to unmarshal JSONStringSlice value")
	}
	return json.Unmarshal(bytes, j)
}

// Post represents an article/blog post in the system
type Post struct {
	ID               uuid.UUID       `gorm:"type:uuid;default:uuid_generate_v4();primary_key" json:"id"`
	Title            string          `gorm:"type:varchar(255);uniqueIndex;not null" json:"title"`
	Slug             string          `gorm:"type:varchar(255);uniqueIndex;not null" json:"slug"`
	Content          string          `gorm:"type:text;not null" json:"content"`
	Thumbnail        string          `gorm:"type:varchar(500)" json:"thumbnail,omitempty"`
	Image            string          `gorm:"type:varchar(500)" json:"image,omitempty"`
	ShortDescription string          `gorm:"type:text" json:"short_description,omitempty"`
	MetaDescription  string          `gorm:"type:text" json:"meta_description,omitempty"`
	Keywords         JSONStringSlice `gorm:"type:jsonb" json:"keywords,omitempty"`
	Tags             JSONStringSlice `gorm:"type:jsonb" json:"tags,omitempty"`

	// Relationships
	Categories []Category `gorm:"many2many:post_categories;" json:"categories,omitempty"`
	AuthorID   uuid.UUID  `gorm:"type:uuid;not null;index" json:"author_id"`
	Author     User       `gorm:"foreignKey:AuthorID;references:ID" json:"author,omitempty"`

	// Metrics
	ReadingTimeMinutes int `gorm:"default:3" json:"reading_time_minutes"`
	Views              int `gorm:"default:23" json:"views"`
	Likes              int `gorm:"default:99" json:"likes"`
	Comments           int `gorm:"default:0" json:"comments"`

	// Publishing
	Published   bool       `gorm:"default:false;index" json:"published"`
	PublishedAt *time.Time `json:"published_at,omitempty"`

	// Timestamps
	CreatedAt time.Time `gorm:"not null;index" json:"created_at"`
	UpdatedAt time.Time `gorm:"not null" json:"updated_at"`
}

// TableName specifies the table name for the Post model
func (Post) TableName() string {
	return "posts"
}

// CreatePostRequest represents the request payload for creating a new post
type CreatePostRequest struct {
	Title              string   `json:"title" binding:"required"`
	Slug               string   `json:"slug" binding:"required"`
	Content            string   `json:"content" binding:"required"`
	Thumbnail          string   `json:"thumbnail" binding:"omitempty,url"`
	Image              string   `json:"image" binding:"omitempty,url"`
	ShortDescription   string   `json:"short_description"`
	MetaDescription    string   `json:"meta_description"`
	Keywords           []string `json:"keywords"`
	Tags               []string `json:"tags"`
	CategoryIDs        []string `json:"category_ids"`
	ReadingTimeMinutes int      `json:"reading_time_minutes"`
	Published          bool     `json:"published"`
}

// UpdatePostRequest represents the request payload for updating a post
type UpdatePostRequest struct {
	Title              *string   `json:"title,omitempty"`
	Slug               *string   `json:"slug,omitempty"`
	Content            *string   `json:"content,omitempty"`
	Thumbnail          *string   `json:"thumbnail,omitempty"`
	Image              *string   `json:"image,omitempty"`
	ShortDescription   *string   `json:"short_description,omitempty"`
	MetaDescription    *string   `json:"meta_description,omitempty"`
	Keywords           *[]string `json:"keywords,omitempty"`
	Tags               *[]string `json:"tags,omitempty"`
	CategoryIDs        *[]string `json:"category_ids,omitempty"`
	Views              *int      `json:"views,omitempty"`
	Likes              *int      `json:"likes,omitempty"`
	Comments           *int      `json:"comments,omitempty"`
	ReadingTimeMinutes *int      `json:"reading_time_minutes,omitempty"`
	Published          *bool     `json:"published,omitempty"`
}

// PostResponse represents the response payload for a post
type PostResponse struct {
	ID                 uuid.UUID    `json:"id"`
	Title              string       `json:"title"`
	Slug               string       `json:"slug"`
	Content            string       `json:"content"`
	Thumbnail          string       `json:"thumbnail,omitempty"`
	Image              string       `json:"image,omitempty"`
	ShortDescription   string       `json:"short_description,omitempty"`
	MetaDescription    string       `json:"meta_description,omitempty"`
	Keywords           []string     `json:"keywords,omitempty"`
	Tags               []string     `json:"tags,omitempty"`
	Categories         []Category   `json:"categories,omitempty"`
	Author             UserResponse `json:"author"`
	ReadingTimeMinutes int          `json:"reading_time_minutes"`
	Views              int          `json:"views"`
	Likes              int          `json:"likes"`
	Comments           int          `json:"comments"`
	Published          bool         `json:"published"`
	PublishedAt        *time.Time   `json:"published_at,omitempty"`
	CreatedAt          time.Time    `json:"created_at"`
	UpdatedAt          time.Time    `json:"updated_at"`
}

// Deprecated: Use UpdatePostRequest instead
type UpdatePost struct {
	Title     string    `json:"title,omitempty"`
	Content   string    `json:"content,omitempty"`
	Image     string    `json:"image,omitempty"`
	User      string    `json:"user,omitempty"`
	CreateAt  time.Time `json:"created_at,omitempty"`
	UpdatedAt time.Time `json:"updated_at,omitempty"`
}
