import { render, screen } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';
import { describe, expect, it, vi } from 'vitest';
import userEvent from '@testing-library/user-event';
import { TagChip } from './TagChip';

describe('TagChip', () => {
    it('links namespaced tags to the canonical tags page', () => {
        render(
            <MemoryRouter>
                <TagChip
                    tag={{
                        id: 'tag-1',
                        name: 'Highlights',
                        slug: 'content/highlights',
                        usage_count: 12,
                        created_at: '2026-08-11T00:00:00Z',
                    }}
                />
            </MemoryRouter>,
        );

        expect(screen.getByRole('link', { name: 'Highlights' })).toHaveAttribute(
            'href',
            '/tags/content%2Fhighlights',
        );
    });
    it('keeps the observed mid-tone language tag readable and removal independent from navigation', async () => {
        const remove = vi.fn();
        render(<MemoryRouter><TagChip tag={{ id: 'english', name: 'English', slug: 'english', color: '#708090', usage_count: 1, created_at: '' }} removable onRemove={remove} /></MemoryRouter>);
        const label = screen.getByRole('link', { name: 'English' });
        expect(label).toHaveStyle({ color: '#000000', backgroundColor: '#708090' });
        const button = screen.getByRole('button', { name: 'Remove English tag' });
        expect(label.contains(button)).toBe(false);
        await userEvent.setup().click(button);
        expect(remove).toHaveBeenCalledWith('english');
    });
});
