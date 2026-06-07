package services

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/microcosm-cc/bluemonday"
	"git.subcult.tv/subculture-collective/clpr/internal/models"
	"git.subcult.tv/subculture-collective/clpr/internal/repository"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer/html"
)

const (
	// MaxCommentLength is the maximum allowed comment length
	MaxCommentLength = 10000
	// MinCommentLength is the minimum allowed comment length
	MinCommentLength = 1
	// MaxNestingDepth is the maximum nesting depth for replies
	MaxNestingDepth = 10
	// EditWindowMinutes is how long a user can edit their comment
	EditWindowMinutes = 15
	// KarmaPerComment is karma awarded for posting a comment
	KarmaPerComment = 1
	// KarmaPerUpvote is karma awarded when comment is upvoted
	KarmaPerUpvote = 1
	// KarmaPerDownvote is karma removed when comment is downvoted
	KarmaPerDownvote = -1
)

// CommentService handles comment business logic
type CommentService struct {
	repo                *repository.CommentRepository
	clipRepo            *repository.ClipRepository
	userRepo            *repository.UserRepository
	markdown            goldmark.Markdown
	sanitizer           *bluemonday.Policy
	notificationService *NotificationService
	toxicityClassifier  *ToxicityClassifier
}

// NewCommentService creates a new CommentService
func NewCommentService(repo *repository.CommentRepository, clipRepo *repository.ClipRepository, userRepo *repository.UserRepository, notificationService *NotificationService, toxicityClassifier *ToxicityClassifier) *CommentService {
	// Configure markdown processor
	md := goldmark.New(
		goldmark.WithExtensions(
			extension.Strikethrough,
			extension.Linkify,
			extension.Table,
		),
		goldmark.WithParserOptions(
			parser.WithAutoHeadingID(),
		),
		goldmark.WithRendererOptions(
			html.WithXHTML(),
			html.WithUnsafe(), // We'll sanitize with bluemonday instead
		),
	)

	// Configure HTML sanitizer - allow safe markdown subset
	sanitizer := bluemonday.UGCPolicy()
	sanitizer.AllowElements("p", "br", "strong", "em", "del", "code", "pre", "blockquote",
		"ul", "ol", "li", "a", "h1", "h2", "h3", "h4", "h5", "h6", "table", "thead",
		"tbody", "tr", "th", "td")
	sanitizer.AllowAttrs("href").OnElements("a")
	sanitizer.AllowAttrs("class").OnElements("code")
	sanitizer.RequireNoFollowOnLinks(true)
	sanitizer.RequireNoReferrerOnLinks(true)
	sanitizer.AddTargetBlankToFullyQualifiedLinks(true)

	return &CommentService{
		repo:                repo,
		clipRepo:            clipRepo,
		userRepo:            userRepo,
		markdown:            md,
		sanitizer:           sanitizer,
		notificationService: notificationService,
		toxicityClassifier:  toxicityClassifier,
	}
}

// CommentTreeNode represents a comment with nested replies
type CommentTreeNode struct {
	repository.CommentWithAuthor
	RenderedContent string            `json:"rendered_content"`
	Replies         []CommentTreeNode `json:"replies,omitempty"`
}

// CreateCommentRequest represents a request to create a comment
type CreateCommentRequest struct {
	Content         string     `json:"content"`
	ParentCommentID *uuid.UUID `json:"parent_comment_id,omitempty"`
}

// ValidateCreateComment validates a comment creation request
func (s *CommentService) ValidateCreateComment(ctx context.Context, req *CreateCommentRequest, clipID uuid.UUID) error {
	// Validate content length
	content := strings.TrimSpace(req.Content)
	if len(content) < MinCommentLength {
		return fmt.Errorf("comment must be at least %d character(s)", MinCommentLength)
	}
	if len(content) > MaxCommentLength {
		return fmt.Errorf("comment must not exceed %d characters", MaxCommentLength)
	}

	// Check if clip exists
	_, err := s.clipRepo.GetByID(ctx, clipID)
	if err != nil {
		return fmt.Errorf("clip not found")
	}

	// If replying to a parent comment, validate it
	if req.ParentCommentID != nil {
		parent, err := s.repo.GetByID(ctx, *req.ParentCommentID, nil)
		if err != nil {
			return fmt.Errorf("parent comment not found")
		}

		// Verify parent belongs to the same clip
		if parent.ClipID != clipID {
			return fmt.Errorf("parent comment does not belong to this clip")
		}

		// Check nesting depth
		depth, err := s.repo.GetNestingDepth(ctx, *req.ParentCommentID)
		if err != nil {
			return fmt.Errorf("failed to check nesting depth: %w", err)
		}

		if depth >= MaxNestingDepth {
			return fmt.Errorf("maximum nesting depth of %d reached", MaxNestingDepth)
		}
	}

	return nil
}

