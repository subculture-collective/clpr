import { Film, MessageSquare, ThumbsUp } from 'lucide-react';
import { Link } from 'react-router-dom';
import type { LeaderboardEntry } from '../../types/reputation';
import { RankBadge } from './ReputationDisplay';

interface LeaderboardTableProps {
  entries: LeaderboardEntry[];
  type: 'karma' | 'engagement';
  currentUserId?: string;
}

const rankColors: Record<number, string> = {
  1: 'text-yellow-400',
  2: 'text-neutral-300',
  3: 'text-amber-600',
};

function RankMedal({ rank }: { rank: number }) {
  const colorClass = rankColors[rank];
  if (!colorClass) return null;
  return (
    <span className={`font-accent font-extrabold text-lg ${colorClass}`}>
      {rank}
    </span>
  );
}

export function LeaderboardTable({ entries, type, currentUserId }: LeaderboardTableProps) {
  if (entries.length === 0) {
    return (
      <div className="py-12 text-center">
        <div className="text-lg text-muted-foreground">No leaderboard data available</div>
      </div>
    );
  }

  return (
    <div className="overflow-hidden bg-surface-raised rounded-lg">
      <table className="w-full">
        <thead className="bg-surface">
          <tr>
            <th className="px-6 py-4 text-sm font-semibold text-left text-foreground">
              Rank
            </th>
            <th className="px-6 py-4 text-sm font-semibold text-left text-foreground">
              User
            </th>
            <th className="px-6 py-4 text-sm font-semibold text-left text-foreground">
              Tier
            </th>
            <th className="px-6 py-4 text-sm font-semibold text-right text-foreground">
              {type === 'karma' ? 'Karma' : 'Engagement'}
            </th>
            {type === 'engagement' && (
              <th className="px-6 py-4 text-sm font-semibold text-right text-foreground">
                Activity
              </th>
            )}
          </tr>
        </thead>
        <tbody className="divide-y divide-border">
          {entries.map((entry) => {
            const isCurrentUser = entry.user_id === currentUserId;
            const rowClasses = isCurrentUser
              ? 'bg-purple-900/20 hover:bg-purple-900/30'
              : 'hover:bg-surface-hover';

            return (
              <tr key={entry.user_id} className={rowClasses}>
                {/* Rank */}
                <td className="px-6 py-4">
                  <div className="flex items-center gap-2">
                    {rankColors[entry.rank] && (
                      <RankMedal rank={entry.rank} />
                    )}
                    <span className="text-lg font-semibold text-white">
                      #{entry.rank}
                    </span>
                  </div>
                </td>

                {/* User */}
                <td className="px-6 py-4">
                  <Link
                    to={`/profile/${entry.username}`}
                    className="group flex items-center gap-3"
                  >
                    {entry.avatar_url ? (
                      <img
                        src={entry.avatar_url}
                        alt={entry.username}
                        className="w-10 h-10 rounded-full"
                      />
                    ) : (
                      <div className="flex items-center justify-center w-10 h-10 bg-surface-raised rounded-full">
                        <span className="text-lg font-semibold text-muted-foreground">
                          {entry.username[0].toUpperCase()}
                        </span>
                      </div>
                    )}
                    <div>
                      <div className="group-hover:text-purple-400 font-semibold text-white transition-colors">
                        {entry.display_name || entry.username}
                      </div>
                      <div className="text-sm text-muted-foreground">
                        @{entry.username}
                      </div>
                    </div>
                  </Link>
                </td>

                {/* Rank/Tier */}
                <td className="px-6 py-4">
                  <RankBadge rank={entry.user_rank} size="sm" />
                </td>

                {/* Score */}
                <td className="px-6 py-4 text-right">
                  <div className="text-xl font-bold text-purple-400">
                    {entry.score.toLocaleString()}
                  </div>
                </td>

                {/* Activity Stats (for engagement leaderboard) */}
                {type === 'engagement' && (
                  <td className="px-6 py-4">
                    <div className="space-y-1 text-sm text-right text-muted-foreground">
                      <div className="flex items-center justify-end gap-1"><MessageSquare size={16} strokeWidth={1.75} /> {entry.total_comments?.toLocaleString() || 0} comments</div>
                      <div className="flex items-center justify-end gap-1"><ThumbsUp size={16} strokeWidth={1.75} /> {entry.total_votes_cast?.toLocaleString() || 0} votes</div>
                      <div className="flex items-center justify-end gap-1"><Film size={16} strokeWidth={1.75} /> {entry.total_clips_submitted?.toLocaleString() || 0} clips</div>
                    </div>
                  </td>
                )}
              </tr>
            );
          })}
        </tbody>
      </table>
    </div>
  );
}

interface LeaderboardSummaryProps {
  entries: LeaderboardEntry[];
  type: 'karma' | 'engagement';
}

export function LeaderboardSummary({ entries, type }: LeaderboardSummaryProps) {
  const topThree = entries.slice(0, 3);

  if (topThree.length === 0) {
    return null;
  }

  return (
    <div className="grid grid-cols-3 gap-4 mb-6">
      {topThree.map((entry, index) => (
        <Link
          key={entry.user_id}
          to={`/profile/${entry.username}`}
          className="hover:bg-surface-hover p-6 text-center transition-colors bg-surface-raised rounded-lg"
        >
          <div className="mb-2 text-4xl"><RankMedal rank={index + 1} /></div>
          <div className="mb-2">
            {entry.avatar_url ? (
              <img
                src={entry.avatar_url}
                alt={entry.username}
                className="w-16 h-16 mx-auto rounded-full"
              />
            ) : (
              <div className="flex items-center justify-center w-16 h-16 mx-auto bg-surface-raised rounded-full">
                <span className="text-2xl font-semibold text-muted-foreground">
                  {entry.username[0].toUpperCase()}
                </span>
              </div>
            )}
          </div>
          <div className="mb-1 font-semibold text-white">
            {entry.display_name || entry.username}
          </div>
          <RankBadge rank={entry.user_rank} size="sm" />
          <div className="mt-3 text-2xl font-bold text-purple-400">
            {entry.score.toLocaleString()}
          </div>
          <div className="mt-1 text-xs text-muted-foreground">
            {type === 'karma' ? 'karma' : 'engagement'}
          </div>
        </Link>
      ))}
    </div>
  );
}
