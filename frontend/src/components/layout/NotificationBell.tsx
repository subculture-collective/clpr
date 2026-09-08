import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { useEffect, useRef, useState } from 'react';
import { Link } from 'react-router-dom';
import { useClickOutside } from '../../hooks/useClickOutside';
import {
    getNotifications,
    getUnreadCount,
    markAllNotificationsAsRead,
    markNotificationAsRead,
} from '../../lib/notification-api';
import type { Notification } from '../../types/notification';
import { NotificationItem } from './NotificationItem';
import { useAuth } from '../../context/AuthContext';

export function NotificationBell() {
    const [isOpen, setIsOpen] = useState(false);
    const dropdownRef = useRef<HTMLDivElement>(null);
    const queryClient = useQueryClient();
    const { isAuthenticated, isLoading } = useAuth();
    const isAuthReady = isAuthenticated && !isLoading;

    // Get unread count
    const { data: unreadCount = 0, refetch: refetchCount } = useQuery({
        queryKey: ['notifications', 'count'],
        queryFn: getUnreadCount,
        enabled: isAuthReady,
        retry: false,
        refetchInterval: isAuthReady ? 30000 : false, // Poll every 30 seconds
    });

    // Get recent notifications (for dropdown)
    const { data: notificationsData, refetch: refetchNotifications } = useQuery(
        {
            queryKey: ['notifications', 'recent'],
            queryFn: () => getNotifications('all', 1, 5),
            enabled: isOpen && isAuthReady, // Only fetch when dropdown is open
            retry: false,
        },
    );

    // Mark notification as read mutation
    const markAsReadMutation = useMutation({
        mutationFn: markNotificationAsRead,
        onSuccess: () => {
            queryClient.invalidateQueries({ queryKey: ['notifications'] });
            refetchCount();
        },
    });

    // Mark all as read mutation
    const markAllAsReadMutation = useMutation({
        mutationFn: markAllNotificationsAsRead,
        onSuccess: () => {
            queryClient.invalidateQueries({ queryKey: ['notifications'] });
            refetchCount();
        },
    });

    // Close dropdown when clicking outside
    useClickOutside(dropdownRef, () => setIsOpen(false), isOpen);

    // Refresh notifications when dropdown opens
    useEffect(() => {
        if (isOpen && isAuthReady) {
            refetchNotifications();
        }
    }, [isOpen, isAuthReady, refetchNotifications]);

    const handleNotificationClick = (notification: Notification) => {
        if (!notification.is_read) {
            markAsReadMutation.mutate(notification.id);
        }
        setIsOpen(false);
    };

    const handleMarkAllAsRead = (e: React.MouseEvent) => {
        e.preventDefault();
        e.stopPropagation();
        markAllAsReadMutation.mutate();
    };

    return (
        <div className='relative' ref={dropdownRef}>
            {/* Bell Icon Button */}
            <button
                onClick={() => setIsOpen(!isOpen)}
                className='hover:bg-muted relative p-2 transition-colors rounded-lg cursor-pointer'
                aria-label='Notifications'
            >
                {/* Bell Icon */}
                <svg
                    className='w-6 h-6'
                    fill='none'
                    stroke='currentColor'
                    viewBox='0 0 24 24'
                >
                    <path
                        strokeLinecap='round'
                        strokeLinejoin='round'
                        strokeWidth={2}
                        d='M15 17h5l-1.405-1.405A2.032 2.032 0 0118 14.158V11a6.002 6.002 0 00-4-5.659V5a2 2 0 10-4 0v.341C7.67 6.165 6 8.388 6 11v3.159c0 .538-.214 1.055-.595 1.436L4 17h5m6 0v1a3 3 0 11-6 0v-1m6 0H9'
                    />
                </svg>

                {/* Badge */}
                {unreadCount > 0 && (
                    <span className='min-w-5 absolute top-0 right-0 inline-flex items-center justify-center px-2 py-1 text-xs font-bold leading-none text-white transform translate-x-1/2 -translate-y-1/2 bg-red-600 rounded-full'>
                        {unreadCount > 99 ? '99+' : unreadCount}
                    </span>
                )}
            </button>

            {/* Dropdown */}
            {isOpen && (
                <div className='w-80 absolute right-0 z-50 mt-2 overflow-hidden bg-background border border-border rounded-md shadow-lg'>
                    {/* Header */}
                    <div className='flex items-center justify-between p-4 border-b border-border'>
                        <h3 className='text-lg font-semibold'>Notifications</h3>
                        {unreadCount > 0 && (
                            <button
                                onClick={handleMarkAllAsRead}
                                className='text-link hover:text-link text-sm cursor-pointer'
                            >
                                Mark all as read
                            </button>
                        )}
                    </div>

                    {/* Notifications List */}
                    <div className='max-h-96 overflow-y-auto'>
                        {(
                            notificationsData &&
                            notificationsData?.notifications?.length > 0
                        ) ?
                            notificationsData.notifications.map(
                                notification => (
                                    <NotificationItem
                                        key={notification.id}
                                        notification={notification}
                                        onClick={() =>
                                            handleNotificationClick(
                                                notification,
                                            )
                                        }
                                    />
                                ),
                            )
                        :   <div className='p-8 text-center text-muted-foreground'>
                                <svg
                                    className='w-12 h-12 mx-auto mb-2 opacity-50'
                                    fill='none'
                                    stroke='currentColor'
                                    viewBox='0 0 24 24'
                                >
                                    <path
                                        strokeLinecap='round'
                                        strokeLinejoin='round'
                                        strokeWidth={2}
                                        d='M15 17h5l-1.405-1.405A2.032 2.032 0 0118 14.158V11a6.002 6.002 0 00-4-5.659V5a2 2 0 10-4 0v.341C7.67 6.165 6 8.388 6 11v3.159c0 .538-.214 1.055-.595 1.436L4 17h5m6 0v1a3 3 0 11-6 0v-1m6 0H9'
                                    />
                                </svg>
                                <p className='text-sm'>No notifications yet</p>
                            </div>
                        }
                    </div>

                    {/* Footer */}
                    {notificationsData &&
                        notificationsData?.notifications?.length > 0 && (
                            <div className='bg-muted/50 p-3 border-t border-border'>
                                <Link
                                    to='/notifications'
                                    className='text-link hover:text-link block text-sm font-medium text-center cursor-pointer'
                                    onClick={() => setIsOpen(false)}
                                >
                                    See all notifications
                                </Link>
                            </div>
                        )}
                </div>
            )}
        </div>
    );
}
