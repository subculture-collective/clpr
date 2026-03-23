import { Link } from 'react-router-dom';
import { ExternalLink, Film } from 'lucide-react';
import { Button } from '../ui';
import type { StreamInfo } from '../../lib/stream-api';

interface StreamOfflineScreenProps {
  channel: string;
  streamInfo?: StreamInfo | null;
}

function formatRelativeTime(dateString?: string | null): string {
  if (!dateString) return '';

  const date = new Date(dateString);
  const now = new Date();
  const diffMs = now.getTime() - date.getTime();
  const diffMins = Math.floor(diffMs / 60000);
  const diffHours = Math.floor(diffMins / 60);
  const diffDays = Math.floor(diffHours / 24);

  if (diffMins < 60) {
    return `${diffMins} minute${diffMins !== 1 ? 's' : ''} ago`;
  } else if (diffHours < 24) {
    return `${diffHours} hour${diffHours !== 1 ? 's' : ''} ago`;
  } else if (diffDays < 7) {
    return `${diffDays} day${diffDays !== 1 ? 's' : ''} ago`;
  } else {
    return date.toLocaleDateString();
  }
}

export function StreamOfflineScreen({ channel, streamInfo }: StreamOfflineScreenProps) {
  return (
    <div className="w-full aspect-video bg-gradient-to-br from-background via-surface-raised to-background flex flex-col items-center justify-center text-center p-8">
      <div className="max-w-2xl">
        {/* Offline Icon */}
        <div className="text-6xl mb-6 opacity-50">📺</div>

        {/* Offline Message */}
        <h2 className="text-3xl font-bold mb-3 text-white">
          {channel} is offline
        </h2>

        {/* Last Streamed Info */}
        {streamInfo?.last_went_offline ? (
          <p className="text-muted-foreground mb-8 text-lg">
            Last streamed {formatRelativeTime(streamInfo.last_went_offline)}
          </p>
        ) : (
          <p className="text-muted-foreground mb-8 text-lg">
            Check back later for the next stream
          </p>
        )}

        {/* Action Buttons */}
        <div className="flex flex-col sm:flex-row gap-3 justify-center items-center">
          <Button
            variant="primary"
            size="lg"
            onClick={() => {
              window.open(`https://twitch.tv/${channel}`, '_blank', 'noopener,noreferrer');
            }}
          >
            <ExternalLink size={16} strokeWidth={1.75} className="mr-2" />
            Visit on Twitch
          </Button>

          <Link to={`/clips?streamer=${channel}`}>
            <Button variant="secondary" size="lg">
              <Film size={16} strokeWidth={1.75} className="mr-2" />
              View Recent Clips
            </Button>
          </Link>
        </div>

        {/* Additional Info */}
        <div className="mt-8 text-sm text-text-secondary">
          <p>This page will automatically update when {channel} goes live</p>
        </div>
      </div>
    </div>
  );
}
