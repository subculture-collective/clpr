#!/usr/bin/env bash

set -euo pipefail

repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$repo_root"

fail() {
    echo "source convergence contract: $*" >&2
    exit 1
}

bash -n scripts/source-convergence.sh
bash -n scripts/install-lychee.sh
bash -n scripts/verify-gitea-actions-runtime.sh
node --check scripts/compare-vitest-runs.mjs
node --test scripts/tests/compare-vitest-runs.test.mjs
bash scripts/tests/configure-test-service-host.test.sh
bash scripts/tests/configure-hosted-test-ports.test.sh
bash scripts/tests/playwright-production-server.test.sh
bash scripts/tests/test-backup-restore-formats-container.test.sh
bash scripts/source-convergence.sh --verify

if bash scripts/source-convergence.sh >/dev/null 2>&1; then
    fail "convergence must require an explicit complete or partial mode"
fi
if bash scripts/source-convergence.sh --phase unknown >/dev/null 2>&1; then
    fail "convergence accepted an unknown phase"
fi

phase_list="$(bash scripts/source-convergence.sh --list)"
for phase in contract backend frontend api-docs docs secrets browser contracts images; do
    grep -Eq "^${phase}[[:space:]]" <<<"$phase_list" || fail "missing phase: $phase"
done

required_commands=(
    'task contract'
    'go test -count=1 ./...'
    'go test -race -count=1 ./...'
    'go vet ./...'
    'govulncheck@v1.6.0'
    'gosec@v2.27.1'
    'npm audit --omit=dev --audit-level=high'
    'vitest-run-1.json'
    'vitest-run-2.json'
    'npm run build'
    'npm run bundle:check'
    'npm run routes:check'
    'npm run openapi:lint'
    'npm run docs:check'
    'verify-secret-history.sh'
    'test:e2e:mocked'
    'run-real-e2e.sh'
    'test-real-e2e-runner-contract.sh'
    'test-edge-contracts.sh --runtime'
    'test-blue-green-contracts.sh'
    'test-operator-preflight.sh'
    'Dockerfile.crawler'
    'HIGH,CRITICAL'
    '-o spdx-json'
)
for command in "${required_commands[@]}"; do
    grep -Fq -- "$command" scripts/source-convergence.sh || fail "missing gate command: $command"
done

python3 - <<'PY'
import pathlib
import yaml

workflow_path = pathlib.Path('.gitea/workflows/source-convergence.yml')
workflow = yaml.safe_load(workflow_path.read_text())
jobs = workflow.get('jobs', {})
if list(jobs) != ['converge']:
    raise SystemExit('Gitea convergence must remain one workspace-preserving job')
steps = jobs['converge'].get('steps', [])
uses = [step.get('uses', '') for step in steps]
if any('download-artifact' in action for action in uses):
    raise SystemExit('cross-run artifact downloads are forbidden in Gitea convergence')
uploads = [step for step in steps if 'upload-artifact@v3' in step.get('uses', '')]
if len(uploads) != 1:
    raise SystemExit('exactly one same-run upload-artifact@v3 step is required')
if not uploads[0]['uses'].startswith('https://github.com/'):
    raise SystemExit('Gitea action source URL must be explicit')
if uploads[0].get('with', {}).get('if-no-files-found') != 'error':
    raise SystemExit('artifact upload must fail when diagnostics are missing')
if not uploads[0].get('with', {}).get('include-hidden-files'):
    raise SystemExit('hidden .tmp diagnostics must be explicitly included')
for step in steps:
    action = step.get('uses')
    if action and not action.startswith('https://github.com/'):
        raise SystemExit(f'action source is not explicit: {action}')

for path in pathlib.Path('.gitea/workflows').glob('*.yml'):
    candidate = yaml.safe_load(path.read_text())
    for job in candidate.get('jobs', {}).values():
        for step in job.get('steps', []):
            action = step.get('uses', '')
            if action and not action.startswith('https://github.com/'):
                raise SystemExit(f'{path}: action source is not explicit: {action}')

release_gates = yaml.safe_load(pathlib.Path('.gitea/workflows/release-gates.yml').read_text())
required_names = {
    'Secret history gate',
    'Documentation and OpenAPI gate',
    'Browser gate (Chromium, Firefox, WebKit)',
    'Image scan and SBOM gate',
}
actual_names = {job.get('name') for job in release_gates['jobs'].values()}
if not required_names <= actual_names:
    raise SystemExit(f'Gitea release gates are missing: {sorted(required_names - actual_names)}')

candidate_text = pathlib.Path('.gitea/workflows/immutable-candidate.yml').read_text()
for required in [
    'candidate-manifest.json', 'docker push', 'cosign sign --yes',
    'cosign verify', '-o spdx-json', 'HIGH,CRITICAL', 'OCI_REVISION',
    'backend_digest', 'frontend_digest', 'crawler_digest',
]:
    if required not in candidate_text:
        raise SystemExit(f'immutable candidate workflow is missing: {required}')
if ':latest' in candidate_text:
    raise SystemExit('immutable candidate workflow must not use latest tags')
if '-v "$evidence_dir:/evidence"' in candidate_text:
    raise SystemExit('immutable candidate evidence must not use runner-private bind mounts')

convergence_text = pathlib.Path('scripts/source-convergence.sh').read_text()
if '-v "$artifact_dir/images:/evidence"' in convergence_text:
    raise SystemExit('image evidence must not use runner-private bind mounts')

for path in [
    '.gitea/workflows/source-convergence.yml',
    '.gitea/workflows/immutable-candidate.yml',
    '.gitea/workflows/release-gates.yml',
]:
    if 'scripts/install-lychee.sh' not in pathlib.Path(path).read_text():
        raise SystemExit(f'{path}: pinned lychee installer is required')
PY

if grep -Eq '(continue-on-error|--exit-code[ =]0|--severity LOW)' \
    scripts/source-convergence.sh .gitea/workflows/source-convergence.yml; then
    fail "a convergence gate is configured to continue after failure"
fi

echo "source convergence contract passed"
