import { describe, it, expect, beforeEach } from 'vitest';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { MemoryRouter, Route, Routes, useLocation } from 'react-router-dom';
import { SavedSearches } from './SavedSearches';

describe('SavedSearches - Integration', () => {
  beforeEach(() => {
    // Clear localStorage before each test
    localStorage.clear();
  });

  const renderComponent = () => {
    const queryClient = new QueryClient({
      defaultOptions: {
        queries: {
          retry: false,
        },
      },
    });

    return render(
      <QueryClientProvider client={queryClient}>
        <MemoryRouter>
          <SavedSearches />
        </MemoryRouter>
      </QueryClientProvider>
    );
  };

  it('should not render when there are no saved searches', () => {
    const { container } = renderComponent();
    expect(container).toBeEmptyDOMElement();
  });

  it('should display saved searches loaded from localStorage', () => {
    // Pre-populate localStorage with saved searches
    const mockSearches = [
      {
        id: '1',
        query: 'test query 1',
        name: 'Test Search 1',
        created_at: new Date().toISOString(),
      },
      {
        id: '2',
        query: 'test query 2',
        created_at: new Date().toISOString(),
      },
    ];
    localStorage.setItem('savedSearches', JSON.stringify(mockSearches));

    renderComponent();

    expect(screen.getByText('Saved Searches')).toBeInTheDocument();
    expect(screen.getByText('Test Search 1')).toBeInTheDocument();
    expect(screen.getByText('test query 2')).toBeInTheDocument();
  });

  it('should display filter count for searches with filters', () => {
    const mockSearches = [
      {
        id: '1',
        query: 'filtered query',
        name: 'Filtered Search',
        filters: {
          language: 'en',
          gameId: 'game-1',
          minVotes: 5,
        },
        created_at: new Date().toISOString(),
      },
    ];
    localStorage.setItem('savedSearches', JSON.stringify(mockSearches));

    renderComponent();

    expect(screen.getByText('Filtered Search')).toBeInTheDocument();
    expect(screen.getByText('3 filter(s)')).toBeInTheDocument();
  });

  it('should remove a saved search when delete button is clicked', async () => {
    const mockSearches = [
      {
        id: '1',
        query: 'test query',
        name: 'Test Search',
        created_at: new Date().toISOString(),
      },
    ];
    localStorage.setItem('savedSearches', JSON.stringify(mockSearches));

    renderComponent();

    const deleteButton = screen.getByTestId('delete-saved-search-1');
    fireEvent.click(deleteButton);

    await waitFor(() => {
      expect(screen.queryByText('Test Search')).not.toBeInTheDocument();
    });

    // Verify it was removed from localStorage
    const stored = JSON.parse(localStorage.getItem('savedSearches') || '[]');
    expect(stored).toHaveLength(0);
  });

  it('should clear all saved searches when "Clear all" is clicked', async () => {
    const mockSearches = [
      {
        id: '1',
        query: 'test query 1',
        name: 'Test Search 1',
        created_at: new Date().toISOString(),
      },
      {
        id: '2',
        query: 'test query 2',
        name: 'Test Search 2',
        created_at: new Date().toISOString(),
      },
    ];
    localStorage.setItem('savedSearches', JSON.stringify(mockSearches));

    renderComponent();

    const clearButton = screen.getByTestId('clear-saved-searches');
    fireEvent.click(clearButton);

    // Confirm in the modal
    const confirmButton = await screen.findByText('Clear All');
    fireEvent.click(confirmButton);

    await waitFor(() => {
      expect(screen.queryByText('Test Search 1')).not.toBeInTheDocument();
      expect(screen.queryByText('Test Search 2')).not.toBeInTheDocument();
    });

    // Verify localStorage is cleared
    const stored = localStorage.getItem('savedSearches');
    expect(stored).toBeNull();
  });

  it('should not clear searches when modal is cancelled', async () => {
    const mockSearches = [
      {
        id: '1',
        query: 'test query',
        name: 'Test Search',
        created_at: new Date().toISOString(),
      },
    ];
    localStorage.setItem('savedSearches', JSON.stringify(mockSearches));

    renderComponent();

    const clearButton = screen.getByTestId('clear-saved-searches');
    fireEvent.click(clearButton);

    // Cancel in the modal
    const cancelButton = await screen.findByText('Cancel');
    fireEvent.click(cancelButton);

    // Search should still be visible
    expect(screen.getByText('Test Search')).toBeInTheDocument();

    // Verify localStorage is unchanged
    const stored = JSON.parse(localStorage.getItem('savedSearches') || '[]');
    expect(stored).toHaveLength(1);
  });

  it('reopens a said-only search with its phrase', () => {
    localStorage.setItem('savedSearches', JSON.stringify([
      { id: '1', query: '', filters: { said: 'cleanest ace' }, created_at: new Date().toISOString() },
    ]));
    const Location = () => <span data-testid='location'>{useLocation().search}</span>;
    render(
      <QueryClientProvider client={new QueryClient()}>
        <MemoryRouter>
          <Routes>
            <Route path='/' element={<SavedSearches />} />
            <Route path='/search' element={<Location />} />
          </Routes>
        </MemoryRouter>
      </QueryClientProvider>,
    );
    expect(screen.getByText('1 filter(s)')).toBeInTheDocument();
    fireEvent.click(screen.getByText('said "cleanest ace"'));
    expect(screen.getByTestId('location')).toHaveTextContent('?said=cleanest+ace');
  });
});
