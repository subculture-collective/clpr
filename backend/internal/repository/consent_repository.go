package repository

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"git.subcult.tv/subculture-collective/clpr/internal/models"
)

var (
	// ErrConsentNotFound is returned when consent record is not found
	ErrConsentNotFound = errors.New("consent not found")

	// ConsentExpirationDuration is 12 months (365 days)
	ConsentExpirationDuration = 365 * 24 * time.Hour
)

// ConsentRepository handles cookie consent database operations
type ConsentRepository struct {
	db *pgxpool.Pool
}

// NewConsentRepository creates a new consent repository
func NewConsentRepository(db *pgxpool.Pool) *ConsentRepository {
	return &ConsentRepository{db: db}
}

// SaveConsent saves or updates user cookie consent preferences
func (r *ConsentRepository) SaveConsent(ctx context.Context, consent *models.CookieConsent, ipAddress, userAgent string) error {
	// Calculate expiration once to avoid race condition
	expiresAt := time.Now().Add(ConsentExpirationDuration)

	query := `
		INSERT INTO user_cookie_consents (
			user_id, essential, functional, analytics, advertising, 
			ip_address, user_agent, expires_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		ON CONFLICT (user_id) 
		DO UPDATE SET
			essential = EXCLUDED.essential,
			functional = EXCLUDED.functional,
			analytics = EXCLUDED.analytics,
			advertising = EXCLUDED.advertising,
			consent_date = NOW(),
			ip_address = EXCLUDED.ip_address,
			user_agent = EXCLUDED.user_agent,
			expires_at = EXCLUDED.expires_at,
			updated_at = NOW()
		RETURNING id, consent_date, created_at, updated_at
	`

	err := r.db.QueryRow(
		ctx, query,
		consent.UserID,
		consent.Essential,
		consent.Functional,
		consent.Analytics,
		consent.Advertising,
		ipAddress,
		userAgent,
		expiresAt,
	).Scan(&consent.ID, &consent.ConsentDate, &consent.CreatedAt, &consent.UpdatedAt)

	if err != nil {
		return err
	}

	consent.IPAddress = &ipAddress
	consent.UserAgent = &userAgent
	consent.ExpiresAt = expiresAt

	return nil
}

// GetConsent retrieves the current consent preferences for a user
func (r *ConsentRepository) GetConsent(ctx context.Context, userID uuid.UUID) (*models.CookieConsent, error) {
	query := `
		SELECT 
			id, user_id, essential, functional, analytics, advertising,
			consent_date, ip_address, user_agent, expires_at, created_at, updated_at
		FROM user_cookie_consents
		WHERE user_id = $1
		ORDER BY consent_date DESC
		LIMIT 1
	`

	var consent models.CookieConsent
	err := r.db.QueryRow(ctx, query, userID).Scan(
		&consent.ID,
		&consent.UserID,
		&consent.Essential,
		&consent.Functional,
		&consent.Analytics,
		&consent.Advertising,
		&consent.ConsentDate,
		&consent.IPAddress,
		&consent.UserAgent,
		&consent.ExpiresAt,
		&consent.CreatedAt,
		&consent.UpdatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, ErrConsentNotFound
		}
		return nil, err
	}

	return &consent, nil
}

// IsConsentExpired checks if the user's consent has expired
func (r *ConsentRepository) IsConsentExpired(ctx context.Context, userID uuid.UUID) (bool, error) {
	consent, err := r.GetConsent(ctx, userID)
	if err != nil {
		if errors.Is(err, ErrConsentNotFound) {
			return true, nil // No consent = expired
		}
		return true, err
	}

	return time.Now().After(consent.ExpiresAt), nil
}
