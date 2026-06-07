package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"git.subcult.tv/subculture-collective/clpr/internal/models"
)

// Sentinel errors for discovery list operations
var (
	// ErrDiscoveryListNotFound is returned when a discovery list is not found
	ErrDiscoveryListNotFound = errors.New("discovery list not found")
	// ErrClipNotFoundInList is returned when a clip is not found in a discovery list
	ErrClipNotFoundInList = errors.New("clip not found in list")
)

// DiscoveryListRepository handles database operations for discovery lists
type DiscoveryListRepository struct {
	db *pgxpool.Pool
}

// NewDiscoveryListRepository creates a new repository instance
func NewDiscoveryListRepository(db *pgxpool.Pool) *DiscoveryListRepository {
	return &DiscoveryListRepository{db: db}
}

// ListDiscoveryLists retrieves discovery lists with optional filtering
func (r *DiscoveryListRepository) ListDiscoveryLists(ctx context.Context, featuredOnly bool, userID *uuid.UUID, limit, offset int) ([]models.DiscoveryListWithStats, error) {
	query := `
		WITH list_stats AS (
			SELECT
				p.id,
				p.title as name,
				p.slug,
				p.description,
				p.is_featured,
				p.visibility = 'public' as is_active,
				p.display_order,
				p.user_id as created_by,
				p.created_at,
				p.updated_at,
				COUNT(DISTINCT pi.clip_id) as clip_count,
				COALESCE(p.follower_count, 0) as follower_count
			FROM playlists p
			LEFT JOIN playlist_items pi ON p.id = pi.playlist_id
			WHERE p.is_curated = true AND p.deleted_at IS NULL
	`

	if featuredOnly {
		query += " AND p.is_featured = true"
	}

	query += `
			GROUP BY p.id
			ORDER BY p.display_order ASC, p.created_at DESC
			LIMIT $1 OFFSET $2
		)
		SELECT * FROM list_stats
	`

	rows, err := r.db.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list discovery lists: %w", err)
	}
	defer rows.Close()

	var lists []models.DiscoveryListWithStats
	for rows.Next() {
		var list models.DiscoveryListWithStats
		err := rows.Scan(
			&list.ID, &list.Name, &list.Slug, &list.Description,
			&list.IsFeatured, &list.IsActive, &list.DisplayOrder,
			&list.CreatedBy, &list.CreatedAt, &list.UpdatedAt,
			&list.ClipCount, &list.FollowerCount,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan discovery list: %w", err)
		}
		lists = append(lists, list)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating discovery lists: %w", err)
	}

	// Get user-specific data if user is authenticated
	if userID != nil && len(lists) > 0 {
		listIDs := make([]uuid.UUID, len(lists))
		for i, list := range lists {
			listIDs[i] = list.ID
		}

		// Get follows
		followQuery := `
			SELECT playlist_id FROM playlist_follows
			WHERE user_id = $1 AND playlist_id = ANY($2)
		`
		followRows, err := r.db.Query(ctx, followQuery, userID, listIDs)
		if err != nil {
			return nil, fmt.Errorf("failed to get followed lists: %w", err)
		}
		defer followRows.Close()

		followedMap := make(map[uuid.UUID]bool)
		for followRows.Next() {
			var listID uuid.UUID
			if err := followRows.Scan(&listID); err != nil {
				return nil, fmt.Errorf("failed to scan followed list ID: %w", err)
			}
			followedMap[listID] = true
		}

		if err := followRows.Err(); err != nil {
			return nil, fmt.Errorf("error iterating followed lists: %w", err)
		}

		// Get bookmarks
		bookmarkQuery := `
			SELECT playlist_id FROM playlist_bookmarks
			WHERE user_id = $1 AND playlist_id = ANY($2)
		`
		bookmarkRows, err := r.db.Query(ctx, bookmarkQuery, userID, listIDs)
		if err != nil {
			return nil, fmt.Errorf("failed to get bookmarked lists: %w", err)
		}
		defer bookmarkRows.Close()

		bookmarkedMap := make(map[uuid.UUID]bool)
		for bookmarkRows.Next() {
			var listID uuid.UUID
			if err := bookmarkRows.Scan(&listID); err != nil {
				return nil, fmt.Errorf("failed to scan bookmarked list ID: %w", err)
			}
			bookmarkedMap[listID] = true
		}

		if err := bookmarkRows.Err(); err != nil {
			return nil, fmt.Errorf("error iterating bookmarked lists: %w", err)
		}

		// Update lists with user data
		for i := range lists {
			lists[i].IsFollowing = followedMap[lists[i].ID]
			lists[i].IsBookmarked = bookmarkedMap[lists[i].ID]
		}
	}

	// Batch fetch preview clips for all lists
	if len(lists) > 0 {
		listIDs := make([]uuid.UUID, len(lists))
		for i, list := range lists {
			listIDs[i] = list.ID
		}

		// Fetch preview clips for all lists in a single query using window function
		previewQuery := `
			WITH ranked_clips AS (
				SELECT
					c.*,
					pi.playlist_id,
					ROW_NUMBER() OVER (PARTITION BY pi.playlist_id ORDER BY pi.order_index ASC, pi.added_at DESC) as rn
				FROM playlist_items pi
				INNER JOIN clips c ON pi.clip_id = c.id
				WHERE pi.playlist_id = ANY($1) AND c.is_removed = false
			)
			SELECT
				playlist_id, rn,
				id, twitch_clip_id, twitch_clip_url, embed_url, title,
				creator_name, creator_id, broadcaster_name, broadcaster_id,
				game_id, game_name, language, thumbnail_url, duration,
				view_count, created_at, imported_at, vote_score, comment_count,
				favorite_count, is_featured, is_nsfw, is_removed, removed_reason, is_hidden,
				submitted_by_user_id, submitted_at
			FROM ranked_clips
			WHERE rn <= 4
			ORDER BY playlist_id, rn
		`

		previewRows, err := r.db.Query(ctx, previewQuery, listIDs)
		if err != nil {
			// Don't fail the whole request if preview clips can't be fetched
			return lists, nil
		}
		defer previewRows.Close()

		// Group clips by list ID
		clipsByList := make(map[uuid.UUID][]models.Clip)
		for previewRows.Next() {
			var listID uuid.UUID
			var rn int
			var clip models.Clip
			err := previewRows.Scan(
				&listID, &rn,
				&clip.ID, &clip.TwitchClipID, &clip.TwitchClipURL, &clip.EmbedURL,
				&clip.Title, &clip.CreatorName, &clip.CreatorID, &clip.BroadcasterName,
				&clip.BroadcasterID, &clip.GameID, &clip.GameName, &clip.Language,
				&clip.ThumbnailURL, &clip.Duration, &clip.ViewCount, &clip.CreatedAt,
				&clip.ImportedAt, &clip.VoteScore, &clip.CommentCount, &clip.FavoriteCount,
				&clip.IsFeatured, &clip.IsNSFW, &clip.IsRemoved, &clip.RemovedReason, &clip.IsHidden,
				&clip.SubmittedByUserID, &clip.SubmittedAt,
			)
			if err != nil {
				// Don't fail the whole request if scanning preview clips fails
				return lists, nil
			}
			clipsByList[listID] = append(clipsByList[listID], clip)
		}

		// Assign preview clips to lists
		for i := range lists {
			if clips, ok := clipsByList[lists[i].ID]; ok {
				lists[i].PreviewClips = clips
			}
		}
	}

	return lists, nil
}

