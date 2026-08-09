const fs = require('node:fs');
const os = require('node:os');
const path = require('node:path');
const test = require('node:test');
const assert = require('node:assert/strict');

const {
  COMMON_FIELDS,
  REQUIREMENTS,
  parseIsoTimestamp,
  verifyReleaseEvidence,
} = require('../verify-release-evidence');
const {
  candidateSha,
  validEvidence,
  validManifest,
  writeFixture,
} = require('./fixtures/release-evidence-fixtures');

const now = Date.parse('2026-08-09T11:00:00Z');

function withFixture(callback, options) {
  const root = fs.mkdtempSync(path.join(os.tmpdir(), 'clpr-release-evidence-'));
  try {
    const fixture = writeFixture(root, options);
    callback({ root, ...fixture });
  } finally {
    fs.rmSync(root, { recursive: true, force: true });
  }
}

function verify(fixture, overrides = {}) {
  return verifyReleaseEvidence({ ...fixture, now, ...overrides });
}

test('accepts six consistent evidence files produced after an immutable candidate', () => {
  withFixture((fixture) => assert.deepEqual(verify(fixture), []));
});

test('requires all six operator evidence files', () => {
  withFixture((fixture) => {
    fs.rmSync(path.join(fixture.evidenceDir, 'restore.json'));
    assert.ok(verify(fixture).includes('missing operator evidence: restore.json'));
  });
});

test('reports malformed manifest and evidence JSON', () => {
  withFixture((fixture) => {
    fs.writeFileSync(fixture.manifestPath, '{');
    assert.match(verify(fixture).join('\n'), /candidate manifest is not valid JSON/);
  });
  withFixture((fixture) => {
    fs.writeFileSync(path.join(fixture.evidenceDir, 'rollback.json'), '{');
    assert.match(verify(fixture).join('\n'), /operator evidence: rollback\.json is not valid JSON/);
  });
});

test('rejects null, scalar, and array manifests and evidence records', () => {
  for (const invalid of [null, false, 0, 'passed', []]) {
    withFixture((fixture) => {
      fs.writeFileSync(fixture.manifestPath, JSON.stringify(invalid));
      assert.match(verify(fixture).join('\n'), /candidate manifest must be a JSON object/);
    });
    withFixture((fixture) => {
      const filename = 'restore.json';
      fs.writeFileSync(path.join(fixture.evidenceDir, filename), JSON.stringify(invalid));
      assert.match(verify(fixture).join('\n'), /operator evidence: restore\.json must be a JSON object/);
    });
  }
});

test('requires every common field in every evidence file', () => {
  for (const field of COMMON_FIELDS) {
    withFixture((fixture) => {
      const filename = 'hosted-ci.json';
      const evidence = validEvidence(filename);
      delete evidence[field];
      fs.writeFileSync(path.join(fixture.evidenceDir, filename), JSON.stringify(evidence));
      assert.ok(verify(fixture).some((failure) => failure === `${filename}: ${field} is required`));
    });
  }
});

test('rejects invalid SHAs, mutable image tags, and non-canonical digests', () => {
  for (const [field, value, message] of [
    ['candidate_sha', 'main', /candidate_sha must be a lowercase 40-character Git SHA/],
    ['backend_digest', 'clpr-backend:latest', /backend_digest must be an immutable sha256 digest/],
    ['frontend_digest', `sha256:${'A'.repeat(64)}`, /frontend_digest must be an immutable sha256 digest/],
  ]) {
    withFixture((fixture) => {
      const filename = 'security-operations.json';
      fs.writeFileSync(path.join(fixture.evidenceDir, filename), JSON.stringify(validEvidence(filename, { [field]: value })));
      assert.match(verify(fixture).join('\n'), message);
    });
  }
});

test('rejects identities that differ from the candidate manifest', () => {
  for (const field of ['candidate_sha', 'backend_digest', 'frontend_digest', 'crawler_digest']) {
    withFixture((fixture) => {
      const filename = 'load-and-soak.json';
      const replacement = field === 'candidate_sha'
        ? 'f'.repeat(40)
        : `sha256:${'d'.repeat(64)}`;
      fs.writeFileSync(path.join(fixture.evidenceDir, filename), JSON.stringify(validEvidence(filename, { [field]: replacement })));
      assert.ok(verify(fixture).includes(`${filename}: ${field} must match the candidate manifest`));
    });
  }
});

