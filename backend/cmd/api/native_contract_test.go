package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"git.subcult.tv/subculture-collective/clpr/internal/handlers"
	"git.subcult.tv/subculture-collective/clpr/internal/models"
	"git.subcult.tv/subculture-collective/clpr/internal/search"
	"git.subcult.tv/subculture-collective/clpr/internal/services"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type nativeFixture struct {
	File       string   `json:"file"`
	Method     string   `json:"method"`
	Path       string   `json:"path"`
	Status     int      `json:"status"`
	Provenance string   `json:"provenance"`
	Sources    []string `json:"sources"`
	Body       any      `json:"-"`
}

type nativeSearchProvider struct{ response *models.SearchResponse }

func (p nativeSearchProvider) Search(context.Context, *models.SearchRequest) (*models.SearchResponse, error) {
	return p.response, nil
}
func (p nativeSearchProvider) SearchWithScores(context.Context, *models.SearchRequest) (*models.SearchResponseWithScores, error) {
	return nil, errors.New("not used")
}

type nativeCountdownStore struct {
	editions []models.PublicCountdownEdition
	next     *string
}

func (p nativeCountdownStore) ListPublicEditions(context.Context, *string, *string, int) ([]models.PublicCountdownEdition, *string, error) {
	return p.editions, p.next, nil
}
func (p nativeCountdownStore) GetPublicEdition(context.Context, string, string) (*models.PublicCountdownEdition, error) {
	return nil, nil
}
func (p nativeCountdownStore) ListCountdownFeatures(context.Context, *uuid.UUID, *string, int) ([]models.CountdownFeature, error) {
	return []models.CountdownFeature{}, nil
}

