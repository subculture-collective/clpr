package services

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"git.subcult.tv/subculture-collective/clpr/config"
	redispkg "git.subcult.tv/subculture-collective/clpr/pkg/redis"
)

// TestDistributedRateLimiter tests the Redis-backed distributed rate limiter
func TestDistributedRateLimiter(t *testing.T) {
	// Skip if Redis is not available
	cfg := &config.RedisConfig{
		Host:     getTestEnv("TEST_REDIS_HOST", "localhost"),
		Port:     getTestEnv("TEST_REDIS_PORT", "6380"),
		Password: "",
		DB:       1, // Use test DB
	}

	redisClient, err := redispkg.NewClient(cfg)
	if err != nil {
		t.Skip("Redis not available for testing:", err)
		return
	}
	defer redisClient.Close()

	ctx := context.Background()

	t.Run("allows requests within limit", func(t *testing.T) {
		limiter := NewDistributedRateLimiter(redisClient, 5, time.Minute)
		key := "test:allows_within_limit"

		// Should allow first 5 requests
		for i := 0; i < 5; i++ {
			allowed, err := limiter.Allow(ctx, key)
			require.NoError(t, err)
			assert.True(t, allowed, "Request %d should be allowed", i+1)
		}

		// 6th request should be denied
		allowed, err := limiter.Allow(ctx, key)
		require.NoError(t, err)
		assert.False(t, allowed, "Request 6 should be denied")

		// Cleanup
		_ = redisClient.Delete(ctx, "ratelimit:"+key)
	})

	t.Run("respects time window", func(t *testing.T) {
		limiter := NewDistributedRateLimiter(redisClient, 2, 2*time.Second)
		key := "test:respects_window"

		// Use up the limit
		allowed, err := limiter.Allow(ctx, key)
		require.NoError(t, err)
		assert.True(t, allowed)

		allowed, err = limiter.Allow(ctx, key)
		require.NoError(t, err)
		assert.True(t, allowed)

		// Should be denied
		allowed, err = limiter.Allow(ctx, key)
		require.NoError(t, err)
		assert.False(t, allowed)

		// Wait for window to expire
		time.Sleep(2100 * time.Millisecond)

		// Should be allowed again
		allowed, err = limiter.Allow(ctx, key)
		require.NoError(t, err)
		assert.True(t, allowed, "Should be allowed after window expires")

		// Cleanup
		_ = redisClient.Delete(ctx, "ratelimit:"+key)
	})

	t.Run("works across multiple keys", func(t *testing.T) {
		limiter := NewDistributedRateLimiter(redisClient, 3, time.Minute)
		key1 := "test:multi_key_1"
		key2 := "test:multi_key_2"

		// Use up limit for key1
		for i := 0; i < 3; i++ {
			allowed, err := limiter.Allow(ctx, key1)
			require.NoError(t, err)
			assert.True(t, allowed)
		}

		// key1 should be denied
		allowed, err := limiter.Allow(ctx, key1)
		require.NoError(t, err)
		assert.False(t, allowed)

		// key2 should still be allowed (different key)
		allowed, err = limiter.Allow(ctx, key2)
		require.NoError(t, err)
		assert.True(t, allowed, "Different key should have separate limit")

		// Cleanup
		_ = redisClient.Delete(ctx, "ratelimit:"+key1)
		_ = redisClient.Delete(ctx, "ratelimit:"+key2)
	})

	t.Run("sliding window behavior", func(t *testing.T) {
		limiter := NewDistributedRateLimiter(redisClient, 3, 3*time.Second)
		key := "test:sliding_window"

		// Make 3 requests at t=0
		for i := 0; i < 3; i++ {
			allowed, err := limiter.Allow(ctx, key)
			require.NoError(t, err)
			assert.True(t, allowed)
		}

		// Should be denied at t=0
		allowed, err := limiter.Allow(ctx, key)
		require.NoError(t, err)
		assert.False(t, allowed)

		// Wait 1.5 seconds (still within window)
		time.Sleep(1500 * time.Millisecond)

		// Should still be denied (requests from t=0 still in window)
		allowed, err = limiter.Allow(ctx, key)
		require.NoError(t, err)
		assert.False(t, allowed)

		// Wait another 2 seconds (t=3.5s, original requests expired)
		time.Sleep(2000 * time.Millisecond)

		// Should be allowed now (original requests outside window)
		allowed, err = limiter.Allow(ctx, key)
		require.NoError(t, err)
		assert.True(t, allowed, "Should be allowed after sliding window moves")

		// Cleanup
		_ = redisClient.Delete(ctx, "ratelimit:"+key)
	})

	t.Run("concurrent requests respect rate limit (no race condition)", func(t *testing.T) {
		limiter := NewDistributedRateLimiter(redisClient, 10, time.Minute)
		key := "test:concurrent"

		// Run 50 concurrent goroutines trying to make requests
		// Only 10 should succeed due to the rate limit
		numGoroutines := 50
		successChan := make(chan bool, numGoroutines)

		var wg sync.WaitGroup
		wg.Add(numGoroutines)

		for i := 0; i < numGoroutines; i++ {
			go func() {
				defer wg.Done()
				allowed, err := limiter.Allow(ctx, key)
				require.NoError(t, err)
				successChan <- allowed
			}()
		}

		wg.Wait()
		close(successChan)

		// Count successful requests
		successCount := 0
		for allowed := range successChan {
			if allowed {
				successCount++
			}
		}

		// Exactly 10 requests should have succeeded (the limit)
		assert.Equal(t, 10, successCount, "Exactly 10 concurrent requests should be allowed")

		// Cleanup
		_ = redisClient.Delete(ctx, "ratelimit:"+key)
	})
}

// TestInMemoryRateLimiterAdapter tests the fallback in-memory adapter
func TestInMemoryRateLimiterAdapter(t *testing.T) {
	ctx := context.Background()

	t.Run("adapter works with SimpleRateLimiter", func(t *testing.T) {
		adapter := NewInMemoryRateLimiterAdapter(3, time.Minute)
		key := "test:adapter"

		// Should allow first 3 requests
		for i := 0; i < 3; i++ {
			allowed, err := adapter.Allow(ctx, key)
			require.NoError(t, err)
			assert.True(t, allowed)
		}

		// 4th request should be denied
		allowed, err := adapter.Allow(ctx, key)
		require.NoError(t, err)
		assert.False(t, allowed)
	})
}

// BenchmarkDistributedRateLimiter benchmarks the rate limiter performance
func BenchmarkDistributedRateLimiter(b *testing.B) {
	cfg := &config.RedisConfig{
		Host:     getTestEnv("TEST_REDIS_HOST", "localhost"),
		Port:     getTestEnv("TEST_REDIS_PORT", "6380"),
		Password: "",
		DB:       1,
	}

	redisClient, err := redispkg.NewClient(cfg)
	if err != nil {
		b.Skip("Redis not available for benchmarking:", err)
		return
	}
	defer redisClient.Close()

	limiter := NewDistributedRateLimiter(redisClient, 1000, time.Minute)
	ctx := context.Background()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			key := "bench:key"
			_, _ = limiter.Allow(ctx, key)
			i++
		}
	})

	// Cleanup
	_ = redisClient.Delete(ctx, "ratelimit:bench:key")
}
