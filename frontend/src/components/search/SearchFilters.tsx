import { useState } from 'react';
import type { SearchFacets, SearchFilters as SearchFiltersType } from '../../types/search';

interface SearchFiltersProps {
    facets?: SearchFacets;
    filters: SearchFiltersType;
    onFiltersChange: (filters: SearchFiltersType) => void;
}

export function SearchFilters({ facets, filters, onFiltersChange }: SearchFiltersProps) {
    const [isExpanded, setIsExpanded] = useState(false);
    const [selectedDateRange, setSelectedDateRange] = useState<string | undefined>();

    if (!facets) {
        return null;
    }

    const hasFilters = filters.language || filters.gameId || filters.dateFrom || filters.dateTo;

    const handleLanguageToggle = (language: string) => {
        onFiltersChange({
            ...filters,
            language: filters.language === language ? undefined : language,
        });
    };

    const handleGameSelect = (gameName: string) => {
        onFiltersChange({
            ...filters,
            gameId: filters.gameId === gameName ? undefined : gameName,
        });
    };

    const handleDateRangeSelect = (range: string) => {
        const now = new Date();
        let dateFrom: string | undefined;

        switch (range) {
            case 'last_hour':
                dateFrom = new Date(now.getTime() - 60 * 60 * 1000).toISOString();
                break;
            case 'last_day':
                dateFrom = new Date(now.getTime() - 24 * 60 * 60 * 1000).toISOString();
                break;
            case 'last_week':
                dateFrom = new Date(now.getTime() - 7 * 24 * 60 * 60 * 1000).toISOString();
                break;
            case 'last_month':
                dateFrom = new Date(now.getTime() - 30 * 24 * 60 * 60 * 1000).toISOString();
                break;
            default:
                dateFrom = undefined;
        }

        setSelectedDateRange(range);
        onFiltersChange({
            ...filters,
            dateFrom,
            dateTo: dateFrom ? undefined : filters.dateTo,
        });
    };

    const handleClearFilters = () => {
        setSelectedDateRange(undefined);
        onFiltersChange({});
    };

    const hasAnyFacets = 
        (facets.languages && facets.languages.length > 0) ||
        (facets.games && facets.games.length > 0) ||
        facets.date_range;

    if (!hasAnyFacets) {
        return null;
    }

    return (
        <div className='border border-border rounded-lg p-4 mb-6'>
            <div className='flex items-center justify-between mb-4'>
                <h3 className='text-lg font-semibold'>Filters</h3>
                <div className='flex items-center gap-2'>
                    {hasFilters && (
                        <button
                            onClick={handleClearFilters}
                            className='text-sm text-primary hover:underline'
                        >
                            Clear all
                        </button>
                    )}
                    <button
                        onClick={() => setIsExpanded(!isExpanded)}
                        className='text-sm text-muted-foreground hover:text-foreground'
                    >
                        {isExpanded ? 'Hide' : 'Show'}
                    </button>
                </div>
            </div>

            {isExpanded && (
                <div className='space-y-4'>
                    {/* Language Facets */}
                    {facets.languages && facets.languages.length > 0 && (
                        <div>
                            <h4 className='text-sm font-medium mb-2'>Language</h4>
                            <div className='flex flex-wrap gap-2'>
                                {facets.languages.map((lang) => (
                                    <button
                                        key={lang.key}
                                        onClick={() => handleLanguageToggle(lang.key)}
                                        className={`px-3 py-1.5 text-sm rounded-full border transition-colors ${
                                            filters.language === lang.key
                                                ? 'bg-primary text-primary-foreground border-primary'
                                                : 'border-border hover:border-primary'
                                        }`}
                                    >
                                        {lang.label || lang.key.toUpperCase()} ({lang.count})
                                    </button>
                                ))}
                            </div>
                        </div>
                    )}

                    {/* Twitch category facets */}
                    {facets.games && facets.games.length > 0 && (
                        <div>
                            <h4 className='text-sm font-medium mb-2'>Twitch Categories</h4>
                            <div className='space-y-1 max-h-48 overflow-y-auto'>
                                {facets.games.slice(0, 10).map((game) => (
                                    <button
                                        key={game.key}
                                        onClick={() => handleGameSelect(game.key)}
                                        className={`w-full text-left px-3 py-2 text-sm rounded-md transition-colors ${
                                            filters.gameId === game.key
                                                ? 'bg-primary text-primary-foreground'
                                                : 'hover:bg-accent'
                                        }`}
                                    >
                                        <span className='font-medium'>{game.label || game.key}</span>
                                        <span className='ml-2 text-muted-foreground'>
                                            ({game.count})
                                        </span>
                                    </button>
                                ))}
                            </div>
                        </div>
                    )}

                    {/* Date Range Facets */}
                    {facets.date_range && (
                        <div>
                            <h4 className='text-sm font-medium mb-2'>Time Range</h4>
                            <div className='space-y-1'>
                                {facets.date_range.last_hour > 0 && (
                                    <button
                                        onClick={() => handleDateRangeSelect('last_hour')}
                                        className={`w-full text-left px-3 py-2 text-sm rounded-md transition-colors ${
                                            selectedDateRange === 'last_hour'
                                                ? 'bg-primary text-primary-foreground'
                                                : 'hover:bg-accent'
                                        }`}
                                    >
                                        Last Hour ({facets.date_range.last_hour})
                                    </button>
                                )}
                                {facets.date_range.last_day > 0 && (
                                    <button
                                        onClick={() => handleDateRangeSelect('last_day')}
                                        className={`w-full text-left px-3 py-2 text-sm rounded-md transition-colors ${
                                            selectedDateRange === 'last_day'
                                                ? 'bg-primary text-primary-foreground'
                                                : 'hover:bg-accent'
                                        }`}
                                    >
                                        Last 24 Hours ({facets.date_range.last_day})
                                    </button>
                                )}
                                {facets.date_range.last_week > 0 && (
                                    <button
                                        onClick={() => handleDateRangeSelect('last_week')}
                                        className={`w-full text-left px-3 py-2 text-sm rounded-md transition-colors ${
                                            selectedDateRange === 'last_week'
                                                ? 'bg-primary text-primary-foreground'
                                                : 'hover:bg-accent'
                                        }`}
                                    >
                                        Last Week ({facets.date_range.last_week})
                                    </button>
                                )}
                                {facets.date_range.last_month > 0 && (
                                    <button
                                        onClick={() => handleDateRangeSelect('last_month')}
                                        className={`w-full text-left px-3 py-2 text-sm rounded-md transition-colors ${
                                            selectedDateRange === 'last_month'
                                                ? 'bg-primary text-primary-foreground'
                                                : 'hover:bg-accent'
                                        }`}
                                    >
                                        Last Month ({facets.date_range.last_month})
                                    </button>
                                )}
                                {facets.date_range.older > 0 && (
                                    <button
                                        onClick={() => handleDateRangeSelect('older')}
                                        className={`w-full text-left px-3 py-2 text-sm rounded-md transition-colors ${
                                            selectedDateRange === 'older'
                                                ? 'bg-primary text-primary-foreground'
                                                : 'hover:bg-accent'
                                        }`}
                                    >
                                        Older ({facets.date_range.older})
                                    </button>
                                )}
                            </div>
                        </div>
                    )}
                </div>
            )}
        </div>
    );
}
