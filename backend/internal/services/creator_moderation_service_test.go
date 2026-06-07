package services

import (
	"context"
	"testing"
	"time"

	"git.subcult.tv/subculture-collective/clpr/internal/models"
	"github.com/google/uuid"
)

type creatorModerationRepoStub struct {
	createModeratorFn         func(context.Context, *models.CreatorModerator) error
	listModeratorsFn          func(context.Context, uuid.UUID) ([]*models.CreatorModerator, error)
	createBanFn               func(context.Context, *models.CreatorBan) error
	getActiveBanForUserFn     func(context.Context, uuid.UUID, uuid.UUID, string) (*models.CreatorBan, error)
	getActiveBanForPlatformFn func(context.Context, uuid.UUID, string, string, string) (*models.CreatorBan, error)

	lastModerator *models.CreatorModerator
	lastBan       *models.CreatorBan
	lastScope     string
}

func (r *creatorModerationRepoStub) CreateModerator(ctx context.Context, moderator *models.CreatorModerator) error {
	r.lastModerator = moderator
	if r.createModeratorFn != nil {
		return r.createModeratorFn(ctx, moderator)
	}
	return nil
}

func (r *creatorModerationRepoStub) ListModeratorsByCreator(ctx context.Context, creatorID uuid.UUID) ([]*models.CreatorModerator, error) {
	if r.listModeratorsFn != nil {
		return r.listModeratorsFn(ctx, creatorID)
	}
	return nil, nil
}

func (r *creatorModerationRepoStub) CreateBan(ctx context.Context, ban *models.CreatorBan) error {
	r.lastBan = ban
	if r.createBanFn != nil {
		return r.createBanFn(ctx, ban)
	}
	return nil
}

func (r *creatorModerationRepoStub) GetActiveBanForUser(ctx context.Context, creatorID, targetUserID uuid.UUID, scope string) (*models.CreatorBan, error) {
	r.lastScope = scope
	if r.getActiveBanForUserFn != nil {
		return r.getActiveBanForUserFn(ctx, creatorID, targetUserID, scope)
	}
	return nil, nil
}

func (r *creatorModerationRepoStub) GetActiveBanForPlatformIdentity(ctx context.Context, creatorID uuid.UUID, platform, platformUserID, scope string) (*models.CreatorBan, error) {
	if r.getActiveBanForPlatformFn != nil {
		return r.getActiveBanForPlatformFn(ctx, creatorID, platform, platformUserID, scope)
	}
	return nil, nil
}

func TestCreatorModerationService_CanAccess(t *testing.T) {
	repo := &creatorModerationRepoStub{}
	service := NewCreatorModerationService(repo)
	creatorID := uuid.New()
	userID := uuid.New()

	tests := []struct {
		name    string
		call    func(context.Context, uuid.UUID, uuid.UUID) (bool, string, error)
		scope   string
		blocked bool
	}{
		{name: "interact allowed", call: service.CanInteract, scope: "interact", blocked: false},
		{name: "submit blocked", call: service.CanSubmit, scope: "submit", blocked: true},
		{name: "comment allowed", call: service.CanComment, scope: "comment", blocked: false},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			repo.getActiveBanForUserFn = func(ctx context.Context, cID, uID uuid.UUID, scope string) (*models.CreatorBan, error) {
				if scope != tt.scope {
					t.Fatalf("scope = %q, want %q", scope, tt.scope)
				}
				if tt.blocked {
					return &models.CreatorBan{ID: uuid.New(), CreatorID: cID, TargetUserID: &uID, Scopes: []string{tt.scope}, SyncStatus: "local_only", CreatedAt: time.Now(), UpdatedAt: time.Now()}, nil
				}
				return nil, nil
			}

			allowed, reason, err := tt.call(context.Background(), creatorID, userID)
			if err != nil {
				t.Fatalf("call error = %v", err)
			}
			if tt.blocked {
				if allowed {
					t.Fatal("expected blocked")
				}
				if reason != creatorModerationRestrictionMessage {
					t.Fatalf("reason = %q, want %q", reason, creatorModerationRestrictionMessage)
				}
				return
			}
			if !allowed || reason != "" {
				t.Fatalf("allowed=%v reason=%q, want allowed true with empty reason", allowed, reason)
			}
		})
	}
}

