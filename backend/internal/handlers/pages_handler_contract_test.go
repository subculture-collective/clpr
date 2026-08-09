package handlers

import (
	"context"
	"database/sql"
	"errors"
	"html/template"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"git.subcult.tv/subculture-collective/clpr/internal/models"
	"github.com/gin-gonic/gin"
)

type pagesClipRepositoryStub struct {
	clips []models.Clip
	total int
	err   error
}

func (s *pagesClipRepositoryStub) ListClipsByBroadcaster(context.Context, string, string, int, int) ([]models.Clip, int, error) {
	return s.clips, s.total, s.err
}
func (s *pagesClipRepositoryStub) ListClipsByGame(context.Context, string, int, int) ([]models.Clip, int, error) {
	return s.clips, s.total, s.err
}
func (s *pagesClipRepositoryStub) ListClipsForBestOf(context.Context, time.Time, time.Time, int, int) ([]models.Clip, int, error) {
	return s.clips, s.total, s.err
}
func (s *pagesClipRepositoryStub) ListClipsForStreamerGame(context.Context, string, string, int, int) ([]models.Clip, int, error) {
	return s.clips, s.total, s.err
}

type pagesBroadcasterRepositoryStub struct{ err error }

func (s *pagesBroadcasterRepositoryStub) GetBroadcasterByName(context.Context, string) (string, error) {
	return "broadcaster-1", s.err
}
func (s *pagesBroadcasterRepositoryStub) GetBroadcasterStats(context.Context, string) (int, int64, float64, error) {
	return 1, 100, 4.5, s.err
}
func (s *pagesBroadcasterRepositoryStub) GetFollowerCount(context.Context, string) (int, error) {
	return 5, s.err
}
func (s *pagesBroadcasterRepositoryStub) ListBroadcasterGames(context.Context, string) ([]models.GameWithClipCount, error) {
	return []models.GameWithClipCount{}, s.err
}

type pagesGameRepositoryStub struct{ err error }

func (s *pagesGameRepositoryStub) GetBySlug(context.Context, string) (*models.GameEntity, error) {
	return &models.GameEntity{TwitchGameID: "game-1", Name: "A Game", Slug: "a-game"}, s.err
}
func (s *pagesGameRepositoryStub) GetByTwitchGameID(context.Context, string) (*models.GameEntity, error) {
	return &models.GameEntity{TwitchGameID: "game-1", Name: "A Game", Slug: "a-game"}, s.err
}
func (s *pagesGameRepositoryStub) ListTopBroadcastersForGame(context.Context, string, int) ([]models.BroadcasterWithClipCount, error) {
	return []models.BroadcasterWithClipCount{}, s.err
}

func pagesTestRouter(handler *PagesHandler) http.Handler {
	router := gin.New()
	tmpl := template.New("")
	for _, name := range []string{"streamer.html", "game.html", "streamer_game.html", "404.html"} {
		template.Must(tmpl.New(name).Parse("<html><body>{{.Title}}</body></html>"))
	}
	router.SetHTMLTemplate(tmpl)
	router.GET("/clips/streamer/:broadcasterName", handler.GetStreamerPage)
	router.GET("/clips/game/:gameSlug", handler.GetGamePage)
	router.GET("/clips/streamer/:broadcasterName/:gameSlug", handler.GetStreamerGamePage)
	return router
}

func TestPagesHandlerHTMLContracts(t *testing.T) {
	handler := NewPagesHandler(&pagesClipRepositoryStub{total: 1}, &pagesBroadcasterRepositoryStub{}, &pagesGameRepositoryStub{})
	router := pagesTestRouter(handler)
	for _, target := range []string{"/clips/streamer/streamer", "/clips/game/a-game", "/clips/streamer/streamer/a-game"} {
		recorder := httptest.NewRecorder()
		request := httptest.NewRequest(http.MethodGet, target, nil)
		router.ServeHTTP(recorder, request)
		if recorder.Code != http.StatusOK || !strings.HasPrefix(recorder.Header().Get("Content-Type"), "text/html") {
			t.Fatalf("unexpected page response for %s: %d %s", target, recorder.Code, recorder.Body.String())
		}
	}
}

func TestPagesHandlerDistinguishesMissingFromUnavailable(t *testing.T) {
	t.Run("missing broadcaster", func(t *testing.T) {
		handler := NewPagesHandler(&pagesClipRepositoryStub{}, &pagesBroadcasterRepositoryStub{err: sql.ErrNoRows}, &pagesGameRepositoryStub{})
		recorder := httptest.NewRecorder()
		pagesTestRouter(handler).ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/clips/streamer/missing", nil))
		if recorder.Code != http.StatusNotFound {
			t.Fatalf("expected 404, got %d", recorder.Code)
		}
	})

	t.Run("broadcaster dependency failure", func(t *testing.T) {
		handler := NewPagesHandler(&pagesClipRepositoryStub{}, &pagesBroadcasterRepositoryStub{err: errors.New("database unavailable")}, &pagesGameRepositoryStub{})
		recorder := httptest.NewRecorder()
		pagesTestRouter(handler).ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/clips/streamer/streamer", nil))
		if recorder.Code != http.StatusInternalServerError || recorder.Header().Get("Cache-Control") != "no-store" {
			t.Fatalf("expected no-store 500, got %d %s", recorder.Code, recorder.Header().Get("Cache-Control"))
		}
	})

	t.Run("clip dependency failure", func(t *testing.T) {
		handler := NewPagesHandler(&pagesClipRepositoryStub{err: errors.New("database unavailable")}, &pagesBroadcasterRepositoryStub{}, &pagesGameRepositoryStub{})
		recorder := httptest.NewRecorder()
		pagesTestRouter(handler).ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/clips/game/a-game", nil))
		if recorder.Code != http.StatusInternalServerError {
			t.Fatalf("expected 500, got %d", recorder.Code)
		}
	})
}
