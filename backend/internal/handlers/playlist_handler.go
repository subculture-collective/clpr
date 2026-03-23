package handlers

import (
	"errors"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/subculture-collective/clipper/internal/models"
	"github.com/subculture-collective/clipper/internal/services"
)

// PlaylistHandler handles playlist-related requests
type PlaylistHandler struct {
	playlistService *services.PlaylistService
}

// NewPlaylistHandler creates a new PlaylistHandler
func NewPlaylistHandler(playlistService *services.PlaylistService) *PlaylistHandler {
	return &PlaylistHandler{
		playlistService: playlistService,
	}
}

// CreatePlaylist handles POST /api/playlists
func (h *PlaylistHandler) CreatePlaylist(c *gin.Context) {
	// Get user ID from context (set by auth middleware)
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
	if !ok {
		c.JSON(http.StatusInternalServerError, StandardResponse{
			Success: false,
			Error: &ErrorInfo{
				Code:    "INTERNAL_ERROR",
				Message: "Invalid user ID format",
			},
		})
		return
	}

	// Parse request body
	var req models.CreatePlaylistRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, StandardResponse{
			Success: false,
			Error: &ErrorInfo{
				Code:    "INVALID_REQUEST",
				Message: err.Error(),
			},
		})
		return
	}

	// Create playlist
	playlist, err := h.playlistService.CreatePlaylist(c.Request.Context(), userID, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, StandardResponse{
			Success: false,
			Error: &ErrorInfo{
				Code:    "INTERNAL_ERROR",
				Message: "Failed to create playlist",
			},
		})
		return
	}

	c.JSON(http.StatusCreated, StandardResponse{
		Success: true,
		Data:    playlist,
	})
}

// GetPlaylist handles GET /api/playlists/:id
func (h *PlaylistHandler) GetPlaylist(c *gin.Context) {
	// Parse playlist ID
	playlistIDStr := c.Param("id")
	playlistID, err := uuid.Parse(playlistIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, StandardResponse{
			Success: false,
			Error: &ErrorInfo{
				Code:    "INVALID_REQUEST",
				Message: "Invalid playlist ID",
			},
		})
		return
	}

	// Get optional user ID from context
	var userID *uuid.UUID
	if userIDVal, exists := c.Get("user_id"); exists {
		if uid, ok := userIDVal.(uuid.UUID); ok {
			userID = &uid
		}
	}

	// Parse pagination parameters
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	// Validate and constrain parameters
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	// Get playlist with clips
	playlist, err := h.playlistService.GetPlaylist(c.Request.Context(), playlistID, userID, page, limit)
	if err != nil {
		if err.Error() == "playlist not found" {
			c.JSON(http.StatusNotFound, StandardResponse{
				Success: false,
				Error: &ErrorInfo{
					Code:    "NOT_FOUND",
					Message: "Playlist not found",
				},
			})
			return
		}
		if err.Error() == "unauthorized: playlist is private" {
			c.JSON(http.StatusForbidden, StandardResponse{
				Success: false,
				Error: &ErrorInfo{
					Code:    "FORBIDDEN",
					Message: "This playlist is private",
				},
			})
			return
		}
		c.JSON(http.StatusInternalServerError, StandardResponse{
			Success: false,
			Error: &ErrorInfo{
				Code:    "INTERNAL_ERROR",
				Message: "Failed to get playlist",
			},
		})
		return
	}

	// Build pagination metadata
	totalPages := (playlist.ClipCount + limit - 1) / limit
	meta := PaginationMeta{
		Page:       page,
		Limit:      limit,
		Total:      playlist.ClipCount,
		TotalPages: totalPages,
		HasNext:    page < totalPages,
		HasPrev:    page > 1,
	}

	c.JSON(http.StatusOK, StandardResponse{
		Success: true,
		Data:    playlist,
		Meta:    meta,
	})
}

// GetPlaylistByShareToken handles GET /api/playlists/share/:token
func (h *PlaylistHandler) GetPlaylistByShareToken(c *gin.Context) {
	shareToken := c.Param("token")
	if shareToken == "" {
		c.JSON(http.StatusBadRequest, StandardResponse{
			Success: false,
			Error: &ErrorInfo{
				Code:    "INVALID_REQUEST",
				Message: "Invalid share token",
			},
		})
		return
	}

	// Get optional user ID from context
	var userID *uuid.UUID
	if userIDVal, exists := c.Get("user_id"); exists {
		if uid, ok := userIDVal.(uuid.UUID); ok {
			userID = &uid
		}
	}

	// Parse pagination parameters
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	// Validate and constrain parameters
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	playlist, err := h.playlistService.GetPlaylistByShareToken(c.Request.Context(), shareToken, userID, page, limit)
	if err != nil {
		if err.Error() == "playlist not found" {
			c.JSON(http.StatusNotFound, StandardResponse{
				Success: false,
				Error: &ErrorInfo{
					Code:    "NOT_FOUND",
					Message: "Playlist not found",
				},
			})
			return
		}
		c.JSON(http.StatusInternalServerError, StandardResponse{
			Success: false,
			Error: &ErrorInfo{
				Code:    "INTERNAL_ERROR",
				Message: "Failed to get playlist",
			},
		})
		return
	}

	// Build pagination metadata
	totalPages := (playlist.ClipCount + limit - 1) / limit
	meta := PaginationMeta{
		Page:       page,
		Limit:      limit,
		Total:      playlist.ClipCount,
		TotalPages: totalPages,
		HasNext:    page < totalPages,
		HasPrev:    page > 1,
	}

	c.JSON(http.StatusOK, StandardResponse{
		Success: true,
		Data:    playlist,
		Meta:    meta,
	})
}

