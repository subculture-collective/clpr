package handlers

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"git.subcult.tv/subculture-collective/clpr/internal/models"
	"git.subcult.tv/subculture-collective/clpr/internal/repository"
	"git.subcult.tv/subculture-collective/clpr/internal/services"
	"git.subcult.tv/subculture-collective/clpr/pkg/metrics"
)

// SearchHandler handles search-related requests
type SearchHandler struct {
	searchRepo          *repository.SearchRepository
	openSearchService   openSearchProvider
	hybridSearchService hybridSearchProvider
	authService         *services.AuthService
	useOpenSearch       bool
	useHybridSearch     bool
}

// openSearchProvider defines the subset of methods needed from the OpenSearch service.
type openSearchProvider interface {
	Search(ctx context.Context, req *models.SearchRequest) (*models.SearchResponse, error)
	GetSuggestions(ctx context.Context, query string, limit int) ([]models.SearchSuggestion, error)
}

// hybridSearchProvider defines the interface required for hybrid search operations.
type hybridSearchProvider interface {
	Search(ctx context.Context, req *models.SearchRequest) (*models.SearchResponse, error)
	SearchWithScores(ctx context.Context, req *models.SearchRequest) (*models.SearchResponseWithScores, error)
}

// NewSearchHandler creates a new SearchHandler with PostgreSQL FTS
func NewSearchHandler(searchRepo *repository.SearchRepository, authService *services.AuthService) *SearchHandler {
	return &SearchHandler{
		searchRepo:      searchRepo,
		authService:     authService,
		useOpenSearch:   false,
		useHybridSearch: false,
	}
}

// NewSearchHandlerWithOpenSearch creates a new SearchHandler with OpenSearch
func NewSearchHandlerWithOpenSearch(searchRepo *repository.SearchRepository, openSearchService *services.OpenSearchService, authService *services.AuthService) *SearchHandler {
	return &SearchHandler{
		searchRepo:        searchRepo,
		openSearchService: openSearchService,
		authService:       authService,
		useOpenSearch:     true,
		useHybridSearch:   false,
	}
}

// NewSearchHandlerWithOpenSearchProvider allows injecting a custom OpenSearch provider (useful for testing)
func NewSearchHandlerWithOpenSearchProvider(searchRepo *repository.SearchRepository, provider openSearchProvider, authService *services.AuthService) *SearchHandler {
	return &SearchHandler{
		searchRepo:        searchRepo,
		openSearchService: provider,
		authService:       authService,
		useOpenSearch:     provider != nil,
		useHybridSearch:   false,
	}
}

// NewSearchHandlerWithHybridSearch creates a new SearchHandler with hybrid BM25 + vector search
func NewSearchHandlerWithHybridSearch(searchRepo *repository.SearchRepository, hybridSearchService *services.HybridSearchService, authService *services.AuthService) *SearchHandler {
	return &SearchHandler{
		searchRepo:          searchRepo,
		hybridSearchService: hybridSearchService,
		authService:         authService,
		useOpenSearch:       false,
		useHybridSearch:     true,
	}
}

// NewSearchHandlerWithHybridProvider allows injecting a custom hybrid search provider (useful for testing)
func NewSearchHandlerWithHybridProvider(searchRepo *repository.SearchRepository, hybridProvider hybridSearchProvider, authService *services.AuthService) *SearchHandler {
	return &SearchHandler{
		searchRepo:          searchRepo,
		hybridSearchService: hybridProvider,
		authService:         authService,
		useOpenSearch:       false,
		useHybridSearch:     hybridProvider != nil,
	}
}

// parseIntQueryParam safely parses an integer query parameter with default value and bounds
func parseIntQueryParam(c *gin.Context, key string, defaultValue, min, max int) int {
	valueStr := c.Query(key)
	if valueStr == "" {
		return defaultValue
	}

	var value int
	if _, err := fmt.Sscanf(valueStr, "%d", &value); err != nil {
		return defaultValue
	}

	if value < min || value > max {
		return defaultValue
	}

	return value
}

