import { beforeEach, describe, expect, it, vi } from 'vitest';
import { render, screen } from '@testing-library/react';
import { MemoryRouter, Route, Routes } from 'react-router-dom';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { HelmetProvider } from '@dr.pogodin/react-helmet';
import { AxiosError } from 'axios';
import { QueueTheatrePage } from './QueueTheatrePage';
import PersonalStatsPage from './PersonalStatsPage';
import { AdminDiscoveryListFormPage } from './admin/AdminDiscoveryListFormPage';
import { discoveryListApi } from '../lib/discovery-list-api';
import { getUserStats } from '../lib/analytics-api';
import { getAdminNavGroups } from '../components/admin/adminNavigation';

vi.mock('../hooks/useQueue', () => ({
    useQueue: () => ({ data: { items: null }, isLoading: false, isError: false }),
    useRemoveFromQueue: () => ({ mutate: vi.fn() }),
    useReorderQueue: () => ({ mutate: vi.fn() }),
}));
vi.mock('../components/SEO', () => ({ SEO: () => null }));
vi.mock('../context/ToastContext', () => ({ useToast: () => ({ showToast: vi.fn() }) }));
vi.mock('../lib/discovery-list-api', () => ({ discoveryListApi: {
    getDiscoveryList: vi.fn(), getDiscoveryListClips: vi.fn(), admin: {},
} }));
vi.mock('../lib/analytics-api', () => ({ getUserStats: vi.fn() }));

function renderPage(page: React.ReactNode, path: string) {
    const client = new QueryClient({ defaultOptions: { queries: { retry: false } } });
    return render(<HelmetProvider><QueryClientProvider client={client}><MemoryRouter initialEntries={[path]}>
        <Routes><Route path={path} element={page} /></Routes>
    </MemoryRouter></QueryClientProvider></HelmetProvider>);
}

beforeEach(() => vi.clearAllMocks());

describe('authenticated route regressions', () => {
    it('renders the empty theatre state when the API returns null items', () => {
        renderPage(<QueueTheatrePage />, '/queue/theatre');
        expect(screen.getByText('Your queue is empty')).toBeInTheDocument();
    });

    it('opens the static new collection route without fetching an undefined list', () => {
        renderPage(<AdminDiscoveryListFormPage />, '/admin/discovery-lists/new');
        expect(screen.getByRole('heading', { name: 'Create Discovery List' })).toBeInTheDocument();
        expect(screen.getByRole('button', { name: 'Create List' })).toBeInTheDocument();
        expect(discoveryListApi.getDiscoveryList).not.toHaveBeenCalled();
        expect(discoveryListApi.getDiscoveryListClips).not.toHaveBeenCalled();
    });

    it.each([404, 500])('handles a statistics response with status %s', async status => {
        const error = new AxiosError('Statistics request failed');
        error.response = { status } as AxiosError['response'];
        vi.mocked(getUserStats).mockRejectedValue(error);
        renderPage(<PersonalStatsPage />, '/profile/stats');
        expect(await screen.findByText(status === 404 ? /No statistics yet/ : /Failed to load your statistics/)).toBeInTheDocument();
    });

    it('offers moderators review links without administrator-only destinations', () => {
        const paths = getAdminNavGroups(false).flatMap(group => group.items.map(item => item.href));
        expect(paths).toContain('/admin/moderation');
        expect(paths).toContain('/admin/submissions');
        for (const path of ['/admin/moderators']) {
            expect(paths).not.toContain(path);
            expect(getAdminNavGroups(true).flatMap(group => group.items.map(item => item.href))).toContain(path);
        }
    });
});
