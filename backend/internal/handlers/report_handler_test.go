//go:build integration

package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"git.subcult.tv/subculture-collective/clpr/config"
	"git.subcult.tv/subculture-collective/clpr/internal/middleware"
	"git.subcult.tv/subculture-collective/clpr/internal/models"
	"git.subcult.tv/subculture-collective/clpr/internal/repository"
	"git.subcult.tv/subculture-collective/clpr/internal/services"
	"git.subcult.tv/subculture-collective/clpr/pkg/database"
)

func TestAdminReportsEndpoints(t *testing.T) {
	// Skip if not running integration tests
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Set Gin to test mode
	gin.SetMode(gin.TestMode)

	// Setup database connection
	dbConfig := &config.DatabaseConfig{
		Host:     getEnvOrDefault("DB_HOST", "localhost"),
		Port:     getEnvOrDefault("DB_PORT", "5437"),
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
	reportRepo := repository.NewReportRepository(db.Pool)
	clipRepo := repository.NewClipRepository(db.Pool)
	commentRepo := repository.NewCommentRepository(db.Pool)
	userRepo := repository.NewUserRepository(db.Pool)

	// Create a minimal auth service (not used in this test, but required for handler)
	authService := &services.AuthService{}

	// Initialize handler
	reportHandler := NewReportHandler(reportRepo, clipRepo, commentRepo, userRepo, authService)

	// Create test admin user
	adminEmail := "admin@test.com"
	adminTwitchID := "test_admin_" + uuid.New().String()
	adminUser := &models.User{
		ID:          uuid.New(),
		TwitchID:    &adminTwitchID,
		Username:    "testadmin",
		DisplayName: "Test Admin",
		Email:       &adminEmail,
		Role:        "admin",
		CreatedAt:   time.Now(),
	}
	if err := userRepo.Create(ctx, adminUser); err != nil {
		t.Fatalf("Failed to create test admin user: %v", err)
	}
	// Note: User cleanup is handled by database CASCADE on delete

	// Create test regular user (for reporter)
	userEmail := "user@test.com"
	userTwitchID := "test_user_" + uuid.New().String()
	regularUser := &models.User{
		ID:          uuid.New(),
		TwitchID:    &userTwitchID,
		Username:    "testuser",
		DisplayName: "Test User",
		Email:       &userEmail,
		Role:        "user",
		CreatedAt:   time.Now(),
	}
	if err := userRepo.Create(ctx, regularUser); err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}
	// Note: User cleanup is handled by database CASCADE on delete

	// Create test report
	testReport := &models.Report{
		ID:             uuid.New(),
		ReporterID:     regularUser.ID,
		ReportableType: "user",
		ReportableID:   regularUser.ID,
		Reason:         "spam",
		Description:    strPtr("Test report description"),
		Status:         "pending",
		CreatedAt:      time.Now(),
	}
	if err := reportRepo.CreateReport(ctx, testReport); err != nil {
		t.Fatalf("Failed to create test report: %v", err)
	}

	// Setup router
	r := gin.New()
	r.Use(middleware.JSONRecoveryMiddleware())

	// Mock authentication middleware that sets user context
	mockAuthMiddleware := func(c *gin.Context) {
		c.Set("user", adminUser)
		c.Set("user_id", adminUser.ID)
		c.Set("user_role", adminUser.Role)
		c.Next()
	}

	mockRoleMiddleware := func(c *gin.Context) {
		userRole, exists := c.Get("user_role")
		if !exists || userRole != "admin" {
			c.JSON(http.StatusForbidden, gin.H{"error": "Insufficient permissions"})
			c.Abort()
			return
		}
		c.Next()
	}

	// Setup routes with middleware
	admin := r.Group("/api/v1/admin")
	admin.Use(mockAuthMiddleware)
	admin.Use(mockRoleMiddleware)
	{
		adminReports := admin.Group("/reports")
		{
			adminReports.GET("", reportHandler.ListReports)
			adminReports.GET("/:id", reportHandler.GetReport)
			adminReports.PUT("/:id", reportHandler.UpdateReport)
		}
	}

	t.Run("ListReports returns 200 and reports list", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/v1/admin/reports", nil)
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d. Body: %s", w.Code, w.Body.String())
		}

		var response map[string]interface{}
		if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
			t.Fatalf("Failed to parse response: %v", err)
		}

		// Check response structure
		if _, ok := response["data"]; !ok {
			t.Error("Response missing 'data' field")
		}
		if _, ok := response["meta"]; !ok {
			t.Error("Response missing 'meta' field")
		}
	})

	t.Run("ListReports with status filter", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/v1/admin/reports?status=pending", nil)
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d. Body: %s", w.Code, w.Body.String())
		}

		var response map[string]interface{}
		if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
			t.Fatalf("Failed to parse response: %v", err)
		}

		data, ok := response["data"].([]interface{})
		if !ok {
			t.Error("Response data is not an array")
		} else if len(data) == 0 {
			t.Log("No pending reports found (expected at least one)")
		}
	})

	t.Run("GetReport returns 200 for valid report ID", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/v1/admin/reports/"+testReport.ID.String(), nil)
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d. Body: %s", w.Code, w.Body.String())
		}

		var response map[string]interface{}
		if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
			t.Fatalf("Failed to parse response: %v", err)
		}

		if _, ok := response["report"]; !ok {
			t.Error("Response missing 'report' field")
		}
	})

	t.Run("GetReport returns 400 for invalid UUID", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/v1/admin/reports/invalid-uuid", nil)
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status 400, got %d", w.Code)
		}
	})

	t.Run("UpdateReport can dismiss a report", func(t *testing.T) {
		updateData := map[string]interface{}{
			"status": "dismissed",
			"action": "mark_false",
		}
		jsonData, _ := json.Marshal(updateData)

		req := httptest.NewRequest("PUT", "/api/v1/admin/reports/"+testReport.ID.String(), bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d. Body: %s", w.Code, w.Body.String())
		}

		// Verify report was updated
		updatedReport, err := reportRepo.GetReportByID(ctx, testReport.ID)
		if err != nil {
			t.Fatalf("Failed to get updated report: %v", err)
		}
		if updatedReport.Status != "dismissed" {
			t.Errorf("Expected status 'dismissed', got '%s'", updatedReport.Status)
		}
	})

	t.Run("ListReports with pagination", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/v1/admin/reports?page=1&limit=10", nil)
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}

		var response map[string]interface{}
		if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
			t.Fatalf("Failed to parse response: %v", err)
		}

		meta, ok := response["meta"].(map[string]interface{})
		if !ok {
			t.Fatal("Response meta is not an object")
		}

		if page, ok := meta["page"].(float64); !ok || page != 1 {
			t.Errorf("Expected page 1, got %v", meta["page"])
		}
		if limit, ok := meta["limit"].(float64); !ok || limit != 10 {
			t.Errorf("Expected limit 10, got %v", meta["limit"])
		}
	})
}
