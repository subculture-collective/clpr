package services

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"time"

	"github.com/google/uuid"
	"git.subcult.tv/subculture-collective/clpr/internal/models"
	"git.subcult.tv/subculture-collective/clpr/internal/repository"
)

// PlaylistService handles business logic for playlists
type PlaylistService struct {
	playlistRepo *repository.PlaylistRepository
	clipRepo     *repository.ClipRepository
	baseURL      string
}

// NewPlaylistService creates a new PlaylistService
func NewPlaylistService(playlistRepo *repository.PlaylistRepository, clipRepo *repository.ClipRepository, baseURL string) *PlaylistService {
	return &PlaylistService{
		playlistRepo: playlistRepo,
		clipRepo:     clipRepo,
		baseURL:      baseURL,
	}
}

// CreatePlaylist creates a new playlist
func (s *PlaylistService) CreatePlaylist(ctx context.Context, userID uuid.UUID, req *models.CreatePlaylistRequest) (*models.Playlist, error) {
	// Set default visibility if not provided
	visibility := models.PlaylistVisibilityPrivate
	if req.Visibility != nil {
		visibility = *req.Visibility
	}

	playlist := &models.Playlist{
		ID:          uuid.New(),
		UserID:      userID,
		Title:       req.Title,
		Description: req.Description,
		CoverURL:    req.CoverURL,
		Visibility:  visibility,
	}

	err := s.playlistRepo.Create(ctx, playlist)
	if err != nil {
		return nil, fmt.Errorf("failed to create playlist: %w", err)
	}

	return playlist, nil
}

// CopyPlaylist duplicates an existing playlist and its clips
func (s *PlaylistService) CopyPlaylist(ctx context.Context, playlistID, userID uuid.UUID, req *models.CopyPlaylistRequest) (*models.Playlist, error) {
	// Get source playlist
	source, err := s.playlistRepo.GetByID(ctx, playlistID)
	if err != nil {
		return nil, fmt.Errorf("failed to get playlist: %w", err)
	}
	if source == nil {
		return nil, fmt.Errorf("playlist not found")
	}

	// Check view permissions if playlist is private
	if source.Visibility == models.PlaylistVisibilityPrivate && source.UserID != userID {
		permission, err := s.playlistRepo.GetCollaboratorPermission(ctx, playlistID, userID)
		if err != nil {
			return nil, fmt.Errorf("failed to check permissions: %w", err)
		}
		if permission == "" {
			return nil, fmt.Errorf("unauthorized: user does not have permission to copy this playlist")
		}
	}

	// Build new playlist fields
	newTitle := fmt.Sprintf("Copy of %s", source.Title)
	if req != nil && req.Title != nil {
		newTitle = *req.Title
	}
	newDescription := source.Description
	if req != nil && req.Description != nil {
		newDescription = req.Description
	}
	newCoverURL := source.CoverURL
	if req != nil && req.CoverURL != nil {
		newCoverURL = req.CoverURL
	}

	visibility := models.PlaylistVisibilityPrivate
	if req != nil && req.Visibility != nil {
		visibility = *req.Visibility
	}

	playlist := &models.Playlist{
		ID:          uuid.New(),
		UserID:      userID,
		Title:       newTitle,
		Description: newDescription,
		CoverURL:    newCoverURL,
		Visibility:  visibility,
	}

	if err := s.playlistRepo.CreateWithItemsCopy(ctx, playlist, playlistID); err != nil {
		return nil, fmt.Errorf("failed to copy playlist: %w", err)
	}

	return playlist, nil
}

