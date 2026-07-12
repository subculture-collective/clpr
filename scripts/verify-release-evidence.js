const fs = require('node:fs');
const path = require('node:path');

const root = process.cwd();
const failures = [];
const spec = fs.readFileSync(path.join(root, 'docs/openapi/openapi.yaml'), 'utf8');
const transitionalOperations = (spec.match(/x-clpr-router-derived:\s*true/g) || []).length;
if (transitionalOperations > 0) {
  failures.push(`${transitionalOperations} OpenAPI operations still use transitional router-derived contracts`);
}

const requirements = {
  'hosted-ci.json': {
    required_workflow_passed: true, branch_protection_verified: true, webkit_passed: true,
  },
  'security-operations.json': {
    jwt_exposure_reviewed: true, jwt_rotated_or_not_required: true, secret_scan_passed: true,
  },
  'stripe-test-mode.json': {
    mode: 'test', checkout_passed: true, signed_webhooks_passed: true, reconciliation_passed: true,
  },
  'load-and-soak.json': {
    baseline_passed: true, stress_passed: true, soak_passed: true, thresholds_passed: true,
  },
  'restore.json': {
    restore_passed: true, application_smoke_passed: true, rpo_met: true, rto_met: true,
  },
  'rollback.json': {
    canary_passed: true, drain_passed: true, rollback_passed: true, post_rollback_smoke_passed: true,
  },
};

for (const [filename, expected] of Object.entries(requirements)) {
  const evidencePath = path.join(root, 'release-evidence', filename);
  if (!fs.existsSync(evidencePath)) {
    failures.push(`missing operator evidence: release-evidence/${filename}`);
    continue;
  }
  let evidence;
  try {
    evidence = JSON.parse(fs.readFileSync(evidencePath, 'utf8'));
  } catch (error) {
    failures.push(`${filename} is not valid JSON: ${error.message}`);
    continue;
  }
  for (const [key, value] of Object.entries(expected)) {
    if (evidence[key] !== value) failures.push(`${filename}: ${key} must equal ${JSON.stringify(value)}`);
  }
  for (const key of ['candidate_sha', 'executed_at', 'evidence_url', 'owner']) {
    if (typeof evidence[key] !== 'string' || evidence[key].trim() === '') failures.push(`${filename}: ${key} is required`);
  }
}

if (failures.length) {
  console.error(`Release evidence gate failed:\n- ${failures.join('\n- ')}`);
  process.exit(1);
}
console.log('Release evidence gate passed');
