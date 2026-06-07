package services

import (
	"context"
	"testing"
	"time"

	"git.subcult.tv/subculture-collective/clpr/internal/utils"
	"git.subcult.tv/subculture-collective/clpr/pkg/twitch"
)

func TestExtractClipID(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Direct ID",
			input:    "AwkwardHelplessSalamanderSwiftRage",
			expected: "AwkwardHelplessSalamanderSwiftRage",
		},
		{
			name:     "Full URL",
			input:    "https://www.twitch.tv/username/clip/AwkwardHelplessSalamanderSwiftRage",
			expected: "AwkwardHelplessSalamanderSwiftRage",
		},
		{
			name:     "Clips subdomain URL",
			input:    "https://clips.twitch.tv/AwkwardHelplessSalamanderSwiftRage",
			expected: "AwkwardHelplessSalamanderSwiftRage",
		},
		{
			name:     "URL with query params",
			input:    "https://clips.twitch.tv/AwkwardHelplessSalamanderSwiftRage?filter=clips",
			expected: "AwkwardHelplessSalamanderSwiftRage",
		},
		{
			name:     "Empty string",
			input:    "",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ExtractClipID(tt.input)
			if result != tt.expected {
				t.Errorf("ExtractClipID(%s) = %s, want %s", tt.input, result, tt.expected)
			}
		})
	}
}

func TestTransformTwitchClip(t *testing.T) {
	now := time.Now()
	twitchClip := &twitch.Clip{
		ID:              "test-clip-id",
		URL:             "https://clips.twitch.tv/test-clip-id",
		EmbedURL:        "https://clips.twitch.tv/embed?clip=test-clip-id",
		BroadcasterID:   "broadcaster-123",
		BroadcasterName: "TestStreamer",
		CreatorID:       "creator-456",
		CreatorName:     "TestClipCreator",
		GameID:          "game-789",
		Language:        "en",
		Title:           "Amazing Play",
		ViewCount:       1000,
		CreatedAt:       now,
		ThumbnailURL:    "https://example.com/thumb.jpg",
		Duration:        30.5,
	}

	clip := transformTwitchClip(twitchClip)

	if clip.TwitchClipID != twitchClip.ID {
		t.Errorf("Expected TwitchClipID %s, got %s", twitchClip.ID, clip.TwitchClipID)
	}
	if clip.TwitchClipURL != twitchClip.URL {
		t.Errorf("Expected TwitchClipURL %s, got %s", twitchClip.URL, clip.TwitchClipURL)
	}
	if clip.Title != twitchClip.Title {
		t.Errorf("Expected Title %s, got %s", twitchClip.Title, clip.Title)
	}
	if clip.ViewCount != twitchClip.ViewCount {
		t.Errorf("Expected ViewCount %d, got %d", twitchClip.ViewCount, clip.ViewCount)
	}
	if clip.CreatedAt != twitchClip.CreatedAt {
		t.Errorf("Expected CreatedAt %v, got %v", twitchClip.CreatedAt, clip.CreatedAt)
	}
	if clip.BroadcasterName != twitchClip.BroadcasterName {
		t.Errorf("Expected BroadcasterName %s, got %s", twitchClip.BroadcasterName, clip.BroadcasterName)
	}
	if clip.CreatorName != twitchClip.CreatorName {
		t.Errorf("Expected CreatorName %s, got %s", twitchClip.CreatorName, clip.CreatorName)
	}

	// Check pointer fields
	if clip.BroadcasterID == nil || *clip.BroadcasterID != twitchClip.BroadcasterID {
		t.Errorf("Expected BroadcasterID %s, got %v", twitchClip.BroadcasterID, clip.BroadcasterID)
	}
	if clip.CreatorID == nil || *clip.CreatorID != twitchClip.CreatorID {
		t.Errorf("Expected CreatorID %s, got %v", twitchClip.CreatorID, clip.CreatorID)
	}
	if clip.GameID == nil || *clip.GameID != twitchClip.GameID {
		t.Errorf("Expected GameID %s, got %v", twitchClip.GameID, clip.GameID)
	}
	if clip.Language == nil || *clip.Language != twitchClip.Language {
		t.Errorf("Expected Language %s, got %v", twitchClip.Language, clip.Language)
	}
	if clip.Duration == nil || *clip.Duration != twitchClip.Duration {
		t.Errorf("Expected Duration %f, got %v", twitchClip.Duration, clip.Duration)
	}

	// Check default values
	if clip.VoteScore != 0 {
		t.Errorf("Expected VoteScore 0, got %d", clip.VoteScore)
	}
	if clip.CommentCount != 0 {
		t.Errorf("Expected CommentCount 0, got %d", clip.CommentCount)
	}
	if clip.IsFeatured != false {
		t.Errorf("Expected IsFeatured false, got %t", clip.IsFeatured)
	}
	if clip.IsNSFW != false {
		t.Errorf("Expected IsNSFW false, got %t", clip.IsNSFW)
	}
	if clip.IsRemoved != false {
		t.Errorf("Expected IsRemoved false, got %t", clip.IsRemoved)
	}
}

