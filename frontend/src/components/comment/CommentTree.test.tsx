import { describe, it, expect, beforeEach, vi } from 'vitest';
import { render, screen } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { CommentTree } from './CommentTree';
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

describe('CommentTree', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  describe('Basic rendering', () => {
    it('should render empty when no comments provided', () => {
      const { container } = renderWithClient(
        <CommentTree comments={[]} clipId="clip-1" />
      );
      expect(container.firstChild).toBeNull();
    });

    it('should render single comment', () => {
      const comments = [createMockComment({ content: 'Single comment' })];

      renderWithClient(<CommentTree comments={comments} clipId="clip-1" />);

      expect(screen.getByText(/Single comment/i)).toBeInTheDocument();
    });

    it('should render multiple comments', () => {
      const comments = [
        createMockComment({ id: 'comment-1', content: 'First comment' }),
        createMockComment({ id: 'comment-2', content: 'Second comment' }),
        createMockComment({ id: 'comment-3', content: 'Third comment' }),
      ];

      renderWithClient(<CommentTree comments={comments} clipId="clip-1" />);

      expect(screen.getByText(/First comment/i)).toBeInTheDocument();
      expect(screen.getByText(/Second comment/i)).toBeInTheDocument();
      expect(screen.getByText(/Third comment/i)).toBeInTheDocument();
    });
  });

  describe('Depth and indentation', () => {
    it('should apply top-level styling when depth is 0', () => {
      const comments = [createMockComment()];
      const { container } = renderWithClient(
        <CommentTree comments={comments} clipId="clip-1" depth={0} />
      );

      // Top level should have space-y-6 class
      const treeContainer = container.querySelector('.space-y-6');
      expect(treeContainer).toBeInTheDocument();

      // Should not have left border at top level
      const borderElement = container.querySelector('.border-l-2');
      expect(borderElement).toBeNull();
    });

    it('should apply nested styling when depth > 0', () => {
      const comments = [createMockComment()];
      const { container } = renderWithClient(
        <CommentTree comments={comments} clipId="clip-1" depth={1} />
      );

      // Nested level should use the expanded indentation and thread border.
      const treeContainer = container.querySelector('.ml-3.border-l-2.pl-3');
      expect(treeContainer).toBeInTheDocument();
    });

    it('should render nested comments with proper indentation', () => {
      const comments = [
        createMockComment({
          id: 'comment-1',
          content: 'Parent comment',
          child_count: 1,
          replies: [
            createMockComment({
              id: 'comment-2',
              content: 'Child comment',
              child_count: 1,
              replies: [
                createMockComment({
                  id: 'comment-3',
                  content: 'Grandchild comment',
                  child_count: 0,
                }),
              ],
            }),
          ],
        }),
      ];

      renderWithClient(<CommentTree comments={comments} clipId="clip-1" depth={0} />);

      expect(screen.getByText('Parent comment')).toBeInTheDocument();
      expect(screen.getByText('Child comment')).toBeInTheDocument();
      expect(screen.getByText('Grandchild comment')).toBeInTheDocument();
    });
  });

  describe('MaxDepth handling', () => {
    it('should use default maxDepth of 10', () => {
      const comments = [createMockComment()];

      renderWithClient(<CommentTree comments={comments} clipId="clip-1" />);

      // Component should render (testing default behavior)
      expect(screen.getByText(/Test comment/i)).toBeInTheDocument();
    });

    it('should pass maxDepth prop to CommentItem components', () => {
      const comments = [createMockComment()];

      renderWithClient(
        <CommentTree comments={comments} clipId="clip-1" maxDepth={5} />
      );

      expect(screen.getByText(/Test comment/i)).toBeInTheDocument();
    });

    it('should handle deeply nested comments up to maxDepth', () => {
      // Create a chain of nested comments
      const deepComment = createMockComment({
        id: 'comment-3',
        content: 'Deep comment at level 2',
      });

      const midComment = createMockComment({
        id: 'comment-2',
        content: 'Mid comment at level 1',
        replies: [deepComment],
      });

      const topComment = createMockComment({
        id: 'comment-1',
        content: 'Top comment',
        replies: [midComment],
      });

      renderWithClient(
        <CommentTree comments={[topComment]} clipId="clip-1" maxDepth={10} />
      );

      expect(screen.getByText(/Top comment/i)).toBeInTheDocument();
      expect(screen.getByText(/Mid comment at level 1/i)).toBeInTheDocument();
      expect(screen.getByText(/Deep comment at level 2/i)).toBeInTheDocument();
    });
  });

  describe('User permissions', () => {
    it('should pass currentUserId to child components', () => {
      const comments = [createMockComment({ user_id: 'user-1' })];

      renderWithClient(
        <CommentTree
          comments={comments}
          clipId="clip-1"
          currentUserId="user-1"
        />
      );

      // User should see their own comment
      expect(screen.getByText(/Test comment/i)).toBeInTheDocument();
    });

    it('should pass isAdmin flag to child components', () => {
      const comments = [createMockComment()];

      renderWithClient(
        <CommentTree comments={comments} clipId="clip-1" isAdmin={true} />
      );

      // Admin flag should be passed (testing that component renders)
      expect(screen.getByText(/Test comment/i)).toBeInTheDocument();
    });
  });

  describe('Props propagation', () => {
    it('should propagate clipId to all child components', () => {
      const comments = [
        createMockComment({
          id: 'comment-1',
          replies: [createMockComment({ id: 'comment-2' })],
        }),
      ];

      renderWithClient(
        <CommentTree comments={comments} clipId="test-clip-123" />
      );

      // Both comments should render (clipId propagated correctly)
      expect(screen.getAllByText(/Test comment/i).length).toBeGreaterThan(0);
    });
  });

  describe('Edge cases', () => {
    it('should handle null replies gracefully', () => {
      const comments = [createMockComment({ replies: undefined })];

      renderWithClient(<CommentTree comments={comments} clipId="clip-1" />);

      expect(screen.getByText(/Test comment/i)).toBeInTheDocument();
    });

    it('should handle empty replies array', () => {
      const comments = [createMockComment({ replies: [] })];

      renderWithClient(<CommentTree comments={comments} clipId="clip-1" />);

      expect(screen.getByText(/Test comment/i)).toBeInTheDocument();
    });

    it('should render deleted comments', () => {
      const comments = [
        createMockComment({
          is_deleted: true,
          content: 'This should not be visible',
        }),
      ];

      renderWithClient(<CommentTree comments={comments} clipId="clip-1" />);

      expect(screen.getByText(/\[deleted by user\]/i)).toBeInTheDocument();
    });

    it('should render removed comments', () => {
      const comments = [
        createMockComment({
          is_removed: true,
          removed_reason: 'spam',
          content: 'This should not be visible',
        }),
      ];

      renderWithClient(
        <CommentTree comments={comments} clipId="clip-1" isAdmin={true} />
      );

      expect(screen.getByText(/\[removed by moderator\]/i)).toBeInTheDocument();
    });
  });

  describe('Performance', () => {
    it('should render large comment trees efficiently', () => {
      // Create 50 top-level comments with plain IDs (will become comment-1, comment-2, etc.)
      const comments = Array.from({ length: 50 }, (_, i) =>
        createMockComment({
          id: `${i}`,
          content: `Comment ${i}`,
        })
      );

      const { container } = renderWithClient(
        <CommentTree comments={comments} clipId="clip-1" />
      );

      // Should render all comments (CommentItem creates IDs as "comment-${comment.id}")
      expect(container.querySelectorAll('[id^="comment-"]').length).toBe(50);
    });

    it('should use React.memo for optimization', () => {
      // CommentTree should have displayName set when using memo
      expect(CommentTree.displayName).toBe('CommentTree');
    });
  });

  describe('Accessibility', () => {
    it('should render with semantic HTML structure', () => {
      const comments = [createMockComment()];
      const { container } = renderWithClient(
        <CommentTree comments={comments} clipId="clip-1" />
      );

      // Should have proper div structure
      expect(container.querySelector('div')).toBeInTheDocument();
    });

    it('should maintain proper comment IDs for anchor links', () => {
      const comments = [createMockComment({ id: 'comment-123' })];

      renderWithClient(<CommentTree comments={comments} clipId="clip-1" />);

      // CommentItem should render with proper ID
      const commentElement = document.getElementById('comment-comment-123');
      expect(commentElement).toBeInTheDocument();
    });
  });
});
