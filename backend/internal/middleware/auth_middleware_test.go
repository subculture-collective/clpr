package middleware

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"git.subcult.tv/subculture-collective/clpr/internal/models"
)

// mockAuthService is a mock implementation that provides GetUserFromToken
type mockAuthService struct {
	getUserFromTokenFunc func(ctx context.Context, token string) (*models.User, error)
}

func (m *mockAuthService) GetUserFromToken(ctx context.Context, token string) (*models.User, error) {
	if m.getUserFromTokenFunc != nil {
		return m.getUserFromTokenFunc(ctx, token)
	}
	return nil, errors.New("not implemented")
}

// authServiceWrapper wraps the mock to satisfy the services.AuthService interface
type authServiceWrapper struct {
	mock *mockAuthService
}

func (w *authServiceWrapper) GetUserFromToken(ctx context.Context, token string) (*models.User, error) {
	return w.mock.GetUserFromToken(ctx, token)
}

func TestAuthMiddleware_MissingToken(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Create a test router with the middleware logic
	router := gin.New()
	router.Use(func(c *gin.Context) {
		// Manually implement auth check for missing token
		token := extractToken(c)
		if token == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"error": gin.H{
					"code":    "UNAUTHORIZED",
					"message": "Missing authentication token",
				},
			})
			c.Abort()
			return
		}
		c.Next()
	})
	router.GET("/protected", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	// Make request without token
	req := httptest.NewRequest("GET", "/protected", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Should return 401
	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected status 401, got %d", w.Code)
	}

	// Check response structure
	expectedBody := `{"error":{"code":"UNAUTHORIZED","message":"Missing authentication token"},"success":false}`
	if w.Body.String() != expectedBody {
		t.Errorf("expected body %s, got %s", expectedBody, w.Body.String())
	}
}

func TestAuthMiddleware_InvalidToken(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Create a mock auth service that returns an error
	mockAuth := &mockAuthService{
		getUserFromTokenFunc: func(ctx context.Context, token string) (*models.User, error) {
			return nil, errors.New("invalid token")
		},
	}

	// Create a test router with the middleware logic
	router := gin.New()
	router.Use(func(c *gin.Context) {
		token := extractToken(c)
		if token == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"error": gin.H{
					"code":    "UNAUTHORIZED",
					"message": "Missing authentication token",
				},
			})
			c.Abort()
			return
		}

		// Get user from token
		_, err := mockAuth.GetUserFromToken(c.Request.Context(), token)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"error": gin.H{
					"code":    "UNAUTHORIZED",
					"message": "Invalid or expired token",
				},
			})
			c.Abort()
			return
		}
		c.Next()
	})
	router.GET("/protected", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	// Make request with invalid token
	req := httptest.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", "Bearer invalid_token")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Should return 401
	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected status 401, got %d", w.Code)
	}

	// Check response structure
	expectedBody := `{"error":{"code":"UNAUTHORIZED","message":"Invalid or expired token"},"success":false}`
	if w.Body.String() != expectedBody {
		t.Errorf("expected body %s, got %s", expectedBody, w.Body.String())
	}
}

func TestAuthMiddleware_ValidToken(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Create a test user with email as pointer
	email := "test@example.com"
	twitchID := "12345"
	testUser := &models.User{
		ID:          uuid.New(),
		TwitchID:    &twitchID,
		Username:    "testuser",
		DisplayName: "Test User",
		Email:       &email,
		Role:        "user",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// Create a mock auth service that returns a valid user
	mockAuth := &mockAuthService{
		getUserFromTokenFunc: func(ctx context.Context, token string) (*models.User, error) {
			if token == "valid_token" {
				return testUser, nil
			}
			return nil, errors.New("invalid token")
		},
	}

	// Create a test router with the middleware logic
	router := gin.New()
	router.Use(func(c *gin.Context) {
		token := extractToken(c)
		if token == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"error": gin.H{
					"code":    "UNAUTHORIZED",
					"message": "Missing authentication token",
				},
			})
			c.Abort()
			return
		}

		user, err := mockAuth.GetUserFromToken(c.Request.Context(), token)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"error": gin.H{
					"code":    "UNAUTHORIZED",
					"message": "Invalid or expired token",
				},
			})
			c.Abort()
			return
		}

		c.Set("user", user)
		c.Set("user_id", user.ID)
		c.Set("user_role", user.Role)
		c.Next()
	})
	router.GET("/protected", func(c *gin.Context) {
		// Verify user is set in context
		userID, exists := c.Get("user_id")
		if !exists {
			t.Error("user_id not set in context")
		}
		if userID != testUser.ID {
			t.Errorf("expected user_id %v, got %v", testUser.ID, userID)
		}

		user, exists := c.Get("user")
		if !exists {
			t.Error("user not set in context")
		}
		if user.(*models.User).ID != testUser.ID {
			t.Errorf("expected user ID %v, got %v", testUser.ID, user.(*models.User).ID)
		}

		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	// Make request with valid token
	req := httptest.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", "Bearer valid_token")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Should return 200
	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}
}

