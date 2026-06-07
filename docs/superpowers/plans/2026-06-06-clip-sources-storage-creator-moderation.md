# Clip Sources, Hosted Uploads, and Creator Moderation Implementation Plan

> **For agentic workers:** Execute this plan task-by-task. Recommended path:
> dispatch a fresh subagent per task, review each result with `review-quality`,
> then continue. For complex multi-agent splits, use
> `parallel-feature-development`, `team-composition-patterns`, and
> `team-communication-protocols`. Steps use checkbox (`- [ ]`) syntax for
> tracking.

**Goal:** Expand clpr from Twitch-only submissions to support Kick, TikTok, YouTube/Shorts, and hosted uploads on MinIO/S3-compatible storage, while adding creator-scoped moderation foundations.

**Architecture:** Keep the existing submission/moderation approval flow, but generalize clip source metadata and storage so Twitch, external URLs, and hosted uploads pass through the same queue. Build storage against the S3 API now so MinIO works first and Cloudflare R2/AWS S3 can replace it later through config and object migration. Creator moderation is local-first: creator bans block interaction/submission, not viewing; hidden comments show an explanation.

**Tech Stack:** Go backend, Gin HTTP handlers, PostgreSQL migrations, Vite React frontend, existing Twitch submission services, MinIO/S3-compatible object storage, `ffprobe`/`ffmpeg` for hosted upload validation, HLS/video playback via existing `video_url` support.

---

## Product Decisions Captured

- Initial platforms: `twitch`, `kick`, `youtube`, `youtube_shorts`, `tiktok`, and `upload`.
- Deferred platforms: `x/twitter` and `instagram_reels`.
- MinIO is acceptable because it is S3-compatible; implement an S3-compatible storage interface rather than MinIO-specific code.
- Make limits config-driven:
  - `CLIP_MAX_DURATION_SECONDS`, default `600`.
  - `CLIP_RECOMMENDED_DURATION_SECONDS`, default `420`.
  - `CLIP_MAX_UPLOAD_BYTES`, default `1073741824`.
- Length recommendation: hard max 10 minutes, recommended max 7 minutes.
- Upload size recommendation: 1GB initial default, adjustable without code changes.
- Allow 1080p60 within the configured duration and size caps; do not hard-code bitrate math.
- Trusted users may auto-approve URL submissions.
- Hosted uploads require moderation approval at launch, including trusted users, until automated safety scanning exists.
- Creator bans block interaction/submission/commenting, not viewing.
- Hidden comments should show an explanation instead of silently disappearing.
- Twitch ban-list import and Twitch mod delegation are core future goals.
- Kick moderation should be modeled as capability-based, not assumed.
- YouTube/TikTok should be treated as source/metadata platforms first, not moderation-sync platforms.

---

## Relevant Existing Files

### Frontend

- `frontend/src/pages/SubmitClipPage.tsx` — current Twitch-only submission page.
- `frontend/src/lib/submission-api.ts` — current submit/check/metadata client calls.
- `frontend/src/types/submission.ts` — current `SubmitClipRequest` and `ClipSubmission` types.
- `frontend/src/types/clip.ts` — clip type already has `video_url?: string` and limited `stream_source`.
- `frontend/src/pages/ClipDetailPage.tsx` — already chooses hosted/HLS playback when `clip.video_url` exists.
- `frontend/src/components/video/VideoPlayer.tsx` — Twitch embed playback.
- `frontend/src/components/video/HlsPlayer.tsx` — HLS playback support.

### Backend

- `backend/cmd/api/routes_content.go` — authenticated submission routes.
- `backend/cmd/api/routes_clips.go` — clip routes.
- `backend/config/config.go` — add source/storage/duration/upload config.
- `backend/internal/handlers/submission_handler.go` — update submit handler and add upload endpoint.
- `backend/internal/services/submission_service.go` — generalize Twitch-only submission flow.
- `backend/internal/repository/submission_repository.go` — add source/upload fields and duplicate checks.
- `backend/internal/handlers/clip_handler.go` — enforce creator interaction bans where needed.
- `backend/internal/services/clip_service.go` — enforce creator interaction bans where needed.
- `backend/internal/repository/clip_repository.go` — query generalized clip fields.
- `backend/internal/models/models.go` — add source, upload, creator account, and creator moderation models.
- `backend/migrations/000001_initial_schema.up.sql` — reference only; do not edit applied migration.
- `backend/migrations/000004_add_clip_submissions.up.sql` — reference only; do not edit applied migration.
- `backend/migrations/000066_add_clip_stream_support.up.sql` — existing hosted stream fields.
- `backend/migrations/000083_add_clip_mirrors.up.sql` — existing mirror/CDN-related fields.

