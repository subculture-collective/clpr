package twitch

import (
	"context"
	"sync"
	"time"

	"git.subcult.tv/subculture-collective/clpr/pkg/utils"
)

// RateLimiter implements token bucket rate limiting for Twitch API
type RateLimiter struct {
	tokens    int
	maxTokens int
	refillAt  time.Time
	mu        sync.Mutex
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(maxTokens int) *RateLimiter {
	return &RateLimiter{
		tokens:    maxTokens,
		maxTokens: maxTokens,
		refillAt:  time.Now().Add(time.Minute),
	}
}

// Wait blocks until a token is available or context is cancelled
func (rl *RateLimiter) Wait(ctx context.Context) error {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	// Refill tokens if minute has passed
	if time.Now().After(rl.refillAt) {
		rl.tokens = rl.maxTokens
		rl.refillAt = time.Now().Add(time.Minute)
	}

	// Wait if no tokens available
	if rl.tokens <= 0 {
		waitTime := time.Until(rl.refillAt)
		if waitTime > 0 {
			logger := utils.GetLogger()
			logger.Warn("Rate limit reached, waiting", map[string]interface{}{
				"wait_time": waitTime.String(),
			})

			// Release lock while waiting to avoid blocking other operations
			rl.mu.Unlock()
			timer := time.NewTimer(waitTime)

			select {
			case <-timer.C:
				// Reacquire lock and refill
				rl.mu.Lock()
				// Recheck state as it may have changed
				if time.Now().After(rl.refillAt) {
					rl.tokens = rl.maxTokens
					rl.refillAt = time.Now().Add(time.Minute)
				}
				// Don't defer unlock here as we already have deferred unlock at function level
				return nil
			case <-ctx.Done():
				timer.Stop()
				// Need to reacquire lock to satisfy defer unlock
				rl.mu.Lock()
				return ctx.Err()
			}
		}
	}

	rl.tokens--
	return nil
}

// Available returns the number of tokens currently available
func (rl *RateLimiter) Available() int {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	// Refill if needed
	if time.Now().After(rl.refillAt) {
		rl.tokens = rl.maxTokens
		rl.refillAt = time.Now().Add(time.Minute)
	}

	return rl.tokens
}

// ChannelRateLimiter manages per-channel rate limits for moderation actions
// Twitch has per-channel rate limits for ban/unban operations to prevent abuse
// Note: To prevent unbounded memory growth in production, consider implementing
// periodic cleanup of inactive channels or using an LRU cache with a maximum size.
type ChannelRateLimiter struct {
	limiters  map[string]*rateLimiterEntry
	mu        sync.RWMutex
	maxTokens int
}

// rateLimiterEntry wraps a rate limiter with last access time for cleanup
type rateLimiterEntry struct {
	limiter      *RateLimiter
	lastAccessed time.Time
}

// NewChannelRateLimiter creates a new per-channel rate limiter
// maxTokens: maximum number of requests per channel per minute
func NewChannelRateLimiter(maxTokens int) *ChannelRateLimiter {
	return &ChannelRateLimiter{
		limiters:  make(map[string]*rateLimiterEntry),
		maxTokens: maxTokens,
	}
}

// Wait blocks until a token is available for the specified channel
func (crl *ChannelRateLimiter) Wait(ctx context.Context, channelID string) error {
	limiter := crl.getLimiter(channelID)
	return limiter.Wait(ctx)
}

// getLimiter gets or creates a rate limiter for a specific channel
func (crl *ChannelRateLimiter) getLimiter(channelID string) *RateLimiter {
	// First try read lock for fast path
	crl.mu.RLock()
	entry, exists := crl.limiters[channelID]
	crl.mu.RUnlock()

	if exists {
		// Update last accessed time under write lock to avoid race condition
		crl.mu.Lock()
		entry.lastAccessed = time.Now()
		crl.mu.Unlock()
		return entry.limiter
	}

	// Need to create new limiter, acquire write lock
	crl.mu.Lock()
	defer crl.mu.Unlock()

	// Check again in case another goroutine created it
	if entry, exists := crl.limiters[channelID]; exists {
		entry.lastAccessed = time.Now()
		return entry.limiter
	}

	// Create new limiter for this channel
	limiter := NewRateLimiter(crl.maxTokens)
	crl.limiters[channelID] = &rateLimiterEntry{
		limiter:      limiter,
		lastAccessed: time.Now(),
	}
	return limiter
}

// Available returns the number of tokens currently available for a channel
func (crl *ChannelRateLimiter) Available(channelID string) int {
	limiter := crl.getLimiter(channelID)
	return limiter.Available()
}

// CleanupInactive removes rate limiters for channels that haven't been accessed
// within the specified duration. This should be called periodically to prevent
// unbounded memory growth. Returns the number of limiters removed.
func (crl *ChannelRateLimiter) CleanupInactive(inactiveDuration time.Duration) int {
	crl.mu.Lock()
	defer crl.mu.Unlock()

	now := time.Now()
	removed := 0

	for channelID, entry := range crl.limiters {
		if now.Sub(entry.lastAccessed) > inactiveDuration {
			delete(crl.limiters, channelID)
			removed++
		}
	}

	return removed
}
