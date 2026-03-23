import { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { useSavedSearches } from '../../hooks/useSavedSearches';
import type { SavedSearch } from '../../types/search';
import { ConfirmModal } from '../ui/ConfirmModal';

interface SavedSearchesProps {
  className?: string;
}

export function SavedSearches({ className = '' }: SavedSearchesProps) {
  const { savedSearches, deleteSavedSearch, clearSavedSearches } = useSavedSearches();
  const [showConfirmModal, setShowConfirmModal] = useState(false);
  const navigate = useNavigate();

  const handleSearchClick = (search: SavedSearch) => {
    const params = new URLSearchParams({ q: search.query });
    
    if (search.filters) {
      const filters = search.filters;
      if (filters.gameId) params.append('game_id', filters.gameId);
      if (filters.language) params.append('language', filters.language);
      if (filters.minVotes) params.append('min_votes', filters.minVotes.toString());
      if (filters.dateFrom) params.append('date_from', filters.dateFrom);
      if (filters.dateTo) params.append('date_to', filters.dateTo);
      if (filters.tags && filters.tags.length > 0) {
        params.append('tags', filters.tags.join(','));
      }
    }
    
    navigate(`/search?${params.toString()}`);
  };

  const handleDelete = (id: string, e: React.MouseEvent) => {
    e.stopPropagation();
    deleteSavedSearch(id);
  };

  const handleClearAll = () => {
    clearSavedSearches();
  };

  if (savedSearches.length === 0) {
    return null;
  }

  return (
    <>
      <div className={className} data-testid="saved-searches">
        <div className="flex items-center justify-between mb-3">
          <h3 className="text-sm font-semibold text-white">
            Saved Searches
          </h3>
          <button
            onClick={() => setShowConfirmModal(true)}
            className="text-xs text-muted-foreground hover:text-foreground"
            data-testid="clear-saved-searches"
          >
            Clear all
          </button>
        </div>
      <div className="space-y-1">
        {savedSearches.map((search) => (
          <div
            key={search.id}
            className="group relative"
            data-testid={`saved-search-${search.id}`}
          >
            <button
              onClick={() => handleSearchClick(search)}
              className="w-full text-left px-3 py-2 rounded-lg hover:bg-surface-hover transition-colors pr-10"
            >
              <div className="flex flex-col gap-1">
                <span className="text-sm text-white truncate">
                  {search.name || search.query}
                </span>
                {search.name && (
                  <span className="text-xs text-muted-foreground truncate">
                    {search.query}
                  </span>
                )}
                {search.filters && (() => {
                  // Count only filters with actual values
                  const filterCount = [
                    search.filters.language,
                    search.filters.gameId,
                    search.filters.dateFrom,
                    search.filters.dateTo,
                    typeof search.filters.minVotes === 'number' && !isNaN(search.filters.minVotes) && search.filters.minVotes > 0,
                    Array.isArray(search.filters.tags) && search.filters.tags.length > 0,
                  ].filter(Boolean).length;
                  
                  return filterCount > 0 ? (
                    <span className="text-xs text-blue-600 dark:text-blue-400">
                      {filterCount} filter(s)
                    </span>
                  ) : null;
                })()}
              </div>
            </button>
            <button
              onClick={(e) => handleDelete(search.id, e)}
              className="absolute right-2 top-1/2 -translate-y-1/2 p-1.5 rounded opacity-0 group-hover:opacity-100 hover:bg-surface-hover transition-opacity"
              title="Delete saved search"
              data-testid={`delete-saved-search-${search.id}`}
            >
              <svg
                className="w-4 h-4 text-muted-foreground"
                fill="none"
                stroke="currentColor"
                viewBox="0 0 24 24"
              >
                <path
                  strokeLinecap="round"
                  strokeLinejoin="round"
                  strokeWidth={2}
                  d="M6 18L18 6M6 6l12 12"
                />
              </svg>
            </button>
          </div>
        ))}
      </div>
      </div>

      <ConfirmModal
        open={showConfirmModal}
        onClose={() => setShowConfirmModal(false)}
        onConfirm={handleClearAll}
        title="Clear All Saved Searches"
        message="Are you sure you want to clear all saved searches? This action cannot be undone."
        confirmText="Clear All"
        variant="danger"
      />
    </>
  );
}
