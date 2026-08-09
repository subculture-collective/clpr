# CLPR

> A deployed, full-stack Twitch clip curation platform built to make discovery, community context, and moderation usable in one place.

[![CI status](https://git.subcult.tv/subculture-collective/clpr/actions/workflows/verify.yml/badge.svg)](https://git.subcult.tv/subculture-collective/clpr/actions/workflows/verify.yml)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)

**Live product:** [clpr.tv](https://clpr.tv) · **API contract:** [OpenAPI](docs/openapi/openapi.yaml) · **License:** [MIT](LICENSE)

> **Pre-release scope:** The current repository ships a responsive web client and
> Go API. Native mobile clients are design/planning work only; there is no
> buildable `mobile/` workspace in this tree. Stream clip extraction, CDN
> mirroring, live feed, and watch parties remain disabled while their production
> acceptance gates are completed. See the
> [production-readiness remediation plan](docs/superpowers/plans/2026-07-12-production-readiness-remediation.md).

## What it demonstrates

- Built a Go and PostgreSQL application around Twitch OAuth, ingestion, clip submission, playlists, and community interaction.
- Designed hybrid BM25 and semantic search, with queue-backed background work and graceful search fallbacks.
- Shipped moderation and audit controls, media ownership handling, operational runbooks, and a React web client.
- Provides automated migration drills, documentation checks, release convergence, and a production deployment path.

## Architecture

| Area | Implementation |
| --- | --- |
| Product UI | React, TypeScript, Vite, Tailwind |
| API and jobs | Go, Gin, Redis-backed workers |
| Data and search | PostgreSQL/pgvector, OpenSearch |
| Integrations | Twitch OAuth/API, webhooks, object storage |
| Operations | Docker Compose, Kubernetes manifests, Prometheus/Grafana runbooks |

## Run it locally

Backend prerequisites and local setup are documented in the [backend README](backend/README.md). Start by creating the service-specific development configuration files:

```bash
cp backend/.env.example backend/.env
cp frontend/.env.example frontend/.env
```

The root Compose file is an operator deployment path and expects deployment-managed networks and secrets. Use the disposable validation stack below for repository-owned local services.

## Verification

The release-supported clean-clone path uses disposable test services and does
not require production credentials. Install [Task](https://taskfile.dev/) and
`golang-migrate`, then run:

```bash
cp backend/.env.example backend/.env
cp frontend/.env.example frontend/.env
task test:setup
bash scripts/run-release-critical-backend-tests.sh
npm --prefix frontend ci
npm --prefix frontend run test:coverage
npm --prefix frontend run build
task test:teardown
```

Operational backup validation and restore drills require protected provider credentials. Their scheduled public workflows report a safe skip unless the explicit `BACKUP_VALIDATION_ENABLED` or `RESTORE_DRILL_ENABLED` secret is set to `true`; this prevents a missing credential from being misrepresented as a product regression.

## Further reading

- [Backend architecture and design](docs/backend/index.md)
- [API reference](docs/openapi/README.md)
- [Operations runbooks](docs/operations/runbooks/README.md)
- [Contributing](docs/contributing.md)

## License

CLPR is released under the [MIT License](LICENSE).
