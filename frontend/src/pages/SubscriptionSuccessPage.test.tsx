import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { render, screen, waitFor } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import SubscriptionSuccessPage from './SubscriptionSuccessPage';
import { getSubscription } from '../lib/subscription-api';
import { trackConversion } from '../lib/paywall-analytics';

vi.mock('../hooks/useAuth', () => ({
    useAuth: () => ({ user: { id: 'user-1' } }),
}));
vi.mock('../lib/paywall-analytics', () => ({ trackConversion: vi.fn() }));
vi.mock('../lib/subscription-api', () => ({
    getSubscription: vi.fn(),
    isProUser: (subscription: { tier?: string; status?: string } | null) =>
        subscription?.tier === 'pro' &&
        ['active', 'trialing'].includes(subscription.status ?? ''),
}));

function renderPage() {
    const queryClient = new QueryClient({
        defaultOptions: { queries: { retry: false } },
    });
    return render(
        <QueryClientProvider client={queryClient}>
            <MemoryRouter initialEntries={['/subscription/success?session_id=cs_test']}>
                <SubscriptionSuccessPage />
            </MemoryRouter>
        </QueryClientProvider>,
    );
}

describe('SubscriptionSuccessPage', () => {
    beforeEach(() => vi.clearAllMocks());

    it('does not claim or track Pro before backend entitlement exists', async () => {
        vi.mocked(getSubscription).mockResolvedValue(null);
        renderPage();

        expect(
            await screen.findByRole('heading', {
                name: 'Confirming your subscription',
            }),
        ).toBeVisible();
        expect(screen.queryByText('Welcome to clpr Pro!')).not.toBeInTheDocument();
        expect(trackConversion).not.toHaveBeenCalled();
    });

    it('shows success and tracks conversion after backend entitlement confirmation', async () => {
        vi.mocked(getSubscription).mockResolvedValue({
            id: 'sub-1',
            user_id: 'user-1',
            stripe_customer_id: 'cus-1',
            status: 'active',
            tier: 'pro',
            cancel_at_period_end: false,
            created_at: new Date().toISOString(),
            updated_at: new Date().toISOString(),
        });
        renderPage();

        expect(await screen.findByText('Welcome to clpr Pro!')).toBeVisible();
        await waitFor(() => expect(trackConversion).toHaveBeenCalledTimes(1));
    });
});
