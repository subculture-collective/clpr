package services

import (
	"strings"
	"testing"
	"time"

	"git.subcult.tv/subculture-collective/clpr/config"
	"git.subcult.tv/subculture-collective/clpr/pkg/twitch"
)

// TestValidateSubmissionInput tests the comprehensive input validation
func TestValidateSubmissionInput(t *testing.T) {
	service := &SubmissionService{}

	t.Run("Valid submission with all fields", func(t *testing.T) {
		customTitle := "Amazing Gaming Moment"
		broadcasterOverride := "StreamerName"
		reason := "This is an epic moment"
		req := &SubmitClipRequest{
			ClipURL:                 "https://clips.twitch.tv/AwkwardHelplessSalamanderSwiftRage",
			CustomTitle:             &customTitle,
			BroadcasterNameOverride: &broadcasterOverride,
			Tags:                    []string{"epic", "gaming", "highlight"},
			IsNSFW:                  false,
			SubmissionReason:        &reason,
		}

		err := service.validateSubmissionInput(req)
		if err != nil {
			t.Errorf("Valid submission should not return error, got: %v", err)
		}
	})

	t.Run("Valid submission with minimal fields", func(t *testing.T) {
		req := &SubmitClipRequest{
			ClipURL: "https://clips.twitch.tv/AwkwardHelplessSalamanderSwiftRage",
			IsNSFW:  false,
		}

		err := service.validateSubmissionInput(req)
		if err != nil {
			t.Errorf("Minimal valid submission should not return error, got: %v", err)
		}
	})

	t.Run("Clip URL too long", func(t *testing.T) {
		longURL := "https://clips.twitch.tv/" + string(make([]byte, 500))
		req := &SubmitClipRequest{
			ClipURL: longURL,
		}

		err := service.validateSubmissionInput(req)
		if err == nil {
			t.Error("Expected error for URL too long")
		}
		if valErr, ok := err.(*ValidationError); ok {
			if valErr.Field != "clip_url" {
				t.Errorf("Expected field 'clip_url', got '%s'", valErr.Field)
			}
		}
	})

	t.Run("Custom title normalization", func(t *testing.T) {
		customTitle := "  Spaced Title  "
		req := &SubmitClipRequest{
			ClipURL:     "https://clips.twitch.tv/Test",
			CustomTitle: &customTitle,
		}

		err := service.validateSubmissionInput(req)
		if err != nil {
			t.Errorf("Valid title with spaces should not error: %v", err)
		}
		if *req.CustomTitle != "Spaced Title" {
			t.Errorf("Expected trimmed title 'Spaced Title', got '%s'", *req.CustomTitle)
		}
	})

	t.Run("Custom title too short", func(t *testing.T) {
		customTitle := "AB"
		req := &SubmitClipRequest{
			ClipURL:     "https://clips.twitch.tv/Test",
			CustomTitle: &customTitle,
		}

		err := service.validateSubmissionInput(req)
		if err == nil {
			t.Error("Expected error for title too short")
		}
		if valErr, ok := err.(*ValidationError); ok {
			if valErr.Field != "custom_title" {
				t.Errorf("Expected field 'custom_title', got '%s'", valErr.Field)
			}
		}
	})

	t.Run("Custom title too long", func(t *testing.T) {
		customTitle := string(make([]byte, 201))
		req := &SubmitClipRequest{
			ClipURL:     "https://clips.twitch.tv/Test",
			CustomTitle: &customTitle,
		}

		err := service.validateSubmissionInput(req)
		if err == nil {
			t.Error("Expected error for title too long")
		}
	})

	t.Run("Empty custom title after trimming is set to nil", func(t *testing.T) {
		customTitle := "   "
		req := &SubmitClipRequest{
			ClipURL:     "https://clips.twitch.tv/Test",
			CustomTitle: &customTitle,
		}

		err := service.validateSubmissionInput(req)
		if err != nil {
			t.Errorf("Empty title after trimming should not error: %v", err)
		}
		if req.CustomTitle != nil {
			t.Error("Expected CustomTitle to be nil after trimming empty string")
		}
	})

	t.Run("Broadcaster name normalization", func(t *testing.T) {
		broadcaster := "  StreamerName  "
		req := &SubmitClipRequest{
			ClipURL:                 "https://clips.twitch.tv/Test",
			BroadcasterNameOverride: &broadcaster,
		}

		err := service.validateSubmissionInput(req)
		if err != nil {
			t.Errorf("Valid broadcaster name should not error: %v", err)
		}
		if *req.BroadcasterNameOverride != "StreamerName" {
			t.Errorf("Expected trimmed broadcaster 'StreamerName', got '%s'", *req.BroadcasterNameOverride)
		}
	})

	t.Run("Broadcaster name too short", func(t *testing.T) {
		broadcaster := "A"
		req := &SubmitClipRequest{
			ClipURL:                 "https://clips.twitch.tv/Test",
			BroadcasterNameOverride: &broadcaster,
		}

		err := service.validateSubmissionInput(req)
		if err == nil {
			t.Error("Expected error for broadcaster name too short")
		}
	})

	t.Run("Broadcaster name with invalid characters", func(t *testing.T) {
		broadcaster := "Streamer-Name"
		req := &SubmitClipRequest{
			ClipURL:                 "https://clips.twitch.tv/Test",
			BroadcasterNameOverride: &broadcaster,
		}

		err := service.validateSubmissionInput(req)
		if err == nil {
			t.Error("Expected error for broadcaster name with invalid characters")
		}
		if valErr, ok := err.(*ValidationError); ok {
			if valErr.Field != "broadcaster_name_override" {
				t.Errorf("Expected field 'broadcaster_name_override', got '%s'", valErr.Field)
			}
		}
	})

	t.Run("Valid broadcaster name with underscore", func(t *testing.T) {
		broadcaster := "Streamer_Name"
		req := &SubmitClipRequest{
			ClipURL:                 "https://clips.twitch.tv/Test",
			BroadcasterNameOverride: &broadcaster,
		}

		err := service.validateSubmissionInput(req)
		if err != nil {
			t.Errorf("Broadcaster name with underscore should be valid: %v", err)
		}
	})

	t.Run("Tags normalization", func(t *testing.T) {
		req := &SubmitClipRequest{
			ClipURL: "https://clips.twitch.tv/Test",
			Tags:    []string{"  EPIC  ", "Gaming", "highlight"},
		}

		err := service.validateSubmissionInput(req)
		if err != nil {
			t.Errorf("Valid tags should not error: %v", err)
		}
		if len(req.Tags) != 3 {
			t.Errorf("Expected 3 tags, got %d", len(req.Tags))
		}
		if req.Tags[0] != "epic" || req.Tags[1] != "gaming" || req.Tags[2] != "highlight" {
			t.Errorf("Expected normalized tags, got %v", req.Tags)
		}
	})

	t.Run("Duplicate tags removed", func(t *testing.T) {
		req := &SubmitClipRequest{
			ClipURL: "https://clips.twitch.tv/Test",
			Tags:    []string{"epic", "EPIC", "gaming", "Epic"},
		}

		err := service.validateSubmissionInput(req)
		if err != nil {
			t.Errorf("Tags with duplicates should not error: %v", err)
		}
		if len(req.Tags) != 2 {
			t.Errorf("Expected 2 unique tags after deduplication, got %d", len(req.Tags))
		}
	})

	t.Run("Empty tags removed", func(t *testing.T) {
		req := &SubmitClipRequest{
			ClipURL: "https://clips.twitch.tv/Test",
			Tags:    []string{"epic", "", "  ", "gaming"},
		}

		err := service.validateSubmissionInput(req)
		if err != nil {
			t.Errorf("Tags with empty entries should not error: %v", err)
		}
		if len(req.Tags) != 2 {
			t.Errorf("Expected 2 tags after removing empty, got %d", len(req.Tags))
		}
	})

	t.Run("Too many tags", func(t *testing.T) {
		req := &SubmitClipRequest{
			ClipURL: "https://clips.twitch.tv/Test",
			Tags:    []string{"tag1", "tag2", "tag3", "tag4", "tag5", "tag6", "tag7", "tag8", "tag9", "tag10", "tag11"},
		}

		err := service.validateSubmissionInput(req)
		if err == nil {
			t.Error("Expected error for too many tags")
		}
		if valErr, ok := err.(*ValidationError); ok {
			if valErr.Field != "tags" {
				t.Errorf("Expected field 'tags', got '%s'", valErr.Field)
			}
		}
	})

	t.Run("Tag too short", func(t *testing.T) {
		req := &SubmitClipRequest{
			ClipURL: "https://clips.twitch.tv/Test",
			Tags:    []string{"a"},
		}

		err := service.validateSubmissionInput(req)
		if err == nil {
			t.Error("Expected error for tag too short")
		}
	})

	t.Run("Tag too long", func(t *testing.T) {
		longTag := string(make([]byte, 51))
		req := &SubmitClipRequest{
			ClipURL: "https://clips.twitch.tv/Test",
			Tags:    []string{longTag},
		}

		err := service.validateSubmissionInput(req)
		if err == nil {
			t.Error("Expected error for tag too long")
		}
	})

	t.Run("Tag with invalid characters", func(t *testing.T) {
		req := &SubmitClipRequest{
			ClipURL: "https://clips.twitch.tv/Test",
			Tags:    []string{"tag_with_underscore"},
		}

		err := service.validateSubmissionInput(req)
		if err == nil {
			t.Error("Expected error for tag with invalid characters")
		}
	})

	t.Run("Valid tag with hyphen", func(t *testing.T) {
		req := &SubmitClipRequest{
			ClipURL: "https://clips.twitch.tv/Test",
			Tags:    []string{"super-cool"},
		}

		err := service.validateSubmissionInput(req)
		if err != nil {
			t.Errorf("Tag with hyphen should be valid: %v", err)
		}
	})

	t.Run("Submission reason normalization", func(t *testing.T) {
		reason := "  This is an amazing clip  "
		req := &SubmitClipRequest{
			ClipURL:          "https://clips.twitch.tv/Test",
			SubmissionReason: &reason,
		}

		err := service.validateSubmissionInput(req)
		if err != nil {
			t.Errorf("Valid reason should not error: %v", err)
		}
		if *req.SubmissionReason != "This is an amazing clip" {
			t.Errorf("Expected trimmed reason, got '%s'", *req.SubmissionReason)
		}
	})

	t.Run("Submission reason too long", func(t *testing.T) {
		reason := string(make([]byte, 1001))
		req := &SubmitClipRequest{
			ClipURL:          "https://clips.twitch.tv/Test",
			SubmissionReason: &reason,
		}

		err := service.validateSubmissionInput(req)
		if err == nil {
			t.Error("Expected error for reason too long")
		}
	})

	t.Run("Empty submission reason after trimming is set to nil", func(t *testing.T) {
		reason := "   "
		req := &SubmitClipRequest{
			ClipURL:          "https://clips.twitch.tv/Test",
			SubmissionReason: &reason,
		}

		err := service.validateSubmissionInput(req)
		if err != nil {
			t.Errorf("Empty reason after trimming should not error: %v", err)
		}
		if req.SubmissionReason != nil {
			t.Error("Expected SubmissionReason to be nil after trimming empty string")
		}
	})
}

