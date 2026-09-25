---
title: "API Playlist Sharing"
summary: "This document describes the API endpoints for playlist sharing and collaborative features."
tags: ["backend","api"]
area: "backend"
status: "stable"
owner: "team-core"
version: "1.0"
last_reviewed: 2026-01-29
---

# Playlist Sharing & Collaboration API

This document describes the API endpoints for playlist sharing and collaborative features.

## Overview

The playlist sharing system allows users to:
- Share playlists with public/private/unlisted visibility
- Generate share links with embed codes
- Add collaborators with different permission levels
- Track share analytics

## Authentication

Most endpoints require authentication via JWT token in the `Authorization` header:
```
Authorization: Bearer <jwt_token>
```

Some endpoints (marked as "Public") do not require authentication.

## Endpoints

### 1. Get Share Link

Generate or retrieve a share link for a playlist.

**Endpoint:** `GET /api/v1/playlists/:id/share-link`

**Auth:** Required

**Rate Limit:** 10 requests per hour per user

**Response:**
```json
{
  "success": true,
  "data": {
    "share_url": "https://clpr.tv/playlists/abc123token",
    "embed_code": "<iframe src=\"https://clpr.tv/embed/playlist/abc123token\" width=\"800\" height=\"600\" frameborder=\"0\" allowfullscreen></iframe>"
  }
}
```

**Permissions:**
- Playlist owner
- Collaborators with edit or admin permission

---

### 2. Track Share Event

Track when a playlist is shared (for analytics).

**Endpoint:** `POST /api/v1/playlists/:id/track-share`

**Auth:** Public (no authentication required)

**Request Body:**
```json
{
  "platform": "twitter",
  "referrer": "https://example.com/playlist-page"
}
```

**Fields:**
- `platform` (required): One of: `twitter`, `facebook`, `discord`, `embed`, `link`
- `referrer` (optional): Source URL where the share originated

**Response:**
```json
{
  "success": true,
  "data": {
    "message": "Share tracked successfully"
  }
}
```

---

### 3. List Collaborators

Get all collaborators for a playlist.

**Endpoint:** `GET /api/v1/playlists/:id/collaborators`

**Auth:** Optional (required for private playlists)

**Response:**
```json
{
  "success": true,
  "data": [
    {
      "id": "uuid",
      "playlist_id": "uuid",
      "user_id": "uuid",
      "user": {
        "id": "uuid",
        "username": "johndoe",
        "display_name": "John Doe",
        "avatar_url": "https://..."
      },
      "permission": "edit",
      "invited_by": "uuid",
      "invited_at": "2025-12-18T00:00:00Z",
      "created_at": "2025-12-18T00:00:00Z",
      "updated_at": "2025-12-18T00:00:00Z"
    }
  ]
}
```

**Permissions:**
- Public/unlisted playlists: Anyone can view
- Private playlists: Owner and collaborators only

---

### 4. Add Collaborator

Add a collaborator to a playlist.

**Endpoint:** `POST /api/v1/playlists/:id/collaborators`

**Auth:** Required

**Rate Limit:** 20 requests per hour per user

**Request Body:**
```json
{
  "user_id": "uuid",
  "permission": "edit"
}
```

**Fields:**
- `user_id` (required): UUID of the user to add
- `permission` (required): One of: `view`, `edit`, `admin`

**Permission Levels:**
- `view`: Can view private playlists
- `edit`: Can modify playlist content (add/remove/reorder clips)
- `admin`: Can manage collaborators

**Response:**
```json
{
  "success": true,
  "data": {
    "message": "Collaborator added successfully"
  }
}
```

**Permissions:**
- Playlist owner
- Collaborators with admin permission

---

### 5. Remove Collaborator

Remove a collaborator from a playlist.

**Endpoint:** `DELETE /api/v1/playlists/:id/collaborators/:user_id`

**Auth:** Required

**Response:**
```json
{
  "success": true,
  "data": {
    "message": "Collaborator removed successfully"
  }
}
```

**Permissions:**
- Playlist owner
- Collaborators with admin permission

---

### 6. Update Collaborator Permission

Update a collaborator's permission level.

**Endpoint:** `PATCH /api/v1/playlists/:id/collaborators/:user_id`

**Auth:** Required

**Request Body:**
```json
{
  "permission": "admin"
}
```

**Fields:**
- `permission` (required): One of: `view`, `edit`, `admin`

**Response:**
```json
{
  "success": true,
  "data": {
    "message": "Collaborator permission updated successfully"
  }
}
```

**Permissions:**
- Playlist owner
- Collaborators with admin permission

---

### 7. Update Playlist Visibility

Update a playlist's visibility and automatically generate share token if needed.

**Endpoint:** `PATCH /api/v1/playlists/:id`

**Auth:** Required

**Request Body:**
```json
{
  "visibility": "public"
}
```

**Fields:**
- `visibility`: One of: `private`, `public`, `unlisted`

**Notes:**
- Share token is automatically generated when changing from private to public/unlisted
- Only the playlist owner can change visibility
- Collaborators with edit permission can modify other fields but not visibility

**Response:**
```json
{
  "success": true,
  "data": {
    "id": "uuid",
    "title": "My Playlist",
    "visibility": "public",
    "share_token": "abc123token",
    ...
  }
}
```

---

## Permission Matrix

