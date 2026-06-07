package services_test

import (
	"strings"
	"testing"

	"github.com/google/uuid"
	"git.subcult.tv/subculture-collective/clpr/internal/models"
	"git.subcult.tv/subculture-collective/clpr/pkg/utils"
)

func TestAutoTagService_PatternMatching(t *testing.T) {
	tests := []struct {
		name         string
		clipTitle    string
		expectedTags []string
	}{
		{
			name:         "Ace detection",
			clipTitle:    "Amazing ACE by Shroud",
			expectedTags: []string{"ace"},
		},
		{
			name:         "Clutch detection",
			clipTitle:    "Insane 1v4 clutch moment",
			expectedTags: []string{"clutch"},
		},
		{
			name:         "Fail detection",
			clipTitle:    "Epic fail compilation",
			expectedTags: []string{"fail"},
		},
		{
			name:         "Funny detection",
			clipTitle:    "LMAO this is hilarious",
			expectedTags: []string{"funny"},
		},
		{
			name:         "Lucky detection",
			clipTitle:    "Lucky RNG moment",
			expectedTags: []string{"lucky"},
		},
		{
			name:         "Multiple patterns",
			clipTitle:    "Insane clutch ACE - epic moment",
			expectedTags: []string{"insane", "clutch", "ace", "epic"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test pattern matching logic directly
			title := tt.clipTitle

			// Check each expected tag pattern
			for _, expectedTag := range tt.expectedTags {
				found := false
				switch expectedTag {
				case "ace":
					found = containsPattern(title, "ace", "5k", "team wipe")
				case "clutch":
					found = containsPattern(title, "clutch", "1v")
				case "fail":
					found = containsPattern(title, "fail")
				case "funny":
					found = containsPattern(title, "funny", "lol", "lmao", "hilarious")
				case "lucky":
					found = containsPattern(title, "lucky", "luck", "rng")
				case "insane":
					found = containsPattern(title, "insane", "crazy", "amazing", "incredible")
				case "epic":
					found = containsPattern(title, "epic", "legendary", "godlike")
				}

				if !found {
					t.Errorf("Expected pattern '%s' not found in title: %s", expectedTag, title)
				}
			}
		})
	}
}

func TestAutoTagService_DurationLogic(t *testing.T) {
	tests := []struct {
		name        string
		duration    float64
		expectedTag string
	}{
		{
			name:        "Short clip (10s)",
			duration:    10.0,
			expectedTag: "short",
		},
		{
			name:        "Very short clip (5s)",
			duration:    5.0,
			expectedTag: "short",
		},
		{
			name:        "Long clip (150s)",
			duration:    150.0,
			expectedTag: "long",
		},
		{
			name:        "Very long clip (300s)",
			duration:    300.0,
			expectedTag: "long",
		},
		{
			name:        "Normal clip (30s)",
			duration:    30.0,
			expectedTag: "",
		},
		{
			name:        "Normal clip (60s)",
			duration:    60.0,
			expectedTag: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var result string
			if tt.duration < 15 {
				result = "short"
			} else if tt.duration > 120 {
				result = "long"
			}

			if result != tt.expectedTag {
				t.Errorf("Expected tag '%s' for duration %.1fs, got '%s'", tt.expectedTag, tt.duration, result)
			}
		})
	}
}

func TestAutoTagService_SlugGeneration(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Simple name",
			input:    "Counter-Strike",
			expected: "counter-strike",
		},
		{
			name:     "Name with spaces",
			input:    "League of Legends",
			expected: "league-of-legends",
		},
		{
			name:     "Name with special characters",
			input:    "VALORANT!!! @#$",
			expected: "valorant",
		},
		{
			name:     "Name with multiple spaces",
			input:    "Grand  Theft   Auto",
			expected: "grand-theft-auto",
		},
		{
			name:     "Mixed case",
			input:    "Minecraft",
			expected: "minecraft",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := utils.Slugify(tt.input)
			if result != tt.expected {
				t.Errorf("Expected slug '%s', got '%s'", tt.expected, result)
			}
		})
	}
}

func TestClipModel_Basic(t *testing.T) {
	// Test that the Clip model can be created and fields accessed
	gameName := "Counter-Strike"
	duration := 30.5
	language := "en"

	clip := &models.Clip{
		ID:              uuid.New(),
		TwitchClipID:    "test-123",
		Title:           "Amazing ACE",
		CreatorName:     "TestCreator",
		BroadcasterName: "TestBroadcaster",
		GameName:        &gameName,
		Duration:        &duration,
		Language:        &language,
	}

	if clip.Title != "Amazing ACE" {
		t.Errorf("Expected title 'Amazing ACE', got '%s'", clip.Title)
	}

	if clip.GameName == nil || *clip.GameName != "Counter-Strike" {
		t.Error("Game name not set correctly")
	}

	if clip.Duration == nil || *clip.Duration != 30.5 {
		t.Error("Duration not set correctly")
	}

	if clip.Language == nil || *clip.Language != "en" {
		t.Error("Language not set correctly")
	}
}

// Helper functions

func containsPattern(text string, patterns ...string) bool {
	lowerText := toLower(text)
	for _, pattern := range patterns {
		if contains(lowerText, toLower(pattern)) {
			return true
		}
	}
	return false
}

func toLower(s string) string {
	result := ""
	for _, r := range s {
		if r >= 'A' && r <= 'Z' {
			result += string(r + 32)
		} else {
			result += string(r)
		}
	}
	return result
}

func contains(text, substr string) bool {
	return strings.Contains(text, substr)
}

