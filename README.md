# CLPR

> A deployed, full-stack Twitch clip curation platform built to make discovery, community context, and moderation usable in one place.

[![CI status](https://git.subcult.tv/subculture-collective/clpr/actions/workflows/verify.yml/badge.svg)](https://git.subcult.tv/subculture-collective/clpr/actions/workflows/verify.yml)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)

**Live product:** [clpr.tv](https://clpr.tv) · **API contract:** [OpenAPI](docs/openapi/openapi.yaml) · **License:** [MIT](LICENSE)

## What it demonstrates

- Built a Go and PostgreSQL application around Twitch OAuth, ingestion, clip submission, playlists, and community interaction.
- Designed hybrid BM25 and semantic search, with queue-backed background work and graceful search fallbacks.
- Shipped moderation and audit controls, media ownership handling, operational runbooks, and a React web client.
- Operates a public product with automated database migration drills, documentation checks, and a production deployment path.

## Architecture

| Area | Implementation |
| --- | --- |
| Product UI | React, TypeScript, Vite, Tailwind |
| API and jobs | Go, Gin, Redis-backed workers |
| Data and search | PostgreSQL/pgvector, OpenSearch |
| Integrations | Twitch OAuth/API, webhooks, object storage |
| Operations | Docker Compose, Kubernetes manifests, Prometheus/Grafana runbooks |

## Run it locally

The supported development path and required services are documented in [the development guide](docs/setup/development.md). For a quick local stack:

```bash
cp .env.development.example .env
docker compose up -d
cd backend && go run cmd/migrate/main.go up
```

The web app is then available at `http://localhost:5173` and the API at `http://localhost:8080`.

## Verification

```bash
make test-unit
make lint
```

Operational backup validation and restore drills require protected provider credentials. Their scheduled public workflows now report a safe skip unless the explicit `BACKUP_VALIDATION_ENABLED` or `RESTORE_DRILL_ENABLED` secret is set to `true`; this prevents a missing credential from being misrepresented as a product regression.

## Further reading

- [Architecture](docs/backend/architecture.md)
- [API reference](docs/openapi/README.md)
- [Operations runbooks](docs/operations/runbooks/README.md)
- [Contributing](docs/contributing.md)

## License

CLPR is released under the [MIT License](LICENSE).
