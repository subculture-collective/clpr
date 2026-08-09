import assert from 'node:assert/strict';
import test from 'node:test';

import { auditJwtPatch } from '../audit-jwt-secret-history.mjs';

test('rejects literal base64 JWT private signing material', () => {
  const encodedFixture = `${'Ab9/'.repeat(16)}==`;
  assert.throws(
    () => auditJwtPatch(`+JWT_PRIVATE_KEY_B64=${encodedFixture}`),
    /possible literal JWT signing material/,
  );
});

test('accepts externalized base64 JWT signing configuration', () => {
  assert.equal(auditJwtPatch('+JWT_PRIVATE_KEY_B64=${JWT_PRIVATE_KEY_B64}'), 1);
  assert.equal(auditJwtPatch('+JWT_SIGNING_KEY_B64: {{ .JWTSigningKeyB64 }}'), 1);
});
