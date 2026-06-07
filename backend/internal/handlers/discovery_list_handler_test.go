package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"git.subcult.tv/subculture-collective/clpr/internal/models"
)

// ==============================================================================
// Mocks
// ==============================================================================

type MockDiscoveryListRepository struct {
	mock.Mock
}

func (m *MockDiscoveryListRepository) ListDiscoveryLists(ctx context.Context, featuredOnly bool, userID *uuid.UUID, limit, offset int) ([]models.DiscoveryListWithStats, error) {
	args := m.Called(ctx, featuredOnly, userID, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.DiscoveryListWithStats), args.Error(1)
}

func (m *MockDiscoveryListRepository) GetDiscoveryList(ctx context.Context, idOrSlug string, userID *uuid.UUID) (*models.DiscoveryListWithStats, error) {
	args := m.Called(ctx, idOrSlug, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.DiscoveryListWithStats), args.Error(1)
}

func (m *MockDiscoveryListRepository) GetListClips(ctx context.Context, listID uuid.UUID, userID *uuid.UUID, limit, offset int) ([]models.ClipWithSubmitter, int, error) {
	args := m.Called(ctx, listID, userID, limit, offset)
	clips := []models.ClipWithSubmitter{}
	if args.Get(0) != nil {
		clips = args.Get(0).([]models.ClipWithSubmitter)
	}
	return clips, args.Int(1), args.Error(2)
}

func (m *MockDiscoveryListRepository) GetListClipCount(ctx context.Context, listID uuid.UUID) (int, error) {
	args := m.Called(ctx, listID)
	return args.Int(0), args.Error(1)
}

func (m *MockDiscoveryListRepository) GetListClipsForExport(ctx context.Context, listID uuid.UUID, limit int) ([]models.ClipWithSubmitter, error) {
	args := m.Called(ctx, listID, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.ClipWithSubmitter), args.Error(1)
}

func (m *MockDiscoveryListRepository) FollowList(ctx context.Context, userID, listID uuid.UUID) error {
	args := m.Called(ctx, userID, listID)
	return args.Error(0)
}

func (m *MockDiscoveryListRepository) UnfollowList(ctx context.Context, userID, listID uuid.UUID) error {
	args := m.Called(ctx, userID, listID)
	return args.Error(0)
}

func (m *MockDiscoveryListRepository) BookmarkList(ctx context.Context, userID, listID uuid.UUID) error {
	args := m.Called(ctx, userID, listID)
	return args.Error(0)
}

func (m *MockDiscoveryListRepository) UnbookmarkList(ctx context.Context, userID, listID uuid.UUID) error {
	args := m.Called(ctx, userID, listID)
	return args.Error(0)
}

func (m *MockDiscoveryListRepository) GetUserFollowedLists(ctx context.Context, userID uuid.UUID, limit, offset int) ([]models.DiscoveryListWithStats, error) {
	args := m.Called(ctx, userID, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.DiscoveryListWithStats), args.Error(1)
}

func (m *MockDiscoveryListRepository) CreateList(ctx context.Context, name, slug, description string, isFeatured bool, createdBy uuid.UUID) (*models.DiscoveryList, error) {
	args := m.Called(ctx, name, slug, description, isFeatured, createdBy)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.DiscoveryList), args.Error(1)
}

func (m *MockDiscoveryListRepository) UpdateList(ctx context.Context, listID uuid.UUID, name, description *string, isFeatured *bool) (*models.DiscoveryList, error) {
	args := m.Called(ctx, listID, name, description, isFeatured)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.DiscoveryList), args.Error(1)
}

func (m *MockDiscoveryListRepository) DeleteList(ctx context.Context, listID uuid.UUID) error {
	args := m.Called(ctx, listID)
	return args.Error(0)
}

func (m *MockDiscoveryListRepository) AddClipToList(ctx context.Context, listID, clipID uuid.UUID) error {
	args := m.Called(ctx, listID, clipID)
	return args.Error(0)
}

