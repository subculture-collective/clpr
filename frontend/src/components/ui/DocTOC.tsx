import type { TOCEntry } from '../../lib/markdown-utils';

interface DocTOCProps {
    toc: TOCEntry[];
    onNavigate?: (id: string) => void;
}

/**
 * DocTOC component renders a table of contents from headings
 * Generated at render-time, not stored in markdown
 */
export function DocTOC({ toc, onNavigate }: DocTOCProps) {
    if (toc.length === 0) {
        return null;
    }

    const handleClick = (e: React.MouseEvent, id: string) => {
        e.preventDefault();
        
        // Scroll to the heading
        const element = document.getElementById(id);
        if (element) {
            element.scrollIntoView({ behavior: 'smooth', block: 'start' });
            
            // Update URL hash
            window.history.pushState(null, '', `#${id}`);
        }
        
        onNavigate?.(id);
    };

    return (
        <nav className='mb-8 p-6 bg-muted rounded-lg border border-border'>
            <h2 className='text-lg font-semibold mb-4'>Table of Contents</h2>
            <TOCList entries={toc} onNavigate={handleClick} />
        </nav>
    );
}

interface TOCListProps {
    entries: TOCEntry[];
    onNavigate: (e: React.MouseEvent, id: string) => void;
    level?: number;
}

function TOCList({ entries, onNavigate, level = 0 }: TOCListProps) {
    if (entries.length === 0) return null;

    return (
        <ul className={`space-y-2 ${level > 0 ? 'ml-4 mt-2' : ''}`}>
            {entries.map((entry) => (
                <li key={entry.id}>
                    <a
                        href={`#${entry.id}`}
                        onClick={(e) => onNavigate(e, entry.id)}
                        className='text-link underline underline-offset-2 block'
                    >
                        {entry.text}
                    </a>
                    {entry.children && entry.children.length > 0 && (
                        <TOCList
                            entries={entry.children}
                            onNavigate={onNavigate}
                            level={level + 1}
                        />
                    )}
                </li>
            ))}
        </ul>
    );
}
