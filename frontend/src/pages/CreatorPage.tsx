import { useInfiniteQuery } from '@tanstack/react-query';
import { useParams } from 'react-router-dom';
import { Button, Container, SEO, Spinner } from '../components';
import { ClipCard } from '../components/clip/ClipCard';
import { ClipCardSkeleton } from '../components/clip/ClipCardSkeleton';
import { fetchCreatorClips } from '../lib/clip-api';

export function CreatorPage() {
  const { creatorId } = useParams<{ creatorId: string }>();
  const {
    data,
    fetchNextPage,
    hasNextPage,
    isFetchingNextPage,
    isLoading,
    isError,
    refetch,
  } = useInfiniteQuery({
    queryKey: ['creator-clips', creatorId],
    queryFn: ({ pageParam }) => fetchCreatorClips({ creatorId: creatorId!, pageParam }),
    getNextPageParam: lastPage => (lastPage.has_more ? lastPage.page + 1 : undefined),
    initialPageParam: 1,
    enabled: !!creatorId,
  });
  const clips = data?.pages.flatMap(page => page.clips ?? []) ?? [];

  if (!creatorId) {
    return (
      <Container className="py-8">
        <div className="text-center text-muted-foreground py-12">
          <h2 className="text-2xl font-bold mb-2">Creator Not Found</h2>
          <p>No creator ID specified.</p>
        </div>
      </Container>
    );
  }

  return (
    <>
      <SEO
        title={`${creatorId} - Creator Profile`}
        description={`View all clips created by ${creatorId} on clpr`}
      />
      <Container className="py-8">
        <div className="mb-6">
          <h1 className="text-3xl font-bold mb-2">Creator Clips</h1>
          <p className="text-muted-foreground">Viewing clips from: {creatorId}</p>
        </div>
        
        {isLoading && (
          <div className="space-y-4">
            {Array.from({ length: 3 }).map((_, i) => (
              <ClipCardSkeleton key={i} />
            ))}
          </div>
        )}

        {isError && (
          <div role="alert" className="rounded-lg border border-border p-4">
            <p>This creator's clips could not be loaded.</p>
            <Button className="mt-3" variant="outline" onClick={() => void refetch()}>Try again</Button>
          </div>
        )}

        {!isLoading && !isError && clips.length === 0 && (
          <p className="text-center text-muted-foreground py-12">No clips from this creator yet.</p>
        )}

        {clips.length > 0 && (
          <div className="space-y-4">
            {clips.map(clip => (
              <ClipCard key={clip.id} clip={clip} />
            ))}
            {hasNextPage && (
              <div className="py-8 flex justify-center">
                {isFetchingNextPage ? (
                  <Spinner size="lg" />
                ) : (
                  <Button onClick={() => void fetchNextPage()}>Load More</Button>
                )}
              </div>
            )}
          </div>
        )}
      </Container>
    </>
  );
}
