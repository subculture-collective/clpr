const fs = require('node:fs');
const path = require('node:path');

const SHA_PATTERN = /^[0-9a-f]{40}$/;
const DIGEST_PATTERN = /^sha256:[0-9a-f]{64}$/;
const ISO_TIMESTAMP_PATTERN = /^(\d{4})-(\d{2})-(\d{2})T(\d{2}):(\d{2}):(\d{2})(?:\.(\d{1,3}))?(Z|[+-]\d{2}:\d{2})$/;
const COMMON_FIELDS = [
  'candidate_sha',
  'backend_digest',
  'frontend_digest',
  'crawler_digest',
  'executed_at',
  'evidence_url',
  'owner',
];
const IDENTITY_FIELDS = [
  'candidate_sha',
  'backend_digest',
  'frontend_digest',
  'crawler_digest',
];

const REQUIREMENTS = {
  'hosted-ci.json': {
    status: 'passed',
    required_workflow_passed: true, branch_protection_verified: true, webkit_passed: true,
  },
  'security-operations.json': {
    status: 'passed', jwt_exposure_reviewed: true, jwt_rotated_or_not_required: true,
    secret_scan_passed: true, candidate_dispositions_reviewed: true,
  },
  'stripe-test-mode.json': {
    status: 'passed', mode: 'test', checkout_passed: true, signed_webhooks_passed: true,
    duplicate_and_out_of_order_webhooks_passed: true, dunning_recovery_passed: true,
    cancellation_passed: true, reconciliation_passed: true,
    entitlement_activation_and_revocation_passed: true,
  },
  'load-and-soak.json': {
    status: 'passed', baseline_passed: true, stress_passed: true, soak_passed: true,
    thresholds_passed: true,
  },
  'restore.json': {
    status: 'passed', restore_passed: true, application_smoke_passed: true, rpo_met: true,
    rto_met: true,
  },
  'rollback.json': {
    status: 'passed', canary_passed: true, drain_passed: true, rollback_passed: true,
    post_rollback_smoke_passed: true,
  },
};

function isNonEmptyString(value) {
  return typeof value === 'string' && value.trim() !== '';
}

function parseIsoTimestamp(value) {
  if (!isNonEmptyString(value)) return null;
  const match = value.match(ISO_TIMESTAMP_PATTERN);
  if (!match) return null;

  const [, yearText, monthText, dayText, hourText, minuteText, secondText, , zone] = match;
  const year = Number(yearText);
  const month = Number(monthText);
  const day = Number(dayText);
  const hour = Number(hourText);
  const minute = Number(minuteText);
  const second = Number(secondText);
  if (month < 1 || month > 12 || day < 1 || day > new Date(Date.UTC(year, month, 0)).getUTCDate()) return null;
  if (hour > 23 || minute > 59 || second > 59) return null;
  if (zone !== 'Z') {
    const zoneHour = Number(zone.slice(1, 3));
    const zoneMinute = Number(zone.slice(4, 6));
    if (zoneHour > 23 || zoneMinute > 59) return null;
  }

  const timestamp = Date.parse(value);
  return Number.isFinite(timestamp) ? timestamp : null;
}

function isHttpsUrl(value) {
  if (!isNonEmptyString(value)) return false;
  try {
    return new URL(value).protocol === 'https:';
  } catch {
    return false;
  }
}

function readJson(filename, label, failures) {
  if (!fs.existsSync(filename)) {
    failures.push(`missing ${label}`);
    return null;
  }
  try {
    const record = JSON.parse(fs.readFileSync(filename, 'utf8'));
    if (!record || typeof record !== 'object' || Array.isArray(record)) {
      failures.push(`${label} must be a JSON object`);
      return null;
    }
    return record;
  } catch (error) {
    failures.push(`${label} is not valid JSON: ${error.message}`);
    return null;
  }
}

