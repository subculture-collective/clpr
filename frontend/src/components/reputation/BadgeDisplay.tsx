import type { UserBadge } from '../../types/reputation';

interface BadgeDisplayProps {
  badges: UserBadge[];
  maxVisible?: number;
  size?: 'sm' | 'md' | 'lg';
  showTooltip?: boolean;
}

const sizeClasses = {
  sm: 'text-base',
  md: 'text-xl',
  lg: 'text-3xl',
};

export function BadgeDisplay({ 
  badges, 
  maxVisible = 3, 
  size = 'md',
  showTooltip = true 
}: BadgeDisplayProps) {
  const visibleBadges = badges.slice(0, maxVisible);
  const remainingCount = badges.length - maxVisible;

  if (badges.length === 0) {
    return null;
  }

  return (
    <div className="flex items-center gap-2">
      {visibleBadges.map((badge) => (
        <div
          key={badge.id}
          className={`${sizeClasses[size]} ${showTooltip ? 'cursor-help' : ''}`}
          title={showTooltip ? `${badge.name}: ${badge.description}` : undefined}
          aria-label={`${badge.name}: ${badge.description}`}
        >
          {badge.icon}
        </div>
      ))}
      {remainingCount > 0 && (
        <span className="text-sm text-muted-foreground">
          +{remainingCount}
        </span>
      )}
    </div>
  );
}

interface BadgeGridProps {
  badges: UserBadge[];
  columns?: number;
}

export function BadgeGrid({ badges, columns = 3 }: BadgeGridProps) {
  if (badges.length === 0) {
    return (
      <div className="py-8 text-center text-muted-foreground">
        No badges earned yet
      </div>
    );
  }

  return (
    <div 
      className="grid gap-4"
      style={{ gridTemplateColumns: `repeat(${columns}, minmax(0, 1fr))` }}
    >
      {badges.map((badge) => (
        <div
          key={badge.id}
          className="hover:bg-surface-hover p-4 transition-colors bg-surface-raised rounded-lg"
        >
          <div className="flex items-start gap-3">
            <div className="shrink-0 text-3xl">
              {badge.icon}
            </div>
            <div className="flex-1 min-w-0">
              <h3 className="font-semibold text-white truncate">
                {badge.name}
              </h3>
              <p className="mt-1 text-sm text-muted-foreground">
                {badge.description}
              </p>
              <div className="mt-2 text-xs text-text-secondary">
                Earned {new Date(badge.awarded_at).toLocaleDateString()}
              </div>
            </div>
          </div>
        </div>
      ))}
    </div>
  );
}

interface BadgeListProps {
  badges: UserBadge[];
}

export function BadgeList({ badges }: BadgeListProps) {
  if (badges.length === 0) {
    return (
      <div className="py-4 text-center text-muted-foreground">
        No badges earned yet
      </div>
    );
  }

  return (
    <div className="space-y-2">
      {badges.map((badge) => (
        <div
          key={badge.id}
          className="hover:bg-surface-hover flex items-center gap-3 p-3 transition-colors bg-surface-raised rounded-lg"
        >
          <div className="shrink-0 text-2xl">
            {badge.icon}
          </div>
          <div className="flex-1 min-w-0">
            <div className="font-semibold text-white">
              {badge.name}
            </div>
            <div className="text-sm text-muted-foreground">
              {badge.description}
            </div>
          </div>
          <div className="text-xs text-text-secondary">
            {new Date(badge.awarded_at).toLocaleDateString()}
          </div>
        </div>
      ))}
    </div>
  );
}
