import { useState } from 'react';
import { useParams, Navigate } from 'react-router-dom';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { Helmet } from '@dr.pogodin/react-helmet';
import { Pencil, Eye, EyeOff } from 'lucide-react';
import { useAuth } from '../context/AuthContext';
import { fetchCreatorClips, updateClipMetadata, updateClipVisibility } from '../lib/clip-api';
import { Container, Button } from '../components';
import type { Clip } from '../types/clip';

export function CreatorDashboardPage() {
  const { user } = useAuth();
  const { creatorId } = useParams<{ creatorId: string }>();
  const queryClient = useQueryClient();
  const [editingClipId, setEditingClipId] = useState<string | null>(null);
  const [editTitle, setEditTitle] = useState('');

  // Fetch creator clips
  const { data, isLoading, error } = useQuery({
    queryKey: ['creatorClips', creatorId],
    queryFn: () => fetchCreatorClips({ creatorId: creatorId! }),
    enabled: !!creatorId,
  });

  // Mutation for updating metadata
  const updateMetadataMutation = useMutation({
    mutationFn: ({ clipId, title }: { clipId: string; title: string }) =>
      updateClipMetadata(clipId, { title }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['creatorClips', creatorId] });
      setEditingClipId(null);
      setEditTitle('');
    },
  });

  // Mutation for updating visibility
  const updateVisibilityMutation = useMutation({
    mutationFn: ({ clipId, isHidden }: { clipId: string; isHidden: boolean }) =>
      updateClipVisibility(clipId, isHidden),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['creatorClips', creatorId] });
    },
  });

  // Check if user has access to this dashboard
  if (!user || (user.twitch_id !== creatorId && user.role !== 'admin' && user.role !== 'moderator')) {
    return <Navigate to="/" replace />;
  }

  const handleEditClick = (clip: Clip) => {
    setEditingClipId(clip.id);
    setEditTitle(clip.title);
  };

  const handleSaveEdit = (clipId: string) => {
    const trimmedTitle = editTitle.trim();
    if (trimmedTitle && trimmedTitle.length >= 1 && trimmedTitle.length <= 255) {
      updateMetadataMutation.mutate({ clipId, title: trimmedTitle });
    }
  };

  const handleCancelEdit = () => {
    setEditingClipId(null);
    setEditTitle('');
  };

  const handleToggleVisibility = (clip: Clip) => {
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
        <div className="text-center text-red-600">Failed to load creator clips</div>
      </Container>
    );
  }

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
        </div>

        {data?.clips.length === 0 ? (
          <div className="text-center text-muted-foreground py-12">
            <p className="text-lg">No clips found</p>
          </div>
        ) : (
          <div className="space-y-6">
            {data?.clips.map((clip) => (
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
                          aria-describedby={`clip-title-help-${clip.id}`}
                        />
                        <span id={`clip-title-help-${clip.id}`} className="sr-only">
                          Enter a new title for this clip (1-255 characters)
                        </span>
                        <div className="flex gap-2">
                          <Button
                            size="sm"
                            onClick={() => handleSaveEdit(clip.id)}
                            disabled={updateMetadataMutation.isPending}
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
                  </div>
                </div>
              </div>
            ))}
          </div>
        )}

        {/* Pagination info */}
        {data && data.total > 0 && (
          <div className="mt-6 text-center text-sm text-muted-foreground">
            Showing {data.clips.length} of {data.total} clips
          </div>
        )}
      </Container>
    </>
  );
}