// UpdatePlaylist handles PATCH /api/playlists/:id
func (h *PlaylistHandler) UpdatePlaylist(c *gin.Context) {
	// Get user ID from context
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
	if !ok {
		c.JSON(http.StatusInternalServerError, StandardResponse{
			Success: false,
			Error: &ErrorInfo{
				Code:    "INTERNAL_ERROR",
				Message: "Invalid user ID format",
			},
		})
		return
	}

	// Parse playlist ID
	playlistIDStr := c.Param("id")
	playlistID, err := uuid.Parse(playlistIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, StandardResponse{
			Success: false,
			Error: &ErrorInfo{
				Code:    "INVALID_REQUEST",
				Message: "Invalid playlist ID",
			},
		})
		return
	}

	// Parse request body
	var req models.UpdatePlaylistRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, StandardResponse{
			Success: false,
			Error: &ErrorInfo{
				Code:    "INVALID_REQUEST",
				Message: err.Error(),
			},
		})
		return
	}

	// Update playlist
	playlist, err := h.playlistService.UpdatePlaylist(c.Request.Context(), playlistID, userID, &req)
	if err != nil {
		if err.Error() == "playlist not found" {
			c.JSON(http.StatusNotFound, StandardResponse{
				Success: false,
				Error: &ErrorInfo{
					Code:    "NOT_FOUND",
					Message: "Playlist not found",
				},
			})
			return
		}
		// Handle permission errors (both owner-only and collaborator checks)
		if strings.Contains(err.Error(), "unauthorized") {
			c.JSON(http.StatusForbidden, StandardResponse{
				Success: false,
				Error: &ErrorInfo{
					Code:    "FORBIDDEN",
					Message: "You don't have permission to edit this playlist",
				},
			})
			return
		}
		c.JSON(http.StatusInternalServerError, StandardResponse{
			Success: false,
			Error: &ErrorInfo{
				Code:    "INTERNAL_ERROR",
				Message: "Failed to update playlist",
			},
		})
		return
	}

	c.JSON(http.StatusOK, StandardResponse{
		Success: true,
		Data:    playlist,
	})
}

// CopyPlaylist handles POST /api/playlists/:id/copy
func (h *PlaylistHandler) CopyPlaylist(c *gin.Context) {
	// Get user ID from context
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
	if !ok {
		c.JSON(http.StatusInternalServerError, StandardResponse{
			Success: false,
			Error: &ErrorInfo{
				Code:    "INTERNAL_ERROR",
				Message: "Invalid user ID format",
			},
		})
		return
	}

	// Parse playlist ID
	playlistIDStr := c.Param("id")
	playlistID, err := uuid.Parse(playlistIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, StandardResponse{
			Success: false,
			Error: &ErrorInfo{
				Code:    "INVALID_REQUEST",
				Message: "Invalid playlist ID",
			},
		})
		return
	}

	// Parse request body (optional)
	var req models.CopyPlaylistRequest
	if err := c.ShouldBindJSON(&req); err != nil && !errors.Is(err, io.EOF) {
		c.JSON(http.StatusBadRequest, StandardResponse{
			Success: false,
			Error: &ErrorInfo{
				Code:    "INVALID_REQUEST",
				Message: err.Error(),
			},
		})
		return
	}

	playlist, err := h.playlistService.CopyPlaylist(c.Request.Context(), playlistID, userID, &req)
	if err != nil {
		if err.Error() == "playlist not found" {
			c.JSON(http.StatusNotFound, StandardResponse{
				Success: false,
				Error: &ErrorInfo{
					Code:    "NOT_FOUND",
					Message: "Playlist not found",
				},
			})
			return
		}
		if strings.Contains(err.Error(), "unauthorized") {
			c.JSON(http.StatusForbidden, StandardResponse{
				Success: false,
				Error: &ErrorInfo{
					Code:    "FORBIDDEN",
					Message: "You don't have permission to copy this playlist",
				},
			})
			return
		}
		c.JSON(http.StatusInternalServerError, StandardResponse{
			Success: false,
			Error: &ErrorInfo{
				Code:    "INTERNAL_ERROR",
				Message: "Failed to copy playlist",
			},
		})
		return
	}

	c.JSON(http.StatusCreated, StandardResponse{
		Success: true,
		Data:    playlist,
	})
}

// DeletePlaylist handles DELETE /api/playlists/:id
func (h *PlaylistHandler) DeletePlaylist(c *gin.Context) {
	// Get user ID from context
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
	if !ok {
		c.JSON(http.StatusInternalServerError, StandardResponse{
			Success: false,
			Error: &ErrorInfo{
				Code:    "INTERNAL_ERROR",
				Message: "Invalid user ID format",
			},
		})
		return
	}

	// Parse playlist ID
	playlistIDStr := c.Param("id")
	playlistID, err := uuid.Parse(playlistIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, StandardResponse{
			Success: false,
			Error: &ErrorInfo{
				Code:    "INVALID_REQUEST",
				Message: "Invalid playlist ID",
			},
		})
		return
	}

	// Delete playlist
	err = h.playlistService.DeletePlaylist(c.Request.Context(), playlistID, userID)
	if err != nil {
		if err.Error() == "playlist not found" {
			c.JSON(http.StatusNotFound, StandardResponse{
				Success: false,
				Error: &ErrorInfo{
					Code:    "NOT_FOUND",
					Message: "Playlist not found",
				},
			})
			return
		}
		if err.Error() == "unauthorized: user does not own this playlist" {
			c.JSON(http.StatusForbidden, StandardResponse{
				Success: false,
				Error: &ErrorInfo{
					Code:    "FORBIDDEN",
					Message: "You don't have permission to delete this playlist",
				},
			})
			return
		}
		c.JSON(http.StatusInternalServerError, StandardResponse{
			Success: false,
			Error: &ErrorInfo{
				Code:    "INTERNAL_ERROR",
				Message: "Failed to delete playlist",
			},
		})
		return
	}

	c.JSON(http.StatusOK, StandardResponse{
		Success: true,
		Data: map[string]string{
			"message": "Playlist deleted successfully",
		},
	})
}

