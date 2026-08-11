import { useState, useEffect } from 'react';
import { Bell, BellOff } from 'lucide-react';
import { Button } from '../ui/Button';
import { useAuth } from '../../context/AuthContext';
import {
  followStreamer,
  unfollowStreamer,
  getStreamFollowStatus,
  type StreamFollowStatus,
} from '../../lib/stream-api';
import { useToast } from '../../context/ToastContext';

interface StreamFollowButtonProps {
  streamerUsername: string;
  className?: string;
}

export function StreamFollowButton({ streamerUsername, className }: StreamFollowButtonProps) {
  const { user } = useAuth();
  const [followStatus, setFollowStatus] = useState<StreamFollowStatus | null>(null);
  const [isLoading, setIsLoading] = useState(false);
  const [isChecking, setIsChecking] = useState(true);

  // Check follow status on mount
  useEffect(() => {
    const checkStatus = async () => {
      if (!user) {
        setIsChecking(false);
        return;
      }

      try {
        const status = await getStreamFollowStatus(streamerUsername);
        setFollowStatus(status);
      } catch {
        // Not following or error - treat as not following
        setFollowStatus({ following: false, notifications_enabled: false });
      } finally {
        setIsChecking(false);
      }
    };

    checkStatus();
  }, [user, streamerUsername]);

  const { addToast } = useToast();

  const handleFollow = async () => {
    if (!user) {
      addToast('Please log in to follow creators', 'error');
      return;
    }

    setIsLoading(true);
    try {
      const result = await followStreamer(streamerUsername, { notifications_enabled: true });
      setFollowStatus({
        following: result.following,
        notifications_enabled: result.notifications_enabled,
      });
      addToast(result.message || `Following ${streamerUsername}`, 'success');
    } catch (error) {
      console.error('Failed to follow streamer:', error);
      addToast('Failed to follow streamer', 'error');
    } finally {
      setIsLoading(false);
    }
  };

  const handleUnfollow = async () => {
    if (!user) return;

    setIsLoading(true);
    try {
      const result = await unfollowStreamer(streamerUsername);
      setFollowStatus({ following: false, notifications_enabled: false });
      addToast(result.message || `Unfollowed ${streamerUsername}`, 'success');
    } catch (error) {
      console.error('Failed to unfollow streamer:', error);
      addToast('Failed to unfollow streamer', 'error');
    } finally {
      setIsLoading(false);
    }
  };

  if (!user || isChecking) {
    return null; // Don't show button if not logged in or still checking
  }

  const isFollowing = followStatus?.following || false;
  const hasNotifications = followStatus?.notifications_enabled || false;

  return (
    <Button
      onClick={isFollowing ? handleUnfollow : handleFollow}
      disabled={isLoading}
      loading={isLoading}
      variant={isFollowing ? 'secondary' : 'primary'}
      className={className}
      leftIcon={isFollowing ? (hasNotifications ? <Bell className="h-4 w-4" /> : <BellOff className="h-4 w-4" />) : <Bell className="h-4 w-4" />}
    >
      {isFollowing ? 'Following' : 'Follow'}
    </Button>
  );
}