// GetDiscoveryList retrieves a specific discovery list by ID or slug
func (r *DiscoveryListRepository) GetDiscoveryList(ctx context.Context, idOrSlug string, userID *uuid.UUID) (*models.DiscoveryListWithStats, error) {
	query := `
		SELECT
			p.id,
			p.title as name,
			p.slug,
			p.description,
			p.is_featured,
			p.visibility = 'public' as is_active,
			p.display_order,
			p.user_id as created_by,
			p.created_at,
			p.updated_at,
			COUNT(DISTINCT pi.clip_id) as clip_count,
			COALESCE(p.follower_count, 0) as follower_count
		FROM playlists p
		LEFT JOIN playlist_items pi ON p.id = pi.playlist_id
		WHERE p.is_curated = true AND p.deleted_at IS NULL AND (p.id::text = $1 OR p.slug = $1)
		GROUP BY p.id
	`

	var list models.DiscoveryListWithStats
	err := r.db.QueryRow(ctx, query, idOrSlug).Scan(
		&list.ID, &list.Name, &list.Slug, &list.Description,
		&list.IsFeatured, &list.IsActive, &list.DisplayOrder,
		&list.CreatedBy, &list.CreatedAt, &list.UpdatedAt,
		&list.ClipCount, &list.FollowerCount,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrDiscoveryListNotFound
		}
		return nil, fmt.Errorf("failed to get discovery list: %w", err)
	}

	// Get user-specific data if user is authenticated
	if userID != nil {
		// Check if following
		var followCount int
		err = r.db.QueryRow(ctx,
			"SELECT COUNT(*) FROM playlist_follows WHERE user_id = $1 AND playlist_id = $2",
			userID, list.ID).Scan(&followCount)
		if err != nil {
			return nil, fmt.Errorf("failed to check follow status: %w", err)
		}
		list.IsFollowing = followCount > 0

		// Check if bookmarked
		var bookmarkCount int
		err = r.db.QueryRow(ctx,
			"SELECT COUNT(*) FROM playlist_bookmarks WHERE user_id = $1 AND playlist_id = $2",
			userID, list.ID).Scan(&bookmarkCount)
		if err != nil {
			return nil, fmt.Errorf("failed to check bookmark status: %w", err)
		}
		list.IsBookmarked = bookmarkCount > 0
	}

	return &list, nil
}

