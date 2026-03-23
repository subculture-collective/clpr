import { useState } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { fetchUserPlaylists, addClipToPlaylist } from '@/lib/playlist-api';
import { useIsAuthenticated, useToast } from '@/hooks';
import { Link } from 'react-router-dom';
import { Modal } from '@/components/ui';
import { ClipboardList, ListMusic } from 'lucide-react';

interface AddToPlaylistButtonProps {
    clipId: string;
}

export function AddToPlaylistButton({ clipId }: AddToPlaylistButtonProps) {
    const [isOpen, setIsOpen] = useState(false);
    const isAuthenticated = useIsAuthenticated();
    const toast = useToast();
    const queryClient = useQueryClient();

    const { data: playlists = [], isLoading } = useQuery({
        queryKey: ['user-playlists'],
        queryFn: fetchUserPlaylists,
        enabled: isAuthenticated && isOpen,
    });

    const addMutation = useMutation({
        mutationFn: ({ playlistId }: { playlistId: string }) =>
            addClipToPlaylist(playlistId, clipId),
        onSuccess: () => {
            toast.success('Clip added to playlist');
            queryClient.invalidateQueries({ queryKey: ['user-playlists'] });
            setIsOpen(false);
        },
        onError: (error: Error) => {
            toast.error(error.message || 'Failed to add clip to playlist');
        },
    });

    const handleClick = () => {
        if (!isAuthenticated) {
            toast.info('Please log in to add clips to playlists');
            return;
        }
        setIsOpen(true);
    };

    const handleAddToPlaylist = (playlistId: string) => {
        addMutation.mutate({ playlistId });
    };

    return (
        <>
            <button
                onClick={handleClick}
                disabled={!isAuthenticated}
                className={`text-muted-foreground hover:text-foreground flex items-center gap-1.5 transition-colors touch-target min-h-11 ${
                    !isAuthenticated ?
                        'opacity-50 cursor-not-allowed hover:bg-transparent'
                    :   'cursor-pointer'
                }`}
                aria-label={
                    !isAuthenticated ?
                        'Log in to add to playlist'
                    :   'Add to playlist'
                }
                aria-disabled={!isAuthenticated}
                title={
                    !isAuthenticated ? 'Log in to add to playlist' : undefined
                }
            >
                <ClipboardList size={18} className='shrink-0' strokeWidth={1.75} />
                <span className='hidden sm:inline'>Playlist</span>
            </button>

            <Modal
                open={isOpen}
                onClose={() => setIsOpen(false)}
                title='Add to Playlist'
                size='sm'
            >
                {isLoading ?
                    <div className='py-8 text-center text-muted-foreground'>
                        <div className='inline-block w-6 h-6 border-2 border-current border-t-transparent rounded-full animate-spin mb-3' />
                        <p>Loading your playlists...</p>
                    </div>
                : playlists.length === 0 ?
                    <div className='py-8 text-center'>
                        <div className='mb-3 text-text-tertiary'><ClipboardList size={40} strokeWidth={1.5} /></div>
                        <p className='text-muted-foreground mb-4'>
                            You don't have any playlists yet
                        </p>
                        <Link
                            to='/playlists'
                            className='inline-flex items-center gap-2 px-4 py-2 bg-primary-600 text-white rounded-lg hover:bg-primary-700 transition-colors'
                            onClick={() => setIsOpen(false)}
                        >
                            Create your first playlist
                        </Link>
                    </div>
                :   <div className='space-y-1 -mx-2'>
                        {playlists.map(playlist => (
                            <button
                                key={playlist.id}
                                onClick={() => handleAddToPlaylist(playlist.id)}
                                disabled={addMutation.isPending}
                                className='w-full text-left px-4 py-3 rounded-lg hover:bg-muted transition-colors disabled:opacity-50 flex items-center gap-3 cursor-pointer'
                            >
                                <div className='w-10 h-10 bg-primary-100 dark:bg-primary-900/30 rounded-lg flex items-center justify-center text-primary-600 dark:text-primary-400 shrink-0'>
                                    <ListMusic size={20} strokeWidth={1.75} />
                                </div>
                                <div className='min-w-0 flex-1'>
                                    <div className='font-medium truncate'>
                                        {playlist.title}
                                    </div>
                                    {playlist.description && (
                                        <div className='text-sm text-muted-foreground truncate'>
                                            {playlist.description}
                                        </div>
                                    )}
                                </div>
                                {addMutation.isPending &&
                                    addMutation.variables?.playlistId ===
                                        playlist.id && (
                                        <div className='w-5 h-5 border-2 border-primary-600 border-t-transparent rounded-full animate-spin' />
                                    )}
                            </button>
                        ))}
                    </div>
                }
            </Modal>
        </>
    );
}
