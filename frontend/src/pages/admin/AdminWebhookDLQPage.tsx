import { useState, useEffect } from 'react';
import {
    Container,
    Card,
    Button,
    Badge,
    Spinner,
    Alert,
    Modal,
} from '../../components';
import { CheckCircle } from 'lucide-react';
import { useAuth } from '../../context/AuthContext';
import { useNavigate } from 'react-router-dom';
import {
    getWebhookDLQItems,
    replayWebhookDLQItem,
    deleteWebhookDLQItem,
} from '../../lib/webhook-api';
import type { OutboundWebhookDLQItem } from '../../types/webhook';

export function AdminWebhookDLQPage() {
    const { isAuthenticated, isAdmin } = useAuth();
    const navigate = useNavigate();
    const [items, setItems] = useState<OutboundWebhookDLQItem[]>([]);
    const [isLoading, setIsLoading] = useState(true);
    const [error, setError] = useState<string | null>(null);
    const [success, setSuccess] = useState<string | null>(null);
    const [page, setPage] = useState(1);
    const [totalPages, setTotalPages] = useState(1);
    const [total, setTotal] = useState(0);

    // Action modal state
    const [actionModalOpen, setActionModalOpen] = useState(false);
    const [selectedItem, setSelectedItem] =
        useState<OutboundWebhookDLQItem | null>(null);
    const [actionType, setActionType] = useState<'replay' | 'delete'>('replay');
    const [isProcessing, setIsProcessing] = useState(false);

    // Payload modal state
    const [payloadModalOpen, setPayloadModalOpen] = useState(false);
    const [selectedPayload, setSelectedPayload] = useState<string>('');

    const loadDLQItems = async () => {
        try {
            setIsLoading(true);
            setError(null);
            const response = await getWebhookDLQItems(page, 20);
            setItems(response.items || []);
            setTotalPages(response.pagination.total_pages);
            setTotal(response.pagination.total);
        } catch (err: unknown) {
            const error = err as { response?: { data?: { error?: string } } };
            setError(error.response?.data?.error || 'Failed to load DLQ items');
        } finally {
            setIsLoading(false);
        }
    };

    useEffect(() => {
        if (!isAuthenticated || !isAdmin) {
            navigate('/');
            return;
        }

        // Clear any previous success message when (re)loading items, e.g. on page change
        setSuccess(null);
        loadDLQItems();
        // eslint-disable-next-line react-hooks/exhaustive-deps
    }, [isAuthenticated, isAdmin, navigate, page]);

    // Auto-dismiss success and error messages after a short delay
    useEffect(() => {
        if (!success && !error) {
            return;
        }

        const timeoutId = window.setTimeout(() => {
            setSuccess(null);
            setError(null);
        }, 5000);

        return () => {
            window.clearTimeout(timeoutId);
        };
    }, [success, error]);

    const openActionModal = (
        item: OutboundWebhookDLQItem,
        action: 'replay' | 'delete'
    ) => {
        setSelectedItem(item);
        setActionType(action);
        setActionModalOpen(true);
    };

    const handleAction = async () => {
        if (!selectedItem) return;

        try {
            setIsProcessing(true);
            setError(null);

            if (actionType === 'replay') {
                await replayWebhookDLQItem(selectedItem.id);
                setSuccess('Webhook replayed successfully!');
            } else {
                await deleteWebhookDLQItem(selectedItem.id);
                setSuccess('DLQ item deleted successfully!');
            }

            setActionModalOpen(false);
            setSelectedItem(null);
            loadDLQItems();
        } catch (err: unknown) {
            const error = err as { response?: { data?: { error?: string } } };
            setError(
                error.response?.data?.error ||
                    `Failed to ${actionType} DLQ item`
            );
        } finally {
            setIsProcessing(false);
        }
    };

    const viewPayload = (payload: string) => {
        setSelectedPayload(payload);
        setPayloadModalOpen(true);
    };

    const formatDate = (dateString: string) => {
        return new Date(dateString).toLocaleString();
    };

    const formatPayload = (payload: string) => {
        try {
            return JSON.stringify(JSON.parse(payload), null, 2);
        } catch {
            return payload;
        }
    };

    const getStatusBadge = (item: OutboundWebhookDLQItem) => {
        if (item.replayed_at) {
            if (item.replay_successful) {
                return (
                    <Badge variant='success' className='text-xs'>
                        Replayed ✓
                    </Badge>
                );
            }
            return (
                <Badge variant='error' className='text-xs'>
                    Replay Failed
                </Badge>
            );
        }
        return (
            <Badge variant='warning' className='text-xs'>
                Pending
            </Badge>
        );
    };

    if (!isAuthenticated || !isAdmin) {
        return null;
    }

    return (
        <Container className='py-4 xs:py-6 md:py-8'>
            <div className='mb-6 xs:mb-8'>
                <h1 className='text-2xl xs:text-3xl font-bold mb-2'>
                    Webhook Dead-Letter Queue
                </h1>
                <p className='text-sm xs:text-base text-muted-foreground'>
                    Failed webhook deliveries that have exhausted retry attempts
                </p>
            </div>

            {error && (
                <Alert variant='error' className='mb-4'>
                    {error}
                </Alert>
            )}

            {success && (
                <Alert variant='success' className='mb-4'>
                    {success}
                </Alert>
            )}

            {/* Stats */}
            <Card className='mb-6'>
                <div className='p-4'>
                    <div className='flex flex-wrap gap-6'>
                        <div>
                            <p className='text-sm text-muted-foreground'>
                                Total Failed Deliveries
                            </p>
                            <p className='text-2xl font-bold'>{total}</p>
                        </div>
                        <div>
                            <p className='text-sm text-muted-foreground'>
                                Current Page
                            </p>
                            <p className='text-2xl font-bold'>
                                {page} / {totalPages}
                            </p>
                        </div>
                    </div>
                </div>
            </Card>

            {/* DLQ Items Table */}
            {isLoading ? (
                <div className='flex justify-center py-12'>
                    <Spinner size='lg' />
                </div>
            ) : items.length === 0 ? (
                <Card>
                    <div className='p-8 text-center'>
                        <p className='text-muted-foreground'>
                            No failed webhook deliveries found. All deliveries
                            are being processed successfully!{' '}
                            <CheckCircle className="inline" size={16} strokeWidth={1.75} />
                        </p>
                    </div>
                </Card>
            ) : (
                <>
                    <div className='overflow-x-auto'>
                        <table
                            className='w-full border-collapse'
                            role='table'
                            aria-label='Failed webhook deliveries'
                        >
                            <thead>
                                <tr className='bg-muted'>
                                    <th className='text-left p-3 text-sm font-semibold'>
                                        Event Type
                                    </th>
                                    <th className='text-left p-3 text-sm font-semibold'>
                                        Status
                                    </th>
                                    <th className='text-left p-3 text-sm font-semibold'>
                                        Error
                                    </th>
                                    <th className='text-left p-3 text-sm font-semibold'>
                                        Attempts
                                    </th>
                                    <th className='text-left p-3 text-sm font-semibold'>
                                        Failed At
                                    </th>
                                    <th className='text-left p-3 text-sm font-semibold'>
                                        Actions
                                    </th>
                                </tr>
                            </thead>
                            <tbody>
                                {items.map(item => (
                                    <tr
                                        key={item.id}
                                        className='border-b border-border hover:bg-accent/50'
                                    >
                                        <td className='p-3'>
                                            <div>
                                                <p className='font-mono text-sm'>
                                                    {item.event_type}
                                                </p>
                                                <p className='text-xs text-muted-foreground'>
                                                    ID:{' '}
                                                    {item.event_id.substring(
                                                        0,
                                                        8
                                                    )}
                                                    ...
                                                </p>
                                            </div>
                                        </td>
                                        <td className='p-3'>
                                            {getStatusBadge(item)}
                                            {item.http_status_code && (
                                                <p className='text-xs text-muted-foreground mt-1'>
                                                    HTTP{' '}
                                                    {item.http_status_code}
                                                </p>
                                            )}
                                        </td>
                                        <td className='p-3'>
                                            <p className='text-sm line-clamp-2 max-w-xs'>
                                                {item.error_message}
                                            </p>
                                        </td>
                                        <td className='p-3'>
                                            <Badge
                                                variant='default'
                                                className='text-xs'
                                            >
                                                {item.attempt_count} attempts
                                            </Badge>
                                        </td>
                                        <td className='p-3'>
                                            <p className='text-sm'>
                                                {formatDate(
                                                    item.moved_to_dlq_at
                                                )}
                                            </p>
                                            {item.replayed_at && (
                                                <p className='text-xs text-muted-foreground mt-1'>
                                                    Replayed:{' '}
                                                    {formatDate(
                                                        item.replayed_at
                                                    )}
                                                </p>
                                            )}
                                        </td>
                                        <td className='p-3'>
                                            <div className='flex gap-2'>
                                                <Button
                                                    size='sm'
                                                    variant='outline'
                                                    onClick={() =>
                                                        viewPayload(item.payload)
                                                    }
                                                    aria-label={`View payload for ${item.event_type} event ${item.event_id.substring(0, 8)}`}
                                                >
                                                    View
                                                </Button>
                                                <Button
                                                    size='sm'
                                                    variant='primary'
                                                    onClick={() =>
                                                        openActionModal(
                                                            item,
                                                            'replay'
                                                        )
                                                    }
                                                    disabled={
                                                        item.replay_successful
                                                    }
                                                    aria-label={`Replay ${item.event_type} event ${item.event_id.substring(0, 8)}`}
                                                >
                                                    Replay
                                                </Button>
                                                <Button
                                                    size='sm'
                                                    variant='error'
                                                    onClick={() =>
                                                        openActionModal(
                                                            item,
                                                            'delete'
                                                        )
                                                    }
                                                    aria-label={`Delete ${item.event_type} event ${item.event_id.substring(0, 8)}`}
                                                >
                                                    Delete
                                                </Button>
                                            </div>
                                        </td>
                                    </tr>
                                ))}
                            </tbody>
                        </table>
                    </div>

                    {/* Pagination */}
                    <div className='mt-6 flex justify-between items-center'>
                        <Button
                            variant='outline'
                            onClick={() => setPage(p => Math.max(1, p - 1))}
                            disabled={page === 1 || totalPages === 0}
                        >
                            Previous
                        </Button>
                        <span className='text-sm text-muted-foreground'>
                            Page {page} of {totalPages || 1}
                        </span>
                        <Button
                            variant='outline'
                            onClick={() =>
                                setPage(p => Math.min(totalPages, p + 1))
                            }
                            disabled={page >= totalPages || totalPages === 0}
                        >
                            Next
                        </Button>
                    </div>
                </>
            )}

            {/* Action Confirmation Modal */}
            <Modal
                isOpen={actionModalOpen}
                onClose={() => !isProcessing && setActionModalOpen(false)}
                title={
                    actionType === 'replay'
                        ? 'Replay Webhook Delivery'
                        : 'Delete DLQ Item'
                }
            >
                <div className='space-y-4'>
                    <p className='text-sm text-muted-foreground'>
                        {actionType === 'replay'
                            ? 'This will attempt to redeliver the webhook to the subscription endpoint. The webhook will be sent with the same payload and headers as the original delivery.'
                            : 'This will permanently delete this DLQ item. This action cannot be undone.'}
                    </p>

                    {selectedItem && (
                        <Card>
                            <div className='p-4 space-y-2'>
                                <div>
                                    <span className='text-sm font-semibold'>
                                        Event Type:
                                    </span>
                                    <span className='text-sm ml-2'>
                                        {selectedItem.event_type}
                                    </span>
                                </div>
                                <div>
                                    <span className='text-sm font-semibold'>
                                        Error:
                                    </span>
                                    <p className='text-sm ml-2 text-muted-foreground'>
                                        {selectedItem.error_message}
                                    </p>
                                </div>
                                <div>
                                    <span className='text-sm font-semibold'>
                                        Attempts:
                                    </span>
                                    <span className='text-sm ml-2'>
                                        {selectedItem.attempt_count}
                                    </span>
                                </div>
                            </div>
                        </Card>
                    )}

                    <div className='flex justify-end gap-3 mt-6'>
                        <Button
                            variant='outline'
                            onClick={() => setActionModalOpen(false)}
                            disabled={isProcessing}
                        >
                            Cancel
                        </Button>
                        <Button
                            variant={
                                actionType === 'replay' ? 'primary' : 'error'
                            }
                            onClick={handleAction}
                            disabled={isProcessing}
                        >
                            {isProcessing ? (
                                <Spinner size='sm' />
                            ) : actionType === 'replay' ? (
                                'Replay'
                            ) : (
                                'Delete'
                            )}
                        </Button>
                    </div>
                </div>
            </Modal>

            {/* Payload View Modal */}
            <Modal
                isOpen={payloadModalOpen}
                onClose={() => setPayloadModalOpen(false)}
                title='Webhook Payload'
            >
                <div className='space-y-4'>
                    <pre className='bg-muted p-4 rounded-lg overflow-x-auto text-xs'>
                        <code>{formatPayload(selectedPayload)}</code>
                    </pre>
                    <div className='flex justify-end'>
                        <Button
                            variant='outline'
                            onClick={() => setPayloadModalOpen(false)}
                        >
                            Close
                        </Button>
                    </div>
                </div>
            </Modal>
        </Container>
    );
}
