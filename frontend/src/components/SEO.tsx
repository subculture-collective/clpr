import { useEffect } from 'react';
import { Helmet } from '@dr.pogodin/react-helmet';

export interface SEOProps {
    title?: string;
    description?: string;
    canonicalUrl?: string;
    ogType?: 'website' | 'article' | 'video.other';
    ogImage?: string;
    ogVideo?: string;
    ogVideoType?: string;
    ogVideoWidth?: string;
    ogVideoHeight?: string;
    twitterCard?: 'summary' | 'summary_large_image' | 'player';
    twitterPlayer?: string;
    twitterPlayerWidth?: string;
    twitterPlayerHeight?: string;
    noindex?: boolean;
    nofollow?: boolean;
    structuredData?: Record<string, unknown>;
}

const DEFAULT_TITLE = 'Clipper - Community-Driven Twitch Clip Curation';
const DEFAULT_DESCRIPTION =
    'Discover, share, and vote on memorable Twitch clips from creators across live culture.';
const DEFAULT_IMAGE = '/clpr-banner-1021px.png';
const SITE_NAME = 'Clipper';

export function SEO({
    title,
    description = DEFAULT_DESCRIPTION,
    canonicalUrl,
    ogType = 'website',
    ogImage = DEFAULT_IMAGE,
    ogVideo,
    ogVideoType,
    ogVideoWidth,
    ogVideoHeight,
    twitterCard = 'summary_large_image',
    twitterPlayer,
    twitterPlayerWidth,
    twitterPlayerHeight,
    noindex = false,
    nofollow = false,
    structuredData,
}: SEOProps) {
    const fullTitle = title ? `${title} | ${SITE_NAME}` : DEFAULT_TITLE;
    const baseUrl = import.meta.env.VITE_BASE_URL || window.location.origin;
    const fullCanonicalUrl =
        canonicalUrl ?
            `${baseUrl}${canonicalUrl}`
        :   window.location.href.split('?')[0];
    const fullOgImage =
        ogImage.startsWith('http') ? ogImage : `${baseUrl}${ogImage}`;

    useEffect(() => {
        // Update meta theme-color based on dark mode if needed
        const metaTheme = document.querySelector('meta[name="theme-color"]');
        if (metaTheme) {
            const isDark = document.documentElement.classList.contains('dark');
            metaTheme.setAttribute('content', isDark ? '#1a1a1a' : '#ffffff');
        }
    }, []);

    const robotsContent = [];
    if (noindex) robotsContent.push('noindex');
    if (nofollow) robotsContent.push('nofollow');
    const robots =
        robotsContent.length > 0 ? robotsContent.join(', ') : undefined;

    return (
        <Helmet>
            {/* Basic Meta Tags */}
            <title>{fullTitle}</title>
            <meta name='description' content={description} />
            {robots && <meta name='robots' content={robots} />}
            <link rel='canonical' href={fullCanonicalUrl} />

            {/* Open Graph Meta Tags */}
            <meta property='og:site_name' content={SITE_NAME} />
            <meta property='og:title' content={fullTitle} />
            <meta property='og:description' content={description} />
            <meta property='og:type' content={ogType} />
            <meta property='og:url' content={fullCanonicalUrl} />
            <meta property='og:image' content={fullOgImage} />
            {ogVideo && <meta property='og:video' content={ogVideo} />}
            {ogVideoType && (
                <meta property='og:video:type' content={ogVideoType} />
            )}
            {ogVideoWidth && (
                <meta property='og:video:width' content={ogVideoWidth} />
            )}
            {ogVideoHeight && (
                <meta property='og:video:height' content={ogVideoHeight} />
            )}

            {/* Twitter Card Meta Tags */}
            <meta name='twitter:card' content={twitterCard} />
            <meta name='twitter:title' content={fullTitle} />
            <meta name='twitter:description' content={description} />
            <meta name='twitter:image' content={fullOgImage} />
            {twitterPlayer && (
                <meta name='twitter:player' content={twitterPlayer} />
            )}
            {twitterPlayerWidth && (
                <meta
                    name='twitter:player:width'
                    content={twitterPlayerWidth}
                />
            )}
            {twitterPlayerHeight && (
                <meta
                    name='twitter:player:height'
                    content={twitterPlayerHeight}
                />
            )}

            {/* Structured Data (JSON-LD) */}
            {structuredData && (
                <script type='application/ld+json'>
                    {JSON.stringify(structuredData)}
                </script>
            )}
        </Helmet>
    );
}
