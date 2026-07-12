# Production Readiness Remediation Plan

**Status:** In progress

**Created:** 2026-07-12

**Source audit:** `quality-review.md`

**Branch:** `codex/production-readiness-remediation`

**Release posture:** Blocked until gates G0-G5 pass

## Objective

Resolve every critical and important finding in the production-readiness audit, satisfy the refined user-story acceptance criteria, and produce reproducible evidence that a clean checkout is safe to release.

This plan is ordered to contain risk first, establish trustworthy verification second, repair product behavior third, and harden the release candidate last. A passing build alone is not completion.

## Safety and scope decisions

- Do not deploy or mutate production infrastructure from this worktree.
- Use Stripe test mode only. No live-money or live-trading activity is authorized.
- Keep stream clipping, CDN mirroring, live feed, and watch parties disabled by default until their individual release gates pass.
- Treat the responsive web application as the current launch client. Native mobile remains out of launch scope unless a buildable mobile workspace is supplied and accepted.
- Preserve the pre-existing user changes in `Makefile`, `README.md`, `backend/config/config.go`, `backend/internal/services/embedding_service.go`, and `Taskfile.yml`. When remediation overlaps those files, isolate and review the relevant hunks before committing.
- Never hide a failing or empty test suite. Missing prerequisites, zero discovered tests, unexpected skips, and unhandled test warnings are failures.
- Make one focused commit after each independently validated milestone.

## Evidence model

Each task must end with four forms of evidence:

1. **Implementation:** reviewed diff with no unrelated user changes.
2. **Automated verification:** the commands listed for the task pass.
3. **Behavioral proof:** a focused test or inspection proves the requirement rather than merely exercising nearby code.
4. **Traceability:** the task status and evidence are recorded in this plan.

Task states are `TODO`, `IN PROGRESS`, `BLOCKED`, or `DONE`. A phase is complete only when all tasks and its exit gate are `DONE`.

## Dependency flow

```mermaid
flowchart LR
    A["Phase 0: containment and truth"] --> B["Phase 1: reproducible quality gate"]
    A --> C["Phase 2: secrets and runtime security"]
    B --> D["Phase 3: product and contract correctness"]
    C --> D
    D --> E["Phase 4: UX, performance, and maintainability"]
    E --> F["Phase 5: resilience and release rehearsal"]
```

## Release gates

| Gate | Required evidence |
|---|---|
| G0 — containment | No secret logging; incomplete production paths disabled; public claims identify pre-release scope |
| G1 — reproducible verification | Clean install; backend test/race/vet; frontend lint/build/unit; non-zero Playwright discovery; vulnerability and OpenAPI checks in required CI |
| G2 — security | Fail-closed production configuration; zero unaccepted critical/high vulnerabilities; protected operational endpoints; authorization matrix passes |
| G3 — product correctness | Media, payment, analytics consent, account lifecycle, and advertised feature behavior meet acceptance tests |
| G4 — beta quality | Critical journeys pass accessibility, browser, performance, and API contract budgets |
| G5 — release rehearsal | Immutable artifacts, backup/restore, migration rollback, graceful drain, smoke, and rollback drills pass against the candidate |

## Phase 0 — Containment, baseline, and product truth

### R0.1 — Establish remediation tracking

- **State:** DONE
- **Owner role:** Release lead
- **Work:** Create this plan, map all audit findings, establish gates, and work on a non-default branch.
- **Verification:** Branch is not `main`; this plan covers C1-C4, I1-I11, US-R1/A1/C1/P1/C2/M1/O1/X1/D1, and all hardening recommendations.

### R0.2 — Prevent unsafe feature activation

- **State:** DONE
- **Owner role:** Backend + release engineering
- **Work:** Add centrally validated, false-by-default flags for stream clip extraction, CDN, mirroring, live feed, and watch parties. Do not register mutating/expensive backend routes when disabled. Return a stable `feature_disabled` API error where route compatibility is required.
- **Acceptance:** A production-profile test proves every incomplete feature is disabled without an explicit valid configuration; frontend navigation cannot expose disabled features.
- **Verification:** Focused Go configuration/router tests and frontend route tests.

