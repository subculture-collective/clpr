package services

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"git.subcult.tv/subculture-collective/clpr/internal/repository"
	"git.subcult.tv/subculture-collective/clpr/pkg/utils"
)

// AccountMergeService handles merging unclaimed accounts into authenticated accounts
type AccountMergeService struct {
	db               *pgxpool.Pool
	userRepo         *repository.UserRepository
	auditLogRepo     *repository.AuditLogRepository
	voteRepo         *repository.VoteRepository
	favoriteRepo     *repository.FavoriteRepository
	commentRepo      *repository.CommentRepository
	clipRepo         *repository.ClipRepository
	watchHistoryRepo *repository.WatchHistoryRepository
}

// NewAccountMergeService creates a new AccountMergeService
func NewAccountMergeService(
	db *pgxpool.Pool,
	userRepo *repository.UserRepository,
	auditLogRepo *repository.AuditLogRepository,
	voteRepo *repository.VoteRepository,
	favoriteRepo *repository.FavoriteRepository,
	commentRepo *repository.CommentRepository,
	clipRepo *repository.ClipRepository,
	watchHistoryRepo *repository.WatchHistoryRepository,
) *AccountMergeService {
	return &AccountMergeService{
		db:               db,
		userRepo:         userRepo,
		auditLogRepo:     auditLogRepo,
		voteRepo:         voteRepo,
		favoriteRepo:     favoriteRepo,
		commentRepo:      commentRepo,
		clipRepo:         clipRepo,
		watchHistoryRepo: watchHistoryRepo,
	}
}

// MergeResult represents the result of an account merge operation
type MergeResult struct {
	ClipsMerged        int
	VotesMerged        int
	FavoritesMerged    int
	CommentsMerged     int
	FollowsMerged      int
	WatchHistoryMerged int
	PreferencesMerged  bool
	SettingsMerged     bool
	SubscriptionMerged bool
	DuplicatesSkipped  int
	Success            bool
	Error              string
}

// MergeAccounts performs a complete merge of unclaimed account data into authenticated account
func (s *AccountMergeService) MergeAccounts(ctx context.Context, fromUserID, toUserID uuid.UUID) (*MergeResult, error) {
	result := &MergeResult{}

	// Begin transaction
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		if err != nil {
			if rbErr := tx.Rollback(ctx); rbErr != nil {
				utils.Error("Error rolling back transaction", rbErr, nil)
			}
		}
	}()

	// 1. Transfer clips
	clipsTransferred, err := s.transferClips(ctx, tx, fromUserID, toUserID)
	if err != nil {
		return nil, fmt.Errorf("failed to transfer clips: %w", err)
	}
	result.ClipsMerged = clipsTransferred

	// 2. Transfer votes (with duplicate handling)
	votesTransferred, duplicatesSkipped, err := s.transferVotes(ctx, tx, fromUserID, toUserID)
	if err != nil {
		return nil, fmt.Errorf("failed to transfer votes: %w", err)
	}
	result.VotesMerged = votesTransferred
	result.DuplicatesSkipped += duplicatesSkipped

	// 3. Transfer favorites (union of both sets)
	favoritesTransferred, favDuplicates, err := s.transferFavorites(ctx, tx, fromUserID, toUserID)
	if err != nil {
		return nil, fmt.Errorf("failed to transfer favorites: %w", err)
	}
	result.FavoritesMerged = favoritesTransferred
	result.DuplicatesSkipped += favDuplicates

	// 4. Transfer comments and comment votes
	commentsTransferred, err := s.transferComments(ctx, tx, fromUserID, toUserID)
	if err != nil {
		return nil, fmt.Errorf("failed to transfer comments: %w", err)
	}
	result.CommentsMerged = commentsTransferred

	// 5. Transfer follows (broadcaster, stream, game, user follows)
	followsTransferred, err := s.transferFollows(ctx, tx, fromUserID, toUserID)
	if err != nil {
		return nil, fmt.Errorf("failed to transfer follows: %w", err)
	}
	result.FollowsMerged = followsTransferred

	// 6. Transfer watch history (keep most recent per clip)
	watchHistoryTransferred, err := s.transferWatchHistory(ctx, tx, fromUserID, toUserID)
	if err != nil {
		return nil, fmt.Errorf("failed to transfer watch history: %w", err)
	}
	result.WatchHistoryMerged = watchHistoryTransferred

	// 7. Merge user preferences (arrays union)
	preferencesMerged, err := s.mergeUserPreferences(ctx, tx, fromUserID, toUserID)
	if err != nil {
		return nil, fmt.Errorf("failed to merge user preferences: %w", err)
	}
	result.PreferencesMerged = preferencesMerged

	// 8. Keep authenticated user's settings (don't override)
	// Settings are already set for authenticated user, no merge needed
	result.SettingsMerged = true

	// 9. Transfer subscription data if exists
	subscriptionMerged, err := s.transferSubscription(ctx, tx, fromUserID, toUserID)
	if err != nil {
		return nil, fmt.Errorf("failed to transfer subscription: %w", err)
	}
	result.SubscriptionMerged = subscriptionMerged

	// 10. Mark unclaimed account as merged
	if err := s.markAccountAsMerged(ctx, tx, fromUserID, toUserID); err != nil {
		return nil, fmt.Errorf("failed to mark account as merged: %w", err)
	}

	// 11. Create audit log entry
	if err := s.createMergeAuditLog(ctx, tx, fromUserID, toUserID, result); err != nil {
		// Log but don't fail the merge
		utils.Warn("Failed to create merge audit log", map[string]interface{}{"error": err})
	}

	// Commit transaction
	if err = tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	result.Success = true
	return result, nil
}

