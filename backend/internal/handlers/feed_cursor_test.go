package handlers

import (
	"testing"
	"time"

	"git.subcult.tv/subculture-collective/clpr/internal/models"
	"git.subcult.tv/subculture-collective/clpr/internal/utils"
	"github.com/google/uuid"
)

func TestFeedCursorForUsesTheRepositorySortKeys(t *testing.T) {
	created := time.Date(2026, 10, 30, 12, 0, 0, 700_000_000, time.UTC)
	snapshot := time.Date(2026, 11, 2, 9, 30, 0, 0, time.UTC)
	clip := models.Clip{
		ID:              uuid.New(),
		CreatedAt:       created,
		VoteScore:       42,
		ViewCount:       1999,
		CommentCount:    7,
		TrendingScore:   0, // a real score of zero, not "missing"
		PopularityIndex: 0, // never NULL in query results
		EngagementCount: 900,
	}

	tests := []struct {
		sort         string
		wantValue    float64
		wantSnapshot bool
	}{
		{sort: "trending", wantValue: 0},
		{sort: "popular", wantValue: 0},
		{sort: "top", wantValue: 42},
		{sort: "discussed", wantValue: 7},
		{sort: "hot", wantValue: 42, wantSnapshot: true},
		{sort: "rising", wantValue: 42 + 19, wantSnapshot: true},
	}
	for _, tt := range tests {
		t.Run(tt.sort, func(t *testing.T) {
			cursor := feedCursorFor(tt.sort, &clip, "", snapshot)
			if cursor.SortValue != tt.wantValue {
				t.Errorf("SortValue = %v, want %v", cursor.SortValue, tt.wantValue)
			}
			if !utils.CursorTime(cursor.CreatedAt).Equal(created) {
				t.Errorf("CreatedAt lost sub-second precision: %v", utils.CursorTime(cursor.CreatedAt))
			}
			if gotSnapshot := cursor.Snapshot != 0; gotSnapshot != tt.wantSnapshot {
				t.Errorf("snapshot present = %v, want %v", gotSnapshot, tt.wantSnapshot)
			}
			if tt.wantSnapshot && !utils.CursorTime(cursor.Snapshot).Equal(snapshot) {
				t.Errorf("snapshot = %v, want %v", utils.CursorTime(cursor.Snapshot), snapshot)
			}
			if _, err := utils.DecodeCursor(utils.EncodeFullCursor(cursor)); err != nil {
				t.Errorf("cursor does not decode: %v", err)
			}
		})
	}
}
