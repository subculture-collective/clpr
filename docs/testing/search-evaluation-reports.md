---
title: "SEARCH EVALUATION"
summary: "This framework provides tools for evaluating and improving search quality in Clipper using standard Information Retrieval (IR) metrics."
tags: ["docs"]
area: "docs"
status: "stable"
owner: "team-core"
version: "1.0"
last_reviewed: 2026-01-29
---

# Search Relevance Evaluation Framework

This framework provides tools for evaluating and improving search quality in Clipper using standard Information Retrieval (IR) metrics.

## Overview

The search evaluation framework consists of:

1. **Labeled Evaluation Dataset** - 510+ queries with graded relevance judgments
2. **Evaluation Metrics** - nDCG, MRR, Precision, and Recall at various cutoffs
3. **A/B Testing Harness** - Compare different ranking configurations
4. **CI Automation** - Automated nightly evaluation and PR integration

## Quick Start

### Running Evaluation

```bash
# Basic evaluation with default dataset
cd backend
make evaluate-search

# Verbose output showing per-query results
go run ./cmd/evaluate-search -verbose

# Save results to JSON file
go run ./cmd/evaluate-search -output results.json
```

### Running A/B Tests

```bash
# List available configurations
go run ./cmd/search-ab-test -list-configs

# Compare two configurations
go run ./cmd/search-ab-test -config-a baseline -config-b semantic-heavy

# Save comparison results
go run ./cmd/search-ab-test \
  -config-a baseline \
  -config-b engagement-focused \
  -output comparison.json
```

## Evaluation Dataset

The evaluation dataset is stored in `backend/testdata/search_evaluation_dataset.yaml` and contains:

- **510 labeled queries** covering diverse search intents:
  - Game-specific searches (Valorant, League of Legends, CSGO, etc.)
  - Creator-focused searches (streamers, pro players)
  - Educational content (tutorials, guides, tips)
  - Funny/Entertainment content
  - Competitive/Esports content
  - Multi-word complex queries
  - Multilingual queries (Spanish, French, German, Portuguese)
  - Typo tolerance tests
  - Single-word queries

- **Relevance scale (0-4)**:
  - `4` = Perfect - Exactly what the user is looking for
  - `3` = Highly relevant - Strongly satisfies the query intent
  - `2` = Fairly relevant - Partially satisfies the query
  - `1` = Marginally relevant - Tangentially related
  - `0` = Not relevant - Completely unrelated

### Updating the Dataset

The dataset can be regenerated or modified using the Python script:

```bash
cd backend/scripts
python3 generate_search_dataset.py
```

## Metrics

### Ranking Quality Metrics

- **nDCG@5, nDCG@10** - Normalized Discounted Cumulative Gain
  - Measures ranking quality considering both relevance and position
  - Higher is better (0.0 - 1.0)
  - **Targets**: nDCG@5 ≥ 0.75, nDCG@10 ≥ 0.80

- **MRR** - Mean Reciprocal Rank
  - Average of 1/rank for the first relevant document
  - Measures how high the first relevant result appears
  - **Target**: ≥ 0.70

### Precision Metrics

- **Precision@5, @10, @20** - Fraction of top-k results that are relevant
  - Measures result quality at different cutoffs
  - **Targets**: P@5 ≥ 0.60, P@10 ≥ 0.55, P@20 ≥ 0.50

### Recall Metrics

- **Recall@5, @10, @20** - Fraction of relevant documents in top-k
  - Measures coverage of relevant results
  - **Targets**: R@5 ≥ 0.50, R@10 ≥ 0.70, R@20 ≥ 0.85

### Threshold Levels

- ✅ **Pass**: Metric meets or exceeds target
- ⚠️ **Warning**: Metric below target but above critical threshold
- ❌ **Critical**: Metric below critical threshold (CI fails)

## A/B Testing

### Available Configurations

The framework provides 8 pre-defined search configurations:

1. **baseline** - Current production configuration
   - Balanced BM25/semantic weights
   - Standard field boosts

2. **semantic-heavy** - Emphasize semantic understanding
   - Higher weight on vector search (0.6 vs 0.3)
   - Better for conceptual/semantic queries

3. **text-heavy** - Emphasize exact text matching
   - Higher weight on BM25 (0.8 vs 0.2)
   - Better for precise term matching

4. **engagement-focused** - Prioritize popular content
   - Higher engagement boost (0.3 vs 0.1)
   - Surfaces viral/popular clips

5. **recency-focused** - Prioritize recent content
   - Higher recency boost (0.8 vs 0.5)
   - Surfaces fresh content

6. **balanced** - Equal BM25/semantic weights
   - 50/50 split between approaches
   - Moderate engagement/recency boosts

