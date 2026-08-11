//go:build integration

package services

import (
	"context"
	"testing"
	"time"

	"git.subcult.tv/subculture-collective/clpr/internal/models"
	"git.subcult.tv/subculture-collective/clpr/internal/repository"
	"git.subcult.tv/subculture-collective/clpr/internal/testutil"
	"github.com/google/uuid"
)

func TestTopicClassificationUsesDirectSignalsAndSensitiveThreshold(t *testing.T) {
	pool := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(t, pool)
	testutil.TruncateTables(t, pool, "clips", "categories")
	ctx := context.Background()

	_, err := pool.Exec(ctx, `
		INSERT INTO categories (id,name,slug,category_type,is_active) VALUES
		($1,'Gaming','gaming','topic',TRUE),($2,'News & Politics','news-politics','topic',TRUE)
	`, uuid.New(), uuid.New())
	if err != nil {
		t.Fatalf("create topics: %v", err)
	}

	clipRepo := repository.NewClipRepository(pool)
	creatorID, gameID := "classifier-creator", "classifier-game"
	clip := models.Clip{
		ID: uuid.New(), TwitchClipID: uuid.NewString(), TwitchClipURL: "https://clips.twitch.tv/classifier",
		EmbedURL: "https://clips.twitch.tv/embed?clip=classifier",
		Title:    "Gaming and election discussion", CreatorName: creatorID, CreatorID: &creatorID,
		BroadcasterName: creatorID, BroadcasterID: &creatorID, GameID: &gameID,
		CreatedAt: time.Now(), ImportedAt: time.Now(),
	}
	if err := clipRepo.Create(ctx, &clip); err != nil {
		t.Fatalf("create clip: %v", err)
	}

	topicRepo := repository.NewClipTopicRepository(pool)
	classifier := NewTopicClassificationService(topicRepo)
	if err := classifier.ClassifyClip(ctx, clip.ID); err != nil {
		t.Fatalf("classify metadata: %v", err)
	}
	topics, err := topicRepo.ListForClip(ctx, clip.ID)
	if err != nil {
		t.Fatalf("list metadata topics: %v", err)
	}
	if !hasTopicSlug(topics, "gaming") {
		t.Fatalf("gaming metadata topic missing: %+v", topics)
	}
	if hasTopicSlug(topics, "news-politics") {
		t.Fatalf("politics must not be inferred from metadata alone: %+v", topics)
	}

	_, err = pool.Exec(ctx, `INSERT INTO clip_transcripts (clip_id,full_text) VALUES ($1,$2)`, clip.ID, "The election and government policy are discussed here.")
	if err != nil {
		t.Fatalf("create transcript: %v", err)
	}
	if err := classifier.ClassifyClip(ctx, clip.ID); err != nil {
		t.Fatalf("classify transcript: %v", err)
	}
	topics, err = topicRepo.ListForClip(ctx, clip.ID)
	if err != nil || !hasTopicSlug(topics, "news-politics") {
		t.Fatalf("politics transcript topic missing: %+v, err=%v", topics, err)
	}
}

func hasTopicSlug(topics []models.ClipTopic, slug string) bool {
	for _, topic := range topics {
		if topic.TopicSlug == slug {
			return true
		}
	}
	return false
}
