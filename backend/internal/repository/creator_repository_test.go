package repository

import (
	"context"
	"strings"
	"testing"
	"time"

	"git.subcult.tv/subculture-collective/clpr/internal/models"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type fakeCreatorRow struct {
	scanFn func(dest ...any) error
}

func (r fakeCreatorRow) Scan(dest ...any) error {
	return r.scanFn(dest...)
}

type fakeCreatorDB struct {
	lastQuery string
	lastArgs  []any
	row       pgx.Row
}

func (db *fakeCreatorDB) Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error) {
	return pgconn.CommandTag{}, nil
}

func (db *fakeCreatorDB) Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error) {
	return nil, nil
}

func (db *fakeCreatorDB) QueryRow(ctx context.Context, sql string, args ...any) pgx.Row {
	db.lastQuery = sql
	db.lastArgs = append([]any(nil), args...)
	return db.row
}

func TestCreatorRepository_CreateCreatorAccount(t *testing.T) {
	db := &fakeCreatorDB{row: fakeCreatorRow{scanFn: func(dest ...any) error {
		createdAt := time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)
		updatedAt := time.Date(2026, 1, 2, 3, 5, 6, 0, time.UTC)
		*(dest[0].(*time.Time)) = createdAt
		*(dest[1].(*time.Time)) = updatedAt
		return nil
	}}}
	repo := &CreatorRepository{db: db}
	account := &models.CreatorAccount{ID: uuid.New(), OwnerUserID: uuid.New(), DisplayName: "Creator", Slug: "creator"}

	if err := repo.CreateCreatorAccount(context.Background(), account); err != nil {
		t.Fatalf("CreateCreatorAccount() error = %v", err)
	}
	if !strings.Contains(db.lastQuery, "INSERT INTO creator_accounts") {
		t.Fatalf("query = %q, want creator_accounts insert", db.lastQuery)
	}
	if got, want := len(db.lastArgs), 4; got != want {
		t.Fatalf("arg count = %d, want %d", got, want)
	}
	if db.lastArgs[0] != account.ID || db.lastArgs[1] != account.OwnerUserID || db.lastArgs[2] != account.DisplayName || db.lastArgs[3] != account.Slug {
		t.Fatalf("args = %#v, want ID/owner/display_name/slug order", db.lastArgs)
	}
	if account.CreatedAt.IsZero() || account.UpdatedAt.IsZero() {
		t.Fatal("timestamps were not populated")
	}
}

func TestCreatorRepository_GetCreatorAccountBySlug(t *testing.T) {
	accountID := uuid.New()
	ownerID := uuid.New()
	db := &fakeCreatorDB{row: fakeCreatorRow{scanFn: func(dest ...any) error {
		*(dest[0].(*uuid.UUID)) = accountID
		*(dest[1].(*uuid.UUID)) = ownerID
		*(dest[2].(*string)) = "Creator"
		*(dest[3].(*string)) = "creator"
		*(dest[4].(*time.Time)) = time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)
		*(dest[5].(*time.Time)) = time.Date(2026, 1, 2, 3, 5, 6, 0, time.UTC)
		return nil
	}}}
	repo := &CreatorRepository{db: db}

	account, err := repo.GetCreatorAccountBySlug(context.Background(), "creator")
	if err != nil {
		t.Fatalf("GetCreatorAccountBySlug() error = %v", err)
	}
	if account == nil || account.ID != accountID || account.OwnerUserID != ownerID {
		t.Fatalf("unexpected account result: %#v", account)
	}
	if !strings.Contains(db.lastQuery, "WHERE slug = $1") {
		t.Fatalf("query = %q, want slug lookup", db.lastQuery)
	}
	if len(db.lastArgs) != 1 || db.lastArgs[0] != "creator" {
		t.Fatalf("args = %#v, want [creator]", db.lastArgs)
	}
}

func TestCreatorRepository_CreateCreatorPlatformAccount(t *testing.T) {
	creatorID := uuid.New()
	db := &fakeCreatorDB{row: fakeCreatorRow{scanFn: func(dest ...any) error {
		createdAt := time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)
		updatedAt := time.Date(2026, 1, 2, 3, 5, 6, 0, time.UTC)
		*(dest[0].(*time.Time)) = createdAt
		*(dest[1].(*time.Time)) = updatedAt
		return nil
	}}}
	repo := &CreatorRepository{db: db}
	profileURL := "https://youtube.com/@creator"
	now := time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)
	account := &models.CreatorPlatformAccount{
		ID:                    uuid.New(),
		CreatorID:             creatorID,
		Platform:              "youtube",
		PlatformUserID:        "channel-123",
		PlatformDisplayName:   "Creator Channel",
		ProfileURL:            &profileURL,
		CanFetchMetadata:      true,
		AccessTokenEncrypted:  strPtr("access"),
		RefreshTokenEncrypted: strPtr("refresh"),
		TokenExpiresAt:        &now,
	}

	if err := repo.CreateCreatorPlatformAccount(context.Background(), account); err != nil {
		t.Fatalf("CreateCreatorPlatformAccount() error = %v", err)
	}
	if !strings.Contains(db.lastQuery, "INSERT INTO creator_platform_accounts") {
		t.Fatalf("query = %q, want creator_platform_accounts insert", db.lastQuery)
	}
	if got, want := len(db.lastArgs), 14; got != want {
		t.Fatalf("arg count = %d, want %d", got, want)
	}
	if db.lastArgs[1] != creatorID || db.lastArgs[2] != account.Platform || db.lastArgs[3] != account.PlatformUserID {
		t.Fatalf("args = %#v, want creator/platform/platform_user_id order", db.lastArgs)
	}
}

func strPtr(s string) *string { return &s }
