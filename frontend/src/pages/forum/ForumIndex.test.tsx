import { render, screen } from '@/test/test-utils';
import { describe, expect, it, vi } from 'vitest';
import { ForumIndex } from './ForumIndex';

vi.mock('@/components', async (importOriginal) => ({
  ...(await importOriginal<typeof import('@/components')>()),
  SEO: () => null,
}));

vi.mock('@/context/AuthContext', () => ({
  AuthProvider: ({ children }: { children: React.ReactNode }) => children,
  useAuth: () => ({ user: null }),
}));

vi.mock('@/lib/forum-api', () => ({
  forumApi: {
    listThreads: vi.fn().mockResolvedValue({ threads: [], total: 0 }),
    search: vi.fn().mockResolvedValue({ threads: [], total: 0 }),
  },
}));

describe('ForumIndex', () => {
  it('renders the signed-out empty state after an empty response', async () => {
    render(<ForumIndex />);

    expect(await screen.findByText('Join the conversation')).toBeInTheDocument();
    expect(screen.getByRole('link', { name: 'Sign in to post' })).toHaveAttribute(
      'href',
      '/login',
    );
  });

  it('exposes the topic filter as a labelled pressed-state group', async () => {
    render(<ForumIndex />);

    expect(await screen.findByRole('group', { name: 'Filter discussions by topic' }))
      .toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'All' })).toHaveAttribute(
      'aria-pressed',
      'true',
    );
  });
});
