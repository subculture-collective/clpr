# Search incident response

Owner: on-call engineer. Search is optional only when the launch feature
inventory says so; otherwise a complete loss of search is release-blocking.

## High latency

Check OpenSearch cluster health, request p95/p99, database fallback latency,
connection saturation, and recent index or mapping changes. Keep fallback
queries bounded by the database statement timeout. Disable semantic ranking
before disabling basic search when that restores a truthful degraded result.

## Zero results

Compare a known fixture through OpenSearch and the database fallback. Check
index document count, alias target, ingestion lag, filters, and feature flags.
Do not return an empty success when both providers failed; surface the defined
service-unavailable response.

## Embedding failures

Check embedding queue lag, provider errors, rate limits, model/version drift,
and dead-lettered jobs. Pause new embedding work if the queue is growing
without bound. Keyword search must remain independent of embedding health.

## Search failover

Block OpenSearch in staging, run the search smoke fixture, and verify bounded
database fallback, explicit degraded telemetry, and recovery after OpenSearch
returns. Record failover and recovery times plus the candidate image digest.
