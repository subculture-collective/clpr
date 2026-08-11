//go:build integration

package repository

import (
	"context"
	"fmt"
	"reflect"
	"testing"
	"time"

	"git.subcult.tv/subculture-collective/clpr/internal/models"
	"git.subcult.tv/subculture-collective/clpr/internal/testutil"
	"git.subcult.tv/subculture-collective/clpr/internal/utils"
	"github.com/google/uuid"
)

func TestClipRepository_ListWithFilters_Discussed(t *testing.T) {
	// Skip if not in integration test mode
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	pool := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(t, pool)
	// Ensure a clean slate to avoid unique constraint collisions from prior runs
	testutil.TruncateTables(t, pool, "clips", "top_streamers", "clip_tags", "favorites", "comments", "votes", "comment_votes")

	repo := NewClipRepository(pool)
	ctx := context.Background()

	makeID := func(base string) string { return fmt.Sprintf("%s-%s", base, uuid.NewString()) }

	// Create test clips with different comment counts
	clip1 := &models.Clip{
		ID:              uuid.New(),
		TwitchClipID:    makeID("test-clip-1"),
		TwitchClipURL:   "https://clips.twitch.tv/test1",
		EmbedURL:        "https://clips.twitch.tv/embed?clip=test1",
		Title:           "Test Clip 1 - Most Discussed",
		CreatorName:     "creator1",
		BroadcasterName: "broadcaster1",
		BroadcasterID:   testutil.StringPtr("12345"),
		CreatedAt:       time.Now().Add(-1 * time.Hour),
		ImportedAt:      time.Now(),
		CommentCount:    100,
	}

	clip2 := &models.Clip{
		ID:              uuid.New(),
		TwitchClipID:    makeID("test-clip-2"),
		TwitchClipURL:   "https://clips.twitch.tv/test2",
		EmbedURL:        "https://clips.twitch.tv/embed?clip=test2",
		Title:           "Test Clip 2 - Less Discussed",
		CreatorName:     "creator2",
		BroadcasterName: "broadcaster2",
		BroadcasterID:   testutil.StringPtr("23456"),
		CreatedAt:       time.Now().Add(-2 * time.Hour),
		ImportedAt:      time.Now(),
		CommentCount:    50,
	}

	clip3 := &models.Clip{
		ID:              uuid.New(),
		TwitchClipID:    makeID("test-clip-3"),
		TwitchClipURL:   "https://clips.twitch.tv/test3",
		EmbedURL:        "https://clips.twitch.tv/embed?clip=test3",
		Title:           "Test Clip 3 - No Discussion",
		CreatorName:     "creator3",
		BroadcasterName: "broadcaster3",
		BroadcasterID:   testutil.StringPtr("34567"),
		CreatedAt:       time.Now().Add(-3 * time.Hour),
		ImportedAt:      time.Now(),
		CommentCount:    0,
	}

	// Insert clips
	if err := repo.Create(ctx, clip1); err != nil {
		t.Fatalf("Failed to create clip1: %v", err)
	}
	if err := repo.Create(ctx, clip2); err != nil {
		t.Fatalf("Failed to create clip2: %v", err)
	}
	if err := repo.Create(ctx, clip3); err != nil {
		t.Fatalf("Failed to create clip3: %v", err)
	}

	// Test discussed sort
	filters := ClipFilters{
		Sort: "discussed",
	}

	clips, total, err := repo.ListWithFilters(ctx, filters, 10, 0)
	if err != nil {
		t.Fatalf("Failed to list clips: %v", err)
	}

	if total < 3 {
		t.Errorf("Expected at least 3 clips, got %d", total)
	}

	if len(clips) < 3 {
		t.Errorf("Expected at least 3 clips in result, got %d", len(clips))
	}

	// Verify clips are sorted by comment count (descending)
	if clips[0].CommentCount < clips[1].CommentCount {
		t.Errorf("Clips not sorted by comment count: clip[0]=%d, clip[1]=%d",
			clips[0].CommentCount, clips[1].CommentCount)
	}
	if clips[1].CommentCount < clips[2].CommentCount {
		t.Errorf("Clips not sorted by comment count: clip[1]=%d, clip[2]=%d",
			clips[1].CommentCount, clips[2].CommentCount)
	}
}

