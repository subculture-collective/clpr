# CLPR Production Release Readiness Report

**Assessment date:** 2026-08-08 (America/Chicago)
**Repository:** `clpr`
**Candidate:** `codex/production-readiness-remediation` at `b2156caa9b7`
**Production target:** `https://clpr.tv`
**Decision:** **NO-GO**

## Executive summary

CLPR should not be promoted to production in its current state. The source candidate contains a substantial and thoughtful hardening program, and its core local tests, lint, contract checks, and builds pass. Production, however, appears to be running an older June build that does not contain those remediations. The live `/search` route crashes for an anonymous visitor, the candidate's Go dependency scan reports three reachable vulnerabilities, and all six fail-closed operator evidence artifacts required by the repository are missing.

The release is viable after targeted remediation. The shortest safe path is to fix the search crash and dependency findings, produce an immutable candidate from a clean tree, run the complete hosted and operator gates against that exact SHA and image digest, deploy it through the repository's hardened blue/green path, and verify the live site before promotion.

## Readiness scorecard

| Area | Status | Evidence |
| --- | --- | --- |
| Release governance | Blocked | The repository itself marks the plan “Blocked until gates G0-G5 pass”; six required evidence files are absent. |
| Build and static quality | Pass with caveats | Backend test/vet/build, frontend lint/unit/build, task contract, UI boundaries, documentation, and bundle budgets pass locally. |
| Dependency and source security | Blocked | `govulncheck` finds three reachable vulnerabilities; the exact `gosec` CI command also exits non-zero. |
| Functional production health | Blocked | Public home and legal pages render, but `/search` reaches the global error boundary. |
| Production artifact parity | Blocked | Live HTML and containers predate the candidate; current hardening and bundle improvements are not deployed. |
| Runtime hardening | Blocked | Inspected live application containers run without the candidate's non-root/read-only/capability/resource controls. |
| Accessibility | Blocked | Live home has 40 serious color-contrast nodes and one serious link-identification issue in an automated axe pass. |
| Performance | Partially ready | Live health is responsive and the candidate meets source bundle budgets; required staging load/stress/soak evidence is missing. |
| Reliability and recovery | Blocked | Production restore, RPO/RTO, canary, drain, rollback, and post-rollback smoke evidence is missing. |
| Payments and authenticated journeys | Blocked | Test-mode Stripe and provider-backed authenticated journey evidence is missing. |
| TLS and edge delivery | Pass | Valid certificate, HTTPS, HTTP/2, Cloudflare delivery, and a healthy API response were observed. |

## Scope and method

This review covered the repository structure, build and deployment definitions, release plan and evidence gates, automated test inventory, local verification commands, dependency/security scanners, production HTTP behavior, browser behavior at desktop and mobile widths, accessibility automation, TLS, and the running production container configuration.

The production review was non-destructive. No authenticated test identity or Stripe test-mode credential was available, so login/OAuth, account lifecycle, moderation, submission, playback entitlements, and checkout were not exercised end to end. No live-money action was attempted. Load, restore, and rollback operations were not run against production.

The working tree was not clean during the audit: `backend/config/config.go` and `backend/internal/services/embedding_service.go` were modified, and `quality-review.json` was untracked. These changes were preserved and were not created by this review.

## Strengths

- The candidate has broad test depth: 133 frontend unit/component files with 1,669 passing tests, 217 backend test files, and 32 discovered Playwright tests across the configured projects.
- Backend tests, vet, and build pass; frontend lint and production build pass.
- The candidate's initial application JavaScript is about 495 kB raw, versus about 1.26 MB for the live initial application asset. The repository's bundle budget passes.
- Source deployment controls are materially stronger than production: read-only filesystems, dropped capabilities, `no-new-privileges`, PID/memory/CPU limits, and non-root image users are defined.
- The current Caddy policy defines HSTS, clickjacking and MIME protections, referrer controls, site isolation headers, and a substantially stricter CSP.
- The release evidence gate is fail-closed and binds evidence to a candidate SHA, owner, execution time, and evidence URL.
- OpenAPI route coverage, documentation integrity, internal links, UI module boundaries, and asset checks are automated.
- The live API health endpoint returned `200 {"status":"healthy"}` in approximately 181 ms, and the public home page rendered responsively without horizontal overflow.
- TLS was valid for `clpr.tv` and `*.clpr.tv`, with a Google Trust Services certificate valid from 2026-07-21 through 2026-10-19.

## Critical issues

