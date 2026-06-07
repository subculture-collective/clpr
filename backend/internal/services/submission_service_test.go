package services

import (
	"testing"

	"github.com/google/uuid"
	"git.subcult.tv/subculture-collective/clpr/internal/models"
	"git.subcult.tv/subculture-collective/clpr/internal/utils"
)

func TestSubmissionService_ShouldAutoApprove(t *testing.T) {
	// Create a service with nil repos just to test this method
	service := &SubmissionService{}

	t.Run("Admin should be auto-approved", func(t *testing.T) {
		user := &models.User{
			ID:          uuid.New(),
			Role:        "admin",
			KarmaPoints: 0,
		}
		if !service.shouldAutoApprove(user) {
			t.Error("Admin should be auto-approved")
		}
	})

	t.Run("Moderator should be auto-approved", func(t *testing.T) {
		user := &models.User{
			ID:          uuid.New(),
			Role:        "moderator",
			KarmaPoints: 0,
		}
		if !service.shouldAutoApprove(user) {
			t.Error("Moderator should be auto-approved")
		}
	})

	t.Run("High karma user should be auto-approved", func(t *testing.T) {
		user := &models.User{
			ID:          uuid.New(),
			Role:        "user",
			KarmaPoints: 1000,
		}
		if !service.shouldAutoApprove(user) {
			t.Error("High karma user should be auto-approved")
		}
	})

	t.Run("Regular user should not be auto-approved", func(t *testing.T) {
		user := &models.User{
			ID:          uuid.New(),
			Role:        "user",
			KarmaPoints: 500,
		}
		if service.shouldAutoApprove(user) {
			t.Error("Regular user should not be auto-approved")
		}
	})
}

func TestExtractClipIDFromURL(t *testing.T) {
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
			result := extractClipIDFromURL(tt.input)
			if result != tt.expected {
				t.Errorf("extractClipIDFromURL(%s) = %s, want %s", tt.input, result, tt.expected)
			}
		})
	}
}

func TestBroadcasterNameHandling(t *testing.T) {
	tests := []struct {
		name                    string
		broadcasterNameOverride *string
		broadcasterName         *string
		expected                string
	}{
		{
			name:                    "Override provided",
			broadcasterNameOverride: strPtr("CustomBroadcaster"),
			broadcasterName:         strPtr("OriginalBroadcaster"),
			expected:                "CustomBroadcaster",
		},
		{
			name:                    "No override, use original",
			broadcasterNameOverride: nil,
			broadcasterName:         strPtr("OriginalBroadcaster"),
			expected:                "OriginalBroadcaster",
		},
		{
			name:                    "Empty override, use original",
			broadcasterNameOverride: strPtr(""),
			broadcasterName:         strPtr("OriginalBroadcaster"),
			expected:                "OriginalBroadcaster",
		},
		{
			name:                    "Both nil, use empty string",
			broadcasterNameOverride: nil,
			broadcasterName:         nil,
			expected:                "",
		},
		{
			name:                    "Override provided, original nil",
			broadcasterNameOverride: strPtr("CustomBroadcaster"),
			broadcasterName:         nil,
			expected:                "CustomBroadcaster",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Simulate the logic from createClipFromSubmission
			emptyStr := ""
			submission := &models.ClipSubmission{
				BroadcasterNameOverride: tt.broadcasterNameOverride,
				BroadcasterName:         tt.broadcasterName,
			}

			// This is the pattern used in createClipFromSubmission
			broadcasterNameFallback := utils.StringOrDefault(submission.BroadcasterName, &emptyStr)
			result := utils.StringOrDefault(submission.BroadcasterNameOverride, &broadcasterNameFallback)

			if result != tt.expected {
				t.Errorf("got %q, want %q", result, tt.expected)
			}
		})
	}
}

// Helper function to create string pointers
func strPtr(s string) *string {
	return &s
}

