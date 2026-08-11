import { useState } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { Container, Card, CardHeader, CardBody, Button, Input } from '../../components';
import { Search, Shield, Ban, TrendingUp, MessageSquare, MessageSquareOff, Eye } from 'lucide-react';
import axios, { AxiosError } from 'axios';
import { Modal } from '@/components/ui/Modal';

interface User {
  id: string;
  username: string;
  display_name: string;
  email: string;
  avatar_url: string;
  role: string;
  karma_points: number;
  is_banned: boolean;
  account_status: string;
  comment_suspended_until?: string;
  comments_require_review?: boolean;
  comment_warning_count?: number;
  created_at: string;
  last_login_at: string;
}

interface UsersResponse {
  users: User[];
  total: number;
  page: number;
  per_page: number;
}

interface UserActionModalProps {
  user: User;
  actionType: 'ban' | 'unban' | 'promote' | 'demote' | 'karma' | 'suspend_comments' | 'lift_suspension' | 'toggle_review';
  onClose: () => void;
  onConfirm: (params: {
    reason?: string;
    karmaValue?: number;
    suspensionType?: string;
    durationHours?: number;
    requireReview?: boolean;
  }) => void;
}

function UserActionModal({ user, actionType, onClose, onConfirm }: UserActionModalProps) {
  const [reason, setReason] = useState('');
  const [karmaValue, setKarmaValue] = useState(user.karma_points);
  const [suspensionType, setSuspensionType] = useState<'warning' | 'temporary' | 'permanent'>('warning');
  const [durationHours, setDurationHours] = useState(24);
  const [requireReview, setRequireReview] = useState(user.comments_require_review || false);

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    if (actionType === 'karma') {
      onConfirm({ karmaValue });
    } else if (actionType === 'suspend_comments') {
      onConfirm({
        reason,
        suspensionType,
        durationHours: suspensionType === 'temporary' ? durationHours : undefined,
      });
    } else if (actionType === 'toggle_review') {
      onConfirm({ reason, requireReview });
    } else {
      onConfirm({ reason });
    }
  };

  const titles: Record<string, string> = {
    ban: 'Ban User',
    unban: 'Unban User',
    promote: 'Promote User',
    demote: 'Demote User',
    karma: 'Adjust Uppies',
    suspend_comments: 'Suspend Comment Privileges',
    lift_suspension: 'Lift Comment Suspension',
    toggle_review: 'Toggle Comment Review'
  };

  return (
    <Modal open onClose={onClose} title={titles[actionType]} size="md">
          <p className="text-sm text-muted-foreground">
            User: {user.username} ({user.email})
          </p>
          <form onSubmit={handleSubmit} className="space-y-4">
            {actionType === 'karma' ? (
              <div>
                <label htmlFor="admin-user-karma" className="block text-sm font-medium mb-2">
                  New Uppies
                </label>
                <Input
                  id="admin-user-karma"
                  type="number"
                  value={karmaValue}
                  onChange={(e) => {
                    const parsed = parseInt(e.target.value, 10);
                    if (isNaN(parsed)) {
                      setKarmaValue(0);
                    } else {
                      const clamped = Math.min(999999, Math.max(0, parsed));
                      setKarmaValue(clamped);
                    }
                  }}
                  min={0}
                  max={999999}
                  className="w-full"
                  required
                />
                <p className="text-xs text-muted-foreground mt-1">
                  Current: {user.karma_points} points
                </p>
              </div>
            ) : actionType === 'suspend_comments' ? (
              <>
                <div>
                  <label htmlFor="admin-user-suspension-type" className="block text-sm font-medium mb-2">
                    Suspension Type
                  </label>
                  <select
                    id="admin-user-suspension-type"
                    value={suspensionType}
                    onChange={(e) => setSuspensionType(e.target.value as 'warning' | 'temporary' | 'permanent')}
                    className="w-full p-2 border border-border rounded-md bg-background"
                  >
                    <option value="warning">Warning (no suspension)</option>
                    <option value="temporary">Temporary Suspension</option>
                    <option value="permanent">Permanent Suspension</option>
                  </select>
                </div>
                {suspensionType === 'temporary' && (
                  <div>
                    <label htmlFor="admin-user-suspension-duration" className="block text-sm font-medium mb-2">
                      Duration (hours)
                    </label>
                    <Input
                      id="admin-user-suspension-duration"
                      type="number"
                      value={durationHours}
                      onChange={(e) => {
                        const value = parseInt(e.target.value, 10);
                        setDurationHours(isNaN(value) ? 24 : Math.max(1, Math.min(8760, value)));
                      }}
                      min={1}
                      max={8760}
                      className="w-full"
                      required
                    />
                    <p className="text-xs text-muted-foreground mt-1">
                      Common: 24h (1 day), 168h (1 week), 720h (30 days)
                    </p>
                  </div>
                )}
                <div>
                  <label htmlFor="admin-user-suspension-reason" className="block text-sm font-medium mb-2">
                    Reason (Required)
                  </label>
                  <textarea
                    id="admin-user-suspension-reason"
                    value={reason}
                    onChange={(e) => setReason(e.target.value)}
                    className="w-full p-2 border border-border rounded-md min-h-[100px] bg-background"
                    placeholder="Enter reason for suspension..."
                    minLength={10}
                    required
                  />
                </div>
              </>
            ) : actionType === 'toggle_review' ? (
              <>
                <div>
                  <label className="flex items-center space-x-2 cursor-pointer">
                    <input
                      type="checkbox"
                      checked={requireReview}
                      onChange={(e) => setRequireReview(e.target.checked)}
                      className="rounded border-border"
                    />
                    <span className="text-sm font-medium">
                      Require all comments to be reviewed before publishing
                    </span>
                  </label>
                  <p className="text-xs text-muted-foreground mt-2">
                    When enabled, all comments from this user will be queued for moderation before being visible.
                  </p>
                </div>
                <div>
                  <label htmlFor="admin-user-review-reason" className="block text-sm font-medium mb-2">
                    Reason (Required)
                  </label>
                  <textarea
                    id="admin-user-review-reason"
                    value={reason}
                    onChange={(e) => setReason(e.target.value)}
                    className="w-full p-2 border border-border rounded-md min-h-[100px] bg-background"
                    placeholder="Enter reason for this change..."
                    required
                  />
                </div>
              </>
            ) : (
              <div>
                <label htmlFor="admin-user-action-reason" className="block text-sm font-medium mb-2">
                  {actionType === 'ban' || actionType === 'suspend_comments' ? 'Reason (Required)' : 'Action Reason'}
                </label>
                <textarea
                  id="admin-user-action-reason"
                  value={reason}
                  onChange={(e) => setReason(e.target.value)}
                  className="w-full p-2 border border-border rounded-md min-h-[100px] bg-background"
                  placeholder={`Enter reason for ${actionType}...`}
                  required={actionType === 'ban' || actionType === 'lift_suspension'}
                  minLength={actionType === 'lift_suspension' ? 10 : undefined}
                />
              </div>
            )}
            <div className="flex gap-2 justify-end">
              <Button type="button" onClick={onClose} variant="outline">
                Cancel
              </Button>
              <Button type="submit" variant={actionType === 'ban' || actionType === 'suspend_comments' ? 'danger' : 'primary'}>
                Confirm
              </Button>
            </div>
          </form>
    </Modal>
  );
}

