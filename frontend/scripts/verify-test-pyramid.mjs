import { readdir, readFile } from 'node:fs/promises';
import path from 'node:path';

async function filesUnder(directory, predicate) {
    const entries = await readdir(directory, { withFileTypes: true });
    const nested = await Promise.all(entries.map(async entry => {
        const target = path.join(directory, entry.name);
        if (entry.isDirectory()) return filesUnder(target, predicate);
        return predicate(target) ? [target] : [];
    }));
    return nested.flat();
}

const contracts = JSON.parse(await readFile('../config/test-contracts.json', 'utf8'));
const failures = [];
const coverage = JSON.parse(await readFile('config/critical-coverage.json', 'utf8'));
for (const [file, floors] of Object.entries(coverage)) {
    try {
        await readFile(file, 'utf8');
    } catch {
        failures.push('Missing critical coverage source: ' + file);
    }
    for (const metric of ['statements', 'lines', 'branches', 'functions']) {
        if (!Number.isFinite(floors[metric]) || floors[metric] <= 0 || floors[metric] > 100) {
            failures.push('Invalid critical coverage floor: ' + file + ' ' + metric);
        }
    }
}
if (Object.keys(coverage).length === 0) failures.push('No critical coverage contracts');

for (const [behavior, suites] of Object.entries(contracts)) {
    if (!Array.isArray(suites) || suites.length === 0) {
        failures.push(`${behavior}: no owning suite`);
        continue;
    }
    for (const suite of suites) {
        try {
            const source = await readFile(path.join('..', suite), 'utf8');
            if (!/\b(?:it|test)\s*(?:\(|\.each\s*\()|\bfunc Test\w+\(/.test(source)) {
                failures.push(`${behavior}: ${suite} contains no tests`);
            }
        } catch {
            failures.push(`${behavior}: missing suite ${suite}`);
        }
    }
}

const frontendUnit = await filesUnder('src', file => /\.(?:test|spec)\.tsx?$/.test(file));
const backendTests = await filesUnder('../backend', file => /_test\.go$/.test(file));
const browser = await filesUnder('e2e', file => /\.spec\.ts$/.test(file));
for (const file of browser) {
    // Every browser spec must belong to a runnable project. Candidate checks
    // use their own HTTPS-only config; the other tiers use playwright.config.ts.
    if (!/^e2e\/(mocked|real-backend|candidate)\//.test(file)) {
        failures.push(`Unassigned browser spec: ${file}`);
    }
}
for (const file of [...frontendUnit, ...browser]) {
    const source = await readFile(file, 'utf8');
    if (/\b(?:it|test|describe)\.(?:skip|only|todo)\b|\b(?:xit|xtest|xdescribe)\s*\(/.test(source)) {
        failures.push(`Skipped, focused, or placeholder test: ${file}`);
    }
}

// Counts describe the suite; they are not quotas. Behavior ownership, executed
// assertions, coverage thresholds, and the real-backend tier are the gates.
console.log(JSON.stringify({
    criticalBehaviorGroups: Object.keys(contracts).length,
    frontendUnitComponentFiles: frontendUnit.length,
    backendTestFiles: backendTests.length,
    browserFiles: browser.length,
}, null, 2));
if (failures.length) {
    console.error(failures.join('\n'));
    process.exit(1);
}
