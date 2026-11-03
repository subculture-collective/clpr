package utils

import (
	"encoding/base64"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
)

// Cursor represents a pagination cursor containing sort key values and clip ID
type Cursor struct {
	SortKey     string  // The sort field name (e.g., "trending_score", "created_at")
	SortValue   float64 // The value of the sort field for the last item
	ClipID      string  // The clip ID for tie-breaking
	CreatedAt   int64   // created_at for tie-breaking: Unix microseconds (older cursors: Unix seconds)
	ShuffleSeed string  // Optional stateless feed-session seed
	// Snapshot is the Unix-microsecond "now" that time-dependent sorts
	// (rising) are evaluated at. Every page of one feed session uses the same
	// snapshot so the ordering cannot shift between pages. Zero when absent.
	Snapshot int64
}

// legacyCursorSecondsLimit separates timestamps written as Unix seconds by
// older cursors from Unix microseconds (seconds stay below 1e11 until 5138).
const legacyCursorSecondsLimit = 1e11

// CursorTimeValue encodes a cursor timestamp as Unix microseconds.
func CursorTimeValue(t time.Time) int64 {
	return t.UnixMicro()
}

// CursorTime decodes a cursor timestamp into UTC. Values written as Unix
// seconds by older cursors are still accepted.
func CursorTime(value int64) time.Time {
	if value < legacyCursorSecondsLimit && value > -legacyCursorSecondsLimit {
		return time.Unix(value, 0).UTC()
	}
	return time.UnixMicro(value).UTC()
}

// EncodeCursor encodes a cursor into a base64 string
// Format: sort_key:sort_value:clip_id:created_at[:shuffle_seed[:snapshot]]
func EncodeCursor(sortKey string, sortValue float64, clipID uuid.UUID, createdAtUnix int64) string {
	return EncodeCursorWithShuffleSeed(sortKey, sortValue, clipID, createdAtUnix, "")

}

// EncodeCursorWithShuffleSeed keeps a personalized shuffle stable while the
// client paginates, without storing feed sessions on the server.
func EncodeCursorWithShuffleSeed(sortKey string, sortValue float64, clipID uuid.UUID, createdAtUnix int64, shuffleSeed string) string {
	return EncodeFullCursor(Cursor{
		SortKey:     sortKey,
		SortValue:   sortValue,
		ClipID:      clipID.String(),
		CreatedAt:   createdAtUnix,
		ShuffleSeed: shuffleSeed,
	})
}

// EncodeFullCursor encodes every cursor field. The shuffle seed and snapshot
// parts are only written when set; a snapshot without a seed leaves the seed
// part empty.
func EncodeFullCursor(c Cursor) string {
	data := fmt.Sprintf("%s:%s:%s:%d", c.SortKey, strconv.FormatFloat(c.SortValue, 'g', -1, 64), c.ClipID, c.CreatedAt)
	if c.ShuffleSeed != "" || c.Snapshot != 0 {
		data += ":" + c.ShuffleSeed
	}
	if c.Snapshot != 0 {
		data += ":" + strconv.FormatInt(c.Snapshot, 10)
	}
	return base64.URLEncoding.EncodeToString([]byte(data))
}

// DecodeCursor decodes a base64 cursor string into a Cursor struct
func DecodeCursor(cursorStr string) (*Cursor, error) {
	if cursorStr == "" {
		return nil, nil
	}

	decoded, err := base64.URLEncoding.DecodeString(cursorStr)
	if err != nil {
		return nil, fmt.Errorf("invalid cursor format: failed to decode base64")
	}

	parts := strings.Split(string(decoded), ":")
	if len(parts) < 4 || len(parts) > 6 {
		return nil, fmt.Errorf("invalid cursor format: expected 4 to 6 parts, got %d", len(parts))
	}

	// Validate sortKey to prevent injection attacks
	sortKey := parts[0]
	validSortKeys := map[string]bool{
		"trending":  true,
		"popular":   true,
		"new":       true,
		"top":       true,
		"discussed": true,
		"hot":       true,
		"rising":    true,
	}
	if !validSortKeys[sortKey] {
		return nil, fmt.Errorf("invalid cursor format: invalid sort key %q", sortKey)
	}

	sortValue, err := strconv.ParseFloat(parts[1], 64)
	if err != nil {
		return nil, fmt.Errorf("invalid cursor format: invalid sort value")
	}

	clipID := parts[2]
	if _, err := uuid.Parse(clipID); err != nil {
		return nil, fmt.Errorf("invalid cursor format: invalid clip ID")
	}

	createdAt, err := strconv.ParseInt(parts[3], 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid cursor format: invalid created_at timestamp")
	}

	shuffleSeed := ""
	if len(parts) >= 5 && (len(parts) == 5 || parts[4] != "") {
		shuffleSeed = parts[4]
		if _, err := uuid.Parse(shuffleSeed); err != nil {
			return nil, fmt.Errorf("invalid cursor format: invalid shuffle seed")
		}
	}

	var snapshot int64
	if len(parts) == 6 {
		snapshot, err = strconv.ParseInt(parts[5], 10, 64)
		if err != nil || snapshot <= 0 {
			return nil, fmt.Errorf("invalid cursor format: invalid snapshot")
		}
	}

	return &Cursor{
		SortKey:     sortKey,
		SortValue:   sortValue,
		ClipID:      clipID,
		CreatedAt:   createdAt,
		ShuffleSeed: shuffleSeed,
		Snapshot:    snapshot,
	}, nil
}
