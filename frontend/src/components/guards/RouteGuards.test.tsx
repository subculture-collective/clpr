import { beforeEach, describe, expect, it, vi } from 'vitest';
import { render, screen } from '@testing-library/react';
import { MemoryRouter, Route, Routes, useLocation } from 'react-router-dom';
import { AdminRoute } from './AdminRoute';
import { ProtectedRoute } from './ProtectedRoute';
import { GuestRoute } from './GuestRoute';

const auth = vi.hoisted(() => ({ isAuthenticated: false, isLoading: false, isModeratorOrAdmin: false }));
vi.mock('../../context/AuthContext', () => ({ useAuth: () => auth }));

function LoginDestination() {
    const location = useLocation();
    return <p>Login return: {location.state?.from?.pathname}{location.state?.from?.search}</p>;
}
function renderGuard(Guard: typeof ProtectedRoute) {
    return render(
        <MemoryRouter initialEntries={['/private?tab=history']}>
            <Routes>
                <Route path='/private' element={<Guard><h1>Private content</h1></Guard>} />
                <Route path='/login' element={<LoginDestination />} />
                <Route path='/' element={<h1>Home</h1>} />
            </Routes>
        </MemoryRouter>,
    );
}

beforeEach(() => Object.assign(auth, { isAuthenticated: false, isLoading: false, isModeratorOrAdmin: false }));

describe('route access decisions', () => {
    it.each([ProtectedRoute, AdminRoute, GuestRoute])('%s does not expose content or redirect before identity resolves', Guard => {
        auth.isLoading = true;
        renderGuard(Guard);
        expect(screen.queryByRole('heading')).not.toBeInTheDocument();
        expect(screen.queryByText(/Login return:/)).not.toBeInTheDocument();
    });

    it('preserves the account destination and query when asking the visitor to sign in', () => {
        renderGuard(ProtectedRoute);
        expect(screen.getByText('Login return: /private?tab=history')).toBeInTheDocument();
    });

    it('allows an authenticated account into its protected route', () => {
        auth.isAuthenticated = true;
        renderGuard(ProtectedRoute);
        expect(screen.getByRole('heading', { name: 'Private content' })).toBeInTheDocument();
    });

    it('redirects anonymous administration requests to sign-in', () => {
        renderGuard(AdminRoute);
        expect(screen.getByText('Login return:')).toBeInTheDocument();
        expect(screen.queryByText('Private content')).not.toBeInTheDocument();
    });

    it('denies an authenticated account without moderation permission', () => {
        auth.isAuthenticated = true;
        renderGuard(AdminRoute);
        expect(screen.getByRole('heading', { name: /403/ })).toBeInTheDocument();
        expect(screen.queryByText('Private content')).not.toBeInTheDocument();
    });

    it('allows an authenticated moderator or administrator', () => {
        auth.isAuthenticated = true;
        auth.isModeratorOrAdmin = true;
        renderGuard(AdminRoute);
        expect(screen.getByRole('heading', { name: 'Private content' })).toBeInTheDocument();
    });

    it('returns an already authenticated visitor from guest-only pages to home', () => {
        auth.isAuthenticated = true;
        renderGuard(GuestRoute);
        expect(screen.getByRole('heading', { name: 'Home' })).toBeInTheDocument();
    });

    it('allows an anonymous visitor into a guest-only page', () => {
        renderGuard(GuestRoute);
        expect(screen.getByRole('heading', { name: 'Private content' })).toBeInTheDocument();
    });
});
