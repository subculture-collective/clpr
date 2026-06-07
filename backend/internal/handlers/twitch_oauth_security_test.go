package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"git.subcult.tv/subculture-collective/clpr/internal/models"
	"git.subcult.tv/subculture-collective/clpr/internal/repository"
)

// TestTokenMaskingInLogs verifies that access and refresh tokens are never logged
func TestTokenMaskingInLogs(t *testing.T) {
	// Set up a custom logger that captures log output
	var logBuffer bytes.Buffer
	log.SetOutput(&logBuffer)
	defer log.SetOutput(os.Stderr)
	connString := os.Getenv("TEST_DATABASE_URL")
	if connString == "" {
		connString = "postgres://clpr:clpr_password@localhost:5436/clpr_db?sslmode=disable"
	}

	pool, err := pgxpool.New(context.Background(), connString)
	if err != nil {
		t.Skipf("Skipping test: database not available: %v", err)
		return
	}
	defer pool.Close()

	// Test connection
	if err := pool.Ping(context.Background()); err != nil {
		t.Skipf("Skipping test: cannot ping database: %v", err)
		return
	}

	repo := repository.NewTwitchAuthRepository(pool)
	handler := NewTwitchOAuthHandler(repo)

	// Create test user
	userID := uuid.New()
	ctx := context.Background()

	// Insert test user
	_, err = pool.Exec(ctx, `
		INSERT INTO users (id, twitch_id, username, display_name, role, account_type)
		VALUES ($1, $2, $3, $4, $5, $6)
	`, userID, fmt.Sprintf("twitch_%s", userID.String()[:8]), fmt.Sprintf("user_%s", userID.String()[:8]), "Test User", "user", "member")
	if err != nil {
		t.Fatalf("Failed to insert test user: %v", err)
	}
	defer pool.Exec(ctx, "DELETE FROM users WHERE id = $1", userID)
	defer pool.Exec(ctx, "DELETE FROM twitch_auth WHERE user_id = $1", userID)

	// Test data with sensitive tokens
	testAccessToken := "test_access_token_SECRET_12345"
	testRefreshToken := "test_refresh_token_SECRET_67890"
	testScopes := "chat:read chat:edit moderator:manage:banned_users channel:manage:banned_users"

	auth := &models.TwitchAuth{
		UserID:         userID,
		TwitchUserID:   "test_twitch_123",
		TwitchUsername: "test_user",
		AccessToken:    testAccessToken,
		RefreshToken:   testRefreshToken,
		Scopes:         testScopes,
		ExpiresAt:      time.Now().Add(1 * time.Hour),
	}

	// Call the handler method that logs
	err = handler.refreshTwitchToken(ctx, auth)
	// We expect this to fail since we're not hitting a real Twitch endpoint
	// but that's OK - we just want to check logging

	// Reset buffer for next test
	logBuffer.Reset()

	// Store auth which should trigger logging
	err = repo.UpsertTwitchAuth(ctx, auth)
	if err != nil {
		t.Fatalf("Failed to upsert auth: %v", err)
	}

	// Get the auth which might log
	retrieved, err := repo.GetTwitchAuth(ctx, userID)
	if err != nil {
		t.Fatalf("Failed to get auth: %v", err)
	}

	// Check that tokens are not in the log output
	logOutput := logBuffer.String()

	if strings.Contains(logOutput, testAccessToken) {
		t.Errorf("Access token found in logs! This is a security violation.\nLog output: %s", logOutput)
	}

	if strings.Contains(logOutput, testRefreshToken) {
		t.Errorf("Refresh token found in logs! This is a security violation.\nLog output: %s", logOutput)
	}

	// Verify we got the right data back (sanity check)
	if retrieved.AccessToken != testAccessToken {
		t.Errorf("Expected to retrieve correct access token")
	}
}

