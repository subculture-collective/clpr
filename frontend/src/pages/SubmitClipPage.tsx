import { useEffect, useMemo, useState } from 'react';
import { useNavigate, useLocation } from 'react-router-dom';
import { useQueryClient } from '@tanstack/react-query';
import {
    Alert,
    Button,
    Card,
    Checkbox,
    Container,
    Input,
    SubmissionConfirmation,
    TextArea,
} from '../components';
import { RateLimitError } from '../components/clip/RateLimitError';
import { DuplicateClipError } from '../components/clip/DuplicateClipError';
import { useAuth } from '../context/AuthContext';
import {
    checkClipStatus,
    getUserSubmissions,
    submitClip,
    submitClipUpload,
    getClipMetadata,
} from '../lib/submission-api';
import { trackEvent, SubmissionEvents } from '../lib/telemetry';
import type {
    ClipSubmission,
    ClipSourcePlatform,
    ClipSourceType,
    SubmitClipRequest,
    RateLimitErrorResponse,
} from '../types/submission';
import { TagSelector } from '../components/tag/TagSelector';
import { tagApi } from '../lib/tag-api';
import type { Tag } from '../types/tag';
import { useSubmissionDraft } from '../hooks/useSubmissionDraft';

/**
 * Clip-specific duplicate error patterns to avoid false positives
 * from unrelated errors like "Email already taken" or "Username already exists"
 */
const CLIP_DUPLICATE_PATTERNS = [
    /clip.*already/,
    /already.*posted/,
    /already.*submitted/,
    /already.*added.*database/,
    /already.*approved/,
    /already.*pending/,
    /duplicate.*clip/,
    /cannot be submitted again/,
];

/**
 * Helper to check if an error message indicates a duplicate clip
 * Note: Currently uses string matching. For better reliability,
 * consider updating backend to return error.code or error.type field
 * (e.g., { error: "...", code: "DUPLICATE_CLIP" })
 *
 * Uses specific clip-related patterns to avoid false positives from
 * unrelated errors like "Email already taken" or "Username already exists"
 */
function isDuplicateError(message: string): boolean {
    const lowerMsg = message.toLowerCase();
    return CLIP_DUPLICATE_PATTERNS.some(pattern => pattern.test(lowerMsg));
}

/**
 * Helper to extract clip information from error response
 * Looks for clip_id, clip_slug at top level or nested in clip object
 */
function extractClipInfo(responseData: unknown): {
    clipId?: string;
    clipSlug?: string;
} {
    if (!responseData || typeof responseData !== 'object') {
        return {};
    }

    const data = responseData as Record<string, unknown>;
    let clipId: string | undefined;
    let clipSlug: string | undefined;

    // Check for clip_id and clip_slug at top level
    if ('clip_id' in data && typeof data.clip_id === 'string') {
        clipId = data.clip_id;
    }
    if ('clip_slug' in data && typeof data.clip_slug === 'string') {
        clipSlug = data.clip_slug;
    }

    // Check for nested clip object (only use if top-level values not found)
    if ('clip' in data && data.clip && typeof data.clip === 'object') {
        const clip = data.clip as Record<string, unknown>;
        if (!clipId && 'id' in clip && typeof clip.id === 'string') {
            clipId = clip.id;
        }
        if (!clipSlug && 'slug' in clip && typeof clip.slug === 'string') {
            clipSlug = clip.slug;
        }
    }

    return { clipId, clipSlug };
}

const SOURCE_SELECTOR_OPTIONS: Array<{
    value: ClipSourceType;
    label: string;
    description: string;
}> = [
    {
        value: 'twitch',
        label: 'Twitch clip URL',
        description: 'Metadata preview, duplicate checks, and claim flow stay enabled.',
    },
    {
        value: 'external',
        label: 'External URL',
        description: 'Use supported Kick, TikTok, YouTube, or YouTube Shorts links.',
    },
    {
        value: 'upload',
        label: 'Upload video',
        description: 'Upload a file for moderator review before publication.',
    },
];

function createEmptyFormData(): SubmitClipRequest {
    return {
        clip_url: '',
        source_type: 'twitch',
        custom_title: '',
        is_nsfw: false,
        submission_reason: '',
        broadcaster_name_override: '',
    };
}

function formatFileSize(bytes: number): string {
    if (!Number.isFinite(bytes) || bytes <= 0) {
        return '0 B';
    }

    const units = ['B', 'KB', 'MB', 'GB', 'TB'];
    const index = Math.min(Math.floor(Math.log(bytes) / Math.log(1024)), units.length - 1);
    const value = bytes / 1024 ** index;
    return `${value >= 10 || index === 0 ? value.toFixed(0) : value.toFixed(1)} ${units[index]}`;
}

function normalizeUrl(value: string): string {
    return value.trim();
}

