//go:build integration

package repository

import (
	"context"
	"testing"
	"time"

	"git.subcult.tv/subculture-collective/clpr/internal/models"
	"git.subcult.tv/subculture-collective/clpr/internal/testutil"
	"github.com/google/uuid"
)

func TestClipTopicModerationLifecycleAndDirectFeed(t *testing.T) {
	pool := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(t, pool)
	testutil.TruncateTables(t, pool, "clips", "categories", "users")
	ctx := context.Background()

	userID, sourceID, targetID := uuid.New(), uuid.New(), uuid.New()
	_, err := pool.Exec(ctx, `
		INSERT INTO users (id, username, email) VALUES ($1, $2, $3)
	`, userID, "topic-mod-"+userID.String()[:8], userID.String()+"@example.test")
	if err != nil {
		t.Fatalf("create moderator: %v", err)
	}
	_, err = pool.Exec(ctx, `
		INSERT INTO categories (id,name,slug,category_type,is_active) VALUES
		($1,'Source','source-topic','topic',TRUE),($2,'Target','target-topic','topic',TRUE)
	`, sourceID, targetID)
	if err != nil {
		t.Fatalf("create topics: %v", err)
	}

	clipRepo := NewClipRepository(pool)
	gameID, creatorID := "unmapped-game", "topic-creator"
	clip := models.Clip{
		ID: uuid.New(), TwitchClipID: uuid.NewString(), TwitchClipURL: "https://clips.twitch.tv/topic-test",
		EmbedURL: "https://clips.twitch.tv/embed?clip=topic-test", Title: "Topic test",
		CreatorName: creatorID, CreatorID: &creatorID, BroadcasterName: creatorID,
		BroadcasterID: &creatorID, GameID: &gameID, CreatedAt: time.Now(), ImportedAt: time.Now(),
	}
	if err := clipRepo.Create(ctx, &clip); err != nil {
		t.Fatalf("create clip: %v", err)
	}

	topicRepo := NewClipTopicRepository(pool)
	if err := topicRepo.ReplaceManual(ctx, clip.ID, userID, []uuid.UUID{sourceID}); err != nil {
		t.Fatalf("replace manual topics: %v", err)
	}
	listed, err := topicRepo.ListForClip(ctx, clip.ID)
	if err != nil || len(listed) != 1 || listed[0].Source != "manual" {
		t.Fatalf("manual topics = %+v, err=%v", listed, err)
	}

	clips, count, err := clipRepo.ListWithFilters(ctx, ClipFilters{CategoryID: &sourceID}, 20, 0)
	if err != nil || count != 1 || len(clips) != 1 || clips[0].ID != clip.ID {
		t.Fatalf("direct topic feed = %+v count=%d err=%v", clips, count, err)
	}

	if err := topicRepo.Merge(ctx, sourceID, targetID); err != nil {
		t.Fatalf("merge topics: %v", err)
	}
	listed, err = topicRepo.ListForClip(ctx, clip.ID)
	if err != nil || len(listed) != 1 || listed[0].TopicID != targetID {
		t.Fatalf("merged topics = %+v, err=%v", listed, err)
	}
	var active bool
	if err := pool.QueryRow(ctx, `SELECT is_active FROM categories WHERE id=$1`, sourceID).Scan(&active); err != nil || active {
		t.Fatalf("source active=%v, err=%v", active, err)
	}

	newTopic, err := topicRepo.Split(ctx, targetID, userID, models.SplitTopicRequest{
		Name: "Split", Slug: "split-topic", ClipIDs: []uuid.UUID{clip.ID},
	})
	if err != nil {
		t.Fatalf("split topic: %v", err)
	}
	listed, err = topicRepo.ListForClip(ctx, clip.ID)
	if err != nil || len(listed) != 1 || listed[0].TopicID != newTopic.ID {
		t.Fatalf("split topics = %+v, err=%v", listed, err)
	}
}
