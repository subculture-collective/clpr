import { useState, useEffect } from 'react';
import { useSearchParams } from 'react-router-dom';
import { Container, SEO } from '../components';
import { ClipFeed } from '../components/clip';
import type { SortOption } from '../types/clip';

type DiscoveryTab = 'trending' | 'new' | 'views';

export function DiscoveryPage() {
    const [searchParams, setSearchParams] = useSearchParams();
    const [top10kEnabled, setTop10kEnabled] = useState(
        searchParams.get('top10k_streamers') === 'true',
    );

    // Sync state with URL params when searchParams changes
    useEffect(() => {
        setTop10kEnabled(searchParams.get('top10k_streamers') === 'true');
    }, [searchParams]);

    // Get active tab from URL or default to 'trending'
    const tabParam = searchParams.get('tab') as DiscoveryTab | null;
    const activeTab: DiscoveryTab =
        tabParam === 'trending' || tabParam === 'new' || tabParam === 'views' ?
            tabParam
        :   'trending';

    const handleTabChange = (tab: DiscoveryTab) => {
        const params = new URLSearchParams(searchParams);
        params.set('tab', tab);
        setSearchParams(params);
    };

    const handleTop10kToggle = () => {
        const newValue = !top10kEnabled;
        setTop10kEnabled(newValue);

        const params = new URLSearchParams(searchParams);
        if (newValue) {
            params.set('top10k_streamers', 'true');
        } else {
            params.delete('top10k_streamers');
        }
        setSearchParams(params);
    };

    const tabs: { value: DiscoveryTab; label: string; description: string }[] =
        [
            {
                value: 'trending',
                label: 'Trending',
                description: 'What everyone is watching right now',
            },
            {
                value: 'new',
                label: 'Latest',
                description: 'Fresh clips just added',
            },
            {
                value: 'views',
                label: 'Top Views',
                description: 'Most viewed clips across Twitch',
            },
        ];

    return (
        <>
            <SEO
                title='Discover Clips'
                description='Explore trending, latest, and most-viewed Twitch clips, including community posts and scraped content.'
                canonicalUrl='/discover'
            />
            <Container className='py-8'>
                <div className='max-w-4xl mx-auto'>
                    {/* Header */}
                    <div className='mb-6'>
                        <h1 className='text-3xl font-bold text-foreground mb-2'>
                            Discover Clips
                        </h1>
                        <p className='text-muted-foreground'>
                            Explore trending, latest, and top-viewed clips from
                            Twitch, including community posts
                        </p>
                    </div>

                    {/* Tabs */}
                    <div className='bg-card border border-border rounded-xl p-2 mb-6'>
                        <div className='flex flex-wrap gap-2'>
                            {tabs.map(tab => (
                                <button
                                    key={tab.value}
                                    onClick={() => handleTabChange(tab.value)}
                                    className={`
                  flex-1 min-w-[120px] px-4 py-3 rounded-lg text-sm font-medium transition-colors
                  ${
                      activeTab === tab.value ?
                          'bg-primary-500 text-white'
                      :   'text-muted-foreground hover:text-foreground hover:bg-accent'
                  }
                `}
                                >
                                    <div className='font-semibold'>
                                        {tab.label}
                                    </div>
                                    <div className='text-xs mt-0.5 opacity-90'>
                                        {tab.description}
                                    </div>
                                </button>
                            ))}
                        </div>
                    </div>

                    {/* Top 10k Streamers Toggle */}
                    <div className='bg-card border border-border rounded-xl p-4 mb-6'>
                        <div className='flex items-center justify-between'>
                            <div>
                                <div className='font-medium text-foreground'>
                                    Top 10k Streamers Only
                                </div>
                                <div className='text-sm text-muted-foreground mt-1'>
                                    Filter clips to only show content from the
                                    top 10,000 streamers
                                </div>
                            </div>
                            <button
                                onClick={handleTop10kToggle}
                                className={`
                relative inline-flex h-6 w-11 items-center rounded-full transition-colors
                ${
                    top10kEnabled ? 'bg-primary-500' : (
                        'bg-muted'
                    )
                }
              `}
                                role='switch'
                                aria-checked={top10kEnabled}
                            >
                                <span
                                    className={`
                  inline-block h-4 w-4 transform rounded-full bg-white transition-transform
                  ${top10kEnabled ? 'translate-x-6' : 'translate-x-1'}
                `}
                                />
                            </button>
                        </div>
                    </div>

                    {/* Clip Feed */}
                    <ClipFeed
                        title={
                            tabs.find(t => t.value === activeTab)?.label ||
                            'Discover'
                        }
                        description={
                            tabs.find(t => t.value === activeTab)
                                ?.description || ''
                        }
                        defaultSort={activeTab as SortOption}
                        defaultTimeframe={
                            activeTab === 'trending' ? 'day' : undefined
                        }
                        filters={{
                            top10k_streamers: top10kEnabled,
                            show_all_clips: true, // Discovery shows both user-submitted and scraped clips
                        }}
                        discoverMode
                    />
                </div>
            </Container>
        </>
    );
}
