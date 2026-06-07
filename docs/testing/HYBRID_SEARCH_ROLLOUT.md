---
title: "HYBRID SEARCH ROLLOUT"
summary: "This document outlines the rollout plan for optimized hybrid search weights (BM25 vs Vector) as part of Roadmap 5.0 Phase 3.1."
tags: ["docs"]
area: "docs"
status: "stable"
owner: "team-core"
version: "1.0"
last_reviewed: 2026-01-29
---

# Hybrid Search Weight Optimization - Rollout Plan

## Overview
This document outlines the rollout plan for optimized hybrid search weights (BM25 vs Vector) as part of Roadmap 5.0 Phase 3.1.

## Baseline Metrics

### Current Configuration (Baseline)
- **BM25 Weight**: 0.7
- **Vector Weight**: 0.3
- **Title Boost**: 3.0
- **Creator Boost**: 2.0
- **Game Boost**: 1.0
- **Engagement Boost**: 0.1
- **Recency Boost**: 0.5

### Baseline Performance
Captured using `capture-baseline-search` tool:
```bash
cd backend
go run cmd/capture-baseline-search/main.go -output baseline-search-metrics.json
```

Reference baseline file: `backend/baseline-search-metrics.json`

## Grid Search Results

### Methodology
Grid search performed using `grid-search-hybrid-search` tool across:
- **BM25 Weights**: 0.3, 0.4, 0.5, 0.6, 0.7, 0.8, 0.9
- **Vector Weights**: 0.1, 0.2, 0.3, 0.4, 0.5, 0.6, 0.7
- **Constraint**: Weights must sum to 1.0
- **Evaluation**: 510 labeled queries from search_evaluation_dataset.yaml
- **Primary Metric**: nDCG@10 (target: ≥10% improvement)

### Running Grid Search
```bash
cd backend
# Full grid search
go run cmd/grid-search-hybrid-search/main.go \
  -baseline baseline-search-metrics.json \
  -output hybrid-search-grid-results.json

# Quick mode (fewer combinations)
go run cmd/grid-search-hybrid-search/main.go \
  -quick \
  -baseline baseline-search-metrics.json \
  -output hybrid-search-grid-results-quick.json \
  -verbose
```

### Top Candidates
Results stored in: `backend/hybrid-search-grid-results.json`

The grid search will identify:
1. **Best Configuration**: Highest nDCG@10 score
2. **Improvement %**: Percentage improvement vs baseline
3. **Status**: Whether 10% improvement target is met

## Recommended Configuration

### Selected Weights
Based on grid search results, the recommended configuration will be documented here after running the grid search.

**Placeholder for optimal weights:**
- BM25 Weight: TBD
- Vector Weight: TBD
- Expected nDCG@10 improvement: TBD%

### Rationale
The configuration is selected based on:
1. **Primary**: nDCG@10 maximization (weight: 4.0 in scoring)
2. **Secondary**: nDCG@5, MRR balance (weight: 2.0 each)
3. **Supporting**: Precision@10 and Recall@10 (weight: 1.0 each)

## Configuration Management

### Environment Variables
The hybrid search weights are controlled via environment variables for easy rollout and rollback:

```bash
# Hybrid Search Configuration
HYBRID_SEARCH_BM25_WEIGHT=0.7        # Default: 0.7
HYBRID_SEARCH_VECTOR_WEIGHT=0.3      # Default: 0.3
HYBRID_SEARCH_TITLE_BOOST=3.0        # Default: 3.0
HYBRID_SEARCH_CREATOR_BOOST=2.0      # Default: 2.0
HYBRID_SEARCH_GAME_BOOST=1.0         # Default: 1.0
HYBRID_SEARCH_ENGAGEMENT_BOOST=0.1   # Default: 0.1
HYBRID_SEARCH_RECENCY_BOOST=0.5      # Default: 0.5
```

### Quick Rollback
If issues are detected, rollback to baseline by removing/commenting environment variables or setting to baseline values:

```bash
# Rollback to baseline
HYBRID_SEARCH_BM25_WEIGHT=0.7
HYBRID_SEARCH_VECTOR_WEIGHT=0.3
```

No code changes required - restart service with updated environment variables.

## Rollout Strategy

### Phase 1: Staging Deployment (Week 1)
**Objective**: Validate optimized weights in staging environment

1. **Deploy Configuration**:
   ```bash
   # Update .env.staging or k8s ConfigMap
   HYBRID_SEARCH_BM25_WEIGHT=<optimized_value>
   HYBRID_SEARCH_VECTOR_WEIGHT=<optimized_value>
   ```

2. **Validation**:
   - Run evaluation suite: `go run cmd/evaluate-search/main.go`
   - Monitor search metrics in staging
   - Test edge cases and error scenarios

3. **Success Criteria**:
   - ✅ nDCG@10 improvement ≥10% confirmed
   - ✅ No regression in other metrics (nDCG@5, MRR)
   - ✅ No errors or performance degradation

### Phase 2: Production Canary (Week 2)
**Objective**: Gradually roll out to production with monitoring

1. **Canary Deployment** (10% traffic):
   ```bash
   # Enable for 10% of users via A/B testing
   # Or deploy to single region/pod
   ```

