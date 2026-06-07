package middleware

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	goredis "github.com/redis/go-redis/v9"
	"git.subcult.tv/subculture-collective/clpr/internal/models"
	redispkg "git.subcult.tv/subculture-collective/clpr/pkg/redis"
)

var (
	// Global fallback rate limiters (one per middleware instance)
	ipFallbackLimiter   *InMemoryRateLimiter
	userFallbackLimiter *InMemoryRateLimiter

	// IP whitelist for rate limiting bypass (for testing/trusted IPs)
	// Always includes localhost (127.0.0.1, ::1)
	// Additional IPs can be configured via RATE_LIMIT_WHITELIST_IPS environment variable
	rateLimitWhitelist   map[string]bool
	rateLimitWhitelistMu sync.RWMutex
)

// InitRateLimitWhitelist initializes the rate limit whitelist from configuration
// This should be called once at application startup
func InitRateLimitWhitelist(whitelistIPs string) {
	rateLimitWhitelistMu.Lock()
	defer rateLimitWhitelistMu.Unlock()

	// Always keep localhost
	rateLimitWhitelist = map[string]bool{
		"127.0.0.1": true,
		"::1":       true,
	}

	// Add configured IPs
	if whitelistIPs != "" {
		for _, ip := range strings.Split(whitelistIPs, ",") {
			trimmed := strings.TrimSpace(ip)
			if trimmed != "" {
				rateLimitWhitelist[trimmed] = true
			}
		}
	}
}

// isIPWhitelisted checks if an IP is in the whitelist (thread-safe)
func isIPWhitelisted(ip string) bool {
	rateLimitWhitelistMu.RLock()
	defer rateLimitWhitelistMu.RUnlock()
	return rateLimitWhitelist[ip]
}

// getUserRateLimitMultiplier determines the rate limit multiplier based on user tier
// Returns: (multiplier, isAdmin)
// - Admin: unlimited (returns 0, true)
// - Premium (pro tier): 5x multiplier
// - Basic (free tier): 1x multiplier
// - Unauthenticated: returns 1x multiplier; IP-based rate limiting is applied by the parent middleware when user_id is not present
func getUserRateLimitMultiplier(c *gin.Context, subscriptionService SubscriptionChecker) (float64, bool) {
	// Check if user is authenticated
	userID, exists := c.Get("user_id")
	if !exists {
		return 1.0, false
	}

	// Check if user is admin (bypass rate limits)
	if role, exists := c.Get("user_role"); exists {
		if roleStr, ok := role.(string); ok && roleStr == models.RoleAdmin {
			return 0, true // Admin gets unlimited
		}
	}

	// Check subscription tier for premium users
	// First check if already set in context (from EnrichWithSubscriptionMiddleware)
	if tier, exists := c.Get("subscription_tier"); exists {
		if tierStr, ok := tier.(string); ok && tierStr == "pro" {
			return 5.0, false // Premium gets 5x
		}
	} else if subscriptionService != nil {
		// On-demand check if not in context (only when rate limiting is enforced)
		var uid uuid.UUID
		switch v := userID.(type) {
		case uuid.UUID:
			uid = v
		case string:
			parsed, err := uuid.Parse(v)
			if err == nil {
				uid = parsed
			}
		}

		if uid != uuid.Nil && subscriptionService.IsProUser(c.Request.Context(), uid) {
			// Cache the result in context for subsequent checks
			c.Set("subscription_tier", "pro")
			return 5.0, false // Premium gets 5x
		}
	}

	// Default: basic authenticated user (1x multiplier)
	return 1.0, false
}

// RateLimitMiddleware creates rate limiting middleware using sliding window algorithm
// For subscription-aware rate limiting, use RateLimitMiddlewareWithSubscription
func RateLimitMiddleware(redis *redispkg.Client, requests int, window time.Duration) gin.HandlerFunc {
	return RateLimitMiddlewareWithSubscription(redis, requests, window, nil)
}

