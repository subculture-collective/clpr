import { useCallback, useEffect, useRef, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { Gamepad2, User, Tag, Search } from 'lucide-react';
import { searchApi } from '../../lib/search-api';
import type { SearchSuggestion } from '../../types/search';
import { Input } from '../ui/Input';

interface SearchBarProps {
    initialQuery?: string;
    autoFocus?: boolean;
    onSearch?: (query: string) => void;
}

export function SearchBar({
    initialQuery = '',
    autoFocus = false,
    onSearch,
}: SearchBarProps) {
    const [query, setQuery] = useState(initialQuery);
    const [suggestions, setSuggestions] = useState<SearchSuggestion[]>([]);
    const [showSuggestions, setShowSuggestions] = useState(false);
    const [isLoading, setIsLoading] = useState(false);
    const [selectedIndex, setSelectedIndex] = useState(-1);
    const inputRef = useRef<HTMLInputElement>(null);
    const suggestionsRef = useRef<HTMLDivElement>(null);
    const navigate = useNavigate();
    const debounceTimerRef = useRef<ReturnType<typeof setTimeout> | undefined>(
        undefined
    );

    const fetchSuggestions = useCallback(async (searchQuery: string) => {
        if (searchQuery.length < 2) {
            setSuggestions([]);
            return;
        }

        setIsLoading(true);
        try {
            const results = await searchApi.getSuggestions(searchQuery);
            setSuggestions(results);
            setShowSuggestions(true);
        } catch (error) {
            console.error('Failed to fetch suggestions:', error);
            setSuggestions([]);
        } finally {
            setIsLoading(false);
        }
    }, []);

    const handleInputChange = (e: React.ChangeEvent<HTMLInputElement>) => {
        const newQuery = e.target.value;
        setQuery(newQuery);
        setSelectedIndex(-1);

        if (debounceTimerRef.current) {
            clearTimeout(debounceTimerRef.current);
        }

        debounceTimerRef.current = setTimeout(() => {
            fetchSuggestions(newQuery);
        }, 300);
    };

    const handleSearch = (searchQuery: string) => {
        if (!searchQuery.trim()) return;

        setShowSuggestions(false);
        setSuggestions([]);

        if (onSearch) {
            onSearch(searchQuery);
        } else {
            navigate(`/search?q=${encodeURIComponent(searchQuery)}`);
        }
    };

    const handleSubmit = (e: React.FormEvent) => {
        e.preventDefault();
        handleSearch(query);
    };

    const handleKeyDown = (e: React.KeyboardEvent<HTMLInputElement>) => {
        if (!showSuggestions || suggestions.length === 0) return;

        switch (e.key) {
            case 'ArrowDown':
                e.preventDefault();
                setSelectedIndex((prev) =>
                    prev < suggestions.length - 1 ? prev + 1 : prev
                );
                break;
            case 'ArrowUp':
                e.preventDefault();
                setSelectedIndex((prev) => (prev > 0 ? prev - 1 : -1));
                break;
            case 'Enter':
                e.preventDefault();
                if (selectedIndex >= 0 && selectedIndex < suggestions.length) {
                    handleSearch(suggestions[selectedIndex].text);
                    setQuery(suggestions[selectedIndex].text);
                } else {
                    handleSearch(query);
                }
                break;
            case 'Escape':
                setShowSuggestions(false);
                setSelectedIndex(-1);
                break;
        }
    };

    const handleSuggestionClick = (suggestion: SearchSuggestion) => {
        setQuery(suggestion.text);
        handleSearch(suggestion.text);
    };

    useEffect(() => {
        const handleClickOutside = (event: MouseEvent) => {
            if (
                suggestionsRef.current &&
                !suggestionsRef.current.contains(event.target as Node) &&
                inputRef.current &&
                !inputRef.current.contains(event.target as Node)
            ) {
                setShowSuggestions(false);
            }
        };

        document.addEventListener('mousedown', handleClickOutside);
        return () =>
            document.removeEventListener('mousedown', handleClickOutside);
    }, []);

    useEffect(() => {
        return () => {
            if (debounceTimerRef.current) {
                clearTimeout(debounceTimerRef.current);
            }
        };
    }, []);

    const getSuggestionIcon = (type: string) => {
        switch (type) {
            case 'game':
                return <Gamepad2 size={16} strokeWidth={1.75} />;
            case 'creator':
                return <User size={16} strokeWidth={1.75} />;
            case 'tag':
                return <Tag size={16} strokeWidth={1.75} />;
            default:
                return <Search size={16} strokeWidth={1.75} />;
        }
    };

    return (
        <div className='relative w-full max-w-2xl'>
            <form onSubmit={handleSubmit}>
                <div className='relative'>
                    <Input
                        ref={inputRef}
                        type='search'
                        name='q'
                        value={query}
                        onChange={handleInputChange}
                        onKeyDown={handleKeyDown}
                        onFocus={() =>
                            suggestions.length > 0 && setShowSuggestions(true)
                        }
                        placeholder='Search clips, games, creators...'
                        aria-label='Search'
                        autoFocus={autoFocus}
                        className='pr-10'
                        data-testid='search-input'
                    />
                    <button
                        type='submit'
                        className='absolute right-3 top-1/2 -translate-y-1/2 text-muted-foreground hover:text-foreground'
                        tabIndex={-1}
                        aria-label='Search'
                    >
                        <svg
                            xmlns='http://www.w3.org/2000/svg'
                            className='h-5 w-5'
                            fill='none'
                            viewBox='0 0 24 24'
                            stroke='currentColor'
                        >
                            <path
                                strokeLinecap='round'
                                strokeLinejoin='round'
                                strokeWidth={2}
                                d='M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z'
                            />
                        </svg>
                    </button>
                    {query && (
                        <button
                            type='button'
                            onClick={() => {
                                setQuery('');
                                setSuggestions([]);
                                setShowSuggestions(false);
                                inputRef.current?.focus();
                            }}
                            className='absolute right-10 top-1/2 -translate-y-1/2 text-muted-foreground hover:text-foreground'
                            aria-label='Clear search'
                        >
                            <svg
                                xmlns='http://www.w3.org/2000/svg'
                                className='h-4 w-4'
                                fill='none'
                                viewBox='0 0 24 24'
                                stroke='currentColor'
                            >
                                <path
                                    strokeLinecap='round'
                                    strokeLinejoin='round'
                                    strokeWidth={2}
                                    d='M6 18L18 6M6 6l12 12'
                                />
                            </svg>
                        </button>
                    )}
                </div>
            </form>

            {showSuggestions && (suggestions.length > 0 || isLoading) && (
                <div
                    ref={suggestionsRef}
                    className='absolute z-50 mt-2 w-full rounded-lg border border-border bg-background shadow-lg'
                    data-testid='suggestions'
                >
                    {isLoading && (
                        <div className='p-4 text-center text-sm text-muted-foreground'>
                            Loading suggestions...
                        </div>
                    )}
                    {!isLoading && suggestions.length > 0 && (
                        <ul className='max-h-80 overflow-y-auto py-2'>
                            {suggestions.map((suggestion, index) => (
                                <li
                                    key={`${suggestion.type}-${suggestion.text}`}
                                >
                                    <button
                                        type='button'
                                        onClick={() =>
                                            handleSuggestionClick(suggestion)
                                        }
                                        className={`w-full px-4 py-2 text-left hover:bg-accent flex items-center gap-3 ${
                                            index === selectedIndex
                                                ? 'bg-accent'
                                                : ''
                                        }`}
                                    >
                                        <span className='flex-shrink-0'>
                                            {getSuggestionIcon(suggestion.type)}
                                        </span>
                                        <div className='flex-1 min-w-0'>
                                            <div className='font-medium truncate'>
                                                {suggestion.text}
                                            </div>
                                            <div className='text-xs text-muted-foreground capitalize'>
                                                {suggestion.type}
                                            </div>
                                        </div>
                                    </button>
                                </li>
                            ))}
                        </ul>
                    )}
                </div>
            )}
        </div>
    );
}
