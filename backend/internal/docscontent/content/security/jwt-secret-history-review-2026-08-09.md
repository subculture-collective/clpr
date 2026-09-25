# JWT and Secret-History Review — 2026-08-09

## Decision

The repository history and candidate tree pass the pinned Gitleaks v8.28.0
policy. No JWT signing material was found in Git history, and the exact private
JWT key currently configured on the production host does not occur in any Git
revision. This review does not expose or record any secret values.

## Scope and method

- Full history: 1,012 commits and approximately 49.63 MB scanned with
  `gitleaks git --redact=100 --config .gitleaks.toml .`.
- Candidate tree: approximately 19.04 MB scanned with
  `gitleaks dir --redact=100 --config .gitleaks.toml .`.
- Result after reviewed dispositions: zero findings in both scans.
- The JWT/PEM audit reviewed 45 historical signing-configuration assignments
  and the three private-key detector hits. The hits contain PEM markers or
  placeholders, but no encoded private-key payload.
- `/opt/server/projects/clpr/.env` was inspected by key name and comparison only.
  It is mode `0600`; secret values were neither printed nor copied. Its exact
  `JWT_PRIVATE_KEY_B64` value was searched across all Git revisions without a
  match.

## Finding disposition

The original redacted scan reported 105 history findings and 51 candidate-tree
findings. Literal curl placeholders and truncated documentation tokens are
covered by rule-specific regexes. Eleven remaining historical fixture findings
are covered by full commit/path/rule/line fingerprints. The committed policy
test proves the placeholder remains allowed while a new credential and a
same-line replacement of a historical fixture still fail scanning.

The remaining historical categories were reviewed as follows:

- Three private-key findings: marker-only documentation/configuration examples;
  the audit rejects a base64 payload appearing beneath those markers.
- Four Stripe findings: test-mode documentation examples, not live-mode keys.
- Two generic findings: a websocket authentication test fixture and a truncated
  API documentation example.
- Two Twitch findings: historical compose configuration. Both historical values
  differ from the values currently configured on the production host, which is
  evidence of apparent rotation only.

## Residual provider action

Repository evidence cannot prove provider-side revocation. An authorized Twitch
operator must confirm in the provider console that the historical credentials
are revoked. Until that external check is recorded, treat Twitch revocation as
an operator-evidence item rather than claiming it is complete.

## Enforcement

The supply-chain workflow installs Gitleaks from the exact Go module version
`github.com/zricethezav/gitleaks/v8@v8.28.0`, scans full history and the current
tree with 100% redaction, runs the JWT/PEM audit, and executes the fail-closed
policy regression test. A new unreviewed finding fails CI.
