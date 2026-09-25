// Mirrors the public Markdown documentation into the backend module so the API
// binary can embed it. The backend image is built from backend/, which cannot
// see ../docs, so GET /api/v1/docs served nothing in production.
//
// Only the documents listed in docs/public-docs.json are published. Operations,
// security reviews, internal design notes, and test plans stay in the
// repository but are neither embedded nor served. The manifest is copied next to
// the embedded tree, and the API refuses any path it does not list.
//
//   node scripts/embed-docs.mjs          write backend/internal/docscontent/
//   node scripts/embed-docs.mjs --check  fail when the embedded copy differs
//                                        from the allowlisted set
import { cpSync, existsSync, lstatSync, mkdirSync, readdirSync, readFileSync, rmSync } from 'node:fs';
import { dirname, join, relative, sep } from 'node:path';

const source = 'docs';
const manifestPath = join(source, 'public-docs.json');
const packageDirectory = 'backend/internal/docscontent';
const target = join(packageDirectory, 'content');
const embeddedManifestPath = join(packageDirectory, 'public-docs.json');
const documentPattern = /^[A-Za-z0-9][A-Za-z0-9_.-]*(?:\/[A-Za-z0-9][A-Za-z0-9_.-]*)*\.md$/;

function fail(message) {
    console.error(message);
    process.exit(1);
}

function readManifest() {
    let manifest;
    try {
        manifest = JSON.parse(readFileSync(manifestPath, 'utf8'));
    } catch (error) {
        fail(`${manifestPath}: ${error.message}`);
    }
    if (manifest.schema_version !== 1 || !Array.isArray(manifest.documents)) {
        fail(`${manifestPath} must use schema_version 1 and a documents array`);
    }
    const documents = manifest.documents;
    const problems = [];
    for (const document of documents) {
        if (typeof document !== 'string' || !documentPattern.test(document) || document.split('/').includes('..')) {
            problems.push(`invalid document path ${JSON.stringify(document)}`);
            continue;
        }
        // Every path component must be a real directory or file; a symlink
        // could publish content from outside the allowlist.
        const parts = document.split('/');
        for (let index = 1; index <= parts.length; index++) {
            const component = join(source, ...parts.slice(0, index));
            if (!existsSync(component)) {
                problems.push(`missing ${join(source, document)}`);
                break;
            }
            const info = lstatSync(component);
            if (info.isSymbolicLink()) {
                problems.push(`symlink ${component}`);
                break;
            }
            if (index === parts.length && !info.isFile()) {
                problems.push(`not a regular file ${component}`);
            }
        }
    }
    if (new Set(documents).size !== documents.length) {
        problems.push('duplicate documents');
    }
    if (JSON.stringify(documents) !== JSON.stringify([...documents].sort())) {
        problems.push('documents must remain sorted');
    }
    if (problems.length > 0) {
        fail(`${manifestPath}:\n${problems.join('\n')}`);
    }
    return new Map(documents.map(document => [document, join(source, document)]));
}

// Every file under root, keyed by its slash-separated relative path.
function collect(root) {
    const files = new Map();
    if (!existsSync(root)) {
        return files;
    }
    const walk = directory => {
        for (const entry of readdirSync(directory, { withFileTypes: true })) {
            const path = join(directory, entry.name);
            if (entry.isDirectory()) {
                walk(path);
            } else {
                files.set(relative(root, path).split(sep).join('/'), path);
            }
        }
    };
    walk(root);
    return files;
}

const wanted = readManifest();

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
            problems.push(`unexpected ${name} (not listed in ${manifestPath})`);
        }
    }
    if (!existsSync(embeddedManifestPath) || !readFileSync(manifestPath).equals(readFileSync(embeddedManifestPath))) {
        problems.push(`stale ${embeddedManifestPath}`);
    }
    if (problems.length > 0) {
        console.error(problems.slice(0, 20).join('\n'));
        console.error(`embedded documentation differs from ${manifestPath} (${problems.length} differences); run npm run docs:embed`);
        process.exit(1);
    }
    console.log(`Embedded documentation matches ${manifestPath} (${wanted.size} files).`);
} else {
    rmSync(target, { recursive: true, force: true });
    for (const [name, path] of wanted) {
        const destination = join(target, name);
        mkdirSync(dirname(destination), { recursive: true });
        cpSync(path, destination);
    }
    cpSync(manifestPath, embeddedManifestPath);
    console.log(`Embedded ${wanted.size} public documentation files into ${target}.`);
}
