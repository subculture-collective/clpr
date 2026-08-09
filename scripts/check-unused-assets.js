const fs = require('node:fs');
const path = require('node:path');
const { globSync } = require('glob');

const root = process.cwd();
const assets = globSync('docs/**/*.{png,jpg,jpeg,gif,svg,webp,avif,pdf}', { cwd: root, nodir: true });
const markdown = globSync(['README.md', 'docs/**/*.md'], { cwd: root, nodir: true })
  .map(file => fs.readFileSync(path.join(root, file), 'utf8'))
  .join('\n');
const unused = assets.filter(asset => !markdown.includes(asset) && !markdown.includes(path.basename(asset)));

if (unused.length) {
  console.error(`Unused documentation assets (${unused.length}):\n${unused.sort().join('\n')}`);
  process.exit(1);
}
console.log(`Validated references for ${assets.length} documentation assets`);
