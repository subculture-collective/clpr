package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"time"

	"github.com/opensearch-project/opensearch-go/v2/opensearchapi"
	"git.subcult.tv/subculture-collective/clpr/internal/models"
	"git.subcult.tv/subculture-collective/clpr/pkg/opensearch"
)

// OpenSearchService handles search operations using OpenSearch
type OpenSearchService struct {
	osClient  *opensearch.Client
	validator *SearchQueryValidator
}

// NewOpenSearchService creates a new OpenSearchService
func NewOpenSearchService(osClient *opensearch.Client) *OpenSearchService {
	return &OpenSearchService{
		osClient:  osClient,
		validator: NewSearchQueryValidator(DefaultSearchLimits()),
	}
}

// NewOpenSearchServiceWithLimits creates a new OpenSearchService with custom limits
func NewOpenSearchServiceWithLimits(osClient *opensearch.Client, limits SearchLimits) *OpenSearchService {
	return &OpenSearchService{
		osClient:  osClient,
		validator: NewSearchQueryValidator(limits),
	}
}

// Search performs a universal search using OpenSearch
func (s *OpenSearchService) Search(ctx context.Context, req *models.SearchRequest) (*models.SearchResponse, error) {
	// Calculate offset (from) and enforce search limits on offset and limit
	from := (req.Page - 1) * req.Limit
	s.validator.EnforceSearchLimits(&req.Limit, &from)
	// If the offset was changed by the validator, update the page number accordingly
	if from != (req.Page-1)*req.Limit {
		req.Page = (from / req.Limit) + 1
	}

	response := &models.SearchResponse{
		Query:   req.Query,
		Results: models.SearchResultsByType{},
		Counts:  models.SearchCounts{},
		Meta: models.SearchMeta{
			Page:  req.Page,
			Limit: req.Limit,
		},
	}

	// Determine what types to search
	searchAll := req.Type == "" || req.Type == "all"
	searchClips := searchAll || req.Type == "clips"
	searchCreators := searchAll || req.Type == "creators"
	searchGames := searchAll || req.Type == "games"
	searchTags := searchAll || req.Type == "tags"

	var totalCount int

	// Search clips (and get facets if searching clips)
	if searchClips {
		clips, count, facets, err := s.searchClipsWithFacets(ctx, req)
		if err != nil {
			return nil, fmt.Errorf("failed to search clips: %w", err)
		}
		response.Results.Clips = clips
		response.Counts.Clips = count
		totalCount += count

		// Include facets only when searching clips specifically or all
		if facets != nil {
			response.Facets = *facets
		}
	}

	// Search creators
	if searchCreators {
		creators, count, err := s.searchCreators(ctx, req)
		if err != nil {
			return nil, fmt.Errorf("failed to search creators: %w", err)
		}
		response.Results.Creators = creators
		response.Counts.Creators = count
		totalCount += count
	}

	// Search games
	if searchGames {
		games, count, err := s.searchGames(ctx, req)
		if err != nil {
			return nil, fmt.Errorf("failed to search games: %w", err)
		}
		response.Results.Games = games
		response.Counts.Games = count
		totalCount += count
	}

	// Search tags
	if searchTags {
		tags, count, err := s.searchTags(ctx, req)
		if err != nil {
			return nil, fmt.Errorf("failed to search tags: %w", err)
		}
		response.Results.Tags = tags
		response.Counts.Tags = count
		totalCount += count
	}

	// Calculate pagination
	response.Meta.TotalItems = totalCount
	if req.Limit > 0 {
		response.Meta.TotalPages = (totalCount + req.Limit - 1) / req.Limit
	}

	return response, nil
}

