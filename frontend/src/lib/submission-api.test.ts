import { beforeEach, describe, expect, it, vi } from 'vitest';

const { get } = vi.hoisted(() => ({ get: vi.fn() }));

vi.mock('./api', () => ({ default: { get } }));

const { getUserSubmissions } = await import('./submission-api');

describe('getUserSubmissions', () => {
  beforeEach(() => get.mockReset());

  it('normalizes null data and missing pagination from rolling deployments', async () => {
    get.mockResolvedValue({ data: { success: true, data: null, meta: null } });

    await expect(getUserSubmissions(2, 20)).resolves.toEqual({
      success: true,
      data: [],
      meta: { page: 2, limit: 20, total: 0, total_pages: 0 },
    });
  });

  it('preserves populated submissions and valid metadata', async () => {
    const response = {
      success: true,
      data: [{ id: 'submission-1' }],
      meta: { page: 1, limit: 10, total: 1, total_pages: 1 },
    };
    get.mockResolvedValue({ data: response });

    await expect(getUserSubmissions(1, 10)).resolves.toEqual(response);
  });
});
