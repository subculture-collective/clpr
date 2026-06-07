package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/subculture-collective/clipper/internal/models"
	"github.com/subculture-collective/clipper/internal/utils"
)

const (
	// HotClipsMaterializedView is the name of the materialized view for hot clips
	HotClipsMaterializedView = "hot_clips_materialized"

	// Pagination limits for clip queries
	DefaultClipLimit = 50
	MaxClipLimit     = 1000
	MaxClipOffset    = 1000
)

// ClipRepository handles database operations for clips
type ClipRepository struct {
	pool   *pgxpool.Pool
	helper *RepositoryHelper
}

// NewClipRepository creates a new ClipRepository
func NewClipRepository(pool *pgxpool.Pool) *ClipRepository {
	return &ClipRepository{
		pool:   pool,
		helper: NewRepositoryHelper(pool),
	}
}

// Create inserts a new clip into the database
func (r *ClipRepository) Create(ctx context.Context, clip *models.Clip) error {
	query := `
		INSERT INTO clips (
			id, twitch_clip_id, twitch_clip_url, embed_url, title,
			creator_name, creator_id, broadcaster_name, broadcaster_id,
			game_id, game_name, language, thumbnail_url, duration,
			view_count, created_at, imported_at, vote_score, comment_count, favorite_count,
			is_featured, is_nsfw, is_removed, is_hidden,
			submitted_by_user_id, submitted_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17,
			$18, $19, $20, $21, $22, $23, $24, $25, $26
		)
	`

	_, err := r.pool.Exec(ctx, query,
		clip.ID, clip.TwitchClipID, clip.TwitchClipURL, clip.EmbedURL,
		clip.Title, clip.CreatorName, clip.CreatorID, clip.BroadcasterName,
		clip.BroadcasterID, clip.GameID, clip.GameName, clip.Language,
		clip.ThumbnailURL, clip.Duration, clip.ViewCount, clip.CreatedAt,
		clip.ImportedAt, clip.VoteScore, clip.CommentCount, clip.FavoriteCount,
		clip.IsFeatured, clip.IsNSFW, clip.IsRemoved, clip.IsHidden,
		clip.SubmittedByUserID, clip.SubmittedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create clip: %w", err)
	}

	return nil
}

// CreateStreamClip inserts a new clip created from a stream into the database
func (r *ClipRepository) CreateStreamClip(ctx context.Context, clip *models.Clip) error {
	query := `
		INSERT INTO clips (
			id, twitch_clip_id, twitch_clip_url, embed_url, title,
			creator_name, creator_id, broadcaster_name, broadcaster_id,
			game_id, game_name, language, thumbnail_url, duration,
			view_count, created_at, imported_at, vote_score, comment_count, favorite_count,
			is_featured, is_nsfw, is_removed, is_hidden,
			submitted_by_user_id, submitted_at,
			stream_source, status, quality, start_time, end_time
		) VALUES (
			$1, $2, $3, $4, $5,
			$6, $7, $8, $9,
			$10, $11, $12, $13, $14,
			$15, $16, $17, $18, $19,
			$20, $21, $22, $23,
			$24, $25,
			$26, $27, $28, $29, $30, $31
		)
	`

	_, err := r.pool.Exec(ctx, query,
		clip.ID, clip.TwitchClipID, clip.TwitchClipURL, clip.EmbedURL, clip.Title,
		clip.CreatorName, clip.CreatorID, clip.BroadcasterName, clip.BroadcasterID,
		clip.GameID, clip.GameName, clip.Language, clip.ThumbnailURL, clip.Duration,
		clip.ViewCount, clip.CreatedAt, clip.ImportedAt, clip.VoteScore, clip.CommentCount, clip.FavoriteCount,
		clip.IsFeatured, clip.IsNSFW, clip.IsRemoved, clip.IsHidden,
		clip.SubmittedByUserID, clip.SubmittedAt,
		clip.StreamSource, clip.Status, clip.Quality, clip.StartTime, clip.EndTime,
	)

	if err != nil {
		return fmt.Errorf("failed to create stream clip: %w", err)
	}

	return nil
}

// GetByTwitchClipID retrieves a clip by its Twitch clip ID
func (r *ClipRepository) GetByTwitchClipID(ctx context.Context, twitchClipID string) (*models.Clip, error) {
	query := `
		SELECT
			id, twitch_clip_id, twitch_clip_url, embed_url, title,
			creator_name, creator_id, broadcaster_name, broadcaster_id,
			game_id, game_name, language, thumbnail_url, duration,
			view_count, created_at, imported_at, vote_score, comment_count,
			favorite_count, is_featured, is_nsfw, is_removed, removed_reason, is_hidden,
			submitted_by_user_id, submitted_at,
			stream_source, status, video_url, processed_at, quality, start_time, end_time
		FROM clips
		WHERE twitch_clip_id = $1
	`

	var clip models.Clip
	err := r.pool.QueryRow(ctx, query, twitchClipID).Scan(
		&clip.ID, &clip.TwitchClipID, &clip.TwitchClipURL, &clip.EmbedURL,
		&clip.Title, &clip.CreatorName, &clip.CreatorID, &clip.BroadcasterName,
		&clip.BroadcasterID, &clip.GameID, &clip.GameName, &clip.Language,
		&clip.ThumbnailURL, &clip.Duration, &clip.ViewCount, &clip.CreatedAt,
		&clip.ImportedAt, &clip.VoteScore, &clip.CommentCount, &clip.FavoriteCount,
		&clip.IsFeatured, &clip.IsNSFW, &clip.IsRemoved, &clip.RemovedReason, &clip.IsHidden,
		&clip.SubmittedByUserID, &clip.SubmittedAt,
		&clip.StreamSource, &clip.Status, &clip.VideoURL, &clip.ProcessedAt, &clip.Quality, &clip.StartTime, &clip.EndTime,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to get clip by twitch ID: %w", err)
	}

	return &clip, nil
}

// GetByTwitchClipIDs retrieves multiple clips by their Twitch clip IDs, preserving the input order.
func (r *ClipRepository) GetByTwitchClipIDs(ctx context.Context, twitchClipIDs []string) ([]models.Clip, error) {
	if len(twitchClipIDs) == 0 {
		return nil, nil
	}

	query := `
		SELECT
			id, twitch_clip_id, twitch_clip_url, embed_url, title,
			creator_name, creator_id, broadcaster_name, broadcaster_id,
			game_id, game_name, language, thumbnail_url, duration,
			view_count, created_at, imported_at, vote_score, comment_count,
			favorite_count, is_featured, is_nsfw, is_removed, removed_reason, is_hidden,
			submitted_by_user_id, submitted_at,
			stream_source, status, video_url, processed_at, quality, start_time, end_time
		FROM clips
		WHERE twitch_clip_id = ANY($1)
		  AND is_removed = false
	`

	rows, err := r.pool.Query(ctx, query, twitchClipIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to get clips by twitch IDs: %w", err)
	}
	defer rows.Close()

	// Build a map for order preservation
	clipMap := make(map[string]models.Clip, len(twitchClipIDs))
	for rows.Next() {
		var clip models.Clip
		err := rows.Scan(
			&clip.ID, &clip.TwitchClipID, &clip.TwitchClipURL, &clip.EmbedURL,
			&clip.Title, &clip.CreatorName, &clip.CreatorID, &clip.BroadcasterName,
			&clip.BroadcasterID, &clip.GameID, &clip.GameName, &clip.Language,
			&clip.ThumbnailURL, &clip.Duration, &clip.ViewCount, &clip.CreatedAt,
			&clip.ImportedAt, &clip.VoteScore, &clip.CommentCount, &clip.FavoriteCount,
			&clip.IsFeatured, &clip.IsNSFW, &clip.IsRemoved, &clip.RemovedReason, &clip.IsHidden,
			&clip.SubmittedByUserID, &clip.SubmittedAt,
			&clip.StreamSource, &clip.Status, &clip.VideoURL, &clip.ProcessedAt, &clip.Quality, &clip.StartTime, &clip.EndTime,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan clip: %w", err)
		}
		clipMap[clip.TwitchClipID] = clip
	}

	// Preserve input order (Twitch returns by view count desc)
	clips := make([]models.Clip, 0, len(clipMap))
	for _, tid := range twitchClipIDs {
		if clip, ok := clipMap[tid]; ok {
			clips = append(clips, clip)
		}
	}

	return clips, nil
}

// UpdateViewCount updates the view count for a clip
func (r *ClipRepository) UpdateViewCount(ctx context.Context, twitchClipID string, viewCount int) error {
	query := `
		UPDATE clips
		SET view_count = $2
		WHERE twitch_clip_id = $1
	`

	_, err := r.pool.Exec(ctx, query, twitchClipID, viewCount)
	if err != nil {
		return fmt.Errorf("failed to update view count: %w", err)
	}

	return nil
}