### R0.3 — Remove immediate secret exposure

- **State:** DONE
- **Owner role:** Backend security
- **Work:** Remove private-key logging immediately. Permit ephemeral JWT keys only in an explicit development profile. Add a regression test that captures logs and proves key material is absent.
- **Acceptance:** Staging/production startup fails if JWT key material is missing; development logs a fingerprint only.
- **Verification:** Focused infrastructure/config tests plus repository secret-pattern scan.

### R0.4 — Publish accurate pre-release scope

- **State:** DONE
- **Owner role:** Product + documentation
- **Work:** Correct the README and feature inventory immediately: web client only; hidden/disabled features marked planned; CI/test counts derived from current evidence; obsolete quick-start links and commands removed or fixed.
- **Acceptance:** No native-mobile, workflow-count, test-count, CDN, mirroring, live-feed, or watch-party completeness claim exceeds executable evidence.
- **Verification:** Documentation checks and clean-checkout command validation.

### Phase 0 exit — G0

- **State:** TODO
- No unsafe fallback or placeholder path can be enabled accidentally.
- Release status and available clients/features are truthful.
- Any JWT key that may have reached non-local logs is identified for operator rotation; rotation itself is an external operational action and must be reported if access is unavailable.

## Phase 1 — Reproducible build, test, and CI gate

### R1.1 — Repair the task-runner contract

- **State:** DONE
- **Owner role:** Developer experience
- **Depends on:** R0.1
- **Work:** Reconcile `Taskfile.yml` and the compatibility `Makefile` with files that actually exist. Remove unsupported targets or implement their assets. Replace silent `|| true` behavior for required setup with explicit failure. Add a side-effect-free target verifier.
- **Acceptance:** Every public task either runs, clearly reports an optional prerequisite, or is removed. Required setup cannot succeed with absent seeds/migrations. Package-specific integration targets validate that the package exists and discovers tests.
- **Verification:** `task --list`, task contract test, and dry-run/summary checks.

### R1.2 — Add mandatory source CI

- **State:** IN PROGRESS — source workflow now enforces backend test/vet/build,
  frontend audit/lint/test/build, non-zero browser discovery, and real-backend
  cross-browser smoke with diagnostic artifacts. Hosted execution and skip-budget
  integration remain before this task is `DONE`.
- **Owner role:** Platform engineering
- **Depends on:** R1.1
- **Work:** Add a pull-request workflow for clean dependency install, backend test/race/vet, frontend lint/build/unit, Playwright discovery/smoke, OpenAPI validation/coverage, docs validation, dependency scanning, secret scanning, SBOM generation, and container build/scan.
- **Acceptance:** Required jobs fail on zero tests, unexpected skips, lint errors, vulnerable high/critical dependencies without an unexpired exception, or missing artifacts.
- **Artifacts:** Test counts, coverage, skip inventory, bundle sizes, SBOM, migration version, OpenAPI coverage, and image digest.

### R1.3 — Resolve frontend lint errors

- **State:** DONE
- **Owner role:** Frontend
- **Depends on:** R1.2 can run in parallel
- **Work:** Correct all 13 errors, including set-state-in-effect, impure render, render-time ref access, and invalid Vite configuration. Triage warnings and establish a warning budget that trends to zero.
- **Acceptance:** `npm run lint` exits zero; no disable comment is added without a precise rationale and focused test.
- **Verification:** Frontend lint and affected unit tests.

### R1.4 — Repair frontend unit/component tests

- **State:** DONE
- **Owner role:** Frontend + QA
- **Depends on:** R1.3
- **Work:** Classify all 69 failures as implementation defects or stale tests. Repair analytics configuration injection, provider-aware render helpers, modal state restoration, accessibility primitives, and order-dependent search-state tests. Replace exact Tailwind assertions with accessible/behavioral outcomes.
- **Acceptance:** Two consecutive full runs pass with identical totals; React `act`, unhandled-request, canvas, and navigation warnings are either eliminated or intentionally provided by shared test adapters.
- **Verification:** `npm test -- --run` twice, with machine-readable result comparison.