// CreateComment creates a new comment
func (s *CommentService) CreateComment(ctx context.Context, req *CreateCommentRequest, clipID, userID uuid.UUID) (*repository.CommentWithAuthor, error) {
	// Check if user can comment (not suspended)
	canComment, err := s.userRepo.CanUserComment(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to check comment privileges: %w", err)
	}
	if !canComment {
		// Get suspension details for better error message
		suspendedUntil, err := s.userRepo.GetCommentSuspensionInfo(ctx, userID)
		if err == nil && suspendedUntil != nil {
			// Check if it's a permanent suspension (year 9999)
			if suspendedUntil.Year() >= 9999 {
				return nil, fmt.Errorf("your comment privileges have been permanently suspended. Please contact support if you believe this is a mistake")
			}
			return nil, fmt.Errorf("your comment privileges are suspended until %s. Please contact support if you believe this is a mistake", suspendedUntil.Format("2006-01-02 15:04 MST"))
		}
		return nil, fmt.Errorf("comment privileges are currently suspended. Please try again later or contact support if you believe this is a mistake")
	}

	// Validate request
	if err := s.ValidateCreateComment(ctx, req, clipID); err != nil {
		return nil, err
	}

	// Create comment
	comment := &models.Comment{
		ID:              uuid.New(),
		ClipID:          clipID,
		UserID:          userID,
		ParentCommentID: req.ParentCommentID,
		Content:         strings.TrimSpace(req.Content),
		VoteScore:       0,
		ReplyCount:      0,
		IsEdited:        false,
		IsRemoved:       false,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	// Save to database
	if err := s.repo.Create(ctx, comment); err != nil {
		return nil, fmt.Errorf("failed to create comment: %w", err)
	}

	// Auto-upvote: Create an upvote from the comment creator
	// This encourages engagement and shows creator approval
	// Note: We call the repository method directly instead of s.VoteOnComment() to avoid
	// giving the user karma for voting on their own comment. The VoteOnComment service
	// method awards karma to the comment author, which would be circular in this case.
	if err := s.repo.VoteOnComment(ctx, userID, comment.ID, 1); err != nil {
		// Log error but don't fail the comment creation
		fmt.Printf("Warning: failed to auto-upvote comment for user %s: %v\n", userID, err)
	}

	// Award karma to user
	if err := s.repo.UpdateUserKarma(ctx, userID, KarmaPerComment); err != nil {
		// Log error but don't fail the comment creation
		fmt.Printf("Warning: failed to update karma for user %s: %v\n", userID, err)
	}

	// Get clip for creator notification
	clip, err := s.clipRepo.GetByID(ctx, clipID)
	if err == nil && clip.CreatorID != nil && s.notificationService != nil {
		// Send notification to clip creator about the comment
		if err := s.notificationService.NotifyClipComment(ctx, clipID, userID, *clip.CreatorID); err != nil {
			// Log error but don't fail the comment creation
			fmt.Printf("Warning: failed to send clip comment notification: %v\n", err)
		}
	}

	// Send notification for reply if this is a reply to a parent comment
	if s.notificationService != nil && req.ParentCommentID != nil {
		if err := s.notificationService.NotifyCommentReply(ctx, clipID, *req.ParentCommentID, userID); err != nil {
			// Log error but don't fail the comment creation
			fmt.Printf("Warning: failed to send reply notification: %v\n", err)
		}
	}

	// Perform toxicity classification (async - don't block comment creation)
	if s.toxicityClassifier != nil {
		go func() {
			// Create a new context with timeout for async processing
			asyncCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()

			// Classify the comment content
			score, err := s.toxicityClassifier.ClassifyComment(asyncCtx, comment.Content)
			if err != nil {
				// Log error but don't fail - this is a non-critical enhancement
				fmt.Printf("Warning: failed to classify comment %s for toxicity: %v\n", comment.ID, err)
				return
			}

			// Record the prediction for metrics
			if err := s.toxicityClassifier.RecordPrediction(asyncCtx, comment.ID, score); err != nil {
				fmt.Printf("Warning: failed to record toxicity prediction for comment %s: %v\n", comment.ID, err)
				// Continue even if recording fails
			}

			// Note: The database trigger will automatically add high-confidence toxic comments
			// to the moderation queue when RecordPrediction inserts into toxicity_predictions.
			// No need to call AddToModerationQueue explicitly here.
		}()
	}

	// Fetch the complete comment with author info and vote status
	commentWithAuthor, err := s.repo.GetByID(ctx, comment.ID, &userID)
	if err != nil {
		// If we can't get the full comment after creation, this is a serious issue.
		// Return an error since the API contract expects a complete CommentWithAuthor
		// with author info and vote status, which we cannot provide.
		return nil, fmt.Errorf("failed to fetch created comment with author info: %w", err)
	}

	return commentWithAuthor, nil
}

// UpdateComment updates a comment's content
func (s *CommentService) UpdateComment(ctx context.Context, commentID, userID uuid.UUID, content string, isAdmin bool) error {
	// Validate content length
	content = strings.TrimSpace(content)
	if len(content) < MinCommentLength {
		return fmt.Errorf("comment must be at least %d character(s)", MinCommentLength)
	}
	if len(content) > MaxCommentLength {
		return fmt.Errorf("comment must not exceed %d characters", MaxCommentLength)
	}

	// Get the comment
	comment, err := s.repo.GetByID(ctx, commentID, nil)
	if err != nil {
		return fmt.Errorf("comment not found")
	}

	// Check if user can edit
	if !isAdmin && comment.UserID != userID {
		return fmt.Errorf("you can only edit your own comments")
	}

	// Check edit time window (unless admin)
	if !isAdmin {
		canEdit, err := s.repo.CanUserEdit(ctx, commentID, userID, EditWindowMinutes)
		if err != nil {
			return fmt.Errorf("failed to check edit permission: %w", err)
		}
		if !canEdit {
			return fmt.Errorf("edit window of %d minutes has expired", EditWindowMinutes)
		}
	}

	// Update the comment
	if err := s.repo.Update(ctx, commentID, content); err != nil {
		return fmt.Errorf("failed to update comment: %w", err)
	}

	return nil
}

// DeleteComment soft-deletes a comment
func (s *CommentService) DeleteComment(ctx context.Context, commentID, userID uuid.UUID, role string, reason *string) error {
	// Get the comment
	comment, err := s.repo.GetByID(ctx, commentID, nil)
	if err != nil {
		return fmt.Errorf("comment not found")
	}

	// Determine if this is a moderator/admin action
	isMod := role == "moderator" || role == "admin"
	isAuthor := comment.UserID == userID

	if !isMod && !isAuthor {
		return fmt.Errorf("you can only delete your own comments")
	}

	// Delete the comment
	if err := s.repo.Delete(ctx, commentID, isMod && !isAuthor, reason); err != nil {
		return fmt.Errorf("failed to delete comment: %w", err)
	}

	// Remove karma from user if author deleted their own comment
	if isAuthor {
		if err := s.repo.UpdateUserKarma(ctx, comment.UserID, -KarmaPerComment); err != nil {
			fmt.Printf("Warning: failed to update karma for user %s: %v\n", comment.UserID, err)
		}
	}

	return nil
}

// VoteOnComment creates or updates a vote on a comment
func (s *CommentService) VoteOnComment(ctx context.Context, commentID, userID uuid.UUID, voteType int16) error {
	// Validate vote type
	if voteType != 1 && voteType != -1 && voteType != 0 {
		return fmt.Errorf("invalid vote type: must be 1 (upvote), -1 (downvote), or 0 (remove vote)")
	}

	// Get the comment
	comment, err := s.repo.GetByID(ctx, commentID, nil)
	if err != nil {
		return fmt.Errorf("comment not found")
	}

	// Get previous vote
	prevVote, err := s.repo.GetUserVote(ctx, userID, commentID)
	if err != nil {
		return fmt.Errorf("failed to get previous vote: %w", err)
	}

	// Handle vote removal
	if voteType == 0 {
		if prevVote != nil {
			if err := s.repo.RemoveVote(ctx, userID, commentID); err != nil {
				return fmt.Errorf("failed to remove vote: %w", err)
			}

			// Update karma (reverse previous vote)
			var karmaDelta int
			if *prevVote == 1 {
				karmaDelta = -KarmaPerUpvote
			} else {
				karmaDelta = -KarmaPerDownvote
			}
			if err := s.repo.UpdateUserKarma(ctx, comment.UserID, karmaDelta); err != nil {
				fmt.Printf("Warning: failed to update karma for user %s: %v\n", comment.UserID, err)
			}
		}
		return nil
	}

	// Create or update vote
	if err := s.repo.VoteOnComment(ctx, userID, commentID, voteType); err != nil {
		return fmt.Errorf("failed to vote on comment: %w", err)
	}

	// Update karma
	var karmaDelta int
	if prevVote != nil {
		// Reverse previous vote and apply new vote
		if *prevVote == 1 {
			karmaDelta = -KarmaPerUpvote
		} else {
			karmaDelta = -KarmaPerDownvote
		}
	}

	if voteType == 1 {
		karmaDelta += KarmaPerUpvote
	} else {
		karmaDelta += KarmaPerDownvote
	}

	if karmaDelta != 0 {
		if err := s.repo.UpdateUserKarma(ctx, comment.UserID, karmaDelta); err != nil {
			fmt.Printf("Warning: failed to update karma for user %s: %v\n", comment.UserID, err)
		}
	}

	return nil
}

// ListComments retrieves comments for a clip with sorting
func (s *CommentService) ListComments(ctx context.Context, clipID uuid.UUID, sortBy string, limit, offset int, userID *uuid.UUID) ([]CommentTreeNode, error) {
	// Get top-level comments
	comments, err := s.repo.ListByClipID(ctx, clipID, sortBy, limit, offset, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to list comments: %w", err)
	}

	// Return empty slice if no comments (not nil)
	if len(comments) == 0 {
		return []CommentTreeNode{}, nil
	}

	// Build tree nodes with rendered content
	var nodes []CommentTreeNode
	for _, c := range comments {
		node := CommentTreeNode{
			CommentWithAuthor: c,
			RenderedContent:   s.RenderMarkdown(c.Content),
			Replies:           []CommentTreeNode{},
		}
		nodes = append(nodes, node)
	}

	return nodes, nil
}

// ListCommentsWithReplies retrieves comments for a clip with optional nested replies
func (s *CommentService) ListCommentsWithReplies(ctx context.Context, clipID uuid.UUID, sortBy string, limit, offset int, userID *uuid.UUID, includeReplies bool) ([]CommentTreeNode, error) {
	// Get top-level comments
	comments, err := s.repo.ListByClipID(ctx, clipID, sortBy, limit, offset, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to list comments: %w", err)
	}

	// Return empty slice if no comments (not nil)
	if len(comments) == 0 {
		return []CommentTreeNode{}, nil
	}

	// Build tree nodes with rendered content
	var nodes []CommentTreeNode
	for _, c := range comments {
		node := CommentTreeNode{
			CommentWithAuthor: c,
			RenderedContent:   s.RenderMarkdown(c.Content),
			Replies:           []CommentTreeNode{},
		}

		// Recursively load replies if requested
		if includeReplies {
			replies, err := s.buildReplyTree(ctx, c.ID, userID, 1)
			if err != nil {
				return nil, fmt.Errorf("failed to build reply tree: %w", err)
			}
			node.Replies = replies
		}

		nodes = append(nodes, node)
	}

	return nodes, nil
}

// buildReplyTree recursively builds a tree of replies up to MaxNestingDepth
func (s *CommentService) buildReplyTree(ctx context.Context, parentID uuid.UUID, userID *uuid.UUID, currentDepth int) ([]CommentTreeNode, error) {
	// Stop recursion if we've reached max depth
	if currentDepth >= MaxNestingDepth {
		return []CommentTreeNode{}, nil
	}

	// Get direct replies to this comment
	// Use a reasonable limit for nested replies to prevent performance issues
	// We fetch more replies than typical pagination to provide a better UX for nested threads
	const maxRepliesPerLevel = 50
	replies, err := s.repo.GetReplies(ctx, parentID, maxRepliesPerLevel, 0, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get replies: %w", err)
	}

	// Return empty slice if no replies
	if len(replies) == 0 {
		return []CommentTreeNode{}, nil
	}

	// Build tree nodes with rendered content and recursively load their replies
	var nodes []CommentTreeNode
	for _, r := range replies {
		node := CommentTreeNode{
			CommentWithAuthor: r,
			RenderedContent:   s.RenderMarkdown(r.Content),
			Replies:           []CommentTreeNode{},
		}

		// Recursively load nested replies
		nestedReplies, err := s.buildReplyTree(ctx, r.ID, userID, currentDepth+1)
		if err != nil {
			return nil, fmt.Errorf("failed to build nested reply tree: %w", err)
		}
		node.Replies = nestedReplies

		nodes = append(nodes, node)
	}

	return nodes, nil
}

// GetReplies retrieves replies to a comment
func (s *CommentService) GetReplies(ctx context.Context, parentID uuid.UUID, limit, offset int, userID *uuid.UUID) ([]CommentTreeNode, error) {
	replies, err := s.repo.GetReplies(ctx, parentID, limit, offset, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get replies: %w", err)
	}

	// Build tree nodes with rendered content
	var nodes []CommentTreeNode
	for _, r := range replies {
		node := CommentTreeNode{
			CommentWithAuthor: r,
			RenderedContent:   s.RenderMarkdown(r.Content),
			Replies:           []CommentTreeNode{},
		}
		nodes = append(nodes, node)
	}

	return nodes, nil
}

// RenderMarkdown processes and sanitizes markdown content
func (s *CommentService) RenderMarkdown(content string) string {
	// If content is removed/deleted, don't process markdown
	if content == "[deleted]" || content == "[removed]" {
		return content
	}

	var buf bytes.Buffer
	if err := s.markdown.Convert([]byte(content), &buf); err != nil {
		// On error, return plain text
		return content
	}

	// Sanitize HTML
	sanitized := s.sanitizer.Sanitize(buf.String())
	return sanitized
}