// transferClips transfers all clips submitted by the unclaimed account
func (s *AccountMergeService) transferClips(ctx context.Context, tx pgx.Tx, fromUserID, toUserID uuid.UUID) (int, error) {
	query := `
		UPDATE clips
		SET submitted_by_user_id = $1
		WHERE submitted_by_user_id = $2
	`

	cmdTag, err := tx.Exec(ctx, query, toUserID, fromUserID)
	if err != nil {
		return 0, err
	}

	return int(cmdTag.RowsAffected()), nil
}

// transferVotes transfers votes with duplicate handling (keep authenticated account vote)
func (s *AccountMergeService) transferVotes(ctx context.Context, tx pgx.Tx, fromUserID, toUserID uuid.UUID) (int, int, error) {
	// First, delete any votes from unclaimed account where authenticated account already has a vote
	deleteQuery := `
		DELETE FROM votes
		WHERE user_id = $1
		AND clip_id IN (
			SELECT clip_id FROM votes WHERE user_id = $2
		)
	`

	delCmdTag, err := tx.Exec(ctx, deleteQuery, fromUserID, toUserID)
	if err != nil {
		return 0, 0, err
	}
	duplicatesSkipped := int(delCmdTag.RowsAffected())

	// Transfer remaining votes
	updateQuery := `
		UPDATE votes
		SET user_id = $1
		WHERE user_id = $2
	`

	updCmdTag, err := tx.Exec(ctx, updateQuery, toUserID, fromUserID)
	if err != nil {
		return 0, 0, err
	}

	return int(updCmdTag.RowsAffected()), duplicatesSkipped, nil
}

// transferFavorites transfers favorites (union of both sets)
func (s *AccountMergeService) transferFavorites(ctx context.Context, tx pgx.Tx, fromUserID, toUserID uuid.UUID) (int, int, error) {
	// Count duplicates first
	countQuery := `
		SELECT COUNT(*)
		FROM favorites
		WHERE user_id = $1
		AND clip_id IN (
			SELECT clip_id FROM favorites WHERE user_id = $2
		)
	`

	var duplicates int
	if err := tx.QueryRow(ctx, countQuery, fromUserID, toUserID).Scan(&duplicates); err != nil {
		return 0, 0, err
	}

	// Delete duplicates from unclaimed account
	deleteQuery := `
		DELETE FROM favorites
		WHERE user_id = $1
		AND clip_id IN (
			SELECT clip_id FROM favorites WHERE user_id = $2
		)
	`

	if _, err := tx.Exec(ctx, deleteQuery, fromUserID, toUserID); err != nil {
		return 0, 0, err
	}

	// Transfer remaining favorites
	updateQuery := `
		UPDATE favorites
		SET user_id = $1
		WHERE user_id = $2
	`

	cmdTag, err := tx.Exec(ctx, updateQuery, toUserID, fromUserID)
	if err != nil {
		return 0, 0, err
	}

	return int(cmdTag.RowsAffected()), duplicates, nil
}

