import { beforeEach, describe, expect, it, vi } from 'vitest';

const { get } = vi.hoisted(() => ({ get: vi.fn() }));

vi.mock('./api', () => ({ default: { get } }));

const { listChannels } = await import('./chat-api');

describe('listChannels', () => {
  beforeEach(() => get.mockReset());

  it('uses the versioned API client route and returns channel envelopes', async () => {
    const channel = { id: 'channel-1', name: 'General' };
    get.mockResolvedValue({ data: { channels: [channel], limit: 50, offset: 0 } });

    await expect(listChannels()).resolves.toEqual([channel]);
    expect(get).toHaveBeenCalledWith('/chat/channels');
  });

  it.each([
    ['a direct array', [{ id: 'channel-1', name: 'General' }]],
    ['null channels', { channels: null }],
    ['an omitted envelope', {}],
  ])('normalizes %s', async (_name, response) => {
    get.mockResolvedValue({ data: response });

    const channels = await listChannels();

    expect(channels).toEqual(Array.isArray(response) ? response : []);
  });
});
