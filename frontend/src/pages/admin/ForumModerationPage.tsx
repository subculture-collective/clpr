import { useState, useEffect } from 'react';
import { Helmet } from '@dr.pogodin/react-helmet';
import { useToast } from '../../context/ToastContext';
import { Modal } from '@/components/ui/Modal';
import { Button } from '@/components/ui/Button';

interface FlaggedContent {
  id: string;
  target_type: string;
  target_id: string;
  reason: string;
  details?: string;
  status: string;
  user_id: string;
  username: string;
  title?: string;
  content: string;
  flag_count: number;
  created_at: string;
  reviewed_by?: string;
  reviewed_at?: string;
}

interface BanUserModalProps {
  isOpen: boolean;
  onClose: () => void;
  onBan: (userId: string, reason: string, durationDays: number) => void;
  userId: string;
  username: string;
}

function BanUserModal({ isOpen, onClose, onBan, userId, username }: BanUserModalProps) {
  const [reason, setReason] = useState('');
  const [durationDays, setDurationDays] = useState(0);

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    if (!reason.trim()) {
      return;
    }
    onBan(userId, reason, durationDays);
    setReason('');
    setDurationDays(0);
    onClose();
  };

  return (
    <Modal open={isOpen} onClose={onClose} title={`Ban User: ${username}`} size="md">
        <form onSubmit={handleSubmit}>
          <div className="mb-4">
            <label htmlFor="forum-ban-reason" className="block text-sm font-medium mb-2">Reason</label>
            <textarea
              id="forum-ban-reason"
              className="w-full p-2 border border-gray-300 dark:border-gray-600 rounded bg-white dark:bg-gray-700 text-gray-900 dark:text-white"
              rows={4}
              value={reason}
              onChange={(e) => setReason(e.target.value)}
              placeholder="Explain the reason for this ban..."
              required
            />
          </div>
          <div className="mb-4">
            <label htmlFor="forum-ban-duration" className="block text-sm font-medium mb-2">Duration</label>
            <select
              id="forum-ban-duration"
              className="w-full p-2 border border-gray-300 dark:border-gray-600 rounded bg-white dark:bg-gray-700 text-gray-900 dark:text-white"
              value={durationDays}
              onChange={(e) => setDurationDays(Number(e.target.value))}
            >
              <option value={0}>Permanent</option>
              <option value={1}>1 Day</option>
              <option value={7}>7 Days</option>
              <option value={30}>30 Days</option>
              <option value={90}>90 Days</option>
            </select>
          </div>
          <div className="flex gap-2">
            <Button
              type="submit"
              variant="danger"
              className="flex-1"
            >
              Ban User
            </Button>
            <Button
              type="button"
              onClick={onClose}
              variant="outline"
              className="flex-1"
            >
              Cancel
            </Button>
          </div>
        </form>
    </Modal>
  );
}

interface FlaggedContentCardProps {
  item: FlaggedContent;
  onLock: (threadId: string, reason: string, locked: boolean) => void;
  onPin: (threadId: string, reason: string, pinned: boolean) => void;
  onDelete: (threadId: string, reason: string) => void;
  onBanUser: (userId: string, reason: string, durationDays: number) => void;
}