### R1.5 — Restore Playwright discovery and real-backend smoke coverage

- **State:** IN PROGRESS — maintained mocked and real-backend tiers discover tests;
  Chromium/Firefox real-stack smoke and Chromium mocked smoke pass locally. CI
  must prove WebKit on a supported runner, and the remaining critical journeys
  below still require fixtures and coverage.
- **Owner role:** QA automation
- **Depends on:** R1.1, R1.4
- **Work:** Restore fixture/page/helper modules or rewrite specs around maintained fixtures. Separate mocked UI tests from real-backend smoke tests. Add non-zero discovery enforcement.
- **Critical real-backend journeys:** Login/OAuth boundary, browse/search, clip detail/playback, submission, premium checkout initiation/return in Stripe test mode, settings/export/deletion, and moderator authorization denial/allow.
- **Acceptance:** `playwright test --list` discovers the expected manifest; supported-browser smoke passes against repository-owned services; mocked suites cannot satisfy the real-backend gate.

### R1.6 — Make skip and warning debt visible

- **State:** DONE — 80 active Go skips and the quarantined legacy browser skip
  have owners, reasons, per-file ratchet limits, and 2026-08-15 expiries. CI
  rejects new/unregistered/expired skips, and release-critical backend suites
  pass with zero runtime skips against repository-owned services.
- **Owner role:** QA + backend
- **Depends on:** R1.2
- **Work:** Inventory the 80 active Go skips and all frontend skipped tests. Categorize by environment, obsolete behavior, or missing coverage. Add CI output and a ratcheting budget; release-critical security/payment/migration tests may not skip.
- **Acceptance:** Every skip has owner/reason/expiry or is removed; the gate fails if the count increases or a release-critical test skips.

### Phase 1 exit — G1

- **State:** TODO
- Clean-checkout verification is executable locally and in CI.
- All mandatory suites pass, discover tests, and emit required evidence.

## Phase 2 — Secrets, dependencies, authorization, and runtime security

### R2.1 — Implement coherent environment profiles

- **State:** DONE — `Config.Validate()` rejects unknown/incoherent profiles and
  release profiles fail closed on insecure URLs/origins, debug mode, default or
  missing secrets, disabled database/OpenSearch verification, local storage,
  insecure telemetry, incomplete paid-feature configuration, and contained
  launch features. Table-driven config tests plus full backend test/vet pass.
- **Owner role:** Backend
- **Depends on:** R0.3
- **Work:** Add `Config.Validate()` and explicit development/staging/production profiles. Unify `ENVIRONMENT` and Gin release behavior. Validate URLs, origins, JWT/MFA/Twitch/Stripe/storage/telemetry settings based on enabled features. Require secure cookies, HTTPS, verified OpenSearch TLS, and non-local durable storage in production.
- **Acceptance:** Table-driven tests cover valid/invalid profiles and prove production fails closed.
- **Verification:** Config unit tests and API startup smoke per profile.

### R2.2 — Upgrade vulnerable dependencies and automate policy

- **State:** TODO
- **Owner role:** Frontend + security
- **Depends on:** R1.4
- **Work:** Upgrade Axios, React Router, DOMPurify, form-data/picomatch dependency paths, and any new findings. Install/run `govulncheck`; add `gosec`, secret scanning, dependency review, SBOM, image scanning, and signed-artifact support.
- **Acceptance:** Zero unaccepted critical/high production vulnerabilities. Any residual advisory has applicability analysis, owner, compensating control, and expiry.
- **Verification:** `npm audit --omit=dev`, `govulncheck ./...`, scanners, full frontend/backend tests.

### R2.3 — Restrict operational endpoints

