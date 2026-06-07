package services

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
	"git.subcult.tv/subculture-collective/clpr/internal/models"
	"git.subcult.tv/subculture-collective/clpr/internal/repository"
)

type CommunityService struct {
	communityRepo *repository.CommunityRepository
	clipRepo      *repository.ClipRepository
	userRepo      *repository.UserRepository
	notifService  *NotificationService
}

func NewCommunityService(
	communityRepo *repository.CommunityRepository,
	clipRepo *repository.ClipRepository,
	userRepo *repository.UserRepository,
	notifService *NotificationService,
) *CommunityService {
	return &CommunityService{
		communityRepo: communityRepo,
		clipRepo:      clipRepo,
		userRepo:      userRepo,
		notifService:  notifService,
	}
}

// Role hierarchy levels (admin > mod > member)
var roleHierarchy = map[string]int{
	models.CommunityRoleAdmin:  3,
	models.CommunityRoleMod:    2,
	models.CommunityRoleMember: 1,
}

// generateSlug generates a URL-friendly slug from a name
func generateSlug(name string) string {
	// Convert to lowercase
	slug := strings.ToLower(name)
	// Replace spaces and special characters with hyphens
	reg := regexp.MustCompile("[^a-z0-9]+")
	slug = reg.ReplaceAllString(slug, "-")
	// Remove leading/trailing hyphens
	slug = strings.Trim(slug, "-")
	// Return empty string if no valid characters remain
	return slug
}

