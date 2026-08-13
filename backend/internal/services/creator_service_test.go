package services

import (
	"context"
	"strings"
	"testing"
	"time"

	"git.subcult.tv/subculture-collective/clpr/internal/models"
	"github.com/google/uuid"
)

type creatorRepoStub struct {
	createAccountFn        func(context.Context, *models.CreatorAccount) error
	getAccountByIDFn       func(context.Context, uuid.UUID) (*models.CreatorAccount, error)
	getAccountBySlugFn     func(context.Context, string) (*models.CreatorAccount, error)
	listAccountsByOwnerFn  func(context.Context, uuid.UUID) ([]*models.CreatorAccount, error)
	createPlatformFn       func(context.Context, *models.CreatorPlatformAccount) error
	getPlatformByUniqueFn  func(context.Context, string, string) (*models.CreatorPlatformAccount, error)
	listPlatformsByCreator func(context.Context, uuid.UUID) ([]*models.CreatorPlatformAccount, error)

	lastAccount         *models.CreatorAccount
	lastPlatformAccount *models.CreatorPlatformAccount
}

func (r *creatorRepoStub) CreateCreatorAccount(ctx context.Context, account *models.CreatorAccount) error {
	r.lastAccount = account
	if r.createAccountFn != nil {
		return r.createAccountFn(ctx, account)
	}
	return nil
}

func (r *creatorRepoStub) GetCreatorAccountByID(ctx context.Context, id uuid.UUID) (*models.CreatorAccount, error) {
	if r.getAccountByIDFn != nil {
		return r.getAccountByIDFn(ctx, id)
	}
	return nil, nil
}

func (r *creatorRepoStub) GetCreatorAccountBySlug(ctx context.Context, slug string) (*models.CreatorAccount, error) {
	if r.getAccountBySlugFn != nil {
		return r.getAccountBySlugFn(ctx, slug)
	}
	return nil, nil
}

func (r *creatorRepoStub) ListCreatorAccountsByOwner(ctx context.Context, ownerUserID uuid.UUID) ([]*models.CreatorAccount, error) {
	if r.listAccountsByOwnerFn != nil {
		return r.listAccountsByOwnerFn(ctx, ownerUserID)
	}
	return nil, nil
}

func (r *creatorRepoStub) CreateCreatorPlatformAccount(ctx context.Context, account *models.CreatorPlatformAccount) error {
	r.lastPlatformAccount = account
	if r.createPlatformFn != nil {
		return r.createPlatformFn(ctx, account)
	}
	return nil
}

func (r *creatorRepoStub) GetCreatorPlatformAccountByPlatformAndUserID(ctx context.Context, platform, platformUserID string) (*models.CreatorPlatformAccount, error) {
	if r.getPlatformByUniqueFn != nil {
		return r.getPlatformByUniqueFn(ctx, platform, platformUserID)
	}
	return nil, nil
}

func (r *creatorRepoStub) ListCreatorPlatformAccounts(ctx context.Context, creatorID uuid.UUID) ([]*models.CreatorPlatformAccount, error) {
	if r.listPlatformsByCreator != nil {
		return r.listPlatformsByCreator(ctx, creatorID)
	}
	return nil, nil
}