// searchClipsWithFacets searches for clips with facet aggregations
func (s *OpenSearchService) searchClipsWithFacets(ctx context.Context, req *models.SearchRequest) ([]models.Clip, int, *models.SearchFacets, error) {
	// Start with the base query
	baseQuery := s.buildClipQuery(req)

	// Wrap with function_score only when sorting by relevance (default)
	var finalQuery map[string]interface{}
	if req.Sort == "" || req.Sort == "relevance" {
		finalQuery = map[string]interface{}{
			"function_score": map[string]interface{}{
				"query": baseQuery,
				"functions": []map[string]interface{}{
					{
						"field_value_factor": map[string]interface{}{
							"field":    "engagement_score",
							"modifier": "log1p",
							"factor":   0.1,
							"missing":  0,
						},
					},
					{
						"field_value_factor": map[string]interface{}{
							"field":    "recency_score",
							"modifier": "none",
							"factor":   0.5,
							"missing":  0,
						},
					},
				},
				"score_mode": "sum",
				"boost_mode": "sum",
			},
		}
	} else {
		finalQuery = baseQuery
	}

	// Validate query clauses
	if err := s.validator.ValidateQueryClauses(finalQuery); err != nil {
		return nil, 0, nil, fmt.Errorf("query clause validation failed: %w", err)
	}

	// Validate query structure for security (fields, operators, dangerous patterns)
	if err := s.validator.ValidateQueryStructure(finalQuery); err != nil {
		return nil, 0, nil, fmt.Errorf("query structure validation failed: %w", err)
	}

	// Build aggregations
	aggs := s.buildFacetAggregations()

	// Validate aggregations
	if err := s.validator.ValidateAggregations(aggs); err != nil {
		return nil, 0, nil, fmt.Errorf("aggregation validation failed: %w", err)
	}

	from := (req.Page - 1) * req.Limit
	searchBody := map[string]interface{}{
		"query": finalQuery,
		"from":  from,
		"size":  req.Limit,
		"sort":  s.buildSortClause(req.Sort),
		"aggs":  aggs,
	}

	bodyJSON, err := json.Marshal(searchBody)
	if err != nil {
		return nil, 0, nil, fmt.Errorf("failed to marshal search body: %w", err)
	}

	// Apply timeout to search request
	timeout := s.validator.GetTimeout()

	reqOS := opensearchapi.SearchRequest{
		Index:   []string{ClipsIndex},
		Body:    bytes.NewReader(bodyJSON),
		Timeout: timeout,
	}

	res, err := reqOS.Do(ctx, s.osClient.GetClient())
	if err != nil {
		return nil, 0, nil, fmt.Errorf("search request failed: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		bodyBytes, _ := io.ReadAll(res.Body)
		return nil, 0, nil, fmt.Errorf("search error: %s - %s", res.Status(), string(bodyBytes))
	}

	var result map[string]interface{}
	if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
		return nil, 0, nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// Parse hits
	clips, total, err := s.parseClipsFromResult(result)
	if err != nil {
		return nil, 0, nil, err
	}

	// Parse facets
	facets := s.parseFacetsFromResult(result)

	return clips, total, facets, nil
}

// searchClips searches for clips in OpenSearch (without facets)
func (s *OpenSearchService) searchClips(ctx context.Context, req *models.SearchRequest) ([]models.Clip, int, error) {
	clips, total, _, err := s.searchClipsWithFacets(ctx, req)
	return clips, total, err
}

// searchCreators searches for users/creators in OpenSearch
func (s *OpenSearchService) searchCreators(ctx context.Context, req *models.SearchRequest) ([]models.User, int, error) {
	query := s.buildUserQuery(req)

	// Validate query clauses
	if err := s.validator.ValidateQueryClauses(query); err != nil {
		return nil, 0, fmt.Errorf("query clause validation failed: %w", err)
	}

	// Validate query structure for security
	if err := s.validator.ValidateQueryStructure(query); err != nil {
		return nil, 0, fmt.Errorf("query structure validation failed: %w", err)
	}

	from := (req.Page - 1) * req.Limit
	searchBody := map[string]interface{}{
		"query": query,
		"from":  from,
		"size":  req.Limit,
		"sort":  s.buildUserSortClause(req.Sort),
	}

	hits, total, err := s.executeSearch(ctx, UsersIndex, searchBody)
	if err != nil {
		return nil, 0, err
	}

	users := make([]models.User, 0, len(hits))
	for _, hit := range hits {
		var user models.User
		if err := json.Unmarshal(hit, &user); err != nil {
			continue
		}
		users = append(users, user)
	}

	return users, total, nil
}

// searchGames searches for games in OpenSearch
func (s *OpenSearchService) searchGames(ctx context.Context, req *models.SearchRequest) ([]models.GameSearchResult, int, error) {
	query := s.buildGameQuery(req)

	// Validate query clauses
	if err := s.validator.ValidateQueryClauses(query); err != nil {
		return nil, 0, fmt.Errorf("query clause validation failed: %w", err)
	}

	// Validate query structure for security
	if err := s.validator.ValidateQueryStructure(query); err != nil {
		return nil, 0, fmt.Errorf("query structure validation failed: %w", err)
	}

	from := (req.Page - 1) * req.Limit
	searchBody := map[string]interface{}{
		"query": query,
		"from":  from,
		"size":  req.Limit,
		"sort":  []map[string]interface{}{{"clip_count": "desc"}},
	}

	hits, total, err := s.executeSearch(ctx, GamesIndex, searchBody)
	if err != nil {
		return nil, 0, err
	}

	games := make([]models.GameSearchResult, 0, len(hits))
	for _, hit := range hits {
		var game models.GameSearchResult
		if err := json.Unmarshal(hit, &game); err != nil {
			continue
		}
		games = append(games, game)
	}

	return games, total, nil
}

// searchTags searches for tags in OpenSearch
func (s *OpenSearchService) searchTags(ctx context.Context, req *models.SearchRequest) ([]models.Tag, int, error) {
	query := s.buildTagQuery(req)

	// Validate query clauses
	if err := s.validator.ValidateQueryClauses(query); err != nil {
		return nil, 0, fmt.Errorf("query clause validation failed: %w", err)
	}

	// Validate query structure for security
	if err := s.validator.ValidateQueryStructure(query); err != nil {
		return nil, 0, fmt.Errorf("query structure validation failed: %w", err)
	}

	from := (req.Page - 1) * req.Limit
	searchBody := map[string]interface{}{
		"query": query,
		"from":  from,
		"size":  req.Limit,
		"sort":  []map[string]interface{}{{"usage_count": "desc"}},
	}

	hits, total, err := s.executeSearch(ctx, TagsIndex, searchBody)
	if err != nil {
		return nil, 0, err
	}

	tags := make([]models.Tag, 0, len(hits))
	for _, hit := range hits {
		var tag models.Tag
		if err := json.Unmarshal(hit, &tag); err != nil {
			continue
		}
		tags = append(tags, tag)
	}

	return tags, total, nil
}

// executeSearch performs the actual search request
func (s *OpenSearchService) executeSearch(ctx context.Context, indexName string, searchBody map[string]interface{}) ([]json.RawMessage, int, error) {
	body, err := json.Marshal(searchBody)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to marshal search body: %w", err)
	}

	req := opensearchapi.SearchRequest{
		Index: []string{indexName},
		Body:  bytes.NewReader(body),
	}

	res, err := req.Do(ctx, s.osClient.GetClient())
	if err != nil {
		return nil, 0, fmt.Errorf("search request failed: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		bodyBytes, _ := io.ReadAll(res.Body)
		return nil, 0, fmt.Errorf("search error: %s - %s", res.Status(), string(bodyBytes))
	}

	var result map[string]interface{}
	if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
		return nil, 0, fmt.Errorf("failed to decode response: %w", err)
	}

	hitsRaw, ok := result["hits"]
	if !ok {
		return nil, 0, fmt.Errorf("missing 'hits' in response")
	}
	hits, ok := hitsRaw.(map[string]interface{})
	if !ok {
		return nil, 0, fmt.Errorf("unexpected 'hits' structure in response")
	}
	totalRaw, ok := hits["total"]
	if !ok {
		return nil, 0, fmt.Errorf("missing 'total' in hits")
	}
	totalMap, ok := totalRaw.(map[string]interface{})
	if !ok {
		return nil, 0, fmt.Errorf("unexpected 'total' structure in hits")
	}
	valueRaw, ok := totalMap["value"]
	if !ok {
		return nil, 0, fmt.Errorf("missing 'value' in total")
	}
	valueFloat, ok := valueRaw.(float64)
	if !ok {
		return nil, 0, fmt.Errorf("unexpected 'value' type in total")
	}
	total := int(valueFloat)

	hitsListRaw, ok := hits["hits"]
	if !ok {
		return nil, 0, fmt.Errorf("missing 'hits' list in hits")
	}
	hitsList, ok := hitsListRaw.([]interface{})
	if !ok {
		return nil, 0, fmt.Errorf("unexpected 'hits' list structure in hits")
	}
	documents := make([]json.RawMessage, 0, len(hitsList))

	for _, hit := range hitsList {
		hitMap, ok := hit.(map[string]interface{})
		if !ok {
			return nil, 0, fmt.Errorf("unexpected hit structure in hits list")
		}
		source, ok := hitMap["_source"]
		if !ok {
			return nil, 0, fmt.Errorf("missing '_source' in hit")
		}
		sourceJSON, err := json.Marshal(source)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to marshal _source: %w", err)
		}
		documents = append(documents, sourceJSON)
	}

	return documents, total, nil
}

