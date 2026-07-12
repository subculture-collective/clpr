import { describe, it, expect, beforeEach, vi } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { CommentItem } from './CommentItem';
import type { Comment } from '@/types/comment';

// Mock dependencies
vi.mock('@/hooks', async () => {
  const actual = await vi.importActual('@/hooks');
  return {
    ...actual,
    useIsAuthenticated: vi.fn(() => false),
    useToast: vi.fn(() => ({
      success: vi.fn(),
      error: vi.fn(),
      info: vi.fn(),
    })),
  };
});

const createMockComment = (overrides: Partial<Comment> = {}): Comment => ({
  id: 'comment-1',
  clip_id: 'clip-1',
  user_id: 'user-1',
  username: 'testuser',
  user_avatar: 'https://example.com/avatar.png',
  user_karma: 1234,
  user_role: 'user',
  parent_comment_id: null,
  content: 'Test comment',
  vote_score: 5,
  created_at: '2024-01-01T00:00:00Z',
  updated_at: '2024-01-01T00:00:00Z',
  is_deleted: false,
  is_removed: false,
  reply_count: 0,
  depth: 0,
  child_count: 0,
  user_vote: null,
  replies: [],
  ...overrides,
});

const createQueryClient = () => {
  return new QueryClient({
    defaultOptions: {
      queries: {
        retry: false,
      },
      mutations: {
        retry: false,
      },
    },
  });
};

const renderWithClient = (ui: React.ReactElement) => {
  const queryClient = createQueryClient();
  return render(<QueryClientProvider client={queryClient}>{ui}</QueryClientProvider>);
};

