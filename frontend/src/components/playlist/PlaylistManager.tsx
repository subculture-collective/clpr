import { Button, Spinner } from '@/components/ui';
import { useToast } from '@/hooks';
import {
    useCreatePlaylist,
    useDeletePlaylist,
    usePlaylists,
    useUpdatePlaylist,
} from '@/hooks/usePlaylist';
import { fetchBookmarkedPlaylists } from '@/lib/playlist-api';
import type { CreatePlaylistRequest, Playlist } from '@/types/playlist';
import { Plus, Trash2, Edit2, ListMusic, Bookmark } from 'lucide-react';
import { useState } from 'react';
import { useQuery } from '@tanstack/react-query';
import { PlaylistCard } from './PlaylistCard';

type Tab = 'mine' | 'bookmarked';

export function PlaylistManager() {
    const [activeTab, setActiveTab] = useState<Tab>('mine');

    const { data, isLoading } = usePlaylists();
    const { data: bookmarksData, isLoading: bookmarksLoading } = useQuery({
        queryKey: ['playlists', 'bookmarked'],
        queryFn: () => fetchBookmarkedPlaylists(1, 50),
        enabled: activeTab === 'bookmarked',
    });

    const createMutation = useCreatePlaylist();
    const updateMutation = useUpdatePlaylist();
    const deleteMutation = useDeletePlaylist();
    const toast = useToast();

    const [showCreateForm, setShowCreateForm] = useState(false);
    const [editingPlaylist, setEditingPlaylist] = useState<Playlist | null>(null);
    const [formData, setFormData] = useState<CreatePlaylistRequest>({
        title: '',
        description: '',
        visibility: 'private',
    });

    const handleCreate = async (e: React.FormEvent) => {
        e.preventDefault();
        if (!formData.title.trim()) {
            toast.error('Playlist title is required');
            return;
        }

        try {
            await createMutation.mutateAsync(formData);
            toast.success('Playlist created successfully');
            setShowCreateForm(false);
            setFormData({ title: '', description: '', visibility: 'private' });
        } catch {
            toast.error('Failed to create playlist');
        }
    };

    const handleUpdate = async (e: React.FormEvent) => {
        e.preventDefault();
        if (!editingPlaylist) return;

        try {
            await updateMutation.mutateAsync({
                id: editingPlaylist.id,
                data: formData,
            });
            toast.success('Playlist updated successfully');
            setEditingPlaylist(null);
            setFormData({ title: '', description: '', visibility: 'private' });
        } catch {
            toast.error('Failed to update playlist');
        }
    };

    const handleDelete = async (id: string) => {
        if (!confirm('Are you sure you want to delete this playlist?')) {
            return;
        }

        try {
            await deleteMutation.mutateAsync(id);
            toast.success('Playlist deleted successfully');
        } catch {
            toast.error('Failed to delete playlist');
        }
    };

    const startEdit = (playlist: Playlist) => {
        setEditingPlaylist(playlist);
        setFormData({
            title: playlist.title,
            description: playlist.description || '',
            visibility: playlist.visibility,
        });
        setShowCreateForm(false);
    };

    const myPlaylists = data?.data || [];
    const bookmarkedPlaylists = bookmarksData?.playlists || [];

    return (
        <div className="space-y-6">
            {/* Header with tabs */}
            <div className="flex items-center justify-between">
                <div className="flex items-center gap-1">
                    <button
                        type="button"
                        onClick={() => setActiveTab('mine')}
                        className={`flex items-center gap-2 px-4 py-2 text-sm font-medium rounded-lg transition-colors cursor-pointer ${
                            activeTab === 'mine'
                                ? 'bg-brand text-white'
                                : 'text-muted-foreground hover:text-foreground hover:bg-muted'
                        }`}
                    >
                        <ListMusic className="h-4 w-4" />
                        My Playlists
                    </button>
                    <button
                        type="button"
                        onClick={() => setActiveTab('bookmarked')}
                        className={`flex items-center gap-2 px-4 py-2 text-sm font-medium rounded-lg transition-colors cursor-pointer ${
                            activeTab === 'bookmarked'
                                ? 'bg-brand text-white'
                                : 'text-muted-foreground hover:text-foreground hover:bg-muted'
                        }`}
                    >
                        <Bookmark className="h-4 w-4" />
                        Bookmarked
                    </button>
                </div>

                {activeTab === 'mine' && (
                    <Button
                        onClick={() => {
                            setShowCreateForm(!showCreateForm);
                            setEditingPlaylist(null);
                            setFormData({ title: '', description: '', visibility: 'private' });
                        }}
                    >
                        <Plus className="h-4 w-4 mr-2" />
                        Create Playlist
                    </Button>
                )}
            </div>

            {/* Create/Edit Form (only on mine tab) */}
            {activeTab === 'mine' && (showCreateForm || editingPlaylist) && (
                <div className="bg-card border border-border rounded-lg p-6">
                    <h3 className="text-lg font-semibold text-foreground mb-4">
                        {editingPlaylist ? 'Edit Playlist' : 'Create New Playlist'}
                    </h3>
                    <form onSubmit={editingPlaylist ? handleUpdate : handleCreate} className="space-y-4">
                        <div>
                            <label className="block text-sm font-medium text-muted-foreground mb-2">
                                Title <span className="text-error-500">*</span>
                            </label>
                            <input
                                type="text"
                                value={formData.title}
                                onChange={(e) =>
                                    setFormData({ ...formData, title: e.target.value })
                                }
                                maxLength={100}
                                className="w-full px-3 py-2 bg-background border border-border rounded-md text-foreground focus:outline-none focus:ring-2 focus:ring-primary-500"
                                required
                            />
                        </div>

                        <div>
                            <label className="block text-sm font-medium text-muted-foreground mb-2">
                                Description
                            </label>
                            <textarea
                                value={formData.description}
                                onChange={(e) =>
                                    setFormData({ ...formData, description: e.target.value })
                                }
                                maxLength={500}
                                rows={3}
                                className="w-full px-3 py-2 bg-background border border-border rounded-md text-foreground focus:outline-none focus:ring-2 focus:ring-primary-500"
                            />
                        </div>

                        <div>
                            <label className="block text-sm font-medium text-muted-foreground mb-2">
                                Visibility
                            </label>
                            <select
                                value={formData.visibility}
                                onChange={(e) =>
                                    setFormData({
                                        ...formData,
                                        visibility: e.target.value as 'private' | 'public' | 'unlisted',
                                    })
                                }
                                className="w-full px-3 py-2 bg-background border border-border rounded-md text-foreground focus:outline-none focus:ring-2 focus:ring-primary-500"
                            >
                                <option value="private">Private - Only you can see it</option>
                                <option value="unlisted">Unlisted - Anyone with link can see it</option>
                                <option value="public">Public - Anyone can discover it</option>
                            </select>
                        </div>

                        <div className="flex gap-2">
                            <Button type="submit" disabled={createMutation.isPending || updateMutation.isPending}>
                                {editingPlaylist ? 'Update' : 'Create'}
                            </Button>
                            <Button
                                type="button"
                                variant="outline"
                                onClick={() => {
                                    setShowCreateForm(false);
                                    setEditingPlaylist(null);
                                    setFormData({ title: '', description: '', visibility: 'private' });
                                }}
                            >
                                Cancel
                            </Button>
                        </div>
                    </form>
                </div>
            )}

            {/* My Playlists tab */}
            {activeTab === 'mine' && (
                <>
                    {isLoading ? (
                        <div className="flex justify-center py-12">
                            <Spinner size="lg" />
                        </div>
                    ) : myPlaylists.length === 0 ? (
                        <div className="text-center py-12 text-muted-foreground">
                            <ListMusic className="h-12 w-12 mx-auto mb-3 opacity-40" />
                            <p>You don't have any playlists yet.</p>
                            <p className="text-sm mt-1">Click "Create Playlist" to get started!</p>
                        </div>
                    ) : (
                        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
                            {myPlaylists.map((playlist) => (
                                <div key={playlist.id} className="relative group">
                                    <PlaylistCard playlist={playlist} />
                                    <div className="absolute top-5 right-5 flex gap-2 opacity-0 group-hover:opacity-100 transition-opacity">
                                        <button
                                            onClick={(e) => {
                                                e.preventDefault();
                                                startEdit(playlist);
                                            }}
                                            className="p-2 bg-background/90 rounded-lg hover:bg-surface-hover transition cursor-pointer"
                                        >
                                            <Edit2 className="h-4 w-4 text-muted-foreground" />
                                        </button>
                                        <button
                                            onClick={(e) => {
                                                e.preventDefault();
                                                handleDelete(playlist.id);
                                            }}
                                            className="p-2 bg-background/90 rounded-lg hover:bg-error-900/50 transition cursor-pointer"
                                        >
                                            <Trash2 className="h-4 w-4 text-error-400" />
                                        </button>
                                    </div>
                                </div>
                            ))}
                        </div>
                    )}
                </>
            )}

            {/* Bookmarked tab */}
            {activeTab === 'bookmarked' && (
                <>
                    {bookmarksLoading ? (
                        <div className="flex justify-center py-12">
                            <Spinner size="lg" />
                        </div>
                    ) : bookmarkedPlaylists.length === 0 ? (
                        <div className="text-center py-12 text-muted-foreground">
                            <Bookmark className="h-12 w-12 mx-auto mb-3 opacity-40" />
                            <p>No bookmarked playlists yet.</p>
                            <p className="text-sm mt-1">
                                Bookmark playlists you enjoy to find them here later.
                            </p>
                        </div>
                    ) : (
                        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
                            {bookmarkedPlaylists.map((playlist) => (
                                <PlaylistCard key={playlist.id} playlist={playlist} />
                            ))}
                        </div>
                    )}
                </>
            )}
        </div>
    );
}
