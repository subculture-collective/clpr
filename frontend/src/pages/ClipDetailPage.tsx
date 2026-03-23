import { useParams, Link } from 'react-router-dom';
import {
    Container,
    Spinner,
    CommentSection,
    SEO,
    VideoPlayer,
    TheatreMode,
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
import { useState, useEffect, useRef } from 'react';

export function ClipDetailPage() {
    const { id } = useParams<{ id: string }>();
    const { data: clip, isLoading, error } = useClipById(id || '');
    const user = useUser();
    const isAuthenticated = useIsAuthenticated();
    const voteMutation = useClipVote();
    const favoriteMutation = useClipFavorite();
    const toast = useToast();
    const isVoting = voteMutation.isPending;
    const isBanned = user?.is_banned;
    const banReason = user?.ban_reason;

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
                    <div className='rounded-lg border border-red-200 bg-red-50 dark:bg-red-900/20 dark:border-red-800 p-6 my-8'>
                        <h2 className='text-lg font-bold text-red-800 dark:text-red-400 mb-2'>
                            You are banned
                        </h2>
                        <p className='text-red-700 dark:text-red-300'>
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

    if (error) {
        return (
            <>
                <SEO title='Error Loading Clip' noindex />
                <Container className='py-8'>
                    <div className='text-center py-12'>
                        <h2 className='text-2xl font-bold text-error-600 mb-4'>
                            Error Loading Clip
                        </h2>
                        <p className='text-muted-foreground'>{error.message}</p>
                    </div>
                </Container>
            </>
        );
    }

    if (!clip) {
        return (
            <>
                <SEO title='Clip Not Found' noindex />
                <Container className='py-8'>
                    <div className='text-center py-12'>
                        <h2 className='text-2xl font-bold mb-4'>
                            Clip Not Found
                        </h2>
                        <p className='text-muted-foreground'>
                            The clip you're looking for doesn't exist.
                        </p>
                    </div>
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
        clip.game_name ? ` playing ${clip.game_name}` : ''
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
                twitterCard='player'
                structuredData={structuredData}
            />
            <Container className='py-4 xs:py-6 md:py-8'>
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
                    <h1 className='text-xl font-semibold text-foreground mb-1 leading-tight'>
                        {clip.title}
                    </h1>

                    {/* Metadata row */}
                    <div className='flex flex-wrap items-center gap-x-2 gap-y-1 text-xs text-muted-foreground mb-3'>
                        <Link
                            to={`/broadcaster/${clip.broadcaster_id || clip.broadcaster_name}`}
                            className='font-medium text-foreground/90 hover:text-foreground transition-colors'
                        >
                            {clip.broadcaster_name}
                        </Link>
                        {clip.game_name && (
                            <>
                                <span>•</span>
                                <Link
                                    to={`/game/${clip.game_id}`}
                                    className='hover:text-foreground transition-colors'
                                >
                                    {clip.game_name}
                                </Link>
                            </>
                        )}
                        {clip.submitted_by && (
                            <>
                                <span>•</span>
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
                        <span>•</span>
                        <span>{clip.view_count.toLocaleString()} views</span>
                        <span>•</span>
                        <span>
                            {new Date(clip.created_at).toLocaleDateString('en-US', {
                                year: 'numeric',
                                month: 'short',
                                day: 'numeric',
                            })}
                        </span>
                    </div>

                    {/* Actions row — matching PlaylistDetail stats row */}
                    <div className='flex flex-wrap items-center gap-x-2 gap-y-1 text-xs text-muted-foreground'>
                        <button
                            onClick={() => handleVote(1)}
                            disabled={!isAuthenticated || isVoting || isBanned}
                            className={cn(
                                'inline-flex items-center gap-1 rounded px-1.5 py-0.5 transition-colors cursor-pointer',
                                clip.user_vote === 1
                                    ? 'text-upvote bg-upvote/10'
                                    : 'hover:bg-accent hover:text-foreground',
                                (!isAuthenticated || isVoting || isBanned) &&
                                    'opacity-50 cursor-not-allowed',
                            )}
                            title={isAuthenticated ? 'Upvote' : 'Log in to vote'}
                        >
                            <svg className='h-3.5 w-3.5' fill={clip.user_vote === 1 ? 'currentColor' : 'none'} stroke='currentColor' strokeWidth={2} viewBox='0 0 24 24'>
                                <path d='M12 4l8 8h-6v8h-4v-8H4z' />
                            </svg>
                            <span className='font-medium text-foreground/90'>
                                {clip.vote_score}
                            </span>
                        </button>

                        <button
                            onClick={() => {
                                if (!isBanned) {
                                    document.getElementById('comments')?.scrollIntoView({ behavior: 'smooth' });
                                }
                            }}
                            disabled={isBanned}
                            className='inline-flex items-center gap-1 rounded px-1.5 py-0.5 transition-colors hover:bg-accent hover:text-foreground cursor-pointer'
                        >
                            <svg className='h-3.5 w-3.5' fill='none' stroke='currentColor' strokeWidth={2} viewBox='0 0 24 24'>
                                <path strokeLinecap='round' strokeLinejoin='round' d='M7 8h10M7 12h4m1 8l-4-4H5a2 2 0 01-2-2V6a2 2 0 012-2h14a2 2 0 012 2v8a2 2 0 01-2 2h-3l-4 4z' />
                            </svg>
                            <span className='font-medium text-foreground/90'>
                                {clip.comment_count}
                            </span>
                        </button>

                        <button
                            onClick={() => handleFavorite()}
                            disabled={!isAuthenticated || isBanned}
                            className={cn(
                                'inline-flex items-center gap-1 rounded px-1.5 py-0.5 transition-colors cursor-pointer',
                                clip.is_favorited
                                    ? 'text-red-500'
                                    : 'hover:bg-accent hover:text-foreground',
                                (!isAuthenticated || isBanned) &&
                                    'opacity-50 cursor-not-allowed',
                            )}
                            title={isAuthenticated ? (clip.is_favorited ? 'Unfavorite' : 'Favorite') : 'Log in to favorite'}
                        >
                            <svg className='h-3.5 w-3.5' fill={clip.is_favorited ? 'currentColor' : 'none'} stroke='currentColor' strokeWidth={2} viewBox='0 0 24 24'>
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
                            buttonClassName='inline-flex min-h-0 items-center rounded px-1.5 py-0.5 text-muted-foreground transition-colors hover:bg-accent hover:text-foreground cursor-pointer'
                            iconClassName='h-3.5 w-3.5'
                        />
                    </div>
                </div>

                {/* Comments */}
                <div id='comments'>
                    <CommentSection
                        clipId={clip.id}
                        currentUserId={user?.id}
                        isAdmin={user?.role === 'admin'}
                        isBanned={!!isBanned}
                        banReason={banReason}
                    />
                </div>
            </Container>
        </>
    );
}