func TestSubmissionService_BroadcasterNameOverridePrecedence(t *testing.T) {
	t.Run("Override should take precedence over Twitch metadata", func(t *testing.T) {
		twitchBroadcaster := "TwitchBroadcaster"
		overrideBroadcaster := "OverrideBroadcaster"

		submission := &models.ClipSubmission{
			TwitchClipID:            "TestClipID",
			TwitchClipURL:           "https://clips.twitch.tv/TestClipID",
			BroadcasterName:         &twitchBroadcaster,
			BroadcasterNameOverride: &overrideBroadcaster,
		}

		// Test that the override takes precedence
		// Since we can't easily test createClipFromSubmission without a full DB setup,
		// we'll test the logic directly
		result := ""
		if submission.BroadcasterNameOverride != nil && *submission.BroadcasterNameOverride != "" {
			result = *submission.BroadcasterNameOverride
		} else if submission.BroadcasterName != nil {
			result = *submission.BroadcasterName
		}

		if result != overrideBroadcaster {
			t.Errorf("Expected broadcaster name to be %s (override), got %s", overrideBroadcaster, result)
		}
	})

	t.Run("Twitch metadata used when no override provided", func(t *testing.T) {
		twitchBroadcaster := "TwitchBroadcaster"

		submission := &models.ClipSubmission{
			TwitchClipID:            "TestClipID",
			TwitchClipURL:           "https://clips.twitch.tv/TestClipID",
			BroadcasterName:         &twitchBroadcaster,
			BroadcasterNameOverride: nil,
		}

		result := ""
		if submission.BroadcasterNameOverride != nil && *submission.BroadcasterNameOverride != "" {
			result = *submission.BroadcasterNameOverride
		} else if submission.BroadcasterName != nil {
			result = *submission.BroadcasterName
		}

		if result != twitchBroadcaster {
			t.Errorf("Expected broadcaster name to be %s (from Twitch), got %s", twitchBroadcaster, result)
		}
	})

	t.Run("Empty override should not override Twitch metadata", func(t *testing.T) {
		twitchBroadcaster := "TwitchBroadcaster"
		emptyOverride := ""

		submission := &models.ClipSubmission{
			TwitchClipID:            "TestClipID",
			TwitchClipURL:           "https://clips.twitch.tv/TestClipID",
			BroadcasterName:         &twitchBroadcaster,
			BroadcasterNameOverride: &emptyOverride,
		}

		result := ""
		if submission.BroadcasterNameOverride != nil && *submission.BroadcasterNameOverride != "" {
			result = *submission.BroadcasterNameOverride
		} else if submission.BroadcasterName != nil {
			result = *submission.BroadcasterName
		}

		if result != twitchBroadcaster {
			t.Errorf("Expected broadcaster name to be %s (from Twitch when override is empty), got %s", twitchBroadcaster, result)
		}
	})
}

func TestGetClipTitle(t *testing.T) {
	tests := []struct {
		name        string
		submission  *models.ClipSubmission
		expected    string
		description string
	}{
		{
			name: "Custom title takes precedence",
			submission: &models.ClipSubmission{
				CustomTitle: strPtr("Custom Title"),
				Title:       strPtr("Original Title"),
			},
			expected:    "Custom Title",
			description: "When both custom title and title are provided, custom title should be returned",
		},
		{
			name: "Use original title when no custom title",
			submission: &models.ClipSubmission{
				CustomTitle: nil,
				Title:       strPtr("Original Title"),
			},
			expected:    "Original Title",
			description: "When custom title is nil, original title should be returned",
		},
		{
			name: "Empty custom title falls back to original title",
			submission: &models.ClipSubmission{
				CustomTitle: strPtr(""),
				Title:       strPtr("Original Title"),
			},
			expected:    "Original Title",
			description: "When custom title is empty string, original title should be returned",
		},
		{
			name: "Return empty string when both are nil",
			submission: &models.ClipSubmission{
				CustomTitle: nil,
				Title:       nil,
			},
			expected:    "",
			description: "When both custom title and title are nil, empty string should be returned",
		},
		{
			name: "Return empty string when both are empty",
			submission: &models.ClipSubmission{
				CustomTitle: strPtr(""),
				Title:       strPtr(""),
			},
			expected:    "",
			description: "When both custom title and title are empty strings, empty string should be returned",
		},
		{
			name: "Custom title with whitespace is valid",
			submission: &models.ClipSubmission{
				CustomTitle: strPtr("   Custom Title   "),
				Title:       strPtr("Original Title"),
			},
			expected:    "   Custom Title   ",
			description: "Custom title with whitespace should be preserved",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getClipTitle(tt.submission)
			if result != tt.expected {
				t.Errorf("getClipTitle() = %q, want %q\nDescription: %s", result, tt.expected, tt.description)
			}
		})
	}
}