---

## Proposed Data Model

### `clips` additions

- `source_type VARCHAR(20) NOT NULL DEFAULT 'twitch'`
  - allowed: `twitch`, `external`, `upload`.
- `source_platform VARCHAR(30) NOT NULL DEFAULT 'twitch'`
  - allowed: `twitch`, `kick`, `youtube`, `youtube_shorts`, `tiktok`, `upload`.
- `source_url TEXT`
- `source_id VARCHAR(255)`
- `source_metadata JSONB NOT NULL DEFAULT '{}'::jsonb`
- `duration_seconds INTEGER`
- `duration_verified BOOLEAN NOT NULL DEFAULT false`
- `storage_provider VARCHAR(30)`
- `storage_bucket VARCHAR(255)`
- `storage_key TEXT`
- `original_filename TEXT`
- `mime_type VARCHAR(255)`
- `file_size_bytes BIGINT`

### `clip_submissions` additions

Use the same source/upload fields as `clips`, plus moderation support fields if not already present:

- `upload_status VARCHAR(30) NOT NULL DEFAULT 'none'`
  - allowed: `none`, `pending`, `uploaded`, `validated`, `rejected`.
- `duration_validation_error TEXT`
- `storage_visibility VARCHAR(20) NOT NULL DEFAULT 'private'`
  - allowed: `private`, `public`.

### Creator moderation tables

- `creator_accounts`
  - one local creator identity, independent of platform.
- `creator_platform_accounts`
  - links a creator to Twitch/Kick/YouTube/TikTok accounts and capability flags.
- `creator_moderators`
  - local creator-scoped moderators, including imported Twitch mods later.
- `creator_bans`
  - local bans that block interaction/submission/commenting, not viewing.

---

## Task 1: Add Configurable Clip Source and Upload Limits

**Files:**
- Modify: `backend/config/config.go`
- Test: existing backend config tests if present; otherwise add focused tests near existing config tests.

- [ ] **Step 1: Add config fields**

Add a config section equivalent to:

```go
type ClipSourceConfig struct {
    MaxDurationSeconds         int64    `mapstructure:"max_duration_seconds"`
    RecommendedDurationSeconds int64    `mapstructure:"recommended_duration_seconds"`
    MaxUploadBytes             int64    `mapstructure:"max_upload_bytes"`
    AllowedUploadMimeTypes     []string `mapstructure:"allowed_upload_mime_types"`
    RequireModerationForUpload bool     `mapstructure:"require_moderation_for_upload"`
}

type ClipStorageConfig struct {
    Provider       string `mapstructure:"provider"`
    Endpoint       string `mapstructure:"endpoint"`
    Bucket         string `mapstructure:"bucket"`
    Region         string `mapstructure:"region"`
    AccessKey      string `mapstructure:"access_key"`
    SecretKey      string `mapstructure:"secret_key"`
    ForcePathStyle bool   `mapstructure:"force_path_style"`
    PublicBaseURL  string `mapstructure:"public_base_url"`
}
```

Defaults:

```go
MaxDurationSeconds: 600
RecommendedDurationSeconds: 420
MaxUploadBytes: 1073741824
AllowedUploadMimeTypes: []string{"video/mp4", "video/webm", "video/quicktime"}
RequireModerationForUpload: true
Provider: "s3"
Region: "us-east-1"
ForcePathStyle: true
```

- [ ] **Step 2: Add environment variables**

Support these env names:

```env
CLIP_MAX_DURATION_SECONDS=600
CLIP_RECOMMENDED_DURATION_SECONDS=420
CLIP_MAX_UPLOAD_BYTES=1073741824
CLIP_ALLOWED_UPLOAD_MIME_TYPES=video/mp4,video/webm,video/quicktime
CLIP_REQUIRE_MODERATION_FOR_UPLOAD=true
CLIP_STORAGE_PROVIDER=s3
CLIP_STORAGE_ENDPOINT=http://localhost:9000
CLIP_STORAGE_BUCKET=clpr-clips
CLIP_STORAGE_REGION=us-east-1
CLIP_STORAGE_ACCESS_KEY=minioadmin
CLIP_STORAGE_SECRET_KEY=minioadmin
CLIP_STORAGE_FORCE_PATH_STYLE=true
CLIP_STORAGE_PUBLIC_BASE_URL=http://localhost:9000/clpr-clips
```

