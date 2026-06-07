package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"git.subcult.tv/subculture-collective/clpr/internal/models"
	"git.subcult.tv/subculture-collective/clpr/internal/services"
)

// AccountTypeHandler handles account type-related HTTP requests
type AccountTypeHandler struct {
	accountTypeService *services.AccountTypeService
	authService        *services.AuthService
}

// NewAccountTypeHandler creates a new account type handler
func NewAccountTypeHandler(
	accountTypeService *services.AccountTypeService,
	authService *services.AuthService,
) *AccountTypeHandler {
	return &AccountTypeHandler{
		accountTypeService: accountTypeService,
		authService:        authService,
	}
}

// GetAccountType retrieves a user's account type information
// GET /api/v1/users/:id/account-type
func (h *AccountTypeHandler) GetAccountType(c *gin.Context) {
	userIDStr := c.Param("id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "INVALID_USER_ID",
				"message": "Invalid user ID format",
			},
		})
		return
	}

	// Get account type information
	accountTypeInfo, err := h.accountTypeService.GetUserAccountType(c.Request.Context(), userID)
	if err != nil {
		if err == services.ErrUserNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"success": false,
				"error": gin.H{
					"code":    "USER_NOT_FOUND",
					"message": "User not found",
				},
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "INTERNAL_ERROR",
				"message": "Failed to retrieve account type",
			},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    accountTypeInfo,
	})
}

// ConvertToBroadcaster converts the current user to broadcaster account type
// POST /api/v1/users/me/convert-to-broadcaster
func (h *AccountTypeHandler) ConvertToBroadcaster(c *gin.Context) {
	// Get current user from context
	userInterface, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "UNAUTHORIZED",
				"message": "Authentication required",
			},
		})
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
		return
	}

	// Parse request body
	var req models.ConvertToBroadcasterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "INVALID_REQUEST",
				"message": "Invalid request body",
			},
		})
		return
	}

	// Convert to broadcaster
	// In production, you might want to verify the user has a Twitch broadcaster profile
	twitchVerified := user.TwitchID != nil && *user.TwitchID != ""
	err := h.accountTypeService.ConvertToBroadcaster(c.Request.Context(), user.ID, req.Reason, twitchVerified)
	if err != nil {
		if err == services.ErrCannotDowngradeAccountType {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"error": gin.H{
					"code":    "INVALID_CONVERSION",
					"message": "Cannot downgrade account type",
				},
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "CONVERSION_FAILED",
				"message": "Failed to convert to broadcaster",
			},
		})
		return
	}

	// Get updated account type info
	accountTypeInfo, err := h.accountTypeService.GetUserAccountType(c.Request.Context(), user.ID)
	if err != nil {
		// Still return success since conversion worked
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"message": "Successfully converted to broadcaster",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Successfully converted to broadcaster",
		"data":    accountTypeInfo,
	})
}

// ConvertToModerator converts a user to moderator account type (admin only)
// POST /api/v1/admin/users/:id/convert-to-moderator
func (h *AccountTypeHandler) ConvertToModerator(c *gin.Context) {
	// Get target user ID from URL
	userIDStr := c.Param("id")
	targetUserID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "INVALID_USER_ID",
				"message": "Invalid user ID format",
			},
		})
		return
	}

	// Get admin user from context
	adminInterface, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "UNAUTHORIZED",
				"message": "Authentication required",
			},
		})
		return
	}

	adminUser, ok := adminInterface.(*models.User)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "INTERNAL_ERROR",
				"message": "Invalid user format",
			},
		})
		return
	}

	// Parse request body
	var req models.ConvertToModeratorRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "INVALID_REQUEST",
				"message": "Invalid request body",
			},
		})
		return
	}

	// Convert to moderator
	err = h.accountTypeService.ConvertToModerator(c.Request.Context(), targetUserID, adminUser.ID, req.Reason)
	if err != nil {
		if err == services.ErrUserNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"success": false,
				"error": gin.H{
					"code":    "USER_NOT_FOUND",
					"message": "User not found",
				},
			})
			return
		}
		if err == services.ErrCannotDowngradeAccountType {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"error": gin.H{
					"code":    "INVALID_CONVERSION",
					"message": "Cannot downgrade account type",
				},
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "CONVERSION_FAILED",
				"message": "Failed to convert to moderator",
			},
		})
		return
	}

	// Get updated account type info
	accountTypeInfo, err := h.accountTypeService.GetUserAccountType(c.Request.Context(), targetUserID)
	if err != nil {
		// Still return success since conversion worked
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"message": "Successfully converted user to moderator",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Successfully converted user to moderator",
		"data":    accountTypeInfo,
	})
}

// GetConversionHistory retrieves conversion history for a user
// GET /api/v1/users/:id/account-type/history
func (h *AccountTypeHandler) GetConversionHistory(c *gin.Context) {
	userIDStr := c.Param("id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "INVALID_USER_ID",
				"message": "Invalid user ID format",
			},
		})
		return
	}

	// Parse pagination parameters
	limit := 20
	if limitStr := c.Query("limit"); limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 && parsedLimit <= 100 {
			limit = parsedLimit
		}
	}

	offset := 0
	if offsetStr := c.Query("offset"); offsetStr != "" {
		if parsedOffset, err := strconv.Atoi(offsetStr); err == nil && parsedOffset >= 0 {
			offset = parsedOffset
		}
	}

	// Get conversion history
	conversions, total, err := h.accountTypeService.GetConversionHistory(c.Request.Context(), userID, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "INTERNAL_ERROR",
				"message": "Failed to retrieve conversion history",
			},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"conversions": conversions,
			"total":       total,
			"limit":       limit,
			"offset":      offset,
		},
	})
}

// GetAccountTypeStats retrieves account type statistics (admin only)
// GET /api/v1/admin/account-types/stats
func (h *AccountTypeHandler) GetAccountTypeStats(c *gin.Context) {
	stats, err := h.accountTypeService.GetAccountTypeStats(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "INTERNAL_ERROR",
				"message": "Failed to retrieve statistics",
			},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    stats,
	})
}

// GetRecentConversions retrieves recent conversions (admin only)
// GET /api/v1/admin/account-types/conversions
func (h *AccountTypeHandler) GetRecentConversions(c *gin.Context) {
	// Parse pagination parameters
	limit := 50
	if limitStr := c.Query("limit"); limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 && parsedLimit <= 100 {
			limit = parsedLimit
		}
	}

	offset := 0
	if offsetStr := c.Query("offset"); offsetStr != "" {
		if parsedOffset, err := strconv.Atoi(offsetStr); err == nil && parsedOffset >= 0 {
			offset = parsedOffset
		}
	}

	conversions, total, err := h.accountTypeService.GetRecentConversions(c.Request.Context(), limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "INTERNAL_ERROR",
				"message": "Failed to retrieve conversions",
			},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"conversions": conversions,
			"total":       total,
			"limit":       limit,
			"offset":      offset,
		},
	})
}
