import { useEffect, useRef, useState } from 'react';
import { Link } from 'react-router-dom';
import { ChevronLeft, ChevronRight } from 'lucide-react';
import { categoryApi } from '../../lib/category-api';
import { tagApi } from '../../lib/tag-api';
import {
    fetchPopularBroadcasters,
    type PopularBroadcaster,
} from '../../lib/broadcaster-api';
import { CategoryIcon } from '../ui/CategoryIcon';
import type { Category } from '../../types/category';
import type { Tag } from '../../types/tag';

type NavTab = 'categories' | 'tags' | 'streamers';

export function CategoriesNav() {
    const [activeTab, setActiveTab] = useState<NavTab>('categories');
    const [categories, setCategories] = useState<Category[]>([]);
    const [tags, setTags] = useState<Tag[]>([]);
    const [streamers, setStreamers] = useState<PopularBroadcaster[]>([]);
    const [loading, setLoading] = useState(true);
    const scrollRef = useRef<HTMLDivElement>(null);
    const [canScrollLeft, setCanScrollLeft] = useState(false);
    const [canScrollRight, setCanScrollRight] = useState(false);

    useEffect(() => {
        const fetchAll = async () => {
            try {
                const [featuredRes, tagsRes, streamersRes] = await Promise.all([
                    categoryApi.listCategories({ featured: true }),
                    tagApi.listTags({ sort: 'popularity', limit: 20 }),
                    fetchPopularBroadcasters(20),
                ]);

                let cats = featuredRes.categories || [];
                if (cats.length === 0) {
                    const all = await categoryApi.listCategories();
                    cats = all.categories || [];
                }
                setCategories(cats);
                setTags(tagsRes.tags || []);
                setStreamers(streamersRes);
            } catch (err) {
                console.error('Failed to fetch nav data:', err);
            } finally {
                setLoading(false);
            }
        };

        fetchAll();
    }, []);

    useEffect(() => {
        const el = scrollRef.current;
        if (!el) return;

        const updateScrollState = () => {
            const { scrollLeft, scrollWidth, clientWidth } = el;
            setCanScrollLeft(scrollLeft > 0);
            setCanScrollRight(scrollLeft + clientWidth < scrollWidth - 1);
        };

        // Reset scroll position when tab changes
        el.scrollLeft = 0;
        updateScrollState();
        el.addEventListener('scroll', updateScrollState);
        window.addEventListener('resize', updateScrollState);
        return () => {
            el.removeEventListener('scroll', updateScrollState);
            window.removeEventListener('resize', updateScrollState);
        };
    }, [activeTab, categories.length, tags.length, streamers.length]);

    const scrollByAmount = (direction: 'left' | 'right') => {
        const el = scrollRef.current;
        if (!el) return;
        const delta = direction === 'left' ? -200 : 200;
        el.scrollBy({ left: delta, behavior: 'smooth' });
    };

    if (loading) {
        return null;
    }

    // Hide if no data at all
    if (
        categories.length === 0 &&
        tags.length === 0 &&
        streamers.length === 0
    ) {
        return null;
    }

    const tabs: { key: NavTab; label: string; count: number }[] = [
        { key: 'categories', label: 'Categories', count: categories.length },
        { key: 'tags', label: 'Tags', count: tags.length },
        { key: 'streamers', label: 'Streamers', count: streamers.length },
    ];

    // Only show tabs that have data
    const visibleTabs = tabs.filter(t => t.count > 0);

    return (
        <div className='border-b border-border bg-background'>
            <div className='container mx-auto px-4'>
                <div className='relative flex items-center'>
                    {/* Tab selector */}
                    {visibleTabs.length > 1 && (
                        <div className='flex items-center gap-1 pr-3 mr-3 border-r border-border shrink-0'>
                            {visibleTabs.map(tab => (
                                <button
                                    key={tab.key}
                                    type='button'
                                    onClick={() => setActiveTab(tab.key)}
                                    className={`px-2.5 py-1.5 text-xs font-medium rounded-md transition-colors whitespace-nowrap cursor-pointer ${
                                        activeTab === tab.key ?
                                            'bg-brand text-white'
                                        :   'text-muted-foreground hover:text-foreground hover:bg-muted'
                                    }`}
                                >
                                    {tab.label}
                                </button>
                            ))}
                        </div>
                    )}

                    {/* Scrollable items */}
                    <div className='relative flex-1 min-w-0'>
                        {canScrollLeft && (
                            <button
                                type='button'
                                onClick={() => scrollByAmount('left')}
                                className='absolute left-0 top-1/2 -translate-y-1/2 z-10 p-1.5 rounded-full bg-background/90 shadow hover:bg-surface-hover cursor-pointer'
                                aria-label='Scroll left'
                            >
                                <ChevronLeft size={16} strokeWidth={2} aria-hidden='true' />
                            </button>
                        )}

                        {canScrollRight && (
                            <button
                                type='button'
                                onClick={() => scrollByAmount('right')}
                                className='absolute right-0 top-1/2 -translate-y-1/2 z-10 p-1.5 rounded-full bg-background/90 shadow hover:bg-surface-hover cursor-pointer'
                                aria-label='Scroll right'
                            >
                                <ChevronRight size={16} strokeWidth={2} aria-hidden='true' />
                            </button>
                        )}

                        <div
                            ref={scrollRef}
                            className='flex items-center gap-2 overflow-x-auto py-2 scrollbar-hide px-6'
                            role='list'
                            aria-label={`Browse ${activeTab}`}
                        >
                            {activeTab === 'categories' &&
                                categories.map(category => (
                                    <Link
                                        key={category.id}
                                        to={`/category/${category.slug}`}
                                        className='flex items-center gap-2 px-3 py-1.5 rounded-full hover:bg-surface-hover whitespace-nowrap text-sm transition-colors'
                                        role='listitem'
                                    >
                                        <CategoryIcon icon={category.icon} size='sm' />
                                        <span>{category.name}</span>
                                    </Link>
                                ))}

                            {activeTab === 'tags' &&
                                tags.map(tag => (
                                    <Link
                                        key={tag.id}
                                        to={`/tag/${tag.slug}`}
                                        className='flex items-center gap-1.5 px-3 py-1.5 rounded-full hover:bg-surface-hover whitespace-nowrap text-sm transition-colors'
                                        role='listitem'
                                    >
                                        <span
                                            className='w-2 h-2 rounded-full shrink-0'
                                            style={{
                                                backgroundColor:
                                                    tag.color || '#6366f1',
                                            }}
                                        />
                                        <span>#{tag.name}</span>
                                        {tag.usage_count > 0 && (
                                            <span className='text-xs text-muted-foreground'>
                                                {tag.usage_count}
                                            </span>
                                        )}
                                    </Link>
                                ))}

                            {activeTab === 'streamers' &&
                                streamers.map(streamer => (
                                    <Link
                                        key={streamer.broadcaster_id}
                                        to={`/broadcaster/${streamer.broadcaster_id}`}
                                        className='flex items-center gap-1.5 px-3 py-1.5 rounded-full hover:bg-surface-hover whitespace-nowrap text-sm transition-colors'
                                        role='listitem'
                                    >
                                        <span className='w-2 h-2 rounded-full bg-primary-500 shrink-0' />
                                        <span>{streamer.broadcaster_name}</span>
                                        <span className='text-xs text-muted-foreground'>
                                            {streamer.clip_count} clips
                                        </span>
                                    </Link>
                                ))}
                        </div>
                    </div>
                </div>
            </div>
        </div>
    );
}
