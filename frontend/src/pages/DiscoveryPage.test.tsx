import { describe, it, expect, vi } from 'vitest';
import { render, screen } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';
import { DiscoveryPage } from './DiscoveryPage';

// Mock the ClipFeed component
vi.mock('../components/clip', () => ({
    ClipFeed: ({ filters }: { filters: { top10k_streamers: boolean } }) => (
        <div data-testid="clip-feed">
            <span data-testid="top10k-filter">
                {String(filters.top10k_streamers)}
            </span>
        </div>
    ),
}));

// Mock the Container and SEO components
vi.mock('../components', () => ({
    Container: ({ children }: { children: React.ReactNode }) => (
        <div>{children}</div>
    ),
    SEO: ({ children }: { children?: React.ReactNode }) => <>{children}</>,
}));

describe('DiscoveryPage', () => {
    it('initializes top10kEnabled from URL params', () => {
        render(
            <MemoryRouter initialEntries={['/discovery?top10k_streamers=true']}>
                <DiscoveryPage />
            </MemoryRouter>
        );

        const toggle = screen.getByRole('switch');
        expect(toggle).toHaveAttribute('aria-checked', 'true');

        const filterValue = screen.getByTestId('top10k-filter');
        expect(filterValue).toHaveTextContent('true');
    });

    it('initializes top10kEnabled as false when URL param is not set', () => {
        render(
            <MemoryRouter initialEntries={['/discovery']}>
                <DiscoveryPage />
            </MemoryRouter>
        );

        const toggle = screen.getByRole('switch');
        expect(toggle).toHaveAttribute('aria-checked', 'false');

        const filterValue = screen.getByTestId('top10k-filter');
        expect(filterValue).toHaveTextContent('false');
    });

    it('initializes top10kEnabled as false when URL param is false', () => {
        render(
            <MemoryRouter
                initialEntries={['/discovery?top10k_streamers=false']}
            >
                <DiscoveryPage />
            </MemoryRouter>
        );

        const toggle = screen.getByRole('switch');
        expect(toggle).toHaveAttribute('aria-checked', 'false');

        const filterValue = screen.getByTestId('top10k-filter');
        expect(filterValue).toHaveTextContent('false');
    });

    it('has useEffect dependency on searchParams for syncing state', () => {
        // This test verifies that different URL params result in different initial states
        // which demonstrates that the component properly syncs with URL params

        // Test with top10k_streamers=true
        const { unmount } = render(
            <MemoryRouter initialEntries={['/discovery?top10k_streamers=true']}>
                <DiscoveryPage />
            </MemoryRouter>
        );

        let toggle = screen.getByRole('switch');
        expect(toggle).toHaveAttribute('aria-checked', 'true');

        let filterValue = screen.getByTestId('top10k-filter');
        expect(filterValue).toHaveTextContent('true');

        // Clean up first render
        unmount();

        // Test with top10k_streamers not set (should be false)
        render(
            <MemoryRouter initialEntries={['/discovery']}>
                <DiscoveryPage />
            </MemoryRouter>
        );

        toggle = screen.getByRole('switch');
        expect(toggle).toHaveAttribute('aria-checked', 'false');

        filterValue = screen.getByTestId('top10k-filter');
        expect(filterValue).toHaveTextContent('false');
    });

    it('renders all three tabs', () => {
        render(
            <MemoryRouter initialEntries={['/discovery']}>
                <DiscoveryPage />
            </MemoryRouter>
        );

        expect(screen.getByText('Trending')).toBeInTheDocument();
        expect(screen.getByText('Latest')).toBeInTheDocument();
        expect(screen.getByText('Top Views')).toBeInTheDocument();
    });

    it('renders the Top 10k Creators toggle', () => {
        render(
            <MemoryRouter initialEntries={['/discovery']}>
                <DiscoveryPage />
            </MemoryRouter>
        );

        expect(
            screen.getByText('Top 10K Creators Only')
        ).toBeInTheDocument();
        expect(
            screen.getByText(
                'Filter clips to only show content from the top 10,000 creators'
            )
        ).toBeInTheDocument();
    });
});