// GetPlaylist retrieves a playlist by ID with clips and additional data
func (s *PlaylistService) GetPlaylist(ctx context.Context, playlistID uuid.UUID, userID *uuid.UUID, page, limit int) (*models.PlaylistWithClips, error) {
	// Get the playlist
	playlist, err := s.playlistRepo.GetByID(ctx, playlistID)
	if err != nil {
		return nil, fmt.Errorf("failed to get playlist: %w", err)
	}
	if playlist == nil {
		return nil, fmt.Errorf("playlist not found")
	}

	// Check visibility permissions
	var currentPermission *string
	if userID != nil && *userID != playlist.UserID {
		permission, err := s.playlistRepo.GetCollaboratorPermission(ctx, playlistID, *userID)
		if err != nil {
			return nil, fmt.Errorf("failed to check permissions: %w", err)
		}
		if permission != "" {
			currentPermission = &permission
		}
	}

	if playlist.Visibility == models.PlaylistVisibilityPrivate {
		if userID == nil {
			return nil, fmt.Errorf("unauthorized: playlist is private")
		}
		if *userID != playlist.UserID && currentPermission == nil {
			return nil, fmt.Errorf("unauthorized: playlist is private")
		}
	}

	// Get clip count
	clipCount, err := s.playlistRepo.GetClipCount(ctx, playlistID)
	if err != nil {
		return nil, fmt.Errorf("failed to get clip count: %w", err)
	}

	// Get clips with pagination
	offset := (page - 1) * limit
	clips, _, err := s.playlistRepo.GetClips(ctx, playlistID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get clips: %w", err)
	}

	// Get creator information
	creator, err := s.playlistRepo.GetCreator(ctx, playlist.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to get creator: %w", err)
	}

	// Check if user has liked/bookmarked the playlist
	isLiked := false
	isBookmarked := false
	if userID != nil {
		isLiked, err = s.playlistRepo.IsLiked(ctx, *userID, playlistID)
		if err != nil {
			return nil, fmt.Errorf("failed to check if liked: %w", err)
		}
		isBookmarked, err = s.playlistRepo.IsBookmarked(ctx, *userID, playlistID)
		if err != nil {
			return nil, fmt.Errorf("failed to check if bookmarked: %w", err)
		}
	}

	result := &models.PlaylistWithClips{
		Playlist:  *playlist,
		ClipCount: clipCount,
		Clips:     clips,
		IsLiked:   isLiked,
		IsBookmarked: isBookmarked,
		Creator:   creator,
	}
	result.CurrentUserPermission = currentPermission

	return result, nil
}

// GetPlaylistByShareToken retrieves a playlist by share token (public or private link access)
func (s *PlaylistService) GetPlaylistByShareToken(ctx context.Context, shareToken string, userID *uuid.UUID, page, limit int) (*models.PlaylistWithClips, error) {
	playlist, err := s.playlistRepo.GetByShareToken(ctx, shareToken)
	if err != nil {
		return nil, fmt.Errorf("failed to get playlist: %w", err)
	}
	if playlist == nil {
		return nil, fmt.Errorf("playlist not found")
	}

	// Determine current user permission if available
	var currentPermission *string
	if userID != nil && *userID != playlist.UserID {
		permission, err := s.playlistRepo.GetCollaboratorPermission(ctx, playlist.ID, *userID)
		if err != nil {
			return nil, fmt.Errorf("failed to check permissions: %w", err)
		}
		if permission != "" {
			currentPermission = &permission
		}
	}

	// Get clip count
	clipCount, err := s.playlistRepo.GetClipCount(ctx, playlist.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get clip count: %w", err)
	}

	// Get clips with pagination
	offset := (page - 1) * limit
	clips, _, err := s.playlistRepo.GetClips(ctx, playlist.ID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get clips: %w", err)
	}

	// Get creator information
	creator, err := s.playlistRepo.GetCreator(ctx, playlist.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to get creator: %w", err)
	}

	// Check if user has liked/bookmarked the playlist
	isLiked := false
	isBookmarked := false
	if userID != nil {
		isLiked, err = s.playlistRepo.IsLiked(ctx, *userID, playlist.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to check if liked: %w", err)
		}
		isBookmarked, err = s.playlistRepo.IsBookmarked(ctx, *userID, playlist.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to check if bookmarked: %w", err)
		}
	}

	result := &models.PlaylistWithClips{
		Playlist:  *playlist,
		ClipCount: clipCount,
		Clips:     clips,
		IsLiked:   isLiked,
		IsBookmarked: isBookmarked,
		Creator:   creator,
	}
	result.CurrentUserPermission = currentPermission

	return result, nil
}

