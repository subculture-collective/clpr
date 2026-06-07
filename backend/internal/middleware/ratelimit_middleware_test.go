package middleware

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	redispkg "git.subcult.tv/subculture-collective/clpr/pkg/redis"
)

func TestRateLimitMiddleware_FallbackInitialization(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Reset global fallback limiters for test isolation
	ipFallbackLimiter = nil

	// Create the middleware with a low limit for testing
	// This should initialize the fallback limiter even if Redis is nil
	_ = RateLimitMiddleware((*redispkg.Client)(nil), 3, time.Second)

	// Verify that the fallback limiter was initialized
	if ipFallbackLimiter == nil {
		t.Fatal("ipFallbackLimiter should be initialized when creating RateLimitMiddleware")
	}

	// Verify the fallback limiter works correctly
	key := "test-key"

	// Make requests up to the limit
	for i := 0; i < 3; i++ {
		allowed, _ := ipFallbackLimiter.Allow(key)
		if !allowed {
			t.Errorf("request %d should be allowed", i+1)
		}
	}

	// The 4th request should be blocked
	allowed, remaining := ipFallbackLimiter.Allow(key)
	if allowed {
		t.Error("4th request should be blocked")
	}
	if remaining != 0 {
		t.Errorf("expected remaining=0, got %d", remaining)
	}
}

func TestRateLimitMiddleware_FallbackHeaders(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Reset global fallback limiters
	ipFallbackLimiter = NewInMemoryRateLimiter(5, time.Second)

	router := gin.New()
	// We'll create a middleware that always uses fallback by simulating Redis failure
	router.Use(func(c *gin.Context) {
		// Simulate using fallback
		key := fmt.Sprintf("ratelimit:%s:%s", c.Request.URL.Path, c.ClientIP())
		allowed, remaining := ipFallbackLimiter.Allow(key)

		if !allowed {
			c.Header("X-RateLimit-Limit", "5")
			c.Header("X-RateLimit-Remaining", "0")
			c.Header("X-RateLimit-Fallback", "true")
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": "Rate limit exceeded. Please try again later.",
			})
			c.Abort()
			return
		}

		c.Header("X-RateLimit-Limit", "5")
		c.Header("X-RateLimit-Remaining", fmt.Sprintf("%d", remaining))
		c.Header("X-RateLimit-Fallback", "true")
		c.Next()
	})

	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	// Make a request and check headers
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	req.RemoteAddr = "127.0.0.1:1234"
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}

	// Verify headers
	if limit := w.Header().Get("X-RateLimit-Limit"); limit != "5" {
		t.Errorf("expected X-RateLimit-Limit=5, got %s", limit)
	}

	if remaining := w.Header().Get("X-RateLimit-Remaining"); remaining == "" {
		t.Error("X-RateLimit-Remaining header should be set")
	}

	if fallback := w.Header().Get("X-RateLimit-Fallback"); fallback != "true" {
		t.Errorf("expected X-RateLimit-Fallback=true, got %s", fallback)
	}
}

func TestRateLimitByUserMiddleware_FallbackInitialization(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Reset global fallback limiters
	userFallbackLimiter = nil

	// Create the middleware
	_ = RateLimitByUserMiddleware((*redispkg.Client)(nil), 2, time.Second)

	// Verify fallback limiter is created
	if userFallbackLimiter == nil {
		t.Fatal("userFallbackLimiter should be initialized when creating RateLimitByUserMiddleware")
	}

	// Verify the fallback limiter works correctly
	key := "test-user-key"

	// Make requests up to the limit
	for i := 0; i < 2; i++ {
		allowed, _ := userFallbackLimiter.Allow(key)
		if !allowed {
			t.Errorf("request %d should be allowed", i+1)
		}
	}

	// The 3rd request should be blocked
	allowed, remaining := userFallbackLimiter.Allow(key)
	if allowed {
		t.Error("3rd request should be blocked")
	}
	if remaining != 0 {
		t.Errorf("expected remaining=0, got %d", remaining)
	}
}

func TestInMemoryRateLimiter_Integration(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Create a fresh limiter for this test
	limiter := NewInMemoryRateLimiter(5, 200*time.Millisecond)

	router := gin.New()
	router.Use(func(c *gin.Context) {
		key := c.ClientIP()
		allowed, remaining := limiter.Allow(key)

		if !allowed {
			c.Header("X-RateLimit-Limit", "5")
			c.Header("X-RateLimit-Remaining", "0")
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": "Rate limit exceeded",
			})
			c.Abort()
			return
		}

		c.Header("X-RateLimit-Remaining", fmt.Sprintf("%d", remaining))
		c.Next()
	})

	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	// Make 5 requests (should all succeed)
	for i := 0; i < 5; i++ {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/test", nil)
		req.RemoteAddr = "192.168.1.1:1234"
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("request %d: expected status 200, got %d", i+1, w.Code)
		}
	}

	// 6th request should fail
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	req.RemoteAddr = "192.168.1.1:1234"
	router.ServeHTTP(w, req)

	if w.Code != http.StatusTooManyRequests {
		t.Errorf("6th request: expected status 429, got %d", w.Code)
	}

	// Wait for window to expire
	time.Sleep(250 * time.Millisecond)

	// Should be able to make requests again
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/test", nil)
	req.RemoteAddr = "192.168.1.1:1234"
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("request after window: expected status 200, got %d", w.Code)
	}
}

