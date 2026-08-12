import { useState, useEffect, useMemo } from 'react';
import ReactMarkdown from 'react-markdown';
import remarkGfm from 'remark-gfm';
import { Container, Card, CardBody, CardHeader, SEO } from '../../components';
import {
    parseMarkdown,
    headingToId,
    extractTextFromChildren,
} from '../../lib/markdown-utils';
import type { ProcessedMarkdown } from '../../lib/markdown-utils';

interface APIVersion {
    version: string;
    label: string;
    file: string;
}

const API_VERSIONS: APIVersion[] = [
    { version: '1.0.0', label: 'v1.0.0 (Current)', file: 'api-reference.md' },
    { version: 'baseline', label: 'Baseline', file: 'api-baseline.md' },
    { version: 'changelog', label: 'Changelog', file: 'api-changelog.md' },
];

interface OpenAPIOperation {
    summary?: string;
    description?: string;
    tags?: string[];
}

interface OpenAPIDocument {
    info?: { title?: string; version?: string; description?: string };
    paths?: Record<string, Record<string, OpenAPIOperation>>;
}

function openAPIToMarkdown(document: OpenAPIDocument): string {
    const sections = new Map<string, string[]>();
    for (const [path, pathItem] of Object.entries(document.paths ?? {})) {
        for (const [method, operation] of Object.entries(pathItem)) {
            if (!['get', 'post', 'put', 'patch', 'delete'].includes(method))
                continue;
            const tag = operation.tags?.[0] ?? 'Other';
            const entries = sections.get(tag) ?? [];
            entries.push(
                `### ${method.toUpperCase()} ${path}\n\n${operation.summary ?? operation.description ?? 'No description available.'}`,
            );
            sections.set(tag, entries);
        }
    }
    const heading = `# ${document.info?.title ?? 'clpr API'}\n\nVersion ${document.info?.version ?? 'unknown'}\n\n${document.info?.description ?? ''}`;
    return [
        heading,
        ...[...sections.entries()]
            .sort(([a], [b]) => a.localeCompare(b))
            .map(([tag, entries]) => `## ${tag}\n\n${entries.join('\n\n')}`),
    ].join('\n\n');
}

/**
 * Admin API Documentation Page
 *
 * Displays generated API reference documentation from OpenAPI spec
 * with version selection and search capabilities.
 */
