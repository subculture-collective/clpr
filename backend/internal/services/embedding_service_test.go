package services

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"git.subcult.tv/subculture-collective/clpr/internal/models"
)

func TestBuildClipText(t *testing.T) {
	service := &EmbeddingService{}

	gameName := "League of Legends"
	clip := &models.Clip{
		ID:              uuid.New(),
		Title:           "Amazing pentakill!",
		BroadcasterName: "Faker",
		CreatorName:     "fan123",
		GameName:        &gameName,
	}

	text := service.buildClipText(clip)

	assert.Contains(t, text, "Amazing pentakill!")
	assert.Contains(t, text, "Faker")
	assert.Contains(t, text, "fan123")
	assert.Contains(t, text, "League of Legends")
}

func TestBuildClipText_MinimalData(t *testing.T) {
	service := &EmbeddingService{}

	clip := &models.Clip{
		ID:              uuid.New(),
		Title:           "Great play",
		BroadcasterName: "streamer1",
	}

	text := service.buildClipText(clip)

	assert.Contains(t, text, "Great play")
	assert.Contains(t, text, "streamer1")
}

func TestBuildClipText_SameCreatorAndBroadcaster(t *testing.T) {
	service := &EmbeddingService{}

	clip := &models.Clip{
		ID:              uuid.New(),
		Title:           "Self clip",
		BroadcasterName: "streamer1",
		CreatorName:     "streamer1",
	}

	text := service.buildClipText(clip)

	assert.Contains(t, text, "Self clip")
	assert.Contains(t, text, "Broadcaster: streamer1")
	// Should not have "Clipped by" since it's the same as broadcaster
	assert.NotContains(t, text, "Clipped by")
}

func TestGetCacheKey(t *testing.T) {
	service := &EmbeddingService{
		model: "text-embedding-3-small",
	}

	key1 := service.getCacheKey("test text")
	key2 := service.getCacheKey("test text")
	key3 := service.getCacheKey("different text")

	// Same text should produce same key
	assert.Equal(t, key1, key2)

	// Different text should produce different key
	assert.NotEqual(t, key1, key3)

	// Key should have correct prefix
	assert.Contains(t, key1, "embedding:")
}

func TestGetCacheKey_DifferentModels(t *testing.T) {
	service1 := &EmbeddingService{
		model: "text-embedding-3-small",
	}
	service2 := &EmbeddingService{
		model: "text-embedding-ada-002",
	}

	key1 := service1.getCacheKey("test text")
	key2 := service2.getCacheKey("test text")

	// Same text but different models should produce different keys
	assert.NotEqual(t, key1, key2)
}

func TestNewEmbeddingService_DefaultValues(t *testing.T) {
	config := &EmbeddingConfig{
		APIKey: "test-key",
	}

	service := NewEmbeddingService(config)

	assert.NotNil(t, service)
	assert.Equal(t, DefaultEmbeddingModel, service.model)
	assert.NotNil(t, service.httpClient)
	assert.NotNil(t, service.rateLimiter)
	assert.Equal(t, 30*time.Second, service.httpClient.Timeout)
}

func TestNewEmbeddingService_CustomValues(t *testing.T) {
	config := &EmbeddingConfig{
		APIKey:            "test-key",
		Model:             "custom-model",
		RequestsPerMinute: 100,
	}

	service := NewEmbeddingService(config)

	assert.NotNil(t, service)
	assert.Equal(t, "custom-model", service.model)
}

func TestGenerateBatchEmbeddings_EmptyInput(t *testing.T) {
	service := &EmbeddingService{}

	result, err := service.GenerateBatchEmbeddings(context.Background(), []string{})

	assert.NoError(t, err)
	assert.Nil(t, result)
}

// TestRecordEmbeddingMetrics tests the metrics helper functions
// These functions record to Prometheus metrics and don't return errors
func TestRecordEmbeddingMetrics(t *testing.T) {
	// Test that metrics functions don't panic
	t.Run("CacheHit", func(t *testing.T) {
		// Should not panic
		recordEmbeddingCacheHit()
	})

	t.Run("CacheMiss", func(t *testing.T) {
		// Should not panic
		recordEmbeddingCacheMiss()
	})

	t.Run("Generation", func(t *testing.T) {
		// Should not panic
		recordEmbeddingGeneration("query", 100.0)
		recordEmbeddingGeneration("clip", 200.0)
	})

	t.Run("GenerationError", func(t *testing.T) {
		// Should not panic
		recordEmbeddingGenerationError("query")
		recordEmbeddingGenerationError("clip")
	})
}

// Note: Testing actual API calls would require mocking the HTTP client
// or using integration tests with a real API key
