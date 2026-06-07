package repository

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"git.subcult.tv/subculture-collective/clpr/internal/models"
	"git.subcult.tv/subculture-collective/clpr/internal/utils"
)

// DiscoveryClipRepository handles database operations for the discovery_clips staging table.
type DiscoveryClipRepository struct {
	pool *pgxpool.Pool
}

// NewDiscoveryClipRepository creates a new DiscoveryClipRepository.
func NewDiscoveryClipRepository(pool *pgxpool.Pool) *DiscoveryClipRepository {
	return &DiscoveryClipRepository{pool: pool}
}

// DiscoveryClipFilters holds query filters for listing discovery clips.
type DiscoveryClipFilters struct {
	GameID          *string
	BroadcasterID   *string
	CreatorID       *string
	Tag             *string
	ExcludeTags     []string
	Search          *string
	Language        *string
	Timeframe       *string
	DateFrom        *string
	DateTo          *string
	Sort            string // hot, new (default), top, views, trending
	Top10kStreamers bool
	ShowHidden      bool
}

// discoveryClipColumns is the standard SELECT column list.
const discoveryClipColumns = `
	d.id, d.twitch_clip_id, d.twitch_clip_url, d.embed_url, d.title,
	d.creator_name, d.creator_id, d.broadcaster_name, d.broadcaster_id,
	d.game_id, d.game_name, d.language, d.thumbnail_url, d.duration,
	d.view_count, d.created_at, d.imported_at, d.is_nsfw, d.is_removed, d.is_hidden
`

func scanDiscoveryClip(row pgx.Row) (*models.DiscoveryClip, error) {
	var dc models.DiscoveryClip
	err := row.Scan(
		&dc.ID, &dc.TwitchClipID, &dc.TwitchClipURL, &dc.EmbedURL, &dc.Title,
		&dc.CreatorName, &dc.CreatorID, &dc.BroadcasterName, &dc.BroadcasterID,
		&dc.GameID, &dc.GameName, &dc.Language, &dc.ThumbnailURL, &dc.Duration,
		&dc.ViewCount, &dc.CreatedAt, &dc.ImportedAt, &dc.IsNSFW, &dc.IsRemoved, &dc.IsHidden,
	)
	return &dc, err
}

func scanDiscoveryClips(rows pgx.Rows) ([]models.DiscoveryClip, error) {
	var clips []models.DiscoveryClip
	for rows.Next() {
		var dc models.DiscoveryClip
		err := rows.Scan(
			&dc.ID, &dc.TwitchClipID, &dc.TwitchClipURL, &dc.EmbedURL, &dc.Title,
			&dc.CreatorName, &dc.CreatorID, &dc.BroadcasterName, &dc.BroadcasterID,
			&dc.GameID, &dc.GameName, &dc.Language, &dc.ThumbnailURL, &dc.Duration,
			&dc.ViewCount, &dc.CreatedAt, &dc.ImportedAt, &dc.IsNSFW, &dc.IsRemoved, &dc.IsHidden,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan discovery clip: %w", err)
		}
		clips = append(clips, dc)
	}
	return clips, rows.Err()
}

// Create inserts a new discovery clip (used by the scraper).
func (r *DiscoveryClipRepository) Create(ctx context.Context, dc *models.DiscoveryClip) error {
	query := `
		INSERT INTO discovery_clips (
			id, twitch_clip_id, twitch_clip_url, embed_url, title,
			creator_name, creator_id, broadcaster_name, broadcaster_id,
			game_id, game_name, language, thumbnail_url, duration,
			view_count, created_at, imported_at, is_nsfw, is_removed, is_hidden
		) VALUES (
			$1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19,$20
		)
	`
	_, err := r.pool.Exec(ctx, query,
		dc.ID, dc.TwitchClipID, dc.TwitchClipURL, dc.EmbedURL, dc.Title,
		dc.CreatorName, dc.CreatorID, dc.BroadcasterName, dc.BroadcasterID,
		dc.GameID, dc.GameName, dc.Language, dc.ThumbnailURL, dc.Duration,
		dc.ViewCount, dc.CreatedAt, dc.ImportedAt, dc.IsNSFW, dc.IsRemoved, dc.IsHidden,
	)
	return err
}

