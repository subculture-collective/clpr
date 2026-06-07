# Vault Integration for Clipper

This directory contains the files required to fetch runtime secrets for the Clipper backend and
build-time secrets for the frontend from HashiCorp Vault.

## Layout

- `config/clpr-backend-agent.hcl` – Vault agent configuration (AppRole auth + template rendering).
- `templates/backend.env.ctmpl` – Consul-template file that renders the environment variables consumed by the backend.
- `templates/frontend.env.ctmpl` – Consul-template file that renders Sentry/build environment variables for the frontend.
- `rendered/` – Output directory for the processed `backend.env` and `frontend.env` files (ignored by git).
- `approle/` – Placeholders for the AppRole `role_id` and `secret_id` files (ignored by git).
- `policies/clpr-backend.hcl` – Vault policy granting the backend read-only access to `kv/clpr/backend`.
- `policies/clpr-frontend.hcl` – Vault policy granting the frontend build read-only access to `kv/clpr/frontend`.

## Bootstrapping Steps

### Backend Setup

1. **Initialize and unseal Vault** (if you haven't already) and set `VAULT_ADDR=https://vault.subcult.tv` on the machine that
   runs the commands below.
2. **Enable KV v2** (run once):

   ```bash
   vault secrets enable -path=kv kv-v2
   ```

3. **Write the backend secret data** (replace the placeholder values with the real ones):

   ```bash
   vault kv put kv/clpr/backend \
     PORT=8080 \
     GIN_MODE=release \
     ENVIRONMENT=production \
     BASE_URL=https://clpr.tv \
     LOG_LEVEL=info \
     DB_HOST=postgres \
     DB_PORT=5432 \
     DB_USER=clpr \
     DB_PASSWORD='changeme' \
     DB_NAME=clpr_db \
     DB_SSLMODE=disable \
     REDIS_HOST=redis \
     REDIS_PORT=6379 \
     REDIS_PASSWORD='' \
     REDIS_DB=0 \
     TWITCH_CLIENT_ID='...' \
     TWITCH_CLIENT_SECRET='...' \
     TWITCH_REDIRECT_URI=https://clpr.tv/api/v1/auth/twitch/callback \
     CORS_ALLOWED_ORIGINS=https://clpr.tv \
     OPENSEARCH_URL=http://opensearch:9200 \
     OPENSEARCH_INSECURE_SKIP_VERIFY=false
   ```

4. **Create the backend policy**:

   ```bash
   vault policy write clpr-backend vault/policies/clpr-backend.hcl
   ```

5. **Create the backend AppRole** (run once):

   ```bash
   vault write auth/approle/role/clpr-backend \
     token_policies="clpr-backend" \
     token_ttl="24h" \
     token_max_ttl="72h" \
     secret_id_ttl="24h" \
     secret_id_num_uses=0
   ```

6. **Capture backend AppRole credentials** and place them in `vault/approle/`:

   ```bash
   vault read -field=role_id auth/approle/role/clpr-backend/role-id > vault/approle/role_id
   vault write -field=secret_id -f auth/approle/role/clpr-backend/secret-id > vault/approle/secret_id
   ```

   > Treat `role_id` and `secret_id` like passwords. The directory is git-ignored by default.

7. **(Optional) Rotate secrets** by re-running step 6 whenever you want to mint a new `secret_id`.

Once the files exist, `docker compose` (or the `clpr-prod` systemd unit) will start the `clpr-vault-agent`
sidecar, which writes `vault/rendered/backend.env`. The backend container waits for that file, sources it, and then
starts the API process with the injected configuration.

### Frontend Setup

The frontend build requires Sentry credentials at build time to upload sourcemaps and enable runtime observability.

1. **Write the frontend secret data**:

   ```bash
   vault kv put kv/clpr/frontend \
     VITE_SENTRY_ENABLED=true \
     VITE_SENTRY_DSN="https://your-dsn@o123.ingest.sentry.io/456" \
     VITE_SENTRY_ENVIRONMENT=production \
     VITE_SENTRY_RELEASE="$(git rev-parse --short HEAD)" \
     VITE_SENTRY_TRACES_SAMPLE_RATE=0.1 \
     SENTRY_AUTH_TOKEN="sntrys_your_auth_token" \
     SENTRY_ORG="your-org" \
     SENTRY_PROJECT="clpr-frontend" \
     SENTRY_RELEASE="$(git rev-parse --short HEAD)"
   ```

2. **Create the frontend policy**:

   ```bash
   vault policy write clpr-frontend vault/policies/clpr-frontend.hcl
   ```

3. **Create the frontend AppRole** (run once):

   ```bash
   vault write auth/approle/role/clpr-frontend \
     token_policies="clpr-frontend" \
     token_ttl="1h" \
     token_max_ttl="2h" \
     secret_id_ttl="24h" \
     secret_id_num_uses=0
   ```

4. **Capture frontend AppRole credentials**:

   ```bash
   vault read -field=role_id auth/approle/role/clpr-frontend/role-id > vault/approle/frontend_role_id
   vault write -field=secret_id -f auth/approle/role/clpr-frontend/secret-id > vault/approle/frontend_secret_id
   ```

5. **Build the frontend** using the script:

   ```bash
   cd frontend
   ./scripts/build-with-vault.sh
   ```

   The script will:
   - Authenticate with Vault using the AppRole credentials.
   - Render `frontend.env` from the template.
   - Export the environment variables for Vite and the Sentry plugin.
   - Run `npm run build` with sourcemap upload enabled.
   - Delete `.map` files after upload (per plugin configuration).

## Expected Keys

### Backend (`kv/clpr/backend`)

The secret at `kv/clpr/backend` mirrors every variable in `backend/.env.example` (plus a few infrastructure
helpers like `POSTGRES_PASSWORD`). The rendered template now includes:

- Core service settings (`PORT`, `BASE_URL`, `ENVIRONMENT`, etc.)
- Database/Redis credentials (`DB_*`, `REDIS_*`, `POSTGRES_PASSWORD`)
- JWT key pair placeholders (`JWT_PRIVATE_KEY`, `JWT_PUBLIC_KEY`)
- Twitch, Stripe, Sentry, SendGrid, and OpenAI credentials
- Feature flags and scheduler knobs (`FEATURE_*`, `HOT_CLIPS_*`, `WEBHOOK_*`, etc.)

Run the following to review what is currently stored:

```bash
export VAULT_ADDR=https://vault.subcult.tv
vault login
vault kv get kv/clpr/backend
```

To update individual values, use `vault kv patch kv/clpr/backend KEY=value`. For example:

```bash
vault kv patch kv/clpr/backend \
   STRIPE_SECRET_KEY="sk_live_xxx" \
   STRIPE_WEBHOOK_SECRET="whsec_primary" \
   STRIPE_WEBHOOK_SECRET_ALT="whsec_secondary" \
   STRIPE_WEBHOOK_SECRETS="whsec_old1,whsec_old2" \
   SENTRY_DSN="https://example@o123.ingest.sentry.io/456"
```

### Frontend (`kv/clpr/frontend`)

The secret at `kv/clpr/frontend` contains Sentry credentials for runtime SDK and build-time sourcemap upload:

- `VITE_SENTRY_ENABLED` – Enable Sentry SDK (true/false)
- `VITE_SENTRY_DSN` – Sentry DSN for error reporting
- `VITE_SENTRY_ENVIRONMENT` – Environment name (production, staging, etc.)
- `VITE_SENTRY_RELEASE` – Release identifier (git SHA, version tag, or timestamp)
- `VITE_SENTRY_TRACES_SAMPLE_RATE` – Sample rate for performance traces (0.0 to 1.0)
- `SENTRY_AUTH_TOKEN` – Auth token for sourcemap upload
- `SENTRY_ORG` – Sentry organization slug
- `SENTRY_PROJECT` – Sentry project slug
- `SENTRY_RELEASE` – Optional override for upload release (defaults to `VITE_SENTRY_RELEASE`)

To update frontend secrets:

```bash
vault kv patch kv/clpr/frontend \
  VITE_SENTRY_DSN="https://new-dsn@sentry.io/123" \
  SENTRY_AUTH_TOKEN="sntrys_new_token"
```

Vault is now the only source of truth for backend and frontend secrets—do **not** maintain production `.env` files anymore.
