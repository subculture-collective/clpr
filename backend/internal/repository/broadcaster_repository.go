package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/subculture-collective/clipper/internal/models"
)

// BroadcasterRepository handles database operations for broadcasters
type BroadcasterRepository struct {
	pool *pgxpool.Pool
}

// NewBroadcasterRepository creates a new broadcaster repository
func NewBroadcasterRepository(pool *pgxpool.Pool) *BroadcasterRepository {
	return &BroadcasterRepository{pool: pool}
}

// FollowBroadcaster adds a follow relationship between a user and broadcaster
func (r *BroadcasterRepository) FollowBroadcaster(ctx context.Context, userID uuid.UUID, broadcasterID, broadcasterName string) error {
	query := `
		INSERT INTO broadcaster_follows (user_id, broadcaster_id, broadcaster_name)
		VALUES ($1, $2, $3)
		ON CONFLICT (user_id, broadcaster_id) DO NOTHING
	`
	_, err := r.pool.Exec(ctx, query, userID, broadcasterID, broadcasterName)
	if err != nil {
		return fmt.Errorf("failed to follow broadcaster: %w", err)
	}
	return nil
}

// UnfollowBroadcaster removes a follow relationship between a user and broadcaster
func (r *BroadcasterRepository) UnfollowBroadcaster(ctx context.Context, userID uuid.UUID, broadcasterID string) error {
	query := `
		DELETE FROM broadcaster_follows
		WHERE user_id = $1 AND broadcaster_id = $2
	`
	result, err := r.pool.Exec(ctx, query, userID, broadcasterID)
	if err != nil {
		return fmt.Errorf("failed to unfollow broadcaster: %w", err)
	}
	if result.RowsAffected() == 0 {
		return errors.New("follow relationship not found")
	}
	return nil
}

// IsFollowing checks if a user is following a broadcaster
func (r *BroadcasterRepository) IsFollowing(ctx context.Context, userID uuid.UUID, broadcasterID string) (bool, error) {
	query := `
		SELECT EXISTS(
			SELECT 1 FROM broadcaster_follows
			WHERE user_id = $1 AND broadcaster_id = $2
		)
	`
	var exists bool
	err := r.pool.QueryRow(ctx, query, userID, broadcasterID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check follow status: %w", err)
	}
	return exists, nil
}

// GetFollowerCount returns the number of followers for a broadcaster
func (r *BroadcasterRepository) GetFollowerCount(ctx context.Context, broadcasterID string) (int, error) {
	query := `
		SELECT COUNT(*) FROM broadcaster_follows
		WHERE broadcaster_id = $1
	`
	var count int
	err := r.pool.QueryRow(ctx, query, broadcasterID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to get follower count: %w", err)
	}
	return count, nil
}

// GetBroadcasterStats returns statistics for a broadcaster from the clips table
func (r *BroadcasterRepository) GetBroadcasterStats(ctx context.Context, broadcasterID string) (totalClips int, totalViews int64, avgVoteScore float64, err error) {
	query := `
		SELECT
			COUNT(*) as total_clips,
			COALESCE(SUM(view_count), 0) as total_views,
			COALESCE(AVG(vote_score), 0) as avg_vote_score
		FROM clips
		WHERE broadcaster_id = $1 AND is_removed = false
	`
	err = r.pool.QueryRow(ctx, query, broadcasterID).Scan(&totalClips, &totalViews, &avgVoteScore)
	if err != nil && err != pgx.ErrNoRows {
		return 0, 0, 0, fmt.Errorf("failed to get broadcaster stats: %w", err)
	}
	return totalClips, totalViews, avgVoteScore, nil
}

// GetBroadcasterByName returns broadcaster ID from clips table by name.
// Display name should be fetched from Twitch API separately.
func (r *BroadcasterRepository) GetBroadcasterByName(ctx context.Context, broadcasterName string) (broadcasterID string, err error) {
	query := `
		SELECT broadcaster_id
		FROM clips
		WHERE broadcaster_name = $1
		LIMIT 1
	`
	err = r.pool.QueryRow(ctx, query, broadcasterName).Scan(&broadcasterID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return "", sql.ErrNoRows
		}
		return "", fmt.Errorf("failed to get broadcaster by name: %w", err)
	}
	return broadcasterID, nil
}

