//go:build integration

package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"git.subcult.tv/subculture-collective/clpr/internal/models"
	"git.subcult.tv/subculture-collective/clpr/internal/repository"
	"git.subcult.tv/subculture-collective/clpr/internal/services"
	"git.subcult.tv/subculture-collective/clpr/internal/testutil"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// modQueueFixture seeds users, clips, comments and queue items for the
// moderation queue handlers and removes them when the test ends.
type modQueueFixture struct {
	t         *testing.T
	pool      *pgxpool.Pool
	handler   *ModerationHandler
	moderator uuid.UUID
	author    uuid.UUID
}

func newModQueueFixture(t *testing.T) *modQueueFixture {
	t.Helper()
	gin.SetMode(gin.TestMode)
	pool := testutil.SetupTestDB(t)
	f := &modQueueFixture{
		t:       t,
		pool:    pool,
		handler: NewModerationHandler(nil, nil, nil, nil, repository.NewCommunityRepository(pool), repository.NewAuditLogRepository(pool), pool),
	}
	f.moderator = f.user("mq-moderator", models.RoleModerator)
	f.author = f.user("mq-author", models.RoleUser)
	return f
}

func (f *modQueueFixture) exec(sql string, args ...any) {
	f.t.Helper()
	if _, err := f.pool.Exec(context.Background(), sql, args...); err != nil {
		f.t.Fatalf("exec %q: %v", sql, err)
	}
}

func (f *modQueueFixture) cleanup(sql string, args ...any) {
	f.t.Cleanup(func() { _, _ = f.pool.Exec(context.Background(), sql, args...) })
}

// user inserts a user whose moderator_scope and moderation_channels are NULL.
func (f *modQueueFixture) user(name, role string) uuid.UUID {
	f.t.Helper()
	id := uuid.New()
	f.exec(`INSERT INTO users (id, twitch_id, username, display_name, role, moderator_scope, moderation_channels)
		VALUES ($1, $2, $3, $3, $4, NULL, NULL)`, id, uuid.NewString(), name+"-"+id.String()[:8], role)
	f.cleanup(`DELETE FROM users WHERE id = $1`, id)
	// The audit log's actor reference blocks deleting the user, so it goes first.
	f.cleanup(`DELETE FROM moderation_audit_logs WHERE moderator_id = $1`, id)
	return id
}

func (f *modQueueFixture) username(id uuid.UUID) string {
	f.t.Helper()
	var name string
	if err := f.pool.QueryRow(context.Background(), `SELECT username FROM users WHERE id = $1`, id).Scan(&name); err != nil {
		f.t.Fatalf("read username: %v", err)
	}
	return name
}

func (f *modQueueFixture) clip(title string) uuid.UUID {
	f.t.Helper()
	id, twitchID := uuid.New(), uuid.NewString()
	f.exec(`INSERT INTO clips (id, twitch_clip_id, twitch_clip_url, embed_url, title, creator_name, broadcaster_name, created_at)
		VALUES ($1, $2, $3, $4, $5, 'mq-creator', 'mq-broadcaster', NOW())`,
		id, twitchID, "https://clips.twitch.tv/"+twitchID, "https://clips.twitch.tv/embed?clip="+twitchID, title)
	f.cleanup(`DELETE FROM clips WHERE id = $1`, id)
	return id
}

func (f *modQueueFixture) comment(clipID uuid.UUID, content string) uuid.UUID {
	f.t.Helper()
	id := uuid.New()
	f.exec(`INSERT INTO comments (id, clip_id, user_id, content) VALUES ($1, $2, $3, $4)`, id, clipID, f.author, content)
	return id
}

func (f *modQueueFixture) queueItem(contentType string, contentID uuid.UUID, status string) uuid.UUID {
	f.t.Helper()
	id := uuid.New()
	f.exec(`INSERT INTO moderation_queue (id, content_type, content_id, reason, status, report_count, created_at, reviewed_by, reviewed_at)
		VALUES ($1, $2, $3, 'spam', $4::varchar, 1, NOW() - interval '3 hours',
			CASE WHEN $4 <> 'pending' THEN $5::uuid END, CASE WHEN $4 <> 'pending' THEN NOW() END)`,
		id, contentType, contentID, status, f.moderator)
	f.cleanup(`DELETE FROM moderation_queue WHERE id = $1`, id)
	return id
}

