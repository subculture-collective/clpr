package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"git.subcult.tv/subculture-collective/clpr/internal/models"
	"git.subcult.tv/subculture-collective/clpr/internal/utils"
)

// AuditLogRepository handles database operations for moderation audit logs
type AuditLogRepository struct {
	db *pgxpool.Pool
}

// NewAuditLogRepository creates a new AuditLogRepository
func NewAuditLogRepository(db *pgxpool.Pool) *AuditLogRepository {
	return &AuditLogRepository{db: db}
}

// Create creates a new audit log entry
func (r *AuditLogRepository) Create(ctx context.Context, log *models.ModerationAuditLog) error {
	if log.ID == uuid.Nil {
		log.ID = uuid.New()
	}

	if log.CreatedAt.IsZero() {
		log.CreatedAt = time.Now()
	}

	// If no moderator is provided, skip creating the audit log to avoid invalid foreign keys.
	// System-generated events should either supply a system moderator user or opt-out explicitly.
	if log.ModeratorID == uuid.Nil {
		return nil
	}

	// Convert metadata to JSON
	var metadataJSON []byte
	var err error
	if log.Metadata != nil {
		metadataJSON, err = json.Marshal(log.Metadata)
		if err != nil {
			return fmt.Errorf("failed to marshal metadata: %w", err)
		}
	}

	query := `
		INSERT INTO moderation_audit_logs (
			id, action, entity_type, entity_id, moderator_id, actor_id, reason, metadata,
			ip_address, user_agent, channel_id, created_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12
		)`

	// Ensure timestamp is in UTC to avoid timezone offset issues
	utcTime := log.CreatedAt.UTC()

	_, err = r.db.Exec(ctx, query,
		log.ID,
		log.Action,
		log.EntityType,
		log.EntityID,
		log.ModeratorID,
		// Populate actor_id with moderator_id for now until services provide explicit actor
		log.ModeratorID,
		log.Reason,
		metadataJSON,
		log.IPAddress,
		log.UserAgent,
		log.ChannelID,
		utcTime,
	)

	return err
}

// List retrieves audit logs with optional filters
func (r *AuditLogRepository) List(ctx context.Context, filters AuditLogFilters, page, limit int) ([]*models.ModerationAuditLogWithUser, int, error) {
	offset := (page - 1) * limit

	// Build query with filters
	whereClause := "WHERE 1=1"
	args := []interface{}{}
	placeholderIndex := 1

	if filters.ModeratorID != nil {
		whereClause += fmt.Sprintf(" AND mal.moderator_id = %s", utils.SQLPlaceholder(placeholderIndex))
		args = append(args, *filters.ModeratorID)
		placeholderIndex++
	}

	if filters.Action != "" {
		whereClause += fmt.Sprintf(" AND mal.action = %s", utils.SQLPlaceholder(placeholderIndex))
		args = append(args, filters.Action)
		placeholderIndex++
	}

	if filters.EntityType != "" {
		whereClause += fmt.Sprintf(" AND mal.entity_type = %s", utils.SQLPlaceholder(placeholderIndex))
		args = append(args, filters.EntityType)
		placeholderIndex++
	}

	if filters.EntityID != nil {
		whereClause += fmt.Sprintf(" AND mal.entity_id = %s", utils.SQLPlaceholder(placeholderIndex))
		args = append(args, *filters.EntityID)
		placeholderIndex++
	}

	if filters.ChannelID != nil {
		whereClause += fmt.Sprintf(" AND mal.channel_id = %s", utils.SQLPlaceholder(placeholderIndex))
		args = append(args, *filters.ChannelID)
		placeholderIndex++
	}

	if filters.StartDate != nil {
		whereClause += fmt.Sprintf(" AND mal.created_at >= %s", utils.SQLPlaceholder(placeholderIndex))
		args = append(args, *filters.StartDate)
		placeholderIndex++
	}

	if filters.EndDate != nil {
		whereClause += fmt.Sprintf(" AND mal.created_at <= %s", utils.SQLPlaceholder(placeholderIndex))
		args = append(args, *filters.EndDate)
		placeholderIndex++
	}

	if filters.Search != "" {
		whereClause += fmt.Sprintf(" AND mal.reason ILIKE %s", utils.SQLPlaceholder(placeholderIndex))
		args = append(args, "%"+filters.Search+"%")
		placeholderIndex++
	}

	// Count total
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM moderation_audit_logs mal %s", whereClause)
	var total int
	if err := r.db.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	// Get logs with moderator info
	query := fmt.Sprintf(`
		SELECT
			mal.id, mal.action, mal.entity_type, mal.entity_id, mal.moderator_id,
			mal.reason, mal.metadata, mal.ip_address, mal.user_agent, mal.channel_id, mal.created_at,
			u.id, u.twitch_id, u.username, u.display_name, u.email, u.avatar_url,
			u.bio, u.karma_points, u.role, u.is_banned, u.created_at, u.updated_at, u.last_login_at
		FROM moderation_audit_logs mal
		JOIN users u ON mal.moderator_id = u.id
		%s
		ORDER BY mal.created_at DESC
		LIMIT %s OFFSET %s`, whereClause, utils.SQLPlaceholder(placeholderIndex), utils.SQLPlaceholder(placeholderIndex+1))

	args = append(args, limit, offset)

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var logs []*models.ModerationAuditLogWithUser
	for rows.Next() {
		var log models.ModerationAuditLogWithUser
		var user models.User
		var metadataJSON []byte

		err := rows.Scan(
			&log.ID,
			&log.Action,
			&log.EntityType,
			&log.EntityID,
			&log.ModeratorID,
			&log.Reason,
			&metadataJSON,
			&log.IPAddress,
			&log.UserAgent,
			&log.ChannelID,
			&log.CreatedAt,
			&user.ID,
			&user.TwitchID,
			&user.Username,
			&user.DisplayName,
			&user.Email,
			&user.AvatarURL,
			&user.Bio,
			&user.KarmaPoints,
			&user.Role,
			&user.IsBanned,
			&user.CreatedAt,
			&user.UpdatedAt,
			&user.LastLoginAt,
		)
		if err != nil {
			return nil, 0, err
		}

		// Unmarshal metadata
		if metadataJSON != nil {
			if err := json.Unmarshal(metadataJSON, &log.Metadata); err != nil {
				return nil, 0, fmt.Errorf("failed to unmarshal metadata: %w", err)
			}
		}

		log.Moderator = &user
		logs = append(logs, &log)
	}

	return logs, total, rows.Err()
}

