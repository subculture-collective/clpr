package middleware

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	redispkg "git.subcult.tv/subculture-collective/clpr/pkg/redis"
)

const (
	// Abuse detection thresholds
	abuseDetectionWindow = 1 * time.Hour
	abuseThreshold       = 1000 // requests per hour (increased for legitimate use with multiple API calls per page)
	abuseBanDuration     = 24 * time.Hour

	// Progressive rate limit penalties
	warningThreshold  = 0.8  // 80% of rate limit
	criticalThreshold = 0.95 // 95% of rate limit
)

// Endpoints excluded from abuse detection (auth endpoints already have their own rate limiting)
var abuseDetectionExemptPaths = map[string]bool{
	"/health":                      true,
	"/health/ready":                true,
	"/health/live":                 true,
	"/api/v1/auth/twitch":          true,
	"/api/v1/auth/twitch/callback": true,
	"/api/v1/auth/refresh":         true,
	"/api/v1/auth/me":              true,
}

// Localhost IPs are always allowed (developer workflows)
var abuseDetectionLocalIPs = map[string]bool{
	"127.0.0.1": true,
	"::1":       true,
}

// AbuseDetectionMiddleware monitors and blocks abusive IPs
func AbuseDetectionMiddleware(redis *redispkg.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Skip exempt endpoints (health checks and auth endpoints)
		if abuseDetectionExemptPaths[c.Request.URL.Path] {
			c.Next()
			return
		}

		// Skip if Redis client is nil (for testing)
		if redis == nil {
			c.Next()
			return
		}

		ip := c.ClientIP()
		if abuseDetectionLocalIPs[ip] {
			c.Next()
			return
		}
		ctx := c.Request.Context()

		// Check if IP is banned
		banKey := fmt.Sprintf("abuse:ban:%s", ip)
		banned, err := redis.Exists(ctx, banKey)
		if err != nil {
			// Log error but don't block request
			log.Printf("Error checking ban status: %v", err)
		} else if banned {
			c.JSON(http.StatusForbidden, gin.H{
				"error": "Access denied due to abusive behavior",
			})
			c.Abort()
			return
		}

		// Track request for abuse detection
		abuseKey := fmt.Sprintf("abuse:track:%s", ip)
		count, err := redis.Increment(ctx, abuseKey)
		if err != nil {
			log.Printf("Error tracking abuse: %v", err)
		} else {
			// Set expiration on first request
			if count == 1 {
				_ = redis.Expire(ctx, abuseKey, abuseDetectionWindow)
			}

			// Check if threshold exceeded
			if count > int64(abuseThreshold) {
				// Ban the IP
				if err := redis.Set(ctx, banKey, "1", abuseBanDuration); err != nil {
					log.Printf("Error setting ban: %v", err)
				} else {
					log.Printf("IP %s banned for abuse (exceeded %d requests in %v)",
						ip, abuseThreshold, abuseDetectionWindow)

					c.JSON(http.StatusForbidden, gin.H{
						"error": "Access denied due to abusive behavior",
					})
					c.Abort()
					return
				}
			}
		}

		c.Next()
	}
}

// EnhancedRateLimitMiddleware extends standard rate limiting with warnings
func EnhancedRateLimitMiddleware(redis *redispkg.Client, requests int, window time.Duration) gin.HandlerFunc {
	baseLimiter := RateLimitMiddleware(redis, requests, window)

	return func(c *gin.Context) {
		// Skip if Redis client is nil (for testing)
		if redis == nil {
			c.Next()
			return
		}

		// Get current rate limit stats before applying limiter
		ip := c.ClientIP()
		endpoint := c.Request.URL.Path
		key := fmt.Sprintf("ratelimit:%s:%s", endpoint, ip)

		ctx := c.Request.Context()
		now := time.Now()
		currentWindow := now.Unix() / int64(window.Seconds())
		currentKey := fmt.Sprintf("%s:%d", key, currentWindow)

		// Get current count
		currentCount := int64(0)
		if val, err := redis.Get(ctx, currentKey); err == nil {
			if parsed, err := strconv.ParseInt(val, 10, 64); err != nil {
				// Log warning and continue with 0 (safer to not apply warnings than to block legitimate traffic)
				log.Printf("Warning: failed to parse rate limit count for enhanced warnings, defaulting to 0: %v", err)
				currentCount = 0
			} else {
				currentCount = parsed
			}
		}

		// Calculate utilization percentage
		utilization := float64(currentCount) / float64(requests)

		// Add warning headers if approaching limit
		if utilization >= warningThreshold && utilization < criticalThreshold {
			c.Header("X-RateLimit-Warning", "approaching-limit")
		} else if utilization >= criticalThreshold {
			c.Header("X-RateLimit-Warning", "critical")
		}

		// Apply base rate limiter
		baseLimiter(c)
	}
}

// UnbanIP removes a ban for a specific IP (admin function)
func UnbanIP(ctx context.Context, redis *redispkg.Client, ip string) error {
	banKey := fmt.Sprintf("abuse:ban:%s", ip)
	return redis.Delete(ctx, banKey)
}

// GetBannedIPs returns a list of currently banned IPs (admin function)
func GetBannedIPs(ctx context.Context, redis *redispkg.Client) ([]string, error) {
	const prefix = "abuse:ban:"
	keys, err := redis.Keys(ctx, prefix+"*")
	if err != nil {
		return nil, err
	}

	// Extract IPs from keys
	ips := make([]string, 0, len(keys))
	for _, key := range keys {
		// Remove "abuse:ban:" prefix
		if len(key) > len(prefix) {
			ips = append(ips, key[len(prefix):])
		}
	}

	return ips, nil
}

// GetAbuseStats returns abuse statistics for an IP (admin function)
func GetAbuseStats(ctx context.Context, redis *redispkg.Client, ip string) (int64, error) {
	abuseKey := fmt.Sprintf("abuse:track:%s", ip)
	val, err := redis.Get(ctx, abuseKey)
	if err != nil {
		return 0, err
	}

	count, err := strconv.ParseInt(val, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("failed to parse abuse count: %w", err)
	}
	return count, nil
}
