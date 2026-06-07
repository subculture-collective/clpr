package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"git.subcult.tv/subculture-collective/clpr/internal/models"
)

// ReputationRepository handles reputation-related database operations
type ReputationRepository struct {
	db *pgxpool.Pool
}

// NewReputationRepository creates a new reputation repository
func NewReputationRepository(db *pgxpool.Pool) *ReputationRepository {
	return &ReputationRepository{db: db}
}

// GetUserKarmaHistory retrieves karma history for a user
func (r *ReputationRepository) GetUserKarmaHistory(ctx context.Context, userID uuid.UUID, limit int) ([]models.KarmaHistory, error) {
	query := `
		SELECT id, user_id, amount, source, source_id, created_at
		FROM karma_history
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT $2
	`

	rows, err := r.db.Query(ctx, query, userID, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query karma history: %w", err)
	}
	defer rows.Close()

	var history []models.KarmaHistory
	for rows.Next() {
		var h models.KarmaHistory
		err := rows.Scan(&h.ID, &h.UserID, &h.Amount, &h.Source, &h.SourceID, &h.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan karma history: %w", err)
		}
		history = append(history, h)
	}

	return history, rows.Err()
}

// GetKarmaBreakdown calculates karma breakdown by source
func (r *ReputationRepository) GetKarmaBreakdown(ctx context.Context, userID uuid.UUID) (*models.KarmaBreakdown, error) {
	query := `
		SELECT
			COALESCE(SUM(CASE WHEN source = 'clip_vote' THEN amount ELSE 0 END), 0) as clip_karma,
			COALESCE(SUM(CASE WHEN source = 'comment_vote' THEN amount ELSE 0 END), 0) as comment_karma,
			COALESCE(SUM(amount), 0) as total_karma
		FROM karma_history
		WHERE user_id = $1
	`

	var breakdown models.KarmaBreakdown
	err := r.db.QueryRow(ctx, query, userID).Scan(
		&breakdown.ClipKarma,
		&breakdown.CommentKarma,
		&breakdown.TotalKarma,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get karma breakdown: %w", err)
	}

	return &breakdown, nil
}

// GetUserBadges retrieves all badges for a user
func (r *ReputationRepository) GetUserBadges(ctx context.Context, userID uuid.UUID) ([]models.UserBadge, error) {
	query := `
		SELECT id, user_id, badge_id, awarded_at, awarded_by
		FROM user_badges
		WHERE user_id = $1
		ORDER BY awarded_at DESC
	`

	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to query user badges: %w", err)
	}
	defer rows.Close()

	var badges []models.UserBadge
	for rows.Next() {
		var badge models.UserBadge
		err := rows.Scan(&badge.ID, &badge.UserID, &badge.BadgeID, &badge.AwardedAt, &badge.AwardedBy)
		if err != nil {
			return nil, fmt.Errorf("failed to scan badge: %w", err)
		}
		badges = append(badges, badge)
	}

	return badges, rows.Err()
}

// AwardBadge awards a badge to a user
func (r *ReputationRepository) AwardBadge(ctx context.Context, userID uuid.UUID, badgeID string, awardedBy *uuid.UUID) error {
	query := `
		INSERT INTO user_badges (user_id, badge_id, awarded_by)
		VALUES ($1, $2, $3)
		ON CONFLICT (user_id, badge_id) DO NOTHING
	`

	_, err := r.db.Exec(ctx, query, userID, badgeID, awardedBy)
	if err != nil {
		return fmt.Errorf("failed to award badge: %w", err)
	}

	return nil
}

// RemoveBadge removes a badge from a user
func (r *ReputationRepository) RemoveBadge(ctx context.Context, userID uuid.UUID, badgeID string) error {
	query := `DELETE FROM user_badges WHERE user_id = $1 AND badge_id = $2`

	_, err := r.db.Exec(ctx, query, userID, badgeID)
	if err != nil {
		return fmt.Errorf("failed to remove badge: %w", err)
	}

	return nil
}

