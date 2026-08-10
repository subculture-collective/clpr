package services

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"git.subcult.tv/subculture-collective/clpr/internal/models"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrPromotionNotFound      = errors.New("tag promotion queue entry not found")
	ErrPromotionAlreadyPending = errors.New("tag already has a pending promotion request")
	ErrPromotionNotPending     = errors.New("tag promotion entry is not in pending status")
	ErrTagNotFound             = errors.New("tag not found") // shadows repository.ErrTagNotFound
)

// TagPromotionService manages the user tag promotion pipeline.
// Community tags that reach usage thresholds appear in the moderation
// queue. Moderators can approve (move to content/) or reject them.
type TagPromotionService struct {
	pool *pgxpool.Pool
}

// NewTagPromotionService creates a new TagPromotionService.
func NewTagPromotionService(pool *pgxpool.Pool) *TagPromotionService {
	return &TagPromotionService{pool: pool}
}

// CheckPromotionCandidates queries the tag_promotion_candidates view and
// inserts any new candidates into the tag_promotion_queue that aren't
// already pending. Returns the slugs of newly queued candidates.
func (s *TagPromotionService) CheckPromotionCandidates(ctx context.Context) ([]string, error) {
	candidates, err := s.queryCandidates(ctx)
	if err != nil {
		return nil, fmt.Errorf("querying promotion candidates: %w", err)
	}

	var queued []string
	for _, c := range candidates {
		if c.Slug == "" {
			continue
		}
		inserted, err := s.insertIfNotPending(ctx, c)
		if err != nil {
			log.Printf("checkPromotionCandidates: failed to queue %q: %v", c.Slug, err)
			continue
		}
		if inserted {
			queued = append(queued, c.Slug)
		}
	}
	return queued, nil
}

// queryCandidates reads all rows from the tag_promotion_candidates view.
func (s *TagPromotionService) queryCandidates(ctx context.Context) ([]models.TagPromotionCandidate, error) {
	query := `SELECT slug, name, clip_count, unique_users, parent_slug FROM tag_promotion_candidates`
	rows, err := s.pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var candidates []models.TagPromotionCandidate
	for rows.Next() {
		var c models.TagPromotionCandidate
		if err := rows.Scan(&c.Slug, &c.Name, &c.ClipCount, &c.UniqueUsers, &c.ParentSlug); err != nil {
			return nil, fmt.Errorf("scanning candidate: %w", err)
		}
		candidates = append(candidates, c)
	}
	return candidates, rows.Err()
}

// insertIfNotPending inserts a candidate into the queue unless a pending
// entry already exists for the same tag_slug.
func (s *TagPromotionService) insertIfNotPending(ctx context.Context, c models.TagPromotionCandidate) (bool, error) {
	now := time.Now()
	id := uuid.New()

	query := `
		INSERT INTO tag_promotion_queue (id, tag_slug, usage_count, unique_users, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, 'pending', $5, $5)
		ON CONFLICT (tag_slug) WHERE status = 'pending' DO NOTHING
	`

	tag, err := s.pool.Exec(ctx, query, id, c.Slug, c.ClipCount, c.UniqueUsers, now)
	if err != nil {
		return false, err
	}
	return tag.RowsAffected() > 0, nil
}

// ApprovePromotion moves a tag from community/ to the content/ parent and
// marks its queue entry as approved.
func (s *TagPromotionService) ApprovePromotion(ctx context.Context, tagSlug string, reviewerID uuid.UUID) error {
	now := time.Now()

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("beginning transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// Update the tag's parent_slug from community to content
	tagResult, err := tx.Exec(ctx,
		`UPDATE tags SET parent_slug = 'content' WHERE slug = $1 AND parent_slug = 'community'`,
		tagSlug,
	)
	if err != nil {
		return fmt.Errorf("updating tag parent: %w", err)
	}
	if tagResult.RowsAffected() == 0 {
		return fmt.Errorf("%w: %s", ErrTagNotFound, tagSlug)
	}

	// Mark the pending queue entry as approved
	qResult, err := tx.Exec(ctx,
		`UPDATE tag_promotion_queue
		 SET status = 'approved', reviewed_by = $1, reviewed_at = $2, promoted_at = $2, updated_at = $2
		 WHERE tag_slug = $3 AND status = 'pending'`,
		reviewerID, now, tagSlug,
	)
	if err != nil {
		return fmt.Errorf("updating promotion queue: %w", err)
	}
	if qResult.RowsAffected() == 0 {
		return fmt.Errorf("%w: %s", ErrPromotionNotPending, tagSlug)
	}

	return tx.Commit(ctx)
}

// RejectPromotion marks a pending queue entry as rejected.
func (s *TagPromotionService) RejectPromotion(ctx context.Context, tagSlug string, reviewerID uuid.UUID) error {
	now := time.Now()

	result, err := s.pool.Exec(ctx,
		`UPDATE tag_promotion_queue
		 SET status = 'rejected', reviewed_by = $1, reviewed_at = $2, updated_at = $2
		 WHERE tag_slug = $3 AND status = 'pending'`,
		reviewerID, now, tagSlug,
	)
	if err != nil {
		return fmt.Errorf("rejecting promotion: %w", err)
	}
	if result.RowsAffected() == 0 {
		return fmt.Errorf("%w: %s", ErrPromotionNotPending, tagSlug)
	}
	return nil
}

// GetPendingQueue returns all pending entries in the promotion queue.
func (s *TagPromotionService) GetPendingQueue(ctx context.Context) ([]models.TagPromotionQueueItem, error) {
	query := `
		SELECT id, tag_slug, usage_count, unique_users, status,
		       reviewed_by, reviewed_at, promoted_at, created_at, updated_at
		FROM tag_promotion_queue
		WHERE status = 'pending'
		ORDER BY unique_users DESC, usage_count DESC
	`
	return s.queryQueueItems(ctx, query)
}

// GetQueueItemBySlug returns a specific queue entry by tag slug and optional status filter.
func (s *TagPromotionService) GetQueueItemBySlug(ctx context.Context, tagSlug string) (*models.TagPromotionQueueItem, error) {
	query := `
		SELECT id, tag_slug, usage_count, unique_users, status,
		       reviewed_by, reviewed_at, promoted_at, created_at, updated_at
		FROM tag_promotion_queue
		WHERE tag_slug = $1 AND status = 'pending'
		ORDER BY created_at DESC
		LIMIT 1
	`

	var item models.TagPromotionQueueItem
	err := s.pool.QueryRow(ctx, query, tagSlug).Scan(
		&item.ID, &item.TagSlug, &item.UsageCount, &item.UniqueUsers,
		&item.Status, &item.ReviewedBy, &item.ReviewedAt, &item.PromotedAt,
		&item.CreatedAt, &item.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrPromotionNotFound
		}
		return nil, fmt.Errorf("querying queue item: %w", err)
	}
	return &item, nil
}

// queryQueueItems is a helper to run a query and scan queue item rows.
func (s *TagPromotionService) queryQueueItems(ctx context.Context, query string) ([]models.TagPromotionQueueItem, error) {
	rows, err := s.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("querying queue items: %w", err)
	}
	defer rows.Close()

	var items []models.TagPromotionQueueItem
	for rows.Next() {
		var item models.TagPromotionQueueItem
		if err := rows.Scan(
			&item.ID, &item.TagSlug, &item.UsageCount, &item.UniqueUsers,
			&item.Status, &item.ReviewedBy, &item.ReviewedAt, &item.PromotedAt,
			&item.CreatedAt, &item.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scanning queue item: %w", err)
		}
		items = append(items, item)
	}
	return items, rows.Err()
}