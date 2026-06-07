---
title: Clipper Development Plan
summary: Development plan documenting architecture, service setup, and implementation strategy for the project.
tags: ["archive", "implementation"]
area: docs
status: archived
owner: team-core
version: "1.0"
last_reviewed: 2026-01-29
---


# Clipper Development Plan

## Current System Status

### Running Services

- ✅ PostgreSQL (clpr-postgres:5436) - HEALTHY
- ✅ Redis (clpr-redis:6379) - HEALTHY
- ✅ Frontend (nginx:80) - UNHEALTHY (disconnected from backend)
- ❌ Backend (port 8080) - RESTARTING (connection issue fixed)
- ✅ Vault Agent - running for secrets management
- ✅ Monitoring Stack - Prometheus, Grafana, Loki, AlertManager

## Architecture

```
Frontend Dev Server (Port 5173)
        │
        ├──> API calls
        ▼
Backend API (Port 8080)
        │
        ├──> Database queries
        ▼
PostgreSQL + Redis
```

## Development Workflows

### Workflow 1: Local Development (Recommended for Features)

**Use case:** Building new features, testing locally, no VPS involvement

```bash
# 1. Create feature branch
git checkout -b feature/my-awesome-feature develop

# 2. Start local development environment
make docker-up       # PostgreSQL + Redis only
make backend-dev     # Terminal 2: Backend on port 8080
make frontend-dev    # Terminal 3: Frontend dev server on port 5173

# 3. Open http://localhost:5173 in browser

# 4. Make changes, test with hot reload

# 5. Run tests
make test-unit
make test            # All tests

# 6. Commit and push
git add .
git commit -m "feat: description"
git push origin feature/my-awesome-feature

# 7. Create PR: feature/my-awesome-feature -> develop
# (Auto-deploys to staging)

# 8. After review, merge to develop
```

### Workflow 2: Staging Testing

**Use case:** Testing on a real server before production

```bash
# Automatic process:
# 1. Merge PR to develop branch
# 2. GitHub Actions auto-builds Docker images
# 3. Staging environment auto-deploys
# 4. Team tests on staging.your-domain.com

# If issues found:
git checkout -b hotfix/issue develop
# ... fix and test locally ...
git push origin hotfix/issue
# Create PR to develop
# Merges and re-deploys staging
```

### Workflow 3: Production Deployment

**Use case:** Releasing to production users

```bash
# 1. Promote from develop to main
git checkout main
git pull origin main
git merge develop
git push origin main

# 2. GitHub Actions builds + deploys
# 3. Manual approval in GitHub (if configured)
# 4. Production auto-deploys with zero downtime (blue-green)

# Alternative: Manual blue-green deploy
bash scripts/deploy-blue-green.sh
```

## Environment Setup

### Prerequisites

- Go 1.24+ (check: `go version`)
- Node 20+ (check: `node --version`)
- npm 11+ (check: `npm --version`)
- Docker & Docker Compose (check: `docker ps`)
- Git (check: `git --version`)

### First-Time Setup

```bash
cd /home/onnwee/projects/clpr

# 1. Install dependencies
make install

# 2. Create/verify .env files
# backend/.env - already configured
# frontend/.env - check values

# 3. Verify environment variables
cat backend/.env
cat frontend/.env

# 4. Start local services
make docker-up
sleep 10  # Wait for DB

# 5. Run migrations (if needed)
make migrate-up

# 6. Test connection
make backend-dev
# Should see: "ListenAndServe on :8080"
```

## Development Commands

### Database

```bash
make migrate-up              # Run pending migrations
make migrate-down            # Rollback last migration
make migrate-status          # Check current version
make migrate-create NAME=foo # Create new migration
make migrate-seed            # Seed with test data
```

### Testing

```bash
make test                # All tests
make test-unit           # Unit tests only
make test-coverage        # With coverage report
make test-load-mixed      # Load testing
```

### Building

```bash
make build              # Build backend + frontend
make backend-build      # Build backend only
make frontend-build     # Build frontend only
make lint               # Run linters
```

### Development

```bash
make dev               # Start everything (backend + frontend)
make backend-dev       # Backend with hot reload
make frontend-dev      # Frontend with hot reload
make docker-up         # PostgreSQL + Redis only
make docker-down       # Stop containers
```

## Git Workflow

### Branch Strategy

```
main (production)
  │
  ├── (PR) ← develop (staging)
  │            │
  │            └── (PR) ← feature/my-feature
  │                          └── Your work
```

**Rules:**
- Never push directly to `main` - always PR
- PR to `develop` requires passing CI/CD
- Feature branches are your sandbox
- Use descriptive branch names: `feature/add-voting`, `fix/chunking-issue`, `docs/api-guide`

### Commit Messages

```bash
# Good commits
git commit -m "feat: add clip voting feature"
git commit -m "fix: resolve database connection timeout"
git commit -m "docs: update deployment guide"
git commit -m "refactor: simplify auth middleware"
git commit -m "test: add unit tests for voting service"

# Bad commits
git commit -m "update stuff"
git commit -m "wip"
git commit -m "asdf"
```

### Pull Requests

1. **Create PR with clear title**
   ```
   Title: "feat: add clip voting feature"
   Description:
   - Adds upvote/downvote buttons to clip cards
   - Persists votes to database
   - Counts votes in clip feed listing
   - Fixes #123
   ```

2. **Request review** from team members