7. **title-priority** - Heavily prioritize title matches
   - Higher title boost (5.0 vs 3.0)
   - Better for title-specific queries

8. **creator-priority** - Prioritize creator name matches
   - Higher creator boost (4.0 vs 2.0)
   - Better for creator-focused queries

### Interpreting A/B Results

The A/B test tool provides:

1. **Metrics Comparison** - Side-by-side metric values with percentage changes
2. **Recommendation** - Suggested action based on improvements:
   - **Significant improvement** (>5% avg, 3+ metrics up): Switch to Config B
   - **Moderate improvement** (>2% avg, 2+ metrics up): A/B test with live traffic
   - **Similar performance** (±2%): Stay with Config A
   - **Degradation** (<-2% avg): Stay with Config A

3. **Statistical Summary** - Count of significant changes (>5%)

### Best Practices

1. **Always test before deploying** - Use A/B harness to compare changes
2. **Monitor multiple metrics** - Don't optimize for a single metric
3. **Consider query diversity** - Ensure improvements across query types
4. **Validate with live traffic** - Simulated results may differ from production
5. **Document changes** - Track configuration changes and their impact

## CI Integration

The GitHub Actions workflow (`.github/workflows/search-evaluation.yml`) runs:

- **Schedule**: Nightly at 3 AM UTC
- **On push**: When search-related code changes
- **Manual**: Via workflow_dispatch

### Workflow Outputs

1. **Console Summary** - Aggregate metrics displayed in workflow logs
2. **Artifacts** - Full JSON results uploaded for 90 days
3. **PR Comments** - Automatic commenting on pull requests with metric comparison
4. **Status Checks** - Fails if any metric is in critical status

## Baseline Metrics

Current baseline (with simulated results):

| Metric | Value | Status | Target |
|--------|-------|--------|--------|
| nDCG@5 | 0.8299 | ✅ Pass | 0.75 |
| nDCG@10 | 0.8299 | ✅ Pass | 0.80 |
| MRR | 0.8294 | ✅ Pass | 0.70 |
| Precision@5 | 0.4502 | ❌ Fail | 0.60 |
| Precision@10 | 0.2251 | ❌ Fail | 0.55 |
| Precision@20 | 0.1125 | ❌ Fail | 0.50 |
| Recall@5 | 0.8288 | ✅ Pass | 0.50 |
| Recall@10 | 0.8288 | ✅ Pass | 0.70 |
| Recall@20 | 0.8288 | ⚠️ Warning | 0.85 |

**Note**: Precision targets may need adjustment based on production data with real relevance judgments.

## Makefile Targets

```bash
# Run evaluation
make evaluate-search

# Run evaluation with JSON output
make evaluate-search-json

# Run A/B test
make search-ab-test  # (if target added to Makefile)
```

## Architecture

### Components

```
backend/
├── cmd/
│   ├── evaluate-search/      # CLI tool for running evaluations
│   └── search-ab-test/        # CLI tool for A/B testing
├── internal/services/
│   ├── search_evaluation_service.go    # Core evaluation logic
│   └── search_ab_testing.go            # A/B testing harness
├── testdata/
│   └── search_evaluation_dataset.yaml  # Labeled queries
└── scripts/
    └── generate_search_dataset.py      # Dataset generation
```

### Evaluation Flow

1. Load labeled dataset from YAML
2. For each query, retrieve search results
3. Map results to relevance judgments
4. Calculate metrics (nDCG, MRR, Precision, Recall)
5. Aggregate across all queries
6. Compare against targets
7. Generate report

### A/B Testing Flow

1. Define two configurations (A and B)
2. Run evaluation for Config A
3. Run evaluation for Config B
4. Calculate percentage improvements
5. Generate recommendation based on changes
6. Provide statistical summary

## Future Enhancements

- [ ] Live search integration (currently uses simulated results)
- [ ] Statistical significance testing (t-tests, bootstrap)
- [ ] Query segmentation (by type, difficulty, language)
- [ ] Diversity metrics (result diversity, game coverage)
- [ ] User behavior simulation (click models)
- [ ] Temporal evaluation (track metrics over time)
- [ ] Query intent classification
- [ ] Personalization evaluation

## References

- [nDCG Wikipedia](https://en.wikipedia.org/wiki/Discounted_cumulative_gain)
- [Mean Reciprocal Rank](https://en.wikipedia.org/wiki/Mean_reciprocal_rank)
- [Precision and Recall](https://en.wikipedia.org/wiki/Precision_and_recall)
- Roadmap Issue: subculture-collective/clpr#805

## Support

For questions or issues with the evaluation framework, please:

1. Check this README
2. Review the evaluation dataset format
3. Run with `-verbose` flag for debugging
4. Open an issue linking to Roadmap tracker #805
