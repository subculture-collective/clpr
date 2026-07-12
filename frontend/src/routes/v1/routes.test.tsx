import { render, screen } from '@testing-library/react';
import { MemoryRouter, Routes } from 'react-router-dom';
import { describe, expect, it, vi } from 'vitest';
import { accountRoutes, type AccountRoutePages } from './AccountRoutes';
import { adminRoutes, type AdminRoutePages } from './AdminRoutes';

vi.mock('@/components/guards/AdminRoute', () => ({
    AdminRoute: ({ children }: { children: React.ReactNode }) => (
        <div data-testid='admin-boundary'>{children}</div>
    ),
}));

vi.mock('@/components/guards/ProtectedRoute', () => ({
    ProtectedRoute: ({ children }: { children: React.ReactNode }) => (
        <div data-testid='account-boundary'>{children}</div>
    ),
}));

function Page() {
    return <h1>Matched page</h1>;
}

const adminPages = new Proxy({}, { get: () => Page }) as AdminRoutePages;
const accountPages = new Proxy({}, { get: () => Page }) as AccountRoutePages;

describe('versioned route boundaries', () => {
    it('wraps administration paths in the admin boundary', () => {
        render(
            <MemoryRouter initialEntries={['/admin/moderation/analytics']}>
                <Routes>{adminRoutes(adminPages)}</Routes>
            </MemoryRouter>,
        );

        expect(screen.getByTestId('admin-boundary')).toContainElement(
            screen.getByRole('heading', { name: 'Matched page' }),
        );
    });

    it('wraps signed-in account paths in the account boundary', () => {
        render(
            <MemoryRouter initialEntries={['/settings']}>
                <Routes>{accountRoutes(accountPages)}</Routes>
            </MemoryRouter>,
        );

        expect(screen.getByTestId('account-boundary')).toContainElement(
            screen.getByRole('heading', { name: 'Matched page' }),
        );
    });

    it('keeps explicitly public playlist paths outside the account boundary', () => {
        render(
            <MemoryRouter initialEntries={['/playlists/discover']}>
                <Routes>{accountRoutes(accountPages)}</Routes>
            </MemoryRouter>,
        );

        expect(screen.queryByTestId('account-boundary')).not.toBeInTheDocument();
        expect(screen.getByRole('heading', { name: 'Matched page' })).toBeInTheDocument();
    });
});
