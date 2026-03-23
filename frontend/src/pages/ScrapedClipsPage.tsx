import { useState, useEffect } from 'react';
import { useSearchParams } from 'react-router-dom';
import { Container, SEO } from '../components';
import { ScrapedClipFeed } from '../components/clip/ScrapedClipFeed';
import type { SortOption } from '../types/clip';

type ScrapedClipsTab = 'trending' | 'new' | 'views';

export function ScrapedClipsPage() {
  const [searchParams, setSearchParams] = useSearchParams();
  const [top10kEnabled, setTop10kEnabled] = useState(
    searchParams.get('top10k_streamers') === 'true'
  );

  // Sync state with URL params when searchParams changes
  useEffect(() => {
    setTop10kEnabled(searchParams.get('top10k_streamers') === 'true');
  }, [searchParams]);

  // Get active tab from URL or default to 'trending'
  const activeTab = (searchParams.get('tab') as ScrapedClipsTab) || 'trending';

  const handleTabChange = (tab: ScrapedClipsTab) => {
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

  const tabs: { value: ScrapedClipsTab; label: string; description: string }[] = [
    {
      value: 'trending',
      label: 'Trending',
      description: 'Most popular on Twitch',
    },
    {
      value: 'new',
      label: 'Latest',
      description: 'Recently scraped clips',
    },
    {
      value: 'views',
      label: 'Top Views',
      description: 'Most viewed clips',
    },
  ];

  return (
    <>
      <SEO
        title="Discover Clips"
        description="Discover trending, latest, and most viewed clips from Twitch. Submit them as posts to share with the community."
        canonicalUrl="/discover"
      />
      <Container className="py-8">
      <div className="max-w-4xl mx-auto">
        {/* Header */}
        <div className="mb-6">
          <h1 className="text-3xl font-bold text-foreground mb-2">
            Discover on Twitch
          </h1>
          <p className="text-muted-foreground">
            Found these great clips from Twitch. You can submit them as posts to the community.
          </p>
        </div>

        {/* Info Badge */}
        <div className="bg-blue-50 dark:bg-blue-900/20 border border-blue-200 dark:border-blue-800 rounded-xl p-4 mb-6">
          <div className="flex items-start gap-3">
            <svg 
              className="w-5 h-5 text-blue-600 dark:text-blue-400 mt-0.5 shrink-0" 
              fill="none" 
              stroke="currentColor" 
              viewBox="0 0 24 24"
            >
              <path 
                strokeLinecap="round" 
                strokeLinejoin="round" 
                strokeWidth={2} 
                d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" 
              />
            </svg>
            <div className="text-sm">
              <p className="font-medium text-blue-900 dark:text-blue-100 mb-1">
                From Twitch
              </p>
              <p className="text-blue-700 dark:text-blue-300">
                These clips are automatically discovered from Twitch and have not been submitted by community members yet. 
                Click "Post This Clip" to share them with everyone!
              </p>
            </div>
          </div>
        </div>

        {/* Tabs */}
        <div className="bg-card border border-border rounded-xl p-2 mb-6">
          <div className="flex flex-wrap gap-2">
            {tabs.map((tab) => (
              <button
                key={tab.value}
                onClick={() => handleTabChange(tab.value)}
                role="tab"
                aria-selected={activeTab === tab.value}
                className={`
                  flex-1 min-w-[120px] px-4 py-3 rounded-lg text-sm font-medium transition-colors
                  ${
                    activeTab === tab.value
                      ? 'bg-primary-500 text-white'
                      : 'text-muted-foreground hover:text-foreground hover:bg-accent'
                  }
                `}
              >
                <div className="font-semibold">{tab.label}</div>
                <div className="text-xs mt-0.5 opacity-90">
                  {tab.description}
                </div>
              </button>
            ))}
          </div>
        </div>

        {/* Top 10k Streamers Toggle */}
        <div className="bg-card border border-border rounded-xl p-4 mb-6">
          <div className="flex items-center justify-between">
            <div>
              <div className="font-medium text-foreground">
                Top 10k Streamers Only
              </div>
              <div className="text-sm text-muted-foreground mt-1">
                Filter clips to only show content from the top 10,000 streamers
              </div>
            </div>
            <button
              onClick={handleTop10kToggle}
              className={`
                relative inline-flex h-6 w-11 items-center rounded-full transition-colors
                ${top10kEnabled ? 'bg-primary-500' : 'bg-muted'}
              `}
              role="switch"
              aria-checked={top10kEnabled}
              aria-label="Toggle Top 10k Streamers filter"
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

        {/* Scraped Clip Feed */}
        <ScrapedClipFeed
          title={tabs.find((t) => t.value === activeTab)?.label || 'Discover'}
          description={
            tabs.find((t) => t.value === activeTab)?.description || ''
          }
          defaultSort={activeTab as SortOption}
          filters={{
            top10k_streamers: top10kEnabled,
          }}
        />
      </div>
    </Container>
    </>
  );
}