- **State:** TODO
- **Owner role:** Backend + platform
- **Depends on:** R2.1
- **Work:** Keep minimal liveness/readiness public. Move Prometheus, DB pool, cache, and webhook retry details to an internal listener or protect them with service authentication and network policy. Add endpoint-specific limits.
- **Acceptance:** Anonymous requests cannot read internal metrics/stats; trusted scrape identity can; public health contains no capacity/failure details.
- **Verification:** Router/middleware tests and deployment network-policy validation.

### R2.4 — Complete object-level authorization matrix

- **State:** TODO
- **Owner role:** Backend security
- **Depends on:** R1.6
- **Work:** Enumerate every `/:id` read/mutation and test owner, non-owner, moderator, admin, anonymous, deleted/banned, and cross-tenant/resource cases. Restore IDOR/security suites referenced by tasks.
- **Acceptance:** Default deny is demonstrated for every production resource mutation; moderator/admin elevation is explicit and audited.
- **Verification:** Generated route/authorization matrix and integration tests.

### R2.5 — Harden CSP, client storage, and log redaction

- **State:** TODO
- **Owner role:** Security + frontend
- **Depends on:** R2.1
- **Work:** Replace incompatible/report-only CSP with a tested policy for Twitch/media/analytics, remove avoidable unsafe directives, and enforce after report review. Stop describing session-storage ciphertext with a co-located key as an XSS security boundary. Redact authorization, cookies, OAuth codes, key material, and PII webhook content.
- **Acceptance:** Browser CSP tests pass required embeds while rejecting unapproved origins; structured logging tests prove sensitive fields are absent.

### R2.6 — Harden HTTP and container runtime

- **State:** TODO
- **Owner role:** Backend + platform
- **Depends on:** R2.1
- **Work:** Add bounded read/write/idle/header limits with explicit streaming/WebSocket handling and graceful-drain tests. Run containers as non-root, pin bases/digests, set read-only roots/tmpfs, drop capabilities, enable `no-new-privileges`, add resource limits, health checks, and immutable release image references.
- **Acceptance:** Slow-client and shutdown tests pass; images run without root/write access; deployment manifests use immutable references and constrained security contexts.

### Phase 2 exit — G2

- **State:** TODO
- Production starts only with a valid secure profile.
- Operational data and resource mutations are access-controlled.
- Dependency/scanner policy passes with documented exceptions only.

## Phase 3 — Product, media, API, payment, and privacy correctness

### R3.1 — Implement or formally disable stream clip creation

- **State:** TODO
- **Owner role:** Media/backend
- **Depends on:** R0.2, R2.1
- **Work:** Replace placeholder VOD/media URLs with real source resolution and a durable idempotent job state machine (`queued`, `processing`, `ready`, `failed`, `retrying`, `cancelled`), or keep the feature entirely disabled for launch. Enqueue and record creation must be transactional or compensating.
- **Acceptance:** No normal success can create a permanent placeholder or stranded processing record. Duplicate request, enqueue failure, retry exhaustion, cancellation, and user-visible recovery are tested.

### R3.2 — Implement or formally disable CDN and mirroring

- **State:** TODO
- **Owner role:** Media/platform
- **Depends on:** R0.2, R2.1
- **Work:** Replace fabricated URLs, zero metrics, no-op purge, and placeholder distribution/host values with provider API calls, object verification, explicit unsupported errors, and cost/health telemetry—or remove providers from launch registration.
- **Acceptance:** Provider initialization rejects incomplete config; success proves the object exists; purge/metrics failures propagate; no `clpr.cdn` or placeholder distribution data remains in enabled paths.

### R3.3 — Define live feed and watch-party launch contract

- **State:** TODO
- **Owner role:** Product + frontend/backend
- **Depends on:** R0.2, R0.4
- **Work:** Keep both out of launch scope by default and gate frontend routes plus backend route registration from one feature source. If promoted later, require dedicated privacy, authorization, rate-limit, WebSocket resilience, and E2E acceptance evidence.
- **Acceptance:** Feature inventory, server flags, API exposure, and navigation agree in every environment.