// ListUserPlaylists handles GET /api/playlists
func (h *PlaylistHandler) ListUserPlaylists(c *gin.Context) {
	// Get user ID from context
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
	if !ok {
		c.JSON(http.StatusInternalServerError, StandardResponse{
			Success: false,
			Error: &ErrorInfo{
				Code:    "INTERNAL_ERROR",
				Message: "Invalid user ID format",
			},
		})
		return
	}

	// Parse pagination parameters
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	// Validate and constrain parameters
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	// Get user's playlists
	playlists, total, err := h.playlistService.ListUserPlaylists(c.Request.Context(), userID, page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, StandardResponse{
			Success: false,
			Error: &ErrorInfo{
				Code:    "INTERNAL_ERROR",
				Message: "Failed to list playlists",
			},
		})
		return
	}

	// Build pagination metadata
	totalPages := (total + limit - 1) / limit
	meta := PaginationMeta{
		Page:       page,
		Limit:      limit,
		Total:      total,
		TotalPages: totalPages,
		HasNext:    page < totalPages,
		HasPrev:    page > 1,
	}

	c.JSON(http.StatusOK, StandardResponse{
		Success: true,
		Data:    playlists,
		Meta:    meta,
	})
}

// ListPublicPlaylists handles GET /api/playlists/public
func (h *PlaylistHandler) ListPublicPlaylists(c *gin.Context) {
	var userID *uuid.UUID
	if userIDVal, exists := c.Get("user_id"); exists {
		if uid, ok := userIDVal.(uuid.UUID); ok {
			userID = &uid
		}
	}

	// Parse pagination parameters
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	// Validate and constrain parameters
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	// Get public playlists
	playlists, total, err := h.playlistService.ListPublicPlaylists(c.Request.Context(), userID, page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, StandardResponse{
			Success: false,
			Error: &ErrorInfo{
				Code:    "INTERNAL_ERROR",
				Message: "Failed to list public playlists",
			},
		})
		return
	}

	// Build pagination metadata
	totalPages := (total + limit - 1) / limit
	meta := PaginationMeta{
		Page:       page,
		Limit:      limit,
		Total:      total,
		TotalPages: totalPages,
		HasNext:    page < totalPages,
		HasPrev:    page > 1,
	}

	c.JSON(http.StatusOK, StandardResponse{
		Success: true,
		Data:    playlists,
		Meta:    meta,
	})
}

// ListBookmarkedPlaylists handles GET /api/playlists/bookmarks
func (h *PlaylistHandler) ListBookmarkedPlaylists(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)

	// Parse pagination parameters
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	// Validate and constrain parameters
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	// Get bookmarked playlists
	playlists, total, err := h.playlistService.ListBookmarkedPlaylists(c.Request.Context(), userID, page, limit)
	if err != nil {
		log.Printf("ERROR: ListBookmarkedPlaylists failed for user %s: %v", userID, err)
		c.JSON(http.StatusInternalServerError, StandardResponse{
			Success: false,
			Error: &ErrorInfo{
				Code:    "INTERNAL_ERROR",
				Message: "Failed to list bookmarked playlists",
			},
		})
		return
	}

	// Build pagination metadata
	totalPages := (total + limit - 1) / limit
	meta := PaginationMeta{
		Page:       page,
		Limit:      limit,
		Total:      total,
		TotalPages: totalPages,
		HasNext:    page < totalPages,
		HasPrev:    page > 1,
	}

	c.JSON(http.StatusOK, StandardResponse{
		Success: true,
		Data:    playlists,
		Meta:    meta,
	})
}

// AddClipsToPlaylist handles POST /api/playlists/:id/clips
func (h *PlaylistHandler) AddClipsToPlaylist(c *gin.Context) {
	// Get user ID from context
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
	if !ok {
		c.JSON(http.StatusInternalServerError, StandardResponse{
			Success: false,
			Error: &ErrorInfo{
				Code:    "INTERNAL_ERROR",
				Message: "Invalid user ID format",
			},
		})
		return
	}

	// Parse playlist ID
	playlistIDStr := c.Param("id")
	playlistID, err := uuid.Parse(playlistIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, StandardResponse{
			Success: false,
			Error: &ErrorInfo{
				Code:    "INVALID_REQUEST",
				Message: "Invalid playlist ID",
			},
		})
		return
	}

	// Parse request body
	var req models.AddClipsToPlaylistRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, StandardResponse{
			Success: false,
			Error: &ErrorInfo{
				Code:    "INVALID_REQUEST",
				Message: err.Error(),
			},
		})
		return
	}

	// Add clips to playlist
	err = h.playlistService.AddClipsToPlaylist(c.Request.Context(), playlistID, userID, req.ClipIDs)
	if err != nil {
		if err.Error() == "playlist not found" {
			c.JSON(http.StatusNotFound, StandardResponse{
				Success: false,
				Error: &ErrorInfo{
					Code:    "NOT_FOUND",
					Message: "Playlist not found",
				},
			})
			return
		}
		if strings.Contains(err.Error(), "unauthorized") {
			c.JSON(http.StatusForbidden, StandardResponse{
				Success: false,
				Error: &ErrorInfo{
					Code:    "FORBIDDEN",
					Message: "You don't have permission to edit this playlist",
				},
			})
			return
		}
		if err.Error() == "playlist cannot exceed 1000 clips" {
			c.JSON(http.StatusBadRequest, StandardResponse{
				Success: false,
				Error: &ErrorInfo{
					Code:    "LIMIT_EXCEEDED",
					Message: "Playlist cannot exceed 1000 clips",
				},
			})
			return
		}
		c.JSON(http.StatusInternalServerError, StandardResponse{
			Success: false,
			Error: &ErrorInfo{
				Code:    "INTERNAL_ERROR",
				Message: "Failed to add clips to playlist",
			},
		})
		return
	}

	c.JSON(http.StatusOK, StandardResponse{
		Success: true,
		Data: map[string]string{
			"message": "Clips added to playlist successfully",
		},
	})
}