// buildClipQuery builds a query for clips with filters and enhanced relevance
func (s *OpenSearchService) buildClipQuery(req *models.SearchRequest) map[string]interface{} {
	must := []map[string]interface{}{}
	filter := []map[string]interface{}{
		{"term": map[string]interface{}{"is_removed": false}},
		{"exists": map[string]interface{}{"field": "submitted_by_user_id"}},
	}

	// Add text search if query is provided with language-specific fields
	if req.Query != "" {
		// Build fields list based on language if specified
		fields := []string{"title^3", "creator_name^2", "broadcaster_name^2", "game_name"}

		// Add language-specific field with higher boost if language is specified
		if req.Language != nil && *req.Language != "" {
			switch *req.Language {
			case "en":
				fields = append([]string{"title.en^4"}, fields...)
			case "es":
				fields = append([]string{"title.es^4"}, fields...)
			case "fr":
				fields = append([]string{"title.fr^4"}, fields...)
			case "de":
				fields = append([]string{"title.de^4"}, fields...)
			}
		}

		must = append(must, map[string]interface{}{
			"multi_match": map[string]interface{}{
				"query":     req.Query,
				"fields":    fields,
				"fuzziness": "AUTO",
				"operator":  "and",
			},
		})
	}

	// Add filters
	if req.GameID != nil && *req.GameID != "" {
		filter = append(filter, map[string]interface{}{
			"term": map[string]interface{}{"game_id": *req.GameID},
		})
	}

	if req.CreatorID != nil && *req.CreatorID != "" {
		filter = append(filter, map[string]interface{}{
			"term": map[string]interface{}{"creator_id": *req.CreatorID},
		})
	}

	if req.Language != nil && *req.Language != "" {
		filter = append(filter, map[string]interface{}{
			"term": map[string]interface{}{"language": *req.Language},
		})
	}

	if req.MinVotes != nil {
		filter = append(filter, map[string]interface{}{
			"range": map[string]interface{}{
				"vote_score": map[string]interface{}{"gte": *req.MinVotes},
			},
		})
	}

	if req.DateFrom != nil && *req.DateFrom != "" {
		filter = append(filter, map[string]interface{}{
			"range": map[string]interface{}{
				"created_at": map[string]interface{}{"gte": *req.DateFrom},
			},
		})
	}

	if req.DateTo != nil && *req.DateTo != "" {
		filter = append(filter, map[string]interface{}{
			"range": map[string]interface{}{
				"created_at": map[string]interface{}{"lte": *req.DateTo},
			},
		})
	}

	// If no query text, use match_all
	if len(must) == 0 {
		must = append(must, map[string]interface{}{"match_all": map[string]interface{}{}})
	}

	// Base bool query used by clip search
	baseQuery := map[string]interface{}{
		"bool": map[string]interface{}{
			"must":   must,
			"filter": filter,
		},
	}

	// Return base query here; caller decides whether to wrap with function_score based on sort
	return baseQuery
}

