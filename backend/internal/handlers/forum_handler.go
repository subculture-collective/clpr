package handlers

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// ForumHandler handles forum thread and reply operations
type ForumHandler struct {
	db *pgxpool.Pool
}

// NewForumHandler creates a new ForumHandler
func NewForumHandler(db *pgxpool.Pool) *ForumHandler {
	return &ForumHandler{
		db: db,
	}
}

// ForumThread represents a forum discussion thread
type ForumThread struct {
	ID         uuid.UUID  `json:"id"`
	UserID     uuid.UUID  `json:"user_id"`
	Username   string     `json:"username"`
	Title      string     `json:"title"`
	Content    string     `json:"content"`
	GameID     *uuid.UUID `json:"game_id,omitempty"`
	GameName   *string    `json:"game_name,omitempty"`
	Tags       []string   `json:"tags"`
	ViewCount  int        `json:"view_count"`
	ReplyCount int        `json:"reply_count"`
	Locked     bool       `json:"locked"`
	LockedAt   *time.Time `json:"locked_at,omitempty"`
	Pinned     bool       `json:"pinned"`
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`
}

// ForumReply represents a reply in a forum thread
type ForumReply struct {
	ID            uuid.UUID    `json:"id"`
	UserID        uuid.UUID    `json:"user_id"`
	Username      string       `json:"username"`
	ThreadID      uuid.UUID    `json:"thread_id"`
	ParentReplyID *uuid.UUID   `json:"parent_reply_id,omitempty"`
	Content       string       `json:"content"`
	Depth         int          `json:"depth"`
	Path          string       `json:"path"`
	CreatedAt     time.Time    `json:"created_at"`
	UpdatedAt     time.Time    `json:"updated_at"`
	Replies       []ForumReply `json:"replies,omitempty"`
}

// Vote represents a vote on a reply
type Vote struct {
	ID        uuid.UUID `json:"id"`
	UserID    uuid.UUID `json:"user_id"`
	ReplyID   uuid.UUID `json:"reply_id"`
	VoteValue int       `json:"vote_value"` // -1, 0, 1
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// ReputationScore represents a user's reputation
type ReputationScore struct {
	UserID    uuid.UUID `json:"user_id"`
	Score     int       `json:"score"`
	Badge     string    `json:"badge"` // new, contributor, expert, moderator
	Votes     int       `json:"votes"`
	Threads   int       `json:"threads"`
	Replies   int       `json:"replies"`
	UpdatedAt time.Time `json:"updated_at"`
}

// VoteStats represents vote statistics for a reply
type VoteStats struct {
	Upvotes   int `json:"upvotes"`
	Downvotes int `json:"downvotes"`
	NetVotes  int `json:"net_votes"`
	UserVote  int `json:"user_vote"` // -1, 0, 1 (0 means no vote)
}

// CreateThreadRequest represents the request to create a thread
type CreateThreadRequest struct {
	Title   string   `json:"title" binding:"required,min=3,max=200"`
	Content string   `json:"content" binding:"required,min=10,max=5000"`
	GameID  *string  `json:"game_id"`
	Tags    []string `json:"tags" binding:"max=5"`
}

// CreateThread creates a new forum thread
// POST /api/v1/forum/threads
func (h *ForumHandler) CreateThread(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "User not authenticated",
		})
		return
	}

	var req CreateThreadRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request: " + err.Error(),
		})
		return
	}

	// Validate tags
	for _, tag := range req.Tags {
		if len(tag) > 50 {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Tag cannot exceed 50 characters",
			})
			return
		}
	}

	// Parse game_id if provided
	var gameID *uuid.UUID
	if req.GameID != nil && *req.GameID != "" {
		parsedGameID, err := uuid.Parse(*req.GameID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid game_id format",
			})
			return
		}
		gameID = &parsedGameID
	}

	// Insert thread
	threadID := uuid.New()
	query := `
		INSERT INTO forum_threads (id, user_id, title, content, game_id, tags)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING created_at
	`

	var createdAt time.Time
	err := h.db.QueryRow(c.Request.Context(), query,
		threadID, userID, req.Title, req.Content, gameID, req.Tags,
	).Scan(&createdAt)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to create thread",
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"data": gin.H{
			"id":         threadID,
			"created_at": createdAt,
		},
	})
}

// ListThreads retrieves a list of forum threads with filters
// GET /api/v1/forum/threads?page=1&sort=recent&game_filter=<game_id>&search=<query>
func (h *ForumHandler) ListThreads(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	if page < 1 {
		page = 1
	}
	limit := 20
	offset := (page - 1) * limit

	sort := c.DefaultQuery("sort", "recent") // recent, popular, replies
	gameFilter := c.Query("game_filter")
	searchQuery := c.Query("search")

	// Build query
	queryBuilder := strings.Builder{}
	queryBuilder.WriteString(`
		SELECT 
			ft.id, ft.user_id, u.username, ft.title, ft.content,
			ft.game_id, g.name as game_name, ft.tags, ft.view_count, ft.reply_count,
			ft.locked, ft.locked_at, ft.pinned, ft.created_at, ft.updated_at
		FROM forum_threads ft
		JOIN users u ON ft.user_id = u.id
		LEFT JOIN games g ON ft.game_id = g.id
		WHERE ft.is_deleted = FALSE
	`)

	args := []interface{}{}
	argCount := 1

	// Filter by game
	if gameFilter != "" {
		gameID, err := uuid.Parse(gameFilter)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid game_filter format",
			})
			return
		}
		queryBuilder.WriteString(fmt.Sprintf(" AND ft.game_id = $%d", argCount))
		args = append(args, gameID)
		argCount++
	}

	// Full-text search
	if searchQuery != "" {
		queryBuilder.WriteString(fmt.Sprintf(" AND ft.search_vector @@ plainto_tsquery('english', $%d)", argCount))
		args = append(args, searchQuery)
		argCount++
	}

	// Sorting
	switch sort {
	case "popular":
		queryBuilder.WriteString(" ORDER BY ft.pinned DESC, ft.view_count DESC, ft.created_at DESC")
	case "replies":
		queryBuilder.WriteString(" ORDER BY ft.pinned DESC, ft.reply_count DESC, ft.created_at DESC")
	default: // recent
		queryBuilder.WriteString(" ORDER BY ft.pinned DESC, ft.updated_at DESC")
	}

	queryBuilder.WriteString(fmt.Sprintf(" LIMIT $%d OFFSET $%d", argCount, argCount+1))
	args = append(args, limit, offset)

	rows, err := h.db.Query(c.Request.Context(), queryBuilder.String(), args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to retrieve threads",
		})
		return
	}
	defer rows.Close()

	threads := []ForumThread{}
	for rows.Next() {
		var thread ForumThread
		var tags []string

		err := rows.Scan(
			&thread.ID, &thread.UserID, &thread.Username, &thread.Title, &thread.Content,
			&thread.GameID, &thread.GameName, &tags, &thread.ViewCount, &thread.ReplyCount,
			&thread.Locked, &thread.LockedAt, &thread.Pinned, &thread.CreatedAt, &thread.UpdatedAt,
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to scan thread",
			})
			return
		}

		thread.Tags = []string(tags)
		if thread.Tags == nil {
			thread.Tags = []string{}
		}
		threads = append(threads, thread)
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    threads,
		"meta": gin.H{
			"page":  page,
			"limit": limit,
			"count": len(threads),
		},
	})
}

// GetThread retrieves a specific thread with its full reply tree
// GET /api/v1/forum/threads/:id
func (h *ForumHandler) GetThread(c *gin.Context) {
	threadID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid thread ID",
		})
		return
	}

	// Get thread details
	var thread ForumThread
	var tags []string

	query := `
		SELECT 
			ft.id, ft.user_id, u.username, ft.title, ft.content,
			ft.game_id, g.name as game_name, ft.tags, ft.view_count, ft.reply_count,
			ft.locked, ft.locked_at, ft.pinned, ft.created_at, ft.updated_at
		FROM forum_threads ft
		JOIN users u ON ft.user_id = u.id
		LEFT JOIN games g ON ft.game_id = g.id
		WHERE ft.id = $1 AND ft.is_deleted = FALSE
	`

	err = h.db.QueryRow(c.Request.Context(), query, threadID).Scan(
		&thread.ID, &thread.UserID, &thread.Username, &thread.Title, &thread.Content,
		&thread.GameID, &thread.GameName, &tags, &thread.ViewCount, &thread.ReplyCount,
		&thread.Locked, &thread.LockedAt, &thread.Pinned, &thread.CreatedAt, &thread.UpdatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Thread not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to retrieve thread",
		})
		return
	}

	thread.Tags = []string(tags)
	if thread.Tags == nil {
		thread.Tags = []string{}
	}

	// Increment view count atomically in the database
	// Using UPDATE directly ensures atomic increment without race conditions
	// Errors are ignored as view count increment is non-critical
	_, _ = h.db.Exec(c.Request.Context(),
		"UPDATE forum_threads SET view_count = view_count + 1 WHERE id = $1",
		threadID)

	// Get all replies for the thread using ltree path for efficient hierarchical queries
	replies, err := h.getThreadRepliesHierarchical(c.Request.Context(), threadID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to retrieve replies",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"thread":  thread,
			"replies": replies,
		},
	})
}

// getThreadRepliesHierarchical retrieves all replies for a thread in hierarchical structure using ltree
func (h *ForumHandler) getThreadRepliesHierarchical(ctx context.Context, threadID uuid.UUID) ([]ForumReply, error) {
	// Get all replies ordered by path for hierarchical reconstruction
	query := `
		SELECT 
			fr.id, fr.user_id, u.username, fr.thread_id, fr.parent_reply_id,
			fr.content, fr.depth, fr.path::text, fr.created_at, fr.updated_at
		FROM forum_replies fr
		JOIN users u ON fr.user_id = u.id
		WHERE fr.thread_id = $1 AND fr.is_deleted = FALSE
		ORDER BY fr.path
	`

	rows, err := h.db.Query(ctx, query, threadID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// Map to store replies by ID for building hierarchy
	replyMap := make(map[uuid.UUID]*ForumReply)
	var rootReplies []*ForumReply

	// First pass: create all reply objects and store in map
	for rows.Next() {
		var reply ForumReply
		err := rows.Scan(
			&reply.ID, &reply.UserID, &reply.Username, &reply.ThreadID, &reply.ParentReplyID,
			&reply.Content, &reply.Depth, &reply.Path, &reply.CreatedAt, &reply.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		reply.Replies = []ForumReply{}
		replyMap[reply.ID] = &reply
	}

	// Second pass: build hierarchy by connecting children to parents
	for _, reply := range replyMap {
		if reply.ParentReplyID == nil {
			// Root level reply
			rootReplies = append(rootReplies, reply)
		} else {
			// Child reply - add to parent's replies
			if parent, ok := replyMap[*reply.ParentReplyID]; ok {
				// Add reference to parent so all nested children are included
				parent.Replies = append(parent.Replies, *reply)
			}
			// If parent not found, this is an orphaned reply - skip it
		}
	}

	// Third pass: recursively copy the tree structure to handle deep nesting
	// This ensures that when we dereference pointers, all nested children are included
	var copyReplyTree func(*ForumReply) ForumReply
	copyReplyTree = func(r *ForumReply) ForumReply {
		copied := *r
		copied.Replies = make([]ForumReply, 0, len(r.Replies))
		for i := range r.Replies {
			// Get the pointer from the map to ensure we have the latest children
			if childPtr, ok := replyMap[r.Replies[i].ID]; ok {
				copied.Replies = append(copied.Replies, copyReplyTree(childPtr))
			} else {
				// Fallback to value if not in map
				copied.Replies = append(copied.Replies, r.Replies[i])
			}
		}
		return copied
	}

	// Convert root replies with full nested structure
	result := make([]ForumReply, 0, len(rootReplies))
	for _, rootPtr := range rootReplies {
		result = append(result, copyReplyTree(rootPtr))
	}

	return result, nil
}

// CreateReplyRequest represents the request to create a reply
type CreateReplyRequest struct {
	Content       string  `json:"content" binding:"required,min=1,max=3000"`
	ParentReplyID *string `json:"parent_reply_id"`
}

// CreateReply creates a new reply to a thread or another reply
// POST /api/v1/forum/threads/:id/replies
func (h *ForumHandler) CreateReply(c *gin.Context) {
	threadID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid thread ID",
		})
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "User not authenticated",
		})
		return
	}

	var req CreateReplyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request: " + err.Error(),
		})
		return
	}

	tx, err := h.db.Begin(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to start transaction",
		})
		return
	}
	defer tx.Rollback(c.Request.Context())

	// Check if thread exists and is not locked
	var isLocked bool
	var isDeleted bool
	err = tx.QueryRow(c.Request.Context(),
		`SELECT locked, is_deleted FROM forum_threads WHERE id = $1`,
		threadID).Scan(&isLocked, &isDeleted)

	if err != nil {
		if err == pgx.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Thread not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to check thread status",
		})
		return
	}

	if isDeleted {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Thread not found",
		})
		return
	}

	if isLocked {
		c.JSON(http.StatusForbidden, gin.H{
			"error": "Thread is locked",
		})
		return
	}

	// Determine depth and path
	var depth int
	var path string
	replyID := uuid.New()
	// Use full UUID string for ltree path to avoid collisions
	// Replace hyphens with underscores as ltree doesn't support hyphens
	replyIDForPath := strings.ReplaceAll(replyID.String(), "-", "_")

	var parentReplyID *uuid.UUID
	if req.ParentReplyID != nil && *req.ParentReplyID != "" {
		parsedParentID, err := uuid.Parse(*req.ParentReplyID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid parent_reply_id format",
			})
			return
		}
		parentReplyID = &parsedParentID

		// Get parent's depth and path
		var parentDepth int
		var parentPath string
		err = tx.QueryRow(c.Request.Context(),
			`SELECT depth, path::text FROM forum_replies 
			 WHERE id = $1 AND thread_id = $2 AND is_deleted = FALSE`,
			parentReplyID, threadID).Scan(&parentDepth, &parentPath)

		if err != nil {
			if err == pgx.ErrNoRows {
				c.JSON(http.StatusNotFound, gin.H{
					"error": "Parent reply not found",
				})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to get parent reply",
			})
			return
		}

		// Check depth limit (max 10 levels: 0-9)
		if parentDepth >= 9 {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Maximum nesting depth (10 levels) reached",
			})
			return
		}

		depth = parentDepth + 1
		path = fmt.Sprintf("%s.%s", parentPath, replyIDForPath)
	} else {
		// Root level reply
		depth = 0
		path = replyIDForPath
	}

	// Insert reply
	query := `
		INSERT INTO forum_replies (id, user_id, thread_id, parent_reply_id, content, depth, path)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING created_at
	`

	var createdAt time.Time
	err = tx.QueryRow(c.Request.Context(), query,
		replyID, userID, threadID, parentReplyID, req.Content, depth, path,
	).Scan(&createdAt)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to create reply",
		})
		return
	}

	if err := tx.Commit(c.Request.Context()); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to commit transaction",
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"data": gin.H{
			"id":         replyID,
			"depth":      depth,
			"path":       path,
			"created_at": createdAt,
		},
	})
}

// UpdateReplyRequest represents the request to update a reply
type UpdateReplyRequest struct {
	Content string `json:"content" binding:"required,min=1,max=3000"`
}

// UpdateReply updates an existing reply
// PATCH /api/v1/forum/replies/:id
func (h *ForumHandler) UpdateReply(c *gin.Context) {
	replyID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid reply ID",
		})
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "User not authenticated",
		})
		return
	}

	var req UpdateReplyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request: " + err.Error(),
		})
		return
	}

	// Check ownership and update
	result, err := h.db.Exec(c.Request.Context(),
		`UPDATE forum_replies 
		 SET content = $1, updated_at = NOW()
		 WHERE id = $2 AND user_id = $3 AND is_deleted = FALSE`,
		req.Content, replyID, userID)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to update reply",
		})
		return
	}

	if result.RowsAffected() == 0 {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Reply not found or you don't have permission to edit it",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Reply updated successfully",
	})
}

// DeleteReply soft deletes a reply
// DELETE /api/v1/forum/replies/:id
func (h *ForumHandler) DeleteReply(c *gin.Context) {
	replyID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid reply ID",
		})
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "User not authenticated",
		})
		return
	}

	// Soft delete by setting is_deleted flag
	result, err := h.db.Exec(c.Request.Context(),
		`UPDATE forum_replies 
		 SET is_deleted = TRUE, updated_at = NOW()
		 WHERE id = $1 AND user_id = $2 AND is_deleted = FALSE`,
		replyID, userID)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to delete reply",
		})
		return
	}

	if result.RowsAffected() == 0 {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Reply not found or you don't have permission to delete it",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Reply deleted successfully",
	})
}

// ForumSearchResult represents a unified search result that can be either a thread or reply
type ForumSearchResult struct {
	Type       string     `json:"type"` // 'thread' or 'reply'
	ID         uuid.UUID  `json:"id"`
	Title      *string    `json:"title,omitempty"` // Only for threads
	Body       string     `json:"body"`            // Thread content or reply content
	AuthorID   uuid.UUID  `json:"author_id"`
	AuthorName string     `json:"author_name"`
	ThreadID   *uuid.UUID `json:"thread_id,omitempty"` // Only for replies
	VoteCount  int        `json:"vote_count"`
	CreatedAt  time.Time  `json:"created_at"`
	Headline   string     `json:"headline"` // Highlighted snippet
	Rank       float64    `json:"rank"`     // Relevance score
}

// SearchThreads performs full-text search on threads and replies with filters
// GET /api/v1/forum/search?q=<query>&author=<username>&sort=<relevance|date|votes>&page=1
func (h *ForumHandler) SearchThreads(c *gin.Context) {
	searchQuery := c.Query("q")
	if searchQuery == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Search query is required",
		})
		return
	}

	author := c.Query("author")
	sortBy := c.DefaultQuery("sort", "relevance") // relevance, date, votes
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	if page < 1 {
		page = 1
	}
	limit := 50
	offset := (page - 1) * limit

	// Validate sort parameter
	if sortBy != "relevance" && sortBy != "date" && sortBy != "votes" {
		sortBy = "relevance"
	}

	// Build the query with UNION of threads and replies
	var queryBuilder strings.Builder
	args := []interface{}{searchQuery}
	argIdx := 2

	// Search threads
	queryBuilder.WriteString(`
		SELECT 
			'thread' as type,
			ft.id,
			ft.title,
			ft.content as body,
			ft.user_id as author_id,
			u.username as author_name,
			NULL::uuid as thread_id,
			-- Threads do not support direct voting; this placeholder keeps UNION schema compatible with replies
			0 as vote_count,
			ft.created_at,
			ts_headline('english', 
				COALESCE(ft.title, '') || ' ' || COALESCE(ft.content, ''),
				plainto_tsquery('english', $1),
				'MaxWords=50, MinWords=30, MaxFragments=1'
			) as headline,
			ts_rank(ft.search_vector, plainto_tsquery('english', $1)) as rank
		FROM forum_threads ft
		JOIN users u ON ft.user_id = u.id
		WHERE ft.is_deleted = FALSE
			AND ft.search_vector @@ plainto_tsquery('english', $1)
	`)

	// Store the author parameter index if author filter is used
	authorParamIdx := 0
	if author != "" {
		queryBuilder.WriteString(fmt.Sprintf(" AND u.username = $%d", argIdx))
		args = append(args, author)
		authorParamIdx = argIdx
		argIdx++
	}

	// Union with replies
	queryBuilder.WriteString(`
		UNION ALL
		SELECT 
			'reply' as type,
			fr.id,
			NULL as title,
			fr.content as body,
			fr.user_id as author_id,
			u.username as author_name,
			fr.thread_id,
			COALESCE(vc.net_votes, 0) as vote_count,
			fr.created_at,
			ts_headline('english',
				COALESCE(fr.content, ''),
				plainto_tsquery('english', $1),
				'MaxWords=50, MinWords=30, MaxFragments=1'
			) as headline,
			ts_rank(fr.search_vector, plainto_tsquery('english', $1)) as rank
		FROM forum_replies fr
		JOIN users u ON fr.user_id = u.id
		LEFT JOIN forum_vote_counts vc ON fr.id = vc.reply_id
		WHERE fr.is_deleted = FALSE
			AND fr.search_vector @@ plainto_tsquery('english', $1)
	`)

	// Reuse the same author parameter index for replies
	if author != "" {
		queryBuilder.WriteString(fmt.Sprintf(" AND u.username = $%d", authorParamIdx))
	}

	// Add ORDER BY based on sort parameter
	switch sortBy {
	case "date":
		queryBuilder.WriteString(" ORDER BY created_at DESC")
	case "votes":
		queryBuilder.WriteString(" ORDER BY vote_count DESC, created_at DESC")
	default: // relevance
		queryBuilder.WriteString(" ORDER BY rank DESC, created_at DESC")
	}

	// Add pagination
	queryBuilder.WriteString(fmt.Sprintf(" LIMIT $%d OFFSET $%d", argIdx, argIdx+1))
	args = append(args, limit, offset)

	// Execute query
	rows, err := h.db.Query(c.Request.Context(), queryBuilder.String(), args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to search forum",
		})
		return
	}
	defer rows.Close()

	results := []ForumSearchResult{}
	for rows.Next() {
		var result ForumSearchResult
		err := rows.Scan(
			&result.Type,
			&result.ID,
			&result.Title,
			&result.Body,
			&result.AuthorID,
			&result.AuthorName,
			&result.ThreadID,
			&result.VoteCount,
			&result.CreatedAt,
			&result.Headline,
			&result.Rank,
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to scan search result",
			})
			return
		}
		results = append(results, result)
	}

	// Check for errors during iteration
	if err := rows.Err(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to read search results",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    results,
		"meta": gin.H{
			"page":     page,
			"limit":    limit,
			"query":    searchQuery,
			"author":   author,
			"sort":     sortBy,
			"count":    len(results),
			"has_more": len(results) == limit, // Indicates if there might be more results
		},
	})
}

// VoteOnReply votes on a reply (upvote, downvote, or neutral)
// POST /api/v1/forum/replies/:id/vote
func (h *ForumHandler) VoteOnReply(c *gin.Context) {
	replyID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid reply ID",
		})
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "User not authenticated",
		})
		return
	}

	var req struct {
		VoteValue int `json:"vote_value" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request: " + err.Error(),
		})
		return
	}

	// Validate vote value
	if req.VoteValue != -1 && req.VoteValue != 0 && req.VoteValue != 1 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Vote value must be -1, 0, or 1",
		})
		return
	}

	// Check if reply exists and is not deleted
	var replyUserID uuid.UUID
	var isDeleted bool
	err = h.db.QueryRow(c.Request.Context(),
		`SELECT user_id, is_deleted FROM forum_replies WHERE id = $1`,
		replyID).Scan(&replyUserID, &isDeleted)

	if err != nil {
		if err == pgx.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Reply not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to check reply",
		})
		return
	}

	if isDeleted {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Reply not found",
		})
		return
	}

	// Prevent self-voting
	if replyUserID == userID.(uuid.UUID) {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Cannot vote on your own reply",
		})
		return
	}

	// Insert or update vote
	_, err = h.db.Exec(c.Request.Context(), `
		INSERT INTO forum_votes (user_id, reply_id, vote_value)
		VALUES ($1, $2, $3)
		ON CONFLICT (user_id, reply_id) 
		DO UPDATE SET vote_value = EXCLUDED.vote_value, updated_at = NOW()
	`, userID, replyID, req.VoteValue)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to save vote",
		})
		return
	}

	// Update reputation for reply author (async via goroutine for performance)
	// Use a separate context with timeout to prevent goroutine leaks
	handlerBackgroundJobs.Submit(func() {
		// Create background context with timeout
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		// Update reputation - errors are logged for debugging
		if _, err := h.db.Exec(ctx, "SELECT update_reputation_score($1)", replyUserID); err != nil {
			// Log error but don't fail the vote operation
			fmt.Printf("Warning: Failed to update reputation for user %s: %v\n", replyUserID, err)
		}
	})

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Vote recorded successfully",
	})
}

