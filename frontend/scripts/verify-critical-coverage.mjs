import { readFile } from 'node:fs/promises';
import path from 'node:path';

// Threshold matching alone can silently ignore an excluded or renamed file.
// Require measured statements for every protected source in the actual report.
const contracts = JSON.parse(await readFile('config/critical-coverage.json', 'utf8'));
const report = JSON.parse(await readFile(process.argv[2] || 'coverage/coverage-final.json', 'utf8'));
const failures = [];
for (const source of Object.keys(contracts)) {
    const measured = report[path.resolve(source)];
    if (!measured || Object.keys(measured.statementMap || {}).length === 0 ||
        !Object.values(measured.s || {}).some(count => count > 0)) {
        failures.push(`Critical source was not exercised: ${source}`);
    }
}
if (failures.length) {
    console.error(failures.join('\n'));
    process.exit(1);
}
console.log(`Verified measured coverage for ${Object.keys(contracts).length} protected modules`);
