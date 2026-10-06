import { render, screen } from '@testing-library/react';
import { MemoryRouter, Route, Routes } from 'react-router-dom';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { TagPage } from './TagPage';
import type { Tag } from '../types/tag';

vi.mock('../components', () => ({
    Container: ({ children }: { children: React.ReactNode }) => (
        <div data-testid="container">{children}</div>
    ),
    SEO: () => null,
}));

vi.mock('../components/clip', () => ({
    ClipFeed: ({ title, filters }: { title?: string; filters?: { tags?: string[] } }) => (
        <div data-testid="clip-feed">
            <h2>{title}</h2>
            <div data-testid="filters">{JSON.stringify(filters)}</div>
        </div>
    ),
}));

const useTagMock = vi.fn();
vi.mock('../hooks/useTags', () => ({ useTag: (slug: string) => useTagMock(slug) }));

function renderAt(path: string, route = '/tags/:tagSlug') {
    return render(
        <MemoryRouter initialEntries={[path]}>
            <Routes>
                <Route path={route} element={<TagPage />} />
            </Routes>
        </MemoryRouter>,
    );
}

const facecam: Tag = {
    id: 't1',
    name: 'Content: facecam',
    slug: 'content/facecam',
    usage_count: 435,
    created_at: '',
    lane: 'detected',
    display_name: 'Facecam',
    evidence: 'visible',
};

describe('TagPage', () => {
    beforeEach(() => useTagMock.mockReset().mockReturnValue({ data: undefined }));

    it('decodes namespaced slugs from the canonical route and filters the feed', () => {
        renderAt('/tags/content%2Fhighlights');
        expect(useTagMock).toHaveBeenCalledWith('content/highlights');
        expect(screen.getByTestId('filters').textContent).toContain('"tags":["content/highlights"]');
    });

    it('shows the tag name and clip count without provenance', () => {
        useTagMock.mockReturnValue({ data: { tag: facecam, clip_count: 1234 } });
        renderAt('/tags/content%2Ffacecam');
        expect(screen.getByRole('heading', { level: 1, name: 'Facecam' })).toBeInTheDocument();
        expect(screen.queryByText('What clpr saw')).not.toBeInTheDocument();
        expect(screen.queryByText(/Seen tag/)).not.toBeInTheDocument();
        expect(screen.getByText('1,234')).toBeInTheDocument();
    });

    it('falls back to the slug while details are unavailable', () => {
        renderAt('/tags/funny');
        expect(screen.getByRole('heading', { level: 1, name: 'funny' })).toBeInTheDocument();
        expect(screen.getByTestId('filters').textContent).toContain('"tags":["funny"]');
    });

    it('handles missing tagSlug gracefully', () => {
        renderAt('/tags/', '/tags/:tagSlug?');
        expect(screen.getByText('No tag specified')).toBeInTheDocument();
        expect(screen.queryByTestId('clip-feed')).not.toBeInTheDocument();
    });
});
