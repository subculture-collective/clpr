package services

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"git.subcult.tv/subculture-collective/clpr/internal/models"
)

// mockExportRepository is a mock implementation of ExportRepository for testing
type mockExportRepository struct {
	requests map[uuid.UUID]*models.ExportRequest
	clips    map[string][]*models.Clip
}

func newMockExportRepository() *mockExportRepository {
	return &mockExportRepository{
		requests: make(map[uuid.UUID]*models.ExportRequest),
		clips:    make(map[string][]*models.Clip),
	}
}

func (m *mockExportRepository) CreateExportRequest(ctx context.Context, req *models.ExportRequest) error {
	m.requests[req.ID] = req
	return nil
}

func (m *mockExportRepository) GetExportRequestByID(ctx context.Context, id uuid.UUID) (*models.ExportRequest, error) {
	req, ok := m.requests[id]
	if !ok {
		return nil, os.ErrNotExist
	}
	return req, nil
}

func (m *mockExportRepository) GetUserExportRequests(ctx context.Context, userID uuid.UUID, limit int) ([]*models.ExportRequest, error) {
	var results []*models.ExportRequest
	for _, req := range m.requests {
		if req.UserID == userID {
			results = append(results, req)
		}
	}
	return results, nil
}

func (m *mockExportRepository) UpdateExportStatus(ctx context.Context, id uuid.UUID, status string, errorMsg *string) error {
	if req, ok := m.requests[id]; ok {
		req.Status = status
		req.ErrorMessage = errorMsg
		req.UpdatedAt = time.Now()
	}
	return nil
}

func (m *mockExportRepository) CompleteExportRequest(ctx context.Context, id uuid.UUID, filePath string, fileSize int64, expiresAt time.Time) error {
	if req, ok := m.requests[id]; ok {
		req.Status = models.ExportStatusCompleted
		req.FilePath = &filePath
		req.FileSizeBytes = &fileSize
		req.ExpiresAt = &expiresAt
		now := time.Now()
		req.CompletedAt = &now
		req.UpdatedAt = now
	}
	return nil
}

func (m *mockExportRepository) MarkEmailSent(ctx context.Context, id uuid.UUID) error {
	if req, ok := m.requests[id]; ok {
		req.EmailSent = true
		req.UpdatedAt = time.Now()
	}
	return nil
}

func (m *mockExportRepository) GetPendingExportRequests(ctx context.Context, limit int) ([]*models.ExportRequest, error) {
	var results []*models.ExportRequest
	for _, req := range m.requests {
		if req.Status == models.ExportStatusPending {
			results = append(results, req)
		}
	}
	return results, nil
}

func (m *mockExportRepository) GetExpiredExportRequests(ctx context.Context) ([]*models.ExportRequest, error) {
	var results []*models.ExportRequest
	now := time.Now()
	for _, req := range m.requests {
		if req.Status == models.ExportStatusCompleted && req.ExpiresAt != nil && req.ExpiresAt.Before(now) {
			results = append(results, req)
		}
	}
	return results, nil
}

func (m *mockExportRepository) MarkExportExpired(ctx context.Context, id uuid.UUID) error {
	if req, ok := m.requests[id]; ok {
		req.Status = models.ExportStatusExpired
		req.UpdatedAt = time.Now()
	}
	return nil
}

func (m *mockExportRepository) GetCreatorClipsForExport(ctx context.Context, creatorName string) ([]*models.Clip, error) {
	clips, ok := m.clips[creatorName]
	if !ok {
		return []*models.Clip{}, nil
	}
	return clips, nil
}

