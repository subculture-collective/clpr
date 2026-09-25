---
title: "RECOMMENDATION EVALUATION"
summary: "This document describes the evaluation framework for the recommendation algorithm, including metrics definitions, baseline measurements, and guidelines for continuous improvement."
tags: ["docs"]
area: "docs"
status: "stable"
owner: "team-core"
version: "1.0"
last_reviewed: 2026-01-29
---

# Recommendation Algorithm Evaluation Framework

## Overview

This document describes the evaluation framework for the recommendation algorithm, including metrics definitions, baseline measurements, and guidelines for continuous improvement.

## Metrics Defined

### Accuracy Metrics

#### Precision@k
**Definition**: The fraction of recommended items in the top k that are relevant.

```
Precision@k = (# of relevant items in top k) / k
```

**Targets**:
- Precision@5: **≥ 0.70** (70% of top 5 recommendations should be relevant)
- Precision@10: **≥ 0.60** (60% of top 10 recommendations should be relevant)

**Thresholds**:
- Warning: Precision@5 < 0.70, Precision@10 < 0.60
- Critical: Precision@5 < 0.50, Precision@10 < 0.40

#### Recall@k
**Definition**: The fraction of all relevant items that are present in the top k recommendations.

```
Recall@k = (# of relevant items in top k) / (total # of relevant items)
```

**Targets**:
- Recall@5: **≥ 0.50** (capture at least 50% of relevant items in top 5)
- Recall@10: **≥ 0.70** (capture at least 70% of relevant items in top 10)

**Thresholds**:
- Warning: Recall@5 < 0.50, Recall@10 < 0.70
- Critical: Recall@5 < 0.30, Recall@10 < 0.50

### Ranking Quality

#### nDCG@k (Normalized Discounted Cumulative Gain)
**Definition**: Measures ranking quality by giving more weight to highly relevant items ranked higher.

```
DCG@k = Σ(2^relevance - 1) / log2(position + 1)
nDCG@k = DCG@k / IDCG@k
```

Where IDCG (Ideal DCG) is the DCG of the perfect ranking.

**Targets**:
- nDCG@5: **≥ 0.75** (good ranking quality)
- nDCG@10: **≥ 0.70** (good ranking quality)

**Thresholds**:
- Warning: nDCG@5 < 0.75, nDCG@10 < 0.70
- Critical: nDCG@5 < 0.55, nDCG@10 < 0.50

**Relevance Scale**:
- 0: Not relevant
- 1: Marginally relevant
- 2: Fairly relevant
- 3: Highly relevant
- 4: Perfect match

### Diversity Metrics

#### Diversity@k
**Definition**: Number of unique games represented in the top k recommendations.

```
Diversity@k = # of unique game_ids in top k
```

**Targets**:
- Diversity@5: **≥ 3.0 games** (variety in top 5)
- Diversity@10: **≥ 5.0 games** (variety in top 10)

**Rationale**: Prevents recommendation bubbles and ensures users discover content from different games.

**Thresholds**:
- Warning: Diversity@5 < 3.0, Diversity@10 < 5.0
- Critical: Diversity@5 < 2.0, Diversity@10 < 3.0

### Serendipity

**Definition**: The fraction of relevant recommendations that come from games NOT in the user's favorite games list.

```
Serendipity = (# of relevant non-favorite items) / (# of relevant items)
```

**Target**: **≥ 0.25** (at least 25% of relevant recommendations should be serendipitous)

**Rationale**: Balances personalization with discovery. Helps users find new content they'll enjoy.

**Thresholds**:
- Warning: < 0.25
- Critical: < 0.10

### Cold-Start Performance

#### Cold-Start Precision@5
**Definition**: Precision@5 specifically for users with no or minimal interaction history.

**Target**: **≥ 0.60** (maintain 60% precision even for new users)

**Rationale**: Ensures good experience for new users through trending and popular content.

**Thresholds**:
- Warning: < 0.60
- Critical: < 0.40

## Baseline Measurements

**Version**: 1.0  
**Date**: 2024-12-30  
**Dataset**: recommendation_evaluation_dataset.yaml v1.0  
**Scenarios**: 8 evaluation scenarios (2 cold-start, 6 active users)

### Simulated Ideal Performance

Using simulated perfect ranking (for baseline establishment):

| Metric | Value | Status | Target |
|--------|-------|--------|--------|
| **Precision@5** | 0.9750 | ✅ Pass | 0.70 |
| **Precision@10** | 0.5125 | ⚠️ Warning | 0.60 |
| **Recall@5** | 0.9583 | ✅ Pass | 0.50 |
| **Recall@10** | 1.0000 | ✅ Pass | 0.70 |
| **nDCG@5** | 1.0000 | ✅ Pass | 0.75 |
| **nDCG@10** | 1.0000 | ✅ Pass | 0.70 |
| **Diversity@5** | 3.88 | ✅ Pass | 3.00 |
| **Diversity@10** | 5.25 | ✅ Pass | 5.00 |
| **Serendipity** | 0.5000 | ✅ Pass | 0.25 |

### Cold-Start Performance