// RateLimitMiddlewareWithSubscription creates rate limiting middleware with subscription-aware multipliers
func RateLimitMiddlewareWithSubscription(redis *redispkg.Client, requests int, window time.Duration, subscriptionService SubscriptionChecker) gin.HandlerFunc {
	// Initialize fallback limiter on first call
	if ipFallbackLimiter == nil {
		ipFallbackLimiter = NewInMemoryRateLimiter(requests, window)
	}
	return func(c *gin.Context) {
		// Get client IP and endpoint for granular rate limiting
		ip := c.ClientIP()

		// Skip rate limiting for whitelisted IPs (thread-safe check)
		if isIPWhitelisted(ip) {
			c.Header("X-RateLimit-Bypass", "whitelisted")
			c.Next()
			return
		}

		// Check if user is admin (bypass rate limits)
		multiplier, isAdmin := getUserRateLimitMultiplier(c, subscriptionService)
		if isAdmin {
			c.Header("X-RateLimit-Bypass", "admin")
			c.Next()
			return
		}

		// Apply multiplier to rate limit for premium users
		adjustedLimit := int(float64(requests) * multiplier)
		if adjustedLimit == 0 {
			adjustedLimit = requests
		}

		endpoint := c.Request.URL.Path
		key := fmt.Sprintf("ratelimit:%s:%s", endpoint, ip)

		ctx := c.Request.Context()

		// Use sliding window algorithm for more accurate rate limiting
		// Key format: ratelimit:{endpoint}:{ip}:{timestamp_bucket}
		now := time.Now()
		currentWindow := now.Unix() / int64(window.Seconds())
		previousWindow := currentWindow - 1

		currentKey := fmt.Sprintf("%s:%d", key, currentWindow)
		previousKey := fmt.Sprintf("%s:%d", key, previousWindow)

		// Get counts from current and previous windows
		pipe := redis.Pipeline()
		currentCmd := pipe.Get(ctx, currentKey)
		previousCmd := pipe.Get(ctx, previousKey)
		if _, err := pipe.Exec(ctx); err != nil && !errors.Is(err, goredis.Nil) {
			// If Redis pipeline fails, use in-memory fallback rate limiter
			log.Printf("Redis pipeline failed for rate limiting, using in-memory fallback: %v", err)
			allowed, remaining := ipFallbackLimiter.Allow(key)

			if !allowed {
				c.Header("X-RateLimit-Limit", fmt.Sprintf("%d", adjustedLimit))
				c.Header("X-RateLimit-Remaining", "0")
				c.Header("X-RateLimit-Fallback", "true")

				c.JSON(http.StatusTooManyRequests, gin.H{
					"error": "Rate limit exceeded. Please try again later.",
				})
				c.Abort()
				return
			}

			// Add rate limit headers for fallback
			c.Header("X-RateLimit-Limit", fmt.Sprintf("%d", adjustedLimit))
			c.Header("X-RateLimit-Remaining", fmt.Sprintf("%d", remaining))
			c.Header("X-RateLimit-Fallback", "true")
			c.Next()
			return
		}

		currentCount := int64(0)
		if val, err := currentCmd.Result(); err == nil {
			if parsed, err := strconv.ParseInt(val, 10, 64); err != nil {
				log.Printf("Warning: failed to parse currentCount from Redis value '%s': err=%v", val, err)
			} else {
				currentCount = parsed
			}
		}

		previousCount := int64(0)
		if val, err := previousCmd.Result(); err == nil {
			if parsed, err := strconv.ParseInt(val, 10, 64); err != nil {
				log.Printf("Warning: failed to parse previousCount from Redis value '%s': err=%v", val, err)
			} else {
				previousCount = parsed
			}
		}

		// Calculate weighted count for sliding window
		elapsed := float64(now.Unix() % int64(window.Seconds()))
		windowSeconds := float64(window.Seconds())
		weight := (windowSeconds - elapsed) / windowSeconds

		weightedCount := int64(float64(previousCount)*weight) + currentCount

		// Check if rate limit exceeded
		if weightedCount >= int64(adjustedLimit) {
			retryAfter := int(windowSeconds - elapsed)
			c.Header("X-RateLimit-Limit", fmt.Sprintf("%d", adjustedLimit))
			c.Header("X-RateLimit-Remaining", "0")
			c.Header("X-RateLimit-Reset", fmt.Sprintf("%d", now.Unix()+int64(retryAfter)))
			c.Header("Retry-After", fmt.Sprintf("%d", retryAfter))

			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":       "Rate limit exceeded. Please try again later.",
				"retry_after": retryAfter,
			})
			c.Abort()
			return
		}

		// Increment current window counter
		count, err := redis.Increment(ctx, currentKey)
		if err != nil {
			// If Redis fails, use in-memory fallback rate limiter
			log.Printf("Redis increment failed for rate limiting, using in-memory fallback: %v", err)
			allowed, remaining := ipFallbackLimiter.Allow(key)

			if !allowed {
				c.Header("X-RateLimit-Limit", fmt.Sprintf("%d", adjustedLimit))
				c.Header("X-RateLimit-Remaining", "0")
				c.Header("X-RateLimit-Fallback", "true")

				c.JSON(http.StatusTooManyRequests, gin.H{
					"error": "Rate limit exceeded. Please try again later.",
				})
				c.Abort()
				return
			}

			// Add rate limit headers for fallback
			c.Header("X-RateLimit-Limit", fmt.Sprintf("%d", adjustedLimit))
			c.Header("X-RateLimit-Remaining", fmt.Sprintf("%d", remaining))
			c.Header("X-RateLimit-Fallback", "true")
			c.Next()
			return
		}

		// Set expiration on first request (2x window to keep previous window)
		if count == 1 {
			_ = redis.Expire(ctx, currentKey, window*2)
		}

		// Add rate limit headers
		remaining := int64(adjustedLimit) - (weightedCount + 1)
		if remaining < 0 {
			remaining = 0
		}
		c.Header("X-RateLimit-Limit", fmt.Sprintf("%d", adjustedLimit))
		c.Header("X-RateLimit-Remaining", fmt.Sprintf("%d", remaining))
		c.Header("X-RateLimit-Reset", fmt.Sprintf("%d", (currentWindow+1)*int64(window.Seconds())))

		c.Next()
	}
}