func (m *mockExportRepository) addTestClips(creatorName string, count int) {
	clips := make([]*models.Clip, count)
	for i := 0; i < count; i++ {
		gameName := "Test Game"
		language := "en"
		duration := 30.0
		thumbnailURL := "https://example.com/thumb.jpg"
		clips[i] = &models.Clip{
			ID:              uuid.New(),
			TwitchClipID:    uuid.New().String(),
			TwitchClipURL:   "https://twitch.tv/clip/" + uuid.New().String(),
			EmbedURL:        "https://clips.twitch.tv/embed?clip=" + uuid.New().String(),
			Title:           "Test Clip",
			CreatorName:     creatorName,
			BroadcasterName: "TestBroadcaster",
			GameName:        &gameName,
			Language:        &language,
			ThumbnailURL:    &thumbnailURL,
			Duration:        &duration,
			ViewCount:       100,
			VoteScore:       10,
			CommentCount:    5,
			FavoriteCount:   3,
			IsFeatured:      false,
			IsNSFW:          false,
			IsRemoved:       false,
			IsHidden:        false,
			CreatedAt:       time.Now().Add(-time.Duration(i) * 24 * time.Hour),
			ImportedAt:      time.Now(),
		}
	}
	m.clips[creatorName] = clips
}

// mockExportUserRepository is a mock implementation of UserRepoInterface for export testing
type mockExportUserRepository struct {
	users map[uuid.UUID]*models.User
}

func newMockExportUserRepository() *mockExportUserRepository {
	return &mockExportUserRepository{
		users: make(map[uuid.UUID]*models.User),
	}
}

func (m *mockExportUserRepository) GetByID(ctx context.Context, userID uuid.UUID) (*models.User, error) {
	user, ok := m.users[userID]
	if !ok {
		return nil, fmt.Errorf("user not found")
	}
	return user, nil
}

func (m *mockExportUserRepository) addTestUser(userID uuid.UUID, email string, displayName string) {
	m.users[userID] = &models.User{
		ID:          userID,
		Email:       &email,
		DisplayName: displayName,
	}
}

func TestExportService_CreateExportRequest(t *testing.T) {
	tmpDir := t.TempDir()
	mockRepo := newMockExportRepository()
	mockUserRepo := newMockExportUserRepository()
	service := NewExportService(mockRepo, mockUserRepo, nil, nil, tmpDir, "http://localhost:8080", 7)

	userID := uuid.New()
	creatorName := "testcreator"

	tests := []struct {
		name    string
		format  string
		wantErr bool
	}{
		{
			name:    "Create CSV export request",
			format:  models.ExportFormatCSV,
			wantErr: false,
		},
		{
			name:    "Create JSON export request",
			format:  models.ExportFormatJSON,
			wantErr: false,
		},
		{
			name:    "Invalid format",
			format:  "xml",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := service.CreateExportRequest(context.Background(), userID, creatorName, tt.format)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.NotNil(t, req)
			assert.Equal(t, userID, req.UserID)
			assert.Equal(t, creatorName, req.CreatorName)
			assert.Equal(t, tt.format, req.Format)
			assert.Equal(t, models.ExportStatusPending, req.Status)
		})
	}
}

func TestExportService_ProcessExportRequest_CSV(t *testing.T) {
	tmpDir := t.TempDir()
	mockRepo := newMockExportRepository()
	service := NewExportService(mockRepo, nil, nil, nil, tmpDir, "http://localhost:8080", 7)

	userID := uuid.New()
	creatorName := "testcreator"

	// Add test clips
	mockRepo.addTestClips(creatorName, 5)

	// Create export request
	req, err := service.CreateExportRequest(context.Background(), userID, creatorName, models.ExportFormatCSV)
	require.NoError(t, err)

	// Process export
	err = service.ProcessExportRequest(context.Background(), req)
	require.NoError(t, err)

	// Verify export was completed
	updated, err := mockRepo.GetExportRequestByID(context.Background(), req.ID)
	require.NoError(t, err)
	assert.Equal(t, models.ExportStatusCompleted, updated.Status)
	assert.NotNil(t, updated.FilePath)
	assert.NotNil(t, updated.FileSizeBytes)
	assert.NotNil(t, updated.ExpiresAt)

	// Verify file exists
	_, err = os.Stat(*updated.FilePath)
	assert.NoError(t, err)

	// Verify file contains data
	assert.Greater(t, *updated.FileSizeBytes, int64(0))

	// Clean up
	os.Remove(*updated.FilePath)
}

