import { act, render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { MemoryRouter } from 'react-router-dom';
import { beforeEach, expect, it, vi } from 'vitest';
import { SettingsPage } from './SettingsPage';
import { getUserSettings, updateUserSettings } from '@/lib/user-settings-api';

const auth = vi.hoisted(() => ({ user: { id: 'member', username: 'member', display_name: 'Member', bio: '' }, refreshUser: vi.fn() }));
vi.mock('@/context/AuthContext', () => ({ useAuth: () => auth }));
vi.mock('@/context/ConsentContext', () => ({ useConsent: () => ({ consent: { functional: false, analytics: false, advertising: false }, updateConsent: vi.fn(), resetConsent: vi.fn(), doNotTrack: false }) }));
vi.mock('@dr.pogodin/react-helmet', () => ({ Helmet: () => null }));
vi.mock('@/lib/user-settings-api', () => ({ getUserSettings: vi.fn(), updateProfile: vi.fn(), updateUserSettings: vi.fn() }));
const saved = { user_id: 'member', profile_visibility: 'public' as const, show_karma_publicly: true, created_at: '', updated_at: '' };
function setup() {
    const client = new QueryClient({ defaultOptions: { queries: { retry: false } } });
    const view = () => <QueryClientProvider client={client}><MemoryRouter><SettingsPage /></MemoryRouter></QueryClientProvider>;
    const rendered = render(view());
    return { client, refresh: () => rendered.rerender(view()) };
}
beforeEach(() => {
    vi.clearAllMocks();
    auth.user = { id: 'member', username: 'member', display_name: 'Member', bio: '' };
    vi.mocked(getUserSettings).mockResolvedValue(saved);
    vi.mocked(updateUserSettings).mockResolvedValue(undefined);
});
it('preserves entered profile and privacy values when fresh server data arrives', async () => {
    const user = userEvent.setup();
    const { client, refresh } = setup();
    const visibility = await screen.findByLabelText('Profile Visibility');
    await user.selectOptions(visibility, 'private');
    const display = screen.getByLabelText('Display Name');
    await user.clear(display);
    await user.type(display, 'Unsaved name');
    await act(async () => { client.setQueryData(['userSettings', 'member'], { ...saved, show_karma_publicly: false }); });
    auth.user = { ...auth.user, display_name: 'Fresh server name' };
    refresh();
    expect(visibility).toHaveValue('private');
    expect(display).toHaveValue('Unsaved name');
    await user.click(screen.getByRole('button', { name: 'Save Settings' }));
    await waitFor(() => expect(updateUserSettings).toHaveBeenCalledWith({ profile_visibility: 'private', show_karma_publicly: true }));
});
it('does not offer invented privacy defaults after a failed load and recovers on retry', async () => {
    vi.mocked(getUserSettings).mockRejectedValueOnce(new Error('Unavailable'));
    setup();
    expect(await screen.findByRole('alert')).toHaveTextContent('Privacy settings could not be loaded');
    expect(screen.queryByLabelText('Profile Visibility')).not.toBeInTheDocument();
    await userEvent.setup().click(screen.getByRole('button', { name: 'Try again' }));
    expect(await screen.findByLabelText('Profile Visibility')).toHaveValue('public');
});
