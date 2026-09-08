package handlers

import (
	"net/url"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestEngagementCursorBindsGenerationAndFilters(t *testing.T) {
	generation, clip := uuid.New(), uuid.New()
	filters := url.Values{"sort": {"trending"}, "timeframe": {"week"}, "language": {"en"}}
	raw := encodeEngagementCursor(generation, clip, 12.5, filters)
	filters.Set("cursor", raw)
	filters.Set("limit", "10")
	cursor, err := decodeEngagementCursor(raw, filters)
	require.NoError(t, err)
	require.Equal(t, generation, cursor.Generation)
	require.Equal(t, clip, cursor.ClipID)
	require.Equal(t, 12.5, cursor.Score)
	filters.Set("timeframe", "day")
	_, err = decodeEngagementCursor(raw, filters)
	require.Error(t, err)
	for _, invalid := range []string{"bad", "e30", encodeEngagementCursor(uuid.Nil, clip, 1, nil), encodeEngagementCursor(generation, clip, 0, nil)} {
		_, err = decodeEngagementCursor(invalid, nil)
		require.Error(t, err)
	}
	require.Equal(t, engagementFilterKey(nil), engagementFilterKey(url.Values{"sort": {"trending"}, "timeframe": {"day"}}))
}