func (m *MockDiscoveryListRepository) RemoveClipFromList(ctx context.Context, listID, clipID uuid.UUID) error {
	args := m.Called(ctx, listID, clipID)
	return args.Error(0)
}

func (m *MockDiscoveryListRepository) ReorderClips(ctx context.Context, listID uuid.UUID, clipIDs []uuid.UUID) error {
	args := m.Called(ctx, listID, clipIDs)
	return args.Error(0)
}

func (m *MockDiscoveryListRepository) ReorderListClips(ctx context.Context, listID uuid.UUID, clipIDs []uuid.UUID) error {
	args := m.Called(ctx, listID, clipIDs)
	return args.Error(0)
}

func (m *MockDiscoveryListRepository) GetListClipsCount(ctx context.Context, listID uuid.UUID) (int, error) {
	args := m.Called(ctx, listID)
	return args.Int(0), args.Error(1)
}

func (m *MockDiscoveryListRepository) ListAllDiscoveryLists(ctx context.Context, limit, offset int) ([]models.DiscoveryListWithStats, error) {
	args := m.Called(ctx, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.DiscoveryListWithStats), args.Error(1)
}

func (m *MockDiscoveryListRepository) CreateDiscoveryList(ctx context.Context, name, slug, description string, isFeatured bool, createdBy uuid.UUID) (*models.DiscoveryList, error) {
	args := m.Called(ctx, name, slug, description, isFeatured, createdBy)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.DiscoveryList), args.Error(1)
}

func (m *MockDiscoveryListRepository) UpdateDiscoveryList(ctx context.Context, listID uuid.UUID, name, description *string, isFeatured *bool) (*models.DiscoveryList, error) {
	args := m.Called(ctx, listID, name, description, isFeatured)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.DiscoveryList), args.Error(1)
}

func (m *MockDiscoveryListRepository) DeleteDiscoveryList(ctx context.Context, listID uuid.UUID) error {
	args := m.Called(ctx, listID)
	return args.Error(0)
}

// ==============================================================================
// ListDiscoveryLists Tests
// ==============================================================================

func TestListDiscoveryLists_DefaultPagination(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRepo := new(MockDiscoveryListRepository)
	mockRepo.On("ListDiscoveryLists", mock.Anything, false, (*uuid.UUID)(nil), 20, 0).
		Return([]models.DiscoveryListWithStats{}, nil)

	handler := &DiscoveryListHandler{
		repo:          mockRepo,
		analyticsRepo: nil,
	}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/discovery-lists", http.NoBody)
	w := httptest.NewRecorder()

	c, _ := gin.CreateTestContext(w)
	c.Request = req

	handler.ListDiscoveryLists(c)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}

	var response []interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	if err != nil {
		t.Errorf("response is not valid JSON: %v", err)
	}
}

func TestListDiscoveryLists_WithFeaturedFilter(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRepo := new(MockDiscoveryListRepository)
	mockRepo.On("ListDiscoveryLists", mock.Anything, true, (*uuid.UUID)(nil), 20, 0).
		Return([]models.DiscoveryListWithStats{}, nil)

	handler := &DiscoveryListHandler{
		repo:          mockRepo,
		analyticsRepo: nil,
	}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/discovery-lists?featured=true", http.NoBody)
	w := httptest.NewRecorder()

	c, _ := gin.CreateTestContext(w)
	c.Request = req

	handler.ListDiscoveryLists(c)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}
}

