package services

import (
	"errors"
	"fmt"
	"strings"
	"time"
)

var (
	// ErrSearchResultSizeTooLarge is returned when search result size exceeds limit
	ErrSearchResultSizeTooLarge = errors.New("search result size exceeds maximum allowed")
	// ErrAggregationSizeTooLarge is returned when aggregation size exceeds limit
	ErrAggregationSizeTooLarge = errors.New("aggregation size exceeds maximum allowed")
	// ErrAggregationDepthTooDeep is returned when aggregation nesting is too deep
	ErrAggregationDepthTooDeep = errors.New("aggregation nesting exceeds maximum depth")
	// ErrTooManyQueryClauses is returned when there are too many query clauses
	ErrTooManyQueryClauses = errors.New("number of query clauses exceeds maximum allowed")
	// ErrDangerousQueryPattern is returned when a dangerous pattern is detected
	ErrDangerousQueryPattern = errors.New("dangerous query pattern detected")
	// ErrUnauthorizedFieldAccess is returned when accessing unauthorized field
	ErrUnauthorizedFieldAccess = errors.New("unauthorized field access attempted")
	// ErrInvalidQueryOperator is returned when using invalid query operator
	ErrInvalidQueryOperator = errors.New("invalid query operator")
)

// structuralKeys contains OpenSearch query structure keys that are not field names
var structuralKeys = map[string]bool{
	"bool": true, "must": true, "should": true, "must_not": true, "filter": true,
	"query": true, "aggs": true, "aggregations": true, "terms": true,
	"range": true, "gte": true, "lte": true, "gt": true, "lt": true,
	"from": true, "size": true, "sort": true, "match": true, "term": true,
	"multi_match": true, "match_phrase": true, "match_phrase_prefix": true,
	"match_all": true, "function_score": true, "functions": true,
	"field_value_factor": true, "modifier": true, "factor": true, "missing": true,
	"score_mode": true, "boost_mode": true, "boost": true, "fuzziness": true,
	"operator": true, "type": true, "source": true, "inline": true, "stored": true,
}

// operatorLikeKeys contains keys that look like query operators
var operatorLikeKeys = map[string]bool{
	"match": true, "term": true, "range": true, "bool": true,
	"must": true, "should": true, "filter": true, "must_not": true,
	"query": true, "aggs": true, "aggregations": true,
	"script": true, "source": true, "inline": true, "stored": true,
	"multi_match": true, "match_phrase": true, "match_phrase_prefix": true,
	"terms": true, "match_all": true, "function_score": true,
}

// SearchLimits defines limits for OpenSearch queries
type SearchLimits struct {
	MaxResultSize      int           // Maximum number of results per query
	MaxAggregationSize int           // Maximum aggregation bucket size
	MaxAggregationNest int           // Maximum nested aggregation depth
	MaxQueryClauses    int           // Maximum number of bool query clauses
	MaxSearchTime      time.Duration // Maximum search timeout
	MaxShardSize       int           // Maximum number of shards to query
	MaxOffset          int           // Maximum pagination offset
}

// DefaultSearchLimits returns sensible defaults for search limits
func DefaultSearchLimits() SearchLimits {
	return SearchLimits{
		MaxResultSize:      100,
		MaxAggregationSize: 100,
		MaxAggregationNest: 2,
		MaxQueryClauses:    20,
		MaxSearchTime:      5 * time.Second,
		MaxShardSize:       5,
		MaxOffset:          1000,
	}
}

// SearchQueryValidator validates OpenSearch queries against limits
type SearchQueryValidator struct {
	limits            SearchLimits
	allowedFields     map[string]bool
	allowedOperators  map[string]bool
	dangerousPatterns []string
}

// NewSearchQueryValidator creates a new search query validator
func NewSearchQueryValidator(limits SearchLimits) *SearchQueryValidator {
	return &SearchQueryValidator{
		limits:            limits,
		allowedFields:     getAllowedSearchFields(),
		allowedOperators:  getAllowedQueryOperators(),
		dangerousPatterns: getDangerousPatterns(),
	}
}

// GetLimits returns the configured search limits
func (v *SearchQueryValidator) GetLimits() SearchLimits {
	return v.limits
}

// ValidateSearchSize validates the search result size
func (v *SearchQueryValidator) ValidateSearchSize(size int) error {
	if size > v.limits.MaxResultSize {
		return fmt.Errorf("%w: requested %d (max: %d)", ErrSearchResultSizeTooLarge, size, v.limits.MaxResultSize)
	}
	return nil
}