// GetReplyVotes retrieves vote statistics for a reply
// GET /api/v1/forum/replies/:id/votes
func (h *ForumHandler) GetReplyVotes(c *gin.Context) {
	replyID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid reply ID",
		})
		return
	}

	// Get authenticated user ID (may be nil for guests)
	userID, exists := c.Get("user_id")
	var userUUID *uuid.UUID
	if exists {
		if id, ok := userID.(uuid.UUID); ok {
			userUUID = &id
		}
	}

	var stats VoteStats

	// Build query based on whether user is authenticated
	if userUUID != nil {
		err = h.db.QueryRow(c.Request.Context(), `
			SELECT 
				COALESCE(vc.upvote_count, 0) as upvotes,
				COALESCE(vc.downvote_count, 0) as downvotes,
				COALESCE(vc.net_votes, 0) as net_votes,
				COALESCE(v.vote_value, 0) as user_vote
			FROM forum_replies fr
			LEFT JOIN forum_vote_counts vc ON fr.id = vc.reply_id
			LEFT JOIN forum_votes v ON fr.id = v.reply_id AND v.user_id = $2
			WHERE fr.id = $1
		`, replyID, userUUID).Scan(&stats.Upvotes, &stats.Downvotes, &stats.NetVotes, &stats.UserVote)
	} else {
		// Guest user - no user vote
		err = h.db.QueryRow(c.Request.Context(), `
			SELECT 
				COALESCE(vc.upvote_count, 0) as upvotes,
				COALESCE(vc.downvote_count, 0) as downvotes,
				COALESCE(vc.net_votes, 0) as net_votes
			FROM forum_replies fr
			LEFT JOIN forum_vote_counts vc ON fr.id = vc.reply_id
			WHERE fr.id = $1
		`, replyID).Scan(&stats.Upvotes, &stats.Downvotes, &stats.NetVotes)
		stats.UserVote = 0
	}

	if err != nil {
		if err == pgx.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Reply not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to retrieve vote statistics",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    stats,
	})
}