func TestListDiscoveryLists_WithCustomPagination(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRepo := new(MockDiscoveryListRepository)
	mockRepo.On("ListDiscoveryLists", mock.Anything, false, (*uuid.UUID)(nil), mock.Anything, mock.Anything).
		Return([]models.DiscoveryListWithStats{}, nil)

	handler := &DiscoveryListHandler{
		repo:          mockRepo,
		analyticsRepo: nil,
	}

	tests := []struct {
		name   string
		url    string
		status int
	}{
		{
			name:   "Valid pagination",
			url:    "/api/v1/discovery-lists?limit=10&offset=5",
			status: http.StatusOK,
		},
		{
			name:   "Limit exceeds max",
			url:    "/api/v1/discovery-lists?limit=200",
			status: http.StatusOK, // Should be clamped to maximum allowed limit (100)
		},
		{
			name:   "Zero limit",
			url:    "/api/v1/discovery-lists?limit=0",
			status: http.StatusOK, // Should use default limit (20)
		},
		{
			name:   "Negative offset",
			url:    "/api/v1/discovery-lists?offset=-5",
			status: http.StatusOK, // Should default to 0
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tt.url, http.NoBody)
			w := httptest.NewRecorder()

			c, _ := gin.CreateTestContext(w)
			c.Request = req

			handler.ListDiscoveryLists(c)

			if w.Code != tt.status {
				t.Errorf("expected status %d, got %d", tt.status, w.Code)
			}
		})
	}
}

// ==============================================================================
// GetDiscoveryList Tests
// ==============================================================================

func TestGetDiscoveryList_NotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRepo := new(MockDiscoveryListRepository)
	mockRepo.On("GetDiscoveryList", mock.Anything, mock.Anything, (*uuid.UUID)(nil)).
		Return(nil, errors.New("discovery list not found"))

	handler := &DiscoveryListHandler{
		repo:          mockRepo,
		analyticsRepo: nil,
	}

	listID := uuid.New().String()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/discovery-lists/"+listID, http.NoBody)
	w := httptest.NewRecorder()

	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{
		{Key: "id", Value: listID},
	}

	handler.GetDiscoveryList(c)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected status %d for non-existent list, got %d", http.StatusNotFound, w.Code)
	}
}

func TestGetDiscoveryList_WithSlug(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRepo := new(MockDiscoveryListRepository)
	mockRepo.On("GetDiscoveryList", mock.Anything, "top-clips", (*uuid.UUID)(nil)).
		Return(nil, errors.New("discovery list not found"))

	handler := &DiscoveryListHandler{
		repo:          mockRepo,
		analyticsRepo: nil,
	}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/discovery-lists/top-clips", http.NoBody)
	w := httptest.NewRecorder()

	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{
		{Key: "id", Value: "top-clips"},
	}

	handler.GetDiscoveryList(c)

	// Repository stub returns not found, which is expected
	if w.Code != http.StatusNotFound {
		t.Errorf("expected status %d for non-existent list, got %d", http.StatusNotFound, w.Code)
	}
}

// ==============================================================================
// GetDiscoveryListClips Tests
// ==============================================================================

func TestGetDiscoveryListClips_InvalidListID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRepo := new(MockDiscoveryListRepository)
	mockRepo.On("GetDiscoveryList", mock.Anything, "invalid-uuid", (*uuid.UUID)(nil)).
		Return(nil, errors.New("discovery list not found"))

	handler := &DiscoveryListHandler{
		repo:          mockRepo,
		analyticsRepo: nil,
	}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/discovery-lists/invalid-uuid/clips", http.NoBody)
	w := httptest.NewRecorder()

	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{
		{Key: "id", Value: "invalid-uuid"},
	}

	handler.GetDiscoveryListClips(c)

	// Invalid UUID should try to resolve as slug, which will return not found
	if w.Code != http.StatusNotFound {
		t.Errorf("expected status %d for invalid list ID, got %d", http.StatusNotFound, w.Code)
	}
}

func TestGetDiscoveryListClips_WithPagination(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRepo := new(MockDiscoveryListRepository)
	mockRepo.On("GetDiscoveryList", mock.Anything, mock.Anything, (*uuid.UUID)(nil)).
		Return(&models.DiscoveryListWithStats{}, nil)
	mockRepo.On("GetListClips", mock.Anything, mock.Anything, (*uuid.UUID)(nil), 50, 10).
		Return([]models.ClipWithSubmitter{}, 0, nil)

	handler := &DiscoveryListHandler{
		repo:          mockRepo,
		analyticsRepo: nil,
	}

	listID := uuid.New().String()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/discovery-lists/"+listID+"/clips?limit=50&offset=10", http.NoBody)
	w := httptest.NewRecorder()

	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{
		{Key: "id", Value: listID},
	}

	handler.GetDiscoveryListClips(c)

	// Repository stub will succeed with empty response
	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	if err != nil {
		t.Errorf("response is not valid JSON: %v", err)
	}

	// Verify response structure
	if _, exists := response["clips"]; !exists {
		t.Error("expected clips field in response")
	}
	if _, exists := response["total"]; !exists {
		t.Error("expected total field in response")
	}
	if _, exists := response["has_more"]; !exists {
		t.Error("expected has_more field in response")
	}
}

