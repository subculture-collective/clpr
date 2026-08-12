import { render, screen } from '@testing-library/react';
import { MemoryRouter, Route, Routes } from 'react-router-dom';
import { describe, expect, it, vi } from 'vitest';
import { TagPage } from './TagPage';

// Mock the components
vi.mock('../components', () => ({
    Container: ({ children }: { children: React.ReactNode }) => (
        <div data-testid="container">{children}</div>
    ),
    SEO: () => null,
}));

// Mock ClipFeed component
vi.mock('../components/clip', () => ({
    ClipFeed: ({ title, description, filters }: { 
        title?: string; 
        description?: string; 
        filters?: { tags?: string[] };
    }) => (
        <div data-testid="clip-feed">
            <h1>{title}</h1>
            <p>{description}</p>
            <div data-testid="filters">{JSON.stringify(filters)}</div>
        </div>
    ),
}));

describe('TagPage', () => {
    it('renders namespaced tags from the canonical route', () => {
        render(
            <MemoryRouter initialEntries={['/tags/content%2Fhighlights']}>
                <Routes>
                    <Route path="/tags/:tagSlug" element={<TagPage />} />
                </Routes>
            </MemoryRouter>
        );

        expect(screen.getByText('#content/highlights')).toBeInTheDocument();
        expect(screen.getByTestId('filters').textContent).toContain(
            '"tags":["content/highlights"]',
        );
    });

    it('renders ClipFeed with tag filter when tagSlug is provided', () => {
        render(
            <MemoryRouter initialEntries={['/tags/gaming']}>
                <Routes>
                    <Route path="/tags/:tagSlug" element={<TagPage />} />
                </Routes>
            </MemoryRouter>
        );

        // Should render ClipFeed
        expect(screen.getByTestId('clip-feed')).toBeInTheDocument();
        
        // Should display the tag in the title
        expect(screen.getByText('#gaming')).toBeInTheDocument();
        
        // Should display the tag in the description
        expect(screen.getByText('Clips tagged with gaming')).toBeInTheDocument();
        
        // Should pass the tag as a filter
        const filtersElement = screen.getByTestId('filters');
        expect(filtersElement.textContent).toContain('"tags":["gaming"]');
    });

    it('renders different tag correctly', () => {
        render(
            <MemoryRouter initialEntries={['/tags/funny']}>
                <Routes>
                    <Route path="/tags/:tagSlug" element={<TagPage />} />
                </Routes>
            </MemoryRouter>
        );

        expect(screen.getByText('#funny')).toBeInTheDocument();
        expect(screen.getByText('Clips tagged with funny')).toBeInTheDocument();
        
        const filtersElement = screen.getByTestId('filters');
        expect(filtersElement.textContent).toContain('"tags":["funny"]');
    });

    it('handles missing tagSlug gracefully', () => {
        render(
            <MemoryRouter initialEntries={['/tags/']}>
                <Routes>
                    <Route path="/tags/:tagSlug?" element={<TagPage />} />
                </Routes>
            </MemoryRouter>
        );

        // Should show error message when no tag is specified
        expect(screen.getByText('No tag specified')).toBeInTheDocument();
        
        // Should not render ClipFeed
        expect(screen.queryByTestId('clip-feed')).not.toBeInTheDocument();
    });
});
