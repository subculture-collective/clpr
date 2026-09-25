package middleware

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"git.subcult.tv/subculture-collective/clpr/config"
	redispkg "git.subcult.tv/subculture-collective/clpr/pkg/redis"
	"github.com/gin-gonic/gin"
	goredis "github.com/redis/go-redis/v9"
)

const (
	// Progressive rate limit penalties
	warningThreshold  = 0.8  // 80% of rate limit
	criticalThreshold = 0.95 // 95% of rate limit

	abuseBanPrefix     = "abuse:ban:"
	abuseOffensePrefix = "abuse:offenses:"
	abuseCountPrefix   = "abuse:count:"
)

// abuseTrafficClass separates cheap reads from state-changing requests so a
// browsing session (dozens of parallel GETs per page) is not judged against
// the budget meant for writes and sign-in attempts.
type abuseTrafficClass string

const (
	abuseClassRead  abuseTrafficClass = "read"
	abuseClassWrite abuseTrafficClass = "write"
)

// Endpoints excluded from abuse detection: health probes only.
var abuseDetectionExemptPaths = map[string]bool{
	"/health":       true,
	"/health/ready": true,
	"/health/live":  true,
}

// Mutations that are per-view telemetry or read-only lookups. They fire as
// often as page views, so they share the read budget.
var abuseReadLikeMutations = map[string]bool{
	"POST /api/v1/clips/batch-media":         true,
	"POST /api/v1/clips/:id/track-view":      true,
	"POST /api/v1/events":                    true,
	"POST /api/v1/ads/track/:id":             true,
	"POST /api/v1/logs":                      true,
	"POST /api/v1/playlists/:id/track-share": true,
}

// Localhost IPs are always allowed (developer workflows)
var abuseDetectionLocalIPs = map[string]bool{
	"127.0.0.1": true,
	"::1":       true,
}

// DefaultAbuseDetectionConfig mirrors the configuration defaults.
func DefaultAbuseDetectionConfig() config.AbuseDetectionConfig {
	return config.AbuseDetectionConfig{
		Enabled:        true,
		ReadPerMinute:  1200,
		ReadPerHour:    12000,
		WritePerMinute: 120,
		WritePerHour:   1500,
		BanDurations:   append([]time.Duration(nil), config.DefaultAbuseBanDurations...),
		OffenseWindow:  24 * time.Hour,
	}
}

// classifyAbuseTraffic returns the budget a request counts against. Auth
// endpoints count as writes even for GET because they start sign-in flows.
func classifyAbuseTraffic(c *gin.Context) abuseTrafficClass {
	path := c.Request.URL.Path
	if strings.HasPrefix(path, "/api/v1/auth/") && path != "/api/v1/auth/me" {
		return abuseClassWrite
	}
	switch c.Request.Method {
	case http.MethodGet, http.MethodHead, http.MethodOptions:
		return abuseClassRead
	}
	if abuseReadLikeMutations[c.Request.Method+" "+c.FullPath()] {
		return abuseClassRead
	}
	return abuseClassWrite
}

type abuseWindow struct {
	name   string
	length time.Duration
	limit  int
}

func abuseWindowsFor(cfg config.AbuseDetectionConfig, class abuseTrafficClass) []abuseWindow {
	if class == abuseClassRead {
		return []abuseWindow{{"m", time.Minute, cfg.ReadPerMinute}, {"h", time.Hour, cfg.ReadPerHour}}
	}
	return []abuseWindow{{"m", time.Minute, cfg.WritePerMinute}, {"h", time.Hour, cfg.WritePerHour}}
}

// abuseBanDuration picks the ban for the nth offense (1-based). The last
// configured duration repeats for later offenses.
func abuseBanDuration(durations []time.Duration, offense int64) time.Duration {
	if len(durations) == 0 {
		durations = config.DefaultAbuseBanDurations
	}
	idx := int(offense) - 1
	if idx < 0 {
		idx = 0
	}
	if idx >= len(durations) {
		idx = len(durations) - 1
	}
	return durations[idx]
}

func abortAbuseBanned(c *gin.Context, retryAfter time.Duration) {
	seconds := int(retryAfter.Round(time.Second) / time.Second)
	if seconds < 1 {
		seconds = 1
	}
	c.Header("Retry-After", strconv.Itoa(seconds))
	c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
		"error":       "Too many requests from this network. Please try again later.",
		"retry_after": seconds,
	})
}

func abuseCountKey(class abuseTrafficClass, window, ip string, bucket int64) string {
	return fmt.Sprintf("%s%s:%s:%s:%d", abuseCountPrefix, class, window, ip, bucket)
}