// ExistsByTwitchClipID checks if a clip exists by Twitch clip ID
func (r *ClipRepository) ExistsByTwitchClipID(ctx context.Context, twitchClipID string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM clips WHERE twitch_clip_id = $1)`

	var exists bool
	err := r.pool.QueryRow(ctx, query, twitchClipID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check clip existence: %w", err)
	}

	return exists, nil
}

// ClaimScrapedClip atomically updates a scraped clip to mark it as claimed by a user.
// This method performs a check-and-update operation to prevent race conditions where
// multiple users attempt to claim the same clip simultaneously.
//
// Parameters:
//   - ctx: Context for the database operation
//   - clipID: The UUID of the clip to claim
//   - userID: The UUID of the user claiming the clip
//   - title: Optional custom title to override the clip's current title (can be nil)
//   - isNSFW: Whether the clip should be marked as NSFW
//   - broadcasterName: Optional broadcaster name override (can be nil)
//   - submittedAt: The timestamp when the clip was claimed/submitted
//
// Returns:
//   - error: Returns "clip not found" if clip doesn't exist, or "clip has already been claimed by another user" if already claimed
//
// The WHERE clause (submitted_by_user_id IS NULL) ensures atomicity - the update will only
// succeed if the clip hasn't been claimed yet, preventing duplicate claims.
//
// Example usage:
//
//	err := repo.ClaimScrapedClip(ctx, clipID, userID, &customTitle, false, &broadcasterOverride, time.Now())
//	if err != nil {
//	    // Handle error - clip may not exist or already claimed
//	}
func (r *ClipRepository) ClaimScrapedClip(ctx context.Context, clipID uuid.UUID, userID uuid.UUID, title *string, isNSFW bool, broadcasterName *string, submittedAt time.Time) error {
	query := `
		UPDATE clips
		SET submitted_by_user_id = $2,
		    submitted_at = $3,
		    title = COALESCE($4, title),
		    is_nsfw = $5,
		    broadcaster_name = COALESCE($6, broadcaster_name)
		WHERE id = $1 AND submitted_by_user_id IS NULL
	`

	result, err := r.pool.Exec(ctx, query, clipID, userID, submittedAt, title, isNSFW, broadcasterName)
	if err != nil {
		return fmt.Errorf("failed to claim scraped clip: %w", err)
	}

	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		// Check if clip exists at all
		var exists bool
		checkQuery := `SELECT EXISTS(SELECT 1 FROM clips WHERE id = $1)`
		if checkErr := r.pool.QueryRow(ctx, checkQuery, clipID).Scan(&exists); checkErr != nil {
			return fmt.Errorf("failed to verify clip existence: %w", checkErr)
		}

		if !exists {
			return fmt.Errorf("clip not found")
		}

		// Clip exists but update didn't happen - must be already claimed
		return fmt.Errorf("clip has already been claimed by another user")
	}

	return nil
}

// GetByID retrieves a clip by its ID
func (r *ClipRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Clip, error) {
	query := `
		SELECT
			id, twitch_clip_id, twitch_clip_url, embed_url, title,
			creator_name, creator_id, broadcaster_name, broadcaster_id,
			game_id, game_name, language, thumbnail_url, duration,
			view_count, created_at, imported_at, vote_score, comment_count,
			favorite_count, is_featured, is_nsfw, is_removed, removed_reason, is_hidden,
			submitted_by_user_id, submitted_at,
			stream_source, status, video_url, processed_at, quality, start_time, end_time
		FROM clips
		WHERE id = $1 AND is_removed = false
	`

	var clip models.Clip
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&clip.ID, &clip.TwitchClipID, &clip.TwitchClipURL, &clip.EmbedURL,
		&clip.Title, &clip.CreatorName, &clip.CreatorID, &clip.BroadcasterName,
		&clip.BroadcasterID, &clip.GameID, &clip.GameName, &clip.Language,
		&clip.ThumbnailURL, &clip.Duration, &clip.ViewCount, &clip.CreatedAt,
		&clip.ImportedAt, &clip.VoteScore, &clip.CommentCount, &clip.FavoriteCount,
		&clip.IsFeatured, &clip.IsNSFW, &clip.IsRemoved, &clip.RemovedReason, &clip.IsHidden,
		&clip.SubmittedByUserID, &clip.SubmittedAt,
		&clip.StreamSource, &clip.Status, &clip.VideoURL, &clip.ProcessedAt, &clip.Quality, &clip.StartTime, &clip.EndTime,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to get clip by ID: %w", err)
	}

	return &clip, nil
}

// List retrieves clips with pagination
func (r *ClipRepository) List(ctx context.Context, limit, offset int) ([]models.Clip, error) {
	// Enforce pagination limits
	r.helper.EnforcePaginationLimits(&limit, &offset)

	// Delegate to ListWithFilters with empty filters for reuse and to reduce duplication.
	clips, total, err := r.ListWithFilters(ctx, ClipFilters{Sort: "new"}, limit, offset)
	if err != nil {
		return nil, err
	}
	_ = total // total is not needed for simple list
	return clips, nil
}

// GetRecentClips gets clips from the last N hours
func (r *ClipRepository) GetRecentClips(ctx context.Context, hours int, limit int) ([]models.Clip, error) {
	query := `
		SELECT
			id, twitch_clip_id, twitch_clip_url, embed_url, title,
			creator_name, creator_id, broadcaster_name, broadcaster_id,
			game_id, game_name, language, thumbnail_url, duration,
			view_count, created_at, imported_at, vote_score, comment_count,
			favorite_count, is_featured, is_nsfw, is_removed, removed_reason,
			submitted_by_user_id, submitted_at
		FROM clips
		WHERE is_removed = false AND created_at > NOW() - INTERVAL '1 hour' * $1
		ORDER BY view_count DESC, created_at DESC
		LIMIT $2
	`

	rows, err := r.pool.Query(ctx, query, hours, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get recent clips: %w", err)
	}
	defer rows.Close()

	var clips []models.Clip
	for rows.Next() {
		var clip models.Clip
		err := rows.Scan(
			&clip.ID, &clip.TwitchClipID, &clip.TwitchClipURL, &clip.EmbedURL,
			&clip.Title, &clip.CreatorName, &clip.CreatorID, &clip.BroadcasterName,
			&clip.BroadcasterID, &clip.GameID, &clip.GameName, &clip.Language,
			&clip.ThumbnailURL, &clip.Duration, &clip.ViewCount, &clip.CreatedAt,
			&clip.ImportedAt, &clip.VoteScore, &clip.CommentCount, &clip.FavoriteCount,
			&clip.IsFeatured, &clip.IsNSFW, &clip.IsRemoved, &clip.RemovedReason,
			&clip.SubmittedByUserID, &clip.SubmittedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan clip: %w", err)
		}
		clips = append(clips, clip)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating clips: %w", err)
	}

	return clips, nil
}

// CountImportedToday counts the number of clips imported today
func (r *ClipRepository) CountImportedToday(ctx context.Context) (int, error) {
	query := `
		SELECT COUNT(*)
		FROM clips
		WHERE imported_at > CURRENT_DATE
	`

	var count int
	err := r.pool.QueryRow(ctx, query).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count imported clips: %w", err)
	}

	return count, nil
}

// GetLastSyncTime gets the last time clips were synced
func (r *ClipRepository) GetLastSyncTime(ctx context.Context) (*time.Time, error) {
	query := `
		SELECT MAX(imported_at)
		FROM clips
	`

	var lastSync *time.Time
	err := r.pool.QueryRow(ctx, query).Scan(&lastSync)
	if err != nil {
		return nil, fmt.Errorf("failed to get last sync time: %w", err)
	}

	return lastSync, nil
}

// ClipFilters represents filters for listing clips
type ClipFilters struct {
	GameID            *string
	BroadcasterID     *string
	Tag               *string
	ExcludeTags       []string // Exclude clips with any of these tag slugs
	Search            *string
	Language          *string // Language code (e.g., en, es, fr)
	Timeframe         *string // hour, day, week, month, year, all
	DateFrom          *string // ISO 8601 date string for custom date range start
	DateTo            *string // ISO 8601 date string for custom date range end
	Sort              string  // hot, new, top, rising, discussed, trending
	Top10kStreamers   bool    // Filter clips to only top 10k streamers
	ShowHidden        bool    // If true, include hidden clips (for owners/admins)
	CreatorID         *string // Filter by creator ID (for creator dashboard)
	SubmittedByUserID *string // Filter by submitted_by_user_id (for user profile submissions)
	UserSubmittedOnly bool    // If true, only show clips with submitted_by_user_id IS NOT NULL
	Cursor            *string // Cursor for cursor-based pagination (base64 encoded)
}

// buildDateFilterClauses adds date range and timeframe filtering clauses
// Note: DateFrom and DateTo are validated before being passed to this function
// to prevent SQL injection. See validateDateFilter in handlers.
func buildDateFilterClauses(filters ClipFilters, whereClauses []string, args []interface{}, argIndex int) ([]string, []interface{}, int) {
	// Add custom date range filter (overrides timeframe if provided)
	if filters.DateFrom != nil && *filters.DateFrom != "" {
		whereClauses = append(whereClauses, fmt.Sprintf("c.created_at >= %s", utils.SQLPlaceholder(argIndex)))
		args = append(args, *filters.DateFrom)
		argIndex++
	}
	if filters.DateTo != nil && *filters.DateTo != "" {
		whereClauses = append(whereClauses, fmt.Sprintf("c.created_at <= %s", utils.SQLPlaceholder(argIndex)))
		args = append(args, *filters.DateTo)
		argIndex++
	}

	// Only apply timeframe if custom date range is not provided
	customDateRangeProvided := (filters.DateFrom != nil && *filters.DateFrom != "") || (filters.DateTo != nil && *filters.DateTo != "")

	if !customDateRangeProvided {
		// Add timeframe filter for top sort
		if filters.Sort == "top" && filters.Timeframe != nil {
			switch *filters.Timeframe {
			case "hour":
				whereClauses = append(whereClauses, "c.created_at > NOW() - INTERVAL '1 hour'")
			case "day":
				whereClauses = append(whereClauses, "c.created_at > NOW() - INTERVAL '1 day'")
			case "week":
				whereClauses = append(whereClauses, "c.created_at > NOW() - INTERVAL '7 days'")
			case "month":
				whereClauses = append(whereClauses, "c.created_at > NOW() - INTERVAL '30 days'")
			case "year":
				whereClauses = append(whereClauses, "c.created_at > NOW() - INTERVAL '365 days'")
			}
		}

		// Add timeframe for rising (recent clips only)
		if filters.Sort == "rising" {
			whereClauses = append(whereClauses, "c.created_at > NOW() - INTERVAL '48 hours'")
		}

		// Add timeframe for discussed (recent clips only, optional)
		if filters.Sort == "discussed" && filters.Timeframe != nil {
			switch *filters.Timeframe {
			case "hour":
				whereClauses = append(whereClauses, "c.created_at > NOW() - INTERVAL '1 hour'")
			case "day":
				whereClauses = append(whereClauses, "c.created_at > NOW() - INTERVAL '1 day'")
			case "week":
				whereClauses = append(whereClauses, "c.created_at > NOW() - INTERVAL '7 days'")
			case "month":
				whereClauses = append(whereClauses, "c.created_at > NOW() - INTERVAL '30 days'")
			case "year":
				whereClauses = append(whereClauses, "c.created_at > NOW() - INTERVAL '365 days'")
			}
		}
	}

	return whereClauses, args, argIndex
}

// ListWithFilters retrieves clips with filters, sorting, and pagination
func (r *ClipRepository) ListWithFilters(ctx context.Context, filters ClipFilters, limit, offset int) ([]models.Clip, int, error) {
	// Enforce pagination limits
	r.helper.EnforcePaginationLimits(&limit, &offset)

	// Build WHERE clause
	whereClauses := []string{"c.is_removed = false"}

	// Filter hidden clips unless ShowHidden is true
	if !filters.ShowHidden {
		whereClauses = append(whereClauses, "c.is_hidden = false")
	}

	// Filter to only user-submitted clips if UserSubmittedOnly is true
	if filters.UserSubmittedOnly {
		whereClauses = append(whereClauses, "c.submitted_by_user_id IS NOT NULL")
	}

	args := []interface{}{}
	argIndex := 1

	if filters.GameID != nil {
		whereClauses = append(whereClauses, fmt.Sprintf("c.game_id = %s", utils.SQLPlaceholder(argIndex)))
		args = append(args, *filters.GameID)
		argIndex++
	}

	if filters.BroadcasterID != nil {
		whereClauses = append(whereClauses, fmt.Sprintf("c.broadcaster_id = %s", utils.SQLPlaceholder(argIndex)))
		args = append(args, *filters.BroadcasterID)
		argIndex++
	}

	if filters.CreatorID != nil {
		whereClauses = append(whereClauses, fmt.Sprintf("c.creator_id = %s", utils.SQLPlaceholder(argIndex)))
		args = append(args, *filters.CreatorID)
		argIndex++
	}

	if filters.SubmittedByUserID != nil {
		whereClauses = append(whereClauses, fmt.Sprintf("c.submitted_by_user_id = %s", utils.SQLPlaceholder(argIndex)))
		args = append(args, *filters.SubmittedByUserID)
		argIndex++
	}

	if filters.Tag != nil {
		whereClauses = append(whereClauses, fmt.Sprintf(`EXISTS (
			SELECT 1 FROM clip_tags ct
			JOIN tags t ON ct.tag_id = t.id
			WHERE ct.clip_id = c.id AND t.slug = %s
		)`, utils.SQLPlaceholder(argIndex)))
		args = append(args, *filters.Tag)
		argIndex++
	}

	// Exclude clips with any of the specified tags
	if len(filters.ExcludeTags) > 0 {
		whereClauses = append(whereClauses, fmt.Sprintf(`NOT EXISTS (
			SELECT 1 FROM clip_tags ct
			JOIN tags t ON ct.tag_id = t.id
			WHERE ct.clip_id = c.id AND t.slug = ANY(%s)
		)`, utils.SQLPlaceholder(argIndex)))
		args = append(args, filters.ExcludeTags)
		argIndex++
	}

	if filters.Search != nil && *filters.Search != "" {
		whereClauses = append(whereClauses, fmt.Sprintf("c.title ILIKE %s", utils.SQLPlaceholder(argIndex)))
		args = append(args, "%"+*filters.Search+"%")
		argIndex++
	}

	if filters.Language != nil && *filters.Language != "" {
		placeholder := utils.SQLPlaceholder(argIndex)
		whereClauses = append(whereClauses, fmt.Sprintf("(c.language = %s OR c.language = split_part(%s, '-', 1))", placeholder, placeholder))
		args = append(args, *filters.Language)
		argIndex++
	}

	// Filter by top 10k streamers if requested
	if filters.Top10kStreamers {
		whereClauses = append(whereClauses, `EXISTS (
			SELECT 1 FROM top_streamers ts
			WHERE ts.broadcaster_id = c.broadcaster_id
		)`)
	}

	// Add date range and timeframe filtering
	whereClauses, args, argIndex = buildDateFilterClauses(filters, whereClauses, args, argIndex)

	// Add cursor-based filtering if cursor is provided
	if filters.Cursor != nil && *filters.Cursor != "" {
		cursor, err := utils.DecodeCursor(*filters.Cursor)
		if err != nil {
			return nil, 0, fmt.Errorf("invalid cursor: %w", err)
		}

		// Validate that cursor sort key matches requested sort
		if cursor.SortKey != filters.Sort {
			return nil, 0, fmt.Errorf("cursor sort key %q does not match requested sort %q", cursor.SortKey, filters.Sort)
		}

		// Add cursor WHERE clause based on sort type
		// For DESC sorts: WHERE (sort_field < cursor_value) OR (sort_field = cursor_value AND id < cursor_id)
		// This ensures stable pagination even with duplicate sort values
		cursorTimestamp := time.Unix(cursor.CreatedAt, 0)

		switch filters.Sort {
		case "trending":
			whereClauses = append(whereClauses, fmt.Sprintf(
				"(COALESCE(c.trending_score, calculate_trending_score(c.view_count, c.vote_score, c.comment_count, c.favorite_count, c.created_at)) < %s OR (COALESCE(c.trending_score, calculate_trending_score(c.view_count, c.vote_score, c.comment_count, c.favorite_count, c.created_at)) = %s AND (c.created_at < %s OR (c.created_at = %s AND c.id < %s))))",
				utils.SQLPlaceholder(argIndex), utils.SQLPlaceholder(argIndex+1), utils.SQLPlaceholder(argIndex+2), utils.SQLPlaceholder(argIndex+3), utils.SQLPlaceholder(argIndex+4)))
			args = append(args, cursor.SortValue, cursor.SortValue, cursorTimestamp, cursorTimestamp, cursor.ClipID)
			argIndex += 5
		case "popular":
			whereClauses = append(whereClauses, fmt.Sprintf(
				"(COALESCE(c.popularity_index, c.engagement_count, (c.view_count + c.vote_score * 2 + c.comment_count * 3 + c.favorite_count * 2)) < %s OR (COALESCE(c.popularity_index, c.engagement_count, (c.view_count + c.vote_score * 2 + c.comment_count * 3 + c.favorite_count * 2)) = %s AND (c.created_at < %s OR (c.created_at = %s AND c.id < %s))))",
				utils.SQLPlaceholder(argIndex), utils.SQLPlaceholder(argIndex+1), utils.SQLPlaceholder(argIndex+2), utils.SQLPlaceholder(argIndex+3), utils.SQLPlaceholder(argIndex+4)))
			args = append(args, cursor.SortValue, cursor.SortValue, cursorTimestamp, cursorTimestamp, cursor.ClipID)
			argIndex += 5
		case "new":
			whereClauses = append(whereClauses, fmt.Sprintf(
				"(c.created_at < %s OR (c.created_at = %s AND c.id < %s))",
				utils.SQLPlaceholder(argIndex), utils.SQLPlaceholder(argIndex+1), utils.SQLPlaceholder(argIndex+2)))
			args = append(args, cursorTimestamp, cursorTimestamp, cursor.ClipID)
			argIndex += 3
		case "top":
			whereClauses = append(whereClauses, fmt.Sprintf(
				"(c.vote_score < %s OR (c.vote_score = %s AND (c.created_at < %s OR (c.created_at = %s AND c.id < %s))))",
				utils.SQLPlaceholder(argIndex), utils.SQLPlaceholder(argIndex+1), utils.SQLPlaceholder(argIndex+2), utils.SQLPlaceholder(argIndex+3), utils.SQLPlaceholder(argIndex+4)))
			args = append(args, cursor.SortValue, cursor.SortValue, cursorTimestamp, cursorTimestamp, cursor.ClipID)
			argIndex += 5
		case "discussed":
			whereClauses = append(whereClauses, fmt.Sprintf(
				"(c.comment_count < %s OR (c.comment_count = %s AND (c.created_at < %s OR (c.created_at = %s AND c.id < %s))))",
				utils.SQLPlaceholder(argIndex), utils.SQLPlaceholder(argIndex+1), utils.SQLPlaceholder(argIndex+2), utils.SQLPlaceholder(argIndex+3), utils.SQLPlaceholder(argIndex+4)))
			args = append(args, cursor.SortValue, cursor.SortValue, cursorTimestamp, cursorTimestamp, cursor.ClipID)
			argIndex += 5
		case "hot", "rising":
			// For hot and rising, we use created_at as the cursor since the score is dynamically calculated
			// Note: This means pagination is based on created_at, not the hot/rising score
			// This is a known limitation but avoids storing calculated scores
			whereClauses = append(whereClauses, fmt.Sprintf(
				"(c.created_at < %s OR (c.created_at = %s AND c.id < %s))",
				utils.SQLPlaceholder(argIndex), utils.SQLPlaceholder(argIndex+1), utils.SQLPlaceholder(argIndex+2)))
			args = append(args, cursorTimestamp, cursorTimestamp, cursor.ClipID)
			argIndex += 3
		default:
			// Default to hot score behavior
			whereClauses = append(whereClauses, fmt.Sprintf(
				"(c.created_at < %s OR (c.created_at = %s AND c.id < %s))",
				utils.SQLPlaceholder(argIndex), utils.SQLPlaceholder(argIndex+1), utils.SQLPlaceholder(argIndex+2)))
			args = append(args, cursorTimestamp, cursorTimestamp, cursor.ClipID)
			argIndex += 3
		}
	}

	whereClause := "WHERE " + whereClauses[0]
	for i := 1; i < len(whereClauses); i++ {
		whereClause += " AND " + whereClauses[i]
	}

	// Build ORDER BY clause
	var orderBy string
	switch filters.Sort {
	case "hot":
		orderBy = "ORDER BY calculate_hot_score(c.vote_score, c.created_at) DESC, c.created_at DESC, c.id DESC"
	case "new":
		orderBy = "ORDER BY COALESCE(c.submitted_at, c.created_at) DESC, c.id DESC"
	case "top":
		orderBy = "ORDER BY c.vote_score DESC, c.created_at DESC, c.id DESC"
	case "trending":
		// Trending: uses pre-calculated trending_score (engagement/age) with fallback to real-time calculation
		orderBy = "ORDER BY COALESCE(c.trending_score, calculate_trending_score(c.view_count, c.vote_score, c.comment_count, c.favorite_count, c.created_at)) DESC, c.created_at DESC, c.id DESC"
	case "popular":
		// Popular: uses pre-calculated popularity_index (total engagement) with fallback
		orderBy = "ORDER BY COALESCE(c.popularity_index, c.engagement_count, (c.view_count + c.vote_score * 2 + c.comment_count * 3 + c.favorite_count * 2)) DESC, c.created_at DESC, c.id DESC"
	case "rising":
		// Rising: recent clips with high velocity (view_count + vote_score combined with recency)
		orderBy = "ORDER BY (c.vote_score + (c.view_count / 100)) * (1 + 1.0 / (EXTRACT(EPOCH FROM (NOW() - c.created_at)) / 3600.0 + 2)) DESC, c.created_at DESC, c.id DESC"
	case "discussed":
		// Discussed: clips with most comments, breaking ties by creation date
		orderBy = "ORDER BY c.comment_count DESC, c.created_at DESC, c.id DESC"
	default:
		orderBy = "ORDER BY calculate_hot_score(c.vote_score, c.created_at) DESC, c.created_at DESC, c.id DESC"
	}

	// Count query
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM clips c %s", whereClause)
	var total int
	err := r.pool.QueryRow(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count clips: %w", err)
	}

	// Main query
	args = append(args, limit, offset)
	query := fmt.Sprintf(`
		SELECT
			c.id, c.twitch_clip_id, c.twitch_clip_url, c.embed_url, c.title,
			c.creator_name, c.creator_id, c.broadcaster_name, c.broadcaster_id,
			c.game_id, c.game_name, c.language, c.thumbnail_url, c.duration,
			c.view_count, c.created_at, c.imported_at, c.vote_score, c.comment_count,
			c.favorite_count, c.is_featured, c.is_nsfw, c.is_removed, c.removed_reason, c.is_hidden,
			c.submitted_by_user_id, c.submitted_at,
			c.trending_score, c.hot_score, c.popularity_index, c.engagement_count
		FROM clips c
		%s
		%s
		LIMIT %s OFFSET %s
	`, whereClause, orderBy, utils.SQLPlaceholder(argIndex), utils.SQLPlaceholder(argIndex+1))

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list clips: %w", err)
	}
	defer rows.Close()

	var clips []models.Clip
	for rows.Next() {
		var clip models.Clip
		err := rows.Scan(
			&clip.ID, &clip.TwitchClipID, &clip.TwitchClipURL, &clip.EmbedURL,
			&clip.Title, &clip.CreatorName, &clip.CreatorID, &clip.BroadcasterName,
			&clip.BroadcasterID, &clip.GameID, &clip.GameName, &clip.Language,
			&clip.ThumbnailURL, &clip.Duration, &clip.ViewCount, &clip.CreatedAt,
			&clip.ImportedAt, &clip.VoteScore, &clip.CommentCount, &clip.FavoriteCount,
			&clip.IsFeatured, &clip.IsNSFW, &clip.IsRemoved, &clip.RemovedReason, &clip.IsHidden,
			&clip.SubmittedByUserID, &clip.SubmittedAt,
			&clip.TrendingScore, &clip.HotScore, &clip.PopularityIndex, &clip.EngagementCount,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan clip: %w", err)
		}
		clips = append(clips, clip)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("error iterating clips: %w", err)
	}

	return clips, total, nil
}

// ListScrapedClipsWithFilters retrieves only scraped clips (submitted_by_user_id IS NULL) with filters, sorting, and pagination
func (r *ClipRepository) ListScrapedClipsWithFilters(ctx context.Context, filters ClipFilters, limit, offset int) ([]models.Clip, int, error) {
	// Build WHERE clause - start with scraped clips filter
	whereClauses := []string{"c.is_removed = false", "c.submitted_by_user_id IS NULL"}

	// Filter hidden clips unless ShowHidden is true
	if !filters.ShowHidden {
		whereClauses = append(whereClauses, "c.is_hidden = false")
	}

	args := []interface{}{}
	argIndex := 1

	if filters.GameID != nil {
		whereClauses = append(whereClauses, fmt.Sprintf("c.game_id = %s", utils.SQLPlaceholder(argIndex)))
		args = append(args, *filters.GameID)
		argIndex++
	}

	if filters.BroadcasterID != nil {
		whereClauses = append(whereClauses, fmt.Sprintf("c.broadcaster_id = %s", utils.SQLPlaceholder(argIndex)))
		args = append(args, *filters.BroadcasterID)
		argIndex++
	}

	if filters.CreatorID != nil {
		whereClauses = append(whereClauses, fmt.Sprintf("c.creator_id = %s", utils.SQLPlaceholder(argIndex)))
		args = append(args, *filters.CreatorID)
		argIndex++
	}

	// Note: SubmittedByUserID filter is intentionally not applied here since this method
	// retrieves only scraped clips (submitted_by_user_id IS NULL). Use ListWithFilters instead.

	if filters.Tag != nil {
		whereClauses = append(whereClauses, fmt.Sprintf(`EXISTS (
			SELECT 1 FROM clip_tags ct
			JOIN tags t ON ct.tag_id = t.id
			WHERE ct.clip_id = c.id AND t.slug = %s
		)`, utils.SQLPlaceholder(argIndex)))
		args = append(args, *filters.Tag)
		argIndex++
	}

	// Exclude clips with any of the specified tags
	if len(filters.ExcludeTags) > 0 {
		whereClauses = append(whereClauses, fmt.Sprintf(`NOT EXISTS (
			SELECT 1 FROM clip_tags ct
			JOIN tags t ON ct.tag_id = t.id
			WHERE ct.clip_id = c.id AND t.slug = ANY(%s)
		)`, utils.SQLPlaceholder(argIndex)))
		args = append(args, filters.ExcludeTags)
		argIndex++
	}

	if filters.Search != nil && *filters.Search != "" {
		whereClauses = append(whereClauses, fmt.Sprintf("c.title ILIKE %s", utils.SQLPlaceholder(argIndex)))
		args = append(args, "%"+*filters.Search+"%")
		argIndex++
	}

	if filters.Language != nil && *filters.Language != "" {
		placeholder := utils.SQLPlaceholder(argIndex)
		whereClauses = append(whereClauses, fmt.Sprintf("(c.language = %s OR c.language = split_part(%s, '-', 1) OR c.language IS NULL OR c.language = '')", placeholder, placeholder))
		args = append(args, *filters.Language)
		argIndex++
	}

	// Filter by top 10k streamers if requested
	if filters.Top10kStreamers {
		whereClauses = append(whereClauses, `EXISTS (
			SELECT 1 FROM top_streamers ts
			WHERE ts.broadcaster_id = c.broadcaster_id
		)`)
	}

	// Add date range and timeframe filtering
	whereClauses, args, argIndex = buildDateFilterClauses(filters, whereClauses, args, argIndex)

	whereClause := "WHERE " + whereClauses[0]
	for i := 1; i < len(whereClauses); i++ {
		whereClause += " AND " + whereClauses[i]
	}

	// Build ORDER BY clause
	var orderBy string
	switch filters.Sort {
	case "hot":
		orderBy = "ORDER BY calculate_hot_score(c.vote_score, c.created_at) DESC"
	case "new":
		orderBy = "ORDER BY c.created_at DESC"
	case "top":
		orderBy = "ORDER BY c.vote_score DESC, c.created_at DESC"
	case "views":
		orderBy = "ORDER BY c.view_count DESC, c.created_at DESC"
	case "trending":
		// Similar to rising but with higher view count weight
		orderBy = "ORDER BY (c.view_count / 10 + c.vote_score) * (1 + 1.0 / (EXTRACT(EPOCH FROM (NOW() - c.created_at)) / 3600.0 + 2)) DESC"
	case "rising":
		// Rising: recent clips with high velocity (view_count + vote_score combined with recency)
		orderBy = "ORDER BY (c.vote_score + (c.view_count / 100)) * (1 + 1.0 / (EXTRACT(EPOCH FROM (NOW() - c.created_at)) / 3600.0 + 2)) DESC"
	case "discussed":
		// Discussed: clips with most comments, breaking ties by creation date
		orderBy = "ORDER BY c.comment_count DESC, c.created_at DESC"
	default:
		orderBy = "ORDER BY c.created_at DESC"
	}

	// Count query
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM clips c %s", whereClause)
	var total int
	err := r.pool.QueryRow(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count scraped clips: %w", err)
	}

	// Main query
	args = append(args, limit, offset)
	query := fmt.Sprintf(`
		SELECT
			c.id, c.twitch_clip_id, c.twitch_clip_url, c.embed_url, c.title,
			c.creator_name, c.creator_id, c.broadcaster_name, c.broadcaster_id,
			c.game_id, c.game_name, c.language, c.thumbnail_url, c.duration,
			c.view_count, c.created_at, c.imported_at, c.vote_score, c.comment_count,
			c.favorite_count, c.is_featured, c.is_nsfw, c.is_removed, c.removed_reason, c.is_hidden,
			c.submitted_by_user_id, c.submitted_at
		FROM clips c
		%s
		%s
		LIMIT %s OFFSET %s
	`, whereClause, orderBy, utils.SQLPlaceholder(argIndex), utils.SQLPlaceholder(argIndex+1))

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query scraped clips: %w", err)
	}
	defer rows.Close()

	var clips []models.Clip
	for rows.Next() {
		var clip models.Clip
		err := rows.Scan(
			&clip.ID, &clip.TwitchClipID, &clip.TwitchClipURL, &clip.EmbedURL,
			&clip.Title, &clip.CreatorName, &clip.CreatorID, &clip.BroadcasterName,
			&clip.BroadcasterID, &clip.GameID, &clip.GameName, &clip.Language,
			&clip.ThumbnailURL, &clip.Duration, &clip.ViewCount, &clip.CreatedAt,
			&clip.ImportedAt, &clip.VoteScore, &clip.CommentCount, &clip.FavoriteCount,
			&clip.IsFeatured, &clip.IsNSFW, &clip.IsRemoved, &clip.RemovedReason, &clip.IsHidden,
			&clip.SubmittedByUserID, &clip.SubmittedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan scraped clip: %w", err)
		}
		clips = append(clips, clip)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("error iterating scraped clips: %w", err)
	}

	return clips, total, nil
}

// IncrementViewCount atomically increments the view count for a clip and returns the new count
func (r *ClipRepository) IncrementViewCount(ctx context.Context, clipID uuid.UUID) (int64, error) {
	query := `
		UPDATE clips
		SET view_count = view_count + 1
		WHERE id = $1
		RETURNING view_count
	`

	var newViewCount int64
	err := r.pool.QueryRow(ctx, query, clipID).Scan(&newViewCount)
	if err != nil {
		return 0, fmt.Errorf("failed to increment view count: %w", err)
	}

	return newViewCount, nil
}

// UpdateVoteScore increments the vote_score by the provided delta and returns the new score
func (r *ClipRepository) UpdateVoteScore(ctx context.Context, clipID uuid.UUID, delta int64) (int64, error) {
	query := `
		UPDATE clips
		SET vote_score = vote_score + $2
		WHERE id = $1
		RETURNING vote_score
	`

	var newScore int64
	err := r.pool.QueryRow(ctx, query, clipID, delta).Scan(&newScore)
	if err != nil {
		return 0, fmt.Errorf("failed to update vote score: %w", err)
	}

	return newScore, nil
}

// Update updates a clip (for admin operations)
func (r *ClipRepository) Update(ctx context.Context, clipID uuid.UUID, updates map[string]interface{}) error {
	if len(updates) == 0 {
		return nil
	}

	// Build dynamic update query
	setClauses := []string{}
	args := []interface{}{}
	argIndex := 1

	for field, value := range updates {
		setClauses = append(setClauses, fmt.Sprintf("%s = %s", field, utils.SQLPlaceholder(argIndex)))
		args = append(args, value)
		argIndex++
	}

	args = append(args, clipID)
	setClause := setClauses[0]
	if len(setClauses) > 1 {
		for i := 1; i < len(setClauses); i++ {
			setClause += ", " + setClauses[i]
		}
	}
	query := fmt.Sprintf("UPDATE clips SET %s WHERE id = %s", setClause, utils.SQLPlaceholder(argIndex))

	_, err := r.pool.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to update clip: %w", err)
	}

	return nil
}

// SoftDelete marks a clip as removed
func (r *ClipRepository) SoftDelete(ctx context.Context, clipID uuid.UUID, reason string) error {
	query := `
		UPDATE clips
		SET is_removed = true, removed_reason = $2
		WHERE id = $1
	`

	_, err := r.pool.Exec(ctx, query, clipID, reason)
	if err != nil {
		return fmt.Errorf("failed to soft delete clip: %w", err)
	}

	return nil
}

// Delete removes a clip record permanently. Accepts either uuid.UUID or string identifiers.
func (r *ClipRepository) Delete(ctx context.Context, clipID interface{}) error {
	var id uuid.UUID

	switch v := clipID.(type) {
	case uuid.UUID:
		id = v
	case string:
		parsed, err := uuid.Parse(v)
		if err != nil {
			return fmt.Errorf("invalid clip id: %w", err)
		}
		id = parsed
	default:
		return fmt.Errorf("unsupported clip id type %T", v)
	}

	if _, err := r.pool.Exec(ctx, "DELETE FROM clips WHERE id = $1", id); err != nil {
		return fmt.Errorf("failed to delete clip: %w", err)
	}

	return nil
}

// GetRelated finds related clips based on game, broadcaster, and tags
func (r *ClipRepository) GetRelated(ctx context.Context, clipID uuid.UUID, limit int) ([]models.Clip, error) {
	query := `
		WITH current_clip AS (
			SELECT game_id, broadcaster_id
			FROM clips
			WHERE id = $1
		),
		current_tags AS (
			SELECT tag_id
			FROM clip_tags
			WHERE clip_id = $1
		)
		SELECT
			c.id, c.twitch_clip_id, c.twitch_clip_url, c.embed_url, c.title,
			c.creator_name, c.creator_id, c.broadcaster_name, c.broadcaster_id,
			c.game_id, c.game_name, c.language, c.thumbnail_url, c.duration,
			c.view_count, c.created_at, c.imported_at, c.vote_score, c.comment_count,
			c.favorite_count, c.is_featured, c.is_nsfw, c.is_removed, c.removed_reason,
			(
				CASE WHEN c.game_id = (SELECT game_id FROM current_clip) THEN 3 ELSE 0 END +
				CASE WHEN c.broadcaster_id = (SELECT broadcaster_id FROM current_clip) THEN 2 ELSE 0 END +
				COALESCE((
					SELECT COUNT(*)
					FROM clip_tags ct
					WHERE ct.clip_id = c.id AND ct.tag_id IN (SELECT tag_id FROM current_tags)
				), 0)
			) as relevance_score
		FROM clips c
		WHERE c.id != $1 AND c.is_removed = false
		ORDER BY relevance_score DESC, c.vote_score DESC
		LIMIT $2
	`

	rows, err := r.pool.Query(ctx, query, clipID, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get related clips: %w", err)
	}
	defer rows.Close()

	var clips []models.Clip
	for rows.Next() {
		var clip models.Clip
		var relevanceScore int
		err := rows.Scan(
			&clip.ID, &clip.TwitchClipID, &clip.TwitchClipURL, &clip.EmbedURL,
			&clip.Title, &clip.CreatorName, &clip.CreatorID, &clip.BroadcasterName,
			&clip.BroadcasterID, &clip.GameID, &clip.GameName, &clip.Language,
			&clip.ThumbnailURL, &clip.Duration, &clip.ViewCount, &clip.CreatedAt,
			&clip.ImportedAt, &clip.VoteScore, &clip.CommentCount, &clip.FavoriteCount,
			&clip.IsFeatured, &clip.IsNSFW, &clip.IsRemoved, &clip.RemovedReason,
			&relevanceScore,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan related clip: %w", err)
		}
		clips = append(clips, clip)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating related clips: %w", err)
	}

	return clips, nil
}

// RemoveClip marks a clip as removed with a reason
func (r *ClipRepository) RemoveClip(ctx context.Context, clipID uuid.UUID, reason *string) error {
	query := `
		UPDATE clips
		SET is_removed = true, removed_reason = $2
		WHERE id = $1
	`

	_, err := r.pool.Exec(ctx, query, clipID, reason)
	return err
}

// RefreshHotScores refreshes the materialized view for hot clips
// This should be called periodically to update hot scores for discovery lists
func (r *ClipRepository) RefreshHotScores(ctx context.Context) error {
	// Note: HotClipsMaterializedView is a compile-time constant, not user input,
	// so this is safe from SQL injection. PostgreSQL does not support parameterized
	// table/view names in DDL statements like REFRESH MATERIALIZED VIEW.
	query := fmt.Sprintf("REFRESH MATERIALIZED VIEW CONCURRENTLY %s", HotClipsMaterializedView)

	_, err := r.pool.Exec(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to refresh hot scores: %w", err)
	}

	return nil
}

// UpsertTopStreamer inserts or updates a top streamer record
func (r *ClipRepository) UpsertTopStreamer(ctx context.Context, broadcasterID, broadcasterName string, rank int, followerCount, viewCount int64) error {
	query := `
		INSERT INTO top_streamers (broadcaster_id, broadcaster_name, rank, follower_count, view_count, last_updated)
		VALUES ($1, $2, $3, $4, $5, NOW())
		ON CONFLICT (broadcaster_id)
		DO UPDATE SET
			broadcaster_name = EXCLUDED.broadcaster_name,
			rank = EXCLUDED.rank,
			follower_count = EXCLUDED.follower_count,
			view_count = EXCLUDED.view_count,
			last_updated = NOW()
	`

	_, err := r.pool.Exec(ctx, query, broadcasterID, broadcasterName, rank, followerCount, viewCount)
	if err != nil {
		return fmt.Errorf("failed to upsert top streamer: %w", err)
	}

	return nil
}

// ClearTopStreamers clears all top streamer records (useful before bulk insert)
func (r *ClipRepository) ClearTopStreamers(ctx context.Context) error {
	query := `TRUNCATE TABLE top_streamers`

	_, err := r.pool.Exec(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to clear top streamers: %w", err)
	}

	return nil
}

// GetTopStreamersCount returns the count of top streamers in the database
func (r *ClipRepository) GetTopStreamersCount(ctx context.Context) (int, error) {
	query := `SELECT COUNT(*) FROM top_streamers`

	var count int
	err := r.pool.QueryRow(ctx, query).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count top streamers: %w", err)
	}

	return count, nil
}

// IsTopStreamer checks if a broadcaster is in the top streamers list
func (r *ClipRepository) IsTopStreamer(ctx context.Context, broadcasterID string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM top_streamers WHERE broadcaster_id = $1)`

	var exists bool
	err := r.pool.QueryRow(ctx, query, broadcasterID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check top streamer status: %w", err)
	}

	return exists, nil
}

