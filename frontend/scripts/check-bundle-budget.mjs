import { readdir, stat } from 'node:fs/promises';
import path from 'node:path';

const assetDirectory = path.resolve('dist/assets');
const files = await readdir(assetDirectory);
const sizes = await Promise.all(
  files.map(async (file) => ({ file, bytes: (await stat(path.join(assetDirectory, file))).size })),
);

const budgets = [
  { label: 'initial app JavaScript', match: /^app-.*\.js$/, max: 550 * 1024 },
  { label: 'lazy JavaScript chunk', match: /^chunk-.*\.js$/, max: 600 * 1024 },
  { label: 'application CSS', match: /^index-.*\.css$/, max: 120 * 1024 },
];

const failures = [];
for (const budget of budgets) {
  for (const asset of sizes.filter(({ file }) => budget.match.test(file))) {
    if (asset.bytes > budget.max) {
      failures.push(
        `${budget.label} ${asset.file} is ${(asset.bytes / 1024).toFixed(1)} KiB ` +
        `(budget ${(budget.max / 1024).toFixed(0)} KiB)`,
      );
    }
  }
}

if (failures.length > 0) {
  console.error(`Bundle budget exceeded:\n${failures.map((failure) => `- ${failure}`).join('\n')}`);
  process.exit(1);
}

const app = sizes.find(({ file }) => /^app-.*\.js$/.test(file));
console.log(`Bundle budgets pass (initial app ${(app?.bytes ?? 0) / 1024 | 0} KiB).`);
