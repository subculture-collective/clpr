package handlers

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"git.subcult.tv/subculture-collective/clpr/internal/models"
	"git.subcult.tv/subculture-collective/clpr/internal/repository"
	"git.subcult.tv/subculture-collective/clpr/internal/services"
)

// ExportHandler handles data export HTTP requests
type ExportHandler struct {
	exportService *services.ExportService
	authService   *services.AuthService
	userRepo      *repository.UserRepository
}

// NewExportHandler creates a new export handler
func NewExportHandler(exportService *services.ExportService, authService *services.AuthService, userRepo *repository.UserRepository) *ExportHandler {
	return &ExportHandler{
		exportService: exportService,
		authService:   authService,
		userRepo:      userRepo,
	}
}

// RequestExport creates a new export request for the authenticated user
// POST /api/v1/creators/me/export/request
func (h *ExportHandler) RequestExport(c *gin.Context) {
	// Get authenticated user
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	user, err := h.userRepo.GetByID(c.Request.Context(), userID.(uuid.UUID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get user"})
		return
	}

	// Parse request body
	var req models.CreateExportRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request", "details": err.Error()})
		return
	}

	// Create export request using the user's Twitch username as creator name
	exportReq, err := h.exportService.CreateExportRequest(
		c.Request.Context(),
		user.ID,
		user.Username,
		req.Format,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create export request", "details": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"export_request": exportReq,
		"message":        "Export request created successfully. You will receive an email when it's ready.",
	})
}

// GetExportStatus retrieves the status of an export request
// GET /api/v1/creators/me/export/status/:id
func (h *ExportHandler) GetExportStatus(c *gin.Context) {
	// Get authenticated user
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	// Parse export ID
	exportIDStr := c.Param("id")
	exportID, err := uuid.Parse(exportIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid export ID"})
		return
	}

	// Get export request
	exportReq, err := h.exportService.GetExportRequest(c.Request.Context(), exportID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "export request not found"})
		return
	}

	// Verify ownership
	if exportReq.UserID != userID.(uuid.UUID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
		return
	}

	// Build response with download URL if completed
	response := models.ExportRequestResponse{
		ExportRequest: *exportReq,
	}

	if exportReq.Status == models.ExportStatusCompleted && exportReq.ExpiresAt != nil {
		baseURL := c.GetString("base_url")
		if baseURL == "" {
			baseURL = "http://localhost:8080"
		}
		downloadURL := fmt.Sprintf("%s/api/v1/creators/me/export/download/%s", baseURL, exportReq.ID)
		response.DownloadURL = &downloadURL
	}

	c.JSON(http.StatusOK, response)
}

// DownloadExport downloads a completed export file
// GET /api/v1/creators/me/export/download/:id
func (h *ExportHandler) DownloadExport(c *gin.Context) {
	// Get authenticated user
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	// Parse export ID
	exportIDStr := c.Param("id")
	exportID, err := uuid.Parse(exportIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid export ID"})
		return
	}

	// Get export request
	exportReq, err := h.exportService.GetExportRequest(c.Request.Context(), exportID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "export request not found"})
		return
	}

	// Verify ownership
	if exportReq.UserID != userID.(uuid.UUID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
		return
	}

	// Check if export is completed
	if exportReq.Status != models.ExportStatusCompleted {
		c.JSON(http.StatusBadRequest, gin.H{"error": "export is not ready yet", "status": exportReq.Status})
		return
	}

	// Get file path
	filePath, err := h.exportService.GetExportFilePath(c.Request.Context(), exportID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "export file not found", "details": err.Error()})
		return
	}

	// Set appropriate content type and headers
	var contentType string
	var filename string
	switch exportReq.Format {
	case models.ExportFormatCSV:
		contentType = "text/csv"
		filename = fmt.Sprintf("clips_export_%s.csv", exportReq.CreatorName)
	case models.ExportFormatJSON:
		contentType = "application/json"
		filename = fmt.Sprintf("clips_export_%s.json", exportReq.CreatorName)
	default:
		contentType = "application/octet-stream"
		filename = fmt.Sprintf("clips_export_%s", exportReq.CreatorName)
	}

	c.Header("Content-Type", contentType)
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))
	c.File(filePath)
}

// ListExportRequests lists all export requests for the authenticated user
// GET /api/v1/creators/me/exports
func (h *ExportHandler) ListExportRequests(c *gin.Context) {
	// Get authenticated user
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	// Get export requests
	exportReqs, err := h.exportService.GetUserExportRequests(c.Request.Context(), userID.(uuid.UUID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to retrieve export requests"})
		return
	}

	// Build response with download URLs
	baseURL := c.GetString("base_url")
	if baseURL == "" {
		baseURL = "http://localhost:8080"
	}

	responses := make([]models.ExportRequestResponse, len(exportReqs))
	for i, req := range exportReqs {
		responses[i] = models.ExportRequestResponse{
			ExportRequest: *req,
		}
		if req.Status == models.ExportStatusCompleted && req.ExpiresAt != nil {
			downloadURL := fmt.Sprintf("%s/api/v1/creators/me/export/download/%s", baseURL, req.ID)
			responses[i].DownloadURL = &downloadURL
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"export_requests": responses,
		"count":           len(responses),
	})
}
