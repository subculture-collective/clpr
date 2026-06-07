package repository

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"git.subcult.tv/subculture-collective/clpr/internal/models"
)

// VoteRepository handles database operations for votes
type VoteRepository struct {
	pool *pgxpool.Pool
}

// NewVoteRepository creates a new VoteRepository
func NewVoteRepository(pool *pgxpool.Pool) *VoteRepository {
	return &VoteRepository{
		pool: pool,
	}
}

// UpsertVote creates or updates a vote
func (r *VoteRepository) UpsertVote(ctx context.Context, userID, clipID uuid.UUID, voteType int16) error {
	query := `
		INSERT INTO votes (user_id, clip_id, vote_type)
		VALUES ($1, $2, $3)
		ON CONFLICT (user_id, clip_id)
		DO UPDATE SET vote_type = EXCLUDED.vote_type
	`

	_, err := r.pool.Exec(ctx, query, userID, clipID, voteType)
	if err != nil {
		return fmt.Errorf("failed to upsert vote: %w", err)
	}

	return nil
}

// DeleteVote removes a vote
func (r *VoteRepository) DeleteVote(ctx context.Context, userID, clipID uuid.UUID) error {
	query := `DELETE FROM votes WHERE user_id = $1 AND clip_id = $2`

	_, err := r.pool.Exec(ctx, query, userID, clipID)
	if err != nil {
		return fmt.Errorf("failed to delete vote: %w", err)
	}

	return nil
}

// GetVote retrieves a user's vote on a clip
func (r *VoteRepository) GetVote(ctx context.Context, userID, clipID uuid.UUID) (*models.Vote, error) {
	query := `
		SELECT id, user_id, clip_id, vote_type, created_at
		FROM votes
		WHERE user_id = $1 AND clip_id = $2
	`

	var vote models.Vote
	err := r.pool.QueryRow(ctx, query, userID, clipID).Scan(
		&vote.ID, &vote.UserID, &vote.ClipID, &vote.VoteType, &vote.CreatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to get vote: %w", err)
	}

	return &vote, nil
}

// GetVoteCounts returns the upvote and downvote counts for a clip
func (r *VoteRepository) GetVoteCounts(ctx context.Context, clipID uuid.UUID) (upvotes, downvotes int, err error) {
	query := `
		SELECT
			COUNT(CASE WHEN vote_type = 1 THEN 1 END) as upvotes,
			COUNT(CASE WHEN vote_type = -1 THEN 1 END) as downvotes
		FROM votes
		WHERE clip_id = $1
	`

	err = r.pool.QueryRow(ctx, query, clipID).Scan(&upvotes, &downvotes)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to get vote counts: %w", err)
	}

	return upvotes, downvotes, nil
}

// GetUserVotedClips returns clips that a user has voted on (upvoted or downvoted)
func (r *VoteRepository) GetUserVotedClips(ctx context.Context, userID uuid.UUID, voteType int16, limit, offset int) ([]uuid.UUID, int, error) {
	// Get total count
	countQuery := `
		SELECT COUNT(*)
		FROM votes
		WHERE user_id = $1 AND vote_type = $2
	`
	var total int
	if err := r.pool.QueryRow(ctx, countQuery, userID, voteType).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("failed to count voted clips: %w", err)
	}

	// Get clip IDs
	query := `
		SELECT clip_id
		FROM votes
		WHERE user_id = $1 AND vote_type = $2
		ORDER BY created_at DESC
		LIMIT $3 OFFSET $4
	`

	rows, err := r.pool.Query(ctx, query, userID, voteType, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query voted clips: %w", err)
	}
	defer rows.Close()

	var clipIDs []uuid.UUID
	for rows.Next() {
		var clipID uuid.UUID
		if err := rows.Scan(&clipID); err != nil {
			return nil, 0, fmt.Errorf("failed to scan clip ID: %w", err)
		}
		clipIDs = append(clipIDs, clipID)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("error iterating voted clips: %w", err)
	}

	return clipIDs, total, nil
}
