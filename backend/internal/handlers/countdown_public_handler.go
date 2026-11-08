package handlers

import (
	"context"
	"net/http"
	"regexp"
	"strconv"
	"time"

	"git.subcult.tv/subculture-collective/clpr/internal/models"
	"git.subcult.tv/subculture-collective/clpr/pkg/utils"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// CountdownPublicStore is the storage the public countdown pages read.
type CountdownPublicStore interface {
	ListPublicEditions(ctx context.Context, seriesSlug, before *string, days int) ([]models.PublicCountdownEdition, *string, error)
	GetPublicEdition(ctx context.Context, date, seriesSlug string) (*models.PublicCountdownEdition, error)
	ListCountdownFeatures(ctx context.Context, clipID *uuid.UUID, broadcasterID *string, limit int) ([]models.CountdownFeature, error)
}

// CountdownPublicHandler serves the public countdown pages (#82).
type CountdownPublicHandler struct {
	store CountdownPublicStore
}

// NewCountdownPublicHandler creates a CountdownPublicHandler.
func NewCountdownPublicHandler(store CountdownPublicStore) *CountdownPublicHandler {
	return &CountdownPublicHandler{store: store}
}

var publicCountdownSlug = regexp.MustCompile(`^[a-z0-9]+(-[a-z0-9]+)*$`)

func validCountdownDate(value string) bool {
	_, err := time.Parse("2006-01-02", value)
	return err == nil
}

func validCountdownSlug(value string) bool {
	return len(value) <= 64 && publicCountdownSlug.MatchString(value)
}

func (h *CountdownPublicHandler) fail(c *gin.Context, err error, action string) {
	utils.Error("Public countdown request failed", err, map[string]interface{}{"action": action})
	countdownFail(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to "+action)
}

// ListEditions handles GET /api/v1/countdowns.
func (h *CountdownPublicHandler) ListEditions(c *gin.Context) {
	days := 14
	if value := c.Query("days"); value != "" {
		parsed, err := strconv.Atoi(value)
		if err != nil || parsed < 1 || parsed > 60 {
			countdownFail(c, http.StatusBadRequest, "INVALID_REQUEST", "days must be between 1 and 60")
			return
		}
		days = parsed
	}
	var before, series *string
	if value := c.Query("before"); value != "" {
		if !validCountdownDate(value) {
			countdownFail(c, http.StatusBadRequest, "INVALID_REQUEST", "before must be YYYY-MM-DD")
			return
		}
		before = &value
	}
	if value := c.Query("series"); value != "" {
		if !validCountdownSlug(value) {
			countdownFail(c, http.StatusBadRequest, "INVALID_REQUEST", "Invalid series")
			return
		}
		series = &value
	}
	editions, next, err := h.store.ListPublicEditions(c.Request.Context(), series, before, days)
	if err != nil {
		h.fail(c, err, "list countdowns")
		return
	}
	c.Header("Cache-Control", "public, max-age=300")
	c.JSON(http.StatusOK, StandardResponse{Success: true, Data: editions, Meta: gin.H{"next_before": next}})
}

// GetEdition handles GET /api/v1/countdowns/:date/:series.
func (h *CountdownPublicHandler) GetEdition(c *gin.Context) {
	date, series := c.Param("date"), c.Param("series")
	if !validCountdownDate(date) || !validCountdownSlug(series) {
		countdownFail(c, http.StatusNotFound, "NOT_FOUND", "Countdown not found")
		return
	}
	edition, err := h.store.GetPublicEdition(c.Request.Context(), date, series)
	if err != nil {
		h.fail(c, err, "load countdown")
		return
	}
	if edition == nil {
		countdownFail(c, http.StatusNotFound, "NOT_FOUND", "Countdown not found")
		return
	}
	c.Header("Cache-Control", "public, max-age=300")
	c.JSON(http.StatusOK, StandardResponse{Success: true, Data: edition})
}

// ListFeatures handles GET /api/v1/countdowns/featured. Exactly one of
// clip_id and broadcaster_id is required.
func (h *CountdownPublicHandler) ListFeatures(c *gin.Context) {
	clipValue, broadcasterValue := c.Query("clip_id"), c.Query("broadcaster_id")
	if (clipValue == "") == (broadcasterValue == "") {
		countdownFail(c, http.StatusBadRequest, "INVALID_REQUEST", "Pass exactly one of clip_id and broadcaster_id")
		return
	}
	var clipID *uuid.UUID
	var broadcasterID *string
	if clipValue != "" {
		parsed, err := uuid.Parse(clipValue)
		if err != nil {
			countdownFail(c, http.StatusBadRequest, "INVALID_REQUEST", "Invalid clip_id")
			return
		}
		clipID = &parsed
	} else {
		if len(broadcasterValue) > 64 {
			countdownFail(c, http.StatusBadRequest, "INVALID_REQUEST", "Invalid broadcaster_id")
			return
		}
		broadcasterID = &broadcasterValue
	}
	features, err := h.store.ListCountdownFeatures(c.Request.Context(), clipID, broadcasterID, 20)
	if err != nil {
		h.fail(c, err, "list countdown features")
		return
	}
	c.Header("Cache-Control", "public, max-age=300")
	c.JSON(http.StatusOK, StandardResponse{Success: true, Data: features})
}