### C1 — Production is not running the release candidate

**Location:** production deployment; source controls at `docker-compose.yml:36-44`, `docker-compose.yml:61-68`, `docker-compose.yml:82-89`, `backend/Dockerfile:22-38`, and `frontend/Dockerfile:31-43`.

**What is wrong:** Live HTML reports a 2026-06-07 last-modified date and loads `/assets/app-DSS5KtSQ.js` at 1,258,500 bytes. The candidate produces an initial application asset of about 494.72 kB. Inspected live application images were created in June, use mutable `latest` tags, and do not expose a revision label that binds them to the audited SHA. This strongly indicates that production corresponds to the older pre-remediation line, although an exact deployed commit cannot be proven from the current image metadata.

Live frontend, backend, and crawler containers also had an empty configured user, writable root filesystems, no dropped capabilities, no `no-new-privileges`, and no configured CPU or memory limits. Those controls exist in the candidate but are not active in production.

**Why it matters:** A source-only security fix provides no production protection. Mutable, unattributed images also make rollback, incident attribution, and reproducible release approval unreliable.

**Minimum fix:** Build clean candidate images once; scan and sign them; record SHA and digest; deploy only digest-pinned images through the hardened blue/green definition; verify non-root UID, read-only root, dropped capabilities, security options, and resource limits from the running containers; then attach the digest and inspection output to release evidence.

### C2 — The live search route crashes

**Location:** `https://clpr.tv/search`; live JavaScript chunk `chunk-BuSsURm8.js`.

**What is wrong:** Opening `/search` as an anonymous desktop visitor reaches the global “Something went wrong” boundary. The page error is `TypeError: Cannot read properties of null (reading 'length')` at the deployed chunk's search code.

**Why it matters:** Search is a primary public journey. A deterministic crash is release-blocking and also indicates a missing production-contract test for nullable API data.

**Minimum fix:** Reproduce against the production API response, normalize nullable collections at the API boundary, add a focused regression test and real-backend browser test for empty and populated searches, then verify the fixed digest on desktop and mobile before promotion.

### C3 — Three reachable Go dependency vulnerabilities fail the candidate security gate

**Location:** `backend/go.mod` dependency graph and reachable call paths reported by `govulncheck`.

**What is wrong:** The exact source CI scan reports:

- `GO-2026-6061`: `google.golang.org/grpc` 1.79.3, fixed in 1.82.1; affected code is reachable.
- `GO-2026-5970`: `golang.org/x/text` 0.37.0, fixed in 0.39.0; affected invalid UTF-8 handling is reachable.
- `GO-2026-5158`: `go.opentelemetry.io/otel` 1.43.0, fixed in 1.44.0; affected baggage-header processing is reachable through HTTP middleware.

**Why it matters:** These include authentication-bypass/denial-of-service and unbounded-processing risks on reachable paths. The repository's release policy explicitly rejects vulnerable high/critical dependencies without an approved, unexpired exception.

**Minimum fix:** Upgrade to fixed compatible versions, run `go mod tidy`, repeat all backend tests/vet/build and `govulncheck`, and record the successful scan and SBOM against the candidate SHA. If compatibility prevents an immediate upgrade, document a time-bounded exception with exposure analysis and a compensating control; do not silently waive the gate.

### C4 — Required operator and hosted release evidence is absent

**Location:** `scripts/verify-release-evidence.js:12-57`; release plan at `docs/superpowers/plans/2026-07-12-production-readiness-remediation.md:118-126`, `149-161`, `176-180`, `300-304`, `332-336`, and `417-447`.

**What is wrong:** `node scripts/verify-release-evidence.js` fails because all required artifacts are missing:

- `release-evidence/hosted-ci.json`
- `release-evidence/security-operations.json`
- `release-evidence/stripe-test-mode.json`
- `release-evidence/load-and-soak.json`
- `release-evidence/restore.json`
- `release-evidence/rollback.json`

The plan separately records hosted CI/branch protection/WebKit, JWT exposure and rotation review, provider identity, Stripe test mode, staging performance, production restore, and canary/rollback work as blocked.

**Why it matters:** These are not documentation niceties; they cover failure modes that local unit tests cannot establish. Bypassing them would contradict the repository's explicit release contract.

**Minimum fix:** Execute every gate against the same immutable SHA and image digest, store authoritative evidence with the required metadata, and require `verify-release-evidence.js` to pass in the promotion job.

## Important issues

### I1 — Live response hardening is inconsistent and stale

