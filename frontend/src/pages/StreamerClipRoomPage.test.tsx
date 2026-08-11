import { fireEvent, render, screen } from '@testing-library/react';
import type { ReactNode } from 'react';
import { MemoryRouter, Route, Routes } from 'react-router-dom';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { StreamerClipRoomPage } from './StreamerClipRoomPage';
import type {
    StreamerClipRoom,
    StreamerClipRoomItem,
    StreamerClipRoomWithItems,
} from '@/types/streamerClipRoom';

const mockUseStreamerClipRoom = vi.fn();
const mockUseStartStreamerClipRoom = vi.fn();
const mockUseStopStreamerClipRoom = vi.fn();
const mockUseApproveStreamerClipRoomItem = vi.fn();
const mockUseRejectStreamerClipRoomItem = vi.fn();
const mockUseReorderStreamerClipRoomItems = vi.fn();
const mockUseStreamerClipRoomWebSocket = vi.fn();
const mockUseTwitchBotAuthorization = vi.fn();
const mockUseUpdateStreamerClipRoomSubmissions = vi.fn();

vi.mock('@/hooks/useStreamerClipRoom', () => ({
    useStreamerClipRoom: (...args: unknown[]) => mockUseStreamerClipRoom(...args),
    useStartStreamerClipRoom: (...args: unknown[]) =>
        mockUseStartStreamerClipRoom(...args),
    useStopStreamerClipRoom: (...args: unknown[]) =>
        mockUseStopStreamerClipRoom(...args),
    useApproveStreamerClipRoomItem: (...args: unknown[]) =>
        mockUseApproveStreamerClipRoomItem(...args),
    useRejectStreamerClipRoomItem: (...args: unknown[]) =>
        mockUseRejectStreamerClipRoomItem(...args),
    useReorderStreamerClipRoomItems: (...args: unknown[]) =>
        mockUseReorderStreamerClipRoomItems(...args),
    useTwitchBotAuthorization: () => mockUseTwitchBotAuthorization(),
    useUpdateStreamerClipRoomSubmissions: () =>
        mockUseUpdateStreamerClipRoomSubmissions(),
}));

vi.mock('@/hooks/useStreamerClipRoomWebSocket', () => ({
    useStreamerClipRoomWebSocket: (...args: unknown[]) =>
        mockUseStreamerClipRoomWebSocket(...args),
}));

vi.mock('@/components/playlist/PlaylistTheatreMode', () => ({
    PlaylistTheatreMode: ({
        title,
        items,
        currentItemId,
        pendingLabel,
        commentsLabel,
        extraSidebarTab,
        onClose,
    }: {
        title: string;
        items: Array<{ id: string; clip?: { title?: string } }>;
        currentItemId: string | null;
        pendingLabel?: string;
        commentsLabel?: string;
        extraSidebarTab?: { label: string; content: ReactNode };
        onClose?: () => void;
    }) => (
        <div data-testid='playlist-theatre-mode'>
            <h2>{title}</h2>
            <div data-testid='approved-items'>
                {items.map(item => item.clip?.title ?? item.id).join(', ')}
            </div>
            <div data-testid='current-item'>{currentItemId}</div>
            <div data-testid='sidebar-labels'>
                {pendingLabel} | {commentsLabel}
            </div>
            {extraSidebarTab && (
                <div data-testid='extra-sidebar'>
                    <h3>{extraSidebarTab.label}</h3>
                    {extraSidebarTab.content}
                </div>
            )}
            {onClose && (
                <button type='button' onClick={onClose}>
                    close
                </button>
            )}
        </div>
    ),
}));

vi.mock('@/components/SEO', () => ({
    SEO: () => null,
}));

const startMutate = vi.fn();
const stopMutate = vi.fn();
const approveMutate = vi.fn();
const rejectMutate = vi.fn();
const reorderMutate = vi.fn();
const updateSubmissionsMutate = vi.fn();

const baseRoom: StreamerClipRoom = {
    id: 'room-1',
    owner_user_id: 'user-1',
    twitch_channel: 'teststreamer',
    approval_mode: 'manual',
    is_active: true,
    submissions_open: false,
    listener_started_at: '2026-06-06T10:00:00Z',
    created_at: '2026-06-06T09:00:00Z',
    updated_at: '2026-06-06T10:05:00Z',
};

