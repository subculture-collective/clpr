import {
    useCallback,
    useEffect,
    useMemo,
    useRef,
    useState,
    type ChangeEvent,
    type KeyboardEvent,
} from 'react';
import { cn } from '@/lib/utils';
import {
    parseQuery,
    QueryParseError,
    type FilterName,
    SORT_ORDERS,
    RESULT_TYPES,
    LANGUAGE_CODES,
    FLAGS,
    USER_ROLES,
    RELATIVE_DATES,
} from '@/lib/query-parser';

/**
 * Configuration constants
 */
const MAX_SUGGESTIONS = 10;
const VALIDATION_DEBOUNCE_MS = 300;

/**
 * Operator descriptions for range filters
 */
const OPERATOR_DESCRIPTIONS: Record<string, string> = {
    '>': 'Greater than',
    '>=': 'Greater or equal',
    '<': 'Less than',
    '<=': 'Less or equal',
    '..': 'Range (e.g., 10..100)',
};

/**
 * Autocomplete suggestion item
 */
interface AutocompleteSuggestion {
    /** Display text for the suggestion */
    text: string;
    /** Category label for grouping */
    category: string;
    /** Description/help text */
    description: string;
    /** Value to insert when selected */
    insertValue: string;
}

/**
 * Validation error with position information
 */
interface ValidationError {
    message: string;
    code: string;
    suggestions: string[];
    position?: {
        line: number;
        column: number;
        offset: number;
    };
}

/**
 * Filter metadata for autocomplete
 */
interface FilterMeta {
    name: FilterName;
    description: string;
    examples: string[];
    valueType: 'string' | 'range' | 'date' | 'enum' | 'flag';
    enumValues?: readonly string[];
}

/**
 * Filter metadata configuration
 */
const FILTER_METADATA: FilterMeta[] = [
    {
        name: 'game',
        description: 'Filter by Twitch category (legacy game: syntax)',
        examples: ['game:valorant', 'game:"League of Legends"'],
        valueType: 'string',
    },
    {
        name: 'creator',
        description: 'Filter by clip creator',
        examples: ['creator:shroud', 'creator:"ninja"'],
        valueType: 'string',
    },
    {
        name: 'broadcaster',
        description: 'Filter by creator channel',
        examples: ['broadcaster:pokimane'],
        valueType: 'string',
    },
    {
        name: 'tag',
        description: 'Filter by tag',
        examples: ['tag:clutch', 'tag:funny'],
        valueType: 'string',
    },
    {
        name: 'language',
        description: 'Filter by language code',
        examples: ['language:en', 'language:es'],
        valueType: 'enum',
        enumValues: LANGUAGE_CODES,
    },
    {
        name: 'duration',
        description: 'Filter by clip duration (seconds)',
        examples: ['duration:>30', 'duration:30..60'],
        valueType: 'range',
    },
    {
        name: 'views',
        description: 'Filter by view count',
        examples: ['views:>1000', 'views:1000..10000'],
        valueType: 'range',
    },
    {
        name: 'votes',
        description: 'Filter by vote count',
        examples: ['votes:>50', 'votes:>=10'],
        valueType: 'range',
    },
    {
        name: 'after',
        description: 'Clips after date',
        examples: ['after:2025-01-01', 'after:yesterday'],
        valueType: 'date',
    },
    {
        name: 'before',
        description: 'Clips before date',
        examples: ['before:2025-12-31', 'before:last-week'],
        valueType: 'date',
    },
    {
        name: 'is',
        description: 'Boolean flag filters',
        examples: ['is:featured', 'is:nsfw'],
        valueType: 'flag',
        enumValues: FLAGS,
    },
    {
        name: 'sort',
        description: 'Sort order for results',
        examples: ['sort:popular', 'sort:recent'],
        valueType: 'enum',
        enumValues: SORT_ORDERS,
    },
    {
        name: 'type',
        description: 'Result type filter',
        examples: ['type:clips', 'type:creators'],
        valueType: 'enum',
        enumValues: RESULT_TYPES,
    },
    {
        name: 'karma',
        description: 'Filter by user uppies',
        examples: ['karma:>100', 'karma:>=50'],
        valueType: 'range',
    },
    {
        name: 'role',
        description: 'Filter by user role',
        examples: ['role:moderator', 'role:admin'],
        valueType: 'enum',
        enumValues: USER_ROLES,
    },
];

