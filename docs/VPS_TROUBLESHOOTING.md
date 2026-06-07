# VPS Deployment Troubleshooting Guide

This guide covers common issues when deploying Clipper to a VPS and how to resolve them.

## Quick Diagnostics

Run the verification script first:

```bash
./scripts/verify-vps-deployment.sh
```

This will identify most common issues automatically.

## Common Issues

### 1. Vault Agent Cannot Connect to Vault

**Symptoms:**
- `clpr-vault-agent` logs show "connection refused" or "no such host: vault"
- Backend containers wait forever for secrets
- Secrets files not appearing in `/vault-agent/rendered/`

**Diagnosis:**
```bash
# Check if Vault is running
docker ps | grep vault

# Check vault-agent logs
docker logs clpr-vault-agent | tail -50

# Try to reach Vault from vault-agent container
docker exec clpr-vault-agent wget -qO- http://vault:8200/v1/sys/health
```

**Solutions:**

**Option 1: Connect Vault to shared networks**
```bash
# Get the actual Vault container name
VAULT_CONTAINER=$(docker ps --format '{{.Names}}' | grep vault | head -1)

# Connect Vault to web network
docker network connect web "$VAULT_CONTAINER"

# Restart vault-agent to retry connection
docker restart clpr-vault-agent
```

**Option 2: Update VAULT_ADDR if Vault container has different name**
```bash
# Find Vault container name
docker ps | grep vault

# If it's not named 'vault', update docker-compose.vps.yml:
# Change VAULT_ADDR: http://vault:8200
# To:    VAULT_ADDR: http://<actual-vault-name>:8200

# Then restart
docker compose -f docker-compose.vps.yml restart vault-agent
```

**Option 3: Use host networking (last resort)**
If Vault is not in Docker, update `docker-compose.vps.yml`:
```yaml
vault-agent:
  environment:
    VAULT_ADDR: http://host.docker.internal:8200  # For Docker Desktop
    # OR
    VAULT_ADDR: http://172.17.0.1:8200  # For Linux (Docker host IP)
```

---

### 2. Secrets Not Rendering

**Symptoms:**
- `/vault-agent/rendered/backend.env` or `postgres.env` files are empty or missing
- Containers wait forever at "waiting for Vault secrets..."

**Diagnosis:**
```bash
# Check if vault-agent is running
docker ps | grep vault-agent

# Check vault-agent logs for errors
docker logs -f clpr-vault-agent

# Check if AppRole files exist
ls -l vault/approle/

# Check file permissions
docker exec clpr-vault-agent ls -lh /vault-agent/rendered/
```

**Solutions:**

**Missing AppRole credentials:**
```bash
# Generate new AppRole credentials
export VAULT_ADDR=http://localhost:8200
vault login

vault read -field=role_id auth/approle/role/clpr-backend/role-id > vault/approle/role_id
vault write -field=secret_id -f auth/approle/role/clpr-backend/secret-id > vault/approle/secret_id

# Restart vault-agent
docker compose -f docker-compose.vps.yml restart vault-agent
```

**Wrong AppRole role name:**
Check if the role exists:
```bash
vault list auth/approle/role
```

If `clpr-backend` is not listed, create it:
```bash
vault write auth/approle/role/clpr-backend \
  token_policies="clpr-backend" \
  token_ttl="24h" \
  token_max_ttl="72h"
```

**Secrets not in Vault:**
```bash
# Check if secrets exist
vault kv get kv/clpr/backend
vault kv get kv/clpr/postgres

# If missing, add them (see docs/VPS_DEPLOYMENT.md)
```

**Template syntax errors:**
```bash
# Check template files
cat vault/templates/backend.env.ctmpl
cat vault/templates/postgres.env.ctmpl

# Look for syntax errors in the Consul Template format
```

---

### 3. Port Conflicts

**Symptoms:**
- "port is already allocated" error when starting containers
- Cannot start postgres, redis, or other services

**Diagnosis:**
```bash
# Check what's using port 5436 (Postgres)
sudo lsof -i :5436
# OR
netstat -tulpn | grep 5436

# Check for other clpr projects
docker ps -a | grep clpr
```

**Solutions:**

**Change Postgres port:**
Edit `docker-compose.vps.yml`:
```yaml
postgres:
  ports:
    - '5437:5432'  # Use different external port
```

**Stop conflicting containers:**
```bash
# List all running containers
docker ps

# Stop specific container
docker stop <container-name>

# Or stop all clpr containers from other projects
docker ps -a --filter "name=clpr" --format "{{.Names}}" | xargs docker stop
```

---

### 4. Caddy Cannot Reach Backend/Frontend

**Symptoms:**
- 502 Bad Gateway on clpr.tv
- Caddy logs show "dial tcp: lookup clpr-backend: no such host"
- Website doesn't load

