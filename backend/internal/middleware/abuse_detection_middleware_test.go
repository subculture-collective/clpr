package middleware

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"sync"
	"testing"
	"time"

	"git.subcult.tv/subculture-collective/clpr/config"
	redispkg "git.subcult.tv/subculture-collective/clpr/pkg/redis"
	"github.com/gin-gonic/gin"
)

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

// newAbuseTestRedis connects to the test Redis and removes every abuse key
// for the given IPs before and after the test.
func newAbuseTestRedis(t *testing.T, ips ...string) *redispkg.Client {
	t.Helper()
	if testing.Short() {
		t.Skip("Skipping integration test")
	}
	cfg := &config.RedisConfig{
		Host: getEnv("TEST_REDIS_HOST", "localhost"),
		Port: getEnv("TEST_REDIS_PORT", "6380"),
	}
	client, err := redispkg.NewClient(cfg)
	if err != nil {
		t.Skip("Redis not available, skipping integration test")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if err := client.HealthCheck(ctx); err != nil {
		t.Skip("Redis not available, skipping integration test")
	}
	cleanup := func() {
		ctx := context.Background()
		for _, ip := range ips {
			for _, pattern := range []string{abuseBanPrefix + ip, abuseOffensePrefix + ip, abuseCountPrefix + "*:" + ip + ":*"} {
				keys, _ := client.Keys(ctx, pattern)
				for _, key := range keys {
					_ = client.Delete(ctx, key)
				}
			}
		}
	}
	cleanup()
	t.Cleanup(func() {
		cleanup()
		_ = client.Close()
	})
	return client
}

func newAbuseTestRouter(redis *redispkg.Client, cfg config.AbuseDetectionConfig) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(AbuseDetectionMiddleware(redis, cfg))
	ok := func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"message": "ok"}) }
	r.GET("/api/v1/clips/:id/tags", ok)
	r.POST("/api/v1/clips/:id/vote", ok)
	r.POST("/api/v1/clips/:id/track-view", ok)
	r.GET("/api/v1/auth/twitch", ok)
	r.GET("/health", ok)
	return r
}

func doAbuseRequest(r http.Handler, method, path, ip string) *httptest.ResponseRecorder {
	w := httptest.NewRecorder()
	req := httptest.NewRequest(method, path, nil)
	req.RemoteAddr = ip + ":12345"
	r.ServeHTTP(w, req)
	return w
}

func testAbuseConfig() config.AbuseDetectionConfig {
	cfg := DefaultAbuseDetectionConfig()
	cfg.ReadPerMinute = 10
	cfg.ReadPerHour = 10
	cfg.WritePerMinute = 3
	cfg.WritePerHour = 3
	return cfg
}

func TestAbuseDetectionMiddleware_NilRedisPassesThrough(t *testing.T) {
	r := newAbuseTestRouter(nil, DefaultAbuseDetectionConfig())
	if w := doAbuseRequest(r, http.MethodGet, "/api/v1/clips/1/tags", "198.51.100.1"); w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestDefaultAbuseDetectionConfigAllowsBrowsingBursts(t *testing.T) {
	cfg := DefaultAbuseDetectionConfig()
	// A home page load issues roughly 150 API reads; the old limit was 1000/hour.
	if cfg.ReadPerMinute < 600 || cfg.ReadPerHour < 5000 {
		t.Fatalf("read limits too low for normal browsing: %+v", cfg)
	}
	if cfg.WritePerMinute >= cfg.ReadPerMinute || cfg.WritePerHour >= cfg.ReadPerHour {
		t.Fatalf("writes must be limited more strictly than reads: %+v", cfg)
	}
	if cfg.BanDurations[0] > 15*time.Minute {
		t.Fatalf("first offense ban should be short, got %v", cfg.BanDurations[0])
	}
}

func TestAbuseBanDurationEscalates(t *testing.T) {
	durations := []time.Duration{15 * time.Minute, time.Hour, 24 * time.Hour}
	cases := map[int64]time.Duration{0: 15 * time.Minute, 1: 15 * time.Minute, 2: time.Hour, 3: 24 * time.Hour, 9: 24 * time.Hour}
	for offense, want := range cases {
		if got := abuseBanDuration(durations, offense); got != want {
			t.Errorf("offense %d: got %v, want %v", offense, got, want)
		}
	}
	if got := abuseBanDuration(nil, 1); got != config.DefaultAbuseBanDurations[0] {
		t.Errorf("empty durations should use defaults, got %v", got)
	}
}

func TestClassifyAbuseTraffic(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	got := map[string]abuseTrafficClass{}
	record := func(c *gin.Context) { got[c.Request.Method+" "+c.Request.URL.Path] = classifyAbuseTraffic(c) }
	r.GET("/api/v1/clips/:id/tags", record)
	r.GET("/api/v1/auth/me", record)
	r.GET("/api/v1/auth/twitch", record)
	r.POST("/api/v1/auth/refresh", record)
	r.POST("/api/v1/clips/:id/track-view", record)
	r.POST("/api/v1/events", record)
	r.POST("/api/v1/clips/:id/vote", record)
	r.DELETE("/api/v1/clips/:id/favorite", record)

	want := map[string]abuseTrafficClass{
		"GET /api/v1/clips/1/tags":        abuseClassRead,
		"GET /api/v1/auth/me":             abuseClassRead,
		"GET /api/v1/auth/twitch":         abuseClassWrite,
		"POST /api/v1/auth/refresh":       abuseClassWrite,
		"POST /api/v1/clips/1/track-view": abuseClassRead,
		"POST /api/v1/events":             abuseClassRead,
		"POST /api/v1/clips/1/vote":       abuseClassWrite,
		"DELETE /api/v1/clips/1/favorite": abuseClassWrite,
	}
	for key := range want {
		var method, path string
		_, _ = fmt.Sscanf(key, "%s %s", &method, &path)
		r.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(method, path, nil))
	}
	for key, class := range want {
		if got[key] != class {
			t.Errorf("%s classified %q, want %q", key, got[key], class)
		}
	}
}

