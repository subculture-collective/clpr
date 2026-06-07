package repository

import (
	"context"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5"
)

type fakeSourceIdentityQuerier struct {
	lastQuery string
	lastArgs  []any
	row       pgx.Row
}

func (f *fakeSourceIdentityQuerier) QueryRow(ctx context.Context, sql string, args ...any) pgx.Row {
	f.lastQuery = sql
	f.lastArgs = append([]any(nil), args...)
	return f.row
}

type noRowsRow struct{}

func (noRowsRow) Scan(dest ...any) error { return pgx.ErrNoRows }

func TestGetSubmissionBySourceIdentity_UsesPlatformAndSourceID(t *testing.T) {
	querier := &fakeSourceIdentityQuerier{row: noRowsRow{}}

	submission, err := getSubmissionBySourceIdentity(context.Background(), querier, "youtube", "abc123")
	if err != nil {
		t.Fatalf("getSubmissionBySourceIdentity() error = %v", err)
	}
	if submission != nil {
		t.Fatalf("getSubmissionBySourceIdentity() submission = %#v, want nil", submission)
	}
	if querier.lastArgs == nil || len(querier.lastArgs) != 2 {
		t.Fatalf("expected 2 query args, got %#v", querier.lastArgs)
	}
	if !strings.Contains(querier.lastQuery, "source_platform = $1") || !strings.Contains(querier.lastQuery, "source_id = $2") {
		t.Fatalf("query = %q, want source_platform/source_id lookup", querier.lastQuery)
	}
	if querier.lastArgs[0] != "youtube" || querier.lastArgs[1] != "abc123" {
		t.Fatalf("query args = %#v, want [youtube abc123]", querier.lastArgs)
	}
}
