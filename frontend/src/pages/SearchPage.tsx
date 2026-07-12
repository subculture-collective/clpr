import { useQuery } from '@tanstack/react-query';
import { useEffect, useMemo, useRef, useState } from 'react';
import { useSearchParams } from 'react-router-dom';
import { Container, SEO } from '../components';
import { ClipCard } from '../components/clip';
import { SearchBar, SearchFilters } from '../components/search';
import { SearchErrorAlert } from '../components/search/SearchErrorAlert';
import { TrendingSearches } from '../components/search/TrendingSearches';
import { SearchHistory } from '../components/search/SearchHistory';
import { SavedSearches } from '../components/search/SavedSearches';
import { SearchResultSkeleton, EmptyStateWithAction } from '../components/ui';
import { SaveSearchModal } from '../components/ui/SaveSearchModal';
import { useSearchErrorState } from '../hooks/useSearchErrorState';
import { useSearchHistory } from '../hooks/useSearchHistory';
import { useSavedSearches } from '../hooks/useSavedSearches';
import { useToast } from '../context/ToastContext';
import { searchApi } from '../lib/search-api';
import type {
    SearchRequest,
    SearchResponse,
    SearchFilters as SearchFiltersType,
} from '../types/search';

export function SearchPage() {
    const [searchParams, setSearchParams] = useSearchParams();
    const toast = useToast();
    const query = searchParams.get('q') || '';
    const typeParam = searchParams.get('type') || 'all';
    const sortParam = searchParams.get('sort') || 'relevance';
    const pageParam = (() => {
        const p = parseInt(searchParams.get('page') || '1', 10);
        return isNaN(p) || p < 1 ? 1 : p;
    })();
    const languageParam = searchParams.get('language') || undefined;
    const gameIdParam = searchParams.get('game_id') || undefined;
    const dateFromParam = searchParams.get('date_from') || undefined;
    const dateToParam = searchParams.get('date_to') || undefined;
    const minVotesParam = searchParams.get('min_votes');
    const tagsParam = searchParams.get('tags');

    type SearchType = NonNullable<SearchRequest['type']>;
    type SortType = NonNullable<SearchRequest['sort']>;
    const [activeTab, setActiveTab] = useState<SearchType>(
        typeParam as SearchType,
    );

    const [filters, setFilters] = useState<SearchFiltersType>({
        language: languageParam,
        gameId: gameIdParam,
        dateFrom: dateFromParam,
        dateTo: dateToParam,
        minVotes: minVotesParam ? parseInt(minVotesParam, 10) : undefined,
        tags: tagsParam ? tagsParam.split(',') : undefined,
    });

    const [showSaveSearchModal, setShowSaveSearchModal] = useState(false);

    // Persist the latest search params so navigation back restores state
    const searchParamsString = useMemo(
        () => searchParams.toString(),
        [searchParams],
    );

    const hasRestoredParams = useRef(false);

    useEffect(() => {
        if (searchParamsString) {
            sessionStorage.setItem('search:lastParams', searchParamsString);
        }
    }, [searchParamsString]);

    useEffect(() => {
        if (query || hasRestoredParams.current) return;

        const stored = sessionStorage.getItem('search:lastParams');
        if (stored) {
            hasRestoredParams.current = true;
            setSearchParams(new URLSearchParams(stored), { replace: true });
        }
    }, [query, setSearchParams]);

    useEffect(() => {
        const handlePopState = () => {
            const current = new URL(window.location.href);
            if (current.searchParams.get('q')) return;

            const stored = sessionStorage.getItem('search:lastParams');
            if (stored) {
                const restored = new URLSearchParams(stored);
                const nextUrl = `${current.pathname}?${restored.toString()}`;
                window.history.replaceState(window.history.state, '', nextUrl);
                setSearchParams(restored, { replace: true });
            }
        };

        window.addEventListener('popstate', handlePopState);
        return () => window.removeEventListener('popstate', handlePopState);
    }, [setSearchParams]);

    // Ensure the search input receives the first Tab focus on this page
    useEffect(() => {
        const handler = (event: KeyboardEvent) => {
            if (event.key !== 'Tab') return;
            const input = document.querySelector(
                '[data-testid="search-input"]',
            ) as HTMLInputElement | null;
            if (input) {
                event.preventDefault();
                input.focus();
            }
        };

        window.addEventListener('keydown', handler, { once: true });
        return () => window.removeEventListener('keydown', handler);
    }, []);

    // Search error state management
    const {
        errorState,
        handleSearchError,
        handleSearchSuccess,
        retry,
        cancelRetry,
        dismissError,
    } = useSearchErrorState();

    // Search history management
    const { addToHistory } = useSearchHistory();

    // Saved searches management
    const { saveSearch } = useSavedSearches();

    // Fetch search results
    const { data, isLoading, error, refetch } = useQuery<SearchResponse>({
        queryKey: ['search', query, activeTab, sortParam, pageParam, filters],
        queryFn: async () => {
            try {
                const result = await searchApi.search({
                    query,
                    type: activeTab,
                    sort: sortParam as SortType,
                    page: pageParam,
                    limit: 20,
                    language: filters.language,
                    gameId: filters.gameId,
                    dateFrom: filters.dateFrom,
                    dateTo: filters.dateTo,
                    minVotes: filters.minVotes,
                    tags: filters.tags,
                });

                // Add to search history
                const totalResults =
                    (result.counts.clips || 0) +
                    (result.counts.creators || 0) +
                    (result.counts.games || 0) +
                    (result.counts.tags || 0);
                addToHistory(query, totalResults);

                // Check for failover indicators even on successful responses
                // Note: This requires the API client to expose response headers
                // For now, we'll just clear the error on success
                handleSearchSuccess();
                return result;
            } catch (err) {
                handleSearchError(err);
                throw err;
            }
        },
        enabled: query.length > 0,
        retry: false, // We handle retries manually
    });

    // Trigger automatic retry when an error occurs
    useEffect(() => {
        if (
            error &&
            errorState.type === 'error' &&
            !errorState.isRetrying &&
            errorState.retryCount === 0 &&
            !errorState.isCircuitOpen
        ) {
            // Only trigger on initial error (retryCount === 0) to avoid retry loops
            // The retry() function itself handles the retry loop with exponential backoff
            retry(() => refetch());
        }
    }, [
        error,
        errorState.type,
        errorState.isRetrying,
        errorState.retryCount,
        errorState.isCircuitOpen,
        retry,
        refetch,
    ]);

    // Update tab when type param changes
    useEffect(() => {
        queueMicrotask(() => {
            setActiveTab(typeParam as SearchType);
        });
    }, [typeParam]);

    // Update filters from URL params
    useEffect(() => {
        queueMicrotask(() => {
            setFilters({
                language: languageParam,
                gameId: gameIdParam,
                dateFrom: dateFromParam,
                dateTo: dateToParam,
                minVotes:
                    minVotesParam ? parseInt(minVotesParam, 10) : undefined,
                tags: tagsParam ? tagsParam.split(',') : undefined,
            });
        });
    }, [
        languageParam,
        gameIdParam,
        dateFromParam,
        dateToParam,
        minVotesParam,
        tagsParam,
    ]);

    // Handle tab change
    const handleTabChange = (newTab: SearchType) => {
        setActiveTab(newTab);
        const newParams = new URLSearchParams(searchParams);
        newParams.set('type', newTab);
        // Reset to first page when changing tab
        newParams.set('page', '1');
        setSearchParams(newParams);
    };

    // Handle sort change
    const handleSortChange = (newSort: SortType) => {
        const newParams = new URLSearchParams(searchParams);
        newParams.set('sort', newSort);
        // Reset to first page when changing sort
        newParams.set('page', '1');
        setSearchParams(newParams);
    };

    // Handle search from SearchBar
    const handleSearch = (newQuery: string) => {
        const newParams = new URLSearchParams();
        newParams.set('q', newQuery);
        newParams.set('type', 'all');
        newParams.set('sort', 'relevance');
        newParams.set('page', '1');
        setSearchParams(newParams);
    };

    // Handle retry from error alert
    const handleRetry = () => {
        retry(() => refetch());
    };

    // Handle filters change
    const handleFiltersChange = (newFilters: SearchFiltersType) => {
        setFilters(newFilters);
        const newParams = new URLSearchParams(searchParams);

        // Update URL params
        if (newFilters.language) {
            newParams.set('language', newFilters.language);
        } else {
            newParams.delete('language');
        }

        if (newFilters.gameId) {
            newParams.set('game_id', newFilters.gameId);
        } else {
            newParams.delete('game_id');
        }

        if (newFilters.dateFrom) {
            newParams.set('date_from', newFilters.dateFrom);
        } else {
            newParams.delete('date_from');
        }

        // Add dateTo
        if (newFilters.dateTo) {
            newParams.set('date_to', newFilters.dateTo);
        } else {
            newParams.delete('date_to');
        }

        // Add minVotes
        if (
            typeof newFilters.minVotes === 'number' &&
            !isNaN(newFilters.minVotes) &&
            newFilters.minVotes > 0
        ) {
            newParams.set('min_votes', String(newFilters.minVotes));
        } else {
            newParams.delete('min_votes');
        }

        // Add tags (as comma-separated)
        if (Array.isArray(newFilters.tags) && newFilters.tags.length > 0) {
            newParams.set('tags', newFilters.tags.join(','));
        } else {
            newParams.delete('tags');
        }

        setSearchParams(newParams);
    };

    // Handle save search
    const handleSaveSearch = (name: string) => {
        // Only treat filters as active if at least one has a meaningful value
        const hasActiveFilters =
            !!filters.language ||
            !!filters.gameId ||
            !!filters.dateFrom ||
            !!filters.dateTo ||
            (typeof filters.minVotes === 'number' &&
                !isNaN(filters.minVotes) &&
                filters.minVotes > 0) ||
            (Array.isArray(filters.tags) && filters.tags.length > 0);

        const activeFilters = hasActiveFilters ? filters : undefined;
        saveSearch(query, activeFilters, name || undefined);

        toast.success('Search saved successfully!');
    };

    const goToNextPage = () => {
        const nextPage = (data?.meta.page || pageParam) + 1;
        const totalPages = data?.meta.total_pages || nextPage;
        const targetPage = Math.min(nextPage, totalPages);
        const newParams = new URLSearchParams(searchParams);
        newParams.set('page', String(targetPage));
        setSearchParams(newParams);
    };

    const goToPreviousPage = () => {
        const prevPage = Math.max(1, (data?.meta.page || pageParam) - 1);
        const newParams = new URLSearchParams(searchParams);
        // Either set to 1 or remove param to satisfy tests
        newParams.set('page', String(prevPage));
        setSearchParams(newParams);
    };

    const seoTitle = query ? `Search: ${query}` : 'Search Clips';
    const seoDescription =
        query ?
            `Search results for "${query}" on Clipper. Find Twitch clips, games, creators, and tags matching your query.`
        :   'Search for Twitch clips, games, creators, and tags on Clipper. Discover amazing gaming moments from your favorite streamers.';

    if (!query.trim()) {
        return (
            <>
                <SEO
                    title={seoTitle}
                    description={seoDescription}
                    canonicalUrl='/search'
                    noindex
                />
                <Container className='py-4 xs:py-6 md:py-8'>
                    <div className='mb-6 xs:mb-8 flex justify-center'>
                        <SearchBar
                            initialQuery={query}
                            onSearch={handleSearch}
                            autoFocus
                        />
                    </div>

                    {/* Search Discovery Components */}
                    <div className='max-w-4xl mx-auto'>
                        <div className='text-center text-muted-foreground py-8 xs:py-12 mb-8'>
                            <p className='text-base xs:text-lg px-4'>
                                Enter a search query to find clips, games,
                                creators, and tags.
                            </p>
                        </div>

                        <div className='grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-8 px-4'>
                            <TrendingSearches limit={10} days={7} />
                            <SearchHistory maxItems={10} />
                            <SavedSearches />
                        </div>
                    </div>
                </Container>
            </>
        );
    }

    const tabs = [
        {
            id: 'all',
            label: 'All',
            count:
                (data?.counts.clips || 0) +
                (data?.counts.creators || 0) +
                (data?.counts.games || 0) +
                (data?.counts.tags || 0),
        },
        { id: 'clips', label: 'Clips', count: data?.counts.clips || 0 },
        {
            id: 'creators',
            label: 'Creators',
            count: data?.counts.creators || 0,
        },
        { id: 'games', label: 'Games', count: data?.counts.games || 0 },
        { id: 'tags', label: 'Tags', count: data?.counts.tags || 0 },
    ];

    return (
        <>
            <SEO
                title={seoTitle}
                description={seoDescription}
                canonicalUrl='/search'
                noindex={!query} // Don't index empty search page
            />
            <Container className='py-4 xs:py-6 md:py-8'>
                {/* Search Bar */}
                <div className='mb-6 xs:mb-8 flex justify-center'>
                    <SearchBar
                        initialQuery={query}
                        onSearch={handleSearch}
                        autoFocus
                    />
                </div>

                {/* Header */}
                <div className='mb-4 xs:mb-6 flex items-start justify-between gap-4'>
                    <div className='flex-1'>
                        <h1 className='text-2xl xs:text-3xl font-bold mb-2'>
                            Search Results
                        </h1>
                        <p
                            className='text-sm xs:text-base text-muted-foreground'
                            data-testid='results-count'
                        >
                            Found {data?.meta.total_items || 0} results for:{' '}
                            <span className='font-semibold'>"{query}"</span>
                        </p>
                    </div>
                    <button
                        onClick={() => setShowSaveSearchModal(true)}
                        className='flex items-center gap-2 px-3 py-2 text-sm rounded-lg border border-border hover:bg-accent transition-colors'
                        title='Save this search'
                        data-testid='save-search-button'
                    >
                        <svg
                            className='w-4 h-4'
                            fill='none'
                            stroke='currentColor'
                            viewBox='0 0 24 24'
                        >
                            <path
                                strokeLinecap='round'
                                strokeLinejoin='round'
                                strokeWidth={2}
                                d='M5 5a2 2 0 012-2h10a2 2 0 012 2v16l-7-3.5L5 21V5z'
                            />
                        </svg>
                        <span className='hidden xs:inline'>Save</span>
                    </button>
                </div>

                {/* Tabs and Sort */}
                <div className='flex flex-col xs:flex-row xs:items-center xs:justify-between gap-3 xs:gap-0 mb-4 xs:mb-6 border-b border-border'>
                    <div className='flex gap-0.5 xs:gap-1 overflow-x-auto scrollbar-hide'>
                        {tabs.map(tab => (
                            <button
                                key={tab.id}
                                onClick={() =>
                                    handleTabChange(tab.id as SearchType)
                                }
                                className={`px-3 xs:px-4 py-2 font-medium transition-colors relative whitespace-nowrap touch-target text-sm xs:text-base ${
                                    activeTab === tab.id ?
                                        'text-primary-300'
                                    :   'text-muted-foreground hover:text-foreground'
                                }`}
                                data-testid={`tab-${tab.id}`}
                            >
                                {tab.label}
                                {tab.count > 0 && (
                                    <span className='ml-1.5 xs:ml-2 text-xs bg-accent px-1.5 xs:px-2 py-0.5 rounded-full'>
                                        {tab.count}
                                    </span>
                                )}
                                {activeTab === tab.id && (
                                    <div className='absolute bottom-0 left-0 right-0 h-0.5 bg-primary' />
                                )}
                            </button>
                        ))}
                    </div>

                    <select
                        aria-label='Sort search results'
                        value={sortParam}
                        onChange={e =>
                            handleSortChange(e.target.value as SortType)
                        }
                        className='px-3 py-2 rounded-md border border-border bg-background text-foreground text-sm focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-primary-500 focus-visible:ring-offset-2 dark:focus-visible:ring-offset-neutral-900 touch-target'
                        data-testid='search-sort-select'
                    >
                        <option value='relevance'>Relevance</option>
                        <option value='recent'>Recent</option>
                        <option value='popular'>Popular</option>
                    </select>
                </div>

                {/* Filters */}
                {(activeTab === 'all' || activeTab === 'clips') &&
                    data?.facets && (
                        <SearchFilters
                            facets={data.facets}
                            filters={filters}
                            onFiltersChange={handleFiltersChange}
                        />
                    )}

                {/* Search Error/Failover Alert */}
                <SearchErrorAlert
                    type={errorState.type}
                    message={errorState.message}
                    retryCount={errorState.retryCount}
                    maxRetries={errorState.maxRetries}
                    onRetry={handleRetry}
                    isRetrying={errorState.isRetrying}
                    onCancelRetry={cancelRetry}
                    onDismiss={dismissError}
                    isCircuitOpen={errorState.isCircuitOpen}
                />

                {/* Loading State */}
                {isLoading && <SearchResultSkeleton />}

                {/* Error State - Only show fallback error UI if SearchErrorAlert is not displaying an error */}
                {error && errorState.type === 'none' && (
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
                        title='Search failed'
                        description='Failed to load search results. Please try again.'
                        primaryAction={{
                            label: 'Try Again',
                            onClick: () => window.location.reload(),
                        }}
                        secondaryAction={{
                            label: 'Go Home',
                            href: '/',
                        }}
                    />
                )}

                {/* Results */}
                {data && !isLoading && (
                    <div className='space-y-8' data-testid='search-results'>
                        {/* Clips */}
                        {(activeTab === 'all' || activeTab === 'clips') &&
                            data.results.clips &&
                            data.results.clips.length > 0 && (
                                <section>
                                    <h2 className='text-xl font-bold mb-4'>
                                        Clips{' '}
                                        {activeTab === 'all' &&
                                            `(${data.counts.clips})`}
                                    </h2>
                                    <div className='space-y-6'>
                                        {data.results.clips.map(clip => (
                                            <ClipCard
                                                key={clip.id}
                                                clip={clip}
                                            />
                                        ))}
                                    </div>
                                </section>
                            )}

                        {/* Creators */}
                        {(activeTab === 'all' || activeTab === 'creators') &&
                            data.results.creators &&
                            data.results.creators.length > 0 && (
                                <section>
                                    <h2 className='text-xl font-bold mb-4'>
                                        Creators{' '}
                                        {activeTab === 'all' &&
                                            `(${data.counts.creators})`}
                                    </h2>
                                    <div className='grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-6'>
                                        {data.results.creators.map(creator => (
                                            <div
                                                key={creator.id}
                                                className='p-4 rounded-lg border border-border hover:border-primary transition-colors search-result-card'
                                            >
                                                <div className='flex items-center gap-3'>
                                                    {creator.avatar_url && (
                                                        <img
                                                            src={
                                                                creator.avatar_url
                                                            }
                                                            alt={
                                                                creator.display_name
                                                            }
                                                            className='w-12 h-12 rounded-full'
                                                        />
                                                    )}
                                                    <div>
                                                        <h3 className='font-semibold'>
                                                            {
                                                                creator.display_name
                                                            }
                                                        </h3>
                                                        <p className='text-sm text-muted-foreground'>
                                                            @{creator.username}
                                                        </p>
                                                    </div>
                                                </div>
                                                {creator.bio && (
                                                    <p className='mt-2 text-sm text-muted-foreground line-clamp-2'>
                                                        {creator.bio}
                                                    </p>
                                                )}
                                                <div className='mt-2 text-xs text-muted-foreground'>
                                                    {creator.karma_points} karma
                                                </div>
                                            </div>
                                        ))}
                                    </div>
                                </section>
                            )}

                        {/* Games */}
                        {(activeTab === 'all' || activeTab === 'games') &&
                            data.results.games &&
                            data.results.games.length > 0 && (
                                <section>
                                    <h2 className='text-xl font-bold mb-4'>
                                        Games{' '}
                                        {activeTab === 'all' &&
                                            `(${data.counts.games})`}
                                    </h2>
                                    <div className='grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-6'>
                                        {data.results.games.map(game => (
                                            <div
                                                key={game.id}
                                                className='p-4 rounded-lg border border-border hover:border-primary transition-colors search-result-card'
                                            >
                                                <h3 className='font-semibold text-lg'>
                                                    {game.name}
                                                </h3>
                                                <p className='text-sm text-muted-foreground mt-1'>
                                                    {game.clip_count} clips
                                                </p>
                                            </div>
                                        ))}
                                    </div>
                                </section>
                            )}

                        {/* Tags */}
                        {(activeTab === 'all' || activeTab === 'tags') &&
                            data.results.tags &&
                            data.results.tags.length > 0 && (
                                <section>
                                    <h2 className='text-xl font-bold mb-4'>
                                        Tags{' '}
                                        {activeTab === 'all' &&
                                            `(${data.counts.tags})`}
                                    </h2>
                                    <div className='flex flex-wrap gap-2'>
                                        {data.results.tags.map(tag => (
                                            <div
                                                key={tag.id}
                                                className='px-3 py-1.5 rounded-full border border-border hover:border-primary transition-colors search-result-card'
                                                style={
                                                    tag.color ?
                                                        {
                                                            borderColor:
                                                                tag.color,
                                                        }
                                                    :   undefined
                                                }
                                            >
                                                <span className='font-medium'>
                                                    #{tag.name}
                                                </span>
                                                <span className='ml-2 text-xs text-muted-foreground'>
                                                    {tag.usage_count}
                                                </span>
                                            </div>
                                        ))}
                                    </div>
                                </section>
                            )}

                        {/* Empty State */}
                        {data.meta.total_items === 0 && (
                            <div data-testid='empty-state'>
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
                                                d='M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z'
                                            />
                                        </svg>
                                    }
                                    title={`No results found for "${query}"`}
                                    description='Try different keywords or adjust your filters'
                                    primaryAction={{
                                        label: 'Clear Filters',
                                        onClick: () => {
                                            setSearchParams({
                                                q: query,
                                                type: 'all',
                                                sort: 'relevance',
                                                page: '1',
                                            });
                                        },
                                    }}
                                    secondaryAction={{
                                        label: 'Browse All Clips',
                                        href: '/',
                                    }}
                                    tips={[
                                        'Check your spelling and try again',
                                        'Try more general keywords',
                                        'Remove filters to broaden your search',
                                    ]}
                                />
                            </div>
                        )}

                        {/* Pagination */}
                        {data.meta.total_pages > 1 || pageParam > 1 ?
                            <div className='mt-6 flex items-center justify-center gap-4'>
                                <button
                                    className='px-3 py-2 rounded-md border border-border bg-background text-foreground text-sm disabled:opacity-50'
                                    onClick={goToPreviousPage}
                                    disabled={
                                        (data?.meta.page ?? pageParam) <= 1
                                    }
                                    data-testid='pagination-prev'
                                >
                                    Prev
                                </button>
                                <span
                                    className='text-sm text-muted-foreground'
                                    data-testid='page-number'
                                >
                                    Page {data?.meta.page ?? pageParam} of{' '}
                                    {data?.meta.total_pages ?? 1}
                                </span>
                                <button
                                    className='px-3 py-2 rounded-md border border-border bg-background text-foreground text-sm disabled:opacity-50'
                                    onClick={goToNextPage}
                                    disabled={
                                        (data?.meta.page ?? pageParam) >=
                                        (data?.meta.total_pages ?? pageParam)
                                    }
                                    data-testid='pagination-next'
                                >
                                    Next
                                </button>
                            </div>
                        :   null}
                    </div>
                )}
            </Container>

            <SaveSearchModal
                open={showSaveSearchModal}
                onClose={() => setShowSaveSearchModal(false)}
                onSave={handleSaveSearch}
            />
        </>
    );
}
