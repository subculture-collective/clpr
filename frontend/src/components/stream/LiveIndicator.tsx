interface LiveIndicatorProps {
  viewerCount: number;
}

function formatViewerCount(count: number): string {
  if (count >= 1000000) {
    return `${(count / 1000000).toFixed(1)}M`;
  } else if (count >= 1000) {
    return `${(count / 1000).toFixed(1)}K`;
  }
  return count.toString();
}

// Rendered beside the stream title, never over the player (Twitch forbids covering embeds).
export function LiveIndicator({ viewerCount }: LiveIndicatorProps) {
  return (
    <div className="inline-flex items-center gap-2 rounded-md bg-red-600 px-3 py-1.5">
      {/* Pulsing Live Dot */}
      <span className="relative flex h-2 w-2">
        <span className="animate-ping absolute inline-flex h-full w-full rounded-full bg-white opacity-75"></span>
        <span className="relative inline-flex rounded-full h-2 w-2 bg-white"></span>
      </span>

      {/* LIVE Text */}
      <span className="font-bold text-sm text-white uppercase">LIVE</span>

      {/* Viewer Count */}
      <span className="text-xs text-white/90 font-medium">
        {formatViewerCount(viewerCount)} viewers
      </span>
    </div>
  );
}