func (f *modQueueFixture) do(method, route, path, body string, handler gin.HandlerFunc) *httptest.ResponseRecorder {
	f.t.Helper()
	router := gin.New()
	router.Handle(method, route, func(c *gin.Context) { c.Set("user_id", f.moderator) }, handler)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(method, path, bytes.NewBufferString(body))
	if body != "" {
		req.Header.Set("Content-Type", "application/json")
	}
	router.ServeHTTP(rec, req)
	return rec
}

func (f *modQueueFixture) reject(itemID uuid.UUID, body string) *httptest.ResponseRecorder {
	return f.do(http.MethodPost, "/admin/moderation/:id/reject", "/admin/moderation/"+itemID.String()+"/reject", body, f.handler.RejectContent)
}

func (f *modQueueFixture) removal(table string, id uuid.UUID) (bool, string) {
	f.t.Helper()
	var removed bool
	var reason *string
	if err := f.pool.QueryRow(context.Background(),
		`SELECT COALESCE(is_removed, false), removed_reason FROM `+table+` WHERE id = $1`, id).Scan(&removed, &reason); err != nil {
		f.t.Fatalf("read %s removal: %v", table, err)
	}
	if reason == nil {
		return removed, ""
	}
	return removed, *reason
}

func (f *modQueueFixture) itemStatus(itemID uuid.UUID) string {
	f.t.Helper()
	var status string
	var reviewed bool
	if err := f.pool.QueryRow(context.Background(),
		`SELECT status, reviewed_at IS NOT NULL AND reviewed_by = $2 FROM moderation_queue WHERE id = $1`,
		itemID, f.moderator).Scan(&status, &reviewed); err != nil {
		f.t.Fatalf("read queue item: %v", err)
	}
	if status != "pending" && !reviewed {
		f.t.Fatalf("item %s is %s without reviewed_at and reviewed_by", itemID, status)
	}
	return status
}

func (f *modQueueFixture) count(sql string, args ...any) int {
	f.t.Helper()
	var n int
	if err := f.pool.QueryRow(context.Background(), sql, args...).Scan(&n); err != nil {
		f.t.Fatalf("count %q: %v", sql, err)
	}
	return n
}

func (f *modQueueFixture) auditRows(action string, itemID uuid.UUID) int {
	return f.count(`SELECT COUNT(*) FROM moderation_audit_logs
		WHERE action = $1 AND entity_type = 'moderation_queue' AND entity_id = $2 AND moderator_id = $3 AND actor_id = $3`,
		action, itemID, f.moderator)
}

func TestModerationQueueListsContentPreview(t *testing.T) {
	f := newModQueueFixture(t)
	clipID := f.clip("Flagged clip title")
	commentID := f.comment(clipID, "flagged comment text")
	commentItem := f.queueItem("comment", commentID, "pending")
	clipItem := f.queueItem("clip", clipID, "pending")
	missingItem := f.queueItem("comment", uuid.New(), "pending")

	rec := f.do(http.MethodGet, "/admin/moderation/queue", "/admin/moderation/queue?limit=100", "", f.handler.GetModerationQueue)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s; want 200", rec.Code, rec.Body.String())
	}
	var resp struct {
		Data []struct {
			ID             uuid.UUID `json:"id"`
			ContentExcerpt *string   `json:"content_excerpt"`
			ContentAuthor  *string   `json:"content_author"`
			ContentURL     *string   `json:"content_url"`
		} `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	want := map[uuid.UUID][3]string{
		commentItem: {"flagged comment text", f.username(f.author), "/clip/" + clipID.String() + "#comment-" + commentID.String()},
		clipItem:    {"Flagged clip title", "mq-creator", "/clip/" + clipID.String()},
		missingItem: {},
	}
	str := func(s *string) string {
		if s == nil {
			return ""
		}
		return *s
	}
	found := 0
	for _, item := range resp.Data {
		expected, ok := want[item.ID]
		if !ok {
			continue
		}
		found++
		got := [3]string{str(item.ContentExcerpt), str(item.ContentAuthor), str(item.ContentURL)}
		if got != expected {
			t.Errorf("item %s preview = %q, want %q", item.ID, got, expected)
		}
	}
	if found != len(want) {
		t.Fatalf("found %d of %d seeded items in %s", found, len(want), rec.Body.String())
	}
}

func TestModerationStatsCountsPendingItems(t *testing.T) {
	f := newModQueueFixture(t)
	f.queueItem("clip", f.clip("Stats clip"), "pending")

	rec := f.do(http.MethodGet, "/admin/moderation/queue/stats", "/admin/moderation/queue/stats", "", f.handler.GetModerationStats)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s; want 200", rec.Code, rec.Body.String())
	}
	var resp struct {
		Data models.ModerationQueueStats `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode %s: %v", rec.Body.String(), err)
	}
	stats := resp.Data
	if stats.TotalPending < 1 || stats.UserReportedCount < 1 || stats.ByContentType["clip"] < 1 || stats.ByReason["spam"] < 1 {
		t.Fatalf("stats = %+v, want non-zero pending, user-reported, clip and spam counts", stats)
	}
	if stats.OldestPendingAge == nil || *stats.OldestPendingAge < 3 {
		t.Fatalf("oldest pending age = %v, want at least 3 hours", stats.OldestPendingAge)
	}
}

