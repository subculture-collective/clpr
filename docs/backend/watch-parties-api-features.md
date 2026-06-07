---
title: "WATCH PARTIES API"
summary: "Watch Parties enable synchronized video watching with friends and community members. The system provides real-time synchronization with a target tolerance of ±2 seconds, role-based access control, and"
tags: ["docs","api"]
area: "docs"
status: "stable"
owner: "team-core"
version: "1.0"
last_reviewed: 2026-01-29
---

# Watch Parties API Documentation

## Overview

Watch Parties enable synchronized video watching with friends and community members. The system provides real-time synchronization with a target tolerance of ±2 seconds, role-based access control, and WebSocket-based communication for low-latency sync events.

## Architecture

### Components

- **WatchPartyService**: Business logic for party creation and management
- **WatchPartyHub**: Real-time WebSocket synchronization engine
- **WatchPartyHubManager**: Manages multiple concurrent party hubs
- **Database**: Persistent storage for party state and participants

### Sync Tolerance

The system is designed with a **±2 second sync tolerance** for video playback synchronization:
- Server maintains authoritative playback state
- Clients adjust their playback position if drift exceeds ±2 seconds
- Network latency and client processing time are factored into sync calculations

## HTTP API Endpoints

### Create Watch Party

**Endpoint**: `POST /api/v1/watch-parties`

**Authentication**: Required

**Request Body**:
```json
{
  "title": "Friday Night Movies",
  "playlist_id": "optional-uuid",
  "visibility": "private|public|friends",
  "max_participants": 100
}
```

**Response** (201 Created):
```json
{
  "id": "party-uuid",
  "host_user_id": "user-uuid",
  "title": "Friday Night Movies",
  "playlist_id": "optional-uuid",
  "visibility": "private",
  "invite_code": "ABC123",
  "max_participants": 100,
  "status": "active",
  "created_at": "2025-12-31T15:00:00Z"
}
```

**Rate Limit**: 10 creates per hour per user

---

### Join Watch Party

**Endpoint**: `POST /api/v1/watch-parties/:code/join`

**Authentication**: Required

**Path Parameters**:
- `code`: 6-character invite code

**Response** (200 OK):
```json
{
  "party_id": "party-uuid",
  "ws_url": "wss://api.clpr.com/api/v1/watch-parties/:id/ws"
}
```

**Rate Limit**: 30 joins per hour per user

---

### Get Watch Party Details

**Endpoint**: `GET /api/v1/watch-parties/:id`

**Authentication**: Required (must be participant)

**Response** (200 OK):
```json
{
  "id": "party-uuid",
  "host_user_id": "user-uuid",
  "title": "Friday Night Movies",
  "current_clip_id": "clip-uuid",
  "current_position": 120,
  "is_playing": true,
  "participant_count": 5,
  "max_participants": 100,
  "status": "active"
}
```

---

### Get Participants

**Endpoint**: `GET /api/v1/watch-parties/:id/participants`

**Authentication**: Required (must be participant)

**Response** (200 OK):
```json
{
  "participants": [
    {
      "user_id": "user-uuid",
      "display_name": "Alice",
      "avatar_url": "https://...",
      "role": "host",
      "is_active": true,
      "last_sync": "2025-12-31T15:00:00Z"
    }
  ]
}
```

---

### Leave Watch Party

**Endpoint**: `DELETE /api/v1/watch-parties/:id/leave`

**Authentication**: Required (must be participant)

**Response** (204 No Content)

---

### End Watch Party

**Endpoint**: `POST /api/v1/watch-parties/:id/end`

**Authentication**: Required (must be host)

**Response** (204 No Content)

## WebSocket Protocol

### Connection

**Endpoint**: `GET /api/v1/watch-parties/:id/ws`

**Authentication**: Required via JWT in the `Authorization` header (e.g., `Authorization: Bearer <jwt>`).

**Security Note**: Do not send JWT tokens in URL query parameters as they will be logged by proxies, browsers, and servers, creating a security risk.

**Upgrade Headers**:
```
Upgrade: websocket
Connection: Upgrade
```

### Message Types

#### Client → Server Commands

**Play Command**:
```json
{
  "type": "play",
  "party_id": "party-uuid",
  "timestamp": 1735660800
}
```

**Pause Command**:
```json
{
  "type": "pause",
  "party_id": "party-uuid",
  "timestamp": 1735660800
}
```

**Seek Command**:
```json
{
  "type": "seek",
  "party_id": "party-uuid",
  "position": 120,
  "timestamp": 1735660800
}
```