describe('CommentItem - Collapse/Expand Badge', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  describe('Badge visibility', () => {
    it('should show collapse badge when comment has replies (child_count > 0)', () => {
      const comment = createMockComment({
        child_count: 3,
        replies: [
          createMockComment({ id: 'reply-1', content: 'Reply 1' }),
          createMockComment({ id: 'reply-2', content: 'Reply 2' }),
          createMockComment({ id: 'reply-3', content: 'Reply 3' }),
        ],
      });

      renderWithClient(<CommentItem comment={comment} clipId="clip-1" />);

      // Should show the badge with child count
      const badge = screen.getByTitle('Collapse thread');
      expect(badge).toBeInTheDocument();
      expect(badge).toHaveTextContent('3');
    });

    it('should not show collapse badge when comment has no replies (child_count = 0)', () => {
      const comment = createMockComment({
        child_count: 0,
        replies: [],
      });

      renderWithClient(<CommentItem comment={comment} clipId="clip-1" />);

      // Should not show collapse badge
      const badge = screen.queryByTitle('Collapse thread');
      expect(badge).not.toBeInTheDocument();
    });

    it('should not show collapse badge at max depth even with replies', () => {
      const comment = createMockComment({
        child_count: 2,
        replies: [
          createMockComment({ id: 'reply-1' }),
          createMockComment({ id: 'reply-2' }),
        ],
      });

      renderWithClient(<CommentItem comment={comment} clipId="clip-1" depth={10} maxDepth={10} />);

      // Badge should not be shown because we're at max depth
      const badge = screen.queryByTitle('Collapse thread');
      expect(badge).not.toBeInTheDocument();
    });
  });

  describe('Badge icon and text', () => {
    it('should show ▼ icon when expanded (default state)', () => {
      const comment = createMockComment({
        child_count: 2,
        replies: [
          createMockComment({ id: 'reply-1' }),
          createMockComment({ id: 'reply-2' }),
        ],
      });

      renderWithClient(<CommentItem comment={comment} clipId="clip-1" />);

      const badge = screen.getByTitle('Collapse thread');
      expect(badge).toHaveTextContent('▼');
      expect(badge).toHaveTextContent('2');
    });

    it('should show ▶ icon when collapsed', () => {
      const comment = createMockComment({
        child_count: 2,
        replies: [
          createMockComment({ id: 'reply-1' }),
          createMockComment({ id: 'reply-2' }),
        ],
      });

      renderWithClient(<CommentItem comment={comment} clipId="clip-1" />);

      // Click to collapse
      const badge = screen.getByTitle('Collapse thread');
      fireEvent.click(badge);

      // Icon should change to ▶
      const collapsedBadge = screen.getByTitle('Expand thread');
      expect(collapsedBadge).toHaveTextContent('▶');
      expect(collapsedBadge).toHaveTextContent('2');
    });

    it('should display correct reply count for single reply', () => {
      const comment = createMockComment({
        child_count: 1,
        replies: [createMockComment({ id: 'reply-1' })],
      });

      renderWithClient(<CommentItem comment={comment} clipId="clip-1" />);

      const badge = screen.getByTitle('Collapse thread');
      expect(badge).toHaveAttribute('aria-label', 'Collapse 1 reply');
    });

    it('should display correct reply count for multiple replies', () => {
      const comment = createMockComment({
        child_count: 5,
        replies: [
          createMockComment({ id: 'reply-1' }),
          createMockComment({ id: 'reply-2' }),
        ],
      });

      renderWithClient(<CommentItem comment={comment} clipId="clip-1" />);

      const badge = screen.getByTitle('Collapse thread');
      expect(badge).toHaveAttribute('aria-label', 'Collapse 5 replies');
    });
  });

  describe('Toggle behavior', () => {
    it('should toggle between collapsed and expanded states', () => {
      const comment = createMockComment({
        child_count: 2,
        replies: [
          createMockComment({ id: 'reply-1', content: 'Reply 1', child_count: 0 }),
          createMockComment({ id: 'reply-2', content: 'Reply 2', child_count: 0 }),
        ],
      });

      renderWithClient(<CommentItem comment={comment} clipId="clip-1" />);

      // Initially expanded - replies should be visible
      expect(screen.getByText('Reply 1')).toBeInTheDocument();
      expect(screen.getByText('Reply 2')).toBeInTheDocument();

      // Collapse
      const badges = screen.getAllByTitle('Collapse thread');
      fireEvent.click(badges[0]); // Click the first badge (parent comment)

      // Replies should be hidden
      expect(screen.queryByText('Reply 1')).not.toBeInTheDocument();
      expect(screen.queryByText('Reply 2')).not.toBeInTheDocument();

      // Expand again
      const expandBadge = screen.getByTitle('Expand thread');
      fireEvent.click(expandBadge);

      // Replies should be visible again
      expect(screen.getByText('Reply 1')).toBeInTheDocument();
      expect(screen.getByText('Reply 2')).toBeInTheDocument();
    });

    it('should hide nested replies when collapsed', () => {
      const comment = createMockComment({
        child_count: 3,
        replies: [
          createMockComment({
            id: 'reply-1',
            content: 'Reply 1',
            child_count: 1,
            replies: [
              createMockComment({ id: 'nested-1', content: 'Nested Reply', child_count: 0 }),
            ],
          }),
        ],
      });

      renderWithClient(<CommentItem comment={comment} clipId="clip-1" />);

      // Initially, all replies visible
      expect(screen.getByText('Reply 1')).toBeInTheDocument();
      expect(screen.getByText('Nested Reply')).toBeInTheDocument();

      // Collapse the parent comment
      const badges = screen.getAllByTitle('Collapse thread');
      fireEvent.click(badges[0]); // Click the parent comment badge

      // All replies and nested replies should be hidden
      expect(screen.queryByText('Reply 1')).not.toBeInTheDocument();
      expect(screen.queryByText('Nested Reply')).not.toBeInTheDocument();
    });
  });

  describe('Badge styling', () => {
    it('should have correct CSS classes for styling', () => {
      const comment = createMockComment({
        child_count: 2,
        replies: [createMockComment({ id: 'reply-1' })],
      });

      renderWithClient(<CommentItem comment={comment} clipId="clip-1" />);

      const badge = screen.getByTitle('Collapse thread');

      // Check for expected classes
      expect(badge).toHaveClass('text-xs');
      expect(badge).toHaveClass('text-muted-foreground');
      expect(badge).toHaveClass('hover:text-foreground');
      expect(badge).toHaveClass('hover:bg-muted');
      expect(badge).toHaveClass('cursor-pointer');
      expect(badge).toHaveClass('py-1');
      expect(badge).toHaveClass('px-2');
      expect(badge).toHaveClass('rounded');
    });
  });

  describe('Deleted/removed comments', () => {
    it('should show collapse badge for deleted comment with replies', () => {
      const comment = createMockComment({
        is_deleted: true,
        child_count: 2,
        replies: [
          createMockComment({ id: 'reply-1', content: 'Reply 1', child_count: 0 }),
          createMockComment({ id: 'reply-2', content: 'Reply 2', child_count: 0 }),
        ],
      });

      renderWithClient(<CommentItem comment={comment} clipId="clip-1" />);

      // Should show deleted message
      expect(screen.getByText('[deleted by user]')).toBeInTheDocument();

      // Should show collapse badge
      const badges = screen.getAllByTitle('Collapse thread');
      expect(badges[0]).toBeInTheDocument();
      expect(badges[0]).toHaveTextContent('▼');
    });

    it('should collapse/expand replies for deleted comment', () => {
      const comment = createMockComment({
        is_deleted: true,
        child_count: 1,
        replies: [createMockComment({ id: 'reply-1', content: 'Reply 1', child_count: 0 })],
      });

      renderWithClient(<CommentItem comment={comment} clipId="clip-1" />);

      // Initially expanded
      expect(screen.getByText('Reply 1')).toBeInTheDocument();

      // Collapse
      const badges = screen.getAllByTitle('Collapse thread');
      fireEvent.click(badges[0]);

      // Reply should be hidden
      expect(screen.queryByText('Reply 1')).not.toBeInTheDocument();
    });

    it('should show collapse badge for removed comment with replies', () => {
      const comment = createMockComment({
        is_removed: true,
        removed_reason: 'spam',
        child_count: 1,
        replies: [createMockComment({ id: 'reply-1', content: 'Reply 1', child_count: 0 })],
      });

      renderWithClient(<CommentItem comment={comment} clipId="clip-1" isAdmin={true} />);

      // Should show removed message
      expect(screen.getByText(/\[removed by moderator\]/)).toBeInTheDocument();

      // Should show collapse badge
      const badges = screen.getAllByTitle('Collapse thread');
      expect(badges[0]).toBeInTheDocument();
    });
  });

  describe('Accessibility', () => {
    it('should have proper aria-label for screen readers', () => {
      const comment = createMockComment({
        child_count: 3,
        replies: [createMockComment({ id: 'reply-1' })],
      });

      renderWithClient(<CommentItem comment={comment} clipId="clip-1" />);

      const badge = screen.getByTitle('Collapse thread');
      expect(badge).toHaveAttribute('aria-label', 'Collapse 3 replies');
    });

    it('should update aria-label when toggled', () => {
      const comment = createMockComment({
        child_count: 2,
        replies: [createMockComment({ id: 'reply-1' })],
      });

      renderWithClient(<CommentItem comment={comment} clipId="clip-1" />);

      // Initially collapsed state aria-label
      let badge = screen.getByTitle('Collapse thread');
      expect(badge).toHaveAttribute('aria-label', 'Collapse 2 replies');

      // Toggle to collapsed
      fireEvent.click(badge);

      // Check expanded state aria-label
      badge = screen.getByTitle('Expand thread');
      expect(badge).toHaveAttribute('aria-label', 'Expand 2 replies');
    });
  });

  describe('Depth-limited thread navigation', () => {
    it('should show "View N more replies" link when at max depth with replies', () => {
      const comment = createMockComment({
        child_count: 5,
        replies: [
          createMockComment({ id: 'reply-1' }),
          createMockComment({ id: 'reply-2' }),
        ],
      });

      renderWithClient(<CommentItem comment={comment} clipId="clip-1" depth={10} maxDepth={10} />);

      // Should show the link with count
      const link = screen.getByText(/View 5 more replies in thread/i);
      expect(link).toBeInTheDocument();
      expect(link).toHaveAttribute('href', '/clips/clip-1/comments/comment-1');
    });

    it('should use singular "reply" when child_count is 1', () => {
      const comment = createMockComment({
        child_count: 1,
        replies: [createMockComment({ id: 'reply-1' })],
      });

      renderWithClient(<CommentItem comment={comment} clipId="clip-1" depth={10} maxDepth={10} />);

      // Should show singular form
      const link = screen.getByText(/View 1 more reply in thread/i);
      expect(link).toBeInTheDocument();
    });

    it('should use plural "replies" when child_count is greater than 1', () => {
      const comment = createMockComment({
        child_count: 3,
        replies: [
          createMockComment({ id: 'reply-1' }),
          createMockComment({ id: 'reply-2' }),
          createMockComment({ id: 'reply-3' }),
        ],
      });

      renderWithClient(<CommentItem comment={comment} clipId="clip-1" depth={10} maxDepth={10} />);

      // Should show plural form
      const link = screen.getByText(/View 3 more replies in thread/i);
      expect(link).toBeInTheDocument();
    });

    it('should not show continue thread link when below max depth', () => {
      const comment = createMockComment({
        child_count: 2,
        replies: [
          createMockComment({ id: 'reply-1', content: 'Reply 1', child_count: 0 }),
          createMockComment({ id: 'reply-2', content: 'Reply 2', child_count: 0 }),
        ],
      });

      renderWithClient(<CommentItem comment={comment} clipId="clip-1" depth={5} maxDepth={10} />);

      // Should not show continue thread link
      expect(screen.queryByText(/View.*more.*in thread/i)).not.toBeInTheDocument();

      // Should show replies normally
      expect(screen.getByText('Reply 1')).toBeInTheDocument();
      expect(screen.getByText('Reply 2')).toBeInTheDocument();
    });

    it('should not show continue thread link when no replies', () => {
      const comment = createMockComment({
        child_count: 0,
        replies: [],
      });

      renderWithClient(<CommentItem comment={comment} clipId="clip-1" depth={10} maxDepth={10} />);

      // Should not show link when there are no replies
      expect(screen.queryByText(/View.*more.*in thread/i)).not.toBeInTheDocument();
    });

    it('should hide nested replies when at max depth', () => {
      const comment = createMockComment({
        child_count: 2,
        replies: [
          createMockComment({ id: 'reply-1', content: 'Hidden Reply 1', child_count: 0 }),
          createMockComment({ id: 'reply-2', content: 'Hidden Reply 2', child_count: 0 }),
        ],
      });

      renderWithClient(<CommentItem comment={comment} clipId="clip-1" depth={10} maxDepth={10} />);

      // Replies should not be rendered
      expect(screen.queryByText('Hidden Reply 1')).not.toBeInTheDocument();
      expect(screen.queryByText('Hidden Reply 2')).not.toBeInTheDocument();

      // Should show continue thread link instead
      expect(screen.getByText(/View 2 more replies in thread/i)).toBeInTheDocument();
    });

    it('should show continue thread link for deleted comment at max depth', () => {
      const comment = createMockComment({
        is_deleted: true,
        child_count: 3,
        replies: [
          createMockComment({ id: 'reply-1' }),
          createMockComment({ id: 'reply-2' }),
          createMockComment({ id: 'reply-3' }),
        ],
      });

      renderWithClient(<CommentItem comment={comment} clipId="clip-1" depth={10} maxDepth={10} />);

      // Should show deleted message
      expect(screen.getByText('[deleted by user]')).toBeInTheDocument();

      // Should show continue thread link with count
      expect(screen.getByText(/View 3 more replies in thread/i)).toBeInTheDocument();
    });

    it('should have correct link styling and classes', () => {
      const comment = createMockComment({
        child_count: 5,
        replies: [createMockComment({ id: 'reply-1' })],
      });

      renderWithClient(<CommentItem comment={comment} clipId="clip-1" depth={10} maxDepth={10} />);

      const link = screen.getByText(/View 5 more replies in thread/i);

      // Check for expected classes
      expect(link).toHaveClass('text-sm');
      expect(link).toHaveClass('text-brand');
      expect(link).toHaveClass('hover:text-brand-hover');
      expect(link).toHaveClass('inline-block');
    });

    it('should use correct clipId in link URL', () => {
      const comment = createMockComment({
        id: 'test-comment-456',
        child_count: 2,
        replies: [createMockComment({ id: 'reply-1' })],
      });

      renderWithClient(<CommentItem comment={comment} clipId="test-clip-789" depth={10} maxDepth={10} />);

      const link = screen.getByText(/View 2 more replies in thread/i);
      expect(link).toHaveAttribute('href', '/clips/test-clip-789/comments/test-comment-456');
    });
  });
});
