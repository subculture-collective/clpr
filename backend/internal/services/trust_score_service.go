package services

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"git.subcult.tv/subculture-collective/clpr/internal/models"
)

// TrustScoreRepositoryInterface defines repository operations for trust scores
type TrustScoreRepositoryInterface interface {
	CalculateTrustScore(ctx context.Context, userID uuid.UUID) (int, error)
	CalculateTrustScoreBreakdown(ctx context.Context, userID uuid.UUID) (*models.TrustScoreBreakdown, error)
	UpdateUserTrustScore(ctx context.Context, userID uuid.UUID, newScore int, reason string, componentScores map[string]interface{}, changedBy *uuid.UUID, notes *string) error
	GetTrustScoreHistory(ctx context.Context, userID uuid.UUID, limit int) ([]models.TrustScoreHistory, error)
	GetTrustScoreLeaderboard(ctx context.Context, limit, offset int) ([]models.LeaderboardEntry, error)
}

// TrustScoreUserRepositoryInterface defines user repository operations needed
type TrustScoreUserRepositoryInterface interface {
	GetByID(ctx context.Context, id uuid.UUID) (*models.User, error)
}

// TrustScoreService handles trust score calculation and management
type TrustScoreService struct {
	reputationRepo TrustScoreRepositoryInterface
	userRepo       TrustScoreUserRepositoryInterface
	cacheService   CacheServiceInterface
}

// CacheServiceInterface defines cache operations needed by TrustScoreService
type CacheServiceInterface interface {
	Get(ctx context.Context, key string) (string, error)
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error
	Delete(ctx context.Context, key string) error
}

// NewTrustScoreService creates a new trust score service
func NewTrustScoreService(
	reputationRepo TrustScoreRepositoryInterface,
	userRepo TrustScoreUserRepositoryInterface,
	cacheService CacheServiceInterface,
) *TrustScoreService {
	return &TrustScoreService{
		reputationRepo: reputationRepo,
		userRepo:       userRepo,
		cacheService:   cacheService,
	}
}

// CalculateScore calculates the trust score for a user with caching
func (s *TrustScoreService) CalculateScore(ctx context.Context, userID uuid.UUID) (int, error) {
	// Try to get from cache first
	cacheKey := fmt.Sprintf("trust_score:%s", userID.String())
	cached, err := s.cacheService.Get(ctx, cacheKey)
	if err == nil && cached != "" {
		var score int
		if err := json.Unmarshal([]byte(cached), &score); err == nil {
			return score, nil
		}
	}

	// Calculate fresh score
	score, err := s.reputationRepo.CalculateTrustScore(ctx, userID)
	if err != nil {
		return 0, fmt.Errorf("failed to calculate trust score: %w", err)
	}

	// Cache the result for 1 hour
	_ = s.cacheService.Set(ctx, cacheKey, score, 1*time.Hour)

	return score, nil
}

// CalculateScoreWithBreakdown calculates trust score and returns detailed breakdown
func (s *TrustScoreService) CalculateScoreWithBreakdown(ctx context.Context, userID uuid.UUID) (*models.TrustScoreBreakdown, error) {
	breakdown, err := s.reputationRepo.CalculateTrustScoreBreakdown(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate trust score breakdown: %w", err)
	}

	return breakdown, nil
}

// UpdateScore recalculates and updates a user's trust score
func (s *TrustScoreService) UpdateScore(ctx context.Context, userID uuid.UUID, reason string) error {
	// Calculate new score
	breakdown, err := s.reputationRepo.CalculateTrustScoreBreakdown(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to calculate trust score: %w", err)
	}

	// Convert breakdown to JSONB for storage
	componentScores := map[string]interface{}{
		"account_age_score": breakdown.AccountAgeScore,
		"karma_score":       breakdown.KarmaScore,
		"report_accuracy":   breakdown.ReportAccuracy,
		"activity_score":    breakdown.ActivityScore,
		"account_age_days":  breakdown.AccountAgeDays,
		"karma_points":      breakdown.KarmaPoints,
		"correct_reports":   breakdown.CorrectReports,
		"incorrect_reports": breakdown.IncorrectReports,
		"total_comments":    breakdown.TotalComments,
		"total_votes":       breakdown.TotalVotes,
		"days_active":       breakdown.DaysActive,
		"is_banned":         breakdown.IsBanned,
	}

	// Update in database with history
	err = s.reputationRepo.UpdateUserTrustScore(ctx, userID, breakdown.TotalScore, reason, componentScores, nil, nil)
	if err != nil {
		return fmt.Errorf("failed to update trust score: %w", err)
	}

	// Invalidate cache
	cacheKey := fmt.Sprintf("trust_score:%s", userID.String())
	_ = s.cacheService.Delete(ctx, cacheKey)

	return nil
}

