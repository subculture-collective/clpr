import { Link } from 'react-router-dom';
import { formatTimestamp, cn } from '@/lib/utils';
import type { ForumThread } from '@/types/forum';

interface ThreadCardProps {
  thread: ForumThread;
  className?: string;
}

export function ThreadCard({ thread, className }: ThreadCardProps) {
  const timestamp = formatTimestamp(thread.updated_at);

  return (
    <Link
      to={`/forum/threads/${thread.id}`}
      className={cn(
        'block p-4 bg-surface hover:bg-surface-hover rounded-lg border border-border',
        'transition-colors duration-200',
        className
      )}
    >
      <div className="flex justify-between items-start gap-4">
        <div className="flex-1 min-w-0">
          <div className="flex items-center gap-2 mb-1">
            {thread.pinned && (
              <span className="px-2 py-0.5 bg-primary-600 text-white text-xs font-semibold rounded">
                Pinned
              </span>
            )}
            {thread.locked && (
              <span className="px-2 py-0.5 bg-muted text-foreground text-xs font-semibold rounded">
                Locked
              </span>
            )}
          </div>
          
          <h3 className="text-lg font-semibold text-white mb-1 truncate">
            {thread.title}
          </h3>
          
          <p className="text-sm text-muted-foreground line-clamp-2 mb-2">
            {thread.content}
          </p>

          <div className="flex items-center gap-2 text-xs text-muted-foreground">
            <span>by {thread.username}</span>
            {thread.game_name && (
              <>
                <span>•</span>
                <span>{thread.game_name}</span>
              </>
            )}
          </div>

          {thread.tags && thread.tags.length > 0 && (
            <div className="mt-3 flex gap-2 flex-wrap">
              {thread.tags.map((tag) => (
                <span
                  key={tag}
                  className="px-2 py-1 bg-surface-raised text-foreground rounded text-xs"
                >
                  {tag}
                </span>
              ))}
            </div>
          )}
        </div>

        <div className="flex-shrink-0 text-right text-sm">
          <div className="text-muted-foreground mb-1">
            <span className="font-semibold text-white">{thread.reply_count}</span> replies
          </div>
          <div className="text-muted-foreground mb-1">
            <span className="font-semibold text-white">{thread.view_count}</span> views
          </div>
          <div className="text-xs text-muted-foreground" title={timestamp.title}>
            {timestamp.display}
          </div>
        </div>
      </div>
    </Link>
  );
}
