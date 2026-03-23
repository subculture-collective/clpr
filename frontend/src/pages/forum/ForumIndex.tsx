import { useState, useEffect, useMemo } from 'react';
import { useQuery } from '@tanstack/react-query';
import { Link, useSearchParams, useNavigate } from 'react-router-dom';
import { Plus } from 'lucide-react';
import { Container, SEO } from '@/components';
import {
  ThreadList,
  ForumSearch,
  SortSelector,
} from '@/components/forum';
import { forumApi } from '@/lib/forum-api';
import { useAuth } from '@/context/AuthContext';
import { FORUM_TOPICS } from './CreateThread';
import { cn } from '@/lib/utils';
import type { ForumSort, ForumFilters as ForumFiltersType } from '@/types/forum';

export function ForumIndex() {
  const { user } = useAuth();
  const [searchParams, setSearchParams] = useSearchParams();
  const navigate = useNavigate();

  // Get filters from URL params
  const sortParam = (searchParams.get('sort') as ForumSort) || 'newest';
  const gameIdParam = searchParams.get('game_id') || undefined;
  const tagsParam = searchParams.get('tags');
  const searchQuery = searchParams.get('q') || undefined;

  const [sort, setSort] = useState<ForumSort>(sortParam);
  const [filters, setFilters] = useState<ForumFiltersType>({
    game_id: gameIdParam,
    tags: tagsParam ? tagsParam.split(',') : [],
  });

  // Fetch threads
  const { data, isLoading, error } = useQuery({
    queryKey: ['forum-threads', sort, filters, searchQuery],
    queryFn: async () => {
      if (searchQuery) {
        return forumApi.search({ q: searchQuery, page: 1, limit: 20 });
      }
      return forumApi.listThreads({
        sort,
        game_id: filters.game_id,
        tags: filters.tags,
        page: 1,
        limit: 20,
      });
    },
  });

  // Update URL params when filters change
  useEffect(() => {
    const params = new URLSearchParams();
    if (sort !== 'newest') params.set('sort', sort);
    if (filters.game_id) params.set('game_id', filters.game_id);
    if (filters.tags && filters.tags.length > 0) {
      params.set('tags', filters.tags.join(','));
    }
    if (searchQuery) params.set('q', searchQuery);
    setSearchParams(params);
  }, [sort, filters, searchQuery, setSearchParams]);

  const handleSearch = (query: string) => {
    if (query) {
      // Use navigate to stay within React Router
      navigate(`/forum/search?q=${encodeURIComponent(query)}`);
    }
  };

  const handleSortChange = (newSort: ForumSort) => {
    setSort(newSort);
  };

  const threads = data?.threads || [];

  const sortedThreads = useMemo(() => {
    if (!threads) return [];
    return [...threads].sort((a, b) => {
      if (a.pinned && !b.pinned) return -1;
      if (!a.pinned && b.pinned) return 1;
      return 0;
    });
  }, [threads]);

  return (
    <>
      <SEO
        title="Forum - Community Discussions"
        description="Join the community discussion. Share your thoughts, ask questions, and connect with other members."
      />
      <Container className="py-6">
        <div className="max-w-4xl mx-auto">
          {/* Header */}
          <div className="flex flex-col sm:flex-row sm:justify-between sm:items-center gap-4 mb-6">
            <div className="flex-1">
              <h1 className="text-3xl font-bold text-foreground">Forum Discussions</h1>
              <Link
                to="/forum/analytics"
                className="text-sm text-primary-500 hover:text-primary-600 transition-colors mt-1 inline-block"
              >
                View Analytics →
              </Link>
            </div>
            {user && (
              <Link
                to="/forum/new"
                className="inline-flex items-center justify-center gap-2 px-4 py-2 bg-primary-600 hover:bg-primary-700 text-white font-medium rounded-lg transition-colors"
              >
                <Plus className="w-5 h-5" />
                <span>Start Discussion</span>
              </Link>
            )}
          </div>

          {/* Search */}
          <ForumSearch onSearch={handleSearch} className="mb-4" />

          {/* Topic filter buttons */}
          <div className="flex flex-wrap gap-2 mb-4">
            <button
              onClick={() => setFilters({ ...filters, tags: [] })}
              className={cn(
                'px-3 py-1.5 rounded-full text-sm font-medium border transition-colors cursor-pointer',
                (!filters.tags || filters.tags.length === 0)
                  ? 'bg-brand text-white border-transparent'
                  : 'text-text-secondary border-border hover:border-text-tertiary hover:text-text-primary',
              )}
            >
              All
            </button>
            {FORUM_TOPICS.map((t) => {
              const isActive = filters.tags?.includes(t.value);
              return (
                <button
                  key={t.value}
                  onClick={() => setFilters({
                    ...filters,
                    tags: isActive ? [] : [t.value],
                  })}
                  className={cn(
                    'px-3 py-1.5 rounded-full text-sm font-medium border transition-colors cursor-pointer',
                    isActive
                      ? 'text-white border-transparent'
                      : 'text-text-secondary border-border hover:border-text-tertiary hover:text-text-primary',
                  )}
                  style={isActive ? { backgroundColor: t.color } : undefined}
                >
                  {t.label}
                </button>
              );
            })}
          </div>

          {/* Sort */}
          <div className="flex justify-end mb-6">
            <SortSelector value={sort} onChange={handleSortChange} />
          </div>

          {/* Error State */}
          {error && (
            <div className="p-4 bg-error-900 border border-error-800 rounded-lg mb-6">
              <p className="text-error-400">
                Failed to load threads. Please try again later.
              </p>
            </div>
          )}

          {/* Thread List */}
          <ThreadList threads={sortedThreads} loading={isLoading} />

          {/* Empty state for unauthenticated users */}
          {!user && threads.length === 0 && !isLoading && (
            <div className="text-center py-12 bg-card rounded-lg border border-border mt-6">
              <p className="text-muted-foreground text-lg mb-4">
                Join the conversation
              </p>
              <Link
                to="/login"
                className="inline-block px-6 py-2 bg-primary-600 hover:bg-primary-700 text-white font-medium rounded-lg transition-colors"
              >
                Sign in to post
              </Link>
            </div>
          )}
        </div>
      </Container>
    </>
  );
}
