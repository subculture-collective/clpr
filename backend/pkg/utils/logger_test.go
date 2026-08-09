package utils

import (
	"reflect"
	"strings"
	"testing"
)

func TestRedactPII(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Redact email",
			input:    "User email is test@example.com",
			expected: "User email is [REDACTED_EMAIL]",
		},
		{
			name:     "Redact phone number",
			input:    "Call me at 555-123-4567",
			expected: "Call me at [REDACTED_PHONE]",
		},
		{
			name:     "Redact credit card",
			input:    "Card number: 4111-1111-1111-1111",
			expected: "Card number: [REDACTED_CARD]",
		},
		{
			name:     "Redact password",
			input:    `{"password":"secret123"}`,
			expected: `{"password":"[REDACTED]"}`,
		},
		{
			name:     "Redact Bearer token",
			input:    "Authorization: Bearer abc123def456",
			expected: "Authorization: Bearer [REDACTED_TOKEN]",
		},
		{
			name:     "No PII",
			input:    "This is a normal log message",
			expected: "This is a normal log message",
		},
		{
			name:     "Redact OAuth callback query",
			input:    "GET /callback?code=oauth-value&state=session-state",
			expected: "GET /callback?code=[REDACTED]&state=[REDACTED]",
		},
		{
			name:     "Redact cookie header",
			input:    "Cookie: access_token=secret; csrf=value",
			expected: "Cookie: [REDACTED]",
		},
		{
			name:     "Redact private key material",
			input:    "key=-----BEGIN PRIVATE KEY-----\nsecret\n-----END PRIVATE KEY-----",
			expected: "key=[REDACTED_PRIVATE_KEY]",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := RedactPII(tt.input)
			if result != tt.expected {
				t.Errorf("RedactPII() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestRedactPIIFromFields(t *testing.T) {
	tests := []struct {
		name     string
		input    map[string]interface{}
		expected map[string]interface{}
	}{
		{
			name: "Redact password field",
			input: map[string]interface{}{
				"username": "john",
				"password": "secret123",
			},
			expected: map[string]interface{}{
				"username": "john",
				"password": "[REDACTED]",
			},
		},
		{
			name: "Redact email in string value",
			input: map[string]interface{}{
				"message": "Contact user@example.com",
			},
			expected: map[string]interface{}{
				"message": "Contact [REDACTED_EMAIL]",
			},
		},
		{
			name: "Redact token field",
			input: map[string]interface{}{
				"user_id":      123,
				"access_token": "abc123",
			},
			expected: map[string]interface{}{
				"user_id":      123,
				"access_token": "[REDACTED]",
			},
		},
		{
			name: "No sensitive fields",
			input: map[string]interface{}{
				"name":    "John Doe",
				"user_id": 123,
			},
			expected: map[string]interface{}{
				"name":    "John Doe",
				"user_id": 123,
			},
		},
		{
			name: "Redact nested and list credentials",
			input: map[string]interface{}{
				"nested": map[string]interface{}{"cookie": "session=secret"},
				"events": []interface{}{map[string]interface{}{"code": "oauth-code"}, "user@example.com"},
			},
			expected: map[string]interface{}{
				"nested": map[string]interface{}{"cookie": "[REDACTED]"},
				"events": []interface{}{map[string]interface{}{"code": "[REDACTED]"}, "[REDACTED_EMAIL]"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := RedactPIIFromFields(tt.input)

			// Check each field
			for key, expectedValue := range tt.expected {
				actualValue, ok := result[key]
				if !ok {
					t.Errorf("Field %s missing in result", key)
					continue
				}

				// Compare values
				if !reflect.DeepEqual(actualValue, expectedValue) {
					t.Errorf("Field %s = %v, want %v", key, actualValue, expectedValue)
				}
			}
		})
	}
}

func TestLogLevels(t *testing.T) {
	logger := NewStructuredLogger(LogLevelInfo)

	// Test that shouldLog correctly filters levels
	if !logger.shouldLog(LogLevelInfo) {
		t.Error("Should log INFO when min level is INFO")
	}

	if !logger.shouldLog(LogLevelError) {
		t.Error("Should log ERROR when min level is INFO")
	}

	if logger.shouldLog(LogLevelDebug) {
		t.Error("Should not log DEBUG when min level is INFO")
	}

	if !logger.shouldLog(LogLevelFatal) {
		t.Error("Should log FATAL when min level is INFO")
	}
}

func TestHashForLogging(t *testing.T) {
	// Test that hashing is consistent
	value := "test@example.com"
	hash1 := hashForLogging(value)
	hash2 := hashForLogging(value)

	if hash1 != hash2 {
		t.Error("Hash should be consistent for the same value")
	}

	// Test that different values produce different hashes
	hash3 := hashForLogging("different@example.com")
	if hash1 == hash3 {
		t.Error("Different values should produce different hashes")
	}

	// Test that hash doesn't contain original value
	if strings.Contains(hash1, "test") || strings.Contains(hash1, "example") {
		t.Error("Hash should not contain parts of original value")
	}
}