func TestRateLimitMiddleware_AdminBypass(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(func(c *gin.Context) {
		// Simulate auth middleware setting admin role
		c.Set("user_id", "admin-123")
		c.Set("user_role", "admin")
		c.Next()
	})
	router.Use(func(c *gin.Context) {
		// Check admin bypass logic
		_, isAdmin := getUserRateLimitMultiplier(c, nil)
		if isAdmin {
			c.Header("X-RateLimit-Bypass", "admin")
			c.Next()
			return
		}
		// For non-admin, simulate rate limiting
		c.JSON(http.StatusTooManyRequests, gin.H{"error": "rate limited"})
		c.Abort()
	})
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	// Admin can make unlimited requests
	for i := 0; i < 10; i++ {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/test", nil)
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("admin request %d: expected status 200, got %d", i+1, w.Code)
		}

		// Check for bypass header
		if bypass := w.Header().Get("X-RateLimit-Bypass"); bypass != "admin" {
			t.Errorf("admin request %d: expected X-RateLimit-Bypass=admin, got %s", i+1, bypass)
		}
	}
}

func TestRateLimitMiddleware_PremiumMultiplier(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Reset global fallback limiters
	userFallbackLimiter = NewInMemoryRateLimiter(10, time.Second)

	router := gin.New()
	router.Use(func(c *gin.Context) {
		// Simulate auth middleware setting user with premium subscription
		c.Set("user_id", "premium-user-123")
		c.Set("user_role", "user")
		c.Set("subscription_tier", "pro")
		c.Next()
	})
	router.Use(func(c *gin.Context) {
		// Test multiplier calculation
		multiplier, _ := getUserRateLimitMultiplier(c, nil)
		baseLimit := 2
		effectiveLimit := int(float64(baseLimit) * multiplier)

		// Use fallback limiter with effective limit
		key := fmt.Sprintf("test-premium-%v", c.GetString("user_id"))
		allowed, remaining := userFallbackLimiter.Allow(key)

		if !allowed {
			c.JSON(http.StatusTooManyRequests, gin.H{"error": "rate limited"})
			c.Abort()
			return
		}

		c.Header("X-RateLimit-Limit", fmt.Sprintf("%d", effectiveLimit))
		c.Header("X-RateLimit-Remaining", fmt.Sprintf("%d", remaining))
		c.Next()
	})
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	// Premium user gets 5x multiplier, so 2 * 5 = 10 requests
	for i := 0; i < 10; i++ {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/test", nil)
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("premium request %d: expected status 200, got %d", i+1, w.Code)
		}

		// Check rate limit header shows effective limit
		if limit := w.Header().Get("X-RateLimit-Limit"); limit != "10" {
			t.Errorf("premium request %d: expected X-RateLimit-Limit=10, got %s", i+1, limit)
		}
	}

	// 11th request should be blocked
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusTooManyRequests {
		t.Errorf("11th premium request: expected status 429, got %d", w.Code)
	}
}

func TestRateLimitMiddleware_BasicUserLimit(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Reset global fallback limiters
	userFallbackLimiter = NewInMemoryRateLimiter(3, time.Second)

	router := gin.New()
	router.Use(func(c *gin.Context) {
		// Simulate auth middleware setting basic user (free tier)
		c.Set("user_id", "basic-user-456")
		c.Set("user_role", "user")
		c.Set("subscription_tier", "free")
		c.Next()
	})
	router.Use(func(c *gin.Context) {
		// Test multiplier calculation
		multiplier, _ := getUserRateLimitMultiplier(c, nil)
		baseLimit := 3
		effectiveLimit := int(float64(baseLimit) * multiplier)

		// Use fallback limiter
		key := fmt.Sprintf("test-basic-%v", c.GetString("user_id"))
		allowed, remaining := userFallbackLimiter.Allow(key)

		if !allowed {
			c.JSON(http.StatusTooManyRequests, gin.H{"error": "rate limited"})
			c.Abort()
			return
		}

		c.Header("X-RateLimit-Limit", fmt.Sprintf("%d", effectiveLimit))
		c.Header("X-RateLimit-Remaining", fmt.Sprintf("%d", remaining))
		c.Next()
	})
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	// Basic user gets 1x multiplier, so 3 * 1 = 3 requests
	for i := 0; i < 3; i++ {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/test", nil)
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("basic request %d: expected status 200, got %d", i+1, w.Code)
		}

		// Check rate limit header shows effective limit
		if limit := w.Header().Get("X-RateLimit-Limit"); limit != "3" {
			t.Errorf("basic request %d: expected X-RateLimit-Limit=3, got %s", i+1, limit)
		}
	}

	// 4th request should be blocked
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusTooManyRequests {
		t.Errorf("4th basic request: expected status 429, got %d", w.Code)
	}
}