// UpdatePlaylist updates a playlist
func (s *PlaylistService) UpdatePlaylist(ctx context.Context, playlistID, userID uuid.UUID, req *models.UpdatePlaylistRequest) (*models.Playlist, error) {
	// Get the playlist to verify ownership or permission
	playlist, err := s.playlistRepo.GetByID(ctx, playlistID)
	if err != nil {
		return nil, fmt.Errorf("failed to get playlist: %w", err)
	}
	if playlist == nil {
		return nil, fmt.Errorf("playlist not found")
	}

	// Check edit permission (owner or edit/admin collaborator)
	hasPermission, err := s.CheckEditPermission(ctx, playlistID, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to check permissions: %w", err)
	}
	if !hasPermission {
		return nil, fmt.Errorf("unauthorized: user does not have permission to edit this playlist")
	}

	// Update fields if provided
	if req.Title != nil {
		playlist.Title = *req.Title
	}
	if req.Description != nil {
		playlist.Description = req.Description
	}
	if req.CoverURL != nil {
		playlist.CoverURL = req.CoverURL
	}
	if req.Visibility != nil {
		// Only owner can change visibility
		if playlist.UserID != userID {
			return nil, fmt.Errorf("unauthorized: only the owner can change playlist visibility")
		}

		oldVisibility := playlist.Visibility
		playlist.Visibility = *req.Visibility

		// Generate share token if making public/unlisted from private
		if oldVisibility == models.PlaylistVisibilityPrivate &&
			(*req.Visibility == models.PlaylistVisibilityPublic || *req.Visibility == models.PlaylistVisibilityUnlisted) {
			if playlist.ShareToken == nil || *playlist.ShareToken == "" {
				token, err := generateShareToken()
				if err != nil {
					return nil, fmt.Errorf("failed to generate share token: %w", err)
				}
				playlist.ShareToken = &token
			}
		}
	}

	err = s.playlistRepo.Update(ctx, playlist)
	if err != nil {
		return nil, fmt.Errorf("failed to update playlist: %w", err)
	}

	return playlist, nil
}

// DeletePlaylist soft deletes a playlist
func (s *PlaylistService) DeletePlaylist(ctx context.Context, playlistID, userID uuid.UUID) error {
	// Get the playlist to verify ownership
	playlist, err := s.playlistRepo.GetByID(ctx, playlistID)
	if err != nil {
		return fmt.Errorf("failed to get playlist: %w", err)
	}
	if playlist == nil {
		return fmt.Errorf("playlist not found")
	}

	// Verify ownership
	if playlist.UserID != userID {
		return fmt.Errorf("unauthorized: user does not own this playlist")
	}

	err = s.playlistRepo.SoftDelete(ctx, playlistID)
	if err != nil {
		return fmt.Errorf("failed to delete playlist: %w", err)
	}

	return nil
}

// ListUserPlaylists retrieves playlists owned by a user
func (s *PlaylistService) ListUserPlaylists(ctx context.Context, userID uuid.UUID, page, limit int) ([]*models.PlaylistListItem, int, error) {
	offset := (page - 1) * limit
	playlists, total, err := s.playlistRepo.ListByUserID(ctx, userID, &userID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list playlists: %w", err)
	}

	return playlists, total, nil
}

// ListPublicPlaylists retrieves public playlists for discovery
func (s *PlaylistService) ListPublicPlaylists(ctx context.Context, userID *uuid.UUID, page, limit int) ([]*models.PlaylistListItem, int, error) {
	offset := (page - 1) * limit
	playlists, total, err := s.playlistRepo.ListPublic(ctx, userID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list public playlists: %w", err)
	}

	return playlists, total, nil
}

// ListBookmarkedPlaylists retrieves playlists bookmarked by a user
func (s *PlaylistService) ListBookmarkedPlaylists(ctx context.Context, userID uuid.UUID, page, limit int) ([]*models.PlaylistListItem, int, error) {
	offset := (page - 1) * limit
	playlists, total, err := s.playlistRepo.ListBookmarkedByUser(ctx, userID, &userID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list bookmarked playlists: %w", err)
	}

	return playlists, total, nil
}

// ListFeaturedPlaylists returns featured/curated playlists for public discovery.
func (s *PlaylistService) ListFeaturedPlaylists(ctx context.Context, userID *uuid.UUID, page, limit int) ([]*models.PlaylistListItem, int, error) {
	offset := (page - 1) * limit
	playlists, total, err := s.playlistRepo.ListFeatured(ctx, userID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list featured playlists: %w", err)
	}
	return playlists, total, nil
}

// GetPlaylistOfTheDay returns the most recent daily-generated playlist.
func (s *PlaylistService) GetPlaylistOfTheDay(ctx context.Context, userID *uuid.UUID) (*models.PlaylistListItem, error) {
	return s.playlistRepo.GetPlaylistOfTheDay(ctx, userID)
}

