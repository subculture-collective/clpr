import { apiClient } from './api';

export interface PlatformModerator {
    id: string;
    username: string;
    display_name?: string;
    avatar_url?: string;
    role: 'moderator';
}

export interface PlatformModeratorPage {
    items: PlatformModerator[];
    total: number;
    page: number;
    limit: number;
    total_pages: number;
}

export const platformModeratorApi = {
    async list(page = 1, limit = 25): Promise<PlatformModeratorPage> {
        const response = await apiClient.get<PlatformModeratorPage>(
            '/admin/moderators',
            { params: { page, limit } },
        );
        return response.data;
    },
    async add(userId: string, reason: string): Promise<void> {
        await apiClient.post('/admin/moderators', { user_id: userId, reason });
    },
    async update(userId: string, reason: string): Promise<void> {
        await apiClient.patch(`/admin/moderators/${userId}`, { reason });
    },
    async revoke(userId: string, reason: string): Promise<void> {
        await apiClient.delete(`/admin/moderators/${userId}`, {
            data: { reason },
        });
    },
};