func TestGetDiscoveryListClips_BoundaryValues(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRepo := new(MockDiscoveryListRepository)
	mockRepo.On("GetDiscoveryList", mock.Anything, mock.Anything, (*uuid.UUID)(nil)).
		Return(&models.DiscoveryListWithStats{}, nil)
	mockRepo.On("GetListClips", mock.Anything, mock.Anything, (*uuid.UUID)(nil), mock.Anything, mock.Anything).
		Return([]models.ClipWithSubmitter{}, 0, nil)

	handler := &DiscoveryListHandler{
		repo:          mockRepo,
		analyticsRepo: nil,
	}

	tests := []struct {
		name   string
		limit  string
		offset string
	}{
		{
			name:   "Max limit",
			limit:  "100",
			offset: "0",
		},
		{
			name:   "Over max limit",
			limit:  "200",
			offset: "0",
		},
		{
			name:   "Zero limit",
			limit:  "0",
			offset: "0",
		},
		{
			name:   "Negative offset",
			limit:  "20",
			offset: "-10",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			listID := uuid.New().String()
			req := httptest.NewRequest(http.MethodGet, "/api/v1/discovery-lists/"+listID+"/clips?limit="+tt.limit+"&offset="+tt.offset, http.NoBody)
			w := httptest.NewRecorder()

			c, _ := gin.CreateTestContext(w)
			c.Request = req
			c.Params = gin.Params{
				{Key: "id", Value: listID},
			}

			handler.GetDiscoveryListClips(c)

			if w.Code != http.StatusOK {
				t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
			}
		})
	}
}

// ==============================================================================
// FollowDiscoveryList Tests
// ==============================================================================

func TestFollowDiscoveryList_Unauthenticated(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRepo := new(MockDiscoveryListRepository)

	handler := &DiscoveryListHandler{
		repo:          mockRepo,
		analyticsRepo: nil,
	}

	listID := uuid.New().String()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/discovery-lists/"+listID+"/follow", http.NoBody)
	w := httptest.NewRecorder()

	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{
		{Key: "id", Value: listID},
	}
	// Don't set user_id to simulate unauthenticated request

	handler.FollowDiscoveryList(c)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected status %d for unauthenticated request, got %d", http.StatusUnauthorized, w.Code)
	}
}

func TestFollowDiscoveryList_InvalidListID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRepo := new(MockDiscoveryListRepository)

	handler := &DiscoveryListHandler{
		repo:          mockRepo,
		analyticsRepo: nil,
	}

	req := httptest.NewRequest(http.MethodPost, "/api/v1/discovery-lists/invalid-uuid/follow", http.NoBody)
	w := httptest.NewRecorder()

	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{
		{Key: "id", Value: "invalid-uuid"},
	}
	c.Set("user_id", uuid.New())

	handler.FollowDiscoveryList(c)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status %d for invalid list ID, got %d", http.StatusBadRequest, w.Code)
	}
}

// ==============================================================================
// UnfollowDiscoveryList Tests
// ==============================================================================

func TestUnfollowDiscoveryList_Unauthenticated(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRepo := new(MockDiscoveryListRepository)

	handler := &DiscoveryListHandler{
		repo:          mockRepo,
		analyticsRepo: nil,
	}

	listID := uuid.New().String()
	req := httptest.NewRequest(http.MethodDelete, "/api/v1/discovery-lists/"+listID+"/follow", http.NoBody)
	w := httptest.NewRecorder()

	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{
		{Key: "id", Value: listID},
	}

	handler.UnfollowDiscoveryList(c)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected status %d for unauthenticated request, got %d", http.StatusUnauthorized, w.Code)
	}
}

