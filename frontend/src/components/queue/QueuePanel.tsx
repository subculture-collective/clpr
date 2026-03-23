import { useQueue, useRemoveFromQueue, useClearQueue, useReorderQueue } from '@/hooks/useQueue';
import { formatDuration } from '@/lib/utils';
import { X, Trash2 } from 'lucide-react';
import { useState } from 'react';
import { Button } from '@/components/ui';

export function QueuePanel() {
    const { data: queue, isLoading } = useQueue(100);
    const removeFromQueue = useRemoveFromQueue();
    const clearQueue = useClearQueue();
    const reorderQueue = useReorderQueue();
    const [draggedId, setDraggedId] = useState<string | null>(null);
    const [dragOverId, setDragOverId] = useState<string | null>(null);

    const handleRemove = (itemId: string) => {
        removeFromQueue.mutate(itemId);
    };

    const handleClearQueue = () => {
        if (window.confirm('Are you sure you want to clear the entire queue?')) {
            clearQueue.mutate();
        }
    };

    const handleDragStart = (id: string) => {
        setDraggedId(id);
    };

    const handleDragOver = (e: React.DragEvent, id: string) => {
        e.preventDefault();
        setDragOverId(id);
    };

    const handleDragLeave = () => {
        setDragOverId(null);
    };

    const handleDrop = (e: React.DragEvent, targetId: string) => {
        e.preventDefault();
        
        if (!draggedId || draggedId === targetId) {
            setDraggedId(null);
            setDragOverId(null);
            return;
        }

        const items = queue?.items || [];
        const draggedIndex = items.findIndex(item => item.id === draggedId);
        const targetIndex = items.findIndex(item => item.id === targetId);

        if (draggedIndex === -1 || targetIndex === -1) {
            setDraggedId(null);
            setDragOverId(null);
            return;
        }

        // Use target position as the new position
        const newPosition = targetIndex;

        // Call the reorder API
        reorderQueue.mutate({
            item_id: draggedId,
            new_position: newPosition,
        });

        setDraggedId(null);
        setDragOverId(null);
    };

    const handleDragEnd = () => {
        setDraggedId(null);
        setDragOverId(null);
    };

    const handleKeyDown = (e: React.KeyboardEvent, itemId: string, currentIndex: number) => {
        // Only handle Alt+ArrowUp and Alt+ArrowDown for reordering
        if (!e.altKey || (e.key !== 'ArrowUp' && e.key !== 'ArrowDown')) {
            return;
        }

        e.preventDefault();

        const items = queue?.items || [];
        const newIndex = e.key === 'ArrowUp' ? currentIndex - 1 : currentIndex + 1;

        // Check bounds
        if (newIndex < 0 || newIndex >= items.length) {
            return;
        }

        // Call the reorder API with the new position
        reorderQueue.mutate({
            item_id: itemId,
            new_position: newIndex,
        });
    };

    if (isLoading) {
        return (
            <div className="w-80 bg-surface border-l border-border flex flex-col h-screen">
                <div className="p-4 border-b border-border">
                    <h2 className="text-lg font-bold text-foreground">Queue</h2>
                </div>
                <div className="flex-1 flex items-center justify-center">
                    <div className="text-text-secondary">Loading...</div>
                </div>
            </div>
        );
    }

    const queueItems = queue?.items || [];
    const total = queue?.total || 0;

    return (
        <div className="w-80 bg-surface border-l border-border flex flex-col h-screen">
            {/* Header */}
            <div className="p-4 border-b border-border flex items-center justify-between">
                <h2 className="text-lg font-bold text-foreground">
                    Queue ({total})
                </h2>
                {total > 0 && (
                    <Button
                        variant="ghost"
                        size="sm"
                        onClick={handleClearQueue}
                        className="text-red-400 hover:text-red-300 hover:bg-red-900/20"
                    >
                        <Trash2 className="h-4 w-4" />
                    </Button>
                )}
            </div>

            {/* Queue Items */}
            <div className="flex-1 overflow-y-auto">
                {queueItems.length === 0 ? (
                    <div className="p-4 text-text-secondary text-center">
                        Queue is empty
                    </div>
                ) : (
                    <div className="divide-y divide-border" role="list" aria-label="Queue items">
                        {queueItems.map((item, idx) => (
                            <div
                                key={item.id}
                                draggable
                                tabIndex={0}
                                onDragStart={() => handleDragStart(item.id)}
                                onDragOver={(e) => handleDragOver(e, item.id)}
                                onDragLeave={handleDragLeave}
                                onDrop={(e) => handleDrop(e, item.id)}
                                onDragEnd={handleDragEnd}
                                onKeyDown={(e) => handleKeyDown(e, item.id, idx)}
                                className={`p-3 hover:bg-surface-hover cursor-move transition-colors ${
                                    draggedId === item.id ? 'opacity-50 bg-surface-raised' : ''
                                } ${
                                    dragOverId === item.id ? 'border-t-2 border-primary' : ''
                                }`}
                                role="listitem"
                                aria-label={`Queue item ${idx + 1}: ${item.clip?.title || 'Unknown Clip'}. Press Alt+Arrow Up or Alt+Arrow Down to reorder.`}
                            >
                                <div className="flex gap-3">
                                    {/* Position Number */}
                                    <span className="text-text-secondary text-sm font-mono w-6 flex-shrink-0">
                                        {idx + 1}.
                                    </span>

                                    {/* Thumbnail */}
                                    {item.clip?.thumbnail_url && (
                                        <div className="w-20 h-12 flex-shrink-0 rounded overflow-hidden">
                                            <img
                                                src={item.clip.thumbnail_url}
                                                alt={item.clip.title}
                                                className="w-full h-full object-cover"
                                            />
                                        </div>
                                    )}

                                    {/* Clip Info */}
                                    <div className="flex-1 min-w-0">
                                        <p className="text-sm font-medium text-foreground truncate">
                                            {item.clip?.title || 'Unknown Clip'}
                                        </p>
                                        <div className="flex items-center gap-2 text-xs text-text-secondary mt-1">
                                            <span>{item.clip?.broadcaster_name}</span>
                                            {item.clip?.duration && (
                                                <>
                                                    <span>•</span>
                                                    <span>{formatDuration(item.clip.duration)}</span>
                                                </>
                                            )}
                                        </div>
                                    </div>

                                    {/* Remove Button */}
                                    <button
                                        onClick={() => handleRemove(item.id)}
                                        className="text-text-secondary hover:text-red-400 transition-colors flex-shrink-0"
                                        aria-label="Remove from queue"
                                    >
                                        <X className="h-4 w-4" />
                                    </button>
                                </div>
                            </div>
                        ))}
                    </div>
                )}
            </div>

            {/* Next Up Preview (if available) */}
            {queue?.next_clip && (
                <div className="p-4 border-t border-border bg-surface-raised/50">
                    <div className="text-xs text-muted-foreground mb-2">Up Next</div>
                    <div className="text-sm font-medium text-foreground truncate">
                        {queue.next_clip.title}
                    </div>
                    <div className="text-xs text-text-secondary mt-1">
                        {queue.next_clip.broadcaster_name}
                    </div>
                </div>
            )}
        </div>
    );
}