// RateLimitByUserMiddleware creates rate limiting middleware based on authenticated user
// For subscription-aware rate limiting, use RateLimitByUserMiddlewareWithSubscription
func RateLimitByUserMiddleware(redis *redispkg.Client, requests int, window time.Duration) gin.HandlerFunc {
	return RateLimitByUserMiddlewareWithSubscription(redis, requests, window, nil)
}

// RateLimitByUserMiddlewareWithSubscription creates rate limiting middleware with subscription-aware multipliers
func RateLimitByUserMiddlewareWithSubscription(redis *redispkg.Client, requests int, window time.Duration, subscriptionService SubscriptionChecker) gin.HandlerFunc {
	// Initialize fallback limiter on first call
	if userFallbackLimiter == nil {
		userFallbackLimiter = NewInMemoryRateLimiter(requests, window)
	}
	return func(c *gin.Context) {
		// Get user ID from context (set by auth middleware)
		userID, exists := c.Get("user_id")
		if !exists {
			// Fall back to IP-based rate limiting
			RateLimitMiddlewareWithSubscription(redis, requests, window, subscriptionService)(c)
			return
		}

		// Check if user is admin (bypass rate limits)
		multiplier, isAdmin := getUserRateLimitMultiplier(c, subscriptionService)
		if isAdmin {
			c.Header("X-RateLimit-Bypass", "admin")
			c.Next()
			return
		}

		// Apply multiplier to rate limit for premium users
		adjustedLimit := int(float64(requests) * multiplier)
		if adjustedLimit == 0 {
			adjustedLimit = requests
		}

		endpoint := c.Request.URL.Path
		key := fmt.Sprintf("ratelimit:%s:user:%v", endpoint, userID)

		ctx := c.Request.Context()

		// Simple counter approach for authenticated users
		count, err := redis.Increment(ctx, key)
		if err != nil {
			// If Redis fails, use in-memory fallback rate limiter
			log.Printf("Redis increment failed for user rate limiting, using in-memory fallback: %v", err)
			allowed, remaining := userFallbackLimiter.Allow(key)

			if !allowed {
				c.Header("X-RateLimit-Limit", fmt.Sprintf("%d", adjustedLimit))
				c.Header("X-RateLimit-Remaining", "0")
				c.Header("X-RateLimit-Fallback", "true")

				c.JSON(http.StatusTooManyRequests, gin.H{
					"error": "Rate limit exceeded. Please try again later.",
				})
				c.Abort()
				return
			}

			// Add rate limit headers for fallback
			c.Header("X-RateLimit-Limit", fmt.Sprintf("%d", adjustedLimit))
			c.Header("X-RateLimit-Remaining", fmt.Sprintf("%d", remaining))
			c.Header("X-RateLimit-Fallback", "true")
			c.Next()
			return
		}

		// Set expiration on first request
		if count == 1 {
			_ = redis.Expire(ctx, key, window)
		}

		// Check if rate limit exceeded
		if count > int64(adjustedLimit) {
			// Calculate reset time
			now := time.Now()
			resetTime := now.Unix() + int64(window.Seconds())
			retryAfter := int(window.Seconds())

			c.Header("X-RateLimit-Limit", fmt.Sprintf("%d", adjustedLimit))
			c.Header("X-RateLimit-Remaining", "0")
			c.Header("X-RateLimit-Reset", fmt.Sprintf("%d", resetTime))
			c.Header("Retry-After", fmt.Sprintf("%d", retryAfter))

			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":       "Rate limit exceeded. Please try again later.",
				"retry_after": retryAfter,
			})
			c.Abort()
			return
		}

		// Add rate limit headers
		remaining := int64(adjustedLimit) - count
		now := time.Now()
		resetTime := now.Unix() + int64(window.Seconds())
		c.Header("X-RateLimit-Limit", fmt.Sprintf("%d", adjustedLimit))
		c.Header("X-RateLimit-Remaining", fmt.Sprintf("%d", remaining))
		c.Header("X-RateLimit-Reset", fmt.Sprintf("%d", resetTime))

		c.Next()
	}
}
