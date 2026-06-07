# Streamer Clip Room Implementation Plan

> **For agentic workers:** Execute this plan task-by-task. Recommended path:
> dispatch a fresh subagent per task, review each result with `review-quality`,
> then continue. For complex multi-agent splits, use
> `parallel-feature-development`, `team-composition-patterns`, and
> `team-communication-protocols`. Steps use checkbox (`- [ ]`) syntax for
> tracking.

**Goal:** Build a Twitch streamer tool where Clpr clip links posted in Twitch chat enter a quick approval queue, approved clips populate a theatre-mode playlist, streamers can reorder playback, and comments/discussion for the current clip are one click away.

**Architecture:** Add a first-class backend `streamer_clip_rooms` resource with persistent detected items, approval/rejection state, ordered approved playback, and a server-side Twitch chat listener. Reuse existing Clpr playback/comment/reorder UI by wrapping `PlaylistTheatreMode` with a streamer-room data source and a pending approval rail.

**Tech Stack:** Go/Gin/pgx/PostgreSQL/Redis backend, Twitch OAuth `chat:read`, Gorilla WebSocket patterns already in `backend/internal/websocket`, React 19/Vite/TypeScript/React Router/TanStack Query frontend.

**pipeline-type:** full
**persistence-mode:** hybrid
**change-name:** streamer-clip-room

---

## Current Codebase Anchors

- Frontend routing: `frontend/src/App.tsx`.
- Page exports: `frontend/src/pages/index.ts`.
- Theatre playlist/reorder/comments UI: `frontend/src/components/playlist/PlaylistTheatreMode.tsx`.
- Playlist theatre route wrapper: `frontend/src/pages/PlaylistTheatrePage.tsx`.
- Queue hooks and reorder style: `frontend/src/hooks/useQueue.ts`.
- Playlist hooks and reorder style: `frontend/src/hooks/usePlaylist.ts`.
- Comments component: `frontend/src/components/comment/CommentSection.tsx`.
- Stream page/Twitch embed surface: `frontend/src/pages/StreamPage.tsx`.
- Existing frontend Twitch OAuth client: `frontend/src/lib/twitch-api.ts`.
- Backend route registration: `backend/cmd/api/routes_platform.go`, `backend/cmd/api/routes_social.go`.
- Existing queue backend: `backend/internal/handlers/queue_handler.go`, `backend/internal/services/queue_service.go`, `backend/internal/repository/queue_repository.go`.
- Existing playlist backend: `backend/internal/services/playlist_service.go`, `backend/internal/repository/playlist_repository.go`.
- Existing OAuth scopes include `chat:read`: `backend/internal/handlers/twitch_oauth_handler.go`.
- Existing websocket patterns: `backend/internal/websocket/server.go`, `backend/internal/services/watch_party_hub.go`.
- Clip lookup by Twitch ID exists: `backend/internal/repository/clip_repository.go` has `GetByTwitchClipID` and `GetByTwitchClipIDs`.

## Product Decisions for v1

1. **Capture mode:** backend bot/listener monitors Twitch chat server-side.
2. **Approval mode:** approval required by default. Auto-approve can be a room setting, but default remains manual.
3. **Supported links in v1:**
   - Clpr clip detail links containing `/clip/:id` or `/clips/:id`.
   - Twitch clip links when the Twitch clip slug already exists in Clpr via `clips.twitch_clip_id`.
4. **Unsupported Twitch links:** persist as `skipped` with reason `clip_not_imported`, so streamers can see that the bot caught the link.
5. **Approvers:** room owner only in v1. Moderator approvals can be added later using existing moderation scope concepts.
6. **Playback:** approved items are ordered independently from personal queue and playlist tables.
7. **Discussion:** reuse current clip comments via `CommentSection` in the theatre sidebar.

## File Structure

### Backend files to create

- `backend/migrations/000117_create_streamer_clip_rooms.up.sql` — create room/item tables, indexes, constraints.
- `backend/migrations/000117_create_streamer_clip_rooms.down.sql` — drop tables in reverse order.
- `backend/internal/models/streamer_clip_room.go` — room/item structs, statuses, request/response types.
- `backend/internal/repository/streamer_clip_room_repository.go` — pgx persistence and ordering operations.
- `backend/internal/services/clip_link_parser.go` — pure URL parsing for Clpr and Twitch clip links.
- `backend/internal/services/clip_link_parser_test.go` — parser unit tests.
- `backend/internal/services/streamer_clip_room_service.go` — business rules for room lifecycle, message ingestion, approve/reject/reorder.
- `backend/internal/services/streamer_clip_room_service_test.go` — service tests with fakes where practical.
- `backend/internal/services/twitch_chat_listener.go` — IRC-over-WebSocket chat listener abstraction and manager.
- `backend/internal/services/twitch_chat_listener_test.go` — message parsing/reconnect decision tests.
- `backend/internal/handlers/streamer_clip_room_handler.go` — REST + websocket HTTP handlers.

### Backend files to modify

- `backend/cmd/api/main.go` or existing service wiring file — instantiate repository/service/handler.
- `backend/cmd/api/routes_platform.go` — register `/streamer-clip-rooms` routes.
- `backend/internal/repository/clip_repository.go` — add `GetByIDString` only if handler/service needs a helper; prefer parsing UUID in service and using existing `GetByID`.
- `backend/internal/handlers/twitch_oauth_handler.go` — no scope change needed for v1 because `chat:read` already exists; update comments if needed.

### Frontend files to create

- `frontend/src/types/streamerClipRoom.ts` — frontend type contract.
- `frontend/src/lib/streamer-clip-room-api.ts` — API client functions.
- `frontend/src/hooks/useStreamerClipRoom.ts` — React Query hooks and mutations.
- `frontend/src/hooks/useStreamerClipRoomWebSocket.ts` — websocket update hook.
- `frontend/src/pages/StreamerClipRoomPage.tsx` — streamer-facing theatre page.
- `frontend/src/pages/StreamerClipRoomPage.test.tsx` — page behavior tests.

### Frontend files to modify

- `frontend/src/App.tsx` — add protected route.
- `frontend/src/pages/index.ts` — export `StreamerClipRoomPage`.
- `frontend/src/components/playlist/PlaylistTheatreMode.tsx` — add optional labels and optional custom sidebar tab content while preserving current playlist/queue behavior.
- `frontend/src/components/stream/` or `frontend/src/pages/StreamPage.tsx` — add navigation entry from streamer page for authenticated users.

---

## Backend API Contract

All endpoints require auth unless explicitly marked optional.

```http
GET    /api/v1/streamer-clip-rooms/:channel
POST   /api/v1/streamer-clip-rooms/:channel/start
POST   /api/v1/streamer-clip-rooms/:channel/stop
GET    /api/v1/streamer-clip-rooms/:roomId/items?status=pending|approved|rejected|skipped|all
POST   /api/v1/streamer-clip-rooms/:roomId/items/:itemId/approve
POST   /api/v1/streamer-clip-rooms/:roomId/items/:itemId/reject
PUT    /api/v1/streamer-clip-rooms/:roomId/items/order
GET    /api/v1/streamer-clip-rooms/:roomId/ws
```

### Response shapes

