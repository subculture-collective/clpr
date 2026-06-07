package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"git.subcult.tv/subculture-collective/clpr/internal/models"
)

// CommentRepository handles database operations for comments
type CommentRepository struct {
	pool *pgxpool.Pool
}

// NewCommentRepository creates a new CommentRepository
func NewCommentRepository(pool *pgxpool.Pool) *CommentRepository {
	return &CommentRepository{
		pool: pool,
	}
}

// CommentWithAuthor represents a comment with author information
type CommentWithAuthor struct {
	models.Comment
	AuthorUsername    string  `json:"author_username" db:"author_username"`
	AuthorDisplayName string  `json:"author_display_name" db:"author_display_name"`
	AuthorAvatarURL   *string `json:"author_avatar_url" db:"author_avatar_url"`
	AuthorKarma       int     `json:"author_karma" db:"author_karma"`
	AuthorRole        string  `json:"author_role" db:"author_role"`
	ReplyCount        int     `json:"reply_count" db:"reply_count"`
	UserVote          *int16  `json:"user_vote,omitempty" db:"user_vote"`
}

// ListByClipID retrieves comments for a clip with sorting and pagination
func (r *CommentRepository) ListByClipID(ctx context.Context, clipID uuid.UUID, sortBy string, limit, offset int, userID *uuid.UUID) ([]CommentWithAuthor, error) {
	var orderClause string

	switch sortBy {
	case "new":
		orderClause = "created_at DESC"
	case "old":
		orderClause = "created_at ASC"
	case "controversial":
		// Controversial: high vote count but score near zero
		orderClause = `
			(ABS(vote_score) / NULLIF(GREATEST(ABS(vote_score), 1), 0)) DESC,
			ABS(vote_score) DESC
		`
	case "best":
		fallthrough
	default:
		// Wilson score confidence interval for "best" sorting
		// Uses vote counts from the CTE to avoid N+1 query pattern
		orderClause = `
			CASE
				WHEN total_votes = 0 THEN 0
				ELSE (
					((upvotes + 1.9208) / total_votes -
					1.96 * SQRT((upvotes * downvotes) / total_votes + 0.9604) / total_votes)
				) / (1 + 3.8416 / total_votes)
			END DESC,
			vote_score DESC,
			created_at DESC
		`
	}

	query := fmt.Sprintf(`
		WITH vote_counts AS (
			-- Pre-calculate vote counts to avoid N+1 query pattern
			SELECT
				comment_id,
				COUNT(*) AS total_votes,
				COUNT(*) FILTER (WHERE vote_type = 1) AS upvotes,
				COUNT(*) FILTER (WHERE vote_type = -1) AS downvotes
			FROM comment_votes
			WHERE comment_id IN (
				SELECT id FROM comments WHERE clip_id = $1 AND parent_comment_id IS NULL
			)
			GROUP BY comment_id
		),
		comment_tree AS (
			-- Get top-level comments for this clip
			SELECT
				c.id, c.clip_id, c.user_id, c.parent_comment_id, c.content,
				c.vote_score, c.reply_count, c.is_edited, c.is_removed, c.removed_reason,
				c.created_at, c.updated_at,
				u.username AS author_username,
				u.display_name AS author_display_name,
				u.avatar_url AS author_avatar_url,
				u.karma_points AS author_karma,
				u.role AS author_role,
				COALESCE(cv.vote_type, NULL) AS user_vote,
				0 AS depth,
				COALESCE(vc.total_votes, 0) AS total_votes,
				COALESCE(vc.upvotes, 0) AS upvotes,
				COALESCE(vc.downvotes, 0) AS downvotes
			FROM comments c
			INNER JOIN users u ON c.user_id = u.id
			LEFT JOIN comment_votes cv ON c.id = cv.comment_id AND cv.user_id = $2
			LEFT JOIN vote_counts vc ON c.id = vc.comment_id
			WHERE c.clip_id = $1 AND c.parent_comment_id IS NULL
		)
		SELECT * FROM comment_tree
		ORDER BY %s
		LIMIT $3 OFFSET $4
	`, orderClause)

	var args []interface{}
	if userID != nil {
		args = []interface{}{clipID, *userID, limit, offset}
	} else {
		// Use a nil UUID for non-authenticated requests
		args = []interface{}{clipID, uuid.Nil, limit, offset}
	}

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list comments: %w", err)
	}
	defer rows.Close()

	var comments []CommentWithAuthor
	for rows.Next() {
		var c CommentWithAuthor
		var depth int
		var totalVotes, upvotes, downvotes int // Vote count columns from CTE
		err := rows.Scan(
			&c.ID, &c.ClipID, &c.UserID, &c.ParentCommentID, &c.Content,
			&c.VoteScore, &c.ReplyCount, &c.IsEdited, &c.IsRemoved, &c.RemovedReason,
			&c.CreatedAt, &c.UpdatedAt,
			&c.AuthorUsername, &c.AuthorDisplayName, &c.AuthorAvatarURL,
			&c.AuthorKarma, &c.AuthorRole, &c.UserVote,
			&depth, &totalVotes, &upvotes, &downvotes,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan comment: %w", err)
		}
		comments = append(comments, c)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating comments: %w", err)
	}

	return comments, nil
}

