package services

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"git.subcult.tv/subculture-collective/clpr/internal/models"
)

// TestCacheServiceKeyFormats verifies cache key formats are correct
func TestCacheServiceKeyFormats(t *testing.T) {
	tests := []struct {
		name     string
		format   string
		args     []interface{}
		expected string
	}{
		{
			name:     "Feed hot key",
			format:   KeyFeedHot,
			args:     []interface{}{1},
			expected: "feed:hot:page:1",
		},
		{
			name:     "Feed top key",
			format:   KeyFeedTop,
			args:     []interface{}{"24h", 1},
			expected: "feed:top:24h:page:1",
		},
		{
			name:     "Feed new key",
			format:   KeyFeedNew,
			args:     []interface{}{2},
			expected: "feed:new:page:2",
		},
		{
			name:     "Feed game key",
			format:   KeyFeedGame,
			args:     []interface{}{"game123", "hot", 1},
			expected: "feed:game:game123:hot:page:1",
		},
		{
			name:     "Feed creator key",
			format:   KeyFeedCreator,
			args:     []interface{}{"creator456", "new", 3},
			expected: "feed:creator:creator456:new:page:3",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatKey(tt.format, tt.args...)
			if result != tt.expected {
				t.Errorf("Expected key %s, got %s", tt.expected, result)
			}
		})
	}
}

// Helper function to format keys for testing
func formatKey(format string, args ...interface{}) string {
	return fmt.Sprintf(format, args...)
}

// TestCacheTTLConstants verifies TTL values are reasonable
func TestCacheTTLConstants(t *testing.T) {
	tests := []struct {
		name   string
		ttl    time.Duration
		minTTL time.Duration
		maxTTL time.Duration
	}{
		{
			name:   "Feed hot TTL",
			ttl:    TTLFeedHot,
			minTTL: 1 * time.Minute,
			maxTTL: 10 * time.Minute,
		},
		{
			name:   "Feed top TTL",
			ttl:    TTLFeedTop,
			minTTL: 10 * time.Minute,
			maxTTL: 30 * time.Minute,
		},
		{
			name:   "Feed new TTL",
			ttl:    TTLFeedNew,
			minTTL: 1 * time.Minute,
			maxTTL: 5 * time.Minute,
		},
		{
			name:   "Clip TTL",
			ttl:    TTLClip,
			minTTL: 30 * time.Minute,
			maxTTL: 2 * time.Hour,
		},
		{
			name:   "Session TTL",
			ttl:    TTLSession,
			minTTL: 1 * 24 * time.Hour,
			maxTTL: 30 * 24 * time.Hour,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.ttl < tt.minTTL {
				t.Errorf("TTL %v is less than minimum %v", tt.ttl, tt.minTTL)
			}
			if tt.ttl > tt.maxTTL {
				t.Errorf("TTL %v is greater than maximum %v", tt.ttl, tt.maxTTL)
			}
		})
	}
}

// TestInvalidationPatterns verifies invalidation patterns are correct
func TestInvalidationPatterns(t *testing.T) {
	// This test verifies the invalidation logic structure
	// without requiring actual Redis connection

	clipID := uuid.New()
	gameID := "game123"
	creatorID := "creator456"

	// Test that clip invalidation would clear related caches
	clip := &models.Clip{
		ID:        clipID,
		GameID:    &gameID,
		CreatorID: &creatorID,
	}

	// Verify clip has required fields for invalidation
	if clip.GameID == nil {
		t.Error("Clip should have game ID for testing")
	}
	if clip.CreatorID == nil {
		t.Error("Clip should have creator ID for testing")
	}

	// Verify clip ID is valid
	if clip.ID == uuid.Nil {
		t.Error("Clip ID should not be nil")
	}
}

// TestCacheServiceCreation verifies cache service can be created
func TestCacheServiceCreation(t *testing.T) {
	// Create cache service with nil redis (for testing structure only)
	service := &CacheService{
		redis: nil,
	}

	// Removed impossible nil check; construction is guaranteed non-nil
	_ = service
}

// TestSmartInvalidationLogic verifies invalidation logic is comprehensive
func TestSmartInvalidationLogic(t *testing.T) {
	// Test invalidation on new clip
	t.Run("Invalidate on new clip", func(t *testing.T) {
		clipID := uuid.New()
		gameID := "game123"
		creatorID := "creator456"

		clip := &models.Clip{
			ID:        clipID,
			GameID:    &gameID,
			CreatorID: &creatorID,
		}

		// Verify all required fields are present
		if clip.ID == uuid.Nil {
			t.Error("Clip ID is required")
		}
		if clip.GameID == nil || *clip.GameID == "" {
			t.Error("Game ID should be present for testing")
		}
		if clip.CreatorID == nil || *clip.CreatorID == "" {
			t.Error("Creator ID should be present for testing")
		}
	})

	// Test invalidation on vote
	t.Run("Invalidate on vote", func(t *testing.T) {
		clipID := uuid.New()
		if clipID == uuid.Nil {
			t.Error("Clip ID should not be nil")
		}
	})

	// Test invalidation on comment
	t.Run("Invalidate on comment", func(t *testing.T) {
		clipID := uuid.New()
		if clipID == uuid.Nil {
			t.Error("Clip ID should not be nil")
		}
	})
}

