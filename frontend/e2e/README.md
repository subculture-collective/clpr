# Browser test tiers

Playwright has two explicit tiers:

- `mocked-chromium` checks isolated UI behavior and may intercept network calls.
- `real-*` runs launch smoke tests against the repository-owned frontend and a
  real backend at `PLAYWRIGHT_API_BASE_URL` (default
  `http://127.0.0.1:18088`, overridable with `PLAYWRIGHT_API_BASE_URL`). These
  projects are the release gate. The non-default port avoids colliding with
  services commonly bound to port 8080 on developer and CI hosts.

The older specs in `e2e/tests/` are quarantined because they depend on missing
fixtures, fabricate backend state, and inject fallback UI. They are intentionally
excluded from Playwright projects until rewritten or removed; they do not count
as coverage.

Commands:

```bash
npm run test:e2e:list
npm run test:e2e:mocked
npm run test:e2e:real
```

The real tier must never skip because services or seed data are missing. CI is
responsible for starting dependencies and fails if a real project cannot run.
