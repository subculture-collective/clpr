---
title: "Query Limits"
summary: "To prevent resource exhaustion attacks and ensure system stability, Clipper implements comprehensive query cost analysis and complexity limits for both database queries and search operations."
tags: ["operations"]
area: "operations"
status: "stable"
owner: "team-core"
version: "1.0"
last_reviewed: 2026-01-29
---

# Query Cost Analysis and Complexity Limits

## Overview

To prevent resource exhaustion attacks and ensure system stability, Clipper implements comprehensive query cost analysis and complexity limits for both database queries and search operations.

## Database Query Limits

### Configuration

Database query limits can be configured via environment variables:

```bash
# Maximum number of rows per query (default: 1000)
QUERY_MAX_RESULT_SIZE=1000

# Maximum pagination offset (default: 1000)
QUERY_MAX_OFFSET=1000

# Maximum number of JOIN clauses (default: 3)
QUERY_MAX_JOIN_DEPTH=3

# Maximum query execution time in seconds (default: 10)
QUERY_MAX_TIME_SEC=10
```

### Enforced Limits

| Limit | Default Value | Purpose |
|-------|---------------|---------|
| Max Result Size | 1000 rows | Prevents large result sets that consume excessive memory |
| Max Offset | 1000 | Discourages inefficient offset-based pagination |
| Max Join Depth | 3 | Limits query complexity to prevent expensive joins |
| Max Query Time | 10 seconds | Prevents long-running queries from blocking resources |

### Query Cost Analysis

The system includes infrastructure for analyzing SQL queries for:

- **Join Count**: Number of JOIN clauses in the query
- **Scan Size**: Estimated number of rows to scan
- **Complexity Score**: Overall complexity based on multiple factors
- **Offset Usage**: Presence and size of OFFSET clauses

**Current Enforcement Status**:
- ✅ **Pagination Limits**: Result size and offset limits are enforced
- ✅ **Timeouts**: Query execution timeouts are enforced
- ⚠️ **Join Depth**: Analysis is implemented but not yet integrated into execution paths

Queries can be validated using the `QueryCostAnalyzer`:

```go
analyzer := NewQueryCostAnalyzer(limits)
cost, err := analyzer.AnalyzeQuery(query)
// Can check join count, complexity, etc.
```

To enforce join depth limits in production, integrate the `ExecuteWithTimeout` method into repository query execution.

### Pagination Best Practices

#### ❌ Avoid: Offset-Based Pagination (Inefficient)

```go
// Bad: Offset-based pagination becomes expensive at large offsets
clips, err := repo.List(ctx, 100, 5000) // Scans and skips 5000 rows
```

#### ✅ Prefer: Cursor-Based Pagination (Efficient)

```go
// Good: Cursor-based pagination using WHERE clause
query := `
  SELECT * FROM clips
  WHERE (created_at, id) > ($1, $2)
  ORDER BY created_at, id
  LIMIT $3
`
```

### Repository Integration

The `RepositoryHelper` automatically enforces limits:

```go
type ClipRepository struct {
    pool   *pgxpool.Pool
    helper *RepositoryHelper
}

func (r *ClipRepository) List(ctx context.Context, limit, offset int) ([]models.Clip, error) {
    // Limits are automatically enforced
    r.helper.EnforcePaginationLimits(&limit, &offset)
    
    // Execute query...
}
```

## Search Query Limits

### Configuration

Search limits can be configured via environment variables:

```bash
# Maximum results per search (default: 100)
SEARCH_MAX_RESULT_SIZE=100

# Maximum aggregation bucket size (default: 100)
SEARCH_MAX_AGGREGATION_SIZE=100

# Maximum aggregation nesting depth (default: 2)
SEARCH_MAX_AGGREGATION_NEST=2

# Maximum query clauses in bool queries (default: 20)
SEARCH_MAX_QUERY_CLAUSES=20

# Maximum search timeout in seconds (default: 5)
SEARCH_MAX_TIME_SEC=5
```

### Enforced Limits

| Limit | Default Value | Purpose |
|-------|---------------|---------|
| Max Result Size | 100 | Prevents large result sets in search responses |
| Max Aggregation Size | 100 buckets | Limits aggregation bucket counts |
| Max Aggregation Depth | 2 levels | Prevents deeply nested aggregations |
| Max Query Clauses | 20 | Limits bool query complexity |
| Max Search Time | 5 seconds | Prevents long-running searches |

### Search Query Validation

The system validates OpenSearch queries for:

- **Result Size**: Number of documents requested
- **Aggregation Depth**: Nesting level of aggregations
- **Aggregation Sizes**: Bucket sizes in term aggregations
- **Query Clauses**: Number of clauses in bool queries

### Aggregation Examples

#### ❌ Avoid: Deeply Nested Aggregations

```json
{
  "aggs": {
    "level1": {
      "terms": {"field": "game_id", "size": 1000},
      "aggs": {
        "level2": {
          "terms": {"field": "broadcaster_id", "size": 1000},
          "aggs": {
            "level3": {
              "terms": {"field": "language", "size": 1000},
              "aggs": {
                "level4": {
                  "terms": {"field": "duration", "size": 1000}
                }
              }
            }
          }
        }
      }
    }
  }
}
```