**Diagnosis:**
```bash
# Check if Caddy is running
docker ps | grep caddy

# Check which network Caddy is on
docker inspect <caddy-container> | grep -A 10 Networks

# Check if backend/frontend are on web network
docker network inspect web | grep -E "(clpr-backend|clpr-frontend)"

# Try to reach backend from Caddy
CADDY_CONTAINER=$(docker ps --format '{{.Names}}' | grep caddy | head -1)
docker exec "$CADDY_CONTAINER" ping clpr-backend
docker exec "$CADDY_CONTAINER" wget -qO- http://clpr-backend:8080/api/v1/health
```

**Solutions:**

**Connect containers to web network:**
```bash
docker network connect web clpr-backend
docker network connect web clpr-frontend

# Reload Caddy
docker exec <caddy-container> caddy reload --config /etc/caddy/Caddyfile
```

**Check container names in Caddyfile:**
Ensure Caddyfile uses correct container names (`clpr-backend`, not `backend`):
```bash
grep -E "clpr-backend|clpr-frontend" Caddyfile
```

**Restart Caddy:**
```bash
docker restart <caddy-container>
```

---

### 5. Database Migration Failures

**Symptoms:**
- Deploy script fails at migration step
- Error: "migration failed"
- Cannot connect to database

**Diagnosis:**
```bash
# Check if Postgres is running and healthy
docker exec clpr-postgres pg_isready -U clpr -d clpr_db

# Check Postgres logs
docker logs clpr-postgres | tail -50

# Check if password is set correctly in Vault
vault kv get kv/clpr/postgres

# Try manual migration
docker run --rm \
  --network container:clpr-postgres \
  --volumes-from clpr-vault-agent \
  -v "$PWD/backend/migrations:/migrations:ro" \
  --entrypoint /bin/sh migrate/migrate:latest \
  -c 'set -a; . /vault-agent/rendered/postgres.env; set +a; \
      echo "Password: ${POSTGRES_PASSWORD:0:4}..."; \
      migrate -path /migrations \
              -database "postgresql://clpr:${POSTGRES_PASSWORD}@localhost:5432/clpr_db?sslmode=disable" -verbose up'
```

**Solutions:**

**Wrong password:**
```bash
# Update password in Vault
vault kv patch kv/clpr/postgres POSTGRES_PASSWORD='new-secure-password'

# Restart vault-agent and postgres
docker compose -f docker-compose.vps.yml restart vault-agent
sleep 5
docker compose -f docker-compose.vps.yml restart postgres
```

**Database not ready:**
```bash
# Wait for Postgres to be fully ready
sleep 10

# Then retry migration
./scripts/deploy-vps.sh --skip-git --skip-migrations
# Run migrations manually after confirming Postgres is ready
```

**Migration already applied:**
This is usually safe to ignore. To check migration status:
```bash
docker run --rm \
  --network container:clpr-postgres \
  --volumes-from clpr-vault-agent \
  -v "$PWD/backend/migrations:/migrations:ro" \
  --entrypoint /bin/sh migrate/migrate:latest \
  -c 'set -a; . /vault-agent/rendered/postgres.env; set +a; \
      migrate -path /migrations \
              -database "postgresql://clpr:${POSTGRES_PASSWORD}@localhost:5432/clpr_db?sslmode=disable" version'
```

---

### 6. Backend Not Starting or Unhealthy

**Symptoms:**
- Backend container exits immediately
- Health checks fail
- Backend shows as unhealthy in `docker ps`

**Diagnosis:**
```bash
# Check backend logs
docker logs clpr-backend | tail -100

# Check if secrets are loaded
docker exec clpr-backend env | grep -v PASSWORD | grep -v SECRET

# Try to access health endpoint manually
docker exec clpr-backend wget -qO- http://localhost:8080/api/v1/health
```

**Solutions:**

**Missing environment variables:**
```bash
# Check if backend.env was rendered
docker exec clpr-vault-agent cat /vault-agent/rendered/backend.env

# If empty or missing, check vault-agent logs
docker logs clpr-vault-agent

# Restart vault-agent
docker compose -f docker-compose.vps.yml restart vault-agent

# Wait for secrets, then restart backend
sleep 10
docker compose -f docker-compose.vps.yml restart backend
```

**Database connection issues:**
```bash
# Check if backend can reach postgres
docker exec clpr-backend getent hosts postgres

# Check if postgres is ready
docker exec clpr-postgres pg_isready -U clpr -d clpr_db

# Check database credentials in Vault match
vault kv get kv/clpr/backend | grep DB_
vault kv get kv/clpr/postgres
```

