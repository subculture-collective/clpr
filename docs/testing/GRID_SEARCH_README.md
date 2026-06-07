---
title: "GRID SEARCH README"
summary: "This directory contains tools for optimizing hybrid search (BM25 + Vector) weights through grid search and evaluation."
tags: ["docs"]
area: "docs"
status: "stable"
owner: "team-core"
version: "1.0"
last_reviewed: 2026-01-29
---

# Hybrid Search Weight Optimization

This directory contains tools for optimizing hybrid search (BM25 + Vector) weights through grid search and evaluation.

## Overview

As part of Roadmap 5.0 Phase 3.1, this implementation provides:

1. **Baseline Capture**: Tool to capture current search performance metrics
2. **Grid Search**: Comprehensive parameter search for optimal BM25/Vector weights
3. **Configuration Management**: Environment-based configuration for easy rollout/rollback
4. **Rollout Plan**: Documented deployment strategy with success metrics

## Quick Start

### Step 1: Capture Baseline

```bash
cd backend
go run cmd/capture-baseline-search/main.go -output baseline-search-metrics.json
```

This captures current performance metrics using the evaluation dataset.

### Step 2: Run Grid Search

```bash
# Full grid search (recommended)
go run cmd/grid-search-hybrid-search/main.go \
  -baseline baseline-search-metrics.json \
  -output hybrid-search-grid-results.json

# Quick mode (faster, fewer combinations)
go run cmd/grid-search-hybrid-search/main.go \
  -quick \
  -baseline baseline-search-metrics.json \
  -output hybrid-search-grid-results-quick.json \
  -verbose
```

### Step 3: Review Results

The grid search outputs:
- Best BM25/Vector weight combination
- Percentage improvement vs baseline
- Full metrics comparison
- Success/failure against 10% nDCG@10 improvement target

```bash
# View results
cat hybrid-search-grid-results.json | jq .best_config
```

### Step 4: Deploy Optimal Configuration

Update environment variables:

```bash
# .env or k8s ConfigMap
HYBRID_SEARCH_BM25_WEIGHT=<optimal_value>
HYBRID_SEARCH_VECTOR_WEIGHT=<optimal_value>
```

Restart service to apply changes.

## Tools

### 1. capture-baseline-search

Captures current search performance as a baseline for comparison.

**Usage:**
```bash
go run cmd/capture-baseline-search/main.go [options]

Options:
  -dataset string
        Path to evaluation dataset YAML file (default "testdata/search_evaluation_dataset.yaml")
  -output string
        Path to output JSON file (default "baseline-search-metrics.json")
  -notes string
        Notes about this baseline capture (default "Baseline capture for hybrid search weight optimization")
  -help
        Show help message
```

**Output:** JSON file with baseline metrics and configuration

### 2. grid-search-hybrid-search

Tests different BM25/Vector weight combinations to find optimal parameters.

**Usage:**
```bash
go run cmd/grid-search-hybrid-search/main.go [options]

Options:
  -dataset string
        Path to evaluation dataset YAML file (default "testdata/search_evaluation_dataset.yaml")
  -output string
        Path to output JSON file (default "hybrid-search-grid-results.json")
  -baseline string
        Path to baseline results JSON file (optional)
  -quick
        Quick mode - test fewer combinations
  -verbose
        Print detailed results for each configuration
  -help
        Show help message
```

**Output:** JSON file with grid search results, best configuration, and improvement metrics

### 3. evaluate-search (existing)

Evaluates current search configuration against the evaluation dataset.

**Usage:**
```bash
go run cmd/evaluate-search/main.go [options]
```

### 4. search-ab-test (existing)

Compares two search configurations side-by-side.

**Usage:**
```bash
go run cmd/search-ab-test/main.go \
  -config-a baseline \
  -config-b semantic-heavy
```

## Configuration

### Environment Variables

```bash
# Hybrid Search Weights (must sum to 1.0)
HYBRID_SEARCH_BM25_WEIGHT=0.7        # Default: 0.7
HYBRID_SEARCH_VECTOR_WEIGHT=0.3      # Default: 0.3

# Field Boost Parameters (for BM25)
HYBRID_SEARCH_TITLE_BOOST=3.0        # Default: 3.0
HYBRID_SEARCH_CREATOR_BOOST=2.0      # Default: 2.0
HYBRID_SEARCH_GAME_BOOST=1.0         # Default: 1.0

# Scoring Boost Parameters
HYBRID_SEARCH_ENGAGEMENT_BOOST=0.1   # Default: 0.1
HYBRID_SEARCH_RECENCY_BOOST=0.5      # Default: 0.5
```

### Quick Rollback

Rollback to baseline by resetting environment variables:

```bash
HYBRID_SEARCH_BM25_WEIGHT=0.7
HYBRID_SEARCH_VECTOR_WEIGHT=0.3
# ... restart service
```

No code changes required!

## Evaluation Dataset

Location: `backend/testdata/search_evaluation_dataset.yaml`

Contains:
- 510 labeled evaluation queries
- Relevance judgments (0-4 scale)
- Diverse query types: game-specific, creator-focused, multilingual, typos, etc.
- Metric targets for quality gates

## Metrics