// RemoveClipFromPlaylist handles DELETE /api/playlists/:id/clips/:clip_id
func (h *PlaylistHandler) RemoveClipFromPlaylist(c *gin.Context) {
	// Get user ID from context
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
	if !ok {
		c.JSON(http.StatusInternalServerError, StandardResponse{
			Success: false,
			Error: &ErrorInfo{
				Code:    "INTERNAL_ERROR",
				Message: "Invalid user ID format",
			},
		})
		return
	}

	// Parse playlist ID
	playlistIDStr := c.Param("id")
	playlistID, err := uuid.Parse(playlistIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, StandardResponse{
			Success: false,
			Error: &ErrorInfo{
				Code:    "INVALID_REQUEST",
				Message: "Invalid playlist ID",
			},
		})
		return
	}

	// Parse clip ID
	clipIDStr := c.Param("clip_id")
	clipID, err := uuid.Parse(clipIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, StandardResponse{
			Success: false,
			Error: &ErrorInfo{
				Code:    "INVALID_REQUEST",
				Message: "Invalid clip ID",
			},
		})
		return
	}

	// Remove clip from playlist
	err = h.playlistService.RemoveClipFromPlaylist(c.Request.Context(), playlistID, clipID, userID)
	if err != nil {
		if err.Error() == "playlist not found" {
			c.JSON(http.StatusNotFound, StandardResponse{
				Success: false,
				Error: &ErrorInfo{
					Code:    "NOT_FOUND",
					Message: "Playlist not found",
				},
			})
			return
		}
		if strings.Contains(err.Error(), "unauthorized") {
			c.JSON(http.StatusForbidden, StandardResponse{
				Success: false,
				Error: &ErrorInfo{
					Code:    "FORBIDDEN",
					Message: "You don't have permission to edit this playlist",
				},
			})
			return
		}
		if err.Error() == "clip not found in playlist" {
			c.JSON(http.StatusNotFound, StandardResponse{
				Success: false,
				Error: &ErrorInfo{
					Code:    "NOT_FOUND",
					Message: "Clip not found in playlist",
				},
			})
			return
		}
		c.JSON(http.StatusInternalServerError, StandardResponse{
			Success: false,
			Error: &ErrorInfo{
				Code:    "INTERNAL_ERROR",
				Message: "Failed to remove clip from playlist",
			},
		})
		return
	}

	c.JSON(http.StatusOK, StandardResponse{
		Success: true,
		Data: map[string]string{
			"message": "Clip removed from playlist successfully",
		},
	})
}

// ReorderPlaylistClips handles PUT /api/playlists/:id/clips/order
func (h *PlaylistHandler) ReorderPlaylistClips(c *gin.Context) {
	// Get user ID from context
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
	if !ok {
		c.JSON(http.StatusInternalServerError, StandardResponse{
			Success: false,
			Error: &ErrorInfo{
				Code:    "INTERNAL_ERROR",
				Message: "Invalid user ID format",
			},
		})
		return
	}

	// Parse playlist ID
	playlistIDStr := c.Param("id")
	playlistID, err := uuid.Parse(playlistIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, StandardResponse{
			Success: false,
			Error: &ErrorInfo{
				Code:    "INVALID_REQUEST",
				Message: "Invalid playlist ID",
			},
		})
		return
	}

	// Parse request body
	var req models.ReorderPlaylistClipsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, StandardResponse{
			Success: false,
			Error: &ErrorInfo{
				Code:    "INVALID_REQUEST",
				Message: err.Error(),
			},
		})
		return
	}

	// Reorder clips
	err = h.playlistService.ReorderPlaylistClips(c.Request.Context(), playlistID, userID, req.ClipIDs)
	if err != nil {
		if err.Error() == "playlist not found" {
			c.JSON(http.StatusNotFound, StandardResponse{
				Success: false,
				Error: &ErrorInfo{
					Code:    "NOT_FOUND",
					Message: "Playlist not found",
				},
			})
			return
		}
		if strings.Contains(err.Error(), "unauthorized") {
			c.JSON(http.StatusForbidden, StandardResponse{
				Success: false,
				Error: &ErrorInfo{
					Code:    "FORBIDDEN",
					Message: "You don't have permission to edit this playlist",
				},
			})
			return
		}
		c.JSON(http.StatusInternalServerError, StandardResponse{
			Success: false,
			Error: &ErrorInfo{
				Code:    "INTERNAL_ERROR",
				Message: "Failed to reorder clips",
			},
		})
		return
	}

	c.JSON(http.StatusOK, StandardResponse{
		Success: true,
		Data: map[string]string{
			"message": "Clips reordered successfully",
		},
	})
}

