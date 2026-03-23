import { cn } from '@/lib/utils';

interface ReputationBadgeProps {
  badge: 'new' | 'contributor' | 'expert' | 'moderator';
  score?: number;
  showScore?: boolean;
  size?: 'sm' | 'md' | 'lg';
  className?: string;
}

const badgeConfig = {
  new: {
    color: 'bg-muted',
    label: 'New Member',
    textColor: 'text-white',
  },
  contributor: {
    color: 'bg-blue-500',
    label: 'Contributor',
    textColor: 'text-white',
  },
  expert: {
    color: 'bg-yellow-500',
    label: 'Expert',
    textColor: 'text-background',
  },
  moderator: {
    color: 'bg-red-500',
    label: 'Moderator',
    textColor: 'text-white',
  },
};

const sizeClasses = {
  sm: 'px-1.5 py-0.5 text-xs',
  md: 'px-2 py-1 text-xs',
  lg: 'px-3 py-1.5 text-sm',
};

export function ReputationBadge({
  badge,
  score,
  showScore = true,
  size = 'md',
  className,
}: ReputationBadgeProps) {
  const config = badgeConfig[badge] || badgeConfig.new;
  const sizeClass = sizeClasses[size];

  return (
    <div className={cn('flex items-center gap-2', className)}>
      <span
        className={cn(
          'rounded font-bold',
          config.color,
          config.textColor,
          sizeClass
        )}
        title={`${config.label}${score !== undefined ? ` - ${score} reputation` : ''}`}
      >
        {config.label}
      </span>
      {showScore && score !== undefined && (
        <span className="text-xs text-muted-foreground">{score} rep</span>
      )}
    </div>
  );
}

interface ReputationProgressBarProps {
  score: number;
  className?: string;
}

export function ReputationProgressBar({
  score,
  className,
}: ReputationProgressBarProps) {
  // Calculate progress within current tier
  let progress = 0;
  let nextTier: string;
  let pointsToNext: number;

  if (score < 50) {
    progress = (score / 50) * 100;
    nextTier = 'Contributor';
    pointsToNext = 50 - score;
  } else if (score < 250) {
    progress = ((score - 50) / 200) * 100;
    nextTier = 'Expert';
    pointsToNext = 250 - score;
  } else {
    progress = 100;
    nextTier = 'Max Level';
    pointsToNext = 0;
  }

  return (
    <div className={cn('space-y-2', className)}>
      <div className="flex justify-between text-xs text-muted-foreground">
        <span>{score} reputation</span>
        {pointsToNext > 0 && (
          <span>{pointsToNext} to {nextTier}</span>
        )}
      </div>
      <div className="w-full bg-surface-raised rounded-full h-2 overflow-hidden">
        <div
          className={cn(
            'h-full transition-all duration-500',
            score < 50 && 'bg-blue-500',
            score >= 50 && score < 250 && 'bg-yellow-500',
            score >= 250 && 'bg-green-500'
          )}
          style={{ width: `${progress}%` }}
        />
      </div>
    </div>
  );
}
