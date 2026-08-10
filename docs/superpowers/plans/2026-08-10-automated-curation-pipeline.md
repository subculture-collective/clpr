# Automated Curation Pipeline Implementation Plan

> **For agentic workers:** Execute this plan task-by-task. Recommended path:
> dispatch a fresh subagent per task, review each result with `review-quality`,
> then continue. For complex multi-agent splits, use
> `parallel-feature-development`, `team-composition-patterns`, and
> `team-communication-protocols`. Steps use checkbox (`- [ ]`) syntax for
> tracking.

**Goal:** Automatically pull clips from Twitch across 50+ categories and followed broadcasters, tag them with hierarchical structural + AI-derived tags, and feed them into time-scheduled curated playlists that keep the site feeling alive with zero manual curation.

**Architecture:** Add a tag hierarchy (`parent_slug`) to the tags table. Expand clip ingestion from 10 hardcoded games to dynamic top-50 categories + followed broadcasters + global trending. Build an auto-tagging pipeline: structural tags from metadata (free), content tags from Whisper transcription (reusing hasanara's `whisper_runner.py`) + thumbnail vision AI. Extend playlist scripts to filter by multiple tags. Merge discovery lists into playlist scripts via `is_curated`/`is_featured` flags. Add user-tag promotion with a moderation gate.

**Tech Stack:** Go (backend), TypeScript/React (frontend), PostgreSQL + migrations, faster-whisper (CTranslate2 on GPU), ffmpeg (audio/thumbnail extraction), vision model API (GPT-4o-mini or Claude Haiku), hasanara `worker/whisper_runner.py` (reused as-is)

---

## Scope Check

This plan covers six subsystems. Each phase produces independently working, testable software. Phases should be executed in order but Phase 3 (auto-tagging) and Phase 4 (multi-tag scripts) can run in parallel.

| Phase | Description | Delivers |
|-------|-------------|----------|
| 1 | Tag hierarchy + CRUD | Schema, models, API, admin UI |
| 2 | Expanded clip ingestion | Dynamic categories, followed broadcasters, global trending |
| 3 | Auto-tagging pipeline | Whisper, thumbnails, vision AI, structural tags |
| 4 | Multi-tag playlist scripts | AND/OR tag filters in scripts |
| 5 | Discovery list merge | DB migration, deprecate discovery API, playlist flags |
| 6 | User tag promotion | Threshold gate, moderator queue, AI vocabulary pool |

---

## File Structure

```
backend/
├── cmd/api/schedulers.go                          (modify: add WhisperTagScheduler)
├── internal/
│   ├── models/models.go                           (modify: tag hierarchy, clip metadata)
│   ├── repository/
│   │   ├── tag_repository.go                      (modify: hierarchy queries)
│   │   ├── clip_repository.go                     (modify: auto-tag upsert)
│   │   ├── playlist_script_repository.go          (modify: multi-tag filters)
│   │   └── broadcaster_repository.go              (modify: GetAllFollowedBroadcasterIDs already exists)
│   ├── services/
│   │   ├── clip_sync_service.go                   (modify: dynamic categories, followed broadcasters)
│   │   ├── playlist_script_service.go             (modify: multi-tag, discovery list flags)
│   │   ├── auto_tagger_service.go                 (CREATE: structural + AI tagging)
│   │   ├── whisper_service.go                     (CREATE: calls hasanara whisper)
│   │   ├── thumbnail_service.go                   (CREATE: extract + classify thumbnails)
│   │   └── tag_promotion_service.go               (CREATE: user tag → AI pool)
│   ├── scheduler/
│   │   ├── clip_sync_scheduler.go                 (modify: new ingestion sources)
│   │   ├── playlist_script_scheduler.go           (modify: discovery list generation)
│   │   └── auto_tag_scheduler.go                  (CREATE: near-real-time tagging)
│   └── handlers/
│       ├── playlist_script_handler.go             (modify: multi-tag, discovery mode)
│       └── tag_handler.go                         (modify: hierarchy endpoints)
├── migrations/
│   ├── NNNNN_add_tag_parent_slug.sql              (CREATE)
│   └── NNNNN_add_playlist_curation_flags.sql      (CREATE)
│
├── whisper/
│   └── whisper_runner.py                          (symlink or copy from ../hasanara/worker/)
│
frontend/src/
├── types/
│   ├── tag.ts                                     (modify: parent_slug, hierarchy)
│   └── playlistScript.ts                          (modify: tags[], discovery_mode)
├── lib/
│   ├── playlist-script-utils.ts                   (modify: strategy meta for discovery)
│   └── tag-api.ts                                 (modify: hierarchy endpoints)
├── components/
│   └── admin/
│       └── PlaylistScriptForm.tsx                 (modify: multi-tag UI, discovery toggle)
```

---

## Phase 1: Tag Hierarchy

### Task 1.1: Add parent_slug to tags table

**Files:**
- Create: `backend/migrations/NNNNN_add_tag_parent_slug.sql`
- Modify: `backend/internal/models/models.go` (Tag model)

- [ ] **Step 1: Write the migration**

```sql
-- Migration: NNNNN_add_tag_parent_slug
-- Add hierarchical parent_slug to tags table

ALTER TABLE tags ADD COLUMN IF NOT EXISTS parent_slug VARCHAR(100);
CREATE INDEX IF NOT EXISTS idx_tags_parent_slug ON tags(parent_slug);
CREATE INDEX IF NOT EXISTS idx_tags_slug_parent ON tags(slug, parent_slug);

-- Add constraint: parent_slug must reference existing tag slug (nullable roots ok)
-- Deferred to avoid circular dependency during insert
```

- [ ] **Step 2: Run migration**

```bash
cd backend && go run cmd/migrate/main.go up
```

- [ ] **Step 3: Update Tag model**

In `backend/internal/models/models.go`, add to the Tag struct:

```go
type Tag struct {
    ID          uuid.UUID  `json:"id" db:"id"`
    Name        string     `json:"name" db:"name"`
    Slug        string     `json:"slug" db:"slug"`
    ParentSlug  *string    `json:"parent_slug,omitempty" db:"parent_slug"`
    Description *string    `json:"description,omitempty" db:"description"`
    Color       *string    `json:"color,omitempty" db:"color"`
    UsageCount  int        `json:"usage_count" db:"usage_count"`
    CreatedAt   time.Time  `json:"created_at" db:"created_at"`
}
```

- [ ] **Step 4: Update TagRepository with hierarchy methods**

In `backend/internal/repository/tag_repository.go`, add:

```go
// GetChildren returns all child tags for a parent slug
func (r *TagRepository) GetChildren(ctx context.Context, parentSlug string) ([]*models.Tag, error) {
    query := `SELECT id, name, slug, parent_slug, description, color, usage_count, created_at
              FROM tags WHERE parent_slug = $1 ORDER BY usage_count DESC`
    // ... standard rows iteration
}

// GetTagTree returns a tag and its full subtree
func (r *TagRepository) GetTagTree(ctx context.Context, rootSlug string) ([]*models.Tag, error) {
    query := `WITH RECURSIVE tag_tree AS (
        SELECT id, name, slug, parent_slug, description, color, usage_count, created_at, 0 AS depth
        FROM tags WHERE slug = $1
        UNION ALL
        SELECT t.id, t.name, t.slug, t.parent_slug, t.description, t.color, t.usage_count, t.created_at, tt.depth + 1
        FROM tags t INNER JOIN tag_tree tt ON t.parent_slug = tt.slug
    ) SELECT * FROM tag_tree ORDER BY depth, usage_count DESC`
    // ... standard rows iteration
}

// GetRootTags returns all tags with no parent
func (r *TagRepository) GetRootTags(ctx context.Context) ([]*models.Tag, error) {
    query := `SELECT id, name, slug, parent_slug, description, color, usage_count, created_at
              FROM tags WHERE parent_slug IS NULL ORDER BY usage_count DESC`
    // ... standard rows iteration
}
```

- [ ] **Step 5: Verify with integration test**

```bash
cd backend && go test ./internal/repository/ -run TagRepository -v -count=1
```

- [ ] **Step 6: Commit**

```bash
git add backend/migrations/ backend/internal/models/models.go backend/internal/repository/tag_repository.go
git commit -m "feat: add tag hierarchy with parent_slug and recursive tree queries"
```

---

### Task 1.2: Update tag API handlers for hierarchy

**Files:**
- Modify: `backend/internal/handlers/tag_handler.go`

- [ ] **Step 1: Add GET /tags/tree endpoint**

```go
// GetTagTree returns the full tag hierarchy
// GET /api/v1/tags/tree
func (h *TagHandler) GetTagTree(c *gin.Context) {
    rootSlug := c.DefaultQuery("root", "")
    var tags []*models.Tag
    var err error
    if rootSlug != "" {
        tags, err = h.tagRepo.GetTagTree(c.Request.Context(), rootSlug)
    } else {
        tags, err = h.tagRepo.GetRootTags(c.Request.Context())
        // Also fetch children for each root
        for _, root := range tags {
            children, _ := h.tagRepo.GetChildren(c.Request.Context(), root.Slug)
            // Attach children to response structure
        }
    }
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch tag tree"})
        return
    }
    c.JSON(http.StatusOK, gin.H{"tags": tags})
}
```

- [ ] **Step 2: Add parent_slug to CreateTag/UpdateTag**

Accept `parent_slug` in the request body for tag creation and updates. Validate that parent_slug references an existing tag.

```go
type CreateTagRequest struct {
    Name       string  `json:"name" binding:"required,min=1,max=50"`
    Slug       string  `json:"slug" binding:"required,min=1,max=100"`
    ParentSlug *string `json:"parent_slug,omitempty" binding:"omitempty,max=100"`
    // ...
}
```

- [ ] **Step 3: Register new route**

In the router setup:

```go
tagGroup.GET("/tree", tagHandler.GetTagTree)
```

- [ ] **Step 4: Commit**

```bash
git add backend/internal/handlers/tag_handler.go
git commit -m "feat: add tag hierarchy API endpoints (tree, parent_slug creation)"
```

---

### Task 1.3: Seed structural tag hierarchy

**Files:**
- Create: `backend/scripts/seed_tag_hierarchy.sql`

- [ ] **Step 1: Create seed SQL**

```sql
-- Structural tags: assigned programmatically from clip metadata
INSERT INTO tags (id, name, slug, parent_slug) VALUES
  (gen_random_uuid(), 'Game', 'game', NULL),
  (gen_random_uuid(), 'Duration', 'duration', NULL),
  (gen_random_uuid(), 'Language', 'lang', NULL),
  (gen_random_uuid(), 'Broadcaster Tier', 'tier', NULL);

-- Duration children
INSERT INTO tags (id, name, slug, parent_slug)
  SELECT gen_random_uuid(), v.name, v.slug, 'duration'
  FROM (VALUES
    ('Short (0-30s)', 'duration/short'),
    ('Medium (30-90s)', 'duration/medium'),
    ('Long (90s+)', 'duration/long')
  ) AS v(name, slug);

-- Language children (common Twitch languages)
INSERT INTO tags (id, name, slug, parent_slug)
  SELECT gen_random_uuid(), v.name, v.slug, 'lang'
  FROM (VALUES
    ('English', 'lang/en'),
    ('Spanish', 'lang/es'),
    ('Portuguese', 'lang/pt'),
    ('French', 'lang/fr'),
    ('German', 'lang/de'),
    ('Russian', 'lang/ru'),
    ('Japanese', 'lang/ja'),
    ('Korean', 'lang/ko')
  ) AS v(name, slug);

-- Broadcaster tier children
INSERT INTO tags (id, name, slug, parent_slug)
  SELECT gen_random_uuid(), v.name, v.slug, 'tier'
  FROM (VALUES
    ('Partner', 'tier/partner'),
    ('Affiliate', 'tier/affiliate'),
    ('Non-affiliate', 'tier/non-affiliate')
  ) AS v(name, slug);

-- Content tags: assigned by AI (roots for future children)
INSERT INTO tags (id, name, slug, parent_slug) VALUES
  (gen_random_uuid(), 'Content', 'content', NULL);

INSERT INTO tags (id, name, slug, parent_slug)
  SELECT gen_random_uuid(), v.name, v.slug, 'content'
  FROM (VALUES
    ('Clutch Play', 'content/clutch'),
    ('Funny Moment', 'content/funny'),
    ('Fail', 'content/fail'),
    ('Educational', 'content/educational'),
    ('Highlights', 'content/highlights'),
    ('Reaction', 'content/reaction'),
    ('Speedrun', 'content/speedrun'),
    ('Music', 'content/music'),
    ('Art/Creative', 'content/creative'),
    ('IRL/Just Chatting', 'content/irl')
  ) AS v(name, slug);

-- Community tags root (user-promoted tags go here)
INSERT INTO tags (id, name, slug, parent_slug) VALUES
  (gen_random_uuid(), 'Community', 'community', NULL);
```

- [ ] **Step 2: Run the seed**

```bash
psql $DATABASE_URL -f backend/scripts/seed_tag_hierarchy.sql
```

- [ ] **Step 3: Commit**

```bash
git add backend/scripts/seed_tag_hierarchy.sql
git commit -m "feat: seed structural and content tag hierarchy"
```

---

### Task 1.4: Update frontend tag types

**Files:**
- Modify: `frontend/src/types/tag.ts`

- [ ] **Step 1: Update Tag interface**

```typescript
export interface Tag {
  id: string;
  name: string;
  slug: string;
  parent_slug?: string | null;
  description?: string;
  color?: string;
  usage_count: number;
  created_at: string;
}

export interface TagTreeResponse {
  tags: Tag[];
  children?: Record<string, Tag[]>;  // slug -> children
}
```

- [ ] **Step 2: Commit**

```bash
git add frontend/src/types/tag.ts
git commit -m "feat: add parent_slug and tree types to Tag interface"
```

---

## Phase 2: Expanded Clip Ingestion

### Task 2.1: Replace hardcoded game IDs with dynamic top categories

**Files:**
- Modify: `backend/internal/services/clip_sync_service.go`

- [ ] **Step 1: Add fetchTopCategories method**

```go
// TopCategory represents a Twitch category with viewer count
type TopCategory struct {
    GameID      string
    GameName    string
    ViewerCount int
}

// fetchTopCategories retrieves the top N categories by current viewers
func (s *ClipSyncService) fetchTopCategories(ctx context.Context, limit int) ([]TopCategory, error) {
    categories, err := s.twitchClient.GetTopGames(ctx, &twitch.TopGamesParams{
        First: limit,
    })
    if err != nil {
        return nil, fmt.Errorf("fetching top categories: %w", err)
    }

    result := make([]TopCategory, 0, len(categories))
    for _, cat := range categories {
        result = append(result, TopCategory{
            GameID:      cat.ID,
            GameName:    cat.Name,
            ViewerCount: cat.ViewerCount,
        })
    }
    return result, nil
}
```

- [ ] **Step 2: Replace SyncTrendingClips to use dynamic categories**

Replace the `defaultTrendingGameIDs` hardcoded list with a call to `fetchTopCategories`:

```go
func (s *ClipSyncService) SyncTrendingClips(ctx context.Context, hours int, opts *TrendingSyncOptions) (*SyncStats, error) {
    // Fetch top 50 categories dynamically
    categories, err := s.fetchTopCategories(ctx, 50)
    if err != nil {
        // Fallback to hardcoded list if Twitch API fails
        utils.Warn("Failed to fetch top categories, using fallback", map[string]interface{}{
            "error": err.Error(),
        })
        categories = fallbackCategories()
    }

    // Build trending game configs: top 3 get 5 clips each, rest get 3
    configs := make([]TrendingGameConfig, 0, len(categories))
    for i, cat := range categories {
        limit := 3
        if i < 3 {
            limit = 5
        }
        configs = append(configs, TrendingGameConfig{
            GameID: cat.GameID,
            Name:   cat.GameName,
            Limit:  limit,
        })
    }

    return s.syncClipsByGames(ctx, hours, configs, opts)
}
```

- [ ] **Step 3: Add fallbackCategories for when Twitch API is down**

```go
func fallbackCategories() []TopCategory {
    // Keep original 10 as fallback
    ids := defaultTrendingGameIDs
    result := make([]TopCategory, len(ids))
    for i, id := range ids {
        result[i] = TopCategory{GameID: id, GameName: "fallback"}
    }
    return result
}
```

- [ ] **Step 4: Verify with existing tests**

```bash
cd backend && go test ./internal/services/ -run ClipSync -v -count=1
```

- [ ] **Step 5: Commit**

```bash
git add backend/internal/services/clip_sync_service.go
git commit -m "feat: replace hardcoded game IDs with dynamic top-50 categories from Twitch API"
```

---

### Task 2.2: Add followed-broadcaster clip ingestion

**Files:**
- Modify: `backend/internal/services/clip_sync_service.go`
- Modify: `backend/internal/scheduler/clip_sync_scheduler.go`

- [ ] **Step 1: Add SyncFollowedBroadcasterClips method**

```go
// FollowedBroadcasterSyncOptions configures the followed broadcaster sync
type FollowedBroadcasterSyncOptions struct {
    MinFollowers     int  // minimum clpr followers to include broadcaster
    ClipsPerBroadcaster int  // max clips per broadcaster
    MaxTotalClips    int  // hard cap per sync cycle
    PreferLive       bool // prioritize currently-live broadcasters
    LanguageFilter   string
}

// SyncFollowedBroadcasterClips syncs clips from broadcasters with clpr followers
func (s *ClipSyncService) SyncFollowedBroadcasterClips(ctx context.Context, opts *FollowedBroadcasterSyncOptions) (*SyncStats, error) {
    if opts == nil {
        opts = &FollowedBroadcasterSyncOptions{
            MinFollowers:        3,
            ClipsPerBroadcaster: 5,
            MaxTotalClips:       200,
            PreferLive:          true,
        }
    }

    // Get all broadcaster IDs that have >= MinFollowers on clpr
    broadcasterIDs, err := s.userRepo.GetBroadcastersWithMinFollowers(ctx, opts.MinFollowers)
    if err != nil {
        return nil, fmt.Errorf("fetching followed broadcasters: %w", err)
    }

    if len(broadcasterIDs) == 0 {
        return &SyncStats{}, nil
    }

    // Shuffle and cap
    // ... fetch clips per broadcaster, respecting MaxTotalClips
}
```

- [ ] **Step 2: Add GetBroadcastersWithMinFollowers to user repository**

In `backend/internal/repository/user_repository.go`:

```go
func (r *UserRepository) GetBroadcastersWithMinFollowers(ctx context.Context, minFollowers int) ([]string, error) {
    query := `SELECT broadcaster_id FROM broadcaster_follows
              GROUP BY broadcaster_id HAVING COUNT(*) >= $1`
    // ...
}
```

- [ ] **Step 3: Integrate into ClipSyncScheduler**

In `backend/internal/scheduler/clip_sync_scheduler.go`, add a second sync path:

```go
func (s *ClipSyncScheduler) runSync(ctx context.Context) {
    // 1. Trending categories sync (existing)
    // 2. Followed broadcaster sync (new)
    followerOpts := &services.FollowedBroadcasterSyncOptions{
        MinFollowers:        3,
        ClipsPerBroadcaster: 5,
        MaxTotalClips:       200,
        PreferLive:          true,
    }
    followerStats, err := s.syncService.SyncFollowedBroadcasterClips(ctx, followerOpts)
    // ... metrics
}
```

- [ ] **Step 4: Verify**

```bash
cd backend && go test ./internal/services/ -run ClipSync -v -count=1
```

- [ ] **Step 5: Commit**

```bash
git add backend/internal/services/clip_sync_service.go backend/internal/scheduler/clip_sync_scheduler.go backend/internal/repository/user_repository.go
git commit -m "feat: add followed-broadcaster clip ingestion with caps and live prioritization"
```

---

### Task 2.3: Add global trending clip source

**Files:**
- Modify: `backend/internal/services/clip_sync_service.go`

- [ ] **Step 1: Add SyncGlobalTrending method**

```go
func (s *ClipSyncService) SyncGlobalTrending(ctx context.Context, limit int) (*SyncStats, error) {
    if limit <= 0 {
        limit = 20
    }

    clips, err := s.twitchClient.GetClips(ctx, &twitch.ClipsParams{
        First:      limit,
        StartedAt:  time.Now().Add(-24 * time.Hour),
        EndedAt:    time.Now(),
    })
    if err != nil {
        return nil, fmt.Errorf("fetching global trending: %w", err)
    }

    return s.ingestClips(ctx, clips)
}
```

- [ ] **Step 2: Integrate into scheduler**

```go
func (s *ClipSyncScheduler) runSync(ctx context.Context) {
    // ... existing syncs ...
    globalStats, _ := s.syncService.SyncGlobalTrending(ctx, 20)
    // ... metrics
}
```

- [ ] **Step 3: Commit**

```bash
git add backend/internal/services/clip_sync_service.go backend/internal/scheduler/clip_sync_scheduler.go
git commit -m "feat: add global trending clip ingestion from Twitch"
```

---

## Phase 3: Auto-Tagging Pipeline

### Task 3.1: Create structural tagger (programmatic, free)

**Files:**
- Create: `backend/internal/services/auto_tagger_service.go`

- [ ] **Step 1: Write the service**

```go
package services

import (
    "context"
    "fmt"

    "git.subcult.tv/subculture-collective/clpr/internal/models"
    "git.subcult.tv/subculture-collective/clpr/internal/repository"
    "github.com/google/uuid"
)

// AutoTaggerService handles automatic clip tagging
type AutoTaggerService struct {
    tagRepo  *repository.TagRepository
    clipRepo *repository.ClipRepository
}

func NewAutoTaggerService(tagRepo *repository.TagRepository, clipRepo *repository.ClipRepository) *AutoTaggerService {
    return &AutoTaggerService{tagRepo: tagRepo, clipRepo: clipRepo}
}

// TagClip assigns structural tags to a clip based on metadata
func (s *AutoTaggerService) TagClip(ctx context.Context, clip *models.Clip) ([]string, error) {
    var slugs []string

    // Duration tags
    slugs = append(slugs, durationTag(clip.Duration))

    // Language tag
    if clip.Language != "" {
        slugs = append(slugs, fmt.Sprintf("lang/%s", normalizeLanguage(clip.Language)))
    }

    // Store tags on clip
    for _, slug := range slugs {
        if err := s.clipRepo.AddTagBySlug(ctx, clip.ID, slug); err != nil {
            // Log but don't fail — tagging is best-effort
            utils.Warn("Failed to add structural tag", map[string]interface{}{
                "clip_id": clip.ID.String(),
                "tag":     slug,
                "error":   err.Error(),
            })
        }
    }

    return slugs, nil
}

func durationTag(seconds float64) string {
    switch {
    case seconds <= 30:
        return "duration/short"
    case seconds <= 90:
        return "duration/medium"
    default:
        return "duration/long"
    }
}

func normalizeLanguage(lang string) string {
    // Map Twitch language codes to our slugs
    langMap := map[string]string{
        "en": "en", "es": "es", "pt": "pt", "fr": "fr",
        "de": "de", "ru": "ru", "ja": "ja", "ko": "ko",
        "zh": "zh", "it": "it", "tr": "tr", "ar": "ar",
    }
    if mapped, ok := langMap[lang]; ok {
        return mapped
    }
    return "other"
}
```

- [ ] **Step 2: Add AddTagBySlug to ClipRepository**

```go
func (r *ClipRepository) AddTagBySlug(ctx context.Context, clipID uuid.UUID, tagSlug string) error {
    query := `INSERT INTO clip_tags (clip_id, tag_id)
              SELECT $1, id FROM tags WHERE slug = $2
              ON CONFLICT DO NOTHING`
    _, err := r.pool.Exec(ctx, query, clipID, tagSlug)
    return err
}
```

- [ ] **Step 3: Commit**

```bash
git add backend/internal/services/auto_tagger_service.go backend/internal/repository/clip_repository.go
git commit -m "feat: add structural auto-tagger (duration, language) from clip metadata"
```

---

### Task 3.2: Integrate Whisper transcription (reuse hasanara)

**Files:**
- Create: `backend/internal/services/whisper_service.go`
- Create or symlink: `backend/whisper/whisper_runner.py`

- [ ] **Step 1: Symlink hasanara whisper runner**

```bash
cd backend/whisper
ln -s ../../hasanara/worker/whisper_runner.py whisper_runner.py
```

- [ ] **Step 2: Write Go service that calls the Python whisper subprocess**

```go
package services

import (
    "context"
    "encoding/json"
    "fmt"
    "os/exec"
    "path/filepath"
)

type WhisperService struct {
    pythonPath  string
    runnerPath  string
    modelName   string
}

func NewWhisperService(pythonPath, runnerDir, modelName string) *WhisperService {
    return &WhisperService{
        pythonPath: pythonPath,
        runnerPath: filepath.Join(runnerDir, "whisper_runner.py"),
        modelName:  modelName,
    }
}

type WhisperSegment struct {
    Start       float64 `json:"start"`
    End         float64 `json:"end"`
    Text        string  `json:"text"`
    AvgLogprob  float64 `json:"avg_logprob"`
}

type WhisperResult struct {
    Segments []WhisperSegment `json:"segments"`
    Language string           `json:"language"`
    FullText string           `json:"full_text"`
}

// TranscribeAudio transcribes a WAV file and returns the full text
func (s *WhisperService) TranscribeAudio(ctx context.Context, wavPath string) (*WhisperResult, error) {
    // Use a small wrapper script that imports transcribe_chunk and prints JSON
    script := fmt.Sprintf(`
import sys, json
sys.path.insert(0, "%s")
from whisper_runner import transcribe_chunk
segments, lang_info = transcribe_chunk("%s")
full_text = " ".join(s["text"] for s in segments)
result = {
    "segments": segments,
    "language": lang_info.get("language"),
    "full_text": full_text,
}
print(json.dumps(result))
`, filepath.Dir(s.runnerPath), wavPath)

    cmd := exec.CommandContext(ctx, s.pythonPath, "-c", script)
    output, err := cmd.Output()
    if err != nil {
        return nil, fmt.Errorf("whisper transcription failed: %w", err)
    }

    var result WhisperResult
    if err := json.Unmarshal(output, &result); err != nil {
        return nil, fmt.Errorf("parsing whisper output: %w", err)
    }
    return &result, nil
}
```

- [ ] **Step 3: Write test**

```go
func TestWhisperService_TranscribeAudio(t *testing.T) {
    if os.Getenv("WHISPER_TEST") != "1" {
        t.Skip("Skipping Whisper test (set WHISPER_TEST=1 to run)")
    }
    svc := NewWhisperService("python3", "../whisper", "tiny")
    result, err := svc.TranscribeAudio(context.Background(), "testdata/sample.wav")
    require.NoError(t, err)
    assert.NotEmpty(t, result.FullText)
}
```

- [ ] **Step 4: Commit**

```bash
git add backend/internal/services/whisper_service.go backend/whisper/
git commit -m "feat: add Whisper transcription service reusing hasanara whisper_runner.py"
```

---

### Task 3.3: Add thumbnail extraction and vision classification

**Files:**
- Create: `backend/internal/services/thumbnail_service.go`

- [ ] **Step 1: Write thumbnail extraction (ffmpeg)**

```go
package services

import (
    "context"
    "encoding/base64"
    "fmt"
    "os/exec"
    "path/filepath"
)

type ThumbnailService struct {
    ffmpegPath string
    outputDir  string
}

func NewThumbnailService(ffmpegPath, outputDir string) *ThumbnailService {
    return &ThumbnailService{ffmpegPath: ffmpegPath, outputDir: outputDir}
}

// ExtractThumbnails extracts 3 thumbnails at 25%, 50%, 75% of the clip duration
func (s *ThumbnailService) ExtractThumbnails(ctx context.Context, videoPath string, duration float64) ([]string, error) {
    positions := []float64{0.25, 0.50, 0.75}
    thumbnails := make([]string, 0, len(positions))

    for i, pos := range positions {
        timestamp := duration * pos
        outPath := filepath.Join(s.outputDir, fmt.Sprintf("%s_%d.jpg",
            filepath.Base(videoPath), i))

        cmd := exec.CommandContext(ctx, s.ffmpegPath,
            "-ss", fmt.Sprintf("%.1f", timestamp),
            "-i", videoPath,
            "-vframes", "1",
            "-q:v", "2",
            "-y", outPath,
        )
        if err := cmd.Run(); err != nil {
            return nil, fmt.Errorf("extracting thumbnail at %.1fs: %w", timestamp, err)
        }
        thumbnails = append(thumbnails, outPath)
    }
    return thumbnails, nil
}

// ClassifyThumbnails sends thumbnails to a vision model for content classification
func (s *ThumbnailService) ClassifyThumbnails(ctx context.Context, imagePaths []string, gameName string) ([]string, error) {
    // Read images as base64
    images := make([]string, len(imagePaths))
    for i, p := range imagePaths {
        data, err := os.ReadFile(p)
        if err != nil {
            return nil, err
        }
        images[i] = base64.StdEncoding.EncodeToString(data)
    }

    // Send to vision model API
    prompt := fmt.Sprintf(
        `You are analyzing 3 frames from a Twitch clip in the game/category "%s".
        Return a JSON array of 1-3 content tags that apply, chosen from:
        ["clutch", "funny", "fail", "educational", "highlights", "reaction",
         "speedrun", "music", "creative", "irl"]
        Only include tags you're confident about. Example: ["clutch", "highlights"]`,
        gameName,
    )

    // ... API call to GPT-4o-mini or Claude Haiku ...
    // Parse response into []string tags
    return tags, nil
}
```

- [ ] **Step 2: Commit**

```bash
git add backend/internal/services/thumbnail_service.go
git commit -m "feat: add thumbnail extraction (ffmpeg) and vision AI classification service"
```

---

### Task 3.4: Build auto-tagging scheduler (near-real-time)

**Files:**
- Create: `backend/internal/scheduler/auto_tag_scheduler.go`
- Modify: `backend/cmd/api/schedulers.go`

- [ ] **Step 1: Write the scheduler**

```go
package scheduler