function validateIdentity(record, label, failures) {
  if (!SHA_PATTERN.test(record.candidate_sha || '')) {
    failures.push(`${label}: candidate_sha must be a lowercase 40-character Git SHA`);
  }
  for (const key of ['backend_digest', 'frontend_digest', 'crawler_digest']) {
    if (!DIGEST_PATTERN.test(record[key] || '')) {
      failures.push(`${label}: ${key} must be an immutable sha256 digest; mutable tags are forbidden`);
    }
  }
}

function verifyReleaseEvidence({ evidenceDir, manifestPath, now = Date.now(), expectedCandidateSha } = {}) {
  const failures = [];
  const manifest = readJson(manifestPath, 'candidate manifest', failures);
  if (!manifest) return failures;

  if (manifest.schema_version !== 1) failures.push('candidate manifest: schema_version must equal 1');
  validateIdentity(manifest, 'candidate manifest', failures);
  const builtAt = parseIsoTimestamp(manifest.built_at);
  if (builtAt === null) failures.push('candidate manifest: built_at must be a valid ISO-8601 timestamp with timezone');
  else if (builtAt > now) failures.push('candidate manifest: built_at must not be in the future');
  if (expectedCandidateSha && manifest.candidate_sha !== expectedCandidateSha) {
    failures.push('candidate manifest: candidate_sha must match the checked-out workflow commit');
  }

  for (const [filename, expected] of Object.entries(REQUIREMENTS)) {
    const evidence = readJson(path.join(evidenceDir, filename), `operator evidence: ${filename}`, failures);
    if (!evidence) continue;

    for (const key of COMMON_FIELDS) {
      if (!isNonEmptyString(evidence[key])) failures.push(`${filename}: ${key} is required`);
    }
    validateIdentity(evidence, filename, failures);
    if (!isHttpsUrl(evidence.evidence_url)) failures.push(`${filename}: evidence_url must be an absolute HTTPS URL`);

    const executedAt = parseIsoTimestamp(evidence.executed_at);
    if (executedAt === null) failures.push(`${filename}: executed_at must be a valid ISO-8601 timestamp with timezone`);
    else {
      if (executedAt > now) failures.push(`${filename}: executed_at must not be in the future`);
      if (builtAt !== null && executedAt <= builtAt) failures.push(`${filename}: executed_at must postdate candidate manifest built_at`);
    }

    for (const key of IDENTITY_FIELDS) {
      if (evidence[key] !== manifest[key]) failures.push(`${filename}: ${key} must match the candidate manifest`);
    }
    for (const [key, value] of Object.entries(expected)) {
      if (evidence[key] !== value) failures.push(`${filename}: ${key} must equal ${JSON.stringify(value)}`);
    }
  }

  return failures;
}

function main() {
  const root = process.cwd();
  const failures = [];
  const spec = fs.readFileSync(path.join(root, 'docs/openapi/openapi.yaml'), 'utf8');
  const transitionalOperations = (spec.match(/x-clpr-router-derived:\s*true/g) || []).length;
  if (transitionalOperations > 0) {
    failures.push(`${transitionalOperations} OpenAPI operations still use transitional router-derived contracts`);
  }

  const nowOverride = process.env.RELEASE_EVIDENCE_NOW;
  const now = nowOverride ? parseIsoTimestamp(nowOverride) : Date.now();
  if (now === null) failures.push('RELEASE_EVIDENCE_NOW must be a valid ISO-8601 timestamp with timezone');
  else {
    failures.push(...verifyReleaseEvidence({
      evidenceDir: process.env.RELEASE_EVIDENCE_DIR || path.join(root, 'release-evidence'),
      manifestPath: process.env.RELEASE_CANDIDATE_MANIFEST || path.join(root, 'release-evidence', 'candidate-manifest.json'),
      now,
      expectedCandidateSha: process.env.RELEASE_EXPECTED_CANDIDATE_SHA,
    }));
  }

  if (failures.length) {
    console.error(`Release evidence gate failed:\n- ${failures.join('\n- ')}`);
    process.exit(1);
  }
  console.log('Release evidence gate passed');
}

if (require.main === module) main();

module.exports = {
  COMMON_FIELDS,
  IDENTITY_FIELDS,
  REQUIREMENTS,
  parseIsoTimestamp,
  verifyReleaseEvidence,
};
