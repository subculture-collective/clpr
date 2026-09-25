import { render, screen } from '@/test/test-utils';
import { describe, expect, it, vi } from 'vitest';
import { forumApi } from '@/lib/forum-api';
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

  it('reads a legacy trending link as the view-ranked sort', async () => {
    window.history.pushState({}, '', '/forum?sort=trending');
    render(<ForumIndex />);

    expect(await screen.findByRole('combobox', { name: 'Sort discussions' })).toHaveValue('popular');
    expect(forumApi.listThreads).toHaveBeenCalledWith(expect.objectContaining({ sort: 'popular' }));
    window.history.pushState({}, '', '/');
  });

  it('shows the load error without the signed-out empty state', async () => {
    vi.mocked(forumApi.listThreads).mockRejectedValueOnce(new Error('Request failed with status code 400'));
    render(<ForumIndex />);

    expect(await screen.findByRole('alert')).toHaveTextContent('Failed to load threads');
    expect(screen.queryByText('Join the conversation')).not.toBeInTheDocument();
    expect(screen.queryByText(/status code/)).not.toBeInTheDocument();
  });
});
