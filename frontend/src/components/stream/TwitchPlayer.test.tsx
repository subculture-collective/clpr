import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { render, screen, waitFor } from '@testing-library/react';
import { afterEach, beforeEach, expect, it, vi } from 'vitest';
import { TwitchPlayer } from './TwitchPlayer';

const fetchStatus = vi.hoisted(() => vi.fn());
vi.mock('../../lib/stream-api', () => ({ fetchStreamStatus: fetchStatus }));
vi.mock('./StreamOfflineScreen', () => ({ StreamOfflineScreen: () => <p>Stream offline</p> }));

beforeEach(() => fetchStatus.mockReset());
afterEach(() => { delete window.Twitch; });

it.each([false, true])('replaces and releases SDK frames when destroy is available: %s', async (hasDestroy) => {
    fetchStatus.mockResolvedValue({ is_live: true, viewer_count: 10 });
    const destroy = vi.fn();
    const Embed = vi.fn(function (elementId: string, _options: Record<string, unknown>) {
        document.getElementById(elementId)?.appendChild(document.createElement('iframe'));
        return { addEventListener: vi.fn(), getPlayer: vi.fn(), ...(hasDestroy ? { destroy } : {}) };
    });
    window.Twitch = { Embed };
    const client = new QueryClient();
    const view = (showChat: boolean) => <QueryClientProvider client={client}><TwitchPlayer channel='streamer' showChat={showChat} /></QueryClientProvider>;
    const { rerender, unmount, container } = render(view(false));
    await waitFor(() => expect(Embed).toHaveBeenCalledTimes(1));
    expect(Embed.mock.calls[0][1]).toMatchObject({ channel: 'streamer', parent: [window.location.hostname], layout: 'video' });
    rerender(view(true));
    await waitFor(() => expect(Embed).toHaveBeenCalledTimes(2));
    expect(container.querySelectorAll('iframe')).toHaveLength(1);
    expect(Embed.mock.calls[1][1]).toMatchObject({ layout: 'video-with-chat' });
    unmount();
    expect(destroy).toHaveBeenCalledTimes(hasDestroy ? 2 : 0);
    expect(container.childElementCount).toBe(0);
    client.clear();
});

it('does not create a live embed for an offline channel', async () => {
    fetchStatus.mockResolvedValue({ is_live: false });
    const Embed = vi.fn();
    window.Twitch = { Embed };
    const client = new QueryClient();
    render(<QueryClientProvider client={client}><TwitchPlayer channel='offline' /></QueryClientProvider>);
    expect(await screen.findByText('Stream offline')).toBeInTheDocument();
    expect(Embed).not.toHaveBeenCalled();
    client.clear();
});
