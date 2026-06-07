package services

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"git.subcult.tv/subculture-collective/clpr/internal/models"
)

// MockReputationRepo is a mock of ReputationRepository
type MockReputationRepo struct {
	mock.Mock
}

func (m *MockReputationRepo) CalculateTrustScore(ctx context.Context, userID uuid.UUID) (int, error) {
	args := m.Called(ctx, userID)
	return args.Int(0), args.Error(1)
}

func (m *MockReputationRepo) CalculateTrustScoreBreakdown(ctx context.Context, userID uuid.UUID) (*models.TrustScoreBreakdown, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.TrustScoreBreakdown), args.Error(1)
}

func (m *MockReputationRepo) UpdateUserTrustScore(ctx context.Context, userID uuid.UUID, newScore int, reason string, componentScores map[string]interface{}, changedBy *uuid.UUID, notes *string) error {
	args := m.Called(ctx, userID, newScore, reason, componentScores, changedBy, notes)
	return args.Error(0)
}

func (m *MockReputationRepo) GetTrustScoreHistory(ctx context.Context, userID uuid.UUID, limit int) ([]models.TrustScoreHistory, error) {
	args := m.Called(ctx, userID, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.TrustScoreHistory), args.Error(1)
}

func (m *MockReputationRepo) GetTrustScoreLeaderboard(ctx context.Context, limit, offset int) ([]models.LeaderboardEntry, error) {
	args := m.Called(ctx, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.LeaderboardEntry), args.Error(1)
}

// MockUserRepo is a mock of UserRepository
type MockUserRepo struct {
	mock.Mock
}

func (m *MockUserRepo) GetByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

// MockCacheService is a mock of CacheService
type MockCacheService struct {
	mock.Mock
}

func (m *MockCacheService) Get(ctx context.Context, key string) (string, error) {
	args := m.Called(ctx, key)
	return args.String(0), args.Error(1)
}

func (m *MockCacheService) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	args := m.Called(ctx, key, value, expiration)
	return args.Error(0)
}

func (m *MockCacheService) Delete(ctx context.Context, key string) error {
	args := m.Called(ctx, key)
	return args.Error(0)
}

func TestCalculateScore(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()

	t.Run("returns cached score when available", func(t *testing.T) {
		mockRepo := new(MockReputationRepo)
		mockUserRepo := new(MockUserRepo)
		mockCache := new(MockCacheService)

		service := NewTrustScoreService(mockRepo, mockUserRepo, mockCache)

		cacheKey := "trust_score:" + userID.String()
		cachedScore := 75
		cachedJSON, _ := json.Marshal(cachedScore)

		mockCache.On("Get", ctx, cacheKey).Return(string(cachedJSON), nil)

		score, err := service.CalculateScore(ctx, userID)

		assert.NoError(t, err)
		assert.Equal(t, 75, score)
		mockCache.AssertExpectations(t)
		mockRepo.AssertNotCalled(t, "CalculateTrustScore")
	})

	t.Run("calculates and caches score when not in cache", func(t *testing.T) {
		mockRepo := new(MockReputationRepo)
		mockUserRepo := new(MockUserRepo)
		mockCache := new(MockCacheService)

		service := NewTrustScoreService(mockRepo, mockUserRepo, mockCache)

		cacheKey := "trust_score:" + userID.String()

		mockCache.On("Get", ctx, cacheKey).Return("", assert.AnError)
		mockRepo.On("CalculateTrustScore", ctx, userID).Return(80, nil)
		mockCache.On("Set", ctx, cacheKey, 80, 1*time.Hour).Return(nil)

		score, err := service.CalculateScore(ctx, userID)

		assert.NoError(t, err)
		assert.Equal(t, 80, score)
		mockCache.AssertExpectations(t)
		mockRepo.AssertExpectations(t)
	})
}

func TestCalculateScoreWithBreakdown(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()

	mockRepo := new(MockReputationRepo)
	mockUserRepo := new(MockUserRepo)
	mockCache := new(MockCacheService)

	service := NewTrustScoreService(mockRepo, mockUserRepo, mockCache)

	expectedBreakdown := &models.TrustScoreBreakdown{
		TotalScore:       75,
		AccountAgeScore:  15,
		KarmaScore:       30,
		ReportAccuracy:   15,
		ActivityScore:    15,
		MaxScore:         100,
		AccountAgeDays:   365,
		KarmaPoints:      1000,
		CorrectReports:   10,
		IncorrectReports: 2,
		TotalComments:    50,
		TotalVotes:       200,
		DaysActive:       30,
		IsBanned:         false,
	}

	mockRepo.On("CalculateTrustScoreBreakdown", ctx, userID).Return(expectedBreakdown, nil)

	breakdown, err := service.CalculateScoreWithBreakdown(ctx, userID)

	assert.NoError(t, err)
	assert.Equal(t, expectedBreakdown, breakdown)
	mockRepo.AssertExpectations(t)
}

