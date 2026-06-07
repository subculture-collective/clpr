package scheduler

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"git.subcult.tv/subculture-collective/clpr/internal/models"
)

// MockEmbeddingService implements EmbeddingServiceInterface for testing
type MockEmbeddingService struct {
	GenerateClipEmbeddingFunc func(ctx context.Context, clip *models.Clip) ([]float32, error)
	CloseFunc                 func()
	CallCount                 int
}

func (m *MockEmbeddingService) GenerateClipEmbedding(ctx context.Context, clip *models.Clip) ([]float32, error) {
	m.CallCount++
	if m.GenerateClipEmbeddingFunc != nil {
		return m.GenerateClipEmbeddingFunc(ctx, clip)
	}
	// Return a mock embedding
	return make([]float32, 768), nil
}

func (m *MockEmbeddingService) Close() {
	if m.CloseFunc != nil {
		m.CloseFunc()
	}
}

func TestNewEmbeddingScheduler(t *testing.T) {
	mockService := &MockEmbeddingService{}

	scheduler := NewEmbeddingScheduler(nil, mockService, 60, "test-model")

	assert.NotNil(t, scheduler)
	assert.Equal(t, 60*time.Minute, scheduler.interval)
	assert.Equal(t, "test-model", scheduler.model)
	assert.NotNil(t, scheduler.stopChan)
}

func TestEmbeddingScheduler_Stop(t *testing.T) {
	mockService := &MockEmbeddingService{}
	scheduler := NewEmbeddingScheduler(nil, mockService, 60, "test-model")

	// Stop should not panic
	scheduler.Stop()

	// Stop should be idempotent
	scheduler.Stop()
	scheduler.Stop()
}

func TestEmbeddingScheduler_StopWithContext(t *testing.T) {
	mockService := &MockEmbeddingService{}
	scheduler := NewEmbeddingScheduler(nil, mockService, 1, "test-model")

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	// Start in goroutine
	done := make(chan struct{})
	go func() {
		scheduler.Start(ctx)
		close(done)
	}()

	// Wait for context to cancel
	<-done

	// Scheduler should have stopped cleanly
}

func TestUpdateEmbeddingCoverageMetrics_NilDB(t *testing.T) {
	mockService := &MockEmbeddingService{}
	scheduler := NewEmbeddingScheduler(nil, mockService, 60, "test-model")

	// Should not panic with nil db
	scheduler.updateEmbeddingCoverageMetrics(context.Background())
}

func TestRunEmbedding_NilDB(t *testing.T) {
	mockService := &MockEmbeddingService{}
	scheduler := NewEmbeddingScheduler(nil, mockService, 60, "test-model")

	// Should not panic and should return early with nil db
	scheduler.runEmbedding(context.Background())
}
