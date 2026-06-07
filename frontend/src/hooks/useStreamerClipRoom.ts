import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { streamerClipRoomApi } from '@/lib/streamer-clip-room-api';

const streamerClipRoomQueryKey = ['streamer-clip-room'] as const;

export const useStreamerClipRoom = (channel: string, enabled = true) => {
    return useQuery({
        queryKey: [...streamerClipRoomQueryKey, channel],
        queryFn: () => streamerClipRoomApi.getRoom(channel),
        enabled: enabled && !!channel,
    });
};

export const useStreamerClipRoomItems = (
    roomId: string,
    status: 'pending' | 'approved' | 'rejected' | 'skipped' | 'all' = 'all',
    enabled = true,
) => {
    return useQuery({
        queryKey: [...streamerClipRoomQueryKey, roomId, 'items', status],
        queryFn: () => streamerClipRoomApi.getRoomItems(roomId, status),
        enabled: enabled && !!roomId,
    });
};

export const useStartStreamerClipRoom = () => {
    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: (channel: string) => streamerClipRoomApi.startRoom(channel),
        onSuccess: () => {
            queryClient.invalidateQueries({ queryKey: streamerClipRoomQueryKey });
        },
    });
};

export const useStopStreamerClipRoom = () => {
    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: (channel: string) => streamerClipRoomApi.stopRoom(channel),
        onSuccess: () => {
            queryClient.invalidateQueries({ queryKey: streamerClipRoomQueryKey });
        },
    });
};

export const useApproveStreamerClipRoomItem = () => {
    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: ({ roomId, itemId }: { roomId: string; itemId: string }) =>
            streamerClipRoomApi.approveItem(roomId, itemId),
        onSuccess: () => {
            queryClient.invalidateQueries({ queryKey: streamerClipRoomQueryKey });
        },
    });
};

export const useRejectStreamerClipRoomItem = () => {
    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: ({ roomId, itemId }: { roomId: string; itemId: string }) =>
            streamerClipRoomApi.rejectItem(roomId, itemId),
        onSuccess: () => {
            queryClient.invalidateQueries({ queryKey: streamerClipRoomQueryKey });
        },
    });
};

export const useReorderStreamerClipRoomItems = () => {
    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: ({ roomId, itemIds }: { roomId: string; itemIds: string[] }) =>
            streamerClipRoomApi.reorderItems(roomId, itemIds),
        onSuccess: () => {
            queryClient.invalidateQueries({ queryKey: streamerClipRoomQueryKey });
        },
    });
};
