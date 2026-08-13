import { SEO } from '../components';
import { ClipFeed } from '../components/clip';
import { PlaylistCard } from '../components/playlist/PlaylistCard';
import { FeedLayout } from '../components/layout/FeedLayout';
import { FeedSidebar } from '../components/layout/FeedSidebar';
import { useFeaturedPlaylists } from '../hooks/usePlaylist';
import { ChevronLeft, ChevronRight } from 'lucide-react';
import { Link } from 'react-router-dom';
import { useCallback, useEffect, useRef, useState } from 'react';

export function HomePage() {
    const carouselRef = useRef<HTMLDivElement>(null);
    const [canScrollLeft, setCanScrollLeft] = useState(false);
    const [canScrollRight, setCanScrollRight] = useState(false);
    const scrollCarousel = useCallback((direction: 'left' | 'right') => {
        const container = carouselRef.current;
        if (!container) return;
        const scrollAmount = Math.max(240, container.clientWidth * 0.9);
        container.scrollBy({
            left: direction === 'left' ? -scrollAmount : scrollAmount,
            behavior: 'smooth',
        });
    }, []);

    const updateCarouselControls = useCallback(() => {
        const container = carouselRef.current;
        if (!container) return;
        const maxScrollLeft = container.scrollWidth - container.clientWidth;
        const hasOverflow = maxScrollLeft > 4;
        setCanScrollLeft(hasOverflow && container.scrollLeft > 4);
        setCanScrollRight(
            hasOverflow && container.scrollLeft < maxScrollLeft - 4,
        );
    }, []);

    const { data: featuredPlaylistsResponse, isLoading } = useFeaturedPlaylists(
        1,
        8,
    );
    const featuredPlaylists = (featuredPlaylistsResponse?.data ?? []).filter(
        playlist => playlist.clip_count > 0,
    );

    useEffect(() => {
        const container = carouselRef.current;
        if (!container) return;

        const handle = () => updateCarouselControls();
        const initialFrame = requestAnimationFrame(handle);
        container.addEventListener('scroll', handle, { passive: true });
        const resizeObserver = new ResizeObserver(handle);
        resizeObserver.observe(container);

        return () => {
            cancelAnimationFrame(initialFrame);
            container.removeEventListener('scroll', handle);
            resizeObserver.disconnect();
        };
    }, [featuredPlaylists.length, updateCarouselControls]);

    const collectionsModule = isLoading ? (
        <section className='min-h-64 animate-pulse rounded-2xl border border-border bg-card p-4' aria-label='Loading curated collections'>
            <div className='mb-4 h-7 w-52 rounded bg-surface-raised' />
            <div className='h-44 rounded-xl bg-surface-raised' />
        </section>
    ) : featuredPlaylists.length > 0 ? (
        <section className='overflow-hidden' aria-labelledby='curated-collections-title'>
            <div className='flex items-end justify-between gap-4 mb-4 px-4 md:px-0'>
                <div>
                    <p className='mb-1 text-xs font-semibold uppercase tracking-[0.18em] text-primary-400'>Keep exploring</p>
                    <h2 id='curated-collections-title' className='text-xl md:text-2xl font-bold text-foreground'>Curated Collections</h2>
                    <p className='text-muted-foreground text-sm mt-1'>Handpicked creator moments, assembled for a longer watch.</p>
                </div>
                <Link to='/discover/lists' className='flex min-h-11 shrink-0 items-center gap-1 text-primary-400 hover:text-primary-300 text-sm font-semibold'>
                    View all <ChevronRight className='w-4 h-4' />
                </Link>
            </div>
            <div className='relative'>
                <div className='absolute inset-y-0 right-0 w-10 bg-linear-to-l from-background to-transparent pointer-events-none z-10' />
                <div ref={carouselRef} className='flex gap-4 md:gap-6 overflow-x-auto scroll-smooth snap-x snap-mandatory pb-2 px-4 md:px-0 scrollbar-hide scrolling-touch touch-pan-x overscroll-x-contain' aria-label='Curated collections carousel'>
                    {featuredPlaylists.map(playlist => (
                        <div key={playlist.id} className='snap-start shrink-0 w-[82vw] max-w-90 lg:w-100'>
                            <PlaylistCard playlist={playlist} />
                        </div>
                    ))}
                </div>
                {canScrollLeft && <div className='absolute top-1/2 -translate-y-1/2 left-2 hidden sm:flex'><button type='button' onClick={() => scrollCarousel('left')} className='h-9 w-9 rounded-full border border-border bg-background/90 shadow-md hover:bg-background flex items-center justify-center' aria-label='Scroll curated collections left'><ChevronLeft className='h-4 w-4' /></button></div>}
                {canScrollRight && <div className='absolute top-1/2 -translate-y-1/2 right-2 hidden sm:flex'><button type='button' onClick={() => scrollCarousel('right')} className='h-9 w-9 rounded-full border border-border bg-background/90 shadow-md hover:bg-background flex items-center justify-center' aria-label='Scroll curated collections right'><ChevronRight className='h-4 w-4' /></button></div>}
            </div>
        </section>
    ) : null;

    return (
        <>
            <SEO
                title='Home'
                description='Discover the creators and live moments people are talking about. Explore fresh clips, follow creators, and join the conversation.'
                canonicalUrl='/'
            />
            <FeedLayout sidebar={<FeedSidebar />}>
                {/* Main Clip Feed */}
                <ClipFeed
                    title='Live culture, clipped'
                    description='Fresh moments from creators across Twitch'
                    defaultSort='trending'
                    showSearch
                    insertAfter={5}
                    insertedContent={collectionsModule}
                />
            </FeedLayout>
        </>
    );
}