func TestAbuseDetectionMiddleware_ReadBurstGetsShortBan(t *testing.T) {
	ip := "198.51.100.10"
	redis := newAbuseTestRedis(t, ip)
	r := newAbuseTestRouter(redis, testAbuseConfig())

	for i := 0; i < 10; i++ {
		if w := doAbuseRequest(r, http.MethodGet, "/api/v1/clips/1/tags", ip); w.Code != http.StatusOK {
			t.Fatalf("read %d: expected 200, got %d", i, w.Code)
		}
	}
	w := doAbuseRequest(r, http.MethodGet, "/api/v1/clips/1/tags", ip)
	if w.Code != http.StatusTooManyRequests {
		t.Fatalf("expected 429 after read limit, got %d", w.Code)
	}
	retry, _ := strconv.Atoi(w.Header().Get("Retry-After"))
	if retry <= 0 || retry > int((15*time.Minute).Seconds()) {
		t.Fatalf("first offense Retry-After = %q, want within 15 minutes", w.Header().Get("Retry-After"))
	}
	if w := doAbuseRequest(r, http.MethodGet, "/api/v1/clips/2/tags", ip); w.Code != http.StatusTooManyRequests {
		t.Fatalf("expected banned IP to stay blocked, got %d", w.Code)
	}
	// Health probes stay exempt.
	if w := doAbuseRequest(r, http.MethodGet, "/health", ip); w.Code != http.StatusOK {
		t.Fatalf("health must be exempt, got %d", w.Code)
	}
}

func TestAbuseDetectionMiddleware_WritesHaveSeparateStricterBudget(t *testing.T) {
	ip := "198.51.100.11"
	redis := newAbuseTestRedis(t, ip)
	r := newAbuseTestRouter(redis, testAbuseConfig())

	// Reads and per-view telemetry do not consume the write budget.
	for i := 0; i < 5; i++ {
		if w := doAbuseRequest(r, http.MethodGet, "/api/v1/clips/1/tags", ip); w.Code != http.StatusOK {
			t.Fatalf("read %d: got %d", i, w.Code)
		}
		if w := doAbuseRequest(r, http.MethodPost, "/api/v1/clips/1/track-view", ip); w.Code != http.StatusOK {
			t.Fatalf("telemetry %d: got %d", i, w.Code)
		}
	}
	for i := 0; i < 3; i++ {
		if w := doAbuseRequest(r, http.MethodPost, "/api/v1/clips/1/vote", ip); w.Code != http.StatusOK {
			t.Fatalf("write %d: expected 200, got %d", i, w.Code)
		}
	}
	if w := doAbuseRequest(r, http.MethodGet, "/api/v1/auth/twitch", ip); w.Code != http.StatusTooManyRequests {
		t.Fatalf("auth endpoint should count against the write budget, got %d", w.Code)
	}
}

func TestAbuseDetectionMiddleware_EscalatesRepeatOffenses(t *testing.T) {
	ip := "198.51.100.12"
	redis := newAbuseTestRedis(t, ip)
	r := newAbuseTestRouter(redis, testAbuseConfig())
	ctx := context.Background()

	wants := []time.Duration{15 * time.Minute, time.Hour, 6 * time.Hour, 24 * time.Hour, 24 * time.Hour}
	for n, want := range wants {
		var last *httptest.ResponseRecorder
		for i := 0; i < 4; i++ {
			last = doAbuseRequest(r, http.MethodPost, "/api/v1/clips/1/vote", ip)
		}
		if last.Code != http.StatusTooManyRequests {
			t.Fatalf("offense %d: expected 429, got %d", n+1, last.Code)
		}
		_, ttl, err := redis.GetWithTTL(ctx, abuseBanPrefix+ip)
		if err != nil {
			t.Fatalf("offense %d: ban key missing: %v", n+1, err)
		}
		if ttl > want || ttl < want-time.Minute {
			t.Fatalf("offense %d: ban ttl %v, want about %v", n+1, ttl, want)
		}
		// Simulate the ban expiring. Counters were reset when the ban was set,
		// so the next request after expiry is allowed.
		_ = redis.Delete(ctx, abuseBanPrefix+ip)
		if w := doAbuseRequest(r, http.MethodPost, "/api/v1/clips/1/vote", ip); w.Code != http.StatusOK {
			t.Fatalf("offense %d: first request after ban expiry got %d", n+1, w.Code)
		}
	}
}

