import { useCallback, useEffect, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import {
    Alert,
    Badge,
    Button,
    Card,
    ConfirmModal,
    Container,
    Modal,
    Spinner,
    TextArea,
} from '../../components';
import { useAuth } from '../../context/AuthContext';
import {
    approveQueueItem,
    bulkModerate,
    getModerationQueue,
    getModerationStats,
    rejectQueueItem,
    type ModerationQueueItem,
    type ModerationQueueStats,
} from '../../lib/moderation-api';
import { getErrorMessage } from '../../lib/error-utils';
import { QueueItemContent } from '../../components/moderation/QueueItemContent';
import {
    bulkResultMessage,
    isQueueShortcutEvent,
    isStaleQueueError,
} from '../../components/moderation/queueUtils';

// Helper function to format confidence score
function formatConfidenceScore(score: number | undefined): string {
    if (score === undefined) return '';
    return ` (${(score * 100).toFixed(0)}%)`;
}

export function AdminModerationQueuePage() {
    const { isAuthenticated, isModeratorOrAdmin } = useAuth();
    const navigate = useNavigate();
    const [items, setItems] = useState<ModerationQueueItem[]>([]);
    const [stats, setStats] = useState<ModerationQueueStats | null>(null);
    const [isLoading, setIsLoading] = useState(true);
    const [error, setError] = useState<string | null>(null);
    const [success, setSuccess] = useState<string | null>(null);

    // Filters
    const [statusFilter, setStatusFilter] = useState<string>('pending');
    const [typeFilter, setTypeFilter] = useState<string>('');

    // Bulk selection
    const [selectedIds, setSelectedIds] = useState<Set<string>>(new Set());

    // Rejection modal state
    const [rejectModalOpen, setRejectModalOpen] = useState(false);
    const [selectedItemId, setSelectedItemId] = useState<string | null>(null);
    const [rejectionReason, setRejectionReason] = useState('');
    const [bulkRejectOpen, setBulkRejectOpen] = useState(false);

    // Data loaders
    const loadQueue = useCallback(async () => {
        try {
            setIsLoading(true);
            setError(null);
            const response = await getModerationQueue(
                statusFilter,
                typeFilter || undefined,
                50
            );
            setItems(response.data || []);
        } catch (err: unknown) {
            setError(getErrorMessage(err, 'Failed to load moderation queue'));
        } finally {
            setIsLoading(false);
        }
    }, [statusFilter, typeFilter]);

    const loadStats = useCallback(async () => {
        try {
            const response = await getModerationStats();
            setStats(response.data);
        } catch (err: unknown) {
            console.error('Failed to load stats:', err);
        }
    }, []);

    // Shows the error and, when the item is gone or already reviewed,
    // reloads so the list matches the server again.
    const handleActionError = useCallback(
        (err: unknown, fallback: string) => {
            if (isStaleQueueError(err)) {
                setSelectedIds(new Set());
                loadQueue();
                loadStats();
            }
            setError(getErrorMessage(err, fallback));
        },
        [loadQueue, loadStats]
    );

    // Bulk action handlers must be defined before effects that reference them
    const handleBulkApprove = useCallback(async () => {
        if (selectedIds.size === 0) return;

        try {
            const result = await bulkModerate({
                item_ids: Array.from(selectedIds),
                action: 'approve',
            });
            setSuccess(bulkResultMessage(result, 'items', 'approved successfully'));
            setSelectedIds(new Set());
            loadQueue();
            loadStats();
        } catch (err: unknown) {
            handleActionError(err, 'Failed to approve items');
        }
    }, [selectedIds, loadQueue, loadStats, handleActionError]);

    const handleBulkReject = useCallback(async () => {
        if (selectedIds.size === 0) return;

        try {
            const result = await bulkModerate({
                item_ids: Array.from(selectedIds),
                action: 'reject',
            });
            setSuccess(bulkResultMessage(result, 'items', 'rejected and removed'));
            setSelectedIds(new Set());
            loadQueue();
            loadStats();
        } catch (err: unknown) {
            handleActionError(err, 'Failed to reject items');
        }
    }, [selectedIds, loadQueue, loadStats, handleActionError]);

    // Keyboard shortcuts
    useEffect(() => {
        const handleKeyPress = (e: KeyboardEvent) => {
            // Shortcuts apply to pending items only, never with Ctrl/Cmd/Alt,
            // and not while typing in a field or while a dialog is open.
            if (
                statusFilter !== 'pending' ||
                selectedIds.size === 0 ||
                rejectModalOpen ||
                bulkRejectOpen ||
                !isQueueShortcutEvent(e)
            ) {
                return;
            }

            if (e.key === 'a') {
                // Approve selected
                e.preventDefault();
                handleBulkApprove();
            } else if (e.key === 'r') {
                // Rejecting removes content, so ask first.
                e.preventDefault();
                setBulkRejectOpen(true);
            }
        };

        document.addEventListener('keydown', handleKeyPress);
        return () => document.removeEventListener('keydown', handleKeyPress);
    }, [selectedIds, statusFilter, rejectModalOpen, bulkRejectOpen, handleBulkApprove]);

    // (moved above)

    useEffect(() => {
        if (!isAuthenticated || !isModeratorOrAdmin) {
            navigate('/');
            return;
        }

        loadQueue();
        loadStats();
    }, [isAuthenticated, isModeratorOrAdmin, navigate, loadQueue, loadStats]);

    // Clear selection when filters change
    useEffect(() => {
        setSelectedIds(new Set());
    }, [statusFilter, typeFilter]);

    // Auto-dismiss success messages after 5 seconds
    useEffect(() => {
        if (success) {
            const timer = setTimeout(() => {
                setSuccess(null);
            }, 5000);
            return () => clearTimeout(timer);
        }
    }, [success]);

    const handleApprove = async (itemId: string) => {
        try {
            await approveQueueItem(itemId);
            setSuccess('Item approved successfully!');
            loadQueue();
            loadStats();
        } catch (err: unknown) {
            handleActionError(err, 'Failed to approve item');
        }
    };

    const openRejectModal = (itemId: string) => {
        setSelectedItemId(itemId);
        setRejectModalOpen(true);
        setRejectionReason('');
    };

    const handleReject = async () => {
        if (!selectedItemId) {
            return;
        }

        try {
            await rejectQueueItem(selectedItemId, rejectionReason || undefined);
            setSuccess('Item rejected. Rejected comments and clips are removed from the site.');
            setRejectModalOpen(false);
            setSelectedItemId(null);
            setRejectionReason('');
            loadQueue();
            loadStats();
        } catch (err: unknown) {
            // Close the dialog so the error and the reloaded list are visible.
            setRejectModalOpen(false);
            setSelectedItemId(null);
            handleActionError(err, 'Failed to reject item');
        }
    };

    // (moved above)

    const toggleSelection = (itemId: string) => {
        const newSelection = new Set(selectedIds);
        if (newSelection.has(itemId)) {
            newSelection.delete(itemId);
        } else {
            newSelection.add(itemId);
        }
        setSelectedIds(newSelection);
    };

    const toggleSelectAll = () => {
        if (selectedIds.size === items.length) {
            setSelectedIds(new Set());
        } else {
            setSelectedIds(new Set(items.map((item) => item.id)));
        }
    };

    if (!isAuthenticated || !isModeratorOrAdmin) {
        return null;
    }

    return (
        <Container className='py-8'>
            <div className='max-w-6xl mx-auto'>
                <div className='mb-6'>
                    <h1 className='mb-2 text-3xl font-bold'>
                        Moderation Queue
                    </h1>
                    <p className='text-muted-foreground'>
                        Review content that was flagged automatically. Approve keeps it; Reject removes it.
                    </p>
                    <p className='text-muted-foreground text-sm mt-2'>
                        Keyboard shortcuts: A approve • R reject (when items chosen)
                    </p>
                </div>

                {error && (
                    <Alert
                        variant='error'
                        className='mb-6'
                    >
                        {error}
                    </Alert>
                )}

                {success && (
                    <Alert
                        variant='success'
                        className='mb-6'
                    >
                        {success}
                    </Alert>
                )}

                {/* Stats */}
                {stats && (
                    <Card className='p-4 mb-6'>
                        <div className='gap-6 grid grid-cols-2 md:grid-cols-4'>
                            <div>
                                <div className='text-2xl font-bold'>
                                    {stats.total_pending}
                                </div>
                                <div className='text-muted-foreground text-sm'>
                                    Pending
                                </div>
                            </div>
                            <div>
                                <div className='text-2xl font-bold text-green-600'>
                                    {stats.total_approved}
                                </div>
                                <div className='text-muted-foreground text-sm'>
                                    Approved
                                </div>
                            </div>
                            <div>
                                <div className='text-2xl font-bold text-red-600'>
                                    {stats.total_rejected}
                                </div>
                                <div className='text-muted-foreground text-sm'>
                                    Rejected
                                </div>
                            </div>
                            <div>
                                <div className='text-2xl font-bold text-yellow-600'>
                                    {stats.high_priority_count}
                                </div>
                                <div className='text-muted-foreground text-sm'>
                                    High Priority
                                </div>
                            </div>
                        </div>
                    </Card>
                )}

                {/* Filters */}
                <Card className='p-4 mb-6'>
                    <div className='flex gap-4 items-center'>
                        <div className='flex-1'>
                            <label className='block mb-2 text-sm font-medium'>
                                Status
                            </label>
                            <select
                                value={statusFilter}
                                onChange={(e) => setStatusFilter(e.target.value)}
                                className='bg-background-secondary border-border focus:ring-primary focus:border-primary block px-3 py-2 w-full rounded border'
                            >
                                <option value='pending'>Pending</option>
                                <option value='approved'>Approved</option>
                                <option value='rejected'>Rejected</option>
                                <option value='escalated'>Escalated</option>
                            </select>
                        </div>
                        <div className='flex-1'>
                            <label className='block mb-2 text-sm font-medium'>
                                Content Type
                            </label>
                            <select
                                value={typeFilter}
                                onChange={(e) => setTypeFilter(e.target.value)}
                                className='bg-background-secondary border-border focus:ring-primary focus:border-primary block px-3 py-2 w-full rounded border'
                            >
                                <option value=''>All Types</option>
                                <option value='comment'>Comments</option>
                                <option value='clip'>Clips</option>
                                <option value='user'>Users</option>
                                <option value='submission'>Submissions</option>
                            </select>
                        </div>
                        <div className='self-end'>
                            <Button
                                onClick={() => {
                                    loadQueue();
                                    loadStats();
                                }}
                                variant='secondary'
                                disabled={isLoading}
                            >
                                Refresh
                            </Button>
                        </div>
                    </div>
                </Card>

                {/* Bulk Actions Bar */}
                {statusFilter === 'pending' && selectedIds.size > 0 && (
                    <Card className='bg-blue-50 dark:bg-blue-900/20 p-4 mb-4'>
                        <div className='flex justify-between items-center'>
                            <span className='font-medium'>
                                {selectedIds.size} item{selectedIds.size !== 1 ? 's' : ''} selected
                            </span>
                            <div className='space-x-2'>
                                <Button
                                    onClick={handleBulkApprove}
                                    variant='primary'
                                    className='hover:bg-green-700 bg-green-600'
                                >
                                    Approve All (A)
                                </Button>
                                <Button
                                    onClick={() => setBulkRejectOpen(true)}
                                    className='bg-red-600/90 hover:bg-red-600 text-white'
                                >
                                    Reject All (R)
                                </Button>
                                <Button
                                    onClick={() => setSelectedIds(new Set())}
                                    variant='secondary'
                                >
                                    Clear
                                </Button>
                            </div>
                        </div>
                    </Card>
                )}

                {/* Queue Items */}
                <Card className='p-6'>
                    {isLoading ? (
                        <div className='flex justify-center py-12'>
                            <Spinner size='lg' />
                        </div>
                    ) : !items || items.length === 0 ? (
                        <div className='py-12 text-center'>
                            <p className='text-muted-foreground'>
                                No items in the queue.
                            </p>
                        </div>
                    ) : (
                        <div className='space-y-4'>
                            {/* Select All */}
                            {statusFilter === 'pending' && (
                                <div className='flex items-center gap-2 pb-2 border-b'>
                                    <input
                                        type='checkbox'
                                        checked={selectedIds.size === items.length && items.length > 0}
                                        onChange={toggleSelectAll}
                                        className='rounded'
                                        aria-label='Select all items'
                                    />
                                    <span className='text-sm font-medium'>
                                        Select All
                                    </span>
                                </div>
                            )}

                            {items.map((item) => (
                                <div
                                    key={item.id}
                                    className='bg-background-secondary hover:bg-background-tertiary p-4 rounded-lg transition-colors'
                                >
                                    <div className='flex gap-4 items-start'>
                                        {/* Checkbox */}
                                        {statusFilter === 'pending' && (
                                            <input
                                                type='checkbox'
                                                checked={selectedIds.has(item.id)}
                                                onChange={() => toggleSelection(item.id)}
                                                className='mt-1 rounded'
                                                aria-label={`Select ${item.content_type} ${item.content_id}`}
                                            />
                                        )}

                                        {/* Content */}
                                        <div className='flex-1'>
                                            <div className='flex gap-2 items-center mb-2'>
                                                <Badge variant='default'>
                                                    {item.content_type}
                                                </Badge>
                                                <Badge
                                                    variant={
                                                        item.priority >= 75
                                                            ? 'error'
                                                            : item.priority >= 50
                                                            ? 'warning'
                                                            : 'default'
                                                    }
                                                >
                                                    Priority: {item.priority}
                                                </Badge>
                                                {item.auto_flagged && (
                                                    <Badge variant='warning'>
                                                        Auto-flagged
                                                        {formatConfidenceScore(item.confidence_score)}
                                                    </Badge>
                                                )}
                                                {item.report_count > 0 && (
                                                    <Badge variant='error'>
                                                        {item.report_count} report{item.report_count !== 1 ? 's' : ''}
                                                    </Badge>
                                                )}
                                            </div>

                                            <div className='text-muted-foreground space-y-1 text-sm'>
                                                <p>
                                                    <span className='font-medium'>Reason:</span> {item.reason}
                                                </p>
                                                <QueueItemContent
                                                    item={item}
                                                    idLabel='Content ID'
                                                />
                                                <p>
                                                    <span className='font-medium'>Created:</span>{' '}
                                                    {new Date(item.created_at).toLocaleString()}
                                                </p>
                                            </div>
                                        </div>

                                        {/* Actions */}
                                        {statusFilter === 'pending' && (
                                            <div className='flex gap-2'>
                                                <Button
                                                    onClick={() => handleApprove(item.id)}
                                                    variant='primary'
                                                    size='sm'
                                                    className='hover:bg-green-700 bg-green-600'
                                                >
                                                    Approve
                                                </Button>
                                                <Button
                                                    onClick={() => openRejectModal(item.id)}
                                                    size='sm'
                                                    className='bg-red-600/90 hover:bg-red-600 text-white'
                                                >
                                                    Reject
                                                </Button>
                                            </div>
                                        )}
                                    </div>
                                </div>
                            ))}
                        </div>
                    )}
                </Card>
            </div>

            {/* Reject Modal */}
            {rejectModalOpen && (
                <Modal
                    open={rejectModalOpen}
                    onClose={() => {
                        setRejectModalOpen(false);
                        setSelectedItemId(null);
                        setRejectionReason('');
                    }}
                    title='Reject Item'
                >
                    <div className='space-y-4'>
                        <p className='text-muted-foreground'>
                            Rejecting a comment or clip removes it from the site. You can add a reason (optional).
                        </p>
                        <TextArea
                            value={rejectionReason}
                            onChange={(e) => setRejectionReason(e.target.value)}
                            placeholder='Reason for rejection...'
                            rows={4}
                        />
                        <div className='flex gap-3'>
                            <Button
                                onClick={handleReject}
                                className='hover:bg-red-700 flex-1 bg-red-600'
                            >
                                Reject Item
                            </Button>
                            <Button
                                onClick={() => {
                                    setRejectModalOpen(false);
                                    setSelectedItemId(null);
                                    setRejectionReason('');
                                }}
                                variant='secondary'
                                className='flex-1'
                            >
                                Cancel
                            </Button>
                        </div>
                    </div>
                </Modal>
            )}

            <ConfirmModal
                open={bulkRejectOpen}
                onClose={() => setBulkRejectOpen(false)}
                onConfirm={handleBulkReject}
                title={`Reject ${selectedIds.size} item${selectedIds.size !== 1 ? 's' : ''}?`}
                message='Rejected comments and clips are removed from the site.'
                confirmText='Reject and remove'
                variant='danger'
            />
        </Container>
    );
}
