package handlers

import (
	"errors"
	"net/http"
	"strconv"

	"git.subcult.tv/subculture-collective/clpr/internal/models"
	"git.subcult.tv/subculture-collective/clpr/internal/repository"
	"git.subcult.tv/subculture-collective/clpr/internal/services"
	"git.subcult.tv/subculture-collective/clpr/pkg/utils"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// CountdownHandler serves the admin countdown endpoints (#76).
type CountdownHandler struct {
	service *services.CountdownService
}

// NewCountdownHandler creates a CountdownHandler.
func NewCountdownHandler(service *services.CountdownService) *CountdownHandler {
	return &CountdownHandler{service: service}
}

func countdownFail(c *gin.Context, status int, code, message string) {
	c.JSON(status, StandardResponse{Success: false, Error: &ErrorInfo{Code: code, Message: message}})
}

// countdownError maps service and repository errors to responses.
func countdownError(c *gin.Context, err error, action string) {
	switch {
	case errors.Is(err, services.ErrCountdownValidation):
		countdownFail(c, http.StatusBadRequest, "INVALID_REQUEST", err.Error())
	case errors.Is(err, repository.ErrCountdownSeriesNotFound):
		countdownFail(c, http.StatusNotFound, "NOT_FOUND", "Countdown series not found")
	case errors.Is(err, repository.ErrCountdownExclusionNotFound):
		countdownFail(c, http.StatusNotFound, "NOT_FOUND", "Countdown exclusion not found")
	case errors.Is(err, repository.ErrCountdownExclusionClipUnset):
		countdownFail(c, http.StatusBadRequest, "INVALID_REQUEST", "clip_id is not a clip")
	case errors.Is(err, repository.ErrCountdownSeriesSlugTaken):
		countdownFail(c, http.StatusConflict, "CONFLICT", "A countdown series with this slug already exists")
	case errors.Is(err, repository.ErrCountdownExclusionExists):
		countdownFail(c, http.StatusConflict, "CONFLICT", "This broadcaster or clip is already excluded")
	case errors.Is(err, repository.ErrCountdownEditionInProgress):
		countdownFail(c, http.StatusConflict, "CONFLICT", "The edition has moved past selection and cannot be selected again")
	default:
		utils.Error("Countdown admin request failed", err, map[string]interface{}{"action": action})
		countdownFail(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to "+action)
	}
}

func countdownPathID(c *gin.Context) (uuid.UUID, bool) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		countdownFail(c, http.StatusBadRequest, "INVALID_REQUEST", "Invalid ID")
		return uuid.Nil, false
	}
	return id, true
}

// ListSeries handles GET /api/v1/admin/countdowns/series.
func (h *CountdownHandler) ListSeries(c *gin.Context) {
	series, err := h.service.ListSeries(c.Request.Context())
	if err != nil {
		countdownError(c, err, "list countdown series")
		return
	}
	c.JSON(http.StatusOK, StandardResponse{Success: true, Data: series})
}

// CreateSeries handles POST /api/v1/admin/countdowns/series.
func (h *CountdownHandler) CreateSeries(c *gin.Context) {
	var in models.CountdownSeriesInput
	if err := c.ShouldBindJSON(&in); err != nil {
		countdownFail(c, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body")
		return
	}
	series, err := h.service.CreateSeries(c.Request.Context(), &in)
	if err != nil {
		countdownError(c, err, "create countdown series")
		return
	}
	c.JSON(http.StatusCreated, StandardResponse{Success: true, Data: series})
}

// UpdateSeries handles PUT /api/v1/admin/countdowns/series/:id.
func (h *CountdownHandler) UpdateSeries(c *gin.Context) {
	id, ok := countdownPathID(c)
	if !ok {
		return
	}
	var in models.CountdownSeriesInput
	if err := c.ShouldBindJSON(&in); err != nil {
		countdownFail(c, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body")
		return
	}
	series, err := h.service.UpdateSeries(c.Request.Context(), id, &in)
	if err != nil {
		countdownError(c, err, "update countdown series")
		return
	}
	c.JSON(http.StatusOK, StandardResponse{Success: true, Data: series})
}

// ReselectEdition handles POST /api/v1/admin/countdowns/series/:id/editions.
// It selects the series edition for the given date again.
func (h *CountdownHandler) ReselectEdition(c *gin.Context) {
	id, ok := countdownPathID(c)
	if !ok {
		return
	}
	var body struct {
		Date string `json:"date"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		countdownFail(c, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body")
		return
	}
	edition, err := h.service.ReselectEdition(c.Request.Context(), id, body.Date)
	if err != nil {
		countdownError(c, err, "select countdown edition")
		return
	}
	c.JSON(http.StatusOK, StandardResponse{Success: true, Data: edition})
}

// ListEditions handles GET /api/v1/admin/countdowns/editions.
func (h *CountdownHandler) ListEditions(c *gin.Context) {
	limit := 30
	if value := c.Query("limit"); value != "" {
		parsed, err := strconv.Atoi(value)
		if err != nil || parsed < 1 || parsed > 200 {
			countdownFail(c, http.StatusBadRequest, "INVALID_REQUEST", "limit must be between 1 and 200")
			return
		}
		limit = parsed
	}
	var date *string
	if value := c.Query("date"); value != "" {
		date = &value
	}
	var seriesID *uuid.UUID
	if value := c.Query("series_id"); value != "" {
		parsed, err := uuid.Parse(value)
		if err != nil {
			countdownFail(c, http.StatusBadRequest, "INVALID_REQUEST", "Invalid series_id")
			return
		}
		seriesID = &parsed
	}
	editions, err := h.service.ListEditions(c.Request.Context(), date, seriesID, limit)
	if err != nil {
		countdownError(c, err, "list countdown editions")
		return
	}
	c.JSON(http.StatusOK, StandardResponse{Success: true, Data: editions})
}

// ListExclusions handles GET /api/v1/admin/countdowns/exclusions.
func (h *CountdownHandler) ListExclusions(c *gin.Context) {
	exclusions, err := h.service.ListExclusions(c.Request.Context())
	if err != nil {
		countdownError(c, err, "list countdown exclusions")
		return
	}
	c.JSON(http.StatusOK, StandardResponse{Success: true, Data: exclusions})
}

// CreateExclusion handles POST /api/v1/admin/countdowns/exclusions.
func (h *CountdownHandler) CreateExclusion(c *gin.Context) {
	adminID, ok := c.Get("user_id")
	userID, isUUID := adminID.(uuid.UUID)
	if !ok || !isUUID || userID == uuid.Nil {
		countdownFail(c, http.StatusUnauthorized, "UNAUTHORIZED", "Authentication required")
		return
	}
	var in models.CountdownExclusionInput
	if err := c.ShouldBindJSON(&in); err != nil {
		countdownFail(c, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body")
		return
	}
	exclusion, err := h.service.CreateExclusion(c.Request.Context(), &in, userID)
	if err != nil {
		countdownError(c, err, "create countdown exclusion")
		return
	}
	c.JSON(http.StatusCreated, StandardResponse{Success: true, Data: exclusion})
}

// DeleteExclusion handles DELETE /api/v1/admin/countdowns/exclusions/:id.
func (h *CountdownHandler) DeleteExclusion(c *gin.Context) {
	id, ok := countdownPathID(c)
	if !ok {
		return
	}
	if err := h.service.DeleteExclusion(c.Request.Context(), id); err != nil {
		countdownError(c, err, "delete countdown exclusion")
		return
	}
	c.JSON(http.StatusOK, StandardResponse{Success: true, Data: gin.H{"id": id}})
}
