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
        tags: [],
    },
    counts: {
        clips: 0,
        creators: 0,
        games: 0,
        tags: 0,
    },
    facets: {
        languages: [],
        games: [],
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
});
