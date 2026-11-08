import { useState } from 'react';
import { useQuery } from '@tanstack/react-query';
import { Link, useParams } from 'react-router-dom';
import { Container, SEO, Spinner } from '../components';
import { TwitchEmbed } from '../components/clip/TwitchEmbed';
import { countdownApi, formatCountdownDate } from '../lib/countdown-api';
import { isNotFoundError } from '../lib/error-utils';
import { NotFoundPage } from './NotFoundPage';

export function CountdownEditionPage() {
  const { date = '', series = '' } = useParams<{ date: string; series: string }>();
  const [active, setActive] = useState<string | null>(null);
  const { data: edition, isLoading, error } = useQuery({
    queryKey: ['countdown', date, series],
    queryFn: () => countdownApi.get(date, series),
    retry: (count, err) => !isNotFoundError(err) && count < 2,
  });

  if (isLoading) {
    return (
      <div className='flex justify-center py-24'>
        <Spinner size='xl' />
      </div>
    );
  }
  if (error && isNotFoundError(error)) return <NotFoundPage />;
  if (error || !edition) {
    return (
      <Container className='py-16'>
        <p className='text-center text-muted-foreground'>This countdown could not be loaded.</p>
      </Container>
    );
  }

  const day = formatCountdownDate(edition.edition_date);
  const title = `Top ${edition.size} Twitch clips: ${edition.series_name}, ${day}`;
  // Count down from the lowest rank to #1, as the video does.
  const clips = [...edition.clips].sort((a, b) => b.position - a.position);
  const structuredData = {
    '@context': 'https://schema.org',
    '@type': 'ItemList',
    name: title,
    itemListOrder: 'https://schema.org/ItemListOrderAscending',
    numberOfItems: edition.clips.length,
    itemListElement: edition.clips.map(clip => ({
      '@type': 'ListItem',
      position: clip.position,
      url: `${window.location.origin}/clip/${clip.clip_id}`,
      name: clip.title,
    })),
  };

  return (
    <>
      <SEO
        title={title}
        description={`The ${edition.size} Twitch clips that drew views fastest on ${day}, counted down to #1.`}
        canonicalUrl={`/countdowns/${edition.edition_date}/${edition.series_slug}`}
        ogImage={edition.clips[0]?.thumbnail_url}
        structuredData={structuredData}
      />
      <Container className='py-8'>
        <nav className='mb-4 text-sm'>
          <Link to='/countdowns' className='text-muted-foreground hover:text-foreground'>
            ← All countdowns
          </Link>
        </nav>
        <p className='kicker mb-2'>{day}</p>
        <h1 className='text-3xl lg:text-4xl text-foreground mb-2'>
          Top {edition.size}: {edition.series_name}
        </h1>
        <p className='text-muted-foreground mb-8'>
          Ranked by how fast each clip drew views. Every clip belongs to its broadcaster; follow them on Twitch.
        </p>
        {clips.length === 0 ? (
          <p className='text-center text-muted-foreground py-16'>The clips in this countdown are no longer available.</p>
        ) : (
          <ol className='mx-auto max-w-4xl space-y-10'>
            {clips.map(clip => (
              <li key={clip.clip_id} aria-label={`Number ${clip.position}`}>
                <div className='mb-3 flex items-baseline gap-4'>
                  <span className='font-heading text-4xl font-bold text-brand' aria-hidden='true'>
                    #{clip.position}
                  </span>
                  <div className='min-w-0'>
                    <h2 className='text-xl text-foreground'>
                      <Link to={`/clip/${clip.clip_id}`} className='hover:underline'>
                        {clip.title}
                      </Link>
                    </h2>
                    <Link
                      to={`/broadcaster/${clip.broadcaster_id || clip.broadcaster_name}`}
                      className='font-mono text-[11px] uppercase tracking-[0.04em] text-muted-foreground hover:text-foreground'
                    >
                      {clip.broadcaster_name}
                    </Link>
                  </div>
                </div>
                <TwitchEmbed
                  clipId={clip.twitch_clip_id}
                  thumbnailUrl={clip.thumbnail_url}
                  title={clip.title}
                  active={active === clip.clip_id}
                  onActivate={() => setActive(clip.clip_id)}
                />
              </li>
            ))}
          </ol>
        )}
      </Container>
    </>
  );
}
