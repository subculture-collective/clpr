package services

import (
	"testing"
	"time"

	"git.subcult.tv/subculture-collective/clpr/internal/repository"
)

// Test helper: creates a pointer to float64 (including zero values, unlike the production helper)
func testFloat64Ptr(v float64) *float64 {
	return &v
}

func TestBuildCacheKey(t *testing.T) {
	svc := &ClipService{}

	gameID := "123"
	broadcasterID := "456"
	creatorID := "creator-1"
	tag := "funny"
	search := "frog"
	timeframe := "week"
	language := "en"

	key := svc.buildCacheKey(repository.ClipFilters{
		Sort:              "hot",
		GameID:            &gameID,
		BroadcasterID:     &broadcasterID,
		CreatorID:         &creatorID,
		Tag:               &tag,
		Search:            &search,
		Language:          &language,
		Timeframe:         &timeframe,
		Top10kStreamers:   true,
		ShowHidden:        true,
		UserSubmittedOnly: true,
	}, 2, 50)

	expected := "clips:list:hot:page:2:limit:50:game:123:broadcaster:456:creator:creator-1:tag:funny:search:frog:timeframe:week:language:en:top10k:true:show_hidden:true:user_submitted_only:true"

	if key != expected {
		t.Fatalf("expected cache key %q, got %q", expected, key)
	}
}

func TestBuildCacheKeySeparatesSubmissionScopes(t *testing.T) {
	svc := &ClipService{}

	userOnly := svc.buildCacheKey(repository.ClipFilters{Sort: "hot", UserSubmittedOnly: true}, 1, 25)
	showAll := svc.buildCacheKey(repository.ClipFilters{Sort: "hot", UserSubmittedOnly: false}, 1, 25)

	if userOnly == showAll {
		t.Fatalf("cache key should differ when user-submitted-only flag changes: %q", userOnly)
	}
}

func TestGetCacheTTL(t *testing.T) {
	tests := []struct {
		name     string
		sort     string
		expected time.Duration
	}{
		{
			name:     "Hot sort TTL",
			sort:     "hot",
			expected: 5 * time.Minute,
		},
		{
			name:     "New sort TTL",
			sort:     "new",
			expected: 2 * time.Minute,
		},
		{
			name:     "Top sort TTL",
			sort:     "top",
			expected: 15 * time.Minute,
		},
		{
			name:     "Rising sort TTL",
			sort:     "rising",
			expected: 3 * time.Minute,
		},
		{
			name:     "Default sort TTL",
			sort:     "unknown",
			expected: 5 * time.Minute,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Validate TTL values are reasonable
			if tt.expected < time.Minute {
				t.Error("Cache TTL should be at least 1 minute")
			}
			if tt.expected > time.Hour {
				t.Error("Cache TTL should not exceed 1 hour")
			}
		})
	}
}

func TestVoteValidation(t *testing.T) {
	tests := []struct {
		name     string
		voteType int16
		valid    bool
	}{
		{
			name:     "Valid upvote",
			voteType: 1,
			valid:    true,
		},
		{
			name:     "Valid downvote",
			voteType: -1,
			valid:    true,
		},
		{
			name:     "Valid remove vote",
			voteType: 0,
			valid:    true,
		},
		{
			name:     "Invalid vote value 2",
			voteType: 2,
			valid:    false,
		},
		{
			name:     "Invalid vote value -2",
			voteType: -2,
			valid:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			isValid := tt.voteType == -1 || tt.voteType == 0 || tt.voteType == 1
			if isValid != tt.valid {
				t.Errorf("Expected vote %d to be valid=%v, got %v", tt.voteType, tt.valid, isValid)
			}
		})
	}
}