2. **Monitor KPIs** (24-48 hours):
   - Search result click-through rate (CTR)
   - Search result dwell time
   - User engagement metrics
   - Zero-result search rate
   - Search latency (p50, p95, p99)

3. **Success Criteria**:
   - ✅ CTR improvement or stable
   - ✅ Dwell time improvement or stable
   - ✅ No increase in zero-result searches
   - ✅ Latency within SLA (p95 < 500ms)

### Phase 3: Production Full Rollout (Week 3)
**Objective**: Roll out to all production traffic

1. **Gradual Rollout**:
   - Day 1: 25% traffic
   - Day 2: 50% traffic
   - Day 3: 75% traffic
   - Day 4: 100% traffic

2. **Continuous Monitoring**:
   - Real-time metrics dashboard
   - Automated alerts for anomalies
   - User feedback monitoring

3. **Final Success Criteria**:
   - ✅ nDCG@10 improvement sustained
   - ✅ Positive or neutral user engagement
   - ✅ No critical incidents
   - ✅ System stability maintained

## Monitoring & Metrics

### Real-Time Metrics
Monitor via Prometheus/Grafana dashboards:

1. **Search Quality**:
   - CTR by query type
   - Zero-result search rate
   - Average result relevance

2. **User Engagement**:
   - Dwell time on search results
   - Bounce rate
   - Engagement rate

3. **System Performance**:
   - Search latency (p50, p95, p99)
   - BM25 search duration
   - Vector search duration
   - Cache hit rate

### Alerts
Configure alerts for:
- nDCG@10 drops > 5% below target
- Search latency p95 > 500ms
- Zero-result search rate increase > 20%
- Error rate increase > 1%

## Rollback Procedure

### Trigger Conditions
Rollback if any of the following occur:
- nDCG@10 improvement < 5% (half of target)
- Search latency increases > 50ms
- CTR drops > 10%
- Critical incidents or errors

### Rollback Steps
1. **Immediate** (< 5 minutes):
   ```bash
   # Revert environment variables to baseline
   HYBRID_SEARCH_BM25_WEIGHT=0.7
   HYBRID_SEARCH_VECTOR_WEIGHT=0.3
   
   # Restart service
   kubectl rollout restart deployment/backend
   ```

2. **Verify Rollback**:
   - Check metrics return to baseline
   - Run evaluation suite
   - Monitor for 30 minutes

3. **Post-Mortem**:
   - Document issue
   - Analyze root cause
   - Plan remediation

## Success Metrics

### Primary
- **nDCG@10**: ≥10% improvement vs baseline ✅

### Secondary
- **CTR**: No degradation (ideally +5%)
- **Dwell Time**: No degradation (ideally +10%)
- **Zero-Result Rate**: No increase
- **Latency**: p95 < 500ms maintained

## Documentation & Communication

### Stakeholder Updates
- **Week 0**: Share baseline metrics and plan
- **Week 1**: Share staging results
- **Week 2**: Share canary results
- **Week 3**: Share full rollout results
- **Week 4**: Final report and retrospective

### Documentation Updates
- [x] Update environment variable documentation
- [ ] Update search architecture docs
- [ ] Update deployment runbooks
- [ ] Update monitoring dashboards

## Dependencies

### Related Issues
- [#837](https://git.subcult.tv/subculture-collective/clpr/issues/837) - Search Relevance Evaluation Framework (completed)
- [#805](https://git.subcult.tv/subculture-collective/clpr/issues/805) - Related search improvements

### Technical Dependencies
- Evaluation dataset: `backend/testdata/search_evaluation_dataset.yaml`
- Configuration system: `backend/config/config.go`
- Search services: `backend/internal/services/hybrid_search_service.go`

## Timeline

| Week | Phase | Activities |
|------|-------|------------|
| Week 0 | Preparation | Capture baseline, run grid search, document plan |
| Week 1 | Staging | Deploy to staging, validate metrics, test edge cases |
| Week 2 | Canary | Deploy to 10% production, monitor KPIs |
| Week 3 | Rollout | Gradual rollout to 100%, continuous monitoring |
| Week 4 | Stabilization | Final validation, documentation, retrospective |

## Team & Responsibilities

- **Engineering Lead**: Configuration deployment, monitoring
- **Data Science**: Metric analysis, validation
- **DevOps**: Deployment automation, rollback procedures
- **Product**: User engagement tracking, success criteria validation

## Notes

### Assumptions
- Semantic search feature flag is enabled (`FEATURE_SEMANTIC_SEARCH=true`)
- Embedding service is operational
- Sufficient evaluation data (510 queries)

### Risks & Mitigations
1. **Risk**: Grid search doesn't find 10% improvement
   - **Mitigation**: Document rationale, consider alternative approaches (field boost tuning, candidate pool size)

2. **Risk**: Production behavior differs from offline evaluation
   - **Mitigation**: Canary deployment with real user traffic, quick rollback capability

3. **Risk**: Latency increase with different weights
   - **Mitigation**: Performance testing in staging, latency SLA monitoring

## References

- [Search Evaluation Framework Docs](../docs/search-evaluation.md)
- [Hybrid Search Architecture](../docs/hybrid-search.md)
- [Configuration Management](../docs/configuration.md)
- [Deployment Procedures](../docs/deployment.md)
