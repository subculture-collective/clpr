import { useQuery } from '@tanstack/react-query';
import { Link } from 'react-router-dom';
import {
  TrendingUp,
  Users,
  MessageSquare,
  FileText,
  Award,
  ArrowLeft,
  Activity,
  Clock,
  Calendar,
  AlertCircle
} from 'lucide-react';
import { Container, SEO } from '@/components';
import { forumApi } from '@/lib/forum-api';
import { ThreadCard } from '@/components/forum/ThreadCard';
import type { ForumThread, UserContribution, HelpfulReply } from '@/types/forum';

export function ForumAnalyticsPage() {
  // Fetch analytics data
  const { data: analytics, isLoading: analyticsLoading, error: analyticsError } = useQuery({
    queryKey: ['forum-analytics'],
    queryFn: () => forumApi.getAnalytics(),
    // Refetch every 5 minutes to reduce backend load while keeping data reasonably fresh
    refetchInterval: 300000,
    // Consider data fresh for 5 minutes to avoid unnecessary refetches on navigation/focus
    staleTime: 300000,
  });

  // Fetch popular discussions
  const { data: popular, isLoading: popularLoading, error: popularError } = useQuery({
    queryKey: ['forum-popular', 'week'],
    queryFn: () => forumApi.getPopularDiscussions('week', 10),
    staleTime: 300000,
  });

  // Fetch helpful replies
  const { data: helpful, isLoading: helpfulLoading, error: helpfulError } = useQuery({
    queryKey: ['forum-helpful', 'month'],
    queryFn: () => forumApi.getMostHelpfulReplies('month', 10),
    staleTime: 300000,
  });

  const analyticsData = analytics?.data;
  const popularThreads = popular?.data || [];
  const helpfulReplies = helpful?.data || [];

  const hasError = analyticsError || popularError || helpfulError;

  return (
    <>
      <SEO
        title="Forum Analytics"
        description="Discover trending topics, popular discussions, and top contributors in the Clipper community forum"
      />
      <Container className="py-6">
        <div className="max-w-7xl mx-auto">
          {/* Header */}
          <Link
            to="/forum"
            className="inline-flex items-center gap-2 text-muted-foreground hover:text-white mb-6 transition-colors"
          >
            <ArrowLeft className="w-4 h-4" />
            <span>Back to Forum</span>
          </Link>

          <div className="flex items-center justify-between mb-8">
            <div>
              <h1 className="text-3xl font-bold text-white mb-2">Forum Analytics</h1>
              <p className="text-muted-foreground">
                Community insights and trending discussions
              </p>
            </div>
          </div>

          {/* Error State */}
          {hasError && (
            <div className="bg-red-900/20 border border-red-800 rounded-lg p-6 mb-8">
              <div className="flex items-start gap-3">
                <AlertCircle className="w-5 h-5 text-red-400 shrink-0 mt-0.5" />
                <div>
                  <h3 className="text-red-400 font-semibold mb-1">Failed to load analytics</h3>
                  <p className="text-foreground text-sm">
                    {analyticsError && 'Unable to fetch analytics data. '}
                    {popularError && 'Unable to fetch popular discussions. '}
                    {helpfulError && 'Unable to fetch helpful replies. '}
                    Please try refreshing the page.
                  </p>
                </div>
              </div>
            </div>
          )}

          {/* Loading State */}
          {analyticsLoading && (
            <div className="text-center py-12">
              <div className="inline-block animate-spin rounded-full h-8 w-8 border-b-2 border-primary-500"></div>
              <p className="text-muted-foreground mt-4">Loading analytics...</p>
            </div>
          )}

          {/* Stats Grid */}
          {!analyticsLoading && analyticsData && (
            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4 mb-8">
              {/* Total Threads */}
              <div className="bg-background border border-border rounded-lg p-6">
                <div className="flex items-center gap-3 mb-3">
                  <div className="p-2 bg-primary-500/10 rounded-lg">
                    <FileText className="w-5 h-5 text-primary-400" />
                  </div>
                  <h3 className="text-sm font-medium text-muted-foreground uppercase tracking-wider">
                    Total Threads
                  </h3>
                </div>
                <p className="text-3xl font-bold text-white">{analyticsData.total_threads.toLocaleString()}</p>
              </div>

              {/* Total Replies */}
              <div className="bg-background border border-border rounded-lg p-6">
                <div className="flex items-center gap-3 mb-3">
                  <div className="p-2 bg-green-500/10 rounded-lg">
                    <MessageSquare className="w-5 h-5 text-green-400" />
                  </div>
                  <h3 className="text-sm font-medium text-muted-foreground uppercase tracking-wider">
                    Total Replies
                  </h3>
                </div>
                <p className="text-3xl font-bold text-white">{analyticsData.total_replies.toLocaleString()}</p>
              </div>

              {/* Active Users */}
              <div className="bg-background border border-border rounded-lg p-6">
                <div className="flex items-center gap-3 mb-3">
                  <div className="p-2 bg-purple-500/10 rounded-lg">
                    <Users className="w-5 h-5 text-purple-400" />
                  </div>
                  <h3 className="text-sm font-medium text-muted-foreground uppercase tracking-wider">
                    Total Users
                  </h3>
                </div>
                <p className="text-3xl font-bold text-white">{analyticsData.total_users.toLocaleString()}</p>
              </div>

              {/* Posts Today */}
              <div className="bg-background border border-border rounded-lg p-6">
                <div className="flex items-center gap-3 mb-3">
                  <div className="p-2 bg-orange-500/10 rounded-lg">
                    <Activity className="w-5 h-5 text-orange-400" />
                  </div>
                  <h3 className="text-sm font-medium text-muted-foreground uppercase tracking-wider">
                    Posts Today
                  </h3>
                </div>
                <p className="text-3xl font-bold text-white">{analyticsData.posts_today.toLocaleString()}</p>
              </div>
            </div>
          )}

          {/* Activity Stats Row */}
          {!analyticsLoading && analyticsData && (
            <div className="grid grid-cols-1 md:grid-cols-3 gap-4 mb-8">
              {/* This Week */}
              <div className="bg-background border border-border rounded-lg p-6">
                <div className="flex items-center gap-3 mb-3">
                  <Clock className="w-5 h-5 text-primary-400" />
                  <h3 className="text-sm font-medium text-muted-foreground uppercase tracking-wider">
                    This Week
                  </h3>
                </div>
                <div className="space-y-2">
                  <div className="flex justify-between">
                    <span className="text-muted-foreground">Posts:</span>
                    <span className="text-white font-semibold">{analyticsData.posts_this_week}</span>
                  </div>
                  <div className="flex justify-between">
                    <span className="text-muted-foreground">Active Users:</span>
                    <span className="text-white font-semibold">{analyticsData.active_users_week}</span>
                  </div>
                </div>
              </div>

              {/* This Month */}
              <div className="bg-background border border-border rounded-lg p-6">
                <div className="flex items-center gap-3 mb-3">
                  <Calendar className="w-5 h-5 text-green-400" />
                  <h3 className="text-sm font-medium text-muted-foreground uppercase tracking-wider">
                    This Month
                  </h3>
                </div>
                <div className="space-y-2">
                  <div className="flex justify-between">
                    <span className="text-muted-foreground">Posts:</span>
                    <span className="text-white font-semibold">{analyticsData.posts_this_month}</span>
                  </div>
                </div>
              </div>

              {/* Today's Activity */}
              <div className="bg-background border border-border rounded-lg p-6">
                <div className="flex items-center gap-3 mb-3">
                  <Activity className="w-5 h-5 text-orange-400" />
                  <h3 className="text-sm font-medium text-muted-foreground uppercase tracking-wider">
                    Today
                  </h3>
                </div>
                <div className="space-y-2">
                  <div className="flex justify-between">
                    <span className="text-muted-foreground">Posts:</span>
                    <span className="text-white font-semibold">{analyticsData.posts_today}</span>
                  </div>
                  <div className="flex justify-between">
                    <span className="text-muted-foreground">Active Users:</span>
                    <span className="text-white font-semibold">{analyticsData.active_users_today}</span>
                  </div>
                </div>
              </div>
            </div>
          )}

          <div className="grid grid-cols-1 lg:grid-cols-2 gap-8">
            {/* Trending Topics */}
            {!analyticsLoading && analyticsData && (
              <div className="bg-background border border-border rounded-lg p-6">
                <div className="flex items-center gap-3 mb-6">
                  <TrendingUp className="w-5 h-5 text-orange-400" />
                  <h2 className="text-xl font-semibold text-white">Trending Topics</h2>
                </div>
                {analyticsData.trending_topics && analyticsData.trending_topics.length > 0 ? (
                  <div className="flex flex-wrap gap-2">
                    {analyticsData.trending_topics.map((topic: string) => (
                      <Link
                        key={topic}
                        to={`/forum?tags=${encodeURIComponent(topic)}`}
                        className="px-3 py-1.5 bg-surface hover:bg-surface-hover border border-subtle rounded-full text-sm text-foreground hover:text-white transition-colors"
                      >
                        #{topic}
                      </Link>
                    ))}
                  </div>
                ) : (
                  <p className="text-muted-foreground">No trending topics yet</p>
                )}
              </div>
            )}

            {/* Top Contributors */}
            {!analyticsLoading && analyticsData && analyticsData.top_contributors && (
              <div className="bg-background border border-border rounded-lg p-6">
                <div className="flex items-center gap-3 mb-6">
                  <Award className="w-5 h-5 text-yellow-400" />
                  <h2 className="text-xl font-semibold text-white">Top Contributors</h2>
                </div>
                {analyticsData.top_contributors.length > 0 ? (
                  <div className="space-y-3">
                    {analyticsData.top_contributors.map((contributor: UserContribution, index: number) => (
                      <div
                        key={contributor.user_id}
                        className="flex items-center justify-between p-3 bg-surface rounded-lg"
                      >
                        <div className="flex items-center gap-3">
                          <span className="text-muted-foreground font-semibold w-6 text-center">
                            #{index + 1}
                          </span>
                          <div>
                            <p className="text-white font-medium">{contributor.username}</p>
                            <p className="text-sm text-muted-foreground">
                              {contributor.thread_count} threads, {contributor.reply_count} replies
                            </p>
                          </div>
                        </div>
                        <div className="text-right">
                          <p className="text-sm text-muted-foreground">Reputation</p>
                          <p className="text-white font-semibold">{contributor.reputation_score}</p>
                        </div>
                      </div>
                    ))}
                  </div>
                ) : (
                  <p className="text-muted-foreground">No contributors yet</p>
                )}
              </div>
            )}
          </div>

          {/* Popular Discussions */}
          {!popularLoading && popularThreads.length > 0 && (
            <div className="mt-8">
              <h2 className="text-2xl font-bold text-white mb-6">Popular This Week</h2>
              <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                {popularThreads.map((thread: ForumThread) => (
                  <ThreadCard key={thread.id} thread={thread} />
                ))}
              </div>
            </div>
          )}

          {/* Most Helpful Replies */}
          {!helpfulLoading && helpfulReplies.length > 0 && (
            <div className="mt-8">
              <h2 className="text-2xl font-bold text-white mb-6">Most Helpful Replies This Month</h2>
              <div className="space-y-4">
                {helpfulReplies.map((reply: HelpfulReply) => (
                  <div
                    key={reply.id}
                    className="bg-background border border-border rounded-lg p-6 hover:border-subtle transition-colors"
                  >
                    <div className="flex items-start justify-between mb-3">
                      <div className="flex-1">
                        <Link
                          to={`/forum/threads/${reply.thread_id}`}
                          className="text-lg font-semibold text-white hover:text-primary-400 transition-colors"
                        >
                          {reply.thread_title}
                        </Link>
                        <p className="text-sm text-muted-foreground mt-1">
                          by {reply.username}
                        </p>
                      </div>
                      <div className="flex flex-col items-end gap-1">
                        <div className="flex items-center gap-2">
                          <span className="text-green-400 font-semibold">
                            +{reply.upvotes}
                          </span>
                          <span className="text-red-400 font-semibold">
                            -{reply.downvotes}
                          </span>
                        </div>
                        <span className={`text-sm font-semibold ${
                          reply.net_votes > 0 ? 'text-green-400' :
                          reply.net_votes < 0 ? 'text-red-400' :
                          'text-muted-foreground'
                        }`}>
                          Net: {reply.net_votes}
                        </span>
                      </div>
                    </div>
                    <p className="text-foreground line-clamp-3">{reply.content}</p>
                  </div>
                ))}
              </div>
            </div>
          )}
        </div>
      </Container>
    </>
  );
}