### R3.4 — Verify subscription lifecycle in Stripe test mode

- **State:** TODO
- **Owner role:** Payments/backend/frontend
- **Depends on:** R1.5, R2.1
- **Work:** Add contract/integration coverage for checkout creation, signed webhooks, duplicates, out-of-order delivery, invalid signatures, cancellation, dunning/payment failure, reconciliation, and entitlement activation/revocation. Browser smoke covers initiation and return without real charges.
- **Acceptance:** UI never grants/claims premium before confirmed backend entitlement; webhook processing is idempotent and replay-safe.

### R3.5 — Make analytics consent testable and enforce withdrawal

- **State:** TODO
- **Owner role:** Frontend + privacy
- **Depends on:** R1.4
- **Work:** Inject analytics configuration through a runtime adapter instead of mocking `import.meta`. Test consent, DNT, reload, withdrawal, backend reconciliation, and PII minimization.
- **Acceptance:** No analytics request before consent/DNT; no further request after withdrawal; consent storage and backend record converge.

### R3.6 — Cover account export/deletion and moderation lifecycle

- **State:** TODO
- **Owner role:** Backend/frontend + privacy
- **Depends on:** R1.5, R2.4
- **Work:** Add integration and browser evidence for export completeness, deletion confirmation/grace behavior, session revocation, retained/legal data policy, moderator authorization, and destructive-action audit trails.
- **Acceptance:** User-visible state, stored state, and documented policy agree after each lifecycle transition.

### R3.7 — Bring OpenAPI coverage to supported-route parity

- **State:** TODO
- **Owner role:** API/backend
- **Depends on:** R2.4, R3.1-R3.6 scope decisions
- **Work:** Generate a route-to-operation coverage manifest. Document every supported production route, cookie/CSRF and service auth, schemas, errors, and status codes. Exclude disabled routes explicitly rather than silently.
- **Acceptance:** CI reports 100% supported-route coverage and contract tests compare live handlers to the spec.

### Phase 3 exit — G3

- **State:** TODO
- No enabled path returns placeholder media, URLs, metrics, or fake success.
- Payment/privacy/account/moderation critical journeys have real integration evidence.
- API and feature inventory match runtime behavior.

## Phase 4 — Accessibility, performance, maintainability, and beta quality

### R4.1 — Repair shared accessibility primitives

- **State:** TODO
- **Owner role:** Frontend design systems
- **Depends on:** R1.4
- **Work:** Enforce 44 px touch targets, associated labels/errors, visible `focus-visible`, reduced motion, explicit transitions, intrinsic image dimensions/aspect ratios, focus trap/restoration, prior scroll-state restoration, and overscroll containment in shared components.
- **Targets:** Button, Modal, PaywallModal, OptimizedImage, ClipGridCard, ThreadDetail forms, SearchResultCard.
- **Acceptance:** Shared primitive tests assert behavior/semantics, not exact utility strings.

### R4.2 — Validate accessible critical journeys

- **State:** TODO
- **Owner role:** QA + accessibility
- **Depends on:** R1.5, R4.1
- **Work:** Run axe plus manual keyboard, screen-reader, zoom/reflow, reduced-motion, and touch checks on login, search, clip detail, submission, payment, settings, moderation, and destructive confirmations.
- **Acceptance:** No critical/serious axe violations; manual checklist has evidence and ownership for approved lower-severity debt.

### R4.3 — Enforce frontend performance budgets

- **State:** TODO
- **Owner role:** Frontend performance
- **Depends on:** R1.4
- **Work:** Generate bundle analysis, remove broad entry barrels, fix empty/manual chunks and mixed imports, route-split expensive features, and set entry/route gzip and browser performance budgets in CI.
- **Acceptance:** No raised warning threshold as a substitute for optimization; budgets are based on measured critical journeys and fail CI on regression.

### R4.4 — Establish maintainable module and UI boundaries