// GetUserReputation retrieves reputation information for a user
// GET /api/v1/forum/users/:id/reputation
func (h *ForumHandler) GetUserReputation(c *gin.Context) {
	userID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid user ID",
		})
		return
	}

	var rep ReputationScore
	err = h.db.QueryRow(c.Request.Context(), `
		SELECT 
			user_id, 
			reputation_score, 
			reputation_badge, 
			total_votes, 
			threads_created, 
			replies_created, 
			last_updated
		FROM user_reputation
		WHERE user_id = $1
	`, userID).Scan(
		&rep.UserID,
		&rep.Score,
		&rep.Badge,
		&rep.Votes,
		&rep.Threads,
		&rep.Replies,
		&rep.UpdatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			// User has no reputation yet - return default
			rep = ReputationScore{
				UserID:    userID,
				Score:     0,
				Badge:     "new",
				Votes:     0,
				Threads:   0,
				Replies:   0,
				UpdatedAt: time.Now(),
			}
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to retrieve reputation",
			})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    rep,
	})
}

// DetectSpamReplies identifies and flags replies as spam based on vote patterns
// This is typically run as a background job
func (h *ForumHandler) DetectSpamReplies(ctx context.Context) error {
	query := `
		UPDATE forum_replies 
		SET flagged_as_spam = TRUE
		WHERE id IN (
			SELECT reply_id 
			FROM forum_vote_counts
			WHERE downvote_count > 5 AND net_votes < -2
		)
		AND flagged_as_spam = FALSE
		AND is_deleted = FALSE
	`
	_, err := h.db.Exec(ctx, query)
	return err
}