func TestClipRepository_UpdateViewCountRecordsVelocity(t *testing.T) {
	pool := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(t, pool)
	testutil.TruncateTables(t, pool, "clips")

	repo := NewClipRepository(pool)
	ctx := context.Background()
	clip := &models.Clip{
		ID: uuid.New(), TwitchClipID: "velocity-" + uuid.NewString(),
		TwitchClipURL: "https://clips.twitch.tv/velocity", EmbedURL: "https://clips.twitch.tv/embed?clip=velocity",
		Title: "Velocity", CreatorName: "creator", BroadcasterName: "broadcaster",
		ViewCount: 100, CreatedAt: time.Now().Add(-4 * time.Hour), ImportedAt: time.Now().Add(-2 * time.Hour),
	}
	if err := repo.Create(ctx, clip); err != nil {
		t.Fatalf("create clip: %v", err)
	}
	if _, err := pool.Exec(ctx, `UPDATE clips SET view_count_observed_at = NOW() - INTERVAL '2 hours' WHERE id = $1`, clip.ID); err != nil {
		t.Fatalf("arrange prior observation: %v", err)
	}

	if err := repo.UpdateViewCount(ctx, clip.TwitchClipID, 200); err != nil {
		t.Fatalf("update view count: %v", err)
	}

	var previous int
	var velocity float64
	if err := pool.QueryRow(ctx, `SELECT previous_view_count, view_velocity FROM clips WHERE id = $1`, clip.ID).Scan(&previous, &velocity); err != nil {
		t.Fatalf("read velocity: %v", err)
	}
	if previous != 100 {
		t.Fatalf("previous view count = %d, want 100", previous)
	}
	if velocity < 49 || velocity > 51 {
		t.Fatalf("view velocity = %.2f, want approximately 50 views/hour", velocity)
	}
}

func TestClipRepository_PublishAutomatedClipRecordsVelocity(t *testing.T) {
	pool := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(t, pool)
	testutil.TruncateTables(t, pool, "clips")

	repo := NewClipRepository(pool)
	ctx := context.Background()
	clip := &models.Clip{
		ID: uuid.New(), TwitchClipID: "publish-velocity-" + uuid.NewString(),
		TwitchClipURL: "https://clips.twitch.tv/publish-velocity", EmbedURL: "https://clips.twitch.tv/embed?clip=publish-velocity",
		Title: "Publish Velocity", CreatorName: "creator", BroadcasterName: "broadcaster",
		ViewCount: 100, CreatedAt: time.Now().Add(-4 * time.Hour), ImportedAt: time.Now().Add(-2 * time.Hour),
	}
	if _, inserted, err := repo.PublishAutomatedClip(ctx, clip); err != nil || !inserted {
		t.Fatalf("initial publish: inserted=%v err=%v", inserted, err)
	}
	if _, err := pool.Exec(ctx, `UPDATE clips SET view_count_observed_at = NOW() - INTERVAL '2 hours' WHERE id = $1`, clip.ID); err != nil {
		t.Fatalf("arrange prior observation: %v", err)
	}

	clip.ViewCount = 200
	if _, inserted, err := repo.PublishAutomatedClip(ctx, clip); err != nil || inserted {
		t.Fatalf("refresh publish: inserted=%v err=%v", inserted, err)
	}

	var previous int
	var velocity float64
	if err := pool.QueryRow(ctx, `SELECT previous_view_count, view_velocity FROM clips WHERE id = $1`, clip.ID).Scan(&previous, &velocity); err != nil {
		t.Fatalf("read velocity: %v", err)
	}
	if previous != 100 {
		t.Fatalf("previous view count = %d, want 100", previous)
	}
	if velocity < 49 || velocity > 51 {
		t.Fatalf("view velocity = %.2f, want approximately 50 views/hour", velocity)
	}
}

