const fs = require('node:fs');
const path = require('node:path');

const { REQUIREMENTS } = require('../../verify-release-evidence');

const candidateSha = '0123456789abcdef0123456789abcdef01234567';
const digests = {
  backend_digest: `sha256:${'a'.repeat(64)}`,
  frontend_digest: `sha256:${'b'.repeat(64)}`,
  crawler_digest: `sha256:${'c'.repeat(64)}`,
};

function validManifest(overrides = {}) {
  return {
    schema_version: 1,
    candidate_sha: candidateSha,
    ...digests,
    built_at: '2026-08-09T10:00:00Z',
    ...overrides,
  };
}

function validEvidence(filename, overrides = {}) {
  return {
    status: 'passed',
    candidate_sha: candidateSha,
    ...digests,
    executed_at: '2026-08-09T10:05:00Z',
    evidence_url: `https://evidence.example.test/runs/${filename}`,
    owner: 'release-test-fixture',
    ...REQUIREMENTS[filename],
    ...overrides,
  };
}

function writeFixture(root, { manifest = validManifest(), evidenceOverrides = {} } = {}) {
  const evidenceDir = path.join(root, 'release-evidence');
  fs.mkdirSync(evidenceDir, { recursive: true });
  const manifestPath = path.join(root, 'candidate-manifest.json');
  fs.writeFileSync(manifestPath, `${JSON.stringify(manifest, null, 2)}\n`);
  for (const filename of Object.keys(REQUIREMENTS)) {
    const evidence = validEvidence(filename, evidenceOverrides[filename]);
    fs.writeFileSync(path.join(evidenceDir, filename), `${JSON.stringify(evidence, null, 2)}\n`);
  }
  return { evidenceDir, manifestPath };
}

module.exports = { candidateSha, digests, validEvidence, validManifest, writeFixture };