// GetListClips retrieves clips from a discovery list
func (r *DiscoveryListRepository) GetListClips(ctx context.Context, listID uuid.UUID, userID *uuid.UUID, limit, offset int) ([]models.ClipWithSubmitter, int, error) {
	// First get the total count
	var totalCount int
	countQuery := `
		SELECT COUNT(*)
		FROM playlist_items pi
		INNER JOIN clips c ON pi.clip_id = c.id
		WHERE pi.playlist_id = $1 AND c.is_removed = false
	`
	err := r.db.QueryRow(ctx, countQuery, listID).Scan(&totalCount)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get list clips count: %w", err)
	}

	// Then get the clips with pagination
	query := `
		SELECT
			c.id, c.twitch_clip_id, c.twitch_clip_url, c.embed_url, c.title,
			c.creator_name, c.creator_id, c.broadcaster_name, c.broadcaster_id,
			c.game_id, c.game_name, c.language, c.thumbnail_url, c.duration,
			c.view_count, c.created_at, c.imported_at, c.vote_score, c.comment_count,
			c.favorite_count, c.is_featured, c.is_nsfw, c.is_removed, c.removed_reason, c.is_hidden,
			c.submitted_by_user_id, c.submitted_at,
			COALESCE(v.vote_type, 0) as user_vote
		FROM playlist_items pi
		INNER JOIN clips c ON pi.clip_id = c.id
		LEFT JOIN votes v ON c.id = v.clip_id AND v.user_id = $2
		WHERE pi.playlist_id = $1 AND c.is_removed = false
		ORDER BY pi.order_index ASC, pi.added_at DESC
		LIMIT $3 OFFSET $4
	`

	rows, err := r.db.Query(ctx, query, listID, userID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get list clips: %w", err)
	}
	defer rows.Close()

	var clips []models.ClipWithSubmitter
	for rows.Next() {
		var clip models.ClipWithSubmitter
		var userVote int
		err := rows.Scan(
			&clip.ID, &clip.TwitchClipID, &clip.TwitchClipURL, &clip.EmbedURL,
			&clip.Title, &clip.CreatorName, &clip.CreatorID, &clip.BroadcasterName,
			&clip.BroadcasterID, &clip.GameID, &clip.GameName, &clip.Language,
			&clip.ThumbnailURL, &clip.Duration, &clip.ViewCount, &clip.CreatedAt,
			&clip.ImportedAt, &clip.VoteScore, &clip.CommentCount, &clip.FavoriteCount,
			&clip.IsFeatured, &clip.IsNSFW, &clip.IsRemoved, &clip.RemovedReason, &clip.IsHidden,
			&clip.SubmittedByUserID, &clip.SubmittedAt,
			&userVote,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan clip: %w", err)
		}
		clips = append(clips, clip)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("error iterating clips: %w", err)
	}

	// Batch fetch submitter info for clips that have been submitted
	if len(clips) > 0 {
		submitterIDs := make(map[uuid.UUID]bool)
		for _, clip := range clips {
			if clip.SubmittedByUserID != nil {
				submitterIDs[*clip.SubmittedByUserID] = true
			}
		}

		// Fetch all submitters in a single query if there are any
		if len(submitterIDs) > 0 {
			ids := make([]uuid.UUID, 0, len(submitterIDs))
			for id := range submitterIDs {
				ids = append(ids, id)
			}

			submitterQuery := `SELECT id, username, display_name, avatar_url FROM users WHERE id = ANY($1)`
			submitterRows, err := r.db.Query(ctx, submitterQuery, ids)
			if err != nil {
				// Don't fail the whole request if submitters can't be fetched
				return clips, totalCount, nil
			}
			defer submitterRows.Close()

			// Map submitters by ID for O(1) lookup
			submitterMap := make(map[uuid.UUID]*models.ClipSubmitterInfo)
			for submitterRows.Next() {
				var submitter models.ClipSubmitterInfo
				if err := submitterRows.Scan(&submitter.ID, &submitter.Username, &submitter.DisplayName, &submitter.AvatarURL); err != nil {
					// Continue on error
					continue
				}
				submitterMap[submitter.ID] = &submitter
			}

			// Attach submitters to clips
			for i := range clips {
				if clips[i].SubmittedByUserID != nil {
					if submitter, ok := submitterMap[*clips[i].SubmittedByUserID]; ok {
						clips[i].SubmittedBy = submitter
					}
				}
			}
		}
	}

	return clips, totalCount, nil
}

