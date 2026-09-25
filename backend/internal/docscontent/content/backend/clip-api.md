---
title: 'Clip Management API Documentation'
summary: 'The Clip Management API provides comprehensive endpoints for managing clips with CRUD operations, fi'
tags: ['backend', 'api']
area: 'backend'
status: 'stable'
owner: 'team-core'
version: '1.0'
last_reviewed: 2025-12-11
---

# Clip Management API Documentation

## Overview

The Clip Management API provides comprehensive endpoints for managing clips with CRUD operations, filtering, sorting, and pagination.

## Base URL

```
/api/v1
```

## Authentication

Some endpoints require authentication. Include the JWT token in the `Authorization` header:

```
Authorization: Bearer <token>
```

---

## Endpoints

### 1. List Clips

**GET** `/clips`

Retrieve a paginated list of clips with optional filtering and sorting.

#### Query Parameters

| Parameter          | Type    | Default | Description                                                                                         |
| ------------------ | ------- | ------- | --------------------------------------------------------------------------------------------------- |
| `sort`             | string  | `hot`   | Sorting method: `hot`, `new`, `top`, `rising`, `discussed`                                          |
| `timeframe`        | string  | -       | Time filter for `top` sort: `hour`, `day`, `week`, `month`, `year`, `all`                           |
| `game_id`          | string  | -       | Filter by game ID                                                                                   |
| `broadcaster_id`   | string  | -       | Filter by broadcaster ID                                                                            |
| `tag`              | string  | -       | Filter by tag slug                                                                                  |
| `search`           | string  | -       | Full-text search in title                                                                           |
| `show_all_clips`   | boolean | `false` | If `true`, includes both user-submitted and scraped clips. Default only shows user-submitted clips. |
| `top10k_streamers` | boolean | `false` | If `true`, only shows clips from top 10k streamers                                                  |
| `page`             | integer | `1`     | Page number (min: 1)                                                                                |
| `limit`            | integer | `25`    | Results per page (min: 1, max: 100)                                                                 |

#### Response

```json
{
    "success": true,
    "data": [
        {
            "id": "uuid",
            "twitch_clip_id": "string",
            "title": "string",
            "broadcaster_name": "string",
            "game_name": "string",
            "view_count": 1234,
            "vote_score": 42,
            "comment_count": 10,
            "favorite_count": 5,
            "created_at": "2024-01-01T00:00:00Z",
            "user_vote": 1,
            "is_favorited": true,
            "upvote_count": 45,
            "downvote_count": 3
        }
    ],
    "meta": {
        "page": 1,
        "limit": 25,
        "total": 150,
        "total_pages": 6,
        "has_next": true,
        "has_prev": false
    }
}
```

#### Sorting Algorithms

- **hot**: Wilson score + time decay (default)
- **new**: Most recent clips (created_at DESC)
- **top**: Highest vote_score (requires timeframe)
- **rising**: Recent clips with high velocity (48 hours, views + votes)
- **discussed**: Clips with most comments (requires timeframe)

#### Content Filtering

By default, the `/clips` endpoint returns **only user-submitted content** (clips where `submitted_by_user_id IS NOT NULL`). This ensures that the main feed pages (home, hot, new, top, rising, discussed) show only community-curated content.

To include all clips (both user-submitted and scraped clips), set `show_all_clips=true`. This is used on the Discovery page to show all available content.

#### Examples

```bash
# Get hot clips (user-submitted only, default)
GET /clips?sort=hot

# Get all clips including scraped content (for discovery)
GET /clips?sort=hot&show_all_clips=true

# Get top clips from the last week
GET /clips?sort=top&timeframe=week

# Get clips for a specific game
GET /clips?game_id=123456

# Search clips
GET /clips?search=epic

# Get new clips with pagination
GET /clips?sort=new&page=2&limit=50
```

---

### 2. Get Single Clip

**GET** `/clips/:id`

Retrieve details for a specific clip.

#### Path Parameters

| Parameter | Type | Required | Description |
| --------- | ---- | -------- | ----------- |
| `id`      | UUID | Yes      | Clip ID     |

#### Response