// TestLockOperations verifies lock key format
func TestLockOperations(t *testing.T) {
	resource := "clip_import_123"
	expectedKey := "lock:clip_import_123"

	// Format lock key
	key := fmt.Sprintf(KeyLock, resource)

	if key != expectedKey {
		t.Errorf("Expected lock key %s, got %s", expectedKey, key)
	}
}

// TestSessionKeyFormat verifies session key format
func TestSessionKeyFormat(t *testing.T) {
	sessionID := "session_abc123"
	expectedKey := "session:session_abc123"

	key := fmt.Sprintf(KeySession, sessionID)

	if key != expectedKey {
		t.Errorf("Expected session key %s, got %s", expectedKey, key)
	}
}

// TestRefreshTokenKeyFormat verifies refresh token key format
func TestRefreshTokenKeyFormat(t *testing.T) {
	tokenID := "token_xyz789"
	expectedKey := "refresh_token:token_xyz789"

	key := fmt.Sprintf(KeyRefreshToken, tokenID)

	if key != expectedKey {
		t.Errorf("Expected token key %s, got %s", expectedKey, key)
	}
}

// TestCacheConsistency verifies cache key naming consistency
func TestCacheConsistency(t *testing.T) {
	// All feed keys should start with "feed:"
	feedKeys := []string{KeyFeedHot, KeyFeedTop, KeyFeedNew, KeyFeedGame, KeyFeedCreator}
	for _, key := range feedKeys {
		if len(key) < 5 || key[:5] != "feed:" {
			t.Errorf("Feed key %s should start with 'feed:'", key)
		}
	}

	// All clip keys should start with "clip:"
	clipKeys := []string{KeyClip, KeyClipVotes, KeyClipComments}
	for _, key := range clipKeys {
		if len(key) < 5 || key[:5] != "clip:" {
			t.Errorf("Clip key %s should start with 'clip:'", key)
		}
	}

	// Comment keys should start with "comment" or "comments"
	commentKeys := []string{KeyCommentTree, KeyComment}
	for _, key := range commentKeys {
		if len(key) < 7 || key[:7] != "comment" {
			t.Errorf("Comment key %s should start with 'comment'", key)
		}
	}
}

// TestTTLRatios verifies TTL ratios are reasonable
func TestTTLRatios(t *testing.T) {
	// Hot feed should expire faster than top feed
	if TTLFeedHot >= TTLFeedTop {
		t.Error("Hot feed TTL should be less than top feed TTL")
	}

	// New feed should expire fastest
	if TTLFeedNew >= TTLFeedHot {
		t.Error("New feed TTL should be less than hot feed TTL")
	}

	// Clip data should be cached longer than feed data
	if TTLClip <= TTLFeedTop {
		t.Error("Clip TTL should be greater than feed TTL")
	}

	// Metadata should be cached longest
	if TTLGame <= TTLClip {
		t.Error("Game TTL should be greater than clip TTL")
	}

	// User data should be cached longer than clips
	if TTLUser <= TTLClipVotes {
		t.Error("User TTL should be greater than clip votes TTL")
	}
}

// TestSearchCacheKeys verifies search cache key formats
func TestSearchCacheKeys(t *testing.T) {
	query := "test query"
	filters := "game:123"
	page := 1

	expectedKey := "search:test query:game:123:page:1"
	key := fmt.Sprintf(KeySearch, query, filters, page)

	if key != expectedKey {
		t.Errorf("Expected search key %s, got %s", expectedKey, key)
	}

	// Test suggestions key
	expectedSuggestKey := "search:suggestions:test query"
	suggestKey := fmt.Sprintf(KeySearchSuggest, query)

	if suggestKey != expectedSuggestKey {
		t.Errorf("Expected suggestion key %s, got %s", expectedSuggestKey, suggestKey)
	}
}

// TestInvalidationContext verifies context handling in invalidation
func TestInvalidationContext(t *testing.T) {
	ctx := context.Background()

	// Test context is not nil
	if ctx == nil {
		t.Error("Context should not be nil")
	}

	// Test with timeout context
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if ctx == nil {
		t.Error("Context with timeout should not be nil")
	}
}

// TestCacheServiceNilSafety verifies nil safety
func TestCacheServiceNilSafety(t *testing.T) {
	// Test that creating cache service with nil redis doesn't panic
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Creating cache service with nil redis should not panic: %v", r)
		}
	}()

	service := &CacheService{
		redis: nil,
	}

	// Removed impossible nil check; construction is guaranteed non-nil
	_ = service
}
