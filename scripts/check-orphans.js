const fs = require('node:fs');
const path = require('node:path');
const { globSync } = require('glob');

const root = process.cwd();
const files = globSync('docs/**/*.md', { cwd: root, nodir: true });
const known = new Set(files);
const reachable = new Set();
const queue = ['docs/index.md'];
const intentionallyStandalone = [
  /^docs\/superpowers\/plans\//,
  /^docs\/archive\//,
  /^docs\/examples\//,
];

while (queue.length) {
  const file = queue.shift();
  if (!known.has(file) || reachable.has(file)) continue;
  reachable.add(file);
  const contents = fs.readFileSync(path.join(root, file), 'utf8');
  for (const match of contents.matchAll(/\[[^\]]*\]\(([^)#]+)(?:#[^)]+)?\)/g)) {
    const raw = match[1].trim().replace(/^<|>$/g, '');
    if (/^(?:https?:|mailto:|tel:)/i.test(raw)) continue;
    const target = path.normalize(path.join(path.dirname(file), decodeURI(raw)));
    if (known.has(target) && !reachable.has(target)) queue.push(target);
  }
}

const orphans = files.filter(file => !reachable.has(file) && !intentionallyStandalone.some(pattern => pattern.test(file)));
if (orphans.length) {
  console.error(`Unreachable documentation (${orphans.length}):\n${orphans.sort().join('\n')}`);
  process.exit(1);
}
console.log(`Documentation graph reaches ${reachable.size}/${files.length} files (${files.length - reachable.size} intentional standalone files)`);