// GetBroadcasterByID returns broadcaster name from clips.
// Display name should be fetched from Twitch API separately.
func (r *BroadcasterRepository) GetBroadcasterByID(ctx context.Context, broadcasterID string) (broadcasterName string, err error) {
	query := `
		SELECT broadcaster_name
		FROM clips
		WHERE broadcaster_id = $1
		LIMIT 1
	`
	err = r.pool.QueryRow(ctx, query, broadcasterID).Scan(&broadcasterName)
	if err != nil {
		if err == pgx.ErrNoRows {
			return "", sql.ErrNoRows
		}
		return "", fmt.Errorf("failed to get broadcaster by id: %w", err)
	}
	return broadcasterName, nil
}

// ListUserFollows returns all broadcasters a user is following
func (r *BroadcasterRepository) ListUserFollows(ctx context.Context, userID uuid.UUID, limit, offset int) ([]models.BroadcasterFollow, int, error) {
	// Get total count
	countQuery := `SELECT COUNT(*) FROM broadcaster_follows WHERE user_id = $1`
	var total int
	if err := r.pool.QueryRow(ctx, countQuery, userID).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("failed to count user follows: %w", err)
	}

	// Get paginated results
	query := `
		SELECT id, user_id, broadcaster_id, broadcaster_name, created_at
		FROM broadcaster_follows
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`
	rows, err := r.pool.Query(ctx, query, userID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list user follows: %w", err)
	}
	defer rows.Close()

	var follows []models.BroadcasterFollow
	for rows.Next() {
		var follow models.BroadcasterFollow
		if err := rows.Scan(&follow.ID, &follow.UserID, &follow.BroadcasterID, &follow.BroadcasterName, &follow.CreatedAt); err != nil {
			return nil, 0, fmt.Errorf("failed to scan follow: %w", err)
		}
		follows = append(follows, follow)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("error iterating follows: %w", err)
	}

	return follows, total, nil
}

// UpsertLiveStatus updates or inserts broadcaster live status
func (r *BroadcasterRepository) UpsertLiveStatus(ctx context.Context, status *models.BroadcasterLiveStatus) error {
	query := `
		INSERT INTO broadcaster_live_status (
			broadcaster_id, user_login, user_name, is_live, stream_title, game_name, viewer_count, started_at, last_checked
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		ON CONFLICT (broadcaster_id)
		DO UPDATE SET
			user_login = EXCLUDED.user_login,
			user_name = EXCLUDED.user_name,
			is_live = EXCLUDED.is_live,
			stream_title = EXCLUDED.stream_title,
			game_name = EXCLUDED.game_name,
			viewer_count = EXCLUDED.viewer_count,
			started_at = EXCLUDED.started_at,
			last_checked = EXCLUDED.last_checked,
			updated_at = NOW()
	`
	_, err := r.pool.Exec(ctx, query,
		status.BroadcasterID,
		status.UserLogin,
		status.UserName,
		status.IsLive,
		status.StreamTitle,
		status.GameName,
		status.ViewerCount,
		status.StartedAt,
		status.LastChecked,
	)
	if err != nil {
		return fmt.Errorf("failed to upsert live status: %w", err)
	}
	return nil
}

// GetLiveStatus retrieves live status for a broadcaster
func (r *BroadcasterRepository) GetLiveStatus(ctx context.Context, broadcasterID string) (*models.BroadcasterLiveStatus, error) {
	query := `
		SELECT broadcaster_id, user_login, user_name, is_live, stream_title, game_name, viewer_count,
		       started_at, last_checked, created_at, updated_at
		FROM broadcaster_live_status
		WHERE broadcaster_id = $1
	`
	var status models.BroadcasterLiveStatus
	err := r.pool.QueryRow(ctx, query, broadcasterID).Scan(
		&status.BroadcasterID,
		&status.UserLogin,
		&status.UserName,
		&status.IsLive,
		&status.StreamTitle,
		&status.GameName,
		&status.ViewerCount,
		&status.StartedAt,
		&status.LastChecked,
		&status.CreatedAt,
		&status.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, sql.ErrNoRows
		}
		return nil, fmt.Errorf("failed to get live status: %w", err)
	}
	return &status, nil
}