// ExistsByTwitchClipID checks if a discovery clip exists with the given Twitch clip ID.
func (r *DiscoveryClipRepository) ExistsByTwitchClipID(ctx context.Context, twitchClipID string) (bool, error) {
	var exists bool
	err := r.pool.QueryRow(ctx,
		`SELECT EXISTS(SELECT 1 FROM discovery_clips WHERE twitch_clip_id = $1)`,
		twitchClipID,
	).Scan(&exists)
	return exists, err
}

// GetByTwitchClipID retrieves a discovery clip by its Twitch clip ID.
func (r *DiscoveryClipRepository) GetByTwitchClipID(ctx context.Context, twitchClipID string) (*models.DiscoveryClip, error) {
	query := fmt.Sprintf(`SELECT %s FROM discovery_clips d WHERE d.twitch_clip_id = $1 AND d.is_removed = false`, discoveryClipColumns)
	return scanDiscoveryClip(r.pool.QueryRow(ctx, query, twitchClipID))
}

// GetByID retrieves a discovery clip by its UUID.
func (r *DiscoveryClipRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.DiscoveryClip, error) {
	query := fmt.Sprintf(`SELECT %s FROM discovery_clips d WHERE d.id = $1 AND d.is_removed = false`, discoveryClipColumns)
	return scanDiscoveryClip(r.pool.QueryRow(ctx, query, id))
}

// Delete removes a discovery clip by ID (used after claiming).
func (r *DiscoveryClipRepository) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM discovery_clips WHERE id = $1`, id)
	return err
}

// DeleteByTwitchClipID removes a discovery clip by its Twitch clip ID.
func (r *DiscoveryClipRepository) DeleteByTwitchClipID(ctx context.Context, twitchClipID string) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM discovery_clips WHERE twitch_clip_id = $1`, twitchClipID)
	return err
}

// ClaimDiscoveryClip atomically moves a discovery clip into the main clips table.
// Returns the new Clip (now in the clips table) or an error if the clip doesn't
// exist in discovery_clips, or a race leads to a conflict.
func (r *DiscoveryClipRepository) ClaimDiscoveryClip(ctx context.Context, twitchClipID string, userID uuid.UUID, customTitle *string, isNSFW bool, broadcasterNameOverride *string) (*models.DiscoveryClip, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// Lock and fetch the discovery clip in one step
	lockQuery := fmt.Sprintf(`SELECT %s FROM discovery_clips d WHERE d.twitch_clip_id = $1 AND d.is_removed = false FOR UPDATE SKIP LOCKED`,
		discoveryClipColumns)
	dc, err := scanDiscoveryClip(tx.QueryRow(ctx, lockQuery, twitchClipID))
	if err != nil {
		if strings.Contains(err.Error(), "no rows") {
			return nil, fmt.Errorf("discovery clip not found or already claimed")
		}
		return nil, fmt.Errorf("failed to lock discovery clip: %w", err)
	}

	// Apply overrides
	title := dc.Title
	if customTitle != nil && *customTitle != "" {
		title = *customTitle
	}
	broadcasterName := dc.BroadcasterName
	if broadcasterNameOverride != nil && *broadcasterNameOverride != "" {
		broadcasterName = *broadcasterNameOverride
	}

	now := time.Now()

	// Insert into the main clips table
	insertQuery := `
		INSERT INTO clips (
			id, twitch_clip_id, twitch_clip_url, embed_url, title,
			creator_name, creator_id, broadcaster_name, broadcaster_id,
			game_id, game_name, language, thumbnail_url, duration,
			view_count, created_at, imported_at,
			vote_score, comment_count, favorite_count,
			is_featured, is_nsfw, is_removed, is_hidden,
			submitted_by_user_id, submitted_at
		) VALUES (
			$1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,
			0, 0, 0,
			false, $18, false, false,
			$19, $20
		)
	`
	_, err = tx.Exec(ctx, insertQuery,
		dc.ID, dc.TwitchClipID, dc.TwitchClipURL, dc.EmbedURL, title,
		dc.CreatorName, dc.CreatorID, broadcasterName, dc.BroadcasterID,
		dc.GameID, dc.GameName, dc.Language, dc.ThumbnailURL, dc.Duration,
		dc.ViewCount, dc.CreatedAt, now,
		isNSFW,
		userID, now,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to insert claimed clip: %w", err)
	}

	// Delete from discovery_clips
	_, err = tx.Exec(ctx, `DELETE FROM discovery_clips WHERE id = $1`, dc.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to delete discovery clip after claim: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("failed to commit claim transaction: %w", err)
	}

	return dc, nil
}