function FlaggedContentCard({ item, onLock, onPin, onDelete, onBanUser }: FlaggedContentCardProps) {
  const [showBanModal, setShowBanModal] = useState(false);
  const [actionReason, setActionReason] = useState('');
  const [showActionInput, setShowActionInput] = useState<string | null>(null);

  const handleAction = (action: string) => {
    if (!actionReason.trim()) {
      return;
    }

    switch (action) {
      case 'lock':
        onLock(item.target_id, actionReason, true);
        break;
      case 'pin':
        onPin(item.target_id, actionReason, true);
        break;
      case 'delete':
        onDelete(item.target_id, actionReason);
        break;
    }

    setActionReason('');
    setShowActionInput(null);
  };

  const getActionButtonText = (action: string): string => {
    const actionLabels: Record<string, string> = {
      lock: 'Lock Thread',
      pin: 'Pin Thread',
      delete: 'Delete Thread',
    };
    return actionLabels[action] || 'Confirm';
  };

  return (
    <div className="bg-white dark:bg-gray-800 rounded-lg shadow p-4 mb-4">
      <div className="flex justify-between items-start mb-2">
        <div className="flex-1">
          <div className="flex items-center gap-2 mb-2">
            <span className="px-2 py-1 bg-red-100 dark:bg-red-900 text-red-800 dark:text-red-200 text-xs rounded">
              {item.target_type}
            </span>
            <span className="px-2 py-1 bg-yellow-100 dark:bg-yellow-900 text-yellow-800 dark:text-yellow-200 text-xs rounded">
              {item.reason}
            </span>
            {item.flag_count > 1 && (
              <span className="px-2 py-1 bg-orange-100 dark:bg-orange-900 text-orange-800 dark:text-orange-200 text-xs rounded">
                {item.flag_count} flags
              </span>
            )}
          </div>
          {item.title && (
            <h3 className="text-lg font-semibold mb-1">{item.title}</h3>
          )}
          <p className="text-gray-700 dark:text-gray-300 mb-2">{item.content}</p>
          <div className="text-sm text-gray-500 dark:text-gray-400">
            Reported by: <span className="font-medium">{item.username}</span> • {new Date(item.created_at).toLocaleString()}
          </div>
        </div>
      </div>

      <div className="mt-4 border-t pt-4 dark:border-gray-700">
        <div className="flex flex-wrap gap-2">
          {item.target_type === 'thread' && (
            <>
              <button
                onClick={() => setShowActionInput(showActionInput === 'lock' ? null : 'lock')}
                className="px-3 py-1 bg-blue-600 hover:bg-blue-700 text-white text-sm rounded"
              >
                Lock Thread
              </button>
              <button
                onClick={() => setShowActionInput(showActionInput === 'pin' ? null : 'pin')}
                className="px-3 py-1 bg-green-600 hover:bg-green-700 text-white text-sm rounded"
              >
                Pin Thread
              </button>
              <button
                onClick={() => setShowActionInput(showActionInput === 'delete' ? null : 'delete')}
                className="px-3 py-1 bg-red-600 hover:bg-red-700 text-white text-sm rounded"
              >
                Delete Thread
              </button>
            </>
          )}
          <button
            onClick={() => setShowBanModal(true)}
            className="px-3 py-1 bg-red-800 hover:bg-red-900 text-white text-sm rounded"
          >
            Ban User
          </button>
        </div>

        {showActionInput && (
          <div className="mt-3">
            <input
              type="text"
              className="w-full p-2 border border-gray-300 dark:border-gray-600 rounded bg-white dark:bg-gray-700 text-gray-900 dark:text-white"
              placeholder="Enter reason for this action..."
              value={actionReason}
              onChange={(e) => setActionReason(e.target.value)}
            />
            <button
              onClick={() => handleAction(showActionInput)}
              disabled={!actionReason.trim()}
              className="mt-2 px-4 py-2 bg-purple-600 hover:bg-purple-700 disabled:bg-gray-400 text-white rounded"
            >
              Confirm {getActionButtonText(showActionInput)}
            </button>
          </div>
        )}
      </div>

      <BanUserModal
        isOpen={showBanModal}
        onClose={() => setShowBanModal(false)}
        onBan={onBanUser}
        userId={item.user_id}
        username={item.username}
      />
    </div>
  );
}

