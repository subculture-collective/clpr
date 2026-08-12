#!/usr/bin/env bash

set -euo pipefail

repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$repo_root"

readonly TRIVY_IMAGE='aquasec/trivy:0.69.2@sha256:3d1f862cb6c4fe13c1506f96f816096030d8d5ccdb2380a3069f7bf07daa86aa'
readonly SYFT_IMAGE='anchore/syft:v1.30.0@sha256:bd5357d2cd087f03af748dac24df48bfbc1723080d78f75f69aca1f2d429060e'
readonly OCI_SOURCE='https://git.subcult.tv/subculture-collective/clpr'
readonly all_phases=(contract backend frontend api-docs docs secrets browser contracts images)

declare -a selected_phases=()
mode=""
current_phase="initialization"
started_at="$(date -u +%Y-%m-%dT%H:%M:%SZ)"
candidate_sha=""
artifact_dir=""
result_file=""

usage() {
    cat <<'EOF'
Usage:
  source-convergence.sh --all
  source-convergence.sh --phase <name> [--phase <name> ...]
  source-convergence.sh --list
  source-convergence.sh --verify

--all is the only mode that constitutes a complete local source gate. A phase
selection is intentionally partial and is recorded as such. Every executing
mode requires a clean tracked/untracked checkout; unavailable tools, networks,
containers, vulnerability databases, or scanners fail the selected phase.
EOF
}

list_phases() {
    cat <<'EOF'
contract   task contract and clean-checkout invariants
backend    test, race, vet, build, govulncheck, and high-severity gosec
frontend   audit, lint, inventory, two stable test runs, build, budget, routes
api-docs   OpenAPI verifier tests, zero-unsuppressed lint, bundle, route contract
docs       markdown/local-link/anchor/orphan/asset checks and inventory parity
secrets    full-history/current-tree Gitleaks and JWT exposure audit
browser    mocked accessibility and real-backend Chromium/Firefox/WebKit gates
contracts  edge runtime, blue-green, and operator preflight contracts
images     three OCI-labelled builds, HIGH/CRITICAL scans, and image SBOMs
EOF
}

fail() {
    printf 'source convergence failed in %s: %s\n' "$current_phase" "$*" >&2
    exit 1
}

record_result() {
    local outcome="$1"
    local finished_at
    finished_at="$(date -u +%Y-%m-%dT%H:%M:%SZ)"
    [[ -n "$result_file" ]] || return 0
    python3 - "$result_file" "$candidate_sha" "$mode" "$outcome" "$current_phase" "$started_at" "$finished_at" "${selected_phases[*]}" <<'PY'
import json
import pathlib
import sys

path, sha, mode, outcome, phase, started, finished, phases = sys.argv[1:]
pathlib.Path(path).write_text(json.dumps({
    "schema_version": 1,
    "candidate_sha": sha,
    "mode": mode,
    "outcome": outcome,
    "last_phase": phase,
    "selected_phases": phases.split(),
    "started_at": started,
    "finished_at": finished,
}, indent=2) + "\n")
PY
}

on_exit() {
    local code=$?
    if (( code == 0 )); then
        record_result passed
    else
        record_result failed || true
    fi
}
trap on_exit EXIT

require_command() {
    command -v "$1" >/dev/null || fail "required command is unavailable: $1"
}

require_selected_tools() {
    local phase command
    for command in git bash python3; do require_command "$command"; done
    for phase in "${selected_phases[@]}"; do
        case "$phase" in
            contract)
                require_command task
                ;;
            backend)
                require_command go
                ;;
            frontend)
                require_command node
                require_command npm
                ;;
            api-docs)
                require_command go
                require_command node
                require_command npm
                ;;
            docs)
                require_command lychee
                require_command node
                require_command npm
                ;;
            secrets)
                require_command gitleaks
                require_command node
                ;;
            browser)
                require_command curl
                require_command docker
                require_command go
                require_command npm
                require_command task
                ;;
            contracts)
                require_command curl
                require_command docker
                require_command node
                ;;
            images)
                require_command docker
                ;;
        esac
    done
}

require_clean_checkout() {
    local changes
    changes="$(git status --porcelain=v1 --untracked-files=all)"
    [[ -z "$changes" ]] || fail "checkout is not clean; convergence evidence requires one immutable source tree"
    if [[ -n "$candidate_sha" ]]; then
        [[ "$(git rev-parse HEAD)" == "$candidate_sha" ]] \
            || fail "HEAD changed after source convergence began"
    fi
}

verify_orchestration() {
    bash -n scripts/source-convergence.sh
    node --check scripts/compare-vitest-runs.mjs
    python3 - <<'PY'
import pathlib
import yaml
yaml.safe_load(pathlib.Path('.gitea/workflows/source-convergence.yml').read_text())
PY
    echo "source convergence orchestration verified"
}

