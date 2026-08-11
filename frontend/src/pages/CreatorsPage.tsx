import { useQuery } from '@tanstack/react-query';
import { Link } from 'react-router-dom';
import { ArrowUpRight, Radio, Users } from 'lucide-react';
import { SEO } from '@/components/SEO';
import { CreatorDiscoveryRail } from '@/components/broadcaster/CreatorDiscoveryRail';
import {
    fetchCreatorDiscovery,
    fetchLiveBroadcasters,
} from '@/lib/broadcaster-api';

export function CreatorsPage() {
    const discoveryQuery = useQuery({
        queryKey: ['creator-discovery', 12],
        queryFn: () => fetchCreatorDiscovery(12),
        staleTime: 5 * 60 * 1000,
    });
    const liveQuery = useQuery({
        queryKey: ['live-creators', 12],
        queryFn: () => fetchLiveBroadcasters(1, 12),
        refetchInterval: 60_000,
    });

    const rails = discoveryQuery.data;
    const liveCreators = liveQuery.data?.data ?? [];
    const isEmpty =
        rails &&
        rails.trending.length === 0 &&
        rails.rising.length === 0 &&
        rails.new.length === 0 &&
        liveCreators.length === 0;

    return (
        <>
            <SEO
                title='Discover Creators'
                description='Find trending, rising, live, and newly discovered creators shaping live culture.'
                canonicalUrl='/creators'
            />
            <main className='container mx-auto px-4 py-8 sm:py-12'>
                <header className='relative mb-12 overflow-hidden border-y border-border py-10 sm:py-14'>
                    <div className='absolute right-0 top-0 h-full w-1/2 bg-[radial-gradient(circle_at_center,rgb(var(--color-brand)/0.13),transparent_65%)]' />
                    <div className='relative max-w-4xl'>
                        <p className='mb-4 text-[11px] font-semibold uppercase tracking-[0.24em] text-brand'>
                            Creator index / live culture
                        </p>
                        <h1 className='font-heading text-4xl font-bold leading-[1.05] text-foreground sm:text-6xl'>
                            Follow the people,
                            <br />
                            <span className='text-text-secondary'>not just the category.</span>
                        </h1>
                        <p className='mt-5 max-w-2xl text-base leading-relaxed text-muted-foreground sm:text-lg'>
                            Meet the creators driving today&apos;s conversations, performances,
                            reactions, adventures, and unforgettable live moments.
                        </p>
                    </div>
                </header>

                {discoveryQuery.isLoading && (
                    <div className='grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-4' aria-label='Loading creators'>
                        {Array.from({ length: 8 }, (_, index) => (
                            <div key={index} className='aspect-[4/3] animate-pulse rounded-xl bg-surface-raised' />
                        ))}
                    </div>
                )}

                {discoveryQuery.isError && (
                    <div className='rounded-xl border border-error-500/30 bg-error-500/5 p-6 text-sm text-error-600'>
                        Creator discovery is temporarily unavailable. Please try again shortly.
                    </div>
                )}

                {!discoveryQuery.isLoading && !discoveryQuery.isError && (
                    <div className='space-y-14'>
                        <CreatorDiscoveryRail
                            title='Creators moving now'
                            description='Current view velocity, fresh clips, and community response—weighted toward what is happening now.'
                            tone='trending'
                            creators={rails?.trending ?? []}
                        />
                        <CreatorDiscoveryRail
                            title='On the rise'
                            description='Momentum relative to catalog size, giving smaller and emerging creators room to break through.'
                            tone='rising'
                            creators={rails?.rising ?? []}
                        />

                        <section aria-labelledby='live-creators-heading'>
                            <div className='mb-4 flex items-center gap-2'>
                                <span className='relative flex h-2 w-2'>
                                    <span className='absolute inline-flex h-full w-full animate-ping rounded-full bg-red-500 opacity-70' />
                                    <span className='relative inline-flex h-2 w-2 rounded-full bg-red-600' />
                                </span>
                                <span className='text-[11px] font-semibold uppercase tracking-[0.18em] text-text-secondary'>Live</span>
                            </div>
                            <h2 id='live-creators-heading' className='font-heading text-2xl font-bold'>Live right now</h2>
                            <p className='mt-1 text-sm text-muted-foreground'>Creators currently broadcasting, ordered by the live audience gathered around them.</p>
                            {liveCreators.length > 0 ? (
                                <div className='mt-4 flex snap-x gap-4 overflow-x-auto pb-3 scrollbar-hide'>
                                    {liveCreators.map(creator => (
                                        <Link
                                            key={creator.broadcaster_id}
                                            to={`/broadcaster/${creator.broadcaster_id}`}
                                            className='group flex w-72 shrink-0 snap-start items-center gap-4 rounded-xl border border-border bg-surface p-4 hover:border-red-500/50'
                                        >
                                            <div className='flex h-12 w-12 shrink-0 items-center justify-center rounded-full bg-red-500/10 text-red-500'>
                                                <Radio className='h-5 w-5' />
                                            </div>
                                            <div className='min-w-0 flex-1'>
                                                <h3 className='truncate font-heading font-bold'>{creator.user_name || creator.user_login || creator.broadcaster_id}</h3>
                                                <p className='truncate text-xs text-muted-foreground'>{creator.stream_title || 'Live on Twitch'}</p>
                                                <p className='mt-1 text-xs font-medium text-red-500'>{creator.viewer_count.toLocaleString()} watching</p>
                                            </div>
                                            <ArrowUpRight className='h-4 w-4 text-text-tertiary group-hover:text-red-500' />
                                        </Link>
                                    ))}
                                </div>
                            ) : (
                                <div className='mt-4 flex items-center gap-3 rounded-xl border border-dashed border-border px-5 py-4 text-sm text-muted-foreground'>
                                    <Radio className='h-4 w-4' />
                                    Live tracking is quiet right now. Check back as creators go live.
                                </div>
                            )}
                        </section>

                        <CreatorDiscoveryRail
                            title='Fresh faces'
                            description='Creators first discovered by Clpr in the last seven days, ordered by recency.'
                            tone='new'
                            creators={rails?.new ?? []}
                        />
                    </div>
                )}

                {isEmpty && (
                    <div className='flex flex-col items-center py-16 text-center'>
                        <Users className='mb-4 h-10 w-10 text-text-tertiary' />
                        <h2 className='font-heading text-xl font-bold'>The creator index is warming up</h2>
                        <p className='mt-2 max-w-md text-sm text-muted-foreground'>Fresh creator signals will appear as new clips arrive and audiences respond.</p>
                    </div>
                )}
            </main>
        </>
    );
}
