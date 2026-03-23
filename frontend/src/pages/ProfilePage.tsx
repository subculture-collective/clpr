import { formatTimestamp } from '../lib/utils';
import { useCallback, useState, useEffect } from 'react';
import { Link } from 'react-router-dom';
import { apiClient } from '../lib/api';

import {
    Card,
    CardBody,
    Container,
    Stack,
    Button,
    Skeleton,
    ProfileSkeleton,
    EmptyStateWithAction,
    UserRoleBadge,
} from '../components';
import { VerifiedBadge } from '../components/user';
import { ClipCard } from '../components/clip/ClipCard';
import { ClipCardSkeleton } from '../components/clip/ClipCardSkeleton';
import { CommentSkeleton } from '../components/ui';
import {
    BadgeGrid,
    KarmaBreakdownChart,
    ReputationDisplay,
} from '../components/reputation';
import { useAuth } from '../context/AuthContext';
import { useToast } from '../context/ToastContext';
import type { KarmaBreakdown, UserReputation } from '../types/reputation';
import type { Clip } from '../types/clip';
import type { Comment } from '../types/comment';
import type { UserRole } from '../lib/roles';
import {
    fetchUserComments,
    fetchUserUpvotedClips,
    fetchUserDownvotedClips,
    reauthorizeTwitch,
} from '../lib/user-api';

type TabType =
    | 'overview'
    | 'badges'
    | 'karma'
    | 'comments'
    | 'upvoted'
    | 'downvoted';

