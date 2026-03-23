import { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { useSearchHistory } from '../../hooks/useSearchHistory';
import { ConfirmModal } from '../ui/ConfirmModal';

interface SearchHistoryProps {
  className?: string;
  maxItems?: number;
}

export function SearchHistory({ className = '', maxItems = 10 }: SearchHistoryProps) {
  const { history, loading, clearHistory } = useSearchHistory();
  const navigate = useNavigate();
  const [showConfirmModal, setShowConfirmModal] = useState(false);

  const handleSearchClick = (query: string) => {
    navigate(`/search?q=${encodeURIComponent(query)}`);
  };

  const handleClearHistory = async () => {
    await clearHistory();
  };

  if (loading) {
    return (
      <div className={`animate-pulse ${className}`}>
        <div className="h-6 bg-surface-raised rounded w-32 mb-3"></div>
        {Array.from({ length: 3 }).map((_, i) => (
          <div key={i} className="h-8 bg-surface-raised rounded mb-2"></div>
        ))}
      </div>
    );
  }

  if (history.length === 0) {
    return null;
  }

  const displayHistory = history.slice(0, maxItems);

  return (
    <>
      <div className={className} data-testid="search-history">
        <div className="flex items-center justify-between mb-3">
          <h3 className="text-sm font-semibold text-white">
            Recent Searches
          </h3>
          <button
            onClick={() => setShowConfirmModal(true)}
            className="text-xs text-muted-foreground hover:text-foreground"
            data-testid="clear-history-button"
          >
            Clear
          </button>
        </div>
      <div className="space-y-1">
        {displayHistory.map((item, index) => (
          <button
            key={`${item.query}-${index}`}
            onClick={() => handleSearchClick(item.query)}
            className="w-full text-left px-3 py-2 rounded-lg hover:bg-surface-hover transition-colors group"
            data-testid={`history-item-${index}`}
          >
            <div className="flex items-center justify-between gap-2">
              <span className="text-sm text-white truncate flex-1 group-hover:text-blue-400">
                {item.query}
              </span>
              <span className="text-xs text-muted-foreground">
                {item.result_count > 0 ? `${item.result_count}` : '0'}
              </span>
            </div>
          </button>
        ))}
      </div>
      </div>

      <ConfirmModal
        open={showConfirmModal}
        onClose={() => setShowConfirmModal(false)}
        onConfirm={handleClearHistory}
        title="Clear Search History"
        message="Are you sure you want to clear your search history? This action cannot be undone."
        confirmText="Clear"
        variant="danger"
      />
    </>
  );
}
