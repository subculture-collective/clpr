//go:build integration

package services

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"git.subcult.tv/subculture-collective/clpr/internal/models"
	"git.subcult.tv/subculture-collective/clpr/internal/repository"
	"git.subcult.tv/subculture-collective/clpr/internal/testutil"
	"git.subcult.tv/subculture-collective/clpr/pkg/twitch"
	"github.com/google/uuid"
)

func TestTwitchClipPublisher_PublishIsIdempotent(t *testing.T) {
	pool := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(t, pool)
	testutil.TruncateTables(t, pool, "clips", "users")

	publisher := NewTwitchClipPublisher(repository.NewClipRepository(pool))
	ctx := context.Background()
	clip := &twitch.Clip{
		ID:              "publisher-idempotent",
		URL:             "https://clips.twitch.tv/publisher-idempotent",
		EmbedURL:        "https://clips.twitch.tv/embed?clip=publisher-idempotent",
		Title:           "First title",
		CreatorName:     "creator",
		CreatorID:       "creator-id",
		BroadcasterName: "broadcaster",
		BroadcasterID:   "broadcaster-id",
		GameID:          "509658",
		Language:        "en",
		ThumbnailURL:    "https://example.invalid/preview.jpg",
		Duration:        30,
		ViewCount:       10,
		CreatedAt:       time.Now().Add(-time.Hour),
	}

	first, err := publisher.Publish(ctx, clip)
	if err != nil {
		t.Fatalf("first Publish() error = %v", err)
	}
	if first.Disposition != PublishCreated {
		t.Fatalf("first disposition = %q, want %q", first.Disposition, PublishCreated)
	}
	ownerID := uuid.New()
	_, err = pool.Exec(ctx, `
		INSERT INTO users (id, username, display_name, email, role, account_type, account_status)
		VALUES ($1, 'publisher-owner', 'Publisher Owner', 'publisher-owner@example.com', 'user', 'member', 'active')
	`, ownerID)
	if err != nil {
		t.Fatalf("insert owner: %v", err)
	}
	_, err = pool.Exec(ctx, `
		UPDATE clips SET submitted_by_user_id = $2, submitted_at = NOW(), is_removed = TRUE, is_hidden = TRUE
		WHERE id = $1
	`, first.Clip.ID, ownerID)
	if err != nil {
		t.Fatalf("claim and moderate clip: %v", err)
	}

	clip.Title = "Updated title"
	clip.ViewCount = 25
	second, err := publisher.Publish(ctx, clip)
	if err != nil {
		t.Fatalf("second Publish() error = %v", err)
	}
	if second.Disposition != PublishUpdated {
		t.Fatalf("second disposition = %q, want %q", second.Disposition, PublishUpdated)
	}
	if second.Clip.ID != first.Clip.ID {
		t.Fatalf("duplicate publish changed clip ID: first=%s second=%s", first.Clip.ID, second.Clip.ID)
	}
	if second.Clip.Title != "Updated title" || second.Clip.ViewCount != 25 {
		t.Fatalf("duplicate publish did not refresh metadata: title=%q views=%d", second.Clip.Title, second.Clip.ViewCount)
	}
	if second.Clip.SubmittedByUserID == nil || *second.Clip.SubmittedByUserID != ownerID {
		t.Fatalf("automated republish changed clip owner: got %v, want %s", second.Clip.SubmittedByUserID, ownerID)
	}
	if !second.Clip.IsRemoved || !second.Clip.IsHidden {
		t.Fatalf("automated republish reset moderation state: removed=%t hidden=%t", second.Clip.IsRemoved, second.Clip.IsHidden)
	}
}

