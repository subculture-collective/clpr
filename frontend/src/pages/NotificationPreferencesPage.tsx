import { useState, useEffect } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { Helmet } from '@dr.pogodin/react-helmet';
import { Link } from 'react-router-dom';
import {
  getNotificationPreferences,
  updateNotificationPreferences,
  resetNotificationPreferences,
} from '../lib/notification-api';
import type { NotificationPreferences } from '../types/notification';
import { Button } from '../components/ui';
import { Container } from '../components/layout';

export function NotificationPreferencesPage() {
  const queryClient = useQueryClient();
  const [isSaving, setIsSaving] = useState(false);
  const [isResetting, setIsResetting] = useState(false);
  const [formData, setFormData] = useState<Partial<NotificationPreferences>>({});

  const { data: preferences, isLoading } = useQuery({
    queryKey: ['notifications', 'preferences'],
    queryFn: getNotificationPreferences,
  });

  // Update form data when preferences load
  useEffect(() => {
    if (preferences) {
      queueMicrotask(() => {
        setFormData(preferences);
      });
    }
  }, [preferences]);

  const updateMutation = useMutation({
    mutationFn: updateNotificationPreferences,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['notifications'] });
      setIsSaving(false);
    },
    onError: () => {
      setIsSaving(false);
    },
  });

  const resetMutation = useMutation({
    mutationFn: resetNotificationPreferences,
    onSuccess: (data) => {
      setFormData(data);
      queryClient.invalidateQueries({ queryKey: ['notifications'] });
      setIsResetting(false);
    },
    onError: () => {
      setIsResetting(false);
    },
  });

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    setIsSaving(true);
    updateMutation.mutate(formData);
  };

  const handleReset = () => {
    if (confirm('Are you sure you want to reset all notification preferences to defaults?')) {
      setIsResetting(true);
      resetMutation.mutate();
    }
  };

  const handleToggle = (field: keyof NotificationPreferences) => {
    setFormData((prev) => ({
      ...prev,
      [field]: !(prev[field] as boolean),
    }));
  };

  const handleEmailDigestChange = (value: 'immediate' | 'daily' | 'weekly' | 'never') => {
    setFormData((prev) => ({
      ...prev,
      email_digest: value,
    }));
  };

  if (isLoading) {
    return (
      <Container>
        <div className="max-w-2xl mx-auto py-8">
          <div className="text-center py-12">
            <div className="inline-block animate-spin rounded-full h-8 w-8 border-4 border-border border-t-primary-600"></div>
            <p className="mt-4 text-muted-foreground">Loading preferences...</p>
          </div>
        </div>
      </Container>
    );
  }

  return (
    <>
      <Helmet>
        <title>Notification Preferences - clpr</title>
      </Helmet>

      <Container>
        <div className="max-w-2xl mx-auto py-8">
          {/* Header */}
          <div className="mb-6">
            <Link
              to="/notifications"
              className="text-sm text-primary-600 hover:text-primary-700 dark:text-primary-400 dark:hover:text-primary-300 mb-2 inline-block"
            >
              ← Back to Notifications
            </Link>
            <h1 className="text-3xl font-bold text-foreground mb-2">
              Notification Preferences
            </h1>
            <p className="text-muted-foreground">
              Customize how and when you receive notifications
            </p>
          </div>

          {/* Form */}
          <form onSubmit={handleSubmit} className="space-y-6">
            {/* Master Toggle */}
            <div className="bg-surface rounded-lg shadow-sm border border-border p-6">
              <h2 className="text-lg font-semibold text-foreground mb-4">
                Master Settings
              </h2>

              <div className="space-y-4">
                <ToggleSwitch
                  label="Enable All Email Notifications"
                  description="Master toggle to enable or disable all email notifications"
                  checked={formData.email_enabled ?? true}
                  onChange={() => handleToggle('email_enabled')}
                />

                {formData.email_enabled && (
                  <div className="ml-6 pt-2 space-y-2">
                    <label className="block text-sm font-medium text-foreground">
                      Email Frequency
                    </label>
                    <select
                      value={formData.email_digest ?? 'immediate'}
                      onChange={(e) =>
                        handleEmailDigestChange(e.target.value as 'immediate' | 'daily' | 'weekly' | 'never')
                      }
                      className="block w-full rounded-md border-border bg-surface text-foreground shadow-sm focus:border-primary-500 focus:ring-primary-500"
                    >
                      <option value="immediate">Immediate</option>
                      <option value="daily">Daily Digest</option>
                      <option value="weekly">Weekly Digest</option>
                      <option value="never">Never (Disabled)</option>
                    </select>
                  </div>
                )}

                <ToggleSwitch
                  label="In-App Notifications"
                  description="Show notifications within the application"
                  checked={formData.in_app_enabled ?? true}
                  onChange={() => handleToggle('in_app_enabled')}
                />
              </div>
            </div>

            {/* Account & Security */}
            <div className="bg-surface rounded-lg shadow-sm border border-border p-6">
              <h2 className="text-lg font-semibold text-foreground mb-4">
                Account &amp; Security
              </h2>
              <p className="text-sm text-muted-foreground mb-4">
                Important security alerts for your account
              </p>

              <div className="space-y-4">
                <ToggleSwitch
                  label="Login from New Device"
                  description="Notify when your account is accessed from a new device"
                  checked={formData.notify_login_new_device ?? true}
                  onChange={() => handleToggle('notify_login_new_device')}
                />

                <ToggleSwitch
                  label="Failed Login Attempts"
                  description="Alert when there are failed login attempts on your account"
                  checked={formData.notify_failed_login ?? true}
                  onChange={() => handleToggle('notify_failed_login')}
                />

                <ToggleSwitch
                  label="Password Changed"
                  description="Notify when your password is changed"
                  checked={formData.notify_password_changed ?? true}
                  onChange={() => handleToggle('notify_password_changed')}
                />

                <ToggleSwitch
                  label="Email Verified/Changed"
                  description="Notify when your email is verified or changed"
                  checked={formData.notify_email_changed ?? true}
                  onChange={() => handleToggle('notify_email_changed')}
                />
              </div>
            </div>

            {/* Content Notifications */}
            <div className="bg-surface rounded-lg shadow-sm border border-border p-6">
              <h2 className="text-lg font-semibold text-foreground mb-4">
                Content
              </h2>
              <p className="text-sm text-muted-foreground mb-4">
                Notifications about content and submissions
              </p>

              <div className="space-y-4">
                <ToggleSwitch
                  label="Reply to My Comment"
                  description="When someone replies to your comment"
                  checked={formData.notify_replies ?? true}
                  onChange={() => handleToggle('notify_replies')}
                />

                <ToggleSwitch
                  label="Mention in Comment"
                  description="When someone mentions you in a comment"
                  checked={formData.notify_mentions ?? true}
                  onChange={() => handleToggle('notify_mentions')}
                />

                <ToggleSwitch
                  label="Content Trending"
                  description="When your content starts trending"
                  checked={formData.notify_content_trending ?? false}
                  onChange={() => handleToggle('notify_content_trending')}
                />

                <ToggleSwitch
                  label="Content Flagged"
                  description="When your content is flagged for review"
                  checked={formData.notify_content_flagged ?? true}
                  onChange={() => handleToggle('notify_content_flagged')}
                />

                <ToggleSwitch
                  label="Vote Milestones"
                  description="When your comments reach vote milestones (10, 25, 50, etc.)"
                  checked={formData.notify_votes ?? false}
                  onChange={() => handleToggle('notify_votes')}
                />

                <ToggleSwitch
                  label="Favorited Clip Comments"
                  description="When someone comments on a clip you favorited"
                  checked={formData.notify_favorited_clip_comment ?? true}
                  onChange={() => handleToggle('notify_favorited_clip_comment')}
                />
              </div>
            </div>

            {/* Community Notifications */}
            <div className="bg-surface rounded-lg shadow-sm border border-border p-6">
              <h2 className="text-lg font-semibold text-foreground mb-4">
                Community
              </h2>
              <p className="text-sm text-muted-foreground mb-4">
                Notifications about community interactions
              </p>

              <div className="space-y-4">
                <ToggleSwitch
                  label="Community Moderator Message"
                  description="Messages from community moderators"
                  checked={formData.notify_moderator_message ?? true}
                  onChange={() => handleToggle('notify_moderator_message')}
                />

                <ToggleSwitch
                  label="User Followed Me"
                  description="When another user follows you"
                  checked={formData.notify_user_followed ?? true}
                  onChange={() => handleToggle('notify_user_followed')}
                />

                <ToggleSwitch
                  label="Comment on My Content"
                  description="When someone comments on your content"
                  checked={formData.notify_comment_on_content ?? true}
                  onChange={() => handleToggle('notify_comment_on_content')}
                />

                <ToggleSwitch
                  label="Reply to Discussion I'm In"
                  description="Replies to discussions you're participating in"
                  checked={formData.notify_discussion_reply ?? true}
                  onChange={() => handleToggle('notify_discussion_reply')}
                />

                <ToggleSwitch
                  label="Badges &amp; Achievements"
                  description="When you earn a badge or achievement"
                  checked={formData.notify_badges ?? true}
                  onChange={() => handleToggle('notify_badges')}
                />

                <ToggleSwitch
                  label="Rank Up"
                  description="When you advance to a new rank"
                  checked={formData.notify_rank_up ?? true}
                  onChange={() => handleToggle('notify_rank_up')}
                />

                <ToggleSwitch
                  label="Moderation Actions"
                  description="When your content is removed, warnings, or account actions"
                  checked={formData.notify_moderation ?? true}
                  onChange={() => handleToggle('notify_moderation')}
                  disabled
                />
              </div>
            </div>

            {/* Creator Notifications */}
            <div className="bg-surface rounded-lg shadow-sm border border-border p-6">
              <h2 className="text-lg font-semibold text-foreground mb-4">
                Creator Notifications
              </h2>
              <p className="text-sm text-muted-foreground mb-4">
                Get notified about activity on clips you've created
              </p>

              <div className="space-y-4">
                <ToggleSwitch
                  label="Clip Approvals"
                  description="When your submitted clip is approved"
                  checked={formData.notify_clip_approved ?? true}
                  onChange={() => handleToggle('notify_clip_approved')}
                />

                <ToggleSwitch
                  label="Clip Rejections"
                  description="When your submitted clip is rejected"
                  checked={formData.notify_clip_rejected ?? true}
                  onChange={() => handleToggle('notify_clip_rejected')}
                />

                <ToggleSwitch
                  label="Comments on Your Clips"
                  description="When someone comments on a clip you created"
                  checked={formData.notify_clip_comments ?? true}
                  onChange={() => handleToggle('notify_clip_comments')}
                />

                <ToggleSwitch
                  label="Clip Threshold Alerts"
                  description="When your clips reach view/vote milestones (100, 1K, 10K views, etc.)"
                  checked={formData.notify_clip_threshold ?? true}
                  onChange={() => handleToggle('notify_clip_threshold')}
                />
              </div>
            </div>

            {/* Stream & Broadcaster Notifications */}
            <div className="bg-surface rounded-lg shadow-sm border border-border p-6">
              <h2 className="text-lg font-semibold text-foreground mb-4">
                Stream & Broadcaster Notifications
              </h2>
              <p className="text-sm text-muted-foreground mb-4">
                Get notified when streamers and broadcasters you follow go live
              </p>

              <div className="space-y-4">
                <ToggleSwitch
                  label="Broadcaster Live"
                  description="When broadcasters you follow on the platform go live"
                  checked={formData.notify_broadcaster_live ?? true}
                  onChange={() => handleToggle('notify_broadcaster_live')}
                />

                <ToggleSwitch
                  label="Stream Live"
                  description="When streamers you follow go live on Twitch"
                  checked={formData.notify_stream_live ?? true}
                  onChange={() => handleToggle('notify_stream_live')}
                />
              </div>
            </div>

            {/* Global Preferences */}
            <div className="bg-surface rounded-lg shadow-sm border border-border p-6">
              <h2 className="text-lg font-semibold text-foreground mb-4">
                Global &amp; Marketing
              </h2>
              <p className="text-sm text-muted-foreground mb-4">
                Platform updates and promotional emails
              </p>

              <div className="space-y-4">
                <ToggleSwitch
                  label="Marketing &amp; Product Emails"
                  description="Receive emails about new features, tips, and promotions"
                  checked={formData.notify_marketing ?? false}
                  onChange={() => handleToggle('notify_marketing')}
                />

                <ToggleSwitch
                  label="Policy Updates"
                  description="Important updates to terms of service and privacy policy"
                  checked={formData.notify_policy_updates ?? true}
                  onChange={() => handleToggle('notify_policy_updates')}
                />

                <ToggleSwitch
                  label="Platform Announcements"
                  description="News about platform updates and changes"
                  checked={formData.notify_platform_announcements ?? true}
                  onChange={() => handleToggle('notify_platform_announcements')}
                />
              </div>
            </div>

            {/* Actions */}
            <div className="flex justify-between items-center gap-3">
              <Button
                type="button"
                variant="ghost"
                onClick={handleReset}
                disabled={isResetting}
              >
                {isResetting ? 'Resetting...' : 'Reset to Defaults'}
              </Button>

              <div className="flex gap-3">
                <Link to="/notifications">
                  <Button variant="ghost">Cancel</Button>
                </Link>
                <Button type="submit" variant="primary" disabled={isSaving}>
                  {isSaving ? 'Saving...' : 'Save Preferences'}
                </Button>
              </div>
            </div>

            {updateMutation.isSuccess && (
              <div className="rounded-lg bg-green-50 dark:bg-green-900/20 p-4 text-green-800 dark:text-green-200">
                Preferences saved successfully!
              </div>
            )}

            {resetMutation.isSuccess && (
              <div className="rounded-lg bg-green-50 dark:bg-green-900/20 p-4 text-green-800 dark:text-green-200">
                Preferences reset to defaults successfully!
              </div>
            )}

            {(updateMutation.isError || resetMutation.isError) && (
              <div className="rounded-lg bg-red-50 dark:bg-red-900/20 p-4 text-red-800 dark:text-red-200">
                Failed to update preferences. Please try again.
              </div>
            )}
          </form>
        </div>
      </Container>
    </>
  );
}

