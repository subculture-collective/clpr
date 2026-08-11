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

func createCurationClip(t *testing.T, repo *ClipRepository, title, gameID, creatorID string, views int, velocity float64, createdAt time.Time) models.Clip {
	t.Helper()
	clip := models.Clip{
		ID: uuid.New(), TwitchClipID: uuid.NewString(), TwitchClipURL: "https://clips.twitch.tv/" + title,
		EmbedURL: "https://clips.twitch.tv/embed?clip=" + title, Title: title,
		CreatorName: creatorID, CreatorID: &creatorID, BroadcasterName: creatorID,
		BroadcasterID: &creatorID, GameID: &gameID, ViewCount: views,
		CreatedAt: createdAt, ImportedAt: time.Now(),
	}
	if err := repo.Create(context.Background(), &clip); err != nil {
		t.Fatalf("create clip %s: %v", title, err)
	}
	if _, err := repo.pool.Exec(context.Background(), `UPDATE clips SET view_velocity = $1 WHERE id = $2`, velocity, clip.ID); err != nil {
		t.Fatalf("set clip velocity: %v", err)
	}
	if _, err := repo.pool.Exec(context.Background(), `
		INSERT INTO clip_topics (clip_id, topic_id, source, confidence)
		SELECT $1, cg.category_id, 'twitch_category', .55
		FROM games g JOIN category_games cg ON cg.game_id = g.id
		WHERE g.twitch_game_id = $2
		ON CONFLICT DO NOTHING
	`, clip.ID, gameID); err != nil {
		t.Fatalf("classify clip topic: %v", err)
	}
	return clip
}

func mapCurationGameToTopic(t *testing.T, repo *ClipRepository, gameID, topicSlug string) {
	t.Helper()
	gameUUID, topicUUID := uuid.New(), uuid.New()
	ctx := context.Background()
	if _, err := repo.pool.Exec(ctx, `
		INSERT INTO games (id, twitch_game_id, name, slug) VALUES ($1, $2, $3, $4)
		ON CONFLICT (twitch_game_id) DO UPDATE SET name = EXCLUDED.name
	`, gameUUID, gameID, gameID, gameID); err != nil {
		t.Fatalf("create game %s: %v", gameID, err)
	}
	if _, err := repo.pool.Exec(ctx, `
		INSERT INTO categories (id, name, slug, position) VALUES ($1, $2, $2, 100)
		ON CONFLICT (slug) DO NOTHING
	`, topicUUID, topicSlug); err != nil {
		t.Fatalf("create topic %s: %v", topicSlug, err)
	}
	if _, err := repo.pool.Exec(ctx, `
		INSERT INTO category_games (game_id, category_id)
		SELECT g.id, c.id FROM games g, categories c
		WHERE g.twitch_game_id = $1 AND c.slug = $2
		ON CONFLICT DO NOTHING
	`, gameID, topicSlug); err != nil {
		t.Fatalf("map game %s to topic %s: %v", gameID, topicSlug, err)
	}
}

func TestDiversityRoulettePrioritizesCreatorAndTopicDiversity(t *testing.T) {
	pool := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(t, pool)
	testutil.TruncateTables(t, pool, "clips")
	clipRepo := NewClipRepository(pool)
	mapCurationGameToTopic(t, clipRepo, "game-a", "topic-a")
	mapCurationGameToTopic(t, clipRepo, "game-b", "topic-a")
	now := time.Now()
	clips := []models.Clip{
		createCurationClip(t, clipRepo, "a1", "game-a", "creator-a", 1000, 20, now.Add(-time.Hour)),
		createCurationClip(t, clipRepo, "a2", "game-a", "creator-b", 900, 18, now.Add(-2*time.Hour)),
		createCurationClip(t, clipRepo, "b1", "game-b", "creator-c", 800, 16, now.Add(-time.Hour)),
	}
	creatorByClip := map[uuid.UUID]string{}
	for _, clip := range clips {
		creatorByClip[clip.ID] = *clip.CreatorID
	}

	week := "week"
	result, err := NewPlaylistCurationRepository(pool).DiversityRoulette(context.Background(), &models.PlaylistScript{Timeframe: &week, ClipLimit: 10})
	if err != nil {
		t.Fatalf("diversity roulette: %v", err)
	}
	if len(result) != 3 {
		t.Fatalf("result count = %d, want one clip for each creator", len(result))
	}
	seen := map[string]bool{}
	for _, clip := range result {
		creator := creatorByClip[clip.ID]
		if seen[creator] {
			t.Fatalf("creator %q appeared more than once", creator)
		}
		seen[creator] = true
	}
}

