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

type fakeCreatorModerationRow struct {
	scanFn func(dest ...any) error
}

func (r fakeCreatorModerationRow) Scan(dest ...any) error { return r.scanFn(dest...) }

type fakeCreatorModerationDB struct {
	lastQuery string
	lastArgs  []any
	row       pgx.Row
}

func (db *fakeCreatorModerationDB) Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error) {
	return pgconn.CommandTag{}, nil
}

func (db *fakeCreatorModerationDB) Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error) {
	db.lastQuery = sql
	db.lastArgs = append([]any(nil), args...)
	return nil, nil
}

func (db *fakeCreatorModerationDB) QueryRow(ctx context.Context, sql string, args ...any) pgx.Row {
	db.lastQuery = sql
	db.lastArgs = append([]any(nil), args...)
	return db.row
}

func TestCreatorModerationRepository_GetActiveBanForUser_Query(t *testing.T) {
	banID := uuid.New()
	creatorID := uuid.New()
	targetUserID := uuid.New()
	db := &fakeCreatorModerationDB{row: fakeCreatorModerationRow{scanFn: func(dest ...any) error {
		*(dest[0].(*uuid.UUID)) = banID
		*(dest[1].(*uuid.UUID)) = creatorID
		*(dest[2].(**uuid.UUID)) = &targetUserID
		*(dest[3].(**string)) = nil
		*(dest[4].(**string)) = nil
		*(dest[5].(*string)) = "manual"
		*(dest[6].(**string)) = nil
		*(dest[7].(*[]string)) = []string{"interact"}
		*(dest[8].(**time.Time)) = nil
		*(dest[9].(**uuid.UUID)) = nil
		*(dest[10].(*string)) = "local_only"
		*(dest[11].(*time.Time)) = time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)
		*(dest[12].(*time.Time)) = time.Date(2026, 1, 2, 3, 5, 6, 0, time.UTC)
		return nil
	}}}
	repo := &CreatorModerationRepository{db: db}

	ban, err := repo.GetActiveBanForUser(context.Background(), creatorID, targetUserID, "interact")
	if err != nil {
		t.Fatalf("GetActiveBanForUser() error = %v", err)
	}
	if ban == nil || ban.ID != banID {
		t.Fatalf("unexpected ban result: %#v", ban)
	}
	if !strings.Contains(db.lastQuery, "creator_id = $1") || !strings.Contains(db.lastQuery, "target_user_id = $2") {
		t.Fatalf("query = %q, want creator_id and target_user_id filters", db.lastQuery)
	}
	if !strings.Contains(db.lastQuery, "scopes @> ARRAY[$3]::TEXT[]") {
		t.Fatalf("query = %q, want scope overlap filter", db.lastQuery)
	}
	if !strings.Contains(db.lastQuery, "expires_at IS NULL OR expires_at > NOW()") {
		t.Fatalf("query = %q, want active expiry filter", db.lastQuery)
	}
	if got, want := len(db.lastArgs), 3; got != want {
		t.Fatalf("args = %#v, want %d args", db.lastArgs, want)
	}
}

func TestCreatorModerationRepository_CreateModerator_UsesInsert(t *testing.T) {
	db := &fakeCreatorModerationDB{row: fakeCreatorModerationRow{scanFn: func(dest ...any) error {
		*(dest[0].(*time.Time)) = time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)
		return nil
	}}}
	repo := &CreatorModerationRepository{db: db}
	creatorID := uuid.New()
	userID := uuid.New()
	moderator := &models.CreatorModerator{ID: uuid.New(), CreatorID: creatorID, UserID: &userID, Permissions: []string{"manage_creator_clips"}, Source: "manual"}

	if err := repo.CreateModerator(context.Background(), moderator); err != nil {
		t.Fatalf("CreateModerator() error = %v", err)
	}
	if !strings.Contains(db.lastQuery, "INSERT INTO creator_moderators") {
		t.Fatalf("query = %q, want insert", db.lastQuery)
	}
}