// Search handles universal search requests
// GET /api/v1/search
func (h *SearchHandler) Search(c *gin.Context) {
	var req models.SearchRequest

	// Bind query parameters
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid query parameters",
		})
		return
	}

	// Validate and set defaults
	if err := h.validateAndSetDefaults(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	// Perform search using hybrid search, OpenSearch, or PostgreSQL fallback
	var results *models.SearchResponse
	var err error
	var usedFallback bool
	var failoverReason string
	var fallbackStartTime time.Time

	if h.useHybridSearch && h.hybridSearchService != nil {
		// Use hybrid BM25 + vector similarity search
		// Note: Hybrid search does not have a fallback - it requires both OpenSearch and embeddings
		results, err = h.hybridSearchService.Search(c.Request.Context(), &req)
		if err != nil {
			// Hybrid search failure - return 503 (no fallback available)
			fmt.Printf("Hybrid search error: %v\n", err)
			failoverReason = getFailoverReason(err)
			// Track hybrid search unavailability separately (not a true failover since no fallback)
			metrics.SearchQueriesTotal.WithLabelValues("hybrid", "unavailable").Inc()

			c.Header("Retry-After", "60") // Suggest retry after 60 seconds
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"error": "Search service is temporarily unavailable. Please try again in about 1 minute.",
			})
			return
		}
	} else if h.useOpenSearch && h.openSearchService != nil {
		// Use OpenSearch BM25 only
		results, err = h.openSearchService.Search(c.Request.Context(), &req)
		if err != nil {
			// Fall back to PostgreSQL FTS
			fmt.Printf("OpenSearch error, falling back to PostgreSQL: %v\n", err)
			failoverReason = getFailoverReason(err)
			usedFallback = true
			fallbackStartTime = time.Now()

			// Track failover event
			metrics.SearchFallbackTotal.WithLabelValues(failoverReason).Inc()

			results, err = h.searchRepo.Search(c.Request.Context(), &req)
			if err != nil {
				// PostgreSQL fallback also failed
				fmt.Printf("PostgreSQL fallback error: %v\n", err)
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": "Failed to perform search",
				})
				return
			}

			// Track fallback latency
			fallbackDuration := time.Since(fallbackStartTime).Milliseconds()
			metrics.SearchFallbackDuration.WithLabelValues(failoverReason).Observe(float64(fallbackDuration))
		}
	} else {
		// Fall back to PostgreSQL FTS
		results, err = h.searchRepo.Search(c.Request.Context(), &req)
		if err != nil {
			fmt.Printf("Search error: %v\n", err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to perform search",
			})
			return
		}
	}

	// Add failover headers if fallback was used
	if usedFallback {
		c.Header("X-Search-Failover", "true")
		c.Header("X-Search-Failover-Reason", failoverReason)
		c.Header("X-Search-Failover-Service", "opensearch")
	}

	// Track search analytics (optional, get user ID if authenticated)
	totalResults := results.Counts.Clips + results.Counts.Creators + results.Counts.Games + results.Counts.Tags

	// Try to get user from context (if authenticated)
	if userVal, exists := c.Get("user"); exists {
		if user, ok := userVal.(*models.User); ok {
			// Track search with user ID
			_ = h.searchRepo.TrackSearch(c.Request.Context(), &user.ID, req.Query, totalResults)
		}
	} else {
		// Track anonymous search
		_ = h.searchRepo.TrackSearch(c.Request.Context(), nil, req.Query, totalResults)
	}

	c.JSON(http.StatusOK, results)
}

// getFailoverReason determines the reason for search failover based on the error
func getFailoverReason(err error) string {
	if err == nil {
		return "unknown"
	}

	errStr := err.Error()

	// Check for timeout errors
	if errors.Is(err, context.DeadlineExceeded) || strings.Contains(errStr, "timeout") || strings.Contains(errStr, "deadline exceeded") {
		return "timeout"
	}

	// Check for 5xx errors
	if strings.Contains(errStr, "503") || strings.Contains(errStr, "500") || strings.Contains(errStr, "502") || strings.Contains(errStr, "504") {
		return "error"
	}

	// Check for connection errors
	if strings.Contains(errStr, "connection refused") || strings.Contains(errStr, "connection reset") || strings.Contains(errStr, "no such host") {
		return "error"
	}

	// Default to error for any other failures
	return "error"
}

// validateAndSetDefaults validates search request parameters and sets defaults
func (h *SearchHandler) validateAndSetDefaults(req *models.SearchRequest) error {
	if req.Query == "" {
		return fmt.Errorf("Query parameter 'q' is required")
	}

	if req.Page < 1 {
		req.Page = 1
	}

	if req.Limit < 1 || req.Limit > 100 {
		req.Limit = 20
	}

	// Set default sort if not specified
	if req.Sort == "" {
		req.Sort = "relevance"
	}

	return nil
}

