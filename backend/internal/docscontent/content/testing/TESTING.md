# Testing CLPR

Tests protect observable behavior at the lowest layer that can prove it.
Prefer a regression that fails when a user loses access, data is corrupted,
a permission is bypassed, or a request changes its public contract.

## Commands

Run from the repository root unless a command changes directory.

| Check | Command | What it proves |
| --- | --- | --- |
| Frontend maintenance gate | `cd frontend && npm run check` | Type safety, lint, source reachability, test ownership, and critical coverage |
| Backend unit and package tests | `cd backend && go test ./...` | Domain rules, handlers, security, and package contracts |
| Frontend tests and coverage | `cd frontend && npm run test:coverage` | Component interaction, state, API adaptation, and coverage thresholds |
| Unused frontend source | `cd frontend && npm run code:unused` | All application files are reachable from the entry point |
| Test ownership and discovery | `cd frontend && npm run test:inventory` | Critical suites exist; browser specs belong to an executable tier; no focused or placeholder frontend tests |
| Disposable services | `task test:setup` | Starts PostgreSQL, Redis, and OpenSearch and applies test migrations |
| Backend integration packages | `task test:integration` | Integration-tagged tests in `internal/handlers`, `internal/repository`, `internal/services`, and `tests/migrations`, run in parallel |
| Release-critical backend | `task test:release-critical` | Runs security and mutation tests without skips, plus real webhook persistence and idempotency |
| Mocked browser | `task test:e2e:mocked` | Browser layout, keyboard access, consent, and public UI |
| Real backend browser | `task test:e2e:real` | Builds the API, seeds disposable data, and checks the real application in three browsers |
| Migration drills | See [migration tests](../../backend/tests/migrations/README.md) | Schema upgrade, rollback, and recovery on dedicated databases |

Use the Go toolchain in `backend/go.mod`. After service-backed checks, run
`task test:teardown`. Never point test fixtures or migration drills at production.
If the default ports are occupied, use the `TEST_DATABASE_PORT`,
`TEST_REDIS_PORT`, and `TEST_OPENSEARCH_PORT` overrides before setup.

Each database-backed Go package calls `testutil.RunWithDatabase` from
`TestMain`. It clones a private database from `clpr_test_template`, which is
refreshed from the migrated `clpr_test` database whenever their migration
versions differ, and drops the clone when the package finishes. Packages
therefore run in parallel without sharing rows or schema state, and
`clpr_test` itself is left for the API used by browser tests. A new
database-backed package must add the same `TestMain`; `testutil.SetupTestDB`
fails without it. Set `TEST_DATABASE_KEEP=1` to keep the clones for
inspection. Run one package with `task test:integration:pkg PKG=internal/repository`.

## Behavior ownership

[The critical contract registry](../../config/test-contracts.json) identifies
the suites responsible for identity, authorization, consent, feed and search,
playback and queues, submissions, payments, accessibility, and real API smoke.
Update it when moving or replacing a suite. It is a review map, not a claim that
file existence proves coverage; the listed tests must execute and pass.

Preserve focused cases for OAuth state and replay, CSRF, cross-user writes,
permission checks, upload validation, payment signatures and duplicate delivery,
pagination boundaries, retry behavior, and recovery from failed requests.
A test should fail because the product contract broke, not because a mock was
renamed or markup was rearranged.

## Choosing a test layer

- Use Go unit tests for domain rules and request validation. Use real PostgreSQL
  or Redis tests for transactions, concurrent claims, query behavior, and expiry.
- Use Vitest for state transitions and component interactions. Assert accessible
  roles, request parameters, returned data, and effects visible to users.
- Use Playwright for complete browser journeys, layout, focus, contrast, and
  browser APIs. Mocked browser checks do not prove database or authentication
  integration. The real tier must fail when its services or seed data are absent.
- Keep deployment, restore, migration, and load checks in their dedicated tiers.

Do not test that a struct can be initialized, a constructor returns its declared
type, every copy string is unchanged, or a component contains specific utility
classes. Avoid rendering the same component repeatedly for single-attribute
assertions. Table-driven boundary cases are useful when inputs represent
distinct failure modes; a large case count alone is not useful.

## Coverage and skips

Coverage floors live in `frontend/config/critical-coverage.json` and are enforced
per module by Vitest in CI and `npm run check`. The complete application coverage
report remains available; its aggregate percentage is diagnostic, not a gate.
Authentication, route guards, encrypted storage, consent, URL validation, API
session handling, search recovery, playlists, and queues have independent floors.
This replaces the inherited failing global average, which could hide a critical
regression behind coverage elsewhere.

Guards, PKCE, and playlist state require 100% coverage. Existing lower floors on
the auth context and API client are explicit remaining gaps, not completeness
claims. Raise the relevant floor when strengthening its tests. Do not lower a
floor to accommodate deletions; moving a module requires moving its coverage
contract, and the inventory check rejects missing paths. Remove a test only when its behavior
is obsolete, covered by a better test, or was never exercised by its assertions.
A green unit run does not prove service-backed integration tests ran.

Backend prerequisite skips are explicitly inventoried in
`config/test-skip-inventory.tsv`. Temporary disabled tests require an expiry;
service or short-mode prerequisites use the `prerequisite` classification.
The release-critical runner separately rejects skipped tests in its required
handler, middleware, and JWT suites.

Browser specs must live under `e2e/mocked`, `e2e/real-backend`, or
`e2e/candidate`. Candidate performance checks run only in Chromium because the
long-task measurements require its browser API. Project selection expresses
that constraint; runtime skips do not.

## Maintaining the suite

For a bug fix, add a regression at the boundary that failed. For a refactor,
keep the existing behavioral assertions and change only setup that needs it.
For a removed feature, remove its unreachable implementation and tests together.
Keep test helpers out of production Go files unless production calls them.

Vitest rejects focused tests even when run directly and caps worker concurrency
at four for predictable memory use. CI requires production TypeScript checks;
tests remain executable behavior checks rather than part of the app build.

Keep run logs, screenshots, coverage output, and release artifacts outside
tracked source. Git history holds old plans and completion reports; operational
runbooks and machine-readable release evidence templates remain maintained
documentation.

See [browser tiers](../../frontend/e2e/README.md),
[Stripe lifecycle testing](stripe-subscription-testing.md), and
[release evidence requirements](../../release-evidence/README.md).
