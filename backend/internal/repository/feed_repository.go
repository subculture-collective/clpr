package repository

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"git.subcult.tv/subculture-collective/clpr/internal/models"
)

type FeedRepository struct {
	pool *pgxpool.Pool
}

func NewFeedRepository(pool *pgxpool.Pool) *FeedRepository {
	return &FeedRepository{pool: pool}
}

// CreateFeed creates a new feed
func (r *FeedRepository) CreateFeed(ctx context.Context, feed *models.Feed) error {
	query := `
		INSERT INTO feeds (id, user_id, name, description, icon, is_public, follower_count, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id, created_at, updated_at
	`
	return r.pool.QueryRow(ctx, query,
		feed.ID, feed.UserID, feed.Name, feed.Description, feed.Icon,
		feed.IsPublic, feed.FollowerCount, feed.CreatedAt, feed.UpdatedAt,
	).Scan(&feed.ID, &feed.CreatedAt, &feed.UpdatedAt)
}

// GetFeedByID retrieves a feed by ID
func (r *FeedRepository) GetFeedByID(ctx context.Context, feedID uuid.UUID) (*models.Feed, error) {
	query := `
		SELECT id, user_id, name, description, icon, is_public, follower_count, created_at, updated_at
		FROM feeds
		WHERE id = $1
	`
	feed := &models.Feed{}
	err := r.pool.QueryRow(ctx, query, feedID).Scan(
		&feed.ID, &feed.UserID, &feed.Name, &feed.Description, &feed.Icon,
		&feed.IsPublic, &feed.FollowerCount, &feed.CreatedAt, &feed.UpdatedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, fmt.Errorf("feed not found")
	}
	return feed, err
}

