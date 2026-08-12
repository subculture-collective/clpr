import { useQuery } from '@tanstack/react-query';
import { Link } from 'react-router-dom';
import { Container, SEO, Spinner } from '../components';
import { tagApi } from '../lib/tag-api';

export function TagsPage() {
    const { data, isLoading, isError } = useQuery({
        queryKey: ['tags', 'alphabetical', 100],
        queryFn: () => tagApi.listTags({ sort: 'alphabetical', limit: 100 }),
    });
    const tags = data?.tags ?? [];

    return (
        <>
            <SEO title='Tags' description='Browse Twitch clips by community tag.' canonicalUrl='/tags' />
            <Container className='py-8'>
                <h1 className='text-3xl font-bold text-foreground mb-2'>Tags</h1>
                <p className='text-muted-foreground mb-8'>Explore the labels people use to describe clips and moments.</p>
                {isLoading ? (
                    <div className='flex justify-center py-16'><Spinner size='xl' /></div>
                ) : isError ? (
                    <p className='text-center text-muted-foreground py-16'>Tags could not be loaded.</p>
                ) : tags.length === 0 ? (
                    <p className='text-center text-muted-foreground py-16'>No tags are available yet.</p>
                ) : (
                    <div className='flex flex-wrap gap-3'>
                        {tags.map(tag => (
                            <Link key={tag.id} to={`/tags/${encodeURIComponent(tag.slug)}`} className='inline-flex items-center gap-2 rounded-full border border-border bg-surface px-4 py-2 text-sm text-foreground hover:border-brand hover:text-brand transition-colors'>
                                <span>#{tag.name}</span>
                                {tag.usage_count > 0 && <span className='text-xs text-muted-foreground'>{tag.usage_count.toLocaleString()}</span>}
                            </Link>
                        ))}
                    </div>
                )}
            </Container>
        </>
    );
}
