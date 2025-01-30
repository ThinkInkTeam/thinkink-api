package middleware

import (
	"github.com/ThinkInkTeam/thinkink-core-backend/database"
	"github.com/ThinkInkTeam/thinkink-core-backend/models"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

func JWTAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Authorization header missing"})
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")

		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			return []byte("your_jwt_secret"), nil
		})

		if err != nil || !token.Valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			return
		}

		var blacklistedToken models.BlacklistedToken
		if err := database.DB.Where("token = ?", tokenString).First(&blacklistedToken).Error; err == nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Token revoked. Please login again"})
			return
		}

		claims := token.Claims.(jwt.MapClaims)
		userID := uint(claims["userID"].(float64))
		c.Set("userID", userID)

		c.Next()
	}
}
