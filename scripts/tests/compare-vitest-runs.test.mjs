import assert from 'node:assert/strict';
import { execFileSync } from 'node:child_process';
import { mkdtempSync, writeFileSync } from 'node:fs';
import { tmpdir } from 'node:os';
import { join } from 'node:path';
import test from 'node:test';

const comparator = new URL('../compare-vitest-runs.mjs', import.meta.url);
const passing = {
  numTotalTestSuites: 4,
  numPassedTestSuites: 4,
  numFailedTestSuites: 0,
  numTotalTests: 12,
  numPassedTests: 12,
  numFailedTests: 0,
  numPendingTests: 0,
  success: true,
};

function compare(first, second) {
  const directory = mkdtempSync(join(tmpdir(), 'clpr-vitest-runs-'));
  const firstPath = join(directory, 'first.json');
  const secondPath = join(directory, 'second.json');
  writeFileSync(firstPath, JSON.stringify(first));
  writeFileSync(secondPath, JSON.stringify(second));
  return execFileSync(process.execPath, [comparator.pathname, firstPath, secondPath], {
    encoding: 'utf8',
    stdio: ['ignore', 'pipe', 'pipe'],
  });
}

test('accepts two successful runs with identical totals', () => {
  assert.match(compare(passing, { ...passing }), /Stable frontend tests: 12 tests in 4 suites/);
});

test('rejects different totals', () => {
  assert.throws(() => compare(passing, { ...passing, numPassedTests: 11, numTotalTests: 11 }));
});

test('rejects failures even when both summaries match', () => {
  const failed = {
    ...passing,
    numPassedTestSuites: 3,
    numFailedTestSuites: 1,
    numPassedTests: 11,
    numFailedTests: 1,
    success: false,
  };
  assert.throws(() => compare(failed, { ...failed }));
});

test('rejects a malformed summary', () => {
  const malformed = { ...passing };
  delete malformed.numPendingTests;
  assert.throws(() => compare(malformed, passing));
});
