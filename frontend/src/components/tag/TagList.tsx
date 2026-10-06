import React from 'react';
import { useClipTags } from '../../hooks/useTags';
import { isChipLane, sortByPopularity, tagLane } from '@/lib/tag-lanes';
import { TagChip } from './TagChip';

interface TagListProps {
    clipId: string;
    maxVisible?: number;
}

/** Clip tags share one presentation; length and language stay in metadata. */
export const TagList: React.FC<TagListProps> = ({ clipId, maxVisible = 5 }) => {
    const { data, isLoading } = useClipTags(clipId);

    if (isLoading) {
        return (
            <div className='flex gap-1.5'>
                {[...Array(3)].map((_, i) => (
                    <div key={i} className='h-6 w-16 bg-surface-raised animate-pulse motion-reduce:animate-none' />
                ))}
            </div>
        );
    }

    const tags = sortByPopularity((data?.tags ?? []).filter(tag => isChipLane(tagLane(tag))));
    if (tags.length === 0) return null;

    const visibleTags = tags.slice(0, maxVisible);
    const remainingCount = tags.length - visibleTags.length;

    return (
        <div className='flex flex-wrap gap-1.5 items-center'>
            {visibleTags.map(tag => (
                <TagChip key={tag.id} tag={tag} size='small' />
            ))}
            {remainingCount > 0 && (
                <span className='font-mono text-[11px] text-text-tertiary'>+{remainingCount}</span>
            )}
        </div>
    );
};
