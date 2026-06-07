package services

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"git.subcult.tv/subculture-collective/clpr/internal/models"
)

func TestAdService_filterByTargeting(t *testing.T) {
	s := &AdService{}

	tests := []struct {
		name     string
		ads      []models.Ad
		req      models.AdSelectionRequest
		expected int
	}{
		{
			name: "No targeting criteria - all ads pass",
			ads: []models.Ad{
				{ID: uuid.New(), Name: "Ad 1", TargetingCriteria: nil},
				{ID: uuid.New(), Name: "Ad 2", TargetingCriteria: nil},
			},
			req:      models.AdSelectionRequest{Platform: "web"},
			expected: 2,
		},
		{
			name: "Game targeting - matches",
			ads: []models.Ad{
				{
					ID:   uuid.New(),
					Name: "Ad 1",
					TargetingCriteria: map[string]interface{}{
						"game_ids": []interface{}{"game123", "game456"},
					},
				},
			},
			req: models.AdSelectionRequest{
				Platform: "web",
				GameID:   strPtr("game123"),
			},
			expected: 1,
		},
		{
			name: "Game targeting - no match",
			ads: []models.Ad{
				{
					ID:   uuid.New(),
					Name: "Ad 1",
					TargetingCriteria: map[string]interface{}{
						"game_ids": []interface{}{"game123", "game456"},
					},
				},
			},
			req: models.AdSelectionRequest{
				Platform: "web",
				GameID:   strPtr("game789"),
			},
			expected: 0,
		},
		{
			name: "Language targeting - matches",
			ads: []models.Ad{
				{
					ID:   uuid.New(),
					Name: "Ad 1",
					TargetingCriteria: map[string]interface{}{
						"languages": []interface{}{"en", "es"},
					},
				},
			},
			req: models.AdSelectionRequest{
				Platform: "web",
				Language: strPtr("en"),
			},
			expected: 1,
		},
		{
			name: "Platform targeting - matches",
			ads: []models.Ad{
				{
					ID:   uuid.New(),
					Name: "Ad 1",
					TargetingCriteria: map[string]interface{}{
						"platforms": []interface{}{"web", "ios"},
					},
				},
			},
			req: models.AdSelectionRequest{
				Platform: "web",
			},
			expected: 1,
		},
		{
			name: "Platform targeting - no match",
			ads: []models.Ad{
				{
					ID:   uuid.New(),
					Name: "Ad 1",
					TargetingCriteria: map[string]interface{}{
						"platforms": []interface{}{"ios", "android"},
					},
				},
			},
			req: models.AdSelectionRequest{
				Platform: "web",
			},
			expected: 0,
		},
		{
			name: "Mixed targeting - all conditions met",
			ads: []models.Ad{
				{
					ID:   uuid.New(),
					Name: "Ad 1",
					TargetingCriteria: map[string]interface{}{
						"game_ids":  []interface{}{"game123"},
						"languages": []interface{}{"en"},
						"platforms": []interface{}{"web"},
					},
				},
			},
			req: models.AdSelectionRequest{
				Platform: "web",
				GameID:   strPtr("game123"),
				Language: strPtr("en"),
			},
			expected: 1,
		},
		{
			name: "Game targeting - request has no game ID",
			ads: []models.Ad{
				{
					ID:   uuid.New(),
					Name: "Ad 1",
					TargetingCriteria: map[string]interface{}{
						"game_ids": []interface{}{"game123"},
					},
				},
			},
			req: models.AdSelectionRequest{
				Platform: "web",
				// GameID is nil
			},
			expected: 0, // Should not match because ad targets specific games
		},
		{
			name: "Language targeting - request has no language",
			ads: []models.Ad{
				{
					ID:   uuid.New(),
					Name: "Ad 1",
					TargetingCriteria: map[string]interface{}{
						"languages": []interface{}{"en"},
					},
				},
			},
			req: models.AdSelectionRequest{
				Platform: "web",
				// Language is nil
			},
			expected: 0, // Should not match because ad targets specific languages
		},
		// New tests for enhanced targeting
		{
			name: "Country targeting - matches",
			ads: []models.Ad{
				{
					ID:   uuid.New(),
					Name: "Ad 1",
					TargetingCriteria: map[string]interface{}{
						"countries": []interface{}{"US", "CA"},
					},
				},
			},
			req: models.AdSelectionRequest{
				Platform: "web",
				Country:  strPtr("US"),
			},
			expected: 1,
		},
		{
			name: "Country targeting - no match",
			ads: []models.Ad{
				{
					ID:   uuid.New(),
					Name: "Ad 1",
					TargetingCriteria: map[string]interface{}{
						"countries": []interface{}{"US", "CA"},
					},
				},
			},
			req: models.AdSelectionRequest{
				Platform: "web",
				Country:  strPtr("GB"),
			},
			expected: 0,
		},
		{
			name: "Device targeting - matches",
			ads: []models.Ad{
				{
					ID:   uuid.New(),
					Name: "Ad 1",
					TargetingCriteria: map[string]interface{}{
						"devices": []interface{}{"desktop", "tablet"},
					},
				},
			},
			req: models.AdSelectionRequest{
				Platform:   "web",
				DeviceType: strPtr("desktop"),
			},
			expected: 1,
		},
		{
			name: "Device targeting - no match",
			ads: []models.Ad{
				{
					ID:   uuid.New(),
					Name: "Ad 1",
					TargetingCriteria: map[string]interface{}{
						"devices": []interface{}{"mobile"},
					},
				},
			},
			req: models.AdSelectionRequest{
				Platform:   "web",
				DeviceType: strPtr("desktop"),
			},
			expected: 0,
		},
		{
			name: "Interests targeting - matches",
			ads: []models.Ad{
				{
					ID:   uuid.New(),
					Name: "Ad 1",
					TargetingCriteria: map[string]interface{}{
						"interests": []interface{}{"gaming", "esports"},
					},
				},
			},
			req: models.AdSelectionRequest{
				Platform:  "web",
				Interests: []string{"gaming", "tech"},
			},
			expected: 1,
		},
		{
			name: "Interests targeting - no match",
			ads: []models.Ad{
				{
					ID:   uuid.New(),
					Name: "Ad 1",
					TargetingCriteria: map[string]interface{}{
						"interests": []interface{}{"fashion", "sports"},
					},
				},
			},
			req: models.AdSelectionRequest{
				Platform:  "web",
				Interests: []string{"gaming", "tech"},
			},
			expected: 0,
		},
		{
			name: "Interests targeting - request has no interests",
			ads: []models.Ad{
				{
					ID:   uuid.New(),
					Name: "Ad 1",
					TargetingCriteria: map[string]interface{}{
						"interests": []interface{}{"gaming"},
					},
				},
			},
			req: models.AdSelectionRequest{
				Platform:  "web",
				Interests: []string{}, // empty
			},
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := s.filterByTargeting(tt.ads, tt.req)
			assert.Equal(t, tt.expected, len(result))
		})
	}
}