func TestAbuseDetectionMiddleware_ParallelBurstIsOneOffense(t *testing.T) {
	ip := "198.51.100.13"
	redis := newAbuseTestRedis(t, ip)
	r := newAbuseTestRouter(redis, testAbuseConfig())

	var wg sync.WaitGroup
	for i := 0; i < 40; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			doAbuseRequest(r, http.MethodGet, "/api/v1/clips/1/tags", ip)
		}()
	}
	wg.Wait()

	val, err := redis.Get(context.Background(), abuseOffensePrefix+ip)
	if err != nil {
		t.Fatalf("offense counter missing: %v", err)
	}
	if val != "1" {
		t.Fatalf("parallel burst recorded %s offenses, want 1", val)
	}
}

func TestAbuseDetectionMiddleware_DisabledAndWhitelisted(t *testing.T) {
	ip := "198.51.100.14"
	redis := newAbuseTestRedis(t, ip)

	cfg := testAbuseConfig()
	cfg.Enabled = false
	r := newAbuseTestRouter(redis, cfg)
	for i := 0; i < 20; i++ {
		if w := doAbuseRequest(r, http.MethodGet, "/api/v1/clips/1/tags", ip); w.Code != http.StatusOK {
			t.Fatalf("disabled limiter blocked request %d: %d", i, w.Code)
		}
	}

	InitRateLimitWhitelist(ip)
	t.Cleanup(func() { InitRateLimitWhitelist("") })
	r = newAbuseTestRouter(redis, testAbuseConfig())
	for i := 0; i < 20; i++ {
		if w := doAbuseRequest(r, http.MethodGet, "/api/v1/clips/1/tags", ip); w.Code != http.StatusOK {
			t.Fatalf("whitelisted IP blocked on request %d: %d", i, w.Code)
		}
	}
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
	testIP := "198.51.100.20"
	redis := newAbuseTestRedis(t, testIP)
	ctx := context.Background()

	if err := redis.Set(ctx, abuseBanPrefix+testIP, "1", time.Hour); err != nil {
		t.Fatalf("Failed to set ban: %v", err)
	}
	if err := redis.Set(ctx, abuseOffensePrefix+testIP, "3", time.Hour); err != nil {
		t.Fatalf("Failed to set offenses: %v", err)
	}

	if err := UnbanIP(ctx, redis, testIP); err != nil {
		t.Fatalf("Failed to unban IP: %v", err)
	}

	for _, key := range []string{abuseBanPrefix + testIP, abuseOffensePrefix + testIP} {
		exists, err := redis.Exists(ctx, key)
		if err != nil {
			t.Fatalf("Failed to check %s: %v", key, err)
		}
		if exists {
			t.Errorf("%s was not removed", key)
		}
	}
}

func TestGetBannedIPs(t *testing.T) {
	testIPs := []string{"198.51.100.21", "198.51.100.22", "2001:db8::21"}
	redis := newAbuseTestRedis(t, testIPs...)
	ctx := context.Background()

	for _, ip := range testIPs {
		if err := redis.Set(ctx, abuseBanPrefix+ip, "1", time.Hour); err != nil {
			t.Fatalf("Failed to ban IP %s: %v", ip, err)
		}
	}

	bannedIPs, err := GetBannedIPs(ctx, redis)
	if err != nil {
		t.Fatalf("Failed to get banned IPs: %v", err)
	}

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
	testIP := "198.51.100.23"
	redis := newAbuseTestRedis(t, testIP)
	ctx := context.Background()

	bucket := time.Now().Unix() / int64(time.Hour/time.Second)
	if err := redis.Set(ctx, abuseCountKey(abuseClassRead, "h", testIP, bucket), "40", time.Hour); err != nil {
		t.Fatalf("Failed to set read count: %v", err)
	}
	if err := redis.Set(ctx, abuseCountKey(abuseClassWrite, "h", testIP, bucket), "2", time.Hour); err != nil {
		t.Fatalf("Failed to set write count: %v", err)
	}

	count, err := GetAbuseStats(ctx, redis, testIP)
	if err != nil {
		t.Fatalf("Failed to get abuse stats: %v", err)
	}
	if count != 42 {
		t.Errorf("Expected count 42, got %d", count)
	}
}
