import { Link } from 'react-router-dom';
import { formatTimestamp } from '@/lib/utils';
import type { Notification } from '../../types/notification';
import {
    MessageSquare,
    AtSign,
    TrendingUp,
    Award,
    ArrowUpCircle,
    Star,
    Trash2,
    AlertTriangle,
    Ban,
    CheckCircle,
    XCircle,
    Bell,
} from 'lucide-react';
import type { LucideIcon } from 'lucide-react';

interface NotificationItemProps {
    notification: Notification;
    onClick?: () => void;
}

const notificationIconMap: Record<string, LucideIcon> = {
    reply: MessageSquare,
    mention: AtSign,
    vote_milestone: TrendingUp,
    badge_earned: Award,
    rank_up: ArrowUpCircle,
    favorited_clip_comment: Star,
    content_removed: Trash2,
    warning: AlertTriangle,
    ban: Ban,
    submission_approved: CheckCircle,
    submission_rejected: XCircle,
};

export function NotificationItem({
    notification,
    onClick,
}: NotificationItemProps) {
    const IconComponent = notificationIconMap[notification.type] || Bell;

    const content = (
        <div
            className={`p-4 border-b border-border transition-colors cursor-pointer ${
                !notification.is_read
                    ? 'bg-brand-tint hover:bg-brand-tint-hover border-l-[3px] border-l-[rgb(var(--color-brand))]'
                    : 'hover:bg-surface-hover'
            }`}
            onClick={onClick}
        >
            <div className='flex gap-3'>
                {/* Icon */}
                <div className='flex-shrink-0 text-text-secondary' aria-hidden='true'>
                    <IconComponent size={20} strokeWidth={1.75} />
                </div>

                {/* Content */}
                <div className='flex-1 min-w-0'>
                    <div className='flex items-start justify-between gap-2'>
                        <div className='flex-1'>
                            <p className='text-sm font-medium'>
                                {notification.title}
                            </p>
                            {notification.message && (
                                <p className='text-sm text-muted-foreground mt-1'>
                                    {notification.message}
                                </p>
                            )}
                        </div>

                        {/* Unread indicator */}
                        {!notification.is_read && (
                            <div
                                className='flex-shrink-0 w-2 h-2 bg-primary-600 rounded-full mt-1'
                                aria-label='Unread'
                            />
                        )}
                    </div>

                    {/* Source user */}
                    {notification.source_display_name && (
                        <div className='flex items-center gap-2 mt-2'>
                            {notification.source_avatar_url && (
                                <img
                                    src={notification.source_avatar_url}
                                    alt={notification.source_display_name}
                                    className='w-5 h-5 rounded-full'
                                />
                            )}
                            <span className='text-xs text-muted-foreground'>
                                {notification.source_display_name}
                            </span>
                        </div>
                    )}

                    {/* Timestamp */}
                    <p
                        className='text-xs text-muted-foreground mt-2'
                        title={formatTimestamp(notification.created_at).title}
                    >
                        {formatTimestamp(notification.created_at).display}
                    </p>
                </div>
            </div>
        </div>
    );

    // Wrap with link if notification has a link
    if (notification.link) {
        return (
            <Link to={notification.link} className='block cursor-pointer'>
                {content}
            </Link>
        );
    }

    return content;
}