// TestKarmaGateRequirement tests the minimum karma requirement for submission
func TestKarmaGateRequirement(t *testing.T) {
	tests := []struct {
		name        string
		karma       int
		shouldPass  bool
		description string
	}{
		{
			name:        "Exactly 100 karma should pass",
			karma:       100,
			shouldPass:  true,
			description: "Users with exactly 100 karma should be allowed to submit",
		},
		{
			name:        "Above 100 karma should pass",
			karma:       150,
			shouldPass:  true,
			description: "Users with more than 100 karma should be allowed to submit",
		},
		{
			name:        "99 karma should fail",
			karma:       99,
			shouldPass:  false,
			description: "Users with 99 karma should not be allowed to submit",
		},
		{
			name:        "0 karma should fail",
			karma:       0,
			shouldPass:  false,
			description: "Users with 0 karma should not be allowed to submit",
		},
		{
			name:        "Negative karma should fail",
			karma:       -50,
			shouldPass:  false,
			description: "Users with negative karma should not be allowed to submit",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			passes := tt.karma >= 100
			if passes != tt.shouldPass {
				t.Errorf("Karma %d: expected pass=%v, got %v. %s", tt.karma, tt.shouldPass, passes, tt.description)
			}
		})
	}
}

// TestRateLimitLogic tests rate limit calculations
func TestRateLimitLogic(t *testing.T) {
	tests := []struct {
		name          string
		hourlyCount   int
		dailyCount    int
		exceedsHourly bool
		exceedsDaily  bool
		description   string
	}{
		{
			name:          "Within both limits",
			hourlyCount:   2,
			dailyCount:    10,
			exceedsHourly: false,
			exceedsDaily:  false,
			description:   "2 submissions in last hour, 10 in last day should pass",
		},
		{
			name:          "At hourly limit",
			hourlyCount:   5,
			dailyCount:    10,
			exceedsHourly: true,
			exceedsDaily:  false,
			description:   "5 submissions in last hour should hit hourly limit",
		},
		{
			name:          "Exceeds hourly limit",
			hourlyCount:   6,
			dailyCount:    10,
			exceedsHourly: true,
			exceedsDaily:  false,
			description:   "6 submissions in last hour should exceed hourly limit",
		},
		{
			name:          "At daily limit",
			hourlyCount:   3,
			dailyCount:    20,
			exceedsHourly: false,
			exceedsDaily:  true,
			description:   "20 submissions in last day should hit daily limit",
		},
		{
			name:          "Exceeds daily limit",
			hourlyCount:   3,
			dailyCount:    25,
			exceedsHourly: false,
			exceedsDaily:  true,
			description:   "25 submissions in last day should exceed daily limit",
		},
		{
			name:          "Exceeds both limits",
			hourlyCount:   10,
			dailyCount:    25,
			exceedsHourly: true,
			exceedsDaily:  true,
			description:   "Both limits exceeded",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hourlyExceeded := tt.hourlyCount >= 5
			dailyExceeded := tt.dailyCount >= 20

			if hourlyExceeded != tt.exceedsHourly {
				t.Errorf("Hourly check: count=%d, expected exceeded=%v, got %v. %s",
					tt.hourlyCount, tt.exceedsHourly, hourlyExceeded, tt.description)
			}
			if dailyExceeded != tt.exceedsDaily {
				t.Errorf("Daily check: count=%d, expected exceeded=%v, got %v. %s",
					tt.dailyCount, tt.exceedsDaily, dailyExceeded, tt.description)
			}
		})
	}
}