```json
{
  "success": true,
  "data": {
    "room": {
      "id": "uuid",
      "owner_user_id": "uuid",
      "twitch_channel": "moonmoon",
      "approval_mode": "manual",
      "is_active": true,
      "created_at": "2026-06-06T00:00:00Z",
      "updated_at": "2026-06-06T00:00:00Z"
    },
    "items": []
  }
}
```

```json
{
  "id": "uuid",
  "room_id": "uuid",
  "clip_id": "uuid-or-null",
  "clip": null,
  "source_url": "https://clpr.tv/clip/uuid",
  "source_type": "clpr",
  "status": "pending",
  "position": null,
  "twitch_message_id": "abc",
  "twitch_user_id": "123",
  "twitch_username": "viewer",
  "message_text": "watch this https://clpr.tv/clip/...",
  "skip_reason": null,
  "detected_at": "2026-06-06T00:00:00Z",
  "approved_at": null,
  "rejected_at": null
}
```

---

## Task 1: Create Database Tables

**Files:**
- Create: `backend/migrations/000117_create_streamer_clip_rooms.up.sql`
- Create: `backend/migrations/000117_create_streamer_clip_rooms.down.sql`

- [ ] **Step 1: Write the migration**

Create `backend/migrations/000117_create_streamer_clip_rooms.up.sql`:

```sql
CREATE TABLE streamer_clip_rooms (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    owner_user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    twitch_channel TEXT NOT NULL,
    approval_mode TEXT NOT NULL DEFAULT 'manual' CHECK (approval_mode IN ('manual', 'auto')),
    is_active BOOLEAN NOT NULL DEFAULT false,
    last_listener_error TEXT,
    listener_started_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (owner_user_id, lower(twitch_channel))
);

CREATE TABLE streamer_clip_room_items (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    room_id UUID NOT NULL REFERENCES streamer_clip_rooms(id) ON DELETE CASCADE,
    clip_id UUID REFERENCES clips(id) ON DELETE SET NULL,
    source_url TEXT NOT NULL,
    source_type TEXT NOT NULL CHECK (source_type IN ('clpr', 'twitch')),
    status TEXT NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'approved', 'rejected', 'skipped')),
    position INTEGER,
    twitch_message_id TEXT,
    twitch_user_id TEXT,
    twitch_username TEXT,
    message_text TEXT,
    skip_reason TEXT,
    detected_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    approved_at TIMESTAMPTZ,
    approved_by_user_id UUID REFERENCES users(id) ON DELETE SET NULL,
    rejected_at TIMESTAMPTZ,
    rejected_by_user_id UUID REFERENCES users(id) ON DELETE SET NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT streamer_clip_room_items_approved_position CHECK (
        (status = 'approved' AND position IS NOT NULL)
        OR (status <> 'approved')
    )
);

CREATE UNIQUE INDEX streamer_clip_rooms_owner_channel_idx
    ON streamer_clip_rooms (owner_user_id, lower(twitch_channel));

CREATE INDEX streamer_clip_room_items_room_status_position_idx
    ON streamer_clip_room_items (room_id, status, position NULLS LAST, detected_at DESC);

CREATE INDEX streamer_clip_room_items_room_detected_idx
    ON streamer_clip_room_items (room_id, detected_at DESC);

CREATE UNIQUE INDEX streamer_clip_room_items_message_url_idx
    ON streamer_clip_room_items (room_id, twitch_message_id, source_url)
    WHERE twitch_message_id IS NOT NULL;

CREATE UNIQUE INDEX streamer_clip_room_items_approved_position_idx
    ON streamer_clip_room_items (room_id, position)
    WHERE status = 'approved';
```

- [ ] **Step 2: Write rollback migration**

Create `backend/migrations/000117_create_streamer_clip_rooms.down.sql`:

```sql
DROP INDEX IF EXISTS streamer_clip_room_items_approved_position_idx;
DROP INDEX IF EXISTS streamer_clip_room_items_message_url_idx;
DROP INDEX IF EXISTS streamer_clip_room_items_room_detected_idx;
DROP INDEX IF EXISTS streamer_clip_room_items_room_status_position_idx;
DROP INDEX IF EXISTS streamer_clip_rooms_owner_channel_idx;
DROP TABLE IF EXISTS streamer_clip_room_items;
DROP TABLE IF EXISTS streamer_clip_rooms;
```

- [ ] **Step 3: Validate migration syntax**

Run from worktree root:

```bash
go test ./backend/tests/migrations/...
```

Expected: migration tests pass or, if local DB is not configured, fail only with the repository's existing test DB connection error.

- [ ] **Step 4: Commit**

```bash
git add backend/migrations/000117_create_streamer_clip_rooms.up.sql backend/migrations/000117_create_streamer_clip_rooms.down.sql
git commit -m "feat: add streamer clip room tables"
```

---

## Task 2: Add Backend Models

**Files:**
- Create: `backend/internal/models/streamer_clip_room.go`

- [ ] **Step 1: Add model types**

Create `backend/internal/models/streamer_clip_room.go`:

```go
package models

import (
    "time"

    "github.com/google/uuid"
)

const (
    StreamerClipRoomApprovalManual = "manual"
    StreamerClipRoomApprovalAuto   = "auto"

    StreamerClipRoomItemStatusPending  = "pending"
    StreamerClipRoomItemStatusApproved = "approved"
    StreamerClipRoomItemStatusRejected = "rejected"
    StreamerClipRoomItemStatusSkipped  = "skipped"

    StreamerClipRoomSourceClpr   = "clpr"
    StreamerClipRoomSourceTwitch = "twitch"
)

type StreamerClipRoom struct {
    ID                uuid.UUID  `json:"id" db:"id"`
    OwnerUserID       uuid.UUID  `json:"owner_user_id" db:"owner_user_id"`
    TwitchChannel     string     `json:"twitch_channel" db:"twitch_channel"`
    ApprovalMode      string     `json:"approval_mode" db:"approval_mode"`
    IsActive          bool       `json:"is_active" db:"is_active"`
    LastListenerError *string    `json:"last_listener_error,omitempty" db:"last_listener_error"`
    ListenerStartedAt *time.Time `json:"listener_started_at,omitempty" db:"listener_started_at"`
    CreatedAt         time.Time  `json:"created_at" db:"created_at"`
    UpdatedAt         time.Time  `json:"updated_at" db:"updated_at"`
}

type StreamerClipRoomItem struct {
    ID               uuid.UUID  `json:"id" db:"id"`
    RoomID           uuid.UUID  `json:"room_id" db:"room_id"`
    ClipID           *uuid.UUID `json:"clip_id,omitempty" db:"clip_id"`
    Clip             *Clip      `json:"clip,omitempty" db:"-"`
    SourceURL        string     `json:"source_url" db:"source_url"`
    SourceType       string     `json:"source_type" db:"source_type"`
    Status           string     `json:"status" db:"status"`
    Position         *int       `json:"position,omitempty" db:"position"`
    TwitchMessageID  *string    `json:"twitch_message_id,omitempty" db:"twitch_message_id"`
    TwitchUserID     *string    `json:"twitch_user_id,omitempty" db:"twitch_user_id"`
    TwitchUsername   *string    `json:"twitch_username,omitempty" db:"twitch_username"`
    MessageText      *string    `json:"message_text,omitempty" db:"message_text"`
    SkipReason       *string    `json:"skip_reason,omitempty" db:"skip_reason"`
    DetectedAt       time.Time  `json:"detected_at" db:"detected_at"`
    ApprovedAt       *time.Time `json:"approved_at,omitempty" db:"approved_at"`
    ApprovedByUserID *uuid.UUID `json:"approved_by_user_id,omitempty" db:"approved_by_user_id"`
    RejectedAt       *time.Time `json:"rejected_at,omitempty" db:"rejected_at"`
    RejectedByUserID *uuid.UUID `json:"rejected_by_user_id,omitempty" db:"rejected_by_user_id"`
    CreatedAt        time.Time  `json:"created_at" db:"created_at"`
    UpdatedAt        time.Time  `json:"updated_at" db:"updated_at"`
}

type StreamerClipRoomWithItems struct {
    Room          StreamerClipRoom       `json:"room"`
    PendingItems  []StreamerClipRoomItem `json:"pending_items"`
    ApprovedItems []StreamerClipRoomItem `json:"approved_items"`
    SkippedItems  []StreamerClipRoomItem `json:"skipped_items"`
}

type ReorderStreamerClipRoomItemsRequest struct {
    ItemIDs []string `json:"item_ids" binding:"required,min=1"`
}

type StreamerClipRoomEvent struct {
    Type string                 `json:"type"`
    Data map[string]interface{} `json:"data"`
}
```