// AddClipsToPlaylist adds multiple clips to a playlist
func (s *PlaylistService) AddClipsToPlaylist(ctx context.Context, playlistID, userID uuid.UUID, clipIDs []uuid.UUID) error {
	// Get the playlist to verify ownership
	playlist, err := s.playlistRepo.GetByID(ctx, playlistID)
	if err != nil {
		return fmt.Errorf("failed to get playlist: %w", err)
	}
	if playlist == nil {
		return fmt.Errorf("playlist not found")
	}

	// Verify edit permission (owner or edit/admin collaborator)
	hasPermission, err := s.CheckEditPermission(ctx, playlistID, userID)
	if err != nil {
		return err
	}
	if !hasPermission {
		return fmt.Errorf("unauthorized: user does not have permission to edit this playlist")
	}

	// Get current clip count
	currentCount, err := s.playlistRepo.GetClipCount(ctx, playlistID)
	if err != nil {
		return fmt.Errorf("failed to get clip count: %w", err)
	}

	// Add clips one by one, checking for duplicates and existence
	addedCount := 0
	for _, clipID := range clipIDs {
		// Check if clip exists
		clip, err := s.clipRepo.GetByID(ctx, clipID)
		if err != nil {
			return fmt.Errorf("failed to check clip existence: %w", err)
		}
		if clip == nil {
			return fmt.Errorf("clip %s not found", clipID)
		}

		// Check if clip is already in playlist
		exists, err := s.playlistRepo.HasClip(ctx, playlistID, clipID)
		if err != nil {
			return fmt.Errorf("failed to check if clip exists in playlist: %w", err)
		}
		if exists {
			// Skip duplicate clips
			continue
		}

		// Check if adding this clip would exceed the limit
		if currentCount+addedCount >= 1000 {
			return fmt.Errorf("playlist cannot exceed 1000 clips")
		}

		// Add clip with order index based on actual position
		orderIndex := currentCount + addedCount
		err = s.playlistRepo.AddClip(ctx, playlistID, clipID, orderIndex)
		if err != nil {
			return fmt.Errorf("failed to add clip to playlist: %w", err)
		}
		addedCount++
	}

	return nil
}

// RemoveClipFromPlaylist removes a clip from a playlist
func (s *PlaylistService) RemoveClipFromPlaylist(ctx context.Context, playlistID, clipID, userID uuid.UUID) error {
	// Get the playlist to verify ownership
	playlist, err := s.playlistRepo.GetByID(ctx, playlistID)
	if err != nil {
		return fmt.Errorf("failed to get playlist: %w", err)
	}
	if playlist == nil {
		return fmt.Errorf("playlist not found")
	}

	// Verify edit permission (owner or edit/admin collaborator)
	hasPermission, err := s.CheckEditPermission(ctx, playlistID, userID)
	if err != nil {
		return err
	}
	if !hasPermission {
		return fmt.Errorf("unauthorized: user does not have permission to edit this playlist")
	}

	err = s.playlistRepo.RemoveClip(ctx, playlistID, clipID)
	if err != nil {
		return fmt.Errorf("failed to remove clip from playlist: %w", err)
	}

	return nil
}

// ReorderPlaylistClips updates the order of clips in a playlist
func (s *PlaylistService) ReorderPlaylistClips(ctx context.Context, playlistID, userID uuid.UUID, clipIDs []uuid.UUID) error {
	// Get the playlist to verify ownership
	playlist, err := s.playlistRepo.GetByID(ctx, playlistID)
	if err != nil {
		return fmt.Errorf("failed to get playlist: %w", err)
	}
	if playlist == nil {
		return fmt.Errorf("playlist not found")
	}

	// Verify edit permission (owner or edit/admin collaborator)
	hasPermission, err := s.CheckEditPermission(ctx, playlistID, userID)
	if err != nil {
		return err
	}
	if !hasPermission {
		return fmt.Errorf("unauthorized: user does not have permission to edit this playlist")
	}

	// Verify all clips exist in the playlist
	for _, clipID := range clipIDs {
		exists, err := s.playlistRepo.HasClip(ctx, playlistID, clipID)
		if err != nil {
			return fmt.Errorf("failed to check if clip exists in playlist: %w", err)
		}
		if !exists {
			return fmt.Errorf("clip %s not found in playlist", clipID)
		}
	}

	err = s.playlistRepo.ReorderClips(ctx, playlistID, clipIDs)
	if err != nil {
		return fmt.Errorf("failed to reorder clips: %w", err)
	}

	return nil
}

