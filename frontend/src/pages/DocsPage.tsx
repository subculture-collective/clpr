import { useState, useEffect, useCallback, useMemo } from 'react';
import { useSearchParams } from 'react-router-dom';
import ReactMarkdown from 'react-markdown';
import remarkGfm from 'remark-gfm';
import {
    Container,
    Card,
    CardBody,
    SEO,
    DocHeader,
    DocTOC,
} from '../components';
import {
    parseMarkdown,
    convertWikilinks,
    headingToId,
    extractTextFromChildren,
} from '../lib/markdown-utils';
import type { ProcessedMarkdown } from '../lib/markdown-utils';
import { PenLine } from 'lucide-react';
import axios from 'axios';

interface DocNode {
    name: string;
    path: string;
    type: 'file' | 'directory';
    children?: DocNode[];
}

interface DocContent {
    path: string;
    content: string;
    github_url?: string;
}

interface SearchResult {
    path: string;
    name: string;
    matches: string[];
    score: number;
}

/**
 * Documentation Hub Page
 * Displays documentation served from the backend /docs folder
 */
export function DocsPage() {
    const [searchParams] = useSearchParams();
    const [docs, setDocs] = useState<DocNode[]>([]);
    const [selectedDoc, setSelectedDoc] = useState<DocContent | null>(null);
    const [processedDoc, setProcessedDoc] = useState<ProcessedMarkdown | null>(
        null,
    );
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState<string | null>(null);
    const [viewingDoc, setViewingDoc] = useState(false);
    const [searchQuery, setSearchQuery] = useState('');
    const [searchResults, setSearchResults] = useState<SearchResult[]>([]);
    const [searching, setSearching] = useState(false);

    const fetchDocsList = useCallback(async () => {
        try {
            const response = await axios.get('/api/v1/docs');
            setDocs(response.data.docs || []);
            setLoading(false);
        } catch (err) {
            console.error('Failed to load documentation', err);
            setError('Failed to load documentation');
            setLoading(false);
        }
    }, []);

    const fetchDoc = useCallback(async (path: string) => {
        try {
            setLoading(true);
            const response = await axios.get(`/api/v1/docs/${path}`);
            setSelectedDoc(response.data);

            // Process markdown: parse frontmatter, remove doctoc, handle Dataview, generate TOC
            const processed = parseMarkdown(response.data.content);

            // Convert wikilinks to regular links
            processed.content = convertWikilinks(processed.content);

            setProcessedDoc(processed);
            setViewingDoc(true);
            setLoading(false);
            window.scrollTo(0, 0);
        } catch (err) {
            console.error('Failed to load document', err);
            setError('Failed to load document');
            setLoading(false);
        }
    }, []);

    useEffect(() => {
        // eslint-disable-next-line react-hooks/set-state-in-effect
        fetchDocsList();
    }, [fetchDocsList]);

    useEffect(() => {
        // Check for doc parameter in URL
        const docParam = searchParams.get('doc');
        if (docParam) {
            // eslint-disable-next-line react-hooks/set-state-in-effect
            fetchDoc(docParam);
        }
    }, [searchParams, fetchDoc]);

    const handleBackToIndex = () => {
        setViewingDoc(false);
        setSelectedDoc(null);
        setProcessedDoc(null);
        setSearchQuery('');
        setSearchResults([]);
    };

    const handleSearch = async (query: string) => {
        setSearchQuery(query);

        if (!query.trim()) {
            setSearchResults([]);
            return;
        }

        try {
            setSearching(true);
            const response = await axios.get(
                `/api/v1/docs/search?q=${encodeURIComponent(query)}`,
            );
            setSearchResults(response.data.results || []);
            setSearching(false);
        } catch (err) {
            console.error('Search failed:', err);
            setSearching(false);
        }
    };

    const renderDocTree = (nodes: DocNode[], level = 0) => {
        if (!nodes || !Array.isArray(nodes)) {
            return null;
        }

        return (
            <div className={level > 0 ? 'ml-4' : ''}>
                {nodes.map(node => (
                    <div key={node.path} className='mb-2'>
                        {node.type === 'directory' ?
                            <>
                                <h3
                                    className={`font-semibold ${
                                        level === 0 ? 'text-xl mt-4 mb-2' : (
                                            'text-lg mt-2 mb-1'
                                        )
                                    }`}
                                >
                                    {node.name.charAt(0).toUpperCase() +
                                        node.name.slice(1).replace(/-/g, ' ')}
                                </h3>
                                {node.children &&
                                    renderDocTree(node.children, level + 1)}
                            </>
                        :   <button
                                onClick={() => fetchDoc(node.path)}
                                className='text-left text-primary hover:underline block py-1'
                            >
                                {node.name.replace(/-/g, ' ')}
                            </button>
                        }
                    </div>
                ))}
            </div>
        );
    };

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
            }) => {
                // Handle relative doc links (wikilinks are already converted)
                if (href && !href.startsWith('http') && !href.startsWith('#')) {
                    const cleanPath = href
                        .replace(/^\.\.\//, '')
                        .replace(/\.md$/, '')
                        .replace(/^\//, '');
                    return (
                        <button
                            onClick={() => fetchDoc(cleanPath)}
                            className='text-primary hover:underline'
                        >
                            {children}
                        </button>
                    );
                }

                // External links
                return (
                    <a
                        href={href}
                        target={href?.startsWith('http') ? '_blank' : undefined}
                        rel={
                            href?.startsWith('http') ?
                                'noopener noreferrer'
                            :   undefined
                        }
                        className='text-primary hover:underline'
                    >
                        {children}
                    </a>
                );
            },
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
                        <code className='px-1 py-0.5 bg-muted rounded text-sm'>
                            {children}
                        </code>
                    );
                }
                return (
                    <code
                        className={`block p-4 bg-muted rounded-lg overflow-x-auto ${
                            className || ''
                        }`}
                    >
                        {children}
                    </code>
                );
            },
            // Style headings with IDs for TOC navigation
            h1: ({ children }: { children?: React.ReactNode }) => {
                const text = extractTextFromChildren(children);
                const id = headingToId(text);
                return (
                    <h1 id={id} className='text-4xl font-bold mb-4 mt-6'>
                        {children}
                    </h1>
                );
            },
            h2: ({ children }: { children?: React.ReactNode }) => {
                const text = extractTextFromChildren(children);
                const id = headingToId(text);
                return (
                    <h2 id={id} className='text-3xl font-semibold mb-3 mt-5'>
                        {children}
                    </h2>
                );
            },
            h3: ({ children }: { children?: React.ReactNode }) => {
                const text = extractTextFromChildren(children);
                const id = headingToId(text);
                return (
                    <h3 id={id} className='text-2xl font-semibold mb-2 mt-4'>
                        {children}
                    </h3>
                );
            },
            h4: ({ children }: { children?: React.ReactNode }) => {
                const text = extractTextFromChildren(children);
                const id = headingToId(text);
                return (
                    <h4 id={id} className='text-xl font-semibold mb-2 mt-3'>
                        {children}
                    </h4>
                );
            },
            h5: ({ children }: { children?: React.ReactNode }) => {
                const text = extractTextFromChildren(children);
                const id = headingToId(text);
                return (
                    <h5 id={id} className='text-lg font-semibold mb-2 mt-3'>
                        {children}
                    </h5>
                );
            },
            h6: ({ children }: { children?: React.ReactNode }) => {
                const text = extractTextFromChildren(children);
                const id = headingToId(text);
                return (
                    <h6 id={id} className='text-base font-semibold mb-2 mt-3'>
                        {children}
                    </h6>
                );
            },
            // Style lists
            ul: ({ children }: { children?: React.ReactNode }) => (
                <ul className='list-disc list-inside mb-4 space-y-1'>
                    {children}
                </ul>
            ),
            ol: ({ children }: { children?: React.ReactNode }) => (
                <ol className='list-decimal list-inside mb-4 space-y-1'>
                    {children}
                </ol>
            ),
            // Style blockquotes
            blockquote: ({ children }: { children?: React.ReactNode }) => (
                <blockquote className='border-l-4 border-primary pl-4 italic my-4'>
                    {children}
                </blockquote>
            ),
        }),
        [fetchDoc],
    );

    if (loading && !viewingDoc) {
        return (
            <Container className='py-8'>
                <p className='text-center text-muted-foreground'>
                    Loading documentation...
                </p>
            </Container>
        );
    }

    if (error) {
        return (
            <Container className='py-8'>
                <Card>
                    <CardBody>
                        <p className='text-center text-destructive'>{error}</p>
                    </CardBody>
                </Card>
            </Container>
        );
    }

    return (
        <>
            <SEO
                title={
                    viewingDoc && selectedDoc ?
                        selectedDoc.path
                    :   'Documentation'
                }
                description='Comprehensive documentation for Clipper - architecture, APIs, operations, user guides, and contributor information.'
                canonicalUrl='/docs'
            />
            <Container className='py-8 max-w-6xl'>
                {
                    viewingDoc && selectedDoc && processedDoc ?
                        // Document viewer
                        <div>
                            <div className='flex justify-between items-center mb-4'>
                                <button
                                    onClick={handleBackToIndex}
                                    className='text-primary hover:underline flex items-center gap-2'
                                >
                                    ← Back to Documentation Index
                                </button>
                                {selectedDoc.github_url && (
                                    <a
                                        href={selectedDoc.github_url}
                                        target='_blank'
                                        rel='noopener noreferrer'
                                        className='text-sm text-primary hover:underline flex items-center gap-1'
                                    >
                                        <PenLine size={14} strokeWidth={1.75} className='inline mr-1' /> Edit on GitHub
                                    </a>
                                )}
                            </div>
                            <Card>
                                <CardBody className='prose prose-invert max-w-none'>
                                    {/* Render frontmatter as DocHeader */}
                                    <DocHeader
                                        frontmatter={processedDoc.frontmatter}
                                    />

                                    {/* Render TOC if available */}
                                    {processedDoc.toc.length > 0 && (
                                        <DocTOC toc={processedDoc.toc} />
                                    )}

                                    {/* Render processed markdown content */}
                                    <ReactMarkdown
                                        remarkPlugins={[remarkGfm]}
                                        components={markdownComponents}
                                    >
                                        {processedDoc.content}
                                    </ReactMarkdown>
                                </CardBody>
                            </Card>
                        </div>
                        // Documentation index
                    :   <div>
                            <div className='mb-8'>
                                <h1 className='text-4xl font-bold mb-4'>
                                    Documentation Hub
                                </h1>
                                <p className='text-lg text-muted-foreground mb-4'>
                                    Comprehensive guides, API references, and
                                    operational procedures
                                </p>

                                {/* Search Bar */}
                                <div className='relative'>
                                    <input
                                        type='text'
                                        placeholder='Search documentation...'
                                        value={searchQuery}
                                        onChange={e =>
                                            handleSearch(e.target.value)
                                        }
                                        className='w-full px-4 py-3 bg-muted border border-border rounded-lg focus:outline-none focus:ring-2 focus:ring-primary'
                                    />
                                    {searching && (
                                        <div className='absolute right-3 top-3 text-muted-foreground'>
                                            Searching...
                                        </div>
                                    )}
                                </div>
                            </div>

                            {/* Search Results */}
                            {searchQuery && searchResults.length > 0 && (
                                <Card className='mb-8'>
                                    <CardBody>
                                        <h2 className='text-2xl font-semibold mb-4'>
                                            Search Results (
                                            {searchResults.length})
                                        </h2>
                                        <div className='space-y-4'>
                                            {searchResults.map(result => (
                                                <div
                                                    key={result.path}
                                                    className='border-b border-border pb-4 last:border-0'
                                                >
                                                    <button
                                                        onClick={() => {
                                                            fetchDoc(
                                                                result.path,
                                                            );
                                                            setSearchQuery('');
                                                            setSearchResults(
                                                                [],
                                                            );
                                                        }}
                                                        className='text-left w-full hover:bg-accent p-2 rounded transition-colors'
                                                    >
                                                        <h3 className='font-semibold text-primary mb-1'>
                                                            {result.name}
                                                        </h3>
                                                        <p className='text-sm text-muted-foreground mb-2'>
                                                            {result.path}
                                                        </p>
                                                        {result.matches.map(
                                                            (match, idx) => (
                                                                <p
                                                                    key={idx}
                                                                    className='text-sm text-muted-foreground italic'
                                                                >
                                                                    {match}
                                                                </p>
                                                            ),
                                                        )}
                                                    </button>
                                                </div>
                                            ))}
                                        </div>
                                    </CardBody>
                                </Card>
                            )}

                            {/* No Results */}
                            {searchQuery &&
                                searchResults.length === 0 &&
                                !searching && (
                                    <Card className='mb-8'>
                                        <CardBody>
                                            <p className='text-center text-muted-foreground'>
                                                No results found for "
                                                {searchQuery}"
                                            </p>
                                        </CardBody>
                                    </Card>
                                )}

                            {/* Documentation Tree (show only when not searching) */}
                            {!searchQuery && (
                                <Card>
                                    <CardBody>{renderDocTree(docs)}</CardBody>
                                </Card>
                            )}

                            {/* Additional Resources */}
                            <Card className='mt-8'>
                                <CardBody>
                                    <h2 className='text-2xl font-semibold mb-4'>
                                        External Resources
                                    </h2>
                                    <div className='space-y-3'>
                                        <div>
                                            <a
                                                href='https://github.com/subculture-collective/clipper'
                                                target='_blank'
                                                rel='noopener noreferrer'
                                                className='text-primary hover:underline font-medium'
                                            >
                                                GitHub Repository
                                            </a>
                                            <p className='text-sm text-muted-foreground'>
                                                View source code and open issues
                                            </p>
                                        </div>
                                        <div>
                                            <a
                                                href='https://github.com/subculture-collective/clipper/issues'
                                                target='_blank'
                                                rel='noopener noreferrer'
                                                className='text-primary hover:underline font-medium'
                                            >
                                                Issue Tracker
                                            </a>
                                            <p className='text-sm text-muted-foreground'>
                                                Report bugs or request features
                                            </p>
                                        </div>
                                        <div>
                                            <a
                                                href='https://github.com/subculture-collective/clipper/discussions'
                                                target='_blank'
                                                rel='noopener noreferrer'
                                                className='text-primary hover:underline font-medium'
                                            >
                                                Discussions
                                            </a>
                                            <p className='text-sm text-muted-foreground'>
                                                Ask questions and share ideas
                                            </p>
                                        </div>
                                    </div>
                                </CardBody>
                            </Card>
                        </div>

                }
            </Container>
        </>
    );
}