// GetByIDs retrieves clips by their IDs, maintaining the order of the provided IDs
func (r *ClipRepository) GetByIDs(ctx context.Context, clipIDs []uuid.UUID) ([]models.Clip, error) {
	if len(clipIDs) == 0 {
		return []models.Clip{}, nil
	}

	query := `
		SELECT
			id, twitch_clip_id, twitch_clip_url, embed_url, title,
			creator_name, creator_id, broadcaster_name, broadcaster_id,
			game_id, game_name, language, thumbnail_url, duration,
			view_count, created_at, imported_at, vote_score, comment_count,
			favorite_count, is_featured, is_nsfw, is_removed, removed_reason,
			submitted_by_user_id, submitted_at
		FROM clips
		WHERE id = ANY($1)
	`

	rows, err := r.pool.Query(ctx, query, clipIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to query clips by IDs: %w", err)
	}
	defer rows.Close()

	// Create a map to store clips by ID
	clipMap := make(map[uuid.UUID]models.Clip)
	for rows.Next() {
		var clip models.Clip
		if err := rows.Scan(
			&clip.ID, &clip.TwitchClipID, &clip.TwitchClipURL, &clip.EmbedURL,
			&clip.Title, &clip.CreatorName, &clip.CreatorID, &clip.BroadcasterName,
			&clip.BroadcasterID, &clip.GameID, &clip.GameName, &clip.Language,
			&clip.ThumbnailURL, &clip.Duration, &clip.ViewCount, &clip.CreatedAt,
			&clip.ImportedAt, &clip.VoteScore, &clip.CommentCount, &clip.FavoriteCount,
			&clip.IsFeatured, &clip.IsNSFW, &clip.IsRemoved, &clip.RemovedReason,
			&clip.SubmittedByUserID, &clip.SubmittedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan clip: %w", err)
		}
		clipMap[clip.ID] = clip
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating clips: %w", err)
	}

	// Maintain the order of the provided IDs
	clips := make([]models.Clip, 0, len(clipIDs))
	for _, id := range clipIDs {
		if clip, ok := clipMap[id]; ok {
			clips = append(clips, clip)
		}
	}

	return clips, nil
}

