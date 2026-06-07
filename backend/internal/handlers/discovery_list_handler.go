package handlers

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"git.subcult.tv/subculture-collective/clpr/internal/models"
	"git.subcult.tv/subculture-collective/clpr/internal/repository"
	pkgutils "git.subcult.tv/subculture-collective/clpr/pkg/utils"
)

// DiscoveryListHandler handles discovery list-related requests
type DiscoveryListHandler struct {
	repo          repository.DiscoveryListRepositoryInterface
	analyticsRepo *repository.AnalyticsRepository
}

// NewDiscoveryListHandler creates a new handler instance
func NewDiscoveryListHandler(repo repository.DiscoveryListRepositoryInterface, analyticsRepo *repository.AnalyticsRepository) *DiscoveryListHandler {
	return &DiscoveryListHandler{
		repo:          repo,
		analyticsRepo: analyticsRepo,
	}
}

// ListDiscoveryLists godoc
// @Summary List discovery lists
// @Description Get a list of discovery lists, optionally filtering for featured only
// @Tags discovery-lists
// @Accept json
// @Produce json
// @Param featured query bool false "Filter for featured lists only"
// @Param limit query int false "Number of lists to return" default(20)
// @Param offset query int false "Offset for pagination" default(0)
// @Success 200 {array} models.DiscoveryListWithStats
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/discovery-lists [get]
func (h *DiscoveryListHandler) ListDiscoveryLists(c *gin.Context) {
	logger := pkgutils.GetLogger()
	ctx := c.Request.Context()

	// Parse query parameters
	featuredOnly := c.Query("featured") == "true"
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	// Validate limits
	if limit < 1 || limit > 100 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}

	// Get user ID if authenticated
	var userID *uuid.UUID
	if userIDVal, exists := c.Get("user_id"); exists {
		if id, ok := userIDVal.(uuid.UUID); ok {
			userID = &id
		}
	}

	// Get lists from repository
	lists, err := h.repo.ListDiscoveryLists(ctx, featuredOnly, userID, limit, offset)
	if err != nil {
		logger.Error("Failed to list discovery lists", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve discovery lists"})
		return
	}

	c.JSON(http.StatusOK, lists)
}

// GetDiscoveryList godoc
// @Summary Get a single discovery list
// @Description Get details of a discovery list by ID or slug
// @Tags discovery-lists
// @Accept json
// @Produce json
// @Param id path string true "List ID or slug"
// @Success 200 {object} models.DiscoveryListWithStats
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/discovery-lists/{id} [get]
func (h *DiscoveryListHandler) GetDiscoveryList(c *gin.Context) {
	logger := pkgutils.GetLogger()
	ctx := c.Request.Context()
	idOrSlug := c.Param("id")

	// Get user ID if authenticated
	var userID *uuid.UUID
	if userIDVal, exists := c.Get("user_id"); exists {
		if id, ok := userIDVal.(uuid.UUID); ok {
			userID = &id
		}
	}

	// Get list from repository
	list, err := h.repo.GetDiscoveryList(ctx, idOrSlug, userID)
	if err != nil {
		if err.Error() == "discovery list not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "Discovery list not found"})
			return
		}
		logger.Error("Failed to get discovery list", err, map[string]interface{}{"id": idOrSlug})
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve discovery list"})
		return
	}

	// Track list view for analytics
	if h.analyticsRepo != nil {
		go func() {
			event := &models.AnalyticsEvent{
				EventType: "discovery_list_view",
				UserID:    userID,
			}
			// Note: We can't easily track the list ID as clip_id here since the schema expects a clip
			_ = h.analyticsRepo.TrackEvent(ctx, event)
		}()
	}

	c.JSON(http.StatusOK, list)
}

