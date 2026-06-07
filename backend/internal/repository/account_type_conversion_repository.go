package repository

import (
	"context"
	"encoding/json"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"git.subcult.tv/subculture-collective/clpr/internal/models"
)

// AccountTypeConversionRepository handles account type conversion database operations
type AccountTypeConversionRepository struct {
	db *pgxpool.Pool
}

// NewAccountTypeConversionRepository creates a new account type conversion repository
func NewAccountTypeConversionRepository(db *pgxpool.Pool) *AccountTypeConversionRepository {
	return &AccountTypeConversionRepository{db: db}
}

// Create records a new account type conversion
func (r *AccountTypeConversionRepository) Create(ctx context.Context, conversion *models.AccountTypeConversion) error {
	query := `
		INSERT INTO account_type_conversions (
			id, user_id, old_type, new_type, reason, converted_by, metadata
		) VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING converted_at, created_at
	`

	var metadataJSON []byte
	var err error
	if conversion.Metadata != nil {
		metadataJSON, err = json.Marshal(conversion.Metadata)
		if err != nil {
			return err
		}
	}

	err = r.db.QueryRow(
		ctx, query,
		conversion.ID,
		conversion.UserID,
		conversion.OldType,
		conversion.NewType,
		conversion.Reason,
		conversion.ConvertedBy,
		metadataJSON,
	).Scan(&conversion.ConvertedAt, &conversion.CreatedAt)

	return err
}

// GetByUserID retrieves all conversions for a user
func (r *AccountTypeConversionRepository) GetByUserID(ctx context.Context, userID uuid.UUID, limit, offset int) ([]models.AccountTypeConversion, error) {
	query := `
		SELECT 
			id, user_id, old_type, new_type, reason, converted_by,
			converted_at, metadata, created_at
		FROM account_type_conversions
		WHERE user_id = $1
		ORDER BY converted_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.Query(ctx, query, userID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var conversions []models.AccountTypeConversion
	for rows.Next() {
		var conversion models.AccountTypeConversion
		var metadataJSON []byte

		err := rows.Scan(
			&conversion.ID,
			&conversion.UserID,
			&conversion.OldType,
			&conversion.NewType,
			&conversion.Reason,
			&conversion.ConvertedBy,
			&conversion.ConvertedAt,
			&metadataJSON,
			&conversion.CreatedAt,
		)
		if err != nil {
			return nil, err
		}

		if len(metadataJSON) > 0 {
			err = json.Unmarshal(metadataJSON, &conversion.Metadata)
			if err != nil {
				return nil, err
			}
		}

		conversions = append(conversions, conversion)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return conversions, nil
}

// GetLatestByUserID retrieves the most recent conversion for a user
func (r *AccountTypeConversionRepository) GetLatestByUserID(ctx context.Context, userID uuid.UUID) (*models.AccountTypeConversion, error) {
	query := `
		SELECT 
			id, user_id, old_type, new_type, reason, converted_by,
			converted_at, metadata, created_at
		FROM account_type_conversions
		WHERE user_id = $1
		ORDER BY converted_at DESC
		LIMIT 1
	`

	var conversion models.AccountTypeConversion
	var metadataJSON []byte

	err := r.db.QueryRow(ctx, query, userID).Scan(
		&conversion.ID,
		&conversion.UserID,
		&conversion.OldType,
		&conversion.NewType,
		&conversion.Reason,
		&conversion.ConvertedBy,
		&conversion.ConvertedAt,
		&metadataJSON,
		&conversion.CreatedAt,
	)
	if err != nil {
		return nil, err
	}

	if len(metadataJSON) > 0 {
		err = json.Unmarshal(metadataJSON, &conversion.Metadata)
		if err != nil {
			return nil, err
		}
	}

	return &conversion, nil
}

// GetByID retrieves a conversion by ID
func (r *AccountTypeConversionRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.AccountTypeConversion, error) {
	query := `
		SELECT 
			id, user_id, old_type, new_type, reason, converted_by,
			converted_at, metadata, created_at
		FROM account_type_conversions
		WHERE id = $1
	`

	var conversion models.AccountTypeConversion
	var metadataJSON []byte

	err := r.db.QueryRow(ctx, query, id).Scan(
		&conversion.ID,
		&conversion.UserID,
		&conversion.OldType,
		&conversion.NewType,
		&conversion.Reason,
		&conversion.ConvertedBy,
		&conversion.ConvertedAt,
		&metadataJSON,
		&conversion.CreatedAt,
	)
	if err != nil {
		return nil, err
	}

	if len(metadataJSON) > 0 {
		err = json.Unmarshal(metadataJSON, &conversion.Metadata)
		if err != nil {
			return nil, err
		}
	}

	return &conversion, nil
}

// CountByUserID returns the total number of conversions for a user
func (r *AccountTypeConversionRepository) CountByUserID(ctx context.Context, userID uuid.UUID) (int, error) {
	query := `
		SELECT COUNT(*)
		FROM account_type_conversions
		WHERE user_id = $1
	`

	var count int
	err := r.db.QueryRow(ctx, query, userID).Scan(&count)
	return count, err
}

// GetRecentConversions retrieves recent conversions across all users (for admin)
func (r *AccountTypeConversionRepository) GetRecentConversions(ctx context.Context, limit, offset int) ([]models.AccountTypeConversion, error) {
	query := `
		SELECT 
			id, user_id, old_type, new_type, reason, converted_by,
			converted_at, metadata, created_at
		FROM account_type_conversions
		ORDER BY converted_at DESC
		LIMIT $1 OFFSET $2
	`

	rows, err := r.db.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var conversions []models.AccountTypeConversion
	for rows.Next() {
		var conversion models.AccountTypeConversion
		var metadataJSON []byte

		err := rows.Scan(
			&conversion.ID,
			&conversion.UserID,
			&conversion.OldType,
			&conversion.NewType,
			&conversion.Reason,
			&conversion.ConvertedBy,
			&conversion.ConvertedAt,
			&metadataJSON,
			&conversion.CreatedAt,
		)
		if err != nil {
			return nil, err
		}

		if len(metadataJSON) > 0 {
			err = json.Unmarshal(metadataJSON, &conversion.Metadata)
			if err != nil {
				return nil, err
			}
		}

		conversions = append(conversions, conversion)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return conversions, nil
}

// CountTotal returns the total number of conversions
func (r *AccountTypeConversionRepository) CountTotal(ctx context.Context) (int, error) {
	query := `
		SELECT COUNT(*)
		FROM account_type_conversions
	`

	var count int
	err := r.db.QueryRow(ctx, query).Scan(&count)
	return count, err
}

// CountByAccountType returns the current count of users with a specific account type
func (r *AccountTypeConversionRepository) CountByAccountType(ctx context.Context, accountType string) (int, error) {
	query := `
		SELECT COUNT(*)
		FROM users
		WHERE account_type = $1
	`

	var count int
	err := r.db.QueryRow(ctx, query, accountType).Scan(&count)
	return count, err
}
