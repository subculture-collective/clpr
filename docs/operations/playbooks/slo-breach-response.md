# SLO breach response

Owner: on-call engineer. Escalation owner: release lead. Security or privacy
symptoms also page the security owner.

## Immediate response

1. Acknowledge within 15 minutes and identify the affected `journey` label.
2. Check deployment time, `http_requests_total`, latency histograms, dependency
   health, queue lag, and the current error-budget burn.
3. Stop an active rollout. Roll back when the breach began with the candidate
   and rollback is safer than forward repair.
4. For dependency failure, keep required dependencies unready and explicitly
   report optional dependency degradation. Never return fabricated success.
5. Validate recovery with both synthetic traffic and the journey smoke test.

## Availability breach

Group failures by status and route template. Check database saturation and
timeouts, Redis availability, external-provider errors, panics, and rate-limit
configuration. Do not classify expected `4xx` responses as downtime.

## Latency breach

Inspect p50/p95/p99, in-flight requests, database acquire duration, slow-query
logs, and downstream latency. Prefer shedding optional work or disabling an
optional feature over allowing unbounded queues or database waits.

## Resolution

Resolve only after the alert is green for 15 minutes and the critical smoke
passes. Record start/end time, user impact, failed SLI, mitigation, rollback or
release digest, remaining error budget, and follow-up owner/due date. A breach
caused by a release blocks promotion until its corrective test is merged.

