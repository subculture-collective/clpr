package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"git.subcult.tv/subculture-collective/clpr/internal/models"
	"git.subcult.tv/subculture-collective/clpr/pkg/utils"
)

// RequirePermission creates middleware that requires a specific permission
// For community moderators, it also validates channel scope if a channel_id is provided in the request
func RequirePermission(permission string) gin.HandlerFunc {
	return func(c *gin.Context) {
		logger := utils.GetLogger()

		// Get user from context (set by AuthMiddleware)
		userInterface, exists := c.Get("user")
		if !exists {
			logger.Warn("Permission check failed: user not authenticated", map[string]interface{}{
				"path":       c.Request.URL.Path,
				"permission": permission,
			})
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
			logger.Error("Permission check failed: invalid user format", nil, map[string]interface{}{
				"path":       c.Request.URL.Path,
				"permission": permission,
			})
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

		// Check if user has the required permission
		if !user.Can(permission) {
			logger.Warn("Permission denied", map[string]interface{}{
				"user_id":      user.ID.String(),
				"username":     user.Username,
				"account_type": user.GetAccountType(),
				"permission":   permission,
				"path":         c.Request.URL.Path,
			})
			c.JSON(http.StatusForbidden, gin.H{
				"success": false,
				"error": gin.H{
					"code":    "FORBIDDEN",
					"message": "Insufficient permissions",
					"details": gin.H{
						"required_permission": permission,
						"account_type":        user.GetAccountType(),
					},
				},
			})
			c.Abort()
			return
		}

		// For community moderators, validate channel scope
		if user.ModeratorScope == models.ModeratorScopeCommunity && len(user.ModerationChannels) > 0 {
			// Check if a channel_id is provided in path params, query params, or JSON body
			channelID := getChannelIDFromRequest(c)
			if channelID != uuid.Nil {
				// Validate that the moderator has access to this channel
				hasAccess := false
				for _, moderatedChannel := range user.ModerationChannels {
					if moderatedChannel == channelID {
						hasAccess = true
						break
					}
				}

				if !hasAccess {
					logger.Warn("Channel scope violation", map[string]interface{}{
						"user_id":    user.ID.String(),
						"username":   user.Username,
						"permission": permission,
						"channel_id": channelID.String(),
						"path":       c.Request.URL.Path,
					})
					c.JSON(http.StatusForbidden, gin.H{
						"success": false,
						"error": gin.H{
							"code":    "FORBIDDEN",
							"message": "Access denied: channel not in moderation scope",
							"details": gin.H{
								"required_permission": permission,
								"account_type":        user.GetAccountType(),
								"channel_id":          channelID,
							},
						},
					})
					c.Abort()
					return
				}
			}
		}

		// Log successful permission check
		logger.Info("Permission granted", map[string]interface{}{
			"user_id":      user.ID.String(),
			"account_type": user.GetAccountType(),
			"permission":   permission,
			"path":         c.Request.URL.Path,
		})

		c.Next()
	}
}

// getChannelIDFromRequest extracts channel_id from request path params, query params, or JSON body
func getChannelIDFromRequest(c *gin.Context) uuid.UUID {
	// Try path parameter first
	if channelIDStr := c.Param("channel_id"); channelIDStr != "" {
		if channelID, err := uuid.Parse(channelIDStr); err == nil {
			return channelID
		}
	}

	// Try query parameter
	if channelIDStr := c.Query("channel_id"); channelIDStr != "" {
		if channelID, err := uuid.Parse(channelIDStr); err == nil {
			return channelID
		}
	}

	// For JSON body requests, we check if it's been set in context first
	// This avoids consuming the request body which would prevent the handler from reading it
	if channelIDInterface, exists := c.Get("channel_id"); exists {
		if channelID, ok := channelIDInterface.(uuid.UUID); ok {
			return channelID
		}
	}

	// Note: We don't read from JSON body directly to avoid consuming the request body.
	// If channel scope validation is needed for JSON payloads, the application should
	// set the channel_id in the context before calling this middleware, or pass it
	// as a path/query parameter.

	return uuid.Nil
}

// RequireAnyPermission creates middleware that requires any of the specified permissions
func RequireAnyPermission(permissions ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		logger := utils.GetLogger()

		// Get user from context (set by AuthMiddleware)
		userInterface, exists := c.Get("user")
		if !exists {
			logger.Warn("Permission check failed: user not authenticated", map[string]interface{}{
				"path":        c.Request.URL.Path,
				"permissions": permissions,
			})
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
			logger.Error("Permission check failed: invalid user format", nil, map[string]interface{}{
				"path":        c.Request.URL.Path,
				"permissions": permissions,
			})
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

		// Check if user has any of the required permissions
		for _, permission := range permissions {
			if user.Can(permission) {
				// For community moderators, validate channel scope
				if user.ModeratorScope == models.ModeratorScopeCommunity && len(user.ModerationChannels) > 0 {
					channelID := getChannelIDFromRequest(c)
					if channelID != uuid.Nil {
						// Validate that the moderator has access to this channel
						hasAccess := false
						for _, moderatedChannel := range user.ModerationChannels {
							if moderatedChannel == channelID {
								hasAccess = true
								break
							}
						}

						if !hasAccess {
							logger.Warn("Channel scope violation", map[string]interface{}{
								"user_id":     user.ID.String(),
								"username":    user.Username,
								"permissions": permissions,
								"channel_id":  channelID.String(),
								"path":        c.Request.URL.Path,
							})
							c.JSON(http.StatusForbidden, gin.H{
								"success": false,
								"error": gin.H{
									"code":    "FORBIDDEN",
									"message": "Access denied: channel not in moderation scope",
									"details": gin.H{
										"required_permissions": permissions,
										"account_type":         user.GetAccountType(),
										"channel_id":           channelID,
									},
								},
							})
							c.Abort()
							return
						}
					}
				}

				logger.Info("Permission granted", map[string]interface{}{
					"user_id":      user.ID.String(),
					"account_type": user.GetAccountType(),
					"permission":   permission,
					"permissions":  permissions,
					"path":         c.Request.URL.Path,
				})
				c.Next()
				return
			}
		}

		logger.Warn("Permission denied: lacks all required permissions", map[string]interface{}{
			"user_id":      user.ID.String(),
			"username":     user.Username,
			"account_type": user.GetAccountType(),
			"permissions":  permissions,
			"path":         c.Request.URL.Path,
		})
		c.JSON(http.StatusForbidden, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "FORBIDDEN",
				"message": "Insufficient permissions",
				"details": gin.H{
					"required_permissions": permissions,
					"account_type":         user.GetAccountType(),
				},
			},
		})
		c.Abort()
	}
}

// RequireAccountType creates middleware that requires a specific account type
func RequireAccountType(accountTypes ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get user from context (set by AuthMiddleware)
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

		userAccountType := user.GetAccountType()

		// Check if user has the required account type
		for _, accountType := range accountTypes {
			if userAccountType == accountType {
				c.Next()
				return
			}
		}

		c.JSON(http.StatusForbidden, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "FORBIDDEN",
				"message": "Insufficient account type",
				"details": gin.H{
					"required_account_types": accountTypes,
					"current_account_type":   userAccountType,
				},
			},
		})
		c.Abort()
	}
}
