---
title: "Semantic Search Architecture"
summary: "This document describes the architecture for semantic search in Clipper, implementing hybrid BM25 + "
tags: ['backend']
area: "backend"
status: "stable"
owner: "team-core"
version: "1.0"
last_reviewed: 2025-12-11
---

# Semantic Search Architecture

This document describes the architecture for semantic search in Clipper, implementing hybrid BM25 + vector similarity re-ranking for improved search relevance.

## Overview

Semantic search combines traditional keyword-based search (BM25) with vector embeddings to understand query intent and context. This enables finding clips based on meaning, not just exact keyword matches.

**Key Components**:

- **OpenSearch**: BM25 full-text search for candidate selection
- **pgvector**: PostgreSQL extension for vector similarity search
- **Embedding Service**: Generates vector embeddings from text
- **Hybrid Search Service**: Orchestrates BM25 + vector re-ranking

## Architecture Decision

See [[../decisions/adr-1-semantic-search-vector-db|ADR-001: Semantic Search Vector Database Selection]] for the full decision rationale.

**Selected Approach**: pgvector in PostgreSQL with hybrid BM25 + vector re-ranking

## System Architecture

### High-Level Component Diagram

```
┌─────────────────────────────────────────────────────────────────┐
│                         Client Application                       │
│                    (Web Browser / Mobile App)                    │
└────────────────────────────┬────────────────────────────────────┘
                             │
                             │ HTTP/REST
                             ▼
                  ┌──────────────────────┐
                  │    API Gateway       │
                  │   (Go + Gin)         │
                  └──────────┬───────────┘
                             │
              ┌──────────────┼──────────────┐
              │              │              │
              ▼              ▼              ▼
    ┌─────────────┐  ┌─────────────┐  ┌─────────────┐
    │   Search    │  │  Embedding  │  │   Clip      │
    │  Handler    │  │   Service   │  │  Handler    │
    └──────┬──────┘  └──────┬──────┘  └──────┬──────┘
           │                │                 │
           │                │                 │
           ▼                ▼                 ▼
    ┌──────────────────────────────────────────────┐
    │         Hybrid Search Service                │
    │  (Orchestrates BM25 + Vector Search)         │
    └──────┬──────────────────────────┬────────────┘
           │                          │
           │                          │
    ┌──────▼──────┐            ┌──────▼──────────┐
    │  OpenSearch │            │   PostgreSQL    │
    │   Service   │            │   + pgvector    │
    │             │            │                 │
    │  BM25 Index │            │ Vector Embeddings│
    └─────────────┘            └─────────────────┘
```

## Indexing Path

The indexing path shows how clip data flows from creation/update to searchable embeddings.

### Indexing Flow Diagram

```
┌──────────────────────────────────────────────────────────────────┐
│                    1. CLIP CREATION / UPDATE                      │
└────────────────────────────────┬─────────────────────────────────┘
                                 │
                                 │ HTTP POST/PUT
                                 ▼
                    ┌─────────────────────────┐
                    │   Clip Handler          │
                    │   (Validation)          │
                    └────────────┬────────────┘
                                 │
                                 │ 2. Save Metadata
                                 ▼
                    ┌─────────────────────────┐
                    │   Clip Repository       │
                    │   (Database Writer)     │
                    └────────────┬────────────┘
                                 │
                    ┌────────────┼────────────┐
                    │            │            │
         3. Write to DB     4. Write to    5. Generate
                    │       Search Index     Embedding
                    │            │            │
                    ▼            ▼            ▼
        ┌───────────────┐  ┌──────────┐  ┌──────────────┐
        │  PostgreSQL   │  │OpenSearch│  │  Embedding   │
        │               │  │  BM25    │  │   Service    │
        │ clips table   │  │  Index   │  │              │
        └───────────────┘  └──────────┘  └──────┬───────┘
                                                 │
                                                 │ 6. Return Vector
                                                 │    (768 dims)
                                                 ▼
                                      ┌──────────────────┐
                                      │  Update Clips    │
                                      │  with Embedding  │
                                      └────────┬─────────┘
                                               │
                                               │ 7. Store Vector
                                               ▼
                                      ┌──────────────────┐
                                      │   PostgreSQL     │
                                      │                  │
                                      │ clips.embedding  │
                                      │   vector(768)    │
                                      │                  │
                                      │   HNSW Index     │
                                      └──────────────────┘
```