const pendingItem: StreamerClipRoomItem = {
    id: 'pending-1',
    room_id: 'room-1',
    source_url: 'https://clpr.tv/clip/pending-1',
    source_type: 'clpr',
    status: 'pending',
    detected_at: '2026-06-06T10:06:00Z',
    twitch_username: 'viewer1',
    message_text: 'please play this one',
    created_at: '2026-06-06T10:06:00Z',
    updated_at: '2026-06-06T10:06:00Z',
};

const approvedItem: StreamerClipRoomItem = {
    id: 'approved-1',
    room_id: 'room-1',
    clip_id: 'clip-1',
    clip: {
        id: 'clip-1',
        title: 'Approved clip title',
        broadcaster_name: 'TestStreamer',
        thumbnail_url: 'https://example.com/thumb.jpg',
        duration: 42,
        video_url: 'https://video.example.com/clip.m3u8',
    } as StreamerClipRoomItem['clip'],
    source_url: 'https://clpr.tv/clip/approved-1',
    source_type: 'clpr',
    status: 'approved',
    position: 1,
    detected_at: '2026-06-06T09:58:00Z',
    approved_at: '2026-06-06T10:02:00Z',
    created_at: '2026-06-06T09:58:00Z',
    updated_at: '2026-06-06T10:02:00Z',
};

