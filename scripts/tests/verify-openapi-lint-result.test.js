const test = require('node:test');
const assert = require('node:assert/strict');

const { validateLintResult } = require('../verify-openapi-lint-result');

test('accepts only the exact owned OpenAPI warning budget', () => {
  assert.deepEqual(validateLintResult({ totals: { errors: 0, warnings: 0, ignored: 25 } }), []);
});

test('rejects errors and unsuppressed warnings', () => {
  assert.deepEqual(validateLintResult({ totals: { errors: 1, warnings: 2, ignored: 25 } }), [
    'Redocly reported 1 errors',
    'Redocly reported 2 unsuppressed warnings',
  ]);
});

test('rejects any change to the exception count', () => {
  assert.match(
    validateLintResult({ totals: { errors: 0, warnings: 0, ignored: 26 } })[0],
    /expected exactly 25/,
  );
});

test('rejects malformed Redocly output', () => {
  assert.deepEqual(validateLintResult({}), ['Redocly output is missing totals']);
});