// transferComments transfers all comments and comment votes
func (s *AccountMergeService) transferComments(ctx context.Context, tx pgx.Tx, fromUserID, toUserID uuid.UUID) (int, error) {
	// Transfer comments
	commentsQuery := `
		UPDATE comments
		SET user_id = $1
		WHERE user_id = $2
	`

	cmdTag, err := tx.Exec(ctx, commentsQuery, toUserID, fromUserID)
	if err != nil {
		return 0, err
	}
	commentsTransferred := int(cmdTag.RowsAffected())

	// Transfer comment votes (with duplicate handling)
	// Delete duplicates first
	deleteVotesQuery := `
		DELETE FROM comment_votes
		WHERE user_id = $1
		AND comment_id IN (
			SELECT comment_id FROM comment_votes WHERE user_id = $2
		)
	`

	if _, err := tx.Exec(ctx, deleteVotesQuery, fromUserID, toUserID); err != nil {
		return 0, err
	}

	// Transfer remaining comment votes
	updateVotesQuery := `
		UPDATE comment_votes
		SET user_id = $1
		WHERE user_id = $2
	`

	if _, err := tx.Exec(ctx, updateVotesQuery, toUserID, fromUserID); err != nil {
		return 0, err
	}

	return commentsTransferred, nil
}

// transferFollows transfers all follow relationships (broadcaster, stream, game, user)
// Note: Individual follow table transfers are logged as warnings on failure but don't fail the transaction.
// This is intentional because some follow tables may not exist in all deployments (e.g., optional features).
// The transaction will still commit successfully even if some follow tables fail to transfer.
func (s *AccountMergeService) transferFollows(ctx context.Context, tx pgx.Tx, fromUserID, toUserID uuid.UUID) (int, error) {
	totalTransferred := 0

	// Handle broadcaster_follows: UNIQUE(user_id, broadcaster_id)
	if transferred, err := s.transferBroadcasterFollows(ctx, tx, fromUserID, toUserID); err != nil {
		utils.Warn("Failed to transfer broadcaster follows", map[string]interface{}{"error": err})
	} else {
		totalTransferred += transferred
	}

	// Handle stream_follows: UNIQUE(user_id, streamer_username)
	if transferred, err := s.transferStreamFollows(ctx, tx, fromUserID, toUserID); err != nil {
		utils.Warn("Failed to transfer stream follows", map[string]interface{}{"error": err})
	} else {
		totalTransferred += transferred
	}

	// Handle game_follows: UNIQUE(user_id, game_id)
	if transferred, err := s.transferGameFollows(ctx, tx, fromUserID, toUserID); err != nil {
		utils.Warn("Failed to transfer game follows", map[string]interface{}{"error": err})
	} else {
		totalTransferred += transferred
	}

	// Handle user_follows: UNIQUE(follower_id, following_id) - uses follower_id instead of user_id
	if transferred, err := s.transferUserFollows(ctx, tx, fromUserID, toUserID); err != nil {
		utils.Warn("Failed to transfer user follows", map[string]interface{}{"error": err})
	} else {
		totalTransferred += transferred
	}

	return totalTransferred, nil
}

// transferBroadcasterFollows handles broadcaster_follows table
func (s *AccountMergeService) transferBroadcasterFollows(ctx context.Context, tx pgx.Tx, fromUserID, toUserID uuid.UUID) (int, error) {
	// Check if table exists
	var exists bool
	err := tx.QueryRow(ctx,
		`SELECT EXISTS (
			SELECT FROM information_schema.tables
			WHERE table_schema = 'public' AND table_name = 'broadcaster_follows'
		)`).Scan(&exists)

	if err != nil {
		return 0, err
	}

	if !exists {
		return 0, nil
	}

	// Delete duplicates (same user_id and broadcaster_id)
	_, err = tx.Exec(ctx, `
		DELETE FROM broadcaster_follows
		WHERE user_id = $1
		AND broadcaster_id IN (
			SELECT broadcaster_id FROM broadcaster_follows WHERE user_id = $2
		)
	`, fromUserID, toUserID)

	if err != nil {
		return 0, err
	}

	// Transfer remaining follows
	cmdTag, err := tx.Exec(ctx, `
		UPDATE broadcaster_follows
		SET user_id = $1
		WHERE user_id = $2
	`, toUserID, fromUserID)

	if err != nil {
		return 0, err
	}

	return int(cmdTag.RowsAffected()), nil
}

