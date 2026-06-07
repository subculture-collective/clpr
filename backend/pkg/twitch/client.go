package twitch

// TWITCH COMPLIANCE:
// This package implements Twitch API integration following Twitch's Developer Services Agreement.
// See: https://legal.twitch.com/legal/developer-agreement/
// See: https://dev.twitch.tv/docs/api/
// See: docs/compliance/twitch-api-usage.md for full compliance documentation
//
// COMPLIANCE REQUIREMENTS:
// - Uses ONLY official Twitch Helix API (no scraping, no unofficial endpoints)
// - Respects 800 requests/minute rate limit via token bucket algorithm
// - Implements proper caching to reduce API load
// - Handles authentication via OAuth 2.0 (app access tokens + user access tokens)
// - Never re-hosts or proxies video files (only metadata)
// - Stores only public data or user-authorized data

import (
	"context"
	"crypto/rand"
	"fmt"
	"math/big"
	"net/http"
	"net/url"
	"sync"
	"time"

	"git.subcult.tv/subculture-collective/clpr/config"
	redispkg "git.subcult.tv/subculture-collective/clpr/pkg/redis"
	"git.subcult.tv/subculture-collective/clpr/pkg/utils"
)

const (
	// baseURL is the official Twitch Helix API endpoint
	// COMPLIANCE: Must ONLY use official Twitch API endpoints
	// See: https://dev.twitch.tv/docs/api/reference
	baseURL = "https://api.twitch.tv/helix"

	// rateLimitPerMin enforces Twitch's rate limit of 800 requests per minute
	// COMPLIANCE: Twitch enforces 800 req/min limit, we must respect it
	// See: https://dev.twitch.tv/docs/api/guide/#rate-limits
	rateLimitPerMin = 800
)

// Client wraps the Twitch API with authentication, rate limiting, and caching
type Client struct {
	clientID           string
	httpClient         *http.Client
	cache              TwitchCache
	authManager        *AuthManager
	rateLimiter        *RateLimiter
	channelRateLimiter *ChannelRateLimiter
	circuitBreaker     *CircuitBreaker
}

// CircuitBreaker implements circuit breaker pattern for API availability
type CircuitBreaker struct {
	mu           sync.RWMutex
	failureCount int
	lastFailure  time.Time
	state        string // "closed", "open", "half-open"
	failureLimit int
	timeout      time.Duration
}

// NewCircuitBreaker creates a new circuit breaker
func NewCircuitBreaker(failureLimit int, timeout time.Duration) *CircuitBreaker {
	return &CircuitBreaker{
		state:        "closed",
		failureLimit: failureLimit,
		timeout:      timeout,
	}
}

// Allow checks if requests should be allowed
func (cb *CircuitBreaker) Allow() error {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	if cb.state == "open" {
		if time.Since(cb.lastFailure) > cb.timeout {
			// Transition to half-open state
			cb.state = "half-open"
			return nil
		}
		return &CircuitBreakerError{Message: "circuit breaker is open, API unavailable"}
	}

	return nil
}

// RecordSuccess records a successful request
func (cb *CircuitBreaker) RecordSuccess() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	if cb.state == "half-open" {
		cb.state = "closed"
		cb.failureCount = 0
	} else if cb.state == "closed" {
		// Reset failure count on success in closed state
		cb.failureCount = 0
	}
}

// RecordFailure records a failed request
func (cb *CircuitBreaker) RecordFailure() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.failureCount++
	cb.lastFailure = time.Now()

	if cb.failureCount >= cb.failureLimit {
		cb.state = "open"
		logger := utils.GetLogger()
		logger.Warn("Circuit breaker opening", map[string]interface{}{
			"failure_count": cb.failureCount,
			"component":     "twitch_client",
		})
	}
}