// AuditLogFilters represents filters for querying audit logs
type AuditLogFilters struct {
	ModeratorID *uuid.UUID
	Action      string
	EntityType  string
	EntityID    *uuid.UUID
	ChannelID   *uuid.UUID
	StartDate   *time.Time
	EndDate     *time.Time
	Search      string // Search term for filtering by reason
}

// Export retrieves all audit logs matching filters for export (no pagination)
func (r *AuditLogRepository) Export(ctx context.Context, filters AuditLogFilters) ([]*models.ModerationAuditLogWithUser, error) {
	// Build query with filters
	whereClause := "WHERE 1=1"
	args := []interface{}{}
	placeholderIndex := 1

	if filters.ModeratorID != nil {
		whereClause += fmt.Sprintf(" AND mal.moderator_id = %s", utils.SQLPlaceholder(placeholderIndex))
		args = append(args, *filters.ModeratorID)
		placeholderIndex++
	}

	if filters.Action != "" {
		whereClause += fmt.Sprintf(" AND mal.action = %s", utils.SQLPlaceholder(placeholderIndex))
		args = append(args, filters.Action)
		placeholderIndex++
	}

	if filters.EntityType != "" {
		whereClause += fmt.Sprintf(" AND mal.entity_type = %s", utils.SQLPlaceholder(placeholderIndex))
		args = append(args, filters.EntityType)
		placeholderIndex++
	}

	if filters.EntityID != nil {
		whereClause += fmt.Sprintf(" AND mal.entity_id = %s", utils.SQLPlaceholder(placeholderIndex))
		args = append(args, *filters.EntityID)
		placeholderIndex++
	}

	if filters.ChannelID != nil {
		whereClause += fmt.Sprintf(" AND mal.channel_id = %s", utils.SQLPlaceholder(placeholderIndex))
		args = append(args, *filters.ChannelID)
		placeholderIndex++
	}

	if filters.StartDate != nil {
		whereClause += fmt.Sprintf(" AND mal.created_at >= %s", utils.SQLPlaceholder(placeholderIndex))
		args = append(args, *filters.StartDate)
		placeholderIndex++
	}

	if filters.EndDate != nil {
		whereClause += fmt.Sprintf(" AND mal.created_at <= %s", utils.SQLPlaceholder(placeholderIndex))
		args = append(args, *filters.EndDate)
		placeholderIndex++
	}

	if filters.Search != "" {
		whereClause += fmt.Sprintf(" AND mal.reason ILIKE %s", utils.SQLPlaceholder(placeholderIndex))
		args = append(args, "%"+filters.Search+"%")
		placeholderIndex++
	}

	// Get logs with moderator info (no limit)
	query := fmt.Sprintf(`
		SELECT
			mal.id, mal.action, mal.entity_type, mal.entity_id, mal.moderator_id,
			mal.reason, mal.metadata, mal.ip_address, mal.user_agent, mal.channel_id, mal.created_at,
			u.id, u.twitch_id, u.username, u.display_name, u.email, u.avatar_url,
			u.bio, u.karma_points, u.role, u.is_banned, u.created_at, u.updated_at, u.last_login_at
		FROM moderation_audit_logs mal
		JOIN users u ON mal.moderator_id = u.id
		%s
		ORDER BY mal.created_at DESC`, whereClause)

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []*models.ModerationAuditLogWithUser
	for rows.Next() {
		var log models.ModerationAuditLogWithUser
		var user models.User
		var metadataJSON []byte

		err := rows.Scan(
			&log.ID,
			&log.Action,
			&log.EntityType,
			&log.EntityID,
			&log.ModeratorID,
			&log.Reason,
			&metadataJSON,
			&log.IPAddress,
			&log.UserAgent,
			&log.ChannelID,
			&log.CreatedAt,
			&user.ID,
			&user.TwitchID,
			&user.Username,
			&user.DisplayName,
			&user.Email,
			&user.AvatarURL,
			&user.Bio,
			&user.KarmaPoints,
			&user.Role,
			&user.IsBanned,
			&user.CreatedAt,
			&user.UpdatedAt,
			&user.LastLoginAt,
		)
		if err != nil {
			return nil, err
		}

		// Unmarshal metadata
		if metadataJSON != nil {
			if err := json.Unmarshal(metadataJSON, &log.Metadata); err != nil {
				return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
			}
		}

		log.Moderator = &user
		logs = append(logs, &log)
	}

	return logs, rows.Err()
}