func TestClipRepository_MixedTrendingRankUsesVelocityAndUserBoost(t *testing.T) {
	pool := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(t, pool)
	testutil.TruncateTables(t, pool, "clips", "users")

	ctx := context.Background()
	userRepo := NewUserRepository(pool)
	user := &models.User{ID: uuid.New(), TwitchID: testutil.StringPtr("rank-user-" + uuid.NewString()), Username: "rank-user", DisplayName: "Rank User"}
	if err := userRepo.Create(ctx, user); err != nil {
		t.Fatalf("create user: %v", err)
	}

	clipRepo := NewClipRepository(pool)
	createdAt := time.Now().Add(-12 * time.Hour)
	makeClip := func(label string) *models.Clip {
		return &models.Clip{
			ID: uuid.New(), TwitchClipID: label + "-" + uuid.NewString(),
			TwitchClipURL: "https://clips.twitch.tv/" + label, EmbedURL: "https://clips.twitch.tv/embed?clip=" + label,
			Title: label, CreatorName: "creator", BroadcasterName: "broadcaster",
			ViewCount: 100, CreatedAt: createdAt, ImportedAt: time.Now(),
		}
	}
	baseline := makeClip("baseline")
	velocity := makeClip("velocity")
	community := makeClip("community")
	community.SubmittedByUserID = &user.ID
	community.SubmittedAt = testutil.TimePtr(time.Now())
	for _, clip := range []*models.Clip{baseline, velocity, community} {
		if err := clipRepo.Create(ctx, clip); err != nil {
			t.Fatalf("create %s: %v", clip.Title, err)
		}
	}
	if _, err := pool.Exec(ctx, `UPDATE clips SET previous_view_count = 50, view_velocity = 100 WHERE id = $1`, velocity.ID); err != nil {
		t.Fatalf("arrange velocity: %v", err)
	}

	if _, err := clipRepo.UpdateTrendingScores(ctx); err != nil {
		t.Fatalf("update mixed ranks: %v", err)
	}
	first, _, err := clipRepo.ListWithFilters(ctx, ClipFilters{Sort: "trending"}, 10, 0)
	if err != nil {
		t.Fatalf("list mixed rank: %v", err)
	}
	scores := map[uuid.UUID]float64{}
	order := make([]uuid.UUID, 0, len(first))
	for _, clip := range first {
		scores[clip.ID] = clip.TrendingScore
		order = append(order, clip.ID)
	}
	if scores[community.ID]-scores[baseline.ID] < 0.14 {
		t.Fatalf("community boost too small: community %.3f baseline %.3f", scores[community.ID], scores[baseline.ID])
	}
	if scores[velocity.ID]-scores[baseline.ID] < 0.14 {
		t.Fatalf("velocity signal too small: velocity %.3f baseline %.3f", scores[velocity.ID], scores[baseline.ID])
	}

	if _, err := clipRepo.UpdateTrendingScores(ctx); err != nil {
		t.Fatalf("repeat mixed rank update: %v", err)
	}
	second, _, err := clipRepo.ListWithFilters(ctx, ClipFilters{Sort: "trending"}, 10, 0)
	if err != nil {
		t.Fatalf("repeat mixed rank list: %v", err)
	}
	for i := range order {
		if second[i].ID != order[i] {
			t.Fatalf("rank changed inside exploration window at %d: got %s want %s", i, second[i].ID, order[i])
		}
	}
}

