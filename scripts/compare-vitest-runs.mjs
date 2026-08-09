#!/usr/bin/env node

import fs from 'node:fs';

if (process.argv.length !== 4) {
  console.error('usage: compare-vitest-runs.mjs <run-one.json> <run-two.json>');
  process.exit(2);
}

function summary(filename) {
  const result = JSON.parse(fs.readFileSync(filename, 'utf8'));
  const fields = ['numTotalTestSuites', 'numPassedTestSuites', 'numFailedTestSuites', 'numTotalTests', 'numPassedTests', 'numFailedTests', 'numPendingTests'];
  const values = Object.fromEntries(fields.map(field => [field, result[field]]));
  for (const [field, value] of Object.entries(values)) {
    if (!Number.isInteger(value) || value < 0) throw new Error(`${filename}: invalid ${field}`);
  }
  if (values.numFailedTestSuites !== 0 || values.numFailedTests !== 0) {
    throw new Error(`${filename}: test failures are not allowed`);
  }
  if (result.success !== true) throw new Error(`${filename}: success must equal true`);
  return values;
}

const first = summary(process.argv[2]);
const second = summary(process.argv[3]);
if (JSON.stringify(first) !== JSON.stringify(second)) {
  console.error('Frontend test runs are not stable');
  console.error(JSON.stringify({ first, second }, null, 2));
  process.exit(1);
}
console.log(`Stable frontend tests: ${first.numPassedTests} tests in ${first.numPassedTestSuites} suites`);
