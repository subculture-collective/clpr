import { execFileSync } from 'node:child_process';
import { mkdtempSync, readFileSync, rmSync } from 'node:fs';
import { tmpdir } from 'node:os';
import { join } from 'node:path';

const temporaryDirectory = mkdtempSync(join(tmpdir(), 'clpr-openapi-'));
const generated = join(temporaryDirectory, 'openapi.json');
const committed = 'backend/internal/openapi/generated/openapi.json';

try {
    execFileSync('./node_modules/.bin/redocly', [
        'bundle',
        'docs/openapi/openapi.yaml',
        '-o',
        generated,
        '--ext',
        'json',
    ], { stdio: 'ignore' });
    const specification = JSON.parse(readFileSync(generated, 'utf8'));
    for (const [route, operations] of Object.entries(specification.paths)) {
        for (const [method, operation] of Object.entries(operations)) {
            if (operation?.['x-clpr-router-derived'] === true) {
                throw new Error(`transitional OpenAPI contract: ${method.toUpperCase()} ${route}; document the actual handler contract before qualification`);
            }
        }
    }
    if (readFileSync(generated, 'utf8') !== readFileSync(committed, 'utf8')) {
        throw new Error('embedded OpenAPI artifact is stale; run npm run openapi:embed');
    }
    console.log('Embedded OpenAPI artifact is current.');
} finally {
    rmSync(temporaryDirectory, { recursive: true, force: true });
}
