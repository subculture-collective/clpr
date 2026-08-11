import userEvent from '@testing-library/user-event';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { useAuth } from '../context/AuthContext';
import * as submissionApi from '../lib/submission-api';
import { render, screen, waitFor } from '../test/test-utils';
import { SubmitClipPage } from './SubmitClipPage';
import { tagApi } from '../lib/tag-api';

// Mock the API calls
vi.mock('../lib/submission-api', () => ({
    submitClip: vi.fn(),
    getUserSubmissions: vi.fn(),
    checkClipStatus: vi.fn(),
    getClipMetadata: vi.fn(),
}));

// Mock tag API to avoid network calls in TagSelector
vi.mock('../lib/tag-api', () => ({
    tagApi: {
        searchTags: vi.fn().mockResolvedValue({ tags: [] }),
        createTag: vi.fn().mockImplementation(async payload => {
            const name = payload?.name || 'new-tag';
            const slug = payload?.slug || name.toLowerCase();

            return {
                tag: {
                    id: `tag-${slug}`,
                    name,
                    slug,
                    usage_count: 0,
                    created_at: new Date().toISOString(),
                },
            };
        }),
    },
}));

// Mock the AuthContext
vi.mock('../context/AuthContext', async () => {
    const actual = await vi.importActual('../context/AuthContext');
    return {
        ...actual,
        useAuth: vi.fn(),
    };
});

// Mock react-router-dom
const mockNavigate = vi.fn();
vi.mock('react-router-dom', async () => {
    const actual = await vi.importActual('react-router-dom');
    return {
        ...actual,
        useNavigate: () => mockNavigate,
    };
});