// GetUserStats retrieves user statistics
func (r *ReputationRepository) GetUserStats(ctx context.Context, userID uuid.UUID) (*models.UserStats, error) {
	query := `
		SELECT user_id, trust_score, engagement_score, total_comments,
		       total_votes_cast, total_clips_submitted, correct_reports,
		       incorrect_reports, days_active, last_active_date, updated_at
		FROM user_stats
		WHERE user_id = $1
	`

	var stats models.UserStats
	err := r.db.QueryRow(ctx, query, userID).Scan(
		&stats.UserID,
		&stats.TrustScore,
		&stats.EngagementScore,
		&stats.TotalComments,
		&stats.TotalVotesCast,
		&stats.TotalClipsSubmit,
		&stats.CorrectReports,
		&stats.IncorrectReports,
		&stats.DaysActive,
		&stats.LastActiveDate,
		&stats.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get user stats: %w", err)
	}

	return &stats, nil
}

// UpdateUserStats updates user statistics
func (r *ReputationRepository) UpdateUserStats(ctx context.Context, stats *models.UserStats) error {
	query := `
		INSERT INTO user_stats (
			user_id, trust_score, engagement_score, total_comments,
			total_votes_cast, total_clips_submitted, correct_reports,
			incorrect_reports, days_active, last_active_date, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, NOW())
		ON CONFLICT (user_id) DO UPDATE SET
			trust_score = EXCLUDED.trust_score,
			engagement_score = EXCLUDED.engagement_score,
			total_comments = EXCLUDED.total_comments,
			total_votes_cast = EXCLUDED.total_votes_cast,
			total_clips_submitted = EXCLUDED.total_clips_submitted,
			correct_reports = EXCLUDED.correct_reports,
			incorrect_reports = EXCLUDED.incorrect_reports,
			days_active = EXCLUDED.days_active,
			last_active_date = EXCLUDED.last_active_date,
			updated_at = NOW()
	`

	_, err := r.db.Exec(ctx, query,
		stats.UserID,
		stats.TrustScore,
		stats.EngagementScore,
		stats.TotalComments,
		stats.TotalVotesCast,
		stats.TotalClipsSubmit,
		stats.CorrectReports,
		stats.IncorrectReports,
		stats.DaysActive,
		stats.LastActiveDate,
	)
	if err != nil {
		return fmt.Errorf("failed to update user stats: %w", err)
	}

	return nil
}

// CalculateTrustScore calculates trust score for a user using database function
func (r *ReputationRepository) CalculateTrustScore(ctx context.Context, userID uuid.UUID) (int, error) {
	query := `SELECT calculate_trust_score($1)`

	var trustScore int
	err := r.db.QueryRow(ctx, query, userID).Scan(&trustScore)
	if err != nil {
		return 0, fmt.Errorf("failed to calculate trust score: %w", err)
	}

	return trustScore, nil
}

// CalculateEngagementScore calculates engagement score for a user
func (r *ReputationRepository) CalculateEngagementScore(ctx context.Context, userID uuid.UUID) (int, error) {
	query := `SELECT calculate_engagement_score($1)`

	var engagementScore int
	err := r.db.QueryRow(ctx, query, userID).Scan(&engagementScore)
	if err != nil {
		return 0, fmt.Errorf("failed to calculate engagement score: %w", err)
	}

	return engagementScore, nil
}

// GetKarmaLeaderboard returns top users by karma score
func (r *ReputationRepository) GetKarmaLeaderboard(ctx context.Context, limit, offset int) ([]models.LeaderboardEntry, error) {
	query := `
		SELECT id, username, display_name, avatar_url, karma_points, rank
		FROM karma_leaderboard
		LIMIT $1 OFFSET $2
	`

	rows, err := r.db.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to query karma leaderboard: %w", err)
	}
	defer rows.Close()

	var entries []models.LeaderboardEntry
	rank := offset + 1
	for rows.Next() {
		var entry models.LeaderboardEntry
		err := rows.Scan(
			&entry.UserID,
			&entry.Username,
			&entry.DisplayName,
			&entry.AvatarURL,
			&entry.Score,
			&entry.UserRank,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan leaderboard entry: %w", err)
		}
		entry.Rank = rank
		rank++
		entries = append(entries, entry)
	}

	return entries, rows.Err()
}

