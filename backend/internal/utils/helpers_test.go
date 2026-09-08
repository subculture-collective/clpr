package utils

import "testing"

func TestStringPtr(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected *string
	}{
		{
			name:     "Non-empty string",
			input:    "test",
			expected: StringPtr("test"),
		},
		{
			name:     "Empty string",
			input:    "",
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := StringPtr(tt.input)
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
			expected: Float64Ptr(30.5),
		},
		{
			name:     "Zero value",
			input:    0,
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Float64Ptr(tt.input)
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
			result := Min(tt.a, tt.b)
			if result != tt.expected {
				t.Errorf("Expected %d, got %d", tt.expected, result)
			}
		})
	}
}

func TestStringOrDefault(t *testing.T) {
	tests := []struct {
		name         string
		input        *string
		defaultValue *string
		expected     string
	}{
		{
			name:         "Non-nil pointer with non-empty string",
			input:        StringPtr("test"),
			defaultValue: StringPtr("default"),
			expected:     "test",
		},
		{
			name:         "Non-nil pointer with empty string",
			input:        func() *string { s := ""; return &s }(),
			defaultValue: StringPtr("default"),
			expected:     "default",
		},
		{
			name:         "Nil pointer",
			input:        nil,
			defaultValue: StringPtr("default"),
			expected:     "default",
		},
		{
			name:         "Nil pointer with empty default",
			input:        nil,
			defaultValue: StringPtr(""),
			expected:     "",
		},
		{
			name:         "Both nil",
			input:        nil,
			defaultValue: nil,
			expected:     "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := StringOrDefault(tt.input, tt.defaultValue)
			if result != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestSQLPlaceholder(t *testing.T) {
	tests := []struct {
		name     string
		position int
		expected string
	}{
		{
			name:     "Position 1",
			position: 1,
			expected: "$1",
		},
		{
			name:     "Position 10",
			position: 10,
			expected: "$10",
		},
		{
			name:     "Position 100",
			position: 100,
			expected: "$100",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SQLPlaceholder(tt.position)
			if result != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, result)
			}
		})
	}
}
