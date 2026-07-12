import { type ReactNode } from 'react';
import { Link } from 'react-router-dom';
import { Button } from './Button';

interface EmptyStateWithActionProps {
  /**
   * Icon to display above the title
   */
  icon?: ReactNode;
  /**
   * Title of the empty state
   */
  title: string;
  /**
   * Description or message
   */
  description: string;
  /**
   * Primary action button
   */
  primaryAction?: {
    label: string;
    href?: string;
    onClick?: () => void;
  };
  /**
   * Secondary action button (optional)
   */
  secondaryAction?: {
    label: string;
    href?: string;
    onClick?: () => void;
  };
  /**
   * Additional helpful tips or guidance
   */
  tips?: string[];
}

/**
 * EmptyStateWithAction component displays a helpful empty state with actionable next steps
 */
export function EmptyStateWithAction({
  icon,
  title,
  description,
  primaryAction,
  secondaryAction,
  tips,
}: EmptyStateWithActionProps) {
  return (
    <div className="bg-card border border-border rounded-xl p-8 text-center" data-testid="empty-state">
      {/* Icon */}
      {icon && (
        <div className="flex justify-center mb-4 text-muted-foreground">
          {icon}
        </div>
      )}

      {/* Title */}
      <h3 className="text-xl font-semibold text-foreground mb-2">
        {title}
      </h3>

      {/* Description */}
      <p className="text-muted-foreground mb-6" data-testid="empty-state-message">
        {description}
      </p>

      {/* Actions */}
      {(primaryAction || secondaryAction) && (
        <div className="flex flex-col sm:flex-row gap-3 justify-center mb-6">
          {primaryAction && (
            primaryAction.href ? (
              <Button asChild size="lg">
                <Link to={primaryAction.href}>
                  {primaryAction.label}
                </Link>
              </Button>
            ) : (
              <Button size="lg" onClick={primaryAction.onClick}>
                {primaryAction.label}
              </Button>
            )
          )}
          {secondaryAction && (
            secondaryAction.href ? (
              <Button asChild size="lg" variant="outline">
                <Link to={secondaryAction.href}>
                  {secondaryAction.label}
                </Link>
              </Button>
            ) : (
              <Button size="lg" variant="outline" onClick={secondaryAction.onClick}>
                {secondaryAction.label}
              </Button>
            )
          )}
        </div>
      )}

      {/* Tips */}
      {tips && tips.length > 0 && (
        <div className="mt-6 pt-6 border-t border-border">
          <p className="text-sm font-medium text-foreground mb-3">
            Helpful tips:
          </p>
          <ul className="text-sm text-muted-foreground space-y-2 text-left max-w-md mx-auto">
            {tips.map((tip, index) => (
              <li key={index} className="flex items-start gap-2">
                <span className="text-primary-500 mt-0.5">•</span>
                <span>{tip}</span>
              </li>
            ))}
          </ul>
        </div>
      )}
    </div>
  );
}
