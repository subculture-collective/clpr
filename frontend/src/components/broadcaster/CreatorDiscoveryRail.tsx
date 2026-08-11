import { Link } from 'react-router-dom';
import { ArrowUpRight, Eye, Film } from 'lucide-react';
import type { CreatorDiscoveryProfile } from '@/lib/broadcaster-api';

type RailTone = 'trending' | 'rising' | 'new';

const toneStyles: Record<RailTone, { label: string; accent: string }> = {
    trending: { label: 'Trending', accent: 'bg-orange-500' },
    rising: { label: 'Rising', accent: 'bg-emerald-500' },
    new: { label: 'New here', accent: 'bg-cyan-500' },
};

function compactNumber(value: number): string {
    return new Intl.NumberFormat('en', {
        notation: 'compact',
        maximumFractionDigits: 1,
    }).format(value);
}

export function CreatorDiscoveryRail({
    title,
    description,
    tone,
    creators,
}: {
    title: string;
    description: string;
    tone: RailTone;
    creators: CreatorDiscoveryProfile[];
}) {
    if (creators.length === 0) return null;

    const toneStyle = toneStyles[tone];
    return (
        <section aria-labelledby={`${tone}-creators-heading`}>
            <div className='mb-4 flex items-end justify-between gap-4'>
                <div>
                    <div className='mb-2 flex items-center gap-2'>
                        <span className={`h-2 w-2 rounded-full ${toneStyle.accent}`} />
                        <span className='text-[11px] font-semibold uppercase tracking-[0.18em] text-text-secondary'>
                            {toneStyle.label}
                        </span>
                    </div>
                    <h2
                        id={`${tone}-creators-heading`}
                        className='font-heading text-2xl font-bold text-foreground'
                    >
                        {title}
                    </h2>
                    <p className='mt-1 max-w-2xl text-sm text-muted-foreground'>
                        {description}
                    </p>
                </div>
            </div>

            <div className='flex snap-x snap-mandatory gap-4 overflow-x-auto pb-3 scrollbar-hide'>
                {creators.map((creator, index) => (
                    <Link
                        key={creator.broadcaster_id}
                        to={`/broadcaster/${creator.broadcaster_id}`}
                        className='group relative w-[252px] shrink-0 snap-start overflow-hidden rounded-xl border border-border bg-surface transition-transform duration-200 hover:-translate-y-1 hover:border-brand/50'
                    >
                        <div className='relative aspect-[4/3] overflow-hidden bg-surface-raised'>
                            {creator.latest_clip_thumbnail ? (
                                <img
                                    src={creator.latest_clip_thumbnail}
                                    alt=''
                                    className='h-full w-full object-cover transition-transform duration-500 group-hover:scale-105'
                                    loading='lazy'
                                />
                            ) : (
                                <div className='flex h-full items-center justify-center bg-[radial-gradient(circle_at_top_left,rgb(var(--color-brand)/0.28),transparent_55%)] font-heading text-6xl font-bold text-brand/40'>
                                    {creator.broadcaster_name.slice(0, 1).toUpperCase()}
                                </div>
                            )}
                            <div className='absolute inset-0 bg-linear-to-t from-black/90 via-black/15 to-transparent' />
                            <span className='absolute left-3 top-3 rounded-full bg-black/65 px-2 py-1 text-[11px] font-semibold text-white backdrop-blur-sm'>
                                {String(index + 1).padStart(2, '0')}
                            </span>
                            <ArrowUpRight className='absolute right-3 top-3 h-4 w-4 text-white opacity-0 transition-opacity group-hover:opacity-100' />
                            <div className='absolute inset-x-0 bottom-0 p-4 text-white'>
                                <h3 className='truncate font-heading text-xl font-bold'>
                                    {creator.broadcaster_name}
                                </h3>
                                {creator.latest_clip_title && (
                                    <p className='mt-1 line-clamp-1 text-xs text-white/75'>
                                        {creator.latest_clip_title}
                                    </p>
                                )}
                            </div>
                        </div>
                        <div className='flex items-center gap-4 px-4 py-3 text-xs text-text-secondary'>
                            <span className='flex items-center gap-1.5'>
                                <Eye className='h-3.5 w-3.5' />
                                {compactNumber(creator.recent_views)} recent
                            </span>
                            <span className='flex items-center gap-1.5'>
                                <Film className='h-3.5 w-3.5' />
                                {creator.total_clips}
                            </span>
                            {creator.twitch_category_name && (
                                <span className='ml-auto max-w-24 truncate' title={creator.twitch_category_name}>
                                    {creator.twitch_category_name}
                                </span>
                            )}
                        </div>
                    </Link>
                ))}
            </div>
        </section>
    );
}
