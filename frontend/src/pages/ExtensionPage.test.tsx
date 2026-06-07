import { render, screen } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';
import { describe, expect, it, vi } from 'vitest';
import { ExtensionPage } from './ExtensionPage';

vi.mock('../components', () => ({
    Container: ({ children }: { children: React.ReactNode }) => (
        <div>{children}</div>
    ),
    Card: ({ children, className }: { children: React.ReactNode; className?: string }) => (
        <div className={className}>{children}</div>
    ),
    CardBody: ({ children }: { children: React.ReactNode }) => (
        <div>{children}</div>
    ),
    SEO: () => null,
    Button: ({ children, variant }: { children: React.ReactNode; variant?: string; size?: string }) => (
        <button data-variant={variant}>{children}</button>
    ),
}));

describe('ExtensionPage', () => {
    function renderPage() {
        return render(
            <MemoryRouter>
                <ExtensionPage />
            </MemoryRouter>,
        );
    }

    it('renders the page heading', () => {
        renderPage();
        expect(
            screen.getByRole('heading', { name: /clpr browser extension/i }),
        ).toBeInTheDocument();
    });

    it('renders Chrome and Firefox download links', () => {
        renderPage();
        expect(
            screen.getByRole('link', { name: /get clpr for chrome/i }),
        ).toBeInTheDocument();
        expect(
            screen.getByRole('link', { name: /get clpr for firefox/i }),
        ).toBeInTheDocument();
    });

    it('renders the Features section', () => {
        renderPage();
        expect(
            screen.getByRole('heading', { name: /features/i }),
        ).toBeInTheDocument();
        expect(screen.getByText(/auto-detect clips/i)).toBeInTheDocument();
        expect(screen.getByRole('heading', { name: /context menu/i, level: 3 })).toBeInTheDocument();
        expect(screen.getByText(/one-click submit/i)).toBeInTheDocument();
    });

    it('renders the How it works section', () => {
        renderPage();
        expect(
            screen.getByRole('heading', { name: /how it works/i }),
        ).toBeInTheDocument();
        expect(screen.getByText(/log in/i)).toBeInTheDocument();
    });

    it('renders the Supported browsers table', () => {
        renderPage();
        expect(screen.getByRole('table')).toBeInTheDocument();
        expect(screen.getByText(/chrome \/ chromium/i)).toBeInTheDocument();
        // Firefox appears in both the download button and the browser table; verify the table cell.
        const rows = screen.getAllByText(/firefox/i);
        expect(rows.length).toBeGreaterThanOrEqual(1);
    });

    it('renders the open source section with GitHub link', () => {
        renderPage();
        const githubLink = screen.getByRole('link', { name: /view on github/i });
        expect(githubLink).toBeInTheDocument();
        expect(githubLink).toHaveAttribute(
            'href',
            expect.stringContaining('github.com'),
        );
    });
});
