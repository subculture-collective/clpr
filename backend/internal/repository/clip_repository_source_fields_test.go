package repository

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"git.subcult.tv/subculture-collective/clpr/internal/models"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type captureClipDB struct {
	query string
	args  []any
}

func (c *captureClipDB) Begin(context.Context) (pgx.Tx, error) {
	panic("Begin is not used by this test double")
}

func (c *captureClipDB) Exec(ctx context.Context, query string, args ...any) (pgconn.CommandTag, error) {
	c.query = query
	c.args = append([]any(nil), args...)
	var tag pgconn.CommandTag
	return tag, nil
}

func (c *captureClipDB) Query(ctx context.Context, query string, args ...any) (pgx.Rows, error) {
	return nil, nil
}

func (c *captureClipDB) QueryRow(ctx context.Context, query string, args ...any) pgx.Row {
	return nil
}

func TestClipRepository_Create_IncludesSourceFields(t *testing.T) {
	pool := &captureClipDB{}
	repo := &ClipRepository{pool: pool}
	creatorID := uuid.NewString()
	broadcasterID := uuid.NewString()
	clip := &models.Clip{
		ID:                uuid.New(),
		TwitchClipID:      "source-123",
		TwitchClipURL:     "https://example.com/source-123",
		EmbedURL:          "https://player.example/embed/source-123",
		Title:             "Example clip",
		CreatorName:       "Creator",
		CreatorID:         &creatorID,
		CreatorAccountID:  func() *uuid.UUID { id := uuid.New(); return &id }(),
		BroadcasterName:   "Broadcaster",
		BroadcasterID:     &broadcasterID,
		Duration:          repoFloat64Ptr(12.5),
		ViewCount:         17,
		CreatedAt:         time.Now().UTC(),
		ImportedAt:        time.Now().UTC(),
		SubmittedByUserID: func() *uuid.UUID { id := uuid.New(); return &id }(),
		SubmittedAt:       func() *time.Time { ts := time.Now().UTC(); return &ts }(),
		SourceType:        "external",
		SourcePlatform:    "youtube",
		SourceURL:         repoStringPtr("https://www.youtube.com/watch?v=abc123"),
		SourceID:          repoStringPtr("abc123"),
		SourceMetadata:    []byte(`{"embed_url":"https://player.example/embed/abc123"}`),
		DurationSeconds:   repoIntPtr(42),
		DurationVerified:  true,
		StorageProvider:   repoStringPtr("s3"),
		StorageBucket:     repoStringPtr("clips"),
		StorageKey:        repoStringPtr("uploads/public/abc123.mp4"),
		OriginalFilename:  repoStringPtr("abc123.mp4"),
		MimeType:          repoStringPtr("video/mp4"),
		FileSizeBytes:     repoInt64Ptr(2048),
		StreamSource:      repoStringPtr("external"),
		Status:            repoStringPtr("ready"),
		VideoURL:          repoStringPtr("https://cdn.example.com/uploads/public/abc123.mp4"),
	}

	if err := repo.Create(context.Background(), clip); err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if !strings.Contains(pool.query, "source_type") || !strings.Contains(pool.query, "video_url") {
		t.Fatalf("query missing source/upload columns: %s", pool.query)
	}
	if got := len(pool.args); got != 47 {
		t.Fatalf("args len = %d, want 47", got)
	}
	if pool.args[7] != clip.CreatorAccountID {
		t.Fatalf("creator_account_id arg = %v, want %v", pool.args[7], clip.CreatorAccountID)
	}
	if pool.args[27] != clip.SourceType || pool.args[28] != clip.SourcePlatform {
		t.Fatalf("source args = %v/%v, want %q/%q", pool.args[27], pool.args[28], clip.SourceType, clip.SourcePlatform)
	}
	if pool.args[29] != clip.SourceURL {
		t.Fatalf("source_url arg = %v, want %v", pool.args[29], clip.SourceURL)
	}
	if pool.args[30] != clip.SourceID {
		t.Fatalf("source_id arg = %v, want %v", pool.args[30], clip.SourceID)
	}
	if string(pool.args[31].(json.RawMessage)) != string(clip.SourceMetadata) {
		t.Fatalf("source_metadata arg = %s, want %s", string(pool.args[31].(json.RawMessage)), string(clip.SourceMetadata))
	}
	if pool.args[40] != clip.StreamSource || pool.args[41] != clip.Status {
		t.Fatalf("stream args = %v/%v, want %v/%v", pool.args[40], pool.args[41], clip.StreamSource, clip.Status)
	}
	if pool.args[42] != clip.VideoURL {
		t.Fatalf("video_url arg = %v, want %v", pool.args[42], clip.VideoURL)
	}
}

func repoStringPtr(s string) *string    { return &s }
func repoIntPtr(v int) *int             { return &v }
func repoInt64Ptr(v int64) *int64       { return &v }
func repoFloat64Ptr(v float64) *float64 { return &v }