// TestIsValidUsername tests username validation
func TestIsValidUsername(t *testing.T) {
	tests := []struct {
		name     string
		username string
		expected bool
	}{
		{"Valid alphanumeric", "StreamerName123", true},
		{"Valid with underscore", "Streamer_Name", true},
		{"Valid lowercase", "streamername", true},
		{"Valid uppercase", "STREAMERNAME", true},
		{"Invalid with hyphen", "Streamer-Name", false},
		{"Invalid with space", "Streamer Name", false},
		{"Invalid with special char", "Streamer@Name", false},
		{"Invalid with dot", "Streamer.Name", false},
		{"Empty string", "", false}, // Empty string should be invalid
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isValidUsername(tt.username)
			if result != tt.expected {
				t.Errorf("isValidUsername(%s) = %v, want %v", tt.username, result, tt.expected)
			}
		})
	}
}

// TestIsValidTag tests tag validation
func TestIsValidTag(t *testing.T) {
	tests := []struct {
		name     string
		tag      string
		expected bool
	}{
		{"Valid alphanumeric", "gaming123", true},
		{"Valid with hyphen", "super-cool", true},
		{"Valid lowercase", "epic", true},
		{"Invalid uppercase", "EPIC", false}, // isValidTag expects normalized (lowercase) input; uppercase should fail
		{"Invalid with underscore", "tag_name", false},
		{"Invalid with space", "tag name", false},
		{"Invalid with special char", "tag@name", false},
		{"Empty string", "", false}, // Empty string should be invalid
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isValidTag(tt.tag)
			if result != tt.expected {
				t.Errorf("isValidTag(%s) = %v, want %v", tt.tag, result, tt.expected)
			}
		})
	}
}