func TestClipRepository_TrendingShuffleIsSeededAndAutomatedOnly(t *testing.T) {
	pool := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(t, pool)
	testutil.TruncateTables(t, pool, "clips", "users")

	ctx := context.Background()
	userRepo := NewUserRepository(pool)
	user := &models.User{ID: uuid.New(), TwitchID: testutil.StringPtr("shuffle-user-" + uuid.NewString()), Username: "shuffle-user", DisplayName: "Shuffle User"}
	if err := userRepo.Create(ctx, user); err != nil {
		t.Fatalf("create user: %v", err)
	}

	clipRepo := NewClipRepository(pool)
	createdAt := time.Now().Add(-6 * time.Hour)
	automatedIDs := make(map[uuid.UUID]struct{})
	for i := 0; i < 12; i++ {
		clip := &models.Clip{
			ID: uuid.New(), TwitchClipID: fmt.Sprintf("shuffle-auto-%d-%s", i, uuid.NewString()),
			TwitchClipURL: "https://clips.twitch.tv/shuffle-auto", EmbedURL: "https://clips.twitch.tv/embed?clip=shuffle-auto",
			Title: fmt.Sprintf("Automated %d", i), CreatorName: "creator", BroadcasterName: "broadcaster",
			ViewCount: 100, CreatedAt: createdAt, ImportedAt: time.Now(),
		}
		if err := clipRepo.Create(ctx, clip); err != nil {
			t.Fatalf("create automated clip %d: %v", i, err)
		}
		automatedIDs[clip.ID] = struct{}{}
	}
	community := &models.Clip{
		ID: uuid.New(), TwitchClipID: "shuffle-community-" + uuid.NewString(),
		TwitchClipURL: "https://clips.twitch.tv/shuffle-community", EmbedURL: "https://clips.twitch.tv/embed?clip=shuffle-community",
		Title: "Community", CreatorName: "creator", BroadcasterName: "broadcaster",
		ViewCount: 100, CreatedAt: createdAt, ImportedAt: time.Now(),
		SubmittedByUserID: &user.ID, SubmittedAt: testutil.TimePtr(time.Now()),
	}
	if err := clipRepo.Create(ctx, community); err != nil {
		t.Fatalf("create community clip: %v", err)
	}
	if _, err := clipRepo.UpdateTrendingScores(ctx); err != nil {
		t.Fatalf("update mixed ranks: %v", err)
	}

	seedA := "11111111-1111-4111-8111-111111111111"
	seedB := "22222222-2222-4222-8222-222222222222"
	list := func(seed string) []models.Clip {
		clips, _, err := clipRepo.ListWithFilters(ctx, ClipFilters{Sort: "trending", TrendingShuffleSeed: &seed}, 20, 0)
		if err != nil {
			t.Fatalf("list seed %s: %v", seed, err)
		}
		return clips
	}
	first := list(seedA)
	repeated := list(seedA)
	secondSeed := list(seedB)
	for i := range first {
		if repeated[i].ID != first[i].ID {
			t.Fatalf("same seed changed order at %d", i)
		}
	}
	autoOrder := func(clips []models.Clip) []uuid.UUID {
		ids := make([]uuid.UUID, 0, len(automatedIDs))
		for _, clip := range clips {
			if _, ok := automatedIDs[clip.ID]; ok {
				ids = append(ids, clip.ID)
			}
		}
		return ids
	}
	if reflect.DeepEqual(autoOrder(first), autoOrder(secondSeed)) {
		t.Fatal("different seeds produced the same automated clip order")
	}

	var rawCommunityScore float64
	if err := pool.QueryRow(ctx, `SELECT trending_score FROM clips WHERE id = $1`, community.ID).Scan(&rawCommunityScore); err != nil {
		t.Fatalf("read community score: %v", err)
	}
	for _, clips := range [][]models.Clip{first, secondSeed} {
		for _, clip := range clips {
			if clip.ID == community.ID && clip.TrendingScore != rawCommunityScore {
				t.Fatalf("community score was shuffled: got %.17g want %.17g", clip.TrendingScore, rawCommunityScore)
			}
		}
	}

	pageOne, _, err := clipRepo.ListWithFilters(ctx, ClipFilters{Sort: "trending", TrendingShuffleSeed: &seedA}, 6, 0)
	if err != nil {
		t.Fatalf("list first page: %v", err)
	}
	last := pageOne[len(pageOne)-1]
	cursor := utils.EncodeCursorWithShuffleSeed("trending", last.TrendingScore, last.ID, last.CreatedAt.Unix(), seedA)
	pageTwo, _, err := clipRepo.ListWithFilters(ctx, ClipFilters{Sort: "trending", TrendingShuffleSeed: &seedA, Cursor: &cursor}, 6, 0)
	if err != nil {
		t.Fatalf("list second page: %v", err)
	}
	seen := make(map[uuid.UUID]struct{}, len(pageOne))
	for _, clip := range pageOne {
		seen[clip.ID] = struct{}{}
	}
	for _, clip := range pageTwo {
		if _, duplicate := seen[clip.ID]; duplicate {
			t.Fatalf("clip %s repeated across shuffled pages", clip.ID)
		}
	}
}