phase_contract() {
    task contract
}

phase_backend() {
    mkdir -p "$artifact_dir/backend"
    (
        cd backend
        go test -count=1 ./...
        go test -race -count=1 ./...
        go vet ./...
        go build -o "$artifact_dir/backend/api" ./cmd/api
        go run golang.org/x/vuln/cmd/govulncheck@v1.6.0 ./...
        go run github.com/securego/gosec/v2/cmd/gosec@v2.27.1 \
            -quiet -severity high ./...
    )
}

phase_frontend() {
    mkdir -p "$artifact_dir/frontend"
    (
        cd frontend
        npm audit --omit=dev --audit-level=high
        npm run lint
        npm run test:inventory
        run_vitest_with_diagnostics "$artifact_dir/frontend/vitest-run-1.json"
        run_vitest_with_diagnostics "$artifact_dir/frontend/vitest-run-2.json"
        node ../scripts/compare-vitest-runs.mjs \
            "$artifact_dir/frontend/vitest-run-1.json" \
            "$artifact_dir/frontend/vitest-run-2.json"
        npm run build
        npm run bundle:check
        npm run routes:check
    )
}

run_vitest_with_diagnostics() {
    local report="$1"
    local status
    set +e
    npm run test -- run --reporter=json --outputFile="$report"
    status=$?
    set -e
    if (( status == 0 )); then
        return 0
    fi

    node - "$report" <<'NODE'
const fs = require('node:fs');

const reportPath = process.argv[2];
if (!fs.existsSync(reportPath)) {
  console.error(`Vitest failed before writing its JSON report: ${reportPath}`);
  process.exit(0);
}

const report = JSON.parse(fs.readFileSync(reportPath, 'utf8'));
console.error(JSON.stringify({
  success: report.success,
  numTotalTestSuites: report.numTotalTestSuites,
  numFailedTestSuites: report.numFailedTestSuites,
  numTotalTests: report.numTotalTests,
  numFailedTests: report.numFailedTests,
  failedResults: (report.testResults || [])
    .filter((suite) => suite.status === 'failed')
    .map((suite) => ({
      name: suite.name,
      message: suite.message,
      failedAssertions: (suite.assertionResults || [])
        .filter((assertion) => assertion.status === 'failed')
        .map((assertion) => ({
          fullName: assertion.fullName,
          failureMessages: assertion.failureMessages,
        })),
    })),
}, null, 2));
NODE
    return "$status"
}

phase_api_docs() {
    mkdir -p "$artifact_dir/openapi"
    node --test scripts/tests/verify-openapi-lint-result.test.js
    npm run openapi:lint
    npm run openapi:embed:check
    ./node_modules/.bin/redocly bundle docs/openapi/openapi.yaml \
        -o "$artifact_dir/openapi/openapi-bundled.yaml"
    ./node_modules/.bin/redocly build-docs docs/openapi/openapi.yaml \
        -o "$artifact_dir/openapi/api-docs.html"
    (cd backend && go test -count=1 ./cmd/api -run 'Test(AmbiguousOpenAPIRoutes|OpenAPIAmbiguousRoute)')
}

phase_docs() {
    npm run docs:check
    npm run docs:inventory
    git diff --exit-code -- docs/reference-inventory.md
}

phase_secrets() {
    GITLEAKS_BIN="$(command -v gitleaks)" bash scripts/verify-secret-history.sh
    GITLEAKS_BIN="$(command -v gitleaks)" bash scripts/test-gitleaks-policy.sh
}

phase_browser() {
    (
        # shellcheck disable=SC2329 # Invoked through the EXIT trap.
        cleanup_browser_services() {
            task test:teardown >/dev/null 2>&1 || true
        }
        trap cleanup_browser_services EXIT
        bash scripts/verify-playwright-manifest.sh
        task test:setup
        bash scripts/run-release-critical-backend-tests.sh
        bash scripts/run-dependency-chaos.sh
        (cd frontend && npm run test:e2e:mocked)
        bash scripts/run-real-e2e.sh
    )
}

phase_contracts() {
    bash scripts/test-edge-contracts.sh --runtime
    bash scripts/test-blue-green-contracts.sh
    bash scripts/test-operator-preflight.sh
}

assert_image_contract() {
    local image="$1"
    local expected_user="$2"
    local revision source user
    revision="$(docker image inspect --format '{{index .Config.Labels "org.opencontainers.image.revision"}}' "$image")"
    source="$(docker image inspect --format '{{index .Config.Labels "org.opencontainers.image.source"}}' "$image")"
    user="$(docker image inspect --format '{{.Config.User}}' "$image")"
    [[ "$revision" == "$candidate_sha" ]] || fail "$image has incorrect OCI revision label"
    [[ "$source" == "$OCI_SOURCE" ]] || fail "$image has incorrect OCI source label"
    [[ "$user" == "$expected_user" ]] || fail "$image runs as $user, expected $expected_user"
}

