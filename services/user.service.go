package services

import (
	"app/models"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"
)

type UserService struct {
	DB *gorm.DB
}

func NewUserService(db *gorm.DB) *UserService {
	return &UserService{DB: db}
}

// GetUserResponse converts a User model to UserResponse
func (s *UserService) GetUserResponse(user *models.User) *models.UserResponse {
	return &models.UserResponse{
		ID:        user.ID,
		Name:      user.Name,
		Email:     user.Email,
		Photo:     user.Photo,
		Role:      user.Role,
		Provider:  user.Provider,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}
}

// GetUsers retrieves all users (admin only)
func (s *UserService) GetUsers() ([]models.UserResponse, error) {
	var users []models.User
	result := s.DB.Find(&users)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to fetch users: %w", result.Error)
	}

	var userResponses []models.UserResponse
	for _, user := range users {
		userResponses = append(userResponses, *s.GetUserResponse(&user))
	}

	return userResponses, nil
}

// UpdateUser updates a user by ID (admin only)
func (s *UserService) UpdateUser(userID string, input *models.UpdateUserInput) (*models.UserResponse, error) {
	var user models.User
	result := s.DB.First(&user, "id = ?", userID)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, errors.New("user not found")
		}
		return nil, fmt.Errorf("failed to fetch user: %w", result.Error)
	}

	// Update fields if provided
	if input.Name != "" {
		user.Name = input.Name
	}
	if input.Email != "" {
		user.Email = input.Email
	}
	if input.Role != "" {
		user.Role = input.Role
	}
	if input.Photo != "" {
		user.Photo = input.Photo
	}

	user.UpdatedAt = time.Now()

	result = s.DB.Save(&user)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to update user: %w", result.Error)
	}

	return s.GetUserResponse(&user), nil
}

// DeleteUser deletes a user by ID (admin only)
func (s *UserService) DeleteUser(userID string) error {
	var user models.User
	result := s.DB.First(&user, "id = ?", userID)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return errors.New("user not found")
		}
		return fmt.Errorf("failed to fetch user: %w", result.Error)
	}

	result = s.DB.Delete(&user)
	if result.Error != nil {
		return fmt.Errorf("failed to delete user: %w", result.Error)
	}

	return nil
}