**Location:** live `/` and `/api/v1/*`; candidate policy at `Caddyfile:26-46`.

The live HTML shell did not include CSP, HSTS, `X-Frame-Options`, `X-Content-Type-Options`, `Referrer-Policy`, or `Permissions-Policy` in the observed response. API responses included several protections, but the live API CSP still allowed `script-src 'unsafe-inline' 'unsafe-eval'`. The candidate Caddy policy removes those script exceptions and supplies stronger headers.

**Fix:** Deploy one edge policy for HTML and API responses, test it with automated assertions, and begin CSP enforcement after confirming required providers. Add `Permissions-Policy` explicitly if browser features should be restricted.

### I2 — Live accessibility has serious contrast defects

**Location:** live home page.

An axe scan at desktop width reported 40 serious `color-contrast` nodes and one serious `link-in-text-block` issue. Automated checks do not replace keyboard and screen-reader review, but this result is too broad to treat as isolated noise.

**Fix:** Correct shared tokens and affected link styles, rerun axe at supported breakpoints, and complete keyboard/focus/name/role/value checks on release-critical journeys. Attach zero-serious/critical results to the deployed candidate.

### I3 — Static metadata and unknown routes are served as the SPA shell

**Location:** live `/robots.txt`, `/sitemap.xml`, `/.well-known/security.txt`, `/manifest.webmanifest`, `/ready`, `/health`, and unknown paths; candidate routing around `Caddyfile:56-85`.

The tested paths returned `200 text/html` with the application shell rather than the requested file or a correct error. `/manifest.json` and `/sw.js` did work. This creates soft 404s, invalid discovery metadata, and misleading health checks.

**Fix:** Add explicit handlers and content types for supported SEO/PWA/security files, return actual readiness/health payloads only on canonical endpoints, and return a real 404 where no route exists. Add HTTP contract tests at the edge layer.

### I4 — Authenticated, provider, payment, and cross-browser evidence is incomplete

**Location:** release plan `R1.5`, `R3.4`, and `G3`.

Playwright discovery finds 32 tests, but discovery is not proof of a live journey. The plan records missing hosted WebKit and maintained fixtures for authenticated OAuth, submission, export/deletion, moderator allow/deny, playback, search, and Stripe test-mode checkout.

**Fix:** Run real-backend journeys with disposable identities and provider test fixtures on supported hosted browsers. Keep mocked UI tests separate from the real-backend promotion gate.

### I5 — Capacity and recovery objectives are unproven

**Location:** release plan `R5.4-R5.5`.

The repository contains k6 profiles and backup/restore tooling, but required staging execution, production-class restore validation, measured RPO/RTO, canary, graceful drain, rollback, and post-rollback smoke evidence are missing.

**Fix:** Establish explicit SLO/capacity thresholds, seed a disposable staging environment, run baseline/stress/soak, restore a production-like backup into isolation, and exercise canary plus rollback using the exact candidate digest.

### I6 — The exact `gosec` CI invocation fails on a non-security use of `math/rand`

**Location:** `backend/internal/services/webhook_retry_service.go:198`.

The scanner emits high-severity `G404` for retry scheduling jitter even though the value is not used for a token, identity, authorization, or other security decision. The existing `//nolint:gosec` comment is not recognized by `gosec`, so the mandatory command exits non-zero.

**Fix:** Use a scanner-recognized, narrowly scoped `#nosec G404` annotation with the existing rationale, or switch to a concurrency-safe cryptographic random source. Keep the scan fail-closed and verify the exception in CI.

### I7 — API lint passes with 84 warnings

**Location:** `docs/openapi/openapi.yaml`.

Redocly exits successfully but emits 84 warnings, principally ambiguous route shapes and invalid examples. Route coverage is strong, but a warning-heavy contract can obscure regressions and may produce surprising client routing or generation behavior.

**Fix:** Triage every warning, repair invalid examples and ambiguous paths, establish an explicit warning budget, and make new warnings fail CI.

### I8 — The candidate is not yet an immutable release input

**Location:** repository working tree.

The audited branch is 186 commits ahead of `main`, with two modified tracked files and one untracked review artifact. Local results therefore do not yet describe a clean, reproducible commit.

**Fix:** Deliberately resolve or preserve the unrelated changes, create the final candidate commit, verify a clean checkout, and bind every artifact and operator result to its SHA and image digest.

## Minor issues