// GetEngagementLeaderboard returns top users by engagement score
func (r *ReputationRepository) GetEngagementLeaderboard(ctx context.Context, limit, offset int) ([]models.LeaderboardEntry, error) {
	query := `
		SELECT id, username, display_name, avatar_url, engagement_score,
		       total_comments, total_votes_cast, total_clips_submitted
		FROM engagement_leaderboard
		LIMIT $1 OFFSET $2
	`

	rows, err := r.db.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to query engagement leaderboard: %w", err)
	}
	defer rows.Close()

	var entries []models.LeaderboardEntry
	rank := offset + 1
	for rows.Next() {
		var entry models.LeaderboardEntry
		err := rows.Scan(
			&entry.UserID,
			&entry.Username,
			&entry.DisplayName,
			&entry.AvatarURL,
			&entry.Score,
			&entry.TotalComments,
			&entry.TotalVotesCast,
			&entry.TotalClipsSubmit,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan leaderboard entry: %w", err)
		}
		entry.Rank = rank
		rank++
		entries = append(entries, entry)
	}

	return entries, rows.Err()
}

// IncrementUserActivity updates user activity statistics
func (r *ReputationRepository) IncrementUserActivity(ctx context.Context, userID uuid.UUID, activityType string, count int) error {
	// Ensure user_stats exists
	_, err := r.db.Exec(ctx, `
		INSERT INTO user_stats (user_id)
		VALUES ($1)
		ON CONFLICT (user_id) DO NOTHING
	`, userID)
	if err != nil {
		return fmt.Errorf("failed to ensure user stats: %w", err)
	}

	// Update activity based on type
	var query string
	switch activityType {
	case "comment":
		query = `
			UPDATE user_stats
			SET total_comments = total_comments + $2,
			    last_active_date = CURRENT_DATE,
				updated_at = NOW()
			WHERE user_id = $1
		`
	case "vote":
		query = `
			UPDATE user_stats
			SET total_votes_cast = total_votes_cast + $2,
			    last_active_date = CURRENT_DATE,
				updated_at = NOW()
			WHERE user_id = $1
		`
	case "submission":
		query = `
			UPDATE user_stats
			SET total_clips_submitted = total_clips_submitted + $2,
			    last_active_date = CURRENT_DATE,
				updated_at = NOW()
			WHERE user_id = $1
		`
	default:
		return fmt.Errorf("unknown activity type: %s", activityType)
	}

	_, err = r.db.Exec(ctx, query, userID, count)
	if err != nil {
		return fmt.Errorf("failed to increment user activity: %w", err)
	}

	// Update days active if last active date changed
	_, err = r.db.Exec(ctx, `
		UPDATE user_stats
		SET days_active = (
			SELECT COUNT(DISTINCT DATE(created_at))
			FROM karma_history
			WHERE user_id = $1
		)
		WHERE user_id = $1
	`, userID)
	if err != nil {
		return fmt.Errorf("failed to update days active: %w", err)
	}

	return nil
}