func TestUpdateScore(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()

	mockRepo := new(MockReputationRepo)
	mockUserRepo := new(MockUserRepo)
	mockCache := new(MockCacheService)

	service := NewTrustScoreService(mockRepo, mockUserRepo, mockCache)

	breakdown := &models.TrustScoreBreakdown{
		TotalScore:      75,
		AccountAgeScore: 15,
		KarmaScore:      30,
		ReportAccuracy:  15,
		ActivityScore:   15,
	}

	cacheKey := "trust_score:" + userID.String()

	mockRepo.On("CalculateTrustScoreBreakdown", ctx, userID).Return(breakdown, nil)
	mockRepo.On("UpdateUserTrustScore", ctx, userID, 75, "test_reason", mock.AnythingOfType("map[string]interface {}"), (*uuid.UUID)(nil), (*string)(nil)).Return(nil)
	mockCache.On("Delete", ctx, cacheKey).Return(nil)

	err := service.UpdateScore(ctx, userID, "test_reason")

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
	mockCache.AssertExpectations(t)
}

func TestManuallyAdjustScore(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()
	adminID := uuid.New()
	newScore := 90
	notes := "Manual adjustment for testing"

	t.Run("successfully adjusts score", func(t *testing.T) {
		mockRepo := new(MockReputationRepo)
		mockUserRepo := new(MockUserRepo)
		mockCache := new(MockCacheService)

		service := NewTrustScoreService(mockRepo, mockUserRepo, mockCache)

		cacheKey := "trust_score:" + userID.String()

		var nilMap map[string]interface{}
		mockRepo.On("UpdateUserTrustScore", ctx, userID, newScore, models.TrustScoreReasonManualAdjustment, nilMap, &adminID, &notes).Return(nil)
		mockCache.On("Delete", ctx, cacheKey).Return(nil)

		err := service.ManuallyAdjustScore(ctx, userID, newScore, adminID, "test", &notes)

		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
		mockCache.AssertExpectations(t)
	})

	t.Run("rejects score out of range", func(t *testing.T) {
		mockRepo := new(MockReputationRepo)
		mockUserRepo := new(MockUserRepo)
		mockCache := new(MockCacheService)

		service := NewTrustScoreService(mockRepo, mockUserRepo, mockCache)

		err := service.ManuallyAdjustScore(ctx, userID, 150, adminID, "test", &notes)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "must be between 0 and 100")

		err = service.ManuallyAdjustScore(ctx, userID, -10, adminID, "test", &notes)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "must be between 0 and 100")
	})
}

func TestGetScoreHistory(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()

	mockRepo := new(MockReputationRepo)
	mockUserRepo := new(MockUserRepo)
	mockCache := new(MockCacheService)

	service := NewTrustScoreService(mockRepo, mockUserRepo, mockCache)

	expectedHistory := []models.TrustScoreHistory{
		{
			ID:           uuid.New(),
			UserID:       userID,
			OldScore:     70,
			NewScore:     75,
			ChangeReason: models.TrustScoreReasonScheduledRecalc,
			CreatedAt:    time.Now(),
		},
	}

	mockRepo.On("GetTrustScoreHistory", ctx, userID, 50).Return(expectedHistory, nil)

	history, err := service.GetScoreHistory(ctx, userID, 50)

	assert.NoError(t, err)
	assert.Equal(t, expectedHistory, history)
	mockRepo.AssertExpectations(t)
}

func TestGetTrustScoreLeaderboard(t *testing.T) {
	ctx := context.Background()

	mockRepo := new(MockReputationRepo)
	mockUserRepo := new(MockUserRepo)
	mockCache := new(MockCacheService)

	service := NewTrustScoreService(mockRepo, mockUserRepo, mockCache)

	expectedLeaderboard := []models.LeaderboardEntry{
		{
			Rank:        1,
			UserID:      uuid.New(),
			Username:    "user1",
			DisplayName: "User One",
			Score:       95,
		},
	}

	mockRepo.On("GetTrustScoreLeaderboard", ctx, 10, 0).Return(expectedLeaderboard, nil)

	leaderboard, err := service.GetTrustScoreLeaderboard(ctx, 10, 0)

	assert.NoError(t, err)
	assert.Equal(t, expectedLeaderboard, leaderboard)
	mockRepo.AssertExpectations(t)
}

func TestBatchUpdateScores(t *testing.T) {
	ctx := context.Background()
	userIDs := []uuid.UUID{uuid.New(), uuid.New(), uuid.New()}

	mockRepo := new(MockReputationRepo)
	mockUserRepo := new(MockUserRepo)
	mockCache := new(MockCacheService)

	service := NewTrustScoreService(mockRepo, mockUserRepo, mockCache)

	breakdown := &models.TrustScoreBreakdown{
		TotalScore:      75,
		AccountAgeScore: 15,
		KarmaScore:      30,
		ReportAccuracy:  15,
		ActivityScore:   15,
	}

	for _, userID := range userIDs {
		cacheKey := "trust_score:" + userID.String()
		mockRepo.On("CalculateTrustScoreBreakdown", ctx, userID).Return(breakdown, nil)
		mockRepo.On("UpdateUserTrustScore", ctx, userID, 75, "batch_update", mock.AnythingOfType("map[string]interface {}"), (*uuid.UUID)(nil), (*string)(nil)).Return(nil)
		mockCache.On("Delete", ctx, cacheKey).Return(nil)
	}

	successCount, errorCount, err := service.BatchUpdateScores(ctx, userIDs, "batch_update")

	assert.NoError(t, err)
	assert.Equal(t, 3, successCount)
	assert.Equal(t, 0, errorCount)
	mockRepo.AssertExpectations(t)
	mockCache.AssertExpectations(t)
}
