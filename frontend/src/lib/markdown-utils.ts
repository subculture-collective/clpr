/**
 * Markdown utilities for documentation rendering
 * Handles frontmatter parsing, doctoc removal, Dataview blocks, and TOC generation
 */

/**
 * Frontmatter metadata interface
 */
export interface DocFrontmatter {
    title?: string;
    summary?: string;
    tags?: string[];
    area?: string;
    status?: string;
    owner?: string;
    version?: string;
    last_reviewed?: string;
    aliases?: string[];
    [key: string]: unknown;
}

/**
 * Processed markdown result
 */
export interface ProcessedMarkdown {
    content: string;
    frontmatter: DocFrontmatter;
    toc: TOCEntry[];
}

/**
 * Table of Contents entry
 */
export interface TOCEntry {
    level: number;
    text: string;
    id: string;
    children?: TOCEntry[];
}

function parseFrontmatterValue(value: string): unknown {
    const trimmed = value.trim();
    if (trimmed.startsWith('[') && trimmed.endsWith(']')) {
        const entries = trimmed.slice(1, -1).trim();
        return entries ? entries.split(',').map((entry) => parseFrontmatterValue(entry)) : [];
    }
    if (
        (trimmed.startsWith('"') && trimmed.endsWith('"')) ||
        (trimmed.startsWith("'") && trimmed.endsWith("'"))
    ) {
        return trimmed.slice(1, -1);
    }
    if (trimmed === 'true') return true;
    if (trimmed === 'false') return false;
    if (trimmed === 'null') return null;
    if (/^-?\d+(?:\.\d+)?$/.test(trimmed)) return Number(trimmed);
    return trimmed;
}

function extractFrontmatter(markdown: string): {
    data: DocFrontmatter;
    content: string;
} {
    if (!markdown.startsWith('---\n')) return { data: {}, content: markdown };
    const closingDelimiter = markdown.indexOf('\n---', 4);
    if (closingDelimiter === -1) return { data: {}, content: markdown };

    const data: DocFrontmatter = {};
    const block = markdown.slice(4, closingDelimiter);
    for (const line of block.split('\n')) {
        const separator = line.indexOf(':');
        if (separator <= 0) continue;
        const key = line.slice(0, separator).trim();
        if (!/^[A-Za-z_][A-Za-z0-9_-]*$/.test(key)) continue;
        data[key] = parseFrontmatterValue(line.slice(separator + 1));
    }
    return {
        data,
        content: markdown.slice(closingDelimiter + 4).replace(/^\r?\n/, ''),
    };
}

/**
 * Parse frontmatter from markdown and return processed content
 */
export function parseMarkdown(markdown: string): ProcessedMarkdown {
    // Parse frontmatter
    const { data, content } = extractFrontmatter(markdown);

    // Remove doctoc blocks
    const contentWithoutDoctoc = removeDoctocBlocks(content);

    // Process Dataview blocks
    const processedContent = processDataviewBlocks(contentWithoutDoctoc);

    // Generate TOC from headings
    const toc = generateTOC(processedContent);

    return {
        content: processedContent,
        frontmatter: data as DocFrontmatter,
        toc,
    };
}

/**
 * Remove doctoc comment blocks from markdown
 * Removes everything between <!-- START doctoc ... --> and <!-- END doctoc -->
 */
export function removeDoctocBlocks(markdown: string): string {
    return markdown.replace(
        /<!-- START doctoc[\s\S]*?<!-- END doctoc[^\n]*-->/g,
        ''
    ).trim();
}

// Dataview callout constants
const DATAVIEW_CALLOUT_TITLE = 'Obsidian Dataview Query';
const DATAVIEW_CALLOUT_MESSAGE = 'This section uses Dataview queries that are only visible in Obsidian.';

/**
 * Process Dataview blocks and convert them to "Obsidian-only" callouts
 * Dataview blocks are not executed, just rendered as informational callouts
 */
export function processDataviewBlocks(markdown: string): string {
    // Match dataview code blocks: ```dataview ... ```
    return markdown.replace(
        /```dataview\n([\s\S]*?)```/g,
        (_, query) => {
            return `> [!note] ${DATAVIEW_CALLOUT_TITLE}\n> ${DATAVIEW_CALLOUT_MESSAGE}\n> \n> \`\`\`dataview\n> ${query.trim()}\n> \`\`\``;
        }
    );
}

/**
 * Generate table of contents from markdown headings
 */
export function generateTOC(markdown: string): TOCEntry[] {
    const headings: TOCEntry[] = [];
    const headingRegex = /^(#{1,6})\s+(.+)$/gm;
    let match;

    while ((match = headingRegex.exec(markdown)) !== null) {
        const level = match[1].length;
        const text = match[2].trim();

        // Generate ID from heading text (GitHub-style)
        const id = headingToId(text);

        headings.push({
            level,
            text,
            id,
        });
    }

    return buildTOCTree(headings);
}

/**
 * Convert heading text to anchor ID (GitHub-style)
 */
export function headingToId(heading: string): string {
    return heading
        .toLowerCase()
        .replace(/[^\w\s-]/g, '') // Remove special chars
        .replace(/\s+/g, '-')      // Replace spaces with hyphens
        .replace(/^-+|-+$/g, '');  // Trim hyphens
}

/**
 * Extract text content from React children (handles strings and nested elements)
 */
export function extractTextFromChildren(children: React.ReactNode): string {
    if (typeof children === 'string') {
        return children;
    }
    if (Array.isArray(children)) {
        return children.map(extractTextFromChildren).join('');
    }
    if (children && typeof children === 'object' && 'props' in children) {
        return extractTextFromChildren((children as { props: { children: React.ReactNode } }).props.children);
    }
    return '';
}

/**
 * Build hierarchical TOC tree from flat list of headings
 */
function buildTOCTree(headings: TOCEntry[]): TOCEntry[] {
    if (headings.length === 0) return [];

    const root: TOCEntry[] = [];
    const stack: TOCEntry[] = [];

    for (const heading of headings) {
        // Find the appropriate parent
        while (stack.length > 0 && stack[stack.length - 1].level >= heading.level) {
            stack.pop();
        }

        if (stack.length === 0) {
            // Top-level heading
            root.push(heading);
        } else {
            // Child heading
            const parent = stack[stack.length - 1];
            if (!parent.children) {
                parent.children = [];
            }
            parent.children.push(heading);
        }

        stack.push(heading);
    }

    return root;
}

/**
 * Convert wikilinks to markdown links
 * [[page]] -> [page](page)
 * [[page|alias]] -> [alias](page)
 * [[page#anchor]] -> [page](page#anchor)
 */
export function convertWikilinks(markdown: string): string {
    return markdown.replace(
        /\[\[([^\]|]+)(?:\|([^\]]+))?\]\]/g,
        (_, target, alias) => {
            const linkText = alias || target;
            // Remove .md extension if present
            const linkTarget = target.replace(/\.md$/, '');
            return `[${linkText}](${linkTarget})`;
        }
    );
}
