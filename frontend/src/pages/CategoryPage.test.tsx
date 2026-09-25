import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { MemoryRouter, Route, Routes } from 'react-router-dom';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { allowTestConsole } from '../test/setup';

const { getCategory, getCategoryGames, getCategoryClips } = vi.hoisted(() => ({
    getCategory: vi.fn(),
    getCategoryGames: vi.fn(),
    getCategoryClips: vi.fn(),
}));

vi.mock('../lib/category-api', () => ({
    categoryApi: { getCategory, getCategoryGames, getCategoryClips },
}));
vi.mock('../components/SEO', () => ({ SEO: () => null, default: () => null }));
vi.mock('../components/clip', () => ({
    ClipGridCard: ({ clip }: { clip: { title: string } }) => <article>{clip.title}</article>,
}));

import { CategoryPage } from './CategoryPage';

function axiosError(status: number) {
    return Object.assign(new Error(`Request failed with status code ${status}`), {
        isAxiosError: true,
        response: { status, data: { error: 'Category not found' } },
    });
}

function renderTopic(path: string) {
    return render(
        <MemoryRouter initialEntries={[path]}>
            <Routes>
                <Route path='/topics/:categorySlug' element={<CategoryPage />} />
            </Routes>
        </MemoryRouter>,
    );
}

const category = {
    id: 't1',
    name: 'News & Politics',
    slug: 'news-politics',
    position: 1,
    created_at: '',
    updated_at: '',
};

describe('CategoryPage', () => {
    beforeEach(() => {
        getCategory.mockReset();
        getCategoryGames.mockReset();
        getCategoryClips.mockReset();
    });

    it('shows a not-found state for an unknown topic', async () => {
        getCategory.mockRejectedValue(axiosError(404));
        getCategoryGames.mockRejectedValue(axiosError(404));
        getCategoryClips.mockRejectedValue(axiosError(404));
        renderTopic('/topics/does-not-exist');

        expect(await screen.findByRole('heading', { level: 1, name: "This topic isn't here" })).toBeInTheDocument();
        expect(screen.getByRole('link', { name: 'Browse topics' })).toHaveAttribute('href', '/topics');
        expect(screen.queryByText(/Failed to load/)).not.toBeInTheDocument();
    });

    it('offers a retry for other failures', async () => {
        allowTestConsole('error', /Failed to fetch category data:.*500/);
        getCategory.mockRejectedValueOnce(axiosError(500)).mockResolvedValue({ category });
        getCategoryGames.mockResolvedValue({ games: [] });
        getCategoryClips.mockResolvedValue({ clips: [], has_more: false });
        renderTopic('/topics/news-politics');

        expect(await screen.findByRole('alert')).toHaveTextContent("We couldn't load this topic");
        await userEvent.click(screen.getByRole('button', { name: 'Try again' }));
        expect(await screen.findByRole('heading', { level: 1, name: 'News & Politics' })).toBeInTheDocument();
    });

    it('refetches only the clips when the sort changes', async () => {
        getCategory.mockResolvedValue({ category });
        getCategoryGames.mockResolvedValue({ games: [] });
        getCategoryClips.mockResolvedValue({ clips: [{ id: 'c1', title: 'Clip one' }], has_more: false });
        renderTopic('/topics/news-politics');

        expect(await screen.findByText('Clip one')).toBeInTheDocument();
        await userEvent.selectOptions(screen.getAllByRole('combobox')[0], 'new');
        await screen.findByText('Clip one');

        expect(getCategoryClips).toHaveBeenCalledTimes(2);
        expect(getCategory).toHaveBeenCalledTimes(1);
        expect(getCategoryGames).toHaveBeenCalledTimes(1);
    });
});