func TestUnfollowDiscoveryList_InvalidListID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRepo := new(MockDiscoveryListRepository)

	handler := &DiscoveryListHandler{
		repo:          mockRepo,
		analyticsRepo: nil,
	}

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/discovery-lists/invalid-uuid/follow", http.NoBody)
	w := httptest.NewRecorder()

	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{
		{Key: "id", Value: "invalid-uuid"},
	}
	c.Set("user_id", uuid.New())

	handler.UnfollowDiscoveryList(c)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status %d for invalid list ID, got %d", http.StatusBadRequest, w.Code)
	}
}

// ==============================================================================
// BookmarkDiscoveryList Tests
// ==============================================================================

func TestBookmarkDiscoveryList_Unauthenticated(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRepo := new(MockDiscoveryListRepository)

	handler := &DiscoveryListHandler{
		repo:          mockRepo,
		analyticsRepo: nil,
	}

	listID := uuid.New().String()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/discovery-lists/"+listID+"/bookmark", http.NoBody)
	w := httptest.NewRecorder()

	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{
		{Key: "id", Value: listID},
	}

	handler.BookmarkDiscoveryList(c)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected status %d for unauthenticated request, got %d", http.StatusUnauthorized, w.Code)
	}
}

func TestBookmarkDiscoveryList_InvalidListID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRepo := new(MockDiscoveryListRepository)

	handler := &DiscoveryListHandler{
		repo:          mockRepo,
		analyticsRepo: nil,
	}

	req := httptest.NewRequest(http.MethodPost, "/api/v1/discovery-lists/invalid-uuid/bookmark", http.NoBody)
	w := httptest.NewRecorder()

	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{
		{Key: "id", Value: "invalid-uuid"},
	}
	c.Set("user_id", uuid.New())

	handler.BookmarkDiscoveryList(c)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status %d for invalid list ID, got %d", http.StatusBadRequest, w.Code)
	}
}

// ==============================================================================
// UnbookmarkDiscoveryList Tests
// ==============================================================================

func TestUnbookmarkDiscoveryList_Unauthenticated(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRepo := new(MockDiscoveryListRepository)

	handler := &DiscoveryListHandler{
		repo:          mockRepo,
		analyticsRepo: nil,
	}

	listID := uuid.New().String()
	req := httptest.NewRequest(http.MethodDelete, "/api/v1/discovery-lists/"+listID+"/bookmark", http.NoBody)
	w := httptest.NewRecorder()

	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{
		{Key: "id", Value: listID},
	}

	handler.UnbookmarkDiscoveryList(c)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected status %d for unauthenticated request, got %d", http.StatusUnauthorized, w.Code)
	}
}

// ==============================================================================
// GetUserFollowedLists Tests
// ==============================================================================

func TestGetUserFollowedLists_Unauthenticated(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRepo := new(MockDiscoveryListRepository)

	handler := &DiscoveryListHandler{
		repo:          mockRepo,
		analyticsRepo: nil,
	}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/users/me/discovery-list-follows", http.NoBody)
	w := httptest.NewRecorder()

	c, _ := gin.CreateTestContext(w)
	c.Request = req

	handler.GetUserFollowedLists(c)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected status %d for unauthenticated request, got %d", http.StatusUnauthorized, w.Code)
	}
}

func TestGetUserFollowedLists_WithPagination(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRepo := new(MockDiscoveryListRepository)
	mockRepo.On("GetUserFollowedLists", mock.Anything, mock.Anything, 10, 5).
		Return([]models.DiscoveryListWithStats{}, nil)

	handler := &DiscoveryListHandler{
		repo:          mockRepo,
		analyticsRepo: nil,
	}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/users/me/discovery-list-follows?limit=10&offset=5", http.NoBody)
	w := httptest.NewRecorder()

	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("user_id", uuid.New())

	handler.GetUserFollowedLists(c)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}
}
