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

import { useEffect, useRef, useState } from 'react';
import { useQuery } from '@tanstack/react-query';
import { AlertTriangle } from 'lucide-react';
import { fetchStreamStatus } from '../../lib/stream-api';
import { StreamOfflineScreen } from './StreamOfflineScreen';
import { LiveIndicator } from './LiveIndicator';

// Declare Twitch type for the embed SDK
declare global {
  interface Window {
    Twitch?: {
      Embed: new (elementId: string, options: TwitchEmbedOptions) => TwitchEmbed;
    };
  }
}

export interface TwitchEmbedOptions {
  width: string | number;
  height: string | number;
  channel: string;
  layout: 'video' | 'video-with-chat';
  autoplay: boolean;
  muted: boolean;
  parent: string[];
}

export interface TwitchEmbed {
  destroy: () => void;
}

export interface TwitchPlayerProps {
  channel: string;
  showChat?: boolean;
}

export function TwitchPlayer({ channel, showChat = false }: TwitchPlayerProps) {
  const [isScriptLoaded, setIsScriptLoaded] = useState(false);
  const embedRef = useRef<HTMLDivElement>(null);
  const embedInstanceRef = useRef<TwitchEmbed | null>(null);

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

  // Initialize Twitch Embed when script loads and stream is live
  useEffect(() => {
    if (!isScriptLoaded || !streamInfo?.is_live || !embedRef.current || !window.Twitch) {
      return;
    }

    // Destroy existing embed if it exists
    if (embedInstanceRef.current) {
      embedInstanceRef.current.destroy();
      embedInstanceRef.current = null;
    }

    // Get the parent domain for embed security
    const parentDomain = window.location.hostname;

    // Create new embed
    try {
      const embed = new window.Twitch.Embed(embedRef.current.id, {
        width: '100%',
        height: '100%',
        channel: channel,
        layout: showChat ? 'video-with-chat' : 'video',
        autoplay: true,
        muted: false,
        parent: [parentDomain],
      });
      embedInstanceRef.current = embed;
    } catch (error) {
      console.error('Failed to initialize Twitch Embed:', error);
    }

    return () => {
      if (embedInstanceRef.current) {
        embedInstanceRef.current.destroy();
        embedInstanceRef.current = null;
      }
    };
  }, [channel, showChat, isScriptLoaded, streamInfo?.is_live]);

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

  return (
    <div className="relative w-full aspect-video bg-black">
      <div
        id={`twitch-embed-${channel}`}
        ref={embedRef}
        className="w-full h-full"
      />
      <LiveIndicator viewerCount={streamInfo.viewer_count} />
    </div>
  );
}
