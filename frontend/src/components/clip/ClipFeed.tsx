import { useEffect, useRef, useState, useCallback, memo, type ReactNode } from 'react';
import { useInView } from 'react-intersection-observer';
import { useSearchParams } from 'react-router-dom';
import { Spinner, Button, ScrollToTop } from '@/components/ui';
import { MiniFooter } from '@/components/layout';
import { ClipCard } from './ClipCard';
import { DiscoverClipCard } from './DiscoverClipCard';
import { ClipCardSkeleton } from './ClipCardSkeleton';
import { EmptyState } from './EmptyState';
import { FeedHeader } from './FeedHeader';
import { useClipFeed } from '@/hooks/useClips';
import type { SortOption, TimeFrame, ClipFeedFilters } from '@/types/clip';
import { useFeedAutoplayPreference } from '@/hooks';
import { SettingsEvents, SubmissionEvents, trackEvent } from '@/lib/telemetry';

interface ClipFeedProps {
    title?: string;
    description?: string;
    defaultSort?: SortOption;
    defaultTimeframe?: TimeFrame;
    filters?: Partial<ClipFeedFilters>;
    showSearch?: boolean;
    useSortTitle?: boolean;
    /** When true, uses simplified cards focused on discovery and posting */
    discoverMode?: boolean;
    insertAfter?: number;
    insertedContent?: ReactNode;
}

// Map legacy 'hot' to 'trending' for consistency
const normalizeSortOption = (sort: SortOption): SortOption => {
    return sort === 'hot' ? 'trending' : sort;
};

// Memoized ClipCard wrapper for performance
const MemoizedClipCard = memo(ClipCard, (prevProps, nextProps) => {
    return (
        prevProps.active === nextProps.active &&
        prevProps.autoplay === nextProps.autoplay &&
        prevProps.clip.id === nextProps.clip.id &&
        prevProps.clip.vote_score === nextProps.clip.vote_score &&
        prevProps.clip.user_vote === nextProps.clip.user_vote &&
        prevProps.clip.is_favorited === nextProps.clip.is_favorited &&
        prevProps.clip.comment_count === nextProps.clip.comment_count &&
        prevProps.clip.favorite_count === nextProps.clip.favorite_count &&
        prevProps.clip.watch_progress?.progress_percent ===
            nextProps.clip.watch_progress?.progress_percent &&
        prevProps.clip.watch_progress?.completed ===
            nextProps.clip.watch_progress?.completed &&
        // Detect when watch_progress changes from undefined to defined or vice versa
        (prevProps.clip.watch_progress === undefined) ===
            (nextProps.clip.watch_progress === undefined)
    );
});

// Memoized DiscoverClipCard wrapper for performance
const MemoizedDiscoverClipCard = memo(
    DiscoverClipCard,
    (prevProps, nextProps) => {
        return (
            prevProps.clip.id === nextProps.clip.id &&
            prevProps.clip.view_count === nextProps.clip.view_count &&
            prevProps.clip.submitted_by?.id === nextProps.clip.submitted_by?.id
        );
    },
);

