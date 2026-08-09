import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { BrowserRouter } from 'react-router-dom';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { PaywallModal } from './PaywallModal';
import * as subscriptionApi from '../../lib/subscription-api';

// Mock the subscription API
vi.mock('../../lib/subscription-api', () => ({
    createCheckoutSession: vi.fn(),
}));

// Mock useAuth hook
vi.mock('../../hooks/useAuth', () => ({
    useAuth: () => ({
        user: { id: 'test-user-id', username: 'testuser' },
        isAuthenticated: true,
    }),
}));

const createWrapper = () => {
    const queryClient = new QueryClient({
        defaultOptions: {
            queries: { retry: false },
            mutations: { retry: false },
        },
    });

    return ({ children }: { children: React.ReactNode }) => (
        <QueryClientProvider client={queryClient}>
            <BrowserRouter>{children}</BrowserRouter>
        </QueryClientProvider>
    );
};

describe('PaywallModal', () => {
    const mockOnClose = vi.fn();
    const mockOnUpgradeClick = vi.fn();
    const mockCheckoutRedirect = vi.fn();

    beforeEach(() => {
        vi.clearAllMocks();
        // Set environment variables for price IDs using Vitest's stubEnv
        vi.stubEnv('VITE_STRIPE_PRO_MONTHLY_PRICE_ID', 'price_monthly_test');
        vi.stubEnv('VITE_STRIPE_PRO_YEARLY_PRICE_ID', 'price_yearly_test');
        // Mock window.alert and window.location
        vi.stubGlobal('alert', vi.fn());
        Object.defineProperty(window, 'location', {
            value: { href: '' },
            writable: true,
        });
    });

    afterEach(() => {
        vi.unstubAllEnvs();
        vi.unstubAllGlobals();
    });

    it('should not render when isOpen is false', () => {
        const { container } = render(
            <PaywallModal isOpen={false} onClose={mockOnClose} />,
            { wrapper: createWrapper() }
        );

        expect(container).toBeEmptyDOMElement();
    });

    it('should render modal when isOpen is true', () => {
        render(
            <PaywallModal
                isOpen={true}
                onClose={mockOnClose}
                priceIds={{ yearly: 'price_yearly_test' }}
            />,
            { wrapper: createWrapper() },
        );

        expect(
            screen.getByText(/This feature is a Pro Feature/i)
        ).toBeInTheDocument();
    });

    it('should display custom title and description', () => {
        render(
            <PaywallModal
                isOpen={true}
                onClose={mockOnClose}
                title='Custom Title'
                description='Custom description text'
            />,
            { wrapper: createWrapper() }
        );

        expect(screen.getByText('Custom Title')).toBeInTheDocument();
        expect(screen.getByText('Custom description text')).toBeInTheDocument();
    });

    it('should display feature name in title when provided', () => {
        render(
            <PaywallModal
                isOpen={true}
                onClose={mockOnClose}
                featureName='Collections'
            />,
            { wrapper: createWrapper() }
        );

        expect(
            screen.getByText(/Collections is a Pro Feature/i)
        ).toBeInTheDocument();
    });

    it('should close modal when close button is clicked', async () => {
        const user = userEvent.setup();

        render(<PaywallModal isOpen={true} onClose={mockOnClose} />, {
            wrapper: createWrapper(),
        });

        const closeButton = screen.getByLabelText('Close modal');
        await user.click(closeButton);

        expect(mockOnClose).toHaveBeenCalledTimes(1);
    });

    it('should toggle between monthly and yearly billing periods', async () => {
        const user = userEvent.setup();

        render(<PaywallModal isOpen={true} onClose={mockOnClose} />, {
            wrapper: createWrapper(),
        });

        const monthlyButton = screen.getByRole('button', { name: /Monthly/i });
        const yearlyButton = screen.getByRole('button', { name: /Yearly/i });

        // Should default to yearly
        expect(yearlyButton).toHaveClass('bg-purple-600');

        // Switch to monthly
        await user.click(monthlyButton);
        await waitFor(() => {
            expect(monthlyButton).toHaveClass('bg-purple-600');
        });
    });

    it('should display monthly pricing when monthly is selected', async () => {
        const user = userEvent.setup();

        render(<PaywallModal isOpen={true} onClose={mockOnClose} />, {
            wrapper: createWrapper(),
        });

        const monthlyButton = screen.getByRole('button', { name: /Monthly/i });
        await user.click(monthlyButton);

        await waitFor(() => {
            expect(screen.getByText('$9.99')).toBeInTheDocument();
        });
    });

    it('should display correct pricing for yearly plan', () => {
        render(<PaywallModal isOpen={true} onClose={mockOnClose} />, {
            wrapper: createWrapper(),
        });

        expect(screen.getByText('$8.33')).toBeInTheDocument();
        expect(screen.getByText('Billed $99.99/year')).toBeInTheDocument();
    });

    it('should display all Pro features', () => {
        render(<PaywallModal isOpen={true} onClose={mockOnClose} />, {
            wrapper: createWrapper(),
        });

        expect(screen.getByText('Ad-free browsing')).toBeInTheDocument();
        expect(screen.getByText('Unlimited favorites')).toBeInTheDocument();
        expect(screen.getByText('Custom collections')).toBeInTheDocument();
        expect(
            screen.getByText('Advanced search & filters')
        ).toBeInTheDocument();
        expect(screen.getByText('Cross-device sync')).toBeInTheDocument();
        expect(screen.getByText('Export your data')).toBeInTheDocument();
        expect(screen.getByText('5x higher rate limits')).toBeInTheDocument();
        expect(screen.getByText('Priority support')).toBeInTheDocument();
    });

    it('should call analytics when upgrade button is clicked', async () => {
        const user = userEvent.setup();

        vi.mocked(subscriptionApi.createCheckoutSession).mockResolvedValue({
            session_id: 'test-session-id',
            session_url: 'https://checkout.stripe.com/test',
        });

        render(
            <PaywallModal
                isOpen={true}
                onClose={mockOnClose}
                onUpgradeClick={mockOnUpgradeClick}
                onCheckoutRedirect={mockCheckoutRedirect}
            />,
            { wrapper: createWrapper() }
        );

        const upgradeButton = screen.getByRole('button', {
            name: /Upgrade to Pro/i,
        });
        await user.click(upgradeButton);

        await waitFor(() => {
            expect(mockOnUpgradeClick).toHaveBeenCalledTimes(1);
        });
    });

    it('should show loading state when processing upgrade', async () => {
        const user = userEvent.setup();

        // Create a promise that we can control
        let resolveCheckout!: (value: {
            session_id: string;
            session_url: string;
        }) => void;
        const controlledPromise = new Promise<{
            session_id: string;
            session_url: string;
        }>(resolve => {
            resolveCheckout = resolve;
        });

        vi.mocked(subscriptionApi.createCheckoutSession).mockReturnValue(
            controlledPromise
        );

        render(
            <PaywallModal
                isOpen={true}
                onClose={mockOnClose}
                priceIds={{ yearly: 'price_yearly_test' }}
                onCheckoutRedirect={mockCheckoutRedirect}
            />,
            { wrapper: createWrapper() },
        );

        const upgradeButton = screen.getByRole('button', {
            name: /^Upgrade to Pro$/i,
        });

        // Click the button
        await user.click(upgradeButton);

        // The button should immediately show Processing... since setIsLoading is called before await
        expect(await screen.findByText('Processing...')).toBeInTheDocument();

        // The button should be disabled
        expect(upgradeButton).toBeDisabled();

        resolveCheckout({
            session_id: 'test-id',
            session_url: 'https://test.com',
        });
        await waitFor(() =>
            expect(subscriptionApi.createCheckoutSession).toHaveBeenCalled(),
        );
    });

    it('should display savings percentage for yearly plan', () => {
        render(<PaywallModal isOpen={true} onClose={mockOnClose} />, {
            wrapper: createWrapper(),
        });

        expect(screen.getByText(/Save 17%/i)).toBeInTheDocument();
    });
});
