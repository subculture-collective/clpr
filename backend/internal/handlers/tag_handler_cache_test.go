package handlers

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"git.subcult.tv/subculture-collective/clpr/internal/models"
	"git.subcult.tv/subculture-collective/clpr/internal/repository"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeTagCache struct {
	values  map[string]string
	ttls    map[string]time.Duration
	deleted []string
}

func newFakeTagCache() *fakeTagCache {
	return &fakeTagCache{values: map[string]string{}, ttls: map[string]time.Duration{}}
}

func (f *fakeTagCache) Get(_ context.Context, key string) (string, error) {
	value, ok := f.values[key]
	if !ok {
		return "", errors.New("redis: nil")
	}
	return value, nil
}

func (f *fakeTagCache) Set(_ context.Context, key string, value interface{}, ttl time.Duration) error {
	f.values[key] = string(value.([]byte))
	f.ttls[key] = ttl
	return nil
}

func (f *fakeTagCache) DeletePattern(_ context.Context, pattern string) error {
	f.deleted = append(f.deleted, pattern)
	return nil
}

func TestListTagsServesCachedResponse(t *testing.T) {
	gin.SetMode(gin.TestMode)
	cache := newFakeTagCache()
	// Unknown sort and lane values normalize to the default key.
	cache.values["tags:api:list:popularity::20:1"] = `{"tags":[],"total":7}`
	handler := NewTagHandler(nil, nil, nil) // a cache miss would dereference the nil repository
	handler.SetResponseCache(cache)
	router := gin.New()
	router.GET("/api/v1/tags", handler.ListTags)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/v1/tags?sort=bogus&lane=bogus&limit=20", nil))

	require.Equal(t, http.StatusOK, w.Code)
	assert.JSONEq(t, `{"tags":[],"total":7}`, w.Body.String())
	assert.Equal(t, "application/json; charset=utf-8", w.Header().Get("Content-Type"))
}

func TestTagWritesInvalidateCachedResponsesOnlyOnSuccess(t *testing.T) {
	gin.SetMode(gin.TestMode)
	cache := newFakeTagCache()
	handler := NewTagHandler(nil, nil, nil)
	handler.SetResponseCache(cache)
	router := gin.New()
	router.DELETE("/admin/tags/:id", handler.DeleteTag)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, httptest.NewRequest(http.MethodDelete, "/admin/tags/not-a-uuid", nil))
	require.Equal(t, http.StatusBadRequest, w.Code)
	assert.Empty(t, cache.deleted, "a rejected write must not clear the cache")

	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest(http.MethodDelete, "/admin/tags/x", nil)
	c.Status(http.StatusOK)
	handler.invalidateTagResponses(c)
	assert.Equal(t, []string{"tags:api:*"}, cache.deleted)
}

func TestBuildTagTreeNestsRowsAndKeepsChildCounts(t *testing.T) {
	content := "content"
	rows := []repository.TagTreeRow{
		{Tag: models.Tag{ID: uuid.New(), Name: "Taxonomy: Content", Slug: "content"}, Depth: 0, ChildCount: 3},
		{Tag: models.Tag{ID: uuid.New(), Name: "Taxonomy: Game", Slug: "game"}, Depth: 0, ChildCount: 1},
		{Tag: models.Tag{ID: uuid.New(), Name: "Content: Funny", Slug: "content/funny", ParentSlug: &content, UsageCount: 9}, Depth: 1},
		{Tag: models.Tag{ID: uuid.New(), Name: "Content: Fail", Slug: "content/fail", ParentSlug: &content, UsageCount: 4}, Depth: 1},
	}

	tree := buildTagTree(rows)

	require.Len(t, tree, 2)
	assert.Equal(t, "content", tree[0].Slug)
	assert.Equal(t, 3, tree[0].ChildCount, "child_count reports children beyond the limit")
	require.Len(t, tree[0].Children, 2)
	assert.Equal(t, "content/funny", tree[0].Children[0].Slug)
	assert.Equal(t, "game", tree[1].Slug)
	assert.Empty(t, tree[1].Children)
}
