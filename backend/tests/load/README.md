# Release load tests

`release.js` covers feed, clip detail, search, comments, authentication,
submission, rate limiting, and moderation with versioned latency/error
thresholds. It has three profiles:

- `baseline`: five virtual users for one minute; required before a candidate.
- `stress`: ramp through 25 and 75 virtual users, then return to zero.
- `soak`: ten users for 30 minutes by default; set `SOAK_DURATION` for the
  pre-release window.

Each user waits four seconds between actions. This models sustained browsing
within the ordinary 1,000-request/hour IP abuse budget, with at most 900
journey requests per user per hour. The separate burst scenario must still
receive HTTP 429. Do not disable abuse detection, clear bans, or accept denied
journey requests to make qualification pass.

Use independent client addresses for virtual users (`STAGING_LOCAL_IPS` in the
staging runner), and allocate a fresh address range for each qualification run.
Reusing addresses from earlier load or abuse tests can contaminate the hourly
budget. Record pacing and addresses alongside the fixture revision. At 75
users this profile offers at most 18.75 journey requests/second, before response
time and authentication overhead; a pass establishes capacity for this workload,
not maximum unthrottled throughput or behavior behind a shared NAT address.

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
Submission and moderation journeys read their authenticated lists by default;
the real-browser suite separately verifies persisted writes and authorization.