// CheckAndAwardAutomaticBadges checks and awards automatic badges based on user activity
func (r *ReputationRepository) CheckAndAwardAutomaticBadges(ctx context.Context, userID uuid.UUID) ([]string, error) {
	// Get user data
	var karma int
	var createdAt time.Time
	err := r.db.QueryRow(ctx, `
		SELECT karma_points, created_at FROM users WHERE id = $1
	`, userID).Scan(&karma, &createdAt)
	if err != nil {
		return nil, fmt.Errorf("failed to get user data: %w", err)
	}

	// Get user stats
	stats, err := r.GetUserStats(ctx, userID)
	if err != nil {
		stats = &models.UserStats{UserID: userID}
	}

	var awardedBadges []string

	// Check karma-based badges
	if karma >= 10000 {
		if err := r.AwardBadge(ctx, userID, "influencer", nil); err == nil {
			awardedBadges = append(awardedBadges, "influencer")
		}
	}
	if karma >= 1000 {
		if err := r.AwardBadge(ctx, userID, "trusted_user", nil); err == nil {
			awardedBadges = append(awardedBadges, "trusted_user")
		}
	}

	// Check account age badge (1 year = 365 days)
	accountAge := time.Since(createdAt)
	if accountAge.Hours()/24 >= 365 {
		if err := r.AwardBadge(ctx, userID, "veteran", nil); err == nil {
			awardedBadges = append(awardedBadges, "veteran")
		}
	}

	// Check activity badges
	if stats.TotalComments >= 100 {
		if err := r.AwardBadge(ctx, userID, "conversationalist", nil); err == nil {
			awardedBadges = append(awardedBadges, "conversationalist")
		}
	}
	if stats.TotalVotesCast >= 1000 {
		if err := r.AwardBadge(ctx, userID, "curator", nil); err == nil {
			awardedBadges = append(awardedBadges, "curator")
		}
	}
	if stats.TotalClipsSubmit >= 50 {
		if err := r.AwardBadge(ctx, userID, "submitter", nil); err == nil {
			awardedBadges = append(awardedBadges, "submitter")
		}
	}

	return awardedBadges, nil
}

// CalculateTrustScoreBreakdown calculates trust score with detailed component breakdown
func (r *ReputationRepository) CalculateTrustScoreBreakdown(ctx context.Context, userID uuid.UUID) (*models.TrustScoreBreakdown, error) {
	query := `
		SELECT
			u.karma_points,
			u.is_banned,
			EXTRACT(EPOCH FROM (NOW() - u.created_at))::INT / 86400 as account_age_days,
			COALESCE(us.correct_reports, 0) as correct_reports,
			COALESCE(us.incorrect_reports, 0) as incorrect_reports,
			COALESCE(us.total_comments, 0) as total_comments,
			COALESCE(us.total_votes_cast, 0) as total_votes_cast,
			COALESCE(us.days_active, 0) as days_active
		FROM users u
		LEFT JOIN user_stats us ON u.id = us.user_id
		WHERE u.id = $1
	`

	var breakdown models.TrustScoreBreakdown
	var accountAgeDays, correctReports, incorrectReports, totalComments, totalVotes, daysActive int
	var karmaPoints int
	var isBanned bool

	err := r.db.QueryRow(ctx, query, userID).Scan(
		&karmaPoints,
		&isBanned,
		&accountAgeDays,
		&correctReports,
		&incorrectReports,
		&totalComments,
		&totalVotes,
		&daysActive,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get user data for trust score: %w", err)
	}

	// Store raw data
	breakdown.AccountAgeDays = accountAgeDays
	breakdown.KarmaPoints = karmaPoints
	breakdown.CorrectReports = correctReports
	breakdown.IncorrectReports = incorrectReports
	breakdown.TotalComments = totalComments
	breakdown.TotalVotes = totalVotes
	breakdown.DaysActive = daysActive
	breakdown.IsBanned = isBanned
	breakdown.MaxScore = 100

	// Calculate component scores (matching database function logic)
	// Account age contribution (max 20 points)
	// Database uses: accountAgeDays / 18, we use proportional calculation
	if accountAgeDays >= 365 {
		breakdown.AccountAgeScore = 20
	} else {
		breakdown.AccountAgeScore = (accountAgeDays * 20) / 365
	}

	// Karma contribution (max 40 points)
	// Database uses: karma / 250, we use proportional calculation
	if karmaPoints >= 10000 {
		breakdown.KarmaScore = 40
	} else {
		breakdown.KarmaScore = (karmaPoints * 40) / 10000
	}

	// Report accuracy contribution (max 20 points)
	totalReports := correctReports + incorrectReports
	if totalReports > 0 {
		breakdown.ReportAccuracy = int((20.0 * float64(correctReports)) / float64(totalReports))
	} else {
		breakdown.ReportAccuracy = 0
	}

	// Activity contribution (max 20 points)
	activityScore := (totalComments / 10) + (totalVotes / 100) + (daysActive / 5)
	breakdown.ActivityScore = min(activityScore, 20)

	// Calculate total score
	breakdown.TotalScore = breakdown.AccountAgeScore + breakdown.KarmaScore + breakdown.ReportAccuracy + breakdown.ActivityScore

	// Apply penalty for banned users
	if isBanned {
		breakdown.BanPenalty = 0.5
		breakdown.TotalScore = breakdown.TotalScore / 2
	}

	// Clamp to 0-100 range
	if breakdown.TotalScore < 0 {
		breakdown.TotalScore = 0
	}
	if breakdown.TotalScore > 100 {
		breakdown.TotalScore = 100
	}

	return &breakdown, nil
}

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// UpdateUserTrustScore updates a user's trust score and logs the change
func (r *ReputationRepository) UpdateUserTrustScore(
	ctx context.Context,
	userID uuid.UUID,
	newScore int,
	reason string,
	componentScores map[string]interface{},
	changedBy *uuid.UUID,
	notes *string,
) error {
	// Convert component scores to JSONB
	var componentScoresJSON interface{}
	if componentScores != nil {
		componentScoresJSON = componentScores
	}

	query := `SELECT update_user_trust_score($1, $2, $3, $4, $5, $6)`

	_, err := r.db.Exec(ctx, query, userID, newScore, reason, componentScoresJSON, changedBy, notes)
	if err != nil {
		return fmt.Errorf("failed to update user trust score: %w", err)
	}

	return nil
}

