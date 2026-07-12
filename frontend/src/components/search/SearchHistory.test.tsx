import { describe, it, expect, beforeEach, vi } from 'vitest';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { MemoryRouter } from 'react-router-dom';
import { SearchHistory } from './SearchHistory';

vi.mock('../../context/AuthContext', () => ({
  useAuth: () => ({ isAuthenticated: false }),
}));

describe('SearchHistory - Integration', () => {
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
          <SearchHistory />
        </MemoryRouter>
      </QueryClientProvider>
    );
  };

  it('should not render when there is no search history', () => {
    const { container } = renderComponent();
    expect(container).toBeEmptyDOMElement();
  });

  it('should display search history loaded from localStorage', () => {
    // Pre-populate localStorage with search history
    const mockHistory = [
      {
        query: 'test query 1',
        result_count: 10,
        created_at: new Date().toISOString(),
      },
      {
        query: 'test query 2',
        result_count: 5,
        created_at: new Date().toISOString(),
      },
    ];
    localStorage.setItem('searchHistory', JSON.stringify(mockHistory));

    renderComponent();

    expect(screen.getByText('Recent Searches')).toBeInTheDocument();
    expect(screen.getByText('test query 1')).toBeInTheDocument();
    expect(screen.getByText('test query 2')).toBeInTheDocument();
    expect(screen.getByText('10')).toBeInTheDocument();
    expect(screen.getByText('5')).toBeInTheDocument();
  });

  it('should clear history when clear button is clicked and confirmed', async () => {
    const mockHistory = [
      {
        query: 'test query',
        result_count: 10,
        created_at: new Date().toISOString(),
      },
    ];
    localStorage.setItem('searchHistory', JSON.stringify(mockHistory));

    renderComponent();

    const clearButton = screen.getByTestId('clear-history-button');
    fireEvent.click(clearButton);

    // Confirm in the modal - use getAllByText and select the confirm button (index 1)
    const clearButtons = await screen.findAllByText('Clear');
    fireEvent.click(clearButtons[1]); // The confirm button in the modal

    await waitFor(() => {
      expect(screen.queryByText('test query')).not.toBeInTheDocument();
    });

    // Verify localStorage is cleared
    const stored = localStorage.getItem('searchHistory');
    expect(stored).toBeNull();
  });

  it('should not clear history when modal is cancelled', async () => {
    const mockHistory = [
      {
        query: 'test query',
        result_count: 10,
        created_at: new Date().toISOString(),
      },
    ];
    localStorage.setItem('searchHistory', JSON.stringify(mockHistory));

    renderComponent();

    const clearButton = screen.getByTestId('clear-history-button');
    fireEvent.click(clearButton);

    // Cancel in the modal
    const cancelButton = await screen.findByText('Cancel');
    fireEvent.click(cancelButton);

    // History should still be visible
    expect(screen.getByText('test query')).toBeInTheDocument();

    // Verify localStorage is unchanged
    const stored = JSON.parse(localStorage.getItem('searchHistory') || '[]');
    expect(stored).toHaveLength(1);
  });

  it('should limit displayed items based on maxItems prop', () => {
    // Create 15 history items
    const mockHistory = Array.from({ length: 15 }, (_, i) => ({
      query: `test query ${i + 1}`,
      result_count: i,
      created_at: new Date().toISOString(),
    }));
    localStorage.setItem('searchHistory', JSON.stringify(mockHistory));

    const queryClient = new QueryClient({
      defaultOptions: {
        queries: {
          retry: false,
        },
      },
    });

    render(
      <QueryClientProvider client={queryClient}>
        <MemoryRouter>
          <SearchHistory maxItems={5} />
        </MemoryRouter>
      </QueryClientProvider>
    );

    // Should only show 5 items
    expect(screen.getByText('test query 1')).toBeInTheDocument();
    expect(screen.getByText('test query 5')).toBeInTheDocument();
    expect(screen.queryByText('test query 6')).not.toBeInTheDocument();
  });

  it('should handle corrupted localStorage data gracefully', () => {
    // Set invalid JSON in localStorage
    localStorage.setItem('searchHistory', 'invalid json');

    const { container } = renderComponent();

    // Should not crash and should not render anything
    expect(container).toBeEmptyDOMElement();
  });
});
