package models

import (
	"time"
)

type SignalFile struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	UserID     uint      `gorm:"not null" json:"user_id"`
	Filename   string    `gorm:"not null" json:"filename"`
	FilePath   string    `gorm:"not null" json:"file_path"`
	UploadedAt time.Time `gorm:"default:CURRENT_TIMESTAMP" json:"uploaded_at"`
	Status     string    `gorm:"default:'pending'" json:"status"`
}
