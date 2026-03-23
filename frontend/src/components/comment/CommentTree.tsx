import React from 'react';
import { CommentItem } from './CommentItem';
import type { Comment } from '@/types/comment';

export interface CommentTreeProps {
  comments: Comment[];
  depth?: number;
  maxDepth?: number;
  currentUserId?: string;
  isAdmin?: boolean;
  clipId: string;
  variant?: 'expanded' | 'compact';
}

function getThreadColor(depth: number): string {
  switch (depth) {
    case 0:
      return 'border-thread-0';
    case 1:
      return 'border-thread-1';
    case 2:
      return 'border-thread-2';
    case 3:
      return 'border-thread-3';
    case 4:
      return 'border-thread-4';
    default:
      return 'border-border';
  }
}

/**
 * CommentTree - Reusable component for rendering nested comment threads
 *
 * Features:
 * - Recursive rendering of nested comments
 * - Visual indentation with depth-colored thread borders
 * - Depth tracking with configurable maxDepth
 * - Performance optimized with React.memo
 * - Supports both flat and nested comment structures
 * - Handles deleted/removed comments gracefully
 * - Expanded variant: full indentation and spacing
 * - Compact variant: reduced indentation, maxDepth capped at 2
 *
 * @param comments - Array of comments to render (can be top-level or nested)
 * @param depth - Current nesting depth (default: 0)
 * @param maxDepth - Maximum nesting depth before showing "Continue thread" (default: 10)
 * @param currentUserId - Current user's ID for edit/delete/vote checks
 * @param isAdmin - Whether current user has admin privileges
 * @param clipId - ID of the clip these comments belong to
 * @param variant - Display variant: 'expanded' (default) or 'compact'
 */
export const CommentTree: React.FC<CommentTreeProps> = React.memo(({
  comments,
  depth = 0,
  maxDepth = 10,
  currentUserId,
  isAdmin = false,
  clipId,
  variant = 'expanded',
}) => {
  if (!comments || comments.length === 0) {
    return null;
  }

  const isCompact = variant === 'compact';
  const effectiveMaxDepth = isCompact ? Math.min(maxDepth, 2) : maxDepth;
  const threadColor = getThreadColor(depth);

  const nestedClassName =
    depth > 0
      ? isCompact
        ? `ml-2 space-y-2 border-l-2 pl-2 ${threadColor}`
        : `ml-3 space-y-3 border-l-2 pl-3 ${threadColor}`
      : isCompact
        ? 'space-y-3'
        : 'space-y-6';

  return (
    <div className={nestedClassName}>
      {comments.map((comment) => (
        <CommentItem
          key={comment.id}
          comment={comment}
          clipId={clipId}
          currentUserId={currentUserId}
          isAdmin={isAdmin}
          depth={depth}
          maxDepth={effectiveMaxDepth}
          variant={variant}
        />
      ))}
    </div>
  );
});

CommentTree.displayName = 'CommentTree';