3. **Address feedback** in new commits (don't force push)

4. **Merge** using "Squash and merge" for clean history

5. **Delete branch** after merge

## Common Development Tasks

### Adding a New API Endpoint

```bash
# 1. Create feature branch
git checkout -b feature/new-endpoint develop

# 2. Edit backend code
# File: backend/internal/handlers/clips.go
# File: backend/internal/services/clips.go
# File: backend/internal/repository/clips.go

# 3. Create migration if needed
make migrate-create NAME=add_new_column

# 4. Test locally
make backend-dev
# Test with: curl http://localhost:8080/api/v1/new-endpoint

# 5. Run tests
make test

# 6. Update API docs
# File: docs/API.md

# 7. Commit and push
git add .
git commit -m "feat: add new API endpoint"
git push origin feature/new-endpoint

# 8. Create PR to develop
```

### Adding Frontend Component

```bash
# 1. Create feature branch
git checkout -b feature/new-component develop

# 2. Start dev server
make frontend-dev

# 3. Edit frontend code
# File: frontend/src/components/NewComponent.tsx
# File: frontend/src/pages/SomePage.tsx

# 4. Test with hot reload (automatic)

# 5. Test different screen sizes
# Chrome DevTools: F12 > Toggle device toolbar

# 6. Run tests
npm test  # in frontend/

# 7. Commit and push
git add frontend/src
git commit -m "feat: add new component"
git push origin feature/new-component

# 8. Create PR to develop
```

### Debugging Issues

```bash
# Backend
docker logs clpr-backend -f --tail 100
docker logs clpr-postgres -f --tail 100

# Frontend
npm run dev -- --debug  # In frontend/

# Database
psql -h localhost -p 5436 -U clpr -d clpr_db
> SELECT * FROM clips LIMIT 5;

# Redis
redis-cli -p 6379
> KEYS *
> GET some_key

# Network requests
# Browser DevTools: F12 > Network tab
# Look for 404s, 500s, slow requests
```

## Troubleshooting

### Backend won't connect to database

```bash
# Check backend environment
cat backend/.env

# Should have:
DB_HOST=clpr-postgres    # NOT localhost
DB_PORT=5432                # NOT 5436 (5436 is host port)
DB_NAME=clpr_db

# Verify PostgreSQL is running and healthy
docker ps | grep postgres
# Should see: UP (healthy)

# Check connection directly
psql -h 127.0.0.1 -p 5436 -U clpr -d clpr_db
```

### Frontend can't reach backend

```bash
# Check frontend environment
cat frontend/.env

# Should have:
VITE_API_URL=http://localhost:8080

# Test backend is running
curl http://localhost:8080/health

# Check CORS settings in backend/.env
CORS_ALLOWED_ORIGINS=http://localhost:5173

# If using Docker, verify network
docker network inspect clpr-network
```

### Hot reload not working

```bash
# Restart frontend dev server
# In terminal running: make frontend-dev
# Press Ctrl+C to stop
# Run again: make frontend-dev

# Check file watchers
# Issue: too many open files
# Solution: increase limit
ulimit -n 4096
```

### Tests failing

```bash
# Run with verbose output
cd backend && go test -v ./...

# Run specific test
cd backend && go test -run TestFunctionName ./...

# Check test database
docker compose -f docker-compose.test.yml up -d
docker logs -f clpr-postgres-test
```

## Performance Optimization

### Frontend Bundle Size

```bash
cd frontend
npm run build
npm run analyze  # Visual breakdown
# Look for:
# - Large dependencies
# - Duplicate code
# - Opportunities for code splitting
```

### Database Queries

```bash
# Enable slow query logging
# In backend logs, look for slow queries
docker logs clpr-backend | grep "duration"

# Optimize with indexes
make migrate-create NAME=add_clips_index
# Edit migration to add index
make migrate-up
```

### Caching Strategy

- Frontend: Browser caching for static assets
- Backend: Redis for session data, expensive queries
- Database: Query result caching via Redis

## Monitoring in Development

### Logs

```bash
docker logs -f clpr-postgres   # DB logs
docker logs -f clpr-redis      # Cache logs
make backend-dev                  # Backend logs in terminal
```

### Health Checks

```bash
bash scripts/health-check.sh

# Or individual checks
curl http://localhost:8080/health
curl http://localhost:80/health.html
```

### Metrics (Staging/Production)

- Grafana: <http://your-domain:3000>
- Prometheus: <http://your-domain:9090>
- AlertManager: <http://your-domain:9093>

## Next Steps

1. **Verify setup works**
   ```bash
   make docker-up
   sleep 5
   docker ps | grep clpr
   ```

2. **Create first feature branch**
   ```bash
   git checkout -b feature/your-feature develop
   ```

3. **Start development**
   ```bash
   make backend-dev   # Terminal 1
   make frontend-dev  # Terminal 2
   ```

4. **Make changes and test**

5. **Push and create PR**
   ```bash
   git push origin feature/your-feature
   # Create PR on GitHub
   ```

6. **Merge to develop** (after review and CI passes)

7. **Staging tests** (auto-deploys)

8. **Promote to main** when ready for production

## Resources

- [Development Workflow Guide](docs/development-workflow.md)
- [Deployment Guide](docs/deployment-live-development.md)
- [API Documentation](docs/API.md)
- [Architecture](docs/ARCHITECTURE.md)
- [Contributing Guide](CONTRIBUTING.md)