**Skip Command** (change clip):
```json
{
  "type": "skip",
  "party_id": "party-uuid",
  "clip_id": "new-clip-uuid",
  "timestamp": 1735660800
}
```

**Sync Request** (request current state):
```json
{
  "type": "sync-request",
  "party_id": "party-uuid",
  "timestamp": 1735660800
}
```

#### Server → Client Events

**Sync Event** (sent after any state change):
```json
{
  "type": "sync",
  "party_id": "party-uuid",
  "clip_id": "clip-uuid",
  "position": 120,
  "is_playing": true,
  "server_timestamp": 1735660800
}
```

**Play Event**:
```json
{
  "type": "play",
  "party_id": "party-uuid",
  "position": 120,
  "is_playing": true,
  "server_timestamp": 1735660800
}
```

**Pause Event**:
```json
{
  "type": "pause",
  "party_id": "party-uuid",
  "position": 120,
  "is_playing": false,
  "server_timestamp": 1735660800
}
```

**Seek Event**:
```json
{
  "type": "seek",
  "party_id": "party-uuid",
  "position": 150,
  "is_playing": false,
  "server_timestamp": 1735660800
}
```

**Note**: Seek commands automatically pause playback (`is_playing: false`) to ensure all clients are synchronized at the new position.

**Skip Event**:
```json
{
  "type": "skip",
  "party_id": "party-uuid",
  "clip_id": "new-clip-uuid",
  "position": 0,
  "is_playing": true,
  "server_timestamp": 1735660800
}
```

**Participant Joined**:
```json
{
  "type": "participant-joined",
  "party_id": "party-uuid",
  "server_timestamp": 1735660800,
  "participant": {
    "user_id": "user-uuid",
    "display_name": "Bob",
    "avatar_url": "https://...",
    "role": "viewer"
  }
}
```

**Participant Left**:
```json
{
  "type": "participant-left",
  "party_id": "party-uuid",
  "server_timestamp": 1735660800,
  "participant": {
    "user_id": "user-uuid",
    "display_name": "Bob",
    "role": "viewer"
  }
}
```

## Role Permissions

### Host
- **Can**: Create party, control playback (play/pause/seek/skip), end party, promote to co-host, kick participants
- **Control Commands**: All playback commands authorized

### Co-Host
- **Can**: Control playback (play/pause/seek/skip)
- **Cannot**: End party, manage participants
- **Control Commands**: All playback commands authorized

### Viewer
- **Can**: Watch synchronized video, see participant list, request sync
- **Cannot**: Control playback
- **Control Commands**: Playback commands return `401 Unauthorized`

### Permission Enforcement

The server enforces role permissions before processing commands:

1. Command received from client
2. Server validates user role
3. If role lacks permission: Command is silently ignored (no error response, no state change, no broadcast)
4. If role has permission: Process command, update state, broadcast to all

**Note**: Unauthorized playback commands are silently rejected without sending an error response to the client.

## Synchronization Behavior

### Normal Operation

1. **Host/Co-host sends command** (e.g., play, pause, seek)
2. **Server receives command** and validates permissions
3. **Server updates authoritative state** in database
4. **Server broadcasts sync event** to all connected clients
5. **Clients receive event** and adjust their playback within ±2s tolerance

### Drift Tolerance

Clients should implement the following logic:

```javascript
function handleSyncEvent(event) {
  const serverPosition = event.position;
  const serverTime = event.server_timestamp;
  const clientTime = Date.now() / 1000;
  const latency = clientTime - serverTime;

  // Calculate expected position accounting for latency
  const expectedPosition = event.is_playing
    ? serverPosition + latency
    : serverPosition;

  const currentPosition = videoPlayer.currentTime;
  const drift = Math.abs(currentPosition - expectedPosition);

  // Only adjust if drift exceeds tolerance
  if (drift > 2.0) {
    videoPlayer.currentTime = expectedPosition;
  }

  if (event.is_playing !== videoPlayer.playing) {
    event.is_playing ? videoPlayer.play() : videoPlayer.pause();
  }
}
```

### Reconnection Recovery

When a client reconnects after disconnection:

1. **Client sends `sync-request` command**
2. **Server responds with current state** (sync event)
3. **Client adjusts to current position**
4. **Client updates UI** with participant list

**Best Practice**: Request sync immediately after WebSocket connection established.

### Network Conditions

The system handles various network conditions:

- **Normal latency** (<100ms): No noticeable sync issues
- **Mild packet loss** (1-5%): Clients may miss events; sync-request recovers state
- **High latency** (>500ms): Increased drift; clients adjust more frequently
- **Disconnection**: Client buffer continues; reconnection triggers sync-request

