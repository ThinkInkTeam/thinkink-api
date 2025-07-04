package middleware

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/ThinkInkTeam/thinkink-core-backend/database"
	"github.com/ThinkInkTeam/thinkink-core-backend/models"
	"github.com/ThinkInkTeam/thinkink-core-backend/utils"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// AuthMiddleware validates JWT tokens and sets the user ID in the context
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header is required"})
			c.Abort()
			return
		}

		// Check if the format is correct
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization format must be Bearer {token}"})
			c.Abort()
			return
		}

	tokenString := parts[1]

	// Check if token is blacklisted
	isBlacklisted, err := models.IsTokenBlacklisted(database.DB, tokenString)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Authentication error"})
		c.Abort()
		return
	}
	if isBlacklisted {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Token has been revoked"})
		c.Abort()
		return
	}

	// Get JWT secret from environment variable or use a default for development
	jwtSecret := utils.GetEnvWithDefault("JWT_SECRET", "your_jwt_secret")

		// Parse and validate token
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			// Validate signing method
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return []byte(jwtSecret), nil
		})

		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
			c.Abort()
			return
		}

		// Validate token is valid
		if !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			c.Abort()
			return
		}

		// Extract claims
		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Failed to extract token claims"})
			c.Abort()
			return
		}

		// Extract user ID from claims
		userID, ok := claims["userID"]
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token claims"})
			c.Abort()
			return
		}

		// Set user ID in context for later use in handlers
		c.Set("userID", uint(userID.(float64)))
		c.Next()
	}
}