// ValidateAggregations validates aggregation structure
func (v *SearchQueryValidator) ValidateAggregations(aggs map[string]interface{}) error {
	if len(aggs) == 0 {
		return nil
	}

	// Check aggregation depth
	maxDepth := v.getMaxAggregationDepth(aggs, 0)
	if maxDepth > v.limits.MaxAggregationNest {
		return fmt.Errorf("%w: depth %d (max: %d)", ErrAggregationDepthTooDeep, maxDepth, v.limits.MaxAggregationNest)
	}

	// Check aggregation sizes
	if err := v.validateAggregationSizes(aggs); err != nil {
		return err
	}

	return nil
}

// getMaxAggregationDepth recursively calculates the maximum nesting depth of aggregations
func (v *SearchQueryValidator) getMaxAggregationDepth(aggs map[string]interface{}, currentDepth int) int {
	maxDepth := currentDepth

	for _, aggValue := range aggs {
		if aggMap, ok := aggValue.(map[string]interface{}); ok {
			// Check if there are nested aggregations
			if nestedAggs, hasNested := aggMap["aggs"].(map[string]interface{}); hasNested {
				depth := v.getMaxAggregationDepth(nestedAggs, currentDepth+1)
				if depth > maxDepth {
					maxDepth = depth
				}
			}
			// Also check "aggregations" key (alternative naming)
			if nestedAggs, hasNested := aggMap["aggregations"].(map[string]interface{}); hasNested {
				depth := v.getMaxAggregationDepth(nestedAggs, currentDepth+1)
				if depth > maxDepth {
					maxDepth = depth
				}
			}
		}
	}

	return maxDepth
}