// NewClient creates a new Twitch API client
func NewClient(cfg *config.TwitchConfig, redis *redispkg.Client) (*Client, error) {
	if cfg.ClientID == "" || cfg.ClientSecret == "" {
		return nil, fmt.Errorf("twitch client ID and secret are required")
	}

	httpClient := &http.Client{
		Timeout: 30 * time.Second,
	}

	cache := NewRedisCache(redis)
	authManager := NewAuthManager(cfg.ClientID, cfg.ClientSecret, httpClient, cache)
	rateLimiter := NewRateLimiter(rateLimitPerMin)
	// Per-channel rate limiter: 100 moderation actions per channel per minute
	channelRateLimiter := NewChannelRateLimiter(100)
	circuitBreaker := NewCircuitBreaker(5, 30*time.Second)

	client := &Client{
		clientID:           cfg.ClientID,
		httpClient:         httpClient,
		cache:              cache,
		authManager:        authManager,
		rateLimiter:        rateLimiter,
		channelRateLimiter: channelRateLimiter,
		circuitBreaker:     circuitBreaker,
	}

	// Try to load token from cache
	if err := authManager.LoadFromCache(context.Background()); err != nil {
		logger := utils.GetLogger()
		logger.Warn("Failed to load token from cache", map[string]interface{}{
			"error": err.Error(),
		})
		// Get a new token
		if err := authManager.RefreshToken(context.Background()); err != nil {
			return nil, fmt.Errorf("failed to get access token: %w", err)
		}
	}

	return client, nil
}

// doRequest performs an HTTP request with authentication, rate limiting, retry logic, and circuit breaker
// nolint:gocyclo // Complexity stems from retry and status handling; kept readable.
func (c *Client) doRequest(ctx context.Context, method, endpoint string, params url.Values) (*http.Response, error) {
	// Check circuit breaker
	if err := c.circuitBreaker.Allow(); err != nil {
		return nil, err
	}

	// Get valid token
	token, err := c.authManager.GetToken(ctx)
	if err != nil {
		c.circuitBreaker.RecordFailure()
		return nil, err
	}

	// Apply rate limiting
	if err := c.rateLimiter.Wait(ctx); err != nil {
		return nil, fmt.Errorf("rate limit wait cancelled: %w", err)
	}

	// Build URL
	reqURL := baseURL + endpoint
	if len(params) > 0 {
		reqURL += "?" + params.Encode()
	}

	// Retry logic with exponential backoff
	var resp *http.Response
	maxRetries := 3
	baseDelay := time.Second

	for attempt := 0; attempt < maxRetries; attempt++ {
		req, reqErr := http.NewRequestWithContext(ctx, method, reqURL, http.NoBody)
		if reqErr != nil {
			return nil, fmt.Errorf("failed to create request: %w", reqErr)
		}

		req.Header.Set("Authorization", "Bearer "+token) // #nosec G101 (value is an OAuth token, not hardcoded secret)
		req.Header.Set("Client-Id", c.clientID)

		logger := utils.GetLogger()
		logger.Debug("Twitch API request", map[string]interface{}{
			"method":   method,
			"endpoint": endpoint,
		})

		resp, err = c.httpClient.Do(req)
		if err != nil {
			c.circuitBreaker.RecordFailure()
			if attempt < maxRetries-1 {
				delay := baseDelay * time.Duration(1<<uint(attempt))
				logger := utils.GetLogger()
				logger.Warn("Request failed, retrying", map[string]interface{}{
					"attempt": attempt + 1,
					"max":     maxRetries,
					"delay":   delay.String(),
					"error":   err.Error(),
				})
				time.Sleep(delay)
				continue
			}
			return nil, fmt.Errorf("request failed after %d attempts: %w", maxRetries, err)
		}

		// Handle specific status codes
		switch resp.StatusCode {
		case http.StatusOK:
			c.circuitBreaker.RecordSuccess()
			return resp, nil
		case http.StatusUnauthorized:
			resp.Body.Close()
			// Token might be invalid, refresh and retry
			if err := c.authManager.RefreshToken(ctx); err != nil {
				c.circuitBreaker.RecordFailure()
				return nil, &AuthError{Message: "failed to refresh token", Err: err}
			}
			// Get new token for retry
			token, err = c.authManager.GetToken(ctx)
			if err != nil {
				c.circuitBreaker.RecordFailure()
				return nil, &AuthError{Message: "failed to get token after refresh", Err: err}
			}
			if attempt < maxRetries-1 {
				logger := utils.GetLogger()
				logger.Info("Token refreshed, retrying request", map[string]interface{}{
					"attempt": attempt + 1,
					"max":     maxRetries,
				})
				continue
			}
		case http.StatusTooManyRequests:
			resp.Body.Close()
			// Rate limited by Twitch, back off
			delay := baseDelay * time.Duration(1<<uint(attempt))
			if attempt < maxRetries-1 {
				logger := utils.GetLogger()
				logger.Warn("Rate limited by Twitch", map[string]interface{}{
					"attempt": attempt + 1,
					"max":     maxRetries,
					"delay":   delay.String(),
				})
				time.Sleep(delay)
				continue
			}
			return nil, &RateLimitError{Message: "rate limited by Twitch", RetryAfter: int(delay.Seconds())}
		case http.StatusServiceUnavailable, http.StatusBadGateway, http.StatusGatewayTimeout:
			resp.Body.Close()
			// Twitch is down, retry with backoff
			c.circuitBreaker.RecordFailure()
			delay := baseDelay * time.Duration(1<<uint(attempt))
			if attempt < maxRetries-1 {
				logger := utils.GetLogger()
				logger.Warn("Twitch service unavailable", map[string]interface{}{
					"attempt":     attempt + 1,
					"max":         maxRetries,
					"delay":       delay.String(),
					"status_code": resp.StatusCode,
				})
				time.Sleep(delay)
				continue
			}
		case http.StatusNotFound:
			// Don't retry on 404, don't count as failure
			c.circuitBreaker.RecordSuccess()
			return resp, nil
		default:
			// Other errors, don't retry but don't count as circuit breaker failure
			c.circuitBreaker.RecordSuccess()
			return resp, nil
		}
	}

	c.circuitBreaker.RecordFailure()
	return resp, fmt.Errorf("request failed after %d attempts", maxRetries)
}

