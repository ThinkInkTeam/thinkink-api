package models

import (
	"fmt"
	"time"

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
	Mobile       string        `gorm:"type:varchar(15)" json:"mobile,omitempty"`
	CountryCode  string        `gorm:"type:varchar(5)" json:"country_code,omitempty"`
	Address      string        `gorm:"type:text" json:"address,omitempty"`
	City         string        `gorm:"type:text" json:"city,omitempty"`
	Country      string        `gorm:"type:text" json:"country,omitempty"`
	PostalCode   string        `gorm:"type:text" json:"postal_code,omitempty"`
	PaymentInfo  datatypes.JSON `gorm:"type:json" json:"payment_info,omitempty" swaggertype:"string"`
	CreatedAt    time.Time      `gorm:"type:timestamp;default:CURRENT_TIMESTAMP" json:"created_at"`
	LastLogin    *time.Time     `gorm:"type:timestamp" json:"last_login,omitempty"`
	Reports      []Report       `gorm:"foreignKey:UserID" json:"reports"`
}

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

func (u *User) ValidatePassword(password string) error {
	return bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password))
}
