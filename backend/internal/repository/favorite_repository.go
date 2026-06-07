package repository

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"git.subcult.tv/subculture-collective/clpr/internal/models"
)

// FavoriteRepository handles database operations for favorites
type FavoriteRepository struct {
	pool *pgxpool.Pool
}

// NewFavoriteRepository creates a new FavoriteRepository
func NewFavoriteRepository(pool *pgxpool.Pool) *FavoriteRepository {
	return &FavoriteRepository{
		pool: pool,
	}
}

// Create adds a clip to favorites
func (r *FavoriteRepository) Create(ctx context.Context, userID, clipID uuid.UUID) error {
	query := `
		INSERT INTO favorites (user_id, clip_id)
		VALUES ($1, $2)
		ON CONFLICT (user_id, clip_id) DO NOTHING
	`

	_, err := r.pool.Exec(ctx, query, userID, clipID)
	if err != nil {
		return fmt.Errorf("failed to create favorite: %w", err)
	}

	return nil
}

// Delete removes a clip from favorites
func (r *FavoriteRepository) Delete(ctx context.Context, userID, clipID uuid.UUID) error {
	query := `DELETE FROM favorites WHERE user_id = $1 AND clip_id = $2`

	_, err := r.pool.Exec(ctx, query, userID, clipID)
	if err != nil {
		return fmt.Errorf("failed to delete favorite: %w", err)
	}

	return nil
}

// IsFavorited checks if a user has favorited a clip
func (r *FavoriteRepository) IsFavorited(ctx context.Context, userID, clipID uuid.UUID) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM favorites WHERE user_id = $1 AND clip_id = $2)`

	var exists bool
	err := r.pool.QueryRow(ctx, query, userID, clipID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check favorite status: %w", err)
	}

	return exists, nil
}

// GetByUserID retrieves all favorites for a user
func (r *FavoriteRepository) GetByUserID(ctx context.Context, userID uuid.UUID, limit, offset int) ([]models.Favorite, error) {
	query := `
		SELECT id, user_id, clip_id, created_at
		FROM favorites
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.pool.Query(ctx, query, userID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get favorites: %w", err)
	}
	defer rows.Close()

	var favorites []models.Favorite
	for rows.Next() {
		var fav models.Favorite
		err := rows.Scan(&fav.ID, &fav.UserID, &fav.ClipID, &fav.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan favorite: %w", err)
		}
		favorites = append(favorites, fav)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating favorites: %w", err)
	}

	return favorites, nil
}

// GetByID retrieves a favorite by ID
func (r *FavoriteRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Favorite, error) {
	query := `
		SELECT id, user_id, clip_id, created_at
		FROM favorites
		WHERE id = $1
	`

	var fav models.Favorite
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&fav.ID, &fav.UserID, &fav.ClipID, &fav.CreatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get favorite: %w", err)
	}

	return &fav, nil
}

// GetByClipID retrieves all favorites for a clip
func (r *FavoriteRepository) GetByClipID(ctx context.Context, clipID uuid.UUID) ([]models.Favorite, error) {
	query := `
		SELECT id, user_id, clip_id, created_at
		FROM favorites
		WHERE clip_id = $1
		ORDER BY created_at DESC
	`

	rows, err := r.pool.Query(ctx, query, clipID)
	if err != nil {
		return nil, fmt.Errorf("failed to get favorites by clip: %w", err)
	}
	defer rows.Close()

	var favorites []models.Favorite
	for rows.Next() {
		var fav models.Favorite
		err := rows.Scan(&fav.ID, &fav.UserID, &fav.ClipID, &fav.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan favorite: %w", err)
		}
		favorites = append(favorites, fav)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating favorites: %w", err)
	}

	return favorites, nil
}

// CountByUserID returns the total count of favorites for a user
func (r *FavoriteRepository) CountByUserID(ctx context.Context, userID uuid.UUID) (int, error) {
	query := `SELECT COUNT(*) FROM favorites WHERE user_id = $1`

	var count int
	err := r.pool.QueryRow(ctx, query, userID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count favorites: %w", err)
	}

	return count, nil
}

// GetClipsByUserID retrieves clips that are favorited by a user with sorting support
func (r *FavoriteRepository) GetClipsByUserID(ctx context.Context, userID uuid.UUID, sort string, limit, offset int) ([]models.Clip, int, error) {
	// First get the total count
	countQuery := `
		SELECT COUNT(*)
		FROM favorites f
		INNER JOIN clips c ON f.clip_id = c.id
		WHERE f.user_id = $1 AND c.is_removed = false
	`

	var total int
	err := r.pool.QueryRow(ctx, countQuery, userID).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count favorite clips: %w", err)
	}

	// Build ORDER BY clause based on sort parameter using a map for safety
	orderByClauses := map[string]string{
		"top":       "ORDER BY c.vote_score DESC",
		"discussed": "ORDER BY c.comment_count DESC",
		"newest":    "ORDER BY f.created_at DESC",
	}

	orderBy, ok := orderByClauses[sort]
	if !ok {
		orderBy = orderByClauses["newest"] // default to newest
	}

	// Query to get clips with favorites - using parameterized orderBy from map
	query := `
		SELECT 
			c.id, c.twitch_clip_id, c.twitch_clip_url, c.embed_url, c.title,
			c.creator_name, c.creator_id, c.broadcaster_name, c.broadcaster_id,
			c.game_id, c.game_name, c.language, c.thumbnail_url, c.duration,
			c.view_count, c.created_at, c.imported_at, c.vote_score, c.comment_count,
			c.favorite_count, c.is_featured, c.is_nsfw, c.is_removed, c.removed_reason
		FROM favorites f
		INNER JOIN clips c ON f.clip_id = c.id
		WHERE f.user_id = $1 AND c.is_removed = false
		` + orderBy + `
		LIMIT $2 OFFSET $3
	`

	rows, err := r.pool.Query(ctx, query, userID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get favorite clips: %w", err)
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
