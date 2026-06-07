package repository

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"git.subcult.tv/subculture-collective/clpr/internal/models"
)

func setupTwitchAuthTestDB(t *testing.T) (*pgxpool.Pool, func()) {
	t.Helper()

	// Build connection string from environment or defaults
	connString := os.Getenv("TEST_DATABASE_URL")
	if connString == "" {
		connString = "postgres://clpr:clpr_password@localhost:5437/clpr_test?sslmode=disable"
	}

	pool, err := pgxpool.New(context.Background(), connString)
	if err != nil {
		t.Skipf("Skipping test: database not available: %v", err)
		return nil, func() {}
	}

	// Test connection
	if err := pool.Ping(context.Background()); err != nil {
		pool.Close()
		t.Skipf("Skipping test: cannot ping database: %v", err)
		return nil, func() {}
	}

	cleanup := func() {
		pool.Close()
	}

	return pool, cleanup
}

func insertTestUser(t *testing.T, pool *pgxpool.Pool, userID uuid.UUID) {
	t.Helper()

	ctx := context.Background()
	prefix := userID.String()[:8]

	_, err := pool.Exec(ctx, `
		INSERT INTO users (id, twitch_id, username, display_name, role, account_type)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (id) DO NOTHING
	`, userID, fmt.Sprintf("twitch_%s", prefix), fmt.Sprintf("user_%s", prefix), fmt.Sprintf("Test User %s", prefix), "user", "member")
	if err != nil {
		t.Fatalf("Failed to insert test user: %v", err)
	}

	t.Cleanup(func() {
		_, _ = pool.Exec(ctx, "DELETE FROM users WHERE id = $1", userID)
	})
}

func TestTwitchAuthRepository_UpsertTwitchAuth(t *testing.T) {
	pool, cleanup := setupTwitchAuthTestDB(t)
	if pool == nil {
		return
	}
	defer cleanup()

	repo := NewTwitchAuthRepository(pool)
	ctx := context.Background()

	userID := uuid.New()
	insertTestUser(t, pool, userID)
	twitchUserID := fmt.Sprintf("tw_%s", userID.String()[:8])
	twitchUsername := fmt.Sprintf("testuser_%s", userID.String()[:8])
	expiresAt := time.Now().Add(4 * time.Hour)

	auth := &models.TwitchAuth{
		UserID:         userID,
		TwitchUserID:   twitchUserID,
		TwitchUsername: twitchUsername,
		AccessToken:    "test_access_token",
		RefreshToken:   "test_refresh_token",
		Scopes:         "chat:read chat:edit moderator:manage:banned_users channel:manage:banned_users",
		ExpiresAt:      expiresAt,
	}

	// Use t.Cleanup for automatic cleanup
	t.Cleanup(func() {
		_ = repo.DeleteTwitchAuth(ctx, userID)
	})

	// Insert
	err := repo.UpsertTwitchAuth(ctx, auth)
	if err != nil {
		t.Fatalf("Failed to insert twitch auth: %v", err)
	}

	// Verify fields were populated
	if auth.CreatedAt.IsZero() {
		t.Error("CreatedAt should be set")
	}
	if auth.UpdatedAt.IsZero() {
		t.Error("UpdatedAt should be set")
	}

	// Update
	auth.AccessToken = "new_access_token"
	err = repo.UpsertTwitchAuth(ctx, auth)
	if err != nil {
		t.Fatalf("Failed to update twitch auth: %v", err)
	}
}

func TestTwitchAuthRepository_GetTwitchAuth(t *testing.T) {
	pool, cleanup := setupTwitchAuthTestDB(t)
	if pool == nil {
		return
	}
	defer cleanup()

	repo := NewTwitchAuthRepository(pool)
	ctx := context.Background()

	userID := uuid.New()
	insertTestUser(t, pool, userID)
	twitchUserID := fmt.Sprintf("tw_%s", userID.String()[:8])
	twitchUsername := fmt.Sprintf("testuser_%s", userID.String()[:8])
	expiresAt := time.Now().Add(4 * time.Hour)
	scopes := "chat:read chat:edit moderator:manage:banned_users channel:manage:banned_users"

	auth := &models.TwitchAuth{
		UserID:         userID,
		TwitchUserID:   twitchUserID,
		TwitchUsername: twitchUsername,
		AccessToken:    "test_access_token_2",
		RefreshToken:   "test_refresh_token_2",
		Scopes:         scopes,
		ExpiresAt:      expiresAt,
	}

	// Use t.Cleanup for automatic cleanup
	t.Cleanup(func() {
		_ = repo.DeleteTwitchAuth(ctx, userID)
	})

	// Insert
	err := repo.UpsertTwitchAuth(ctx, auth)
	if err != nil {
		t.Fatalf("Failed to insert twitch auth: %v", err)
	}

	// Get
	retrieved, err := repo.GetTwitchAuth(ctx, userID)
	if err != nil {
		t.Fatalf("Failed to get twitch auth: %v", err)
	}

	if retrieved == nil {
		t.Fatal("Expected to retrieve twitch auth, got nil")
	}

	if retrieved.UserID != userID {
		t.Errorf("Expected UserID %v, got %v", userID, retrieved.UserID)
	}
	if retrieved.TwitchUserID != twitchUserID {
		t.Errorf("Expected TwitchUserID '%s', got '%s'", twitchUserID, retrieved.TwitchUserID)
	}
	if retrieved.TwitchUsername != twitchUsername {
		t.Errorf("Expected TwitchUsername '%s', got '%s'", twitchUsername, retrieved.TwitchUsername)
	}
	if retrieved.AccessToken != "test_access_token_2" {
		t.Errorf("Expected AccessToken 'test_access_token_2', got '%s'", retrieved.AccessToken)
	}
	if retrieved.Scopes != scopes {
		t.Errorf("Expected Scopes '%s', got '%s'", scopes, retrieved.Scopes)
	}

	// Test non-existent user
	nonExistentID := uuid.New()
	retrieved, err = repo.GetTwitchAuth(ctx, nonExistentID)
	if err != nil {
		t.Fatalf("Error should be nil for non-existent user: %v", err)
	}
	if retrieved != nil {
		t.Error("Expected nil for non-existent user")
	}
}

