package middleware

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"sync"
	"time"

	redispkg "git.subcult.tv/subculture-collective/clpr/pkg/redis"
	"github.com/gin-gonic/gin"
	goredis "github.com/redis/go-redis/v9"
)

const publicResponseCachePrefix = "respcache:v1:"

// PublicResponseCacheConfig controls the anonymous response cache.
type PublicResponseCacheConfig struct {
	// TTL is how long a response is served without re-running the handler.
	TTL time.Duration
	// StaleTTL is how long the last good response is kept to answer requests
	// whose handler fails with a 5xx (for example a statement timeout).
	StaleTTL time.Duration
}

type cachedPublicResponse struct {
	Status      int    `json:"s"`
	ContentType string `json:"t"`
	Body        []byte `json:"b"`
}

// bufferedResponseWriter holds the handler's response so the cache can decide
// whether to send it or a stale copy instead.
type bufferedResponseWriter struct {
	gin.ResponseWriter
	body   bytes.Buffer
	status int
}

func (w *bufferedResponseWriter) WriteHeader(code int) { w.status = code }
func (w *bufferedResponseWriter) WriteHeaderNow()      {}
func (w *bufferedResponseWriter) Write(b []byte) (int, error) {
	if w.status == 0 {
		w.status = http.StatusOK
	}
	return w.body.Write(b)
}
func (w *bufferedResponseWriter) WriteString(s string) (int, error) { return w.Write([]byte(s)) }
func (w *bufferedResponseWriter) Status() int {
	if w.status == 0 {
		return http.StatusOK
	}
	return w.status
}
func (w *bufferedResponseWriter) Size() int     { return w.body.Len() }
func (w *bufferedResponseWriter) Written() bool { return w.body.Len() > 0 || w.status != 0 }

// inflightCalls lets concurrent misses for the same key share one handler run.
type inflightCalls struct {
	mu    sync.Mutex
	calls map[string]*inflightCall
}

type inflightCall struct {
	done chan struct{}
	resp *cachedPublicResponse
}

func (g *inflightCalls) join(key string) (*inflightCall, bool) {
	g.mu.Lock()
	defer g.mu.Unlock()
	if call, ok := g.calls[key]; ok {
		return call, false
	}
	call := &inflightCall{done: make(chan struct{})}
	g.calls[key] = call
	return call, true
}

func (g *inflightCalls) finish(key string, call *inflightCall, resp *cachedPublicResponse) {
	g.mu.Lock()
	delete(g.calls, key)
	g.mu.Unlock()
	call.resp = resp
	close(call.done)
}

func publicResponseCacheKey(c *gin.Context) string {
	raw := c.Request.URL.Path + "?" + c.Request.URL.Query().Encode()
	sum := sha256.Sum256([]byte(raw))
	return publicResponseCachePrefix + hex.EncodeToString(sum[:16])
}

func writeCachedPublicResponse(c *gin.Context, resp *cachedPublicResponse, state string) {
	c.Header("X-Cache", state)
	c.Data(resp.Status, resp.ContentType, resp.Body)
	c.Abort()
}

