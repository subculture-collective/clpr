package handlers

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"git.subcult.tv/subculture-collective/clpr/internal/models"
	"git.subcult.tv/subculture-collective/clpr/internal/repository"
	"git.subcult.tv/subculture-collective/clpr/internal/services"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type auditLogServiceStub struct {
	logs      []*models.ModerationAuditLogWithUser
	total     int
	exportCSV string
	err       error
}

func (s *auditLogServiceStub) GetAuditLogs(context.Context, repository.AuditLogFilters, int, int) ([]*models.ModerationAuditLogWithUser, int, error) {
	return s.logs, s.total, s.err
}
func (s *auditLogServiceStub) ExportAuditLogsCSV(_ context.Context, _ repository.AuditLogFilters, writer io.Writer) error {
	if s.err != nil {
		return s.err
	}
	_, err := io.WriteString(writer, s.exportCSV)
	return err
}
func (s *auditLogServiceStub) GetAuditLogByID(context.Context, uuid.UUID) (*models.ModerationAuditLogWithUser, error) {
	if len(s.logs) == 0 {
		return nil, s.err
	}
	return s.logs[0], s.err
}

func TestAdminAuditLogResponseContracts(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logEntry := &models.ModerationAuditLogWithUser{ModerationAuditLog: models.ModerationAuditLog{
		ID: uuid.New(), Action: "ban_user", EntityType: "user", EntityID: uuid.New(), ModeratorID: uuid.New(), CreatedAt: time.Now(),
	}}
	handler := NewAuditLogHandler(&auditLogServiceStub{logs: []*models.ModerationAuditLogWithUser{logEntry}, total: 1, exportCSV: "ID,Action\n1,ban_user\n"})

	t.Run("list", func(t *testing.T) {
		recorder := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(recorder)
		ctx.Request = httptest.NewRequest(http.MethodGet, "/api/v1/admin/audit-logs?page=1&limit=50", nil)
		handler.ListAuditLogs(ctx)
		if recorder.Code != http.StatusOK || !strings.Contains(recorder.Body.String(), `"total":1`) {
			t.Fatalf("unexpected audit list: %d %s", recorder.Code, recorder.Body.String())
		}
	})

	t.Run("buffered export", func(t *testing.T) {
		recorder := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(recorder)
		ctx.Request = httptest.NewRequest(http.MethodGet, "/api/v1/admin/audit-logs/export", nil)
		handler.ExportAuditLogs(ctx)
		if recorder.Code != http.StatusOK || !strings.HasPrefix(recorder.Header().Get("Content-Type"), "text/csv") || !strings.Contains(recorder.Body.String(), "ban_user") {
			t.Fatalf("unexpected audit export: %d %s", recorder.Code, recorder.Body.String())
		}
	})
}

func TestAdminAuditLogExportFailsBeforeCSVHeaders(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := NewAuditLogHandler(&auditLogServiceStub{err: services.ErrAuditLogExportTooLarge})
	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Request = httptest.NewRequest(http.MethodGet, "/api/v1/admin/audit-logs/export", nil)
	handler.ExportAuditLogs(ctx)
	if recorder.Code != http.StatusRequestEntityTooLarge || strings.HasPrefix(recorder.Header().Get("Content-Type"), "text/csv") {
		t.Fatalf("expected JSON 413 before CSV headers, got %d %s", recorder.Code, recorder.Header().Get("Content-Type"))
	}

	handler = NewAuditLogHandler(&auditLogServiceStub{err: errors.New("database unavailable")})
	recorder = httptest.NewRecorder()
	ctx, _ = gin.CreateTestContext(recorder)
	ctx.Request = httptest.NewRequest(http.MethodGet, "/api/v1/admin/audit-logs/export", nil)
	handler.ExportAuditLogs(ctx)
	if recorder.Code != http.StatusInternalServerError || strings.HasPrefix(recorder.Header().Get("Content-Type"), "text/csv") {
		t.Fatalf("expected JSON 500 before CSV headers, got %d %s", recorder.Code, recorder.Header().Get("Content-Type"))
	}
}