export function ForumModerationPage() {
  const [flaggedContent, setFlaggedContent] = useState<FlaggedContent[]>([]);
  const [loading, setLoading] = useState(true);
  const [status, setStatus] = useState('pending');
  const { showToast } = useToast();

  const loadFlaggedContent = async () => {
    try {
      setLoading(true);
      const response = await fetch(`/api/v1/admin/forum/flagged?status=${status}`, {
        credentials: 'include',
      });

      if (!response.ok) {
        throw new Error('Failed to load flagged content');
      }

      const data = await response.json();
      setFlaggedContent(data.data || []);
    } catch (error) {
      showToast('Failed to load flagged content', 'error');
      console.error('Error loading flagged content:', error);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    loadFlaggedContent();
  }, [status, showToast]); // eslint-disable-line react-hooks/exhaustive-deps

  const lockThread = async (threadId: string, reason: string, locked: boolean) => {
    try {
      const response = await fetch(`/api/v1/admin/forum/threads/${threadId}/lock`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        credentials: 'include',
        body: JSON.stringify({ reason, locked }),
      });

      if (!response.ok) {
        throw new Error('Failed to lock thread');
      }

      showToast(`Thread ${locked ? 'locked' : 'unlocked'} successfully`, 'success');
      loadFlaggedContent();
    } catch (error) {
      showToast('Failed to lock thread', 'error');
      console.error('Error locking thread:', error);
    }
  };

  const pinThread = async (threadId: string, reason: string, pinned: boolean) => {
    try {
      const response = await fetch(`/api/v1/admin/forum/threads/${threadId}/pin`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        credentials: 'include',
        body: JSON.stringify({ reason, pinned }),
      });

      if (!response.ok) {
        throw new Error('Failed to pin thread');
      }

      showToast(`Thread ${pinned ? 'pinned' : 'unpinned'} successfully`, 'success');
      loadFlaggedContent();
    } catch (error) {
      showToast('Failed to pin thread', 'error');
      console.error('Error pinning thread:', error);
    }
  };

  const deleteThread = async (threadId: string, reason: string) => {
    try {
      const response = await fetch(`/api/v1/admin/forum/threads/${threadId}/delete`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        credentials: 'include',
        body: JSON.stringify({ reason }),
      });

      if (!response.ok) {
        throw new Error('Failed to delete thread');
      }

      showToast('Thread deleted successfully', 'success');
      loadFlaggedContent();
    } catch (error) {
      showToast('Failed to delete thread', 'error');
      console.error('Error deleting thread:', error);
    }
  };

  const banUser = async (userId: string, reason: string, durationDays: number) => {
    try {
      const response = await fetch(`/api/v1/admin/forum/users/${userId}/ban`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        credentials: 'include',
        body: JSON.stringify({ reason, duration_days: durationDays }),
      });

      if (!response.ok) {
        throw new Error('Failed to ban user');
      }

      showToast('User banned successfully', 'success');
      loadFlaggedContent();
    } catch (error) {
      showToast('Failed to ban user', 'error');
      console.error('Error banning user:', error);
    }
  };

  return (
    <>
      <Helmet>
        <title>Forum Moderation - Clipper Admin</title>
      </Helmet>
      <div className="container mx-auto py-6 px-4">
        <div className="mb-6">
          <h1 className="text-3xl font-bold mb-2">Forum Moderation</h1>
          <p className="text-gray-600 dark:text-gray-400">
            Manage flagged content and moderate forum threads
          </p>
        </div>

        <div className="mb-6">
          <div className="flex gap-2">
            <button
              onClick={() => setStatus('pending')}
              className={`px-4 py-2 rounded ${
                status === 'pending'
                  ? 'bg-purple-600 text-white'
                  : 'bg-gray-200 dark:bg-gray-700 text-gray-700 dark:text-gray-300'
              }`}
            >
              Pending
            </button>
            <button
              onClick={() => setStatus('reviewed')}
              className={`px-4 py-2 rounded ${
                status === 'reviewed'
                  ? 'bg-purple-600 text-white'
                  : 'bg-gray-200 dark:bg-gray-700 text-gray-700 dark:text-gray-300'
              }`}
            >
              Reviewed
            </button>
            <button
              onClick={() => setStatus('resolved')}
              className={`px-4 py-2 rounded ${
                status === 'resolved'
                  ? 'bg-purple-600 text-white'
                  : 'bg-gray-200 dark:bg-gray-700 text-gray-700 dark:text-gray-300'
              }`}
            >
              Resolved
            </button>
          </div>
        </div>

        {loading ? (
          <div className="text-center py-8">
            <div className="inline-block animate-spin rounded-full h-8 w-8 border-b-2 border-purple-600"></div>
            <p className="mt-2 text-gray-600 dark:text-gray-400">Loading...</p>
          </div>
        ) : flaggedContent.length === 0 ? (
          <div className="bg-white dark:bg-gray-800 rounded-lg shadow p-8 text-center">
            <p className="text-gray-600 dark:text-gray-400">
              No {status} content to display
            </p>
          </div>
        ) : (
          <div className="space-y-4">
            {flaggedContent.map((item) => (
              <FlaggedContentCard
                key={item.id}
                item={item}
                onLock={lockThread}
                onPin={pinThread}
                onDelete={deleteThread}
                onBanUser={banUser}
              />
            ))}
          </div>
        )}
      </div>
    </>
  );
}

export default ForumModerationPage;