// TestNormalizeClipURL tests URL normalization
func TestNormalizeClipURL(t *testing.T) {
	service := &SubmissionService{}

	tests := []struct {
		name        string
		input       string
		expectedID  string
		expectedURL string
		description string
	}{
		{
			name:        "clips.twitch.tv URL",
			input:       "https://clips.twitch.tv/AwkwardHelplessSalamanderSwiftRage",
			expectedID:  "AwkwardHelplessSalamanderSwiftRage",
			expectedURL: "https://clips.twitch.tv/AwkwardHelplessSalamanderSwiftRage",
			description: "Should extract ID and return normalized URL",
		},
		{
			name:        "www.twitch.tv clip URL",
			input:       "https://www.twitch.tv/username/clip/AwkwardHelplessSalamanderSwiftRage",
			expectedID:  "AwkwardHelplessSalamanderSwiftRage",
			expectedURL: "https://clips.twitch.tv/AwkwardHelplessSalamanderSwiftRage",
			description: "Should extract ID and return normalized URL",
		},
		{
			name:        "URL with query parameters",
			input:       "https://clips.twitch.tv/AwkwardHelplessSalamanderSwiftRage?filter=clips",
			expectedID:  "AwkwardHelplessSalamanderSwiftRage",
			expectedURL: "https://clips.twitch.tv/AwkwardHelplessSalamanderSwiftRage",
			description: "Should strip query parameters",
		},
		{
			name:        "Direct clip ID",
			input:       "AwkwardHelplessSalamanderSwiftRage",
			expectedID:  "AwkwardHelplessSalamanderSwiftRage",
			expectedURL: "https://clips.twitch.tv/AwkwardHelplessSalamanderSwiftRage",
			description: "Should accept direct clip ID",
		},
		{
			name:        "Invalid URL - will be caught by Twitch API",
			input:       "not-a-valid-url",
			expectedID:  "not-a-valid-url",
			expectedURL: "https://clips.twitch.tv/not-a-valid-url",
			description: "Invalid formats are accepted but will fail at Twitch API validation",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clipID, normalizedURL := service.normalizeClipURL(tt.input)
			if clipID != tt.expectedID {
				t.Errorf("Expected clipID '%s', got '%s'. %s", tt.expectedID, clipID, tt.description)
			}
			if normalizedURL != tt.expectedURL {
				t.Errorf("Expected normalizedURL '%s', got '%s'. %s", tt.expectedURL, normalizedURL, tt.description)
			}
		})
	}
}