// GetListClipCount returns the number of clips in a list
func (r *DiscoveryListRepository) GetListClipCount(ctx context.Context, listID uuid.UUID) (int, error) {
	query := `
		SELECT COUNT(*)
		FROM playlist_items pi
		INNER JOIN clips c ON pi.clip_id = c.id
		WHERE pi.playlist_id = $1 AND c.is_removed = false
	`

	var count int
	err := r.db.QueryRow(ctx, query, listID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to get list clips count: %w", err)
	}

	return count, nil
}

// GetListClipsForExport retrieves all clips from a list for export
func (r *DiscoveryListRepository) GetListClipsForExport(ctx context.Context, listID uuid.UUID, limit int) ([]models.ClipWithSubmitter, error) {
	query := `
		SELECT
			c.id, c.twitch_clip_id, c.twitch_clip_url, c.embed_url, c.title,
			c.creator_name, c.creator_id, c.broadcaster_name, c.broadcaster_id,
			c.game_id, c.game_name, c.language, c.thumbnail_url, c.duration,
			c.view_count, c.created_at, c.imported_at, c.vote_score, c.comment_count,
			c.favorite_count, c.is_featured, c.is_nsfw, c.is_removed, c.removed_reason, c.is_hidden,
			c.submitted_by_user_id, c.submitted_at
		FROM playlist_items pi
		INNER JOIN clips c ON pi.clip_id = c.id
		WHERE pi.playlist_id = $1 AND c.is_removed = false
		ORDER BY pi.order_index ASC, pi.added_at DESC
		LIMIT $2
	`

	rows, err := r.db.Query(ctx, query, listID, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get list clips for export: %w", err)
	}
	defer rows.Close()

	var clips []models.ClipWithSubmitter
	for rows.Next() {
		var clip models.ClipWithSubmitter
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
			return nil, fmt.Errorf("failed to scan clip: %w", err)
		}
		clips = append(clips, clip)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating clips: %w", err)
	}

	return clips, nil
}