- [ ] **Step 3: Verify config loads**

Run the backend config test suite or the closest existing backend test command.

Expected: defaults load with 10-minute duration, 7-minute recommendation, 1GB max upload size, and upload moderation enabled.

- [ ] **Step 4: Commit**

```bash
git add backend/config/config.go backend/**/**/*config*test*.go
git commit -m "feat: add clip source storage config"
```

---

## Task 2: Add Source and Upload Schema Migration

**Files:**
- Create: `backend/migrations/000084_add_clip_source_upload_support.up.sql`
- Create: `backend/migrations/000084_add_clip_source_upload_support.down.sql`
- Modify: `backend/internal/models/models.go`

- [ ] **Step 1: Create up migration**

Use a new migration number after the current latest migration. If `000084` already exists when executing, choose the next number.

```sql
ALTER TABLE clips
    ADD COLUMN IF NOT EXISTS source_type VARCHAR(20) NOT NULL DEFAULT 'twitch',
    ADD COLUMN IF NOT EXISTS source_platform VARCHAR(30) NOT NULL DEFAULT 'twitch',
    ADD COLUMN IF NOT EXISTS source_url TEXT,
    ADD COLUMN IF NOT EXISTS source_id VARCHAR(255),
    ADD COLUMN IF NOT EXISTS source_metadata JSONB NOT NULL DEFAULT '{}'::jsonb,
    ADD COLUMN IF NOT EXISTS duration_seconds INTEGER,
    ADD COLUMN IF NOT EXISTS duration_verified BOOLEAN NOT NULL DEFAULT false,
    ADD COLUMN IF NOT EXISTS storage_provider VARCHAR(30),
    ADD COLUMN IF NOT EXISTS storage_bucket VARCHAR(255),
    ADD COLUMN IF NOT EXISTS storage_key TEXT,
    ADD COLUMN IF NOT EXISTS original_filename TEXT,
    ADD COLUMN IF NOT EXISTS mime_type VARCHAR(255),
    ADD COLUMN IF NOT EXISTS file_size_bytes BIGINT;

ALTER TABLE clip_submissions
    ADD COLUMN IF NOT EXISTS source_type VARCHAR(20) NOT NULL DEFAULT 'twitch',
    ADD COLUMN IF NOT EXISTS source_platform VARCHAR(30) NOT NULL DEFAULT 'twitch',
    ADD COLUMN IF NOT EXISTS source_url TEXT,
    ADD COLUMN IF NOT EXISTS source_id VARCHAR(255),
    ADD COLUMN IF NOT EXISTS source_metadata JSONB NOT NULL DEFAULT '{}'::jsonb,
    ADD COLUMN IF NOT EXISTS duration_seconds INTEGER,
    ADD COLUMN IF NOT EXISTS duration_verified BOOLEAN NOT NULL DEFAULT false,
    ADD COLUMN IF NOT EXISTS storage_provider VARCHAR(30),
    ADD COLUMN IF NOT EXISTS storage_bucket VARCHAR(255),
    ADD COLUMN IF NOT EXISTS storage_key TEXT,
    ADD COLUMN IF NOT EXISTS original_filename TEXT,
    ADD COLUMN IF NOT EXISTS mime_type VARCHAR(255),
    ADD COLUMN IF NOT EXISTS file_size_bytes BIGINT,
    ADD COLUMN IF NOT EXISTS upload_status VARCHAR(30) NOT NULL DEFAULT 'none',
    ADD COLUMN IF NOT EXISTS duration_validation_error TEXT,
    ADD COLUMN IF NOT EXISTS storage_visibility VARCHAR(20) NOT NULL DEFAULT 'private';

ALTER TABLE clips DROP CONSTRAINT IF EXISTS clips_stream_source_check;
ALTER TABLE clips ADD CONSTRAINT clips_stream_source_check
    CHECK (stream_source IN ('twitch', 'stream', 'external', 'upload'));

ALTER TABLE clips ADD CONSTRAINT clips_source_type_check
    CHECK (source_type IN ('twitch', 'external', 'upload'));
ALTER TABLE clip_submissions ADD CONSTRAINT clip_submissions_source_type_check
    CHECK (source_type IN ('twitch', 'external', 'upload'));

ALTER TABLE clips ADD CONSTRAINT clips_source_platform_check
    CHECK (source_platform IN ('twitch', 'kick', 'youtube', 'youtube_shorts', 'tiktok', 'upload'));
ALTER TABLE clip_submissions ADD CONSTRAINT clip_submissions_source_platform_check
    CHECK (source_platform IN ('twitch', 'kick', 'youtube', 'youtube_shorts', 'tiktok', 'upload'));

ALTER TABLE clip_submissions ADD CONSTRAINT clip_submissions_upload_status_check
    CHECK (upload_status IN ('none', 'pending', 'uploaded', 'validated', 'rejected'));
ALTER TABLE clip_submissions ADD CONSTRAINT clip_submissions_storage_visibility_check
    CHECK (storage_visibility IN ('private', 'public'));

CREATE INDEX IF NOT EXISTS idx_clips_source_platform ON clips(source_platform);
CREATE INDEX IF NOT EXISTS idx_clips_source_id ON clips(source_platform, source_id) WHERE source_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_clip_submissions_source_id ON clip_submissions(source_platform, source_id) WHERE source_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_clip_submissions_upload_status ON clip_submissions(upload_status);
```

