package models

import (
	"encoding/json"
	"testing"
)

func TestSearchResponseAlwaysSerializesResultCollectionsAsArrays(t *testing.T) {
	response := SearchResponse{Results: EmptySearchResults()}

	payload, err := json.Marshal(response)
	if err != nil {
		t.Fatalf("marshal search response: %v", err)
	}

	var decoded struct {
		Results map[string]json.RawMessage `json:"results"`
	}
	if err := json.Unmarshal(payload, &decoded); err != nil {
		t.Fatalf("decode search response: %v", err)
	}

	for _, field := range []string{"clips", "creators", "games", "tags"} {
		value, ok := decoded.Results[field]
		if !ok {
			t.Errorf("results.%s is missing", field)
			continue
		}
		if string(value) != "[]" {
			t.Errorf("results.%s = %s, want []", field, value)
		}
	}
}

func TestSearchResultsNormalizeLegacyNilCollections(t *testing.T) {
	var results SearchResultsByType
	results.Normalize()

	payload, err := json.Marshal(results)
	if err != nil {
		t.Fatalf("marshal normalized search results: %v", err)
	}
	if string(payload) != `{"clips":[],"creators":[],"games":[],"tags":[]}` {
		t.Fatalf("normalized search results = %s", payload)
	}
}