// LikePlaylist handles POST /api/playlists/:id/like
func (h *PlaylistHandler) LikePlaylist(c *gin.Context) {
	// Get user ID from context
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
	if !ok {
		c.JSON(http.StatusInternalServerError, StandardResponse{
			Success: false,
			Error: &ErrorInfo{
				Code:    "INTERNAL_ERROR",
				Message: "Invalid user ID format",
			},
		})
		return
	}

	// Parse playlist ID
	playlistIDStr := c.Param("id")
	playlistID, err := uuid.Parse(playlistIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, StandardResponse{
			Success: false,
			Error: &ErrorInfo{
				Code:    "INVALID_REQUEST",
				Message: "Invalid playlist ID",
			},
		})
		return
	}

	// Like playlist
	err = h.playlistService.LikePlaylist(c.Request.Context(), playlistID, userID)
	if err != nil {
		if err.Error() == "playlist not found" {
			c.JSON(http.StatusNotFound, StandardResponse{
				Success: false,
				Error: &ErrorInfo{
					Code:    "NOT_FOUND",
					Message: "Playlist not found",
				},
			})
			return
		}
		if err.Error() == "cannot like private playlists" {
			c.JSON(http.StatusForbidden, StandardResponse{
				Success: false,
				Error: &ErrorInfo{
					Code:    "FORBIDDEN",
					Message: "Cannot like private playlists",
				},
			})
			return
		}
		c.JSON(http.StatusInternalServerError, StandardResponse{
			Success: false,
			Error: &ErrorInfo{
				Code:    "INTERNAL_ERROR",
				Message: "Failed to like playlist",
			},
		})
		return
	}

	c.JSON(http.StatusOK, StandardResponse{
		Success: true,
		Data: map[string]string{
			"message": "Playlist liked successfully",
		},
	})
}

// UnlikePlaylist handles DELETE /api/playlists/:id/like
func (h *PlaylistHandler) UnlikePlaylist(c *gin.Context) {
	// Get user ID from context
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
	if !ok {
		c.JSON(http.StatusInternalServerError, StandardResponse{
			Success: false,
			Error: &ErrorInfo{
				Code:    "INTERNAL_ERROR",
				Message: "Invalid user ID format",
			},
		})
		return
	}

	// Parse playlist ID
	playlistIDStr := c.Param("id")
	playlistID, err := uuid.Parse(playlistIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, StandardResponse{
			Success: false,
			Error: &ErrorInfo{
				Code:    "INVALID_REQUEST",
				Message: "Invalid playlist ID",
			},
		})
		return
	}

	// Unlike playlist
	err = h.playlistService.UnlikePlaylist(c.Request.Context(), playlistID, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, StandardResponse{
			Success: false,
			Error: &ErrorInfo{
				Code:    "INTERNAL_ERROR",
				Message: "Failed to unlike playlist",
			},
		})
		return
	}

	c.JSON(http.StatusOK, StandardResponse{
		Success: true,
		Data: map[string]string{
			"message": "Playlist unliked successfully",
		},
	})
}

// GetShareLink handles GET /api/playlists/:id/share-link
func (h *PlaylistHandler) GetShareLink(c *gin.Context) {
	// Get user ID from context
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
	if !ok {
		c.JSON(http.StatusInternalServerError, StandardResponse{
			Success: false,
			Error: &ErrorInfo{
				Code:    "INTERNAL_ERROR",
				Message: "Invalid user ID format",
			},
		})
		return
	}

	// Parse playlist ID
	playlistIDStr := c.Param("id")
	playlistID, err := uuid.Parse(playlistIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, StandardResponse{
			Success: false,
			Error: &ErrorInfo{
				Code:    "INVALID_REQUEST",
				Message: "Invalid playlist ID",
			},
		})
		return
	}

	// Get share link
	shareLink, err := h.playlistService.GetShareLink(c.Request.Context(), playlistID, userID)
	if err != nil {
		if err.Error() == "playlist not found" {
			c.JSON(http.StatusNotFound, StandardResponse{
				Success: false,
				Error: &ErrorInfo{
					Code:    "NOT_FOUND",
					Message: "Playlist not found",
				},
			})
			return
		}
		if err.Error() == "unauthorized: user does not have permission to share this playlist" {
			c.JSON(http.StatusForbidden, StandardResponse{
				Success: false,
				Error: &ErrorInfo{
					Code:    "FORBIDDEN",
					Message: "You don't have permission to share this playlist",
				},
			})
			return
		}
		c.JSON(http.StatusInternalServerError, StandardResponse{
			Success: false,
			Error: &ErrorInfo{
				Code:    "INTERNAL_ERROR",
				Message: "Failed to get share link",
			},
		})
		return
	}

	c.JSON(http.StatusOK, StandardResponse{
		Success: true,
		Data:    shareLink,
	})
}

