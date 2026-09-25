# Billing retirement verification

Paid enrollment and the legacy Stripe runtime are retired after a complete
read-only provider and database inventory found zero outstanding obligations.
Historical billing tables, migration history, and audit records remain intact.

Run `go test ./cmd/api -run TestBillingRetirement` from `backend` to verify that
billing endpoints are absent and unrelated webhook routes remain registered.
The maintained [release evidence contract](../../release-evidence/README.md)
requires current provider inventory; empty application tables alone are insufficient.
No live customer mutation belongs in qualification.