// CreateCommunity creates a new community
func (s *CommunityService) CreateCommunity(ctx context.Context, ownerID uuid.UUID, req *models.CreateCommunityRequest) (*models.Community, error) {
	// Generate slug from name
	slug := generateSlug(req.Name)

	// Validate slug is not empty
	if slug == "" {
		return nil, fmt.Errorf("community name must contain at least one alphanumeric character")
	}

	// Check if slug already exists
	existing, _ := s.communityRepo.GetCommunityBySlug(ctx, slug)
	if existing != nil {
		return nil, fmt.Errorf("community with this name already exists")
	}

	community := &models.Community{
		ID:          uuid.New(),
		Name:        req.Name,
		Slug:        slug,
		Description: req.Description,
		Icon:        req.Icon,
		OwnerID:     ownerID,
		IsPublic:    true,
		MemberCount: 0, // Let DB trigger increment when owner is added
		Rules:       req.Rules,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if req.IsPublic != nil {
		community.IsPublic = *req.IsPublic
	}

	// Create the community
	err := s.communityRepo.CreateCommunity(ctx, community)
	if err != nil {
		return nil, fmt.Errorf("failed to create community: %w", err)
	}

	// Automatically add owner as admin
	member := &models.CommunityMember{
		ID:          uuid.New(),
		CommunityID: community.ID,
		UserID:      ownerID,
		Role:        models.CommunityRoleAdmin,
		JoinedAt:    time.Now(),
	}
	err = s.communityRepo.AddMember(ctx, member)
	if err != nil {
		// If adding member fails, try to clean up the community
		_ = s.communityRepo.DeleteCommunity(ctx, community.ID)
		return nil, fmt.Errorf("failed to add owner as member: %w", err)
	}

	return community, nil
}

// GetCommunity retrieves a community by ID
func (s *CommunityService) GetCommunity(ctx context.Context, communityID uuid.UUID, requestingUserID *uuid.UUID) (*models.Community, error) {
	community, err := s.communityRepo.GetCommunityByID(ctx, communityID)
	if err != nil {
		return nil, err
	}

	// Check if the requesting user has access to this community
	if !community.IsPublic {
		if requestingUserID == nil {
			return nil, fmt.Errorf("unauthorized access to private community")
		}
		isMember, err := s.communityRepo.IsMember(ctx, communityID, *requestingUserID)
		if err != nil {
			return nil, err
		}
		if !isMember {
			return nil, fmt.Errorf("unauthorized access to private community")
		}
	}

	return community, nil
}

// GetCommunityBySlug retrieves a community by slug
func (s *CommunityService) GetCommunityBySlug(ctx context.Context, slug string, requestingUserID *uuid.UUID) (*models.Community, error) {
	community, err := s.communityRepo.GetCommunityBySlug(ctx, slug)
	if err != nil {
		return nil, err
	}

	// Check if the requesting user has access to this community
	if !community.IsPublic {
		if requestingUserID == nil {
			return nil, fmt.Errorf("unauthorized access to private community")
		}
		isMember, err := s.communityRepo.IsMember(ctx, community.ID, *requestingUserID)
		if err != nil {
			return nil, err
		}
		if !isMember {
			return nil, fmt.Errorf("unauthorized access to private community")
		}
	}

	return community, nil
}

// ListCommunities retrieves all communities with optional filters
func (s *CommunityService) ListCommunities(ctx context.Context, page, limit int, sort string) ([]*models.Community, int, error) {
	offset := (page - 1) * limit
	return s.communityRepo.ListCommunities(ctx, limit, offset, true, sort)
}

// SearchCommunities searches communities by name
func (s *CommunityService) SearchCommunities(ctx context.Context, query string, page, limit int) ([]*models.Community, int, error) {
	offset := (page - 1) * limit
	return s.communityRepo.SearchCommunities(ctx, query, limit, offset)
}

// UpdateCommunity updates a community
func (s *CommunityService) UpdateCommunity(ctx context.Context, communityID, userID uuid.UUID, req *models.UpdateCommunityRequest) (*models.Community, error) {
	// Check if user has permission to update
	hasPermission, err := s.HasPermission(ctx, communityID, userID, models.CommunityRoleAdmin)
	if err != nil {
		return nil, err
	}
	if !hasPermission {
		return nil, fmt.Errorf("unauthorized to update this community")
	}

	community, err := s.communityRepo.GetCommunityByID(ctx, communityID)
	if err != nil {
		return nil, err
	}

	if req.Name != nil {
		newSlug := generateSlug(*req.Name)
		// Check if slug already exists (and is not the current community)
		existing, _ := s.communityRepo.GetCommunityBySlug(ctx, newSlug)
		if existing != nil && existing.ID != communityID {
			return nil, fmt.Errorf("community with this name already exists")
		}
		community.Name = *req.Name
		community.Slug = newSlug
	}
	if req.Description != nil {
		community.Description = req.Description
	}
	if req.Icon != nil {
		community.Icon = req.Icon
	}
	if req.IsPublic != nil {
		community.IsPublic = *req.IsPublic
	}
	if req.Rules != nil {
		community.Rules = req.Rules
	}

	err = s.communityRepo.UpdateCommunity(ctx, community)
	if err != nil {
		return nil, fmt.Errorf("failed to update community: %w", err)
	}

	return community, nil
}

// DeleteCommunity deletes a community
func (s *CommunityService) DeleteCommunity(ctx context.Context, communityID, userID uuid.UUID) error {
	// Only owner can delete
	community, err := s.communityRepo.GetCommunityByID(ctx, communityID)
	if err != nil {
		return err
	}

	if community.OwnerID != userID {
		return fmt.Errorf("only the community owner can delete the community")
	}

	return s.communityRepo.DeleteCommunity(ctx, communityID)
}

// JoinCommunity adds a user to a community
func (s *CommunityService) JoinCommunity(ctx context.Context, communityID, userID uuid.UUID) error {
	// Check if user is banned
	isBanned, err := s.communityRepo.IsBanned(ctx, communityID, userID)
	if err != nil {
		return err
	}
	if isBanned {
		return fmt.Errorf("you are banned from this community")
	}

	// Check if already a member
	isMember, err := s.communityRepo.IsMember(ctx, communityID, userID)
	if err != nil {
		return err
	}
	if isMember {
		return fmt.Errorf("already a member of this community")
	}

	member := &models.CommunityMember{
		ID:          uuid.New(),
		CommunityID: communityID,
		UserID:      userID,
		Role:        models.CommunityRoleMember,
		JoinedAt:    time.Now(),
	}

	return s.communityRepo.AddMember(ctx, member)
}

// LeaveCommunity removes a user from a community
func (s *CommunityService) LeaveCommunity(ctx context.Context, communityID, userID uuid.UUID) error {
	// Check if user is the owner
	community, err := s.communityRepo.GetCommunityByID(ctx, communityID)
	if err != nil {
		return err
	}

	if community.OwnerID == userID {
		return fmt.Errorf("owner cannot leave the community; transfer ownership or delete the community instead")
	}

	return s.communityRepo.RemoveMember(ctx, communityID, userID)
}

// GetMembers retrieves members of a community
func (s *CommunityService) GetMembers(ctx context.Context, communityID uuid.UUID, role string, page, limit int) ([]*models.CommunityMember, int, error) {
	offset := (page - 1) * limit
	return s.communityRepo.ListMembers(ctx, communityID, role, limit, offset)
}

// UpdateMemberRole updates a member's role in a community
func (s *CommunityService) UpdateMemberRole(ctx context.Context, communityID, requestingUserID, targetUserID uuid.UUID, newRole string) error {
	// Check if requesting user has permission (must be admin)
	hasPermission, err := s.HasPermission(ctx, communityID, requestingUserID, models.CommunityRoleAdmin)
	if err != nil {
		return err
	}
	if !hasPermission {
		return fmt.Errorf("unauthorized to update member roles")
	}

	// Cannot change the owner's role
	community, err := s.communityRepo.GetCommunityByID(ctx, communityID)
	if err != nil {
		return err
	}
	if community.OwnerID == targetUserID {
		return fmt.Errorf("cannot change the owner's role")
	}

	return s.communityRepo.UpdateMemberRole(ctx, communityID, targetUserID, newRole)
}

// BanMember bans a user from a community
func (s *CommunityService) BanMember(ctx context.Context, communityID, requestingUserID, targetUserID uuid.UUID, reason *string) error {
	// Check if requesting user has permission (must be admin or mod)
	hasPermission, err := s.HasPermission(ctx, communityID, requestingUserID, models.CommunityRoleMod)
	if err != nil {
		return err
	}
	if !hasPermission {
		return fmt.Errorf("unauthorized to ban members")
	}

	// Cannot ban the owner
	community, err := s.communityRepo.GetCommunityByID(ctx, communityID)
	if err != nil {
		return err
	}
	if community.OwnerID == targetUserID {
		return fmt.Errorf("cannot ban the community owner")
	}

	// Check role hierarchy: cannot ban users with equal or higher role
	requestingMember, err := s.communityRepo.GetMember(ctx, communityID, requestingUserID)
	if err != nil {
		return err
	}
	if requestingMember == nil {
		return fmt.Errorf("requesting user is not a member")
	}

	targetMember, err := s.communityRepo.GetMember(ctx, communityID, targetUserID)
	if err != nil {
		return err
	}
	if targetMember != nil {
		requestingRoleLevel := roleHierarchy[requestingMember.Role]
		targetRoleLevel := roleHierarchy[targetMember.Role]

		if targetRoleLevel >= requestingRoleLevel {
			return fmt.Errorf("cannot ban users with equal or higher role")
		}
	}

	// Remove user from community if they are a member
	_ = s.communityRepo.RemoveMember(ctx, communityID, targetUserID)

	// Add to ban list
	ban := &models.CommunityBan{
		ID:             uuid.New(),
		CommunityID:    communityID,
		BannedUserID:   targetUserID,
		BannedByUserID: &requestingUserID,
		Reason:         reason,
		BannedAt:       time.Now(),
	}

	return s.communityRepo.BanMember(ctx, ban)
}

// UnbanMember unbans a user from a community
func (s *CommunityService) UnbanMember(ctx context.Context, communityID, requestingUserID, targetUserID uuid.UUID) error {
	// Check if requesting user has permission (must be admin or mod)
	hasPermission, err := s.HasPermission(ctx, communityID, requestingUserID, models.CommunityRoleMod)
	if err != nil {
		return err
	}
	if !hasPermission {
		return fmt.Errorf("unauthorized to unban members")
	}

	return s.communityRepo.UnbanMember(ctx, communityID, targetUserID)
}

// GetBannedMembers retrieves banned members of a community
func (s *CommunityService) GetBannedMembers(ctx context.Context, communityID, requestingUserID uuid.UUID, page, limit int) ([]*models.CommunityBan, int, error) {
	// Check if requesting user has permission (must be admin or mod)
	hasPermission, err := s.HasPermission(ctx, communityID, requestingUserID, models.CommunityRoleMod)
	if err != nil {
		return nil, 0, err
	}
	if !hasPermission {
		return nil, 0, fmt.Errorf("unauthorized to view banned members")
	}

	offset := (page - 1) * limit
	return s.communityRepo.ListBans(ctx, communityID, limit, offset)
}

// AddClipToCommunity adds a clip to a community feed
func (s *CommunityService) AddClipToCommunity(ctx context.Context, communityID, userID, clipID uuid.UUID) error {
	// Check if user is a member
	isMember, err := s.communityRepo.IsMember(ctx, communityID, userID)
	if err != nil {
		return err
	}
	if !isMember {
		return fmt.Errorf("must be a member to post clips")
	}

	// Verify clip exists
	_, err = s.clipRepo.GetByID(ctx, clipID)
	if err != nil {
		return fmt.Errorf("clip not found")
	}

	communityClip := &models.CommunityClip{
		ID:            uuid.New(),
		CommunityID:   communityID,
		ClipID:        clipID,
		AddedByUserID: &userID,
		AddedAt:       time.Now(),
	}

	return s.communityRepo.AddClipToCommunity(ctx, communityClip)
}

// RemoveClipFromCommunity removes a clip from a community feed
func (s *CommunityService) RemoveClipFromCommunity(ctx context.Context, communityID, userID, clipID uuid.UUID) error {
	// Check if user has permission (must be admin or mod)
	hasPermission, err := s.HasPermission(ctx, communityID, userID, models.CommunityRoleMod)
	if err != nil {
		return err
	}
	if !hasPermission {
		return fmt.Errorf("unauthorized to remove clips")
	}

	// Get requesting user's member info
	requestingMember, err := s.communityRepo.GetMember(ctx, communityID, userID)
	if err != nil {
		return err
	}
	if requestingMember == nil {
		return fmt.Errorf("requesting user is not a member")
	}

	// Query to get who added this specific clip
	clips, _, err := s.communityRepo.GetCommunityClips(ctx, communityID, "recent", 100, 0)
	if err != nil {
		return fmt.Errorf("failed to get community clips: %w", err)
	}

	// Find the clip and check role hierarchy if needed
	for _, c := range clips {
		if c.ClipID == clipID {
			if c.AddedByUserID != nil {
				addedByMember, err := s.communityRepo.GetMember(ctx, communityID, *c.AddedByUserID)
				if err != nil {
					return err
				}
				if addedByMember != nil {
					requestingRoleLevel := roleHierarchy[requestingMember.Role]
					addedByRoleLevel := roleHierarchy[addedByMember.Role]

					// Cannot remove clips added by users with equal or higher role
					if addedByRoleLevel >= requestingRoleLevel {
						return fmt.Errorf("cannot remove clips added by users with equal or higher role")
					}
				}
			}
			break
		}
	}

	return s.communityRepo.RemoveClipFromCommunity(ctx, communityID, clipID)
}

// GetCommunityFeed retrieves the community feed
func (s *CommunityService) GetCommunityFeed(ctx context.Context, communityID uuid.UUID, sort string, page, limit int) ([]*models.CommunityClipWithClip, int, error) {
	offset := (page - 1) * limit
	return s.communityRepo.GetCommunityClips(ctx, communityID, sort, limit, offset)
}

// CreateDiscussion creates a new discussion thread
func (s *CommunityService) CreateDiscussion(ctx context.Context, communityID, userID uuid.UUID, req *models.CreateDiscussionRequest) (*models.CommunityDiscussion, error) {
	// Check if user is a member
	isMember, err := s.communityRepo.IsMember(ctx, communityID, userID)
	if err != nil {
		return nil, err
	}
	if !isMember {
		return nil, fmt.Errorf("must be a member to create discussions")
	}

	discussion := &models.CommunityDiscussion{
		ID:           uuid.New(),
		CommunityID:  communityID,
		UserID:       userID,
		Title:        req.Title,
		Content:      req.Content,
		IsPinned:     false,
		IsResolved:   false,
		VoteScore:    0,
		CommentCount: 0,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	err = s.communityRepo.CreateDiscussion(ctx, discussion)
	if err != nil {
		return nil, fmt.Errorf("failed to create discussion: %w", err)
	}

	return discussion, nil
}

// GetDiscussion retrieves a discussion thread
func (s *CommunityService) GetDiscussion(ctx context.Context, discussionID uuid.UUID) (*models.CommunityDiscussion, error) {
	return s.communityRepo.GetDiscussion(ctx, discussionID)
}

// ListDiscussions retrieves discussions for a community
func (s *CommunityService) ListDiscussions(ctx context.Context, communityID uuid.UUID, sort string, page, limit int) ([]*models.CommunityDiscussion, int, error) {
	offset := (page - 1) * limit
	return s.communityRepo.ListDiscussions(ctx, communityID, sort, limit, offset)
}

// UpdateDiscussion updates a discussion thread
func (s *CommunityService) UpdateDiscussion(ctx context.Context, discussionID, userID uuid.UUID, req *models.UpdateDiscussionRequest) (*models.CommunityDiscussion, error) {
	discussion, err := s.communityRepo.GetDiscussion(ctx, discussionID)
	if err != nil {
		return nil, err
	}

	// Check if user has permission to update (must be author, admin, or mod)
	isAuthor := discussion.UserID == userID
	hasModPermission, err := s.HasPermission(ctx, discussion.CommunityID, userID, models.CommunityRoleMod)
	if err != nil {
		return nil, err
	}

	if !isAuthor && !hasModPermission {
		return nil, fmt.Errorf("unauthorized to update this discussion")
	}

	// Regular members can only update title and content
	if req.Title != nil && (isAuthor || hasModPermission) {
		discussion.Title = *req.Title
	}
	if req.Content != nil && (isAuthor || hasModPermission) {
		discussion.Content = *req.Content
	}

	// Only mods/admins can pin or resolve
	if req.IsPinned != nil && hasModPermission {
		discussion.IsPinned = *req.IsPinned
	}
	if req.IsResolved != nil && hasModPermission {
		discussion.IsResolved = *req.IsResolved
	}

	err = s.communityRepo.UpdateDiscussion(ctx, discussion)
	if err != nil {
		return nil, fmt.Errorf("failed to update discussion: %w", err)
	}

	return discussion, nil
}

// DeleteDiscussion deletes a discussion thread
func (s *CommunityService) DeleteDiscussion(ctx context.Context, discussionID, userID uuid.UUID) error {
	discussion, err := s.communityRepo.GetDiscussion(ctx, discussionID)
	if err != nil {
		return err
	}

	// Check if user has permission (must be author, admin, or mod)
	isAuthor := discussion.UserID == userID
	hasModPermission, err := s.HasPermission(ctx, discussion.CommunityID, userID, models.CommunityRoleMod)
	if err != nil {
		return err
	}

	if !isAuthor && !hasModPermission {
		return fmt.Errorf("unauthorized to delete this discussion")
	}

	return s.communityRepo.DeleteDiscussion(ctx, discussionID)
}

// HasPermission checks if a user has a specific role or higher in a community
func (s *CommunityService) HasPermission(ctx context.Context, communityID, userID uuid.UUID, requiredRole string) (bool, error) {
	member, err := s.communityRepo.GetMember(ctx, communityID, userID)
	if err != nil {
		return false, err
	}
	if member == nil {
		return false, nil // Not a member
	}

	// Role hierarchy: admin > mod > member
	roleHierarchy := map[string]int{
		models.CommunityRoleAdmin:  3,
		models.CommunityRoleMod:    2,
		models.CommunityRoleMember: 1,
	}

	userRoleLevel := roleHierarchy[member.Role]
	requiredRoleLevel := roleHierarchy[requiredRole]

	return userRoleLevel >= requiredRoleLevel, nil
}

// GetUserRole retrieves the role of a user in a community
func (s *CommunityService) GetUserRole(ctx context.Context, communityID, userID uuid.UUID) (*string, error) {
	member, err := s.communityRepo.GetMember(ctx, communityID, userID)
	if err != nil {
		return nil, err
	}
	if member == nil {
		return nil, nil // Not a member
	}
	return &member.Role, nil
}