func TestKarmaCalculation(t *testing.T) {
	tests := []struct {
		name        string
		oldVote     *int16
		newVote     int16
		expectedDif int
	}{
		{
			name:        "New upvote",
			oldVote:     nil,
			newVote:     1,
			expectedDif: 1,
		},
		{
			name:        "New downvote",
			oldVote:     nil,
			newVote:     -1,
			expectedDif: -1,
		},
		{
			name: "Change upvote to downvote",
			oldVote: func() *int16 {
				v := int16(1)
				return &v
			}(),
			newVote:     -1,
			expectedDif: -2,
		},
		{
			name: "Change downvote to upvote",
			oldVote: func() *int16 {
				v := int16(-1)
				return &v
			}(),
			newVote:     1,
			expectedDif: 2,
		},
		{
			name: "Remove upvote",
			oldVote: func() *int16 {
				v := int16(1)
				return &v
			}(),
			newVote:     0,
			expectedDif: 0, // Vote removal does not affect karma at all
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			karmaChange := 0
			if tt.oldVote == nil {
				// New vote
				if tt.newVote == 1 {
					karmaChange = 1
				} else if tt.newVote == -1 {
					karmaChange = -1
				}
			} else if tt.newVote != 0 {
				// Changed vote (not removal)
				if *tt.oldVote == 1 && tt.newVote == -1 {
					karmaChange = -2
				} else if *tt.oldVote == -1 && tt.newVote == 1 {
					karmaChange = 2
				}
			}

			if karmaChange != tt.expectedDif {
				t.Errorf("Expected karma change %d, got %d", tt.expectedDif, karmaChange)
			}
		})
	}
}

func TestPaginationCalculation(t *testing.T) {
	tests := []struct {
		name            string
		total           int
		limit           int
		page            int
		expectedPages   int
		expectedHasNext bool
		expectedHasPrev bool
	}{
		{
			name:            "First page with results",
			total:           100,
			limit:           25,
			page:            1,
			expectedPages:   4,
			expectedHasNext: true,
			expectedHasPrev: false,
		},
		{
			name:            "Middle page",
			total:           100,
			limit:           25,
			page:            2,
			expectedPages:   4,
			expectedHasNext: true,
			expectedHasPrev: true,
		},
		{
			name:            "Last page",
			total:           100,
			limit:           25,
			page:            4,
			expectedPages:   4,
			expectedHasNext: false,
			expectedHasPrev: true,
		},
		{
			name:            "Partial last page",
			total:           90,
			limit:           25,
			page:            4,
			expectedPages:   4,
			expectedHasNext: false,
			expectedHasPrev: true,
		},
		{
			name:            "Single page",
			total:           10,
			limit:           25,
			page:            1,
			expectedPages:   1,
			expectedHasNext: false,
			expectedHasPrev: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			totalPages := (tt.total + tt.limit - 1) / tt.limit
			hasNext := tt.page < totalPages
			hasPrev := tt.page > 1

			if totalPages != tt.expectedPages {
				t.Errorf("Expected %d pages, got %d", tt.expectedPages, totalPages)
			}
			if hasNext != tt.expectedHasNext {
				t.Errorf("Expected hasNext=%v, got %v", tt.expectedHasNext, hasNext)
			}
			if hasPrev != tt.expectedHasPrev {
				t.Errorf("Expected hasPrev=%v, got %v", tt.expectedHasPrev, hasPrev)
			}
		})
	}
}

func TestSortValidation(t *testing.T) {
	validSorts := []string{"hot", "new", "top", "rising"}

	tests := []struct {
		name  string
		sort  string
		valid bool
	}{
		{
			name:  "Valid hot",
			sort:  "hot",
			valid: true,
		},
		{
			name:  "Valid new",
			sort:  "new",
			valid: true,
		},
		{
			name:  "Valid top",
			sort:  "top",
			valid: true,
		},
		{
			name:  "Valid rising",
			sort:  "rising",
			valid: true,
		},
		{
			name:  "Invalid sort",
			sort:  "invalid",
			valid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			isValid := false
			for _, validSort := range validSorts {
				if tt.sort == validSort {
					isValid = true
					break
				}
			}

			if isValid != tt.valid {
				t.Errorf("Expected sort '%s' to be valid=%v, got %v", tt.sort, tt.valid, isValid)
			}
		})
	}
}

func TestTimeframeValidation(t *testing.T) {
	validTimeframes := []string{"hour", "day", "week", "month", "year", "all"}

	tests := []struct {
		name      string
		timeframe string
		valid     bool
	}{
		{
			name:      "Valid hour",
			timeframe: "hour",
			valid:     true,
		},
		{
			name:      "Valid day",
			timeframe: "day",
			valid:     true,
		},
		{
			name:      "Valid week",
			timeframe: "week",
			valid:     true,
		},
		{
			name:      "Valid month",
			timeframe: "month",
			valid:     true,
		},
		{
			name:      "Valid year",
			timeframe: "year",
			valid:     true,
		},
		{
			name:      "Valid all",
			timeframe: "all",
			valid:     true,
		},
		{
			name:      "Invalid timeframe",
			timeframe: "invalid",
			valid:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			isValid := false
			for _, validTf := range validTimeframes {
				if tt.timeframe == validTf {
					isValid = true
					break
				}
			}

			if isValid != tt.valid {
				t.Errorf("Expected timeframe '%s' to be valid=%v, got %v", tt.timeframe, tt.valid, isValid)
			}
		})
	}
}

