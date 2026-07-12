# Release load tests

`release.js` covers feed, clip detail, search, comments, authentication,
submission, rate limiting, and moderation with versioned latency/error
thresholds. It has three profiles:

- `baseline`: five virtual users for one minute; required before a candidate.
- `stress`: increasing load to identify the first failed threshold.
- `soak`: ten users for 30 minutes by default; set `SOAK_DURATION` for the
  pre-release window.

Use a staging database seeded with a public `CLIP_ID`, a disposable user
`AUTH_TOKEN`, and a disposable moderator `ADMIN_TOKEN`. Set `REQUIRE_MUTATIONS`
and `SUBMISSION_URL` only where duplicate submissions can be safely discarded.
Never run mutation or rate-limit scenarios against production.

```bash
BASE_URL=https://staging.clpr.tv \
CLIP_ID=00000000-0000-0000-0000-000000000001 \
AUTH_TOKEN=load-user-token ADMIN_TOKEN=load-admin-token \
k6 run backend/tests/load/release.js

PROFILE=stress k6 run backend/tests/load/release.js
PROFILE=soak SOAK_DURATION=2h k6 run backend/tests/load/release.js
```

Store the k6 summary, candidate image digest, fixture revision, database size,
instance count, and first saturated resource as release evidence.