export function AdminAPIDocsPage() {
    const [selectedVersion, setSelectedVersion] = useState<string>('1.0.0');
    const [content, setContent] = useState<string>('');
    const [processedDoc, setProcessedDoc] = useState<ProcessedMarkdown | null>(
        null,
    );
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState<string | null>(null);
    const [searchQuery, setSearchQuery] = useState('');
    const [selectedTag, setSelectedTag] = useState<string>('all');

    // Extract tags from table of contents
    const tags = useMemo(() => {
        if (!processedDoc?.toc) return [];

        // Get level 2 headings as tags
        const tagHeadings = processedDoc.toc.filter((item) => item.level === 2);
        return ['all', ...tagHeadings.map((item) => item.text)];
    }, [processedDoc]);

    useEffect(() => {
        const loadAPIDoc = async () => {
            setLoading(true);
            setError(null);

            try {
                const versionInfo = API_VERSIONS.find(
                    (v) => v.version === selectedVersion,
                );
                if (!versionInfo) {
                    throw new Error('Version not found');
                }

                const response = await fetch(
                    selectedVersion === '1.0.0'
                        ? '/api/v1/admin/openapi.json'
                        : `/docs/openapi/generated/${versionInfo.file}`,
                );
                if (!response.ok)
                    throw new Error(
                        'This generated document is not available in the release artifact.',
                    );
                const text =
                    selectedVersion === '1.0.0'
                        ? openAPIToMarkdown(
                              (await response.json()) as OpenAPIDocument,
                          )
                        : await response.text();
                setContent(text);

                // Process markdown
                const processed = parseMarkdown(text);
                setProcessedDoc(processed);
            } catch (err) {
                console.error('Error loading API docs:', err);
                setError(
                    err instanceof Error
                        ? err.message
                        : 'Failed to load API documentation',
                );
            } finally {
                setLoading(false);
            }
        };

        loadAPIDoc();
    }, [selectedVersion]);

    // Filter content based on search and tag
    const filteredContent = useMemo(() => {
        if (!processedDoc) return '';

        let filtered = processedDoc.content;

        // Filter by tag (category)
        if (selectedTag !== 'all') {
            // Extract section for the selected tag
            const lines = filtered.split('\n');
            // Escape special regex characters in selectedTag
            const escapedTag = selectedTag.replace(
                /[.*+?^${}()|[\]\\]/g,
                '\\$&',
            );
            const startPattern = new RegExp(`^## ${escapedTag}$`, 'i');
            const endPattern = /^## /;

            let inSection = false;
            const sectionLines: string[] = [];

            for (let i = 0; i < lines.length; i++) {
                const line = lines[i];

                if (startPattern.test(line)) {
                    inSection = true;
                    sectionLines.push(line);
                } else if (inSection) {
                    if (endPattern.test(line)) {
                        break;
                    }
                    sectionLines.push(line);
                }
            }

            filtered = sectionLines.join('\n');
        }

        // Filter by search query
        if (searchQuery.trim()) {
            const query = searchQuery.toLowerCase();
            const lines = filtered.split('\n');
            const matchingLines: string[] = [];
            const addedLineIndices = new Set<number>();
            let inMatchingSection = false;
            let sectionStart = 0;

            for (let i = 0; i < lines.length; i++) {
                const line = lines[i];
                const lowerLine = line.toLowerCase();

                // Check if line matches search
                if (lowerLine.includes(query)) {
                    inMatchingSection = true;
                    sectionStart = i;
                }

                // Include heading context
                if (line.startsWith('###')) {
                    inMatchingSection = false;
                    sectionStart = i;
                }

                if (inMatchingSection || lowerLine.includes(query)) {
                    // Include some context around the match
                    const contextStart = Math.max(0, sectionStart);
                    for (let j = contextStart; j <= i; j++) {
                        if (!addedLineIndices.has(j)) {
                            matchingLines.push(lines[j]);
                            addedLineIndices.add(j);
                        }
                    }
                }
            }

            filtered = matchingLines.join('\n');
        }

        return filtered;
    }, [processedDoc, selectedTag, searchQuery]);

    // Custom components for markdown rendering
    const markdownComponents = useMemo(
        () => ({
            // Style links
            a: ({
                href,
                children,
            }: {
                href?: string;
                children?: React.ReactNode;
            }) => (
                <a
                    href={href}
                    target={href?.startsWith('http') ? '_blank' : undefined}
                    rel={
                        href?.startsWith('http')
                            ? 'noopener noreferrer'
                            : undefined
                    }
                    className="text-primary hover:underline"
                >
                    {children}
                </a>
            ),
            // Style code blocks
            code: ({
                inline,
                className,
                children,
            }: {
                inline?: boolean;
                className?: string;
                children?: React.ReactNode;
            }) => {
                if (inline) {
                    return (
                        <code className="px-1 py-0.5 bg-muted rounded text-sm">
                            {children}
                        </code>
                    );
                }
                return (
                    <code
                        className={`block p-4 bg-muted rounded-lg overflow-x-auto text-sm ${
                            className || ''
                        }`}
                    >
                        {children}
                    </code>
                );
            },
            // Style headings with IDs for navigation
            h1: ({ children }: { children?: React.ReactNode }) => {
                const text = extractTextFromChildren(children);
                const id = headingToId(text);
                return (
                    <h1 id={id} className="text-4xl font-bold mb-4 mt-6">
                        {children}
                    </h1>
                );
            },
            h2: ({ children }: { children?: React.ReactNode }) => {
                const text = extractTextFromChildren(children);
                const id = headingToId(text);
                return (
                    <h2 id={id} className="text-3xl font-semibold mb-3 mt-5">
                        {children}
                    </h2>
                );
            },
            h3: ({ children }: { children?: React.ReactNode }) => {
                const text = extractTextFromChildren(children);
                const id = headingToId(text);
                return (
                    <h3 id={id} className="text-2xl font-semibold mb-2 mt-4">
                        {children}
                    </h3>
                );
            },
            h4: ({ children }: { children?: React.ReactNode }) => {
                const text = extractTextFromChildren(children);
                const id = headingToId(text);
                return (
                    <h4 id={id} className="text-xl font-semibold mb-2 mt-3">
                        {children}
                    </h4>
                );
            },
            h5: ({ children }: { children?: React.ReactNode }) => {
                const text = extractTextFromChildren(children);
                const id = headingToId(text);
                return (
                    <h5 id={id} className="text-lg font-semibold mb-2 mt-3">
                        {children}
                    </h5>
                );
            },
            // Style tables
            table: ({ children }: { children?: React.ReactNode }) => (
                <div className="overflow-x-auto my-4">
                    <table className="min-w-full border-collapse border border-border">
                        {children}
                    </table>
                </div>
            ),
            thead: ({ children }: { children?: React.ReactNode }) => (
                <thead className="bg-muted">{children}</thead>
            ),
            th: ({ children }: { children?: React.ReactNode }) => (
                <th className="border border-border px-4 py-2 text-left">
                    {children}
                </th>
            ),
            td: ({ children }: { children?: React.ReactNode }) => (
                <td className="border border-border px-4 py-2">{children}</td>
            ),
            // Style lists
            ul: ({ children }: { children?: React.ReactNode }) => (
                <ul className="list-disc list-inside mb-4 space-y-1">
                    {children}
                </ul>
            ),
            ol: ({ children }: { children?: React.ReactNode }) => (
                <ol className="list-decimal list-inside mb-4 space-y-1">
                    {children}
                </ol>
            ),
            // Style blockquotes
            blockquote: ({ children }: { children?: React.ReactNode }) => (
                <blockquote className="border-l-4 border-primary pl-4 italic my-4">
                    {children}
                </blockquote>
            ),
        }),
        [],
    );

    if (loading) {
        return (
            <Container className="py-8">
                <p className="text-center text-muted-foreground">
                    Loading API documentation...
                </p>
            </Container>
        );
    }

    if (error) {
        return (
            <Container className="py-8">
                <Card>
                    <CardBody>
                        <p className="text-center text-destructive">{error}</p>
                        <p className="text-center text-sm text-muted-foreground mt-2">
                            Make sure to generate API docs first by running:
                            <code className="block mt-2 p-2 bg-muted rounded">
                                npm run openapi:generate-docs
                            </code>
                        </p>
                    </CardBody>
                </Card>
            </Container>
        );
    }

    return (
        <>
            <SEO
                title="API Documentation"
                description="Complete API reference documentation for Clipper platform"
                canonicalUrl="/admin/api-docs"
            />
            <Container className="py-8 max-w-7xl">
                {/* Header with controls */}
                <Card className="mb-6">
                    <CardHeader>
                        <div className="flex flex-col gap-4">
                            <div className="flex justify-between items-center">
                                <h1 className="text-3xl font-bold">
                                    📚 API Documentation
                                </h1>
                                <a
                                    href="/docs/openapi/openapi.yaml"
                                    target="_blank"
                                    rel="noopener noreferrer"
                                    className="text-sm text-primary hover:underline"
                                >
                                    📄 View OpenAPI Spec
                                </a>
                            </div>

                            {/* Version Selector */}
                            <div className="flex gap-4 items-center flex-wrap">
                                <label
                                    htmlFor="version-select"
                                    className="text-sm font-medium"
                                >
                                    Version:
                                </label>
                                <select
                                    id="version-select"
                                    value={selectedVersion}
                                    onChange={(e) =>
                                        setSelectedVersion(e.target.value)
                                    }
                                    className="px-3 py-2 bg-muted border border-border rounded-lg focus:outline-none focus:ring-2 focus:ring-primary"
                                >
                                    {API_VERSIONS.map((version) => (
                                        <option
                                            key={version.version}
                                            value={version.version}
                                        >
                                            {version.label}
                                        </option>
                                    ))}
                                </select>
                            </div>

                            {/* Search and Filter */}
                            <div className="flex gap-4 flex-col sm:flex-row">
                                <input
                                    type="text"
                                    placeholder="Search endpoints..."
                                    value={searchQuery}
                                    onChange={(e) =>
                                        setSearchQuery(e.target.value)
                                    }
                                    aria-label="Search API endpoints"
                                    className="flex-1 px-4 py-2 bg-muted border border-border rounded-lg focus:outline-none focus:ring-2 focus:ring-primary"
                                />
                                <label
                                    htmlFor="category-filter"
                                    className="sr-only"
                                >
                                    Filter by category
                                </label>
                                <select
                                    id="category-filter"
                                    value={selectedTag}
                                    onChange={(e) =>
                                        setSelectedTag(e.target.value)
                                    }
                                    className="px-3 py-2 bg-muted border border-border rounded-lg focus:outline-none focus:ring-2 focus:ring-primary"
                                >
                                    {tags.map((tag) => (
                                        <option key={tag} value={tag}>
                                            {tag === 'all'
                                                ? 'All Categories'
                                                : tag}
                                        </option>
                                    ))}
                                </select>
                            </div>
                        </div>
                    </CardHeader>
                </Card>

                {/* API Documentation Content */}
                <Card>
                    <CardBody className="prose prose-invert max-w-none">
                        <ReactMarkdown
                            remarkPlugins={[remarkGfm]}
                            components={markdownComponents}
                        >
                            {filteredContent || content}
                        </ReactMarkdown>
                    </CardBody>
                </Card>

                {/* Try-it-out Info */}
                <Card className="mt-6">
                    <CardHeader>
                        <h2 className="text-xl font-semibold">
                            🧪 Try It Out (Optional)
                        </h2>
                    </CardHeader>
                    <CardBody>
                        <p className="text-muted-foreground mb-4">
                            You can test API endpoints using the following
                            tools:
                        </p>
                        <ul className="space-y-2">
                            <li>
                                <strong>Swagger UI:</strong> Run{' '}
                                <code className="px-2 py-1 bg-muted rounded">
                                    npm run openapi:serve
                                </code>{' '}
                                and visit{' '}
                                <a
                                    href="http://localhost:8081"
                                    target="_blank"
                                    rel="noopener noreferrer"
                                    className="text-primary hover:underline"
                                >
                                    http://localhost:8081
                                </a>
                            </li>
                            <li>
                                <strong>Postman:</strong> Import the OpenAPI
                                spec from{' '}
                                <code className="px-2 py-1 bg-muted rounded">
                                    docs/openapi/openapi.yaml
                                </code>
                            </li>
                            <li>
                                <strong>Insomnia:</strong> Import the spec for
                                organized API testing
                            </li>
                            <li>
                                <strong>cURL:</strong> Copy code samples from
                                the documentation above
                            </li>
                        </ul>
                    </CardBody>
                </Card>
            </Container>
        </>
    );
}