func TestExportService_ProcessExportRequest_JSON(t *testing.T) {
	tmpDir := t.TempDir()
	mockRepo := newMockExportRepository()
	service := NewExportService(mockRepo, nil, nil, nil, tmpDir, "http://localhost:8080", 7)

	userID := uuid.New()
	creatorName := "testcreator"

	// Add test clips
	mockRepo.addTestClips(creatorName, 3)

	// Create export request
	req, err := service.CreateExportRequest(context.Background(), userID, creatorName, models.ExportFormatJSON)
	require.NoError(t, err)

	// Process export
	err = service.ProcessExportRequest(context.Background(), req)
	require.NoError(t, err)

	// Verify export was completed
	updated, err := mockRepo.GetExportRequestByID(context.Background(), req.ID)
	require.NoError(t, err)
	assert.Equal(t, models.ExportStatusCompleted, updated.Status)
	assert.NotNil(t, updated.FilePath)
	assert.NotNil(t, updated.FileSizeBytes)

	// Verify file exists
	_, err = os.Stat(*updated.FilePath)
	assert.NoError(t, err)

	// Clean up
	os.Remove(*updated.FilePath)
}

func TestExportService_CleanupExpiredExports(t *testing.T) {
	tmpDir := t.TempDir()
	mockRepo := newMockExportRepository()
	service := NewExportService(mockRepo, nil, nil, nil, tmpDir, "http://localhost:8080", 7)

	userID := uuid.New()
	creatorName := "testcreator"

	// Add test clips
	mockRepo.addTestClips(creatorName, 2)

	// Create and process export
	req, err := service.CreateExportRequest(context.Background(), userID, creatorName, models.ExportFormatCSV)
	require.NoError(t, err)

	err = service.ProcessExportRequest(context.Background(), req)
	require.NoError(t, err)

	// Get the completed request
	completed, err := mockRepo.GetExportRequestByID(context.Background(), req.ID)
	require.NoError(t, err)

	// Manually set expiration to the past
	pastTime := time.Now().Add(-1 * time.Hour)
	completed.ExpiresAt = &pastTime

	// Run cleanup
	err = service.CleanupExpiredExports(context.Background())
	require.NoError(t, err)

	// Verify file was deleted
	if completed.FilePath != nil {
		_, err = os.Stat(*completed.FilePath)
		assert.True(t, os.IsNotExist(err))
	}

	// Verify status was updated
	expired, err := mockRepo.GetExportRequestByID(context.Background(), req.ID)
	require.NoError(t, err)
	assert.Equal(t, models.ExportStatusExpired, expired.Status)
}

func TestExportService_GetExportFilePath(t *testing.T) {
	tmpDir := t.TempDir()
	mockRepo := newMockExportRepository()
	service := NewExportService(mockRepo, nil, nil, nil, tmpDir, "http://localhost:8080", 7)

	userID := uuid.New()
	creatorName := "testcreator"

	// Add test clips
	mockRepo.addTestClips(creatorName, 1)

	// Create and process export
	req, err := service.CreateExportRequest(context.Background(), userID, creatorName, models.ExportFormatCSV)
	require.NoError(t, err)

	err = service.ProcessExportRequest(context.Background(), req)
	require.NoError(t, err)

	// Get file path
	filePath, err := service.GetExportFilePath(context.Background(), req.ID)
	require.NoError(t, err)
	assert.NotEmpty(t, filePath)

	// Verify file exists
	_, err = os.Stat(filePath)
	assert.NoError(t, err)

	// Clean up
	os.Remove(filePath)
}

func TestExportService_CSVGeneration(t *testing.T) {
	tmpDir := t.TempDir()
	mockRepo := newMockExportRepository()
	service := NewExportService(mockRepo, nil, nil, nil, tmpDir, "http://localhost:8080", 7)

	creatorName := "testcreator"
	mockRepo.addTestClips(creatorName, 10)

	clips, err := mockRepo.GetCreatorClipsForExport(context.Background(), creatorName)
	require.NoError(t, err)

	exportID := uuid.New()
	filePath, fileSize, err := service.generateCSVExport(exportID, clips)
	require.NoError(t, err)

	// Verify file was created
	assert.NotEmpty(t, filePath)
	assert.Greater(t, fileSize, int64(0))
	assert.Equal(t, filepath.Ext(filePath), ".csv")

	// Verify file exists
	_, err = os.Stat(filePath)
	assert.NoError(t, err)

	// Clean up
	os.Remove(filePath)
}

