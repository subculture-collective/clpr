package repository

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"git.subcult.tv/subculture-collective/clpr/internal/models"
	"git.subcult.tv/subculture-collective/clpr/pkg/utils"
)

// GameRepository handles database operations for games
type GameRepository struct {
	pool *pgxpool.Pool
}

// NewGameRepository creates a new GameRepository
func NewGameRepository(pool *pgxpool.Pool) *GameRepository {
	return &GameRepository{
		pool: pool,
	}
}

// Create inserts a new game into the database
func (r *GameRepository) Create(ctx context.Context, game *models.GameEntity) error {
	slug := utils.Slugify(game.Name)
	query := `
		INSERT INTO games (id, twitch_game_id, name, slug, box_art_url, igdb_id, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		ON CONFLICT (twitch_game_id) DO UPDATE SET
			name = EXCLUDED.name,
			slug = EXCLUDED.slug,
			box_art_url = EXCLUDED.box_art_url,
			igdb_id = EXCLUDED.igdb_id,
			updated_at = EXCLUDED.updated_at
		RETURNING id
	`

	err := r.pool.QueryRow(ctx, query,
		game.ID, game.TwitchGameID, game.Name, slug, game.BoxArtURL,
		game.IGDBID, game.CreatedAt, game.UpdatedAt,
	).Scan(&game.ID)

	if err != nil {
		return fmt.Errorf("failed to create game: %w", err)
	}

	return nil
}

// GetByID retrieves a game by its internal ID
func (r *GameRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.GameEntity, error) {
	query := `
		SELECT id, twitch_game_id, name, slug, box_art_url, igdb_id, created_at, updated_at
		FROM games
		WHERE id = $1
	`

	var game models.GameEntity
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&game.ID, &game.TwitchGameID, &game.Name, &game.Slug, &game.BoxArtURL,
		&game.IGDBID, &game.CreatedAt, &game.UpdatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("game not found")
		}
		return nil, fmt.Errorf("failed to get game by ID: %w", err)
	}

	return &game, nil
}

// GetByTwitchGameID retrieves a game by its Twitch game ID
func (r *GameRepository) GetByTwitchGameID(ctx context.Context, twitchGameID string) (*models.GameEntity, error) {
	query := `
		SELECT id, twitch_game_id, name, slug, box_art_url, igdb_id, created_at, updated_at
		FROM games
		WHERE twitch_game_id = $1
	`

	var game models.GameEntity
	err := r.pool.QueryRow(ctx, query, twitchGameID).Scan(
		&game.ID, &game.TwitchGameID, &game.Name, &game.Slug, &game.BoxArtURL,
		&game.IGDBID, &game.CreatedAt, &game.UpdatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("game not found")
		}
		return nil, fmt.Errorf("failed to get game by Twitch ID: %w", err)
	}

	return &game, nil
}

// GetWithStats retrieves a game with statistics
func (r *GameRepository) GetWithStats(ctx context.Context, gameID uuid.UUID, userID *uuid.UUID) (*models.GameWithStats, error) {
	query := `
		SELECT
			g.id, g.twitch_game_id, g.name, g.slug, g.box_art_url, g.igdb_id,
			g.created_at, g.updated_at,
			COALESCE(COUNT(DISTINCT c.id), 0) as clip_count,
			COALESCE(COUNT(DISTINCT gf.id), 0) as follower_count,
			BOOL_OR(ugf.id IS NOT NULL) as is_following
		FROM games g
		LEFT JOIN clips c ON c.game_id = g.twitch_game_id AND c.is_removed = false
		LEFT JOIN game_follows gf ON gf.game_id = g.id
		LEFT JOIN game_follows ugf ON ugf.game_id = g.id AND ugf.user_id = $2
		WHERE g.id = $1
		GROUP BY g.id, g.twitch_game_id, g.name, g.slug, g.box_art_url, g.igdb_id, g.created_at, g.updated_at
	`

	var game models.GameWithStats
	err := r.pool.QueryRow(ctx, query, gameID, userID).Scan(
		&game.ID, &game.TwitchGameID, &game.Name, &game.Slug, &game.BoxArtURL, &game.IGDBID,
		&game.CreatedAt, &game.UpdatedAt,
		&game.ClipCount, &game.FollowerCount, &game.IsFollowing,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("game not found")
		}
		return nil, fmt.Errorf("failed to get game with stats: %w", err)
	}

	return &game, nil
}

// GetTrending retrieves trending games (most clips in last 7 days)
func (r *GameRepository) GetTrending(ctx context.Context, limit, offset int) ([]*models.TrendingGame, error) {
	query := `
		SELECT 
			id, twitch_game_id, name, box_art_url, 
			recent_clip_count, total_vote_score, follower_count
		FROM trending_games
		LIMIT $1 OFFSET $2
	`

	rows, err := r.pool.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get trending games: %w", err)
	}
	defer rows.Close()

	var games []*models.TrendingGame
	for rows.Next() {
		var game models.TrendingGame
		err := rows.Scan(
			&game.ID, &game.TwitchGameID, &game.Name, &game.BoxArtURL,
			&game.RecentClipCount, &game.TotalVoteScore, &game.FollowerCount,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan trending game: %w", err)
		}
		games = append(games, &game)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating trending games: %w", err)
	}

	return games, nil
}

// FollowGame creates a game follow relationship
func (r *GameRepository) FollowGame(ctx context.Context, userID, gameID uuid.UUID) error {
	query := `
		INSERT INTO game_follows (id, user_id, game_id, followed_at)
		VALUES ($1, $2, $3, NOW())
		ON CONFLICT (user_id, game_id) DO NOTHING
	`

	_, err := r.pool.Exec(ctx, query, uuid.New(), userID, gameID)
	if err != nil {
		return fmt.Errorf("failed to follow game: %w", err)
	}

	return nil
}