// transferStreamFollows handles stream_follows table
func (s *AccountMergeService) transferStreamFollows(ctx context.Context, tx pgx.Tx, fromUserID, toUserID uuid.UUID) (int, error) {
	// Check if table exists
	var exists bool
	err := tx.QueryRow(ctx,
		`SELECT EXISTS (
			SELECT FROM information_schema.tables
			WHERE table_schema = 'public' AND table_name = 'stream_follows'
		)`).Scan(&exists)

	if err != nil {
		return 0, err
	}

	if !exists {
		return 0, nil
	}

	// Delete duplicates (same user_id and streamer_username)
	_, err = tx.Exec(ctx, `
		DELETE FROM stream_follows
		WHERE user_id = $1
		AND streamer_username IN (
			SELECT streamer_username FROM stream_follows WHERE user_id = $2
		)
	`, fromUserID, toUserID)

	if err != nil {
		return 0, err
	}

	// Transfer remaining follows
	cmdTag, err := tx.Exec(ctx, `
		UPDATE stream_follows
		SET user_id = $1
		WHERE user_id = $2
	`, toUserID, fromUserID)

	if err != nil {
		return 0, err
	}

	return int(cmdTag.RowsAffected()), nil
}

// transferGameFollows handles game_follows table
func (s *AccountMergeService) transferGameFollows(ctx context.Context, tx pgx.Tx, fromUserID, toUserID uuid.UUID) (int, error) {
	// Check if table exists
	var exists bool
	err := tx.QueryRow(ctx,
		`SELECT EXISTS (
			SELECT FROM information_schema.tables
			WHERE table_schema = 'public' AND table_name = 'game_follows'
		)`).Scan(&exists)

	if err != nil {
		return 0, err
	}

	if !exists {
		return 0, nil
	}

	// Delete duplicates (same user_id and game_id)
	_, err = tx.Exec(ctx, `
		DELETE FROM game_follows
		WHERE user_id = $1
		AND game_id IN (
			SELECT game_id FROM game_follows WHERE user_id = $2
		)
	`, fromUserID, toUserID)

	if err != nil {
		return 0, err
	}

	// Transfer remaining follows
	cmdTag, err := tx.Exec(ctx, `
		UPDATE game_follows
		SET user_id = $1
		WHERE user_id = $2
	`, toUserID, fromUserID)

	if err != nil {
		return 0, err
	}

	return int(cmdTag.RowsAffected()), nil
}

// transferUserFollows handles user_follows table (uses follower_id instead of user_id)
func (s *AccountMergeService) transferUserFollows(ctx context.Context, tx pgx.Tx, fromUserID, toUserID uuid.UUID) (int, error) {
	// Check if table exists
	var exists bool
	err := tx.QueryRow(ctx,
		`SELECT EXISTS (
			SELECT FROM information_schema.tables
			WHERE table_schema = 'public' AND table_name = 'user_follows'
		)`).Scan(&exists)

	if err != nil {
		return 0, err
	}

	if !exists {
		return 0, nil
	}

	totalTransferred := 0

	// Part 1: Transfer follows where unclaimed user is the follower
	// Delete duplicates (same follower_id and following_id)
	_, err = tx.Exec(ctx, `
		DELETE FROM user_follows
		WHERE follower_id = $1
		AND following_id IN (
			SELECT following_id FROM user_follows WHERE follower_id = $2
		)
	`, fromUserID, toUserID)

	if err != nil {
		return 0, err
	}

	// Transfer remaining follows (update follower_id)
	cmdTag, err := tx.Exec(ctx, `
		UPDATE user_follows
		SET follower_id = $1
		WHERE follower_id = $2
	`, toUserID, fromUserID)

	if err != nil {
		return 0, err
	}

	totalTransferred += int(cmdTag.RowsAffected())

	// Part 2: Transfer follows where unclaimed user is being followed
	// Delete duplicates (same follower_id and following_id)
	_, err = tx.Exec(ctx, `
		DELETE FROM user_follows
		WHERE following_id = $1
		AND follower_id IN (
			SELECT follower_id FROM user_follows WHERE following_id = $2
		)
	`, fromUserID, toUserID)

	if err != nil {
		return 0, err
	}

	// Transfer remaining follows (update following_id)
	cmdTag, err = tx.Exec(ctx, `
		UPDATE user_follows
		SET following_id = $1
		WHERE following_id = $2
	`, toUserID, fromUserID)

	if err != nil {
		return 0, err
	}

	totalTransferred += int(cmdTag.RowsAffected())

	return totalTransferred, nil
}

