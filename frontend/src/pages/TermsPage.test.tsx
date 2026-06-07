import { render, screen } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';
import { describe, expect, it, vi } from 'vitest';
import { TermsPage } from './TermsPage';

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

describe('TermsPage', () => {
    it('renders the page title', () => {
        render(
            <MemoryRouter>
                <TermsPage />
            </MemoryRouter>
        );

        expect(screen.getByText('Terms of Service')).toBeInTheDocument();
    });

    it('displays the last updated date', () => {
        render(
            <MemoryRouter>
                <TermsPage />
            </MemoryRouter>
        );

        expect(screen.getByText(/Last updated:/i)).toBeInTheDocument();
    });

    it('renders key terms sections', () => {
        render(
            <MemoryRouter>
                <TermsPage />
            </MemoryRouter>
        );

        expect(screen.getByText(/1\. Eligibility/i)).toBeInTheDocument();
        expect(
            screen.getByText(/2\. Account Registration and Security/i)
        ).toBeInTheDocument();
        expect(
            screen.getByText(/3\. Acceptable Use Policy/i)
        ).toBeInTheDocument();
        expect(
            screen.getByText(/4\. Content and Intellectual Property/i)
        ).toBeInTheDocument();
        expect(
            screen.getByText(/8\. Disclaimers and Limitation of Liability/i)
        ).toBeInTheDocument();
    });

    it('has anchor IDs for navigation', () => {
        const { container } = render(
            <MemoryRouter>
                <TermsPage />
            </MemoryRouter>
        );

        expect(container.querySelector('#eligibility')).toBeInTheDocument();
        expect(
            container.querySelector('#account-registration')
        ).toBeInTheDocument();
        expect(container.querySelector('#acceptable-use')).toBeInTheDocument();
        expect(
            container.querySelector('#content-licensing')
        ).toBeInTheDocument();
        expect(container.querySelector('#disclaimers')).toBeInTheDocument();
        expect(container.querySelector('#contact')).toBeInTheDocument();
    });

    it('links to privacy policy', () => {
        render(
            <MemoryRouter>
                <TermsPage />
            </MemoryRouter>
        );

        const privacyLinks = screen.getAllByText('Privacy Policy');
        expect(privacyLinks.length).toBeGreaterThan(0);
        expect(privacyLinks[0]).toHaveAttribute('href', '/privacy');
    });

    it('links to community rules', () => {
        render(
            <MemoryRouter>
                <TermsPage />
            </MemoryRouter>
        );

        const rulesLinks = screen.getAllByText('Community Rules');
        expect(rulesLinks.length).toBeGreaterThan(0);
        expect(rulesLinks[0]).toHaveAttribute('href', '/community-rules');
    });

    it('includes legal disclaimers', () => {
        render(
            <MemoryRouter>
                <TermsPage />
            </MemoryRouter>
        );

        const asIsElements = screen.getAllByText(/AS IS/i);
        expect(asIsElements.length).toBeGreaterThan(0);
        const liabilityElements = screen.getAllByText(
            /Limitation of Liability/i
        );
        expect(liabilityElements.length).toBeGreaterThan(0);
    });

    it('includes contact information', () => {
        render(
            <MemoryRouter>
                <TermsPage />
            </MemoryRouter>
        );

        expect(
            screen.getByText(/14\. Contact Information/i)
        ).toBeInTheDocument();
        expect(screen.getByText(/legal@clpr.com/i)).toBeInTheDocument();
    });

    it('mentions dispute resolution', () => {
        render(
            <MemoryRouter>
                <TermsPage />
            </MemoryRouter>
        );

        expect(
            screen.getByText(/11\. Dispute Resolution and Arbitration/i)
        ).toBeInTheDocument();
        const arbitrationElements = screen.getAllByText(/Binding Arbitration/i);
        expect(arbitrationElements.length).toBeGreaterThan(0);
    });
});
