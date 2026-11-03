package twitch

import (
	"context"
	"testing"
	"time"
)

func TestNewRateLimiter(t *testing.T) {
	maxTokens := 100
	rl := NewRateLimiter(maxTokens)

	if rl == nil {
		t.Fatal("NewRateLimiter returned nil")
	}

	if rl.maxTokens != maxTokens {
		t.Errorf("expected maxTokens=%d, got=%d", maxTokens, rl.maxTokens)
	}

	if rl.tokens != maxTokens {
		t.Errorf("expected initial tokens=%d, got=%d", maxTokens, rl.tokens)
	}
}

func TestRateLimiter_Available(t *testing.T) {
	rl := NewRateLimiter(10)

	available := rl.Available()
	if available != 10 {
		t.Errorf("expected 10 tokens available, got %d", available)
	}
}

func TestRateLimiter_TokenRefill(t *testing.T) {
	rl := NewRateLimiter(5)

	// Check initial tokens
	if rl.Available() != 5 {
		t.Errorf("expected 5 initial tokens, got %d", rl.Available())
	}

	// Consume one token
	ctx := context.Background()
	if err := rl.Wait(ctx); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should have 4 tokens left
	if rl.Available() != 4 {
		t.Errorf("expected 4 tokens after consuming one, got %d", rl.Available())
	}
}

func TestNewChannelRateLimiter(t *testing.T) {
	maxTokens := 100
	crl := NewChannelRateLimiter(maxTokens)

	if crl == nil {
		t.Fatal("NewChannelRateLimiter returned nil")
	}

	if crl.maxTokens != maxTokens {
		t.Errorf("expected maxTokens=%d, got=%d", maxTokens, crl.maxTokens)
	}

	if crl.limiters == nil {
		t.Error("expected limiters map to be initialized")
	}
}

func TestChannelRateLimiter_PerChannel(t *testing.T) {
	crl := NewChannelRateLimiter(10)
	ctx := context.Background()

	// Channel 1 should have independent tokens
	channel1 := "12345"
	channel2 := "67890"

	// Use 3 tokens from channel1
	for i := 0; i < 3; i++ {
		if err := crl.Wait(ctx, channel1); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	}

	// Channel1 should have 7 tokens left
	if available := crl.Available(channel1); available != 7 {
		t.Errorf("expected 7 tokens for channel1, got %d", available)
	}

	// Channel2 should still have all 10 tokens
	if available := crl.Available(channel2); available != 10 {
		t.Errorf("expected 10 tokens for channel2, got %d", available)
	}

	// Use 5 tokens from channel2
	for i := 0; i < 5; i++ {
		if err := crl.Wait(ctx, channel2); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	}

	// Channel1 should still have 7 tokens
	if available := crl.Available(channel1); available != 7 {
		t.Errorf("expected 7 tokens for channel1, got %d", available)
	}

	// Channel2 should have 5 tokens left
	if available := crl.Available(channel2); available != 5 {
		t.Errorf("expected 5 tokens for channel2, got %d", available)
	}
}

func TestChannelRateLimiter_ConcurrentAccess(t *testing.T) {
	crl := NewChannelRateLimiter(100)
	ctx := context.Background()
	channelID := "test-channel"

	// Create multiple goroutines accessing the same channel
	done := make(chan error, 10)
	for i := 0; i < 10; i++ {
		go func() {
			for j := 0; j < 5; j++ {
				if err := crl.Wait(ctx, channelID); err != nil {
					done <- err
					return
				}
			}
			done <- nil
		}()
	}

	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		if err := <-done; err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	}

	// Should have consumed 50 tokens total (10 goroutines * 5 tokens each)
	expected := 100 - 50
	if available := crl.Available(channelID); available != expected {
		t.Errorf("expected %d tokens remaining, got %d", expected, available)
	}
}

func TestChannelRateLimiter_CleanupInactive(t *testing.T) {
	crl := NewChannelRateLimiter(100)
	ctx := context.Background()

	// Create limiters for multiple channels
	channel1 := "channel1"
	channel2 := "channel2"
	channel3 := "channel3"

	// Access all channels (each uses 1 token)
	_ = crl.Wait(ctx, channel1)
	_ = crl.Wait(ctx, channel2)
	_ = crl.Wait(ctx, channel3)

	// Wait a bit, then access only channel1 (uses another token)
	time.Sleep(100 * time.Millisecond)
	_ = crl.Wait(ctx, channel1)

	// Cleanup channels inactive for more than 50ms
	// channel1 was just accessed, so should remain
	// channel2 and channel3 should be removed
	removed := crl.CleanupInactive(50 * time.Millisecond)

	if removed != 2 {
		t.Errorf("expected 2 limiters removed, got %d", removed)
	}

	// Verify channel1 still exists and has used 2 tokens
	// Note: getLimiter is called in Available which updates lastAccessed,
	// so we check that it has 98 tokens (100 - 2 from Wait calls)
	if available := crl.Available(channel1); available != 98 {
		t.Errorf("expected channel1 to have 98 tokens (used 2), got %d", available)
	}

	// channel2 and channel3 should have been recreated with full tokens
	if available := crl.Available(channel2); available != 100 {
		t.Errorf("expected channel2 to be recreated with 100 tokens, got %d", available)
	}
}

// TestRateLimiter_ConcurrentWaitersRespectCap checks that waiters woken by a
// refill each consume a token, so no more than maxTokens calls succeed per
// window no matter how many goroutines were blocked.
func TestRateLimiter_ConcurrentWaitersRespectCap(t *testing.T) {
	const (
		maxTokens = 2
		waiters   = 8
		window    = 150 * time.Millisecond
	)
	rl := newRateLimiterWithWindow(maxTokens, window)
	start := time.Now()

	// Drain the initial bucket.
	for i := 0; i < maxTokens; i++ {
		if err := rl.Wait(context.Background()); err != nil {
			t.Fatalf("initial wait: %v", err)
		}
	}

	// Only one refill (maxTokens tokens) fits before the deadline, at ~1 window.
	ctx, cancel := context.WithTimeout(context.Background(), window+window/2)
	defer cancel()

	results := make(chan error, waiters)
	for i := 0; i < waiters; i++ {
		go func() { results <- rl.Wait(ctx) }()
	}

	succeeded := 0
	for i := 0; i < waiters; i++ {
		if err := <-results; err == nil {
			succeeded++
		} else if err != context.DeadlineExceeded {
			t.Fatalf("unexpected error: %v", err)
		}
	}

	if succeeded != maxTokens {
		t.Fatalf("%d waiters succeeded after one refill, want %d", succeeded, maxTokens)
	}
	if elapsed := time.Since(start); elapsed < window {
		t.Fatalf("waiters returned after %v, before the %v refill", elapsed, window)
	}
}

func TestRateLimiter_WaitHonorsCancellation(t *testing.T) {
	rl := newRateLimiterWithWindow(1, time.Hour)
	if err := rl.Wait(context.Background()); err != nil {
		t.Fatalf("first wait: %v", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()
	if err := rl.Wait(ctx); err != context.DeadlineExceeded {
		t.Fatalf("Wait = %v, want deadline exceeded", err)
	}
}
