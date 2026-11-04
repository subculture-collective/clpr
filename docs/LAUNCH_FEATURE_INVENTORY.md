# Launch Feature Inventory

This inventory is the authoritative release-scope contract for the first
production launch. A feature marked **disabled** must not be advertised,
navigable, registered as a production API route, or enabled by release-profile
configuration.

| Capability | Launch state | Backend contract | Frontend contract | Promotion gate |
| --- | --- | --- | --- | --- |
| Automated account deletion | Disabled | Deletion request, cancellation, and status routes are not registered | Settings does not promise or offer automated deletion | Transactional erasure/anonymization worker, session and OAuth revocation, retained-data policy, recovery/retry behavior, and lifecycle integration evidence |
| Automated data-subject export | Disabled | `/api/v1/users/me/export` is not registered; creator-content exports are unaffected | Settings does not claim a complete automated privacy export | Inventory every personal-data category, fail closed on partial collection, redact third-party data, retention-policy review, and full archive contract tests |

## Executable evidence

- `backend/cmd/api/routes_feature_flags_test.go` proves the listed routes are
  absent from the production router.
- Frontend lint and route tests protect the active route tree; disabled pages
  may remain as non-routable implementation work but are not launch features.

## Change control

Promoting a capability requires updating this inventory, its server validation,
backend route tests, frontend route/navigation tests, public documentation, and
the named promotion-gate evidence in one reviewed change. A runtime flag alone
does not make an incomplete capability production-supported.
