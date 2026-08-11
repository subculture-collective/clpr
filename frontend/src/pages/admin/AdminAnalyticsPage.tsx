import React, { useState } from 'react';
import { useQuery } from '@tanstack/react-query';
import { Helmet } from '@dr.pogodin/react-helmet';
import {
  getPlatformOverview,
  getContentMetrics,
  getPlatformTrends,
} from '../../lib/analytics-api';
import {
  MetricCard,
  LineChartComponent,
  BarChartComponent,
  DateRangeSelector,
} from '../../components/analytics';

const AdminAnalyticsPage: React.FC = () => {
  const [timeRange, setTimeRange] = useState(30);

  // Fetch platform overview
  const { data: overview, isLoading: overviewLoading } = useQuery({
    queryKey: ['platformOverview'],
    queryFn: getPlatformOverview,
  });

  // Fetch content metrics
  const { data: content, isLoading: contentLoading } = useQuery({
    queryKey: ['contentMetrics'],
    queryFn: getContentMetrics,
  });

  // Fetch user growth trend
  const { data: usersTrend, isLoading: usersTrendLoading } = useQuery({
    queryKey: ['platformTrends', 'users', timeRange],
    queryFn: () => getPlatformTrends({ metric: 'users', days: timeRange }),
  });

  // Fetch clips growth trend
  const { data: clipsTrend, isLoading: clipsTrendLoading } = useQuery({
    queryKey: ['platformTrends', 'clips', timeRange],
    queryFn: () => getPlatformTrends({ metric: 'clips', days: timeRange }),
  });

  return (
    <>
      <Helmet>
        <title>Platform Analytics - Admin Dashboard</title>
      </Helmet>

      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        {/* Header */}
        <div className="mb-8">
          <h1 className="text-3xl font-bold text-gray-900 dark:text-white">
            Platform Analytics
          </h1>
          <p className="mt-2 text-gray-600 dark:text-gray-400">
            Monitor platform health and user engagement
          </p>
        </div>

        {/* Overview KPIs */}
        <div className="mb-8">
          <h2 className="text-xl font-bold text-gray-900 dark:text-white mb-4">
            Overview Metrics
          </h2>

          {overviewLoading ? (
            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
              {[1, 2, 3, 4, 5, 6, 7, 8].map((i) => (
                <div
                  key={i}
                  className="bg-white dark:bg-gray-800 rounded-lg shadow p-6 animate-pulse"
                >
                  <div className="h-4 bg-gray-200 dark:bg-gray-700 rounded w-1/2 mb-4"></div>
                  <div className="h-8 bg-gray-200 dark:bg-gray-700 rounded w-3/4"></div>
                </div>
              ))}
            </div>
          ) : overview ? (
            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
              <MetricCard
                title="Total Users"
                value={overview.total_users}
                subtitle="Registered accounts"
              />
              <MetricCard
                title="Daily Active Users"
                value={overview.active_users_daily}
                subtitle="Active today"
              />
              <MetricCard
                title="Monthly Active Users"
                value={overview.active_users_monthly}
                subtitle="Active this month"
              />
              <MetricCard
                title="Total Clips"
                value={overview.total_clips}
                subtitle="All clips"
              />
              <MetricCard
                title="Clips Added Today"
                value={overview.clips_added_today}
                subtitle="New clips"
              />
              <MetricCard
                title="Total Votes"
                value={overview.total_votes}
                subtitle="All time"
              />
              <MetricCard
                title="Total Comments"
                value={overview.total_comments}
                subtitle="All time"
              />
              <MetricCard
                title="Avg. Session Duration"
                value={`${overview.avg_session_duration.toFixed(1)} min`}
                subtitle="User engagement"
              />
            </div>
          ) : null}
        </div>

        {/* Growth Metrics */}
        <div className="mb-8">
          <div className="flex items-center justify-between mb-4">
            <h2 className="text-xl font-bold text-gray-900 dark:text-white">
              Growth Trends
            </h2>
            <DateRangeSelector value={timeRange} onChange={setTimeRange} />
          </div>

          <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
            {usersTrendLoading ? (
              <div className="bg-white dark:bg-gray-800 rounded-lg shadow p-6 h-80 animate-pulse">
                <div className="h-6 bg-gray-200 dark:bg-gray-700 rounded w-1/3 mb-4"></div>
                <div className="h-full bg-gray-200 dark:bg-gray-700 rounded"></div>
              </div>
            ) : usersTrend?.data ? (
              <LineChartComponent
                data={usersTrend.data}
                title="New Users Over Time"
                valueLabel="New Users"
                color="#10b981"
              />
            ) : null}

            {clipsTrendLoading ? (
              <div className="bg-white dark:bg-gray-800 rounded-lg shadow p-6 h-80 animate-pulse">
                <div className="h-6 bg-gray-200 dark:bg-gray-700 rounded w-1/3 mb-4"></div>
                <div className="h-full bg-gray-200 dark:bg-gray-700 rounded"></div>
              </div>
            ) : clipsTrend?.data ? (
              <LineChartComponent
                data={clipsTrend.data}
                title="New Clips Over Time"
                valueLabel="New Clips"
                color="#8b5cf6"
              />
            ) : null}
          </div>
        </div>

        {/* Content Metrics */}
        <div className="mb-8">
          <h2 className="text-xl font-bold text-gray-900 dark:text-white mb-4">
            Content Metrics
          </h2>

          {contentLoading ? (
            <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
              {[1, 2].map((i) => (
                <div
                  key={i}
                  className="bg-white dark:bg-gray-800 rounded-lg shadow p-6 h-80 animate-pulse"
                >
                  <div className="h-6 bg-gray-200 dark:bg-gray-700 rounded w-1/3 mb-4"></div>
                  <div className="h-full bg-gray-200 dark:bg-gray-700 rounded"></div>
                </div>
              ))}
            </div>
          ) : content ? (
            <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
              {content.most_popular_games && (
                <BarChartComponent
                  data={content.most_popular_games.map((game) => ({
                    name: game.game_name,
                    value: game.clip_count,
                  }))}
                  title="Most Popular Twitch Categories"
                  valueLabel="Clip Count"
                  color="#3b82f6"
                  height={350}
                />
              )}

              {content.most_popular_creators && (
                <BarChartComponent
                  data={content.most_popular_creators.map((creator) => ({
                    name: creator.creator_name,
                    value: creator.view_count,
                  }))}
                  title="Most Popular Creators"
                  valueLabel="Views"
                  color="#ec4899"
                  height={350}
                />
              )}
            </div>
          ) : null}
        </div>

        {/* Trending Tags */}
        {content?.trending_tags && content.trending_tags.length > 0 && (
          <div className="mb-8">
            <h2 className="text-xl font-bold text-gray-900 dark:text-white mb-4">
              Trending Tags (Last 7 Days)
            </h2>

            <div className="bg-white dark:bg-gray-800 rounded-lg shadow p-6">
              <div className="flex flex-wrap gap-2">
                {content.trending_tags.map((tag) => (
                  <span
                    key={tag.tag_id}
                    className="inline-flex items-center px-3 py-1 rounded-full text-sm font-medium bg-purple-100 dark:bg-purple-900 text-purple-800 dark:text-purple-200"
                  >
                    {tag.tag_name}
                    <span className="ml-2 text-xs text-purple-600 dark:text-purple-400">
                      {tag.usage_count}
                    </span>
                  </span>
                ))}
              </div>
            </div>
          </div>
        )}
      </div>
    </>
  );
};

export default AdminAnalyticsPage;