func TestLanguageMatches(t *testing.T) {
	tests := []struct {
		name    string
		clip    string
		filter  string
		expects bool
	}{
		{"empty filter allows all", "en", "", true},
		{"all keyword allows all", "fr", "all", true},
		{"exact match", "en", "en", true},
		{"case insensitive", "EN", "en", true},
		{"prefix match", "en-us", "en", true},
		{"non match", "fr", "en", false},
		{"empty clip lang", "", "en", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := languageMatches(tt.clip, tt.filter); got != tt.expects {
				t.Fatalf("languageMatches(%q, %q) = %v, want %v", tt.clip, tt.filter, got, tt.expects)
			}
		})
	}
}

func TestStringPtr(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected *string
	}{
		{
			name:     "Non-empty string",
			input:    "test",
			expected: utils.StringPtr("test"),
		},
		{
			name:     "Empty string",
			input:    "",
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := utils.StringPtr(tt.input)
			if tt.expected == nil {
				if result != nil {
					t.Errorf("Expected nil, got %v", result)
				}
			} else {
				if result == nil || *result != *tt.expected {
					t.Errorf("Expected %v, got %v", tt.expected, result)
				}
			}
		})
	}
}

func TestFloat64Ptr(t *testing.T) {
	tests := []struct {
		name     string
		input    float64
		expected *float64
	}{
		{
			name:     "Non-zero value",
			input:    30.5,
			expected: utils.Float64Ptr(30.5),
		},
		{
			name:     "Zero value",
			input:    0,
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := utils.Float64Ptr(tt.input)
			if tt.expected == nil {
				if result != nil {
					t.Errorf("Expected nil, got %v", result)
				}
			} else {
				if result == nil || *result != *tt.expected {
					t.Errorf("Expected %v, got %v", tt.expected, result)
				}
			}
		})
	}
}

func TestMin(t *testing.T) {
	tests := []struct {
		name     string
		a        int
		b        int
		expected int
	}{
		{
			name:     "a is smaller",
			a:        5,
			b:        10,
			expected: 5,
		},
		{
			name:     "b is smaller",
			a:        10,
			b:        5,
			expected: 5,
		},
		{
			name:     "equal values",
			a:        5,
			b:        5,
			expected: 5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := utils.Min(tt.a, tt.b)
			if result != tt.expected {
				t.Errorf("utils.Min(%d, %d) = %d, want %d", tt.a, tt.b, result, tt.expected)
			}
		})
	}
}

type mockTrendingStateStore struct {
	cursorState map[string]*TrendingCursorState
	gameIDs     []string
}

func newMockTrendingStateStore(gameIDs []string) *mockTrendingStateStore {
	return &mockTrendingStateStore{cursorState: make(map[string]*TrendingCursorState), gameIDs: gameIDs}
}