func TestAutoTagQueueUsesProcessingStateInsteadOfExistingTags(t *testing.T) {
	pool := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(t, pool)
	testutil.TruncateTables(t, pool, "clip_tags", "tags", "clips")

	ctx := context.Background()
	clipRepo := repository.NewClipRepository(pool)
	publisher := NewTwitchClipPublisher(clipRepo)
	result, err := publisher.Publish(ctx, &twitch.Clip{
		ID: "processing-state-test", URL: "https://clips.twitch.tv/processing-state-test",
		EmbedURL: "https://clips.twitch.tv/embed?clip=processing-state-test", Title: "Already structurally tagged",
		CreatorName: "creator", BroadcasterName: "broadcaster", CreatedAt: time.Now(),
	})
	if err != nil {
		t.Fatalf("Publish() error = %v", err)
	}

	_, err = pool.Exec(ctx, `INSERT INTO tags (id, name, slug) VALUES ($1, 'English', 'english')`, uuid.New())
	if err != nil {
		t.Fatalf("insert tag: %v", err)
	}
	_, err = pool.Exec(ctx, `INSERT INTO clip_tags (clip_id, tag_id) SELECT $1, id FROM tags WHERE slug = 'english'`, result.Clip.ID)
	if err != nil {
		t.Fatalf("attach tag: %v", err)
	}

	queued, err := clipRepo.GetUntaggedClips(ctx, 10)
	if err != nil {
		t.Fatalf("GetUntaggedClips() error = %v", err)
	}
	if len(queued) != 1 {
		t.Fatalf("queued clips = %d, want 1; existing structural tags must not starve media processing", len(queued))
	}

	if err := clipRepo.MarkAutoTagged(ctx, result.Clip.ID); err != nil {
		t.Fatalf("MarkAutoTagged() error = %v", err)
	}
	queued, err = clipRepo.GetUntaggedClips(ctx, 10)
	if err != nil {
		t.Fatalf("GetUntaggedClips() after mark error = %v", err)
	}
	if len(queued) != 0 {
		t.Fatalf("queued clips after mark = %d, want 0", len(queued))
	}
}

func TestThumbnailEnrichmentTitleSurvivesProviderRefresh(t *testing.T) {
	pool := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(t, pool)
	testutil.TruncateTables(t, pool, "clip_enrichments", "clips")

	ctx := context.Background()
	clipRepo := repository.NewClipRepository(pool)
	publisher := NewTwitchClipPublisher(clipRepo)
	twitchClip := &twitch.Clip{
		ID: "enriched-title-refresh", URL: "https://clips.twitch.tv/enriched-title-refresh",
		EmbedURL: "https://clips.twitch.tv/embed?clip=enriched-title-refresh", Title: "lol",
		CreatorName: "creator", BroadcasterName: "broadcaster", CreatedAt: time.Now(),
	}
	published, err := publisher.Publish(ctx, twitchClip)
	if err != nil {
		t.Fatalf("Publish() error = %v", err)
	}

	err = clipRepo.RecordThumbnailEnrichment(ctx, &models.ClipEnrichment{
		ClipID: published.Clip.ID, SourceTitle: "lol",
		SuggestedTitle: "A Surprised Reaction on Stream", Confidence: 0.96,
		Basis: "visible", Evidence: []string{"A surprised face is visible."},
		Tags: []string{"reaction"}, TitleAccepted: true,
	})
	if err != nil {
		t.Fatalf("RecordThumbnailEnrichment() error = %v", err)
	}

	twitchClip.Title = "provider refresh title"
	refreshed, err := publisher.Publish(ctx, twitchClip)
	if err != nil {
		t.Fatalf("second Publish() error = %v", err)
	}
	if refreshed.Clip.Title != "A Surprised Reaction on Stream" {
		t.Fatalf("provider refresh overwrote enriched title: %q", refreshed.Clip.Title)
	}

	var titleSource string
	var visionProcessed bool
	err = pool.QueryRow(ctx, `
		SELECT title_source, vision_processed_at IS NOT NULL FROM clips WHERE id = $1
	`, published.Clip.ID).Scan(&titleSource, &visionProcessed)
	if err != nil {
		t.Fatalf("query enrichment state: %v", err)
	}
	if titleSource != "ai" || !visionProcessed {
		t.Fatalf("state = source %q processed %t, want ai/true", titleSource, visionProcessed)
	}
}

