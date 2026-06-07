---
title: Recommendation Engine Implementation
summary: This document describes the implementation of the discovery algorithm and recommendation engine for the Clipper platform, as outlined in issue #668.
tags: ["archive", "implementation"]
area: docs
status: archived
owner: team-core
version: "1.0"
last_reviewed: 2026-01-29
---

# Recommendation Engine Implementation

## Overview

This document describes the implementation of the discovery algorithm and recommendation engine for the Clipper platform, as outlined in issue #668.

## Architecture

The recommendation system consists of three main layers:

### 1. Database Layer (`migrations/000055_add_recommendation_system.up.sql`)

Two new tables were added:

- **`user_preferences`**: Stores user preferences for games, streamers, tags, and categories
  - `user_id`: UUID reference to users table
  - `favorite_games`: Array of game IDs
  - `followed_streamers`: Array of broadcaster IDs
  - `preferred_categories`: Array of category names
  - `preferred_tags`: Array of tag UUIDs
  - Includes function to auto-populate from user interactions

- **`user_clip_interactions`**: Tracks all user interactions with clips
  - `user_id`, `clip_id`: References to users and clips
  - `interaction_type`: 'view', 'like', 'share', 'dwell'
  - `dwell_time`: Time spent watching (in seconds)
  - `timestamp`: When the interaction occurred

### 2. Repository Layer (`internal/repository/recommendation_repository.go`)

Handles all database operations:

- **Preference Management**
  - `GetUserPreferences()`: Retrieves user preferences
  - `UpdateUserPreferences()`: Updates or creates preferences
  - `UpdateUserPreferencesFromInteractions()`: Auto-learns from interactions

- **Recommendation Queries**
  - `GetContentBasedRecommendations()`: Content-based filtering
  - `GetCollaborativeRecommendations()`: Collaborative filtering
  - `GetTrendingClips()`: Trending clips for cold start
  - `GetClipsByIDs()`: Retrieves clips while preserving order

- **Interaction Tracking**
  - `RecordInteraction()`: Records user interactions
  - `HasUserInteractions()`: Checks if user has history

### 3. Service Layer (`internal/services/recommendation_service.go`)

Implements the recommendation algorithms:

#### Content-Based Filtering

- Recommends clips similar to user's preferences
- Weights: 50% game match, 30% streamer match, 20% popularity
- Uses user's favorite games and followed streamers

#### Collaborative Filtering

- Finds similar users based on shared likes
- Recommends clips liked by similar users
- Uses Jaccard similarity for user matching

#### Hybrid Algorithm

- Combines multiple signals with weighted scoring:
  - Content-based: 50%
  - Collaborative: 30%
  - Trending: 20%
- Enforces game diversity (max 3 consecutive clips from same game)
- Provides best overall recommendations

#### Cold Start Handling

- Detects users without interaction history
- Falls back to trending clips
- Ensures new users get quality content immediately

### 4. Handler Layer (`internal/handlers/recommendation_handler.go`)

Exposes HTTP endpoints:

- `GET /api/v1/recommendations/clips`
  - Query params: `limit` (default 20), `algorithm` (content/collaborative/hybrid/trending)
  - Returns personalized recommendations with scores and reasons
  - Rate limited: 60 requests/minute

- `POST /api/v1/recommendations/feedback`
  - Body: `clip_id`, `feedback_type` (positive/negative)
  - Records user feedback
  - Rate limited: 100 requests/minute

- `GET /api/v1/recommendations/preferences`
  - Returns user's current preferences
  - No rate limit (read-only)

- `PUT /api/v1/recommendations/preferences`
  - Body: `favorite_games`, `followed_streamers`, `preferred_categories`, `preferred_tags`
  - Updates user preferences
  - Rate limited: 10 requests/minute

- `POST /api/v1/recommendations/track-view/:id`
  - Body: `dwell_time` (optional)
  - Tracks clip views for better recommendations
  - Rate limited: 200 requests/minute

## Key Features

### 1. Intelligent Scoring

The hybrid algorithm combines multiple signals:

```go
merged_score = (content_score * 0.5) + 
               (collaborative_score * 0.3) + 
               (trending_score * 0.2)
```

### 2. Game Diversity

Prevents repetitive recommendations by:
- Limiting consecutive clips from same game to 3
- Ensuring variety across different games
- Maintaining score-based ordering when possible

### 3. Caching

- Redis cache with 24-hour TTL
- Cache key includes: `user_id`, `algorithm`, `limit`
- Automatic cache invalidation on preference updates
- Improves performance for repeated requests

### 4. Reason Generation

