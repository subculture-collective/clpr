import { useCallback, useEffect, useMemo, useState } from 'react';
import type { ReactNode } from 'react';
import { useNavigate, useParams } from 'react-router-dom';
import {
    AlertTriangle,
    Check,
    CircleAlert,
    Clock3,
    Radio,
    RefreshCw,
    Square,
    X,
} from 'lucide-react';
import { SEO } from '@/components/SEO';
import { PlaylistTheatreMode, type PlaylistItem } from '@/components/playlist/PlaylistTheatreMode';
import { Button, Spinner } from '@/components/ui';
import {
    useApproveStreamerClipRoomItem,
    useRejectStreamerClipRoomItem,
    useReorderStreamerClipRoomItems,
    useStartStreamerClipRoom,
    useStopStreamerClipRoom,
    useStreamerClipRoom,
    useTwitchBotAuthorization,
    useUpdateStreamerClipRoomSubmissions,
} from '@/hooks/useStreamerClipRoom';
import { useStreamerClipRoomWebSocket } from '@/hooks/useStreamerClipRoomWebSocket';
import { getTwitchBotAuthorizationUrl } from '@/lib/twitch-api';
import type { StreamerClipRoomItem } from '@/types/streamerClipRoom';

const dateTimeFormatter = new Intl.DateTimeFormat(undefined, {
    dateStyle: 'medium',
    timeStyle: 'short',
});

function formatDateTime(value?: string) {
    if (!value) return 'Unknown';

    return dateTimeFormatter.format(new Date(value));
}

function formatRemaining(milliseconds: number) {
    const totalSeconds = Math.max(0, Math.ceil(milliseconds / 1000));
    const hours = Math.floor(totalSeconds / 3600);
    const minutes = Math.floor((totalSeconds % 3600) / 60);
    const seconds = totalSeconds % 60;

    return hours > 0
        ? `${hours}:${String(minutes).padStart(2, '0')}:${String(seconds).padStart(2, '0')}`
        : `${minutes}:${String(seconds).padStart(2, '0')}`;
}

function ItemBadge({
    tone,
    children,
}: {
    tone: 'brand' | 'neutral' | 'success' | 'warning' | 'danger';
    children: ReactNode;
}) {
    const toneClass =
        tone === 'brand' ? 'border-brand/30 bg-brand/10 text-brand-200'
        : tone === 'success' ? 'border-emerald-400/30 bg-emerald-500/10 text-emerald-200'
        : tone === 'warning' ? 'border-amber-400/30 bg-amber-500/10 text-amber-200'
        : tone === 'danger' ? 'border-rose-400/30 bg-rose-500/10 text-rose-200'
        : 'border-white/10 bg-white/5 text-white/70';

    return (
        <span
            className={`inline-flex items-center rounded-full border px-2 py-0.5 text-[11px] font-medium ${toneClass}`}
        >
            {children}
        </span>
    );
}

