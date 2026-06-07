package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"git.subcult.tv/subculture-collective/clpr/internal/models"
	"git.subcult.tv/subculture-collective/clpr/internal/repository"
)

// ConsentHandler handles cookie consent HTTP requests
type ConsentHandler struct {
	consentRepo *repository.ConsentRepository
}

// NewConsentHandler creates a new consent handler
func NewConsentHandler(consentRepo *repository.ConsentRepository) *ConsentHandler {
	return &ConsentHandler{
		consentRepo: consentRepo,
	}
}

// SaveConsent saves or updates user cookie consent preferences
// POST /api/v1/users/me/consent
func (h *ConsentHandler) SaveConsent(c *gin.Context) {
	// Get user ID from context (set by auth middleware)
	userIDInterface, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return
	}

	userID, ok := userIDInterface.(uuid.UUID)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "invalid user ID"})
		return
	}

	var req models.UpdateConsentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	// Essential cookies are always required
	req.Essential = true

	consent := &models.CookieConsent{
		UserID:      userID,
		Essential:   req.Essential,
		Functional:  req.Functional,
		Analytics:   req.Analytics,
		Advertising: req.Advertising,
	}

	// Get IP address and user agent for audit trail
	ipAddress := c.ClientIP()
	userAgent := c.Request.UserAgent()

	err := h.consentRepo.SaveConsent(c.Request.Context(), consent, ipAddress, userAgent)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save consent"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    consent,
	})
}

// GetConsent retrieves the current consent preferences for a user
// GET /api/v1/users/me/consent
func (h *ConsentHandler) GetConsent(c *gin.Context) {
	// Get user ID from context (set by auth middleware)
	userIDInterface, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return
	}

	userID, ok := userIDInterface.(uuid.UUID)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "invalid user ID"})
		return
	}

	consent, err := h.consentRepo.GetConsent(c.Request.Context(), userID)
	if err != nil {
		if err == repository.ErrConsentNotFound {
			// No consent saved yet - return default values
			c.JSON(http.StatusOK, gin.H{
				"success": true,
				"data": map[string]interface{}{
					"essential":   true,
					"functional":  false,
					"analytics":   false,
					"advertising": false,
					"expires_at":  nil,
				},
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to retrieve consent"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    consent,
	})
}
