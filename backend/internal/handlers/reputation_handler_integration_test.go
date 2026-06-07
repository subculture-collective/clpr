//go:build integration

package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"git.subcult.tv/subculture-collective/clpr/config"
	"git.subcult.tv/subculture-collective/clpr/internal/middleware"
	"git.subcult.tv/subculture-collective/clpr/internal/repository"
	"git.subcult.tv/subculture-collective/clpr/internal/services"
	"git.subcult.tv/subculture-collective/clpr/pkg/database"
)

func TestLeaderboardIntegration(t *testing.T) {
	// Skip if not running integration tests
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Set Gin to test mode
	gin.SetMode(gin.TestMode)

	// Setup database connection
	dbConfig := &config.DatabaseConfig{
		Host:     getEnvOrDefault("DB_HOST", "localhost"),
		Port:     getEnvOrDefault("DB_PORT", "5437"), // Test DB port (see docker-compose.test.yml)
		User:     getEnvOrDefault("DB_USER", "clpr"),
		Password: getEnvOrDefault("DB_PASSWORD", "clpr_password"),
		Name:     getEnvOrDefault("DB_NAME", "clpr_test"),
		SSLMode:  "disable",
	}

	db, err := database.NewDB(dbConfig)
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}
	defer db.Close()

	// Check if database is ready
	ctx := context.Background()
	if err := db.HealthCheck(ctx); err != nil {
		t.Fatalf("Test database is not ready: %v", err)
	}

	// Initialize repositories
	reputationRepo := repository.NewReputationRepository(db.Pool)
	userRepo := repository.NewUserRepository(db.Pool)

	// Initialize services
	reputationService := services.NewReputationService(reputationRepo, userRepo)

	// Initialize handler
	handler := NewReputationHandler(reputationService, nil)

	// Create router with recovery middleware
	r := gin.New()
	r.Use(middleware.JSONRecoveryMiddleware())
	r.GET("/leaderboards/:type", handler.GetLeaderboard)

	tests := []struct {
		name            string
		leaderboardType string
		expectedStatus  int
		checkJSON       bool
	}{
		{
			name:            "karma leaderboard returns valid JSON",
			leaderboardType: "karma",
			expectedStatus:  http.StatusOK,
			checkJSON:       true,
		},
		{
			name:            "engagement leaderboard returns valid JSON",
			leaderboardType: "engagement",
			expectedStatus:  http.StatusOK,
			checkJSON:       true,
		},
		{
			name:            "invalid type returns JSON error",
			leaderboardType: "invalid",
			expectedStatus:  http.StatusBadRequest,
			checkJSON:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create request
			req := httptest.NewRequest("GET", "/leaderboards/"+tt.leaderboardType, nil)
			w := httptest.NewRecorder()

			// Serve request
			r.ServeHTTP(w, req)

			// Check status code
			if w.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			// Verify response is JSON
			if tt.checkJSON {
				contentType := w.Header().Get("Content-Type")
				if contentType != "application/json; charset=utf-8" {
					t.Errorf("expected JSON content type, got '%s'", contentType)
				}

				// Verify we can parse as JSON
				var response map[string]interface{}
				if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
					t.Errorf("response is not valid JSON: %v, body: %s", err, w.Body.String())
				}

				// For success cases, check structure
				if tt.expectedStatus == http.StatusOK {
					if _, ok := response["type"]; !ok {
						t.Error("type field missing in success response")
					}
					if _, ok := response["entries"]; !ok {
						t.Error("entries field missing in success response")
					}
				}

				// For error cases, check error structure
				if tt.expectedStatus >= 400 {
					if _, ok := response["error"]; !ok {
						t.Error("error field missing in error response")
					}
					if _, ok := response["code"]; !ok {
						t.Error("code field missing in error response")
					}
					if _, ok := response["message"]; !ok {
						t.Error("message field missing in error response")
					}
				}
			}
		})
	}
}
