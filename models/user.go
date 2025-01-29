package models

import (
	"gorm.io/datatypes"
	"gorm.io/gorm"
	"time"
)

type User struct {
	ID           uint           `gorm:"primaryKey;autoIncrement" json:"id"`
	Name         string         `gorm:"type:text;not null" json:"name"`
	Email        string         `gorm:"type:text;unique;not null" json:"email"`
	PasswordHash string         `gorm:"type:text;not null" json:"-"`
	DateOfBirth  time.Time      `gorm:"type:date;not null" json:"date_of_birth"`
	Mobile       *string        `gorm:"type:varchar(15)" json:"mobile,omitempty"`
	CountryCode  *string        `gorm:"type:varchar(5)" json:"country_code,omitempty"`
	Address      *string        `gorm:"type:text" json:"address,omitempty"`
	City         *string        `gorm:"type:text" json:"city,omitempty"`
	Country      *string        `gorm:"type:text" json:"country,omitempty"`
	PostalCode   *string        `gorm:"type:text" json:"postal_code,omitempty"`
	PaymentInfo  datatypes.JSON `gorm:"type:json" json:"payment_info,omitempty"`
	CreatedAt    time.Time      `gorm:"type:timestamp;default:CURRENT_TIMESTAMP" json:"created_at"`
	LastLogin    *time.Time     `gorm:"type:timestamp" json:"last_login,omitempty"`
	Role         string         `gorm:"type:text;default:'user'" json:"role"`
	Reports      []Report       `gorm:"foreignKey:UserID" json:"-"`
}
