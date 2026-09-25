package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"git.subcult.tv/subculture-collective/clpr/internal/models"
	"git.subcult.tv/subculture-collective/clpr/internal/tagtaxonomy"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrTagNotFound            = errors.New("tag not found")
	ErrTagConflict            = errors.New("tag already exists")
	ErrBlacklistedTagNotFound = errors.New("blacklisted tag not found")
	ErrBlacklistedTagConflict = errors.New("blacklisted tag already exists")
)

func tagWriteError(err error) error {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" {
		return ErrTagConflict
	}
	return err
}

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
		INSERT INTO tags (id, name, slug, parent_slug, description, color, usage_count, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`

	_, err := r.pool.Exec(ctx, query,
		tag.ID, tag.Name, tag.Slug, tag.ParentSlug, tag.Description, tag.Color,
		tag.UsageCount, tag.CreatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create tag: %w", tagWriteError(err))
	}

	return nil
}

// GetByID retrieves a tag by its ID
func (r *TagRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Tag, error) {
	query := `
		SELECT id, name, slug, parent_slug, description, color, usage_count, created_at
		FROM tags
		WHERE id = $1
	`

	var tag models.Tag
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&tag.ID, &tag.Name, &tag.Slug, &tag.ParentSlug, &tag.Description,
		&tag.Color, &tag.UsageCount, &tag.CreatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, ErrTagNotFound
		}
		return nil, fmt.Errorf("failed to get tag by ID: %w", err)
	}

	return &tag, nil
}

// GetBySlug retrieves a tag by its slug
func (r *TagRepository) GetBySlug(ctx context.Context, slug string) (*models.Tag, error) {
	query := `
		SELECT t.id, t.name, t.slug, t.parent_slug, t.description, t.color, t.usage_count, t.created_at
		FROM tags t
		WHERE (t.slug = $1 OR t.id = (SELECT canonical_tag_id FROM tag_aliases WHERE alias_slug = $1))
		  AND NOT EXISTS (SELECT 1 FROM tag_suppressions s WHERE s.tag_id = t.id)
		ORDER BY CASE WHEN t.slug = $1 THEN 0 ELSE 1 END
		LIMIT 1
	`

	var tag models.Tag
	err := r.pool.QueryRow(ctx, query, slug).Scan(
		&tag.ID, &tag.Name, &tag.Slug, &tag.ParentSlug, &tag.Description,
		&tag.Color, &tag.UsageCount, &tag.CreatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, ErrTagNotFound
		}
		return nil, fmt.Errorf("failed to get tag by slug: %w", err)
	}

	return &tag, nil
}

// visibleTagPredicate excludes blacklisted and suppressed tags for the tags
// row aliased as alias. It matches is_tag_blacklisted() but is written inline
// so PostgreSQL hash-joins exact-match patterns (blacklisted_tags.literal_slug,
// migration 000144) instead of calling the function once per tag.
func visibleTagPredicate(alias string) string {
	return `NOT EXISTS (SELECT 1 FROM blacklisted_tags b WHERE b.literal_slug = lower(` + alias + `.slug))
		  AND NOT EXISTS (SELECT 1 FROM blacklisted_tags b WHERE b.literal_slug IS NULL AND ` + alias + `.slug ILIKE b.like_pattern ESCAPE '\')
		  AND NOT EXISTS (SELECT 1 FROM tag_suppressions s WHERE s.tag_id = ` + alias + `.id)`
}

// tagLanePredicate returns a SQL predicate on alias t for a tagtaxonomy lane.
// slugs is the placeholder bound to the catalog content slugs; every branch
// references it so the argument count never varies. An empty or unknown lane
// matches every tag.
func tagLanePredicate(lane, slugs string) string {
	bound := `cardinality(` + slugs + `::text[]) >= 0`
	var predicate string
	switch tagtaxonomy.Lane(lane) {
	case tagtaxonomy.LaneCategory:
		predicate = `t.slug LIKE 'game/%'`
	case tagtaxonomy.LaneDetected:
		predicate = `t.slug = ANY(` + slugs + `::text[])`
	case tagtaxonomy.LaneStreamer:
		predicate = `(t.slug LIKE 'streamer/%' OR (t.slug LIKE 'content/%' AND NOT t.slug = ANY(` + slugs + `::text[])))`
	case tagtaxonomy.LaneCommunity:
		predicate = `(t.slug LIKE 'community/%' OR (strpos(t.slug, '/') = 0 AND t.slug NOT IN ('content','game','duration','lang','community','streamer')))`
	case tagtaxonomy.LaneDuration:
		predicate = `t.slug LIKE 'duration/%'`
	case tagtaxonomy.LaneLanguage:
		predicate = `t.slug LIKE 'lang/%'`
	default:
		return bound
	}
	return `(` + predicate + ` AND ` + bound + `)`
}

// List retrieves tags with optional lane filtering, sorting and pagination.
// The "curated" sort is kept for older clients and means the detected lane.
func (r *TagRepository) List(ctx context.Context, sort, lane string, limit, offset int) ([]*models.Tag, error) {
	if sort == "curated" {
		sort, lane = "popularity", string(tagtaxonomy.LaneDetected)
	}
	visible := visibleTagPredicate("t")
	laneClause := tagLanePredicate(lane, "$3")
	var query string
	switch sort {
	case "trending":
		query = `
		SELECT t.id, t.name, t.slug, t.parent_slug, t.description, t.color,
		       COUNT(ct.clip_id)::int AS usage_count, t.created_at
		FROM tags t JOIN clip_tags ct ON ct.tag_id = t.id
		JOIN clips c ON c.id = ct.clip_id
		WHERE ct.created_at >= NOW() - INTERVAL '7 days'
		  AND c.is_removed = FALSE AND c.is_hidden = FALSE
		  AND ` + visible + ` AND ` + laneClause + `
		GROUP BY t.id ORDER BY usage_count DESC, t.name ASC LIMIT $1 OFFSET $2`
	case "alphabetical":
		query = `
		SELECT t.id, t.name, t.slug, t.parent_slug, t.description, t.color, t.usage_count, t.created_at
		FROM tags t WHERE ` + visible + ` AND ` + laneClause + `
		ORDER BY t.name ASC LIMIT $1 OFFSET $2`
	case "recent":
		query = `
		SELECT t.id, t.name, t.slug, t.parent_slug, t.description, t.color, t.usage_count, t.created_at
		FROM tags t WHERE ` + visible + ` AND ` + laneClause + `
		ORDER BY t.created_at DESC LIMIT $1 OFFSET $2`
	default:
		query = `
		SELECT t.id, t.name, t.slug, t.parent_slug, t.description, t.color, t.usage_count, t.created_at
		FROM tags t WHERE ` + visible + ` AND ` + laneClause + `
		ORDER BY t.usage_count DESC, t.name ASC LIMIT $1 OFFSET $2`
	}

	rows, err := r.pool.Query(ctx, query, limit, offset, tagtaxonomy.DetectedSlugs())
	if err != nil {
		return nil, fmt.Errorf("failed to list tags: %w", err)
	}
	defer rows.Close()

	var tags []*models.Tag
	for rows.Next() {
		var tag models.Tag
		err := rows.Scan(
			&tag.ID, &tag.Name, &tag.Slug, &tag.ParentSlug, &tag.Description,
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

// Count returns the number of visible tags, optionally within one lane.
func (r *TagRepository) Count(ctx context.Context, lane string) (int, error) {
	var count int
	query := `SELECT COUNT(*) FROM tags t
		WHERE ` + visibleTagPredicate("t") + `
		  AND ` + tagLanePredicate(lane, "$1")
	err := r.pool.QueryRow(ctx, query, tagtaxonomy.DetectedSlugs()).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count tags: %w", err)
	}
	return count, nil
}

// Search searches for tags by name
func (r *TagRepository) Search(ctx context.Context, query string, limit int) ([]*models.Tag, error) {
	searchQuery := `
		SELECT t.id, t.name, t.slug, t.parent_slug, t.description, t.color, t.usage_count, t.created_at
		FROM tags t
		WHERE (t.name ILIKE $1 OR t.slug ILIKE $1)
		  AND ` + visibleTagPredicate("t") + `
		ORDER BY t.usage_count DESC
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
			&tag.ID, &tag.Name, &tag.Slug, &tag.ParentSlug, &tag.Description,
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

func (r *TagRepository) ListAdmin(ctx context.Context, limit, offset int) ([]*models.Tag, error) {
	rows, err := r.pool.Query(ctx, `SELECT t.id,t.name,t.slug,t.parent_slug,t.description,t.color,t.usage_count,t.created_at,
		s.suppressed_at,s.suppressed_by,s.reason FROM tags t LEFT JOIN tag_suppressions s ON s.tag_id=t.id
		ORDER BY t.name LIMIT $1 OFFSET $2`, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	tags := []*models.Tag{}
	for rows.Next() {
		var tag models.Tag
		if err := rows.Scan(&tag.ID, &tag.Name, &tag.Slug, &tag.ParentSlug, &tag.Description, &tag.Color, &tag.UsageCount, &tag.CreatedAt, &tag.SuppressedAt, &tag.SuppressedBy, &tag.SuppressionReason); err != nil {
			return nil, err
		}
		tags = append(tags, &tag)
	}
	return tags, rows.Err()
}

func (r *TagRepository) Suppress(ctx context.Context, id, actor uuid.UUID, reason string) error {
	result, err := r.pool.Exec(ctx, `INSERT INTO tag_suppressions(tag_id,suppressed_by,reason) VALUES($1,$2,$3)
		ON CONFLICT(tag_id) DO UPDATE SET suppressed_by=EXCLUDED.suppressed_by,reason=EXCLUDED.reason,suppressed_at=NOW()`, id, actor, reason)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return ErrTagNotFound
	}
	return nil
}

func (r *TagRepository) Restore(ctx context.Context, id uuid.UUID) error {
	result, err := r.pool.Exec(ctx, `DELETE FROM tag_suppressions WHERE tag_id=$1`, id)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return ErrTagNotFound
	}
	return nil
}

// Update updates an existing tag
func (r *TagRepository) Update(ctx context.Context, tag *models.Tag) error {
	query := `
		UPDATE tags
		SET name = $2, slug = $3, parent_slug = $4, description = $5, color = $6
		WHERE id = $1
	`

	result, err := r.pool.Exec(ctx, query,
		tag.ID, tag.Name, tag.Slug, tag.ParentSlug, tag.Description, tag.Color,
	)

	if err != nil {
		return fmt.Errorf("failed to update tag: %w", tagWriteError(err))
	}

	if result.RowsAffected() == 0 {
		return ErrTagNotFound
	}

	return nil
}

// Delete deletes a tag and its associations
func (r *TagRepository) Delete(ctx context.Context, id uuid.UUID) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	// Delete all clip-tag associations first
	_, err = tx.Exec(ctx, "DELETE FROM clip_tags WHERE tag_id = $1", id)
	if err != nil {
		return fmt.Errorf("failed to delete clip-tag associations: %w", err)
	}

	// Delete the tag
	query := `DELETE FROM tags WHERE id = $1`
	result, err := tx.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete tag: %w", err)
	}

	if result.RowsAffected() == 0 {
		return ErrTagNotFound
	}
	return tx.Commit(ctx)
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
		SELECT t.id, t.name, t.slug, t.parent_slug, t.description, t.color, t.usage_count, t.created_at
		FROM tags t
		INNER JOIN clip_tags ct ON t.id = ct.tag_id
		WHERE ct.clip_id = $1
		  AND NOT is_tag_blacklisted(t.slug)
		  AND NOT EXISTS (SELECT 1 FROM tag_suppressions s WHERE s.tag_id = t.id)
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
			&tag.ID, &tag.Name, &tag.Slug, &tag.ParentSlug, &tag.Description,
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
		  AND NOT is_tag_blacklisted(t.slug)
		  AND NOT EXISTS (SELECT 1 FROM tag_suppressions s WHERE s.tag_id=t.id)
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
		  AND NOT is_tag_blacklisted(t.slug)
		  AND NOT EXISTS (SELECT 1 FROM tag_suppressions s WHERE s.tag_id=t.id)
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
	return r.GetOrCreateTagWithParent(ctx, name, slug, nil, color)
}