**JWT keys missing:**
```bash
# Check if JWT keys are set in Vault
vault kv get kv/clpr/backend | grep JWT

# If missing, generate and add them
cd backend
go run cmd/keygen/main.go

# Add to Vault (base64-encoded)
vault kv patch kv/clpr/backend \
  JWT_PRIVATE_KEY_B64="<base64-private-key>" \
  JWT_PUBLIC_KEY_B64="<base64-public-key>"
```

---

### 7. Frontend Not Loading

**Symptoms:**
- Frontend shows blank page
- 404 errors for static assets
- Console errors about API

**Diagnosis:**
```bash
# Check frontend logs
docker logs clpr-frontend | tail -50

# Check if frontend is serving files
docker exec clpr-frontend ls -l /usr/share/nginx/html/

# Check nginx configuration
docker exec clpr-frontend cat /etc/nginx/conf.d/default.conf
```

**Solutions:**

**Frontend not built correctly:**
```bash
# Rebuild frontend
cd frontend
npm run build

# Or rebuild container
docker compose -f docker-compose.vps.yml build frontend
docker compose -f docker-compose.vps.yml up -d frontend
```

**API connection issues:**
Check if frontend can reach backend:
```bash
# Frontend should proxy API calls through Caddy
# Check browser console for API errors

# Test API through Caddy
curl https://clpr.tv/api/v1/health
```

---

### 8. HTTPS/TLS Certificate Issues

**Symptoms:**
- Browser shows "Not secure" warning
- Certificate errors
- Cannot access site via HTTPS

**Diagnosis:**
```bash
# Check Caddy logs for cert errors
docker logs <caddy-container> | grep -i cert

# List current certificates
docker exec <caddy-container> caddy list-certificates

# Check if ports 80/443 are accessible from internet
curl -I http://clpr.tv
curl -I https://clpr.tv
```

**Solutions:**

**Ports not open:**
```bash
# Check firewall rules (UFW on Ubuntu)
sudo ufw status

# Allow ports 80 and 443
sudo ufw allow 80/tcp
sudo ufw allow 443/tcp
```

**DNS not pointing to server:**
```bash
# Check DNS
dig +short clpr.tv

# Should return your VPS IP address
```

**Force certificate renewal:**
```bash
# Delete existing certificates and restart Caddy
docker exec <caddy-container> rm -rf /data/caddy/certificates
docker restart <caddy-container>

# Caddy will automatically request new certificates
```

---

## General Debugging Commands

### View All Container Logs
```bash
# All clpr containers
docker ps --filter "name=clpr" --format "{{.Names}}" | xargs -I {} sh -c 'echo "=== {} ==="; docker logs --tail=20 {}'

# Specific container with follow
docker logs -f clpr-backend
```

### Check Resource Usage
```bash
# Container stats
docker stats clpr-backend clpr-frontend clpr-postgres clpr-redis

# Disk usage
docker system df
```

### Network Debugging
```bash
# List all networks
docker network ls

# Inspect web network
docker network inspect web

# Check what containers can reach each other (using getent for compatibility)
docker exec clpr-backend getent hosts postgres
docker exec clpr-backend getent hosts redis
docker exec clpr-backend getent hosts vault
```

### Reset Everything (Nuclear Option)
```bash
# Stop and remove all clpr containers
docker compose -f docker-compose.vps.yml down -v

# Remove images (optional)
docker images | grep clpr | awk '{print $3}' | xargs docker rmi -f

# Redeploy from scratch
./scripts/deploy-vps.sh
```

## Getting Help

If you're still stuck after trying these solutions:

1. Run the verification script and save output:
   ```bash
   ./scripts/verify-vps-deployment.sh > verification-output.txt
   ```

2. Collect logs:
   ```bash
   docker logs clpr-backend > backend.log 2>&1
   docker logs clpr-vault-agent > vault-agent.log 2>&1
   docker logs clpr-postgres > postgres.log 2>&1
   ```

3. Check configuration:
   ```bash
   docker compose -f docker-compose.vps.yml config > compose-config.yml
   ```

4. Share these files when asking for help.

## Prevention

To avoid these issues in the future:

1. **Always run verification after deployment:**
   ```bash
   ./scripts/deploy-vps.sh && ./scripts/verify-vps-deployment.sh
   ```

2. **Keep Vault secrets up to date:**
   ```bash
   # Document all required secrets
   vault kv get kv/clpr/backend
   vault kv get kv/clpr/postgres
   ```

3. **Monitor container health:**
   ```bash
   # Set up health check monitoring
   watch 'docker ps --filter "name=clpr" --format "table {{.Names}}\t{{.Status}}"'
   ```

4. **Backup before changes:**
   ```bash
   # Backup database before updates
   docker exec clpr-postgres pg_dump -U clpr clpr_db | gzip > backup-$(date +%Y%m%d).sql.gz
   ```
