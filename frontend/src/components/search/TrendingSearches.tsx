import { useEffect, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { searchApi } from '../../lib/search-api';
import type { TrendingSearch } from '../../types/search';

interface TrendingSearchesProps {
  limit?: number;
  days?: number;
  className?: string;
}

export function TrendingSearches({ limit = 10, days = 7, className = '' }: TrendingSearchesProps) {
  const [trending, setTrending] = useState<TrendingSearch[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const navigate = useNavigate();

  useEffect(() => {
    let mounted = true;

    async function loadTrending() {
      try {
        setLoading(true);
        const data = await searchApi.getTrendingSearches(days, limit);
        if (mounted) {
          setTrending(data);
          setError(null);
        }
      } catch (err) {
        if (mounted) {
          setError('Failed to load trending searches');
          console.error('Error loading trending searches:', err);
        }
      } finally {
        if (mounted) {
          setLoading(false);
        }
      }
    }

    loadTrending();

    return () => {
      mounted = false;
    };
  }, [days, limit]);

  const handleSearchClick = (query: string) => {
    navigate(`/search?q=${encodeURIComponent(query)}`);
  };

  if (loading) {
    return (
      <div className={`animate-pulse ${className}`}>
        <div className="h-6 bg-surface-raised rounded w-32 mb-3"></div>
        {Array.from({ length: 5 }).map((_, i) => (
          <div key={i} className="h-8 bg-surface-raised rounded mb-2"></div>
        ))}
      </div>
    );
  }

  if (error || trending.length === 0) {
    return null;
  }

  return (
    <div className={className} data-testid="trending-searches">
      <h3 className="text-sm font-semibold text-white mb-3">
        Trending Searches
      </h3>
      <div className="space-y-1">
        {trending.map((item, index) => (
          <button
            key={item.query}
            onClick={() => handleSearchClick(item.query)}
            className="w-full text-left px-3 py-2 rounded-lg hover:bg-surface-hover transition-colors group"
            data-testid={`trending-search-${index}`}
          >
            <div className="flex items-center justify-between gap-2">
              <div className="flex items-center gap-2 min-w-0 flex-1">
                <span className="text-xs font-medium text-muted-foreground w-5">
                  {index + 1}
                </span>
                <span className="text-sm text-white truncate group-hover:text-blue-400">
                  {item.query}
                </span>
              </div>
              <div className="flex items-center gap-2 text-xs text-muted-foreground">
                <span title="Search count">{item.search_count}</span>
                {item.avg_results > 0 && (
                  <span className="text-green-600 dark:text-green-400" title="Average results">
                    {item.avg_results}
                  </span>
                )}
              </div>
            </div>
          </button>
        ))}
      </div>
    </div>
  );
}
