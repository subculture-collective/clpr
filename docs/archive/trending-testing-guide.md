---
title: Feed Sorting & Trending Algorithms - Testing Guide
summary: This document outlines how to test the new trending and popular sorting algorithms implemented for the feed.
tags: ["testing", "archive", "implementation", "guide"]
area: docs
status: archived
owner: team-core
version: "1.0"
last_reviewed: 2026-01-29
---

# Feed Sorting & Trending Algorithms - Testing Guide

This document outlines how to test the new trending and popular sorting algorithms implemented for the feed.

## Backend Testing

### Unit Tests

Run the following to test the scheduler and repository:
```bash
cd backend
go test ./internal/scheduler/trending_score_scheduler_test.go ./internal/scheduler/trending_score_scheduler.go -v
go test ./internal/repository/... -v
go test ./internal/models/... -v
```

### Database Migration Testing

1. Start the database (if using Docker):
```bash
docker-compose up -d postgres
```

2. Run migrations:
```bash
migrate -path backend/migrations -database "postgresql://clpr:clpr_password@localhost:5432/clpr?sslmode=disable" up
```

3. Verify the new columns exist:
```sql
-- Connect to database
psql -h localhost -U clpr -d clpr

-- Check for new columns
\d clips

-- Should show:
-- trending_score | double precision | | default 0
-- hot_score | double precision | | default 0
-- popularity_index | integer | | default 0
-- engagement_count | integer | | default 0

-- Verify functions exist
\df calculate_trending_score
\df calculate_hot_score_value

-- Check indexes
\di idx_clips_trending_score
\di idx_clips_hot_score
\di idx_clips_popularity_index
\di idx_clips_engagement_count
```

### API Endpoint Testing

Test the trending and popular sort options with curl or Postman:

```bash
# Test trending sort
curl "http://localhost:8080/api/v1/feeds/clips?sort=trending&limit=10"

# Test trending with time window
curl "http://localhost:8080/api/v1/feeds/clips?sort=trending&timeframe=week&limit=10"

# Test popular sort
curl "http://localhost:8080/api/v1/feeds/clips?sort=popular&limit=10"

# Test hot sort (existing)
curl "http://localhost:8080/api/v1/feeds/clips?sort=hot&limit=10"

# Test new sort (existing)
curl "http://localhost:8080/api/v1/feeds/clips?sort=new&limit=10"
```

Expected response structure:
```json
{
  "clips": [
    {
      "id": "...",
      "title": "...",
      "trending_score": 123.45,
      "hot_score": 98.76,
      "popularity_index": 1500,
      "engagement_count": 2000,
      ...
    }
  ],
  "pagination": {
    "limit": 10,
    "offset": 0,
    "total": 100,
    "has_more": true
  }
}
```

### Scheduler Testing

To test the trending score scheduler:

1. Check logs for scheduler startup:
```bash
# Look for this in the logs
grep "Starting trending score scheduler" backend/logs/app.log
```

2. Manually trigger score updates:
```sql
-- Run the update function manually
SELECT COUNT(*) FROM clips WHERE is_removed = false AND is_hidden = false;

-- Update scores
UPDATE clips
SET 
    engagement_count = view_count + (vote_score * 2) + (comment_count * 3) + (favorite_count * 2),
    trending_score = calculate_trending_score(view_count, vote_score, comment_count, favorite_count, created_at),
    hot_score = calculate_trending_score(view_count, vote_score, comment_count, favorite_count, created_at),
    popularity_index = view_count + (vote_score * 2) + (comment_count * 3) + (favorite_count * 2)
WHERE is_removed = false AND is_hidden = false;

-- Verify scores were updated
SELECT id, title, trending_score, hot_score, popularity_index, engagement_count 
FROM clips 
ORDER BY trending_score DESC 
LIMIT 10;
```

## Frontend Testing

### Manual Browser Testing

1. Start the frontend dev server:
```bash
cd frontend
npm run dev
```

2. Navigate to the feed page (usually <http://localhost:5173>)

3. Test the sort dropdown:
   - Select "Trending 🔥" from the sort dropdown
   - Verify the timeframe selector appears
   - Select different timeframes (Past Hour, Past Day, Past Week, etc.)
   - Verify clips are sorted correctly
   - Check browser console for errors

4. Test "Most Popular ⭐":
   - Select "Most Popular ⭐" from the sort dropdown
   - Verify clips with high engagement appear first
   - Check that timeframe selector does NOT appear (it's only for top and trending)

5. Test localStorage persistence:
   - Select "Trending 🔥"
   - Refresh the page
   - Open browser DevTools > Application > Local Storage
   - Verify `feedSort` key exists with value "trending"
   - Refresh page again and verify "Trending" is still selected

6. Test URL parameters:
   - Manually navigate to: `http://localhost:5173/?sort=trending&timeframe=week`
   - Verify the correct sort and timeframe are selected
   - Change sort to "Popular" and verify URL updates

### Visual Testing Checklist

- [ ] Sort dropdown shows all options including "Trending 🔥" and "Most Popular ⭐"
- [ ] Timeframe selector appears for "Trending" and "Top" sorts
- [ ] Timeframe selector does NOT appear for "Popular", "Hot", "New", "Rising", "Discussed"
- [ ] Page title updates correctly: "Trending — Past Day Feed", "Most Popular Feed", etc.
- [ ] Sort preference persists after page refresh
- [ ] No console errors when switching between sorts
- [ ] Clips load correctly for all sort options

## Performance Testing

### Query Performance

Test that trending queries execute within acceptable time:

```sql
-- Test trending query performance
EXPLAIN ANALYZE
SELECT *
FROM clips c
WHERE c.is_removed = false AND c.is_hidden = false
ORDER BY COALESCE(c.trending_score, calculate_trending_score(c.view_count, c.vote_score, c.comment_count, c.favorite_count, c.created_at)) DESC, c.created_at DESC
LIMIT 20;

-- Should use idx_clips_trending_score index and complete in < 200ms

-- Test popular query performance
EXPLAIN ANALYZE
SELECT *
FROM clips c
WHERE c.is_removed = false AND c.is_hidden = false
ORDER BY COALESCE(c.popularity_index, c.engagement_count, (c.view_count + c.vote_score * 2 + c.comment_count * 3 + c.favorite_count * 2)) DESC, c.created_at DESC
LIMIT 20;

-- Should use idx_clips_popularity_index index and complete in < 200ms
```

### Load Testing

If k6 is available, test concurrent requests:

```bash
# Test trending endpoint with 100 concurrent users
k6 run -u 100 -d 30s backend/tests/load/scenarios/feed_browsing.js
```

## Known Issues & Limitations

1. **Initial Score Calculation**: Scores are initialized to 0 for new clips and updated hourly. Very recent clips (< 1 hour old) may not have accurate trending scores yet.

2. **Timeframe Filtering**: The timeframe parameter for trending currently works the same as for "top" sort - it filters clips by creation date, not by trending score calculation window.

3. **Real-time Updates**: Scores are updated hourly by the scheduler, not in real-time. For real-time trending, consider implementing a materialized view refresh or more frequent updates.

## Success Criteria

- [x] Database migration runs without errors
- [x] All new columns and indexes are created
- [x] SQL functions for score calculation work correctly
- [x] Backend builds without errors
- [x] Unit tests pass for scheduler
- [ ] API returns correct data for trending and popular sorts
- [ ] Frontend displays new sort options correctly
- [ ] localStorage persists user sort preference
- [ ] Query performance meets p95 < 200ms target
- [ ] No console errors in browser
- [ ] Scheduler updates scores hourly in production
