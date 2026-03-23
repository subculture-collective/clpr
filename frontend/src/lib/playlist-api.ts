import apiClient from './api';
import type { Playlist } from '@/types/playlist';

interface AddToPlaylistResponse {
    success: boolean;
    message?: string;
}

/**
 * Fetch user's playlists
 */
export async function fetchUserPlaylists(): Promise<Playlist[]> {
    const response = await apiClient.get<{
        success: boolean;
        data: Playlist[];
        meta: {
            page: number;
            limit: number;
            total: number;
            total_pages: number;
            has_next: boolean;
            has_prev: boolean;
        };
    }>('/playlists');
    return response.data.data || [];
}

/**
 * Add a clip to a playlist
 */
export async function addClipToPlaylist(
    playlistId: string,
    clipId: string
): Promise<AddToPlaylistResponse> {
    const response = await apiClient.post<AddToPlaylistResponse>(
        `/playlists/${playlistId}/clips`,
        { clip_id: clipId }
    );
    return response.data;
}

/**
 * Remove a clip from a playlist
 */
export async function removeClipFromPlaylist(
    playlistId: string,
    clipId: string
): Promise<void> {
    await apiClient.delete(`/playlists/${playlistId}/clips/${clipId}`);
}

/**
 * Fetch bookmarked playlists for the current user
 */
export async function fetchBookmarkedPlaylists(page = 1, limit = 20): Promise<{
    playlists: Playlist[];
    total: number;
    has_next: boolean;
}> {
    const response = await apiClient.get<{
        success: boolean;
        data: Playlist[];
        meta: { total: number; has_next: boolean };
    }>(`/playlists/bookmarks?page=${page}&limit=${limit}`);
    return {
        playlists: response.data.data || [],
        total: response.data.meta?.total || 0,
        has_next: response.data.meta?.has_next || false,
    };
}
