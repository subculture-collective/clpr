import { render, screen } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';
import { describe, expect, it } from 'vitest';
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
});