// TrackShare handles POST /api/playlists/:id/track-share
func (h *PlaylistHandler) TrackShare(c *gin.Context) {
	// Parse playlist ID
	playlistIDStr := c.Param("id")
	playlistID, err := uuid.Parse(playlistIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, StandardResponse{
			Success: false,
			Error: &ErrorInfo{
				Code:    "INVALID_REQUEST",
				Message: "Invalid playlist ID",
			},
		})
		return
	}

	// Parse request body
	var req models.TrackShareRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, StandardResponse{
			Success: false,
			Error: &ErrorInfo{
				Code:    "INVALID_REQUEST",
				Message: err.Error(),
			},
		})
		return
	}

	// Track share
	referrer := ""
	if req.Referrer != nil {
		referrer = *req.Referrer
	}
	err = h.playlistService.TrackShare(c.Request.Context(), playlistID, req.Platform, referrer)
	if err != nil {
		if err.Error() == "playlist not found" {
			c.JSON(http.StatusNotFound, StandardResponse{
				Success: false,
				Error: &ErrorInfo{
					Code:    "NOT_FOUND",
					Message: "Playlist not found",
				},
			})
			return
		}
		c.JSON(http.StatusInternalServerError, StandardResponse{
			Success: false,
			Error: &ErrorInfo{
				Code:    "INTERNAL_ERROR",
				Message: "Failed to track share",
			},
		})
		return
	}

	c.JSON(http.StatusOK, StandardResponse{
		Success: true,
		Data: map[string]string{
			"message": "Share tracked successfully",
		},
	})
}

// AddCollaborator handles POST /api/playlists/:id/collaborators
func (h *PlaylistHandler) AddCollaborator(c *gin.Context) {
	// Get user ID from context
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
	if !ok {
		c.JSON(http.StatusInternalServerError, StandardResponse{
			Success: false,
			Error: &ErrorInfo{
				Code:    "INTERNAL_ERROR",
				Message: "Invalid user ID format",
			},
		})
		return
	}

	// Parse playlist ID
	playlistIDStr := c.Param("id")
	playlistID, err := uuid.Parse(playlistIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, StandardResponse{
			Success: false,
			Error: &ErrorInfo{
				Code:    "INVALID_REQUEST",
				Message: "Invalid playlist ID",
			},
		})
		return
	}

	// Parse request body
	var req models.AddCollaboratorRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, StandardResponse{
			Success: false,
			Error: &ErrorInfo{
				Code:    "INVALID_REQUEST",
				Message: err.Error(),
			},
		})
		return
	}

	// Parse collaborator user ID
	collaboratorUserID, err := uuid.Parse(req.UserID)
	if err != nil {
		c.JSON(http.StatusBadRequest, StandardResponse{
			Success: false,
			Error: &ErrorInfo{
				Code:    "INVALID_REQUEST",
				Message: "Invalid collaborator user ID",
			},
		})
		return
	}

	// Add collaborator
	err = h.playlistService.AddCollaborator(c.Request.Context(), playlistID, userID, collaboratorUserID, req.Permission)
	if err != nil {
		if err.Error() == "playlist not found" {
			c.JSON(http.StatusNotFound, StandardResponse{
				Success: false,
				Error: &ErrorInfo{
					Code:    "NOT_FOUND",
					Message: "Playlist not found",
				},
			})
			return
		}
		if err.Error() == "unauthorized: user does not have permission to add collaborators" {
			c.JSON(http.StatusForbidden, StandardResponse{
				Success: false,
				Error: &ErrorInfo{
					Code:    "FORBIDDEN",
					Message: "You don't have permission to add collaborators",
				},
			})
			return
		}
		if err.Error() == "cannot add playlist owner as a collaborator" {
			c.JSON(http.StatusBadRequest, StandardResponse{
				Success: false,
				Error: &ErrorInfo{
					Code:    "INVALID_REQUEST",
					Message: "Cannot add playlist owner as a collaborator",
				},
			})
			return
		}
		c.JSON(http.StatusInternalServerError, StandardResponse{
			Success: false,
			Error: &ErrorInfo{
				Code:    "INTERNAL_ERROR",
				Message: "Failed to add collaborator",
			},
		})
		return
	}

	c.JSON(http.StatusCreated, StandardResponse{
		Success: true,
		Data: map[string]string{
			"message": "Collaborator added successfully",
		},
	})
}

// GetCollaborators handles GET /api/playlists/:id/collaborators
func (h *PlaylistHandler) GetCollaborators(c *gin.Context) {
	// Get user ID from context (optional for public playlists)
	var userID *uuid.UUID
	if userIDVal, exists := c.Get("user_id"); exists {
		if uid, ok := userIDVal.(uuid.UUID); ok {
			userID = &uid
		}
	}

	// Parse playlist ID
	playlistIDStr := c.Param("id")
	playlistID, err := uuid.Parse(playlistIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, StandardResponse{
			Success: false,
			Error: &ErrorInfo{
				Code:    "INVALID_REQUEST",
				Message: "Invalid playlist ID",
			},
		})
		return
	}

	// Get collaborators
	var collaborators []*models.PlaylistCollaborator
	if userID != nil {
		collaborators, err = h.playlistService.GetCollaborators(c.Request.Context(), playlistID, *userID)
	} else {
		// For anonymous users, try to get collaborators (will only work for public playlists)
		collaborators, err = h.playlistService.GetCollaborators(c.Request.Context(), playlistID, uuid.Nil)
	}

	if err != nil {
		if err.Error() == "playlist not found" {
			c.JSON(http.StatusNotFound, StandardResponse{
				Success: false,
				Error: &ErrorInfo{
					Code:    "NOT_FOUND",
					Message: "Playlist not found",
				},
			})
			return
		}
		if err.Error() == "unauthorized: user does not have permission to view collaborators" {
			c.JSON(http.StatusForbidden, StandardResponse{
				Success: false,
				Error: &ErrorInfo{
					Code:    "FORBIDDEN",
					Message: "You don't have permission to view collaborators",
				},
			})
			return
		}
		c.JSON(http.StatusInternalServerError, StandardResponse{
			Success: false,
			Error: &ErrorInfo{
				Code:    "INTERNAL_ERROR",
				Message: "Failed to get collaborators",
			},
		})
		return
	}

	c.JSON(http.StatusOK, StandardResponse{
		Success: true,
		Data:    collaborators,
	})
}

