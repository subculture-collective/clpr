import { useState, useCallback, useMemo, useEffect, useRef } from 'react';
import { cn, formatDuration } from '@/lib/utils';
import { TheatreMode, VideoPlayer } from '@/components/video';
import {
    Play,
    X,
    GripVertical,
    SkipForward,
    SkipBack,
    Minimize2,
    Maximize2,
    ChevronLeft,
    Check,
} from 'lucide-react';
import { Button } from '@/components/ui';
import { Tab } from '@/components/ui/Tab';
import { CommentSection } from '@/components/comment/CommentSection';
import type { Clip } from '@/types/clip';

export interface PlaylistItem {
    id: string;
    clip?: Clip;
    clip_id: string;
    order?: number;
    played_at?: string;
}

interface PlaylistTheatreModeProps {
    title: string;
    items: PlaylistItem[];
    currentItemId: string | null;
    onItemClick: (item: PlaylistItem) => void;
    onItemRemove?: (itemId: string) => void;
    onReorder?: (itemId: string, newPosition: number) => void;
    onClipUpdated?: (clipId: string) => void;
    onClose?: () => void;
    isQueue?: boolean; // True for queue, false for playlist
    contained?: boolean; // True for embedded mode, false for full-screen
    className?: string;
}

export function PlaylistTheatreMode({
    title,
    items,
    currentItemId,
    onItemClick,
    onItemRemove,
    onReorder,
    onClipUpdated,
    onClose,
    isQueue = false,
    contained = false,
    className,
}: PlaylistTheatreModeProps) {
    const [draggedId, setDraggedId] = useState<string | null>(null);
    const [dragOverId, setDragOverId] = useState<string | null>(null);
    const [showSidebar, setShowSidebar] = useState(true);
    const [activeTab, setActiveTab] = useState<'queue' | 'chat'>('queue');
    // Find current item and clip
    const currentItem = useMemo(
        () => items.find(item => item.id === currentItemId),
        [items, currentItemId],
    );
    const currentClip = currentItem?.clip;

    // Auto-advance to the next clip when current one finishes
    const handleClipEnd = useCallback(() => {
        const currentIndex = items.findIndex(item => item.id === currentItemId);
        if (currentIndex === -1) return;

        const nextItem = items[currentIndex + 1];
        if (nextItem) {
            onItemClick(nextItem);
        }
    }, [items, currentItemId, onItemClick]);

    // Skip to next clip
    const handleSkipNext = useCallback(() => {
        const currentIndex = items.findIndex(item => item.id === currentItemId);
        if (currentIndex < items.length - 1) {
            onItemClick(items[currentIndex + 1]);
        }
    }, [items, currentItemId, onItemClick]);

    // Skip to previous clip
    const handleSkipPrev = useCallback(() => {
        const currentIndex = items.findIndex(item => item.id === currentItemId);
        if (currentIndex > 0) {
            onItemClick(items[currentIndex - 1]);
        }
    }, [items, currentItemId, onItemClick]);

    // Drag and drop handlers
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

            if (!draggedId || !onReorder || draggedId === targetId) {
                setDraggedId(null);
                setDragOverId(null);
                return;
            }

            const draggedIndex = items.findIndex(item => item.id === draggedId);
            const targetIndex = items.findIndex(item => item.id === targetId);

            if (draggedIndex === -1 || targetIndex === -1) {
                setDraggedId(null);
                setDragOverId(null);
                return;
            }

            onReorder(draggedId, targetIndex);

            setDraggedId(null);
            setDragOverId(null);
        },
        [draggedId, items, onReorder],
    );

    // Keyboard shortcuts
    useEffect(() => {
        const handleKeyDown = (e: KeyboardEvent) => {
            if (e.key === 'ArrowDown' || e.key === 'n') {
                e.preventDefault();
                handleSkipNext();
            }
            if (e.key === 'ArrowUp' || e.key === 'p') {
                e.preventDefault();
                handleSkipPrev();
            }
            if (e.key === 's') {
                e.preventDefault();
                setShowSidebar(prev => !prev);
            }
            if (e.key === 'c') {
                e.preventDefault();
                setActiveTab('chat');
            }
            if (e.key === 'q') {
                e.preventDefault();
                setActiveTab('queue');
            }
        };

        window.addEventListener('keydown', handleKeyDown);
        return () => window.removeEventListener('keydown', handleKeyDown);
    }, [handleSkipNext, handleSkipPrev]);

    // Auto-advance countdown for Twitch iframe clips (no onEnded event available)
    // Shows a visible countdown the user can pause/resume
    const [countdown, setCountdown] = useState<number | null>(null);
    const [countdownPaused, setCountdownPaused] = useState(false);
    const countdownIntervalRef = useRef<ReturnType<typeof setInterval> | null>(null);
    const isIframeClip = currentClip && !currentClip.video_url && !!currentClip.duration;
    const hasNextClip = (() => {
        const idx = items.findIndex(item => item.id === currentItemId);
        return idx >= 0 && idx < items.length - 1;
    })();

    // Reset countdown when clip changes
    useEffect(() => {
        if (countdownIntervalRef.current) {
            clearInterval(countdownIntervalRef.current);
            countdownIntervalRef.current = null;
        }
        setCountdownPaused(false);

        if (!isIframeClip || !hasNextClip) {
            setCountdown(null);
            return;
        }

        // Start countdown from clip duration + 3s buffer
        const totalSeconds = Math.ceil(currentClip!.duration! + 3);
        setCountdown(totalSeconds);

        return () => {
            if (countdownIntervalRef.current) {
                clearInterval(countdownIntervalRef.current);
            }
        };
    }, [currentClip?.id]); // eslint-disable-line react-hooks/exhaustive-deps

    // Tick the countdown every second
    useEffect(() => {
        if (countdownIntervalRef.current) {
            clearInterval(countdownIntervalRef.current);
            countdownIntervalRef.current = null;
        }

        if (countdown === null || countdown <= 0 || countdownPaused) {
            if (countdown !== null && countdown <= 0 && !countdownPaused) {
                handleClipEnd();
                setCountdown(null);
            }
            return;
        }

        countdownIntervalRef.current = setInterval(() => {
            setCountdown(prev => {
                if (prev === null || prev <= 1) return 0;
                return prev - 1;
            });
        }, 1000);

        return () => {
            if (countdownIntervalRef.current) {
                clearInterval(countdownIntervalRef.current);
            }
        };
    }, [countdown, countdownPaused, handleClipEnd]);

    // Pause countdown when tab is hidden
    useEffect(() => {
        const handleVisibility = () => {
            if (document.hidden && countdown !== null && !countdownPaused) {
                setCountdownPaused(true);
            }
        };
        document.addEventListener('visibilitychange', handleVisibility);
        return () => document.removeEventListener('visibilitychange', handleVisibility);
    }, [countdown, countdownPaused]);

    return (
        <div
            className={cn(
                contained ?
                    'relative w-full bg-background rounded-xl overflow-hidden border border-border'
                :   'fixed inset-0 z-50 bg-black',
                className,
            )}
        >
            {/* Main container */}
            <div
                className={cn(
                    'relative w-full flex',
                    contained ? 'h-[600px]' : 'h-full',
                )}
            >
                {/* Video player area */}
                <div className='flex-1 flex flex-col items-center justify-center transition-all'>
                    {/* Video player */}
                    <div className='w-full flex-1 flex items-center justify-center overflow-hidden'>
                        {currentClip ?
                            currentClip.video_url ?
                                <TheatreMode
                                    title={currentClip.title}
                                    hlsUrl={currentClip.video_url}
                                    onEnded={handleClipEnd}
                                    fit='height'
                                />
                            :   <VideoPlayer
                                    clipId={currentClip.id}
                                    title={currentClip.title}
                                    embedUrl={currentClip.embed_url}
                                    twitchClipId={currentClip.twitch_clip_id}
                                    fit='height'
                                    className='max-h-full'
                                    onEnded={handleClipEnd}
                                    onPlay={() => setCountdownPaused(false)}
                                    onPause={() => setCountdownPaused(true)}
                                />

                        :   <div className='text-center text-text-secondary'>
                                <p className='text-lg'>No clip selected</p>
                                <p className='text-sm mt-2'>
                                    Select a clip from the{' '}
                                    {isQueue ? 'queue' : 'playlist'} to start
                                    watching
                                </p>
                            </div>
                        }
                    </div>

                    {/* Auto-advance countdown bar for Twitch iframe clips */}
                    {isIframeClip && countdown !== null && hasNextClip && (
                        <div className='w-full px-4 py-1.5 flex items-center justify-center gap-3 bg-surface/80 text-[12px]'>
                            {countdownPaused ? (
                                <>
                                    <span className='text-text-secondary'>
                                        Auto-advance paused
                                    </span>
                                    <button
                                        onClick={() => setCountdownPaused(false)}
                                        className='text-cta hover:text-cta-hover transition-colors cursor-pointer font-medium'
                                    >
                                        Resume
                                    </button>
                                    <button
                                        onClick={() => setCountdown(null)}
                                        className='text-text-tertiary hover:text-text-secondary transition-colors cursor-pointer'
                                    >
                                        Cancel
                                    </button>
                                </>
                            ) : (
                                <>
                                    <span className='text-text-secondary'>
                                        Next clip in {countdown}s
                                    </span>
                                    <button
                                        onClick={() => setCountdownPaused(true)}
                                        className='text-cta hover:text-cta-hover transition-colors cursor-pointer font-medium'
                                    >
                                        Pause
                                    </button>
                                    <button
                                        onClick={() => handleClipEnd()}
                                        className='text-text-tertiary hover:text-text-secondary transition-colors cursor-pointer'
                                    >
                                        Skip now
                                    </button>
                                </>
                            )}
                        </div>
                    )}
                </div>

                {/* Sidebar toggle when hidden */}
                {!showSidebar && (
                    <button
                        onClick={() => setShowSidebar(true)}
                        className='absolute top-4 right-4 z-10 p-2 bg-surface/80 hover:bg-surface-hover border border-border rounded-lg transition-colors cursor-pointer'
                        aria-label='Show playlist'
                    >
                        <ChevronLeft className='h-5 w-5 text-text-primary' />
                    </button>
                )}

                {/* Playlist/Queue sidebar */}
                {showSidebar && (
                    <div
                        className={cn(
                            'h-full bg-surface border-l border-border flex flex-col',
                            contained ? 'w-80' : 'w-96',
                        )}
                    >
                        {/* Sidebar header */}
                        <div className='px-4 py-2.5 border-b border-border'>
                            <div className='flex items-center justify-between'>
                                <div className='flex items-center gap-2'>
                                    <p className='text-text-secondary text-xs'>
                                        <span className='font-medium text-text-primary'>
                                            {items.findIndex(
                                                item => item.id === currentItemId,
                                            ) + 1}
                                        </span>
                                        {' / '}
                                        {items.length}{' '}
                                        {items.length === 1 ? 'clip' : 'clips'}
                                    </p>
                                </div>
                                <button
                                    onClick={() => setShowSidebar(false)}
                                    className='p-1.5 hover:bg-surface-hover rounded-lg transition-colors cursor-pointer'
                                    aria-label='Hide sidebar'
                                >
                                    <ChevronLeft className='h-4 w-4 text-text-primary rotate-180' />
                                </button>
                            </div>
                        </div>

                        <Tab
                            tabs={[
                                { id: 'queue', label: isQueue ? 'Queue' : 'Playlist' },
                                { id: 'chat', label: 'Chat', badge: currentClip?.comment_count },
                            ]}
                            activeTab={activeTab}
                            onTabChange={(id) => setActiveTab(id as 'queue' | 'chat')}
                        />

                        {/* Scrollable items list */}
                        {activeTab === 'queue' && (
                        <div className='flex-1 overflow-y-auto'>
                            {items.map((item, idx) => {
                                const isCurrentItem = item.id === currentItemId;
                                const isPlayed = !!item.played_at;

                                return (
                                    <div
                                        key={item.id}
                                        draggable={!!onReorder}
                                        onDragStart={() =>
                                            handleDragStart(item.id)
                                        }
                                        onDragOver={e =>
                                            handleDragOver(e, item.id)
                                        }
                                        onDragLeave={handleDragLeave}
                                        onDrop={e => handleDrop(e, item.id)}
                                        className={cn(
                                            'group relative border-b border-border hover:bg-surface-hover transition-colors',
                                            isCurrentItem &&
                                                'bg-primary-500/20',
                                            draggedId === item.id &&
                                                'opacity-50',
                                            dragOverId === item.id &&
                                                'border-t-2 border-primary-500',
                                        )}
                                    >
                                        <div className='flex gap-2 p-3'>
                                            {/* Drag handle and number */}
                                            <div className='flex items-center gap-2 text-text-tertiary'>
                                                {onReorder && (
                                                    <GripVertical className='h-4 w-4 cursor-grab active:cursor-grabbing' />
                                                )}
                                                <span className='text-xs font-mono w-6 text-right'>
                                                    {idx + 1}
                                                </span>
                                            </div>

                                            {/* Thumbnail */}
                                            <button
                                                onClick={() =>
                                                    onItemClick(item)
                                                }
                                                className='relative w-20 h-12 shrink-0 rounded overflow-hidden group/thumb'
                                            >
                                                {item.clip?.thumbnail_url && (
                                                    <img
                                                        src={
                                                            item.clip
                                                                .thumbnail_url
                                                        }
                                                        alt={item.clip.title}
                                                        className='w-full h-full object-cover'
                                                    />
                                                )}
                                                <div className='absolute inset-0 bg-black/50 opacity-0 group-hover/thumb:opacity-100 transition-opacity flex items-center justify-center'>
                                                    <Play className='h-5 w-5 text-white fill-white' />
                                                </div>
                                                {item.clip?.duration && (
                                                    <div className='absolute bottom-0.5 right-0.5 px-1 py-0.5 text-xs text-white bg-black/75 rounded'>
                                                        {formatDuration(
                                                            item.clip.duration,
                                                        )}
                                                    </div>
                                                )}
                                                {isCurrentItem && (
                                                    <div className='absolute top-0.5 left-0.5 p-1 bg-primary-500 rounded'>
                                                        <Play className='h-3 w-3 text-white fill-white' />
                                                    </div>
                                                )}
                                            </button>

                                            {/* Info */}
                                            <div className='flex-1 min-w-0'>
                                                <button
                                                    onClick={() =>
                                                        onItemClick(item)
                                                    }
                                                    className='text-left w-full'
                                                >
                                                    <p
                                                        className={cn(
                                                            'text-sm font-medium line-clamp-2',
                                                            isCurrentItem ?
                                                                'text-primary-400'
                                                            :   'text-text-primary',
                                                            isPlayed &&
                                                                'opacity-60',
                                                        )}
                                                    >
                                                        {item.clip?.title ||
                                                            'Unknown Clip'}
                                                    </p>
                                                    <p className='text-xs text-text-secondary mt-0.5'>
                                                        {
                                                            item.clip
                                                                ?.broadcaster_name
                                                        }
                                                    </p>
                                                </button>
                                                {isPlayed && (
                                                    <div className='flex items-center gap-1 text-xs text-text-tertiary mt-1'>
                                                        <Check className='h-3 w-3' />
                                                        Watched
                                                    </div>
                                                )}
                                            </div>

                                            {/* Actions */}
                                            {onItemRemove && (
                                                <button
                                                    onClick={() =>
                                                        onItemRemove(item.id)
                                                    }
                                                    className='p-1.5 opacity-0 group-hover:opacity-100 hover:bg-error-500/20 text-error-400 rounded transition-all'
                                                    aria-label='Remove from playlist'
                                                >
                                                    <X className='h-4 w-4' />
                                                </button>
                                            )}
                                        </div>
                                    </div>
                                );
                            })}

                            {items.length === 0 && (
                                <div className='text-center py-12 text-text-tertiary'>
                                    <p>
                                        No clips in{' '}
                                        {isQueue ? 'queue' : 'playlist'}
                                    </p>
                                </div>
                            )}
                        </div>
                        )}

                        {activeTab === 'chat' && currentClip && (
                            <div className="flex-1 overflow-y-auto">
                                <CommentSection
                                    clipId={currentClip.id}
                                    className="p-3"
                                />
                            </div>
                        )}
                        {activeTab === 'chat' && !currentClip && (
                            <div className="flex-1 flex items-center justify-center text-text-tertiary text-sm">
                                Select a clip to see comments
                            </div>
                        )}

                        {/* Controls + keyboard shortcuts */}
                        <div className='border-t border-border bg-background/50'>
                            {/* Prev / Next / Fullscreen buttons with inline key hints */}
                            <div className='flex items-center justify-between px-3 py-2'>
                                <div className='flex items-center gap-1'>
                                    <Button
                                        variant='ghost'
                                        size='sm'
                                        onClick={handleSkipPrev}
                                        disabled={
                                            items.findIndex(
                                                item => item.id === currentItemId,
                                            ) === 0
                                        }
                                        className='text-text-primary hover:bg-surface-hover'
                                    >
                                        <SkipBack className='h-4 w-4 mr-1' />
                                        Prev
                                        <kbd className='ml-1.5 px-1 py-0.5 text-[10px] bg-surface-raised rounded text-text-tertiary'>
                                            P
                                        </kbd>
                                    </Button>
                                    <Button
                                        variant='ghost'
                                        size='sm'
                                        onClick={handleSkipNext}
                                        disabled={
                                            items.findIndex(
                                                item => item.id === currentItemId,
                                            ) ===
                                            items.length - 1
                                        }
                                        className='text-text-primary hover:bg-surface-hover'
                                    >
                                        <SkipForward className='h-4 w-4 mr-1' />
                                        Next
                                        <kbd className='ml-1.5 px-1 py-0.5 text-[10px] bg-surface-raised rounded text-text-tertiary'>
                                            N
                                        </kbd>
                                    </Button>
                                </div>
                                {onClose && (
                                    <Button
                                        variant='ghost'
                                        size='sm'
                                        onClick={onClose}
                                        className='text-text-primary hover:bg-surface-hover'
                                    >
                                        {contained ?
                                            <>
                                                <Maximize2 className='h-4 w-4 mr-1' />
                                                Full
                                            </>
                                        :   <>
                                                <Minimize2 className='h-4 w-4 mr-1' />
                                                Exit
                                            </>
                                        }
                                    </Button>
                                )}
                            </div>
                            {/* Key hints */}
                            <div className='px-3 pb-2'>
                                <p className='text-xs text-text-tertiary'>
                                    <kbd className='px-1.5 py-0.5 bg-surface-raised rounded text-text-secondary'>
                                        S
                                    </kbd>{' '}
                                    Sidebar
                                    {' • '}
                                    <kbd className='px-1.5 py-0.5 bg-surface-raised rounded text-text-secondary'>
                                        Q
                                    </kbd>{' '}
                                    Queue
                                    {' • '}
                                    <kbd className='px-1.5 py-0.5 bg-surface-raised rounded text-text-secondary'>
                                        C
                                    </kbd>{' '}
                                    Chat
                                </p>
                            </div>
                        </div>
                    </div>
                )}
            </div>
        </div>
    );
}
