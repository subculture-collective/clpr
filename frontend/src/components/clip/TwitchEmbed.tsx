// TWITCH COMPLIANCE:
// This component embeds Twitch clips using ONLY official Twitch embed URLs.
// See: https://dev.twitch.tv/docs/embed/video-and-clips/
// See: https://legal.twitch.com/legal/developer-agreement/
// See: docs/compliance/twitch-embeds.md for full compliance documentation
//
// COMPLIANCE REQUIREMENTS:
// - Uses official clips.twitch.tv/embed URL only (no custom players)
// - Includes 'parent' parameter with actual domain (required by Twitch)
// - HTTPS only (required by Twitch)
// - No re-hosting, proxying, or downloading of video files
// - No stripping of Twitch branding or attribution
// - Respects creator's right to delete clips (graceful error handling)

import { useState, useEffect } from 'react';
import { useVolumePreference } from '@/hooks';
import { MutedIcon } from '@/components/ui';
import { Play } from 'lucide-react';

interface TwitchEmbedProps {
  clipId: string;
  autoplay?: boolean;
  muted?: boolean;
  thumbnailUrl?: string;
  title?: string;
}

export function TwitchEmbed({
  clipId,
  autoplay = false,
  muted = true,
  thumbnailUrl,
  title = 'Twitch Clip'
}: TwitchEmbedProps) {
  const [isLoaded, setIsLoaded] = useState(autoplay);

  // Keep feed cards poster-first and allow only one active third-party embed.
  useEffect(() => {
    const deactivate = (event: Event) => {
      const activeClip = (event as CustomEvent<string>).detail;
      if (activeClip !== clipId) setIsLoaded(false);
    };
    window.addEventListener('clpr:twitch-play', deactivate);
    return () => window.removeEventListener('clpr:twitch-play', deactivate);
  }, [clipId]);
  const [hasError, setHasError] = useState(false);
  const [showMutedIndicator, setShowMutedIndicator] = useState(true);
  const { embedMuted: volumePreferredMuted, hasSetPreference, setUnmutedPreference } = useVolumePreference();

  // Determine mute state for embed:
  // - Before loaded (thumbnail shown): use the prop default (typically muted=true)
  // - After loaded (iframe shown): use user's volume preference from localStorage
  // This ensures the iframe URL is generated with the correct mute parameter
  const embedMuted = isLoaded ? volumePreferredMuted : muted;

  // COMPLIANCE: Get the actual parent domain for Twitch embed
  // Twitch requires the 'parent' parameter to match the actual domain hosting the embed
  // This is a security measure to prevent unauthorized embedding
  // See: https://dev.twitch.tv/docs/embed/video-and-clips/#embedded-experiences
  const parentDomain = typeof window !== 'undefined'
    ? window.location.hostname
    : 'localhost';

  // Auto-hide muted indicator after 3 seconds
  useEffect(() => {
    if (!embedMuted || hasSetPreference) {
      return;
    }

    const timer = setTimeout(() => {
      setShowMutedIndicator(false);
    }, 3000);

    return () => clearTimeout(timer);
  }, [embedMuted, hasSetPreference]);

  // COMPLIANCE: Official Twitch embed URL only
  // MUST use https://clips.twitch.tv/embed (no custom players, no video re-hosting)
  // Per Twitch Developer Agreement, we MUST NOT:
  // - Re-host or proxy video files
  // - Use unofficial embed methods
  // - Strip Twitch branding or attribution
  // - Download or cache video content
  const embedUrl = `https://clips.twitch.tv/embed?clip=${clipId}&parent=${parentDomain}&autoplay=${isLoaded ? 'true' : 'false'}&muted=${embedMuted}`;

  const handleError = () => {
    setHasError(true);
  };

  if (hasError) {
    return (
      <div className="relative w-full pt-[56.25%] bg-neutral-900 rounded-lg flex items-center justify-center">
        <div className="absolute inset-0 flex flex-col items-center justify-center text-white p-4">
          <svg className="w-12 h-12 mb-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
          </svg>
          <p className="text-sm text-center">This clip is no longer available</p>
          <a
            href={`https://clips.twitch.tv/${clipId}`}
            target="_blank"
            rel="noopener noreferrer"
            className="mt-2 text-primary-400 hover:text-primary-300 text-sm underline"
          >
            Try viewing on Twitch
          </a>
        </div>
      </div>
    );
  }

  if (!isLoaded) {
    return (
      <div
        className="relative w-full pt-[56.25%] bg-black rounded-lg overflow-hidden"
      >
        {/* Thumbnail shown while waiting for viewport intersection */}
        {thumbnailUrl && (
          <img
            src={thumbnailUrl}
            alt={title}
            loading="lazy"
            decoding="async"
            className="absolute inset-0 w-full h-full object-cover"
            width="1920"
            height="1080"
          />
        )}
        <button
          type="button"
          className="absolute inset-0 m-auto flex h-14 w-14 items-center justify-center rounded-full bg-black/75 text-white transition-colors hover:bg-primary-600 focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-white"
          aria-label={`Play ${title}`}
          onClick={() => {
            window.dispatchEvent(new CustomEvent('clpr:twitch-play', { detail: clipId }));
            setIsLoaded(true);
          }}
        >
          <Play className="h-6 w-6 fill-current" aria-hidden="true" />
        </button>
      </div>
    );
  }

  return (
    <div className="relative w-full pt-[56.25%] bg-black rounded-lg overflow-hidden">
      <iframe
        src={embedUrl}
        className="absolute inset-0 w-full h-full"
        allowFullScreen
        title={title}
        onError={handleError}
        allow="autoplay; fullscreen"
      />

      {/* Muted indicator - shown when video is muted and user hasn't set a preference yet */}
      {embedMuted && !hasSetPreference && showMutedIndicator && (
        <div
          className="absolute top-3 left-3 bg-black/70 hover:bg-black/90 text-white px-2 py-1 rounded text-xs font-medium flex items-center gap-1 cursor-pointer transition-opacity duration-500 pointer-events-auto"
          onClick={setUnmutedPreference}
          role="button"
          tabIndex={0}
          onKeyDown={(e) => {
            if (e.key === 'Enter' || e.key === ' ') {
              e.preventDefault();
              setUnmutedPreference();
            }
          }}
          aria-label="Video is muted, click to enable sound on future videos"
          title="Video starts muted for autoplay compatibility. Click to enable sound on future videos. (Reload this clip to hear it)"
        >
          <MutedIcon size="sm" />
          <span>Muted</span>
        </div>
      )}
    </div>
  );
}