// RemoveCollaborator handles DELETE /api/playlists/:id/collaborators/:user_id
func (h *PlaylistHandler) RemoveCollaborator(c *gin.Context) {
	// Get user ID from context
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
	if !ok {
		c.JSON(http.StatusInternalServerError, StandardResponse{
			Success: false,
			Error: &ErrorInfo{
				Code:    "INTERNAL_ERROR",
				Message: "Invalid user ID format",
			},
		})
		return
	}

	// Parse playlist ID
	playlistIDStr := c.Param("id")
	playlistID, err := uuid.Parse(playlistIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, StandardResponse{
			Success: false,
			Error: &ErrorInfo{
				Code:    "INVALID_REQUEST",
				Message: "Invalid playlist ID",
			},
		})
		return
	}

	// Parse collaborator user ID
	collaboratorUserIDStr := c.Param("user_id")
	collaboratorUserID, err := uuid.Parse(collaboratorUserIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, StandardResponse{
			Success: false,
			Error: &ErrorInfo{
				Code:    "INVALID_REQUEST",
				Message: "Invalid collaborator user ID",
			},
		})
		return
	}

	// Remove collaborator
	err = h.playlistService.RemoveCollaborator(c.Request.Context(), playlistID, userID, collaboratorUserID)
	if err != nil {
		if err.Error() == "playlist not found" {
			c.JSON(http.StatusNotFound, StandardResponse{
				Success: false,
				Error: &ErrorInfo{
					Code:    "NOT_FOUND",
					Message: "Playlist not found",
				},
			})
			return
		}
		if err.Error() == "unauthorized: user does not have permission to remove collaborators" {
			c.JSON(http.StatusForbidden, StandardResponse{
				Success: false,
				Error: &ErrorInfo{
					Code:    "FORBIDDEN",
					Message: "You don't have permission to remove collaborators",
				},
			})
			return
		}
		if err.Error() == "collaborator not found" {
			c.JSON(http.StatusNotFound, StandardResponse{
				Success: false,
				Error: &ErrorInfo{
					Code:    "NOT_FOUND",
					Message: "Collaborator not found",
				},
			})
			return
		}
		c.JSON(http.StatusInternalServerError, StandardResponse{
			Success: false,
			Error: &ErrorInfo{
				Code:    "INTERNAL_ERROR",
				Message: "Failed to remove collaborator",
			},
		})
		return
	}

	c.JSON(http.StatusOK, StandardResponse{
		Success: true,
		Data: map[string]string{
			"message": "Collaborator removed successfully",
		},
	})
}

// UpdateCollaboratorPermission handles PATCH /api/playlists/:id/collaborators/:user_id
func (h *PlaylistHandler) UpdateCollaboratorPermission(c *gin.Context) {
	// Get user ID from context
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
	if !ok {
		c.JSON(http.StatusInternalServerError, StandardResponse{
			Success: false,
			Error: &ErrorInfo{
				Code:    "INTERNAL_ERROR",
				Message: "Invalid user ID format",
			},
		})
		return
	}

	// Parse playlist ID
	playlistIDStr := c.Param("id")
	playlistID, err := uuid.Parse(playlistIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, StandardResponse{
			Success: false,
			Error: &ErrorInfo{
				Code:    "INVALID_REQUEST",
				Message: "Invalid playlist ID",
			},
		})
		return
	}

	// Parse collaborator user ID
	collaboratorUserIDStr := c.Param("user_id")
	collaboratorUserID, err := uuid.Parse(collaboratorUserIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, StandardResponse{
			Success: false,
			Error: &ErrorInfo{
				Code:    "INVALID_REQUEST",
				Message: "Invalid collaborator user ID",
			},
		})
		return
	}

	// Parse request body
	var req models.UpdateCollaboratorRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, StandardResponse{
			Success: false,
			Error: &ErrorInfo{
				Code:    "INVALID_REQUEST",
				Message: err.Error(),
			},
		})
		return
	}

	// Update collaborator permission
	err = h.playlistService.UpdateCollaboratorPermission(c.Request.Context(), playlistID, userID, collaboratorUserID, req.Permission)
	if err != nil {
		if err.Error() == "playlist not found" {
			c.JSON(http.StatusNotFound, StandardResponse{
				Success: false,
				Error: &ErrorInfo{
					Code:    "NOT_FOUND",
					Message: "Playlist not found",
				},
			})
			return
		}
		if err.Error() == "unauthorized: user does not have permission to update collaborators" {
			c.JSON(http.StatusForbidden, StandardResponse{
				Success: false,
				Error: &ErrorInfo{
					Code:    "FORBIDDEN",
					Message: "You don't have permission to update collaborators",
				},
			})
			return
		}
		c.JSON(http.StatusInternalServerError, StandardResponse{
			Success: false,
			Error: &ErrorInfo{
				Code:    "INTERNAL_ERROR",
				Message: "Failed to update collaborator permission",
			},
		})
		return
	}

	c.JSON(http.StatusOK, StandardResponse{
		Success: true,
		Data: map[string]string{
			"message": "Collaborator permission updated successfully",
		},
	})
}