func TestClipRepository_TopStreamers(t *testing.T) {
	// Skip if not in integration test mode
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	pool := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(t, pool)
	// Clean related tables before inserting fixtures
	testutil.TruncateTables(t, pool, "clips", "top_streamers", "clip_tags")

	repo := NewClipRepository(pool)
	ctx := context.Background()

	// Test UpsertTopStreamer
	err := repo.UpsertTopStreamer(ctx, "12345", "TopStreamer", 1, 1000000, 500000000)
	if err != nil {
		t.Fatalf("Failed to upsert top streamer: %v", err)
	}

	// Test IsTopStreamer
	isTop, err := repo.IsTopStreamer(ctx, "12345")
	if err != nil {
		t.Fatalf("Failed to check top streamer: %v", err)
	}
	if !isTop {
		t.Error("Expected broadcaster to be a top streamer")
	}

	// Test non-top streamer
	isTop, err = repo.IsTopStreamer(ctx, "99999")
	if err != nil {
		t.Fatalf("Failed to check non-top streamer: %v", err)
	}
	if isTop {
		t.Error("Expected broadcaster to not be a top streamer")
	}

	// Test GetTopStreamersCount
	count, err := repo.GetTopStreamersCount(ctx)
	if err != nil {
		t.Fatalf("Failed to get top streamers count: %v", err)
	}
	if count != 1 {
		t.Errorf("Expected 1 top streamer, got %d", count)
	}

	// Test ClearTopStreamers
	err = repo.ClearTopStreamers(ctx)
	if err != nil {
		t.Fatalf("Failed to clear top streamers: %v", err)
	}

	count, err = repo.GetTopStreamersCount(ctx)
	if err != nil {
		t.Fatalf("Failed to get count after clear: %v", err)
	}
	if count != 0 {
		t.Errorf("Expected 0 top streamers after clear, got %d", count)
	}
}

