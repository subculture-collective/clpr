package services

import (
	"bytes"
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"git.subcult.tv/subculture-collective/clpr/internal/models"
	"git.subcult.tv/subculture-collective/clpr/internal/repository"
)

// TestParseAuditLogFilters tests the filter parsing function
func TestParseAuditLogFilters(t *testing.T) {
	tests := []struct {
		name        string
		moderatorID string
		action      string
		entityType  string
		entityID    string
		channelID   string
		startDate   string
		endDate     string
		expectError bool
	}{
		{
			name:        "Valid filters",
			moderatorID: uuid.New().String(),
			action:      "ban",
			entityType:  "user",
			entityID:    uuid.New().String(),
			channelID:   uuid.New().String(),
			startDate:   time.Now().Add(-24 * time.Hour).Format(time.RFC3339),
			endDate:     time.Now().Format(time.RFC3339),
			expectError: false,
		},
		{
			name:        "Empty filters",
			moderatorID: "",
			action:      "",
			entityType:  "",
			entityID:    "",
			channelID:   "",
			startDate:   "",
			endDate:     "",
			expectError: false,
		},
		{
			name:        "Invalid moderator ID",
			moderatorID: "invalid-uuid",
			action:      "",
			entityType:  "",
			entityID:    "",
			channelID:   "",
			startDate:   "",
			endDate:     "",
			expectError: true,
		},
		{
			name:        "Invalid entity ID",
			moderatorID: "",
			action:      "",
			entityType:  "",
			entityID:    "invalid-uuid",
			channelID:   "",
			startDate:   "",
			endDate:     "",
			expectError: true,
		},
		{
			name:        "Invalid channel ID",
			moderatorID: "",
			action:      "",
			entityType:  "",
			entityID:    "",
			channelID:   "invalid-uuid",
			startDate:   "",
			endDate:     "",
			expectError: true,
		},
		{
			name:        "Invalid start date",
			moderatorID: "",
			action:      "",
			entityType:  "",
			entityID:    "",
			channelID:   "",
			startDate:   "invalid-date",
			endDate:     "",
			expectError: true,
		},
		{
			name:        "Invalid end date",
			moderatorID: "",
			action:      "",
			entityType:  "",
			entityID:    "",
			channelID:   "",
			startDate:   "",
			endDate:     "invalid-date",
			expectError: true,
		},
		{
			name:        "Partial filters - action and entity type only",
			moderatorID: "",
			action:      "timeout",
			entityType:  "message",
			entityID:    "",
			channelID:   "",
			startDate:   "",
			endDate:     "",
			expectError: false,
		},
		{
			name:        "Partial filters - date range only",
			moderatorID: "",
			action:      "",
			entityType:  "",
			entityID:    "",
			channelID:   "",
			startDate:   time.Now().Add(-7 * 24 * time.Hour).Format(time.RFC3339),
			endDate:     time.Now().Format(time.RFC3339),
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filters, err := ParseAuditLogFilters(
				tt.moderatorID,
				tt.action,
				tt.entityType,
				tt.entityID,
				tt.channelID,
				tt.startDate,
				tt.endDate,
				"", // search parameter
			)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				if tt.moderatorID != "" {
					assert.NotNil(t, filters.ModeratorID)
					assert.Equal(t, tt.moderatorID, filters.ModeratorID.String())
				}
				if tt.action != "" {
					assert.Equal(t, tt.action, filters.Action)
				}
				if tt.entityType != "" {
					assert.Equal(t, tt.entityType, filters.EntityType)
				}
				if tt.entityID != "" {
					assert.NotNil(t, filters.EntityID)
					assert.Equal(t, tt.entityID, filters.EntityID.String())
				}
				if tt.channelID != "" {
					assert.NotNil(t, filters.ChannelID)
					assert.Equal(t, tt.channelID, filters.ChannelID.String())
				}
				if tt.startDate != "" {
					assert.NotNil(t, filters.StartDate)
				}
				if tt.endDate != "" {
					assert.NotNil(t, filters.EndDate)
				}
			}
		})
	}
}