- [ ] **Step 2: Run model compile check**

```bash
go test ./backend/internal/models
```

Expected: package compiles and tests pass.

- [ ] **Step 3: Commit**

```bash
git add backend/internal/models/streamer_clip_room.go
git commit -m "feat: add streamer clip room models"
```

---

## Task 3: Implement Clip Link Parser

**Files:**
- Create: `backend/internal/services/clip_link_parser.go`
- Create: `backend/internal/services/clip_link_parser_test.go`

- [ ] **Step 1: Write parser tests**

Create `backend/internal/services/clip_link_parser_test.go`:

```go
package services

import "testing"

func TestExtractClipLinks(t *testing.T) {
    input := "watch https://clpr.tv/clip/123e4567-e89b-12d3-a456-426614174000 and https://clips.twitch.tv/FunnySlug"
    links := ExtractClipLinks(input)
    if len(links) != 2 {
        t.Fatalf("expected 2 links, got %d", len(links))
    }
    if links[0].SourceType != "clpr" || links[0].ClprClipID != "123e4567-e89b-12d3-a456-426614174000" {
        t.Fatalf("unexpected first link: %#v", links[0])
    }
    if links[1].SourceType != "twitch" || links[1].TwitchClipID != "FunnySlug" {
        t.Fatalf("unexpected second link: %#v", links[1])
    }
}

func TestExtractClipLinksDeduplicates(t *testing.T) {
    input := "https://clips.twitch.tv/FunnySlug https://clips.twitch.tv/FunnySlug"
    links := ExtractClipLinks(input)
    if len(links) != 1 {
        t.Fatalf("expected 1 link, got %d", len(links))
    }
}
```

- [ ] **Step 2: Run tests and verify failure**

```bash
go test ./backend/internal/services -run TestExtractClipLinks -count=1
```

Expected: FAIL because `ExtractClipLinks` is not defined.

- [ ] **Step 3: Implement parser**

Create `backend/internal/services/clip_link_parser.go`:

```go
package services

import (
    "net/url"
    "regexp"
    "strings"
)

type DetectedClipLink struct {
    SourceURL    string
    SourceType   string
    ClprClipID   string
    TwitchClipID string
}

var httpURLPattern = regexp.MustCompile(`https?://[^\s<>()]+`)

func ExtractClipLinks(message string) []DetectedClipLink {
    matches := httpURLPattern.FindAllString(message, -1)
    seen := map[string]bool{}
    links := make([]DetectedClipLink, 0, len(matches))

    for _, raw := range matches {
        cleaned := strings.TrimRight(raw, ".,!?)]}")
        if seen[cleaned] {
            continue
        }
        parsed, err := url.Parse(cleaned)
        if err != nil || parsed.Host == "" {
            continue
        }

        host := strings.ToLower(parsed.Host)
        path := strings.Trim(parsed.Path, "/")
        parts := strings.Split(path, "/")

        if (host == "clpr.tv" || strings.HasSuffix(host, ".clpr.tv")) && len(parts) == 2 && (parts[0] == "clip" || parts[0] == "clips") {
            seen[cleaned] = true
            links = append(links, DetectedClipLink{SourceURL: cleaned, SourceType: "clpr", ClprClipID: parts[1]})
            continue
        }

        if host == "clips.twitch.tv" && len(parts) >= 1 && parts[0] != "" {
            seen[cleaned] = true
            links = append(links, DetectedClipLink{SourceURL: cleaned, SourceType: "twitch", TwitchClipID: parts[0]})
            continue
        }

        if (host == "www.twitch.tv" || host == "twitch.tv") && len(parts) >= 3 && parts[1] == "clip" {
            seen[cleaned] = true
            links = append(links, DetectedClipLink{SourceURL: cleaned, SourceType: "twitch", TwitchClipID: parts[2]})
        }
    }

    return links
}
```

- [ ] **Step 4: Run parser tests**

```bash
go test ./backend/internal/services -run TestExtractClipLinks -count=1
```

Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add backend/internal/services/clip_link_parser.go backend/internal/services/clip_link_parser_test.go
git commit -m "feat: parse streamer clip links"
```

---

## Task 4: Add Repository Layer

**Files:**
- Create: `backend/internal/repository/streamer_clip_room_repository.go`
- Test: use service-level tests with fake repository first; add integration tests only if repository test harness exists for similar tables.

- [ ] **Step 1: Implement repository skeleton and core methods**

Create `backend/internal/repository/streamer_clip_room_repository.go` with these exact public methods:

```go
package repository

import (
    "context"
    "fmt"
    "strings"

    "github.com/google/uuid"
    "github.com/jackc/pgx/v5"
    "github.com/jackc/pgx/v5/pgxpool"
    "github.com/subculture-collective/clipper/internal/models"
)

type StreamerClipRoomRepository struct {
    pool *pgxpool.Pool
}

func NewStreamerClipRoomRepository(pool *pgxpool.Pool) *StreamerClipRoomRepository {
    return &StreamerClipRoomRepository{pool: pool}
}

func (r *StreamerClipRoomRepository) GetOrCreateRoom(ctx context.Context, ownerUserID uuid.UUID, channel string) (*models.StreamerClipRoom, error) {
    normalized := strings.ToLower(strings.TrimSpace(channel))
    var room models.StreamerClipRoom
    err := r.pool.QueryRow(ctx, `
        INSERT INTO streamer_clip_rooms (owner_user_id, twitch_channel)
        VALUES ($1, $2)
        ON CONFLICT (owner_user_id, lower(twitch_channel)) DO UPDATE
            SET updated_at = NOW()
        RETURNING id, owner_user_id, twitch_channel, approval_mode, is_active,
            last_listener_error, listener_started_at, created_at, updated_at
    `, ownerUserID, normalized).Scan(
        &room.ID, &room.OwnerUserID, &room.TwitchChannel, &room.ApprovalMode, &room.IsActive,
        &room.LastListenerError, &room.ListenerStartedAt, &room.CreatedAt, &room.UpdatedAt,
    )
    if err != nil {
        return nil, fmt.Errorf("failed to get or create streamer clip room: %w", err)
    }
    return &room, nil
}

func (r *StreamerClipRoomRepository) GetRoomByID(ctx context.Context, roomID uuid.UUID) (*models.StreamerClipRoom, error) {
    var room models.StreamerClipRoom
    err := r.pool.QueryRow(ctx, `
        SELECT id, owner_user_id, twitch_channel, approval_mode, is_active,
            last_listener_error, listener_started_at, created_at, updated_at
        FROM streamer_clip_rooms
        WHERE id = $1
    `, roomID).Scan(&room.ID, &room.OwnerUserID, &room.TwitchChannel, &room.ApprovalMode, &room.IsActive,
        &room.LastListenerError, &room.ListenerStartedAt, &room.CreatedAt, &room.UpdatedAt)
    if err == pgx.ErrNoRows {
        return nil, nil
    }
    if err != nil {
        return nil, fmt.Errorf("failed to get streamer clip room: %w", err)
    }
    return &room, nil
}
```

