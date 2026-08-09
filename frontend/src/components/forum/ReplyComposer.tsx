import { useId, useState } from 'react';
import { cn } from '@/lib/utils';

interface ReplyComposerProps {
  onSubmit: (content: string) => Promise<void>;
  onCancel?: () => void;
  placeholder?: string;
  submitLabel?: string;
  isMobile?: boolean;
  className?: string;
}

export function ReplyComposer({
  onSubmit,
  onCancel,
  placeholder = 'Write your reply... (Markdown supported)',
  submitLabel = 'Post Reply',
  isMobile = false,
  className,
}: ReplyComposerProps) {
  const [content, setContent] = useState('');
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [submitError, setSubmitError] = useState('');
  const composerId = useId();
  const helpId = `${composerId}-help`;
  const errorId = `${composerId}-error`;

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    
    if (!content.trim() || isSubmitting) {
      return;
    }

    setIsSubmitting(true);
    setSubmitError('');
    try {
      await onSubmit(content.trim());
      setContent('');
    } catch (error) {
      console.error('Failed to post reply:', error);
      setSubmitError('Your reply could not be posted. Please try again.');
    } finally {
      setIsSubmitting(false);
    }
  };

  return (
    <form
      onSubmit={handleSubmit}
      className={cn(
        'bg-surface rounded-lg border border-border p-4',
        isMobile && 'rounded-t-lg border-x-0 border-b-0',
        className
      )}
    >
      <label htmlFor={composerId} className="sr-only">Reply</label>
      <textarea
        id={composerId}
        value={content}
        onChange={(e) => setContent(e.target.value)}
        placeholder={placeholder}
        className={cn(
          'w-full bg-surface-raised text-white rounded-lg p-3 mb-3',
          'border border-border focus:border-primary-500 focus:outline-none',
          'resize-none placeholder-muted-foreground'
        )}
        rows={isMobile ? 6 : 4}
        disabled={isSubmitting}
        aria-describedby={`${helpId}${submitError ? ` ${errorId}` : ''}`}
        aria-invalid={submitError ? true : undefined}
      />

      {submitError && (
        <p id={errorId} role="alert" className="mb-3 text-sm text-error-500">
          {submitError}
        </p>
      )}

      <div className="flex justify-between items-center">
        <p id={helpId} className="text-xs text-muted-foreground">
          Markdown formatting is supported
        </p>
        <div className="flex gap-2">
          {onCancel && (
            <button
              type="button"
              onClick={onCancel}
              disabled={isSubmitting}
              className={cn(
                'min-h-[44px] px-4 py-2 bg-surface-raised hover:bg-surface-hover text-white rounded-lg',
                'transition-colors motion-reduce:transition-none disabled:opacity-50 disabled:cursor-not-allowed focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-primary-500'
              )}
            >
              Cancel
            </button>
          )}
          <button
            type="submit"
            disabled={!content.trim() || isSubmitting}
            className={cn(
              'min-h-[44px] px-4 py-2 bg-primary-600 hover:bg-primary-700 text-white rounded-lg',
              'transition-colors motion-reduce:transition-none disabled:opacity-50 disabled:cursor-not-allowed focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-primary-500'
            )}
            aria-busy={isSubmitting}
          >
            {isSubmitting ? 'Posting...' : submitLabel}
          </button>
        </div>
      </div>
    </form>
  );
}
