import { Link } from 'react-router-dom';
import { Badge, Button, Card } from '../ui';
import type { ClipSubmission } from '@/types/submission';

export interface SubmissionConfirmationProps {
    /**
     * The submission data to display
     */
    submission: ClipSubmission;
    /**
     * Callback when "Submit Another" is clicked
     */
    onSubmitAnother?: () => void;
    /**
     * If true, shows a "Back to Discovery" link
     */
    fromDiscover?: boolean;
}

/**
 * SubmissionConfirmation displays a summary of a successfully submitted clip
 * with quick action buttons and status information
 */
export function SubmissionConfirmation({
    submission,
    onSubmitAnother,
    fromDiscover = false,
}: SubmissionConfirmationProps) {
    const statusVariant =
        submission.status === 'approved' ? 'success' : 'warning';
    const isAutoApproved = submission.status === 'approved';

    // Format date for display
    const submittedDate = new Date(submission.created_at).toLocaleString(
        'en-US',
        {
            month: 'short',
            day: 'numeric',
            year: 'numeric',
            hour: 'numeric',
            minute: '2-digit',
        },
    );

    return (
        <Card className='max-w-2xl mx-auto p-6'>
            {/* Success Icon and Title */}
            <div className='text-center mb-6'>
                <div className='inline-flex items-center justify-center w-16 h-16 rounded-full bg-success-100 dark:bg-success-900 mb-4'>
                    <svg
                        className='w-8 h-8 text-success-600 dark:text-success-400'
                        fill='none'
                        stroke='currentColor'
                        viewBox='0 0 24 24'
                        xmlns='http://www.w3.org/2000/svg'
                    >
                        <path
                            strokeLinecap='round'
                            strokeLinejoin='round'
                            strokeWidth={2}
                            d='M5 13l4 4L19 7'
                        />
                    </svg>
                </div>
                <h2 className='text-2xl font-bold mb-2'>
                    Submission Successful!
                </h2>
                <p className='text-muted-foreground'>
                    Your clip has been submitted for review
                </p>
            </div>

            {/* Submission Details */}
            <div className='space-y-4 mb-6'>
                {/* Title */}
                <div>
                    <label className='text-sm font-medium text-muted-foreground block mb-1'>
                        Title
                    </label>
                    <p className='text-base font-medium'>
                        {submission.custom_title ||
                            submission.title ||
                            'Untitled'}
                    </p>
                </div>

                {/* Streamer */}
                {(submission.broadcaster_name || submission.creator_name) && (
                    <div>
                        <label className='text-sm font-medium text-muted-foreground block mb-1'>
                            Streamer
                        </label>
                        <p className='text-base'>
                            {submission.broadcaster_name ||
                                submission.creator_name}
                        </p>
                    </div>
                )}

                {/* Tags */}
                {submission.tags && submission.tags.length > 0 && (
                    <div>
                        <label className='text-sm font-medium text-muted-foreground block mb-1'>
                            Tags
                        </label>
                        <div className='flex flex-wrap gap-2'>
                            {submission.tags.map(tag => (
                                <span
                                    key={tag}
                                    className='px-3 py-1 bg-primary/10 text-primary rounded-full text-sm'
                                >
                                    {tag}
                                </span>
                            ))}
                        </div>
                    </div>
                )}

                {/* Submitted Date */}
                <div>
                    <label className='text-sm font-medium text-muted-foreground block mb-1'>
                        Submitted
                    </label>
                    <p className='text-base'>{submittedDate}</p>
                </div>

                {/* Status */}
                <div>
                    <label className='text-sm font-medium text-muted-foreground block mb-1'>
                        Status
                    </label>
                    <Badge variant={statusVariant} size='md'>
                        {submission.status}
                    </Badge>
                </div>
            </div>

            {/* Auto-approval message and clip link */}
            {isAutoApproved && (
                <div className='bg-success-50 dark:bg-success-950 border border-success-200 dark:border-success-800 rounded-lg p-4 mb-6'>
                    <div className='flex items-start gap-3'>
                        <div className='flex-shrink-0'>
                            <svg
                                className='w-5 h-5 text-success-600 dark:text-success-400'
                                fill='currentColor'
                                viewBox='0 0 20 20'
                                xmlns='http://www.w3.org/2000/svg'
                            >
                                <path
                                    fillRule='evenodd'
                                    d='M10 18a8 8 0 100-16 8 8 0 000 16zm3.707-9.293a1 1 0 00-1.414-1.414L9 10.586 7.707 9.293a1 1 0 00-1.414 1.414l2 2a1 1 0 001.414 0l4-4z'
                                    clipRule='evenodd'
                                />
                            </svg>
                        </div>
                        <div className='flex-1'>
                            <h3 className='text-sm font-semibold text-success-800 dark:text-success-200 mb-1'>
                                Auto-approved!
                            </h3>
                            <p className='text-sm text-success-700 dark:text-success-300 mb-3'>
                                Your clip has been automatically approved and is
                                now live on the platform.
                            </p>
                            <Link
                                to={
                                    submission.clip_id ?
                                        `/clip/${submission.clip_id}`
                                    :   `/clip/${submission.twitch_clip_id}`
                                }
                                className='text-sm font-medium text-success-700 dark:text-success-300 hover:text-success-800 dark:hover:text-success-200 underline cursor-pointer'
                            >
                                View your clip →
                            </Link>
                        </div>
                    </div>
                </div>
            )}

            {/* Action Buttons */}
            <div className='flex flex-col sm:flex-row gap-3'>
                {fromDiscover && (
                    <Button asChild variant='primary' fullWidth>
                        <Link to='/discover' className='sm:flex-1'>
                            Back to Discovery
                        </Link>
                    </Button>
                )}
                <Button
                    asChild
                        variant={fromDiscover ? 'secondary' : 'primary'}
                        fullWidth
                    >
                    <Link to='/submissions' className='sm:flex-1'>
                        View My Submissions
                    </Link>
                </Button>
                <Button
                    variant='secondary'
                    fullWidth
                    className='sm:flex-1'
                    onClick={onSubmitAnother}
                >
                    Submit Another Clip
                </Button>
            </div>
        </Card>
    );
}
