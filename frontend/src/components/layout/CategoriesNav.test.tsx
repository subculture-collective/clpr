import { fireEvent, render, screen } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { MemoryRouter } from 'react-router-dom';
import { describe, expect, it, vi } from 'vitest';
import { categoryApi } from '../../lib/category-api';
import { tagApi } from '../../lib/tag-api';
import { CategoriesNav } from './CategoriesNav';

function renderNav() {
    const client = new QueryClient({ defaultOptions: { queries: { retry: false } } });
    return render(
        <QueryClientProvider client={client}>
            <MemoryRouter>
                <CategoriesNav />
            </MemoryRouter>
        </QueryClientProvider>,
    );
}

vi.mock('../../lib/category-api', () => ({
    categoryApi: {
        listCategories: vi.fn().mockResolvedValue({
            categories: [
                {
                    id: 'topic-1',
                    name: 'IRL',
                    slug: 'irl',
                    position: 1,
                    is_featured: true,
                    created_at: '',
                    updated_at: '',
                },
                {
                    id: 'topic-2',
                    name: 'Unfeatured',
                    slug: 'unfeatured',
                    position: 2,
                    is_featured: false,
                    created_at: '',
                    updated_at: '',
                },
            ],
        }),
    },
}));

vi.mock('../../lib/tag-api', () => ({
    tagApi: {
        listTags: vi.fn().mockResolvedValue({ tags: [] }),
    },
}));

vi.mock('../../lib/broadcaster-api', () => ({
    fetchPopularBroadcasters: vi.fn().mockResolvedValue([
        {
            broadcaster_id: 'creator-1',
            broadcaster_name: 'Creator One',
            clip_count: 12,
        },
    ]),
}));

describe('CategoriesNav', () => {
    it('leads with creators and presents topics as the second discovery axis', async () => {
        renderNav();

        expect(await screen.findByRole('button', { name: 'Creators' })).toBeInTheDocument();
        expect(screen.getByRole('button', { name: 'Topics' })).toBeInTheDocument();
        expect(screen.getByRole('link', { name: /Creator One/ })).toHaveAttribute(
            'href',
            '/broadcaster/creator-1',
        );
        expect(screen.queryByText('Streamers')).not.toBeInTheDocument();
        expect(screen.queryByText('Games')).not.toBeInTheDocument();
    });

    it('reuses the shared topic and tag queries instead of its own variants', async () => {
        vi.mocked(categoryApi.listCategories).mockClear();
        vi.mocked(tagApi.listTags).mockClear();
        renderNav();

        fireEvent.click(await screen.findByRole('button', { name: 'Topics' }));
        expect(await screen.findByRole('link', { name: /IRL/ })).toHaveAttribute('href', '/topics/irl');
        expect(screen.queryByRole('link', { name: /Unfeatured/ })).not.toBeInTheDocument();
        expect(categoryApi.listCategories).toHaveBeenCalledTimes(1);
        expect(categoryApi.listCategories).toHaveBeenCalledWith({ type: 'topic', public: true });
        expect(tagApi.listTags).toHaveBeenCalledWith({ sort: 'popularity', limit: 20 });
    });
});