// TestParseAuditLogFiltersDateRangeOrder tests that date ranges are correctly parsed
func TestParseAuditLogFiltersDateRangeOrder(t *testing.T) {
	startDate := time.Now().Add(-7 * 24 * time.Hour)
	endDate := time.Now()

	filters, err := ParseAuditLogFilters(
		"",
		"",
		"",
		"",
		"",
		startDate.Format(time.RFC3339),
		endDate.Format(time.RFC3339),
		"",
	)

	assert.NoError(t, err)
	assert.NotNil(t, filters.StartDate)
	assert.NotNil(t, filters.EndDate)
	assert.True(t, filters.StartDate.Before(*filters.EndDate), "start date should be before end date")
}

// TestAuditLogFiltersStructure tests the AuditLogFilters structure
func TestAuditLogFiltersStructure(t *testing.T) {
	moderatorID := uuid.New()
	entityID := uuid.New()
	channelID := uuid.New()
	startDate := time.Now().Add(-24 * time.Hour)
	endDate := time.Now()

	filters := repository.AuditLogFilters{
		ModeratorID: &moderatorID,
		Action:      "ban",
		EntityType:  "user",
		EntityID:    &entityID,
		ChannelID:   &channelID,
		StartDate:   &startDate,
		EndDate:     &endDate,
	}

	assert.Equal(t, "ban", filters.Action)
	assert.Equal(t, "user", filters.EntityType)
	assert.NotNil(t, filters.ModeratorID)
	assert.Equal(t, moderatorID, *filters.ModeratorID)
	assert.NotNil(t, filters.EntityID)
	assert.Equal(t, entityID, *filters.EntityID)
	assert.NotNil(t, filters.ChannelID)
	assert.Equal(t, channelID, *filters.ChannelID)
	assert.NotNil(t, filters.StartDate)
	assert.NotNil(t, filters.EndDate)
}

// TestAuditLogFiltersOptional tests that filters work with optional fields
func TestAuditLogFiltersOptional(t *testing.T) {
	filters := repository.AuditLogFilters{
		Action:     "approve",
		EntityType: "clip_submission",
	}

	assert.Equal(t, "approve", filters.Action)
	assert.Equal(t, "clip_submission", filters.EntityType)
	assert.Nil(t, filters.ModeratorID)
	assert.Nil(t, filters.EntityID)
	assert.Nil(t, filters.ChannelID)
	assert.Nil(t, filters.StartDate)
	assert.Nil(t, filters.EndDate)
}

// TestAuditLogFiltersEmpty tests empty filters
func TestAuditLogFiltersEmpty(t *testing.T) {
	filters := repository.AuditLogFilters{}

	assert.Equal(t, "", filters.Action)
	assert.Equal(t, "", filters.EntityType)
	assert.Nil(t, filters.ModeratorID)
	assert.Nil(t, filters.EntityID)
	assert.Nil(t, filters.ChannelID)
	assert.Nil(t, filters.StartDate)
	assert.Nil(t, filters.EndDate)
}

// Mock AuditLogRepository for testing service methods
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

// TestAuditLogService_GetAuditLogs tests retrieving audit logs with filters
func TestAuditLogService_GetAuditLogs(t *testing.T) {
	ctx := context.Background()
	mockRepo := new(MockAuditLogRepository)

	moderatorID := uuid.New()
	filters := repository.AuditLogFilters{
		ModeratorID: &moderatorID,
		Action:      "ban",
	}

	expectedLogs := []*models.ModerationAuditLogWithUser{
		{
			ModerationAuditLog: models.ModerationAuditLog{
				ID:          uuid.New(),
				Action:      "ban",
				EntityType:  "user",
				EntityID:    uuid.New(),
				ModeratorID: moderatorID,
				CreatedAt:   time.Now(),
			},
			Moderator: &models.User{
				ID:       moderatorID,
				Username: "test_mod",
			},
		},
	}

	mockRepo.On("List", ctx, filters, 1, 10).Return(expectedLogs, 1, nil)

	service := NewAuditLogService(mockRepo)
	logs, total, err := service.GetAuditLogs(ctx, filters, 1, 10)

	assert.NoError(t, err)
	assert.Equal(t, 1, total)
	assert.Equal(t, 1, len(logs))
	assert.Equal(t, "ban", logs[0].Action)

	mockRepo.AssertExpectations(t)
}

