/**
 * Path helpers for the /docs page. Document paths are relative to the docs
 * root, use "/" between directories, and omit the ".md" extension, e.g.
 * "compliance/twitch-embeds".
 */

/** API URL for a document; each segment is encoded, slashes are kept. */
export function docContentUrl(docPath: string): string {
    const segments = docPath
        .split('/')
        .filter(segment => segment !== '')
        .map(segment => encodeURIComponent(segment));
    return `/api/v1/docs/content/${segments.join('/')}`;
}

export type DocLinkTarget =
    | { kind: 'document'; path: string; hash: string }
    | { kind: 'anchor'; hash: string }
    | { kind: 'external'; href: string }
    | { kind: 'site'; href: string }
    | { kind: 'unpublished' };

function normalize(parts: string[]): string[] | null {
    const resolved: string[] = [];
    for (const part of parts) {
        if (part === '' || part === '.') continue;
        if (part === '..') {
            if (resolved.length === 0) return null;
            resolved.pop();
            continue;
        }
        resolved.push(part);
    }
    return resolved;
}

function safeDecode(value: string): string {
    try {
        return decodeURIComponent(value);
    } catch {
        return value;
    }
}

/**
 * Resolves a link inside a rendered document. Relative Markdown links resolve
 * against the current document's directory; converted wikilinks, which are
 * relative to the docs root, are tried second. Links to documents that are not
 * published resolve to "unpublished" so they render as plain text instead of
 * a link that fails to load.
 */
export function resolveDocLink(
    currentDocPath: string,
    href: string,
    publishedPaths: ReadonlySet<string>,
): DocLinkTarget {
    const trimmed = href.trim();
    if (trimmed.startsWith('#')) {
        return { kind: 'anchor', hash: trimmed.slice(1) };
    }
    if (/^[a-z][a-z0-9+.-]*:/i.test(trimmed) || trimmed.startsWith('//')) {
        return { kind: 'external', href: trimmed };
    }
    if (trimmed.startsWith('/')) {
        return { kind: 'site', href: trimmed };
    }

    const hashIndex = trimmed.indexOf('#');
    const hash = hashIndex === -1 ? '' : trimmed.slice(hashIndex + 1);
    let target = hashIndex === -1 ? trimmed : trimmed.slice(0, hashIndex);
    const queryIndex = target.indexOf('?');
    if (queryIndex !== -1) target = target.slice(0, queryIndex);
    target = target.split('/').map(safeDecode).join('/');
    if (target === '') {
        return { kind: 'anchor', hash };
    }

    const lastSegment = target.split('/').pop() ?? '';
    if (lastSegment.includes('.') && !lastSegment.endsWith('.md')) {
        // Source files, specs, and images are not published documents.
        return { kind: 'unpublished' };
    }
    target = target.replace(/\.md$/, '');

    const currentDirectory = currentDocPath.split('/').slice(0, -1);
    const candidates = [
        normalize([...currentDirectory, ...target.split('/')]),
        normalize(target.split('/')),
    ];
    for (const candidate of candidates) {
        if (candidate && candidate.length > 0) {
            const path = candidate.join('/');
            if (publishedPaths.has(path)) {
                return { kind: 'document', path, hash };
            }
        }
    }
    return { kind: 'unpublished' };
}

interface DocTreeNode {
    path: string;
    type: 'file' | 'directory';
    children?: DocTreeNode[];
}

/** Paths of every document in a /api/v1/docs tree. */
export function publishedDocPaths(nodes: readonly DocTreeNode[]): Set<string> {
    const paths = new Set<string>();
    const visit = (list: readonly DocTreeNode[]) => {
        for (const node of list) {
            if (node.type === 'file') {
                paths.add(node.path);
            } else if (node.children) {
                visit(node.children);
            }
        }
    };
    visit(nodes);
    return paths;
}
