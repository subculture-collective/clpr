package main

import (
	"context"
	"errors"
	"testing"

	"git.subcult.tv/subculture-collective/clpr/internal/models"
	"github.com/google/uuid"
)

type fakeAuditRepo struct {
	users     []*models.User
	revokeErr error
	revoked   []uuid.UUID
	logs      []*models.VerificationAuditLog
}

func (f *fakeAuditRepo) GetVerifiedUsersForAudit(context.Context, int, int) ([]*models.User, error) {
	return f.users, nil
}

func (f *fakeAuditRepo) RevokeUserVerification(_ context.Context, userID uuid.UUID) error {
	if f.revokeErr != nil {
		return f.revokeErr
	}
	f.revoked = append(f.revoked, userID)
	return nil
}

func (f *fakeAuditRepo) CreateAuditLog(_ context.Context, log *models.VerificationAuditLog) error {
	f.logs = append(f.logs, log)
	return nil
}

func bannedUser() *models.User {
	return &models.User{ID: uuid.New(), Username: "banned", IsBanned: true, TrustScore: 100}
}

func TestRunVerificationAudit_FailedRevocationIsNotLoggedAsRevoked(t *testing.T) {
	repo := &fakeAuditRepo{users: []*models.User{bannedUser()}, revokeErr: errors.New("db down")}

	stats, err := runVerificationAudit(context.Background(), repo, false, 10, 90)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if stats.RevokedAudits != 0 || stats.FlaggedAudits != 1 || stats.ErrorCount != 1 {
		t.Fatalf("stats revoked=%d flagged=%d errors=%d, want 0/1/1",
			stats.RevokedAudits, stats.FlaggedAudits, stats.ErrorCount)
	}
	if len(repo.logs) != 1 {
		t.Fatalf("expected one audit log, got %d", len(repo.logs))
	}
	got := repo.logs[0]
	if got.Status != models.AuditStatusFlagged {
		t.Fatalf("logged status = %q, want %q", got.Status, models.AuditStatusFlagged)
	}
	if got.ActionTaken == nil || *got.ActionTaken != models.AuditActionFurtherReviewRequired {
		t.Fatalf("logged action = %v, want %q", got.ActionTaken, models.AuditActionFurtherReviewRequired)
	}
	if got.Findings["revocation_failed"] != true {
		t.Fatalf("findings must record the failed revocation: %v", got.Findings)
	}
}

func TestRunVerificationAudit_SuccessfulRevocationIsLogged(t *testing.T) {
	user := bannedUser()
	repo := &fakeAuditRepo{users: []*models.User{user}}

	stats, err := runVerificationAudit(context.Background(), repo, false, 10, 90)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if stats.RevokedAudits != 1 || stats.ErrorCount != 0 {
		t.Fatalf("stats revoked=%d errors=%d, want 1/0", stats.RevokedAudits, stats.ErrorCount)
	}
	if len(repo.revoked) != 1 || repo.revoked[0] != user.ID {
		t.Fatalf("revoked = %v, want [%s]", repo.revoked, user.ID)
	}
	if len(repo.logs) != 1 || repo.logs[0].Status != models.AuditStatusRevoked {
		t.Fatalf("expected one revoked audit log, got %+v", repo.logs)
	}
}

func TestRunVerificationAudit_DryRunWritesNothing(t *testing.T) {
	repo := &fakeAuditRepo{users: []*models.User{bannedUser()}}
	stats, err := runVerificationAudit(context.Background(), repo, true, 10, 90)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(repo.revoked) != 0 || len(repo.logs) != 0 || stats.ErrorCount != 0 {
		t.Fatalf("dry run wrote data: revoked=%v logs=%d", repo.revoked, len(repo.logs))
	}
}
