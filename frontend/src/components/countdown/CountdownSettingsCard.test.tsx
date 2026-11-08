import { http, HttpResponse } from 'msw';
import userEvent from '@testing-library/user-event';
import { describe, expect, it } from 'vitest';
import { server } from '../../test/mocks/server';
import { render, screen } from '../../test/test-utils';
import { CountdownSettingsCard } from './CountdownSettingsCard';

const API = '*/api/v1';
const base = { broadcaster_id: '1234', excluded: false, featured: false, clip_download_authorized: false, consent_text_version: '2026-11-07' };

describe('CountdownSettingsCard', () => {
  it('turns "leave my clips out" on and off', async () => {
    const sent: unknown[] = [];
    server.use(
      http.get(`${API}/users/me/countdowns`, () => HttpResponse.json({ success: true, data: base })),
      http.put(`${API}/users/me/countdowns/exclusion`, async ({ request }) => {
        const body = (await request.json()) as { excluded: boolean };
        sent.push(body);
        return HttpResponse.json({ success: true, data: { ...base, excluded: body.excluded, exclusion_source: 'self_service' } });
      }),
    );
    render(<CountdownSettingsCard />);
    const toggle = await screen.findByRole('switch', { name: /Leave my clips out/ });
    expect(toggle).not.toBeChecked();
    await userEvent.click(toggle);
    await screen.findByRole('switch', { name: /Leave my clips out/, checked: true });
    await userEvent.click(screen.getByRole('switch', { name: /Leave my clips out/ }));
    await screen.findByRole('switch', { name: /Leave my clips out/, checked: false });
    expect(sent).toEqual([{ excluded: true }, { excluded: false }]);
    expect(screen.getByText('Wording version 2026-11-07')).toBeInTheDocument();
  });

  it('locks the toggle when staff excluded the broadcaster', async () => {
    server.use(http.get(`${API}/users/me/countdowns`, () =>
      HttpResponse.json({ success: true, data: { ...base, excluded: true, exclusion_source: 'admin' } })));
    render(<CountdownSettingsCard />);
    const toggle = await screen.findByRole('switch', { name: /Leave my clips out/ });
    expect(toggle).toBeChecked();
    expect(toggle).toBeDisabled();
    expect(screen.getByText(/left out at your request/)).toBeInTheDocument();
  });

  it('offers "feature my clips" through Twitch and lets the broadcaster stop', async () => {
    server.use(
      http.get(`${API}/users/me/countdowns`, () => HttpResponse.json({ success: true, data: base })),
    );
    const { unmount } = render(<CountdownSettingsCard />);
    const link = await screen.findByRole('link', { name: 'Feature my clips' });
    expect(link.getAttribute('href')).toBe('/api/v1/twitch/oauth/authorize?purpose=feature&return_to=%2Fsettings');
    unmount();

    server.use(
      http.get(`${API}/users/me/countdowns`, () => HttpResponse.json({ success: true, data: { ...base, featured: true } })),
      http.delete(`${API}/users/me/countdowns/feature`, () => HttpResponse.json({ success: true, data: base })),
    );
    render(<CountdownSettingsCard />);
    await userEvent.click(await screen.findByRole('button', { name: 'Stop featuring my clips' }));
    expect(await screen.findByRole('link', { name: 'Feature my clips' })).toBeInTheDocument();
  });

  it('asks users without a Twitch account to sign in with Twitch', async () => {
    server.use(http.get(`${API}/users/me/countdowns`, () =>
      HttpResponse.json({ success: false, error: { code: 'NO_TWITCH_ACCOUNT' } }, { status: 409 })));
    render(<CountdownSettingsCard />);
    expect(await screen.findByText(/Sign in with Twitch/)).toBeInTheDocument();
  });
});