// GetDiscoveryListClips godoc
// @Summary Get clips in a discovery list
// @Description Get all clips in a discovery list with pagination
// @Tags discovery-lists
// @Accept json
// @Produce json
// @Param id path string true "List ID"
// @Param limit query int false "Number of clips to return" default(20)
// @Param offset query int false "Offset for pagination" default(0)
// @Success 200 {array} models.ClipWithSubmitter
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/discovery-lists/{id}/clips [get]
func (h *DiscoveryListHandler) GetDiscoveryListClips(c *gin.Context) {
	logger := pkgutils.GetLogger()
	ctx := c.Request.Context()
	listIDStr := c.Param("id")

	// Get user ID if authenticated
	var userID *uuid.UUID
	if userIDVal, exists := c.Get("user_id"); exists {
		if id, ok := userIDVal.(uuid.UUID); ok {
			userID = &id
		}
	}

	// Resolve listIDStr to UUID (accept both UUID and slug)
	var listID uuid.UUID
	listID, err := uuid.Parse(listIDStr)
	if err != nil {
		// Not a UUID, try to resolve as slug
		list, err2 := h.repo.GetDiscoveryList(ctx, listIDStr, userID)
		if err2 != nil {
			if err2.Error() == "discovery list not found" {
				c.JSON(http.StatusNotFound, gin.H{"error": "Discovery list not found"})
				return
			}
			logger.Error("Failed to resolve discovery list slug", err2, map[string]interface{}{"slug": listIDStr})
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve discovery list"})
			return
		}
		listID = list.ID
	}

	// Parse query parameters
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	// Validate limits
	if limit < 1 || limit > 100 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}

	// Get clips from repository
	clips, total, err := h.repo.GetListClips(ctx, listID, userID, limit, offset)
	if err != nil {
		logger.Error("Failed to get discovery list clips", err, map[string]interface{}{"list_id": listID})
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve clips"})
		return
	}

	// Calculate pagination metadata
	page := offset / limit
	hasMore := len(clips) == limit && (offset+limit) < total

	c.JSON(http.StatusOK, gin.H{
		"clips":    clips,
		"total":    total,
		"page":     page,
		"limit":    limit,
		"has_more": hasMore,
	})
}

// FollowDiscoveryList godoc
// @Summary Follow a discovery list
// @Description Follow a discovery list to get notifications when new clips are added
// @Tags discovery-lists
// @Accept json
// @Produce json
// @Param id path string true "List ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/discovery-lists/{id}/follow [post]
func (h *DiscoveryListHandler) FollowDiscoveryList(c *gin.Context) {
	logger := pkgutils.GetLogger()
	ctx := c.Request.Context()
	listIDStr := c.Param("id")

	// Get user ID from context
	userIDVal, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	userID := userIDVal.(uuid.UUID)

	// Parse list ID
	listID, err := uuid.Parse(listIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid list ID"})
		return
	}

	// Verify list exists
	_, err = h.repo.GetDiscoveryList(ctx, listIDStr, &userID)
	if err != nil {
		if err.Error() == "discovery list not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "Discovery list not found"})
			return
		}
		logger.Error("Failed to verify discovery list", err, map[string]interface{}{"id": listIDStr})
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to follow list"})
		return
	}

	// Follow list
	err = h.repo.FollowList(ctx, userID, listID)
	if err != nil {
		logger.Error("Failed to follow discovery list", err, map[string]interface{}{"user_id": userID, "list_id": listID})
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to follow list"})
		return
	}

	// Track follow event for analytics
	if h.analyticsRepo != nil {
		go func() {
			event := &models.AnalyticsEvent{
				EventType: "discovery_list_follow",
				UserID:    &userID,
			}
			_ = h.analyticsRepo.TrackEvent(ctx, event)
		}()
	}

	c.JSON(http.StatusOK, gin.H{"message": "Successfully followed list"})
}

// UnfollowDiscoveryList godoc
// @Summary Unfollow a discovery list
// @Description Stop following a discovery list
// @Tags discovery-lists
// @Accept json
// @Produce json
// @Param id path string true "List ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/discovery-lists/{id}/follow [delete]
func (h *DiscoveryListHandler) UnfollowDiscoveryList(c *gin.Context) {
	logger := pkgutils.GetLogger()
	ctx := c.Request.Context()
	listIDStr := c.Param("id")

	// Get user ID from context
	userIDVal, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	userID := userIDVal.(uuid.UUID)

	// Parse list ID
	listID, err := uuid.Parse(listIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid list ID"})
		return
	}

	// Unfollow list
	err = h.repo.UnfollowList(ctx, userID, listID)
	if err != nil {
		if err.Error() == "not following this list" {
			c.JSON(http.StatusNotFound, gin.H{"error": "Not following this list"})
			return
		}
		logger.Error("Failed to unfollow discovery list", err, map[string]interface{}{"user_id": userID, "list_id": listID})
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to unfollow list"})
		return
	}

	// Track unfollow event for analytics
	if h.analyticsRepo != nil {
		go func() {
			event := &models.AnalyticsEvent{
				EventType: "discovery_list_unfollow",
				UserID:    &userID,
			}
			_ = h.analyticsRepo.TrackEvent(ctx, event)
		}()
	}

	c.JSON(http.StatusOK, gin.H{"message": "Successfully unfollowed list"})
}

