package models

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"time"

	"github.com/ThinkInkTeam/thinkink-core-backend/utils"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stripe/stripe-go/v72"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type User struct {
	ID           uint           `gorm:"primaryKey;autoIncrement" json:"id"`
	Name         string         `gorm:"type:text;not null" json:"name"`
	Email        string         `gorm:"type:text;unique;not null" json:"email"`
	PasswordHash string         `gorm:"type:text;not null" json:"password"`
	DateOfBirth  time.Time      `gorm:"type:date;not null" json:"date_of_birth"`
	Mobile       string         `gorm:"type:varchar(15)" json:"mobile,omitempty"`
	CountryCode  string         `gorm:"type:varchar(5)" json:"country_code,omitempty"`
	Address      string         `gorm:"type:text" json:"address,omitempty"`
	City         string         `gorm:"type:text" json:"city,omitempty"`
	Country      string         `gorm:"type:text" json:"country,omitempty"`
	PostalCode   string         `gorm:"type:text" json:"postal_code,omitempty"`
	PaymentInfo  datatypes.JSON `gorm:"type:json" json:"payment_info,omitempty" swaggertype:"string" example:"{\"card_type\":\"visa\"}"`
	CreatedAt    time.Time      `gorm:"type:timestamp;default:CURRENT_TIMESTAMP" json:"created_at"`
	LastLogin    *time.Time     `gorm:"type:timestamp" json:"last_login,omitempty"`
	Reports      []Report       `gorm:"foreignKey:UserID" json:"reports"`
	// Stripe fields
	StripeCustomerID   *string    `gorm:"type:text;uniqueIndex" json:"stripe_customer_id,omitempty"`
	StripeDefaultPM    *string    `gorm:"type:text" json:"stripe_default_payment_method,omitempty"`
	CurrentPlanID      *string    `gorm:"type:text" json:"current_plan_id,omitempty"`
	SubscriptionID     *string    `gorm:"type:text" json:"subscription_id,omitempty"`
	SubscriptionStatus *string    `gorm:"type:text" json:"subscription_status,omitempty"`
	SubscriptionEndsAt *time.Time `gorm:"type:timestamp" json:"subscription_ends_at,omitempty"`
}

// New function for Stripe integration

// ToStripeCustomerParams converts user data to Stripe customer parameters
func (u *User) ToStripeCustomerParams() *stripe.CustomerParams {
	params := &stripe.CustomerParams{
		Name:  stripe.String(u.Name),
		Email: stripe.String(u.Email),
	}

	if u.Mobile != "" && u.CountryCode != "" {
		phone := u.CountryCode + u.Mobile
		params.Phone = stripe.String(phone)
	}

	if u.Address != "" && u.City != "" && u.Country != "" {
		params.Address = &stripe.AddressParams{
			Line1:      stripe.String(u.Address),
			City:       stripe.String(u.City),
			Country:    stripe.String(u.Country),
			PostalCode: stripe.String(u.PostalCode),
		}
	}

	return params
}

// UpdateStripeData updates the Stripe-related user data
func (u *User) UpdateStripeData(db *gorm.DB, customerID string, defaultPM string) error {
	u.StripeCustomerID = &customerID
	u.StripeDefaultPM = &defaultPM
	return db.Model(u).Updates(map[string]interface{}{
		"stripe_customer_id": customerID,
		"stripe_default_pm":  defaultPM,
	}).Error
}

// UpdateSubscriptionData updates the subscription data for the user
func (u *User) UpdateSubscriptionData(db *gorm.DB, subscriptionID, planID, status string, endsAt *time.Time) error {
	u.SubscriptionID = &subscriptionID
	u.CurrentPlanID = &planID
	u.SubscriptionStatus = &status
	u.SubscriptionEndsAt = endsAt

	return db.Model(u).Updates(map[string]interface{}{
		"subscription_id":      subscriptionID,
		"current_plan_id":      planID,
		"subscription_status":  status,
		"subscription_ends_at": endsAt,
	}).Error
}

// IsSubscribed checks if the user has an active subscription
func (u *User) IsSubscribed() bool {
	if u.SubscriptionStatus == nil {
		return false
	}
	return *u.SubscriptionStatus == "active" || *u.SubscriptionStatus == "trialing"
}

// Original User functions
func (u *User) BeforeCreate(tx *gorm.DB) (err error) {
	var existingUser User
	if err := tx.Where("email = ?", u.Email).First(&existingUser).Error; err == nil {
		return fmt.Errorf("email already exists")
	}

	return nil
}

func (u *User) BeforeSave(tx *gorm.DB) (err error) {
	var existingUser User
	result := tx.Where("email = ?", u.Email).First(&existingUser)

	if u.ID == 0 && result.Error == nil {
		return fmt.Errorf("email already exists")
	}

	if u.ID != 0 && result.Error == nil && existingUser.ID != u.ID {
		return fmt.Errorf("email already exists")
	}
	return nil
}

// ValidatePassword checks if the provided password matches the stored hash
func (u *User) ValidatePassword(password string) error {
	return bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password))
}

// HashPassword generates a bcrypt hash from a plain-text password
func (u *User) HashPassword(password string) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	u.PasswordHash = string(hashedPassword)
	return nil
}

