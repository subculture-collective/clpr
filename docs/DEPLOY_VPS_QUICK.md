# Clipper VPS Deployment - Quick Reference

## One-Command Deploy

```bash
cd ~/projects/clpr && ./scripts/deploy-vps.sh
```

## Deploy Options

```bash
# Deploy without git pull
./scripts/deploy-vps.sh --skip-git

# Deploy without git operations at all  
./scripts/deploy-vps.sh --skip-git --no-pull

# Deploy without running migrations
./scripts/deploy-vps.sh --skip-migrations

# Deploy with all options
./scripts/deploy-vps.sh --skip-git --skip-migrations
```

## Prerequisites (One-Time Setup)

```bash
# 1. Create web network
docker network create web

# 2. Ensure Vault is running
docker ps | grep vault

# 3. Connect Vault to web network (if needed)
docker network connect web vault

# 4. Generate Vault AppRole credentials
cd ~/projects/clpr
vault read -field=role_id auth/approle/role/clpr-backend/role-id > vault/approle/role_id
vault write -field=secret_id -f auth/approle/role/clpr-backend/secret-id > vault/approle/secret_id
```

## Post-Deploy: Configure Caddy

```bash
# If Caddy is already running, reload it
docker exec <caddy-container-name> caddy reload --config /etc/caddy/Caddyfile

# Or restart Caddy to pick up new configuration
docker restart <caddy-container-name>
```

## Verify Deployment

```bash
# 1. Check all containers are healthy
docker compose -f docker-compose.vps.yml ps

# 2. Test backend health
docker exec clpr-backend wget -qO- http://localhost:8080/api/v1/health

# 3. Test via Caddy
curl https://clpr.tv/api/v1/health

# 4. Visit in browser
https://clpr.tv
```

## Common Issues

### Vault agent can't connect
```bash
# Find the actual clpr network name
# It will be based on your project directory name (e.g., clpr_clpr-network)
docker network ls | grep clpr

# Connect Vault to networks (replace <clpr-network> with actual name)
docker network connect web vault
docker network connect <clpr-network> vault

# Restart vault-agent
docker compose -f docker-compose.vps.yml restart vault-agent
```

### Secrets not rendering
```bash
# Check vault-agent logs
docker logs -f clpr-vault-agent

# Verify AppRole files exist
ls -l vault/approle/role_id vault/approle/secret_id

# Test Vault connectivity
docker exec clpr-vault-agent wget -qO- http://vault:8200/v1/sys/health
```

### Caddy shows 502
```bash
# Ensure backend is on web network
docker network inspect web | grep clpr-backend

# If not, connect it
docker network connect web clpr-backend
docker network connect web clpr-frontend

# Reload Caddy
docker exec <caddy-container> caddy reload --config /etc/caddy/Caddyfile
```

## Service Management

```bash
# View logs
docker logs -f clpr-backend
docker logs -f clpr-frontend
docker logs -f clpr-vault-agent

# Restart a service
docker compose -f docker-compose.vps.yml restart backend

# Stop all services
docker compose -f docker-compose.vps.yml down

# Start all services
docker compose -f docker-compose.vps.yml up -d

# View resource usage
docker stats clpr-backend clpr-frontend clpr-postgres
```

## Emergency Rollback

```bash
cd ~/projects/clpr

# Find previous commit
git log --oneline -10

# Rollback
git reset --hard <commit-sha>

# Redeploy
./scripts/deploy-vps.sh --skip-git
```

## Update Secrets

```bash
# Update in Vault
vault kv patch kv/clpr/backend KEY=new_value

# Restart vault-agent
docker compose -f docker-compose.vps.yml restart vault-agent

# Wait 5 seconds, then restart backend
sleep 5
docker compose -f docker-compose.vps.yml restart backend
```

## Troubleshooting Commands

```bash
# Check what's on web network
docker network inspect web --format '{{range .Containers}}{{.Name}} {{end}}'

# Check vault-agent rendered files
docker exec clpr-vault-agent ls -lh /vault-agent/rendered/

# Test postgres connectivity
docker exec clpr-postgres pg_isready -U clpr -d clpr_db

# Test backend can resolve and reach postgres
docker exec clpr-backend wget -qO- --timeout=5 http://postgres:5432 2>&1 || echo "Can resolve postgres"
docker exec clpr-backend getent hosts postgres

# Test backend can resolve and reach redis  
docker exec clpr-backend wget -qO- --timeout=5 http://redis:6379 2>&1 || echo "Can resolve redis"
docker exec clpr-backend getent hosts redis

# See all environment variables (sanitized)
docker exec clpr-backend env | grep -v PASSWORD | grep -v SECRET | grep -v KEY
```

## Full Documentation

See `docs/VPS_DEPLOYMENT.md` for complete guide.
