# VPS Deployment Guide for Clipper (clpr.tv)

This guide covers deploying Clipper on a VPS where Vault and Caddy run as separate services.

## Prerequisites

Before deploying, ensure you have:

1. **Vault Server Running**
   - Located in `~/projects/vault`
   - Container named `vault` on Docker network
   - Accessible at `http://vault:8200` from Docker containers

2. **External Docker Network**
   - Network named `web` exists: `docker network create web`
   - Shared between Caddy and application containers

3. **Caddy Reverse Proxy** (optional for initial setup)
   - Will serve `clpr.tv` domain
   - Connected to `web` network
   - Configured to proxy to `clpr-backend` and `clpr-frontend`

4. **Vault Secrets Configured**
   - Backend secrets at `kv/clpr/backend`
   - Postgres password set in Vault
   - AppRole credentials generated

## Directory Structure

```
~/projects/
├── vault/          # External Vault server
├── caddy/          # External Caddy server (optional location)
├── clpr/        # This repository
└── systemd/        # Systemd service files (optional)
```

## Initial Setup

### 1. Vault Configuration

If Vault is not already configured, set it up:

```bash
# From ~/projects/vault or wherever Vault is running
export VAULT_ADDR=http://localhost:8200

# Login to Vault
vault login

# Enable KV v2 secrets engine (if not already enabled)
vault secrets enable -path=kv kv-v2

# Create backend secrets
vault kv put kv/clpr/backend \
  PORT=8080 \
  GIN_MODE=release \
  ENVIRONMENT=production \
  BASE_URL=https://clpr.tv \
  LOG_LEVEL=info \
  DB_HOST=postgres \
  DB_PORT=5432 \
  DB_USER=clpr \
  DB_PASSWORD='CHANGE_ME_SECURE_PASSWORD' \
  DB_NAME=clpr_db \
  DB_SSLMODE=disable \
  REDIS_HOST=redis \
  REDIS_PORT=6379 \
  REDIS_PASSWORD='' \
  REDIS_DB=0 \
  TWITCH_CLIENT_ID='your_twitch_client_id' \
  TWITCH_CLIENT_SECRET='your_twitch_client_secret' \
  TWITCH_REDIRECT_URI=https://clpr.tv/api/v1/auth/twitch/callback \
  CORS_ALLOWED_ORIGINS=https://clpr.tv \
  MFA_ENCRYPTION_KEY="$(openssl rand -base64 32)"

# Create Postgres secrets
vault kv put kv/clpr/postgres \
  POSTGRES_DB=clpr_db \
  POSTGRES_USER=clpr \
  POSTGRES_PASSWORD='CHANGE_ME_SECURE_PASSWORD'

# Create backend policy
vault policy write clpr-backend vault/policies/clpr-backend.hcl

# Create AppRole
vault write auth/approle/role/clpr-backend \
  token_policies="clpr-backend" \
  token_ttl="24h" \
  token_max_ttl="72h" \
  secret_id_ttl="24h" \
  secret_id_num_uses=0

# Generate AppRole credentials
cd ~/projects/clpr
mkdir -p vault/approle
vault read -field=role_id auth/approle/role/clpr-backend/role-id > vault/approle/role_id
vault write -field=secret_id -f auth/approle/role/clpr-backend/secret-id > vault/approle/secret_id
chmod 600 vault/approle/role_id vault/approle/secret_id
```

### 2. Create External Networks

```bash
# Create the shared 'web' network if it doesn't exist
docker network create web
```

### 3. Verify Vault Connectivity

The clpr containers need to reach the Vault container. Ensure Vault is on a network accessible to clpr services:

```bash
# Check which networks Vault is on
docker inspect vault | grep -A 10 Networks

# If needed, connect Vault to the web network
docker network connect web vault
```

## Deployment

### Quick Deploy

From `~/projects/clpr`:

```bash
# First time deployment
./scripts/deploy-vps.sh

# Update deployment (with git pull)
./scripts/deploy-vps.sh

# Deploy without git operations
./scripts/deploy-vps.sh --skip-git

# Deploy without migrations
./scripts/deploy-vps.sh --skip-migrations
```

### What the Deploy Script Does

1. **Environment Check**: Verifies Vault and network prerequisites
2. **Git Operations**: Optionally pulls latest code
3. **Network Setup**: Ensures `web` network exists
4. **Build Images**: Builds backend, frontend, and postgres containers
5. **Start Infrastructure**: Launches vault-agent, postgres, redis
6. **Wait for Secrets**: Ensures Vault agent renders secrets
7. **Run Migrations**: Applies database migrations
8. **Start Applications**: Launches backend and frontend
9. **Health Checks**: Verifies all services are healthy
10. **Caddy Verification**: Checks Caddy configuration

### Manual Deployment Steps

If you prefer manual control:

