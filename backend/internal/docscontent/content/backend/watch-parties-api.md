---
title: "Watch Parties Api"
summary: "Watch Parties enable synchronized video watching with friends. Users can create parties, invite others via invite codes, and watch clips together with synchronized playback controls."
tags: ["backend","api"]
area: "backend"
status: "stable"
owner: "team-core"
version: "1.0"
last_reviewed: 2026-01-29
---

# Watch Parties API Documentation

## Overview

Watch Parties enable synchronized video watching with friends. Users can create parties, invite others via invite codes, and watch clips together with synchronized playback controls.

## Features

- **Real-time synchronization** via WebSocket
- **Host controls** for play, pause, seek, and skip
- **Participant management** with roles (host, co-host, viewer)
- **Invite system** with shareable codes
- **±2 second sync tolerance** for smooth playback
- **Up to 100 participants** per party (configurable to 1000)

## API Endpoints

### Create Watch Party

Creates a new watch party.

**Endpoint:** `POST /api/v1/watch-parties`

**Authentication:** Required

**Rate Limit:** 10 requests per hour

**Request Body:**
```json
{
  "title": "Friday Night Gaming",
  "playlist_id": "uuid-optional",
  "visibility": "private",
  "max_participants": 50
}
```

**Response:** `201 Created`
```json
{
  "success": true,
  "data": {
    "id": "party-uuid",
    "invite_code": "ABC123",
    "invite_url": "https://clpr.tv/watch-parties/ABC123",
    "party": {
      "id": "party-uuid",
      "host_user_id": "user-uuid",
      "title": "Friday Night Gaming",
      "visibility": "private",
      "max_participants": 50,
      "created_at": "2024-01-01T00:00:00Z"
    }
  }
}
```

---

### Join Watch Party

Join an existing watch party using an invite code.

**Endpoint:** `POST /api/v1/watch-parties/:code/join`

**Authentication:** Required

**Rate Limit:** 30 requests per hour

**Parameters:**
- `code` (path): The invite code (e.g., ABC123)

**Response:** `200 OK`
```json
{
  "success": true,
  "data": {
    "party": {
      "id": "party-uuid",
      "title": "Friday Night Gaming",
      "current_clip_id": "clip-uuid",
      "current_position_seconds": 42,
      "is_playing": true,
      "participants": [
        {
          "user_id": "host-uuid",
          "role": "host",
          "user": {
            "username": "host123",
            "display_name": "The Host"
          }
        }
      ]
    },
    "invite_url": "https://clpr.tv/watch-parties/ABC123"
  }
}
```

**Error Responses:**
- `404 Not Found`: Party not found or has ended
- `403 Forbidden`: Party is full

---

### Get Watch Party Details

Retrieve details about a watch party.

**Endpoint:** `GET /api/v1/watch-parties/:id`

**Authentication:** Optional (required for private parties)

**Parameters:**
- `id` (path): The party UUID

**Response:** `200 OK`
```json
{
  "success": true,
  "data": {
    "id": "party-uuid",
    "host_user_id": "user-uuid",
    "title": "Friday Night Gaming",
    "current_clip_id": "clip-uuid",
    "current_position_seconds": 42,
    "is_playing": true,
    "visibility": "private",
    "invite_code": "ABC123",
    "max_participants": 100,
    "participants": [...],
    "created_at": "2024-01-01T00:00:00Z",
    "started_at": "2024-01-01T00:05:00Z"
  }
}
```

---

### Get Participants

Get list of active participants in a watch party.

**Endpoint:** `GET /api/v1/watch-parties/:id/participants`

**Authentication:** Not required

**Parameters:**
- `id` (path): The party UUID

**Response:** `200 OK`
```json
{
  "success": true,
  "data": {
    "participants": [
      {
        "id": "participant-uuid",
        "user_id": "user-uuid",
        "role": "host",
        "joined_at": "2024-01-01T00:05:00Z",
        "user": {
          "username": "host123",
          "display_name": "The Host",
          "avatar_url": "https://..."
        }
      }
    ],
    "count": 1
  }
}
```

---

### Leave Watch Party

Leave a watch party you've joined.

**Endpoint:** `DELETE /api/v1/watch-parties/:id/leave`

**Authentication:** Required

**Parameters:**
- `id` (path): The party UUID

**Response:** `200 OK`
```json
{
  "success": true,
  "data": {
    "message": "Successfully left watch party"
  }
}
```

---

### End Watch Party

End a watch party (host only).

**Endpoint:** `POST /api/v1/watch-parties/:id/end`

**Authentication:** Required

**Parameters:**
- `id` (path): The party UUID

**Response:** `200 OK`
```json
{
  "success": true,
  "data": {
    "message": "Watch party ended successfully"
  }
}
```

**Error Responses:**
- `403 Forbidden`: Only the host can end the party

---

## WebSocket Protocol

### Connect to Party

**Endpoint:** `GET /api/v1/watch-parties/:id/ws`

**Authentication:** Required (via auth middleware)

**Protocol:** WebSocket

