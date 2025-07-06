package handlers

import (
	"net/http"
	"strconv"

	"github.com/ThinkInkTeam/thinkink-core-backend/database"
	"github.com/ThinkInkTeam/thinkink-core-backend/models"
	"github.com/gin-gonic/gin"
)

// ReportsResponse represents a response containing a list of reports
type ReportsResponse struct {
	Reports []models.Report `json:"reports"`
}

// SortedReportsResponse represents a response containing sorted reports
type SortedReportsResponse struct {
	Reports []models.Report `json:"reports"`
	Sorting SortingInfo     `json:"sorting"`
}

// SortingInfo represents sorting information
type SortingInfo struct {
	Field string `json:"field" example:"matching_scale"`
	Order string `json:"order" example:"descending"`
}

// GetUserReports retrieves all reports for the authenticated user
// @Summary Get all user reports
// @Description Retrieves all reports belonging to the authenticated user
// @Tags reports
// @Produce json
// @Success 200 {object} ReportsResponse "List of user reports"
// @Failure 401 {object} ErrorResponse "Unauthorized"
// @Failure 500 {object} ErrorResponse "Internal Server Error"
// @Security BearerAuth
// @Router /reports [get]
func GetUserReports(c *gin.Context) {
	// Get authenticated user ID
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "Unauthorized"})
		return
	}

	// Fetch user from database
	user, err := models.FindUserByID(database.DB, userID.(uint))
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to fetch user"})
		return
	}

	// Get all reports for the user
	reports, err := user.FindAllUserReports(database.DB)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to fetch reports"})
		return
	}

	c.JSON(http.StatusOK, ReportsResponse{
		Reports: reports,
	})
}

// GetUserReportsSortedByScale retrieves all reports for the authenticated user sorted by matching scale
// @Summary Get user reports sorted by matching scale
// @Description Retrieves all reports belonging to the authenticated user, sorted by matching scale
// @Tags reports
// @Produce json
// @Param asc query string false "Sort ascending (true) or descending (false, default)"
// @Success 200 {object} SortedReportsResponse "List of user reports sorted by matching scale"
// @Failure 401 {object} ErrorResponse "Unauthorized"
// @Failure 500 {object} ErrorResponse "Internal Server Error"
// @Security BearerAuth
// @Router /reports/sorted [get]
func GetUserReportsSortedByScale(c *gin.Context) {
	// Get authenticated user ID
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "Unauthorized"})
		return
	}

	// Parse sort direction from query parameter
	ascendingParam := c.DefaultQuery("asc", "false")
	ascending, err := strconv.ParseBool(ascendingParam)
	if err != nil {
		ascending = false // Default to descending (highest matching scale first)
	}

	// Fetch user from database
	user, err := models.FindUserByID(database.DB, userID.(uint))
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to fetch user"})
		return
	}

	// Get reports sorted by matching scale
	reports, err := user.FindAllUserReportsSortedByScale(database.DB, ascending)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to fetch sorted reports"})
		return
	}

	orderText := "descending"
	if ascending {
		orderText = "ascending"
	}

	c.JSON(http.StatusOK, SortedReportsResponse{
		Reports: reports,
		Sorting: SortingInfo{
			Field: "matching_scale",
			Order: orderText,
		},
	})
}