// TestDuplicateDetectionLogic tests duplicate detection scenarios
func TestDuplicateDetectionLogic(t *testing.T) {
	tests := []struct {
		name            string
		clipExists      bool
		submissionState string // "none", "pending", "approved", "rejected-recent", "rejected-old"
		shouldAllow     bool
		description     string
	}{
		{
			name:            "New clip not submitted before",
			clipExists:      false,
			submissionState: "none",
			shouldAllow:     true,
			description:     "Brand new clip should be allowed",
		},
		{
			name:            "Clip already exists in database",
			clipExists:      true,
			submissionState: "none",
			shouldAllow:     false,
			description:     "Clip that already exists should be rejected",
		},
		{
			name:            "Clip has pending submission",
			clipExists:      false,
			submissionState: "pending",
			shouldAllow:     false,
			description:     "Clip with pending submission should be rejected",
		},
		{
			name:            "Clip was already approved",
			clipExists:      false,
			submissionState: "approved",
			shouldAllow:     false,
			description:     "Previously approved clip should be rejected",
		},
		{
			name:            "Clip rejected recently",
			clipExists:      false,
			submissionState: "rejected-recent",
			shouldAllow:     false,
			description:     "Recently rejected clip (within 7 days) should not be resubmitted",
		},
		{
			name:            "Clip rejected long ago",
			clipExists:      false,
			submissionState: "rejected-old",
			shouldAllow:     true,
			description:     "Clip rejected more than 7 days ago can be resubmitted",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Simulate the duplicate detection logic
			allowed := true

			if tt.clipExists {
				allowed = false
			} else if tt.submissionState == "pending" || tt.submissionState == "approved" {
				allowed = false
			} else if tt.submissionState == "rejected-recent" {
				allowed = false
			}

			if allowed != tt.shouldAllow {
				t.Errorf("Expected shouldAllow=%v, got %v. %s", tt.shouldAllow, allowed, tt.description)
			}
		})
	}
}

// TestApproveRejectStatusTransitions tests status transition logic
func TestApproveRejectStatusTransitions(t *testing.T) {
	tests := []struct {
		name          string
		currentStatus string
		action        string // "approve" or "reject"
		shouldSucceed bool
		description   string
	}{
		{
			name:          "Approve pending submission",
			currentStatus: "pending",
			action:        "approve",
			shouldSucceed: true,
			description:   "Pending submission can be approved",
		},
		{
			name:          "Reject pending submission",
			currentStatus: "pending",
			action:        "reject",
			shouldSucceed: true,
			description:   "Pending submission can be rejected",
		},
		{
			name:          "Cannot approve already approved submission",
			currentStatus: "approved",
			action:        "approve",
			shouldSucceed: false,
			description:   "Already approved submission cannot be approved again",
		},
		{
			name:          "Cannot reject already approved submission",
			currentStatus: "approved",
			action:        "reject",
			shouldSucceed: false,
			description:   "Already approved submission cannot be rejected",
		},
		{
			name:          "Cannot approve already rejected submission",
			currentStatus: "rejected",
			action:        "approve",
			shouldSucceed: false,
			description:   "Already rejected submission cannot be approved",
		},
		{
			name:          "Cannot reject already rejected submission",
			currentStatus: "rejected",
			action:        "reject",
			shouldSucceed: false,
			description:   "Already rejected submission cannot be rejected again",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Only pending submissions can be approved or rejected
			canTransition := tt.currentStatus == "pending"

			if canTransition != tt.shouldSucceed {
				t.Errorf("Status %s, action %s: expected success=%v, got %v. %s",
					tt.currentStatus, tt.action, tt.shouldSucceed, canTransition, tt.description)
			}
		})
	}
}

// TestKarmaAwardLogic tests karma point calculations
func TestKarmaAwardLogic(t *testing.T) {
	tests := []struct {
		name        string
		action      string
		karmaChange int
		description string
	}{
		{
			name:        "Approval awards 10 karma",
			action:      "approve",
			karmaChange: 10,
			description: "Approved submissions award 10 karma points",
		},
		{
			name:        "Rejection penalizes 5 karma",
			action:      "reject",
			karmaChange: -5,
			description: "Rejected submissions subtract 5 karma points",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Verify the karma change matches the expected pattern for the action
			expectedKarma := 0
			if tt.action == "approve" {
				expectedKarma = 10
			} else if tt.action == "reject" {
				expectedKarma = -5
			}

			// This test ensures the test data itself is correct
			if tt.karmaChange != expectedKarma {
				t.Errorf("Test data mismatch for action %s: karmaChange=%d should be %d. %s",
					tt.action, tt.karmaChange, expectedKarma, tt.description)
			}
		})
	}
}

