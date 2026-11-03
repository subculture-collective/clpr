package repository

import (
	"testing"
	"time"

	"git.subcult.tv/subculture-collective/clpr/internal/models"
)

func TestNewSortCursorValueUsesSubmittedAt(t *testing.T) {
	created := time.Date(2026, 1, 2, 3, 4, 5, 123456000, time.UTC)
	submitted := time.Date(2026, 3, 4, 5, 6, 7, 654321000, time.UTC)

	withSubmission := models.Clip{CreatedAt: created, SubmittedAt: &submitted}
	if got := NewSortCursorTime(NewSortCursorValue(&withSubmission)); !got.Equal(submitted) {
		t.Fatalf("submitted clip cursor time = %v, want %v", got, submitted)
	}

	imported := models.Clip{CreatedAt: created}
	if got := NewSortCursorTime(NewSortCursorValue(&imported)); !got.Equal(created) {
		t.Fatalf("imported clip cursor time = %v, want %v", got, created)
	}
}

func TestNewSortCursorTimeAcceptsLegacySeconds(t *testing.T) {
	legacy := time.Date(2025, 6, 7, 8, 9, 10, 0, time.UTC)
	if got := NewSortCursorTime(float64(legacy.Unix())); !got.Equal(legacy) {
		t.Fatalf("legacy cursor time = %v, want %v", got, legacy)
	}
}
