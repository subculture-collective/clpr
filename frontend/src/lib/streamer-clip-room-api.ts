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

const unwrapResponse = <T>(response: StandardResponse<T>, fallbackMessage: string): T => {
    if (!response.success || response.data === undefined) {
        throw new Error(response.error?.message || fallbackMessage);
    }

    return response.data;
};

export const streamerClipRoomApi = {
    async getRoom(channel: string): Promise<StreamerClipRoomWithItems> {
        const response = await apiClient.get<StandardResponse<StreamerClipRoomWithItems>>(
            `/streamer-clip-rooms/${encodeURIComponent(channel)}`,
        );

        return unwrapResponse(response.data, 'Failed to load streamer clip room');
    },

    async startRoom(channel: string): Promise<void> {
        const response = await apiClient.post<StandardResponse<StreamerClipRoom>>(
            `/streamer-clip-rooms/${encodeURIComponent(channel)}/start`,
        );

        unwrapResponse(response.data, 'Failed to start streamer clip room');
    },

    async stopRoom(channel: string): Promise<void> {
        const response = await apiClient.post<StandardResponse<StreamerClipRoom>>(
            `/streamer-clip-rooms/${encodeURIComponent(channel)}/stop`,
        );

        unwrapResponse(response.data, 'Failed to stop streamer clip room');
    },

    async getRoomItems(
        roomId: string,
        status: 'pending' | 'approved' | 'rejected' | 'skipped' | 'all' = 'all',
    ): Promise<StreamerClipRoomItem[]> {
        const response = await apiClient.get<StandardResponse<StreamerClipRoomItem[]>>(
            `/streamer-clip-rooms/${encodeURIComponent(roomId)}/items`,
            {
                params: { status },
            },
        );

        return unwrapResponse(response.data, 'Failed to load streamer clip room items');
    },

    async approveItem(roomId: string, itemId: string): Promise<StreamerClipRoomItem> {
        const response = await apiClient.post<StandardResponse<StreamerClipRoomItem>>(
            `/streamer-clip-rooms/${encodeURIComponent(roomId)}/items/${encodeURIComponent(itemId)}/approve`,
        );

        return unwrapResponse(response.data, 'Failed to approve streamer clip room item');
    },

    async rejectItem(roomId: string, itemId: string): Promise<StreamerClipRoomItem> {
        const response = await apiClient.post<StandardResponse<StreamerClipRoomItem>>(
            `/streamer-clip-rooms/${encodeURIComponent(roomId)}/items/${encodeURIComponent(itemId)}/reject`,
        );

        return unwrapResponse(response.data, 'Failed to reject streamer clip room item');
    },

    async reorderItems(roomId: string, itemIds: string[]): Promise<void> {
        const response = await apiClient.put<StandardResponse<{ message: string }>>(
            `/streamer-clip-rooms/${encodeURIComponent(roomId)}/items/order`,
            { item_ids: itemIds },
        );

        unwrapResponse(response.data, 'Failed to reorder streamer clip room items');
    },
};
