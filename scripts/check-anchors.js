const fs = require('node:fs');
const path = require('node:path');
const { globSync } = require('glob');

const root = process.cwd();
const files = globSync(['README.md', 'docs/**/*.md'], { cwd: root, nodir: true });
const anchorsByFile = new Map();

function stripFencedCode(contents) {
  return contents.replace(/^(```|~~~)[\s\S]*?^\1.*$/gm, '');
}

function slugHeadings(contents) {
	contents = stripFencedCode(contents);
  const anchors = new Set();
  const occurrences = new Map();
  for (const line of contents.split(/\r?\n/)) {
    const heading = line.match(/^#{1,6}\s+(.+?)\s*#*$/);
    if (!heading) continue;
    const base = heading[1]
      .toLowerCase()
      .replace(/<[^>]+>/g, '')
      .replace(/[^\p{L}\p{N}\s_-]/gu, '')
      .trim()
      .replace(/\s/g, '-');
    const count = occurrences.get(base) || 0;
    anchors.add(count === 0 ? base : `${base}-${count}`);
    occurrences.set(base, count + 1);
  }
  for (const match of contents.matchAll(/<a\s+(?:name|id)=["']([^"']+)["']/gi)) anchors.add(match[1]);
  return anchors;
}

for (const file of files) anchorsByFile.set(file, slugHeadings(fs.readFileSync(path.join(root, file), 'utf8')));

const failures = [];
for (const file of files) {
  const contents = stripFencedCode(fs.readFileSync(path.join(root, file), 'utf8'));
  for (const match of contents.matchAll(/\[[^\]]*\]\(([^)]+)\)/g)) {
    const raw = match[1].trim().replace(/^<|>$/g, '');
    if (!raw.includes('#') || /^(?:https?:|mailto:|tel:)/i.test(raw)) continue;
    const [targetPart, encodedAnchor] = raw.split('#', 2);
    const target = targetPart ? path.normalize(path.join(path.dirname(file), decodeURI(targetPart))) : file;
    const anchor = decodeURIComponent(encodedAnchor || '').toLowerCase();
    if (!anchorsByFile.has(target)) failures.push(`${file}: missing anchor target file ${target}`);
    else if (anchor && !anchorsByFile.get(target).has(anchor)) failures.push(`${file}: missing #${anchor} in ${target}`);
  }
}

if (failures.length) {
  console.error(failures.join('\n'));
  process.exit(1);
}
console.log(`Validated anchors in ${files.length} Markdown files`);
