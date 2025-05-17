package models

import (
	"time"

	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// Report defines the structure for an API report
type Report struct {
	ID            uint           `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID        uint           `gorm:"not null" json:"user_id"`
	Title         string         `gorm:"type:varchar(255);not null" json:"title"`
	Description   string         `gorm:"type:text" json:"description"`
	Content       datatypes.JSON `gorm:"type:json" json:"content" swaggertype:"string" example:"{\"key\":\"value\"}"`
	CreatedAt     time.Time      `gorm:"type:timestamp;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt     time.Time      `gorm:"type:timestamp;default:CURRENT_TIMESTAMP" json:"updated_at"`
	MatchingScale int        `gorm:"type:numeric(5,2);default:0" json:"matching_scale"`
}

// BeforeSave automatically updates the UpdatedAt field
func (r *Report) BeforeSave(tx *gorm.DB) (err error) {
	r.UpdatedAt = time.Now()
	return
}

// FindReportsByUserID gets all reports for a user
func FindReportsByUserID(db *gorm.DB, userID uint) ([]Report, error) {
	var reports []Report
	result := db.Where("user_id = ?", userID).Find(&reports)
	return reports, result.Error
}

// CreateReport creates a new report directly with the provided data
func (r *Report) CreateReport(db *gorm.DB, userID uint) (*Report, error) {
	if err := db.Create(r).Error; err != nil {
		return nil, err
	}

	return r, nil
}
