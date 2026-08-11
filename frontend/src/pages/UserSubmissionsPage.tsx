import { useEffect, useState } from 'react';
import { useNavigate, Link } from 'react-router-dom';
import { Alert, Badge, Button, Card, Container, Spinner } from '../components';
import { useAuth } from '../context/AuthContext';
import { getSubmissionStats, getUserSubmissions } from '../lib/submission-api';
import type { ClipSubmission, SubmissionStats } from '../types/submission';

export function UserSubmissionsPage() {
    const { isAuthenticated } = useAuth();
    const navigate = useNavigate();
    const [submissions, setSubmissions] = useState<ClipSubmission[]>([]);
    const [stats, setStats] = useState<SubmissionStats | null>(null);
    const [isLoading, setIsLoading] = useState(true);
    const [error, setError] = useState<string | null>(null);
    const [page, setPage] = useState(1);
    const [totalPages, setTotalPages] = useState(1);

    const loadSubmissions = async () => {
        try {
            setIsLoading(true);
            const response = await getUserSubmissions(page, 20);
            setSubmissions(response.data);
            setTotalPages(response.meta.total_pages);
        } catch (err: unknown) {
            const error = err as { response?: { data?: { error?: string } } };
            setError(
                error.response?.data?.error || 'Failed to load submissions'
            );
        } finally {
            setIsLoading(false);
        }
    };

    const loadStats = async () => {
        try {
            const response = await getSubmissionStats();
            setStats(response.data);
        } catch (err) {
            console.error('Failed to load stats:', err);
        }
    };

    useEffect(() => {
        if (!isAuthenticated) {
            navigate('/login');
            return;
        }

        loadSubmissions();
        loadStats();
        // eslint-disable-next-line react-hooks/exhaustive-deps
    }, [isAuthenticated, navigate, page]);

    const getStatusColor = (status: string) => {
        switch (status) {
            case 'approved':
                return 'success';
            case 'rejected':
                return 'error';
            case 'pending':
                return 'warning';
            default:
                return 'default';
        }
    };

    if (!isAuthenticated) {
        return null;
    }

    return (
        <Container className='py-8'>
            <div className='max-w-6xl mx-auto'>
                <div className='flex items-center justify-between mb-6'>
                    <div>
                        <h1 className='mb-2 text-3xl font-bold'>
                            My Submissions
                        </h1>
                        <p className='text-muted-foreground'>
                            Track the status of your submitted clips
                        </p>
                    </div>
                    <Button onClick={() => navigate('/submit')}>
                        Submit New Clip
                    </Button>
                </div>

                {error && (
                    <Alert
                        variant='error'
                        className='mb-6'
                    >
                        {error}
                    </Alert>
                )}

                {/* Stats Cards */}
                {stats && (
                    <div className='md:grid-cols-4 grid grid-cols-1 gap-4 mb-8'>
                        <Card className='p-4'>
                            <div className='text-2xl font-bold'>
                                {stats.total_submissions}
                            </div>
                            <div className='text-muted-foreground text-sm'>
                                Total Submissions
                            </div>
                        </Card>
                        <Card className='p-4'>
                            <div className='text-2xl font-bold text-green-500'>
                                {stats.approved_count}
                            </div>
                            <div className='text-muted-foreground text-sm'>
                                Approved
                            </div>
                        </Card>
                        <Card className='p-4'>
                            <div className='text-2xl font-bold text-yellow-500'>
                                {stats.pending_count}
                            </div>
                            <div className='text-muted-foreground text-sm'>
                                Pending
                            </div>
                        </Card>
                        <Card className='p-4'>
                            <div className='text-2xl font-bold'>
                                {stats.approval_rate.toFixed(1)}%
                            </div>
                            <div className='text-muted-foreground text-sm'>
                                Approval Rate
                            </div>
                        </Card>
                    </div>
                )}

                {/* Submissions List */}
                <Card className='p-6'>
                    {isLoading ? (
                        <div className='flex justify-center py-12'>
                            <Spinner size='lg' />
                        </div>
                    ) : submissions.length === 0 ? (
                        <div className='py-12 text-center'>
                            <p className='text-muted-foreground mb-4'>
                                You haven't submitted any clips yet.
                            </p>
                            <Button onClick={() => navigate('/submit')}>
                                Submit Your First Clip
                            </Button>
                        </div>
                    ) : (
                        <div className='space-y-4'>
                            {submissions.map((submission) => {
                                // Only approved submissions can link to a clip page
                                const isApproved = submission.status === 'approved' && submission.clip_id;
                                const clipUrl = isApproved ? `/clip/${submission.clip_id}` : null;

                                const cardContent = (
                                    <>
                                        {/* Thumbnail */}
                                        {submission.thumbnail_url && (
                                            <div className='shrink-0'>
                                                <img
                                                    src={submission.thumbnail_url}
                                                    alt={
                                                        submission.title ||
                                                        'Clip thumbnail'
                                                    }
                                                    className='md:w-48 object-cover w-full h-32 rounded'
                                                />
                                            </div>
                                        )}

                                    {/* Content */}
                                    <div className='flex-1 min-w-0'>
                                        <div className='flex items-start justify-between gap-2 mb-2'>
                                            <h3 className='text-lg font-semibold truncate hover:underline'>
                                                {submission.custom_title ||
                                                    submission.title ||
                                                    'Untitled'}
                                            </h3>
                                            <Badge
                                                variant={getStatusColor(
                                                    submission.status
                                                )}
                                            >
                                                {submission.status}
                                            </Badge>
                                        </div>

                                        <div className='text-muted-foreground space-y-1 text-sm'>
                                            {submission.broadcaster_name && (
                                                <p>
                                                    Broadcaster:{' '}
                                                    {
                                                        submission.broadcaster_name
                                                    }
                                                </p>
                                            )}
                                            {submission.game_name && (
                                                <p>
                                                    Twitch category: {submission.game_name}
                                                </p>
                                            )}
                                            <p>
                                                Submitted:{' '}
                                                {new Date(
                                                    submission.created_at
                                                ).toLocaleDateString()}
                                            </p>
                                            {submission.reviewed_at && (
                                                <p>
                                                    Reviewed:{' '}
                                                    {new Date(
                                                        submission.reviewed_at
                                                    ).toLocaleDateString()}
                                                </p>
                                            )}
                                        </div>

                                        {/* Tags */}
                                        {submission.tags &&
                                            submission.tags.length > 0 && (
                                                <div className='flex flex-wrap gap-2 mt-2'>
                                                    {submission.tags.map(
                                                        (tag) => (
                                                            <span
                                                                key={tag}
                                                                className='bg-primary/10 text-primary px-2 py-1 text-xs rounded'
                                                            >
                                                                {tag}
                                                            </span>
                                                        )
                                                    )}
                                                </div>
                                            )}

                                        {/* Rejection Reason */}
                                        {submission.status === 'rejected' &&
                                            submission.rejection_reason && (
                                                <div className='bg-red-500/10 border-red-500/20 p-3 mt-3 border rounded'>
                                                    <p className='mb-1 text-sm font-medium text-red-500'>
                                                        Rejection Reason:
                                                    </p>
                                                    <p className='text-sm text-red-400'>
                                                        {
                                                            submission.rejection_reason
                                                        }
                                                    </p>
                                                </div>
                                            )}
                                    </div>
                                    </>
                                );

                                return clipUrl ? (
                                    <Link
                                        key={submission.id}
                                        to={clipUrl}
                                        className='md:flex-row bg-background-secondary hover:bg-background-tertiary flex flex-col gap-4 p-4 transition-colors rounded-lg'
                                    >
                                        {cardContent}
                                    </Link>
                                ) : (
                                    <div
                                        key={submission.id}
                                        className='md:flex-row bg-background-secondary flex flex-col gap-4 p-4 rounded-lg'
                                    >
                                        {cardContent}
                                    </div>
                                );
                            })}
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
        </Container>
    );
}