// ListLiveBroadcasters retrieves all currently live broadcasters
func (r *BroadcasterRepository) ListLiveBroadcasters(ctx context.Context, limit, offset int) ([]models.BroadcasterLiveStatus, int, error) {
	// Get total count of live broadcasters
	countQuery := `SELECT COUNT(*) FROM broadcaster_live_status WHERE is_live = true`
	var total int
	if err := r.pool.QueryRow(ctx, countQuery).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("failed to count live broadcasters: %w", err)
	}

	// Get paginated results, ordered by viewer count
	query := `
		SELECT broadcaster_id, user_login, user_name, is_live, stream_title, game_name, viewer_count,
		       started_at, last_checked, created_at, updated_at
		FROM broadcaster_live_status
		WHERE is_live = true
		ORDER BY viewer_count DESC
		LIMIT $1 OFFSET $2
	`
	rows, err := r.pool.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list live broadcasters: %w", err)
	}
	defer rows.Close()

	var broadcasters []models.BroadcasterLiveStatus
	for rows.Next() {
		var status models.BroadcasterLiveStatus
		if err := rows.Scan(
			&status.BroadcasterID,
			&status.UserLogin,
			&status.UserName,
			&status.IsLive,
			&status.StreamTitle,
			&status.GameName,
			&status.ViewerCount,
			&status.StartedAt,
			&status.LastChecked,
			&status.CreatedAt,
			&status.UpdatedAt,
		); err != nil {
			return nil, 0, fmt.Errorf("failed to scan live broadcaster: %w", err)
		}
		broadcasters = append(broadcasters, status)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("error iterating live broadcasters: %w", err)
	}

	return broadcasters, total, nil
}

// GetFollowedLiveBroadcasters retrieves live broadcasters that a user follows
func (r *BroadcasterRepository) GetFollowedLiveBroadcasters(ctx context.Context, userID uuid.UUID) ([]models.BroadcasterLiveStatus, error) {
	query := `
		SELECT bls.broadcaster_id, bls.user_login, bls.user_name, bls.is_live, bls.stream_title, bls.game_name,
		       bls.viewer_count, bls.started_at, bls.last_checked,
		       bls.created_at, bls.updated_at
		FROM broadcaster_live_status bls
		INNER JOIN broadcaster_follows bf ON bls.broadcaster_id = bf.broadcaster_id
		WHERE bf.user_id = $1 AND bls.is_live = true
		ORDER BY bls.viewer_count DESC
	`
	rows, err := r.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get followed live broadcasters: %w", err)
	}
	defer rows.Close()

	var broadcasters []models.BroadcasterLiveStatus
	for rows.Next() {
		var status models.BroadcasterLiveStatus
		if err := rows.Scan(
			&status.BroadcasterID,
			&status.UserLogin,
			&status.UserName,
			&status.IsLive,
			&status.StreamTitle,
			&status.GameName,
			&status.ViewerCount,
			&status.StartedAt,
			&status.LastChecked,
			&status.CreatedAt,
			&status.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan followed live broadcaster: %w", err)
		}
		broadcasters = append(broadcasters, status)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating followed live broadcasters: %w", err)
	}

	return broadcasters, nil
}

// GetFollowedBroadcasterIDs retrieves broadcaster IDs that a user follows
func (r *BroadcasterRepository) GetFollowedBroadcasterIDs(ctx context.Context, userID uuid.UUID) ([]string, error) {
	query := `
		SELECT broadcaster_id
		FROM broadcaster_follows
		WHERE user_id = $1
	`
	rows, err := r.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get followed broadcaster IDs: %w", err)
	}
	defer rows.Close()

	var broadcasterIDs []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("failed to scan broadcaster ID: %w", err)
		}
		broadcasterIDs = append(broadcasterIDs, id)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating broadcaster IDs: %w", err)
	}

	return broadcasterIDs, nil
}

// GetAllFollowedBroadcasterIDs retrieves all unique broadcaster IDs that are followed by any user
func (r *BroadcasterRepository) GetAllFollowedBroadcasterIDs(ctx context.Context) ([]string, error) {
	query := `
		SELECT DISTINCT broadcaster_id
		FROM broadcaster_follows
	`
	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get all followed broadcaster IDs: %w", err)
	}
	defer rows.Close()

	var broadcasterIDs []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("failed to scan broadcaster ID: %w", err)
		}
		broadcasterIDs = append(broadcasterIDs, id)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating broadcaster IDs: %w", err)
	}

	return broadcasterIDs, nil
}

