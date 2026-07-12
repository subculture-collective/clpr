import { describe, it, expect, vi } from 'vitest';
import { render, screen, fireEvent } from '@/test/test-utils';
import { ReplyTree } from './ReplyTree';
import type { ForumReply } from '@/types/forum';

const mockReplies: ForumReply[] = [
  {
    id: 'reply1',
    user_id: 'user1',
    username: 'testuser',
    thread_id: 'thread1',
    content: 'This is a root reply',
    depth: 0,
    path: 'reply1',
    created_at: '2024-01-01T00:00:00Z',
    updated_at: '2024-01-01T00:00:00Z',
    replies: [
      {
        id: 'reply2',
        user_id: 'user2',
        username: 'anotheruser',
        thread_id: 'thread1',
        parent_reply_id: 'reply1',
        content: 'This is a nested reply',
        depth: 1,
        path: 'reply1.reply2',
        created_at: '2024-01-01T01:00:00Z',
        updated_at: '2024-01-01T01:00:00Z',
      },
    ],
  },
  {
    id: 'reply3',
    user_id: 'user3',
    username: 'thirduser',
    thread_id: 'thread1',
    content: 'This is another root reply',
    depth: 0,
    path: 'reply3',
    created_at: '2024-01-01T02:00:00Z',
    updated_at: '2024-01-01T02:00:00Z',
    is_deleted: true,
  },
];

describe('ReplyTree', () => {
  const mockOnReply = vi.fn();
  const mockOnEdit = vi.fn();
  const mockOnDelete = vi.fn();

  it('renders replies hierarchically', () => {
    render(
      <ReplyTree
        replies={mockReplies}
        threadId="thread1"
        onReply={mockOnReply}
        onEdit={mockOnEdit}
        onDelete={mockOnDelete}
      />
    );

    expect(screen.getByText('This is a root reply')).toBeInTheDocument();
    expect(screen.getByText('This is a nested reply')).toBeInTheDocument();
  });

  it('displays usernames correctly', () => {
    render(
      <ReplyTree
        replies={mockReplies}
        threadId="thread1"
        onReply={mockOnReply}
        onEdit={mockOnEdit}
        onDelete={mockOnDelete}
      />
    );

    expect(screen.getByText('testuser')).toBeInTheDocument();
    expect(screen.getByText('anotheruser')).toBeInTheDocument();
  });

  it('shows deleted reply placeholder', () => {
    render(
      <ReplyTree
        replies={mockReplies}
        threadId="thread1"
        onReply={mockOnReply}
        onEdit={mockOnEdit}
        onDelete={mockOnDelete}
      />
    );

    expect(screen.getByText('[deleted]')).toBeInTheDocument();
  });

  it('shows reply buttons for non-deleted replies', () => {
    render(
      <ReplyTree
        replies={mockReplies}
        threadId="thread1"
        currentUserId="user1"
        onReply={mockOnReply}
        onEdit={mockOnEdit}
        onDelete={mockOnDelete}
      />
    );

    const replyButtons = screen.getAllByText('Reply');
    expect(replyButtons.length).toBeGreaterThan(0);
  });

  it('shows edit and delete buttons for own replies', () => {
    render(
      <ReplyTree
        replies={mockReplies}
        threadId="thread1"
        currentUserId="user1"
        onReply={mockOnReply}
        onEdit={mockOnEdit}
        onDelete={mockOnDelete}
      />
    );

    expect(screen.getByText('Edit')).toBeInTheDocument();
    expect(screen.getByText('Delete')).toBeInTheDocument();
  });

  it('calls onReply when reply button is clicked', () => {
    render(
      <ReplyTree
        replies={mockReplies}
        threadId="thread1"
        onReply={mockOnReply}
        onEdit={mockOnEdit}
        onDelete={mockOnDelete}
      />
    );

    const replyButton = screen.getAllByText('Reply')[0];
    fireEvent.click(replyButton);

    expect(mockOnReply).toHaveBeenCalledWith('reply1');
  });

  it('applies indentation for nested replies', () => {
    const { container } = render(
      <ReplyTree
        replies={mockReplies}
        threadId="thread1"
        depth={1}
        onReply={mockOnReply}
        onEdit={mockOnEdit}
        onDelete={mockOnDelete}
      />
    );

    // Check for border-l class which indicates indentation
    const borderedElement = container.querySelector('.border-l-2');
    expect(borderedElement).toBeInTheDocument();
  });
});
