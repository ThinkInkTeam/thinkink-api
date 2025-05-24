package handlers

import (
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/ThinkInkTeam/thinkink-core-backend/database"
	"github.com/ThinkInkTeam/thinkink-core-backend/models"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// SignUpRequest represents the request for user registration
type SignUpRequest struct {
	Name         string                 `json:"name" binding:"required" example:"John Doe"`
	Email        string                 `json:"email" binding:"required,email" example:"john@example.com"`
	Password     string                 `json:"password" binding:"required,min=8" example:"password123"`
	DateOfBirth  time.Time              `json:"date_of_birth" binding:"required" example:"1990-01-01T00:00:00Z"`
	Mobile       string                 `json:"mobile" example:"5551234567"`
	CountryCode  string                 `json:"country_code" example:"+1"`
	Address      string                 `json:"address" example:"123 Main St"`
	City         string                 `json:"city" example:"New York"`
	Country      string                 `json:"country" example:"US"`
	PostalCode   string                 `json:"postal_code" example:"10001"`
	PaymentInfo  map[string]interface{} `json:"payment_info" swaggertype:"object,string" example:"{\"card_type\":\"visa\"}"`
}

// SignInRequest represents the request for user authentication
type SignInRequest struct {
	Email    string `json:"email" binding:"required,email" example:"john@example.com"`
	Password string `json:"password" binding:"required" example:"password123"`
}

// AuthResponse represents the response for authentication endpoints
type AuthResponse struct {
	Message string      `json:"message" example:"Login successful"`
	User    UserInfo    `json:"user"`
	Token   string      `json:"token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
}

// UserInfo represents basic user information
type UserInfo struct {
	ID    uint   `json:"id" example:"1"`
	Name  string `json:"name" example:"John Doe"`
	Email string `json:"email" example:"john@example.com"`
}

// TokenResponse represents a response containing just a token
type TokenResponse struct {
	Message string `json:"message" example:"Token refreshed successfully"`
	Token   string `json:"token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
}

// MessageResponse represents a simple message response
type MessageResponse struct {
	Message string `json:"message" example:"Operation completed successfully"`
}

// ForgotPasswordRequest represents the request for password reset initiation
type ForgotPasswordRequest struct {
	Email string `json:"email" binding:"required,email" example:"user@example.com"`
}

// ForgotPasswordResponse represents the response for password reset initiation
type ForgotPasswordResponse struct {
	Message string `json:"message" example:"Password reset instructions sent to your email"`
}

// ResetPasswordRequest represents the request for password reset completion
type ResetPasswordRequest struct {
	Token    string `json:"token" binding:"required" example:"reset-token"`
	Password string `json:"password" binding:"required,min=8" example:"new-password"`
}

// SignUp handles user registration
// @Summary Register a new user
// @Description Register a new user with the provided information
// @Tags auth
// @Accept json
// @Produce json
// @Param user body SignUpRequest true "User Registration Information"
// @Success 201 {object} AuthResponse "User created successfully with token"
// @Failure 400 {object} ErrorResponse "Bad Request - Invalid input"
// @Failure 500 {object} ErrorResponse "Internal Server Error"
// @Router /signup [post]
func SignUp(c *gin.Context) {
	var req SignUpRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	user, err := models.CreateUser(
		database.DB,
		req.Name,
		req.Email,
		req.Password,
		req.DateOfBirth,
		req.Mobile,
		req.CountryCode,
		req.Address,
		req.City,
		req.Country,
		req.PostalCode,
		req.PaymentInfo,
	)

	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	token, err := user.GenerateJWT()
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to generate token"})
		return
	}

	c.JSON(http.StatusCreated, AuthResponse{
		Message: "User registered successfully",
		User: UserInfo{
			ID:    user.ID,
			Name:  user.Name,
			Email: user.Email,
		},
		Token: token,
	})
}

// SignIn handles user authentication
// @Summary Authenticate a user
// @Description Authenticate a user with email and password
// @Tags auth
// @Accept json
// @Produce json
// @Param credentials body SignInRequest true "User Credentials"
// @Success 200 {object} AuthResponse "User authenticated successfully with token"
// @Failure 400 {object} ErrorResponse "Bad Request - Invalid input"
// @Failure 401 {object} ErrorResponse "Unauthorized - Invalid credentials"
// @Failure 500 {object} ErrorResponse "Internal Server Error"
// @Router /signin [post]
func SignIn(c *gin.Context) {
	var req SignInRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	user, err := models.FindUserByEmail(database.DB, req.Email)
	if err != nil {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "Invalid email or password"})
		return
	}

	if err := user.ValidatePassword(req.Password); err != nil {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "Invalid email or password"})
		return
	}

	token, err := user.GenerateJWT()
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to generate token"})
		return
	}

	// Update last login time
	if err := user.UpdateLastLogin(database.DB); err != nil {
		// Non-critical error, just log it
		log.Printf("Failed to update last login time: %v", err)
	}

	c.JSON(http.StatusOK, AuthResponse{
		Message: "Login successful",
		User: UserInfo{
			ID:    user.ID,
			Name:  user.Name,
			Email: user.Email,
		},
		Token: token,
	})
}

