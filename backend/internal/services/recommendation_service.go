package services

import (
	"context"
	"fmt"

	"git.subcult.tv/subculture-collective/clpr/internal/models"
	"git.subcult.tv/subculture-collective/clpr/internal/repository"
	"github.com/google/uuid"
)

// RecommendationService keeps the creator-first onboarding preferences that the
// web client collects. The personalised recommendation engine that once used
// them was removed in November 2026; the feed ranks by Twitch view velocity.
type RecommendationService struct {
	repo *repository.RecommendationRepository
}

// NewRecommendationService creates a preference service.
func NewRecommendationService(repo *repository.RecommendationRepository) *RecommendationService {
	return &RecommendationService{repo: repo}
}

// GetUserPreferences retrieves user preferences
func (s *RecommendationService) GetUserPreferences(ctx context.Context, userID uuid.UUID) (*models.UserPreference, error) {
	return s.repo.GetUserPreferences(ctx, userID)
}

// UpdateUserPreferences updates user preferences
func (s *RecommendationService) UpdateUserPreferences(ctx context.Context, pref *models.UserPreference) error {
	if err := s.repo.UpdateUserPreferences(ctx, pref); err != nil {
		return fmt.Errorf("failed to update user preferences: %w", err)
	}
	return nil
}

// CompleteOnboarding saves initial onboarding preferences for a user
func (s *RecommendationService) CompleteOnboarding(ctx context.Context, pref *models.UserPreference) error {
	if err := s.repo.CompleteOnboarding(ctx, pref); err != nil {
		return fmt.Errorf("failed to complete onboarding: %w", err)
	}
	RecordOnboardingCompleted()
	return nil
}
