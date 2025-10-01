package main

import (
	"log"
	"net/http"

	"app/controllers"
	"app/initializers"
	"app/routes"
	"app/services"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

var (
	server              *gin.Engine
	AuthController      *controllers.AuthController
	AuthRouteController routes.AuthRouteController

	UserController      *controllers.UserController
	UserRouteController routes.UserRouteController

	PostController      *controllers.PostController
	PostRouteController routes.PostRouteController

	CategoryController      *controllers.CategoryController
	CategoryRouteController routes.CategoryRouteController
)

func init() {
	config, err := initializers.LoadConfig(".")
	if err != nil {
		log.Fatal("🚀 Could not load environment variables", err)
	}

	initializers.ConnectDB(&config)

	// Initialize Services
	authService := services.NewAuthService(initializers.DB)
	userService := services.NewUserService(initializers.DB)
	postService := services.NewPostService(initializers.DB)
	categoryService := services.NewCategoryService(initializers.DB)

	// Initialize Controllers with Services
	AuthController = controllers.NewAuthController(authService)
	AuthRouteController = routes.NewAuthRouteController(AuthController)

	UserController = controllers.NewUserController(userService)
	UserRouteController = routes.NewRouteUserController(UserController)

	PostController = controllers.NewPostController(postService)
	PostRouteController = routes.NewRoutePostController(PostController)

	CategoryController = controllers.NewCategoryController(categoryService)
	CategoryRouteController = routes.NewCategoryRouteController(CategoryController)

	server = gin.Default()
}

func main() {
	config, err := initializers.LoadConfig(".")
	if err != nil {
		log.Fatal("🚀 Could not load environment variables", err)
	}

	corsConfig := cors.DefaultConfig()
	corsConfig.AllowOrigins = []string{"http://localhost:8000", "http://localhost:3000", "http://localhost:5173", config.ClientOrigin}
	corsConfig.AllowCredentials = true
	corsConfig.AllowMethods = []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"}
	corsConfig.AllowHeaders = []string{"Origin", "Content-Type", "Accept", "Authorization", "X-Requested-With"}
	corsConfig.ExposeHeaders = []string{"Content-Length"}

	server.Use(cors.New(corsConfig))

	router := server.Group("/api")
	router.GET("/healthchecker", func(ctx *gin.Context) {
		message := "Welcome to Golang with Gorm and Postgres"
		ctx.JSON(http.StatusOK, gin.H{"status": "success", "message": message})
	})

	AuthRouteController.AuthRoute(router)
	UserRouteController.UserRoute(router)
	PostRouteController.PostRoute(router)
	CategoryRouteController.CategoryRoute(router)
	log.Fatal(server.Run(":" + config.ServerPort))
}