Add methods with signatures used by the service:

```go
func (r *StreamerClipRoomRepository) SetRoomActive(ctx context.Context, roomID uuid.UUID, active bool, listenerError *string) error
func (r *StreamerClipRoomRepository) CreateItem(ctx context.Context, item *models.StreamerClipRoomItem) error
func (r *StreamerClipRoomRepository) ListItems(ctx context.Context, roomID uuid.UUID, status string, limit int) ([]models.StreamerClipRoomItem, error)
func (r *StreamerClipRoomRepository) GetItem(ctx context.Context, roomID, itemID uuid.UUID) (*models.StreamerClipRoomItem, error)
func (r *StreamerClipRoomRepository) ApproveItem(ctx context.Context, roomID, itemID, approverID uuid.UUID) (*models.StreamerClipRoomItem, error)
func (r *StreamerClipRoomRepository) RejectItem(ctx context.Context, roomID, itemID, rejecterID uuid.UUID) (*models.StreamerClipRoomItem, error)
func (r *StreamerClipRoomRepository) ReorderApprovedItems(ctx context.Context, roomID uuid.UUID, itemIDs []uuid.UUID) error
```

Implementation details:
- `ApproveItem` must run in a transaction, lock the item row, assign `position = COALESCE(MAX(position), 0)+1`, set `status='approved'`, `approved_at=NOW()`, and `approved_by_user_id`.
- `RejectItem` must set `status='rejected'`, clear `position`, and set rejection metadata.
- `ReorderApprovedItems` must verify all provided item IDs belong to the room and are approved, then temporarily offset positions by `+100000` before writing final positions `1..n` to avoid unique index collisions.

- [ ] **Step 2: Run repository compile check**

```bash
go test ./backend/internal/repository -run TestNonExistent -count=1
```

Expected: package compiles; no tests run is acceptable.

- [ ] **Step 3: Commit**

```bash
git add backend/internal/repository/streamer_clip_room_repository.go
git commit -m "feat: add streamer clip room repository"
```

---

## Task 5: Add Service Layer and Ingestion Rules

**Files:**
- Create: `backend/internal/services/streamer_clip_room_service.go`
- Create: `backend/internal/services/streamer_clip_room_service_test.go`

- [ ] **Step 1: Write service tests for ingest behavior**

Create `backend/internal/services/streamer_clip_room_service_test.go` with tests for:

```go
func TestStreamerClipRoomServiceIngestsClprClipAsPending(t *testing.T)
func TestStreamerClipRoomServiceSkipsUnknownTwitchClip(t *testing.T)
func TestStreamerClipRoomServiceAutoApprovesWhenConfigured(t *testing.T)
func TestStreamerClipRoomServiceRejectsNonOwnerApproval(t *testing.T)
```

Use fake repository and fake clip lookup interfaces in the test file. The first test must assert a Clpr URL with an existing clip creates an item with `Status == models.StreamerClipRoomItemStatusPending` and the correct `ClipID`. The second must assert an unknown Twitch slug creates an item with `Status == models.StreamerClipRoomItemStatusSkipped` and `SkipReason == "clip_not_imported"`.

- [ ] **Step 2: Run tests and verify failure**

```bash
go test ./backend/internal/services -run TestStreamerClipRoomService -count=1
```

Expected: FAIL because the service is not implemented.

- [ ] **Step 3: Implement service interfaces and methods**

Create `backend/internal/services/streamer_clip_room_service.go` with public constructor and methods:

```go
package services

import (
    "context"
    "errors"
    "fmt"

    "github.com/google/uuid"
    "github.com/subculture-collective/clipper/internal/models"
    "github.com/subculture-collective/clipper/internal/repository"
)

var ErrStreamerClipRoomForbidden = errors.New("streamer clip room forbidden")

type StreamerClipRoomService struct {
    rooms *repository.StreamerClipRoomRepository
    clips *repository.ClipRepository
}

func NewStreamerClipRoomService(rooms *repository.StreamerClipRoomRepository, clips *repository.ClipRepository) *StreamerClipRoomService {
    return &StreamerClipRoomService{rooms: rooms, clips: clips}
}

type TwitchChatClipMessage struct {
    RoomID          uuid.UUID
    TwitchMessageID string
    TwitchUserID    string
    TwitchUsername  string
    MessageText     string
}
```

Add methods:

```go
func (s *StreamerClipRoomService) GetOrCreateRoom(ctx context.Context, ownerUserID uuid.UUID, channel string) (*models.StreamerClipRoomWithItems, error)
func (s *StreamerClipRoomService) StartRoom(ctx context.Context, ownerUserID uuid.UUID, channel string) (*models.StreamerClipRoom, error)
func (s *StreamerClipRoomService) StopRoom(ctx context.Context, ownerUserID uuid.UUID, channel string) (*models.StreamerClipRoom, error)
func (s *StreamerClipRoomService) IngestChatMessage(ctx context.Context, msg TwitchChatClipMessage) ([]models.StreamerClipRoomItem, error)
func (s *StreamerClipRoomService) ApproveItem(ctx context.Context, actorUserID, roomID, itemID uuid.UUID) (*models.StreamerClipRoomItem, error)
func (s *StreamerClipRoomService) RejectItem(ctx context.Context, actorUserID, roomID, itemID uuid.UUID) (*models.StreamerClipRoomItem, error)
func (s *StreamerClipRoomService) ReorderApprovedItems(ctx context.Context, actorUserID, roomID uuid.UUID, itemIDs []uuid.UUID) error
```

Ownership check used by approve/reject/reorder:

```go
func (s *StreamerClipRoomService) requireOwner(ctx context.Context, actorUserID, roomID uuid.UUID) (*models.StreamerClipRoom, error) {
    room, err := s.rooms.GetRoomByID(ctx, roomID)
    if err != nil {
        return nil, err
    }
    if room == nil || room.OwnerUserID != actorUserID {
        return nil, ErrStreamerClipRoomForbidden
    }
    return room, nil
}
```

