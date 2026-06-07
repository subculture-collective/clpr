---
title: Search Index Versioning and Management
summary: This document describes the index versioning system for semantic search in Clipper, enabling zero-downtime index rebuilds, rollbacks, and version...
tags: ["archive", "implementation"]
area: docs
status: archived
owner: team-core
version: "1.0"
last_reviewed: 2026-01-29
---

# Search Index Versioning and Management

This document describes the index versioning system for semantic search in Clipper, enabling zero-downtime index rebuilds, rollbacks, and version management.

## Overview

The search index versioning system provides:

- **Versioned indices**: All indices use versioned names (e.g., `clips_v1`, `clips_v2`)
- **Alias-based routing**: Searches use an alias (e.g., `clips`) that points to the active version
- **Zero-downtime rebuilds**: New indices are built in the background before swapping
- **Quick rollbacks**: Previous versions are retained for fast rollback if needed
- **Automatic cleanup**: Old versions are automatically removed to save disk space

## Versioning Scheme

### Index Naming Convention

```
{base_index}_v{version}
```

Examples:
- `clips_v1` - Version 1 of the clips index
- `clips_v2` - Version 2 of the clips index
- `users_v3` - Version 3 of the users index

### Alias Structure

Each base index name (`clips`, `users`, `tags`, `games`) is an alias that points to the currently active versioned index.

```
clips (alias) → clips_v2 (active index)
              → clips_v1 (previous, kept for rollback)
              → clips_v0 (old, eligible for cleanup)
```

## Architecture

### Components

```
┌─────────────────────────────────────────────────────────────────┐
│                     Search Index Manager                          │
│                     (CLI + Services)                              │
└────────────────────────────┬────────────────────────────────────┘
                             │
            ┌────────────────┼────────────────┐
            │                │                │
            ▼                ▼                ▼
   ┌─────────────┐   ┌─────────────┐   ┌─────────────┐
   │  Index      │   │  Index      │   │  Index      │
   │  Version    │   │  Rebuild    │   │  Indexer    │
   │  Service    │   │  Service    │   │  Service    │
   └─────────────┘   └─────────────┘   └─────────────┘
            │                │                │
            └────────────────┼────────────────┘
                             │
                             ▼
                    ┌─────────────────┐
                    │   OpenSearch    │
                    │                 │
                    │  clips (alias)  │
                    │  clips_v1       │
                    │  clips_v2       │
                    │  ...            │
                    └─────────────────┘
```

### Services

1. **IndexVersionService**: Manages index versions, aliases, and version metadata
2. **IndexRebuildService**: Orchestrates index rebuilds with data from PostgreSQL
3. **SearchIndexerService**: Handles individual document indexing

## CLI Tool: search-index-manager

The `search-index-manager` CLI tool provides commands for managing search indices.

### Installation

```bash
cd backend
go build -o bin/search-index-manager ./cmd/search-index-manager
```

### Commands

#### Show Index Status

```bash
# Show all indices
./bin/search-index-manager status

# Show specific index
./bin/search-index-manager status -index clips

# JSON output
./bin/search-index-manager status -json
```

Example output:
```
=== Search Index Versions ===

Index: clips
  Total Versions: 3
  Latest Version: v3
  Active Version: v3 (clips_v3)
    Documents: 15432
  All Versions:
    - clips_v3 (v3): 15432 docs [ACTIVE]
    - clips_v2 (v2): 15200 docs
    - clips_v1 (v1): 14800 docs
```

#### Rebuild Index

```bash
# Rebuild specific index with zero-downtime swap
./bin/search-index-manager rebuild -index clips

# Rebuild all indices
./bin/search-index-manager rebuild -index all

# Rebuild with custom batch size
./bin/search-index-manager rebuild -index clips -batch 500

# Rebuild without swapping (for testing)
./bin/search-index-manager rebuild -index clips -no-swap

# Dry run
./bin/search-index-manager rebuild -index clips -dry-run
```

#### Swap Alias

```bash
# Swap to latest version
./bin/search-index-manager swap -index clips

# Swap to specific version
./bin/search-index-manager swap -index clips -version 2
```