// validateAggregationSizes checks that aggregation bucket sizes are within limits
func (v *SearchQueryValidator) validateAggregationSizes(aggs map[string]interface{}) error {
	for aggName, aggValue := range aggs {
		if aggMap, ok := aggValue.(map[string]interface{}); ok {
			// Check terms aggregation size
			if terms, hasTerms := aggMap["terms"].(map[string]interface{}); hasTerms {
				if size, hasSize := terms["size"].(float64); hasSize {
					if int(size) > v.limits.MaxAggregationSize {
						return fmt.Errorf("%w for '%s': requested %d (max: %d)",
							ErrAggregationSizeTooLarge, aggName, int(size), v.limits.MaxAggregationSize)
					}
				}
				// Also check as int
				if size, hasSize := terms["size"].(int); hasSize {
					if size > v.limits.MaxAggregationSize {
						return fmt.Errorf("%w for '%s': requested %d (max: %d)",
							ErrAggregationSizeTooLarge, aggName, size, v.limits.MaxAggregationSize)
					}
				}
			}

			// Recursively check nested aggregations
			if nestedAggs, hasNested := aggMap["aggs"].(map[string]interface{}); hasNested {
				if err := v.validateAggregationSizes(nestedAggs); err != nil {
					return err
				}
			}
			if nestedAggs, hasNested := aggMap["aggregations"].(map[string]interface{}); hasNested {
				if err := v.validateAggregationSizes(nestedAggs); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

// ValidateQueryClauses validates the number of query clauses
func (v *SearchQueryValidator) ValidateQueryClauses(query map[string]interface{}) error {
	clauseCount := v.countQueryClauses(query)
	if clauseCount > v.limits.MaxQueryClauses {
		return fmt.Errorf("%w: found %d (max: %d)", ErrTooManyQueryClauses, clauseCount, v.limits.MaxQueryClauses)
	}
	return nil
}

// countQueryClauses recursively counts the number of clauses in a query
func (v *SearchQueryValidator) countQueryClauses(query map[string]interface{}) int {
	count := 0

	for key, value := range query {
		switch key {
		case "must", "should", "must_not", "filter":
			// These are arrays of clauses in bool queries
			if clauseArray, ok := value.([]interface{}); ok {
				// Count each clause in the array
				for _, clause := range clauseArray {
					count++ // Count this clause
					// Recursively count nested clauses within this clause
					if clauseMap, ok := clause.(map[string]interface{}); ok {
						count += v.countQueryClauses(clauseMap)
					}
				}
			}
		case "bool":
			// Recursively process bool queries
			if boolMap, ok := value.(map[string]interface{}); ok {
				count += v.countQueryClauses(boolMap)
			}
		}
	}

	return count
}

// EnforceSearchLimits enforces search limits on a search request
func (v *SearchQueryValidator) EnforceSearchLimits(size *int, from *int) {
	// Enforce maximum result size
	if size != nil && *size > v.limits.MaxResultSize {
		*size = v.limits.MaxResultSize
	}

	// Ensure size is at least 1 if specified
	if size != nil && *size < 1 {
		defaultSize := 10
		*size = defaultSize
	}

	// Enforce maximum and minimum offset for 'from'
	if from != nil {
		if *from > v.limits.MaxOffset {
			*from = v.limits.MaxOffset
		}
		if *from < 0 {
			*from = 0
		}
	}
}

// GetTimeout returns the maximum search timeout
func (v *SearchQueryValidator) GetTimeout() time.Duration {
	return v.limits.MaxSearchTime
}

// getAllowedSearchFields returns the whitelist of allowed search fields
func getAllowedSearchFields() map[string]bool {
	fields := []string{
		// Clip fields
		"title",
		// Language-specific title fields for better search relevance
		"title.en", // English
		"title.es", // Spanish
		"title.fr", // French
		"title.de", // German
		"creator_name",
		"creator_id",
		"broadcaster_name",
		"broadcaster_id",
		"transcript",
		"transcript.en",
		"game_name",
		"game_name.keyword",
		"game_id",
		"language",
		"tags",
		"created_at",
		"view_count",
		"vote_score",
		"duration",
		"is_removed",
		"is_featured",
		"is_nsfw",
		"engagement_score",
		"recency_score",
		// User fields
		"username",
		"display_name",
		"bio",
		"is_banned",
		"karma_points",
		// Tag fields
		"name",
		"description",
		"usage_count",
		// Sort fields
		"_score",
		"comment_count",
		"favorite_count",
	}

	fieldMap := make(map[string]bool, len(fields))
	for _, field := range fields {
		fieldMap[field] = true
	}
	return fieldMap
}

// getAllowedQueryOperators returns the whitelist of allowed query operators
func getAllowedQueryOperators() map[string]bool {
	operators := []string{
		"match",
		"multi_match",
		"match_phrase",
		"match_phrase_prefix",
		"term",
		"terms",
		"range",
		"bool",
		"must",
		"should",
		"must_not",
		"filter",
		"match_all",
		"function_score",
		"field_value_factor",
		"query", // Add query key which is commonly used
		"aggs",
		"aggregations",
	}

	operatorMap := make(map[string]bool, len(operators))
	for _, op := range operators {
		operatorMap[op] = true
	}
	return operatorMap
}

// getDangerousPatterns returns patterns that should be blocked
func getDangerousPatterns() []string {
	return []string{
		"script",
		"_source",
		"painless",
		"groovy",
		"expression",
		"inline",
		"stored",
		"file",
		"../",
		"..\\",
		"eval(",
		"exec(",
		"system(",
		"Runtime",
		"ProcessBuilder",
		"__import__",
	}
}

// ValidateQueryStructure validates the entire query structure for security
func (v *SearchQueryValidator) ValidateQueryStructure(query map[string]interface{}) error {
	// Check for dangerous patterns
	if err := v.checkDangerousPatterns(query); err != nil {
		return err
	}

	// Validate field access
	if err := v.validateFieldAccess(query); err != nil {
		return err
	}

	// Validate operators
	if err := v.validateOperators(query); err != nil {
		return err
	}

	return nil
}

// checkDangerousPatterns recursively checks for dangerous patterns in query
func (v *SearchQueryValidator) checkDangerousPatterns(obj interface{}) error {
	switch val := obj.(type) {
	case map[string]interface{}:
		for key, value := range val {
			// Check key for dangerous patterns
			for _, pattern := range v.dangerousPatterns {
				if contains(key, pattern) {
					return fmt.Errorf("%w: found '%s' in query key", ErrDangerousQueryPattern, pattern)
				}
			}

			// Recursively check value
			if err := v.checkDangerousPatterns(value); err != nil {
				return err
			}
		}
	case []interface{}:
		for _, item := range val {
			if err := v.checkDangerousPatterns(item); err != nil {
				return err
			}
		}
	case string:
		// Check string values for dangerous patterns
		for _, pattern := range v.dangerousPatterns {
			if contains(val, pattern) {
				return fmt.Errorf("%w: found '%s' in query value", ErrDangerousQueryPattern, pattern)
			}
		}
	}

	return nil
}

// validateFieldAccess validates that only allowed fields are accessed
func (v *SearchQueryValidator) validateFieldAccess(obj interface{}) error {
	switch val := obj.(type) {
	case map[string]interface{}:
		for key, value := range val {
			// Check if this is a field reference (common keys that contain field names)
			if key == "field" || key == "fields" {
				if err := v.checkFieldAllowed(value); err != nil {
					return err
				}
				continue // Skip further processing for this key
			}

			// For leaf-level query operators (match, term, etc.), check the field names
			if v.isLeafQueryOperator(key) {
				if fieldMap, ok := value.(map[string]interface{}); ok {
					for fieldName := range fieldMap {
						// Skip if this is a structural parameter, not a field name
						if !structuralKeys[fieldName] && !v.isFieldAllowed(fieldName) {
							return fmt.Errorf("%w: field '%s'", ErrUnauthorizedFieldAccess, fieldName)
						}
					}
				}
			}

			// Recursively validate nested structures
			if err := v.validateFieldAccess(value); err != nil {
				return err
			}
		}
	case []interface{}:
		for _, item := range val {
			if err := v.validateFieldAccess(item); err != nil {
				return err
			}
		}
	}

	return nil
}

// checkFieldAllowed checks if a field value is allowed
func (v *SearchQueryValidator) checkFieldAllowed(value interface{}) error {
	switch val := value.(type) {
	case string:
		if !v.isFieldAllowed(val) {
			return fmt.Errorf("%w: field '%s'", ErrUnauthorizedFieldAccess, val)
		}
	case []interface{}:
		for _, item := range val {
			if fieldStr, ok := item.(string); ok {
				if !v.isFieldAllowed(fieldStr) {
					return fmt.Errorf("%w: field '%s'", ErrUnauthorizedFieldAccess, fieldStr)
				}
			}
		}
	}
	return nil
}

// isFieldAllowed checks if a field is in the whitelist
func (v *SearchQueryValidator) isFieldAllowed(field string) bool {
	// Handle field boost notation (e.g., "title^3")
	for idx := 0; idx < len(field); idx++ {
		if field[idx] == '^' {
			field = field[:idx]
			break
		}
	}

	// Handle wildcard notation (e.g., "title.*")
	for idx := 0; idx < len(field); idx++ {
		if field[idx] == '*' {
			field = field[:idx]
			if len(field) > 0 && field[len(field)-1] == '.' {
				field = field[:len(field)-1]
			}
			break
		}
	}

	return v.allowedFields[field]
}

// validateOperators validates that only allowed operators are used
func (v *SearchQueryValidator) validateOperators(obj interface{}) error {
	switch val := obj.(type) {
	case map[string]interface{}:
		for key, value := range val {
			// Check if this key is an operator
			if v.looksLikeOperator(key) && !v.allowedOperators[key] {
				return fmt.Errorf("%w: '%s'", ErrInvalidQueryOperator, key)
			}

			// Recursively validate nested structures
			if err := v.validateOperators(value); err != nil {
				return err
			}
		}
	case []interface{}:
		for _, item := range val {
			if err := v.validateOperators(item); err != nil {
				return err
			}
		}
	}

	return nil
}

// isQueryOperator checks if a key is a known query operator
func (v *SearchQueryValidator) isQueryOperator(key string) bool {
	return v.allowedOperators[key]
}

// isLeafQueryOperator checks if a key is a leaf-level query operator that directly contains field names
func (v *SearchQueryValidator) isLeafQueryOperator(key string) bool {
	leafOperators := map[string]bool{
		"match":               true,
		"term":                true,
		"terms":               true,
		"match_phrase":        true,
		"match_phrase_prefix": true,
		"range":               true,
	}
	return leafOperators[key]
}

// looksLikeOperator checks if a key looks like an operator
// (operators are typically lowercase with underscores, not field names)
func (v *SearchQueryValidator) looksLikeOperator(key string) bool {
	return operatorLikeKeys[key]
}

// contains checks if a string contains a substring (case-insensitive)
func contains(s, substr string) bool {
	return strings.Contains(strings.ToLower(s), strings.ToLower(substr))
}
