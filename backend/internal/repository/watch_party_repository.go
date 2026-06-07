package repository

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"git.subcult.tv/subculture-collective/clpr/internal/models"
)

// WatchPartyRepository handles database operations for watch parties
type WatchPartyRepository struct {
	pool *pgxpool.Pool
}

// NewWatchPartyRepository creates a new WatchPartyRepository
func NewWatchPartyRepository(pool *pgxpool.Pool) *WatchPartyRepository {
	return &WatchPartyRepository{
		pool: pool,
	}
}

// Create creates a new watch party
func (r *WatchPartyRepository) Create(ctx context.Context, party *models.WatchParty) error {
	query := `
		INSERT INTO watch_parties (
			id, host_user_id, title, playlist_id, visibility, 
			invite_code, max_participants
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING created_at
	`

	err := r.pool.QueryRow(ctx, query,
		party.ID,
		party.HostUserID,
		party.Title,
		party.PlaylistID,
		party.Visibility,
		party.InviteCode,
		party.MaxParticipants,
	).Scan(&party.CreatedAt)

	if err != nil {
		return fmt.Errorf("failed to create watch party: %w", err)
	}

	return nil
}

// GetByID retrieves a watch party by its ID
func (r *WatchPartyRepository) GetByID(ctx context.Context, partyID uuid.UUID) (*models.WatchParty, error) {
	query := `
		SELECT id, host_user_id, title, playlist_id, current_clip_id,
		       current_position_seconds, is_playing, visibility, password, invite_code,
		       max_participants, created_at, started_at, ended_at
		FROM watch_parties
		WHERE id = $1
	`

	var party models.WatchParty
	err := r.pool.QueryRow(ctx, query, partyID).Scan(
		&party.ID,
		&party.HostUserID,
		&party.Title,
		&party.PlaylistID,
		&party.CurrentClipID,
		&party.CurrentPositionSeconds,
		&party.IsPlaying,
		&party.Visibility,
		&party.Password,
		&party.InviteCode,
		&party.MaxParticipants,
		&party.CreatedAt,
		&party.StartedAt,
		&party.EndedAt,
	)

	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get watch party: %w", err)
	}

	return &party, nil
}

// GetByInviteCode retrieves a watch party by its invite code
func (r *WatchPartyRepository) GetByInviteCode(ctx context.Context, inviteCode string) (*models.WatchParty, error) {
	query := `
		SELECT id, host_user_id, title, playlist_id, current_clip_id,
		       current_position_seconds, is_playing, visibility, password, invite_code,
		       max_participants, created_at, started_at, ended_at
		FROM watch_parties
		WHERE invite_code = $1 AND ended_at IS NULL
	`

	var party models.WatchParty
	err := r.pool.QueryRow(ctx, query, inviteCode).Scan(
		&party.ID,
		&party.HostUserID,
		&party.Title,
		&party.PlaylistID,
		&party.CurrentClipID,
		&party.CurrentPositionSeconds,
		&party.IsPlaying,
		&party.Visibility,
		&party.Password,
		&party.InviteCode,
		&party.MaxParticipants,
		&party.CreatedAt,
		&party.StartedAt,
		&party.EndedAt,
	)

	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get watch party by invite code: %w", err)
	}

	return &party, nil
}

// UpdatePlaybackState updates the playback state of a watch party
func (r *WatchPartyRepository) UpdatePlaybackState(ctx context.Context, partyID uuid.UUID, isPlaying bool, position int) error {
	query := `
		UPDATE watch_parties
		SET is_playing = $1, current_position_seconds = $2
		WHERE id = $3 AND ended_at IS NULL
	`

	_, err := r.pool.Exec(ctx, query, isPlaying, position, partyID)
	if err != nil {
		return fmt.Errorf("failed to update playback state: %w", err)
	}

	return nil
}

// UpdateCurrentClip updates the current clip being played
func (r *WatchPartyRepository) UpdateCurrentClip(ctx context.Context, partyID uuid.UUID, clipID uuid.UUID, position int) error {
	query := `
		UPDATE watch_parties
		SET current_clip_id = $1, current_position_seconds = $2
		WHERE id = $3 AND ended_at IS NULL
	`

	_, err := r.pool.Exec(ctx, query, clipID, position, partyID)
	if err != nil {
		return fmt.Errorf("failed to update current clip: %w", err)
	}

	return nil
}

