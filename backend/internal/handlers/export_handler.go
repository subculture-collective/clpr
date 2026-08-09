package handlers

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"time"

	"git.subcult.tv/subculture-collective/clpr/internal/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type exportService interface {
	CreateExportRequest(context.Context, uuid.UUID, string, string) (*models.ExportRequest, error)
	GetExportRequest(context.Context, uuid.UUID) (*models.ExportRequest, error)
	GetExportFilePath(context.Context, uuid.UUID) (string, error)
	GetUserExportRequests(context.Context, uuid.UUID) ([]*models.ExportRequest, error)
}

type exportUserRepository interface {
	GetByID(context.Context, uuid.UUID) (*models.User, error)
}

var unsafeExportFilename = regexp.MustCompile(`[^A-Za-z0-9._-]+`)

// ExportHandler handles data export HTTP requests
type ExportHandler struct {
	exportService exportService
	userRepo      exportUserRepository
}

// NewExportHandler creates a new export handler
func NewExportHandler(exportService exportService, userRepo exportUserRepository) *ExportHandler {
	return &ExportHandler{
		exportService: exportService,
		userRepo:      userRepo,
	}
}

func exportUserID(c *gin.Context) (uuid.UUID, bool) {
	value, exists := c.Get("user_id")
	if !exists {
		return uuid.Nil, false
	}
	userID, ok := value.(uuid.UUID)
	return userID, ok && userID != uuid.Nil
}

func exportDownloadURL(c *gin.Context, req *models.ExportRequest) *string {
	if req.Status != models.ExportStatusCompleted || req.ExpiresAt == nil || !req.ExpiresAt.After(time.Now()) {
		return nil
	}
	url := fmt.Sprintf("%s/api/v1/creators/me/export/download/%s", getBaseURL(c), req.ID)
	return &url
}

func sanitizeExportFilename(value string) string {
	value = unsafeExportFilename.ReplaceAllString(strings.TrimSpace(value), "_")
	value = strings.Trim(value, "._-")
	if value == "" {
		return "creator"
	}
	if len(value) > 80 {
		value = value[:80]
	}
	return value
}

// RequestExport creates a new export request for the authenticated user
// POST /api/v1/creators/me/export/request
func (h *ExportHandler) RequestExport(c *gin.Context) {
	// Get authenticated user
	userID, ok := exportUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	user, err := h.userRepo.GetByID(c.Request.Context(), userID)
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
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create export request"})
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
	userID, ok := exportUserID(c)
	if !ok {
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
		if errors.Is(err, pgx.ErrNoRows) {
			c.JSON(http.StatusNotFound, gin.H{"error": "export request not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to retrieve export request"})
		}
		return
	}

	// Verify ownership
	if exportReq.UserID != userID {
		c.JSON(http.StatusNotFound, gin.H{"error": "export request not found"})
		return
	}

	// Build response with download URL if completed
	response := models.ExportRequestResponse{
		ExportRequest: *exportReq,
	}

	response.DownloadURL = exportDownloadURL(c, exportReq)

	c.JSON(http.StatusOK, response)
}

// DownloadExport downloads a completed export file
// GET /api/v1/creators/me/export/download/:id
func (h *ExportHandler) DownloadExport(c *gin.Context) {
	// Get authenticated user
	userID, ok := exportUserID(c)
	if !ok {
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
		if errors.Is(err, pgx.ErrNoRows) {
			c.JSON(http.StatusNotFound, gin.H{"error": "export request not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to retrieve export request"})
		}
		return
	}

	// Verify ownership
	if exportReq.UserID != userID {
		c.JSON(http.StatusNotFound, gin.H{"error": "export request not found"})
		return
	}

	// Check if export is completed
	if exportReq.Status != models.ExportStatusCompleted {
		c.JSON(http.StatusBadRequest, gin.H{"error": "export is not ready yet", "status": exportReq.Status})
		return
	}
	if exportReq.ExpiresAt == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "export expiration is unavailable"})
		return
	}
	if !exportReq.ExpiresAt.After(time.Now()) {
		c.JSON(http.StatusGone, gin.H{"error": "export has expired"})
		return
	}

	// Get file path
	filePath, err := h.exportService.GetExportFilePath(c.Request.Context(), exportID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "export file not found"})
		return
	}

	// Set appropriate content type and headers
	var contentType string
	var filename string
	switch exportReq.Format {
	case models.ExportFormatCSV:
		contentType = "text/csv"
		filename = fmt.Sprintf("clips_export_%s.csv", sanitizeExportFilename(exportReq.CreatorName))
	case models.ExportFormatJSON:
		contentType = "application/json"
		filename = fmt.Sprintf("clips_export_%s.json", sanitizeExportFilename(exportReq.CreatorName))
	default:
		contentType = "application/octet-stream"
		filename = fmt.Sprintf("clips_export_%s", sanitizeExportFilename(exportReq.CreatorName))
	}

	c.Header("Content-Type", contentType)
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))
	c.File(filePath)
}

// ListExportRequests lists all export requests for the authenticated user
// GET /api/v1/creators/me/exports
func (h *ExportHandler) ListExportRequests(c *gin.Context) {
	// Get authenticated user
	userID, ok := exportUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	// Get export requests
	exportReqs, err := h.exportService.GetUserExportRequests(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to retrieve export requests"})
		return
	}

	responses := make([]models.ExportRequestResponse, len(exportReqs))
	for i, req := range exportReqs {
		responses[i] = models.ExportRequestResponse{
			ExportRequest: *req,
		}
		responses[i].DownloadURL = exportDownloadURL(c, req)
	}

	c.JSON(http.StatusOK, gin.H{
		"export_requests": responses,
		"count":           len(responses),
	})
}
