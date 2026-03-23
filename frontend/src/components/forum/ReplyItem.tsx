import { useState, useEffect } from 'react';
import { useMutation } from '@tanstack/react-query';
import { Flag } from 'lucide-react';
import ReactMarkdown from 'react-markdown';
import remarkGfm from 'remark-gfm';
import { formatTimestamp, cn } from '@/lib/utils';
import { Avatar, Modal } from '@/components/ui';
import { ConfirmDialog } from './ConfirmDialog';
import { ReputationBadge } from './ReputationBadge';
import { VoteButtons } from './VoteButtons';
import { forumApi } from '@/lib/forum-api';
import { useToast } from '@/context/ToastContext';
import type { ForumReply, VoteStats, FlagContentRequest } from '@/types/forum';

interface ReplyItemProps {
  reply: ForumReply;
  depth: number;
  currentUserId?: string;
  onReply: (replyId: string) => void;
  onEdit: (replyId: string, content: string) => void;
  onDelete: (replyId: string) => void;
  className?: string;
}

export function ReplyItem({
  reply,
  depth,
  currentUserId,
  onReply,
  onEdit,
  onDelete,
  className,
}: ReplyItemProps) {
  const { showToast } = useToast();
  const [isEditing, setIsEditing] = useState(false);
  const [editContent, setEditContent] = useState(reply.content);
  const [showDeleteConfirm, setShowDeleteConfirm] = useState(false);
  const [showReportModal, setShowReportModal] = useState(false);
  const [reportReason, setReportReason] = useState<FlagContentRequest['reason']>('spam');
  const [reportDetails, setReportDetails] = useState('');
  const [voteStats, setVoteStats] = useState<VoteStats | null>(reply.vote_stats || null);
  const [isLoadingVotes, setIsLoadingVotes] = useState(!reply.vote_stats);
  const [hasFetchedStats, setHasFetchedStats] = useState(!!reply.vote_stats);

  const flagContentMutation = useMutation({
    mutationFn: (data: FlagContentRequest) => forumApi.flagContent(data),
    onSuccess: () => {
      showToast('Reply has been flagged for review', 'success');
      setShowReportModal(false);
      setReportReason('spam');
      setReportDetails('');
    },
    onError: () => {
      showToast('Failed to flag reply', 'error');
    },
  });

  const handleReportReply = () => {
    flagContentMutation.mutate({
      target_type: 'reply',
      target_id: reply.id,
      reason: reportReason,
      details: reportDetails || undefined,
    });
  };

  const timestamp = formatTimestamp(reply.created_at);
  const isAuthor = currentUserId === reply.user_id;
  const maxDepth = 10;

  // Fetch vote stats if not provided
  useEffect(() => {
    if (!reply.vote_stats && !hasFetchedStats) {
      const fetchVoteStats = async () => {
        try {
          const response = await fetch(`/api/v1/forum/replies/${reply.id}/votes`, {
            credentials: 'include',
          });
          if (response.ok) {
            const { data } = await response.json();
            setVoteStats(data);
          }
        } catch (error) {
          console.error('Failed to fetch vote stats:', error);
          // Set default stats on error
          setVoteStats({
            upvotes: 0,
            downvotes: 0,
            net_votes: 0,
            user_vote: 0,
          });
        } finally {
          setIsLoadingVotes(false);
          setHasFetchedStats(true);
        }
      };
      fetchVoteStats();
    } else if (reply.vote_stats) {
      setIsLoadingVotes(false);
      setHasFetchedStats(true);
    }
  }, [reply.id, reply.vote_stats, hasFetchedStats]);

  // Soft-deleted reply
  if (reply.is_deleted) {
    return (
      <div
        className={cn(
          'p-3 bg-surface-raised rounded border border-border',
          'text-sm text-muted-foreground italic',
          className
        )}
      >
        [deleted]
      </div>
    );
  }

  const handleSaveEdit = () => {
    if (editContent.trim() && editContent !== reply.content) {
      onEdit(reply.id, editContent.trim());
    }
    setIsEditing(false);
  };

  const handleCancelEdit = () => {
    setEditContent(reply.content);
    setIsEditing(false);
  };

  return (
    <div
      className={cn(
        'bg-surface rounded-lg border border-border p-4',
        className
      )}
    >
      <div className="flex gap-3">
        {/* Avatar */}
        <div className="flex-shrink-0">
          <Avatar
            src={`https://api.dicebear.com/7.x/avataaars/svg?seed=${reply.username}`}
            alt={reply.username}
            size="md"
          />
        </div>

        {/* Content */}
        <div className="flex-1 min-w-0">
          {/* Header */}
          <div className="flex justify-between items-start mb-2">
            <div>
              <div className="flex items-center gap-2">
                <p className="font-semibold text-white">{reply.username}</p>
                {reply.reputation && (
                  <ReputationBadge score={reply.reputation.score} badge={reply.reputation.badge} />
                )}
              </div>
              <p className="text-xs text-muted-foreground" title={timestamp.title}>
                {timestamp.display}
              </p>
            </div>

            {/* Actions */}
            {isAuthor && !isEditing && (
              <div className="flex gap-2 text-xs">
                <button
                  onClick={() => setIsEditing(true)}
                  className="text-muted-foreground hover:text-white transition-colors"
                >
                  Edit
                </button>
                <button
                  onClick={() => setShowDeleteConfirm(true)}
                  className="text-muted-foreground hover:text-red-500 transition-colors"
                >
                  Delete
                </button>
              </div>
            )}
          </div>

          {/* Content or Editor */}
          {isEditing ? (
            <div className="space-y-2">
              <textarea
                value={editContent}
                onChange={(e) => setEditContent(e.target.value)}
                className={cn(
                  'w-full bg-surface-raised text-white rounded-lg p-3',
                  'border border-border focus:border-primary-500 focus:outline-none',
                  'resize-none'
                )}
                rows={4}
                placeholder="Edit your reply..."
              />
              <div className="flex justify-end gap-2">
                <button
                  onClick={handleCancelEdit}
                  className="px-3 py-1.5 text-sm bg-surface-raised hover:bg-surface-hover text-white rounded transition-colors"
                >
                  Cancel
                </button>
                <button
                  onClick={handleSaveEdit}
                  disabled={!editContent.trim()}
                  className="px-3 py-1.5 text-sm bg-primary-600 hover:bg-primary-700 disabled:opacity-50 disabled:cursor-not-allowed text-white rounded transition-colors"
                >
                  Save
                </button>
              </div>
            </div>
          ) : (
            <>
              <div className="forum-body max-w-none">
                <ReactMarkdown remarkPlugins={[remarkGfm]}>
                  {reply.content}
                </ReactMarkdown>
              </div>

              {/* Vote buttons and reply actions */}
              <div className="mt-3 flex items-center gap-4">
                {/* Vote buttons */}
                {!isLoadingVotes && voteStats && (
                  <VoteButtons
                    replyId={reply.id}
                    initialStats={voteStats}
                    onVoteChange={setVoteStats}
                    disabled={!currentUserId}
                  />
                )}

                {/* Reply button */}
                {depth < maxDepth && (
                  <button
                    onClick={() => onReply(reply.id)}
                    className="text-sm text-primary-400 hover:text-primary-300 transition-colors"
                  >
                    Reply
                  </button>
                )}

                {/* Report button */}
                {currentUserId && !isAuthor && (
                  <button
                    onClick={() => setShowReportModal(true)}
                    className="flex items-center gap-1 text-sm text-muted-foreground hover:text-red-400 transition-colors"
                  >
                    <Flag className="w-3.5 h-3.5" />
                    Report
                  </button>
                )}
              </div>
            </>
          )}
        </div>
      </div>

      {/* Delete confirmation dialog */}
      <ConfirmDialog
        open={showDeleteConfirm}
        onClose={() => setShowDeleteConfirm(false)}
        onConfirm={() => onDelete(reply.id)}
        title="Delete Reply"
        message="Are you sure you want to delete this reply? This action cannot be undone."
        confirmLabel="Delete"
        cancelLabel="Cancel"
        variant="danger"
      />

      {/* Report modal */}
      <Modal
        open={showReportModal}
        onClose={() => setShowReportModal(false)}
        title="Report Reply"
        size="sm"
      >
        <div className="p-6">
          <label className="block text-sm font-medium text-foreground mb-2">
            Reason
          </label>
          <select
            value={reportReason}
            onChange={(e) => setReportReason(e.target.value as FlagContentRequest['reason'])}
            className="w-full bg-surface-raised text-white rounded-lg p-2.5 border border-border focus:border-primary-500 focus:outline-none mb-4"
          >
            <option value="spam">Spam</option>
            <option value="harassment">Harassment</option>
            <option value="off-topic">Off-topic</option>
            <option value="misinformation">Misinformation</option>
            <option value="other">Other</option>
          </select>

          <label className="block text-sm font-medium text-foreground mb-2">
            Additional details (optional)
          </label>
          <textarea
            value={reportDetails}
            onChange={(e) => setReportDetails(e.target.value)}
            className="w-full bg-surface-raised text-white rounded-lg p-3 border border-border focus:border-primary-500 focus:outline-none resize-none mb-4"
            rows={3}
            placeholder="Provide any additional context..."
          />

          <div className="flex gap-3 justify-end">
            <button
              onClick={() => setShowReportModal(false)}
              className="px-4 py-2 bg-surface-raised hover:bg-surface-hover text-white rounded-lg transition-colors"
            >
              Cancel
            </button>
            <button
              onClick={handleReportReply}
              disabled={flagContentMutation.isPending}
              className="px-4 py-2 bg-red-600 hover:bg-red-700 disabled:opacity-50 disabled:cursor-not-allowed text-white rounded-lg transition-colors"
            >
              {flagContentMutation.isPending ? 'Submitting...' : 'Submit Report'}
            </button>
          </div>
        </div>
      </Modal>
    </div>
  );
}