// GetTrustScoreHistory retrieves trust score history for a user
func (r *ReputationRepository) GetTrustScoreHistory(ctx context.Context, userID uuid.UUID, limit int) ([]models.TrustScoreHistory, error) {
	query := `
		SELECT id, user_id, old_score, new_score, change_reason, 
		       component_scores, changed_by, notes, created_at
		FROM trust_score_history
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT $2
	`

	rows, err := r.db.Query(ctx, query, userID, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query trust score history: %w", err)
	}
	defer rows.Close()

	var history []models.TrustScoreHistory
	for rows.Next() {
		var h models.TrustScoreHistory
		var componentScoresJSON []byte

		err := rows.Scan(
			&h.ID,
			&h.UserID,
			&h.OldScore,
			&h.NewScore,
			&h.ChangeReason,
			&componentScoresJSON,
			&h.ChangedBy,
			&h.Notes,
			&h.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan trust score history: %w", err)
		}

		// Parse component scores JSON if present
		if len(componentScoresJSON) > 0 {
			var scores map[string]interface{}
			if err := json.Unmarshal(componentScoresJSON, &scores); err == nil {
				h.ComponentScores = scores
			}
		}

		history = append(history, h)
	}

	return history, rows.Err()
}

// GetTrustScoreLeaderboard returns top users by trust score
func (r *ReputationRepository) GetTrustScoreLeaderboard(ctx context.Context, limit, offset int) ([]models.LeaderboardEntry, error) {
	query := `
		SELECT id, username, display_name, avatar_url, trust_score, karma_points, rank
		FROM trust_score_leaderboard
		LIMIT $1 OFFSET $2
	`

	rows, err := r.db.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to query trust score leaderboard: %w", err)
	}
	defer rows.Close()

	var entries []models.LeaderboardEntry
	rank := offset + 1
	for rows.Next() {
		var entry models.LeaderboardEntry
		var trustScore int
		var karmaPoints int

		err := rows.Scan(
			&entry.UserID,
			&entry.Username,
			&entry.DisplayName,
			&entry.AvatarURL,
			&trustScore,
			&karmaPoints,
			&entry.UserRank,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan leaderboard entry: %w", err)
		}
		entry.Rank = rank
		entry.Score = trustScore // Use trust score as the primary score
		rank++
		entries = append(entries, entry)
	}

	return entries, rows.Err()
}
