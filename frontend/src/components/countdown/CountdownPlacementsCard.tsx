import { useQuery } from '@tanstack/react-query';
import { Link } from 'react-router-dom';
import { ExternalLink } from 'lucide-react';
import { Alert, Button, Card, CardBody, CardHeader } from '@/components';
import { countdownPath, formatCountdownDate } from '@/lib/countdown-api';
import {
  countdownControlsApi,
  featureMyClipsUrl,
  youTubeVideoUrl,
  type CountdownPlacement,
  type CountdownVideo,
} from '@/lib/countdown-controls-api';
import { getHttpStatus } from '@/lib/error-utils';

const views = new Intl.NumberFormat('en-US');

function VideoLink({ video, kind }: { video: CountdownVideo; kind: 'countdown' | 'short' }) {
  const label = kind === 'short' ? 'Short' : 'Countdown video';
  return (
    <a
      href={youTubeVideoUrl(video, kind)}
      target='_blank'
      rel='noopener noreferrer'
      className='inline-flex min-h-9 items-center gap-1.5 border border-line-strong px-3 text-sm hover:border-brand hover:text-foreground transition-colors'
    >
      {label}{' '}
      <span className='font-mono text-muted-foreground'>
        {video.view_count === undefined ? 'views pending' : `${views.format(video.view_count)} views`}
      </span>
      <ExternalLink size={14} aria-hidden='true' />
      <span className='sr-only'> (opens YouTube)</span>
    </a>
  );
}

function PlacementRow({ placement }: { placement: CountdownPlacement }) {
  return (
    <li className='flex items-start gap-4 py-4 first:pt-0 last:pb-0'>
      <div className='w-28 shrink-0'>
        {placement.thumbnail_url && (
          <img src={placement.thumbnail_url} alt='' loading='lazy' width='112' height='63' className='aspect-video w-full object-cover' />
        )}
      </div>
      <div className='min-w-0 flex-1'>
        <p className='font-semibold text-foreground truncate'>{placement.title}</p>
        <Link to={countdownPath(placement)} className='text-sm text-muted-foreground hover:text-foreground'>
          <span className='font-heading font-bold text-brand'>#{placement.position}</span> · {placement.series_name} ·{' '}
          {formatCountdownDate(placement.edition_date, 'short')}
        </Link>
        {(placement.countdown || placement.short) && (
          <div className='mt-2 flex flex-wrap gap-2'>
            {placement.countdown && <VideoLink video={placement.countdown} kind='countdown' />}
            {placement.short && <VideoLink video={placement.short} kind='short' />}
          </div>
        )}
      </div>
    </li>
  );
}

/**
 * The creator dashboard's countdown card (#84): where the signed-in
 * broadcaster's clips placed and the views their videos got, with the offer to
 * connect the channel when the broadcaster has not yet featured their clips.
 */
export function CountdownPlacementsCard({ returnTo }: { returnTo: string }) {
  const controls = useQuery({ queryKey: ['countdown-controls'], queryFn: countdownControlsApi.get, retry: false });
  const placements = useQuery({
    queryKey: ['countdown-placements'],
    queryFn: countdownControlsApi.placements,
    retry: false,
    staleTime: 5 * 60 * 1000,
  });

  // A user without a Twitch account, or an API without this feature yet.
  const status = getHttpStatus(controls.error) ?? getHttpStatus(placements.error);
  if (status === 409 || status === 404) return null;

  let list;
  if (placements.isLoading) {
    list = <p className='text-muted-foreground'>Loading…</p>;
  } else if (placements.error) {
    list = <Alert variant='error'>Your countdown placements could not be loaded.</Alert>;
  } else if (!placements.data?.length) {
    list = <p className='text-muted-foreground'>None of your clips has placed in a countdown yet.</p>;
  } else {
    list = (
      <ul className='divide-y divide-border' aria-label='Your countdown placements'>
        {placements.data.map(p => (
          <PlacementRow key={`${p.edition_date}-${p.series_slug}-${p.clip_id}`} placement={p} />
        ))}
      </ul>
    );
  }

  const excluded = controls.data?.excluded;
  const offer = controls.data && !controls.data.featured && !excluded && (
    <div className='mb-6 border border-line-strong p-4'>
      <h3 className='font-semibold text-foreground mb-1'>Feature your clips in the daily countdowns</h3>
      <p className='text-sm text-muted-foreground mb-3'>
        Connect your channel and clpr downloads your clips through Twitch&apos;s official API, so they can rank in the daily
        countdowns and Shorts on YouTube. You see here where each clip placed and the views it got. Twitch asks you to allow
        &ldquo;Manage clips&rdquo;; you can stop at any time in Settings.
      </p>
      <Button asChild variant='primary'>
        <a href={featureMyClipsUrl(returnTo)}>Connect my channel</a>
      </Button>
    </div>
  );

  return (
    <Card className='mb-8' data-testid='countdown-placements'>
      <CardHeader>
        <h2 className='text-xl font-semibold'>Your clips in countdowns</h2>
      </CardHeader>
      <CardBody>
        {offer}
        {excluded && (
          <p className='mb-4 text-sm text-muted-foreground'>
            Your clips are left out of countdowns. You can change this in <Link to='/settings' className='underline'>Settings</Link>.
          </p>
        )}
        {list}
        <p className='mt-4 text-xs text-muted-foreground'>
          View counts come from YouTube and update every few hours. Videos appear here once they are public.
        </p>
      </CardBody>
    </Card>
  );
}
