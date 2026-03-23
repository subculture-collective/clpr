import React from 'react';
import { cn } from '@/lib/utils';
import { useCommentVote } from '@/hooks';
import { useIsAuthenticated, useToast } from '@/hooks';

interface CommentVoteButtonsProps {
  commentId: string;
  score: number;
  userVote?: 1 | -1 | null;
  variant?: 'expanded' | 'compact';
  className?: string;
}

export const CommentVoteButtons: React.FC<CommentVoteButtonsProps> = ({
  commentId,
  score,
  userVote,
  variant = 'expanded',
  className,
}) => {
  const isAuthenticated = useIsAuthenticated();
  const { mutate: vote, isPending } = useCommentVote();
  const [isVoting, setIsVoting] = React.useState(false);
  const toast = useToast();

  const handleVote = (voteType: 1 | -1) => {
    if (!isAuthenticated) {
      toast.info('Please log in to vote');
      return;
    }

    if (isVoting) return;

    setIsVoting(true);
    vote(
      { comment_id: commentId, vote_type: voteType },
      {
        onSettled: () => {
          setTimeout(() => setIsVoting(false), 200);
        },
      }
    );
  };

  const isUpvoted = userVote === 1;
  const isDownvoted = userVote === -1;

  const isCompact = variant === 'compact';

  return (
    <div
      className={cn(
        isCompact
          ? 'flex flex-row items-center gap-1'
          : 'flex flex-col items-center gap-1',
        className
      )}
    >
      <button
        onClick={() => handleVote(1)}
        disabled={isPending || isVoting}
        className={cn(
          'rounded hover:bg-surface-hover transition-colors cursor-pointer',
          'disabled:opacity-50 disabled:cursor-not-allowed',
          isCompact ? 'p-0.5' : 'min-w-8 min-h-8 p-1.5',
          isUpvoted ? 'text-upvote' : 'text-text-tertiary'
        )}
        aria-label="Upvote"
        title="Upvote"
      >
        <svg
          xmlns="http://www.w3.org/2000/svg"
          viewBox="0 0 24 24"
          fill={isUpvoted ? 'currentColor' : 'none'}
          stroke="currentColor"
          strokeWidth="2"
          className={isCompact ? 'w-4 h-4' : 'w-5 h-5'}
        >
          <path
            strokeLinecap="round"
            strokeLinejoin="round"
            d="M4.5 15.75l7.5-7.5 7.5 7.5"
          />
        </svg>
      </button>

      <span
        aria-live="polite"
        className={cn(
          isCompact
            ? 'text-xs font-medium min-w-[1.5rem] text-center'
            : 'text-sm font-medium min-w-[2rem] text-center',
          isUpvoted && 'text-upvote',
          isDownvoted && 'text-downvote',
          !userVote && 'text-text-tertiary'
        )}
      >
        {score}
      </span>

      <button
        onClick={() => handleVote(-1)}
        disabled={isPending || isVoting}
        className={cn(
          'rounded hover:bg-surface-hover transition-colors cursor-pointer',
          'disabled:opacity-50 disabled:cursor-not-allowed',
          isCompact ? 'p-0.5' : 'min-w-8 min-h-8 p-1.5',
          isDownvoted ? 'text-downvote' : 'text-text-tertiary'
        )}
        aria-label="Downvote"
        title="Downvote"
      >
        <svg
          xmlns="http://www.w3.org/2000/svg"
          viewBox="0 0 24 24"
          fill={isDownvoted ? 'currentColor' : 'none'}
          stroke="currentColor"
          strokeWidth="2"
          className={isCompact ? 'w-4 h-4' : 'w-5 h-5'}
        >
          <path
            strokeLinecap="round"
            strokeLinejoin="round"
            d="M19.5 8.25l-7.5 7.5-7.5-7.5"
          />
        </svg>
      </button>
    </div>
  );
};
