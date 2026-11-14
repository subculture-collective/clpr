package handlers

import (
	"context"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// fakeQueueTx stands in for a transaction: it serves one moderation_queue row
// per item ID and records every statement that would change data.
type fakeQueueTx struct {
	pgx.Tx
	items map[uuid.UUID]fakeQueueItem
	execs []fakeExec
}

type fakeQueueItem struct {
	contentType string
	contentID   uuid.UUID
	reason      string
	status      string
}

type fakeExec struct {
	sql  string
	args []any
}

type fakeQueueRow struct {
	item  fakeQueueItem
	found bool
}

func (r fakeQueueRow) Scan(dest ...any) error {
	if !r.found {
		return pgx.ErrNoRows
	}
	*dest[0].(*string) = r.item.contentType
	*dest[1].(*uuid.UUID) = r.item.contentID
	*dest[2].(*string) = r.item.reason
	*dest[3].(*string) = r.item.status
	return nil
}

func (tx *fakeQueueTx) QueryRow(_ context.Context, _ string, args ...any) pgx.Row {
	item, found := tx.items[args[0].(uuid.UUID)]
	return fakeQueueRow{item: item, found: found}
}

func (tx *fakeQueueTx) Exec(_ context.Context, sql string, args ...any) (pgconn.CommandTag, error) {
	tx.execs = append(tx.execs, fakeExec{sql: strings.Join(strings.Fields(sql), " "), args: args})
	return pgconn.NewCommandTag("UPDATE 1"), nil
}

// execsMatching returns the recorded statements that start with prefix.
func (tx *fakeQueueTx) execsMatching(prefix string) []fakeExec {
	var matches []fakeExec
	for _, e := range tx.execs {
		if strings.HasPrefix(e.sql, prefix) {
			matches = append(matches, e)
		}
	}
	return matches
}

func newFakeQueueTx(contentType, status string) (*fakeQueueTx, uuid.UUID, uuid.UUID) {
	itemID, contentID := uuid.New(), uuid.New()
	return &fakeQueueTx{items: map[uuid.UUID]fakeQueueItem{
		itemID: {contentType: contentType, contentID: contentID, reason: "toxic", status: status},
	}}, itemID, contentID
}

func TestApplyQueueDecision_RejectSoftRemovesComment(t *testing.T) {
	tx, itemID, commentID := newFakeQueueTx("comment", "pending")
	moderatorID := uuid.New()
	reason := "Harassment"

	outcome, status, err := applyQueueDecision(context.Background(), tx, itemID, moderatorID, "reject", "reject", &reason)
	require.NoError(t, err)
	assert.Equal(t, queueDecisionApplied, outcome)
	assert.Equal(t, "rejected", status)

	removed := tx.execsMatching("UPDATE comments SET is_removed = true, removed_reason = $2 WHERE id = $1")
	require.Len(t, removed, 1)
	assert.Equal(t, []any{commentID, "Harassment"}, removed[0].args)
	assert.Empty(t, tx.execsMatching("UPDATE clips"))
	assert.Empty(t, tx.execsMatching("DELETE"))

	queue := tx.execsMatching("UPDATE moderation_queue")
	require.Len(t, queue, 1)
	assert.Equal(t, []any{"rejected", moderatorID, itemID}, queue[0].args)
	require.Len(t, tx.execsMatching("INSERT INTO moderation_decisions"), 1)

	audit := tx.execsMatching("INSERT INTO moderation_audit_logs")
	require.Len(t, audit, 1)
	assert.Equal(t, "reject", audit[0].args[1])
	assert.Equal(t, itemID, audit[0].args[2])
	assert.Equal(t, moderatorID, audit[0].args[3])
}

func TestApplyQueueDecision_RejectSoftRemovesClipWithDefaultReason(t *testing.T) {
	tx, itemID, clipID := newFakeQueueTx("clip", "pending")

	outcome, _, err := applyQueueDecision(context.Background(), tx, itemID, uuid.New(), "reject", "bulk_reject", nil)
	require.NoError(t, err)
	assert.Equal(t, queueDecisionApplied, outcome)

	removed := tx.execsMatching("UPDATE clips SET is_removed = true, removed_reason = $2 WHERE id = $1")
	require.Len(t, removed, 1)
	assert.Equal(t, []any{clipID, "Content removed: toxic"}, removed[0].args)
	assert.Empty(t, tx.execsMatching("UPDATE comments"))
	assert.Equal(t, "bulk_reject", tx.execsMatching("INSERT INTO moderation_audit_logs")[0].args[1])
}

func TestApplyQueueDecision_ApproveAndEscalateLeaveContentAlone(t *testing.T) {
	for _, action := range []string{"approve", "escalate"} {
		tx, itemID, _ := newFakeQueueTx("comment", "pending")

		outcome, _, err := applyQueueDecision(context.Background(), tx, itemID, uuid.New(), action, action, nil)
		require.NoError(t, err, action)
		assert.Equal(t, queueDecisionApplied, outcome, action)
		assert.Empty(t, tx.execsMatching("UPDATE comments"), action)
		assert.Empty(t, tx.execsMatching("UPDATE clips"), action)
		assert.Len(t, tx.execsMatching("INSERT INTO moderation_audit_logs"), 1, action)
	}
}

func TestApplyQueueDecision_RejectOfUserItemRemovesNothing(t *testing.T) {
	tx, itemID, _ := newFakeQueueTx("user", "pending")

	outcome, _, err := applyQueueDecision(context.Background(), tx, itemID, uuid.New(), "reject", "reject", nil)
	require.NoError(t, err)
	assert.Equal(t, queueDecisionApplied, outcome)
	assert.Empty(t, tx.execsMatching("UPDATE comments"))
	assert.Empty(t, tx.execsMatching("UPDATE clips"))
}

func TestApplyQueueDecision_EscalatedItemsCanBeResolved(t *testing.T) {
	for _, action := range []string{"approve", "reject"} {
		tx, itemID, _ := newFakeQueueTx("comment", "escalated")

		outcome, _, err := applyQueueDecision(context.Background(), tx, itemID, uuid.New(), action, action, nil)
		require.NoError(t, err, action)
		assert.Equal(t, queueDecisionApplied, outcome, action)
	}

	// Escalating an already escalated item changes nothing.
	tx, itemID, _ := newFakeQueueTx("comment", "escalated")
	outcome, status, err := applyQueueDecision(context.Background(), tx, itemID, uuid.New(), "escalate", "bulk_escalate", nil)
	require.NoError(t, err)
	assert.Equal(t, queueDecisionStale, outcome)
	assert.Equal(t, "escalated", status)
	assert.Empty(t, tx.execs)
}

func TestApplyQueueDecision_HandledAndMissingItemsChangeNothing(t *testing.T) {
	for _, handled := range []string{"approved", "rejected"} {
		tx, itemID, _ := newFakeQueueTx("comment", handled)

		outcome, status, err := applyQueueDecision(context.Background(), tx, itemID, uuid.New(), "reject", "reject", nil)
		require.NoError(t, err, handled)
		assert.Equal(t, queueDecisionStale, outcome, handled)
		assert.Equal(t, handled, status, handled)
		assert.Empty(t, tx.execs, handled)
	}

	tx, _, _ := newFakeQueueTx("comment", "pending")
	outcome, _, err := applyQueueDecision(context.Background(), tx, uuid.New(), uuid.New(), "approve", "approve", nil)
	require.NoError(t, err)
	assert.Equal(t, queueDecisionNotFound, outcome)
	assert.Empty(t, tx.execs)
}

// A batch with one stale item still applies the rest; this is the loop BulkModerate runs.
func TestApplyQueueDecision_BatchSkipsStaleItem(t *testing.T) {
	open1, stale, open2 := uuid.New(), uuid.New(), uuid.New()
	tx := &fakeQueueTx{items: map[uuid.UUID]fakeQueueItem{
		open1: {contentType: "comment", contentID: uuid.New(), reason: "toxic", status: "pending"},
		stale: {contentType: "comment", contentID: uuid.New(), reason: "toxic", status: "approved"},
		open2: {contentType: "clip", contentID: uuid.New(), reason: "nsfw", status: "pending"},
	}}

	var processed, skipped []uuid.UUID
	for _, id := range []uuid.UUID{open1, stale, open2} {
		outcome, _, err := applyQueueDecision(context.Background(), tx, id, uuid.New(), "reject", "bulk_reject", nil)
		require.NoError(t, err)
		if outcome == queueDecisionApplied {
			processed = append(processed, id)
		} else {
			skipped = append(skipped, id)
		}
	}

	assert.Equal(t, []uuid.UUID{open1, open2}, processed)
	assert.Equal(t, []uuid.UUID{stale}, skipped)
	assert.Len(t, tx.execsMatching("UPDATE moderation_queue"), 2)
	assert.Len(t, tx.execsMatching("INSERT INTO moderation_audit_logs"), 2)
}