// Toggle Switch Component
interface ToggleSwitchProps {
  label: string;
  description?: string;
  checked: boolean;
  onChange: () => void;
  disabled?: boolean;
}

function ToggleSwitch({ label, description, checked, onChange, disabled = false }: ToggleSwitchProps) {
  return (
    <div className="flex items-start justify-between">
      <div className="flex-1">
        <label className="text-sm font-medium text-foreground">{label}</label>
        {description && (
          <p className="text-sm text-muted-foreground mt-0.5">{description}</p>
        )}
      </div>
      <button
        type="button"
        onClick={onChange}
        disabled={disabled}
        className={`relative inline-flex h-6 w-11 flex-shrink-0 cursor-pointer rounded-full border-2 border-transparent transition-colors duration-200 ease-in-out focus:outline-none focus:ring-2 focus:ring-primary-500 focus:ring-offset-2 ${
          checked ? 'bg-primary-600' : 'bg-muted'
        } ${disabled ? 'opacity-50 cursor-not-allowed' : ''}`}
      >
        <span
          className={`pointer-events-none inline-block h-5 w-5 transform rounded-full bg-white shadow ring-0 transition duration-200 ease-in-out ${
            checked ? 'translate-x-5' : 'translate-x-0'
          }`}
        />
      </button>
    </div>
  );
}
