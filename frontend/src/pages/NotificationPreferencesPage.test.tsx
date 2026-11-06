import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { render, screen } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';
import { describe, expect, it, vi } from 'vitest';
import { NotificationPreferencesPage, ToggleSwitch } from './NotificationPreferencesPage';

vi.mock('@dr.pogodin/react-helmet', () => ({ Helmet: () => null }));
vi.mock('../lib/notification-api', () => ({
    getNotificationPreferences: vi.fn(async () => ({ notify_broadcaster_live: true, notify_stream_live: true })),
    updateNotificationPreferences: vi.fn(),
    resetNotificationPreferences: vi.fn(),
}));

describe('notification preference switch', () => {
    it('exposes its visible label and checked state to assistive technology', () => {
        render(
            <ToggleSwitch
                label='New device login'
                description='Notify me after a new device signs in.'
                checked
                onChange={vi.fn()}
            />,
        );

        expect(
            screen.getByRole('switch', { name: 'New device login' }),
        ).toHaveAttribute('aria-checked', 'true');
    });

    it('no longer offers the retired Twitch stream-live notification', async () => {
        render(
            <QueryClientProvider client={new QueryClient({ defaultOptions: { queries: { retry: false } } })}>
                <MemoryRouter>
                    <NotificationPreferencesPage />
                </MemoryRouter>
            </QueryClientProvider>,
        );

        expect(await screen.findByRole('switch', { name: 'Creator Live on Clpr' })).toBeInTheDocument();
        expect(screen.queryByRole('switch', { name: 'Stream Live' })).not.toBeInTheDocument();
    });
});
