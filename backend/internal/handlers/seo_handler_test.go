package handlers

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"git.subcult.tv/subculture-collective/clpr/internal/models"
)

// mockClipRepository is a mock implementation of ClipRepository for testing
type mockClipRepository struct {
	clips []models.Clip
}

func (m *mockClipRepository) ListForSitemap(ctx context.Context) ([]models.Clip, error) {
	return m.clips, nil
}

func (m *mockClipRepository) ListForSitemapBroadcasters(ctx context.Context) ([]models.BroadcasterWithClipCount, error) {
	return nil, nil
}

func TestGetSitemap(t *testing.T) {
	// Set Gin to test mode
	gin.SetMode(gin.TestMode)

	// Create test clips
	clip1 := models.Clip{
		ID:        uuid.New(),
		CreatedAt: time.Now().Add(-24 * time.Hour),
	}
	clip2 := models.Clip{
		ID:        uuid.New(),
		CreatedAt: time.Now().Add(-48 * time.Hour),
	}

	mockRepo := &mockClipRepository{
		clips: []models.Clip{clip1, clip2},
	}

	handler := NewSEOHandler(mockRepo, nil)

	// Create router
	r := gin.New()
	// Middleware to set base_url
	r.Use(func(c *gin.Context) {
		c.Set("base_url", "https://test.clpr.app")
		c.Next()
	})
	r.GET("/sitemap.xml", handler.GetSitemap)

	// Create request
	req := httptest.NewRequest("GET", "/sitemap.xml", nil)
	w := httptest.NewRecorder()

	// Serve request
	r.ServeHTTP(w, req)

	// Check status code
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	// Check content type
	contentType := w.Header().Get("Content-Type")
	if contentType != "application/xml" {
		t.Errorf("Expected Content-Type application/xml, got %s", contentType)
	}

	// Check response contains expected elements
	body := w.Body.String()
	expectedStrings := []string{
		"<?xml version=\"1.0\" encoding=\"UTF-8\"?>",
		"<urlset xmlns=\"http://www.sitemaps.org/schemas/sitemap/0.9\">",
		"https://test.clpr.app/",
		"https://test.clpr.app/discover",
		clip1.ID.String(),
		clip2.ID.String(),
		"</urlset>",
	}

	for _, expected := range expectedStrings {
		if !strings.Contains(body, expected) {
			t.Errorf("Expected sitemap to contain '%s'", expected)
		}
	}
}

func TestGetRobotsTxtProduction(t *testing.T) {
	// Set Gin to test mode
	gin.SetMode(gin.TestMode)

	handler := NewSEOHandler(nil, nil) // No repo needed for robots.txt

	// Create router
	r := gin.New()
	// Middleware to set environment
	r.Use(func(c *gin.Context) {
		c.Set("environment", "production")
		c.Next()
	})
	r.GET("/robots.txt", handler.GetRobotsTxt)

	// Create request
	req := httptest.NewRequest("GET", "/robots.txt", nil)
	w := httptest.NewRecorder()

	// Serve request
	r.ServeHTTP(w, req)

	// Check status code
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	// Check content type
	contentType := w.Header().Get("Content-Type")
	if contentType != "text/plain" {
		t.Errorf("Expected Content-Type text/plain, got %s", contentType)
	}

	// Check response contains expected content for production
	body := w.Body.String()
	expectedStrings := []string{
		"User-agent: *",
		"Allow: /",
		"Sitemap:",
		"Disallow: /api/",
		"Disallow: /admin/",
	}

	for _, expected := range expectedStrings {
		if !strings.Contains(body, expected) {
			t.Errorf("Expected robots.txt to contain '%s'", expected)
		}
	}
}

func TestGetRobotsTxtDevelopment(t *testing.T) {
	// Set Gin to test mode
	gin.SetMode(gin.TestMode)

	handler := NewSEOHandler(nil, nil) // No repo needed for robots.txt

	// Create router
	r := gin.New()
	// Middleware to set environment
	r.Use(func(c *gin.Context) {
		c.Set("environment", "development")
		c.Next()
	})
	r.GET("/robots.txt", handler.GetRobotsTxt)

	// Create request
	req := httptest.NewRequest("GET", "/robots.txt", nil)
	w := httptest.NewRecorder()

	// Serve request
	r.ServeHTTP(w, req)

	// Check status code
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	// Check response disallows all in non-production
	body := w.Body.String()
	expectedStrings := []string{
		"User-agent: *",
		"Disallow: /",
	}

	for _, expected := range expectedStrings {
		if !strings.Contains(body, expected) {
			t.Errorf("Expected robots.txt to contain '%s'", expected)
		}
	}
}
