import React from 'react';
import { Link } from 'react-router-dom';
import { cn } from '@/lib/utils';
import { TAG_EVIDENCE, TAG_LANES, tagHref, tagLabel, tagLane } from '@/lib/tag-lanes';
import type { Tag, TagLane } from '../../types/tag';

interface TagChipProps {
    tag: Tag;
    size?: 'small' | 'medium';
    removable?: boolean;
    onRemove?: (slug: string) => void;
    onClick?: (slug: string) => void;
    /** Show the usage count after the label. */
    showCount?: boolean;
}

const laneClasses: Record<TagLane, string> = {
    detected: 'border-line-strong text-foreground hover:border-text-tertiary',
    category: 'border-category/40 text-category hover:border-category',
    community: 'border-primary-800 text-link hover:border-primary-400',
    streamer: 'border-dashed border-line-strong text-text-tertiary hover:text-text-secondary',
    duration: 'border-border text-text-secondary',
    language: 'border-border text-text-secondary',
    root: 'border-border text-text-secondary',
};

const evidenceClasses = {
    visible: 'bg-seen text-background',
    contextual: 'bg-context text-background',
    strong: 'bg-primary-400 text-background',
} as const;

export const TagChip: React.FC<TagChipProps> = ({
    tag,
    size = 'medium',
    removable = false,
    onRemove,
    onClick,
    showCount = false,
}) => {
    const lane = tagLane(tag);
    const label = tagLabel(tag);
    const evidence = lane === 'detected' ? tag.evidence : undefined;
    const title = [
        TAG_LANES[lane].label,
        evidence ? `${TAG_EVIDENCE[evidence].label}: ${TAG_EVIDENCE[evidence].description}` : TAG_LANES[lane].description,
    ].join(' · ');
    const accessibleName =
        evidence ? `${label}, ${TAG_EVIDENCE[evidence].label.toLowerCase()}`
        : lane === 'community' ? `#${label}`
        : lane === 'streamer' ? `${label}, streamer tag`
        : lane === 'category' ? `${label}, Twitch category`
        : label;

    const chipClasses = cn(
        'inline-flex items-center gap-1.5 border font-mono font-medium leading-none transition-colors cursor-pointer',
        'focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-focus-ring',
        size === 'small' ? 'min-h-6 px-1.5 text-[11px]' : 'min-h-8 px-2 text-xs',
        laneClasses[lane],
    );

    const content = (
        <>
            <span>{lane === 'community' ? `#${label}` : label}</span>
            {evidence && (
                <span className={cn('px-1 py-0.5 text-[9px] font-semibold uppercase tracking-[0.06em]', evidenceClasses[evidence])}>
                    {TAG_EVIDENCE[evidence].short}
                </span>
            )}
            {showCount && tag.usage_count > 0 && (
                <span className='tabular-nums text-text-tertiary'>{tag.usage_count.toLocaleString()}</span>
            )}
        </>
    );

    const handleClick = (e: React.MouseEvent) => {
        if (onClick) {
            e.preventDefault();
            onClick(tag.slug);
        }
    };

    const handleRemove = (e: React.MouseEvent) => {
        e.preventDefault();
        e.stopPropagation();
        onRemove?.(tag.slug);
    };

    const chip = onClick ? (
        <button type='button' onClick={handleClick} className={chipClasses} title={title} data-lane={lane} aria-label={accessibleName}>
            {content}
        </button>
    ) : (
        <Link to={tagHref(tag.slug)} className={chipClasses} title={title} data-lane={lane} aria-label={accessibleName}>
            {content}
        </Link>
    );

    if (!removable) return chip;
    return (
        <span className='inline-flex items-center'>
            {chip}
            <button
                type='button'
                onClick={handleRemove}
                className='-ml-px flex min-h-8 min-w-8 items-center justify-center border border-line-strong text-text-secondary hover:bg-surface-hover hover:text-foreground'
                aria-label={`Remove ${label} tag`}
            >
                <svg aria-hidden='true' className='h-3 w-3' fill='none' stroke='currentColor' viewBox='0 0 24 24'>
                    <path strokeLinecap='round' strokeLinejoin='round' strokeWidth={2} d='M6 18L18 6M6 6l12 12' />
                </svg>
            </button>
        </span>
    );
};
