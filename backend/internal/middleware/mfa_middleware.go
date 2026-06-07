package middleware

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"git.subcult.tv/subculture-collective/clpr/internal/models"
)

// MFAServiceInterface defines the interface for MFA service operations used by middleware
type MFAServiceInterface interface {
	IsAdminActionAllowed(ctx context.Context, userID uuid.UUID) (bool, string, error)
	CheckMFARequired(ctx context.Context, userID uuid.UUID) (required bool, enabled bool, inGracePeriod bool, err error)
	SetMFARequired(ctx context.Context, userID uuid.UUID) error
}

// RequireMFAForAdminMiddleware creates middleware that enforces MFA for admin/moderator actions
func RequireMFAForAdminMiddleware(mfaService MFAServiceInterface) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get user from context (set by auth middleware)
		userInterface, exists := c.Get("user")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"error": gin.H{
					"code":    "UNAUTHORIZED",
					"message": "Authentication required",
				},
			})
			c.Abort()
			return
		}

		user, ok := userInterface.(*models.User)
		if !ok {
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"error": gin.H{
					"code":    "INTERNAL_ERROR",
					"message": "Invalid user format",
				},
			})
			c.Abort()
			return
		}

		// Only enforce MFA for admin and moderator roles
		if !user.IsModeratorOrAdmin() {
			c.Next()
			return
		}

		// Check if admin action is allowed based on MFA status
		allowed, message, err := mfaService.IsAdminActionAllowed(c.Request.Context(), user.ID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"error": gin.H{
					"code":    "INTERNAL_ERROR",
					"message": "Failed to check MFA status",
				},
			})
			c.Abort()
			return
		}

		if !allowed {
			c.JSON(http.StatusForbidden, gin.H{
				"success": false,
				"error": gin.H{
					"code":    "MFA_REQUIRED",
					"message": message,
				},
			})
			c.Abort()
			return
		}

		// If there's a warning message (grace period), add it to response headers
		if message != "" {
			c.Header("X-MFA-Warning", message)
		}

		c.Next()
	}
}

// CheckMFARequirementMiddleware checks if MFA should be required for a user after role change
// This is used after role/account_type updates to trigger MFA requirement
func CheckMFARequirementMiddleware(mfaService MFAServiceInterface) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next() // Execute the handler first

		// Only proceed if the request was successful
		if c.Writer.Status() >= 400 {
			return
		}

		// Get user from context
		userInterface, exists := c.Get("updated_user")
		if !exists {
			// If no updated_user in context, check regular user
			userInterface, exists = c.Get("user")
			if !exists {
				return
			}
		}

		user, ok := userInterface.(*models.User)
		if !ok {
			return
		}

		// Check if user is now admin or moderator
		if user.IsModeratorOrAdmin() {
			// Check if MFA is already set as required
			required, enabled, _, err := mfaService.CheckMFARequired(c.Request.Context(), user.ID)
			if err != nil {
				// Fail closed - abort the request if we cannot verify MFA requirement
				c.Error(fmt.Errorf("SECURITY WARNING: Failed to check MFA requirement for user %s: %w", user.ID, err))
				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
					"error": "Failed to verify MFA requirement. Please try again later.",
				})
				return
			}

			// If MFA is not yet required and not enabled, set it as required
			if !required && !enabled {
				if err := mfaService.SetMFARequired(c.Request.Context(), user.ID); err != nil {
					// Log error - this is a security critical operation
					c.Error(fmt.Errorf("SECURITY WARNING: Failed to set MFA requirement for user %s: %w", user.ID, err))
				}
			}
		}
	}
}
