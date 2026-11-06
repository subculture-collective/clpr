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

type settingsRepoReader struct {
	repo *repository.UserSettingsRepository
}

func (r settingsRepoReader) GetSettings(ctx context.Context, userID uuid.UUID) (*models.UserSettings, error) {
	return r.repo.GetByUserID(ctx, userID)
}

func TestProfilePrivacySettingsAreEnforced(t *testing.T) {
	pool := testutil.SetupTestDB(t)
	ctx := context.Background()
	gin.SetMode(gin.TestMode)

	suffix := uuid.NewString()[:8]
	createUser := func(name string, karma int) (uuid.UUID, string) {
		t.Helper()
		id, username := uuid.New(), name+"-"+suffix
		if _, err := pool.Exec(ctx, `
			INSERT INTO users (id, twitch_id, username, display_name, role, karma_points)
			VALUES ($1, $2, $3, $3, 'user', $4)`, id, uuid.NewString(), username, karma); err != nil {
			t.Fatalf("insert %s: %v", name, err)
		}
		t.Cleanup(func() { _, _ = pool.Exec(context.Background(), `DELETE FROM users WHERE id = $1`, id) })
		return id, username
	}
	owner, ownerName := createUser("pp-owner", 4321)
	follower, _ := createUser("pp-follower", 0)
	stranger, _ := createUser("pp-stranger", 0)
	if _, err := pool.Exec(ctx, `INSERT INTO user_follows (follower_id, following_id) VALUES ($1, $2)`, follower, owner); err != nil {
		t.Fatalf("insert follow: %v", err)
	}
	setPrivacy := func(visibility string, showKarma bool) {
		t.Helper()
		if _, err := pool.Exec(ctx, `
			INSERT INTO user_settings (user_id, profile_visibility, show_karma_publicly) VALUES ($1, $2, $3)
			ON CONFLICT (user_id) DO UPDATE SET profile_visibility = $2, show_karma_publicly = $3`,
			owner, visibility, showKarma); err != nil {
			t.Fatalf("set privacy: %v", err)
		}
	}

	userRepo := repository.NewUserRepository(pool)
	handler := &UserHandler{userRepo: userRepo}
	privacy := handler.ProfilePrivacy(settingsRepoReader{repository.NewUserSettingsRepository(pool)}, false)
	get := func(viewer uuid.UUID, path string) *httptest.ResponseRecorder {
		router := gin.New()
		identity := func(c *gin.Context) {
			if viewer != uuid.Nil {
				c.Set("user_id", viewer)
				c.Set("user_role", models.RoleUser)
			}
		}
		router.GET("/users/by-username/:username", identity, privacy, handler.GetUserByUsername)
		router.GET("/users/:id", identity, privacy, handler.GetUserProfile)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, path, nil))
		return rec
	}
	karmaOf := func(rec *httptest.ResponseRecorder) int {
		t.Helper()
		var body struct {
			Data struct {
				KarmaPoints int `json:"karma_points"`
			} `json:"data"`
		}
		if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
			t.Fatalf("decode %s: %v", rec.Body.String(), err)
		}
		return body.Data.KarmaPoints
	}
	byID := "/users/" + owner.String()
	byName := "/users/by-username/" + ownerName + "?full=true"

	setPrivacy("followers", true)
	for _, path := range []string{byID, byName} {
		if rec := get(uuid.Nil, path); rec.Code != http.StatusNotFound {
			t.Errorf("anonymous %s: status %d, want 404", path, rec.Code)
		}
		if rec := get(stranger, path); rec.Code != http.StatusNotFound {
			t.Errorf("stranger %s: status %d, want 404", path, rec.Code)
		}
		if rec := get(follower, path); rec.Code != http.StatusOK {
			t.Errorf("follower %s: status %d, want 200", path, rec.Code)
		}
	}

	setPrivacy("public", false)
	for _, path := range []string{byID, byName, "/users/by-username/" + ownerName} {
		rec := get(stranger, path)
		if rec.Code != http.StatusOK || karmaOf(rec) != 0 {
			t.Errorf("stranger %s: status %d karma %d, want 200 with hidden karma", path, rec.Code, karmaOf(rec))
		}
		rec = get(owner, path)
		if rec.Code != http.StatusOK || karmaOf(rec) != 4321 {
			t.Errorf("owner %s: status %d karma %d, want 200 with karma 4321", path, rec.Code, karmaOf(rec))
		}
	}
}
