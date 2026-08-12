import { useCallback, useEffect, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import {
    Alert,
    Badge,
    Button,
    Card,
    Container,
    Modal,
    Spinner,
    TextArea,
} from '../../components';
import { useAuth } from '../../context/AuthContext';
import {
    getVerificationApplications,
    getVerificationStats,
    reviewVerificationApplication,
    type VerificationApplication,
    type VerificationApplicationStats,
} from '../../lib/verification-api';

// Helper function to extract error message from API error
function getErrorMessage(error: unknown): string {
    if (typeof error === 'object' && error !== null && 'response' in error) {
        const apiError = error as { response?: { data?: { error?: string | { message?: string }; message?: string } } };
        const responseError = apiError.response?.data?.error;
        if (typeof responseError === 'string') return responseError;
        return responseError?.message || apiError.response?.data?.message || 'The review could not be saved.';
    }
    return error instanceof Error ? error.message : 'The review could not be saved.';
}

export function AdminVerificationQueuePage() {
    const { isAuthenticated, isAdmin } = useAuth();
    const navigate = useNavigate();
    const [applications, setApplications] = useState<VerificationApplication[]>([]);
    const [stats, setStats] = useState<VerificationApplicationStats | null>(null);
    const [isLoading, setIsLoading] = useState(true);
    const [error, setError] = useState<string | null>(null);
    const [success, setSuccess] = useState<string | null>(null);
    
    // Filters
    const [statusFilter, setStatusFilter] = useState<string>('pending');
    
    // Review modal state
    const [reviewModalOpen, setReviewModalOpen] = useState(false);
    const [selectedApplication, setSelectedApplication] = useState<VerificationApplication | null>(null);
    const [reviewNotes, setReviewNotes] = useState('');
    const [reviewDecision, setReviewDecision] = useState<'approved' | 'rejected'>('approved');
    const [isReviewing, setIsReviewing] = useState(false);
    const [reviewError, setReviewError] = useState<string | null>(null);

    const loadApplications = useCallback(async () => {
        try {
            setIsLoading(true);
            setError(null);
            const response = await getVerificationApplications(statusFilter, 50);
            setApplications(response.data || []);
        } catch (err: unknown) {
            setError(getErrorMessage(err));
        } finally {
            setIsLoading(false);
        }
    }, [statusFilter]);

    const loadStats = useCallback(async () => {
        try {
            const response = await getVerificationStats();
            setStats(response.data);
        } catch (err: unknown) {
            console.error('Failed to load stats:', err);
        }
    }, []);

    useEffect(() => {
        if (!isAuthenticated || !isAdmin) {
            navigate('/');
            return;
        }

        loadApplications();
        loadStats();
    }, [isAuthenticated, isAdmin, navigate, loadApplications, loadStats]);

    // Auto-dismiss success messages after 5 seconds
    useEffect(() => {
        if (success) {
            const timer = setTimeout(() => {
                setSuccess(null);
            }, 5000);
            return () => clearTimeout(timer);
        }
    }, [success]);

    const openReviewModal = (app: VerificationApplication, decision: 'approved' | 'rejected') => {
        setSelectedApplication(app);
        setReviewDecision(decision);
        setReviewNotes('');
        setReviewError(null);
        setReviewModalOpen(true);
    };

    const handleReview = async () => {
        if (!selectedApplication) {
            return;
        }

        try {
            setIsReviewing(true);
            setReviewError(null);
            await reviewVerificationApplication(selectedApplication.id, {
                decision: reviewDecision,
                notes: reviewNotes || undefined,
            });
            setSuccess(`Application ${reviewDecision} successfully!`);
            setReviewModalOpen(false);
            setSelectedApplication(null);
            setReviewNotes('');
            await Promise.all([loadApplications(), loadStats()]);
        } catch (err: unknown) {
            setReviewError(getErrorMessage(err));
        } finally {
            setIsReviewing(false);
        }
    };

    const getStatusBadgeColor = (status: string) => {
        switch (status) {
            case 'approved':
                return 'green';
            case 'rejected':
                return 'red';
            case 'pending':
            default:
                return 'yellow';
        }
    };

    const formatDate = (dateString: string) => {
        return new Date(dateString).toLocaleDateString('en-US', {
            year: 'numeric',
            month: 'short',
            day: 'numeric',
            hour: '2-digit',
            minute: '2-digit',
        });
    };

    return (
        <Container maxWidth="full" className="px-0">
            <div>
                <div className="mb-6">
                    <p className="mb-2 text-xs font-bold uppercase tracking-[0.16em] text-brand">Trust &amp; identity</p>
                    <h1 className="text-3xl font-bold text-text-primary mb-2">
                        Creator Verification Queue
                    </h1>
                    <p className="text-text-secondary">
                        Review and manage creator verification applications
                    </p>
                </div>

                {/* Stats Dashboard */}
                {stats && (
                    <div className="grid grid-cols-1 md:grid-cols-4 gap-4 mb-6">
                        <Card className="p-4">
                            <div className="text-sm text-gray-600 dark:text-gray-400">
                                Pending
                            </div>
                            <div className="text-2xl font-bold text-yellow-600 dark:text-yellow-500">
                                {stats.total_pending}
                            </div>
                        </Card>
                        <Card className="p-4">
                            <div className="text-sm text-gray-600 dark:text-gray-400">
                                Approved
                            </div>
                            <div className="text-2xl font-bold text-green-600 dark:text-green-500">
                                {stats.total_approved}
                            </div>
                        </Card>
                        <Card className="p-4">
                            <div className="text-sm text-gray-600 dark:text-gray-400">
                                Rejected
                            </div>
                            <div className="text-2xl font-bold text-red-600 dark:text-red-500">
                                {stats.total_rejected}
                            </div>
                        </Card>
                        <Card className="p-4">
                            <div className="text-sm text-gray-600 dark:text-gray-400">
                                Total Verified
                            </div>
                            <div className="text-2xl font-bold text-blue-600 dark:text-blue-500">
                                {stats.total_verified}
                            </div>
                        </Card>
                    </div>
                )}

                {/* Status Filter */}
                <div className="flex gap-4 mb-6">
                    <div className="flex gap-2">
                        <Button
                            size="sm"
                            variant={statusFilter === 'pending' ? 'primary' : 'secondary'}
                            onClick={() => setStatusFilter('pending')}
                        >
                            Pending
                        </Button>
                        <Button
                            size="sm"
                            variant={statusFilter === 'approved' ? 'primary' : 'secondary'}
                            onClick={() => setStatusFilter('approved')}
                        >
                            Approved
                        </Button>
                        <Button
                            size="sm"
                            variant={statusFilter === 'rejected' ? 'primary' : 'secondary'}
                            onClick={() => setStatusFilter('rejected')}
                        >
                            Rejected
                        </Button>
                    </div>
                </div>

                {/* Error and Success Messages */}
                {error && (
                    <Alert
                        variant="error"
                        className="mb-4"
                        onDismiss={() => setError(null)}
                    >
                        {error}
                    </Alert>
                )}
                {success && (
                    <Alert
                        variant="success"
                        className="mb-4"
                        onDismiss={() => setSuccess(null)}
                    >
                        {success}
                    </Alert>
                )}

                {/* Loading State */}
                {isLoading && (
                    <div className="flex justify-center py-12">
                        <Spinner size="lg" />
                    </div>
                )}

                {/* Applications List */}
                {!isLoading && applications.length === 0 && (
                    <Card className="p-8 text-center text-gray-600 dark:text-gray-400">
                        No {statusFilter} applications found.
                    </Card>
                )}

                {!isLoading && applications.length > 0 && (
                    <div className="space-y-4">
                        {applications.map((app) => (
                            <Card key={app.id} className="p-6">
                                <div className="flex flex-col items-start justify-between gap-5 lg:flex-row">
                                    <div className="flex-1">
                                        <div className="flex flex-wrap items-center gap-3 mb-2">
                                            <h3 className="text-lg font-semibold text-gray-900 dark:text-white">
                                                Application #{app.id.slice(0, 8)}
                                            </h3>
                                            <Badge color={getStatusBadgeColor(app.status)}>
                                                {app.status}
                                            </Badge>
                                            {app.priority > 70 && (
                                                <Badge color="red">High Priority</Badge>
                                            )}
                                        </div>
                                        
                                        <div className="space-y-2 text-sm text-gray-600 dark:text-gray-400">
                                            <div>
                                                <span className="font-medium">Twitch Channel:</span>{' '}
                                                <a
                                                    href={app.twitch_channel_url}
                                                    target="_blank"
                                                    rel="noopener noreferrer"
                                                    className="text-blue-600 hover:underline"
                                                >
                                                    {app.twitch_channel_url}
                                                </a>
                                            </div>
                                            {app.follower_count && (
                                                <div>
                                                    <span className="font-medium">Followers:</span>{' '}
                                                    {app.follower_count.toLocaleString()}
                                                </div>
                                            )}
                                            {app.subscriber_count && (
                                                <div>
                                                    <span className="font-medium">Subscribers:</span>{' '}
                                                    {app.subscriber_count.toLocaleString()}
                                                </div>
                                            )}
                                            {app.avg_viewers && (
                                                <div>
                                                    <span className="font-medium">Avg Viewers:</span>{' '}
                                                    {app.avg_viewers.toLocaleString()}
                                                </div>
                                            )}
                                            {app.content_description && (
                                                <div>
                                                    <span className="font-medium">Description:</span>{' '}
                                                    {app.content_description}
                                                </div>
                                            )}
                                            <div>
                                                <span className="font-medium">Submitted:</span>{' '}
                                                {formatDate(app.created_at)}
                                            </div>
                                            {app.reviewed_at && (
                                                <div>
                                                    <span className="font-medium">Reviewed:</span>{' '}
                                                    {formatDate(app.reviewed_at)}
                                                </div>
                                            )}
                                            {app.reviewer_notes && (
                                                <div>
                                                    <span className="font-medium">Notes:</span>{' '}
                                                    {app.reviewer_notes}
                                                </div>
                                            )}
                                        </div>
                                    </div>

                                    {app.status === 'pending' && (
                                        <div className="flex w-full gap-2 lg:ml-4 lg:w-auto">
                                            <Button
                                                size="sm"
                                                variant="primary"
                                                onClick={() => openReviewModal(app, 'approved')}
                                                className="flex-1 lg:flex-none"
                                            >
                                                Approve
                                            </Button>
                                            <Button
                                                size="sm"
                                                variant="danger"
                                                onClick={() => openReviewModal(app, 'rejected')}
                                                className="flex-1 lg:flex-none"
                                            >
                                                Reject
                                            </Button>
                                        </div>
                                    )}
                                </div>
                            </Card>
                        ))}
                    </div>
                )}

                {/* Review Modal */}
                <Modal
                    open={reviewModalOpen}
                    onClose={() => { if (!isReviewing) setReviewModalOpen(false); }}
                    title={`${reviewDecision === 'approved' ? 'Approve' : 'Reject'} Application`}
                >
                    <div className="space-y-4">
                        {reviewError && <Alert variant="error" onDismiss={() => setReviewError(null)}>{reviewError}</Alert>}
                        <p className="text-gray-600 dark:text-gray-400">
                            {reviewDecision === 'approved'
                                ? 'Are you sure you want to approve this verification application? The user will receive a verified badge.'
                                : 'Are you sure you want to reject this verification application? Please provide a reason.'}
                        </p>

                        {selectedApplication && (
                            <div className="bg-gray-100 dark:bg-gray-800 p-4 rounded-md text-sm">
                                <div><strong>Channel:</strong> {selectedApplication.twitch_channel_url}</div>
                                {selectedApplication.follower_count && (
                                    <div><strong>Followers:</strong> {selectedApplication.follower_count.toLocaleString()}</div>
                                )}
                            </div>
                        )}

                        <div>
                            <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                                Notes (Optional)
                            </label>
                            <TextArea
                                value={reviewNotes}
                                onChange={(e) => setReviewNotes(e.target.value)}
                                placeholder={reviewDecision === 'rejected' ? 'Reason for rejection...' : 'Additional notes...'}
                                rows={4}
                            />
                        </div>

                        <div className="flex justify-end gap-3">
                            <Button
                                variant="secondary"
                                onClick={() => setReviewModalOpen(false)}
                                disabled={isReviewing}
                            >
                                Cancel
                            </Button>
                            <Button
                                variant={reviewDecision === 'approved' ? 'primary' : 'danger'}
                                onClick={handleReview}
                                loading={isReviewing}
                            >
                                {isReviewing ? 'Saving decision' : reviewDecision === 'approved' ? 'Approve' : 'Reject'}
                            </Button>
                        </div>
                    </div>
                </Modal>
            </div>
        </Container>
    );
}