func TestValidateClipQuality_UsesConfiguredMaxDuration(t *testing.T) {
	service := &SubmissionService{
		cfg: &config.Config{Clip: config.ClipConfig{MaxDurationSeconds: 30}},
	}
	clip := &twitch.Clip{
		CreatedAt:       time.Now().Add(-24 * time.Hour),
		Duration:        31,
		Title:           "Clip title",
		BroadcasterName: "StreamerName",
	}

	err := service.validateClipQuality(clip)
	if err == nil {
		t.Fatal("expected error for clip longer than configured maximum")
	}
	valErr, ok := err.(*ValidationError)
	if !ok {
		t.Fatalf("expected ValidationError, got %T", err)
	}
	if valErr.Field != "clip" {
		t.Fatalf("Field = %q, want clip", valErr.Field)
	}
	if !strings.Contains(valErr.Message, "at most 30 seconds") {
		t.Fatalf("Message = %q, want configured max duration", valErr.Message)
	}

	clip.Duration = 30
	if err := service.validateClipQuality(clip); err != nil {
		t.Fatalf("expected clip at configured max duration to pass, got %v", err)
	}
}

// TestValidationErrorMessages tests that validation errors have actionable messages
func TestValidationErrorMessages(t *testing.T) {
	service := &SubmissionService{}

	tests := []struct {
		name          string
		req           *SubmitClipRequest
		expectedField string
		shouldContain string
		description   string
	}{
		{
			name: "Custom title too short",
			req: &SubmitClipRequest{
				ClipURL:     "https://clips.twitch.tv/Test",
				CustomTitle: strPtr("AB"),
			},
			expectedField: "custom_title",
			shouldContain: "at least 3 characters",
			description:   "Error message should be actionable",
		},
		{
			name: "Broadcaster name with invalid chars",
			req: &SubmitClipRequest{
				ClipURL:                 "https://clips.twitch.tv/Test",
				BroadcasterNameOverride: strPtr("Name-With-Hyphens"),
			},
			expectedField: "broadcaster_name_override",
			shouldContain: "only contain",
			description:   "Error message should explain valid characters",
		},
		{
			name: "Too many tags",
			req: &SubmitClipRequest{
				ClipURL: "https://clips.twitch.tv/Test",
				Tags:    []string{"1", "2", "3", "4", "5", "6", "7", "8", "9", "10", "11"},
			},
			expectedField: "tags",
			shouldContain: "no more than 10",
			description:   "Error message should specify the limit",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.validateSubmissionInput(tt.req)
			if err == nil {
				t.Fatalf("Expected validation error, got nil. %s", tt.description)
			}

			valErr, ok := err.(*ValidationError)
			if !ok {
				t.Fatalf("Expected ValidationError, got %T", err)
			}

			if valErr.Field != tt.expectedField {
				t.Errorf("Expected field '%s', got '%s'", tt.expectedField, valErr.Field)
			}

			if !strings.Contains(valErr.Message, tt.shouldContain) {
				t.Errorf("Expected message to contain '%s', got '%s'. %s",
					tt.shouldContain, valErr.Message, tt.description)
			}
		})
	}
}
