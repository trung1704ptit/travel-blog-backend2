package controllers

import (
	"net/http"

	"app/models"
	"app/services"

	"github.com/gin-gonic/gin"
)

type UserController struct {
	userService *services.UserService
}

func NewUserController(userService *services.UserService) *UserController {
	return &UserController{userService: userService}
}

// GetMe returns the current authenticated user's information
func (uc *UserController) GetMe(ctx *gin.Context) {
	currentUser := ctx.MustGet("currentUser").(models.User)

	userResponse := uc.userService.GetUserResponse(&currentUser)

	ctx.JSON(http.StatusOK, gin.H{"status": "success", "data": gin.H{"user": userResponse}})
}

// GetUsers returns all users (admin only)
func (uc *UserController) GetUsers(ctx *gin.Context) {
	users, err := uc.userService.GetUsers()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"status": "success", "data": gin.H{"users": users}})
}

// UpdateUser updates a user by ID (admin only)
func (uc *UserController) UpdateUser(ctx *gin.Context) {
	userID := ctx.Param("id")

	var input models.UpdateUserInput
	if err := ctx.ShouldBindJSON(&input); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"status": "fail", "message": err.Error()})
		return
	}

	user, err := uc.userService.UpdateUser(userID, &input)
	if err != nil {
		if err.Error() == "user not found" {
			ctx.JSON(http.StatusNotFound, gin.H{"status": "fail", "message": err.Error()})
		} else {
			ctx.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": err.Error()})
		}
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"status": "success", "data": gin.H{"user": user}})
}

// DeleteUser deletes a user by ID (admin only)
func (uc *UserController) DeleteUser(ctx *gin.Context) {
	userID := ctx.Param("id")

	err := uc.userService.DeleteUser(userID)
	if err != nil {
		if err.Error() == "user not found" {
			ctx.JSON(http.StatusNotFound, gin.H{"status": "fail", "message": err.Error()})
		} else {
			ctx.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": err.Error()})
		}
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"status": "success", "message": "User deleted successfully"})
}