- [ ] **Step 2: Create down migration**

Drop constraints, indexes, and new columns. Recreate the prior `stream_source` check with only the old values used by the current schema.

- [ ] **Step 3: Update Go models**

Add matching fields to `models.Clip` and `models.ClipSubmission` using pointer types where nullable.

- [ ] **Step 4: Run migrations locally**

Run the repo's established migration command against a dev DB.

Expected: migration applies cleanly and rollback applies cleanly.

- [ ] **Step 5: Commit**

```bash
git add backend/migrations/000084_add_clip_source_upload_support.*.sql backend/internal/models/models.go
git commit -m "feat: add generalized clip source schema"
```

---

## Task 3: Add Platform URL Detection and Normalization

**Files:**
- Create: `backend/internal/services/source_detector.go`
- Create: `backend/internal/services/source_detector_test.go`
- Modify: `backend/internal/services/submission_service.go`

- [ ] **Step 1: Write detector tests**

Cover these cases:

```text
https://clips.twitch.tv/SomeClip -> twitch
https://www.twitch.tv/example/clip/SomeClip -> twitch
https://kick.com/example?clip=123 -> kick
https://www.youtube.com/watch?v=abc123 -> youtube
https://youtu.be/abc123 -> youtube
https://www.youtube.com/shorts/abc123 -> youtube_shorts
https://www.tiktok.com/@creator/video/123456789 -> tiktok
https://x.com/user/status/123 -> unsupported
https://www.instagram.com/reel/abc -> unsupported
```

- [ ] **Step 2: Implement detector**

Expose:

```go
type SourcePlatform string

const (
    SourcePlatformTwitch        SourcePlatform = "twitch"
    SourcePlatformKick          SourcePlatform = "kick"
    SourcePlatformYouTube       SourcePlatform = "youtube"
    SourcePlatformYouTubeShorts SourcePlatform = "youtube_shorts"
    SourcePlatformTikTok        SourcePlatform = "tiktok"
)

type DetectedSource struct {
    Type       string
    Platform   SourcePlatform
    URL        string
    SourceID   string
    IsExternal bool
}

func DetectClipSource(rawURL string) (DetectedSource, error)
```

- [ ] **Step 3: Integrate detection into submission service without changing behavior yet**

For Twitch URLs, route to the current Twitch path. For non-Twitch supported URLs, return a clear `external sources are not enabled yet` error until Task 6.

- [ ] **Step 4: Run tests**

```bash
go test ./backend/internal/services -run 'TestDetectClipSource'
```

- [ ] **Step 5: Commit**

```bash
git add backend/internal/services/source_detector.go backend/internal/services/source_detector_test.go backend/internal/services/submission_service.go
git commit -m "feat: detect supported clip source urls"
```

---

## Task 4: Add S3-Compatible Storage Adapter for MinIO/R2

**Files:**
- Create: `backend/internal/storage/clip_storage.go`
- Create: `backend/internal/storage/s3_clip_storage.go`
- Create: `backend/internal/storage/s3_clip_storage_test.go`
- Modify: `backend/go.mod`
- Modify: `backend/go.sum`

- [ ] **Step 1: Define storage interface**