// FollowList creates a follow relationship for a user and list
func (r *DiscoveryListRepository) FollowList(ctx context.Context, userID, listID uuid.UUID) error {
	query := `
		INSERT INTO playlist_follows (user_id, playlist_id)
		VALUES ($1, $2)
		ON CONFLICT (user_id, playlist_id) DO NOTHING
	`

	_, err := r.db.Exec(ctx, query, userID, listID)
	if err != nil {
		return fmt.Errorf("failed to follow list: %w", err)
	}

	return nil
}

// UnfollowList removes a follow relationship
func (r *DiscoveryListRepository) UnfollowList(ctx context.Context, userID, listID uuid.UUID) error {
	query := `
		DELETE FROM playlist_follows
		WHERE user_id = $1 AND playlist_id = $2
	`

	result, err := r.db.Exec(ctx, query, userID, listID)
	if err != nil {
		return fmt.Errorf("failed to unfollow list: %w", err)
	}

	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("not following this list")
	}

	return nil
}

// BookmarkList creates a bookmark for a user and list
func (r *DiscoveryListRepository) BookmarkList(ctx context.Context, userID, listID uuid.UUID) error {
	query := `
		INSERT INTO playlist_bookmarks (user_id, playlist_id)
		VALUES ($1, $2)
		ON CONFLICT (user_id, playlist_id) DO NOTHING
	`

	_, err := r.db.Exec(ctx, query, userID, listID)
	if err != nil {
		return fmt.Errorf("failed to bookmark list: %w", err)
	}

	return nil
}

// UnbookmarkList removes a bookmark
func (r *DiscoveryListRepository) UnbookmarkList(ctx context.Context, userID, listID uuid.UUID) error {
	query := `
		DELETE FROM playlist_bookmarks
		WHERE user_id = $1 AND playlist_id = $2
	`

	result, err := r.db.Exec(ctx, query, userID, listID)
	if err != nil {
		return fmt.Errorf("failed to unbookmark list: %w", err)
	}

	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("list not bookmarked")
	}

	return nil
}