```bash
cd ~/projects/clpr

# 1. Build images
docker compose -f docker-compose.vps.yml build

# 2. Start infrastructure
docker compose -f docker-compose.vps.yml up -d vault-agent postgres redis

# 3. Wait for secrets to render (check logs)
docker logs -f clpr-vault-agent

# 4. Run migrations
docker run --rm \
  --network container:clpr-postgres \
  --volumes-from clpr-vault-agent \
  -v "$PWD/backend/migrations:/migrations:ro" \
  --entrypoint /bin/sh migrate/migrate:latest \
  -c 'set -e; set -a; . /vault-agent/rendered/postgres.env; set +a; \
      migrate -path /migrations \
              -database "postgresql://clpr:${POSTGRES_PASSWORD}@localhost:5432/clpr_db?sslmode=disable" up'

# 5. Start application services
docker compose -f docker-compose.vps.yml up -d backend frontend

# 6. Check status
docker compose -f docker-compose.vps.yml ps
```

## Caddy Configuration

### Option 1: Update Existing Caddy

If Caddy is already running (e.g., in `~/projects/caddy`):

1. **Connect Caddy to web network** (if not already):
   ```bash
   docker network connect web <caddy-container-name>
   ```

2. **Update Caddyfile** to include clpr routing:
   ```caddy
   clpr.tv {
       # Security headers
       header {
           Strict-Transport-Security "max-age=31536000; includeSubDomains; preload"
           X-Content-Type-Options "nosniff"
           X-Frame-Options "DENY"
           X-XSS-Protection "1; mode=block"
           Referrer-Policy "strict-origin-when-cross-origin"
           -Server
       }
       
       # API routes -> backend
       handle /api/* {
           reverse_proxy clpr-backend:8080 {
               header_up Host {host}
               header_up X-Real-IP {remote_host}
               header_up X-Forwarded-For {remote_host}
               header_up X-Forwarded-Proto {scheme}
               
               health_uri /api/v1/health
               health_interval 30s
               health_timeout 10s
               health_status 200
           }
       }
       
       # Health check endpoint
       handle /health {
           reverse_proxy clpr-backend:8080
       }
       
       # WebSocket support
       handle /ws/* {
           reverse_proxy clpr-backend:8080 {
               header_up Upgrade {>Upgrade}
               header_up Connection {>Connection}
           }
       }
       
       # Frontend (SPA)
       handle /* {
           reverse_proxy clpr-frontend:80
       }
       
       # Compression
       encode gzip
   }
   
   # HTTP to HTTPS redirect
   http://clpr.tv {
       redir https://{host}{uri} permanent
   }
   ```

3. **Reload Caddy**:
   ```bash
   docker exec <caddy-container> caddy reload --config /etc/caddy/Caddyfile
   ```

### Option 2: Deploy Caddy from Clipper Repo

The repository includes a Caddyfile that can be used:

```bash
cd ~/projects/clpr

# Start Caddy container
docker run -d \
  --name clpr-caddy \
  --network web \
  -p 80:80 \
  -p 443:443 \
  -v "$PWD/Caddyfile:/etc/caddy/Caddyfile:ro" \
  -v caddy_data:/data \
  -v caddy_config:/config \
  --restart unless-stopped \
  caddy:2-alpine
```

## Verification

### 1. Check Container Status

```bash
cd ~/projects/clpr
docker compose -f docker-compose.vps.yml ps
```

All services should show as "healthy" or "running".

### 2. Verify Network Connectivity

```bash
# Check which containers are on the web network
docker network inspect web | grep Name

# Should include: clpr-backend, clpr-frontend, caddy, vault (optional)
```

### 3. Test Backend Health

```bash
# Internal health check
docker exec clpr-backend wget -qO- http://localhost:8080/api/v1/health

# Through Caddy (if configured)
curl https://clpr.tv/api/v1/health
```

### 4. Test Frontend

```bash
# Visit in browser
https://clpr.tv
```

### 5. Check Logs

```bash
# Backend logs
docker logs -f clpr-backend

# Frontend logs
docker logs -f clpr-frontend

# Vault agent logs
docker logs -f clpr-vault-agent

# Postgres logs
docker logs -f clpr-postgres
```

## Troubleshooting

### Vault Agent Can't Connect to Vault

**Symptom**: Vault agent logs show "connection refused" or "no such host"

**Solution**:
```bash
# Check Vault is running
docker ps | grep vault

# Check Vault networks
docker inspect vault | grep -A 10 Networks

# Ensure Vault is reachable from clpr network
# Option 1: Connect Vault to web network
docker network connect web vault

# Option 2: Connect Vault to clpr-network
# Find your actual clpr network name first
docker network ls | grep clpr
# Then connect (replace <clpr-network> with actual name)
docker network connect <clpr-network> vault
```

### Secrets Not Rendering

**Symptom**: Backend waits forever for `/vault-agent/rendered/backend.env`

