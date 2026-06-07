package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"git.subcult.tv/subculture-collective/clpr/internal/models"
	"git.subcult.tv/subculture-collective/clpr/internal/services"
)

type CommunityHandler struct {
	communityService *services.CommunityService
	authService      *services.AuthService
}

func NewCommunityHandler(communityService *services.CommunityService, authService *services.AuthService) *CommunityHandler {
	return &CommunityHandler{
		communityService: communityService,
		authService:      authService,
	}
}

// CreateCommunity creates a new community
func (h *CommunityHandler) CreateCommunity(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var req models.CreateCommunityRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	community, err := h.communityService.CreateCommunity(c.Request.Context(), userID.(uuid.UUID), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, community)
}

// GetCommunity retrieves a community by ID or slug
func (h *CommunityHandler) GetCommunity(c *gin.Context) {
	idOrSlug := c.Param("id")

	var requestingUserID *uuid.UUID
	if id, exists := c.Get("user_id"); exists {
		uid := id.(uuid.UUID)
		requestingUserID = &uid
	}

	var community *models.Community
	var err error

	// Try parsing as UUID first
	if communityID, parseErr := uuid.Parse(idOrSlug); parseErr == nil {
		community, err = h.communityService.GetCommunity(c.Request.Context(), communityID, requestingUserID)
	} else {
		// If not a UUID, treat as slug
		community, err = h.communityService.GetCommunityBySlug(c.Request.Context(), idOrSlug, requestingUserID)
	}

	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, community)
}

// ListCommunities lists all communities
func (h *CommunityHandler) ListCommunities(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	sort := c.DefaultQuery("sort", "recent")

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	communities, total, err := h.communityService.ListCommunities(c.Request.Context(), page, limit, sort)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"communities": communities,
		"pagination": gin.H{
			"page":  page,
			"limit": limit,
			"total": total,
		},
	})
}

// SearchCommunities searches communities by name
func (h *CommunityHandler) SearchCommunities(c *gin.Context) {
	query := c.Query("q")
	if query == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "query parameter is required"})
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	communities, total, err := h.communityService.SearchCommunities(c.Request.Context(), query, page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"communities": communities,
		"pagination": gin.H{
			"page":  page,
			"limit": limit,
			"total": total,
		},
	})
}

// UpdateCommunity updates a community
func (h *CommunityHandler) UpdateCommunity(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	communityIDParam := c.Param("id")
	communityID, err := uuid.Parse(communityIDParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid community ID"})
		return
	}

	var req models.UpdateCommunityRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	community, err := h.communityService.UpdateCommunity(c.Request.Context(), communityID, userID.(uuid.UUID), &req)
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, community)
}

// DeleteCommunity deletes a community
func (h *CommunityHandler) DeleteCommunity(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	communityIDParam := c.Param("id")
	communityID, err := uuid.Parse(communityIDParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid community ID"})
		return
	}

	err = h.communityService.DeleteCommunity(c.Request.Context(), communityID, userID.(uuid.UUID))
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusNoContent, nil)
}

// JoinCommunity adds the current user to a community
func (h *CommunityHandler) JoinCommunity(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	communityIDParam := c.Param("id")
	communityID, err := uuid.Parse(communityIDParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid community ID"})
		return
	}

	err = h.communityService.JoinCommunity(c.Request.Context(), communityID, userID.(uuid.UUID))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Successfully joined community"})
}

// LeaveCommunity removes the current user from a community
func (h *CommunityHandler) LeaveCommunity(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	communityIDParam := c.Param("id")
	communityID, err := uuid.Parse(communityIDParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid community ID"})
		return
	}

	err = h.communityService.LeaveCommunity(c.Request.Context(), communityID, userID.(uuid.UUID))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Successfully left community"})
}

// GetMembers retrieves members of a community
func (h *CommunityHandler) GetMembers(c *gin.Context) {
	communityIDParam := c.Param("id")
	communityID, err := uuid.Parse(communityIDParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid community ID"})
		return
	}

	role := c.Query("role")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 50
	}

	members, total, err := h.communityService.GetMembers(c.Request.Context(), communityID, role, page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"members": members,
		"pagination": gin.H{
			"page":  page,
			"limit": limit,
			"total": total,
		},
	})
}

