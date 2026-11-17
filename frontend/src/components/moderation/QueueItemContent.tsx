import type { ModerationQueueItem } from '../../lib/moderation-api';
import { safeContentUrl } from './queueUtils';

interface QueueItemContentProps {
    item: ModerationQueueItem;
    /** Label for the ID shown when the API sends no excerpt. */
    idLabel: string;
}

/**
 * Shows what was flagged: the excerpt, its author and a link to it.
 * Falls back to the content ID when the API sends no excerpt.
 */
export function QueueItemContent({ item, idLabel }: QueueItemContentProps) {
    const excerpt = item.content_excerpt?.trim();
    const url = safeContentUrl(item.content_url);

    return (
        <>
            {excerpt ? (
                <blockquote className='border-border text-foreground pl-3 border-l-2 whitespace-pre-wrap break-words'>
                    {excerpt}
                </blockquote>
            ) : (
                <p>
                    <span className='font-medium'>{idLabel}:</span>{' '}
                    {item.content_id}
                </p>
            )}
            {item.content_author && (
                <p>
                    <span className='font-medium'>Posted by:</span>{' '}
                    {item.content_author}
                </p>
            )}
            {url && (
                <p>
                    <a
                        href={url}
                        target='_blank'
                        rel='noopener noreferrer'
                        className='text-primary hover:underline'
                    >
                        Open {item.content_type}
                    </a>
                </p>
            )}
        </>
    );
}
