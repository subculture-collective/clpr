# Comment Repository Methods Documentation

## Overview

The Comment Repository provides optimized methods for working with nested comment threading in the Clipper application. The system supports Reddit-style nested comments with efficient tree traversal and pagination.

## Key Features

- **Maximum Nesting Depth**: 10 levels
- **Denormalized Reply Counts**: Avoid expensive COUNT subqueries
- **Recursive CTEs**: Efficient tree traversal without N+1 queries
- **Vote Integration**: User votes included in all comment queries
- **Author Information**: Full user profile data embedded in responses

## Core Methods

### GetCommentTree(ctx, parentID, userID)

Retrieves a complete nested comment tree starting from a specific parent comment.

**Use Case**: Load a full conversation thread when user clicks "View All Replies"

**Features**:
- Uses recursive CTE for efficient tree traversal
- Returns comments in depth-first order
- Includes all replies up to max nesting depth (10)
- Single database query eliminates N+1 problem

**Returns**: Flat list of `CommentWithAuthor` sorted by conversation path

**Example**:
```go
comments, err := repo.GetCommentTree(ctx, parentCommentID, &userID)
```

### GetTopLevelComments(ctx, clipID, limit, offset, userID)

Retrieves top-level comments for a clip with pagination.

**Use Case**: Initial page load of comment section

**Features**:
- Returns only top-level comments (where `parent_comment_id IS NULL`)
- Sorted by vote score DESC, then created_at DESC
- Includes `reply_count` for showing "X replies" badge
- Supports pagination for large comment sections

**Returns**: Array of `CommentWithAuthor` with pagination support

**Example**:
```go
comments, err := repo.GetTopLevelComments(ctx, clipID, 20, 0, &userID)
```

### GetReplies(ctx, parentID, limit, offset, userID)

Retrieves direct replies to a specific comment with pagination.

**Use Case**: Load immediate children when user clicks "Load More Replies"

**Features**:
- Returns only direct children (one level deep)
- Sorted by vote score and creation time
- Includes each reply's `reply_count` for nested reply indicators
- Supports pagination for comments with many replies

**Returns**: Array of `CommentWithAuthor` with pagination

**Example**:
```go
replies, err := repo.GetReplies(ctx, parentCommentID, 10, 0, &userID)
```

### ListByClipID(ctx, clipID, sortBy, limit, offset, userID)

Advanced listing with multiple sort options.

**Sort Options**:
- `"best"` (default): Wilson score confidence interval
- `"new"`: Most recent first
- `"old"`: Oldest first
- `"controversial"`: High engagement, mixed votes

**Use Case**: Primary method for fetching comments with user-selected sorting

**Example**:
```go
comments, err := repo.ListByClipID(ctx, clipID, "best", 20, 0, &userID)
```

## Database Schema

### Comments Table

```sql
CREATE TABLE comments (
    id UUID PRIMARY KEY,
    clip_id UUID REFERENCES clips(id),
    user_id UUID REFERENCES users(id),
    parent_comment_id UUID REFERENCES comments(id),
    content TEXT,
    vote_score INT DEFAULT 0,
    reply_count INT DEFAULT 0,  -- Denormalized for performance
    is_edited BOOLEAN DEFAULT false,
    is_removed BOOLEAN DEFAULT false,
    removed_reason TEXT,
    created_at TIMESTAMP,
    updated_at TIMESTAMP
);
```

### Key Indexes

```sql
-- Efficient reply queries
CREATE INDEX idx_comments_parent_comment_id ON comments(parent_comment_id) 
WHERE parent_comment_id IS NOT NULL;

-- Top-level comment queries
CREATE INDEX idx_comments_clip ON comments(clip_id, created_at);

-- Vote score sorting
CREATE INDEX idx_comments_score ON comments(vote_score DESC);
```

## Reply Count Maintenance

Reply counts are automatically maintained by database triggers:

- **INSERT**: Increments parent's `reply_count`
- **DELETE**: Decrements parent's `reply_count`
- **UPDATE**: Handles parent_comment_id changes

This eliminates the need for expensive `COUNT(*)` subqueries in SELECT statements.

## Performance Considerations

### Avoiding N+1 Queries

The repository uses several strategies to avoid N+1 query patterns:

1. **Recursive CTEs**: `GetCommentTree()` fetches entire subtrees in one query
2. **Denormalized Counts**: `reply_count` avoids COUNT subqueries
3. **Vote Preloading**: User votes joined in primary query
4. **Author Embedding**: User data included via JOIN

### Query Complexity

- **Top-level comments**: O(1) with proper indexes
- **Comment tree**: O(depth) with recursive CTE
- **Reply pagination**: O(log n) with indexed parent_id

### Recommended Usage Patterns

**Initial Page Load**:
```go
// Get first 20 top-level comments
topComments := repo.GetTopLevelComments(ctx, clipID, 20, 0, userID)
```

**Load More Top-Level**:
```go
// Get next 20 top-level comments
moreComments := repo.GetTopLevelComments(ctx, clipID, 20, 20, userID)
```

**Expand Reply Thread**:
```go
// Get direct replies to a comment
replies := repo.GetReplies(ctx, commentID, 10, 0, userID)
```

**View Full Conversation**:
```go
// Get entire nested thread
fullThread := repo.GetCommentTree(ctx, commentID, userID)
```

## Depth Validation

Maximum nesting depth is enforced in the service layer:

```go
const MaxNestingDepth = 10
```

The service validates depth before allowing comment creation:

```go
depth, _ := repo.GetNestingDepth(ctx, parentID)
if depth >= MaxNestingDepth {
    return fmt.Errorf("maximum nesting depth of %d reached", MaxNestingDepth)
}
```

## Load Testing

The system has been designed to handle:

- 1000+ comments per clip
- Deep conversation threads (10 levels)
- High-traffic scenarios with proper indexing

For load testing, use the provided seed data:
```bash
psql -U clpr -d clpr < migrations/seed_load_test.sql
```

## Migration History

- **000046**: Added `reply_count` column, triggers, and parent_comment_id index

## Future Enhancements

Potential improvements for v2:

- Comment tree caching (Redis)
- Real-time updates via WebSocket
- Comment search and filtering
- Threaded notification system
- AI-powered comment moderation
