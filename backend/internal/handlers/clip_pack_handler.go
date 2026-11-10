package handlers

import (
	"context"
	"net/http"
	"time"

	"git.subcult.tv/subculture-collective/clpr/internal/models"
	"git.subcult.tv/subculture-collective/clpr/internal/services"
	"git.subcult.tv/subculture-collective/clpr/pkg/utils"
	"github.com/gin-gonic/gin"
)

// ClipPackStore lists a broadcaster's weekly clip packs.
type ClipPackStore interface {
	ListBroadcasterPacks(ctx context.Context, broadcasterID string, limit int) ([]models.ClipPack, error)
}

// ClipPackHandler serves the broadcaster's weekly clip packs (#86).
type ClipPackHandler struct {
	store ClipPackStore
	users countdownUserLookup
	// links signs download links; without a secret, packs are listed
	// without one.
	links services.ClipPackLinks
	now   func() time.Time
}

// NewClipPackHandler creates a ClipPackHandler.
func NewClipPackHandler(store ClipPackStore, users countdownUserLookup, links services.ClipPackLinks) *ClipPackHandler {
	return &ClipPackHandler{store: store, users: users, links: links, now: time.Now}
}

// ListMyPacks handles GET /api/v1/users/me/clip-packs: the signed-in
// broadcaster's packs, newest week first, with a download link valid for a
// day for each ready pack.
func (h *ClipPackHandler) ListMyPacks(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		return
	}
	user, err := h.users.GetByID(c.Request.Context(), userID)
	if err != nil {
		utils.Error("Clip pack request failed", err, nil)
		countdownFail(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to load your account")
		return
	}
	if user == nil || user.TwitchID == nil || *user.TwitchID == "" {
		countdownFail(c, http.StatusConflict, "NO_TWITCH_ACCOUNT", "Sign in with Twitch to get clip packs")
		return
	}
	packs, err := h.store.ListBroadcasterPacks(c.Request.Context(), *user.TwitchID, 8)
	if err != nil {
		utils.Error("Clip pack request failed", err, nil)
		countdownFail(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to list your clip packs")
		return
	}
	now := h.now()
	canSign := len(h.links.Secret) > 0 && h.links.BaseURL != ""
	for i := range packs {
		p := &packs[i]
		if canSign && p.Status == models.ClipPackReady && p.ExpiresAt != nil && now.Before(*p.ExpiresAt) {
			link := h.links.URL(p.ID, *p.ExpiresAt, now)
			p.DownloadURL = &link
		}
	}
	c.Header("Cache-Control", "private, no-store")
	c.JSON(http.StatusOK, StandardResponse{Success: true, Data: packs})
}
