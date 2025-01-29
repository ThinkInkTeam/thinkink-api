package main

import (
	"thinkink-core-backend/database"
	"thinkink-core-backend/handlers"
	"thinkink-core-backend/middleware"

	"github.com/gin-gonic/gin"
)

func main() {
	database.ConnectDB()

	router := gin.Default()

	router.POST("/register", handlers.Register)
	router.POST("/login", handlers.Login)

	authRoutes := router.Group("/users")
	authRoutes.Use(middleware.JWTAuth())
	{
		authRoutes.PUT("/profile", handlers.UpdateUser)
		authRoutes.POST("/logout", handlers.Logout)
	}

	router.Run(":8080")
}
