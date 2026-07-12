import { useState, useCallback, useMemo } from 'react';
import { useQueryClient } from '@tanstack/react-query';
import { useNavigate, useParams } from 'react-router-dom';
import {
    usePlaylist,
    useRemoveClipFromPlaylist,
    useReorderPlaylistClips,
} from '@/hooks/usePlaylist';
import { PlaylistTheatreMode } from '@/components/playlist/PlaylistTheatreMode';
import { Spinner } from '@/components/ui';
import { SEO } from '@/components/SEO';
import type { PlaylistItem } from '@/components/playlist/PlaylistTheatreMode';

export function PlaylistTheatrePage() {
    const { id } = useParams<{ id: string }>();
    const navigate = useNavigate();
    const {
        data: playlist,
        isLoading,
        isError,
    } = usePlaylist(id || '', 1, 500);
    const removeClip = useRemoveClipFromPlaylist();
    const reorderClips = useReorderPlaylistClips();
    const queryClient = useQueryClient();

    const [selectedItemId, setCurrentItemId] = useState<string | null>(null);

    // Convert playlist clips to playlist items format
    const playlistData = playlist?.data;
    const playlistItems: PlaylistItem[] = useMemo(
        () =>
            playlistData?.clips?.map(clip => ({
                id: `${playlistData.id}-${clip.id}`, // Composite ID
                clip,
                clip_id: clip.id,
                order: clip.order,
            })) ?? [],
        [playlistData],
    );
    const currentItemId =
        selectedItemId &&
        playlistItems.some(item => item.id === selectedItemId)
            ? selectedItemId
            : (playlistItems[0]?.id ?? null);

    const handleItemClick = useCallback((item: PlaylistItem) => {
        setCurrentItemId(item.id);
    }, []);

    const handleItemRemove = useCallback(
        (itemId: string) => {
            if (!id) return;

            // Extract clip ID from composite ID
            const item = playlistItems.find(i => i.id === itemId);
            if (!item) return;

            // If removing current item, move to next
            if (itemId === currentItemId) {
                const currentIndex = playlistItems.findIndex(
                    i => i.id === itemId,
                );
                if (currentIndex < playlistItems.length - 1) {
                    setCurrentItemId(playlistItems[currentIndex + 1].id);
                } else {
                    setCurrentItemId(null);
                }
            }

            removeClip.mutate({ playlistId: id, clipId: item.clip_id });
        },
        [id, currentItemId, playlistItems, removeClip],
    );

    const handleReorder = useCallback(
        (itemId: string, newPosition: number) => {
            if (!id) return;

            // Extract clip ID from composite ID
            const item = playlistItems.find(i => i.id === itemId);
            if (!item) return;

            const draggedIndex = playlistItems.findIndex(i => i.id === itemId);
            if (draggedIndex === -1) return;

            // Reorder array
            const newOrder = [...playlistItems];
            const [draggedItem] = newOrder.splice(draggedIndex, 1);
            newOrder.splice(newPosition, 0, draggedItem);

            // Get clip IDs in new order
            const clipIds = newOrder.map(i => i.clip_id);

            reorderClips.mutate({ id, data: { clip_ids: clipIds } });
        },
        [id, playlistItems, reorderClips],
    );

    const handleClose = useCallback(() => {
        // Always navigate to playlist detail page when closing
        navigate(`/playlists/${id}`);
    }, [navigate, id]);

    const handleClipUpdated = useCallback(() => {
        if (!id) return;
        queryClient.invalidateQueries({ queryKey: ['playlist', id] });
    }, [id, queryClient]);

    if (isLoading) {
        return (
            <>
                <SEO title='Playlist Theatre Mode' />
                <div className='fixed inset-0 bg-black flex items-center justify-center'>
                    <Spinner size='lg' />
                </div>
            </>
        );
    }

    if (isError || !playlist) {
        return (
            <>
                <SEO title='Playlist Theatre Mode' />
                <div className='fixed inset-0 bg-black flex items-center justify-center text-white'>
                    <div className='text-center'>
                        <p className='text-xl mb-4'>Failed to load playlist</p>
                        <button
                            onClick={() => navigate('/playlists')}
                            className='px-4 py-2 bg-primary-500 hover:bg-primary-600 rounded-lg transition-colors'
                        >
                            Back to Playlists
                        </button>
                    </div>
                </div>
            </>
        );
    }

    if (!playlistData || playlistItems.length === 0) {
        return (
            <>
                <SEO
                    title={`${playlistData?.title || 'Playlist'} - Theatre Mode`}
                />
                <div className='fixed inset-0 bg-black flex items-center justify-center text-white'>
                    <div className='text-center'>
                        <p className='text-xl mb-4'>This playlist is empty</p>
                        <button
                            onClick={() => navigate(`/playlists/${id}`)}
                            className='px-4 py-2 bg-primary-500 hover:bg-primary-600 rounded-lg transition-colors'
                        >
                            Back to Playlist
                        </button>
                    </div>
                </div>
            </>
        );
    }

    return (
        <>
            <SEO
                title={`${playlistData.title} - Theatre Mode`}
                description={
                    playlistData.description ||
                    `Watch ${playlistData.title} in theatre mode`
                }
            />
            <div className='min-h-screen bg-black'>
                <PlaylistTheatreMode
                    title={playlistData.title}
                    items={playlistItems}
                    currentItemId={currentItemId}
                    onItemClick={handleItemClick}
                    onItemRemove={handleItemRemove}
                    onReorder={handleReorder}
                    onClipUpdated={handleClipUpdated}
                    onClose={handleClose}
                    isQueue={false}
                    contained={false}
                />
            </div>
        </>
    );
}
