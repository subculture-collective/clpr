package repository

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/subculture-collective/clipper/internal/models"
)

// TagRepository handles database operations for tags
type TagRepository struct {
	pool *pgxpool.Pool
}

// NewTagRepository creates a new TagRepository
func NewTagRepository(pool *pgxpool.Pool) *TagRepository {
	return &TagRepository{
		pool: pool,
	}
}

// Create inserts a new tag into the database
func (r *TagRepository) Create(ctx context.Context, tag *models.Tag) error {
	query := `
		INSERT INTO tags (id, name, slug, description, color, usage_count, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`

	_, err := r.pool.Exec(ctx, query,
		tag.ID, tag.Name, tag.Slug, tag.Description, tag.Color,
		tag.UsageCount, tag.CreatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create tag: %w", err)
	}

	return nil
}

// GetByID retrieves a tag by its ID
func (r *TagRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Tag, error) {
	query := `
		SELECT id, name, slug, description, color, usage_count, created_at
		FROM tags
		WHERE id = $1
	`

	var tag models.Tag
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&tag.ID, &tag.Name, &tag.Slug, &tag.Description,
		&tag.Color, &tag.UsageCount, &tag.CreatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("tag not found")
		}
		return nil, fmt.Errorf("failed to get tag by ID: %w", err)
	}

	return &tag, nil
}

// GetBySlug retrieves a tag by its slug
func (r *TagRepository) GetBySlug(ctx context.Context, slug string) (*models.Tag, error) {
	query := `
		SELECT id, name, slug, description, color, usage_count, created_at
		FROM tags
		WHERE slug = $1
	`

	var tag models.Tag
	err := r.pool.QueryRow(ctx, query, slug).Scan(
		&tag.ID, &tag.Name, &tag.Slug, &tag.Description,
		&tag.Color, &tag.UsageCount, &tag.CreatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("tag not found")
		}
		return nil, fmt.Errorf("failed to get tag by slug: %w", err)
	}

	return &tag, nil
}

// List retrieves tags with optional sorting and pagination
func (r *TagRepository) List(ctx context.Context, sort string, limit, offset int) ([]*models.Tag, error) {
	var query string
	switch sort {
	case "alphabetical":
		query = `
		SELECT id, name, slug, description, color, usage_count, created_at
		FROM tags
		WHERE slug NOT IN (SELECT LOWER(pattern) FROM blacklisted_tags)
		ORDER BY name ASC
		LIMIT $1 OFFSET $2
		`
	case "recent":
		query = `
		SELECT id, name, slug, description, color, usage_count, created_at
		FROM tags
		WHERE slug NOT IN (SELECT LOWER(pattern) FROM blacklisted_tags)
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
		`
	case "popularity":
		fallthrough
	default:
		query = `
		SELECT id, name, slug, description, color, usage_count, created_at
		FROM tags
		WHERE slug NOT IN (SELECT LOWER(pattern) FROM blacklisted_tags)
		ORDER BY usage_count DESC
		LIMIT $1 OFFSET $2
		`
	}

	rows, err := r.pool.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list tags: %w", err)
	}
	defer rows.Close()

	var tags []*models.Tag
	for rows.Next() {
		var tag models.Tag
		err := rows.Scan(
			&tag.ID, &tag.Name, &tag.Slug, &tag.Description,
			&tag.Color, &tag.UsageCount, &tag.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan tag: %w", err)
		}
		tags = append(tags, &tag)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating tags: %w", err)
	}

	return tags, nil
}

// Count returns the total number of tags
func (r *TagRepository) Count(ctx context.Context) (int, error) {
	var count int
	query := `SELECT COUNT(*) FROM tags`
	err := r.pool.QueryRow(ctx, query).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count tags: %w", err)
	}
	return count, nil
}

