import { useState, useEffect } from 'react';
import { Helmet } from '@dr.pogodin/react-helmet';
import { useToast } from '../../context/ToastContext';

interface ModerationAction {
  id: string;
  moderator_id: string;
  moderator: string;
  action_type: string;
  target_type: string;
  target_id: string;
  reason?: string;
  metadata?: string;
  created_at: string;
}

export function ModerationLogPage() {
  const [actions, setActions] = useState<ModerationAction[]>([]);
  const [loading, setLoading] = useState(true);
  const [actionTypeFilter, setActionTypeFilter] = useState('');
  const [targetTypeFilter, setTargetTypeFilter] = useState('');
  const { showToast } = useToast();

  const loadModerationLog = async () => {
    try {
      setLoading(true);
      const params = new URLSearchParams();
      if (actionTypeFilter) params.append('action_type', actionTypeFilter);
      if (targetTypeFilter) params.append('target_type', targetTypeFilter);

      const response = await fetch(`/api/v1/admin/forum/moderation-log?${params.toString()}`, {
        credentials: 'include',
      });

      if (!response.ok) {
        throw new Error('Failed to load moderation log');
      }

      const data = await response.json();
      setActions(data.data || []);
    } catch (error) {
      showToast('Failed to load moderation log', 'error');
      console.error('Error loading moderation log:', error);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    loadModerationLog();
  }, [actionTypeFilter, targetTypeFilter, showToast]); // eslint-disable-line react-hooks/exhaustive-deps

  const getActionBadgeColor = (actionType: string) => {
    const colors: Record<string, string> = {
      lock_thread: 'bg-blue-100 dark:bg-blue-900 text-blue-800 dark:text-blue-200',
      unlock_thread: 'bg-blue-100 dark:bg-blue-900 text-blue-800 dark:text-blue-200',
      pin_thread: 'bg-green-100 dark:bg-green-900 text-green-800 dark:text-green-200',
      unpin_thread: 'bg-green-100 dark:bg-green-900 text-green-800 dark:text-green-200',
      delete_thread: 'bg-red-100 dark:bg-red-900 text-red-800 dark:text-red-200',
      ban_user: 'bg-red-100 dark:bg-red-900 text-red-800 dark:text-red-200',
      unban_user: 'bg-yellow-100 dark:bg-yellow-900 text-yellow-800 dark:text-yellow-200',
    };
    return colors[actionType] || 'bg-gray-100 dark:bg-gray-900 text-gray-800 dark:text-gray-200';
  };

  const formatActionType = (actionType: string) => {
    return actionType
      .split('_')
      .map((word) => word.charAt(0).toUpperCase() + word.slice(1))
      .join(' ');
  };

  return (
    <>
      <Helmet>
        <title>Moderation Log - Clipper Admin</title>
      </Helmet>
      <div className="page-container mx-auto py-6 px-4">
        <div className="mb-6">
          <h1 className="text-3xl font-bold mb-2">Moderation Log</h1>
          <p className="text-gray-600 dark:text-gray-400">
            View all moderation actions and audit trail
          </p>
        </div>

        <div className="mb-6 flex flex-wrap gap-4">
          <div>
            <label className="block text-sm font-medium mb-2">Action Type</label>
            <select
              className="p-2 border border-gray-300 dark:border-gray-600 rounded bg-white dark:bg-gray-700 text-gray-900 dark:text-white"
              value={actionTypeFilter}
              onChange={(e) => setActionTypeFilter(e.target.value)}
            >
              <option value="">All Actions</option>
              <option value="lock_thread">Lock Thread</option>
              <option value="unlock_thread">Unlock Thread</option>
              <option value="pin_thread">Pin Thread</option>
              <option value="unpin_thread">Unpin Thread</option>
              <option value="delete_thread">Delete Thread</option>
              <option value="ban_user">Ban User</option>
              <option value="unban_user">Unban User</option>
            </select>
          </div>

          <div>
            <label className="block text-sm font-medium mb-2">Target Type</label>
            <select
              className="p-2 border border-gray-300 dark:border-gray-600 rounded bg-white dark:bg-gray-700 text-gray-900 dark:text-white"
              value={targetTypeFilter}
              onChange={(e) => setTargetTypeFilter(e.target.value)}
            >
              <option value="">All Targets</option>
              <option value="thread">Thread</option>
              <option value="reply">Reply</option>
              <option value="user">User</option>
            </select>
          </div>
        </div>

        {loading ? (
          <div className="text-center py-8">
            <div className="inline-block animate-spin rounded-full h-8 w-8 border-b-2 border-purple-600"></div>
            <p className="mt-2 text-gray-600 dark:text-gray-400">Loading...</p>
          </div>
        ) : actions.length === 0 ? (
          <div className="bg-white dark:bg-gray-800 rounded-lg shadow p-8 text-center">
            <p className="text-gray-600 dark:text-gray-400">No moderation actions found</p>
          </div>
        ) : (
          <div className="bg-white dark:bg-gray-800 rounded-lg shadow overflow-hidden">
            <div className="overflow-x-auto">
              <table className="min-w-full divide-y divide-gray-200 dark:divide-gray-700">
                <thead className="bg-gray-50 dark:bg-gray-900">
                  <tr>
                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">
                      Date
                    </th>
                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">
                      Moderator
                    </th>
                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">
                      Action
                    </th>
                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">
                      Target
                    </th>
                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">
                      Reason
                    </th>
                  </tr>
                </thead>
                <tbody className="bg-white dark:bg-gray-800 divide-y divide-gray-200 dark:divide-gray-700">
                  {actions.map((action) => (
                    <tr key={action.id}>
                      <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-900 dark:text-gray-100">
                        {new Date(action.created_at).toLocaleString()}
                      </td>
                      <td className="px-6 py-4 whitespace-nowrap text-sm font-medium text-gray-900 dark:text-gray-100">
                        {action.moderator}
                      </td>
                      <td className="px-6 py-4 whitespace-nowrap">
                        <span className={`px-2 py-1 text-xs rounded ${getActionBadgeColor(action.action_type)}`}>
                          {formatActionType(action.action_type)}
                        </span>
                      </td>
                      <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500 dark:text-gray-400">
                        <div>
                          <div className="font-medium">{action.target_type}</div>
                          <div className="text-xs text-gray-400 dark:text-gray-500 font-mono">
                            {action.target_id.substring(0, 8)}...
                          </div>
                        </div>
                      </td>
                      <td className="px-6 py-4 text-sm text-gray-500 dark:text-gray-400">
                        {action.reason || <em className="text-gray-400">No reason provided</em>}
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          </div>
        )}
      </div>
    </>
  );
}

export default ModerationLogPage;
