package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"testing/fstest"

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
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "internal.md"), []byte("# Internal"), 0644))

	// Create a subdirectory with files
	subDir := filepath.Join(tmpDir, "subdir")
	require.NoError(t, os.Mkdir(subDir, 0755))
	require.NoError(t, os.WriteFile(filepath.Join(subDir, "test3.md"), []byte("# Test 3"), 0644))

	// Setup
	gin.SetMode(gin.TestMode)
	router := gin.New()
	handler := NewDocsHandler(tmpDir, []string{"test1.md", "test2.md", "subdir/test3.md"}, "test-owner", "test-repo", "main")
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
	assert.NotContains(t, w.Body.String(), "internal")
}

func TestDocsHandler_GetNestedDoc(t *testing.T) {
	tmpDir := t.TempDir()
	require.NoError(t, os.Mkdir(filepath.Join(tmpDir, "guides"), 0755))
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "guides", "release.md"), []byte("# Release Guide"), 0644))
	router := gin.New()
	handler := NewDocsHandler(tmpDir, []string{"guides/release.md"}, "owner", "repo", "main")
	router.GET("/api/v1/docs/content/*path", handler.GetDoc)
	router.GET("/api/v1/docs/:path", handler.GetDoc)
	for _, target := range []string{"/api/v1/docs/content/guides/release", "/api/v1/docs/content/guides/release.md"} {
		w := httptest.NewRecorder()
		router.ServeHTTP(w, httptest.NewRequest(http.MethodGet, target, nil))
		assert.Equal(t, http.StatusOK, w.Code, target)
		assert.Contains(t, w.Body.String(), "Release Guide", target)
	}
}

// Production routes set UseRawPath, so an encoded slash reaches the
// single-segment route as part of the document path.
func TestDocsHandler_GetNestedDocWithEncodedSlash(t *testing.T) {
	docs := fstest.MapFS{"compliance/twitch-embeds.md": {Data: []byte("# Twitch Embed Compliance")}}
	handler := NewDocsHandlerFS(docs, []string{"compliance/twitch-embeds.md"}, "", "", "")
	router := gin.New()
	router.UseRawPath = true
	router.UnescapePathValues = true
	router.GET("/api/v1/docs/:path", handler.GetDoc)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/v1/docs/compliance%2Ftwitch-embeds", nil))
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "Twitch Embed Compliance")
}

func TestDocsHandler_BlocksSymlinkEscape(t *testing.T) {
	tmpDir := t.TempDir()
	outsideDir := t.TempDir()
	outside := filepath.Join(outsideDir, "secret.md")
	require.NoError(t, os.WriteFile(outside, []byte("secret"), 0644))
	require.NoError(t, os.Symlink(outside, filepath.Join(tmpDir, "leak.md")))
	router := gin.New()
	handler := NewDocsHandler(tmpDir, []string{"leak.md"}, "owner", "repo", "main")
	router.GET("/api/v1/docs", handler.GetDocsList)
	router.GET("/api/v1/docs/:path", handler.GetDoc)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/v1/docs/leak", nil))
	assert.Equal(t, http.StatusForbidden, w.Code)
	assert.NotContains(t, w.Body.String(), "secret")

	w = httptest.NewRecorder()
	router.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/v1/docs", nil))
	assert.JSONEq(t, `{"docs":[]}`, w.Body.String())
}

// Documents that exist in the source but are not allowlisted are neither
// listed, searchable, nor readable.
func TestDocsHandler_FailsClosedForUnlistedDocuments(t *testing.T) {
	docs := fstest.MapFS{
		"compliance/twitch-embeds.md":                      {Data: []byte("# Twitch Embed Compliance\n\nembed parent domain")},
		"security/jwt-secret-history-review-2026-08-09.md": {Data: []byte("# JWT review\n\nembed secret rotation")},
		"operations/runbooks/background-jobs.md":           {Data: []byte("# Background Jobs\n\nembed worker")},
		"NSFW_ENV_VARS.md":                                 {Data: []byte("# NSFW env\n\nembed")},
	}
	handler := NewDocsHandlerFS(docs, []string{"compliance/twitch-embeds.md"}, "owner", "repo", "main")
	router := gin.New()
	router.UseRawPath = true
	router.UnescapePathValues = true
	router.GET("/api/v1/docs", handler.GetDocsList)
	router.GET("/api/v1/docs/search", handler.SearchDocs)
	router.GET("/api/v1/docs/content/*path", handler.GetDoc)
	router.GET("/api/v1/docs/:path", handler.GetDoc)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/v1/docs", nil))
	require.Equal(t, http.StatusOK, w.Code)
	assert.JSONEq(t, `{"docs":[
		{"name":"compliance","path":"compliance","type":"directory","children":[
			{"name":"twitch-embeds","path":"compliance/twitch-embeds","type":"file","title":"Twitch Embed Compliance"}
		]}
	]}`, w.Body.String())

	for _, target := range []string{
		"/api/v1/docs/content/security/jwt-secret-history-review-2026-08-09",
		"/api/v1/docs/security%2Fjwt-secret-history-review-2026-08-09",
		"/api/v1/docs/content/operations/runbooks/background-jobs.md",
		"/api/v1/docs/NSFW_ENV_VARS",
	} {
		w = httptest.NewRecorder()
		router.ServeHTTP(w, httptest.NewRequest(http.MethodGet, target, nil))
		assert.Equal(t, http.StatusNotFound, w.Code, target)
		assert.NotContains(t, w.Body.String(), "#", target)
	}

	w = httptest.NewRecorder()
	router.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/v1/docs/search?q=embed", nil))
	require.Equal(t, http.StatusOK, w.Code)
	var search struct {
		Results []SearchResult `json:"results"`
		Count   int            `json:"count"`
	}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &search))
	require.Equal(t, 1, search.Count)
	assert.Equal(t, "compliance/twitch-embeds", search.Results[0].Path)
	assert.Equal(t, "Twitch Embed Compliance", search.Results[0].Title)
}