function getExternalSourcePlatform(urlString: string): ClipSourcePlatform | null {
    try {
        const parsed = new URL(urlString);
        const hostname = parsed.hostname.replace(/^www\./, '').toLowerCase();
        const pathname = parsed.pathname.toLowerCase();

        if (hostname === 'kick.com') {
            return 'kick';
        }

        if (hostname === 'tiktok.com' || hostname.endsWith('.tiktok.com')) {
            return 'tiktok';
        }

        if (
            hostname === 'youtube.com' ||
            hostname.endsWith('.youtube.com') ||
            hostname === 'youtu.be'
        ) {
            return pathname.startsWith('/shorts/') ? 'youtube_shorts' : 'youtube';
        }

        return null;
    } catch {
        return null;
    }
}

function validateSourceUrl(sourceType: ClipSourceType, value: string): string | null {
    const url = normalizeUrl(value);

    if (!url) {
        return null;
    }

    try {
        const parsed = new URL(url);
        if (parsed.protocol !== 'http:' && parsed.protocol !== 'https:') {
            return 'Invalid URL format - please enter a valid URL';
        }
    } catch {
        return 'Invalid URL format - please enter a valid URL';
    }

    if (sourceType === 'twitch') {
        const twitchClipPattern = /^(https?:\/\/)?(clips\.twitch\.tv\/[a-zA-Z0-9_-]+|www\.twitch\.tv\/[^/]+\/clip\/[a-zA-Z0-9_-]+|twitch\.tv\/[^/]+\/clip\/[a-zA-Z0-9_-]+)$/;
        return twitchClipPattern.test(url)
            ? null
            : 'Please enter a valid Twitch clip URL';
    }

    return getExternalSourcePlatform(url)
        ? null
        : 'Use a supported Kick, TikTok, YouTube, or YouTube Shorts URL';
}

