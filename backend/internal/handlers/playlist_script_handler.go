package handlers

import (
	"errors"
	"net/http"
	"strconv"

	"git.subcult.tv/subculture-collective/clpr/internal/models"
	"git.subcult.tv/subculture-collective/clpr/internal/services"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// PlaylistScriptHandler handles playlist script admin endpoints
type PlaylistScriptHandler struct {
	service *services.PlaylistScriptService
}

// NewPlaylistScriptHandler creates a new PlaylistScriptHandler
func NewPlaylistScriptHandler(service *services.PlaylistScriptService) *PlaylistScriptHandler {
	return &PlaylistScriptHandler{service: service}
}

// ListScripts handles GET /api/v1/admin/playlist-scripts
func (h *PlaylistScriptHandler) ListScripts(c *gin.Context) {
	limit := 100
	if value := c.Query("limit"); value != "" {
		parsed, err := strconv.Atoi(value)
		if err != nil || parsed < 1 || parsed > 500 {
			c.JSON(http.StatusBadRequest, StandardResponse{Success: false, Error: &ErrorInfo{Code: "INVALID_REQUEST", Message: "limit must be between 1 and 500"}})
			return
		}
		limit = parsed
	}
	scripts, err := h.service.ListScripts(c.Request.Context(), limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, StandardResponse{
			Success: false,
			Error: &ErrorInfo{
				Code:    "INTERNAL_ERROR",
				Message: "Failed to list playlist scripts",
			},
		})
		return
	}

	c.JSON(http.StatusOK, StandardResponse{
		Success: true,
		Data:    scripts,
	})
}

// CreateScript handles POST /api/v1/admin/playlist-scripts
func (h *PlaylistScriptHandler) CreateScript(c *gin.Context) {
	userIDVal, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, StandardResponse{
			Success: false,
			Error: &ErrorInfo{
				Code:    "UNAUTHORIZED",
				Message: "Authentication required",
			},
		})
		return
	}

	userID, ok := userIDVal.(uuid.UUID)
	if !ok || userID == uuid.Nil {
		c.JSON(http.StatusUnauthorized, StandardResponse{
			Success: false,
			Error: &ErrorInfo{
				Code:    "UNAUTHORIZED",
				Message: "Authentication required",
			},
		})
		return
	}

	var req models.CreatePlaylistScriptRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, StandardResponse{
			Success: false,
			Error: &ErrorInfo{
				Code:    "INVALID_REQUEST",
				Message: "Invalid request body",
			},
		})
		return
	}

	script, err := h.service.CreateScript(c.Request.Context(), userID, &req)
	if err != nil {
		if errors.Is(err, services.ErrPlaylistScriptValidation) {
			c.JSON(http.StatusBadRequest, StandardResponse{Success: false, Error: &ErrorInfo{Code: "INVALID_REQUEST", Message: "Playlist script configuration is invalid"}})
			return
		}
		c.JSON(http.StatusInternalServerError, StandardResponse{
			Success: false,
			Error: &ErrorInfo{
				Code:    "INTERNAL_ERROR",
				Message: "Failed to create playlist script",
			},
		})
		return
	}

	c.JSON(http.StatusCreated, StandardResponse{
		Success: true,
		Data:    script,
	})
}

// UpdateScript handles PUT /api/v1/admin/playlist-scripts/:id
func (h *PlaylistScriptHandler) UpdateScript(c *gin.Context) {
	scriptIDStr := c.Param("id")
	scriptID, err := uuid.Parse(scriptIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, StandardResponse{
			Success: false,
			Error: &ErrorInfo{
				Code:    "INVALID_REQUEST",
				Message: "Invalid script ID",
			},
		})
		return
	}

	var req models.UpdatePlaylistScriptRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, StandardResponse{
			Success: false,
			Error: &ErrorInfo{
				Code:    "INVALID_REQUEST",
				Message: "Invalid request body",
			},
		})
		return
	}

	script, err := h.service.UpdateScript(c.Request.Context(), scriptID, &req)
	if err != nil {
		if errors.Is(err, services.ErrPlaylistScriptNotFound) {
			c.JSON(http.StatusNotFound, StandardResponse{
				Success: false,
				Error: &ErrorInfo{
					Code:    "NOT_FOUND",
					Message: "Playlist script not found",
				},
			})
			return
		}
		if errors.Is(err, services.ErrPlaylistScriptValidation) {
			c.JSON(http.StatusBadRequest, StandardResponse{Success: false, Error: &ErrorInfo{Code: "INVALID_REQUEST", Message: "Playlist script configuration is invalid"}})
			return
		}
		c.JSON(http.StatusInternalServerError, StandardResponse{
			Success: false,
			Error: &ErrorInfo{
				Code:    "INTERNAL_ERROR",
				Message: "Failed to update playlist script",
			},
		})
		return
	}

	c.JSON(http.StatusOK, StandardResponse{
		Success: true,
		Data:    script,
	})
}

