import { useClipById } from '@/hooks/useClips';
import { VideoPlayer } from '@/components/video';

export interface SyncedVideoPlayerProps {
  clipId: string | undefined;
  /**
   * Reserved for future synchronized playback support.
   * Currently unused because the Twitch embed cannot be controlled programmatically.
   */
  currentPosition?: number;
  /**
   * Reserved for future synchronized playback support.
   * Currently unused because the Twitch embed cannot be controlled programmatically.
   */
  isPlaying?: boolean;
  className?: string;
}

/**
 * Video player component for watch parties.
 * Fetches clip data and displays a Twitch-embedded video player.
 *
 * Note: The Twitch iframe embed handles its own playback and does not currently
 * support programmatic synchronization. The `currentPosition` and `isPlaying`
 * props are accepted for forward compatibility but are not yet used.
 * Future enhancement: Implement an HLS player with custom controls for full sync support.
 */
export function SyncedVideoPlayer({
  clipId,
  className = '',
}: SyncedVideoPlayerProps) {
  // Fetch clip data when clipId is available using the standard hook
  const { data: clip, isLoading, error } = useClipById(clipId || '');

  // No clip selected state
  if (!clipId) {
    return (
      <div className={`relative bg-surface-secondary rounded-lg aspect-video flex items-center justify-center ${className}`}>
        <div className="text-center text-content-secondary">
          <p className="text-lg mb-2">No Video Selected</p>
          <p className="text-sm">The host hasn't selected a video yet</p>
        </div>
      </div>
    );
  }

  // Loading state
  if (isLoading) {
    return (
      <div className={`relative bg-surface-secondary rounded-lg aspect-video flex items-center justify-center ${className}`}>
        <div className="text-center text-content-secondary">
          <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-primary-500 mx-auto mb-4"></div>
          <p className="text-sm">Loading video...</p>
        </div>
      </div>
    );
  }

  // Error state
  if (error || !clip) {
    return (
      <div className={`relative bg-surface-secondary rounded-lg aspect-video flex items-center justify-center ${className}`}>
        <div className="text-center text-content-secondary">
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
          <p className="text-lg font-semibold mb-2">Failed to Load Video</p>
          <p className="text-sm">
            {error instanceof Error ? error.message : 'Unable to load the video clip'}
          </p>
        </div>
      </div>
    );
  }

  // Render video player
  // Note: VideoPlayer uses Twitch iframe embed which handles its own playback.
  // Full synchronized playback control requires implementing an HLS player with
  // custom controls that can respond to WebSocket sync events.
  return (
    <div className={`relative ${className}`}>
      <VideoPlayer
        clipId={clip.id}
        title={clip.title}
        embedUrl={clip.embed_url}
        twitchClipId={clip.twitch_clip_id}
      />
    </div>
  );
}