test('enforces status passed and every evidence-specific success field', () => {
  for (const [filename, expected] of Object.entries(REQUIREMENTS)) {
    for (const [field, requiredValue] of Object.entries(expected)) {
      const flippedValue = typeof requiredValue === 'boolean' ? !requiredValue : `not-${requiredValue}`;
      withFixture((fixture) => {
        fs.writeFileSync(path.join(fixture.evidenceDir, filename), JSON.stringify(validEvidence(filename, { [field]: flippedValue })));
        assert.ok(verify(fixture).some((failure) => failure.startsWith(`${filename}: ${field} must equal`)));
      });
    }
  }
});

test('the verifier owns every boolean declared by an evidence template', () => {
  for (const [filename, expected] of Object.entries(REQUIREMENTS)) {
    const template = JSON.parse(fs.readFileSync(path.join(__dirname, '..', '..', 'release-evidence', 'templates', filename), 'utf8'));
    for (const [field, value] of Object.entries(template)) {
      if (typeof value === 'boolean') assert.equal(expected[field], true, `${filename}.${field} must be required`);
    }
  }
});

test('rejects invalid, future, and pre-build evidence timestamps', () => {
  for (const [executedAt, message] of [
    ['2026-02-30T10:05:00Z', /valid ISO-8601/],
    ['2026-08-09', /valid ISO-8601/],
    ['2026-08-09T12:00:00Z', /must not be in the future/],
    ['2026-08-09T10:00:00Z', /must postdate candidate manifest built_at/],
    ['2026-08-09T09:59:59Z', /must postdate candidate manifest built_at/],
  ]) {
    withFixture((fixture) => {
      const filename = 'stripe-test-mode.json';
      fs.writeFileSync(path.join(fixture.evidenceDir, filename), JSON.stringify(validEvidence(filename, { executed_at: executedAt })));
      assert.match(verify(fixture).join('\n'), message);
    });
  }
});

test('rejects invalid or future candidate build timestamps', () => {
  for (const [builtAt, message] of [
    ['not-a-date', /built_at must be a valid ISO-8601/],
    ['2026-08-09T12:00:00Z', /built_at must not be in the future/],
  ]) {
    withFixture((fixture) => {
      fs.writeFileSync(fixture.manifestPath, JSON.stringify(validManifest({ built_at: builtAt })));
      assert.match(verify(fixture).join('\n'), message);
    });
  }
});

test('requires the versioned immutable candidate manifest identity', () => {
  for (const [overrides, message] of [
    [{ schema_version: 2 }, /schema_version must equal 1/],
    [{ candidate_sha: 'main' }, /candidate_sha must be a lowercase 40-character Git SHA/],
    [{ crawler_digest: 'clpr-crawler:latest' }, /crawler_digest must be an immutable sha256 digest/],
  ]) {
    withFixture((fixture) => {
      fs.writeFileSync(fixture.manifestPath, JSON.stringify(validManifest(overrides)));
      assert.match(verify(fixture).join('\n'), message);
    });
  }
});

test('requires HTTPS evidence URLs and a non-empty owner', () => {
  withFixture((fixture) => {
    const filename = 'rollback.json';
    fs.writeFileSync(path.join(fixture.evidenceDir, filename), JSON.stringify(validEvidence(filename, {
      evidence_url: 'http://evidence.example.test/run',
      owner: ' ',
    })));
    const failures = verify(fixture).join('\n');
    assert.match(failures, /evidence_url must be an absolute HTTPS URL/);
    assert.match(failures, /owner is required/);
  });
});

test('binds the manifest to the checked-out workflow commit', () => {
  withFixture((fixture) => {
    assert.deepEqual(verify(fixture, { expectedCandidateSha: candidateSha }), []);
    assert.match(verify(fixture, { expectedCandidateSha: 'f'.repeat(40) }).join('\n'), /checked-out workflow commit/);
  });
});

test('strict ISO parser accepts timezone timestamps and rejects normalized invalid dates', () => {
  assert.equal(parseIsoTimestamp('2026-08-09T10:05:00.123-05:00'), Date.parse('2026-08-09T10:05:00.123-05:00'));
  assert.equal(parseIsoTimestamp('2026-02-30T10:05:00Z'), null);
  assert.equal(parseIsoTimestamp('2026-08-09T25:05:00Z'), null);
});