Ingestion rules:
- Run `ExtractClipLinks(msg.MessageText)`.
- Resolve Clpr links by UUID using `clips.GetByID`.
- Resolve Twitch links using `clips.GetByTwitchClipID`.
- If no clip is found, create a skipped item with `skip_reason='clip_not_imported'`.
- If clip is removed/hidden/DMCA removed, create skipped item with `skip_reason='clip_unavailable'`.
- If room approval mode is `auto`, create pending then approve in the same service call or insert directly as approved with repository support. Prefer create pending then approve to keep logic consistent.

- [ ] **Step 4: Run service tests**

```bash
go test ./backend/internal/services -run TestStreamerClipRoomService -count=1
```

Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add backend/internal/services/streamer_clip_room_service.go backend/internal/services/streamer_clip_room_service_test.go
git commit -m "feat: add streamer clip room service"
```

---

## Task 6: Add Twitch Chat Listener Manager

**Files:**
- Create: `backend/internal/services/twitch_chat_listener.go`
- Create: `backend/internal/services/twitch_chat_listener_test.go`

- [ ] **Step 1: Write message parsing tests**

Create `backend/internal/services/twitch_chat_listener_test.go` with a test that parses IRC tags and PRIVMSG text:

```go
package services

import "testing"

func TestParseTwitchIRCPrivmsg(t *testing.T) {
    raw := "@id=msg-1;user-id=42;display-name=Viewer :viewer!viewer@viewer.tmi.twitch.tv PRIVMSG #moonmoon :watch https://clpr.tv/clip/123e4567-e89b-12d3-a456-426614174000"
    msg, ok := ParseTwitchIRCPrivmsg(raw)
    if !ok {
        t.Fatal("expected PRIVMSG")
    }
    if msg.MessageID != "msg-1" || msg.UserID != "42" || msg.Username != "Viewer" || msg.Channel != "moonmoon" {
        t.Fatalf("unexpected parsed message: %#v", msg)
    }
}
```

- [ ] **Step 2: Implement listener parser and manager interface**

Create `backend/internal/services/twitch_chat_listener.go`:

```go
package services

import (
    "context"
    "strings"
    "sync"

    "github.com/google/uuid"
)

type TwitchIRCMessage struct {
    MessageID string
    UserID    string
    Username  string
    Channel   string
    Text      string
}

type TwitchChatListenerManager struct {
    mu        sync.Mutex
    listeners map[uuid.UUID]context.CancelFunc
    service   *StreamerClipRoomService
}

func NewTwitchChatListenerManager(service *StreamerClipRoomService) *TwitchChatListenerManager {
    return &TwitchChatListenerManager{listeners: map[uuid.UUID]context.CancelFunc{}, service: service}
}
```

Implement `ParseTwitchIRCPrivmsg(raw string) (TwitchIRCMessage, bool)` by:
- Returning false unless raw contains ` PRIVMSG `.
- Reading tags before the first space when raw begins with `@`.
- Extracting `id`, `user-id`, and `display-name` from semicolon-separated tags.
- Extracting channel after ` PRIVMSG #` until ` :`.
- Extracting message text after ` :`.

Add manager methods:

```go
func (m *TwitchChatListenerManager) Start(ctx context.Context, roomID uuid.UUID, channel string, oauthToken string) error
func (m *TwitchChatListenerManager) Stop(roomID uuid.UUID)
func (m *TwitchChatListenerManager) IsRunning(roomID uuid.UUID) bool
```

For v1 implementation, `Start` should create a cancellable goroutine that connects to Twitch IRC over TLS or WebSocket, joins the channel, reads `PRIVMSG`, calls `service.IngestChatMessage`, and exits on context cancellation. Use Twitch's documented IRC endpoint `irc.chat.twitch.tv:6697` with PASS `oauth:<token>` and NICK from OAuth username when available. If username is not available in this manager signature, extend `Start` to accept `username`.

- [ ] **Step 3: Run listener tests**

```bash
go test ./backend/internal/services -run TestParseTwitchIRCPrivmsg -count=1
```

Expected: PASS.

- [ ] **Step 4: Commit**

```bash
git add backend/internal/services/twitch_chat_listener.go backend/internal/services/twitch_chat_listener_test.go
git commit -m "feat: add twitch chat listener manager"
```

---

## Task 7: Add Backend Handler and Routes

**Files:**
- Create: `backend/internal/handlers/streamer_clip_room_handler.go`
- Modify: `backend/cmd/api/routes_platform.go`
- Modify: backend service wiring file that defines `Handlers` and `Services` structs.

- [ ] **Step 1: Add handler**

Create `backend/internal/handlers/streamer_clip_room_handler.go` with handler methods:

```go
package handlers

import (
    "net/http"

    "github.com/gin-gonic/gin"
    "github.com/google/uuid"
    "github.com/subculture-collective/clipper/internal/models"
    "github.com/subculture-collective/clipper/internal/services"
)

type StreamerClipRoomHandler struct {
    service *services.StreamerClipRoomService
}

func NewStreamerClipRoomHandler(service *services.StreamerClipRoomService) *StreamerClipRoomHandler {
    return &StreamerClipRoomHandler{service: service}
}
```

Add methods:

```go
func (h *StreamerClipRoomHandler) GetRoom(c *gin.Context)
func (h *StreamerClipRoomHandler) StartRoom(c *gin.Context)
func (h *StreamerClipRoomHandler) StopRoom(c *gin.Context)
func (h *StreamerClipRoomHandler) ListItems(c *gin.Context)
func (h *StreamerClipRoomHandler) ApproveItem(c *gin.Context)
func (h *StreamerClipRoomHandler) RejectItem(c *gin.Context)
func (h *StreamerClipRoomHandler) ReorderItems(c *gin.Context)
func (h *StreamerClipRoomHandler) WebSocket(c *gin.Context)
```

Use the existing response shape:

```go
c.JSON(http.StatusOK, StandardResponse{Success: true, Data: data})
```

For `ReorderItems`, bind `models.ReorderStreamerClipRoomItemsRequest`, parse UUIDs in service, and return `{"message":"Items reordered successfully"}`.

- [ ] **Step 2: Register routes**

Modify `backend/cmd/api/routes_platform.go` after Twitch routes or near stream routes:

```go
if h.StreamerClipRoom != nil {
    streamerClipRooms := v1.Group("/streamer-clip-rooms")
    streamerClipRooms.Use(middleware.AuthMiddleware(svcs.Auth))
    {
        streamerClipRooms.GET("/:channel", h.StreamerClipRoom.GetRoom)
        streamerClipRooms.POST("/:channel/start", middleware.RateLimitMiddleware(infra.Redis, 10, time.Minute), h.StreamerClipRoom.StartRoom)
        streamerClipRooms.POST("/:channel/stop", h.StreamerClipRoom.StopRoom)
        streamerClipRooms.GET("/:roomId/items", h.StreamerClipRoom.ListItems)
        streamerClipRooms.POST("/:roomId/items/:itemId/approve", middleware.RateLimitMiddleware(infra.Redis, 120, time.Minute), h.StreamerClipRoom.ApproveItem)
        streamerClipRooms.POST("/:roomId/items/:itemId/reject", middleware.RateLimitMiddleware(infra.Redis, 120, time.Minute), h.StreamerClipRoom.RejectItem)
        streamerClipRooms.PUT("/:roomId/items/order", h.StreamerClipRoom.ReorderItems)
        streamerClipRooms.GET("/:roomId/ws", h.StreamerClipRoom.WebSocket)
    }
}
```