// HideLowQualityReplies hides replies with very negative vote scores
// This is typically run as a background job
func (h *ForumHandler) HideLowQualityReplies(ctx context.Context) error {
	query := `
		UPDATE forum_replies 
		SET hidden = TRUE
		WHERE id IN (
			SELECT reply_id 
			FROM forum_vote_counts
			WHERE net_votes <= -5
		)
		AND hidden = FALSE
		AND is_deleted = FALSE
	`
	_, err := h.db.Exec(ctx, query)
	return err
}

// ForumAnalytics represents analytics data for the forum
type ForumAnalytics struct {
	TotalThreads     int                `json:"total_threads"`
	TotalReplies     int                `json:"total_replies"`
	TotalUsers       int                `json:"total_users"`
	PostsToday       int                `json:"posts_today"`
	PostsThisWeek    int                `json:"posts_this_week"`
	PostsThisMonth   int                `json:"posts_this_month"`
	ActiveUsersToday int                `json:"active_users_today"`
	ActiveUsersWeek  int                `json:"active_users_week"`
	TrendingTopics   []string           `json:"trending_topics"`
	PopularThreads   []ForumThread      `json:"popular_threads"`
	TopContributors  []UserContribution `json:"top_contributors"`
	LastUpdated      time.Time          `json:"last_updated"`
}

