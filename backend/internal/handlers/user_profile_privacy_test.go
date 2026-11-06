package handlers

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"git.subcult.tv/subculture-collective/clpr/internal/models"
	"git.subcult.tv/subculture-collective/clpr/internal/repository"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type fakeProfileSettings struct {
	settings *models.UserSettings
	err      error
}

func (f fakeProfileSettings) GetSettings(context.Context, uuid.UUID) (*models.UserSettings, error) {
	return f.settings, f.err
}

func TestProfilePrivacyGuard(t *testing.T) {
	gin.SetMode(gin.TestMode)
	owner := uuid.New()
	stranger := uuid.New()
	settings := func(visibility string, showKarma bool) fakeProfileSettings {
		return fakeProfileSettings{settings: &models.UserSettings{UserID: owner, ProfileVisibility: visibility, ShowKarmaPublicly: showKarma}}
	}

	tests := []struct {
		name        string
		settings    fakeProfileSettings
		viewer      uuid.UUID
		role        string
		karmaRoute  bool
		path        string
		wantStatus  int
		wantHidden  bool
		wantReached bool
	}{
		{name: "public anonymous", settings: settings("public", true), wantStatus: http.StatusOK, wantReached: true},
		{name: "private anonymous", settings: settings("private", true), wantStatus: http.StatusNotFound},
		{name: "private stranger", settings: settings("private", true), viewer: stranger, role: models.RoleUser, wantStatus: http.StatusNotFound},
		{name: "private owner", settings: settings("private", false), viewer: owner, role: models.RoleUser, wantStatus: http.StatusOK, wantReached: true},
		{name: "private moderator", settings: settings("private", true), viewer: stranger, role: models.RoleModerator, wantStatus: http.StatusOK, wantReached: true},
		{name: "private admin", settings: settings("private", true), viewer: stranger, role: models.RoleAdmin, wantStatus: http.StatusOK, wantReached: true},
		{name: "followers anonymous", settings: settings("followers", true), wantStatus: http.StatusNotFound},
		{name: "unknown visibility fails closed", settings: settings("friends", true), wantStatus: http.StatusNotFound},
		{name: "missing settings row uses public defaults", settings: fakeProfileSettings{err: repository.ErrUserNotFound}, wantStatus: http.StatusOK, wantReached: true},
		{name: "settings failure", settings: fakeProfileSettings{err: errors.New("db down")}, wantStatus: http.StatusInternalServerError},
		{name: "hidden karma on profile", settings: settings("public", false), wantStatus: http.StatusOK, wantReached: true, wantHidden: true},
		{name: "hidden karma for moderator", settings: settings("public", false), viewer: stranger, role: models.RoleModerator, wantStatus: http.StatusOK, wantReached: true, wantHidden: true},
		{name: "hidden karma route", settings: settings("public", false), karmaRoute: true, wantStatus: http.StatusNotFound},
		{name: "hidden karma route owner", settings: settings("public", false), viewer: owner, karmaRoute: true, wantStatus: http.StatusOK, wantReached: true},
		{name: "invalid id reaches handler", settings: fakeProfileSettings{err: errors.New("unused")}, path: "/users/not-a-uuid", wantStatus: http.StatusOK, wantReached: true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			reached, hidden := false, false
			router := gin.New()
			router.GET("/users/:id", func(c *gin.Context) {
				if test.viewer != uuid.Nil {
					c.Set("user_id", test.viewer)
					c.Set("user_role", test.role)
				}
			}, (&UserHandler{}).ProfilePrivacy(test.settings, test.karmaRoute), func(c *gin.Context) {
				reached, hidden = true, hideProfileKarma(c)
				c.Status(http.StatusOK)
			})
			path := test.path
			if path == "" {
				path = "/users/" + owner.String()
			}
			rec := httptest.NewRecorder()
			router.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, path, nil))
			if rec.Code != test.wantStatus || reached != test.wantReached || hidden != test.wantHidden {
				t.Fatalf("status=%d reached=%v hidden=%v body=%s; want status=%d reached=%v hidden=%v",
					rec.Code, reached, hidden, rec.Body.String(), test.wantStatus, test.wantReached, test.wantHidden)
			}
			if rec.Code == http.StatusNotFound && rec.Body.String() != `{"error":"user not found"}` {
				t.Fatalf("404 body = %s, want the unknown-user body", rec.Body.String())
			}
		})
	}
}

func TestRedactPrivateProfileFields(t *testing.T) {
	gin.SetMode(gin.TestMode)
	owner := uuid.New()
	email := "owner@example.test"
	now := time.Now()
	tests := []struct {
		name     string
		viewer   uuid.UUID
		role     string
		wantKept bool
	}{
		{name: "anonymous", wantKept: false},
		{name: "another user", viewer: uuid.New(), role: models.RoleUser, wantKept: false},
		{name: "owner", viewer: owner, role: models.RoleUser, wantKept: true},
		{name: "moderator", viewer: uuid.New(), role: models.RoleModerator, wantKept: true},
		{name: "admin", viewer: uuid.New(), role: models.RoleAdmin, wantKept: true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			c, _ := gin.CreateTestContext(httptest.NewRecorder())
			if test.viewer != uuid.Nil {
				c.Set("user_id", test.viewer)
				c.Set("user_role", test.role)
			}
			profile := &models.UserProfile{User: models.User{
				ID: owner, Email: &email, TrustScore: 70, TrustScoreUpdatedAt: &now, LastLoginAt: &now,
			}}
			redactPrivateProfileFields(c, profile)
			kept := profile.Email != nil && profile.TrustScore == 70 &&
				profile.TrustScoreUpdatedAt != nil && profile.LastLoginAt != nil
			cleared := profile.Email == nil && profile.TrustScore == 0 &&
				profile.TrustScoreUpdatedAt == nil && profile.LastLoginAt == nil
			if test.wantKept && !kept {
				t.Fatalf("private fields were cleared for %s", test.name)
			}
			if !test.wantKept && !cleared {
				t.Fatalf("private fields leaked to %s: %+v", test.name, profile.User)
			}
		})
	}
}