export function SubmitClipPage() {
    const { user, isAuthenticated } = useAuth();
    const navigate = useNavigate();
    const location = useLocation();
    const queryClient = useQueryClient();
    const [fromDiscover, setFromDiscover] = useState(false);
    const [formData, setFormData] = useState<SubmitClipRequest>(
        createEmptyFormData(),
    );
    const [selectedTags, setSelectedTags] = useState<Tag[]>([]);
    const [uploadFile, setUploadFile] = useState<File | null>(null);
    const [uploadProgress, setUploadProgress] = useState<number | null>(null);
    const [uploadError, setUploadError] = useState<string | null>(null);
    const [tagQueryLoading, setTagQueryLoading] = useState(false);
    const [isSubmitting, setIsSubmitting] = useState(false);
    const [error, setError] = useState<string | null>(null);
    const [urlError, setUrlError] = useState<string | null>(null);
    const [rateLimitError, setRateLimitError] =
        useState<RateLimitErrorResponse | null>(null);
    const [duplicateError, setDuplicateError] = useState<{
        message: string;
        clipId?: string;
        clipSlug?: string;
    } | null>(null);
    const [submittedClip, setSubmittedClip] = useState<ClipSubmission | null>(
        null,
    );
    const [recentSubmissions, setRecentSubmissions] = useState<
        ClipSubmission[]
    >([]);
    const selectedSource = formData.source_type ?? 'twitch';
    const selectedExternalPlatform =
        selectedSource === 'external' && formData.clip_url ?
            getExternalSourcePlatform(formData.clip_url)
        :   null;
    // Draft management
    const draft = useSubmissionDraft();
    const [showDraftRestored, setShowDraftRestored] = useState(false);

    // Clip submission is available to every authenticated, non-banned user.
    const canSubmit =
        isAuthenticated &&
        user &&
        !rateLimitError &&
        !duplicateError;

    // Helper function to convert text to slug format
    const slugify = useMemo(
        () => (value: string) =>
            value
                .toLowerCase()
                .trim()
                .replace(/[^a-z0-9\s-]/g, '')
                .replace(/\s+/g, '-')
                .replace(/-+/g, '-')
                .replace(/^-|-$/g, ''),
        [],
    );

    // Pre-fill from navigation state or URL query (e.g., when claiming a scraped clip)
    // or restore from draft
    useEffect(() => {
        let timeoutId: number | undefined;

        const state = location.state as {
            clipUrl?: string;
            fromDiscover?: boolean;
        } | null;
        const searchParams = new URLSearchParams(location.search);
        const urlFromQuery =
            searchParams.get('url') || searchParams.get('clip_url');

        // Detect if user came from discover page
        const cameFromDiscover =
            state?.fromDiscover ||
            searchParams.get('from') === 'discover' ||
            document.referrer.includes('/discover');
        if (cameFromDiscover) {
            setFromDiscover(true);
        }

        if (state?.clipUrl) {
            setFormData(prev => ({
                ...prev,
                clip_url: state.clipUrl!,
                source_type: 'twitch',
            }));
        } else if (urlFromQuery) {
            setFormData(prev => ({
                ...prev,
                clip_url: urlFromQuery,
                source_type: 'twitch',
            }));
        } else {
            // Try to load draft if no state from navigation
            const savedDraft = draft.loadDraft();
            if (savedDraft) {
                setFormData({
                    ...createEmptyFormData(),
                    ...savedDraft.formData,
                    source_type: savedDraft.formData.source_type ?? 'twitch',
                });
                setSelectedTags(savedDraft.selectedTags);
                setShowDraftRestored(true);
                // Auto-hide the restored message after 5 seconds
                timeoutId = setTimeout(
                    () => setShowDraftRestored(false),
                    5000,
                ) as unknown as number;
            }
        }

        return () => {
            if (timeoutId !== undefined) {
                clearTimeout(timeoutId);
            }
        };
    }, [location.state, location.search, draft.loadDraft]); // eslint-disable-line react-hooks/exhaustive-deps

    // Load rate limit from localStorage on mount
    useEffect(() => {
        const storedRateLimit = localStorage.getItem('submission_rate_limit');
        if (storedRateLimit) {
            try {
                const rateLimitData: RateLimitErrorResponse =
                    JSON.parse(storedRateLimit);
                const now = Math.floor(Date.now() / 1000);
                // Only restore if still active
                if (rateLimitData.retry_after > now) {
                    setRateLimitError(rateLimitData);
                } else {
                    // Clear expired rate limit
                    localStorage.removeItem('submission_rate_limit');
                }
            } catch (err) {
                console.error('Failed to parse stored rate limit:', err);
                localStorage.removeItem('submission_rate_limit');
            }
        }
    }, []);

    // Load recent submissions
    useEffect(() => {
        let isMounted = true;

        if (isAuthenticated) {
            getUserSubmissions(1, 5)
                .then(response => {
                    if (isMounted) {
                        setRecentSubmissions(response.data || []);
                    }
                })
                .catch(err => {
                    if (isMounted) {
                        console.error('Failed to load submissions:', err);
                        setRecentSubmissions([]);
                    }
                });
        }

        return () => {
            isMounted = false;
        };
    }, [isAuthenticated]);

    // Helper function to extract clip ID from Twitch URL
    const extractClipIDFromURL = (url: string): string | null => {
        if (!url) return null;

        // Match patterns like:
        // https://clips.twitch.tv/AwkwardHelplessSalamanderSwiftRage
        // https://www.twitch.tv/broadcaster/clip/AwkwardHelplessSalamanderSwiftRage
        const clipsTwitchPattern = /clips\.twitch\.tv\/([a-zA-Z0-9_-]+)/;
        const twitchClipPattern = /twitch\.tv\/[^/]+\/clip\/([a-zA-Z0-9_-]+)/;

        let match = url.match(clipsTwitchPattern);
        if (match) return match[1];

        match = url.match(twitchClipPattern);
        if (match) return match[1];

        return null;
    };

    // Auto-set NSFW if clip already marked (best effort) when URL changes
    // Also check for duplicates and show error proactively
    useEffect(() => {
        if (selectedSource !== 'twitch') {
            setDuplicateError(null);
            return;
        }

        const clipID = extractClipIDFromURL(formData.clip_url);

        // Clear duplicate error when URL is empty or invalid
        if (!clipID) {
            setDuplicateError(null);
            return;
        }

        let isActive = true;
        checkClipStatus(clipID)
            .then(resp => {
                if (!isActive) return;

                // Auto-set NSFW if clip already marked
                if (resp?.clip?.is_nsfw) {
                    setFormData(prev => ({ ...prev, is_nsfw: true }));
                }

                // Check if clip already exists (duplicate detection)
                if (resp?.exists && !resp?.can_be_claimed) {
                    setDuplicateError({
                        message:
                            'This clip has already been submitted to the database.',
                        clipId: resp.clip?.id,
                        clipSlug: resp.clip?.twitch_clip_id,
                    });
                } else {
                    setDuplicateError(null);
                }
            })
            .catch(() => {
                // ignore; optional helper
                setDuplicateError(null);
            });

        return () => {
            isActive = false;
        };
    }, [formData.clip_url, selectedSource]);

    // Auto-fill title and tags when clip URL is pasted
    useEffect(() => {
        if (selectedSource !== 'twitch' || !formData.clip_url) return;

        let isActive = true;
        getClipMetadata(formData.clip_url)
            .then(metadata => {
                if (!isActive) return;

                // Auto-fill custom title if not already filled
                if (!formData.custom_title) {
                    setFormData(prev => ({
                        ...prev,
                        custom_title: metadata.title,
                    }));
                }

                // Auto-add game tag if not already added
                if (metadata.game_name && selectedTags.length === 0) {
                    const gameTag: Tag = {
                        id: `temp-${slugify(metadata.game_name)}`,
                        name: metadata.game_name,
                        slug: slugify(metadata.game_name),
                        usage_count: 0,
                        created_at: new Date().toISOString(),
                    };
                    setSelectedTags([gameTag]);
                }
            })
            .catch(() => {
                // ignore; optional helper
            });

        return () => {
            isActive = false;
        };
    }, [
        formData.clip_url,
        formData.custom_title,
        selectedTags.length,
        slugify,
        selectedSource,
    ]);

    // Auto-save draft every 30 seconds when form has content
    useEffect(() => {
        draft.startAutoSave(formData, selectedTags);
    }, [formData, selectedTags, draft]);

    const handleCreateTag = async (name: string): Promise<Tag | null> => {
        const slug = slugify(name);
        setTagQueryLoading(true);
        try {
            const resp = await tagApi.createTag({ name: name.trim(), slug });
            return resp.tag;
        } catch (err) {
            // Fallback: return a local tag representation even if API fails
            console.error('Failed to create tag, using local fallback', err);
            return {
                id: `temp-${slug}`,
                name: name.trim(),
                slug,
                usage_count: 0,
                created_at: new Date().toISOString(),
            } as Tag;
        } finally {
            setTagQueryLoading(false);
        }
    };

    const handleTagsChange = (tags: Tag[]) => {
        setSelectedTags(tags);
    };

    const tagsToSubmit = selectedTags.map(tag => tag.slug || slugify(tag.name));

    const handleSourceChange = (nextSource: ClipSourceType) => {
        setFormData(prev => ({
            ...prev,
            source_type: nextSource,
        }));
        setError(null);
        setUrlError(null);
        setUploadError(null);
        setDuplicateError(null);
        setUploadProgress(null);
        if (nextSource === 'upload') {
            setSelectedTags([]);
        } else {
            setUploadFile(null);
        }
    };

    const handleUrlChange = (value: string) => {
        setFormData(prev => ({
            ...prev,
            clip_url: value,
        }));
        setError(null);
        setUrlError(null);
        setDuplicateError(null);
    };

    const handleUrlBlur = (value: string) => {
        if (!value) {
            setUrlError(null);
            return;
        }

        const validationError = validateSourceUrl(selectedSource, value);
        setUrlError(validationError);
    };

    const handleUploadChange = (file: File | null) => {
        setUploadFile(file);
        setError(null);
        setUploadError(null);
        setUploadProgress(null);
    };

    const resetSubmissionForm = () => {
        setFormData(createEmptyFormData());
        setSelectedTags([]);
        setUploadFile(null);
        setUploadProgress(null);
        setUploadError(null);
        setUrlError(null);
        setDuplicateError(null);
        setError(null);
    };

    const handleSubmit = async (e: React.FormEvent) => {
        e.preventDefault();

        if (selectedSource === 'upload') {
            if (!uploadFile) {
                setUploadError('Please choose a video file to upload');
                return;
            }
        } else {
            const normalizedClipUrl = normalizeUrl(formData.clip_url);
            if (!normalizedClipUrl) {
                setUrlError(
                    selectedSource === 'external' ?
                        'Please enter an external URL'
                    :   'Please enter a Twitch clip URL',
                );
                return;
            }

            const validationError = validateSourceUrl(
                selectedSource,
                normalizedClipUrl,
            );
            if (validationError) {
                setUrlError(validationError);
                return;
            }
        }
        setUrlError(null);
        setUploadError(null);

        if (!canSubmit) {
            setError('You must be logged in to submit clips');
            return;
        }

        setError(null);
        setSubmittedClip(null);
        setDuplicateError(null);
        setRateLimitError(null);
        setUploadProgress(selectedSource === 'upload' ? 0 : null);
        setIsSubmitting(true);

        try {
            const response =
                selectedSource === 'upload' && uploadFile ?
                    await submitClipUpload(
                        {
                            file: uploadFile,
                            custom_title: formData.custom_title?.trim() || '',
                            is_nsfw: formData.is_nsfw,
                            submission_reason:
                                formData.submission_reason?.trim() || '',
                        },
                        event => {
                            const progressFraction =
                                typeof event.progress === 'number' ?
                                    event.progress
                                : event.total && event.total > 0 ?
                                    event.loaded / event.total
                                :   null;

                            if (progressFraction !== null) {
                                setUploadProgress(
                                    Math.min(
                                        100,
                                        Math.round(progressFraction * 100),
                                    ),
                                );
                            }
                        },
                    )
                :   await submitClip({
                        ...formData,
                        clip_url: normalizeUrl(formData.clip_url),
                        source_type: selectedSource,
                        source_platform:
                            selectedSource === 'twitch' ? 'twitch' : selectedExternalPlatform ?? undefined,
                        source_url: normalizeUrl(formData.clip_url),
                        source_id:
                            selectedSource === 'twitch' ?
                                extractClipIDFromURL(formData.clip_url) ?? undefined
                            :   undefined,
                        tags: tagsToSubmit,
                        // omit broadcaster override if empty to let backend auto-detect
                        broadcaster_name_override:
                            formData.broadcaster_name_override?.trim() ?
                                formData.broadcaster_name_override
                            :   undefined,
                    });
            // Set the submitted clip to show confirmation
            setSubmittedClip(response.submission);

            // Track successful submission
            trackEvent(SubmissionEvents.SUBMISSION_CREATE_COMPLETED, {
                submission_id: response.submission.id,
                title: response.submission.title,
                is_nsfw: response.submission.is_nsfw,
                tags: Array.isArray(tagsToSubmit) ? tagsToSubmit : [],
            });

            // Clear draft on successful submission
            draft.clearDraft();

            // Invalidate scraped-clips cache so the submitted clip disappears from discovery
            queryClient.invalidateQueries({ queryKey: ['scraped-clips'] });

            // Reset form
            resetSubmissionForm();
        } catch (err: unknown) {
            const error = err as {
                response?: {
                    status?: number;
                    data?: unknown;
                };
            };

            // Check for rate limit error (429)
            if (error.response?.status === 429) {
                const data = error.response.data;
                // Type guard to verify rate limit error structure
                if (
                    data &&
                    typeof data === 'object' &&
                    'error' in data &&
                    data.error === 'rate_limit_exceeded' &&
                    'retry_after' in data &&
                    typeof data.retry_after === 'number' &&
                    'limit' in data &&
                    typeof data.limit === 'number' &&
                    'window' in data &&
                    typeof data.window === 'number'
                ) {
                    const rateLimitData: RateLimitErrorResponse = {
                        error: data.error,
                        limit: data.limit,
                        window: data.window,
                        retry_after: data.retry_after,
                    };

                    setRateLimitError(rateLimitData);
                    setError(null);
                    // Store in localStorage for persistence
                    try {
                        localStorage.setItem(
                            'submission_rate_limit',
                            JSON.stringify(rateLimitData),
                        );
                    } catch (storageError) {
                        // Ignore localStorage errors - rate limit will still work for current session
                        console.warn(
                            'Failed to persist rate limit to localStorage:',
                            storageError,
                        );
                    }
                    // Track rate limit hit
                    trackEvent(SubmissionEvents.SUBMISSION_RATE_LIMIT_HIT, {
                        limit: rateLimitData.limit,
                        window: rateLimitData.window,
                        retry_after: rateLimitData.retry_after,
                    });
                    return;
                }
            }

            // Handle other errors
            const data = error.response?.data;
            let errorMessage = 'Failed to submit clip';
            if (
                data &&
                typeof data === 'object' &&
                'error' in data &&
                typeof data.error === 'string'
            ) {
                errorMessage = data.error;

                if (isDuplicateError(errorMessage)) {
                    const { clipId, clipSlug } = extractClipInfo(data);
                    setDuplicateError({
                        message: errorMessage,
                        clipId,
                        clipSlug,
                    });
                    setError(null);
                } else {
                    setError(errorMessage);
                    setDuplicateError(null);
                }
            } else {
                setError(errorMessage);
                setDuplicateError(null);
            }

            if (selectedSource === 'upload') {
                setUploadError(errorMessage);
                setUploadProgress(null);
            }

            // Track failed submission
            trackEvent(SubmissionEvents.SUBMISSION_CREATE_FAILED, {
                error: errorMessage,
            });
        } finally {
            setIsSubmitting(false);
        }
    };

    const handleRateLimitExpire = () => {
        setRateLimitError(null);

        // Read stored rate limit metadata before clearing
        let metadata: Record<string, unknown> = {};
        try {
            const storedRateLimit = localStorage.getItem(
                'submission_rate_limit',
            );
            if (storedRateLimit) {
                const parsed = JSON.parse(
                    storedRateLimit,
                ) as RateLimitErrorResponse;
                metadata = {
                    limit: parsed.limit,
                    window: parsed.window,
                };
            }
        } catch (error) {
            // Ignore parsing errors
            console.warn(
                'Failed to read rate limit metadata for analytics:',
                error,
            );
        }

        // Clear from localStorage
        try {
            localStorage.removeItem('submission_rate_limit');
        } catch (error) {
            // Ignore localStorage errors
            console.warn(
                'Failed to remove rate limit from localStorage:',
                error,
            );
        }

        // Track rate limit expiration with metadata
        trackEvent(SubmissionEvents.SUBMISSION_RATE_LIMIT_EXPIRED, metadata);
    };

    const handleRateLimitDismiss = () => {
        setRateLimitError(null);
        try {
            localStorage.removeItem('submission_rate_limit');
        } catch (error) {
            // Ignore localStorage errors
            console.warn(
                'Failed to remove rate limit from localStorage:',
                error,
            );
        }
    };

    const handleSubmitAnother = () => {
        setSubmittedClip(null);
        setError(null);
        setDuplicateError(null);
    };

    if (!isAuthenticated) {
        return (
            <Container className='py-4 xs:py-6 md:py-8'>
                <Card className='max-w-2xl mx-auto p-4 xs:p-6 md:p-8 text-center'>
                    <h1 className='text-2xl xs:text-3xl font-bold mb-4'>
                        Submit a Clip
                    </h1>
                    <p className='text-sm xs:text-base text-muted-foreground mb-6'>
                        You must be logged in to submit clips.
                    </p>
                    <Button onClick={() => navigate('/login')}>Log In</Button>
                </Card>
            </Container>
        );
    }

    // Show confirmation view after successful submission
    if (submittedClip) {
        return (
            <Container className='py-4 xs:py-6 md:py-8'>
                <SubmissionConfirmation
                    submission={submittedClip}
                    onSubmitAnother={handleSubmitAnother}
                    fromDiscover={fromDiscover}
                />
            </Container>
        );
    }

    return (
        <Container className='py-4 xs:py-6 md:py-8'>
            <div data-testid='submit-clip-main-content' className='max-w-3xl mx-auto'>
                <div className='mb-4 xs:mb-6'>
                    <h1 className='text-2xl xs:text-3xl font-bold mb-2'>
                        Submit a Clip
                    </h1>
                    <p className='text-sm xs:text-base text-muted-foreground'>
                        Share a memorable moment from any Twitch creator or community
                    </p>
                </div>

                {rateLimitError && (
                    <div className='mb-4 xs:mb-6'>
                        <RateLimitError
                            retryAfter={rateLimitError.retry_after}
                            limit={rateLimitError.limit}
                            window={rateLimitError.window}
                            onExpire={handleRateLimitExpire}
                            onDismiss={handleRateLimitDismiss}
                        />
                    </div>
                )}

                {duplicateError && (
                    <div className='mb-4 xs:mb-6'>
                        <DuplicateClipError
                            message={duplicateError.message}
                            clipId={duplicateError.clipId}
                            clipSlug={duplicateError.clipSlug}
                            onDismiss={() => setDuplicateError(null)}
                        />
                    </div>
                )}

                {error && (
                    <Alert variant='error' className='mb-6'>
                        {error}
                    </Alert>
                )}

                {showDraftRestored && (
                    <Alert
                        variant='info'
                        className='mb-6'
                        dismissible={true}
                        onDismiss={() => setShowDraftRestored(false)}
                    >
                        Draft restored from your last session
                    </Alert>
                )}

                {draft.hasDraft && draft.lastSaved && (
                    <div className='mb-4 xs:mb-6 flex items-center justify-between bg-blue-500/10 border border-blue-500/20 rounded-lg p-3'>
                        <div className='flex items-center gap-2 text-sm text-blue-400'>
                            <svg
                                className='w-4 h-4'
                                fill='none'
                                stroke='currentColor'
                                viewBox='0 0 24 24'
                            >
                                <path
                                    strokeLinecap='round'
                                    strokeLinejoin='round'
                                    strokeWidth={2}
                                    d='M5 13l4 4L19 7'
                                />
                            </svg>
                            <span>
                                Draft saved{' '}
                                {new Date(draft.lastSaved).toLocaleTimeString()}
                            </span>
                        </div>
                        <Button
                            type='button'
                            variant='ghost'
                            size='sm'
                            onClick={() => {
                                draft.clearDraft();
                                resetSubmissionForm();
                            }}
                            className='text-blue-400 hover:text-blue-300'
                        >
                            Clear Draft
                        </Button>
                    </div>
                )}

                <Card className='p-6 mb-8'>
                    <form onSubmit={handleSubmit} noValidate>
                        <div className='space-y-6'>
                            {/* Source Selector */}
                            <div>
                                <label className='block text-sm font-medium mb-3'>
                                    Clip Source <span className='text-red-500'>*</span>
                                </label>
                                <div
                                    className='grid gap-3 md:grid-cols-3'
                                    role='radiogroup'
                                    aria-label='Clip source'
                                >
                                    {SOURCE_SELECTOR_OPTIONS.map(option => {
                                        const selected =
                                            selectedSource === option.value;

                                        return (
                                            <button
                                                key={option.value}
                                                type='button'
                                                role='radio'
                                                aria-checked={selected}
                                                disabled={!canSubmit || isSubmitting}
                                                onClick={() =>
                                                    handleSourceChange(option.value)
                                                }
                                                className={`rounded-xl border p-4 text-left transition-all duration-200 ${
                                                    selected ?
                                                        'border-primary-500 bg-primary-500/10 shadow-sm'
                                                    :   'border-border bg-background hover:border-primary-300 hover:bg-primary-50/40 dark:hover:bg-primary-950/20'
                                                } ${!canSubmit || isSubmitting ? 'opacity-60 cursor-not-allowed' : 'cursor-pointer'}`}
                                            >
                                                <div className='flex items-start justify-between gap-3'>
                                                    <div>
                                                        <p className='font-medium'>
                                                            {option.label}
                                                        </p>
                                                        <p className='mt-1 text-xs text-muted-foreground leading-relaxed'>
                                                            {option.description}
                                                        </p>
                                                    </div>
                                                    <span
                                                        className={`mt-0.5 h-2.5 w-2.5 rounded-full border ${
                                                            selected ?
                                                                'border-primary-500 bg-primary-500'
                                                            :   'border-muted-foreground/30'
                                                        }`}
                                                    />
                                                </div>
                                            </button>
                                        );
                                    })}
                                </div>
                            </div>

                            {/* Source-specific URL / Upload Control */}
                            {selectedSource !== 'upload' ? (
                                <div>
                                    <label
                                        htmlFor='clip_url'
                                        className='block text-sm font-medium mb-2'
                                    >
                                        {selectedSource === 'external' ?
                                            'External URL'
                                        :   'Twitch Clip URL'}{' '}
                                        <span className='text-red-500'>*</span>
                                    </label>
                                    <Input
                                        id='clip_url'
                                        name='url'
                                        type='url'
                                        value={formData.clip_url}
                                        onChange={e => {
                                            handleUrlChange(e.target.value);
                                        }}
                                        onBlur={e => {
                                            handleUrlBlur(e.target.value);
                                        }}
                                        placeholder={
                                            selectedSource === 'external' ?
                                                'https://www.youtube.com/watch?v=...'
                                            :   'https://clips.twitch.tv/...'
                                        }
                                        required
                                        disabled={!canSubmit || isSubmitting}
                                    />
                                    {urlError && (
                                        <p className='text-xs text-red-500 mt-1'>
                                            {urlError}
                                        </p>
                                    )}
                                    <p className='text-xs text-muted-foreground mt-1'>
                                        {selectedSource === 'external' ?
                                            'Supported platforms: Kick, TikTok, YouTube, and YouTube Shorts.'
                                        :   'Paste the full URL of a Twitch clip.'}
                                    </p>
                                </div>
                            ) : (
                                <div className='space-y-3'>
                                    <div>
                                        <label
                                            htmlFor='clip_file'
                                            className='block text-sm font-medium mb-2'
                                        >
                                            Video File <span className='text-red-500'>*</span>
                                        </label>
                                        <Input
                                            id='clip_file'
                                            type='file'
                                            accept='video/*'
                                            disabled={!canSubmit || isSubmitting}
                                            onChange={e =>
                                                handleUploadChange(
                                                    e.currentTarget.files?.[0] ?? null,
                                                )
                                            }
                                            className='cursor-pointer file:mr-4 file:rounded-md file:border-0 file:bg-primary-500 file:px-4 file:py-2 file:text-sm file:font-medium file:text-white hover:file:bg-primary-600'
                                        />
                                    </div>
                                    <div className='rounded-xl border border-border/70 bg-background-secondary/40 p-4 text-sm text-muted-foreground space-y-2'>
                                        <p>
                                            Maximum duration: 10 minutes. Recommended: 7 minutes or less.
                                        </p>
                                        <p>
                                            Maximum upload size is configured by the server. 1080p60 is allowed if it fits within the duration and size limits.
                                        </p>
                                        <p>
                                            Uploads require moderator approval before publication.
                                        </p>
                                    </div>
                                    {uploadFile && (
                                        <div className='rounded-xl border border-primary-500/20 bg-primary-500/5 p-4'>
                                            <div className='flex items-start justify-between gap-4'>
                                                <div className='min-w-0'>
                                                    <p className='font-medium truncate'>
                                                        {uploadFile.name}
                                                    </p>
                                                    <p className='text-xs text-muted-foreground mt-1'>
                                                        {formatFileSize(uploadFile.size)}
                                                        {uploadFile.type ? ` • ${uploadFile.type}` : ''}
                                                    </p>
                                                </div>
                                                <span className='rounded-full border border-primary-500/20 bg-primary-500/10 px-2.5 py-1 text-xs font-medium text-primary-500'>
                                                    Ready to upload
                                                </span>
                                            </div>
                                            {uploadProgress !== null && isSubmitting && (
                                                <div className='mt-4'>
                                                    <div className='h-2 overflow-hidden rounded-full bg-muted/80'>
                                                        <div
                                                            className='h-full rounded-full bg-primary-500 transition-all duration-200'
                                                            style={{ width: `${uploadProgress}%` }}
                                                        />
                                                    </div>
                                                    <p className='mt-2 text-xs text-muted-foreground'>
                                                        {uploadProgress}% uploaded
                                                    </p>
                                                </div>
                                            )}
                                        </div>
                                    )}
                                    {uploadError && (
                                        <p className='text-xs text-red-500'>
                                            {uploadError}
                                        </p>
                                    )}
                                </div>
                            )}

                            {/* Custom Title */}
                            <div>
                                <label
                                    htmlFor='custom_title'
                                    className='block text-sm font-medium mb-2'
                                >
                                    Custom Title (Optional)
                                </label>
                                <Input
                                    id='custom_title'
                                    type='text'
                                    value={formData.custom_title}
                                    onChange={e =>
                                        setFormData({
                                            ...formData,
                                            custom_title: e.target.value,
                                        })
                                    }
                                    placeholder='Give your clip a catchy title (optional)'
                                    disabled={!canSubmit || isSubmitting}
                                />
                            </div>

                            {/* Tags */}
                            {selectedSource !== 'upload' ? (
                                <div>
                                    <label
                                        htmlFor='tags'
                                        className='block text-sm font-medium mb-2'
                                    >
                                        Tags (Optional)
                                    </label>
                                    <TagSelector
                                        selectedTags={selectedTags}
                                        onTagsChange={handleTagsChange}
                                        maxTags={10}
                                        allowCreate
                                        onCreateTag={handleCreateTag}
                                        helperText='Search popular tags or add your own. New tags will be saved.'
                                        placeholder='Search or add tags...'
                                    />
                                    {tagQueryLoading && (
                                        <p className='text-xs text-muted-foreground mt-1'>
                                            Creating tag...
                                        </p>
                                    )}
                                </div>
                            ) : (
                                <div className='rounded-xl border border-border/70 bg-background-secondary/40 p-4 text-sm text-muted-foreground'>
                                    Tags are available for URL submissions only.
                                </div>
                            )}

                            {/* NSFW Checkbox */}
                            <div className='flex items-center gap-2'>
                                <Checkbox
                                    id='is_nsfw'
                                    checked={formData.is_nsfw}
                                    onChange={e =>
                                        setFormData({
                                            ...formData,
                                            is_nsfw: e.target.checked,
                                        })
                                    }
                                    disabled={!canSubmit || isSubmitting}
                                />
                                <label
                                    htmlFor='is_nsfw'
                                    className='text-sm font-medium cursor-pointer'
                                >
                                    Mark as NSFW
                                </label>
                            </div>

                            {/* Submission Reason */}
                            <div>
                                <label
                                    htmlFor='submission_reason'
                                    className='block text-sm font-medium mb-2'
                                >
                                    Submission Reason (Optional)
                                </label>
                                <TextArea
                                    id='submission_reason'
                                    value={formData.submission_reason}
                                    onChange={e =>
                                        setFormData({
                                            ...formData,
                                            submission_reason: e.target.value,
                                        })
                                    }
                                    placeholder='Why is this clip noteworthy?'
                                    rows={3}
                                    disabled={!canSubmit || isSubmitting}
                                />
                            </div>

                            {/* Submit Button */}
                            <div className='flex gap-3'>
                                <Button
                                    type='submit'
                                    disabled={
                                        !canSubmit ||
                                        isSubmitting ||
                                        (selectedSource === 'upload' ?
                                            !uploadFile
                                        :   !normalizeUrl(formData.clip_url))
                                    }
                                    className='flex-1'
                                >
                                    {isSubmitting ?
                                        selectedSource === 'upload' ?
                                            'Uploading...'
                                        :   'Submitting...'
                                    :   selectedSource === 'upload' ?
                                        'Upload Clip'
                                    :   'Submit Clip'}
                                </Button>
                                <Button
                                    type='button'
                                    variant='secondary'
                                    onClick={() => navigate('/submissions')}
                                >
                                    My Submissions
                                </Button>
                            </div>
                        </div>
                    </form>
                </Card>

                {/* Recent Submissions */}
                {Array.isArray(recentSubmissions) &&
                    recentSubmissions.length > 0 && (
                        <Card className='p-6'>
                            <h2 className='text-xl font-bold mb-4'>
                                Your Recent Submissions
                            </h2>
                            <div className='space-y-3'>
                                {recentSubmissions.map(submission => (
                                    <div
                                        key={submission.id}
                                        className='flex items-center justify-between p-3 bg-background-secondary rounded-lg'
                                    >
                                        <div className='flex-1'>
                                            <p className='font-medium truncate'>
                                                {submission.custom_title ||
                                                    submission.title ||
                                                    'Untitled'}
                                            </p>
                                            <p className='text-xs text-muted-foreground'>
                                                {new Date(
                                                    submission.created_at,
                                                ).toLocaleDateString()}
                                            </p>
                                        </div>
                                        <span
                                            className={`px-3 py-1 rounded-full text-xs font-medium ${
                                                (
                                                    submission.status ===
                                                    'approved'
                                                ) ?
                                                    'bg-green-500/20 text-green-500'
                                                : (
                                                    submission.status ===
                                                    'rejected'
                                                ) ?
                                                    'bg-red-500/20 text-red-500'
                                                :   'bg-yellow-500/20 text-yellow-500'
                                            }`}
                                        >
                                            {submission.status}
                                        </span>
                                    </div>
                                ))}
                            </div>
                        </Card>
                    )}
            </div>
        </Container>
    );
}