#### Rollback

```bash
# Rollback to previous version
./bin/search-index-manager rollback -index clips

# Rollback to specific version
./bin/search-index-manager rollback -index clips -version 1
```

#### Cleanup Old Versions

```bash
# Clean up old versions, keeping 2 most recent
./bin/search-index-manager cleanup -index clips -keep 2

# Clean up all indices
./bin/search-index-manager cleanup -index all -keep 2

# Dry run
./bin/search-index-manager cleanup -index clips -keep 2 -dry-run
```

## Rebuild Workflow

### Zero-Downtime Rebuild Process

```
1. Create New Version
   ┌──────────────────────────────────────────────┐
   │  clips (alias) → clips_v1 (active)           │
   │                                              │
   │  Create: clips_v2 (new, empty)               │
   └──────────────────────────────────────────────┘
                         │
                         ▼
2. Index Data to New Version
   ┌──────────────────────────────────────────────┐
   │  clips (alias) → clips_v1 (active, serving)  │
   │                                              │
   │  clips_v2 (building, indexing data)          │
   │                                              │
   │  PostgreSQL ──batch──> clips_v2              │
   └──────────────────────────────────────────────┘
                         │
                         ▼
3. Atomic Alias Swap
   ┌──────────────────────────────────────────────┐
   │  clips (alias) → clips_v2 (now active)       │
   │                                              │
   │  clips_v1 (previous, kept for rollback)      │
   └──────────────────────────────────────────────┘
                         │
                         ▼
4. Cleanup (Optional)
   ┌──────────────────────────────────────────────┐
   │  clips (alias) → clips_v2 (active)           │
   │                                              │
   │  clips_v1 (kept - within retention limit)    │
   │  clips_v0 (deleted - exceeded retention)     │
   └──────────────────────────────────────────────┘
```

### Performance Characteristics

| Operation | Typical Duration | Impact on Availability |
|-----------|-----------------|----------------------|
| Create versioned index | < 1s | None |
| Rebuild clips (10K) | ~2-5 min | None |
| Rebuild clips (100K) | ~15-30 min | None |
| Alias swap | < 100ms | None (atomic) |
| Rollback | < 100ms | None (atomic) |
| Cleanup | < 5s | None |

## Rollback Plan

### Quick Rollback Procedure

If issues are detected after a rebuild:

```bash
# 1. Check current status
./bin/search-index-manager status -index clips

# 2. Rollback to previous version
./bin/search-index-manager rollback -index clips

# 3. Verify rollback
./bin/search-index-manager status -index clips
```

### Rollback Scenarios

#### Scenario 1: Data Quality Issues

If the new index has data quality issues:

```bash
# Rollback to previous version
./bin/search-index-manager rollback -index clips -version 2

# Investigate and fix data issues
# Then rebuild when ready
./bin/search-index-manager rebuild -index clips
```

#### Scenario 2: Performance Degradation

If search performance degrades:

```bash
# Check version info
./bin/search-index-manager status -index clips -json

# Rollback to known good version
./bin/search-index-manager rollback -index clips -version 1

# Monitor performance
curl http://localhost:8080/health/ready
```

#### Scenario 3: Mapping Incompatibility

If new mapping causes issues:

```bash
# Rollback to version with old mapping
./bin/search-index-manager rollback -index clips -version 1

# Fix mapping definition
# Then rebuild with corrected mapping
./bin/search-index-manager rebuild -index clips
```

## Configuration

### Default Settings

| Setting | Default | Description |
|---------|---------|-------------|
| Batch Size | 100 | Documents per batch during rebuild |
| Keep Versions | 2 | Number of old versions to retain |
| Swap After Build | true | Automatically swap alias after rebuild |
| Verbose | true | Show progress during rebuild |

### Environment Variables

The index manager uses standard Clipper configuration:

```bash
# OpenSearch connection
OPENSEARCH_URL=http://localhost:9200
OPENSEARCH_USERNAME=
OPENSEARCH_PASSWORD=
OPENSEARCH_INSECURE_SKIP_VERIFY=true  # DEV ONLY

# Database connection
DATABASE_HOST=localhost
DATABASE_PORT=5436
DATABASE_USER=clpr
DATABASE_PASSWORD=clpr_password
DATABASE_NAME=clpr_db
```