// GetFeedsByUserID retrieves all feeds for a user
func (r *FeedRepository) GetFeedsByUserID(ctx context.Context, userID uuid.UUID, includePrivate bool) ([]*models.Feed, error) {
	query := `
		SELECT id, user_id, name, description, icon, is_public, follower_count, created_at, updated_at
		FROM feeds
		WHERE user_id = $1
	`
	if !includePrivate {
		query += " AND is_public = true"
	}
	query += " ORDER BY created_at DESC"

	rows, err := r.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	feeds := []*models.Feed{}
	for rows.Next() {
		feed := &models.Feed{}
		err := rows.Scan(
			&feed.ID, &feed.UserID, &feed.Name, &feed.Description, &feed.Icon,
			&feed.IsPublic, &feed.FollowerCount, &feed.CreatedAt, &feed.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		feeds = append(feeds, feed)
	}
	return feeds, rows.Err()
}

// UpdateFeed updates a feed
func (r *FeedRepository) UpdateFeed(ctx context.Context, feed *models.Feed) error {
	query := `
		UPDATE feeds
		SET name = $2, description = $3, icon = $4, is_public = $5, updated_at = NOW()
		WHERE id = $1
		RETURNING updated_at
	`
	return r.pool.QueryRow(ctx, query,
		feed.ID, feed.Name, feed.Description, feed.Icon, feed.IsPublic,
	).Scan(&feed.UpdatedAt)
}

// DeleteFeed deletes a feed
func (r *FeedRepository) DeleteFeed(ctx context.Context, feedID uuid.UUID) error {
	query := `DELETE FROM feeds WHERE id = $1`
	_, err := r.pool.Exec(ctx, query, feedID)
	return err
}

// AddClipToFeed adds a clip to a feed
func (r *FeedRepository) AddClipToFeed(ctx context.Context, feedItem *models.FeedItem) error {
	// Get the next position
	var maxPosition *int
	err := r.pool.QueryRow(ctx, `SELECT MAX(position) FROM feed_items WHERE feed_id = $1`, feedItem.FeedID).Scan(&maxPosition)
	if err != nil {
		return err
	}

	position := 0
	if maxPosition != nil {
		position = *maxPosition + 1
	}
	feedItem.Position = position

	query := `
		INSERT INTO feed_items (id, feed_id, clip_id, position, added_at)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (feed_id, clip_id) DO UPDATE SET position = EXCLUDED.position
		RETURNING id, position, added_at
	`
	return r.pool.QueryRow(ctx, query,
		feedItem.ID, feedItem.FeedID, feedItem.ClipID, feedItem.Position, feedItem.AddedAt,
	).Scan(&feedItem.ID, &feedItem.Position, &feedItem.AddedAt)
}

// RemoveClipFromFeed removes a clip from a feed
func (r *FeedRepository) RemoveClipFromFeed(ctx context.Context, feedID, clipID uuid.UUID) error {
	query := `DELETE FROM feed_items WHERE feed_id = $1 AND clip_id = $2`
	_, err := r.pool.Exec(ctx, query, feedID, clipID)
	return err
}

// GetFeedClips retrieves all clips in a feed
func (r *FeedRepository) GetFeedClips(ctx context.Context, feedID uuid.UUID) ([]*models.FeedItemWithClip, error) {
	query := `
		SELECT 
			fi.id, fi.feed_id, fi.clip_id, fi.position, fi.added_at,
			c.id, c.twitch_clip_id, c.twitch_clip_url, c.embed_url, c.title,
			c.creator_name, c.creator_id, c.broadcaster_name, c.broadcaster_id,
			c.game_id, c.game_name, c.language, c.thumbnail_url, c.duration,
			c.view_count, c.created_at, c.imported_at, c.vote_score, c.comment_count,
			c.favorite_count, c.is_featured, c.is_nsfw, c.is_removed, c.removed_reason, c.is_hidden
		FROM feed_items fi
		JOIN clips c ON fi.clip_id = c.id
		WHERE fi.feed_id = $1
		ORDER BY fi.position ASC
	`
	rows, err := r.pool.Query(ctx, query, feedID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := []*models.FeedItemWithClip{}
	for rows.Next() {
		item := &models.FeedItemWithClip{
			Clip: &models.Clip{},
		}
		err := rows.Scan(
			&item.ID, &item.FeedID, &item.ClipID, &item.Position, &item.AddedAt,
			&item.Clip.ID, &item.Clip.TwitchClipID, &item.Clip.TwitchClipURL, &item.Clip.EmbedURL,
			&item.Clip.Title, &item.Clip.CreatorName, &item.Clip.CreatorID, &item.Clip.BroadcasterName,
			&item.Clip.BroadcasterID, &item.Clip.GameID, &item.Clip.GameName, &item.Clip.Language,
			&item.Clip.ThumbnailURL, &item.Clip.Duration, &item.Clip.ViewCount, &item.Clip.CreatedAt,
			&item.Clip.ImportedAt, &item.Clip.VoteScore, &item.Clip.CommentCount, &item.Clip.FavoriteCount,
			&item.Clip.IsFeatured, &item.Clip.IsNSFW, &item.Clip.IsRemoved, &item.Clip.RemovedReason, &item.Clip.IsHidden,
		)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

// ReorderFeedClips reorders clips in a feed
func (r *FeedRepository) ReorderFeedClips(ctx context.Context, feedID uuid.UUID, clipIDs []uuid.UUID) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	for i, clipID := range clipIDs {
		query := `UPDATE feed_items SET position = $1 WHERE feed_id = $2 AND clip_id = $3`
		_, err := tx.Exec(ctx, query, i, feedID, clipID)
		if err != nil {
			return err
		}
	}

	return tx.Commit(ctx)
}

// FollowFeed adds a follow relationship
func (r *FeedRepository) FollowFeed(ctx context.Context, feedFollow *models.FeedFollow) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	query := `
		INSERT INTO feed_follows (id, user_id, feed_id, followed_at)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (user_id, feed_id) DO NOTHING
		RETURNING id, followed_at
	`
	err = tx.QueryRow(ctx, query,
		feedFollow.ID, feedFollow.UserID, feedFollow.FeedID, feedFollow.FollowedAt,
	).Scan(&feedFollow.ID, &feedFollow.FollowedAt)

	// Only update follower count if a new row was inserted
	if err != pgx.ErrNoRows {
		if err != nil {
			return err
		}
		updateQuery := `UPDATE feeds SET follower_count = follower_count + 1 WHERE id = $1`
		_, err = tx.Exec(ctx, updateQuery, feedFollow.FeedID)
		if err != nil {
			return err
		}
	}

	return tx.Commit(ctx)
}

// UnfollowFeed removes a follow relationship
func (r *FeedRepository) UnfollowFeed(ctx context.Context, userID, feedID uuid.UUID) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	query := `DELETE FROM feed_follows WHERE user_id = $1 AND feed_id = $2`
	result, err := tx.Exec(ctx, query, userID, feedID)
	if err != nil {
		return err
	}

	rowsAffected := result.RowsAffected()
	if rowsAffected > 0 {
		// Update follower count only if a row was deleted
		updateQuery := `UPDATE feeds SET follower_count = GREATEST(0, follower_count - 1) WHERE id = $1`
		_, err = tx.Exec(ctx, updateQuery, feedID)
		if err != nil {
			return err
		}
	}

	return tx.Commit(ctx)
}

// IsFollowingFeed checks if a user is following a feed
func (r *FeedRepository) IsFollowingFeed(ctx context.Context, userID, feedID uuid.UUID) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM feed_follows WHERE user_id = $1 AND feed_id = $2)`
	var exists bool
	err := r.pool.QueryRow(ctx, query, userID, feedID).Scan(&exists)
	return exists, err
}

// GetFollowedFeeds retrieves all feeds a user is following
func (r *FeedRepository) GetFollowedFeeds(ctx context.Context, userID uuid.UUID) ([]*models.Feed, error) {
	query := `
		SELECT f.id, f.user_id, f.name, f.description, f.icon, f.is_public, f.follower_count, f.created_at, f.updated_at
		FROM feeds f
		JOIN feed_follows ff ON f.id = ff.feed_id
		WHERE ff.user_id = $1
		ORDER BY ff.followed_at DESC
	`
	rows, err := r.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	feeds := []*models.Feed{}
	for rows.Next() {
		feed := &models.Feed{}
		err := rows.Scan(
			&feed.ID, &feed.UserID, &feed.Name, &feed.Description, &feed.Icon,
			&feed.IsPublic, &feed.FollowerCount, &feed.CreatedAt, &feed.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		feeds = append(feeds, feed)
	}
	return feeds, rows.Err()
}

// DiscoverPublicFeeds retrieves public feeds for discovery
func (r *FeedRepository) DiscoverPublicFeeds(ctx context.Context, limit, offset int) ([]*models.FeedWithOwner, error) {
	query := `
		SELECT 
			f.id, f.user_id, f.name, f.description, f.icon, f.is_public, f.follower_count, f.created_at, f.updated_at,
			u.id, u.username, u.display_name, u.avatar_url
		FROM feeds f
		JOIN users u ON f.user_id = u.id
		WHERE f.is_public = true
		ORDER BY f.follower_count DESC, f.created_at DESC
		LIMIT $1 OFFSET $2
	`
	rows, err := r.pool.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	feeds := []*models.FeedWithOwner{}
	for rows.Next() {
		feed := &models.FeedWithOwner{
			Owner: &models.User{},
		}
		err := rows.Scan(
			&feed.ID, &feed.UserID, &feed.Name, &feed.Description, &feed.Icon,
			&feed.IsPublic, &feed.FollowerCount, &feed.CreatedAt, &feed.UpdatedAt,
			&feed.Owner.ID, &feed.Owner.Username, &feed.Owner.DisplayName, &feed.Owner.AvatarURL,
		)
		if err != nil {
			return nil, err
		}
		feeds = append(feeds, feed)
	}
	return feeds, rows.Err()
}

// SearchFeeds searches for public feeds by name
func (r *FeedRepository) SearchFeeds(ctx context.Context, query string, limit, offset int) ([]*models.FeedWithOwner, error) {
	searchQuery := `
		SELECT 
			f.id, f.user_id, f.name, f.description, f.icon, f.is_public, f.follower_count, f.created_at, f.updated_at,
			u.id, u.username, u.display_name, u.avatar_url
		FROM feeds f
		JOIN users u ON f.user_id = u.id
		WHERE f.is_public = true AND (f.name ILIKE $1 OR f.description ILIKE $1)
		ORDER BY f.follower_count DESC, f.created_at DESC
		LIMIT $2 OFFSET $3
	`
	searchTerm := "%" + query + "%"
	rows, err := r.pool.Query(ctx, searchQuery, searchTerm, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	feeds := []*models.FeedWithOwner{}
	for rows.Next() {
		feed := &models.FeedWithOwner{
			Owner: &models.User{},
		}
		err := rows.Scan(
			&feed.ID, &feed.UserID, &feed.Name, &feed.Description, &feed.Icon,
			&feed.IsPublic, &feed.FollowerCount, &feed.CreatedAt, &feed.UpdatedAt,
			&feed.Owner.ID, &feed.Owner.Username, &feed.Owner.DisplayName, &feed.Owner.AvatarURL,
		)
		if err != nil {
			return nil, err
		}
		feeds = append(feeds, feed)
	}
	return feeds, rows.Err()
}
