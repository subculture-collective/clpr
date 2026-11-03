package handlers

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

// Regression: cursors were formatted with time.RFC3339, dropping sub-second
// precision, and had no tie-breaker, so messages were skipped between pages.
func TestChatHistoryCursor_RoundTripKeepsPrecisionAndID(t *testing.T) {
	createdAt := time.Date(2026, 11, 2, 10, 30, 15, 123456000, time.UTC)
	id := uuid.New()

	cursor := formatChatHistoryCursor(createdAt, id)
	gotTime, gotID, err := parseChatHistoryCursor(cursor)
	if err != nil {
		t.Fatalf("parse %q: %v", cursor, err)
	}
	if !gotTime.Equal(createdAt) {
		t.Fatalf("time = %v, want %v (sub-second precision lost)", gotTime, createdAt)
	}
	if gotID == nil || *gotID != id {
		t.Fatalf("id = %v, want %v", gotID, id)
	}
}

func TestParseChatHistoryCursor_AcceptsLegacyTimestamps(t *testing.T) {
	for _, cursor := range []string{"2026-11-02T10:30:15Z", "2026-11-02T10:30:15.123456Z", "2026-11-02T05:30:15-05:00"} {
		_, id, err := parseChatHistoryCursor(cursor)
		if err != nil {
			t.Errorf("legacy cursor %q rejected: %v", cursor, err)
		}
		if id != nil {
			t.Errorf("legacy cursor %q should have no tie-breaker", cursor)
		}
	}
}

func TestParseChatHistoryCursor_RejectsMalformed(t *testing.T) {
	for _, cursor := range []string{"yesterday", "2026-11-02T10:30:15Z_not-a-uuid", "_" + uuid.New().String()} {
		if _, _, err := parseChatHistoryCursor(cursor); err == nil {
			t.Errorf("cursor %q should be rejected", cursor)
		}
	}
}