func TestTwitchAuthRepository_RefreshToken(t *testing.T) {
	pool, cleanup := setupTwitchAuthTestDB(t)
	if pool == nil {
		return
	}
	defer cleanup()

	repo := NewTwitchAuthRepository(pool)
	ctx := context.Background()

	userID := uuid.New()
	insertTestUser(t, pool, userID)
	twitchUserID := fmt.Sprintf("tw_%s", userID.String()[:8])
	twitchUsername := fmt.Sprintf("testuser_%s", userID.String()[:8])
	expiresAt := time.Now().Add(4 * time.Hour)

	auth := &models.TwitchAuth{
		UserID:         userID,
		TwitchUserID:   twitchUserID,
		TwitchUsername: twitchUsername,
		AccessToken:    "old_access_token",
		RefreshToken:   "old_refresh_token",
		Scopes:         "chat:read chat:edit",
		ExpiresAt:      expiresAt,
	}

	// Use t.Cleanup for automatic cleanup
	t.Cleanup(func() {
		_ = repo.DeleteTwitchAuth(ctx, userID)
	})

	// Insert
	err := repo.UpsertTwitchAuth(ctx, auth)
	if err != nil {
		t.Fatalf("Failed to insert twitch auth: %v", err)
	}

	// Refresh token with new scopes
	newExpiresAt := time.Now().Add(8 * time.Hour)
	newScopes := "chat:read chat:edit moderator:manage:banned_users channel:manage:banned_users"
	err = repo.RefreshToken(ctx, userID, "new_access_token", "new_refresh_token", newScopes, newExpiresAt)
	if err != nil {
		t.Fatalf("Failed to refresh token: %v", err)
	}

	// Verify update
	retrieved, err := repo.GetTwitchAuth(ctx, userID)
	if err != nil {
		t.Fatalf("Failed to get twitch auth: %v", err)
	}

	if retrieved.AccessToken != "new_access_token" {
		t.Errorf("Expected AccessToken 'new_access_token', got '%s'", retrieved.AccessToken)
	}
	if retrieved.RefreshToken != "new_refresh_token" {
		t.Errorf("Expected RefreshToken 'new_refresh_token', got '%s'", retrieved.RefreshToken)
	}
	if retrieved.Scopes != newScopes {
		t.Errorf("Expected Scopes '%s', got '%s'", newScopes, retrieved.Scopes)
	}
}

func TestTwitchAuthRepository_IsTokenExpired(t *testing.T) {
	repo := &TwitchAuthRepository{}

	tests := []struct {
		name     string
		auth     *models.TwitchAuth
		expected bool
	}{
		{
			name: "token expires in 10 minutes - not expired",
			auth: &models.TwitchAuth{
				ExpiresAt: time.Now().Add(10 * time.Minute),
			},
			expected: false,
		},
		{
			name: "token expires in 3 minutes - considered expired",
			auth: &models.TwitchAuth{
				ExpiresAt: time.Now().Add(3 * time.Minute),
			},
			expected: true,
		},
		{
			name: "token already expired",
			auth: &models.TwitchAuth{
				ExpiresAt: time.Now().Add(-1 * time.Hour),
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := repo.IsTokenExpired(tt.auth)
			if result != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestTwitchAuthRepository_DeleteTwitchAuth(t *testing.T) {
	pool, cleanup := setupTwitchAuthTestDB(t)
	if pool == nil {
		return
	}
	defer cleanup()

	repo := NewTwitchAuthRepository(pool)
	ctx := context.Background()

	userID := uuid.New()
	insertTestUser(t, pool, userID)
	twitchUserID := fmt.Sprintf("tw_%s", userID.String()[:8])
	twitchUsername := fmt.Sprintf("testuser_%s", userID.String()[:8])
	expiresAt := time.Now().Add(4 * time.Hour)

	auth := &models.TwitchAuth{
		UserID:         userID,
		TwitchUserID:   twitchUserID,
		TwitchUsername: twitchUsername,
		AccessToken:    "test_access_token_4",
		RefreshToken:   "test_refresh_token_4",
		Scopes:         "chat:read chat:edit",
		ExpiresAt:      expiresAt,
	}

	// Insert
	err := repo.UpsertTwitchAuth(ctx, auth)
	if err != nil {
		t.Fatalf("Failed to insert twitch auth: %v", err)
	}

	// Delete
	err = repo.DeleteTwitchAuth(ctx, userID)
	if err != nil {
		t.Fatalf("Failed to delete twitch auth: %v", err)
	}

	// Verify deletion
	retrieved, err := repo.GetTwitchAuth(ctx, userID)
	if err != nil {
		t.Fatalf("Error checking deleted auth: %v", err)
	}
	if retrieved != nil {
		t.Error("Expected nil after deletion")
	}
}