export function ProfilePage() {
    const { user, isAdmin } = useAuth();
    const toast = useToast();
    const [reputation, setReputation] = useState<UserReputation | null>(null);
    const [karmaBreakdown, setKarmaBreakdown] = useState<KarmaBreakdown | null>(
        null,
    );
    const [loadingReputation, setLoadingReputation] = useState(true);
    const [reputationError, setReputationError] = useState(false);
    const [activeTab, setActiveTab] = useState<TabType>('overview');

    // Tab-specific data states
    const [comments, setComments] = useState<Comment[]>([]);
    const [upvotedClips, setUpvotedClips] = useState<Clip[]>([]);
    const [downvotedClips, setDownvotedClips] = useState<Clip[]>([]);
    const [loadingTabData, setLoadingTabData] = useState(false);
    const [currentPage, setCurrentPage] = useState(1);
    const [hasMore, setHasMore] = useState(false);
    const [reauthorizing, setReauthorizing] = useState(false);

    const fetchReputation = useCallback(async () => {
        if (!user) return;
        try {
            setLoadingReputation(true);
            setReputationError(false);
            const { data } = await apiClient.get(`/users/${user.id}/reputation`);
            setReputation(data);
        } catch (error) {
            console.error('Failed to fetch reputation:', error);
            setReputationError(true);
        } finally {
            setLoadingReputation(false);
        }
    }, [user]);

    const fetchKarmaBreakdown = useCallback(async () => {
        if (!user) return;
        try {
            const { data } = await apiClient.get(`/users/${user.id}/karma?limit=10`);
            setKarmaBreakdown(data.breakdown);
        } catch (error) {
            console.error('Failed to fetch karma breakdown:', error);
        }
    }, [user]);

    const fetchTabData = useCallback(
        async (tab: TabType, page: number = 1, append: boolean = false) => {
            if (!user) return;

            setLoadingTabData(true);
            try {
                switch (tab) {
                    case 'comments': {
                        const data = await fetchUserComments(user.id, page);
                        setComments(prev =>
                            append ?
                                [...prev, ...data.comments]
                            :   data.comments,
                        );
                        setHasMore(data.has_more);
                        break;
                    }
                    case 'upvoted': {
                        const data = await fetchUserUpvotedClips(user.id, page);
                        setUpvotedClips(prev =>
                            append ? [...prev, ...data.clips] : data.clips,
                        );
                        setHasMore(data.has_more);
                        break;
                    }
                    case 'downvoted': {
                        const data = await fetchUserDownvotedClips(
                            user.id,
                            page,
                        );
                        setDownvotedClips(prev =>
                            append ? [...prev, ...data.clips] : data.clips,
                        );
                        setHasMore(data.has_more);
                        break;
                    }
                }
                setCurrentPage(page);
            } catch (error) {
                console.error(`Failed to fetch ${tab} data:`, error);
                toast.error(`Failed to load ${tab}`);
            } finally {
                setLoadingTabData(false);
            }
        },
        [user, toast],
    );

    const handleTabChange = (tab: TabType) => {
        setActiveTab(tab);
        if (tab === 'comments' || tab === 'upvoted' || tab === 'downvoted') {
            // Reset only the relevant tab data when switching
            if (tab === 'comments') {
                setComments([]);
            } else if (tab === 'upvoted') {
                setUpvotedClips([]);
            } else if (tab === 'downvoted') {
                setDownvotedClips([]);
            }
            setCurrentPage(1);
            setHasMore(false);
            fetchTabData(tab, 1, false);
        }
    };

    const handleReauthorize = async () => {
        try {
            setReauthorizing(true);
            const data = await reauthorizeTwitch();
            // Redirect to the auth URL
            window.location.href = data.auth_url;
        } catch (error) {
            console.error('Failed to reauthorize:', error);
            toast.error('Failed to initiate reauthorization');
            setReauthorizing(false);
        }
    };

    useEffect(() => {
        if (user) {
            fetchReputation();
            fetchKarmaBreakdown();
        }
    }, [user, fetchReputation, fetchKarmaBreakdown]);

    if (!user) {
        return null;
    }

    return (
        <Container className='py-4 xs:py-6 md:py-8'>
            <div className='max-w-4xl mx-auto'>
                {/* Banned User Message */}
                {user?.is_banned && (
                    <Card className='mb-4 xs:mb-6 border-red-500 bg-red-50 dark:bg-red-950'>
                        <CardBody>
                            <div className='flex items-start gap-3 xs:gap-4'>
                                <div className='flex-shrink-0'>
                                    <svg
                                        className='w-5 h-5 xs:w-6 xs:h-6 text-red-600'
                                        fill='none'
                                        stroke='currentColor'
                                        viewBox='0 0 24 24'
                                    >
                                        <path
                                            strokeLinecap='round'
                                            strokeLinejoin='round'
                                            strokeWidth={2}
                                            d='M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z'
                                        />
                                    </svg>
                                </div>
                                <div>
                                    <h3 className='text-base xs:text-lg font-semibold text-red-900 dark:text-red-100'>
                                        Account Restricted
                                    </h3>
                                    <p className='mt-1 text-sm xs:text-base text-red-800 dark:text-red-200'>
                                        Your account has been restricted from
                                        interacting with certain content. Please
                                        contact support if you believe this is
                                        an error.
                                    </p>
                                </div>
                            </div>
                        </CardBody>
                    </Card>
                )}

                {/* Profile Header */}
                <Card className='mb-4 xs:mb-6'>
                    <CardBody>
                        <div className='flex flex-col xs:flex-row items-start gap-4 xs:gap-6'>
                            {/* Avatar */}
                            <div className='shrink-0 mx-auto xs:mx-0'>
                                {user.avatar_url ?
                                    <img
                                        src={user.avatar_url}
                                        alt={user.username}
                                        className='border-border w-20 h-20 xs:w-24 xs:h-24 border-2 rounded-full'
                                    />
                                :   <div className='bg-primary-100 dark:bg-primary-900 text-primary-600 flex items-center justify-center w-20 h-20 xs:w-24 xs:h-24 text-2xl xs:text-3xl font-bold rounded-full'>
                                        {user.username.charAt(0).toUpperCase()}
                                    </div>
                                }
                            </div>

                            {/* User Info */}
                            <div className='flex-1 w-full text-center xs:text-left'>
                                <div className='flex flex-col xs:flex-row items-center xs:items-start justify-between gap-3 xs:gap-0'>
                                    <div>
                                        <div className='flex items-center justify-center xs:justify-start gap-2 mb-1'>
                                            <h1 className='text-2xl xs:text-3xl font-bold'>
                                                {user.display_name}
                                            </h1>
                                            {user.is_verified && (
                                                <VerifiedBadge size='lg' />
                                            )}
                                        </div>
                                        <p className='text-sm xs:text-base text-muted-foreground mb-2'>
                                            @{user.username}
                                        </p>
                                        {!user.is_verified && (
                                            <Link
                                                to='/verification/apply'
                                                className='inline-block text-xs xs:text-sm text-primary-600 hover:underline'
                                            >
                                                Apply for verification →
                                            </Link>
                                        )}
                                    </div>
                                    <Button
                                        onClick={handleReauthorize}
                                        disabled={reauthorizing}
                                        variant='outline'
                                        size='sm'
                                        className='w-full xs:w-auto'
                                    >
                                        {reauthorizing ?
                                            'Redirecting...'
                                        :   'Reauthorize with Twitch'}
                                    </Button>
                                </div>

                                {user.bio && (
                                    <p className='text-sm xs:text-base text-foreground mb-4'>
                                        {user.bio}
                                    </p>
                                )}

                                <div className='flex flex-wrap justify-center xs:justify-start gap-3 xs:gap-4 text-xs xs:text-sm'>
                                    <div className='flex items-center gap-2'>
                                        <span className='text-muted-foreground'>
                                            Karma:
                                        </span>
                                        <span className='text-primary-600 font-semibold'>
                                            {user.karma_points}
                                        </span>
                                    </div>
                                    <div className='flex items-center gap-2'>
                                        <span className='text-muted-foreground'>
                                            Role:
                                        </span>
                                        <UserRoleBadge
                                            role={user.role as UserRole}
                                            size='sm'
                                        />
                                    </div>
                                    {user.created_at && (
                                        <div className='flex items-center gap-2'>
                                            <span className='text-muted-foreground'>
                                                Joined:
                                            </span>
                                            <span
                                                className='font-semibold'
                                                title={
                                                    formatTimestamp(
                                                        user.created_at,
                                                    ).title
                                                }
                                            >
                                                {
                                                    formatTimestamp(
                                                        user.created_at,
                                                    ).display
                                                }
                                            </span>
                                        </div>
                                    )}
                                </div>

                                {/* RBAC Navigation */}
                                {isAdmin && (
                                    <div className='mt-4 pt-4 border-t border-border'>
                                        <p className='text-sm text-muted-foreground mb-2'>
                                            Admin Tools:
                                        </p>
                                        <div className='flex gap-3'>
                                            <Link
                                                to='/admin/dashboard'
                                                className='text-sm text-primary-600 hover:text-primary-700 font-medium'
                                            >
                                                Dashboard
                                            </Link>
                                            <Link
                                                to='/admin/reports'
                                                className='text-sm text-primary-600 hover:text-primary-700 font-medium'
                                            >
                                                Reports
                                            </Link>
                                            <Link
                                                to='/admin/submissions'
                                                className='text-sm text-primary-600 hover:text-primary-700 font-medium'
                                            >
                                                Submissions
                                            </Link>
                                        </div>
                                    </div>
                                )}
                            </div>
                        </div>
                    </CardBody>
                </Card>

                {/* Tabs Section */}
                <Card>
                    <CardBody>
                        <Stack direction='vertical' gap={4}>
                            <div className='border-border border-b'>
                                <nav
                                    className='flex gap-4 overflow-x-auto'
                                    role='tablist'
                                >
                                    <button
                                        onClick={() =>
                                            handleTabChange('overview')
                                        }
                                        className={`px-4 py-2 border-b-2 font-semibold whitespace-nowrap ${
                                            activeTab === 'overview' ?
                                                'border-primary-500 text-primary-600'
                                            :   'border-transparent text-muted-foreground hover:text-foreground'
                                        }`}
                                        role='tab'
                                        aria-selected={activeTab === 'overview'}
                                    >
                                        Overview
                                    </button>
                                    <button
                                        onClick={() =>
                                            handleTabChange('badges')
                                        }
                                        className={`px-4 py-2 border-b-2 font-semibold whitespace-nowrap ${
                                            activeTab === 'badges' ?
                                                'border-primary-500 text-primary-600'
                                            :   'border-transparent text-muted-foreground hover:text-foreground'
                                        }`}
                                        role='tab'
                                        aria-selected={activeTab === 'badges'}
                                    >
                                        Badges
                                    </button>
                                    <button
                                        onClick={() => handleTabChange('karma')}
                                        className={`px-4 py-2 border-b-2 font-semibold whitespace-nowrap ${
                                            activeTab === 'karma' ?
                                                'border-primary-500 text-primary-600'
                                            :   'border-transparent text-muted-foreground hover:text-foreground'
                                        }`}
                                        role='tab'
                                        aria-selected={activeTab === 'karma'}
                                    >
                                        Karma
                                    </button>
                                    <button
                                        onClick={() =>
                                            handleTabChange('comments')
                                        }
                                        className={`px-4 py-2 border-b-2 font-semibold whitespace-nowrap ${
                                            activeTab === 'comments' ?
                                                'border-primary-500 text-primary-600'
                                            :   'border-transparent text-muted-foreground hover:text-foreground'
                                        }`}
                                        role='tab'
                                        aria-selected={activeTab === 'comments'}
                                    >
                                        Comments
                                    </button>
                                    <button
                                        onClick={() =>
                                            handleTabChange('upvoted')
                                        }
                                        className={`px-4 py-2 border-b-2 font-semibold whitespace-nowrap ${
                                            activeTab === 'upvoted' ?
                                                'border-primary-500 text-primary-600'
                                            :   'border-transparent text-muted-foreground hover:text-foreground'
                                        }`}
                                        role='tab'
                                        aria-selected={activeTab === 'upvoted'}
                                    >
                                        Upvoted
                                    </button>
                                    <button
                                        onClick={() =>
                                            handleTabChange('downvoted')
                                        }
                                        className={`px-4 py-2 border-b-2 font-semibold whitespace-nowrap ${
                                            activeTab === 'downvoted' ?
                                                'border-primary-500 text-primary-600'
                                            :   'border-transparent text-muted-foreground hover:text-foreground'
                                        }`}
                                        role='tab'
                                        aria-selected={
                                            activeTab === 'downvoted'
                                        }
                                    >
                                        Downvoted
                                    </button>
                                </nav>
                            </div>

                            {/* Tab Content */}
                            {activeTab === 'overview' && (
                                <div>
                                    {loadingReputation ?
                                        <ProfileSkeleton />
                                    : reputationError ?
                                        <EmptyStateWithAction
                                            icon={
                                                <svg
                                                    className='w-16 h-16'
                                                    fill='none'
                                                    stroke='currentColor'
                                                    viewBox='0 0 24 24'
                                                >
                                                    <path
                                                        strokeLinecap='round'
                                                        strokeLinejoin='round'
                                                        strokeWidth={2}
                                                        d='M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z'
                                                    />
                                                </svg>
                                            }
                                            title='Failed to load reputation'
                                            description="We couldn't load your reputation data. Please try again."
                                            primaryAction={{
                                                label: 'Retry',
                                                onClick: fetchReputation,
                                            }}
                                        />
                                    : reputation ?
                                        <ReputationDisplay
                                            reputation={reputation}
                                        />
                                    :   <EmptyStateWithAction
                                            title='No reputation data'
                                            description='Your reputation data is not available at this time.'
                                            primaryAction={{
                                                label: 'Refresh',
                                                onClick: fetchReputation,
                                            }}
                                        />
                                    }
                                </div>
                            )}

                            {activeTab === 'badges' && (
                                <div>
                                    {loadingReputation ?
                                        <div className='grid grid-cols-3 gap-4'>
                                            {[...Array(6)].map((_, i) => (
                                                <Skeleton
                                                    key={i}
                                                    variant='rectangular'
                                                    height={120}
                                                />
                                            ))}
                                        </div>
                                    : (
                                        reputation &&
                                        reputation.badges &&
                                        reputation.badges.length > 0
                                    ) ?
                                        <BadgeGrid
                                            badges={reputation.badges}
                                            columns={3}
                                        />
                                    :   <div className='py-12 text-center'>
                                            <p className='text-muted-foreground'>
                                                No badges earned yet. Keep
                                                contributing to earn badges!
                                            </p>
                                        </div>
                                    }
                                </div>
                            )}

                            {activeTab === 'karma' && (
                                <div>
                                    {karmaBreakdown ?
                                        <KarmaBreakdownChart
                                            breakdown={karmaBreakdown}
                                        />
                                    :   <div className='py-12 text-center'>
                                            <p className='text-muted-foreground'>
                                                Loading karma data...
                                            </p>
                                        </div>
                                    }
                                </div>
                            )}

                            {activeTab === 'comments' && (
                                <div>
                                    {loadingTabData ?
                                        <CommentSkeleton />
                                    : comments.length > 0 ?
                                        <div className='space-y-4'>
                                            {comments.map(comment => (
                                                <Card key={comment.id}>
                                                    <CardBody>
                                                        <Link
                                                            to={`/clips/${comment.clip_id}`}
                                                            className='text-sm text-primary-600 hover:underline mb-2 block'
                                                        >
                                                            View on clip
                                                        </Link>
                                                        <p className='text-foreground'>
                                                            {comment.content}
                                                        </p>
                                                        <div className='mt-2 text-sm text-muted-foreground'>
                                                            {comment.vote_score}{' '}
                                                            points •{' '}
                                                            <span
                                                                title={
                                                                    formatTimestamp(
                                                                        comment.created_at,
                                                                    ).title
                                                                }
                                                            >
                                                                {
                                                                    formatTimestamp(
                                                                        comment.created_at,
                                                                    ).display
                                                                }
                                                            </span>
                                                        </div>
                                                    </CardBody>
                                                </Card>
                                            ))}
                                            {hasMore && (
                                                <div className='text-center pt-4'>
                                                    <Button
                                                        onClick={() =>
                                                            fetchTabData(
                                                                'comments',
                                                                currentPage + 1,
                                                                true,
                                                            )
                                                        }
                                                        variant='outline'
                                                    >
                                                        Load More
                                                    </Button>
                                                </div>
                                            )}
                                        </div>
                                    :   <EmptyStateWithAction
                                            icon={
                                                <svg
                                                    className='w-16 h-16'
                                                    fill='none'
                                                    stroke='currentColor'
                                                    viewBox='0 0 24 24'
                                                >
                                                    <path
                                                        strokeLinecap='round'
                                                        strokeLinejoin='round'
                                                        strokeWidth={2}
                                                        d='M8 12h.01M12 12h.01M16 12h.01M21 12c0 4.418-4.03 8-9 8a9.863 9.863 0 01-4.255-.949L3 20l1.395-3.72C3.512 15.042 3 13.574 3 12c0-4.418 4.03-8 9-8s9 3.582 9 8z'
                                                    />
                                                </svg>
                                            }
                                            title='No comments yet'
                                            description='Start engaging with the community by commenting on clips.'
                                            primaryAction={{
                                                label: 'Browse Clips',
                                                href: '/',
                                            }}
                                            tips={[
                                                'Share your thoughts on clips you enjoy',
                                                'Engage in discussions with other users',
                                                'Earn karma by posting quality comments',
                                            ]}
                                        />
                                    }
                                </div>
                            )}

                            {activeTab === 'upvoted' && (
                                <div>
                                    {loadingTabData ?
                                        <div className='space-y-6'>
                                            {[...Array(5)].map((_, i) => (
                                                <ClipCardSkeleton key={i} />
                                            ))}
                                        </div>
                                    : upvotedClips.length > 0 ?
                                        <div className='space-y-6'>
                                            {upvotedClips.map(clip => (
                                                <ClipCard
                                                    key={clip.id}
                                                    clip={clip}
                                                />
                                            ))}
                                            {hasMore && (
                                                <div className='text-center pt-4'>
                                                    <Button
                                                        onClick={() =>
                                                            fetchTabData(
                                                                'upvoted',
                                                                currentPage + 1,
                                                                true,
                                                            )
                                                        }
                                                        variant='outline'
                                                    >
                                                        Load More
                                                    </Button>
                                                </div>
                                            )}
                                        </div>
                                    :   <EmptyStateWithAction
                                            icon={
                                                <svg
                                                    className='w-16 h-16'
                                                    fill='none'
                                                    stroke='currentColor'
                                                    viewBox='0 0 24 24'
                                                >
                                                    <path
                                                        strokeLinecap='round'
                                                        strokeLinejoin='round'
                                                        strokeWidth={2}
                                                        d='M14 10h4.764a2 2 0 011.789 2.894l-3.5 7A2 2 0 0115.263 21h-4.017c-.163 0-.326-.02-.485-.06L7 20m7-10V5a2 2 0 00-2-2h-.095c-.5 0-.905.405-.905.905 0 .714-.211 1.412-.608 2.006L7 11v9m7-10h-2M7 20H5a2 2 0 01-2-2v-6a2 2 0 012-2h2.5'
                                                    />
                                                </svg>
                                            }
                                            title='No upvoted clips'
                                            description='Start upvoting clips you enjoy to see them here.'
                                            primaryAction={{
                                                label: 'Discover Clips',
                                                href: '/discover',
                                            }}
                                            secondaryAction={{
                                                label: 'Browse Top Clips',
                                                href: '/top',
                                            }}
                                        />
                                    }
                                </div>
                            )}

                            {activeTab === 'downvoted' && (
                                <div>
                                    {loadingTabData ?
                                        <div className='space-y-6'>
                                            {[...Array(5)].map((_, i) => (
                                                <ClipCardSkeleton key={i} />
                                            ))}
                                        </div>
                                    : downvotedClips.length > 0 ?
                                        <div className='space-y-6'>
                                            {downvotedClips.map(clip => (
                                                <ClipCard
                                                    key={clip.id}
                                                    clip={clip}
                                                />
                                            ))}
                                            {hasMore && (
                                                <div className='text-center pt-4'>
                                                    <Button
                                                        onClick={() =>
                                                            fetchTabData(
                                                                'downvoted',
                                                                currentPage + 1,
                                                                true,
                                                            )
                                                        }
                                                        variant='outline'
                                                    >
                                                        Load More
                                                    </Button>
                                                </div>
                                            )}
                                        </div>
                                    :   <EmptyStateWithAction
                                            icon={
                                                <svg
                                                    className='w-16 h-16'
                                                    fill='none'
                                                    stroke='currentColor'
                                                    viewBox='0 0 24 24'
                                                >
                                                    <path
                                                        strokeLinecap='round'
                                                        strokeLinejoin='round'
                                                        strokeWidth={2}
                                                        d='M10 14H5.236a2 2 0 01-1.789-2.894l3.5-7A2 2 0 018.736 3h4.018a2 2 0 01.485.06l3.76.94m-7 10v5a2 2 0 002 2h.096c.5 0 .905-.405.905-.904 0-.715.211-1.413.608-2.008L17 13V4m-7 10h2m5-10h2a2 2 0 012 2v6a2 2 0 01-2 2h-2.5'
                                                    />
                                                </svg>
                                            }
                                            title='No downvoted clips'
                                            description='Downvoted clips will appear here.'
                                            primaryAction={{
                                                label: 'Browse Clips',
                                                href: '/',
                                            }}
                                        />
                                    }
                                </div>
                            )}
                        </Stack>
                    </CardBody>
                </Card>
            </div>
        </Container>
    );
}
