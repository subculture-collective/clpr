package handlers

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"git.subcult.tv/subculture-collective/clpr/config"
	"git.subcult.tv/subculture-collective/clpr/internal/models"
	"git.subcult.tv/subculture-collective/clpr/internal/services"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// mfaLoginVerifier is the part of MFAService used to finish a sign-in.
type mfaLoginVerifier interface {
	VerifyTOTP(ctx context.Context, userID uuid.UUID, code string, ipAddress, userAgent *string) error
	VerifyBackupCode(ctx context.Context, userID uuid.UUID, code string) error
	CreateTrustedDevice(ctx context.Context, userID uuid.UUID, fingerprint, deviceName string, ipAddress, userAgent *string) error
}

// mfaLoginUsers loads the user a pending MFA sign-in belongs to.
type mfaLoginUsers interface {
	GetUserByID(ctx context.Context, id uuid.UUID) (*models.User, error)
}

// MFAHandler handles MFA-related endpoints
type MFAHandler struct {
	mfaService *services.MFAService
	cfg        *config.Config

	loginVerifier mfaLoginVerifier
	users         mfaLoginUsers
	challenges    mfaChallengeStore
	session       *loginSession
}

// NewMFAHandler creates a new MFA handler. authService and challenges finish
// sign-ins that AuthHandler paused for an MFA code.
func NewMFAHandler(mfaService *services.MFAService, cfg *config.Config, authService *services.AuthService, challenges mfaChallengeStore) *MFAHandler {
	h := &MFAHandler{
		mfaService: mfaService,
		cfg:        cfg,
		challenges: challenges,
	}
	if mfaService != nil {
		h.loginVerifier = mfaService
	}
	if authService != nil {
		h.users = authService
		// The code was just verified, so this session issuer has no MFA gate.
		h.session = newLoginSession(cfg, authService, nil, challenges)
	}
	return h
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

	// The authenticator app shows this account name next to the issuer.
	accountName := ""
	if userValue, ok := c.Get("user"); ok {
		if user, ok := userValue.(*models.User); ok && user != nil {
			accountName = user.Username
		}
	}

	// Start enrollment
	response, err := h.mfaService.StartEnrollment(c.Request.Context(), userIDUUID, accountName)
	if err != nil {
		if err == services.ErrMFAAlreadyEnabled {
			c.JSON(http.StatusConflict, gin.H{
				"error": "MFA is already enabled for this account",
			})
			return
		}

		c.Error(err)
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
// Completes a sign-in that AuthHandler paused for an MFA code. The pending
// sign-in is identified by the mfa_challenge cookie; on success the challenge
// is consumed and session cookies are issued.
func (h *MFAHandler) VerifyLogin(c *gin.Context) {
	if h.challenges == nil || h.session == nil || h.loginVerifier == nil || h.users == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "MFA sign-in is unavailable"})
		return
	}
	ctx := c.Request.Context()

	token, cookieErr := c.Cookie(mfaChallengeCookie)
	if cookieErr != nil || token == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "No pending MFA verification",
			"code":  "MFA_CHALLENGE_MISSING",
		})
		return
	}

	userID, err := h.challenges.Resolve(ctx, token)
	if err != nil {
		if errors.Is(err, services.ErrMFAChallengeNotFound) {
			clearChallengeCookie(c)
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Your sign-in expired. Please sign in again.",
				"code":  "MFA_CHALLENGE_EXPIRED",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to verify MFA code"})
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

	// Backup codes are 8 characters; authenticator codes are 6 digits.
	code := normalizeMFALoginCode(req.Code)
	if len(code) == 8 {
		err = h.loginVerifier.VerifyBackupCode(ctx, userID, code)
	} else {
		err = h.loginVerifier.VerifyTOTP(ctx, userID, code, &ipAddress, &userAgent)
	}

	if err != nil {
		switch {
		case errors.Is(err, services.ErrInvalidMFACode), errors.Is(err, services.ErrNoBackupCodesRemaining):
			exhausted, recordErr := h.challenges.RecordFailure(ctx, token)
			if recordErr != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to verify MFA code"})
				return
			}
			if exhausted {
				clearChallengeCookie(c)
				c.JSON(http.StatusTooManyRequests, gin.H{
					"error": "Too many incorrect codes. Please sign in again.",
					"code":  "MFA_CHALLENGE_EXHAUSTED",
				})
				return
			}
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid MFA code",
				"code":  "MFA_CODE_INVALID",
			})
		case errors.Is(err, services.ErrTooManyFailedAttempts):
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": "Too many failed attempts. Account temporarily locked.",
				"code":  "MFA_LOCKED",
			})
		case errors.Is(err, services.ErrMFANotEnabled):
			// MFA was turned off after this sign-in started; starting over
			// signs the user in without a code.
			_, _ = h.challenges.Consume(ctx, token)
			clearChallengeCookie(c)
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Your sign-in expired. Please sign in again.",
				"code":  "MFA_CHALLENGE_EXPIRED",
			})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to verify MFA code",
			})
		}
		return
	}

	// Consume atomically: a challenge completes at most one sign-in.
	consumedID, err := h.challenges.Consume(ctx, token)
	if err != nil || consumedID != userID {
		clearChallengeCookie(c)
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Your sign-in expired. Please sign in again.",
			"code":  "MFA_CHALLENGE_EXPIRED",
		})
		return
	}
	clearChallengeCookie(c)

	user, err := h.users.GetUserByID(ctx, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to complete sign-in"})
		return
	}
	if _, _, err := h.session.issueTokens(c, user); err != nil {
		if errors.Is(err, services.ErrUserBanned) {
			c.JSON(http.StatusForbidden, gin.H{"error": "Your account has been banned"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to complete sign-in"})
		return
	}

	// If trust device was requested, create trusted device entry
	if req.TrustDevice != nil && *req.TrustDevice {
		fingerprint := requestDeviceFingerprint(c)
		_ = h.loginVerifier.CreateTrustedDevice(ctx, userID, fingerprint, userAgent, &ipAddress, &userAgent)
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "MFA verification successful",
		"user":    user,
	})
}

// normalizeMFALoginCode removes the spaces and dashes people type when
// copying a code and upper-cases backup codes.
func normalizeMFALoginCode(code string) string {
	code = strings.NewReplacer(" ", "", "-", "", "\t", "").Replace(strings.TrimSpace(code))
	return strings.ToUpper(code)
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