// ListForSitemap retrieves all non-removed clips with minimal info for sitemap generation
// Limits to 10,000 clips to keep sitemap size manageable (Google's recommended limit is 50,000 URLs).
// For sites with more clips, consider implementing a sitemap index with multiple sitemap files.
// Returns clips ordered by creation date (newest first).
func (r *ClipRepository) ListForSitemap(ctx context.Context) ([]models.Clip, error) {
	query := `
		SELECT id, created_at
		FROM clips
		WHERE is_removed = false
		ORDER BY created_at DESC
		LIMIT 10000
	`

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to list clips for sitemap: %w", err)
	}
	defer rows.Close()

	var clips []models.Clip
	for rows.Next() {
		var clip models.Clip
		if err := rows.Scan(&clip.ID, &clip.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan clip: %w", err)
		}
		clips = append(clips, clip)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating clips: %w", err)
	}

	return clips, nil
}

// ListClipsByBroadcaster retrieves clips for a specific broadcaster with pagination and sorting
func (r *ClipRepository) ListClipsByBroadcaster(ctx context.Context, broadcasterID, sort string, limit, offset int) ([]models.Clip, int, error) {
	// Get total count
	countQuery := `
		SELECT COUNT(*)
		FROM clips
		WHERE broadcaster_id = $1 AND is_removed = false AND is_hidden = false AND submitted_by_user_id IS NOT NULL
	`
	var total int
	if err := r.pool.QueryRow(ctx, countQuery, broadcasterID).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("failed to count broadcaster clips: %w", err)
	}

	// Build order by clause based on sort parameter
	orderBy := "created_at DESC" // default: recent
	switch sort {
	case "popular":
		orderBy = "vote_score DESC, created_at DESC"
	case "trending":
		// Sort by popularity with recency tiebreaker (not true trending/hot score)
		orderBy = "vote_score DESC, view_count DESC, created_at DESC"
	}

	// Get clips
	query := fmt.Sprintf(`
		SELECT
			id, twitch_clip_id, twitch_clip_url, embed_url, title,
			creator_name, creator_id, broadcaster_name, broadcaster_id,
			game_id, game_name, language, thumbnail_url, duration,
			view_count, created_at, imported_at, vote_score, comment_count,
			favorite_count, is_featured, is_nsfw, is_removed, removed_reason, is_hidden,
			submitted_by_user_id, submitted_at
		FROM clips
		WHERE broadcaster_id = $1 AND is_removed = false AND is_hidden = false AND submitted_by_user_id IS NOT NULL
		ORDER BY %s
		LIMIT $2 OFFSET $3
	`, orderBy)

	rows, err := r.pool.Query(ctx, query, broadcasterID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list broadcaster clips: %w", err)
	}
	defer rows.Close()

	var clips []models.Clip
	for rows.Next() {
		var clip models.Clip
		if err := rows.Scan(
			&clip.ID, &clip.TwitchClipID, &clip.TwitchClipURL, &clip.EmbedURL,
			&clip.Title, &clip.CreatorName, &clip.CreatorID, &clip.BroadcasterName,
			&clip.BroadcasterID, &clip.GameID, &clip.GameName, &clip.Language,
			&clip.ThumbnailURL, &clip.Duration, &clip.ViewCount, &clip.CreatedAt,
			&clip.ImportedAt, &clip.VoteScore, &clip.CommentCount, &clip.FavoriteCount,
			&clip.IsFeatured, &clip.IsNSFW, &clip.IsRemoved, &clip.RemovedReason, &clip.IsHidden,
			&clip.SubmittedByUserID, &clip.SubmittedAt,
		); err != nil {
			return nil, 0, fmt.Errorf("failed to scan clip: %w", err)
		}
		clips = append(clips, clip)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("error iterating clips: %w", err)
	}

	return clips, total, nil
}

