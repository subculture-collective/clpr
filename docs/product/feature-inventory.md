---
title: "Feature inventory"
summary: "Source-backed scope and verification boundaries for the current product."
tags: ["product", "status"]
area: "product"
status: "stable"
owner: "team-core"
version: "1.0"
last_reviewed: 2026-11-16
---

# Feature inventory and verification map

Reviewed November 16, 2026 against main `5591a517`. This is an implementation
map, not a production acceptance certificate. A closed issue or a passing unit
test alone does not establish live operation. The
[launch feature inventory](../LAUNCH_FEATURE_INVENTORY.md) controls disabled
scope; [development status](development-status.md) records outstanding work.

## Maintained implementation

| Capability | Source and verification references | Qualification boundary |
| --- | --- | --- |
| Twitch identity, sessions and MFA | `backend/cmd/api/routes_auth.go`, `backend/internal/handlers/login_session.go`, `backend/internal/handlers/mfa_login_test.go` | MFA is retained by the owner's November 7 decision in [#1395](https://github.com/subculture-collective/clpr/issues/1395); it was not removed. Fresh provider sign-in and enrolled/staff MFA need live evidence. |
| Clip discovery and submission | `backend/cmd/api/routes_content.go`, `frontend/src/pages/SubmitClipPage.tsx`, [clip API](../backend/clip-api.md) | Lists, filters, existing Twitch URLs and reviewed archive handoffs are maintained; extraction from live streams is removed. |
| Text, hybrid and transcript search | `backend/internal/search`, [transcript search](../features/transcript-search.md), [search artifact](../development/search-wasm.md) | Transcript matches require actual transcripts/indexing. Browser Wasm is separately tested and versioned; it does not prove production deployment. |
| Saved clips, playlists, queue and history | `frontend/src/components/clip`, [playlists](../features/feature-playlists.md), [queue/history](../features/feature-queue-history.md) | Account permissions, paging and real playback are checked through the relevant backend/browser suites. |
| Comments, votes, reports and site moderation | [comments](../features/comments.md), `backend/internal/handlers/moderation_handler.go`, `frontend/src/pages/admin` | Submission review, reports, bans, appeals, audit and DMCA stay. Platform roles and community-scoped roles are distinct. |
| Countdown selection, fetch, render and public pages | [countdown guide](../features/countdown-editions.md), [operations runbook](../operations/runbooks/countdowns.md) | Enabled series and actual edition outcomes must be inspected; rendered media is not proof of a YouTube publication. |
| YouTube publisher, metrics and alerts | `backend/internal/services/countdown_youtube.go`, `monitoring/prometheus/clpr-countdown-rules.test.yml` | Code exists. [#1408](https://github.com/subculture-collective/clpr/issues/1408) and [#1409](https://github.com/subculture-collective/clpr/issues/1409) retain provider and operational acceptance. |
| Creator connect, placements, clip packs, digest and clip-room series | [countdown guide](../features/countdown-editions.md), `frontend/src/pages/CreatorDashboardPage.tsx`, `backend/cmd/countdown-media` | Source work is closed in the streamer-service milestone. Packs need a signing secret/public route; email and YouTube features require their own configuration. Do not imply they run from issue closure alone. |
| Outbound webhooks and email infrastructure | `backend/internal/services`, [webhook subscriptions](../backend/webhook-subscription-management.md) | External delivery requires configuration and provider receipts; the email service was retained with MFA and digests. |
| Community APIs | `backend/cmd/api/routes_content.go`, `backend/internal/repository/community_repository.go` | Backend scope does not establish a complete community product UI or private invitation lifecycle. |
| Native iOS and Android | [New native app plan](native-apps-plan.md) | Planning only. No native workspace, binary, platform CI or device qualification exists. |

## Disabled and retired scope

Automated account deletion and full personal-data export remain disabled until
their implementation and acceptance gates pass. Creator-content exports are a
separate capability. Native planning must resolve account/privacy requirements
before promising support or store eligibility.

Paid subscriptions/Stripe runtime, watch parties, live feed/stream extraction,
site chat, forums, advertising, CDN orchestration/mirroring, collaborative
filtering, account merge/types and creator verification were retired. The
streamer Twitch chat listener remains. Retired-feature tables were dropped in
migration 148 ([#1401](https://github.com/subculture-collective/clpr/issues/1401));
retained billing/audit history follows its separate retention contract.

## Verification and delivery

The maintained complete source gate is `scripts/source-convergence.sh --all`,
run by GitHub Mirror CI on branches. It covers backend tests and integration
contracts, frontend type/lint/build/budget checks, browser journeys and
repository/documentation checks. Use the actual run to establish coverage for
an exact source tree. Gitea main release gates additionally qualify browser
engines and immutable images.

Frontend-only delivery verifies a signed image and a host admission contract.
Backend source, schema, runtime, startup and image environment changes can
require a native release and reviewed admission refresh. Neither the presence
of workflows nor a healthy API response proves complete release acceptance.
[Release evidence](../../release-evidence/README.md) describes the separate
operator contracts.

There is no percentage-complete estimate. The former inventory mixed planned,
removed and unverified features and listed obsolete counts and dates. Add a
capability here only with source, tests, scope and its qualification limits.
