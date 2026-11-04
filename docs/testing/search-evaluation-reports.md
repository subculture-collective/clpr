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

```

## Architecture

### Components

```
backend/
├── cmd/
│   ├── evaluate-search/      # CLI tool for running evaluations
├── internal/services/
│   ├── search_evaluation_service.go    # Core evaluation logic
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
