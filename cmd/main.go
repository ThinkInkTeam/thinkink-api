package main

import (
	"github.com/ThinkInkTeam/thinkink-core-backend/handlers"

	"github.com/gin-gonic/gin"
)

func main() {
	//database.ConnectDB()

	router := gin.Default()

	router.POST("/register", handlers.Register)
	router.POST("/login", handlers.Login)

	authRoutes := router.Group("/users")
	//authRoutes.Use(middleware.JWTAuth())
	{
		authRoutes.PUT("/profile", handlers.UpdateUser)
		authRoutes.POST("/logout", handlers.Logout)
		authRoutes.POST("/upload", handlers.UploadSignalFile)
	}

	router.Run(":8080")
}