// BookmarkDiscoveryList godoc
// @Summary Bookmark a discovery list
// @Description Bookmark a discovery list for quick access later
// @Tags discovery-lists
// @Accept json
// @Produce json
// @Param id path string true "List ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/discovery-lists/{id}/bookmark [post]
func (h *DiscoveryListHandler) BookmarkDiscoveryList(c *gin.Context) {
	logger := pkgutils.GetLogger()
	ctx := c.Request.Context()
	listIDStr := c.Param("id")

	// Get user ID from context
	userIDVal, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	userID := userIDVal.(uuid.UUID)

	// Parse list ID
	listID, err := uuid.Parse(listIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid list ID"})
		return
	}

	// Verify list exists
	_, err = h.repo.GetDiscoveryList(ctx, listIDStr, &userID)
	if err != nil {
		if err.Error() == "discovery list not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "Discovery list not found"})
			return
		}
		logger.Error("Failed to verify discovery list", err, map[string]interface{}{"id": listIDStr})
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to bookmark list"})
		return
	}

	// Bookmark list
	err = h.repo.BookmarkList(ctx, userID, listID)
	if err != nil {
		logger.Error("Failed to bookmark discovery list", err, map[string]interface{}{"user_id": userID, "list_id": listID})
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to bookmark list"})
		return
	}

	// Track bookmark event for analytics
	if h.analyticsRepo != nil {
		go func() {
			event := &models.AnalyticsEvent{
				EventType: "discovery_list_bookmark",
				UserID:    &userID,
			}
			_ = h.analyticsRepo.TrackEvent(ctx, event)
		}()
	}

	c.JSON(http.StatusOK, gin.H{"message": "Successfully bookmarked list"})
}

// UnbookmarkDiscoveryList godoc
// @Summary Remove bookmark from a discovery list
// @Description Remove a bookmark from a discovery list
// @Tags discovery-lists
// @Accept json
// @Produce json
// @Param id path string true "List ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/discovery-lists/{id}/bookmark [delete]
func (h *DiscoveryListHandler) UnbookmarkDiscoveryList(c *gin.Context) {
	logger := pkgutils.GetLogger()
	ctx := c.Request.Context()
	listIDStr := c.Param("id")

	// Get user ID from context
	userIDVal, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	userID := userIDVal.(uuid.UUID)

	// Parse list ID
	listID, err := uuid.Parse(listIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid list ID"})
		return
	}

	// Unbookmark list
	err = h.repo.UnbookmarkList(ctx, userID, listID)
	if err != nil {
		if err.Error() == "list not bookmarked" {
			c.JSON(http.StatusNotFound, gin.H{"error": "List not bookmarked"})
			return
		}
		logger.Error("Failed to unbookmark discovery list", err, map[string]interface{}{"user_id": userID, "list_id": listID})
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to unbookmark list"})
		return
	}

	// Track unbookmark event for analytics
	if h.analyticsRepo != nil {
		go func() {
			event := &models.AnalyticsEvent{
				EventType: "discovery_list_unbookmark",
				UserID:    &userID,
			}
			_ = h.analyticsRepo.TrackEvent(ctx, event)
		}()
	}

	c.JSON(http.StatusOK, gin.H{"message": "Successfully removed bookmark"})
}

// GetUserFollowedLists godoc
// @Summary Get user's followed discovery lists
// @Description Get all discovery lists that the current user follows
// @Tags discovery-lists
// @Accept json
// @Produce json
// @Param limit query int false "Number of lists to return" default(20)
// @Param offset query int false "Offset for pagination" default(0)
// @Success 200 {array} models.DiscoveryListWithStats
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/users/me/discovery-list-follows [get]
func (h *DiscoveryListHandler) GetUserFollowedLists(c *gin.Context) {
	logger := pkgutils.GetLogger()
	ctx := c.Request.Context()

	// Get user ID from context
	userIDVal, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	userID := userIDVal.(uuid.UUID)

	// Parse query parameters
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	// Validate limits
	if limit < 1 || limit > 100 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}

	// Get followed lists from repository
	lists, err := h.repo.GetUserFollowedLists(ctx, userID, limit, offset)
	if err != nil {
		logger.Error("Failed to get user followed lists", err, map[string]interface{}{"user_id": userID})
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve followed lists"})
		return
	}

	c.JSON(http.StatusOK, lists)
}