## Error Handling

### HTTP Errors

- `400 Bad Request`: Invalid request body or parameters
- `401 Unauthorized`: Missing or invalid authentication
- `403 Forbidden`: Insufficient permissions (e.g., non-host ending party)
- `404 Not Found`: Party or resource not found
- `429 Too Many Requests`: Rate limit exceeded
- `500 Internal Server Error`: Server-side error

### WebSocket Errors

**Error Event**:
```json
{
  "type": "error",
  "message": "Error description",
  "code": "ERROR_CODE"
}
```

**Error Codes**:
- `UNAUTHORIZED_COMMAND`: User lacks permission for command
- `INVALID_COMMAND`: Malformed command structure
- `PARTY_ENDED`: Party has ended
- `PARTICIPANT_LIMIT`: Maximum participants reached
- `RATE_LIMITED`: Command rate limit exceeded

## Rate Limiting

### HTTP Endpoints
- **Create Party**: 10 requests per hour
- **Join Party**: 30 requests per hour

### WebSocket Commands
- **Chat Messages**: 10 messages per minute per user
- **Reactions**: 30 reactions per minute per user
- **Playback Commands**: Unlimited (but best practice: avoid rapid commands)

## Best Practices

### Client Implementation

1. **Always send sync-request on connection**: Ensures state consistency
2. **Implement ±2s tolerance**: Avoid constant micro-adjustments
3. **Handle reconnection gracefully**: Buffer continues playing during brief disconnects
4. **Show sync status indicator**: Visual feedback when syncing
5. **Log sync events for debugging**: Timestamp, drift, action taken

### Performance

1. **Minimize command frequency**: Avoid sending commands during drag events
2. **Batch reactions**: Accumulate rapid reactions and send in batches
3. **Use WebSocket ping/pong**: Detect disconnections early
4. **Implement exponential backoff**: For reconnection attempts

### Testing

1. **Test with simulated latency**: Use network throttling tools
2. **Test reconnection scenarios**: Simulate disconnections
3. **Test permission enforcement**: Verify viewers cannot control
4. **Test with multiple clients**: Ensure sync events broadcast correctly

## Example Client Code

### Connection Setup

```javascript
const partyId = 'party-uuid';
// Authentication is handled via Authorization header during WebSocket upgrade
// The server checks for a valid JWT in the Authorization header
const wsUrl = `wss://api.clpr.com/api/v1/watch-parties/${partyId}/ws`;

// Note: WebSocket API doesn't directly support custom headers in browser environments
// For browser clients, use an existing authenticated session (HttpOnly cookies)
// or implement a token exchange mechanism on the server side
const ws = new WebSocket(wsUrl);

ws.onopen = () => {
  console.log('Connected to watch party');

  // Request current state
  ws.send(JSON.stringify({
    type: 'sync-request',
    party_id: partyId,
    timestamp: Math.floor(Date.now() / 1000)
  }));
};

ws.onmessage = (event) => {
  const data = JSON.parse(event.data);
  handleSyncEvent(data);
};

ws.onerror = (error) => {
  console.error('WebSocket error:', error);
};

ws.onclose = () => {
  console.log('Disconnected from watch party');
  // Implement reconnection logic
};
```

### Sending Commands

```javascript
function sendPlayCommand() {
  if (userRole === 'viewer') {
    console.error('Viewers cannot control playback');
    return;
  }

  ws.send(JSON.stringify({
    type: 'play',
    party_id: partyId,
    timestamp: Math.floor(Date.now() / 1000)
  }));
}

function sendSeekCommand(position) {
  if (userRole === 'viewer') {
    console.error('Viewers cannot control playback');
    return;
  }

  ws.send(JSON.stringify({
    type: 'seek',
    party_id: partyId,
    position: Math.floor(position),
    timestamp: Math.floor(Date.now() / 1000)
  }));
}
```

## Monitoring and Metrics

The system logs the following metrics for observability:

- **Sync event latency**: Time from command received to event broadcast
- **Client drift**: Measured drift when clients report position
- **Command frequency**: Commands per minute per user
- **Participant churn**: Join/leave rate
- **Error rate**: Failed commands and reasons

These metrics should be monitored in production to ensure sync quality and identify issues.

## References

- Implementation: `backend/internal/services/watch_party_hub.go`
- Models: `backend/internal/models/models.go`
- Tests: `backend/tests/integration/watch_parties/`
- Testing Guide: `docs/testing/TESTING.md`