// EndParty marks a watch party as ended
func (r *WatchPartyRepository) EndParty(ctx context.Context, partyID uuid.UUID) error {
	query := `
		UPDATE watch_parties
		SET ended_at = NOW()
		WHERE id = $1 AND ended_at IS NULL
	`

	_, err := r.pool.Exec(ctx, query, partyID)
	if err != nil {
		return fmt.Errorf("failed to end watch party: %w", err)
	}

	return nil
}

// GetActiveParticipantCount returns the count of active participants in a party
func (r *WatchPartyRepository) GetActiveParticipantCount(ctx context.Context, partyID uuid.UUID) (int, error) {
	query := `
		SELECT COUNT(*)
		FROM watch_party_participants
		WHERE party_id = $1 AND left_at IS NULL
	`

	var count int
	err := r.pool.QueryRow(ctx, query, partyID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to get active participant count: %w", err)
	}

	return count, nil
}

// AddParticipant adds a participant to a watch party or reactivates them if they left
func (r *WatchPartyRepository) AddParticipant(ctx context.Context, participant *models.WatchPartyParticipant) error {
	query := `
		INSERT INTO watch_party_participants (
			id, party_id, user_id, role
		)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (party_id, user_id)
		DO UPDATE SET left_at = NULL
		RETURNING joined_at
	`

	err := r.pool.QueryRow(ctx, query,
		participant.ID,
		participant.PartyID,
		participant.UserID,
		participant.Role,
	).Scan(&participant.JoinedAt)

	if err != nil {
		return fmt.Errorf("failed to add participant: %w", err)
	}

	return nil
}

// RemoveParticipant marks a participant as having left the party
func (r *WatchPartyRepository) RemoveParticipant(ctx context.Context, partyID, userID uuid.UUID) error {
	query := `
		UPDATE watch_party_participants
		SET left_at = NOW()
		WHERE party_id = $1 AND user_id = $2 AND left_at IS NULL
	`

	_, err := r.pool.Exec(ctx, query, partyID, userID)
	if err != nil {
		return fmt.Errorf("failed to remove participant: %w", err)
	}

	return nil
}

// GetParticipant retrieves a specific participant
func (r *WatchPartyRepository) GetParticipant(ctx context.Context, partyID, userID uuid.UUID) (*models.WatchPartyParticipant, error) {
	query := `
		SELECT id, party_id, user_id, role, joined_at, left_at, last_sync_at, sync_offset_ms
		FROM watch_party_participants
		WHERE party_id = $1 AND user_id = $2
		ORDER BY joined_at DESC
		LIMIT 1
	`

	var participant models.WatchPartyParticipant
	err := r.pool.QueryRow(ctx, query, partyID, userID).Scan(
		&participant.ID,
		&participant.PartyID,
		&participant.UserID,
		&participant.Role,
		&participant.JoinedAt,
		&participant.LeftAt,
		&participant.LastSyncAt,
		&participant.SyncOffsetMS,
	)

	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get participant: %w", err)
	}

	return &participant, nil
}

// GetActiveParticipants retrieves all active participants in a party
func (r *WatchPartyRepository) GetActiveParticipants(ctx context.Context, partyID uuid.UUID) ([]models.WatchPartyParticipant, error) {
	query := `
		SELECT wpp.id, wpp.party_id, wpp.user_id, wpp.role, 
		       wpp.joined_at, wpp.left_at, wpp.last_sync_at, wpp.sync_offset_ms,
		       u.username, u.display_name, u.avatar_url
		FROM watch_party_participants wpp
		JOIN users u ON u.id = wpp.user_id
		WHERE wpp.party_id = $1 AND wpp.left_at IS NULL
		ORDER BY wpp.joined_at ASC
	`

	rows, err := r.pool.Query(ctx, query, partyID)
	if err != nil {
		return nil, fmt.Errorf("failed to get active participants: %w", err)
	}
	defer rows.Close()

	var participants []models.WatchPartyParticipant
	for rows.Next() {
		var p models.WatchPartyParticipant
		var user models.User

		err := rows.Scan(
			&p.ID,
			&p.PartyID,
			&p.UserID,
			&p.Role,
			&p.JoinedAt,
			&p.LeftAt,
			&p.LastSyncAt,
			&p.SyncOffsetMS,
			&user.Username,
			&user.DisplayName,
			&user.AvatarURL,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan participant: %w", err)
		}

		user.ID = p.UserID
		p.User = &user
		participants = append(participants, p)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating participants: %w", err)
	}

	return participants, nil
}

