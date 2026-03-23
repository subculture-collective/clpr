import { useEffect, useState, type ReactNode } from 'react';
import { getUserWatchPartyStats } from '@/lib/watch-party-api';
import type { HostStats } from '@/types/watchParty';
import { PartyPopper, Users, BarChart3, MessageSquare, Heart, AlertTriangle } from 'lucide-react';

interface Props {
  userId: string;
}

interface StatItemProps {
  label: string;
  value: string | number;
  icon: ReactNode;
}

function StatItem({ label, value, icon }: StatItemProps) {
  return (
    <div className="flex items-center justify-between p-4 bg-gray-800 rounded-lg border border-gray-700">
      <div>
        <div className="text-gray-400 text-sm mb-1">{label}</div>
        <div className="text-2xl font-bold text-white">{value}</div>
      </div>
      <div className="text-gray-400">{icon}</div>
    </div>
  );
}

export default function HostStatsPanel({ userId }: Props) {
  const [stats, setStats] = useState<HostStats | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    loadStats();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [userId]); // loadStats captures userId from props

  const loadStats = async () => {
    try {
      setLoading(true);
      setError(null);
      const data = await getUserWatchPartyStats(userId);
      setStats(data);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to load stats');
    } finally {
      setLoading(false);
    }
  };

  if (loading) {
    return (
      <div className="p-6 bg-gray-900 rounded-lg">
        <div className="animate-pulse">
          <div className="h-8 bg-gray-700 rounded w-1/2 mb-6"></div>
          <div className="space-y-4">
            {[1, 2, 3, 4, 5].map((i) => (
              <div key={i} className="bg-gray-800 rounded-lg p-4 h-20"></div>
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
            onClick={loadStats}
            className="mt-4 px-4 py-2 bg-blue-600 hover:bg-blue-700 rounded-lg text-white"
          >
            Retry
          </button>
        </div>
      </div>
    );
  }

  if (!stats) {
    return null;
  }

  return (
    <div className="p-6 bg-gray-900 rounded-lg">
      <h2 className="text-2xl font-bold text-white mb-6">Your Watch Party Stats</h2>

      <div className="space-y-4">
        <StatItem
          label="Parties Hosted"
          value={stats.total_parties_hosted}
          icon={<PartyPopper size={20} strokeWidth={1.75} />}
        />
        <StatItem
          label="Total Viewers"
          value={stats.total_viewers}
          icon={<Users size={20} strokeWidth={1.75} />}
        />
        <StatItem
          label="Avg Viewers per Party"
          value={stats.avg_viewers_per_party.toFixed(1)}
          icon={<BarChart3 size={20} strokeWidth={1.75} />}
        />
        <StatItem
          label="Total Chat Messages"
          value={stats.total_chat_messages}
          icon={<MessageSquare size={20} strokeWidth={1.75} />}
        />
        <StatItem
          label="Total Reactions"
          value={stats.total_reactions}
          icon={<Heart size={20} strokeWidth={1.75} />}
        />
      </div>

      {stats.total_parties_hosted > 0 && (
        <div className="mt-6 p-4 bg-gray-800 rounded-lg border border-gray-700">
          <h3 className="text-lg font-semibold text-white mb-2">Performance</h3>
          <p className="text-gray-300 text-sm">
            You've hosted <strong>{stats.total_parties_hosted}</strong> watch parties with an
            average of <strong>{stats.avg_viewers_per_party.toFixed(1)}</strong> viewers each. Your
            parties generated <strong>{stats.total_chat_messages + stats.total_reactions}</strong>{' '}
            total interactions!
          </p>
        </div>
      )}
    </div>
  );
}
