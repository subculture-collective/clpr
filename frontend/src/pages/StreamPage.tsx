import { Link, useParams } from 'react-router-dom';
import { useQuery } from '@tanstack/react-query';
import { Container, SEO } from '../components';
import { TwitchPlayer, ClipCreator, StreamFollowButton, TwitchChatEmbed } from '../components/stream';
import { fetchStreamStatus } from '../lib/stream-api';
import { ClipCard } from '../components/clip';
import { fetchBroadcasterClips } from '../lib/broadcaster-api';
import { useAuth } from '../context/AuthContext';
import { useState } from 'react';

export function StreamPage() {
  const { streamer } = useParams<{ streamer: string }>();
  const { isAuthenticated } = useAuth();
  const [chatPosition, setChatPosition] = useState<'side' | 'bottom'>('side');
  const [showChat, setShowChat] = useState(true);

  // Fetch stream status
  const { data: streamInfo } = useQuery({
    queryKey: ['streamStatus', streamer],
    queryFn: () => fetchStreamStatus(streamer!),
    enabled: !!streamer,
    refetchInterval: 60000, // Refresh every 60 seconds
  });

  // Fetch recent clips for the streamer
  const { data: clipsData, isLoading: isLoadingClips } = useQuery({
    queryKey: ['broadcasterClips', streamer],
    queryFn: () => fetchBroadcasterClips(streamer!, { page: 1, limit: 12, sort: 'recent' }),
    enabled: !!streamer,
  });

  if (!streamer) {
    return (
      <Container>
        <div className="text-center py-12">
          <p className="text-muted-foreground">Invalid stream URL</p>
        </div>
      </Container>
    );
  }

  return (
    <>
      <SEO
        title={`${streamer} - Live Stream`}
        description={
          streamInfo?.is_live
            ? `${streamer} is live! ${streamInfo.title || ''}`
            : `Watch ${streamer}'s live stream and recent clips`
        }
      />

      <div className="min-h-screen bg-background">
        {/* Stream Player and Chat Container */}
        <div className="w-full bg-black">
          <div className="max-w-[2000px] mx-auto">
            <div className={`flex ${chatPosition === 'side' ? 'flex-col lg:flex-row' : 'flex-col'} gap-0`}>
              {/* Stream Player */}
              <div className={chatPosition === 'side' && showChat ? 'lg:flex-[3]' : 'w-full'}>
                <TwitchPlayer channel={streamer} />
              </div>
              
              {/* Chat Sidebar/Bottom */}
              {showChat && (
                <div className={chatPosition === 'side' ? 'lg:flex-[1] lg:min-w-[340px] lg:max-w-[400px]' : 'w-full'}>
                  <TwitchChatEmbed channel={streamer} position={chatPosition} />
                </div>
              )}
            </div>
          </div>
        </div>

        {/* Stream Info */}
        <Container className="py-6">
          <div className="mb-8">
            <div className="flex items-center gap-4 mb-4 flex-wrap">
              <h1 className="text-3xl font-bold text-foreground">
                {streamer}
              </h1>
              {streamInfo?.is_live && (
                <span className="inline-flex items-center px-3 py-1 rounded-full text-sm font-medium bg-red-100 text-red-800 dark:bg-red-900 dark:text-red-200">
                  <span className="w-2 h-2 mr-2 bg-red-500 rounded-full animate-pulse"></span>
                  LIVE
                </span>
              )}
                {/* Follow Button */}
                <StreamFollowButton streamerUsername={streamer} />

                {/* Chat Controls */}
                <div className="ml-auto flex items-center gap-2 flex-wrap justify-end">
                  {isAuthenticated && (
                    <Link
                      to={`/streamer-tools/${encodeURIComponent(streamer)}/clips`}
                      className="px-3 py-1 text-sm rounded border border-purple-500 text-purple-700 dark:text-purple-300 hover:bg-purple-50 dark:hover:bg-purple-900/20"
                    >
                      Clip Room
                    </Link>
                  )}
                  <button
                    onClick={() => setShowChat(!showChat)}
                    className="px-3 py-1 text-sm rounded border border-border hover:bg-surface-hover text-foreground"
                  >
                    {showChat ? 'Hide Chat' : 'Show Chat'}
                  </button>
                  {showChat && (
                    <select
                      value={chatPosition}
                      onChange={(e) => setChatPosition(e.target.value as 'side' | 'bottom')}
                      className="px-3 py-1 text-sm rounded border border-border bg-surface text-foreground"
                    >
                      <option value="side">Side</option>
                      <option value="bottom">Bottom</option>
                    </select>
                  )}
                  {/* Create Clip Button - only show if stream is live and user is authenticated */}
                  {streamInfo?.is_live && isAuthenticated && (
                    <ClipCreator streamer={streamer} />
                  )}
                </div>
            </div>

            {streamInfo?.is_live && streamInfo.title && (
              <div className="space-y-2">
                <h2 className="text-xl text-foreground">
                  {streamInfo.title}
                </h2>
                {streamInfo.game_name && (
                  <p className="text-muted-foreground">
                    Playing: {streamInfo.game_name}
                  </p>
                )}
              </div>
            )}
          </div>

          {/* Recent Clips */}
          <div className="mt-8">
            <h2 className="text-2xl font-bold mb-6 text-foreground">
              Recent Clips
            </h2>

            {isLoadingClips ? (
              <div className="flex justify-center items-center py-12">
                <div className="inline-block animate-spin rounded-full h-12 w-12 border-t-2 border-b-2 border-purple-500"></div>
              </div>
            ) : clipsData?.data.length ? (
              <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-6">
                {clipsData.data.map((clip) => (
                  <ClipCard key={clip.id} clip={clip} />
                ))}
              </div>
            ) : (
              <div className="text-center py-12">
                <p className="text-muted-foreground">
                  No clips available yet
                </p>
              </div>
            )}
          </div>
        </Container>
      </div>
    </>
  );
}
