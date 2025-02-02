package handlers

import (
	"fmt"

	"github.com/ThinkInkTeam/thinkink-core-backend/models"
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

// UploadSignalFile handles the upload of signal files.
//
// @Summary Upload a signal file
// @Description Uploads a signal file and stores metadata in the database
// @Tags Signal Files
// @Accept multipart/form-data
// @Produce json
// @Param file formData file true "File to upload"
// @Success 200 {object} map[string]interface{} "File uploaded successfully"
// @Failure 400 {object} map[string]string "Bad Request - No file uploaded or file too large"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 500 {object} map[string]string "Internal Server Error - Could not create upload directory or failed to save file"
// @Security BearerAuth
// @Router /upload [post]

func UploadSignalFile(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, MaxUploadSize)
	if err := c.Request.ParseMultipartForm(MaxUploadSize); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "File too large (max 50MB)"})
		return
	}

	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No file uploaded"})
		return
	}

	if err := os.MkdirAll(UploadDir, os.ModePerm); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not create upload directory"})
		return
	}

	ext := filepath.Ext(file.Filename)
	filename := fmt.Sprintf("%d-%s%s", userID, uuid.New().String(), ext)
	filePath := filepath.Join(UploadDir, filename)

	if err := c.SaveUploadedFile(file, filePath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save file"})
		return
	}

	signalFile := models.SignalFile{
		UserID:   userID.(uint),
		Filename: file.Filename,
		FilePath: filePath,
		Status:   "pending",
	}

	// if err := database.DB.Create(&signalFile).Error; err != nil {
	// 	c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to record file"})
	// 	return
	// }

	c.JSON(http.StatusOK, gin.H{
		"message": "File uploaded successfully",
		"file_id": signalFile.ID,
		"status":  signalFile.Status,
	})
}
