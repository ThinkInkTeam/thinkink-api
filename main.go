package main

import (
	"github.com/ThinkInkTeam/thinkink-core-backend/handlers"

	_ "github.com/ThinkInkTeam/thinkink-core-backend/docs" // Adjust package path
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func main() {
	//database.ConnectDB()

	router := gin.Default()

	router.POST("/register", handlers.Register)
	router.POST("/login", handlers.Login)

	authRoutes := router.Group("/users")
	router.GET("/swagger/*any",ginSwagger.WrapHandler(swaggerFiles.Handler))
	//authRoutes.Use(middleware.JWTAuth())
	{
		authRoutes.PUT("/update", handlers.UpdateUser)
		authRoutes.POST("/logout", handlers.Logout)
		authRoutes.POST("/upload", handlers.UploadSignalFile)
		authRoutes.POST("/profile", handlers.GetUser)
	}

	router.Run(":8080")
}
