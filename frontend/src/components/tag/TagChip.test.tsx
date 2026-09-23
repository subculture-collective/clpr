import { render, screen } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';
import { describe, expect, it, vi } from 'vitest';
import userEvent from '@testing-library/user-event';
import { TagChip } from './TagChip';
import type { Tag } from '../../types/tag';

const base = { id: 'tag-1', usage_count: 12, created_at: '2026-08-11T00:00:00Z' };

function renderChip(tag: Tag, props: Partial<React.ComponentProps<typeof TagChip>> = {}) {
    return render(<MemoryRouter><TagChip tag={tag} {...props} /></MemoryRouter>);
}

describe('TagChip', () => {
    it('links namespaced tags through one encoded path segment', () => {
        renderChip({ ...base, name: 'Content: highlights', slug: 'content/highlights', lane: 'detected', display_name: 'Highlights', evidence: 'contextual' });
        expect(screen.getByRole('link', { name: 'Highlights, context' })).toHaveAttribute('href', '/tags/content%2Fhighlights');
    });

    it('shows the evidence a detected tag definition requires', () => {
        renderChip({ ...base, name: 'Content: boss-fight', slug: 'content/boss-fight', lane: 'detected', display_name: 'Boss Fight', evidence: 'strong' });
        const link = screen.getByRole('link', { name: 'Boss Fight, outcome' });
        expect(link).toHaveTextContent('Boss Fight');
        expect(link).toHaveTextContent('Outcome');
        expect(link).toHaveAttribute('data-lane', 'detected');
    });

    it('strips stored prefixes and marks streamer channel tags when the API predates lanes', () => {
        renderChip({ ...base, name: 'Streamer: Community Oriented', slug: 'streamer/community-oriented' });
        const link = screen.getByRole('link', { name: 'Community Oriented, streamer tag' });
        expect(link).toHaveAttribute('data-lane', 'streamer');
        expect(link.className).toContain('border-dashed');
    });

    it('prefixes community tags with a hash and ignores provider colours', () => {
        renderChip({ ...base, name: 'goosebumps', slug: 'goosebumps', color: '#708090', lane: 'community', display_name: 'goosebumps' });
        const link = screen.getByRole('link', { name: '#goosebumps' });
        expect(link).toHaveTextContent('#goosebumps');
        expect(link).not.toHaveStyle({ backgroundColor: '#708090' });
    });

    it('keeps removal independent from navigation', async () => {
        const remove = vi.fn();
        renderChip({ ...base, id: 'english', name: 'English', slug: 'english', lane: 'community', display_name: 'English' }, { removable: true, onRemove: remove });
        const label = screen.getByRole('link', { name: '#English' });
        const button = screen.getByRole('button', { name: 'Remove English tag' });
        expect(label.contains(button)).toBe(false);
        await userEvent.setup().click(button);
        expect(remove).toHaveBeenCalledWith('english');
    });
});