func TestAdService_weightedRandomSelect(t *testing.T) {
	s := &AdService{}

	t.Run("Single ad returns that ad", func(t *testing.T) {
		ad := models.Ad{ID: uuid.New(), Name: "Single Ad", Priority: 1, Weight: 50}
		result := s.weightedRandomSelect([]models.Ad{ad})
		assert.Equal(t, ad.ID, result.ID)
	})

	t.Run("Highest priority ad is selected from different priorities", func(t *testing.T) {
		ads := []models.Ad{
			{ID: uuid.New(), Name: "Low Priority", Priority: 1, Weight: 100},
			{ID: uuid.New(), Name: "High Priority", Priority: 10, Weight: 1},
		}
		// High priority should always be selected
		result := s.weightedRandomSelect(ads)
		assert.Equal(t, "High Priority", result.Name)
	})

	t.Run("Same priority uses weighted random", func(t *testing.T) {
		ads := []models.Ad{
			{ID: uuid.New(), Name: "Ad 1", Priority: 5, Weight: 50},
			{ID: uuid.New(), Name: "Ad 2", Priority: 5, Weight: 50},
		}
		// Run multiple times to verify both can be selected
		selections := make(map[string]int)
		for i := 0; i < 100; i++ {
			result := s.weightedRandomSelect(ads)
			selections[result.Name]++
		}
		// Both should be selected at least once (probabilistically)
		assert.True(t, selections["Ad 1"] > 0 && selections["Ad 2"] > 0)
	})
}