// GetSuggestions handles autocomplete suggestions
// GET /api/v1/search/suggestions
func (h *SearchHandler) GetSuggestions(c *gin.Context) {
	query := c.Query("q")

	if query == "" || len(query) < 2 {
		c.JSON(http.StatusOK, gin.H{
			"suggestions": []models.SearchSuggestion{},
		})
		return
	}

	limit := 10 // Default limit for suggestions

	var suggestions []models.SearchSuggestion
	var err error
	var usedFallback bool
	var failoverReason string
	var fallbackStartTime time.Time

	// Use OpenSearch or PostgreSQL fallback
	if h.useOpenSearch && h.openSearchService != nil {
		suggestions, err = h.openSearchService.GetSuggestions(c.Request.Context(), query, limit)
		if err != nil {
			// Fall back to PostgreSQL
			fmt.Printf("OpenSearch suggestions error, falling back to PostgreSQL: %v\n", err)
			failoverReason = getFailoverReason(err)
			usedFallback = true
			fallbackStartTime = time.Now()

			// Track failover event
			metrics.SearchFallbackTotal.WithLabelValues(failoverReason).Inc()

			suggestions, err = h.searchRepo.GetSuggestions(c.Request.Context(), query, limit)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": "Failed to get suggestions",
				})
				return
			}

			// Track fallback latency
			fallbackDuration := time.Since(fallbackStartTime).Milliseconds()
			metrics.SearchFallbackDuration.WithLabelValues(failoverReason).Observe(float64(fallbackDuration))
		}
	} else {
		suggestions, err = h.searchRepo.GetSuggestions(c.Request.Context(), query, limit)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to get suggestions",
			})
			return
		}
	}

	// Add failover headers if fallback was used
	if usedFallback {
		c.Header("X-Search-Failover", "true")
		c.Header("X-Search-Failover-Reason", failoverReason)
		c.Header("X-Search-Failover-Service", "opensearch")
	}

	c.JSON(http.StatusOK, gin.H{
		"query":       query,
		"suggestions": suggestions,
	})
}

// SearchWithScores handles search requests that include similarity scores
// GET /api/v1/search/scores
func (h *SearchHandler) SearchWithScores(c *gin.Context) {
	var req models.SearchRequest

	// Bind query parameters
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid query parameters",
		})
		return
	}

	// Validate and set defaults
	if err := h.validateAndSetDefaults(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	// Only hybrid search supports scores
	if !h.useHybridSearch || h.hybridSearchService == nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Hybrid search with similarity scores is not enabled",
		})
		return
	}

	// Perform search with scores
	results, err := h.hybridSearchService.SearchWithScores(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to perform search",
		})
		return
	}

	// Track search analytics
	totalResults := results.Counts.Clips + results.Counts.Creators + results.Counts.Games + results.Counts.Tags

	// Try to get user from context (if authenticated)
	if userVal, exists := c.Get("user"); exists {
		if user, ok := userVal.(*models.User); ok {
			_ = h.searchRepo.TrackSearch(c.Request.Context(), &user.ID, req.Query, totalResults)
		}
	} else {
		_ = h.searchRepo.TrackSearch(c.Request.Context(), nil, req.Query, totalResults)
	}

	c.JSON(http.StatusOK, results)
}

// GetTrendingSearches returns the most popular search queries
// GET /api/v1/search/trending
func (h *SearchHandler) GetTrendingSearches(c *gin.Context) {
	days := parseIntQueryParam(c, "days", 7, 1, 365)

	limit := parseIntQueryParam(c, "limit", 20, 1, 100)

	searches, err := h.searchRepo.GetTrendingSearches(c.Request.Context(), days, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get trending searches",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"trending_searches": searches,
		"days":              days,
		"limit":             limit,
	})
}

// GetFailedSearches returns searches that returned no results
// GET /api/v1/search/failed
func (h *SearchHandler) GetFailedSearches(c *gin.Context) {
	days := parseIntQueryParam(c, "days", 7, 1, 365)

	limit := parseIntQueryParam(c, "limit", 20, 1, 100)

	searches, err := h.searchRepo.GetFailedSearches(c.Request.Context(), days, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get failed searches",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"failed_searches": searches,
		"days":            days,
		"limit":           limit,
	})
}

// GetSearchHistory returns a user's recent search queries
// GET /api/v1/search/history
func (h *SearchHandler) GetSearchHistory(c *gin.Context) {
	// Get user from context (requires authentication)
	userVal, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Authentication required",
		})
		return
	}

	user, ok := userVal.(*models.User)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Invalid user context",
		})
		return
	}

	limit := parseIntQueryParam(c, "limit", 20, 1, 100)

	history, err := h.searchRepo.GetUserSearchHistory(c.Request.Context(), user.ID, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get search history",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"search_history": history,
		"limit":          limit,
	})
}

// GetSearchAnalytics returns overall search analytics (admin only)
// GET /api/v1/search/analytics
func (h *SearchHandler) GetSearchAnalytics(c *gin.Context) {
	// Check if user is admin (requires authentication and admin role)
	userVal, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Authentication required",
		})
		return
	}

	user, ok := userVal.(*models.User)
	if !ok || user.Role != "admin" {
		c.JSON(http.StatusForbidden, gin.H{
			"error": "Admin access required",
		})
		return
	}

	days := parseIntQueryParam(c, "days", 7, 1, 365)

	summary, err := h.searchRepo.GetSearchAnalyticsSummary(c.Request.Context(), days)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get search analytics",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"analytics": summary,
		"days":      days,
	})
}
