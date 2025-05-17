package models

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"gorm.io/datatypes"
)

// SingleFile represents a temporarily uploaded file that will be processed into a Report
type SingleFile struct {
	ID          uint      `json:"id"`
	UserID      uint      `json:"user_id"`
	Filename    string    `json:"filename"`
	FilePath    string    `json:"file_path"`
	UploadedAt  time.Time `json:"uploaded_at"`
	FileSize    int64
	Description string `json:"description"`
}

// ConvertToReport reads the file, parses the JSON content into a Report object and returns it
// Does not save to database
func (sf *SingleFile) ConvertToReport() (*Report, error) {
	// Read the file content
	fileData, err := os.ReadFile(sf.FilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}
	
	// Attempt to parse the JSON
	var jsonData map[string]interface{}
	if err := json.Unmarshal(fileData, &jsonData); err != nil {
		return nil, fmt.Errorf("invalid JSON format: %w", err)
	}
	content, err := json.Marshal(jsonData)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal JSON: %w", err)
	}
	
	// Create and return the report without saving to database
	report := &Report{
		UserID:			sf.UserID,
		Title:			sf.Filename,
		Description: 		sf.Description,
		Content:       	datatypes.JSON(content),
		MatchingScale: 	0,
		CreatedAt:  	time.Now(),
	}
	
	return report, nil
}

// CreateSingleFile creates a new single file entry from a file path
func CreateSingleFile(userID uint, originalFilename, filePath, description string) (*SingleFile, error) {
	// Check if file exists
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return nil, fmt.Errorf("file error: %w", err)
	}
	
	singleFile := &SingleFile{
		UserID:     userID,
		Filename:   originalFilename,
		FilePath:   filePath,
		Description: description,
		UploadedAt: time.Now(),
		FileSize:   fileInfo.Size(),
	}
	
	return singleFile, nil
}

