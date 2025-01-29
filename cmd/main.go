package main

import (
	"thinkink-core-backend/database"
	"thinkink-core-backend/handlers"

	"github.com/gin-gonic/gin"
)

func main() {
	database.ConnectDB()

	router := gin.Default()

	router.POST("/register", handlers.Register)
	router.POST("/login", handlers.Login)

	router.Run(":8080")
}
