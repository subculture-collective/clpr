import type { FormEvent } from 'react';
import { useRef, useState } from 'react';
import { useTranslation } from 'react-i18next';
import { useNavigate } from 'react-router-dom';
import { Clock, Crown, Eye, Flame, MessageSquare, PlayCircle, Star, TrendingUp } from 'lucide-react';
import { Button, Input } from '@/components/ui';
import { useKeyboardShortcuts } from '@/hooks/useKeyboardShortcuts';
import { cn } from '@/lib/utils';
import type { SortOption, TimeFrame } from '@/types/clip';
import type { FeedAutoplayPreference } from '@/hooks';

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

interface FeedHeaderProps {
    title: string;
    description?: string;
    showSearch?: boolean;
    // Filter props (optional - if provided, shows inline filter controls)
    sort?: SortOption;
    timeframe?: TimeFrame;
    onSortChange?: (sort: SortOption) => void;
    onTimeframeChange?: (timeframe: TimeFrame) => void;
    autoplayPreference?: FeedAutoplayPreference;
    onAutoplayPreferenceChange?: (value: FeedAutoplayPreference) => void;
}

const sortOptions: { value: SortOption; label: string }[] = [
    { value: 'trending', label: 'Trending' },
    { value: 'popular', label: 'Most Popular' },
    { value: 'top', label: 'Top Rated' },
    { value: 'rising', label: 'Rising' },
    { value: 'new', label: 'Newest' },
    { value: 'discussed', label: 'Most Discussed' },
];

const timeframeOptions: { value: TimeFrame; label: string }[] = [
    { value: 'hour', label: 'Hour' },
    { value: 'day', label: 'Day' },
    { value: 'week', label: 'Week' },
    { value: 'month', label: 'Month' },
    { value: 'year', label: 'Year' },
    { value: 'all', label: 'All' },
];

export function FeedHeader({
    title,
    description,
    showSearch = false,
    sort,
    timeframe,
    onSortChange,
    onTimeframeChange,
    autoplayPreference,
    onAutoplayPreferenceChange,
}: FeedHeaderProps) {
    const { t } = useTranslation();
    const navigate = useNavigate();
    const [searchQuery, setSearchQuery] = useState('');
    const searchInputRef = useRef<HTMLInputElement>(null);

    useKeyboardShortcuts(
        showSearch ?
            [
                {
                    key: '/',
                    callback: () => searchInputRef.current?.focus(),
                    description: 'Focus search',
                },
            ]
        :   [],
    );

    const handleSearch = (e: FormEvent) => {
        e.preventDefault();
        const trimmed = searchQuery.trim();
        if (!trimmed) return;

        navigate(`/search?q=${encodeURIComponent(trimmed)}`);
        setSearchQuery('');
        searchInputRef.current?.blur();
    };

    const showFilters = sort !== undefined && onSortChange !== undefined;
    const showTimeframe =
        showFilters &&
        (sort === 'top' || sort === 'trending') &&
        onTimeframeChange;

    return (
        <div className='mb-4 px-4 md:mb-6 md:px-0'>
            {/* Main header row */}
            <div className='flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between'>
                <div className='flex-1 min-w-0'>
                    <h1 className='text-2xl sm:text-3xl font-bold truncate'>
                        {title}
                    </h1>
                    {description && (
                        <p className='text-muted-foreground text-sm mt-1'>
                            {description}
                        </p>
                    )}
                </div>

                <div className='flex flex-row flex-wrap gap-2 sm:items-center'>
                    {autoplayPreference && onAutoplayPreferenceChange && (
                        <button
                            type='button'
                            onClick={() => onAutoplayPreferenceChange(autoplayPreference === 'manual' ? 'muted' : 'manual')}
                            className='inline-flex min-h-11 shrink-0 items-center justify-center gap-2 rounded-lg border border-border px-3 text-sm text-foreground hover:bg-muted'
                            aria-pressed={autoplayPreference === 'muted'}
                            title='Choose whether visible clips play automatically'
                        >
                            <PlayCircle size={17} aria-hidden='true' />
                            {autoplayPreference === 'muted' ? 'Autoplay on' : 'Tap to play'}
                        </button>
                    )}
                    {/* Sort dropdown */}
                    {showFilters && (
                        <div className='relative flex flex-1 items-center sm:flex-none'>
                            <span className='absolute left-2.5 pointer-events-none text-muted-foreground'>
                                {sort && sortIcons[sort]}
                            </span>
                            <select
                                id='sort-select'
                                value={sort}
                                onChange={e =>
                                    onSortChange(e.target.value as SortOption)
                                }
                                className={cn(
                                    'w-full min-h-11 sm:w-auto pl-8 pr-3 py-2 rounded-lg border text-sm transition-colors',
                                    'bg-background text-foreground',
                                    'focus:outline-none focus:ring-2 focus:ring-primary-500 focus:border-transparent',
                                    'border-border hover:border-primary-300',
                                )}
                                aria-label='Sort clips by'
                            >
                                {sortOptions.map(option => (
                                    <option key={option.value} value={option.value}>
                                        {option.label}
                                    </option>
                                ))}
                            </select>
                        </div>
                    )}

                    {showSearch && (
                        <form
                            onSubmit={handleSearch}
                            className='hidden w-full sm:block sm:w-auto sm:min-w-60'
                            role='search'
                            aria-label={t('nav.search')}
                        >
                            <Input
                                ref={searchInputRef}
                                type='search'
                                placeholder={`${t('nav.search')} (/)`}
                                value={searchQuery}
                                onChange={e => setSearchQuery(e.target.value)}
                                aria-label={t('nav.search')}
                                className='text-sm'
                            />
                        </form>
                    )}
                </div>
            </div>

            {/* Timeframe row (only when top/trending selected) */}
            {showTimeframe && (
                <div className='flex flex-wrap gap-1.5 mt-3'>
                    {timeframeOptions.map(option => (
                        <Button
                            key={option.value}
                            variant={
                                timeframe === option.value ? 'primary' : 'ghost'
                            }
                            size='sm'
                            onClick={() => onTimeframeChange(option.value)}
                            className={cn(
                                'text-xs px-2.5 py-1',
                                timeframe === option.value && 'font-semibold',
                            )}
                        >
                            {option.label}
                        </Button>
                    ))}
                </div>
            )}
        </div>
    );
}
