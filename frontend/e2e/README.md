# Browser test tiers

- `mocked-chromium` checks UI behavior with controlled API responses.
- `real-chromium`, `real-firefox`, and `real-webkit` run against the built
  frontend and repository-owned API. Use `task test:setup` followed by
  `task test:e2e:real` from the repository root.
- Candidate projects use `playwright.candidate.config.ts`, an explicit HTTPS
  URL, and explicit fixture identifiers. Accessibility runs in all three
  browsers; performance budgets run in Chromium.

From `frontend`:

```bash
npm run test:e2e:list
npm run test:e2e:mocked
npm run test:e2e:candidate
```

The real tier must never skip because services or seed data are missing.
Mocked checks cannot satisfy the real-backend release gate. Every spec must
belong to a configured project; do not keep quarantined or demonstration suites
outside test discovery.

See the [testing guide](../../docs/testing/TESTING.md) for ownership and rules.