// GenerateJWT creates a JWT token for the user
func (u *User) GenerateJWT() (string, error) {
	// Set JWT expiration to 24 hours
	expirationTime := time.Now().Add(24 * time.Hour)

	claims := jwt.MapClaims{
		"userID": u.ID,
		"email":  u.Email,
		"exp":    expirationTime.Unix(),
	}

	// Get JWT secret from environment variable or use a default for development
	jwtSecret := utils.GetEnvWithDefault("JWT_SECRET", "your_jwt_secret")

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(jwtSecret))

	return tokenString, err
}

// UpdateLastLogin updates the user's last login timestamp
func (u *User) UpdateLastLogin(db *gorm.DB) error {
	now := time.Now()
	u.LastLogin = &now
	return db.Model(u).Update("last_login", now).Error
}

// CreateUser creates a new user in the database with the provided information
func CreateUser(db *gorm.DB, name, email, password string, dateOfBirth time.Time, mobile, countryCode, address, city, country, postalCode string, paymentInfo map[string]interface{}) (*User, error) {
	// Check if user with email already exists
	var existingUser User
	if err := db.Where("email = ?", email).First(&existingUser).Error; err == nil {
		return nil, fmt.Errorf("email already exists")
	} else if err != gorm.ErrRecordNotFound {
		return nil, err
	}

	// Create payment info JSON if provided
	var paymentInfoJSON datatypes.JSON
	if paymentInfo != nil {
		paymentInfoBytes, err := json.Marshal(paymentInfo)
		if err != nil {
			return nil, fmt.Errorf("error encoding payment info: %w", err)
		}
		paymentInfoJSON = datatypes.JSON(paymentInfoBytes)
	}

	// Create new user
	user := &User{
		Name:        name,
		Email:       email,
		DateOfBirth: dateOfBirth,
		Mobile:      mobile,
		CountryCode: countryCode,
		Address:     address,
		City:        city,
		Country:     country,
		PostalCode:  postalCode,
		PaymentInfo: paymentInfoJSON,
		CreatedAt:   time.Now(),
	}

	// Hash the password
	if err := user.HashPassword(password); err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// Save to database
	if err := db.Create(user).Error; err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return user, nil
}

// FindAllUserReports retrieves all reports belonging to the user
func (u *User) FindAllUserReports(db *gorm.DB) ([]Report, error) {
	var reports []Report

	err := db.Where("user_id = ?", u.ID).Find(&reports).Error
	if err != nil {
		return nil, fmt.Errorf("failed to fetch reports: %w", err)
	}

	return reports, nil
}

// FindAllUserReportsSortedByScale retrieves all reports belonging to the user sorted by matching scale
func (u *User) FindAllUserReportsSortedByScale(db *gorm.DB, ascending bool) ([]Report, error) {
	var reports []Report

	query := db.Where("user_id = ?", u.ID)

	if ascending {
		query = query.Order("matching_scale asc")
	} else {
		query = query.Order("matching_scale desc")
	}

	err := query.Find(&reports).Error
	if err != nil {
		return nil, fmt.Errorf("failed to fetch sorted reports: %w", err)
	}

	return reports, nil
}

// FindUserByID retrieves a user by their ID
func FindUserByID(db *gorm.DB, id uint) (*User, error) {
	var user User
	if err := db.First(&user, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("database error: %w", err)
	}
	return &user, nil
}

// FindUserByEmail retrieves a user by their email address
func FindUserByEmail(db *gorm.DB, email string) (*User, error) {
	var user User
	if err := db.Where("email = ?", email).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("database error: %w", err)
	}
	return &user, nil
}

// PasswordReset represents a password reset request
type PasswordReset struct {
	gorm.Model
	UserID    uint      `gorm:"not null"`
	Token     string    `gorm:"uniqueIndex;not null"`
	ExpiresAt time.Time `gorm:"not null"`
	Used      bool      `gorm:"default:false"`
}

// GeneratePasswordResetToken creates a token for password reset
func (u *User) GeneratePasswordResetToken(db *gorm.DB) (string, error) {
	// Generate random token
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		return "", fmt.Errorf("error generating token: %w", err)
	}

	token := base64.URLEncoding.EncodeToString(b)

	// Set expiration to 1 hour from now
	expiresAt := time.Now().Add(1 * time.Hour)

	// Create password reset record
	reset := PasswordReset{
		UserID:    u.ID,
		Token:     token,
		ExpiresAt: expiresAt,
		Used:      false,
	}

	// Save to database
	if err := db.Create(&reset).Error; err != nil {
		return "", fmt.Errorf("error saving reset token: %w", err)
	}

	return token, nil
}

// VerifyPasswordResetToken validates a reset token and returns the associated user
func VerifyPasswordResetToken(db *gorm.DB, token string) (*User, error) {
	var reset PasswordReset
	if err := db.Where("token = ? AND used = ? AND expires_at > ?", token, false, time.Now()).First(&reset).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("invalid or expired token")
		}
		return nil, fmt.Errorf("database error: %w", err)
	}

	// Mark token as used
	if err := db.Model(&reset).Updates(map[string]interface{}{"used": true}).Error; err != nil {
		return nil, fmt.Errorf("error updating token: %w", err)
	}

	// Get associated user
	user, err := FindUserByID(db, reset.UserID)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	return user, nil
}

// UpdatePassword changes a user's password
func (u *User) UpdatePassword(db *gorm.DB, newPassword string) error {
	// Hash the new password
	if err := u.HashPassword(newPassword); err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	// Update in database
	return db.Model(u).Update("password_hash", u.PasswordHash).Error
}