// UnfollowGame removes a game follow relationship
func (r *GameRepository) UnfollowGame(ctx context.Context, userID, gameID uuid.UUID) error {
	query := `
		DELETE FROM game_follows
		WHERE user_id = $1 AND game_id = $2
	`

	_, err := r.pool.Exec(ctx, query, userID, gameID)
	if err != nil {
		return fmt.Errorf("failed to unfollow game: %w", err)
	}

	return nil
}

// IsFollowing checks if a user is following a game
func (r *GameRepository) IsFollowing(ctx context.Context, userID, gameID uuid.UUID) (bool, error) {
	query := `
		SELECT EXISTS(
			SELECT 1 FROM game_follows
			WHERE user_id = $1 AND game_id = $2
		)
	`

	var exists bool
	err := r.pool.QueryRow(ctx, query, userID, gameID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check game follow status: %w", err)
	}

	return exists, nil
}

// GetFollowedGames retrieves games followed by a user
func (r *GameRepository) GetFollowedGames(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*models.GameWithStats, error) {
	query := `
		SELECT
			g.id, g.twitch_game_id, g.name, g.slug, g.box_art_url, g.igdb_id,
			g.created_at, g.updated_at,
			COALESCE(COUNT(DISTINCT c.id), 0) as clip_count,
			COALESCE(COUNT(DISTINCT gf.id), 0) as follower_count,
			true as is_following
		FROM games g
		INNER JOIN game_follows ugf ON ugf.game_id = g.id AND ugf.user_id = $1
		LEFT JOIN clips c ON c.game_id = g.twitch_game_id AND c.is_removed = false
		LEFT JOIN game_follows gf ON gf.game_id = g.id
		GROUP BY g.id, g.twitch_game_id, g.name, g.slug, g.box_art_url, g.igdb_id, g.created_at, g.updated_at
		ORDER BY ugf.followed_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.pool.Query(ctx, query, userID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get followed games: %w", err)
	}
	defer rows.Close()

	var games []*models.GameWithStats
	for rows.Next() {
		var game models.GameWithStats
		err := rows.Scan(
			&game.ID, &game.TwitchGameID, &game.Name, &game.Slug, &game.BoxArtURL, &game.IGDBID,
			&game.CreatedAt, &game.UpdatedAt,
			&game.ClipCount, &game.FollowerCount, &game.IsFollowing,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan followed game: %w", err)
		}
		games = append(games, &game)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating followed games: %w", err)
	}

	return games, nil
}

// GetBySlug finds a game by a URL slug using the indexed slug column.
func (r *GameRepository) GetBySlug(ctx context.Context, slug string) (*models.GameEntity, error) {
	query := `
		SELECT id, twitch_game_id, name, slug, box_art_url, igdb_id, created_at, updated_at
		FROM games
		WHERE slug = $1
		LIMIT 1
	`

	var game models.GameEntity
	err := r.pool.QueryRow(ctx, query, slug).Scan(
		&game.ID, &game.TwitchGameID, &game.Name, &game.Slug, &game.BoxArtURL,
		&game.IGDBID, &game.CreatedAt, &game.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("game not found")
		}
		return nil, fmt.Errorf("failed to get game by slug: %w", err)
	}
	return &game, nil
}

// ListAllWithClipCounts returns games that have at least 10 clips, ordered by clip count.
func (r *GameRepository) ListAllWithClipCounts(ctx context.Context, limit, offset int) ([]*models.GameWithStats, error) {
	query := `
		SELECT
			g.id, g.twitch_game_id, g.name, g.slug, g.box_art_url, g.igdb_id,
			g.created_at, g.updated_at,
			COUNT(DISTINCT c.id) as clip_count,
			0 as follower_count,
			false as is_following
		FROM games g
		INNER JOIN clips c ON c.game_id = g.twitch_game_id AND c.is_removed = false
		GROUP BY g.id, g.twitch_game_id, g.name, g.slug, g.box_art_url, g.igdb_id, g.created_at, g.updated_at
		HAVING COUNT(DISTINCT c.id) >= 10
		ORDER BY clip_count DESC
		LIMIT $1 OFFSET $2
	`

	rows, err := r.pool.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list games with clip counts: %w", err)
	}
	defer rows.Close()

	var games []*models.GameWithStats
	for rows.Next() {
		var game models.GameWithStats
		err := rows.Scan(
			&game.ID, &game.TwitchGameID, &game.Name, &game.Slug, &game.BoxArtURL, &game.IGDBID,
			&game.CreatedAt, &game.UpdatedAt,
			&game.ClipCount, &game.FollowerCount, &game.IsFollowing,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan game: %w", err)
		}
		games = append(games, &game)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating games: %w", err)
	}
	return games, nil
}

// ListTopBroadcastersForGame returns the top broadcasters by clip count for a given game.
func (r *GameRepository) ListTopBroadcastersForGame(ctx context.Context, gameID string, limit int) ([]models.BroadcasterWithClipCount, error) {
	query := `
		SELECT broadcaster_id, broadcaster_name, COUNT(*) as clip_count, COALESCE(SUM(view_count), 0) as total_views
		FROM clips
		WHERE game_id = $1 AND is_removed = false AND broadcaster_id IS NOT NULL
		GROUP BY broadcaster_id, broadcaster_name
		ORDER BY clip_count DESC
		LIMIT $2
	`
	rows, err := r.pool.Query(ctx, query, gameID, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to list top broadcasters for game: %w", err)
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
