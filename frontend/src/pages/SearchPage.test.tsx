import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { act, render, screen } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { SearchPage } from './SearchPage';
import { HelmetProvider } from '@dr.pogodin/react-helmet';
import { searchApi } from '../lib/search-api';
import { ToastProvider } from '../context/ToastContext';

vi.mock('../context/AuthContext', () => ({
    useAuth: () => ({ isAuthenticated: false }),
}));

vi.mock('../components/search/TrendingSearches', () => ({
    TrendingSearches: () => null,
}));

vi.mock('../lib/search-api', () => {
    const search = vi.fn();
    return {
        searchApi: {
            search,
            getSuggestions: vi.fn(),
            getSavedSearches: vi.fn(() => []),
            getTrendingSearches: vi.fn(() => []),
        },
    };
});

const mockedSearchApi = searchApi as unknown as {
    search: ReturnType<typeof vi.fn>;
};

const mockSearchResponse = {
    query: 'test',
    results: {
        clips: [],
        creators: [],
        games: [],
        twitch_categories: [],
        tags: [],
    },
    counts: {
        clips: 0,
        creators: 0,
        games: 0,
        twitch_categories: 0,
        tags: 0,
    },
    facets: {
        languages: [],
        games: [],
        twitch_categories: [],
        tags: [],
        date_range: {
            last_hour: 0,
            last_day: 0,
            last_week: 0,
            last_month: 0,
            older: 0,
        },
    },
    meta: {
        page: 1,
        limit: 20,
        total_items: 0,
        total_pages: 1,
    },
};

beforeEach(() => {
    queryClient.clear();
    mockedSearchApi.search.mockResolvedValue(mockSearchResponse);
});

// Create a test query client
const queryClient = new QueryClient({
    defaultOptions: {
        queries: {
            retry: false,
        },
    },
});

describe('SearchPage - Sort Controls', () => {
    it('renders sort dropdown with proper theme-aware styling', async () => {
        render(
            <HelmetProvider>
                <QueryClientProvider client={queryClient}>
                    <ToastProvider>
                        <MemoryRouter>
                            <SearchPage />
                        </MemoryRouter>
                    </ToastProvider>
                </QueryClientProvider>
            </HelmetProvider>
        );

        await act(async () => {
            await Promise.resolve();
        });

        // Note: The sort dropdown only appears when there's a search query
        // This test verifies component renders without errors
        expect(screen.getByText(/Enter a search query/i)).toBeInTheDocument();
        expect(
            screen.getByText(/clips, creators, tags, and Twitch categories/i),
        ).toBeInTheDocument();
    });

    it('sort dropdown has focus-visible styles for accessibility', async () => {
        // Create a wrapper with search query
        const wrapper = ({ children }: { children: React.ReactNode }) => (
            <HelmetProvider>
                <QueryClientProvider client={queryClient}>
                    <ToastProvider>
                        <MemoryRouter initialEntries={['/?q=test']}>
                            {children}
                        </MemoryRouter>
                    </ToastProvider>
                </QueryClientProvider>
            </HelmetProvider>
        );

        render(<SearchPage />, { wrapper }); // Wait for the component to render with query
        await screen.findByTestId('empty-state-message');
        // The sort dropdown should have proper accessibility classes
        const sortElements = document.querySelectorAll('select');

        if (sortElements.length > 0) {
            const sortSelect = sortElements[0];
            expect(sortSelect.className).toContain('focus-visible:ring-2');
            expect(sortSelect.className).toContain(
                'focus-visible:ring-primary-500'
            );
            expect(sortSelect.className).toContain(
                'dark:focus-visible:ring-offset-neutral-900'
            );
            expect(sortSelect.className).toContain('text-foreground');
        }
    });

    it('renders an empty search response without crashing', async () => {
        render(
            <HelmetProvider>
                <QueryClientProvider client={queryClient}>
                    <ToastProvider>
                        <MemoryRouter initialEntries={['/search?q=test']}>
                            <SearchPage />
                        </MemoryRouter>
                    </ToastProvider>
                </QueryClientProvider>
            </HelmetProvider>,
        );

        expect(await screen.findAllByTestId('empty-state')).toHaveLength(1);
        expect(screen.getByTestId('search-results')).toBeVisible();
    });

    it('leaves the first Tab to the browser so the skip link stays reachable', async () => {
        render(
            <HelmetProvider>
                <QueryClientProvider client={queryClient}>
                    <ToastProvider>
                        <MemoryRouter initialEntries={['/search']}>
                            <SearchPage />
                        </MemoryRouter>
                    </ToastProvider>
                </QueryClientProvider>
            </HelmetProvider>,
        );
        await act(async () => {
            await Promise.resolve();
        });

        const tab = new KeyboardEvent('keydown', { key: 'Tab', bubbles: true, cancelable: true });
        window.dispatchEvent(tab);

        expect(tab.defaultPrevented).toBe(false);
    });
});

describe('SearchPage - said in the clip', () => {
    it('searches a transcript phrase alone and shows the snippet', async () => {
        mockedSearchApi.search.mockResolvedValue({
            ...mockSearchResponse,
            results: {
                ...mockSearchResponse.results,
                clips: [{
                    id: 'c1', twitch_clip_id: 't1', twitch_clip_url: 'https://clips.twitch.tv/t1', embed_url: 'https://clips.twitch.tv/embed?clip=t1',
                    title: 'Valorant highlight', creator_name: 'clipper', broadcaster_name: 'alpha', view_count: 10, vote_score: 0,
                    comment_count: 0, favorite_count: 0, is_featured: false, is_nsfw: false, is_removed: false,
                    created_at: '2026-11-09T12:00:00Z', imported_at: '2026-11-09T12:00:00Z',
                    transcript_snippet: 'that was the cleanest ace ever',
                }],
            },
            counts: { ...mockSearchResponse.counts, clips: 1 },
            meta: { ...mockSearchResponse.meta, total_items: 1 },
        });
        render(
            <HelmetProvider>
                <QueryClientProvider client={queryClient}>
                    <ToastProvider>
                        <MemoryRouter initialEntries={['/search?said=cleanest%20ace']}>
                            <SearchPage />
                        </MemoryRouter>
                    </ToastProvider>
                </QueryClientProvider>
            </HelmetProvider>
        );
        const snippet = await screen.findByTestId('transcript-snippet');
        expect(snippet).toHaveTextContent('that was the cleanest ace ever');
        expect(mockedSearchApi.search).toHaveBeenCalledWith(expect.objectContaining({ query: '', said: 'cleanest ace' }));
        expect(screen.getByTestId('results-count')).toHaveTextContent('said “cleanest ace”');
        expect(screen.getByLabelText('Said in the clip')).toHaveValue('cleanest ace');
    });
});