func TestExportService_JSONGeneration(t *testing.T) {
	tmpDir := t.TempDir()
	mockRepo := newMockExportRepository()
	service := NewExportService(mockRepo, nil, nil, nil, tmpDir, "http://localhost:8080", 7)

	creatorName := "testcreator"
	mockRepo.addTestClips(creatorName, 10)

	clips, err := mockRepo.GetCreatorClipsForExport(context.Background(), creatorName)
	require.NoError(t, err)

	exportID := uuid.New()
	filePath, fileSize, err := service.generateJSONExport(exportID, clips)
	require.NoError(t, err)

	// Verify file was created
	assert.NotEmpty(t, filePath)
	assert.Greater(t, fileSize, int64(0))
	assert.Equal(t, filepath.Ext(filePath), ".json")

	// Verify file exists
	_, err = os.Stat(filePath)
	assert.NoError(t, err)

	// Clean up
	os.Remove(filePath)
}

func TestExportService_NotificationsSent(t *testing.T) {
	tmpDir := t.TempDir()
	mockRepo := newMockExportRepository()
	mockUserRepo := newMockExportUserRepository()

	// Create test user
	userID := uuid.New()
	userEmail := "test@example.com"
	mockUserRepo.addTestUser(userID, userEmail, "Test User")

	// Create service (without actual email/notification services for this unit test)
	service := NewExportService(mockRepo, mockUserRepo, nil, nil, tmpDir, "http://localhost:8080", 7)

	creatorName := "testcreator"
	mockRepo.addTestClips(creatorName, 5)

	// Create and process export request
	req, err := service.CreateExportRequest(context.Background(), userID, creatorName, models.ExportFormatCSV)
	require.NoError(t, err)

	err = service.ProcessExportRequest(context.Background(), req)
	require.NoError(t, err)

	// Verify export was completed
	updated, err := mockRepo.GetExportRequestByID(context.Background(), req.ID)
	require.NoError(t, err)
	assert.Equal(t, models.ExportStatusCompleted, updated.Status)
	assert.NotNil(t, updated.FilePath)
	assert.NotNil(t, updated.FileSizeBytes)
	assert.NotNil(t, updated.ExpiresAt)

	// Clean up
	if updated.FilePath != nil {
		os.Remove(*updated.FilePath)
	}
}

func TestExportService_FailedExportHandling(t *testing.T) {
	tmpDir := t.TempDir()
	mockRepo := newMockExportRepository()
	mockUserRepo := newMockExportUserRepository()

	// Create test user
	userID := uuid.New()
	mockUserRepo.addTestUser(userID, "test@example.com", "Test User")

	service := NewExportService(mockRepo, mockUserRepo, nil, nil, tmpDir, "http://localhost:8080", 7)

	// Create export request with invalid format to trigger failure
	req, err := service.CreateExportRequest(context.Background(), userID, "testcreator", models.ExportFormatCSV)
	require.NoError(t, err)

	// Manually change format to invalid after creation
	req.Format = "invalid"

	// Process export - should fail
	err = service.ProcessExportRequest(context.Background(), req)
	assert.Error(t, err)

	// Verify export was marked as failed
	updated, err := mockRepo.GetExportRequestByID(context.Background(), req.ID)
	require.NoError(t, err)
	assert.Equal(t, models.ExportStatusFailed, updated.Status)
	assert.NotNil(t, updated.ErrorMessage)
}

func TestFormatFileSize(t *testing.T) {
	tests := []struct {
		name     string
		bytes    int64
		expected string
	}{
		{"Zero bytes", 0, "0 B"},
		{"Bytes", 500, "500 B"},
		{"Kilobytes", 1024, "1.0 KB"},
		{"Kilobytes decimal", 1536, "1.5 KB"},
		{"Megabytes", 1048576, "1.0 MB"},
		{"Megabytes decimal", 1572864, "1.5 MB"},
		{"Gigabytes", 1073741824, "1.0 GB"},
		{"Large file", 5368709120, "5.0 GB"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatFileSize(tt.bytes)
			assert.Equal(t, tt.expected, result)
		})
	}
}