**Solution**:
```bash
# Check vault-agent logs
docker logs -f clpr-vault-agent

# Verify AppRole credentials exist
ls -l vault/approle/

# Test Vault connectivity from vault-agent container
docker exec clpr-vault-agent wget -qO- http://vault:8200/v1/sys/health

# If vault-agent can't resolve 'vault', update VAULT_ADDR in docker-compose.vps.yml
# to use the actual Vault container name or IP
```

### Port Conflicts

**Symptom**: "port already allocated" error

**Solution**:
```bash
# Check what's using port 5436 (Postgres)
sudo lsof -i :5436
netstat -tulpn | grep 5436

# If another project is using it, change the port in docker-compose.vps.yml:
# ports:
#   - '5437:5432'  # Use different external port
```

### Caddy Can't Reach Backend

**Symptom**: 502 Bad Gateway on clpr.tv

**Solution**:
```bash
# Ensure Caddy and clpr-backend are on same network
docker network inspect web

# If backend is not listed, something went wrong during deployment
# Redeploy or manually connect:
docker network connect web clpr-backend
docker network connect web clpr-frontend

# Reload Caddy
docker exec <caddy-container> caddy reload --config /etc/caddy/Caddyfile
```

### Database Migration Failures

**Symptom**: Migration step fails in deploy script

**Solution**:
```bash
# Check Postgres is healthy
docker exec clpr-postgres pg_isready -U clpr -d clpr_db

# Check Postgres password is correct in Vault
vault kv get kv/clpr/postgres

# Run migrations manually with verbose output
docker run --rm \
  --network container:clpr-postgres \
  --volumes-from clpr-vault-agent \
  -v "$PWD/backend/migrations:/migrations:ro" \
  --entrypoint /bin/sh migrate/migrate:latest \
  -c 'set -e; set -a; . /vault-agent/rendered/postgres.env; set +a; \
      echo "Using password: ${POSTGRES_PASSWORD:0:4}..."; \
      migrate -path /migrations \
              -database "postgresql://clpr:${POSTGRES_PASSWORD}@localhost:5432/clpr_db?sslmode=disable" -verbose up'
```

## Updating Deployment

### Code Updates

```bash
cd ~/projects/clpr
./scripts/deploy-vps.sh
```

This will:
- Pull latest code
- Rebuild images
- Apply new migrations
- Restart services with zero downtime (containers restart)

### Secrets Updates

```bash
# Update secrets in Vault
vault kv patch kv/clpr/backend \
  TWITCH_CLIENT_ID="new_value" \
  TWITCH_CLIENT_SECRET="new_secret"

# Restart vault-agent to fetch new secrets
cd ~/projects/clpr
docker compose -f docker-compose.vps.yml restart vault-agent

# Wait a few seconds for secrets to render, then restart backend
docker compose -f docker-compose.vps.yml restart backend
```

## Rollback

If a deployment fails:

```bash
cd ~/projects/clpr

# Check git history
git log --oneline -10

# Rollback to previous commit
git reset --hard <previous-commit-sha>

# Redeploy
./scripts/deploy-vps.sh --skip-git
```

## Monitoring

### Service Health

```bash
# Quick health check
docker compose -f docker-compose.vps.yml ps

# Detailed health for all services
docker ps --filter "name=clpr" --format "table {{.Names}}\t{{.Status}}"
```

### Logs

```bash
# Follow all clpr logs
docker compose -f docker-compose.vps.yml logs -f

# Specific service
docker logs -f clpr-backend
```

### Resource Usage

```bash
# Container resource usage
docker stats clpr-backend clpr-frontend clpr-postgres clpr-redis
```

## Backup and Recovery

### Database Backup

```bash
# Create backup
docker exec clpr-postgres pg_dump -U clpr clpr_db | gzip > ~/backups/clpr-db-$(date +%Y%m%d-%H%M%S).sql.gz

# Restore from backup
gunzip < ~/backups/clpr-db-TIMESTAMP.sql.gz | docker exec -i clpr-postgres psql -U clpr clpr_db
```

### Volume Backup

```bash
# Backup postgres data volume
docker run --rm \
  -v clpr_postgres_data:/data:ro \
  -v ~/backups:/backup \
  alpine \
  tar czf /backup/postgres-data-$(date +%Y%m%d-%H%M%S).tar.gz -C /data .
```

## Security Notes

1. **Secrets Management**: All secrets are in Vault, never committed to git
2. **AppRole Credentials**: Stored in `vault/approle/` which is git-ignored
3. **Network Isolation**: Internal services use private network, only Caddy is public
4. **Port Exposure**: Only Postgres port 5436 is exposed for admin access
5. **TLS**: Caddy handles automatic HTTPS with Let's Encrypt

## Support

For issues or questions:
- Check container logs: `docker logs -f clpr-backend`
- Review Vault agent logs: `docker logs -f clpr-vault-agent`
- Inspect network: `docker network inspect web`
- Test connectivity: `docker exec clpr-backend getent hosts postgres`
