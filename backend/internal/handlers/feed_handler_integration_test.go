//go:build integration

package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"git.subcult.tv/subculture-collective/clpr/internal/models"
	"git.subcult.tv/subculture-collective/clpr/internal/middleware"
	"git.subcult.tv/subculture-collective/clpr/internal/repository"
	"git.subcult.tv/subculture-collective/clpr/internal/services"
	"git.subcult.tv/subculture-collective/clpr/internal/testutil"
	"git.subcult.tv/subculture-collective/clpr/internal/utils"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func TestMainFeedIncludesAutomatedTwitchClips(t *testing.T) {
	gin.SetMode(gin.TestMode)
	pool := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(t, pool)
	testutil.TruncateTables(t, pool, "clips", "votes", "favorites")

	clipRepo := repository.NewClipRepository(pool)
	voteRepo := repository.NewVoteRepository(pool)
	favoriteRepo := repository.NewFavoriteRepository(pool)
	feedService := services.NewFeedService(nil, clipRepo, nil, nil, voteRepo, favoriteRepo)
	handler := NewFeedHandler(feedService, nil, voteRepo, favoriteRepo, nil)

	automated := &models.Clip{
		ID:              uuid.New(),
		TwitchClipID:    "automated-main-feed-" + uuid.NewString(),
		TwitchClipURL:   "https://clips.twitch.tv/automated-main-feed",
		EmbedURL:        "https://clips.twitch.tv/embed?clip=automated-main-feed",
		Title:           "Automated clip appears in the main feed",
		CreatorName:     "creator",
		BroadcasterName: "broadcaster",
		BroadcasterID:   testutil.StringPtr("12345"),
		CreatedAt:       time.Now(),
		ImportedAt:      time.Now(),
	}
	if err := clipRepo.Create(context.Background(), automated); err != nil {
		t.Fatalf("create automated clip: %v", err)
	}

	router := gin.New()
	router.GET("/api/v1/feeds/clips", handler.GetFilteredClips)
	request := httptest.NewRequest(http.MethodGet, "/api/v1/feeds/clips?limit=20&sort=new", nil)
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200: %s", response.Code, response.Body.String())
	}
	var payload struct {
		Clips []models.Clip `json:"clips"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if len(payload.Clips) != 1 || payload.Clips[0].ID != automated.ID {
		t.Fatalf("main feed clips = %#v, want automated clip %s", payload.Clips, automated.ID)
	}
}

func TestMainFeedFiltersClipsByFlatTagParameter(t *testing.T) {
	gin.SetMode(gin.TestMode)
	pool := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(t, pool)
	testutil.TruncateTables(t, pool, "clips", "tags", "votes", "favorites")

	clipRepo := repository.NewClipRepository(pool)
	tagRepo := repository.NewTagRepository(pool)
	voteRepo := repository.NewVoteRepository(pool)
	favoriteRepo := repository.NewFavoriteRepository(pool)
	feedService := services.NewFeedService(nil, clipRepo, nil, nil, voteRepo, favoriteRepo)
	handler := NewFeedHandler(feedService, nil, voteRepo, favoriteRepo, nil)

	tag := &models.Tag{
		ID: uuid.New(), Name: "Highlights", Slug: "content/highlights", CreatedAt: time.Now(),
	}
	if err := tagRepo.Create(context.Background(), tag); err != nil {
		t.Fatalf("create tag: %v", err)
	}

	newClip := func(title string) *models.Clip {
		return &models.Clip{
			ID: uuid.New(), TwitchClipID: uuid.NewString(),
			TwitchClipURL: "https://clips.twitch.tv/" + title,
			EmbedURL: "https://clips.twitch.tv/embed?clip=" + title,
			Title: title, CreatorName: "creator", BroadcasterName: "broadcaster",
			CreatedAt: time.Now(), ImportedAt: time.Now(),
		}
	}
	tagged := newClip("tagged")
	untagged := newClip("untagged")
	for _, clip := range []*models.Clip{tagged, untagged} {
		if err := clipRepo.Create(context.Background(), clip); err != nil {
			t.Fatalf("create clip %s: %v", clip.Title, err)
		}
	}
	if err := clipRepo.AddTagBySlug(context.Background(), tagged.ID, tag.Slug); err != nil {
		t.Fatalf("tag clip: %v", err)
	}

	router := gin.New()
	router.Use(middleware.InputValidationMiddleware())
	router.GET("/api/v1/feeds/clips", handler.GetFilteredClips)
	request := httptest.NewRequest(http.MethodGet, "/api/v1/feeds/clips?limit=20&sort=new&tag=content%2Fhighlights", nil)
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200: %s", response.Code, response.Body.String())
	}
	var payload struct {
		Clips []models.Clip `json:"clips"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if len(payload.Clips) != 1 || payload.Clips[0].ID != tagged.ID {
		t.Fatalf("filtered clips = %#v, want only %s", payload.Clips, tagged.ID)
	}
}

