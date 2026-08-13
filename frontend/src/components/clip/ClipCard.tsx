import { TagList } from '@/components/tag/TagList';
import { Badge } from '@/components/ui';
import { VerifiedBadge } from '@/components/user';
import { useClipFavorite, useClipVote } from '@/hooks/useClips';
import { useIsAuthenticated, useMediaQuery, useToast } from '@/hooks';
import { cn, formatCompactNumber, formatDuration, formatTimestamp } from '@/lib/utils';
import type { Clip } from '@/types/clip';
import { Link } from 'react-router-dom';
import { TwitchEmbed } from './TwitchEmbed';
import { AddToPlaylistButton } from './AddToPlaylistButton';
import { AddToQueueButton } from './AddToQueueButton';
import { ShareButton } from './ShareButton';
import {
    ArrowBigUp,
    ArrowBigDown,
    MessageSquare,
    Heart,
    Eye,
    Check,
    MoreHorizontal,
} from 'lucide-react';
import { useEffect } from 'react';
import { useInView } from 'react-intersection-observer';

interface ClipCardProps {
    clip: Clip;
    active?: boolean;
    autoplay?: boolean;
    onActivate?: (clipId: string) => void;
    onVisibilityChange?: (clipId: string, visible: boolean) => void;
}

export function ClipCard({
    clip,
    active = false,
    autoplay = false,
    onActivate,
    onVisibilityChange,
}: ClipCardProps) {
    const isAuthenticated = useIsAuthenticated();
    const voteMutation = useClipVote();
    const favoriteMutation = useClipFavorite();
    const isVoting = voteMutation.isPending;
    const toast = useToast();
    const isDesktop = useMediaQuery('(min-width: 768px)');
    const { ref: visibilityRef, inView } = useInView({ threshold: 0.6 });

    useEffect(() => {
        onVisibilityChange?.(clip.id, inView);
    }, [clip.id, inView, onVisibilityChange]);

    const handleVote = (voteType: 1 | -1) => {
        // Avoid duplicate same-direction votes
        if (isVoting || clip.user_vote === voteType) return;
        if (!isAuthenticated) {
            toast.info('Please log in to vote on clips');
            return;
        }
        voteMutation.mutate({ clip_id: clip.id, vote_type: voteType });
    };

    const handleFavorite = () => {
        if (!isAuthenticated) {
            toast.info('Please log in to favorite clips');
            return;
        }
        favoriteMutation.mutate({ clip_id: clip.id });
    };

    const voteColor =
        clip.vote_score > 0 ? 'text-upvote'
        : clip.vote_score < 0 ? 'text-downvote'
        : 'text-muted-foreground';

    const timestamp = formatTimestamp(clip.created_at);

    return (
        <div
            ref={visibilityRef}
            className='bg-card border-y md:border border-border md:rounded-xl md:hover:shadow-lg transition-shadow lazy-render scroll-mt-28 snap-start'
            data-testid='clip-card'
        >
            <div className='flex flex-col md:flex-row gap-3 md:gap-6 py-4 md:p-6'>
                {/* Vote sidebar - horizontal on mobile, vertical on larger screens */}
                {isDesktop && <div className='flex flex-col items-center justify-start w-10 gap-2 order-1 shrink-0'>
                    <button
                        onClick={() => handleVote(1)}
                        disabled={!isAuthenticated || isVoting}
                        className={cn(
                            'w-11 h-11 xs:w-10 xs:h-10 rounded hover:bg-surface-hover flex items-center justify-center transition-colors touch-target',
                            clip.user_vote === 1 &&
                                'text-upvote',
                            !isAuthenticated || isVoting ?
                                'opacity-50 cursor-not-allowed hover:bg-transparent'
                            :   'cursor-pointer',
                        )}
                        aria-label={
                            isAuthenticated ? 'Upvote' : 'Log in to upvote'
                        }
                        aria-disabled={!isAuthenticated || isVoting}
                        title={isAuthenticated ? 'Upvote' : 'Log in to vote'}
                    >
                        <ArrowBigUp
                            size={22}
                            fill={clip.user_vote === 1 ? 'currentColor' : 'none'}
                            strokeWidth={1.75}
                        />
                    </button>

                    <span
                        className={cn(
                            'text-xs font-bold min-w-8 text-center',
                            voteColor,
                        )}
                    >
                        {formatCompactNumber(clip.vote_score)}
                    </span>

                    <button
                        onClick={() => handleVote(-1)}
                        disabled={!isAuthenticated || isVoting}
                        className={cn(
                            'w-11 h-11 xs:w-10 xs:h-10 rounded hover:bg-surface-hover flex items-center justify-center transition-colors touch-target',
                            clip.user_vote === -1 &&
                                'text-downvote',
                            !isAuthenticated || isVoting ?
                                'opacity-50 cursor-not-allowed hover:bg-transparent'
                            :   'cursor-pointer',
                        )}
                        aria-label={
                            isAuthenticated ? 'Downvote' : 'Log in to downvote'
                        }
                        aria-disabled={!isAuthenticated || isVoting}
                        title={isAuthenticated ? 'Downvote' : 'Log in to vote'}
                    >
                        <ArrowBigDown
                            size={22}
                            fill={clip.user_vote === -1 ? 'currentColor' : 'none'}
                            strokeWidth={1.75}
                        />
                    </button>
                </div>}

                {/* Main content */}
                <div className='flex-1 min-w-0 order-1 md:order-2'>
                    {/* Title */}
                    <Link
                        to={`/clip/${clip.id}`}
                        className='hover:text-primary-400 block mb-2 px-4 md:px-0 transition-colors touch-target cursor-pointer'
                    >
                        <h2 className='line-clamp-2 text-lg md:text-xl font-semibold leading-snug'>
                            {clip.title}
                        </h2>
                    </Link>
                    {/* Metadata */}
                    <div className='text-muted-foreground flex flex-wrap items-center gap-1.5 md:gap-2 mb-3 px-4 md:px-0 text-xs leading-tight'>
                        <span className='flex items-center gap-1 font-medium'>
                            <Link
                                to={`/broadcaster/${
                                    clip.broadcaster_id || clip.broadcaster_name
                                }`}
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
                                        to={`/twitch-category/${clip.twitch_category_id || clip.game_id}`}
                                        className='hover:text-foreground transition-colors cursor-pointer'
                                    >
                                        {clip.game_name}
                                    </Link>
                                </span>
                            </>
                        )}

                        {(clip.submitted_by ||
                            (clip.creator_id &&
                                clip.creator_id.trim() !== '' &&
                                clip.creator_name)) && (
                            <span className='hidden xs:inline'>•</span>
                        )}

                        {clip.submitted_by ?
                            <span className='inline-flex items-center gap-1'>
                                <Link
                                    to={`/user/${clip.submitted_by.username}`}
                                    className='hover:text-foreground transition-colors cursor-pointer'
                                >
                                    {clip.submitted_by.display_name}
                                </Link>
                                {clip.submitted_by.is_verified && (
                                    <VerifiedBadge size='sm' />
                                )}
                            </span>
                        : (
                            clip.creator_id &&
                            clip.creator_id.trim() !== '' &&
                            clip.creator_name
                        ) ?
                            <span className='inline-flex items-center gap-1'>
                                Clipped by{' '}
                                <Link
                                    to={`/user/${clip.creator_id}`}
                                    className='hover:text-foreground transition-colors cursor-pointer'
                                >
                                    {clip.creator_name}
                                </Link>
                            </span>
                        :   null}

                        <span className='hidden xs:inline'>•</span>

                        <span
                            className='truncate align-middle'
                            title={timestamp.title}
                        >
                            {timestamp.display}
                        </span>
                    </div>
                    {/* Thumbnail/Embed */}
                    <div className='relative mb-3'>
                        <TwitchEmbed
                            clipId={clip.twitch_clip_id}
                            thumbnailUrl={clip.thumbnail_url}
                            title={clip.title}
                            active={active}
                            autoplay={autoplay}
                            onActivate={() => onActivate?.(clip.id)}
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

                        {/* Featured badge */}
                        {clip.is_featured && (
                            <div className='top-2 right-2 absolute'>
                                <Badge variant='default'>Featured</Badge>
                            </div>
                        )}

                        {/* Watch progress indicator */}
                        {clip.watch_progress && (
                            <>
                                <div
                                    className='bottom-0 left-0 right-0 absolute h-1 bg-surface-raised'
                                    role='progressbar'
                                    aria-valuenow={Math.round(
                                        Math.min(
                                            100,
                                            Math.max(
                                                0,
                                                clip.watch_progress
                                                    .progress_percent,
                                            ),
                                        ),
                                    )}
                                    aria-valuemin={0}
                                    aria-valuemax={100}
                                    aria-label={`${Math.round(
                                        Math.min(
                                            100,
                                            Math.max(
                                                0,
                                                clip.watch_progress
                                                    .progress_percent,
                                            ),
                                        ),
                                    )}% watched`}
                                >
                                    <div
                                        className='h-full bg-primary-600'
                                        style={{
                                            width: `${Math.min(100, Math.max(0, clip.watch_progress.progress_percent))}%`,
                                        }}
                                    />
                                </div>
                                {clip.watch_progress.completed && (
                                    <div className='bottom-2 left-2 absolute px-2 py-1 text-xs font-medium text-white bg-success-600 bg-opacity-90 rounded flex items-center gap-1'>
                                        <Check className='w-3 h-3' />
                                        Watched
                                    </div>
                                )}
                            </>
                        )}
                    </div>

                    {/* Tags */}
                    <div className='mb-2 px-4 md:px-0'>
                        <TagList clipId={clip.id} maxVisible={2} />
                    </div>

                    {/* Action bar */}
                    <div className='flex items-center gap-1 px-2 md:px-0 text-xs'>
                        {!isDesktop && <div className='flex items-center min-h-11'>
                            <button
                                onClick={() => handleVote(1)}
                                disabled={!isAuthenticated || isVoting}
                                className={cn('grid size-11 place-items-center rounded-full', clip.user_vote === 1 ? 'text-upvote' : 'text-muted-foreground')}
                                aria-label={isAuthenticated ? 'Upvote' : 'Log in to upvote'}
                            >
                                <ArrowBigUp size={21} fill={clip.user_vote === 1 ? 'currentColor' : 'none'} />
                            </button>
                            <span className={cn('min-w-7 text-center text-xs font-bold', voteColor)}>{formatCompactNumber(clip.vote_score)}</span>
                            <button
                                onClick={() => handleVote(-1)}
                                disabled={!isAuthenticated || isVoting}
                                className={cn('grid size-11 place-items-center rounded-full', clip.user_vote === -1 ? 'text-downvote' : 'text-muted-foreground')}
                                aria-label={isAuthenticated ? 'Downvote' : 'Log in to downvote'}
                            >
                                <ArrowBigDown size={21} fill={clip.user_vote === -1 ? 'currentColor' : 'none'} />
                            </button>
                        </div>}
                        <Link
                            to={`/clip/${clip.id}#comments`}
                            className='text-muted-foreground hover:text-foreground flex min-w-11 items-center justify-center gap-1.5 transition-colors touch-target min-h-11 cursor-pointer'
                        >
                            <MessageSquare size={18} className='shrink-0' strokeWidth={1.75} />
                            <span className='hidden md:inline'>
                                {formatCompactNumber(clip.comment_count)} comments
                            </span>
                            <span className='md:hidden'>
                                {formatCompactNumber(clip.comment_count)}
                            </span>
                        </Link>

                        <button
                            onClick={handleFavorite}
                            disabled={!isAuthenticated}
                            className={cn(
                                'flex min-w-11 items-center justify-center gap-1.5 transition-colors touch-target min-h-11',
                                clip.is_favorited ?
                                    'text-primary-500'
                                :   'text-muted-foreground hover:text-foreground',
                                !isAuthenticated ?
                                    'opacity-50 cursor-not-allowed hover:bg-transparent'
                                :   'cursor-pointer',
                            )}
                            aria-label={
                                !isAuthenticated ? 'Log in to favorite'
                                : clip.is_favorited ?
                                    'Remove from favorites'
                                :   'Add to favorites'
                            }
                            aria-disabled={!isAuthenticated}
                            title={
                                !isAuthenticated ? 'Log in to favorite' : (
                                    undefined
                                )
                            }
                        >
                            <Heart
                                size={18}
                                className='shrink-0'
                                fill={clip.is_favorited ? 'currentColor' : 'none'}
                                strokeWidth={1.75}
                            />
                            <span>{formatCompactNumber(clip.favorite_count)}</span>
                        </button>

                        <ShareButton clipId={clip.id} clipTitle={clip.title} />

                        <span className='ml-auto text-muted-foreground hidden sm:flex items-center gap-1'>
                            <Eye size={18} strokeWidth={1.75} />
                            <span>{formatCompactNumber(clip.view_count)}</span>
                        </span>

                        <details className='relative md:hidden'>
                            <summary className='grid size-11 list-none place-items-center rounded-full text-muted-foreground hover:bg-surface-hover cursor-pointer' aria-label='More clip actions'>
                                <MoreHorizontal size={21} aria-hidden='true' />
                            </summary>
                            <div className='absolute bottom-12 right-0 z-30 flex min-w-44 flex-col gap-1 rounded-xl border border-border bg-background p-2 shadow-xl'>
                                <AddToPlaylistButton clipId={clip.id} />
                                <AddToQueueButton clipId={clip.id} />
                                <Link to={`/clip/${clip.id}`} className='flex min-h-11 items-center px-2 text-sm text-muted-foreground'>View clip details</Link>
                            </div>
                        </details>
                        <div className='hidden md:flex items-center gap-4'>
                            <AddToPlaylistButton clipId={clip.id} />
                            <AddToQueueButton clipId={clip.id} />
                        </div>
                    </div>
                </div>
            </div>
        </div>
    );
}
