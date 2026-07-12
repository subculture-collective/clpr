package handlers

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDocsHandler_GetDocsList(t *testing.T) {
	// Create a temporary docs directory
	tmpDir := t.TempDir()

	// Create some test markdown files
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "test1.md"), []byte("# Test 1"), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "test2.md"), []byte("# Test 2"), 0644))

	// Create a subdirectory with files
	subDir := filepath.Join(tmpDir, "subdir")
	require.NoError(t, os.Mkdir(subDir, 0755))
	require.NoError(t, os.WriteFile(filepath.Join(subDir, "test3.md"), []byte("# Test 3"), 0644))

	// Setup
	gin.SetMode(gin.TestMode)
	router := gin.New()
	handler := NewDocsHandler(tmpDir, "test-owner", "test-repo", "main")
	router.GET("/api/v1/docs", handler.GetDocsList)

	// Test
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/docs", nil)
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "test1")
	assert.Contains(t, w.Body.String(), "test2")
	assert.Contains(t, w.Body.String(), "subdir")
}

func TestDocsHandler_GetNestedDoc(t *testing.T) {
	tmpDir := t.TempDir()
	require.NoError(t, os.Mkdir(filepath.Join(tmpDir, "guides"), 0755))
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "guides", "release.md"), []byte("# Release Guide"), 0644))
	router := gin.New()
	handler := NewDocsHandler(tmpDir, "owner", "repo", "main")
	router.GET("/api/v1/docs/content/*path", handler.GetDoc)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/v1/docs/content/guides/release", nil))
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "Release Guide")
}

func TestDocsHandler_BlocksSymlinkEscape(t *testing.T) {
	tmpDir := t.TempDir()
	outsideDir := t.TempDir()
	outside := filepath.Join(outsideDir, "secret.md")
	require.NoError(t, os.WriteFile(outside, []byte("secret"), 0644))
	require.NoError(t, os.Symlink(outside, filepath.Join(tmpDir, "leak.md")))
	router := gin.New()
	handler := NewDocsHandler(tmpDir, "owner", "repo", "main")
	router.GET("/api/v1/docs/:path", handler.GetDoc)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/v1/docs/leak", nil))
	assert.Equal(t, http.StatusForbidden, w.Code)
	assert.NotContains(t, w.Body.String(), "secret")
}

func TestDocsHandler_BoundsDocumentsAndSearch(t *testing.T) {
	tmpDir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "large.md"), []byte(strings.Repeat("x", maxDocumentBytes+1)), 0644))
	handler := NewDocsHandler(tmpDir, "owner", "repo", "main")
	router := gin.New()
	router.GET("/api/v1/docs/search", handler.SearchDocs)
	router.GET("/api/v1/docs/:path", handler.GetDoc)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/v1/docs/large", nil))
	assert.Equal(t, http.StatusRequestEntityTooLarge, w.Code)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/v1/docs/search?q=x", nil))
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestDocsHandler_SearchUnavailableSource(t *testing.T) {
	handler := NewDocsHandler(filepath.Join(t.TempDir(), "missing"), "owner", "repo", "main")
	router := gin.New()
	router.GET("/api/v1/docs/search", handler.SearchDocs)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/v1/docs/search?q=release", nil))
	assert.Equal(t, http.StatusServiceUnavailable, w.Code)
}

func TestDocsHandler_ListRejectsUnsupportedAccept(t *testing.T) {
	handler := NewDocsHandler(t.TempDir(), "owner", "repo", "main")
	router := gin.New()
	router.GET("/api/v1/docs", handler.GetDocsList)
	w := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/v1/docs", nil)
	request.Header.Set("Accept", "text/html")
	router.ServeHTTP(w, request)
	assert.Equal(t, http.StatusNotAcceptable, w.Code)
}

func TestDocsHandler_GetDoc(t *testing.T) {
	// Create a temporary docs directory
	tmpDir := t.TempDir()
	testContent := "# Test Document\n\nThis is a test document."
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "test.md"), []byte(testContent), 0644))

	// Setup
	gin.SetMode(gin.TestMode)
	router := gin.New()
	handler := NewDocsHandler(tmpDir, "test-owner", "test-repo", "main")
	router.GET("/api/v1/docs/:path", handler.GetDoc)

	// Test
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/docs/test", nil)
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "Test Document")
	assert.Contains(t, w.Body.String(), "github_url")
}

func TestDocsHandler_GetDoc_NotFound(t *testing.T) {
	// Create a temporary docs directory
	tmpDir := t.TempDir()

	// Setup
	gin.SetMode(gin.TestMode)
	router := gin.New()
	handler := NewDocsHandler(tmpDir, "test-owner", "test-repo", "main")
	router.GET("/api/v1/docs/:path", handler.GetDoc)

	// Test
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/docs/nonexistent", nil)
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestDocsHandler_SearchDocs(t *testing.T) {
	// Create a temporary docs directory
	tmpDir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "searchable.md"), []byte("# Searchable\n\nThis document contains the word banana."), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "other.md"), []byte("# Other\n\nThis document is about apples."), 0644))

	// Setup
	gin.SetMode(gin.TestMode)
	router := gin.New()
	handler := NewDocsHandler(tmpDir, "test-owner", "test-repo", "main")
	router.GET("/api/v1/docs/search", handler.SearchDocs)

	// Test
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/docs/search?q=banana", nil)
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "searchable")
	assert.NotContains(t, w.Body.String(), "other")
}
