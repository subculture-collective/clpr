import { useEffect, useState, useCallback } from 'react';
import { subDays } from 'date-fns';
import { Gavel, Clock, Users, BarChart3 } from 'lucide-react';
import {
    getModerationAnalytics,
    type ModerationAnalytics,
} from '../../lib/moderation-api';
import MetricCard from '../analytics/MetricCard';
import PieChartComponent from '../analytics/PieChartComponent';
import BarChartComponent from '../analytics/BarChartComponent';
import LineChartComponent from '../analytics/LineChartComponent';
import DateRangeSelector from '../analytics/DateRangeSelector';

export function ModerationAnalyticsDashboard() {
    const [analytics, setAnalytics] = useState<ModerationAnalytics | null>(null);
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState<string | null>(null);
    const [dateRangeDays, setDateRangeDays] = useState(30);

    const loadAnalytics = useCallback(async () => {
        try {
            setLoading(true);
            setError(null);
            const startDate = subDays(new Date(), dateRangeDays)
                .toISOString()
                .split('T')[0];
            const endDate = new Date().toISOString().split('T')[0];
            
            const response = await getModerationAnalytics({
                start_date: startDate,
                end_date: endDate,
            });
            setAnalytics(response.data);
        } catch (err) {
            setError('Failed to load analytics data');
            console.error('Analytics error:', err);
        } finally {
            setLoading(false);
        }
    }, [dateRangeDays]);

    useEffect(() => {
        loadAnalytics();
    }, [loadAnalytics]);

    if (loading) {
        return (
            <div className="flex h-64 items-center justify-center">
                <div className="text-lg text-gray-600">Loading analytics...</div>
            </div>
        );
    }

    if (error) {
        return (
            <div className="rounded-lg border border-red-200 bg-red-50 p-4">
                <p className="text-red-800">{error}</p>
                <button
                    onClick={loadAnalytics}
                    className="mt-2 text-sm text-red-600 underline hover:text-red-800"
                >
                    Retry
                </button>
            </div>
        );
    }

    if (!analytics) {
        return null;
    }

    // Transform data for charts
    const actionTypeData = Object.entries(analytics.actions_by_type).map(
        ([name, value]) => ({ name, value })
    );

    const moderatorData = Object.entries(analytics.actions_by_moderator)
        .slice(0, 10)
        .map(([name, value]) => ({ name, value }));

    const contentTypeData = Object.entries(analytics.content_type_breakdown).map(
        ([name, value]) => ({ name, value })
    );

    const banReasonsData = Object.entries(analytics.ban_reasons || {}).map(
        ([name, value]) => ({ name, value })
    );

    const bannedUsersData = analytics.most_banned_users || [];

    const activeModerators = Object.keys(analytics.actions_by_moderator).length;
    const avgResponseTime = analytics.average_response_time_minutes
        ? `${Math.round(analytics.average_response_time_minutes)} min`
        : 'N/A';
    
    const falsePositiveRate = analytics.appeals?.false_positive_rate
        ? `${analytics.appeals.false_positive_rate.toFixed(1)}%`
        : 'N/A';

    return (
        <div className="container mx-auto space-y-6 py-6">
            <div className="flex items-center justify-between">
                <h1 className="text-3xl font-bold text-gray-900">
                    Moderation Analytics
                </h1>

                {/* Date Range Selector */}
                <DateRangeSelector
                    value={dateRangeDays}
                    onChange={setDateRangeDays}
                />
            </div>

            {/* Metrics Cards */}
            <div className="grid grid-cols-1 gap-4 md:grid-cols-2 lg:grid-cols-4">
                <MetricCard
                    title="Total Actions"
                    value={analytics.total_actions}
                    icon={<Gavel size={20} />}
                />
                <MetricCard
                    title="Avg Response Time"
                    value={avgResponseTime}
                    icon={<Clock size={20} />}
                />
                <MetricCard
                    title="Active Moderators"
                    value={activeModerators}
                    icon={<Users size={20} />}
                />
                <MetricCard
                    title="False Positive Rate"
                    value={falsePositiveRate}
                    icon={<BarChart3 size={20} />}
                />
            </div>

            {/* Charts Grid */}
            <div className="grid grid-cols-1 gap-6 lg:grid-cols-2">
                {/* Actions by Type */}
                <PieChartComponent
                    data={actionTypeData}
                    title="Actions by Type"
                />

                {/* Top Moderators */}
                <div className="rounded-lg border border-gray-200 bg-white p-6 shadow-sm">
                    <h3 className="mb-4 text-lg font-bold text-gray-900">
                        Top Moderators
                    </h3>
                    <BarChartComponent
                        data={moderatorData}
                        xAxisKey="name"
                        yAxisKey="value"
                        barColor="#3b82f6"
                    />
                </div>

                {/* Content Type Breakdown */}
                <div className="rounded-lg border border-gray-200 bg-white p-6 shadow-sm">
                    <h3 className="mb-4 text-lg font-bold text-gray-900">
                        Content Type Breakdown
                    </h3>
                    <BarChartComponent
                        data={contentTypeData}
                        xAxisKey="name"
                        yAxisKey="value"
                        barColor="#10b981"
                    />
                </div>

                {/* Actions Over Time */}
                <div className="rounded-lg border border-gray-200 bg-white p-6 shadow-sm">
                    <h3 className="mb-4 text-lg font-bold text-gray-900">
                        Actions Over Time
                    </h3>
                    <LineChartComponent
                        data={analytics.actions_over_time}
                        xAxisKey="date"
                        yAxisKey="count"
                        lineColor="#8b5cf6"
                    />
                </div>

                {/* Ban Reasons Distribution */}
                {banReasonsData.length > 0 && (
                    <PieChartComponent
                        data={banReasonsData}
                        title="Ban Reasons Distribution"
                    />
                )}

                {/* Most Banned Users */}
                {bannedUsersData.length > 0 && (
                    <div className="rounded-lg border border-gray-200 bg-white p-6 shadow-sm">
                        <h3 className="mb-4 text-lg font-bold text-gray-900">
                            Most Banned Users
                        </h3>
                        <div className="overflow-x-auto">
                            <table className="min-w-full divide-y divide-gray-200">
                                <thead className="bg-gray-50">
                                    <tr>
                                        <th className="px-4 py-2 text-left text-xs font-medium uppercase tracking-wider text-gray-500">
                                            Username
                                        </th>
                                        <th className="px-4 py-2 text-left text-xs font-medium uppercase tracking-wider text-gray-500">
                                            Ban Count
                                        </th>
                                        <th className="px-4 py-2 text-left text-xs font-medium uppercase tracking-wider text-gray-500">
                                            Last Ban
                                        </th>
                                    </tr>
                                </thead>
                                <tbody className="divide-y divide-gray-200 bg-white">
                                    {bannedUsersData.map((user) => (
                                        <tr key={user.user_id} className="hover:bg-gray-50">
                                            <td className="whitespace-nowrap px-4 py-2 text-sm font-medium text-gray-900">
                                                {user.username}
                                            </td>
                                            <td className="whitespace-nowrap px-4 py-2 text-sm text-gray-500">
                                                {user.ban_count}
                                            </td>
                                            <td className="whitespace-nowrap px-4 py-2 text-sm text-gray-500">
                                                {new Date(user.last_ban_at).toLocaleDateString()}
                                            </td>
                                        </tr>
                                    ))}
                                </tbody>
                            </table>
                        </div>
                    </div>
                )}
            </div>

            {/* Appeals Statistics Section */}
            {analytics.appeals && analytics.appeals.total_appeals > 0 && (
                <div className="mt-6 rounded-lg border border-gray-200 bg-white p-6 shadow-sm">
                    <h3 className="mb-4 text-lg font-bold text-gray-900">
                        Appeals & Reversals
                    </h3>
                    <div className="grid grid-cols-1 gap-4 md:grid-cols-2 lg:grid-cols-4">
                        <div className="rounded-lg bg-blue-50 p-4">
                            <div className="text-sm font-medium text-blue-600">Total Appeals</div>
                            <div className="mt-1 text-2xl font-bold text-blue-900">
                                {analytics.appeals.total_appeals}
                            </div>
                        </div>
                        <div className="rounded-lg bg-yellow-50 p-4">
                            <div className="text-sm font-medium text-yellow-600">Pending</div>
                            <div className="mt-1 text-2xl font-bold text-yellow-900">
                                {analytics.appeals.pending_appeals}
                            </div>
                        </div>
                        <div className="rounded-lg bg-green-50 p-4">
                            <div className="text-sm font-medium text-green-600">Approved</div>
                            <div className="mt-1 text-2xl font-bold text-green-900">
                                {analytics.appeals.approved_appeals}
                            </div>
                        </div>
                        <div className="rounded-lg bg-red-50 p-4">
                            <div className="text-sm font-medium text-red-600">Rejected</div>
                            <div className="mt-1 text-2xl font-bold text-red-900">
                                {analytics.appeals.rejected_appeals}
                            </div>
                        </div>
                    </div>
                </div>
            )}
        </div>
    );
}
