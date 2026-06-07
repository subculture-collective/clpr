package services

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
	"git.subcult.tv/subculture-collective/clpr/internal/models"
	"git.subcult.tv/subculture-collective/clpr/pkg/metrics"
)

const (
	// DefaultEmbeddingModel is the default OpenAI embedding model
	DefaultEmbeddingModel = "text-embedding-3-small"
	// EmbeddingDimensions is the number of dimensions in the embedding vector
	EmbeddingDimensions = 768
	// EmbeddingCacheTTL is how long to cache embeddings (30 days)
	EmbeddingCacheTTL = 30 * 24 * time.Hour
	// MaxRetries for API calls
	MaxRetries = 3
	// RetryDelay between retries
	RetryDelay = 2 * time.Second
)

// EmbeddingService handles generating and caching text embeddings
type EmbeddingService struct {
	apiKey      string
	apiBaseURL  string
	model       string
	redisClient *redis.Client
	httpClient  *http.Client
	rateLimiter *time.Ticker
}

// EmbeddingRequest represents OpenAI embedding API request
type EmbeddingRequest struct {
	Input          interface{} `json:"input"`
	Model          string      `json:"model"`
	EncodingFormat string      `json:"encoding_format,omitempty"`
}

// EmbeddingResponse represents OpenAI embedding API response
type EmbeddingResponse struct {
	Object string `json:"object"`
	Data   []struct {
		Object    string    `json:"object"`
		Embedding []float32 `json:"embedding"`
		Index     int       `json:"index"`
	} `json:"data"`
	Model string `json:"model"`
	Usage struct {
		PromptTokens int `json:"prompt_tokens"`
		TotalTokens  int `json:"total_tokens"`
	} `json:"usage"`
}

// EmbeddingConfig holds configuration for the embedding service
type EmbeddingConfig struct {
	APIKey            string
	APIBaseURL        string
	Model             string
	RedisClient       *redis.Client
	RequestsPerMinute int // Rate limiting
}

// NewEmbeddingService creates a new embedding service
func NewEmbeddingService(config *EmbeddingConfig) *EmbeddingService {
	model := config.Model
	if model == "" {
		model = DefaultEmbeddingModel
	}

	rpm := config.RequestsPerMinute
	if rpm <= 0 {
		rpm = 500 // Default: 500 requests per minute for tier 1
	}

	// Validate API key
	apiBaseURL := config.APIBaseURL
	if apiBaseURL == "" {
		apiBaseURL = "https://api.openai.com/v1/embeddings"
	}

	if config.APIKey == "" && apiBaseURL == "https://api.openai.com/v1/embeddings" {
		log.Println("WARNING: Embedding API key is empty - embedding service will fail at runtime")
	}

	return &EmbeddingService{
		apiKey:      config.APIKey,
		apiBaseURL:  apiBaseURL,
		model:       model,
		redisClient: config.RedisClient,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		rateLimiter: time.NewTicker(time.Minute / time.Duration(rpm)),
	}
}

// GenerateEmbedding generates an embedding for a single text
func (s *EmbeddingService) GenerateEmbedding(ctx context.Context, text string) ([]float32, error) {
	return s.generateEmbeddingWithType(ctx, text, "query")
}

// generateEmbeddingWithType is the core embedding generation logic with metrics tracking
func (s *EmbeddingService) generateEmbeddingWithType(ctx context.Context, text string, embeddingType string) ([]float32, error) {
	start := time.Now()

	// Check cache first
	cacheKey := s.getCacheKey(text)
	if cached, err := s.getFromCache(ctx, cacheKey); err == nil && cached != nil {
		recordEmbeddingCacheHit()
		return cached, nil
	}
	recordEmbeddingCacheMiss()

	// Rate limit
	select {
	case <-s.rateLimiter.C:
		// Continue
	case <-ctx.Done():
		return nil, ctx.Err()
	}

	// Generate embedding with retries
	var embedding []float32
	var lastErr error

	for attempt := 0; attempt < MaxRetries; attempt++ {
		if attempt > 0 {
			// Exponential backoff: 2^attempt * RetryDelay
			time.Sleep(RetryDelay * time.Duration(1<<uint(attempt)))
		}

		embedding, lastErr = s.callEmbeddingAPI(ctx, text)
		if lastErr == nil {
			break
		}
	}

	duration := float64(time.Since(start).Milliseconds())

	if lastErr != nil {
		recordEmbeddingGenerationError(embeddingType)
		return nil, fmt.Errorf("failed to generate embedding after %d attempts: %w", MaxRetries, lastErr)
	}

	recordEmbeddingGeneration(embeddingType, duration)

	// Cache the result
	if err := s.saveToCache(ctx, cacheKey, embedding); err != nil {
		// Log but don't fail - cache is optional
		log.Printf("Warning: failed to cache embedding: %v", err)
	}

	return embedding, nil
}

