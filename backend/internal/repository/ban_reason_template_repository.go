package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"git.subcult.tv/subculture-collective/clpr/internal/models"
)

type BanReasonTemplateRepository struct {
	pool *pgxpool.Pool
}

func NewBanReasonTemplateRepository(pool *pgxpool.Pool) *BanReasonTemplateRepository {
	return &BanReasonTemplateRepository{pool: pool}
}

// GetByID retrieves a template by ID
func (r *BanReasonTemplateRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.BanReasonTemplate, error) {
	var template models.BanReasonTemplate
	query := `SELECT id, name, reason, duration_seconds, is_default, broadcaster_id, created_by, created_at, updated_at, usage_count, last_used_at FROM ban_reason_templates WHERE id = $1`

	err := r.pool.QueryRow(ctx, query, id).Scan(
		&template.ID,
		&template.Name,
		&template.Reason,
		&template.DurationSeconds,
		&template.IsDefault,
		&template.BroadcasterID,
		&template.CreatedBy,
		&template.CreatedAt,
		&template.UpdatedAt,
		&template.UsageCount,
		&template.LastUsedAt,
	)

	if err == pgx.ErrNoRows {
		return nil, nil
	}
	return &template, err
}

// List retrieves templates with optional filtering
func (r *BanReasonTemplateRepository) List(ctx context.Context, broadcasterID *string, includeDefaults bool) ([]models.BanReasonTemplate, error) {
	var templates []models.BanReasonTemplate

	query := `SELECT id, name, reason, duration_seconds, is_default, broadcaster_id, created_by, created_at, updated_at, usage_count, last_used_at FROM ban_reason_templates WHERE `
	args := []interface{}{}

	if broadcasterID != nil {
		if includeDefaults {
			query += `(broadcaster_id = $1 OR is_default = true)`
			args = append(args, *broadcasterID)
		} else {
			query += `broadcaster_id = $1`
			args = append(args, *broadcasterID)
		}
	} else if includeDefaults {
		query += `is_default = true`
	} else {
		query += `broadcaster_id IS NULL AND is_default = false`
	}

	query += ` ORDER BY is_default DESC, usage_count DESC, name ASC`

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var template models.BanReasonTemplate
		err := rows.Scan(
			&template.ID,
			&template.Name,
			&template.Reason,
			&template.DurationSeconds,
			&template.IsDefault,
			&template.BroadcasterID,
			&template.CreatedBy,
			&template.CreatedAt,
			&template.UpdatedAt,
			&template.UsageCount,
			&template.LastUsedAt,
		)
		if err != nil {
			return nil, err
		}
		templates = append(templates, template)
	}

	return templates, rows.Err()
}

// Create creates a new template
func (r *BanReasonTemplateRepository) Create(ctx context.Context, template *models.BanReasonTemplate) error {
	query := `
		INSERT INTO ban_reason_templates (name, reason, duration_seconds, is_default, broadcaster_id, created_by, created_at, updated_at, usage_count)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id
	`
	now := time.Now()
	template.CreatedAt = now
	template.UpdatedAt = now
	template.UsageCount = 0

	err := r.pool.QueryRow(ctx, query,
		template.Name,
		template.Reason,
		template.DurationSeconds,
		template.IsDefault,
		template.BroadcasterID,
		template.CreatedBy,
		template.CreatedAt,
		template.UpdatedAt,
		template.UsageCount,
	).Scan(&template.ID)

	return err
}

// Update updates an existing template
func (r *BanReasonTemplateRepository) Update(ctx context.Context, id uuid.UUID, updates map[string]interface{}) error {
	// Whitelist of allowed fields to prevent SQL injection
	allowedFields := map[string]bool{
		"name":             true,
		"reason":           true,
		"duration_seconds": true,
		"updated_at":       true,
	}

	updates["updated_at"] = time.Now()

	// Build dynamic update query with whitelisted fields
	query := `UPDATE ban_reason_templates SET `
	args := []interface{}{}
	argNum := 1
	first := true

	for field, value := range updates {
		// Validate field name against whitelist
		if !allowedFields[field] {
			return fmt.Errorf("invalid field name: %s", field)
		}

		if !first {
			query += ", "
		}
		query += fmt.Sprintf("%s = $%d", field, argNum)
		args = append(args, value)
		argNum++
		first = false
	}

	query += fmt.Sprintf(" WHERE id = $%d", argNum)
	args = append(args, id)

	_, err := r.pool.Exec(ctx, query, args...)
	return err
}

// Delete deletes a template
func (r *BanReasonTemplateRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM ban_reason_templates WHERE id = $1 AND is_default = false`
	result, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}

	return nil
}

// IncrementUsage increments the usage count for a template
func (r *BanReasonTemplateRepository) IncrementUsage(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE ban_reason_templates SET usage_count = usage_count + 1, last_used_at = $1 WHERE id = $2`
	_, err := r.pool.Exec(ctx, query, time.Now(), id)
	return err
}

// GetUsageStats retrieves usage statistics for templates
func (r *BanReasonTemplateRepository) GetUsageStats(ctx context.Context, broadcasterID *string) ([]models.BanReasonTemplate, error) {
	var templates []models.BanReasonTemplate

	query := `SELECT id, name, reason, duration_seconds, is_default, broadcaster_id, created_by, created_at, updated_at, usage_count, last_used_at FROM ban_reason_templates WHERE `
	args := []interface{}{}

	if broadcasterID != nil {
		query += `broadcaster_id = $1 ORDER BY usage_count DESC, name ASC`
		args = append(args, *broadcasterID)
	} else {
		query += `is_default = true ORDER BY usage_count DESC, name ASC`
	}

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var template models.BanReasonTemplate
		err := rows.Scan(
			&template.ID,
			&template.Name,
			&template.Reason,
			&template.DurationSeconds,
			&template.IsDefault,
			&template.BroadcasterID,
			&template.CreatedBy,
			&template.CreatedAt,
			&template.UpdatedAt,
			&template.UsageCount,
			&template.LastUsedAt,
		)
		if err != nil {
			return nil, err
		}
		templates = append(templates, template)
	}

	return templates, rows.Err()
}
