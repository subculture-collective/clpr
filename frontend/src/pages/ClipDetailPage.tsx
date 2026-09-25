import { useParams, Link } from 'react-router-dom';
import {
    Container,
    Spinner,
    CommentSection,
    SEO,
    VideoPlayer,
    TheatreMode,
    ResourceUnavailable,
} from '../components';
import {
    useClipById,
    useUser,
    useClipVote,
    useClipFavorite,
    useIsAuthenticated,
    useToast,
    useWatchHistory,
} from '../hooks';
import { cn } from '@/lib/utils';
import { apiClient } from '@/lib/api';
import { ShareButton } from '@/components/clip/ShareButton';
import { TagList } from '@/components/tag/TagList';
import { useEffect, useRef } from 'react';
import { useQuery } from '@tanstack/react-query';
import { topicApi } from '@/lib/topic-api';
import { isNotFoundError } from '@/lib/error-utils';

export function ClipDetailPage() {
    const { id } = useParams<{ id: string }>();
    const { data: clip, isLoading, error, refetch } = useClipById(id || '');
    const user = useUser();
    const isAuthenticated = useIsAuthenticated();
    const voteMutation = useClipVote();
    const favoriteMutation = useClipFavorite();
    const toast = useToast();
    const isVoting = voteMutation.isPending;
    const isBanned = user?.is_banned;
    const banReason = user?.ban_reason;
    const { data: clipTopics } = useQuery({
        queryKey: ['clip-topics', id],
        queryFn: () => topicApi.getClipTopics(id || ''),
        enabled: Boolean(id),
    });

    // Watch history integration - full progress tracking for HLS clips only
    const {
        progress: resumePosition,
        hasProgress,
        isLoading: isLoadingProgress,
        recordProgress,
        recordProgressOnPause,
    } = useWatchHistory({
        clipId: id || '',
        duration: clip?.duration || 0,
        enabled: isAuthenticated && !!clip?.video_url,
    });

    // For iframe-only clips (no video_url), record a basic "viewed" entry
    // since Twitch embeds don't expose playback events per TOS
    const viewRecorded = useRef(false);
    useEffect(() => {
        if (!isAuthenticated || !clip || clip.video_url || viewRecorded.current)
            return;
        viewRecorded.current = true;
        apiClient.post('/watch-history', {
            clip_id: clip.id,
            progress_seconds: Math.floor(clip.duration || 0),
            duration_seconds: Math.floor(clip.duration || 30),
            session_id: `embed_${Date.now()}`,
        }).catch(() => {
            /* ignore */
        });
    }, [isAuthenticated, clip]);

    const clipUrl = clip ? `${window.location.origin}/clip/${clip.id}` : '';
    // Show ban message if user is banned (before clip loading checks)
    if (isBanned) {
        return (
            <>
                <SEO title='Banned' noindex />
                <Container className='py-8'>
                    <div className='border border-error-800 border-l-[3px] border-l-error-400 bg-surface p-6 my-8'>
                        <h2 className='text-2xl text-error-300 mb-2'>
                            You are banned
                        </h2>
                        <p className='text-text-secondary'>
                            You are banned and cannot interact with clips
                            {banReason ? `: ${banReason}` : ''}.
                        </p>
                    </div>
                </Container>
            </>
        );
    }


    const handleVote = (voteType: 1 | -1) => {
        if (!isAuthenticated) {
            toast.info('Please log in to vote on clips');
            return;
        }
        if (!clip || isVoting) return;
        if (clip.user_vote === voteType) return;
        voteMutation.mutate({ clip_id: clip.id, vote_type: voteType }, {
            onError: (error) => {
                const msg = (error as { response?: { data?: { error?: string } } })?.response?.data?.error;
                toast.error(msg || 'Failed to vote. Please try again.');
            },
        });
    };

    const handleFavorite = () => {
        if (!isAuthenticated) {
            toast.info('Please log in to favorite clips');
            return;
        }
        if (!clip) return;
        favoriteMutation.mutate({ clip_id: clip.id }, {
            onError: (error) => {
                const msg = (error as { response?: { data?: { error?: string } } })?.response?.data?.error;
                toast.error(msg || 'Failed to favorite. Please try again.');
            },
        });
    };

    if (isLoading) {
        return (
            <>
                <SEO title='Loading Clip...' noindex />
                <Container className='py-8'>
                    <div className='flex justify-center items-center min-h-[400px]'>
                        <Spinner size='lg' />
                    </div>
                </Container>
            </>
        );
    }

    if (error || !clip) {
        const notFound = !error || isNotFoundError(error);
        return (
            <>
                <SEO title={notFound ? 'Clip not found' : 'Clip unavailable'} noindex />
                <Container className='py-8'>
                    {notFound ?
                        <ResourceUnavailable
                            kind='not-found'
                            title="This clip isn't here"
                            description='It may have been removed, or the link is incomplete.'
                            links={[
                                { label: 'Back to the feed', href: '/' },
                                { label: 'Search clips', href: '/search' },
                            ]}
                        />
                    :   <ResourceUnavailable
                            kind='error'
                            title="We couldn't load this clip"
                            description='Check your connection and try again.'
                            onRetry={() => void refetch()}
                            links={[{ label: 'Back to the feed', href: '/' }]}
                        />
                    }
                </Container>
            </>
        );
    }

    // Format duration for display
    const formatDuration = (seconds: number | null | undefined) => {
        if (!seconds) return '';
        return `PT${Math.round(seconds)}S`;
    };

    // Generate rich description
    const description = `Watch "${clip.title}" by ${clip.creator_name} on ${
        clip.broadcaster_name
    }'s channel${
        clip.game_name ? ` in the ${clip.game_name} Twitch category` : ''
    }. ${clip.view_count.toLocaleString()} views, ${clip.vote_score} votes.`;

    // Schema.org VideoObject structured data
    const structuredData = {
        '@context': 'https://schema.org',
        '@type': 'VideoObject',
        name: clip.title,
        description: description,
        thumbnailUrl: clip.thumbnail_url || '',
        uploadDate: clip.created_at,
        duration: formatDuration(clip.duration),
        embedUrl: clip.embed_url,
        contentUrl: clip.twitch_clip_url,
        interactionStatistic: [
            {
                '@type': 'InteractionCounter',
                interactionType: 'https://schema.org/WatchAction',
                userInteractionCount: clip.view_count,
            },
            {
                '@type': 'InteractionCounter',
                interactionType: 'https://schema.org/LikeAction',
                userInteractionCount: clip.vote_score > 0 ? clip.vote_score : 0,
            },
            {
                '@type': 'InteractionCounter',
                interactionType: 'https://schema.org/CommentAction',
                userInteractionCount: clip.comment_count,
            },
        ],
        creator: {
            '@type': 'Person',
            name: clip.creator_name,
        },
    };

    return (
        <>
            <SEO
                title={clip.title}
                description={description}
                canonicalUrl={`/clip/${clip.id}`}
                ogType='video.other'
                ogImage={clip.thumbnail_url || undefined}
                imageAlt={`${clip.title} — ${clip.creator_name}`}
                structuredData={structuredData}
            />
            <Container className='py-4 md:py-8'>
                <div className='grid min-w-0 gap-8 xl:grid-cols-[minmax(0,1fr)_24rem]'>
                <div className='min-w-0'>
                {/* Video Player — full width */}
                <div className='mb-6'>
                    {clip.video_url ?
                        <TheatreMode
                            title={clip.title}
                            hlsUrl={clip.video_url}
                            resumePosition={resumePosition}
                            hasProgress={hasProgress}
                            isLoadingProgress={isLoadingProgress}
                            onProgressUpdate={recordProgress}
                            onPause={recordProgressOnPause}
                            onEnded={recordProgressOnPause}
                        />
                    :   <VideoPlayer
                            clipId={clip.id}
                            title={clip.title}
                            embedUrl={clip.embed_url}
                            twitchClipId={clip.twitch_clip_id}
                        />
                    }
                </div>

                {/* Header — compact, matching PlaylistDetail style */}
                <div className='mb-6'>
                    <h1 className='text-3xl lg:text-4xl text-foreground mb-2'>
                        {clip.title}
                    </h1>

                    {/* Metadata row */}
                    <div className='flex flex-wrap items-center gap-x-2 gap-y-1 font-mono text-[11px] uppercase tracking-[0.04em] text-muted-foreground mb-4'>
                        <Link
                            to={`/broadcaster/${clip.broadcaster_id || clip.broadcaster_name}`}
                            className='font-medium text-foreground/90 hover:text-foreground transition-colors'
                        >
                            {clip.broadcaster_name}
                        </Link>
                        {clip.game_name && (
                            <>
                                <span className='text-text-disabled'>·</span>
                                <Link
                                    to={`/twitch-category/${clip.twitch_category_id || clip.game_id}`}
                                    className='hover:text-foreground transition-colors'
                                >
                                    {clip.game_name}
                                </Link>
                            </>
                        )}
                        {clip.submitted_by && (
                            <>
                                <span className='text-text-disabled'>·</span>
                                <span>
                                    by{' '}
                                    <Link
                                        to={`/user/${clip.submitted_by.username}`}
                                        className='hover:text-foreground transition-colors'
                                    >
                                        {clip.submitted_by.display_name}
                                    </Link>
                                </span>
                            </>
                        )}
                        <span className='text-text-disabled'>·</span>
                        {/* view_count is clpr's last Twitch sync; the embed shows Twitch's live count. */}
                        <span title='Twitch view count when clpr last synced this clip. The player shows the live count.'>
                            {clip.view_count.toLocaleString()} views at last sync
                        </span>
                        <span className='text-text-disabled'>·</span>
                        <span>
                            {new Date(clip.created_at).toLocaleDateString('en-US', {
                                year: 'numeric',
                                month: 'short',
                                day: 'numeric',
                            })}
                        </span>
                    </div>

                    <div className='mb-4 grid gap-2 sm:grid-cols-[6rem_1fr] sm:items-start'>
                        <p className='kicker pt-1.5'>Tags</p>
                        <TagList clipId={clip.id} maxVisible={16} />
                    </div>

                    {clipTopics && clipTopics.topics.length > 0 && (
                        <div className='mb-4 flex flex-wrap items-center gap-1.5 sm:grid sm:grid-cols-[6rem_1fr]' aria-label='Clip topics'>
                            <p className='kicker'>Topics</p>
                            <div className='flex flex-wrap gap-1.5'>
                            {clipTopics.topics.map(topic => (
                                <Link
                                    key={topic.topic_id}
                                    to={`/topics/${topic.topic_slug}`}
                                    className='inline-flex min-h-8 items-center border border-line-strong px-2 font-heading text-[13px] font-bold uppercase tracking-[0.04em] text-foreground transition-colors hover:border-text-tertiary'
                                >
                                    {topic.topic_name}
                                </Link>
                            ))}
                            </div>
                        </div>
                    )}

                    {/* Actions row — matching PlaylistDetail stats row */}
                    <div className='flex flex-wrap items-center gap-x-2 gap-y-1 border-t border-border pt-3 font-mono text-xs text-muted-foreground'>
                        <button
                            aria-label={`Upvote: ${clip.vote_score} votes`}
                            aria-pressed={clip.user_vote === 1}
                            onClick={() => handleVote(1)}
                            disabled={!isAuthenticated || isVoting || isBanned}
                            className={cn(
                                'inline-flex min-h-11 min-w-11 items-center justify-center gap-2 px-3 py-2 transition-colors cursor-pointer',
                                clip.user_vote === 1
                                    ? 'text-upvote bg-upvote/10'
                                    : 'hover:bg-accent hover:text-foreground',
                                (!isAuthenticated || isVoting || isBanned) &&
                                    'opacity-50 cursor-not-allowed',
                            )}
                            title={isAuthenticated ? 'Upvote' : 'Log in to vote'}
                        >
                            <svg className='h-5 w-5' fill={clip.user_vote === 1 ? 'currentColor' : 'none'} stroke='currentColor' strokeWidth={2} viewBox='0 0 24 24'>
                                <path d='M12 4l8 8h-6v8h-4v-8H4z' />
                            </svg>
                            <span className='font-medium text-foreground/90'>
                                {clip.vote_score}
                            </span>
                        </button>

                        <button
                            onClick={() => {
                                const discussion = document.getElementById('comments');
                                discussion?.focus({ preventScroll: true });
                                discussion?.scrollIntoView({ behavior: window.matchMedia('(prefers-reduced-motion: reduce)').matches ? 'instant' : 'smooth', block: 'start' });
                            }}
                            aria-label={`Read discussion: ${clip.comment_count} comments`}
                            className='inline-flex min-h-11 min-w-11 items-center justify-center gap-2 px-3 py-2 transition-colors hover:bg-accent hover:text-foreground cursor-pointer'
                        >
                            <svg className='h-5 w-5' fill='none' stroke='currentColor' strokeWidth={2} viewBox='0 0 24 24'>
                                <path strokeLinecap='round' strokeLinejoin='round' d='M7 8h10M7 12h4m1 8l-4-4H5a2 2 0 01-2-2V6a2 2 0 012-2h14a2 2 0 012 2v8a2 2 0 01-2 2h-3l-4 4z' />
                            </svg>
                            <span className='font-medium text-foreground/90'>
                                {clip.comment_count}
                            </span>
                        </button>

                        <button
                            aria-label={clip.is_favorited ? 'Remove from saved clips' : 'Save clip'}
                            aria-pressed={!!clip.is_favorited}
                            onClick={() => handleFavorite()}
                            disabled={!isAuthenticated || isBanned}
                            className={cn(
                                'inline-flex min-h-11 min-w-11 items-center justify-center gap-2 px-3 py-2 transition-colors cursor-pointer',
                                clip.is_favorited
                                    ? 'text-red-500'
                                    : 'hover:bg-accent hover:text-foreground',
                                (!isAuthenticated || isBanned) &&
                                    'opacity-50 cursor-not-allowed',
                            )}
                            title={isAuthenticated ? (clip.is_favorited ? 'Unfavorite' : 'Favorite') : 'Log in to favorite'}
                        >
                            <svg className='h-5 w-5' fill={clip.is_favorited ? 'currentColor' : 'none'} stroke='currentColor' strokeWidth={2} viewBox='0 0 24 24'>
                                <path strokeLinecap='round' strokeLinejoin='round' d='M4.318 6.318a4.5 4.5 0 000 6.364L12 20.364l7.682-7.682a4.5 4.5 0 00-6.364-6.364L12 7.636l-1.318-1.318a4.5 4.5 0 00-6.364 0z' />
                            </svg>
                            <span className='font-medium text-foreground/90'>
                                {clip.favorite_count}
                            </span>
                        </button>

                        <ShareButton
                            shareUrl={clipUrl}
                            shareTitle={clip.title}
                            showLabel={false}
                            buttonClassName='inline-flex min-h-11 min-w-11 items-center justify-center rounded px-3 py-2 text-muted-foreground transition-colors hover:bg-accent hover:text-foreground cursor-pointer'
                            iconClassName='h-5 w-5'
                        />
                    </div>
                </div>

                </div>
                <section id='comments' aria-label='Discussion' tabIndex={-1} className='min-w-0 scroll-mt-24 border border-border bg-surface p-4 tally-bar xl:sticky xl:top-24 xl:max-h-[calc(100dvh-8rem)] xl:self-start xl:overflow-hidden'>
                    <CommentSection
                        clipId={clip.id}
                        variant='compact'
                        className='xl:h-[calc(100dvh-10rem)]'
                        currentUserId={user?.id}
                        isAdmin={user?.role === 'admin'}
                        isBanned={!!isBanned}
                        banReason={banReason}
                    />
                </section>
                </div>
            </Container>
        </>
    );
}
