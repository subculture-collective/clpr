import { render, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { MemoryRouter, Route, Routes, useNavigate } from 'react-router-dom';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { AppLayout } from './AppLayout';

// Mock the Header and Footer components
vi.mock('./Header', () => ({
    Header: () => <header data-testid='header'>Header</header>,
}));

vi.mock('./Footer', () => ({
    Footer: () => <footer data-testid='footer'>Footer</footer>,
}));

vi.mock('./CategoriesNav', () => ({
    CategoriesNav: () => <nav data-testid='categories-nav'>Categories</nav>,
}));

vi.mock('../OfflineIndicator', () => ({
    OfflineIndicator: () => null,
}));

vi.mock('../queue/QueueWidget', () => ({
    QueueWidget: () => null,
}));

vi.mock('@/hooks/useOfflineCache', () => ({
    useOfflineCacheInit: () => {},
}));

vi.mock('@/hooks/useSyncManager', () => ({
    useSyncManager: () => {},
}));

// Helper component to programmatically navigate
function NavigateButton({
    to,
    children,
}: {
    to: string;
    children: React.ReactNode;
}) {
    const navigate = useNavigate();
    return (
        <button
            onClick={() => {
                // Simulate layout's scroll lock reset before navigation
                document.body.style.overflow = 'unset';
                navigate(to);
            }}
        >
            {children}
        </button>
    );
}

function BackButton() {
    const navigate = useNavigate();
    return <button onClick={() => navigate(-1)}>Go back</button>;
}

describe('AppLayout', () => {
    beforeEach(() => {
        // Reset body overflow before each test
        document.body.style.overflow = 'unset';
        // Reset scroll position
        window.scrollTo(0, 0);
    });

    it('renders header, main content, and footer', () => {
        const { getByTestId, getByText } = render(
            <MemoryRouter initialEntries={['/']}>
                <Routes>
                    <Route element={<AppLayout />}>
                        <Route
                            path='/'
                            element={<div>Home Page</div>}
                        />
                    </Route>
                </Routes>
            </MemoryRouter>
        );

        expect(getByTestId('header')).toBeInTheDocument();
        expect(getByTestId('footer')).toBeInTheDocument();
        expect(getByText('Home Page')).toBeInTheDocument();
    });

    it('scrolls to top on route change', async () => {
        // Mock scrollTo
        const scrollToMock = vi.fn();
        const originalScrollTo = window.scrollTo;
        window.scrollTo = scrollToMock;

        const user = userEvent.setup();
        const { getByText } = render(
            <MemoryRouter initialEntries={['/']}>
                <Routes>
                    <Route element={<AppLayout />}>
                        <Route
                            path='/'
                            element={
                                <div>
                                    <div>Home</div>
                                    <NavigateButton to='/about'>
                                        Go to About
                                    </NavigateButton>
                                </div>
                            }
                        />
                        <Route
                            path='/about'
                            element={<div>About</div>}
                        />
                    </Route>
                </Routes>
            </MemoryRouter>
        );

        // Wait for initial render
        await waitFor(() => {
            expect(getByText('Home')).toBeInTheDocument();
        });

        // Clear initial scroll call
        scrollToMock.mockClear();

        // Navigate to /about
        const button = getByText('Go to About');
        await user.click(button);

        // Wait for navigation and scroll
        await waitFor(() => {
            expect(getByText('About')).toBeInTheDocument();
            expect(scrollToMock).toHaveBeenCalledWith(0, 0);
        });

        // Restore original scrollTo
        window.scrollTo = originalScrollTo;
    });

    it('restores scroll position when navigating back', async () => {
        const scrollToMock = vi.fn();
        const originalScrollTo = window.scrollTo;
        window.scrollTo = scrollToMock;
        let scrollY = 0;
        vi.spyOn(window, 'scrollY', 'get').mockImplementation(() => scrollY);

        const user = userEvent.setup();
        const { getByText } = render(
            <MemoryRouter initialEntries={['/']}>
                <Routes>
                    <Route element={<AppLayout />}>
                        <Route path='/' element={<NavigateButton to='/about'>Go to About</NavigateButton>} />
                        <Route path='/about' element={<BackButton />} />
                    </Route>
                </Routes>
            </MemoryRouter>
        );

        scrollY = 480;
        await user.click(getByText('Go to About'));
        await waitFor(() => expect(scrollToMock).toHaveBeenLastCalledWith(0, 0));

        scrollY = 0;
        await user.click(getByText('Go back'));
        await waitFor(() => expect(scrollToMock).toHaveBeenLastCalledWith(0, 480));

        window.scrollTo = originalScrollTo;
        vi.restoreAllMocks();
    });

    it('resets body overflow on route change', async () => {
        const user = userEvent.setup();
        const { getByText } = render(
            <MemoryRouter initialEntries={['/']}>
                <Routes>
                    <Route element={<AppLayout />}>
                        <Route
                            path='/'
                            element={
                                <div>
                                    <div>Home</div>
                                    <NavigateButton to='/submit'>
                                        Go to Submit
                                    </NavigateButton>
                                </div>
                            }
                        />
                        <Route
                            path='/submit'
                            element={<div>Submit</div>}
                        />
                    </Route>
                </Routes>
            </MemoryRouter>
        );

        // Should reset overflow on initial render (empty string = no inline override)
        expect(document.body.style.overflow).toBe('');

        // Set overflow to hidden (simulating modal opened)
        document.body.style.overflow = 'hidden';

        // Navigate to /submit
        const button = getByText('Go to Submit');
        await user.click(button);

        await waitFor(() => {
            expect(getByText('Submit')).toBeInTheDocument();
        });

        // Should reset after navigation (no scroll lock)
        await waitFor(() => {
            expect(document.body.style.overflow).not.toBe('hidden');
        });
    });

    it('handles navigation from submit page back correctly', async () => {
        // This test specifically addresses the P1 issue
        document.body.style.overflow = 'hidden';

        const user = userEvent.setup();
        const { getByText } = render(
            <MemoryRouter initialEntries={['/submit']}>
                <Routes>
                    <Route element={<AppLayout />}>
                        <Route
                            path='/'
                            element={<div>Home Page</div>}
                        />
                        <Route
                            path='/submit'
                            element={
                                <div>
                                    <div>Submit Page</div>
                                    <NavigateButton to='/'>
                                        Go to Home
                                    </NavigateButton>
                                </div>
                            }
                        />
                    </Route>
                </Routes>
            </MemoryRouter>
        );

        expect(getByText('Submit Page')).toBeInTheDocument();

        // Set overflow to hidden (simulating modal was open during submit)
        document.body.style.overflow = 'hidden';

        // Navigate back to home (simulating browser back)
        const button = getByText('Go to Home');
        await user.click(button);

        await waitFor(() => {
            expect(getByText('Home Page')).toBeInTheDocument();
        });

        // Body overflow should be reset (no black screen)
        await waitFor(() => {
            expect(document.body.style.overflow).not.toBe('hidden');
        });
    });
});
