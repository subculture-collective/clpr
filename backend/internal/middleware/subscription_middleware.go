package middleware

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"git.subcult.tv/subculture-collective/clpr/internal/models"
)

// sharedSubscriptionGuard centralizes extracting the current user and performing
// a predicate check on their subscription state. If predicate returns false it
// logs and writes the provided error/message pair and aborts the context.
func sharedSubscriptionGuard(
	c *gin.Context,
	subscriptionService SubscriptionChecker,
	auditLogService AuditLogger,
	predicate func(ctx *gin.Context, userID uuid.UUID) bool,
	denialCode string,
	required string,
	errorTitle string,
	message string,
) bool {
	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		c.Abort()
		return false
	}

	currentUser, ok := user.(*models.User)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user information"})
		c.Abort()
		return false
	}

	if !predicate(c, currentUser.ID) {
		if auditLogService != nil {
			if err := auditLogService.LogEntitlementDenial(c.Request.Context(), currentUser.ID, denialCode, map[string]interface{}{
				"endpoint": c.Request.URL.Path,
				"method":   c.Request.Method,
				"required": required,
			}); err != nil {
				log.Printf("[WARN] Failed to log entitlement denial: %v", err)
			}
		}

		c.JSON(http.StatusForbidden, gin.H{
			"error":   errorTitle,
			"message": message,
		})
		c.Abort()
		return false
	}

	return true
}

// RequireProSubscription middleware ensures the user has an active Pro subscription
func RequireProSubscription(subscriptionService SubscriptionChecker, auditLogService AuditLogger) gin.HandlerFunc {
	return func(c *gin.Context) {
		ok := sharedSubscriptionGuard(
			c,
			subscriptionService,
			auditLogService,
			func(ctx *gin.Context, userID uuid.UUID) bool {
				return subscriptionService.IsProUser(ctx.Request.Context(), userID)
			},
			"pro_subscription_required",
			"pro",
			"Pro subscription required",
			"This feature requires an active Pro subscription",
		)
		if !ok {
			return
		}
		c.Next()
	}
}

// RequireActiveSubscription middleware ensures the user has any active subscription
func RequireActiveSubscription(subscriptionService SubscriptionChecker, auditLogService AuditLogger) gin.HandlerFunc {
	return func(c *gin.Context) {
		ok := sharedSubscriptionGuard(
			c,
			subscriptionService,
			auditLogService,
			func(ctx *gin.Context, userID uuid.UUID) bool {
				return subscriptionService.HasActiveSubscription(ctx.Request.Context(), userID)
			},
			"active_subscription_required",
			"active",
			"Active subscription required",
			"This feature requires an active subscription",
		)
		if !ok {
			return
		}
		c.Next()
	}
}

// EnrichWithSubscriptionMiddleware adds subscription tier information to the context
// This should be used after AuthMiddleware to enrich authenticated requests with subscription data
func EnrichWithSubscriptionMiddleware(subscriptionService SubscriptionChecker) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Only enrich if user is authenticated
		userID, exists := c.Get("user_id")
		if !exists {
			c.Next()
			return
		}

		// Get user ID as UUID
		var uid uuid.UUID
		switch v := userID.(type) {
		case uuid.UUID:
			uid = v
		case string:
			parsed, err := uuid.Parse(v)
			if err != nil {
				c.Next()
				return
			}
			uid = parsed
		default:
			c.Next()
			return
		}

		// Check if user has pro subscription
		if subscriptionService.IsProUser(c.Request.Context(), uid) {
			c.Set("subscription_tier", "pro")
		} else {
			c.Set("subscription_tier", "free")
		}

		c.Next()
	}
}
