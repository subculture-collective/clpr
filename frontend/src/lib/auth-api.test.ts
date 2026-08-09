import { describe, expect, it, vi } from 'vitest';

const { get } = vi.hoisted(() => ({ get: vi.fn() }));

vi.mock('./api', () => ({ apiClient: { get } }));

const { getCurrentUser } = await import('./auth-api');

describe('getCurrentUser', () => {
    it('marks the anonymous session probe as ineligible for token refresh', async () => {
        get.mockResolvedValue({ data: { id: 'user-1' } });

        await getCurrentUser({ anonymousProbe: true });

        expect(get).toHaveBeenCalledWith('/auth/me', {
            skipAuthRefresh: true,
        });
    });

    it('leaves authenticated restoration eligible for token refresh', async () => {
        get.mockResolvedValue({ data: { id: 'user-1' } });

        await getCurrentUser();

        expect(get).toHaveBeenCalledWith('/auth/me', undefined);
    });
});
