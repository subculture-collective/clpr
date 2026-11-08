package handlers

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"git.subcult.tv/subculture-collective/clpr/internal/models"
	"git.subcult.tv/subculture-collective/clpr/internal/repository"
	"git.subcult.tv/subculture-collective/clpr/internal/services"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type countdownHandlerStore struct {
	services.CountdownStore
}

func (countdownHandlerStore) CreateExclusion(_ context.Context, in *models.CountdownExclusionInput, source string, _ *uuid.UUID) (*models.CountdownExclusion, error) {
	if in.BroadcasterID != nil && *in.BroadcasterID == "taken" {
		return nil, repository.ErrCountdownExclusionExists
	}
	return &models.CountdownExclusion{ID: uuid.New(), BroadcasterID: in.BroadcasterID, Reason: in.Reason, Source: source}, nil
}

func (countdownHandlerStore) DeleteExclusion(context.Context, uuid.UUID) error {
	return repository.ErrCountdownExclusionNotFound
}

func (countdownHandlerStore) CreateSeries(context.Context, *models.CountdownSeriesInput) (*models.CountdownSeries, error) {
	return nil, repository.ErrCountdownSeriesSlugTaken
}

func TestCountdownHandlerStatusCodes(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewCountdownHandler(services.NewCountdownService(countdownHandlerStore{}))
	admin := uuid.New()
	router := gin.New()
	router.Use(func(c *gin.Context) {
		if c.GetHeader("X-Test-User") != "" {
			c.Set("user_id", admin)
		}
	})
	router.POST("/exclusions", h.CreateExclusion)
	router.DELETE("/exclusions/:id", h.DeleteExclusion)
	router.POST("/series", h.CreateSeries)

	cases := []struct {
		name, method, path, body string
		auth                     bool
		want                     int
	}{
		{"exclusion without a user", http.MethodPost, "/exclusions", `{"broadcaster_id":"1","reason":"request"}`, false, http.StatusUnauthorized},
		{"exclusion with both targets", http.MethodPost, "/exclusions", `{"broadcaster_id":"1","clip_id":"` + uuid.NewString() + `","reason":"request"}`, true, http.StatusBadRequest},
		{"exclusion with an unknown reason", http.MethodPost, "/exclusions", `{"broadcaster_id":"1","reason":"dislike"}`, true, http.StatusBadRequest},
		{"exclusion already present", http.MethodPost, "/exclusions", `{"broadcaster_id":"taken","reason":"request"}`, true, http.StatusConflict},
		{"exclusion created", http.MethodPost, "/exclusions", `{"broadcaster_id":"1","reason":"request"}`, true, http.StatusCreated},
		{"delete with a bad id", http.MethodDelete, "/exclusions/nope", ``, true, http.StatusBadRequest},
		{"delete a missing exclusion", http.MethodDelete, "/exclusions/" + uuid.NewString(), ``, true, http.StatusNotFound},
		{"series with a bad slug", http.MethodPost, "/series", `{"slug":"Bad Slug","name":"x"}`, true, http.StatusBadRequest},
		{"series slug taken", http.MethodPost, "/series", `{"slug":"overall","name":"Overall"}`, true, http.StatusConflict},
	}
	for _, tc := range cases {
		req := httptest.NewRequest(tc.method, tc.path, strings.NewReader(tc.body))
		req.Header.Set("Content-Type", "application/json")
		if tc.auth {
			req.Header.Set("X-Test-User", "1")
		}
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)
		if rec.Code != tc.want {
			t.Errorf("%s: status %d, want %d (%s)", tc.name, rec.Code, tc.want, rec.Body.String())
		}
	}
}