// These responses use owned values and no network, credentials or production rows.
// Handler captures use stub providers; constructed envelopes deliberately say so.
func nativeFixtures(t *testing.T) []nativeFixture {
	t.Helper()
	gin.SetMode(gin.TestMode)
	date := time.Date(2026, 11, 16, 12, 0, 0, 123000000, time.UTC)
	id := uuid.MustParse("00000000-0000-4000-8000-000000000001")
	owner := uuid.MustParse("00000000-0000-4000-8000-000000000002")
	category, name, snippet := "fixture-category", "Fixture category", "A \ue000fixture\ue001 phrase"
	clip := models.Clip{ID: id, TwitchClipID: "CLPRNativeFixture", TwitchClipURL: "https://clips.twitch.tv/CLPRNativeFixture", EmbedURL: "https://clips.twitch.tv/embed?clip=CLPRNativeFixture", Title: "Native contract fixture", CreatorName: "Fixture creator", BroadcasterName: "Fixture channel", GameID: &category, GameName: &name, ViewCount: 42, VoteScore: 3, CreatedAt: date, ImportedAt: date, SourceType: "twitch", SourcePlatform: "twitch"}
	userClip := services.ClipWithUserData{Clip: clip}
	meta := handlers.PaginationMeta{Page: 1, Limit: 25, Total: 1, TotalPages: 1}
	emptyMeta := handlers.PaginationMeta{Page: 1, Limit: 25}
	tag := models.Tag{ID: owner, Name: "Content: Funny", Slug: "content/funny", UsageCount: 1, CreatedAt: date}
	playlist := models.PlaylistWithClips{Playlist: models.Playlist{ID: owner, UserID: owner, Title: "Public fixture playlist", Visibility: "public", CreatedAt: date, UpdatedAt: date}, ClipCount: 1, Clips: []models.PlaylistClipRef{{Clip: clip, OrderIndex: 0}}}
	edition := models.PublicCountdownEdition{EditionDate: "2026-11-16", SeriesSlug: "overall", SeriesName: "Overall", Size: 10, Clips: []models.PublicCountdownClip{{Position: 2, ClipID: id, Title: clip.Title, BroadcasterName: clip.BroadcasterName, TwitchClipID: clip.TwitchClipID, TwitchClipURL: clip.TwitchClipURL}}}
	var fixtures []nativeFixture
	add := func(file, path string, status int, body any, sources ...string) {
		fixtures = append(fixtures, nativeFixture{File: file + ".json", Method: "GET", Path: path, Status: status, Provenance: "constructed-from-wire-types", Sources: sources, Body: body})
	}
	capture := func(file, path, target string, status int, handler gin.HandlerFunc, sources ...string) {
		router := gin.New()
		router.GET(path, handler)
		response := httptest.NewRecorder()
		router.ServeHTTP(response, httptest.NewRequest(http.MethodGet, target, nil))
		if response.Code != status {
			t.Fatalf("%s: status %d, want %d: %s", file, response.Code, status, response.Body.String())
		}
		var body any
		if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil {
			t.Fatal(err)
		}
		add(file, ginParameter.ReplaceAllString(path, `{$1}`), status, body, sources...)
		fixtures[len(fixtures)-1].Provenance = "handler-with-owned-stub"
	}
	add("clips-populated", "/api/v1/clips", 200, handlers.StandardResponse{Success: true, Data: []services.ClipWithUserData{userClip}, Meta: meta}, "backend/internal/handlers/clip_handler.go#ListClips", "backend/internal/services/clip_service.go#ClipWithUserData")
	add("clips-empty", "/api/v1/clips", 200, handlers.StandardResponse{Success: true, Data: []services.ClipWithUserData{}, Meta: emptyMeta}, "backend/internal/handlers/clip_handler.go#ListClips")
	add("clip-detail", "/api/v1/clips/{id}", 200, handlers.StandardResponse{Success: true, Data: userClip}, "backend/internal/handlers/clip_handler.go#GetClip")
	capture("clips-invalid-query", "/api/v1/clips", "/api/v1/clips?submitted_by_user_id=not-a-uuid", 400, handlers.NewClipHandler(nil, nil).ListClips, "backend/internal/handlers/clip_handler.go#ListClips")
	add("clip-removed", "/api/v1/clips/{id}", 404, handlers.StandardResponse{Success: false, Error: &handlers.ErrorInfo{Code: "CLIP_NOT_FOUND", Message: "Clip not found or has been removed"}}, "backend/internal/handlers/clip_handler.go#GetClip")
	fixtures[len(fixtures)-1].Provenance = "constructed-negative-case"
	searchResponse := &models.SearchResponse{Query: "fixture", Results: models.EmptySearchResults(), Meta: models.SearchMeta{Page: 1, Limit: 20}}
	capture("search-empty", "/api/v1/search", "/api/v1/search?q=fixture", 200, search.NewHandlerWithHybridProvider(nil, nativeSearchProvider{searchResponse}).Search, "backend/internal/search/handler.go#Search")
	clip.TranscriptSnippet = &snippet
	searchResponse.Results.Clips = []models.Clip{clip}
	searchResponse.Results.Games = []models.GameSearchResult{{ID: category, Name: name, ClipCount: 1}}
	searchResponse.Results.TwitchCategories = searchResponse.Results.Games
	searchResponse.Results.Tags = []models.Tag{tag}
	searchResponse.Counts = models.SearchCounts{Clips: 1, Games: 1, TwitchCategories: 1, Tags: 1}
	searchResponse.Meta.TotalItems = 1
	searchResponse.Meta.TotalPages = 1
	capture("search-populated", "/api/v1/search", "/api/v1/search?q=fixture&said=fixture", 200, search.NewHandlerWithHybridProvider(nil, nativeSearchProvider{searchResponse}).Search, "backend/internal/search/handler.go#Search", "backend/internal/models/models.go#Clip")
	capture("search-invalid", "/api/v1/search", "/api/v1/search", 400, search.NewHandlerWithHybridProvider(nil, nativeSearchProvider{}).Search, "backend/internal/search/handler.go#Search")
	add("search-rate-limited", "/api/v1/search", 429, gin.H{"error": "Rate limit exceeded. Please try again later.", "retry_after": 60}, "backend/internal/middleware/ratelimit_middleware.go#RateLimitMiddleware")
	fixtures[len(fixtures)-1].Provenance = "constructed-negative-case"
	add("tags-populated", "/api/v1/tags", 200, gin.H{"tags": []*models.Tag{&tag}, "total": 1, "page": 1, "limit": 50, "has_more": false}, "backend/internal/handlers/tag_handler.go#ListTags", "backend/internal/models/tag_json.go#MarshalJSON")
	add("tags-empty", "/api/v1/tags", 200, gin.H{"tags": ([]*models.Tag)(nil), "total": 0, "page": 1, "limit": 50, "has_more": false}, "backend/internal/handlers/tag_handler.go#ListTags", "backend/internal/repository/tag_repository.go#List")
	add("categories-populated", "/api/v1/games/trending", 200, gin.H{"games": []*models.TrendingGame{{ID: owner, TwitchGameID: category, Name: name, RecentClipCount: 1}}, "page": 1, "limit": 20, "has_more": false}, "backend/internal/handlers/game_handler.go#GetTrendingGames")
	add("categories-empty", "/api/v1/games/trending", 200, gin.H{"games": ([]*models.TrendingGame)(nil), "page": 1, "limit": 20, "has_more": false}, "backend/internal/handlers/game_handler.go#GetTrendingGames", "backend/internal/repository/game_repository.go#GetTrending")
	add("clip-topics", "/api/v1/clips/{id}/topics", 200, gin.H{"topics": []models.ClipTopic{{ClipID: id, TopicID: owner, TopicName: "Fixture topic", TopicSlug: "fixture-topic", Source: "manual", Confidence: 1, Evidence: json.RawMessage(`{}`), CreatedAt: date, UpdatedAt: date}}}, "backend/internal/handlers/topic_handler.go#ListClipTopics")
	add("clip-topics-empty", "/api/v1/clips/{id}/topics", 200, gin.H{"topics": []models.ClipTopic{}}, "backend/internal/handlers/topic_handler.go#ListClipTopics")
	rails := models.CreatorDiscoveryRails{Trending: []models.CreatorDiscoveryProfile{{BroadcasterID: "fixture-channel", BroadcasterName: "Fixture channel", TotalClips: 1, FirstDiscoveredAt: date, LatestClipAt: date}}, Rising: []models.CreatorDiscoveryProfile{}, New: []models.CreatorDiscoveryProfile{}}
	add("creators-populated", "/api/v1/broadcasters/discover", 200, handlers.StandardResponse{Success: true, Data: rails}, "backend/internal/handlers/creator_discovery_handler.go#ListCreatorDiscovery")
	rails.Trending = []models.CreatorDiscoveryProfile{}
	add("creators-empty", "/api/v1/broadcasters/discover", 200, handlers.StandardResponse{Success: true, Data: rails}, "backend/internal/handlers/creator_discovery_handler.go#ListCreatorDiscovery")
	previous := "2026-11-15"
	capture("countdowns-populated", "/api/v1/countdowns", "/api/v1/countdowns?days=1", 200, handlers.NewCountdownPublicHandler(nativeCountdownStore{[]models.PublicCountdownEdition{edition}, &previous}).ListEditions, "backend/internal/handlers/countdown_public_handler.go#ListEditions")
	capture("countdowns-empty", "/api/v1/countdowns", "/api/v1/countdowns", 200, handlers.NewCountdownPublicHandler(nativeCountdownStore{[]models.PublicCountdownEdition{}, nil}).ListEditions, "backend/internal/handlers/countdown_public_handler.go#ListEditions")
	add("countdown-detail", "/api/v1/countdowns/{date}/{series}", 200, handlers.StandardResponse{Success: true, Data: edition}, "backend/internal/handlers/countdown_public_handler.go#GetEdition")
	capture("countdown-unavailable", "/api/v1/countdowns/:date/:series", "/api/v1/countdowns/2026-11-16/overall", 404, handlers.NewCountdownPublicHandler(nativeCountdownStore{}).GetEdition, "backend/internal/handlers/countdown_public_handler.go#GetEdition")
	add("playlist-populated", "/api/v1/playlists/{id}", 200, handlers.StandardResponse{Success: true, Data: playlist, Meta: meta}, "backend/internal/handlers/playlist_handler.go#GetPlaylist", "backend/internal/models/models.go#PlaylistWithClips")
	playlist.Clips = nil
	playlist.ClipCount = 0
	add("playlist-empty", "/api/v1/playlists/{id}", 200, handlers.StandardResponse{Success: true, Data: playlist, Meta: emptyMeta}, "backend/internal/handlers/playlist_handler.go#GetPlaylist")
	add("playlist-private-or-missing", "/api/v1/playlists/{id}", 404, handlers.StandardResponse{Success: false, Error: &handlers.ErrorInfo{Code: "NOT_FOUND", Message: "Playlist not found"}}, "backend/internal/handlers/playlist_handler.go#GetPlaylist")
	fixtures[len(fixtures)-1].Provenance = "constructed-negative-case"
	return fixtures
}