### Indexing Steps Explained

1. **Clip Creation/Update**: User submits new clip or updates existing clip metadata
2. **Validation**: Clip handler validates input data
3. **Database Write**: Clip metadata saved to PostgreSQL `clips` table
4. **Search Index**: Clip indexed in OpenSearch for BM25 full-text search
5. **Embedding Generation**: Asynchronous task generates vector embedding
   - Combines: title, description, game name, broadcaster name, tags
   - Model: OpenAI text-embedding-ada-002 (or open-source alternative)
   - Output: 768-dimensional vector
6. **Vector Storage**: Embedding stored in `clips.embedding` column
7. **HNSW Indexing**: PostgreSQL automatically updates HNSW index for fast similarity search

### Batch Indexing (Backfill)

For existing clips without embeddings:

```
┌──────────────────────────────────────────────────────────────────┐
│                     BACKFILL PROCESS                              │
└────────────────────────────────┬─────────────────────────────────┘
                                 │
                                 │ 1. Start Backfill Job
                                 ▼
                    ┌─────────────────────────┐
                    │  Backfill Service       │
                    │  (Batch Processor)      │
                    └────────────┬────────────┘
                                 │
                                 │ 2. Query Clips Without Embeddings
                                 │    (BATCH_SIZE = 100)
                                 ▼
                    ┌─────────────────────────┐
                    │   PostgreSQL            │
                    │   SELECT WHERE          │
                    │   embedding IS NULL     │
                    └────────────┬────────────┘
                                 │
                                 │ 3. Batch of Clips
                                 ▼
                    ┌─────────────────────────┐
                    │  Embedding Service      │
                    │  (Batch API Call)       │
                    └────────────┬────────────┘
                                 │
                                 │ 4. Batch of Vectors
                                 ▼
                    ┌─────────────────────────┐
                    │  Bulk Update            │
                    │  PostgreSQL             │
                    └────────────┬────────────┘
                                 │
                                 │ 5. Loop Until Complete
                                 │
                                 ▼
                    ┌─────────────────────────┐
                    │  Progress Log           │
                    │  "Processed 1000/5000"  │
                    └─────────────────────────┘
```

## Query Path

The query path shows how user search queries are processed through hybrid search.

### Query Flow Diagram

```
┌──────────────────────────────────────────────────────────────────┐
│                   USER SEARCH QUERY                               │
│              "funny valorant clutch moments"                      │
└────────────────────────────────┬─────────────────────────────────┘
                                 │
                                 │ HTTP GET /api/v1/search
                                 ▼
                    ┌─────────────────────────┐
                    │  Search Handler         │
                    │  (Parse & Validate)     │
                    └────────────┬────────────┘
                                 │
                                 │ 1. Query + Filters
                                 ▼
                    ┌─────────────────────────┐
                    │ Hybrid Search Service   │
                    │ (Orchestrator)          │
                    └────────┬────────────────┘
                             │
            ┌────────────────┼────────────────┐
            │                │                │
            │ 2a. BM25    2b. Generate     2c. Cache Check
            │    Query       Embedding       (Redis)
            │                │                │
            ▼                ▼                ▼
  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐
  │ OpenSearch   │  │  Embedding   │  │    Redis     │
  │   Service    │  │   Service    │  │   Cache      │
  │              │  │              │  │              │
  │  BM25 Query  │  │ Query Vector │  │ Query Vector │
  │              │  │  (768 dims)  │  │  (cached)    │
  └──────┬───────┘  └──────┬───────┘  └──────┬───────┘
         │                 │                 │
         │ 3. Top 100      │ 4. Vector       │ (or cached)
         │    Candidates   │    [0.234,...]  │
         │                 └─────────┬────────┘
         │                           │
         │                           │ 5. Cache Vector (if new)
         │                           │
         └───────────┬───────────────┘
                     │
                     │ 6. Candidate IDs + Query Vector
                     ▼
          ┌──────────────────────────┐
          │   PostgreSQL + pgvector  │
          │                          │
          │   SELECT * FROM clips    │
          │   WHERE id = ANY($1)     │
          │   ORDER BY               │
          │   embedding <=>          │
          │   $2::vector(768)        │
          │   LIMIT 20               │
          └────────────┬─────────────┘
                       │
                       │ 7. Top 20 Results (Re-ranked)
                       ▼
          ┌──────────────────────────┐
          │  Response Formatter      │
          │  (Add Metadata)          │
          └────────────┬─────────────┘
                       │
                       │ 8. JSON Response
                       ▼
          ┌──────────────────────────┐
          │      Client              │
          │  Display Results         │
          └──────────────────────────┘
```