```go
package storage

import (
    "context"
    "io"
)

type PutObjectInput struct {
    Key         string
    Body        io.Reader
    Size        int64
    ContentType string
    Metadata    map[string]string
}

type PutObjectResult struct {
    Provider string
    Bucket   string
    Key      string
    PublicURL string
}

type ClipStorage interface {
    PutObject(ctx context.Context, input PutObjectInput) (PutObjectResult, error)
    DeleteObject(ctx context.Context, key string) error
    PublicURL(key string) string
}
```

- [ ] **Step 2: Implement S3 adapter**

Use AWS SDK v2. Configure endpoint override and path-style addressing so MinIO works. Keep the interface generic so R2 works later.

- [ ] **Step 3: Test public URL/key behavior**

Unit-test key joining and public URL generation without hitting real MinIO.

- [ ] **Step 4: Run tests**

```bash
go test ./backend/internal/storage
```

- [ ] **Step 5: Commit**

```bash
git add backend/internal/storage backend/go.mod backend/go.sum
git commit -m "feat: add s3 compatible clip storage"
```

---

## Task 5: Add Hosted Upload Validation and Submission Endpoint

**Files:**
- Modify: `backend/cmd/api/routes_content.go`
- Modify: `backend/internal/handlers/submission_handler.go`
- Modify: `backend/internal/services/submission_service.go`
- Create: `backend/internal/services/upload_validator.go`
- Create: `backend/internal/services/upload_validator_test.go`

- [ ] **Step 1: Add upload validation tests**

Test:

- Reject unsupported MIME type.
- Reject size greater than `CLIP_MAX_UPLOAD_BYTES`.
- Reject duration greater than `CLIP_MAX_DURATION_SECONDS`.
- Accept valid MP4/WebM/QuickTime metadata under limits.

- [ ] **Step 2: Implement validator**

Use `ffprobe` for duration inspection. Keep the command wrapper injectable so tests can stub it.

Expose:

```go
type UploadValidationResult struct {
    DurationSeconds int64
    DurationVerified bool
    MimeType string
    FileSizeBytes int64
}
```

- [ ] **Step 3: Add authenticated upload route**

Add route under authenticated `/submissions`, for example:

```text
POST /submissions/upload
```

This endpoint accepts multipart form data with:

- `file`
- `custom_title`
- `is_nsfw`
- `submission_reason`
- optional `creator_id` once creator accounts exist

- [ ] **Step 4: Store upload privately**

Save to S3/MinIO key:

```text
uploads/pending/{userID}/{submissionID}/original.{ext}
```

Set submission fields:

```text
source_type=upload
source_platform=upload
stream_source=upload
upload_status=validated
storage_visibility=private
video_url=NULL until approval or private-preview policy exists
```

- [ ] **Step 5: Force moderation for upload**

Even trusted users should not auto-approve uploads while `RequireModerationForUpload=true`.

- [ ] **Step 6: Run backend tests**

Run focused service/handler tests first, then broader backend tests.

- [ ] **Step 7: Commit**

```bash
git add backend/cmd/api/routes_content.go backend/internal/handlers/submission_handler.go backend/internal/services/submission_service.go backend/internal/services/upload_validator.go backend/internal/services/upload_validator_test.go
git commit -m "feat: add hosted clip upload submissions"
```

---

## Task 6: Generalize Submission Creation for External URLs

**Files:**
- Modify: `backend/internal/services/submission_service.go`
- Modify: `backend/internal/repository/submission_repository.go`
- Create: `backend/internal/services/external_metadata.go`
- Create: `backend/internal/services/external_metadata_test.go`

- [ ] **Step 1: Add metadata interface**

```go
type ExternalMetadata struct {
    Title string
    AuthorName string
    ThumbnailURL string
    DurationSeconds *int64
    DurationVerified bool
    EmbedURL string
    Raw map[string]any
}

type ExternalMetadataFetcher interface {
    Fetch(ctx context.Context, source DetectedSource) (ExternalMetadata, error)
}
```

- [ ] **Step 2: Implement YouTube first**

Support YouTube/Shorts URL normalization and basic metadata using oEmbed/OpenGraph where available. If duration is unavailable, set `DurationVerified=false` and allow submission for moderation.

- [ ] **Step 3: Add Kick/TikTok stubs**

Add adapters that fetch best-effort OpenGraph metadata, but keep duration unverified if unavailable.

- [ ] **Step 4: Enforce duration if known**

Reject external URLs with verified duration greater than `CLIP_MAX_DURATION_SECONDS`. If duration is unknown, allow submission but mark it clearly for moderators.

- [ ] **Step 5: Preserve Twitch path**

