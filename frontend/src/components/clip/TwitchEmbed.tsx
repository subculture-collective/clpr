// Twitch playback is deliberately controlled by the owning feed. Inactive clips
// remain thumbnails so long feeds never accumulate off-screen iframes.
// Twitch requires the player to be at least 400×300 and never covered by page
// elements, so nothing is drawn over the iframe and small screens link out.
import { useRef, useState } from 'react';
import { ExternalLink, Play, RotateCw } from 'lucide-react';
import { useVolumePreference } from '@/hooks';
import { useTwitchEmbedFits } from '@/hooks/useTwitchEmbedFits';
import { useRegisterTwitchPlayer } from '@/hooks/useTwitchPlayerLayer';
import { MutedIcon } from '@/components/ui';
import { ErrorEvents, trackEvent } from '@/lib/telemetry';

interface TwitchEmbedProps {
  clipId: string;
  active?: boolean;
  autoplay?: boolean;
  muted?: boolean;
  thumbnailUrl?: string;
  title?: string;
  onActivate?: () => void;
}

export function TwitchEmbed({
  clipId,
  active = false,
  autoplay = false,
  muted = true,
  thumbnailUrl,
  title = 'Twitch Clip',
  onActivate,
}: TwitchEmbedProps) {
  const boxRef = useRef<HTMLDivElement>(null);
  const fits = useTwitchEmbedFits(boxRef);
  const [hasError, setHasError] = useState(false);
  const { embedMuted: volumePreferredMuted, hasSetPreference, setUnmutedPreference } =
    useVolumePreference();
  const embedMuted = active ? volumePreferredMuted : muted;
  const parentDomain = typeof window !== 'undefined' ? window.location.hostname : 'localhost';
  const shouldAutoplay = active || autoplay;
  const embedUrl = `https://clips.twitch.tv/embed?clip=${encodeURIComponent(clipId)}&parent=${encodeURIComponent(parentDomain)}&autoplay=${shouldAutoplay ? 'true' : 'false'}&muted=${embedMuted ? 'true' : 'false'}`;
  const playing = active && fits === true && !hasError;
  useRegisterTwitchPlayer(boxRef, playing);

  const thumbnail = thumbnailUrl ? (
    <img
      src={thumbnailUrl}
      alt=''
      loading='lazy'
      decoding='async'
      className='absolute inset-0 h-full w-full object-cover transition-opacity duration-150 group-hover:opacity-90 motion-reduce:transition-none'
      width='1920'
      height='1080'
    />
  ) : (
    <span className='absolute inset-0 bg-surface-raised' />
  );

  let content;
  if (!active) {
    content = (
      <button
        type='button'
        onClick={() => {
          setHasError(false);
          onActivate?.();
        }}
        className='absolute inset-0 block w-full group cursor-pointer'
        aria-label={`Play ${title}`}
      >
        {thumbnail}
        <span className='absolute inset-0 bg-black/20' />
        <span className='absolute left-1/2 top-1/2 grid size-14 -translate-x-1/2 -translate-y-1/2 place-items-center bg-primary-400 text-background transition-colors group-hover:bg-primary-300'>
          <Play size={28} fill='currentColor' className='translate-x-0.5' aria-hidden='true' />
        </span>
        <span className='absolute bottom-3 left-3 inline-flex items-center gap-1.5 bg-black/75 px-2.5 py-1 text-[11px] font-semibold md:hidden'>
          <RotateCw size={13} aria-hidden='true' /> Landscape recommended
        </span>
      </button>
    );
  } else if (hasError || fits === false) {
    content = (
      <>
        {!hasError && thumbnail}
        <div className='absolute inset-0 grid place-items-center bg-black/70 p-5'>
          <div className='text-center'>
            <p className='font-semibold'>
              {hasError
                ? 'This clip is unavailable here.'
                : 'Not enough room for the Twitch player.'}
            </p>
            <a
              href={`https://clips.twitch.tv/${clipId}`}
              target='_blank'
              rel='noopener noreferrer'
              className='mt-3 inline-flex min-h-11 items-center gap-2 text-sm font-semibold text-cyan-300 underline'
            >
              Watch on Twitch <ExternalLink size={16} aria-hidden='true' />
            </a>
          </div>
        </div>
      </>
    );
  } else if (fits) {
    content = (
      <iframe
        src={embedUrl}
        className='absolute inset-0 h-full w-full border-0'
        allowFullScreen
        title={title}
        onError={() => {
          setHasError(true);
          trackEvent(ErrorEvents.ERROR_OCCURRED, {
            error_type: 'twitch_embed',
            error_message: 'Twitch clip embed failed to load',
            clip_id: clipId,
          });
        }}
        allow='autoplay; fullscreen'
      />
    );
  }

  return (
    <div>
      <div
        ref={boxRef}
        className='relative aspect-[4/3] md:aspect-video w-full overflow-hidden md:rounded-lg bg-neutral-950 text-white'
      >
        {content}
      </div>
      {playing && embedMuted && !hasSetPreference && (
        <button
          type='button'
          onClick={setUnmutedPreference}
          className='mt-1 inline-flex min-h-11 items-center gap-1.5 px-3 md:px-0 text-xs font-semibold text-text-secondary hover:text-text-primary'
        >
          <MutedIcon size='sm' /> Sound off · Turn on for future clips
        </button>
      )}
    </div>
  );
}