### Query Steps Explained

1. **Parse Query**: Extract search terms, filters, pagination params
2. **Parallel Operations**:
   - **2a. BM25 Search**: OpenSearch returns top 100 candidates by keyword relevance
   - **2b. Embedding Generation**: Convert query text to 768-dim vector
   - **2c. Cache Check**: Check Redis for previously generated query embedding
3. **Candidate Selection**: BM25 narrows search space from 100K clips to 100 candidates
4. **Query Vector**: Embedding service returns vector or retrieves from cache
5. **Cache Update**: Store query vector in Redis for 1 hour (common queries benefit)
6. **Vector Re-ranking**: pgvector compares query vector with 100 candidate embeddings
   - Uses cosine similarity: `embedding <=> query_vector`
   - HNSW index accelerates similarity computation
   - Returns top 20 most similar clips
7. **Result Set**: Final ranked results combining BM25 + semantic relevance
8. **Response**: JSON with clips, metadata, and relevance scores

### Performance Optimization

#### Caching Strategy

```
┌─────────────────────────────────────────────────────────────┐
│                    CACHING LAYERS                            │
└─────────────────────────────────────────────────────────────┘

Layer 1: Query Embedding Cache (Redis)
┌────────────────────────────────────┐
│  Key: "embed:query:<hash>"         │
│  Value: [0.234, 0.456, ...]        │
│  TTL: 1 hour                       │
│  Hit Rate: ~60% (common queries)   │
└────────────────────────────────────┘

Layer 2: BM25 Results Cache (Redis)
┌────────────────────────────────────┐
│  Key: "bm25:<query_hash>:<filters>"│
│  Value: [clip_ids]                 │
│  TTL: 5 minutes                    │
│  Hit Rate: ~40% (trending queries) │
└────────────────────────────────────┘

Layer 3: Clip Metadata Cache (Redis)
┌────────────────────────────────────┐
│  Key: "clip:<clip_id>"             │
│  Value: {clip metadata JSON}       │
│  TTL: 15 minutes                   │
│  Hit Rate: ~80% (popular clips)    │
└────────────────────────────────────┘
```

## Data Models

### Database Schema

```sql
-- Enable pgvector extension
CREATE EXTENSION IF NOT EXISTS vector;

-- Add embedding column to clips table
ALTER TABLE clips 
ADD COLUMN embedding vector(768),
ADD COLUMN embedding_model VARCHAR(50) DEFAULT 'text-embedding-ada-002',
ADD COLUMN embedding_generated_at TIMESTAMP;

-- Create HNSW index for fast similarity search
-- HNSW parameters: m=16 (connections), ef_construction=64 (build quality)
CREATE INDEX clips_embedding_idx ON clips 
USING hnsw (embedding vector_cosine_ops)
WITH (m = 16, ef_construction = 64);

-- Index for finding clips without embeddings (backfill)
CREATE INDEX clips_embedding_null_idx ON clips (id) 
WHERE embedding IS NULL;
```

### Embedding Generation

```go
type EmbeddingService struct {
    client *openai.Client
    cache  *redis.Client
}

type ClipEmbeddingInput struct {
    Title         string
    Description   string
    GameName      string
    BroadcasterName string
    CreatorName   string
    Tags          []string
}

// Generate embedding from clip metadata
func (s *EmbeddingService) GenerateClipEmbedding(ctx context.Context, input *ClipEmbeddingInput) ([]float32, error) {
    // Combine fields into single text
    text := fmt.Sprintf("%s. %s. Game: %s. Broadcaster: %s. Tags: %s",
        input.Title,
        input.Description,
        input.GameName,
        input.BroadcasterName,
        strings.Join(input.Tags, ", "),
    )
    
    // Check cache
    cacheKey := fmt.Sprintf("embed:clip:%s", hashText(text))
    if cached, err := s.cache.Get(ctx, cacheKey).Result(); err == nil {
        return parseVector(cached), nil
    }
    
    // Generate embedding via API
    resp, err := s.client.CreateEmbeddings(ctx, openai.EmbeddingRequest{
        Model: "text-embedding-ada-002",
        Input: []string{text},
    })
    if err != nil {
        return nil, err
    }
    
    embedding := resp.Data[0].Embedding
    
    // Cache for 24 hours
    s.cache.Set(ctx, cacheKey, embedding, 24*time.Hour)
    
    return embedding, nil
}
```