// UpdateMetadata updates the title of a clip
func (r *ClipRepository) UpdateMetadata(ctx context.Context, clipID uuid.UUID, title *string) error {
	// Whitelist of allowed fields for metadata update
	allowedFields := map[string]struct{}{
		"title": {},
	}

	updates := make(map[string]interface{})
	if title != nil {
		updates["title"] = *title
	}

	// Filter updates to only include allowed fields
	filteredUpdates := make(map[string]interface{})
	for field, value := range updates {
		if _, ok := allowedFields[field]; ok {
			filteredUpdates[field] = value
		}
	}

	if len(filteredUpdates) == 0 {
		return nil
	}

	return r.Update(ctx, clipID, filteredUpdates)
}

// UpdateVisibility updates the visibility status of a clip
func (r *ClipRepository) UpdateVisibility(ctx context.Context, clipID uuid.UUID, isHidden bool) error {
	query := `
UPDATE clips
SET is_hidden = $2
WHERE id = $1
`

	_, err := r.pool.Exec(ctx, query, clipID, isHidden)
	if err != nil {
		return fmt.Errorf("failed to update clip visibility: %w", err)
	}

	return nil
}

// GetFollowingFeedClips retrieves clips from users and broadcasters that the user follows
func (r *ClipRepository) GetFollowingFeedClips(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*models.ClipWithSubmitter, int, error) {
	query := `
WITH followed_users AS (
SELECT following_id FROM user_follows WHERE follower_id = $1
),
followed_broadcasters AS (
SELECT broadcaster_id FROM broadcaster_follows WHERE user_id = $1
)
,
blocked_users AS (
SELECT blocked_user_id FROM user_blocks WHERE user_id = $1
)
SELECT
c.id, c.twitch_clip_id, c.twitch_clip_url, c.embed_url,
c.title, c.creator_name, c.creator_id, c.broadcaster_name, c.broadcaster_id,
c.game_id, c.game_name, c.language, c.thumbnail_url, c.duration,
c.view_count, c.created_at, c.imported_at, c.vote_score, c.comment_count,
c.favorite_count, c.is_featured, c.is_nsfw, c.is_removed, c.removed_reason,
c.is_hidden, c.submitted_by_user_id, c.submitted_at,
u.id as submitter_id, u.username as submitter_username,
u.display_name as submitter_display_name, u.avatar_url as submitter_avatar_url
FROM clips c
LEFT JOIN users u ON c.submitted_by_user_id = u.id
WHERE c.is_removed = false
AND c.is_hidden = false
AND (
c.submitted_by_user_id IN (SELECT following_id FROM followed_users)
OR c.broadcaster_id IN (SELECT broadcaster_id FROM followed_broadcasters)
)
AND c.submitted_by_user_id NOT IN (SELECT blocked_user_id FROM blocked_users)
ORDER BY c.created_at DESC
LIMIT $2 OFFSET $3
`

	rows, err := r.pool.Query(ctx, query, userID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var clips []*models.ClipWithSubmitter
	for rows.Next() {
		var clip models.ClipWithSubmitter
		var submitterID *uuid.UUID
		var submitterUsername, submitterDisplayName *string
		var submitterAvatarURL *string

		err := rows.Scan(
			&clip.ID, &clip.TwitchClipID, &clip.TwitchClipURL, &clip.EmbedURL,
			&clip.Title, &clip.CreatorName, &clip.CreatorID, &clip.BroadcasterName, &clip.BroadcasterID,
			&clip.GameID, &clip.GameName, &clip.Language, &clip.ThumbnailURL, &clip.Duration,
			&clip.ViewCount, &clip.CreatedAt, &clip.ImportedAt, &clip.VoteScore, &clip.CommentCount,
			&clip.FavoriteCount, &clip.IsFeatured, &clip.IsNSFW, &clip.IsRemoved, &clip.RemovedReason,
			&clip.IsHidden, &clip.SubmittedByUserID, &clip.SubmittedAt,
			&submitterID, &submitterUsername, &submitterDisplayName, &submitterAvatarURL,
		)
		if err != nil {
			return nil, 0, err
		}

		if submitterID != nil {
			clip.SubmittedBy = &models.ClipSubmitterInfo{
				ID:          *submitterID,
				Username:    *submitterUsername,
				DisplayName: *submitterDisplayName,
				AvatarURL:   submitterAvatarURL,
			}
		}

		clips = append(clips, &clip)
	}

	if err = rows.Err(); err != nil {
		return nil, 0, err
	}

	// Get total count
	countQuery := `
WITH followed_users AS (
SELECT following_id FROM user_follows WHERE follower_id = $1
),
followed_broadcasters AS (
SELECT broadcaster_id FROM broadcaster_follows WHERE user_id = $1
),
blocked_users AS (
SELECT blocked_user_id FROM user_blocks WHERE user_id = $1
)
SELECT COUNT(*)
FROM clips c
WHERE c.is_removed = false
AND c.is_hidden = false
AND (
c.submitted_by_user_id IN (SELECT following_id FROM followed_users)
OR c.broadcaster_id IN (SELECT broadcaster_id FROM followed_broadcasters)
)
AND c.submitted_by_user_id NOT IN (SELECT blocked_user_id FROM blocked_users)
`

	var total int
	err = r.pool.QueryRow(ctx, countQuery, userID).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	return clips, total, nil
}

// UpdateTrendingScores updates trending_score, hot_score, popularity_index, and engagement_count for all clips
// This should be called periodically (e.g., hourly) by a scheduler job
func (r *ClipRepository) UpdateTrendingScores(ctx context.Context) (int64, error) {
	query := `
UPDATE clips
SET
engagement_count = view_count + (vote_score * 2) + (comment_count * 3) + (favorite_count * 2),
trending_score = calculate_trending_score(view_count, vote_score, comment_count, favorite_count, created_at),
hot_score = trending_score,
popularity_index = view_count + (vote_score * 2) + (comment_count * 3) + (favorite_count * 2)
WHERE is_removed = false AND is_hidden = false
`

	result, err := r.pool.Exec(ctx, query)
	if err != nil {
		return 0, fmt.Errorf("failed to update trending scores: %w", err)
	}

	return result.RowsAffected(), nil
}

// UpdateTrendingScoresForTimeWindow updates trending scores for clips within a specific time window
// This can be used to update only recent clips for better performance
func (r *ClipRepository) UpdateTrendingScoresForTimeWindow(ctx context.Context, hours int) (int64, error) {
	query := `
UPDATE clips
SET
engagement_count = view_count + (vote_score * 2) + (comment_count * 3) + (favorite_count * 2),
trending_score = calculate_trending_score(view_count, vote_score, comment_count, favorite_count, created_at),
hot_score = trending_score,
popularity_index = view_count + (vote_score * 2) + (comment_count * 3) + (favorite_count * 2)
WHERE is_removed = false
AND is_hidden = false
AND created_at > NOW() - INTERVAL '1 hour' * $1
`

	result, err := r.pool.Exec(ctx, query, hours)
	if err != nil {
		return 0, fmt.Errorf("failed to update trending scores for time window: %w", err)
	}

	return result.RowsAffected(), nil
}

// GetClipsByIDs retrieves multiple clips by their IDs
func (r *ClipRepository) GetClipsByIDs(ctx context.Context, clipIDs []uuid.UUID) ([]models.Clip, error) {
	if len(clipIDs) == 0 {
		return []models.Clip{}, nil
	}

	query := `
SELECT
id, twitch_clip_id, twitch_clip_url, embed_url, title,
creator_name, creator_id, broadcaster_name, broadcaster_id,
game_id, game_name, language, thumbnail_url, duration,
view_count, created_at, imported_at, vote_score, comment_count,
favorite_count, is_featured, is_nsfw, is_removed, removed_reason,
is_hidden, submitted_by_user_id, submitted_at
FROM clips
WHERE id = ANY($1)
AND is_removed = false
`

	rows, err := r.pool.Query(ctx, query, clipIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to query clips: %w", err)
	}
	defer rows.Close()

	var clips []models.Clip
	for rows.Next() {
		var clip models.Clip
		err := rows.Scan(
			&clip.ID, &clip.TwitchClipID, &clip.TwitchClipURL, &clip.EmbedURL, &clip.Title,
			&clip.CreatorName, &clip.CreatorID, &clip.BroadcasterName, &clip.BroadcasterID,
			&clip.GameID, &clip.GameName, &clip.Language, &clip.ThumbnailURL, &clip.Duration,
			&clip.ViewCount, &clip.CreatedAt, &clip.ImportedAt, &clip.VoteScore, &clip.CommentCount,
			&clip.FavoriteCount, &clip.IsFeatured, &clip.IsNSFW, &clip.IsRemoved, &clip.RemovedReason,
			&clip.IsHidden, &clip.SubmittedByUserID, &clip.SubmittedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan clip: %w", err)
		}
		clips = append(clips, clip)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}

	return clips, nil
}

// ListForSitemapBroadcasters returns broadcasters with 5+ clips for sitemap generation.
func (r *ClipRepository) ListForSitemapBroadcasters(ctx context.Context) ([]models.BroadcasterWithClipCount, error) {
	query := `
		SELECT broadcaster_id, broadcaster_name, COUNT(*) as clip_count, COALESCE(SUM(view_count), 0) as total_views
		FROM clips
		WHERE is_removed = false AND broadcaster_id IS NOT NULL
		GROUP BY broadcaster_id, broadcaster_name
		HAVING COUNT(*) >= 5
		ORDER BY clip_count DESC
	`
	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to list sitemap broadcasters: %w", err)
	}
	defer rows.Close()

	var broadcasters []models.BroadcasterWithClipCount
	for rows.Next() {
		var b models.BroadcasterWithClipCount
		if err := rows.Scan(&b.BroadcasterID, &b.BroadcasterName, &b.ClipCount, &b.TotalViews); err != nil {
			return nil, fmt.Errorf("failed to scan broadcaster: %w", err)
		}
		broadcasters = append(broadcasters, b)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating broadcasters: %w", err)
	}
	return broadcasters, nil
}

