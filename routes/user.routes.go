package routes

import (
	"app/controllers"
	"app/middleware"

	"github.com/gin-gonic/gin"
)

type UserRouteController struct {
	userController *controllers.UserController
}

func NewRouteUserController(userController *controllers.UserController) UserRouteController {
	return UserRouteController{userController}
}

func (uc *UserRouteController) UserRoute(rg *gin.RouterGroup) {

	router := rg.Group("users")
	router.GET("/me", middleware.DeserializeUser(), uc.userController.GetMe)

	// Admin-only routes
	adminRouter := router.Group("")
	adminRouter.Use(middleware.DeserializeUser())
	{
		adminRouter.GET("", uc.userController.GetUsers)
		adminRouter.PUT("/:id", uc.userController.UpdateUser)
		adminRouter.DELETE("/:id", uc.userController.DeleteUser)
	}
}
