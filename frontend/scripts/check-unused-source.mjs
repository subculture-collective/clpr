import ts from 'typescript';
import { readFile } from 'node:fs/promises';
import path from 'node:path';

const root = process.cwd();
const config = ts.readConfigFile('tsconfig.app.json', ts.sys.readFile);
if (config.error) throw new Error(ts.flattenDiagnosticMessageText(config.error.messageText, '\n'));
const parsed = ts.parseJsonConfigFileContent(config.config, ts.sys, root);
if (parsed.errors.length) throw new Error(ts.formatDiagnosticsWithColorAndContext(parsed.errors, {
    getCanonicalFileName: file => file,
    getCurrentDirectory: () => root,
    getNewLine: () => '\n',
}));
const edges = new Map();
for (const file of parsed.fileNames) {
    const ast = ts.createSourceFile(file, await readFile(file, 'utf8'), ts.ScriptTarget.Latest, true);
    const dependencies = [];
    function visit(node) {
        let specifier;
        if (ts.isImportDeclaration(node) || ts.isExportDeclaration(node)) specifier = node.moduleSpecifier;
        if (ts.isCallExpression(node) && node.expression.kind === ts.SyntaxKind.ImportKeyword) {
            specifier = node.arguments[0];
            if (!specifier || !ts.isStringLiteral(specifier)) {
                throw new Error(`Nonliteral dynamic import in ${file}: declare its entry points before changing this audit`);
            }
        }
        if (specifier && ts.isStringLiteral(specifier)) {
            const resolved = ts.resolveModuleName(specifier.text, file, parsed.options, ts.sys).resolvedModule;
            if (resolved) dependencies.push(resolved.resolvedFileName);
        }
        ts.forEachChild(node, visit);
    }
    visit(ast);
    edges.set(file, dependencies);
}
const reachable = new Set();
function walk(file) {
    if (reachable.has(file)) return;
    reachable.add(file);
    for (const dependency of edges.get(file) ?? []) walk(dependency);
}
walk(path.join(root, 'src/main.tsx'));
const unused = [...edges.keys()].filter(file => !reachable.has(file) && !file.endsWith('.d.ts'));
if (unused.length) {
    console.error('Source files unreachable from src/main.tsx:\n' + unused.map(file => path.relative(root, file)).join('\n'));
    process.exit(1);
}
console.log('All application source files are reachable from src/main.tsx.');