### Hybrid Search Query

```go
type HybridSearchService struct {
    opensearch *OpenSearchService
    postgres   *sql.DB
    embedding  *EmbeddingService
    cache      *redis.Client
}

func (s *HybridSearchService) Search(ctx context.Context, query string, opts SearchOptions) (*SearchResults, error) {
    // Step 1: BM25 candidate selection (OpenSearch)
    candidates, err := s.opensearch.Search(ctx, query, SearchOptions{
        Limit: 100,
        Filters: opts.Filters,
    })
    if err != nil {
        return nil, err
    }
    
    if len(candidates) == 0 {
        return &SearchResults{}, nil
    }
    
    // Step 2: Generate query embedding (with caching)
    queryVector, err := s.embedding.GenerateQueryEmbedding(ctx, query)
    if err != nil {
        return nil, err
    }
    
    // Step 3: Vector similarity re-ranking (pgvector)
    candidateIDs := extractIDs(candidates)
    
    query := `
        SELECT id, title, description, game_name, broadcaster_name,
               embedding <=> $1 AS similarity_score
        FROM clips
        WHERE id = ANY($2)
        ORDER BY embedding <=> $1
        LIMIT $3
    `
    
    rows, err := s.postgres.QueryContext(ctx, query, 
        pgvector.NewVector(queryVector),
        pq.Array(candidateIDs),
        opts.Limit,
    )
    if err != nil {
        return nil, err
    }
    defer rows.Close()
    
    // Step 4: Build results
    results := &SearchResults{}
    for rows.Next() {
        var clip Clip
        if err := rows.Scan(&clip.ID, &clip.Title, /* ... other fields ... */); err != nil {
            return nil, err
        }
        results.Clips = append(results.Clips, clip)
    }
    
    return results, nil
}
```

## Performance Targets

| Metric | Target | Notes |
|--------|--------|-------|
| BM25 Candidate Selection | <20ms | OpenSearch query |
| Query Embedding Generation | <50ms | <5ms if cached |
| Vector Re-ranking | <30ms | 100 similarity comparisons |
| **Total Query Latency** | **<100ms** | P95 latency |
| Indexing Latency | <1s | Async, non-blocking |
| Backfill Throughput | 100+ clips/sec | Batch processing |

## Monitoring & Observability

### Key Metrics

```
Search Metrics:
- search_query_duration_ms (histogram)
  - Labels: search_type (bm25, vector, hybrid)
- search_results_count (histogram)
- search_cache_hit_rate (gauge)
- embedding_generation_duration_ms (histogram)

Indexing Metrics:
- embedding_generation_total (counter)
- embedding_generation_errors (counter)
- clips_with_embeddings (gauge)
- clips_without_embeddings (gauge)

Database Metrics:
- pgvector_query_duration_ms (histogram)
- pgvector_index_size_bytes (gauge)
- postgres_connections_used (gauge)
```

### Alerting Rules

```yaml
- alert: SemanticSearchHighLatency
  expr: histogram_quantile(0.95, search_query_duration_ms) > 100
  for: 5m
  severity: warning

- alert: EmbeddingGenerationFailing
  expr: rate(embedding_generation_errors[5m]) > 0.1
  for: 5m
  severity: critical

- alert: MissingEmbeddings
  expr: clips_without_embeddings / (clips_with_embeddings + clips_without_embeddings) > 0.1
  for: 30m
  severity: warning
```

## Infrastructure Requirements

### Resource Allocation

```
PostgreSQL Configuration:
- shared_buffers = 2GB
- effective_cache_size = 6GB
- maintenance_work_mem = 1GB
- max_connections = 100
- work_mem = 64MB

pgvector Settings:
- hnsw.ef_search = 40 (query-time search depth)
  - Higher = better recall, slower queries
  - Lower = faster queries, lower recall
  - 40 is good balance for <100ms queries

OpenSearch Configuration:
- Heap size: 512MB (unchanged)
- Index refresh interval: 5s

Redis Configuration:
- maxmemory = 256MB (unchanged)
- maxmemory-policy = allkeys-lru
```

