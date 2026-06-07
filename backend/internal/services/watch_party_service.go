package services

import (
	"context"
	"crypto/rand"
	"fmt"

	"github.com/google/uuid"
	"git.subcult.tv/subculture-collective/clpr/internal/models"
	"git.subcult.tv/subculture-collective/clpr/internal/repository"
)

// WatchPartyService handles business logic for watch parties
type WatchPartyService struct {
	watchPartyRepo *repository.WatchPartyRepository
	playlistRepo   *repository.PlaylistRepository
	clipRepo       *repository.ClipRepository
	baseURL        string
}

// NewWatchPartyService creates a new WatchPartyService
func NewWatchPartyService(
	watchPartyRepo *repository.WatchPartyRepository,
	playlistRepo *repository.PlaylistRepository,
	clipRepo *repository.ClipRepository,
	baseURL string,
) *WatchPartyService {
	return &WatchPartyService{
		watchPartyRepo: watchPartyRepo,
		playlistRepo:   playlistRepo,
		clipRepo:       clipRepo,
		baseURL:        baseURL,
	}
}

// CreateWatchParty creates a new watch party
func (s *WatchPartyService) CreateWatchParty(ctx context.Context, userID uuid.UUID, req *models.CreateWatchPartyRequest) (*models.WatchParty, error) {
	// Set default visibility if not provided
	visibility := "private"
	if req.Visibility != "" {
		visibility = req.Visibility
	}

	// Set default max participants if not provided
	maxParticipants := 100
	if req.MaxParticipants != nil {
		maxParticipants = *req.MaxParticipants
	}

	// Validate playlist if provided
	if req.PlaylistID != nil {
		playlist, err := s.playlistRepo.GetByID(ctx, *req.PlaylistID)
		if err != nil {
			return nil, fmt.Errorf("failed to validate playlist: %w", err)
		}
		if playlist == nil {
			return nil, fmt.Errorf("playlist not found")
		}
		// Check if user has access to the playlist
		if playlist.Visibility == "private" && playlist.UserID != userID {
			return nil, fmt.Errorf("unauthorized: cannot use private playlist")
		}
	}

	// Generate unique invite code
	inviteCode, err := s.generateInviteCode(6)
	if err != nil {
		return nil, fmt.Errorf("failed to generate invite code: %w", err)
	}

	party := &models.WatchParty{
		ID:              uuid.New(),
		HostUserID:      userID,
		Title:           req.Title,
		PlaylistID:      req.PlaylistID,
		Visibility:      visibility,
		InviteCode:      inviteCode,
		MaxParticipants: maxParticipants,
	}

	err = s.watchPartyRepo.Create(ctx, party)
	if err != nil {
		return nil, fmt.Errorf("failed to create watch party: %w", err)
	}

	// Add host as first participant
	participant := &models.WatchPartyParticipant{
		ID:      uuid.New(),
		PartyID: party.ID,
		UserID:  userID,
		Role:    "host",
	}

	err = s.watchPartyRepo.AddParticipant(ctx, participant)
	if err != nil {
		return nil, fmt.Errorf("failed to add host as participant: %w", err)
	}

	return party, nil
}

// GetWatchParty retrieves a watch party by ID
func (s *WatchPartyService) GetWatchParty(ctx context.Context, partyID uuid.UUID, userID *uuid.UUID) (*models.WatchParty, error) {
	party, err := s.watchPartyRepo.GetByID(ctx, partyID)
	if err != nil {
		return nil, fmt.Errorf("failed to get watch party: %w", err)
	}
	if party == nil {
		return nil, fmt.Errorf("watch party not found")
	}

	// Check visibility permissions
	if party.Visibility == "private" {
		if userID == nil {
			return nil, fmt.Errorf("unauthorized: party is private")
		}
		// Check if user is a participant
		participant, err := s.watchPartyRepo.GetParticipant(ctx, partyID, *userID)
		if err != nil {
			return nil, fmt.Errorf("failed to check participant status: %w", err)
		}
		if participant == nil || participant.LeftAt != nil {
			return nil, fmt.Errorf("unauthorized: not a participant")
		}
	}

	// Load participants
	participants, err := s.watchPartyRepo.GetActiveParticipants(ctx, partyID)
	if err != nil {
		return nil, fmt.Errorf("failed to get participants: %w", err)
	}
	party.Participants = participants

	return party, nil
}

