import { Clock, Crown, Eye, Flame, MessageSquare, Star, TrendingUp } from 'lucide-react';
import { Button } from '@/components/ui';
import { cn } from '@/lib/utils';
import type { SortOption, TimeFrame } from '@/types/clip';

const sortIcons: Record<SortOption, React.ReactNode> = {
    trending: <Flame size={16} strokeWidth={1.75} />,
    popular: <Star size={16} strokeWidth={1.75} />,
    top: <Crown size={16} strokeWidth={1.75} />,
    rising: <TrendingUp size={16} strokeWidth={1.75} />,
    new: <Clock size={16} strokeWidth={1.75} />,
    discussed: <MessageSquare size={16} strokeWidth={1.75} />,
    hot: <Flame size={16} strokeWidth={1.75} />,
    views: <Eye size={16} strokeWidth={1.75} />,
};

interface FeedFiltersProps {
    sort: SortOption;
    timeframe?: TimeFrame;
    onSortChange: (sort: SortOption) => void;
    onTimeframeChange: (timeframe: TimeFrame) => void;
}

export function FeedFilters({
    sort,
    timeframe,
    onSortChange,
    onTimeframeChange,
}: FeedFiltersProps) {
    const sortOptions: { value: SortOption; label: string }[] = [
        { value: 'trending', label: 'Trending' },
        { value: 'popular', label: 'Most Popular' },
        { value: 'top', label: 'Top Rated' },
        { value: 'rising', label: 'Rising' },
        { value: 'new', label: 'Newest' },
        { value: 'discussed', label: 'Most Discussed' },
    ];

    const timeframeOptions: { value: TimeFrame; label: string }[] = [
        { value: 'hour', label: 'Past Hour' },
        { value: 'day', label: 'Past Day' },
        { value: 'week', label: 'Past Week' },
        { value: 'month', label: 'Past Month' },
        { value: 'year', label: 'Past Year' },
        { value: 'all', label: 'All Time' },
    ];

    return (
        <div className='bg-card border border-border rounded-xl p-4 mb-6'>
            {/* Sort dropdown */}
            <div className='flex flex-col gap-2 sm:flex-row sm:items-center sm:gap-2 mb-4'>
                <label
                    className='text-sm font-medium text-foreground'
                    htmlFor='sort-select'
                >
                    Sort by
                </label>
                <div className='relative flex items-center'>
                    <span className='absolute left-2.5 pointer-events-none text-muted-foreground'>
                        {sortIcons[sort]}
                    </span>
                    <select
                        id='sort-select'
                        value={sort}
                        onChange={e => onSortChange(e.target.value as SortOption)}
                        className={cn(
                            'w-full sm:w-48 pl-8 pr-2.5 py-2 rounded-lg border text-sm transition-colors min-h-10',
                            'bg-background text-foreground',
                            'focus:outline-none focus:ring-2 focus:ring-primary-500 focus:border-transparent',
                            'border-border hover:border-primary-300'
                        )}
                    >
                        {sortOptions.map(option => (
                            <option key={option.value} value={option.value}>
                                {option.label}
                            </option>
                        ))}
                    </select>
                </div>
            </div>

            {/* Timeframe selector (shown when Top or Trending is selected) */}
            {(sort === 'top' || sort === 'trending') && (
                <div className='flex flex-wrap gap-2'>
                    <span className='text-sm text-muted-foreground flex items-center mr-2'>
                        Timeframe:
                    </span>
                    {timeframeOptions.map(option => (
                        <Button
                            key={option.value}
                            variant={
                                timeframe === option.value ? 'primary' : 'ghost'
                            }
                            size='sm'
                            onClick={() => onTimeframeChange(option.value)}
                            className={cn(
                                'cursor-pointer',
                                timeframe === option.value &&
                                    'font-bold shadow-md'
                            )}
                        >
                            {timeframe === option.value && (
                                <svg
                                    className='w-4 h-4 mr-1'
                                    fill='currentColor'
                                    viewBox='0 0 20 20'
                                    aria-hidden='true'
                                >
                                    <path
                                        fillRule='evenodd'
                                        d='M16.707 5.293a1 1 0 010 1.414l-8 8a1 1 0 01-1.414 0l-4-4a1 1 0 011.414-1.414L8 12.586l7.293-7.293a1 1 0 011.414 0z'
                                        clipRule='evenodd'
                                    />
                                </svg>
                            )}
                            {option.label}
                        </Button>
                    ))}
                </div>
            )}
        </div>
    );
}
