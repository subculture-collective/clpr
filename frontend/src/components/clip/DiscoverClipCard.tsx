import { Badge } from '@/components/ui';
import { cn, formatCompactNumber, formatDuration, formatTimestamp } from '@/lib/utils';
import type { Clip } from '@/types/clip';
import { Link, useNavigate } from 'react-router-dom';
import { TwitchEmbed } from './TwitchEmbed';
import { ShareButton } from './ShareButton';
import { useIsAuthenticated, useToast } from '@/hooks';
import { Plus, Eye, ExternalLink } from 'lucide-react';

interface DiscoverClipCardProps {
    clip: Clip;
}

/**
 * DiscoverClipCard - A simplified clip card for the discover page
 *
 * This card focuses on clip discovery and posting, without the full
 * social features (voting, comments, favorites) that are available
 * on posted clips.
 */
export function DiscoverClipCard({ clip }: DiscoverClipCardProps) {
    const isAuthenticated = useIsAuthenticated();
    const navigate = useNavigate();
    const toast = useToast();


    const timestamp = formatTimestamp(clip.created_at);

    const handlePostClip = () => {
        if (!isAuthenticated) {
            toast.info('Please log in to post clips');
            return;
        }
        // Navigate to submit page with the clip URL prefilled
        navigate(
            `/submit?url=${encodeURIComponent(clip.twitch_clip_url)}&from=discover`,
            {
                state: { clipUrl: clip.twitch_clip_url, fromDiscover: true },
            },
        );
    };

    return (
        <div
            className='bg-card border-border rounded-xl hover:shadow-lg transition-shadow border'
            data-testid='discover-clip-card'
        >
            <div className='p-3 xs:p-4'>
                {/* Thumbnail/Embed */}
                <div className='relative mb-3'>
                    <TwitchEmbed
                        clipId={clip.twitch_clip_id}
                        thumbnailUrl={clip.thumbnail_url}
                        title={clip.title}
                    />

                    {/* Duration badge */}
                    {clip.duration && (
                        <div className='bottom-2 right-2 absolute px-2 py-1 text-xs font-medium text-white bg-black bg-opacity-75 rounded'>
                            {formatDuration(clip.duration)}
                        </div>
                    )}

                    {/* NSFW badge */}
                    {clip.is_nsfw && (
                        <div className='top-2 left-2 absolute'>
                            <Badge variant='error'>NSFW</Badge>
                        </div>
                    )}
                </div>

                {/* Title */}
                <h3 className='line-clamp-2 text-base xs:text-lg font-semibold leading-snug mb-2'>
                    {clip.title}
                </h3>

                {/* Metadata */}
                <div className='text-muted-foreground flex flex-wrap items-center gap-1.5 xs:gap-2 mb-3 text-xs xs:text-sm leading-tight'>
                    <span className='flex items-center gap-1 font-medium'>
                        <Link
                            to={`/broadcaster/${clip.broadcaster_id || clip.broadcaster_name}`}
                            className='hover:text-foreground transition-colors cursor-pointer'
                        >
                            {clip.broadcaster_name}
                        </Link>
                    </span>

                    {clip.game_name && (
                        <>
                            <span className='hidden xs:inline'>•</span>
                            <span className='flex items-center gap-1'>
                                <Link
                                    to={`/game/${clip.game_id}`}
                                    className='hover:text-foreground transition-colors cursor-pointer'
                                >
                                    {clip.game_name}
                                </Link>
                            </span>
                        </>
                    )}

                    <span className='hidden xs:inline'>•</span>
                    <span
                        className='truncate align-middle'
                        title={timestamp.title}
                    >
                        {timestamp.display}
                    </span>
                </div>

                {/* Action bar */}
                <div className='flex flex-wrap items-center gap-2 xs:gap-3'>
                    {/* View count */}
                    <span className='text-muted-foreground flex items-center gap-1 text-xs xs:text-sm'>
                        <Eye className='w-4 h-4' />
                        <span>{formatCompactNumber(clip.view_count)} views</span>
                    </span>

                    <div className='flex-1' />

                    {/* Share button */}
                    <ShareButton clipId={clip.id} clipTitle={clip.title} />

                    {/* Watch on Twitch */}
                    <a
                        href={clip.twitch_clip_url}
                        target='_blank'
                        rel='noopener noreferrer'
                        className='text-muted-foreground hover:text-foreground flex items-center gap-1.5 transition-colors text-xs xs:text-sm'
                    >
                        <ExternalLink className='w-4 h-4' />
                        <span className='hidden sm:inline'>Twitch</span>
                    </a>

                    {/* Post button */}
                    <button
                        onClick={handlePostClip}
                        className={cn(
                            'flex items-center gap-1.5 px-3 py-1.5 rounded-lg text-xs xs:text-sm font-medium transition-colors',
                            'bg-primary-500 text-white hover:bg-primary-600',
                            !isAuthenticated && 'opacity-70',
                        )}
                    >
                        <Plus className='w-4 h-4' />
                        Post Clip
                    </button>
                </div>
            </div>
        </div>
    );
}
