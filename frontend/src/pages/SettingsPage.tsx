import { useQuery, useQueryClient } from '@tanstack/react-query';
import { useEffect, useState, useRef } from 'react';
import { Helmet } from '@dr.pogodin/react-helmet';
import { Link } from 'react-router-dom';
import { Alert, Button, Card, CardBody, CardHeader, Container, Input, Modal, Stack, Toggle } from '../components';
import { useAuth } from '../context/AuthContext';
import { useConsent } from '../context/ConsentContext';
import type { UpdateProfileRequest, UpdateSettingsRequest } from '../lib/user-settings-api';
import { getUserSettings, updateProfile, updateUserSettings } from '../lib/user-settings-api';
import {
    getSubscription,
    cancelSubscription,
    reactivateSubscription,
    createPortalSession,
} from '../lib/subscription-api';
import type { Subscription } from '../lib/subscription-api';

// Constants for billing period calculation
const DAYS_IN_YEAR = 365; // Threshold for yearly subscription (allowing minimal variance)
const DAYS_IN_MONTH_MIN = 28; // Minimum days for monthly subscription
const DAYS_IN_MONTH_MAX = 32; // Maximum days for monthly subscription

export function SettingsPage() {
    const { user, refreshUser } = useAuth();
    const queryClient = useQueryClient();
    const { consent, updateConsent, doNotTrack, resetConsent } = useConsent();

    // Refs to store timeout IDs for success messages (cleaned up on unmount)
    const successTimeoutRef = useRef<ReturnType<typeof setTimeout> | null>(null);
    const profileTimeoutRef = useRef<ReturnType<typeof setTimeout> | null>(null);
    const settingsTimeoutRef = useRef<ReturnType<typeof setTimeout> | null>(null);
    const consentTimeoutRef = useRef<ReturnType<typeof setTimeout> | null>(null);

    // Cleanup all timeouts on unmount
    useEffect(() => {
        return () => {
            if (successTimeoutRef.current) clearTimeout(successTimeoutRef.current);
            if (profileTimeoutRef.current) clearTimeout(profileTimeoutRef.current);
            if (settingsTimeoutRef.current) clearTimeout(settingsTimeoutRef.current);
            if (consentTimeoutRef.current) clearTimeout(consentTimeoutRef.current);
        };
    }, []);

    // Profile state
    const [profileData, setProfileData] = useState<UpdateProfileRequest>({
        display_name: '',
        bio: null,
    });
    const [isSavingProfile, setIsSavingProfile] = useState(false);
    const [profileSuccess, setProfileSuccess] = useState(false);
    const [profileError, setProfileError] = useState<string | null>(null);

    // Settings state
    const [settingsData, setSettingsData] = useState<UpdateSettingsRequest>({});
    const [isSavingSettings, setIsSavingSettings] = useState(false);
    const [settingsSuccess, setSettingsSuccess] = useState(false);
    const [settingsError, setSettingsError] = useState<string | null>(null);

    // Consent state
    const [consentSuccess, setConsentSuccess] = useState(false);

    // Subscription state
    const [showCancelModal, setShowCancelModal] = useState(false);
    const [cancelImmediate, setCancelImmediate] = useState(false);
    const [subscriptionError, setSubscriptionError] = useState<string | null>(null);
    const [subscriptionSuccess, setSubscriptionSuccess] = useState<string | null>(null);

    // Load subscription
    const { data: subscription, refetch: refetchSubscription } = useQuery<Subscription | null>({
        queryKey: ['subscription'],
        queryFn: getSubscription,
    });

    // Load user settings
    const { data: settings, isLoading: settingsLoading } = useQuery({
        queryKey: ['userSettings'],
        queryFn: getUserSettings,
    });

    // Initialize form data when user or settings load
    useEffect(() => {
        if (user) {
            setProfileData({
                display_name: user.display_name,
                bio: user.bio || null,
            });
        }
    }, [user]);

    useEffect(() => {
        if (settings) {
            setSettingsData({
                profile_visibility: settings.profile_visibility,
                show_karma_publicly: settings.show_karma_publicly,
            });
        }
    }, [settings]);

    // Profile update
    const handleProfileSubmit = async (e: React.FormEvent) => {
        e.preventDefault();
        setIsSavingProfile(true);
        setProfileError(null);
        setProfileSuccess(false);

        try {
            await updateProfile(profileData);
            await refreshUser();
            setProfileSuccess(true);
            profileTimeoutRef.current = setTimeout(() => setProfileSuccess(false), 3000);
        } catch {
            setProfileError('Failed to update profile. Please try again.');
        } finally {
            setIsSavingProfile(false);
        }
    };

    // Settings update
    const handleSettingsSubmit = async (e: React.FormEvent) => {
        e.preventDefault();
        setIsSavingSettings(true);
        setSettingsError(null);
        setSettingsSuccess(false);

        try {
            await updateUserSettings(settingsData);
            queryClient.invalidateQueries({ queryKey: ['userSettings'] });
            setSettingsSuccess(true);
            settingsTimeoutRef.current = setTimeout(() => setSettingsSuccess(false), 3000);
        } catch {
            setSettingsError('Failed to update settings. Please try again.');
        } finally {
            setIsSavingSettings(false);
        }
    };

    // Subscription management handlers
    const handleCancelSubscription = async () => {
        setSubscriptionError(null);
        setSubscriptionSuccess(null);
        try {
            await cancelSubscription(cancelImmediate);
            await refetchSubscription();
            setShowCancelModal(false);
            setCancelImmediate(false); // Reset to default

            // Clear any existing timeout before setting a new one
            if (successTimeoutRef.current) {
                clearTimeout(successTimeoutRef.current);
            }

            setSubscriptionSuccess(
                cancelImmediate
                    ? 'Subscription canceled immediately'
                    : 'Subscription will be canceled at the end of the billing period',
            );
            successTimeoutRef.current = setTimeout(() => setSubscriptionSuccess(null), 5000);
        } catch (error: unknown) {
            const err = error as { response?: { data?: { error?: string } } };
            setSubscriptionError(err.response?.data?.error || 'Failed to cancel subscription');
        }
    };

    const handleReactivateSubscription = async () => {
        setSubscriptionError(null);
        setSubscriptionSuccess(null);

        // Clear any existing timeout before setting a new one
        if (successTimeoutRef.current) {
            clearTimeout(successTimeoutRef.current);
        }

        try {
            await reactivateSubscription();
            await refetchSubscription();
            setSubscriptionSuccess('Subscription reactivated successfully');
            successTimeoutRef.current = setTimeout(() => setSubscriptionSuccess(null), 5000);
        } catch (error: unknown) {
            const err = error as { response?: { data?: { error?: string } } };
            setSubscriptionError(err.response?.data?.error || 'Failed to reactivate subscription');
        }
    };

    const handleManageSubscription = async () => {
        setSubscriptionError(null);
        try {
            const { portal_url } = await createPortalSession();
            window.open(portal_url, '_blank');
        } catch (error: unknown) {
            const err = error as { response?: { data?: { error?: string } } };
            setSubscriptionError(err.response?.data?.error || 'Failed to open customer portal');
        }
    };

    if (!user) {
        return null;
    }

    const showTwitchName = user.username.toLowerCase() !== user.display_name.toLowerCase();

    return (
        <>
            <Helmet>
                <title>Settings - clpr</title>
            </Helmet>

            <Container className='py-4 xs:py-6 md:py-8'>
                <div className='max-w-3xl mx-auto'>
                    <h1 className='text-2xl xs:text-3xl font-bold mb-4 xs:mb-6'>Settings</h1>

                    {/* Profile Settings */}
                    <Card className='mb-4 xs:mb-6'>
                        <CardHeader>
                            <h2 className='text-lg xs:text-xl font-semibold'>Profile</h2>
                        </CardHeader>
                        <CardBody>
                            <form onSubmit={handleProfileSubmit}>
                                <Stack direction='vertical' gap={4}>
                                    {showTwitchName && (
                                        <Input
                                            label='Twitch Username'
                                            value={user.username}
                                            disabled
                                            helperText='This is your Twitch username and cannot be changed'
                                        />
                                    )}
                                    <Input
                                        label='Display Name'
                                        value={profileData.display_name}
                                        onChange={(e) =>
                                            setProfileData({
                                                ...profileData,
                                                display_name: e.target.value,
                                            })
                                        }
                                        required
                                        maxLength={100}
                                        helperText='This is how your name appears on the site'
                                    />
                                    <TextArea
                                        label='Bio'
                                        value={profileData.bio || ''}
                                        onChange={(e) =>
                                            setProfileData({
                                                ...profileData,
                                                bio: e.target.value || null,
                                            })
                                        }
                                        rows={4}
                                        maxLength={500}
                                        placeholder='Tell us about yourself...'
                                        helperText={`${(profileData.bio || '').length}/500 characters`}
                                    />
                                    {user.email && (
                                        <Input
                                            label='Email'
                                            value={user.email}
                                            disabled
                                            helperText='Email is managed through your Twitch account'
                                        />
                                    )}
                                    <div className='flex gap-3'>
                                        <Button type='submit' variant='primary' disabled={isSavingProfile}>
                                            {isSavingProfile ? 'Saving...' : 'Save Profile'}
                                        </Button>
                                    </div>
                                    {profileSuccess && <Alert variant='success'>Profile updated successfully!</Alert>}
                                    {profileError && <Alert variant='error'>{profileError}</Alert>}
                                </Stack>
                            </form>
                        </CardBody>
                    </Card>

                    {/* Privacy Settings */}
                    <Card className='mb-6'>
                        <CardHeader>
                            <h2 className='text-xl font-semibold'>Privacy Settings</h2>
                        </CardHeader>
                        <CardBody>
                            {settingsLoading ? (
                                <div className='text-center py-4'>Loading settings...</div>
                            ) : (
                                <form onSubmit={handleSettingsSubmit}>
                                    <Stack direction='vertical' gap={4}>
                                        <div>
                                            <label className='block text-sm font-medium mb-2'>Profile Visibility</label>
                                            <select
                                                value={settingsData.profile_visibility || 'public'}
                                                onChange={(e) =>
                                                    setSettingsData({
                                                        ...settingsData,
                                                        profile_visibility: e.target.value as
                                                            'public' | 'private' | 'followers',
                                                    })
                                                }
                                                className='w-full px-3 py-2 border border-border rounded-md bg-background text-foreground'
                                            >
                                                <option value='public'>Public - Anyone can view your profile</option>
                                                <option value='private'>
                                                    Private - Only you can view your profile
                                                </option>
                                                <option value='followers'>Followers - Only followers can view</option>
                                            </select>
                                        </div>
                                        <Toggle
                                            label='Show Uppies Publicly'
                                            helperText='Display your uppies on your public profile'
                                            checked={settingsData.show_karma_publicly ?? true}
                                            onChange={(e) =>
                                                setSettingsData({
                                                    ...settingsData,
                                                    show_karma_publicly: e.target.checked,
                                                })
                                            }
                                        />
                                        <div className='flex gap-3'>
                                            <Button type='submit' variant='primary' disabled={isSavingSettings}>
                                                {isSavingSettings ? 'Saving...' : 'Save Settings'}
                                            </Button>
                                        </div>
                                        {settingsSuccess && (
                                            <Alert variant='success'>Settings updated successfully!</Alert>
                                        )}
                                        {settingsError && <Alert variant='error'>{settingsError}</Alert>}
                                    </Stack>
                                </form>
                            )}
                        </CardBody>
                    </Card>

                    {/* Advertising & Privacy Consent */}
                    <Card className='mb-6'>
                        <CardHeader>
                            <h2 className='text-xl font-semibold'>Advertising & Privacy</h2>
                        </CardHeader>
                        <CardBody>
                            {doNotTrack && (
                                <Alert variant='info' className='mb-4'>
                                    <strong>Do Not Track enabled:</strong> Your browser has Do Not Track enabled.
                                    Personalized ads and analytics will be automatically disabled regardless of your
                                    consent settings.
                                </Alert>
                            )}
                            <Stack direction='vertical' gap={4}>
                                <Toggle
                                    label='Functional Cookies'
                                    helperText='Remember your preferences like language, theme, and other settings'
                                    checked={consent.functional}
                                    onChange={(e) => {
                                        updateConsent({
                                            functional: e.target.checked,
                                        });
                                        setConsentSuccess(true);
                                        if (consentTimeoutRef.current) clearTimeout(consentTimeoutRef.current);
                                        consentTimeoutRef.current = setTimeout(() => setConsentSuccess(false), 3000);
                                    }}
                                    disabled={doNotTrack}
                                />
                                <Toggle
                                    label='Analytics Tracking'
                                    helperText='Help us improve clpr by allowing anonymous usage analytics'
                                    checked={consent.analytics}
                                    onChange={(e) => {
                                        updateConsent({
                                            analytics: e.target.checked,
                                        });
                                        setConsentSuccess(true);
                                        if (consentTimeoutRef.current) clearTimeout(consentTimeoutRef.current);
                                        consentTimeoutRef.current = setTimeout(() => setConsentSuccess(false), 3000);
                                    }}
                                    disabled={doNotTrack}
                                />
                                <Toggle
                                    label='Personalized Advertising'
                                    helperText='Allow ads tailored to your interests. Without this, you will see contextual ads based on page content.'
                                    checked={consent.advertising}
                                    onChange={(e) => {
                                        updateConsent({
                                            advertising: e.target.checked,
                                        });
                                        setConsentSuccess(true);
                                        if (consentTimeoutRef.current) clearTimeout(consentTimeoutRef.current);
                                        consentTimeoutRef.current = setTimeout(() => setConsentSuccess(false), 3000);
                                    }}
                                    disabled={doNotTrack}
                                />
                                <div className='pt-2 border-t border-border'>
                                    <div className='flex flex-wrap gap-2'>
                                        <Button asChild variant='outline' size='sm'>
                                            <Link to='/settings/cookies'>
                                                Manage Cookie Settings
                                            </Link>
                                        </Button>
                                        <Button
                                            variant='ghost'
                                            size='sm'
                                            onClick={() => {
                                                resetConsent();
                                            }}
                                        >
                                            Reset Consent (Show Banner Again)
                                        </Button>
                                    </div>
                                </div>
                                {consentSuccess && <Alert variant='success'>Privacy preferences updated!</Alert>}
                                <p className='text-xs text-muted-foreground'>
                                    Learn more about how we use your data in our{' '}
                                    <Link to='/privacy' className='text-primary-500 hover:underline'>
                                        Privacy Policy
                                    </Link>
                                </p>
                            </Stack>
                        </CardBody>
                    </Card>

                    {/* Subscription Management */}
                    <Card className='mb-6' data-testid='subscription-section'>
                        <CardHeader>
                            <h2 className='text-xl font-semibold'>Subscription</h2>
                        </CardHeader>
                        <CardBody>
                            {subscriptionError && (
                                <Alert variant='error' className='mb-4'>
                                    {subscriptionError}
                                </Alert>
                            )}
                            {subscriptionSuccess && (
                                <Alert variant='success' className='mb-4'>
                                    {subscriptionSuccess}
                                </Alert>
                            )}
                            {subscription ? (
                                <Stack direction='vertical' gap={4}>
                                    <div>
                                        <label className='block text-sm font-medium mb-1'>Current Plan</label>
                                        <p className='text-lg font-semibold' data-testid='current-plan'>
                                            {subscription.tier === 'pro' ? 'Pro' : 'Free'}
                                        </p>
                                    </div>

                                    {subscription.tier === 'pro' && subscription.status && (
                                        <div>
                                            <label className='block text-sm font-medium mb-1'>Status</label>
                                            <p data-testid='subscription-status'>
                                                {subscription.cancel_at_period_end ? (
                                                    <span className='text-warning-600'>Will cancel at period end</span>
                                                ) : (
                                                    <span className='text-success-600 capitalize'>
                                                        {subscription.status}
                                                    </span>
                                                )}
                                            </p>
                                        </div>
                                    )}

                                    {subscription.tier === 'pro' &&
                                        subscription.current_period_start &&
                                        subscription.current_period_end && (
                                            <div>
                                                <label className='block text-sm font-medium mb-1'>Billing Period</label>
                                                <p data-testid='billing-period'>
                                                    {(() => {
                                                        const start = new Date(subscription.current_period_start);
                                                        const end = new Date(subscription.current_period_end);
                                                        const daysDiff = Math.round(
                                                            (end.getTime() - start.getTime()) / (1000 * 60 * 60 * 24),
                                                        );

                                                        if (daysDiff >= DAYS_IN_YEAR) {
                                                            return 'Yearly';
                                                        } else if (
                                                            daysDiff >= DAYS_IN_MONTH_MIN &&
                                                            daysDiff <= DAYS_IN_MONTH_MAX
                                                        ) {
                                                            return 'Monthly';
                                                        } else {
                                                            return `${daysDiff} days`;
                                                        }
                                                    })()}
                                                </p>
                                            </div>
                                        )}

                                    {subscription.current_period_end && (
                                        <div>
                                            <label className='block text-sm font-medium mb-1'>Next Billing Date</label>
                                            <p data-testid='next-billing-date'>
                                                {new Date(subscription.current_period_end).toLocaleDateString()}
                                            </p>
                                        </div>
                                    )}

                                    <div className='flex flex-wrap gap-3 pt-2'>
                                        {subscription.tier === 'pro' && !subscription.cancel_at_period_end && (
                                            <>
                                                <Button variant='outline' onClick={handleManageSubscription}>
                                                    Manage Subscription
                                                </Button>
                                                <Button
                                                    variant='outline'
                                                    className='text-error-600 border-error-600 hover:bg-error-50'
                                                    onClick={() => setShowCancelModal(true)}
                                                >
                                                    Cancel Subscription
                                                </Button>
                                            </>
                                        )}
                                        {subscription.cancel_at_period_end && (
                                            <Button variant='primary' onClick={handleReactivateSubscription}>
                                                Reactivate Subscription
                                            </Button>
                                        )}
                                        {subscription.tier === 'free' && (
                                            <Button asChild variant='primary'><Link to='/pricing'>Upgrade to Pro</Link></Button>
                                        )}
                                    </div>
                                </Stack>
                            ) : (
                                <div>
                                    <p className='text-muted-foreground mb-4'>You are currently on the free plan.</p>
                                    <Button asChild variant='primary'><Link to='/pricing'>View Pro Plans</Link></Button>
                                </div>
                            )}
                        </CardBody>
                    </Card>

                    {/* Notification Settings */}
                    <Card className='mb-6'>
                        <CardHeader>
                            <h2 className='text-xl font-semibold'>Notifications</h2>
                        </CardHeader>
                        <CardBody>
                            <p className='text-muted-foreground mb-4'>
                                Manage your notification preferences including email and reply notifications.
                            </p>
                            <Button asChild variant='outline'><Link to='/notifications/preferences'>Manage Notification Preferences</Link></Button>
                        </CardBody>
                    </Card>

                    {/* Webhook Settings */}
                    <Card className='mb-6'>
                        <CardHeader>
                            <h2 className='text-xl font-semibold'>Webhooks</h2>
                        </CardHeader>
                        <CardBody>
                            <p className='text-muted-foreground mb-4'>
                                Configure webhook endpoints to receive real-time notifications when events occur.
                                Integrate with your own applications.
                            </p>
                            <Button asChild variant='outline'><Link to='/settings/webhooks'>Manage Webhook Subscriptions</Link></Button>
                        </CardBody>
                    </Card>

                </div>
            </Container>

            {/* Cancel Subscription Modal */}
            <Modal
                open={showCancelModal}
                onClose={() => {
                    setShowCancelModal(false);
                    setCancelImmediate(false); // Reset to default when closing
                }}
                title='Cancel Subscription'
            >
                <div className='space-y-4'>
                    <Alert variant='warning'>
                        <strong>Warning:</strong> Are you sure you want to cancel your subscription?
                    </Alert>
                    <p className='text-sm text-muted-foreground'>
                        Choose when you would like your subscription to end:
                    </p>
                    <div className='space-y-3'>
                        <label className='flex items-start gap-3 cursor-pointer'>
                            <input
                                type='radio'
                                name='cancelType'
                                checked={!cancelImmediate}
                                onChange={() => setCancelImmediate(false)}
                                className='mt-1'
                            />
                            <div>
                                <div className='font-medium'>End at period end</div>
                                <div className='text-sm text-muted-foreground'>
                                    You'll retain access until{' '}
                                    {subscription?.current_period_end
                                        ? new Date(subscription.current_period_end).toLocaleDateString()
                                        : 'the end of your billing period'}
                                </div>
                            </div>
                        </label>
                        <label className='flex items-start gap-3 cursor-pointer'>
                            <input
                                type='radio'
                                name='cancelType'
                                checked={cancelImmediate}
                                onChange={() => setCancelImmediate(true)}
                                className='mt-1'
                            />
                            <div>
                                <div className='font-medium'>Cancel immediately</div>
                                <div className='text-sm text-muted-foreground'>
                                    Access will end immediately (no refund for remaining time)
                                </div>
                            </div>
                        </label>
                    </div>
                    <div className='flex gap-3 justify-end pt-4'>
                        <Button variant='ghost' onClick={() => setShowCancelModal(false)}>
                            Keep Subscription
                        </Button>
                        <Button
                            variant='primary'
                            onClick={handleCancelSubscription}
                            className='bg-error-600 hover:bg-error-700'
                        >
                            Cancel Subscription
                        </Button>
                    </div>
                </div>
            </Modal>
        </>
    );
}
