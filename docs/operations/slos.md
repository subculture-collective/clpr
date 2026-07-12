# Service level objectives

These objectives cover the user journeys that gate a clpr release. They are
measured at the API boundary over a rolling 30-day window. Planned maintenance
is not excluded; a maintenance event consumes error budget just like any other
user-visible outage.

| Journey | Route matcher | Availability | p95 latency | 30-day error budget |
|---|---|---:|---:|---:|
| Authentication | `/api/v1/auth/.*` | 99.5% | 500 ms | 3 h 39 m |
| Feed | `/api/v1/(feed|feeds)(/.*)?` | 99.5% | 500 ms | 3 h 39 m |
| Search | `/api/v1/search(/.*)?` | 99.5% | 750 ms | 3 h 39 m |
| Clip playback/detail | `/api/v1/clips/[^/]+` | 99.5% | 500 ms | 3 h 39 m |
| Submission | `/api/v1/submissions(/.*)?` | 99.0% | 1 s | 7 h 18 m |
| Checkout | `/api/v1/subscriptions/(checkout|portal)` | 99.0% | 1.5 s | 7 h 18 m |
| Moderation | `/api/v1/(moderation|admin/moderation)(/.*)?` | 99.0% | 1 s | 7 h 18 m |

Availability counts responses other than `5xx` as successful API handling.
Expected validation, authentication, authorization, and rate-limit responses
are therefore not outages. Product-level correctness (for example entitlement
activation after a Stripe webhook) remains covered by journey contract tests.

## Release policy

- Page when the 1-hour error ratio is above 5% or p95 latency is above twice
  the journey target for 10 minutes.
- Create a ticket when the 6-hour error ratio exhausts more than 10% of the
  monthly budget.
- Freeze risky releases when a journey has consumed 75% of its monthly error
  budget. Only reliability fixes and approved emergency changes may ship.
- The release lead records the current 30-day SLO values in the candidate
  checklist. Missing telemetry is a no-go, not a passing value.

The executable alert and recording rules are in
`monitoring/prometheus/clpr-slo-rules.yml`. Respond using the
[SLO breach playbook](playbooks/slo-breach-response.md).

