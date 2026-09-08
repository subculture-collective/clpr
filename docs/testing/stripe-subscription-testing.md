---
title: "Legacy Billing Testing"
summary: "Verify free access and retained billing obligations without creating new paid enrollment."
tags: ["testing"]
area: "testing"
status: "stable"
owner: "team-core"
version: "2.0"
last_reviewed: 2026-09-07
---

# Legacy billing and retirement testing

CLPR accounts are free. New subscriptions, paid entitlements, plan changes,
reactivation, and billing portal enrollment are retired. Optional external
support does not change account permissions or rate limits.

Historical subscriptions, invoices, failures, receipts, and audit records remain
intact. Signed Stripe webhooks reconcile those records without granting paid
access. Cancellation and invoice access remain authenticated and rate limited.
They require configured credentials and `LEGACY_BILLING_SERVICING=true`.
Ordinary outbound webhook subscriptions are a separate supported capability.

## Automated persistence contract

Against the disposable test database, run:

```bash
cd backend
go test -tags=integration ./tests/integration/premium \
  -run TestSignedStripeWebhookLifecycleReconcilesLegacyBillingIdempotently
```

The retained package name reflects its migration history. Its contract now
checks signed receipts, duplicate delivery, out-of-order subscription updates,
payment failure and recovery, cancellation events, and rejected signatures with
real PostgreSQL persistence. Middleware tests verify that historical paid status
does not increase member rate limits. These tests do not establish the state of
provider-side obligations.

## Provider inventory and servicing

Use a read-only inventory of application billing records, pending work, webhook
retries, and the complete provider subscription catalog. Report only counts and
states. Credentials and customer details belong in protected operator storage.

Empty application tables are insufficient to prove retirement is complete.
An unavailable or rejected provider credential leaves the inventory unverified.

For existing obligations, the protected test-provider runner invoked by
`scripts/run-stripe-test-evidence.sh --execute` must verify cancellation,
invoice access, signed and replayed webhooks, and reconciliation. It accepts
explicit test credentials only. Do not create new paid subscriptions, cancel
real subscriptions, issue refunds, or contact customers during qualification.

## Release evidence

The [release evidence contract](../../release-evidence/README.md) requires
`billing-retirement.json` with the exact candidate SHA and immutable image
digests, verified free access, no new paid enrollment, and a complete inventory.

All obligation counts must be zero to select `obligation_status=zero`.
Otherwise select `legacy_servicing` and supply passing servicing evidence.
Missing credentials, incomplete inventory, or failed servicing block acceptance.
