import { AxiosError, type AxiosResponse } from 'axios';
import { beforeEach, describe, expect, it, vi } from 'vitest';

const post = vi.fn();
const put = vi.fn();
vi.mock('./api', () => ({
    default: {
        get: vi.fn(),
        post: (...args: unknown[]) => post(...args),
        put: (...args: unknown[]) => put(...args),
    },
}));

import { StreamerClipRoomApiError, streamerClipRoomApi } from './streamer-clip-room-api';

function httpError(status: number, data: unknown) {
    return new AxiosError(
        `Request failed with status code ${status}`,
        'ERR_BAD_REQUEST',
        undefined,
        undefined,
        { status, data } as AxiosResponse,
    );
}

describe('streamerClipRoomApi errors', () => {
    beforeEach(() => {
        post.mockReset();
        put.mockReset();
    });

    it('throws the backend code and message for a failed request', async () => {
        put.mockRejectedValue(
            httpError(409, {
                success: false,
                error: {
                    code: 'LISTENER_INACTIVE',
                    message: 'Start the Twitch listener before opening submissions',
                },
            }),
        );

        const error = await streamerClipRoomApi
            .updateSubmissions('teststreamer', true)
            .catch((err: unknown) => err);

        expect(error).toBeInstanceOf(StreamerClipRoomApiError);
        expect(error).toMatchObject({
            message: 'Start the Twitch listener before opening submissions',
            code: 'LISTENER_INACTIVE',
            status: 409,
        });
    });

    it('reads the rate limiter body, where error is a plain string', async () => {
        post.mockRejectedValue(
            httpError(429, { error: 'Rate limit exceeded. Please try again later.' }),
        );

        await expect(streamerClipRoomApi.startRoom('teststreamer')).rejects.toMatchObject({
            message: 'Rate limit exceeded. Please try again later.',
            code: 'RATE_LIMITED',
            status: 429,
        });
    });

    it('falls back to a plain message when there is no response', async () => {
        post.mockRejectedValue(new AxiosError('Network Error'));

        await expect(
            streamerClipRoomApi.approveItem('room-1', 'item-1'),
        ).rejects.toMatchObject({ message: 'Could not approve the clip. Try again.' });
    });
});
