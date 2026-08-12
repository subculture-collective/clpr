import { useEffect } from 'react';
import { Helmet } from '@dr.pogodin/react-helmet';

export interface SEOProps {
    title?: string;
    description?: string;
    canonicalUrl?: string;
    ogType?: 'website' | 'article' | 'video.other';
    ogImage?: string;
    imageAlt?: string;
    imageWidth?: string;
    imageHeight?: string;
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

const DEFAULT_TITLE = 'clpr - Discover Creators and Live Moments';
const DEFAULT_DESCRIPTION =
    'Discover the creators and moments shaping live culture. Browse Twitch clips by creator, topic, tag, or collection.';
const DEFAULT_IMAGE = '/social-card.png';
const DEFAULT_IMAGE_ALT = 'clpr — Discover creators and live moments';
const SITE_NAME = 'clpr';
const TWITTER_HANDLE = '@clpr_tv';

export function SEO({
    title,
    description = DEFAULT_DESCRIPTION,
    canonicalUrl,
    ogType = 'website',
    ogImage = DEFAULT_IMAGE,
    imageAlt = DEFAULT_IMAGE_ALT,
    imageWidth,
    imageHeight,
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
    const resolvedImageWidth =
        imageWidth ?? (ogImage === DEFAULT_IMAGE ? '1200' : undefined);
    const resolvedImageHeight =
        imageHeight ?? (ogImage === DEFAULT_IMAGE ? '630' : undefined);

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
            <meta property='og:locale' content='en_US' />
            <meta property='og:title' content={fullTitle} />
            <meta property='og:description' content={description} />
            <meta property='og:type' content={ogType} />
            <meta property='og:url' content={fullCanonicalUrl} />
            <meta property='og:image' content={fullOgImage} />
            <meta property='og:image:secure_url' content={fullOgImage} />
            {resolvedImageWidth && (
                <meta property='og:image:width' content={resolvedImageWidth} />
            )}
            {resolvedImageHeight && (
                <meta property='og:image:height' content={resolvedImageHeight} />
            )}
            <meta property='og:image:alt' content={imageAlt} />
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
            <meta name='twitter:site' content={TWITTER_HANDLE} />
            <meta name='twitter:creator' content={TWITTER_HANDLE} />
            <meta name='twitter:domain' content='clpr.tv' />
            <meta name='twitter:url' content={fullCanonicalUrl} />
            <meta name='twitter:title' content={fullTitle} />
            <meta name='twitter:description' content={description} />
            <meta name='twitter:image' content={fullOgImage} />
            <meta name='twitter:image:alt' content={imageAlt} />
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
