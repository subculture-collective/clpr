import { useEffect, useRef, useState } from 'react';
import { SkipBack, SkipForward } from 'lucide-react';
import { Link } from 'react-router-dom';
import { TwitchEmbed } from '../clip/TwitchEmbed';
import { AutoAdvanceCountdown } from '../playlist/AutoAdvanceCountdown';
import { Button } from '../ui';
import { cn } from '@/lib/utils';
import type { PublicCountdownClip } from '@/lib/countdown-api';

interface CountdownPlayerProps {
  /** In playing order, lowest rank first. */
  clips: PublicCountdownClip[];
  autoStart?: boolean;
}

/** Plays one countdown as a playlist: one Twitch player and the ranked list. */
export function CountdownPlayer({ clips, autoStart = false }: CountdownPlayerProps) {
  const [index, setIndex] = useState(0);
  const [started, setStarted] = useState(autoStart);
  const [showing, setShowing] = useState(false);
  const [paused, setPaused] = useState(false);
  const playerRef = useRef<HTMLDivElement>(null);
  const listRef = useRef<HTMLOListElement>(null);
  const currentRowRef = useRef<HTMLLIElement>(null);

  const clip = clips[index];
  const hasPrevious = index > 0;
  const hasNext = index < clips.length - 1;

  const play = (next: number) => {
    setIndex(next);
    setStarted(true);
    setPaused(false);
  };

  // Only the list scrolls, so an advancing playlist never moves the page.
  useEffect(() => {
    if (listRef.current && currentRowRef.current) {
      listRef.current.scrollTop = currentRowRef.current.offsetTop - 8;
    }
  }, [index]);

  return (
    <div className='grid gap-6 lg:grid-cols-[minmax(0,1fr)_22rem]'>
      <div ref={playerRef} className='min-w-0 scroll-mt-20'>
        {/* The key gives each clip a fresh player and clears a previous clip's error. */}
        <TwitchEmbed
          key={clip.twitch_clip_id}
          clipId={clip.twitch_clip_id}
          thumbnailUrl={clip.thumbnail_url}
          title={clip.title}
          active={started}
          onActivate={() => setStarted(true)}
          onPlayingChange={setShowing}
        />
        {showing && hasNext && clip.duration ? (
          <AutoAdvanceCountdown
            key={clip.clip_id}
            duration={clip.duration}
            paused={paused}
            onPausedChange={setPaused}
            onComplete={() => play(index + 1)}
          />
        ) : null}
        <div className='mt-3 flex items-start gap-4'>
          <span className='font-heading text-4xl font-bold text-brand'>#{clip.position}</span>
          <div className='min-w-0 flex-1'>
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
          <div className='flex shrink-0 gap-2'>
            <Button variant='outline' size='sm' aria-label='Previous clip' disabled={!hasPrevious} onClick={() => play(index - 1)}>
              <SkipBack size={18} aria-hidden='true' />
            </Button>
            <Button variant='outline' size='sm' aria-label='Next clip' disabled={!hasNext} onClick={() => play(index + 1)}>
              <SkipForward size={18} aria-hidden='true' />
            </Button>
          </div>
        </div>
      </div>

      <ol ref={listRef} className='relative space-y-1 lg:max-h-[34rem] lg:overflow-y-auto'>
        {clips.map((item, i) => (
          <li key={item.clip_id} ref={i === index ? currentRowRef : undefined} aria-label={`Number ${item.position}`}>
            <button
              type='button'
              aria-current={i === index ? 'true' : undefined}
              aria-label={`#${item.position} ${item.title}, ${item.broadcaster_name}`}
              onClick={() => {
                play(i);
                playerRef.current?.scrollIntoView?.({ block: 'nearest' });
              }}
              className={cn(
                'flex min-h-11 w-full cursor-pointer items-center gap-3 rounded-lg p-2 text-left transition-colors',
                i === index ? 'bg-surface-raised' : 'hover:bg-surface-hover',
              )}
            >
              <span className='w-9 shrink-0 text-right font-heading font-bold text-brand'>#{item.position}</span>
              {item.thumbnail_url ? (
                <img src={item.thumbnail_url} alt='' loading='lazy' decoding='async' className='h-12 w-20 shrink-0 rounded object-cover' />
              ) : (
                <span className='h-12 w-20 shrink-0 rounded bg-surface-raised' />
              )}
              <span className='min-w-0'>
                <span className='line-clamp-2 text-sm font-medium text-foreground'>{item.title}</span>
                <span className='block truncate text-xs text-muted-foreground'>{item.broadcaster_name}</span>
              </span>
            </button>
          </li>
        ))}
      </ol>
    </div>
  );
}
