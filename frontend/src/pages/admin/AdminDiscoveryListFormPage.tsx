import { useState, useEffect } from 'react';
import { useParams, useNavigate, Link } from 'react-router-dom';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { ArrowLeft, Plus, X, ChevronUp, ChevronDown, Search } from 'lucide-react';
import {
  Container,
  Card,
  CardHeader,
  CardBody,
  Button,
  Spinner,
  Input,
  TextArea,
} from '../../components';
import { discoveryListApi } from '../../lib/discovery-list-api';
import { searchApi } from '../../lib/search-api';
import { useToast } from '../../context/ToastContext';
import type { Clip } from '../../types/clip';

// Constants
const MAX_CLIPS_PER_LIST = 200;

export function AdminDiscoveryListFormPage() {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const { showToast } = useToast();
  const queryClient = useQueryClient();
  const isEditing = id !== 'new';

  // Form state
  const [name, setName] = useState('');
  const [description, setDescription] = useState('');
  const [isFeatured, setIsFeatured] = useState(false);
  const [clips, setClips] = useState<Clip[]>([]);

  // Clip search state
  const [searchQuery, setSearchQuery] = useState('');
  const [showClipSearch, setShowClipSearch] = useState(false);

  // Fetch existing list if editing
  const { data: existingList, isLoading: isLoadingList } = useQuery({
    queryKey: ['admin', 'discovery-list', id],
    queryFn: () => discoveryListApi.getDiscoveryList(id!),
    enabled: isEditing,
  });

  // Fetch clips for the list if editing
  const { data: listClipsData, isLoading: isLoadingClips } = useQuery({
    queryKey: ['admin', 'discovery-list-clips', id],
    queryFn: () => discoveryListApi.getDiscoveryListClips(id!, { limit: MAX_CLIPS_PER_LIST }),
    enabled: isEditing,
  });

  // Search clips
  const { data: searchResults, isLoading: isSearching } = useQuery({
    queryKey: ['clip-search', searchQuery],
    queryFn: () => searchApi.search({ query: searchQuery, type: 'clips', limit: 20 }),
    enabled: searchQuery.length >= 3,
  });

  // Populate form with existing data
  useEffect(() => {
    if (existingList) {
      queueMicrotask(() => {
        setName(existingList.name);
        setDescription(existingList.description || '');
        setIsFeatured(existingList.is_featured);
      });
    }
  }, [existingList]);

  useEffect(() => {
    if (listClipsData?.clips) {
      queueMicrotask(() => {
        setClips(listClipsData.clips);
      });
    }
  }, [listClipsData]);

  // Create mutation
  const createMutation = useMutation({
    mutationFn: (data: { name: string; description?: string; is_featured: boolean }) =>
      discoveryListApi.admin.createDiscoveryList(data),
    onSuccess: (newList) => {
      queryClient.invalidateQueries({ queryKey: ['admin', 'discovery-lists'] });
      showToast('Discovery list created successfully', 'success');
      // Navigate to edit page to add clips
      navigate(`/admin/discovery-lists/${newList.id}/edit`);
    },
    onError: () => {
      showToast('Failed to create discovery list', 'error');
    },
  });

  // Update mutation
  const updateMutation = useMutation({
    mutationFn: (data: {
      name?: string;
      description?: string;
      is_featured?: boolean;
    }) => discoveryListApi.admin.updateDiscoveryList(id!, data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['admin', 'discovery-lists'] });
      queryClient.invalidateQueries({ queryKey: ['admin', 'discovery-list', id] });
      showToast('Discovery list updated successfully', 'success');
    },
    onError: () => {
      showToast('Failed to update discovery list', 'error');
    },
  });

  // Add clip mutation
  const addClipMutation = useMutation({
    mutationFn: (clipId: string) => discoveryListApi.admin.addClipToList(id!, clipId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['admin', 'discovery-list-clips', id] });
      showToast('Clip added to list', 'success');
      setSearchQuery('');
      setShowClipSearch(false);
    },
    onError: () => {
      showToast('Failed to add clip to list', 'error');
    },
  });

  // Remove clip mutation
  const removeClipMutation = useMutation({
    mutationFn: (clipId: string) => discoveryListApi.admin.removeClipFromList(id!, clipId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['admin', 'discovery-list-clips', id] });
      showToast('Clip removed from list', 'success');
    },
    onError: () => {
      showToast('Failed to remove clip from list', 'error');
    },
  });

  // Reorder clips mutation
  const reorderMutation = useMutation({
    mutationFn: (clipIds: string[]) =>
      discoveryListApi.admin.reorderListClips(id!, clipIds),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['admin', 'discovery-list-clips', id] });
      showToast('Clips reordered successfully', 'success');
    },
    onError: () => {
      showToast('Failed to reorder clips', 'error');
    },
  });

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();

    if (!name.trim()) {
      showToast('Please enter a list name', 'error');
      return;
    }

    const data = {
      name: name.trim(),
      description: description.trim() || undefined,
      is_featured: isFeatured,
    };

    if (isEditing) {
      updateMutation.mutate(data);
    } else {
      createMutation.mutate(data);
    }
  };

  const handleAddClip = (clip: Clip) => {
    if (!isEditing) {
      showToast('Please save the list first before adding clips', 'info');
      return;
    }
    addClipMutation.mutate(clip.id);
  };

  const handleRemoveClip = (clipId: string) => {
    removeClipMutation.mutate(clipId);
  };

  const moveClip = (index: number, direction: 'up' | 'down') => {
    const newClips = [...clips];
    const targetIndex = direction === 'up' ? index - 1 : index + 1;

    if (targetIndex < 0 || targetIndex >= newClips.length) return;

    [newClips[index], newClips[targetIndex]] = [newClips[targetIndex], newClips[index]];
    setClips(newClips);

    // Save the new order
    const clipIds = newClips.map((c) => c.id);
    reorderMutation.mutate(clipIds);
  };

  if (isEditing && (isLoadingList || isLoadingClips)) {
    return (
      <Container className="py-8 flex justify-center">
        <Spinner size="xl" />
      </Container>
    );
  }

  return (
    <Container className="py-8 max-w-4xl">
      <div className="mb-6">
        <Link to="/admin/discovery-lists" className="flex items-center text-sm text-muted-foreground hover:text-foreground mb-2">
          <ArrowLeft className="w-4 h-4 mr-1" />
          Back to Discovery Lists
        </Link>
        <h1 className="text-3xl font-bold">
          {isEditing ? 'Edit Discovery List' : 'Create Discovery List'}
        </h1>
      </div>

      <form onSubmit={handleSubmit} className="space-y-6">
        {/* List Details */}
        <Card>
          <CardHeader>
            <h2 className="text-xl font-semibold">List Details</h2>
          </CardHeader>
          <CardBody className="space-y-4">
            <div>
              <label htmlFor="name" className="block text-sm font-medium mb-2">
                Name <span className="text-red-500">*</span>
              </label>
              <Input
                id="name"
                value={name}
                onChange={(e) => setName(e.target.value)}
                placeholder="e.g., Wildest Live Moments"
                required
                maxLength={200}
              />
            </div>

            <div>
              <label htmlFor="description" className="block text-sm font-medium mb-2">
                Description
              </label>
              <TextArea
                id="description"
                value={description}
                onChange={(e) => setDescription(e.target.value)}
                placeholder="A brief description of this discovery list..."
                rows={3}
                maxLength={1000}
              />
            </div>

            <div className="flex items-center gap-2">
              <input
                type="checkbox"
                id="isFeatured"
                checked={isFeatured}
                onChange={(e) => setIsFeatured(e.target.checked)}
                className="w-4 h-4 rounded border-gray-300 text-primary focus:ring-primary"
              />
              <label htmlFor="isFeatured" className="text-sm font-medium">
                Featured List
              </label>
            </div>

            <div className="flex gap-3 pt-4">
              <Button
                type="submit"
                disabled={createMutation.isPending || updateMutation.isPending}
              >
                {createMutation.isPending || updateMutation.isPending ? (
                  <Spinner size="sm" className="mr-2" />
                ) : null}
                {isEditing ? 'Update List' : 'Create List'}
              </Button>
              <Button type="button" variant="outline" onClick={() => navigate('/admin/discovery-lists')}>
                Cancel
              </Button>
            </div>
          </CardBody>
        </Card>

        {/* Clip Management - Only show when editing */}
        {isEditing && (
          <Card>
            <CardHeader>
              <div className="flex justify-between items-center">
                <h2 className="text-xl font-semibold">Clips ({clips.length})</h2>
                <Button
                  type="button"
                  size="sm"
                  onClick={() => setShowClipSearch(!showClipSearch)}
                >
                  <Plus className="w-4 h-4 mr-2" />
                  Add Clip
                </Button>
              </div>
            </CardHeader>
            <CardBody>
              {/* Clip Search */}
              {showClipSearch && (
                <div className="mb-4 p-4 border border-border rounded-lg">
                  <div className="flex gap-2 mb-3">
                    <div className="flex-1 relative">
                      <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 w-4 h-4 text-muted-foreground" />
                      <Input
                        type="text"
                        placeholder="Search clips by title or broadcaster..."
                        value={searchQuery}
                        onChange={(e) => setSearchQuery(e.target.value)}
                        className="pl-10"
                      />
                    </div>
                    <Button
                      type="button"
                      variant="ghost"
                      onClick={() => {
                        setShowClipSearch(false);
                        setSearchQuery('');
                      }}
                    >
                      Cancel
                    </Button>
                  </div>

                  {isSearching && (
                    <div className="flex justify-center py-4">
                      <Spinner />
                    </div>
                  )}

                  {searchQuery.length >= 3 && searchResults && searchResults.results.clips && searchResults.results.clips.length > 0 && (
                    <div className="space-y-2 max-h-80 overflow-y-auto">
                      {searchResults.results.clips.map((clip: Clip) => (
                        <div
                          key={clip.id}
                          className="flex items-center gap-3 p-2 hover:bg-accent rounded transition-colors"
                        >
                          <img
                            src={clip.thumbnail_url || ''}
                            alt={clip.title}
                            className="w-20 h-11 object-cover rounded"
                          />
                          <div className="flex-1 min-w-0">
                            <div className="font-medium text-sm truncate">{clip.title}</div>
                            <div className="text-xs text-muted-foreground">
                              {clip.broadcaster_name}
                            </div>
                          </div>
                          <Button
                            type="button"
                            size="sm"
                            onClick={() => handleAddClip(clip)}
                            disabled={clips.some((c) => c.id === clip.id)}
                          >
                            {clips.some((c) => c.id === clip.id) ? 'Added' : 'Add'}
                          </Button>
                        </div>
                      ))}
                    </div>
                  )}

                  {searchQuery.length >= 3 && searchResults && searchResults.results.clips && searchResults.results.clips.length === 0 && (
                    <div className="text-center py-4 text-muted-foreground">
                      No clips found
                    </div>
                  )}
                </div>
              )}

              {/* Current Clips */}
              {clips.length === 0 ? (
                <div className="text-center py-12 text-muted-foreground">
                  <p>No clips in this list yet</p>
                  <p className="text-sm mt-1">Click "Add Clip" to start adding clips</p>
                </div>
              ) : (
                <div className="space-y-2">
                  {clips.map((clip, index) => (
                    <div
                      key={clip.id}
                      className="flex items-center gap-3 p-3 border border-border rounded-lg hover:bg-accent transition-colors"
                    >
                      <div className="flex flex-col gap-1">
                        <button
                          type="button"
                          onClick={() => moveClip(index, 'up')}
                          disabled={index === 0}
                          className="text-muted-foreground hover:text-foreground disabled:opacity-30"
                          title="Move up"
                        >
                          <ChevronUp className="w-4 h-4" />
                        </button>
                        <button
                          type="button"
                          onClick={() => moveClip(index, 'down')}
                          disabled={index === clips.length - 1}
                          className="text-muted-foreground hover:text-foreground disabled:opacity-30"
                          title="Move down"
                        >
                          <ChevronDown className="w-4 h-4" />
                        </button>
                      </div>
                      <div className="text-sm text-muted-foreground w-8">{index + 1}</div>
                      <img
                        src={clip.thumbnail_url || ''}
                        alt={clip.title}
                        className="w-24 h-14 object-cover rounded"
                      />
                      <div className="flex-1 min-w-0">
                        <div className="font-medium truncate">{clip.title}</div>
                        <div className="text-sm text-muted-foreground">
                          {clip.broadcaster_name} • {clip.view_count.toLocaleString()} views
                        </div>
                      </div>
                      <Button
                        type="button"
                        variant="ghost"
                        size="sm"
                        onClick={() => handleRemoveClip(clip.id)}
                      >
                        <X className="w-4 h-4" />
                      </Button>
                    </div>
                  ))}
                </div>
              )}
            </CardBody>
          </Card>
        )}
      </form>
    </Container>
  );
}