// GetReplies retrieves replies to a comment
func (r *CommentRepository) GetReplies(ctx context.Context, parentID uuid.UUID, limit, offset int, userID *uuid.UUID) ([]CommentWithAuthor, error) {
	query := `
		SELECT
			c.id, c.clip_id, c.user_id, c.parent_comment_id, c.content,
			c.vote_score, c.reply_count, c.is_edited, c.is_removed, c.removed_reason,
			c.created_at, c.updated_at,
			u.username AS author_username,
			u.display_name AS author_display_name,
			u.avatar_url AS author_avatar_url,
			u.karma_points AS author_karma,
			u.role AS author_role,
			COALESCE(cv.vote_type, NULL) AS user_vote
		FROM comments c
		INNER JOIN users u ON c.user_id = u.id
		LEFT JOIN comment_votes cv ON c.id = cv.comment_id AND cv.user_id = $2
		WHERE c.parent_comment_id = $1
		ORDER BY c.vote_score DESC, c.created_at DESC
		LIMIT $3 OFFSET $4
	`

	var args []interface{}
	if userID != nil {
		args = []interface{}{parentID, *userID, limit, offset}
	} else {
		args = []interface{}{parentID, uuid.Nil, limit, offset}
	}

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get replies: %w", err)
	}
	defer rows.Close()

	var comments []CommentWithAuthor
	for rows.Next() {
		var c CommentWithAuthor
		err := rows.Scan(
			&c.ID, &c.ClipID, &c.UserID, &c.ParentCommentID, &c.Content,
			&c.VoteScore, &c.ReplyCount, &c.IsEdited, &c.IsRemoved, &c.RemovedReason,
			&c.CreatedAt, &c.UpdatedAt,
			&c.AuthorUsername, &c.AuthorDisplayName, &c.AuthorAvatarURL,
			&c.AuthorKarma, &c.AuthorRole, &c.UserVote,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan reply: %w", err)
		}
		comments = append(comments, c)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating replies: %w", err)
	}

	return comments, nil
}

// GetByID retrieves a comment by ID with author info
func (r *CommentRepository) GetByID(ctx context.Context, id uuid.UUID, userID *uuid.UUID) (*CommentWithAuthor, error) {
	query := `
		SELECT
			c.id, c.clip_id, c.user_id, c.parent_comment_id, c.content,
			c.vote_score, c.reply_count, c.is_edited, c.is_removed, c.removed_reason,
			c.created_at, c.updated_at,
			u.username AS author_username,
			u.display_name AS author_display_name,
			u.avatar_url AS author_avatar_url,
			u.karma_points AS author_karma,
			u.role AS author_role,
			COALESCE(cv.vote_type, NULL) AS user_vote
		FROM comments c
		INNER JOIN users u ON c.user_id = u.id
		LEFT JOIN comment_votes cv ON c.id = cv.comment_id AND cv.user_id = $2
		WHERE c.id = $1
	`

	var c CommentWithAuthor
	var args []interface{}
	if userID != nil {
		args = []interface{}{id, *userID}
	} else {
		args = []interface{}{id, uuid.Nil}
	}

	err := r.pool.QueryRow(ctx, query, args...).Scan(
		&c.ID, &c.ClipID, &c.UserID, &c.ParentCommentID, &c.Content,
		&c.VoteScore, &c.ReplyCount, &c.IsEdited, &c.IsRemoved, &c.RemovedReason,
		&c.CreatedAt, &c.UpdatedAt,
		&c.AuthorUsername, &c.AuthorDisplayName, &c.AuthorAvatarURL,
		&c.AuthorKarma, &c.AuthorRole, &c.UserVote,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("comment not found")
		}
		return nil, fmt.Errorf("failed to get comment: %w", err)
	}

	return &c, nil
}