// GetCachedUser retrieves user data from cache
func (c *Client) GetCachedUser(ctx context.Context, userID string) (*User, error) {
	return c.cache.CachedUser(ctx, userID)
}

// GetCachedGame retrieves game data from cache
func (c *Client) GetCachedGame(ctx context.Context, gameID string) (*Game, error) {
	return c.cache.CachedGame(ctx, gameID)
}

// jitteredBackoff calculates exponential backoff with jitter using crypto/rand for thread safety
// attempt: retry attempt number (0-indexed)
// baseDelay: base delay duration
// maxDelay: maximum delay duration
// Returns a duration with random jitter applied
// Uses the "Decorrelated Jitter" approach: returns delay/2 + random(0, delay/2)
// This ensures a minimum delay of delay/2 while still providing randomization
func jitteredBackoff(attempt int, baseDelay, maxDelay time.Duration) time.Duration {
	// Cap attempt to prevent overflow in exponential calculation
	// On 64-bit systems, 1<<63 would overflow, so we cap at 62
	if attempt > 62 {
		attempt = 62
	}

	// Exponential backoff: baseDelay * 2^attempt
	delay := baseDelay * time.Duration(1<<uint(attempt))

	// Cap at max delay
	if delay > maxDelay {
		delay = maxDelay
	}

	// Calculate jitter range: delay/2 to delay
	// This ensures minimum backoff of delay/2 while providing randomization
	halfDelay := delay / 2

	// Prevent overflow and ensure we have a valid range
	if halfDelay <= 0 {
		// Fallback to 75% of delay for edge cases
		return delay * 3 / 4
	}

	maxJitter := big.NewInt(int64(halfDelay))

	// Use crypto/rand for thread-safe random number generation
	jitterBig, err := rand.Int(rand.Reader, maxJitter)
	if err != nil {
		// Fallback to 75% of delay if random generation fails
		return delay * 3 / 4
	}

	// Return delay/2 + random(0, delay/2)
	return halfDelay + time.Duration(jitterBig.Int64())
}
