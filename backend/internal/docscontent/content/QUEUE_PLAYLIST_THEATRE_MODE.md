# Queue to Playlist & Theatre Mode Implementation

## Overview

This document describes the implementation of two major features:

1. **Convert Queue to Playlist** - Transform your watching queue into a persistent playlist
2. **Theatre Mode** - A cinematic viewing experience for both queues and playlists with an interactive sidebar

## Features Implemented

### 1. Convert Queue to Playlist

#### Backend Implementation

- **API Endpoint**: `POST /api/queue/convert-to-playlist`
- **Location**: `/backend/internal/handlers/queue_handler.go` (ConvertToPlaylist)
- **Service**: `/backend/internal/services/queue_service.go` (ConvertQueueToPlaylist)

**Request Parameters**:

```json
{
    "title": "My Playlist", // Required: Playlist title
    "description": "Optional desc", // Optional: Playlist description
    "only_unplayed": false, // Optional: Only convert unplayed clips
    "clear_queue": false // Optional: Clear queue after conversion
}
```

**Features**:

- Creates a new private playlist from queue items
- Option to include only unplayed clips
- Option to clear queue after conversion
- Preserves queue order in the new playlist
- Rate limited to 10 requests per minute

#### Frontend Implementation

- **Dialog Component**: `/frontend/src/components/queue/ConvertToPlaylistDialog.tsx`
- **Hook**: `/frontend/src/hooks/useQueue.ts` (useConvertQueueToPlaylist)
- **Integration**: Added button to Queue page header

**User Flow**:

1. User clicks "Convert to Playlist" button on Queue page
2. Dialog opens with options:
    - Playlist title (required)
    - Description (optional)
    - "Only include unplayed clips" checkbox
    - "Clear queue after conversion" checkbox
3. User fills out form and submits
4. Playlist is created and user is redirected to the new playlist page
5. Success toast notification appears

### 2. Theatre Mode

#### Unified Theatre Mode Component

- **Component**: `/frontend/src/components/playlist/PlaylistTheatreMode.tsx`
- **Features**:
    - Full-screen cinematic video player
    - Interactive sidebar showing playlist/queue items
    - Drag-and-drop reordering (when supported)
    - Auto-advance to next unplayed clip
    - Current clip highlighting
    - Keyboard shortcuts:
        - `N` or `Arrow Down` - Skip to next clip
        - `S` - Toggle sidebar visibility
    - Visual indicators:
        - Currently playing clip
        - Watched/played clips
        - Clip thumbnails with duration
    - Remove clips functionality

#### Queue Theatre Mode

- **Route**: `/queue/theatre`
- **Page**: `/frontend/src/pages/QueueTheatrePage.tsx`
- **Access**: "Theatre Mode" button on Queue page

**Features**:

- Watch queue in full-screen mode
- Auto-advances to next unplayed clip
- Marks clips as played when watched
- Drag-and-drop to reorder queue
- Remove individual clips
- Shows played status with reduced opacity

#### Playlist Theatre Mode

- **Route**: `/playlists/:id/theatre`
- **Page**: `/frontend/src/pages/PlaylistTheatrePage.tsx`
- **Access**: "Theatre Mode" button on Playlist detail page

**Features**:

- Watch playlist in full-screen mode
- Auto-advances to next clip
- Drag-and-drop to reorder playlist (if user has permission)
- Remove clips from playlist (if user has permission)
- Continuous playback through entire playlist

## UI/UX Improvements

### Queue Page Enhancements

- Added "Theatre Mode" button (primary action)
- Added "Convert to Playlist" button
- Reorganized action buttons for better hierarchy

### Playlist Detail Page Enhancements

- Added "Theatre Mode" button next to the Like button
- Consistent action button placement

### Theatre Mode Interface

- **Video Player Area**:
    - Large, centered video player
    - Top bar with playlist/queue title and current clip info
    - Exit button (returns to list view)
    - Sidebar toggle button

- **Sidebar** (400px wide):
    - Header with item count
    - Scrollable list of clips
    - Drag handles for reordering
    - Numbered list (1, 2, 3...)
    - Clip thumbnails with duration overlay
    - Currently playing indicator (play icon badge)
    - Watched indicator (checkmark)
    - Hover effects and smooth transitions
    - Remove button on hover
    - Keyboard shortcuts hint at bottom

## Technical Architecture

### Component Hierarchy

```
QueueTheatrePage / PlaylistTheatrePage
  └── PlaylistTheatreMode
        ├── VideoPlayer (existing component)
        └── Sidebar
              └── PlaylistItem[] (with drag & drop)
```

### Data Flow

**Queue Theatre Mode**:

1. Fetch queue data using `useQueue` hook
2. Transform queue items to `PlaylistItem[]` format
3. Pass to `PlaylistTheatreMode` with queue-specific handlers
4. Handle interactions (mark as played, remove, reorder)