// Search searches for tags by name
func (r *TagRepository) Search(ctx context.Context, query string, limit int) ([]*models.Tag, error) {
	searchQuery := `
		SELECT id, name, slug, description, color, usage_count, created_at
		FROM tags
		WHERE (name ILIKE $1 OR slug ILIKE $1)
		AND slug NOT IN (SELECT LOWER(pattern) FROM blacklisted_tags)
		ORDER BY usage_count DESC
		LIMIT $2
	`

	rows, err := r.pool.Query(ctx, searchQuery, "%"+query+"%", limit)
	if err != nil {
		return nil, fmt.Errorf("failed to search tags: %w", err)
	}
	defer rows.Close()

	var tags []*models.Tag
	for rows.Next() {
		var tag models.Tag
		err := rows.Scan(
			&tag.ID, &tag.Name, &tag.Slug, &tag.Description,
			&tag.Color, &tag.UsageCount, &tag.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan tag: %w", err)
		}
		tags = append(tags, &tag)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating tags: %w", err)
	}

	return tags, nil
}

// Update updates an existing tag
func (r *TagRepository) Update(ctx context.Context, tag *models.Tag) error {
	query := `
		UPDATE tags
		SET name = $2, slug = $3, description = $4, color = $5
		WHERE id = $1
	`

	result, err := r.pool.Exec(ctx, query,
		tag.ID, tag.Name, tag.Slug, tag.Description, tag.Color,
	)

	if err != nil {
		return fmt.Errorf("failed to update tag: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("tag not found")
	}

	return nil
}

// Delete deletes a tag and its associations
func (r *TagRepository) Delete(ctx context.Context, id uuid.UUID) error {
	// Delete all clip-tag associations first
	_, err := r.pool.Exec(ctx, "DELETE FROM clip_tags WHERE tag_id = $1", id)
	if err != nil {
		return fmt.Errorf("failed to delete clip-tag associations: %w", err)
	}

	// Delete the tag
	query := `DELETE FROM tags WHERE id = $1`
	result, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete tag: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("tag not found")
	}

	return nil
}

// AddTagToClip associates a tag with a clip
func (r *TagRepository) AddTagToClip(ctx context.Context, clipID, tagID uuid.UUID) error {
	query := `
		INSERT INTO clip_tags (clip_id, tag_id, created_at)
		VALUES ($1, $2, NOW())
		ON CONFLICT (clip_id, tag_id) DO NOTHING
	`

	_, err := r.pool.Exec(ctx, query, clipID, tagID)
	if err != nil {
		return fmt.Errorf("failed to add tag to clip: %w", err)
	}

	return nil
}

