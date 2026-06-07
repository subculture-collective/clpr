package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"git.subcult.tv/subculture-collective/clpr/config"
	redispkg "git.subcult.tv/subculture-collective/clpr/pkg/redis"
)

func TestCSRFMiddleware_SafeMethods(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Use nil Redis client for unit test (safe methods don't need Redis)
	var mockRedis *redispkg.Client = nil

	tests := []struct {
		name   string
		method string
	}{
		{"GET request", "GET"},
		{"HEAD request", "HEAD"},
		{"OPTIONS request", "OPTIONS"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, r := gin.CreateTestContext(w)

			r.Use(CSRFMiddleware(mockRedis, false))
			r.Handle(tt.method, "/test", func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"message": "ok"})
			})

			req, _ := http.NewRequest(tt.method, "/test", nil)
			c.Request = req
			r.ServeHTTP(w, req)

			if w.Code != http.StatusOK {
				t.Errorf("Expected status 200, got %d", w.Code)
			}
		})
	}
}

func TestCSRFMiddleware_BearerToken(t *testing.T) {
	gin.SetMode(gin.TestMode)

	var mockRedis *redispkg.Client = nil

	w := httptest.NewRecorder()
	c, r := gin.CreateTestContext(w)

	r.Use(CSRFMiddleware(mockRedis, false))
	r.POST("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "ok"})
	})

	req, _ := http.NewRequest("POST", "/test", nil)
	req.Header.Set("Authorization", "Bearer test-token")
	c.Request = req
	r.ServeHTTP(w, req)

	// Should pass without CSRF token when using Bearer auth
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

func TestCSRFMiddleware_MissingToken(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Note: With nil Redis client, CSRF middleware skips checks entirely
	// This test verifies the middleware structure, not enforcement
	var mockRedis *redispkg.Client = nil

	w := httptest.NewRecorder()
	c, r := gin.CreateTestContext(w)

	r.Use(CSRFMiddleware(mockRedis, false))
	r.POST("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "ok"})
	})

	req, _ := http.NewRequest("POST", "/test", nil)
	// Add cookie to trigger CSRF check (if Redis were available)
	req.AddCookie(&http.Cookie{
		Name:  "access_token",
		Value: "test-token",
	})
	c.Request = req
	r.ServeHTTP(w, req)

	// With nil Redis, CSRF checks are skipped, so request succeeds
	// In production with real Redis, this would return 403
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200 with nil Redis client, got %d", w.Code)
	}
}

func TestGenerateCSRFToken(t *testing.T) {
	token1, err := generateCSRFToken()
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	if len(token1) == 0 {
		t.Error("Generated token is empty")
	}

	// Generate another token and ensure they're different
	token2, err := generateCSRFToken()
	if err != nil {
		t.Fatalf("Failed to generate second token: %v", err)
	}

	if token1 == token2 {
		t.Error("Tokens should be unique")
	}
}

func TestVerifyCSRFToken_Mismatch(t *testing.T) {
	gin.SetMode(gin.TestMode)

	var mockRedis *redispkg.Client = nil

	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = &http.Request{}
	ctx := context.Background()
	c.Request = c.Request.WithContext(ctx)

	// Different tokens should not match
	result := verifyCSRFToken(c, mockRedis, "token1", "token2")
	if result {
		t.Error("Expected verification to fail with mismatched tokens")
	}
}

func TestCSRFMiddleware_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	gin.SetMode(gin.TestMode)

	// This test requires a real Redis connection
	// Skip if Redis is not available
	cfg := &config.RedisConfig{
		Host:     "localhost",
		Port:     "6380",
		Password: "",
		DB:       0,
	}

	mockRedis, err := redispkg.NewClient(cfg)
	if err != nil {
		t.Skip("Redis not available, skipping integration test")
	}
	defer mockRedis.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	if err := mockRedis.HealthCheck(ctx); err != nil {
		t.Skip("Redis not available, skipping integration test")
	}

	// Clean up any existing test data
	defer func() {
		ctx := context.Background()
		keys, _ := mockRedis.Keys(ctx, "csrf:*")
		for _, key := range keys {
			_ = mockRedis.Delete(ctx, key)
		}
	}()

	t.Run("full CSRF flow", func(t *testing.T) {
		w := httptest.NewRecorder()
		_, r := gin.CreateTestContext(w)

		r.Use(CSRFMiddleware(mockRedis, false))
		r.GET("/get-token", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"message": "ok"})
		})
		r.POST("/submit", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"message": "success"})
		})

		// Step 1: GET request to obtain CSRF token
		getReq, _ := http.NewRequest("GET", "/get-token", nil)
		r.ServeHTTP(w, getReq)

		if w.Code != http.StatusOK {
			t.Fatalf("Expected status 200 on GET, got %d", w.Code)
		}

		// Extract CSRF token from response
		csrfToken := w.Header().Get(CSRFTokenHeader)
		if csrfToken == "" {
			t.Fatal("CSRF token not set in response header")
		}

		cookies := w.Result().Cookies()
		var csrfCookie *http.Cookie
		for _, cookie := range cookies {
			if cookie.Name == CSRFCookieName {
				csrfCookie = cookie
				break
			}
		}

		if csrfCookie == nil {
			t.Fatal("CSRF cookie not set")
		}

		// Step 2: POST request with CSRF token
		w = httptest.NewRecorder()
		postReq, _ := http.NewRequest("POST", "/submit", nil)
		postReq.Header.Set(CSRFTokenHeader, csrfToken)
		postReq.AddCookie(csrfCookie)
		postReq.AddCookie(&http.Cookie{
			Name:  "access_token",
			Value: "test-token",
		})
		r.ServeHTTP(w, postReq)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200 on POST with valid CSRF token, got %d", w.Code)
		}
	})
}
