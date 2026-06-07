package services

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"git.subcult.tv/subculture-collective/clpr/config"
	"git.subcult.tv/subculture-collective/clpr/internal/models"
)

// MockMirrorRepository is a mock implementation of MirrorRepository
type MockMirrorRepository struct {
	mock.Mock
}

func (m *MockMirrorRepository) Create(ctx context.Context, mirror *models.ClipMirror) error {
	args := m.Called(ctx, mirror)
	return args.Error(0)
}

func (m *MockMirrorRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.ClipMirror, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.ClipMirror), args.Error(1)
}

func (m *MockMirrorRepository) GetByClipAndRegion(ctx context.Context, clipID uuid.UUID, region string) (*models.ClipMirror, error) {
	args := m.Called(ctx, clipID, region)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.ClipMirror), args.Error(1)
}

func (m *MockMirrorRepository) ListByClip(ctx context.Context, clipID uuid.UUID) ([]*models.ClipMirror, error) {
	args := m.Called(ctx, clipID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.ClipMirror), args.Error(1)
}

func (m *MockMirrorRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status string, failureReason *string) error {
	args := m.Called(ctx, id, status, failureReason)
	return args.Error(0)
}

func (m *MockMirrorRepository) RecordAccess(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockMirrorRepository) DeleteExpired(ctx context.Context) (int64, error) {
	args := m.Called(ctx)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockMirrorRepository) CreateMetric(ctx context.Context, metric *models.MirrorMetrics) error {
	args := m.Called(ctx, metric)
	return args.Error(0)
}

func (m *MockMirrorRepository) GetMirrorHitRate(ctx context.Context, startTime time.Time) (float64, error) {
	args := m.Called(ctx, startTime)
	return args.Get(0).(float64), args.Error(1)
}

func (m *MockMirrorRepository) GetPopularClipsForMirroring(ctx context.Context, threshold int, maxMirrors int, limit int) ([]uuid.UUID, error) {
	args := m.Called(ctx, threshold, maxMirrors, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]uuid.UUID), args.Error(1)
}

// MockClipRepository is a mock implementation of ClipRepository
type MockClipRepository struct {
	mock.Mock
}

func (m *MockClipRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Clip, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Clip), args.Error(1)
}

func TestClipMirrorService_GetMirrorURL(t *testing.T) {
	mockMirrorRepo := new(MockMirrorRepository)
	mockClipRepo := new(MockClipRepository)

	mirrorConfig := &config.MirrorConfig{
		Enabled: true,
		Regions: []string{"us-east-1", "eu-west-1"},
	}

	service := NewClipMirrorService(mockMirrorRepo, mockClipRepo, mirrorConfig)

	ctx := context.Background()
	clipID := uuid.New()
	userRegion := "us-east-1"

	t.Run("mirror found in user region", func(t *testing.T) {
		mirror := &models.ClipMirror{
			ID:        uuid.New(),
			ClipID:    clipID,
			Region:    userRegion,
			MirrorURL: "https://s3.us-east-1.clpr.cdn/test.mp4",
			Status:    models.MirrorStatusActive,
		}

		mockMirrorRepo.On("GetByClipAndRegion", ctx, clipID, userRegion).Return(mirror, nil).Once()
		mockMirrorRepo.On("RecordAccess", ctx, mirror.ID).Return(nil).Once()
		mockMirrorRepo.On("CreateMetric", ctx, mock.Anything).Return(nil).Once()

		url, found, err := service.GetMirrorURL(ctx, clipID, userRegion)

		assert.NoError(t, err)
		assert.True(t, found)
		assert.Equal(t, mirror.MirrorURL, url)
		mockMirrorRepo.AssertExpectations(t)
	})

	t.Run("service disabled", func(t *testing.T) {
		disabledConfig := &config.MirrorConfig{Enabled: false}
		disabledService := NewClipMirrorService(mockMirrorRepo, mockClipRepo, disabledConfig)

		url, found, err := disabledService.GetMirrorURL(ctx, clipID, userRegion)

		assert.NoError(t, err)
		assert.False(t, found)
		assert.Empty(t, url)
	})
}

func TestClipMirrorService_CleanupExpiredMirrors(t *testing.T) {
	mockMirrorRepo := new(MockMirrorRepository)
	mockClipRepo := new(MockClipRepository)

	mirrorConfig := &config.MirrorConfig{
		Enabled: true,
	}

	service := NewClipMirrorService(mockMirrorRepo, mockClipRepo, mirrorConfig)

	ctx := context.Background()

	t.Run("cleanup successful", func(t *testing.T) {
		expectedCount := int64(5)
		mockMirrorRepo.On("DeleteExpired", ctx).Return(expectedCount, nil).Once()

		count, err := service.CleanupExpiredMirrors(ctx)

		assert.NoError(t, err)
		assert.Equal(t, expectedCount, count)
		mockMirrorRepo.AssertExpectations(t)
	})

	t.Run("service disabled", func(t *testing.T) {
		disabledConfig := &config.MirrorConfig{Enabled: false}
		disabledService := NewClipMirrorService(mockMirrorRepo, mockClipRepo, disabledConfig)

		count, err := disabledService.CleanupExpiredMirrors(ctx)

		assert.NoError(t, err)
		assert.Equal(t, int64(0), count)
	})
}

func TestClipMirrorService_GetMirrorHitRate(t *testing.T) {
	mockMirrorRepo := new(MockMirrorRepository)
	mockClipRepo := new(MockClipRepository)

	mirrorConfig := &config.MirrorConfig{
		Enabled: true,
	}

	service := NewClipMirrorService(mockMirrorRepo, mockClipRepo, mirrorConfig)

	ctx := context.Background()

	t.Run("get hit rate", func(t *testing.T) {
		expectedRate := 75.5
		mockMirrorRepo.On("GetMirrorHitRate", ctx, mock.Anything).Return(expectedRate, nil).Once()

		rate, err := service.GetMirrorHitRate(ctx)

		assert.NoError(t, err)
		assert.Equal(t, expectedRate, rate)
		mockMirrorRepo.AssertExpectations(t)
	})
}

func TestClipMirrorService_IdentifyPopularClips(t *testing.T) {
	mockMirrorRepo := new(MockMirrorRepository)
	mockClipRepo := new(MockClipRepository)

	mirrorConfig := &config.MirrorConfig{
		Enabled:              true,
		ReplicationThreshold: 1000,
		MaxMirrorsPerClip:    3,
	}

	service := NewClipMirrorService(mockMirrorRepo, mockClipRepo, mirrorConfig)

	ctx := context.Background()

	t.Run("identify popular clips", func(t *testing.T) {
		expectedClips := []uuid.UUID{uuid.New(), uuid.New(), uuid.New()}
		mockMirrorRepo.On("GetPopularClipsForMirroring", ctx, 1000, 3, 100).Return(expectedClips, nil).Once()

		clips, err := service.IdentifyPopularClips(ctx)

		assert.NoError(t, err)
		assert.Equal(t, len(expectedClips), len(clips))
		mockMirrorRepo.AssertExpectations(t)
	})
}