Existing Twitch metadata validation and duplicate behavior must keep passing.

- [ ] **Step 6: Commit**

```bash
git add backend/internal/services/submission_service.go backend/internal/repository/submission_repository.go backend/internal/services/external_metadata.go backend/internal/services/external_metadata_test.go
git commit -m "feat: support external clip url submissions"
```

---

## Task 7: Make Approval Source-Aware

**Files:**
- Modify: `backend/internal/services/submission_service.go`
- Modify: `backend/internal/repository/clip_repository.go`
- Modify: `backend/internal/models/models.go`

- [ ] **Step 1: Split clip creation by source**

Refactor `createClipFromSubmission` into source-aware helpers:

```go
func (s *SubmissionService) createClipFromSubmission(ctx context.Context, submission *models.ClipSubmission) (*models.Clip, error) {
    switch submission.SourceType {
    case "twitch":
        return s.createTwitchClipFromSubmission(ctx, submission)
    case "external":
        return s.createExternalClipFromSubmission(ctx, submission)
    case "upload":
        return s.createUploadClipFromSubmission(ctx, submission)
    default:
        return nil, fmt.Errorf("unsupported source type: %s", submission.SourceType)
    }
}
```

- [ ] **Step 2: Twitch path**

Keep existing fields populated exactly as before, while also filling `source_type=twitch`, `source_platform=twitch`, and `source_url`.

- [ ] **Step 3: External path**

Create clips with:

```text
source_type=external
stream_source=external
source_platform=kick|youtube|youtube_shorts|tiktok
source_url=<normalized URL>
embed_url=<metadata embed URL if available, otherwise source URL>
video_url=NULL
```

- [ ] **Step 4: Upload path**

Create clips with:

```text
source_type=upload
stream_source=upload
source_platform=upload
video_url=<public storage URL>
storage_visibility=public
```

If storage promotion requires copying from pending key to public key, do it here and delete the pending object only after DB update succeeds.

- [ ] **Step 5: Commit**

```bash
git add backend/internal/services/submission_service.go backend/internal/repository/clip_repository.go backend/internal/models/models.go
git commit -m "feat: create clips from multiple source types"
```

---

## Task 8: Update Frontend Submission UX

**Files:**
- Modify: `frontend/src/pages/SubmitClipPage.tsx`
- Modify: `frontend/src/lib/submission-api.ts`
- Modify: `frontend/src/types/submission.ts`
- Modify: `frontend/src/types/clip.ts`

- [ ] **Step 1: Update frontend types**

Add:

```ts
export type ClipSourceType = 'twitch' | 'external' | 'upload';
export type ClipSourcePlatform = 'twitch' | 'kick' | 'youtube' | 'youtube_shorts' | 'tiktok' | 'upload';
```

Extend `SubmitClipRequest` and `ClipSubmission` with source/upload fields from the backend.

- [ ] **Step 2: Add source selector**

In `SubmitClipPage.tsx`, add options:

- Twitch clip URL
- External URL
- Upload video

Text for external URL should list supported platforms only: Kick, TikTok, YouTube, YouTube Shorts.

- [ ] **Step 3: Add upload control**

Show file picker, size/duration copy, progress, and post-upload validation errors.

UX copy:

```text
Maximum duration: 10 minutes. Recommended: 7 minutes or less.
Maximum upload size is configured by the server. 1080p60 is allowed if it fits within the duration and size limits.
Uploads require moderator approval before publication.
```

- [ ] **Step 4: Preserve Twitch metadata preview**

Do not regress existing Twitch duplicate/metadata behavior.

- [ ] **Step 5: Commit**

```bash
git add frontend/src/pages/SubmitClipPage.tsx frontend/src/lib/submission-api.ts frontend/src/types/submission.ts frontend/src/types/clip.ts
git commit -m "feat: add source aware clip submission ui"
```

---

## Task 9: Add Creator Account Foundations

**Files:**
- Create: `backend/migrations/000085_add_creator_accounts.up.sql`
- Create: `backend/migrations/000085_add_creator_accounts.down.sql`
- Modify: `backend/internal/models/models.go`
- Create: `backend/internal/repository/creator_repository.go`
- Create: `backend/internal/services/creator_service.go`

- [ ] **Step 1: Add creator tables**

Create tables:

