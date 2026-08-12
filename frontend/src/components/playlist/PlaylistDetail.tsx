import { Badge } from '@/components/ui';
import { useAuth, useIsAuthenticated, useToast } from '@/hooks';
import {
    useInfinitePlaylist,
    useUpdatePlaylist,
    useCopyPlaylist,
    useLikePlaylist,
    useUnlikePlaylist,
    useBookmarkPlaylist,
    useUnbookmarkPlaylist,
    useRemoveClipFromPlaylist,
    useReorderPlaylistClips,
} from '@/hooks/usePlaylist';
import apiClient from '@/lib/api';
import { useParams, useNavigate } from 'react-router-dom';
import {
    Heart,
    Lock,
    Globe,
    Users,
    Copy,
    Bookmark,
} from 'lucide-react';
import { cn, formatRelativeTimestamp } from '@/lib/utils';
import { PlaylistTheatreMode } from './PlaylistTheatreMode';
import type { PlaylistItem } from './PlaylistTheatreMode';
import { useState, useCallback, useMemo } from 'react';
import { useQueryClient } from '@tanstack/react-query';
import { ShareButton } from '../clip/ShareButton';
import { CollaboratorManager } from './CollaboratorManager';
import { PlaylistCopyModal } from './PlaylistCopyModal';