Note: `/:channel` and `/:roomId/items` are unambiguous because the latter has extra path segments.

- [ ] **Step 3: Wire service and handler**

Find the file that constructs `Services` and `Handlers` in `backend/cmd/api`. Add fields:

```go
StreamerClipRoom *services.StreamerClipRoomService
```

and:

```go
StreamerClipRoom *handlers.StreamerClipRoomHandler
```

Instantiate:

```go
streamerClipRoomRepo := repository.NewStreamerClipRoomRepository(infra.DB)
streamerClipRoomService := services.NewStreamerClipRoomService(streamerClipRoomRepo, clipRepo)
streamerClipRoomHandler := handlers.NewStreamerClipRoomHandler(streamerClipRoomService)
```

- [ ] **Step 4: Compile backend**

```bash
go test ./backend/cmd/api ./backend/internal/handlers -run TestNonExistent -count=1
```

Expected: packages compile; no tests run is acceptable.

- [ ] **Step 5: Commit**

```bash
git add backend/internal/handlers/streamer_clip_room_handler.go backend/cmd/api
git commit -m "feat: expose streamer clip room API"
```

---

## Task 8: Add Frontend Types, API Client, and Hooks

**Files:**
- Create: `frontend/src/types/streamerClipRoom.ts`
- Create: `frontend/src/lib/streamer-clip-room-api.ts`
- Create: `frontend/src/hooks/useStreamerClipRoom.ts`
- Create: `frontend/src/hooks/useStreamerClipRoomWebSocket.ts`

- [ ] **Step 1: Add types**

Create `frontend/src/types/streamerClipRoom.ts`:

```ts
import type { Clip } from './clip';

export type StreamerClipRoomApprovalMode = 'manual' | 'auto';
export type StreamerClipRoomItemStatus = 'pending' | 'approved' | 'rejected' | 'skipped';
export type StreamerClipRoomSourceType = 'clpr' | 'twitch';

export interface StreamerClipRoom {
  id: string;
  owner_user_id: string;
  twitch_channel: string;
  approval_mode: StreamerClipRoomApprovalMode;
  is_active: boolean;
  last_listener_error?: string;
  listener_started_at?: string;
  created_at: string;
  updated_at: string;
}

export interface StreamerClipRoomItem {
  id: string;
  room_id: string;
  clip_id?: string;
  clip?: Clip;
  source_url: string;
  source_type: StreamerClipRoomSourceType;
  status: StreamerClipRoomItemStatus;
  position?: number;
  twitch_message_id?: string;
  twitch_user_id?: string;
  twitch_username?: string;
  message_text?: string;
  skip_reason?: string;
  detected_at: string;
  approved_at?: string;
  rejected_at?: string;
}

export interface StreamerClipRoomWithItems {
  room: StreamerClipRoom;
  pending_items: StreamerClipRoomItem[];
  approved_items: StreamerClipRoomItem[];
  skipped_items: StreamerClipRoomItem[];
}

export interface StreamerClipRoomEvent {
  type: 'item_detected' | 'item_approved' | 'item_rejected' | 'items_reordered' | 'room_status_changed';
  data: Record<string, unknown>;
}
```

- [ ] **Step 2: Add API client**

Create `frontend/src/lib/streamer-clip-room-api.ts`:

```ts
import apiClient from './api';
import type { StreamerClipRoomItem, StreamerClipRoomWithItems } from '@/types/streamerClipRoom';

interface StandardResponse<T> {
  success: boolean;
  data: T;
}

export const streamerClipRoomApi = {
  async getRoom(channel: string): Promise<StreamerClipRoomWithItems> {
    const response = await apiClient.get<StandardResponse<StreamerClipRoomWithItems>>(`/streamer-clip-rooms/${encodeURIComponent(channel)}`);
    return response.data.data;
  },

  async startRoom(channel: string): Promise<void> {
    await apiClient.post(`/streamer-clip-rooms/${encodeURIComponent(channel)}/start`);
  },

  async stopRoom(channel: string): Promise<void> {
    await apiClient.post(`/streamer-clip-rooms/${encodeURIComponent(channel)}/stop`);
  },

  async approveItem(roomId: string, itemId: string): Promise<StreamerClipRoomItem> {
    const response = await apiClient.post<StandardResponse<StreamerClipRoomItem>>(`/streamer-clip-rooms/${roomId}/items/${itemId}/approve`);
    return response.data.data;
  },

  async rejectItem(roomId: string, itemId: string): Promise<StreamerClipRoomItem> {
    const response = await apiClient.post<StandardResponse<StreamerClipRoomItem>>(`/streamer-clip-rooms/${roomId}/items/${itemId}/reject`);
    return response.data.data;
  },

  async reorderItems(roomId: string, itemIds: string[]): Promise<void> {
    await apiClient.put(`/streamer-clip-rooms/${roomId}/items/order`, { item_ids: itemIds });
  },
};
```

- [ ] **Step 3: Add React Query hooks**

Create `frontend/src/hooks/useStreamerClipRoom.ts` with:

```ts
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { streamerClipRoomApi } from '@/lib/streamer-clip-room-api';

export function useStreamerClipRoom(channel: string) {
  return useQuery({
    queryKey: ['streamer-clip-room', channel],
    queryFn: () => streamerClipRoomApi.getRoom(channel),
    enabled: !!channel,
  });
}

export function useStartStreamerClipRoom(channel: string) {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: () => streamerClipRoomApi.startRoom(channel),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ['streamer-clip-room', channel] }),
  });
}

export function useStopStreamerClipRoom(channel: string) {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: () => streamerClipRoomApi.stopRoom(channel),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ['streamer-clip-room', channel] }),
  });
}

export function useApproveStreamerClipRoomItem(channel: string, roomId?: string) {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (itemId: string) => streamerClipRoomApi.approveItem(roomId!, itemId),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ['streamer-clip-room', channel] }),
  });
}

export function useRejectStreamerClipRoomItem(channel: string, roomId?: string) {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (itemId: string) => streamerClipRoomApi.rejectItem(roomId!, itemId),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ['streamer-clip-room', channel] }),
  });
}

export function useReorderStreamerClipRoomItems(channel: string, roomId?: string) {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (itemIds: string[]) => streamerClipRoomApi.reorderItems(roomId!, itemIds),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ['streamer-clip-room', channel] }),
  });
}
```

- [ ] **Step 4: Add websocket invalidation hook**

Create `frontend/src/hooks/useStreamerClipRoomWebSocket.ts`:

```ts
import { useEffect } from 'react';
import { useQueryClient } from '@tanstack/react-query';

export function useStreamerClipRoomWebSocket(channel: string, roomId?: string) {
  const queryClient = useQueryClient();

  useEffect(() => {
    if (!roomId) return;

    const wsProtocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
    const wsHost = import.meta.env.VITE_WS_HOST || window.location.host;
    const ws = new WebSocket(`${wsProtocol}//${wsHost}/api/v1/streamer-clip-rooms/${roomId}/ws`);

    ws.onmessage = () => {
      queryClient.invalidateQueries({ queryKey: ['streamer-clip-room', channel] });
    };

    return () => ws.close();
  }, [channel, queryClient, roomId]);
}
```

- [ ] **Step 5: Compile frontend**

```bash
pnpm --dir frontend build
```

Expected: TypeScript and Vite build pass.

- [ ] **Step 6: Commit**

```bash
git add frontend/src/types/streamerClipRoom.ts frontend/src/lib/streamer-clip-room-api.ts frontend/src/hooks/useStreamerClipRoom.ts frontend/src/hooks/useStreamerClipRoomWebSocket.ts
git commit -m "feat: add streamer clip room frontend data layer"
```

---

## Task 9: Extend Theatre Component for Approval Sidebar

**Files:**
- Modify: `frontend/src/components/playlist/PlaylistTheatreMode.tsx`

- [ ] **Step 1: Add optional props without breaking existing callers**

In `PlaylistTheatreModeProps`, add:

```ts
pendingLabel?: string;
commentsLabel?: string;
extraSidebarTab?: {
  id: string;
  label: string;
  content: React.ReactNode;
  count?: number;
};
```

Change internal tab state from:

```ts
const [activeTab, setActiveTab] = useState<'queue' | 'chat'>('queue');
```

to:

```ts
const [activeTab, setActiveTab] = useState<'queue' | 'chat' | 'extra'>('queue');
```

Keep keyboard shortcuts: `q` opens queue, `c` opens comments. Do not bind a shortcut for the approval tab in v1 to avoid conflict.

- [ ] **Step 2: Rename visible comments label only**

Where the tab currently displays `Chat` for `CommentSection`, render:

```tsx
{commentsLabel ?? 'Comments'}
```

Keep the internal state value `'chat'` to reduce churn.

- [ ] **Step 3: Render extra tab when provided**

Near existing tab buttons, add:

```tsx
{extraSidebarTab && (
  <Tab
    active={activeTab === 'extra'}
    onClick={() => setActiveTab('extra')}
  >
    {extraSidebarTab.label}
    {typeof extraSidebarTab.count === 'number' && extraSidebarTab.count > 0 && (
      <span className="ml-2 rounded-full bg-primary-500 px-2 py-0.5 text-xs text-white">
        {extraSidebarTab.count}
      </span>
    )}
  </Tab>
)}
```

Where tab content is rendered, add:

```tsx
{activeTab === 'extra' && extraSidebarTab?.content}
```

- [ ] **Step 4: Build frontend**

```bash
pnpm --dir frontend build
```

Expected: build passes and existing playlist theatre callers compile unchanged.

- [ ] **Step 5: Commit**

```bash
git add frontend/src/components/playlist/PlaylistTheatreMode.tsx
git commit -m "feat: allow theatre mode approval sidebar"
```

---

## Task 10: Build Streamer Clip Room Page

**Files:**
- Create: `frontend/src/pages/StreamerClipRoomPage.tsx`
- Create: `frontend/src/pages/StreamerClipRoomPage.test.tsx`
- Modify: `frontend/src/pages/index.ts`
- Modify: `frontend/src/App.tsx`

- [ ] **Step 1: Add page component**

Create `frontend/src/pages/StreamerClipRoomPage.tsx`:

```tsx
import { useCallback, useMemo, useState } from 'react';
import { useNavigate, useParams } from 'react-router-dom';
import { Check, X } from 'lucide-react';
import { SEO } from '@/components/SEO';
import { Button, Spinner } from '@/components/ui';
import { PlaylistTheatreMode, type PlaylistItem } from '@/components/playlist/PlaylistTheatreMode';
import {
  useApproveStreamerClipRoomItem,
  useRejectStreamerClipRoomItem,
  useReorderStreamerClipRoomItems,
  useStartStreamerClipRoom,
  useStopStreamerClipRoom,
  useStreamerClipRoom,
} from '@/hooks/useStreamerClipRoom';
import { useStreamerClipRoomWebSocket } from '@/hooks/useStreamerClipRoomWebSocket';
import type { StreamerClipRoomItem } from '@/types/streamerClipRoom';

function ApprovalQueue({ items, onApprove, onReject }: {
  items: StreamerClipRoomItem[];
  onApprove: (id: string) => void;
  onReject: (id: string) => void;
}) {
  if (items.length === 0) {
    return <p className="p-4 text-sm text-muted-foreground">No clips waiting for approval.</p>;
  }

  return (
    <div className="space-y-3 p-3">
      {items.map(item => (
        <div key={item.id} className="rounded-lg border border-border bg-surface p-3">
          <p className="text-sm font-medium text-foreground line-clamp-2">{item.clip?.title ?? item.source_url}</p>
          <p className="mt-1 text-xs text-muted-foreground">Posted by {item.twitch_username ?? 'unknown viewer'}</p>
          {item.message_text && <p className="mt-2 text-xs text-muted-foreground line-clamp-3">{item.message_text}</p>}
          <div className="mt-3 flex gap-2">
            <Button size="sm" onClick={() => onApprove(item.id)}><Check className="mr-1 h-4 w-4" />Approve</Button>
            <Button size="sm" variant="outline" onClick={() => onReject(item.id)}><X className="mr-1 h-4 w-4" />Reject</Button>
          </div>
        </div>
      ))}
    </div>
  );
}