**Playlist Theatre Mode**:

1. Fetch playlist using `usePlaylist` hook
2. Extract clips from `playlist.data.clips`
3. Transform to `PlaylistItem[]` format
4. Pass to `PlaylistTheatreMode` with playlist-specific handlers
5. Handle interactions (remove, reorder)

### State Management

- Uses React Query for server state
- Local state for:
    - Current playing item
    - Drag-and-drop operations
    - Sidebar visibility
    - Dialog open/close

## API Changes

### New Endpoint

```
POST /api/queue/convert-to-playlist
Authorization: Required
Rate Limit: 10/minute
Content-Type: application/json
```

### Modified Services

- `QueueService`: Added `ConvertQueueToPlaylist` method
- Updated constructor to include `PlaylistService` dependency

### Database

No schema changes required - uses existing tables:

- `queue_items`
- `playlists`
- `playlist_clips`

## Testing Recommendations

### Manual Testing Checklist

- [ ] Convert empty queue to playlist (should error)
- [ ] Convert queue with all played items (with only_unplayed=true, should error)
- [ ] Convert queue to playlist successfully
- [ ] Convert queue with only_unplayed option
- [ ] Convert queue with clear_queue option
- [ ] Open theatre mode with empty queue/playlist
- [ ] Play clips in theatre mode
- [ ] Auto-advance to next clip
- [ ] Drag and drop to reorder
- [ ] Remove clips
- [ ] Keyboard shortcuts (N, S)
- [ ] Toggle sidebar
- [ ] Exit theatre mode
- [ ] Responsive behavior (sidebar should work on smaller screens)

### Edge Cases

- Queue/playlist becomes empty during playback
- Network errors during conversion
- Permission issues (private playlists, etc.)
- Very long playlists (1000+ clips)
- Rapid reordering/removal operations

## Future Enhancements

### Potential Improvements

1. **Watch Party Integration**: Reuse `PlaylistTheatreMode` in watch parties
2. **Playback Controls**: Add more controls (loop, shuffle)
3. **Picture-in-Picture**: Allow smaller player while browsing
4. **Keyboard Shortcuts**: Add more shortcuts (play/pause, volume, etc.)
5. **View Modes**: Grid view option for sidebar
6. **Filters**: Filter played/unplayed in theatre mode
7. **Export**: Export playlist as JSON or share links
8. **Analytics**: Track completion rates, popular playlists
9. **Recommendations**: Suggest next playlist based on watching history

### Code Unification Opportunities

- VideoPlayer component could be further optimized for theatre mode
- Watch party sync logic could be integrated
- Shared UI components for playlist/queue items

## Files Modified

### Backend

- `/backend/internal/services/queue_service.go`
- `/backend/internal/handlers/queue_handler.go`
- `/backend/internal/models/models.go`
- `/backend/cmd/api/main.go`

### Frontend - New Files

- `/frontend/src/components/queue/ConvertToPlaylistDialog.tsx`
- `/frontend/src/components/playlist/PlaylistTheatreMode.tsx`
- `/frontend/src/pages/QueueTheatrePage.tsx`
- `/frontend/src/pages/PlaylistTheatrePage.tsx`

### Frontend - Modified Files

- `/frontend/src/App.tsx`
- `/frontend/src/pages/QueuePage.tsx`
- `/frontend/src/components/playlist/PlaylistDetail.tsx`
- `/frontend/src/hooks/useQueue.ts`
- `/frontend/src/types/queue.ts`

## Migration Guide

### For Users

1. No database migrations required
2. Feature is immediately available after deployment
3. Existing queues and playlists work without changes

### For Developers

1. Pull latest changes
2. Rebuild backend: `cd backend && go build`
3. Rebuild frontend: `cd frontend && npm install && npm run build`
4. Restart services
5. Feature should be accessible at:
    - `/queue/theatre` (queue theatre mode)
    - `/playlists/:id/theatre` (playlist theatre mode)

## Performance Considerations

### Optimizations

- Theatre mode components use React.memo where appropriate
- Lazy loading of video player
- Virtualized list for very long playlists (could be added)
- Debounced drag-and-drop updates

### Loading States

- Spinner shown while loading queue/playlist
- Skeleton UI for sidebar items (could be added)
- Graceful error handling with fallback UI

## Accessibility

### Features

- Keyboard navigation support
- ARIA labels on interactive elements
- Focus management for dialogs
- Semantic HTML structure
- High contrast indicators

### Improvements Needed

- Screen reader announcements for playback events
- Better focus trap in dialogo
- Keyboard-only drag-and-drop alternative

---

**Implementation Date**: February 2026
**Authors**: GitHub Copilot
**Version**: 1.0.0
