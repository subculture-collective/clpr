const fs = require('node:fs');
const path = require('node:path');
const { globSync } = require('glob');

const root = process.cwd();
const docs = globSync(['README.md', 'docs/**/*.md'], { cwd: root, nodir: true });
const repositoryFiles = globSync('**/*', {
  cwd: root,
  nodir: true,
  dot: true,
  ignore: ['.git/**', 'node_modules/**', 'frontend/node_modules/**'],
});
const byBasename = new Map();
for (const file of repositoryFiles) {
  const basename = path.basename(file);
  if (!byBasename.has(basename)) byBasename.set(basename, []);
  byBasename.get(basename).push(file);
}

let repaired = 0;
let removed = 0;
for (const file of docs) {
  const absolute = path.join(root, file);
  const original = fs.readFileSync(absolute, 'utf8');
  const updated = original.replace(/(!?)\[([^\]]*)\]\(([^)]+)\)/g, (whole, image, label, rawTarget) => {
    const target = rawTarget.trim().replace(/^<|>$/g, '');
    if (/^(?:https?:|mailto:|tel:|data:|#)/i.test(target)) return whole;
    const [targetWithoutAnchor, anchor = ''] = target.split('#', 2);
    const cleanTarget = decodeURI(targetWithoutAnchor.split('?', 1)[0]);
    if (!cleanTarget) return whole;
    const resolved = cleanTarget.startsWith('/')
      ? path.join(root, cleanTarget.slice(1))
      : path.resolve(root, path.dirname(file), cleanTarget);
    if (fs.existsSync(resolved)) return whole;

    const candidates = byBasename.get(path.basename(cleanTarget)) || [];
    if (candidates.length === 1) {
      let relative = path.relative(path.dirname(file), candidates[0]).split(path.sep).join('/');
      if (!relative.startsWith('.')) relative = `./${relative}`;
      repaired += 1;
      return `${image}[${label}](${relative}${anchor ? `#${anchor}` : ''})`;
    }

    removed += 1;
    return image ? label : label;
  });
  if (updated !== original) fs.writeFileSync(absolute, updated);
}

console.log(`Corrected ${repaired} uniquely relocatable links and removed ${removed} broken link targets`);
