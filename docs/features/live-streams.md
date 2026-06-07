---
title: "Live Streams"
summary: "Complete guide to Clipper's live stream integration features, allowing users to watch Twitch streams directly on the platform with integrated chat, notifications, and clip creation."
tags: ["features"]
area: "features"
status: "stable"
owner: "team-core"
version: "1.0"
last_reviewed: 2026-01-29
---

# Live Stream Watching & Integration

Complete guide to Clipper's live stream integration features, allowing users to watch Twitch streams directly on the platform with integrated chat, notifications, and clip creation.

## Table of Contents

- [Overview](#overview)
- [Features](#features)
- [Architecture](#architecture)
- [User Guide](#user-guide)
- [API Reference](#api-reference)
- [Technical Implementation](#technical-implementation)
- [Performance](#performance)
- [Security](#security)
- [Monitoring](#monitoring)

## Overview

Clipper's live stream integration provides a seamless experience for watching Twitch streams without leaving the platform. Users can:

- Watch live Twitch streams with embedded player
- View and interact with Twitch chat
- Follow streamers for live notifications
- Create clips directly from live streams
- View recent clips from broadcasters

## Features

### 1. Stream Embedding & Playback

**Status**: ✅ Implemented

The Twitch Player component (`TwitchPlayer.tsx`) provides:

- **Live Stream Playback**: Embedded Twitch player using Twitch Embed SDK
- **Stream Status Detection**: Automatic detection of live/offline/ended states
- **Quality Selection**: Users can choose stream quality via Twitch player controls
- **Stream Metadata**: Display of stream title, game, and viewer count
- **Offline Screen**: Friendly offline screen when streamer is not live
- **Mobile Responsive**: Fully responsive player that adapts to screen size
- **Error Handling**: Graceful error handling for stream loading failures
- **Live Indicator**: Real-time viewer count and LIVE badge overlay

**Key Implementation Details**:

```typescript
// TwitchPlayer component
- Uses Twitch Embed SDK v1
- Auto-refresh stream status every 60 seconds
- Lazy-loads Twitch SDK for performance
- Properly cleans up embed instances
- Supports autoplay and muted options
```

### 2. Integrated Stream Chat Layer

**Status**: ✅ Implemented

The Chat Embed component (`TwitchChatEmbed.tsx`) provides:

- **Twitch Chat Embedding**: Official Twitch chat via iframe
- **Chat Position Toggle**: Side-by-side or bottom layout options
- **Authentication Status**: Shows connection status and login button
- **Dark Mode Support**: Chat inherits dark mode from Twitch
- **Mobile Responsive**: Adapts chat layout for mobile devices
- **Sandbox Security**: Proper iframe sandbox attributes for security

**Key Implementation Details**:

```typescript
// TwitchChatEmbed component
- Uses Twitch chat embed iframe
- Supports OAuth for authenticated chat participation
- Position options: 'side' (default) or 'bottom'
- Auto-detects parent domain for security
- Shows authentication status indicator
```

### 3. Stream Notifications & Scheduling

**Status**: ✅ Implemented

Stream follow system allows users to:

- **Follow Streamers**: Follow favorite streamers for updates
- **Notification Preferences**: Enable/disable live notifications per streamer
- **Follow Status Tracking**: View follow status and notification settings
- **Follow List**: View all followed streamers in one place
- **Real-time Updates**: Stream status updates via periodic polling (60s interval)

**API Endpoints**:

- `POST /api/v1/streams/:streamer/follow` - Follow a streamer
- `DELETE /api/v1/streams/:streamer/follow` - Unfollow a streamer
- `GET /api/v1/streams/:streamer/follow-status` - Check follow status
- `GET /api/v1/streams/following` - Get list of followed streamers

**Database Schema**:

```sql
stream_follows table:
- id (UUID)
- user_id (UUID, FK to users)
- streamer_username (VARCHAR)
- notifications_enabled (BOOLEAN)
- created_at (TIMESTAMP)
- updated_at (TIMESTAMP)
```

### 4. Stream Clip Submission & Watch-Along

**Status**: ✅ Implemented

The Clip Creator component (`ClipCreator.tsx`) provides:

- **Live Clip Creation**: Create clips from live streams
- **Timestamp Selection**: Choose start and end times (5-60 seconds)
- **Quality Selection**: Choose clip quality (source, 1080p, 720p)
- **Title Input**: Add custom titles to clips
- **Validation**: Client-side and server-side validation
- **Processing Status**: Shows clip processing status
- **Automatic Redirect**: Redirects to clip page after creation

**API Endpoints**:

- `POST /api/v1/streams/:streamer/clips` - Create clip from stream

**Clip Metadata**:

```typescript
{
  start_time: number;      // Seconds from stream start
  end_time: number;        // Seconds from stream start
  quality: 'source' | '1080p' | '720p';
  title: string;           // 3-255 characters
}
```

## Architecture

### Frontend Components

```
StreamPage.tsx
├── TwitchPlayer.tsx          # Main stream player
│   ├── StreamOfflineScreen   # Shown when offline
│   └── LiveIndicator         # Live status overlay
├── TwitchChatEmbed.tsx       # Twitch chat iframe
├── StreamFollowButton.tsx    # Follow/unfollow controls
└── ClipCreator.tsx           # Clip creation modal
```

### Backend Services

```
StreamHandler
├── GetStreamStatus          # Fetch stream status from Twitch API
├── CreateClipFromStream     # Create clip from live stream
├── FollowStreamer          # Follow a streamer
├── UnfollowStreamer        # Unfollow a streamer
├── GetFollowedStreamers    # List followed streamers
└── GetStreamFollowStatus   # Check follow status
```

### Data Flow

1. **Stream Status Check**:
   ```
   User → Frontend → Backend API → Twitch API → Redis Cache → Response
   ```

2. **Follow Streamer**:
   ```
   User → Frontend → Backend API → Database → Response
   ```

3. **Create Clip**:
   ```
   User → ClipCreator Modal → Backend API → Database → Processing Queue
   ```

## User Guide

### Watching a Live Stream

1. Navigate to `/stream/:streamer` (e.g., `/stream/shroud`)
2. The page automatically checks if the streamer is live
3. If live, the stream player loads with chat
4. If offline, an offline screen is displayed with recent clips

### Customizing Chat Layout

- **Toggle Chat**: Use "Hide Chat" / "Show Chat" button
- **Change Position**: Select "Side" or "Bottom" from dropdown
- **Login to Chat**: Click "Login to Chat" to participate (requires Twitch OAuth)

### Following a Streamer

1. Click the "Follow" button on the stream page
2. Choose whether to enable notifications (enabled by default)
3. View followed streamers at `/streams/following` (when implemented)

### Creating Clips from Live Streams

1. While watching a live stream, click "Create Clip" button
2. Enter clip title (3-255 characters)
3. Adjust start and end times (5-60 second duration)
4. Select quality preference
5. Click "Create Clip"
6. Wait for processing (usually < 30 seconds)
7. Automatically redirected to clip page

## API Reference

### Get Stream Status

**Endpoint**: `GET /api/v1/streams/:streamer`

**Description**: Fetch current stream status for a streamer

**Parameters**:
- `streamer` (path, required): Twitch username

**Response** (200 OK):
```json
{
  "streamer_username": "shroud",
  "is_live": true,
  "title": "Valorant Ranked | !pc !sponsor",
  "game_name": "VALORANT",
  "viewer_count": 15234,
  "thumbnail_url": "https://...",
  "started_at": "2024-01-15T18:30:00Z",
  "last_went_offline": null
}
```

**Caching**: Responses are cached in Redis for 60 seconds

### Create Clip from Stream

**Endpoint**: `POST /api/v1/streams/:streamer/clips`

**Description**: Create a clip from a live stream

**Authentication**: Required (JWT token)

**Rate Limit**: 10 requests per hour per user

**Request Body**:
```json
{
  "streamer_username": "shroud",
  "start_time": 150.5,
  "end_time": 180.5,
  "quality": "1080p",
  "title": "Amazing ace clutch!"
}
```

**Validation**:
- `title`: 3-255 characters
- `start_time` < `end_time`
- Duration: 5-60 seconds
- Stream must be live

**Response** (201 Created):
```json
{
  "clip_id": "550e8400-e29b-41d4-a716-446655440000",
  "status": "processing"
}
```

### Follow Streamer

**Endpoint**: `POST /api/v1/streams/:streamer/follow`

**Description**: Follow a streamer for live notifications

**Authentication**: Required

**Rate Limit**: 20 requests per minute per user

**Request Body**:
```json
{
  "notifications_enabled": true
}
```

**Response** (200 OK):
```json
{
  "following": true,
  "notifications_enabled": true,
  "message": "Successfully following shroud"
}
```

### Unfollow Streamer

**Endpoint**: `DELETE /api/v1/streams/:streamer/follow`

**Authentication**: Required

**Response** (200 OK):
```json
{
  "following": false,
  "message": "Successfully unfollowed shroud"
}
```

### Get Follow Status

**Endpoint**: `GET /api/v1/streams/:streamer/follow-status`

**Authentication**: Required

**Response** (200 OK):
```json
{
  "following": true,
  "notifications_enabled": true
}
```

### Get Followed Streamers

**Endpoint**: `GET /api/v1/streams/following`

**Authentication**: Required

**Response** (200 OK):
```json
{
  "follows": [
    {
      "id": "...",
      "user_id": "...",
      "streamer_username": "shroud",
      "notifications_enabled": true,
      "created_at": "2024-01-15T10:00:00Z",
      "updated_at": "2024-01-15T10:00:00Z"
    }
  ],
  "count": 1
}
```

## Technical Implementation

### Twitch Embed SDK Integration

```typescript
// Load Twitch Embed SDK
const script = document.createElement('script');
script.src = 'https://embed.twitch.tv/embed/v1.js';
script.async = true;

// Initialize embed
const embed = new window.Twitch.Embed('twitch-embed-container', {
  width: '100%',
  height: '100%',
  channel: 'streamer_name',
  layout: 'video', // or 'video-with-chat'
  autoplay: true,
  muted: false,
  parent: [window.location.hostname],
});
```

### Stream Status Polling

The frontend uses React Query with a 60-second refetch interval:

```typescript
const { data: streamInfo } = useQuery({
  queryKey: ['streamStatus', streamer],
  queryFn: () => fetchStreamStatus(streamer!),
  enabled: !!streamer,
  refetchInterval: 60000, // 60 seconds
});
```

### Database Indexes

Optimized queries with indexes:

```sql
-- Stream lookups
CREATE INDEX idx_streams_username ON streams(streamer_username);
CREATE INDEX idx_streams_live ON streams(is_live, streamer_username);

-- Follow lookups
CREATE INDEX idx_stream_follows_user ON stream_follows(user_id);
CREATE INDEX idx_stream_follows_streamer_notifications 
  ON stream_follows(streamer_username, notifications_enabled) 
  WHERE notifications_enabled = TRUE;
```

### Caching Strategy

**Stream Status Cache**:
- Key: `stream:status:{streamer_username}`
- TTL: 60 seconds
- Invalidation: On stream state change

**Twitch API Rate Limits**:
- Protected by Redis-based rate limiting
- Graceful degradation on API failures

## Performance

### Metrics & Targets

| Metric | Target | Current Status |
|--------|--------|----------------|
| Stream Load Time | < 2s | ✅ Optimized |
| Status API Response | < 100ms | ✅ Cached |
| Concurrent Viewers | 100+ | ✅ Supported |
| Chat Load Time | < 1s | ✅ Iframe |
| Clip Creation Time | < 30s | 🔄 Processing |

### Optimizations

1. **Lazy Loading**: Twitch SDK loaded on-demand
2. **Caching**: Stream status cached in Redis
3. **CDN**: Static assets served via CDN
4. **Code Splitting**: Stream page lazy-loaded
5. **Image Optimization**: Thumbnails optimized

## Security

### Authentication

- JWT tokens required for:
  - Following streamers
  - Creating clips
  - Viewing followed streamers

### Authorization

- Users can only manage their own follows
- Clip creation requires valid user account
- Rate limiting prevents abuse

### Input Validation

**Streamer Username**:
- Length: 4-25 characters
- Pattern: Alphanumeric + underscore only

**Clip Title**:
- Length: 3-255 characters
- XSS protection via sanitization

**Time Ranges**:
- Start time < End time
- Duration: 5-60 seconds
- Non-negative values only

### Rate Limiting

- Follow/Unfollow: 20 requests/minute/user
- Clip Creation: 10 requests/hour/user
- Status Checks: Unlimited (cached)

### Iframe Security

Chat embed uses sandbox attributes:

```html
<iframe
  sandbox="allow-storage-access-by-user-activation 
           allow-scripts 
           allow-same-origin 
           allow-popups 
           allow-popups-to-escape-sandbox 
           allow-modals"
/>
```

## Monitoring

### Metrics to Track

1. **Stream Viewing**:
   - Concurrent stream viewers
   - Average watch duration
   - Stream load failures

2. **Engagement**:
   - Clips created per stream
   - Follow conversion rate
   - Notification click-through rate

3. **Performance**:
   - Stream load time (p50, p95, p99)
   - API response times
   - Cache hit rate

4. **Errors**:
   - Twitch API failures
   - Embed load failures
   - Clip creation failures

### Logging

Key events logged:

```go
// Stream viewing
logger.Info("Stream viewed", map[string]interface{}{
  "streamer": streamer,
  "is_live": isLive,
})

// Clip creation
logger.Info("Clip from stream created", map[string]interface{}{
  "clip_id": clipID,
  "streamer": streamer,
  "user_id": userID,
})

// Follow actions
logger.Info("User followed streamer", map[string]interface{}{
  "user_id": userID,
  "streamer": streamer,
})
```

### Alerts

Set up alerts for:

- Stream load time > 3s (p95)
- Twitch API error rate > 5%
- Clip creation failure rate > 10%
- Cache miss rate > 50%

## Future Enhancements

### Planned Features

1. **WebSocket Updates**: Real-time stream status without polling
2. **Stream Schedule**: Import and display stream schedules from Twitch
3. **Email Notifications**: Email alerts when followed streamers go live
4. **Push Notifications**: Browser push for live alerts
5. **Custom Chat**: Optional Clipper-specific chat overlay
6. **VOD Playback**: Watch past broadcasts
7. **Picture-in-Picture**: Native PiP support
8. **Stream Calendar**: Calendar view of upcoming streams
9. **Watch Parties**: Synchronized watching with friends
10. **Stream Analytics**: Detailed viewing statistics

### API Improvements

- GraphQL API for flexible stream data queries
- Webhook support for stream events
- Batch operations for following multiple streamers
- Advanced filtering and sorting for followed streams

## Troubleshooting

### Common Issues

**Stream won't load**:
- Check that Twitch Embed SDK loaded (check browser console)
- Verify parent domain is whitelisted
- Check if adblockers are interfering
- Try refreshing the page

**Chat not showing**:
- Verify iframe sandbox permissions
- Check browser console for errors
- Ensure Twitch chat is accessible
- Try logging in to Twitch

**Can't create clips**:
- Ensure stream is currently live
- Check authentication status
- Verify clip duration is 5-60 seconds
- Check rate limit status

**Follow button not working**:
- Ensure you're logged in
- Check network tab for API errors
- Verify rate limits not exceeded

## Related Documentation

- [Twitch Integration Guide](../backend/twitch-integration.md)
- [API Reference](../backend/api.md)
- [Caching Strategy](../backend/caching-strategy.md)
- [User Guide](../users/user-guide.md)

## Support

For issues or questions:

- GitHub Issues: [Create an issue](https://git.subcult.tv/subculture-collective/clpr/issues)
- Documentation: [View full docs](../index.md)
- Community: [Join discussions](https://git.subcult.tv/subculture-collective/clpr/discussions)
