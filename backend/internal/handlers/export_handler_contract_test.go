package handlers

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"git.subcult.tv/subculture-collective/clpr/internal/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type exportServiceStub struct {
	request  *models.ExportRequest
	requests []*models.ExportRequest
	filePath string
	err      error
}

func (s *exportServiceStub) CreateExportRequest(context.Context, uuid.UUID, string, string) (*models.ExportRequest, error) {
	return s.request, s.err
}
func (s *exportServiceStub) GetExportRequest(context.Context, uuid.UUID) (*models.ExportRequest, error) {
	return s.request, s.err
}
func (s *exportServiceStub) GetExportFilePath(context.Context, uuid.UUID) (string, error) {
	return s.filePath, s.err
}
func (s *exportServiceStub) GetUserExportRequests(context.Context, uuid.UUID) ([]*models.ExportRequest, error) {
	return s.requests, s.err
}

type exportUserRepositoryStub struct {
	user *models.User
	err  error
}

func (s *exportUserRepositoryStub) GetByID(context.Context, uuid.UUID) (*models.User, error) {
	return s.user, s.err
}

func TestExportHandlerResponseContracts(t *testing.T) {
	gin.SetMode(gin.TestMode)
	userID := uuid.New()
	exportID := uuid.New()
	expires := time.Now().Add(time.Hour)
	internalPath := "/private/exports/secret.json"
	req := &models.ExportRequest{
		ID: exportID, UserID: userID, CreatorName: "creator", Format: models.ExportFormatJSON,
		Status: models.ExportStatusCompleted, FilePath: &internalPath, ExpiresAt: &expires,
		CreatedAt: time.Now(), UpdatedAt: time.Now(),
	}
	service := &exportServiceStub{request: req, requests: []*models.ExportRequest{req}}
	handler := NewExportHandler(service, &exportUserRepositoryStub{user: &models.User{ID: userID, Username: "creator"}})

	t.Run("status hides internals and uses public URL", func(t *testing.T) {
		recorder := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(recorder)
		ctx.Params = gin.Params{{Key: "id", Value: exportID.String()}}
		ctx.Set("user_id", userID)
		ctx.Set("base_url", "https://clpr.example")
		ctx.Request = httptest.NewRequest(http.MethodGet, "/api/v1/creators/me/export/status/"+exportID.String(), nil)
		handler.GetExportStatus(ctx)
		if recorder.Code != http.StatusOK || strings.Contains(recorder.Body.String(), "private/exports") || !strings.Contains(recorder.Body.String(), "https://clpr.example/api/v1") {
			t.Fatalf("unexpected status response: %d %s", recorder.Code, recorder.Body.String())
		}
	})

	t.Run("list", func(t *testing.T) {
		recorder := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(recorder)
		ctx.Set("user_id", userID)
		ctx.Request = httptest.NewRequest(http.MethodGet, "/api/v1/creators/me/exports", nil)
		handler.ListExportRequests(ctx)
		if recorder.Code != http.StatusOK || !strings.Contains(recorder.Body.String(), `"count":1`) || strings.Contains(recorder.Body.String(), "private/exports") {
			t.Fatalf("unexpected list response: %d %s", recorder.Code, recorder.Body.String())
		}
	})
}

func TestExportHandlerProtectsExpiredAndForeignArtifacts(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ownerID := uuid.New()
	requesterID := uuid.New()
	exportID := uuid.New()
	expired := time.Now().Add(-time.Minute)
	req := &models.ExportRequest{ID: exportID, UserID: ownerID, CreatorName: "creator", Format: models.ExportFormatJSON, Status: models.ExportStatusCompleted, ExpiresAt: &expired}

	t.Run("expired download", func(t *testing.T) {
		handler := NewExportHandler(&exportServiceStub{request: req}, &exportUserRepositoryStub{})
		recorder := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(recorder)
		ctx.Params = gin.Params{{Key: "id", Value: exportID.String()}}
		ctx.Set("user_id", ownerID)
		ctx.Request = httptest.NewRequest(http.MethodGet, "/api/v1/creators/me/export/download/"+exportID.String(), nil)
		handler.DownloadExport(ctx)
		if recorder.Code != http.StatusGone {
			t.Fatalf("expected 410, got %d", recorder.Code)
		}
	})

	t.Run("foreign request is concealed", func(t *testing.T) {
		handler := NewExportHandler(&exportServiceStub{request: req}, &exportUserRepositoryStub{})
		recorder := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(recorder)
		ctx.Params = gin.Params{{Key: "id", Value: exportID.String()}}
		ctx.Set("user_id", requesterID)
		ctx.Request = httptest.NewRequest(http.MethodGet, "/api/v1/creators/me/export/status/"+exportID.String(), nil)
		handler.GetExportStatus(ctx)
		if recorder.Code != http.StatusNotFound {
			t.Fatalf("expected concealed 404, got %d", recorder.Code)
		}
	})

	t.Run("malformed identity does not panic", func(t *testing.T) {
		handler := NewExportHandler(&exportServiceStub{}, &exportUserRepositoryStub{})
		recorder := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(recorder)
		ctx.Set("user_id", "invalid")
		ctx.Request = httptest.NewRequest(http.MethodGet, "/api/v1/creators/me/exports", nil)
		handler.ListExportRequests(ctx)
		if recorder.Code != http.StatusUnauthorized {
			t.Fatalf("expected 401, got %d", recorder.Code)
		}
	})
}

func TestSanitizeExportFilename(t *testing.T) {
	if got := sanitizeExportFilename("bad\"\r\nname"); got != "bad_name" {
		t.Fatalf("unexpected sanitized filename %q", got)
	}
}