func TestGetUserRateLimitMultiplier(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Generate test UUIDs
	adminUUID := uuid.New()
	premiumUUID := uuid.New()
	basicUUID := uuid.New()

	tests := []struct {
		name             string
		userID           interface{}
		userRole         string
		subscriptionTier string
		wantMultiplier   float64
		wantIsAdmin      bool
	}{
		{
			name:           "unauthenticated user",
			wantMultiplier: 1.0,
			wantIsAdmin:    false,
		},
		{
			name:           "admin user with UUID",
			userID:         adminUUID,
			userRole:       "admin",
			wantMultiplier: 0,
			wantIsAdmin:    true,
		},
		{
			name:           "admin user with string UUID",
			userID:         adminUUID.String(),
			userRole:       "admin",
			wantMultiplier: 0,
			wantIsAdmin:    true,
		},
		{
			name:             "premium user with UUID",
			userID:           premiumUUID,
			userRole:         "user",
			subscriptionTier: "pro",
			wantMultiplier:   5.0,
			wantIsAdmin:      false,
		},
		{
			name:             "basic user with UUID",
			userID:           basicUUID,
			userRole:         "user",
			subscriptionTier: "free",
			wantMultiplier:   1.0,
			wantIsAdmin:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, _ := gin.CreateTestContext(httptest.NewRecorder())

			if tt.userID != nil {
				c.Set("user_id", tt.userID)
			}
			if tt.userRole != "" {
				c.Set("user_role", tt.userRole)
			}
			if tt.subscriptionTier != "" {
				c.Set("subscription_tier", tt.subscriptionTier)
			}

			multiplier, isAdmin := getUserRateLimitMultiplier(c, nil)

			if multiplier != tt.wantMultiplier {
				t.Errorf("got multiplier=%f, want %f", multiplier, tt.wantMultiplier)
			}
			if isAdmin != tt.wantIsAdmin {
				t.Errorf("got isAdmin=%v, want %v", isAdmin, tt.wantIsAdmin)
			}
		})
	}
}

func TestInitRateLimitWhitelist(t *testing.T) {
	tests := []struct {
		name         string
		whitelistIPs string
		checkIPs     map[string]bool // IP -> should be whitelisted
	}{
		{
			name:         "empty whitelist",
			whitelistIPs: "",
			checkIPs: map[string]bool{
				"127.0.0.1":      true,  // localhost always included
				"::1":            true,  // IPv6 localhost always included
				"192.168.1.1":    false, // not whitelisted
				"173.165.22.142": false, // not whitelisted
			},
		},
		{
			name:         "single IP",
			whitelistIPs: "192.168.1.100",
			checkIPs: map[string]bool{
				"127.0.0.1":     true,  // localhost always included
				"::1":           true,  // IPv6 localhost always included
				"192.168.1.100": true,  // whitelisted
				"192.168.1.101": false, // not whitelisted
			},
		},
		{
			name:         "multiple IPs",
			whitelistIPs: "192.168.1.100,10.0.0.50,173.165.22.142",
			checkIPs: map[string]bool{
				"127.0.0.1":      true,  // localhost always included
				"::1":            true,  // IPv6 localhost always included
				"192.168.1.100":  true,  // whitelisted
				"10.0.0.50":      true,  // whitelisted
				"173.165.22.142": true,  // whitelisted
				"192.168.1.101":  false, // not whitelisted
			},
		},
		{
			name:         "IPs with spaces",
			whitelistIPs: " 192.168.1.100 , 10.0.0.50 ",
			checkIPs: map[string]bool{
				"192.168.1.100": true,  // whitelisted (trimmed)
				"10.0.0.50":     true,  // whitelisted (trimmed)
				"192.168.1.101": false, // not whitelisted
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Initialize whitelist with test data
			InitRateLimitWhitelist(tt.whitelistIPs)

			// Check each IP using thread-safe accessor
			for ip, expectedWhitelisted := range tt.checkIPs {
				actualWhitelisted := isIPWhitelisted(ip)
				if actualWhitelisted != expectedWhitelisted {
					t.Errorf("IP %s: got whitelisted=%v, want %v", ip, actualWhitelisted, expectedWhitelisted)
				}
			}
		})
	}
}