```json
{
    "success": true,
    "data": {
        "id": "uuid",
        "twitch_clip_id": "string",
        "twitch_clip_url": "string",
        "embed_url": "string",
        "title": "string",
        "creator_name": "string",
        "broadcaster_name": "string",
        "game_name": "string",
        "thumbnail_url": "string",
        "duration": 30.5,
        "view_count": 1234,
        "vote_score": 42,
        "comment_count": 10,
        "favorite_count": 5,
        "is_featured": false,
        "is_nsfw": false,
        "created_at": "2024-01-01T00:00:00Z",
        "user_vote": 1,
        "is_favorited": true,
        "upvote_count": 45,
        "downvote_count": 3
    }
}
```

#### Notes

- View count is incremented asynchronously
- Returns 404 if clip not found or removed
- User-specific data (`user_vote`, `is_favorited`) only included if authenticated

---

### 3. Vote on Clip

**POST** `/clips/:id/vote`

Vote on a clip (upvote, downvote, or remove vote).

**Authentication Required** | **Rate Limited** (20 per minute)

#### Path Parameters

| Parameter | Type | Required | Description |
| --------- | ---- | -------- | ----------- |
| `id`      | UUID | Yes      | Clip ID     |

#### Request Body

```json
{
    "vote": 1
}
```

| Field  | Type    | Required | Description                                             |
| ------ | ------- | -------- | ------------------------------------------------------- |
| `vote` | integer | Yes      | Vote value: `1` (upvote), `-1` (downvote), `0` (remove) |

#### Response

```json
{
    "success": true,
    "data": {
        "message": "Vote processed successfully",
        "vote_score": 43,
        "upvote_count": 46,
        "downvote_count": 3,
        "user_vote": 1
    }
}
```

#### Notes

- Voting updates user karma asynchronously
- Vote changes are upserted (previous vote is replaced)
- Triggers update clip vote_score automatically

---

### 4. Add to Favorites

**POST** `/clips/:id/favorite`

Add a clip to user's favorites.

**Authentication Required**

#### Path Parameters

| Parameter | Type | Required | Description |
| --------- | ---- | -------- | ----------- |
| `id`      | UUID | Yes      | Clip ID     |

#### Response

```json
{
    "success": true,
    "data": {
        "message": "Clip added to favorites",
        "is_favorited": true
    }
}
```

#### Notes

- Idempotent operation (returns 200 if already favorited)
- Triggers increment clip favorite_count automatically

---

### 5. Remove from Favorites

**DELETE** `/clips/:id/favorite`

Remove a clip from user's favorites.

**Authentication Required**

#### Path Parameters

| Parameter | Type | Required | Description |
| --------- | ---- | -------- | ----------- |
| `id`      | UUID | Yes      | Clip ID     |

#### Response

```json
{
    "success": true,
    "data": {
        "message": "Clip removed from favorites",
        "is_favorited": false
    }
}
```

#### Notes

- Idempotent operation (returns 200 even if not favorited)
- Triggers decrement clip favorite_count automatically

---

### 6. Get Related Clips

**GET** `/clips/:id/related`

Get clips related to the specified clip.

#### Path Parameters

| Parameter | Type | Required | Description |
| --------- | ---- | -------- | ----------- |
| `id`      | UUID | Yes      | Clip ID     |

#### Response

```json
{
    "success": true,
    "data": [
        {
            "id": "uuid",
            "title": "string",
            "broadcaster_name": "string",
            "game_name": "string",
            "vote_score": 42,
            "created_at": "2024-01-01T00:00:00Z"
        }
    ]
}
```

#### Relevance Algorithm

Clips are ranked by:

1. Same game (priority 3)
2. Same broadcaster (priority 2)
3. Similar tags (priority 1 per matching tag)
4. Vote score as tiebreaker

#### Notes

- Excludes the current clip
- Limit to 10 results
- Returns empty array if no related clips found

---

### 7. Update Clip (Admin)

**PUT** `/clips/:id`

Update clip properties.

**Authentication Required** | **Admin/Moderator Only**

#### Path Parameters

| Parameter | Type | Required | Description |
| --------- | ---- | -------- | ----------- |
| `id`      | UUID | Yes      | Clip ID     |

#### Request Body

```json
{
    "is_featured": true,
    "is_nsfw": false,
    "is_removed": false,
    "removed_reason": null
}
```

#### Allowed Fields

- `is_featured` (boolean)
- `is_nsfw` (boolean)
- `is_removed` (boolean)
- `removed_reason` (string, nullable)

#### Response

```json
{
    "success": true,
    "data": {
        "id": "uuid",
        "is_featured": true,
        "is_nsfw": false,
        "is_removed": false
    }
}
```

#### Notes

- Only allowed fields can be updated
- Cache is invalidated after update
- Admin actions should be logged (TODO)

