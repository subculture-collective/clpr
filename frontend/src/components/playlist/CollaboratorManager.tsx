import { useState, useEffect, useCallback } from 'react';
import { UserPlus, Shield, Edit, Eye, Trash2 } from 'lucide-react';
import apiClient from '@/lib/api';
import type { PlaylistCollaborator, AddCollaboratorRequest } from '@/types/playlist';

interface CollaboratorManagerProps {
    playlistId: string;
    isOwner: boolean;
    canManageCollaborators: boolean;
}

export function CollaboratorManager({ playlistId, isOwner, canManageCollaborators }: CollaboratorManagerProps) {
    const [collaborators, setCollaborators] = useState<PlaylistCollaborator[]>([]);
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState<string | null>(null);
    const [showAddForm, setShowAddForm] = useState(false);
    const [newCollaboratorUserId, setNewCollaboratorUserId] = useState('');
    const [newCollaboratorPermission, setNewCollaboratorPermission] = useState<'view' | 'edit' | 'admin'>('edit');
    const [submitting, setSubmitting] = useState(false);

    const fetchCollaborators = useCallback(async () => {
        try {
            setLoading(true);
            setError(null); // Clear any previous errors
            const response = await apiClient.get<{ success: boolean; data: PlaylistCollaborator[]; error?: { message: string } }>(
                `/playlists/${playlistId}/collaborators`
            );

            if (response.data.success) {
                setCollaborators(response.data.data || []);
            } else {
                throw new Error(response.data.error?.message || 'Failed to fetch collaborators');
            }
        } catch (err) {
            setError(err instanceof Error ? err.message : 'An error occurred');
        } finally {
            setLoading(false);
        }
    }, [playlistId]);

    useEffect(() => {
        fetchCollaborators();
    }, [fetchCollaborators]);

    const addCollaborator = async () => {
        if (!newCollaboratorUserId.trim()) {
            setError('Please enter a user ID');
            return;
        }

        try {
            setSubmitting(true);
            setError(null);

            const requestBody: AddCollaboratorRequest = {
                user_id: newCollaboratorUserId.trim(),
                permission: newCollaboratorPermission,
            };

            await apiClient.post(
                `/playlists/${playlistId}/collaborators`,
                requestBody
            );

            // Reset form and refresh list
            setNewCollaboratorUserId('');
            setNewCollaboratorPermission('edit');
            setShowAddForm(false);
            await fetchCollaborators();
        } catch (err) {
            setError(err instanceof Error ? err.message : 'An error occurred');
        } finally {
            setSubmitting(false);
        }
    };

    const removeCollaborator = async (userId: string) => {
        if (!confirm('Are you sure you want to remove this collaborator?')) {
            return;
        }

        try {
            setError(null); // Clear any previous errors
            await apiClient.delete(`/playlists/${playlistId}/collaborators/${userId}`);
            await fetchCollaborators();
        } catch (err) {
            setError(err instanceof Error ? err.message : 'An error occurred');
        }
    };

    const updatePermission = async (userId: string, newPermission: 'view' | 'edit' | 'admin') => {
        try {
            setError(null); // Clear any previous errors
            await apiClient.patch(
                `/playlists/${playlistId}/collaborators/${userId}`,
                { permission: newPermission }
            );
            await fetchCollaborators();
        } catch (err) {
            setError(err instanceof Error ? err.message : 'An error occurred');
        }
    };

    const getPermissionIcon = (permission: string) => {
        switch (permission) {
            case 'admin':
                return <Shield className="h-4 w-4 text-red-400" />;
            case 'edit':
                return <Edit className="h-4 w-4 text-blue-400" />;
            case 'view':
                return <Eye className="h-4 w-4 text-green-400" />;
            default:
                return null;
        }
    };

    const getPermissionLabel = (permission: string) => {
        switch (permission) {
            case 'admin':
                return 'Admin';
            case 'edit':
                return 'Can Edit';
            case 'view':
                return 'View Only';
            default:
                return permission;
        }
    };

    return (
        <div className="bg-surface rounded-lg border border-border p-6">
            {/* Header */}
            <div className="flex items-center justify-between mb-6">
                <div>
                    <h3 className="text-lg font-bold text-white">Collaborators</h3>
                    <p className="text-sm text-muted-foreground mt-1">
                        {collaborators.length === 0
                            ? 'No collaborators yet'
                            : `${collaborators.length} collaborator${collaborators.length === 1 ? '' : 's'}`
                        }
                    </p>
                </div>
                {canManageCollaborators && (
                    <button
                        onClick={() => setShowAddForm(!showAddForm)}
                        className="px-4 py-2 bg-purple-600 hover:bg-purple-700 text-white rounded-lg transition-colors flex items-center gap-2"
                    >
                        <UserPlus className="h-4 w-4" />
                        Add
                    </button>
                )}
            </div>

            {/* Error Display */}
            {error && (
                <div className="mb-4 bg-red-500/10 border border-red-500/20 rounded-lg p-3">
                    <p className="text-sm text-red-400">{error}</p>
                </div>
            )}

            {/* Add Collaborator Form */}
            {showAddForm && canManageCollaborators && (
                <div className="mb-6 bg-surface-raised rounded-lg p-4 border border-border">
                    <div className="space-y-3">
                        <div>
                            <label className="block text-sm font-medium text-muted-foreground mb-2">
                                User ID
                            </label>
                            <input
                                type="text"
                                value={newCollaboratorUserId}
                                onChange={(e) => setNewCollaboratorUserId(e.target.value)}
                                placeholder="Enter user UUID"
                                className="w-full bg-surface text-white px-3 py-2 rounded-lg border border-border focus:outline-none focus:border-purple-500"
                                disabled={submitting}
                            />
                            <p className="text-xs text-text-secondary mt-1">
                                Enter the user's UUID. You can find this in their profile URL.
                            </p>
                        </div>
                        <div>
                            <label className="block text-sm font-medium text-muted-foreground mb-2">
                                Permission Level
                            </label>
                            <select
                                value={newCollaboratorPermission}
                                onChange={(e) => setNewCollaboratorPermission(e.target.value as 'view' | 'edit' | 'admin')}
                                className="w-full bg-surface text-white px-3 py-2 rounded-lg border border-border focus:outline-none focus:border-purple-500"
                                disabled={submitting}
                            >
                                <option value="view">View Only</option>
                                <option value="edit">Can Edit</option>
                                {isOwner && <option value="admin">Admin</option>}
                            </select>
                        </div>
                        <div className="flex gap-2">
                            <button
                                onClick={addCollaborator}
                                disabled={submitting}
                                className="flex-1 px-4 py-2 bg-purple-600 hover:bg-purple-700 disabled:bg-surface-raised disabled:cursor-not-allowed text-white rounded-lg transition-colors"
                            >
                                {submitting ? 'Adding...' : 'Add Collaborator'}
                            </button>
                            <button
                                onClick={() => {
                                    setShowAddForm(false);
                                    setError(null);
                                }}
                                disabled={submitting}
                                className="px-4 py-2 bg-surface-raised hover:bg-surface-hover disabled:cursor-not-allowed text-white rounded-lg transition-colors"
                            >
                                Cancel
                            </button>
                        </div>
                    </div>
                </div>
            )}

            {/* Collaborators List */}
            {loading ? (
                <div className="text-center py-8">
                    <div className="inline-block h-8 w-8 animate-spin rounded-full border-4 border-solid border-purple-500 border-r-transparent"></div>
                    <p className="mt-4 text-muted-foreground">Loading collaborators...</p>
                </div>
            ) : collaborators.length === 0 ? (
                <div className="text-center py-8">
                    <UserPlus className="h-12 w-12 text-text-tertiary mx-auto mb-3" />
                    <p className="text-muted-foreground">No collaborators yet</p>
                    {canManageCollaborators && (
                        <p className="text-sm text-text-secondary mt-2">
                            Add collaborators to work on this playlist together
                        </p>
                    )}
                </div>
            ) : (
                <div className="space-y-2">
                    {collaborators.map((collab) => (
                        <div
                            key={collab.id}
                            className="flex items-center justify-between p-3 bg-surface-raised rounded-lg border border-border"
                        >
                            <div className="flex items-center gap-3 flex-1 min-w-0">
                                {/* Avatar */}
                                {collab.user?.avatar_url ? (
                                    <img
                                        src={collab.user.avatar_url}
                                        alt={collab.user.display_name}
                                        className="w-10 h-10 rounded-full"
                                    />
                                ) : (
                                    <div className="w-10 h-10 rounded-full bg-gradient-to-br from-purple-500 to-blue-500 flex items-center justify-center text-white font-bold">
                                        {collab.user?.display_name?.charAt(0) || '?'}
                                    </div>
                                )}

                                {/* User Info */}
                                <div className="flex-1 min-w-0">
                                    <p className="font-medium text-white truncate">
                                        {collab.user?.display_name || 'Unknown User'}
                                    </p>
                                    <p className="text-sm text-muted-foreground truncate">
                                        @{collab.user?.username || collab.user_id}
                                    </p>
                                </div>
                            </div>

                            {/* Permission Badge and Actions */}
                            <div className="flex items-center gap-2">
                                {canManageCollaborators ? (
                                    <select
                                        value={collab.permission}
                                        onChange={(e) => updatePermission(collab.user_id, e.target.value as 'view' | 'edit' | 'admin')}
                                        className="bg-surface text-white px-3 py-1 rounded text-sm border border-border focus:outline-none focus:border-purple-500"
                                    >
                                        <option value="view">View Only</option>
                                        <option value="edit">Can Edit</option>
                                        {isOwner && <option value="admin">Admin</option>}
                                    </select>
                                ) : (
                                    <div className="flex items-center gap-2 px-3 py-1 bg-surface rounded text-sm">
                                        {getPermissionIcon(collab.permission)}
                                        <span className="text-foreground">{getPermissionLabel(collab.permission)}</span>
                                    </div>
                                )}

                                {canManageCollaborators && (
                                    <button
                                        onClick={() => removeCollaborator(collab.user_id)}
                                        className="p-1 text-muted-foreground hover:text-red-400 transition-colors"
                                        title="Remove collaborator"
                                    >
                                        <Trash2 className="h-4 w-4" />
                                    </button>
                                )}
                            </div>
                        </div>
                    ))}
                </div>
            )}
        </div>
    );
}