### Scaling Considerations

| Metric | Current | 1 Year | 3 Years | Action Required |
|--------|---------|--------|---------|-----------------|
| Total Clips | 10K | 100K | 500K | None (within pgvector limits) |
| Embeddings Size | 30MB | 300MB | 1.5GB | Monitor RAM usage |
| HNSW Index Size | 60MB | 600MB | 3GB | Consider read replica |
| Query Load | 10/sec | 50/sec | 200/sec | Add read replica |

**Migration Point**: If exceeding 1M clips or 500 QPS, consider:

1. Read replica for search queries
2. Separate vector database (Qdrant, Milvus)
3. Sharding by time period (recent vs archive)

## Testing Strategy

### Unit Tests

```go
func TestEmbeddingGeneration(t *testing.T) {
    // Test embedding dimensions
    // Test caching behavior
    // Test error handling
}

func TestHybridSearch(t *testing.T) {
    // Test BM25 candidate selection
    // Test vector re-ranking
    // Test result ordering
}
```

### Integration Tests

```go
func TestEndToEndSearch(t *testing.T) {
    // Create test clips with embeddings
    // Execute hybrid search query
    // Verify result relevance
    // Verify performance < 100ms
}
```

### Relevance Testing

Use the search relevance dataset (`backend/testdata/search_relevance_dataset.yaml`) to validate:

- Semantic similarity matches
- Query intent understanding
- Ranking improvements over BM25-only

## Rollout Plan

### Phase 1: Foundation (2 weeks)

- ✅ Install pgvector extension
- ✅ Add embedding column and index
- ✅ Implement embedding service

### Phase 2: Backfill (1 week)

- ✅ Backfill existing clips
- ✅ Monitor index build performance
- ✅ Validate embedding quality

### Phase 3: Hybrid Search (2 weeks)

- ✅ Implement hybrid search service
- ✅ Create API endpoint
- ✅ Add caching layer

### Phase 4: Production (1 week)

- ✅ A/B test semantic vs traditional
- ✅ Monitor performance metrics
- ✅ Adjust HNSW parameters
- ✅ Documentation updates

### Phase 5: Optimization (ongoing)

- ✅ Fine-tune relevance scoring
- ✅ Optimize cache hit rates
- ✅ Monitor resource usage
- ✅ Gather user feedback

## Security Considerations

1. **API Key Management**: Store OpenAI API key in secrets management
2. **Rate Limiting**: Limit embedding generation to prevent abuse
3. **Input Validation**: Sanitize text before embedding generation
4. **Query Injection**: Use parameterized queries for vector search
5. **Cache Poisoning**: Hash query inputs for cache keys

## Index Management and Versioning

For zero-downtime index rebuilds, rollbacks, and version management, see:

- **Index Versioning Documentation** - Complete guide to index versioning scheme, rebuild workflows, and rollback procedures

### Key Features

- **Versioned Indices**: All indices use versioned names (e.g., `clips_v1`, `clips_v2`)
- **Alias-based Routing**: Searches use aliases that point to active versions
- **Zero-downtime Rebuilds**: New indices are built before swapping
- **Quick Rollbacks**: Previous versions retained for fast recovery

### CLI Tool

```bash
# Check index status
./bin/search-index-manager status

# Rebuild with zero-downtime swap
./bin/search-index-manager rebuild -index clips

# Rollback to previous version
./bin/search-index-manager rollback -index clips
```

## Future Enhancements

1. **Fine-tuned Models**: Train custom embedding model on clip data
2. **Multi-modal Embeddings**: Include video thumbnails in embeddings
3. **Personalized Re-ranking**: User preference-based similarity
4. **Cross-lingual Search**: Support searching across languages
5. **Query Expansion**: Automatically expand queries with related terms
6. **Federated Search**: Combine clips, users, and games in single query

## References

- [[../decisions/adr-1-semantic-search-vector-db|ADR-001: Vector Database Selection]]
- [OpenSearch Documentation](https://opensearch.org/docs/latest/)
- [pgvector Documentation](https://github.com/pgvector/pgvector)
- [OpenAI Embeddings Guide](https://platform.openai.com/docs/guides/embeddings)
- [HNSW Algorithm Paper](https://arxiv.org/abs/1603.09320)
- [Hybrid Search Best Practices](https://www.pinecone.io/learn/hybrid-search/)