func TestRejectContentRemovesFlaggedContent(t *testing.T) {
	f := newModQueueFixture(t)
	clipID := f.clip("Rejected clip")
	commentID := f.comment(f.clip("Clip with a rejected comment"), "rejected comment")
	commentItem := f.queueItem("comment", commentID, "pending")
	clipItem := f.queueItem("clip", clipID, "pending")

	rec := f.reject(commentItem, `{"reason":"breaks the rules"}`)
	if rec.Code != http.StatusOK {
		t.Fatalf("reject comment status = %d, body = %s; want 200", rec.Code, rec.Body.String())
	}
	if removed, reason := f.removal("comments", commentID); !removed || reason != "breaks the rules" {
		t.Fatalf("comment removed=%v reason=%q, want removed with the moderator's reason", removed, reason)
	}
	if status := f.itemStatus(commentItem); status != "rejected" {
		t.Fatalf("comment item status = %q, want rejected", status)
	}
	if n := f.count(`SELECT COUNT(*) FROM moderation_decisions WHERE queue_item_id = $1 AND action = 'reject'`, commentItem); n != 1 {
		t.Fatalf("decision rows = %d, want 1", n)
	}
	if n := f.auditRows("reject", commentItem); n != 1 {
		t.Fatalf("audit rows = %d, want 1", n)
	}

	rec = f.reject(commentItem, "")
	var conflict struct {
		Code   string `json:"code"`
		Status string `json:"status"`
	}
	_ = json.Unmarshal(rec.Body.Bytes(), &conflict)
	if rec.Code != http.StatusConflict || conflict.Code != "ALREADY_HANDLED" || conflict.Status != "rejected" {
		t.Fatalf("second reject status = %d, body = %s; want 409 ALREADY_HANDLED", rec.Code, rec.Body.String())
	}
	if n := f.auditRows("reject", commentItem); n != 1 {
		t.Fatalf("audit rows after second reject = %d, want 1", n)
	}

	rec = f.reject(clipItem, "")
	if rec.Code != http.StatusOK {
		t.Fatalf("reject clip status = %d, body = %s; want 200", rec.Code, rec.Body.String())
	}
	if removed, reason := f.removal("clips", clipID); !removed || reason != "Content removed: spam" {
		t.Fatalf("clip removed=%v reason=%q, want removed with the default reason", removed, reason)
	}
	if n := f.auditRows("reject", clipItem); n != 1 {
		t.Fatalf("clip audit rows = %d, want 1", n)
	}
}