import (
    "context"
    "time"

    "git.subcult.tv/subculture-collective/clpr/internal/services"
)

type AutoTagScheduler struct {
    autoTagger    *services.AutoTaggerService
    whisper       *services.WhisperService
    thumbnail     *services.ThumbnailService
    clipRepo      ClipRepository // minimal interface
    interval      time.Duration
    stopChan      chan struct{}
}

func NewAutoTagScheduler(
    autoTagger *services.AutoTaggerService,
    whisper *services.WhisperService,
    thumbnail *services.ThumbnailService,
    clipRepo ClipRepository,
    intervalMinutes int,
) *AutoTagScheduler {
    return &AutoTagScheduler{
        autoTagger: autoTagger,
        whisper:    whisper,
        thumbnail:  thumbnail,
        clipRepo:   clipRepo,
        interval:   time.Duration(intervalMinutes) * time.Minute,
        stopChan:   make(chan struct{}),
    }
}

func (s *AutoTagScheduler) Start(ctx context.Context) {
    ticker := time.NewTicker(s.interval)
    defer ticker.Stop()

    for {
        select {
        case <-ticker.C:
            s.processUntaggedClips(ctx)
        case <-s.stopChan:
            return
        case <-ctx.Done():
            return
        }
    }
}

func (s *AutoTagScheduler) processUntaggedClips(ctx context.Context) {
    // 1. Fetch clips without structural tags
    untagged, err := s.clipRepo.GetUntaggedClips(ctx, 100)
    if err != nil {
        return
    }

    for _, clip := range untagged {
        // 2. Assign structural tags (fast, free)
        s.autoTagger.TagClip(ctx, &clip)

        // 3. Download clip, extract audio, run Whisper
        //    (heavy — consider separate worker pool)
        // wavPath := downloadAndExtractAudio(clip.URL)
        // result := s.whisper.TranscribeAudio(ctx, wavPath)

        // 4. Extract thumbnails, classify with vision
        // thumbs := s.thumbnail.ExtractThumbnails(ctx, videoPath, clip.Duration)
        // contentTags := s.thumbnail.ClassifyThumbnails(ctx, thumbs, clip.GameName)

        // 5. Combine: structural + content tags → store
    }
}
```

- [ ] **Step 2: Register in schedulers.go**

```go
// Start auto-tagging scheduler (runs every 30 seconds)
if svcs.AutoTagger != nil && svcs.Whisper != nil {
    sg.AutoTag = scheduler.NewAutoTagScheduler(
        svcs.AutoTagger, svcs.Whisper, svcs.Thumbnail,
        repos.Clip, 30, // run every 30 seconds
    )
    go sg.AutoTag.Start(context.Background())
}
```

- [ ] **Step 3: Commit**

```bash
git add backend/internal/scheduler/auto_tag_scheduler.go backend/cmd/api/schedulers.go
git commit -m "feat: add auto-tagging scheduler (structural + Whisper + vision pipeline)"
```

---

### Task 3.5: Add game-based structural tags

**Files:**
- Modify: `backend/internal/services/auto_tagger_service.go`

- [ ] **Step 1: Add game-to-genre mapping**

```go
// gameToGenres maps Twitch game IDs (and names) to genre tags
var gameToGenres = map[string][]string{
    "509658": {"game/just-chatting", "game/irl"},
    "33214":  {"game/fortnite", "game/battle-royale", "game/shooter"},
    "516575": {"game/valorant", "game/tactical-shooter", "game/fps"},
    "21779":  {"game/league-of-legends", "game/moba"},
    "511224": {"game/apex-legends", "game/battle-royale", "game/shooter"},
    "29595":  {"game/dota-2", "game/moba"},
    "27471":  {"game/minecraft", "game/sandbox"},
    "32982":  {"game/gta-v", "game/open-world"},
    // ... expand with ~500 popular games
    "1469308723": {"game/software", "game/programming"}, // Software & Game Dev
}
```

- [ ] **Step 2: Add game tags to TagClip**

```go
func (s *AutoTaggerService) TagClip(ctx context.Context, clip *models.Clip) ([]string, error) {
    var slugs []string

    // Game/genre tags from game_id
    if tags, ok := gameToGenres[clip.GameID]; ok {
        slugs = append(slugs, tags...)
    } else if clip.GameName != "" {
        // Auto-generate game tag from game name
        slug := slugify(clip.GameName)
        if err := s.ensureTag(ctx, clip.GameName, slug, "game"); err == nil {
            slugs = append(slugs, "game/"+slug)
        }
    }

    // Duration, language, broadcaster tier...
    slugs = append(slugs, durationTag(clip.Duration))
    slugs = append(slugs, fmt.Sprintf("lang/%s", normalizeLanguage(clip.Language)))

    // Add all tags
    for _, slug := range slugs {
        s.clipRepo.AddTagBySlug(ctx, clip.ID, slug)
    }

    return slugs, nil
}
```

- [ ] **Step 3: Commit**

```bash
git add backend/internal/services/auto_tagger_service.go
git commit -m "feat: add game-to-genre structural tag mapping for auto-tagging"
```

---

## Phase 4: Multi-Tag Playlist Scripts

### Task 4.1: Update PlaylistScript model for multi-tag support

**Files:**
- Modify: `backend/internal/models/models.go`

- [ ] **Step 1: Change Tag from *string to []string**

```go
type PlaylistScript struct {
    // ... existing fields ...
    Tag         *string   `json:"tag,omitempty" db:"tag"`           // DEPRECATED: kept for migration
    Tags        []string  `json:"tags,omitempty" db:"tags"`         // NEW: multiple tags with AND logic
    TagsLogic   string    `json:"tags_logic,omitempty" db:"tags_logic"` // "and" | "or", default "and"
    ExcludeTags []string  `json:"exclude_tags,omitempty" db:"exclude_tags"`
    // ... rest ...
}
```

- [ ] **Step 2: Write migration to add tags[] and tags_logic columns**

```sql
ALTER TABLE playlist_scripts ADD COLUMN IF NOT EXISTS tags TEXT[] DEFAULT '{}';
ALTER TABLE playlist_scripts ADD COLUMN IF NOT EXISTS tags_logic VARCHAR(3) DEFAULT 'and';