// PublicResponseCache caches successful anonymous GET responses in Redis for
// cfg.TTL. Concurrent misses in this process share one handler run, and when
// the handler fails with a 5xx the last good response (kept for
// cfg.StaleTTL) is served instead. Requests from signed-in users (user_id set
// by an earlier auth middleware, or a bearer token) bypass the cache, so
// personalised responses are never shared. Redis errors fall through to the
// handler.
func PublicResponseCache(redis *redispkg.Client, cfg PublicResponseCacheConfig) gin.HandlerFunc {
	if cfg.TTL <= 0 {
		cfg.TTL = 30 * time.Second
	}
	if cfg.StaleTTL < cfg.TTL {
		cfg.StaleTTL = cfg.TTL
	}
	inflight := &inflightCalls{calls: map[string]*inflightCall{}}

	return func(c *gin.Context) {
		if redis == nil || c.Request.Method != http.MethodGet {
			c.Next()
			return
		}
		if _, signedIn := c.Get("user_id"); signedIn || c.GetHeader("Authorization") != "" {
			c.Next()
			return
		}

		ctx := c.Request.Context()
		key := publicResponseCacheKey(c)
		if resp, ok := loadPublicResponse(ctx, redis, key); ok {
			writeCachedPublicResponse(c, resp, "HIT")
			return
		}

		call, leader := inflight.join(key)
		if !leader {
			select {
			case <-call.done:
				if call.resp != nil && call.resp.Status == http.StatusOK {
					writeCachedPublicResponse(c, call.resp, "SHARED")
					return
				}
			case <-ctx.Done():
				c.AbortWithStatus(http.StatusServiceUnavailable)
				return
			}
			// The shared run failed; run this request's own handler below
			// without joining another flight.
			runUncachedWithStale(c, redis, key)
			return
		}

		var resp *cachedPublicResponse
		// A panicking handler must still release waiting followers.
		defer func() {
			if resp == nil {
				inflight.finish(key, call, nil)
			}
		}()
		resp = runBuffered(c)
		inflight.finish(key, call, resp)
		if resp.Status == http.StatusOK {
			storePublicResponse(redis, key, resp, cfg)
			writeRawResponse(c, resp, "MISS")
			return
		}
		if resp.Status >= 500 {
			if stale, ok := loadPublicResponse(context.Background(), redis, key+":stale"); ok {
				log.Printf("Serving stale cached response for %s after status %d", c.Request.URL.Path, resp.Status)
				writeRawResponse(c, stale, "STALE")
				return
			}
		}
		writeRawResponse(c, resp, "BYPASS")
	}
}

func runUncachedWithStale(c *gin.Context, redis *redispkg.Client, key string) {
	resp := runBuffered(c)
	if resp.Status >= 500 {
		if stale, ok := loadPublicResponse(context.Background(), redis, key+":stale"); ok {
			writeRawResponse(c, stale, "STALE")
			return
		}
	}
	writeRawResponse(c, resp, "BYPASS")
}

// runBuffered runs the remaining handlers with a buffering writer.
func runBuffered(c *gin.Context) *cachedPublicResponse {
	original := c.Writer
	buffered := &bufferedResponseWriter{ResponseWriter: original}
	c.Writer = buffered
	// Restore on panic too, so recovery middleware writes to the client.
	defer func() { c.Writer = original }()
	c.Next()
	return &cachedPublicResponse{
		Status:      buffered.Status(),
		ContentType: original.Header().Get("Content-Type"),
		Body:        buffered.body.Bytes(),
	}
}

func writeRawResponse(c *gin.Context, resp *cachedPublicResponse, state string) {
	c.Header("X-Cache", state)
	if resp.ContentType != "" {
		c.Header("Content-Type", resp.ContentType)
	}
	c.Writer.WriteHeader(resp.Status)
	_, _ = c.Writer.Write(resp.Body)
}

func loadPublicResponse(ctx context.Context, redis *redispkg.Client, key string) (*cachedPublicResponse, bool) {
	raw, err := redis.Get(ctx, key)
	if err != nil {
		if !errors.Is(err, goredis.Nil) {
			log.Printf("Response cache read failed: %v", err)
		}
		return nil, false
	}
	var resp cachedPublicResponse
	if err := json.Unmarshal([]byte(raw), &resp); err != nil || resp.Status == 0 {
		return nil, false
	}
	return &resp, true
}

func storePublicResponse(redis *redispkg.Client, key string, resp *cachedPublicResponse, cfg PublicResponseCacheConfig) {
	encoded, err := json.Marshal(resp)
	if err != nil {
		return
	}
	// Store even if the client has gone away; the next visitor benefits.
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if err := redis.Set(ctx, key, encoded, cfg.TTL); err != nil {
		log.Printf("Response cache write failed: %v", err)
		return
	}
	_ = redis.Set(ctx, key+":stale", encoded, cfg.StaleTTL)
}
