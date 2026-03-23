import React from 'react';
import { useClipTags } from '../../hooks/useTags';
import { TagChip } from './TagChip';

interface TagListProps {
    clipId: string;
    maxVisible?: number;
}

export const TagList: React.FC<TagListProps> = ({ clipId, maxVisible = 5 }) => {
    const { data, isLoading } = useClipTags(clipId);

    if (isLoading) {
        return (
            <div className='flex gap-2'>
                {[...Array(3)].map((_, i) => (
                    <div
                        key={i}
                        className='h-6 w-16 bg-surface-raised rounded-full animate-pulse'
                    />
                ))}
            </div>
        );
    }

    if (!data || !data.tags || data.tags.length === 0) {
        return null;
    }

    const visibleTags = data.tags.slice(0, maxVisible);
    const remainingCount = data.tags.length - maxVisible;

    return (
        <div className='flex flex-wrap gap-2 items-center'>
            {visibleTags.map((tag) => (
                <TagChip
                    key={tag.id}
                    tag={tag}
                    size='small'
                />
            ))}
            {remainingCount > 0 && (
                <span className='text-xs text-muted-foreground'>
                    +{remainingCount} more
                </span>
            )}
        </div>
    );
};
