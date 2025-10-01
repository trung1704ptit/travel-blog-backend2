package services

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"app/initializers"
	"app/models"
	"app/utils"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type AuthService struct {
	DB *gorm.DB
}

func NewAuthService(db *gorm.DB) *AuthService {
	return &AuthService{DB: db}
}

// SignUp creates a new user account
func (s *AuthService) SignUp(payload *models.SignUpInput) (*models.UserResponse, error) {
	if payload.Password != payload.PasswordConfirm {
		return nil, errors.New("passwords do not match")
	}

	hashedPassword, err := utils.HashPassword(payload.Password)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	now := time.Now()
	newUser := models.User{
		Name:      payload.Name,
		Email:     strings.ToLower(payload.Email),
		Password:  hashedPassword,
		Role:      "user",
		Verified:  true,
		Photo:     payload.Photo,
		Provider:  "local",
		CreatedAt: now,
		UpdatedAt: now,
	}

	result := s.DB.Create(&newUser)
	if result.Error != nil {
		if strings.Contains(result.Error.Error(), "duplicate key value violates unique") {
			return nil, errors.New("user with that email already exists")
		}
		return nil, fmt.Errorf("failed to create user: %w", result.Error)
	}

	userResponse := &models.UserResponse{
		ID:        newUser.ID,
		Name:      newUser.Name,
		Email:     newUser.Email,
		Photo:     newUser.Photo,
		Role:      newUser.Role,
		Provider:  newUser.Provider,
		CreatedAt: newUser.CreatedAt,
		UpdatedAt: newUser.UpdatedAt,
	}

	return userResponse, nil
}

// SignIn authenticates a user and returns tokens
func (s *AuthService) SignIn(payload *models.SignInInput) (*models.User, string, string, error) {
	var user models.User
	result := s.DB.First(&user, "email = ?", strings.ToLower(payload.Email))
	if result.Error != nil {
		return nil, "", "", errors.New("invalid email or password")
	}

	if err := utils.VerifyPassword(user.Password, payload.Password); err != nil {
		return nil, "", "", errors.New("invalid email or password")
	}

	config, err := initializers.LoadConfig(".")
	if err != nil {
		return nil, "", "", fmt.Errorf("failed to load config: %w", err)
	}

	// Generate Access Token
	accessToken, err := utils.CreateToken(config.AccessTokenExpiresIn, user.ID, config.AccessTokenPrivateKey)
	if err != nil {
		return nil, "", "", fmt.Errorf("failed to create access token: %w", err)
	}

	// Generate Refresh Token
	refreshToken, err := utils.CreateToken(config.RefreshTokenExpiresIn, user.ID, config.RefreshTokenPrivateKey)
	if err != nil {
		return nil, "", "", fmt.Errorf("failed to create refresh token: %w", err)
	}

	return &user, accessToken, refreshToken, nil
}

// RefreshAccessToken generates a new access token from a refresh token
func (s *AuthService) RefreshAccessToken(refreshToken string) (*models.User, string, error) {
	config, err := initializers.LoadConfig(".")
	if err != nil {
		return nil, "", fmt.Errorf("failed to load config: %w", err)
	}

	sub, err := utils.ValidateToken(refreshToken, config.RefreshTokenPublicKey)
	if err != nil {
		return nil, "", fmt.Errorf("invalid refresh token: %w", err)
	}

	var user models.User
	result := s.DB.First(&user, "id = ?", fmt.Sprint(sub))
	if result.Error != nil {
		return nil, "", errors.New("the user belonging to this token no longer exists")
	}

	accessToken, err := utils.CreateToken(config.AccessTokenExpiresIn, user.ID, config.AccessTokenPrivateKey)
	if err != nil {
		return nil, "", fmt.Errorf("failed to create access token: %w", err)
	}

	return &user, accessToken, nil
}

// GetUserByID retrieves a user by their ID
func (s *AuthService) GetUserByID(userID uuid.UUID) (*models.User, error) {
	var user models.User
	result := s.DB.First(&user, "id = ?", userID)
	if result.Error != nil {
		return nil, errors.New("user not found")
	}
	return &user, nil
}
