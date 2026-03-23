import React from 'react';
import { cn } from '@/lib/utils';
import { Button, Spinner } from '@/components/ui';
import { CommentTree } from './CommentTree';
import { CommentForm } from './CommentForm';
import { useComments, useIsAuthenticated } from '@/hooks';
import type { CommentSortOption } from '@/types/comment';

interface CommentSectionProps {
    clipId: string;
    currentUserId?: string;
    isAdmin?: boolean;
    variant?: 'expanded' | 'compact';
    className?: string;
    isBanned?: boolean;
    banReason?: string;
}

export const CommentSection: React.FC<CommentSectionProps> = ({
    clipId,
    currentUserId,
    isAdmin = false,
    variant = 'expanded',
    className,
    isBanned = false,
    banReason,
}) => {
    const [sort, setSort] = React.useState<CommentSortOption>('best');
    const isAuthenticated = useIsAuthenticated();
    const isCompact = variant === 'compact';

    const {
        data,
        isLoading,
        error,
        fetchNextPage,
        hasNextPage,
        isFetchingNextPage,
    } = useComments(clipId, sort);

    const totalComments = data?.pages[0]?.total || 0;
    const allComments = React.useMemo(
        () => data?.pages.flatMap(page => page.comments) || [],
        [data],
    );

    if (error) {
        return (
            <div className={cn('space-y-4', className)}>
                <div className='text-center py-8'>
                    <p className='text-error-500'>Error loading comments</p>
                    <p className='text-sm text-muted-foreground mt-2'>
                        {error instanceof Error ?
                            error.message
                        :   'Something went wrong'}
                    </p>
                </div>
            </div>
        );
    }

    if (isCompact) {
        return (
            <div className={cn('flex flex-col', className)}>
                {/* Header */}
                <div className='flex items-center justify-between mb-3'>
                    <h2 className='text-[14px] font-semibold'>
                        Comments ({totalComments.toLocaleString()})
                    </h2>
                </div>

                {isBanned && (
                    <div
                        role='alert'
                        className='rounded-lg border border-red-200 bg-red-50 dark:bg-red-900/20 dark:border-red-800 p-3 text-sm text-red-800 mb-3'
                    >
                        You are banned and cannot comment
                        {banReason ? `: ${banReason}` : ''}.
                    </div>
                )}

                {/* Scrollable comments area */}
                <div className='flex-1 overflow-y-auto'>
                    {isLoading ?
                        <div className='flex justify-center py-8'>
                            <Spinner size='lg' />
                        </div>
                    : allComments.length === 0 ?
                        <div className='text-center py-8'>
                            <p className='text-sm font-semibold mb-1'>
                                No comments yet
                            </p>
                            <p className='text-xs text-muted-foreground'>
                                Be the first to comment!
                            </p>
                        </div>
                    :   <>
                            <CommentTree
                                comments={allComments}
                                clipId={clipId}
                                currentUserId={currentUserId}
                                isAdmin={isAdmin}
                                depth={0}
                                maxDepth={2}
                                variant='compact'
                            />

                            {hasNextPage && (
                                <div className='flex justify-center pt-3'>
                                    <Button
                                        onClick={() => fetchNextPage()}
                                        disabled={isFetchingNextPage}
                                        loading={isFetchingNextPage}
                                        variant='outline'
                                    >
                                        {isFetchingNextPage ?
                                            'Loading...'
                                        :   'Load More'}
                                    </Button>
                                </div>
                            )}
                        </>
                    }
                </div>

                {/* Sticky comment form at bottom */}
                {isAuthenticated && !isBanned && (
                    <div className='sticky bottom-0 pt-3'>
                        <CommentForm
                            clipId={clipId}
                            placeholder='Add a comment...'
                        />
                    </div>
                )}
            </div>
        );
    }

    return (
        <div className={cn('space-y-6', className)}>
            {/* Header */}
            <div className='flex items-center justify-between'>
                <h2 className='text-2xl font-bold'>
                    Comments ({totalComments.toLocaleString()})
                </h2>

                {/* Sort dropdown */}
                <div className='flex items-center gap-2'>
                    <label
                        htmlFor='sort-select'
                        className='text-sm text-muted-foreground'
                    >
                        Sort by:
                    </label>
                    <select
                        id='sort-select'
                        value={sort}
                        onChange={e =>
                            setSort(e.target.value as CommentSortOption)
                        }
                        className='px-3 py-1.5 rounded-md border border-border bg-background text-foreground text-sm focus:outline-none focus:ring-2 focus:ring-primary-500 cursor-pointer'
                    >
                        <option value='best'>Best</option>
                        <option value='top'>Top</option>
                        <option value='new'>New</option>
                        <option value='old'>Old</option>
                        <option value='controversial'>Controversial</option>
                    </select>
                </div>
            </div>

            {isBanned && (
                <div
                    role='alert'
                    className='rounded-lg border border-red-200 bg-red-50 dark:bg-red-900/20 dark:border-red-800 p-3 text-sm text-red-800'
                >
                    You are banned and cannot comment
                    {banReason ? `: ${banReason}` : ''}.
                </div>
            )}

            {/* Add comment button/form */}
            <div>
                {!isAuthenticated ?
                    <div className='text-center py-6 border border-border rounded-lg'>
                        <p className='text-muted-foreground mb-3'>
                            Please log in to comment
                        </p>
                        <Button
                            onClick={() => (window.location.href = '/login')}
                            variant='primary'
                        >
                            Log In
                        </Button>
                    </div>
                : !isBanned ?
                    <CommentForm
                        clipId={clipId}
                        placeholder='What are your thoughts?'
                    />
                :   null}
            </div>

            {/* Loading state */}
            {isLoading ?
                <div className='flex justify-center py-12'>
                    <Spinner size='lg' />
                </div>
            : allComments.length === 0 ?
                /* Empty state */
                <div className='text-center py-12 border border-border rounded-lg'>
                    <p className='text-xl font-semibold mb-2'>
                        No comments yet
                    </p>
                    <p className='text-muted-foreground mb-4'>
                        Be the first to comment!
                    </p>
                </div>
            :   <>
                    {/* Comments list */}
                    <CommentTree
                        comments={allComments}
                        clipId={clipId}
                        currentUserId={currentUserId}
                        isAdmin={isAdmin}
                        depth={0}
                        maxDepth={10}
                    />

                    {/* Load more button */}
                    {hasNextPage && (
                        <div className='flex justify-center pt-4'>
                            <Button
                                onClick={() => fetchNextPage()}
                                disabled={isFetchingNextPage}
                                loading={isFetchingNextPage}
                                variant='outline'
                            >
                                {isFetchingNextPage ?
                                    'Loading...'
                                :   'Load More Comments'}
                            </Button>
                        </div>
                    )}
                </>
            }
        </div>
    );
};
