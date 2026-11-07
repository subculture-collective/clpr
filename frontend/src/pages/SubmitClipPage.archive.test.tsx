import { createHash } from 'node:crypto';
import userEvent from '@testing-library/user-event';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { useAuth } from '../context/AuthContext';
import * as submissionApi from '../lib/submission-api';
import { render, screen, waitFor } from '../test/test-utils';
import fixture from '../test/fixtures/rekolekt-clip-handoff.v1.json';
import { SubmitClipPage } from './SubmitClipPage';

vi.mock('../lib/submission-api', () => ({
    submitClip: vi.fn(),
    submitClipUpload: vi.fn(),
    getUserSubmissions: vi.fn(),
    checkClipStatus: vi.fn(),
    getClipMetadata: vi.fn(),
}));

vi.mock('../lib/tag-api', () => ({
    tagApi: { searchTags: vi.fn().mockResolvedValue({ tags: [] }), createTag: vi.fn() },
}));

vi.mock('../context/AuthContext', async () => {
    const actual = await vi.importActual('../context/AuthContext');
    return { ...actual, useAuth: vi.fn() };
});

// The rendered MP4 the archive export describes, and a handoff naming it.
const mp4Bytes = new TextEncoder().encode('rendered passage bytes');
const handoffRaw = JSON.stringify({
    ...fixture,
    media: {
        ...fixture.media,
        bytes: mp4Bytes.length,
        sha256: createHash('sha256').update(mp4Bytes).digest('hex'),
    },
});
const handoffFile = () => new File([handoffRaw], 'passage.handoff.json', { type: 'application/json' });
const mp4File = () => new File([mp4Bytes], 'passage.mp4', { type: 'video/mp4' });

