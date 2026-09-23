import { useState, useEffect, useCallback } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { Spinner } from '@/components/ui/Spinner';
import { Alert } from '@/components/ui/Alert';
import { Button } from '@/components/ui/Button';
import type { ChannelMember, ChannelRole } from '@/types/chat';
import {
  getCurrentUserRole,
  listChannelMembers,
  removeChannelMember,
  updateChannelMemberRole,
  deleteChannel,
} from '@/lib/chat-api';

/**
 * ChannelSettingsPage - Manage channel settings and members
 * Features:
 * - View channel members with roles
 * - Invite members (owner/admin only)
 * - Remove members (owner/admin only)
 * - Update member roles (owner only)
 * - Delete channel (owner only)
 */
export function ChannelSettingsPage() {
  const { id: channelId } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const [currentUserRole, setCurrentUserRole] = useState<ChannelRole | null>(null);
  const [members, setMembers] = useState<ChannelMember[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [actionLoading, setActionLoading] = useState<string | null>(null);

  const fetchData = useCallback(async () => {
    if (!channelId) return;

    try {
      setLoading(true);
      setError(null);

      // Fetch current user's role and members in parallel
      const [roleData, membersData] = await Promise.all([
        getCurrentUserRole(channelId),
        listChannelMembers(channelId),
      ]);

      setCurrentUserRole(roleData.role);
      setMembers(membersData.members);
    } catch (err) {
      console.error('Error fetching channel data:', err);
      setError(err instanceof Error ? err.message : 'Failed to load channel data');
    } finally {
      setLoading(false);
    }
  }, [channelId]);

  useEffect(() => {
    if (!channelId) {
      setError('Channel ID is required');
      setLoading(false);
      return;
    }

    fetchData();
  }, [fetchData, channelId]);

  const handleRemoveMember = async (userId: string) => {
    if (!channelId) return;

    try {
      setActionLoading(`remove-${userId}`);
      await removeChannelMember(channelId, userId);
      await fetchData();
    } catch (err) {
      console.error('Error removing member:', err);
      setError(err instanceof Error ? err.message : 'Failed to remove member');
    } finally {
      setActionLoading(null);
    }
  };

  const handleUpdateRole = async (userId: string, newRole: Exclude<ChannelRole, 'owner'>) => {
    if (!channelId) return;

    try {
      setActionLoading(`role-${userId}`);
      await updateChannelMemberRole(channelId, userId, newRole);
      await fetchData();
    } catch (err) {
      console.error('Error updating role:', err);
      setError(err instanceof Error ? err.message : 'Failed to update role');
    } finally {
      setActionLoading(null);
    }
  };

  const handleDeleteChannel = async () => {
    if (!channelId) return;
    if (!window.confirm('Are you sure you want to delete this channel? This action cannot be undone.')) {
      return;
    }

    try {
      setActionLoading('delete-channel');
      await deleteChannel(channelId);
      navigate('/chat');
    } catch (err) {
      console.error('Error deleting channel:', err);
      setError(err instanceof Error ? err.message : 'Failed to delete channel');
      setActionLoading(null);
    }
  };

  const canInvite = currentUserRole === 'owner' || currentUserRole === 'admin';
  const canRemove = currentUserRole === 'owner' || currentUserRole === 'admin';
  const canUpdateRoles = currentUserRole === 'owner';
  const canDelete = currentUserRole === 'owner';

  if (loading) {
    return (
      <div className="min-h-screen flex items-center justify-center">
        <Spinner size="xl" />
      </div>
    );
  }

  if (error && !members.length) {
    return (
      <div className="min-h-screen p-8">
        <Alert variant="error" title="Error">
          {error}
        </Alert>
        <Button onClick={() => navigate('/chat')} className="mt-4">
          Back to Chat
        </Button>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-background">
      <div className="max-w-4xl mx-auto p-8">
        <div className="mb-8">
          <h1 className="text-3xl font-bold mb-2">Channel Settings</h1>
          <p className="text-muted-foreground">
            Your role: <span className="font-semibold capitalize">{currentUserRole}</span>
          </p>
        </div>

        {error && (
          <Alert variant="error" title="Error" className="mb-4">
            {error}
          </Alert>
        )}

        {/* Members Section */}
        <div className="bg-card rounded-lg shadow-sm p-6 mb-6">
          <div className="flex items-center justify-between mb-4">
            <h2 className="text-xl font-semibold">Members ({members.length})</h2>
            {canInvite && (
              <Button size="sm" disabled>
                Invite Member
              </Button>
            )}
          </div>

          <div className="space-y-2">
            {members.map((member) => (
              <div
                key={member.id}
                className="flex items-center justify-between p-3 bg-muted/50 rounded-lg"
              >
                <div className="flex items-center gap-3">
                  {member.avatar_url ? (
                    <img
                      src={member.avatar_url}
                      alt={member.username}
                      className="w-10 h-10 rounded-full"
                    />
                  ) : (
                    <div className="w-10 h-10 rounded-full bg-primary-400 flex items-center justify-center text-background font-heading font-bold uppercase tracking-[0.06em]">
                      {member.username?.[0]?.toUpperCase() || '?'}
                    </div>
                  )}
                  <div>
                    <div className="font-medium">{member.display_name || member.username}</div>
                    <div className="text-sm text-muted-foreground">@{member.username}</div>
                  </div>
                </div>

                <div className="flex items-center gap-2">
                  {canUpdateRoles && member.role !== 'owner' ? (
                    <select
                      value={member.role}
                      onChange={(e) => {
                        const value = e.target.value;
                        if (value === 'member' || value === 'moderator' || value === 'admin') {
                          handleUpdateRole(member.user_id, value);
                        }
                      }}
                      disabled={actionLoading === `role-${member.user_id}`}
                      className="px-3 py-1 bg-background border border-border rounded text-sm"
                    >
                      <option value="member">Member</option>
                      <option value="moderator">Moderator</option>
                      <option value="admin">Admin</option>
                    </select>
                  ) : (
                    <span className="px-3 py-1 bg-background border border-border rounded text-sm capitalize">
                      {member.role}
                    </span>
                  )}

                  {canRemove && member.role !== 'owner' && (
                    <Button
                      size="sm"
                      variant="danger"
                      onClick={() => handleRemoveMember(member.user_id)}
                      disabled={actionLoading === `remove-${member.user_id}`}
                    >
                      {actionLoading === `remove-${member.user_id}` ? 'Removing...' : 'Remove'}
                    </Button>
                  )}
                </div>
              </div>
            ))}
          </div>
        </div>

        {/* Danger Zone */}
        {canDelete && (
          <div className="bg-card rounded-lg shadow-sm p-6 border-2 border-error-500">
            <h2 className="text-xl font-semibold text-error-600 mb-4">Danger Zone</h2>
            <p className="text-muted-foreground mb-4">
              Deleting this channel will permanently remove all messages and member data. This action
              cannot be undone.
            </p>
            <Button
              variant="danger"
              onClick={handleDeleteChannel}
              disabled={actionLoading === 'delete-channel'}
            >
              {actionLoading === 'delete-channel' ? 'Deleting...' : 'Delete Channel'}
            </Button>
          </div>
        )}

        <div className="mt-6">
          <Button variant="secondary" onClick={() => navigate('/chat')}>
            Back to Chat
          </Button>
        </div>
      </div>
    </div>
  );
}