```sql
CREATE TABLE creator_accounts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    owner_user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    display_name VARCHAR(255) NOT NULL,
    slug VARCHAR(255) UNIQUE NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE creator_platform_accounts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    creator_id UUID NOT NULL REFERENCES creator_accounts(id) ON DELETE CASCADE,
    platform VARCHAR(30) NOT NULL,
    platform_user_id VARCHAR(255) NOT NULL,
    platform_display_name VARCHAR(255) NOT NULL,
    profile_url TEXT,
    can_import_bans BOOLEAN NOT NULL DEFAULT false,
    can_sync_bans_outbound BOOLEAN NOT NULL DEFAULT false,
    can_import_moderators BOOLEAN NOT NULL DEFAULT false,
    can_verify_ownership BOOLEAN NOT NULL DEFAULT false,
    can_fetch_metadata BOOLEAN NOT NULL DEFAULT true,
    access_token_encrypted TEXT,
    refresh_token_encrypted TEXT,
    token_expires_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(platform, platform_user_id)
);
```

- [ ] **Step 2: Link clips/submissions to creators**

Add nullable `creator_id` to `clips` and `clip_submissions`.

- [ ] **Step 3: Add repository/service skeletons**

Implement create/get/list/link-platform basics only. Do not implement OAuth yet.

- [ ] **Step 4: Commit**

```bash
git add backend/migrations/000085_add_creator_accounts.*.sql backend/internal/models/models.go backend/internal/repository/creator_repository.go backend/internal/services/creator_service.go
git commit -m "feat: add creator account foundations"
```

---

## Task 10: Add Local Creator Bans and Moderator Permissions

**Files:**
- Create: `backend/migrations/000086_add_creator_moderation.up.sql`
- Create: `backend/migrations/000086_add_creator_moderation.down.sql`
- Modify: `backend/internal/models/models.go`
- Create: `backend/internal/repository/creator_moderation_repository.go`
- Create: `backend/internal/services/creator_moderation_service.go`

- [ ] **Step 1: Add creator moderation tables**

```sql
CREATE TABLE creator_moderators (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    creator_id UUID NOT NULL REFERENCES creator_accounts(id) ON DELETE CASCADE,
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    platform VARCHAR(30),
    platform_user_id VARCHAR(255),
    permissions TEXT[] NOT NULL DEFAULT ARRAY[]::TEXT[],
    source VARCHAR(30) NOT NULL DEFAULT 'manual',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(creator_id, user_id),
    UNIQUE(creator_id, platform, platform_user_id)
);

CREATE TABLE creator_bans (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    creator_id UUID NOT NULL REFERENCES creator_accounts(id) ON DELETE CASCADE,
    target_user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    target_platform VARCHAR(30),
    target_platform_user_id VARCHAR(255),
    source VARCHAR(30) NOT NULL DEFAULT 'manual',
    reason TEXT,
    scopes TEXT[] NOT NULL DEFAULT ARRAY['interact', 'submit', 'comment']::TEXT[],
    expires_at TIMESTAMPTZ,
    created_by_user_id UUID REFERENCES users(id) ON DELETE SET NULL,
    sync_status VARCHAR(30) NOT NULL DEFAULT 'local_only',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
```

Allowed permissions:

- `manage_creator_clips`
- `approve_creator_submissions`
- `remove_creator_comments`
- `ban_creator_users`
- `sync_platform_bans`

Allowed ban scopes:

- `interact`
- `submit`
- `comment`

Do not add a `view` scope for initial launch.

- [ ] **Step 2: Add moderation service checks**

Expose:

```go
func (s *CreatorModerationService) CanInteract(ctx context.Context, creatorID uuid.UUID, userID uuid.UUID) (bool, string, error)
func (s *CreatorModerationService) CanSubmit(ctx context.Context, creatorID uuid.UUID, userID uuid.UUID) (bool, string, error)
func (s *CreatorModerationService) CanComment(ctx context.Context, creatorID uuid.UUID, userID uuid.UUID) (bool, string, error)
```

If blocked, return explanation:

```text
You cannot interact with this creator's content because of a creator-level moderation restriction.
```

- [ ] **Step 3: Commit**

```bash
git add backend/migrations/000086_add_creator_moderation.*.sql backend/internal/models/models.go backend/internal/repository/creator_moderation_repository.go backend/internal/services/creator_moderation_service.go
git commit -m "feat: add creator scoped moderation controls"
```

---

## Task 11: Enforce Creator Bans on Interaction, Submission, and Comments