export function AdminUsersPage() {
  const [search, setSearch] = useState('');
  const [roleFilter, setRoleFilter] = useState('all');
  const [statusFilter, setStatusFilter] = useState('all');
  const [page, setPage] = useState(1);
  const [selectedUser, setSelectedUser] = useState<User | null>(null);
  const [actionType, setActionType] = useState<'ban' | 'unban' | 'promote' | 'demote' | 'karma' | 'suspend_comments' | 'lift_suspension' | 'toggle_review' | null>(null);
  const [errorMessage, setErrorMessage] = useState<string | null>(null);

  const queryClient = useQueryClient();
  const perPage = 25;

  // Fetch users
  const { data, isLoading, error } = useQuery<UsersResponse>({
    queryKey: ['admin-users', search, roleFilter, statusFilter, page],
    queryFn: async () => {
      const params = new URLSearchParams({
        page: page.toString(),
        per_page: perPage.toString(),
      });
      if (search) params.append('search', search);
      if (roleFilter !== 'all') params.append('role', roleFilter);
      if (statusFilter !== 'all') params.append('status', statusFilter);

      const response = await axios.get(`/api/v1/admin/users?${params.toString()}`);
      return response.data;
    },
  });

  // Ban user mutation
  const banMutation = useMutation({
    mutationFn: async ({ userId, reason }: { userId: string; reason: string }) => {
      await axios.post(`/api/v1/admin/users/${userId}/ban`, { reason });
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['admin-users'] });
      setSelectedUser(null);
      setActionType(null);
      setErrorMessage(null);
    },
    onError: (error: AxiosError<{ error?: string }>) => {
      setErrorMessage(error.response?.data?.error || 'Failed to ban user');
    },
  });

  // Unban user mutation
  const unbanMutation = useMutation({
    mutationFn: async ({ userId, reason }: { userId: string; reason?: string }) => {
      await axios.post(`/api/v1/admin/users/${userId}/unban`, { reason });
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['admin-users'] });
      setSelectedUser(null);
      setActionType(null);
      setErrorMessage(null);
    },
    onError: (error: AxiosError<{ error?: string }>) => {
      setErrorMessage(error.response?.data?.error || 'Failed to unban user');
    },
  });

  // Update role mutation
  const updateRoleMutation = useMutation({
    mutationFn: async ({ userId, role, reason }: { userId: string; role: string; reason?: string }) => {
      await axios.patch(`/api/v1/admin/users/${userId}/role`, { role, reason });
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['admin-users'] });
      setSelectedUser(null);
      setActionType(null);
      setErrorMessage(null);
    },
    onError: (error: AxiosError<{ error?: string }>) => {
      setErrorMessage(error.response?.data?.error || 'Failed to update user role');
    },
  });

  // Update karma mutation
  const updateKarmaMutation = useMutation({
    mutationFn: async ({ userId, karma }: { userId: string; karma: number }) => {
      await axios.patch(`/api/v1/admin/users/${userId}/karma`, { karma_points: karma });
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['admin-users'] });
      setSelectedUser(null);
      setActionType(null);
      setErrorMessage(null);
    },
    onError: (error: AxiosError<{ error?: string }>) => {
      setErrorMessage(error.response?.data?.error || 'Failed to update user uppies');
    },
  });

  // Suspend comment privileges mutation
  const suspendCommentsMutation = useMutation({
    mutationFn: async ({ userId, suspensionType, reason, durationHours }: {
      userId: string;
      suspensionType: string;
      reason: string;
      durationHours?: number;
    }) => {
      await axios.post(`/api/v1/admin/users/${userId}/suspend-comments`, {
        suspension_type: suspensionType,
        reason,
        duration_hours: durationHours,
      });
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['admin-users'] });
      setSelectedUser(null);
      setActionType(null);
      setErrorMessage(null);
    },
    onError: (error: AxiosError<{ error?: string }>) => {
      setErrorMessage(error.response?.data?.error || 'Failed to suspend comment privileges');
    },
  });

  // Lift comment suspension mutation
  const liftSuspensionMutation = useMutation({
    mutationFn: async ({ userId, reason }: { userId: string; reason: string }) => {
      await axios.post(`/api/v1/admin/users/${userId}/lift-comment-suspension`, { reason });
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['admin-users'] });
      setSelectedUser(null);
      setActionType(null);
      setErrorMessage(null);
    },
    onError: (error: AxiosError<{ error?: string }>) => {
      setErrorMessage(error.response?.data?.error || 'Failed to lift comment suspension');
    },
  });

  // Toggle comment review requirement mutation
  const toggleReviewMutation = useMutation({
    mutationFn: async ({ userId, requireReview, reason }: {
      userId: string;
      requireReview: boolean;
      reason: string;
    }) => {
      await axios.post(`/api/v1/admin/users/${userId}/toggle-comment-review`, {
        require_review: requireReview,
        reason,
      });
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['admin-users'] });
      setSelectedUser(null);
      setActionType(null);
      setErrorMessage(null);
    },
    onError: (error: AxiosError<{ error?: string }>) => {
      setErrorMessage(error.response?.data?.error || 'Failed to toggle comment review');
    },
  });

  const handleAction = (user: User, action: 'ban' | 'unban' | 'promote' | 'demote' | 'karma' | 'suspend_comments' | 'lift_suspension' | 'toggle_review') => {
    setSelectedUser(user);
    setActionType(action);
  };

  const handleConfirmAction = (params: {
    reason?: string;
    karmaValue?: number;
    suspensionType?: string;
    durationHours?: number;
    requireReview?: boolean;
  }) => {
    if (!selectedUser) return;

    if (actionType === 'ban') {
      banMutation.mutate({ userId: selectedUser.id, reason: params.reason || '' });
    } else if (actionType === 'unban') {
      unbanMutation.mutate({ userId: selectedUser.id, reason: params.reason });
    } else if (actionType === 'promote') {
      const newRole = selectedUser.role === 'user' ? 'moderator' : 'admin';
      updateRoleMutation.mutate({ userId: selectedUser.id, role: newRole, reason: params.reason });
    } else if (actionType === 'demote') {
      const newRole = selectedUser.role === 'admin' ? 'moderator' : 'user';
      updateRoleMutation.mutate({ userId: selectedUser.id, role: newRole, reason: params.reason });
    } else if (actionType === 'karma' && params.karmaValue !== undefined) {
      updateKarmaMutation.mutate({ userId: selectedUser.id, karma: params.karmaValue });
    } else if (actionType === 'suspend_comments' && params.suspensionType) {
      suspendCommentsMutation.mutate({
        userId: selectedUser.id,
        suspensionType: params.suspensionType,
        reason: params.reason || '',
        durationHours: params.durationHours,
      });
    } else if (actionType === 'lift_suspension') {
      liftSuspensionMutation.mutate({ userId: selectedUser.id, reason: params.reason || '' });
    } else if (actionType === 'toggle_review' && params.requireReview !== undefined) {
      toggleReviewMutation.mutate({
        userId: selectedUser.id,
        requireReview: params.requireReview,
        reason: params.reason || '',
      });
    }
  };

  const totalPages = data ? Math.ceil(data.total / perPage) : 0;

  return (
    <Container className="py-8">
      <div className="mb-6">
        <h1 className="text-3xl font-bold mb-2">User Management</h1>
        <p className="text-muted-foreground">Manage user accounts, roles, and permissions</p>
      </div>

      {/* Error Message */}
      {errorMessage && (
        <Card className="mb-6">
          <CardBody>
            <div className="bg-destructive/10 border border-destructive text-destructive px-4 py-3 rounded">
              <p className="font-bold">Error</p>
              <p>{errorMessage}</p>
              <button
                onClick={() => setErrorMessage(null)}
                className="mt-2 text-sm underline"
              >
                Dismiss
              </button>
            </div>
          </CardBody>
        </Card>
      )}

      {/* Filters */}
      <Card className="mb-6">
        <CardBody>
          <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
            <div className="md:col-span-2">
              <div className="relative">
                <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 text-muted-foreground w-4 h-4" />
                <Input
                  type="text"
                  placeholder="Search by username, email..."
                  value={search}
                  onChange={(e) => setSearch(e.target.value)}
                  className="pl-10"
                />
              </div>
            </div>
            <div>
              <select
                value={roleFilter}
                onChange={(e) => setRoleFilter(e.target.value)}
                className="w-full p-2 border border-border rounded-md bg-background"
              >
                <option value="all">All Roles</option>
                <option value="user">User</option>
                <option value="moderator">Moderator</option>
                <option value="admin">Admin</option>
              </select>
            </div>
            <div>
              <select
                value={statusFilter}
                onChange={(e) => setStatusFilter(e.target.value)}
                className="w-full p-2 border border-border rounded-md bg-background"
              >
                <option value="all">All Status</option>
                <option value="active">Active</option>
                <option value="banned">Banned</option>
                <option value="unclaimed">Unclaimed</option>
              </select>
            </div>
          </div>
        </CardBody>
      </Card>

      {/* Stats */}
      {data && (
        <div className="grid grid-cols-1 md:grid-cols-3 gap-4 mb-6">
          <Card>
            <CardBody className="flex items-center justify-between">
              <div>
                <p className="text-sm text-muted-foreground">Total Users</p>
                <p className="text-2xl font-bold">{data.total}</p>
              </div>
            </CardBody>
          </Card>
          <Card>
            <CardBody className="flex items-center justify-between">
              <div>
                <p className="text-sm text-muted-foreground">On This Page</p>
                <p className="text-2xl font-bold">{data?.users?.length ?? 0}</p>
              </div>
            </CardBody>
          </Card>
          <Card>
            <CardBody className="flex items-center justify-between">
              <div>
                <p className="text-sm text-muted-foreground">Page</p>
                <p className="text-2xl font-bold">{page} / {totalPages}</p>
              </div>
            </CardBody>
          </Card>
        </div>
      )}

      {/* Users Table */}
      <Card>
        <CardHeader>
          <h2 className="text-xl font-semibold">Users</h2>
        </CardHeader>
        <CardBody>
          {isLoading && (
            <div className="text-center py-8">
              <p className="text-muted-foreground">Loading users...</p>
            </div>
          )}

          {error && (
            <div className="bg-destructive/10 border border-destructive text-destructive px-4 py-3 rounded">
              <p className="font-bold">Error loading users</p>
              <p>Please try again later.</p>
            </div>
          )}

          {data?.users?.length === 0 && (
            <div className="text-center py-8">
              <p className="text-muted-foreground">No users found matching your criteria.</p>
            </div>
          )}

          {data?.users?.length > 0 && (
            <div className="overflow-x-auto">
              <table className="w-full">
                <thead>
                  <tr className="border-b border-border">
                    <th className="text-left p-3">User</th>
                    <th className="text-left p-3">Email</th>
                    <th className="text-left p-3">Role</th>
                    <th className="text-left p-3">Uppies</th>
                    <th className="text-left p-3">Status</th>
                    <th className="text-left p-3">Joined</th>
                    <th className="text-right p-3">Actions</th>
                  </tr>
                </thead>
                <tbody>
                  {(() => {
                    const now = new Date();
                    return data.users.map((user) => (
                    <tr key={user.id} className="border-b border-border hover:bg-accent">
                      <td className="p-3">
                        <div className="flex items-center gap-2">
                          {user.avatar_url && (
                            <img
                              src={user.avatar_url}
                              alt={user.username}
                              className="w-8 h-8 rounded-full"
                            />
                          )}
                          <div>
                            <p className="font-medium">{user.username}</p>
                            {user.display_name && (
                              <p className="text-xs text-muted-foreground">{user.display_name}</p>
                            )}
                          </div>
                        </div>
                      </td>
                      <td className="p-3 text-sm">{user.email}</td>
                      <td className="p-3">
                        <span
                          className={`inline-flex items-center gap-1 px-2 py-1 rounded text-xs font-medium ${
                            user.role === 'admin'
                              ? 'bg-purple-500/20 text-purple-400'
                              : user.role === 'moderator'
                              ? 'bg-blue-500/20 text-blue-400'
                              : 'bg-gray-500/20 text-gray-400'
                          }`}
                        >
                          {user.role === 'admin' || user.role === 'moderator' ? (
                            <Shield className="w-3 h-3" />
                          ) : null}
                          {user.role}
                        </span>
                      </td>
                      <td className="p-3 text-sm">{user.karma_points}</td>
                      <td className="p-3">
                        <span
                          className={`inline-flex items-center gap-1 px-2 py-1 rounded text-xs font-medium ${
                            user.is_banned
                              ? 'bg-red-500/20 text-red-400'
                              : user.account_status === 'unclaimed'
                              ? 'bg-yellow-500/20 text-yellow-400'
                              : 'bg-green-500/20 text-green-400'
                          }`}
                        >
                          {user.is_banned ? (
                            <>
                              <Ban className="w-3 h-3" />
                              Banned
                            </>
                          ) : user.account_status === 'unclaimed' ? (
                            'Unclaimed'
                          ) : (
                            'Active'
                          )}
                        </span>
                        {user.comment_suspended_until && new Date(user.comment_suspended_until) > now && (
                          <span className="inline-flex items-center gap-1 px-2 py-1 rounded text-xs font-medium bg-orange-500/20 text-orange-400 ml-1">
                            <MessageSquareOff className="w-3 h-3" />
                            Comment Suspended
                          </span>
                        )}
                        {user.comments_require_review && (
                          <span className="inline-flex items-center gap-1 px-2 py-1 rounded text-xs font-medium bg-blue-500/20 text-blue-400 ml-1">
                            <Eye className="w-3 h-3" />
                            Review Required
                          </span>
                        )}
                        {user.comment_warning_count && user.comment_warning_count > 0 && (
                          <span className="text-xs text-yellow-400 ml-1">
                            ({user.comment_warning_count} warning{user.comment_warning_count > 1 ? 's' : ''})
                          </span>
                        )}
                      </td>
                      <td className="p-3 text-sm">
                        {new Date(user.created_at).toLocaleDateString()}
                      </td>
                      <td className="p-3">
                        <div className="flex gap-1 justify-end flex-wrap">
                          <Button
                            size="sm"
                            variant="ghost"
                            onClick={() => handleAction(user, 'karma')}
                            title="Adjust Uppies"
                          >
                            <TrendingUp className="w-4 h-4" />
                          </Button>
                          <Button
                            size="sm"
                            variant="ghost"
                            onClick={() =>
                              handleAction(user, user.role === 'user' ? 'promote' : 'demote')
                            }
                            title={user.role === 'user' ? 'Promote' : 'Demote'}
                          >
                            <Shield className="w-4 h-4" />
                          </Button>
                          <Button
                            size="sm"
                            variant="ghost"
                            onClick={() => handleAction(user, user.is_banned ? 'unban' : 'ban')}
                            title={user.is_banned ? 'Unban' : 'Ban'}
                          >
                            <Ban className="w-4 h-4" />
                          </Button>
                          {user.comment_suspended_until && new Date(user.comment_suspended_until) > now ? (
                            <Button
                              size="sm"
                              variant="ghost"
                              onClick={() => handleAction(user, 'lift_suspension')}
                              title="Lift Comment Suspension"
                              className="text-orange-500"
                            >
                              <MessageSquareOff className="w-4 h-4" />
                            </Button>
                          ) : (
                            <Button
                              size="sm"
                              variant="ghost"
                              onClick={() => handleAction(user, 'suspend_comments')}
                              title="Suspend Comment Privileges"
                            >
                              <MessageSquare className="w-4 h-4" />
                            </Button>
                          )}
                          <Button
                            size="sm"
                            variant="ghost"
                            onClick={() => handleAction(user, 'toggle_review')}
                            title={user.comments_require_review ? 'Disable Comment Review' : 'Require Comment Review'}
                            className={user.comments_require_review ? 'text-blue-500' : ''}
                          >
                            <Eye className="w-4 h-4" />
                          </Button>
                        </div>
                      </td>
                    </tr>
                  ));
                  })()}
                </tbody>
              </table>
            </div>
          )}

          {/* Pagination */}
          {data && totalPages > 1 && (
            <div className="flex justify-between items-center mt-4 pt-4 border-t border-border">
              <Button
                onClick={() => setPage((p) => Math.max(1, p - 1))}
                disabled={page === 1}
                variant="outline"
              >
                Previous
              </Button>
              <span className="text-sm text-muted-foreground">
                Page {page} of {totalPages}
              </span>
              <Button
                onClick={() => setPage((p) => Math.min(totalPages, p + 1))}
                disabled={page === totalPages}
                variant="outline"
              >
                Next
              </Button>
            </div>
          )}
        </CardBody>
      </Card>

      {/* Action Modal */}
      {selectedUser && actionType && (
        <UserActionModal
          user={selectedUser}
          actionType={actionType}
          onClose={() => {
            setSelectedUser(null);
            setActionType(null);
          }}
          onConfirm={handleConfirmAction}
        />
      )}
    </Container>
  );
}
