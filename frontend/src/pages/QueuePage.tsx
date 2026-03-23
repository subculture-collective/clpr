import { Link, useNavigate } from 'react-router-dom';
import { Container } from '@/components/layout/Container';
import {
    useQueue,
    useRemoveFromQueue,
    useClearQueue,
    useReorderQueue,
} from '@/hooks/useQueue';
import { cn } from '@/lib/utils';
import {
    Trash2,
    ListPlus,
    Repeat,
    Shuffle,
    ClipboardList,
} from 'lucide-react';
import { Button, Spinner } from '@/components/ui';
import { SEO } from '@/components/SEO';
import { useState, useCallback } from 'react';
import { ConvertToPlaylistDialog } from '@/components/queue/ConvertToPlaylistDialog';
import { PlaylistTheatreMode } from '@/components/playlist/PlaylistTheatreMode';
import type { PlaylistItem } from '@/components/playlist/PlaylistTheatreMode';
import { useQueryClient } from '@tanstack/react-query';

export function QueuePage() {
    const { data: queue, isLoading, isError } = useQueue(100);
    const removeFromQueue = useRemoveFromQueue();
    const clearQueue = useClearQueue();
    const reorderQueue = useReorderQueue();
    const queryClient = useQueryClient();
    const navigate = useNavigate();

    const [showConvertDialog, setShowConvertDialog] = useState(false);
    const [currentItemId, setCurrentItemId] = useState<string | null>(null);
    const [loopEnabled, setLoopEnabled] = useState(false);
    const [shuffleEnabled, setShuffleEnabled] = useState(false);

    const queueItems = queue?.items || [];
    const total = queue?.total || 0;

    // Convert queue items to PlaylistTheatreMode format
    const playlistItems: PlaylistItem[] = queueItems.map(item => ({
        id: item.id,
        clip: item.clip,
        clip_id: item.clip_id,
        played_at: item.played_at,
    }));

    // Auto-select first item
    if (!currentItemId && playlistItems.length > 0) {
        setCurrentItemId(playlistItems[0].id);
    }

    const handleItemClick = useCallback(
        (item: PlaylistItem) => {
            setCurrentItemId(item.id);
        },
        [],
    );

    const handleItemRemove = useCallback(
        (itemId: string) => {
            if (itemId === currentItemId) {
                const currentIndex = playlistItems.findIndex(
                    item => item.id === itemId,
                );
                const nextItem = playlistItems[currentIndex + 1] || playlistItems[currentIndex - 1];
                setCurrentItemId(nextItem?.id || null);
            }
            removeFromQueue.mutate(itemId);
        },
        [currentItemId, playlistItems, removeFromQueue],
    );

    const handleReorder = useCallback(
        (itemId: string, newPosition: number) => {
            reorderQueue.mutate({
                item_id: itemId,
                new_position: newPosition + 1,
            });
        },
        [reorderQueue],
    );

    const handleClipUpdated = useCallback(() => {
        queryClient.invalidateQueries({ queryKey: ['queue'] });
    }, [queryClient]);

    const handleClearQueue = () => {
        if (
            window.confirm('Are you sure you want to clear the entire queue?')
        ) {
            clearQueue.mutate();
        }
    };

    return (
        <>
            <SEO
                title='My Queue'
                description="Your clip queue - clips you've saved to watch later"
            />

            <Container className='py-8'>
                {/* Loading State */}
                {isLoading && (
                    <div className='flex items-center justify-center py-16'>
                        <Spinner size='lg' />
                    </div>
                )}

                {/* Error State */}
                {isError && (
                    <div className='text-center py-16'>
                        <p className='text-error-600 mb-4'>
                            Failed to load queue
                        </p>
                        <Button
                            variant='outline'
                            onClick={() => window.location.reload()}
                        >
                            Try Again
                        </Button>
                    </div>
                )}

                {/* Empty State */}
                {!isLoading && !isError && queueItems.length === 0 && (
                    <div className='text-center py-16 bg-card rounded-xl border border-border'>
                        <div className='mb-4 text-text-tertiary'><ClipboardList size={48} strokeWidth={1.5} /></div>
                        <h2 className='text-xl font-semibold mb-2'>
                            Your queue is empty
                        </h2>
                        <p className='text-muted-foreground mb-6'>
                            Add clips to your queue to watch them later
                        </p>
                        <Link to='/'>
                            <Button variant='primary'>Browse Clips</Button>
                        </Link>
                    </div>
                )}

                {/* Queue with Theatre Mode */}
                {!isLoading && !isError && queueItems.length > 0 && (
                    <>
                        {/* Embedded Theatre Mode Player */}
                        <div className='mb-6'>
                            <PlaylistTheatreMode
                                title='My Queue'
                                items={playlistItems}
                                currentItemId={currentItemId}
                                onItemClick={handleItemClick}
                                onItemRemove={handleItemRemove}
                                onReorder={handleReorder}
                                onClipUpdated={handleClipUpdated}
                                onClose={() => navigate('/queue/theatre')}
                                isQueue={true}
                                contained={true}
                            />
                        </div>

                        {/* Header — compact, matches PlaylistDetail style */}
                        <div className='mb-6'>
                            <h1 className='text-xl font-semibold text-foreground mb-1 leading-tight'>
                                My Queue
                            </h1>
                            <p className='text-sm text-muted-foreground mb-3'>
                                {total} {total === 1 ? 'clip' : 'clips'} saved for later
                            </p>

                            {/* Actions row */}
                            <div className='flex flex-wrap items-center gap-2'>
                                <Button
                                    variant={shuffleEnabled ? 'primary' : 'outline'}
                                    size='sm'
                                    onClick={() => setShuffleEnabled(!shuffleEnabled)}
                                >
                                    <Shuffle className='h-4 w-4 mr-1' />
                                    Shuffle
                                </Button>
                                <Button
                                    variant={loopEnabled ? 'primary' : 'outline'}
                                    size='sm'
                                    onClick={() => setLoopEnabled(!loopEnabled)}
                                >
                                    <Repeat className='h-4 w-4 mr-1' />
                                    Loop
                                </Button>
                                <Button
                                    variant='outline'
                                    size='sm'
                                    onClick={() => setShowConvertDialog(true)}
                                >
                                    <ListPlus className='h-4 w-4 mr-1' />
                                    Convert to Playlist
                                </Button>
                                <Button
                                    variant='outline'
                                    size='sm'
                                    onClick={handleClearQueue}
                                    className='text-error-600 hover:text-error-700 hover:border-error-600'
                                >
                                    <Trash2 className='h-4 w-4 mr-1' />
                                    Clear Queue
                                </Button>
                            </div>
                        </div>
                    </>
                )}

                {/* Convert to Playlist Dialog */}
                <ConvertToPlaylistDialog
                    isOpen={showConvertDialog}
                    onClose={() => setShowConvertDialog(false)}
                    queueItemCount={total}
                />
            </Container>
        </>
    );
}