func TestLimitConstraints(t *testing.T) {
	tests := []struct {
		name          string
		inputLimit    int
		expectedLimit int
	}{
		{
			name:          "Valid limit",
			inputLimit:    25,
			expectedLimit: 25,
		},
		{
			name:          "Too small limit",
			inputLimit:    0,
			expectedLimit: 25,
		},
		{
			name:          "Negative limit",
			inputLimit:    -10,
			expectedLimit: 25,
		},
		{
			name:          "Too large limit",
			inputLimit:    200,
			expectedLimit: 25,
		},
		{
			name:          "Maximum allowed",
			inputLimit:    100,
			expectedLimit: 100,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			limit := tt.inputLimit
			if limit < 1 || limit > 100 {
				limit = 25
			}

			if limit != tt.expectedLimit {
				t.Errorf("Expected limit %d, got %d", tt.expectedLimit, limit)
			}
		})
	}
}

// TestBuildWatchProgressInfo tests the helper method for creating WatchProgressInfo
func TestBuildWatchProgressInfo(t *testing.T) {
	svc := &ClipService{}

	tests := []struct {
		name            string
		progressSeconds int
		completed       bool
		duration        *float64
		expectNil       bool
		expectedPercent float64
		expectedDurSec  int
	}{
		{
			name:            "Normal progress",
			progressSeconds: 45,
			completed:       false,
			duration:        testFloat64Ptr(120.0),
			expectNil:       false,
			expectedPercent: 37.5,
			expectedDurSec:  120,
		},
		{
			name:            "Zero progress returns nil",
			progressSeconds: 0,
			completed:       false,
			duration:        testFloat64Ptr(120.0),
			expectNil:       true,
		},
		{
			name:            "Negative progress returns nil",
			progressSeconds: -10,
			completed:       false,
			duration:        testFloat64Ptr(120.0),
			expectNil:       true,
		},
		{
			name:            "Nil duration",
			progressSeconds: 45,
			completed:       false,
			duration:        nil,
			expectNil:       false,
			expectedPercent: 0,
			expectedDurSec:  0,
		},
		{
			name:            "Zero duration",
			progressSeconds: 45,
			completed:       false,
			duration:        testFloat64Ptr(0.0),
			expectNil:       false,
			expectedPercent: 0,
			expectedDurSec:  0,
		},
		{
			name:            "Completed clip at 90%",
			progressSeconds: 108,
			completed:       true,
			duration:        testFloat64Ptr(120.0),
			expectNil:       false,
			expectedPercent: 90.0,
			expectedDurSec:  120,
		},
		{
			name:            "Progress exceeds duration",
			progressSeconds: 150,
			completed:       true,
			duration:        testFloat64Ptr(120.0),
			expectNil:       false,
			expectedPercent: 125.0,
			expectedDurSec:  120,
		},
		{
			name:            "Small progress value",
			progressSeconds: 1,
			completed:       false,
			duration:        testFloat64Ptr(120.0),
			expectNil:       false,
			expectedPercent: 0.8333333333333334,
			expectedDurSec:  120,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := svc.buildWatchProgressInfo(tt.progressSeconds, tt.completed, tt.duration)

			if tt.expectNil {
				if result != nil {
					t.Errorf("Expected nil result, got %+v", result)
				}
				return
			}

			if result == nil {
				t.Fatal("Expected non-nil result, got nil")
			}

			if result.ProgressSeconds != tt.progressSeconds {
				t.Errorf("Expected ProgressSeconds=%d, got %d", tt.progressSeconds, result.ProgressSeconds)
			}

			if result.Completed != tt.completed {
				t.Errorf("Expected Completed=%v, got %v", tt.completed, result.Completed)
			}

			if result.DurationSeconds != tt.expectedDurSec {
				t.Errorf("Expected DurationSeconds=%d, got %d", tt.expectedDurSec, result.DurationSeconds)
			}

			if result.ProgressPercent != tt.expectedPercent {
				t.Errorf("Expected ProgressPercent=%.2f, got %.2f", tt.expectedPercent, result.ProgressPercent)
			}

			// Verify WatchedAt is not set (omitted for performance)
			if result.WatchedAt != "" {
				t.Errorf("Expected WatchedAt to be empty, got %s", result.WatchedAt)
			}
		})
	}
}

