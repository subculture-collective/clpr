import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { render, screen, within } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { MemoryRouter, Route, Routes } from 'react-router-dom';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { CountdownEditionPage } from './CountdownEditionPage';
import { CountdownsPage } from './CountdownsPage';
import { CountdownFeatures } from '../components/countdown/CountdownFeatures';
import type { PublicCountdownEdition } from '../lib/countdown-api';

vi.mock('../components', () => ({
  Container: ({ children }: { children: React.ReactNode }) => <div>{children}</div>,
  SEO: ({ title }: { title?: string }) => <meta data-testid='seo-title' content={title} />,
  Spinner: () => <div role='status'>Loading</div>,
}));

vi.mock('../components/clip/TwitchEmbed', async () => {
  const { useEffect } = await import('react');
  return {
    TwitchEmbed: ({
      clipId,
      active,
      onActivate,
      onPlayingChange,
    }: {
      clipId: string;
      active?: boolean;
      onActivate?: () => void;
      onPlayingChange?: (playing: boolean) => void;
    }) => {
      useEffect(() => onPlayingChange?.(Boolean(active)), [active, onPlayingChange]);
      return (
        <button type='button' data-testid='embed' data-clip={clipId} data-active={active ? 'yes' : 'no'} onClick={onActivate}>
          play {clipId}
        </button>
      );
    },
  };
});

vi.mock('../components/countdown/CreatorConnectOffer', () => ({ CreatorConnectOffer: () => null }));

vi.mock('./NotFoundPage', () => ({ NotFoundPage: () => <h1>Page not found</h1> }));

const api = vi.hoisted(() => ({ list: vi.fn(), get: vi.fn(), featured: vi.fn() }));
vi.mock('../lib/countdown-api', async importOriginal => ({
  ...(await importOriginal<typeof import('../lib/countdown-api')>()),
  countdownApi: api,
}));

const edition: PublicCountdownEdition = {
  edition_date: '2026-11-07',
  series_slug: 'overall',
  series_name: 'Overall',
  size: 3,
  clips: [
    { position: 1, duration: 30, clip_id: 'c1', title: 'Clutch ace', broadcaster_name: 'alpha', broadcaster_id: '11', twitch_clip_id: 'tw1', twitch_clip_url: '' },
    { position: 3, duration: 30, clip_id: 'c3', title: 'Chat goes wild', broadcaster_name: 'charlie', twitch_clip_id: 'tw3', twitch_clip_url: '' },
  ],
};

function renderAt(path: string | { pathname: string; state: unknown }, element: React.ReactNode, route: string) {
  const client = new QueryClient({ defaultOptions: { queries: { retry: false } } });
  return render(
    <QueryClientProvider client={client}>
      <MemoryRouter initialEntries={[path]}>
        <Routes>
          <Route path={route} element={element} />
        </Routes>
      </MemoryRouter>
    </QueryClientProvider>,
  );
}

