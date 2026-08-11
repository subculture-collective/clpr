import { useEffect, useState } from 'react';
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
    UserRoleBadge,
} from '../../components';
import { useAuth } from '../../context/AuthContext';
import {
    approveSubmission,
    getPendingSubmissions,
    rejectSubmission,
} from '../../lib/submission-api';
import type { ClipSubmissionWithUser } from '../../types/submission';
import type { UserRole } from '../../lib/roles';

export function ModerationQueuePage() {
    const { isAuthenticated, isModeratorOrAdmin } = useAuth();
    const navigate = useNavigate();
    const [submissions, setSubmissions] = useState<ClipSubmissionWithUser[]>(
        []
    );
    const [isLoading, setIsLoading] = useState(true);
    const [error, setError] = useState<string | null>(null);
    const [success, setSuccess] = useState<string | null>(null);
    const [page, setPage] = useState(1);
    const [totalPages, setTotalPages] = useState(1);
    const [total, setTotal] = useState(0);

    // Rejection modal state
    const [rejectModalOpen, setRejectModalOpen] = useState(false);
    const [selectedSubmissionId, setSelectedSubmissionId] = useState<
        string | null
    >(null);
    const [rejectionReason, setRejectionReason] = useState('');

    const loadSubmissions = async () => {
        try {
            setIsLoading(true);
            setError(null);
            const response = await getPendingSubmissions(page, 20);
            setSubmissions(response.data || []);
            setTotalPages(response.meta.total_pages);
            setTotal(response.meta.total);
        } catch (err: unknown) {
            const error = err as { response?: { data?: { error?: string } } };
            setError(
                error.response?.data?.error || 'Failed to load submissions'
            );
        } finally {
            setIsLoading(false);
        }
    };

    useEffect(() => {
        if (!isAuthenticated || !isModeratorOrAdmin) {
            navigate('/');
            return;
        }

        loadSubmissions();
        // eslint-disable-next-line react-hooks/exhaustive-deps
    }, [isAuthenticated, isModeratorOrAdmin, navigate, page]);

    const handleApprove = async (submissionId: string) => {
        try {
            await approveSubmission(submissionId);
            setSuccess('Submission approved successfully!');
            loadSubmissions(); // Reload the list
        } catch (err: unknown) {
            const error = err as { response?: { data?: { error?: string } } };
            setError(
                error.response?.data?.error || 'Failed to approve submission'
            );
        }
    };

    const openRejectModal = (submissionId: string) => {
        setSelectedSubmissionId(submissionId);
        setRejectModalOpen(true);
        setRejectionReason('');
    };

    const handleReject = async () => {
        if (!selectedSubmissionId || !rejectionReason.trim()) {
            setError('Rejection reason is required');
            return;
        }

        try {
            await rejectSubmission(selectedSubmissionId, rejectionReason);
            setSuccess('Submission rejected successfully!');
            setRejectModalOpen(false);
            setSelectedSubmissionId(null);
            setRejectionReason('');
            loadSubmissions(); // Reload the list
        } catch (err: unknown) {
            const error = err as { response?: { data?: { error?: string } } };
            setError(
                error.response?.data?.error || 'Failed to reject submission'
            );
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
                        Review and moderate user-submitted clips
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
                <Card className='p-4 mb-6'>
                    <div className='flex items-center justify-between'>
                        <div>
                            <div className='text-2xl font-bold'>{total}</div>
                            <div className='text-muted-foreground text-sm'>
                                Pending Submissions
                            </div>
                        </div>
                        <Button
                            onClick={loadSubmissions}
                            variant='secondary'
                            disabled={isLoading}
                        >
                            Refresh
                        </Button>
                    </div>
                </Card>

                {/* Submissions List */}
                <Card className='p-6'>
                    {isLoading ? (
                        <div className='flex justify-center py-12'>
                            <Spinner size='lg' />
                        </div>
                    ) : !submissions || submissions.length === 0 ? (
                        <div className='py-12 text-center'>
                            <p className='text-muted-foreground'>
                                No pending submissions to review.
                            </p>
                        </div>
                    ) : (
                        <div className='space-y-6'>
                            {submissions.map((submission) => (
                                <div
                                    key={submission.id}
                                    className='bg-background-secondary p-6 rounded-lg'
                                >
                                    <div className='lg:flex-row flex flex-col gap-4'>
                                        {/* Thumbnail and Preview */}
                                        <div className='shrink-0'>
                                            {submission.thumbnail_url ? (
                                                <img
                                                    src={
                                                        submission.thumbnail_url
                                                    }
                                                    alt={
                                                        submission.title ||
                                                        'Clip thumbnail'
                                                    }
                                                    className='lg:w-64 object-cover w-full h-40 rounded'
                                                />
                                            ) : (
                                                <div className='lg:w-64 bg-background-tertiary flex items-center justify-center w-full h-40 rounded'>
                                                    <span className='text-muted-foreground'>
                                                        No thumbnail
                                                    </span>
                                                </div>
                                            )}
                                            {submission.is_nsfw && (
                                                <Badge
                                                    variant='error'
                                                    className='mt-2'
                                                >
                                                    NSFW
                                                </Badge>
                                            )}
                                        </div>

                                        {/* Content */}
                                        <div className='flex-1'>
                                            <div className='mb-4'>
                                                <h3 className='mb-2 text-xl font-bold'>
                                                    {submission.custom_title ||
                                                        submission.title ||
                                                        'Untitled'}
                                                </h3>

                                                {/* Submitter Info */}
                                                {submission.user && (
                                                    <div className='text-muted-foreground flex items-center gap-2 mb-2 text-sm'>
                                                        <span>
                                                            Submitted by
                                                        </span>
                                                        <span className='text-foreground font-medium'>
                                                            {submission.user
                                                                .display_name ||
                                                                submission.user
                                                                    .username}
                                                        </span>
                                                        <Badge variant='default'>
                                                            {
                                                                submission.user
                                                                    .karma_points
                                                            }{' '}
                                                            uppies
                                                        </Badge>
                                                        {submission.user.role !== 'user' && (
                                                            <UserRoleBadge role={submission.user.role as UserRole} size="sm" />
                                                        )}
                                                    </div>
                                                )}

                                                {/* Clip Info */}
                                                <div className='text-muted-foreground space-y-1 text-sm'>
                                                    {submission.broadcaster_name && (
                                                        <p>
                                                            Broadcaster:{' '}
                                                            {
                                                                submission.broadcaster_name
                                                            }
                                                        </p>
                                                    )}
                                                    {submission.broadcaster_name_override && (
                                                        <p className='text-warning font-medium'>
                                                            Broadcaster Override:{' '}
                                                            {
                                                                submission.broadcaster_name_override
                                                            }
                                                        </p>
                                                    )}
                                                    {submission.creator_name && (
                                                        <p>
                                                            Creator:{' '}
                                                            {
                                                                submission.creator_name
                                                            }
                                                        </p>
                                                    )}
                                                    {submission.game_name && (
                                                        <p>
                                                            Game:{' '}
                                                            {
                                                                submission.game_name
                                                            }
                                                        </p>
                                                    )}
                                                    {submission.duration && (
                                                        <p>
                                                            Duration:{' '}
                                                            {submission.duration.toFixed(
                                                                1
                                                            )}
                                                            s
                                                        </p>
                                                    )}
                                                    <p>
                                                        Views:{' '}
                                                        {submission.view_count}
                                                    </p>
                                                    <p>
                                                        Submitted:{' '}
                                                        {new Date(
                                                            submission.created_at
                                                        ).toLocaleString()}
                                                    </p>
                                                </div>

                                                {/* Tags */}
                                                {submission.tags &&
                                                    submission.tags.length >
                                                        0 && (
                                                        <div className='flex flex-wrap gap-2 mt-3'>
                                                            {submission.tags.map(
                                                                (tag) => (
                                                                    <span
                                                                        key={
                                                                            tag
                                                                        }
                                                                        className='bg-primary/10 text-primary px-2 py-1 text-xs rounded'
                                                                    >
                                                                        {tag}
                                                                    </span>
                                                                )
                                                            )}
                                                        </div>
                                                    )}

                                                {/* Submission Reason */}
                                                {submission.submission_reason && (
                                                    <div className='bg-background-tertiary p-3 mt-3 rounded'>
                                                        <p className='mb-1 text-sm font-medium'>
                                                            Submission Reason:
                                                        </p>
                                                        <p className='text-muted-foreground text-sm'>
                                                            {
                                                                submission.submission_reason
                                                            }
                                                        </p>
                                                    </div>
                                                )}
                                            </div>

                                            {/* Actions */}
                                            <div className='flex gap-3'>
                                                <Button
                                                    onClick={() =>
                                                        handleApprove(
                                                            submission.id
                                                        )
                                                    }
                                                    variant='primary'
                                                    className='hover:bg-green-700 flex-1 bg-green-600'
                                                >
                                                    Approve
                                                </Button>
                                                <Button
                                                    onClick={() =>
                                                        openRejectModal(
                                                            submission.id
                                                        )
                                                    }
                                                    variant='secondary'
                                                    className='bg-red-600/20 hover:bg-red-600/30 flex-1 text-red-500'
                                                >
                                                    Reject
                                                </Button>
                                                <Button
                                                    onClick={() =>
                                                        window.open(
                                                            submission.twitch_clip_url,
                                                            '_blank'
                                                        )
                                                    }
                                                    variant='secondary'
                                                >
                                                    View on Twitch
                                                </Button>
                                            </div>
                                        </div>
                                    </div>
                                </div>
                            ))}
                        </div>
                    )}

                    {/* Pagination */}
                    {totalPages > 1 && (
                        <div className='flex justify-center gap-2 mt-6'>
                            <Button
                                variant='secondary'
                                onClick={() => setPage(page - 1)}
                                disabled={page === 1}
                            >
                                Previous
                            </Button>
                            <span className='flex items-center px-4'>
                                Page {page} of {totalPages}
                            </span>
                            <Button
                                variant='secondary'
                                onClick={() => setPage(page + 1)}
                                disabled={page === totalPages}
                            >
                                Next
                            </Button>
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
                        setSelectedSubmissionId(null);
                        setRejectionReason('');
                    }}
                    title='Reject Submission'
                >
                    <div className='space-y-4'>
                        <p className='text-muted-foreground'>
                            Please provide a reason for rejecting this
                            submission. This will be shown to the user.
                        </p>
                        <div>
                            <label htmlFor='rejection-reason' className='block mb-2 text-sm font-medium'>
                                Rejection Reason
                            </label>
                            <TextArea
                                id='rejection-reason'
                                value={rejectionReason}
                                onChange={(e) => setRejectionReason(e.target.value)}
                                placeholder='Reason for rejection...'
                                rows={4}
                                required
                            />
                        </div>
                        <div className='flex gap-3'>
                            <Button
                                onClick={handleReject}
                                disabled={!rejectionReason.trim()}
                                className='hover:bg-red-700 flex-1 bg-red-600'
                            >
                                Reject Submission
                            </Button>
                            <Button
                                onClick={() => {
                                    setRejectModalOpen(false);
                                    setSelectedSubmissionId(null);
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
        </Container>
    );
}