// transferWatchHistory transfers watch history, keeping authenticated user's version for conflicting clips
func (s *AccountMergeService) transferWatchHistory(ctx context.Context, tx pgx.Tx, fromUserID, toUserID uuid.UUID) (int, error) {
	// For clips where authenticated user already has watch history, keep the authenticated user's version
	// Delete unclaimed user's watch history for those clips
	deleteQuery := `
		DELETE FROM watch_history
		WHERE user_id = $1
		AND clip_id IN (
			SELECT clip_id FROM watch_history WHERE user_id = $2
		)
	`

	if _, err := tx.Exec(ctx, deleteQuery, fromUserID, toUserID); err != nil {
		return 0, err
	}

	// Transfer remaining watch history
	updateQuery := `
		UPDATE watch_history
		SET user_id = $1
		WHERE user_id = $2
	`

	cmdTag, err := tx.Exec(ctx, updateQuery, toUserID, fromUserID)
	if err != nil {
		return 0, err
	}

	return int(cmdTag.RowsAffected()), nil
}

// mergeUserPreferences merges user preferences (union of arrays)
func (s *AccountMergeService) mergeUserPreferences(ctx context.Context, tx pgx.Tx, fromUserID, toUserID uuid.UUID) (bool, error) {
	// Check if user_preferences table exists in the public schema
	checkQuery := `
		SELECT EXISTS (
			SELECT FROM information_schema.tables
			WHERE table_schema = 'public' AND table_name = 'user_preferences'
		)
	`

	var exists bool
	if err := tx.QueryRow(ctx, checkQuery).Scan(&exists); err != nil {
		return false, err
	}

	if !exists {
		return false, nil
	}

	// Merge preferences by taking union of arrays
	mergeQuery := `
		INSERT INTO user_preferences (
			user_id,
			favorite_games,
			followed_streamers,
			preferred_categories,
			preferred_tags,
			updated_at
		)
		SELECT
			$1 as user_id,
			ARRAY(SELECT DISTINCT unnest(
				COALESCE(to_prefs.favorite_games, '{}') ||
				COALESCE(from_prefs.favorite_games, '{}')
			)) as favorite_games,
			ARRAY(SELECT DISTINCT unnest(
				COALESCE(to_prefs.followed_streamers, '{}') ||
				COALESCE(from_prefs.followed_streamers, '{}')
			)) as followed_streamers,
			ARRAY(SELECT DISTINCT unnest(
				COALESCE(to_prefs.preferred_categories, '{}') ||
				COALESCE(from_prefs.preferred_categories, '{}')
			)) as preferred_categories,
			ARRAY(SELECT DISTINCT unnest(
				COALESCE(to_prefs.preferred_tags, '{}') ||
				COALESCE(from_prefs.preferred_tags, '{}')
			)) as preferred_tags,
			NOW() as updated_at
		FROM
			(SELECT * FROM user_preferences WHERE user_id = $1) to_prefs
		FULL OUTER JOIN
			(SELECT * FROM user_preferences WHERE user_id = $2) from_prefs
			ON true
		ON CONFLICT (user_id)
		DO UPDATE SET
			favorite_games = EXCLUDED.favorite_games,
			followed_streamers = EXCLUDED.followed_streamers,
			preferred_categories = EXCLUDED.preferred_categories,
			preferred_tags = EXCLUDED.preferred_tags,
			updated_at = NOW()
	`

	if _, err := tx.Exec(ctx, mergeQuery, toUserID, fromUserID); err != nil {
		return false, err
	}

	// Delete unclaimed user's preferences
	deleteQuery := `DELETE FROM user_preferences WHERE user_id = $1`
	if _, err := tx.Exec(ctx, deleteQuery, fromUserID); err != nil {
		utils.Warn("Failed to delete old preferences", map[string]interface{}{"error": err})
	}

	return true, nil
}