-- Migrate existing single tag to tags array
UPDATE playlist_scripts SET tags = ARRAY[tag] WHERE tag IS NOT NULL AND array_length(tags, 1) IS NULL;
```

- [ ] **Step 3: Commit**

```bash
git add backend/internal/models/models.go backend/migrations/
git commit -m "feat: add multi-tag support to PlaylistScript (tags[], tags_logic)"
```

---

### Task 4.2: Update buildFiltersFromScript for multi-tag

**Files:**
- Modify: `backend/internal/services/playlist_script_service.go`

- [ ] **Step 1: Pass Tags into ClipFilters**

```go
func buildFiltersFromScript(script *models.PlaylistScript) *models.ClipFilters {
    filters := &models.ClipFilters{
        Sort:           script.Sort,
        Timeframe:      script.Timeframe,
        GameID:         script.GameID,
        BroadcasterID:  script.BroadcasterID,
        Language:       script.Language,
        ExcludeNSFW:    script.ExcludeNSFW,
        Top10kStreamers: script.Top10kStreamers,
    }

    // Multi-tag filter: pass tags array directly
    if len(script.Tags) > 0 {
        filters.Tags = script.Tags
        // TagsLogic controls AND vs OR — backend defaults to AND
        if script.TagsLogic == "or" {
            filters.TagsLogic = "or"
        }
    }

    if len(script.ExcludeTags) > 0 {
        filters.ExcludeTags = script.ExcludeTags
    }

    return filters
}
```

- [ ] **Step 2: Update ClipFilters model to support tags logic**

```go
type ClipFilters struct {
    // ... existing ...
    Tags        []string `json:"tags,omitempty"`
    TagsLogic   string   `json:"tags_logic,omitempty"` // "and" | "or"
    ExcludeTags []string `json:"exclude_tags,omitempty"`
}
```

- [ ] **Step 3: Update ListWithFilters to support multi-tag AND/OR**

In `backend/internal/repository/clip_repository.go`, handle `TagsLogic`:

```go
// When TagsLogic is "or", use ANY(); when "and" (default), require ALL tags
if filters.TagsLogic == "or" {
    query += ` AND ct.tag_slug = ANY($N)`
} else {
    // AND logic: clip must have ALL specified tags
    query += ` AND (
        SELECT COUNT(DISTINCT ct2.tag_slug) FROM clip_tags ct2
        WHERE ct2.clip_id = c.id AND ct2.tag_slug = ANY($N)
    ) = $M`
}
```

- [ ] **Step 4: Commit**

```bash
git add backend/internal/services/playlist_script_service.go backend/internal/repository/clip_repository.go backend/internal/models/models.go
git commit -m "feat: add multi-tag AND/OR filtering to playlist script clip queries"
```

---

### Task 4.3: Update frontend for multi-tag scripts

**Files:**
- Modify: `frontend/src/types/playlistScript.ts`
- Modify: `frontend/src/lib/playlist-script-form.ts`
- Modify: `frontend/src/components/admin/PlaylistScriptForm.tsx`

- [ ] **Step 1: Update PlaylistScript type**

```typescript
export interface PlaylistScript {
    // ... existing ...
    tag?: string;              // DEPRECATED
    tags: string[];            // NEW
    tags_logic: 'and' | 'or';  // NEW
    exclude_tags: string[];
}
```

- [ ] **Step 2: Update scriptToFormValues**

```typescript
export function scriptToFormValues(script: PlaylistScript): PlaylistScriptFormValues {
    return {
        // ... existing ...
        tags: script.tags || [],
        tags_logic: script.tags_logic || 'and',
        // ...
    };
}
```

- [ ] **Step 3: Add multi-tag selector to PlaylistScriptForm**

Replace the single `tag` input with a tag multi-select using the existing `TagSelector` component:

```tsx
<div className="space-y-2">
    <label className="block text-sm font-medium">Tags (clip must have ALL selected)</label>
    <TagSelector
        value={form.tags}
        onChange={(tags) => setForm({ ...form, tags })}
        placeholder="Select tags to include..."
    />
    {form.tags.length > 0 && (
        <select
            value={form.tags_logic}
            onChange={(e) => setForm({ ...form, tags_logic: e.target.value as 'and' | 'or' })}
            className="mt-1 block w-full rounded-md border-gray-300 text-sm"
        >
            <option value="and">Must have ALL tags (AND)</option>
            <option value="or">Must have ANY tag (OR)</option>
        </select>
    )}