// Logout logs out a user
// @Summary User logout
// @Description Logs out a user and invalidates the session token
// @Tags auth
// @Produce json
// @Security BearerAuth
// @Success 200 {object} MessageResponse "Logged out successfully"
// @Failure 401 {object} ErrorResponse "Unauthorized - Invalid or missing token"
// @Failure 500 {object} ErrorResponse "Internal Server Error"
// @Router /logout [post]
func Logout(c *gin.Context) {
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "Authorization header is required"})
		return
	}

	tokenString := strings.TrimPrefix(authHeader, "Bearer ")

	// Parse the token to get expiration time
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return []byte(os.Getenv("JWT_SECRET")), nil
	})

	if err != nil {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "Invalid token"})
		return
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok {
		var exp time.Time

		// Check if exp exists and convert to time.Time
		if expFloat, ok := claims["exp"].(float64); ok {
			exp = time.Unix(int64(expFloat), 0)
		} else {
			// If no expiration, set a default (1 day)
			exp = time.Now().Add(24 * time.Hour)
		}

		// Add the token to blacklist
		if err := models.AddToBlacklist(database.DB, tokenString, exp); err != nil {
			c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Logout failed"})
			return
		}

		c.JSON(http.StatusOK, MessageResponse{Message: "Logged out successfully"})
	} else {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "Invalid token claims"})
	}
}

// RefreshToken generates a new JWT token for the user
// @Summary Refresh authentication token
// @Description Generate a new JWT token using a valid existing token
// @Tags auth
// @Produce json
// @Security BearerAuth
// @Success 200 {object} TokenResponse "Token refreshed successfully"
// @Failure 401 {object} ErrorResponse "Unauthorized - Invalid or blacklisted token"
// @Failure 500 {object} ErrorResponse "Internal Server Error"
// @Router /refresh-token [post]
func RefreshToken(c *gin.Context) {
	// Get the user ID from the middleware context
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "Unauthorized"})
		return
	}

	// Get user from database
	user, err := models.FindUserByID(database.DB, userID.(uint))
	if err != nil {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "User not found"})
		return
	}

	// Generate a new token
	token, err := user.GenerateJWT()
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to generate token"})
		return
	}

	c.JSON(http.StatusOK, TokenResponse{
		Message: "Token refreshed successfully",
		Token:   token,
	})
}

// CheckAuth validates if a user's token is valid
// @Summary Validate authentication token
// @Description Check if the current token is valid and not blacklisted
// @Tags auth
// @Produce json
// @Security BearerAuth
// @Success 200 {object} AuthResponse "User authentication status"
// @Failure 401 {object} ErrorResponse "Unauthorized - Invalid or blacklisted token"
// @Router /check-auth [get]
func CheckAuth(c *gin.Context) {
	// Get the user ID from the middleware context
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "Unauthorized"})
		return
	}

	// Get user from database
	user, err := models.FindUserByID(database.DB, userID.(uint))
	if err != nil {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "User not found"})
		return
	}

	c.JSON(http.StatusOK, AuthResponse{
		Message: "User authentication status",
		User: UserInfo{
			ID:    user.ID,
			Name:  user.Name,
			Email: user.Email,
		},
	})
}

// ForgotPassword initiates the password reset process
// @Summary Request password reset
// @Description Send a password reset link to the user's email
// @Tags auth
// @Accept json
// @Produce json
// @Param request body ForgotPasswordRequest true "User email"
// @Success 200 {object} ForgotPasswordResponse "Password reset email sent"
// @Failure 400 {object} ErrorResponse "Bad Request - Invalid input"
// @Failure 404 {object} ErrorResponse "Not Found - User not found"
// @Failure 500 {object} ErrorResponse "Internal Server Error"
// @Router /forgot-password [post]
func ForgotPassword(c *gin.Context) {
	var req ForgotPasswordRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	user, err := models.FindUserByEmail(database.DB, req.Email)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	// Generate a reset token (in a real app, you'd send this via email)
	resetToken, err := user.GeneratePasswordResetToken(database.DB)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate reset token"})
		return
	}

	// Here you would normally send an email with the reset link
	// For this implementation, we'll just return the token in the response
	// In production, you should send an email and not expose the token in the response

	response := ForgotPasswordResponse{
		Message: "Password reset instructions sent to your email",
	}
	
	// In development mode, you might want to include the token for testing
	if os.Getenv("APP_ENV") != "production" {
		c.JSON(http.StatusOK, gin.H{
			"message": response.Message,
			"reset_token": resetToken, // Only included in non-production environments
		})
	} else {
		c.JSON(http.StatusOK, response)
	}
}

// ResetPassword completes the password reset process
// @Summary Reset user password
// @Description Reset the user's password using a valid reset token
// @Tags auth
// @Accept json
// @Produce json
// @Param request body ResetPasswordRequest true "Reset token and new password"
// @Success 200 {object} MessageResponse "Password reset successful"
// @Failure 400 {object} ErrorResponse "Bad Request - Invalid input"
// @Failure 401 {object} ErrorResponse "Unauthorized - Invalid or expired token"
// @Failure 500 {object} ErrorResponse "Internal Server Error"
// @Router /reset-password [post]
func ResetPassword(c *gin.Context) {
	var req ResetPasswordRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	// Verify the reset token and get the associated user
	user, err := models.VerifyPasswordResetToken(database.DB, req.Token)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired reset token"})
		return
	}

	// Update the user's password
	if err := user.UpdatePassword(database.DB, req.Password); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update password"})
		return
	}

	c.JSON(http.StatusOK, MessageResponse{Message: "Password reset successful"})
}
