package repository

import (
	"context"
	"encoding/json"
	"fmt"

	"git.subcult.tv/subculture-collective/clpr/internal/models"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ClipTopicRepository struct{ pool *pgxpool.Pool }

func NewClipTopicRepository(pool *pgxpool.Pool) *ClipTopicRepository {
	return &ClipTopicRepository{pool: pool}
}

func (r *ClipTopicRepository) FindActiveTopicIDBySlug(ctx context.Context, slug string) (uuid.UUID, error) {
	var id uuid.UUID
	err := r.pool.QueryRow(ctx, `SELECT id FROM categories WHERE slug = $1 AND is_active = TRUE`, slug).Scan(&id)
	return id, err
}

func (r *ClipTopicRepository) ListForClip(ctx context.Context, clipID uuid.UUID) ([]models.ClipTopic, error) {
	rows, err := r.pool.Query(ctx, `SELECT ct.clip_id,ct.topic_id,c.name,c.slug,ct.source,ct.confidence,ct.evidence,ct.assigned_by_user_id,ct.created_at,ct.updated_at FROM clip_topics ct JOIN categories c ON c.id=ct.topic_id WHERE ct.clip_id=$1 AND c.is_active=TRUE ORDER BY ct.confidence DESC,c.position,c.name`, clipID)
	if err != nil {
		return nil, fmt.Errorf("list clip topics: %w", err)
	}
	defer rows.Close()
	var topics []models.ClipTopic
	for rows.Next() {
		var topic models.ClipTopic
		if err := rows.Scan(&topic.ClipID, &topic.TopicID, &topic.TopicName, &topic.TopicSlug, &topic.Source, &topic.Confidence, &topic.Evidence, &topic.AssignedByUserID, &topic.CreatedAt, &topic.UpdatedAt); err != nil {
			return nil, err
		}
		topics = append(topics, topic)
	}
	return topics, rows.Err()
}

func (r *ClipTopicRepository) GetClassificationInput(ctx context.Context, clipID uuid.UUID) (*models.ClipTopicClassificationInput, error) {
	var input models.ClipTopicClassificationInput
	err := r.pool.QueryRow(ctx, `SELECT c.id,c.title,COALESCE(c.game_id,''),COALESCE(c.game_name,''),COALESCE((SELECT array_agg(DISTINCT t.slug) FROM clip_tags ct JOIN tags t ON t.id=ct.tag_id WHERE ct.clip_id=c.id),'{}'),COALESCE((SELECT full_text FROM clip_transcripts WHERE clip_id=c.id),'') FROM clips c WHERE c.id=$1`, clipID).Scan(&input.ClipID, &input.Title, &input.TwitchCategoryID, &input.TwitchCategoryName, &input.Tags, &input.Transcript)
	if err != nil {
		return nil, fmt.Errorf("load topic classification input: %w", err)
	}
	rows, err := r.pool.Query(ctx, `SELECT category.id,category.name,category.slug,category.description,category.icon,category.position,category.category_type,category.is_featured,category.is_custom,category.created_by_user_id,category.created_at,category.updated_at FROM games g JOIN category_games cg ON cg.game_id=g.id JOIN categories category ON category.id=cg.category_id WHERE g.twitch_game_id=$1 AND category.is_active=TRUE AND category.category_type='topic'`, input.TwitchCategoryID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var topic models.Category
		if err := rows.Scan(&topic.ID, &topic.Name, &topic.Slug, &topic.Description, &topic.Icon, &topic.Position, &topic.CategoryType, &topic.IsFeatured, &topic.IsCustom, &topic.CreatedByUserID, &topic.CreatedAt, &topic.UpdatedAt); err != nil {
			return nil, err
		}
		input.MappedTopics = append(input.MappedTopics, topic)
	}
	return &input, rows.Err()
}

func (r *ClipTopicRepository) UpsertCandidates(ctx context.Context, clipID uuid.UUID, candidates []models.TopicCandidate) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	if _, err = tx.Exec(ctx, `DELETE FROM clip_topics WHERE clip_id=$1 AND source<>'manual'`, clipID); err != nil {
		return err
	}
	for _, candidate := range candidates {
		evidence := candidate.Evidence
		if len(evidence) == 0 {
			evidence = json.RawMessage(`{}`)
		}
		_, err = tx.Exec(ctx, `INSERT INTO clip_topics (clip_id,topic_id,source,confidence,evidence) VALUES ($1,$2,$3,$4,$5) ON CONFLICT (clip_id,topic_id) DO UPDATE SET source=CASE WHEN clip_topics.source='manual' THEN clip_topics.source ELSE EXCLUDED.source END,confidence=CASE WHEN clip_topics.source='manual' THEN clip_topics.confidence ELSE EXCLUDED.confidence END,evidence=CASE WHEN clip_topics.source='manual' THEN clip_topics.evidence ELSE EXCLUDED.evidence END,updated_at=NOW()`, clipID, candidate.TopicID, candidate.Source, candidate.Confidence, evidence)
		if err != nil {
			return fmt.Errorf("upsert clip topic: %w", err)
		}
	}
	if _, err = tx.Exec(ctx, `UPDATE clips SET topics_classified_at=NOW() WHERE id=$1`, clipID); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func (r *ClipTopicRepository) ListUnclassifiedClipIDs(ctx context.Context, limit int) ([]uuid.UUID, error) {
	rows, err := r.pool.Query(ctx, `SELECT id FROM clips WHERE topics_classified_at IS NULL AND is_removed=FALSE AND is_hidden=FALSE ORDER BY imported_at LIMIT $1`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var ids []uuid.UUID
	for rows.Next() {
		var id uuid.UUID
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}

func (r *ClipTopicRepository) ReplaceManual(ctx context.Context, clipID, userID uuid.UUID, topicIDs []uuid.UUID) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	if _, err = tx.Exec(ctx, `DELETE FROM clip_topics WHERE clip_id=$1`, clipID); err != nil {
		return err
	}
	seen := make(map[uuid.UUID]struct{}, len(topicIDs))
	for _, topicID := range topicIDs {
		if _, duplicate := seen[topicID]; duplicate {
			continue
		}
		seen[topicID] = struct{}{}
		result, insertErr := tx.Exec(ctx, `INSERT INTO clip_topics (clip_id,topic_id,source,confidence,evidence,assigned_by_user_id) SELECT $1,id,'manual',1,'{}',$3 FROM categories WHERE id=$2 AND is_active=TRUE AND category_type='topic'`, clipID, topicID, userID)
		if insertErr != nil {
			return insertErr
		}
		if result.RowsAffected() != 1 {
			return fmt.Errorf("topic %s is not active", topicID)
		}
	}
	_, err = tx.Exec(ctx, `UPDATE clips SET topics_classified_at=NOW() WHERE id=$1`, clipID)
	if err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func (r *ClipTopicRepository) Merge(ctx context.Context, sourceID, targetID uuid.UUID) error {
	if sourceID == targetID {
		return fmt.Errorf("source and target topics must differ")
	}
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	var targetActive bool
	if err = tx.QueryRow(ctx, `SELECT is_active FROM categories WHERE id=$1 FOR UPDATE`, targetID).Scan(&targetActive); err != nil {
		return err
	}
	if !targetActive {
		return fmt.Errorf("target topic is inactive")
	}
	_, err = tx.Exec(ctx, `INSERT INTO clip_topics (clip_id,topic_id,source,confidence,evidence,assigned_by_user_id) SELECT clip_id,$2,source,confidence,evidence,assigned_by_user_id FROM clip_topics WHERE topic_id=$1 ON CONFLICT (clip_id,topic_id) DO UPDATE SET confidence=GREATEST(clip_topics.confidence,EXCLUDED.confidence),updated_at=NOW()`, sourceID, targetID)
	if err != nil {
		return err
	}
	if _, err = tx.Exec(ctx, `DELETE FROM clip_topics WHERE topic_id=$1`, sourceID); err != nil {
		return err
	}
	result, err := tx.Exec(ctx, `UPDATE categories SET is_active=FALSE,merged_into_id=$2,is_featured=FALSE,updated_at=NOW() WHERE id=$1 AND is_active=TRUE`, sourceID, targetID)
	if err != nil {
		return err
	}
	if result.RowsAffected() != 1 {
		return pgx.ErrNoRows
	}
	return tx.Commit(ctx)
}

func (r *ClipTopicRepository) Split(ctx context.Context, sourceID, userID uuid.UUID, request models.SplitTopicRequest) (*models.Category, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)
	topic := &models.Category{ID: uuid.New(), Name: request.Name, Slug: request.Slug, Description: request.Description, CategoryType: "topic", IsCustom: true, CreatedByUserID: &userID}
	err = tx.QueryRow(ctx, `INSERT INTO categories (id,name,slug,description,category_type,is_custom,created_by_user_id) VALUES ($1,$2,$3,$4,'topic',TRUE,$5) RETURNING created_at,updated_at`, topic.ID, topic.Name, topic.Slug, topic.Description, userID).Scan(&topic.CreatedAt, &topic.UpdatedAt)
	if err != nil {
		return nil, err
	}
	result, err := tx.Exec(ctx, `WITH moved AS (
		DELETE FROM clip_topics WHERE topic_id=$1 AND clip_id=ANY($2::uuid[]) RETURNING clip_id
	) INSERT INTO clip_topics (clip_id,topic_id,source,confidence,evidence,assigned_by_user_id)
	SELECT clip_id,$3,'manual',1,'{}',$4 FROM moved`, sourceID, request.ClipIDs, topic.ID, userID)
	if err != nil {
		return nil, err
	}
	if result.RowsAffected() != int64(len(request.ClipIDs)) {
		return nil, fmt.Errorf("every split clip must belong to the source topic")
	}
	if err = tx.Commit(ctx); err != nil {
		return nil, err
	}
	return topic, nil
}
