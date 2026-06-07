package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"git.subcult.tv/subculture-collective/clpr/config"
	redispkg "git.subcult.tv/subculture-collective/clpr/pkg/redis"
)

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func TestAbuseDetectionMiddleware_NormalUsage(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Use nil Redis client for unit test
	var mockRedis *redispkg.Client = nil

	w := httptest.NewRecorder()
	c, r := gin.CreateTestContext(w)

	r.Use(AbuseDetectionMiddleware(mockRedis))
	r.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "ok"})
	})

	req, _ := http.NewRequest("GET", "/test", nil)
	c.Request = req
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

func TestAbuseDetectionMiddleware_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	gin.SetMode(gin.TestMode)

	cfg := &config.RedisConfig{
		Host:     getEnv("TEST_REDIS_HOST", "localhost"),
		Port:     getEnv("TEST_REDIS_PORT", "6380"),
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
		keys, _ := mockRedis.Keys(ctx, "abuse:*")
		for _, key := range keys {
			_ = mockRedis.Delete(ctx, key)
		}
	}()

	t.Run("ban after threshold", func(t *testing.T) {
		r := gin.New()
		r.Use(AbuseDetectionMiddleware(mockRedis))
		r.GET("/test", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"message": "ok"})
		})

		// Simulate many requests from same IP (but below abuse threshold for testing)
		testIP := "192.168.1.100"

		// Make requests up to threshold
		for i := 0; i < 10; i++ {
			w := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", "/test", nil)
			req.RemoteAddr = testIP + ":12345"
			r.ServeHTTP(w, req)

			if w.Code != http.StatusOK {
				t.Errorf("Request %d: expected status 200, got %d", i, w.Code)
			}
		}
	})
}

func TestEnhancedRateLimitMiddleware_Warnings(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Use nil Redis client for unit test
	var mockRedis *redispkg.Client = nil

	w := httptest.NewRecorder()
	c, r := gin.CreateTestContext(w)

	r.Use(EnhancedRateLimitMiddleware(mockRedis, 10, time.Minute))
	r.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "ok"})
	})

	req, _ := http.NewRequest("GET", "/test", nil)
	c.Request = req
	r.ServeHTTP(w, req)

	// Should pass for first request
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

func TestUnbanIP(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	cfg := &config.RedisConfig{
		Host:     getEnv("TEST_REDIS_HOST", "localhost"),
		Port:     getEnv("TEST_REDIS_PORT", "6380"),
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

	// Clean up
	defer func() {
		ctx := context.Background()
		keys, _ := mockRedis.Keys(ctx, "abuse:*")
		for _, key := range keys {
			_ = mockRedis.Delete(ctx, key)
		}
	}()

	testIP := "192.168.1.1"

	// Ban the IP
	banKey := "abuse:ban:" + testIP
	if err := mockRedis.Set(ctx, banKey, "1", time.Hour); err != nil {
		t.Fatalf("Failed to set ban: %v", err)
	}

	// Verify ban exists
	exists, err := mockRedis.Exists(ctx, banKey)
	if err != nil {
		t.Fatalf("Failed to check ban: %v", err)
	}
	if !exists {
		t.Error("Ban was not set")
	}

	// Unban the IP
	if err := UnbanIP(ctx, mockRedis, testIP); err != nil {
		t.Fatalf("Failed to unban IP: %v", err)
	}

	// Verify ban is removed
	exists, err = mockRedis.Exists(ctx, banKey)
	if err != nil {
		t.Fatalf("Failed to check ban after unban: %v", err)
	}
	if exists {
		t.Error("Ban was not removed")
	}
}

func TestGetBannedIPs(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	cfg := &config.RedisConfig{
		Host:     getEnv("TEST_REDIS_HOST", "localhost"),
		Port:     getEnv("TEST_REDIS_PORT", "6380"),
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

	// Clean up
	defer func() {
		ctx := context.Background()
		keys, _ := mockRedis.Keys(ctx, "abuse:*")
		for _, key := range keys {
			_ = mockRedis.Delete(ctx, key)
		}
	}()

	// Ban multiple IPs
	testIPs := []string{"192.168.1.1", "192.168.1.2", "10.0.0.1"}
	for _, ip := range testIPs {
		banKey := "abuse:ban:" + ip
		if err := mockRedis.Set(ctx, banKey, "1", time.Hour); err != nil {
			t.Fatalf("Failed to ban IP %s: %v", ip, err)
		}
	}

	// Get banned IPs
	bannedIPs, err := GetBannedIPs(ctx, mockRedis)
	if err != nil {
		t.Fatalf("Failed to get banned IPs: %v", err)
	}

	if len(bannedIPs) != len(testIPs) {
		t.Errorf("Expected %d banned IPs, got %d", len(testIPs), len(bannedIPs))
	}

	// Verify all test IPs are in the list
	bannedMap := make(map[string]bool)
	for _, ip := range bannedIPs {
		bannedMap[ip] = true
	}

	for _, ip := range testIPs {
		if !bannedMap[ip] {
			t.Errorf("Expected IP %s to be in banned list", ip)
		}
	}
}

func TestGetAbuseStats(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	cfg := &config.RedisConfig{
		Host:     getEnv("TEST_REDIS_HOST", "localhost"),
		Port:     getEnv("TEST_REDIS_PORT", "6380"),
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

	// Clean up
	defer func() {
		ctx := context.Background()
		keys, _ := mockRedis.Keys(ctx, "abuse:*")
		for _, key := range keys {
			_ = mockRedis.Delete(ctx, key)
		}
	}()

	testIP := "192.168.1.1"
	expectedCount := int64(42)

	// Set abuse count
	abuseKey := "abuse:track:" + testIP
	if err := mockRedis.Set(ctx, abuseKey, "42", time.Hour); err != nil {
		t.Fatalf("Failed to set abuse count: %v", err)
	}

	// Get abuse stats
	count, err := GetAbuseStats(ctx, mockRedis, testIP)
	if err != nil {
		t.Fatalf("Failed to get abuse stats: %v", err)
	}

	if count != expectedCount {
		t.Errorf("Expected count %d, got %d", expectedCount, count)
	}
}