func TestNormalizeCreatorSlug(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{name: "trim and lowercase", input: "  Creator-One  ", want: "creator-one"},
		{name: "reject spaces", input: "Creator One", wantErr: true},
		{name: "reject punctuation", input: "creator!!", wantErr: true},
		{name: "reject empty", input: "   ", wantErr: true},
		{name: "reject all hyphens", input: "---", wantErr: true},
		{name: "reject too long", input: strings.Repeat("a", 256), wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := normalizeCreatorSlug(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("normalizeCreatorSlug(%q) error = nil, want error", tt.input)
				}
				return
			}
			if err != nil {
				t.Fatalf("normalizeCreatorSlug(%q) error = %v", tt.input, err)
			}
			if got != tt.want {
				t.Fatalf("normalizeCreatorSlug(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestCreateCreatorAccount_NormalizesAndStores(t *testing.T) {
	repo := &creatorRepoStub{}
	service := NewCreatorService(repo)
	ownerID := uuid.New()

	account, err := service.CreateCreatorAccount(context.Background(), ownerID, &CreateCreatorAccountRequest{
		DisplayName: "  My Creator  ",
		Slug:        "  My-Creator  ",
	})
	if err != nil {
		t.Fatalf("CreateCreatorAccount() error = %v", err)
	}
	if account == nil {
		t.Fatal("CreateCreatorAccount() account = nil")
	}
	if account.OwnerUserID != ownerID {
		t.Fatalf("OwnerUserID = %v, want %v", account.OwnerUserID, ownerID)
	}
	if account.DisplayName != "My Creator" {
		t.Fatalf("DisplayName = %q, want %q", account.DisplayName, "My Creator")
	}
	if account.Slug != "my-creator" {
		t.Fatalf("Slug = %q, want %q", account.Slug, "my-creator")
	}
	if repo.lastAccount == nil || repo.lastAccount != account {
		t.Fatal("repository did not receive created account")
	}
}

func TestCreateCreatorAccount_RejectsInvalidDisplayNameAndSlug(t *testing.T) {
	repo := &creatorRepoStub{createAccountFn: func(context.Context, *models.CreatorAccount) error {
		t.Fatal("repository should not be called for invalid input")
		return nil
	}}
	service := NewCreatorService(repo)
	ownerID := uuid.New()

	if _, err := service.CreateCreatorAccount(context.Background(), ownerID, &CreateCreatorAccountRequest{DisplayName: "   ", Slug: "creator-one"}); err == nil {
		t.Fatal("expected display name validation error")
	}
	if _, err := service.CreateCreatorAccount(context.Background(), ownerID, &CreateCreatorAccountRequest{DisplayName: "Creator", Slug: "bad slug"}); err == nil {
		t.Fatal("expected slug validation error")
	}
}

func TestLinkPlatformAccount_DefaultsCapabilities(t *testing.T) {
	repo := &creatorRepoStub{}
	service := NewCreatorService(repo)
	creatorID := uuid.New()
	profileURL := "  https://youtube.com/@creator  "
	createdAt := time.Now()
	refreshToken := "refresh-token"

	account, err := service.LinkPlatformAccount(context.Background(), creatorID, &LinkCreatorPlatformAccountRequest{
		Platform:              "YouTube",
		PlatformUserID:        "  channel-123  ",
		PlatformDisplayName:   "  Creator Channel  ",
		ProfileURL:            &profileURL,
		RefreshTokenEncrypted: &refreshToken,
		TokenExpiresAt:        &createdAt,
	})
	if err != nil {
		t.Fatalf("LinkPlatformAccount() error = %v", err)
	}
	if account == nil {
		t.Fatal("LinkPlatformAccount() account = nil")
	}
	if account.Platform != "youtube" {
		t.Fatalf("Platform = %q, want %q", account.Platform, "youtube")
	}
	if account.PlatformUserID != "channel-123" {
		t.Fatalf("PlatformUserID = %q, want %q", account.PlatformUserID, "channel-123")
	}
	if account.PlatformDisplayName != "Creator Channel" {
		t.Fatalf("PlatformDisplayName = %q, want %q", account.PlatformDisplayName, "Creator Channel")
	}
	if account.ProfileURL == nil || *account.ProfileURL != "https://youtube.com/@creator" {
		t.Fatalf("ProfileURL = %#v, want trimmed URL", account.ProfileURL)
	}
	if !account.CanFetchMetadata || account.CanImportBans || account.CanSyncBansOutbound || account.CanImportModerators || !account.CanVerifyOwnership {
		t.Fatalf("unexpected capability defaults: %+v", account)
	}
	if account.RefreshTokenEncrypted == nil || *account.RefreshTokenEncrypted != refreshToken {
		t.Fatalf("RefreshTokenEncrypted = %#v, want preserved token", account.RefreshTokenEncrypted)
	}
	if repo.lastPlatformAccount == nil || repo.lastPlatformAccount.CreatorID != creatorID {
		t.Fatal("repository did not receive linked platform account")
	}
}

func TestLinkPlatformAccount_RejectsUnsupportedPlatform(t *testing.T) {
	repo := &creatorRepoStub{createPlatformFn: func(context.Context, *models.CreatorPlatformAccount) error {
		t.Fatal("repository should not be called for invalid platform")
		return nil
	}}
	service := NewCreatorService(repo)

	for _, platform := range []string{"x/twitter", "youtube_shorts"} {
		if _, err := service.LinkPlatformAccount(context.Background(), uuid.New(), &LinkCreatorPlatformAccountRequest{
			Platform:            platform,
			PlatformUserID:      "user-1",
			PlatformDisplayName: "User",
		}); err == nil {
			t.Fatalf("expected unsupported platform error for %q", platform)
		}
	}
}