// Create inserts a new comment
func (r *CommentRepository) Create(ctx context.Context, comment *models.Comment) error {
	query := `
		INSERT INTO comments (
			id, clip_id, user_id, parent_comment_id, content,
			vote_score, reply_count, is_edited, is_removed, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11
		)
	`

	_, err := r.pool.Exec(ctx, query,
		comment.ID, comment.ClipID, comment.UserID, comment.ParentCommentID,
		comment.Content, comment.VoteScore, comment.ReplyCount, comment.IsEdited, comment.IsRemoved,
		comment.CreatedAt, comment.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create comment: %w", err)
	}

	return nil
}

// Update updates a comment's content
// Note: updated_at is automatically updated by the update_comments_updated_at database trigger
func (r *CommentRepository) Update(ctx context.Context, id uuid.UUID, content string) error {
	query := `
		UPDATE comments
		SET content = $2, is_edited = true
		WHERE id = $1
	`

	result, err := r.pool.Exec(ctx, query, id, content)
	if err != nil {
		return fmt.Errorf("failed to update comment: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("comment not found")
	}

	return nil
}

// Delete soft-deletes a comment
func (r *CommentRepository) Delete(ctx context.Context, id uuid.UUID, isModAction bool, reason *string) error {
	var content string
	if isModAction {
		content = "[removed]"
	} else {
		content = "[deleted]"
	}

	query := `
		UPDATE comments
		SET content = $2, is_removed = true, removed_reason = $3
		WHERE id = $1
	`

	result, err := r.pool.Exec(ctx, query, id, content, reason)
	if err != nil {
		return fmt.Errorf("failed to delete comment: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("comment not found")
	}

	return nil
}

// GetNestingDepth calculates the nesting depth of a comment
func (r *CommentRepository) GetNestingDepth(ctx context.Context, parentID uuid.UUID) (int, error) {
	query := `
		WITH RECURSIVE parent_chain AS (
			SELECT id, parent_comment_id, 1 AS depth
			FROM comments
			WHERE id = $1

			UNION ALL

			SELECT c.id, c.parent_comment_id, pc.depth + 1
			FROM comments c
			INNER JOIN parent_chain pc ON c.id = pc.parent_comment_id
		)
		SELECT COALESCE(MAX(depth), 0) FROM parent_chain
	`

	var depth int
	err := r.pool.QueryRow(ctx, query, parentID).Scan(&depth)
	if err != nil {
		return 0, fmt.Errorf("failed to get nesting depth: %w", err)
	}

	return depth, nil
}

// VoteOnComment creates or updates a vote on a comment
func (r *CommentRepository) VoteOnComment(ctx context.Context, userID, commentID uuid.UUID, voteType int16) error {
	query := `
		INSERT INTO comment_votes (id, user_id, comment_id, vote_type, created_at)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (user_id, comment_id)
		DO UPDATE SET vote_type = EXCLUDED.vote_type
	`

	_, err := r.pool.Exec(ctx, query,
		uuid.New(), userID, commentID, voteType, time.Now(),
	)

	if err != nil {
		return fmt.Errorf("failed to vote on comment: %w", err)
	}

	return nil
}

// RemoveVote removes a vote from a comment
func (r *CommentRepository) RemoveVote(ctx context.Context, userID, commentID uuid.UUID) error {
	query := `DELETE FROM comment_votes WHERE user_id = $1 AND comment_id = $2`

	_, err := r.pool.Exec(ctx, query, userID, commentID)
	if err != nil {
		return fmt.Errorf("failed to remove vote: %w", err)
	}

	return nil
}

// GetUserVote gets a user's vote on a comment
func (r *CommentRepository) GetUserVote(ctx context.Context, userID, commentID uuid.UUID) (*int16, error) {
	query := `SELECT vote_type FROM comment_votes WHERE user_id = $1 AND comment_id = $2`

	var voteType int16
	err := r.pool.QueryRow(ctx, query, userID, commentID).Scan(&voteType)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get user vote: %w", err)
	}

	return &voteType, nil
}

// CanUserEdit checks if a user can edit a comment (within time limit)
func (r *CommentRepository) CanUserEdit(ctx context.Context, commentID, userID uuid.UUID, editWindowMinutes int) (bool, error) {
	query := `
		SELECT
			user_id = $2 AND
			created_at > NOW() - INTERVAL '1 minute' * $3
		FROM comments
		WHERE id = $1
	`

	var canEdit bool
	err := r.pool.QueryRow(ctx, query, commentID, userID, editWindowMinutes).Scan(&canEdit)
	if err != nil {
		if err == pgx.ErrNoRows {
			return false, nil
		}
		return false, fmt.Errorf("failed to check edit permission: %w", err)
	}

	return canEdit, nil
}

// UpdateUserKarma updates a user's karma points
func (r *CommentRepository) UpdateUserKarma(ctx context.Context, userID uuid.UUID, delta int) error {
	query := `UPDATE users SET karma_points = karma_points + $2 WHERE id = $1`

	_, err := r.pool.Exec(ctx, query, userID, delta)
	if err != nil {
		return fmt.Errorf("failed to update user karma: %w", err)
	}

	return nil
}

// RemoveComment marks a comment as removed with a reason
func (r *CommentRepository) RemoveComment(ctx context.Context, commentID uuid.UUID, reason *string) error {
	query := `
		UPDATE comments
		SET is_removed = true, removed_reason = $2
		WHERE id = $1
	`

	_, err := r.pool.Exec(ctx, query, commentID, reason)
	return err
}

// GetCommentTree retrieves a full nested comment tree starting from a parent comment
// This method uses a recursive CTE to efficiently fetch the entire subtree in a single query
func (r *CommentRepository) GetCommentTree(ctx context.Context, parentID uuid.UUID, userID *uuid.UUID) ([]CommentWithAuthor, error) {
	query := `
		WITH RECURSIVE comment_tree AS (
			-- Base case: get the parent comment
			SELECT
				c.id, c.clip_id, c.user_id, c.parent_comment_id, c.content,
				c.vote_score, c.reply_count, c.is_edited, c.is_removed, c.removed_reason,
				c.created_at, c.updated_at,
				u.username AS author_username,
				u.display_name AS author_display_name,
				u.avatar_url AS author_avatar_url,
				u.karma_points AS author_karma,
				u.role AS author_role,
				COALESCE(cv.vote_type, NULL) AS user_vote,
				0 AS depth,
				ARRAY[c.created_at] AS path
			FROM comments c
			INNER JOIN users u ON c.user_id = u.id
			LEFT JOIN comment_votes cv ON c.id = cv.comment_id AND cv.user_id = $2
			WHERE c.id = $1

			UNION ALL

			-- Recursive case: get all replies
			SELECT
				c.id, c.clip_id, c.user_id, c.parent_comment_id, c.content,
				c.vote_score, c.reply_count, c.is_edited, c.is_removed, c.removed_reason,
				c.created_at, c.updated_at,
				u.username AS author_username,
				u.display_name AS author_display_name,
				u.avatar_url AS author_avatar_url,
				u.karma_points AS author_karma,
				u.role AS author_role,
				COALESCE(cv.vote_type, NULL) AS user_vote,
				ct.depth + 1,
				ct.path || c.created_at
			FROM comments c
			INNER JOIN users u ON c.user_id = u.id
			INNER JOIN comment_tree ct ON c.parent_comment_id = ct.id
			LEFT JOIN comment_votes cv ON c.id = cv.comment_id AND cv.user_id = $2
			WHERE ct.depth < 10 -- Prevent infinite recursion
		)
		SELECT
			id, clip_id, user_id, parent_comment_id, content,
			vote_score, reply_count, is_edited, is_removed, removed_reason,
			created_at, updated_at,
			author_username, author_display_name, author_avatar_url,
			author_karma, author_role, user_vote
		FROM comment_tree
		ORDER BY path
	`

	var args []interface{}
	if userID != nil {
		args = []interface{}{parentID, *userID}
	} else {
		args = []interface{}{parentID, uuid.Nil}
	}

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get comment tree: %w", err)
	}
	defer rows.Close()

	var comments []CommentWithAuthor
	for rows.Next() {
		var c CommentWithAuthor
		err := rows.Scan(
			&c.ID, &c.ClipID, &c.UserID, &c.ParentCommentID, &c.Content,
			&c.VoteScore, &c.ReplyCount, &c.IsEdited, &c.IsRemoved, &c.RemovedReason,
			&c.CreatedAt, &c.UpdatedAt,
			&c.AuthorUsername, &c.AuthorDisplayName, &c.AuthorAvatarURL,
			&c.AuthorKarma, &c.AuthorRole, &c.UserVote,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan comment in tree: %w", err)
		}
		comments = append(comments, c)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating comment tree: %w", err)
	}

	return comments, nil
}

