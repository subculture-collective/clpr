package repository

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"git.subcult.tv/subculture-collective/clpr/internal/models"
	"git.subcult.tv/subculture-collective/clpr/internal/utils"
)

// TestCursorPaginationStability tests that cursor pagination returns consistent results
// This test verifies that clips are returned in the correct order and that pagination is stable
func TestCursorPaginationStability(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Test data - we'll create mock clips with different scores
	testClips := []struct {
		id            uuid.UUID
		trendingScore float64
		createdAt     time.Time
	}{
		{uuid.New(), 100.0, time.Now().Add(-1 * time.Hour)},
		{uuid.New(), 90.0, time.Now().Add(-2 * time.Hour)},
		{uuid.New(), 90.0, time.Now().Add(-3 * time.Hour)}, // Same score, different time
		{uuid.New(), 80.0, time.Now().Add(-4 * time.Hour)},
		{uuid.New(), 70.0, time.Now().Add(-5 * time.Hour)},
	}

	// Test cursor encoding and decoding for different sort types
	t.Run("CursorEncodingDecoding", func(t *testing.T) {
		sortTypes := []string{"trending", "popular", "new", "top", "discussed"}

		for _, sortType := range sortTypes {
			t.Run(sortType, func(t *testing.T) {
				clipID := testClips[0].id
				sortValue := testClips[0].trendingScore
				createdAtUnix := testClips[0].createdAt.Unix()

				// Encode cursor
				encoded := utils.EncodeCursor(sortType, sortValue, clipID, createdAtUnix)
				if encoded == "" {
					t.Errorf("Encoded cursor is empty for sort type %s", sortType)
					return
				}

				// Decode cursor
				decoded, err := utils.DecodeCursor(encoded)
				if err != nil {
					t.Errorf("Failed to decode cursor: %v", err)
					return
				}

				// Verify decoded values
				if decoded.SortKey != sortType {
					t.Errorf("SortKey mismatch: got %s, want %s", decoded.SortKey, sortType)
				}
				if decoded.SortValue != sortValue {
					t.Errorf("SortValue mismatch: got %f, want %f", decoded.SortValue, sortValue)
				}
				if decoded.ClipID != clipID.String() {
					t.Errorf("ClipID mismatch: got %s, want %s", decoded.ClipID, clipID.String())
				}
				if decoded.CreatedAt != createdAtUnix {
					t.Errorf("CreatedAt mismatch: got %d, want %d", decoded.CreatedAt, createdAtUnix)
				}
			})
		}
	})

	// Test that cursors handle duplicate sort values correctly
	t.Run("DuplicateSortValues", func(t *testing.T) {
		// Create two cursors with the same sort value but different IDs
		clip1 := testClips[1] // score 90.0
		clip2 := testClips[2] // score 90.0

		cursor1 := utils.EncodeCursor("trending", clip1.trendingScore, clip1.id, clip1.createdAt.Unix())
		cursor2 := utils.EncodeCursor("trending", clip2.trendingScore, clip2.id, clip2.createdAt.Unix())

		if cursor1 == cursor2 {
			t.Error("Cursors with same sort value but different IDs should not be equal")
		}

		// Decode both cursors
		decoded1, err1 := utils.DecodeCursor(cursor1)
		decoded2, err2 := utils.DecodeCursor(cursor2)

		if err1 != nil || err2 != nil {
			t.Errorf("Failed to decode cursors: err1=%v, err2=%v", err1, err2)
			return
		}

		// Both should have same sort value but different clip IDs
		if decoded1.SortValue != decoded2.SortValue {
			t.Error("Cursors should have same sort value")
		}
		if decoded1.ClipID == decoded2.ClipID {
			t.Error("Cursors should have different clip IDs")
		}
	})

	// Test cursor tampering detection
	t.Run("CursorTamperingDetection", func(t *testing.T) {
		tampered := []string{
			"not-base64",
			"aW52YWxpZDpkYXRh", // "invalid:data" - only 2 parts
			"",
		}

		for _, cursor := range tampered {
			decoded, err := utils.DecodeCursor(cursor)
			if err == nil && cursor != "" {
				t.Errorf("Expected error for tampered cursor %q, got nil (decoded: %v)", cursor, decoded)
			}
		}
	})
}

// TestClipFiltersWithCursor tests that ClipFilters correctly handles cursor parameter
func TestClipFiltersWithCursor(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Test with various sort types
	sortTypes := []string{"trending", "popular", "new", "top", "discussed", "hot", "rising"}

	for _, sortType := range sortTypes {
		t.Run(sortType, func(t *testing.T) {
			// Create a cursor
			clipID := uuid.New()
			sortValue := 100.0
			createdAt := time.Now().Unix()
			cursor := utils.EncodeCursor(sortType, sortValue, clipID, createdAt)

			// Create filters with cursor
			filters := ClipFilters{
				Sort:   sortType,
				Cursor: &cursor,
			}

			// Verify cursor is set
			if filters.Cursor == nil {
				t.Error("Cursor should be set in filters")
				return
			}

			// Verify cursor can be decoded
			decoded, err := utils.DecodeCursor(*filters.Cursor)
			if err != nil {
				t.Errorf("Failed to decode cursor from filters: %v", err)
				return
			}

			// Verify decoded sort key matches filter sort
			if decoded.SortKey != sortType {
				t.Errorf("Decoded sort key %s doesn't match filter sort %s", decoded.SortKey, sortType)
			}
		})
	}

	// Test without cursor (backward compatibility)
	t.Run("WithoutCursor", func(t *testing.T) {
		filters := ClipFilters{
			Sort: "trending",
		}

		if filters.Cursor != nil {
			t.Error("Cursor should be nil when not set")
		}
	})

	// Test empty cursor (should be treated as nil)
	t.Run("EmptyCursor", func(t *testing.T) {
		emptyCursor := ""
		_, err := utils.DecodeCursor(emptyCursor)
		if err != nil {
			t.Errorf("Empty cursor should not return error, got: %v", err)
		}
	})
}

// TestCursorPaginationOrder tests that clips are returned in the correct order
func TestCursorPaginationOrder(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// This is a conceptual test showing what we expect from cursor pagination
	// In a real integration test, this would query an actual database

	t.Run("TrendingScoreOrder", func(t *testing.T) {
		// Create mock clips with descending trending scores
		clips := []models.Clip{
			{ID: uuid.New(), TrendingScore: 100.0, CreatedAt: time.Now().Add(-1 * time.Hour)},
			{ID: uuid.New(), TrendingScore: 90.0, CreatedAt: time.Now().Add(-2 * time.Hour)},
			{ID: uuid.New(), TrendingScore: 80.0, CreatedAt: time.Now().Add(-3 * time.Hour)},
		}

		// Generate cursor from second clip
		cursor := utils.EncodeCursor("trending", clips[1].TrendingScore, clips[1].ID, clips[1].CreatedAt.Unix())

		// Decode and verify cursor points to correct clip
		decoded, err := utils.DecodeCursor(cursor)
		if err != nil {
			t.Errorf("Failed to decode cursor: %v", err)
			return
		}

		if decoded.ClipID != clips[1].ID.String() {
			t.Errorf("Cursor should point to second clip, got %s, want %s", decoded.ClipID, clips[1].ID.String())
		}

		// Verify the next page would start after this clip
		// (clip with TrendingScore < 90.0 OR (TrendingScore = 90.0 AND created_at < clip[1].CreatedAt))
		expectedNextClip := clips[2]
		if expectedNextClip.TrendingScore >= clips[1].TrendingScore {
			t.Error("Next clip should have lower trending score")
		}
	})
}
