package models

import (
	"time"

	"gorm.io/gorm"
)

type BlacklistedToken struct {
	gorm.Model
	Token     string    `gorm:"uniqueIndex;not null"`
	ExpiresAt time.Time `gorm:"not null"`
}

// IsExpired checks if the token has expired
func (bt *BlacklistedToken) IsExpired() bool {
	return time.Now().After(bt.ExpiresAt)
}

// AddToBlacklist adds a token to the blacklist
func AddToBlacklist(db *gorm.DB, token string, expiresAt time.Time) error {
	blacklistedToken := BlacklistedToken{
		Token:     token,
		ExpiresAt: expiresAt,
	}
	return db.Create(&blacklistedToken).Error
}

// IsTokenBlacklisted checks if a token is in the blacklist
func IsTokenBlacklisted(db *gorm.DB, token string) (bool, error) {
	var blacklistedToken BlacklistedToken
	err := db.Where("token = ?", token).First(&blacklistedToken).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

// CleanupExpiredTokens removes all expired tokens from the blacklist
func CleanupExpiredTokens(db *gorm.DB) error {
	return db.Where("expires_at < ?", time.Now()).Delete(&BlacklistedToken{}).Error
}