func TestClipRepository_ListWithFilters_Top10kStreamers(t *testing.T) {
	// Skip if not in integration test mode
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	pool := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(t, pool)
	// Clean related tables before inserting fixtures
	testutil.TruncateTables(t, pool, "clips", "top_streamers", "clip_tags")

	repo := NewClipRepository(pool)
	ctx := context.Background()

	// Add a top streamer
	err := repo.UpsertTopStreamer(ctx, "12345", "TopStreamer", 1, 1000000, 500000000)
	if err != nil {
		t.Fatalf("Failed to add top streamer: %v", err)
	}

	// Create clips - one from top streamer, one from regular streamer
	topClip := &models.Clip{
		ID:              uuid.New(),
		TwitchClipID:    fmt.Sprintf("top-clip-1-%s", uuid.NewString()),
		TwitchClipURL:   "https://clips.twitch.tv/top1",
		EmbedURL:        "https://clips.twitch.tv/embed?clip=top1",
		Title:           "Clip from Top Streamer",
		CreatorName:     "creator1",
		BroadcasterName: "TopStreamer",
		BroadcasterID:   testutil.StringPtr("12345"),
		CreatedAt:       time.Now().Add(-1 * time.Hour),
		ImportedAt:      time.Now(),
	}

	regularClip := &models.Clip{
		ID:              uuid.New(),
		TwitchClipID:    fmt.Sprintf("regular-clip-1-%s", uuid.NewString()),
		TwitchClipURL:   "https://clips.twitch.tv/regular1",
		EmbedURL:        "https://clips.twitch.tv/embed?clip=regular1",
		Title:           "Clip from Regular Streamer",
		CreatorName:     "creator2",
		BroadcasterName: "RegularStreamer",
		BroadcasterID:   testutil.StringPtr("99999"),
		CreatedAt:       time.Now().Add(-2 * time.Hour),
		ImportedAt:      time.Now(),
	}

	if err := repo.Create(ctx, topClip); err != nil {
		t.Fatalf("Failed to create top clip: %v", err)
	}
	if err := repo.Create(ctx, regularClip); err != nil {
		t.Fatalf("Failed to create regular clip: %v", err)
	}

	// Test filter with top 10k streamers enabled
	filters := ClipFilters{
		Sort:            "new",
		Top10kStreamers: true,
	}

	clips, total, err := repo.ListWithFilters(ctx, filters, 10, 0)
	if err != nil {
		t.Fatalf("Failed to list clips with top10k filter: %v", err)
	}

	// Should only return clip from top streamer
	if total < 1 {
		t.Errorf("Expected at least 1 clip from top streamers, got %d", total)
	}

	// Verify all returned clips are from top streamers
	for i, clip := range clips {
		if clip.BroadcasterID == nil {
			t.Errorf("Clip %d has nil BroadcasterID", i)
			continue
		}
		isTop, err := repo.IsTopStreamer(ctx, *clip.BroadcasterID)
		if err != nil {
			t.Fatalf("Failed to check streamer status: %v", err)
		}
		if !isTop {
			t.Errorf("Clip %d is not from a top streamer: %s", i, *clip.BroadcasterID)
		}
	}
}

func TestClipRepository_ListWithFilters_Language(t *testing.T) {
	// Skip if not in integration test mode
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	pool := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(t, pool)

	repo := NewClipRepository(pool)
	ctx := context.Background()

	// Create test clips with different languages
	englishClip := &models.Clip{
		ID:              uuid.New(),
		TwitchClipID:    fmt.Sprintf("en-clip-1-%s", uuid.NewString()),
		TwitchClipURL:   "https://clips.twitch.tv/en1",
		EmbedURL:        "https://clips.twitch.tv/embed?clip=en1",
		Title:           "English Clip",
		CreatorName:     "creator1",
		BroadcasterName: "EnglishStreamer",
		BroadcasterID:   testutil.StringPtr("11111"),
		Language:        testutil.StringPtr("en"),
		CreatedAt:       time.Now().Add(-1 * time.Hour),
		ImportedAt:      time.Now(),
	}

	spanishClip := &models.Clip{
		ID:              uuid.New(),
		TwitchClipID:    fmt.Sprintf("es-clip-1-%s", uuid.NewString()),
		TwitchClipURL:   "https://clips.twitch.tv/es1",
		EmbedURL:        "https://clips.twitch.tv/embed?clip=es1",
		Title:           "Spanish Clip",
		CreatorName:     "creator2",
		BroadcasterName: "SpanishStreamer",
		BroadcasterID:   testutil.StringPtr("22222"),
		Language:        testutil.StringPtr("es"),
		CreatedAt:       time.Now().Add(-2 * time.Hour),
		ImportedAt:      time.Now(),
	}

	frenchClip := &models.Clip{
		ID:              uuid.New(),
		TwitchClipID:    fmt.Sprintf("fr-clip-1-%s", uuid.NewString()),
		TwitchClipURL:   "https://clips.twitch.tv/fr1",
		EmbedURL:        "https://clips.twitch.tv/embed?clip=fr1",
		Title:           "French Clip",
		CreatorName:     "creator3",
		BroadcasterName: "FrenchStreamer",
		BroadcasterID:   testutil.StringPtr("33333"),
		Language:        testutil.StringPtr("fr"),
		CreatedAt:       time.Now().Add(-3 * time.Hour),
		ImportedAt:      time.Now(),
	}

	if err := repo.Create(ctx, englishClip); err != nil {
		t.Fatalf("Failed to create English clip: %v", err)
	}
	if err := repo.Create(ctx, spanishClip); err != nil {
		t.Fatalf("Failed to create Spanish clip: %v", err)
	}
	if err := repo.Create(ctx, frenchClip); err != nil {
		t.Fatalf("Failed to create French clip: %v", err)
	}

	// Test filter by English
	enLang := "en"
	filters := ClipFilters{
		Sort:     "new",
		Language: &enLang,
	}

	clips, total, err := repo.ListWithFilters(ctx, filters, 10, 0)
	if err != nil {
		t.Fatalf("Failed to list clips with language filter: %v", err)
	}

	if total != 1 {
		t.Errorf("Expected 1 English clip, got %d", total)
	}

	if len(clips) > 0 && clips[0].Language != nil && *clips[0].Language != "en" {
		t.Errorf("Expected English clip, got language: %s", *clips[0].Language)
	}

	// Test filter by Spanish
	esLang := "es"
	filters.Language = &esLang

	clips, total, err = repo.ListWithFilters(ctx, filters, 10, 0)
	if err != nil {
		t.Fatalf("Failed to list clips with Spanish filter: %v", err)
	}

	if total != 1 {
		t.Errorf("Expected 1 Spanish clip, got %d", total)
	}

	if len(clips) > 0 && clips[0].Language != nil && *clips[0].Language != "es" {
		t.Errorf("Expected Spanish clip, got language: %s", *clips[0].Language)
	}
}