// UserContribution represents a user's forum contributions
type UserContribution struct {
	UserID          uuid.UUID `json:"user_id"`
	Username        string    `json:"username"`
	ThreadCount     int       `json:"thread_count"`
	ReplyCount      int       `json:"reply_count"`
	ReputationScore int       `json:"reputation_score"`
}

// GetForumAnalytics retrieves forum analytics data
// GET /api/v1/forum/analytics
func (h *ForumHandler) GetForumAnalytics(c *gin.Context) {
	ctx := c.Request.Context()

	var analytics ForumAnalytics
	analytics.LastUpdated = time.Now()

	// Get total threads
	err := h.db.QueryRow(ctx, `
		SELECT COUNT(*) FROM forum_threads WHERE is_deleted = FALSE
	`).Scan(&analytics.TotalThreads)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get thread count"})
		return
	}

	// Get total replies
	err = h.db.QueryRow(ctx, `
		SELECT COUNT(*) FROM forum_replies WHERE is_deleted = FALSE
	`).Scan(&analytics.TotalReplies)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get reply count"})
		return
	}

	// Get total unique forum users
	err = h.db.QueryRow(ctx, `
		SELECT COUNT(DISTINCT user_id) FROM (
			SELECT user_id FROM forum_threads WHERE is_deleted = FALSE
			UNION
			SELECT user_id FROM forum_replies WHERE is_deleted = FALSE
		) u
	`).Scan(&analytics.TotalUsers)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user count"})
		return
	}

	// Get posts today (threads + replies)
	err = h.db.QueryRow(ctx, `
		SELECT COUNT(*) FROM (
			SELECT created_at FROM forum_threads 
			WHERE is_deleted = FALSE AND created_at >= CURRENT_DATE
			UNION ALL
			SELECT created_at FROM forum_replies 
			WHERE is_deleted = FALSE AND created_at >= CURRENT_DATE
		) posts
	`).Scan(&analytics.PostsToday)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get today's posts"})
		return
	}

	// Get posts this week
	err = h.db.QueryRow(ctx, `
		SELECT COUNT(*) FROM (
			SELECT created_at FROM forum_threads 
			WHERE is_deleted = FALSE AND created_at >= CURRENT_DATE - INTERVAL '7 days'
			UNION ALL
			SELECT created_at FROM forum_replies 
			WHERE is_deleted = FALSE AND created_at >= CURRENT_DATE - INTERVAL '7 days'
		) posts
	`).Scan(&analytics.PostsThisWeek)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get week's posts"})
		return
	}

	// Get posts this month
	err = h.db.QueryRow(ctx, `
		SELECT COUNT(*) FROM (
			SELECT created_at FROM forum_threads 
			WHERE is_deleted = FALSE AND created_at >= CURRENT_DATE - INTERVAL '30 days'
			UNION ALL
			SELECT created_at FROM forum_replies 
			WHERE is_deleted = FALSE AND created_at >= CURRENT_DATE - INTERVAL '30 days'
		) posts
	`).Scan(&analytics.PostsThisMonth)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get month's posts"})
		return
	}

	// Get active users today
	err = h.db.QueryRow(ctx, `
		SELECT COUNT(DISTINCT user_id) FROM (
			SELECT user_id FROM forum_threads 
			WHERE is_deleted = FALSE AND created_at >= CURRENT_DATE
			UNION
			SELECT user_id FROM forum_replies 
			WHERE is_deleted = FALSE AND created_at >= CURRENT_DATE
		) u
	`).Scan(&analytics.ActiveUsersToday)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get active users today"})
		return
	}

	// Get active users this week
	err = h.db.QueryRow(ctx, `
		SELECT COUNT(DISTINCT user_id) FROM (
			SELECT user_id FROM forum_threads 
			WHERE is_deleted = FALSE AND created_at >= CURRENT_DATE - INTERVAL '7 days'
			UNION
			SELECT user_id FROM forum_replies 
			WHERE is_deleted = FALSE AND created_at >= CURRENT_DATE - INTERVAL '7 days'
		) u
	`).Scan(&analytics.ActiveUsersWeek)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get active users this week"})
		return
	}

	// Get trending topics (most used tags in the last 7 days)
	rows, err := h.db.Query(ctx, `
		SELECT unnest(tags) as tag, COUNT(*) as count
		FROM forum_threads
		WHERE is_deleted = FALSE 
			AND created_at >= CURRENT_DATE - INTERVAL '7 days'
			AND tags IS NOT NULL
		GROUP BY tag
		ORDER BY count DESC
		LIMIT 10
	`)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get trending topics"})
		return
	}
	defer rows.Close()

	analytics.TrendingTopics = make([]string, 0)
	for rows.Next() {
		var topic string
		var count int
		if err := rows.Scan(&topic, &count); err != nil {
			continue
		}
		analytics.TrendingTopics = append(analytics.TrendingTopics, topic)
	}

	// Get popular threads (most replies + views in last 30 days)
	threadRows, err := h.db.Query(ctx, `
		SELECT 
			ft.id, ft.user_id, u.username, ft.title, ft.content,
			ft.game_id, g.title as game_name, ft.tags,
			ft.view_count, ft.reply_count, ft.locked, ft.locked_at,
			ft.pinned, ft.created_at, ft.updated_at
		FROM forum_threads ft
		INNER JOIN users u ON ft.user_id = u.id
		LEFT JOIN games g ON ft.game_id = g.id
		WHERE ft.is_deleted = FALSE
			AND ft.created_at >= CURRENT_DATE - INTERVAL '30 days'
		ORDER BY (ft.reply_count * 2 + ft.view_count) DESC
		LIMIT 10
	`)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get popular threads"})
		return
	}
	defer threadRows.Close()

	analytics.PopularThreads = make([]ForumThread, 0)
	for threadRows.Next() {
		var thread ForumThread
		var gameName *string
		err := threadRows.Scan(
			&thread.ID, &thread.UserID, &thread.Username, &thread.Title, &thread.Content,
			&thread.GameID, &gameName, &thread.Tags,
			&thread.ViewCount, &thread.ReplyCount, &thread.Locked, &thread.LockedAt,
			&thread.Pinned, &thread.CreatedAt, &thread.UpdatedAt,
		)
		if err != nil {
			continue
		}
		thread.GameName = gameName
		analytics.PopularThreads = append(analytics.PopularThreads, thread)
	}

	// Get top contributors (users with most activity in last 30 days)
	contribRows, err := h.db.Query(ctx, `
		SELECT 
			u.id, u.username,
			COALESCE(t.thread_count, 0) as thread_count,
			COALESCE(r.reply_count, 0) as reply_count,
			COALESCE(ur.reputation_score, 0) as reputation_score
		FROM users u
		LEFT JOIN (
			SELECT user_id, COUNT(*) as thread_count
			FROM forum_threads
			WHERE is_deleted = FALSE AND created_at >= CURRENT_DATE - INTERVAL '30 days'
			GROUP BY user_id
		) t ON u.id = t.user_id
		LEFT JOIN (
			SELECT user_id, COUNT(*) as reply_count
			FROM forum_replies
			WHERE is_deleted = FALSE AND created_at >= CURRENT_DATE - INTERVAL '30 days'
			GROUP BY user_id
		) r ON u.id = r.user_id
		LEFT JOIN user_reputation ur ON u.id = ur.user_id
		WHERE (t.thread_count > 0 OR r.reply_count > 0)
		ORDER BY (COALESCE(t.thread_count, 0) + COALESCE(r.reply_count, 0)) DESC
		LIMIT 10
	`)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get top contributors"})
		return
	}
	defer contribRows.Close()

	analytics.TopContributors = make([]UserContribution, 0)
	for contribRows.Next() {
		var contrib UserContribution
		if err := contribRows.Scan(&contrib.UserID, &contrib.Username, &contrib.ThreadCount, &contrib.ReplyCount, &contrib.ReputationScore); err != nil {
			continue
		}
		analytics.TopContributors = append(analytics.TopContributors, contrib)
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    analytics,
	})
}