// Admin-only handlers

// AdminListDiscoveryLists godoc
// @Summary List all discovery lists (admin)
// @Description Get all discovery lists including inactive ones
// @Tags admin, discovery-lists
// @Accept json
// @Produce json
// @Param limit query int false "Number of lists to return" default(50)
// @Param offset query int false "Offset for pagination" default(0)
// @Success 200 {array} models.DiscoveryListWithStats
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/admin/discovery-lists [get]
func (h *DiscoveryListHandler) AdminListDiscoveryLists(c *gin.Context) {
	logger := pkgutils.GetLogger()
	ctx := c.Request.Context()

	// Parse query parameters
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	// Validate limits
	if limit < 1 || limit > 100 {
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}

	// Get all lists from repository
	lists, err := h.repo.ListAllDiscoveryLists(ctx, limit, offset)
	if err != nil {
		logger.Error("Failed to list all discovery lists", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve discovery lists"})
		return
	}

	c.JSON(http.StatusOK, lists)
}

// AdminCreateDiscoveryList godoc
// @Summary Create a discovery list (admin)
// @Description Create a new discovery list
// @Tags admin, discovery-lists
// @Accept json
// @Produce json
// @Param list body models.CreateDiscoveryListRequest true "List details"
// @Success 201 {object} models.DiscoveryList
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/admin/discovery-lists [post]
func (h *DiscoveryListHandler) AdminCreateDiscoveryList(c *gin.Context) {
	logger := pkgutils.GetLogger()
	ctx := c.Request.Context()

	// Get user ID from context
	userIDVal, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	userID := userIDVal.(uuid.UUID)

	// Parse request body
	var req models.CreateDiscoveryListRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body", "details": err.Error()})
		return
	}

	// Generate slug from name
	slug := pkgutils.Slugify(req.Name)

	// Set default for IsFeatured
	isFeatured := false
	if req.IsFeatured != nil {
		isFeatured = *req.IsFeatured
	}

	// Set default description
	description := ""
	if req.Description != nil {
		description = *req.Description
	}

	// Create list
	list, err := h.repo.CreateDiscoveryList(ctx, req.Name, slug, description, isFeatured, userID)
	if err != nil {
		logger.Error("Failed to create discovery list", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create discovery list"})
		return
	}

	c.JSON(http.StatusCreated, list)
}

// AdminUpdateDiscoveryList godoc
// @Summary Update a discovery list (admin)
// @Description Update an existing discovery list
// @Tags admin, discovery-lists
// @Accept json
// @Produce json
// @Param id path string true "List ID"
// @Param list body models.UpdateDiscoveryListRequest true "Updated list details"
// @Success 200 {object} models.DiscoveryList
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/admin/discovery-lists/{id} [put]
func (h *DiscoveryListHandler) AdminUpdateDiscoveryList(c *gin.Context) {
	logger := pkgutils.GetLogger()
	ctx := c.Request.Context()
	listIDStr := c.Param("id")

	// Parse list ID
	listID, err := uuid.Parse(listIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid list ID"})
		return
	}

	// Parse request body
	var req models.UpdateDiscoveryListRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body", "details": err.Error()})
		return
	}

	// Update list
	list, err := h.repo.UpdateDiscoveryList(ctx, listID, req.Name, req.Description, req.IsFeatured)
	if err != nil {
		if errors.Is(err, repository.ErrDiscoveryListNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Discovery list not found"})
			return
		}
		logger.Error("Failed to update discovery list", err, map[string]interface{}{"list_id": listID})
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update discovery list"})
		return
	}

	c.JSON(http.StatusOK, list)
}

