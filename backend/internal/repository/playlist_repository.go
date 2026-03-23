package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/subculture-collective/clipper/internal/models"
)

// PlaylistRepository handles database operations for playlists
type PlaylistRepository struct {
	pool *pgxpool.Pool
}

// NewPlaylistRepository creates a new PlaylistRepository
func NewPlaylistRepository(pool *pgxpool.Pool) *PlaylistRepository {
	return &PlaylistRepository{
		pool: pool,
	}
}

// Create creates a new playlist
func (r *PlaylistRepository) Create(ctx context.Context, playlist *models.Playlist) error {
	query := `
		INSERT INTO playlists (id, user_id, title, description, cover_url, visibility, is_curated, is_featured, display_order, script_id)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING created_at, updated_at
	`

	err := r.pool.QueryRow(ctx, query,
		playlist.ID,
		playlist.UserID,
		playlist.Title,
		playlist.Description,
		playlist.CoverURL,
		playlist.Visibility,
		playlist.IsCurated,
		playlist.IsFeatured,
		playlist.DisplayOrder,
		playlist.ScriptID,
	).Scan(&playlist.CreatedAt, &playlist.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to create playlist: %w", err)
	}

	return nil
}

// CreateWithItemsCopy creates a new playlist and copies items from another playlist
func (r *PlaylistRepository) CreateWithItemsCopy(ctx context.Context, playlist *models.Playlist, sourcePlaylistID uuid.UUID) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	insertPlaylist := `
		INSERT INTO playlists (id, user_id, title, description, cover_url, visibility, is_curated, is_featured, display_order, script_id)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING created_at, updated_at
	`

	err = tx.QueryRow(ctx, insertPlaylist,
		playlist.ID,
		playlist.UserID,
		playlist.Title,
		playlist.Description,
		playlist.CoverURL,
		playlist.Visibility,
		playlist.IsCurated,
		playlist.IsFeatured,
		playlist.DisplayOrder,
		playlist.ScriptID,
	).Scan(&playlist.CreatedAt, &playlist.UpdatedAt)
	if err != nil {
		return fmt.Errorf("failed to create playlist: %w", err)
	}

	copyItems := `
		INSERT INTO playlist_items (playlist_id, clip_id, order_index)
		SELECT $1, clip_id, order_index
		FROM playlist_items
		WHERE playlist_id = $2
		ORDER BY order_index ASC
	`

	if _, err = tx.Exec(ctx, copyItems, playlist.ID, sourcePlaylistID); err != nil {
		return fmt.Errorf("failed to copy playlist items: %w", err)
	}

	if err = tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// GetByID retrieves a playlist by its ID
func (r *PlaylistRepository) GetByID(ctx context.Context, playlistID uuid.UUID) (*models.Playlist, error) {
	query := `
		SELECT id, user_id, title, description, cover_url, visibility, share_token,
		       view_count, share_count, like_count, follower_count, bookmark_count,
		       is_curated, is_featured, display_order, script_id, slug,
		       created_at, updated_at, deleted_at
		FROM playlists
		WHERE id = $1 AND deleted_at IS NULL
	`

	var playlist models.Playlist
	err := r.pool.QueryRow(ctx, query, playlistID).Scan(
		&playlist.ID,
		&playlist.UserID,
		&playlist.Title,
		&playlist.Description,
		&playlist.CoverURL,
		&playlist.Visibility,
		&playlist.ShareToken,
		&playlist.ViewCount,
		&playlist.ShareCount,
		&playlist.LikeCount,
		&playlist.FollowerCount,
		&playlist.BookmarkCount,
		&playlist.IsCurated,
		&playlist.IsFeatured,
		&playlist.DisplayOrder,
		&playlist.ScriptID,
		&playlist.Slug,
		&playlist.CreatedAt,
		&playlist.UpdatedAt,
		&playlist.DeletedAt,
	)

	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get playlist: %w", err)
	}

	return &playlist, nil
}

// GetByShareToken retrieves a playlist by its share token
func (r *PlaylistRepository) GetByShareToken(ctx context.Context, shareToken string) (*models.Playlist, error) {
	query := `
		SELECT id, user_id, title, description, cover_url, visibility, share_token,
		       view_count, share_count, like_count, follower_count, bookmark_count,
		       is_curated, is_featured, display_order, script_id, slug,
		       created_at, updated_at, deleted_at
		FROM playlists
		WHERE share_token = $1 AND deleted_at IS NULL
	`

	var playlist models.Playlist
	err := r.pool.QueryRow(ctx, query, shareToken).Scan(
		&playlist.ID,
		&playlist.UserID,
		&playlist.Title,
		&playlist.Description,
		&playlist.CoverURL,
		&playlist.Visibility,
		&playlist.ShareToken,
		&playlist.ViewCount,
		&playlist.ShareCount,
		&playlist.LikeCount,
		&playlist.FollowerCount,
		&playlist.BookmarkCount,
		&playlist.IsCurated,
		&playlist.IsFeatured,
		&playlist.DisplayOrder,
		&playlist.ScriptID,
		&playlist.Slug,
		&playlist.CreatedAt,
		&playlist.UpdatedAt,
		&playlist.DeletedAt,
	)

	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get playlist by share token: %w", err)
	}

	return &playlist, nil
}

// Update updates a playlist
func (r *PlaylistRepository) Update(ctx context.Context, playlist *models.Playlist) error {
	query := `
		UPDATE playlists
		SET title = $1, description = $2, cover_url = $3, visibility = $4, share_token = $5
		WHERE id = $6 AND deleted_at IS NULL
		RETURNING updated_at
	`

	err := r.pool.QueryRow(ctx, query,
		playlist.Title,
		playlist.Description,
		playlist.CoverURL,
		playlist.Visibility,
		playlist.ShareToken,
		playlist.ID,
	).Scan(&playlist.UpdatedAt)

	if err == pgx.ErrNoRows {
		return fmt.Errorf("playlist not found")
	}
	if err != nil {
		return fmt.Errorf("failed to update playlist: %w", err)
	}

	return nil
}

