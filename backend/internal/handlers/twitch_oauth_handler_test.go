package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"git.subcult.tv/subculture-collective/clpr/internal/models"
	"git.subcult.tv/subculture-collective/clpr/internal/repository"
)

func setupTwitchOAuthTestHandler(t *testing.T) (*TwitchOAuthHandler, *pgxpool.Pool, func()) {
	t.Helper()

	connString := os.Getenv("TEST_DATABASE_URL")
	if connString == "" {
		connString = "postgres://clpr:clpr_password@localhost:5436/clpr_db?sslmode=disable"
	}

	pool, err := pgxpool.New(context.Background(), connString)
	if err != nil {
		t.Skipf("Skipping test: database not available: %v", err)
		return nil, nil, func() {}
	}

	// Test connection
	if err := pool.Ping(context.Background()); err != nil {
		pool.Close()
		t.Skipf("Skipping test: cannot ping database: %v", err)
		return nil, nil, func() {}
	}

	repo := repository.NewTwitchAuthRepository(pool)
	handler := NewTwitchOAuthHandler(repo)

	cleanup := func() {
		pool.Close()
	}

	return handler, pool, cleanup
}

func insertTestUser(t *testing.T, pool *pgxpool.Pool, userID uuid.UUID) {
	t.Helper()

	ctx := context.Background()
	prefix := userID.String()[:8]

	_, err := pool.Exec(ctx, `
			INSERT INTO users (id, twitch_id, username, display_name, role, account_type)
			VALUES ($1, $2, $3, $4, $5, $6)
		`, userID, fmt.Sprintf("twitch_%s", prefix), fmt.Sprintf("user_%s", prefix), fmt.Sprintf("Test User %s", prefix), "user", "member")
	if err != nil {
		t.Fatalf("Failed to insert test user: %v", err)
	}

	t.Cleanup(func() {
		_, _ = pool.Exec(ctx, "DELETE FROM users WHERE id = $1", userID)
	})
}

func TestTwitchOAuthHandler_GetTwitchAuthStatus_NotAuthenticated(t *testing.T) {
	handler, _, cleanup := setupTwitchOAuthTestHandler(t)
	if handler == nil {
		return
	}
	defer cleanup()

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	// No user_id in context (not authenticated)
	handler.GetTwitchAuthStatus(c)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	var response models.TwitchAuthStatusResponse
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if response.Authenticated {
		t.Error("Expected authenticated to be false")
	}
}

func TestTwitchOAuthHandler_GetTwitchAuthStatus_NoTwitchAuth(t *testing.T) {
	handler, pool, cleanup := setupTwitchOAuthTestHandler(t)
	if handler == nil {
		return
	}
	defer cleanup()

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	// Set user_id but no Twitch auth exists
	userID := uuid.New()
	c.Set("user_id", userID)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/twitch/auth/status", nil)

	insertTestUser(t, pool, userID)

	handler.GetTwitchAuthStatus(c)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	var response models.TwitchAuthStatusResponse
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if response.Authenticated {
		t.Error("Expected authenticated to be false")
	}
}

func TestTwitchOAuthHandler_RevokeTwitchAuth_NotAuthenticated(t *testing.T) {
	handler, _, cleanup := setupTwitchOAuthTestHandler(t)
	if handler == nil {
		return
	}
	defer cleanup()

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	// No user_id in context
	handler.RevokeTwitchAuth(c)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status %d, got %d", http.StatusUnauthorized, w.Code)
	}
}

func TestTwitchOAuthHandler_RevokeTwitchAuth_Success(t *testing.T) {
	handler, pool, cleanup := setupTwitchOAuthTestHandler(t)
	if handler == nil {
		return
	}
	defer cleanup()

	gin.SetMode(gin.TestMode)

	// Create a context for the database operation
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	// Create a mock request first
	req, _ := http.NewRequest("DELETE", "/api/v1/twitch/auth", nil)
	c.Request = req

	// First, create a Twitch auth record
	userID := uuid.New()
	insertTestUser(t, pool, userID)
	twitchUserID := fmt.Sprintf("tw_%s", userID.String()[:8])
	twitchUsername := fmt.Sprintf("testuser_%s", userID.String()[:8])
	auth := &models.TwitchAuth{
		UserID:         userID,
		TwitchUserID:   twitchUserID,
		TwitchUsername: twitchUsername,
		AccessToken:    "test_token",
		RefreshToken:   "test_refresh",
		ExpiresAt:      time.Now().Add(4 * time.Hour),
	}

	err := handler.twitchAuthRepo.UpsertTwitchAuth(c.Request.Context(), auth)
	if err != nil {
		t.Fatalf("Failed to create test auth: %v", err)
	}

	// Now test revoke
	c.Set("user_id", userID)

	handler.RevokeTwitchAuth(c)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	// Verify auth was deleted
	retrieved, _ := handler.twitchAuthRepo.GetTwitchAuth(c.Request.Context(), userID)
	if retrieved != nil {
		t.Error("Expected auth to be deleted")
	}
}

func TestTwitchOAuthHandler_InitiateTwitchOAuth(t *testing.T) {
	handler, _, cleanup := setupTwitchOAuthTestHandler(t)
	if handler == nil {
		return
	}
	defer cleanup()

	// Set environment variables for testing
	os.Setenv("TWITCH_CLIENT_ID", "test_client_id")
	os.Setenv("TWITCH_REDIRECT_URI", "http://localhost:8080/api/v1/twitch/oauth/callback")
	defer func() {
		os.Unsetenv("TWITCH_CLIENT_ID")
		os.Unsetenv("TWITCH_REDIRECT_URI")
	}()

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	// Create a mock request
	req, _ := http.NewRequest("GET", "/api/v1/twitch/oauth/authorize", nil)
	c.Request = req

	handler.InitiateTwitchOAuth(c)

	// Should redirect
	if w.Code != http.StatusTemporaryRedirect {
		t.Errorf("Expected status %d, got %d", http.StatusTemporaryRedirect, w.Code)
	}

	location := w.Header().Get("Location")
	if location == "" {
		t.Error("Expected redirect location to be set")
	}

	// Verify redirect URL contains necessary parameters
	if !strings.Contains(location, "id.twitch.tv/oauth2/authorize") {
		t.Error("Expected redirect to Twitch OAuth")
	}

	parsed, err := url.Parse(location)
	if err != nil {
		t.Fatalf("Failed to parse redirect URL: %v", err)
	}

	query := parsed.Query()
	scope := query.Get("scope")
	if scope == "" {
		t.Error("Expected scope parameter in redirect URL")
	}
	if !strings.Contains(scope, "chat:read") || !strings.Contains(scope, "chat:edit") {
		t.Error("Expected chat scopes in redirect URL")
	}
	if !strings.Contains(scope, "moderator:manage:banned_users") {
		t.Error("Expected moderator:manage:banned_users scope in redirect URL")
	}
	if !strings.Contains(scope, "channel:manage:banned_users") {
		t.Error("Expected channel:manage:banned_users scope in redirect URL")
	}

	// Verify client_id and redirect_uri are present and non-empty
	if query.Get("client_id") != "test_client_id" {
		t.Error("Expected non-empty client_id parameter in redirect URL")
	}
	if query.Get("redirect_uri") == "" {
		t.Error("Expected non-empty redirect_uri parameter in redirect URL")
	}
}
