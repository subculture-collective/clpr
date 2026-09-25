import { render, screen } from '@/test/test-utils';
import { describe, expect, it, vi } from 'vitest';
import { SearchBar } from './SearchBar';

vi.mock('../../lib/search-api', () => ({
    searchApi: { getSuggestions: vi.fn(async () => []) },
}));

describe('SearchBar', () => {
    it('shows only the custom clear button, hiding the native search cancel control', () => {
        render(<SearchBar initialQuery='moonmoon' />);

        const input = screen.getByTestId('search-input');
        expect(input.className).toContain('[&::-webkit-search-cancel-button]:appearance-none');
        expect(screen.getAllByRole('button', { name: 'Clear search' })).toHaveLength(1);
    });
});