// TestAuditLogService_LogAction tests logging a moderation action
func TestAuditLogService_LogAction(t *testing.T) {
	ctx := context.Background()
	mockRepo := new(MockAuditLogRepository)

	actorID := uuid.New()
	targetID := uuid.New()
	channelID := uuid.New()
	reason := "spam"

	mockRepo.On("Create", ctx, mock.MatchedBy(func(log *models.ModerationAuditLog) bool {
		return log.Action == "ban_user" &&
			log.EntityType == "user" &&
			log.ModeratorID == actorID &&
			log.EntityID == targetID &&
			log.Reason != nil &&
			*log.Reason == reason &&
			log.ChannelID != nil &&
			*log.ChannelID == channelID
	})).Return(nil)

	service := NewAuditLogService(mockRepo)

	opts := AuditLogOptions{
		Channel: &channelID,
		Reason:  &reason,
		Metadata: map[string]interface{}{
			"key": "value",
		},
	}

	err := service.LogAction(ctx, "ban_user", actorID, targetID, "user", opts)
	assert.NoError(t, err)

	mockRepo.AssertExpectations(t)
}

// TestAuditLogService_LogAction_MinimalOptions tests logging with minimal options
func TestAuditLogService_LogAction_MinimalOptions(t *testing.T) {
	ctx := context.Background()
	mockRepo := new(MockAuditLogRepository)

	actorID := uuid.New()
	targetID := uuid.New()

	mockRepo.On("Create", ctx, mock.MatchedBy(func(log *models.ModerationAuditLog) bool {
		return log.Action == "view_logs" &&
			log.EntityType == "audit_log" &&
			log.ModeratorID == actorID &&
			log.EntityID == targetID &&
			log.Reason == nil &&
			log.ChannelID == nil
	})).Return(nil)

	service := NewAuditLogService(mockRepo)

	opts := AuditLogOptions{}
	err := service.LogAction(ctx, "view_logs", actorID, targetID, "audit_log", opts)
	assert.NoError(t, err)

	mockRepo.AssertExpectations(t)
}

// TestAuditLogService_ExportAuditLogsCSV tests CSV export functionality
func TestAuditLogService_ExportAuditLogsCSV(t *testing.T) {
	ctx := context.Background()
	mockRepo := new(MockAuditLogRepository)

	moderatorID := uuid.New()
	entityID := uuid.New()
	channelID := uuid.New()
	reason := "test reason"
	ipAddress := "192.168.1.1"
	userAgent := "test-agent"

	filters := repository.AuditLogFilters{
		Action: "ban",
	}

	exportedLogs := []*models.ModerationAuditLogWithUser{
		{
			ModerationAuditLog: models.ModerationAuditLog{
				ID:          uuid.New(),
				Action:      "ban",
				EntityType:  "user",
				EntityID:    entityID,
				ModeratorID: moderatorID,
				Reason:      &reason,
				Metadata: map[string]interface{}{
					"key": "value",
				},
				IPAddress: &ipAddress,
				UserAgent: &userAgent,
				ChannelID: &channelID,
				CreatedAt: time.Now(),
			},
			Moderator: &models.User{
				ID:       moderatorID,
				Username: "test_mod",
			},
		},
	}

	mockRepo.On("Export", ctx, filters).Return(exportedLogs, nil)

	service := NewAuditLogService(mockRepo)

	var buf bytes.Buffer
	err := service.ExportAuditLogsCSV(ctx, filters, &buf)

	assert.NoError(t, err)
	assert.Contains(t, buf.String(), "ID,Action,Entity Type")
	assert.Contains(t, buf.String(), "ban")
	assert.Contains(t, buf.String(), "test_mod")

	mockRepo.AssertExpectations(t)
}

