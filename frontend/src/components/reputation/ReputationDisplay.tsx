import type { UserReputation } from '../../types/reputation';
import { BadgeDisplay } from './BadgeDisplay';

interface ReputationDisplayProps {
  reputation: UserReputation;
  compact?: boolean;
}

const rankColors: Record<string, string> = {
  'Newcomer': 'text-muted-foreground',
  'Member': 'text-green-400',
  'Regular': 'text-blue-400',
  'Contributor': 'text-purple-400',
  'Veteran': 'text-yellow-400',
  'Legend': 'text-red-400',
};

export function ReputationDisplay({ reputation, compact = false }: ReputationDisplayProps) {
  const rankColor = rankColors[reputation.rank] || 'text-muted-foreground';

  if (compact) {
    return (
      <div className="flex items-center gap-3">
        {/* Karma Points */}
        <div className="flex items-center gap-1">
          <span className="font-semibold text-purple-400">
            {reputation.karma_points.toLocaleString()}
          </span>
          <span className="text-xs text-muted-foreground">karma</span>
        </div>

        {/* Rank */}
        <div className={`text-sm font-semibold ${rankColor}`}>
          {reputation.rank}
        </div>

        {/* Badges */}
        {reputation.badges && reputation.badges.length > 0 && (
          <BadgeDisplay badges={reputation.badges} maxVisible={3} size="sm" />
        )}
      </div>
    );
  }

  return (
    <div className="p-6 bg-surface-raised rounded-lg">
      {/* Header */}
      <div className="flex items-center justify-between mb-6">
        <div>
          <h2 className="text-2xl font-bold text-white">
            {reputation.display_name || reputation.username}
          </h2>
          <div className={`text-lg font-semibold ${rankColor} mt-1`}>
            {reputation.rank}
          </div>
        </div>
        {reputation.avatar_url && (
          <img
            src={reputation.avatar_url}
            alt={reputation.username}
            className="w-16 h-16 rounded-full"
          />
        )}
      </div>

      {/* Karma */}
      <div className="mb-6">
        <div className="p-4 text-center bg-surface rounded-lg">
          <div className="text-4xl font-bold text-purple-400">
            {((reputation?.karma_points) || 0).toLocaleString()}
          </div>
          <div className="mt-1 text-sm text-muted-foreground">Total Karma</div>
        </div>
      </div>

      {/* Scores */}
      <div className="grid grid-cols-2 gap-4 mb-6">
        <div className="p-4 text-center bg-surface rounded-lg">
          <div className="text-2xl font-bold text-green-400">
            {((reputation?.trust_score) || 0)}
          </div>
          <div className="mt-1 text-sm text-muted-foreground">Trust Score</div>
        </div>
        <div className="p-4 text-center bg-surface rounded-lg">
          <div className="text-2xl font-bold text-blue-400">
            {((reputation?.engagement_score) || 0).toLocaleString()}
          </div>
          <div className="mt-1 text-sm text-muted-foreground">Engagement</div>
        </div>
      </div>

      {/* Badges */}
      {reputation.badges && reputation.badges.length > 0 && (
        <div>
          <h3 className="mb-3 text-lg font-semibold text-white">Badges</h3>
          <BadgeDisplay badges={reputation.badges} maxVisible={5} size="lg" />
        </div>
      )}

      {/* Stats */}
      {reputation.stats ? (
        <div className="pt-6 mt-6 border-t border-border">
          <h3 className="mb-3 text-lg font-semibold text-white">Activity</h3>
          <div className="grid grid-cols-3 gap-4 text-center">
            <div>
              <div className="text-xl font-bold text-white">
                {((reputation.stats?.total_comments) || 0).toLocaleString()}
              </div>
              <div className="mt-1 text-xs text-muted-foreground">Comments</div>
            </div>
            <div>
              <div className="text-xl font-bold text-white">
                {((reputation.stats?.total_votes_cast) || 0).toLocaleString()}
              </div>
              <div className="mt-1 text-xs text-muted-foreground">Votes</div>
            </div>
            <div>
              <div className="text-xl font-bold text-white">
                {((reputation.stats?.total_clips_submitted) || 0).toLocaleString()}
              </div>
              <div className="mt-1 text-xs text-muted-foreground">Submissions</div>
            </div>
          </div>
        </div>
      ) : null}
    </div>
  );
}

interface RankBadgeProps {
  rank: string;
  size?: 'sm' | 'md' | 'lg';
}

export function RankBadge({ rank, size = 'md' }: RankBadgeProps) {
  const rankColor = rankColors[rank] || 'text-muted-foreground';
  const sizeClasses = {
    sm: 'text-xs px-2 py-0.5',
    md: 'text-sm px-3 py-1',
    lg: 'text-base px-4 py-2',
  };

  return (
    <span className={`${rankColor} ${sizeClasses[size]} bg-surface-raised rounded-full font-semibold`}>
      {rank}
    </span>
  );
}
