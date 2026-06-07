# Playlist Features Implementation Status

## ✅ COMPLETED FIXES

### Critical Fixes (High Impact)

1. **Create Playlist Button** ✅
    - Added to PublicPlaylistsPage (Discover Playlists)
    - Navigates to `/playlists/new` route
    - File: `frontend/src/pages/PublicPlaylistsPage.tsx`

2. **Hide Playlist Button UI** ✅
    - Changed from `MoreVertical` icon to `ChevronLeft` arrow
    - Rotates 180° when sidebar is hidden (visual feedback)
    - File: `frontend/src/components/playlist/PlaylistTheatreMode.tsx`

3. **Request Processing Overlay - Dismiss Button** ✅
    - Added X button to dismiss/collapse overlay
    - Positioned at top-right of overlay
    - Calls `handleSkipNext()` when clicked
    - File: `frontend/src/components/playlist/PlaylistTheatreMode.tsx`

4. **Video Frame Size in Theatre Mode** ✅
    - Fixed from `h-[600px]` (constrained) to `flex-1` (flexible)
    - Now properly uses available vertical space
    - File: `frontend/src/components/playlist/PlaylistTheatreMode.tsx`

5. **Overlay Positioning Blocking Video Controls** ✅
    - Changed from `bottom-6` to `bottom-24` to clear bottom controls
    - Added `z-30` for proper layering
    - Now doesn't block bottom playback controls (Next button, progress)
    - File: `frontend/src/components/playlist/PlaylistTheatreMode.tsx`

6. **403 Error on Playlist Edit (Ownership) ** ✅
    - Fixed error message matching in handler
    - Changed from exact string comparison to `strings.Contains("unauthorized")`
    - Now catches all permission errors properly
    - File: `backend/internal/handlers/playlist_handler.go`
    - Added `strings` import

## 🚧 NEXT STEPS (Priority Order)

### 1. Create Playlist Creation Page (CRITICAL)

**Current Status:** Button navigates to `/playlists/new` but route/page doesn't exist yet

**What needs to be done:**

- Create `frontend/src/pages/PlaylistCreatePage.tsx` (new or edit form)
- Create `frontend/src/components/playlist/PlaylistForm.tsx` (reusable form)
- Add route to `frontend/src/App.tsx`: `<Route path="/playlists/new" element={<PlaylistCreatePage />} />`
- Add mutations hook: `useCreatePlaylist()` in `frontend/src/hooks/usePlaylist.ts`
- Backend endpoint exists: `POST /api/v1/playlists`

**Frontend needed:**

- Title field
- Description field
- Visibility selector (private/unlisted/public)
- Cover image upload
- Submit handler

### 2. Copy Playlist Feature

**What needs:**

- Backend: `POST /api/v1/playlists/{id}/copy` endpoint
- Frontend: Copy button on PlaylistDetail header
- Copy modal/dialog asking new name
- Copy should:
    - Duplicate all clips in order
    - Allow editing playlist name/description before saving
    - Make user the owner of copy

### 3. Playlist Card Styling (Like Collections Card)

**What needs:**

- Update `frontend/src/components/playlist/PlaylistCard.tsx`
- Match visual style of `DiscoveryListDetailCard.tsx` (curated collections)
- Current likely uses basic grid, needs:
    - Better image aspect ratio/hover effects
    - Similar text layout and typography
    - Same color palette and spacing

### 4. Permissions System (Multi-Editor)

**Backend exists:** `playlist_collaborators` table with permission levels
**Frontend needed:**

- Add collaborators button on PlaylistDetail
- Share dialog interface
- Show list of editors with permission levels
- Remove editor button (owner only)
- Change permission level dropdown (read/edit/admin)

**Permissions in DB:**

- `read`: View only
- `edit`: Can modify clips and metadata
- `admin`: Can edit + manage collaborators

### 5. Private Playlist Sharing with Link/Permission

**What already exists:**

- Database: `playlists.share_token` column
- Backend: Share token generation in UpdatePlaylist service

**Frontend needed:**

- Share button that shows:
    - Share link (includes share token)
    - QR code (optional nice-to-have)
- Access control selector when sharing
- Show who has access

### 6. Script-Based Playlists (Daily Top 10, etc.)

**Database:** No tables yet
**What needs:**

- New table: `playlist_scripts` (template definitions)
- New table: `generated_playlists` (playlists created from scripts)
- Backend endpoints for:
    - Create script template (admin?)
    - Generate playlist from script (scheduled or on-demand)
    - Regenerate playlist from script
- Frontend:
    - Admin interface for creating scripts
    - Schedule management for auto-generation
    - Display indicator for auto-generated playlists

---

## Implementation Notes

### Route Structure

Current route for playlist button:

```tsx
onClick={() => navigate('/playlists/new')}
```

Need to add to App.tsx:

```tsx
<Route path="/playlists/new" element={<PlaylistCreatePage />} />
<Route path="/playlists/:id/edit" element={<PlaylistEditPage />} />
```

### Notification Seeding

Note: Seeded notifications in `backend/migrations/seed.sql` line 851-869 don't cause 404s for end users because they're generic placeholder links like `/clips` and `/profile`. The 404 issues were from auto-generated notification links (submission approved notifications) which have already been fixed in the service layer.

---

## Testing Checklist

### Fixes Verification

- [ ] Create button appears on Discover Playlists
- [ ] Hide button is arrow icon, rotates when toggled
- [ ] dismiss button (X) appears on processing overlay
- [ ] Overlay doesn't cover bottom video controls
- [ ] Can click play/pause while overlay showing
- [ ] Edit any playlist without 403 error
- [ ] Video frame fills available space in theatre mode

### New Features Testing

- [ ] /playlists/new page loads
- [ ] Can create new playlist with form
- [ ] Can copy existing playlist
- [ ] Team members can see and edit shared playlists
- [ ] Share link generates correctly
- [ ] Private playlists only accessible with link/permission

---

## Files Changed This Session

### Frontend

- `frontend/src/pages/PublicPlaylistsPage.tsx`
    - Added Create button with Plus icon
    - Added navigate hook

- `frontend/src/components/playlist/PlaylistTheatreMode.tsx`
    - Import change: `MoreVertical` → `ChevronLeft`
    - Arrow button with rotation animation
    - Dismiss button on overlay
    - Overlay positioning fix (bottom-24, z-30)
    - Video player height: `h-full` → `flex-1`

### Backend

- `backend/internal/handlers/playlist_handler.go`
    - Added `strings` import
    - Fixed error handling for unauthorized access
    - Changed exact match to `strings.Contains("unauthorized")`

---

## Command Reference

### Build Backend

```bash
cd backend
go build -o bin/api ./cmd/api
```

### Restart Services

```bash
make restart-backend
# Or
docker-compose restart clpr-backend
```

### Run Tests

```bash
go test ./...
```

---

## Known Limitations & Notes

1. **Seeded Notification Links**: Generic `/clips`, `/profile` links in seed.sql don't match specific resources, but this is acceptable for test data.

2. **Permission Model**: Current DB supports read/edit/admin but frontend currently doesn't have UI for managing these.

3. **Script-Based Playlists**: Would require new DB schema and admin interface - more complex feature.

4. **Copy Feature**: Needs careful handling of clip order and metadata preservation.

---

**Status as of Feb 8, 2026**
Primary issues fixed, main blocker is missing Create page for the new button route.
