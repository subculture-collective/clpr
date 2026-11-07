import { useEffect, useRef, useState } from 'react';
import Hls from 'hls.js';
import type { VideoQuality } from '@/lib/adaptive-bitrate';
import { AdaptiveBitrateSelector } from '@/lib/adaptive-bitrate';

export interface HlsPlayerProps {
  src: string;
  quality: VideoQuality;
  autoQuality: boolean;
  videoRef?: React.RefObject<HTMLVideoElement | null>;
  onQualityChange?: (quality: VideoQuality) => void;
  onBandwidthUpdate?: (bandwidth: number) => void;
  onBufferHealthUpdate?: (health: number) => void;
  className?: string;
  /** WebVTT captions shown as a selectable caption track. */
  captionsUrl?: string;
  captionsLanguage?: string;
}

/**
 * HLS video player component with adaptive bitrate support
 * Handles HLS stream playback, quality selection, and network monitoring
 */
export function HlsPlayer({
  src,
  quality,
  autoQuality,
  videoRef: externalVideoRef,
  onQualityChange,
  onBandwidthUpdate,
  onBufferHealthUpdate,
  className,
  captionsUrl,
  captionsLanguage,
}: HlsPlayerProps) {
  const internalVideoRef = useRef<HTMLVideoElement>(null);
  const videoRef = externalVideoRef || internalVideoRef;
  const hlsRef = useRef<Hls | null>(null);
  const bitrateSelector = useRef(new AdaptiveBitrateSelector());
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  // Initialize HLS player
  useEffect(() => {
    const video = videoRef.current;
    if (!video || !src) return;

    // Check if HLS is supported
    if (!Hls.isSupported()) {
      // For Safari and other browsers with native HLS support
      if (video.canPlayType('application/vnd.apple.mpegurl')) {
        video.src = src;
        queueMicrotask(() => setIsLoading(false));
      } else {
        queueMicrotask(() => {
          setError('HLS is not supported in this browser');
          setIsLoading(false);
        });
      }
      return;
    }

    // Initialize HLS.js
    const hls = new Hls({
      enableWorker: true,
      lowLatencyMode: false,
      backBufferLength: 90,
    });

    hlsRef.current = hls;

    // Load source
    hls.loadSource(src);
    hls.attachMedia(video);

    // Handle events
    hls.on(Hls.Events.MANIFEST_PARSED, () => {
      queueMicrotask(() => {
        setIsLoading(false);
        setError(null);
      });
      video.play().catch(() => {
        // Autoplay might be prevented, that's okay
      });
    });

    hls.on(Hls.Events.ERROR, (_event, data) => {
      if (data.fatal) {
        switch (data.type) {
          case Hls.ErrorTypes.NETWORK_ERROR:
            queueMicrotask(() => setError('Network error occurred'));
            hls.startLoad();
            break;
          case Hls.ErrorTypes.MEDIA_ERROR:
            queueMicrotask(() => setError('Media error occurred'));
            hls.recoverMediaError();
            break;
          default:
            queueMicrotask(() => setError('Fatal error occurred'));
            hls.destroy();
            break;
        }
      }
    });

    // Monitor bandwidth for adaptive quality
    hls.on(Hls.Events.FRAG_LOADED, (_event, data) => {
      const stats = data.frag.stats;
      if (stats && stats.loaded && stats.total) {
        const loadTimeSeconds = (stats.loading.end - stats.loading.start) / 1000;
        const bandwidthMbps = (stats.loaded * 8) / (loadTimeSeconds * 1000000);
        onBandwidthUpdate?.(bandwidthMbps);

        // Auto quality selection
        if (autoQuality) {
          let bufferLength = 0;
          try {
            if (video.buffered && video.buffered.length > 0) {
              bufferLength = video.buffered.end(0) - video.currentTime;
            }
          } catch {
            // In case accessing buffered ranges throws, fall back to 0 buffer length
            bufferLength = 0;
          }
          const bufferHealth = Math.min(100, (bufferLength / 30) * 100);
          onBufferHealthUpdate?.(bufferHealth);

          const selectedQuality = bitrateSelector.current.selectQuality(
            bandwidthMbps,
            bufferHealth
          );
          onQualityChange?.(selectedQuality);
        }
      }
    });

    // Cleanup
    return () => {
      if (hls) {
        hls.destroy();
        hlsRef.current = null;
      }
    };
  }, [src, autoQuality, onQualityChange, onBandwidthUpdate, onBufferHealthUpdate, videoRef]);

  // Handle manual quality selection
  useEffect(() => {
    const hls = hlsRef.current;
    if (!hls || autoQuality) return;

    const levels = hls.levels;
    if (!levels || levels.length === 0) {
      return;
    }

    // Map desired quality labels to target video heights
    const targetHeightMap: Record<string, number> = {
      '480p': 480,
      '720p': 720,
      '1080p': 1080,
      '2K': 1440,
      '4K': 2160,
    };

    if (quality === 'auto') {
      // Let HLS.js handle adaptive bitrate selection
      hls.currentLevel = -1;
      return;
    }

    const targetHeight = targetHeightMap[quality];
    if (typeof targetHeight !== 'number') {
      // Unknown quality label; do not change the current level
      return;
    }

    // Find the level with the closest matching resolution
    let bestIndex = -1;
    let bestDiff = Number.POSITIVE_INFINITY;

    levels.forEach((level, index) => {
      const height = level.height;
      if (typeof height !== 'number' || height <= 0) {
        return;
      }
      const diff = Math.abs(height - targetHeight);
      if (diff < bestDiff) {
        bestDiff = diff;
        bestIndex = index;
      }
    });

    if (bestIndex !== -1) {
      hls.currentLevel = bestIndex;
    }
  }, [quality, autoQuality]);

  if (error) {
    return (
      <div className="absolute inset-0 flex items-center justify-center bg-black text-white">
        <div className="text-center">
          <svg
            className="w-16 h-16 mx-auto mb-4 text-error-500"
            fill="none"
            stroke="currentColor"
            viewBox="0 0 24 24"
          >
            <path
              strokeLinecap="round"
              strokeLinejoin="round"
              strokeWidth={2}
              d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"
            />
          </svg>
          <p className="text-lg font-semibold mb-2">Playback Error</p>
          <p className="text-sm text-white/70">{error}</p>
        </div>
      </div>
    );
  }

  return (
    <>
      <video
        ref={videoRef}
        className={className}
        playsInline
        preload="metadata"
      >
        {captionsUrl && (
          <track
            kind="captions"
            src={captionsUrl}
            srcLang={captionsLanguage || 'en'}
            label="Captions"
            default
          />
        )}
        Your browser does not support the video tag.
      </video>
      {isLoading && (
        <div className="absolute inset-0 flex items-center justify-center bg-black">
          <div className="flex flex-col items-center gap-3">
            <div className="w-12 h-12 border-4 border-white/20 border-t-white rounded-full animate-spin" />
            <p className="text-white text-sm">Loading video...</p>
          </div>
        </div>
      )}
    </>
  );
}
