package services

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"git.subcult.tv/subculture-collective/clpr/config"
	"git.subcult.tv/subculture-collective/clpr/internal/models"
)

// MockCDNRepository is a mock implementation of CDNRepository
type MockCDNRepository struct {
	mock.Mock
}

func (m *MockCDNRepository) CreateMetric(ctx context.Context, metric *models.CDNMetrics) error {
	args := m.Called(ctx, metric)
	return args.Error(0)
}

func (m *MockCDNRepository) GetTotalCost(ctx context.Context, startTime time.Time, endTime time.Time) (float64, error) {
	if len(m.ExpectedCalls) == 0 {
		return 0, nil
	}

	args := m.Called(ctx, startTime, endTime)
	return args.Get(0).(float64), args.Error(1)
}

func (m *MockCDNRepository) GetCacheHitRate(ctx context.Context, provider string, startTime time.Time) (float64, error) {
	if len(m.ExpectedCalls) == 0 {
		return 0, nil
	}

	args := m.Called(ctx, provider, startTime)
	return args.Get(0).(float64), args.Error(1)
}

func (m *MockCDNRepository) GetMetricsSummary(ctx context.Context, provider string, metricType string, startTime time.Time) (float64, error) {
	if len(m.ExpectedCalls) == 0 {
		return 0, nil
	}

	args := m.Called(ctx, provider, metricType, startTime)
	return args.Get(0).(float64), args.Error(1)
}

func TestCDNService_GetCacheHeaders(t *testing.T) {
	mockCDNRepo := new(MockCDNRepository)

	cdnConfig := &config.CDNConfig{
		Enabled:  true,
		Provider: models.CDNProviderCloudflare,
		CacheTTL: 3600,
	}

	service := NewCDNService(mockCDNRepo, cdnConfig)

	t.Run("get cache headers when enabled", func(t *testing.T) {
		headers := service.GetCacheHeaders()

		assert.NotEmpty(t, headers)
		assert.Contains(t, headers, "Cache-Control")
	})

	t.Run("get default headers when disabled", func(t *testing.T) {
		disabledConfig := &config.CDNConfig{Enabled: false}
		disabledService := NewCDNService(mockCDNRepo, disabledConfig)

		headers := disabledService.GetCacheHeaders()

		assert.NotEmpty(t, headers)
		assert.Contains(t, headers, "Cache-Control")
	})
}

func TestCDNService_GetCDNURL_Disabled(t *testing.T) {
	mockCDNRepo := new(MockCDNRepository)

	cdnConfig := &config.CDNConfig{
		Enabled: false,
	}

	service := NewCDNService(mockCDNRepo, cdnConfig)

	ctx := context.Background()
	clip := &models.Clip{
		TwitchClipID: "test-clip-123",
	}

	url, err := service.GetCDNURL(ctx, clip)

	assert.NoError(t, err)
	assert.Empty(t, url)
}

func TestCloudflareProvider_GenerateURL(t *testing.T) {
	provider := NewCloudflareProvider("zone-id", "api-key", 3600)

	videoURL := "https://clips.twitch.tv/test.mp4"
	clip := &models.Clip{
		TwitchClipID: "test-clip-123",
		VideoURL:     &videoURL,
	}

	url, err := provider.GenerateURL(clip)

	assert.NoError(t, err)
	assert.NotEmpty(t, url)
	assert.Contains(t, url, "cdn.cloudflare.clpr.gg")
}

func TestCloudflareProvider_GetCacheHeaders(t *testing.T) {
	provider := NewCloudflareProvider("zone-id", "api-key", 7200)

	headers := provider.GetCacheHeaders()

	assert.NotEmpty(t, headers)
	assert.Contains(t, headers, "Cache-Control")
	assert.Contains(t, headers["Cache-Control"], "7200")
}

func TestBunnyProvider_GenerateURL(t *testing.T) {
	provider := NewBunnyProvider("api-key", "storage-zone", 3600)

	videoURL := "https://clips.twitch.tv/test.mp4"
	clip := &models.Clip{
		TwitchClipID: "test-clip-123",
		VideoURL:     &videoURL,
	}

	url, err := provider.GenerateURL(clip)

	assert.NoError(t, err)
	assert.NotEmpty(t, url)
	assert.Contains(t, url, "b-cdn.net")
}

func TestBunnyProvider_GetCacheHeaders(t *testing.T) {
	provider := NewBunnyProvider("api-key", "storage-zone", 3600)

	headers := provider.GetCacheHeaders()

	assert.NotEmpty(t, headers)
	assert.Contains(t, headers, "Cache-Control")
}

func TestAWSCloudFrontProvider_GenerateURL(t *testing.T) {
	provider := NewAWSCloudFrontProvider("access-key", "secret-key", "us-east-1", 3600)

	videoURL := "https://clips.twitch.tv/test.mp4"
	clip := &models.Clip{
		TwitchClipID: "test-clip-123",
		VideoURL:     &videoURL,
	}

	url, err := provider.GenerateURL(clip)

	assert.NoError(t, err)
	assert.NotEmpty(t, url)
	assert.Contains(t, url, "cloudfront.net")
}

func TestAWSCloudFrontProvider_GetCacheHeaders(t *testing.T) {
	provider := NewAWSCloudFrontProvider("access-key", "secret-key", "us-east-1", 3600)

	headers := provider.GetCacheHeaders()

	assert.NotEmpty(t, headers)
	assert.Contains(t, headers, "Cache-Control")
	assert.Contains(t, headers, "X-Cache")
}
