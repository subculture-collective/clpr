import { useState, useEffect } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { Helmet } from '@dr.pogodin/react-helmet';
import { Link } from 'react-router-dom';
import {
    Alert,
    Button,
    Card,
    CardBody,
    CardHeader,
    Container,
    Input,
    Modal,
    Stack,
    TextArea,
    Toggle,
} from '../components';
import { useToast } from '../context/ToastContext';
import {
    createWebhookSubscription,
    deleteWebhookSubscription,
    getWebhookDeliveries,
    getSupportedWebhookEvents,
    listWebhookSubscriptions,
    regenerateWebhookSecret,
    updateWebhookSubscription,
} from '../lib/webhook-api';
import type {
    CreateWebhookSubscriptionRequest,
    UpdateWebhookSubscriptionRequest,
    WebhookSubscription,
    WebhookDelivery,
} from '../types/webhook';

export function WebhookSubscriptionsPage() {
    const queryClient = useQueryClient();
    const { error: showErrorToast } = useToast();
    const [showCreateModal, setShowCreateModal] = useState(false);
    const [showEditModal, setShowEditModal] = useState(false);
    const [showDeleteModal, setShowDeleteModal] = useState(false);
    const [showSecretModal, setShowSecretModal] = useState(false);
    const [showDeliveriesModal, setShowDeliveriesModal] = useState(false);
    const [selectedSubscription, setSelectedSubscription] =
        useState<WebhookSubscription | null>(null);
    const [newSecret, setNewSecret] = useState<string | null>(null);
    const [copiedSecret, setCopiedSecret] = useState(false);

    // Form state for create
    const [createForm, setCreateForm] =
        useState<CreateWebhookSubscriptionRequest>({
            url: '',
            events: [],
            description: '',
        });

    // Form state for edit
    const [editForm, setEditForm] =
        useState<UpdateWebhookSubscriptionRequest>({});

    // Load subscriptions
    const { data: subscriptions, isLoading } = useQuery({
        queryKey: ['webhookSubscriptions'],
        queryFn: listWebhookSubscriptions,
    });

    // Load supported events
    const { data: supportedEvents } = useQuery({
        queryKey: ['webhookEvents'],
        queryFn: getSupportedWebhookEvents,
    });

    // Create mutation
    const createMutation = useMutation({
        mutationFn: createWebhookSubscription,
        onSuccess: (data) => {
            queryClient.invalidateQueries({ queryKey: ['webhookSubscriptions'] });
            setNewSecret(data.secret);
            setShowCreateModal(false);
            setShowSecretModal(true);
            setCreateForm({ url: '', events: [], description: '' });
        },
    });

    // Update mutation
    const updateMutation = useMutation({
        mutationFn: ({
            id,
            data,
        }: {
            id: string;
            data: UpdateWebhookSubscriptionRequest;
        }) => updateWebhookSubscription(id, data),
        onSuccess: () => {
            queryClient.invalidateQueries({ queryKey: ['webhookSubscriptions'] });
            setShowEditModal(false);
            setSelectedSubscription(null);
        },
    });

    // Delete mutation
    const deleteMutation = useMutation({
        mutationFn: deleteWebhookSubscription,
        onSuccess: () => {
            queryClient.invalidateQueries({ queryKey: ['webhookSubscriptions'] });
            setShowDeleteModal(false);
            setSelectedSubscription(null);
        },
    });

    // Regenerate secret mutation
    const regenerateSecretMutation = useMutation({
        mutationFn: regenerateWebhookSecret,
        onSuccess: (secret) => {
            setNewSecret(secret);
            setShowSecretModal(true);
        },
    });

    // Handle create submit
    const handleCreateSubmit = (e: React.FormEvent) => {
        e.preventDefault();
        createMutation.mutate(createForm);
    };

    // Handle edit submit
    const handleEditSubmit = (e: React.FormEvent) => {
        e.preventDefault();
        if (selectedSubscription) {
            updateMutation.mutate({
                id: selectedSubscription.id,
                data: editForm,
            });
        }
    };

    // Handle delete
    const handleDelete = () => {
        if (selectedSubscription) {
            deleteMutation.mutate(selectedSubscription.id);
        }
    };

    // Handle regenerate secret
    const handleRegenerateSecret = (subscription: WebhookSubscription) => {
        setSelectedSubscription(subscription);
        regenerateSecretMutation.mutate(subscription.id);
    };

    // Copy secret to clipboard
    const copySecretToClipboard = async () => {
        if (newSecret) {
            try {
                await navigator.clipboard.writeText(newSecret);
                setCopiedSecret(true);
                setTimeout(() => setCopiedSecret(false), 2000);
            } catch (error) {
                console.error('Failed to copy to clipboard:', error);
                showErrorToast('Failed to copy to clipboard. Please copy manually.');
            }
        }
    };

    // Open edit modal
    const openEditModal = (subscription: WebhookSubscription) => {
        setSelectedSubscription(subscription);
        setEditForm({
            url: subscription.url,
            events: subscription.events,
            is_active: subscription.is_active,
            description: subscription.description,
        });
        setShowEditModal(true);
    };

    // Open delete modal
    const openDeleteModal = (subscription: WebhookSubscription) => {
        setSelectedSubscription(subscription);
        setShowDeleteModal(true);
    };

    // Open deliveries modal
    const openDeliveriesModal = (subscription: WebhookSubscription) => {
        setSelectedSubscription(subscription);
        setShowDeliveriesModal(true);
    };

    // Toggle event in create form
    const toggleCreateEvent = (event: string) => {
        setCreateForm({
            ...createForm,
            events: createForm.events.includes(event)
                ? createForm.events.filter((e) => e !== event)
                : [...createForm.events, event],
        });
    };

    // Toggle event in edit form
    const toggleEditEvent = (event: string) => {
        const currentEvents = editForm.events || [];
        setEditForm({
            ...editForm,
            events: currentEvents.includes(event)
                ? currentEvents.filter((e) => e !== event)
                : [...currentEvents, event],
        });
    };

    return (
        <>
            <Helmet>
                <title>Webhook Subscriptions - clpr</title>
            </Helmet>

            <Container className="py-4 xs:py-6 md:py-8">
                <div className="max-w-5xl mx-auto">
                    <div className="flex justify-between items-center mb-4 xs:mb-6">
                        <h1 className="text-2xl xs:text-3xl font-bold">
                            Webhook Subscriptions
                        </h1>
                        <Button
                            variant="primary"
                            onClick={() => setShowCreateModal(true)}
                        >
                            Create Webhook
                        </Button>
                    </div>

                    <Alert variant="info" className="mb-6">
                        <p className="text-sm">
                            Webhooks allow you to receive real-time notifications
                            when events occur. Configure webhook endpoints to
                            integrate with your own applications.{' '}
                            <Link
                                to="/docs"
                                className="text-primary-500 hover:underline"
                            >
                                Learn more
                            </Link>
                        </p>
                    </Alert>

                    {isLoading ? (
                        <div className="text-center py-8">
                            Loading subscriptions...
                        </div>
                    ) : subscriptions && subscriptions.length > 0 ? (
                        <div className="space-y-4">
                            {subscriptions.map((subscription) => (
                                <Card key={subscription.id}>
                                    <CardHeader>
                                        <div className="flex justify-between items-start">
                                            <div>
                                                <div className="flex items-center gap-2">
                                                    <h3 className="text-lg font-semibold">
                                                        {subscription.url}
                                                    </h3>
                                                    {subscription.is_active ? (
                                                        <span className="px-2 py-0.5 text-xs bg-success-100 text-success-700 rounded">
                                                            Active
                                                        </span>
                                                    ) : (
                                                        <span className="px-2 py-0.5 text-xs bg-muted text-muted-foreground rounded">
                                                            Inactive
                                                        </span>
                                                    )}
                                                </div>
                                                {subscription.description && (
                                                    <p className="text-sm text-muted-foreground mt-1">
                                                        {subscription.description}
                                                    </p>
                                                )}
                                            </div>
                                        </div>
                                    </CardHeader>
                                    <CardBody>
                                        <div className="space-y-3">
                                            <div>
                                                <p className="text-sm font-medium mb-1">
                                                    Subscribed Events:
                                                </p>
                                                <div className="flex flex-wrap gap-2">
                                                    {subscription.events.map(
                                                        (event) => (
                                                            <span
                                                                key={event}
                                                                className="px-2 py-1 text-xs bg-primary-100 text-primary-700 rounded"
                                                            >
                                                                {event}
                                                            </span>
                                                        )
                                                    )}
                                                </div>
                                            </div>
                                            <div className="text-xs text-muted-foreground">
                                                <p>
                                                    Created:{' '}
                                                    {new Date(
                                                        subscription.created_at
                                                    ).toLocaleString()}
                                                </p>
                                                {subscription.last_delivery_at && (
                                                    <p>
                                                        Last Delivery:{' '}
                                                        {new Date(
                                                            subscription.last_delivery_at
                                                        ).toLocaleString()}
                                                    </p>
                                                )}
                                            </div>
                                            <div className="flex flex-wrap gap-2 pt-2">
                                                <Button
                                                    variant="outline"
                                                    size="sm"
                                                    onClick={() =>
                                                        openEditModal(
                                                            subscription
                                                        )
                                                    }
                                                >
                                                    Edit
                                                </Button>
                                                <Button
                                                    variant="outline"
                                                    size="sm"
                                                    onClick={() =>
                                                        handleRegenerateSecret(
                                                            subscription
                                                        )
                                                    }
                                                >
                                                    Regenerate Secret
                                                </Button>
                                                <Button
                                                    variant="outline"
                                                    size="sm"
                                                    onClick={() =>
                                                        openDeliveriesModal(
                                                            subscription
                                                        )
                                                    }
                                                >
                                                    View Deliveries
                                                </Button>
                                                <Button
                                                    variant="outline"
                                                    size="sm"
                                                    onClick={() =>
                                                        openDeleteModal(
                                                            subscription
                                                        )
                                                    }
                                                    className="text-error-600 border-error-600 hover:bg-error-50"
                                                >
                                                    Delete
                                                </Button>
                                            </div>
                                        </div>
                                    </CardBody>
                                </Card>
                            ))}
                        </div>
                    ) : (
                        <Card>
                            <CardBody>
                                <div className="text-center py-8">
                                    <p className="text-muted-foreground mb-4">
                                        No webhook subscriptions yet
                                    </p>
                                    <Button
                                        variant="primary"
                                        onClick={() => setShowCreateModal(true)}
                                    >
                                        Create Your First Webhook
                                    </Button>
                                </div>
                            </CardBody>
                        </Card>
                    )}
                </div>
            </Container>

            {/* Create Modal */}
            <Modal
                open={showCreateModal}
                onClose={() => {
                    setShowCreateModal(false);
                    setCreateForm({ url: '', events: [], description: '' });
                    createMutation.reset();
                }}
                title="Create Webhook Subscription"
            >
                <form onSubmit={handleCreateSubmit}>
                    <Stack direction="vertical" gap={4}>
                        <Input
                            label="Webhook URL"
                            value={createForm.url}
                            onChange={(e) =>
                                setCreateForm({
                                    ...createForm,
                                    url: e.target.value,
                                })
                            }
                            placeholder="https://example.com/webhook"
                            required
                        />
                        <TextArea
                            label="Description (optional)"
                            value={createForm.description || ''}
                            onChange={(e) =>
                                setCreateForm({
                                    ...createForm,
                                    description: e.target.value,
                                })
                            }
                            rows={2}
                            placeholder="What is this webhook for?"
                        />
                        <div>
                            <label className="block text-sm font-medium mb-2">
                                Select Events *
                            </label>
                            <div className="space-y-2">
                                {supportedEvents && supportedEvents.length > 0 ? (
                                    supportedEvents.map((event) => (
                                        <Toggle
                                            key={event}
                                            label={event}
                                            checked={createForm.events.includes(
                                                event
                                            )}
                                            onChange={() => toggleCreateEvent(event)}
                                        />
                                    ))
                                ) : (
                                    <p className="text-sm text-muted-foreground">
                                        Loading events...
                                    </p>
                                )}
                            </div>
                        </div>
                        {createMutation.isError && (
                            <Alert variant="error">
                                <p className="font-medium">Failed to create webhook subscription</p>
                                {createMutation.error instanceof Error && createMutation.error.message && (
                                    <p className="mt-1 text-sm">{createMutation.error.message}</p>
                                )}
                            </Alert>
                        )}
                        <div className="flex gap-3 justify-end">
                            <Button
                                type="button"
                                variant="ghost"
                                onClick={() => setShowCreateModal(false)}
                            >
                                Cancel
                            </Button>
                            <Button
                                type="submit"
                                variant="primary"
                                disabled={
                                    createMutation.isPending ||
                                    createForm.events.length === 0
                                }
                            >
                                {createMutation.isPending
                                    ? 'Creating...'
                                    : 'Create Webhook'}
                            </Button>
                        </div>
                    </Stack>
                </form>
            </Modal>

            {/* Edit Modal */}
            <Modal
                open={showEditModal}
                onClose={() => {
                    setShowEditModal(false);
                    setSelectedSubscription(null);
                    setEditForm({});
                    updateMutation.reset();
                }}
                title="Edit Webhook Subscription"
            >
                <form onSubmit={handleEditSubmit}>
                    <Stack direction="vertical" gap={4}>
                        <Input
                            label="Webhook URL"
                            value={editForm.url || ''}
                            onChange={(e) =>
                                setEditForm({ ...editForm, url: e.target.value })
                            }
                            placeholder="https://example.com/webhook"
                        />
                        <TextArea
                            label="Description (optional)"
                            value={editForm.description || ''}
                            onChange={(e) =>
                                setEditForm({
                                    ...editForm,
                                    description: e.target.value,
                                })
                            }
                            rows={2}
                        />
                        <Toggle
                            label="Active"
                            checked={editForm.is_active ?? true}
                            onChange={(e) =>
                                setEditForm({
                                    ...editForm,
                                    is_active: e.target.checked,
                                })
                            }
                        />
                        <div>
                            <label className="block text-sm font-medium mb-2">
                                Select Events
                            </label>
                            <div className="space-y-2">
                                {supportedEvents && supportedEvents.length > 0 ? (
                                    supportedEvents.map((event) => (
                                        <Toggle
                                            key={event}
                                            label={event}
                                            checked={
                                                editForm.events?.includes(event) ??
                                                false
                                            }
                                            onChange={() => toggleEditEvent(event)}
                                        />
                                    ))
                                ) : (
                                    <p className="text-sm text-muted-foreground">
                                        Loading events...
                                    </p>
                                )}
                            </div>
                        </div>
                        {updateMutation.isError && (
                            <Alert variant="error">
                                <p className="font-medium">Failed to update webhook subscription</p>
                                {updateMutation.error instanceof Error && updateMutation.error.message && (
                                    <p className="mt-1 text-sm">{updateMutation.error.message}</p>
                                )}
                            </Alert>
                        )}
                        <div className="flex gap-3 justify-end">
                            <Button
                                type="button"
                                variant="ghost"
                                onClick={() => setShowEditModal(false)}
                            >
                                Cancel
                            </Button>
                            <Button
                                type="submit"
                                variant="primary"
                                disabled={updateMutation.isPending}
                            >
                                {updateMutation.isPending
                                    ? 'Saving...'
                                    : 'Save Changes'}
                            </Button>
                        </div>
                    </Stack>
                </form>
            </Modal>

            {/* Delete Confirmation Modal */}
            <Modal
                open={showDeleteModal}
                onClose={() => {
                    setShowDeleteModal(false);
                    setSelectedSubscription(null);
                    deleteMutation.reset();
                }}
                title="Delete Webhook Subscription"
            >
                <Stack direction="vertical" gap={4}>
                    <Alert variant="warning">
                        Are you sure you want to delete this webhook subscription?
                        This action cannot be undone.
                    </Alert>
                    {selectedSubscription && (
                        <div className="text-sm">
                            <p className="font-medium">URL:</p>
                            <p className="text-muted-foreground">
                                {selectedSubscription.url}
                            </p>
                        </div>
                    )}
                    {deleteMutation.isError && (
                        <Alert variant="error">
                            <p className="font-medium">Failed to delete webhook subscription</p>
                            {deleteMutation.error instanceof Error && deleteMutation.error.message && (
                                <p className="mt-1 text-sm">{deleteMutation.error.message}</p>
                            )}
                        </Alert>
                    )}
                    <div className="flex gap-3 justify-end">
                        <Button
                            variant="ghost"
                            onClick={() => setShowDeleteModal(false)}
                        >
                            Cancel
                        </Button>
                        <Button
                            variant="primary"
                            onClick={handleDelete}
                            disabled={deleteMutation.isPending}
                            className="bg-error-600 hover:bg-error-700"
                        >
                            {deleteMutation.isPending ? 'Deleting...' : 'Delete'}
                        </Button>
                    </div>
                </Stack>
            </Modal>

            {/* Secret Display Modal */}
            <Modal
                open={showSecretModal}
                onClose={() => {
                    setShowSecretModal(false);
                    setNewSecret(null);
                    setCopiedSecret(false);
                }}
                title="Webhook Secret"
            >
                <Stack direction="vertical" gap={4}>
                    <Alert variant="warning">
                        <strong>Important:</strong> Save this secret securely. It
                        will only be shown once. You'll need it to verify webhook
                        signatures.
                    </Alert>
                    <div>
                        <label className="block text-sm font-medium mb-2">
                            Secret
                        </label>
                        <div className="flex gap-2">
                            <Input
                                value={newSecret || ''}
                                readOnly
                                className="font-mono text-sm"
                            />
                            <Button
                                variant="outline"
                                onClick={copySecretToClipboard}
                            >
                                {copiedSecret ? 'Copied!' : 'Copy'}
                            </Button>
                        </div>
                    </div>
                    <div className="flex gap-3 justify-end">
                        <Button
                            variant="primary"
                            onClick={() => {
                                setShowSecretModal(false);
                                setNewSecret(null);
                                setCopiedSecret(false);
                            }}
                        >
                            I've Saved It
                        </Button>
                    </div>
                </Stack>
            </Modal>

            {/* Deliveries Modal */}
            {selectedSubscription && (
                <DeliveriesModal
                    open={showDeliveriesModal}
                    onClose={() => setShowDeliveriesModal(false)}
                    subscriptionId={selectedSubscription.id}
                />
            )}
        </>
    );
}