// buildUserQuery builds a query for users
func (s *OpenSearchService) buildUserQuery(req *models.SearchRequest) map[string]interface{} {
	must := []map[string]interface{}{}
	filter := []map[string]interface{}{
		{"term": map[string]interface{}{"is_banned": false}},
	}

	if req.Query != "" {
		must = append(must, map[string]interface{}{
			"multi_match": map[string]interface{}{
				"query":     req.Query,
				"fields":    []string{"username^3", "display_name^2", "bio"},
				"fuzziness": "AUTO",
			},
		})
	}

	if len(must) == 0 {
		must = append(must, map[string]interface{}{"match_all": map[string]interface{}{}})
	}

	return map[string]interface{}{
		"bool": map[string]interface{}{
			"must":   must,
			"filter": filter,
		},
	}
}

// buildGameQuery builds a query for games
func (s *OpenSearchService) buildGameQuery(req *models.SearchRequest) map[string]interface{} {
	if req.Query == "" {
		return map[string]interface{}{"match_all": map[string]interface{}{}}
	}

	return map[string]interface{}{
		"match": map[string]interface{}{
			"name": map[string]interface{}{
				"query":     req.Query,
				"fuzziness": "AUTO",
			},
		},
	}
}

// buildTagQuery builds a query for tags
func (s *OpenSearchService) buildTagQuery(req *models.SearchRequest) map[string]interface{} {
	if req.Query == "" {
		return map[string]interface{}{"match_all": map[string]interface{}{}}
	}

	return map[string]interface{}{
		"multi_match": map[string]interface{}{
			"query":     req.Query,
			"fields":    []string{"name^2", "description"},
			"fuzziness": "AUTO",
		},
	}
}

