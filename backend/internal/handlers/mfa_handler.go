package handlers

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"git.subcult.tv/subculture-collective/clpr/config"
	"git.subcult.tv/subculture-collective/clpr/internal/models"
	"git.subcult.tv/subculture-collective/clpr/internal/services"
)

// MFAHandler handles MFA-related endpoints
type MFAHandler struct {
	mfaService *services.MFAService
	cfg        *config.Config
}

// NewMFAHandler creates a new MFA handler
func NewMFAHandler(mfaService *services.MFAService, cfg *config.Config) *MFAHandler {
	return &MFAHandler{
		mfaService: mfaService,
		cfg:        cfg,
	}
}

// StartEnrollment handles POST /api/v1/auth/mfa/enroll
// Initiates MFA enrollment and returns QR code + backup codes
func (h *MFAHandler) StartEnrollment(c *gin.Context) {
	// Get authenticated user from context
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Unauthorized",
		})
		return
	}

	userIDUUID, ok := userID.(uuid.UUID)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Invalid user ID format",
		})
		return
	}

	// Get user email from context (set by auth middleware)
	email, exists := c.Get("user_email")
	if !exists {
		email = ""
	}

	emailStr, _ := email.(string)

	// Start enrollment
	response, err := h.mfaService.StartEnrollment(c.Request.Context(), userIDUUID, emailStr)
	if err != nil {
		if err == services.ErrMFAAlreadyEnabled {
			c.JSON(http.StatusConflict, gin.H{
				"error": "MFA is already enabled for this account",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to start MFA enrollment",
		})
		return
	}

	c.JSON(http.StatusOK, response)
}

// VerifyEnrollment handles POST /api/v1/auth/mfa/verify-enrollment
// Verifies TOTP code and enables MFA
func (h *MFAHandler) VerifyEnrollment(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Unauthorized",
		})
		return
	}

	userIDUUID, ok := userID.(uuid.UUID)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Invalid user ID format",
		})
		return
	}

	var req models.VerifyMFAEnrollmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request body",
		})
		return
	}

	// Verify code and enable MFA
	err := h.mfaService.VerifyEnrollment(c.Request.Context(), userIDUUID, req.Code)
	if err != nil {
		if err == services.ErrInvalidMFACode {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid MFA code",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to verify MFA enrollment",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "MFA enabled successfully",
	})
}

// VerifyLogin handles POST /api/v1/auth/mfa/verify-login
// Verifies MFA code during login
func (h *MFAHandler) VerifyLogin(c *gin.Context) {
	// This endpoint is called during login flow, user_id comes from session
	userID, exists := c.Get("pending_user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "No pending MFA verification",
		})
		return
	}

	userIDUUID, ok := userID.(uuid.UUID)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Invalid user ID format",
		})
		return
	}

	var req models.VerifyMFALoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request body",
		})
		return
	}

	// Get IP address and user agent for audit logging
	ipAddress := c.ClientIP()
	userAgent := c.GetHeader("User-Agent")

	// Check if this is a backup code (8 characters) or TOTP (6 digits)
	var err error
	if len(req.Code) == 8 {
		err = h.mfaService.VerifyBackupCode(c.Request.Context(), userIDUUID, req.Code)
	} else {
		err = h.mfaService.VerifyTOTP(c.Request.Context(), userIDUUID, req.Code, &ipAddress, &userAgent)
	}

	if err != nil {
		if err == services.ErrInvalidMFACode {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid MFA code",
			})
			return
		}

		if err == services.ErrTooManyFailedAttempts {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": "Too many failed attempts. Account temporarily locked.",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to verify MFA code",
		})
		return
	}

	// If trust device was requested, create trusted device entry
	if req.TrustDevice != nil && *req.TrustDevice {
		deviceName := c.GetHeader("User-Agent")
		fingerprint := services.GenerateDeviceFingerprint(userAgent, ipAddress)
		_ = h.mfaService.CreateTrustedDevice(c.Request.Context(), userIDUUID, fingerprint, deviceName, &ipAddress, &userAgent)
	}

	// MFA verification successful - auth middleware will handle session creation
	c.JSON(http.StatusOK, gin.H{
		"message": "MFA verification successful",
		"user_id": userIDUUID,
	})
}

// RegenerateBackupCodes handles POST /api/v1/auth/mfa/regenerate-backup-codes
// Generates new backup codes after verifying TOTP
func (h *MFAHandler) RegenerateBackupCodes(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Unauthorized",
		})
		return
	}

	userIDUUID, ok := userID.(uuid.UUID)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Invalid user ID format",
		})
		return
	}

	var req models.RegenerateBackupCodesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request body",
		})
		return
	}

	// Regenerate backup codes
	codes, err := h.mfaService.RegenerateBackupCodes(c.Request.Context(), userIDUUID, req.Code)
	if err != nil {
		if err == services.ErrInvalidMFACode {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid MFA code",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to regenerate backup codes",
		})
		return
	}

	c.JSON(http.StatusOK, models.RegenerateBackupCodesResponse{
		BackupCodes: codes,
	})
}

// GetTrustedDevices handles GET /api/v1/auth/mfa/trusted-devices
// Lists all trusted devices for the user
func (h *MFAHandler) GetTrustedDevices(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Unauthorized",
		})
		return
	}

	userIDUUID, ok := userID.(uuid.UUID)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Invalid user ID format",
		})
		return
	}

	devices, err := h.mfaService.GetTrustedDevices(c.Request.Context(), userIDUUID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get trusted devices",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"devices": devices,
	})
}

// RevokeTrustedDevice handles DELETE /api/v1/auth/mfa/trusted-devices/:id
// Revokes a trusted device
func (h *MFAHandler) RevokeTrustedDevice(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Unauthorized",
		})
		return
	}

	userIDUUID, ok := userID.(uuid.UUID)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Invalid user ID format",
		})
		return
	}

	// Get device ID from URL
	deviceID := c.Param("id")
	var deviceIDInt int
	if _, err := fmt.Sscanf(deviceID, "%d", &deviceIDInt); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid device ID",
		})
		return
	}

	err := h.mfaService.RevokeTrustedDevice(c.Request.Context(), userIDUUID, deviceIDInt)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to revoke trusted device",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Trusted device revoked successfully",
	})
}

// DisableMFA handles POST /api/v1/auth/mfa/disable
// Disables MFA after password and code verification
func (h *MFAHandler) DisableMFA(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Unauthorized",
		})
		return
	}

	userIDUUID, ok := userID.(uuid.UUID)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Invalid user ID format",
		})
		return
	}

	var req models.DisableMFARequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request body",
		})
		return
	}

	err := h.mfaService.DisableMFA(c.Request.Context(), userIDUUID, req.Code)
	if err != nil {
		if err == services.ErrInvalidMFACode {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid MFA code",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to disable MFA",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "MFA disabled successfully",
	})
}

// GetStatus handles GET /api/v1/auth/mfa/status
// Returns the current MFA status for the user
func (h *MFAHandler) GetStatus(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Unauthorized",
		})
		return
	}

	userIDUUID, ok := userID.(uuid.UUID)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Invalid user ID format",
		})
		return
	}

	status, err := h.mfaService.GetMFAStatus(c.Request.Context(), userIDUUID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get MFA status",
		})
		return
	}

	c.JSON(http.StatusOK, status)
}
