import { http, HttpResponse } from 'msw';
import { describe, expect, it } from 'vitest';
import { server } from '../../test/mocks/server';
import { act, render, screen, within } from '../../test/test-utils';
import { CountdownPlacementsCard } from './CountdownPlacementsCard';

const API = '*/api/v1';
const controls = { broadcaster_id: '1234', excluded: false, featured: true, clip_download_authorized: true, consent_text_version: '2026-11-07' };

describe('CountdownPlacementsCard', () => {
  it('lists placements with the views of their videos', async () => {
    server.use(
      http.get(`${API}/users/me/countdowns`, () => HttpResponse.json({ success: true, data: controls })),
      http.get(`${API}/users/me/countdowns/placements`, () =>
        HttpResponse.json({
          success: true,
          data: [
            {
              edition_date: '2026-11-09', series_slug: 'overall', series_name: 'Overall', position: 2,
              clip_id: 'c1', title: 'Huge clutch',
              countdown: { youtube_video_id: 'vidCountdown1', view_count: 12345 },
              short: { youtube_video_id: 'vidShort1' },
            },
            { edition_date: '2026-11-08', series_slug: 'just-chatting', series_name: 'Just Chatting', position: 7, clip_id: 'c2', title: 'Quiet one' },
          ],
        })),
    );
    render(<CountdownPlacementsCard returnTo='/creator/1234/dashboard' />);
    const list = await screen.findByRole('list', { name: 'Your countdown placements' });
    const rows = within(list).getAllByRole('listitem');
    expect(rows).toHaveLength(2);
    expect(within(rows[0]).getByRole('link', { name: /#2 · Overall/ })).toHaveAttribute('href', '/countdowns/2026-11-09/overall');
    expect(within(rows[0]).getByRole('link', { name: /Countdown video 12,345 views/ })).toHaveAttribute(
      'href', 'https://www.youtube.com/watch?v=vidCountdown1');
    expect(within(rows[0]).getByRole('link', { name: /Short views pending/ })).toHaveAttribute(
      'href', 'https://www.youtube.com/shorts/vidShort1');
    expect(within(rows[1]).queryByRole('link', { name: /YouTube/ })).not.toBeInTheDocument();
    expect(screen.queryByRole('link', { name: 'Connect my channel' })).not.toBeInTheDocument();
  });

  it('offers to connect the channel until the broadcaster features their clips', async () => {
    server.use(
      http.get(`${API}/users/me/countdowns`, () => HttpResponse.json({ success: true, data: { ...controls, featured: false } })),
      http.get(`${API}/users/me/countdowns/placements`, () => HttpResponse.json({ success: true, data: [] })),
    );
    render(<CountdownPlacementsCard returnTo='/creator/1234/dashboard' />);
    const connect = await screen.findByRole('link', { name: 'Connect my channel' });
    expect(connect.getAttribute('href')).toContain('purpose=feature');
    expect(connect.getAttribute('href')).toContain(encodeURIComponent('/creator/1234/dashboard'));
    expect(await screen.findByText(/None of your clips has placed/)).toBeInTheDocument();
  });

  it('renders nothing for an account without Twitch', async () => {
    server.use(
      http.get(`${API}/users/me/countdowns`, () => HttpResponse.json({ success: false }, { status: 409 })),
      http.get(`${API}/users/me/countdowns/placements`, () => HttpResponse.json({ success: false }, { status: 409 })),
    );
    const { container } = render(<CountdownPlacementsCard returnTo='/creator/1234/dashboard' />);
    await act(async () => {
      await new Promise(resolve => setTimeout(resolve, 100));
    });
    expect(container.querySelector('[data-testid="countdown-placements"]')).toBeNull();
  });
});
