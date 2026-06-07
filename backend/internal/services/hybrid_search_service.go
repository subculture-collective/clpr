package services

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pgvector/pgvector-go"
	"github.com/redis/go-redis/v9"
	"git.subcult.tv/subculture-collective/clpr/internal/models"
	"git.subcult.tv/subculture-collective/clpr/pkg/metrics"
)

// HybridSearchService orchestrates BM25 + vector similarity search
type HybridSearchService struct {
	pool              *pgxpool.Pool
	openSearchService *OpenSearchService
	embeddingService  *EmbeddingService
	redisClient       *redis.Client
}

// HybridSearchConfig holds configuration for hybrid search
type HybridSearchConfig struct {
	Pool              *pgxpool.Pool
	OpenSearchService *OpenSearchService
	EmbeddingService  *EmbeddingService
	RedisClient       *redis.Client
}

// NewHybridSearchService creates a new hybrid search service
func NewHybridSearchService(config *HybridSearchConfig) *HybridSearchService {
	return &HybridSearchService{
		pool:              config.Pool,
		openSearchService: config.OpenSearchService,
		embeddingService:  config.EmbeddingService,
		redisClient:       config.RedisClient,
	}
}

// Search performs hybrid search combining BM25 and vector similarity
func (s *HybridSearchService) Search(ctx context.Context, req *models.SearchRequest) (*models.SearchResponse, error) {
	searchStart := time.Now()
	searchType := "hybrid"

	// If semantic search is disabled or embedding service not available, fall back to BM25 only
	if s.embeddingService == nil || req.Query == "" {
		searchType = "bm25"
		result, err := s.openSearchService.Search(ctx, req)
		s.recordSearchMetrics(searchType, searchStart, result, err)
		return result, err
	}

	// Get BM25 candidates and query embedding
	candidates, queryEmbedding, err := s.getBM25CandidatesWithEmbedding(ctx, req)
	if err != nil {
		// Fall back to BM25 results on error
		metrics.SearchFallbackTotal.WithLabelValues("embedding_error").Inc()
		searchType = "bm25"
		result, err := s.openSearchService.Search(ctx, req)
		s.recordSearchMetrics(searchType, searchStart, result, err)
		return result, err
	}

	// If no clips found, return empty results
	if len(candidates.Results.Clips) == 0 {
		s.recordSearchMetrics(searchType, searchStart, candidates, nil)
		return candidates, nil
	}

	// Extract candidate IDs
	candidateIDs := make([]string, len(candidates.Results.Clips))
	for i, clip := range candidates.Results.Clips {
		candidateIDs[i] = clip.ID.String()
	}

	// Re-rank using vector similarity - note: no offset, we select from all candidates
	vectorStart := time.Now()
	rerankedClips, err := s.rerankByVectorSimilarity(ctx, candidateIDs, queryEmbedding, req.Limit, 0)
	metrics.VectorSearchDuration.Observe(float64(time.Since(vectorStart).Milliseconds()))

	if err != nil {
		log.Printf("Warning: vector re-ranking failed, falling back to BM25: %v", err)
		// Fall back to BM25 results
		metrics.SearchFallbackTotal.WithLabelValues("vector_search_error").Inc()
		searchType = "bm25"
		result, err := s.openSearchService.Search(ctx, req)
		s.recordSearchMetrics(searchType, searchStart, result, err)
		return result, err
	}

	// Step 5: Build response
	response := &models.SearchResponse{
		Query: req.Query,
		Results: models.SearchResultsByType{
			Clips: rerankedClips,
		},
		Counts: models.SearchCounts{
			Clips: candidates.Counts.Clips, // Use total count from BM25
		},
		Facets: candidates.Facets,
		Meta: models.SearchMeta{
			Page:       req.Page,
			Limit:      req.Limit,
			TotalItems: candidates.Counts.Clips,
		},
	}

	// Calculate total pages
	if req.Limit > 0 {
		response.Meta.TotalPages = (response.Meta.TotalItems + req.Limit - 1) / req.Limit
	}

	s.recordSearchMetrics(searchType, searchStart, response, nil)
	return response, nil
}

// getBM25CandidatesWithEmbedding retrieves BM25 candidates and generates query embedding
// This helper method extracts common logic used by both Search and SearchWithScores
func (s *HybridSearchService) getBM25CandidatesWithEmbedding(ctx context.Context, req *models.SearchRequest) (*models.SearchResponse, []float32, error) {
	// Calculate candidate pool size - fetch enough candidates to support pagination
	// We need to fetch candidates for all pages up to the requested page
	totalNeeded := req.Page * req.Limit
	candidateLimit := totalNeeded * 5
	if candidateLimit > 500 {
		candidateLimit = 500 // Cap at 500 to maintain performance
	}
	if candidateLimit < 100 {
		candidateLimit = 100 // Minimum of 100 for quality re-ranking
	}

	candidateReq := *req
	candidateReq.Limit = candidateLimit
	candidateReq.Page = 1 // Get from first page

	bm25Start := time.Now()
	bm25Results, err := s.openSearchService.Search(ctx, &candidateReq)
	metrics.BM25SearchDuration.Observe(float64(time.Since(bm25Start).Milliseconds()))

	if err != nil {
		return nil, nil, fmt.Errorf("BM25 search failed: %w", err)
	}

	// Generate query embedding
	queryEmbedding, err := s.embeddingService.GenerateEmbedding(ctx, req.Query)
	if err != nil {
		log.Printf("Warning: failed to generate query embedding: %v", err)
		return nil, nil, err
	}

	return bm25Results, queryEmbedding, nil
}

