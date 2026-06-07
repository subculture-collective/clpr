package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"git.subcult.tv/subculture-collective/clpr/internal/models"
	"git.subcult.tv/subculture-collective/clpr/internal/repository"
	"git.subcult.tv/subculture-collective/clpr/internal/services"
)

// MockAuditLogRepository is a mock implementation of the AuditLogRepository interface
type MockAuditLogRepository struct {
	mock.Mock
}

func (m *MockAuditLogRepository) List(ctx context.Context, filters repository.AuditLogFilters, page, limit int) ([]*models.ModerationAuditLogWithUser, int, error) {
	args := m.Called(ctx, filters, page, limit)
	if args.Get(0) == nil {
		return nil, args.Int(1), args.Error(2)
	}
	return args.Get(0).([]*models.ModerationAuditLogWithUser), args.Int(1), args.Error(2)
}

func (m *MockAuditLogRepository) Create(ctx context.Context, log *models.ModerationAuditLog) error {
	args := m.Called(ctx, log)
	return args.Error(0)
}

func (m *MockAuditLogRepository) Export(ctx context.Context, filters repository.AuditLogFilters) ([]*models.ModerationAuditLogWithUser, error) {
	args := m.Called(ctx, filters)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.ModerationAuditLogWithUser), args.Error(1)
}

func (m *MockAuditLogRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.ModerationAuditLogWithUser, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.ModerationAuditLogWithUser), args.Error(1)
}

func TestListModerationAuditLogs(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		queryParams    string
		mockLogs       []*models.ModerationAuditLogWithUser
		mockTotal      int
		mockError      error
		expectedStatus int
		expectedLogs   int
	}{
		{
			name:        "successful list with default pagination",
			queryParams: "",
			mockLogs: []*models.ModerationAuditLogWithUser{
				{
					ModerationAuditLog: models.ModerationAuditLog{
						ID:          uuid.New(),
						Action:      "ban",
						EntityType:  "user",
						EntityID:    uuid.New(),
						ModeratorID: uuid.New(),
						CreatedAt:   time.Now(),
					},
					Moderator: &models.User{
						Username: "test_moderator",
					},
				},
			},
			mockTotal:      1,
			mockError:      nil,
			expectedStatus: http.StatusOK,
			expectedLogs:   1,
		},
		{
			name:        "list with custom limit and offset",
			queryParams: "?limit=10&offset=5",
			mockLogs: []*models.ModerationAuditLogWithUser{
				{
					ModerationAuditLog: models.ModerationAuditLog{
						ID:          uuid.New(),
						Action:      "timeout",
						EntityType:  "user",
						EntityID:    uuid.New(),
						ModeratorID: uuid.New(),
						CreatedAt:   time.Now(),
					},
					Moderator: &models.User{
						Username: "mod2",
					},
				},
			},
			mockTotal:      15,
			mockError:      nil,
			expectedStatus: http.StatusOK,
			expectedLogs:   1,
		},
		{
			name:        "list with action filter",
			queryParams: "?action=ban",
			mockLogs: []*models.ModerationAuditLogWithUser{
				{
					ModerationAuditLog: models.ModerationAuditLog{
						ID:          uuid.New(),
						Action:      "ban",
						EntityType:  "user",
						EntityID:    uuid.New(),
						ModeratorID: uuid.New(),
						CreatedAt:   time.Now(),
					},
					Moderator: &models.User{
						Username: "admin",
					},
				},
			},
			mockTotal:      1,
			mockError:      nil,
			expectedStatus: http.StatusOK,
			expectedLogs:   1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockAuditLogRepository)
			mockRepo.On("List", mock.Anything, mock.AnythingOfType("repository.AuditLogFilters"), mock.AnythingOfType("int"), mock.AnythingOfType("int")).
				Return(tt.mockLogs, tt.mockTotal, tt.mockError)

			service := services.NewAuditLogService(mockRepo)
			handler := NewAuditLogHandler(service)

			router := gin.New()
			router.GET("/api/v1/moderation/audit-logs", handler.ListModerationAuditLogs)

			req := httptest.NewRequest(http.MethodGet, "/api/v1/moderation/audit-logs"+tt.queryParams, nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedStatus == http.StatusOK {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)

				// Check response structure
				assert.Contains(t, response, "logs")
				assert.Contains(t, response, "total")
				assert.Contains(t, response, "limit")
				assert.Contains(t, response, "offset")

				logs := response["logs"].([]interface{})
				assert.Equal(t, tt.expectedLogs, len(logs))

				if len(logs) > 0 {
					// Check log structure
					log := logs[0].(map[string]interface{})
					assert.Contains(t, log, "id")
					assert.Contains(t, log, "action")
					assert.Contains(t, log, "actor")
					assert.Contains(t, log, "target")
					assert.Contains(t, log, "createdAt")
					assert.Contains(t, log, "metadata")
				}
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestGetModerationAuditLog(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		logID          string
		mockLog        *models.ModerationAuditLogWithUser
		mockError      error
		expectedStatus int
	}{
		{
			name:  "successful get",
			logID: uuid.New().String(),
			mockLog: &models.ModerationAuditLogWithUser{
				ModerationAuditLog: models.ModerationAuditLog{
					ID:          uuid.New(),
					Action:      "ban",
					EntityType:  "user",
					EntityID:    uuid.New(),
					ModeratorID: uuid.New(),
					CreatedAt:   time.Now(),
				},
				Moderator: &models.User{
					Username: "test_moderator",
				},
			},
			mockError:      nil,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "invalid UUID",
			logID:          "invalid-uuid",
			mockLog:        nil,
			mockError:      nil,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "not found",
			logID:          uuid.New().String(),
			mockLog:        nil,
			mockError:      pgx.ErrNoRows,
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockAuditLogRepository)
			if tt.logID != "invalid-uuid" {
				mockRepo.On("GetByID", mock.Anything, mock.AnythingOfType("uuid.UUID")).
					Return(tt.mockLog, tt.mockError)
			}

			service := services.NewAuditLogService(mockRepo)
			handler := NewAuditLogHandler(service)

			router := gin.New()
			router.GET("/api/v1/moderation/audit-logs/:id", handler.GetModerationAuditLog)

			req := httptest.NewRequest(http.MethodGet, "/api/v1/moderation/audit-logs/"+tt.logID, nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedStatus == http.StatusOK {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)

				// Check response structure
				assert.Contains(t, response, "id")
				assert.Contains(t, response, "action")
				assert.Contains(t, response, "actor")
				assert.Contains(t, response, "target")
				assert.Contains(t, response, "createdAt")
				assert.Contains(t, response, "metadata")
			}

			if tt.logID != "invalid-uuid" {
				mockRepo.AssertExpectations(t)
			}
		})
	}
}