// DeleteScript handles DELETE /api/v1/admin/playlist-scripts/:id
func (h *PlaylistScriptHandler) DeleteScript(c *gin.Context) {
	scriptIDStr := c.Param("id")
	scriptID, err := uuid.Parse(scriptIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, StandardResponse{
			Success: false,
			Error: &ErrorInfo{
				Code:    "INVALID_REQUEST",
				Message: "Invalid script ID",
			},
		})
		return
	}

	if err := h.service.DeleteScript(c.Request.Context(), scriptID); err != nil {
		if errors.Is(err, services.ErrPlaylistScriptNotFound) {
			c.JSON(http.StatusNotFound, StandardResponse{
				Success: false,
				Error: &ErrorInfo{
					Code:    "NOT_FOUND",
					Message: "Playlist script not found",
				},
			})
			return
		}
		c.JSON(http.StatusInternalServerError, StandardResponse{
			Success: false,
			Error: &ErrorInfo{
				Code:    "INTERNAL_ERROR",
				Message: "Failed to delete playlist script",
			},
		})
		return
	}

	c.JSON(http.StatusOK, StandardResponse{
		Success: true,
		Data: map[string]string{
			"message": "Playlist script deleted successfully",
		},
	})
}

// GeneratePlaylist handles POST /api/v1/admin/playlist-scripts/:id/generate
func (h *PlaylistScriptHandler) GeneratePlaylist(c *gin.Context) {
	scriptIDStr := c.Param("id")
	scriptID, err := uuid.Parse(scriptIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, StandardResponse{
			Success: false,
			Error: &ErrorInfo{
				Code:    "INVALID_REQUEST",
				Message: "Invalid script ID",
			},
		})
		return
	}

	playlist, err := h.service.GeneratePlaylist(c.Request.Context(), scriptID)
	if err != nil {
		if errors.Is(err, services.ErrPlaylistScriptNotFound) {
			c.JSON(http.StatusNotFound, StandardResponse{
				Success: false,
				Error: &ErrorInfo{
					Code:    "NOT_FOUND",
					Message: "Playlist script not found",
				},
			})
			return
		}
		if errors.Is(err, services.ErrPlaylistScriptInactive) {
			c.JSON(http.StatusBadRequest, StandardResponse{
				Success: false,
				Error: &ErrorInfo{
					Code:    "INVALID_REQUEST",
					Message: "Playlist script is inactive",
				},
			})
			return
		}
		if errors.Is(err, services.ErrPlaylistGenerationEmpty) {
			c.JSON(http.StatusConflict, StandardResponse{Success: false, Error: &ErrorInfo{Code: "NO_MATCHING_CLIPS", Message: "No clips matched this playlist script"}})
			return
		}
		c.JSON(http.StatusInternalServerError, StandardResponse{
			Success: false,
			Error: &ErrorInfo{
				Code:    "INTERNAL_ERROR",
				Message: "Failed to generate playlist",
			},
		})
		return
	}

	c.JSON(http.StatusCreated, StandardResponse{
		Success: true,
		Data:    playlist,
	})
}

// ---- User-scoped endpoints ----

// extractUserID is a helper that pulls the authenticated user ID from the context.
func (h *PlaylistScriptHandler) extractUserID(c *gin.Context) (uuid.UUID, bool) {
	userIDVal, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, StandardResponse{
			Success: false,
			Error:   &ErrorInfo{Code: "UNAUTHORIZED", Message: "Authentication required"},
		})
		return uuid.Nil, false
	}
	userID, ok := userIDVal.(uuid.UUID)
	if !ok || userID == uuid.Nil {
		c.JSON(http.StatusUnauthorized, StandardResponse{
			Success: false,
			Error:   &ErrorInfo{Code: "UNAUTHORIZED", Message: "Authentication required"},
		})
		return uuid.Nil, false
	}
	return userID, true
}

// ListMyScripts handles GET /api/v1/playlist-scripts
func (h *PlaylistScriptHandler) ListMyScripts(c *gin.Context) {
	userID, ok := h.extractUserID(c)
	if !ok {
		return
	}

	scripts, err := h.service.ListUserScripts(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, StandardResponse{
			Success: false,
			Error:   &ErrorInfo{Code: "INTERNAL_ERROR", Message: "Failed to list playlist scripts"},
		})
		return
	}

	c.JSON(http.StatusOK, StandardResponse{Success: true, Data: scripts})
}

// CreateMyScript handles POST /api/v1/playlist-scripts
func (h *PlaylistScriptHandler) CreateMyScript(c *gin.Context) {
	userID, ok := h.extractUserID(c)
	if !ok {
		return
	}

	var req models.CreatePlaylistScriptRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, StandardResponse{
			Success: false,
			Error:   &ErrorInfo{Code: "INVALID_REQUEST", Message: "Invalid request body"},
		})
		return
	}

	script, err := h.service.CreateUserScript(c.Request.Context(), userID, &req)
	if err != nil {
		status := http.StatusInternalServerError
		code, message := "INTERNAL_ERROR", "Failed to create playlist script"
		if errors.Is(err, services.ErrPlaylistScriptValidation) {
			status, code, message = http.StatusBadRequest, "INVALID_REQUEST", "Invalid playlist script"
		}
		c.JSON(status, StandardResponse{
			Success: false,
			Error:   &ErrorInfo{Code: code, Message: message},
		})
		return
	}

	c.JSON(http.StatusCreated, StandardResponse{Success: true, Data: script})
}

