import { http, HttpResponse } from 'msw';
import { describe, expect, it } from 'vitest';
import { server } from '../../test/mocks/server';
import { act, render, screen, within } from '../../test/test-utils';
import { ClipPacksCard } from './ClipPacksCard';

const API = '*/api/v1';

describe('ClipPacksCard', () => {
  it('lists packs with a download link for ready ones', async () => {
    server.use(
      http.get(`${API}/users/me/clip-packs`, () =>
        HttpResponse.json({
          success: true,
          data: [
            { id: 'p1', broadcaster_id: '1', broadcaster_name: 'a', week_start: '2026-11-02', status: 'ready', clip_count: 8,
              bytes: 154_000_000, expires_at: '2026-11-16T07:00:00Z', download_url: 'https://packs.example/packs/p1.zip?expires=1&sig=x' },
            { id: 'p2', broadcaster_id: '1', broadcaster_name: 'a', week_start: '2026-10-26', status: 'expired' },
            { id: 'p3', broadcaster_id: '1', broadcaster_name: 'a', week_start: '2026-10-19', status: 'failed' },
            { id: 'p4', broadcaster_id: '1', broadcaster_name: 'a', week_start: '2026-10-12', status: 'ready', withdrawn: true,
              download_url: 'https://packs.example/packs/p4.zip' },
          ],
        })),
    );
    render(<ClipPacksCard />);
    const list = await screen.findByRole('list', { name: 'Your clip packs' });
    const rows = within(list).getAllByRole('listitem');
    expect(rows[0]).toHaveTextContent('Week of Nov 2, 2026');
    expect(rows[0]).toHaveTextContent('8 clips · 154 MB · available until Nov 16');
    expect(within(rows[0]).getByRole('link', { name: /Download/ })).toHaveAttribute(
      'href', 'https://packs.example/packs/p1.zip?expires=1&sig=x');
    expect(rows[1]).toHaveTextContent('Expired after 7 days');
    expect(within(rows[1]).queryByRole('link')).not.toBeInTheDocument();
    expect(rows[2]).toHaveTextContent('Could not be made');
    expect(rows[3]).toHaveTextContent('No longer available');
    expect(within(rows[3]).queryByRole('link')).not.toBeInTheDocument();
  });

  it('renders nothing before the first pack', async () => {
    const { container } = render(<ClipPacksCard />);
    await act(async () => {
      await new Promise(resolve => setTimeout(resolve, 100));
    });
    expect(container.querySelector('[data-testid="clip-packs"]')).toBeNull();
  });
});