func (r *TagRepository) GetOrCreateTagWithParent(ctx context.Context, name, slug string, parentSlug, color *string) (*models.Tag, error) {
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
		ParentSlug: parentSlug,
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
	query := `
		SELECT COUNT(*)
		FROM clip_tags ct
		JOIN tags t ON t.id = ct.tag_id
		WHERE ct.clip_id = $1
		  AND NOT EXISTS (SELECT 1 FROM tag_suppressions s WHERE s.tag_id = t.id)
	`
	var count int
	err := r.pool.QueryRow(ctx, query, clipID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count clip tags: %w", err)
	}
	return count, nil
}

// IsBlacklisted checks if a tag slug matches any blacklisted pattern
func (r *TagRepository) IsBlacklisted(ctx context.Context, slug string) (bool, error) {
	query := `SELECT is_tag_blacklisted($1)`
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
	result, err := r.pool.Exec(ctx, query, pattern, reason, createdBy)
	if err != nil {
		return fmt.Errorf("failed to add blacklisted tag: %w", err)
	}
	if result.RowsAffected() == 0 {
		return ErrBlacklistedTagConflict
	}
	return nil
}

// MaxTagTreeDepth bounds tag subtree traversal in case parent_slug data ever
// contains a cycle.
const MaxTagTreeDepth = 10