</div>
```

- [ ] **Step 4: Commit**

```bash
git add frontend/src/types/playlistScript.ts frontend/src/lib/playlist-script-form.ts frontend/src/components/admin/PlaylistScriptForm.tsx
git commit -m "feat: add multi-tag selector with AND/OR logic to playlist script form"
```

---

## Phase 5: Discovery List Merge

### Task 5.1: Add is_curated and is_featured to playlists

**Files:**
- Create: `backend/migrations/NNNNN_add_playlist_curation_flags.sql`
- Modify: `backend/internal/models/models.go`

- [ ] **Step 1: Write migration**

```sql
-- Add discovery list fields to playlists table
ALTER TABLE playlists ADD COLUMN IF NOT EXISTS is_curated BOOLEAN DEFAULT false;
ALTER TABLE playlists ADD COLUMN IF NOT EXISTS is_featured BOOLEAN DEFAULT false;
ALTER TABLE playlists ADD COLUMN IF NOT EXISTS display_order INTEGER DEFAULT 0;
ALTER TABLE playlists ADD COLUMN IF NOT EXISTS script_id UUID REFERENCES playlist_scripts(id);

-- Migrate existing discovery lists into playlists
INSERT INTO playlists (id, user_id, title, description, visibility, is_curated, is_featured, display_order, created_at, updated_at)
SELECT id, COALESCE(created_by, '00000000-0000-0000-0000-000000000001'),
       name, description, 'public', true, is_featured, display_order,
       created_at, updated_at