// GetPopularDiscussions retrieves popular discussions dashboard
// GET /api/v1/forum/popular
func (h *ForumHandler) GetPopularDiscussions(c *gin.Context) {
	ctx := c.Request.Context()
	timeframe := c.DefaultQuery("timeframe", "week") // day, week, month, all
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	if limit < 1 || limit > 100 {
		limit = 20
	}

	var intervalDays int
	switch timeframe {
	case "day":
		intervalDays = 1
	case "week":
		intervalDays = 7
	case "month":
		intervalDays = 30
	case "all":
		intervalDays = 3650 // ~10 years
	default:
		intervalDays = 7
	}

	query := `
		SELECT 
			ft.id, ft.user_id, u.username, ft.title, ft.content,
			ft.game_id, g.title as game_name, ft.tags,
			ft.view_count, ft.reply_count, ft.locked, ft.locked_at,
			ft.pinned, ft.created_at, ft.updated_at
		FROM forum_threads ft
		INNER JOIN users u ON ft.user_id = u.id
		LEFT JOIN games g ON ft.game_id = g.id
		WHERE ft.is_deleted = FALSE
			AND ft.created_at >= CURRENT_DATE - ($2 || ' days')::INTERVAL
		ORDER BY (ft.reply_count * 3 + ft.view_count / 10) DESC
		LIMIT $1
	`

	rows, err := h.db.Query(ctx, query, limit, intervalDays)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get popular discussions"})
		return
	}
	defer rows.Close()

	threads := make([]ForumThread, 0)
	for rows.Next() {
		var thread ForumThread
		var gameName *string
		err := rows.Scan(
			&thread.ID, &thread.UserID, &thread.Username, &thread.Title, &thread.Content,
			&thread.GameID, &gameName, &thread.Tags,
			&thread.ViewCount, &thread.ReplyCount, &thread.Locked, &thread.LockedAt,
			&thread.Pinned, &thread.CreatedAt, &thread.UpdatedAt,
		)
		if err != nil {
			continue
		}
		thread.GameName = gameName
		threads = append(threads, thread)
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    threads,
		"meta": gin.H{
			"timeframe": timeframe,
			"count":     len(threads),
			"limit":     limit,
		},
	})
}