// UpdateParticipantSync updates the last sync time for a participant
func (r *WatchPartyRepository) UpdateParticipantSync(ctx context.Context, partyID, userID uuid.UUID, offsetMS int) error {
	query := `
		UPDATE watch_party_participants
		SET last_sync_at = NOW(), sync_offset_ms = $1
		WHERE party_id = $2 AND user_id = $3 AND left_at IS NULL
	`

	_, err := r.pool.Exec(ctx, query, offsetMS, partyID, userID)
	if err != nil {
		return fmt.Errorf("failed to update participant sync: %w", err)
	}

	return nil
}

// StartParty marks a party as started
func (r *WatchPartyRepository) StartParty(ctx context.Context, partyID uuid.UUID) error {
	query := `
		UPDATE watch_parties
		SET started_at = NOW()
		WHERE id = $1 AND started_at IS NULL AND ended_at IS NULL
	`

	_, err := r.pool.Exec(ctx, query, partyID)
	if err != nil {
		return fmt.Errorf("failed to start party: %w", err)
	}

	return nil
}

// CleanupStaleParticipants removes participants who haven't synced recently
func (r *WatchPartyRepository) CleanupStaleParticipants(ctx context.Context, partyID uuid.UUID, staleDuration time.Duration) error {
	query := `
		UPDATE watch_party_participants
		SET left_at = NOW()
		WHERE party_id = $1 
		  AND left_at IS NULL
		  AND (last_sync_at IS NULL OR last_sync_at < NOW() - $2::interval)
	`

	_, err := r.pool.Exec(ctx, query, partyID, staleDuration)
	if err != nil {
		return fmt.Errorf("failed to cleanup stale participants: %w", err)
	}

	return nil
}

// CreateMessage creates a new chat message in a watch party
func (r *WatchPartyRepository) CreateMessage(ctx context.Context, message *models.WatchPartyMessage) error {
	query := `
		INSERT INTO watch_party_messages (id, watch_party_id, user_id, message)
		VALUES ($1, $2, $3, $4)
		RETURNING created_at
	`

	err := r.pool.QueryRow(ctx, query,
		message.ID,
		message.WatchPartyID,
		message.UserID,
		message.Message,
	).Scan(&message.CreatedAt)

	if err != nil {
		return fmt.Errorf("failed to create message: %w", err)
	}

	return nil
}

// GetMessages retrieves recent messages for a watch party
func (r *WatchPartyRepository) GetMessages(ctx context.Context, partyID uuid.UUID, limit int) ([]models.WatchPartyMessage, error) {
	query := `
		SELECT m.id, m.watch_party_id, m.user_id, m.message, m.created_at,
		       u.username, u.display_name, u.avatar_url
		FROM watch_party_messages m
		JOIN users u ON u.id = m.user_id
		WHERE m.watch_party_id = $1
		ORDER BY m.created_at DESC
		LIMIT $2
	`

	rows, err := r.pool.Query(ctx, query, partyID, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get messages: %w", err)
	}
	defer rows.Close()

	var messages []models.WatchPartyMessage
	for rows.Next() {
		var m models.WatchPartyMessage
		err := rows.Scan(
			&m.ID,
			&m.WatchPartyID,
			&m.UserID,
			&m.Message,
			&m.CreatedAt,
			&m.Username,
			&m.DisplayName,
			&m.AvatarURL,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan message: %w", err)
		}
		messages = append(messages, m)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating messages: %w", err)
	}

	// Reverse to get chronological order
	for i := 0; i < len(messages)/2; i++ {
		j := len(messages) - 1 - i
		messages[i], messages[j] = messages[j], messages[i]
	}

	return messages, nil
}

// CreateReaction creates a new emoji reaction in a watch party
func (r *WatchPartyRepository) CreateReaction(ctx context.Context, reaction *models.WatchPartyReaction) error {
	query := `
		INSERT INTO watch_party_reactions (id, watch_party_id, user_id, emoji, video_timestamp)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING created_at
	`

	err := r.pool.QueryRow(ctx, query,
		reaction.ID,
		reaction.WatchPartyID,
		reaction.UserID,
		reaction.Emoji,
		reaction.VideoTimestamp,
	).Scan(&reaction.CreatedAt)

	if err != nil {
		return fmt.Errorf("failed to create reaction: %w", err)
	}

	return nil
}