// JoinWatchParty allows a user to join a watch party
func (s *WatchPartyService) JoinWatchParty(ctx context.Context, inviteCode string, userID uuid.UUID) (*models.WatchParty, error) {
	// Get party by invite code
	party, err := s.watchPartyRepo.GetByInviteCode(ctx, inviteCode)
	if err != nil {
		return nil, fmt.Errorf("failed to find watch party: %w", err)
	}
	if party == nil {
		return nil, fmt.Errorf("watch party not found or has ended")
	}

	// Check participant limit
	activeCount, err := s.watchPartyRepo.GetActiveParticipantCount(ctx, party.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to check participant count: %w", err)
	}

	// Check if user is already a participant
	existingParticipant, err := s.watchPartyRepo.GetParticipant(ctx, party.ID, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to check existing participant: %w", err)
	}

	// If not rejoining, check participant limit
	if existingParticipant == nil || existingParticipant.LeftAt != nil {
		if activeCount >= party.MaxParticipants {
			return nil, fmt.Errorf("party is full: %d/%d participants", activeCount, party.MaxParticipants)
		}
	}

	// Add user as participant
	participant := &models.WatchPartyParticipant{
		ID:      uuid.New(),
		PartyID: party.ID,
		UserID:  userID,
		Role:    "viewer",
	}

	err = s.watchPartyRepo.AddParticipant(ctx, participant)
	if err != nil {
		return nil, fmt.Errorf("failed to add participant: %w", err)
	}

	// Start party if this is the first join and it hasn't started yet
	if party.StartedAt == nil {
		err = s.watchPartyRepo.StartParty(ctx, party.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to start party: %w", err)
		}
	}

	// Reload party with participants
	party, err = s.GetWatchParty(ctx, party.ID, &userID)
	if err != nil {
		return nil, fmt.Errorf("failed to reload party: %w", err)
	}

	return party, nil
}

// EndWatchParty ends a watch party
func (s *WatchPartyService) EndWatchParty(ctx context.Context, partyID, userID uuid.UUID) error {
	// Verify user is the host
	party, err := s.watchPartyRepo.GetByID(ctx, partyID)
	if err != nil {
		return fmt.Errorf("failed to get watch party: %w", err)
	}
	if party == nil {
		return fmt.Errorf("watch party not found")
	}

	if party.HostUserID != userID {
		return fmt.Errorf("unauthorized: only host can end the party")
	}

	err = s.watchPartyRepo.EndParty(ctx, partyID)
	if err != nil {
		return fmt.Errorf("failed to end watch party: %w", err)
	}

	return nil
}

// LeaveWatchParty removes a user from a watch party
func (s *WatchPartyService) LeaveWatchParty(ctx context.Context, partyID, userID uuid.UUID) error {
	err := s.watchPartyRepo.RemoveParticipant(ctx, partyID, userID)
	if err != nil {
		return fmt.Errorf("failed to leave watch party: %w", err)
	}

	return nil
}

// UpdatePlaybackState updates the playback state of a watch party
func (s *WatchPartyService) UpdatePlaybackState(ctx context.Context, partyID uuid.UUID, isPlaying bool, position int) error {
	err := s.watchPartyRepo.UpdatePlaybackState(ctx, partyID, isPlaying, position)
	if err != nil {
		return fmt.Errorf("failed to update playback state: %w", err)
	}

	return nil
}

// SkipToClip updates the current clip in a watch party
func (s *WatchPartyService) SkipToClip(ctx context.Context, partyID, clipID uuid.UUID) error {
	// Verify clip exists
	clip, err := s.clipRepo.GetByID(ctx, clipID)
	if err != nil {
		return fmt.Errorf("failed to validate clip: %w", err)
	}
	if clip == nil {
		return fmt.Errorf("clip not found")
	}

	err = s.watchPartyRepo.UpdateCurrentClip(ctx, partyID, clipID, 0)
	if err != nil {
		return fmt.Errorf("failed to skip to clip: %w", err)
	}

	return nil
}

// VerifyHostOrCoHost checks if a user is a host or co-host of a party
func (s *WatchPartyService) VerifyHostOrCoHost(ctx context.Context, partyID, userID uuid.UUID) (bool, string, error) {
	participant, err := s.watchPartyRepo.GetParticipant(ctx, partyID, userID)
	if err != nil {
		return false, "", fmt.Errorf("failed to get participant: %w", err)
	}
	if participant == nil || participant.LeftAt != nil {
		return false, "", nil
	}

	isHostOrCoHost := participant.Role == "host" || participant.Role == "co-host"
	return isHostOrCoHost, participant.Role, nil
}

// GetInviteURL generates the full invite URL for a watch party
func (s *WatchPartyService) GetInviteURL(inviteCode string) string {
	return fmt.Sprintf("%s/watch-parties/%s", s.baseURL, inviteCode)
}

// generateInviteCode generates a random invite code
func (s *WatchPartyService) generateInviteCode(length int) (string, error) {
	const charset = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789"
	b := make([]byte, length)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}

	for i := range b {
		b[i] = charset[int(b[i])%len(charset)]
	}

	return string(b), nil
}
