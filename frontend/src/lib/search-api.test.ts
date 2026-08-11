import { beforeEach, describe, expect, it, vi } from 'vitest';

const { get } = vi.hoisted(() => ({ get: vi.fn() }));

vi.mock('./api', () => ({ default: { get } }));

const { searchApi } = await import('./search-api');

const responseBase = {
    query: 'nothing',
    counts: { clips: 0, creators: 0, games: 0, twitch_categories: 0, tags: 0 },
    meta: { page: 1, limit: 20, total_items: 0, total_pages: 0 },
};

describe('searchApi.search', () => {
    beforeEach(() => get.mockReset());

    it.each([
        ['omitted result collections', {}],
        ['null result collections', { clips: null, creators: null, games: null, tags: null }],
        ['a null results object', null],
    ])('normalizes %s from rolling-deployment responses', async (_name, results) => {
        get.mockResolvedValue({ data: { ...responseBase, results } });

        const response = await searchApi.search({ query: 'nothing' });

        expect(response.results).toEqual({
            clips: [],
            creators: [],
            games: [],
            twitch_categories: [],
            tags: [],
        });
    });

    it('normalizes an omitted results object', async () => {
        get.mockResolvedValue({ data: responseBase });

        const response = await searchApi.search({ query: 'nothing' });

        expect(response.results).toEqual({ clips: [], creators: [], games: [], twitch_categories: [], tags: [] });
    });

    it('preserves explicitly empty result collections', async () => {
        const results = { clips: [], creators: [], games: [], tags: [] };
        get.mockResolvedValue({ data: { ...responseBase, results } });

        await expect(searchApi.search({ query: 'nothing' })).resolves.toMatchObject({ results });
    });

    it('preserves populated result collections', async () => {
        const results = {
            clips: [{ id: 'clip-1' }],
            creators: [{ id: 'creator-1' }],
            games: [{ id: 'game-1' }],
            tags: [{ id: 'tag-1' }],
        };
        get.mockResolvedValue({
            data: {
                ...responseBase,
                results,
            },
        });

        const response = await searchApi.search({ query: 'clip' });

        expect(response.results).toEqual({
            ...results,
            twitch_categories: results.games,
        });
    });

    it('serializes filtered search parameters', async () => {
        get.mockResolvedValue({
            data: {
                ...responseBase,
                results: { clips: [], creators: [], games: [], tags: [] },
            },
        });

        await searchApi.search({
            query: 'clip',
            type: 'clips',
            sort: 'recent',
            gameId: 'game-1',
            creatorId: 'creator-1',
            language: 'en',
            tags: ['speedrun'],
            minVotes: 5,
            dateFrom: '2026-01-01',
            dateTo: '2026-01-31',
            page: 2,
            limit: 10,
        });

        expect(get).toHaveBeenCalledWith('/search', {
            params: expect.objectContaining({
                q: 'clip',
                type: 'clips',
                sort: 'recent',
                twitch_category_id: 'game-1',
                creator_id: 'creator-1',
                language: 'en',
                tags: ['speedrun'],
                min_votes: 5,
                date_from: '2026-01-01',
                date_to: '2026-01-31',
                page: 2,
                limit: 10,
            }),
        });
    });

    it('propagates search failures', async () => {
        get.mockImplementationOnce(async () => {
            throw new Error('search unavailable');
        });

        const message = await searchApi.search({ query: 'clip' }).then(
            () => 'unexpected success',
            (error: Error) => error.message,
        );
        expect(message).toBe('search unavailable');
    });
});
