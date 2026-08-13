// Twitch playback is deliberately controlled by the owning feed. Inactive clips
// remain thumbnails so long feeds never accumulate off-screen iframes.
import { useEffect, useState } from 'react';
import { ExternalLink, Play, RotateCw } from 'lucide-react';
import { useVolumePreference } from '@/hooks';
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
  const [hasError, setHasError] = useState(false);
  const [showMutedIndicator, setShowMutedIndicator] = useState(true);
  const { embedMuted: volumePreferredMuted, hasSetPreference, setUnmutedPreference } =
    useVolumePreference();
  const embedMuted = active ? volumePreferredMuted : muted;
  const parentDomain = typeof window !== 'undefined' ? window.location.hostname : 'localhost';
  const shouldAutoplay = active || autoplay;
  const embedUrl = `https://clips.twitch.tv/embed?clip=${encodeURIComponent(clipId)}&parent=${encodeURIComponent(parentDomain)}&autoplay=${shouldAutoplay ? 'true' : 'false'}&muted=${embedMuted ? 'true' : 'false'}`;

  useEffect(() => {
    if (!active || !embedMuted || hasSetPreference) return;
    const timer = window.setTimeout(() => setShowMutedIndicator(false), 3000);
    return () => window.clearTimeout(timer);
  }, [active, embedMuted, hasSetPreference]);

  if (active && hasError) {
    return (
      <div className='relative aspect-[4/3] md:aspect-video w-full overflow-hidden md:rounded-lg bg-neutral-950 text-white'>
      <div className='absolute inset-0 grid place-items-center p-5'>
        <div className='text-center'>
          <p className='font-semibold'>This clip is unavailable here.</p>
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
      </div>
    );
  }

  if (!active) {
    return (
      <button
        type='button'
        onClick={() => {
          setHasError(false);
          onActivate?.();
        }}
        className='relative block aspect-[4/3] md:aspect-video w-full group overflow-hidden md:rounded-lg bg-neutral-950 text-white cursor-pointer'
        aria-label={`Play ${title}`}
      >
        {thumbnailUrl ? (
          <img
            src={thumbnailUrl}
            alt=''
            loading='lazy'
            decoding='async'
            className='absolute inset-0 h-full w-full object-cover transition-transform duration-500 group-hover:scale-[1.02] motion-reduce:transition-none'
            width='1920'
            height='1080'
          />
        ) : (
          <span className='absolute inset-0 bg-[radial-gradient(circle_at_30%_20%,rgba(124,58,237,.3),transparent_55%)]' />
        )}
        <span className='absolute inset-0 bg-linear-to-t from-black/60 via-black/5 to-black/15' />
        <span className='absolute left-1/2 top-1/2 grid size-16 -translate-x-1/2 -translate-y-1/2 place-items-center rounded-full border border-white/35 bg-black/65 shadow-2xl backdrop-blur-sm transition-transform group-hover:scale-105'>
          <Play size={28} fill='currentColor' className='translate-x-0.5' aria-hidden='true' />
        </span>
        <span className='absolute bottom-3 left-3 inline-flex items-center gap-1.5 rounded-full bg-black/75 px-2.5 py-1 text-[11px] font-semibold md:hidden'>
          <RotateCw size={13} aria-hidden='true' /> Landscape recommended
        </span>
      </button>
    );
  }

  return (
    <div className='relative aspect-[4/3] md:aspect-video w-full overflow-hidden md:rounded-lg bg-black'>
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
      {embedMuted && !hasSetPreference && showMutedIndicator && (
        <button
          type='button'
          onClick={setUnmutedPreference}
          className='absolute left-3 top-3 inline-flex min-h-11 items-center gap-1.5 rounded-full bg-black/80 px-3 text-xs font-semibold text-white backdrop-blur'
          aria-label='Video is muted; enable sound for future videos'
        >
          <MutedIcon size='sm' /> Muted
        </button>
      )}
    </div>
  );
}
