import { Link } from 'react-router-dom';
import { Bookmark, Users } from 'lucide-react';
import {
    useFollowDiscoveryList,
    useUnfollowDiscoveryList,
} from '@/hooks/useDiscoveryLists';
import { useIsAuthenticated, useToast } from '@/hooks';
import type { DiscoveryListWithStats } from '../../types/discoveryList';
import { cn, formatCompactNumber } from '../../lib/utils';

interface DiscoveryListCardProps {
    list: DiscoveryListWithStats;
    showPreviewClips?: boolean;
    size?: 'default' | 'compact';
    className?: string;
}

export function DiscoveryListCard({
    list,
    showPreviewClips = true,
    size = 'default',
    className,
}: DiscoveryListCardProps) {
    const isCompact = size === 'compact';
    const isAuthenticated = useIsAuthenticated();
    const toast = useToast();
    const followMutation = useFollowDiscoveryList();
    const unfollowMutation = useUnfollowDiscoveryList();
    const isFollowing = list.is_following;
    const isUpdating = followMutation.isPending || unfollowMutation.isPending;

    return (
        <Link
            to={`/discover/lists/${list.slug}`}
            className={cn(
                'block bg-card border border-border rounded-xl hover:shadow-lg transition-all hover:border-primary-500/30 overflow-hidden group',
                isCompact && 'h-[325px] flex flex-col',
                className,
            )}
        >
            {/* Preview Clips Thumbnails */}
            {showPreviewClips &&
                list.preview_clips &&
                list.preview_clips.length > 0 && (
                    <div
                        className={cn(
                            'grid grid-cols-2 gap-1 p-2',
                            isCompact && 'p-1.5 h-[216px] overflow-hidden',
                        )}
                    >
                        {list.preview_clips.slice(0, 4).map((clip, idx) => (
                            <div
                                key={clip.id}
                                className={cn(
                                    'aspect-video rounded-lg overflow-hidden bg-accent',
                                    list.preview_clips!.length === 1 &&
                                        'col-span-2',
                                    list.preview_clips!.length === 3 &&
                                        idx === 0 &&
                                        'col-span-2',
                                )}
                            >
                                {clip.thumbnail_url ?
                                    <img
                                        src={clip.thumbnail_url}
                                        alt={clip.title}
                                        className='w-full h-full object-cover group-hover:scale-105 transition-transform duration-300'
                                        loading='lazy'
                                    />
                                :   <div className='w-full h-full bg-gradient-to-br from-accent to-accent/50' />
                                }
                            </div>
                        ))}
                    </div>
                )}

            {/* List Info */}
            <div className={cn('p-4', isCompact && 'p-3 flex-1 flex flex-col')}>
                <div className='flex items-start justify-between gap-3 mb-2'>
                    <div className='flex-1 min-w-0'>
                        <h3
                            className={cn(
                                'font-semibold text-foreground text-lg mb-1 line-clamp-1 group-hover:text-primary-500 transition-colors',
                                isCompact && 'text-base',
                            )}
                        >
                            {list.name}
                        </h3>
                        {list.description && (
                            <p
                                className={cn(
                                    'text-sm text-muted-foreground line-clamp-2',
                                    isCompact && 'text-xs',
                                )}
                            >
                                {list.description}
                            </p>
                        )}
                    </div>
                    {list.is_featured && (
                        <span className='inline-flex items-center px-2 py-1 rounded-md text-xs font-medium bg-primary-500/10 text-primary-500 border border-primary-500/20 shrink-0'>
                            Featured
                        </span>
                    )}
                </div>

                {/* Stats + Follow */}
                <div
                    className={cn(
                        'flex items-center justify-between gap-3 mt-3',
                        isCompact && 'mt-auto',
                    )}
                >
                    <div
                        className={cn(
                            'flex items-center gap-4 text-sm text-muted-foreground',
                            isCompact && 'gap-3 text-xs',
                        )}
                    >
                        <div className='flex items-center gap-1.5'>
                            <svg
                                className='w-4 h-4'
                                fill='none'
                                stroke='currentColor'
                                viewBox='0 0 24 24'
                                aria-hidden='true'
                            >
                                <path
                                    strokeLinecap='round'
                                    strokeLinejoin='round'
                                    strokeWidth={2}
                                    d='M7 4v16M17 4v16M3 8h4m10 0h4M3 12h18M3 16h4m10 0h4M4 20h16a1 1 0 001-1V5a1 1 0 00-1-1H4a1 1 0 00-1 1v14a1 1 0 001 1z'
                                />
                            </svg>
                            <span>{formatCompactNumber(list.clip_count)} clips</span>
                        </div>
                        <div className='flex items-center gap-1.5'>
                            <Users className='w-4 h-4' aria-hidden='true' />
                            <span>
                                {formatCompactNumber(list.follower_count)} followers
                            </span>
                        </div>
                    </div>

                    <button
                        type='button'
                        onClick={event => {
                            event.preventDefault();
                            event.stopPropagation();
                            if (!isAuthenticated) {
                                toast.info('Please log in to follow lists');
                                return;
                            }
                            if (isFollowing) {
                                unfollowMutation.mutate({
                                    id: list.id,
                                    slug: list.slug,
                                    idOrSlug: list.slug,
                                });
                            } else {
                                followMutation.mutate({
                                    id: list.id,
                                    slug: list.slug,
                                    idOrSlug: list.slug,
                                });
                            }
                        }}
                        disabled={isUpdating}
                        className={cn(
                            'inline-flex items-center justify-center rounded-md border px-2.5 py-1 text-xs font-medium transition-colors',
                            isFollowing ?
                                'border-primary-500/30 bg-primary-500/10 text-primary-500'
                            :   'border-border bg-background text-foreground hover:bg-accent',
                            isUpdating && 'opacity-50 cursor-not-allowed',
                        )}
                        aria-label={
                            isFollowing ? 'Unfollow list' : 'Follow list'
                        }
                    >
                        <Bookmark
                            className={cn(
                                'w-4 h-4',
                                isFollowing && 'fill-current',
                            )}
                            aria-hidden='true'
                        />
                        <span className='sr-only'>
                            {isFollowing ? 'Unfollow list' : 'Follow list'}
                        </span>
                    </button>
                </div>
            </div>
        </Link>
    );
}