// rerankByVectorSimilarity re-ranks clips using pgvector similarity
func (s *HybridSearchService) rerankByVectorSimilarity(ctx context.Context, candidateIDs []string, queryEmbedding []float32, limit, offset int) ([]models.Clip, error) {
	if len(candidateIDs) == 0 {
		return []models.Clip{}, nil
	}

	// Convert to pgvector format
	queryVector := pgvector.NewVector(queryEmbedding)

	// Query with vector similarity re-ranking
	// Using cosine distance operator <=> for similarity
	query := `
		SELECT 
			id, twitch_clip_id, twitch_clip_url, embed_url, title,
			creator_name, creator_id, broadcaster_name, broadcaster_id,
			game_id, game_name, language, thumbnail_url, duration,
			view_count, created_at, imported_at, vote_score, comment_count,
			favorite_count, is_featured, is_nsfw, is_removed, removed_reason,
			embedding <=> $1 AS similarity_distance
		FROM clips
		WHERE id = ANY($2)
			AND embedding IS NOT NULL
			AND is_removed = false
		ORDER BY embedding <=> $1
		LIMIT $3 OFFSET $4
	`

	rows, err := s.pool.Query(ctx, query, queryVector, candidateIDs, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to query vector similarity: %w", err)
	}
	defer rows.Close()

	clips := []models.Clip{}
	for rows.Next() {
		var clip models.Clip
		var similarityDistance float64

		err := rows.Scan(
			&clip.ID, &clip.TwitchClipID, &clip.TwitchClipURL, &clip.EmbedURL,
			&clip.Title, &clip.CreatorName, &clip.CreatorID, &clip.BroadcasterName,
			&clip.BroadcasterID, &clip.GameID, &clip.GameName, &clip.Language,
			&clip.ThumbnailURL, &clip.Duration, &clip.ViewCount, &clip.CreatedAt,
			&clip.ImportedAt, &clip.VoteScore, &clip.CommentCount, &clip.FavoriteCount,
			&clip.IsFeatured, &clip.IsNSFW, &clip.IsRemoved, &clip.RemovedReason,
			&similarityDistance,
		)

		if err != nil {
			return nil, fmt.Errorf("failed to scan clip: %w", err)
		}

		clips = append(clips, clip)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating clips: %w", err)
	}

	return clips, nil
}

// SearchWithScores performs hybrid search and includes similarity scores in response
func (s *HybridSearchService) SearchWithScores(ctx context.Context, req *models.SearchRequest) (*models.SearchResponseWithScores, error) {
	// If semantic search is disabled or embedding service not available, fall back to BM25 only
	if s.embeddingService == nil || req.Query == "" {
		baseResponse, err := s.openSearchService.Search(ctx, req)
		if err != nil {
			return nil, err
		}

		// Convert to response with scores (no semantic scores available)
		return &models.SearchResponseWithScores{
			SearchResponse: *baseResponse,
			Scores:         []models.ClipScore{},
		}, nil
	}

	// Get BM25 candidates and query embedding
	candidates, queryEmbedding, err := s.getBM25CandidatesWithEmbedding(ctx, req)
	if err != nil {
		// Fall back to BM25 results on error
		baseResponse, err := s.openSearchService.Search(ctx, req)
		if err != nil {
			return nil, err
		}
		return &models.SearchResponseWithScores{
			SearchResponse: *baseResponse,
			Scores:         []models.ClipScore{},
		}, nil
	}

	if len(candidates.Results.Clips) == 0 {
		return &models.SearchResponseWithScores{
			SearchResponse: *candidates,
			Scores:         []models.ClipScore{},
		}, nil
	}

	// Extract candidate IDs
	candidateIDs := make([]string, len(candidates.Results.Clips))
	for i, clip := range candidates.Results.Clips {
		candidateIDs[i] = clip.ID.String()
	}

	// Re-rank with scores - calculate offset for the current page within all candidates
	offset := (req.Page - 1) * req.Limit
	rerankedClips, scores, err := s.rerankByVectorSimilarityWithScores(ctx, candidateIDs, queryEmbedding, req.Limit, offset)
	if err != nil {
		log.Printf("Warning: vector re-ranking failed, falling back to BM25: %v", err)
		baseResponse, err := s.openSearchService.Search(ctx, req)
		if err != nil {
			return nil, err
		}
		return &models.SearchResponseWithScores{
			SearchResponse: *baseResponse,
			Scores:         []models.ClipScore{},
		}, nil
	}

	// Build response with scores
	response := &models.SearchResponseWithScores{
		SearchResponse: models.SearchResponse{
			Query: req.Query,
			Results: models.SearchResultsByType{
				Clips: rerankedClips,
			},
			Counts: models.SearchCounts{
				Clips: candidates.Counts.Clips,
			},
			Facets: candidates.Facets,
			Meta: models.SearchMeta{
				Page:       req.Page,
				Limit:      req.Limit,
				TotalItems: candidates.Counts.Clips,
			},
		},
		Scores: scores,
	}

	if req.Limit > 0 {
		response.SearchResponse.Meta.TotalPages = (response.SearchResponse.Meta.TotalItems + req.Limit - 1) / req.Limit
	}

	return response, nil
}