func TestClipOfTheDayPrioritizesCurrentVelocity(t *testing.T) {
	pool := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(t, pool)
	testutil.TruncateTables(t, pool, "clips")
	clipRepo := NewClipRepository(pool)
	now := time.Now()
	createCurationClip(t, clipRepo, "slow", "game-a", "creator-a", 100000, 0.1, now.Add(-20*time.Hour))
	fast := createCurationClip(t, clipRepo, "fast", "game-b", "creator-b", 5000, 500, now.Add(-time.Hour))

	day := "day"
	result, err := NewPlaylistCurationRepository(pool).ClipOfTheDay(context.Background(), &models.PlaylistScript{Timeframe: &day, ClipLimit: 1})
	if err != nil {
		t.Fatalf("clip of the day: %v", err)
	}
	if len(result) != 1 || result[0].ID != fast.ID {
		t.Fatalf("clip of the day = %+v, want fast clip %s", result, fast.ID)
	}
}

func TestWeekendMixEnforcesCreatorAndTopicDiversity(t *testing.T) {
	pool := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(t, pool)
	testutil.TruncateTables(t, pool, "clips")
	clipRepo := NewClipRepository(pool)
	mapCurationGameToTopic(t, clipRepo, "game-a", "topic-weekend-a")
	mapCurationGameToTopic(t, clipRepo, "game-b", "topic-weekend-b")
	now := time.Now()
	clips := []models.Clip{
		createCurationClip(t, clipRepo, "a1", "game-a", "creator-a", 1000, 30, now.Add(-time.Hour)),
		createCurationClip(t, clipRepo, "a2", "game-a", "creator-a", 900, 20, now.Add(-2*time.Hour)),
		createCurationClip(t, clipRepo, "b1", "game-a", "creator-b", 800, 18, now.Add(-time.Hour)),
		createCurationClip(t, clipRepo, "c1", "game-a", "creator-c", 700, 16, now.Add(-time.Hour)),
		createCurationClip(t, clipRepo, "d1", "game-b", "creator-d", 600, 14, now.Add(-time.Hour)),
	}
	creatorByClip, gameByClip := map[uuid.UUID]string{}, map[uuid.UUID]string{}
	for _, clip := range clips {
		creatorByClip[clip.ID] = *clip.CreatorID
		gameByClip[clip.ID] = *clip.GameID
	}

	week := "week"
	result, err := NewPlaylistCurationRepository(pool).WeekendMix(context.Background(), &models.PlaylistScript{Timeframe: &week, ClipLimit: 20})
	if err != nil {
		t.Fatalf("weekend mix: %v", err)
	}
	seenCreators, topicCounts := map[string]bool{}, map[string]int{}
	for _, clip := range result {
		creator := creatorByClip[clip.ID]
		if seenCreators[creator] {
			t.Fatalf("creator %q appeared more than once", creator)
		}
		seenCreators[creator] = true
		topic := "topic-weekend-a"
		if gameByClip[clip.ID] == "game-b" {
			topic = "topic-weekend-b"
		}
		topicCounts[topic]++
	}
	if topicCounts["topic-weekend-a"] > 4 {
		t.Fatalf("topic-weekend-a appeared %d times, want at most 4", topicCounts["topic-weekend-a"])
	}
}
