# Abuse detection

Request abuse protection is implemented in
[AbuseDetectionMiddleware](../../backend/internal/middleware/abuse_detection_middleware.go).
It uses Redis to track request activity and temporary IP bans, with exclusions
for health and authentication endpoints. Route-specific rate limits and
submission abuse checks provide additional boundaries.

The request middleware is covered by
[its Redis-backed tests](../../backend/internal/middleware/abuse_detection_middleware_test.go).
Submission checks are covered by
[submission abuse tests](../../backend/internal/services/submission_abuse_detection_test.go).
Run the [release-critical backend tier](../testing/TESTING.md) with disposable
services to verify required middleware cases without skips.

The previous anomaly-scoring and auto-flagging service graph was never wired
into the API. Its implementation and proposed analytics endpoints have been
removed. It is not part of the active abuse protection contract.