func TestTrendingFeedRefreshesPersonalizedAutoShuffleAndKeepsCursorStable(t *testing.T) {
	gin.SetMode(gin.TestMode)
	pool := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(t, pool)
	testutil.TruncateTables(t, pool, "clips", "votes", "favorites")

	clipRepo := repository.NewClipRepository(pool)
	voteRepo := repository.NewVoteRepository(pool)
	favoriteRepo := repository.NewFavoriteRepository(pool)
	feedService := services.NewFeedService(nil, clipRepo, nil, nil, voteRepo, favoriteRepo)
	handler := NewFeedHandler(feedService, nil, voteRepo, favoriteRepo, nil)
	createdAt := time.Now().Add(-6 * time.Hour)
	for i := 0; i < 25; i++ {
		clip := &models.Clip{
			ID: uuid.New(), TwitchClipID: uuid.NewString(),
			TwitchClipURL: "https://clips.twitch.tv/personalized", EmbedURL: "https://clips.twitch.tv/embed?clip=personalized",
			Title: "Personalized automated clip", CreatorName: "creator", BroadcasterName: "broadcaster",
			ViewCount: 100, CreatedAt: createdAt, ImportedAt: time.Now(),
		}
		if err := clipRepo.Create(context.Background(), clip); err != nil {
			t.Fatalf("create automated clip %d: %v", i, err)
		}
	}
	if _, err := clipRepo.UpdateTrendingScores(context.Background()); err != nil {
		t.Fatalf("update trending scores: %v", err)
	}

	router := gin.New()
	viewerID := uuid.New()
	router.Use(func(c *gin.Context) {
		c.Set("user_id", viewerID)
		c.Next()
	})
	router.GET("/api/v1/feeds/clips", handler.GetFilteredClips)
	type feedPayload struct {
		Clips      []models.Clip `json:"clips"`
		Pagination struct {
			Cursor string `json:"cursor"`
		} `json:"pagination"`
	}
	requestFeed := func(path string) feedPayload {
		request := httptest.NewRequest(http.MethodGet, path, nil)
		response := httptest.NewRecorder()
		router.ServeHTTP(response, request)
		if response.Code != http.StatusOK {
			t.Fatalf("status = %d, want 200: %s", response.Code, response.Body.String())
		}
		var payload feedPayload
		if err := json.Unmarshal(response.Body.Bytes(), &payload); err != nil {
			t.Fatalf("decode response: %v", err)
		}
		return payload
	}

	first := requestFeed("/api/v1/feeds/clips?limit=10&sort=trending")
	refreshed := requestFeed("/api/v1/feeds/clips?limit=10&sort=trending")
	firstCursor, err := utils.DecodeCursor(first.Pagination.Cursor)
	if err != nil {
		t.Fatalf("decode first cursor: %v", err)
	}
	refreshedCursor, err := utils.DecodeCursor(refreshed.Pagination.Cursor)
	if err != nil {
		t.Fatalf("decode refreshed cursor: %v", err)
	}
	if firstCursor.ShuffleSeed == refreshedCursor.ShuffleSeed {
		t.Fatal("refresh reused the previous feed-session seed")
	}
	orderChanged := false
	for i := range first.Clips {
		if first.Clips[i].ID != refreshed.Clips[i].ID {
			orderChanged = true
			break
		}
	}
	if !orderChanged {
		t.Fatal("refresh produced the same automated clip order")
	}

	next := requestFeed("/api/v1/feeds/clips?limit=10&sort=trending&cursor=" + url.QueryEscape(first.Pagination.Cursor))
	seen := make(map[uuid.UUID]struct{}, len(first.Clips))
	for _, clip := range first.Clips {
		seen[clip.ID] = struct{}{}
	}
	for _, clip := range next.Clips {
		if _, duplicate := seen[clip.ID]; duplicate {
			t.Fatalf("clip %s repeated across cursor pages", clip.ID)
		}
	}
}
