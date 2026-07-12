import { describe, it, expect } from 'vitest';
import { render, screen } from '@testing-library/react';
import { VoteButtons } from './VoteButtons';

describe('VoteButtons', () => {
  it('renders vote buttons with initial stats', () => {
    const initialStats = {
      upvotes: 5,
      downvotes: 2,
      net_votes: 3,
      user_vote: 0 as const,
    };

    render(
      <VoteButtons
        replyId="test-reply-id"
        initialStats={initialStats}
      />
    );

    // Check that vote count is displayed
    expect(screen.getByText('+3')).toBeInTheDocument();
    
    // Check that buttons are present
    expect(screen.getByLabelText('Upvote')).toBeInTheDocument();
    expect(screen.getByLabelText('Downvote')).toBeInTheDocument();
  });

  it('displays negative vote count correctly', () => {
    const initialStats = {
      upvotes: 2,
      downvotes: 5,
      net_votes: -3,
      user_vote: 0 as const,
    };

    render(
      <VoteButtons
        replyId="test-reply-id"
        initialStats={initialStats}
      />
    );

    expect(screen.getByText('-3')).toBeInTheDocument();
  });

  it('displays zero vote count correctly', () => {
    const initialStats = {
      upvotes: 0,
      downvotes: 0,
      net_votes: 0,
      user_vote: 0 as const,
    };

    render(
      <VoteButtons
        replyId="test-reply-id"
        initialStats={initialStats}
      />
    );

    expect(screen.getByText('0')).toBeInTheDocument();
  });

  it('highlights upvote button when user has upvoted', () => {
    const initialStats = {
      upvotes: 5,
      downvotes: 2,
      net_votes: 3,
      user_vote: 1 as const,
    };

    render(
      <VoteButtons
        replyId="test-reply-id"
        initialStats={initialStats}
      />
    );

    const upvoteButton = screen.getByLabelText('Upvote');
    expect(upvoteButton).toHaveClass('text-upvote', 'bg-upvote/10');
  });

  it('highlights downvote button when user has downvoted', () => {
    const initialStats = {
      upvotes: 2,
      downvotes: 5,
      net_votes: -3,
      user_vote: -1 as const,
    };

    render(
      <VoteButtons
        replyId="test-reply-id"
        initialStats={initialStats}
      />
    );

    const downvoteButton = screen.getByLabelText('Downvote');
    expect(downvoteButton).toHaveClass('text-downvote', 'bg-downvote/10');
  });

  it('disables buttons when disabled prop is true', () => {
    const initialStats = {
      upvotes: 5,
      downvotes: 2,
      net_votes: 3,
      user_vote: 0 as const,
    };

    render(
      <VoteButtons
        replyId="test-reply-id"
        initialStats={initialStats}
        disabled={true}
      />
    );

    expect(screen.getByLabelText('Upvote')).toBeDisabled();
    expect(screen.getByLabelText('Downvote')).toBeDisabled();
  });
});