// TagTreeRow is a visible tag returned by a tag tree query. ChildCount is the
// total number of visible children, which may exceed the children returned.
type TagTreeRow struct {
	models.Tag
	Depth      int
	ChildCount int
}

// GetTagTree returns the tag rootSlug and its visible descendants, keeping at
// most childLimit children (highest usage first) under each parent.
func (r *TagRepository) GetTagTree(ctx context.Context, rootSlug string, childLimit int) ([]TagTreeRow, error) {
	query := `
		WITH RECURSIVE tag_tree AS (
			SELECT t.id, t.name, t.slug, t.parent_slug, t.description, t.color, t.usage_count, t.created_at, 0 AS depth
			FROM tags t
			WHERE t.slug = $1
			  AND ` + visibleTagPredicate("t") + `
			UNION ALL
			SELECT t.id, t.name, t.slug, t.parent_slug, t.description, t.color, t.usage_count, t.created_at, tt.depth + 1
			FROM tags t
			INNER JOIN tag_tree tt ON t.parent_slug = tt.slug
			WHERE tt.depth < $3
			  AND ` + visibleTagPredicate("t") + `
		),
		child_counts AS (
			SELECT parent_slug, COUNT(*)::int AS child_count
			FROM tag_tree
			WHERE depth > 0
			GROUP BY parent_slug
		),
		ranked AS (
			SELECT tt.*, row_number() OVER (PARTITION BY tt.parent_slug ORDER BY tt.usage_count DESC, tt.name) AS sibling_rank
			FROM tag_tree tt
		)
		SELECT r.id, r.name, r.slug, r.parent_slug, r.description, r.color, r.usage_count, r.created_at,
		       r.depth, COALESCE(cc.child_count, 0)
		FROM ranked r
		LEFT JOIN child_counts cc ON cc.parent_slug = r.slug
		WHERE r.depth = 0 OR r.sibling_rank <= $2
		ORDER BY r.depth, r.usage_count DESC, r.name
	`
	return r.queryTagTree(ctx, query, rootSlug, childLimit, MaxTagTreeDepth)
}

