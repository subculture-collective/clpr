import { useInfiniteQuery } from '@tanstack/react-query';
import { Link } from 'react-router-dom';
import { Container, SEO, Spinner } from '../components';
import { Button } from '../components/ui';
import { CreatorConnectOffer } from '../components/countdown/CreatorConnectOffer';
import {
  countdownApi,
  countdownPath,
  formatCountdownDate,
  type PublicCountdownEdition,
} from '../lib/countdown-api';

function EditionCard({ edition }: { edition: PublicCountdownEdition }) {
  const top = edition.clips[0];
  return (
    <Link
      to={countdownPath(edition)}
      state={{ play: true }}
      className='group block overflow-hidden rounded-xl border border-border bg-surface hover:border-brand transition-colors'
    >
      <div className='relative aspect-video bg-neutral-950'>
        {top?.thumbnail_url && (
          <img
            src={top.thumbnail_url}
            alt=''
            loading='lazy'
            decoding='async'
            className='absolute inset-0 h-full w-full object-cover group-hover:opacity-90'
          />
        )}
        <span className='absolute left-3 top-3 bg-black/80 px-2.5 py-1 font-heading text-sm font-bold uppercase tracking-[0.04em] text-white'>
          Top {edition.size}
        </span>
      </div>
      <div className='p-4'>
        <h3 className='font-semibold text-foreground'>{edition.series_name}</h3>
        {top && (
          <p className='mt-1 text-sm text-muted-foreground line-clamp-2'>
            #{top.position}: {top.title} · {top.broadcaster_name}
          </p>
        )}
      </div>
    </Link>
  );
}

export function CountdownsPage() {
  const { data, isLoading, isError, hasNextPage, fetchNextPage, isFetchingNextPage } = useInfiniteQuery({
    queryKey: ['countdowns'],
    queryFn: ({ pageParam }) => countdownApi.list(pageParam ? { before: pageParam } : {}),
    initialPageParam: '',
    getNextPageParam: last => last.nextBefore ?? undefined,
  });

  const days = new Map<string, PublicCountdownEdition[]>();
  for (const edition of data?.pages.flatMap(page => page.editions) ?? []) {
    days.set(edition.edition_date, [...(days.get(edition.edition_date) ?? []), edition]);
  }

  return (
    <>
      <SEO
        title='Daily countdowns'
        description='The most-watched Twitch clips of each day, counted down, overall and by category.'
        canonicalUrl='/countdowns'
      />
      <Container className='py-8'>
        <h1 className='text-3xl font-bold text-foreground mb-2'>Daily countdowns</h1>
        <p className='text-muted-foreground mb-8'>
          The Twitch clips that drew views fastest each day, counted down. Every clip belongs to its broadcaster.
        </p>
        <CreatorConnectOffer />
        {isLoading ? (
          <div className='flex justify-center py-16'>
            <Spinner size='xl' />
          </div>
        ) : isError ? (
          <p className='text-center text-muted-foreground py-16'>Countdowns could not be loaded.</p>
        ) : days.size === 0 ? (
          <p className='text-center text-muted-foreground py-16'>No countdowns have been published yet.</p>
        ) : (
          <>
            {[...days.entries()].map(([date, editions]) => (
              <section key={date} className='mb-10' aria-labelledby={`countdowns-${date}`}>
                <h2 id={`countdowns-${date}`} className='kicker mb-4'>
                  {formatCountdownDate(date)}
                </h2>
                <div className='grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4'>
                  {editions.map(edition => (
                    <EditionCard key={edition.series_slug} edition={edition} />
                  ))}
                </div>
              </section>
            ))}
            {hasNextPage && (
              <div className='text-center'>
                <Button variant='outline' size='lg' onClick={() => void fetchNextPage()} disabled={isFetchingNextPage}>
                  {isFetchingNextPage ? 'Loading…' : 'Older countdowns'}
                </Button>
              </div>
            )}
          </>
        )}
      </Container>
    </>
  );
}