// GenerateBatchEmbeddings generates embeddings for multiple texts in a single API call
func (s *EmbeddingService) GenerateBatchEmbeddings(ctx context.Context, texts []string) ([][]float32, error) {
	if len(texts) == 0 {
		return nil, nil
	}

	// OpenAI supports up to 2048 inputs per request, but we'll use smaller batches
	// to avoid timeouts and make caching more effective
	const batchSize = 100

	results := make([][]float32, len(texts))

	for i := 0; i < len(texts); i += batchSize {
		end := i + batchSize
		if end > len(texts) {
			end = len(texts)
		}

		batch := texts[i:end]
		batchResults, err := s.generateBatch(ctx, batch)
		if err != nil {
			return nil, fmt.Errorf("failed to generate batch embeddings (batch %d-%d): %w", i, end, err)
		}

		copy(results[i:end], batchResults)
	}

	return results, nil
}

// generateBatch generates embeddings for a batch of texts
func (s *EmbeddingService) generateBatch(ctx context.Context, texts []string) ([][]float32, error) {
	// Check cache once and store results
	needsGeneration := make([]string, 0, len(texts))
	cachedResults := make(map[int][]float32) // maps index to cached embedding
	cacheKeys := make([]string, len(texts))

	for i, text := range texts {
		cacheKeys[i] = s.getCacheKey(text)

		if cached, err := s.getFromCache(ctx, cacheKeys[i]); err == nil {
			cachedResults[i] = cached
		} else {
			needsGeneration = append(needsGeneration, text)
		}
	}

	// If all are cached, return cached results
	if len(needsGeneration) == 0 {
		results := make([][]float32, len(texts))
		for i := range results {
			results[i] = cachedResults[i]
		}
		return results, nil
	}

	// Rate limit
	select {
	case <-s.rateLimiter.C:
		// Continue
	case <-ctx.Done():
		return nil, ctx.Err()
	}

	// Generate embeddings for texts that need it
	var embeddings [][]float32
	var lastErr error

	for attempt := 0; attempt < MaxRetries; attempt++ {
		if attempt > 0 {
			// Exponential backoff: 2^attempt * RetryDelay
			time.Sleep(RetryDelay * time.Duration(1<<uint(attempt)))
		}

		embeddings, lastErr = s.callBatchEmbeddingAPI(ctx, needsGeneration)
		if lastErr == nil {
			break
		}
	}

	if lastErr != nil {
		return nil, fmt.Errorf("failed to generate batch embeddings after %d attempts: %w", MaxRetries, lastErr)
	}

	// Cache the new embeddings
	for i, text := range needsGeneration {
		cacheKey := s.getCacheKey(text)
		if err := s.saveToCache(ctx, cacheKey, embeddings[i]); err != nil {
			log.Printf("Warning: failed to cache embedding: %v", err)
		}
	}

	// Combine cached and newly generated embeddings
	results := make([][]float32, len(texts))
	generatedIdx := 0

	for i := range texts {
		if cached, ok := cachedResults[i]; ok {
			results[i] = cached
		} else {
			results[i] = embeddings[generatedIdx]
			generatedIdx++
		}
	}

	return results, nil
}

// GenerateClipEmbedding generates an embedding for a clip based on its content
func (s *EmbeddingService) GenerateClipEmbedding(ctx context.Context, clip *models.Clip) ([]float32, error) {
	text := s.buildClipText(clip)
	return s.generateEmbeddingWithType(ctx, text, "clip")
}

// buildClipText constructs the text representation of a clip for embedding
func (s *EmbeddingService) buildClipText(clip *models.Clip) string {
	var parts []string

	// Title is most important
	if clip.Title != "" {
		parts = append(parts, "Title: "+clip.Title)
	}

	// Broadcaster and creator names
	if clip.BroadcasterName != "" {
		parts = append(parts, "Broadcaster: "+clip.BroadcasterName)
	}
	if clip.CreatorName != "" && clip.CreatorName != clip.BroadcasterName {
		parts = append(parts, "Clipped by: "+clip.CreatorName)
	}

	// Game name
	if clip.GameName != nil && *clip.GameName != "" {
		parts = append(parts, "Game: "+*clip.GameName)
	}

	return strings.Join(parts, ". ")
}

