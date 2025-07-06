package handlers

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/ThinkInkTeam/thinkink-core-backend/database"
	"github.com/ThinkInkTeam/thinkink-core-backend/models"
	"github.com/ThinkInkTeam/thinkink-core-backend/services"
	"github.com/google/uuid"

	"net/http"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
)

const (
	UploadDir     = "./uploads"
	MaxUploadSize = 50 << 20
)

// FileUploadResponse represents a successful file upload response
type FileUploadResponse struct {
	Message       string `json:"message" example:"File processed successfully"`
	FileID        uint   `json:"file_id" example:"1"`
	ReportID      uint   `json:"report_id" example:"2"`
	Description   string `json:"description" example:"Sample brain activity data"`
	MatchingScale int    `json:"matching_scale" example:"7"`
}

// UploadSignalFile handles the upload of signal files.
// @Summary Upload a signal file
// @Description Uploads a signal file and stores metadata in the database with matching scale
// @Tags files
// @Accept multipart/form-data
// @Produce json
// @Param file formData file true "File to upload"
// @Param matchingScale formData int false "Matching scale (1-10)" default(5)
// @Param description formData string false "Description of the file" default("")
// @Success 200 {object} FileUploadResponse "File uploaded successfully"
// @Failure 400 {object} ErrorResponse "Bad Request - No file uploaded, file too large, or invalid matching scale"
// @Failure 401 {object} ErrorResponse "Unauthorized"
// @Failure 500 {object} ErrorResponse "Internal Server Error"
// @Security BearerAuth
// @Router /upload [post]
func UploadSignalFile(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "Unauthorized"})
		return
	}

	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, MaxUploadSize)
	if err := c.Request.ParseMultipartForm(MaxUploadSize); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "File too large (max 50MB)"})
		return
	}

	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "No file uploaded"})
		return
	}

	// Get matching scale from form, default to 5 if not provided
	matchingScaleStr := c.DefaultPostForm("matchingScale", "5")
	matchingScale, err := strconv.Atoi(matchingScaleStr)
	if err != nil || matchingScale < 1 || matchingScale > 10 {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Matching scale must be between 1 and 10"})
		return
	}

	if err := os.MkdirAll(UploadDir, os.ModePerm); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Could not create upload directory"})
		return
	}

	ext := filepath.Ext(file.Filename)
	filename := fmt.Sprintf("%d-%s%s", userID, uuid.New().String(), ext)
	filePath := filepath.Join(UploadDir, filename)

	if err := c.SaveUploadedFile(file, filePath); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to save file"})
		return
	}

	// Get description from form, default to empty string if not provided
	description := ""

	// If no description provided, try to get translation from ML server
	if description == "" {
		if authHeader := c.GetHeader("Authorization"); authHeader != "" {
			// Connect to translation service
			translationClient, err := services.NewTranslationClient("ml-service:50052")
			if err == nil {
				defer translationClient.Close()
				fileData, err := os.ReadFile(filePath)
				if err != nil {
					c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to read file"})
					return
				}
				// Get translation using the file data
				translations, err := translationClient.TranslateEEGFromBytes(authHeader, fileData)
				if err == nil && len(translations) > 0 {
					description = strings.Join(translations, " ")
				}
			}
		}
	}

	signalFile, err := models.CreateSingleFile(
		userID.(uint),
		file.Filename,
		filePath,
		description,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to process file: " + err.Error()})
		return
	}

	// Convert the file to a report
	report, err := signalFile.ConvertToReport()
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Failed to convert file to report: " + err.Error()})
		// Clean up the file
		_ = os.Remove(filePath)
		return
	}

	// Set the matching scale provided by the user
	report.MatchingScale = matchingScale

	// Use the CreateReport method to save the report to the database
	savedReport, err := report.CreateReport(database.DB, userID.(uint))
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to save report: " + err.Error()})
		// Clean up the file
		_ = os.Remove(filePath)
		return
	}

	c.JSON(http.StatusOK, FileUploadResponse{
		Message:       "File processed successfully",
		FileID:        signalFile.ID,
		ReportID:      savedReport.ID,
		Description:   signalFile.Description,
		MatchingScale: savedReport.MatchingScale,
	})
}