func TestCreatorModerationService_CreateModeratorRejectsInvalidPermission(t *testing.T) {
	repo := &creatorModerationRepoStub{createModeratorFn: func(context.Context, *models.CreatorModerator) error {
		t.Fatal("repository should not be called")
		return nil
	}}
	service := NewCreatorModerationService(repo)
	userID := uuid.New()

	if _, err := service.CreateModerator(context.Background(), uuid.New(), &CreateCreatorModeratorRequest{UserID: &userID, Permissions: []string{"view"}}); err == nil {
		t.Fatal("expected invalid permission error")
	}
}

func TestCreatorModerationService_CreateModeratorRejectsPartialIdentityAfterTrim(t *testing.T) {
	repo := &creatorModerationRepoStub{createModeratorFn: func(context.Context, *models.CreatorModerator) error {
		t.Fatal("repository should not be called")
		return nil
	}}
	service := NewCreatorModerationService(repo)
	userID := uuid.New()
	platformUserID := "abc"
	blank := " "

	for _, tc := range []struct {
		name string
		req  *CreateCreatorModeratorRequest
	}{
		{name: "blank platform", req: &CreateCreatorModeratorRequest{UserID: &userID, Platform: stringPtr(" "), PlatformUserID: &platformUserID}},
		{name: "blank platform user id", req: &CreateCreatorModeratorRequest{UserID: &userID, Platform: stringPtr("twitch"), PlatformUserID: &blank}},
		{name: "platform without user id", req: &CreateCreatorModeratorRequest{Platform: stringPtr("twitch")}},
		{name: "platform user id without platform", req: &CreateCreatorModeratorRequest{PlatformUserID: &platformUserID}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := service.CreateModerator(context.Background(), uuid.New(), tc.req); err == nil {
				t.Fatal("expected identity validation error")
			}
		})
	}
}

func TestCreatorModerationService_CreateModerator_AllowsLocalOnly(t *testing.T) {
	repo := &creatorModerationRepoStub{createModeratorFn: func(ctx context.Context, moderator *models.CreatorModerator) error {
		return nil
	}}
	service := NewCreatorModerationService(repo)
	creatorID := uuid.New()
	userID := uuid.New()

	moderator, err := service.CreateModerator(context.Background(), creatorID, &CreateCreatorModeratorRequest{
		UserID:      &userID,
		Permissions: []string{"manage_creator_clips"},
	})
	if err != nil {
		t.Fatalf("CreateModerator() error = %v", err)
	}
	if moderator == nil {
		t.Fatal("CreateModerator() moderator = nil")
	}
	if moderator.CreatorID != creatorID {
		t.Fatalf("CreatorID = %v, want %v", moderator.CreatorID, creatorID)
	}
	if moderator.UserID == nil || *moderator.UserID != userID {
		t.Fatalf("UserID = %#v, want %v", moderator.UserID, userID)
	}
	if moderator.Platform != nil || moderator.PlatformUserID != nil {
		t.Fatalf("unexpected platform identity: platform=%#v platformUserID=%#v", moderator.Platform, moderator.PlatformUserID)
	}
	if moderator.Source != "manual" {
		t.Fatalf("Source = %q, want %q", moderator.Source, "manual")
	}
}

func TestCreatorModerationService_CreateModerator_AllowsPlatformOnly(t *testing.T) {
	repo := &creatorModerationRepoStub{createModeratorFn: func(ctx context.Context, moderator *models.CreatorModerator) error {
		return nil
	}}
	service := NewCreatorModerationService(repo)
	creatorID := uuid.New()
	platformUserID := "abc"

	moderator, err := service.CreateModerator(context.Background(), creatorID, &CreateCreatorModeratorRequest{
		Platform:       stringPtr("twitch"),
		PlatformUserID: &platformUserID,
		Permissions:    []string{"ban_creator_users"},
	})
	if err != nil {
		t.Fatalf("CreateModerator() error = %v", err)
	}
	if moderator == nil {
		t.Fatal("CreateModerator() moderator = nil")
	}
	if moderator.UserID != nil {
		t.Fatalf("UserID = %#v, want nil", moderator.UserID)
	}
	if moderator.Platform == nil || *moderator.Platform != "twitch" {
		t.Fatalf("Platform = %#v, want %q", moderator.Platform, "twitch")
	}
	if moderator.PlatformUserID == nil || *moderator.PlatformUserID != platformUserID {
		t.Fatalf("PlatformUserID = %#v, want %q", moderator.PlatformUserID, platformUserID)
	}
}

