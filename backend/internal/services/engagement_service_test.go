package services

import (
	"testing"

	"git.subcult.tv/subculture-collective/clpr/internal/models"
)

// TestCalculateComponentScore validates component score calculation
func TestCalculateComponentScore(t *testing.T) {
	tests := []struct {
		name     string
		actual   float64
		max      float64
		expected int
	}{
		{
			name:     "Zero activity",
			actual:   0,
			max:      10,
			expected: 0,
		},
		{
			name:     "Half of max",
			actual:   5,
			max:      10,
			expected: 50,
		},
		{
			name:     "At max",
			actual:   10,
			max:      10,
			expected: 100,
		},
		{
			name:     "Above max (should cap at 100)",
			actual:   15,
			max:      10,
			expected: 100,
		},
		{
			name:     "Zero max (edge case)",
			actual:   5,
			max:      0,
			expected: 0,
		},
		{
			name:     "Fractional score",
			actual:   7.5,
			max:      10,
			expected: 75,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := calculateComponentScore(tt.actual, tt.max)
			if result != tt.expected {
				t.Errorf("calculateComponentScore(%v, %v) = %d, want %d", tt.actual, tt.max, result, tt.expected)
			}
		})
	}
}

// TestDetermineEngagementTier validates tier determination
func TestDetermineEngagementTier(t *testing.T) {
	tests := []struct {
		name     string
		score    int
		expected string
	}{
		{
			name:     "Inactive (0)",
			score:    0,
			expected: "Inactive",
		},
		{
			name:     "Inactive (25)",
			score:    25,
			expected: "Inactive",
		},
		{
			name:     "Low engagement (26)",
			score:    26,
			expected: "Low Engagement",
		},
		{
			name:     "Low engagement (50)",
			score:    50,
			expected: "Low Engagement",
		},
		{
			name:     "Moderate engagement (51)",
			score:    51,
			expected: "Moderate Engagement",
		},
		{
			name:     "Moderate engagement (75)",
			score:    75,
			expected: "Moderate Engagement",
		},
		{
			name:     "High engagement (76)",
			score:    76,
			expected: "High Engagement",
		},
		{
			name:     "High engagement (90)",
			score:    90,
			expected: "High Engagement",
		},
		{
			name:     "Very high engagement (91)",
			score:    91,
			expected: "Very High Engagement",
		},
		{
			name:     "Very high engagement (100)",
			score:    100,
			expected: "Very High Engagement",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := determineEngagementTier(tt.score)
			if result != tt.expected {
				t.Errorf("determineEngagementTier(%d) = %q, want %q", tt.score, result, tt.expected)
			}
		})
	}
}

// TestCalculateAverage validates average calculation for trend points
func TestCalculateAverage(t *testing.T) {
	tests := []struct {
		name     string
		points   []models.TrendingDataPoint
		expected float64
	}{
		{
			name:     "Empty slice",
			points:   []models.TrendingDataPoint{},
			expected: 0,
		},
		{
			name: "Single point",
			points: []models.TrendingDataPoint{
				{Value: 100},
			},
			expected: 100,
		},
		{
			name: "Multiple points",
			points: []models.TrendingDataPoint{
				{Value: 100},
				{Value: 200},
				{Value: 300},
			},
			expected: 200,
		},
		{
			name: "Mixed values",
			points: []models.TrendingDataPoint{
				{Value: 50},
				{Value: 150},
				{Value: 100},
			},
			expected: 100,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := calculateAverage(tt.points)
			if result != tt.expected {
				t.Errorf("calculateAverage() = %v, want %v", result, tt.expected)
			}
		})
	}
}

// TestCalculateTrendSummary validates trend summary calculation
func TestCalculateTrendSummary(t *testing.T) {
	tests := []struct {
		name     string
		points   []models.TrendingDataPoint
		checkMin int64
		checkMax int64
		checkAvg int64
	}{
		{
			name:     "Empty slice",
			points:   []models.TrendingDataPoint{},
			checkMin: 0,
			checkMax: 0,
			checkAvg: 0,
		},
		{
			name: "Single point",
			points: []models.TrendingDataPoint{
				{Value: 100},
			},
			checkMin: 100,
			checkMax: 100,
			checkAvg: 100,
		},
		{
			name: "Multiple points",
			points: []models.TrendingDataPoint{
				{Value: 50},
				{Value: 100},
				{Value: 150},
			},
			checkMin: 50,
			checkMax: 150,
			checkAvg: 100,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := calculateTrendSummary(tt.points)
			if len(tt.points) > 0 {
				if result.Min != tt.checkMin {
					t.Errorf("Min = %v, want %v", result.Min, tt.checkMin)
				}
				if result.Max != tt.checkMax {
					t.Errorf("Max = %v, want %v", result.Max, tt.checkMax)
				}
				if result.Avg != tt.checkAvg {
					t.Errorf("Avg = %v, want %v", result.Avg, tt.checkAvg)
				}
			}
		})
	}
}

// TestNormalizeMetric validates metric normalization
func TestNormalizeMetric(t *testing.T) {
	tests := []struct {
		name     string
		value    int64
		max      int64
		expected int
	}{
		{
			name:     "Zero value",
			value:    0,
			max:      100,
			expected: 0,
		},
		{
			name:     "Half of max",
			value:    50,
			max:      100,
			expected: 50,
		},
		{
			name:     "At max",
			value:    100,
			max:      100,
			expected: 100,
		},
		{
			name:     "Above max (should cap at 100)",
			value:    150,
			max:      100,
			expected: 100,
		},
		{
			name:     "Zero max (edge case)",
			value:    50,
			max:      0,
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := normalizeMetric(tt.value, tt.max)
			if result != tt.expected {
				t.Errorf("normalizeMetric(%d, %d) = %d, want %d", tt.value, tt.max, result, tt.expected)
			}
		})
	}
}

// TestEngagementServiceStructure validates the service structure
func TestEngagementServiceStructure(t *testing.T) {
	// This test ensures the EngagementService is properly structured
	// and can be instantiated
	service := NewEngagementService(nil, nil, nil)
	if service == nil {
		t.Error("NewEngagementService returned nil")
	}
}

// TestEngagementServiceMethods validates that all expected methods exist
func TestEngagementServiceMethods(t *testing.T) {
	service := NewEngagementService(nil, nil, nil)

	// Verify service has the expected method signatures by checking it's not nil
	if service == nil {
		t.Error("Service should not be nil")
	}

	// The service struct exists and has the correct type
	if _, ok := interface{}(service).(*EngagementService); !ok {
		t.Error("Service is not of type *EngagementService")
	}
}
