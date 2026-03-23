import { useState } from 'react';
import { Link } from 'react-router-dom';
import { Settings } from 'lucide-react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { Helmet } from '@dr.pogodin/react-helmet';
import {
  getNotifications,
  markNotificationAsRead,
  markAllNotificationsAsRead,
  deleteNotification,
} from '../lib/notification-api';
import type { NotificationFilter } from '../types/notification';
import { Button, NotificationSkeleton, EmptyStateWithAction } from '../components/ui';
import { NotificationItem } from '../components/layout/NotificationItem';
import { Container } from '../components/layout';

export function NotificationsPage() {
  const [filter, setFilter] = useState<NotificationFilter>('all');
  const [page, setPage] = useState(1);
  const queryClient = useQueryClient();

  const { data, isLoading, isError } = useQuery({
    queryKey: ['notifications', 'list', filter, page],
    queryFn: () => getNotifications(filter, page, 20),
  });

  const markAsReadMutation = useMutation({
    mutationFn: markNotificationAsRead,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['notifications'] });
    },
  });

  const markAllAsReadMutation = useMutation({
    mutationFn: markAllNotificationsAsRead,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['notifications'] });
    },
  });

  const deleteMutation = useMutation({
    mutationFn: deleteNotification,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['notifications'] });
    },
  });

  const handleNotificationClick = (notificationId: string, isRead: boolean) => {
    if (!isRead) {
      markAsReadMutation.mutate(notificationId);
    }
  };

  const handleMarkAllAsRead = () => {
    markAllAsReadMutation.mutate();
  };

  return (
    <>
      <Helmet>
        <title>Notifications - clpr</title>
      </Helmet>

      <Container>
        <div className="max-w-4xl mx-auto py-8">
          {/* Header */}
          <div className="mb-6">
            <h1 className="text-3xl font-bold text-foreground mb-2">
              Notifications
            </h1>
            <p className="text-muted-foreground">
              Stay updated with your activity
            </p>
          </div>

          {/* Actions Bar */}
          <div className="bg-surface rounded-lg shadow-sm border border-border p-4 mb-6">
            <div className="flex flex-col sm:flex-row items-start sm:items-center justify-between gap-4">
              {/* Filter Tabs */}
              <div className="flex gap-2 flex-wrap">
                <Button
                  variant={filter === 'all' ? 'primary' : 'ghost'}
                  size="sm"
                  onClick={() => {
                    setFilter('all');
                    setPage(1);
                  }}
                >
                  All
                </Button>
                <Button
                  variant={filter === 'unread' ? 'primary' : 'ghost'}
                  size="sm"
                  onClick={() => {
                    setFilter('unread');
                    setPage(1);
                  }}
                >
                  Unread
                  {data && data.unread_count > 0 && (
                    <span className="ml-2 px-2 py-0.5 bg-red-600 text-white text-xs rounded-full">
                      {data.unread_count}
                    </span>
                  )}
                </Button>
                <Button
                  variant={filter === 'read' ? 'primary' : 'ghost'}
                  size="sm"
                  onClick={() => {
                    setFilter('read');
                    setPage(1);
                  }}
                >
                  Read
                </Button>
              </div>

              {/* Actions */}
              <div className="flex gap-2">
                {data && data.unread_count > 0 && (
                  <Button
                    variant="ghost"
                    size="sm"
                    onClick={handleMarkAllAsRead}
                    disabled={markAllAsReadMutation.isPending}
                  >
                    Mark all as read
                  </Button>
                )}
                <Link to="/notifications/preferences">
                  <Button variant="ghost" size="sm">
                    <Settings size={16} strokeWidth={1.75} className='mr-1.5 inline' /> Preferences
                  </Button>
                </Link>
              </div>
            </div>
          </div>

          {/* Notifications List */}
          <div className="bg-surface rounded-lg shadow-sm border border-border overflow-hidden">
            {isLoading ? (
              <div className="p-4">
                <NotificationSkeleton />
              </div>
            ) : isError ? (
              <div className="p-8">
                <EmptyStateWithAction
                  icon={
                    <svg className="w-16 h-16" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
                    </svg>
                  }
                  title="Failed to load notifications"
                  description="Something went wrong while loading your notifications."
                  primaryAction={{
                    label: "Try Again",
                    onClick: () => window.location.reload()
                  }}
                  secondaryAction={{
                    label: "Go Home",
                    href: "/"
                  }}
                />
              </div>
            ) : data && data.notifications && data.notifications.length > 0 ? (
              <>
                <div className="divide-y divide-border">
                  {data.notifications.map((notification) => (
                    <div key={notification.id} className="relative group">
                      <NotificationItem
                        notification={notification}
                        onClick={() => handleNotificationClick(notification.id, notification.is_read)}
                      />
                      {/* Delete button */}
                      <button
                        onClick={(e) => {
                          e.preventDefault();
                          e.stopPropagation();
                          if (window.confirm('Delete this notification?')) {
                            deleteMutation.mutate(notification.id);
                          }
                        }}
                        className="absolute top-4 right-4 opacity-0 group-hover:opacity-100 transition-opacity p-1 rounded hover:bg-surface-hover"
                        aria-label="Delete notification"
                      >
                        <svg
                          className="w-4 h-4 text-muted-foreground"
                          fill="none"
                          stroke="currentColor"
                          viewBox="0 0 24 24"
                        >
                          <path
                            strokeLinecap="round"
                            strokeLinejoin="round"
                            strokeWidth={2}
                            d="M6 18L18 6M6 6l12 12"
                          />
                        </svg>
                      </button>
                    </div>
                  ))}
                </div>

                {/* Pagination */}
                {(data.has_more || page > 1) && (
                  <div className="p-4 border-t border-border flex justify-center gap-2">
                    <Button
                      variant="ghost"
                      size="sm"
                      onClick={() => setPage((p) => Math.max(1, p - 1))}
                      disabled={page === 1}
                    >
                      Previous
                    </Button>
                    <span className="px-4 py-2 text-sm text-muted-foreground">
                      Page {page}
                    </span>
                    <Button
                      variant="ghost"
                      size="sm"
                      onClick={() => setPage((p) => p + 1)}
                      disabled={!data.has_more}
                    >
                      Next
                    </Button>
                  </div>
                )}
              </>
            ) : (
              <div className="p-8">
                <EmptyStateWithAction
                  icon={
                    <svg className="w-16 h-16" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15 17h5l-1.405-1.405A2.032 2.032 0 0118 14.158V11a6.002 6.002 0 00-4-5.659V5a2 2 0 10-4 0v.341C7.67 6.165 6 8.388 6 11v3.159c0 .538-.214 1.055-.595 1.436L4 17h5m6 0v1a3 3 0 11-6 0v-1m6 0H9" />
                    </svg>
                  }
                  title={
                    filter === 'unread'
                      ? "No unread notifications"
                      : filter === 'read'
                        ? "No read notifications"
                        : "No notifications"
                  }
                  description={
                    filter === 'unread'
                      ? "You're all caught up! Check back later for new notifications."
                      : filter === 'read'
                        ? "You haven't read any notifications yet."
                        : "You'll see notifications here when you get them."
                  }
                  primaryAction={
                    filter !== 'all' ? {
                      label: "View All Notifications",
                      onClick: () => {
                        setFilter('all');
                        setPage(1);
                      }
                    } : undefined
                  }
                  secondaryAction={{
                    label: "Browse Clips",
                    href: "/"
                  }}
                  tips={
                    filter === 'all' ? [
                      "Like or comment on clips to get notifications",
                      "Follow your favorite streamers",
                      "Submit clips to the community"
                    ] : undefined
                  }
                />
              </div>
            )}
          </div>
        </div>
      </Container>
    </>
  );
}