- Anonymous page loads produce visible browser-console 401 failures for session/refresh and some user-specific endpoints. Expected anonymous checks should avoid error-level noise so real authentication failures remain actionable.
- The frontend build reports Browserslist data about eight months old. Refresh it in a controlled dependency update and repeat cross-browser checks.
- The live root references a nested favicon path while `/favicon.ico` returns 404. Provide a conventional root favicon or redirect it.
- Frontend unit tests pass, but the run emits substantial expected console output and React `act` warnings. The plan's stated warning-acceptance criterion should be enforced machine-readably rather than relying only on exit status.

## Security findings appendix

### SEC-1 — Reachable gRPC vulnerability

- **Rule ID:** `GO-2026-6061`
- **Severity:** High
- **Location:** `google.golang.org/grpc` 1.79.3 in the backend dependency graph
- **Evidence:** `govulncheck` reports reachable affected traces; fixed version is 1.82.1.
- **Impact:** Authentication bypass and denial-of-service conditions may be reachable depending on exercised gRPC behavior.
- **Fix:** Upgrade to 1.82.1 or later compatible release and retest.
- **Mitigation:** Restrict exposed gRPC paths and traffic limits only as a temporary compensating control.
- **False-positive assessment:** Not treated as false positive because vulnerable symbols are reachable.
- **Advisories:** <https://pkg.go.dev/vuln/GO-2026-6061>, <https://github.com/grpc/grpc-go/security/advisories/GHSA-hrxh-6v49-42gf>

### SEC-2 — Reachable invalid UTF-8 infinite loop

- **Rule ID:** `GO-2026-5970`
- **Severity:** High availability risk
- **Location:** `golang.org/x/text` 0.37.0 in the backend dependency graph
- **Evidence:** `govulncheck` reports reachable affected traces; fixed version is 0.39.0.
- **Impact:** Crafted invalid UTF-8 input can cause unbounded processing and denial of service in an affected path.
- **Fix:** Upgrade to 0.39.0 or later compatible release and retest.
- **Mitigation:** Bound request size and processing time while the upgrade is prepared.
- **False-positive assessment:** Not treated as false positive because vulnerable symbols are reachable.
- **Advisory:** <https://pkg.go.dev/vuln/GO-2026-5970>

### SEC-3 — Reachable OpenTelemetry baggage-header denial of service

- **Rule ID:** `GO-2026-5158`
- **Severity:** Moderate
- **Location:** `go.opentelemetry.io/otel` 1.43.0 in the backend dependency graph
- **Evidence:** `govulncheck` reports a reachable HTTP middleware path; fixed version is 1.44.0.
- **Impact:** Oversized baggage headers can consume excessive resources and degrade availability.
- **Fix:** Upgrade to 1.44.0 or later compatible release and retest.
- **Mitigation:** Enforce conservative request-header limits at the edge and application server.
- **False-positive assessment:** Not treated as false positive because the middleware path is reachable.
- **Advisories:** <https://pkg.go.dev/vuln/GO-2026-5158>, <https://github.com/open-telemetry/opentelemetry-go/security/advisories/GHSA-5wrp-cwcj-q835>

### SEC-4 — Production container isolation is below the candidate baseline

- **Rule ID:** `RUNTIME-CONTAINER-001`
- **Severity:** High
- **Location:** live frontend, backend, and crawler containers
- **Evidence:** Empty configured user, writable root filesystem, no capability drop, no `no-new-privileges`, and no CPU/memory limits were observed.
- **Impact:** A compromised process has a broader ability to modify its container and consume host resources; blast radius is larger than intended.
- **Fix:** Deploy the candidate hardening, enforce a non-root UID, and verify effective runtime configuration.
- **Mitigation:** Host-level isolation and edge filtering may reduce exposure but do not replace container controls.
- **False-positive assessment:** This is an effective-runtime observation, not a source inference.

### SEC-5 — HTML security headers are missing in production

- **Rule ID:** `WEB-HEADERS-001`
- **Severity:** Medium
- **Location:** live root HTML response
- **Evidence:** The observed response lacked CSP, HSTS, frame, MIME, referrer, and permissions headers; the candidate defines most of them.
- **Impact:** Browser-side containment for injection, clickjacking, content sniffing, and referrer leakage is weaker than designed.
- **Fix:** Apply and test the candidate edge header policy to all HTML responses.
- **Mitigation:** Cloudflare and HTTPS reduce transport exposure but do not supply equivalent browser containment.
- **False-positive assessment:** Header presence can vary by route; the finding is scoped to the observed live HTML response and should be tested across all entry points.