// GetTopLevelComments retrieves top-level comments for a clip with pagination
// This is optimized for the initial page load of a comment section
func (r *CommentRepository) GetTopLevelComments(ctx context.Context, clipID uuid.UUID, limit, offset int, userID *uuid.UUID) ([]CommentWithAuthor, error) {
	query := `
		SELECT
			c.id, c.clip_id, c.user_id, c.parent_comment_id, c.content,
			c.vote_score, c.reply_count, c.is_edited, c.is_removed, c.removed_reason,
			c.created_at, c.updated_at,
			u.username AS author_username,
			u.display_name AS author_display_name,
			u.avatar_url AS author_avatar_url,
			u.karma_points AS author_karma,
			u.role AS author_role,
			COALESCE(cv.vote_type, NULL) AS user_vote
		FROM comments c
		INNER JOIN users u ON c.user_id = u.id
		LEFT JOIN comment_votes cv ON c.id = cv.comment_id AND cv.user_id = $2
		WHERE c.clip_id = $1 AND c.parent_comment_id IS NULL
		ORDER BY c.vote_score DESC, c.created_at DESC
		LIMIT $3 OFFSET $4
	`

	var args []interface{}
	if userID != nil {
		args = []interface{}{clipID, *userID, limit, offset}
	} else {
		args = []interface{}{clipID, uuid.Nil, limit, offset}
	}

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get top-level comments: %w", err)
	}
	defer rows.Close()

	var comments []CommentWithAuthor
	for rows.Next() {
		var c CommentWithAuthor
		err := rows.Scan(
			&c.ID, &c.ClipID, &c.UserID, &c.ParentCommentID, &c.Content,
			&c.VoteScore, &c.ReplyCount, &c.IsEdited, &c.IsRemoved, &c.RemovedReason,
			&c.CreatedAt, &c.UpdatedAt,
			&c.AuthorUsername, &c.AuthorDisplayName, &c.AuthorAvatarURL,
			&c.AuthorKarma, &c.AuthorRole, &c.UserVote,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan top-level comment: %w", err)
		}
		comments = append(comments, c)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating top-level comments: %w", err)
	}

	return comments, nil
}