function PendingApprovalCard({
    item,
    onApprove,
    onReject,
}: {
    item: StreamerClipRoomItem;
    onApprove: (itemId: string) => void;
    onReject: (itemId: string) => void;
}) {
    const clipTitle = item.clip?.title ?? 'Awaiting clip import';
    const sourceLabel = item.source_type === 'twitch' ? 'Twitch link' : 'Clpr link';

    return (
        <article className='rounded-2xl border border-white/10 bg-white/[0.04] p-4 shadow-[0_0_0_1px_rgba(255,255,255,0.01)]'>
            <div className='flex items-start justify-between gap-3'>
                <div className='min-w-0 space-y-2'>
                    <div className='flex flex-wrap items-center gap-2'>
                        <ItemBadge tone={item.source_type === 'twitch' ? 'warning' : 'brand'}>
                            {sourceLabel}
                        </ItemBadge>
                        {item.twitch_username && (
                            <ItemBadge tone='neutral'>@{item.twitch_username}</ItemBadge>
                        )}
                    </div>

                    <h3 className='line-clamp-2 text-sm font-semibold text-white'>
                        {clipTitle}
                    </h3>

                    {item.message_text && (
                        <p className='line-clamp-3 text-sm text-white/70'>
                            {item.message_text}
                        </p>
                    )}
                </div>
            </div>

            <dl className='mt-3 grid gap-2 text-xs text-white/60 sm:grid-cols-2'>
                <div>
                    <dt className='uppercase tracking-[0.2em] text-white/35'>Source</dt>
                    <dd className='mt-1 break-all text-white/85'>{item.source_url}</dd>
                </div>
                <div>
                    <dt className='uppercase tracking-[0.2em] text-white/35'>Detected</dt>
                    <dd className='mt-1 text-white/85'>{formatDateTime(item.detected_at)}</dd>
                </div>
                {item.skip_reason && (
                    <div className='sm:col-span-2'>
                        <dt className='uppercase tracking-[0.2em] text-white/35'>Skip reason</dt>
                        <dd className='mt-1 text-amber-200'>{item.skip_reason}</dd>
                    </div>
                )}
            </dl>

            <div className='mt-4 flex flex-wrap gap-2'>
                <Button
                    size='sm'
                    onClick={() => onApprove(item.id)}
                    leftIcon={<Check className='h-4 w-4' />}
                >
                    Approve
                </Button>
                <Button
                    size='sm'
                    variant='outline'
                    onClick={() => onReject(item.id)}
                    leftIcon={<X className='h-4 w-4' />}
                >
                    Reject
                </Button>
            </div>
        </article>
    );
}

function PendingApprovalsRail({
    pendingItems,
    skippedItems,
    onApprove,
    onReject,
}: {
    pendingItems: StreamerClipRoomItem[];
    skippedItems: StreamerClipRoomItem[];
    onApprove: (itemId: string) => void;
    onReject: (itemId: string) => void;
}) {
    return (
        <div className='space-y-5 p-4'>
            <section className='space-y-3'>
                <div className='flex items-center justify-between gap-3'>
                    <div>
                        <p className='text-sm font-semibold text-white'>Pending approvals</p>
                        <p className='text-xs text-white/55'>
                            Clips caught from Twitch chat before they reach playback.
                        </p>
                    </div>
                    <ItemBadge tone='brand'>{pendingItems.length}</ItemBadge>
                </div>

                {pendingItems.length > 0 ? (
                    <div className='space-y-3'>
                        {pendingItems.map(item => (
                            <PendingApprovalCard
                                key={item.id}
                                item={item}
                                onApprove={onApprove}
                                onReject={onReject}
                            />
                        ))}
                    </div>
                ) : (
                    <div className='rounded-2xl border border-dashed border-white/10 bg-white/[0.03] p-5 text-sm text-white/60'>
                        No clips waiting for approval.
                    </div>
                )}
            </section>

            {skippedItems.length > 0 && (
                <section className='space-y-3'>
                    <div className='flex items-center justify-between gap-3'>
                        <p className='text-sm font-semibold text-white'>Skipped by listener</p>
                        <ItemBadge tone='warning'>{skippedItems.length}</ItemBadge>
                    </div>

                    <div className='space-y-2'>
                        {skippedItems.map(item => (
                            <div
                                key={item.id}
                                className='rounded-xl border border-amber-400/20 bg-amber-500/5 p-3 text-xs text-white/70'
                            >
                                <div className='flex flex-wrap items-center gap-2'>
                                    {item.twitch_username && (
                                        <ItemBadge tone='neutral'>@{item.twitch_username}</ItemBadge>
                                    )}
                                    <span className='text-white/55'>{formatDateTime(item.detected_at)}</span>
                                </div>
                                <p className='mt-2 line-clamp-2 text-white/85'>
                                    {item.source_url}
                                </p>
                                {item.skip_reason && (
                                    <p className='mt-1 text-amber-200'>
                                        {item.skip_reason}
                                    </p>
                                )}
                            </div>
                        ))}
                    </div>
                </section>
            )}
        </div>
    );
}

