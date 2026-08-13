//go:build integration

package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"git.subcult.tv/subculture-collective/clpr/internal/models"
	"git.subcult.tv/subculture-collective/clpr/internal/repository"
	"git.subcult.tv/subculture-collective/clpr/internal/testutil"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func TestAdminListUsersReturnsIdentityTaxonomySummary(t *testing.T) {
	pool := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(t, pool)
	testutil.TruncateTables(t, pool, "users")

	ctx := context.Background()
	for _, identity := range []struct {
		username string
		role     string
		status   string
	}{
		{"signed-in-user", "user", "active"},
		{"imported-creator", "user", "unclaimed"},
		{"staff-admin", "admin", "active"},
		{"other-service", "service", "active"},
	} {
		if _, err := pool.Exec(ctx, `
			INSERT INTO users (id, twitch_id, username, display_name, role, account_status)
			VALUES ($1, $2, $3, $3, $4, $5)
		`, uuid.New(), uuid.NewString(), identity.username, identity.role, identity.status); err != nil {
			t.Fatalf("insert identity: %v", err)
		}
	}

	gin.SetMode(gin.TestMode)
	recorder := httptest.NewRecorder()
	ginContext, _ := gin.CreateTestContext(recorder)
	ginContext.Request = httptest.NewRequest(http.MethodGet, "/api/v1/admin/users", http.NoBody)

	(&AdminUserHandler{userRepo: repository.NewUserRepository(pool)}).ListUsers(ginContext)
	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", recorder.Code, recorder.Body.String())
	}
	var response struct {
		Total   int                         `json:"total"`
		Summary models.AdminIdentitySummary `json:"summary"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if response.Total != 4 || response.Summary.SignedInUsers != 1 || response.Summary.UnclaimedCreators != 1 || response.Summary.Staff != 1 || response.Summary.Other != 1 {
		t.Fatalf("unexpected response: %#v", response)
	}
}
