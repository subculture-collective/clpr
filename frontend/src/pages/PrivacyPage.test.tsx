import { render, screen } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';
import { describe, expect, it, vi } from 'vitest';
import { PrivacyPage } from './PrivacyPage';

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

describe('PrivacyPage', () => {
    it('renders the page title', () => {
        render(
            <MemoryRouter>
                <PrivacyPage />
            </MemoryRouter>
        );

        expect(screen.getByText('Privacy Policy')).toBeInTheDocument();
    });

    it('displays the last updated date', () => {
        render(
            <MemoryRouter>
                <PrivacyPage />
            </MemoryRouter>
        );

        expect(screen.getByText(/Last updated:/i)).toBeInTheDocument();
    });

    it('renders key privacy sections', () => {
        render(
            <MemoryRouter>
                <PrivacyPage />
            </MemoryRouter>
        );

        expect(screen.getByText('Information We Collect')).toBeInTheDocument();
        expect(
            screen.getByText('How We Use Your Information')
        ).toBeInTheDocument();
        expect(
            screen.getByText('Cookies and Tracking Technologies')
        ).toBeInTheDocument();
        expect(
            screen.getByText('How We Share Your Information')
        ).toBeInTheDocument();
        expect(screen.getByText('Data Security')).toBeInTheDocument();
        expect(screen.getByText('Your Privacy Rights')).toBeInTheDocument();
    });

    it('has anchor IDs for navigation', () => {
        const { container } = render(
            <MemoryRouter>
                <PrivacyPage />
            </MemoryRouter>
        );

        expect(
            container.querySelector('#information-we-collect')
        ).toBeInTheDocument();
        expect(
            container.querySelector('#how-we-use-information')
        ).toBeInTheDocument();
        expect(
            container.querySelector('#cookies-tracking')
        ).toBeInTheDocument();
        expect(container.querySelector('#data-sharing')).toBeInTheDocument();
        expect(container.querySelector('#data-security')).toBeInTheDocument();
        expect(container.querySelector('#your-rights')).toBeInTheDocument();
    });

    it('mentions GDPR-relevant information', () => {
        render(
            <MemoryRouter>
                <PrivacyPage />
            </MemoryRouter>
        );

        // Check for user rights which are a key part of GDPR
        expect(screen.getByText(/Access:/i)).toBeInTheDocument();
        expect(screen.getByText(/Deletion:/i)).toBeInTheDocument();
        expect(screen.getByText(/Portability:/i)).toBeInTheDocument();
    });

    it('includes contact information', () => {
        render(
            <MemoryRouter>
                <PrivacyPage />
            </MemoryRouter>
        );

        expect(screen.getByText('Contact Us')).toBeInTheDocument();
        expect(screen.getByText(/privacy@clpr.com/i)).toBeInTheDocument();
    });

    it('mentions third-party services', () => {
        render(
            <MemoryRouter>
                <PrivacyPage />
            </MemoryRouter>
        );

        expect(screen.getByText('Third-Party Services')).toBeInTheDocument();
        const twitchElements = screen.getAllByText(/Twitch/i);
        expect(twitchElements.length).toBeGreaterThan(0);
    });
});
