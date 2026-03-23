import type { PlaylistClipRef } from '@/types/playlist';
import { ListMusic } from 'lucide-react';

interface PlaylistThumbnailProps {
    clips?: PlaylistClipRef[];
    className?: string;
}

export function PlaylistThumbnail({
    clips,
    className = '',
}: PlaylistThumbnailProps) {
    // Get first 4 clips for the grid
    const thumbnailClips = (clips || [])
        .slice(0, 4)
        .filter(c => c.thumbnail_url);

    // If no clips with thumbnails, show gradient with music note
    if (thumbnailClips.length === 0) {
        return (
            <div
                className={`bg-gradient-to-br from-purple-900/20 to-blue-900/20 flex items-center justify-center ${className}`}
            >
                <div className='text-text-tertiary'><ListMusic size={48} strokeWidth={1.5} /></div>
            </div>
        );
    }

    // Show collage of clip thumbnails
    const gridClass =
        thumbnailClips.length === 1 ? 'grid-cols-1'
        : thumbnailClips.length === 2 ? 'grid-cols-2 grid-rows-2'
        : 'grid-cols-2 grid-rows-2';

    return (
        <div
            className={`bg-black grid ${gridClass} gap-0.5 overflow-hidden ${className}`}
        >
            {thumbnailClips.map((clip, idx) => (
                <div
                    key={`${clip.id}-${idx}`}
                    className='w-full h-full overflow-hidden'
                >
                    <img
                        src={clip.thumbnail_url}
                        alt={clip.title}
                        className='w-full h-full object-cover'
                        loading='lazy'
                    />
                </div>
            ))}
        </div>
    );
}
