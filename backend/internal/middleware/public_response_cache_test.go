package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	redispkg "git.subcult.tv/subculture-collective/clpr/pkg/redis"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func newPublicCacheTestRedis(t *testing.T) *redispkg.Client {
	t.Helper()
	redis := newAbuseTestRedis(t)
	t.Cleanup(func() {
		keys, _ := redis.Keys(context.Background(), publicResponseCachePrefix+"*")
		for _, key := range keys {
			_ = redis.Delete(context.Background(), key)
		}
	})
	return redis
}

type publicCacheHarness struct {
	router *gin.Engine
	path   string
	calls  atomic.Int32
	status atomic.Int32
	delay  time.Duration
	user   bool
}

func newPublicCacheHarness(redis *redispkg.Client) *publicCacheHarness {
	gin.SetMode(gin.TestMode)
	h := &publicCacheHarness{path: "/api/v1/test-" + uuid.NewString()[:8] + "/clips"}
	h.status.Store(http.StatusOK)
	h.router = gin.New()
	h.router.Use(JSONRecoveryMiddleware())
	h.router.GET(h.path, func(c *gin.Context) {
		if h.user {
			c.Set("user_id", uuid.New())
		}
		c.Next()
	}, PublicResponseCache(redis, PublicResponseCacheConfig{TTL: time.Minute, StaleTTL: time.Hour}), func(c *gin.Context) {
		n := h.calls.Add(1)
		if h.delay > 0 {
			time.Sleep(h.delay)
		}
		status := int(h.status.Load())
		if status == 999 {
			panic("handler exploded")
		}
		c.JSON(status, gin.H{"call": n, "q": c.Query("sort")})
	})
	return h
}

func (h *publicCacheHarness) get(query string) *httptest.ResponseRecorder {
	rec := httptest.NewRecorder()
	h.router.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, h.path+query, nil))
	return rec
}

func TestPublicResponseCacheServesAnonymousHits(t *testing.T) {
	h := newPublicCacheHarness(newPublicCacheTestRedis(t))

	first := h.get("?sort=hot&limit=20")
	if first.Code != http.StatusOK || first.Header().Get("X-Cache") != "MISS" {
		t.Fatalf("first request: %d %q", first.Code, first.Header().Get("X-Cache"))
	}
	// Same parameters in a different order hit the same entry.
	second := h.get("?limit=20&sort=hot")
	if second.Header().Get("X-Cache") != "HIT" || second.Body.String() != first.Body.String() {
		t.Fatalf("second request: %q %s (first %s)", second.Header().Get("X-Cache"), second.Body.String(), first.Body.String())
	}
	if ct := second.Header().Get("Content-Type"); ct != "application/json; charset=utf-8" {
		t.Fatalf("cached Content-Type = %q", ct)
	}
	if h.calls.Load() != 1 {
		t.Fatalf("handler ran %d times, want 1", h.calls.Load())
	}
	if other := h.get("?sort=new&limit=20"); other.Header().Get("X-Cache") != "MISS" {
		t.Fatalf("different query should miss, got %q", other.Header().Get("X-Cache"))
	}
}

func TestPublicResponseCacheBypassesSignedInUsers(t *testing.T) {
	h := newPublicCacheHarness(newPublicCacheTestRedis(t))
	h.user = true
	h.get("?sort=hot")
	h.get("?sort=hot")
	if h.calls.Load() != 2 {
		t.Fatalf("signed-in requests ran handler %d times, want 2", h.calls.Load())
	}

	h.user = false
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, h.path+"?sort=hot", nil)
	req.Header.Set("Authorization", "Bearer token")
	h.router.ServeHTTP(rec, req)
	if h.calls.Load() != 3 {
		t.Fatalf("bearer request was served from cache")
	}
}

func TestPublicResponseCacheDoesNotStoreErrors(t *testing.T) {
	h := newPublicCacheHarness(newPublicCacheTestRedis(t))
	h.status.Store(http.StatusBadRequest)
	h.get("?sort=bad")
	if rec := h.get("?sort=bad"); rec.Code != http.StatusBadRequest || h.calls.Load() != 2 {
		t.Fatalf("400 was cached: code %d calls %d", rec.Code, h.calls.Load())
	}
	h.status.Store(http.StatusInternalServerError)
	if rec := h.get("?sort=err"); rec.Code != http.StatusInternalServerError {
		t.Fatalf("500 without a stale copy should pass through, got %d", rec.Code)
	}
}

func TestPublicResponseCacheServesStaleOnServerError(t *testing.T) {
	redis := newPublicCacheTestRedis(t)
	h := newPublicCacheHarness(redis)
	good := h.get("?sort=hot")

	// Expire the fresh copy, then make the database fail.
	req := httptest.NewRequest(http.MethodGet, h.path+"?sort=hot", nil)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = req
	_ = redis.Delete(context.Background(), publicResponseCacheKey(c))
	h.status.Store(http.StatusInternalServerError)

	rec := h.get("?sort=hot")
	if rec.Code != http.StatusOK || rec.Header().Get("X-Cache") != "STALE" || rec.Body.String() != good.Body.String() {
		t.Fatalf("expected stale 200, got %d %q %s", rec.Code, rec.Header().Get("X-Cache"), rec.Body.String())
	}
}

func TestPublicResponseCacheCollapsesConcurrentMisses(t *testing.T) {
	h := newPublicCacheHarness(newPublicCacheTestRedis(t))
	h.delay = 100 * time.Millisecond

	var wg sync.WaitGroup
	codes := make([]int, 20)
	for i := range codes {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			codes[i] = h.get("?sort=trending").Code
		}(i)
	}
	wg.Wait()

	for i, code := range codes {
		if code != http.StatusOK {
			t.Fatalf("request %d: status %d", i, code)
		}
	}
	if calls := h.calls.Load(); calls != 1 {
		t.Fatalf("20 concurrent misses ran the handler %d times, want 1", calls)
	}
}

func TestPublicResponseCachePanicStillReachesRecovery(t *testing.T) {
	h := newPublicCacheHarness(newPublicCacheTestRedis(t))
	h.status.Store(999)
	if rec := h.get("?sort=panic"); rec.Code != http.StatusInternalServerError || rec.Body.Len() == 0 {
		t.Fatalf("panic response = %d %q", rec.Code, rec.Body.String())
	}
}

func TestPublicResponseCacheNilRedisPassesThrough(t *testing.T) {
	h := newPublicCacheHarness(nil)
	h.get("?sort=hot")
	h.get("?sort=hot")
	if h.calls.Load() != 2 {
		t.Fatalf("nil redis should not cache; handler ran %d times", h.calls.Load())
	}
}
