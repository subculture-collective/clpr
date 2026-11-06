//go:build integration

package handlers

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"git.subcult.tv/subculture-collective/clpr/internal/models"
	"git.subcult.tv/subculture-collective/clpr/internal/repository"
	"git.subcult.tv/subculture-collective/clpr/internal/testutil"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// A report's ban_user action must follow the rules of POST /admin/users/:id/ban:
// only administrators may ban, administrators cannot be banned, and the ban is
// recorded in the moderation audit log against the acting staff member.
func TestReportBanUsesGuardedAdminMutation(t *testing.T) {
	pool := testutil.SetupTestDB(t)
	ctx := context.Background()
	gin.SetMode(gin.TestMode)

	suffix := uuid.NewString()[:8]
	createUser := func(name, role string) uuid.UUID {
		t.Helper()
		id := uuid.New()
		if _, err := pool.Exec(ctx, `
			INSERT INTO users (id, twitch_id, username, display_name, role)
			VALUES ($1, $2, $3, $3, $4)`, id, uuid.NewString(), name+"-"+suffix, role); err != nil {
			t.Fatalf("insert %s: %v", name, err)
		}
		t.Cleanup(func() { _, _ = pool.Exec(context.Background(), `DELETE FROM users WHERE id = $1`, id) })
		return id
	}
	admin := createUser("rb-admin", models.RoleAdmin)
	moderator := createUser("rb-moderator", models.RoleModerator)
	adminTarget := createUser("rb-admin-target", models.RoleAdmin)
	userTarget := createUser("rb-user-target", models.RoleUser)
	reporter := createUser("rb-reporter", models.RoleUser)

	reportRepo := repository.NewReportRepository(pool)
	createReport := func(target uuid.UUID) uuid.UUID {
		t.Helper()
		report := &models.Report{
			ID: uuid.New(), ReporterID: reporter, ReportableType: "user", ReportableID: target,
			Reason: "harassment", Status: "pending", CreatedAt: time.Now(),
		}
		if err := reportRepo.CreateReport(ctx, report); err != nil {
			t.Fatalf("create report: %v", err)
		}
		t.Cleanup(func() { _, _ = pool.Exec(context.Background(), `DELETE FROM reports WHERE id = $1`, report.ID) })
		return report.ID
	}
	userRepo := repository.NewUserRepository(pool)
	handler := NewReportHandler(reportRepo, nil, nil, userRepo, nil)
	update := func(actor uuid.UUID, role string, reportID uuid.UUID) *httptest.ResponseRecorder {
		router := gin.New()
		router.PUT("/admin/reports/:id", func(c *gin.Context) {
			c.Set("user_id", actor)
			c.Set("user_role", role)
		}, handler.UpdateReport)
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPut, "/admin/reports/"+reportID.String(),
			bytes.NewBufferString(`{"status":"actioned","action":"ban_user"}`))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(rec, req)
		return rec
	}
	isBanned := func(id uuid.UUID) bool {
		t.Helper()
		var banned bool
		if err := pool.QueryRow(ctx, `SELECT is_banned FROM users WHERE id = $1`, id).Scan(&banned); err != nil {
			t.Fatalf("read is_banned: %v", err)
		}
		return banned
	}

	t.Run("moderator cannot ban an administrator", func(t *testing.T) {
		rec := update(moderator, models.RoleModerator, createReport(adminTarget))
		if rec.Code != http.StatusForbidden {
			t.Fatalf("status = %d, body = %s; want 403", rec.Code, rec.Body.String())
		}
		if isBanned(adminTarget) {
			t.Fatal("administrator was banned through a moderator's report action")
		}
	})

	t.Run("moderator cannot ban a user", func(t *testing.T) {
		rec := update(moderator, models.RoleModerator, createReport(userTarget))
		if rec.Code != http.StatusForbidden {
			t.Fatalf("status = %d, body = %s; want 403", rec.Code, rec.Body.String())
		}
		if isBanned(userTarget) {
			t.Fatal("user was banned by a moderator")
		}
	})

	t.Run("administrator ban is applied and audited", func(t *testing.T) {
		reportID := createReport(userTarget)
		rec := update(admin, models.RoleAdmin, reportID)
		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, body = %s; want 200", rec.Code, rec.Body.String())
		}
		if !isBanned(userTarget) {
			t.Fatal("user was not banned")
		}
		var audited int
		if err := pool.QueryRow(ctx, `
			SELECT COUNT(*) FROM moderation_audit_logs
			WHERE action = $1 AND target_user_id = $2 AND moderator_id = $3`,
			repository.AdminUserActionBan, userTarget, admin).Scan(&audited); err != nil {
			t.Fatalf("read audit log: %v", err)
		}
		if audited != 1 {
			t.Fatalf("audit rows = %d, want 1", audited)
		}
		t.Cleanup(func() {
			_, _ = pool.Exec(context.Background(), `DELETE FROM moderation_audit_logs WHERE target_user_id = $1`, userTarget)
		})
		var status string
		if err := pool.QueryRow(ctx, `SELECT status FROM reports WHERE id = $1`, reportID).Scan(&status); err != nil {
			t.Fatalf("read report: %v", err)
		}
		if status != "actioned" {
			t.Fatalf("report status = %q, want actioned", status)
		}
	})
}
