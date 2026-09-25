---
title: 'Clip Serving Strategy'
summary: 'Documentation for how clips are served across different pages - user-submitted vs. scraped content'
tags: ['product', 'architecture']
area: 'product'
status: 'stable'
owner: 'team-core'
version: '1.0'
last_reviewed: 2025-12-13
---

# Clip Serving Strategy

## Overview

This document outlines the strategy for serving clips across different pages of the application. The key distinction is between **user-submitted content** (clips that have been posted by community members) and **scraped content** (clips automatically discovered from Twitch).

## Content Types

### User-Submitted Clips

- Clips where `submitted_by_user_id IS NOT NULL`
- Content that has been curated and posted by community members
- Appears in main feed pages by default
- Represents community-curated, high-quality content

### Scraped Clips

- Clips where `submitted_by_user_id IS NULL`
- Content automatically discovered from Twitch
- Available for users to submit/claim
- Used for discovery and content sourcing

## Page-by-Page Breakdown

### Main Feed Pages (User-Submitted Only)

The following pages show **only user-submitted clips**:

| Page       | Route     | Sort     | Description                        |
| ---------- | --------- | -------- | ---------------------------------- |
| **Home**   | `/`       | `hot`    | Community-curated trending content |
| **New**    | `/new`    | `new`    | Latest community submissions       |
| **Top**    | `/top`    | `top`    | Highest-rated community content    |
| **Rising** | `/rising` | `rising` | Trending upward community posts    |

**Implementation:**

- Backend: `GET /api/v1/clips` with default `show_all_clips=false`
- Filter: `UserSubmittedOnly: true` in `ClipFilters`
- Query: `WHERE submitted_by_user_id IS NOT NULL`

### Discovery Page (All Content)

The Discovery page shows **all clips** (both user-submitted and scraped):

| Page          | Route       | Sort Options              | Description                          |
| ------------- | ----------- | ------------------------- | ------------------------------------ |
| **Discovery** | `/discover` | `top`, `new`, `discussed` | All available clips from all sources |

**Purpose:**

- Allow users to explore all available content
- Find interesting scraped clips to submit
- Discover trending content from Twitch

**Implementation:**

- Backend: `GET /api/v1/clips?show_all_clips=true`
- Filter: `UserSubmittedOnly: false` in `ClipFilters`
- Query: No filter on `submitted_by_user_id`

### Scraped Clips Page (Scraped Only)

The Scraped Clips page shows **only scraped clips**:

| Page        | Route               | Sort Options               | Description                 |
| ----------- | ------------------- | -------------------------- | --------------------------- |
| **Scraped** | `/discover/scraped` | `trending`, `new`, `views` | Only unclaimed Twitch clips |

**Implementation:**

- Backend: `GET /api/v1/scraped-clips`
- Filter: Dedicated endpoint for scraped content
- Query: `WHERE submitted_by_user_id IS NULL`

## API Implementation

### Backend Query Parameter

```
GET /api/v1/clips?show_all_clips=true
```

| Parameter        | Type    | Default | Description                                                 |
| ---------------- | ------- | ------- | ----------------------------------------------------------- |
| `show_all_clips` | boolean | `false` | When `true`, includes both user-submitted and scraped clips |

### Code Structure

**Repository Layer** (`backend/internal/repository/clip_repository.go`):

```go
type ClipFilters struct {
    // ... other fields
    UserSubmittedOnly bool // If true, only show clips with submitted_by_user_id IS NOT NULL
}
```

**Handler Layer** (`backend/internal/handlers/clip_handler.go`):

```go
filters := repository.ClipFilters{
    Sort:              sort,
    UserSubmittedOnly: !showAllClips, // Only show user-submitted unless explicitly requesting all
}
```

**Frontend Type** (`frontend/src/types/clip.ts`):

```typescript
export interface ClipFeedFilters {
    // ... other fields
    show_all_clips?: boolean; // If true, show both user-submitted and scraped clips
}
```

## User Experience

### Main Feed Pages

- **What users see:** High-quality, community-curated content
- **Why:** These clips have been vetted and submitted by community members
- **Benefit:** Better signal-to-noise ratio, trusted content

### Discovery Page

- **What users see:** All available clips from all sources
- **Why:** Maximum content discovery including trending Twitch clips
- **Benefit:** Users can find and submit great content that hasn't been shared yet

### Scraped Clips Page

- **What users see:** Raw Twitch clips not yet submitted
- **Why:** Dedicated space for discovering submittable content
- **Benefit:** Clear call-to-action to "Post This Clip"

## Migration Notes

This strategy was implemented to create a clearer distinction between:

1. **Curated Community Content** - Main feeds show only what the community has chosen to share
2. **Discovery Content** - Discovery page shows everything, helping users find new clips to share
3. **Source Content** - Scraped page shows the raw pool of clips available for submission

### Previous Behavior

- All pages showed mixed content (both user-submitted and scraped)
- No distinction between community-curated and auto-discovered content

### New Behavior (Current)

- Main feeds are user-submitted only (home, hot, new, top, rising)
- Discovery page shows all content
- Clear separation improves content quality perception

## Testing

### Verify User-Submitted Only

```bash
# Should return only clips with submitted_by_user_id
curl "http://localhost:8080/api/v1/clips?sort=hot"

# Verify by checking response clips have submitted_by field
```

### Verify All Content (Discovery)

```bash
# Should return both user-submitted and scraped clips
curl "http://localhost:8080/api/v1/clips?sort=hot&show_all_clips=true"

# Mix of clips with and without submitted_by field
```

### Verify Scraped Only

```bash
# Should return only clips without submitted_by_user_id
curl "http://localhost:8080/api/v1/scraped-clips?sort=new"

# All clips should have submitted_by: null
```

## Future Considerations

1. **Metrics to Track:**
    - Ratio of user-submitted to scraped clips in discovery
    - Conversion rate: discovery views → clip submissions
    - User engagement on main feeds vs. discovery

2. **Potential Enhancements:**
    - Allow users to opt-in to see scraped content in main feeds
    - Add "recommended for you" section with scraped clips
    - Personalized discovery based on user preferences

3. **Content Quality:**
    - Monitor if filtering improves perceived content quality
    - Track user feedback on main feed vs. discovery page
    - Adjust strategy based on user behavior

## Related Documentation

- [Clip Management API](../backend/clip-api.md)
- [Scraped Clips Feature](./scraped-clips.md)
- Discovery Lists Feature