func TestDocsHandler_EmptyAllowlistPublishesNothing(t *testing.T) {
	docs := fstest.MapFS{"index.md": {Data: []byte("# Home")}}
	handler := NewDocsHandlerFS(docs, nil, "owner", "repo", "main")
	router := gin.New()
	router.GET("/api/v1/docs", handler.GetDocsList)
	router.GET("/api/v1/docs/:path", handler.GetDoc)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/v1/docs", nil))
	assert.JSONEq(t, `{"docs":[]}`, w.Body.String())
	w = httptest.NewRecorder()
	router.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/v1/docs/index", nil))
	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestDocsHandler_BoundsDocumentsAndSearch(t *testing.T) {
	tmpDir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "large.md"), []byte(strings.Repeat("x", maxDocumentBytes+1)), 0644))
	handler := NewDocsHandler(tmpDir, []string{"large.md"}, "owner", "repo", "main")
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
	handler := NewDocsHandler(filepath.Join(t.TempDir(), "missing"), []string{"index.md"}, "owner", "repo", "main")
	router := gin.New()
	router.GET("/api/v1/docs/search", handler.SearchDocs)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/v1/docs/search?q=release", nil))
	assert.Equal(t, http.StatusServiceUnavailable, w.Code)
}

func TestDocsHandler_ListRejectsUnsupportedAccept(t *testing.T) {
	handler := NewDocsHandler(t.TempDir(), nil, "owner", "repo", "main")
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
	handler := NewDocsHandler(tmpDir, []string{"test.md"}, "test-owner", "test-repo", "main")
	router.GET("/api/v1/docs/:path", handler.GetDoc)

	// Test
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/docs/test", nil)
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "Test Document")
	assert.Contains(t, w.Body.String(), `"title":"Test Document"`)
	assert.Contains(t, w.Body.String(), "github_url")
}

