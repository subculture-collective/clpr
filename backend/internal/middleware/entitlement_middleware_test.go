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

// mockSubscriptionService is a minimal mock for testing
type mockSubscriptionService struct {
	isProUserFunc             func(ctx context.Context, userID uuid.UUID) bool
	hasActiveSubscriptionFunc func(ctx context.Context, userID uuid.UUID) bool
}

func (m *mockSubscriptionService) IsProUser(ctx context.Context, userID uuid.UUID) bool {
	if m.isProUserFunc != nil {
		return m.isProUserFunc(ctx, userID)
	}
	return false
}

func (m *mockSubscriptionService) HasActiveSubscription(ctx context.Context, userID uuid.UUID) bool {
	if m.hasActiveSubscriptionFunc != nil {
		return m.hasActiveSubscriptionFunc(ctx, userID)
	}
	return false
}

// mockAuditLogService is a minimal mock for testing
type mockAuditLogService struct {
	logEntitlementDenialFunc func(ctx context.Context, userID uuid.UUID, action string, metadata map[string]interface{}) error
	callCount                int
}

func (m *mockAuditLogService) LogEntitlementDenial(ctx context.Context, userID uuid.UUID, action string, metadata map[string]interface{}) error {
	m.callCount++
	if m.logEntitlementDenialFunc != nil {
		return m.logEntitlementDenialFunc(ctx, userID, action, metadata)
	}
	return nil
}

func TestRequireProSubscription_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Create test user
	userID := uuid.New()
	user := &models.User{
		ID:       userID,
		Username: "testuser",
		Role:     "user",
	}

	// Create mock services
	mockSubService := &mockSubscriptionService{
		isProUserFunc: func(ctx context.Context, uid uuid.UUID) bool {
			return uid == userID // User is Pro
		},
	}
	mockAuditService := &mockAuditLogService{}

	// Create test router
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("user", user)
		c.Next()
	})
	router.Use(RequireProSubscription(mockSubService, mockAuditService))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	// Make request
	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}
	if mockAuditService.callCount != 0 {
		t.Errorf("expected no audit log calls, got %d", mockAuditService.callCount)
	}
}

func TestRequireProSubscription_Denied(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Create test user
	userID := uuid.New()
	user := &models.User{
		ID:       userID,
		Username: "testuser",
		Role:     "user",
	}

	// Create mock services
	mockSubService := &mockSubscriptionService{
		isProUserFunc: func(ctx context.Context, uid uuid.UUID) bool {
			return false // User is not Pro
		},
	}
	mockAuditService := &mockAuditLogService{}

	// Create test router
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("user", user)
		c.Next()
	})
	router.Use(RequireProSubscription(mockSubService, mockAuditService))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	// Make request
	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	if w.Code != http.StatusForbidden {
		t.Errorf("expected status 403, got %d", w.Code)
	}
	if mockAuditService.callCount != 1 {
		t.Errorf("expected 1 audit log call, got %d", mockAuditService.callCount)
	}
}

func TestRequireProSubscription_NoUser(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Create mock services
	mockSubService := &mockSubscriptionService{}
	mockAuditService := &mockAuditLogService{}

	// Create test router
	router := gin.New()
	router.Use(RequireProSubscription(mockSubService, mockAuditService))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	// Make request without user in context
	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected status 401, got %d", w.Code)
	}
	if mockAuditService.callCount != 0 {
		t.Errorf("expected no audit log calls, got %d", mockAuditService.callCount)
	}
}

func TestRequireActiveSubscription_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Create test user
	userID := uuid.New()
	user := &models.User{
		ID:       userID,
		Username: "testuser",
		Role:     "user",
	}

	// Create mock services
	mockSubService := &mockSubscriptionService{
		hasActiveSubscriptionFunc: func(ctx context.Context, uid uuid.UUID) bool {
			return uid == userID // User has active subscription
		},
	}
	mockAuditService := &mockAuditLogService{}

	// Create test router
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("user", user)
		c.Next()
	})
	router.Use(RequireActiveSubscription(mockSubService, mockAuditService))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	// Make request
	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}
	if mockAuditService.callCount != 0 {
		t.Errorf("expected no audit log calls, got %d", mockAuditService.callCount)
	}
}

func TestRequireActiveSubscription_Denied(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Create test user
	userID := uuid.New()
	user := &models.User{
		ID:       userID,
		Username: "testuser",
		Role:     "user",
	}

	// Create mock services
	mockSubService := &mockSubscriptionService{
		hasActiveSubscriptionFunc: func(ctx context.Context, uid uuid.UUID) bool {
			return false // User doesn't have active subscription
		},
	}
	mockAuditService := &mockAuditLogService{}

	// Create test router
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("user", user)
		c.Next()
	})
	router.Use(RequireActiveSubscription(mockSubService, mockAuditService))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	// Make request
	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	if w.Code != http.StatusForbidden {
		t.Errorf("expected status 403, got %d", w.Code)
	}
	if mockAuditService.callCount != 1 {
		t.Errorf("expected 1 audit log call, got %d", mockAuditService.callCount)
	}
}

func TestRequireProSubscription_NilAuditService(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Create test user
	userID := uuid.New()
	user := &models.User{
		ID:       userID,
		Username: "testuser",
		Role:     "user",
	}

	// Create mock services
	mockSubService := &mockSubscriptionService{
		isProUserFunc: func(ctx context.Context, uid uuid.UUID) bool {
			return false // User is not Pro
		},
	}

	// Create test router with nil audit service
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("user", user)
		c.Next()
	})
	router.Use(RequireProSubscription(mockSubService, nil))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	// Make request
	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert - should still deny access
	if w.Code != http.StatusForbidden {
		t.Errorf("expected status 403, got %d", w.Code)
	}
}
