package handlers

import (
	"context"
	"errors"
	"net/http"

	"git.subcult.tv/subculture-collective/clpr/internal/models"
	"git.subcult.tv/subculture-collective/clpr/internal/repository"
	"git.subcult.tv/subculture-collective/clpr/pkg/utils"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// CountdownControlsStore is the storage behind the broadcaster's countdown
// controls and the admin removal list (#75).
type CountdownControlsStore interface {
	BroadcasterCountdownState(ctx context.Context, broadcasterID string) (*models.BroadcasterCountdownState, error)
	SetSelfExclusion(ctx context.Context, broadcasterID string, userID uuid.UUID, excluded bool) error
	DeleteFeatureOptIns(ctx context.Context, userID uuid.UUID) error
	ListRemovalTasks(ctx context.Context) ([]models.CountdownRemovalTask, error)
	ResolveRemoval(ctx context.Context, in *models.CountdownRemovalInput, adminID uuid.UUID) error
	ListBroadcasterPlacements(ctx context.Context, broadcasterID string, limit int) ([]models.CountdownPlacement, error)
}

type countdownUserLookup interface {
	GetByID(ctx context.Context, id uuid.UUID) (*models.User, error)
}

type countdownTwitchAuthLookup interface {
	GetTwitchAuth(ctx context.Context, userID uuid.UUID) (*models.TwitchAuth, error)
}

// CountdownControlsHandler serves /users/me/countdowns and the admin
// removal list.
type CountdownControlsHandler struct {
	store  CountdownControlsStore
	users  countdownUserLookup
	twitch countdownTwitchAuthLookup
}

// NewCountdownControlsHandler creates a CountdownControlsHandler.
func NewCountdownControlsHandler(store CountdownControlsStore, users countdownUserLookup, twitch countdownTwitchAuthLookup) *CountdownControlsHandler {
	return &CountdownControlsHandler{store: store, users: users, twitch: twitch}
}

func (h *CountdownControlsHandler) internal(c *gin.Context, err error, action string) {
	utils.Error("Countdown controls request failed", err, map[string]interface{}{"action": action})
	countdownFail(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to "+action)
}

// broadcaster returns the signed-in user's Twitch ID, which is their
// broadcaster ID on clips.
func (h *CountdownControlsHandler) broadcaster(c *gin.Context) (uuid.UUID, string, bool) {
	userID, ok := currentUserID(c)
	if !ok {
		return uuid.Nil, "", false
	}
	user, err := h.users.GetByID(c.Request.Context(), userID)
	if err != nil {
		h.internal(c, err, "load your account")
		return uuid.Nil, "", false
	}
	if user == nil || user.TwitchID == nil || *user.TwitchID == "" {
		countdownFail(c, http.StatusConflict, "NO_TWITCH_ACCOUNT", "Sign in with Twitch to manage your clips in countdowns")
		return uuid.Nil, "", false
	}
	return userID, *user.TwitchID, true
}

func (h *CountdownControlsHandler) respondState(c *gin.Context, userID uuid.UUID, broadcasterID string) {
	ctx := c.Request.Context()
	state, err := h.store.BroadcasterCountdownState(ctx, broadcasterID)
	if err != nil {
		h.internal(c, err, "load your countdown settings")
		return
	}
	auth, err := h.twitch.GetTwitchAuth(ctx, userID)
	if err != nil {
		h.internal(c, err, "load your Twitch connection")
		return
	}
	state.ClipDownloadAuthorized = auth != nil && auth.TwitchUserID == broadcasterID && hasTwitchScope(auth.Scopes, "channel:manage:clips")
	c.JSON(http.StatusOK, StandardResponse{Success: true, Data: state})
}

// GetMyCountdownControls handles GET /api/v1/users/me/countdowns.
func (h *CountdownControlsHandler) GetMyCountdownControls(c *gin.Context) {
	userID, broadcasterID, ok := h.broadcaster(c)
	if !ok {
		return
	}
	h.respondState(c, userID, broadcasterID)
}

// SetMyCountdownExclusion handles PUT /api/v1/users/me/countdowns/exclusion.
// Turning it on keeps the broadcaster's clips out of future editions and
// adds their clips in uploaded videos to the admin removal list.
func (h *CountdownControlsHandler) SetMyCountdownExclusion(c *gin.Context) {
	userID, broadcasterID, ok := h.broadcaster(c)
	if !ok {
		return
	}
	var body struct {
		Excluded *bool `json:"excluded"`
	}
	if err := c.ShouldBindJSON(&body); err != nil || body.Excluded == nil {
		countdownFail(c, http.StatusBadRequest, "INVALID_REQUEST", "excluded (true or false) is required")
		return
	}
	if err := h.store.SetSelfExclusion(c.Request.Context(), broadcasterID, userID, *body.Excluded); err != nil {
		h.internal(c, err, "save your countdown setting")
		return
	}
	h.respondState(c, userID, broadcasterID)
}

// DeleteMyFeatureOptIn handles DELETE /api/v1/users/me/countdowns/feature.
func (h *CountdownControlsHandler) DeleteMyFeatureOptIn(c *gin.Context) {
	userID, broadcasterID, ok := h.broadcaster(c)
	if !ok {
		return
	}
	if err := h.store.DeleteFeatureOptIns(c.Request.Context(), userID); err != nil {
		h.internal(c, err, "stop featuring your clips")
		return
	}
	h.respondState(c, userID, broadcasterID)
}

// ListMyPlacements handles GET /api/v1/users/me/countdowns/placements: the
// signed-in broadcaster's clips in public countdowns, newest first, with the
// YouTube videos they appear in and those videos' view counts (#84).
func (h *CountdownControlsHandler) ListMyPlacements(c *gin.Context) {
	_, broadcasterID, ok := h.broadcaster(c)
	if !ok {
		return
	}
	placements, err := h.store.ListBroadcasterPlacements(c.Request.Context(), broadcasterID, 50)
	if err != nil {
		h.internal(c, err, "list your countdown placements")
		return
	}
	c.JSON(http.StatusOK, StandardResponse{Success: true, Data: placements})
}

// ListRemovals handles GET /api/v1/admin/countdowns/removals.
func (h *CountdownControlsHandler) ListRemovals(c *gin.Context) {
	tasks, err := h.store.ListRemovalTasks(c.Request.Context())
	if err != nil {
		h.internal(c, err, "list countdown removals")
		return
	}
	c.JSON(http.StatusOK, StandardResponse{Success: true, Data: tasks})
}

// ResolveRemoval handles POST /api/v1/admin/countdowns/removals.
func (h *CountdownControlsHandler) ResolveRemoval(c *gin.Context) {
	adminID, ok := currentUserID(c)
	if !ok {
		return
	}
	var in models.CountdownRemovalInput
	if err := c.ShouldBindJSON(&in); err != nil || in.RenderID == uuid.Nil || in.ClipID == uuid.Nil {
		countdownFail(c, http.StatusBadRequest, "INVALID_REQUEST", "render_id and clip_id are required")
		return
	}
	switch in.Resolution {
	case "removed", "edited", "kept":
	default:
		countdownFail(c, http.StatusBadRequest, "INVALID_REQUEST", "resolution must be removed, edited or kept")
		return
	}
	if in.Note != nil && len(*in.Note) > 500 {
		countdownFail(c, http.StatusBadRequest, "INVALID_REQUEST", "note must be at most 500 characters")
		return
	}
	err := h.store.ResolveRemoval(c.Request.Context(), &in, adminID)
	if errors.Is(err, repository.ErrCountdownRemovalNotFound) {
		countdownFail(c, http.StatusNotFound, "NOT_FOUND", "No uploaded video matches that render and clip")
		return
	}
	if err != nil {
		h.internal(c, err, "resolve the removal")
		return
	}
	c.JSON(http.StatusOK, StandardResponse{Success: true})
}