---

### 8. Delete Clip (Admin)

**DELETE** `/clips/:id`

Soft delete a clip (mark as removed).

**Authentication Required** | **Admin Only**

#### Path Parameters

| Parameter | Type | Required | Description |
| --------- | ---- | -------- | ----------- |
| `id`      | UUID | Yes      | Clip ID     |

#### Request Body

```json
{
    "reason": "Violates community guidelines"
}
```

| Field    | Type   | Required | Description         |
| -------- | ------ | -------- | ------------------- |
| `reason` | string | Yes      | Reason for deletion |

#### Response

```json
{
    "success": true,
    "data": {
        "message": "Clip deleted successfully"
    }
}
```

#### Notes

- Soft delete (sets `is_removed=true`)
- Data retained for audit trail
- Removed from public feeds
- Cache is invalidated after deletion

---

## Error Responses

All errors follow this format:

```json
{
    "success": false,
    "error": {
        "code": "ERROR_CODE",
        "message": "Human-readable error message"
    }
}
```

### Common Error Codes

| Code              | Status | Description               |
| ----------------- | ------ | ------------------------- |
| `INVALID_CLIP_ID` | 400    | Invalid clip ID format    |
| `INVALID_REQUEST` | 400    | Invalid request body      |
| `INVALID_VOTE`    | 400    | Vote must be -1, 0, or 1  |
| `UNAUTHORIZED`    | 401    | Authentication required   |
| `FORBIDDEN`       | 403    | Insufficient permissions  |
| `CLIP_NOT_FOUND`  | 404    | Clip not found or removed |
| `INTERNAL_ERROR`  | 500    | Server error              |

---

## Caching

Clip feeds are cached in Redis with the following TTLs:

| Feed Type | TTL        | Notes             |
| --------- | ---------- | ----------------- |
| Hot       | 5 minutes  | Most dynamic feed |
| New       | 2 minutes  | Frequent updates  |
| Top       | 15 minutes | More stable       |
| Rising    | 3 minutes  | Moderate updates  |

Cache is invalidated on:

- New votes
- Clip updates
- Clip deletions

---

## Rate Limits

| Endpoint    | Limit         |
| ----------- | ------------- |
| Vote        | 20 per minute |
| Comment     | 10 per minute |
| Submit Clip | 5 per hour    |
| List Clips  | No limit      |
| Get Clip    | No limit      |

---

## Database Performance

The API leverages several database optimizations:

### Indexes

- `idx_clips_vote_score` - For top sorting
- `idx_clips_created` - For new sorting
- `idx_clips_hot` - For hot sorting (composite)
- `idx_clips_game` - For game filtering
- `idx_clips_broadcaster` - For broadcaster filtering

### Database Functions

- `calculate_hot_score()` - Wilson score + time decay algorithm
- Triggers for automatic vote_score, comment_count, favorite_count updates

### Query Optimization

- Covering indexes for common queries
- Related clips use CTE for efficient relevance calculation
- Pagination uses LIMIT/OFFSET

---

## Examples

### JavaScript (Fetch API)

```javascript
// List hot clips
const response = await fetch('/api/v1/clips?sort=hot&limit=10');
const data = await response.json();
console.log(data.data); // Array of clips

// Vote on clip
const voteResponse = await fetch('/api/v1/clips/clip-id/vote', {
    method: 'POST',
    headers: {
        'Content-Type': 'application/json',
        Authorization: 'Bearer ' + token,
    },
    body: JSON.stringify({ vote: 1 }),
});
```

### cURL

```bash
# List clips
curl -X GET 'http://localhost:8080/api/v1/clips?sort=top&timeframe=week'

# Vote on clip
curl -X POST 'http://localhost:8080/api/v1/clips/{id}/vote' \
  -H 'Content-Type: application/json' \
  -H 'Authorization: Bearer YOUR_TOKEN' \
  -d '{"vote": 1}'

# Add to favorites
curl -X POST 'http://localhost:8080/api/v1/clips/{id}/favorite' \
  -H 'Authorization: Bearer YOUR_TOKEN'
```

---

## Changelog

### Version 1.0 (Current)

- ✅ Full CRUD operations for clips
- ✅ Advanced filtering and sorting
- ✅ Vote system with karma
- ✅ Favorites system
- ✅ Related clips algorithm
- ✅ Redis caching
- ✅ Admin controls
- ✅ Rate limiting
- ✅ Comprehensive error handling