// ListClipsForBestOf returns top clips within a date range, ordered by vote_score.
func (r *ClipRepository) ListClipsForBestOf(ctx context.Context, startDate, endDate time.Time, limit, offset int) ([]models.Clip, int, error) {
	countQuery := `
		SELECT COUNT(*)
		FROM clips
		WHERE is_removed = false AND created_at >= $1 AND created_at < $2
	`
	var total int
	if err := r.pool.QueryRow(ctx, countQuery, startDate, endDate).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("failed to count best-of clips: %w", err)
	}

	query := `
		SELECT
			id, twitch_clip_id, twitch_clip_url, embed_url, title,
			creator_name, creator_id, broadcaster_name, broadcaster_id,
			game_id, game_name, language, thumbnail_url, duration,
			view_count, created_at, imported_at, vote_score, comment_count,
			favorite_count, is_featured, is_nsfw, is_removed, removed_reason, is_hidden,
			submitted_by_user_id, submitted_at
		FROM clips
		WHERE is_removed = false AND created_at >= $1 AND created_at < $2
		ORDER BY vote_score DESC, view_count DESC
		LIMIT $3 OFFSET $4
	`
	rows, err := r.pool.Query(ctx, query, startDate, endDate, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list best-of clips: %w", err)
	}
	defer rows.Close()

	var clips []models.Clip
	for rows.Next() {
		var clip models.Clip
		if err := rows.Scan(
			&clip.ID, &clip.TwitchClipID, &clip.TwitchClipURL, &clip.EmbedURL,
			&clip.Title, &clip.CreatorName, &clip.CreatorID, &clip.BroadcasterName,
			&clip.BroadcasterID, &clip.GameID, &clip.GameName, &clip.Language,
			&clip.ThumbnailURL, &clip.Duration, &clip.ViewCount, &clip.CreatedAt,
			&clip.ImportedAt, &clip.VoteScore, &clip.CommentCount, &clip.FavoriteCount,
			&clip.IsFeatured, &clip.IsNSFW, &clip.IsRemoved, &clip.RemovedReason, &clip.IsHidden,
			&clip.SubmittedByUserID, &clip.SubmittedAt,
		); err != nil {
			return nil, 0, fmt.Errorf("failed to scan clip: %w", err)
		}
		clips = append(clips, clip)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("error iterating clips: %w", err)
	}
	return clips, total, nil
}