func (m *mockTrendingStateStore) GetCursor(_ context.Context, gameID string) (*TrendingCursorState, error) {
	return m.cursorState[gameID], nil
}

func (m *mockTrendingStateStore) SaveCursor(_ context.Context, gameID string, state *TrendingCursorState) error {
	m.cursorState[gameID] = state
	return nil
}

func (m *mockTrendingStateStore) ClearCursor(_ context.Context, gameID string) error {
	delete(m.cursorState, gameID)
	return nil
}

func (m *mockTrendingStateStore) SaveGameIDs(_ context.Context, gameIDs []string) error {
	m.gameIDs = gameIDs
	return nil
}

func (m *mockTrendingStateStore) LoadGameIDs(_ context.Context) ([]string, error) {
	return m.gameIDs, nil
}

func TestEnsureJustChattingGameIDs(t *testing.T) {
	ids := []string{"111", justChattingGameID, "222", "111"}
	result := ensureJustChattingGameIDs(ids)

	if len(result) == 0 || result[0] != justChattingGameID {
		t.Fatalf("expected Just Chatting (%s) to be first, got %v", justChattingGameID, result)
	}

	seen := map[string]bool{}
	for _, id := range result {
		if seen[id] {
			t.Fatalf("duplicate id %s in result %v", id, result)
		}
		seen[id] = true
	}

	if len(result) > maxTrendingGames {
		t.Fatalf("expected result to be trimmed to %d ids, got %d", maxTrendingGames, len(result))
	}
}

func TestBuildTrendingGameConfigs(t *testing.T) {
	configs := buildTrendingGameConfigs([]string{justChattingGameID, "111"})
	if len(configs) != 2 {
		t.Fatalf("expected 2 configs, got %d", len(configs))
	}

	if configs[0].Limit != justChattingPerGameLimit {
		t.Fatalf("expected Just Chatting limit %d, got %d", justChattingPerGameLimit, configs[0].Limit)
	}
	if configs[1].Limit != defaultPerGameLimit {
		t.Fatalf("expected default per-game limit %d, got %d", defaultPerGameLimit, configs[1].Limit)
	}
}

func TestApplyTrendingDefaults(t *testing.T) {
	store := newMockTrendingStateStore(nil)
	svc := &ClipSyncService{maxPages: 5, stateStore: store}

	resolved := svc.applyTrendingDefaults(nil)
	if resolved.MaxPages != 5 {
		t.Fatalf("expected max pages 5 from service defaults, got %d", resolved.MaxPages)
	}
	if resolved.StateStore != store {
		t.Fatal("expected state store to default to service store")
	}

	customStore := newMockTrendingStateStore(nil)
	opts := &TrendingSyncOptions{MaxPages: 2, StateStore: customStore, ForceResetPagination: true, Games: []TrendingGameConfig{{GameID: "123", Limit: 5}}}
	resolved = svc.applyTrendingDefaults(opts)
	if resolved.MaxPages != 2 {
		t.Fatalf("expected max pages 2 from options, got %d", resolved.MaxPages)
	}
	if resolved.StateStore != customStore {
		t.Fatal("expected state store override from options")
	}
	if !resolved.ForceResetPagination {
		t.Fatal("expected force reset pagination to be true")
	}
	if len(resolved.Games) != 1 || resolved.Games[0].GameID != "123" {
		t.Fatalf("expected custom games to be copied, got %+v", resolved.Games)
	}
}

func TestResolveTrendingGamesUsesCachedList(t *testing.T) {
	cached := []string{"999", "888"}
	store := newMockTrendingStateStore(cached)
	svc := &ClipSyncService{stateStore: store}

	configs, err := svc.resolveTrendingGames(context.Background(), store)
	if err != nil {
		t.Fatalf("expected no error when using cached games, got %v", err)
	}

	if len(configs) < 2 {
		t.Fatalf("expected cached games to be used, got %d configs", len(configs))
	}

	if configs[0].GameID != justChattingGameID {
		t.Fatalf("expected Just Chatting to be first, got %s", configs[0].GameID)
	}
}
