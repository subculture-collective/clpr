import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { useState } from 'react';
import { useParams } from 'react-router-dom';
import { Container, SEO } from '../components';
import { ClipGridCard } from '../components/clip';
import { Button } from '../components/ui';
import { Spinner } from '../components';
import { LiveBadge } from '../components/broadcaster';
import {
    fetchBroadcasterProfile,
    fetchBroadcasterClips,
    followBroadcaster,
    unfollowBroadcaster,
    fetchBroadcasterLiveStatus,
} from '../lib/broadcaster-api';
import { apiClient } from '../lib/api';
import { useAuth } from '../context/AuthContext';
import { useToast } from '../context/ToastContext';

export function BroadcasterPage() {
    const { broadcasterId } = useParams<{ broadcasterId: string }>();
    const { isAuthenticated } = useAuth();
    const { showToast } = useToast();
    const queryClient = useQueryClient();
    const [page, setPage] = useState(1);
    const [sortBy, setSortBy] = useState<'recent' | 'popular' | 'trending'>(
        'recent',
    );

    // Fetch broadcaster profile
    const {
        data: profile,
        isLoading: isLoadingProfile,
        error: profileError,
    } = useQuery({
        queryKey: ['broadcaster', broadcasterId],
        queryFn: () => fetchBroadcasterProfile(broadcasterId!),
        enabled: !!broadcasterId,
    });

    // Fetch live status
    const { data: liveStatus } = useQuery({
        queryKey: ['broadcasterLiveStatus', broadcasterId],
        queryFn: () => fetchBroadcasterLiveStatus(broadcasterId!),
        enabled: !!broadcasterId,
        refetchInterval: 30000, // Refetch every 30 seconds
    });

    // Fetch broadcaster clips
    const { data: clipsData, isLoading: isLoadingClips } = useQuery({
        queryKey: ['broadcasterClips', broadcasterId, page, sortBy],
        queryFn: () =>
            fetchBroadcasterClips(broadcasterId!, {
                page,
                limit: 20,
                sort: sortBy,
            }),
        enabled: !!broadcasterId,
    });

    // Fetch broadcaster ranking data
    const { data: rankingData } = useQuery({
        queryKey: ['broadcaster-ranking', broadcasterId],
        queryFn: async () => {
            const res = await apiClient.get<{
                success: boolean;
                data: Array<{
                    broadcaster_id: string;
                    engagement_score: number;
                }>;
            }>('/broadcasters/rankings?limit=100');
            const rankings = res.data?.data || [];
            const idx = rankings.findIndex(
                r => r.broadcaster_id === broadcasterId,
            );
            if (idx >= 0) {
                return {
                    rank: idx + 1,
                    engagement_score: rankings[idx].engagement_score,
                };
            }
            return null;
        },
        enabled: !!broadcasterId,
    });

    // Follow mutation
    const followMutation = useMutation({
        mutationFn: () => followBroadcaster(broadcasterId!),
        onSuccess: () => {
            queryClient.invalidateQueries({
                queryKey: ['broadcaster', broadcasterId],
            });
            showToast('Successfully followed broadcaster', 'success');
        },
        onError: () => {
            showToast('Failed to follow broadcaster', 'error');
        },
    });

    // Unfollow mutation
    const unfollowMutation = useMutation({
        mutationFn: () => unfollowBroadcaster(broadcasterId!),
        onSuccess: () => {
            queryClient.invalidateQueries({
                queryKey: ['broadcaster', broadcasterId],
            });
            showToast('Successfully unfollowed broadcaster', 'success');
        },
        onError: () => {
            showToast('Failed to unfollow broadcaster', 'error');
        },
    });

    const handleFollowToggle = () => {
        if (!isAuthenticated) {
            showToast('Please log in to follow broadcasters', 'error');
            return;
        }

        if (profile?.is_following) {
            unfollowMutation.mutate();
        } else {
            followMutation.mutate();
        }
    };

    if (isLoadingProfile) {
        return (
            <Container className='py-8'>
                <div className='flex justify-center items-center min-h-100'>
                    <Spinner size='xl' />
                </div>
            </Container>
        );
    }

    if (profileError || !profile) {
        return (
            <Container className='py-8'>
                <div className='text-center text-muted-foreground py-12'>
                    <h2 className='text-2xl font-bold mb-2'>
                        Broadcaster Not Found
                    </h2>
                    <p>
                        The broadcaster you're looking for doesn't exist or has
                        no clips yet.
                    </p>
                </div>
            </Container>
        );
    }

    return (
        <>
            <SEO
                title={`${profile.display_name} - Broadcaster Profile`}
                description={`View all clips featuring ${profile.display_name} on Clipper`}
            />
            <Container className='py-8'>
                {/* Broadcaster Profile Header */}
                <div className='mb-8'>
                    <div className='flex flex-col md:flex-row items-start md:items-center gap-6 mb-6'>
                        {/* Avatar */}
                        {profile.avatar_url && (
                            <img
                                src={profile.avatar_url}
                                alt={`${profile.display_name} profile picture`}
                                className='w-24 h-24 md:w-32 md:h-32 rounded-full border-4 border-primary'
                            />
                        )}

                        {/* Info */}
                        <div className='flex-1'>
                            <div className='flex items-center gap-3 mb-2'>
                                <h1 className='text-4xl font-bold'>
                                    {profile.display_name}
                                </h1>
                                <LiveBadge
                                    liveStatus={liveStatus}
                                    showViewers={true}
                                    showDuration={true}
                                    size='lg'
                                    clickable={true}
                                    onClick={() =>
                                        window.open(
                                            `https://twitch.tv/${broadcasterId}`,
                                            '_blank',
                                            'noopener,noreferrer',
                                        )
                                    }
                                />
                            </div>
                            {profile.bio && (
                                <p className='text-muted-foreground mb-4'>
                                    {profile.bio}
                                </p>
                            )}

                            {/* Stats */}
                            <div className='flex flex-wrap gap-6 text-sm'>
                                <div>
                                    <span className='font-semibold'>
                                        {profile.total_clips}
                                    </span>
                                    <span className='text-muted-foreground ml-1'>
                                        Clips
                                    </span>
                                </div>
                                <div>
                                    <span className='font-semibold'>
                                        {profile.follower_count}
                                    </span>
                                    <span className='text-muted-foreground ml-1'>
                                        Followers
                                    </span>
                                </div>
                                <div>
                                    <span className='font-semibold'>
                                        {profile.total_views.toLocaleString()}
                                    </span>
                                    <span className='text-muted-foreground ml-1'>
                                        Total Views
                                    </span>
                                </div>
                                <div>
                                    <span className='font-semibold'>
                                        {profile.avg_vote_score.toFixed(1)}
                                    </span>
                                    <span className='text-muted-foreground ml-1'>
                                        Avg Score
                                    </span>
                                </div>
                                {rankingData && (
                                    <>
                                        <div>
                                            <span className='font-semibold'>
                                                {rankingData.engagement_score.toLocaleString()}
                                            </span>
                                            <span className='text-muted-foreground ml-1'>
                                                Engagement Score
                                            </span>
                                        </div>
                                        <div>
                                            <span className='font-semibold'>
                                                #{rankingData.rank}
                                            </span>
                                            <span className='text-muted-foreground ml-1'>
                                                Rank
                                            </span>
                                        </div>
                                    </>
                                )}
                            </div>
                        </div>

                        {/* Actions */}
                        <div className='flex flex-col gap-2'>
                            <Button
                                onClick={handleFollowToggle}
                                disabled={
                                    followMutation.isPending ||
                                    unfollowMutation.isPending
                                }
                                variant={
                                    profile.is_following ? 'outline' : 'primary'
                                }
                            >
                                {(
                                    followMutation.isPending ||
                                    unfollowMutation.isPending
                                ) ?
                                    <Spinner size='sm' />
                                : profile.is_following ?
                                    'Unfollow'
                                :   'Follow'}
                            </Button>
                            <Button
                                variant='outline'
                                onClick={() =>
                                    window.open(
                                        profile.twitch_url,
                                        '_blank',
                                        'noopener,noreferrer',
                                    )
                                }
                            >
                                View on Twitch
                            </Button>
                        </div>
                    </div>
                </div>

                {/* Sort Controls */}
                <div className='mb-6 flex items-center gap-4'>
                    <h2 className='text-2xl font-bold'>Clips</h2>
                    <div className='flex gap-2'>
                        <Button
                            size='sm'
                            variant={
                                sortBy === 'recent' ? 'primary' : 'outline'
                            }
                            onClick={() => {
                                setSortBy('recent');
                                setPage(1);
                            }}
                            aria-label='Sort clips by most recent'
                        >
                            Recent
                        </Button>
                        <Button
                            size='sm'
                            variant={
                                sortBy === 'popular' ? 'primary' : 'outline'
                            }
                            onClick={() => {
                                setSortBy('popular');
                                setPage(1);
                            }}
                            aria-label='Sort clips by most popular'
                        >
                            Popular
                        </Button>
                        <Button
                            size='sm'
                            variant={
                                sortBy === 'trending' ? 'primary' : 'outline'
                            }
                            onClick={() => {
                                setSortBy('trending');
                                setPage(1);
                            }}
                            aria-label='Sort clips by trending'
                        >
                            Trending
                        </Button>
                    </div>
                </div>

                {/* Clips Grid */}
                {isLoadingClips ?
                    <div className='flex justify-center items-center min-h-50'>
                        <Spinner size='lg' />
                    </div>
                : clipsData && (clipsData.data ?? []).length > 0 ?
                    <>
                        <div className='grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-6 mb-8'>
                            {(clipsData.data ?? []).map(clip => (
                                <ClipGridCard key={clip.id} clip={clip} />
                            ))}
                        </div>

                        {/* Pagination */}
                        {clipsData.meta.total_pages > 1 && (
                            <div className='flex justify-center gap-2'>
                                <Button
                                    variant='outline'
                                    onClick={() =>
                                        setPage(p => Math.max(1, p - 1))
                                    }
                                    disabled={page === 1}
                                >
                                    Previous
                                </Button>
                                <span className='flex items-center px-4'>
                                    Page {page} of {clipsData.meta.total_pages}
                                </span>
                                <Button
                                    variant='outline'
                                    onClick={() => setPage(p => p + 1)}
                                    disabled={
                                        page >= clipsData.meta.total_pages
                                    }
                                >
                                    Next
                                </Button>
                            </div>
                        )}
                    </>
                :   <div className='text-center text-muted-foreground py-12'>
                        <p>No clips found for this broadcaster.</p>
                    </div>
                }
            </Container>
        </>
    );
}