// ListClipsByGame returns top clips for a game, ordered by vote_score.
func (r *ClipRepository) ListClipsByGame(ctx context.Context, gameID string, limit, offset int) ([]models.Clip, int, error) {
	countQuery := `
		SELECT COUNT(*)
		FROM clips
		WHERE game_id = $1 AND is_removed = false
	`
	var total int
	if err := r.pool.QueryRow(ctx, countQuery, gameID).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("failed to count game clips: %w", err)
	}

	query := `
		SELECT
			id, twitch_clip_id, twitch_clip_url, embed_url, title,
			creator_name, creator_id, broadcaster_name, broadcaster_id,
			game_id, game_name, language, thumbnail_url, duration,
			view_count, created_at, imported_at, vote_score, comment_count,
			favorite_count, is_featured, is_nsfw, is_removed, removed_reason, is_hidden,
			submitted_by_user_id, submitted_at
		FROM clips
		WHERE game_id = $1 AND is_removed = false
		ORDER BY vote_score DESC, view_count DESC
		LIMIT $2 OFFSET $3
	`
	rows, err := r.pool.Query(ctx, query, gameID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list game clips: %w", err)
	}
	defer rows.Close()

	var clips []models.Clip
	for rows.Next() {
		var clip models.Clip
		if err := rows.Scan(
			&clip.ID, &clip.TwitchClipID, &clip.TwitchClipURL, &clip.EmbedURL,
			&clip.Title, &clip.CreatorName, &clip.CreatorID, &clip.BroadcasterName,
			&clip.BroadcasterID, &clip.GameID, &clip.GameName, &clip.Language,
			&clip.ThumbnailURL, &clip.Duration, &clip.ViewCount, &clip.CreatedAt,
			&clip.ImportedAt, &clip.VoteScore, &clip.CommentCount, &clip.FavoriteCount,
			&clip.IsFeatured, &clip.IsNSFW, &clip.IsRemoved, &clip.RemovedReason, &clip.IsHidden,
			&clip.SubmittedByUserID, &clip.SubmittedAt,
		); err != nil {
			return nil, 0, fmt.Errorf("failed to scan clip: %w", err)
		}
		clips = append(clips, clip)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("error iterating clips: %w", err)
	}
	return clips, total, nil
}