- **State:** TODO
- **Owner role:** Frontend/backend architecture
- **Depends on:** R3.7, R4.1
- **Work:** Consolidate modal/button/input/date/number/API-state primitives, prevent broad barrel imports on entry paths, split the route surface into versioned bounded modules, and add dependency/layering checks.
- **Acceptance:** Boundary rules are executable; duplicate patterns targeted by the audit are removed or documented as intentional.

### R4.5 — Build the intended test pyramid

- **State:** TODO
- **Owner role:** QA architecture
- **Depends on:** R1.4-R1.6, R3.7
- **Work:** Keep pure logic in unit tests, accessible behavior in component tests, schemas in contract tests, PostgreSQL/Redis/OpenSearch behavior in containerized integration tests, and only critical user paths in real-backend browser tests.
- **Coverage areas:** Auth/OAuth/CSRF/abuse, submissions, engagement, search fallback, repositories, notifications, migrations, payments, clips, and security/IDOR.
- **Acceptance:** Advertised suites exist and discover meaningful tests; skip budget and coverage reports reflect actual tiers.

### Phase 4 exit — G4

- **State:** TODO
- Critical journeys meet accessibility and performance budgets.
- Test and module boundaries are enforceable and documented.

## Phase 5 — Reliability, observability, documentation, and release rehearsal

### R5.1 — Add resilience controls

- **State:** TODO
- **Owner role:** Backend reliability
- **Depends on:** R3.1-R3.6
- **Work:** Add job idempotency, bounded queues, dead-letter behavior, retry jitter, database statement timeouts, feature-aware readiness/degraded states, migration compatibility policy, and restore-point objectives.
- **Acceptance:** Dependency failure does not create false success or unbounded resource use; required dependencies fail readiness and optional dependencies report degradation.

### R5.2 — Define SLOs and actionable observability

- **State:** TODO
- **Owner role:** SRE + product
- **Depends on:** R2.3, R5.1
- **Work:** Define SLOs for auth, feed, search, playback, submission, checkout, and moderation. Add structured error codes, trace IDs in user-safe errors, RED metrics, job lag/failure metrics, alert thresholds, runbooks, and owners.
- **Acceptance:** Every release-critical alert links to a tested runbook and avoids sensitive log data.

### R5.3 — Run dependency and chaos cases

- **State:** TODO
- **Owner role:** QA reliability
- **Depends on:** R5.1, R5.2
- **Work:** Exercise Redis, OpenSearch, Twitch, Stripe test mode, SendGrid, storage/CDN, and database degradation. Test webhook replay, queue recovery, and partial dependency startup.
- **Acceptance:** Expected degraded/readiness behavior and recovery time are recorded; no data corruption or false success occurs.

### R5.4 — Restore load, stress, and soak suites

- **State:** TODO
- **Owner role:** Performance engineering
- **Depends on:** R1.1, R5.1
- **Work:** Implement or remove every advertised k6 target. Create repository-owned realistic fixtures and thresholds for feed, clip detail, search, comments, auth, submission, rate limiting, moderation, stress, and a practical pre-release soak.
- **Acceptance:** Release thresholds and capacity assumptions are versioned; failures produce diagnostic artifacts.

### R5.5 — Validate migration, backup, restore, and rollback

- **State:** TODO
- **Owner role:** Database/platform
- **Depends on:** R1.2, R5.1
- **Work:** Run up/down compatibility, backup validation, restore to isolated environment, post-restore application smoke, immutable canary deployment, graceful drain, and automated rollback using release telemetry.
- **Acceptance:** RPO/RTO evidence, restored data validation, migration version, image digest, and rollback result are attached to the candidate.

### R5.6 — Finish documentation and mobile disposition