// RemoveTagFromClip removes a tag association from a clip
func (r *TagRepository) RemoveTagFromClip(ctx context.Context, clipID, tagID uuid.UUID) error {
	query := `DELETE FROM clip_tags WHERE clip_id = $1 AND tag_id = $2`
	result, err := r.pool.Exec(ctx, query, clipID, tagID)
	if err != nil {
		return fmt.Errorf("failed to remove tag from clip: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("tag association not found")
	}

	return nil
}

// GetClipTags retrieves all tags for a clip
func (r *TagRepository) GetClipTags(ctx context.Context, clipID uuid.UUID) ([]*models.Tag, error) {
	query := `
		SELECT t.id, t.name, t.slug, t.description, t.color, t.usage_count, t.created_at
		FROM tags t
		INNER JOIN clip_tags ct ON t.id = ct.tag_id
		WHERE ct.clip_id = $1
		ORDER BY t.name ASC
	`

	rows, err := r.pool.Query(ctx, query, clipID)
	if err != nil {
		return nil, fmt.Errorf("failed to get clip tags: %w", err)
	}
	defer rows.Close()

	var tags []*models.Tag
	for rows.Next() {
		var tag models.Tag
		err := rows.Scan(
			&tag.ID, &tag.Name, &tag.Slug, &tag.Description,
			&tag.Color, &tag.UsageCount, &tag.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan tag: %w", err)
		}
		tags = append(tags, &tag)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating tags: %w", err)
	}

	return tags, nil
}

// GetClipsByTag retrieves clips that have a specific tag
func (r *TagRepository) GetClipsByTag(ctx context.Context, tagSlug string, limit, offset int) ([]uuid.UUID, error) {
	query := `
		SELECT ct.clip_id
		FROM clip_tags ct
		INNER JOIN tags t ON ct.tag_id = t.id
		WHERE t.slug = $1
		ORDER BY ct.created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.pool.Query(ctx, query, tagSlug, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get clips by tag: %w", err)
	}
	defer rows.Close()

	var clipIDs []uuid.UUID
	for rows.Next() {
		var clipID uuid.UUID
		if err := rows.Scan(&clipID); err != nil {
			return nil, fmt.Errorf("failed to scan clip ID: %w", err)
		}
		clipIDs = append(clipIDs, clipID)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating clip IDs: %w", err)
	}

	return clipIDs, nil
}

// CountClipsByTag counts clips with a specific tag
func (r *TagRepository) CountClipsByTag(ctx context.Context, tagSlug string) (int, error) {
	query := `
		SELECT COUNT(*)
		FROM clip_tags ct
		INNER JOIN tags t ON ct.tag_id = t.id
		WHERE t.slug = $1
	`

	var count int
	err := r.pool.QueryRow(ctx, query, tagSlug).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count clips by tag: %w", err)
	}

	return count, nil
}

// GetOrCreateTag gets a tag by slug or creates it if it doesn't exist
func (r *TagRepository) GetOrCreateTag(ctx context.Context, name, slug string, color *string) (*models.Tag, error) {
	// Try to get existing tag
	tag, err := r.GetBySlug(ctx, slug)
	if err == nil {
		return tag, nil
	}

	// Create new tag if not found
	newTag := &models.Tag{
		ID:         uuid.New(),
		Name:       name,
		Slug:       slug,
		Color:      color,
		UsageCount: 0,
		CreatedAt:  time.Now(),
	}

	err = r.Create(ctx, newTag)
	if err != nil {
		// Handle race condition where tag was created between check and insert
		if strings.Contains(err.Error(), "unique") || strings.Contains(err.Error(), "duplicate") {
			return r.GetBySlug(ctx, slug)
		}
		return nil, err
	}

	// Fetch the created tag to get the database-generated created_at
	return r.GetBySlug(ctx, slug)
}

// GetClipTagCount returns the number of tags associated with a clip
func (r *TagRepository) GetClipTagCount(ctx context.Context, clipID uuid.UUID) (int, error) {
	query := `SELECT COUNT(*) FROM clip_tags WHERE clip_id = $1`
	var count int
	err := r.pool.QueryRow(ctx, query, clipID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count clip tags: %w", err)
	}
	return count, nil
}

// IsBlacklisted checks if a tag slug matches any blacklisted pattern
func (r *TagRepository) IsBlacklisted(ctx context.Context, slug string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM blacklisted_tags WHERE LOWER(pattern) = LOWER($1))`
	var exists bool
	err := r.pool.QueryRow(ctx, query, slug).Scan(&exists)
	return exists, err
}

// GetBlacklistedTags returns all blacklisted tag patterns
func (r *TagRepository) GetBlacklistedTags(ctx context.Context) ([]models.BlacklistedTag, error) {
	query := `SELECT id, pattern, reason, created_by, created_at FROM blacklisted_tags ORDER BY pattern ASC`
	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get blacklisted tags: %w", err)
	}
	defer rows.Close()

	var tags []models.BlacklistedTag
	for rows.Next() {
		var t models.BlacklistedTag
		if err := rows.Scan(&t.ID, &t.Pattern, &t.Reason, &t.CreatedBy, &t.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan blacklisted tag: %w", err)
		}
		tags = append(tags, t)
	}
	return tags, nil
}

// AddBlacklistedTag adds a pattern to the blacklist
func (r *TagRepository) AddBlacklistedTag(ctx context.Context, pattern string, reason *string, createdBy *uuid.UUID) error {
	query := `INSERT INTO blacklisted_tags (pattern, reason, created_by) VALUES ($1, $2, $3) ON CONFLICT (pattern) DO NOTHING`
	_, err := r.pool.Exec(ctx, query, pattern, reason, createdBy)
	if err != nil {
		return fmt.Errorf("failed to add blacklisted tag: %w", err)
	}
	return nil
}

// RemoveBlacklistedTag removes a pattern from the blacklist
func (r *TagRepository) RemoveBlacklistedTag(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM blacklisted_tags WHERE id = $1`
	result, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to remove blacklisted tag: %w", err)
	}
	if result.RowsAffected() == 0 {
		return fmt.Errorf("blacklisted tag not found")
	}
	return nil
}
