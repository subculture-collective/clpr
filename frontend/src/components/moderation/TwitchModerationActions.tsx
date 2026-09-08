import { useState, useEffect } from 'react';
import { Button, Modal, ModalFooter, Alert, Input, TextArea } from '../ui';
import {
    banUserOnTwitch,
    unbanUserOnTwitch,
    getBanReasonTemplates,
    type TwitchModerationError,
    type BanReasonTemplate,
} from '../../lib/moderation-api';
import { getErrorMessage } from '../../lib/error-utils';
import { useAuth } from '../../context/AuthContext';
import { useToast } from '../../context/ToastContext';
import { Ban, ShieldCheck } from 'lucide-react';
import { useQueryClient } from '@tanstack/react-query';

export interface TwitchModerationActionsProps {
    /**
     * Twitch broadcaster ID (channel ID)
     */
    broadcasterID: string;
    /**
     * Twitch user ID to ban/unban
     */
    userID: string;
    /**
     * Username for display purposes
     */
    username?: string;
    /**
     * Whether the user is currently banned
     */
    isBanned?: boolean;
    /**
     * Whether the current user is the broadcaster
     */
    isBroadcaster?: boolean;
    /**
     * Whether the current user is a Twitch-recognized moderator
     */
    isTwitchModerator?: boolean;
    /**
     * Callback when action completes successfully
     */
    onSuccess?: () => void;
    /**
     * Disable trigger buttons when another modal is open
     */
    disableTriggers?: boolean;
    /**
     * Callback when a modal opens
     */
    onModalOpen?: () => void;
    /**
     * Callback when a modal closes
     */
    onModalClose?: () => void;
}

/**
 * TwitchModerationActions component for banning/unbanning users on Twitch
 *
 * Features:
 * - Permission gating: only visible to broadcaster or Twitch-recognized moderators
 * - Site moderators are view-only and cannot perform Twitch actions
 * - Optimistic loading states
 * - Structured error handling for scope, rate-limit, and unknown errors
 * - Confirmation modals
 * - Supports both permanent bans and timeouts
 */
/**
 * Helper function to determine if user can perform Twitch moderation actions
 */
function canUserPerformTwitchActions(
    isBroadcaster: boolean,
    isTwitchModerator: boolean,
): boolean {
    // Only broadcaster or Twitch-recognized moderators can perform actions.
    // Site moderators without these roles are view-only (their flags will both be false).
    return isBroadcaster || isTwitchModerator;
}

/**
 * Helper function to handle moderation API errors
 */
function handleModerationError(err: unknown): string {
    const apiError = err as { response?: { data?: TwitchModerationError } };
    const errorData = apiError.response?.data;

    if (!errorData?.code) {
        return getErrorMessage(err, 'Failed to perform action on Twitch');
    }

    switch (errorData.code) {
        case 'SITE_MODERATORS_READ_ONLY':
            return 'Site moderators cannot perform Twitch actions. You must be the channel broadcaster or a Twitch-recognized moderator.';
        case 'NOT_AUTHENTICATED':
            return 'You are not authenticated with Twitch. Please reconnect your Twitch account.';
        case 'INSUFFICIENT_SCOPES':
            return 'You do not have the required Twitch permissions. Please reconnect your Twitch account with moderator permissions.';
        case 'NOT_BROADCASTER':
            return 'Only the broadcaster can perform this action.';
        case 'RATE_LIMIT_EXCEEDED':
            return 'Rate limit exceeded. Please wait a moment and try again.';
        default:
            return (
                errorData.detail ||
                errorData.error ||
                'Failed to perform action on Twitch'
            );
    }
}

