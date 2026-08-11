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

	for _, field := range []string{"clips", "creators", "games", "twitch_categories", "tags"} {
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
	if string(payload) != `{"clips":[],"creators":[],"games":[],"twitch_categories":[],"tags":[]}` {
		t.Fatalf("normalized search results = %s", payload)
	}
}

func TestClipSerializesTwitchCategoryAliasesAlongsideLegacyFields(t *testing.T) {
	categoryID, categoryName := "509658", "Just Chatting"
	payload, err := json.Marshal(Clip{GameID: &categoryID, GameName: &categoryName})
	if err != nil {
		t.Fatalf("marshal clip: %v", err)
	}
	var decoded map[string]any
	if err := json.Unmarshal(payload, &decoded); err != nil {
		t.Fatalf("decode clip: %v", err)
	}
	for _, field := range []string{"game_id", "game_name", "twitch_category_id", "twitch_category_name"} {
		if decoded[field] == nil {
			t.Errorf("clip.%s is missing", field)
		}
	}
}

func TestEmbeddedClipResponseRetainsOuterFieldsAndCategoryAliases(t *testing.T) {
	categoryID := "509658"
	payload, err := json.Marshal(ClipWithSubmitter{
		Clip:        Clip{GameID: &categoryID},
		SubmittedBy: &ClipSubmitterInfo{Username: "creator"},
	})
	if err != nil {
		t.Fatalf("marshal clip with submitter: %v", err)
	}
	var decoded map[string]any
	if err := json.Unmarshal(payload, &decoded); err != nil {
		t.Fatalf("decode clip with submitter: %v", err)
	}
	if decoded["submitted_by"] == nil || decoded["twitch_category_id"] != categoryID {
		t.Fatalf("embedded clip response lost fields: %s", payload)
	}
}

func TestAllEmbeddedClipResponsesRetainCompatibilityAliases(t *testing.T) {
	categoryID := "509658"
	for name, value := range map[string]any{
		"hot score":      ClipWithHotScore{Clip: Clip{GameID: &categoryID}, HotScore: 3},
		"creator top":    CreatorTopClip{Clip: Clip{GameID: &categoryID}, Views: 4},
		"recommendation": ClipRecommendation{Clip: Clip{GameID: &categoryID}, Score: .8},
		"playlist ref":   PlaylistClipRef{Clip: Clip{GameID: &categoryID}, OrderIndex: 2},
	} {
		t.Run(name, func(t *testing.T) {
			payload, err := json.Marshal(value)
			if err != nil {
				t.Fatalf("marshal: %v", err)
			}
			var decoded map[string]any
			if err := json.Unmarshal(payload, &decoded); err != nil {
				t.Fatalf("decode: %v", err)
			}
			if decoded["twitch_category_id"] != categoryID || decoded["game_id"] != categoryID {
				t.Fatalf("category aliases missing: %s", payload)
			}
		})
	}
}