// GetMostHelpfulReplies retrieves most helpful (highly voted) replies
// GET /api/v1/forum/helpful-replies
func (h *ForumHandler) GetMostHelpfulReplies(c *gin.Context) {
	ctx := c.Request.Context()
	timeframe := c.DefaultQuery("timeframe", "month") // week, month, all
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	if limit < 1 || limit > 100 {
		limit = 20
	}

	var intervalDays int
	switch timeframe {
	case "week":
		intervalDays = 7
	case "month":
		intervalDays = 30
	case "all":
		intervalDays = 3650 // ~10 years
	default:
		intervalDays = 30
	}

	query := `
		SELECT 
			fr.id, fr.user_id, u.username, fr.thread_id, 
			fr.parent_reply_id, fr.content, fr.depth, fr.path,
			fr.created_at, fr.updated_at,
			ft.title as thread_title,
			COALESCE(fvc.net_votes, 0) as net_votes,
			COALESCE(fvc.upvote_count, 0) as upvotes,
			COALESCE(fvc.downvote_count, 0) as downvotes
		FROM forum_replies fr
		INNER JOIN users u ON fr.user_id = u.id
		INNER JOIN forum_threads ft ON fr.thread_id = ft.id
		LEFT JOIN forum_vote_counts fvc ON fr.id = fvc.reply_id
		WHERE fr.is_deleted = FALSE
			AND fr.created_at >= CURRENT_DATE - ($2 || ' days')::INTERVAL
		ORDER BY COALESCE(fvc.net_votes, 0) DESC, fr.created_at DESC
		LIMIT $1
	`

	rows, err := h.db.Query(ctx, query, limit, intervalDays)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get helpful replies"})
		return
	}
	defer rows.Close()

	type HelpfulReply struct {
		ForumReply
		ThreadTitle string `json:"thread_title"`
		NetVotes    int    `json:"net_votes"`
		Upvotes     int    `json:"upvotes"`
		Downvotes   int    `json:"downvotes"`
	}

	replies := make([]HelpfulReply, 0)
	for rows.Next() {
		var reply HelpfulReply
		err := rows.Scan(
			&reply.ID, &reply.UserID, &reply.Username, &reply.ThreadID,
			&reply.ParentReplyID, &reply.Content, &reply.Depth, &reply.Path,
			&reply.CreatedAt, &reply.UpdatedAt,
			&reply.ThreadTitle, &reply.NetVotes, &reply.Upvotes, &reply.Downvotes,
		)
		if err != nil {
			continue
		}
		replies = append(replies, reply)
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    replies,
		"meta": gin.H{
			"timeframe": timeframe,
			"count":     len(replies),
			"limit":     limit,
		},
	})
}