scan_image() {
    local image="$1"
    local report="$2"
    docker run --rm \
        -v /var/run/docker.sock:/var/run/docker.sock \
        -v "$artifact_dir/trivy-cache:/root/.cache/trivy" \
        "$TRIVY_IMAGE" image --scanners vuln --ignore-unfixed \
        --severity HIGH,CRITICAL --exit-code 1 "$image" >/dev/null
    # Produce the retained report only after the blocking scan passes.
    docker run --rm \
        -v /var/run/docker.sock:/var/run/docker.sock \
        -v "$artifact_dir/trivy-cache:/root/.cache/trivy" \
        -v "$artifact_dir/images:/evidence" \
        "$TRIVY_IMAGE" image --scanners vuln --ignore-unfixed \
        --severity HIGH,CRITICAL --format json \
        --output "/evidence/$report" "$image"
}

generate_image_sbom() {
    local image="$1"
    local report="$2"
    docker run --rm \
        -v /var/run/docker.sock:/var/run/docker.sock \
        -v "$artifact_dir/images:/evidence" \
        "$SYFT_IMAGE" "docker:$image" -o "spdx-json=/evidence/$report"
    [[ -s "$artifact_dir/images/$report" ]] || fail "$image SBOM was not produced"
}

phase_images() {
    require_command docker
    docker info >/dev/null || fail "Docker daemon is unavailable"
    mkdir -p "$artifact_dir/images" "$artifact_dir/trivy-cache"
    local backend_image="clpr-convergence-backend:$candidate_sha"
    local crawler_image="clpr-convergence-crawler:$candidate_sha"
    local frontend_image="clpr-convergence-frontend:$candidate_sha"
    local build_args=(--build-arg "OCI_REVISION=$candidate_sha" --build-arg "OCI_SOURCE=$OCI_SOURCE")

    docker build "${build_args[@]}" -f backend/Dockerfile -t "$backend_image" backend
    docker build "${build_args[@]}" -f backend/Dockerfile.crawler -t "$crawler_image" backend
    docker build "${build_args[@]}" -f frontend/Dockerfile -t "$frontend_image" .
    assert_image_contract "$backend_image" '10001:10001'
    assert_image_contract "$crawler_image" '10001:10001'
    assert_image_contract "$frontend_image" '101:101'
    scan_image "$backend_image" backend-trivy.json
    scan_image "$crawler_image" crawler-trivy.json
    scan_image "$frontend_image" frontend-trivy.json
    generate_image_sbom "$backend_image" backend.spdx.json
    generate_image_sbom "$crawler_image" crawler.spdx.json
    generate_image_sbom "$frontend_image" frontend.spdx.json
}

run_phase() {
    current_phase="$1"
    printf '\n=== source convergence: %s ===\n' "$current_phase"
    "phase_${current_phase//-/_}"
    require_clean_checkout
}

while (( $# > 0 )); do
    case "$1" in
        --all)
            [[ -z "$mode" ]] || { usage >&2; exit 2; }
            mode=all
            selected_phases=("${all_phases[@]}")
            shift
            ;;
        --phase)
            [[ "$mode" != all ]] || { usage >&2; exit 2; }
            [[ $# -ge 2 ]] || { usage >&2; exit 2; }
            mode=partial
            selected_phases+=("$2")
            shift 2
            ;;
        --list)
            list_phases
            exit 0
            ;;
        --verify)
            verify_orchestration
            exit 0
            ;;
        --help|-h)
            usage
            exit 0
            ;;
        *)
            usage >&2
            exit 2
            ;;
    esac
done

[[ ${#selected_phases[@]} -gt 0 ]] || { usage >&2; exit 2; }
for phase in "${selected_phases[@]}"; do
    [[ " ${all_phases[*]} " == *" $phase "* ]] || fail "unknown phase: $phase"
done

current_phase=preflight
require_selected_tools
require_clean_checkout
candidate_sha="$(git rev-parse HEAD)"
[[ "$candidate_sha" =~ ^[0-9a-f]{40}$ ]] || fail "HEAD is not a full Git commit"
artifact_dir="$repo_root/.tmp/source-convergence/$candidate_sha"
result_file="$artifact_dir/result.json"
mkdir -p "$artifact_dir"

for phase in "${selected_phases[@]}"; do run_phase "$phase"; done
current_phase=complete
if [[ "$mode" == all ]]; then
    echo "Complete local source convergence passed for $candidate_sha"
else
    echo "Selected phases passed for $candidate_sha; this is not a complete source gate"
fi