// ListFeaturedPlaylists handles GET /api/v1/playlists/featured
func (h *PlaylistHandler) ListFeaturedPlaylists(c *gin.Context) {
	var userID *uuid.UUID
	if userIDVal, exists := c.Get("user_id"); exists {
		if uid, ok := userIDVal.(uuid.UUID); ok {
			userID = &uid
		}
	}

	page := 1
	limit := 20

	if p := c.Query("page"); p != "" {
		if val, err := strconv.Atoi(p); err == nil && val > 0 {
			page = val
		}
	}
	if l := c.Query("limit"); l != "" {
		if val, err := strconv.Atoi(l); err == nil && val > 0 && val <= 100 {
			limit = val
		}
	}

	playlists, total, err := h.playlistService.ListFeaturedPlaylists(c.Request.Context(), userID, page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, StandardResponse{
			Success: false,
			Error: &ErrorInfo{
				Code:    "INTERNAL_ERROR",
				Message: "Failed to list featured playlists",
			},
		})
		return
	}

	totalPages := (total + limit - 1) / limit
	meta := PaginationMeta{
		Page:       page,
		Limit:      limit,
		Total:      total,
		TotalPages: totalPages,
		HasNext:    page < totalPages,
		HasPrev:    page > 1,
	}

	c.JSON(http.StatusOK, StandardResponse{
		Success: true,
		Data:    playlists,
		Meta:    meta,
	})
}

// GetPlaylistOfTheDay handles GET /api/v1/playlists/today
func (h *PlaylistHandler) GetPlaylistOfTheDay(c *gin.Context) {
	var userID *uuid.UUID
	if userIDVal, exists := c.Get("user_id"); exists {
		if uid, ok := userIDVal.(uuid.UUID); ok {
			userID = &uid
		}
	}

	playlist, err := h.playlistService.GetPlaylistOfTheDay(c.Request.Context(), userID)
	if err != nil {
		if strings.Contains(err.Error(), "no rows") {
			c.JSON(http.StatusNotFound, StandardResponse{
				Success: false,
				Error: &ErrorInfo{
					Code:    "NOT_FOUND",
					Message: "No playlist of the day available",
				},
			})
			return
		}
		c.JSON(http.StatusInternalServerError, StandardResponse{
			Success: false,
			Error: &ErrorInfo{
				Code:    "INTERNAL_ERROR",
				Message: "Failed to get playlist of the day",
			},
		})
		return
	}

	c.JSON(http.StatusOK, StandardResponse{
		Success: true,
		Data:    playlist,
	})
}

// BookmarkPlaylist handles POST /api/playlists/:id/bookmark
func (h *PlaylistHandler) BookmarkPlaylist(c *gin.Context) {
	userIDVal, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, StandardResponse{
			Success: false,
			Error: &ErrorInfo{Code: "UNAUTHORIZED", Message: "Authentication required"},
		})
		return
	}

	userID, ok := userIDVal.(uuid.UUID)
	if !ok {
		c.JSON(http.StatusInternalServerError, StandardResponse{
			Success: false,
			Error: &ErrorInfo{Code: "INTERNAL_ERROR", Message: "Invalid user ID format"},
		})
		return
	}

	playlistID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, StandardResponse{
			Success: false,
			Error: &ErrorInfo{Code: "INVALID_REQUEST", Message: "Invalid playlist ID"},
		})
		return
	}

	err = h.playlistService.BookmarkPlaylist(c.Request.Context(), playlistID, userID)
	if err != nil {
		if err.Error() == "playlist not found" {
			c.JSON(http.StatusNotFound, StandardResponse{
				Success: false,
				Error: &ErrorInfo{Code: "NOT_FOUND", Message: "Playlist not found"},
			})
			return
		}
		if err.Error() == "cannot bookmark private playlists" {
			c.JSON(http.StatusForbidden, StandardResponse{
				Success: false,
				Error: &ErrorInfo{Code: "FORBIDDEN", Message: "Cannot bookmark private playlists"},
			})
			return
		}
		c.JSON(http.StatusInternalServerError, StandardResponse{
			Success: false,
			Error: &ErrorInfo{Code: "INTERNAL_ERROR", Message: "Failed to bookmark playlist"},
		})
		return
	}

	c.JSON(http.StatusOK, StandardResponse{
		Success: true,
		Data: map[string]string{"message": "Playlist bookmarked successfully"},
	})
}

// UnbookmarkPlaylist handles DELETE /api/playlists/:id/bookmark
func (h *PlaylistHandler) UnbookmarkPlaylist(c *gin.Context) {
	userIDVal, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, StandardResponse{
			Success: false,
			Error: &ErrorInfo{Code: "UNAUTHORIZED", Message: "Authentication required"},
		})
		return
	}

	userID, ok := userIDVal.(uuid.UUID)
	if !ok {
		c.JSON(http.StatusInternalServerError, StandardResponse{
			Success: false,
			Error: &ErrorInfo{Code: "INTERNAL_ERROR", Message: "Invalid user ID format"},
		})
		return
	}

	playlistID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, StandardResponse{
			Success: false,
			Error: &ErrorInfo{Code: "INVALID_REQUEST", Message: "Invalid playlist ID"},
		})
		return
	}

	err = h.playlistService.UnbookmarkPlaylist(c.Request.Context(), playlistID, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, StandardResponse{
			Success: false,
			Error: &ErrorInfo{Code: "INTERNAL_ERROR", Message: "Failed to unbookmark playlist"},
		})
		return
	}

	c.JSON(http.StatusOK, StandardResponse{
		Success: true,
		Data: map[string]string{"message": "Playlist unbookmarked successfully"},
	})
}
