import { Link, useParams } from 'react-router-dom';
import { Container, SEO } from '../components';
import { ClipFeed } from '../components/clip';
import { useTag } from '../hooks/useTags';
import { TAG_EVIDENCE, TAG_LANES, tagHref, tagLabel, tagLane } from '../lib/tag-lanes';

export function TagPage() {
    const { tagSlug = '' } = useParams<{ tagSlug: string }>();
    const { data } = useTag(tagSlug);

    if (!tagSlug) {
        return (
            <Container className='py-8'>
                <div className='text-center text-muted-foreground'>
                    <p>No tag specified</p>
                </div>
            </Container>
        );
    }

    const tag = data?.tag;
    const lane = tagLane(tag ?? { slug: tagSlug });
    const label = tag ? tagLabel(tag) : tagSlug;
    const evidence = lane === 'detected' ? tag?.evidence : undefined;

    return (
        <>
            <SEO
                title={label}
                description={`Twitch clips tagged ${label} on clpr.`}
                canonicalUrl={tagHref(tagSlug)}
            />
            <Container className='py-8'>
                <nav aria-label='Breadcrumb' className='kicker mb-4'>
                    <Link to='/tags' className='text-text-tertiary hover:text-foreground'>Tags</Link>
                    <span aria-hidden='true'> / </span>
                    <span>{TAG_LANES[lane].label}</span>
                </nav>
                <div className='mb-8 grid gap-4 border-b border-line-strong pb-6 md:grid-cols-[1fr_auto] md:items-end'>
                    <div>
                        <h1 className='text-4xl lg:text-5xl'>{lane === 'community' ? `#${label}` : label}</h1>
                        <p className='mt-2 max-w-2xl text-text-secondary'>
                            {evidence ? `${TAG_EVIDENCE[evidence].label} tag. ${TAG_EVIDENCE[evidence].description}` : TAG_LANES[lane].description}
                        </p>
                    </div>
                    {data && (
                        <p className='font-mono text-sm text-text-secondary'>
                            <span className='display text-3xl text-foreground tabular-nums'>{data.clip_count.toLocaleString()}</span> clips
                        </p>
                    )}
                </div>
                <ClipFeed
                    title='Clips'
                    headingLevel='h2'
                    filters={{ tags: [tagSlug] }}
                    showSearch={false}
                    useSortTitle={false}
                />
            </Container>
        </>
    );
}
