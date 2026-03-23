import { Link } from 'react-router-dom';
import { MessageSquare, ThumbsUp, User, Calendar } from 'lucide-react';
import DOMPurify from 'dompurify';
import { cn } from '@/lib/utils';
import type { SearchResult } from '@/types/forum';

interface SearchResultCardProps {
  result: SearchResult;
  className?: string;
}

// Format date relative to now, moved outside component to avoid recreation on every render
const formatDate = (dateString: string): string => {
  const date = new Date(dateString);
  const now = new Date();
  const diffMs = now.getTime() - date.getTime();
  
  // Handle future dates or invalid dates
  if (diffMs < 0 || isNaN(diffMs)) {
    return date.toLocaleDateString();
  }
  
  const diffMins = Math.floor(diffMs / 60000);
  const diffHours = Math.floor(diffMins / 60);
  const diffDays = Math.floor(diffHours / 24);

  if (diffMins < 60) return `${diffMins}m ago`;
  if (diffHours < 24) return `${diffHours}h ago`;
  if (diffDays < 7) return `${diffDays}d ago`;
  return date.toLocaleDateString();
};

export function SearchResultCard({ result, className }: SearchResultCardProps) {
  // Determine the link based on result type
  // Add safety check for missing thread_id in reply results
  const link = result.type === 'thread' 
    ? `/forum/threads/${result.id}`
    : result.thread_id
      ? `/forum/threads/${result.thread_id}#reply-${result.id}`
      : '/forum';

  // Sanitize the headline to prevent XSS attacks
  const sanitizedHeadline = DOMPurify.sanitize(result.headline, {
    ALLOWED_TAGS: ['b', 'strong', 'em', 'i'],
    ALLOWED_ATTR: []
  });

  return (
    <Link
      to={link}
      className={cn(
        'block p-4 bg-surface border border-border rounded-lg',
        'hover:bg-surface-hover hover:border-subtle transition-all duration-200',
        className
      )}
    >
      <div className="flex items-start justify-between gap-4 mb-3">
        <div className="flex-1 min-w-0">
          {result.title && (
            <h3 className="text-lg font-semibold text-white mb-1 truncate">
              {result.title}
            </h3>
          )}
          <div className="flex items-center gap-3 text-sm text-muted-foreground flex-wrap">
            <div className="flex items-center gap-1">
              <User className="w-4 h-4" />
              <span>{result.author_name}</span>
            </div>
            <span>•</span>
            <div className="flex items-center gap-1">
              <Calendar className="w-4 h-4" />
              <span>{formatDate(result.created_at)}</span>
            </div>
            {result.vote_count !== 0 && (
              <>
                <span>•</span>
                <div className="flex items-center gap-1">
                  <ThumbsUp className="w-4 h-4" />
                  <span>{result.vote_count} votes</span>
                </div>
              </>
            )}
          </div>
        </div>
        <div className="flex-shrink-0">
          <span
            className={cn(
              'inline-flex items-center gap-1 px-2 py-1 text-xs font-medium rounded',
              result.type === 'thread'
                ? 'bg-blue-900/30 text-blue-400 border border-blue-800'
                : 'bg-purple-900/30 text-purple-400 border border-purple-800'
            )}
          >
            {result.type === 'thread' ? (
              <>
                <MessageSquare className="w-3 h-3" />
                <span>Thread</span>
              </>
            ) : (
              <>
                <MessageSquare className="w-3 h-3" />
                <span>Reply</span>
              </>
            )}
          </span>
        </div>
      </div>

      {/* Highlighted snippet */}
      <div
        className="text-sm text-foreground line-clamp-3 search-highlight"
        dangerouslySetInnerHTML={{ __html: sanitizedHeadline }}
      />
    </Link>
  );
}
