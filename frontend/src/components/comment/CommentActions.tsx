import React from 'react';
import { cn } from '@/lib/utils';
import {
    useDeleteComment,
    useReportComment,
    useIsAuthenticated,
    useToast,
} from '@/hooks';
import { Modal } from '@/components/ui';

interface CommentActionsProps {
    commentId: string;
    clipId: string;
    isAuthor: boolean;
    isAdmin: boolean;
    createdAt: string;
    onReply?: () => void;
    onEdit?: () => void;
    depth?: number;
    maxDepth?: number;
    variant?: 'expanded' | 'compact';
    className?: string;
}

const EDIT_WINDOW_MINUTES = 15;

export const CommentActions: React.FC<CommentActionsProps> = ({
    commentId,
    clipId,
    isAuthor,
    isAdmin,
    createdAt,
    onReply,
    onEdit,
    depth = 0,
    maxDepth = 10,
    variant = 'expanded',
    className,
}) => {
    const isAuthenticated = useIsAuthenticated();
    const { mutate: deleteComment } = useDeleteComment();
    const { mutate: reportComment } = useReportComment();
    const toast = useToast();

    const [showDeleteDialog, setShowDeleteDialog] = React.useState(false);
    const [showReportDialog, setShowReportDialog] = React.useState(false);
    const [reportReason, setReportReason] = React.useState<string>('spam');
    const [reportDescription, setReportDescription] = React.useState('');

    const isCompact = variant === 'compact';

    const computeWithinEditWindow = React.useCallback(() => {
        const createdTime = new Date(createdAt).getTime();
        const minutesSinceCreated = (Date.now() - createdTime) / (1000 * 60);
        return minutesSinceCreated <= EDIT_WINDOW_MINUTES;
    }, [createdAt]);

    // Track whether the comment is still within the editable window
    const [isWithinEditWindow, setIsWithinEditWindow] = React.useState(() =>
        computeWithinEditWindow(),
    );

    React.useEffect(() => {
        setIsWithinEditWindow(computeWithinEditWindow());
        const timer = window.setInterval(() => {
            setIsWithinEditWindow(computeWithinEditWindow());
        }, 60_000);

        return () => clearInterval(timer);
    }, [computeWithinEditWindow]);

    const canEdit = isAuthor && isWithinEditWindow;
    const canDelete = isAuthor || isAdmin;
    const canReply = isAuthenticated && depth < maxDepth;

    const handleDelete = () => {
        deleteComment(commentId, {
            onSuccess: () => {
                setShowDeleteDialog(false);
            },
        });
    };

    const handleReport = () => {
        reportComment(
            {
                comment_id: commentId,
                reason: reportReason as
                    | 'spam'
                    | 'harassment'
                    | 'off-topic'
                    | 'misinformation'
                    | 'other',
                description: reportDescription || undefined,
            },
            {
                onSuccess: () => {
                    setShowReportDialog(false);
                    setReportDescription('');
                    toast.success(
                        'Comment reported. Thank you for helping keep our community safe.',
                    );
                },
            },
        );
    };

    const handleShare = () => {
        const url = `${window.location.origin}/clips/${clipId}#comment-${commentId}`;
        navigator.clipboard.writeText(url).then(() => {
            toast.success('Link copied to clipboard!');
        });
    };

    return (
        <>
            <div className={cn('flex items-center gap-3 text-[12px]', className)}>
                {canReply && (
                    <button
                        onClick={onReply}
                        className='text-muted-foreground hover:text-cta transition-colors font-medium cursor-pointer'
                    >
                        Reply
                    </button>
                )}

                {!isCompact && canEdit && (
                    <button
                        onClick={onEdit}
                        className='text-muted-foreground hover:text-cta transition-colors font-medium cursor-pointer'
                    >
                        Edit
                    </button>
                )}

                {!isCompact && canDelete && (
                    <button
                        onClick={() => setShowDeleteDialog(true)}
                        className='text-muted-foreground hover:text-error-500 transition-colors font-medium cursor-pointer'
                    >
                        Delete
                    </button>
                )}

                {!isCompact && (
                    <button
                        onClick={handleShare}
                        className='text-muted-foreground hover:text-cta transition-colors font-medium cursor-pointer'
                    >
                        Share
                    </button>
                )}

                {!isCompact && isAuthenticated && !isAuthor && (
                    <button
                        onClick={() => setShowReportDialog(true)}
                        className='text-muted-foreground hover:text-error-500 transition-colors font-medium cursor-pointer'
                    >
                        Report
                    </button>
                )}
            </div>

            {/* Delete confirmation dialog */}
            {showDeleteDialog && (
                <Modal
                    open={showDeleteDialog}
                    onClose={() => setShowDeleteDialog(false)}
                    title='Delete Comment'
                >
                    <div className='space-y-4'>
                        <p className='text-muted-foreground'>
                            Are you sure you want to delete this comment? This
                            action cannot be undone.
                        </p>
                        <div className='flex justify-end gap-2'>
                            <button
                                onClick={() => setShowDeleteDialog(false)}
                                className='px-4 py-2 rounded-md border border-border hover:bg-muted transition-colors cursor-pointer'
                            >
                                Cancel
                            </button>
                            <button
                                onClick={handleDelete}
                                className='px-4 py-2 rounded-md bg-error-500 text-white hover:bg-error-600 transition-colors cursor-pointer'
                            >
                                Delete
                            </button>
                        </div>
                    </div>
                </Modal>
            )}

            {/* Report dialog */}
            {showReportDialog && (
                <Modal
                    open={showReportDialog}
                    onClose={() => setShowReportDialog(false)}
                    title='Report Comment'
                >
                    <div className='space-y-4'>
                        <div>
                            <label className='block text-sm font-medium mb-2'>
                                Reason
                            </label>
                            <select
                                value={reportReason}
                                onChange={e => setReportReason(e.target.value)}
                                className='w-full px-3 py-2 rounded-md border border-border bg-background cursor-pointer'
                            >
                                <option value='spam'>Spam</option>
                                <option value='harassment'>Harassment</option>
                                <option value='off-topic'>Off-topic</option>
                                <option value='misinformation'>
                                    Misinformation
                                </option>
                                <option value='other'>Other</option>
                            </select>
                        </div>

                        <div>
                            <label className='block text-sm font-medium mb-2'>
                                Additional details (optional)
                            </label>
                            <textarea
                                value={reportDescription}
                                onChange={e =>
                                    setReportDescription(e.target.value)
                                }
                                className='w-full px-3 py-2 rounded-md border border-border bg-background min-h-[100px] resize-y'
                                placeholder='Provide additional context...'
                            />
                        </div>

                        <div className='flex justify-end gap-2'>
                            <button
                                onClick={() => setShowReportDialog(false)}
                                className='px-4 py-2 rounded-md border border-border hover:bg-muted transition-colors cursor-pointer'
                            >
                                Cancel
                            </button>
                            <button
                                onClick={handleReport}
                                className='px-4 py-2 rounded-md bg-primary-500 text-white hover:bg-primary-600 transition-colors cursor-pointer'
                            >
                                Submit Report
                            </button>
                        </div>
                    </div>
                </Modal>
            )}
        </>
    );
};
