package services

import (
	"context"
	"fmt"
	"time"

	redispkg "git.subcult.tv/subculture-collective/clpr/pkg/redis"
)

// DistributedRateLimiter provides a Redis-backed distributed rate limiter
// using a sliding window algorithm. This ensures rate limits are enforced
// across multiple server instances.
type DistributedRateLimiter struct {
	redisClient *redispkg.Client
	limit       int
	window      time.Duration
}

// Rate limiter configuration constants
const (
	// rateLimitExpireBuffer is the additional time to keep rate limit keys in Redis
	// beyond the window duration to ensure proper cleanup
	rateLimitExpireBuffer = time.Minute
)

// NewDistributedRateLimiter creates a new distributed rate limiter
// limit: maximum number of requests allowed in the window
// window: time window for rate limiting
func NewDistributedRateLimiter(redisClient *redispkg.Client, limit int, window time.Duration) *DistributedRateLimiter {
	return &DistributedRateLimiter{
		redisClient: redisClient,
		limit:       limit,
		window:      window,
	}
}

// Allow checks if a request should be allowed for the given key.
// Uses Redis sorted sets with sliding window algorithm for accurate rate limiting.
// The entire operation is atomic via Lua script to prevent race conditions.
// Returns true if the request is allowed, false if rate limit is exceeded.
func (r *DistributedRateLimiter) Allow(ctx context.Context, key string) (bool, error) {
	now := time.Now()
	windowStart := now.Add(-r.window)

	// Use a Redis key prefix to namespace rate limit keys
	redisKey := fmt.Sprintf("ratelimit:%s", key)

	// Get the underlying Redis client for direct operations
	client := r.redisClient.GetClient()

	// Use Lua script for atomic check-and-increment operation
	// This prevents race conditions where multiple requests could exceed the limit
	luaScript := `
		local key = KEYS[1]
		local window_start = tonumber(ARGV[1])
		local limit = tonumber(ARGV[2])
		local now_score = tonumber(ARGV[3])
		local member = ARGV[4]
		local expire_time = tonumber(ARGV[5])
		
		-- Remove old entries outside the window
		redis.call('ZREMRANGEBYSCORE', key, '0', window_start)
		
		-- Count current entries in the window
		local count = redis.call('ZCARD', key)
		
		-- Check if limit exceeded
		if count >= limit then
			return 0
		end
		
		-- Add current request
		redis.call('ZADD', key, now_score, member)
		
		-- Set expiration
		redis.call('EXPIRE', key, expire_time)
		
		return 1
	`

	member := fmt.Sprintf("%d", now.UnixNano())
	score := now.UnixMilli()
	expireSeconds := int64((r.window + rateLimitExpireBuffer).Seconds())

	result, err := client.Eval(ctx, luaScript, []string{redisKey},
		windowStart.UnixMilli(), r.limit, score, member, expireSeconds).Result()

	if err != nil {
		return false, fmt.Errorf("failed to check rate limit: %w", err)
	}

	// Result is 1 if allowed, 0 if denied
	allowed, ok := result.(int64)
	if !ok {
		return false, fmt.Errorf("unexpected result type from rate limit script")
	}

	return allowed == 1, nil
}

// RateLimiter interface for abstraction
type RateLimiter interface {
	Allow(ctx context.Context, key string) (bool, error)
}

// InMemoryRateLimiterAdapter adapts SimpleRateLimiter to the RateLimiter interface
// This is used as a fallback when Redis is not available
type InMemoryRateLimiterAdapter struct {
	limiter *SimpleRateLimiter
}

// NewInMemoryRateLimiterAdapter creates an adapter for SimpleRateLimiter
func NewInMemoryRateLimiterAdapter(limit int, window time.Duration) *InMemoryRateLimiterAdapter {
	return &InMemoryRateLimiterAdapter{
		limiter: NewSimpleRateLimiter(limit, window),
	}
}

// Allow implements the RateLimiter interface for in-memory rate limiting
func (a *InMemoryRateLimiterAdapter) Allow(ctx context.Context, key string) (bool, error) {
	return a.limiter.Allow(key), nil
}