// UpsertSyncStatus updates or inserts broadcaster sync status
func (r *BroadcasterRepository) UpsertSyncStatus(ctx context.Context, status *models.BroadcasterSyncStatus) error {
	query := `
		INSERT INTO broadcaster_sync_status (
			broadcaster_id, is_live, stream_started_at, last_synced, game_name, viewer_count, stream_title
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7
		)
		ON CONFLICT (broadcaster_id)
		DO UPDATE SET
			is_live = EXCLUDED.is_live,
			stream_started_at = EXCLUDED.stream_started_at,
			last_synced = EXCLUDED.last_synced,
			game_name = EXCLUDED.game_name,
			viewer_count = EXCLUDED.viewer_count,
			stream_title = EXCLUDED.stream_title,
			updated_at = NOW()
	`
	_, err := r.pool.Exec(ctx, query,
		status.BroadcasterID,
		status.IsLive,
		status.StreamStartedAt,
		status.LastSynced,
		status.GameName,
		status.ViewerCount,
		status.StreamTitle,
	)
	if err != nil {
		return fmt.Errorf("failed to upsert sync status: %w", err)
	}
	return nil
}

// GetSyncStatus retrieves sync status for a broadcaster
func (r *BroadcasterRepository) GetSyncStatus(ctx context.Context, broadcasterID string) (*models.BroadcasterSyncStatus, error) {
	query := `
SELECT broadcaster_id, is_live, stream_started_at, last_synced, game_name, viewer_count,
       stream_title, created_at, updated_at
FROM broadcaster_sync_status
WHERE broadcaster_id = $1
`
	var status models.BroadcasterSyncStatus
	err := r.pool.QueryRow(ctx, query, broadcasterID).Scan(
		&status.BroadcasterID,
		&status.IsLive,
		&status.StreamStartedAt,
		&status.LastSynced,
		&status.GameName,
		&status.ViewerCount,
		&status.StreamTitle,
		&status.CreatedAt,
		&status.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, sql.ErrNoRows
		}
		return nil, fmt.Errorf("failed to get sync status: %w", err)
	}
	return &status, nil
}

// CreateSyncLog creates a new sync log entry
func (r *BroadcasterRepository) CreateSyncLog(ctx context.Context, log *models.BroadcasterSyncLog) error {
	query := `
INSERT INTO broadcaster_sync_log (id, broadcaster_id, sync_time, status_change, error)
VALUES ($1, $2, $3, $4, $5)
`
	_, err := r.pool.Exec(ctx, query,
		log.ID,
		log.BroadcasterID,
		log.SyncTime,
		log.StatusChange,
		log.Error,
	)
	if err != nil {
		return fmt.Errorf("failed to create sync log: %w", err)
	}
	return nil
}

// GetFollowerUserIDs retrieves user IDs that follow a broadcaster
func (r *BroadcasterRepository) GetFollowerUserIDs(ctx context.Context, broadcasterID string) ([]uuid.UUID, error) {
	query := `
SELECT user_id
FROM broadcaster_follows
WHERE broadcaster_id = $1
`
	rows, err := r.pool.Query(ctx, query, broadcasterID)
	if err != nil {
		return nil, fmt.Errorf("failed to get follower user IDs: %w", err)
	}
	defer rows.Close()

	var userIDs []uuid.UUID
	for rows.Next() {
		var id uuid.UUID
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("failed to scan user ID: %w", err)
		}
		userIDs = append(userIDs, id)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating user IDs: %w", err)
	}

	return userIDs, nil
}

// ListBroadcasterGames returns games a broadcaster has clips in, ordered by clip count.
func (r *BroadcasterRepository) ListBroadcasterGames(ctx context.Context, broadcasterID string) ([]models.GameWithClipCount, error) {
	query := `
		SELECT c.game_id, COALESCE(g.name, c.game_name, 'Unknown'), COUNT(*) as clip_count, g.box_art_url
		FROM clips c
		LEFT JOIN games g ON g.twitch_game_id = c.game_id
		WHERE c.broadcaster_id = $1 AND c.is_removed = false AND c.game_id IS NOT NULL
		GROUP BY c.game_id, g.name, c.game_name, g.box_art_url
		ORDER BY clip_count DESC
		LIMIT 50
	`
	rows, err := r.pool.Query(ctx, query, broadcasterID)
	if err != nil {
		return nil, fmt.Errorf("failed to list broadcaster games: %w", err)
	}
	defer rows.Close()

	var games []models.GameWithClipCount
	for rows.Next() {
		var g models.GameWithClipCount
		if err := rows.Scan(&g.GameID, &g.Name, &g.ClipCount, &g.BoxArtURL); err != nil {
			return nil, fmt.Errorf("failed to scan game: %w", err)
		}
		games = append(games, g)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating games: %w", err)
	}
	return games, nil
}