// LikePlaylist adds a like to a playlist
func (s *PlaylistService) LikePlaylist(ctx context.Context, playlistID, userID uuid.UUID) error {
	// Verify playlist exists and is not private
	playlist, err := s.playlistRepo.GetByID(ctx, playlistID)
	if err != nil {
		return fmt.Errorf("failed to get playlist: %w", err)
	}
	if playlist == nil {
		return fmt.Errorf("playlist not found")
	}

	// Can't like private playlists unless you own them
	if playlist.Visibility == models.PlaylistVisibilityPrivate && playlist.UserID != userID {
		return fmt.Errorf("cannot like private playlists")
	}

	err = s.playlistRepo.LikePlaylist(ctx, userID, playlistID)
	if err != nil {
		return fmt.Errorf("failed to like playlist: %w", err)
	}

	return nil
}

// UnlikePlaylist removes a like from a playlist
func (s *PlaylistService) UnlikePlaylist(ctx context.Context, playlistID, userID uuid.UUID) error {
	err := s.playlistRepo.UnlikePlaylist(ctx, userID, playlistID)
	if err != nil {
		return fmt.Errorf("failed to unlike playlist: %w", err)
	}

	return nil
}

// BookmarkPlaylist adds a bookmark to a playlist.
func (s *PlaylistService) BookmarkPlaylist(ctx context.Context, playlistID, userID uuid.UUID) error {
	playlist, err := s.playlistRepo.GetByID(ctx, playlistID)
	if err != nil {
		return fmt.Errorf("failed to get playlist: %w", err)
	}
	if playlist == nil {
		return fmt.Errorf("playlist not found")
	}

	if playlist.Visibility == models.PlaylistVisibilityPrivate && playlist.UserID != userID {
		return fmt.Errorf("cannot bookmark private playlists")
	}

	err = s.playlistRepo.BookmarkPlaylist(ctx, userID, playlistID)
	if err != nil {
		return fmt.Errorf("failed to bookmark playlist: %w", err)
	}

	return nil
}

// UnbookmarkPlaylist removes a bookmark from a playlist.
func (s *PlaylistService) UnbookmarkPlaylist(ctx context.Context, playlistID, userID uuid.UUID) error {
	err := s.playlistRepo.UnbookmarkPlaylist(ctx, userID, playlistID)
	if err != nil {
		return fmt.Errorf("failed to unbookmark playlist: %w", err)
	}

	return nil
}

// generateShareToken generates a URL-safe random share token
func generateShareToken() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("failed to generate random token: %w", err)
	}
	return base64.URLEncoding.EncodeToString(b)[:22], nil
}

// GetShareLink generates or retrieves the share link for a playlist
func (s *PlaylistService) GetShareLink(ctx context.Context, playlistID, userID uuid.UUID) (*models.GetShareLinkResponse, error) {
	// Get the playlist
	playlist, err := s.playlistRepo.GetByID(ctx, playlistID)
	if err != nil {
		return nil, fmt.Errorf("failed to get playlist: %w", err)
	}
	if playlist == nil {
		return nil, fmt.Errorf("playlist not found")
	}

	// Check permissions: owner or collaborator with edit/admin permission
	hasPermission := false
	if playlist.UserID == userID {
		hasPermission = true
	} else {
		permission, err := s.playlistRepo.GetCollaboratorPermission(ctx, playlistID, userID)
		if err != nil {
			return nil, fmt.Errorf("failed to check permissions: %w", err)
		}
		if permission == models.PlaylistPermissionEdit || permission == models.PlaylistPermissionAdmin {
			hasPermission = true
		}
	}

	if !hasPermission {
		return nil, fmt.Errorf("unauthorized: user does not have permission to share this playlist")
	}

	// Generate share token if it doesn't exist with retry logic for uniqueness
	shareToken := ""
	if playlist.ShareToken != nil && *playlist.ShareToken != "" {
		shareToken = *playlist.ShareToken
	} else {
		const maxShareTokenAttempts = 5
		var lastErr error
		for attempt := 0; attempt < maxShareTokenAttempts; attempt++ {
			candidate, err := generateShareToken()
			if err != nil {
				return nil, fmt.Errorf("failed to generate share token: %w", err)
			}
			if err = s.playlistRepo.UpdateShareToken(ctx, playlistID, candidate); err == nil {
				shareToken = candidate
				break
			}
			lastErr = err
		}
		if shareToken == "" {
			return nil, fmt.Errorf("failed to update share token after %d attempts: %w", maxShareTokenAttempts, lastErr)
		}
	}

	// Build share URL and embed code
	shareURL := fmt.Sprintf("%s/playlists/%s", s.baseURL, shareToken)
	embedCode := fmt.Sprintf(`<iframe src="%s/embed/playlist/%s" width="800" height="600" frameborder="0" allowfullscreen></iframe>`, s.baseURL, shareToken)

	return &models.GetShareLinkResponse{
		ShareURL:  shareURL,
		EmbedCode: embedCode,
	}, nil
}