/**
 * Example queries for quick start
 */
const EXAMPLE_QUERIES = [
    {
        label: 'Popular Valorant clips',
        query: 'game:valorant votes:>50 sort:popular',
    },
    {
        label: 'Recent featured clips',
        query: 'after:last-week is:featured sort:recent',
    },
    {
        label: 'Funny or fail clips',
        query: '(tag:funny OR tag:fail) -is:nsfw',
    },
    {
        label: 'Long clips with many views',
        query: 'duration:>60 views:>10000',
    },
];

interface QueryBuilderProps {
    /** Initial query value */
    initialQuery?: string;
    /** Callback when query is submitted */
    onSearch?: (query: string) => void;
    /** Callback when query changes */
    onChange?: (query: string) => void;
    /** Placeholder text */
    placeholder?: string;
    /** Auto-focus on mount */
    autoFocus?: boolean;
    /** Additional CSS classes */
    className?: string;
    /** Show example queries section */
    showExamples?: boolean;
    /** Aria label for the input */
    ariaLabel?: string;
}

/**
 * QueryBuilder component provides an advanced query input with autocomplete,
 * syntax highlighting, and real-time validation.
 */
export function QueryBuilder({
    initialQuery = '',
    onSearch,
    onChange,
    placeholder = 'Build your query... (e.g., game:valorant tag:clutch votes:>50)',
    autoFocus = false,
    className,
    showExamples = true,
    ariaLabel = 'Advanced query builder',
}: QueryBuilderProps) {
    const [query, setQuery] = useState(initialQuery);
    const [suggestions, setSuggestions] = useState<AutocompleteSuggestion[]>(
        []
    );
    const [showSuggestions, setShowSuggestions] = useState(false);
    const [selectedIndex, setSelectedIndex] = useState(-1);
    const [validationError, setValidationError] =
        useState<ValidationError | null>(null);
    const [isValid, setIsValid] = useState(true);

    const inputRef = useRef<HTMLInputElement>(null);
    const suggestionsRef = useRef<HTMLDivElement>(null);
    const debounceTimerRef = useRef<ReturnType<typeof setTimeout> | undefined>(
        undefined
    );

    /**
     * Generate autocomplete suggestions based on current input and cursor position
     */
    const generateSuggestions = useCallback(
        (
            inputValue: string,
            cursorPosition: number
        ): AutocompleteSuggestion[] => {
            const suggestions: AutocompleteSuggestion[] = [];

            // Get the word at cursor position
            const beforeCursor = inputValue.slice(0, cursorPosition);
            const afterCursor = inputValue.slice(cursorPosition);

            // Find the current word boundary
            const wordMatch = beforeCursor.match(/(\S+)$/);
            const currentWord = wordMatch ? wordMatch[1].toLowerCase() : '';

            // Check if we're after a colon (completing filter value)
            const colonMatch = beforeCursor.match(/(\w+):(\S*)$/);

            if (colonMatch) {
                const filterName = colonMatch[1].toLowerCase();
                const filterMeta = FILTER_METADATA.find(
                    (f) => f.name === filterName
                );
                const partialValue = colonMatch[2].toLowerCase();

                if (filterMeta) {
                    // Suggest enum values
                    if (filterMeta.enumValues) {
                        filterMeta.enumValues
                            .filter(
                                (v) => !partialValue || v.startsWith(partialValue)
                            )
                            .forEach((value) => {
                                suggestions.push({
                                    text: value,
                                    category: `${filterName} values`,
                                    description: `Set ${filterName} to ${value}`,
                                    insertValue: `${filterName}:${value}${afterCursor.startsWith(' ') ? '' : ' '}`,
                                });
                            });
                    }

                    // Suggest date values for date filters
                    if (filterMeta.valueType === 'date') {
                        RELATIVE_DATES.filter(
                            (d) => !partialValue || d.startsWith(partialValue)
                        ).forEach((date) => {
                            suggestions.push({
                                text: date,
                                category: 'Relative dates',
                                description: `Use ${date} as date`,
                                insertValue: `${filterName}:${date}${afterCursor.startsWith(' ') ? '' : ' '}`,
                            });
                        });

                        // Suggest ISO date format
                        const currentYear = new Date().getFullYear().toString();
                        if (!partialValue || currentYear.startsWith(partialValue)) {
                            suggestions.push({
                                text: 'YYYY-MM-DD',
                                category: 'Date format',
                                description: `Use ISO date format (e.g., ${currentYear}-01-15)`,
                                insertValue: `${filterName}:`,
                            });
                        }
                    }

                    // Suggest range operators
                    if (filterMeta.valueType === 'range') {
                        const operators = ['>', '>=', '<', '<=', '..'];
                        operators
                            .filter(
                                (op) => !partialValue || op.startsWith(partialValue)
                            )
                            .forEach((op) => {
                                suggestions.push({
                                    text: op,
                                    category: 'Comparison operators',
                                    description: OPERATOR_DESCRIPTIONS[op] || op,
                                    insertValue: `${filterName}:${op}`,
                                });
                            });
                    }
                }
            } else if (currentWord) {
                // Suggest filter names that match the current word
                FILTER_METADATA.filter((f) =>
                    f.name.startsWith(currentWord)
                ).forEach((filter) => {
                    suggestions.push({
                        text: `${filter.name}:`,
                        category: 'Filters',
                        description: filter.description,
                        insertValue: `${filter.name}:`,
                    });
                });

                // Suggest OR operator
                if ('or'.startsWith(currentWord)) {
                    suggestions.push({
                        text: 'OR',
                        category: 'Operators',
                        description: 'Boolean OR operator',
                        insertValue: 'OR ',
                    });
                }

                // Suggest negation
                if (currentWord === '-') {
                    FILTER_METADATA.forEach((filter) => {
                        suggestions.push({
                            text: `-${filter.name}:`,
                            category: 'Negated filters',
                            description: `Exclude ${filter.description.toLowerCase()}`,
                            insertValue: `-${filter.name}:`,
                        });
                    });
                }
            } else if (!inputValue.trim()) {
                // Show all filters when input is empty
                FILTER_METADATA.forEach((filter) => {
                    suggestions.push({
                        text: `${filter.name}:`,
                        category: 'Filters',
                        description: filter.description,
                        insertValue: `${filter.name}:`,
                    });
                });
            }

            return suggestions.slice(0, MAX_SUGGESTIONS);
        },
        []
    );

    /**
     * Validate query using the parser
     */
    const validateQuery = useCallback((queryStr: string): void => {
        if (!queryStr.trim()) {
            setValidationError(null);
            setIsValid(true);
            return;
        }

        try {
            parseQuery(queryStr);
            setValidationError(null);
            setIsValid(true);
        } catch (error) {
            if (error instanceof QueryParseError) {
                setValidationError({
                    message: error.message,
                    code: error.code,
                    suggestions: error.suggestions,
                    position: error.position,
                });
                setIsValid(false);
            }
        }
    }, []);

    /**
     * Handle input changes with debounced validation
     */
    const handleInputChange = (e: ChangeEvent<HTMLInputElement>) => {
        const newQuery = e.target.value;
        setQuery(newQuery);
        setSelectedIndex(-1);

        // Clear previous debounce timer
        if (debounceTimerRef.current) {
            clearTimeout(debounceTimerRef.current);
        }

        // Generate suggestions immediately
        const cursorPosition = e.target.selectionStart || 0;
        const newSuggestions = generateSuggestions(newQuery, cursorPosition);
        setSuggestions(newSuggestions);
        setShowSuggestions(newSuggestions.length > 0);

        // Debounce validation
        debounceTimerRef.current = setTimeout(() => {
            validateQuery(newQuery);
        }, VALIDATION_DEBOUNCE_MS);

        // Notify parent of change
        onChange?.(newQuery);
    };

    /**
     * Handle suggestion selection
     */
    const selectSuggestion = useCallback(
        (suggestion: AutocompleteSuggestion) => {
            const input = inputRef.current;
            if (!input) return;

            const cursorPosition = input.selectionStart || 0;
            const beforeCursor = query.slice(0, cursorPosition);
            const afterCursor = query.slice(cursorPosition);

            // Find where to replace from
            const wordMatch = beforeCursor.match(/(\S+)$/);
            const replaceFrom = wordMatch
                ? beforeCursor.length - wordMatch[1].length
                : beforeCursor.length;

            // Build new query - ensure trailing space for better UX
            const insertValueWithSpace = suggestion.insertValue.endsWith(' ')
                ? suggestion.insertValue
                : suggestion.insertValue + ' ';
            const newQuery =
                query.slice(0, replaceFrom) +
                insertValueWithSpace +
                afterCursor.trimStart();

            setQuery(newQuery);
            setShowSuggestions(false);
            setSelectedIndex(-1);

            // Validate immediately
            validateQuery(newQuery);

            // Focus input and set cursor position
            requestAnimationFrame(() => {
                input.focus();
                const newCursorPos = replaceFrom + insertValueWithSpace.length;
                input.setSelectionRange(newCursorPos, newCursorPos);
            });

            onChange?.(newQuery);
        },
        [query, validateQuery, onChange]
    );

    /**
     * Handle keyboard navigation
     */
    const handleKeyDown = (e: KeyboardEvent<HTMLInputElement>) => {
        // Handle suggestions navigation
        if (showSuggestions && suggestions.length > 0) {
            switch (e.key) {
                case 'ArrowDown':
                    e.preventDefault();
                    setSelectedIndex((prev) =>
                        prev < suggestions.length - 1 ? prev + 1 : prev
                    );
                    return;
                case 'ArrowUp':
                    e.preventDefault();
                    setSelectedIndex((prev) => (prev > 0 ? prev - 1 : -1));
                    return;
                case 'Tab':
                case 'Enter':
                    if (selectedIndex >= 0) {
                        e.preventDefault();
                        selectSuggestion(suggestions[selectedIndex]);
                        return;
                    }
                    break;
                case 'Escape':
                    e.preventDefault();
                    setShowSuggestions(false);
                    setSelectedIndex(-1);
                    return;
            }
        }

        // Submit on Enter when no suggestion is selected
        if (e.key === 'Enter' && !e.shiftKey) {
            e.preventDefault();
            if (isValid && query.trim()) {
                onSearch?.(query);
            }
        }

        // Ctrl+Enter to force submit even with errors
        if (e.key === 'Enter' && e.ctrlKey) {
            e.preventDefault();
            if (query.trim()) {
                onSearch?.(query);
            }
        }
    };

    /**
     * Handle clicking outside to close suggestions
     */
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

    /**
     * Cleanup debounce timer on unmount
     */
    useEffect(() => {
        return () => {
            if (debounceTimerRef.current) {
                clearTimeout(debounceTimerRef.current);
            }
        };
    }, []);

    /**
     * Scroll selected suggestion into view when navigating with keyboard
     */
    useEffect(() => {
        if (selectedIndex >= 0 && suggestionsRef.current) {
            const selectedElement = suggestionsRef.current.querySelector(
                '[aria-selected="true"]'
            );
            if (selectedElement && typeof selectedElement.scrollIntoView === 'function') {
                selectedElement.scrollIntoView({ block: 'nearest' });
            }
        }
    }, [selectedIndex]);

    /**
     * Handle example query click
     */
    const handleExampleClick = useCallback(
        (exampleQuery: string) => {
            setQuery(exampleQuery);
            validateQuery(exampleQuery);
            onChange?.(exampleQuery);
            inputRef.current?.focus();
        },
        [validateQuery, onChange]
    );

    /**
     * Get categorized suggestions for display
     */
    const categorizedSuggestions = useMemo(() => {
        const categories: Record<string, AutocompleteSuggestion[]> = {};
        suggestions.forEach((suggestion) => {
            if (!categories[suggestion.category]) {
                categories[suggestion.category] = [];
            }
            categories[suggestion.category].push(suggestion);
        });
        return categories;
    }, [suggestions]);

    return (
        <div className={cn('w-full space-y-4', className)}>
            {/* Query input */}
            <div className='relative'>
                <div className='relative'>
                    <input
                        ref={inputRef}
                        type='text'
                        value={query}
                        onChange={handleInputChange}
                        onKeyDown={handleKeyDown}
                        onFocus={() => {
                            const cursorPosition =
                                inputRef.current?.selectionStart || 0;
                            const newSuggestions = generateSuggestions(
                                query,
                                cursorPosition
                            );
                            setSuggestions(newSuggestions);
                            setShowSuggestions(newSuggestions.length > 0);
                        }}
                        placeholder={placeholder}
                        autoFocus={autoFocus}
                        aria-label={ariaLabel}
                        aria-invalid={!isValid}
                        aria-describedby={
                            validationError
                                ? 'query-error'
                                : 'query-helper'
                        }
                        aria-expanded={showSuggestions}
                        aria-haspopup='listbox'
                        aria-autocomplete='list'
                        role='combobox'
                        className={cn(
                            'w-full px-4 py-3 rounded-lg border transition-colors font-mono text-sm',
                            'bg-background text-foreground',
                            'placeholder:text-muted-foreground',
                            'focus:outline-none focus:ring-2 focus:ring-offset-2',
                            isValid
                                ? 'border-border hover:border-primary-300 focus:ring-primary-500 focus:border-transparent'
                                : 'border-error-500 focus:ring-error-500'
                        )}
                    />
                    {/* Search button */}
                    <button
                        type='button'
                        onClick={() => isValid && query.trim() && onSearch?.(query)}
                        disabled={!isValid || !query.trim()}
                        className={cn(
                            'absolute right-2 top-1/2 -translate-y-1/2 px-3 py-1.5 rounded-md text-sm font-medium transition-colors',
                            isValid && query.trim()
                                ? 'bg-primary-500 text-white hover:bg-primary-600'
                                : 'bg-neutral-200 text-neutral-400 cursor-not-allowed dark:bg-neutral-700 dark:text-neutral-500'
                        )}
                        aria-label='Search'
                    >
                        Search
                    </button>
                </div>

                {/* Autocomplete suggestions dropdown */}
                {showSuggestions && suggestions.length > 0 && (
                    <div
                        ref={suggestionsRef}
                        className='absolute z-50 mt-2 w-full rounded-lg border border-border bg-background shadow-lg max-h-80 overflow-y-auto'
                        role='listbox'
                        aria-label='Query suggestions'
                    >
                        {Object.entries(categorizedSuggestions).map(
                            ([category, items]) => (
                                <div key={category}>
                                    <div className='px-3 py-1.5 text-xs font-semibold text-muted-foreground bg-accent/50 sticky top-0'>
                                        {category}
                                    </div>
                                    {items.map((suggestion, idx) => {
                                        const globalIndex =
                                            suggestions.indexOf(suggestion);
                                        return (
                                            <button
                                                key={`${suggestion.text}-${idx}`}
                                                type='button'
                                                role='option'
                                                aria-selected={
                                                    globalIndex === selectedIndex
                                                }
                                                onClick={() =>
                                                    selectSuggestion(suggestion)
                                                }
                                                className={cn(
                                                    'w-full px-3 py-2 text-left hover:bg-accent transition-colors flex items-center justify-between gap-2',
                                                    globalIndex === selectedIndex &&
                                                        'bg-accent'
                                                )}
                                            >
                                                <span className='font-mono text-sm'>
                                                    {suggestion.text}
                                                </span>
                                                <span className='text-xs text-muted-foreground truncate'>
                                                    {suggestion.description}
                                                </span>
                                            </button>
                                        );
                                    })}
                                </div>
                            )
                        )}
                    </div>
                )}
            </div>

            {/* Validation error display */}
            {validationError && (
                <div
                    id='query-error'
                    className='p-3 rounded-lg bg-error-50 dark:bg-error-950 border border-error-200 dark:border-error-800'
                    role='alert'
                >
                    <div className='flex items-start gap-2'>
                        <svg
                            className='w-5 h-5 text-error-500 flex-shrink-0 mt-0.5'
                            fill='none'
                            viewBox='0 0 24 24'
                            stroke='currentColor'
                        >
                            <path
                                strokeLinecap='round'
                                strokeLinejoin='round'
                                strokeWidth={2}
                                d='M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z'
                            />
                        </svg>
                        <div className='flex-1'>
                            <p className='text-sm font-medium text-error-700 dark:text-error-300'>
                                {validationError.message}
                            </p>
                            {validationError.suggestions.length > 0 && (
                                <ul className='mt-2 text-xs text-error-600 dark:text-error-400 space-y-1'>
                                    {validationError.suggestions.map(
                                        (suggestion, idx) => (
                                            <li key={idx}>• {suggestion}</li>
                                        )
                                    )}
                                </ul>
                            )}
                        </div>
                    </div>
                </div>
            )}

            {/* Helper text when valid */}
            {!validationError && (
                <p
                    id='query-helper'
                    className='text-xs text-muted-foreground'
                >
                    Press <kbd className='px-1 py-0.5 bg-accent rounded text-xs font-mono'>Tab</kbd> to autocomplete,{' '}
                    <kbd className='px-1 py-0.5 bg-accent rounded text-xs font-mono'>Enter</kbd> to search
                </p>
            )}

            {/* Example queries */}
            {showExamples && (
                <div className='space-y-2'>
                    <p className='text-sm font-medium text-foreground'>
                        Try an example:
                    </p>
                    <div className='flex flex-wrap gap-2'>
                        {EXAMPLE_QUERIES.map((example) => (
                            <button
                                key={example.query}
                                type='button'
                                onClick={() => handleExampleClick(example.query)}
                                className='px-3 py-1.5 text-xs rounded-full border border-border hover:border-primary-300 hover:bg-accent transition-colors'
                            >
                                {example.label}
                            </button>
                        ))}
                    </div>
                </div>
            )}

            {/* Keyboard shortcuts legend */}
            <div className='text-xs text-muted-foreground space-y-1'>
                <p className='font-medium'>Keyboard shortcuts:</p>
                <ul className='grid grid-cols-2 gap-x-4 gap-y-1'>
                    <li>
                        <kbd className='px-1 py-0.5 bg-accent rounded font-mono'>↑↓</kbd> Navigate suggestions
                    </li>
                    <li>
                        <kbd className='px-1 py-0.5 bg-accent rounded font-mono'>Tab</kbd> Accept suggestion
                    </li>
                    <li>
                        <kbd className='px-1 py-0.5 bg-accent rounded font-mono'>Enter</kbd> Search
                    </li>
                    <li>
                        <kbd className='px-1 py-0.5 bg-accent rounded font-mono'>Esc</kbd> Close suggestions
                    </li>
                </ul>
            </div>
        </div>
    );
}