// UpdateScoreRealtime updates trust score in response to a triggering event
func (s *TrustScoreService) UpdateScoreRealtime(ctx context.Context, userID uuid.UUID, reason string) error {
	// Use a short timeout for real-time updates to avoid blocking
	ctxWithTimeout, cancel := context.WithTimeout(ctx, 100*time.Millisecond)
	defer cancel()

	err := s.UpdateScore(ctxWithTimeout, userID, reason)
	if err != nil {
		// Log error but don't fail the operation - this is graceful degradation
		// The score will be updated on next scheduled run
		return nil
	}

	return nil
}

// ManuallyAdjustScore allows an admin to manually set a user's trust score
func (s *TrustScoreService) ManuallyAdjustScore(
	ctx context.Context,
	userID uuid.UUID,
	newScore int,
	adminID uuid.UUID,
	reason string,
	notes *string,
) error {
	if newScore < 0 || newScore > 100 {
		return fmt.Errorf("trust score must be between 0 and 100")
	}

	// Update score with admin information
	err := s.reputationRepo.UpdateUserTrustScore(ctx, userID, newScore, models.TrustScoreReasonManualAdjustment, nil, &adminID, notes)
	if err != nil {
		return fmt.Errorf("failed to manually adjust trust score: %w", err)
	}

	// Invalidate cache
	cacheKey := fmt.Sprintf("trust_score:%s", userID.String())
	_ = s.cacheService.Delete(ctx, cacheKey)

	return nil
}

// GetScoreHistory retrieves trust score history for a user
func (s *TrustScoreService) GetScoreHistory(ctx context.Context, userID uuid.UUID, limit int) ([]models.TrustScoreHistory, error) {
	if limit <= 0 || limit > 100 {
		limit = 50
	}

	return s.reputationRepo.GetTrustScoreHistory(ctx, userID, limit)
}

// GetScoreHistoryWithUsers retrieves trust score history with user information
func (s *TrustScoreService) GetScoreHistoryWithUsers(ctx context.Context, userID uuid.UUID, limit int) ([]models.TrustScoreHistoryWithUser, error) {
	if limit <= 0 || limit > 100 {
		limit = 50
	}

	history, err := s.reputationRepo.GetTrustScoreHistory(ctx, userID, limit)
	if err != nil {
		return nil, err
	}

	// Fetch user details for the subject and admins
	result := make([]models.TrustScoreHistoryWithUser, len(history))
	for i, h := range history {
		result[i].TrustScoreHistory = h

		// Get user
		user, err := s.userRepo.GetByID(ctx, h.UserID)
		if err == nil {
			result[i].User = user
		}

		// Get admin if manual adjustment
		if h.ChangedBy != nil {
			admin, err := s.userRepo.GetByID(ctx, *h.ChangedBy)
			if err == nil {
				result[i].ChangedBy = admin
			}
		}
	}

	return result, nil
}

// GetTrustScoreLeaderboard retrieves top users by trust score
func (s *TrustScoreService) GetTrustScoreLeaderboard(ctx context.Context, limit, offset int) ([]models.LeaderboardEntry, error) {
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}

	return s.reputationRepo.GetTrustScoreLeaderboard(ctx, limit, offset)
}

// BatchUpdateScores updates trust scores for multiple users efficiently
func (s *TrustScoreService) BatchUpdateScores(ctx context.Context, userIDs []uuid.UUID, reason string) (int, int, error) {
	successCount := 0
	errorCount := 0

	for _, userID := range userIDs {
		err := s.UpdateScore(ctx, userID, reason)
		if err != nil {
			errorCount++
		} else {
			successCount++
		}
	}

	return successCount, errorCount, nil
}

// InvalidateCache removes trust score from cache for a user
func (s *TrustScoreService) InvalidateCache(ctx context.Context, userID uuid.UUID) error {
	cacheKey := fmt.Sprintf("trust_score:%s", userID.String())
	return s.cacheService.Delete(ctx, cacheKey)
}

// WarmCache pre-loads trust scores into cache for active users
func (s *TrustScoreService) WarmCache(ctx context.Context, userIDs []uuid.UUID) error {
	for _, userID := range userIDs {
		// Calculate and cache
		_, err := s.CalculateScore(ctx, userID)
		if err != nil {
			// Continue on error, just skip this user
			continue
		}
	}

	return nil
}
