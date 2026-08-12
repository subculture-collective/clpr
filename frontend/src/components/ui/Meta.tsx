import { Helmet } from '@dr.pogodin/react-helmet';

interface MetaProps {
    title?: string;
    description?: string;
    image?: string;
    imageAlt?: string;
    url?: string;
    type?: string;
}

const DEFAULT_TITLE = 'clpr - Discover Creators and Live Moments';
const DEFAULT_DESCRIPTION =
    'Discover the creators and moments shaping live culture. Browse Twitch clips by creator, topic, tag, or collection.';
const DEFAULT_IMAGE = '/social-card.png';
const DEFAULT_IMAGE_ALT = 'clpr — Discover creators and live moments';
const TWITTER_HANDLE = '@clpr_tv';

export function Meta({
    title = DEFAULT_TITLE,
    description = DEFAULT_DESCRIPTION,
    image = DEFAULT_IMAGE,
    imageAlt = DEFAULT_IMAGE_ALT,
    url,
    type = 'website',
}: MetaProps) {
    const fullTitle = title === DEFAULT_TITLE ? title : `${title} | clpr`;
    const fullUrl = url || window.location.href;
    const baseUrl = import.meta.env.VITE_BASE_URL || window.location.origin;
    const fullImage = image.startsWith('http') ? image : `${baseUrl}${image}`;

    return (
        <Helmet>
            {/* Primary Meta Tags */}
            <title>{fullTitle}</title>
            <meta name='title' content={fullTitle} />
            <meta name='description' content={description} />

            {/* Open Graph / Facebook */}
            <meta property='og:type' content={type} />
            <meta property='og:url' content={fullUrl} />
            <meta property='og:title' content={fullTitle} />
            <meta property='og:description' content={description} />
            <meta property='og:site_name' content='clpr' />
            <meta property='og:locale' content='en_US' />
            <meta property='og:image' content={fullImage} />
            <meta property='og:image:secure_url' content={fullImage} />
            <meta property='og:image:width' content='1200' />
            <meta property='og:image:height' content='630' />
            <meta property='og:image:alt' content={imageAlt} />

            {/* Twitter */}
            <meta name='twitter:card' content='summary_large_image' />
            <meta name='twitter:site' content={TWITTER_HANDLE} />
            <meta name='twitter:creator' content={TWITTER_HANDLE} />
            <meta name='twitter:domain' content='clpr.tv' />
            <meta name='twitter:url' content={fullUrl} />
            <meta name='twitter:title' content={fullTitle} />
            <meta name='twitter:description' content={description} />
            <meta name='twitter:image' content={fullImage} />
            <meta name='twitter:image:alt' content={imageAlt} />
        </Helmet>
    );
}
