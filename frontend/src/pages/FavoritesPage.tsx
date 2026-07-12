import { useEffect } from 'react';
import { useSearchParams, Link } from 'react-router-dom';
import { useInView } from 'react-intersection-observer';
import { Container, Button, Spinner, ScrollToTop } from '../components';
import { MiniFooter } from '../components/layout';
import { ClipCard } from '../components/clip/ClipCard';
import { ClipCardSkeleton } from '../components/clip/ClipCardSkeleton';
import { EmptyState } from '../components/clip/EmptyState';
import { useAuth } from '../context/AuthContext';
import { useFavoritesFeed } from '../hooks/useClips';

type FavoriteSort = 'newest' | 'top' | 'discussed';

export function FavoritesPage() {
  const { isAuthenticated, isLoading: authLoading } = useAuth();
  const [searchParams, setSearchParams] = useSearchParams();

  // Get sort from URL or default to 'newest'
  const sort = (searchParams.get('sort') as FavoriteSort) || 'newest';

  // Fetch favorites with infinite query
  const {
    data,
    fetchNextPage,
    hasNextPage,
    isFetchingNextPage,
    isLoading,
    isError,
  } = useFavoritesFeed(sort);

  // Intersection observer for infinite scroll
  const { ref: loadMoreRef, inView } = useInView({
    threshold: 0.5,
  });

  // Load more when the trigger element comes into view
  useEffect(() => {
    if (inView && hasNextPage && !isFetchingNextPage) {
      fetchNextPage();
    }
  }, [inView, hasNextPage, isFetchingNextPage, fetchNextPage]);

  const handleSortChange = (newSort: FavoriteSort) => {
    const params = new URLSearchParams(searchParams);
    params.set('sort', newSort);
    setSearchParams(params);
  };

  // Get all clips from all pages
  const clips = data?.pages.flatMap(page => page.clips) ?? [];

  // Show login CTA for unauthenticated users
  if (!authLoading && !isAuthenticated) {
    return (
      <Container className="py-8">
        <div className="max-w-4xl mx-auto">
          <div className="mb-6">
            <h1 className="text-3xl font-bold mb-2">My Favorites</h1>
            <p className="text-muted-foreground">Your saved clips</p>
          </div>
          <div className="bg-card border border-border rounded-xl p-8 text-center">
            <div className="flex justify-center mb-4 text-muted-foreground">
              <svg className="w-16 h-16" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M4.318 6.318a4.5 4.5 0 000 6.364L12 20.364l7.682-7.682a4.5 4.5 0 00-6.364-6.364L12 7.636l-1.318-1.318a4.5 4.5 0 00-6.364 0z" />
              </svg>
            </div>
            <h2 className="text-xl font-semibold mb-2">Login to view your favorites</h2>
            <p className="text-muted-foreground mb-6">
              Sign in with your Twitch account to save and organize your favorite clips
            </p>
            <Button asChild size="lg">
              <Link to="/auth/twitch">
                <svg className="w-5 h-5 mr-2" fill="currentColor" viewBox="0 0 24 24">
                  <path d="M11.571 4.714h1.715v5.143H11.57zm4.715 0H18v5.143h-1.714zM6 0L1.714 4.286v15.428h5.143V24l4.286-4.286h3.428L22.286 12V0zm14.571 11.143l-3.428 3.428h-3.429l-3 3v-3H6.857V1.714h13.714Z" />
                </svg>
                Login with Twitch
              </Link>
            </Button>
          </div>
        </div>
      </Container>
    );
  }

  return (
    <Container className="py-8">
      <div className="max-w-4xl mx-auto">
        {/* Header */}
        <div className="mb-6">
          <h1 className="text-3xl font-bold text-foreground mb-2">
            My Favorites
          </h1>
          <p className="text-muted-foreground">
            Your saved clips
          </p>
        </div>

        {/* Sort Options */}
        <div className="bg-card border border-border rounded-xl p-2 mb-6">
          <div className="flex flex-wrap gap-2">
            <button
              onClick={() => handleSortChange('newest')}
              className={`
                flex-1 min-w-[120px] px-4 py-3 rounded-lg text-sm font-medium transition-colors
                ${
                  sort === 'newest'
                    ? 'bg-primary-500 text-white'
                    : 'text-muted-foreground hover:text-foreground hover:bg-accent'
                }
              `}
            >
              <div className="font-semibold">Newest</div>
              <div className="text-xs mt-0.5 opacity-90">
                Recently saved
              </div>
            </button>
            <button
              onClick={() => handleSortChange('top')}
              className={`
                flex-1 min-w-[120px] px-4 py-3 rounded-lg text-sm font-medium transition-colors
                ${
                  sort === 'top'
                    ? 'bg-primary-500 text-white'
                    : 'text-muted-foreground hover:text-foreground hover:bg-accent'
                }
              `}
            >
              <div className="font-semibold">Top</div>
              <div className="text-xs mt-0.5 opacity-90">
                Highest rated
              </div>
            </button>
            <button
              onClick={() => handleSortChange('discussed')}
              className={`
                flex-1 min-w-[120px] px-4 py-3 rounded-lg text-sm font-medium transition-colors
                ${
                  sort === 'discussed'
                    ? 'bg-primary-500 text-white'
                    : 'text-muted-foreground hover:text-foreground hover:bg-accent'
                }
              `}
            >
              <div className="font-semibold">Discussed</div>
              <div className="text-xs mt-0.5 opacity-90">
                Most comments
              </div>
            </button>
          </div>
        </div>

        {/* Loading state */}
        {isLoading && (
          <div className="space-y-4">
            {Array.from({ length: 5 }).map((_, i) => (
              <ClipCardSkeleton key={i} />
            ))}
          </div>
        )}

        {/* Error state */}
        {isError && (
          <EmptyState
            title="Error loading favorites"
            message="Something went wrong. Please try again later."
            icon={
              <svg className="w-16 h-16" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
              </svg>
            }
          />
        )}

        {/* Empty state */}
        {!isLoading && !isError && clips.length === 0 && (
          <div className="bg-card border border-border rounded-xl p-8 text-center">
            <div className="flex justify-center mb-4 text-muted-foreground">
              <svg className="w-16 h-16" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M4.318 6.318a4.5 4.5 0 000 6.364L12 20.364l7.682-7.682a4.5 4.5 0 00-6.364-6.364L12 7.636l-1.318-1.318a4.5 4.5 0 00-6.364 0z" />
              </svg>
            </div>
            <h3 className="text-xl font-semibold mb-2">No favorites yet</h3>
            <p className="text-muted-foreground mb-6">
              Start favoriting clips to see them here!
            </p>
            <Button asChild><Link to="/">Browse Clips</Link></Button>
          </div>
        )}

        {/* Clips list */}
        {!isLoading && !isError && clips.length > 0 && (
          <div className="space-y-4">
            {clips.map((clip) => (
              <ClipCard key={clip.id} clip={clip} />
            ))}

            {/* Load more trigger */}
            {hasNextPage && (
              <div ref={loadMoreRef} className="py-8 flex justify-center">
                {isFetchingNextPage ? (
                  <Spinner size="lg" />
                ) : (
                  <Button onClick={() => fetchNextPage()}>Load More</Button>
                )}
              </div>
            )}

            {/* End of results */}
            {!hasNextPage && clips.length > 0 && (
              <div className="text-center py-8 text-muted-foreground">
                <p>You've reached the end!</p>
              </div>
            )}
          </div>
        )}

        {/* Scroll to top button */}
        <ScrollToTop threshold={500} />

        {/* Mini footer for quick access to footer links */}
        <MiniFooter />
      </div>
    </Container>
  );
}