- **State:** TODO
- **Owner role:** Product/docs
- **Depends on:** All prior scope decisions
- **Work:** Supply safe `.env.example` files and a tested clean-clone setup; repair every README/doc link; generate feature/workflow/test inventories where practical; align legal/privacy/API/deployment/runbooks. Record native mobile as planned or introduce a separately owned, buildable, tested workspace.
- **Acceptance:** US-D1 and US-M1 pass; documentation CI finds no broken internal links or unsubstantiated release claims.

### R5.7 — Final completion audit and release decision

- **State:** TODO
- **Owner role:** Release lead + security + product
- **Depends on:** R5.1-R5.6
- **Work:** Re-run every command and acceptance criterion from a clean checkout, inspect evidence rather than relying on prior intent, and complete the go/no-go checklist.
- **Acceptance:** G0-G5 all pass; every audit item maps to a `DONE` task with authoritative evidence; no required work remains.

## Finding-to-task traceability

| Audit item | Primary remediation tasks |
|---|---|
| C1 — red frontend/E2E gate | R1.2-R1.6 |
| C2 — JWT generation/logging | R0.3, R2.1, R2.5 |
| C3 — placeholder stream clips | R0.2, R3.1, R5.1 |
| C4 — vulnerable dependencies | R2.2 |
| I1 — nonexistent task assets | R1.1, R4.5, R5.4 |
| I2 — overstated docs/features | R0.4, R5.6 |
| I3 — unreachable live/watch-party UI | R0.2, R3.3 |
| I4 — fabricated CDN/mirror success | R0.2, R3.2 |
| I5 — incoherent production config | R2.1 |
| I6 — public operational internals | R2.3 |
| I7 — HTTP/container hardening | R2.6, R5.5 |
| I8 — oversized bundle | R4.3, R4.4 |
| I9 — incomplete OpenAPI | R3.7 |
| I10 — accessibility violations | R4.1, R4.2 |
| I11 — stale tests/missing behavior | R1.4-R1.6, R4.5 |
| US-R1 — reproducible release gate | R1.1-R1.6, R5.7 |
| US-A1 — fail-closed secrets | R0.3, R2.1 |
| US-C1 — reliable clip creation | R3.1, R5.1 |
| US-P1 — subscription lifecycle | R3.4 |
| US-C2 — analytics consent | R3.5 |
| US-M1 — honest mobile availability | R0.4, R5.6 |
| US-O1 — private telemetry | R2.3, R5.2 |
| US-X1 — accessible journeys | R4.1, R4.2 |
| US-D1 — working setup | R1.1, R5.6 |
| Stability hardening | R2.6, R5.1, R5.3-R5.5 |
| Security/privacy hardening | R2.1-R2.6, R3.4-R3.6 |
| Maintainability | R3.7, R4.4, R4.5 |
| Observability | R2.3, R5.2, R5.3 |

## Commit strategy

Use focused conventional-style commits, for example:

1. `docs: add production readiness remediation plan`
2. `fix(security): fail closed on jwt configuration`
3. `chore(ci): add mandatory source quality gate`
4. `fix(test): restore frontend unit and e2e gates`
5. `fix(media): disable incomplete production providers`

Before each commit:

- inspect `git diff` and `git diff --cached`;
- stage only milestone-owned paths or hunks;
- run the milestone verification commands;
- record evidence below;
- do not include unrelated pre-existing user changes.

## Execution log

| Date | Milestone | Result | Evidence/commit |
|---|---|---|---|
| 2026-07-12 | R0.1 plan and branch | Done | Branch created; `git diff --check` and Markdown lint pass |
| 2026-07-12 | R0.2-R0.3 feature/key containment | Done | Focused route/key tests, full `go test ./...`, and `go vet ./...` pass |
| 2026-07-12 | R0.4 truthful release scope | Done | README/inventory corrected; Markdown lint and `git diff --check` pass |
| 2026-07-12 | R1.1 task-runner contract | Done | `task contract`, list/summary/dry-run checks, Make compatibility help, and whitespace checks pass |
| 2026-07-12 | R1.3-R1.4 frontend quality gate | Done | Lint has zero warnings; build passes; two full runs each pass 1,644/1,644 tests |