func TestNativeContractFixtures(t *testing.T) {
	const root = "../../../docs/openapi/mobile"
	fixtures := nativeFixtures(t)
	routes := map[string]bool{}
	for _, r := range supportedRouter().Routes() {
		routes[r.Method+" "+ginParameter.ReplaceAllString(r.Path, `{$1}`)] = true
	}
	check := func(name string, value any) {
		t.Helper()
		data, err := json.MarshalIndent(value, "", "  ")
		if err != nil {
			t.Fatal(err)
		}
		data = append(data, '\n')
		path := filepath.Join(root, name)
		if os.Getenv("UPDATE_NATIVE_FIXTURES") == "1" {
			if err := os.MkdirAll(root, 0755); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(path, data, 0644); err != nil {
				t.Fatal(err)
			}
		}
		actual, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(actual, data) {
			t.Errorf("%s drifted from its Go response: run UPDATE_NATIVE_FIXTURES=1 go test ./cmd/api -run TestNativeContractFixtures and review the diff", name)
		}
	}
	expected := map[string]bool{"manifest.json": true}
	for _, f := range fixtures {
		if !routes[f.Method+" "+f.Path] {
			t.Errorf("fixture references unregistered route: %s %s", f.Method, f.Path)
		}
		if expected[f.File] {
			t.Fatalf("duplicate fixture: %s", f.File)
		}
		expected[f.File] = true
		for _, source := range f.Sources {
			if _, err := os.Stat(filepath.Join("../../..", strings.Split(source, "#")[0])); err != nil {
				t.Fatal(err)
			}
		}
		check(f.File, f.Body)
	}
	check("manifest.json", fixtures)
	files, err := filepath.Glob(filepath.Join(root, "*.json"))
	if err != nil {
		t.Fatal(err)
	}
	for _, file := range files {
		if !expected[filepath.Base(file)] {
			t.Errorf("unowned fixture: %s", file)
		}
	}
}
