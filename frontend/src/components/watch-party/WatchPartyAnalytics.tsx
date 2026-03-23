import { useEffect, useState, type ReactNode } from 'react';
import { getWatchPartyAnalytics } from '@/lib/watch-party-api';
import type { WatchPartyAnalytics } from '@/types/watchParty';
import { Users, TrendingUp, Eye, Clock, MessageSquare, Heart, Flame, AlertTriangle } from 'lucide-react';

interface Props {
  partyId: string;
}

interface StatCardProps {
  label: string;
  value: string | number;
  icon: ReactNode;
  subtext?: string;
}

function StatCard({ label, value, icon, subtext }: StatCardProps) {
  return (
    <div className="bg-gray-800 rounded-lg p-6 border border-gray-700">
      <div className="flex items-center justify-between mb-2">
        <span className="text-gray-400 text-sm font-medium">{label}</span>
        <span className="text-gray-400">{icon}</span>
      </div>
      <div className="text-3xl font-bold text-white mb-1">{value}</div>
      {subtext && <div className="text-gray-500 text-xs">{subtext}</div>}
    </div>
  );
}

export default function WatchPartyAnalytics({ partyId }: Props) {
  const [analytics, setAnalytics] = useState<WatchPartyAnalytics | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    loadAnalytics();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [partyId]); // loadAnalytics captures partyId from props

  const loadAnalytics = async () => {
    try {
      setLoading(true);
      setError(null);
      const data = await getWatchPartyAnalytics(partyId);
      setAnalytics(data);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to load analytics');
    } finally {
      setLoading(false);
    }
  };

  if (loading) {
    return (
      <div className="p-6 bg-gray-900 rounded-lg">
        <div className="animate-pulse">
          <div className="h-8 bg-gray-700 rounded w-1/3 mb-6"></div>
          <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
            {[1, 2, 3, 4].map((i) => (
              <div key={i} className="bg-gray-800 rounded-lg p-6 h-32"></div>
            ))}
          </div>
        </div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="p-6 bg-gray-900 rounded-lg">
        <div className="text-red-500 text-center">
          <p className="text-xl mb-2"><AlertTriangle size={20} strokeWidth={1.75} className="inline" /></p>
          <p>{error}</p>
          <button
            onClick={loadAnalytics}
            className="mt-4 px-4 py-2 bg-blue-600 hover:bg-blue-700 rounded-lg text-white"
          >
            Retry
          </button>
        </div>
      </div>
    );
  }

  if (!analytics) {
    return null;
  }

  const avgWatchMinutes = Math.round(analytics.avg_duration_seconds / 60);
  const totalEngagement = analytics.chat_messages + analytics.reactions;
  const engagementRate =
    analytics.unique_viewers > 0
      ? ((totalEngagement / analytics.unique_viewers) * 100).toFixed(1)
      : '0';

  return (
    <div className="p-6 bg-gray-900 rounded-lg">
      <h2 className="text-2xl font-bold text-white mb-6">Watch Party Analytics</h2>

      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4 mb-6">
        <StatCard
          label="Total Viewers"
          value={analytics.unique_viewers}
          icon={<Users size={20} strokeWidth={1.75} />}
          subtext="Unique participants"
        />
        <StatCard
          label="Peak Concurrent"
          value={analytics.peak_concurrent}
          icon={<TrendingUp size={20} strokeWidth={1.75} />}
          subtext="Maximum viewers at once"
        />
        <StatCard
          label="Current Viewers"
          value={analytics.current_viewers}
          icon={<Eye size={20} strokeWidth={1.75} />}
          subtext="Live now"
        />
        <StatCard
          label="Avg Watch Time"
          value={`${avgWatchMinutes}m`}
          icon={<Clock size={20} strokeWidth={1.75} />}
          subtext={`${analytics.avg_duration_seconds}s total`}
        />
      </div>

      <div className="grid grid-cols-1 sm:grid-cols-3 gap-4">
        <StatCard
          label="Chat Messages"
          value={analytics.chat_messages}
          icon={<MessageSquare size={20} strokeWidth={1.75} />}
        />
        <StatCard
          label="Reactions"
          value={analytics.reactions}
          icon={<Heart size={20} strokeWidth={1.75} />}
        />
        <StatCard
          label="Engagement Rate"
          value={`${engagementRate}%`}
          icon={<Flame size={20} strokeWidth={1.75} />}
          subtext="Interactions per viewer"
        />
      </div>

      <div className="mt-6 p-4 bg-gray-800 rounded-lg border border-gray-700">
        <h3 className="text-lg font-semibold text-white mb-2">Summary</h3>
        <p className="text-gray-300 text-sm">
          This watch party had <strong>{analytics.unique_viewers}</strong> unique viewers with a peak
          of <strong>{analytics.peak_concurrent}</strong> watching simultaneously. Viewers stayed for an
          average of <strong>{avgWatchMinutes} minutes</strong> and generated{' '}
          <strong>{totalEngagement}</strong> total interactions.
        </p>
      </div>
    </div>
  );
}
