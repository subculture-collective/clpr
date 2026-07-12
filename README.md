# Clipper

> A modern, community-driven Twitch clip curation platform

[![CI Status](https://git.subcult.tv/subculture-collective/clpr/workflows/CI/badge.svg)](https://git.subcult.tv/subculture-collective/clpr/actions)
[![License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)

Clipper is a full-stack platform for discovering, curating, and sharing the best Twitch clips. It combines powerful search capabilities, community voting, and social features to help users find and organize memorable gaming moments.

> **Pre-release scope:** The current repository ships a responsive web client and
> Go API. Native mobile clients are design/planning work only; there is no
> buildable `mobile/` workspace in this tree. Stream clip extraction, CDN
> mirroring, live feed, and watch parties are disabled by default while their
> production acceptance gates are completed. See the
> [production-readiness remediation plan](docs/superpowers/plans/2026-07-12-production-readiness-remediation.md).

## ✨ Key Features

- **🔍 Advanced Search**: Hybrid BM25 + semantic vector search with natural language queries
- **⬆️ Community Curation**: Reddit-style voting, comments, and karma system
- **📱 Responsive Web**: React web experience across desktop and mobile browsers
- **💎 Premium Features**: Unlimited collections, advanced filters, and cross-device sync
- **🎮 Twitch Integration**: OAuth login, broadcaster data, and clip submission
- **🚀 Modern Stack**: Go backend and React frontend

## 🚀 Quick Start

### Prerequisites

- **Docker** & **Docker Compose** (recommended)
- **Node.js** 20+ (for frontend development)
- **Go** 1.24+ (for backend development)
- **PostgreSQL** 17+ (if running without Docker)
- **Redis** 8+ (if running without Docker)

### Reproducible local validation

The release-supported clean-clone path uses disposable test services and never
requires production credentials. Install [Task](https://taskfile.dev/) and
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

`task test:setup` creates a separate test database, Redis, and OpenSearch and
applies every migration. The ordinary deployment Compose file remains an
operator path and expects deployment-managed networks and secrets.

## Development Without Docker

Backend commands and prerequisites are documented in the
[backend README](backend/README.md). Frontend commands are defined in
`frontend/package.json`.

## 📚 Documentation

Comprehensive documentation is available in the [\`/docs\`](docs/) directory:

- **[Backend](docs/backend/index.md)** - Backend documentation hub
- **[Frontend](docs/frontend/index.md)** - Frontend documentation hub
- **[API Reference](docs/openapi/README.md)** - OpenAPI documentation
- **[Operations](docs/operations/index.md)** - Deployment and operations hub
- **[Site Freshness Automation](docs/operations/SITE_FRESHNESS_AUTOMATION.md)** - Bootstrapping auto-refreshing public playlists
- **[Contributing](docs/contributing.md)** - Contribution guidelines

**📖 Full documentation index**: [\`docs/index.md\`](docs/index.md)

## 🏗️ Architecture

Clipper is built as a modern, scalable full-stack application:

```
┌─────────────┐                           ┌───────────────┐
│   Web App   │──────────────────────────▶│   Backend API │
│  (React)    │                           │    (Go/Gin)   │
└─────────────┘                           └───────┬───────┘
                                                    │
                          ┌─────────────────────────┼─────────────────────┐
                          ▼                         ▼                     ▼
                    ┌──────────┐            ┌──────────┐         ┌─────────────┐
                    │PostgreSQL│            │  Redis   │         │ OpenSearch  │
                    │  (Data)  │            │ (Cache)  │         │  (Search)   │
                    └──────────┘            └──────────┘         └─────────────┘
```

- **Frontend**: React 19 + TypeScript + Vite + TailwindCSS
- **Backend**: Go 1.24 + Gin + PostgreSQL + Redis
- **Search**: OpenSearch 2.11 with hybrid BM25 + vector search
- **Infrastructure**: Docker, Kubernetes, GitHub Actions

See the [backend documentation hub](docs/backend/index.md) for current details.

## 🛠️ Tech Stack

### Backend
- **Language**: Go 1.24+
- **Framework**: Gin (HTTP web framework)
- **Database**: PostgreSQL 17 with pgx driver
- **Cache**: Redis 8 with go-redis
- **Search**: OpenSearch 2.11
- **Auth**: JWT with Twitch OAuth
- **Queue**: Redis-based background jobs

### Frontend (Web)
- **Framework**: React 19 with TypeScript
- **Build Tool**: Vite
- **Styling**: TailwindCSS
- **Routing**: React Router 7
- **State**: TanStack Query + Zustand
- **Forms**: React Hook Form

### Infrastructure
- **Containers**: Docker & Docker Compose
- **Orchestration**: Kubernetes (production)
- **CI/CD**: GitHub Actions
- **Monitoring**: Prometheus + Grafana
- **Secrets**: Environment variables and platform-managed secret stores

## 🤝 Contributing

We welcome contributions! Please see our [Contributing Guide](docs/contributing.md) for:

- Code of conduct
- Development workflow
- Code style guidelines
- Testing requirements
- Pull request process

### Quick Contribution Steps

1. Fork the repository
2. Create a feature branch (\`git checkout -b feature/amazing-feature\`)
3. Make your changes and add tests
4. Run linters and tests (\`task test\`; \`make test\` remains as a compatibility wrapper)
5. Commit your changes (\`git commit -m 'Add amazing feature'\`)
6. Push to the branch (\`git push origin feature/amazing-feature\`)
7. Open a Pull Request

### Task runner

Common development commands live in `Taskfile.yml`. Use `task --list` to see
organized targets such as `task test:backend`, `task test:frontend`,
`task docker:build`, and `task migrate:up`. The root `Makefile` is intentionally
small and delegates legacy `make <target>` calls to the matching Task target.

## 📝 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🔗 Links

- **Documentation**: [docs/index.md](docs/index.md)
- **Issue Tracker**: [GitHub Issues](https://git.subcult.tv/subculture-collective/clpr/issues)
- **Discussions**: [GitHub Discussions](https://git.subcult.tv/subculture-collective/clpr/discussions)
- **Twitch API**: [Twitch Developer Docs](https://dev.twitch.tv/docs/api/)

## 🙏 Acknowledgments

Built with ❤️ by the [Subculture Collective](https://git.subcult.tv/subculture-collective)

Special thanks to:
- The Twitch developer community
- All our contributors
- Open source projects that make this possible

---

**Status**: Active Development | **Version**: v0.x (Pre-release) | **Release gate**: [Remediation plan](docs/superpowers/plans/2026-07-12-production-readiness-remediation.md)