describe('countdown pages', () => {
  beforeEach(() => {
    api.list.mockReset();
    api.get.mockReset();
    api.featured.mockReset();
  });

  it('lists editions grouped by date and loads older ones', async () => {
    api.list
      .mockResolvedValueOnce({ editions: [edition, { ...edition, series_slug: 'just-chatting', series_name: 'Just Chatting' }], nextBefore: '2026-11-07' })
      .mockResolvedValueOnce({ editions: [{ ...edition, edition_date: '2026-11-06' }], nextBefore: null });
    renderAt('/countdowns', <CountdownsPage />, '/countdowns');

    const day = await screen.findByRole('region', { name: 'November 7, 2026' });
    const links = within(day).getAllByRole('link');
    expect(links.map(link => link.getAttribute('href'))).toEqual(['/countdowns/2026-11-07/overall', '/countdowns/2026-11-07/just-chatting']);
    expect(within(day).getAllByText('Top 3')).toHaveLength(2);

    await userEvent.click(screen.getByRole('button', { name: 'Older countdowns' }));
    expect(await screen.findByRole('region', { name: 'November 6, 2026' })).toBeInTheDocument();
    expect(api.list).toHaveBeenLastCalledWith({ before: '2026-11-07' });
    expect(screen.queryByRole('button', { name: 'Older countdowns' })).not.toBeInTheDocument();
  });

  it('says so when nothing is published', async () => {
    api.list.mockResolvedValue({ editions: [], nextBefore: null });
    renderAt('/countdowns', <CountdownsPage />, '/countdowns');
    expect(await screen.findByText('No countdowns have been published yet.')).toBeInTheDocument();
  });

  it('counts an edition down to #1 with one player at a time', async () => {
    api.get.mockResolvedValue(edition);
    renderAt('/countdowns/2026-11-07/overall', <CountdownEditionPage />, '/countdowns/:date/:series');

    expect(await screen.findByRole('heading', { level: 1, name: 'Top 3: Overall' })).toBeInTheDocument();
    expect(api.get).toHaveBeenCalledWith('2026-11-07', 'overall');
    const items = screen.getAllByRole('listitem');
    expect(items.map(item => item.getAttribute('aria-label'))).toEqual(['Number 3', 'Number 1']);
    expect(screen.getByRole('link', { name: 'charlie' })).toHaveAttribute('href', '/broadcaster/charlie');

    // One player, on the lowest rank, waiting for a press on a direct visit.
    expect(screen.getAllByTestId('embed')).toHaveLength(1);
    expect(screen.getByTestId('embed')).toHaveAttribute('data-clip', 'tw3');
    expect(screen.getByTestId('embed')).toHaveAttribute('data-active', 'no');
    expect(screen.queryByText(/Next clip in/)).not.toBeInTheDocument();

    await userEvent.click(screen.getByTestId('embed'));
    expect(screen.getByTestId('embed')).toHaveAttribute('data-active', 'yes');
    expect(screen.getByText(/Next clip in/)).toBeInTheDocument();

    await userEvent.click(screen.getByRole('button', { name: 'Skip now' }));
    expect(screen.getByTestId('embed')).toHaveAttribute('data-clip', 'tw1');
    expect(screen.getByRole('button', { name: /#1 Clutch ace/ })).toHaveAttribute('aria-current', 'true');
    expect(screen.getByRole('link', { name: 'alpha' })).toHaveAttribute('href', '/broadcaster/11');
    // #1 is the end of the list.
    expect(screen.queryByText(/Next clip in/)).not.toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'Next clip' })).toBeDisabled();

    await userEvent.click(screen.getByRole('button', { name: 'Previous clip' }));
    expect(screen.getByTestId('embed')).toHaveAttribute('data-clip', 'tw3');

    await userEvent.click(screen.getByRole('button', { name: /#1 Clutch ace/ }));
    expect(screen.getByTestId('embed')).toHaveAttribute('data-clip', 'tw1');
    expect(screen.getByTestId('embed')).toHaveAttribute('data-active', 'yes');
  });

  it('starts playing when opened from the list of countdowns', async () => {
    api.get.mockResolvedValue(edition);
    renderAt({ pathname: '/countdowns/2026-11-07/overall', state: { play: true } }, <CountdownEditionPage />, '/countdowns/:date/:series');
    await screen.findByRole('heading', { level: 1, name: 'Top 3: Overall' });
    expect(screen.getByTestId('embed')).toHaveAttribute('data-active', 'yes');
  });

  it('shows the not-found page for an edition that is not public', async () => {
    api.get.mockRejectedValue(Object.assign(new Error('Request failed'), { isAxiosError: true, response: { status: 404 } }));
    renderAt('/countdowns/2026-11-07/missing', <CountdownEditionPage />, '/countdowns/:date/:series');
    expect(await screen.findByRole('heading', { name: 'Page not found' })).toBeInTheDocument();
  });

  it('marks a featured clip and stays hidden otherwise', async () => {
    api.featured.mockResolvedValueOnce([
      { edition_date: '2026-11-07', series_slug: 'overall', series_name: 'Overall', position: 2, clip_id: 'c1', title: 'Clutch ace' },
    ]);
    const { unmount } = renderAt('/clip/c1', <CountdownFeatures clipId='c1' />, '/clip/:id');
    const link = await screen.findByRole('link', { name: '#2 · Overall · Nov 7, 2026' });
    expect(link).toHaveAttribute('href', '/countdowns/2026-11-07/overall');
    expect(api.featured).toHaveBeenCalledWith({ clip_id: 'c1' });
    unmount();

    api.featured.mockResolvedValueOnce([]);
    renderAt('/broadcaster/11', <CountdownFeatures broadcasterId='11' />, '/broadcaster/:id');
    await vi.waitFor(() => expect(api.featured).toHaveBeenCalledWith({ broadcaster_id: '11' }));
    expect(screen.queryByTestId('countdown-features')).not.toBeInTheDocument();
  });
});
