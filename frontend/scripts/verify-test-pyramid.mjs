import { readdir, readFile } from 'node:fs/promises';
import path from 'node:path';

async function filesUnder(directory, predicate) {
  const entries = await readdir(directory, { withFileTypes: true });
  const nested = await Promise.all(entries.map(async (entry) => {
    const target = path.join(directory, entry.name);
    if (entry.isDirectory()) return filesUnder(target, predicate);
    return predicate(target) ? [target] : [];
  }));
  return nested.flat();
}

const frontendUnit = await filesUnder('src', (file) => /\.(?:test|spec)\.tsx?$/.test(file));
const backendTests = await filesUnder('../backend', (file) => /_test\.go$/.test(file));
const mockedBrowser = await filesUnder('e2e/mocked', (file) => /\.spec\.ts$/.test(file));
const realBrowser = await filesUnder('e2e/real-backend', (file) => /\.spec\.ts$/.test(file));
const allBrowser = await filesUnder('e2e', (file) => /\.spec\.ts$/.test(file));
const browserContents = await Promise.all(allBrowser.map((file) => readFile(file, 'utf8')));
const skippedBrowserTests = browserContents.reduce(
  (total, source) => total + (source.match(/\b(?:test|describe)\.skip\s*\(/g)?.length ?? 0),
  0,
);

const inventory = {
  frontendUnitComponentFiles: frontendUnit.length,
  backendTestFiles: backendTests.length,
  mockedBrowserFiles: mockedBrowser.length,
  realBackendBrowserFiles: realBrowser.length,
  skippedBrowserTests,
};

const minimums = {
  frontendUnitComponentFiles: 125,
  backendTestFiles: 150,
  mockedBrowserFiles: 2,
  realBackendBrowserFiles: 1,
};
const failures = Object.entries(minimums)
  .filter(([tier, minimum]) => inventory[tier] < minimum)
  .map(([tier, minimum]) => `${tier}: ${inventory[tier]} (minimum ${minimum})`);
if (skippedBrowserTests > 1) {
  failures.push(`skippedBrowserTests: ${skippedBrowserTests} (maximum 1)`);
}

console.log(JSON.stringify(inventory, null, 2));
if (failures.length > 0) {
  console.error(`Test pyramid regression:\n${failures.map((failure) => `- ${failure}`).join('\n')}`);
  process.exit(1);
}
