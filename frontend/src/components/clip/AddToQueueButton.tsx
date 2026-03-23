import { useAddToQueue, useQueue } from '@/hooks/useQueue';
import { useIsAuthenticated, useToast } from '@/hooks';
import { useAuth } from '@/context/AuthContext';
import { AxiosError } from 'axios';
import { ListPlus, Check } from 'lucide-react';

interface AddToQueueButtonProps {
    clipId: string;
}

export function AddToQueueButton({ clipId }: AddToQueueButtonProps) {
    const isAuthenticated = useIsAuthenticated();
    const { user } = useAuth();
    const addToQueue = useAddToQueue();
    const { data: queue } = useQueue(100, !!user);
    const toast = useToast();

    // Check if clip is already in queue
    const isInQueue = queue?.items?.some(item => item.clip_id === clipId) ?? false;

    const handleAddToQueue = async () => {
        if (!isAuthenticated) {
            toast.info('Please log in to add clips to queue');
            return;
        }

        if (isInQueue) {
            toast.info('This clip is already in your queue');
            return;
        }

        try {
            await addToQueue.mutateAsync({
                clip_id: clipId,
                at_end: true,
            });
            toast.success('Added to queue');
        } catch (error) {
            let message = 'Failed to add to queue';
            if (
                error instanceof AxiosError &&
                error.response?.data?.error?.message
            ) {
                message = error.response.data.error.message;
            } else if (error instanceof Error) {
                message = error.message;
            }
            toast.error(message);
        }
    };

    return (
        <button
            onClick={handleAddToQueue}
            disabled={!isAuthenticated || addToQueue.isPending || isInQueue}
            className={`flex items-center gap-1.5 transition-colors touch-target min-h-11 ${
                isInQueue ?
                    'text-brand cursor-default'
                : !isAuthenticated ?
                    'text-muted-foreground opacity-50 cursor-not-allowed hover:bg-transparent'
                :   'text-muted-foreground hover:text-foreground cursor-pointer'
            }`}
            aria-label={
                isInQueue ? 'Already in queue'
                : !isAuthenticated ? 'Log in to add to queue'
                : 'Add to queue'
            }
            aria-disabled={!isAuthenticated || isInQueue}
            title={
                isInQueue ? 'Already in queue'
                : !isAuthenticated ? 'Log in to add to queue'
                : 'Add to queue'
            }
        >
            {isInQueue ? (
                <Check size={18} className='shrink-0' strokeWidth={1.75} />
            ) : (
                <ListPlus size={18} className='shrink-0' strokeWidth={1.75} />
            )}
            <span className='hidden sm:inline'>
                {isInQueue ? 'In Queue' : 'Queue'}
            </span>
        </button>
    );
}
