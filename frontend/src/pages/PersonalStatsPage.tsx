import React from 'react';
import { useQuery } from '@tanstack/react-query';
import { Helmet } from '@dr.pogodin/react-helmet';
import { getUserStats } from '../lib/analytics-api';
import { MetricCard, PieChartComponent } from '../components/analytics';
import { format } from 'date-fns';

const PersonalStatsPage: React.FC = () => {
  // Fetch user statistics
  const { data: stats, isLoading, error } = useQuery({
    queryKey: ['userStats'],
    queryFn: getUserStats,
  });

  if (error) {
    return (
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        <div className="bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 rounded-lg p-4">
          <p className="text-red-800 dark:text-red-200">
            Failed to load your statistics. Please try again later.
          </p>
        </div>
      </div>
    );
  }

  return (
    <>
      <Helmet>
        <title>My Statistics - Clipper</title>
      </Helmet>

      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        {/* Header */}
        <div className="mb-8">
          <h1 className="text-3xl font-bold text-foreground">
            My Statistics
          </h1>
          <p className="mt-2 text-muted-foreground">
            Your activity and engagement on the platform
          </p>
        </div>

        {/* Activity Summary */}
        {isLoading ? (
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6 mb-8">
            {[1, 2, 3, 4, 5, 6, 7, 8].map((i) => (
              <div
                key={i}
                className="bg-surface rounded-lg shadow p-6 animate-pulse"
              >
                <div className="h-4 bg-muted rounded w-1/2 mb-4"></div>
                <div className="h-8 bg-muted rounded w-3/4"></div>
              </div>
            ))}
          </div>
        ) : stats ? (
          <>
            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6 mb-8">
              <MetricCard
                title="Clips Upvoted"
                value={stats.clips_upvoted}
                subtitle="Positive engagement"
              />
              <MetricCard
                title="Clips Downvoted"
                value={stats.clips_downvoted}
                subtitle="Negative feedback"
              />
              <MetricCard
                title="Comments Posted"
                value={stats.comments_posted}
                subtitle="Your contributions"
              />
              <MetricCard
                title="Clips Favorited"
                value={stats.clips_favorited}
                subtitle="Saved clips"
              />
              <MetricCard
                title="Searches Performed"
                value={stats.searches_performed}
                subtitle="Exploration"
              />
              <MetricCard
                title="Days Active"
                value={stats.days_active}
                subtitle="Platform engagement"
              />
              <MetricCard
                title="Karma Earned"
                value={stats.total_karma_earned}
                subtitle="Reputation points"
              />
              <MetricCard
                title="Last Active"
                value={
                  stats.last_active_at
                    ? format(new Date(stats.last_active_at), 'MMM dd, yyyy')
                    : 'N/A'
                }
                subtitle="Most recent activity"
              />
            </div>

            {/* Voting Distribution */}
            {(stats.clips_upvoted > 0 || stats.clips_downvoted > 0) && (
              <div className="mb-8">
                <h2 className="text-xl font-bold text-foreground mb-4">
                  Voting Behavior
                </h2>

                <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
                  <PieChartComponent
                    data={[
                      { name: 'Upvotes', value: stats.clips_upvoted },
                      { name: 'Downvotes', value: stats.clips_downvoted },
                    ]}
                    title="Upvotes vs Downvotes"
                    colors={['#10b981', '#ef4444']}
                  />

                  <div className="bg-surface rounded-lg shadow p-6">
                    <h3 className="text-lg font-semibold text-foreground mb-4">
                      Engagement Summary
                    </h3>

                    <div className="space-y-4">
                      <div>
                        <div className="flex items-center justify-between">
                          <span className="text-sm text-muted-foreground">
                            Total Votes Cast
                          </span>
                          <span className="text-sm font-semibold text-foreground">
                            {stats.clips_upvoted + stats.clips_downvoted}
                          </span>
                        </div>
                      </div>

                      <div>
                        <div className="flex items-center justify-between mb-2">
                          <span className="text-sm text-muted-foreground">
                            Positivity Rate
                          </span>
                          <span className="text-sm font-semibold text-foreground">
                            {(
                              (stats.clips_upvoted /
                                Math.max(
                                  stats.clips_upvoted + stats.clips_downvoted,
                                  1
                                )) *
                              100
                            ).toFixed(1)}
                            %
                          </span>
                        </div>
                        <div className="w-full bg-muted rounded-full h-2">
                          <div
                            className="bg-green-500 h-2 rounded-full"
                            style={{
                              width: `${
                                (stats.clips_upvoted /
                                  Math.max(
                                    stats.clips_upvoted + stats.clips_downvoted,
                                    1
                                  )) *
                                100
                              }%`,
                            }}
                          ></div>
                        </div>
                      </div>

                      <div>
                        <div className="flex items-center justify-between mb-2">
                          <span className="text-sm text-muted-foreground">
                            Comments per Vote
                          </span>
                          <span className="text-sm font-semibold text-foreground">
                            {(
                              stats.comments_posted /
                              Math.max(
                                stats.clips_upvoted + stats.clips_downvoted,
                                1
                              )
                            ).toFixed(2)}
                          </span>
                        </div>
                      </div>
                    </div>
                  </div>
                </div>
              </div>
            )}

            {/* Account Information */}
            <div className="bg-surface rounded-lg shadow p-6">
              <h3 className="text-lg font-semibold text-foreground mb-4">
                Account Summary
              </h3>

              <dl className="grid grid-cols-1 md:grid-cols-2 gap-6">
                <div>
                  <dt className="text-sm font-medium text-muted-foreground">
                    Total Karma
                  </dt>
                  <dd className="mt-1 text-2xl font-semibold text-foreground">
                    {stats.total_karma_earned}
                  </dd>
                </div>
                <div>
                  <dt className="text-sm font-medium text-muted-foreground">
                    Days Active
                  </dt>
                  <dd className="mt-1 text-2xl font-semibold text-foreground">
                    {stats.days_active}
                  </dd>
                </div>
                <div>
                  <dt className="text-sm font-medium text-muted-foreground">
                    Contribution Score
                  </dt>
                  <dd className="mt-1 text-2xl font-semibold text-foreground">
                    {stats.comments_posted + stats.clips_upvoted}
                  </dd>
                </div>
                <div>
                  <dt className="text-sm font-medium text-muted-foreground">
                    Engagement Level
                  </dt>
                  <dd className="mt-1 text-2xl font-semibold text-foreground">
                    {stats.comments_posted > 50 || stats.clips_upvoted > 100
                      ? 'High'
                      : stats.comments_posted > 10 || stats.clips_upvoted > 20
                      ? 'Medium'
                      : 'Low'}
                  </dd>
                </div>
              </dl>
            </div>
          </>
        ) : null}
      </div>
    </>
  );
};

export default PersonalStatsPage;