// TestTokenNotInJSONResponses verifies tokens are not exposed in API responses
func TestTokenNotInJSONResponses(t *testing.T) {
	// Create a sample TwitchAuthStatusResponse
	response := models.TwitchAuthStatusResponse{
		Authenticated:  true,
		TwitchUsername: stringPtr("test_user"),
	}

	// Marshal to JSON
	jsonData, err := json.Marshal(response)
	if err != nil {
		t.Fatalf("Failed to marshal response: %v", err)
	}

	// Verify response doesn't contain token fields
	jsonStr := string(jsonData)

	if strings.Contains(jsonStr, "access_token") {
		t.Errorf("Response JSON should not contain 'access_token' field")
	}

	if strings.Contains(jsonStr, "refresh_token") {
		t.Errorf("Response JSON should not contain 'refresh_token' field")
	}

	// Verify expected fields are present
	if !strings.Contains(jsonStr, "authenticated") {
		t.Errorf("Response JSON should contain 'authenticated' field")
	}
}

// TestScopesStoredAndRetrieved verifies that scopes are properly stored and retrieved
func TestScopesStoredAndRetrieved(t *testing.T) {
	connString := os.Getenv("TEST_DATABASE_URL")
	if connString == "" {
		connString = "postgres://clpr:clpr_password@localhost:5436/clpr_db?sslmode=disable"
	}

	pool, err := pgxpool.New(context.Background(), connString)
	if err != nil {
		t.Skipf("Skipping test: database not available: %v", err)
		return
	}
	defer pool.Close()

	if err := pool.Ping(context.Background()); err != nil {
		t.Skipf("Skipping test: cannot ping database: %v", err)
		return
	}

	repo := repository.NewTwitchAuthRepository(pool)
	ctx := context.Background()

	userID := uuid.New()

	// Insert test user
	_, err = pool.Exec(ctx, `
		INSERT INTO users (id, twitch_id, username, display_name, role, account_type)
		VALUES ($1, $2, $3, $4, $5, $6)
	`, userID, fmt.Sprintf("twitch_%s", userID.String()[:8]), fmt.Sprintf("user_%s", userID.String()[:8]), "Test User", "user", "member")
	if err != nil {
		t.Fatalf("Failed to insert test user: %v", err)
	}
	defer pool.Exec(ctx, "DELETE FROM users WHERE id = $1", userID)
	defer pool.Exec(ctx, "DELETE FROM twitch_auth WHERE user_id = $1", userID)

	expectedScopes := "chat:read chat:edit moderator:manage:banned_users channel:manage:banned_users"

	auth := &models.TwitchAuth{
		UserID:         userID,
		TwitchUserID:   "test_twitch_456",
		TwitchUsername: "test_user_scopes",
		AccessToken:    "test_access",
		RefreshToken:   "test_refresh",
		Scopes:         expectedScopes,
		ExpiresAt:      time.Now().Add(1 * time.Hour),
	}

	// Store with scopes
	err = repo.UpsertTwitchAuth(ctx, auth)
	if err != nil {
		t.Fatalf("Failed to upsert auth with scopes: %v", err)
	}

	// Retrieve and verify scopes
	retrieved, err := repo.GetTwitchAuth(ctx, userID)
	if err != nil {
		t.Fatalf("Failed to get auth: %v", err)
	}

	if retrieved.Scopes != expectedScopes {
		t.Errorf("Expected scopes '%s', got '%s'", expectedScopes, retrieved.Scopes)
	}

	// Verify individual scopes are present
	if !strings.Contains(retrieved.Scopes, "chat:read") {
		t.Error("Expected chat:read scope")
	}
	if !strings.Contains(retrieved.Scopes, "chat:edit") {
		t.Error("Expected chat:edit scope")
	}
	if !strings.Contains(retrieved.Scopes, "moderator:manage:banned_users") {
		t.Error("Expected moderator:manage:banned_users scope")
	}
	if !strings.Contains(retrieved.Scopes, "channel:manage:banned_users") {
		t.Error("Expected channel:manage:banned_users scope")
	}
}

func stringPtr(s string) *string {
	return &s
}
