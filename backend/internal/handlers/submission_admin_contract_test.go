package handlers

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func TestPendingSubmissionsRejectMalformedFilters(t *testing.T) {
	gin.SetMode(gin.TestMode)
	paths := []string{
		"?page=zero", "?page=0", "?limit=101", "?is_nsfw=yes",
		"?broadcaster=" + strings.Repeat("b", 101),
		"?creator=" + strings.Repeat("c", 101),
		"?tags=one,one",
		"?start_date=2026-07-10T00:00:00Z&end_date=2026-07-01T00:00:00Z",
	}
	for _, query := range paths {
		recorder := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(recorder)
		ctx.Request = httptest.NewRequest(http.MethodGet, "/api/v1/admin/submissions"+query, http.NoBody)
		(&SubmissionHandler{}).ListPendingSubmissions(ctx)
		if recorder.Code != http.StatusBadRequest {
			t.Fatalf("%s: expected 400, got %d", query, recorder.Code)
		}
	}
}

func TestSubmissionModerationRejectsMalformedIdentity(t *testing.T) {
	gin.SetMode(gin.TestMode)
	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Request = httptest.NewRequest(http.MethodPost, "/api/v1/admin/submissions/"+uuid.NewString()+"/approve", http.NoBody)
	ctx.Params = gin.Params{{Key: "id", Value: uuid.NewString()}}
	ctx.Set("user_id", "not-a-uuid")
	(&SubmissionHandler{}).ApproveSubmission(ctx)
	if recorder.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", recorder.Code)
	}
}

func TestBulkSubmissionModerationRequiresUniqueBoundedIDs(t *testing.T) {
	gin.SetMode(gin.TestMode)
	id := uuid.NewString()
	for _, body := range []string{
		`{"submission_ids":["` + id + `","` + id + `"]}`,
		`{"submission_ids":["not-a-uuid"]}`,
	} {
		recorder := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(recorder)
		ctx.Request = httptest.NewRequest(http.MethodPost, "/api/v1/admin/submissions/bulk-approve", strings.NewReader(body))
		ctx.Request.Header.Set("Content-Type", "application/json")
		ctx.Set("user_id", uuid.New())
		(&SubmissionHandler{}).BulkApproveSubmissions(ctx)
		if recorder.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d", recorder.Code)
		}
	}
}
