# Release evidence

This directory intentionally contains no passing evidence in source control.
The release operator supplies four JSON files as a protected workflow artifact
or from an approved evidence store, then runs `node scripts/verify-release-evidence.js`.

Each file must contain `candidate_sha`, ISO-8601 `executed_at`, `evidence_url`,
and `owner`, plus these booleans:

- `stripe-test-mode.json`: `mode` is `test`; checkout, signed webhooks, and
  reconciliation passed.
- `load-and-soak.json`: baseline, stress, soak, and thresholds passed.
- `restore.json`: restore, application smoke, RPO, and RTO passed.
- `rollback.json`: canary, graceful drain, rollback, and post-rollback smoke passed.

Evidence files may contain deployment identifiers and test resource IDs but
must never contain API keys, tokens, customer data, or webhook secrets.

