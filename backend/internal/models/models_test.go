package models

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestRejectionReasonTemplates(t *testing.T) {
	templates := GetRejectionReasonTemplates()

	if len(templates) == 0 {
		t.Error("Expected at least one rejection reason template")
	}

	// Check that all expected templates are present
	expectedReasons := []string{
		RejectionReasonLowQuality,
		RejectionReasonDuplicate,
		RejectionReasonInappropriate,
		RejectionReasonOffTopic,
		RejectionReasonPoorTitle,
		RejectionReasonTooShort,
		RejectionReasonTooLong,
		RejectionReasonSpam,
		RejectionReasonViolatesGuidelines,
		RejectionReasonOther,
	}

	foundReasons := make(map[string]bool)
	for _, template := range templates {
		foundReasons[template] = true
	}

	for _, expected := range expectedReasons {
		if !foundReasons[expected] {
			t.Errorf("Expected rejection reason template '%s' not found", expected)
		}
	}

	// Verify each constant has a non-empty value
	if RejectionReasonLowQuality == "" {
		t.Error("RejectionReasonLowQuality should not be empty")
	}
	if RejectionReasonDuplicate == "" {
		t.Error("RejectionReasonDuplicate should not be empty")
	}
	if RejectionReasonInappropriate == "" {
		t.Error("RejectionReasonInappropriate should not be empty")
	}
	if RejectionReasonOffTopic == "" {
		t.Error("RejectionReasonOffTopic should not be empty")
	}
	if RejectionReasonPoorTitle == "" {
		t.Error("RejectionReasonPoorTitle should not be empty")
	}
	if RejectionReasonTooShort == "" {
		t.Error("RejectionReasonTooShort should not be empty")
	}
	if RejectionReasonTooLong == "" {
		t.Error("RejectionReasonTooLong should not be empty")
	}
	if RejectionReasonSpam == "" {
		t.Error("RejectionReasonSpam should not be empty")
	}
	if RejectionReasonViolatesGuidelines == "" {
		t.Error("RejectionReasonViolatesGuidelines should not be empty")
	}
	if RejectionReasonOther == "" {
		t.Error("RejectionReasonOther should not be empty")
	}
}

func TestUserModeratorFieldsJSONMarshaling(t *testing.T) {
	now := time.Now()
	channelID1 := uuid.New()
	channelID2 := uuid.New()

	tests := []struct {
		name    string
		user    User
		checkFn func(*testing.T, []byte)
		desc    string
	}{
		{
			name: "site moderator JSON marshaling",
			user: User{
				ID:                  uuid.New(),
				Username:            "sitemod",
				DisplayName:         "Site Moderator",
				ModeratorScope:      ModeratorScopeSite,
				ModerationChannels:  []uuid.UUID{},
				ModerationStartedAt: &now,
			},
			checkFn: func(t *testing.T, jsonData []byte) {
				var result map[string]interface{}
				if err := json.Unmarshal(jsonData, &result); err != nil {
					t.Fatalf("Failed to unmarshal JSON: %v", err)
				}
				if result["moderator_scope"] != ModeratorScopeSite {
					t.Errorf("Expected moderator_scope to be %q, got %v", ModeratorScopeSite, result["moderator_scope"])
				}
				// Empty array with omitempty will be omitted in JSON
				// This is expected behavior for omitempty tag
				if result["moderation_started_at"] == nil {
					t.Error("Expected moderation_started_at to be present")
				}
			},
			desc: "Site moderator should marshal with empty channel list",
		},
		{
			name: "community moderator JSON marshaling",
			user: User{
				ID:                  uuid.New(),
				Username:            "commmod",
				DisplayName:         "Community Moderator",
				ModeratorScope:      ModeratorScopeCommunity,
				ModerationChannels:  []uuid.UUID{channelID1, channelID2},
				ModerationStartedAt: &now,
			},
			checkFn: func(t *testing.T, jsonData []byte) {
				var result map[string]interface{}
				if err := json.Unmarshal(jsonData, &result); err != nil {
					t.Fatalf("Failed to unmarshal JSON: %v", err)
				}
				if result["moderator_scope"] != ModeratorScopeCommunity {
					t.Errorf("Expected moderator_scope to be %q, got %v", ModeratorScopeCommunity, result["moderator_scope"])
				}
				channels, ok := result["moderation_channels"].([]interface{})
				if !ok {
					t.Fatalf("Expected moderation_channels to be array, got %T", result["moderation_channels"])
				}
				if len(channels) != 2 {
					t.Errorf("Expected 2 moderation channels, got %d", len(channels))
				}
			},
			desc: "Community moderator should marshal with channel list",
		},
		{
			name: "non-moderator JSON marshaling",
			user: User{
				ID:          uuid.New(),
				Username:    "regular",
				DisplayName: "Regular User",
			},
			checkFn: func(t *testing.T, jsonData []byte) {
				var result map[string]interface{}
				if err := json.Unmarshal(jsonData, &result); err != nil {
					t.Fatalf("Failed to unmarshal JSON: %v", err)
				}
				// Fields with omitempty should not be present for zero values
				if _, exists := result["moderator_scope"]; exists {
					t.Error("Expected moderator_scope to be omitted for non-moderator")
				}
				if _, exists := result["moderation_channels"]; exists {
					t.Error("Expected moderation_channels to be omitted for non-moderator")
				}
				if _, exists := result["moderation_started_at"]; exists {
					t.Error("Expected moderation_started_at to be omitted for non-moderator")
				}
			},
			desc: "Non-moderator should omit moderator fields",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jsonData, err := json.Marshal(tt.user)
			if err != nil {
				t.Fatalf("Failed to marshal user to JSON: %v", err)
			}

			tt.checkFn(t, jsonData)

			// Verify we can unmarshal back
			var unmarshaled User
			if err := json.Unmarshal(jsonData, &unmarshaled); err != nil {
				t.Fatalf("Failed to unmarshal JSON back to User: %v", err)
			}

			// Verify key fields match
			if unmarshaled.Username != tt.user.Username {
				t.Errorf("Username mismatch after unmarshal: got %q, want %q", unmarshaled.Username, tt.user.Username)
			}
			if unmarshaled.ModeratorScope != tt.user.ModeratorScope {
				t.Errorf("ModeratorScope mismatch after unmarshal: got %q, want %q", unmarshaled.ModeratorScope, tt.user.ModeratorScope)
			}
			if len(unmarshaled.ModerationChannels) != len(tt.user.ModerationChannels) {
				t.Errorf("ModerationChannels length mismatch after unmarshal: got %d, want %d",
					len(unmarshaled.ModerationChannels), len(tt.user.ModerationChannels))
			}
		})
	}
}