// GetRecentReactions retrieves recent reactions for a watch party
func (r *WatchPartyRepository) GetRecentReactions(ctx context.Context, partyID uuid.UUID, since time.Time) ([]models.WatchPartyReaction, error) {
	query := `
		SELECT r.id, r.watch_party_id, r.user_id, r.emoji, r.video_timestamp, r.created_at,
		       u.username
		FROM watch_party_reactions r
		JOIN users u ON u.id = r.user_id
		WHERE r.watch_party_id = $1 AND r.created_at >= $2
		ORDER BY r.created_at ASC
	`

	rows, err := r.pool.Query(ctx, query, partyID, since)
	if err != nil {
		return nil, fmt.Errorf("failed to get reactions: %w", err)
	}
	defer rows.Close()

	var reactions []models.WatchPartyReaction
	for rows.Next() {
		var r models.WatchPartyReaction
		err := rows.Scan(
			&r.ID,
			&r.WatchPartyID,
			&r.UserID,
			&r.Emoji,
			&r.VideoTimestamp,
			&r.CreatedAt,
			&r.Username,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan reaction: %w", err)
		}
		reactions = append(reactions, r)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating reactions: %w", err)
	}

	return reactions, nil
}

// UpdateSettings updates the privacy and password settings for a watch party
func (r *WatchPartyRepository) UpdateSettings(ctx context.Context, partyID uuid.UUID, privacy *string, password *string) error {
	// Build dynamic query based on provided fields
	query := `UPDATE watch_parties SET`
	params := []interface{}{}
	paramCount := 1
	updates := []string{}

	if privacy != nil {
		updates = append(updates, fmt.Sprintf(" visibility = $%d", paramCount))
		params = append(params, *privacy)
		paramCount++
	}

	if password != nil {
		// Treat empty password string or nil as a request to clear the password (set to NULL)
		if *password == "" {
			updates = append(updates, " password = NULL")
		} else {
			updates = append(updates, fmt.Sprintf(" password = $%d", paramCount))
			params = append(params, *password)
			paramCount++
		}
	}

	if len(updates) == 0 {
		return nil // Nothing to update
	}

	query += strings.Join(updates, ",")
	query += fmt.Sprintf(" WHERE id = $%d AND ended_at IS NULL", paramCount)
	params = append(params, partyID)

	result, err := r.pool.Exec(ctx, query, params...)
	if err != nil {
		return fmt.Errorf("failed to update watch party settings: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("watch party not found or already ended")
	}

	return nil
}

// GetHistory retrieves past watch parties for a user
func (r *WatchPartyRepository) GetHistory(ctx context.Context, userID uuid.UUID, limit, offset int) ([]models.WatchPartyHistoryEntry, int, error) {
	// First get the total count
	countQuery := `
		SELECT COUNT(DISTINCT wp.id)
		FROM watch_parties wp
		JOIN watch_party_participants wpp ON wp.id = wpp.party_id
		WHERE wpp.user_id = $1 AND wp.ended_at IS NOT NULL
	`

	var totalCount int
	err := r.pool.QueryRow(ctx, countQuery, userID).Scan(&totalCount)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count watch party history: %w", err)
	}

	// Then get the paginated history
	query := `
		SELECT 
			wp.id,
			wp.host_user_id,
			wp.title,
			wp.visibility,
			(SELECT COUNT(*) FROM watch_party_participants WHERE party_id = wp.id) as participant_count,
			wp.created_at,
			wp.started_at,
			wp.ended_at
		FROM watch_parties wp
		JOIN watch_party_participants wpp ON wp.id = wpp.party_id
		WHERE wpp.user_id = $1 AND wp.ended_at IS NOT NULL
		GROUP BY wp.id
		ORDER BY wp.ended_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.pool.Query(ctx, query, userID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get watch party history: %w", err)
	}
	defer rows.Close()

	var history []models.WatchPartyHistoryEntry
	for rows.Next() {
		var entry models.WatchPartyHistoryEntry
		err := rows.Scan(
			&entry.ID,
			&entry.HostUserID,
			&entry.Title,
			&entry.Visibility,
			&entry.ParticipantCount,
			&entry.CreatedAt,
			&entry.StartedAt,
			&entry.EndedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan history entry: %w", err)
		}

		// Calculate duration if both started_at and ended_at are present
		if entry.StartedAt != nil && entry.EndedAt != nil {
			duration := int(entry.EndedAt.Sub(*entry.StartedAt).Seconds())
			if duration < 0 {
				duration = 0
			}
			entry.Duration = &duration
		}

		history = append(history, entry)
	}

	if err = rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("error iterating history: %w", err)
	}

	return history, totalCount, nil
}

// GetPublicParties retrieves public/active watch parties for discovery
func (r *WatchPartyRepository) GetPublicParties(ctx context.Context, limit, offset int) ([]models.WatchParty, int, error) {
	// Get total count
	var totalCount int
	countQuery := `
		SELECT COUNT(*)
		FROM watch_parties
		WHERE visibility = 'public' AND ended_at IS NULL
	`
	err := r.pool.QueryRow(ctx, countQuery).Scan(&totalCount)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count public parties: %w", err)
	}

	// Get parties with participant counts
	query := `
		SELECT 
			wp.id, wp.host_user_id, wp.title, wp.playlist_id, wp.current_clip_id,
			wp.current_position_seconds, wp.is_playing, wp.visibility, wp.invite_code,
			wp.max_participants, wp.created_at, wp.started_at, wp.ended_at,
			COUNT(wpp.id) FILTER (WHERE wpp.left_at IS NULL) as active_participants
		FROM watch_parties wp
		LEFT JOIN watch_party_participants wpp ON wp.id = wpp.party_id
		WHERE wp.visibility = 'public' AND wp.ended_at IS NULL
		GROUP BY wp.id
		ORDER BY active_participants DESC, wp.created_at DESC
		LIMIT $1 OFFSET $2
	`

	rows, err := r.pool.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query public parties: %w", err)
	}
	defer rows.Close()

	var parties []models.WatchParty
	for rows.Next() {
		var party models.WatchParty
		var activeParticipants int
		err := rows.Scan(
			&party.ID,
			&party.HostUserID,
			&party.Title,
			&party.PlaylistID,
			&party.CurrentClipID,
			&party.CurrentPositionSeconds,
			&party.IsPlaying,
			&party.Visibility,
			&party.InviteCode,
			&party.MaxParticipants,
			&party.CreatedAt,
			&party.StartedAt,
			&party.EndedAt,
			&activeParticipants,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan public party: %w", err)
		}
		party.ActiveParticipantCount = activeParticipants
		parties = append(parties, party)
	}

	if err = rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("error iterating public parties: %w", err)
	}

	return parties, totalCount, nil
}

// GetTrendingParties retrieves trending watch parties (by participant count)
func (r *WatchPartyRepository) GetTrendingParties(ctx context.Context, limit int) ([]models.WatchParty, error) {
	query := `
		SELECT 
			wp.id, wp.host_user_id, wp.title, wp.playlist_id, wp.current_clip_id,
			wp.current_position_seconds, wp.is_playing, wp.visibility, wp.invite_code,
			wp.max_participants, wp.created_at, wp.started_at, wp.ended_at,
			COUNT(wpp.id) FILTER (WHERE wpp.left_at IS NULL) as active_participants
		FROM watch_parties wp
		LEFT JOIN watch_party_participants wpp ON wp.id = wpp.party_id
		WHERE wp.visibility = 'public' AND wp.ended_at IS NULL
		GROUP BY wp.id
		HAVING COUNT(wpp.id) FILTER (WHERE wpp.left_at IS NULL) > 0
		ORDER BY active_participants DESC, wp.created_at DESC
		LIMIT $1
	`

	rows, err := r.pool.Query(ctx, query, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query trending parties: %w", err)
	}
	defer rows.Close()

	var parties []models.WatchParty
	for rows.Next() {
		var party models.WatchParty
		var activeParticipants int
		err := rows.Scan(
			&party.ID,
			&party.HostUserID,
			&party.Title,
			&party.PlaylistID,
			&party.CurrentClipID,
			&party.CurrentPositionSeconds,
			&party.IsPlaying,
			&party.Visibility,
			&party.InviteCode,
			&party.MaxParticipants,
			&party.CreatedAt,
			&party.StartedAt,
			&party.EndedAt,
			&activeParticipants,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan trending party: %w", err)
		}
		party.ActiveParticipantCount = activeParticipants
		parties = append(parties, party)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating trending parties: %w", err)
	}

	return parties, nil
}
