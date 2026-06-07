package services

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"git.subcult.tv/subculture-collective/clpr/internal/models"
	"git.subcult.tv/subculture-collective/clpr/internal/repository"
)

// ReputationService handles reputation-related business logic
type ReputationService struct {
	reputationRepo *repository.ReputationRepository
	userRepo       *repository.UserRepository
}

// NewReputationService creates a new reputation service
func NewReputationService(reputationRepo *repository.ReputationRepository, userRepo *repository.UserRepository) *ReputationService {
	return &ReputationService{
		reputationRepo: reputationRepo,
		userRepo:       userRepo,
	}
}

// GetUserReputation retrieves complete reputation info for a user
func (s *ReputationService) GetUserReputation(ctx context.Context, userID uuid.UUID) (*models.UserReputation, error) {
	// Get user basic info
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	// Get badges
	badges, err := s.reputationRepo.GetUserBadges(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get badges: %w", err)
	}

	// Get stats
	stats, err := s.reputationRepo.GetUserStats(ctx, userID)
	if err != nil {
		// Stats might not exist yet, that's ok
		stats = nil
	}

	// Calculate trust score
	trustScore, err := s.reputationRepo.CalculateTrustScore(ctx, userID)
	if err != nil {
		trustScore = 0
	}

	// Calculate engagement score
	engagementScore, err := s.reputationRepo.CalculateEngagementScore(ctx, userID)
	if err != nil {
		engagementScore = 0
	}

	// Get user rank
	rank := GetUserRank(user.KarmaPoints)

	return &models.UserReputation{
		UserID:          user.ID,
		Username:        user.Username,
		DisplayName:     user.DisplayName,
		AvatarURL:       user.AvatarURL,
		KarmaPoints:     user.KarmaPoints,
		Rank:            rank,
		TrustScore:      trustScore,
		EngagementScore: engagementScore,
		Badges:          badges,
		Stats:           stats,
		CreatedAt:       user.CreatedAt,
	}, nil
}

// GetUserKarmaHistory retrieves karma history for a user
func (s *ReputationService) GetUserKarmaHistory(ctx context.Context, userID uuid.UUID, limit int) ([]models.KarmaHistory, error) {
	if limit <= 0 || limit > 100 {
		limit = 50
	}

	return s.reputationRepo.GetUserKarmaHistory(ctx, userID, limit)
}

// GetKarmaBreakdown retrieves karma breakdown by source
func (s *ReputationService) GetKarmaBreakdown(ctx context.Context, userID uuid.UUID) (*models.KarmaBreakdown, error) {
	return s.reputationRepo.GetKarmaBreakdown(ctx, userID)
}

// GetUserBadges retrieves all badges for a user
func (s *ReputationService) GetUserBadges(ctx context.Context, userID uuid.UUID) ([]models.UserBadge, error) {
	return s.reputationRepo.GetUserBadges(ctx, userID)
}

// AwardBadge awards a badge to a user
func (s *ReputationService) AwardBadge(ctx context.Context, userID uuid.UUID, badgeID string, awardedBy *uuid.UUID) error {
	// Validate badge exists
	if !IsValidBadge(badgeID) {
		return fmt.Errorf("invalid badge ID: %s", badgeID)
	}

	return s.reputationRepo.AwardBadge(ctx, userID, badgeID, awardedBy)
}

// RemoveBadge removes a badge from a user
func (s *ReputationService) RemoveBadge(ctx context.Context, userID uuid.UUID, badgeID string) error {
	return s.reputationRepo.RemoveBadge(ctx, userID, badgeID)
}

// UpdateUserStats updates user statistics and recalculates scores
func (s *ReputationService) UpdateUserStats(ctx context.Context, userID uuid.UUID) error {
	// Calculate trust score
	trustScore, err := s.reputationRepo.CalculateTrustScore(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to calculate trust score: %w", err)
	}

	// Calculate engagement score
	engagementScore, err := s.reputationRepo.CalculateEngagementScore(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to calculate engagement score: %w", err)
	}

	// Get current stats or create new
	stats, err := s.reputationRepo.GetUserStats(ctx, userID)
	if err != nil {
		// Create new stats if doesn't exist
		stats = &models.UserStats{
			UserID: userID,
		}
	}

	// Update scores
	stats.TrustScore = trustScore
	stats.EngagementScore = engagementScore

	// Update user stats table
	err = s.reputationRepo.UpdateUserStats(ctx, stats)
	if err != nil {
		return fmt.Errorf("failed to update user stats: %w", err)
	}

	// Also update trust score in users table with history tracking
	breakdown, err := s.reputationRepo.CalculateTrustScoreBreakdown(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to calculate trust score breakdown: %w", err)
	}

	componentScores := map[string]interface{}{
		"account_age_score": breakdown.AccountAgeScore,
		"karma_score":       breakdown.KarmaScore,
		"report_accuracy":   breakdown.ReportAccuracy,
		"activity_score":    breakdown.ActivityScore,
	}

	err = s.reputationRepo.UpdateUserTrustScore(ctx, userID, breakdown.TotalScore, models.TrustScoreReasonScheduledRecalc, componentScores, nil, nil)
	if err != nil {
		return fmt.Errorf("failed to update trust score: %w", err)
	}

	return nil
}

