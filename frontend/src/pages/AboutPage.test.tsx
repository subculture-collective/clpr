import { render, screen } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';
import { describe, expect, it, vi } from 'vitest';
import { AboutPage } from './AboutPage';

// Mock the components (include SEO to avoid helmet interactions in tests)
vi.mock('../components', () => ({
    Container: ({ children }: { children: React.ReactNode }) => (
        <div>{children}</div>
    ),
    Card: ({ children, id }: { children: React.ReactNode; id?: string }) => (
        <div id={id}>{children}</div>
    ),
    CardBody: ({ children }: { children: React.ReactNode }) => (
        <div>{children}</div>
    ),
    SEO: () => null,
}));

describe('AboutPage', () => {
    it('renders the page title', () => {
        render(
            <MemoryRouter>
                <AboutPage />
            </MemoryRouter>
        );

        expect(screen.getByText('About clpr')).toBeInTheDocument();
    });

    it('displays the last updated date', () => {
        render(
            <MemoryRouter>
                <AboutPage />
            </MemoryRouter>
        );

        expect(screen.getByText(/Last updated:/i)).toBeInTheDocument();
    });

    it('renders key sections with headings', () => {
        render(
            <MemoryRouter>
                <AboutPage />
            </MemoryRouter>
        );

        expect(screen.getByText('What is clpr?')).toBeInTheDocument();
        expect(screen.getByText('How It Works')).toBeInTheDocument();
        expect(screen.getByText('Key Features')).toBeInTheDocument();
        expect(screen.getByText('Open Source & Community')).toBeInTheDocument();
        expect(screen.getByText('Technology Stack')).toBeInTheDocument();
        expect(screen.getByText('Get in Touch')).toBeInTheDocument();
    });

    it('has anchor IDs for navigation', () => {
        const { container } = render(
            <MemoryRouter>
                <AboutPage />
            </MemoryRouter>
        );

        expect(container.querySelector('#what-is-clpr')).toBeInTheDocument();
        expect(container.querySelector('#how-it-works')).toBeInTheDocument();
        expect(container.querySelector('#features')).toBeInTheDocument();
        expect(container.querySelector('#open-source')).toBeInTheDocument();
        expect(container.querySelector('#tech-stack')).toBeInTheDocument();
        expect(container.querySelector('#contact')).toBeInTheDocument();
    });

    it('links to GitHub repository', () => {
        render(
            <MemoryRouter>
                <AboutPage />
            </MemoryRouter>
        );

        const githubLinks = screen.getAllByText('View on GitHub');
        expect(githubLinks.length).toBeGreaterThan(0);
        expect(githubLinks[0]).toHaveAttribute(
            'href',
            'https://git.subcult.tv/subculture-collective/clpr'
        );
    });

    it('links to privacy and terms pages', () => {
        render(
            <MemoryRouter>
                <AboutPage />
            </MemoryRouter>
        );

        const privacyLink = screen.getByText('Privacy Policy');
        expect(privacyLink).toHaveAttribute('href', '/privacy');

        const termsLink = screen.getByText('Terms of Service');
        expect(termsLink).toHaveAttribute('href', '/terms');
    });
});