func TestDocsHandler_GetDoc_NotFound(t *testing.T) {
	// Create a temporary docs directory
	tmpDir := t.TempDir()

	// Setup
	gin.SetMode(gin.TestMode)
	router := gin.New()
	handler := NewDocsHandler(tmpDir, []string{"nonexistent.md"}, "test-owner", "test-repo", "main")
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
	handler := NewDocsHandler(tmpDir, []string{"other.md", "searchable.md"}, "test-owner", "test-repo", "main")
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

// Production returned 500 for GET /api/v1/docs because the image had no docs
// directory at the configured DOCS_PATH (../docs relative to /app).
func TestDocsHandler_ListMissingSourceIsEmpty(t *testing.T) {
	handler := NewDocsHandler(filepath.Join(t.TempDir(), "missing"), []string{"index.md"}, "owner", "repo", "main")
	router := gin.New()
	router.GET("/api/v1/docs", handler.GetDocsList)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/v1/docs", nil))
	assert.Equal(t, http.StatusOK, w.Code)
	assert.JSONEq(t, `{"docs":[]}`, w.Body.String())
}

func TestDocsHandler_ServesFileSystem(t *testing.T) {
	docs := fstest.MapFS{
		"index.md":          {Data: []byte("# Home\n\nWelcome to clpr.")},
		"users/guide.md":    {Data: []byte("# Guide\n\nHow to tag clips.")},
		"archive/old.md":    {Data: []byte("# Old")},
		"users/diagram.png": {Data: []byte("png")},
	}
	handler := NewDocsHandlerFS(docs, []string{"index.md", "users/guide.md"}, "owner", "repo", "main")
	router := gin.New()
	router.GET("/api/v1/docs", handler.GetDocsList)
	router.GET("/api/v1/docs/search", handler.SearchDocs)
	router.GET("/api/v1/docs/content/*path", handler.GetDoc)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/v1/docs", nil))
	require.Equal(t, http.StatusOK, w.Code)
	assert.JSONEq(t, `{"docs":[
		{"name":"index","path":"index","type":"file","title":"Home"},
		{"name":"users","path":"users","type":"directory","children":[{"name":"guide","path":"users/guide","type":"file","title":"Guide"}]}
	]}`, w.Body.String())

	w = httptest.NewRecorder()
	router.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/v1/docs/content/users/guide", nil))
	require.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "How to tag clips.")
	assert.Contains(t, w.Body.String(), "https://github.com/owner/repo/edit/main/docs/users/guide.md")

	w = httptest.NewRecorder()
	router.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/v1/docs/search?q=tag+clips", nil))
	require.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"path":"users/guide"`)
}

// Each entry is titled from its own content. Base names repeat across
// directories (README, ARCHITECTURE), and a file's name need not describe it.
func TestDocsHandler_TitlesComeFromEachDocument(t *testing.T) {
	docs := fstest.MapFS{
		"README.md":            {Data: []byte("# Embedding Backfill Tool\n")},
		"ARCHITECTURE.md":      {Data: []byte("# Architecture Diagram: Saved Searches Feature\n")},
		"adr/README.md":        {Data: []byte("---\ntitle: \"README\"\n---\n\n# Architecture Decision Records (ADR)\n")},
		"compliance/README.md": {Data: []byte("---\ntitle: \"README\"\n---\n\n# Twitch Compliance Documentation\n")},
	}
	handler := NewDocsHandlerFS(docs, []string{"ARCHITECTURE.md", "README.md", "adr/README.md", "compliance/README.md"}, "", "", "")
	router := gin.New()
	router.GET("/api/v1/docs", handler.GetDocsList)
	router.GET("/api/v1/docs/content/*path", handler.GetDoc)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/v1/docs", nil))
	require.Equal(t, http.StatusOK, w.Code)
	assert.JSONEq(t, `{"docs":[
		{"name":"ARCHITECTURE","path":"ARCHITECTURE","type":"file","title":"Architecture Diagram: Saved Searches Feature"},
		{"name":"README","path":"README","type":"file","title":"Embedding Backfill Tool"},
		{"name":"adr","path":"adr","type":"directory","children":[
			{"name":"README","path":"adr/README","type":"file","title":"Architecture Decision Records (ADR)"}
		]},
		{"name":"compliance","path":"compliance","type":"directory","children":[
			{"name":"README","path":"compliance/README","type":"file","title":"Twitch Compliance Documentation"}
		]}
	]}`, w.Body.String())

	for target, title := range map[string]string{
		"/api/v1/docs/content/adr/README":        "Architecture Decision Records (ADR)",
		"/api/v1/docs/content/compliance/README": "Twitch Compliance Documentation",
		"/api/v1/docs/content/README":            "Embedding Backfill Tool",
	} {
		w = httptest.NewRecorder()
		router.ServeHTTP(w, httptest.NewRequest(http.MethodGet, target, nil))
		require.Equal(t, http.StatusOK, w.Code, target)
		var body struct {
			Title string `json:"title"`
		}
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
		assert.Equal(t, title, body.Title, target)
	}
}

func TestDocumentTitle(t *testing.T) {
	cases := []struct {
		name    string
		content string
		path    string
		want    string
	}{
		{"first heading", "# Twitch Embeds\n\n## Purpose\n# Later", "compliance/twitch-embeds.md", "Twitch Embeds"},
		{"heading after front matter", "---\ntitle: \"Short\"\nsummary: \"# not a heading\"\n---\n\n# Full Title\n", "a.md", "Full Title"},
		{"front matter when no heading", "---\ntitle: 'Feature Playlists'\n---\n\nBody only.\n", "features/feature-playlists.md", "Feature Playlists"},
		{"ignores fenced headings", "```bash\n# install deps\n```\n\n# Real Title\n", "a.md", "Real Title"},
		{"ignores tilde fences", "~~~\n# comment\n~~~\n# Real Title\n", "a.md", "Real Title"},
		{"closing hashes", "# Title ##\n", "a.md", "Title"},
		{"indented code is not a heading", "    # code\n", "docs/getting-started.md", "getting started"},
		{"level two is not a title", "## Section\n", "NSFW_ENV_VARS.md", "NSFW ENV VARS"},
		{"windows line endings", "---\r\ntitle: X\r\n---\r\n# Windows Title\r\n", "a.md", "Windows Title"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, DocumentTitle([]byte(tc.content), tc.path))
		})
	}
}