func TestCreatorModerationService_CreateBanRejectsViewScope(t *testing.T) {
	repo := &creatorModerationRepoStub{createBanFn: func(context.Context, *models.CreatorBan) error {
		t.Fatal("repository should not be called")
		return nil
	}}
	service := NewCreatorModerationService(repo)
	userID := uuid.New()

	if _, err := service.CreateBan(context.Background(), uuid.New(), &CreateCreatorBanRequest{TargetUserID: &userID, Scopes: []string{"view"}}); err == nil {
		t.Fatal("expected invalid scope error")
	}
}

func TestCreatorModerationService_CreateBanRejectsPartialIdentityAfterTrim(t *testing.T) {
	repo := &creatorModerationRepoStub{createBanFn: func(context.Context, *models.CreatorBan) error {
		t.Fatal("repository should not be called")
		return nil
	}}
	service := NewCreatorModerationService(repo)
	userID := uuid.New()
	targetPlatformUserID := "abc"
	blank := " "

	for _, tc := range []struct {
		name string
		req  *CreateCreatorBanRequest
	}{
		{name: "blank target platform", req: &CreateCreatorBanRequest{TargetUserID: &userID, TargetPlatform: stringPtr(" "), TargetPlatformUserID: &targetPlatformUserID}},
		{name: "blank target platform user id", req: &CreateCreatorBanRequest{TargetUserID: &userID, TargetPlatform: stringPtr("twitch"), TargetPlatformUserID: &blank}},
		{name: "target platform without user id", req: &CreateCreatorBanRequest{TargetPlatform: stringPtr("twitch")}},
		{name: "target platform user id without platform", req: &CreateCreatorBanRequest{TargetPlatformUserID: &targetPlatformUserID}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := service.CreateBan(context.Background(), uuid.New(), tc.req); err == nil {
				t.Fatal("expected identity validation error")
			}
		})
	}
}

func TestCreatorModerationService_CreateBan_AllowsLocalOnly(t *testing.T) {
	repo := &creatorModerationRepoStub{createBanFn: func(ctx context.Context, ban *models.CreatorBan) error {
		return nil
	}}
	service := NewCreatorModerationService(repo)
	creatorID := uuid.New()
	targetUserID := uuid.New()

	ban, err := service.CreateBan(context.Background(), creatorID, &CreateCreatorBanRequest{
		TargetUserID: &targetUserID,
		Scopes:       []string{"interact"},
	})
	if err != nil {
		t.Fatalf("CreateBan() error = %v", err)
	}
	if ban == nil {
		t.Fatal("CreateBan() ban = nil")
	}
	if ban.CreatorID != creatorID {
		t.Fatalf("CreatorID = %v, want %v", ban.CreatorID, creatorID)
	}
	if ban.TargetUserID == nil || *ban.TargetUserID != targetUserID {
		t.Fatalf("TargetUserID = %#v, want %v", ban.TargetUserID, targetUserID)
	}
	if ban.TargetPlatform != nil || ban.TargetPlatformUserID != nil {
		t.Fatalf("unexpected platform identity: platform=%#v platformUserID=%#v", ban.TargetPlatform, ban.TargetPlatformUserID)
	}
	if ban.SyncStatus != "local_only" {
		t.Fatalf("SyncStatus = %q, want %q", ban.SyncStatus, "local_only")
	}
}

func TestCreatorModerationService_CreateBan_AllowsPlatformOnly(t *testing.T) {
	repo := &creatorModerationRepoStub{createBanFn: func(ctx context.Context, ban *models.CreatorBan) error {
		return nil
	}}
	service := NewCreatorModerationService(repo)
	creatorID := uuid.New()
	targetPlatformUserID := "abc"

	ban, err := service.CreateBan(context.Background(), creatorID, &CreateCreatorBanRequest{
		TargetPlatform:       stringPtr("youtube"),
		TargetPlatformUserID: &targetPlatformUserID,
		Scopes:               []string{"comment"},
	})
	if err != nil {
		t.Fatalf("CreateBan() error = %v", err)
	}
	if ban == nil {
		t.Fatal("CreateBan() ban = nil")
	}
	if ban.TargetUserID != nil {
		t.Fatalf("TargetUserID = %#v, want nil", ban.TargetUserID)
	}
	if ban.TargetPlatform == nil || *ban.TargetPlatform != "youtube" {
		t.Fatalf("TargetPlatform = %#v, want %q", ban.TargetPlatform, "youtube")
	}
	if ban.TargetPlatformUserID == nil || *ban.TargetPlatformUserID != targetPlatformUserID {
		t.Fatalf("TargetPlatformUserID = %#v, want %q", ban.TargetPlatformUserID, targetPlatformUserID)
	}
}
