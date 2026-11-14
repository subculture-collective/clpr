import { isAxiosError } from 'axios';
import apiClient from './api';
import type {
    StreamerClipRoom,
    StreamerClipRoomItem,
    StreamerClipRoomWithItems,
} from '@/types/streamerClipRoom';

interface StandardResponse<T> {
    success: boolean;
    data?: T;
    error?: {
        code: string;
        message: string;
    };
}

/** A failed clip room request, carrying the backend's error code and message. */
export class StreamerClipRoomApiError extends Error {
    readonly code?: string;
    readonly status?: number;

    constructor(message: string, code?: string, status?: number) {
        super(message);
        this.name = 'StreamerClipRoomApiError';
        this.code = code;
        this.status = status;
    }
}

const unwrapResponse = <T>(response: StandardResponse<T>, fallbackMessage: string): T => {
    if (!response.success || response.data === undefined) {
        throw new StreamerClipRoomApiError(
            response.error?.message || fallbackMessage,
            response.error?.code,
        );
    }

    return response.data;
};

// axios rejects non-2xx responses before unwrapResponse runs, so the backend's
// code and message have to be read from the error's response body.
const toApiError = (error: unknown, fallbackMessage: string): Error => {
    if (error instanceof StreamerClipRoomApiError) return error;
    if (!isAxiosError(error) || !error.response) {
        return new StreamerClipRoomApiError(fallbackMessage);
    }

    const status = error.response.status;
    // Handlers send { error: { code, message } }; the rate limiter sends { error: string }.
    const body = (error.response.data as { error?: unknown } | undefined)?.error;
    let code: string | undefined;
    let message: string | undefined;
    if (typeof body === 'string') {
        message = body;
    } else if (body && typeof body === 'object') {
        const info = body as { code?: unknown; message?: unknown };
        if (typeof info.code === 'string') code = info.code;
        if (typeof info.message === 'string') message = info.message;
    }
    if (status === 429) code = 'RATE_LIMITED';

    return new StreamerClipRoomApiError(message || fallbackMessage, code, status);
};

const request = async <T>(
    call: () => Promise<{ data: StandardResponse<T> }>,
    fallbackMessage: string,
): Promise<T> => {
    try {
        const response = await call();
        return unwrapResponse(response.data, fallbackMessage);
    } catch (error) {
        throw toApiError(error, fallbackMessage);
    }
};

export const streamerClipRoomApi = {
    async getRoom(channel: string): Promise<StreamerClipRoomWithItems> {
        return request(
            () =>
                apiClient.get<StandardResponse<StreamerClipRoomWithItems>>(
                    `/streamer-clip-rooms/${encodeURIComponent(channel)}`,
                ),
            'Failed to load streamer clip room',
        );
    },

    async startRoom(channel: string): Promise<void> {
        await request(
            () =>
                apiClient.post<StandardResponse<StreamerClipRoom>>(
                    `/streamer-clip-rooms/${encodeURIComponent(channel)}/start`,
                ),
            'Could not start the listener. Try again.',
        );
    },

    async stopRoom(channel: string): Promise<void> {
        await request(
            () =>
                apiClient.post<StandardResponse<StreamerClipRoom>>(
                    `/streamer-clip-rooms/${encodeURIComponent(channel)}/stop`,
                ),
            'Could not stop the listener. Try again.',
        );
    },

    async updateSubmissions(
        channel: string,
        enabled: boolean,
        durationMinutes?: number,
    ): Promise<StreamerClipRoom> {
        return request(
            () =>
                apiClient.put<StandardResponse<StreamerClipRoom>>(
                    `/streamer-clip-rooms/${encodeURIComponent(channel)}/submissions`,
                    {
                        enabled,
                        ...(durationMinutes === undefined
                            ? {}
                            : { duration_minutes: durationMinutes }),
                    },
                ),
            'Could not update submissions. Try again.',
        );
    },

    async getRoomItems(
        roomId: string,
        status: 'pending' | 'approved' | 'rejected' | 'skipped' | 'all' = 'all',
    ): Promise<StreamerClipRoomItem[]> {
        return request(
            () =>
                apiClient.get<StandardResponse<StreamerClipRoomItem[]>>(
                    `/streamer-clip-rooms/${encodeURIComponent(roomId)}/items`,
                    {
                        params: { status },
                    },
                ),
            'Failed to load streamer clip room items',
        );
    },

    async approveItem(roomId: string, itemId: string): Promise<StreamerClipRoomItem> {
        return request(
            () =>
                apiClient.post<StandardResponse<StreamerClipRoomItem>>(
                    `/streamer-clip-rooms/${encodeURIComponent(roomId)}/items/${encodeURIComponent(itemId)}/approve`,
                ),
            'Could not approve the clip. Try again.',
        );
    },

    async rejectItem(roomId: string, itemId: string): Promise<StreamerClipRoomItem> {
        return request(
            () =>
                apiClient.post<StandardResponse<StreamerClipRoomItem>>(
                    `/streamer-clip-rooms/${encodeURIComponent(roomId)}/items/${encodeURIComponent(itemId)}/reject`,
                ),
            'Could not reject the clip. Try again.',
        );
    },

    async reorderItems(roomId: string, itemIds: string[]): Promise<void> {
        await request(
            () =>
                apiClient.put<StandardResponse<{ message: string }>>(
                    `/streamer-clip-rooms/${encodeURIComponent(roomId)}/items/order`,
                    { item_ids: itemIds },
                ),
            'Could not save the new order. Try again.',
        );
    },
};