func TestAdService_calculateWindowStart(t *testing.T) {
	s := &AdService{}

	t.Run("Hourly window truncates to hour", func(t *testing.T) {
		start := s.calculateWindowStart(models.FrequencyWindowHourly)
		assert.Equal(t, 0, start.Minute())
		assert.Equal(t, 0, start.Second())
	})

	t.Run("Daily window starts at midnight", func(t *testing.T) {
		start := s.calculateWindowStart(models.FrequencyWindowDaily)
		assert.Equal(t, 0, start.Hour())
		assert.Equal(t, 0, start.Minute())
	})

	t.Run("Weekly window starts on Sunday", func(t *testing.T) {
		start := s.calculateWindowStart(models.FrequencyWindowWeekly)
		assert.Equal(t, time.Sunday, start.Weekday())
	})

	t.Run("Lifetime window returns zero time", func(t *testing.T) {
		start := s.calculateWindowStart(models.FrequencyWindowLifetime)
		assert.True(t, start.IsZero())
	})
}

func TestViewabilityThreshold(t *testing.T) {
	t.Run("Threshold is set correctly", func(t *testing.T) {
		assert.Equal(t, 1000, models.ViewabilityThresholdMs)
	})
}

func TestAdService_evaluateTargetingRule(t *testing.T) {
	s := &AdService{}

	tests := []struct {
		name     string
		rule     models.AdTargetingRule
		req      models.AdSelectionRequest
		expected bool
	}{
		{
			name: "Country rule - matches",
			rule: models.AdTargetingRule{
				RuleType: models.TargetingRuleTypeCountry,
				Operator: models.TargetingOperatorInclude,
				Values:   []string{"US", "CA"},
			},
			req: models.AdSelectionRequest{
				Platform: "web",
				Country:  strPtr("US"),
			},
			expected: true,
		},
		{
			name: "Country rule - no match",
			rule: models.AdTargetingRule{
				RuleType: models.TargetingRuleTypeCountry,
				Operator: models.TargetingOperatorInclude,
				Values:   []string{"US", "CA"},
			},
			req: models.AdSelectionRequest{
				Platform: "web",
				Country:  strPtr("GB"),
			},
			expected: false,
		},
		{
			name: "Device rule - matches",
			rule: models.AdTargetingRule{
				RuleType: models.TargetingRuleTypeDevice,
				Operator: models.TargetingOperatorInclude,
				Values:   []string{"desktop", "tablet"},
			},
			req: models.AdSelectionRequest{
				Platform:   "web",
				DeviceType: strPtr("desktop"),
			},
			expected: true,
		},
		{
			name: "Interest rule - matches one of many",
			rule: models.AdTargetingRule{
				RuleType: models.TargetingRuleTypeInterest,
				Operator: models.TargetingOperatorInclude,
				Values:   []string{"gaming", "esports"},
			},
			req: models.AdSelectionRequest{
				Platform:  "web",
				Interests: []string{"tech", "gaming", "music"},
			},
			expected: true,
		},
		{
			name: "Interest rule - no match",
			rule: models.AdTargetingRule{
				RuleType: models.TargetingRuleTypeInterest,
				Operator: models.TargetingOperatorInclude,
				Values:   []string{"fashion", "sports"},
			},
			req: models.AdSelectionRequest{
				Platform:  "web",
				Interests: []string{"tech", "gaming"},
			},
			expected: false,
		},
		{
			name: "Platform rule - matches",
			rule: models.AdTargetingRule{
				RuleType: models.TargetingRuleTypePlatform,
				Operator: models.TargetingOperatorInclude,
				Values:   []string{"web", "ios"},
			},
			req: models.AdSelectionRequest{
				Platform: "web",
			},
			expected: true,
		},
		{
			name: "Language rule - matches",
			rule: models.AdTargetingRule{
				RuleType: models.TargetingRuleTypeLanguage,
				Operator: models.TargetingOperatorInclude,
				Values:   []string{"en", "es"},
			},
			req: models.AdSelectionRequest{
				Platform: "web",
				Language: strPtr("en"),
			},
			expected: true,
		},
		{
			name: "Game rule - matches",
			rule: models.AdTargetingRule{
				RuleType: models.TargetingRuleTypeGame,
				Operator: models.TargetingOperatorInclude,
				Values:   []string{"game123", "game456"},
			},
			req: models.AdSelectionRequest{
				Platform: "web",
				GameID:   strPtr("game123"),
			},
			expected: true,
		},
		// Exclude operator tests
		{
			name: "Country rule - exclude matches (should filter out)",
			rule: models.AdTargetingRule{
				RuleType: models.TargetingRuleTypeCountry,
				Operator: models.TargetingOperatorExclude,
				Values:   []string{"US", "CA"},
			},
			req: models.AdSelectionRequest{
				Platform: "web",
				Country:  strPtr("US"),
			},
			expected: true, // Rule matches, so exclude operator will filter it out
		},
		{
			name: "Country rule - exclude no match (should include)",
			rule: models.AdTargetingRule{
				RuleType: models.TargetingRuleTypeCountry,
				Operator: models.TargetingOperatorExclude,
				Values:   []string{"US", "CA"},
			},
			req: models.AdSelectionRequest{
				Platform: "web",
				Country:  strPtr("GB"),
			},
			expected: false, // Rule does not match, so exclude operator keeps it
		},
		{
			name: "Device rule - exclude matches",
			rule: models.AdTargetingRule{
				RuleType: models.TargetingRuleTypeDevice,
				Operator: models.TargetingOperatorExclude,
				Values:   []string{"mobile"},
			},
			req: models.AdSelectionRequest{
				Platform:   "web",
				DeviceType: strPtr("mobile"),
			},
			expected: true, // Rule matches for exclude
		},
		{
			name: "Interest rule - exclude matches one",
			rule: models.AdTargetingRule{
				RuleType: models.TargetingRuleTypeInterest,
				Operator: models.TargetingOperatorExclude,
				Values:   []string{"gambling", "adult"},
			},
			req: models.AdSelectionRequest{
				Platform:  "web",
				Interests: []string{"tech", "gambling"},
			},
			expected: true, // Rule matches for exclude
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := s.evaluateTargetingRule(tt.rule, tt.req)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestAdService_selectAdWithExperiment(t *testing.T) {
	s := &AdService{}

	t.Run("Non-experiment ads use weighted random selection", func(t *testing.T) {
		ads := []models.Ad{
			{ID: uuid.New(), Name: "Ad 1", Priority: 5, Weight: 50},
			{ID: uuid.New(), Name: "Ad 2", Priority: 5, Weight: 50},
		}
		result := s.selectAdWithExperiment(ads, nil, nil)
		assert.NotEmpty(t, result.ID)
	})

	t.Run("Experiment ads are selected consistently for same user", func(t *testing.T) {
		experimentID := uuid.New()
		variantA := "control"
		variantB := "variant_a"
		ads := []models.Ad{
			{ID: uuid.New(), Name: "Control", Priority: 5, Weight: 50, ExperimentID: &experimentID, ExperimentVariant: &variantA},
			{ID: uuid.New(), Name: "Variant A", Priority: 5, Weight: 50, ExperimentID: &experimentID, ExperimentVariant: &variantB},
		}
		userID := uuid.New()

		// Same user should get consistent variant
		result1 := s.selectAdWithExperiment(ads, &userID, nil)
		result2 := s.selectAdWithExperiment(ads, &userID, nil)
		assert.Equal(t, result1.ID, result2.ID)
	})

	t.Run("Different users may get different variants", func(t *testing.T) {
		experimentID := uuid.New()
		variantA := "control"
		variantB := "variant_a"
		ads := []models.Ad{
			{ID: uuid.New(), Name: "Control", Priority: 5, Weight: 50, ExperimentID: &experimentID, ExperimentVariant: &variantA},
			{ID: uuid.New(), Name: "Variant A", Priority: 5, Weight: 50, ExperimentID: &experimentID, ExperimentVariant: &variantB},
		}

		// Different users should potentially get different variants
		variants := make(map[string]int)
		for i := 0; i < 100; i++ {
			userID := uuid.New()
			result := s.selectAdWithExperiment(ads, &userID, nil)
			if result.ExperimentVariant != nil {
				variants[*result.ExperimentVariant]++
			}
		}
		// Both variants should be represented
		assert.True(t, variants["control"] > 0 || variants["variant_a"] > 0)
	})
}

func TestContainsString(t *testing.T) {
	t.Run("Contains element", func(t *testing.T) {
		result := containsString([]string{"a", "b", "c"}, "b")
		assert.True(t, result)
	})

	t.Run("Does not contain element", func(t *testing.T) {
		result := containsString([]string{"a", "b", "c"}, "d")
		assert.False(t, result)
	})

	t.Run("Empty slice", func(t *testing.T) {
		result := containsString([]string{}, "a")
		assert.False(t, result)
	})
}

func TestTargetingRuleTypeConstants(t *testing.T) {
	t.Run("Rule type constants are defined", func(t *testing.T) {
		assert.Equal(t, "country", models.TargetingRuleTypeCountry)
		assert.Equal(t, "device", models.TargetingRuleTypeDevice)
		assert.Equal(t, "interest", models.TargetingRuleTypeInterest)
		assert.Equal(t, "platform", models.TargetingRuleTypePlatform)
		assert.Equal(t, "language", models.TargetingRuleTypeLanguage)
		assert.Equal(t, "game", models.TargetingRuleTypeGame)
	})
}

func TestTargetingOperatorConstants(t *testing.T) {
	t.Run("Operator constants are defined", func(t *testing.T) {
		assert.Equal(t, "include", models.TargetingOperatorInclude)
		assert.Equal(t, "exclude", models.TargetingOperatorExclude)
	})
}

func TestExperimentStatusConstants(t *testing.T) {
	t.Run("Experiment status constants are defined", func(t *testing.T) {
		assert.Equal(t, "draft", models.ExperimentStatusDraft)
		assert.Equal(t, "running", models.ExperimentStatusRunning)
		assert.Equal(t, "paused", models.ExperimentStatusPaused)
		assert.Equal(t, "completed", models.ExperimentStatusCompleted)
	})
}

func TestAdService_filterByContextualTargeting(t *testing.T) {
	s := &AdService{}

	tests := []struct {
		name     string
		ads      []models.Ad
		req      models.AdSelectionRequest
		expected int
	}{
		{
			name: "No targeting criteria - all ads pass",
			ads: []models.Ad{
				{ID: uuid.New(), Name: "Ad 1", TargetingCriteria: nil},
				{ID: uuid.New(), Name: "Ad 2", TargetingCriteria: nil},
			},
			req:      models.AdSelectionRequest{Platform: "web"},
			expected: 2,
		},
		{
			name: "Game targeting - matches (contextual)",
			ads: []models.Ad{
				{
					ID:   uuid.New(),
					Name: "Ad 1",
					TargetingCriteria: map[string]interface{}{
						"game_ids": []interface{}{"game123", "game456"},
					},
				},
			},
			req: models.AdSelectionRequest{
				Platform: "web",
				GameID:   strPtr("game123"),
			},
			expected: 1,
		},
		{
			name: "Language targeting - matches (contextual)",
			ads: []models.Ad{
				{
					ID:   uuid.New(),
					Name: "Ad 1",
					TargetingCriteria: map[string]interface{}{
						"languages": []interface{}{"en", "es"},
					},
				},
			},
			req: models.AdSelectionRequest{
				Platform: "web",
				Language: strPtr("en"),
			},
			expected: 1,
		},
		{
			name: "Platform targeting - matches (contextual)",
			ads: []models.Ad{
				{
					ID:   uuid.New(),
					Name: "Ad 1",
					TargetingCriteria: map[string]interface{}{
						"platforms": []interface{}{"web", "ios"},
					},
				},
			},
			req: models.AdSelectionRequest{
				Platform: "web",
			},
			expected: 1,
		},
		{
			name: "Country targeting - IGNORED in contextual mode (privacy)",
			ads: []models.Ad{
				{
					ID:   uuid.New(),
					Name: "Ad 1",
					TargetingCriteria: map[string]interface{}{
						"countries": []interface{}{"US"},
					},
				},
			},
			req: models.AdSelectionRequest{
				Platform: "web",
				Country:  strPtr("CA"), // Different country, but should still pass in contextual
			},
			expected: 1, // Passes because country targeting is skipped
		},
		{
			name: "Device targeting - IGNORED in contextual mode (privacy)",
			ads: []models.Ad{
				{
					ID:   uuid.New(),
					Name: "Ad 1",
					TargetingCriteria: map[string]interface{}{
						"devices": []interface{}{"mobile"},
					},
				},
			},
			req: models.AdSelectionRequest{
				Platform:   "web",
				DeviceType: strPtr("desktop"), // Different device, but should still pass
			},
			expected: 1, // Passes because device targeting is skipped
		},
		{
			name: "Interests targeting - IGNORED in contextual mode (privacy)",
			ads: []models.Ad{
				{
					ID:   uuid.New(),
					Name: "Ad 1",
					TargetingCriteria: map[string]interface{}{
						"interests": []interface{}{"gaming"},
					},
				},
			},
			req: models.AdSelectionRequest{
				Platform:  "web",
				Interests: []string{"cooking"}, // Different interests, but should still pass
			},
			expected: 1, // Passes because interests targeting is skipped
		},
		{
			name: "Mixed contextual and user-specific targeting - only contextual applied",
			ads: []models.Ad{
				{
					ID:   uuid.New(),
					Name: "Ad 1",
					TargetingCriteria: map[string]interface{}{
						"game_ids":  []interface{}{"game123"},
						"countries": []interface{}{"US"},
						"interests": []interface{}{"gaming"},
					},
				},
			},
			req: models.AdSelectionRequest{
				Platform:  "web",
				GameID:    strPtr("game123"), // Matches contextual targeting
				Country:   strPtr("CA"),      // Would fail user-specific, but ignored
				Interests: []string{},        // Would fail user-specific, but ignored
			},
			expected: 1, // Passes because only game_ids is checked
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := s.filterByContextualTargeting(tt.ads, tt.req)
			assert.Equal(t, tt.expected, len(result))
		})
	}
}

func TestAdService_ValidateCreative(t *testing.T) {
	s := &AdService{}

	tests := []struct {
		name       string
		contentURL string
		adType     string
		width      *int
		height     *int
		expectErr  bool
		errMsg     string
	}{
		{
			name:       "Valid banner with standard size",
			contentURL: "https://example.com/ad.jpg",
			adType:     "banner",
			width:      intPtr(728),
			height:     intPtr(90),
			expectErr:  false,
		},
		{
			name:       "Valid banner with medium rectangle",
			contentURL: "https://example.com/ad.jpg",
			adType:     "banner",
			width:      intPtr(300),
			height:     intPtr(250),
			expectErr:  false,
		},
		{
			name:       "Invalid banner size",
			contentURL: "https://example.com/ad.jpg",
			adType:     "banner",
			width:      intPtr(100),
			height:     intPtr(100),
			expectErr:  true,
			errMsg:     "invalid banner size",
		},
		{
			name:       "Banner without dimensions",
			contentURL: "https://example.com/ad.jpg",
			adType:     "banner",
			width:      nil,
			height:     nil,
			expectErr:  true,
			errMsg:     "width and height are required for banner ads",
		},
		{
			name:       "Valid video (no dimensions required)",
			contentURL: "https://example.com/video.mp4",
			adType:     "video",
			width:      nil,
			height:     nil,
			expectErr:  false,
		},
		{
			name:       "Valid native ad",
			contentURL: "https://example.com/native.json",
			adType:     "native",
			width:      nil,
			height:     nil,
			expectErr:  false,
		},
		{
			name:       "Invalid ad type",
			contentURL: "https://example.com/ad.jpg",
			adType:     "invalid",
			width:      nil,
			height:     nil,
			expectErr:  true,
			errMsg:     "invalid ad type",
		},
		{
			name:       "Empty content URL",
			contentURL: "",
			adType:     "banner",
			width:      intPtr(728),
			height:     intPtr(90),
			expectErr:  true,
			errMsg:     "content URL is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := s.ValidateCreative(context.TODO(), tt.contentURL, tt.adType, tt.width, tt.height)
			if tt.expectErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestAdService_CreateCampaignValidation(t *testing.T) {
	s := &AdService{}

	tests := []struct {
		name      string
		ad        *models.Ad
		expectErr bool
		errMsg    string
	}{
		{
			name: "Valid campaign - all fields",
			ad: &models.Ad{
				Name:           "Test Campaign",
				AdvertiserName: "Test Advertiser",
				AdType:         "banner",
				ContentURL:     "https://example.com/ad.jpg",
				Width:          intPtr(728),
				Height:         intPtr(90),
			},
			expectErr: false,
		},
		{
			name: "Missing name",
			ad: &models.Ad{
				Name:           "",
				AdvertiserName: "Test Advertiser",
				AdType:         "banner",
				ContentURL:     "https://example.com/ad.jpg",
			},
			expectErr: true,
			errMsg:    "campaign name is required",
		},
		{
			name: "Missing advertiser name",
			ad: &models.Ad{
				Name:           "Test Campaign",
				AdvertiserName: "",
				AdType:         "banner",
				ContentURL:     "https://example.com/ad.jpg",
			},
			expectErr: true,
			errMsg:    "advertiser name is required",
		},
		{
			name: "Missing ad type",
			ad: &models.Ad{
				Name:           "Test Campaign",
				AdvertiserName: "Test Advertiser",
				AdType:         "",
				ContentURL:     "https://example.com/ad.jpg",
			},
			expectErr: true,
			errMsg:    "ad type is required",
		},
		{
			name: "Invalid ad type",
			ad: &models.Ad{
				Name:           "Test Campaign",
				AdvertiserName: "Test Advertiser",
				AdType:         "invalid",
				ContentURL:     "https://example.com/ad.jpg",
			},
			expectErr: true,
			errMsg:    "invalid ad type",
		},
		{
			name: "Missing content URL",
			ad: &models.Ad{
				Name:           "Test Campaign",
				AdvertiserName: "Test Advertiser",
				AdType:         "banner",
				ContentURL:     "",
			},
			expectErr: true,
			errMsg:    "content URL is required",
		},
		{
			name: "End date before start date",
			ad: &models.Ad{
				Name:           "Test Campaign",
				AdvertiserName: "Test Advertiser",
				AdType:         "video",
				ContentURL:     "https://example.com/video.mp4",
				StartDate:      testTimePtr(time.Now().Add(48 * time.Hour)),
				EndDate:        testTimePtr(time.Now().Add(24 * time.Hour)),
			},
			expectErr: true,
			errMsg:    "end date must be after start date",
		},
		{
			name: "Valid video campaign",
			ad: &models.Ad{
				Name:           "Video Campaign",
				AdvertiserName: "Test Advertiser",
				AdType:         "video",
				ContentURL:     "https://example.com/video.mp4",
			},
			expectErr: false,
		},
		{
			name: "Valid native campaign",
			ad: &models.Ad{
				Name:           "Native Campaign",
				AdvertiserName: "Test Advertiser",
				AdType:         "native",
				ContentURL:     "https://example.com/native.json",
			},
			expectErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test ValidateCampaign directly to avoid needing a mock repository
			err := s.ValidateCampaign(tt.ad)
			if tt.expectErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// Helper function to create int pointer
func intPtr(i int) *int {
	return &i
}

// Helper function to create time pointer for tests
func testTimePtr(t time.Time) *time.Time {
	return &t
}
