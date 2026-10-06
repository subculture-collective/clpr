import React from 'react';
import { Link } from 'react-router-dom';
import { cn } from '@/lib/utils';
import { tagAccent, tagHref, tagLabel } from '@/lib/tag-lanes';
import type { Tag } from '../../types/tag';

interface TagChipProps {
    tag: Tag;
    size?: 'small' | 'medium';
    removable?: boolean;
    onRemove?: (slug: string) => void;
    onClick?: (slug: string) => void;
    /** Show the usage count after the label. */
    showCount?: boolean;
}

export const TagChip: React.FC<TagChipProps> = ({
    tag,
    size = 'medium',
    removable = false,
    onRemove,
    onClick,
    showCount = false,
}) => {
    const label = tagLabel(tag);

    const chipClasses = cn(
        'inline-flex items-center gap-1.5 border font-mono font-medium leading-none transition-colors cursor-pointer',
        'focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-focus-ring',
        size === 'small' ? 'min-h-6 px-1.5 text-[11px]' : 'min-h-8 px-2 text-xs',
        'border-current/40 bg-current/10 hover:border-current/70 hover:bg-current/20',
        tagAccent(label),
    );

    const content = (
        <>
            <span>{label}</span>
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
        <button type='button' onClick={handleClick} className={chipClasses} aria-label={label}>
            {content}
        </button>
    ) : (
        <Link to={tagHref(tag.slug)} className={chipClasses} aria-label={label}>
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
