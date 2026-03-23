import { cn } from '@/lib/utils';
import { ReplyItem } from './ReplyItem';
import type { ForumReply } from '@/types/forum';

interface ReplyTreeProps {
  replies: ForumReply[];
  threadId: string;
  currentUserId?: string;
  depth?: number;
  onReply: (replyId: string) => void;
  onEdit: (replyId: string, content: string) => void;
  onDelete: (replyId: string) => void;
  className?: string;
}

export function ReplyTree({
  replies,
  threadId,
  currentUserId,
  depth = 0,
  onReply,
  onEdit,
  onDelete,
  className,
}: ReplyTreeProps) {
  if (!replies || replies.length === 0) {
    return null;
  }

  return (
    <div
      className={cn(
        'space-y-4',
        depth > 0 && 'ml-6 pl-4 border-l-2 border-border',
        className
      )}
    >
      {replies.map((reply) => (
        <div key={reply.id}>
          <ReplyItem
            reply={reply}
            depth={depth}
            currentUserId={currentUserId}
            onReply={onReply}
            onEdit={onEdit}
            onDelete={onDelete}
          />

          {/* Nested replies */}
          {reply.replies && reply.replies.length > 0 && (
            <ReplyTree
              replies={reply.replies}
              threadId={threadId}
              currentUserId={currentUserId}
              depth={depth + 1}
              onReply={onReply}
              onEdit={onEdit}
              onDelete={onDelete}
            />
          )}
        </div>
      ))}
    </div>
  );
}
