---
title: "CLPR documentation"
summary: "Current product, development and operations references."
tags: ["docs", "hub", "index"]
area: "docs"
status: "stable"
owner: "team-core"
version: "1.0"
last_reviewed: 2026-11-16
---

# CLPR documentation

CLPR is a deployed responsive web app for Twitch clip discovery, curation and
sharing, with a Go API. Accounts are free. Separate native iOS and Android apps
are now being planned; there is no buildable native app yet.

## Start here

- [Product overview](../README.md)
- [Development setup and verification](../DEVELOPMENT.md)
- [Current development status and open work](product/development-status.md)
- [Launch scope and disabled capabilities](LAUNCH_FEATURE_INVENTORY.md)
- [Feature inventory](product/feature-inventory.md)
- [Native apps plan](product/native-apps-plan.md)
- [Complete documentation reference inventory](reference-inventory.md)

## Product and implementation

- [Product hub](product/index.md)
- [Feature guides](features/index.md)
- [Countdowns and Shorts](features/countdown-editions.md)
- [Transcript search](features/transcript-search.md)
- [Backend documentation](backend/index.md)
- [Frontend documentation](frontend/index.md)
- [OpenAPI reference](openapi/README.md)
- [Architecture decisions](adr/index.md)
- [RFCs](rfcs/index.md)
- [Browser search artifact](development/search-wasm.md)
- [Marketing packet and campaign planning](marketing/README.md)

## Verification and operations

- [Testing](testing/index.md)
- [Self-QA journeys](testing/self-qa-flow-checklist.md)
- [Operations](operations/index.md)
- [Countdown runbook](operations/runbooks/countdowns.md)
- [Deployment runbook](deployment/runbook.md)
- [Release evidence requirements](../release-evidence/README.md)
- [Compliance](compliance/index.md)
- [Contributing](contributing.md)
- [Documentation checks](contributing/docs-quality-checks.md)

## Status and ownership

Use the [current status](product/development-status.md) and
[GitHub milestones](https://github.com/subculture-collective/clpr/milestones).
Old roadmap versions and Q2 2025 launch dates are historical. Source completion,
CI, deployment and provider/operational acceptance are reported separately.

Branches, PRs, issues and discussions live on
[GitHub](https://github.com/subculture-collective/clpr). Gitea mirrors the code
and still runs homelab release workflows. Earlier native and premium documents
are in the [archive](archive/index.md); the current native effort has its own
plan above.

Run `npm run docs:inventory` when adding or removing a document, then
`npm run docs:check` and `npm run docs:embed:check`. Only the explicit documents
in `public-docs.json` are embedded into the API; repository plans and operations
notes are not automatically published on clpr.tv.