// AdminDeleteDiscoveryList godoc
// @Summary Delete a discovery list (admin)
// @Description Delete a discovery list permanently
// @Tags admin, discovery-lists
// @Accept json
// @Produce json
// @Param id path string true "List ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/admin/discovery-lists/{id} [delete]
func (h *DiscoveryListHandler) AdminDeleteDiscoveryList(c *gin.Context) {
	logger := pkgutils.GetLogger()
	ctx := c.Request.Context()
	listIDStr := c.Param("id")

	// Parse list ID
	listID, err := uuid.Parse(listIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid list ID"})
		return
	}

	// Delete list
	err = h.repo.DeleteDiscoveryList(ctx, listID)
	if err != nil {
		if errors.Is(err, repository.ErrDiscoveryListNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Discovery list not found"})
			return
		}
		logger.Error("Failed to delete discovery list", err, map[string]interface{}{"list_id": listID})
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete discovery list"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Discovery list deleted successfully"})
}

// AdminAddClipToList godoc
// @Summary Add a clip to a discovery list (admin)
// @Description Add a clip to a discovery list
// @Tags admin, discovery-lists
// @Accept json
// @Produce json
// @Param id path string true "List ID"
// @Param clip body models.AddClipToListRequest true "Clip to add"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/admin/discovery-lists/{id}/clips [post]
func (h *DiscoveryListHandler) AdminAddClipToList(c *gin.Context) {
	logger := pkgutils.GetLogger()
	ctx := c.Request.Context()
	listIDStr := c.Param("id")

	// Parse list ID
	listID, err := uuid.Parse(listIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid list ID"})
		return
	}

	// Parse request body
	var req models.AddClipToListRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body", "details": err.Error()})
		return
	}

	// Add clip to list
	err = h.repo.AddClipToList(ctx, listID, req.ClipID)
	if err != nil {
		logger.Error("Failed to add clip to list", err, map[string]interface{}{"list_id": listID, "clip_id": req.ClipID})
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add clip to list"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Clip added to list successfully"})
}

// AdminRemoveClipFromList godoc
// @Summary Remove a clip from a discovery list (admin)
// @Description Remove a clip from a discovery list
// @Tags admin, discovery-lists
// @Accept json
// @Produce json
// @Param id path string true "List ID"
// @Param clipId path string true "Clip ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/admin/discovery-lists/{id}/clips/{clipId} [delete]
func (h *DiscoveryListHandler) AdminRemoveClipFromList(c *gin.Context) {
	logger := pkgutils.GetLogger()
	ctx := c.Request.Context()
	listIDStr := c.Param("id")
	clipIDStr := c.Param("clipId")

	// Parse IDs
	listID, err := uuid.Parse(listIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid list ID"})
		return
	}

	clipID, err := uuid.Parse(clipIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid clip ID"})
		return
	}

	// Remove clip from list
	err = h.repo.RemoveClipFromList(ctx, listID, clipID)
	if err != nil {
		if errors.Is(err, repository.ErrClipNotFoundInList) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Clip not found in list"})
			return
		}
		logger.Error("Failed to remove clip from list", err, map[string]interface{}{"list_id": listID, "clip_id": clipID})
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to remove clip from list"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Clip removed from list successfully"})
}

const (
	// MaxReorderClipsLimit defines the maximum number of clips that can be reordered at once
	MaxReorderClipsLimit = 200
)

// AdminReorderListClips godoc
// @Summary Reorder clips in a discovery list (admin)
// @Description Reorder clips in a discovery list by providing ordered clip IDs
// @Tags admin, discovery-lists
// @Accept json
// @Produce json
// @Param id path string true "List ID"
// @Param clips body models.ReorderListClipsRequest true "Ordered clip IDs"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/admin/discovery-lists/{id}/clips/reorder [put]
func (h *DiscoveryListHandler) AdminReorderListClips(c *gin.Context) {
	logger := pkgutils.GetLogger()
	ctx := c.Request.Context()
	listIDStr := c.Param("id")

	// Parse list ID
	listID, err := uuid.Parse(listIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid list ID"})
		return
	}

	// Parse request body
	var req models.ReorderListClipsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body", "details": err.Error()})
		return
	}

	// Validate clip count
	if len(req.ClipIDs) > MaxReorderClipsLimit {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Cannot reorder more than %d clips at once", MaxReorderClipsLimit)})
		return
	}

	// Reorder clips
	err = h.repo.ReorderListClips(ctx, listID, req.ClipIDs)
	if err != nil {
		logger.Error("Failed to reorder clips", err, map[string]interface{}{"list_id": listID})
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to reorder clips"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Clips reordered successfully"})
}
