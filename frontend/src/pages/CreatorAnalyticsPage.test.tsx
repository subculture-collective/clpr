import { render, screen } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { MemoryRouter, Route, Routes } from 'react-router-dom';
import { beforeEach, describe, expect, it, vi } from 'vitest';

const { getCreatorAnalyticsOverview, getCreatorTopClips, getCreatorTrends } = vi.hoisted(() => ({
    getCreatorAnalyticsOverview: vi.fn(),
    getCreatorTopClips: vi.fn(),
    getCreatorTrends: vi.fn(),
}));

vi.mock('../lib/analytics-api', () => ({
    getCreatorAnalyticsOverview,
    getCreatorTopClips,
    getCreatorTrends,
}));
vi.mock('@dr.pogodin/react-helmet', () => ({ Helmet: () => null }));
vi.mock('../components/analytics', async () => {
    const actual = await vi.importActual<typeof import('../components/analytics')>('../components/analytics');
    return {
        ...actual,
        LineChartComponent: ({ title }: { title: string }) => <div>{title}</div>,
        AudienceInsightsSection: () => null,
    };
});

import CreatorAnalyticsPage from './CreatorAnalyticsPage';

function notFound() {
    return Object.assign(new Error('Request failed with status code 404'), {
        isAxiosError: true,
        response: { status: 404, data: { error: 'creator analytics not found' } },
    });
}

function renderAnalytics(name = 'Twitch') {
    const client = new QueryClient({ defaultOptions: { queries: { retry: false } } });
    return render(
        <QueryClientProvider client={client}>
            <MemoryRouter initialEntries={[`/creator/${name}/analytics`]}>
                <Routes>
                    <Route path='/creator/:creatorName/analytics' element={<CreatorAnalyticsPage />} />
                </Routes>
            </MemoryRouter>
        </QueryClientProvider>,
    );
}

const topClip = {
    id: 'c1',
    title: 'Patch Notes',
    views: 18713,
    vote_score: 0,
    comment_count: 0,
    engagement_rate: 0,
};

describe('CreatorAnalyticsPage', () => {
    beforeEach(() => {
        getCreatorAnalyticsOverview.mockReset();
        getCreatorTopClips.mockReset();
        getCreatorTrends.mockReset();
        getCreatorTrends.mockResolvedValue({ data: null, days: 30, metric: 'clip_views' });
    });

    it('explains a missing summary instead of failing when clips are tracked', async () => {
        getCreatorAnalyticsOverview.mockRejectedValue(notFound());
        getCreatorTopClips.mockResolvedValue({ clips: [topClip] });
        renderAnalytics();

        expect(await screen.findByText('Summary not available yet')).toBeInTheDocument();
        expect(await screen.findByRole('link', { name: 'Patch Notes' })).toBeInTheDocument();
        expect(screen.getByText('No view history recorded for this period yet.')).toBeInTheDocument();
        expect(screen.queryByRole('alert')).not.toBeInTheDocument();
    });

    it('shows a friendly page when nothing has been tracked for the creator', async () => {
        getCreatorAnalyticsOverview.mockRejectedValue(notFound());
        getCreatorTopClips.mockResolvedValue({ clips: [] });
        renderAnalytics('nobody');

        expect(await screen.findByRole('heading', { level: 1, name: 'No analytics for nobody yet' })).toBeInTheDocument();
        expect(screen.getByRole('link', { name: 'Browse creators' })).toHaveAttribute('href', '/creators');
        expect(screen.queryByText(/status code/)).not.toBeInTheDocument();
    });

    it('lets the range buttons wrap instead of forcing one row', async () => {
        getCreatorAnalyticsOverview.mockRejectedValue(notFound());
        getCreatorTopClips.mockResolvedValue({ clips: [topClip] });
        renderAnalytics();

        const group = await screen.findByRole('group', { name: 'Date range selection' });
        expect(group.className).toContain('flex-wrap');
        expect(group.parentElement?.className).toContain('flex-wrap');
    });
});
