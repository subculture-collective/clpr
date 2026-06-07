---
title: "Comments"
summary: "The Clipper comment system provides Reddit-style nested comment threading with voting, markdown support, and moderation controls. This document provides a comprehensive overview of the feature, includ"
tags: ["features"]
area: "features"
status: "stable"
owner: "team-core"
version: "1.0"
last_reviewed: 2026-01-29
---

# Comment System Feature Documentation

## Overview

The Clipper comment system provides Reddit-style nested comment threading with voting, markdown support, and moderation controls. This document provides a comprehensive overview of the feature, including functionality, performance characteristics, testing procedures, and usage examples.

## Table of Contents

- [Feature Overview](#feature-overview)
- [Reddit-Style Threading](#reddit-style-threading)
- [Key Features](#key-features)
- [Architecture](#architecture)
- [Performance Targets](#performance-targets)
- [E2E Testing Procedures](#e2e-testing-procedures)
- [API Examples](#api-examples)
- [Migration Guide](#migration-guide)
- [Troubleshooting](#troubleshooting)

## Feature Overview

The nested comment system allows users to engage in threaded discussions on clips, similar to Reddit's comment structure. Comments can be nested up to 10 levels deep, with each level indented to show the conversation hierarchy.

### Key Capabilities

- **Nested Threading**: Reply to any comment, creating conversation threads up to 10 levels deep
- **Collapse/Expand**: Toggle visibility of nested replies at any level
- **Voting**: Upvote/downvote comments with optimistic UI updates
- **Markdown Support**: Rich text formatting with sanitized HTML output
- **Real-time Updates**: Comments appear immediately after submission
- **Moderation**: Soft-delete with [deleted]/[removed] markers preserving thread structure
- **Performance**: Optimized for 1000+ comments per clip with smooth scrolling
- **Cross-Platform**: Identical functionality on web and mobile

## Reddit-Style Threading

### Visual Hierarchy

Comments use visual indentation to show nesting depth:

```
└─ Top-level comment
   ├─ Reply (depth 1)
   │  └─ Nested reply (depth 2)
   │     └─ Deep reply (depth 3)
   └─ Another reply (depth 1)
```

### Depth Limit

**Maximum nesting depth: 10 levels**

When a comment reaches depth 10, the "Reply" button is replaced with a "View N more replies" link that directs users to view the full conversation thread on a dedicated page.

**Implementation Details:**

- Depth is calculated from the root comment (parent_comment_id = NULL)
- Depth limit enforced in backend service layer before creating comments
- Frontend displays appropriate UI based on depth
- Link format: `/clips/{clipId}/comments/{commentId}`

### Collapse/Expand

Users can collapse or expand any comment with nested replies:

- **Collapsed**: Shows comment with badge indicating number of hidden replies (e.g., "▶ 5 replies")
- **Expanded**: Shows all nested replies with badge showing collapse option (e.g., "▼ 5 replies")
- **State Management**: Collapse state persists during the session
- **Cross-Device**: State does not sync across devices (client-side only)

## Key Features

### 1. Nested Comment Creation

Users can reply to any comment, creating a nested conversation:

```typescript
POST /api/v1/clips/{clipId}/comments
{
  "content": "Great point!",
  "parent_comment_id": "uuid-of-parent-comment"
}
```

**Validation:**
- Content: 1-10,000 characters
- Parent comment must belong to the same clip
- Maximum depth of 10 levels enforced
- Rate limit: 10 comments per minute

### 2. Comment Voting

Users can upvote (+1) or downvote (-1) any comment:

```typescript
POST /api/v1/comments/{commentId}/vote
{
  "vote": 1  // 1 for upvote, -1 for downvote, 0 to remove vote
}
```

**Performance:**
- Optimistic UI update: < 100ms
- Vote reflected immediately in UI
- Backend processes asynchronously
- Rate limit: 20 votes per minute

### 3. Markdown Formatting

Comments support GitHub Flavored Markdown:

**Allowed:**
- **Bold**, *italic*, ~~strikethrough~~
- Links (auto-nofollow/noreferrer)
- Inline `code` and code blocks
- > Blockquotes
- Lists (ordered/unordered)
- Headers (H1-H6)
- Tables

**Blocked:**
- HTML tags (XSS prevention)
- Images (prevents hotlinking abuse)
- Embedded content

### 4. Comment Editing

Authors can edit their comments within 15 minutes of posting:

```typescript
PUT /api/v1/comments/{commentId}
{
  "content": "Updated content"
}
```

**Restrictions:**
- Authors: 15-minute edit window
- Admins: Can edit anytime
- Sets `is_edited` flag to true
- Edit history not tracked (future enhancement)

### 5. Comment Deletion

Soft-delete preserves thread structure:

- **Author deletion**: Content replaced with "[deleted]"
- **Moderator/admin removal**: Content replaced with "[removed]"
- Nested replies remain visible
- Author loses -1 karma for self-deletion
- Hard delete not available (preserves conversation context)

### 6. Load More Replies

For comments with many replies, pagination prevents overwhelming the UI:

```typescript
GET /api/v1/comments/{commentId}/replies?limit=10&cursor=0
```

- Shows "Load N more replies" link when more replies are available
- Fetches next batch on click
- Maintains scroll position
- Default limit: 50 replies per page

## Architecture

### Backend Stack

- **Database**: PostgreSQL with recursive CTEs for efficient tree traversal
- **Indexes**: Optimized for clip_id, parent_comment_id, vote_score queries
- **Caching**: Redis caching planned for hot comment trees (future)
- **Rate Limiting**: Redis-based rate limiting on creation and voting

### Database Schema

```sql
CREATE TABLE comments (
    id UUID PRIMARY KEY,
    clip_id UUID REFERENCES clips(id),
    user_id UUID REFERENCES users(id),
    parent_comment_id UUID REFERENCES comments(id),
    content TEXT NOT NULL,
    vote_score INT DEFAULT 0,
    reply_count INT DEFAULT 0,  -- Denormalized for performance
    is_edited BOOLEAN DEFAULT false,
    is_removed BOOLEAN DEFAULT false,
    removed_reason TEXT,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);
```

**Key Indexes:**

```sql
CREATE INDEX idx_comments_clip ON comments(clip_id, created_at);
CREATE INDEX idx_comments_parent ON comments(parent_comment_id) 
    WHERE parent_comment_id IS NOT NULL;
CREATE INDEX idx_comments_score ON comments(vote_score DESC);
```

**Triggers:**

- `update_reply_count`: Auto-maintains denormalized reply_count
- `update_comment_vote_score`: Auto-updates vote_score on vote changes
- `update_clip_comment_count`: Auto-updates clip comment_count

### Frontend Stack

- **React**: Component-based UI
- **TanStack Query**: Data fetching and caching
- **Markdown**: react-markdown with remark-gfm
- **State Management**: Zustand for collapse state
- **Optimistic Updates**: Immediate UI feedback for votes/posts

### Mobile Stack

- **React Native**: iOS and Android apps
- **Expo**: Development and build tooling
- **TanStack Query**: Same data fetching as web
- **Native Gestures**: Swipe-to-collapse (future enhancement)

## Performance Targets

### Load Times

| Metric | Target | Measurement |
|--------|--------|-------------|
| Initial load (1000 comments) | < 2s | Time to first render |
| Expand deep thread | < 500ms | From click to visible |
| Scroll 1000 nested comments | 60 FPS | Smooth scrolling |
| Vote optimistic update | < 100ms | UI feedback time |
| Comment submission | < 200ms | Optimistic render |
| Load more replies | < 300ms | Pagination fetch |

### Database Query Performance

| Query Type | Target | Implementation |
|------------|--------|----------------|
| Top-level comments | < 50ms | Indexed query with limit |
| Get comment tree | < 100ms | Recursive CTE, single query |
| Get replies | < 30ms | Direct children query |
| Vote update | < 20ms | Indexed UPDATE |
| Create comment | < 50ms | INSERT with trigger |

### Load Testing Results

The system has been tested with:

- ✅ **1000+ comments per clip**: Loads in < 2s
- ✅ **Deep nesting (10 levels)**: No performance degradation
- ✅ **Concurrent submissions**: 25 users creating comments simultaneously
- ✅ **High vote frequency**: 100 votes/second sustained

**Load Test Tool**: k6 load testing framework

```bash
# Run comment load test
make test-load-comments

# With authentication
AUTH_TOKEN=your_token k6 run backend/tests/load/scenarios/comments.js
```

## E2E Testing Procedures

### Functionality Testing Checklist

#### Basic Comment Operations

- [ ] **Create top-level comment**
  - Navigate to clip detail page
  - Click "Add Comment" button
  - Enter text in composer
  - Click "Submit"
  - Verify comment appears at top of list
  - Verify comment count incremented

- [ ] **Create nested reply (3+ levels)**
  - Click "Reply" on a comment
  - Enter text in reply composer
  - Submit reply
  - Verify reply appears under parent comment
  - Verify indentation shows nesting level
  - Repeat to create depth 3+ thread
  - Verify visual hierarchy is clear

- [ ] **Collapse/expand at all depths**
  - Expand a comment with replies (shows "▼ N replies")
  - Click collapse badge
  - Verify replies are hidden
  - Verify badge changes to "▶ N replies"
  - Click expand badge
  - Verify replies reappear
  - Test at depth 1, 2, 3, and 10

- [ ] **Vote on nested comments**
  - Click upvote on a nested comment (depth 2+)
  - Verify score increments immediately
  - Verify visual feedback (colored icon)
  - Click upvote again to remove vote
  - Verify score decrements
  - Click downvote
  - Verify score decrements
  - Refresh page and verify vote persists

- [ ] **Edit nested comment**
  - Create a comment (within 15 min)
  - Click "Edit" button
  - Modify text
  - Submit edit
  - Verify content updates
  - Verify "edited" indicator appears
  - Wait 15+ minutes
  - Verify edit button disappears (author only)

- [ ] **Delete nested comment**
  - Create a comment with replies
  - Click "Delete" button
  - Confirm deletion
  - Verify content changes to "[deleted]"
  - Verify replies remain visible
  - Verify thread structure preserved

- [ ] **Load more replies**
  - Find a comment with 50+ replies
  - Verify "Load N more replies" link appears
  - Click link
  - Verify next batch of replies loads
  - Verify scroll position maintained
  - Verify link updates or disappears when all loaded

#### Edge Cases

- [ ] **Maximum depth enforcement (depth 10)**
  - Navigate to a depth-9 comment
  - Click "Reply"
  - Submit reply
  - Verify reply appears at depth 10
  - Click "Reply" on depth-10 comment
  - Verify "View N more replies" link instead of reply button
  - Click link and verify navigation to comment thread page

- [ ] **Deleted parent with nested replies**
  - Create comment with replies
  - Delete parent comment
  - Verify parent shows "[deleted]"
  - Verify nested replies still visible
  - Verify visual hierarchy maintained
  - Verify "Reply" button still works on children

- [ ] **Removed parent with nested replies**
  - As admin/moderator, remove a comment with replies
  - Verify parent shows "[removed]"
  - Verify nested replies still visible
  - Verify removal reason visible to moderators

- [ ] **Very long comment content**
  - Create comment with 9,999 characters
  - Verify content wraps properly
  - Verify no horizontal scroll
  - Verify readability maintained
  - Try creating comment with 10,001 characters
  - Verify validation error

- [ ] **Special characters and markdown**
  - Test with special characters: `<script>alert('xss')</script>`
  - Verify HTML is escaped/sanitized
  - Test markdown: `**bold** *italic* [link](https://example.com)`
  - Verify markdown renders correctly
  - Test code blocks with syntax highlighting
  - Verify inline code rendering

- [ ] **No broken threads or orphaned comments**
  - Load a clip with 100+ comments
  - Verify all top-level comments appear
  - Expand all threads
  - Verify no missing comments
  - Verify reply counts match actual replies
  - Verify no duplicate comments

### Cross-Platform Testing

#### Web and Mobile Parity

- [ ] **Display threads identically**
  - Open same clip on web browser and mobile app
  - Verify same comments appear in same order
  - Verify same nesting structure
  - Verify same vote scores
  - Verify same collapse state defaults

- [ ] **Collapse state behavior**
  - Collapse a thread on web
  - Note: State does NOT sync (client-side only)
  - Reload page and verify state resets
  - Collapse thread on mobile
  - Verify independent from web state

- [ ] **Reply appears immediately**
  - Submit reply on web
  - Verify immediate optimistic render
  - Verify actual comment appears after backend confirmation
  - Repeat on mobile
  - Verify same behavior

#### Browser Testing Matrix

Test on:
- [ ] Chrome (desktop, latest version)
- [ ] Firefox (desktop, latest version)
- [ ] Safari (desktop, latest version)
- [ ] Edge (desktop, latest version)

#### Mobile Testing Matrix

Test on:
- [ ] iOS Safari (iPhone, latest iOS)
- [ ] Android Chrome (Pixel/Samsung, latest Android)
- [ ] iOS Safari (iPad, latest iOS)

### Performance Testing

#### Load Time Tests

- [ ] **Load clip with 1000 comments**
  - Use load test data or seed script
  - Open clip in browser
  - Measure time to first meaningful paint
  - Target: < 2 seconds
  - Use Chrome DevTools Performance tab
  - Verify no layout shifts
  - Verify images lazy load

- [ ] **Expand deep thread**
  - Find collapsed thread with 50+ replies
  - Click expand badge
  - Measure time until all replies visible
  - Target: < 500ms
  - Verify smooth animation
  - Verify no jank

- [ ] **Scroll 1000 nested comments**
  - Load clip with 1000+ comments
  - Scroll from top to bottom continuously
  - Monitor FPS in Chrome DevTools
  - Target: 60 FPS sustained
  - Check for dropped frames
  - Verify smooth scrolling experience

- [ ] **Vote optimistic update**
  - Click upvote on any comment
  - Measure time to UI feedback
  - Target: < 100ms
  - Verify instant visual feedback
  - Verify score updates immediately
  - Verify backend sync happens after

#### Network Conditions Testing

Test with Chrome DevTools Network throttling:

- [ ] **Slow 3G**
  - Load comments
  - Verify loading states appear
  - Verify skeleton screens show
  - Verify progressive loading
  - Verify error handling for timeouts

- [ ] **Offline**
  - Disable network
  - Try loading comments
  - Verify appropriate error message
  - Verify retry mechanism
  - Re-enable network and verify recovery

#### CPU Throttling Testing

Test with Chrome DevTools Performance tab (6x slowdown):

- [ ] **Slow device simulation**
  - Enable 6x CPU throttling
  - Load comments
  - Verify still loads < 5s
  - Scroll comments
  - Verify still smooth (30+ FPS)
  - Collapse/expand threads
  - Verify still responsive

#### Load Testing (Backend)

Run k6 load tests to verify backend performance:

```bash
# Test with 25 concurrent users
make test-load-comments

# Test with authentication (includes write operations)
AUTH_TOKEN=$(cat .env | grep JWT_TOKEN | cut -d= -f2) k6 run backend/tests/load/scenarios/comments.js
```

**Targets:**

- [ ] List comments: p(95) < 50ms, p(99) < 100ms
- [ ] Create comment: p(95) < 100ms, p(99) < 200ms
- [ ] Vote comment: p(95) < 50ms, p(99) < 100ms
- [ ] Error rate: < 2%
- [ ] No failed requests for list operations

## API Examples

### Example 1: Fetch Top-Level Comments

**Request:**

```bash
curl -X GET "https://api.clpr.tv/v1/clips/123e4567-e89b-12d3-a456-426614174000/comments?sort=best&limit=20" \
  -H "Accept: application/json"
```

**Response:**

```json
{
  "comments": [
    {
      "id": "comment-uuid-1",
      "clip_id": "123e4567-e89b-12d3-a456-426614174000",
      "user_id": "user-uuid",
      "parent_comment_id": null,
      "content": "Amazing play!",
      "rendered_content": "<p>Amazing play!</p>",
      "vote_score": 42,
      "reply_count": 5,
      "is_edited": false,
      "is_removed": false,
      "created_at": "2024-12-01T12:00:00Z",
      "updated_at": "2024-12-01T12:00:00Z",
      "author_username": "gamer123",
      "author_display_name": "Gamer 123",
      "author_avatar_url": "https://cdn.clpr.tv/avatars/...",
      "author_karma": 1234,
      "author_role": "user",
      "user_vote": 1,
      "replies": []
    }
  ],
  "next_cursor": 20,
  "has_more": true
}
```

### Example 2: Create a Top-Level Comment

**Request:**

```bash
curl -X POST "https://api.clpr.tv/v1/clips/123e4567-e89b-12d3-a456-426614174000/comments" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -d '{
    "content": "This is **bold** and *italic* text with a [link](https://example.com)"
  }'
```

**Response (201 Created):**

```json
{
  "id": "new-comment-uuid",
  "clip_id": "123e4567-e89b-12d3-a456-426614174000",
  "user_id": "your-user-uuid",
  "parent_comment_id": null,
  "content": "This is **bold** and *italic* text with a [link](https://example.com)",
  "vote_score": 0,
  "is_edited": false,
  "is_removed": false,
  "created_at": "2024-12-15T10:30:00Z",
  "updated_at": "2024-12-15T10:30:00Z"
}
```

### Example 3: Create a Nested Reply

**Request:**

```bash
curl -X POST "https://api.clpr.tv/v1/clips/123e4567-e89b-12d3-a456-426614174000/comments" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -d '{
    "content": "Great point! I agree completely.",
    "parent_comment_id": "parent-comment-uuid"
  }'
```

### Example 4: Get Replies to a Comment

**Request:**

```bash
curl -X GET "https://api.clpr.tv/v1/comments/parent-comment-uuid/replies?limit=10&cursor=0" \
  -H "Accept: application/json" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

**Response:**

```json
{
  "comments": [
    {
      "id": "reply-uuid-1",
      "parent_comment_id": "parent-comment-uuid",
      "content": "I agree!",
      "rendered_content": "<p>I agree!</p>",
      "vote_score": 5,
      "reply_count": 2,
      "user_vote": 0,
      "replies": []
    }
  ],
  "next_cursor": 10,
  "has_more": false
}
```

### Example 5: Vote on a Comment

**Request:**

```bash
# Upvote
curl -X POST "https://api.clpr.tv/v1/comments/comment-uuid/vote" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -d '{
    "vote": 1
  }'

# Downvote
curl -X POST "https://api.clpr.tv/v1/comments/comment-uuid/vote" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -d '{
    "vote": -1
  }'

# Remove vote
curl -X POST "https://api.clpr.tv/v1/comments/comment-uuid/vote" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -d '{
    "vote": 0
  }'
```

**Response (200 OK):**

```json
{
  "message": "Vote recorded successfully"
}
```

### Example 6: Edit a Comment

**Request:**

```bash
curl -X PUT "https://api.clpr.tv/v1/comments/comment-uuid" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -d '{
    "content": "Updated content with corrections"
  }'
```

**Response (200 OK):**

```json
{
  "message": "Comment updated successfully"
}
```

### Example 7: Delete a Comment

**Request:**

```bash
curl -X DELETE "https://api.clpr.tv/v1/comments/comment-uuid" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

**Response (200 OK):**

```json
{
  "message": "Comment deleted successfully"
}
```

### TypeScript/JavaScript Examples

Using axios:

```typescript
import axios from 'axios';

const API_BASE = 'https://api.clpr.tv/v1';
const token = 'YOUR_JWT_TOKEN';

// List comments
async function getComments(clipId: string, sort = 'best', limit = 20) {
  const response = await axios.get(
    `${API_BASE}/clips/${clipId}/comments`,
    {
      params: { sort, limit },
      headers: { Authorization: `Bearer ${token}` }
    }
  );
  return response.data;
}

// Create comment
async function createComment(clipId: string, content: string, parentId?: string) {
  const response = await axios.post(
    `${API_BASE}/clips/${clipId}/comments`,
    { content, parent_comment_id: parentId },
    { headers: { Authorization: `Bearer ${token}` } }
  );
  return response.data;
}

// Vote on comment
async function voteComment(commentId: string, vote: 1 | -1 | 0) {
  const response = await axios.post(
    `${API_BASE}/comments/${commentId}/vote`,
    { vote },
    { headers: { Authorization: `Bearer ${token}` } }
  );
  return response.data;
}

// Get replies
async function getReplies(commentId: string, limit = 10, cursor = 0) {
  const response = await axios.get(
    `${API_BASE}/comments/${commentId}/replies`,
    {
      params: { limit, cursor },
      headers: { Authorization: `Bearer ${token}` }
    }
  );
  return response.data;
}
```

## Migration Guide

### For Existing Comment Data

If you have existing flat comment data that needs to be migrated to the nested structure:

#### 1. Database Migration

No migration needed if:
- Comments table already has `parent_comment_id` column
- Existing comments have `parent_comment_id = NULL` (treated as top-level)

Migration needed if:
- Comments table lacks `parent_comment_id` column
- Need to add `reply_count` denormalized field
- Need to add indexes for performance

**Migration SQL:**

```sql
-- Add parent_comment_id if missing
ALTER TABLE comments 
ADD COLUMN IF NOT EXISTS parent_comment_id UUID REFERENCES comments(id);

-- Add reply_count if missing
ALTER TABLE comments 
ADD COLUMN IF NOT EXISTS reply_count INT DEFAULT 0;

-- Create indexes
CREATE INDEX IF NOT EXISTS idx_comments_parent 
ON comments(parent_comment_id) 
WHERE parent_comment_id IS NOT NULL;

CREATE INDEX IF NOT EXISTS idx_comments_clip 
ON comments(clip_id, created_at);

-- Backfill reply_count
UPDATE comments c
SET reply_count = (
    SELECT COUNT(*) 
    FROM comments r 
    WHERE r.parent_comment_id = c.id
);

-- Create trigger for automatic reply_count maintenance
CREATE OR REPLACE FUNCTION update_reply_count()
RETURNS TRIGGER AS $$
BEGIN
    IF TG_OP = 'INSERT' THEN
        IF NEW.parent_comment_id IS NOT NULL THEN
            UPDATE comments 
            SET reply_count = reply_count + 1 
            WHERE id = NEW.parent_comment_id;
        END IF;
    ELSIF TG_OP = 'DELETE' THEN
        IF OLD.parent_comment_id IS NOT NULL THEN
            UPDATE comments 
            SET reply_count = reply_count - 1 
            WHERE id = OLD.parent_comment_id;
        END IF;
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_update_reply_count
AFTER INSERT OR DELETE ON comments
FOR EACH ROW
EXECUTE FUNCTION update_reply_count();
```

#### 2. Frontend Migration

**Before (flat comments):**

```typescript
// Old: Flat list
comments.map(comment => (
  <CommentItem key={comment.id} comment={comment} />
))
```

**After (nested comments):**

```typescript
// New: Tree structure with CommentTree component
<CommentTree 
  comments={comments} 
  clipId={clipId}
  currentUserId={currentUserId}
  maxDepth={10}
/>
```

**Component Structure:**

```
CommentSection (main container)
└─ CommentTree (renders tree recursively)
   └─ CommentItem (individual comment)
      ├─ CommentActions (vote, reply, edit, delete)
      ├─ CommentForm (reply composer)
      └─ CommentTree (nested replies, recursive)
```

#### 3. Mobile Migration

Similar changes needed in React Native app:

```typescript
// Import nested comment components
import { CommentList } from '@/components/CommentList';

// Render with nesting support
<CommentList 
  clipId={clipId}
  comments={comments}
  maxDepth={10}
/>
```

#### 4. API Changes

If upgrading from a previous comment API:

**Old API (flat):**
```
GET /api/v1/clips/{clipId}/comments
```

**New API (nested, backward compatible):**
```
GET /api/v1/clips/{clipId}/comments
  → Returns top-level comments with reply_count
GET /api/v1/comments/{commentId}/replies
  → Returns nested replies (new endpoint)
```

The API is backward compatible. Old clients will see `reply_count` in responses but can ignore it.

## Troubleshooting

### Common Issues

#### Comments Not Loading

**Symptoms:**
- Empty comment section
- Infinite loading spinner
- Error message displayed

**Diagnosis:**

1. Check browser console for API errors
2. Check network tab for failed requests
3. Verify authentication token is valid
4. Check API endpoint URL is correct

**Solutions:**

```bash
# Verify backend is running
curl http://localhost:8080/health

# Test comment endpoint
curl http://localhost:8080/api/v1/clips/{clipId}/comments

# Check authentication
curl -H "Authorization: Bearer YOUR_TOKEN" \
  http://localhost:8080/api/v1/auth/me
```

#### Nested Replies Not Expanding

**Symptoms:**
- Click expand badge but nothing happens
- Console shows "Failed to load replies"

**Diagnosis:**

1. Check if comment has `reply_count > 0`
2. Verify API endpoint `/comments/{id}/replies` is working
3. Check for JavaScript errors in console

**Solutions:**

- Verify backend route is registered
- Check CORS configuration for API requests
- Ensure recursive query is working in database

#### Maximum Depth Not Enforced

**Symptoms:**
- Users can reply beyond depth 10
- No "View more replies" link appears

**Diagnosis:**

1. Check service layer depth validation
2. Verify frontend is calculating depth correctly
3. Check database query for parent chain

**Solutions:**

```go
// Backend: Ensure depth check exists
depth, err := s.repo.GetNestingDepth(ctx, parentID)
if depth >= MaxNestingDepth {
    return nil, fmt.Errorf("maximum nesting depth reached")
}
```

```typescript
// Frontend: Calculate depth from parent chain
function calculateDepth(comment: Comment, allComments: Comment[]): number {
  let depth = 0;
  let current = comment;
  
  while (current.parent_comment_id) {
    depth++;
    current = allComments.find(c => c.id === current.parent_comment_id);
    if (!current) break;
  }
  
  return depth;
}
```

#### Performance Issues

**Symptoms:**
- Slow page load with many comments
- Laggy scrolling
- Browser becomes unresponsive

**Diagnosis:**

1. Check number of comments being rendered
2. Monitor memory usage in browser
3. Check for unnecessary re-renders
4. Profile with Chrome DevTools

**Solutions:**

- Implement virtualization for long comment lists
- Use pagination (limit to 50-100 comments per load)
- Add lazy loading for nested replies
- Optimize React component memoization
- Enable database query caching

```typescript
// Use React.memo to prevent unnecessary re-renders
const CommentItem = React.memo(({ comment, depth }) => {
  // Component code
}, (prevProps, nextProps) => {
  // Custom comparison function
  return prevProps.comment.id === nextProps.comment.id &&
         prevProps.comment.vote_score === nextProps.comment.vote_score;
});
```

#### Markdown Not Rendering

**Symptoms:**
- Raw markdown visible instead of formatted text
- Links not clickable
- Bold/italic not applied

**Diagnosis:**

1. Check if `rendered_content` field is present in API response
2. Verify markdown parser is configured
3. Check sanitizer is not over-stripping

**Solutions:**

Backend should render markdown:

```go
// Ensure markdown rendering is called
renderedContent := s.RenderMarkdown(comment.Content)
```

Frontend should use rendered HTML:

```typescript
// Use dangerouslySetInnerHTML with sanitized content
<div dangerouslySetInnerHTML={{ __html: comment.rendered_content }} />
```

### Performance Optimization Tips

#### Database Optimization

1. **Add Indexes:**

```sql
-- Essential indexes
CREATE INDEX idx_comments_clip_score ON comments(clip_id, vote_score DESC);
CREATE INDEX idx_comments_parent_score ON comments(parent_comment_id, vote_score DESC);
```

2. **Use Connection Pooling:**

```go
// Configure pgx pool
config.MaxConns = 25
config.MinConns = 5
config.MaxConnLifetime = time.Hour
```

3. **Enable Query Caching:**

```go
// Cache hot comment trees in Redis
cache.Set(ctx, fmt.Sprintf("comments:tree:%s", commentID), data, 10*time.Minute)
```

#### Frontend Optimization

1. **Lazy Load Replies:**

```typescript
// Only load when user expands
const loadReplies = async (commentId: string) => {
  if (!repliesCache[commentId]) {
    const replies = await fetchReplies(commentId);
    setRepliesCache(prev => ({ ...prev, [commentId]: replies }));
  }
};
```

2. **Virtualize Long Lists:**

```typescript
import { useVirtualizer } from '@tanstack/react-virtual';

const virtualizer = useVirtualizer({
  count: comments.length,
  getScrollElement: () => scrollRef.current,
  estimateSize: () => 100,
});
```

3. **Optimize Re-renders:**

```typescript
// Use React Query for caching
const { data: comments } = useQuery({
  queryKey: ['comments', clipId],
  queryFn: () => fetchComments(clipId),
  staleTime: 5 * 60 * 1000, // 5 minutes
});
```

## Related Documentation

- [Comment API Reference](../backend/comment-api.md) - Complete API documentation
- [Comment Repository Methods](../../backend/internal/repository/COMMENT_REPOSITORY.md) - Database layer documentation
- [Backend Architecture](../backend/architecture.md) - Overall system architecture
- [Frontend Testing Guide](../frontend/testing.md) - Component testing strategies
- [Load Testing Guide](../backend/testing-performance.md) - Performance testing with k6

## Support

For issues or questions:

- **GitHub Issues**: [subculture-collective/clpr/issues](https://git.subcult.tv/subculture-collective/clpr/issues)
- **Documentation**: [Full docs](../index.md)
- **API Status**: [status.clpr.tv](https://status.clpr.tv)

## Version History

- **v1.0** (2024-12-15): Initial nested comment threading release
  - Reddit-style threading with 10-level depth
  - Collapse/expand functionality
  - Markdown support with XSS protection
  - Load testing validated for 1000+ comments
  - Cross-platform parity (web + mobile)
