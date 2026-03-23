import { useEffect, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import {
    Alert,
    Button,
    Card,
    Container,
    Input,
    Spinner,
    TextArea,
} from '../components';
import { useAuth } from '../context/AuthContext';
import {
    createVerificationApplication,
    getMyVerificationApplication,
    type VerificationApplication,
} from '../lib/verification-api';

// Helper function to extract error message from API error
function getErrorMessage(error: unknown): string {
    if (typeof error === 'object' && error !== null && 'response' in error) {
        const apiError = error as { response?: { data?: { error?: string } } };
        return apiError.response?.data?.error || 'An error occurred';
    }
    return 'An error occurred';
}

export function VerificationApplicationPage() {
    const { isAuthenticated } = useAuth();
    const navigate = useNavigate();
    const [existingApplication, setExistingApplication] = useState<VerificationApplication | null>(null);
    const [isLoading, setIsLoading] = useState(true);
    const [isSubmitting, setIsSubmitting] = useState(false);
    const [error, setError] = useState<string | null>(null);
    const [success, setSuccess] = useState<string | null>(null);

    // Form fields
    const [twitchChannelUrl, setTwitchChannelUrl] = useState('');
    const [followerCount, setFollowerCount] = useState('');
    const [subscriberCount, setSubscriberCount] = useState('');
    const [avgViewers, setAvgViewers] = useState('');
    const [contentDescription, setContentDescription] = useState('');
    const [socialMediaLinks, setSocialMediaLinks] = useState({
        twitter: '',
        youtube: '',
        instagram: '',
        discord: '',
    });

    useEffect(() => {
        if (!isAuthenticated) {
            navigate('/');
            return;
        }

        // Check if user already has an application
        const loadExistingApplication = async () => {
            try {
                setIsLoading(true);
                const response = await getMyVerificationApplication();
                setExistingApplication(response.data);
            } catch (err: unknown) {
                // 404 is expected if no application exists
                const errorMsg = getErrorMessage(err);
                if (!errorMsg.includes('not found')) {
                    console.error('Failed to load existing application:', err);
                }
            } finally {
                setIsLoading(false);
            }
        };

        loadExistingApplication();
    }, [isAuthenticated, navigate]);

    const handleSubmit = async (e: React.FormEvent) => {
        e.preventDefault();
        setError(null);
        setSuccess(null);

        // Validate required fields
        if (!twitchChannelUrl.trim()) {
            setError('Twitch channel URL is required');
            return;
        }

        try {
            setIsSubmitting(true);

            // Prepare social media links (only include non-empty values)
            const links: Record<string, string> = {};
            if (socialMediaLinks.twitter.trim()) links.twitter = socialMediaLinks.twitter.trim();
            if (socialMediaLinks.youtube.trim()) links.youtube = socialMediaLinks.youtube.trim();
            if (socialMediaLinks.instagram.trim()) links.instagram = socialMediaLinks.instagram.trim();
            if (socialMediaLinks.discord.trim()) links.discord = socialMediaLinks.discord.trim();

            // Parse and validate numeric inputs
            const parsedFollowerCount = followerCount ? parseInt(followerCount, 10) : undefined;
            const parsedSubscriberCount = subscriberCount ? parseInt(subscriberCount, 10) : undefined;
            const parsedAvgViewers = avgViewers ? parseInt(avgViewers, 10) : undefined;

            // Check for invalid numbers
            if ((parsedFollowerCount !== undefined && isNaN(parsedFollowerCount)) ||
                (parsedSubscriberCount !== undefined && isNaN(parsedSubscriberCount)) ||
                (parsedAvgViewers !== undefined && isNaN(parsedAvgViewers))) {
                setError('Please enter valid numbers for follower count, subscriber count, and average viewers');
                return;
            }

            await createVerificationApplication({
                twitch_channel_url: twitchChannelUrl.trim(),
                follower_count: parsedFollowerCount,
                subscriber_count: parsedSubscriberCount,
                avg_viewers: parsedAvgViewers,
                content_description: contentDescription.trim() || undefined,
                social_media_links: Object.keys(links).length > 0 ? links : undefined,
            });

            setSuccess('Your verification application has been submitted successfully! Our team will review it soon.');

            // Reload the application status
            try {
                const response = await getMyVerificationApplication();
                setExistingApplication(response.data);
            } catch (err) {
                console.error('Failed to reload application:', err);
            }
        } catch (err: unknown) {
            setError(getErrorMessage(err));
        } finally {
            setIsSubmitting(false);
        }
    };

    const formatDate = (dateString: string) => {
        return new Date(dateString).toLocaleDateString('en-US', {
            year: 'numeric',
            month: 'long',
            day: 'numeric',
        });
    };

    const getStatusColor = (status: string) => {
        switch (status) {
            case 'approved':
                return 'text-green-600 dark:text-green-400';
            case 'rejected':
                return 'text-red-600 dark:text-red-400';
            case 'pending':
            default:
                return 'text-yellow-600 dark:text-yellow-400';
        }
    };

    if (isLoading) {
        return (
            <Container>
                <div className="flex justify-center py-12">
                    <Spinner size="lg" />
                </div>
            </Container>
        );
    }

    // Show existing application status if one exists
    if (existingApplication) {
        return (
            <Container>
                <div className="py-8 max-w-3xl mx-auto">
                    <h1 className="text-3xl font-bold text-foreground mb-2">
                        Creator Verification Application
                    </h1>
                    <p className="text-muted-foreground mb-6">
                        View the status of your verification application
                    </p>

                    <Card className="p-6">
                        <div className="space-y-4">
                            <div>
                                <h3 className="text-lg font-semibold text-foreground mb-2">
                                    Application Status
                                </h3>
                                <p className={`text-2xl font-bold ${getStatusColor(existingApplication.status)}`}>
                                    {existingApplication.status.charAt(0).toUpperCase() + existingApplication.status.slice(1)}
                                </p>
                            </div>

                            <div>
                                <span className="text-sm font-medium text-foreground">
                                    Submitted on:
                                </span>{' '}
                                <span className="text-sm text-muted-foreground">
                                    {formatDate(existingApplication.created_at)}
                                </span>
                            </div>

                            <div>
                                <span className="text-sm font-medium text-foreground">
                                    Twitch Channel:
                                </span>{' '}
                                <a
                                    href={existingApplication.twitch_channel_url}
                                    target="_blank"
                                    rel="noopener noreferrer"
                                    className="text-sm text-primary-600 hover:underline"
                                >
                                    {existingApplication.twitch_channel_url}
                                </a>
                            </div>

                            {existingApplication.reviewed_at && (
                                <div>
                                    <span className="text-sm font-medium text-foreground">
                                        Reviewed on:
                                    </span>{' '}
                                    <span className="text-sm text-muted-foreground">
                                        {formatDate(existingApplication.reviewed_at)}
                                    </span>
                                </div>
                            )}

                            {existingApplication.reviewer_notes && (
                                <div>
                                    <h4 className="text-sm font-medium text-foreground mb-1">
                                        Reviewer Notes:
                                    </h4>
                                    <p className="text-sm text-muted-foreground bg-surface p-3 rounded">
                                        {existingApplication.reviewer_notes}
                                    </p>
                                </div>
                            )}

                            {existingApplication.status === 'pending' && (
                                <Alert variant="info" className="mt-4">
                                    Your application is currently under review. We'll notify you once a decision has been made.
                                </Alert>
                            )}

                            {existingApplication.status === 'approved' && (
                                <Alert variant="success" className="mt-4">
                                    Congratulations! Your verification application has been approved. You now have a verified badge on your profile.
                                </Alert>
                            )}

                            {existingApplication.status === 'rejected' && (
                                <Alert variant="error" className="mt-4">
                                    Your verification application was not approved. You can reapply after 30 days.
                                </Alert>
                            )}
                        </div>
                    </Card>

                    <div className="mt-6">
                        <Button variant="secondary" onClick={() => navigate('/profile')}>
                            Back to Profile
                        </Button>
                    </div>
                </div>
            </Container>
        );
    }

    // Show application form if no existing application
    return (
        <Container>
            <div className="py-8 max-w-3xl mx-auto">
                <h1 className="text-3xl font-bold text-foreground mb-2">
                    Apply for Creator Verification
                </h1>
                <p className="text-muted-foreground mb-6">
                    Get a verified badge to show that you're an authentic creator
                </p>

                <Card className="p-6 mb-6">
                    <h3 className="text-lg font-semibold text-foreground mb-3">
                        What is Creator Verification?
                    </h3>
                    <p className="text-sm text-muted-foreground mb-3">
                        Creator verification is a trust signal that helps viewers identify authentic creators on Clipper.
                        Verified creators receive a blue checkmark badge displayed next to their name.
                    </p>
                    <h4 className="text-md font-semibold text-foreground mb-2">
                        Eligibility Requirements:
                    </h4>
                    <ul className="list-disc list-inside text-sm text-muted-foreground space-y-1">
                        <li>Active Twitch streaming presence</li>
                        <li>Established follower base</li>
                        <li>Consistent content creation</li>
                        <li>Good standing in the community</li>
                        <li>No history of policy violations</li>
                    </ul>
                </Card>

                {error && (
                    <Alert variant="error" className="mb-4" onDismiss={() => setError(null)}>
                        {error}
                    </Alert>
                )}

                {success && (
                    <Alert variant="success" className="mb-4" onDismiss={() => setSuccess(null)}>
                        {success}
                    </Alert>
                )}

                <form onSubmit={handleSubmit}>
                    <Card className="p-6 space-y-4">
                        <div>
                            <label htmlFor="twitch-channel-url" className="block text-sm font-medium text-foreground mb-2">
                                Twitch Channel URL <span className="text-red-500">*</span>
                            </label>
                            <Input
                                id="twitch-channel-url"
                                type="url"
                                value={twitchChannelUrl}
                                onChange={(e) => setTwitchChannelUrl(e.target.value)}
                                placeholder="https://twitch.tv/your_channel"
                                required
                            />
                            <p className="text-xs text-muted-foreground mt-1">
                                Your primary Twitch channel URL
                            </p>
                        </div>

                        <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
                            <div>
                                <label htmlFor="follower-count" className="block text-sm font-medium text-foreground mb-2">
                                    Follower Count
                                </label>
                                <Input
                                    id="follower-count"
                                    type="number"
                                    value={followerCount}
                                    onChange={(e) => setFollowerCount(e.target.value)}
                                    placeholder="0"
                                    min="0"
                                />
                            </div>

                            <div>
                                <label htmlFor="subscriber-count" className="block text-sm font-medium text-foreground mb-2">
                                    Subscriber Count
                                </label>
                                <Input
                                    id="subscriber-count"
                                    type="number"
                                    value={subscriberCount}
                                    onChange={(e) => setSubscriberCount(e.target.value)}
                                    placeholder="0"
                                    min="0"
                                />
                            </div>

                            <div>
                                <label htmlFor="avg-viewers" className="block text-sm font-medium text-foreground mb-2">
                                    Average Viewers
                                </label>
                                <Input
                                    id="avg-viewers"
                                    type="number"
                                    value={avgViewers}
                                    onChange={(e) => setAvgViewers(e.target.value)}
                                    placeholder="0"
                                    min="0"
                                />
                            </div>
                        </div>

                        <div>
                            <label htmlFor="content-description" className="block text-sm font-medium text-foreground mb-2">
                                Content Description
                            </label>
                            <TextArea
                                id="content-description"
                                value={contentDescription}
                                onChange={(e) => setContentDescription(e.target.value)}
                                placeholder="Tell us about your content (games you play, streaming schedule, community focus, etc.)"
                                rows={4}
                            />
                        </div>

                        <div>
                            <h4 className="text-sm font-medium text-foreground mb-3">
                                Social Media Links (Optional)
                            </h4>
                            <div className="space-y-3">
                                <div>
                                    <label htmlFor="twitter-link" className="block text-xs text-muted-foreground mb-1">
                                        Twitter/X
                                    </label>
                                    <Input
                                        id="twitter-link"
                                        type="url"
                                        value={socialMediaLinks.twitter}
                                        onChange={(e) =>
                                            setSocialMediaLinks({ ...socialMediaLinks, twitter: e.target.value })
                                        }
                                        placeholder="https://twitter.com/your_handle"
                                    />
                                </div>

                                <div>
                                    <label htmlFor="youtube-link" className="block text-xs text-muted-foreground mb-1">
                                        YouTube
                                    </label>
                                    <Input
                                        id="youtube-link"
                                        type="url"
                                        value={socialMediaLinks.youtube}
                                        onChange={(e) =>
                                            setSocialMediaLinks({ ...socialMediaLinks, youtube: e.target.value })
                                        }
                                        placeholder="https://youtube.com/@your_channel"
                                    />
                                </div>

                                <div>
                                    <label htmlFor="instagram-link" className="block text-xs text-muted-foreground mb-1">
                                        Instagram
                                    </label>
                                    <Input
                                        id="instagram-link"
                                        type="url"
                                        value={socialMediaLinks.instagram}
                                        onChange={(e) =>
                                            setSocialMediaLinks({ ...socialMediaLinks, instagram: e.target.value })
                                        }
                                        placeholder="https://instagram.com/your_handle"
                                    />
                                </div>

                                <div>
                                    <label htmlFor="discord-link" className="block text-xs text-muted-foreground mb-1">
                                        Discord Server
                                    </label>
                                    <Input
                                        id="discord-link"
                                        type="url"
                                        value={socialMediaLinks.discord}
                                        onChange={(e) =>
                                            setSocialMediaLinks({ ...socialMediaLinks, discord: e.target.value })
                                        }
                                        placeholder="https://discord.gg/your_server"
                                    />
                                </div>
                            </div>
                        </div>

                        <Alert variant="info">
                            By submitting this application, you confirm that all information provided is accurate and
                            that you are the legitimate owner of the Twitch channel specified.
                        </Alert>

                        <div className="flex gap-3 pt-4">
                            <Button type="submit" variant="primary" disabled={isSubmitting}>
                                {isSubmitting ? 'Submitting...' : 'Submit Application'}
                            </Button>
                            <Button type="button" variant="secondary" onClick={() => navigate('/profile')}>
                                Cancel
                            </Button>
                        </div>
                    </Card>
                </form>
            </div>
        </Container>
    );
}
