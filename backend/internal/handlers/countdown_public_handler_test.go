package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"git.subcult.tv/subculture-collective/clpr/internal/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type fakeCountdownPublicStore struct {
	gotSeries, gotBefore, gotBroadcaster *string
	gotDays                              int
	gotClip                              *uuid.UUID
}

func (f *fakeCountdownPublicStore) ListPublicEditions(_ context.Context, series, before *string, days int) ([]models.PublicCountdownEdition, *string, error) {
	f.gotSeries, f.gotBefore, f.gotDays = series, before, days
	next := "2026-11-06"
	return []models.PublicCountdownEdition{{EditionDate: "2026-11-07", SeriesSlug: "overall", Clips: []models.PublicCountdownClip{}}}, &next, nil
}

func (f *fakeCountdownPublicStore) GetPublicEdition(_ context.Context, date, series string) (*models.PublicCountdownEdition, error) {
	switch {
	case series == "broken":
		return nil, errors.New("database down")
	case date == "2026-11-07" && series == "overall":
		return &models.PublicCountdownEdition{EditionDate: date, SeriesSlug: series, Size: 10, Clips: []models.PublicCountdownClip{}}, nil
	}
	return nil, nil
}

func (f *fakeCountdownPublicStore) ListCountdownFeatures(_ context.Context, clipID *uuid.UUID, broadcasterID *string, _ int) ([]models.CountdownFeature, error) {
	f.gotClip, f.gotBroadcaster = clipID, broadcasterID
	return []models.CountdownFeature{}, nil
}

func TestCountdownPublicHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)
	store := &fakeCountdownPublicStore{}
	h := NewCountdownPublicHandler(store)
	router := gin.New()
	router.GET("/countdowns", h.ListEditions)
	router.GET("/countdowns/featured", h.ListFeatures)
	router.GET("/countdowns/:date/:series", h.GetEdition)

	clip := uuid.New().String()
	cases := []struct {
		path string
		want int
	}{
		{"/countdowns", 200},
		{"/countdowns?days=0", 400},
		{"/countdowns?days=61", 400},
		{"/countdowns?before=yesterday", 400},
		{"/countdowns?series=Bad_Slug", 400},
		{"/countdowns/2026-11-07/overall", 200},
		{"/countdowns/2026-11-06/overall", 404},
		{"/countdowns/2026-13-01/overall", 404},
		{"/countdowns/2026-11-07/NOPE", 404},
		{"/countdowns/2026-11-07/broken", 500},
		{"/countdowns/featured", 400},
		{"/countdowns/featured?clip_id=" + clip + "&broadcaster_id=1", 400},
		{"/countdowns/featured?clip_id=nope", 400},
		{"/countdowns/featured?broadcaster_id=" + strings.Repeat("9", 65), 400},
		{"/countdowns/featured?clip_id=" + clip, 200},
		{"/countdowns/featured?broadcaster_id=123", 200},
	}
	for _, tc := range cases {
		w := httptest.NewRecorder()
		router.ServeHTTP(w, httptest.NewRequest(http.MethodGet, tc.path, nil))
		if w.Code != tc.want {
			t.Errorf("GET %s = %d, want %d: %s", tc.path, w.Code, tc.want, w.Body.String())
		}
		if w.Code == 500 && strings.Contains(w.Body.String(), "database down") {
			t.Errorf("GET %s leaked the internal error", tc.path)
		}
	}

	w := httptest.NewRecorder()
	router.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/countdowns?days=7&before=2026-11-07&series=just-chatting", nil))
	if store.gotDays != 7 || *store.gotBefore != "2026-11-07" || *store.gotSeries != "just-chatting" {
		t.Fatalf("store got days %d before %v series %v", store.gotDays, store.gotBefore, store.gotSeries)
	}
	var body struct {
		Meta struct {
			NextBefore *string `json:"next_before"`
		} `json:"meta"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil || body.Meta.NextBefore == nil || *body.Meta.NextBefore != "2026-11-06" {
		t.Fatalf("meta = %+v, %v", body.Meta, err)
	}
	if w.Header().Get("Cache-Control") != "public, max-age=300" {
		t.Fatalf("Cache-Control = %q", w.Header().Get("Cache-Control"))
	}
}