func TestClipRepository_ListWithFilters_SubmittedByUserID(t *testing.T) {
	// Skip if not in integration test mode
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	pool := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(t, pool)
	// Ensure a clean slate
	testutil.TruncateTables(t, pool, "clips", "users")

	repo := NewClipRepository(pool)
	ctx := context.Background()

	// Create test users
	user1ID := uuid.New()
	insertTestUser(t, pool, user1ID)
	user2ID := uuid.New()
	insertTestUser(t, pool, user2ID)

	// Helper function to create unique clip IDs
	makeID := func(base string) string { return fmt.Sprintf("%s-%s", base, uuid.NewString()) }

	// Create clips submitted by user1
	clip1 := &models.Clip{
		ID:                uuid.New(),
		TwitchClipID:      makeID("test-clip-user1-1"),
		TwitchClipURL:     "https://clips.twitch.tv/user1-1",
		EmbedURL:          "https://clips.twitch.tv/embed?clip=user1-1",
		Title:             "User 1 Clip 1",
		CreatorName:       "creator1",
		BroadcasterName:   "broadcaster1",
		BroadcasterID:     testutil.StringPtr("12345"),
		CreatedAt:         time.Now().Add(-1 * time.Hour),
		ImportedAt:        time.Now(),
		SubmittedByUserID: &user1ID,
		SubmittedAt:       testutil.TimePtr(time.Now().Add(-1 * time.Hour)),
	}

	clip2 := &models.Clip{
		ID:                uuid.New(),
		TwitchClipID:      makeID("test-clip-user1-2"),
		TwitchClipURL:     "https://clips.twitch.tv/user1-2",
		EmbedURL:          "https://clips.twitch.tv/embed?clip=user1-2",
		Title:             "User 1 Clip 2",
		CreatorName:       "creator2",
		BroadcasterName:   "broadcaster2",
		BroadcasterID:     testutil.StringPtr("23456"),
		CreatedAt:         time.Now().Add(-2 * time.Hour),
		ImportedAt:        time.Now(),
		SubmittedByUserID: &user1ID,
		SubmittedAt:       testutil.TimePtr(time.Now().Add(-2 * time.Hour)),
	}

	// Create clip submitted by user2
	clip3 := &models.Clip{
		ID:                uuid.New(),
		TwitchClipID:      makeID("test-clip-user2-1"),
		TwitchClipURL:     "https://clips.twitch.tv/user2-1",
		EmbedURL:          "https://clips.twitch.tv/embed?clip=user2-1",
		Title:             "User 2 Clip 1",
		CreatorName:       "creator3",
		BroadcasterName:   "broadcaster3",
		BroadcasterID:     testutil.StringPtr("34567"),
		CreatedAt:         time.Now().Add(-3 * time.Hour),
		ImportedAt:        time.Now(),
		SubmittedByUserID: &user2ID,
		SubmittedAt:       testutil.TimePtr(time.Now().Add(-3 * time.Hour)),
	}

	// Create clip with no submitter (scraped clip)
	clip4 := &models.Clip{
		ID:              uuid.New(),
		TwitchClipID:    makeID("test-clip-scraped"),
		TwitchClipURL:   "https://clips.twitch.tv/scraped",
		EmbedURL:        "https://clips.twitch.tv/embed?clip=scraped",
		Title:           "Scraped Clip",
		CreatorName:     "creator4",
		BroadcasterName: "broadcaster4",
		BroadcasterID:   testutil.StringPtr("45678"),
		CreatedAt:       time.Now().Add(-4 * time.Hour),
		ImportedAt:      time.Now(),
	}

	// Insert clips
	if err := repo.Create(ctx, clip1); err != nil {
		t.Fatalf("Failed to create clip1: %v", err)
	}
	if err := repo.Create(ctx, clip2); err != nil {
		t.Fatalf("Failed to create clip2: %v", err)
	}
	if err := repo.Create(ctx, clip3); err != nil {
		t.Fatalf("Failed to create clip3: %v", err)
	}
	if err := repo.Create(ctx, clip4); err != nil {
		t.Fatalf("Failed to create clip4: %v", err)
	}

	// Test filtering by user1's submissions
	user1IDStr := user1ID.String()
	filters := ClipFilters{
		SubmittedByUserID: &user1IDStr,
		Sort:              "new",
	}

	clips, total, err := repo.ListWithFilters(ctx, filters, 10, 0)
	if err != nil {
		t.Fatalf("Failed to list clips for user1: %v", err)
	}

	if total != 2 {
		t.Errorf("Expected 2 clips submitted by user1, got %d", total)
	}

	if len(clips) != 2 {
		t.Errorf("Expected 2 clips in result, got %d", len(clips))
	}

	// Verify all returned clips belong to user1
	for _, clip := range clips {
		if clip.SubmittedByUserID == nil {
			t.Errorf("Clip %s has nil SubmittedByUserID", clip.ID)
			continue
		}
		if *clip.SubmittedByUserID != user1ID {
			t.Errorf("Clip %s submitted by %s, expected %s", clip.ID, *clip.SubmittedByUserID, user1ID)
		}
	}

	// Test filtering by user2's submissions
	user2IDStr := user2ID.String()
	filters.SubmittedByUserID = &user2IDStr

	clips, total, err = repo.ListWithFilters(ctx, filters, 10, 0)
	if err != nil {
		t.Fatalf("Failed to list clips for user2: %v", err)
	}

	if total != 1 {
		t.Errorf("Expected 1 clip submitted by user2, got %d", total)
	}

	if len(clips) != 1 {
		t.Errorf("Expected 1 clip in result, got %d", len(clips))
	}

	// Verify the clip belongs to user2
	if len(clips) > 0 {
		if clips[0].SubmittedByUserID == nil {
			t.Errorf("Clip %s has nil SubmittedByUserID", clips[0].ID)
		} else if *clips[0].SubmittedByUserID != user2ID {
			t.Errorf("Clip %s submitted by %s, expected %s", clips[0].ID, *clips[0].SubmittedByUserID, user2ID)
		}
	}

	// Test that scraped clips are not returned when filtering by user ID
	nonExistentUserID := uuid.New()
	nonExistentUserIDStr := nonExistentUserID.String()
	filters.SubmittedByUserID = &nonExistentUserIDStr

	clips, total, err = repo.ListWithFilters(ctx, filters, 10, 0)
	if err != nil {
		t.Fatalf("Failed to list clips for non-existent user: %v", err)
	}

	if total != 0 {
		t.Errorf("Expected 0 clips for non-existent user, got %d", total)
	}
}