// ListClipsForStreamerGame returns top clips for a broadcaster+game combination.
func (r *ClipRepository) ListClipsForStreamerGame(ctx context.Context, broadcasterID, gameID string, limit, offset int) ([]models.Clip, int, error) {
	countQuery := `
		SELECT COUNT(*)
		FROM clips
		WHERE broadcaster_id = $1 AND game_id = $2 AND is_removed = false
	`
	var total int
	if err := r.pool.QueryRow(ctx, countQuery, broadcasterID, gameID).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("failed to count streamer+game clips: %w", err)
	}

	query := `
		SELECT
			id, twitch_clip_id, twitch_clip_url, embed_url, title,
			creator_name, creator_id, broadcaster_name, broadcaster_id,
			game_id, game_name, language, thumbnail_url, duration,
			view_count, created_at, imported_at, vote_score, comment_count,
			favorite_count, is_featured, is_nsfw, is_removed, removed_reason, is_hidden,
			submitted_by_user_id, submitted_at
		FROM clips
		WHERE broadcaster_id = $1 AND game_id = $2 AND is_removed = false
		ORDER BY vote_score DESC, view_count DESC
		LIMIT $3 OFFSET $4
	`
	rows, err := r.pool.Query(ctx, query, broadcasterID, gameID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list streamer+game clips: %w", err)
	}
	defer rows.Close()

	var clips []models.Clip
	for rows.Next() {
		var clip models.Clip
		if err := rows.Scan(
			&clip.ID, &clip.TwitchClipID, &clip.TwitchClipURL, &clip.EmbedURL,
			&clip.Title, &clip.CreatorName, &clip.CreatorID, &clip.BroadcasterName,
			&clip.BroadcasterID, &clip.GameID, &clip.GameName, &clip.Language,
			&clip.ThumbnailURL, &clip.Duration, &clip.ViewCount, &clip.CreatedAt,
			&clip.ImportedAt, &clip.VoteScore, &clip.CommentCount, &clip.FavoriteCount,
			&clip.IsFeatured, &clip.IsNSFW, &clip.IsRemoved, &clip.RemovedReason, &clip.IsHidden,
			&clip.SubmittedByUserID, &clip.SubmittedAt,
		); err != nil {
			return nil, 0, fmt.Errorf("failed to scan clip: %w", err)
		}
		clips = append(clips, clip)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("error iterating clips: %w", err)
	}
	return clips, total, nil
}