// Deliveries Modal Component
function DeliveriesModal({
    open,
    onClose,
    subscriptionId,
}: {
    open: boolean;
    onClose: () => void;
    subscriptionId: string;
}) {
    const [page, setPage] = useState(1);

    // Reset page when subscriptionId changes
    useEffect(() => {
        queueMicrotask(() => {
            setPage(1);
        });
    }, [subscriptionId]);

    const { data, isLoading } = useQuery({
        queryKey: ['webhookDeliveries', subscriptionId, page],
        queryFn: () => getWebhookDeliveries(subscriptionId, page, 20),
        enabled: open,
    });

    const getStatusBadge = (status: string) => {
        const colors = {
            delivered: 'bg-success-100 text-success-700',
            pending: 'bg-warning-100 text-warning-700',
            failed: 'bg-error-100 text-error-700',
        };
        return colors[status as keyof typeof colors] || 'bg-muted text-muted-foreground';
    };

    return (
        <Modal open={open} onClose={onClose} title="Delivery History">
            <div className="space-y-4">
                {isLoading ? (
                    <div className="text-center py-8">Loading deliveries...</div>
                ) : data && data.deliveries.length > 0 ? (
                    <>
                        <div className="space-y-3 max-h-96 overflow-y-auto">
                            {data.deliveries.map((delivery: WebhookDelivery) => (
                                <div
                                    key={delivery.id}
                                    className="p-3 border border-border rounded-md"
                                >
                                    <div className="flex justify-between items-start mb-2">
                                        <div>
                                            <span className="font-medium text-sm">
                                                {delivery.event_type}
                                            </span>
                                            <span
                                                className={`ml-2 px-2 py-0.5 text-xs rounded ${getStatusBadge(
                                                    delivery.status
                                                )}`}
                                            >
                                                {delivery.status}
                                            </span>
                                        </div>
                                        {delivery.http_status_code && (
                                            <span className="text-xs text-muted-foreground">
                                                HTTP {delivery.http_status_code}
                                            </span>
                                        )}
                                    </div>
                                    <p className="text-xs text-muted-foreground">
                                        {new Date(
                                            delivery.created_at
                                        ).toLocaleString()}{' '}
                                        • Attempt {delivery.attempt_count}/
                                        {delivery.max_attempts}
                                    </p>
                                    {delivery.error_message && (
                                        <p className="text-xs text-error-600 mt-1">
                                            {delivery.error_message}
                                        </p>
                                    )}
                                </div>
                            ))}
                        </div>
                        {data.meta.total_pages > 1 && (
                            <div className="flex justify-between items-center pt-2">
                                <Button
                                    variant="outline"
                                    size="sm"
                                    onClick={() => setPage(page - 1)}
                                    disabled={page === 1}
                                >
                                    Previous
                                </Button>
                                <span className="text-sm text-muted-foreground">
                                    Page {page} of {data.meta.total_pages}
                                </span>
                                <Button
                                    variant="outline"
                                    size="sm"
                                    onClick={() => setPage(page + 1)}
                                    disabled={page === data.meta.total_pages}
                                >
                                    Next
                                </Button>
                            </div>
                        )}
                    </>
                ) : (
                    <div className="text-center py-8 text-muted-foreground">
                        No deliveries yet
                    </div>
                )}
            </div>
        </Modal>
    );
}