export function PlaylistDetail() {
    const { id } = useParams<{ id: string }>();
    const navigate = useNavigate();
    const { user } = useAuth();
    const isAuthenticated = useIsAuthenticated();
    const playlistQuery = useInfinitePlaylist(id || '');
    const {
        isLoading,
        isError,
        error,
        refetch,
        hasNextPage,
        fetchNextPage,
        isFetchingNextPage,
    } = playlistQuery;
    const data = useMemo(() => {
        const first = playlistQuery.data?.pages[0];
        if (!first) return undefined;
        return {
            ...first,
            data: {
                ...first.data,
                clips: playlistQuery.data.pages.flatMap(
                    page => page.data.clips ?? [],
                ),
            },
        };
    }, [playlistQuery.data]);
    const likeMutation = useLikePlaylist();
    const unlikeMutation = useUnlikePlaylist();
    const bookmarkMutation = useBookmarkPlaylist();
    const unbookmarkMutation = useUnbookmarkPlaylist();
    const removeClip = useRemoveClipFromPlaylist();
    const reorderClips = useReorderPlaylistClips();
    const updatePlaylist = useUpdatePlaylist();
    const copyPlaylist = useCopyPlaylist();
    const toast = useToast();
    const [selectedItemId, setCurrentItemId] = useState<string | null>(null);
    const [showCopyModal, setShowCopyModal] = useState(false);
    const [visibilityOverride, setVisibilityOverride] = useState<{
        playlistId: string;
        value: 'private' | 'public' | 'unlisted';
    } | null>(null);
    const queryClient = useQueryClient();

    // Convert playlist clips to playlist items format
    const playlistItems: PlaylistItem[] = useMemo(() => {
        if (!data?.data?.clips) return [];
        return data.data.clips.map(clip => ({
            id: `${data.data.id}-${clip.id}`,
            clip,
            clip_id: clip.id,
            order: clip.order,
        }));
    }, [data]);

    const currentItemId =
        selectedItemId &&
        playlistItems.some(item => item.id === selectedItemId)
            ? selectedItemId
            : (playlistItems[0]?.id ?? null);
    const visibility =
        visibilityOverride?.playlistId === id
            ? visibilityOverride.value
            : (data?.data?.visibility ?? 'private');

    const handleItemClick = useCallback((item: PlaylistItem) => {
        setCurrentItemId(item.id);
    }, []);

    const loadNextPage = () => {
        if (!hasNextPage || isFetchingNextPage) return;
        const previousCount = playlistItems.length;
        void fetchNextPage().then(result => {
            const clips = result.data?.pages.flatMap(page => page.data.clips ?? []) ?? [];
            const nextClip = clips[previousCount];
            if (nextClip && data?.data.id) {
                setCurrentItemId(`${data.data.id}-${nextClip.id}`);
            }
        });
    };

    const handleItemRemove = useCallback(
        (itemId: string) => {
            if (!id) return;

            const item = playlistItems.find(i => i.id === itemId);
            if (!item) return;

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

            const item = playlistItems.find(i => i.id === itemId);
            if (!item) return;

            const draggedIndex = playlistItems.findIndex(i => i.id === itemId);
            if (draggedIndex === -1) return;

            const newOrder = [...playlistItems];
            const [draggedItem] = newOrder.splice(draggedIndex, 1);
            newOrder.splice(newPosition, 0, draggedItem);

            const clipIds = newOrder.map(i => i.clip_id);
            reorderClips.mutate({ id, data: { clip_ids: clipIds } });
        },
        [id, playlistItems, reorderClips],
    );

    const handleLike = async (e: React.MouseEvent) => {
        e.preventDefault();
        e.stopPropagation();

        if (!isAuthenticated || !id) {
            toast.info('Please log in to like playlists');
            return;
        }

        try {
            if (data?.data?.is_liked) {
                await unlikeMutation.mutateAsync(id);
            } else {
                await likeMutation.mutateAsync(id);
            }
        } catch {
            toast.error('Failed to update like status');
        }
    };

    const handleBookmark = async (e: React.MouseEvent) => {
        e.preventDefault();
        e.stopPropagation();

        if (!isAuthenticated || !id) {
            toast.info('Please log in to bookmark playlists');
            return;
        }

        try {
            if (data?.data?.is_bookmarked) {
                await unbookmarkMutation.mutateAsync(id);
                toast.success('Bookmark removed', {
                    action: {
                        label: 'Undo',
                        onClick: () => bookmarkMutation.mutate(id),
                    },
                });
            } else {
                await bookmarkMutation.mutateAsync(id);
                toast.success('Playlist bookmarked');
            }
        } catch {
            toast.error('Failed to update bookmark');
        }
    };

    const handleClipUpdated = useCallback(() => {
        if (!id) return;
        queryClient.invalidateQueries({ queryKey: ['playlist', id] });
    }, [id, queryClient]);

    const copyInitialValues = data?.data
        ? {
              title: `Copy of ${data.data.title}`,
              description: data.data.description || '',
              cover_url: data.data.cover_url || '',
              visibility: 'private' as const,
          }
        : {
              title: '',
              description: '',
              cover_url: '',
              visibility: 'private' as const,
          };

    const handleVisibilityChange = useCallback(
        async (nextVisibility: 'private' | 'public' | 'unlisted') => {
            if (!id) return;
            const previous = visibility;
            setVisibilityOverride({ playlistId: id, value: nextVisibility });
            try {
                await updatePlaylist.mutateAsync({
                    id,
                    data: { visibility: nextVisibility },
                });
                toast.success('Visibility updated');
            } catch {
                setVisibilityOverride({ playlistId: id, value: previous });
                toast.error('Failed to update visibility');
            }
        },
        [id, updatePlaylist, toast, visibility],
    );

    const trackShare = async (
        platform: 'link' | 'twitter' | 'facebook' | 'reddit' | 'bluesky',
    ) => {
        if (!id) return;
        const trackedPlatform =
            platform === 'twitter' || platform === 'facebook'
                ? platform
                : 'link';
        try {
            await apiClient.post(`/playlists/${id}/track-share`, {
                platform: trackedPlatform,
                referrer: window.location.href,
            });
        } catch {
            // Tracking failure shouldn't block sharing.
        }
    };

    if (!id) {
        return (
            <div className='text-center py-12 text-muted-foreground'>
                Playlist not found
            </div>
        );
    }

    if (isLoading) {
        return (
            <div className='text-center py-12 text-muted-foreground'>
                Loading...
            </div>
        );
    }

    if (isError) {
        const status = (error as { response?: { status?: number } })?.response
            ?.status;
        const message =
            status === 404 ? 'Playlist not found'
            : status === 403 ? 'You do not have permission to view this playlist'
            : status === 400 ? 'This playlist request is invalid'
            : 'The playlist could not be loaded';
        const actionClass =
            'min-h-11 rounded-md border border-border px-4 focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-primary-400';
        return (
            <div className='py-12 text-center'>
                <p className='mb-4 text-muted-foreground'>{message}</p>
                <div className='flex justify-center gap-3'>
                    <button className={actionClass} onClick={() => void refetch()}>
                        Retry
                    </button>
                    <button className={actionClass} onClick={() => navigate('/playlists')}>
                        Back to playlists
                    </button>
                </div>
            </div>
        );
    }

    if (!data?.data) {
        return (
            <div className='text-center py-12 text-muted-foreground'>
                Playlist not found
            </div>
        );
    }

    const playlist = data.data;
    const currentPermission = playlist.current_user_permission;
    const isOwner = user?.id === playlist.user_id;
    const canEdit =
        isOwner ||
        currentPermission === 'edit' ||
        currentPermission === 'admin';
    const canManageCollaborators = isOwner || currentPermission === 'admin';
    const canCopy = !!user;
    const isLiked = Boolean(playlist.is_liked);
    const isBookmarked = Boolean(playlist.is_bookmarked);
    const createdAt = formatRelativeTimestamp(playlist.created_at);

    const visibilityIconClass = 'h-2.5 w-2.5';
    const getVisibilityIcon = () => {
        switch (playlist.visibility) {
            case 'private':
                return <Lock className={visibilityIconClass} />;
            case 'public':
                return <Globe className={visibilityIconClass} />;
            case 'unlisted':
                return <Users className={visibilityIconClass} />;
            default:
                return null;
        }
    };

    const getVisibilityLabel = () => {
        switch (playlist.visibility) {
            case 'private':
                return 'Private';
            case 'public':
                return 'Public';
            case 'unlisted':
                return 'Unlisted';
            default:
                return '';
        }
    };

    return (
        <div className='w-full'>
            {/* Theatre Mode Player */}
            {playlistItems.length > 0 && (
                <div className='mb-6'>
                    <PlaylistTheatreMode
                        title={playlist.title}
                        items={playlistItems}
                        currentItemId={currentItemId}
                        onItemClick={handleItemClick}
                        onItemRemove={canEdit ? handleItemRemove : undefined}
                        onReorder={canEdit ? handleReorder : undefined}
                        onClipUpdated={handleClipUpdated}
                        onReachEnd={loadNextPage}
                        onClose={() => navigate(`/playlists/${id}/theatre`)}
                        isQueue={false}
                        contained={true}
                    />
                    {hasNextPage && (
                        <div className='mt-4 flex justify-center'>
                            <button
                                type='button'
                                className='min-h-11 rounded-md border border-border px-5 focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-primary-400 disabled:opacity-60'
                                disabled={isFetchingNextPage}
                                onClick={loadNextPage}
                            >
                                {isFetchingNextPage ? 'Loading…' : 'Load more clips'}
                            </button>
                        </div>
                    )}
                </div>
            )}

            {/* Header — compact, card-like */}
            <div className='mb-6'>
                {/* Title + description */}
                <h1 className='text-xl font-semibold text-foreground mb-1 leading-tight'>
                    {playlist.title}
                </h1>

                {playlist.description && (
                    <p className='text-sm text-muted-foreground mb-3'>
                        {playlist.description}
                    </p>
                )}

                {/* Badges row */}
                <div className='flex items-center gap-2 mb-3'>
                    {playlist.script_id && (
                        <Badge
                            variant='secondary'
                            size='sm'
                            className='shrink-0 border border-violet-400/40 bg-violet-500/12 text-[11px] text-violet-100 shadow-xs'
                        >
                            Scripted
                        </Badge>
                    )}

                    <Badge
                        variant='secondary'
                        size='sm'
                        className='shrink-0 border border-border bg-background/90 text-[11px] text-foreground shadow-xs'
                    >
                        {getVisibilityIcon()}
                        <span>{getVisibilityLabel()}</span>
                    </Badge>

                    {isOwner && (
                        <select
                            value={visibility}
                            onChange={e =>
                                handleVisibilityChange(
                                    e.target.value as typeof visibility,
                                )
                            }
                            className='bg-background text-foreground border border-border rounded px-2 py-0.5 text-xs'
                        >
                            <option value='private'>Private</option>
                            <option value='unlisted'>Unlisted</option>
                            <option value='public'>Public</option>
                        </select>
                    )}
                </div>

                {/* Stats + actions row (matches PlaylistCard footer style) */}
                <div className='flex items-center justify-between gap-2 text-xs text-muted-foreground'>
                    <div className='flex flex-wrap items-center gap-x-2 gap-y-1'>
                        {/* Clip count */}
                        <span className='inline-flex items-center gap-1'>
                            <span className='font-medium text-foreground/90'>
                                {playlist.clip_count}
                            </span>
                            <span>clips</span>
                        </span>

                        {/* Like */}
                        <button
                            type='button'
                            onClick={handleLike}
                            disabled={
                                likeMutation.isPending ||
                                unlikeMutation.isPending
                            }
                            className='inline-flex items-center gap-1 rounded px-1 py-0.5 text-muted-foreground transition-colors hover:bg-accent hover:text-primary-500 disabled:cursor-not-allowed disabled:opacity-60'
                            title={
                                isAuthenticated
                                    ? isLiked
                                        ? 'Unlike playlist'
                                        : 'Like playlist'
                                    : 'Log in to like playlists'
                            }
                        >
                            <Heart
                                className={cn(
                                    'h-3.5 w-3.5',
                                    isLiked &&
                                        'fill-current text-primary-500',
                                )}
                            />
                            <span className='font-medium text-foreground/90'>
                                {playlist.like_count}
                            </span>
                        </button>

                        {/* Bookmark */}
                        <button
                            type='button'
                            onClick={handleBookmark}
                            disabled={
                                bookmarkMutation.isPending ||
                                unbookmarkMutation.isPending
                            }
                            className='inline-flex items-center gap-1 rounded px-1 py-0.5 text-muted-foreground transition-colors hover:bg-accent hover:text-primary-500 disabled:cursor-not-allowed disabled:opacity-60'
                            title={
                                isAuthenticated
                                    ? isBookmarked
                                        ? 'Remove bookmark'
                                        : 'Bookmark playlist'
                                    : 'Log in to bookmark playlists'
                            }
                        >
                            <Bookmark
                                className={cn(
                                    'h-3.5 w-3.5',
                                    isBookmarked &&
                                        'fill-current text-primary-500',
                                )}
                            />
                            <span className='font-medium text-foreground/90'>
                                {playlist.bookmark_count}
                            </span>
                        </button>

                        {/* Share */}
                        {playlist.visibility !== 'private' && (
                            <ShareButton
                                shareUrl={`${window.location.origin}/playlists/${playlist.id}`}
                                shareTitle={
                                    playlist.title ||
                                    playlist.description ||
                                    'Check out this playlist!'
                                }
                                onShare={trackShare}
                                showLabel={false}
                                buttonClassName='inline-flex min-h-0 items-center rounded px-1 py-0.5 text-muted-foreground transition-colors hover:bg-accent hover:text-primary-500'
                                iconClassName='h-3.5 w-3.5'
                            />
                        )}

                        {/* Copy Playlist */}
                        {canCopy && (
                            <button
                                type='button'
                                onClick={() => setShowCopyModal(true)}
                                className='inline-flex items-center gap-1 rounded px-1 py-0.5 text-muted-foreground transition-colors hover:bg-accent hover:text-primary-500 cursor-pointer'
                                title='Copy playlist'
                            >
                                <Copy className='h-3.5 w-3.5' />
                                <span>Copy Playlist</span>
                            </button>
                        )}
                    </div>

                    {/* Right side: creator + timestamp */}
                    <div className='flex items-center gap-1.5 shrink-0 text-right'>
                        {playlist.creator && (
                            <>
                                <span className='text-foreground/80'>
                                    {playlist.creator.display_name}
                                </span>
                                <span>·</span>
                            </>
                        )}
                        <span title={createdAt.title}>
                            {createdAt.display}
                        </span>
                    </div>
                </div>
            </div>

            {(isOwner || currentPermission) && (
                <div className='mb-6'>
                    <CollaboratorManager
                        playlistId={playlist.id}
                        isOwner={isOwner}
                        canManageCollaborators={canManageCollaborators}
                    />
                </div>
            )}

            {showCopyModal && (
                <PlaylistCopyModal
                    initialValues={copyInitialValues}
                    isSubmitting={copyPlaylist.isPending}
                    onClose={() => setShowCopyModal(false)}
                    onSubmit={async values => {
                        if (!id) return;
                        try {
                            const copied = await copyPlaylist.mutateAsync({
                                id,
                                data: {
                                    title: values.title,
                                    description:
                                        values.description || undefined,
                                    cover_url: values.cover_url || undefined,
                                    visibility: values.visibility,
                                },
                            });
                            toast.success('Playlist copied');
                            setShowCopyModal(false);
                            navigate(`/playlists/${copied.id}`);
                        } catch {
                            toast.error('Failed to copy playlist');
                        }
                    }}
                />
            )}
        </div>
    );
}
