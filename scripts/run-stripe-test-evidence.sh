#!/usr/bin/env bash

set -euo pipefail

mode="${1:---dry-run}"

describe() {
    cat <<'EOF'
Stripe provider evidence requires only explicitly test-scoped inputs:
  STRIPE_MODE=test
  STRIPE_TEST_SECRET_KEY=sk_test_...
  STRIPE_TEST_WEBHOOK_SECRET=whsec_...
  STRIPE_TEST_MONTHLY_PRICE_ID=price_...
  STRIPE_TEST_YEARLY_PRICE_ID=price_...
  STRIPE_PROVIDER_RUNNER=/protected/path/to/executable-runner

The runner must prove checkout/return, signed and duplicate/out-of-order
webhooks, dunning recovery, cancellation, reconciliation, and entitlement
activation/revocation. No STRIPE_SECRET_KEY fallback is used.
EOF
}

if [[ "$mode" == "--dry-run" ]]; then
    describe
    exit 0
fi
[[ "$mode" == "--execute" ]] || { echo "usage: $0 [--dry-run|--execute]" >&2; exit 2; }

[[ "${STRIPE_MODE:-}" == "test" ]] || { echo "STRIPE_MODE must explicitly equal test" >&2; exit 1; }
[[ "${STRIPE_TEST_SECRET_KEY:-}" == sk_test_* ]] || { echo "STRIPE_TEST_SECRET_KEY must start with sk_test_" >&2; exit 1; }
[[ "${STRIPE_TEST_SECRET_KEY:-}" != sk_live_* ]] || { echo "live Stripe keys are forbidden" >&2; exit 1; }
[[ "${STRIPE_TEST_WEBHOOK_SECRET:-}" == whsec_* ]] || { echo "STRIPE_TEST_WEBHOOK_SECRET must start with whsec_" >&2; exit 1; }
[[ "${STRIPE_TEST_MONTHLY_PRICE_ID:-}" == price_* ]] || { echo "STRIPE_TEST_MONTHLY_PRICE_ID is required" >&2; exit 1; }
[[ "${STRIPE_TEST_YEARLY_PRICE_ID:-}" == price_* ]] || { echo "STRIPE_TEST_YEARLY_PRICE_ID is required" >&2; exit 1; }
[[ -x "${STRIPE_PROVIDER_RUNNER:-}" ]] || { echo "STRIPE_PROVIDER_RUNNER must be an executable protected runner" >&2; exit 1; }

# Export the names consumed by the application only for this child process.
# There is intentionally no lookup of production-style fallback variables.
STRIPE_SECRET_KEY="$STRIPE_TEST_SECRET_KEY" \
STRIPE_WEBHOOK_SECRET="$STRIPE_TEST_WEBHOOK_SECRET" \
STRIPE_PRO_MONTHLY_PRICE_ID="$STRIPE_TEST_MONTHLY_PRICE_ID" \
STRIPE_PRO_YEARLY_PRICE_ID="$STRIPE_TEST_YEARLY_PRICE_ID" \
    "$STRIPE_PROVIDER_RUNNER"