// ListByUserID retrieves comments by a user with pagination
func (r *CommentRepository) ListByUserID(ctx context.Context, userID uuid.UUID, limit, offset int) ([]CommentWithAuthor, int, error) {
	// Get total count
	countQuery := `SELECT COUNT(*) FROM comments WHERE user_id = $1`
	var total int
	if err := r.pool.QueryRow(ctx, countQuery, userID).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("failed to count user comments: %w", err)
	}

	// Get comments
	query := `
		SELECT
			c.id, c.clip_id, c.user_id, c.parent_comment_id, c.content,
			c.vote_score, c.reply_count, c.is_edited, c.is_removed, c.removed_reason,
			c.created_at, c.updated_at,
			u.username AS author_username,
			u.display_name AS author_display_name,
			u.avatar_url AS author_avatar_url,
			u.karma_points AS author_karma,
			u.role AS author_role,
			NULL AS user_vote
		FROM comments c
		INNER JOIN users u ON c.user_id = u.id
		WHERE c.user_id = $1
		ORDER BY c.created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.pool.Query(ctx, query, userID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query user comments: %w", err)
	}
	defer rows.Close()

	var comments []CommentWithAuthor
	for rows.Next() {
		var c CommentWithAuthor
		if err := rows.Scan(
			&c.ID, &c.ClipID, &c.UserID, &c.ParentCommentID, &c.Content,
			&c.VoteScore, &c.ReplyCount, &c.IsEdited, &c.IsRemoved, &c.RemovedReason,
			&c.CreatedAt, &c.UpdatedAt,
			&c.AuthorUsername, &c.AuthorDisplayName, &c.AuthorAvatarURL,
			&c.AuthorKarma, &c.AuthorRole, &c.UserVote,
		); err != nil {
			return nil, 0, fmt.Errorf("failed to scan comment: %w", err)
		}
		comments = append(comments, c)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("error iterating user comments: %w", err)
	}

	return comments, total, nil
}