Provides context for each recommendation:
- "Because you liked clips in Valorant"
- "Because you watched shroud"
- "Popular with users like you"
- "Trending now"

### 5. Performance Optimization

- Efficient PostgreSQL queries with proper indexing
- GIN indexes on array fields
- Composite indexes for common query patterns
- Query result caching in Redis
- Target: p95 < 300ms for 100k+ users

## API Examples

### Get Recommendations

```bash
GET /api/v1/recommendations/clips?limit=20&algorithm=hybrid
Authorization: Bearer <token>
```

Response:
```json
{
  "recommendations": [
    {
      "id": "clip-123",
      "title": "Amazing 1v5 Clutch",
      "game_name": "Valorant",
      "broadcaster_name": "TenZ",
      "score": 0.89,
      "reason": "Because you liked clips in Valorant",
      "algorithm": "hybrid"
    }
  ],
  "metadata": {
    "algorithm_used": "hybrid",
    "diversity_applied": true,
    "cold_start": false,
    "cache_hit": false,
    "processing_time_ms": 145
  }
}
```

### Submit Feedback

```bash
POST /api/v1/recommendations/feedback
Authorization: Bearer <token>
Content-Type: application/json

{
  "clip_id": "clip-123",
  "feedback_type": "positive",
  "algorithm": "hybrid",
  "score": 0.89
}
```

### Update Preferences

```bash
PUT /api/v1/recommendations/preferences
Authorization: Bearer <token>
Content-Type: application/json

{
  "favorite_games": ["game-1", "game-2"],
  "followed_streamers": ["streamer-1", "streamer-2"],
  "preferred_categories": ["fps", "competitive"]
}
```

## Testing

### Unit Tests

Seven comprehensive unit tests cover:
- Game diversity enforcement
- Score merging and ranking
- Reason generation
- Edge cases (empty lists, small inputs)

All tests pass successfully.

### Test Coverage

- `TestEnforceGameDiversity`: Validates diversity rules
- `TestMergeAndRank`: Tests hybrid algorithm scoring
- `TestGenerateReason`: Tests reason text generation
- Edge case tests for robustness

## Database Migrations

To apply the migration:

```bash
# Development
make migrate-up

# Production
migrate -path backend/migrations \
  -database "postgresql://user:pass@host:5432/clpr?sslmode=disable" \
  up
```

To rollback:

```bash
migrate -path backend/migrations \
  -database "postgresql://user:pass@host:5432/clpr?sslmode=disable" \
  down 1
```

## Performance Considerations

### Query Optimization

1. **Indexes**: All foreign keys and array fields are indexed
2. **Limit Early**: Queries fetch 2x limit then apply diversity
3. **Materialized Views**: Could be added for pre-computed similarities
4. **Connection Pooling**: Uses pgx pool for efficient connections

### Caching Strategy

1. **Cache Key**: `recommendations:{user_id}:{algorithm}:{limit}`
2. **TTL**: 24 hours
3. **Invalidation**: On preference updates or feedback
4. **Warm-up**: First request populates cache

### Scalability

- Queries designed for 100k+ users
- Redis reduces database load
- Background jobs can pre-compute recommendations
- Horizontal scaling supported (stateless design)

## Future Enhancements

### Phase 2 (Not in this PR)

- Real-time recommendation updates
- Social recommendations (friends' likes)
- Category-based exploration
- A/B testing framework
- Machine learning models
- Personalized ranking weights

### Monitoring & Analytics

- Track CTR (click-through rate)
- Monitor dwell time
- Measure diversity scores
- A/B test algorithm variants
- User feedback analysis

## Configuration

No additional configuration needed. The system uses:
- Existing database connection
- Existing Redis connection
- Default algorithm weights (can be adjusted in service)

## Security

- All endpoints require authentication
- Rate limiting prevents abuse
- Input validation on all endpoints
- SQL injection prevention via parameterized queries
- No sensitive data in cache keys

## Maintenance

### Regular Tasks

1. Monitor cache hit rate
2. Review recommendation quality metrics
3. Analyze user feedback
4. Optimize slow queries
5. Update algorithm weights based on data

### Troubleshooting

- **Low diversity**: Check game distribution in trending clips
- **Poor recommendations**: Verify user has interaction history
- **Slow queries**: Check database indexes and connection pool
- **Cache issues**: Verify Redis connection and memory

## References

- Issue: #668 - Home Page & Feed Filtering
- Epic: Feed Filtering
- Migration: `000055_add_recommendation_system.up.sql`
- Code: `internal/services/recommendation_service.go`
