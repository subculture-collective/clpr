import { useQuery } from '@tanstack/react-query';
import { useState } from 'react';
import { Container, SEO } from '../components';
import { TagChip } from '../components/tag/TagChip';
import { tagApi } from '../lib/tag-api';
import { TAG_EVIDENCE, TAG_LANES, sortByLane } from '../lib/tag-lanes';
import type { Tag, TagEvidence, TagLane } from '../types/tag';

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
    if (tags.length === 0) return <p className='text-sm text-text-secondary'>{empty}</p>;
    return (
        <div className='flex flex-wrap content-start items-start gap-1.5'>
            {tags.map(tag => <TagChip key={tag.id} tag={tag} showCount />)}
        </div>
    );
}

function LaneSection({ lane, title, children }: { lane: TagLane; title?: string; children: React.ReactNode }) {
    return (
        <section className='border-t border-line-strong pt-4' aria-labelledby={`lane-${lane}`}>
            <div className='mb-3 flex flex-wrap items-baseline justify-between gap-x-6 gap-y-1'>
                <h2 id={`lane-${lane}`} className='text-2xl'>{title ?? TAG_LANES[lane].label}</h2>
                <p className='max-w-xl text-sm text-text-secondary'>{TAG_LANES[lane].description}</p>
            </div>
            {children}
        </section>
    );
}

const EVIDENCE_ORDER: TagEvidence[] = ['visible', 'contextual', 'strong'];

function useLane(lane: TagLane, limit: number) {
    return useQuery({
        queryKey: ['tags', 'lane', lane, limit],
        queryFn: () => tagApi.listTags({ lane, sort: 'popularity', limit }),
    });
}

export function TagsPage() {
    const [search, setSearch] = useState('');
    const searching = search.trim().length >= 2;
    const detected = useLane('detected', 100);
    const category = useLane('category', 30);
    const community = useLane('community', 30);
    const streamer = useLane('streamer', 30);
    const trending = useQuery({ queryKey: ['tags', 'trending', 16], queryFn: () => tagApi.listTags({ sort: 'trending', limit: 16 }) });
    const searchResults = useQuery({
        queryKey: ['tags', 'search', search],
        queryFn: () => tagApi.searchTags(search, 40),
        enabled: searching,
    });

    const detectedTags = detected.data?.tags ?? [];

    return (
        <>
            <SEO title='Tags' description='Browse Twitch clips by what is in them, their Twitch category, and community tags.' canonicalUrl='/tags' />
            <Container className='py-8'>
                <p className='kicker mb-2'>Browse</p>
                <h1 className='mb-3 text-4xl'>Tags</h1>
                <p className='mb-6 max-w-2xl text-text-secondary'>
                    Every tag on clpr shows where it came from: what clpr saw in the clip, the Twitch category it streamed under, what people here added, or the streamer’s own channel tags.
                </p>

                <dl className='mb-8 grid gap-px border border-border bg-border sm:grid-cols-2 lg:grid-cols-4'>
                    {(['detected', 'category', 'community', 'streamer'] as TagLane[]).map(lane => (
                        <div key={lane} className='bg-background p-3'>
                            <dt className='kicker mb-1'>{TAG_LANES[lane].label}</dt>
                            <dd className='text-sm text-text-secondary'>{TAG_LANES[lane].description}</dd>
                        </div>
                    ))}
                </dl>

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
                        <TagRow tags={sortByLane(searchResults.data?.tags ?? [])} query={searchResults} empty={`No tags match “${search.trim()}”.`} />
                    </section>
                ) : (
                    <div className='space-y-10'>
                        <LaneSection lane='detected'>
                            {detected.isLoading || detected.isError || detectedTags.length === 0 ? (
                                <TagRow tags={detectedTags} query={detected} empty='No content tags yet.' />
                            ) : (
                                <div className='space-y-4'>
                                    {EVIDENCE_ORDER.map(evidence => {
                                        const tags = detectedTags.filter(tag => tag.evidence === evidence);
                                        if (tags.length === 0) return null;
                                        return (
                                            <div key={evidence} className='grid gap-2 md:grid-cols-[10rem_1fr]'>
                                                <div>
                                                    <p className='kicker text-foreground'>{TAG_EVIDENCE[evidence].label}</p>
                                                    <p className='text-xs text-text-tertiary'>{TAG_EVIDENCE[evidence].description}</p>
                                                </div>
                                                <TagRow tags={tags} query={detected} empty='' />
                                            </div>
                                        );
                                    })}
                                </div>
                            )}
                        </LaneSection>

                        <section className='border-t border-line-strong pt-4' aria-labelledby='lane-trending'>
                            <div className='mb-3 flex flex-wrap items-baseline justify-between gap-x-6 gap-y-1'>
                                <h2 id='lane-trending' className='text-2xl'>Trending this week</h2>
                                <p className='text-sm text-text-secondary'>Most added to clips in the last seven days, across every lane.</p>
                            </div>
                            <TagRow tags={trending.data?.tags ?? []} query={trending} empty='Nothing is trending yet this week.' />
                        </section>

                        <LaneSection lane='category' title='Twitch categories'>
                            <TagRow tags={category.data?.tags ?? []} query={category} empty='No categories yet.' />
                        </LaneSection>

                        <LaneSection lane='community'>
                            <TagRow tags={community.data?.tags ?? []} query={community} empty='No community tags yet.' />
                        </LaneSection>

                        <LaneSection lane='streamer'>
                            <TagRow tags={streamer.data?.tags ?? []} query={streamer} empty='No streamer tags yet.' />
                        </LaneSection>
                    </div>
                )}
            </Container>
        </>
    );
}
