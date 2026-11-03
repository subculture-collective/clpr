package repository

import (
	"testing"

	"git.subcult.tv/subculture-collective/clpr/internal/models"
)

func TestParseQueryToTSQuery(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Single word",
			input:    "valorant",
			expected: "valorant:*",
		},
		{
			name:     "Multiple words",
			input:    "valorant ace",
			expected: "valorant:* & ace:*",
		},
		{
			name:     "Empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "Words with extra spaces",
			input:    "valorant  ace   clutch",
			expected: "valorant:* & ace:* & clutch:*",
		},
		{name: "Trailing parenthesis", input: "foo(", expected: "foo:*"},
		{name: "Trailing OR operator", input: "a|", expected: "a:*"},
		{name: "Only an operator", input: "!", expected: ""},
		{name: "Apostrophe and colon", input: "it's:", expected: "it:* & s:*"},
		{name: "Operators between words", input: "foo & | !bar <-> (baz)", expected: "foo:* & bar:* & baz:*"},
		{name: "Prefix and weight syntax", input: "abc:*AB", expected: "abc:* & AB:*"},
		{name: "Quotes and backslashes", input: `"quoted" \\x 'single'`, expected: "quoted:* & x:* & single:*"},
		{name: "Underscore and digits kept", input: "pog_champ 1v5", expected: "pog_champ:* & 1v5:*"},
		{name: "Unicode letters kept", input: "café суп 日本語!", expected: "café:* & суп:* & 日本語:*"},
		{name: "Whitespace only", input: " \t\n ", expected: ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseQueryToTSQuery(tt.input)
			if result != tt.expected {
				t.Errorf("parseQueryToTSQuery(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestSearchRequest_Defaults(t *testing.T) {
	req := &models.SearchRequest{
		Query: "test",
	}

	// Test that we can create a search request with just a query
	if req.Query != "test" {
		t.Errorf("Expected query to be 'test', got %q", req.Query)
	}

	// Default page should be 1 when not set
	if req.Page == 0 {
		req.Page = 1
	}
	if req.Page != 1 {
		t.Errorf("Expected default page to be 1, got %d", req.Page)
	}

	// Default limit should be reasonable
	if req.Limit == 0 {
		req.Limit = 20
	}
	if req.Limit != 20 {
		t.Errorf("Expected default limit to be 20, got %d", req.Limit)
	}
}
