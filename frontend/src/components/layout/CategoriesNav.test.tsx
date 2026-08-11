import { render, screen } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';
import { describe, expect, it, vi } from 'vitest';
import { CategoriesNav } from './CategoriesNav';

vi.mock('../../lib/category-api', () => ({
    categoryApi: {
        listCategories: vi.fn().mockResolvedValue({
            categories: [
                {
                    id: 'topic-1',
                    name: 'IRL',
                    slug: 'irl',
                    position: 1,
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
        render(
            <MemoryRouter>
                <CategoriesNav />
            </MemoryRouter>,
        );

        expect(await screen.findByRole('button', { name: 'Creators' })).toBeInTheDocument();
        expect(screen.getByRole('button', { name: 'Topics' })).toBeInTheDocument();
        expect(screen.getByRole('link', { name: /Creator One/ })).toHaveAttribute(
            'href',
            '/broadcaster/creator-1',
        );
        expect(screen.queryByText('Streamers')).not.toBeInTheDocument();
        expect(screen.queryByText('Games')).not.toBeInTheDocument();
    });
});