**Files:**
- Modify: `backend/internal/services/submission_service.go`
- Modify: `backend/internal/handlers/clip_handler.go`
- Modify: comment-related backend handler/service files discovered in repo.
- Modify: frontend comment display component discovered in repo.

- [ ] **Step 1: Find comment files**

Use search for comment handlers/services/components. Do not guess file paths.

- [ ] **Step 2: Enforce submit ban**

When a submission is associated with a creator, call `CanSubmit`. If false, reject with HTTP 403 and the moderation explanation.

- [ ] **Step 3: Enforce interaction/comment ban**

For likes/votes/comments/follows on creator-owned content, call `CanInteract` or `CanComment`. Return HTTP 403 and explanation.

- [ ] **Step 4: Hide blocked user's existing comments with explanation**

For viewers affected by a creator ban, comment UI should show:

```text
This comment is hidden because your account is restricted from interacting with this creator's content.
```

Do not block clip viewing.

- [ ] **Step 5: Commit**

```bash
git add backend frontend
git commit -m "feat: enforce creator moderation restrictions"
```

---

## Task 12: Add Twitch Moderation Import Planning Hooks

**Files:**
- Modify: `backend/internal/services/creator_service.go`
- Create: `backend/internal/services/platform_moderation_capabilities.go`
- Create: `backend/internal/services/platform_moderation_capabilities_test.go`

- [ ] **Step 1: Add capability model**

```go
type PlatformModerationCapabilities struct {
    CanImportBans bool
    CanSyncBansOutbound bool
    CanImportModerators bool
    CanVerifyOwnership bool
    CanFetchMetadata bool
}
```

- [ ] **Step 2: Define platform defaults**

```go
Twitch:  import bans=true, outbound sync=true, import moderators=true, verify ownership=true, fetch metadata=true
Kick:    import bans=false initially, outbound sync=false initially, import moderators=false initially, verify ownership=true if API supports it, fetch metadata=true
YouTube: import bans=false, outbound sync=false, import moderators=false, verify ownership=true, fetch metadata=true
TikTok:  import bans=false, outbound sync=false, import moderators=false, verify ownership=true if API supports it, fetch metadata=true
```

- [ ] **Step 3: Do not implement OAuth in this task**

This task only prevents Twitch assumptions from leaking into the creator model.

- [ ] **Step 4: Commit**

```bash
git add backend/internal/services/creator_service.go backend/internal/services/platform_moderation_capabilities.go backend/internal/services/platform_moderation_capabilities_test.go
git commit -m "feat: model platform moderation capabilities"
```

---

## Task 13: Verification Pass

**Files:**
- No direct source changes expected unless tests fail.

- [ ] **Step 1: Backend tests**

Run the repo's backend test command. If unavailable, run targeted packages changed by this plan.

- [ ] **Step 2: Frontend tests/typecheck**

Run the repo's frontend typecheck/test command.

- [ ] **Step 3: Manual upload smoke test**

With local MinIO configured, upload a short MP4 and verify:

- Submission is created.
- Object exists in MinIO under pending key.
- Submission requires moderation.
- Approval creates clip with `video_url`.
- Clip detail page plays hosted video.

- [ ] **Step 4: Manual external URL smoke test**

Submit YouTube, YouTube Shorts, TikTok, and Kick URLs. Verify supported URLs are accepted into moderation queue and X/Instagram are rejected.

- [ ] **Step 5: Manual creator ban smoke test**

Create a creator ban for a user. Verify:

- User can still view clips.
- User cannot submit to that creator.
- User cannot comment/interact.
- Existing comments show explanation.

- [ ] **Step 6: Final review**

Run `review-quality` or equivalent code review before merge.

---

## Out of Scope for Initial Implementation

- X/Twitter support.
- Instagram Reels support.
- Automatic illegal-content detection beyond validation/scanning primitives.
- Outbound Twitch ban sync.
- Twitch OAuth scopes and token storage hardening.
- Kick ban sync until API capability is confirmed.
- YouTube or TikTok ban-list sync.
- Transcoding pipeline beyond accepting 1080p60 uploads within configured limits.

---

## Open Questions Before Execution

1. Should hosted uploads be served directly from MinIO public URLs, or through a Caddy/app-controlled media domain from day one?
2. Is `1GB` acceptable as the default upload cap in production, or should production start lower while the variable exists?
3. Should external URL submissions with unverified duration be moderator-only forever, or can trusted users auto-approve them after a reputation threshold?
4. Should creator account ownership be created manually first, or should the first implementation include Twitch OAuth verification?