func TestListModerationAuditLogsPagination(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRepo := new(MockAuditLogRepository)
	mockRepo.On("List", mock.Anything, mock.AnythingOfType("repository.AuditLogFilters"), 1, 20).
		Return([]*models.ModerationAuditLogWithUser{}, 100, nil)

	service := services.NewAuditLogService(mockRepo)
	handler := NewAuditLogHandler(service)

	router := gin.New()
	router.GET("/api/v1/moderation/audit-logs", handler.ListModerationAuditLogs)

	// Test default pagination
	req := httptest.NewRequest(http.MethodGet, "/api/v1/moderation/audit-logs", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, float64(20), response["limit"])
	assert.Equal(t, float64(0), response["offset"])
	assert.Equal(t, float64(100), response["total"])

	mockRepo.AssertExpectations(t)
}

func TestListModerationAuditLogsFilters(t *testing.T) {
	gin.SetMode(gin.TestMode)

	actorID := uuid.New()
	targetID := uuid.New()

	mockRepo := new(MockAuditLogRepository)
	mockRepo.On("List", mock.Anything, mock.MatchedBy(func(f repository.AuditLogFilters) bool {
		return f.Action == "ban" &&
			f.ModeratorID != nil && *f.ModeratorID == actorID &&
			f.EntityID != nil && *f.EntityID == targetID
	}), mock.AnythingOfType("int"), mock.AnythingOfType("int")).
		Return([]*models.ModerationAuditLogWithUser{}, 0, nil)

	service := services.NewAuditLogService(mockRepo)
	handler := NewAuditLogHandler(service)

	router := gin.New()
	router.GET("/api/v1/moderation/audit-logs", handler.ListModerationAuditLogs)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/moderation/audit-logs?action=ban&actor="+actorID.String()+"&target="+targetID.String(), nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	mockRepo.AssertExpectations(t)
}

func TestListModerationAuditLogsSearch(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRepo := new(MockAuditLogRepository)
	mockRepo.On("List", mock.Anything, mock.MatchedBy(func(f repository.AuditLogFilters) bool {
		return f.Search == "harassment"
	}), mock.AnythingOfType("int"), mock.AnythingOfType("int")).
		Return([]*models.ModerationAuditLogWithUser{}, 0, nil)

	service := services.NewAuditLogService(mockRepo)
	handler := NewAuditLogHandler(service)

	router := gin.New()
	router.GET("/api/v1/moderation/audit-logs", handler.ListModerationAuditLogs)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/moderation/audit-logs?search=harassment", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	mockRepo.AssertExpectations(t)
}

func TestExportModerationAuditLogs(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		queryParams    string
		mockLogs       []*models.ModerationAuditLogWithUser
		mockError      error
		expectedStatus int
		expectCSV      bool
	}{
		{
			name:        "successful export",
			queryParams: "?action=ban",
			mockLogs: []*models.ModerationAuditLogWithUser{
				{
					ModerationAuditLog: models.ModerationAuditLog{
						ID:          uuid.New(),
						Action:      "ban",
						EntityType:  "user",
						EntityID:    uuid.New(),
						ModeratorID: uuid.New(),
						CreatedAt:   time.Now(),
					},
					Moderator: &models.User{
						Username: "admin",
					},
				},
			},
			mockError:      nil,
			expectedStatus: http.StatusOK,
			expectCSV:      true,
		},
		{
			name:           "export with error",
			queryParams:    "",
			mockLogs:       nil,
			mockError:      assert.AnError,
			expectedStatus: http.StatusInternalServerError,
			expectCSV:      false,
		},
		{
			name:           "export with invalid filter",
			queryParams:    "?actor=invalid-uuid",
			mockLogs:       nil,
			mockError:      nil,
			expectedStatus: http.StatusBadRequest,
			expectCSV:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockAuditLogRepository)
			if tt.expectedStatus != http.StatusBadRequest {
				mockRepo.On("Export", mock.Anything, mock.AnythingOfType("repository.AuditLogFilters")).
					Return(tt.mockLogs, tt.mockError)
			}

			service := services.NewAuditLogService(mockRepo)
			handler := NewAuditLogHandler(service)

			router := gin.New()
			router.GET("/api/v1/moderation/audit-logs/export", handler.ExportModerationAuditLogs)

			req := httptest.NewRequest(http.MethodGet, "/api/v1/moderation/audit-logs/export"+tt.queryParams, nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectCSV {
				contentType := w.Header().Get("Content-Type")
				assert.Equal(t, "text/csv", contentType)
				contentDisposition := w.Header().Get("Content-Disposition")
				assert.Contains(t, contentDisposition, "attachment")
			}

			if tt.expectedStatus != http.StatusBadRequest {
				mockRepo.AssertExpectations(t)
			}
		})
	}
}
