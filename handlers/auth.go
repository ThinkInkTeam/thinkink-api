package handlers

import (
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/ThinkInkTeam/thinkink-core-backend/database"
	"github.com/ThinkInkTeam/thinkink-core-backend/models"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type RegistrationInput struct {
	Name        string                 `json:"name" binding:"required"`
	Email       string                 `json:"email" binding:"required,email"`
	Password    string                 `json:"password" binding:"required"`
	DateOfBirth string                 `json:"date_of_birth" binding:"required"`
	Mobile      *string                `json:"mobile,omitempty"`
	CountryCode *string                `json:"country_code,omitempty"`
	Address     *string                `json:"address,omitempty"`
	City        *string                `json:"city,omitempty"`
	Country     *string                `json:"country,omitempty"`
	PostalCode  *string                `json:"postal_code,omitempty"`
	PaymentInfo map[string]interface{} `json:"payment_info,omitempty"`
}

func Register(c *gin.Context) {
	var input RegistrationInput

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	dob, err := time.Parse("2006-01-02", input.DateOfBirth)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid date format (YYYY-MM-DD)"})
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Password hashing failed"})
		return
	}

	user := models.User{
		Name:        input.Name,
		Email:       input.Email,
		PasswordHash:    string(hashedPassword),
		DateOfBirth: dob,
		Mobile:      input.Mobile,
		CountryCode: input.CountryCode,
		Address:     input.Address,
		City:        input.City,
		Country:     input.Country,
		PostalCode:  input.PostalCode,
	}

	if err := database.DB.Create(&user).Error; err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "User registered successfully"})
}

func Login(c *gin.Context) {
	var credentials struct {
		Email    string `json:"email" binding:"required,email"`
		Password string `json:"password" binding:"required"`
	}

	if err := c.ShouldBindJSON(&credentials); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var user models.User
	if err := database.DB.Where("email = ?", credentials.Email).First(&user).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	if err := user.ValidatePassword(credentials.Password); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	now := time.Now()
	database.DB.Model(&user).Update("last_login", now)

	c.JSON(http.StatusOK, gin.H{
		"message": "Login successful",
		"user": gin.H{
			"id":    user.ID,
			"name":  user.Name,
			"email": user.Email,
		},
	})
}

type UpdateUserInput struct {
	Name        *string                `json:"name,omitempty"`
	Email       *string                `json:"email,omitempty" binding:"omitempty,email"`
	Password    *string                `json:"password,omitempty" binding:"omitempty,min=8"`
	Mobile      *string                `json:"mobile,omitempty"`
	CountryCode *string                `json:"country_code,omitempty"`
	Address     *string                `json:"address,omitempty"`
	City        *string                `json:"city,omitempty"`
	Country     *string                `json:"country,omitempty"`
	PostalCode  *string                `json:"postal_code,omitempty"`
	PaymentInfo map[string]interface{} `json:"payment_info,omitempty"`
}

func UpdateUser(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var input UpdateUserInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var user models.User
	if err := database.DB.First(&user, userID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	if input.Email != nil {
		user.Email = *input.Email
	}
	if input.Name != nil {
		user.Name = *input.Name
	}
	if input.Mobile != nil {
		user.Mobile = input.Mobile
	}

	if input.Password != nil {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(*input.Password), bcrypt.DefaultCost)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Password hashing failed"})
			return
		}
		user.PasswordHash = string(hashedPassword)
	}

	if err := database.DB.Save(&user).Error; err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User updated successfully"})
}

func Logout(c *gin.Context) {
	authHeader := c.GetHeader("Authorization")
	tokenString := strings.TrimPrefix(authHeader, "Bearer ")

	token, _ := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return []byte(os.Getenv("JWT_SECRET")), nil
	})

	claims := token.Claims.(jwt.MapClaims)
	exp := time.Unix(int64(claims["exp"].(float64)), 0)

	blacklistedToken := models.BlacklistedToken{
		Token:     tokenString,
		ExpiresAt: exp,
	}

	if err := database.DB.Create(&blacklistedToken).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Logout failed"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Logged out successfully"})
}
