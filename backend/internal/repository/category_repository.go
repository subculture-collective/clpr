package repository

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"git.subcult.tv/subculture-collective/clpr/internal/models"
)

// CategoryRepository handles database operations for categories
type CategoryRepository struct {
	pool *pgxpool.Pool
}

// NewCategoryRepository creates a new CategoryRepository
func NewCategoryRepository(pool *pgxpool.Pool) *CategoryRepository {
	return &CategoryRepository{
		pool: pool,
	}
}

// List retrieves categories ordered by position with optional filters
func (r *CategoryRepository) List(ctx context.Context, categoryType *string, featured *bool) ([]*models.Category, error) {
	baseQuery := `
		SELECT id, name, slug, description, icon, position,
		       category_type, is_featured, is_custom, created_by_user_id,
		       created_at, updated_at
		FROM categories
	`

	whereClause := "WHERE 1=1"
	args := []interface{}{}

	if categoryType != nil && *categoryType != "" {
		args = append(args, *categoryType)
		whereClause += fmt.Sprintf(" AND category_type = $%d", len(args))
	}
	if featured != nil {
		args = append(args, *featured)
		whereClause += fmt.Sprintf(" AND is_featured = $%d", len(args))
	}

	query := fmt.Sprintf("%s %s ORDER BY position ASC, name ASC", baseQuery, whereClause)

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list categories: %w", err)
	}
	defer rows.Close()

	var categories []*models.Category
	for rows.Next() {
		var category models.Category
		err := rows.Scan(
			&category.ID, &category.Name, &category.Slug, &category.Description,
			&category.Icon, &category.Position,
			&category.CategoryType, &category.IsFeatured, &category.IsCustom, &category.CreatedByUserID,
			&category.CreatedAt, &category.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan category: %w", err)
		}
		categories = append(categories, &category)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating categories: %w", err)
	}

	return categories, nil
}

// GetByID retrieves a category by its ID
func (r *CategoryRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Category, error) {
	query := `
		SELECT id, name, slug, description, icon, position,
		       category_type, is_featured, is_custom, created_by_user_id,
		       created_at, updated_at
		FROM categories
		WHERE id = $1
	`

	var category models.Category
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&category.ID, &category.Name, &category.Slug, &category.Description,
		&category.Icon, &category.Position,
		&category.CategoryType, &category.IsFeatured, &category.IsCustom, &category.CreatedByUserID,
		&category.CreatedAt, &category.UpdatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("category not found")
		}
		return nil, fmt.Errorf("failed to get category by ID: %w", err)
	}

	return &category, nil
}

// GetBySlug retrieves a category by its slug
func (r *CategoryRepository) GetBySlug(ctx context.Context, slug string) (*models.Category, error) {
	query := `
		SELECT id, name, slug, description, icon, position,
		       category_type, is_featured, is_custom, created_by_user_id,
		       created_at, updated_at
		FROM categories
		WHERE slug = $1
	`

	var category models.Category
	err := r.pool.QueryRow(ctx, query, slug).Scan(
		&category.ID, &category.Name, &category.Slug, &category.Description,
		&category.Icon, &category.Position,
		&category.CategoryType, &category.IsFeatured, &category.IsCustom, &category.CreatedByUserID,
		&category.CreatedAt, &category.UpdatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("category not found")
		}
		return nil, fmt.Errorf("failed to get category by slug: %w", err)
	}

	return &category, nil
}

// GetGamesInCategory retrieves games that belong to a specific category
func (r *CategoryRepository) GetGamesInCategory(ctx context.Context, categoryID uuid.UUID, userID *uuid.UUID, limit, offset int) ([]*models.GameWithStats, error) {
	query := `
		SELECT
			g.id, g.twitch_game_id, g.name, g.slug, g.box_art_url, g.igdb_id,
			g.created_at, g.updated_at,
			COALESCE(COUNT(DISTINCT c.id), 0) as clip_count,
			COALESCE(COUNT(DISTINCT gf.id), 0) as follower_count,
			BOOL_OR(ugf.id IS NOT NULL) as is_following
		FROM games g
		INNER JOIN category_games cg ON g.id = cg.game_id
		LEFT JOIN clips c ON c.game_id = g.twitch_game_id AND c.is_removed = false
		LEFT JOIN game_follows gf ON gf.game_id = g.id
		LEFT JOIN game_follows ugf ON ugf.game_id = g.id AND ugf.user_id = $2
		WHERE cg.category_id = $1
		GROUP BY g.id, g.twitch_game_id, g.name, g.slug, g.box_art_url, g.igdb_id, g.created_at, g.updated_at
		ORDER BY clip_count DESC, g.name ASC
		LIMIT $3 OFFSET $4
	`

	rows, err := r.pool.Query(ctx, query, categoryID, userID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get games in category: %w", err)
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
