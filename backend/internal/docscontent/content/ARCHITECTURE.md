# Architecture Diagram: Saved Searches Feature

```
┌─────────────────────────────────────────────────────────────────────┐
│                          SearchPage Component                        │
│                                                                       │
│  ┌────────────────┐                   ┌─────────────────────────┐   │
│  │  SearchBar     │                   │  SaveSearchModal        │   │
│  │                │                   │  (user input)           │   │
│  └────────────────┘                   └───────────┬─────────────┘   │
│                                                    │                 │
│                                                    │ name            │
│  ┌────────────────────────────────────────────────▼──────────────┐  │
│  │           useSavedSearches Hook                                │  │
│  │  • saveSearch(query, filters, name)                           │  │
│  │  • deleteSavedSearch(id)                                      │  │
│  │  • clearSavedSearches()                                       │  │
│  │  • savedSearches (reactive state)                             │  │
│  └───────────┬─────────────────────────────┬─────────────────────┘  │
│              │                              │                        │
└──────────────┼──────────────────────────────┼────────────────────────┘
               │                              │
               │                              │
               ▼                              ▼
    ┌──────────────────┐         ┌───────────────────────┐
    │  search-api.ts   │         │  SavedSearches        │
    │                  │         │  Component            │
    │  • getSaved...() │         │  (displays list)      │
    │  • saveSearch()  │         │                       │
    │  • deleteSaved() │         │  • Click to re-run    │
    │  • clearSaved()  │         │  • Delete button      │
    └────────┬─────────┘         │  • Clear all          │
             │                   └───────────────────────┘
             │
             ▼
    ┌──────────────────┐
    │   localStorage   │
    │                  │
    │  Key: 'saved     │
    │       Searches'  │
    │                  │
    │  [{              │
    │    id,           │
    │    query,        │
    │    filters,      │
    │    name,         │
    │    created_at    │
    │  }]              │
    └──────────────────┘
             │
             │ storage event
             │ (cross-tab sync)
             │
             ▼
    ┌──────────────────┐
    │   Other Tabs     │
    │  (auto-refresh)  │
    └──────────────────┘
```

## Data Flow

### Saving a Search

```
User Input (SearchPage)
    │
    ▼
SaveSearchModal
    │
    │ name
    ▼
handleSaveSearch()
    │
    ▼
useSavedSearches.saveSearch()
    │
    ▼
search-api.saveSearch()
    │
    ├─► Generate UUID
    ├─► Create SavedSearch object
    ├─► Add to array
    └─► Save to localStorage
         │
         ▼
    localStorage.setItem()
         │
         ├─► Storage event fired
         │   (cross-tab sync)
         │
         ▼
    loadSavedSearches()
         │
         ▼
    setSavedSearches()
         │
         ▼
    SavedSearches component re-renders
         │
         ▼
    ✅ New search visible immediately
```

### Deleting a Search

```
User clicks Delete button
    │
    ▼
handleDelete(id)
    │
    ▼
useSavedSearches.deleteSavedSearch(id)
    │
    ▼
search-api.deleteSavedSearch(id)
    │
    ├─► Filter out search by id
    └─► Save to localStorage
         │
         ▼
    setSavedSearches(filtered)
         │
         ▼
    SavedSearches component re-renders
         │
         ▼
    ✅ Search removed immediately
```

### Re-running a Saved Search

```
User clicks on SavedSearch
    │
    ▼
handleSearchClick(search)
    │
    ├─► Build URL with query
    ├─► Add filter params
    │   • game_id
    │   • language
    │   • min_votes
    │   • date_from
    │   • date_to
    │   • tags
    │
    ▼
navigate(`/search?${params}`)
    │
    ▼
SearchPage loads with params
    │
    ▼
useQuery triggers search
    │
    ▼
✅ Results displayed
```

## State Management Flow

```
┌────────────────────────────────────────┐
│        useSavedSearches Hook           │
│                                        │
│  State:                                │
│  • savedSearches: SavedSearch[]        │
│  • loading: boolean                    │
│                                        │
│  Effects:                              │
│  • Load on mount                       │
│  • Listen for storage events           │
│                                        │
│  Methods:                              │
│  • saveSearch                          │
│  • deleteSavedSearch                   │
│  • clearSavedSearches                  │
│  • refresh                             │
│                                        │
└─────────┬──────────────────────────────┘
          │
          │ provides
          │
          ▼
┌────────────────────────────────────────┐
│      SearchPage & SavedSearches        │
│                                        │
│  • Access reactive state               │
│  • Call mutation methods               │
│  • UI auto-updates                     │
│                                        │
└────────────────────────────────────────┘
```

## Component Hierarchy

```
App
└── SearchPage
    ├── SearchBar
    ├── SearchFilters
    ├── SaveSearchModal
    │   └── (triggered by save button)
    │
    ├── Sidebar
    │   ├── TrendingSearches
    │   ├── SearchHistory
    │   │   └── uses: useSearchHistory()
    │   └── SavedSearches
    │       └── uses: useSavedSearches()
    │
    └── SearchResults
        └── (clips, creators, games, tags)
```

## Storage Schema

```typescript
// localStorage key: 'savedSearches'
interface SavedSearch {
  id: string;              // UUID
  query: string;           // Search query
  filters?: {              // Optional filters
    gameId?: string;
    language?: string;
    minVotes?: number;
    dateFrom?: string;
    dateTo?: string;
    tags?: string[];
  };
  name?: string;           // Optional custom name
  created_at: string;      // ISO timestamp
}

// localStorage key: 'searchHistory'
interface SearchHistoryItem {
  query: string;
  result_count: number;
  created_at: string;
}
```

## Synchronization

```
Tab A                          localStorage                    Tab B
  │                                 │                            │
  │  saveSearch()                   │                            │
  ├────────────────────────────────►│                            │
  │                                 │                            │
  │                                 │  storage event             │
  │                                 ├───────────────────────────►│
  │                                 │                            │
  │                                 │              handleStorageChange()
  │                                 │                            │
  │                                 │              loadSavedSearches()
  │                                 │◄───────────────────────────┤
  │                                 │                            │
  │                                 │              setSavedSearches()
  │                                 │                            │
  │                                 │         ✅ UI updated      │
  │                                 │                            │
```

## Error Handling

```
loadSavedSearches()
    │
    ├─► Try: searchApi.getSavedSearches()
    │        │
    │        ├─► Parse localStorage
    │        │
    │        ├─► JSON parse error?
    │        │   └─► Catch: Return []
    │        │
    │        └─► Return parsed data
    │
    ├─► Catch: console.error()
    │   └─► setSavedSearches([])
    │
    └─► Finally: setLoading(false)
```

## Testing Strategy

```
Unit Tests (useSavedSearches.test.ts)
    ├─► Hook initialization
    ├─► Save functionality
    ├─► Delete functionality
    ├─► Clear functionality
    ├─► localStorage persistence
    ├─► Error handling
    └─► Refresh functionality

Integration Tests (SavedSearches.test.tsx)
    ├─► Component rendering
    ├─► User interactions
    ├─► localStorage integration
    ├─► Modal workflows
    └─► State updates

Integration Tests (SearchHistory.test.tsx)
    ├─► Component rendering
    ├─► User interactions
    ├─► localStorage integration
    ├─► Clear functionality
    └─► Error handling
```

This architecture ensures:
- ✅ Reactive updates without page refresh
- ✅ Persistent storage across sessions
- ✅ Cross-tab synchronization
- ✅ Error resilience
- ✅ Clean separation of concerns
- ✅ Testable components