// GetUserFollowedLists retrieves lists followed by a user
func (r *DiscoveryListRepository) GetUserFollowedLists(ctx context.Context, userID uuid.UUID, limit, offset int) ([]models.DiscoveryListWithStats, error) {
	query := `
		SELECT
			p.id,
			p.title as name,
			p.slug,
			p.description,
			p.is_featured,
			p.visibility = 'public' as is_active,
			p.display_order,
			p.user_id as created_by,
			p.created_at,
			p.updated_at,
			COUNT(DISTINCT pi.clip_id) as clip_count,
			COALESCE(p.follower_count, 0) as follower_count,
			true as is_following,
			EXISTS(SELECT 1 FROM playlist_bookmarks pb WHERE pb.user_id = $1 AND pb.playlist_id = p.id) as is_bookmarked
		FROM playlist_follows pf
		INNER JOIN playlists p ON pf.playlist_id = p.id
		LEFT JOIN playlist_items pi ON p.id = pi.playlist_id
		WHERE pf.user_id = $1 AND p.is_curated = true AND p.deleted_at IS NULL
		GROUP BY p.id, pf.followed_at
		ORDER BY pf.followed_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.Query(ctx, query, userID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get user followed lists: %w", err)
	}
	defer rows.Close()

	var lists []models.DiscoveryListWithStats
	for rows.Next() {
		var list models.DiscoveryListWithStats
		err := rows.Scan(
			&list.ID, &list.Name, &list.Slug, &list.Description,
			&list.IsFeatured, &list.IsActive, &list.DisplayOrder,
			&list.CreatedBy, &list.CreatedAt, &list.UpdatedAt,
			&list.ClipCount, &list.FollowerCount,
			&list.IsFollowing, &list.IsBookmarked,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan followed list: %w", err)
		}
		lists = append(lists, list)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating followed lists: %w", err)
	}

	// Batch fetch preview clips for all lists using a single query
	if len(lists) > 0 {
		listIDs := make([]uuid.UUID, len(lists))
		for i, list := range lists {
			listIDs[i] = list.ID
		}

		// Fetch preview clips for all lists in a single query using window function
		previewQuery := `
			WITH ranked_clips AS (
				SELECT
					c.id, c.twitch_clip_id, c.twitch_clip_url, c.embed_url, c.title,
					c.creator_name, c.creator_id, c.broadcaster_name, c.broadcaster_id,
					c.game_id, c.game_name, c.language, c.thumbnail_url, c.duration,
					c.view_count, c.created_at, c.imported_at, c.vote_score, c.comment_count,
					c.favorite_count, c.is_featured, c.is_nsfw, c.is_removed, c.removed_reason, c.is_hidden,
					c.submitted_by_user_id, c.submitted_at,
					pi.playlist_id,
					ROW_NUMBER() OVER (PARTITION BY pi.playlist_id ORDER BY pi.order_index ASC, pi.added_at DESC) as rn
				FROM playlist_items pi
				INNER JOIN clips c ON pi.clip_id = c.id
				WHERE pi.playlist_id = ANY($1) AND c.is_removed = false
			)
			SELECT
				playlist_id,
				id, twitch_clip_id, twitch_clip_url, embed_url, title,
				creator_name, creator_id, broadcaster_name, broadcaster_id,
				game_id, game_name, language, thumbnail_url, duration,
				view_count, created_at, imported_at, vote_score, comment_count,
				favorite_count, is_featured, is_nsfw, is_removed, removed_reason, is_hidden,
				submitted_by_user_id, submitted_at
			FROM ranked_clips
			WHERE rn <= 4
			ORDER BY playlist_id, rn
		`

		previewRows, err := r.db.Query(ctx, previewQuery, listIDs)
		if err != nil {
			// Don't fail the whole request if preview clips can't be fetched
			return lists, nil
		}
		defer previewRows.Close()

		// Group clips by list ID
		clipsByList := make(map[uuid.UUID][]models.Clip)
		for previewRows.Next() {
			var listID uuid.UUID
			var clip models.Clip
			err := previewRows.Scan(
				&listID,
				&clip.ID, &clip.TwitchClipID, &clip.TwitchClipURL, &clip.EmbedURL,
				&clip.Title, &clip.CreatorName, &clip.CreatorID, &clip.BroadcasterName,
				&clip.BroadcasterID, &clip.GameID, &clip.GameName, &clip.Language,
				&clip.ThumbnailURL, &clip.Duration, &clip.ViewCount, &clip.CreatedAt,
				&clip.ImportedAt, &clip.VoteScore, &clip.CommentCount, &clip.FavoriteCount,
				&clip.IsFeatured, &clip.IsNSFW, &clip.IsRemoved, &clip.RemovedReason, &clip.IsHidden,
				&clip.SubmittedByUserID, &clip.SubmittedAt,
			)
			if err != nil {
				// Don't fail the whole request if scanning preview clips fails
				return lists, nil
			}
			clipsByList[listID] = append(clipsByList[listID], clip)
		}

		// Assign preview clips to lists
		for i := range lists {
			if clips, ok := clipsByList[lists[i].ID]; ok {
				lists[i].PreviewClips = clips
			}
		}
	}

	return lists, nil
}

// CreateList creates a new discovery list
func (r *DiscoveryListRepository) CreateList(ctx context.Context, name, slug, description string, isFeatured bool, createdBy uuid.UUID) (*models.DiscoveryList, error) {
	query := `
		INSERT INTO playlists (user_id, title, slug, description, is_featured, is_curated, visibility)
		VALUES ($1, $2, $3, $4, $5, true, 'public')
		RETURNING id, user_id as created_by, title as name, slug, description, is_featured, created_at, updated_at
	`

	var list models.DiscoveryList
	err := r.db.QueryRow(ctx, query, createdBy, name, slug, description, isFeatured).Scan(
		&list.ID, &list.CreatedBy, &list.Name, &list.Slug, &list.Description,
		&list.IsFeatured, &list.CreatedAt, &list.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create discovery list: %w", err)
	}

	// Set IsActive to true since we're creating public playlists
	list.IsActive = true
	list.DisplayOrder = 0

	return &list, nil
}

// UpdateList updates an existing discovery list
func (r *DiscoveryListRepository) UpdateList(ctx context.Context, listID uuid.UUID, name, description *string, isFeatured *bool) (*models.DiscoveryList, error) {
	// Check if at least one field is being updated
	if name == nil && description == nil && isFeatured == nil {
		return nil, fmt.Errorf("at least one field must be provided for update")
	}

	// Build dynamic update query using a safer approach
	setClauses := []string{"updated_at = NOW()"}
	args := []interface{}{listID}
	argIdx := 2

	if name != nil {
		setClauses = append(setClauses, fmt.Sprintf("title = $%d", argIdx))
		args = append(args, *name)
		argIdx++
	}
	if description != nil {
		setClauses = append(setClauses, fmt.Sprintf("description = $%d", argIdx))
		args = append(args, *description)
		argIdx++
	}
	if isFeatured != nil {
		setClauses = append(setClauses, fmt.Sprintf("is_featured = $%d", argIdx))
		args = append(args, *isFeatured)
		argIdx++
	}

	// Join all SET clauses
	query := "UPDATE playlists SET " + setClauses[0]
	for i := 1; i < len(setClauses); i++ {
		query += ", " + setClauses[i]
	}
	query += " WHERE id = $1 AND is_curated = true RETURNING id, user_id as created_by, title as name, slug, description, is_featured, created_at, updated_at"

	var list models.DiscoveryList
	err := r.db.QueryRow(ctx, query, args...).Scan(
		&list.ID, &list.CreatedBy, &list.Name, &list.Slug, &list.Description,
		&list.IsFeatured, &list.CreatedAt, &list.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrDiscoveryListNotFound
		}
		return nil, fmt.Errorf("failed to update discovery list: %w", err)
	}

	list.IsActive = true
	list.DisplayOrder = 0

	return &list, nil
}

// DeleteList deletes a discovery list
func (r *DiscoveryListRepository) DeleteList(ctx context.Context, listID uuid.UUID) error {
	query := "UPDATE playlists SET deleted_at = NOW() WHERE id = $1 AND is_curated = true"

	result, err := r.db.Exec(ctx, query, listID)
	if err != nil {
		return fmt.Errorf("failed to delete discovery list: %w", err)
	}

	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		return ErrDiscoveryListNotFound
	}

	return nil
}

// AddClipToList adds a clip to a discovery list
func (r *DiscoveryListRepository) AddClipToList(ctx context.Context, listID, clipID uuid.UUID) error {
	// Get the current max display order
	var maxOrder int
	err := r.db.QueryRow(ctx,
		"SELECT COALESCE(MAX(order_index), -1) FROM playlist_items WHERE playlist_id = $1", listID).Scan(&maxOrder)
	if err != nil {
		return fmt.Errorf("failed to get max display order: %w", err)
	}

	// Insert the clip with the next display order
	query := `
		INSERT INTO playlist_items (playlist_id, clip_id, order_index)
		VALUES ($1, $2, $3)
		ON CONFLICT (playlist_id, clip_id) DO NOTHING
	`

	_, err = r.db.Exec(ctx, query, listID, clipID, maxOrder+1)
	if err != nil {
		return fmt.Errorf("failed to add clip to list: %w", err)
	}

	// Update the playlists updated_at timestamp
	_, err = r.db.Exec(ctx, "UPDATE playlists SET updated_at = NOW() WHERE id = $1", listID)
	if err != nil {
		return fmt.Errorf("failed to update list timestamp: %w", err)
	}

	return nil
}

// RemoveClipFromList removes a clip from a discovery list
func (r *DiscoveryListRepository) RemoveClipFromList(ctx context.Context, listID, clipID uuid.UUID) error {
	query := "DELETE FROM playlist_items WHERE playlist_id = $1 AND clip_id = $2"

	result, err := r.db.Exec(ctx, query, listID, clipID)
	if err != nil {
		return fmt.Errorf("failed to remove clip from list: %w", err)
	}

	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		return ErrClipNotFoundInList
	}

	// Update the playlists updated_at timestamp
	_, err = r.db.Exec(ctx, "UPDATE playlists SET updated_at = NOW() WHERE id = $1", listID)
	if err != nil {
		return fmt.Errorf("failed to update list timestamp: %w", err)
	}

	return nil
}

// ReorderClips reorders clips in a discovery list
func (r *DiscoveryListRepository) ReorderClips(ctx context.Context, listID uuid.UUID, clipIDs []uuid.UUID) error {
	// Start a transaction
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// Update display order for each clip
	for i, clipID := range clipIDs {
		query := "UPDATE playlist_items SET order_index = $1 WHERE playlist_id = $2 AND clip_id = $3"
		_, err := tx.Exec(ctx, query, i, listID, clipID)
		if err != nil {
			return fmt.Errorf("failed to update clip order: %w", err)
		}
	}

	// Update the playlists updated_at timestamp
	_, err = tx.Exec(ctx, "UPDATE playlists SET updated_at = NOW() WHERE id = $1", listID)
	if err != nil {
		return fmt.Errorf("failed to update list timestamp: %w", err)
	}

	// Commit the transaction
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// GetListClipsCount is an alias for GetListClipCount
func (r *DiscoveryListRepository) GetListClipsCount(ctx context.Context, listID uuid.UUID) (int, error) {
	return r.GetListClipCount(ctx, listID)
}

// ListAllDiscoveryLists retrieves all discovery lists
func (r *DiscoveryListRepository) ListAllDiscoveryLists(ctx context.Context, limit, offset int) ([]models.DiscoveryListWithStats, error) {
	query := `
		SELECT
			p.id,
			p.title as name,
			p.slug,
			p.description,
			p.is_featured,
			p.visibility = 'public' as is_active,
			p.display_order,
			p.user_id as created_by,
			p.created_at,
			p.updated_at,
			COUNT(DISTINCT pi.clip_id) as clip_count,
			COALESCE(p.follower_count, 0) as follower_count
		FROM playlists p
		LEFT JOIN playlist_items pi ON p.id = pi.playlist_id
		WHERE p.is_curated = true AND p.deleted_at IS NULL
		GROUP BY p.id
		ORDER BY p.created_at DESC
		LIMIT $1 OFFSET $2
	`

	rows, err := r.db.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list all discovery lists: %w", err)
	}
	defer rows.Close()

	var lists []models.DiscoveryListWithStats
	for rows.Next() {
		var list models.DiscoveryListWithStats
		err := rows.Scan(
			&list.ID, &list.Name, &list.Slug, &list.Description,
			&list.IsFeatured, &list.IsActive, &list.DisplayOrder,
			&list.CreatedBy, &list.CreatedAt, &list.UpdatedAt,
			&list.ClipCount, &list.FollowerCount,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan discovery list: %w", err)
		}
		lists = append(lists, list)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating discovery lists: %w", err)
	}

	return lists, nil
}

// CreateDiscoveryList is an alias for CreateList
func (r *DiscoveryListRepository) CreateDiscoveryList(ctx context.Context, name, slug, description string, isFeatured bool, createdBy uuid.UUID) (*models.DiscoveryList, error) {
	return r.CreateList(ctx, name, slug, description, isFeatured, createdBy)
}

// UpdateDiscoveryList is an alias for UpdateList
func (r *DiscoveryListRepository) UpdateDiscoveryList(ctx context.Context, listID uuid.UUID, name, description *string, isFeatured *bool) (*models.DiscoveryList, error) {
	return r.UpdateList(ctx, listID, name, description, isFeatured)
}

// DeleteDiscoveryList is an alias for DeleteList
func (r *DiscoveryListRepository) DeleteDiscoveryList(ctx context.Context, listID uuid.UUID) error {
	return r.DeleteList(ctx, listID)
}

// ReorderListClips is an alias for ReorderClips
func (r *DiscoveryListRepository) ReorderListClips(ctx context.Context, listID uuid.UUID, clipIDs []uuid.UUID) error {
	return r.ReorderClips(ctx, listID, clipIDs)
}
