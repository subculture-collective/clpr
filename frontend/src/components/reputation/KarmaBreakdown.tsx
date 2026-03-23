import { Film, MessageSquare } from 'lucide-react';
import type { KarmaBreakdown } from '../../types/reputation';

interface KarmaBreakdownProps {
  breakdown: KarmaBreakdown;
}

export function KarmaBreakdownChart({ breakdown }: KarmaBreakdownProps) {
  const total = breakdown.total_karma;
  const clipPercentage = total > 0 ? (breakdown.clip_karma / total) * 100 : 0;
  const commentPercentage = total > 0 ? (breakdown.comment_karma / total) * 100 : 0;

  return (
    <div className="p-6 bg-surface-raised rounded-lg">
      <h3 className="mb-4 text-lg font-semibold text-white">Karma Breakdown</h3>
      
      {/* Total Karma */}
      <div className="mb-6">
        <div className="text-center">
          <div className="text-4xl font-bold text-purple-400">
            {total.toLocaleString()}
          </div>
          <div className="mt-1 text-sm text-muted-foreground">Total Karma</div>
        </div>
      </div>

      {/* Progress Bars */}
      <div className="space-y-4">
        {/* Clip Karma */}
        <div>
          <div className="flex justify-between mb-1 text-sm">
            <span className="flex items-center gap-1 text-foreground"><Film size={16} strokeWidth={1.75} /> Clip Karma</span>
            <span className="font-semibold text-purple-400">
              {breakdown.clip_karma.toLocaleString()}
            </span>
          </div>
          <div className="w-full h-2 bg-surface-raised rounded-full">
            <div
              className="h-2 transition-all duration-300 bg-purple-500 rounded-full"
              style={{ width: `${clipPercentage}%` }}
            />
          </div>
        </div>

        {/* Comment Karma */}
        <div>
          <div className="flex justify-between mb-1 text-sm">
            <span className="flex items-center gap-1 text-foreground"><MessageSquare size={16} strokeWidth={1.75} /> Comment Karma</span>
            <span className="font-semibold text-blue-400">
              {breakdown.comment_karma.toLocaleString()}
            </span>
          </div>
          <div className="w-full h-2 bg-surface-raised rounded-full">
            <div
              className="h-2 transition-all duration-300 bg-blue-500 rounded-full"
              style={{ width: `${commentPercentage}%` }}
            />
          </div>
        </div>
      </div>

      {/* Percentage Display */}
      <div className="grid grid-cols-2 gap-4 pt-4 mt-6 border-t border-border">
        <div className="text-center">
          <div className="text-2xl font-bold text-purple-400">
            {clipPercentage.toFixed(1)}%
          </div>
          <div className="text-xs text-muted-foreground">from clips</div>
        </div>
        <div className="text-center">
          <div className="text-2xl font-bold text-blue-400">
            {commentPercentage.toFixed(1)}%
          </div>
          <div className="text-xs text-muted-foreground">from comments</div>
        </div>
      </div>
    </div>
  );
}

interface KarmaStatsProps {
  breakdown: KarmaBreakdown;
}

export function KarmaStats({ breakdown }: KarmaStatsProps) {
  return (
    <div className="grid grid-cols-3 gap-4">
      <div className="p-4 text-center bg-surface-raised rounded-lg">
        <div className="text-2xl font-bold text-purple-400">
          {breakdown.total_karma.toLocaleString()}
        </div>
        <div className="mt-1 text-sm text-muted-foreground">Total</div>
      </div>
      <div className="p-4 text-center bg-surface-raised rounded-lg">
        <div className="text-2xl font-bold text-purple-400">
          {breakdown.clip_karma.toLocaleString()}
        </div>
        <div className="mt-1 text-sm text-muted-foreground">Clips</div>
      </div>
      <div className="p-4 text-center bg-surface-raised rounded-lg">
        <div className="text-2xl font-bold text-blue-400">
          {breakdown.comment_karma.toLocaleString()}
        </div>
        <div className="mt-1 text-sm text-muted-foreground">Comments</div>
      </div>
    </div>
  );
}