// GetTagForest returns every visible top-level tag that has visible children,
// each followed by at most childLimit of its children (highest usage first).
// Top-level tags without children are ordinary flat tags; list them with List.
func (r *TagRepository) GetTagForest(ctx context.Context, childLimit int) ([]TagTreeRow, error) {
	query := `
		WITH child_counts AS (
			SELECT c.parent_slug, COUNT(*)::int AS child_count
			FROM tags c
			WHERE c.parent_slug IS NOT NULL
			  AND ` + visibleTagPredicate("c") + `
			GROUP BY c.parent_slug
		),
		roots AS (
			SELECT t.id, t.name, t.slug, t.parent_slug, t.description, t.color, t.usage_count, t.created_at,
			       0 AS depth, cc.child_count
			FROM tags t
			INNER JOIN child_counts cc ON cc.parent_slug = t.slug
			WHERE t.parent_slug IS NULL
			  AND ` + visibleTagPredicate("t") + `
		),
		children AS (
			SELECT c.id, c.name, c.slug, c.parent_slug, c.description, c.color, c.usage_count, c.created_at,
			       1 AS depth, COALESCE(gc.child_count, 0) AS child_count,
			       row_number() OVER (PARTITION BY c.parent_slug ORDER BY c.usage_count DESC, c.name) AS sibling_rank
			FROM tags c
			INNER JOIN roots ON c.parent_slug = roots.slug
			LEFT JOIN child_counts gc ON gc.parent_slug = c.slug
			WHERE ` + visibleTagPredicate("c") + `
		)
		SELECT id, name, slug, parent_slug, description, color, usage_count, created_at, depth, child_count
		FROM (
			SELECT id, name, slug, parent_slug, description, color, usage_count, created_at, depth, child_count FROM roots
			UNION ALL
			SELECT id, name, slug, parent_slug, description, color, usage_count, created_at, depth, child_count
			FROM children WHERE sibling_rank <= $1
		) forest
		ORDER BY depth, usage_count DESC, name
	`
	return r.queryTagTree(ctx, query, childLimit)
}

func (r *TagRepository) queryTagTree(ctx context.Context, query string, args ...any) ([]TagTreeRow, error) {
	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get tag tree: %w", err)
	}
	defer rows.Close()

	var tags []TagTreeRow
	for rows.Next() {
		var row TagTreeRow
		err := rows.Scan(
			&row.ID, &row.Name, &row.Slug, &row.ParentSlug, &row.Description,
			&row.Color, &row.UsageCount, &row.CreatedAt, &row.Depth, &row.ChildCount,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan tag tree row: %w", err)
		}
		tags = append(tags, row)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating tag tree: %w", err)
	}

	return tags, nil
}

// RemoveBlacklistedTag removes a pattern from the blacklist
func (r *TagRepository) RemoveBlacklistedTag(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM blacklisted_tags WHERE id = $1`
	result, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to remove blacklisted tag: %w", err)
	}
	if result.RowsAffected() == 0 {
		return ErrBlacklistedTagNotFound
	}
	return nil
}
