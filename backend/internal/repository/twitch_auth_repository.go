package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"git.subcult.tv/subculture-collective/clpr/internal/models"
)

// TwitchAuthRepository handles Twitch OAuth authentication data persistence
type TwitchAuthRepository struct {
	pool *pgxpool.Pool
}

// NewTwitchAuthRepository creates a new Twitch auth repository
func NewTwitchAuthRepository(pool *pgxpool.Pool) *TwitchAuthRepository {
	return &TwitchAuthRepository{pool: pool}
}

// UpsertTwitchAuth inserts or updates Twitch OAuth credentials
func (r *TwitchAuthRepository) UpsertTwitchAuth(ctx context.Context, auth *models.TwitchAuth) error {
	query := `
		INSERT INTO twitch_auth (user_id, twitch_user_id, twitch_username, access_token, refresh_token, scopes, expires_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (user_id) DO UPDATE
		SET 
			twitch_user_id = EXCLUDED.twitch_user_id,
			twitch_username = EXCLUDED.twitch_username,
			access_token = EXCLUDED.access_token,
			refresh_token = EXCLUDED.refresh_token,
			scopes = EXCLUDED.scopes,
			expires_at = EXCLUDED.expires_at,
			updated_at = NOW()
		RETURNING created_at, updated_at
	`

	err := r.pool.QueryRow(
		ctx,
		query,
		auth.UserID,
		auth.TwitchUserID,
		auth.TwitchUsername,
		auth.AccessToken,
		auth.RefreshToken,
		auth.Scopes,
		auth.ExpiresAt,
	).Scan(&auth.CreatedAt, &auth.UpdatedAt)

	return err
}

// GetTwitchAuth retrieves Twitch OAuth credentials for a user
func (r *TwitchAuthRepository) GetTwitchAuth(ctx context.Context, userID uuid.UUID) (*models.TwitchAuth, error) {
	query := `
		SELECT user_id, twitch_user_id, twitch_username, access_token, refresh_token, scopes, expires_at, created_at, updated_at
		FROM twitch_auth
		WHERE user_id = $1
	`

	auth := &models.TwitchAuth{}
	err := r.pool.QueryRow(ctx, query, userID).Scan(
		&auth.UserID,
		&auth.TwitchUserID,
		&auth.TwitchUsername,
		&auth.AccessToken,
		&auth.RefreshToken,
		&auth.Scopes,
		&auth.ExpiresAt,
		&auth.CreatedAt,
		&auth.UpdatedAt,
	)

	if err == pgx.ErrNoRows {
		return nil, nil
	}

	return auth, err
}

// DeleteTwitchAuth removes Twitch OAuth credentials for a user
func (r *TwitchAuthRepository) DeleteTwitchAuth(ctx context.Context, userID uuid.UUID) error {
	query := `DELETE FROM twitch_auth WHERE user_id = $1`
	_, err := r.pool.Exec(ctx, query, userID)
	return err
}

// RefreshToken updates the access token and expiration time
func (r *TwitchAuthRepository) RefreshToken(ctx context.Context, userID uuid.UUID, newAccessToken string, newRefreshToken string, scopes string, expiresAt time.Time) error {
	query := `
		UPDATE twitch_auth
		SET access_token = $2, refresh_token = $3, scopes = $4, expires_at = $5, updated_at = NOW()
		WHERE user_id = $1
	`

	_, err := r.pool.Exec(ctx, query, userID, newAccessToken, newRefreshToken, scopes, expiresAt)
	return err
}

// IsTokenExpired checks if a token is expired or about to expire (within 5 minutes)
func (r *TwitchAuthRepository) IsTokenExpired(auth *models.TwitchAuth) bool {
	// Consider token expired if it expires within the next 5 minutes
	return time.Now().Add(5 * time.Minute).After(auth.ExpiresAt)
}
