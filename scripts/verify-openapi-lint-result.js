const fs = require('node:fs');

const EXPECTED_AMBIGUOUS_ROUTE_EXCEPTIONS = 25;

function validateLintResult(result) {
  const failures = [];
  if (!result || typeof result !== 'object' || !result.totals) return ['Redocly output is missing totals'];
  if (result.totals.errors !== 0) failures.push(`Redocly reported ${result.totals.errors} errors`);
  if (result.totals.warnings !== 0) failures.push(`Redocly reported ${result.totals.warnings} unsuppressed warnings`);
  if (result.totals.ignored !== EXPECTED_AMBIGUOUS_ROUTE_EXCEPTIONS) {
    failures.push(`Redocly ignored ${result.totals.ignored} problems; expected exactly ${EXPECTED_AMBIGUOUS_ROUTE_EXCEPTIONS} route-precedence exceptions`);
  }
  return failures;
}

function main() {
  const filename = process.argv[2];
  if (!filename) throw new Error('usage: node scripts/verify-openapi-lint-result.js <redocly-json>');
  const result = JSON.parse(fs.readFileSync(filename, 'utf8'));
  const failures = validateLintResult(result);
  if (failures.length > 0) {
    console.error(`OpenAPI lint gate failed:\n- ${failures.join('\n- ')}`);
    process.exit(1);
  }
  console.log(`OpenAPI lint gate passed: 0 errors, 0 warnings, ${EXPECTED_AMBIGUOUS_ROUTE_EXCEPTIONS} owned route-precedence exceptions`);
}

if (require.main === module) main();

module.exports = { EXPECTED_AMBIGUOUS_ROUTE_EXCEPTIONS, validateLintResult };