// GetByID retrieves a single audit log entry by ID
func (r *AuditLogRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.ModerationAuditLogWithUser, error) {
	query := `
		SELECT
			mal.id, mal.action, mal.entity_type, mal.entity_id, mal.moderator_id,
			mal.reason, mal.metadata, mal.ip_address, mal.user_agent, mal.channel_id, mal.created_at,
			u.id, u.twitch_id, u.username, u.display_name, u.email, u.avatar_url,
			u.bio, u.karma_points, u.role, u.is_banned, u.created_at, u.updated_at, u.last_login_at
		FROM moderation_audit_logs mal
		JOIN users u ON mal.moderator_id = u.id
		WHERE mal.id = $1`

	var log models.ModerationAuditLogWithUser
	var user models.User
	var metadataJSON []byte

	err := r.db.QueryRow(ctx, query, id).Scan(
		&log.ID,
		&log.Action,
		&log.EntityType,
		&log.EntityID,
		&log.ModeratorID,
		&log.Reason,
		&metadataJSON,
		&log.IPAddress,
		&log.UserAgent,
		&log.ChannelID,
		&log.CreatedAt,
		&user.ID,
		&user.TwitchID,
		&user.Username,
		&user.DisplayName,
		&user.Email,
		&user.AvatarURL,
		&user.Bio,
		&user.KarmaPoints,
		&user.Role,
		&user.IsBanned,
		&user.CreatedAt,
		&user.UpdatedAt,
		&user.LastLoginAt,
	)
	if err != nil {
		return nil, err
	}

	// Unmarshal metadata
	if metadataJSON != nil {
		if err := json.Unmarshal(metadataJSON, &log.Metadata); err != nil {
			return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
		}
	}

	log.Moderator = &user
	return &log, nil
}

// GetByEntityID returns audit logs for a specific entity with optional type filter
func (r *AuditLogRepository) GetByEntityID(ctx context.Context, entityID uuid.UUID, entityType string, limit, offset int) ([]*models.ModerationAuditLog, error) {
	whereClause := "WHERE entity_id = $1"
	args := []interface{}{entityID}
	placeholderIndex := 2

	if entityType != "" {
		whereClause += fmt.Sprintf(" AND entity_type = %s", utils.SQLPlaceholder(placeholderIndex))
		args = append(args, entityType)
		placeholderIndex++
	}

	query := fmt.Sprintf(`
		SELECT id, action, entity_type, entity_id, moderator_id, reason, metadata,
		       ip_address, user_agent, channel_id, created_at
		FROM moderation_audit_logs
		%s
		ORDER BY created_at DESC
		LIMIT %s OFFSET %s`, whereClause, utils.SQLPlaceholder(placeholderIndex), utils.SQLPlaceholder(placeholderIndex+1))

	args = append(args, limit, offset)

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []*models.ModerationAuditLog
	for rows.Next() {
		var log models.ModerationAuditLog
		var metadataJSON []byte

		err := rows.Scan(
			&log.ID,
			&log.Action,
			&log.EntityType,
			&log.EntityID,
			&log.ModeratorID,
			&log.Reason,
			&metadataJSON,
			&log.IPAddress,
			&log.UserAgent,
			&log.ChannelID,
			&log.CreatedAt,
		)
		if err != nil {
			return nil, err
		}

		if metadataJSON != nil {
			if err := json.Unmarshal(metadataJSON, &log.Metadata); err != nil {
				return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
			}
		}

		logs = append(logs, &log)
	}

	return logs, rows.Err()
}
