import { useQuery, useQueryClient } from '@tanstack/react-query';
import { useEffect, useState, useRef } from 'react';
import { Helmet } from '@dr.pogodin/react-helmet';
import { Link } from 'react-router-dom';
import { Alert, Button, Card, CardBody, CardHeader, Container, Input, Stack, TextArea, Toggle } from '../components';
import { useAuth } from '../context/AuthContext';
import { useConsent } from '../context/ConsentContext';
import { effectiveConsentValue } from '../lib/consent-display';
import type { UpdateProfileRequest, UpdateSettingsRequest } from '../lib/user-settings-api';
import { getUserSettings, updateProfile, updateUserSettings } from '../lib/user-settings-api';
import { useFeedAutoplayPreference } from '../hooks';

export function SettingsPage() {
    const { user, refreshUser } = useAuth();
    const queryClient = useQueryClient();
    const { consent, updateConsent, doNotTrack, resetConsent } = useConsent();
    const { preference: feedAutoplay, setPreference: setFeedAutoplay } = useFeedAutoplayPreference();

    // Refs to store timeout IDs for success messages (cleaned up on unmount)
    const profileTimeoutRef = useRef<ReturnType<typeof setTimeout> | null>(null);
    const settingsTimeoutRef = useRef<ReturnType<typeof setTimeout> | null>(null);
    const consentTimeoutRef = useRef<ReturnType<typeof setTimeout> | null>(null);

    // Cleanup all timeouts on unmount
    useEffect(() => {
        return () => {
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

                    <Card className='mb-6'>
                        <CardHeader>
                            <h2 className='text-xl font-semibold'>Playback</h2>
                        </CardHeader>
                        <CardBody>
                            <Toggle
                                label='Muted feed autoplay'
                                helperText='Automatically play the most visible clip without sound. Only one Twitch player is loaded at a time.'
                                checked={feedAutoplay === 'muted'}
                                onChange={event => setFeedAutoplay(event.target.checked ? 'muted' : 'manual')}
                            />
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
                                    Optional consent is effectively off. Your saved choices are retained but are not
                                    used while this signal remains enabled.
                                </Alert>
                            )}
                            <Stack direction='vertical' gap={4}>
                                <Toggle
                                    label='Functional Cookies'
                                    helperText='Remember your preferences like language, theme, and other settings'
                                    checked={effectiveConsentValue(consent.functional, doNotTrack)}
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
                                    checked={effectiveConsentValue(consent.analytics, doNotTrack)}
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
                                    checked={effectiveConsentValue(consent.advertising, doNotTrack)}
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

                    {/* Account Access */}
                    <Card className='mb-6' data-testid='account-access-section'>
                        <CardHeader>
                            <h2 className='text-xl font-semibold'>Account Access</h2>
                        </CardHeader>
                        <CardBody>
                            <div className='rounded-xl border border-brand/25 bg-brand/5 p-5'>
                                <p className='text-lg font-semibold text-foreground'>Everything is included.</p>
                                <p className='mt-2 text-sm leading-6 text-muted-foreground'>
                                    clpr does not have paid account tiers or feature quotas. Normal safeguards against spam, abuse, and excessive automated traffic still apply to everyone.
                                </p>
                                <div className='mt-4 flex flex-wrap gap-3'>
                                    <Button asChild variant='outline'>
                                        <Link to='/support'>How community support works</Link>
                                    </Button>
                                    <Button asChild variant='primary'>
                                        <a href='https://patreon.com/subcult' target='_blank' rel='noopener noreferrer'>Support on Patreon</a>
                                    </Button>
                                </div>
                            </div>
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

        </>
    );
}