export function ClipFeed({
    title = 'Clip Feed',
    description,
    defaultSort = 'trending',
    defaultTimeframe = 'day',
    filters: additionalFilters = {},
    showSearch = false,
    useSortTitle = true,
    discoverMode = false,
    insertAfter = 5,
    insertedContent,
}: ClipFeedProps) {
    const [searchParams, setSearchParams] = useSearchParams();
    const containerRef = useRef<HTMLDivElement>(null);
    const [isRefreshing, setIsRefreshing] = useState(false);
    const [pullDistance, setPullDistance] = useState(0);
    const [activeClipId, setActiveClipId] = useState<string | null>(null);
    const visibleClipIdsRef = useRef(new Set<string>());
    const { preference: autoplayPreference, setPreference: setAutoplayPreference } =
        useFeedAutoplayPreference();
    const touchStartRef = useRef<number>(0);
    const scrollTopRef = useRef<number>(0);

    // Get filters from URL or use defaults (normalize 'hot' to 'trending')
    // In discover mode, ignore URL sort/timeframe to keep the feed controlled by tabs
    const rawSort =
        discoverMode ? defaultSort : (
            (searchParams.get('sort') as SortOption) || defaultSort
        );
    const sort = normalizeSortOption(rawSort);
    const timeframe =
        discoverMode ? defaultTimeframe : (
            (searchParams.get('timeframe') as TimeFrame) || defaultTimeframe
        );

    // Combine URL filters with additional filters and current language
    const filters: ClipFeedFilters = {
        sort,
        timeframe:
            sort === 'top' || sort === 'trending' ? timeframe : undefined,
        // Do not filter by UI language by default; include language only if explicitly provided
        ...additionalFilters,
    };

    // Fetch clips with infinite query
    const {
        data,
        fetchNextPage,
        hasNextPage,
        isFetchingNextPage,
        isLoading,
        isError,
        refetch,
    } = useClipFeed(filters);

    // Get all clips from all pages
    const clips = data?.pages.flatMap(page => page.clips) ?? [];
    const validClips = clips.filter(clip => clip?.id);

    const handleVisibilityChange = useCallback(
        (clipId: string, visible: boolean) => {
            const visibleIds = visibleClipIdsRef.current;
            if (visible) {
                visibleIds.add(clipId);
                if (autoplayPreference === 'muted') {
                    setActiveClipId(clipId);
                    trackEvent(SubmissionEvents.SUBMISSION_PLAY_STARTED, {
                        clip_id: clipId,
                        playback_mode: 'muted_autoplay',
                        section_name: 'clip_feed',
                    });
                }
                return;
            }

            visibleIds.delete(clipId);
            setActiveClipId(current => {
                if (current !== clipId) return current;
                return autoplayPreference === 'muted' ? (visibleIds.values().next().value ?? null) : null;
            });
        },
        [autoplayPreference],
    );

    const handleAutoplayPreferenceChange = useCallback(
        (value: 'manual' | 'muted') => {
            setAutoplayPreference(value);
            if (value === 'manual') setActiveClipId(null);
            trackEvent(SettingsEvents.FEED_AUTOPLAY_CHANGED, {
                setting_name: 'feed_autoplay',
                new_value: value,
            });
        },
        [setAutoplayPreference],
    );

    const handleActivate = useCallback((clipId: string) => {
        setActiveClipId(clipId);
        trackEvent(SubmissionEvents.SUBMISSION_PLAY_STARTED, {
            clip_id: clipId,
            playback_mode: 'manual',
            section_name: 'clip_feed',
        });
    }, []);

    // Intersection observer for infinite scroll
    const { ref: loadMoreRef, inView } = useInView({
        threshold: 0.5,
    });

    // Load more when the trigger element comes into view
    useEffect(() => {
        if (inView && hasNextPage && !isFetchingNextPage) {
            fetchNextPage();
        }
    }, [inView, hasNextPage, isFetchingNextPage, fetchNextPage]);

    // Pull-to-refresh handlers for mobile web
    const handleTouchStart = useCallback((e: React.TouchEvent) => {
        // Only enable pull-to-refresh when at the top of the page
        scrollTopRef.current =
            window.scrollY || document.documentElement.scrollTop;
        if (scrollTopRef.current === 0) {
            touchStartRef.current = e.touches[0].clientY;
        }
    }, []);

    const handleTouchMove = useCallback((e: React.TouchEvent) => {
        if (touchStartRef.current && scrollTopRef.current === 0) {
            const currentY = e.touches[0].clientY;
            const distance = Math.max(0, currentY - touchStartRef.current);

            // Only activate pull-to-refresh if user is pulling down
            if (distance > 0 && distance < 120) {
                setPullDistance(distance);
                // Prevent default scroll behavior when pulling down
                if (distance > 10) {
                    e.preventDefault();
                }
            }
        }
    }, []);

    const handleTouchEnd = useCallback(async () => {
        if (pullDistance > 80 && !isRefreshing && !isLoading) {
            setIsRefreshing(true);
            try {
                await refetch();
            } finally {
                setIsRefreshing(false);
            }
        }
        setPullDistance(0);
        touchStartRef.current = 0;
    }, [pullDistance, isRefreshing, isLoading, refetch]);

    const handleSortChange = (newSort: SortOption) => {
        const params = new URLSearchParams(searchParams);
        params.set('sort', newSort);
        if (newSort !== 'top' && newSort !== 'trending') {
            params.delete('timeframe');
        }
        setSearchParams(params);
        // Persist sort preference to localStorage
        try {
            localStorage.setItem('feedSort', newSort);
        } catch (error) {
            console.error('Failed to save sort preference:', error);
        }
    };

    const handleTimeframeChange = (newTimeframe: TimeFrame) => {
        const params = new URLSearchParams(searchParams);
        params.set('timeframe', newTimeframe);
        setSearchParams(params);
    };

    const sortLabelMap: Partial<Record<SortOption, string>> = {
        trending: 'Trending',
        popular: 'Most Popular',
        new: 'Newest',
        top: 'Top Rated',
        rising: 'Rising',
        discussed: 'Most Discussed',
    };

    const timeframeLabelMap: Partial<Record<TimeFrame, string>> = {
        hour: 'Past Hour',
        day: 'Past Day',
        week: 'Past Week',
        month: 'Past Month',
        year: 'Past Year',
        all: 'All Time',
    };

    const resolvedTitle =
        useSortTitle ?
            sort === 'top' || sort === 'trending' ?
                `${sortLabelMap[sort] ?? sort} — ${timeframeLabelMap[timeframe] ?? 'Past Day'}`
            :   `${sortLabelMap[sort] ?? sort} Feed`
        :   title;

    return (
        <div className='w-full'>
            {/* Hide FeedHeader in discover mode - DiscoveryPage has its own header/tabs */}
            {!discoverMode && (
                <FeedHeader
                    title={resolvedTitle || title}
                    description={description}
                    showSearch={showSearch}
                    sort={sort}
                    timeframe={timeframe}
                    onSortChange={handleSortChange}
                    onTimeframeChange={handleTimeframeChange}
                    autoplayPreference={autoplayPreference}
                    onAutoplayPreferenceChange={handleAutoplayPreferenceChange}
                />
            )}

            {/* Pull-to-refresh indicator */}
            {pullDistance > 0 && (
                <div
                    className='flex justify-center items-center py-4 text-muted-foreground transition-all'
                    style={{
                        transform: `translateY(${Math.min(pullDistance, 80)}px)`,
                        opacity: Math.min(pullDistance / 80, 1),
                    }}
                >
                    {isRefreshing ?
                        <Spinner size='md' />
                    :   <div className='flex flex-col items-center'>
                            <svg
                                className='w-6 h-6 mb-1'
                                fill='none'
                                stroke='currentColor'
                                viewBox='0 0 24 24'
                                style={{
                                    transform: `rotate(${pullDistance * 4}deg)`,
                                }}
                            >
                                <path
                                    strokeLinecap='round'
                                    strokeLinejoin='round'
                                    strokeWidth={2}
                                    d='M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15'
                                />
                            </svg>
                            <span className='text-xs'>
                                {pullDistance > 80 ?
                                    'Release to refresh'
                                :   'Pull to refresh'}
                            </span>
                        </div>
                    }
                </div>
            )}

            {/* Loading state */}
            {isLoading && (
                <div className='space-y-6'>
                    {Array.from({ length: 5 }).map((_, i) => (
                        <ClipCardSkeleton key={i} />
                    ))}
                </div>
            )}

            {/* Error state */}
            {isError && (
                <EmptyState
                    title='Error loading clips'
                    message='Something went wrong. Please try again later.'
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
                />
            )}

            {/* Empty state */}
            {!isLoading && !isError && validClips.length === 0 && (
                <EmptyState
                    title='No clips found'
                    message='Try adjusting your filters or check back later.'
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
                                d='M15 10l4.553-2.276A1 1 0 0121 8.618v6.764a1 1 0 01-1.447.894L15 14M5 18h8a2 2 0 002-2V8a2 2 0 00-2-2H5a2 2 0 00-2 2v8a2 2 0 002 2z'
                            />
                        </svg>
                    }
                />
            )}

            {/* Clips list with pull-to-refresh */}
            {!isLoading && !isError && validClips.length > 0 && (
                <div
                    ref={containerRef}
                    onTouchStart={handleTouchStart}
                    onTouchMove={handleTouchMove}
                    onTouchEnd={handleTouchEnd}
                >
                    <div className='space-y-5 md:space-y-6 snap-y snap-proximity motion-reduce:snap-none'>
                        {validClips.map((clip, index) => (
                            <div key={clip.id}>
                            {discoverMode ?
                                <MemoizedDiscoverClipCard
                                    clip={clip}
                                />
                            :   <MemoizedClipCard
                                    clip={clip}
                                    active={activeClipId === clip.id}
                                    autoplay={autoplayPreference === 'muted'}
                                    onActivate={handleActivate}
                                    onVisibilityChange={handleVisibilityChange}
                                />}
                            {insertedContent && index + 1 === insertAfter && (
                                <div className='mt-5 md:mt-8'>{insertedContent}</div>
                            )}
                            </div>
                        ))}
                    </div>

                    {/* Load more trigger */}
                    {hasNextPage && (
                        <div
                            ref={loadMoreRef}
                            className='py-8 flex justify-center'
                        >
                            {isFetchingNextPage ?
                                <Spinner size='lg' />
                            :   <Button onClick={() => fetchNextPage()}>
                                    Load More
                                </Button>
                            }
                        </div>
                    )}

                    {/* End of results */}
                    {!hasNextPage && validClips.length > 0 && (
                        <div className='text-center py-8 text-muted-foreground'>
                            <p>You've reached the end!</p>
                        </div>
                    )}
                </div>
            )}

            {/* Scroll to top button */}
            <div className='hidden md:block'><ScrollToTop threshold={500} /></div>

            {/* Mini footer for quick access to footer links */}
            <div className='hidden md:block'><MiniFooter /></div>
        </div>
    );
}
