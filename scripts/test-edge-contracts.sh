#!/usr/bin/env bash

set -euo pipefail

repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$repo_root"

runtime=false
if [[ "${1:-}" == "--runtime" ]]; then
  runtime=true
elif [[ $# -gt 0 ]]; then
  echo "usage: $0 [--runtime]" >&2
  exit 2
fi

fail() {
  echo "edge contract: $*" >&2
  exit 1
}

require_literal() {
  local file="$1"
  local value="$2"
  grep -Fq "$value" "$file" || fail "$file is missing: $value"
}

for header in \
  'Strict-Transport-Security "max-age=31536000; includeSubDomains; preload"' \
  'Cross-Origin-Opener-Policy "same-origin"' \
  'Cross-Origin-Resource-Policy "same-origin"' \
  'X-Content-Type-Options "nosniff"' \
  'X-Frame-Options "DENY"' \
  'Referrer-Policy "strict-origin-when-cross-origin"' \
  'Permissions-Policy "geolocation=(), microphone=(), camera=(), payment=()"'; do
  require_literal Caddyfile "$header"
  require_literal deploy/Caddyfile.blue-green.template "$header"
done

for header in \
  'Strict-Transport-Security "max-age=31536000; includeSubDomains; preload" always' \
  'Cross-Origin-Opener-Policy "same-origin" always' \
  'Cross-Origin-Resource-Policy "same-origin" always' \
  'X-Content-Type-Options "nosniff" always' \
  'X-Frame-Options "DENY" always' \
  'Referrer-Policy "strict-origin-when-cross-origin" always' \
  'Permissions-Policy "geolocation=(), microphone=(), camera=(), payment=()" always'; do
  require_literal frontend/nginx.conf "$header"
done

if grep -Eq "script-src[^;]*(unsafe-inline|unsafe-eval)" \
  Caddyfile deploy/Caddyfile.blue-green.template frontend/nginx.conf; then
  fail "executable inline scripts or eval are allowed by an edge CSP"
fi

require_literal Caddyfile 'handle /health {'
require_literal Caddyfile 'handle /health/ready {'
require_literal Caddyfile 'handle /ready {'
require_literal Caddyfile 'redir * /health/ready 308'
require_literal Caddyfile 'handle /manifest.webmanifest {'
require_literal Caddyfile 'rewrite * /manifest.json'
require_literal Caddyfile 'header Content-Type "application/manifest+json"'
require_literal frontend/nginx.conf 'location = /manifest.webmanifest {'
require_literal frontend/nginx.conf 'include /etc/nginx/edge-routes.conf;'
require_literal frontend/nginx.conf 'location @spa_route {'
require_literal frontend/Dockerfile 'cp /usr/share/nginx/html/favicon_io/favicon.ico /usr/share/nginx/html/favicon.ico'
require_literal frontend/Dockerfile 'COPY --chmod=0444 frontend/nginx.conf /etc/nginx/nginx.conf'
require_literal frontend/Dockerfile 'COPY --from=builder --chmod=0444 /app/edge-routes.conf /etc/nginx/edge-routes.conf'

(cd frontend && node scripts/generate-edge-routes.mjs --check) || fail "React and edge route manifests differ"

for public_file in \
  frontend/public/robots.txt \
  frontend/public/sitemap.xml \
  frontend/public/.well-known/security.txt \
  frontend/public/favicon_io/favicon.ico; do
  [[ -s "$public_file" ]] || fail "$public_file is missing or empty"
done

require_literal frontend/public/robots.txt 'Sitemap: https://clpr.tv/sitemap.xml'
require_literal frontend/public/.well-known/security.txt 'Contact: mailto:security@clpr.tv'
require_literal frontend/public/.well-known/security.txt 'Canonical: https://clpr.tv/.well-known/security.txt'
require_literal frontend/public/.well-known/security.txt 'Preferred-Languages: en'

python3 - <<'PY'
import datetime as dt
import json
import pathlib
import xml.etree.ElementTree as ET

root = pathlib.Path("frontend/public")
json.loads((root / "manifest.json").read_text())
ET.parse(root / "sitemap.xml")

fields = {}
for line in (root / ".well-known/security.txt").read_text().splitlines():
    if ":" in line:
        key, value = line.split(":", 1)
        fields[key] = value.strip()
expires = dt.datetime.fromisoformat(fields["Expires"].replace("Z", "+00:00"))
minimum = dt.datetime.now(dt.timezone.utc) + dt.timedelta(days=180)
if expires <= minimum:
    raise SystemExit("security.txt Expires must remain at least 180 days in the future")
PY

if [[ "$runtime" != true ]]; then
  echo "edge contract: static checks passed"
  exit 0
fi

command -v docker >/dev/null || fail "docker is required for --runtime"
command -v curl >/dev/null || fail "curl is required for --runtime"
[[ -s frontend/dist/index.html ]] || fail "frontend/dist is required; build the frontend first"

tmp_dir="$(mktemp -d)"
caddy_container="clpr-edge-contract-${RANDOM}-$$"
frontend_container="clpr-frontend-contract-${RANDOM}-$$"
validation_container=""
edge_port=$((20000 + RANDOM % 3000))
backend_port=$((23000 + RANDOM % 3000))
frontend_port=$((26000 + RANDOM % 3000))
backend_pid=""
runtime_network=host
if docker inspect "$(hostname)" >/dev/null 2>&1; then
  runtime_network="container:$(hostname)"
fi

cleanup() {
  [[ -z "$validation_container" ]] || docker rm -f "$validation_container" >/dev/null 2>&1 || true
  docker rm -f "$caddy_container" >/dev/null 2>&1 || true
  docker rm -f "$frontend_container" >/dev/null 2>&1 || true
  [[ -z "$backend_pid" ]] || kill "$backend_pid" >/dev/null 2>&1 || true
  rm -rf "$tmp_dir"
}
trap cleanup EXIT

mkdir -p "$tmp_dir/frontend"
cp -R frontend/dist/. "$tmp_dir/frontend/"
cp frontend/public/favicon_io/favicon.ico "$tmp_dir/frontend/favicon.ico"
find "$tmp_dir/frontend" -type d -exec chmod 755 {} +
find "$tmp_dir/frontend" -type f -exec chmod 644 {} +
node frontend/scripts/generate-edge-routes.mjs \
  --nginx-output "$tmp_dir/edge-routes.conf"
sed "s/listen 8080;/listen $frontend_port;/" frontend/nginx.conf \
  > "$tmp_dir/nginx.conf"

python3 - "$backend_port" <<'PY' &
import http.server
import json
import pathlib
import socketserver
import sys

port = int(sys.argv[1])

class Backend(http.server.BaseHTTPRequestHandler):
    def do_GET(self):
        statuses = {
            "/health": (200, {"status": "healthy"}),
            "/health/ready": (200, {"status": "ready"}),
            "/api/v1/config": (200, {"environment": "contract"}),
            "/api/v1/fail": (500, {"error": "expected contract failure"}),
        }
        status, payload = statuses.get(self.path, (404, {"error": "not found"}))
        body = json.dumps(payload).encode()
        self.send_response(status)
        self.send_header("Content-Type", "application/json")
        self.send_header("Content-Length", str(len(body)))
        self.end_headers()
        self.wfile.write(body)

    def log_message(self, *_):
        pass

with socketserver.TCPServer(("127.0.0.1", port), Backend) as server:
    server.serve_forever()
PY
backend_pid=$!

CANARY_TOKEN='edge-contract-token-0123456789-ABCDEFGH' \
TEMPLATE_FILE="$repo_root/deploy/Caddyfile.blue-green.template" \
  bash scripts/blue-green-deploy.sh render blue green "$tmp_dir/Caddyfile.rendered"

awk '/^http:\/\/clpr\.tv/{exit} {print}' "$tmp_dir/Caddyfile.rendered" \
  | sed \
      -e 's/^clpr\.tv {/:18080 {/' \
      -e "s/:18080 {/:$edge_port {/" \
      -E \
      -e "s/clpr-backend-(blue|green):8080/127.0.0.1:$backend_port/g" \
      -e "s/clpr-frontend-(blue|green):8080/127.0.0.1:$frontend_port/g" \
      -e 's|output file /var/log/caddy/access.log|output file /tmp/access.log|' \
  > "$tmp_dir/Caddyfile"

validation_container="$(docker create nginx:1.29-alpine nginx -t)"
docker cp "$tmp_dir/nginx.conf" "$validation_container:/etc/nginx/nginx.conf"
docker cp "$tmp_dir/edge-routes.conf" "$validation_container:/etc/nginx/edge-routes.conf"
docker start --attach "$validation_container" >/dev/null
docker rm "$validation_container" >/dev/null
validation_container=""

validation_container="$(docker create caddy:2 caddy validate --config /etc/caddy/Caddyfile)"
docker cp "$tmp_dir/Caddyfile" "$validation_container:/etc/caddy/Caddyfile"
docker start --attach "$validation_container" >/dev/null
docker rm "$validation_container" >/dev/null
validation_container=""

docker create --name "$frontend_container" --network "$runtime_network" \
  nginx:1.29-alpine nginx -g 'daemon off;' >/dev/null
docker cp "$tmp_dir/nginx.conf" "$frontend_container:/etc/nginx/nginx.conf"
docker cp "$tmp_dir/edge-routes.conf" "$frontend_container:/etc/nginx/edge-routes.conf"
docker cp "$tmp_dir/frontend/." "$frontend_container:/usr/share/nginx/html"
docker start "$frontend_container" >/dev/null

for _ in $(seq 1 30); do
  if curl --silent --fail "http://127.0.0.1:$frontend_port/health.html" >/dev/null; then
    break
  fi
  sleep 1
done
curl --silent --fail "http://127.0.0.1:$frontend_port/health.html" >/dev/null || fail "nginx did not become ready"

docker create --name "$caddy_container" --network "$runtime_network" \
  caddy:2 caddy run --config /etc/caddy/Caddyfile >/dev/null
docker cp "$tmp_dir/Caddyfile" "$caddy_container:/etc/caddy/Caddyfile"
docker start "$caddy_container" >/dev/null

for _ in $(seq 1 30); do
  if curl --silent --fail "http://127.0.0.1:$edge_port/health" >/dev/null; then
    break
  fi
  sleep 1
done
curl --silent --fail "http://127.0.0.1:$edge_port/health" >/dev/null || fail "Caddy did not become ready"

assert_status() {
  local expected="$1"
  local url="$2"
  local actual
  actual="$(curl --silent --output "$tmp_dir/body" --dump-header "$tmp_dir/headers" --write-out '%{http_code}' "$url")"
  [[ "$actual" == "$expected" ]] || fail "$url returned $actual, expected $expected"
}

assert_header() {
  local name="$1"
  local value="$2"
  grep -Eiq "^${name}: ${value}?$" "$tmp_dir/headers" || fail "missing $name: $value for the last request"
}

for path in /health /health/ready /api/v1/config /robots.txt /sitemap.xml /.well-known/security.txt /favicon.ico; do
  assert_status 200 "http://127.0.0.1:$edge_port$path"
  assert_header X-Frame-Options DENY
  assert_header X-Content-Type-Options nosniff
  assert_header Referrer-Policy strict-origin-when-cross-origin
  assert_header Permissions-Policy 'geolocation=\(\), microphone=\(\), camera=\(\), payment=\(\)'
done

# The edge must serve the SPA shell only for declared static and dynamic routes.
for path in \
  / \
  /about \
  /clips/release-smoke-id \
  /playlists/release-smoke-id/theatre \
  /admin/discovery-lists/release-smoke-id/edit \
  /settings/cookies; do
  assert_status 200 "http://127.0.0.1:$edge_port$path"
  grep -Fq '<div id="root"></div>' "$tmp_dir/body" || fail "$path did not return the SPA shell"
done

for path in /definitely-not-a-clpr-route /unknown/nested/path; do
  assert_status 404 "http://127.0.0.1:$edge_port$path"
done

assert_status 200 "http://127.0.0.1:$edge_port/manifest.webmanifest"
assert_header Content-Type 'application/manifest\+json'
python3 -m json.tool "$tmp_dir/body" >/dev/null

assert_status 308 "http://127.0.0.1:$edge_port/ready"
assert_header Location /health/ready

assert_status 500 "http://127.0.0.1:$edge_port/api/v1/fail"
assert_header X-Frame-Options DENY
assert_header Content-Security-Policy '.*'

echo "edge contract: runtime HTTP and configuration checks passed"
