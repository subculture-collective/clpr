import { useCallback, useEffect, useRef, useState } from 'react';
import { User } from 'lucide-react';
import { searchApi } from '../../lib/search-api';
import type { SearchSuggestion } from '../../types/search';
import { Input } from './Input';
import { Badge } from './Badge';

export interface StreamerInputProps {
    value: string;
    onChange: (value: string) => void;
    autoDetected?: boolean;
    disabled?: boolean;
    id?: string;
    required?: boolean;
}

export function StreamerInput({
    value,
    onChange,
    autoDetected = false,
    disabled = false,
    id,
    required = false,
}: StreamerInputProps) {
    const [suggestions, setSuggestions] = useState<SearchSuggestion[]>([]);
    const [showSuggestions, setShowSuggestions] = useState(false);
    const [isLoading, setIsLoading] = useState(false);
    const [selectedIndex, setSelectedIndex] = useState(-1);
    const inputRef = useRef<HTMLInputElement>(null);
    const suggestionsRef = useRef<HTMLDivElement>(null);
    const debounceTimerRef = useRef<ReturnType<typeof setTimeout> | undefined>(
        undefined
    );

    const fetchSuggestions = useCallback(async (query: string) => {
        if (query.length < 2) {
            setSuggestions([]);
            return;
        }

        setIsLoading(true);
        try {
            const results = await searchApi.getSuggestions(query);
            const normalizedResults = Array.isArray(results) ? results : [];
            // Filter to only show creator suggestions
            const creatorSuggestions = normalizedResults.filter(
                s => s.type === 'creator'
            );
            setSuggestions(creatorSuggestions);
            setShowSuggestions(true);
        } catch (error) {
            console.error('Failed to fetch streamer suggestions:', error);
            setSuggestions([]);
        } finally {
            setIsLoading(false);
        }
    }, []);

    const handleInputChange = (e: React.ChangeEvent<HTMLInputElement>) => {
        const newValue = e.target.value;
        onChange(newValue);
        setSelectedIndex(-1);

        if (debounceTimerRef.current) {
            clearTimeout(debounceTimerRef.current);
        }

        debounceTimerRef.current = setTimeout(() => {
            fetchSuggestions(newValue);
        }, 300);
    };

    const handleKeyDown = (e: React.KeyboardEvent<HTMLInputElement>) => {
        if (!showSuggestions || suggestions.length === 0) return;

        switch (e.key) {
            case 'ArrowDown':
                e.preventDefault();
                setSelectedIndex(prev =>
                    prev < suggestions.length - 1 ? prev + 1 : prev
                );
                break;
            case 'ArrowUp':
                e.preventDefault();
                setSelectedIndex(prev => (prev > 0 ? prev - 1 : -1));
                break;
            case 'Enter':
                e.preventDefault();
                if (selectedIndex >= 0 && selectedIndex < suggestions.length) {
                    onChange(suggestions[selectedIndex].text);
                    setShowSuggestions(false);
                }
                break;
            case 'Escape':
                setShowSuggestions(false);
                setSelectedIndex(-1);
                break;
        }
    };

    const handleSuggestionClick = (suggestion: SearchSuggestion) => {
        onChange(suggestion.text);
        setShowSuggestions(false);
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

    return (
        <div>
            <div className='flex items-center justify-between mb-2'>
                <label
                    htmlFor={id || 'streamer_input'}
                    className='block text-sm font-medium'
                >
                    Streamer{' '}
                    {required && <span className='text-red-500'>*</span>}
                </label>
                {autoDetected && (
                    <Badge variant='info' size='sm'>
                        Will be auto-detected from clip
                    </Badge>
                )}
            </div>
            <div className='relative'>
                <Input
                    ref={inputRef}
                    id={id || 'streamer_input'}
                    type='text'
                    value={value}
                    onChange={handleInputChange}
                    onKeyDown={handleKeyDown}
                    onFocus={() =>
                        suggestions.length > 0 && setShowSuggestions(true)
                    }
                    placeholder='Enter streamer name...'
                    disabled={disabled}
                    required={required}
                />
                {showSuggestions && (suggestions.length > 0 || isLoading) && (
                    <div
                        ref={suggestionsRef}
                        className='absolute z-50 mt-2 w-full rounded-lg border border-border bg-background shadow-lg'
                    >
                        {isLoading && (
                            <div className='p-4 text-center text-sm text-muted-foreground'>
                                Loading suggestions...
                            </div>
                        )}
                        {!isLoading && suggestions.length > 0 && (
                            <ul className='max-h-60 overflow-y-auto py-2'>
                                {suggestions.map((suggestion, index) => (
                                    <li
                                        key={`${suggestion.type}-${suggestion.text}`}
                                    >
                                        <button
                                            type='button'
                                            onClick={() =>
                                                handleSuggestionClick(
                                                    suggestion
                                                )
                                            }
                                            className={`w-full px-4 py-2 text-left hover:bg-accent flex items-center gap-3 ${
                                                index === selectedIndex
                                                    ? 'bg-accent'
                                                    : ''
                                            }`}
                                        >
                                            <span className='flex-shrink-0'><User size={16} strokeWidth={1.75} /></span>
                                            <div className='flex-1 min-w-0'>
                                                <div className='font-medium truncate'>
                                                    {suggestion.text}
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
            <p className='text-xs text-muted-foreground mt-1'>
                {autoDetected
                    ? 'Streamer will be detected from the clip URL. You can override by typing here.'
                    : 'Type to search for streamers or enter manually'}
            </p>
        </div>
    );
}
