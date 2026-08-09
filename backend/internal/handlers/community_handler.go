package handlers

import (
	"net/http"
	"strconv"
	"strings"

	"git.subcult.tv/subculture-collective/clpr/internal/models"
	"git.subcult.tv/subculture-collective/clpr/internal/services"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func optionalCommunityUserID(c *gin.Context) (*uuid.UUID, bool) {
	value, exists := c.Get("user_id")
	if !exists {
		return nil, true
	}
	userID, ok := value.(uuid.UUID)
	if !ok || userID == uuid.Nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return nil, false
	}
	return &userID, true
}

func parseCommunityPagination(c *gin.Context, defaultLimit int) (int, int, bool) {
	page, err := strconv.Atoi(c.DefaultQuery("page", "1"))
	if err != nil || page < 1 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "page must be a positive integer"})
		return 0, 0, false
	}
	limit, err := strconv.Atoi(c.DefaultQuery("limit", strconv.Itoa(defaultLimit)))
	if err != nil || limit < 1 || limit > 100 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "limit must be between 1 and 100"})
		return 0, 0, false
	}
	return page, limit, true
}

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
	userID, ok := authenticatedUserID(c)
	if !ok {
		return
	}

	var req models.CreateCommunityRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid community"})
		return
	}
	if req.IsPublic != nil && !*req.IsPublic {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Private communities are not available"})
		return
	}

	community, err := h.communityService.CreateCommunity(c.Request.Context(), userID, &req)
	if err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Community could not be created"})
		return
	}

	c.JSON(http.StatusCreated, community)
}

// GetCommunity retrieves a community by ID or slug
func (h *CommunityHandler) GetCommunity(c *gin.Context) {
	idOrSlug := c.Param("id")

	requestingUserID, ok := optionalCommunityUserID(c)
	if !ok {
		return
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
		c.JSON(http.StatusNotFound, gin.H{"error": "Community not found"})
		return
	}

	c.JSON(http.StatusOK, community)
}

// ListCommunities lists all communities
func (h *CommunityHandler) ListCommunities(c *gin.Context) {
	page, limit, ok := parseCommunityPagination(c, 20)
	if !ok {
		return
	}
	sort := c.DefaultQuery("sort", "recent")
	if sort != "recent" && sort != "members" && sort != "name" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "sort must be recent, members, or name"})
		return
	}

	communities, total, err := h.communityService.ListCommunities(c.Request.Context(), page, limit, sort)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list communities"})
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
	query := strings.TrimSpace(c.Query("q"))
	if query == "" || len(query) > 100 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "query parameter is required"})
		return
	}

	page, limit, ok := parseCommunityPagination(c, 20)
	if !ok {
		return
	}

	communities, total, err := h.communityService.SearchCommunities(c.Request.Context(), query, page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to search communities"})
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
	userID, ok := authenticatedUserID(c)
	if !ok {
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
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid community update"})
		return
	}
	if req.Name == nil && req.Description == nil && req.Icon == nil && req.IsPublic == nil && req.Rules == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "At least one field is required"})
		return
	}
	if req.IsPublic != nil && !*req.IsPublic {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Private communities are not available"})
		return
	}

	community, err := h.communityService.UpdateCommunity(c.Request.Context(), communityID, userID, &req)
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "Community update not permitted"})
		return
	}

	c.JSON(http.StatusOK, community)
}

// DeleteCommunity deletes a community
func (h *CommunityHandler) DeleteCommunity(c *gin.Context) {
	userID, ok := authenticatedUserID(c)
	if !ok {
		return
	}

	communityIDParam := c.Param("id")
	communityID, err := uuid.Parse(communityIDParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid community ID"})
		return
	}

	err = h.communityService.DeleteCommunity(c.Request.Context(), communityID, userID)
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "Community deletion not permitted"})
		return
	}

	c.JSON(http.StatusNoContent, nil)
}

// JoinCommunity adds the current user to a community
func (h *CommunityHandler) JoinCommunity(c *gin.Context) {
	userID, ok := authenticatedUserID(c)
	if !ok {
		return
	}

	communityIDParam := c.Param("id")
	communityID, err := uuid.Parse(communityIDParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid community ID"})
		return
	}

	err = h.communityService.JoinCommunity(c.Request.Context(), communityID, userID)
	if err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Community cannot be joined"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Successfully joined community"})
}

// LeaveCommunity removes the current user from a community
func (h *CommunityHandler) LeaveCommunity(c *gin.Context) {
	userID, ok := authenticatedUserID(c)
	if !ok {
		return
	}

	communityIDParam := c.Param("id")
	communityID, err := uuid.Parse(communityIDParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid community ID"})
		return
	}

	err = h.communityService.LeaveCommunity(c.Request.Context(), communityID, userID)
	if err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Community cannot be left"})
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
	if role != "" && role != models.CommunityRoleAdmin && role != models.CommunityRoleMod && role != models.CommunityRoleMember {
		c.JSON(http.StatusBadRequest, gin.H{"error": "role must be admin, mod, or member"})
		return
	}
	page, limit, ok := parseCommunityPagination(c, 50)
	if !ok {
		return
	}
	requestingUserID, ok := optionalCommunityUserID(c)
	if !ok {
		return
	}
	members, total, err := h.communityService.GetVisibleMembers(c.Request.Context(), communityID, requestingUserID, role, page, limit)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Community not found"})
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
	userID, ok := authenticatedUserID(c)
	if !ok {
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
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid member role"})
		return
	}

	err = h.communityService.UpdateMemberRole(c.Request.Context(), communityID, userID, targetUserID, req.Role)
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "Member role update not permitted"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Member role updated successfully"})
}

