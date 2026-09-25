# Abuse detection

Request abuse protection is implemented in
[AbuseDetectionMiddleware](../../backend/internal/middleware/abuse_detection_middleware.go).
It counts every API request per client IP in Redis, in fixed one-minute and
one-hour windows, with separate budgets for two classes of traffic:

| Class | Requests | Default limit |
|-------|----------|---------------|
| Read | `GET`, `HEAD`, `OPTIONS`, and per-view telemetry (`track-view`, `events`, `ads/track`, `logs`, `track-share`, `clips/batch-media`) | 1,200 per minute, 12,000 per hour |
| Write | Other mutations, and every `/api/v1/auth/*` route except `GET /auth/me` | 120 per minute, 1,500 per hour |

`/health`, `/health/ready` and `/health/live` are not counted. Loopback
addresses and `RATE_LIMIT_WHITELIST_IPS` bypass the limiter.

Exceeding a limit blocks the IP with `429 Too Many Requests` and a
`Retry-After` header. Blocks escalate with each offense inside the offense
window: 15 minutes, 1 hour, 6 hours, then 24 hours for every later offense.
Parallel requests that cross the limit together count as one offense, and the
tripped counters are cleared when the block starts, so the IP starts fresh
when the block expires. If Redis fails, requests are allowed.

| Variable | Default |
|----------|---------|
| `ABUSE_DETECTION_ENABLED` | `true` |
| `ABUSE_READ_PER_MINUTE` / `ABUSE_READ_PER_HOUR` | `1200` / `12000` |
| `ABUSE_WRITE_PER_MINUTE` / `ABUSE_WRITE_PER_HOUR` | `120` / `1500` |
| `ABUSE_BAN_DURATIONS` | `15m,1h,6h,24h` |
| `ABUSE_OFFENSE_WINDOW` | `24h` |

Route-specific rate limits and submission abuse checks provide additional
boundaries.

## Client IP

Limits key on `c.ClientIP()`. In production, requests pass through
Cloudflare, cloudflared and the shared Caddy before reaching the backend.
Cloudflare appends the visitor address to any `X-Forwarded-For` the visitor
sent, and Caddy appends cloudflared's address. The backend trusts only the
proxies in `TRUSTED_PROXIES` (default: loopback and private networks) and
reads only `X-Forwarded-For`, so it takes the first untrusted address from the
right: the one Cloudflare added. A visitor-supplied `X-Forwarded-For`,
`X-Real-IP` or `CF-Connecting-IP` value cannot replace it.

## Tests

The request middleware is covered by
[its Redis-backed tests](../../backend/internal/middleware/abuse_detection_middleware_test.go),
and client IP derivation by
[the router tests](../../backend/cmd/api/router_test.go).
Submission checks are covered by
[submission abuse tests](../../backend/internal/services/submission_abuse_detection_test.go).
Run the [release-critical backend tier](../testing/TESTING.md) with disposable
services to verify required middleware cases without skips.

The previous anomaly-scoring and auto-flagging service graph was never wired
into the API. Its implementation and proposed analytics endpoints have been
removed. It is not part of the active abuse protection contract.