| Metric | Value | Status | Target |
|--------|-------|--------|--------|
| **Cold-Start Precision@5** | 0.9000 | ✅ Pass | 0.60 |
| **Cold-Start Recall@5** | 1.0000 | ✅ Pass | - |
| **Cold-Start Scenarios** | 2 | - | - |

**Note**: Precision@10 shows warning status in simulated results because dataset has limited items per scenario (typically 6-7 items). In production with larger candidate sets, this metric should improve.

## Evaluation Dataset

### Structure

The evaluation dataset (`backend/testdata/recommendation_evaluation_dataset.yaml`) contains:

- **8 evaluation scenarios** covering:
  - Active users with clear preferences (content-based)
  - New users (cold-start, trending)
  - Users with diverse interests (collaborative filtering)
  - Discovery-oriented users (serendipity testing)
  - Single-game enthusiasts (diversity enforcement)
  - Casual viewers (hybrid algorithm)

- **User profiles** including:
  - Favorite games
  - Followed streamers
  - Preferred categories
  - Interaction history

- **Relevant clips** with:
  - Relevance scores (0-4 scale)
  - Game IDs (for diversity calculation)
  - Reasons for relevance

### Versioning

Dataset versions are tracked in the YAML file:
- `version`: Semantic version (currently "1.0")
- `created_at`: Initial creation date
- `last_updated`: Last modification date

**Important**: When updating the dataset, increment the version and document changes. Re-establish baselines for major version changes.

## Running Evaluations

### Via Makefile

```bash
# Run evaluation with verbose output
make evaluate-recommendations

# Run evaluation and save JSON results
make evaluate-recommendations-json
```

### Via CLI Tool

```bash
# Build the tool
cd backend
go build -o bin/evaluate-recommendations ./cmd/evaluate-recommendations

# Run with default dataset
./bin/evaluate-recommendations -verbose

# Run with custom dataset and output
./bin/evaluate-recommendations -dataset path/to/dataset.yaml -output results.json

# Get help
./bin/evaluate-recommendations -help
```

### Via CI/CD

The evaluation runs automatically:
- **Nightly**: Every day at 2 AM UTC
- **On changes**: When recommendation code or dataset changes
- **Manual**: Via GitHub Actions workflow dispatch

Results are stored as artifacts for 90 days.

## Interpreting Results

### Status Indicators

- ✅ **Pass**: Metric meets or exceeds target
- ⚠️ **Warning**: Metric is below target but above critical threshold
- ❌ **Critical**: Metric is below critical threshold (CI fails)

### Per-Scenario Analysis

When running with `-verbose`, review per-scenario results to identify:
- **Low precision**: Irrelevant recommendations being shown
- **Low recall**: Relevant items not being surfaced
- **Low diversity**: Too many recommendations from same game
- **Low serendipity**: Only showing familiar content
- **Poor nDCG**: Good items ranked too low

### Algorithm Comparison

Different scenarios test different algorithms:
- **content**: Tests content-based filtering
- **collaborative**: Tests collaborative filtering
- **trending**: Tests cold-start/trending approach
- **hybrid**: Tests combined approach

Compare performance across algorithms to guide optimization priorities.

## Continuous Improvement

### Establishing New Baselines

When making significant algorithm changes:

1. Run evaluation before changes (capture baseline)
2. Make changes
3. Run evaluation after changes
4. Compare metrics
5. If improved, update baseline documentation

### Tracking Progress

Use GitHub Actions artifacts to track metrics over time:

```bash
# Download artifact from workflow run
gh run download <run-id> -n recommendation-evaluation-results-<number>

# Compare with previous results
jq '.aggregate_metrics' recommendation-evaluation-results.json
```

### Dataset Maintenance

Periodically review and update the evaluation dataset:

1. **Add new scenarios**: Cover edge cases or new features
2. **Update relevance judgments**: Refine based on user behavior
3. **Add diversity**: Ensure coverage of different user types
4. **Version properly**: Increment version and document changes

### Optimization Priorities

Focus optimization efforts based on metric status:

1. **Critical failures**: Immediate attention required
2. **Warning status**: Next priority for improvement
3. **Passing metrics**: Maintain or enhance further

## Links and References

- Issue: [subculture-collective/clpr#805](https://git.subcult.tv/subculture-collective/clpr/issues/805)
- Recommendation Service: `backend/internal/services/recommendation_service.go`
- Evaluation Service: `backend/internal/services/recommendation_evaluation_service.go`
- Dataset: `backend/testdata/recommendation_evaluation_dataset.yaml`
- CI Workflow: `.github/workflows/recommendation-evaluation.yml`

## Metric Definitions Codified

All metric calculations are implemented in:
- `backend/internal/services/recommendation_evaluation_service.go`
- Functions: `CalculatePrecisionAtK`, `CalculateRecallAtK`, `CalculateNDCG`, `CalculateDiversity`, `CalculateSerendipity`, `CalculateCoverage`

Unit tests validate metric calculations in:
- `backend/internal/services/recommendation_evaluation_service_test.go`

All metrics use standard IR (Information Retrieval) definitions and formulas.
