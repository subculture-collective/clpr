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
import { TagChip } from '../tag/TagChip';
import type { Category } from '../../types/category';
import type { Tag } from '../../types/tag';

type NavTab = 'creators' | 'topics' | 'tags';

export function CategoriesNav() {
    const [activeTab, setActiveTab] = useState<NavTab>('creators');
    const [topics, setTopics] = useState<Category[]>([]);
    const [tags, setTags] = useState<Tag[]>([]);
    const [creators, setCreators] = useState<PopularBroadcaster[]>([]);
    const [loading, setLoading] = useState(true);
    const scrollRef = useRef<HTMLDivElement>(null);
    const [canScrollLeft, setCanScrollLeft] = useState(false);
    const [canScrollRight, setCanScrollRight] = useState(false);

    useEffect(() => {
        const fetchAll = async () => {
            try {
                const [featuredRes, tagsRes, creatorsRes] = await Promise.all([
                    categoryApi.listCategories({ type: 'topic', featured: true }),
                    tagApi.listTags({ lane: 'detected', sort: 'popularity', limit: 20 }),
                    fetchPopularBroadcasters(20),
                ]);

                let featuredTopics = featuredRes.categories || [];
                if (featuredTopics.length === 0) {
                    const all = await categoryApi.listCategories({ type: 'topic' });
                    featuredTopics = all.categories || [];
                }
                setTopics(featuredTopics);
                setTags(tagsRes.tags || []);
                setCreators(creatorsRes);
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
    }, [activeTab, topics.length, tags.length, creators.length]);

    const scrollByAmount = (direction: 'left' | 'right') => {
        const el = scrollRef.current;
        if (!el) return;
        const delta = direction === 'left' ? -200 : 200;
        el.scrollBy({ left: delta, behavior: 'smooth' });
    };

    const tabs: { key: NavTab; label: string; count: number }[] = [
        { key: 'creators', label: 'Creators', count: creators.length },
        { key: 'topics', label: 'Topics', count: topics.length },
        { key: 'tags', label: 'Tags', count: tags.length },
    ];

    // Only show tabs that have data
    const visibleTabs = tabs.filter(t => t.count > 0);

    return (
        <div className='border-b border-border bg-background' aria-busy={loading}>
            <nav className='flex min-h-12 items-center gap-2 overflow-x-auto px-4 md:hidden scrollbar-hide' aria-label='Discover creators, topics, and tags'>
                <Link to='/creators' className='inline-flex min-h-9 shrink-0 items-center border border-line-strong bg-transparent px-4 text-foreground font-heading text-[14px] font-bold uppercase tracking-[0.04em]'>Creators</Link>
                <Link to='/topics' className='inline-flex min-h-9 shrink-0 items-center border border-line-strong bg-transparent px-4 text-foreground font-heading text-[14px] font-bold uppercase tracking-[0.04em]'>Topics</Link>
                <Link to='/tags' className='inline-flex min-h-9 shrink-0 items-center border border-line-strong bg-transparent px-4 text-foreground font-heading text-[14px] font-bold uppercase tracking-[0.04em]'>Tags</Link>
            </nav>
            <div className='page-container mx-auto hidden min-h-12 px-4 md:block'>
                <div className='relative flex min-h-12 items-center'>
                    {/* Tab selector */}
                    {visibleTabs.length > 1 && (
                        <div className='flex items-center gap-1 pr-3 mr-3 border-r border-border shrink-0'>
                            {visibleTabs.map(tab => (
                                <button
                                    key={tab.key}
                                    type='button'
                                    onClick={() => setActiveTab(tab.key)}
                                    className={`px-2.5 py-1.5 font-mono text-[11px] font-medium uppercase tracking-[0.08em] transition-colors whitespace-nowrap cursor-pointer ${
                                        activeTab === tab.key ?
                                            'bg-brand text-background'
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
                                className='absolute left-0 top-1/2 -translate-y-1/2 z-10 p-1.5 bg-background border border-line-strong hover:bg-surface-hover cursor-pointer'
                                aria-label='Scroll left'
                            >
                                <ChevronLeft size={16} strokeWidth={2} aria-hidden='true' />
                            </button>
                        )}

                        {canScrollRight && (
                            <button
                                type='button'
                                onClick={() => scrollByAmount('right')}
                                className='absolute right-0 top-1/2 -translate-y-1/2 z-10 p-1.5 bg-background border border-line-strong hover:bg-surface-hover cursor-pointer'
                                aria-label='Scroll right'
                            >
                                <ChevronRight size={16} strokeWidth={2} aria-hidden='true' />
                            </button>
                        )}

                        <div
                            ref={scrollRef}
                            className='flex items-center gap-2 overflow-x-auto py-2 scrollbar-hide px-6'
                            role='navigation'
                            aria-label={`Browse ${activeTab}`}
                        >
                            {activeTab === 'creators' &&
                                <>
                                    <Link
                                        to='/creators'
                                        className='flex items-center gap-1.5 bg-primary-400 px-3 py-1.5 font-heading text-[14px] font-bold uppercase tracking-[0.04em] text-background whitespace-nowrap hover:bg-primary-300 hover:text-background'
                                    >
                                        Explore creators
                                    </Link>
                                    {creators.map(creator => (
                                        <Link
                                            key={creator.broadcaster_id}
                                            to={`/broadcaster/${creator.broadcaster_id}`}
                                            className='flex items-center gap-1.5 px-3 py-1.5 border border-transparent hover:border-line-strong whitespace-nowrap text-sm text-foreground transition-colors'
                                        >
                                            <span className='w-1.5 h-1.5 rounded-full bg-tally shrink-0' />
                                            <span>{creator.broadcaster_name}</span>
                                            <span className='font-mono text-[11px] text-muted-foreground'>
                                                {creator.clip_count} clips
                                            </span>
                                        </Link>
                                    ))}
                                </>
                            }

                            {activeTab === 'topics' &&
                                topics.map(topic => (
                                    <Link
                                        key={topic.id}
                                        to={`/topics/${topic.slug}`}
                                        className='flex items-center gap-2 px-3 py-1.5 border border-transparent hover:border-line-strong whitespace-nowrap text-sm text-foreground transition-colors'
                                    >
                                        <CategoryIcon icon={topic.icon} size='sm' />
                                        <span>{topic.name}</span>
                                    </Link>
                                ))}

                            {activeTab === 'tags' &&
                                tags.map(tag => (
                                    <span key={tag.id} className='shrink-0'>
                                        <TagChip tag={tag} showCount />
                                    </span>
                                ))}

                        </div>
                    </div>
                </div>
            </div>
        </div>
    );
}
