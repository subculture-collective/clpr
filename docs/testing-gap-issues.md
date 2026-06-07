---
title: "Testing Gap Issues"
summary: "This document contains ready-to-file GitHub issue content for closing testing gaps identified in"
tags: ["docs","testing"]
area: "docs"
status: "stable"
owner: "team-core"
version: "1.0"
last_reviewed: 2026-01-29
---

# Testing Coverage Gap Issues (from Feature Test Coverage Audit a9be649)

This document contains ready-to-file GitHub issue content for closing testing gaps identified in
`docs/product/feature-test-coverage.md` (commit a9be649). Each issue is small, well-defined, and
includes clear acceptance criteria, labels, and file pointers. A master tracker is provided at the top
to group issues by phase and priority.

---

## Master Tracker: Testing Coverage Gaps ✅

Live Tracker Issue: [#900](https://git.subcult.tv/subculture-collective/clpr/issues/900)

- Phase 1 — Critical Security & Compliance (P0)
  - [ ] Issue: DMCA Handler Test Suite (P0, MVP) — [#917](https://git.subcult.tv/subculture-collective/clpr/issues/917)
  - [ ] Issue: GDPR Account Deletion Lifecycle Tests (P0, MVP) — [#904](https://git.subcult.tv/subculture-collective/clpr/issues/904)
  - [ ] Issue: Admin User Management Authorization Tests (P0, MVP) — [#912](https://git.subcult.tv/subculture-collective/clpr/issues/912)
  - [ ] Issue: Authorization Test Suite (RBAC Endpoints) (P0, MVP) — [#901](https://git.subcult.tv/subculture-collective/clpr/issues/901)
  - [ ] Issue: Validation Middleware Security Tests (P0, MVP) — [#914](https://git.subcult.tv/subculture-collective/clpr/issues/914)

- Phase 2 — Infrastructure Reliability (P0/P1)
  - [ ] Issue: Deployment Scripts Test Harness & Smoke Tests (P0, Beta) — [#903](https://git.subcult.tv/subculture-collective/clpr/issues/903)
  - [ ] Issue: Database Migration Rollback Tests (P0, Beta) — [#902](https://git.subcult.tv/subculture-collective/clpr/issues/902)
  - [ ] Issue: Backup & Restore Validation (P0, Beta) — [#913](https://git.subcult.tv/subculture-collective/clpr/issues/913)
  - [ ] Issue: Monitoring Alert Rule Validation (P1, Beta) — [#907](https://git.subcult.tv/subculture-collective/clpr/issues/907)

- Phase 3 — Feature Completeness (P1)
  - [ ] Issue: Mobile E2E Test Suite — Core Flows (P1, Beta) — [#910](https://git.subcult.tv/subculture-collective/clpr/issues/910)
  - [ ] Issue: Discovery Lists — Unit + Integration + E2E Coverage (P1, Beta) — [#906](https://git.subcult.tv/subculture-collective/clpr/issues/906)
  - [ ] Issue: Live Status Tracking — Integration Tests (P1, Beta) — [#911](https://git.subcult.tv/subculture-collective/clpr/issues/911)
  - [ ] Issue: Moderation Workflow — E2E Coverage (P1, Beta) — [#915](https://git.subcult.tv/subculture-collective/clpr/issues/915)
  - [ ] Issue: Watch Party Real-time Sync Tests (P1, Beta) — [#905](https://git.subcult.tv/subculture-collective/clpr/issues/905)

- Phase 4 — Performance & Optimization (P2)
  - [ ] Issue: Rate Limiting — Load Tests (P2, GA) — [#908](https://git.subcult.tv/subculture-collective/clpr/issues/908)
  - [ ] Issue: Search Fallback Performance & Failover Tests (P2, GA) — [#897](https://git.subcult.tv/subculture-collective/clpr/issues/897)
  - [ ] Issue: CDN Failover Simulation Tests (P2, GA) — [#898](https://git.subcult.tv/subculture-collective/clpr/issues/898)
  - [ ] Issue: Webhook Delivery at Scale — Load & DLQ Replay (P2, GA) — [#899](https://git.subcult.tv/subculture-collective/clpr/issues/899)

Notes:
- Labels: use `kind/chore`, `area/testing`, plus area-specific labels (e.g., `area/backend`, `area/mobile`,
  `area/infrastructure`) and priorities (`priority/P0`, `priority/P1`, `priority/P2`).
- Milestones: map Phase 1→MVP, Phase 2→Beta, Phase 3→Beta, Phase 4→GA.
- Link existing coverage to avoid duplication (e.g., existing Mobile E2E parent, Backup/Restore epics).

---

## Phase 1 — Critical Security & Compliance (P0)


### 1) [Testing] DMCA Handler Test Suite

- Labels: `kind/chore`, `area/testing`, `area/backend`, `priority/P0`, `milestone/MVP`
- Summary: Add comprehensive unit and integration tests for DMCA takedown handling to ensure legal compliance.
- Scope:
  - Backend: `backend/internal/handlers/dmca_handler.go`
  - Tests: `backend/internal/handlers/dmca_handler_test.go`, `backend/tests/integration/dmca/dmca_integration_test.go`
  - Cases: takedown request intake, validation (required fields, authenticity), moderation linkage,
    notification, audit logging, appeal workflow stubs.
- Acceptance Criteria:
  - [ ] Unit tests cover request validation (required fields, malformed input) and handler branches (success, error).
  - [ ] Integration tests verify: takedown creation, status transitions, audit log entries, and notifications dispatched.
  - [ ] Unauthorized and non-admin access correctly returns 403.
  - [ ] Negative tests: invalid request rejected; duplicate takedown prevented.
  - [ ] Coverage report for handler ≥ 80% lines/branches.


### 2) [Testing] GDPR Account Deletion Lifecycle Tests

- Labels: `kind/chore`, `area/testing`, `area/backend`, `priority/P0`, `milestone/MVP`
- Summary: Implement tests for full account deletion lifecycle (request → soft lock → grace period → hard delete),
  ensuring GDPR compliance.
- Scope:
  - Backend: `backend/internal/services/user_settings_service.go`, relevant handlers for deletion requests.
  - Tests: `backend/tests/integration/users/account_deletion_integration_test.go`
  - Validate data erasure across user-owned resources (clips, favorites, votes), and audit trail entries.
- Acceptance Criteria:
  - [ ] Integration tests confirm staged deletion with grace period and irreversible final deletion.
  - [ ] Deleted accounts cannot authenticate and all personal data is removed/anonymized.
  - [ ] Linked content ownership reassigned or removed according to policy; orphan checks included.
  - [ ] Export endpoint returns no personal data post-deletion.
  - [ ] Coverage added for error/retry paths and cancellation flow.


### 3) [Testing] Admin User Management Authorization Tests

- Labels: `kind/chore`, `area/testing`, `area/backend`, `priority/P0`, `milestone/MVP`
- Summary: Add integration tests for admin user management endpoints to validate strict authorization and audit logging.
- Scope:
  - Backend: `backend/internal/handlers/admin_user_handler.go`
  - Tests: `backend/tests/integration/admin/admin_user_management_test.go`
  - Scenarios: create/update roles, suspend/unsuspend, password reset, audit logging.
- Acceptance Criteria:
  - [ ] Non-admin requests receive 403; admin-only operations succeed.
  - [ ] Role changes are persisted and applied immediately to permissions.
  - [ ] Each admin operation writes a correct audit log entry.
  - [ ] Negative tests: privilege escalation attempts rejected.
  - [ ] Coverage for handler ≥ 80%.


### 4) [Testing] Authorization Test Suite (RBAC Endpoints)

- Labels: `kind/chore`, `area/testing`, `area/backend`, `priority/P0`, `milestone/MVP`
- Summary: Create a reusable integration test suite that iterates protected endpoints to validate
  role-based access control (RBAC).
- Scope:
  - Middleware: `backend/internal/middleware/permission_middleware.go`
  - Endpoints: admin moderation, clip delete, webhook DLQ admin, etc.
  - Tests: `backend/tests/integration/auth/rbac_endpoints_test.go`
- Acceptance Criteria:
  - [ ] For each protected endpoint, tests assert access matrix (guest/user/mod/admin) with expected status (401/403/200).
  - [ ] Privilege escalation attempts (e.g., regular user performing admin action) fail predictably.
  - [ ] Results aggregated; failing endpoint names logged for quick triage.
  - [ ] Documentation in `docs/AUTHORIZATION_FRAMEWORK.md` updated with tested endpoints.


### 5) [Testing] Validation Middleware Security Tests

- Labels: `kind/chore`, `area/testing`, `area/backend`, `priority/P0`, `milestone/MVP`
- Summary: Expand validation middleware tests to include SQLi/XSS edge cases and cross-field validation;
  add integration coverage.
- Scope:
  - Middleware: `backend/internal/middleware/validation_middleware.go` (or relevant path)
  - Tests: `backend/internal/middleware/validation_middleware_test.go`, `backend/tests/integration/security/validation_integration_test.go`
- Acceptance Criteria:
  - [ ] Unit tests include malicious inputs (SQLi/XSS payloads); responses sanitize and reject appropriately.
  - [ ] Integration tests verify middleware applied across key endpoints.
  - [ ] Fuzzer smoke test suite executed for 1000+ random payloads without panics.
  - [ ] Document patterns in `docs/SECURITY_TESTING_RUNBOOK.md`.

---

## Phase 2 — Infrastructure Reliability (P0/P1)


### 6) [Testing] Deployment Scripts Test Harness & Smoke Tests

- Labels: `kind/chore`, `area/testing`, `area/infrastructure`, `priority/P0`, `milestone/Beta`
- Summary: Create a test harness to validate deployment scripts (`scripts/*.sh`) with dry-run and sandbox modes.
- Scope:
  - Scripts: `scripts/*.sh` (deploy, rollback, infra ops)
  - Tests: `scripts/tests/deployment_harness_test.sh`, CI job to run harness.
- Acceptance Criteria:
  - [ ] Harness supports DRY_RUN and MOCK mode to simulate external calls.
  - [ ] Smoke test covers success/failure paths for deploy and rollback.
  - [ ] CI job fails on non-zero exit; logs stored as artifact.
  - [ ] Documentation in `docs/deployment/runbook.md` updated.


### 7) [Testing] Database Migration Rollback Tests

- Labels: `kind/chore`, `area/testing`, `area/infrastructure`, `priority/P0`, `milestone/Beta`
- Summary: Add tests to validate forward and rollback migrations with data integrity checks in a shadow database.
- Scope:
  - Migrations: `backend/migrations/*.sql`
  - Tests: `backend/tests/integration/migrations/migration_rollback_test.go`
- Acceptance Criteria:
  - [ ] Apply latest migration and rollback cleanly; no dangling objects.
  - [ ] Integrity checks: referential integrity preserved; indices restored.
  - [ ] Performance: migration run time under threshold (document baseline).
  - [ ] CI integration: fails on drift; report included.


### 8) [Testing] Backup & Restore Validation

- Labels: `kind/chore`, `area/testing`, `area/infrastructure`, `priority/P0`, `milestone/Beta`
- Summary: Automate validation of backup and restore processes with RPO/RTO targets.
- Scope:
  - Scripts: `scripts/backup.sh`, restore runbooks.
  - Tests: `monitoring/tests/backup_restore_validation_test.sh`
- Acceptance Criteria:
  - [ ] Nightly backup completes and is verified.
  - [ ] Monthly restore validation executed; RTO < 1h, RPO < 15m.
  - [ ] Encrypted backups; cross-region storage confirmed.
  - [ ] Alerts configured on backup failures.


### 9) [Testing] Monitoring Alert Rule Validation

- Labels: `kind/chore`, `area/testing`, `area/monitoring`, `priority/P1`, `milestone/Beta`
- Summary: Validate monitoring alert rules and dashboards for accuracy and actionable thresholds.
- Scope:
  - Monitoring config: `monitoring/` (Prometheus, Grafana, etc.)
  - Tests: `monitoring/tests/alert_rules_validation_test.sh`
- Acceptance Criteria:
  - [ ] Synthetic events trigger alerts as expected.
  - [ ] False positive rate documented; thresholds adjusted.
  - [ ] Dashboard panels reflect accurate metrics.
  - [ ] Runbook updated with alert troubleshooting steps.

---

## Phase 3 — Feature Completeness (P1)


### 10) [Testing] Mobile E2E Test Suite — Core Flows

- Labels: `kind/chore`, `area/testing`, `area/mobile`, `priority/P1`, `milestone/Beta`
- Summary: Build Detox E2E coverage for mobile core flows (auth, feed, submission, search, profile, favorites).
- Scope:
  - Mobile: `mobile/app/(tabs)/*`, `mobile/app/auth/*`
  - Tests: `mobile/e2e/*.e2e.ts`
- Acceptance Criteria:
  - [ ] E2E tests cover auth (PKCE, refresh), feed browsing, submission wizard, search, profile edit, favorites.
  - [ ] iOS and Android configurations passing locally and in CI (minimum one platform in CI).
  - [ ] Flaky test rate < 5% over 10 runs.
  - [ ] Playwright-like reporting enabled for Detox.


### 11) [Testing] Discovery Lists — Unit + Integration + E2E Coverage

- Labels: `kind/chore`, `area/testing`, `area/backend`, `priority/P1`, `milestone/Beta`
- Summary: Implement tests for discovery lists (Top/New/Discussed) including pagination, filters, and caching.
- Scope:
  - Backend: `backend/internal/handlers/discovery_list_handler.go`
  - Tests: `backend/internal/handlers/discovery_list_handler_test.go`, `backend/tests/integration/discovery/discovery_integration_test.go`
  - Frontend (optional E2E): `frontend/e2e/discovery.spec.ts`
- Acceptance Criteria:
  - [ ] Unit tests for each list type including edge cases.
  - [ ] Integration tests validate responses with live DB; pagination and filtering behave correctly.
  - [ ] E2E confirms UI displays lists and filters work.
  - [ ] Cache behavior documented; invalidation tested.


### 12) [Testing] Live Status Tracking — Integration Tests

- Labels: `kind/chore`, `area/testing`, `area/backend`, `priority/P1`, `milestone/Beta`
- Summary: Add integration tests for live status tracking service (online/offline transitions, polling, caching).
- Scope:
  - Backend: `backend/internal/services/live_status_service.go`
  - Tests: `backend/tests/integration/live/live_status_integration_test.go`
- Acceptance Criteria:
  - [ ] Status transitions correctly persisted and exposed via API.
  - [ ] Polling/backoff logic verified with mock upstream.
  - [ ] Cache invalidation correct on status change.
  - [ ] Negative tests for upstream failure and timeouts.


### 13) [Testing] Moderation Workflow — E2E Coverage

- Labels: `kind/chore`, `area/testing`, `area/frontend`, `priority/P1`, `milestone/Beta`
- Summary: Add end-to-end tests for moderator workflow (queue, approve/reject, bulk actions, rejection reasons).
- Scope:
  - Frontend: admin pages for moderation
  - Tests: `frontend/e2e/moderation-workflow.spec.ts`
- Acceptance Criteria:
  - [ ] Admin-only access enforced; non-admin blocked.
  - [ ] Approve/reject workflows succeed; bulk actions applied.
  - [ ] Rejection reasons visible to users; audit logs written.
  - [ ] Performance baseline captured (p95 page load < 200ms in local env).


### 14) [Testing] Watch Party Real-time Sync Tests

- Labels: `kind/chore`, `area/testing`, `area/backend`, `priority/P1`, `milestone/Beta`
- Summary: Validate WebSocket sync for watch parties including ±2s tolerance, seek/play/pause consistency, and reconnection.
- Scope:
  - Backend: `backend/internal/services/watch_party_service.go`, WS hub.
  - Tests: `backend/tests/integration/watch_party/watch_party_sync_test.go`
  - Reference: `docs/WATCH_PARTIES_API.md` (events, tolerance, roles)
- Acceptance Criteria:
  - [ ] Multi-client sync tests keep drift within ±2s under normal conditions.
  - [ ] Commands (play/pause/seek/skip) propagate to all participants.
  - [ ] Reconnection recovers state via initial sync.
  - [ ] Role permissions enforced (viewer cannot control playback).

---

## Phase 4 — Performance & Optimization (P2)


### 15) [Testing] Rate Limiting — Load Tests

- Labels: `kind/chore`, `area/testing`, `area/backend`, `priority/P2`, `milestone/GA`
- Summary: Add k6 scenarios to validate rate limiting accuracy under load across key endpoints.
- Scope:
  - k6: `backend/tests/load/scenarios/rate_limit.js`
  - Endpoints: submission, metadata, watch party create/join.
- Acceptance Criteria:
  - [ ] Thresholds defined per endpoint; no false positives at target load.
  - [ ] Error rate < 1%; p95 latency within targets.
  - [ ] Reports exported as CI artifacts.


### 16) [Testing] Search Fallback Performance & Failover Tests

- Labels: `kind/chore`, `area/testing`, `area/backend`, `priority/P2`, `milestone/GA`
- Summary: Test OpenSearch failover behavior and performance impact; validate fallback to cache/local search.
- Scope:
  - Backend: `backend/internal/services/opensearch_search_service.go`
  - Tests: `backend/tests/integration/search/search_failover_test.go`
- Acceptance Criteria:
  - [ ] Failover triggers correctly when upstream unavailable.
  - [ ] Fallback returns results within acceptable latency.
  - [ ] Alerts/logging capture failover events; runbook updated.


### 17) [Testing] CDN Failover Simulation Tests

- Labels: `kind/chore`, `area/testing`, `area/infrastructure`, `priority/P2`, `milestone/GA`
- Summary: Simulate CDN provider failure and validate automatic fallback between providers.
- Scope:
  - Backend: `backend/internal/services/cdn_service.go`
  - Tests: `backend/internal/services/cdn_service_test.go`, integration harness.
- Acceptance Criteria:
  - [ ] Failover from primary to secondary provider operates seamlessly.
  - [ ] Mirror health checks detect and remediate issues.
  - [ ] Orphaned mirrors cleanup verified.
  - [ ] Documentation updated under `docs/deployment/infra.md`.


### 18) [Testing] Webhook Delivery at Scale — Load & DLQ Replay

- Labels: `kind/chore`, `area/testing`, `area/backend`, `priority/P2`, `milestone/GA`
- Summary: Validate outbound webhook delivery at scale with retries, DLQ insertion, and replay success.
- Scope:
  - Backend: `backend/internal/services/outbound_webhook_service.go`, `backend/internal/services/webhook_retry_service.go`
  - Docs: `docs/WEBHOOK_SIGNATURE_VERIFICATION.md`, `docs/WEBHOOK_SUBSCRIPTION_MANAGEMENT.md`
  - Tests: `backend/tests/load/webhooks/webhook_scale_test.go`
- Acceptance Criteria:
  - [ ] Simulate 10k deliveries/hour; ≥ 99% success with retries.
  - [ ] DLQ entries created for hard failures; replay succeeds ≥ 95%.
  - [ ] Monitoring alerts for elevated failure rates; dashboards reflect metrics.
  - [ ] Signature verification remains correct under high throughput.

---

## Filing Instructions

When creating GitHub issues:
- Use the “Feature Request” template and adjust labels to `kind/chore` plus appropriate `area/*`.
- Set Priority and Milestone based on phases above.
- Copy the issue section content verbatim into the issue body.
- Link related docs and code paths for quick triage.
- Add “Closes #<tracker-issue-number>” in PRs implementing tests to sync labels.

After filing, replace “link TBD” in the Master Tracker with the actual issue links and keep the checklist up to date.
