import { useQuery } from '@tanstack/react-query';
import { useState } from 'react';
import { Container, SEO } from '../components';
import { TagChip } from '../components/tag/TagChip';
import { tagApi } from '../lib/tag-api';
import { isChipLane, tagLane } from '../lib/tag-lanes';
import type { Tag } from '../types/tag';

type TagQuery = { data?: { tags: Tag[] }; isLoading: boolean; isError: boolean };

function TagRow({ tags, query, empty }: { tags: Tag[]; query: TagQuery; empty: string }) {
    if (query.isLoading) {
        return (
            <div className='flex flex-wrap gap-1.5' aria-hidden='true'>
                {[...Array(6)].map((_, i) => <div key={i} className='h-8 w-24 bg-surface-raised animate-pulse motion-reduce:animate-none' />)}
            </div>
        );
    }
    if (query.isError) return <p className='text-sm text-text-secondary'>These tags could not be loaded.</p>;
    const visibleTags = tags.filter(tag => isChipLane(tagLane(tag)));
    if (visibleTags.length === 0) return <p className='text-sm text-text-secondary'>{empty}</p>;
    return (
        <div className='flex flex-wrap content-start items-start gap-1.5'>
            {visibleTags.map(tag => <TagChip key={tag.id} tag={tag} showCount />)}
        </div>
    );
}

export function TagsPage() {
    const [search, setSearch] = useState('');
    const searching = search.trim().length >= 2;
    const popular = useQuery({ queryKey: ['tags', 'popular', 100], queryFn: () => tagApi.listTags({ sort: 'popularity', limit: 100 }) });
    const trending = useQuery({ queryKey: ['tags', 'trending', 16], queryFn: () => tagApi.listTags({ sort: 'trending', limit: 16 }) });
    const searchResults = useQuery({
        queryKey: ['tags', 'search', search],
        queryFn: () => tagApi.searchTags(search, 40),
        enabled: searching,
    });

    return (
        <>
            <SEO title='Tags' description='Find Twitch clips by your favorite games, moments, and interests.' canonicalUrl='/tags' />
            <Container className='py-8'>
                <p className='kicker mb-2'>Browse</p>
                <h1 className='mb-3 text-4xl'>Tags</h1>
                <p className='mb-6 max-w-2xl text-text-secondary'>
                    Find more of what you like. Pick a tag to explore clips.
                </p>

                <label className='mb-10 block max-w-xl'>
                    <span className='kicker mb-1.5 block'>Search tags</span>
                    <input
                        type='search'
                        value={search}
                        onChange={event => setSearch(event.target.value)}
                        placeholder='cooking, just chatting, cozy…'
                        className='min-h-11 w-full border border-line-strong bg-surface px-4 text-foreground placeholder:text-text-tertiary focus:border-primary-400 focus:outline-none focus:ring-1 focus:ring-primary-400'
                    />
                </label>

                {searching ? (
                    <section aria-labelledby='tag-search-results'>
                        <h2 id='tag-search-results' className='mb-3 text-2xl'>Results</h2>
                        <TagRow tags={searchResults.data?.tags ?? []} query={searchResults} empty={`No tags match “${search.trim()}”.`} />
                    </section>
                ) : (
                    <div className='space-y-10'>
                        <section className='border-t border-line-strong pt-4' aria-labelledby='tags-trending'>
                            <div className='mb-3 flex flex-wrap items-baseline justify-between gap-x-6 gap-y-1'>
                                <h2 id='tags-trending' className='text-2xl'>Trending this week</h2>
                                <p className='text-sm text-text-secondary'>Explore tags from the last seven days.</p>
                            </div>
                            <TagRow tags={trending.data?.tags ?? []} query={trending} empty='Nothing is trending yet this week.' />
                        </section>

                        <section className='border-t border-line-strong pt-4' aria-labelledby='tags-popular'>
                            <h2 id='tags-popular' className='mb-3 text-2xl'>Popular tags</h2>
                            <TagRow tags={popular.data?.tags ?? []} query={popular} empty='No tags yet.' />
                        </section>
                    </div>
                )}
            </Container>
        </>
    );
}