// buildSortClause builds sort criteria for clips
func (s *OpenSearchService) buildSortClause(sort string) []map[string]interface{} {
	switch sort {
	case "popular":
		return []map[string]interface{}{
			{"vote_score": "desc"},
			{"created_at": "desc"},
		}
	case "recent":
		return []map[string]interface{}{
			{"created_at": "desc"},
		}
	case "relevance":
		fallthrough
	default:
		return []map[string]interface{}{
			{"_score": "desc"},
			{"vote_score": "desc"},
		}
	}
}

// buildUserSortClause builds sort criteria for users (no vote_score field)
func (s *OpenSearchService) buildUserSortClause(sort string) []map[string]interface{} {
	switch sort {
	case "popular":
		return []map[string]interface{}{
			{"karma_points": "desc"},
			{"created_at": "desc"},
		}
	case "recent":
		return []map[string]interface{}{
			{"created_at": "desc"},
		}
	case "relevance":
		fallthrough
	default:
		return []map[string]interface{}{
			{"_score": "desc"},
			{"karma_points": "desc"},
		}
	}
}

// buildGameSortClause builds sort criteria for games (no vote_score field)
func (s *OpenSearchService) buildGameSortClause(sort string) []map[string]interface{} {
	switch sort {
	case "popular":
		return []map[string]interface{}{
			{"clip_count": "desc"},
		}
	case "recent":
		// Games don't have created_at, sort by name
		return []map[string]interface{}{
			{"name.keyword": "asc"},
		}
	case "relevance":
		fallthrough
	default:
		return []map[string]interface{}{
			{"_score": "desc"},
			{"clip_count": "desc"},
		}
	}
}

// buildTagSortClause builds sort criteria for tags (no vote_score field)
func (s *OpenSearchService) buildTagSortClause(sort string) []map[string]interface{} {
	switch sort {
	case "popular":
		return []map[string]interface{}{
			{"usage_count": "desc"},
			{"created_at": "desc"},
		}
	case "recent":
		return []map[string]interface{}{
			{"created_at": "desc"},
		}
	case "relevance":
		fallthrough
	default:
		return []map[string]interface{}{
			{"_score": "desc"},
			{"usage_count": "desc"},
		}
	}
}

