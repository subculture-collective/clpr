---
title: Development Setup Verification Checklist
summary: Verifies that all required development tools are correctly installed and configured.
tags: ["testing", "archive", "implementation"]
area: docs
status: archived
owner: team-core
version: "1.0"
last_reviewed: 2026-01-29
---

# Development Setup Verification Checklist

## Run this checklist to verify everything is configured correctly

```bash
#!/bin/bash

echo "=== Clipper Development Setup Verification ==="
echo ""

# 1. Check tools
echo "1. Checking required tools..."
go version > /dev/null && echo "  ✓ Go installed" || echo "  ✗ Go NOT installed"
node --version > /dev/null && echo "  ✓ Node installed" || echo "  ✗ Node NOT installed"
npm --version > /dev/null && echo "  ✓ npm installed" || echo "  ✗ npm NOT installed"
docker --version > /dev/null && echo "  ✓ Docker installed" || echo "  ✗ Docker NOT installed"
git --version > /dev/null && echo "  ✓ Git installed" || echo "  ✗ Git NOT installed"
echo ""

# 2. Check Docker services
echo "2. Checking Docker services..."
docker compose ps | grep postgres > /dev/null && echo "  ✓ PostgreSQL running" || echo "  ✗ PostgreSQL NOT running"
docker compose ps | grep redis > /dev/null && echo "  ✓ Redis running" || echo "  ✗ Redis NOT running"
echo ""

# 3. Check database connectivity
echo "3. Checking database connections..."
psql -h 127.0.0.1 -p 5436 -U clpr -d clpr_db -c "SELECT 1" > /dev/null 2>&1 && echo "  ✓ PostgreSQL connected" || echo "  ✗ PostgreSQL NOT connected"
docker exec clpr-redis redis-cli PING > /dev/null 2>&1 && echo "  ✓ Redis connected" || echo "  ✗ Redis NOT connected"
echo ""

# 4. Check environment files
echo "4. Checking configuration files..."
test -f backend/.env && echo "  ✓ backend/.env exists" || echo "  ✗ backend/.env missing"
test -f frontend/.env && echo "  ✓ frontend/.env exists" || echo "  ✗ frontend/.env missing"
echo ""

# 5. Check git setup
echo "5. Checking git setup..."
git remote -v | grep origin > /dev/null && echo "  ✓ Git remote configured" || echo "  ✗ Git remote NOT configured"
git branch | grep -q main && echo "  ✓ main branch exists" || echo "  ✗ main branch NOT found"
echo ""

# 6. Check source code
echo "6. Checking source code..."
test -f backend/cmd/api/main.go && echo "  ✓ Backend code exists" || echo "  ✗ Backend code missing"
test -f frontend/src/main.tsx && echo "  ✓ Frontend code exists" || echo "  ✗ Frontend code missing"
test -f backend/go.mod && echo "  ✓ Go modules configured" || echo "  ✗ Go modules NOT configured"
test -f frontend/package.json && echo "  ✓ npm configured" || echo "  ✗ npm NOT configured"
echo ""

echo "=== Verification Complete ==="
echo ""
echo "If everything shows ✓, you're ready to develop!"
echo ""
echo "To start development:"
echo "  Terminal 1: docker compose up -d postgres redis"
echo "  Terminal 2: cd backend && go run cmd/api/main.go"
echo "  Terminal 3: cd frontend && npm run dev"
echo "  Browser: http://localhost:5173"
```

## Files Checklist

| File | Purpose | Status |
|------|---------|--------|
| backend/.env | Backend config | ✓ Configured |
| frontend/.env | Frontend config | ✓ Configured |
| docker-compose.yml | DB + Redis setup | ✓ Ready |
| Makefile | Development commands | ✓ Available |
| DEVELOPMENT_READY.md | Quick start guide | ✓ Created |
| DEVELOPMENT_PLAN.md | Complete dev guide | ✓ Created |
| DEV_STATUS_AND_PLAN.md | Status & next steps | ✓ Created |
| DEPLOYMENT_IMPROVEMENTS.md | Deployment overview | ✓ Created |
| QUICK_REFERENCE.md | Command reference | ✓ Created |

## Services Status

| Service | Port | Status |
|---------|------|--------|
| PostgreSQL | 5436 | ✓ Running |
| Redis | 6379 | ✓ Running |
| Vault | 8200 | ✓ Running |
| Backend | 8080 | Ready (start manually) |
| Frontend | 5173 | Ready (start manually) |
| Monitoring | Various | ✓ Running (Prometheus, Grafana, etc.) |

## Next Steps

1. Run verification script (above)
2. Fix any issues found
3. Start development services (see quick start)
4. Begin coding!

## Support

- Questions? Check the documentation files
- Issues? See troubleshooting sections
- Need help? Ask team on Slack/Discord
