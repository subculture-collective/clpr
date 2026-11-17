import { useEffect, useState } from 'react';
import { Link, useParams, useSearchParams, Navigate } from 'react-router-dom';
import { useInfiniteQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { Helmet } from '@dr.pogodin/react-helmet';
import { Pencil, Eye, EyeOff } from 'lucide-react';
import { useAuth } from '../context/AuthContext';
import { fetchCreatorClips, updateClipMetadata, updateClipVisibility } from '../lib/clip-api';
import { Alert, Container, Button, ResourceUnavailable } from '../components';
import type { Clip } from '../types/clip';
import { CreatorEvents, trackEvent } from '../lib/telemetry';
import { CountdownPlacementsCard } from '../components/countdown/CountdownPlacementsCard';
import { ClipPacksCard } from '../components/countdown/ClipPacksCard';
import { getErrorMessage, getHttpStatus } from '../lib/error-utils';

const MAX_TITLE_LENGTH = 255;

type ClipAction = 'title' | 'hide' | 'show';

function clipActionErrorMessage(error: unknown, action: ClipAction): string {
  const status = getHttpStatus(error);
  if (status === 429) {
    return 'You are making changes too quickly. Wait a minute and try again.';
  }
  // The server refuses to show a clip again when it was removed, for example
  // for copyright. Its message says why.
  if (action === 'show' && (status === 403 || status === 409)) {
    return getErrorMessage(error, 'This clip cannot be shown again.');
  }
  if (status === 403) {
    return 'You do not have permission to change this clip.';
  }
  if (status === 409) {
    return getErrorMessage(error, 'This clip cannot be changed right now.');
  }
  return getErrorMessage(
    error,
    action === 'title' ? 'The title could not be saved. Try again.' : 'The clip could not be updated. Try again.',
  );
}

type ConnectNotice = { variant: 'success' | 'error'; message: string };

// The Twitch connect flow returns here with ?error=<code> or ?twitch_connected=true.
function readConnectNotice(params: URLSearchParams): ConnectNotice | null {
  const error = params.get('error');
  if (error === 'oauth_denied') {
    return { variant: 'error', message: 'Your channel was not connected because the Twitch request was cancelled.' };
  }
  if (error) {
    return { variant: 'error', message: 'Your channel could not be connected. Try again.' };
  }
  if (params.get('twitch_connected') === 'true') {
    return { variant: 'success', message: 'Your Twitch channel is connected.' };
  }
  return null;
}

export function CreatorDashboardPage() {
  const { user } = useAuth();
  const { creatorId } = useParams<{ creatorId: string }>();
  const queryClient = useQueryClient();
  const [editingClipId, setEditingClipId] = useState<string | null>(null);
  const [editTitle, setEditTitle] = useState('');
  const [actionError, setActionError] = useState<{ clipId: string; message: string } | null>(null);
  const [searchParams, setSearchParams] = useSearchParams();
  const [connectNotice, setConnectNotice] = useState(() => readConnectNotice(searchParams));

  // Keep the message, but take the result out of the URL so a reload or a
  // shared link does not show it again.
  useEffect(() => {
    if (!searchParams.has('error') && !searchParams.has('twitch_connected')) return;
    const next = new URLSearchParams(searchParams);
    next.delete('error');
    next.delete('twitch_connected');
    setSearchParams(next, { replace: true });
  }, [searchParams, setSearchParams]);

  // Fetch creator clips
  const { data, isLoading, error, refetch, fetchNextPage, hasNextPage, isFetchingNextPage } = useInfiniteQuery({
    queryKey: ['creatorClips', creatorId],
    queryFn: ({ pageParam }) => fetchCreatorClips({ creatorId: creatorId!, pageParam }),
    getNextPageParam: lastPage => (lastPage.has_more ? lastPage.page + 1 : undefined),
    initialPageParam: 1,
    enabled: !!creatorId,
  });
  const clips = data?.pages.flatMap(page => page.clips ?? []) ?? [];
  const total = data?.pages[0]?.total ?? 0;

  // Mutation for updating metadata
  const updateMetadataMutation = useMutation({
    mutationFn: ({ clipId, title }: { clipId: string; title: string }) =>
      updateClipMetadata(clipId, { title }),
    onSuccess: (_result, { clipId }) => {
      trackEvent(CreatorEvents.CLIP_TITLE_UPDATED, { clip_id: clipId });
      queryClient.invalidateQueries({ queryKey: ['creatorClips', creatorId] });
      setEditingClipId(null);
      setEditTitle('');
    },
    onError: (err, { clipId }) => {
      setActionError({ clipId, message: clipActionErrorMessage(err, 'title') });
    },
  });

  // Mutation for updating visibility
  const updateVisibilityMutation = useMutation({
    mutationFn: ({ clipId, isHidden }: { clipId: string; isHidden: boolean }) =>
      updateClipVisibility(clipId, isHidden),
    onSuccess: (_result, { clipId, isHidden }) => {
      trackEvent(CreatorEvents.CLIP_VISIBILITY_UPDATED, { clip_id: clipId, is_hidden: isHidden });
      queryClient.invalidateQueries({ queryKey: ['creatorClips', creatorId] });
    },
    onError: (err, { clipId, isHidden }) => {
      setActionError({ clipId, message: clipActionErrorMessage(err, isHidden ? 'hide' : 'show') });
    },
  });

  // Check if user has access to this dashboard
  if (!user || (user.twitch_id !== creatorId && user.role !== 'admin' && user.role !== 'moderator')) {
    return <Navigate to="/" replace />;
  }

  const handleEditClick = (clip: Clip) => {
    setActionError(null);
    setEditingClipId(clip.id);
    setEditTitle(clip.title);
  };

  const handleSaveEdit = (clipId: string) => {
    const trimmedTitle = editTitle.trim();
    if (trimmedTitle && trimmedTitle.length <= MAX_TITLE_LENGTH) {
      setActionError(null);
      updateMetadataMutation.mutate({ clipId, title: trimmedTitle });
    }
  };

  const handleCancelEdit = () => {
    setActionError(null);
    setEditingClipId(null);
    setEditTitle('');
  };

  const handleToggleVisibility = (clip: Clip) => {
    setActionError(null);
    updateVisibilityMutation.mutate({
      clipId: clip.id,
      isHidden: !clip.is_hidden,
    });
  };

  if (isLoading) {
    return (
      <Container className="py-8">
        <div className="text-center">Loading creator clips...</div>
      </Container>
    );
  }

  if (error) {
    return (
      <Container className="py-8">
        <ResourceUnavailable
          kind="error"
          title="We couldn't load your clips"
          description="Check your connection and try again."
          onRetry={() => void refetch()}
          links={[{ label: 'Back to the feed', href: '/' }]}
        />
      </Container>
    );
  }

  const titleIsEmpty = editTitle.trim() === '';
  // Analytics are keyed by the creator name stored on the clips.
  const analyticsName = clips[0]?.creator_name;

  return (
    <>
      <Helmet>
        <title>Creator Dashboard - clpr</title>
      </Helmet>

      <Container className="py-8">
        <div className="mb-8">
          <h1 className="text-3xl font-bold text-foreground mb-2">
            Creator Dashboard
          </h1>
          <p className="text-muted-foreground">
            Manage your clips - edit titles and control visibility
          </p>
          {analyticsName && (
            <Link
              to={`/creator/${encodeURIComponent(analyticsName)}/analytics`}
              className="mt-2 inline-block text-sm underline hover:text-foreground"
            >
              Analytics
            </Link>
          )}
        </div>

        {connectNotice && (
          <Alert
            variant={connectNotice.variant}
            className="mb-6"
            dismissible
            onDismiss={() => setConnectNotice(null)}
          >
            {connectNotice.message}
          </Alert>
        )}

        {user.twitch_id === creatorId && (
          <CountdownPlacementsCard returnTo={`/creator/${encodeURIComponent(creatorId)}/dashboard`} />
        )}
        {user.twitch_id === creatorId && <ClipPacksCard />}

        {clips.length === 0 ? (
          <div className="text-center text-muted-foreground py-12">
            <p className="text-lg">No clips found</p>
          </div>
        ) : (
          <div className="space-y-6">
            {clips.map((clip) => (
              <div
                key={clip.id}
                className="bg-surface rounded-lg shadow-md p-6"
              >
                <div className="flex items-start gap-4">
                  {/* Clip thumbnail */}
                  <div className="flex-shrink-0 w-32">
                    {clip.thumbnail_url && (
                      <img
                        src={clip.thumbnail_url}
                        alt={`Thumbnail for ${clip.title}`}
                        className="w-full rounded"
                        loading="lazy"
                        width="128"
                        height="72"
                      />
                    )}
                  </div>

                  {/* Clip details and controls */}
                  <div className="flex-1">
                    {editingClipId === clip.id ? (
                      <div className="space-y-3">
                        <label htmlFor={`clip-title-${clip.id}`} className="sr-only">
                          Clip title
                        </label>
                        <input
                          id={`clip-title-${clip.id}`}
                          type="text"
                          value={editTitle}
                          onChange={(e) => setEditTitle(e.target.value)}
                          className="w-full px-3 py-2 border border-border rounded-md focus:outline-none focus:ring-2 focus:ring-primary-500 bg-surface text-foreground"
                          placeholder="Clip title"
                          maxLength={MAX_TITLE_LENGTH}
                          aria-describedby={`clip-title-help-${clip.id}`}
                        />
                        <span id={`clip-title-help-${clip.id}`} className="sr-only">
                          Enter a new title for this clip (1-255 characters)
                        </span>
                        {titleIsEmpty && (
                          <p className="text-sm text-muted-foreground">Enter a title.</p>
                        )}
                        <div className="flex gap-2">
                          <Button
                            size="sm"
                            onClick={() => handleSaveEdit(clip.id)}
                            disabled={updateMetadataMutation.isPending || titleIsEmpty}
                          >
                            {updateMetadataMutation.isPending ? 'Saving...' : 'Save'}
                          </Button>
                          <Button
                            size="sm"
                            variant="outline"
                            onClick={handleCancelEdit}
                            disabled={updateMetadataMutation.isPending}
                          >
                            Cancel
                          </Button>
                        </div>
                      </div>
                    ) : (
                      <>
                        <div className="flex items-center justify-between mb-2">
                          <h3 className="text-lg font-semibold text-foreground">
                            {clip.title}
                            {clip.is_hidden && (
                              <span className="ml-2 text-sm text-yellow-600 dark:text-yellow-400">
                                (Hidden)
                              </span>
                            )}
                          </h3>
                          <button
                            onClick={() => handleEditClick(clip)}
                            className="p-2 hover:bg-surface-hover rounded focus:outline-none focus:ring-2 focus:ring-primary-500 focus:ring-offset-2"
                            title="Edit title"
                            aria-label={`Edit title for ${clip.title}`}
                          >
                            <Pencil className="w-4 h-4" aria-hidden="true" />
                          </button>
                        </div>
                        <p className="text-sm text-muted-foreground mb-3">
                          Created by {clip.creator_name} • {clip.view_count} views • Score: {clip.vote_score}
                        </p>
                      </>
                    )}

                    {/* Visibility toggle */}
                    <div className="mt-3">
                      <Button
                        size="sm"
                        variant={clip.is_hidden ? 'outline' : 'secondary'}
                        onClick={() => handleToggleVisibility(clip)}
                        disabled={updateVisibilityMutation.isPending}
                        className="flex items-center gap-2"
                      >
                        {clip.is_hidden ? (
                          <>
                            <Eye className="w-4 h-4" />
                            Show Clip
                          </>
                        ) : (
                          <>
                            <EyeOff className="w-4 h-4" />
                            Hide Clip
                          </>
                        )}
                      </Button>
                    </div>

                    {actionError?.clipId === clip.id && (
                      <Alert variant="error" className="mt-3">
                        {actionError.message}
                      </Alert>
                    )}
                  </div>
                </div>
              </div>
            ))}
          </div>
        )}

        {/* Pagination */}
        {total > 0 && (
          <div className="mt-6 text-center text-sm text-muted-foreground">
            Showing {clips.length} of {total} clips
          </div>
        )}
        {hasNextPage && (
          <div className="mt-4 flex justify-center">
            <Button onClick={() => void fetchNextPage()} disabled={isFetchingNextPage}>
              {isFetchingNextPage ? 'Loading...' : 'Load more'}
            </Button>
          </div>
        )}
      </Container>
    </>
  );
}