| Action | Owner | Admin Collaborator | Edit Collaborator | View Collaborator | Public |
|--------|-------|-------------------|-------------------|-------------------|--------|
| View private playlist | ✅ | ✅ | ✅ | ✅ | ❌ |
| View public/unlisted | ✅ | ✅ | ✅ | ✅ | ✅ |
| Edit content | ✅ | ✅ | ✅ | ❌ | ❌ |
| Change visibility | ✅ | ❌ | ❌ | ❌ | ❌ |
| Get share link | ✅ | ✅ | ✅ | ❌ | ❌ |
| Add collaborators | ✅ | ✅ | ❌ | ❌ | ❌ |
| Remove collaborators | ✅ | ✅ | ❌ | ❌ | ❌ |
| Update permissions | ✅ | ✅ | ❌ | ❌ | ❌ |
| View collaborators (public) | ✅ | ✅ | ✅ | ✅ | ✅ |
| View collaborators (private) | ✅ | ✅ | ✅ | ✅ | ❌ |
| Track shares | ✅ | ✅ | ✅ | ✅ | ✅ |

---

## Error Responses

All endpoints follow a consistent error response format:

```json
{
  "success": false,
  "error": {
    "code": "ERROR_CODE",
    "message": "Human-readable error message"
  }
}
```

Common error codes:
- `UNAUTHORIZED`: Missing or invalid authentication
- `FORBIDDEN`: User lacks permission for this action
- `NOT_FOUND`: Playlist or collaborator not found
- `INVALID_REQUEST`: Invalid request parameters
- `INTERNAL_ERROR`: Server error

---

## Database Schema

### playlists table additions

```sql
share_token VARCHAR(100) UNIQUE
view_count INT DEFAULT 0
share_count INT DEFAULT 0
```

### playlist_collaborators table

```sql
CREATE TABLE playlist_collaborators (
    id UUID PRIMARY KEY,
    playlist_id UUID NOT NULL REFERENCES playlists(id),
    user_id UUID NOT NULL REFERENCES users(id),
    permission VARCHAR(20) CHECK (permission IN ('view', 'edit', 'admin')),
    invited_by UUID REFERENCES users(id),
    invited_at TIMESTAMP DEFAULT NOW(),
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    UNIQUE(playlist_id, user_id)
);
```

### playlist_shares table

```sql
CREATE TABLE playlist_shares (
    id UUID PRIMARY KEY,
    playlist_id UUID NOT NULL REFERENCES playlists(id),
    platform VARCHAR(50),
    referrer VARCHAR(255),
    shared_at TIMESTAMP DEFAULT NOW()
);
```

---

## Rate Limiting

Rate limits are enforced on specific endpoints:

| Endpoint | Limit | Window |
|----------|-------|--------|
| GET /share-link | 10 requests | 1 hour |
| POST /collaborators | 20 requests | 1 hour |
| POST /clips (existing) | 60 requests | 1 minute |
| POST /like (existing) | 30 requests | 1 minute |

When rate limited, the API returns:
```json
{
  "success": false,
  "error": {
    "code": "RATE_LIMIT_EXCEEDED",
    "message": "Rate limit exceeded. Please try again later."
  }
}
```

---

## Frontend Integration

### Using ShareModal Component

```tsx
import { ShareModal } from '@/components/playlist';

function PlaylistPage() {
  const [showShareModal, setShowShareModal] = useState(false);
  const playlistId = "...";

  return (
    <>
      <button onClick={() => setShowShareModal(true)}>
        Share Playlist
      </button>
      
      {showShareModal && (
        <ShareModal
          playlistId={playlistId}
          onClose={() => setShowShareModal(false)}
        />
      )}
    </>
  );
}
```

### Using CollaboratorManager Component

```tsx
import { CollaboratorManager } from '@/components/playlist';

function PlaylistSettings() {
  const playlistId = "...";
  const isOwner = true; // Check if current user owns the playlist
  const canManage = true; // Check if user has admin permission

  return (
    <CollaboratorManager
      playlistId={playlistId}
      isOwner={isOwner}
      canManageCollaborators={canManage}
    />
  );
}
```

### Using PlaylistCard with Share

```tsx
import { PlaylistCard } from '@/components/playlist';

function PlaylistList() {
  const [showShareModal, setShowShareModal] = useState(false);
  const [selectedPlaylistId, setSelectedPlaylistId] = useState<string | null>(null);

  const handleShare = (playlistId: string) => {
    setSelectedPlaylistId(playlistId);
    setShowShareModal(true);
  };

  return (
    <>
      {playlists.map(playlist => (
        <PlaylistCard
          key={playlist.id}
          playlist={playlist}
          onShare={handleShare}
        />
      ))}
      
      {showShareModal && selectedPlaylistId && (
        <ShareModal
          playlistId={selectedPlaylistId}
          onClose={() => {
            setShowShareModal(false);
            setSelectedPlaylistId(null);
          }}
        />
      )}
    </>
  );
}
```

---

## Analytics Queries

### Get most shared playlists

```sql
SELECT p.id, p.title, p.share_count
FROM playlists p
WHERE p.visibility IN ('public', 'unlisted')
ORDER BY p.share_count DESC
LIMIT 10;
```

### Get share breakdown by platform

```sql
SELECT platform, COUNT(*) as share_count
FROM playlist_shares
WHERE playlist_id = $1
GROUP BY platform
ORDER BY share_count DESC;
```

### Get most viewed playlists

```sql
SELECT p.id, p.title, p.view_count
FROM playlists p
WHERE p.visibility = 'public'
ORDER BY p.view_count DESC
LIMIT 10;
```

### Get playlists with most collaborators

```sql
SELECT p.id, p.title, COUNT(pc.id) as collaborator_count
FROM playlists p
JOIN playlist_collaborators pc ON p.id = pc.playlist_id
GROUP BY p.id, p.title
ORDER BY collaborator_count DESC
LIMIT 10;
```
