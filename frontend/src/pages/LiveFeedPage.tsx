import { useQuery } from '@tanstack/react-query';
import { Link } from 'react-router-dom';
import { formatDistanceToNow } from 'date-fns';
import { ExternalLink, Eye, PlayCircle } from 'lucide-react';
import { fetchFollowedLiveBroadcasters } from '@/lib/broadcaster-api';
import { SEO } from '@/components/SEO';
import { Spinner } from '@/components';
import { LiveBadge } from '@/components/broadcaster';

/**
 * LiveFeedPage displays all live broadcasters that the authenticated user follows
 */
export function LiveFeedPage() {
  const {
    data: liveData,
    isLoading,
    error,
  } = useQuery({
    queryKey: ['followed-live-broadcasters'],
    queryFn: fetchFollowedLiveBroadcasters,
    refetchInterval: 30000, // Refetch every 30 seconds
  });

  const liveBroadcasters = liveData?.data || [];

  return (
    <>
      <SEO
        title="Live Now"
        description="Watch live streams from broadcasters you follow"
      />

      <div className="container mx-auto px-4 py-8">
        <div className="mb-8">
          <h1 className="text-3xl font-bold mb-2">Live Now</h1>
          <p className="text-muted-foreground">
            Broadcasters you follow who are currently streaming
          </p>
        </div>

        {isLoading && (
          <div className="flex justify-center items-center py-12">
            <Spinner size="xl" />
          </div>
        )}

        {error && (
          <div className="bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 rounded-lg p-4">
            <p className="text-red-800 dark:text-red-200">
              Failed to load live broadcasters. Please try again.
            </p>
          </div>
        )}

        {!isLoading && !error && liveBroadcasters.length === 0 && (
          <div className="bg-surface border border-border rounded-lg p-8 text-center">
            <PlayCircle className="mx-auto mb-4 text-muted-foreground" size={48} />
            <h3 className="text-xl font-semibold mb-2">No live streams</h3>
            <p className="text-muted-foreground mb-4">
              None of the broadcasters you follow are currently live.
            </p>
            <Link
              to="/discover"
              className="inline-block bg-purple-600 hover:bg-purple-700 text-white px-6 py-2 rounded-lg transition-colors"
            >
              Discover Broadcasters
            </Link>
          </div>
        )}

        {!isLoading && !error && liveBroadcasters.length > 0 && (
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
            {liveBroadcasters.map((broadcaster) => (
              <div
                key={broadcaster.broadcaster_id}
                className="bg-surface border border-border rounded-lg overflow-hidden hover:shadow-lg transition-shadow"
              >
                <div className="relative aspect-video bg-background">
                  <div className="absolute inset-0 flex items-center justify-center">
                    <PlayCircle className="text-white/50" size={64} />
                  </div>
                  <div className="absolute top-3 left-3">
                    <LiveBadge
                      liveStatus={broadcaster}
                      showViewers={true}
                      showDuration={false}
                      size="md"
                    />
                  </div>
                </div>

                <div className="p-4">
                  <div className="flex items-start justify-between gap-3 mb-2">
                    <div className="flex-1 min-w-0">
                      <Link
                        to={`/broadcaster/${broadcaster.broadcaster_id}`}
                        className="text-lg font-semibold hover:text-purple-600 dark:hover:text-purple-400 transition-colors line-clamp-1"
                      >
                        {broadcaster.user_name || broadcaster.user_login || broadcaster.broadcaster_id}
                      </Link>
                    </div>
                    <a
                      href={`https://twitch.tv/${broadcaster.user_login || broadcaster.broadcaster_id}`}
                      target="_blank"
                      rel="noopener noreferrer"
                      className="flex-shrink-0 text-purple-600 dark:text-purple-400 hover:text-purple-700 dark:hover:text-purple-300 transition-colors"
                      title="Watch on Twitch"
                    >
                      <ExternalLink size={20} />
                    </a>
                  </div>

                  {broadcaster.stream_title && (
                    <p className="text-foreground mb-2 line-clamp-2">
                      {broadcaster.stream_title}
                    </p>
                  )}

                  {broadcaster.game_name && (
                    <p className="text-sm text-muted-foreground mb-3">
                      Playing {broadcaster.game_name}
                    </p>
                  )}

                  <div className="flex items-center gap-4 text-sm text-muted-foreground">
                    <div className="flex items-center gap-1">
                      <Eye size={16} />
                      <span>
                        {broadcaster.viewer_count.toLocaleString()} viewers
                      </span>
                    </div>
                    {(() => {
                      if (!broadcaster.started_at) return null;
                      const startedAtDate = new Date(broadcaster.started_at);
                      if (isNaN(startedAtDate.getTime())) return null;
                      return (
                        <span>
                          Started{' '}
                          {formatDistanceToNow(startedAtDate, {
                            addSuffix: true,
                          })}
                        </span>
                      );
                    })()}
                  </div>
                </div>
              </div>
            ))}
          </div>
        )}
      </div>
    </>
  );
}
