package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"git.subcult.tv/subculture-collective/clpr/internal/models"
)

// mockMFAService is a mock implementation of MFAServiceInterface for testing
type mockMFAService struct {
	required      bool
	enabled       bool
	inGracePeriod bool
	shouldError   bool
	errorMessage  error
}

func (m *mockMFAService) IsAdminActionAllowed(ctx context.Context, userID uuid.UUID) (bool, string, error) {
	if m.shouldError {
		return false, "", m.errorMessage
	}

	if !m.required {
		return true, "", nil
	}

	if m.enabled {
		return true, "", nil
	}

	if m.inGracePeriod {
		return true, "MFA setup required: Please enable MFA soon. Your grace period will expire.", nil
	}

	return false, "MFA is required for admin actions. Please enable MFA to continue.", nil
}

func (m *mockMFAService) CheckMFARequired(ctx context.Context, userID uuid.UUID) (required bool, enabled bool, inGracePeriod bool, err error) {
	if m.shouldError {
		return false, false, false, m.errorMessage
	}
	return m.required, m.enabled, m.inGracePeriod, nil
}

func (m *mockMFAService) SetMFARequired(ctx context.Context, userID uuid.UUID) error {
	if m.shouldError {
		return m.errorMessage
	}
	m.required = true
	return nil
}

func TestRequireMFAForAdminMiddleware_NoUser(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	mfaService := &mockMFAService{}
	router.Use(RequireMFAForAdminMiddleware(mfaService))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	req, _ := http.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status 401, got %d", w.Code)
	}
}

func TestRequireMFAForAdminMiddleware_RegularUser(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// Set up a regular user (not admin/moderator)
	router.Use(func(c *gin.Context) {
		user := &models.User{
			ID:          uuid.New(),
			Username:    "regularuser",
			Role:        models.RoleUser,
			AccountType: models.AccountTypeMember,
		}
		c.Set("user", user)
		c.Next()
	})

	mfaService := &mockMFAService{}
	router.Use(RequireMFAForAdminMiddleware(mfaService))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	req, _ := http.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Regular users should pass through without MFA check
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

func TestRequireMFAForAdminMiddleware_AdminWithMFA(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// Set up an admin user
	router.Use(func(c *gin.Context) {
		user := &models.User{
			ID:          uuid.New(),
			Username:    "admin",
			Role:        models.RoleAdmin,
			AccountType: models.AccountTypeAdmin,
		}
		c.Set("user", user)
		c.Next()
	})

	// Mock MFA service - MFA is required and enabled
	mfaService := &mockMFAService{
		required: true,
		enabled:  true,
	}
	router.Use(RequireMFAForAdminMiddleware(mfaService))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	req, _ := http.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Admin with MFA enabled should be allowed
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

func TestRequireMFAForAdminMiddleware_AdminInGracePeriod(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// Set up an admin user
	router.Use(func(c *gin.Context) {
		user := &models.User{
			ID:          uuid.New(),
			Username:    "admin",
			Role:        models.RoleAdmin,
			AccountType: models.AccountTypeAdmin,
		}
		c.Set("user", user)
		c.Next()
	})

	// Mock MFA service - MFA is required but not enabled, still in grace period
	mfaService := &mockMFAService{
		required:      true,
		enabled:       false,
		inGracePeriod: true,
	}
	router.Use(RequireMFAForAdminMiddleware(mfaService))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	req, _ := http.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Admin in grace period should be allowed with a warning
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	// Check for warning header
	warningHeader := w.Header().Get("X-MFA-Warning")
	if warningHeader == "" {
		t.Error("Expected X-MFA-Warning header to be set")
	}
}

func TestRequireMFAForAdminMiddleware_AdminWithoutMFAExpiredGracePeriod(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// Set up an admin user
	router.Use(func(c *gin.Context) {
		user := &models.User{
			ID:          uuid.New(),
			Username:    "admin",
			Role:        models.RoleAdmin,
			AccountType: models.AccountTypeAdmin,
		}
		c.Set("user", user)
		c.Next()
	})

	// Mock MFA service - MFA is required but not enabled, grace period expired
	mfaService := &mockMFAService{
		required:      true,
		enabled:       false,
		inGracePeriod: false,
	}
	router.Use(RequireMFAForAdminMiddleware(mfaService))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	req, _ := http.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Admin without MFA after grace period should be blocked
	if w.Code != http.StatusForbidden {
		t.Errorf("Expected status 403, got %d", w.Code)
	}
}

func TestRequireMFAForAdminMiddleware_ModeratorWithMFA(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// Set up a moderator user
	router.Use(func(c *gin.Context) {
		user := &models.User{
			ID:          uuid.New(),
			Username:    "moderator",
			Role:        models.RoleModerator,
			AccountType: models.AccountTypeModerator,
		}
		c.Set("user", user)
		c.Next()
	})

	// Mock MFA service - MFA is required and enabled
	mfaService := &mockMFAService{
		required: true,
		enabled:  true,
	}
	router.Use(RequireMFAForAdminMiddleware(mfaService))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	req, _ := http.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Moderator with MFA enabled should be allowed
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

func TestRequireMFAForAdminMiddleware_ModeratorWithoutMFAExpiredGracePeriod(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// Set up a moderator user
	router.Use(func(c *gin.Context) {
		user := &models.User{
			ID:          uuid.New(),
			Username:    "moderator",
			Role:        models.RoleModerator,
			AccountType: models.AccountTypeModerator,
		}
		c.Set("user", user)
		c.Next()
	})

	// Mock MFA service - MFA is required but not enabled, grace period expired
	mfaService := &mockMFAService{
		required:      true,
		enabled:       false,
		inGracePeriod: false,
	}
	router.Use(RequireMFAForAdminMiddleware(mfaService))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	req, _ := http.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Moderator without MFA after grace period should be blocked
	if w.Code != http.StatusForbidden {
		t.Errorf("Expected status 403, got %d", w.Code)
	}
}

func TestCheckMFARequirementMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	mfaService := &mockMFAService{
		required: false,
		enabled:  false,
	}

	// Handler that updates user to admin
	router.Use(func(c *gin.Context) {
		user := &models.User{
			ID:          uuid.New(),
			Username:    "newadmin",
			Role:        models.RoleAdmin,
			AccountType: models.AccountTypeAdmin,
		}
		c.Set("updated_user", user)
		c.Next()
	})

	router.Use(CheckMFARequirementMiddleware(mfaService))
	router.POST("/promote", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "promoted"})
	})

	req, _ := http.NewRequest("POST", "/promote", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	// Verify MFA requirement was set
	if !mfaService.required {
		t.Error("Expected MFA to be required after admin promotion")
	}
}
