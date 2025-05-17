package handlers

import (
	"net/http"
	"strconv"

	"github.com/ThinkInkTeam/thinkink-core-backend/database"
	"github.com/ThinkInkTeam/thinkink-core-backend/models"
	"github.com/gin-gonic/gin"
)

// UserResponse represents a response containing user information
type UserResponse struct {
	User models.User `json:"user"`
}

// UserUpdateResponse represents a response after updating a user
type UserUpdateResponse struct {
	Message string     `json:"message" example:"User updated successfully"`
	User    models.User `json:"user"`
}

// UpdateUserRequest represents the request body for updating a user
type UpdateUserRequest struct {
	Name        string                 `json:"name" example:"John Doe"`
	Mobile      string                 `json:"mobile" example:"5551234567"`
	CountryCode string                 `json:"country_code" example:"+1"`
	Address     string                 `json:"address" example:"123 Main St"`
	City        string                 `json:"city" example:"New York"`
	Country     string                 `json:"country" example:"US"`
	PostalCode  string                 `json:"postal_code" example:"10001"`
	PaymentInfo map[string]interface{} `json:"payment_info" swaggertype:"object,string" example:"{\"card_type\":\"visa\"}"`
}

// GetUser handles retrieving a user's profile
// @Summary Get user profile
// @Description Get user details by ID (must be authenticated)
// @Tags users
// @Accept json
// @Produce json
// @Param id path string true "User ID"
// @Success 200 {object} UserResponse "User details"
// @Failure 400 {object} ErrorResponse "Bad Request - Invalid ID"
// @Failure 401 {object} ErrorResponse "Unauthorized"
// @Failure 403 {object} ErrorResponse "Forbidden - Trying to access other user's profile"
// @Failure 404 {object} ErrorResponse "Not Found - User not found"
// @Failure 500 {object} ErrorResponse "Internal Server Error"
// @Security BearerAuth
// @Router /user/{id} [get]
func GetUser(c *gin.Context) {
	// Get authenticated user ID
	authenticatedUserID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "Unauthorized"})
		return
	}
	
	// Get requested user ID from path
	userID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid user ID"})
		return
	}
	
	// Check if user is requesting their own profile
	if authenticatedUserID.(uint) != uint(userID) {
		c.JSON(http.StatusForbidden, ErrorResponse{Error: "You can only view your own profile"})
		return
	}
	
	// Fetch user from database
	user, err := models.FindUserByID(database.DB, uint(userID))
	if err != nil {
		c.JSON(http.StatusNotFound, ErrorResponse{Error: "User not found"})
		return
	}
	
	// Clear sensitive fields
	user.PasswordHash = ""
	
	c.JSON(http.StatusOK, UserResponse{User: *user})
}

// UpdateUser handles updating a user's profile
// @Summary Update user profile
// @Description Update user details (must be authenticated and own profile)
// @Tags users
// @Accept json
// @Produce json
// @Param id path string true "User ID"
// @Param user body UpdateUserRequest true "User details to update"
// @Success 200 {object} UserUpdateResponse "Updated user details"
// @Failure 400 {object} ErrorResponse "Bad Request - Invalid input"
// @Failure 401 {object} ErrorResponse "Unauthorized"
// @Failure 403 {object} ErrorResponse "Forbidden - Trying to update other user's profile"
// @Failure 404 {object} ErrorResponse "Not Found - User not found"
// @Failure 500 {object} ErrorResponse "Internal Server Error"
// @Security BearerAuth
// @Router /user/{id}/update [put]
func UpdateUser(c *gin.Context) {
	// Get authenticated user ID
	authenticatedUserID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "Unauthorized"})
		return
	}
	
	// Get requested user ID from path
	userID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid user ID"})
		return
	}
	
	// Check if user is updating their own profile
	if authenticatedUserID.(uint) != uint(userID) {
		c.JSON(http.StatusForbidden, ErrorResponse{Error: "You can only update your own profile"})
		return
	}
	
	// Parse update request
	var req UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}
	
	// Fetch user from database
	user, err := models.FindUserByID(database.DB, uint(userID))
	if err != nil {
		c.JSON(http.StatusNotFound, ErrorResponse{Error: "User not found"})
		return
	}
	
	// Update fields if provided
	if req.Name != "" {
		user.Name = req.Name
	}
	if req.Mobile != "" {
		user.Mobile = req.Mobile
	}
	if req.CountryCode != "" {
		user.CountryCode = req.CountryCode
	}
	if req.Address != "" {
		user.Address = req.Address
	}
	if req.City != "" {
		user.City = req.City
	}
	if req.Country != "" {
		user.Country = req.Country
	}
	if req.PostalCode != "" {
		user.PostalCode = req.PostalCode
	}
	if req.PaymentInfo != nil {
		// // Convert map to JSON
		// paymentInfoJSON, err := database.DB.Dialector.Translate(req.PaymentInfo)
		// if err != nil {
		// 	c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid payment info format"})
		// 	return
		// }
		// user.PaymentInfo = paymentInfoJSON
	}
	
	// Save to database
	if err := database.DB.Save(user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to update user"})
		return
	}
	
	// Clear sensitive fields
	user.PasswordHash = ""
	
	c.JSON(http.StatusOK, UserUpdateResponse{
		Message: "User updated successfully",
		User:    *user,
	})
}