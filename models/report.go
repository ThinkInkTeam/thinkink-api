package models

import (
	"gorm.io/datatypes"
	"gorm.io/gorm"
	"time"
)

type Report struct {
	ID            uint           `gorm:"primaryKey;autoIncrement" json:"id"`
	Content       string         `gorm:"type:text" json:"content"`
	Metadata      datatypes.JSON `gorm:"type:json" json:"metadata"`
	MatchingScale int            `gorm:"type:int" json:"matching_scale"`
	CreationDate  time.Time      `gorm:"type:timestamp;default:CURRENT_TIMESTAMP" json:"creation_date"`
	LastUpdated   time.Time      `gorm:"type:timestamp;default:CURRENT_TIMESTAMP" json:"last_updated"`
	UserID        uint           `gorm:"type:int" json:"user_id"`
}