// TestAuditLogService_ExportAuditLogsCSV_EmptyResults tests CSV export with no results
func TestAuditLogService_ExportAuditLogsCSV_EmptyResults(t *testing.T) {
	ctx := context.Background()
	mockRepo := new(MockAuditLogRepository)

	filters := repository.AuditLogFilters{
		Action: "nonexistent",
	}

	mockRepo.On("Export", ctx, filters).Return([]*models.ModerationAuditLogWithUser{}, nil)

	service := NewAuditLogService(mockRepo)

	var buf bytes.Buffer
	err := service.ExportAuditLogsCSV(ctx, filters, &buf)

	assert.NoError(t, err)
	assert.Contains(t, buf.String(), "ID,Action,Entity Type") // Header should still be present

	mockRepo.AssertExpectations(t)
}

// TestAuditLogService_LogSubscriptionEvent tests logging subscription events
func TestAuditLogService_LogSubscriptionEvent(t *testing.T) {
	ctx := context.Background()
	mockRepo := new(MockAuditLogRepository)

	userID := uuid.New()
	metadata := map[string]interface{}{
		"plan":   "premium",
		"amount": 9.99,
	}

	mockRepo.On("Create", ctx, mock.MatchedBy(func(log *models.ModerationAuditLog) bool {
		return log.Action == "subscription_created" &&
			log.EntityType == "subscription" &&
			log.ModeratorID == userID &&
			log.EntityID == userID &&
			log.Metadata != nil
	})).Return(nil)

	service := NewAuditLogService(mockRepo)
	err := service.LogSubscriptionEvent(ctx, userID, "subscription_created", metadata)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

// TestAuditLogService_LogAccountDeletionRequested tests logging account deletion request
func TestAuditLogService_LogAccountDeletionRequested(t *testing.T) {
	ctx := context.Background()
	mockRepo := new(MockAuditLogRepository)

	userID := uuid.New()
	reason := "no longer needed"

	mockRepo.On("Create", ctx, mock.MatchedBy(func(log *models.ModerationAuditLog) bool {
		return log.Action == "account_deletion_requested" &&
			log.EntityType == "user" &&
			log.ModeratorID == userID &&
			log.EntityID == userID &&
			log.Reason != nil &&
			*log.Reason == reason
	})).Return(nil)

	service := NewAuditLogService(mockRepo)
	err := service.LogAccountDeletionRequested(ctx, userID, &reason)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

// TestAuditLogService_LogAccountDeletionCancelled tests logging account deletion cancellation
func TestAuditLogService_LogAccountDeletionCancelled(t *testing.T) {
	ctx := context.Background()
	mockRepo := new(MockAuditLogRepository)

	userID := uuid.New()

	mockRepo.On("Create", ctx, mock.MatchedBy(func(log *models.ModerationAuditLog) bool {
		return log.Action == "account_deletion_cancelled" &&
			log.EntityType == "user" &&
			log.ModeratorID == userID &&
			log.EntityID == userID
	})).Return(nil)

	service := NewAuditLogService(mockRepo)
	err := service.LogAccountDeletionCancelled(ctx, userID)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

// TestAuditLogService_LogEntitlementDenial tests logging entitlement denial
func TestAuditLogService_LogEntitlementDenial(t *testing.T) {
	ctx := context.Background()
	mockRepo := new(MockAuditLogRepository)

	userID := uuid.New()
	metadata := map[string]interface{}{
		"feature": "premium_export",
		"tier":    "free",
	}

	mockRepo.On("Create", ctx, mock.MatchedBy(func(log *models.ModerationAuditLog) bool {
		return log.Action == "entitlement_denied" &&
			log.EntityType == "entitlement" &&
			log.ModeratorID == userID &&
			log.EntityID == userID &&
			log.Metadata != nil
	})).Return(nil)

	service := NewAuditLogService(mockRepo)
	err := service.LogEntitlementDenial(ctx, userID, "entitlement_denied", metadata)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

// TestAuditLogService_LogClipMetadataUpdate tests logging clip metadata update
func TestAuditLogService_LogClipMetadataUpdate(t *testing.T) {
	ctx := context.Background()
	mockRepo := new(MockAuditLogRepository)

	userID := uuid.New()
	clipID := uuid.New()
	metadata := map[string]interface{}{
		"field": "title",
		"old":   "Old Title",
		"new":   "New Title",
	}

	mockRepo.On("Create", ctx, mock.MatchedBy(func(log *models.ModerationAuditLog) bool {
		return log.Action == "clip_metadata_updated" &&
			log.EntityType == "clip" &&
			log.ModeratorID == userID &&
			log.EntityID == clipID &&
			log.Metadata != nil
	})).Return(nil)

	service := NewAuditLogService(mockRepo)
	err := service.LogClipMetadataUpdate(ctx, userID, clipID, metadata)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

// TestAuditLogService_LogClipVisibilityChange tests logging clip visibility change
func TestAuditLogService_LogClipVisibilityChange(t *testing.T) {
	ctx := context.Background()
	mockRepo := new(MockAuditLogRepository)

	userID := uuid.New()
	clipID := uuid.New()
	isHidden := true

	mockRepo.On("Create", ctx, mock.MatchedBy(func(log *models.ModerationAuditLog) bool {
		return log.Action == "clip_hidden" &&
			log.EntityType == "clip" &&
			log.ModeratorID == userID &&
			log.EntityID == clipID &&
			log.Metadata != nil
	})).Return(nil)

	service := NewAuditLogService(mockRepo)
	err := service.LogClipVisibilityChange(ctx, userID, clipID, isHidden)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

// TestAuditLogService_LogClipVisibilityChange_Unhide tests logging clip unhide
func TestAuditLogService_LogClipVisibilityChange_Unhide(t *testing.T) {
	ctx := context.Background()
	mockRepo := new(MockAuditLogRepository)

	userID := uuid.New()
	clipID := uuid.New()
	isHidden := false

	mockRepo.On("Create", ctx, mock.MatchedBy(func(log *models.ModerationAuditLog) bool {
		return log.Action == "clip_unhidden" &&
			log.EntityType == "clip" &&
			log.ModeratorID == userID &&
			log.EntityID == clipID &&
			log.Metadata != nil
	})).Return(nil)

	service := NewAuditLogService(mockRepo)
	err := service.LogClipVisibilityChange(ctx, userID, clipID, isHidden)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

// TestAuditLogService_GetAuditLogByID tests retrieving a single audit log by ID
func TestAuditLogService_GetAuditLogByID(t *testing.T) {
	ctx := context.Background()
	mockRepo := new(MockAuditLogRepository)

	logID := uuid.New()
	moderatorID := uuid.New()

	expectedLog := &models.ModerationAuditLogWithUser{
		ModerationAuditLog: models.ModerationAuditLog{
			ID:          logID,
			Action:      "ban",
			EntityType:  "user",
			EntityID:    uuid.New(),
			ModeratorID: moderatorID,
			CreatedAt:   time.Now(),
		},
		Moderator: &models.User{
			ID:       moderatorID,
			Username: "test_mod",
		},
	}

	mockRepo.On("GetByID", ctx, logID).Return(expectedLog, nil)

	service := NewAuditLogService(mockRepo)
	log, err := service.GetAuditLogByID(ctx, logID)

	assert.NoError(t, err)
	assert.NotNil(t, log)
	assert.Equal(t, logID, log.ID)
	assert.Equal(t, "ban", log.Action)

	mockRepo.AssertExpectations(t)
}

// TestAuditLogService_GetAuditLogByID_NotFound tests retrieving non-existent audit log
func TestAuditLogService_GetAuditLogByID_NotFound(t *testing.T) {
	ctx := context.Background()
	mockRepo := new(MockAuditLogRepository)

	logID := uuid.New()

	mockRepo.On("GetByID", ctx, logID).Return(nil, errors.New("not found"))

	service := NewAuditLogService(mockRepo)
	log, err := service.GetAuditLogByID(ctx, logID)

	assert.Error(t, err)
	assert.Nil(t, log)

	mockRepo.AssertExpectations(t)
}