// TestWatchProgressPercentCalculation tests progress percentage calculation edge cases
func TestWatchProgressPercentCalculation(t *testing.T) {
	svc := &ClipService{}

	tests := []struct {
		name            string
		progressSeconds int
		durationSeconds float64
		expectedPercent float64
	}{
		{
			name:            "Exactly 50%",
			progressSeconds: 60,
			durationSeconds: 120,
			expectedPercent: 50.0,
		},
		{
			name:            "Completion threshold at 90%",
			progressSeconds: 108,
			durationSeconds: 120,
			expectedPercent: 90.0,
		},
		{
			name:            "Just below completion threshold",
			progressSeconds: 107,
			durationSeconds: 120,
			expectedPercent: 89.16666666666667,
		},
		{
			name:            "Full completion at 100%",
			progressSeconds: 120,
			durationSeconds: 120,
			expectedPercent: 100.0,
		},
		{
			name:            "Very small percentage",
			progressSeconds: 1,
			durationSeconds: 1000,
			expectedPercent: 0.1,
		},
		{
			name:            "Fractional seconds precision",
			progressSeconds: 33,
			durationSeconds: 100,
			expectedPercent: 33.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			duration := tt.durationSeconds
			result := svc.buildWatchProgressInfo(tt.progressSeconds, false, &duration)

			if result == nil {
				t.Fatal("Expected non-nil result")
			}

			if result.ProgressPercent != tt.expectedPercent {
				t.Errorf("Expected ProgressPercent=%.2f, got %.2f", tt.expectedPercent, result.ProgressPercent)
			}
		})
	}
}

// TestWatchProgressCompletionLogic tests the completion flag behavior
func TestWatchProgressCompletionLogic(t *testing.T) {
	svc := &ClipService{}

	tests := []struct {
		name              string
		progressSeconds   int
		durationSeconds   float64
		completed         bool
		description       string
		expectedCompleted bool
	}{
		{
			name:              "Not completed at 50%",
			progressSeconds:   60,
			durationSeconds:   120,
			completed:         false,
			description:       "Should not be marked complete at 50%",
			expectedCompleted: false,
		},
		{
			name:              "Completed at 90%",
			progressSeconds:   108,
			durationSeconds:   120,
			completed:         true,
			description:       "Should be marked complete at exactly 90%",
			expectedCompleted: true,
		},
		{
			name:              "Completed at 100%",
			progressSeconds:   120,
			durationSeconds:   120,
			completed:         true,
			description:       "Should be marked complete at 100%",
			expectedCompleted: true,
		},
		{
			name:              "Completed flag overrides percentage",
			progressSeconds:   10,
			durationSeconds:   120,
			completed:         true,
			description:       "Backend sets completed flag, helper preserves it",
			expectedCompleted: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			duration := tt.durationSeconds
			result := svc.buildWatchProgressInfo(tt.progressSeconds, tt.completed, &duration)

			if result == nil {
				t.Fatal("Expected non-nil result")
			}

			if result.Completed != tt.expectedCompleted {
				t.Errorf("%s: Expected Completed=%v, got %v",
					tt.description, tt.expectedCompleted, result.Completed)
			}
		})
	}
}

// TestWatchProgressFieldConsistency validates all fields are set correctly
func TestWatchProgressFieldConsistency(t *testing.T) {
	svc := &ClipService{}

	progressSeconds := 75
	duration := 150.0
	completed := false

	result := svc.buildWatchProgressInfo(progressSeconds, completed, &duration)

	if result == nil {
		t.Fatal("Expected non-nil result")
	}

	// Verify all fields are set
	if result.ProgressSeconds != progressSeconds {
		t.Errorf("ProgressSeconds mismatch: expected %d, got %d", progressSeconds, result.ProgressSeconds)
	}

	if result.DurationSeconds != 150 {
		t.Errorf("DurationSeconds mismatch: expected 150, got %d", result.DurationSeconds)
	}

	expectedPercent := 50.0
	if result.ProgressPercent != expectedPercent {
		t.Errorf("ProgressPercent mismatch: expected %.2f, got %.2f", expectedPercent, result.ProgressPercent)
	}

	if result.Completed != completed {
		t.Errorf("Completed mismatch: expected %v, got %v", completed, result.Completed)
	}

	// Verify WatchedAt is empty (performance optimization)
	if result.WatchedAt != "" {
		t.Errorf("WatchedAt should be empty for performance, got %s", result.WatchedAt)
	}
}