## Monitoring

### Health Check Endpoint

The existing health check endpoint includes OpenSearch status:

```bash
curl http://localhost:8080/health/ready
```

### Prometheus Metrics

Relevant metrics to monitor:

```
# Index size and document count
opensearch_indices_docs_count{index="clips_v*"}
opensearch_indices_store_size_bytes{index="clips_v*"}

# Search latency
search_query_duration_ms{search_type="hybrid"}

# Index operations
search_index_rebuild_duration_seconds
search_index_swap_total
search_index_rollback_total
```

### Alerting Rules

```yaml
# Alert on failed rebuild
- alert: SearchIndexRebuildFailed
  expr: search_index_rebuild_errors_total > 0
  for: 5m
  severity: warning
  annotations:
    summary: "Search index rebuild failed"
    description: "Check rebuild logs for errors"

# Alert on missing active alias
- alert: SearchIndexNoActiveAlias
  expr: opensearch_alias_active{alias=~"clips|users|tags|games"} == 0
  for: 5m
  severity: critical
  annotations:
    summary: "Search index has no active alias"
    description: "Search functionality may be impacted"
```

## Operational Procedures

### Scheduled Rebuilds

For regular maintenance, schedule rebuilds during low-traffic periods:

```bash
# Cron job example (daily at 3 AM)
0 3 * * * cd /opt/clpr/backend && ./bin/search-index-manager rebuild -index all -batch 200 -keep 2 >> /var/log/clpr/index-rebuild.log 2>&1
```

### Pre-deployment Checklist

Before a rebuild:

1. [ ] Verify OpenSearch is healthy
2. [ ] Check current index status
3. [ ] Ensure sufficient disk space for new version
4. [ ] Confirm database is accessible
5. [ ] Review any mapping changes

### Post-rebuild Verification

After a rebuild:

1. [ ] Verify new version is active
2. [ ] Test search functionality
3. [ ] Check document counts match expected
4. [ ] Monitor search latency for 15 minutes
5. [ ] Review error logs

## Troubleshooting

### Common Issues

#### Issue: Rebuild takes too long

**Cause**: Large dataset or slow database queries

**Solution**:
- Increase batch size: `-batch 500`
- Run during off-peak hours
- Check database performance

#### Issue: Alias swap fails

**Cause**: New index not found or naming issue

**Solution**:
```bash
# Check index exists
curl http://localhost:9200/_cat/indices?v

# Manually verify alias
curl http://localhost:9200/_alias/clips

# Retry swap
./bin/search-index-manager swap -index clips -version X
```

#### Issue: Rollback target version not found

**Cause**: Version was cleaned up or never created

**Solution**:
```bash
# List all available versions
./bin/search-index-manager status -index clips -json

# Rebuild if no valid rollback target
./bin/search-index-manager rebuild -index clips
```

### Recovery Procedures

#### Complete Index Loss

If all versioned indices are lost:

```bash
# 1. Rebuild from scratch
./bin/search-index-manager rebuild -index all

# 2. Verify all indices
./bin/search-index-manager status

# 3. Test search functionality
curl "http://localhost:8080/api/v1/search?q=test"
```

#### Alias Pointing to Wrong Version

```bash
# 1. Check current alias target
curl http://localhost:9200/_alias/clips

# 2. Swap to correct version
./bin/search-index-manager swap -index clips -version X
```

## Security Considerations

1. **Access Control**: Limit access to the search-index-manager CLI to authorized operators
2. **Audit Logging**: Log all index operations for audit trail
3. **Backup Before Cleanup**: Ensure backups exist before deleting old versions
4. **Validate Mappings**: Review mapping changes for security implications

## Future Enhancements

- [ ] Web UI for index management
- [ ] Automated rollback on degraded search quality
- [ ] Index diff tool for comparing versions
- [ ] Scheduled rebuild automation with monitoring integration
- [ ] Multi-cluster replication support