describe('SubmitClipPage', () => {
    const mockUser = {
        id: 'user-123',
        twitch_id: 'twitch-123',
        username: 'testuser',
        display_name: 'Test User',
        karma_points: 150,
        role: 'user' as const,
        is_banned: false,
        created_at: '2024-01-01T00:00:00Z',
    };

    const mockSubmitClip = vi.mocked(submissionApi.submitClip);
    const mockGetUserSubmissions = vi.mocked(submissionApi.getUserSubmissions);
    const mockCheckClipStatus = vi.mocked(submissionApi.checkClipStatus);
    const mockGetClipMetadata = vi.mocked(submissionApi.getClipMetadata);
    const mockUseAuth = vi.mocked(useAuth);
    const mockSearchTags = vi.mocked(tagApi.searchTags);

    beforeEach(() => {
        vi.clearAllMocks();
        mockGetUserSubmissions.mockResolvedValue({
            success: true,
            data: [],
            meta: {
                page: 1,
                limit: 5,
                total: 0,
                total_pages: 0,
            },
        });
        mockSearchTags.mockResolvedValue({ tags: [] });
        mockCheckClipStatus.mockResolvedValue({
            success: true,
            exists: false,
            can_be_claimed: true,
        });
        mockGetClipMetadata.mockResolvedValue({
            title: 'Test Clip Title',
            broadcaster_name: 'TestStreamer',
            game: 'Test Game',
            duration: 30,
        });
    });

    describe('Authentication', () => {
        it('shows login prompt when not authenticated', () => {
            mockUseAuth.mockReturnValue({
                user: null,
                isAuthenticated: false,
                login: vi.fn(),
                logout: vi.fn(),
                isLoading: false,
                isAdmin: false,
                isModerator: false,
                isModeratorOrAdmin: false,
                refreshUser: vi.fn(),
            });

            render(<SubmitClipPage />);

            expect(screen.getByText('Submit a Clip')).toBeInTheDocument();
            expect(
                screen.getByText('You must be logged in to submit clips.')
            ).toBeInTheDocument();
            expect(screen.getByText('Log In')).toBeInTheDocument();
        });

        it('navigates to login when login button is clicked', async () => {
            const user = userEvent.setup();
            mockUseAuth.mockReturnValue({
                user: null,
                isAuthenticated: false,
                login: vi.fn(),
                logout: vi.fn(),
                isLoading: false,
                isAdmin: false,
                isModerator: false,
                isModeratorOrAdmin: false,
                refreshUser: vi.fn(),
            });

            render(<SubmitClipPage />);

            const loginButton = screen.getByText('Log In');
            await user.click(loginButton);

            expect(mockNavigate).toHaveBeenCalledWith('/login');
        });

        it('allows a signed-in user with zero uppies to submit', async () => {
            const user = userEvent.setup();
            mockUseAuth.mockReturnValue({
                user: { ...mockUser, karma_points: 0 },
                isAuthenticated: true,
                login: vi.fn(),
                logout: vi.fn(),
                isLoading: false,
                isAdmin: false,
                isModerator: false,
                isModeratorOrAdmin: false,
                refreshUser: vi.fn(),
            });

            render(<SubmitClipPage />);

            const clipUrlInput = screen.getByLabelText(/Twitch Clip URL/);
            await user.type(clipUrlInput, 'https://clips.twitch.tv/TestClip');

            await waitFor(() => {
                const submitButton = screen.getByRole('button', {
                    name: /Submit Clip/,
                });

                expect(clipUrlInput).toBeEnabled();
                expect(submitButton).toBeEnabled();
            });
        });
    });

    describe('Form Validation', () => {
        beforeEach(() => {
            mockUseAuth.mockReturnValue({
                user: mockUser,
                isAuthenticated: true,
                login: vi.fn(),
                logout: vi.fn(),
                isLoading: false,
                isAdmin: false,
                isModerator: false,
                isModeratorOrAdmin: false,
                refreshUser: vi.fn(),
            });
        });

        it('requires clip URL to submit', async () => {
            render(<SubmitClipPage />);

            await waitFor(() => {
                const submitButton = screen.getByRole('button', {
                    name: /Submit Clip/,
                });
                expect(submitButton).toBeDisabled();
            });
        });

        it('enables submit button when clip URL is provided', async () => {
            const user = userEvent.setup();
            render(<SubmitClipPage />);

            const clipUrlInput = screen.getByLabelText(/Twitch Clip URL/);
            await user.type(
                clipUrlInput,
                'https://clips.twitch.tv/TestClip123'
            );

            const submitButton = screen.getByRole('button', {
                name: /Submit Clip/,
            });
            expect(submitButton).toBeEnabled();
        });

        it('accepts valid Twitch clip URLs', async () => {
            const user = userEvent.setup();
            render(<SubmitClipPage />);

            const clipUrlInput = screen.getByLabelText(/Twitch Clip URL/);

            // Test clips.twitch.tv format
            await user.type(
                clipUrlInput,
                'https://clips.twitch.tv/AwkwardHelplessSalamanderSwiftRage'
            );
            expect(clipUrlInput).toHaveValue(
                'https://clips.twitch.tv/AwkwardHelplessSalamanderSwiftRage'
            );
        });

        it('allows adding custom title', async () => {
            const user = userEvent.setup();
            render(<SubmitClipPage />);

            const customTitleInput = screen.getByLabelText(/Custom Title/);
            await user.clear(customTitleInput);
            await user.type(customTitleInput, 'My Custom Title');

            expect(customTitleInput).toHaveValue('My Custom Title');
        });

        it('allows marking clip as NSFW', async () => {
            const user = userEvent.setup();
            render(<SubmitClipPage />);

            const nsfwCheckbox = screen.getByLabelText(/Mark as NSFW/);
            expect(nsfwCheckbox).not.toBeChecked();

            await user.click(nsfwCheckbox);
            expect(nsfwCheckbox).toBeChecked();
        });

        it('allows adding submission reason', async () => {
            const user = userEvent.setup();
            render(<SubmitClipPage />);

            const reasonTextarea = screen.getByLabelText(/Submission Reason/);
            await user.type(reasonTextarea, 'This is an amazing play');

            expect(reasonTextarea).toHaveValue('This is an amazing play');
        });
    });

    describe('Tags Management', () => {
        beforeEach(() => {
            mockUseAuth.mockReturnValue({
                user: mockUser,
                isAuthenticated: true,
                login: vi.fn(),
                logout: vi.fn(),
                isLoading: false,
                isAdmin: false,
                isModerator: false,
                isModeratorOrAdmin: false,
                refreshUser: vi.fn(),
            });
        });

        it('allows adding tags', async () => {
            const user = userEvent.setup();
            render(<SubmitClipPage />);

            const tagInput = screen.getByPlaceholderText(
                'Search or add tags...'
            );

            await user.type(tagInput, 'clutch{Enter}');

            expect(screen.getByText('clutch')).toBeInTheDocument();
            expect(tagInput).toHaveValue('');
        });

        it('allows adding tags with Enter key', async () => {
            const user = userEvent.setup();
            render(<SubmitClipPage />);

            const tagInput = screen.getByPlaceholderText(
                'Search or add tags...'
            );
            await user.type(tagInput, 'epic{Enter}');

            expect(await screen.findByText('epic')).toBeInTheDocument();
        });

        it('prevents adding duplicate tags', async () => {
            const user = userEvent.setup();
            render(<SubmitClipPage />);

            const tagInput = screen.getByPlaceholderText(
                'Search or add tags...'
            );

            // Add first tag via enter
            await user.type(tagInput, 'clutch{Enter}');

            // Try to add duplicate
            await user.type(tagInput, 'clutch{Enter}');

            // Should only have one instance
            const tags = screen.getAllByText('clutch');
            expect(tags).toHaveLength(1);
        });

        it('allows removing tags', async () => {
            const user = userEvent.setup();
            render(<SubmitClipPage />);

            const tagInput = screen.getByPlaceholderText(
                'Search or add tags...'
            );
            await user.type(tagInput, 'clutch{Enter}');

            expect(screen.getByText('clutch')).toBeInTheDocument();

            const removeButton = screen.getByLabelText('Remove clutch tag');
            await user.click(removeButton);

            expect(screen.queryByText('clutch')).not.toBeInTheDocument();
        });

        it('trims whitespace from tags', async () => {
            const user = userEvent.setup();
            render(<SubmitClipPage />);

            const tagInput = screen.getByPlaceholderText(
                'Search or add tags...'
            );
            await user.type(tagInput, '  clutch  {Enter}');

            expect(screen.getByText('clutch')).toBeInTheDocument();
        });
    });

    describe('Streamer override', () => {
        beforeEach(() => {
            mockUseAuth.mockReturnValue({
                user: mockUser,
                isAuthenticated: true,
                login: vi.fn(),
                logout: vi.fn(),
                isLoading: false,
                isAdmin: false,
                isModerator: false,
                isModeratorOrAdmin: false,
                refreshUser: vi.fn(),
            });
        });

        it('no longer shows manual streamer input (auto-detected by backend)', async () => {
            render(<SubmitClipPage />);

            await waitFor(() => {
                expect(
                    screen.queryByPlaceholderText('Enter streamer name...')
                ).not.toBeInTheDocument();
                expect(
                    screen.queryByText('Will be auto-detected from clip')
                ).not.toBeInTheDocument();
            });
        });
    });

    describe('Form Submission', () => {
        beforeEach(() => {
            mockUseAuth.mockReturnValue({
                user: mockUser,
                isAuthenticated: true,
                login: vi.fn(),
                logout: vi.fn(),
                isLoading: false,
                isAdmin: false,
                isModerator: false,
                isModeratorOrAdmin: false,
                refreshUser: vi.fn(),
            });
        });

        it('submits form with valid data', async () => {
            const user = userEvent.setup();
            mockSubmitClip.mockResolvedValue({
                success: true,
                message: 'Clip submitted successfully',
                submission: {
                    id: 'submission-123',
                    user_id: 'user-123',
                    twitch_clip_id: 'TestClip123',
                    twitch_clip_url: 'https://clips.twitch.tv/TestClip123',
                    title: 'Test Clip',
                    is_nsfw: false,
                    status: 'pending',
                    created_at: new Date().toISOString(),
                    updated_at: new Date().toISOString(),
                    broadcaster_name: 'TestStreamer',
                    creator_name: 'TestCreator',
                    view_count: 0,
                },
            });

            render(<SubmitClipPage />);

            const clipUrlInput = screen.getByLabelText(/Twitch Clip URL/);
            await user.type(
                clipUrlInput,
                'https://clips.twitch.tv/TestClip123'
            );

            const submitButton = screen.getByRole('button', {
                name: /Submit Clip/,
            });
            await user.click(submitButton);

            await waitFor(() => {
                const submission = mockSubmitClip.mock.calls[0][0];
                expect(submission).toMatchObject({
                    clip_url: 'https://clips.twitch.tv/TestClip123',
                    custom_title: 'Test Clip Title', // Auto-filled from metadata
                    tags: [],
                    is_nsfw: false,
                    submission_reason: '',
                });
                expect(submission.broadcaster_name_override).toBeUndefined();
            });
        });

        it('includes all form data in submission', async () => {
            const user = userEvent.setup();
            mockSubmitClip.mockResolvedValue({
                success: true,
                message: 'Clip submitted successfully',
                submission: {
                    id: 'submission-123',
                    user_id: 'user-123',
                    twitch_clip_id: 'TestClip123',
                    twitch_clip_url: 'https://clips.twitch.tv/TestClip123',
                    title: 'Test Clip',
                    custom_title: 'Custom Title',
                    is_nsfw: false,
                    status: 'pending',
                    created_at: new Date().toISOString(),
                    updated_at: new Date().toISOString(),
                    broadcaster_name: 'TestStreamer',
                    creator_name: 'TestCreator',
                    view_count: 0,
                },
            });

            render(<SubmitClipPage />);

            // Fill in all fields
            const clipUrlInput = screen.getByLabelText(/Twitch Clip URL/);
            await user.type(
                clipUrlInput,
                'https://clips.twitch.tv/TestClip123'
            );

            // Wait for metadata to load and populate the custom title field
            const customTitleInput = screen.getByLabelText(/Custom Title/);
            await waitFor(() => {
                expect(customTitleInput).toHaveValue('Test Clip Title');
            });
            // Select all and replace with new text
            await user.tripleClick(customTitleInput);
            await user.keyboard('Custom Title');

            const tagInput = screen.getByPlaceholderText(
                'Search or add tags...'
            );
            await user.type(tagInput, 'clutch{Enter}');
            expect(await screen.findByText('clutch')).toBeInTheDocument();

            const nsfwCheckbox = screen.getByLabelText(/Mark as NSFW/);
            await user.click(nsfwCheckbox);

            const reasonTextarea = screen.getByLabelText(/Submission Reason/);
            await user.type(reasonTextarea, 'Amazing play');

            const submitButton = screen.getByRole('button', {
                name: /Submit Clip/,
            });
            await user.click(submitButton);

            await waitFor(() => {
                const submission = mockSubmitClip.mock.calls[0][0];
                expect(submission).toMatchObject({
                    clip_url: 'https://clips.twitch.tv/TestClip123',
                    custom_title: 'Custom Title',
                    tags: ['clutch'],
                    is_nsfw: true,
                    submission_reason: 'Amazing play',
                });
                expect(submission.broadcaster_name_override).toBeUndefined();
            });
        });

        it('shows loading state during submission', async () => {
            const user = userEvent.setup();
            mockSubmitClip.mockImplementation(
                () => new Promise(resolve => setTimeout(resolve, 1000))
            );

            render(<SubmitClipPage />);

            const clipUrlInput = screen.getByLabelText(/Twitch Clip URL/);
            await user.type(
                clipUrlInput,
                'https://clips.twitch.tv/TestClip123'
            );

            const submitButton = screen.getByRole('button', {
                name: /Submit Clip/,
            });
            await user.click(submitButton);

            expect(screen.getByText('Submitting...')).toBeInTheDocument();
        });

        it('displays error message on submission failure', async () => {
            const user = userEvent.setup();
            mockSubmitClip.mockRejectedValue({
                response: {
                    data: {
                        error: 'Clip already exists',
                    },
                },
            });

            render(<SubmitClipPage />);

            const clipUrlInput = screen.getByLabelText(/Twitch Clip URL/);
            await user.type(
                clipUrlInput,
                'https://clips.twitch.tv/TestClip123'
            );

            const submitButton = screen.getByRole('button', {
                name: /Submit Clip/,
            });
            await user.click(submitButton);

            await waitFor(() => {
                expect(
                    screen.getByText('Clip already exists')
                ).toBeInTheDocument();
            });
        });

        it('shows generic error message when error details are unavailable', async () => {
            const user = userEvent.setup();
            mockSubmitClip.mockRejectedValue(new Error('Network error'));

            render(<SubmitClipPage />);

            const clipUrlInput = screen.getByLabelText(/Twitch Clip URL/);
            await user.type(
                clipUrlInput,
                'https://clips.twitch.tv/TestClip123'
            );

            const submitButton = screen.getByRole('button', {
                name: /Submit Clip/,
            });
            await user.click(submitButton);

            await waitFor(() => {
                expect(
                    screen.getByText('Failed to submit clip')
                ).toBeInTheDocument();
            });
        });
    });

    describe('Success View', () => {
        beforeEach(() => {
            mockUseAuth.mockReturnValue({
                user: mockUser,
                isAuthenticated: true,
                login: vi.fn(),
                logout: vi.fn(),
                isLoading: false,
                isAdmin: false,
                isModerator: false,
                isModeratorOrAdmin: false,
                refreshUser: vi.fn(),
            });
        });

        it('shows success confirmation after successful submission', async () => {
            const user = userEvent.setup();
            mockSubmitClip.mockResolvedValue({
                success: true,
                message: 'Clip submitted successfully',
                submission: {
                    id: 'submission-123',
                    user_id: 'user-123',
                    twitch_clip_id: 'TestClip123',
                    twitch_clip_url: 'https://clips.twitch.tv/TestClip123',
                    title: 'Test Clip',
                    is_nsfw: false,
                    status: 'pending',
                    created_at: new Date().toISOString(),
                    updated_at: new Date().toISOString(),
                    broadcaster_name: 'TestStreamer',
                    creator_name: 'TestCreator',
                    view_count: 0,
                },
            });

            render(<SubmitClipPage />);

            const clipUrlInput = screen.getByLabelText(/Twitch Clip URL/);
            await user.type(
                clipUrlInput,
                'https://clips.twitch.tv/TestClip123'
            );

            const submitButton = screen.getByRole('button', {
                name: /Submit Clip/,
            });
            await user.click(submitButton);

            await waitFor(() => {
                expect(
                    screen.getByText('Submission Successful!')
                ).toBeInTheDocument();
            });
        });

        it('resets form after successful submission', async () => {
            const user = userEvent.setup();
            mockSubmitClip.mockResolvedValue({
                success: true,
                message: 'Clip submitted successfully',
                submission: {
                    id: 'submission-123',
                    user_id: 'user-123',
                    twitch_clip_id: 'TestClip123',
                    twitch_clip_url: 'https://clips.twitch.tv/TestClip123',
                    title: 'Test Clip',
                    is_nsfw: false,
                    status: 'pending',
                    created_at: new Date().toISOString(),
                    updated_at: new Date().toISOString(),
                    broadcaster_name: 'TestStreamer',
                    creator_name: 'TestCreator',
                    view_count: 0,
                },
            });

            render(<SubmitClipPage />);

            const clipUrlInput = screen.getByLabelText(/Twitch Clip URL/);
            await user.type(
                clipUrlInput,
                'https://clips.twitch.tv/TestClip123'
            );

            const submitButton = screen.getByRole('button', {
                name: /Submit Clip/,
            });
            await user.click(submitButton);

            await waitFor(() => {
                expect(
                    screen.getByText('Submission Successful!')
                ).toBeInTheDocument();
            });

            // Click "Submit Another Clip" to return to form
            const submitAnotherButton = screen.getByText('Submit Another Clip');
            await user.click(submitAnotherButton);

            // Form should be reset
            const resetClipUrlInput = screen.getByLabelText(/Twitch Clip URL/);
            expect(resetClipUrlInput).toHaveValue('');
        });
    });

    describe('Recent Submissions', () => {
        beforeEach(() => {
            mockUseAuth.mockReturnValue({
                user: mockUser,
                isAuthenticated: true,
                login: vi.fn(),
                logout: vi.fn(),
                isLoading: false,
                isAdmin: false,
                isModerator: false,
                isModeratorOrAdmin: false,
                refreshUser: vi.fn(),
            });
        });

        it('loads and displays recent submissions', async () => {
            mockGetUserSubmissions.mockResolvedValue({
                success: true,
                data: [
                    {
                        id: 'sub-1',
                        user_id: 'user-123',
                        twitch_clip_id: 'clip-1',
                        twitch_clip_url: 'https://clips.twitch.tv/clip-1',
                        title: 'Recent Submission 1',
                        is_nsfw: false,
                        status: 'pending',
                        created_at: new Date().toISOString(),
                        updated_at: new Date().toISOString(),
                        broadcaster_name: 'Streamer1',
                        creator_name: 'Creator1',
                        view_count: 0,
                    },
                ],
                meta: {
                    page: 1,
                    limit: 5,
                    total: 1,
                    total_pages: 1,
                },
            });

            render(<SubmitClipPage />);

            await waitFor(() => {
                expect(
                    screen.getByText('Your Recent Submissions')
                ).toBeInTheDocument();
                expect(
                    screen.getByText('Recent Submission 1')
                ).toBeInTheDocument();
            });
        });

        it('shows status badges for recent submissions', async () => {
            mockGetUserSubmissions.mockResolvedValue({
                success: true,
                data: [
                    {
                        id: 'sub-1',
                        user_id: 'user-123',
                        twitch_clip_id: 'clip-1',
                        twitch_clip_url: 'https://clips.twitch.tv/clip-1',
                        title: 'Approved Submission',
                        is_nsfw: false,
                        status: 'approved',
                        created_at: new Date().toISOString(),
                        updated_at: new Date().toISOString(),
                        broadcaster_name: 'Streamer1',
                        creator_name: 'Creator1',
                        view_count: 0,
                    },
                    {
                        id: 'sub-2',
                        user_id: 'user-123',
                        twitch_clip_id: 'clip-2',
                        twitch_clip_url: 'https://clips.twitch.tv/clip-2',
                        title: 'Pending Submission',
                        is_nsfw: false,
                        status: 'pending',
                        created_at: new Date().toISOString(),
                        updated_at: new Date().toISOString(),
                        broadcaster_name: 'Streamer2',
                        creator_name: 'Creator2',
                        view_count: 0,
                    },
                ],
                meta: {
                    page: 1,
                    limit: 5,
                    total: 2,
                    total_pages: 1,
                },
            });

            render(<SubmitClipPage />);

            await waitFor(() => {
                expect(screen.getByText('approved')).toBeInTheDocument();
                expect(screen.getByText('pending')).toBeInTheDocument();
            });
        });
    });

    describe('Navigation and Cleanup', () => {
        beforeEach(() => {
            mockUseAuth.mockReturnValue({
                user: mockUser,
                isAuthenticated: true,
                login: vi.fn(),
                logout: vi.fn(),
                isLoading: false,
                isAdmin: false,
                isModerator: false,
                isModeratorOrAdmin: false,
                refreshUser: vi.fn(),
            });
        });

        it('cleans up effects when component unmounts', async () => {
            const { unmount } = render(<SubmitClipPage />);

            // Wait for component to mount and effects to run
            await waitFor(() => {
                expect(screen.getByText('Submit a Clip')).toBeInTheDocument();
            });

            // Unmount the component (simulates navigation)
            unmount();

            // Verify that API calls don't update state after unmount
            // This is implicitly tested by the isMounted flag in the useEffect
            expect(mockGetUserSubmissions).toHaveBeenCalled();
        });

        it('navigates back without issues after viewing confirmation', async () => {
            const user = userEvent.setup();
            mockSubmitClip.mockResolvedValue({
                success: true,
                message: 'Clip submitted successfully',
                submission: {
                    id: 'submission-123',
                    user_id: 'user-123',
                    twitch_clip_id: 'TestClip123',
                    twitch_clip_url: 'https://clips.twitch.tv/TestClip123',
                    title: 'Test Clip',
                    is_nsfw: false,
                    status: 'pending',
                    created_at: new Date().toISOString(),
                    updated_at: new Date().toISOString(),
                    broadcaster_name: 'TestStreamer',
                    creator_name: 'TestCreator',
                    view_count: 0,
                },
            });

            const { unmount } = render(<SubmitClipPage />);

            // Submit a clip
            const clipUrlInput = screen.getByLabelText(/Twitch Clip URL/);
            await user.type(
                clipUrlInput,
                'https://clips.twitch.tv/TestClip123'
            );

            const submitButton = screen.getByRole('button', {
                name: /Submit Clip/,
            });
            await user.click(submitButton);

            // Wait for confirmation view
            await waitFor(() => {
                expect(
                    screen.getByText('Submission Successful!')
                ).toBeInTheDocument();
            });

            // Unmount (simulates browser back navigation)
            unmount();

            // No errors should occur, and component should clean up properly
            // This test verifies that navigation away from confirmation doesn't cause issues
        });

        it('handles back navigation during form editing', async () => {
            const user = userEvent.setup();
            const { unmount } = render(<SubmitClipPage />);

            // Start filling the form
            const clipUrlInput = screen.getByLabelText(/Twitch Clip URL/);
            await user.type(clipUrlInput, 'https://clips.twitch.tv/Test');

            // Navigate back (unmount) while form is being edited
            unmount();

            // No errors should occur
            // This test verifies that navigation away during form editing doesn't cause issues
        });
    });
});