// ListWithFilters retrieves discovery clips with filters, sorting, and pagination.
func (r *DiscoveryClipRepository) ListWithFilters(ctx context.Context, filters DiscoveryClipFilters, limit, offset int) ([]models.DiscoveryClip, int, error) {
	if limit < 1 || limit > 100 {
		limit = 25
	}
	if offset < 0 {
		offset = 0
	}

	whereClauses := []string{"d.is_removed = false"}
	if !filters.ShowHidden {
		whereClauses = append(whereClauses, "d.is_hidden = false")
	}

	args := []interface{}{}
	argIdx := 1

	if filters.GameID != nil {
		whereClauses = append(whereClauses, fmt.Sprintf("d.game_id = %s", utils.SQLPlaceholder(argIdx)))
		args = append(args, *filters.GameID)
		argIdx++
	}
	if filters.BroadcasterID != nil {
		whereClauses = append(whereClauses, fmt.Sprintf("d.broadcaster_id = %s", utils.SQLPlaceholder(argIdx)))
		args = append(args, *filters.BroadcasterID)
		argIdx++
	}
	if filters.CreatorID != nil {
		whereClauses = append(whereClauses, fmt.Sprintf("d.creator_id = %s", utils.SQLPlaceholder(argIdx)))
		args = append(args, *filters.CreatorID)
		argIdx++
	}
	if filters.Search != nil && *filters.Search != "" {
		whereClauses = append(whereClauses, fmt.Sprintf("d.title ILIKE %s", utils.SQLPlaceholder(argIdx)))
		args = append(args, "%"+*filters.Search+"%")
		argIdx++
	}
	if filters.Language != nil && *filters.Language != "" {
		ph := utils.SQLPlaceholder(argIdx)
		whereClauses = append(whereClauses, fmt.Sprintf("(d.language = %s OR d.language = split_part(%s, '-', 1) OR d.language IS NULL OR d.language = '')", ph, ph))
		args = append(args, *filters.Language)
		argIdx++
	}
	if filters.Top10kStreamers {
		whereClauses = append(whereClauses, `EXISTS (SELECT 1 FROM top_streamers ts WHERE ts.broadcaster_id = d.broadcaster_id)`)
	}

	// Date range / timeframe
	if filters.DateFrom != nil && *filters.DateFrom != "" {
		whereClauses = append(whereClauses, fmt.Sprintf("d.created_at >= %s", utils.SQLPlaceholder(argIdx)))
		args = append(args, *filters.DateFrom)
		argIdx++
	}
	if filters.DateTo != nil && *filters.DateTo != "" {
		whereClauses = append(whereClauses, fmt.Sprintf("d.created_at <= %s", utils.SQLPlaceholder(argIdx)))
		args = append(args, *filters.DateTo)
		argIdx++
	}

	whereClause := "WHERE " + strings.Join(whereClauses, " AND ")

	// ORDER BY
	var orderBy string
	switch filters.Sort {
	case "views":
		orderBy = "ORDER BY d.view_count DESC, d.created_at DESC"
	case "trending":
		orderBy = "ORDER BY d.view_count DESC, d.created_at DESC"
	case "new":
		orderBy = "ORDER BY d.created_at DESC"
	default:
		orderBy = "ORDER BY d.created_at DESC"
	}

	// Count
	var total int
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM discovery_clips d %s", whereClause)
	if err := r.pool.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("failed to count discovery clips: %w", err)
	}

	// Data
	args = append(args, limit, offset)
	dataQuery := fmt.Sprintf(
		`SELECT %s FROM discovery_clips d %s %s LIMIT %s OFFSET %s`,
		discoveryClipColumns, whereClause, orderBy,
		utils.SQLPlaceholder(argIdx), utils.SQLPlaceholder(argIdx+1),
	)

	rows, err := r.pool.Query(ctx, dataQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list discovery clips: %w", err)
	}
	defer rows.Close()

	clips, err := scanDiscoveryClips(rows)
	if err != nil {
		return nil, 0, err
	}

	return clips, total, nil
}
