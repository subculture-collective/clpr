package services

import (
	"testing"

	"git.subcult.tv/subculture-collective/clpr/internal/models"
)

func TestOpenSearchService_BuildClipQuery(t *testing.T) {
	service := &OpenSearchService{}

	t.Run("Basic text search", func(t *testing.T) {
		req := &models.SearchRequest{
			Query: "test query",
		}
		query := service.buildClipQuery(req)

		// Verify query structure
		if query == nil {
			t.Fatal("Expected query to be non-nil")
		}

		boolQuery, ok := query["bool"].(map[string]interface{})
		if !ok {
			t.Fatalf("Expected bool query, got %T", query["bool"])
		}

		must, ok := boolQuery["must"].([]map[string]interface{})
		if !ok {
			t.Fatalf("Expected must clause as []map[string]interface{}, got %T", boolQuery["must"])
		}

		if len(must) != 1 {
			t.Fatalf("Expected 1 must clause, got %d", len(must))
		}

		// Check for multi_match query
		if _, ok := must[0]["multi_match"]; !ok {
			t.Error("Expected multi_match query")
		}
	})

	t.Run("With filters", func(t *testing.T) {
		gameID := "12345"
		language := "en"
		minVotes := 10

		req := &models.SearchRequest{
			Query:    "test",
			GameID:   &gameID,
			Language: &language,
			MinVotes: &minVotes,
		}

		query := service.buildClipQuery(req)

		boolQuery, ok := query["bool"].(map[string]interface{})
		if !ok {
			t.Fatalf("Expected bool query, got %T", query["bool"])
		}

		filter, ok := boolQuery["filter"].([]map[string]interface{})
		if !ok {
			t.Fatalf("Expected filter as []map[string]interface{}, got %T", boolQuery["filter"])
		}

		// Should have at least: is_removed + game_id + language + min_votes = 4 filters
		if len(filter) < 4 {
			t.Fatalf("Expected at least 4 filter clauses, got %d", len(filter))
		}
	})

	t.Run("Empty query uses match_all", func(t *testing.T) {
		req := &models.SearchRequest{
			Query: "",
		}

		query := service.buildClipQuery(req)
		boolQuery := query["bool"].(map[string]interface{})
		must := boolQuery["must"].([]map[string]interface{})

		if len(must) != 1 {
			t.Fatalf("Expected 1 must clause, got %d", len(must))
		}

		if _, ok := must[0]["match_all"]; !ok {
			t.Error("Expected match_all query for empty search")
		}
	})
}

func TestOpenSearchService_BuildSortClause(t *testing.T) {
	service := &OpenSearchService{}

	tests := []struct {
		name     string
		sortType string
		expected string // Expected primary sort field
	}{
		{
			name:     "Popular sort",
			sortType: "popular",
			expected: "vote_score",
		},
		{
			name:     "Recent sort",
			sortType: "recent",
			expected: "created_at",
		},
		{
			name:     "Relevance sort (default)",
			sortType: "relevance",
			expected: "_score",
		},
		{
			name:     "Empty defaults to relevance",
			sortType: "",
			expected: "_score",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.buildSortClause(tt.sortType)

			if len(result) == 0 {
				t.Fatal("Expected sort clause to be non-empty")
			}

			// Check first sort field
			firstSort := result[0]
			if _, ok := firstSort[tt.expected]; !ok {
				t.Errorf("Expected primary sort field to be %s, got %v", tt.expected, firstSort)
			}
		})
	}
}

func TestOpenSearchService_BuildUserQuery(t *testing.T) {
	service := &OpenSearchService{}

	t.Run("User search with query", func(t *testing.T) {
		req := &models.SearchRequest{
			Query: "testuser",
		}

		query := service.buildUserQuery(req)
		boolQuery := query["bool"].(map[string]interface{})
		filter := boolQuery["filter"].([]map[string]interface{})

		// Should have is_banned filter
		foundBannedFilter := false
		for _, f := range filter {
			if term, ok := f["term"].(map[string]interface{}); ok {
				if _, exists := term["is_banned"]; exists {
					foundBannedFilter = true
					break
				}
			}
		}

		if !foundBannedFilter {
			t.Error("Expected is_banned filter in user query")
		}
	})
}

func TestOpenSearchService_BuildGameQuery(t *testing.T) {
	service := &OpenSearchService{}

	t.Run("Game search with query", func(t *testing.T) {
		req := &models.SearchRequest{
			Query: "valorant",
		}

		query := service.buildGameQuery(req)

		if match, ok := query["match"].(map[string]interface{}); !ok {
			t.Error("Expected match query for game search")
		} else if _, ok := match["name"]; !ok {
			t.Error("Expected match on 'name' field")
		}
	})

	t.Run("Empty game query uses match_all", func(t *testing.T) {
		req := &models.SearchRequest{
			Query: "",
		}

		query := service.buildGameQuery(req)

		if _, ok := query["match_all"]; !ok {
			t.Error("Expected match_all for empty game query")
		}
	})
}

func TestOpenSearchService_BuildTagQuery(t *testing.T) {
	service := &OpenSearchService{}

	t.Run("Tag search with query", func(t *testing.T) {
		req := &models.SearchRequest{
			Query: "funny",
		}

		query := service.buildTagQuery(req)

		if multiMatch, ok := query["multi_match"].(map[string]interface{}); !ok {
			t.Error("Expected multi_match query for tag search")
		} else {
			fields, ok := multiMatch["fields"].([]string)
			if !ok {
				t.Error("Expected fields array")
			}
			if len(fields) == 0 {
				t.Error("Expected at least one search field")
			}
		}
	})
}
