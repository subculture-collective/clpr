// TWITCH COMPLIANCE:
// This component embeds live Twitch streams using ONLY the official Twitch Embed SDK.
// See: https://dev.twitch.tv/docs/embed/video-and-clips/
// See: https://legal.twitch.com/legal/developer-agreement/
// See: docs/compliance/twitch-embeds.md for full compliance documentation
//
// COMPLIANCE REQUIREMENTS:
// - Uses official Twitch Embed SDK (embed.twitch.tv/embed/v1.js) - never bundled or re-hosted
// - Loads SDK from official Twitch CDN only (HTTPS)
// - Includes 'parent' parameter with actual domain array (required by Twitch)
// - No custom video streaming or HLS parsing
// - No re-hosting, proxying, or downloading of streams
// - Respects Twitch's embed parameters and layout options
// - Autoplays muted; nothing is drawn over the player (the live badge sits on the page)
// - Links out to Twitch when there is less than the 400×300 minimum player size

import { useEffect, useRef, useState } from 'react';
import { useQuery } from '@tanstack/react-query';
import { AlertTriangle, ExternalLink } from 'lucide-react';
import { useTwitchEmbedFits } from '@/hooks/useTwitchEmbedFits';
import { useRegisterTwitchPlayer } from '@/hooks/useTwitchPlayerLayer';
import { fetchStreamStatus } from '../../lib/stream-api';
import { StreamOfflineScreen } from './StreamOfflineScreen';

import type { TwitchEmbedInstance as TwitchEmbed } from '@/types/twitchEmbed';

export interface TwitchPlayerProps {
  channel: string;
  showChat?: boolean;
}

export function TwitchPlayer({ channel, showChat = false }: TwitchPlayerProps) {
  const [isScriptLoaded, setIsScriptLoaded] = useState(false);

  // Fetch stream status with auto-refresh every 60 seconds
  const {
    data: streamInfo,
    isLoading,
    error
  } = useQuery({
    queryKey: ['streamStatus', channel],
    queryFn: () => fetchStreamStatus(channel),
    refetchInterval: 60000, // Refresh every 60 seconds
    retry: 2,
  });

  // Load Twitch Embed SDK script with reference counting
  useEffect(() => {
    if (window.Twitch) {
      // Use queueMicrotask to defer setState after render
      queueMicrotask(() => setIsScriptLoaded(true));
      return;
    }

    // Check if script is already being loaded
    const existingScript = document.querySelector('script[src="https://embed.twitch.tv/embed/v1.js"]');
    if (existingScript) {
      // Script exists, wait for it to load
      const handleLoad = () => {
        queueMicrotask(() => setIsScriptLoaded(true));
        existingScript.removeEventListener('load', handleLoad);
      };
      existingScript.addEventListener('load', handleLoad);
      return () => existingScript.removeEventListener('load', handleLoad);
    }

    // Create new script if it doesn't exist
    // COMPLIANCE: Load official Twitch Embed SDK from Twitch CDN only
    // Must NOT bundle, re-host, or modify this script
    // See: https://dev.twitch.tv/docs/embed/video-and-clips/
    const script = document.createElement('script');
    script.src = 'https://embed.twitch.tv/embed/v1.js';
    script.async = true;
    script.onload = () => queueMicrotask(() => setIsScriptLoaded(true));
    script.onerror = () => {
      console.error('Failed to load Twitch Embed SDK');
    };

    document.body.appendChild(script);

    // Don't remove script on unmount - it may be used by other instances
  }, []);

  if (isLoading) {
    return (
      <div className="w-full aspect-video bg-surface flex items-center justify-center">
        <div className="text-center">
          <div className="inline-block animate-spin rounded-full h-12 w-12 border-t-2 border-b-2 border-purple-500 mb-4"></div>
          <p className="text-muted-foreground">Loading stream...</p>
        </div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="w-full aspect-video bg-surface flex items-center justify-center">
        <div className="text-center text-red-500">
          <p className="text-xl mb-2"><AlertTriangle size={16} strokeWidth={1.75} className="inline" /></p>
          <p>Failed to load stream information</p>
        </div>
      </div>
    );
  }

  if (!streamInfo?.is_live) {
    return <StreamOfflineScreen channel={channel} streamInfo={streamInfo} />;
  }

  return <LiveEmbed channel={channel} showChat={showChat} isScriptLoaded={isScriptLoaded} />;
}

interface LiveEmbedProps {
  channel: string;
  showChat: boolean;
  isScriptLoaded: boolean;
}

// Mounted only while the channel is live so the size check measures the real box.
function LiveEmbed({ channel, showChat, isScriptLoaded }: LiveEmbedProps) {
  const boxRef = useRef<HTMLDivElement>(null);
  const embedRef = useRef<HTMLDivElement>(null);
  const embedInstanceRef = useRef<TwitchEmbed | null>(null);
  const fits = useTwitchEmbedFits(boxRef);
  useRegisterTwitchPlayer(boxRef, fits === true);

  // Initialize Twitch Embed when the script is loaded and the box is large enough
  useEffect(() => {
    if (!isScriptLoaded || !fits || !embedRef.current || !window.Twitch) {
      return;
    }

    const container = embedRef.current;
    // Clear the previous iframe even when the SDK has no destroy method.
    if (embedInstanceRef.current) {
      embedInstanceRef.current.destroy?.();
      container.replaceChildren();
      embedInstanceRef.current = null;
    }

    // Get the parent domain for embed security
    const parentDomain = window.location.hostname;

    // Create new embed
    try {
      const embed = new window.Twitch.Embed(container.id, {
        width: '100%',
        height: '100%',
        channel: channel,
        layout: showChat ? 'video-with-chat' : 'video',
        autoplay: true,
        // Autoplay is always muted; viewers unmute with Twitch's own control.
        muted: true,
        parent: [parentDomain],
      });
      embedInstanceRef.current = embed;
    } catch (error) {
      console.error('Failed to initialize Twitch Embed:', error);
    }

    return () => {
      if (embedInstanceRef.current) {
        embedInstanceRef.current.destroy?.();
        container.replaceChildren();
        embedInstanceRef.current = null;
      }
    };
  }, [channel, showChat, isScriptLoaded, fits]);

  return (
    <div ref={boxRef} className="relative w-full aspect-video bg-black">
      {fits === false ? (
        <div className="absolute inset-0 grid place-items-center p-4 text-center text-white">
          <div>
            <p className="text-sm font-semibold">Not enough room for the Twitch player.</p>
            <a
              href={`https://www.twitch.tv/${encodeURIComponent(channel)}`}
              target="_blank"
              rel="noopener noreferrer"
              className="mt-2 inline-flex min-h-11 items-center gap-2 text-sm font-semibold text-cyan-300 underline"
            >
              Watch on Twitch <ExternalLink size={16} aria-hidden="true" />
            </a>
          </div>
        </div>
      ) : (
        <div
          id={`twitch-embed-${channel}`}
          ref={embedRef}
          className="w-full h-full"
        />
      )}
    </div>
  );
}
