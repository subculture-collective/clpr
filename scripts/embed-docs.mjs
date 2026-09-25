// Mirrors the published Markdown documentation into the backend module so the
// API binary can embed it. The backend image is built from backend/, which
// cannot see ../docs, so GET /api/v1/docs served nothing in production.
//
//   node scripts/embed-docs.mjs          write backend/internal/docscontent/content
//   node scripts/embed-docs.mjs --check  fail when the embedded copy is stale
import { cpSync, existsSync, mkdirSync, readdirSync, readFileSync, rmSync } from 'node:fs';
import { dirname, join, relative } from 'node:path';

const source = 'docs';
const target = 'backend/internal/docscontent/content';
// Keep in sync with the skip rules in backend/internal/handlers/docs_handler.go.
const skippedDirectories = new Set(['archive', 'vault']);

function collect(root) {
    const files = new Map();
    if (!existsSync(root)) {
        return files;
    }
    const walk = (directory) => {
        for (const entry of readdirSync(directory, { withFileTypes: true })) {
            if (entry.name.startsWith('.') || entry.name.startsWith('_') || entry.isSymbolicLink()) {
                continue;
            }
            const path = join(directory, entry.name);
            if (entry.isDirectory()) {
                if (!skippedDirectories.has(entry.name)) {
                    walk(path);
                }
            } else if (entry.isFile() && entry.name.endsWith('.md')) {
                files.set(relative(root, path), path);
            }
        }
    };
    walk(root);
    return files;
}

const wanted = collect(source);

if (process.argv.includes('--check')) {
    const embedded = collect(target);
    const problems = [];
    for (const [name, path] of wanted) {
        const copy = embedded.get(name);
        if (!copy) {
            problems.push(`missing ${name}`);
        } else if (!readFileSync(path).equals(readFileSync(copy))) {
            problems.push(`stale ${name}`);
        }
    }
    for (const name of embedded.keys()) {
        if (!wanted.has(name)) {
            problems.push(`unexpected ${name}`);
        }
    }
    if (problems.length > 0) {
        console.error(problems.slice(0, 20).join('\n'));
        console.error(`embedded documentation is stale (${problems.length} differences); run npm run docs:embed`);
        process.exit(1);
    }
    console.log(`Embedded documentation is current (${wanted.size} files).`);
} else {
    rmSync(target, { recursive: true, force: true });
    for (const [name, path] of wanted) {
        const destination = join(target, name);
        mkdirSync(dirname(destination), { recursive: true });
        cpSync(path, destination);
    }
    console.log(`Embedded ${wanted.size} documentation files into ${target}.`);
}
