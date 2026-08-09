import { SEO } from '../components';
import { PlaylistCard } from '../components/playlist/PlaylistCard';
import { Button } from '../components/ui';
import { FeedLayout } from '../components/layout/FeedLayout';
import { FeedSidebar } from '../components/layout/FeedSidebar';
import { useMemo } from 'react';
import { useInfiniteFeaturedPlaylists } from '../hooks/usePlaylist';

export function DiscoveryListsPage() {
  const pageSize = 12;

  const {
    data,
    isLoading,
    isFetchingNextPage,
    hasNextPage,
    fetchNextPage,
  } = useInfiniteFeaturedPlaylists(pageSize);
  const lists = useMemo(() => {
    const uniqueLists = new Map(
      data?.pages.flatMap((page) => page.data).map((list) => [list.id, list]),
    );
    return [...(uniqueLists?.values() ?? [])];
  }, [data?.pages]);

  return (
    <>
      <SEO
        title="Discovery Lists"
        description="Browse curated collections of amazing Twitch clips. Find new content organized by theme, game, and community favorites."
        canonicalUrl="/discover/lists"
      />
      <FeedLayout sidebar={<FeedSidebar />}>
          {/* Header */}
          <div className="mb-8">
            <h1 className="text-3xl font-bold text-foreground mb-2">
              Discovery Lists
            </h1>
            <p className="text-muted-foreground">
              Explore handpicked and auto-generated collections of the best Twitch clips
            </p>
          </div>

          {/* Lists Grid */}
          {isLoading ? (
            <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-6">
              {[...Array(6)].map((_, i) => (
                <div
                  key={i}
                  className="h-80 bg-accent rounded-xl animate-pulse"
                />
              ))}
            </div>
          ) : lists && lists.length > 0 ? (
            <>
              <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-6">
                {lists.map((list) => (
                  <PlaylistCard key={list.id} playlist={list} />
                ))}
              </div>

              {/* Load More Button */}
              {hasNextPage && (
                <div className="text-center pt-8">
                  <Button
                    onClick={() => void fetchNextPage()}
                    variant="outline"
                    size="lg"
                    disabled={isFetchingNextPage}
                  >
                    {isFetchingNextPage ? 'Loading…' : 'Load More Lists'}
                  </Button>
                </div>
              )}
            </>
          ) : (
            <div className="text-center py-12 bg-card border border-border rounded-xl">
              <h2 className="text-xl font-semibold mb-2">
                No Discovery Lists Yet
              </h2>
              <p className="text-muted-foreground">
                Check back soon for curated collections of clips
              </p>
            </div>
          )}
      </FeedLayout>
    </>
  );
}