FROM discovery_lists
ON CONFLICT (id) DO NOTHING;

-- Migrate discovery list clips into playlist_items
INSERT INTO playlist_items (id, playlist_id, clip_id, display_order, added_at)
SELECT gen_random_uuid(), list_id, clip_id, display_order, added_at
FROM discovery_list_clips
ON CONFLICT DO NOTHING;
```

- [ ] **Step 2: Run migration**

```bash
cd backend && go run cmd/migrate/main.go up
```

- [ ] **Step 3: Update site freshness to use is_curated**

In `backend/internal/services/playlist_script_service.go`, the `siteFreshnessDisplayOrder` map now checks for `is_curated`:

```go
func generatedPlaylistPresentationForScript(script *models.PlaylistScript, ownerID uuid.UUID) generatedPlaylistPresentation {
    // Scripts owned by the bot user with certain names get curated/featured
    isBot := ownerID == BotUserID
    isCurated := isBot && script.Name != "" && siteFreshnessDisplayOrder[script.Name] > 0
    isFeatured := isCurated // curated playlists are featured by default

    return generatedPlaylistPresentation{
        isCurated:    isCurated,
        isFeatured:   isFeatured,
        displayOrder: siteFreshnessDisplayOrder[script.Name],
    }
}
```

- [ ] **Step 4: Commit**

```bash
git add backend/migrations/ backend/internal/models/models.go backend/internal/services/playlist_script_service.go
git commit -m "feat: merge discovery lists into playlists with is_curated/is_featured flags"
```

---

### Task 5.2: Update frontend to treat curated playlists as discovery lists

**Files:**
- Modify: `frontend/src/pages/DiscoveryListsPage.tsx`

- [ ] **Step 1: Update DiscoveryListsPage to query curated playlists**

Instead of calling `discoveryListApi.listDiscoveryLists()`, query for playlists with `is_curated: true`:

```typescript
// Replace discovery list API call with playlist query
const { data: curatedPlaylists, isLoading } = useQuery({
    queryKey: ['curated-playlists'],
    queryFn: () => playlistApi.listPlaylists({ is_curated: true, is_featured: true }),
});
```

- [ ] **Step 2: Keep old discovery list API for backward compat**

Add a deprecation note and proxy to playlist queries. Remove after next release cycle.

- [ ] **Step 3: Update AdminDiscoveryListFormPage to use PlaylistScriptForm**

Redirect the discovery list admin form to use the playlist script form with `is_curated` preset to true.

- [ ] **Step 4: Commit**

```bash
git add frontend/src/pages/DiscoveryListsPage.tsx frontend/src/pages/admin/AdminDiscoveryListFormPage.tsx
git commit -m "feat: redirect discovery list UI to curated playlist queries"
```

---

## Phase 6: User Tag Promotion (Option A)

### Task 6.1: Add tag_promotion queue table

**Files:**
- Create: `backend/migrations/NNNNN_add_tag_promotion_queue.sql`

- [ ] **Step 1: Write migration**

```sql
CREATE TABLE IF NOT EXISTS tag_promotion_queue (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tag_slug VARCHAR(100) NOT NULL,
    usage_count INTEGER NOT NULL DEFAULT 0,
    unique_users INTEGER NOT NULL DEFAULT 0,
    status VARCHAR(20) NOT NULL DEFAULT 'pending', -- pending, approved, rejected
    reviewed_by UUID REFERENCES users(id),
    reviewed_at TIMESTAMPTZ,
    promoted_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX idx_tag_promotion_slug_status
    ON tag_promotion_queue(tag_slug) WHERE status = 'pending';

-- View: tags that have crossed the promotion threshold
CREATE VIEW tag_promotion_candidates AS
SELECT
    t.slug,
    t.name,
    COUNT(DISTINCT ct.clip_id) AS clip_count,
    COUNT(DISTINCT ct.user_id) AS unique_users,
    t.parent_slug
FROM clip_tags ct
JOIN tags t ON t.slug = ct.tag_slug
WHERE t.parent_slug IS NULL OR t.parent_slug = 'community'
GROUP BY t.slug, t.name, t.parent_slug
HAVING COUNT(DISTINCT ct.user_id) >= 3   -- threshold: 3 unique users
   AND COUNT(DISTINCT ct.clip_id) >= 5;  -- threshold: used on 5+ clips
```

- [ ] **Step 2: Create TagPromotionService**

```go
package services

type TagPromotionService struct {
    tagRepo  *repository.TagRepository
    queueRepo *repository.TagPromotionRepository
}

func (s *TagPromotionService) CheckPromotionCandidates(ctx context.Context) ([]string, error) {
    // Query tag_promotion_candidates view
    // For each candidate that isn't already queued, insert into tag_promotion_queue
    // Return slugs that were newly queued
}

func (s *TagPromotionService) ApprovePromotion(ctx context.Context, slug string, reviewerID uuid.UUID) error {
    // Move tag from community/ to content/ (or other appropriate parent)
    // Mark queue entry as approved
    // The tag is now available for AI to assign
}
```

- [ ] **Step 3: Commit**

```bash
git add backend/migrations/ backend/internal/services/tag_promotion_service.go
git commit -m "feat: add user tag promotion queue with threshold gate and moderator review"
```

---

### Task 6.2: Add admin moderation UI for tag promotion

**Files:**
- Create: `frontend/src/pages/admin/AdminTagPromotionPage.tsx`

- [ ] **Step 1: Write the admin page**

```tsx
export function AdminTagPromotionPage() {
    const { data: queue, isLoading } = useQuery({
        queryKey: ['admin', 'tag-promotion'],
        queryFn: () => tagApi.getPromotionQueue(),
    });

    return (
        <Container>
            <h1>Tag Promotion Queue</h1>
            {queue?.map(item => (
                <div key={item.id} className="flex justify-between p-4 border rounded">
                    <div>
                        <span className="font-bold">#{item.tag_slug}</span>
                        <span className="text-sm text-muted ml-2">
                            {item.unique_users} users, {item.clip_count} clips
                        </span>
                    </div>
                    <div className="space-x-2">
                        <Button onClick={() => approveTag(item.tag_slug)}>Promote to AI Pool</Button>
                        <Button variant="secondary" onClick={() => rejectTag(item.tag_slug)}>Reject</Button>
                    </div>
                </div>
            ))}
        </Container>
    );
}
```

- [ ] **Step 2: Add route to AdminRoutes**

```tsx
tagPromotion: AdminTagPromotionPage,
```

- [ ] **Step 3: Commit**

```bash
git add frontend/src/pages/admin/AdminTagPromotionPage.tsx frontend/src/routes/v1/AdminRoutes.tsx
git commit -m "feat: add admin tag promotion queue UI for moderating community tags"
```

---

## Integration: Wiring Everything Together

### Task I.1: Update clip sync pipeline to trigger auto-tagging

**Files:**
- Modify: `backend/internal/services/clip_sync_service.go`

- [ ] **Step 1: After clip creation, enqueue for auto-tagging**

```go
func (s *ClipSyncService) ingestClips(ctx context.Context, twitchClips []twitch.Clip) (*SyncStats, error) {
    for _, tc := range twitchClips {
        clip := s.twitchClipToModel(tc)
        created, err := s.clipRepo.Upsert(ctx, clip)
        if err != nil {
            stats.Errors = append(stats.Errors, err.Error())
            continue
        }
        if created {
            stats.ClipsCreated++
            // Enqueue for auto-tagging (non-blocking)
            go func() {
                if s.autoTagger != nil {
                    s.autoTagger.TagClip(context.Background(), clip)
                }
            }()
        }
    }
    return stats, nil
}
```

- [ ] **Step 2: Commit**

```bash
git add backend/internal/services/clip_sync_service.go
git commit -m "feat: enqueue new clips for auto-tagging after ingestion"
```

---

---

## Execution Order

```
Phase 1 (Tag Hierarchy)
    │
    ├──► Task 1.1: DB migration + models
    ├──► Task 1.2: Tag API handlers
    ├──► Task 1.3: Seed hierarchy
    └──► Task 1.4: Frontend types
         │
         ▼
Phase 2 (Expanded Ingestion)          Phase 3 (Auto-Tagging)
    │                                      │
    ├──► Task 2.1: Top 50 categories        ├──► Task 3.1: Structural tagger
    ├──► Task 2.2: Followed broadcasters     ├──► Task 3.2: Whisper integration
    └──► Task 2.3: Global trending           ├──► Task 3.3: Thumbnails + vision
                                             ├──► Task 3.4: Tagging scheduler
                                             └──► Task 3.5: Game-to-genre mapping
         │                                      │
         └──────────────┬───────────────────────┘
                        ▼
Phase 4 (Multi-Tag Scripts)           Phase 5 (Discovery Merge)
    │                                      │
    ├──► Task 4.1: Model + migration        ├──► Task 5.1: DB migration
    ├──► Task 4.2: Backend filtering         └──► Task 5.2: Frontend redirect
    └──► Task 4.3: Frontend form
         │                                      │
         └──────────────┬───────────────────────┘
                        ▼
               Phase 6 (Tag Promotion)
                    │
                    ├──► Task 6.1: Queue table + service
                    └──► Task 6.2: Admin UI
                         │
                         ▼
               Integration Tasks
                    │
                    └──► Task I.1: Wire clip sync → auto-tagger
```

---

## Verification Checklist

After all phases are complete, verify:

- [ ] Structural tags appear on newly synced clips within 30 seconds
- [ ] Whisper transcription produces reasonable output for English clips
- [ ] Vision model correctly classifies obvious clip content (FPS kill feed → clutch)
- [ ] Playlist scripts with `tags: ["content/clutch", "game/valorant"]` generate playlists
- [ ] Curated playlists appear on homepage under their display_order names
- [ ] Discovery lists page shows the same content from curated playlists
- [ ] User tag `#satisfying` used by 3+ users on 5+ clips appears in admin queue
- [ ] Top 50 categories include non-gaming content (Just Chatting, Music, Art)
- [ ] Followed broadcaster clips respect the 200-per-cycle cap
- [ ] Clip ingestion continues if Twitch API is briefly unavailable (fallback to hardcoded)