---
title: "Development status and active plans"
summary: "November 2026 source progress, open acceptance work and native app planning."
tags: ["product", "status", "planning"]
area: "product"
status: "stable"
owner: "team-core"
version: "1.0"
last_reviewed: 2026-11-16
---

# Development status and active plans

This is a dated source/tracker snapshot, reconciled November 16, 2026. Use
[live GitHub issues](https://github.com/subculture-collective/clpr/issues),
[PRs](https://github.com/subculture-collective/clpr/pulls) and per-revision CI
and deployment receipts for subsequent changes. There is no percentage-complete
estimate or promised release date.

## Current product

The responsive web app and Go API implement clip discovery, search, saved clips,
playlists, queue/history, submissions, comments, creator tools and moderation.
Countdown and streamer-service code is present; configuration and actual
provider outcomes remain separate evidence. See the
[feature verification map](feature-inventory.md) and
[launch scope](../LAUNCH_FEATURE_INVENTORY.md).

MFA remains supported. The owner chose to retain it on November 7 in
[#1395](https://github.com/subculture-collective/clpr/issues/1395), and the staff
requirement was repaired. That issue's historical removal title must not be
read as evidence of removal. Automated account deletion and full personal-data
export are still disabled.

## November plan

| Workstream | Recorded state | Remaining boundary |
| --- | --- | --- |
| [Stabilize and shrink](https://github.com/subculture-collective/clpr/milestone/6) | Closed, 17 tracked issues closed | Source/tracker completion; removed scope is listed in the launch contract. MFA was retained. |
| [Countdowns and Shorts](https://github.com/subculture-collective/clpr/milestone/5) | Open, eight issues closed and two acceptance issues open | YouTube audit/private-upload acceptance, then seven consecutive publishing days and live alert-delivery proof. |
| [Streamer service](https://github.com/subculture-collective/clpr/milestone/7) | Closed after reconciliation, five tracked issues closed | Clip-pack signing/routing, email configuration, real provider results and natural scheduled execution require their own evidence. |
| [Separate native apps](https://github.com/subculture-collective/clpr/milestone/8) | Planning started by owner request November 16; viewer-first beta confirmed | No app workspace, native binary or platform qualification yet; see the [new plan](native-apps-plan.md). |

The old MVP/Beta/GA/Post-GA buckets were closed as historical or retired
planning buckets. Their closure does not establish a beta or GA release.

## Open acceptance work

- [#1408: YouTube publisher](https://github.com/subculture-collective/clpr/issues/1408)
  has implementation and tests, but is blocked on provider acceptance. Record
  a real private test upload with matching metadata and an API audit-submission
  tracking link. Do not call code completion a public YouTube launch.
- [#1409: Series schedule, metrics, alerts and runbook](https://github.com/subculture-collective/clpr/issues/1409)
  has the metrics, rules, rule tests and [runbook](../operations/runbooks/countdowns.md).
  It requires seven consecutive automated publishing days and a controlled live
  failure that reaches the intended alert receiver. It depends on #1408.
- The [30-day marketing campaign](../marketing/clpr/launch-plan.md) remains a
  draft in this repository. It needs a selected start date/owner, checked
  product journeys, finished assets and actual feedback/measurement evidence.

## Release recovery and maintenance

The assessment found main `5591a517` passing
[Gitea release gates](https://git.subcult.tv/subculture-collective/clpr/actions/runs/22928)
but its [frontend delivery](https://git.subcult.tv/subculture-collective/clpr/actions/runs/22927)
failed. The host admission still pinned the previous API/schema and nginx image
environment. These are compatibility controls, not checks to disable. Native
release recovery must use the qualified images, migration rehearsal, backup,
health checks and a reviewed admission refresh before frontend activation.
Recovery on November 16 activated the CI-built backend/render images and the
signed frontend at `5591a517`. Migrations 161–162 passed an isolated schema
up/down/up rehearsal before production reached clean schema 162. A fresh
2.33 GB dump has a readable table of contents; a full data restore was not
repeated. Public build markers match all three uncached query variants, all
139 frontend asset hashes match, readiness/search/countdown requests pass, and
anonymous account/MFA requests stay unauthorized. The five admitted native
services stayed unchanged during frontend activation; the crawler and unrelated
containers kept their identities and start times across the recovery.

Private deployment receipts retain the exact images, native services, schema
and rollback. Future backend changes still require compatibility review; the
host admission and signed-delivery checks remain enabled. The earlier failed
CI run is historical; a subsequent main delivery must supply its own result.

[Dependency PR #1428](https://github.com/subculture-collective/clpr/pull/1428)
was repaired at `45961667`: latest main was merged without rewriting its
existing commits, the current Sentry build plugin retained, and accessible
names corrected for favorite counts and the mobile playback hint. Local
typecheck, lint, 28 focused component/page tests, the 22-test accessibility
file and its production build/bundle budget passed. Hosted checks on the final
head and current merge state remain the source of truth.

Branches, PRs and issues belong on GitHub. Gitea is the fast-forward code mirror
and continues to run homelab main release workflows. See
[release evidence](../../release-evidence/README.md) for the distinction between
source verification, signed frontend delivery and native operational acceptance.
