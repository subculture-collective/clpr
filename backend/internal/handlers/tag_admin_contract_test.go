package handlers

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestValidateAdminTagFields(t *testing.T) {
	description := strings.Repeat("d", 1001)
	for _, tc := range []struct {
		name, slug  string
		description *string
	}{
		{" ", "valid", nil}, {"Valid", "UPPER_CASE", nil}, {"Valid", "bad--slug", nil}, {"Valid", "valid", &description},
	} {
		if _, _, ok := validateAdminTagFields(tc.name, tc.slug, tc.description); ok {
			t.Fatalf("expected invalid fields: %#v", tc)
		}
	}
	name, slug, ok := validateAdminTagFields("  Good Tag  ", " Good-Tag ", nil)
	if !ok || name != "Good Tag" || slug != "good-tag" {
		t.Fatalf("unexpected canonical fields: %q %q %v", name, slug, ok)
	}
}

func TestAddBlacklistedTagRejectsMalformedIdentityAndPattern(t *testing.T) {
	gin.SetMode(gin.TestMode)
	tests := []struct {
		body     string
		identity any
	}{
		{`{"pattern":"bad pattern"}`, nil},
		{`{"pattern":"valid-tag"}`, "not-a-uuid"},
	}
	for _, tc := range tests {
		recorder := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(recorder)
		ctx.Request = httptest.NewRequest(http.MethodPost, "/api/v1/admin/tags/blacklist", strings.NewReader(tc.body))
		ctx.Request.Header.Set("Content-Type", "application/json")
		if tc.identity != nil {
			ctx.Set("user_id", tc.identity)
		}
		(&TagHandler{}).AddBlacklistedTag(ctx)
		if recorder.Code != http.StatusBadRequest && recorder.Code != http.StatusUnauthorized {
			t.Fatalf("expected 400/401, got %d", recorder.Code)
		}
	}
}