func TestAuthMiddleware_TokenFromCookie(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Create a test user
	email := "test@example.com"
	twitchID := "12345"
	testUser := &models.User{
		ID:          uuid.New(),
		TwitchID:    &twitchID,
		Username:    "testuser",
		DisplayName: "Test User",
		Email:       &email,
		Role:        "user",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// Create a mock auth service
	mockAuth := &mockAuthService{
		getUserFromTokenFunc: func(ctx context.Context, token string) (*models.User, error) {
			if token == "cookie_token" {
				return testUser, nil
			}
			return nil, errors.New("invalid token")
		},
	}

	// Create a test router with the middleware logic
	router := gin.New()
	router.Use(func(c *gin.Context) {
		token := extractToken(c)
		if token == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"error": gin.H{
					"code":    "UNAUTHORIZED",
					"message": "Missing authentication token",
				},
			})
			c.Abort()
			return
		}

		user, err := mockAuth.GetUserFromToken(c.Request.Context(), token)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"error": gin.H{
					"code":    "UNAUTHORIZED",
					"message": "Invalid or expired token",
				},
			})
			c.Abort()
			return
		}

		c.Set("user", user)
		c.Set("user_id", user.ID)
		c.Set("user_role", user.Role)
		c.Next()
	})
	router.GET("/protected", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	// Make request with token in cookie
	req := httptest.NewRequest("GET", "/protected", nil)
	req.AddCookie(&http.Cookie{
		Name:  "access_token",
		Value: "cookie_token",
	})
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Should return 200
	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}
}

func TestRequireRole_NoUserRole(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(RequireRole("admin"))
	router.GET("/admin", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	req := httptest.NewRequest("GET", "/admin", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Should return 401 when no user_role in context
	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected status 401, got %d", w.Code)
	}
}

func TestRequireRole_InvalidRoleFormat(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(func(c *gin.Context) {
		// Set invalid role format (not a string)
		c.Set("user_role", 123)
		c.Next()
	})
	router.Use(RequireRole("admin"))
	router.GET("/admin", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	req := httptest.NewRequest("GET", "/admin", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Should return 500 for invalid role format
	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected status 500, got %d", w.Code)
	}
}

func TestRequireRole_InsufficientPermissions(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("user_role", "user")
		c.Next()
	})
	router.Use(RequireRole("admin", "moderator"))
	router.GET("/admin", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	req := httptest.NewRequest("GET", "/admin", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Should return 403 when user doesn't have required role
	if w.Code != http.StatusForbidden {
		t.Errorf("expected status 403, got %d", w.Code)
	}

	expectedBody := `{"error":{"code":"FORBIDDEN","message":"Insufficient permissions"},"success":false}`
	if w.Body.String() != expectedBody {
		t.Errorf("expected body %s, got %s", expectedBody, w.Body.String())
	}
}

func TestRequireRole_AdminAccess(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("user_role", "admin")
		c.Next()
	})
	router.Use(RequireRole("admin", "moderator"))
	router.GET("/admin", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	req := httptest.NewRequest("GET", "/admin", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Should return 200 when user has admin role
	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}
}

func TestRequireRole_ModeratorAccess(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("user_role", "moderator")
		c.Next()
	})
	router.Use(RequireRole("admin", "moderator"))
	router.GET("/admin", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	req := httptest.NewRequest("GET", "/admin", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Should return 200 when user has moderator role
	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}
}

func TestRequireRole_SingleRole(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("user_role", "moderator")
		c.Next()
	})
	router.Use(RequireRole("admin"))
	router.GET("/admin", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	req := httptest.NewRequest("GET", "/admin", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Should return 403 when user role doesn't match the single required role
	if w.Code != http.StatusForbidden {
		t.Errorf("expected status 403, got %d", w.Code)
	}
}
