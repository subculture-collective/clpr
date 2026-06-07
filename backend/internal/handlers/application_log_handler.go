package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"git.subcult.tv/subculture-collective/clpr/internal/models"
)

// ApplicationLogRepositoryInterface defines the interface for application log repository operations
type ApplicationLogRepositoryInterface interface {
	Create(ctx context.Context, log *models.ApplicationLog) error
	GetLogStats(ctx context.Context) (map[string]interface{}, error)
}

// ApplicationLogHandler handles application log operations
type ApplicationLogHandler struct {
	logRepo ApplicationLogRepositoryInterface
}

// NewApplicationLogHandler creates a new ApplicationLogHandler
func NewApplicationLogHandler(logRepo ApplicationLogRepositoryInterface) *ApplicationLogHandler {
	return &ApplicationLogHandler{
		logRepo: logRepo,
	}
}

// CreateLog handles POST /api/v1/logs
// Accepts log entries from frontend and mobile clients
func (h *ApplicationLogHandler) CreateLog(c *gin.Context) {
	var req models.CreateApplicationLogRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request body",
		})
		return
	}

	// Validate total payload size (max 100KB)
	payload, err := json.Marshal(req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Failed to process log payload",
		})
		return
	}
	if len(payload) > 100000 {
		c.JSON(http.StatusRequestEntityTooLarge, gin.H{
			"error": "Log payload exceeds maximum size of 100KB",
		})
		return
	}

	// Set timestamp if not provided
	timestamp := time.Now()
	if req.Timestamp != nil {
		timestamp = *req.Timestamp

		// Validate that the provided timestamp is within an acceptable range
		maxFuture := time.Now().Add(5 * time.Minute)
		minPast := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
		if timestamp.After(maxFuture) || timestamp.Before(minPast) {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Timestamp is out of acceptable range",
			})
			return
		}
	}

	// Determine service from platform or use provided service
	service := "clpr-frontend"
	if req.Service != "" {
		service = req.Service
	} else {
		// Auto-detect service from platform only if not provided
		if req.Platform == "ios" || req.Platform == "android" {
			service = "clpr-mobile"
		}
	}

	// Get user ID from context if authenticated (optional)
	var userID *uuid.UUID
	if uid, exists := c.Get("user_id"); exists {
		if parsedUID, ok := uid.(uuid.UUID); ok {
			userID = &parsedUID
		}
	}

	// Get IP address from request
	ipAddress := c.ClientIP()

	// Filter sensitive data from message and error
	req.Message = h.filterSensitiveData(req.Message)
	if req.Error != nil {
		filtered := h.filterSensitiveData(*req.Error)
		req.Error = &filtered
	}

	// Filter sensitive data from context
	if req.Context != nil {
		req.Context = h.filterSensitiveContext(req.Context)
	}

	// Convert context to JSON
	var contextJSON json.RawMessage
	if req.Context != nil {
		contextBytes, err := json.Marshal(req.Context)
		if err != nil {
			// Context marshaling failed - continue without context
			// This shouldn't happen in normal operation as we validate input
			contextJSON = nil
		} else {
			contextJSON = contextBytes
		}
	}

	// Create application log entry
	var platformPtr *string
	if req.Platform != "" {
		platformPtr = &req.Platform
	}

	log := &models.ApplicationLog{
		Level:      req.Level,
		Message:    req.Message,
		Timestamp:  timestamp,
		Service:    service,
		Platform:   platformPtr,
		UserID:     userID,
		SessionID:  req.SessionID,
		TraceID:    req.TraceID,
		URL:        req.URL,
		UserAgent:  req.UserAgent,
		DeviceID:   req.DeviceID,
		AppVersion: req.AppVersion,
		Error:      req.Error,
		Stack:      req.Stack,
		Context:    contextJSON,
		IPAddress:  &ipAddress,
	}

	// Store in database
	if err := h.logRepo.Create(c.Request.Context(), log); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to store log",
		})
		return
	}

	// Return 204 No Content on success
	c.Writer.WriteHeader(http.StatusNoContent)
}

// filterSensitiveData removes sensitive information from log strings
func (h *ApplicationLogHandler) filterSensitiveData(text string) string {
	// These patterns are already filtered on the client side, but we add
	// server-side filtering as a defense-in-depth measure

	// Remove passwords and tokens (case-insensitive)
	sensitivePatterns := []string{
		"password", "passwd", "pwd", "secret", "token",
		"apikey", "api_key", "access_token", "auth_token",
		"authorization", "bearer",
	}

	lowerText := strings.ToLower(text)
	for _, pattern := range sensitivePatterns {
		if strings.Contains(lowerText, pattern) {
			// Don't log the actual sensitive data
			return "[REDACTED - contains sensitive data]"
		}
	}

	return text
}

// filterSensitiveContext removes sensitive keys from context map
func (h *ApplicationLogHandler) filterSensitiveContext(context map[string]interface{}) map[string]interface{} {
	filtered := make(map[string]interface{})

	for key, value := range context {
		lowerKey := strings.ToLower(key)

		// Skip sensitive keys
		if strings.Contains(lowerKey, "password") ||
			strings.Contains(lowerKey, "secret") ||
			strings.Contains(lowerKey, "token") ||
			strings.Contains(lowerKey, "api_key") ||
			strings.Contains(lowerKey, "apikey") ||
			strings.Contains(lowerKey, "authorization") ||
			lowerKey == "auth" {
			filtered[key] = "[REDACTED]"
			continue
		}

		// Filter nested maps recursively
		if nestedMap, ok := value.(map[string]interface{}); ok {
			filtered[key] = h.filterSensitiveContext(nestedMap)
		} else {
			filtered[key] = value
		}
	}

	return filtered
}

// GetLogStats handles GET /api/v1/logs/stats (admin only)
// Returns statistics about stored logs
func (h *ApplicationLogHandler) GetLogStats(c *gin.Context) {
	stats, err := h.logRepo.GetLogStats(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to retrieve log statistics",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"stats":   stats,
	})
}