// GetRankedBroadcasters returns broadcasters ordered by engagement score
func (r *BroadcasterRepository) GetRankedBroadcasters(ctx context.Context, limit, offset int) ([]models.BroadcasterRanking, int, error) {
	countQuery := `SELECT COUNT(*) FROM broadcaster_rankings`
	var total int
	if err := r.pool.QueryRow(ctx, countQuery).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("failed to count broadcaster rankings: %w", err)
	}

	query := `
		SELECT broadcaster_id, broadcaster_name, total_clips, human_submitted_clips,
		       total_vote_score, total_views, total_comments, unique_commenters,
		       engagement_score, follower_count, last_calculated
		FROM broadcaster_rankings
		ORDER BY engagement_score DESC
		LIMIT $1 OFFSET $2
	`
	rows, err := r.pool.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get ranked broadcasters: %w", err)
	}
	defer rows.Close()

	var rankings []models.BroadcasterRanking
	for rows.Next() {
		var br models.BroadcasterRanking
		if err := rows.Scan(
			&br.BroadcasterID, &br.BroadcasterName, &br.TotalClips, &br.HumanSubmittedClips,
			&br.TotalVoteScore, &br.TotalViews, &br.TotalComments, &br.UniqueCommenters,
			&br.EngagementScore, &br.FollowerCount, &br.LastCalculated,
		); err != nil {
			return nil, 0, fmt.Errorf("failed to scan broadcaster ranking: %w", err)
		}
		rankings = append(rankings, br)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("error iterating broadcaster rankings: %w", err)
	}

	return rankings, total, nil
}

// GetBroadcasterRank returns the ranking for a specific broadcaster
func (r *BroadcasterRepository) GetBroadcasterRank(ctx context.Context, broadcasterID string) (*models.BroadcasterRanking, error) {
	query := `
		SELECT broadcaster_id, broadcaster_name, total_clips, human_submitted_clips,
		       total_vote_score, total_views, total_comments, unique_commenters,
		       engagement_score, follower_count, last_calculated
		FROM broadcaster_rankings
		WHERE broadcaster_id = $1
	`
	var rank models.BroadcasterRanking
	err := r.pool.QueryRow(ctx, query, broadcasterID).Scan(
		&rank.BroadcasterID, &rank.BroadcasterName, &rank.TotalClips, &rank.HumanSubmittedClips,
		&rank.TotalVoteScore, &rank.TotalViews, &rank.TotalComments, &rank.UniqueCommenters,
		&rank.EngagementScore, &rank.FollowerCount, &rank.LastCalculated,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get broadcaster rank: %w", err)
	}
	return &rank, nil
}

// RefreshRankings refreshes the broadcaster_rankings materialized view
func (r *BroadcasterRepository) RefreshRankings(ctx context.Context) error {
	_, err := r.pool.Exec(ctx, "SELECT refresh_broadcaster_rankings()")
	if err != nil {
		return fmt.Errorf("failed to refresh broadcaster rankings: %w", err)
	}
	return nil
}

// ListPopularBroadcasters returns broadcasters ordered by clip count
func (r *BroadcasterRepository) ListPopularBroadcasters(ctx context.Context, limit int) ([]models.PopularBroadcaster, error) {
	if limit < 1 || limit > 50 {
		limit = 15
	}
	query := `
		SELECT broadcaster_id, broadcaster_name, COUNT(*) as clip_count
		FROM clips
		WHERE is_removed = false AND is_hidden = false AND broadcaster_id IS NOT NULL AND broadcaster_name != ''
		GROUP BY broadcaster_id, broadcaster_name
		ORDER BY clip_count DESC
		LIMIT $1
	`
	rows, err := r.pool.Query(ctx, query, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to list popular broadcasters: %w", err)
	}
	defer rows.Close()

	broadcasters := make([]models.PopularBroadcaster, 0)
	for rows.Next() {
		var b models.PopularBroadcaster
		if err := rows.Scan(&b.BroadcasterID, &b.BroadcasterName, &b.ClipCount); err != nil {
			return nil, fmt.Errorf("failed to scan broadcaster: %w", err)
		}
		broadcasters = append(broadcasters, b)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating broadcasters: %w", err)
	}
	return broadcasters, nil
}