// buildFacetAggregations builds aggregation queries for facets
func (s *OpenSearchService) buildFacetAggregations() map[string]interface{} {
	now := time.Now()
	lastHour := now.Add(-1 * time.Hour)
	lastDay := now.Add(-24 * time.Hour)
	lastWeek := now.Add(-7 * 24 * time.Hour)
	lastMonth := now.Add(-30 * 24 * time.Hour)

	// Get max aggregation size from validator
	limits := s.validator.GetLimits()
	maxSize := limits.MaxAggregationSize

	// Ensure we don't exceed the configured limit
	aggsSize := 20
	if aggsSize > maxSize {
		aggsSize = maxSize
	}

	return map[string]interface{}{
		"languages": map[string]interface{}{
			"terms": map[string]interface{}{
				"field": "language",
				"size":  aggsSize,
			},
		},
		"games": map[string]interface{}{
			"terms": map[string]interface{}{
				"field": "game_name.keyword",
				"size":  aggsSize,
			},
		},
		"date_ranges": map[string]interface{}{
			"range": map[string]interface{}{
				"field": "created_at",
				"ranges": []map[string]interface{}{
					{"from": lastHour.Format(time.RFC3339), "key": "last_hour"},
					{"from": lastDay.Format(time.RFC3339), "key": "last_day"},
					{"from": lastWeek.Format(time.RFC3339), "to": lastDay.Format(time.RFC3339), "key": "last_week"},
					{"from": lastMonth.Format(time.RFC3339), "to": lastWeek.Format(time.RFC3339), "key": "last_month"},
					{"to": lastMonth.Format(time.RFC3339), "key": "older"},
				},
			},
		},
	}
}

// parseClipsFromResult extracts clips from OpenSearch response
func (s *OpenSearchService) parseClipsFromResult(result map[string]interface{}) ([]models.Clip, int, error) {
	hitsRaw, ok := result["hits"]
	if !ok {
		return nil, 0, fmt.Errorf("missing 'hits' in response")
	}
	hits, ok := hitsRaw.(map[string]interface{})
	if !ok {
		return nil, 0, fmt.Errorf("unexpected 'hits' structure in response")
	}

	totalRaw, ok := hits["total"]
	if !ok {
		return nil, 0, fmt.Errorf("missing 'total' in hits")
	}
	totalMap, ok := totalRaw.(map[string]interface{})
	if !ok {
		return nil, 0, fmt.Errorf("unexpected 'total' structure in hits")
	}
	valueRaw, ok := totalMap["value"]
	if !ok {
		return nil, 0, fmt.Errorf("missing 'value' in total")
	}
	valueFloat, ok := valueRaw.(float64)
	if !ok {
		return nil, 0, fmt.Errorf("unexpected 'value' type in total")
	}
	total := int(valueFloat)

	hitsListRaw, ok := hits["hits"]
	if !ok {
		return nil, 0, fmt.Errorf("missing 'hits' list in hits")
	}
	hitsList, ok := hitsListRaw.([]interface{})
	if !ok {
		return nil, 0, fmt.Errorf("unexpected 'hits' list structure in hits")
	}

	clips := make([]models.Clip, 0, len(hitsList))
	for _, hit := range hitsList {
		hitMap, ok := hit.(map[string]interface{})
		if !ok {
			continue
		}
		source, ok := hitMap["_source"]
		if !ok {
			continue
		}
		sourceJSON, err := json.Marshal(source)
		if err != nil {
			continue
		}
		var clip models.Clip
		if err := json.Unmarshal(sourceJSON, &clip); err != nil {
			continue
		}
		clips = append(clips, clip)
	}

	return clips, total, nil
}