### Client → Server Messages

#### Play Command

```json
{
  "type": "play",
  "party_id": "party-uuid",
  "timestamp": 1704067200
}
```

#### Pause Command

```json
{
  "type": "pause",
  "party_id": "party-uuid",
  "timestamp": 1704067200
}
```

#### Seek Command

```json
{
  "type": "seek",
  "party_id": "party-uuid",
  "position": 120,
  "timestamp": 1704067200
}
```

#### Skip to Clip

```json
{
  "type": "skip",
  "party_id": "party-uuid",
  "clip_id": "clip-uuid",
  "timestamp": 1704067200
}
```

#### Request Sync

```json
{
  "type": "sync-request",
  "party_id": "party-uuid",
  "timestamp": 1704067200
}
```

### Server → Client Events

#### Sync Event

```json
{
  "type": "sync",
  "party_id": "party-uuid",
  "clip_id": "clip-uuid",
  "position": 120,
  "is_playing": true,
  "server_timestamp": 1704067200
}
```

#### Play Event

```json
{
  "type": "play",
  "party_id": "party-uuid",
  "position": 120,
  "is_playing": true,
  "server_timestamp": 1704067200
}
```

#### Pause Event

```json
{
  "type": "pause",
  "party_id": "party-uuid",
  "position": 120,
  "is_playing": false,
  "server_timestamp": 1704067200
}
```

#### Seek Event

```json
{
  "type": "seek",
  "party_id": "party-uuid",
  "position": 200,
  "is_playing": false,
  "server_timestamp": 1704067200
}
```

#### Skip Event

```json
{
  "type": "skip",
  "party_id": "party-uuid",
  "clip_id": "new-clip-uuid",
  "position": 0,
  "is_playing": true,
  "server_timestamp": 1704067200
}
```

#### Participant Joined

```json
{
  "type": "participant-joined",
  "party_id": "party-uuid",
  "server_timestamp": 1704067200,
  "participant": {
    "user_id": "user-uuid",
    "display_name": "NewUser",
    "avatar_url": "https://...",
    "role": "viewer"
  }
}
```

#### Participant Left

```json
{
  "type": "participant-left",
  "party_id": "party-uuid",
  "server_timestamp": 1704067200,
  "participant": {
    "user_id": "user-uuid",
    "display_name": "LeftUser",
    "avatar_url": "https://...",
    "role": "viewer"
  }
}
```

## Roles and Permissions

### Host

- Created the party
- Full control over playback (play, pause, seek, skip)
- Can end the party
- Cannot be removed from the party

### Co-Host

- Elevated by the host
- Full control over playback (play, pause, seek, skip)
- Cannot end the party

### Viewer

- Default role for participants
- Can request sync to catch up
- Cannot control playback

## Sync Tolerance

The client should implement a ±2 second tolerance for synchronization:
- If the local playback position differs by more than 2 seconds from the server position, sync immediately
- If the difference is less than 2 seconds, allow natural playback to smooth out minor network delays

## Rate Limits

- **Create Party:** 10 requests per hour per user
- **Join Party:** 30 requests per hour per user
- **Other endpoints:** Standard rate limits apply

## Best Practices

1. **Connection Management:**
   - Implement reconnection logic with exponential backoff
   - Send periodic ping messages to keep the connection alive
   - Handle disconnect gracefully and attempt to rejoin

2. **State Management:**
   - Request initial sync immediately after connecting
   - Update local state based on server events
   - Buffer events during brief disconnections

3. **Error Handling:**
   - Display user-friendly messages for common errors
   - Provide feedback when sync is occurring
   - Show connection status indicator

4. **Performance:**
   - Batch multiple state updates when possible
   - Avoid sending commands too frequently
   - Implement debouncing for seek operations

## Example Usage

### Creating and Joining a Party

1. **Host creates party:**
```bash
curl -X POST https://api.clpr.tv/v1/watch-parties \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{"title": "Movie Night", "visibility": "private"}'
```

2. **Host shares invite code** (e.g., "ABC123") with friends

3. **Friends join party:**
```bash
curl -X POST https://api.clpr.tv/v1/watch-parties/ABC123/join \
  -H "Authorization: Bearer <token>"
```

4. **Connect to WebSocket:**
```javascript
const ws = new WebSocket('wss://api.clpr.tv/v1/watch-parties/<party-id>/ws');

ws.onopen = () => {
  // Request initial sync
  ws.send(JSON.stringify({
    type: 'sync-request',
    party_id: partyId,
    timestamp: Date.now()
  }));
};

ws.onmessage = (event) => {
  const syncEvent = JSON.parse(event.data);
  handleSyncEvent(syncEvent);
};
```

5. **Host controls playback:**
```javascript
// Play
ws.send(JSON.stringify({
  type: 'play',
  party_id: partyId,
  timestamp: Date.now()
}));

// Seek to 2 minutes
ws.send(JSON.stringify({
  type: 'seek',
  party_id: partyId,
  position: 120,
  timestamp: Date.now()
}));
```
