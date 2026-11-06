import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { render, screen } from '@testing-library/react';
import { MemoryRouter, Route, Routes } from 'react-router-dom';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import * as clipApi from '../lib/clip-api';
import type { Clip } from '../types/clip';
import { CreatorPage } from './CreatorPage';

vi.mock('../lib/clip-api');
vi.mock('../components/SEO', () => ({ SEO: () => null }));
vi.mock('../components/clip/ClipCard', () => ({
    ClipCard: ({ clip }: { clip: Clip }) => <div>{clip.title}</div>,
}));

describe('CreatorPage', () => {
    beforeEach(() => vi.clearAllMocks());

    it('lists clips from the creator clips endpoint', async () => {
        vi.mocked(clipApi.fetchCreatorClips).mockResolvedValue({
            clips: [{ id: 'c1', title: 'Creator moment' } as Clip],
            page: 1,
            total: 1,
            limit: 25,
            has_more: false,
        });

        render(
            <QueryClientProvider client={new QueryClient({ defaultOptions: { queries: { retry: false } } })}>
                <MemoryRouter initialEntries={['/creator/12345']}>
                    <Routes>
                        <Route path='/creator/:creatorId' element={<CreatorPage />} />
                    </Routes>
                </MemoryRouter>
            </QueryClientProvider>,
        );

        expect(await screen.findByText('Creator moment')).toBeInTheDocument();
        expect(screen.getByText('Viewing clips from: 12345')).toBeInTheDocument();
        expect(clipApi.fetchCreatorClips).toHaveBeenCalledWith({ creatorId: '12345', pageParam: 1 });
        expect(clipApi.fetchClips).not.toHaveBeenCalled();
    });

    it('handles missing creatorId gracefully', () => {
        render(
            <QueryClientProvider client={new QueryClient()}>
                <MemoryRouter initialEntries={['/creator/']}>
                    <Routes>
                        <Route path='/creator/:creatorId?' element={<CreatorPage />} />
                    </Routes>
                </MemoryRouter>
            </QueryClientProvider>,
        );

        expect(screen.getByText('Creator Not Found')).toBeInTheDocument();
        expect(screen.getByText('No creator ID specified.')).toBeInTheDocument();
        expect(clipApi.fetchCreatorClips).not.toHaveBeenCalled();
    });
});
