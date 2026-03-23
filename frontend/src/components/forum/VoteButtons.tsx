import { useState } from 'react';
import { cn } from '@/lib/utils';
import type { VoteStats } from '@/types/forum';

interface VoteButtonsProps {
  replyId: string;
  initialStats: VoteStats;
  onVoteChange?: (stats: VoteStats) => void;
  disabled?: boolean;
  className?: string;
}

export function VoteButtons({
  replyId,
  initialStats,
  onVoteChange,
  disabled = false,
  className,
}: VoteButtonsProps) {
  const [localVote, setLocalVote] = useState<-1 | 0 | 1>(initialStats.user_vote);
  const [isLoading, setIsLoading] = useState(false);
  const [voteCount, setVoteCount] = useState(initialStats.net_votes);

  const handleVote = async (value: -1 | 0 | 1) => {
    // Determine new vote value (toggle if clicking same button)
    const newVote = localVote === value ? 0 : value;
    
    // Optimistic update
    const oldVote = localVote;
    const oldCount = voteCount;
    
    setLocalVote(newVote);
    // Calculate net change: remove old vote, add new vote
    const netChange = newVote - oldVote;
    setVoteCount(oldCount + netChange);

    setIsLoading(true);
    try {
      const response = await fetch(`/api/v1/forum/replies/${replyId}/vote`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        credentials: 'include',
        body: JSON.stringify({ vote_value: newVote }),
      });

      if (!response.ok) {
        throw new Error('Failed to vote');
      }

      // Fetch updated stats to ensure consistency
      const statsResponse = await fetch(`/api/v1/forum/replies/${replyId}/votes`, {
        credentials: 'include',
      });
      
      if (statsResponse.ok) {
        const { data } = await statsResponse.json();
        if (data) {
          // Validate response data
          const validatedStats: VoteStats = {
            upvotes: typeof data.upvotes === 'number' ? data.upvotes : 0,
            downvotes: typeof data.downvotes === 'number' ? data.downvotes : 0,
            net_votes: typeof data.net_votes === 'number' ? data.net_votes : 0,
            user_vote: (data.user_vote === -1 || data.user_vote === 0 || data.user_vote === 1) 
              ? data.user_vote 
              : 0,
          };
          
          setVoteCount(validatedStats.net_votes);
          setLocalVote(validatedStats.user_vote);
          
          if (onVoteChange) {
            onVoteChange(validatedStats);
          }
        }
      } else {
        // Keep optimistic update but log that stats could not be refreshed
        console.warn(
          'Failed to refresh vote stats for reply',
          replyId,
          'Status:',
          statsResponse.status
        );
      }
    } catch (error) {
      // Revert on error
      setLocalVote(oldVote);
      setVoteCount(oldCount);
      console.error('Failed to vote:', error);
    } finally {
      setIsLoading(false);
    }
  };

  const isPositive = voteCount > 0;
  const isNegative = voteCount < 0;

  return (
    <div className={cn('flex items-center gap-2', className)}>
      {/* Upvote button */}
      <button
        onClick={() => handleVote(1)}
        disabled={isLoading || disabled}
        className={cn(
          'p-1 hover:bg-surface-hover rounded transition-colors',
          'disabled:opacity-50 disabled:cursor-not-allowed',
          localVote === 1 && 'text-upvote bg-upvote/10'
        )}
        title="Upvote"
        aria-label="Upvote"
      >
        <svg
          xmlns="http://www.w3.org/2000/svg"
          className="w-4 h-4"
          viewBox="0 0 20 20"
          fill="currentColor"
        >
          <path
            fillRule="evenodd"
            d="M5.293 9.707a1 1 0 010-1.414l4-4a1 1 0 011.414 0l4 4a1 1 0 01-1.414 1.414L11 7.414V15a1 1 0 11-2 0V7.414L6.707 9.707a1 1 0 01-1.414 0z"
            clipRule="evenodd"
          />
        </svg>
      </button>

      {/* Vote count */}
      <span
        className={cn(
          'text-sm font-semibold min-w-[2.5rem] text-center',
          isPositive && 'text-upvote',
          isNegative && 'text-downvote',
          voteCount === 0 && 'text-muted-foreground'
        )}
      >
        {voteCount > 0 ? '+' : ''}
        {voteCount}
      </span>

      {/* Downvote button */}
      <button
        onClick={() => handleVote(-1)}
        disabled={isLoading || disabled}
        className={cn(
          'p-1 hover:bg-surface-hover rounded transition-colors',
          'disabled:opacity-50 disabled:cursor-not-allowed',
          localVote === -1 && 'text-downvote bg-downvote/10'
        )}
        title="Downvote"
        aria-label="Downvote"
      >
        <svg
          xmlns="http://www.w3.org/2000/svg"
          className="w-4 h-4"
          viewBox="0 0 20 20"
          fill="currentColor"
        >
          <path
            fillRule="evenodd"
            d="M14.707 10.293a1 1 0 010 1.414l-4 4a1 1 0 01-1.414 0l-4-4a1 1 0 111.414-1.414L9 12.586V5a1 1 0 012 0v7.586l2.293-2.293a1 1 0 011.414 0z"
            clipRule="evenodd"
          />
        </svg>
      </button>
    </div>
  );
}