// BanMember bans a user from a community
func (h *CommunityHandler) BanMember(c *gin.Context) {
	userID, ok := authenticatedUserID(c)
	if !ok {
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
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid member ban"})
		return
	}

	err = h.communityService.BanMember(c.Request.Context(), communityID, userID, req.UserID, req.Reason)
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "Member ban not permitted"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Member banned successfully"})
}

// UnbanMember unbans a user from a community
func (h *CommunityHandler) UnbanMember(c *gin.Context) {
	userID, ok := authenticatedUserID(c)
	if !ok {
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

	err = h.communityService.UnbanMember(c.Request.Context(), communityID, userID, targetUserID)
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "Member unban not permitted"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Member unbanned successfully"})
}

// GetBannedMembers retrieves banned members of a community
func (h *CommunityHandler) GetBannedMembers(c *gin.Context) {
	userID, ok := authenticatedUserID(c)
	if !ok {
		return
	}

	communityIDParam := c.Param("id")
	communityID, err := uuid.Parse(communityIDParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid community ID"})
		return
	}

	page, limit, ok := parseCommunityPagination(c, 50)
	if !ok {
		return
	}

	bans, total, err := h.communityService.GetBannedMembers(c.Request.Context(), communityID, userID, page, limit)
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "Banned members are restricted to community staff"})
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
	if sort != "recent" && sort != "trending" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "sort must be recent or trending"})
		return
	}
	page, limit, ok := parseCommunityPagination(c, 20)
	if !ok {
		return
	}
	requestingUserID, ok := optionalCommunityUserID(c)
	if !ok {
		return
	}
	communityClips, total, err := h.communityService.GetVisibleCommunityFeed(c.Request.Context(), communityID, requestingUserID, sort, page, limit)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Community not found"})
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
	userID, ok := authenticatedUserID(c)
	if !ok {
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
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid clip"})
		return
	}

	err = h.communityService.AddClipToCommunity(c.Request.Context(), communityID, userID, req.ClipID)
	if err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Clip could not be added to the community"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Clip added to community successfully"})
}

// RemoveClipFromCommunity removes a clip from the community feed
func (h *CommunityHandler) RemoveClipFromCommunity(c *gin.Context) {
	userID, ok := authenticatedUserID(c)
	if !ok {
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

	err = h.communityService.RemoveClipFromCommunity(c.Request.Context(), communityID, userID, clipID)
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "Clip removal not permitted"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Clip removed from community successfully"})
}

// CreateDiscussion creates a new discussion thread
func (h *CommunityHandler) CreateDiscussion(c *gin.Context) {
	userID, ok := authenticatedUserID(c)
	if !ok {
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
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid discussion"})
		return
	}

	discussion, err := h.communityService.CreateDiscussion(c.Request.Context(), communityID, userID, &req)
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "Discussion creation not permitted"})
		return
	}

	c.JSON(http.StatusCreated, discussion)
}

// GetDiscussion retrieves a discussion thread
func (h *CommunityHandler) GetDiscussion(c *gin.Context) {
	communityID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid community ID"})
		return
	}
	discussionIDParam := c.Param("discussionId")
	discussionID, err := uuid.Parse(discussionIDParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid discussion ID"})
		return
	}

	requestingUserID, ok := optionalCommunityUserID(c)
	if !ok {
		return
	}
	discussion, err := h.communityService.GetVisibleDiscussion(c.Request.Context(), communityID, discussionID, requestingUserID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Discussion not found"})
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
	if sort != "recent" && sort != "trending" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "sort must be recent or trending"})
		return
	}
	page, limit, ok := parseCommunityPagination(c, 20)
	if !ok {
		return
	}
	requestingUserID, ok := optionalCommunityUserID(c)
	if !ok {
		return
	}
	discussions, total, err := h.communityService.ListVisibleDiscussions(c.Request.Context(), communityID, requestingUserID, sort, page, limit)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Community not found"})
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
	userID, ok := authenticatedUserID(c)
	if !ok {
		return
	}
	communityID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid community ID"})
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
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid discussion update"})
		return
	}
	if req.Title == nil && req.Content == nil && req.IsPinned == nil && req.IsResolved == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "At least one field is required"})
		return
	}

	discussion, err := h.communityService.UpdateDiscussion(c.Request.Context(), communityID, discussionID, userID, &req)
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "Discussion update not permitted"})
		return
	}

	c.JSON(http.StatusOK, discussion)
}

// DeleteDiscussion deletes a discussion thread
func (h *CommunityHandler) DeleteDiscussion(c *gin.Context) {
	userID, ok := authenticatedUserID(c)
	if !ok {
		return
	}
	communityID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid community ID"})
		return
	}

	discussionIDParam := c.Param("discussionId")
	discussionID, err := uuid.Parse(discussionIDParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid discussion ID"})
		return
	}

	err = h.communityService.DeleteDiscussion(c.Request.Context(), communityID, discussionID, userID)
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "Discussion deletion not permitted"})
		return
	}

	c.JSON(http.StatusNoContent, nil)
}