// GetKarmaLeaderboard retrieves karma leaderboard
func (s *ReputationService) GetKarmaLeaderboard(ctx context.Context, limit int, offset int) ([]models.LeaderboardEntry, error) {
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}

	return s.reputationRepo.GetKarmaLeaderboard(ctx, limit, offset)
}

// GetEngagementLeaderboard retrieves engagement leaderboard
func (s *ReputationService) GetEngagementLeaderboard(ctx context.Context, limit int, offset int) ([]models.LeaderboardEntry, error) {
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}

	return s.reputationRepo.GetEngagementLeaderboard(ctx, limit, offset)
}

// IncrementUserActivity updates user activity statistics
func (s *ReputationService) IncrementUserActivity(ctx context.Context, userID uuid.UUID, activityType string, count int) error {
	return s.reputationRepo.IncrementUserActivity(ctx, userID, activityType, count)
}

// CheckAndAwardBadges checks and awards automatic badges
func (s *ReputationService) CheckAndAwardBadges(ctx context.Context, userID uuid.UUID) ([]string, error) {
	return s.reputationRepo.CheckAndAwardAutomaticBadges(ctx, userID)
}

// GetUserRank returns the rank name based on karma points
func GetUserRank(karma int) string {
	switch {
	case karma >= 10000:
		return "Legend"
	case karma >= 5000:
		return "Veteran"
	case karma >= 1000:
		return "Contributor"
	case karma >= 500:
		return "Regular"
	case karma >= 100:
		return "Member"
	default:
		return "Newcomer"
	}
}

// Badge definitions
var badgeDefinitions = map[string]models.Badge{
	"veteran": {
		ID:          "veteran",
		Name:        "Veteran",
		Description: "Member for over 1 year",
		Icon:        "🏆",
		Category:    "achievement",
		Requirement: "1 year account age",
	},
	"influencer": {
		ID:          "influencer",
		Name:        "Influencer",
		Description: "Earned 10,000+ karma",
		Icon:        "⭐",
		Category:    "achievement",
		Requirement: "10,000 karma",
	},
	"trusted_user": {
		ID:          "trusted_user",
		Name:        "Trusted User",
		Description: "Earned 1,000+ karma",
		Icon:        "✅",
		Category:    "achievement",
		Requirement: "1,000 karma",
	},
	"conversationalist": {
		ID:          "conversationalist",
		Name:        "Conversationalist",
		Description: "Posted 100+ comments",
		Icon:        "💬",
		Category:    "achievement",
		Requirement: "100 comments",
	},
	"curator": {
		ID:          "curator",
		Name:        "Curator",
		Description: "Cast 1,000+ votes",
		Icon:        "👍",
		Category:    "achievement",
		Requirement: "1,000 votes",
	},
	"submitter": {
		ID:          "submitter",
		Name:        "Submitter",
		Description: "Submitted 50+ clips",
		Icon:        "📹",
		Category:    "achievement",
		Requirement: "50 clip submissions",
	},
	"early_adopter": {
		ID:          "early_adopter",
		Name:        "Early Adopter",
		Description: "Joined during beta",
		Icon:        "🚀",
		Category:    "special",
	},
	"beta_tester": {
		ID:          "beta_tester",
		Name:        "Beta Tester",
		Description: "Participated in beta testing",
		Icon:        "🧪",
		Category:    "special",
	},
	"moderator": {
		ID:          "moderator",
		Name:        "Moderator",
		Description: "Community moderator",
		Icon:        "🛡️",
		Category:    "staff",
	},
	"admin": {
		ID:          "admin",
		Name:        "Admin",
		Description: "Site administrator",
		Icon:        "👑",
		Category:    "staff",
	},
	"developer": {
		ID:          "developer",
		Name:        "Developer",
		Description: "Platform developer",
		Icon:        "💻",
		Category:    "staff",
	},
	"supporter": {
		ID:          "supporter",
		Name:        "Supporter",
		Description: "Financial supporter",
		Icon:        "❤️",
		Category:    "supporter",
	},
}

// GetBadgeDefinition returns badge definition by ID
func GetBadgeDefinition(badgeID string) (*models.Badge, error) {
	badge, ok := badgeDefinitions[badgeID]
	if !ok {
		return nil, fmt.Errorf("badge not found: %s", badgeID)
	}
	return &badge, nil
}

// GetAllBadgeDefinitions returns all badge definitions
func GetAllBadgeDefinitions() []models.Badge {
	badges := make([]models.Badge, 0, len(badgeDefinitions))
	for _, badge := range badgeDefinitions {
		badges = append(badges, badge)
	}
	return badges
}

// IsValidBadge checks if a badge ID is valid
func IsValidBadge(badgeID string) bool {
	_, ok := badgeDefinitions[badgeID]
	return ok
}

// CanUserPerformAction checks if user has enough karma for an action
func CanUserPerformAction(karma int, action string) bool {
	switch action {
	case "create_tags":
		return karma >= 10
	case "report_content":
		return karma >= 50
	case "submit_clips":
		return karma >= 100
	case "nominate_featured":
		return karma >= 500
	default:
		return true
	}
}
