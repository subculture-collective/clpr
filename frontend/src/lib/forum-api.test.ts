import { beforeEach, describe, expect, it, vi } from 'vitest';

const { get, post } = vi.hoisted(() => ({ get: vi.fn(), post: vi.fn() }));

vi.mock('./api', () => ({ apiClient: { get, post } }));

const { forumApi } = await import('./forum-api');
const { parseForumSort, toBackendForumSort } = await import('./forum-sort');

const thread = (id: string, tags: string[]) => ({
    id,
    user_id: 'u1',
    username: 'someone',
    title: `Thread ${id}`,
    content: 'body',
    tags,
    view_count: 0,
    reply_count: 0,
    locked: false,
    pinned: false,
    created_at: '2026-09-24T00:00:00Z',
    updated_at: '2026-09-24T00:00:00Z',
});

function requestedParams() {
    const url = String(get.mock.calls[0][0]);
    return new URLSearchParams(url.slice(url.indexOf('?') + 1));
}

describe('forum thread sort names', () => {
    it.each([
        ['newest', 'recent'],
        ['most-replied', 'replies'],
        ['popular', 'popular'],
    ] as const)('sends %s as the backend value %s', (ui, backend) => {
        expect(toBackendForumSort(ui)).toBe(backend);
    });

    it.each([
        [null, 'newest'],
        ['newest', 'newest'],
        ['recent', 'newest'],
        ['most-replied', 'most-replied'],
        ['replies', 'most-replied'],
        ['trending', 'popular'],
        ['hot', 'popular'],
        ['popular', 'popular'],
        ['bogus', 'newest'],
    ])('reads the URL sort %s as %s', (value, expected) => {
        expect(parseForumSort(value)).toBe(expected);
    });
});

describe('forumApi.listThreads', () => {
    beforeEach(() => get.mockReset());

    it('uses the backend parameter names for sort and game', async () => {
        get.mockResolvedValue({ data: { success: true, data: [], meta: { page: 1, limit: 20, count: 0 } } });
        const gameId = '5b0f4c4e-8f3a-4a57-9a53-2f6c1f0d1e11';

        await forumApi.listThreads({ sort: 'most-replied', game_id: gameId, page: 2 });

        const params = requestedParams();
        expect(params.get('sort')).toBe('replies');
        expect(params.get('game_filter')).toBe(gameId);
        expect(params.has('game_id')).toBe(false);
        expect(params.get('page')).toBe('2');
    });

    it('drops a game filter the backend would reject as malformed', async () => {
        get.mockResolvedValue({ data: { success: true, data: [], meta: {} } });

        await forumApi.listThreads({ sort: 'newest', game_id: '509658' });

        const params = requestedParams();
        expect(params.get('sort')).toBe('recent');
        expect(params.has('game_filter')).toBe(false);
    });

    it('maps the envelope and filters topic tags the backend ignores', async () => {
        get.mockResolvedValue({
            data: {
                success: true,
                data: [thread('a', ['help']), thread('b', ['bug-report']), thread('c', [])],
                meta: { page: 1, limit: 20, count: 3 },
            },
        });

        const response = await forumApi.listThreads({ tags: ['help'] });

        expect(response.threads.map(t => t.id)).toEqual(['a']);
        expect(response.page).toBe(1);
        expect(requestedParams().getAll('tags')).toEqual(['help']);
    });
});

describe('forum reply votes', () => {
    beforeEach(() => {
        get.mockReset();
        post.mockReset();
    });

    it('reads vote stats from the response envelope', async () => {
        const stats = { upvotes: 2, downvotes: 1, net_votes: 1, user_vote: 1 };
        get.mockResolvedValue({ data: { success: true, data: stats } });

        await expect(forumApi.getReplyVotes('r1')).resolves.toEqual(stats);
        expect(get).toHaveBeenCalledWith('/forum/replies/r1/votes');
    });

    it('posts the vote value through the API client', async () => {
        post.mockResolvedValue({ data: { success: true } });

        await forumApi.voteOnReply('r1', -1);

        expect(post).toHaveBeenCalledWith('/forum/replies/r1/vote', { vote_value: -1 });
    });
});