### Primary Metric
- **nDCG@10**: Normalized Discounted Cumulative Gain at position 10
  - Target: ≥10% improvement vs baseline
  - Current baseline: Captured via baseline tool

### Secondary Metrics
- **nDCG@5**: nDCG at position 5
- **MRR**: Mean Reciprocal Rank (first relevant result)
- **Precision@10**: Precision at position 10
- **Recall@10**: Recall at position 10

### User Engagement Metrics (Production)
- Click-through rate (CTR)
- Dwell time on results
- Zero-result search rate
- Engagement rate

## Grid Search Parameters

### Default Grid (Full Mode)
- **BM25 Weights**: 0.3, 0.4, 0.5, 0.6, 0.7, 0.8, 0.9
- **Vector Weights**: 0.1, 0.2, 0.3, 0.4, 0.5, 0.6, 0.7
- **Total Combinations**: ~28 valid configurations
- **Runtime**: ~2-5 minutes (with simulated results)

### Quick Grid (Quick Mode)
- **BM25 Weights**: 0.5, 0.6, 0.7, 0.8
- **Vector Weights**: 0.2, 0.3, 0.4, 0.5
- **Total Combinations**: ~8 valid configurations
- **Runtime**: ~1 minute

## Rollout Plan

See [HYBRID_SEARCH_ROLLOUT.md](../docs/HYBRID_SEARCH_ROLLOUT.md) for detailed rollout strategy including:

- Phased deployment (Staging → Canary → Full)
- Success criteria and KPIs
- Monitoring and alerting
- Rollback procedures
- Timeline and responsibilities

## Examples

### Example 1: Baseline Capture

```bash
go run cmd/capture-baseline-search/main.go \
  -notes "Pre-optimization baseline for Q1 2025" \
  -output baseline-q1-2025.json
```

Output:
```
Baseline Metrics Captured
========================================

Configuration:
-----------------------------------------
  BM25 Weight:      0.70
  Vector Weight:    0.30
  ...

Baseline Metrics:
-----------------------------------------
  nDCG@10:      0.8000 (primary metric)
  nDCG@5:       0.7500
  MRR:          0.7000
  ...

Target for Optimization:
-----------------------------------------
  nDCG@10 improvement: ≥10%
  Target nDCG@10:      ≥0.8800
```

### Example 2: Grid Search with Baseline

```bash
go run cmd/grid-search-hybrid-search/main.go \
  -baseline baseline-q1-2025.json \
  -output results-q1-2025.json \
  -verbose
```

Output:
```
Grid Search Summary
========================================

Tested 28 parameter combinations

Baseline Configuration:
-----------------------------------------
  nDCG@10:      0.8000
  ...

Best Configuration:
-----------------------------------------
  BM25 Weight:   0.60
  Vector Weight: 0.40

Metrics:
-----------------------------------------
  nDCG@10:      0.8900
  ...

Improvement vs Baseline:
-----------------------------------------
  nDCG@10:      +11.25%
  ...

✅ SUCCESS: nDCG@10 improvement ≥10% target met!
```

### Example 3: A/B Testing Configurations

```bash
go run cmd/search-ab-test/main.go \
  -config-a baseline \
  -config-b semantic-heavy \
  -output ab-test-results.json
```

## Dependencies

### Related Issues
- [#837](https://git.subcult.tv/subculture-collective/clpr/issues/837) - Search Relevance Evaluation Framework (completed)
- [#805](https://git.subcult.tv/subculture-collective/clpr/issues/805) - Related search improvements

### Technical Dependencies
- Go 1.21+
- PostgreSQL with pgvector extension
- OpenSearch for BM25 search
- Evaluation dataset with labeled queries

## Troubleshooting

### Issue: Grid search doesn't meet 10% improvement target

**Solutions:**
1. Try expanding the parameter grid
2. Consider tuning field boost parameters (title, creator, game)
3. Adjust candidate pool size in hybrid search service
4. Review evaluation dataset for bias or coverage gaps

### Issue: Production metrics differ from offline evaluation

**Causes:**
- Distribution shift between evaluation queries and real user queries
- Different user behavior patterns
- Cache effects in production

**Solutions:**
- Use canary deployment to validate with real traffic
- Monitor user engagement metrics closely
- Consider online A/B testing with live users

### Issue: Configuration changes not applied

**Checklist:**
- ✅ Environment variables set correctly
- ✅ Service restarted after config change
- ✅ Feature flag `FEATURE_SEMANTIC_SEARCH` enabled
- ✅ Embedding service operational

## Testing

Run evaluation suite:

```bash
# Unit tests
cd backend
go test ./internal/services/...

# Integration tests
go test ./tests/integration/search/...

# Evaluation
go run cmd/evaluate-search/main.go -verbose
```

## Contributing

When adding new features:

1. Update evaluation dataset with new query types
2. Run grid search to validate impact
3. Update baseline metrics
4. Document in rollout plan
5. Add tests

## References

- [Search Architecture](../docs/search-architecture.md)
- [Evaluation Framework](../docs/search-evaluation.md)
- [Configuration Management](../docs/configuration.md)
- [Deployment Guide](../docs/deployment.md)

## Support

For questions or issues:
- GitHub Issues: https://git.subcult.tv/subculture-collective/clpr/issues
- Documentation: https://git.subcult.tv/subculture-collective/clpr/tree/main/docs
