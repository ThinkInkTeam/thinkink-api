package models

import (
	"gorm.io/gorm"
	"time"
)

type BlacklistedToken struct {
	gorm.Model
	Token     string    `gorm:"uniqueIndex;not null"`
	ExpiresAt time.Time `gorm:"not null"`
}