// TrackShare records a share event for analytics
func (s *PlaylistService) TrackShare(ctx context.Context, playlistID uuid.UUID, platform, referrer string) error {
	// Verify playlist exists and is shareable
	playlist, err := s.playlistRepo.GetByID(ctx, playlistID)
	if err != nil {
		return fmt.Errorf("failed to get playlist: %w", err)
	}
	if playlist == nil {
		return fmt.Errorf("playlist not found")
	}

	// Only track shares for public/unlisted playlists
	if playlist.Visibility == models.PlaylistVisibilityPrivate {
		return fmt.Errorf("cannot track shares for private playlists")
	}

	// Validate referrer length (database constraint is VARCHAR(255))
	if len(referrer) > 255 {
		referrer = referrer[:255]
	}

	// Track the share event
	share := &models.PlaylistShare{
		ID:         uuid.New(),
		PlaylistID: playlistID,
		Platform:   &platform,
		SharedAt:   time.Now(),
	}
	if referrer != "" {
		share.Referrer = &referrer
	}

	err = s.playlistRepo.TrackShare(ctx, share)
	if err != nil {
		return fmt.Errorf("failed to track share: %w", err)
	}

	// Increment share count
	err = s.playlistRepo.IncrementShareCount(ctx, playlistID)
	if err != nil {
		return fmt.Errorf("failed to increment share count: %w", err)
	}

	return nil
}

// AddCollaborator adds a collaborator to a playlist
func (s *PlaylistService) AddCollaborator(ctx context.Context, playlistID, userID, collaboratorUserID uuid.UUID, permission string) error {
	// Check if user has permission to add collaborators (owner or admin)
	hasPermission, err := s.CheckManageCollaboratorsPermission(ctx, playlistID, userID)
	if err != nil {
		return err
	}
	if !hasPermission {
		return fmt.Errorf("unauthorized: user does not have permission to add collaborators")
	}

	// Get the playlist to check ownership
	playlist, err := s.playlistRepo.GetByID(ctx, playlistID)
	if err != nil {
		return fmt.Errorf("failed to get playlist: %w", err)
	}
	if playlist == nil {
		return fmt.Errorf("playlist not found")
	}

	// Cannot add the owner as a collaborator
	if collaboratorUserID == playlist.UserID {
		return fmt.Errorf("cannot add playlist owner as a collaborator")
	}

	// Create collaborator
	collaborator := &models.PlaylistCollaborator{
		ID:         uuid.New(),
		PlaylistID: playlistID,
		UserID:     collaboratorUserID,
		Permission: permission,
		InvitedBy:  &userID,
		InvitedAt:  time.Now(),
	}

	err = s.playlistRepo.AddCollaborator(ctx, collaborator)
	if err != nil {
		return fmt.Errorf("failed to add collaborator: %w", err)
	}

	return nil
}

// RemoveCollaborator removes a collaborator from a playlist
func (s *PlaylistService) RemoveCollaborator(ctx context.Context, playlistID, userID, collaboratorUserID uuid.UUID) error {
	// Check if user has permission to remove collaborators (owner or admin)
	hasPermission, err := s.CheckManageCollaboratorsPermission(ctx, playlistID, userID)
	if err != nil {
		return err
	}
	if !hasPermission {
		return fmt.Errorf("unauthorized: user does not have permission to remove collaborators")
	}

	err = s.playlistRepo.RemoveCollaborator(ctx, playlistID, collaboratorUserID)
	if err != nil {
		return fmt.Errorf("failed to remove collaborator: %w", err)
	}

	return nil
}