export function TwitchModerationActions({
    broadcasterID,
    userID,
    username,
    isBanned = false,
    isBroadcaster = false,
    isTwitchModerator = false,
    onSuccess,
    disableTriggers = false,
    onModalOpen,
    onModalClose,
}: TwitchModerationActionsProps) {
    const { user, refreshUser } = useAuth();
    const toast = useToast();
    const queryClient = useQueryClient();
    const [showBanModal, setShowBanModal] = useState(false);
    const [showUnbanModal, setShowUnbanModal] = useState(false);
    const [loading, setLoading] = useState(false);
    const [error, setError] = useState<string | null>(null);
    const [actionAlert, setActionAlert] = useState<{
        type: 'success' | 'error';
        message: string;
    } | null>(null);

    const resolvedIsBroadcaster =
        isBroadcaster;
    const resolvedIsTwitchModerator =
        isTwitchModerator;

    // Ban form state
    const [reason, setReason] = useState('');
    const [isPermanent, setIsPermanent] = useState(true);
    const [duration, setDuration] = useState('600'); // Default 10 minutes in seconds
    const [customDuration, setCustomDuration] = useState(false);

    // Template state
    const [templates, setTemplates] = useState<BanReasonTemplate[]>([]);
    const [selectedTemplate, setSelectedTemplate] = useState<string>('');
    const [templatesLoading, setTemplatesLoading] = useState(false);

    // Fetch templates when ban modal opens
    useEffect(() => {
        if (showBanModal && templates.length === 0) {
            setTemplatesLoading(true);
            getBanReasonTemplates(broadcasterID, true)
                .then(response => {
                    setTemplates(response.templates);
                })
                .catch(err => {
                    console.error('Failed to load templates:', err);
                })
                .finally(() => {
                    setTemplatesLoading(false);
                });
        }
    }, [showBanModal, broadcasterID, templates.length]);

    // Apply template when selected
    useEffect(() => {
        if (selectedTemplate && templates.length > 0) {
            const template = templates.find(t => t.id === selectedTemplate);
            if (template) {
                setReason(template.reason);
                if (
                    template.duration_seconds === null ||
                    template.duration_seconds === undefined
                ) {
                    setIsPermanent(true);
                } else {
                    setIsPermanent(false);
                    setDuration(template.duration_seconds.toString());
                    setCustomDuration(false);
                }
            }
        }
    }, [selectedTemplate, templates]);

    // Permission check
    const canPerformTwitchActions = canUserPerformTwitchActions(
        resolvedIsBroadcaster,
        resolvedIsTwitchModerator,
    );

    // Hide triggers when a modal is open or parent disables them
    const triggersHidden = disableTriggers || showBanModal || showUnbanModal;

    // Notify parent when all modals close
    useEffect(() => {
        if (!showBanModal && !showUnbanModal) {
            onModalClose?.();
        }
    }, [showBanModal, showUnbanModal, onModalClose]);

    // Show read-only notice for site moderators without Twitch permissions
    if (!canPerformTwitchActions) {
        if (user?.role === 'moderator') {
            return (
                <Alert role='alert' variant='warning' className='mb-2'>
                    Site moderators cannot perform Twitch actions. Only the
                    channel broadcaster or Twitch-recognized moderators can ban
                    or unban users on Twitch.
                </Alert>
            );
        }
        return null;
    }

    const handleBan = async () => {
        if (!user) return;

        // Validate duration for timeouts
        if (!isPermanent) {
            const durationNum = parseInt(duration, 10);
            if (
                isNaN(durationNum) ||
                durationNum <= 0 ||
                durationNum > 1209600
            ) {
                setError(
                    'Duration must be between 1 and 1,209,600 seconds (14 days).',
                );
                return;
            }
        }

        try {
            setLoading(true);
            setError(null);

            await banUserOnTwitch({
                broadcasterID,
                userID,
                reason: reason.trim() || undefined,
                duration: isPermanent ? undefined : parseInt(duration, 10),
            });

            setShowBanModal(false);
            onModalClose?.();
            setReason('');
            setIsPermanent(true);
            setDuration('600');

            setCustomDuration(false);

            const successText =
                isPermanent ?
                    `${username || 'User'} has been permanently banned on Twitch`
                :   `${username || 'User'} has been timed out on Twitch`;

            // Show success toast and inline alert for E2E visibility
            toast.success(successText);
            setActionAlert({ type: 'success', message: successText });

            // Invalidate relevant queries to update UI
            queryClient.invalidateQueries({
                queryKey: ['banStatus', broadcasterID],
            });
            queryClient.invalidateQueries({
                queryKey: ['user-profile-by-username'],
            });

            // Refresh user data from auth context
            await refreshUser();

            if (onSuccess) {
                onSuccess();
            }
        } catch (err) {
            const message = handleModerationError(err);
            setError(message);
            setActionAlert({ type: 'error', message });
        } finally {
            setLoading(false);
        }
    };

    const handleUnban = async () => {
        if (!user) return;

        try {
            setLoading(true);
            setError(null);

            await unbanUserOnTwitch({
                broadcasterID,
                userID,
            });

            setShowUnbanModal(false);
            onModalClose?.();

            const successText = `${
                username || 'User'
            } has been unbanned on Twitch`;
            // Show success toast and inline alert for E2E visibility
            toast.success(successText);
            setActionAlert({ type: 'success', message: successText });

            // Invalidate relevant queries to update UI
            queryClient.invalidateQueries({
                queryKey: ['banStatus', broadcasterID],
            });
            queryClient.invalidateQueries({
                queryKey: ['user-profile-by-username'],
            });

            // Refresh user data from auth context
            await refreshUser();

            if (onSuccess) {
                onSuccess();
            }
        } catch (err) {
            const message = handleModerationError(err);
            setError(message);
            setActionAlert({ type: 'error', message });
        } finally {
            setLoading(false);
        }
    };

    const resetBanForm = () => {
        setReason('');
        setIsPermanent(true);
        setDuration('600');
        setCustomDuration(false);
        setError(null);
        setSelectedTemplate('');
    };

    return (
        <>
            {actionAlert && (
                <Alert
                    role='alert'
                    variant={
                        actionAlert.type === 'success' ? 'success' : 'error'
                    }
                    className='mb-2'
                    data-testid={`twitch-action-${actionAlert.type}-alert`}
                >
                    {actionAlert.message}
                </Alert>
            )}
            {!triggersHidden &&
                (!isBanned ?
                    <Button
                        variant='danger'
                        size='sm'
                        leftIcon={<Ban className='h-4 w-4' />}
                        onClick={() => {
                            resetBanForm();
                            setActionAlert(null);
                            setShowBanModal(true);
                            onModalOpen?.();
                        }}
                        aria-label={`Ban ${username || 'user'} on Twitch`}
                    >
                        Ban on Twitch
                    </Button>
                :   <Button
                        variant='secondary'
                        size='sm'
                        leftIcon={<ShieldCheck className='h-4 w-4' />}
                        onClick={() => {
                            setError(null);
                            setActionAlert(null);
                            setShowUnbanModal(true);
                            onModalOpen?.();
                        }}
                        aria-label={`Unban ${username || 'user'} on Twitch`}
                    >
                        Unban on Twitch
                    </Button>)}

            {/* Ban Modal */}
            <Modal
                open={showBanModal}
                onClose={() => {
                    if (!loading) {
                        setShowBanModal(false);
                        resetBanForm();
                        onModalClose?.();
                    }
                }}
                title='Ban User on Twitch'
            >
                <div className='space-y-4'>
                    {error && (
                        <Alert variant='error' role='alert'>
                            {error}
                        </Alert>
                    )}

                    <div>
                        <p className='mb-4 text-sm text-gray-600 dark:text-gray-400'>
                            You are about to ban{' '}
                            <strong>{username || userID}</strong> on Twitch.
                            This will prevent them from chatting or interacting
                            in the Twitch channel.
                        </p>

                        <div className='space-y-4'>
                            {/* Template Selection */}
                            <div>
                                <label
                                    htmlFor='template-select'
                                    className='mb-2 block text-sm font-medium text-gray-700 dark:text-gray-300'
                                >
                                    Ban Reason Template (Optional)
                                </label>
                                <select
                                    id='template-select'
                                    value={selectedTemplate}
                                    onChange={e =>
                                        setSelectedTemplate(e.target.value)
                                    }
                                    disabled={loading || templatesLoading}
                                    className='w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md shadow-sm focus:outline-none focus:ring-primary-500 focus:border-primary-500 dark:bg-gray-700 dark:text-white'
                                >
                                    <option value=''>
                                        -- Select a template --
                                    </option>
                                    {templates.map(template => (
                                        <option
                                            key={template.id}
                                            value={template.id}
                                        >
                                            {template.name}{' '}
                                            {template.is_default ?
                                                '(Default)'
                                            :   ''}
                                        </option>
                                    ))}
                                </select>
                                {templatesLoading && (
                                    <p className='mt-1 text-xs text-gray-500 dark:text-gray-400'>
                                        Loading templates...
                                    </p>
                                )}
                            </div>

                            <div>
                                <label className='mb-2 block text-sm font-medium text-gray-700 dark:text-gray-300'>
                                    Ban Type
                                </label>
                                <div className='space-y-2'>
                                    <div className='flex items-center space-x-2'>
                                        <input
                                            type='radio'
                                            id='ban-type-permanent'
                                            name='banType'
                                            checked={isPermanent}
                                            onChange={() =>
                                                setIsPermanent(true)
                                            }
                                            className='h-4 w-4 text-link'
                                            disabled={loading}
                                        />
                                        <label
                                            htmlFor='ban-type-permanent'
                                            className='text-sm text-gray-700 dark:text-gray-300'
                                        >
                                            Permanent Ban
                                        </label>
                                    </div>
                                    <div className='flex items-center space-x-2'>
                                        <input
                                            type='radio'
                                            id='ban-type-timeout'
                                            name='banType'
                                            checked={!isPermanent}
                                            onChange={() =>
                                                setIsPermanent(false)
                                            }
                                            className='h-4 w-4 text-link'
                                            disabled={loading}
                                        />
                                        <label
                                            htmlFor='ban-type-timeout'
                                            className='text-sm text-gray-700 dark:text-gray-300'
                                        >
                                            Timeout (Temporary)
                                        </label>
                                    </div>
                                </div>
                            </div>

                            {!isPermanent && (
                                <div>
                                    <label
                                        htmlFor='duration'
                                        className='mb-2 block text-sm font-medium text-gray-700 dark:text-gray-300'
                                    >
                                        Duration
                                    </label>
                                    <div className='space-y-3'>
                                        {/* Preset duration buttons */}
                                        <div className='grid grid-cols-2 gap-2'>
                                            <button
                                                type='button'
                                                onClick={() => {
                                                    setDuration('3600');
                                                    setCustomDuration(false);
                                                }}
                                                disabled={loading}
                                                className={`px-3 py-2 text-sm border rounded-md transition-colors ${
                                                    (
                                                        duration === '3600' &&
                                                        !customDuration
                                                    ) ?
                                                        'border-primary-600 bg-primary-50 dark:bg-primary-900/20 text-link dark:text-primary-300'
                                                    :   'border-gray-300 dark:border-gray-600 hover:bg-gray-50 dark:hover:bg-gray-700'
                                                }`}
                                            >
                                                1 hour
                                            </button>
                                            <button
                                                type='button'
                                                onClick={() => {
                                                    setDuration('86400');
                                                    setCustomDuration(false);
                                                }}
                                                disabled={loading}
                                                className={`px-3 py-2 text-sm border rounded-md transition-colors ${
                                                    (
                                                        duration === '86400' &&
                                                        !customDuration
                                                    ) ?
                                                        'border-primary-600 bg-primary-50 dark:bg-primary-900/20 text-link dark:text-primary-300'
                                                    :   'border-gray-300 dark:border-gray-600 hover:bg-gray-50 dark:hover:bg-gray-700'
                                                }`}
                                            >
                                                24 hours
                                            </button>
                                            <button
                                                type='button'
                                                onClick={() => {
                                                    setDuration('604800');
                                                    setCustomDuration(false);
                                                }}
                                                disabled={loading}
                                                className={`px-3 py-2 text-sm border rounded-md transition-colors ${
                                                    (
                                                        duration === '604800' &&
                                                        !customDuration
                                                    ) ?
                                                        'border-primary-600 bg-primary-50 dark:bg-primary-900/20 text-link dark:text-primary-300'
                                                    :   'border-gray-300 dark:border-gray-600 hover:bg-gray-50 dark:hover:bg-gray-700'
                                                }`}
                                            >
                                                7 days
                                            </button>
                                            <button
                                                type='button'
                                                onClick={() => {
                                                    setDuration('1209600');
                                                    setCustomDuration(false);
                                                }}
                                                disabled={loading}
                                                className={`px-3 py-2 text-sm border rounded-md transition-colors ${
                                                    (
                                                        duration ===
                                                            '1209600' &&
                                                        !customDuration
                                                    ) ?
                                                        'border-primary-600 bg-primary-50 dark:bg-primary-900/20 text-link dark:text-primary-300'
                                                    :   'border-gray-300 dark:border-gray-600 hover:bg-gray-50 dark:hover:bg-gray-700'
                                                }`}
                                            >
                                                14 days
                                            </button>
                                        </div>

                                        {/* Custom duration input */}
                                        <div>
                                            <button
                                                type='button'
                                                onClick={() => {
                                                    setCustomDuration(true);
                                                    setDuration('600');
                                                }}
                                                disabled={loading}
                                                className={`w-full px-3 py-2 text-sm border rounded-md transition-colors text-left ${
                                                    customDuration ?
                                                        'border-primary-600 bg-primary-50 dark:bg-primary-900/20 text-link dark:text-primary-300'
                                                    :   'border-gray-300 dark:border-gray-600 hover:bg-gray-50 dark:hover:bg-gray-700'
                                                }`}
                                            >
                                                Custom duration
                                            </button>
                                            {customDuration && (
                                                <div className='mt-2'>
                                                    <label
                                                        htmlFor='custom-duration'
                                                        className='block text-xs text-gray-600 dark:text-gray-400 mb-1'
                                                    >
                                                        Duration (seconds)
                                                    </label>
                                                    <Input
                                                        id='custom-duration'
                                                        type='number'
                                                        min='1'
                                                        max='1209600'
                                                        value={duration}
                                                        onChange={e =>
                                                            setDuration(
                                                                e.target.value,
                                                            )
                                                        }
                                                        disabled={loading}
                                                        placeholder='Enter seconds'
                                                    />
                                                    <p className='mt-1 text-xs text-gray-500 dark:text-gray-400'>
                                                        Max: 1,209,600 seconds
                                                        (14 days)
                                                    </p>
                                                </div>
                                            )}
                                        </div>
                                    </div>
                                </div>
                            )}

                            <div>
                                <label
                                    htmlFor='reason'
                                    className='mb-2 block text-sm font-medium text-gray-700 dark:text-gray-300'
                                >
                                    Reason (Optional)
                                </label>
                                <TextArea
                                    id='reason'
                                    value={reason}
                                    onChange={e => setReason(e.target.value)}
                                    disabled={loading}
                                    placeholder='Reason for ban...'
                                    rows={3}
                                    maxLength={500}
                                />
                            </div>
                        </div>
                    </div>
                </div>

                <ModalFooter>
                    <Button
                        variant='ghost'
                        onClick={() => {
                            setShowBanModal(false);
                            resetBanForm();
                        }}
                        disabled={loading}
                    >
                        Cancel
                    </Button>
                    <Button
                        variant='danger'
                        onClick={handleBan}
                        loading={loading}
                        disabled={
                            loading ||
                            (!isPermanent &&
                                (!duration ||
                                    parseInt(duration, 10) <= 0 ||
                                    parseInt(duration, 10) > 1209600))
                        }
                        data-testid='confirm-ban-action-button'
                    >
                        {isPermanent ? 'Ban User' : 'Timeout User'}
                    </Button>
                </ModalFooter>
            </Modal>

            {/* Unban Modal */}
            <Modal
                open={showUnbanModal}
                onClose={() => {
                    if (!loading) {
                        setShowUnbanModal(false);
                        setError(null);
                        onModalClose?.();
                    }
                }}
                title='Unban User on Twitch'
            >
                <div className='space-y-4'>
                    {error && (
                        <Alert variant='error' role='alert'>
                            {error}
                        </Alert>
                    )}

                    <p className='text-sm text-gray-600 dark:text-gray-400'>
                        Are you sure you want to unban{' '}
                        <strong>{username || userID}</strong> on Twitch? This
                        will allow them to chat and interact in the Twitch
                        channel again.
                    </p>
                </div>

                <ModalFooter>
                    <Button
                        variant='ghost'
                        onClick={() => {
                            setShowUnbanModal(false);
                            setError(null);
                        }}
                        disabled={loading}
                    >
                        Cancel
                    </Button>
                    <Button
                        variant='secondary'
                        onClick={handleUnban}
                        loading={loading}
                        disabled={loading}
                        data-testid='confirm-unban-button'
                    >
                        Unban User
                    </Button>
                </ModalFooter>
            </Modal>
        </>
    );
}