// AbuseDetectionMiddleware applies IP-wide volume limits in fixed minute and
// hour windows, with separate budgets for reads and writes. Exceeding a limit
// blocks the IP for an escalating period (15m, 1h, 6h, 24h by default); the
// offense count resets after cfg.OffenseWindow without another offense.
// Redis failures fail open.
func AbuseDetectionMiddleware(redis *redispkg.Client, cfg config.AbuseDetectionConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !cfg.Enabled || redis == nil || abuseDetectionExemptPaths[c.Request.URL.Path] {
			c.Next()
			return
		}

		ip := c.ClientIP()
		if abuseDetectionLocalIPs[ip] || isIPWhitelisted(ip) {
			c.Next()
			return
		}
		ctx := c.Request.Context()

		banKey := abuseBanPrefix + ip
		if _, ttl, err := redis.GetWithTTL(ctx, banKey); err == nil {
			abortAbuseBanned(c, ttl)
			return
		} else if !errors.Is(err, goredis.Nil) {
			log.Printf("Error checking ban status: %v", err)
		}

		class := classifyAbuseTraffic(c)
		now := time.Now()
		windows := abuseWindowsFor(cfg, class)
		pipe := redis.Pipeline()
		counts := make([]*goredis.IntCmd, len(windows))
		for i, w := range windows {
			key := abuseCountKey(class, w.name, ip, now.Unix()/int64(w.length/time.Second))
			counts[i] = pipe.Incr(ctx, key)
			pipe.Expire(ctx, key, w.length+time.Minute)
		}
		if _, err := pipe.Exec(ctx); err != nil {
			log.Printf("Error tracking abuse: %v", err)
			c.Next()
			return
		}

		for i, w := range windows {
			if w.limit <= 0 || counts[i].Val() <= int64(w.limit) {
				continue
			}
			abortAbuseBanned(c, recordAbuseOffense(ctx, redis, cfg, ip, class, windows, now, w))
			return
		}

		c.Next()
	}
}

// recordAbuseOffense bans ip and returns the ban length. Parallel requests
// that cross the limit together count as one offense: only the request that
// creates the ban key advances the offense counter. The tripped class's
// counters are cleared so the IP starts fresh once the ban expires instead of
// being re-banned by the rest of the hourly window.
func recordAbuseOffense(ctx context.Context, redis *redispkg.Client, cfg config.AbuseDetectionConfig, ip string, class abuseTrafficClass, windows []abuseWindow, now time.Time, tripped abuseWindow) time.Duration {
	banKey := abuseBanPrefix + ip
	offenseKey := abuseOffensePrefix + ip

	var previous int64
	if val, err := redis.Get(ctx, offenseKey); err == nil {
		previous, _ = strconv.ParseInt(val, 10, 64)
	}
	offense := previous + 1
	ban := abuseBanDuration(cfg.BanDurations, offense)

	created, err := redis.SetNX(ctx, banKey, strconv.FormatInt(offense, 10), ban)
	if err != nil {
		log.Printf("Error setting ban: %v", err)
		return ban
	}
	if !created {
		if _, ttl, err := redis.GetWithTTL(ctx, banKey); err == nil && ttl > 0 {
			return ttl
		}
		return ban
	}

	if _, err := redis.Increment(ctx, offenseKey); err != nil {
		log.Printf("Error recording abuse offense: %v", err)
	} else {
		_ = redis.Expire(ctx, offenseKey, cfg.OffenseWindow)
	}
	for _, w := range windows {
		_ = redis.Delete(ctx, abuseCountKey(class, w.name, ip, now.Unix()/int64(w.length/time.Second)))
	}
	log.Printf("IP %s blocked for %v (offense %d: more than %d %s requests per %v)",
		ip, ban, offense, tripped.limit, class, tripped.length)
	return ban
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

// UnbanIP removes a ban and the offense history for a specific IP (admin function)
func UnbanIP(ctx context.Context, redis *redispkg.Client, ip string) error {
	if err := redis.Delete(ctx, abuseBanPrefix+ip); err != nil {
		return err
	}
	return redis.Delete(ctx, abuseOffensePrefix+ip)
}

// GetBannedIPs returns a list of currently banned IPs (admin function)
func GetBannedIPs(ctx context.Context, redis *redispkg.Client) ([]string, error) {
	keys, err := redis.Keys(ctx, abuseBanPrefix+"*")
	if err != nil {
		return nil, err
	}

	ips := make([]string, 0, len(keys))
	for _, key := range keys {
		if len(key) > len(abuseBanPrefix) {
			ips = append(ips, key[len(abuseBanPrefix):])
		}
	}

	return ips, nil
}

// GetAbuseStats returns the IP's read plus write requests in the current
// hourly window (admin function).
func GetAbuseStats(ctx context.Context, redis *redispkg.Client, ip string) (int64, error) {
	bucket := time.Now().Unix() / int64(time.Hour/time.Second)
	var total int64
	for _, class := range []abuseTrafficClass{abuseClassRead, abuseClassWrite} {
		val, err := redis.Get(ctx, abuseCountKey(class, "h", ip, bucket))
		if errors.Is(err, goredis.Nil) {
			continue
		}
		if err != nil {
			return 0, err
		}
		count, err := strconv.ParseInt(val, 10, 64)
		if err != nil {
			return 0, fmt.Errorf("failed to parse abuse count: %w", err)
		}
		total += count
	}
	return total, nil
}
