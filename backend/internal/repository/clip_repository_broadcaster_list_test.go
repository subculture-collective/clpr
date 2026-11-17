package repository

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type broadcasterListDB struct {
	countQuery string
	listQuery  string
}

type broadcasterCountRow struct{ total int }

func (r broadcasterCountRow) Scan(dest ...any) error {
	*(dest[0].(*int)) = r.total
	return nil
}

var errBroadcasterListStop = errors.New("stop after capturing the list query")

func (d *broadcasterListDB) Begin(context.Context) (pgx.Tx, error) {
	panic("Begin is not used by this test double")
}

func (d *broadcasterListDB) Exec(context.Context, string, ...any) (pgconn.CommandTag, error) {
	panic("Exec is not used by this test double")
}

func (d *broadcasterListDB) Query(_ context.Context, query string, _ ...any) (pgx.Rows, error) {
	d.listQuery = query
	return nil, errBroadcasterListStop
}

func (d *broadcasterListDB) QueryRow(_ context.Context, query string, _ ...any) pgx.Row {
	d.countQuery = query
	return broadcasterCountRow{total: 3}
}

// The broadcaster page header counts every visible clip of the channel, so the
// list must not be limited to user-submitted clips and must skip DMCA removals.
func TestClipRepository_ListClipsByBroadcaster_ListsAllVisibleClips(t *testing.T) {
	db := &broadcasterListDB{}
	repo := &ClipRepository{pool: db}

	_, _, err := repo.ListClipsByBroadcaster(context.Background(), "12345", "recent", 20, 0)
	if !errors.Is(err, errBroadcasterListStop) {
		t.Fatalf("ListClipsByBroadcaster() error = %v, want the list query to run", err)
	}

	for name, query := range map[string]string{"count": db.countQuery, "list": db.listQuery} {
		if strings.Contains(query, "submitted_by_user_id IS NOT NULL") {
			t.Errorf("%s query still limits the list to user-submitted clips:\n%s", name, query)
		}
		for _, want := range []string{
			"broadcaster_id = $1",
			"is_removed = false",
			"is_hidden = false",
			"COALESCE(dmca_removed, false) = false",
		} {
			if !strings.Contains(query, want) {
				t.Errorf("%s query is missing %q:\n%s", name, want, query)
			}
		}
	}
}