// UpdateMyScript handles PUT /api/v1/playlist-scripts/:id
func (h *PlaylistScriptHandler) UpdateMyScript(c *gin.Context) {
	userID, ok := h.extractUserID(c)
	if !ok {
		return
	}

	scriptID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, StandardResponse{
			Success: false,
			Error:   &ErrorInfo{Code: "INVALID_REQUEST", Message: "Invalid script ID"},
		})
		return
	}

	var req models.UpdatePlaylistScriptRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, StandardResponse{
			Success: false,
			Error:   &ErrorInfo{Code: "INVALID_REQUEST", Message: "Invalid request body"},
		})
		return
	}

	script, err := h.service.UpdateUserScript(c.Request.Context(), scriptID, userID, &req)
	if err != nil {
		if errors.Is(err, services.ErrPlaylistScriptNotFound) {
			c.JSON(http.StatusNotFound, StandardResponse{
				Success: false,
				Error:   &ErrorInfo{Code: "NOT_FOUND", Message: "Playlist script not found"},
			})
			return
		}
		if errors.Is(err, services.ErrPlaylistScriptValidation) {
			c.JSON(http.StatusBadRequest, StandardResponse{
				Success: false,
				Error:   &ErrorInfo{Code: "INVALID_REQUEST", Message: "Invalid playlist script"},
			})
			return
		}
		c.JSON(http.StatusInternalServerError, StandardResponse{
			Success: false,
			Error:   &ErrorInfo{Code: "INTERNAL_ERROR", Message: "Failed to update playlist script"},
		})
		return
	}

	c.JSON(http.StatusOK, StandardResponse{Success: true, Data: script})
}

// DeleteMyScript handles DELETE /api/v1/playlist-scripts/:id
func (h *PlaylistScriptHandler) DeleteMyScript(c *gin.Context) {
	userID, ok := h.extractUserID(c)
	if !ok {
		return
	}

	scriptID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, StandardResponse{
			Success: false,
			Error:   &ErrorInfo{Code: "INVALID_REQUEST", Message: "Invalid script ID"},
		})
		return
	}

	if err := h.service.DeleteUserScript(c.Request.Context(), scriptID, userID); err != nil {
		if errors.Is(err, services.ErrPlaylistScriptNotFound) {
			c.JSON(http.StatusNotFound, StandardResponse{
				Success: false,
				Error:   &ErrorInfo{Code: "NOT_FOUND", Message: "Playlist script not found"},
			})
			return
		}
		c.JSON(http.StatusInternalServerError, StandardResponse{
			Success: false,
			Error:   &ErrorInfo{Code: "INTERNAL_ERROR", Message: "Failed to delete playlist script"},
		})
		return
	}

	c.JSON(http.StatusOK, StandardResponse{
		Success: true,
		Data:    map[string]string{"message": "Playlist script deleted successfully"},
	})
}

// GenerateMyPlaylist handles POST /api/v1/playlist-scripts/:id/generate
func (h *PlaylistScriptHandler) GenerateMyPlaylist(c *gin.Context) {
	userID, ok := h.extractUserID(c)
	if !ok {
		return
	}

	scriptID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, StandardResponse{
			Success: false,
			Error:   &ErrorInfo{Code: "INVALID_REQUEST", Message: "Invalid script ID"},
		})
		return
	}

	playlist, err := h.service.GenerateUserPlaylist(c.Request.Context(), scriptID, userID)
	if err != nil {
		if errors.Is(err, services.ErrPlaylistScriptNotFound) {
			c.JSON(http.StatusNotFound, StandardResponse{
				Success: false,
				Error:   &ErrorInfo{Code: "NOT_FOUND", Message: "Playlist script not found"},
			})
			return
		}
		if errors.Is(err, services.ErrPlaylistScriptInactive) {
			c.JSON(http.StatusBadRequest, StandardResponse{
				Success: false,
				Error:   &ErrorInfo{Code: "INVALID_REQUEST", Message: "Playlist script is inactive"},
			})
			return
		}
		if errors.Is(err, services.ErrPlaylistGenerationEmpty) {
			c.JSON(http.StatusConflict, StandardResponse{Success: false, Error: &ErrorInfo{Code: "NO_MATCHING_CLIPS", Message: "No clips matched this playlist script"}})
			return
		}
		c.JSON(http.StatusInternalServerError, StandardResponse{
			Success: false,
			Error:   &ErrorInfo{Code: "INTERNAL_ERROR", Message: "Failed to generate playlist"},
		})
		return
	}

	c.JSON(http.StatusCreated, StandardResponse{Success: true, Data: playlist})
}
