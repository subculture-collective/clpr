import { useQuery } from '@tanstack/react-query';
import { Link } from 'react-router-dom';
import { Container, SEO, Spinner } from '../components';
import { tagApi } from '../lib/tag-api';
import { useState } from 'react';

function TagGroup({ title, tags }: { title: string; tags: Array<{ id: string; slug: string; name: string; usage_count: number }> }) {
    if (tags.length === 0) return null;
    return <section className='mb-8'><h2 className='mb-3 text-xl font-semibold'>{title}</h2><div className='flex flex-wrap gap-3'>{tags.map(tag => (
        <Link key={tag.id} to={`/tags/${encodeURIComponent(tag.slug)}`} className='inline-flex min-h-11 items-center gap-2 rounded-full border border-border bg-surface px-4 py-2 text-sm text-foreground hover:border-brand hover:text-brand focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-primary-400'>
            <span>#{tag.name}</span>{tag.usage_count > 0 && <span className='text-xs text-muted-foreground'>{tag.usage_count.toLocaleString()}</span>}
        </Link>
    ))}</div></section>;
}

export function TagsPage() {
    const [search, setSearch] = useState('');
    const alphabetical = useQuery({
        queryKey: ['tags', 'alphabetical', 100],
        queryFn: () => tagApi.listTags({ sort: 'alphabetical', limit: 100 }),
    });
    const popular = useQuery({ queryKey: ['tags', 'popularity', 24], queryFn: () => tagApi.listTags({ sort: 'popularity', limit: 24 }) });
    const trending = useQuery({ queryKey: ['tags', 'trending', 12], queryFn: () => tagApi.listTags({ sort: 'trending', limit: 12 }) });
    const curated = useQuery({ queryKey: ['tags', 'curated', 24], queryFn: () => tagApi.listTags({ sort: 'curated', limit: 24 }) });
    const searchResults = useQuery({ queryKey: ['tags', 'search', search], queryFn: () => tagApi.searchTags(search, 40), enabled: search.trim().length >= 2 });
    const tags = alphabetical.data?.tags ?? [];
    const isLoading = alphabetical.isLoading || popular.isLoading || trending.isLoading || curated.isLoading;
    const isError = alphabetical.isError || popular.isError || trending.isError || curated.isError;

    return (
        <>
            <SEO title='Tags' description='Browse Twitch clips by community tag.' canonicalUrl='/tags' />
            <Container className='py-8'>
                <h1 className='text-3xl font-bold text-foreground mb-2'>Tags</h1>
                <p className='text-muted-foreground mb-8'>Explore the labels people use to describe clips and moments.</p>
                <label className='mb-8 block max-w-xl'>
                    <span className='sr-only'>Search tags</span>
                    <input value={search} onChange={event => setSearch(event.target.value)} placeholder='Search tags' className='min-h-11 w-full rounded-md border border-border bg-background px-4 text-foreground focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-primary-400' />
                </label>
                {isLoading ? (
                    <div className='flex justify-center py-16'><Spinner size='xl' /></div>
                ) : isError ? (
                    <p className='text-center text-muted-foreground py-16'>Tags could not be loaded.</p>
                ) : tags.length === 0 ? (
                    <p className='text-center text-muted-foreground py-16'>No tags are available yet.</p>
                ) : (
                    <>
                        {search.trim().length >= 2 ? <TagGroup title='Search results' tags={searchResults.data?.tags ?? []} /> : <>
                            <TagGroup title='Popular' tags={popular.data?.tags ?? []} />
                            <TagGroup title='Trending this week' tags={trending.data?.tags ?? []} />
                            <TagGroup title='Curated' tags={curated.data?.tags ?? []} />
                        </>}
                        <details className='rounded-lg border border-border p-4'><summary className='min-h-11 cursor-pointer font-semibold'>Alphabetical directory</summary><div className='mt-4'><TagGroup title='All tags' tags={tags} /></div></details>
                    </>
                )}
            </Container>
        </>
    );
}
