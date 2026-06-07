package repository

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"git.subcult.tv/subculture-collective/clpr/internal/models"
	"git.subcult.tv/subculture-collective/clpr/internal/utils"
)

// ContactRepository handles database operations for contact messages
type ContactRepository struct {
	db *pgxpool.Pool
}

// NewContactRepository creates a new contact repository
func NewContactRepository(db *pgxpool.Pool) *ContactRepository {
	return &ContactRepository{db: db}
}

// Create creates a new contact message
func (r *ContactRepository) Create(ctx context.Context, message *models.ContactMessage) error {
	query := `
		INSERT INTO contact_messages (
			id, user_id, email, category, subject, message,
			status, ip_address, user_agent, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`

	_, err := r.db.Exec(ctx, query,
		message.ID,
		message.UserID,
		message.Email,
		message.Category,
		message.Subject,
		message.Message,
		message.Status,
		message.IPAddress,
		message.UserAgent,
		message.CreatedAt,
		message.UpdatedAt,
	)

	return err
}

// List retrieves contact messages with filtering and pagination
func (r *ContactRepository) List(ctx context.Context, page, limit int, category, status string) ([]*models.ContactMessage, int, error) {
	// Build the WHERE clause
	whereClause := "WHERE 1=1"
	args := []interface{}{}
	argIndex := 1

	if category != "" {
		whereClause += fmt.Sprintf(" AND category = %s", utils.SQLPlaceholder(argIndex))
		args = append(args, category)
		argIndex++
	}

	if status != "" {
		whereClause += fmt.Sprintf(" AND status = %s", utils.SQLPlaceholder(argIndex))
		args = append(args, status)
		argIndex++
	}

	// Count total
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM contact_messages %s", whereClause)
	var total int
	err := r.db.QueryRow(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	// Calculate offset
	offset := (page - 1) * limit

	// Query messages
	query := fmt.Sprintf(`
		SELECT id, user_id, email, category, subject, message,
			status, ip_address, user_agent, created_at, updated_at
		FROM contact_messages
		%s
		ORDER BY created_at DESC
		LIMIT %s OFFSET %s
	`, whereClause, utils.SQLPlaceholder(argIndex), utils.SQLPlaceholder(argIndex+1))
	args = append(args, limit, offset)

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	messages := []*models.ContactMessage{}
	for rows.Next() {
		var msg models.ContactMessage
		err := rows.Scan(
			&msg.ID,
			&msg.UserID,
			&msg.Email,
			&msg.Category,
			&msg.Subject,
			&msg.Message,
			&msg.Status,
			&msg.IPAddress,
			&msg.UserAgent,
			&msg.CreatedAt,
			&msg.UpdatedAt,
		)
		if err != nil {
			return nil, 0, err
		}
		messages = append(messages, &msg)
	}

	return messages, total, rows.Err()
}

// UpdateStatus updates the status of a contact message
func (r *ContactRepository) UpdateStatus(ctx context.Context, messageID uuid.UUID, status string) error {
	query := `
		UPDATE contact_messages
		SET status = $1, updated_at = NOW()
		WHERE id = $2
	`

	result, err := r.db.Exec(ctx, query, status, messageID)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("contact message not found")
	}

	return nil
}