export function StreamerClipRoomPage() {
    const { channel = '' } = useParams<{ channel: string }>();
    const navigate = useNavigate();

    const {
        data,
        isLoading,
        isError,
        refetch,
    } = useStreamerClipRoom(channel);

    const room = data?.room;
    const roomId = room?.id;
    const approvedItems = useMemo(
        () => data?.approved_items ?? [],
        [data?.approved_items],
    );
    const pendingItems = useMemo(
        () => data?.pending_items ?? [],
        [data?.pending_items],
    );
    const skippedItems = useMemo(
        () => data?.skipped_items ?? [],
        [data?.skipped_items],
    );

    const startRoom = useStartStreamerClipRoom();
    const stopRoom = useStopStreamerClipRoom();
    const updateSubmissions = useUpdateStreamerClipRoomSubmissions();
    const approveItem = useApproveStreamerClipRoomItem();
    const rejectItem = useRejectStreamerClipRoomItem();
    const reorderItems = useReorderStreamerClipRoomItems();
    const botAuthorization = useTwitchBotAuthorization();
    const isBotAuthorized = botAuthorization.data?.bot_authorized ?? false;
    const isClipDownloadAuthorized =
        botAuthorization.data?.clip_download_authorized ?? false;
    const hasRequiredTwitchAuthorization =
        isBotAuthorized && isClipDownloadAuthorized;
    const [submissionDurationMinutes, setSubmissionDurationMinutes] = useState(10);
    const [nowMs, setNowMs] = useState(() => Date.now());
    const submissionsCloseAtMs = room?.submissions_close_at
        ? new Date(room.submissions_close_at).getTime()
        : null;

    useEffect(() => {
        if (!room?.submissions_open || submissionsCloseAtMs === null) return;

        const intervalId = window.setInterval(() => setNowMs(Date.now()), 1000);
        return () => window.clearInterval(intervalId);
    }, [room?.submissions_open, submissionsCloseAtMs]);

    const { isConnected, error: wsError } = useStreamerClipRoomWebSocket({
        roomId,
        channel,
        enabled: !!roomId,
    });

    const theatreItems: PlaylistItem[] = useMemo(
        () =>
            [...approvedItems]
                .sort((a, b) => {
                    const left = a.position ?? Number.MAX_SAFE_INTEGER;
                    const right = b.position ?? Number.MAX_SAFE_INTEGER;
                    if (left !== right) return left - right;
                    return a.detected_at.localeCompare(b.detected_at);
                })
                .map(item => ({
                    id: item.id,
                    clip: item.clip,
                    clip_id: item.clip?.id ?? item.clip_id ?? item.id,
                    order: item.position ?? undefined,
                })),
        [approvedItems],
    );

    const [selectedItemId, setCurrentItemId] = useState<string | null>(null);
    const currentItemId =
        selectedItemId &&
        theatreItems.some(item => item.id === selectedItemId)
            ? selectedItemId
            : (theatreItems[0]?.id ?? null);

    const currentItem = useMemo(
        () => theatreItems.find(item => item.id === currentItemId) ?? null,
        [currentItemId, theatreItems],
    );

    const handleItemClick = useCallback((item: PlaylistItem) => {
        setCurrentItemId(item.id);
    }, []);

    const handleReorder = useCallback(
        (itemId: string, newPosition: number) => {
            if (!roomId) return;

            const nextOrder = [...theatreItems];
            const draggedIndex = nextOrder.findIndex(item => item.id === itemId);

            if (draggedIndex === -1) return;

            const [draggedItem] = nextOrder.splice(draggedIndex, 1);
            nextOrder.splice(newPosition, 0, draggedItem);

            reorderItems.mutate({
                roomId,
                itemIds: nextOrder.map(item => item.id),
            });
        },
        [reorderItems, roomId, theatreItems],
    );

    const handleApprove = useCallback(
        (itemId: string) => {
            if (!roomId) return;

            approveItem.mutate({ roomId, itemId });
        },
        [approveItem, roomId],
    );

    const handleReject = useCallback(
        (itemId: string) => {
            if (!roomId) return;

            rejectItem.mutate({ roomId, itemId });
        },
        [rejectItem, roomId],
    );

    const handleStartListener = useCallback(() => {
        if (!channel) return;

        startRoom.mutate(channel);
    }, [channel, startRoom]);

    const handleAuthorizeBot = useCallback(() => {
        if (!channel) return;
        window.location.assign(getTwitchBotAuthorizationUrl(channel));
    }, [channel]);

    const handleStopListener = useCallback(() => {
        if (!channel) return;

        stopRoom.mutate(channel);
    }, [channel, stopRoom]);

    const handleOpenSubmissions = useCallback((durationMinutes?: number) => {
        if (!channel) return;
        updateSubmissions.mutate({ channel, enabled: true, durationMinutes });
    }, [channel, updateSubmissions]);

    const handleCloseSubmissions = useCallback(() => {
        if (!channel) return;
        updateSubmissions.mutate({ channel, enabled: false });
    }, [channel, updateSubmissions]);

    const approvalContent = useMemo(
        () => (
            <PendingApprovalsRail
                pendingItems={pendingItems}
                skippedItems={skippedItems}
                onApprove={handleApprove}
                onReject={handleReject}
            />
        ),
        [handleApprove, handleReject, pendingItems, skippedItems],
    );

    const hasApprovedItems = theatreItems.length > 0;
    const isListenerLive = room?.is_active ?? false;
    const areSubmissionsOpen = Boolean(
        room?.submissions_open &&
        (submissionsCloseAtMs === null || submissionsCloseAtMs > nowMs),
    );
    const submissionCountdown =
        areSubmissionsOpen && submissionsCloseAtMs !== null
            ? formatRemaining(submissionsCloseAtMs - nowMs)
            : null;
    const activeConnectionLabel =
        wsError ? 'Sync issue'
        : isConnected ? 'Live sync'
        : 'Connecting';

    if (isLoading) {
        return (
            <>
                <SEO title={`${channel} Clip Room`} />
                <div className='fixed inset-0 flex items-center justify-center bg-black'>
                    <Spinner size='lg' />
                </div>
            </>
        );
    }

    if (isError || !data) {
        return (
            <>
                <SEO title={`${channel} Clip Room`} />
                <div className='fixed inset-0 flex items-center justify-center bg-black px-4 text-white'>
                    <div className='max-w-lg rounded-3xl border border-white/10 bg-white/[0.04] p-6 text-center shadow-2xl'>
                        <div className='mx-auto mb-4 flex h-12 w-12 items-center justify-center rounded-2xl bg-rose-500/10 text-rose-200'>
                            <AlertTriangle className='h-6 w-6' />
                        </div>
                        <h1 className='text-2xl font-semibold'>Failed to load clip room</h1>
                        <p className='mt-2 text-sm text-white/65'>
                            We couldn’t load the streamer clip room for{' '}
                            <span className='font-medium text-white'>
                                {channel}
                            </span>
                            .
                        </p>
                        <div className='mt-6 flex flex-wrap justify-center gap-2'>
                            <Button onClick={() => refetch()} leftIcon={<RefreshCw className='h-4 w-4' />}>Retry</Button>
                            <Button
                                variant='outline'
                                onClick={() => navigate('/playlists')}
                            >
                                Back to playlists
                            </Button>
                        </div>
                    </div>
                </div>
            </>
        );
    }

    return (
        <>
            <SEO
                title={`${channel} Clip Room`}
                description={`Moderate Twitch clip submissions for ${channel}`}
            />

            <div className='min-h-screen bg-black text-white'>
                <div className='fixed inset-x-0 top-0 z-[60] border-b border-white/10 bg-black/85 backdrop-blur-xl'>
                    <div className='mx-auto flex w-full max-w-[1600px] flex-col gap-3 px-4 py-3 lg:px-6'>
                        <div className='flex flex-wrap items-center gap-3'>
                            <div className='flex min-w-0 items-center gap-3'>
                                <div
                                    className={`flex h-10 w-10 items-center justify-center rounded-2xl ${
                                        isListenerLive ? 'bg-emerald-500/15 text-emerald-200' : 'bg-white/8 text-white/55'
                                    }`}
                                >
                                    <Radio className='h-5 w-5' />
                                </div>
                                <div className='min-w-0'>
                                    <p className='text-[11px] uppercase tracking-[0.26em] text-white/40'>
                                        Streamer clip room
                                    </p>
                                    <h1 className='truncate text-lg font-semibold text-white'>
                                        {channel}
                                    </h1>
                                </div>
                            </div>

                            <div className='flex flex-wrap items-center gap-2'>
                                <ItemBadge tone={isListenerLive ? 'success' : 'neutral'}>
                                    {isListenerLive ? 'Listener active' : 'Listener stopped'}
                                </ItemBadge>
                                <ItemBadge tone={areSubmissionsOpen ? 'success' : 'neutral'}>
                                    {areSubmissionsOpen
                                        ? `Submissions open${submissionCountdown ? ` · ${submissionCountdown}` : ''}`
                                        : 'Submissions closed'}
                                </ItemBadge>
                                <ItemBadge tone={isConnected ? 'success' : 'warning'}>
                                    {activeConnectionLabel}
                                </ItemBadge>
                                <ItemBadge tone='brand'>
                                    Approved {theatreItems.length}
                                </ItemBadge>
                                <ItemBadge
                                    tone={isBotAuthorized ? 'success' : 'warning'}
                                >
                                    {isBotAuthorized
                                        ? 'Bot authorized'
                                        : 'Bot needs access'}
                                </ItemBadge>
                                <ItemBadge
                                    tone={isClipDownloadAuthorized ? 'success' : 'warning'}
                                >
                                    {isClipDownloadAuthorized
                                        ? 'Clip downloads authorized'
                                        : 'Clip downloads need access'}
                                </ItemBadge>
                                <ItemBadge tone='neutral'>
                                    Pending {pendingItems.length}
                                </ItemBadge>
                            </div>

                            <div className='ml-auto flex flex-wrap items-center gap-2'>
                                {!hasRequiredTwitchAuthorization && (
                                    <Button
                                        onClick={handleAuthorizeBot}
                                        loading={botAuthorization.isLoading}
                                        leftIcon={<Radio className='h-4 w-4' />}
                                    >
                                        {!isBotAuthorized
                                            ? 'Authorize Clpr bot'
                                            : 'Authorize clip downloads'}
                                    </Button>
                                )}
                                <Button
                                    onClick={handleStartListener}
                                    loading={startRoom.isPending}
                                    disabled={!hasRequiredTwitchAuthorization}
                                    leftIcon={<Radio className='h-4 w-4' />}
                                >
                                    Start listener
                                </Button>
                                <Button
                                    variant='outline'
                                    onClick={handleStopListener}
                                    loading={stopRoom.isPending}
                                    leftIcon={<Square className='h-4 w-4' />}
                                >
                                    Stop listener
                                </Button>
                            </div>
                        </div>

                        <div className='flex flex-wrap items-center gap-x-5 gap-y-2 text-xs text-white/55'>
                            <span>
                                Mode{' '}
                                <span className='text-white/85'>
                                    {room.approval_mode}
                                </span>
                            </span>
                            {currentItem?.clip?.title && (
                                <span>
                                    Current clip{' '}
                                    <span className='text-white/85'>
                                        {currentItem.clip.title}
                                    </span>
                                </span>
                            )}
                            <span>
                                Listener started{' '}
                                <span className='text-white/85'>
                                    {formatDateTime(room.listener_started_at)}
                                </span>
                            </span>
                            <span>
                                Updated{' '}
                                <span className='text-white/85'>
                                    {formatDateTime(room.updated_at)}
                                </span>
                            </span>
                        </div>

                        <div className='flex flex-wrap items-center gap-2 rounded-2xl border border-white/10 bg-white/[0.03] px-3 py-2'>
                            <div className='mr-auto flex min-w-[180px] items-center gap-2 text-sm'>
                                <Clock3 className={`h-4 w-4 ${areSubmissionsOpen ? 'text-emerald-300' : 'text-white/45'}`} />
                                <span className='font-medium text-white'>Viewer submissions</span>
                                <span className='text-white/55'>
                                    {areSubmissionsOpen
                                        ? submissionCountdown
                                            ? `${submissionCountdown} remaining`
                                            : 'open until you close them'
                                        : isListenerLive
                                            ? 'closed'
                                            : 'start the listener first'}
                                </span>
                            </div>
                            <Button
                                size='sm'
                                variant='outline'
                                onClick={() => handleOpenSubmissions()}
                                disabled={!isListenerLive || areSubmissionsOpen}
                                loading={updateSubmissions.isPending}
                            >
                                Open until closed
                            </Button>
                            <label className='flex items-center gap-2 text-xs text-white/60'>
                                <span>Minutes</span>
                                <input
                                    aria-label='Submission duration in minutes'
                                    type='number'
                                    min={1}
                                    max={1440}
                                    value={submissionDurationMinutes}
                                    onChange={event => {
                                        const value = Number(event.target.value);
                                        setSubmissionDurationMinutes(Number.isFinite(value) ? value : 10);
                                    }}
                                    className='h-8 w-20 rounded-lg border border-white/15 bg-black/40 px-2 text-sm text-white outline-none focus:border-brand/60'
                                />
                            </label>
                            <Button
                                size='sm'
                                onClick={() => handleOpenSubmissions(submissionDurationMinutes)}
                                disabled={!isListenerLive || submissionDurationMinutes < 1 || submissionDurationMinutes > 1440}
                                loading={updateSubmissions.isPending}
                            >
                                Start timer
                            </Button>
                            <Button
                                size='sm'
                                variant='outline'
                                onClick={handleCloseSubmissions}
                                disabled={!areSubmissionsOpen}
                                loading={updateSubmissions.isPending}
                            >
                                Close submissions
                            </Button>
                        </div>

                        {updateSubmissions.isError && (
                            <div className='text-xs text-rose-200'>
                                {updateSubmissions.error instanceof Error
                                    ? updateSubmissions.error.message
                                    : 'Failed to update submissions'}
                            </div>
                        )}

                        {room.last_listener_error && (
                            <div className='flex items-start gap-2 rounded-2xl border border-rose-400/20 bg-rose-500/10 px-3 py-2 text-sm text-rose-100'>
                                <CircleAlert className='mt-0.5 h-4 w-4 shrink-0' />
                                <div>
                                    <p className='font-medium'>Listener issue</p>
                                    <p className='text-rose-100/80'>
                                        {room.last_listener_error}
                                    </p>
                                </div>
                            </div>
                        )}
                        {!isBotAuthorized && !botAuthorization.isLoading && (
                            <div className='flex items-start justify-between gap-3 rounded-2xl border border-amber-400/20 bg-amber-500/10 px-3 py-2 text-sm text-amber-100'>
                                <div>
                                    <p className='font-medium'>
                                        Let Clpr watch your Twitch chat
                                    </p>
                                    <p className='text-amber-100/80'>
                                        Twitch will ask you to grant channel bot access. The
                                        bot only queues supported clip links posted by viewers.
                                    </p>
                                </div>
                                <Button size='sm' onClick={handleAuthorizeBot}>
                                    Authorize bot
                                </Button>
                            </div>
                        )}
                        {!isClipDownloadAuthorized && !botAuthorization.isLoading && (
                            <div className='flex items-start justify-between gap-3 rounded-2xl border border-violet-400/20 bg-violet-500/10 px-3 py-2 text-sm text-violet-100'>
                                <div>
                                    <p className='font-medium'>
                                        Allow Clpr to download your Twitch clips
                                    </p>
                                    <p className='text-violet-100/80'>
                                        Twitch will ask you to grant clip-management access.
                                        Clpr uses it to request temporary official download URLs
                                        for clips from your channel so submitted clips can be processed and played.
                                    </p>
                                </div>
                                <Button size='sm' onClick={handleAuthorizeBot}>
                                    Authorize clip downloads
                                </Button>
                            </div>
                        )}
                    </div>
                </div>

                {hasApprovedItems ? (
                    <PlaylistTheatreMode
                        title={`${channel} Clip Room`}
                        items={theatreItems}
                        currentItemId={currentItemId}
                        onItemClick={handleItemClick}
                        onReorder={handleReorder}
                        onClose={() => navigate(`/stream/${channel}`)}
                        pendingLabel='Approved clips'
                        commentsLabel='Discussion'
                        extraSidebarTab={{
                            id: 'approvals',
                            label: 'Approvals',
                            count: pendingItems.length,
                            content: approvalContent,
                        }}
                        isQueue={false}
                        contained={false}
                        className='pt-24'
                    />
                ) : (
                    <div className='mx-auto flex min-h-screen max-w-[1600px] items-center px-4 pt-28 pb-8 lg:px-6'>
                        <div className='grid w-full gap-4 xl:grid-cols-[1.2fr_0.8fr]'>
                            <section className='rounded-[28px] border border-white/10 bg-white/[0.04] p-6 shadow-2xl shadow-black/30'>
                                <div className='max-w-2xl space-y-4'>
                                    <div className='flex items-center gap-2 text-xs uppercase tracking-[0.28em] text-white/35'>
                                        <Radio className='h-4 w-4 text-white/55' />
                                        Waiting for the first approved clip
                                    </div>
                                    <h2 className='text-3xl font-semibold tracking-tight text-white'>
                                        Streamer clip room is ready.
                                    </h2>
                                    <p className='max-w-xl text-sm leading-6 text-white/65'>
                                        Start the listener, then open submissions for as long as you want. Viewer clip links will flow into the pending rail until you close submissions or the timer runs out.
                                    </p>

                                    <div className='flex flex-wrap gap-2'>
                                        <Button
                                            onClick={handleStartListener}
                                            loading={startRoom.isPending}
                                            disabled={!hasRequiredTwitchAuthorization}
                                            leftIcon={<Radio className='h-4 w-4' />}
                                        >
                                            Start listener
                                        </Button>
                                        <Button
                                            variant='outline'
                                            onClick={handleStopListener}
                                            loading={stopRoom.isPending}
                                            leftIcon={<Square className='h-4 w-4' />}
                                        >
                                            Stop listener
                                        </Button>
                                    </div>

                                    <div className='grid gap-3 pt-2 sm:grid-cols-3'>
                                        <div className='rounded-2xl border border-white/10 bg-black/20 p-4'>
                                            <p className='text-[11px] uppercase tracking-[0.22em] text-white/35'>
                                                Listener
                                            </p>
                                            <p className='mt-1 text-sm font-medium text-white'>
                                                {isListenerLive ? 'Active' : 'Stopped'}
                                            </p>
                                        </div>
                                        <div className='rounded-2xl border border-white/10 bg-black/20 p-4'>
                                            <p className='text-[11px] uppercase tracking-[0.22em] text-white/35'>
                                                Approved
                                            </p>
                                            <p className='mt-1 text-sm font-medium text-white'>
                                                0 clips
                                            </p>
                                        </div>
                                        <div className='rounded-2xl border border-white/10 bg-black/20 p-4'>
                                            <p className='text-[11px] uppercase tracking-[0.22em] text-white/35'>
                                                Pending
                                            </p>
                                            <p className='mt-1 text-sm font-medium text-white'>
                                                {pendingItems.length} clips
                                            </p>
                                        </div>
                                    </div>
                                </div>
                            </section>

                            <aside className='rounded-[28px] border border-white/10 bg-white/[0.03] shadow-2xl shadow-black/20'>
                                <PendingApprovalsRail
                                    pendingItems={pendingItems}
                                    skippedItems={skippedItems}
                                    onApprove={handleApprove}
                                    onReject={handleReject}
                                />
                            </aside>
                        </div>
                    </div>
                )}
            </div>
        </>
    );
}