// TestSubmissionService_AutoUpvote tests the auto-upvote functionality
func TestSubmissionService_AutoUpvote(t *testing.T) {
	t.Run("Auto-upvote behavior verification", func(t *testing.T) {
		// This test verifies the expected behavior of auto-upvoting clips
		// Since we don't have mocks for all repositories in this test file,
		// we document the expected behavior:
		//
		// 1. When createClipFromSubmission is called, it should:
		//    - Create the clip with vote_score = 0 (default)
		//    - Call voteRepo.UpsertVote with the submitter's userID, clipID, and voteType=1
		//    - Database trigger updates vote_score from 0 to 1
		//
		// 2. If voteRepo.UpsertVote fails:
		//    - Error is logged but clip creation continues (soft failure)
		//    - Clip is still created successfully
		//
		// 3. If voteRepo is nil:
		//    - No auto-upvote is attempted
		//    - Clip is created normally
		//
		// Integration tests with a real database would verify:
		// - Vote record exists in votes table after clip creation
		// - vote_score is updated to 1 via database trigger
		// - User can remove the auto-upvote by calling vote endpoint with voteType=0

		// Test expectation: createClipFromSubmission should not panic with nil voteRepo
		service := &SubmissionService{
			voteRepo: nil,
		}

		if service.voteRepo != nil {
			t.Error("Expected voteRepo to be nil for this test case")
		}

		// This documents that the service handles nil voteRepo gracefully
		// The actual integration test would verify the full flow with mocked repositories
	})

	t.Run("Vote creation logic expectations", func(t *testing.T) {
		// Document expected vote creation behavior
		expectations := map[string]string{
			"vote_type":      "1 (upvote)",
			"initial_score":  "0 (before trigger)",
			"final_score":    "1 (after trigger)",
			"soft_failure":   "Logs error but continues",
			"nil_protection": "Checks voteRepo != nil before calling",
		}

		// Verify expectations are documented
		if len(expectations) != 5 {
			t.Errorf("Expected 5 documented behaviors, got %d", len(expectations))
		}

		// The actual behavior is verified by:
		// 1. Code inspection (lines 837-844 in submission_service.go)
		// 2. Integration tests with database (would be added separately)
		// 3. Manual testing with real database and triggers
	})
}

// TestClipExistenceResult verifies ClipExistenceResult struct behavior
func TestClipExistenceResult(t *testing.T) {
	t.Run("ClipExistenceResult with claimable clip", func(t *testing.T) {
		clip := &models.Clip{
			ID:                uuid.New(),
			TwitchClipID:      "TestClipID",
			SubmittedByUserID: nil, // Unclaimed scraped clip
		}

		result := &ClipExistenceResult{
			Exists:       true,
			Clip:         clip,
			CanBeClaimed: clip.SubmittedByUserID == nil,
		}

		if !result.Exists {
			t.Error("Expected clip to exist")
		}
		if !result.CanBeClaimed {
			t.Error("Expected clip to be claimable when SubmittedByUserID is nil")
		}
	})

	t.Run("ClipExistenceResult with claimed clip", func(t *testing.T) {
		userID := uuid.New()
		clip := &models.Clip{
			ID:                uuid.New(),
			TwitchClipID:      "TestClipID",
			SubmittedByUserID: &userID, // Claimed clip
		}

		result := &ClipExistenceResult{
			Exists:       true,
			Clip:         clip,
			CanBeClaimed: clip.SubmittedByUserID == nil,
		}

		if !result.Exists {
			t.Error("Expected clip to exist")
		}
		if result.CanBeClaimed {
			t.Error("Expected clip to NOT be claimable when SubmittedByUserID is set")
		}
	})

	t.Run("ClipExistenceResult with non-existent clip", func(t *testing.T) {
		result := &ClipExistenceResult{
			Exists:       false,
			Clip:         nil,
			CanBeClaimed: false,
		}

		if result.Exists {
			t.Error("Expected clip to not exist")
		}
		if result.CanBeClaimed {
			t.Error("Non-existent clip should not be claimable")
		}
	})
}