// SoftDelete soft deletes a playlist
func (r *PlaylistRepository) SoftDelete(ctx context.Context, playlistID uuid.UUID) error {
	query := `
		UPDATE playlists
		SET deleted_at = $1
		WHERE id = $2 AND deleted_at IS NULL
	`

	result, err := r.pool.Exec(ctx, query, time.Now(), playlistID)
	if err != nil {
		return fmt.Errorf("failed to delete playlist: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("playlist not found")
	}

	return nil
}

// ListByUserID retrieves playlists owned by a user
func (r *PlaylistRepository) ListByUserID(ctx context.Context, userID uuid.UUID, currentUserID *uuid.UUID, limit, offset int) ([]*models.PlaylistListItem, int, error) {
	// Get total count
	countQuery := `
		SELECT COUNT(*)
		FROM playlists
		WHERE user_id = $1 AND deleted_at IS NULL
	`

	var total int
	err := r.pool.QueryRow(ctx, countQuery, userID).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count playlists: %w", err)
	}

	// Get playlists with clip count
	query := `
		SELECT
			p.id, p.user_id, p.title, p.description, p.cover_url, p.visibility, p.share_token,
			p.view_count, p.share_count, p.like_count, p.follower_count, p.bookmark_count,
			p.is_curated, p.is_featured, p.display_order, p.script_id, p.slug,
			p.created_at, p.updated_at, p.deleted_at,
			COALESCE(COUNT(pi.id), 0) AS clip_count,
			EXISTS (
				SELECT 1
				FROM playlist_items pi2
				JOIN clips c2 ON pi2.clip_id = c2.id
				WHERE pi2.playlist_id = p.id
				  AND (c2.status = 'processing' OR (c2.stream_source = 'stream' AND c2.video_url IS NULL))
			) AS has_processing_clips
		FROM playlists p
		LEFT JOIN playlist_items pi ON p.id = pi.playlist_id
		WHERE p.user_id = $1 AND p.deleted_at IS NULL
		GROUP BY p.id, p.user_id, p.title, p.description, p.cover_url, p.visibility, p.share_token,
		         p.view_count, p.share_count, p.like_count, p.follower_count, p.bookmark_count,
		         p.is_curated, p.is_featured, p.display_order, p.script_id, p.slug,
		         p.created_at, p.updated_at, p.deleted_at
		ORDER BY p.created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.pool.Query(ctx, query, userID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list playlists: %w", err)
	}
	defer rows.Close()

	var playlists []*models.PlaylistListItem
	for rows.Next() {
		var item models.PlaylistListItem
		err := rows.Scan(
			&item.ID,
			&item.UserID,
			&item.Title,
			&item.Description,
			&item.CoverURL,
			&item.Visibility,
			&item.ShareToken,
			&item.ViewCount,
			&item.ShareCount,
			&item.LikeCount,
			&item.FollowerCount,
			&item.BookmarkCount,
			&item.IsCurated,
			&item.IsFeatured,
			&item.DisplayOrder,
			&item.ScriptID,
			&item.Slug,
			&item.CreatedAt,
			&item.UpdatedAt,
			&item.DeletedAt,
			&item.ClipCount,
			&item.HasProcessingClips,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan playlist: %w", err)
		}
		playlists = append(playlists, &item)
	}

	if currentUserID != nil {
		if err := r.enrichPlaylistInteractionStates(ctx, *currentUserID, playlists); err != nil {
			return nil, 0, err
		}
	}

	// Fetch preview clips for each playlist (first 4)
	for _, playlist := range playlists {
		previewQuery := `
			SELECT c.id, c.twitch_clip_id, c.title, c.broadcaster_name, c.thumbnail_url,
			       c.duration, c.view_count, c.created_at
			FROM clips c
			INNER JOIN playlist_items pi ON c.id = pi.clip_id
			WHERE pi.playlist_id = $1
			ORDER BY pi.order_index ASC
			LIMIT 4
		`
		previewRows, err := r.pool.Query(ctx, previewQuery, playlist.ID)
		if err != nil {
			continue // Skip preview clips on error
		}

		var previewClips []models.Clip
		for previewRows.Next() {
			var clip models.Clip
			err := previewRows.Scan(
				&clip.ID,
				&clip.TwitchClipID,
				&clip.Title,
				&clip.BroadcasterName,
				&clip.ThumbnailURL,
				&clip.Duration,
				&clip.ViewCount,
				&clip.CreatedAt,
			)
			if err == nil {
				previewClips = append(previewClips, clip)
			}
		}
		previewRows.Close()
		playlist.PreviewClips = previewClips
	}

	return playlists, total, nil
}

// ListPublic retrieves public playlists for discovery
func (r *PlaylistRepository) ListPublic(ctx context.Context, currentUserID *uuid.UUID, limit, offset int) ([]*models.PlaylistListItem, int, error) {
	// Get total count
	countQuery := `
		SELECT COUNT(*)
		FROM playlists
		WHERE visibility = 'public' AND deleted_at IS NULL
	`

	var total int
	err := r.pool.QueryRow(ctx, countQuery).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count public playlists: %w", err)
	}

	// Get playlists with clip count
	query := `
		SELECT
			p.id, p.user_id, p.title, p.description, p.cover_url, p.visibility, p.share_token,
			p.view_count, p.share_count, p.like_count, p.follower_count, p.bookmark_count,
			p.is_curated, p.is_featured, p.display_order, p.script_id, p.slug,
			p.created_at, p.updated_at, p.deleted_at,
			COALESCE(COUNT(pi.id), 0) AS clip_count,
			EXISTS (
				SELECT 1
				FROM playlist_items pi2
				JOIN clips c2 ON pi2.clip_id = c2.id
				WHERE pi2.playlist_id = p.id
				  AND (c2.status = 'processing' OR (c2.stream_source = 'stream' AND c2.video_url IS NULL))
			) AS has_processing_clips
		FROM playlists p
		LEFT JOIN playlist_items pi ON p.id = pi.playlist_id
		WHERE p.visibility = 'public' AND p.deleted_at IS NULL
		GROUP BY p.id, p.user_id, p.title, p.description, p.cover_url, p.visibility, p.share_token,
		         p.view_count, p.share_count, p.like_count, p.follower_count, p.bookmark_count,
		         p.is_curated, p.is_featured, p.display_order, p.script_id, p.slug,
		         p.created_at, p.updated_at, p.deleted_at
		ORDER BY p.like_count DESC, p.created_at DESC
		LIMIT $1 OFFSET $2
	`

	rows, err := r.pool.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list public playlists: %w", err)
	}
	defer rows.Close()

	var playlists []*models.PlaylistListItem
	for rows.Next() {
		var item models.PlaylistListItem
		err := rows.Scan(
			&item.ID,
			&item.UserID,
			&item.Title,
			&item.Description,
			&item.CoverURL,
			&item.Visibility,
			&item.ShareToken,
			&item.ViewCount,
			&item.ShareCount,
			&item.LikeCount,
			&item.FollowerCount,
			&item.BookmarkCount,
			&item.IsCurated,
			&item.IsFeatured,
			&item.DisplayOrder,
			&item.ScriptID,
			&item.Slug,
			&item.CreatedAt,
			&item.UpdatedAt,
			&item.DeletedAt,
			&item.ClipCount,
			&item.HasProcessingClips,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan playlist: %w", err)
		}
		playlists = append(playlists, &item)
	}

	if currentUserID != nil {
		if err := r.enrichPlaylistInteractionStates(ctx, *currentUserID, playlists); err != nil {
			return nil, 0, err
		}
	}

	// Fetch preview clips for each playlist (first 4)
	for _, playlist := range playlists {
		previewQuery := `
			SELECT c.id, c.twitch_clip_id, c.title, c.broadcaster_name, c.thumbnail_url,
			       c.duration, c.view_count, c.created_at
			FROM clips c
			INNER JOIN playlist_items pi ON c.id = pi.clip_id
			WHERE pi.playlist_id = $1
			ORDER BY pi.order_index ASC
			LIMIT 4
		`
		previewRows, err := r.pool.Query(ctx, previewQuery, playlist.ID)
		if err != nil {
			continue // Skip preview clips on error
		}

		var previewClips []models.Clip
		for previewRows.Next() {
			var clip models.Clip
			err := previewRows.Scan(
				&clip.ID,
				&clip.TwitchClipID,
				&clip.Title,
				&clip.BroadcasterName,
				&clip.ThumbnailURL,
				&clip.Duration,
				&clip.ViewCount,
				&clip.CreatedAt,
			)
			if err == nil {
				previewClips = append(previewClips, clip)
			}
		}
		previewRows.Close()
		playlist.PreviewClips = previewClips
	}

	return playlists, total, nil
}

// ListBookmarkedByUser retrieves playlists bookmarked by a user
func (r *PlaylistRepository) ListBookmarkedByUser(ctx context.Context, userID uuid.UUID, currentUserID *uuid.UUID, limit, offset int) ([]*models.PlaylistListItem, int, error) {
	// Get total count
	countQuery := `
		SELECT COUNT(*)
		FROM playlist_bookmarks pb
		JOIN playlists p ON pb.playlist_id = p.id
		WHERE pb.user_id = $1 AND p.deleted_at IS NULL
	`

	var total int
	err := r.pool.QueryRow(ctx, countQuery, userID).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count bookmarked playlists: %w", err)
	}

	// Get playlists with clip count
	query := `
		SELECT
			p.id, p.user_id, p.title, p.description, p.cover_url, p.visibility, p.share_token,
			p.view_count, p.share_count, p.like_count, p.follower_count, p.bookmark_count,
			p.is_curated, p.is_featured, p.display_order, p.script_id, p.slug,
			p.created_at, p.updated_at, p.deleted_at,
			COALESCE(COUNT(pi.id), 0) AS clip_count,
			EXISTS (
				SELECT 1
				FROM playlist_items pi2
				JOIN clips c2 ON pi2.clip_id = c2.id
				WHERE pi2.playlist_id = p.id
				  AND (c2.status = 'processing' OR (c2.stream_source = 'stream' AND c2.video_url IS NULL))
			) AS has_processing_clips
		FROM playlists p
		INNER JOIN playlist_bookmarks pb ON p.id = pb.playlist_id
		LEFT JOIN playlist_items pi ON p.id = pi.playlist_id
		WHERE pb.user_id = $1 AND p.deleted_at IS NULL
		GROUP BY p.id, p.user_id, p.title, p.description, p.cover_url, p.visibility, p.share_token,
		         p.view_count, p.share_count, p.like_count, p.follower_count, p.bookmark_count,
		         p.is_curated, p.is_featured, p.display_order, p.script_id, p.slug,
		         p.created_at, p.updated_at, p.deleted_at, pb.bookmarked_at
		ORDER BY pb.bookmarked_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.pool.Query(ctx, query, userID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list bookmarked playlists: %w", err)
	}
	defer rows.Close()

	var playlists []*models.PlaylistListItem
	for rows.Next() {
		var item models.PlaylistListItem
		err := rows.Scan(
			&item.ID,
			&item.UserID,
			&item.Title,
			&item.Description,
			&item.CoverURL,
			&item.Visibility,
			&item.ShareToken,
			&item.ViewCount,
			&item.ShareCount,
			&item.LikeCount,
			&item.FollowerCount,
			&item.BookmarkCount,
			&item.IsCurated,
			&item.IsFeatured,
			&item.DisplayOrder,
			&item.ScriptID,
			&item.Slug,
			&item.CreatedAt,
			&item.UpdatedAt,
			&item.DeletedAt,
			&item.ClipCount,
			&item.HasProcessingClips,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan playlist: %w", err)
		}
		playlists = append(playlists, &item)
	}

	if currentUserID != nil {
		if err := r.enrichPlaylistInteractionStates(ctx, *currentUserID, playlists); err != nil {
			return nil, 0, err
		}
	}

	// Fetch preview clips for each playlist (first 4)
	for _, playlist := range playlists {
		previewQuery := `
			SELECT c.id, c.twitch_clip_id, c.title, c.broadcaster_name, c.thumbnail_url,
			       c.duration, c.view_count, c.created_at
			FROM clips c
			INNER JOIN playlist_items pi ON c.id = pi.clip_id
			WHERE pi.playlist_id = $1
			ORDER BY pi.order_index ASC
			LIMIT 4
		`
		previewRows, err := r.pool.Query(ctx, previewQuery, playlist.ID)
		if err != nil {
			continue // Skip preview clips on error
		}

		var previewClips []models.Clip
		for previewRows.Next() {
			var clip models.Clip
			err := previewRows.Scan(
				&clip.ID,
				&clip.TwitchClipID,
				&clip.Title,
				&clip.BroadcasterName,
				&clip.ThumbnailURL,
				&clip.Duration,
				&clip.ViewCount,
				&clip.CreatedAt,
			)
			if err == nil {
				previewClips = append(previewClips, clip)
			}
		}
		previewRows.Close()
		playlist.PreviewClips = previewClips
	}

	return playlists, total, nil
}