export function StreamerClipRoomPage() {
  const { channel = '' } = useParams<{ channel: string }>();
  const navigate = useNavigate();
  const { data, isLoading, isError } = useStreamerClipRoom(channel);
  const roomId = data?.room.id;
  useStreamerClipRoomWebSocket(channel, roomId);

  const startRoom = useStartStreamerClipRoom(channel);
  const stopRoom = useStopStreamerClipRoom(channel);
  const approveItem = useApproveStreamerClipRoomItem(channel, roomId);
  const rejectItem = useRejectStreamerClipRoomItem(channel, roomId);
  const reorderItems = useReorderStreamerClipRoomItems(channel, roomId);
  const [currentItemId, setCurrentItemId] = useState<string | null>(null);

  const approvedItems = data?.approved_items ?? [];
  const pendingItems = data?.pending_items ?? [];

  const theatreItems: PlaylistItem[] = useMemo(() => approvedItems
    .filter(item => item.clip)
    .sort((a, b) => (a.position ?? 0) - (b.position ?? 0))
    .map(item => ({ id: item.id, clip_id: item.clip_id!, clip: item.clip, order: item.position })), [approvedItems]);

  if (!currentItemId && theatreItems.length > 0) {
    setCurrentItemId(theatreItems[0].id);
  }

  const handleReorder = useCallback((itemId: string, newPosition: number) => {
    const oldIndex = theatreItems.findIndex(item => item.id === itemId);
    if (oldIndex === -1) return;
    const next = [...theatreItems];
    const [moved] = next.splice(oldIndex, 1);
    next.splice(newPosition, 0, moved);
    reorderItems.mutate(next.map(item => item.id));
  }, [reorderItems, theatreItems]);

  if (isLoading) {
    return <div className="fixed inset-0 flex items-center justify-center bg-black"><Spinner size="lg" /></div>;
  }

  if (isError || !data) {
    return <div className="fixed inset-0 flex items-center justify-center bg-black text-white">Failed to load streamer clip room.</div>;
  }

  const approvalContent = (
    <ApprovalQueue
      items={pendingItems}
      onApprove={id => approveItem.mutate(id)}
      onReject={id => rejectItem.mutate(id)}
    />
  );

  if (theatreItems.length === 0) {
    return (
      <>
        <SEO title={`${channel} Clip Room`} />
        <div className="min-h-screen bg-black p-6 text-white">
          <div className="mx-auto max-w-3xl rounded-xl border border-white/10 bg-white/5 p-6">
            <h1 className="text-2xl font-bold">{channel} Clip Room</h1>
            <p className="mt-2 text-white/70">Start the room and approve clips as viewers post Clpr links in Twitch chat.</p>
            <div className="mt-4 flex gap-2">
              <Button onClick={() => startRoom.mutate()} disabled={startRoom.isPending}>Start listener</Button>
              <Button variant="outline" onClick={() => stopRoom.mutate()} disabled={stopRoom.isPending}>Stop listener</Button>
            </div>
            <div className="mt-6 rounded-lg bg-black/40">{approvalContent}</div>
          </div>
        </div>
      </>
    );
  }

  return (
    <>
      <SEO title={`${channel} Clip Room`} />
      <div className="min-h-screen bg-black">
        <PlaylistTheatreMode
          title={`${channel} Clip Room`}
          items={theatreItems}
          currentItemId={currentItemId}
          onItemClick={item => setCurrentItemId(item.id)}
          onReorder={handleReorder}
          onClose={() => navigate(`/stream/${channel}`)}
          commentsLabel="Discussion"
          extraSidebarTab={{ id: 'approval', label: 'Approve', count: pendingItems.length, content: approvalContent }}
          isQueue={false}
          contained={false}
        />
      </div>
    </>
  );
}
```

- [ ] **Step 2: Export page**

Modify `frontend/src/pages/index.ts`:

```ts
export { StreamerClipRoomPage } from './StreamerClipRoomPage';
```

- [ ] **Step 3: Add protected route**

Modify `frontend/src/App.tsx` imports and route list. Add:

```tsx
<Route
  path="/streamer-tools/:channel/clips"
  element={
    <ProtectedRoute>
      <StreamerClipRoomPage />
    </ProtectedRoute>
  }
/>
```

- [ ] **Step 4: Add page test**

Create `frontend/src/pages/StreamerClipRoomPage.test.tsx` with a mocked API response asserting:
- the page renders `Start listener` when no approved clips exist,
- pending clips show `Approve` and `Reject`,
- approved clips map into theatre mode list.

Use the existing test setup from `frontend/src/test/AllTheProviders.tsx` and mock handlers from `frontend/src/test/mocks/handlers.ts` as examples.

- [ ] **Step 5: Run frontend tests and build**

```bash
pnpm --dir frontend test -- StreamerClipRoomPage.test.tsx
pnpm --dir frontend build
```

Expected: test and build pass.

- [ ] **Step 6: Commit**

```bash
git add frontend/src/pages/StreamerClipRoomPage.tsx frontend/src/pages/StreamerClipRoomPage.test.tsx frontend/src/pages/index.ts frontend/src/App.tsx
git commit -m "feat: add streamer clip room page"
```

---

## Task 11: Add Entry Point from Stream Page

**Files:**
- Modify: `frontend/src/pages/StreamPage.tsx`

- [ ] **Step 1: Add link for authenticated users**

In `StreamPage.tsx`, import `Link` from `react-router-dom` if not already imported, then add next to chat controls:

```tsx
{isAuthenticated && (
  <Link
    to={`/streamer-tools/${encodeURIComponent(streamer)}/clips`}
    className="px-3 py-1 text-sm rounded border border-purple-500 text-purple-700 dark:text-purple-300 hover:bg-purple-50 dark:hover:bg-purple-900/20"
  >
    Clip Room
  </Link>
)}
```

- [ ] **Step 2: Build frontend**

```bash
pnpm --dir frontend build
```

Expected: build passes.

- [ ] **Step 3: Commit**

```bash
git add frontend/src/pages/StreamPage.tsx
git commit -m "feat: link streams to clip room"
```

---

## Task 12: Final Verification

**Files:**
- Modify docs only if endpoint docs are required by repository conventions.

- [ ] **Step 1: Run backend focused tests**

```bash
go test ./backend/internal/services -run 'TestExtractClipLinks|TestStreamerClipRoomService|TestParseTwitchIRCPrivmsg' -count=1
go test ./backend/internal/repository ./backend/internal/handlers ./backend/cmd/api -run TestNonExistent -count=1
```

Expected: service tests pass and backend packages compile.

- [ ] **Step 2: Run frontend checks**

```bash
pnpm --dir frontend test -- StreamerClipRoomPage.test.tsx
pnpm --dir frontend build
```

Expected: frontend test and build pass.

- [ ] **Step 3: Manual smoke test**

Start backend and frontend using the repository's normal local commands. In a browser:
1. Log in.
2. Visit `/streamer-tools/moonmoon/clips`.
3. Click `Start listener`.
4. Insert a test row with `status='pending'` for a known clip or post a test Clpr URL in Twitch chat if credentials are configured.
5. Confirm pending item appears in approval queue.
6. Click `Approve`.
7. Confirm clip appears in theatre list and begins as current item when first approved.
8. Drag/reorder approved clips.
9. Press `c` or click `Discussion`; confirm `CommentSection` loads for current clip.

- [ ] **Step 4: Update docs if API docs are expected**

If OpenAPI docs are maintained for all new endpoints, update `docs/openapi/openapi.yaml` with `/streamer-clip-rooms` paths and run:

```bash
pnpm openapi:validate
```

Expected: OpenAPI validation passes.

- [ ] **Step 5: Final commit**

```bash
git status --short
git add docs/openapi/openapi.yaml
git commit -m "docs: document streamer clip room API"
```

Run the docs commit only if `docs/openapi/openapi.yaml` changed. If it did not change, verify `git status --short` is clean after previous commits.

---

## Risks and Mitigations

- **Twitch IRC auth/reconnect complexity:** keep listener manager isolated in `twitch_chat_listener.go`, test parser separately, and expose listener status on the room.
- **Duplicate messages:** enforce `(room_id, twitch_message_id, source_url)` uniqueness where message ID exists and de-duplicate URLs in parser.
- **Unknown Twitch clips:** persist skipped items with clear reason instead of failing silently.
- **Reorder race conditions:** repository reorder uses a transaction and temporary position offset.
- **Theatre component regression:** new props are optional and existing callers compile unchanged.
- **WebSocket auth mismatch:** if the existing API path does not support cookies for WS, copy the authenticated watch-party WS approach rather than inventing a second auth pattern.

## Self-Review

- Spec coverage: backend bot, approval queue, playlist population, reorder, theatre default, and comments/discussion access are each mapped to tasks.
- Placeholder scan: this plan contains no unspecified implementation placeholders; every task has concrete files, methods, commands, and expected behavior.
- Type consistency: backend statuses and frontend types match: `pending`, `approved`, `rejected`, `skipped`; source types match: `clpr`, `twitch`.
- Scope control: v1 intentionally supports existing Clpr clips and known Twitch clips only; automatic importing of unknown Twitch links is excluded and represented as `skipped`.