// UpdateMemberRole updates a member's role in a community
func (h *CommunityHandler) UpdateMemberRole(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	communityIDParam := c.Param("id")
	communityID, err := uuid.Parse(communityIDParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid community ID"})
		return
	}

	targetUserIDParam := c.Param("userId")
	targetUserID, err := uuid.Parse(targetUserIDParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	var req models.UpdateMemberRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err = h.communityService.UpdateMemberRole(c.Request.Context(), communityID, userID.(uuid.UUID), targetUserID, req.Role)
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Member role updated successfully"})
}

// BanMember bans a user from a community
func (h *CommunityHandler) BanMember(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	communityIDParam := c.Param("id")
	communityID, err := uuid.Parse(communityIDParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid community ID"})
		return
	}

	var req models.BanMemberRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err = h.communityService.BanMember(c.Request.Context(), communityID, userID.(uuid.UUID), req.UserID, req.Reason)
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Member banned successfully"})
}

// UnbanMember unbans a user from a community
func (h *CommunityHandler) UnbanMember(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	communityIDParam := c.Param("id")
	communityID, err := uuid.Parse(communityIDParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid community ID"})
		return
	}

	targetUserIDParam := c.Param("userId")
	targetUserID, err := uuid.Parse(targetUserIDParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	err = h.communityService.UnbanMember(c.Request.Context(), communityID, userID.(uuid.UUID), targetUserID)
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Member unbanned successfully"})
}

// GetBannedMembers retrieves banned members of a community
func (h *CommunityHandler) GetBannedMembers(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	communityIDParam := c.Param("id")
	communityID, err := uuid.Parse(communityIDParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid community ID"})
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 50
	}

	bans, total, err := h.communityService.GetBannedMembers(c.Request.Context(), communityID, userID.(uuid.UUID), page, limit)
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"bans": bans,
		"pagination": gin.H{
			"page":  page,
			"limit": limit,
			"total": total,
		},
	})
}

// GetCommunityFeed retrieves the community feed (clips)
func (h *CommunityHandler) GetCommunityFeed(c *gin.Context) {
	communityIDParam := c.Param("id")
	communityID, err := uuid.Parse(communityIDParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid community ID"})
		return
	}

	sort := c.DefaultQuery("sort", "recent")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	communityClips, total, err := h.communityService.GetCommunityFeed(c.Request.Context(), communityID, sort, page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Extract just the clips for the response
	clips := make([]*models.Clip, len(communityClips))
	for i, cc := range communityClips {
		clips[i] = cc.Clip
	}

	c.JSON(http.StatusOK, gin.H{
		"clips": clips,
		"pagination": gin.H{
			"page":  page,
			"limit": limit,
			"total": total,
		},
	})
}

// AddClipToCommunity adds a clip to the community feed
func (h *CommunityHandler) AddClipToCommunity(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	communityIDParam := c.Param("id")
	communityID, err := uuid.Parse(communityIDParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid community ID"})
		return
	}

	var req models.AddClipToCommunityRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err = h.communityService.AddClipToCommunity(c.Request.Context(), communityID, userID.(uuid.UUID), req.ClipID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Clip added to community successfully"})
}

// RemoveClipFromCommunity removes a clip from the community feed
func (h *CommunityHandler) RemoveClipFromCommunity(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	communityIDParam := c.Param("id")
	communityID, err := uuid.Parse(communityIDParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid community ID"})
		return
	}

	clipIDParam := c.Param("clipId")
	clipID, err := uuid.Parse(clipIDParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid clip ID"})
		return
	}

	err = h.communityService.RemoveClipFromCommunity(c.Request.Context(), communityID, userID.(uuid.UUID), clipID)
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Clip removed from community successfully"})
}

// CreateDiscussion creates a new discussion thread
func (h *CommunityHandler) CreateDiscussion(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	communityIDParam := c.Param("id")
	communityID, err := uuid.Parse(communityIDParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid community ID"})
		return
	}

	var req models.CreateDiscussionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	discussion, err := h.communityService.CreateDiscussion(c.Request.Context(), communityID, userID.(uuid.UUID), &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, discussion)
}

// GetDiscussion retrieves a discussion thread
func (h *CommunityHandler) GetDiscussion(c *gin.Context) {
	discussionIDParam := c.Param("discussionId")
	discussionID, err := uuid.Parse(discussionIDParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid discussion ID"})
		return
	}

	discussion, err := h.communityService.GetDiscussion(c.Request.Context(), discussionID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, discussion)
}

// ListDiscussions lists discussions for a community
func (h *CommunityHandler) ListDiscussions(c *gin.Context) {
	communityIDParam := c.Param("id")
	communityID, err := uuid.Parse(communityIDParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid community ID"})
		return
	}

	sort := c.DefaultQuery("sort", "recent")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	discussions, total, err := h.communityService.ListDiscussions(c.Request.Context(), communityID, sort, page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"discussions": discussions,
		"pagination": gin.H{
			"page":  page,
			"limit": limit,
			"total": total,
		},
	})
}

// UpdateDiscussion updates a discussion thread
func (h *CommunityHandler) UpdateDiscussion(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	discussionIDParam := c.Param("discussionId")
	discussionID, err := uuid.Parse(discussionIDParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid discussion ID"})
		return
	}

	var req models.UpdateDiscussionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	discussion, err := h.communityService.UpdateDiscussion(c.Request.Context(), discussionID, userID.(uuid.UUID), &req)
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, discussion)
}

// DeleteDiscussion deletes a discussion thread
func (h *CommunityHandler) DeleteDiscussion(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	discussionIDParam := c.Param("discussionId")
	discussionID, err := uuid.Parse(discussionIDParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid discussion ID"})
		return
	}

	err = h.communityService.DeleteDiscussion(c.Request.Context(), discussionID, userID.(uuid.UUID))
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusNoContent, nil)
}
