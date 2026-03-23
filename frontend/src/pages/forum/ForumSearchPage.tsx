import { useState } from 'react';
import { useSearchParams, Link } from 'react-router-dom';
import { useQuery } from '@tanstack/react-query';
import { ArrowLeft, Search as SearchIcon } from 'lucide-react';
import { Container, SEO } from '@/components';
import { ForumSearch } from '@/components/forum/ForumSearch';
import { SearchResultCard } from '@/components/forum/SearchResultCard';
import { forumApi } from '@/lib/forum-api';

export function ForumSearchPage() {
  const [searchParams, setSearchParams] = useSearchParams();
  
  const query = searchParams.get('q') || '';
  const author = searchParams.get('author') || '';
  const sort = (searchParams.get('sort') as 'relevance' | 'date' | 'votes') || 'relevance';
  const page = parseInt(searchParams.get('page') || '1', 10);

  const [localAuthor, setLocalAuthor] = useState(author);
  const [localSort, setLocalSort] = useState(sort);

  // Fetch search results
  const { data, isLoading, error } = useQuery({
    queryKey: ['forum-search', query, author, sort, page],
    queryFn: () => forumApi.search({ 
      q: query, 
      author: author || undefined,
      sort,
      page,
    }),
    enabled: query.length > 0,
  });

  const handleSearch = (newQuery: string) => {
    const params = new URLSearchParams();
    if (newQuery) params.set('q', newQuery);
    if (localAuthor) params.set('author', localAuthor);
    if (localSort !== 'relevance') params.set('sort', localSort);
    setSearchParams(params);
  };

  const handleFilterChange = () => {
    const params = new URLSearchParams();
    if (query) params.set('q', query);
    if (localAuthor) params.set('author', localAuthor);
    if (localSort !== 'relevance') params.set('sort', localSort);
    setSearchParams(params);
  };

  const results = data?.data || [];
  const meta = data?.meta;

  return (
    <>
      <SEO
        title={query ? `Search: ${query} - Forum` : 'Search Forum'}
        description="Search forum discussions and replies"
      />
      <Container className="py-6">
        <div className="max-w-4xl mx-auto">
          {/* Back Link */}
          <Link
            to="/forum"
            className="inline-flex items-center gap-2 text-muted-foreground hover:text-white mb-6 transition-colors"
          >
            <ArrowLeft className="w-4 h-4" />
            <span>Back to Forum</span>
          </Link>

          {/* Header */}
          <h1 className="text-3xl font-bold text-white mb-6">Search Forum</h1>

          {/* Search Bar */}
          <ForumSearch 
            onSearch={handleSearch}
            placeholder="Search threads and replies..."
            className="mb-6"
          />

          {/* Filters */}
          <div className="bg-background border border-border rounded-lg p-4 mb-6">
            <h2 className="text-sm font-semibold text-muted-foreground uppercase tracking-wider mb-3">
              Filters
            </h2>
            <div className="flex flex-col sm:flex-row gap-4">
              {/* Author Filter */}
              <div className="flex-1">
                <label htmlFor="author-filter" className="block text-sm text-muted-foreground mb-2">
                  Author
                </label>
                <input
                  id="author-filter"
                  type="text"
                  value={localAuthor}
                  onChange={(e) => setLocalAuthor(e.target.value)}
                  placeholder="Filter by username..."
                  className="w-full px-3 py-2 bg-surface border border-border rounded text-white placeholder-muted-foreground focus:outline-none focus:ring-2 focus:ring-primary-500 focus:border-transparent"
                />
              </div>

              {/* Sort Dropdown */}
              <div className="flex-1">
                <label htmlFor="sort-select" className="block text-sm text-muted-foreground mb-2">
                  Sort By
                </label>
                <select
                  id="sort-select"
                  value={localSort}
                  onChange={(e) => setLocalSort(e.target.value as 'relevance' | 'date' | 'votes')}
                  className="w-full px-3 py-2 bg-surface border border-border rounded text-white focus:outline-none focus:ring-2 focus:ring-primary-500 focus:border-transparent"
                >
                  <option value="relevance">Relevance</option>
                  <option value="date">Most Recent</option>
                  <option value="votes">Most Votes</option>
                </select>
              </div>

              {/* Apply Button */}
              <div className="flex items-end">
                <button
                  onClick={handleFilterChange}
                  className="px-6 py-2 bg-primary-600 hover:bg-primary-700 text-white font-medium rounded transition-colors"
                >
                  Apply
                </button>
              </div>
            </div>
          </div>

          {/* Results Info */}
          {query && meta && (
            <div className="text-sm text-muted-foreground mb-4">
              Found {meta.count} result{meta.count !== 1 ? 's' : ''} for "{query}"
              {author && ` by ${author}`}
            </div>
          )}

          {/* Error State */}
          {error && (
            <div className="p-4 bg-red-900/20 border border-red-800 rounded-lg mb-6">
              <p className="text-red-400">
                Failed to search. Please try again later.
              </p>
            </div>
          )}

          {/* Loading State */}
          {isLoading && (
            <div className="text-center py-12">
              <div className="inline-block animate-spin rounded-full h-8 w-8 border-b-2 border-blue-500"></div>
              <p className="text-muted-foreground mt-4">Searching...</p>
            </div>
          )}

          {/* Empty State */}
          {!isLoading && !error && query && results.length === 0 && (
            <div className="text-center py-12 bg-background rounded-lg border border-border">
              <SearchIcon className="w-12 h-12 text-text-tertiary mx-auto mb-4" />
              <p className="text-muted-foreground text-lg mb-2">
                No results found for "{query}"
              </p>
              <p className="text-muted-foreground text-sm">
                Try different keywords or filters
              </p>
            </div>
          )}

          {/* Results List */}
          {!isLoading && !error && results.length > 0 && (
            <div className="space-y-4">
              {results.map((result) => (
                <SearchResultCard
                  key={`${result.type}-${result.id}`}
                  result={result}
                />
              ))}
            </div>
          )}

          {/* Pagination */}
          {meta && (page > 1 || meta.has_more) && (
            <div className="flex justify-center gap-2 mt-8">
              {page > 1 && (
                <button
                  onClick={() => {
                    const params = new URLSearchParams(searchParams);
                    params.set('page', (page - 1).toString());
                    setSearchParams(params);
                  }}
                  className="px-4 py-2 bg-surface hover:bg-surface-hover text-white rounded border border-border transition-colors"
                >
                  Previous
                </button>
              )}
              <span className="px-4 py-2 bg-background text-muted-foreground rounded border border-border">
                Page {page}
              </span>
              {meta.has_more && (
                <button
                  onClick={() => {
                    const params = new URLSearchParams(searchParams);
                    params.set('page', (page + 1).toString());
                    setSearchParams(params);
                  }}
                  className="px-4 py-2 bg-surface hover:bg-surface-hover text-white rounded border border-border transition-colors"
                >
                  Next
                </button>
              )}
            </div>
          )}
        </div>
      </Container>
    </>
  );
}
