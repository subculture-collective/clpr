package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestGetLeaderboardInvalidType(t *testing.T) {
	// Set Gin to test mode
	gin.SetMode(gin.TestMode)

	// Create a minimal handler setup
	// We don't need real services since we're testing the invalid type path
	handler := &ReputationHandler{
		reputationService: nil, // Not accessed in invalid type case
		authService:       nil, // Not accessed in invalid type case
	}

	// Create router
	r := gin.New()
	r.GET("/leaderboards/:type", handler.GetLeaderboard)

	tests := []struct {
		name              string
		leaderboardType   string
		expectedStatus    int
		expectedErrorCode string
	}{
		{
			name:              "invalid leaderboard type",
			leaderboardType:   "invalid",
			expectedStatus:    http.StatusBadRequest,
			expectedErrorCode: "INVALID_LEADERBOARD_TYPE",
		},
		{
			name:              "empty leaderboard type",
			leaderboardType:   "",
			expectedStatus:    http.StatusNotFound, // Gin returns 404 for empty path params
			expectedErrorCode: "",
		},
		{
			name:              "numeric leaderboard type",
			leaderboardType:   "123",
			expectedStatus:    http.StatusBadRequest,
			expectedErrorCode: "INVALID_LEADERBOARD_TYPE",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create request
			path := "/leaderboards/" + tt.leaderboardType
			req := httptest.NewRequest("GET", path, http.NoBody)
			w := httptest.NewRecorder()

			// Serve request
			r.ServeHTTP(w, req)

			// Check status code
			if w.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			// For error responses, check JSON structure
			if tt.expectedStatus >= 400 && tt.expectedErrorCode != "" {
				// Verify response is valid JSON
				var response map[string]interface{}
				if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
					t.Errorf("response is not valid JSON: %v, body: %s", err, w.Body.String())
					return
				}

				// Check content type
				contentType := w.Header().Get("Content-Type")
				if contentType != "application/json; charset=utf-8" {
					t.Errorf("expected Content-Type 'application/json; charset=utf-8', got '%s'", contentType)
				}

				// Verify error response structure
				code, ok := response["code"].(string)
				if !ok {
					t.Error("code field missing or not a string in error response")
				} else if code != tt.expectedErrorCode {
					t.Errorf("expected error code '%s', got '%s'", tt.expectedErrorCode, code)
				}

				// Verify all required error fields are present
				if _, ok := response["error"]; !ok {
					t.Error("error field missing in error response")
				}
				if _, ok := response["message"]; !ok {
					t.Error("message field missing in error response")
				}
			}
		})
	}
}

func TestGetLeaderboardRejectsInvalidPaginationBeforeService(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := &ReputationHandler{}
	router := gin.New()
	router.GET("/leaderboards/:type", handler.GetLeaderboard)
	tests := []string{
		"/leaderboards/karma?limit=0",
		"/leaderboards/karma?limit=101",
		"/leaderboards/karma?limit=many",
		"/leaderboards/engagement?page=0",
		"/leaderboards/engagement?page=1000001",
		"/leaderboards/engagement?page=many",
	}
	for _, target := range tests {
		t.Run(target, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			router.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, target, http.NoBody))
			if recorder.Code != http.StatusBadRequest {
				t.Fatalf("expected 400, got %d: %s", recorder.Code, recorder.Body.String())
			}
			if !strings.Contains(recorder.Body.String(), `"code":"INVALID_PAGINATION"`) {
				t.Fatalf("missing pagination error code: %s", recorder.Body.String())
			}
		})
	}
}

func TestGetUserKarmaRejectsInvalidHistoryLimitBeforeService(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := &ReputationHandler{}
	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Params = gin.Params{{Key: "id", Value: "123e4567-e89b-12d3-a456-426614174000"}}
	ctx.Request = httptest.NewRequest(http.MethodGet, "/api/v1/users/id/karma?limit=101", http.NoBody)
	handler.GetUserKarma(ctx)
	if recorder.Code != http.StatusBadRequest || !strings.Contains(recorder.Body.String(), `"code":"INVALID_PAGINATION"`) {
		t.Fatalf("expected pagination 400 before service work, got %d: %s", recorder.Code, recorder.Body.String())
	}
}

func TestGetBadgeDefinitionsContract(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := &ReputationHandler{}
	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Request = httptest.NewRequest(http.MethodGet, "/api/v1/badges", http.NoBody)
	handler.GetBadgeDefinitions(ctx)
	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", recorder.Code, recorder.Body.String())
	}
	var response struct {
		Badges []struct {
			ID string `json:"id"`
		} `json:"badges"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil || len(response.Badges) == 0 {
		t.Fatalf("invalid badge response: %v %s", err, recorder.Body.String())
	}
	for i := 1; i < len(response.Badges); i++ {
		if response.Badges[i-1].ID > response.Badges[i].ID {
			t.Fatalf("badges are not sorted: %q before %q", response.Badges[i-1].ID, response.Badges[i].ID)
		}
	}
}

func TestGetBadgeDefinitionsRejectsUnsupportedAccept(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := &ReputationHandler{}
	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Request = httptest.NewRequest(http.MethodGet, "/api/v1/badges", http.NoBody)
	ctx.Request.Header.Set("Accept", "text/html")
	handler.GetBadgeDefinitions(ctx)
	if recorder.Code != http.StatusNotAcceptable {
		t.Fatalf("expected 406, got %d: %s", recorder.Code, recorder.Body.String())
	}
}

func TestGetLeaderboardJSONResponse(t *testing.T) {
	// Set Gin to test mode
	gin.SetMode(gin.TestMode)

	// This test verifies that even without a database,
	// the endpoint returns JSON (not HTML) when it fails
	handler := &ReputationHandler{
		reputationService: nil, // Will cause nil pointer if accessed
		authService:       nil,
	}

	r := gin.New()
	// Add recovery middleware to catch panics
	r.Use(func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
					"error":   "Internal server error",
					"code":    "INTERNAL_ERROR",
					"message": "An unexpected error occurred",
				})
			}
		}()
		c.Next()
	})
	r.GET("/leaderboards/:type", handler.GetLeaderboard)

	// Test valid type but with nil service (will panic)
	req := httptest.NewRequest("GET", "/leaderboards/karma", http.NoBody)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	// Should get 500 status
	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected status %d, got %d", http.StatusInternalServerError, w.Code)
	}

	// Verify response is JSON, not HTML
	contentType := w.Header().Get("Content-Type")
	if contentType != "application/json; charset=utf-8" {
		t.Errorf("expected JSON content type, got '%s'", contentType)
	}

	// Verify we can parse the response as JSON
	var response map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Errorf("response is not valid JSON: %v, body: %s", err, w.Body.String())
	}
}