// AddClip adds a clip to a playlist
func (r *PlaylistRepository) AddClip(ctx context.Context, playlistID, clipID uuid.UUID, orderIndex int) error {
	query := `
		INSERT INTO playlist_items (playlist_id, clip_id, order_index)
		VALUES ($1, $2, $3)
		ON CONFLICT (playlist_id, clip_id) DO NOTHING
	`

	_, err := r.pool.Exec(ctx, query, playlistID, clipID, orderIndex)
	if err != nil {
		return fmt.Errorf("failed to add clip to playlist: %w", err)
	}

	return nil
}

// RemoveClip removes a clip from a playlist
func (r *PlaylistRepository) RemoveClip(ctx context.Context, playlistID, clipID uuid.UUID) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	query := `
		DELETE FROM playlist_items
		WHERE playlist_id = $1 AND clip_id = $2
	`

	result, err := tx.Exec(ctx, query, playlistID, clipID)
	if err != nil {
		return fmt.Errorf("failed to remove clip from playlist: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("clip not found in playlist")
	}

	// Reindex order_index for remaining items to keep ordering continuous
	reindexQuery := `
		WITH ordered AS (
			SELECT
				clip_id,
				ROW_NUMBER() OVER (ORDER BY order_index) - 1 AS new_index
			FROM playlist_items
			WHERE playlist_id = $1
		)
		UPDATE playlist_items AS pi
		SET order_index = o.new_index
		FROM ordered AS o
		WHERE pi.playlist_id = $1
		  AND pi.clip_id = o.clip_id
	`

	if _, err := tx.Exec(ctx, reindexQuery, playlistID); err != nil {
		return fmt.Errorf("failed to reindex playlist items after removal: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// GetClips retrieves clips in a playlist with pagination
func (r *PlaylistRepository) GetClips(ctx context.Context, playlistID uuid.UUID, limit, offset int) ([]models.PlaylistClipRef, int, error) {
	// Get total count
	countQuery := `
		SELECT COUNT(*)
		FROM playlist_items
		WHERE playlist_id = $1
	`

	var total int
	err := r.pool.QueryRow(ctx, countQuery, playlistID).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count clips: %w", err)
	}

	// Get clips
	query := `
		SELECT
			c.id, c.twitch_clip_id, c.twitch_clip_url, c.embed_url, c.title,
			c.creator_name, c.creator_id, c.broadcaster_name, c.broadcaster_id,
			c.game_id, c.game_name, c.language, c.thumbnail_url, c.duration,
			c.view_count, c.created_at, c.imported_at, c.vote_score,
			c.comment_count, c.favorite_count, c.is_featured, c.is_nsfw,
			c.is_removed, c.removed_reason, c.is_hidden,
			c.submitted_by_user_id, c.submitted_at,
			c.trending_score, c.hot_score, c.popularity_index, c.engagement_count,
			c.dmca_removed, c.dmca_notice_id, c.dmca_removed_at, c.dmca_reinstated_at,
			c.stream_source, c.status, c.video_url, c.processed_at, c.quality, c.start_time, c.end_time,
			pi.order_index
		FROM playlist_items pi
		JOIN clips c ON pi.clip_id = c.id
		WHERE pi.playlist_id = $1
		ORDER BY pi.order_index ASC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.pool.Query(ctx, query, playlistID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get clips: %w", err)
	}
	defer rows.Close()

	var clips []models.PlaylistClipRef
	for rows.Next() {
		var clipRef models.PlaylistClipRef
		err := rows.Scan(
			&clipRef.ID,
			&clipRef.TwitchClipID,
			&clipRef.TwitchClipURL,
			&clipRef.EmbedURL,
			&clipRef.Title,
			&clipRef.CreatorName,
			&clipRef.CreatorID,
			&clipRef.BroadcasterName,
			&clipRef.BroadcasterID,
			&clipRef.GameID,
			&clipRef.GameName,
			&clipRef.Language,
			&clipRef.ThumbnailURL,
			&clipRef.Duration,
			&clipRef.ViewCount,
			&clipRef.CreatedAt,
			&clipRef.ImportedAt,
			&clipRef.VoteScore,
			&clipRef.CommentCount,
			&clipRef.FavoriteCount,
			&clipRef.IsFeatured,
			&clipRef.IsNSFW,
			&clipRef.IsRemoved,
			&clipRef.RemovedReason,
			&clipRef.IsHidden,
			&clipRef.SubmittedByUserID,
			&clipRef.SubmittedAt,
			&clipRef.TrendingScore,
			&clipRef.HotScore,
			&clipRef.PopularityIndex,
			&clipRef.EngagementCount,
			&clipRef.DMCARemoved,
			&clipRef.DMCANoticeID,
			&clipRef.DMCARemovedAt,
			&clipRef.DMCAReinstatedAt,
			&clipRef.StreamSource,
			&clipRef.Status,
			&clipRef.VideoURL,
			&clipRef.ProcessedAt,
			&clipRef.Quality,
			&clipRef.StartTime,
			&clipRef.EndTime,
			&clipRef.OrderIndex,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan clip: %w", err)
		}
		clips = append(clips, clipRef)
	}

	return clips, total, nil
}

// GetClipCount returns the number of clips in a playlist
func (r *PlaylistRepository) GetClipCount(ctx context.Context, playlistID uuid.UUID) (int, error) {
	query := `
		SELECT COUNT(*)
		FROM playlist_items
		WHERE playlist_id = $1
	`

	var count int
	err := r.pool.QueryRow(ctx, query, playlistID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to get clip count: %w", err)
	}

	return count, nil
}

// HasClip checks if a clip is already in a playlist
func (r *PlaylistRepository) HasClip(ctx context.Context, playlistID, clipID uuid.UUID) (bool, error) {
	query := `
		SELECT EXISTS(
			SELECT 1 FROM playlist_items
			WHERE playlist_id = $1 AND clip_id = $2
		)
	`

	var exists bool
	err := r.pool.QueryRow(ctx, query, playlistID, clipID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check clip existence: %w", err)
	}

	return exists, nil
}

// ReorderClips updates the order of clips in a playlist
func (r *PlaylistRepository) ReorderClips(ctx context.Context, playlistID uuid.UUID, clipIDs []uuid.UUID) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	query := `
		UPDATE playlist_items
		SET order_index = $1
		WHERE playlist_id = $2 AND clip_id = $3
	`

	for i, clipID := range clipIDs {
		_, err := tx.Exec(ctx, query, i, playlistID, clipID)
		if err != nil {
			return fmt.Errorf("failed to update order for clip %s: %w", clipID, err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// LikePlaylist adds a like to a playlist
func (r *PlaylistRepository) LikePlaylist(ctx context.Context, userID, playlistID uuid.UUID) error {
	query := `
		INSERT INTO playlist_likes (user_id, playlist_id)
		VALUES ($1, $2)
		ON CONFLICT (user_id, playlist_id) DO NOTHING
	`

	_, err := r.pool.Exec(ctx, query, userID, playlistID)
	if err != nil {
		return fmt.Errorf("failed to like playlist: %w", err)
	}

	return nil
}

// BookmarkPlaylist adds a bookmark to a playlist.
func (r *PlaylistRepository) BookmarkPlaylist(ctx context.Context, userID, playlistID uuid.UUID) error {
	query := `
		INSERT INTO playlist_bookmarks (user_id, playlist_id)
		VALUES ($1, $2)
		ON CONFLICT (user_id, playlist_id) DO NOTHING
	`

	_, err := r.pool.Exec(ctx, query, userID, playlistID)
	if err != nil {
		return fmt.Errorf("failed to bookmark playlist: %w", err)
	}

	return nil
}

// UnbookmarkPlaylist removes a bookmark from a playlist.
func (r *PlaylistRepository) UnbookmarkPlaylist(ctx context.Context, userID, playlistID uuid.UUID) error {
	query := `
		DELETE FROM playlist_bookmarks
		WHERE user_id = $1 AND playlist_id = $2
	`

	_, err := r.pool.Exec(ctx, query, userID, playlistID)
	if err != nil {
		return fmt.Errorf("failed to unbookmark playlist: %w", err)
	}

	return nil
}

// UnlikePlaylist removes a like from a playlist
func (r *PlaylistRepository) UnlikePlaylist(ctx context.Context, userID, playlistID uuid.UUID) error {
	query := `
		DELETE FROM playlist_likes
		WHERE user_id = $1 AND playlist_id = $2
	`

	_, err := r.pool.Exec(ctx, query, userID, playlistID)
	if err != nil {
		return fmt.Errorf("failed to unlike playlist: %w", err)
	}

	return nil
}

// IsLiked checks if a user has liked a playlist
func (r *PlaylistRepository) IsLiked(ctx context.Context, userID, playlistID uuid.UUID) (bool, error) {
	query := `
		SELECT EXISTS(
			SELECT 1 FROM playlist_likes
			WHERE user_id = $1 AND playlist_id = $2
		)
	`

	var liked bool
	err := r.pool.QueryRow(ctx, query, userID, playlistID).Scan(&liked)
	if err != nil {
		return false, fmt.Errorf("failed to check if liked: %w", err)
	}

	return liked, nil
}

// IsBookmarked checks if a user has bookmarked a playlist.
func (r *PlaylistRepository) IsBookmarked(ctx context.Context, userID, playlistID uuid.UUID) (bool, error) {
	query := `
		SELECT EXISTS(
			SELECT 1 FROM playlist_bookmarks
			WHERE user_id = $1 AND playlist_id = $2
		)
	`

	var bookmarked bool
	err := r.pool.QueryRow(ctx, query, userID, playlistID).Scan(&bookmarked)
	if err != nil {
		return false, fmt.Errorf("failed to check if bookmarked: %w", err)
	}

	return bookmarked, nil
}

func (r *PlaylistRepository) enrichPlaylistInteractionStates(ctx context.Context, userID uuid.UUID, playlists []*models.PlaylistListItem) error {
	if len(playlists) == 0 {
		return nil
	}

	playlistIDs := make([]uuid.UUID, 0, len(playlists))
	for _, playlist := range playlists {
		playlistIDs = append(playlistIDs, playlist.ID)
	}

	likedMap := make(map[uuid.UUID]bool, len(playlists))
	likedRows, err := r.pool.Query(ctx, `
		SELECT playlist_id
		FROM playlist_likes
		WHERE user_id = $1 AND playlist_id = ANY($2)
	`, userID, playlistIDs)
	if err != nil {
		return fmt.Errorf("failed to load liked playlists: %w", err)
	}
	for likedRows.Next() {
		var playlistID uuid.UUID
		if scanErr := likedRows.Scan(&playlistID); scanErr != nil {
			likedRows.Close()
			return fmt.Errorf("failed to scan liked playlist: %w", scanErr)
		}
		likedMap[playlistID] = true
	}
	likedRows.Close()

	bookmarkedMap := make(map[uuid.UUID]bool, len(playlists))
	bookmarkRows, err := r.pool.Query(ctx, `
		SELECT playlist_id
		FROM playlist_bookmarks
		WHERE user_id = $1 AND playlist_id = ANY($2)
	`, userID, playlistIDs)
	if err != nil {
		return fmt.Errorf("failed to load bookmarked playlists: %w", err)
	}
	for bookmarkRows.Next() {
		var playlistID uuid.UUID
		if scanErr := bookmarkRows.Scan(&playlistID); scanErr != nil {
			bookmarkRows.Close()
			return fmt.Errorf("failed to scan bookmarked playlist: %w", scanErr)
		}
		bookmarkedMap[playlistID] = true
	}
	bookmarkRows.Close()

	for _, playlist := range playlists {
		playlist.IsLiked = likedMap[playlist.ID]
		playlist.IsBookmarked = bookmarkedMap[playlist.ID]
	}

	return nil
}

// GetCreator retrieves the creator of a playlist
func (r *PlaylistRepository) GetCreator(ctx context.Context, userID uuid.UUID) (*models.User, error) {
	query := `
		SELECT id, username, display_name, avatar_url
		FROM users
		WHERE id = $1
	`

	var user models.User
	err := r.pool.QueryRow(ctx, query, userID).Scan(
		&user.ID,
		&user.Username,
		&user.DisplayName,
		&user.AvatarURL,
	)

	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get creator: %w", err)
	}

	return &user, nil
}

// UpdateShareToken updates or generates a share token for a playlist
func (r *PlaylistRepository) UpdateShareToken(ctx context.Context, playlistID uuid.UUID, shareToken string) error {
	query := `
		UPDATE playlists
		SET share_token = $1, updated_at = NOW()
		WHERE id = $2 AND deleted_at IS NULL
	`

	result, err := r.pool.Exec(ctx, query, shareToken, playlistID)
	if err != nil {
		return fmt.Errorf("failed to update share token: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("playlist not found")
	}

	return nil
}

// IncrementViewCount increments the view count for a playlist
func (r *PlaylistRepository) IncrementViewCount(ctx context.Context, playlistID uuid.UUID) error {
	query := `
		UPDATE playlists
		SET view_count = view_count + 1
		WHERE id = $1 AND deleted_at IS NULL
	`

	_, err := r.pool.Exec(ctx, query, playlistID)
	if err != nil {
		return fmt.Errorf("failed to increment view count: %w", err)
	}

	return nil
}

// IncrementShareCount increments the share count for a playlist
func (r *PlaylistRepository) IncrementShareCount(ctx context.Context, playlistID uuid.UUID) error {
	query := `
		UPDATE playlists
		SET share_count = share_count + 1
		WHERE id = $1 AND deleted_at IS NULL
	`

	_, err := r.pool.Exec(ctx, query, playlistID)
	if err != nil {
		return fmt.Errorf("failed to increment share count: %w", err)
	}

	return nil
}

// AddCollaborator adds a collaborator to a playlist
func (r *PlaylistRepository) AddCollaborator(ctx context.Context, collaborator *models.PlaylistCollaborator) error {
	query := `
		INSERT INTO playlist_collaborators (id, playlist_id, user_id, permission, invited_by, invited_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (playlist_id, user_id)
		DO UPDATE SET permission = EXCLUDED.permission, updated_at = NOW()
		RETURNING created_at, updated_at
	`

	err := r.pool.QueryRow(ctx, query,
		collaborator.ID,
		collaborator.PlaylistID,
		collaborator.UserID,
		collaborator.Permission,
		collaborator.InvitedBy,
		collaborator.InvitedAt,
	).Scan(&collaborator.CreatedAt, &collaborator.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to add collaborator: %w", err)
	}

	return nil
}

// UpdateCollaboratorPermission updates a collaborator's permission level
func (r *PlaylistRepository) UpdateCollaboratorPermission(ctx context.Context, playlistID, userID uuid.UUID, permission string) error {
	query := `
		UPDATE playlist_collaborators
		SET permission = $1, updated_at = NOW()
		WHERE playlist_id = $2 AND user_id = $3
	`

	result, err := r.pool.Exec(ctx, query, permission, playlistID, userID)
	if err != nil {
		return fmt.Errorf("failed to update collaborator permission: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("collaborator not found")
	}

	return nil
}

// RemoveCollaborator removes a collaborator from a playlist
func (r *PlaylistRepository) RemoveCollaborator(ctx context.Context, playlistID, userID uuid.UUID) error {
	query := `
		DELETE FROM playlist_collaborators
		WHERE playlist_id = $1 AND user_id = $2
	`

	result, err := r.pool.Exec(ctx, query, playlistID, userID)
	if err != nil {
		return fmt.Errorf("failed to remove collaborator: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("collaborator not found")
	}

	return nil
}

// GetCollaborators retrieves all collaborators for a playlist
func (r *PlaylistRepository) GetCollaborators(ctx context.Context, playlistID uuid.UUID) ([]*models.PlaylistCollaborator, error) {
	query := `
		SELECT pc.id, pc.playlist_id, pc.user_id, pc.permission, pc.invited_by, pc.invited_at,
		       pc.created_at, pc.updated_at,
		       u.id, u.username, u.display_name, u.avatar_url
		FROM playlist_collaborators pc
		JOIN users u ON pc.user_id = u.id
		WHERE pc.playlist_id = $1
		ORDER BY pc.created_at ASC
	`

	rows, err := r.pool.Query(ctx, query, playlistID)
	if err != nil {
		return nil, fmt.Errorf("failed to get collaborators: %w", err)
	}
	defer rows.Close()

	var collaborators []*models.PlaylistCollaborator
	for rows.Next() {
		var collab models.PlaylistCollaborator
		collab.User = &models.User{}

		err := rows.Scan(
			&collab.ID,
			&collab.PlaylistID,
			&collab.UserID,
			&collab.Permission,
			&collab.InvitedBy,
			&collab.InvitedAt,
			&collab.CreatedAt,
			&collab.UpdatedAt,
			&collab.User.ID,
			&collab.User.Username,
			&collab.User.DisplayName,
			&collab.User.AvatarURL,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan collaborator: %w", err)
		}
		collaborators = append(collaborators, &collab)
	}

	return collaborators, nil
}

// GetCollaboratorPermission retrieves a user's permission level for a playlist
func (r *PlaylistRepository) GetCollaboratorPermission(ctx context.Context, playlistID, userID uuid.UUID) (string, error) {
	query := `
		SELECT permission
		FROM playlist_collaborators
		WHERE playlist_id = $1 AND user_id = $2
	`

	var permission string
	err := r.pool.QueryRow(ctx, query, playlistID, userID).Scan(&permission)

	if err == pgx.ErrNoRows {
		return "", nil
	}
	if err != nil {
		return "", fmt.Errorf("failed to get collaborator permission: %w", err)
	}

	return permission, nil
}

// IsCollaborator checks if a user is a collaborator on a playlist
func (r *PlaylistRepository) IsCollaborator(ctx context.Context, playlistID, userID uuid.UUID) (bool, error) {
	query := `
		SELECT EXISTS(
			SELECT 1 FROM playlist_collaborators
			WHERE playlist_id = $1 AND user_id = $2
		)
	`

	var exists bool
	err := r.pool.QueryRow(ctx, query, playlistID, userID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check if collaborator: %w", err)
	}

	return exists, nil
}

// ListFeatured returns playlists that are featured or curated, public, ordered by display_order.
func (r *PlaylistRepository) ListFeatured(ctx context.Context, currentUserID *uuid.UUID, limit, offset int) ([]*models.PlaylistListItem, int, error) {
	countQuery := `
		WITH featured_source AS (
			SELECT ROW_NUMBER() OVER (
				PARTITION BY COALESCE(script_id, id)
				ORDER BY created_at DESC
			) AS latest_rank
			FROM playlists
			WHERE (is_featured = true OR is_curated = true)
			  AND visibility = 'public' AND deleted_at IS NULL
		)
		SELECT COUNT(*)
		FROM featured_source
		WHERE latest_rank = 1
	`

	var total int
	err := r.pool.QueryRow(ctx, countQuery).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count featured playlists: %w", err)
	}

	query := `
		WITH featured_source AS (
			SELECT p.*,
			       ROW_NUMBER() OVER (
				   PARTITION BY COALESCE(p.script_id, p.id)
				   ORDER BY p.created_at DESC
			   ) AS latest_rank
			FROM playlists p
			WHERE (p.is_featured = true OR p.is_curated = true)
			  AND p.visibility = 'public' AND p.deleted_at IS NULL
		)
		SELECT
			p.id, p.user_id, p.title, p.description, p.cover_url, p.visibility, p.share_token,
			p.view_count, p.share_count, p.like_count, p.follower_count, p.bookmark_count,
			p.is_curated, p.is_featured, p.display_order, p.script_id, p.slug,
			p.created_at, p.updated_at, p.deleted_at,
			COALESCE(COUNT(pi.id), 0) AS clip_count,
			EXISTS (
				SELECT 1
				FROM playlist_items pi2
				JOIN clips c2 ON pi2.clip_id = c2.id
				WHERE pi2.playlist_id = p.id
				  AND (c2.status = 'processing' OR (c2.stream_source = 'stream' AND c2.video_url IS NULL))
			) AS has_processing_clips
		FROM featured_source p
		LEFT JOIN playlist_items pi ON p.id = pi.playlist_id
		WHERE p.latest_rank = 1
		GROUP BY p.id, p.user_id, p.title, p.description, p.cover_url, p.visibility, p.share_token,
		         p.view_count, p.share_count, p.like_count, p.follower_count, p.bookmark_count,
		         p.is_curated, p.is_featured, p.display_order, p.script_id, p.slug,
		         p.created_at, p.updated_at, p.deleted_at
		ORDER BY p.display_order ASC, p.created_at DESC
		LIMIT $1 OFFSET $2
	`

	rows, err := r.pool.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list featured playlists: %w", err)
	}
	defer rows.Close()

	var playlists []*models.PlaylistListItem
	for rows.Next() {
		var item models.PlaylistListItem
		err := rows.Scan(
			&item.ID,
			&item.UserID,
			&item.Title,
			&item.Description,
			&item.CoverURL,
			&item.Visibility,
			&item.ShareToken,
			&item.ViewCount,
			&item.ShareCount,
			&item.LikeCount,
			&item.FollowerCount,
			&item.BookmarkCount,
			&item.IsCurated,
			&item.IsFeatured,
			&item.DisplayOrder,
			&item.ScriptID,
			&item.Slug,
			&item.CreatedAt,
			&item.UpdatedAt,
			&item.DeletedAt,
			&item.ClipCount,
			&item.HasProcessingClips,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan featured playlist: %w", err)
		}
		playlists = append(playlists, &item)
	}

	if currentUserID != nil {
		if err := r.enrichPlaylistInteractionStates(ctx, *currentUserID, playlists); err != nil {
			return nil, 0, err
		}
	}

	// Fetch preview clips for each playlist (first 4)
	for _, playlist := range playlists {
		previewQuery := `
			SELECT c.id, c.twitch_clip_id, c.title, c.broadcaster_name, c.thumbnail_url,
			       c.duration, c.view_count, c.created_at
			FROM clips c
			INNER JOIN playlist_items pi ON c.id = pi.clip_id
			WHERE pi.playlist_id = $1
			ORDER BY pi.order_index ASC
			LIMIT 4
		`
		previewRows, err := r.pool.Query(ctx, previewQuery, playlist.ID)
		if err != nil {
			continue
		}

		var previewClips []models.Clip
		for previewRows.Next() {
			var clip models.Clip
			err := previewRows.Scan(
				&clip.ID,
				&clip.TwitchClipID,
				&clip.Title,
				&clip.BroadcasterName,
				&clip.ThumbnailURL,
				&clip.Duration,
				&clip.ViewCount,
				&clip.CreatedAt,
			)
			if err == nil {
				previewClips = append(previewClips, clip)
			}
		}
		previewRows.Close()
		playlist.PreviewClips = previewClips
	}

	return playlists, total, nil
}

// GetPlaylistOfTheDay returns the most recently generated playlist from a daily-schedule script.
func (r *PlaylistRepository) GetPlaylistOfTheDay(ctx context.Context, currentUserID *uuid.UUID) (*models.PlaylistListItem, error) {
	query := `
		SELECT
			p.id, p.user_id, p.title, p.description, p.cover_url, p.visibility, p.share_token,
			p.view_count, p.share_count, p.like_count, p.follower_count, p.bookmark_count,
			p.is_curated, p.is_featured, p.display_order, p.script_id, p.slug,
			p.created_at, p.updated_at, p.deleted_at,
			COALESCE(COUNT(pi.id), 0) AS clip_count,
			EXISTS (
				SELECT 1
				FROM playlist_items pi2
				JOIN clips c2 ON pi2.clip_id = c2.id
				WHERE pi2.playlist_id = p.id
				  AND (c2.status = 'processing' OR (c2.stream_source = 'stream' AND c2.video_url IS NULL))
			) AS has_processing_clips
		FROM playlists p
		JOIN generated_playlists gp ON gp.playlist_id = p.id
		JOIN playlist_scripts ps ON ps.id = gp.script_id
		LEFT JOIN playlist_items pi ON p.id = pi.playlist_id
		WHERE ps.schedule = 'daily'
		  AND ps.is_active = true
		  AND p.visibility = 'public'
		  AND p.deleted_at IS NULL
		GROUP BY p.id, p.user_id, p.title, p.description, p.cover_url, p.visibility, p.share_token,
		         p.view_count, p.share_count, p.like_count, p.follower_count, p.bookmark_count,
		         p.is_curated, p.is_featured, p.display_order, p.script_id, p.slug,
		         p.created_at, p.updated_at, p.deleted_at
		ORDER BY gp.generated_at DESC
		LIMIT 1
	`

	var item models.PlaylistListItem
	err := r.pool.QueryRow(ctx, query).Scan(
		&item.ID,
		&item.UserID,
		&item.Title,
		&item.Description,
		&item.CoverURL,
		&item.Visibility,
		&item.ShareToken,
		&item.ViewCount,
		&item.ShareCount,
		&item.LikeCount,
		&item.FollowerCount,
		&item.BookmarkCount,
		&item.IsCurated,
		&item.IsFeatured,
		&item.DisplayOrder,
		&item.ScriptID,
		&item.Slug,
		&item.CreatedAt,
		&item.UpdatedAt,
		&item.DeletedAt,
		&item.ClipCount,
		&item.HasProcessingClips,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to get playlist of the day: %w", err)
	}

	if currentUserID != nil {
		if enrichErr := r.enrichPlaylistInteractionStates(ctx, *currentUserID, []*models.PlaylistListItem{&item}); enrichErr != nil {
			return nil, enrichErr
		}
	}

	// Fetch preview clips
	previewQuery := `
		SELECT c.id, c.twitch_clip_id, c.title, c.broadcaster_name, c.thumbnail_url,
		       c.duration, c.view_count, c.created_at
		FROM clips c
		INNER JOIN playlist_items pi ON c.id = pi.clip_id
		WHERE pi.playlist_id = $1
		ORDER BY pi.order_index ASC
		LIMIT 4
	`
	previewRows, err := r.pool.Query(ctx, previewQuery, item.ID)
	if err == nil {
		var previewClips []models.Clip
		for previewRows.Next() {
			var clip models.Clip
			scanErr := previewRows.Scan(
				&clip.ID,
				&clip.TwitchClipID,
				&clip.Title,
				&clip.BroadcasterName,
				&clip.ThumbnailURL,
				&clip.Duration,
				&clip.ViewCount,
				&clip.CreatedAt,
			)
			if scanErr == nil {
				previewClips = append(previewClips, clip)
			}
		}
		previewRows.Close()
		item.PreviewClips = previewClips
	}

	return &item, nil
}

// TrackShare records a playlist share event
func (r *PlaylistRepository) TrackShare(ctx context.Context, share *models.PlaylistShare) error {
	query := `
		INSERT INTO playlist_shares (id, playlist_id, platform, referrer, shared_at)
		VALUES ($1, $2, $3, $4, $5)
	`

	_, err := r.pool.Exec(ctx, query,
		share.ID,
		share.PlaylistID,
		share.Platform,
		share.Referrer,
		share.SharedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to track share: %w", err)
	}

	return nil
}
