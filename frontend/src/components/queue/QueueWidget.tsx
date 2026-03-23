import { useState, useCallback } from 'react';
import { Link, useLocation } from 'react-router-dom';
import { VideoPlayer } from '@/components/video';
import {
    useQueue,
    useRemoveFromQueue,
    useQueueCount,
    useReorderQueue,
} from '@/hooks/useQueue';
import { useAuth } from '@/context/AuthContext';
import { formatDuration, cn } from '@/lib/utils';
import {
    X,
    ChevronUp,
    ChevronDown,
    Play,
    SkipForward,
    SkipBack,
    Maximize2,
    ExternalLink,
    ListMusic,
    Trash2,
    GripVertical,
    Repeat,
    Shuffle,
} from 'lucide-react';
import type { QueueItemWithClip } from '@/types/queue';

type WidgetState = 'collapsed' | 'expanded' | 'playing';

export function QueueWidget() {
    const { user } = useAuth();
    const isAuthenticated = !!user;
    const { data: queue } = useQueue(20, isAuthenticated);
    const { data: queueCount } = useQueueCount(isAuthenticated);
    const removeFromQueue = useRemoveFromQueue();
    const reorderQueue = useReorderQueue();

    const [widgetState, setWidgetState] = useState<WidgetState>('collapsed');
    const [draggedId, setDraggedId] = useState<string | null>(null);
    const [dragOverId, setDragOverId] = useState<string | null>(null);
    const [currentClip, setCurrentClip] = useState<QueueItemWithClip | null>(null);
    const [currentItemId, setCurrentItemId] = useState<string | null>(null);
    const [loopEnabled, setLoopEnabled] = useState(false);
    const [shuffleEnabled, setShuffleEnabled] = useState(false);

    const queueItems = queue?.items || [];

    // Compute next/prev availability
    const currentIndex = queueItems.findIndex(item => item.id === currentItemId);
    const hasNext = currentIndex >= 0 && currentIndex < queueItems.length - 1;
    const hasPrev = currentIndex > 0;

    const handlePlayClip = useCallback(
        (item: QueueItemWithClip) => {
            setCurrentClip(item);
            setCurrentItemId(item.id);
            setWidgetState('playing');
        },
        [],
    );

    const handlePlayNext = useCallback(() => {
        const items = queue?.items || [];
        const idx = items.findIndex(item => item.id === currentItemId);

        if (shuffleEnabled) {
            // Pick a random different clip
            const otherItems = items.filter(item => item.id !== currentItemId);
            if (otherItems.length > 0) {
                const randomItem = otherItems[Math.floor(Math.random() * otherItems.length)];
                setCurrentClip(randomItem);
                setCurrentItemId(randomItem.id);
            }
            return;
        }

        const nextItem = items[idx + 1];
        if (nextItem) {
            setCurrentClip(nextItem);
            setCurrentItemId(nextItem.id);
        } else if (loopEnabled && items.length > 0) {
            // Loop back to first clip
            setCurrentClip(items[0]);
            setCurrentItemId(items[0].id);
        }
    }, [queue?.items, currentItemId, loopEnabled, shuffleEnabled]);

    const handlePlayPrev = useCallback(() => {
        const items = queue?.items || [];
        const idx = items.findIndex(item => item.id === currentItemId);
        const prevItem = items[idx - 1];
        if (prevItem) {
            setCurrentClip(prevItem);
            setCurrentItemId(prevItem.id);
        }
    }, [queue?.items, currentItemId]);

    const handleRemoveItem = useCallback(
        (itemId: string) => {
            removeFromQueue.mutate(itemId);
            if (itemId === currentItemId) {
                handlePlayNext();
            }
        },
        [removeFromQueue, currentItemId, handlePlayNext],
    );

    const handleClose = useCallback(() => {
        setWidgetState('collapsed');
    }, []);

    const handleExpand = useCallback(() => {
        setWidgetState('expanded');
    }, []);

    const handleMinimizeToPlayer = useCallback(() => {
        if (currentClip) {
            setWidgetState('playing');
        } else {
            setWidgetState('collapsed');
        }
    }, [currentClip]);

    const handleDragStart = useCallback((id: string) => {
        setDraggedId(id);
    }, []);

    const handleDragOver = useCallback((e: React.DragEvent, id: string) => {
        e.preventDefault();
        setDragOverId(id);
    }, []);

    const handleDragLeave = useCallback(() => {
        setDragOverId(null);
    }, []);

    const handleDrop = useCallback(
        (e: React.DragEvent, targetId: string) => {
            e.preventDefault();
            if (!draggedId || draggedId === targetId) {
                setDraggedId(null);
                setDragOverId(null);
                return;
            }
            const targetIndex = queueItems.findIndex(item => item.id === targetId);
            if (targetIndex !== -1) {
                reorderQueue.mutate({ item_id: draggedId, new_position: targetIndex });
            }
            setDraggedId(null);
            setDragOverId(null);
        },
        [draggedId, queueItems, reorderQueue],
    );

    const location = useLocation();

    // Don't show on queue pages, for logged out users, or empty queues
    if (location.pathname.startsWith('/queue')) {
        return null;
    }
    if (!user || !queueCount || queueCount === 0) {
        return null;
    }

    // Collapsed state - small button
    if (widgetState === 'collapsed') {
        return (
            <button
                onClick={handleExpand}
                className='fixed bottom-6 right-6 z-50 flex items-center gap-2 px-4 py-3 bg-primary-600 hover:bg-primary-700 text-white rounded-full shadow-lg transition-all hover:scale-105 cursor-pointer'
                aria-label='Open queue'
            >
                <ListMusic className='h-5 w-5' />
                <span className='font-medium'>{queueCount}</span>
            </button>
        );
    }

    // Playing state - miniplayer with queue
    if (widgetState === 'playing' && currentClip?.clip) {
        return (
            <div className='fixed bottom-6 right-6 z-50 w-80 bg-card border border-border rounded-xl shadow-2xl overflow-hidden'>
                {/* Header */}
                <div className='flex items-center justify-between px-3 py-2 bg-muted/50 border-b border-border'>
                    <div className='flex items-center gap-2'>
                        <ListMusic className='h-4 w-4 text-muted-foreground' />
                        <span className='text-sm font-medium'>Now Playing</span>
                    </div>
                    <div className='flex items-center gap-1'>
                        <Link
                            to='/queue'
                            onClick={handleClose}
                            className='p-1 hover:bg-muted rounded transition-colors cursor-pointer'
                            aria-label='Open queue page'
                            title='Queue page'
                        >
                            <Maximize2 className='h-4 w-4' />
                        </Link>
                        <button
                            onClick={handleExpand}
                            className='p-1 hover:bg-muted rounded transition-colors cursor-pointer'
                            aria-label='Show queue list'
                        >
                            <ChevronDown className='h-4 w-4' />
                        </button>
                        <button
                            onClick={handleClose}
                            className='p-1 hover:bg-muted rounded transition-colors cursor-pointer'
                            aria-label='Close player'
                        >
                            <X className='h-4 w-4' />
                        </button>
                    </div>
                </div>

                {/* Miniplayer */}
                <VideoPlayer
                    clipId={currentClip.clip.id}
                    title={currentClip.clip.title}
                    embedUrl={currentClip.clip.embed_url}
                    twitchClipId={currentClip.clip.twitch_clip_id}
                    onEnded={handlePlayNext}
                />

                {/* Clip Info */}
                <div className='p-3'>
                    <Link
                        to={`/clip/${currentClip.clip_id}`}
                        className='font-medium text-sm line-clamp-1 hover:text-brand transition-colors cursor-pointer'
                    >
                        {currentClip.clip.title}
                    </Link>
                    <p className='text-xs text-muted-foreground mt-0.5 line-clamp-1'>
                        <Link
                            to={`/broadcaster/${currentClip.clip.broadcaster_id || currentClip.clip.broadcaster_name}`}
                            className='hover:text-foreground transition-colors cursor-pointer'
                        >
                            {currentClip.clip.broadcaster_name}
                        </Link>
                        {currentClip.clip.game_name &&
                            ` • ${currentClip.clip.game_name}`}
                    </p>

                    {/* Playback controls */}
                    <div className='flex items-center justify-between mt-3'>
                        <div className='flex items-center gap-1'>
                            <button
                                onClick={() => setShuffleEnabled(!shuffleEnabled)}
                                className={cn(
                                    'p-1.5 rounded transition-colors cursor-pointer',
                                    shuffleEnabled ? 'text-brand bg-brand/10' : 'text-muted-foreground hover:text-foreground',
                                )}
                                aria-label={shuffleEnabled ? 'Disable shuffle' : 'Enable shuffle'}
                                title='Shuffle'
                            >
                                <Shuffle className='h-3.5 w-3.5' />
                            </button>
                            <button
                                onClick={handlePlayPrev}
                                disabled={!hasPrev}
                                className='p-1.5 rounded transition-colors cursor-pointer text-muted-foreground hover:text-foreground disabled:opacity-30 disabled:cursor-not-allowed'
                                aria-label='Previous clip'
                            >
                                <SkipBack className='h-3.5 w-3.5' />
                            </button>
                            <button
                                onClick={handlePlayNext}
                                disabled={!hasNext && !loopEnabled && !shuffleEnabled}
                                className='p-1.5 rounded transition-colors cursor-pointer text-muted-foreground hover:text-foreground disabled:opacity-30 disabled:cursor-not-allowed'
                                aria-label='Next clip'
                            >
                                <SkipForward className='h-3.5 w-3.5' />
                            </button>
                            <button
                                onClick={() => setLoopEnabled(!loopEnabled)}
                                className={cn(
                                    'p-1.5 rounded transition-colors cursor-pointer',
                                    loopEnabled ? 'text-brand bg-brand/10' : 'text-muted-foreground hover:text-foreground',
                                )}
                                aria-label={loopEnabled ? 'Disable loop' : 'Enable loop'}
                                title='Loop'
                            >
                                <Repeat className='h-3.5 w-3.5' />
                            </button>
                        </div>
                        <Link
                            to={`/clip/${currentClip.clip_id}`}
                            className='flex items-center gap-1 px-2 py-1 text-xs text-muted-foreground hover:text-foreground rounded transition-colors cursor-pointer'
                            aria-label='View clip page'
                        >
                            <ExternalLink className='h-3 w-3' />
                            Clip
                        </Link>
                    </div>
                </div>

                {/* Queue list — full style with thumbnails */}
                <div className='border-t border-border max-h-[250px] overflow-y-auto'>
                    {queueItems.map((item) => {
                        const isCurrentItem = item.id === currentItemId;
                        return (
                            <div
                                key={item.id}
                                draggable
                                onDragStart={() => handleDragStart(item.id)}
                                onDragOver={(e) => handleDragOver(e, item.id)}
                                onDragLeave={handleDragLeave}
                                onDrop={(e) => handleDrop(e, item.id)}
                                className={cn(
                                    'flex items-center gap-2 p-2 hover:bg-muted/50 transition-colors group',
                                    isCurrentItem && 'bg-brand/10 border-l-2 border-brand',
                                    draggedId === item.id && 'opacity-50',
                                    dragOverId === item.id && 'border-t-2 border-brand',
                                )}
                            >
                                <GripVertical className='h-3 w-3 text-muted-foreground cursor-grab active:cursor-grabbing shrink-0' />
                                {isCurrentItem ? (
                                    <Play className='h-3 w-3 text-brand fill-brand shrink-0' />
                                ) : null}

                                {/* Thumbnail */}
                                <div
                                    onClick={() => handlePlayClip(item)}
                                    className='relative w-14 h-9 shrink-0 rounded overflow-hidden cursor-pointer group/thumb'
                                >
                                    {item.clip?.thumbnail_url && (
                                        <img
                                            src={item.clip.thumbnail_url}
                                            alt=''
                                            className='w-full h-full object-cover'
                                        />
                                    )}
                                    <div className='absolute inset-0 bg-black/40 opacity-0 group-hover/thumb:opacity-100 flex items-center justify-center transition-opacity'>
                                        <Play className='h-4 w-4 text-white fill-white' />
                                    </div>
                                    {item.clip?.duration && (
                                        <span className='absolute bottom-0.5 right-0.5 text-[10px] bg-black/75 text-white px-0.5 rounded'>
                                            {formatDuration(item.clip.duration)}
                                        </span>
                                    )}
                                </div>

                                {/* Clip info */}
                                <div className='flex-1 min-w-0'>
                                    <p className={cn(
                                        'text-xs font-medium line-clamp-1',
                                        isCurrentItem && 'text-brand',
                                    )}>
                                        {item.clip?.title || 'Unknown Clip'}
                                    </p>
                                    <Link
                                        to={`/broadcaster/${item.clip?.broadcaster_id || item.clip?.broadcaster_name}`}
                                        className='text-[10px] text-muted-foreground hover:text-foreground transition-colors line-clamp-1 cursor-pointer'
                                        onClick={(e) => e.stopPropagation()}
                                    >
                                        {item.clip?.broadcaster_name}
                                    </Link>
                                </div>

                                {/* Remove */}
                                <button
                                    onClick={() => handleRemoveItem(item.id)}
                                    className='p-0.5 opacity-0 group-hover:opacity-100 hover:text-error-600 rounded transition-all cursor-pointer text-muted-foreground'
                                    aria-label='Remove from queue'
                                >
                                    <X className='h-3 w-3' />
                                </button>
                            </div>
                        );
                    })}
                </div>
            </div>
        );
    }

    // Expanded state - full queue list
    return (
        <div className='fixed bottom-6 right-6 z-50 w-80 max-h-[70vh] bg-card border border-border rounded-xl shadow-2xl overflow-hidden flex flex-col'>
            {/* Header */}
            <div className='flex items-center justify-between px-3 py-2 bg-muted/50 border-b border-border shrink-0'>
                <div className='flex items-center gap-2'>
                    <ListMusic className='h-4 w-4 text-muted-foreground' />
                    <span className='text-sm font-medium'>Queue</span>
                    <span className='text-xs text-muted-foreground'>
                        ({queueCount})
                    </span>
                </div>
                <div className='flex items-center gap-1'>
                    <Link
                        to='/queue'
                        onClick={handleClose}
                        className='p-1 hover:bg-muted rounded transition-colors cursor-pointer'
                        aria-label='Open queue page'
                        title='Queue page'
                    >
                        <Maximize2 className='h-4 w-4' />
                    </Link>
                    <button
                        onClick={() => {
                            if (currentClip) {
                                setWidgetState('playing');
                            } else if (queueItems.length > 0) {
                                handlePlayClip(queueItems[0]);
                            }
                        }}
                        className='p-1 hover:bg-muted rounded transition-colors cursor-pointer'
                        aria-label='Show player'
                        title='Show player'
                    >
                        <ChevronUp className='h-4 w-4' />
                    </button>
                    <button
                        onClick={handleClose}
                        className='p-1 hover:bg-muted rounded transition-colors cursor-pointer'
                        aria-label='Close queue'
                    >
                        <X className='h-4 w-4' />
                    </button>
                </div>
            </div>

            {/* Playback mode controls */}
            <div className='flex items-center justify-between px-3 py-1.5 border-b border-border bg-muted/20'>
                <div className='flex items-center gap-1'>
                    <button
                        onClick={() => setShuffleEnabled(!shuffleEnabled)}
                        className={cn(
                            'p-1 rounded text-xs flex items-center gap-1 transition-colors cursor-pointer',
                            shuffleEnabled ? 'text-brand bg-brand/10' : 'text-muted-foreground hover:text-foreground',
                        )}
                        title='Shuffle'
                    >
                        <Shuffle className='h-3 w-3' />
                        <span className='text-[10px]'>Shuffle</span>
                    </button>
                    <button
                        onClick={() => setLoopEnabled(!loopEnabled)}
                        className={cn(
                            'p-1 rounded text-xs flex items-center gap-1 transition-colors cursor-pointer',
                            loopEnabled ? 'text-brand bg-brand/10' : 'text-muted-foreground hover:text-foreground',
                        )}
                        title='Loop'
                    >
                        <Repeat className='h-3 w-3' />
                        <span className='text-[10px]'>Loop</span>
                    </button>
                </div>
            </div>

            {/* Queue List — full thumbnail style */}
            <div className='flex-1 overflow-y-auto'>
                {queueItems.length === 0 ?
                    <div className='p-6 text-center text-muted-foreground'>
                        <ListMusic className='h-8 w-8 mx-auto mb-2 opacity-50' />
                        <p className='text-sm'>Your queue is empty</p>
                    </div>
                :   <div>
                        {queueItems.map((item) => {
                            const isCurrentItem = item.id === currentItemId;
                            return (
                                <div
                                    key={item.id}
                                    draggable
                                    onDragStart={() => handleDragStart(item.id)}
                                    onDragOver={(e) => handleDragOver(e, item.id)}
                                    onDragLeave={handleDragLeave}
                                    onDrop={(e) => handleDrop(e, item.id)}
                                    className={cn(
                                        'flex items-center gap-2 p-2 hover:bg-muted/50 transition-colors group',
                                        isCurrentItem && 'bg-brand/10 border-l-2 border-brand',
                                        draggedId === item.id && 'opacity-50',
                                        dragOverId === item.id && 'border-t-2 border-brand',
                                    )}
                                >
                                    <GripVertical className='h-3 w-3 text-muted-foreground cursor-grab active:cursor-grabbing shrink-0' />
                                    {isCurrentItem ? (
                                        <Play className='h-3 w-3 text-brand fill-brand shrink-0' />
                                    ) : null}

                                    {/* Thumbnail */}
                                    <div
                                        onClick={() => handlePlayClip(item)}
                                        className='relative w-14 h-9 shrink-0 rounded overflow-hidden cursor-pointer group/thumb'
                                    >
                                        {item.clip?.thumbnail_url && (
                                            <img
                                                src={item.clip.thumbnail_url}
                                                alt=''
                                                className='w-full h-full object-cover'
                                            />
                                        )}
                                        <div className='absolute inset-0 bg-black/40 opacity-0 group-hover/thumb:opacity-100 flex items-center justify-center transition-opacity'>
                                            <Play className='h-4 w-4 text-white fill-white' />
                                        </div>
                                        {item.clip?.duration && (
                                            <span className='absolute bottom-0.5 right-0.5 text-[10px] bg-black/75 text-white px-0.5 rounded'>
                                                {formatDuration(item.clip.duration)}
                                            </span>
                                        )}
                                    </div>

                                    {/* Clip info */}
                                    <div className='flex-1 min-w-0'>
                                        <p className={cn(
                                            'text-xs font-medium line-clamp-1',
                                            isCurrentItem && 'text-brand',
                                        )}>
                                            {item.clip?.title || 'Unknown Clip'}
                                        </p>
                                        <Link
                                            to={`/broadcaster/${item.clip?.broadcaster_id || item.clip?.broadcaster_name}`}
                                            className='text-[10px] text-muted-foreground hover:text-foreground transition-colors line-clamp-1 cursor-pointer'
                                            onClick={(e) => e.stopPropagation()}
                                        >
                                            {item.clip?.broadcaster_name}
                                        </Link>
                                    </div>

                                    {/* Remove */}
                                    <button
                                        onClick={() => handleRemoveItem(item.id)}
                                        className='p-0.5 opacity-0 group-hover:opacity-100 hover:text-error-600 rounded transition-all cursor-pointer text-muted-foreground'
                                        aria-label='Remove from queue'
                                    >
                                        <X className='h-3 w-3' />
                                    </button>
                                </div>
                            );
                        })}
                    </div>
                }
            </div>

            {/* Footer */}
            <div className='shrink-0 border-t border-border p-2 flex items-center justify-between bg-muted/30'>
                <Link
                    to='/queue'
                    onClick={handleClose}
                    className='text-xs text-primary-600 hover:text-primary-700 hover:underline cursor-pointer'
                >
                    View full queue
                </Link>
                {queueItems.length > 0 && !currentClip && (
                    <button
                        onClick={() => handlePlayClip(queueItems[0])}
                        className='flex items-center gap-1 px-2 py-1 text-xs bg-primary-600 hover:bg-primary-700 text-white rounded transition-colors cursor-pointer'
                    >
                        <Play className='h-3 w-3 fill-current' />
                        Play All
                    </button>
                )}
            </div>
        </div>
    );
}
