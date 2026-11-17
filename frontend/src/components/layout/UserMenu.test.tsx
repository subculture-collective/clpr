import { act, fireEvent, render, screen } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';
import { describe, expect, it, vi } from 'vitest';
import { UserMenu } from './UserMenu';

const auth = vi.hoisted(() => ({
    user: null as null | Record<string, unknown>,
    logout: vi.fn(),
    isAdmin: false,
    isModeratorOrAdmin: false,
}));

vi.mock('../../context/AuthContext', () => ({ useAuth: () => auth }));

const baseUser = {
    id: 'u1',
    username: 'streamer_one',
    display_name: 'Streamer One',
    role: 'user',
    karma_points: 0,
    is_banned: false,
    created_at: '2026-01-01T00:00:00Z',
};

async function openMenu() {
    render(
        <MemoryRouter>
            <UserMenu />
        </MemoryRouter>,
    );
    // The menu resets its keyboard focus in a microtask after opening.
    await act(async () => {
        fireEvent.click(screen.getByRole('button', { name: 'User menu' }));
    });
}

describe('UserMenu clip room link', () => {
    it('links a Twitch-connected account to its own clip room', async () => {
        auth.user = { ...baseUser, twitch_id: '123' };
        await openMenu();

        expect(screen.getByRole('menuitem', { name: /Stream Clip Room/ })).toHaveAttribute(
            'href',
            '/streamer-tools/streamer_one/clips',
        );
        expect(screen.getByRole('menuitem', { name: /Creator Dashboard/ })).toHaveAttribute('href', '/creator/123/dashboard');
    });

    it('hides the link for accounts without a Twitch login', async () => {
        auth.user = { ...baseUser, twitch_id: '' };
        await openMenu();

        expect(screen.queryByRole('menuitem', { name: /Stream Clip Room/ })).not.toBeInTheDocument();
        expect(screen.queryByRole('menuitem', { name: /Creator Dashboard/ })).not.toBeInTheDocument();
    });
});

describe('UserMenu staff link', () => {
    it('sends a moderator who is not an admin to the review queue', async () => {
        auth.user = { ...baseUser, role: 'moderator' };
        Object.assign(auth, { isAdmin: false, isModeratorOrAdmin: true });
        try {
            await openMenu();
            expect(screen.getByRole('menuitem', { name: /Moderation/ })).toHaveAttribute('href', '/admin/moderation');
        } finally {
            Object.assign(auth, { isAdmin: false, isModeratorOrAdmin: false });
        }
    });

    it('sends an admin to the admin dashboard', async () => {
        auth.user = { ...baseUser, role: 'admin' };
        Object.assign(auth, { isAdmin: true, isModeratorOrAdmin: true });
        try {
            await openMenu();
            expect(screen.getByRole('menuitem', { name: /Admin Panel/ })).toHaveAttribute('href', '/admin/dashboard');
        } finally {
            Object.assign(auth, { isAdmin: false, isModeratorOrAdmin: false });
        }
    });
});