This query would be rejected due to:
- Aggregation depth exceeds 2 levels (has 3 levels of nesting: level1→level2→level3→level4)
- Aggregation sizes exceed 100 buckets

#### ✅ Prefer: Simple Aggregations

```json
{
  "aggs": {
    "games": {
      "terms": {"field": "game_id", "size": 50}
    },
    "languages": {
      "terms": {"field": "language", "size": 20}
    }
  }
}
```

### Search Service Integration

The `OpenSearchService` automatically enforces limits:

```go
service := NewOpenSearchService(osClient)

// Limits are automatically enforced
response, err := service.Search(ctx, searchRequest)
```

## Performance Optimization Strategies

### 1. Use Materialized Views

For expensive aggregations, use materialized views:

```sql
CREATE MATERIALIZED VIEW clip_stats AS
SELECT 
    clip_id,
    COUNT(comments.id) as comment_count,
    SUM(votes.value) as vote_score
FROM clips
LEFT JOIN comments ON clips.id = comments.clip_id
LEFT JOIN votes ON clips.id = votes.clip_id
GROUP BY clips.id;

-- Refresh periodically
REFRESH MATERIALIZED VIEW CONCURRENTLY clip_stats;
```

### 2. Add Appropriate Indexes

Ensure queries can use indexes:

```sql
-- Index for cursor-based pagination
CREATE INDEX idx_clips_created_id ON clips(created_at, id);

-- Index for filtered queries
CREATE INDEX idx_clips_game_created ON clips(game_id, created_at) 
WHERE is_removed = false;
```

### 3. Cache Query Results

Use Redis to cache expensive query results:

```go
// Check cache first
cached, err := redis.Get(ctx, cacheKey).Result()
if err == nil {
    return cached, nil
}

// Execute query and cache result
result, err := executeQuery(ctx)
if err == nil {
    redis.Set(ctx, cacheKey, result, 5*time.Minute)
}
```

### 4. Limit Result Sets Early

Apply filters early in queries:

```sql
-- Good: Filter before join
SELECT c.*, u.username
FROM clips c
WHERE c.created_at > NOW() - INTERVAL '1 day'
  AND c.is_removed = false
JOIN users u ON c.submitted_by_user_id = u.id
LIMIT 100;
```

## Monitoring and Alerts

### Metrics to Track

- Query execution time (p50, p95, p99)
- Slow query count (>1s)
- Query complexity scores
- Search timeout rate
- Cache hit rates

### Example Prometheus Queries

```promql
# Slow queries per minute
rate(slow_queries_total[1m]) > 10

# Average query complexity
avg(query_complexity_score) > 80

# Search timeout rate
rate(search_timeouts_total[5m]) / rate(search_requests_total[5m]) > 0.05
```

## Error Handling

### Database Query Errors

```go
err := repository.List(ctx, 5000, 0)
// Returns: ErrLimitTooLarge: limit 5000 (max: 1000)

err := repository.List(ctx, 100, 2000)
// Returns: ErrOffsetTooLarge: offset 2000 (max: 1000)
```

### Search Query Errors

```go
err := validator.ValidateSearchSize(200)
// Returns: ErrSearchResultSizeTooLarge: requested 200 (max: 100)

err := validator.ValidateAggregations(deeplyNestedAggs)
// Returns: ErrAggregationDepthTooDeep: depth 3 (max: 2)
```

## API Client Guidelines

### Pagination

Always use reasonable page sizes:

```javascript
// Good
const clips = await api.getClips({ page: 1, limit: 50 });

// Bad: Will be capped at 100
const clips = await api.getClips({ page: 1, limit: 1000 });
```

### Search

Keep aggregations simple:

```javascript
// Good
const results = await api.search({
  query: "gameplay",
  filters: { game: "fortnite" },
  page: 1,
  limit: 20
});

// Bad: Too many results requested
const results = await api.search({
  query: "gameplay",
  limit: 500  // Will be capped at 100
});
```

## Security Considerations

These limits protect against:

1. **DoS Attacks**: Prevent malicious queries from exhausting resources
2. **Resource Exhaustion**: Limit memory and CPU usage per query
3. **Database Overload**: Prevent cascading failures from expensive queries
4. **Cost Control**: Reduce infrastructure costs from inefficient queries

## Testing

### Query Analyzer Tests

```bash
go test ./internal/repository -run TestQuery
```

### Search Limits Tests

```bash
go test ./internal/services -run TestSearch
```

### Load Testing

```bash
k6 run backend/tests/load/scenarios/search.js
```

## Related Documentation

- [Security testing runbook](security-testing-runbook.md)
- [Database developer guide](../backend/DEVELOPER_GUIDE.md#database-schema--migrations)
- [Search architecture](../backend/semantic-search-arch.md)
- [API rate limiting](../openapi/generated/api-reference.md#rate-limiting)

## References

- [PostgreSQL Query Performance](https://www.postgresql.org/docs/current/performance-tips.html)
- [OpenSearch Performance Tuning](https://opensearch.org/docs/latest/tuning-your-cluster/)
- [Use The Index, Luke!](https://use-the-index-luke.com/)