// parseFacetsFromResult extracts facets from OpenSearch response
func (s *OpenSearchService) parseFacetsFromResult(result map[string]interface{}) *models.SearchFacets {
	facets := &models.SearchFacets{}

	aggsRaw, ok := result["aggregations"]
	if !ok {
		return facets
	}
	aggs, ok := aggsRaw.(map[string]interface{})
	if !ok {
		return facets
	}

	// Parse language facets
	if languagesRaw, ok := aggs["languages"]; ok {
		if languagesMap, ok := languagesRaw.(map[string]interface{}); ok {
			if bucketsRaw, ok := languagesMap["buckets"]; ok {
				if bucketsList, ok := bucketsRaw.([]interface{}); ok {
					for _, bucket := range bucketsList {
						if bucketMap, ok := bucket.(map[string]interface{}); ok {
							key, _ := bucketMap["key"].(string)
							count, _ := bucketMap["doc_count"].(float64)
							facets.Languages = append(facets.Languages, models.FacetBucket{
								Key:   key,
								Label: key, // Could map to language names
								Count: int(count),
							})
						}
					}
				}
			}
		}
	}

	// Parse game facets
	if gamesRaw, ok := aggs["games"]; ok {
		if gamesMap, ok := gamesRaw.(map[string]interface{}); ok {
			if bucketsRaw, ok := gamesMap["buckets"]; ok {
				if bucketsList, ok := bucketsRaw.([]interface{}); ok {
					for _, bucket := range bucketsList {
						if bucketMap, ok := bucket.(map[string]interface{}); ok {
							key, _ := bucketMap["key"].(string)
							count, _ := bucketMap["doc_count"].(float64)
							facets.Games = append(facets.Games, models.FacetBucket{
								Key:   key,
								Label: key,
								Count: int(count),
							})
						}
					}
				}
			}
		}
	}

	// Parse date range facets
	if dateRangesRaw, ok := aggs["date_ranges"]; ok {
		if dateRangesMap, ok := dateRangesRaw.(map[string]interface{}); ok {
			if bucketsRaw, ok := dateRangesMap["buckets"]; ok {
				if bucketsList, ok := bucketsRaw.([]interface{}); ok {
					for _, bucket := range bucketsList {
						if bucketMap, ok := bucket.(map[string]interface{}); ok {
							key, _ := bucketMap["key"].(string)
							count, _ := bucketMap["doc_count"].(float64)

							switch key {
							case "last_hour":
								facets.DateRange.LastHour = int(count)
							case "last_day":
								facets.DateRange.LastDay = int(count)
							case "last_week":
								facets.DateRange.LastWeek = int(count)
							case "last_month":
								facets.DateRange.LastMonth = int(count)
							case "older":
								facets.DateRange.Older = int(count)
							}
						}
					}
				}
			}
		}
	}

	return facets
}

// GetSuggestions provides autocomplete suggestions
func (s *OpenSearchService) GetSuggestions(ctx context.Context, query string, limit int) ([]models.SearchSuggestion, error) {
	if query == "" || len(query) < 2 {
		return []models.SearchSuggestion{}, nil
	}

	suggestions := []models.SearchSuggestion{}

	// Search games
	gameQuery := map[string]interface{}{
		"query": map[string]interface{}{
			"match_phrase_prefix": map[string]interface{}{
				"name": query,
			},
		},
		"size": limit / 2,
	}

	gameHits, _, err := s.executeSearch(ctx, GamesIndex, gameQuery)
	if err == nil {
		for _, hit := range gameHits {
			var game models.GameEntity
			if err := json.Unmarshal(hit, &game); err == nil {
				suggestions = append(suggestions, models.SearchSuggestion{
					Text: game.Name,
					Type: "game",
				})
			}
		}
	}

	// Search creators
	creatorQuery := map[string]interface{}{
		"query": map[string]interface{}{
			"multi_match": map[string]interface{}{
				"query":  query,
				"fields": []string{"username", "display_name"},
				"type":   "phrase_prefix",
			},
		},
		"size": limit / 2,
	}

	creatorHits, _, err := s.executeSearch(ctx, UsersIndex, creatorQuery)
	if err == nil {
		for _, hit := range creatorHits {
			var user models.User
			if err := json.Unmarshal(hit, &user); err == nil {
				suggestions = append(suggestions, models.SearchSuggestion{
					Text: user.Username,
					Type: "creator",
				})
			}
		}
	}

	return suggestions, nil
}
