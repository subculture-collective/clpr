import { Helmet } from '@dr.pogodin/react-helmet';

interface MetaProps {
    title?: string;
    description?: string;
    image?: string;
    url?: string;
    type?: string;
}

const DEFAULT_TITLE = 'clpr - Discover Creators and Live Moments';
const DEFAULT_DESCRIPTION =
    'Discover the creators and moments shaping live culture. Browse Twitch clips by creator, topic, tag, or collection.';
const DEFAULT_IMAGE = '/clpr-banner-1021px.png';

export function Meta({
    title = DEFAULT_TITLE,
    description = DEFAULT_DESCRIPTION,
    image = DEFAULT_IMAGE,
    url,
    type = 'website',
}: MetaProps) {
    const fullTitle = title === DEFAULT_TITLE ? title : `${title} | clpr`;
    const fullUrl = url || window.location.href;

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
            <meta property='og:image' content={image} />

            {/* Twitter */}
            <meta property='twitter:card' content='summary_large_image' />
            <meta property='twitter:url' content={fullUrl} />
            <meta property='twitter:title' content={fullTitle} />
            <meta property='twitter:description' content={description} />
            <meta property='twitter:image' content={image} />
        </Helmet>
    );
}