// GetCollaborators retrieves all collaborators for a playlist
func (s *PlaylistService) GetCollaborators(ctx context.Context, playlistID, userID uuid.UUID) ([]*models.PlaylistCollaborator, error) {
	// Get the playlist
	playlist, err := s.playlistRepo.GetByID(ctx, playlistID)
	if err != nil {
		return nil, fmt.Errorf("failed to get playlist: %w", err)
	}
	if playlist == nil {
		return nil, fmt.Errorf("playlist not found")
	}

	// Check if user has permission to view collaborators
	hasPermission := false
	if playlist.UserID == userID {
		hasPermission = true
	} else if playlist.Visibility != models.PlaylistVisibilityPrivate {
		// Anyone can see collaborators on public/unlisted playlists
		hasPermission = true
	} else {
		// Check if user is a collaborator
		isCollab, err := s.playlistRepo.IsCollaborator(ctx, playlistID, userID)
		if err != nil {
			return nil, fmt.Errorf("failed to check if collaborator: %w", err)
		}
		hasPermission = isCollab
	}

	if !hasPermission {
		return nil, fmt.Errorf("unauthorized: user does not have permission to view collaborators")
	}

	collaborators, err := s.playlistRepo.GetCollaborators(ctx, playlistID)
	if err != nil {
		return nil, fmt.Errorf("failed to get collaborators: %w", err)
	}

	return collaborators, nil
}

// UpdateCollaboratorPermission updates a collaborator's permission level
func (s *PlaylistService) UpdateCollaboratorPermission(ctx context.Context, playlistID, userID, collaboratorUserID uuid.UUID, permission string) error {
	// Check if user has permission to update collaborators (owner or admin)
	hasPermission, err := s.CheckManageCollaboratorsPermission(ctx, playlistID, userID)
	if err != nil {
		return err
	}
	if !hasPermission {
		return fmt.Errorf("unauthorized: user does not have permission to update collaborators")
	}

	// Update the collaborator's permission
	err = s.playlistRepo.UpdateCollaboratorPermission(ctx, playlistID, collaboratorUserID, permission)
	if err != nil {
		return fmt.Errorf("failed to update collaborator: %w", err)
	}

	return nil
}

// CheckEditPermission checks if a user has edit permission for a playlist
func (s *PlaylistService) CheckEditPermission(ctx context.Context, playlistID, userID uuid.UUID) (bool, error) {
	// Get the playlist
	playlist, err := s.playlistRepo.GetByID(ctx, playlistID)
	if err != nil {
		return false, fmt.Errorf("failed to get playlist: %w", err)
	}
	if playlist == nil {
		return false, fmt.Errorf("playlist not found")
	}

	// Owner always has edit permission
	if playlist.UserID == userID {
		return true, nil
	}

	// Check collaborator permission
	permission, err := s.playlistRepo.GetCollaboratorPermission(ctx, playlistID, userID)
	if err != nil {
		return false, fmt.Errorf("failed to check permissions: %w", err)
	}

	return permission == models.PlaylistPermissionEdit || permission == models.PlaylistPermissionAdmin, nil
}

// CheckManageCollaboratorsPermission checks if a user has permission to manage collaborators (owner or admin)
func (s *PlaylistService) CheckManageCollaboratorsPermission(ctx context.Context, playlistID, userID uuid.UUID) (bool, error) {
	// Get the playlist
	playlist, err := s.playlistRepo.GetByID(ctx, playlistID)
	if err != nil {
		return false, fmt.Errorf("failed to get playlist: %w", err)
	}
	if playlist == nil {
		return false, fmt.Errorf("playlist not found")
	}

	// Owner always has permission
	if playlist.UserID == userID {
		return true, nil
	}

	// Check if user is an admin collaborator
	permission, err := s.playlistRepo.GetCollaboratorPermission(ctx, playlistID, userID)
	if err != nil {
		return false, fmt.Errorf("failed to check permissions: %w", err)
	}

	return permission == models.PlaylistPermissionAdmin, nil
}
