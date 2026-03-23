import React, { useState } from 'react';
import { useParams, Link } from 'react-router-dom';
import { useQuery } from '@tanstack/react-query';
import { Helmet } from '@dr.pogodin/react-helmet';
import {
  getCreatorAnalyticsOverview,
  getCreatorTopClips,
  getCreatorTrends,
} from '../lib/analytics-api';
import {
  MetricCard,
  LineChartComponent,
  DateRangeSelector,
  AudienceInsightsSection,
} from '../components/analytics';
import { useDebounce } from '../hooks/useDebounce';

const CreatorAnalyticsPage: React.FC = () => {
  const { creatorName } = useParams<{ creatorName: string }>();
  const [timeRange, setTimeRange] = useState(30);
  const [sortBy, setSortBy] = useState('views');
  
  // Debounce time range changes to avoid excessive API calls
  const debouncedTimeRange = useDebounce(timeRange, 300);

  // Fetch analytics overview
  const { data: overview, isLoading: overviewLoading } = useQuery({
    queryKey: ['creatorAnalyticsOverview', creatorName],
    queryFn: () => getCreatorAnalyticsOverview(creatorName!),
    enabled: !!creatorName,
  });

  // Fetch top clips
  const { data: topClips, isLoading: clipsLoading } = useQuery({
    queryKey: ['creatorTopClips', creatorName, sortBy],
    queryFn: () => getCreatorTopClips(creatorName!, { sort: sortBy, limit: 10 }),
    enabled: !!creatorName,
  });

  // Fetch views trend (using debounced time range)
  const { data: viewsTrend, isLoading: viewsTrendLoading } = useQuery({
    queryKey: ['creatorTrends', creatorName, 'clip_views', debouncedTimeRange],
    queryFn: () =>
      getCreatorTrends(creatorName!, { metric: 'clip_views', days: debouncedTimeRange }),
    enabled: !!creatorName,
  });

  // Fetch votes trend (using debounced time range)
  const { data: votesTrend, isLoading: votesTrendLoading } = useQuery({
    queryKey: ['creatorTrends', creatorName, 'votes', debouncedTimeRange],
    queryFn: () =>
      getCreatorTrends(creatorName!, { metric: 'votes', days: debouncedTimeRange }),
    enabled: !!creatorName,
  });

  if (!creatorName) {
    return (
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        <p className="text-red-600">Invalid creator name</p>
      </div>
    );
  }

  return (
    <>
      <Helmet>
        <title>{creatorName} Analytics - clpr</title>
      </Helmet>

      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        {/* Header */}
        <div className="mb-8">
          <h1 className="text-3xl font-bold text-foreground">
            {creatorName} Analytics
          </h1>
          <p className="mt-2 text-muted-foreground">
            Performance metrics and insights for clips
          </p>
        </div>

        {/* Overview Metrics */}
        {overviewLoading ? (
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6 mb-8">
            {[1, 2, 3, 4, 5, 6].map((i) => (
              <div
                key={i}
                className="bg-surface rounded-lg shadow p-6 animate-pulse"
              >
                <div className="h-4 bg-muted rounded w-1/2 mb-4"></div>
                <div className="h-8 bg-muted rounded w-3/4"></div>
              </div>
            ))}
          </div>
        ) : overview ? (
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6 mb-8">
            <MetricCard
              title="Total Clips"
              value={overview.total_clips}
              subtitle="Clips on platform"
            />
            <MetricCard
              title="Total Views"
              value={overview.total_views}
              subtitle="Across all clips"
            />
            <MetricCard
              title="Total Upvotes"
              value={overview.total_upvotes}
              subtitle="Community approval"
            />
            <MetricCard
              title="Total Comments"
              value={overview.total_comments}
              subtitle="User engagement"
            />
            <MetricCard
              title="Avg. Engagement Rate"
              value={`${(overview.avg_engagement_rate * 100).toFixed(2)}%`}
              subtitle="Votes + Comments / Views"
            />
            <MetricCard
              title="Follower Count"
              value={overview.follower_count}
              subtitle="Community size"
            />
          </div>
        ) : null}

        {/* Top Clips Section */}
        <div className="mb-8">
          <div className="flex items-center justify-between mb-4">
            <h2 className="text-2xl font-bold text-foreground">
              Top Performing Clips
            </h2>
            <label htmlFor="clip-sort" className="sr-only">
              Sort clips by
            </label>
            <select
              id="clip-sort"
              value={sortBy}
              onChange={(e) => setSortBy(e.target.value)}
              className="px-4 py-2 border border-border rounded-lg bg-surface text-foreground focus:outline-none focus:ring-2 focus:ring-primary-500"
              aria-label="Sort clips by metric"
            >
              <option value="views">By Views</option>
              <option value="votes">By Votes</option>
              <option value="comments">By Comments</option>
            </select>
          </div>

          {clipsLoading ? (
            <div className="space-y-4">
              {[1, 2, 3].map((i) => (
                <div
                  key={i}
                  className="bg-surface rounded-lg shadow p-4 animate-pulse"
                >
                  <div className="h-6 bg-muted rounded w-3/4"></div>
                </div>
              ))}
            </div>
          ) : topClips?.clips && topClips.clips.length > 0 ? (
            <div className="bg-surface rounded-lg shadow overflow-hidden">
              <table className="min-w-full divide-y divide-border">
                <caption className="sr-only">
                  Top performing clips sorted by {sortBy}
                </caption>
                <thead className="bg-background">
                  <tr>
                    <th scope="col" className="px-6 py-3 text-left text-xs font-medium text-muted-foreground uppercase tracking-wider">
                      Clip
                    </th>
                    <th scope="col" className="px-6 py-3 text-left text-xs font-medium text-muted-foreground uppercase tracking-wider">
                      Views
                    </th>
                    <th scope="col" className="px-6 py-3 text-left text-xs font-medium text-muted-foreground uppercase tracking-wider">
                      Votes
                    </th>
                    <th scope="col" className="px-6 py-3 text-left text-xs font-medium text-muted-foreground uppercase tracking-wider">
                      Comments
                    </th>
                    <th scope="col" className="px-6 py-3 text-left text-xs font-medium text-muted-foreground uppercase tracking-wider">
                      Engagement
                    </th>
                  </tr>
                </thead>
                <tbody className="bg-surface divide-y divide-border">
                  {topClips.clips.map((clip) => (
                    <tr key={clip.id} className="hover:bg-surface-hover">
                      <td className="px-6 py-4">
                        <Link
                          to={`/clips/${clip.id}`}
                          className="text-purple-600 dark:text-purple-400 hover:underline"
                        >
                          {clip.title}
                        </Link>
                      </td>
                      <td className="px-6 py-4 text-sm text-foreground">
                        {clip.views.toLocaleString()}
                      </td>
                      <td className="px-6 py-4 text-sm text-foreground">
                        {clip.vote_score}
                      </td>
                      <td className="px-6 py-4 text-sm text-foreground">
                        {clip.comment_count}
                      </td>
                      <td className="px-6 py-4 text-sm text-foreground">
                        {(clip.engagement_rate * 100).toFixed(2)}%
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          ) : (
            <p className="text-muted-foreground">No clips found</p>
          )}
        </div>

        {/* Performance Trends */}
        <div className="mb-8">
          <div className="flex items-center justify-between mb-4">
            <h2 className="text-2xl font-bold text-foreground">
              Performance Trends
            </h2>
            <DateRangeSelector value={timeRange} onChange={setTimeRange} />
          </div>

          <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
            {viewsTrendLoading ? (
              <div className="bg-surface rounded-lg shadow p-6 h-80 animate-pulse">
                <div className="h-6 bg-muted rounded w-1/3 mb-4"></div>
                <div className="h-full bg-muted rounded"></div>
              </div>
            ) : viewsTrend?.data ? (
              <LineChartComponent
                data={viewsTrend.data}
                title="Views Over Time"
                valueLabel="Views"
                color="#8b5cf6"
              />
            ) : null}

            {votesTrendLoading ? (
              <div className="bg-surface rounded-lg shadow p-6 h-80 animate-pulse">
                <div className="h-6 bg-muted rounded w-1/3 mb-4"></div>
                <div className="h-full bg-muted rounded"></div>
              </div>
            ) : votesTrend?.data ? (
              <LineChartComponent
                data={votesTrend.data}
                title="Votes Over Time"
                valueLabel="Votes"
                color="#ec4899"
              />
            ) : null}
          </div>
        </div>

        {/* Audience Insights Section */}
        <div className="mb-8">
          <h2 className="text-2xl font-bold text-foreground mb-4">
            Audience Insights
          </h2>
          <AudienceInsightsSection creatorName={creatorName} />
        </div>
      </div>
    </>
  );
};

export default CreatorAnalyticsPage;