// rerankByVectorSimilarityWithScores re-ranks clips and includes similarity scores
func (s *HybridSearchService) rerankByVectorSimilarityWithScores(ctx context.Context, candidateIDs []string, queryEmbedding []float32, limit, offset int) ([]models.Clip, []models.ClipScore, error) {
	if len(candidateIDs) == 0 {
		return []models.Clip{}, []models.ClipScore{}, nil
	}

	queryVector := pgvector.NewVector(queryEmbedding)

	query := `
		SELECT 
			id, twitch_clip_id, twitch_clip_url, embed_url, title,
			creator_name, creator_id, broadcaster_name, broadcaster_id,
			game_id, game_name, language, thumbnail_url, duration,
			view_count, created_at, imported_at, vote_score, comment_count,
			favorite_count, is_featured, is_nsfw, is_removed, removed_reason,
			embedding <=> $1 AS similarity_distance
		FROM clips
		WHERE id = ANY($2)
			AND embedding IS NOT NULL
			AND is_removed = false
		ORDER BY embedding <=> $1
		LIMIT $3 OFFSET $4
	`

	rows, err := s.pool.Query(ctx, query, queryVector, candidateIDs, limit, offset)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to query vector similarity: %w", err)
	}
	defer rows.Close()

	clips := []models.Clip{}
	scores := []models.ClipScore{}

	for rows.Next() {
		var clip models.Clip
		var similarityDistance float64

		err := rows.Scan(
			&clip.ID, &clip.TwitchClipID, &clip.TwitchClipURL, &clip.EmbedURL,
			&clip.Title, &clip.CreatorName, &clip.CreatorID, &clip.BroadcasterName,
			&clip.BroadcasterID, &clip.GameID, &clip.GameName, &clip.Language,
			&clip.ThumbnailURL, &clip.Duration, &clip.ViewCount, &clip.CreatedAt,
			&clip.ImportedAt, &clip.VoteScore, &clip.CommentCount, &clip.FavoriteCount,
			&clip.IsFeatured, &clip.IsNSFW, &clip.IsRemoved, &clip.RemovedReason,
			&similarityDistance,
		)

		if err != nil {
			return nil, nil, fmt.Errorf("failed to scan clip: %w", err)
		}

		clips = append(clips, clip)

		// Convert cosine distance to similarity score in [0,1] range.
		// The pgvector '<=>' operator returns cosine distance in [0,2]:
		//   0 = identical vectors (perfect match)
		//   1 = orthogonal vectors (no similarity)
		//   2 = opposite vectors (completely different)
		// We normalize to similarity score: 1.0 - (distance / 2.0)
		//   distance 0 (identical)   => similarity 1.0 (maximum)
		//   distance 1 (orthogonal)  => similarity 0.5
		//   distance 2 (opposite)    => similarity 0.0 (minimum)
		similarityScore := 1.0 - (similarityDistance / 2.0)

		scores = append(scores, models.ClipScore{
			ClipID:          clip.ID,
			SimilarityScore: similarityScore,
			SimilarityRank:  len(scores) + 1,
		})
	}

	if err := rows.Err(); err != nil {
		return nil, nil, fmt.Errorf("error iterating clips: %w", err)
	}

	return clips, scores, nil
}

// recordSearchMetrics records Prometheus metrics for search queries
func (s *HybridSearchService) recordSearchMetrics(searchType string, start time.Time, result *models.SearchResponse, err error) {
	duration := float64(time.Since(start).Milliseconds())
	status := "success"
	if err != nil {
		status = "error"
	}

	metrics.SearchQueriesTotal.WithLabelValues(searchType, status).Inc()
	metrics.SearchQueryDuration.WithLabelValues(searchType).Observe(duration)

	if result != nil {
		resultCount := float64(len(result.Results.Clips))
		metrics.SearchResultsCount.WithLabelValues(searchType).Observe(resultCount)

		if resultCount == 0 {
			metrics.SearchZeroResultsTotal.WithLabelValues(searchType).Inc()
		}
	}
}