### SEC-6 — `gosec` G404 is a gate defect, not a demonstrated vulnerability

- **Rule ID:** `G404`
- **Severity:** Scanner reports High; assessed Informational for this use
- **Location:** `backend/internal/services/webhook_retry_service.go:198`
- **Evidence:** `math/rand.Float64` perturbs retry timing by ±20%; it does not protect a secret or security decision.
- **Impact:** No direct security impact was identified, but the unsuppressed finding breaks mandatory CI and encourages unsafe blanket waivers.
- **Fix:** Add the scanner's precise suppression syntax or use a cryptographic random source.
- **Mitigation:** None needed for runtime security; CI must still be repaired.
- **False-positive assessment:** Likely false positive based on data use and local context.

## Verification results

| Check | Result | Notes |
| --- | --- | --- |
| `task contract` | Pass | Task contract is internally consistent. |
| Backend `go test ./...` | Pass | All discovered packages passed. |
| Backend `go vet ./...` | Pass | No vet failure. |
| Backend `go build ./cmd/api` | Pass | API binary builds. |
| Frontend ESLint with zero warnings | Pass | Direct local ESLint invocation exited zero. |
| Frontend Vitest | Pass | 133 files, 1,669 tests; 367.51 seconds. Console/`act` warning debt remains. |
| Frontend Vite production build | Pass | Bundle budget passes; initial app about 494.72 kB raw. |
| UI boundary check | Pass | 460 source modules checked. |
| Playwright discovery | Pass | 32 tests discovered across four configured projects. This is not hosted execution evidence. |
| OpenAPI lint | Pass with warnings | Exit zero with 84 warnings. |
| Documentation checks | Pass | 201 Markdown files, zero lint errors; anchors, graph, and assets pass. |
| `govulncheck` | Fail | Three reachable vulnerabilities. |
| `gosec` exact CI command | Fail | One G404 finding on non-security retry jitter. |
| Release evidence gate | Fail | Six mandatory files missing. |
| Live API health | Pass | HTTP 200 and healthy payload. |
| Live public browser smoke | Fail | `/search` crashes; home/legal/login pages otherwise render. |
| Live accessibility smoke | Fail | 41 serious affected nodes/findings on home. |
| Live static endpoint contracts | Fail | Multiple metadata/health/unknown routes return SPA HTML. |
| Frontend npm audit | Not run | `npm` was unavailable on the audit host; dependency scanning must run in hosted CI. |

## Required release sequence

1. Fix the production search crash and add nullable-response regression coverage.
2. Upgrade the three vulnerable Go dependency paths and make both security scanners pass without broad exclusions.
3. Resolve the working tree into a clean candidate commit; build, scan, sign, and record immutable image digests.
4. Run source CI on a supported hosted runner, including WebKit, browser test execution, dependency audits, secret scan, SBOM, and container scan; verify required branch protection.
5. Run authenticated/provider-backed and Stripe test-mode journeys with disposable fixtures.
6. Run staging baseline, stress, and soak profiles against defined thresholds.
7. Validate backup restore, migration compatibility, RPO/RTO, canary, drain, rollback, and post-rollback smoke.
8. Complete JWT exposure/rotation review and populate all six evidence artifacts against the same candidate SHA.
9. Deploy the digest-pinned candidate using the hardened blue/green configuration.
10. Verify live container controls, response headers, static metadata, search, accessibility, core public and authenticated journeys, monitoring, and rollback readiness before shifting full traffic.

## Go/no-go criteria

Release may change to **GO** only when all of the following are true:

- The search regression is fixed and verified against the production-like backend.
- `govulncheck`, `gosec`, frontend dependency audit, secret scan, and container scan meet policy.
- All source test/build/contract jobs pass from a clean checkout with stable totals and accepted warnings/skips.
- The six release-evidence JSON files are present, valid, authoritative, and bound to the final SHA and image digest.
- Hosted cross-browser and critical authenticated/payment journeys pass.
- Staging performance and production-class restore/rollback evidence meets declared thresholds.
- The deployed images match the approved digest and effective runtime hardening.
- Post-deploy smoke confirms correct headers, routes, accessibility, monitoring, and rollback capability.

## Final assessment

**No — not ready for production release.** The candidate is directionally strong, but four independent release blockers remain: production artifact drift, a live critical-route crash, reachable dependency vulnerabilities, and absent operator evidence. Approval should be reconsidered only after the required release sequence completes against one immutable candidate.
