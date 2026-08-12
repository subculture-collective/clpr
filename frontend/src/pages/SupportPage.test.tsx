import { HelmetProvider } from '@dr.pogodin/react-helmet';
import { render, screen } from '@testing-library/react';
import { describe, expect, it } from 'vitest';
import SupportPage from './SupportPage';

describe('SupportPage', () => {
    it('states that access is free and sends optional support to Patreon', () => {
        render(<HelmetProvider><SupportPage /></HelmetProvider>);

        expect(screen.getByRole('heading', { name: /clpr is for the culture/i })).toBeInTheDocument();
        expect(screen.getByText('No account tiers')).toBeInTheDocument();
        expect(screen.getByText('No feature paywalls')).toBeInTheDocument();
        expect(screen.getByRole('link', { name: /support subcult on patreon/i })).toHaveAttribute(
            'href',
            'https://patreon.com/subcult',
        );
        expect(screen.queryByText(/upgrade to pro/i)).not.toBeInTheDocument();
    });
});
