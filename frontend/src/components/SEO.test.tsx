import { render, waitFor } from '@testing-library/react';
import { HelmetProvider } from '@dr.pogodin/react-helmet';
import { describe, expect, it } from 'vitest';
import { SEO } from './SEO';

const meta = (selector: string) =>
    document.head.querySelector<HTMLMetaElement>(selector)?.content;

describe('SEO social metadata', () => {
    it('publishes the branded default card and clpr social account', async () => {
        render(
            <HelmetProvider>
                <SEO canonicalUrl='/creators' />
            </HelmetProvider>,
        );

        await waitFor(() => {
            expect(meta('meta[name="twitter:site"]')).toBe('@clpr_tv');
        });

        expect(meta('meta[property="og:site_name"]')).toBe('clpr');
        expect(meta('meta[property="og:image"]')).toMatch(
            /\/social-card\.png$/,
        );
        expect(meta('meta[property="og:image:width"]')).toBe('1200');
        expect(meta('meta[property="og:image:height"]')).toBe('630');
        expect(meta('meta[name="twitter:card"]')).toBe(
            'summary_large_image',
        );
        expect(meta('meta[name="twitter:creator"]')).toBe('@clpr_tv');
        expect(meta('meta[name="twitter:image:alt"]')).toContain('clpr');
    });

    it('does not claim default dimensions for a custom clip thumbnail', async () => {
        render(
            <HelmetProvider>
                <SEO
                    title='A memorable clip'
                    ogImage='https://static-cdn.jtvnw.net/clip.jpg'
                    imageAlt='A memorable clip by a creator'
                />
            </HelmetProvider>,
        );

        await waitFor(() => {
            expect(meta('meta[property="og:image"]')).toBe(
                'https://static-cdn.jtvnw.net/clip.jpg',
            );
        });

        expect(meta('meta[property="og:image:width"]')).toBeUndefined();
        expect(meta('meta[property="og:image:height"]')).toBeUndefined();
        expect(meta('meta[property="og:image:alt"]')).toBe(
            'A memorable clip by a creator',
        );
    });
});
