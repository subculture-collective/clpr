# Embedding Backfill Tool

This tool generates embeddings for existing clips in the database using the OpenAI Embeddings API.

## Overview

The embedding backfill tool:

- Generates vector embeddings for clips without embeddings
- Uses batch processing to efficiently handle large datasets
- Caches embeddings in Redis to reduce API costs
- Implements rate limiting to respect OpenAI API quotas
- Provides detailed progress tracking and statistics
- Supports dry-run mode for testing
- Can force-update existing embeddings

## Prerequisites

1. **OpenAI API Key**: Get your API key from [OpenAI Platform](https://platform.openai.com/api-keys)
2. **Database**: PostgreSQL with pgvector extension enabled
3. **Redis**: Running Redis instance for caching
4. **Environment Variables**: Configured in `.env` file

## Configuration

Set the following environment variables in your `.env` file:

```bash
# Required
OPENAI_API_KEY=sk-...                           # Your OpenAI API key

# Optional (with defaults)
EMBEDDING_ENABLED=true                          # Enable embedding generation
EMBEDDING_MODEL=text-embedding-3-small          # Model to use
EMBEDDING_REQUESTS_PER_MINUTE=500               # Rate limit (tier 1: 500, tier 2: 5000)
```

## Usage

### Basic Usage

Process all clips without embeddings:

```bash
cd backend
go run cmd/backfill-embeddings/main.go
```

### Options

#### Batch Size

Control how many clips are processed in each batch:

```bash
go run cmd/backfill-embeddings/main.go -batch 100
```

Default: 50 clips per batch

#### Force Update

Regenerate embeddings for all clips, even those that already have embeddings:

```bash
go run cmd/backfill-embeddings/main.go -force
```

Useful for:

- Migrating to a new embedding model
- Fixing corrupted embeddings
- Updating embeddings with improved text preprocessing

#### Dry Run

Test the backfill process without actually saving embeddings:

```bash
go run cmd/backfill-embeddings/main.go -dry-run
```

Useful for:

- Testing configuration
- Estimating processing time
- Checking for errors before committing changes

### Combined Options

```bash
go run cmd/backfill-embeddings/main.go -batch 100 -force -dry-run
```

## Building

Build a standalone binary:

```bash
cd backend
go build -o bin/backfill-embeddings ./cmd/backfill-embeddings
./bin/backfill-embeddings
```

## Output

The tool provides detailed logging and statistics:

```
Starting embedding backfill job...
Configuration: batch_size=50, force_update=false, dry_run=false
Database connection established
Redis connection established
Embedding service initialized (model: text-embedding-3-small)
Found 1234 clips to process
Processing batch of 50 clips (offset: 0, progress: 0.0%)
Progress: 10/1234 (0.8%) - processed: 10, skipped: 0, failed: 0
...
Progress: 1234/1234 (100.0%) - processed: 1234, skipped: 0, failed: 0

=== Backfill Summary ===
Total clips: 1234
Processed: 1234
Skipped: 0
Failed: 0
Duration: 5m23s
Average time per clip: 262ms

✓ Backfill completed successfully!
```

## Performance

### Estimated Processing Time

With default settings (text-embedding-3-small, batch size 50):

| Clips | Time (approx) | API Calls |
|-------|---------------|-----------|
| 1,000 | ~4 minutes | 1,000 |
| 10,000 | ~40 minutes | 10,000 |
| 100,000 | ~7 hours | 100,000 |

Note: Times assume no cache hits. Cached embeddings are retrieved instantly.

### Rate Limiting

The tool respects OpenAI API rate limits:

- **Tier 1**: 500 requests/minute (default)
- **Tier 2**: 5,000 requests/minute (configure with `EMBEDDING_REQUESTS_PER_MINUTE`)

### Caching

Embeddings are cached in Redis for 30 days:

- Reduces API costs on re-runs
- Speeds up processing significantly
- Cache key based on model + text content hash

## Error Handling

The tool includes robust error handling:

1. **Retry Logic**: Failed API calls are retried up to 3 times with exponential backoff
2. **Graceful Degradation**: Individual clip failures don't stop the entire process
3. **Detailed Logging**: All errors are logged with clip IDs for investigation
4. **Statistics**: Final summary includes count of failed clips

### Common Errors

#### Missing API Key

```
OPENAI_API_KEY is not set. Please set it in your environment or .env file
```

**Solution**: Set `OPENAI_API_KEY` in your `.env` file

#### Rate Limit Exceeded

```
API returned status 429: Rate limit exceeded
```

**Solution**:

- Reduce batch size
- Increase `EMBEDDING_REQUESTS_PER_MINUTE` if you have a higher tier
- Wait and retry later

#### Database Connection Failed

```
Failed to connect to database: ...
```

**Solution**: Check your database configuration in `.env`

## Monitoring

### Database Verification

Check embedding coverage:

```sql
-- Count clips with embeddings
SELECT COUNT(*)
FROM clips
WHERE embedding IS NOT NULL;

-- Get embedding statistics
SELECT
    COUNT(*) as total_clips,
    COUNT(embedding) as clips_with_embeddings,
    COUNT(embedding)::float / COUNT(*)::float * 100 as coverage_percentage
FROM clips
WHERE is_removed = false;
```

### Redis Cache Stats

Check cache hit rate:

```bash
redis-cli INFO stats | grep hits
```

## Best Practices

1. **Run During Low Traffic**: Schedule backfills during off-peak hours
2. **Start Small**: Test with small batch sizes first
3. **Use Dry Run**: Always test with `-dry-run` before committing
4. **Monitor Progress**: Watch logs for errors and adjust batch size if needed
5. **Cache Warmup**: Keep Redis running to benefit from caching
6. **Incremental Processing**: Run regularly for new clips (automated via scheduler)

## Automated Processing

The API server includes an automatic embedding scheduler that:

- Runs every 6 hours
- Processes clips created in the last 7 days without embeddings
- Limits to 100 clips per run to avoid overwhelming the API
- No manual intervention required

For complete backfills of existing data, use this CLI tool.

## Troubleshooting

### Slow Processing

1. Check Redis connection (caching speeds up re-runs significantly)
2. Verify rate limit settings match your OpenAI tier
3. Consider increasing batch size for better throughput

### High API Costs

1. Ensure caching is working (check Redis)
2. Avoid using `-force` unless necessary
3. Use `-dry-run` to estimate costs before running

### Memory Issues

1. Reduce batch size with `-batch` flag
2. Process in smaller increments (can resume where it left off)

## Support

For issues or questions:

- Check the Semantic Search Architecture
- Review [OpenAI Embeddings Documentation](https://platform.openai.com/docs/guides/embeddings)
- Open an issue in the GitHub repository