// transferSubscription transfers subscription data if it exists
func (s *AccountMergeService) transferSubscription(ctx context.Context, tx pgx.Tx, fromUserID, toUserID uuid.UUID) (bool, error) {
	// Check if subscriptions table exists in the public schema
	checkQuery := `
		SELECT EXISTS (
			SELECT FROM information_schema.tables
			WHERE table_schema = 'public' AND table_name = 'subscriptions'
		)
	`

	var exists bool
	if err := tx.QueryRow(ctx, checkQuery).Scan(&exists); err != nil {
		return false, err
	}

	if !exists {
		return false, nil
	}

	// Check if unclaimed user has an active subscription
	hasSubQuery := `
		SELECT EXISTS (
			SELECT 1 FROM subscriptions
			WHERE user_id = $1
			AND status IN ('active', 'trialing')
		)
	`

	var hasSubscription bool
	if err := tx.QueryRow(ctx, hasSubQuery, fromUserID).Scan(&hasSubscription); err != nil {
		return false, err
	}

	if !hasSubscription {
		return false, nil
	}

	// Check if authenticated user already has subscription
	hasToSubQuery := `
		SELECT EXISTS (
			SELECT 1 FROM subscriptions
			WHERE user_id = $1
		)
	`

	var toHasSubscription bool
	if err := tx.QueryRow(ctx, hasToSubQuery, toUserID).Scan(&toHasSubscription); err != nil {
		return false, err
	}

	// If authenticated user already has subscription, don't transfer (keep authenticated)
	if toHasSubscription {
		return false, nil
	}

	// Transfer active/trialing subscription only
	updateQuery := `
		UPDATE subscriptions
		SET user_id = $1
		WHERE user_id = $2
		AND status IN ('active', 'trialing')
	`

	if _, err := tx.Exec(ctx, updateQuery, toUserID, fromUserID); err != nil {
		return false, err
	}

	return true, nil
}

// markAccountAsMerged marks the unclaimed account as merged
func (s *AccountMergeService) markAccountAsMerged(ctx context.Context, tx pgx.Tx, fromUserID, toUserID uuid.UUID) error {
	query := `
		UPDATE users
		SET account_status = 'merged',
		    updated_at = NOW()
		WHERE id = $1
	`

	_, err := tx.Exec(ctx, query, fromUserID)
	return err
}

// createMergeAuditLog creates an audit log entry for the merge operation
func (s *AccountMergeService) createMergeAuditLog(ctx context.Context, tx pgx.Tx, fromUserID, toUserID uuid.UUID, result *MergeResult) error {
	metadata := map[string]interface{}{
		"from_user_id":         fromUserID.String(),
		"to_user_id":           toUserID.String(),
		"clips_merged":         result.ClipsMerged,
		"votes_merged":         result.VotesMerged,
		"favorites_merged":     result.FavoritesMerged,
		"comments_merged":      result.CommentsMerged,
		"follows_merged":       result.FollowsMerged,
		"watch_history_merged": result.WatchHistoryMerged,
		"preferences_merged":   result.PreferencesMerged,
		"subscription_merged":  result.SubscriptionMerged,
		"duplicates_skipped":   result.DuplicatesSkipped,
		"timestamp":            time.Now().UTC().Format(time.RFC3339),
	}

	// Marshal metadata to JSON for proper storage
	metadataJSON, err := json.Marshal(metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal audit log metadata: %w", err)
	}

	query := `
		INSERT INTO moderation_audit_logs (
			id, action, entity_type, entity_id, moderator_id, actor_id, metadata, created_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, NOW()
		)
	`

	// Note: Both moderator_id and actor_id are set to toUserID for backward compatibility
	// during the migration period. The user themselves initiated this merge by claiming
	// their account. This represents a self-service action.
	_, err = tx.Exec(ctx, query,
		uuid.New(),
		"account_merged",
		"user",
		toUserID,
		toUserID, // moderator_id (old column, kept for backward compatibility)
		toUserID, // actor_id (new column)
		metadataJSON,
	)

	return err
}
