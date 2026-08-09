package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"git.subcult.tv/subculture-collective/clpr/internal/models"
	"github.com/gin-gonic/gin"
)

type searchContractProvider struct {
	response *models.SearchResponse
	err      error
	request  *models.SearchRequest
}

func (p *searchContractProvider) Search(_ context.Context, request *models.SearchRequest) (*models.SearchResponse, error) {
	copy := *request
	p.request = &copy
	return p.response, p.err
}

func (p *searchContractProvider) SearchWithScores(context.Context, *models.SearchRequest) (*models.SearchResponseWithScores, error) {
	return nil, errors.New("not implemented by search contract fixture")
}

func performSearchRequest(t *testing.T, provider *searchContractProvider, target string) *httptest.ResponseRecorder {
	t.Helper()
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/api/v1/search", NewSearchHandlerWithHybridProvider(nil, provider, nil).Search)
	response := httptest.NewRecorder()
	router.ServeHTTP(response, httptest.NewRequest(http.MethodGet, target, nil))
	return response
}

func decodeSearchResponse(t *testing.T, response *httptest.ResponseRecorder) map[string]any {
	t.Helper()
	var payload map[string]any
	if err := json.Unmarshal(response.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode search response: %v", err)
	}
	return payload
}

func assertFourResultArrays(t *testing.T, payload map[string]any, lengths map[string]int) {
	t.Helper()
	results, ok := payload["results"].(map[string]any)
	if !ok {
		t.Fatalf("results = %#v, want object", payload["results"])
	}
	for _, field := range []string{"clips", "creators", "games", "tags"} {
		items, ok := results[field].([]any)
		if !ok {
			t.Fatalf("results.%s = %#v, want array", field, results[field])
		}
		if len(items) != lengths[field] {
			t.Errorf("len(results.%s) = %d, want %d", field, len(items), lengths[field])
		}
	}
}

func TestSearchHandlerWireContractNormalizesNilAndEmptyCollections(t *testing.T) {
	for _, test := range []struct {
		name    string
		results models.SearchResultsByType
	}{
		{name: "nil", results: models.SearchResultsByType{}},
		{name: "empty", results: models.EmptySearchResults()},
	} {
		t.Run(test.name, func(t *testing.T) {
			provider := &searchContractProvider{response: &models.SearchResponse{Query: "none", Results: test.results}}
			response := performSearchRequest(t, provider, "/api/v1/search?q=none")
			if response.Code != http.StatusOK {
				t.Fatalf("status = %d, want 200: %s", response.Code, response.Body.String())
			}
			assertFourResultArrays(t, decodeSearchResponse(t, response), map[string]int{})
		})
	}
}

func TestSearchHandlerWireContractPreservesPopulatedAndFilteredResults(t *testing.T) {
	provider := &searchContractProvider{response: &models.SearchResponse{
		Query: "clip",
		Results: models.SearchResultsByType{
			Clips:    []models.Clip{{}},
			Creators: []models.User{{}},
			Games:    []models.GameSearchResult{{}},
			Tags:     []models.Tag{{}},
		},
	}}
	response := performSearchRequest(t, provider, "/api/v1/search?q=clip&type=clips&game_id=game-1&language=en&tags=featured&min_votes=5&page=2&limit=10")
	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200: %s", response.Code, response.Body.String())
	}
	assertFourResultArrays(t, decodeSearchResponse(t, response), map[string]int{
		"clips": 1, "creators": 1, "games": 1, "tags": 1,
	})
	if provider.request == nil || provider.request.Type != "clips" || provider.request.GameID == nil || *provider.request.GameID != "game-1" || provider.request.Language == nil || *provider.request.Language != "en" || provider.request.MinVotes == nil || *provider.request.MinVotes != 5 || provider.request.Page != 2 || provider.request.Limit != 10 {
		t.Fatalf("filtered request was not preserved: %#v", provider.request)
	}
}

func TestSearchHandlerWireContractFailsClosedOnProviderErrorOrNilResponse(t *testing.T) {
	for _, test := range []struct {
		name     string
		provider *searchContractProvider
		status   int
	}{
		{name: "provider error", provider: &searchContractProvider{err: errors.New("unavailable")}, status: http.StatusServiceUnavailable},
		{name: "nil response", provider: &searchContractProvider{}, status: http.StatusInternalServerError},
	} {
		t.Run(test.name, func(t *testing.T) {
			response := performSearchRequest(t, test.provider, "/api/v1/search?q=clip")
			if response.Code != test.status {
				t.Fatalf("status = %d, want %d: %s", response.Code, test.status, response.Body.String())
			}
			if decodeSearchResponse(t, response)["error"] == nil {
				t.Fatal("failure response must contain an error")
			}
		})
	}
}