func TestBulkModerateSkipsHandledItemsAndResolvesEscalated(t *testing.T) {
	f := newModQueueFixture(t)
	clipID := f.clip("Bulk rejected clip")
	commentID := f.comment(f.clip("Clip with a bulk rejected comment"), "bulk rejected comment")
	keptClipID := f.clip("Already approved clip")
	commentItem := f.queueItem("comment", commentID, "pending")
	clipItem := f.queueItem("clip", clipID, "pending")
	handledItem := f.queueItem("clip", keptClipID, "approved")

	body, _ := json.Marshal(models.BulkModerationRequest{
		ItemIDs: []string{handledItem.String(), clipItem.String(), commentItem.String()},
		Action:  "reject",
	})
	rec := f.do(http.MethodPost, "/admin/moderation/bulk", "/admin/moderation/bulk", string(body), f.handler.BulkModerate)
	if rec.Code != http.StatusOK {
		t.Fatalf("bulk status = %d, body = %s; want 200", rec.Code, rec.Body.String())
	}
	var resp struct {
		Processed    int         `json:"processed"`
		ProcessedIDs []uuid.UUID `json:"processed_ids"`
		Skipped      []uuid.UUID `json:"skipped"`
		Total        int         `json:"total"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode %s: %v", rec.Body.String(), err)
	}
	processed := map[uuid.UUID]bool{}
	for _, id := range resp.ProcessedIDs {
		processed[id] = true
	}
	if resp.Processed != 2 || resp.Total != 3 || len(processed) != 2 || !processed[clipItem] || !processed[commentItem] ||
		len(resp.Skipped) != 1 || resp.Skipped[0] != handledItem {
		t.Fatalf("bulk response = %s, want two processed and the approved item skipped", rec.Body.String())
	}
	if removed, _ := f.removal("clips", clipID); !removed {
		t.Fatal("bulk reject did not remove the clip")
	}
	if removed, _ := f.removal("comments", commentID); !removed {
		t.Fatal("bulk reject did not remove the comment")
	}
	if removed, _ := f.removal("clips", keptClipID); removed {
		t.Fatal("bulk reject removed the content of a skipped item")
	}
	if status := f.itemStatus(handledItem); status != "approved" {
		t.Fatalf("skipped item status = %q, want approved", status)
	}
	if a, b, c := f.auditRows("bulk_reject", clipItem), f.auditRows("bulk_reject", commentItem), f.auditRows("bulk_reject", handledItem); a != 1 || b != 1 || c != 0 {
		t.Fatalf("bulk audit rows = %d, %d, %d; want 1, 1, 0", a, b, c)
	}

	escalatedClip := f.clip("Escalated clip")
	escalatedItem := f.queueItem("clip", escalatedClip, "pending")
	body, _ = json.Marshal(models.BulkModerationRequest{ItemIDs: []string{escalatedItem.String()}, Action: "escalate"})
	rec = f.do(http.MethodPost, "/admin/moderation/bulk", "/admin/moderation/bulk", string(body), f.handler.BulkModerate)
	if rec.Code != http.StatusOK || f.itemStatus(escalatedItem) != "escalated" {
		t.Fatalf("escalate status = %d, body = %s; want 200 and an escalated item", rec.Code, rec.Body.String())
	}
	rec = f.do(http.MethodPost, "/admin/moderation/:id/approve", "/admin/moderation/"+escalatedItem.String()+"/approve", "", f.handler.ApproveContent)
	if rec.Code != http.StatusOK {
		t.Fatalf("approve escalated status = %d, body = %s; want 200", rec.Code, rec.Body.String())
	}
	if status := f.itemStatus(escalatedItem); status != "approved" {
		t.Fatalf("escalated item status = %q, want approved", status)
	}
	if removed, _ := f.removal("clips", escalatedClip); removed {
		t.Fatal("approving an escalated item removed its clip")
	}
}

// A user whose moderator_scope and moderation_channels are NULL must get a
// permission answer from the moderator-management lookups, not a scan error.
func TestModeratorLookupsAcceptNullScopeColumns(t *testing.T) {
	f := newModQueueFixture(t)
	ctx := context.Background()
	owner := f.user("mq-owner", models.RoleUser)
	channel := uuid.New()
	slug := "mq-channel-" + channel.String()
	f.exec(`INSERT INTO communities (id, name, slug, owner_id) VALUES ($1, $2, $2, $3)`, channel, slug, owner)

	user, err := f.handler.getUserDetails(ctx, f.author)
	if err != nil {
		t.Fatalf("getUserDetails: %v", err)
	}
	if user.ID != f.author || user.Role != models.RoleUser || user.ModeratorScope != "" {
		t.Fatalf("user details = %+v", user)
	}
	if err := f.handler.validateModeratorScope(ctx, channel, f.author); !errors.Is(err, services.ErrModerationPermissionDenied) {
		t.Fatalf("validateModeratorScope for an outsider = %v, want permission denied", err)
	}
	if err := f.handler.validateModeratorScope(ctx, channel, owner); err != nil {
		t.Fatalf("validateModeratorScope for the owner = %v, want nil", err)
	}

	list := func(actor uuid.UUID) *httptest.ResponseRecorder {
		router := gin.New()
		router.GET("/moderation/moderators", func(c *gin.Context) { c.Set("user_id", actor) }, f.handler.ListModerators)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/moderation/moderators?channelId="+channel.String(), nil))
		return rec
	}
	if rec := list(f.author); rec.Code != http.StatusForbidden {
		t.Fatalf("outsider list status = %d, body = %s; want 403", rec.Code, rec.Body.String())
	}
	if rec := list(owner); rec.Code != http.StatusOK {
		t.Fatalf("owner list status = %d, body = %s; want 200", rec.Code, rec.Body.String())
	}
}
