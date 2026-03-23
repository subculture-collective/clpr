import { cn } from '@/lib/utils';
import type { ForumFilters as ForumFiltersType } from '@/types/forum';

interface ForumFiltersProps {
  filters: ForumFiltersType;
  onFilterChange: (filters: ForumFiltersType) => void;
  className?: string;
}

export function ForumFilters({
  filters,
  onFilterChange,
  className,
}: ForumFiltersProps) {
  const handleTagToggle = (tag: string) => {
    const currentTags = filters.tags || [];
    const newTags = currentTags.includes(tag)
      ? currentTags.filter((t) => t !== tag)
      : [...currentTags, tag];
    
    onFilterChange({ ...filters, tags: newTags });
  };

  const handleClearFilters = () => {
    onFilterChange({ game_id: undefined, tags: [] });
  };

  const hasActiveFilters = filters.game_id || (filters.tags && filters.tags.length > 0);

  return (
    <div className={cn('space-y-3', className)}>
      {/* Quick filters */}
      <div className="flex flex-wrap gap-2">
        <button
          type="button"
          onClick={() => handleTagToggle('help')}
          className={cn(
            'px-3 py-1.5 text-sm rounded-lg border transition-colors',
            filters.tags?.includes('help')
              ? 'bg-primary-600 border-primary-600 text-white'
              : 'bg-surface border-border text-foreground hover:bg-surface-hover'
          )}
        >
          Help
        </button>
        <button
          type="button"
          onClick={() => handleTagToggle('discussion')}
          className={cn(
            'px-3 py-1.5 text-sm rounded-lg border transition-colors',
            filters.tags?.includes('discussion')
              ? 'bg-primary-600 border-primary-600 text-white'
              : 'bg-surface border-border text-foreground hover:bg-surface-hover'
          )}
        >
          Discussion
        </button>
        <button
          type="button"
          onClick={() => handleTagToggle('suggestion')}
          className={cn(
            'px-3 py-1.5 text-sm rounded-lg border transition-colors',
            filters.tags?.includes('suggestion')
              ? 'bg-primary-600 border-primary-600 text-white'
              : 'bg-surface border-border text-foreground hover:bg-surface-hover'
          )}
        >
          Suggestion
        </button>
        <button
          type="button"
          onClick={() => handleTagToggle('bug')}
          className={cn(
            'px-3 py-1.5 text-sm rounded-lg border transition-colors',
            filters.tags?.includes('bug')
              ? 'bg-primary-600 border-primary-600 text-white'
              : 'bg-surface border-border text-foreground hover:bg-surface-hover'
          )}
        >
          Bug Report
        </button>

        {hasActiveFilters && (
          <button
            type="button"
            onClick={handleClearFilters}
            className="px-3 py-1.5 text-sm text-muted-foreground hover:text-white transition-colors"
          >
            Clear filters
          </button>
        )}
      </div>
    </div>
  );
}