describe('SubmitClipPage archive handoff import', () => {
    const mockSubmitClipUpload = vi.mocked(submissionApi.submitClipUpload);

    beforeEach(() => {
        vi.clearAllMocks();
        localStorage.clear();
        vi.mocked(submissionApi.getUserSubmissions).mockResolvedValue({
            success: true,
            data: [],
            meta: { page: 1, limit: 5, total: 0, total_pages: 0 },
        });
        vi.mocked(useAuth).mockReturnValue({
            user: {
                id: 'user-123',
                twitch_id: 'twitch-123',
                username: 'testuser',
                display_name: 'Test User',
                karma_points: 0,
                role: 'user' as const,
                is_banned: false,
                created_at: '2024-01-01T00:00:00Z',
            },
            isAuthenticated: true,
            login: vi.fn(),
            logout: vi.fn(),
            isLoading: false,
            isAdmin: false,
            isModerator: false,
            isModeratorOrAdmin: false,
            refreshUser: vi.fn(),
        } as unknown as ReturnType<typeof useAuth>);
        mockSubmitClipUpload.mockResolvedValue({
            success: true,
            message: 'Upload submitted for review',
            submission: {
                id: 'submission-1',
                user_id: 'user-123',
                twitch_clip_id: 'upload:1',
                twitch_clip_url: '',
                is_nsfw: false,
                status: 'pending',
                view_count: 0,
                source_type: 'upload',
                source_platform: 'upload',
                created_at: new Date().toISOString(),
                updated_at: new Date().toISOString(),
            },
        });
    });

    async function openUpload(user: ReturnType<typeof userEvent.setup>) {
        render(<SubmitClipPage />);
        await user.click(screen.getByRole('radio', { name: /upload video/i }));
    }

    it('shows the handoff for review, and Cancel leaves the draft unchanged', async () => {
        const user = userEvent.setup();
        await openUpload(user);

        await user.upload(screen.getByLabelText(/Archive handoff/), handoffFile());

        const review = await screen.findByTestId('archive-handoff-review');
        expect(review).toHaveTextContent('Transcript Archive');
        expect(review).toHaveTextContent('12:34.0 – 12:43.0 of the recording');
        expect(review).toHaveTextContent('4 cues');
        expect(review).toHaveTextContent('not independent proof of rights');

        await user.click(screen.getByRole('button', { name: 'Cancel' }));

        expect(screen.queryByTestId('archive-handoff-review')).not.toBeInTheDocument();
        expect(screen.queryByTestId('archive-handoff-current')).not.toBeInTheDocument();
        expect(localStorage.getItem('submission_draft')).toBeNull();
        expect(mockSubmitClipUpload).not.toHaveBeenCalled();
    });

    it('reports an unsupported handoff without changing the draft', async () => {
        const user = userEvent.setup();
        await openUpload(user);

        const v2 = new File([JSON.stringify({ ...fixture, schema_version: 2 })], 'h.json', { type: 'application/json' });
        await user.upload(screen.getByLabelText(/Archive handoff/), v2);

        expect(await screen.findByText(/Unsupported archive handoff version/)).toBeInTheDocument();
        expect(screen.queryByTestId('archive-handoff-review')).not.toBeInTheDocument();
    });

    it('asks before replacing a draft that already has content', async () => {
        const user = userEvent.setup();
        await openUpload(user);
        await user.type(screen.getByLabelText(/Custom Title/), 'My earlier title');

        await user.upload(screen.getByLabelText(/Archive handoff/), handoffFile());

        expect(await screen.findByText(/Your draft already has content/)).toBeInTheDocument();
        await user.click(screen.getByRole('button', { name: 'Replace draft' }));

        expect(await screen.findByTestId('archive-handoff-current')).toBeInTheDocument();
        expect(screen.getByLabelText(/Custom Title/)).toHaveValue('');
        expect(screen.getByRole('link', { name: 'Open the source passage' })).toHaveAttribute('href', fixture.passage.url);
        // Importing saves the draft at once but submits nothing.
        const saved = JSON.parse(localStorage.getItem('submission_draft') ?? '{}');
        expect(saved.archiveImport.raw).toBe(handoffRaw);
        expect(mockSubmitClipUpload).not.toHaveBeenCalled();
    });

    it('submits the matching MP4 with the original handoff and the edited captions', async () => {
        const user = userEvent.setup();
        await openUpload(user);
        await user.upload(screen.getByLabelText(/Archive handoff/), handoffFile());
        await user.click(await screen.findByRole('button', { name: 'Use in draft' }));

        const firstCue = screen.getByLabelText('Caption 1 text');
        await user.clear(firstCue);
        await user.type(firstCue, 'and that is the part nobody reads,');
        await user.click(screen.getByRole('button', { name: 'Remove caption 3' }));
        expect(screen.getByRole('button', { name: "Restore the archive's captions" })).toBeInTheDocument();

        await user.upload(screen.getByLabelText(/Video File/i), mp4File());
        expect(await screen.findByText('Matches the archive handoff')).toBeInTheDocument();
        expect(mockSubmitClipUpload).not.toHaveBeenCalled();

        await user.click(screen.getByRole('button', { name: /Upload Clip/ }));

        await waitFor(() => expect(mockSubmitClipUpload).toHaveBeenCalledTimes(1));
        const [request] = mockSubmitClipUpload.mock.calls[0];
        expect(request.archive_handoff).toBe(handoffRaw);
        expect(request.caption_cues).toEqual([
            { start_ms: 0, end_ms: 1840, text: 'and that is the part nobody reads,' },
            { start_ms: 1840, end_ms: 5210, text: 'the footnote on page forty that changes the whole story.' },
            { start_ms: 6900, end_ms: 9000, text: "So let's actually read it." },
        ]);
    });

    it('blocks submission when the chosen MP4 does not match the handoff', async () => {
        const user = userEvent.setup();
        await openUpload(user);
        await user.upload(screen.getByLabelText(/Archive handoff/), handoffFile());
        await user.click(await screen.findByRole('button', { name: 'Use in draft' }));

        await user.upload(
            screen.getByLabelText(/Video File/i),
            new File([new TextEncoder().encode('a different rendering!')], 'other.mp4', { type: 'video/mp4' }),
        );

        expect(await screen.findByText('Does not match the handoff')).toBeInTheDocument();
        expect(screen.getByRole('alert', { name: '' })).toHaveTextContent('expects passage-');
        await user.click(screen.getByRole('button', { name: /Upload Clip/ }));
        expect(mockSubmitClipUpload).not.toHaveBeenCalled();
    });

    it('resumes an archive import and its caption edits from a saved draft', async () => {
        const handoff = JSON.parse(handoffRaw);
        const editedCues = handoff.captions.cues.map((cue: { text: string }, i: number) =>
            i === 0 ? { ...cue, text: 'edited before closing the tab' } : cue,
        );
        localStorage.setItem(
            'submission_draft',
            JSON.stringify({
                formData: { clip_url: '', source_type: 'upload', custom_title: 'Resumed', is_nsfw: false },
                selectedTags: [],
                archiveImport: { raw: handoffRaw, handoff, cues: editedCues },
                lastSaved: Date.now(),
            }),
        );

        render(<SubmitClipPage />);

        expect(await screen.findByText('Draft restored from your last session')).toBeInTheDocument();
        expect(screen.getByTestId('archive-handoff-current')).toHaveTextContent('Transcript Archive');
        expect(screen.getByLabelText('Caption 1 text')).toHaveValue('edited before closing the tab');
        expect(screen.getByLabelText(/Custom Title/)).toHaveValue('Resumed');
    });
});
