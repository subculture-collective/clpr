import React from 'react';
import { Link, useLocation } from 'react-router-dom';
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
    const location = useLocation();
    const sortId = React.useId();

    const {
        data,
        isLoading,
        error,
        fetchNextPage,
        hasNextPage,
        isFetchingNextPage,
        isFetchNextPageError,
        isFetching,
        refetch,
    } = useComments(clipId, sort);

    const totalComments = data?.pages[0]?.total || 0;
    const allComments = React.useMemo(
        () => data?.pages.flatMap(page => page.comments) || [],
        [data],
    );

    const recovery = error ? (
        <div role='alert' className='rounded-lg border border-border p-4 text-sm'>
            <p>{allComments.length ? 'More comments could not be loaded. Your discussion is still here.' : 'Comments could not be loaded.'}</p>
            <Button variant='outline' className='mt-3' disabled={isFetching}
                onClick={() => isFetchNextPageError ? fetchNextPage() : refetch()}>
                Try again
            </Button>
        </div>
    ) : null;
    const signIn = (
        <p className='py-3 text-sm text-muted-foreground'>
            <Link to='/login' state={{ from: location }} className='text-link underline underline-offset-4'>Log in</Link> to join the discussion.
        </p>
    );
    const sortControl = (
        <div className='flex items-center gap-2'>
            <label htmlFor={sortId} className='text-sm text-muted-foreground'>Sort:</label>
            <select id={sortId} value={sort}
                onChange={e => setSort(e.target.value as CommentSortOption)}
                className='min-h-11 rounded-md border border-border bg-background px-2 text-sm'>
                <option value='best'>Best</option>
                <option value='top'>Top</option>
                <option value='new'>New</option>
                <option value='old'>Old</option>
                <option value='controversial'>Controversial</option>
            </select>
        </div>
    );

    if (isCompact) {
        return (
            <div className={cn('flex flex-col', className)}>
                {/* Header */}
                <div className='flex flex-wrap items-center justify-between gap-2 mb-3'>
                    <h2 className='text-[14px] font-semibold'>
                        Comments ({totalComments.toLocaleString()})
                    </h2>
                    {sortControl}
                </div>
                {recovery}

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
                <div className='min-h-0 flex-1 xl:overflow-y-auto'>
                    {isLoading ?
                        <div className='flex justify-center py-8'>
                            <Spinner size='lg' />
                        </div>
                    : allComments.length === 0 && !error ?
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
                                variant='expanded'
                            />

                            {hasNextPage && !error && (
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

                {!isAuthenticated && signIn}
                {/* Comment composer */}
                {isAuthenticated && !isBanned && (
                    <div className='shrink-0 border-t border-border bg-surface-raised pt-3'>
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

                {sortControl}
            </div>
            {recovery}

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
                    signIn
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
            : allComments.length === 0 && !error ?
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
                    {hasNextPage && !error && (
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