// callEmbeddingAPI makes a single API call to OpenAI
func (s *EmbeddingService) callEmbeddingAPI(ctx context.Context, text string) ([]float32, error) {
	reqBody := EmbeddingRequest{
		Input: text,
		Model: s.model,
	}

	embeddingResp, err := s.executeAPIRequest(ctx, reqBody)
	if err != nil {
		return nil, err
	}

	if len(embeddingResp.Data) == 0 {
		return nil, fmt.Errorf("no embeddings returned")
	}

	return embeddingResp.Data[0].Embedding, nil
}

// callBatchEmbeddingAPI makes a batch API call to OpenAI
func (s *EmbeddingService) callBatchEmbeddingAPI(ctx context.Context, texts []string) ([][]float32, error) {
	reqBody := EmbeddingRequest{
		Input: texts,
		Model: s.model,
	}

	embeddingResp, err := s.executeAPIRequest(ctx, reqBody)
	if err != nil {
		return nil, err
	}

	if len(embeddingResp.Data) != len(texts) {
		return nil, fmt.Errorf("expected %d embeddings, got %d", len(texts), len(embeddingResp.Data))
	}

	result := make([][]float32, len(embeddingResp.Data))
	for i, data := range embeddingResp.Data {
		result[i] = data.Embedding
	}

	return result, nil
}

// executeAPIRequest performs the common HTTP request logic to OpenAI
func (s *EmbeddingService) executeAPIRequest(ctx context.Context, reqBody EmbeddingRequest) (*EmbeddingResponse, error) {
	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", s.apiBaseURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	if s.apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+s.apiKey)
	}

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("API request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
	}

	var embeddingResp EmbeddingResponse
	if err := json.NewDecoder(resp.Body).Decode(&embeddingResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &embeddingResp, nil
}

// doAPICall is deprecated - kept for backward compatibility
// Use callEmbeddingAPI instead
func (s *EmbeddingService) doAPICall(ctx context.Context, reqBody EmbeddingRequest) ([]float32, error) {
	embeddingResp, err := s.executeAPIRequest(ctx, reqBody)
	if err != nil {
		return nil, err
	}

	if len(embeddingResp.Data) == 0 {
		return nil, fmt.Errorf("no embeddings returned")
	}

	return embeddingResp.Data[0].Embedding, nil
}

// getCacheKey generates a cache key for a text
func (s *EmbeddingService) getCacheKey(text string) string {
	hash := sha256.Sum256([]byte(s.model + ":" + text))
	return "embedding:" + hex.EncodeToString(hash[:])
}

// getFromCache retrieves an embedding from Redis cache
func (s *EmbeddingService) getFromCache(ctx context.Context, key string) ([]float32, error) {
	if s.redisClient == nil {
		return nil, fmt.Errorf("redis not available")
	}

	data, err := s.redisClient.Get(ctx, key).Bytes()
	if err != nil {
		return nil, err
	}

	var embedding []float32
	if err := json.Unmarshal(data, &embedding); err != nil {
		return nil, fmt.Errorf("failed to unmarshal cached embedding: %w", err)
	}

	return embedding, nil
}

// saveToCache stores an embedding in Redis cache
func (s *EmbeddingService) saveToCache(ctx context.Context, key string, embedding []float32) error {
	if s.redisClient == nil {
		return fmt.Errorf("redis not available")
	}

	data, err := json.Marshal(embedding)
	if err != nil {
		return fmt.Errorf("failed to marshal embedding: %w", err)
	}

	return s.redisClient.Set(ctx, key, data, EmbeddingCacheTTL).Err()
}

// Close closes the rate limiter
func (s *EmbeddingService) Close() {
	if s.rateLimiter != nil {
		s.rateLimiter.Stop()
	}
}

// GetModel returns the embedding model being used
func (s *EmbeddingService) GetModel() string {
	return s.model
}

// Metrics helper functions - wrappers for recording metrics

func recordEmbeddingCacheHit() {
	metrics.EmbeddingCacheHits.Inc()
}

func recordEmbeddingCacheMiss() {
	metrics.EmbeddingCacheMisses.Inc()
}

func recordEmbeddingGeneration(embeddingType string, durationMs float64) {
	metrics.EmbeddingGenerationTotal.WithLabelValues(embeddingType).Inc()
	metrics.EmbeddingGenerationDuration.WithLabelValues(embeddingType).Observe(durationMs)
}

func recordEmbeddingGenerationError(embeddingType string) {
	metrics.EmbeddingGenerationErrors.WithLabelValues(embeddingType).Inc()
}