describe('StreamerClipRoomPage', () => {
    beforeEach(() => {
        vi.clearAllMocks();

        mockUseStreamerClipRoom.mockReturnValue({
            data: {
                room: baseRoom,
                pending_items: [pendingItem],
                approved_items: [],
                skipped_items: [],
            } satisfies StreamerClipRoomWithItems,
            isLoading: false,
            isError: false,
            refetch: vi.fn(),
        });

        mockUseStreamerClipRoomWebSocket.mockReturnValue({
            isConnected: true,
            error: null,
            lastEvent: null,
        });

        mockUseStartStreamerClipRoom.mockReturnValue({
            mutate: startMutate,
            isPending: false,
        });
        mockUseStopStreamerClipRoom.mockReturnValue({
            mutate: stopMutate,
            isPending: false,
        });
        mockUseApproveStreamerClipRoomItem.mockReturnValue({
            mutate: approveMutate,
            isPending: false,
        });
        mockUseRejectStreamerClipRoomItem.mockReturnValue({
            mutate: rejectMutate,
            isPending: false,
        });
        mockUseReorderStreamerClipRoomItems.mockReturnValue({
            mutate: reorderMutate,
            isPending: false,
        });
        mockUseUpdateStreamerClipRoomSubmissions.mockReturnValue({
            mutate: updateSubmissionsMutate,
            isPending: false,
            isError: false,
            error: null,
        });
        mockUseTwitchBotAuthorization.mockReturnValue({
            data: {
                authenticated: true,
                bot_authorized: true,
                clip_download_authorized: true,
                twitch_username: 'teststreamer',
            },
            isLoading: false,
        });
    });

    it('renders listener controls and pending approvals when there are no approved clips', () => {
        render(
            <MemoryRouter initialEntries={['/streamer-tools/teststreamer/clips']}>
                <Routes>
                    <Route
                        path='/streamer-tools/:channel/clips'
                        element={<StreamerClipRoomPage />}
                    />
                </Routes>
            </MemoryRouter>,
        );

        expect(
            screen.getByRole('heading', { name: 'teststreamer' }),
        ).toBeInTheDocument();
        expect(screen.getAllByRole('button', { name: /start listener/i })).not.toHaveLength(0);
        expect(screen.getAllByRole('button', { name: /stop listener/i })).not.toHaveLength(0);
        expect(screen.getByText(/pending approvals/i)).toBeInTheDocument();
        expect(screen.getByRole('button', { name: /approve/i })).toBeInTheDocument();
        expect(screen.getByRole('button', { name: /reject/i })).toBeInTheDocument();
        expect(screen.getByText(/submissions closed/i)).toBeInTheDocument();
        expect(screen.getByRole('button', { name: /start timer/i })).toBeEnabled();
    });

    it('opens timed submissions and lets the streamer close them', () => {
        const timedRoom = {
            ...baseRoom,
            submissions_open: true,
            submissions_close_at: new Date(Date.now() + 10 * 60_000).toISOString(),
        } satisfies StreamerClipRoom;
        mockUseStreamerClipRoom.mockReturnValue({
            data: {
                room: timedRoom,
                pending_items: [],
                approved_items: [],
                skipped_items: [],
            } satisfies StreamerClipRoomWithItems,
            isLoading: false,
            isError: false,
            refetch: vi.fn(),
        });

        render(
            <MemoryRouter initialEntries={['/streamer-tools/teststreamer/clips']}>
                <Routes>
                    <Route
                        path='/streamer-tools/:channel/clips'
                        element={<StreamerClipRoomPage />}
                    />
                </Routes>
            </MemoryRouter>,
        );

        expect(screen.getByText(/remaining/i)).toBeInTheDocument();
        fireEvent.click(screen.getByRole('button', { name: /close submissions/i }));
        expect(updateSubmissionsMutate).toHaveBeenCalledWith({
            channel: 'teststreamer',
            enabled: false,
        });
    });

    it('starts submissions with the selected timer duration', () => {
        render(
            <MemoryRouter initialEntries={['/streamer-tools/teststreamer/clips']}>
                <Routes>
                    <Route
                        path='/streamer-tools/:channel/clips'
                        element={<StreamerClipRoomPage />}
                    />
                </Routes>
            </MemoryRouter>,
        );

        fireEvent.change(screen.getByLabelText(/submission duration/i), {
            target: { value: '25' },
        });
        fireEvent.click(screen.getByRole('button', { name: /start timer/i }));
        expect(updateSubmissionsMutate).toHaveBeenCalledWith({
            channel: 'teststreamer',
            enabled: true,
            durationMinutes: 25,
        });
    });

    it('passes approved clips into theatre mode and keeps the approval rail visible', () => {
        mockUseStreamerClipRoom.mockReturnValue({
            data: {
                room: baseRoom,
                pending_items: [pendingItem],
                approved_items: [approvedItem],
                skipped_items: [],
            } satisfies StreamerClipRoomWithItems,
            isLoading: false,
            isError: false,
            refetch: vi.fn(),
        });

        render(
            <MemoryRouter initialEntries={['/streamer-tools/teststreamer/clips']}>
                <Routes>
                    <Route
                        path='/streamer-tools/:channel/clips'
                        element={<StreamerClipRoomPage />}
                    />
                </Routes>
            </MemoryRouter>,
        );

        expect(screen.getByTestId('playlist-theatre-mode')).toBeInTheDocument();
        expect(screen.getByTestId('approved-items')).toHaveTextContent(
            'Approved clip title',
        );
        expect(screen.getByText(/approved clips/i)).toBeInTheDocument();
        expect(screen.getByText(/discussion/i)).toBeInTheDocument();
        expect(screen.getByText('Approvals')).toBeInTheDocument();
    });

    it('asks the streamer to authorize the Clpr bot before starting', () => {
        mockUseTwitchBotAuthorization.mockReturnValue({
            data: {
                authenticated: true,
                bot_authorized: false,
                clip_download_authorized: true,
                twitch_username: 'teststreamer',
            },
            isLoading: false,
        });

		render(
			<MemoryRouter initialEntries={['/streamer-tools/teststreamer/clips']}>
				<Routes>
					<Route
						path='/streamer-tools/:channel/clips'
						element={<StreamerClipRoomPage />}
					/>
				</Routes>
			</MemoryRouter>,
		);

        expect(
            screen.getByText(/let clpr watch your twitch chat/i),
        ).toBeInTheDocument();
        expect(
            screen.getAllByRole('button', { name: /authorize.*bot/i }).length,
        ).toBeGreaterThan(0);
		for (const startButton of screen.getAllByRole('button', { name: /start listener/i })) {
			expect(startButton).toBeDisabled();
		}
	});

    it('asks the streamer to authorize clip downloads before starting', () => {
        mockUseTwitchBotAuthorization.mockReturnValue({
            data: {
                authenticated: true,
                bot_authorized: true,
                clip_download_authorized: false,
                twitch_username: 'teststreamer',
            },
            isLoading: false,
        });

        render(
            <MemoryRouter initialEntries={['/streamer-tools/teststreamer/clips']}>
                <Routes>
                    <Route
                        path='/streamer-tools/:channel/clips'
                        element={<StreamerClipRoomPage />}
                    />
                </Routes>
            </MemoryRouter>,
        );

        expect(
            screen.getByText(/allow clpr to download your twitch clips/i),
        ).toBeInTheDocument();
        expect(
            screen.getAllByRole('button', { name: /authorize clip downloads/i }).length,
        ).toBeGreaterThan(0);
        for (const startButton of screen.getAllByRole('button', { name: /start listener/i })) {
            expect(startButton).toBeDisabled();
        }
    });
});
