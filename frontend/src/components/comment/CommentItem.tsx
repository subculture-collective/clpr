import React from 'react';
import ReactMarkdown from 'react-markdown';
import { cn, formatTimestamp } from '@/lib/utils';
import { Avatar } from '@/components/ui';
import { UserRoleBadge, VerifiedBadge } from '@/components/user';
import { CommentVoteButtons } from './CommentVoteButtons';
import { CommentActions } from './CommentActions';
import { CommentForm } from './CommentForm';
import { CommentTree } from './CommentTree';
import type { Comment } from '@/types/comment';

interface CommentItemProps {
  comment: Comment;
  clipId: string;
  currentUserId?: string;
  isAdmin?: boolean;
  depth?: number;
  maxDepth?: number;
  variant?: 'expanded' | 'compact';
  className?: string;
}

export const CommentItem: React.FC<CommentItemProps> = ({
  comment,
  clipId,
  currentUserId,
  isAdmin = false,
  depth = 0,
  maxDepth = 10,
  variant = 'expanded',
  className,
}) => {
  const [isCollapsed, setIsCollapsed] = React.useState(false);
  const [showReplyForm, setShowReplyForm] = React.useState(false);
  const [isEditing, setIsEditing] = React.useState(false);

  const isAuthor = currentUserId === comment.user_id;
  const hasReplies = comment.replies && comment.replies.length > 0;
  const shouldShowContinueThread = depth >= maxDepth && hasReplies;
  const isCompact = variant === 'compact';

  const handleReply = () => {
    setShowReplyForm(true);
    setIsCollapsed(false);
  };

  const handleEdit = () => {
    setIsEditing(true);
  };

  const handleCancelReply = () => {
    setShowReplyForm(false);
  };

  const handleCancelEdit = () => {
    setIsEditing(false);
  };

  const handleEditSuccess = () => {
    setIsEditing(false);
  };

  const handleReplySuccess = () => {
    setShowReplyForm(false);
  };

  // Render deleted/removed state
  if (comment.is_deleted || comment.is_removed) {
    return (
      <div className={cn('flex gap-3', isCompact && 'py-2', className)} id={`comment-${comment.id}`}>
        {/* Spacer for alignment with collapse badge */}
        <div className="flex-shrink-0 flex flex-col items-center gap-2">
          <div className="w-12" /> {/* Spacer for vote buttons */}

          {/* Collapse/Expand badge */}
          {comment.child_count > 0 && !shouldShowContinueThread && (
            <button
              onClick={() => setIsCollapsed(!isCollapsed)}
              className="text-xs text-muted-foreground hover:text-foreground hover:bg-muted transition-colors cursor-pointer py-1 px-2 rounded flex items-center gap-1"
              title={`${isCollapsed ? 'Expand' : 'Collapse'} thread`}
              aria-label={`${isCollapsed ? 'Expand' : 'Collapse'} ${comment.child_count} ${comment.child_count === 1 ? 'reply' : 'replies'}`}
            >
              <span className="select-none">{isCollapsed ? '▶' : '▼'}</span>
              <span>{comment.child_count}</span>
            </button>
          )}
        </div>

        <div className="flex-1">
          <div className="text-sm text-muted-foreground italic py-2">
            {comment.is_deleted ? '[deleted by user]' : '[removed by moderator]'}
            {comment.removed_reason && isAdmin && ` - ${comment.removed_reason}`}
          </div>
          {/* Nested replies */}
          {hasReplies && !shouldShowContinueThread && !isCollapsed && (
            <div className="mt-4">
              <CommentTree
                comments={comment.replies!}
                clipId={clipId}
                currentUserId={currentUserId}
                isAdmin={isAdmin}
                depth={depth + 1}
                maxDepth={maxDepth}
                variant={variant}
              />
            </div>
          )}

          {/* Continue thread link for max depth */}
          {shouldShowContinueThread && (
            <a
              href={`/clips/${clipId}/comments/${comment.id}`}
              className="mt-4 inline-block text-sm text-brand hover:text-brand-hover transition-colors cursor-pointer"
            >
              View {comment.child_count} more {comment.child_count === 1 ? 'reply' : 'replies'} in thread →
            </a>
          )}
        </div>
      </div>
    );
  }

  return (
    <div className={cn('flex gap-3', isCompact && 'py-2', className)} id={`comment-${comment.id}`}>
      {/* Vote buttons and collapse badge */}
      <div className="flex-shrink-0 flex flex-col items-center gap-2">
        <CommentVoteButtons
          commentId={comment.id}
          score={comment.vote_score}
          userVote={comment.user_vote}
          variant={variant}
        />

        {/* Collapse/Expand badge */}
        {comment.child_count > 0 && !shouldShowContinueThread && (
          <button
            onClick={() => setIsCollapsed(!isCollapsed)}
            className="text-xs text-muted-foreground hover:text-foreground hover:bg-muted transition-colors cursor-pointer py-1 px-2 rounded flex items-center gap-1"
            title={`${isCollapsed ? 'Expand' : 'Collapse'} thread`}
            aria-label={`${isCollapsed ? 'Expand' : 'Collapse'} ${comment.child_count} ${comment.child_count === 1 ? 'reply' : 'replies'}`}
          >
            <span className="select-none">{isCollapsed ? '▶' : '▼'}</span>
            <span>{comment.child_count}</span>
          </button>
        )}
      </div>

      {/* Comment content */}
      <div className="flex-1 min-w-0">
        {/* Header */}
        <div className="flex items-center gap-2 mb-2">
          <Avatar
            src={comment.user_avatar}
            alt={comment.username}
            size="sm"
            className="flex-shrink-0"
          />
          <span
            className={cn(
              'text-brand',
              isCompact
                ? 'font-medium text-[12px]'
                : 'font-heading text-[13px] font-semibold'
            )}
          >
            {comment.username}
          </span>

          {comment.user_verified && (
            <VerifiedBadge size="sm" />
          )}

          {comment.user_role && comment.user_role !== 'user' && (
            <UserRoleBadge role={comment.user_role} size="sm" />
          )}

          {comment.user_karma !== undefined && (
            <span className="text-xs text-muted-foreground">
              {comment.user_karma.toLocaleString()} karma
            </span>
          )}

          <span className="text-xs text-muted-foreground">•</span>

          <span
            className="text-xs text-muted-foreground"
            title={formatTimestamp(comment.created_at).title}
          >
            {formatTimestamp(comment.created_at).display}
          </span>

          {comment.edited_at && (
            <>
              <span className="text-xs text-muted-foreground">•</span>
              <span className="text-xs text-muted-foreground italic">edited</span>
            </>
          )}
        </div>

        {/* Show content only when expanded (not collapsed) */}
        {!isCollapsed && (
          <>
            {/* Content */}
            {isEditing ? (
              <CommentForm
                clipId={clipId}
                editCommentId={comment.id}
                initialContent={comment.content}
                onCancel={handleCancelEdit}
                onSuccess={handleEditSuccess}
                placeholder="Edit your comment..."
                className="mb-3"
              />
            ) : (
              <div className={cn('comment-body max-w-none mb-3', isCompact && 'line-clamp-2')}>
                <ReactMarkdown
                  components={{
                    // Open external links in new tab
                    a: ({ ...props }) => (
                      <a
                        {...props}
                        target="_blank"
                        rel="noopener noreferrer"
                        className="text-brand hover:text-brand-hover underline"
                      />
                    ),
                    // Code blocks
                    code: ({ className: codeClassName, children, ...props }) => {
                      const isInline = !codeClassName;
                      return isInline ? (
                        <code
                          className="px-1 py-0.5 rounded bg-surface-raised text-sm font-mono"
                          {...props}
                        >
                          {children}
                        </code>
                      ) : (
                        <code
                          className={cn(
                            'block p-3 rounded bg-surface-raised text-sm font-mono overflow-x-auto',
                            codeClassName
                          )}
                          {...props}
                        >
                          {children}
                        </code>
                      );
                    },
                  }}
                >
                  {comment.content}
                </ReactMarkdown>
              </div>
            )}

            {/* Actions */}
            {!isEditing && (
              <CommentActions
                commentId={comment.id}
                clipId={clipId}
                isAuthor={isAuthor}
                isAdmin={isAdmin}
                createdAt={comment.created_at}
                onReply={handleReply}
                onEdit={handleEdit}
                depth={depth}
                maxDepth={maxDepth}
                variant={variant}
                className="mb-3"
              />
            )}

            {/* Reply form */}
            {showReplyForm && (
              <CommentForm
                clipId={clipId}
                parentId={comment.id}
                parentUsername={comment.username}
                onCancel={handleCancelReply}
                onSuccess={handleReplySuccess}
                className="mb-3"
              />
            )}

            {/* Nested replies */}
            {hasReplies && !shouldShowContinueThread && (
              <div className="mt-4">
                <CommentTree
                  comments={comment.replies!}
                  clipId={clipId}
                  currentUserId={currentUserId}
                  isAdmin={isAdmin}
                  depth={depth + 1}
                  maxDepth={maxDepth}
                  variant={variant}
                />
              </div>
            )}

            {/* Continue thread link */}
            {shouldShowContinueThread && (
              <a
                href={`/clips/${clipId}/comments/${comment.id}`}
                className="mt-4 inline-block text-sm text-brand hover:text-brand-hover transition-colors cursor-pointer"
              >
                View {comment.child_count} more {comment.child_count === 1 ? 'reply' : 'replies'} in thread →
              </a>
            )}
          </>
        )}
      </div>
    </div>
  );
};
