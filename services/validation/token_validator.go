package validation

import (
	"fmt"
	"strings"

	"github.com/ThinkInkTeam/thinkink-core-backend/database"
	"github.com/ThinkInkTeam/thinkink-core-backend/models"
	"github.com/ThinkInkTeam/thinkink-core-backend/utils"
	"github.com/golang-jwt/jwt/v5"
)

// TokenValidator handles JWT token validation and user subscription checks
type TokenValidator struct{}

// NewTokenValidator creates a new TokenValidator instance
func NewTokenValidator() *TokenValidator {
	return &TokenValidator{}
}

// ValidateToken validates a JWT token and checks if the user has an active subscription
func (tv *TokenValidator) ValidateToken(tokenString string) bool {
	// Validate token format
	if tokenString == "" {
		return false
	}

	// Remove "Bearer " prefix if present
	tokenString = strings.TrimPrefix(tokenString, "Bearer ")

	// Check if token is blacklisted
	isBlacklisted, err := models.IsTokenBlacklisted(database.DB, tokenString)
	if err != nil || isBlacklisted {
		return false
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

	if err != nil || !token.Valid {
		return false
	}

	// Extract claims
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return false
	}

	// Extract user ID from claims
	userIDFloat, ok := claims["userID"]
	if !ok {
		return false
	}

	userID := uint(userIDFloat.(float64))

	// Find user and check subscription
	user, err := models.FindUserByID(database.DB, userID)
	if err != nil {
		return false
	}

	// Check if user has active subscription
	return user.IsSubscribed()
}