func TestVisionFailureRemainsRetryable(t *testing.T) {
	pool := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(t, pool)
	testutil.TruncateTables(t, pool, "clip_enrichments", "clips")

	ctx := context.Background()
	clipRepo := repository.NewClipRepository(pool)
	publisher := NewTwitchClipPublisher(clipRepo)
	thumbnailURL := "https://static-cdn.jtvnw.net/example.jpg"
	published, err := publisher.Publish(ctx, &twitch.Clip{
		ID: "retryable-vision", URL: "https://clips.twitch.tv/retryable-vision",
		EmbedURL: "https://clips.twitch.tv/embed?clip=retryable-vision", Title: "clip",
		CreatorName: "creator", BroadcasterName: "broadcaster", CreatedAt: time.Now(),
		ThumbnailURL: thumbnailURL,
	})
	if err != nil {
		t.Fatalf("Publish() error = %v", err)
	}
	if err := clipRepo.RecordVisionFailure(ctx, published.Clip.ID, context.DeadlineExceeded); err != nil {
		t.Fatalf("RecordVisionFailure() error = %v", err)
	}

	var processed bool
	var attempts int
	var failure string
	err = pool.QueryRow(ctx, `
		SELECT vision_processed_at IS NOT NULL, vision_attempt_count, vision_error
		FROM clips WHERE id = $1
	`, published.Clip.ID).Scan(&processed, &attempts, &failure)
	if err != nil {
		t.Fatalf("query failure state: %v", err)
	}
	if processed || attempts != 1 || failure == "" {
		t.Fatalf("failure state = processed %t attempts %d error %q", processed, attempts, failure)
	}
}

func TestClipTranscriptCanBeStoredAndRetrieved(t *testing.T) {
	pool := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(t, pool)
	testutil.TruncateTables(t, pool, "clip_transcripts", "clips")

	ctx := context.Background()
	clipRepo := repository.NewClipRepository(pool)
	publisher := NewTwitchClipPublisher(clipRepo)
	broadcasterID := "authorized-broadcaster"
	published, err := publisher.Publish(ctx, &twitch.Clip{
		ID: "transcript-storage", URL: "https://clips.twitch.tv/transcript-storage",
		EmbedURL: "https://clips.twitch.tv/embed?clip=transcript-storage", Title: "clip",
		CreatorName: "creator", BroadcasterName: "broadcaster", BroadcasterID: broadcasterID,
		CreatedAt: time.Now(),
	})
	if err != nil {
		t.Fatalf("Publish() error = %v", err)
	}
	segments := json.RawMessage(`[{
		"start":0,"end":1.2,"text":"what a save","avg_logprob":-0.1
	}]`)
	err = clipRepo.RecordClipTranscript(ctx, &models.ClipTranscript{
		ClipID: published.Clip.ID, Language: "en", FullText: "what a save",
		Segments: segments, Source: "twitch_authorized_whisper",
	})
	if err != nil {
		t.Fatalf("RecordClipTranscript() error = %v", err)
	}

	transcript, err := clipRepo.GetClipTranscript(ctx, published.Clip.ID)
	if err != nil {
		t.Fatalf("GetClipTranscript() error = %v", err)
	}
	if transcript == nil || transcript.FullText != "what a save" || transcript.Language != "en" {
		t.Fatalf("retrieved transcript = %#v", transcript)
	}
}

func TestTranscriptionQueueStaysIdleWithoutBroadcasterAuthorization(t *testing.T) {
	pool := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(t, pool)
	testutil.TruncateTables(t, pool, "clip_transcripts", "twitch_auth", "clips")

	ctx := context.Background()
	clipRepo := repository.NewClipRepository(pool)
	publisher := NewTwitchClipPublisher(clipRepo)
	broadcasterID := "not-yet-authorized"
	_, err := publisher.Publish(ctx, &twitch.Clip{
		ID: "no-transcription-authorization", URL: "https://clips.twitch.tv/no-transcription-authorization",
		EmbedURL: "https://clips.twitch.tv/embed?clip=no-transcription-authorization", Title: "clip",
		CreatorName: "creator", BroadcasterName: "broadcaster", BroadcasterID: broadcasterID,
		CreatedAt: time.Now(),
	})
	if err != nil {
		t.Fatalf("Publish() error = %v", err)
	}

	queued, err := clipRepo.GetClipsNeedingTranscription(ctx, 10)
	if err != nil {
		t.Fatalf("GetClipsNeedingTranscription() error = %v", err)
	}
	if len(queued) != 0 {
		t.Fatalf("unauthorized transcription queue length = %d, want 0", len(queued))
	}
}
