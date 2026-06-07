# Clipper

> A modern, community-driven Twitch clip curation platform

[![CI Status](https://git.subcult.tv/subculture-collective/clpr/workflows/CI/badge.svg)](https://git.subcult.tv/subculture-collective/clpr/actions)
[![License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)

Clipper is a full-stack platform for discovering, curating, and sharing the best Twitch clips. It combines powerful search capabilities, community voting, and social features to help users find and organize memorable gaming moments.

## ✨ Key Features

- **🔍 Advanced Search**: Hybrid BM25 + semantic vector search with natural language queries
- **⬆️ Community Curation**: Reddit-style voting, comments, and karma system
- **📱 Multi-Platform**: Responsive web app + native iOS/Android apps
- **💎 Premium Features**: Unlimited collections, advanced filters, and cross-device sync
- **🎮 Twitch Integration**: OAuth login, live streams, and clip submission
- **🚀 Modern Stack**: Go backend, React frontend, React Native mobile

## 🚀 Quick Start

### Prerequisites

- **Docker** & **Docker Compose** (recommended)
- **Node.js** 20+ (for frontend/mobile development)
- **Go** 1.24+ (for backend development)
- **PostgreSQL** 17+ (if running without Docker)
- **Redis** 8+ (if running without Docker)

### Getting Started with Docker

\`\`\`bash
# Clone the repository
git clone https://git.subcult.tv/subculture-collective/clpr.git
cd clpr

# Copy environment files
cp .env.development.example .env

# Start all services
docker-compose up -d

# Run database migrations
cd backend
go run cmd/migrate/main.go up

# Access the application
# Frontend: http://localhost:5173
# Backend API: http://localhost:8080
# Docs: http://localhost:3000
\`\`\`


This will start all services in Docker containers.

## Development Without Docker

See the complete [Development Setup Guide](docs/setup/development.md) for detailed instructions.

## 📚 Documentation

Comprehensive documentation is available in the [\`/docs\`](docs/) directory:

- **[Getting Started](docs/setup/development.md)** - Development environment setup
- **[User Guide](docs/users/user-guide.md)** - Using the platform
- **[API Reference](docs/backend/api.md)** - REST API documentation
- **[Architecture](docs/backend/architecture.md)** - System design and components
- **[Deployment](docs/operations/deployment.md)** - General deployment guide
- **[Site Freshness Automation](docs/operations/SITE_FRESHNESS_AUTOMATION.md)** - Bootstrapping auto-refreshing public playlists
- **[Contributing](docs/contributing.md)** - Contribution guidelines

**📖 Full documentation index**: [\`docs/index.md\`](docs/index.md)

## 🏗️ Architecture

Clipper is built as a modern, scalable full-stack application:

\`\`\`
┌─────────────┐      ┌──────────────┐      ┌───────────────┐
│   Web App   │      │  Mobile Apps │      │   Backend API │
│  (React)    │─────▶│ (React Native)│─────▶│    (Go/Gin)   │
└─────────────┘      └──────────────┘      └───────┬───────┘
                                                    │
                          ┌─────────────────────────┼─────────────────────┐
                          ▼                         ▼                     ▼
                    ┌──────────┐            ┌──────────┐         ┌─────────────┐
                    │PostgreSQL│            │  Redis   │         │ OpenSearch  │
                    │  (Data)  │            │ (Cache)  │         │  (Search)   │
                    └──────────┘            └──────────┘         └─────────────┘
\`\`\`

- **Frontend**: React 19 + TypeScript + Vite + TailwindCSS
- **Mobile**: React Native 0.76 + Expo 52
- **Backend**: Go 1.24 + Gin + PostgreSQL + Redis
- **Search**: OpenSearch 2.11 with hybrid BM25 + vector search
- **Infrastructure**: Docker, Kubernetes, GitHub Actions

See [Architecture Documentation](docs/backend/architecture.md) for details.

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

### Mobile (iOS/Android)
- **Framework**: React Native 0.76
- **Platform**: Expo 52 with Expo Router
- **Language**: TypeScript (shared types)
- **State**: TanStack Query + Zustand

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
4. Run linters and tests (\`make test\`)
5. Commit your changes (\`git commit -m 'Add amazing feature'\`)
6. Push to the branch (\`git push origin feature/amazing-feature\`)
7. Open a Pull Request

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

**Status**: Active Development | **Version**: v0.x (Pre-release) | **Target**: MVP Release Q2 2025
