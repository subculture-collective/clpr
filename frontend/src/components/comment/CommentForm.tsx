import React from 'react';
import ReactMarkdown from 'react-markdown';
import { cn } from '@/lib/utils';
import { Button } from '@/components/ui';
import { EmojiPicker } from '@/components/ui/EmojiPicker';
import { MarkdownHelpModal } from '@/components/ui/MarkdownHelpModal';
import { LinkInputModal } from '@/components/ui/LinkInputModal';
import { useCreateComment, useUpdateComment, useToast } from '@/hooks';
import { useAutoSave, useDraftStorage } from '@/hooks/useAutoSave';
import { HelpCircle, Smile, Save, Link2, Quote } from 'lucide-react';

interface CommentFormProps {
    clipId: string;
    parentId?: string | null;
    parentUsername?: string;
    editCommentId?: string;
    initialContent?: string;
    onCancel?: () => void;
    onSuccess?: () => void;
    placeholder?: string;
    className?: string;
}

export const CommentForm: React.FC<CommentFormProps> = ({
    clipId,
    parentId,
    parentUsername,
    editCommentId,
    initialContent = '',
    onCancel,
    onSuccess,
    placeholder,
    className,
}) => {
    const [content, setContent] = React.useState(initialContent);
    const [showPreview, setShowPreview] = React.useState(false);
    const [showEmojiPicker, setShowEmojiPicker] = React.useState(false);
    const [showMarkdownHelp, setShowMarkdownHelp] = React.useState(false);
    const [showLinkModal, setShowLinkModal] = React.useState(false);
    const textareaRef = React.useRef<HTMLTextAreaElement>(null);

    const { mutate: createComment, isPending: isCreating } = useCreateComment();
    const { mutate: updateComment, isPending: isUpdating } = useUpdateComment();
    const toast = useToast();

    const isPending = isCreating || isUpdating;
    const maxLength = 5000;
    const isEmpty = content.trim().length === 0;

    // Draft storage key
    const draftKey = `comment-draft-${clipId}-${parentId || 'root'}`;
    const { saveDraft, loadDraft, clearDraft } = useDraftStorage(draftKey);

    // Auto-save functionality
    const { status: autoSaveStatus, lastSaved } = useAutoSave(content, {
        onSave: content => {
            saveDraft(content);
        },
        interval: 30000, // 30 seconds
        minLength: 1,
    });

    // Generate default placeholder based on context
    const defaultPlaceholder = React.useMemo(() => {
        if (placeholder) return placeholder;
        if (parentId && parentUsername) return `Reply to @${parentUsername}...`;
        return 'Write a comment...';
    }, [placeholder, parentId, parentUsername]);

    // Load draft on mount, but clear draft when editing existing comment
    React.useEffect(() => {
        if (editCommentId) {
            // Clear any stored draft when editing to avoid confusion
            clearDraft();
            return;
        }

        if (!initialContent) {
            const draft = loadDraft();
            if (draft) {
                setContent(draft);
                toast.info('Draft restored');
            }
        }
    }, [editCommentId, initialContent, loadDraft, clearDraft, toast]);

    React.useEffect(() => {
        // Focus textarea on mount
        if (textareaRef.current) {
            textareaRef.current.focus();
        }
    }, []);

    const handleSubmit = (e: React.FormEvent) => {
        e.preventDefault();

        if (isEmpty || isPending) return;

        if (editCommentId) {
            updateComment(
                {
                    commentId: editCommentId,
                    payload: { content },
                },
                {
                    onSuccess: () => {
                        setContent('');
                        clearDraft(); // Clear draft on successful submit
                        toast.success('Comment updated successfully');
                        onSuccess?.();
                    },
                    onError: error => {
                        toast.error(
                            error instanceof Error ?
                                error.message
                            :   'Failed to update comment. Please try again.',
                        );
                    },
                },
            );
        } else {
            createComment(
                {
                    clip_id: clipId,
                    content,
                    parent_comment_id: parentId,
                },
                {
                    onSuccess: () => {
                        setContent('');
                        clearDraft(); // Clear draft on successful submit
                        toast.success(
                            parentId ?
                                'Reply posted successfully'
                            :   'Comment posted successfully',
                        );
                        onSuccess?.();
                    },
                    onError: error => {
                        toast.error(
                            error instanceof Error ?
                                error.message
                            :   'Failed to post comment. Please try again.',
                        );
                    },
                },
            );
        }
    };

    const handleCancel = () => {
        setContent(initialContent);
        onCancel?.();
    };

    const handleKeyDown = (e: React.KeyboardEvent) => {
        // Submit on Ctrl+Enter
        if (e.key === 'Enter' && (e.ctrlKey || e.metaKey)) {
            e.preventDefault();
            handleSubmit(e);
        }

        // Cancel on Escape
        if (e.key === 'Escape') {
            handleCancel();
        }
    };

    const insertMarkdown = (before: string, after = '') => {
        if (!textareaRef.current) return;

        const textarea = textareaRef.current;
        const start = textarea.selectionStart;
        const end = textarea.selectionEnd;
        const selectedText = content.substring(start, end);
        const newText =
            content.substring(0, start) +
            before +
            selectedText +
            after +
            content.substring(end);

        setContent(newText);

        // Restore cursor position
        setTimeout(() => {
            textarea.focus();
            const newCursorPos = start + before.length + selectedText.length;
            textarea.setSelectionRange(newCursorPos, newCursorPos);
        }, 0);
    };

    const insertLink = (url: string, linkText: string) => {
        // URL is already validated and normalized by LinkInputModal
        const displayText = linkText || url;
        insertMarkdown(`[${displayText}](`, `${url})`);
    };

    const insertEmoji = (emoji: string) => {
        if (!textareaRef.current) return;

        const textarea = textareaRef.current;
        const start = textarea.selectionStart;
        const end = textarea.selectionEnd;
        const newText =
            content.substring(0, start) + emoji + content.substring(end);

        setContent(newText);
        setShowEmojiPicker(false);

        // Restore focus and cursor position
        setTimeout(() => {
            textarea.focus();
            const newCursorPos = start + emoji.length;
            textarea.setSelectionRange(newCursorPos, newCursorPos);
        }, 0);
    };

    return (
        <form onSubmit={handleSubmit} className={cn('space-y-3', className)}>
            <div className='border border-border rounded-lg overflow-hidden'>
                {/* Toolbar */}
                <div className='flex items-center gap-1 p-2 bg-muted border-b border-border'>
                    <button
                        type='button'
                        onClick={() => insertMarkdown('**', '**')}
                        className='p-1.5 rounded hover:bg-neutral-200 dark:hover:bg-neutral-700 transition-colors cursor-pointer'
                        title='Bold (Ctrl+B)'
                    >
                        <strong className='text-sm'>B</strong>
                    </button>
                    <button
                        type='button'
                        onClick={() => insertMarkdown('*', '*')}
                        className='p-1.5 rounded hover:bg-neutral-200 dark:hover:bg-neutral-700 transition-colors cursor-pointer'
                        title='Italic (Ctrl+I)'
                    >
                        <em className='text-sm'>I</em>
                    </button>
                    <button
                        type='button'
                        onClick={() => insertMarkdown('~~', '~~')}
                        className='p-1.5 rounded hover:bg-neutral-200 dark:hover:bg-neutral-700 transition-colors cursor-pointer'
                        title='Strikethrough'
                    >
                        <s className='text-sm'>S</s>
                    </button>
                    <div className='w-px h-5 bg-border mx-1' />
                    <button
                        type='button'
                        onClick={() => setShowLinkModal(true)}
                        className='p-1.5 rounded hover:bg-neutral-200 dark:hover:bg-neutral-700 transition-colors text-sm cursor-pointer'
                        title='Insert Link'
                    >
                        <Link2 size={16} strokeWidth={1.75} />
                    </button>
                    <button
                        type='button'
                        onClick={() => insertMarkdown('> ', '')}
                        className='p-1.5 rounded hover:bg-neutral-200 dark:hover:bg-neutral-700 transition-colors text-sm cursor-pointer'
                        title='Quote'
                    >
                        <Quote size={16} strokeWidth={1.75} />
                    </button>
                    <button
                        type='button'
                        onClick={() => insertMarkdown('`', '`')}
                        className='p-1.5 rounded hover:bg-neutral-200 dark:hover:bg-neutral-700 transition-colors text-sm cursor-pointer'
                        title='Code'
                    >
                        {'</>'}
                    </button>

                    <div className='w-px h-5 bg-border mx-1' />

                    {/* Emoji picker button */}
                    <div className='relative'>
                        <button
                            type='button'
                            onClick={() => setShowEmojiPicker(!showEmojiPicker)}
                            className='p-1.5 rounded hover:bg-neutral-200 dark:hover:bg-neutral-700 transition-colors cursor-pointer'
                            title='Insert Emoji'
                            aria-label='Emoji picker'
                        >
                            <Smile className='w-4 h-4' />
                        </button>
                        {showEmojiPicker && (
                            <EmojiPicker
                                onEmojiSelect={insertEmoji}
                                onClose={() => setShowEmojiPicker(false)}
                                className='absolute bottom-full mb-2 right-0 z-50'
                            />
                        )}
                    </div>

                    {/* Markdown help button */}
                    <button
                        type='button'
                        onClick={() => setShowMarkdownHelp(true)}
                        className='p-1.5 rounded hover:bg-neutral-200 dark:hover:bg-neutral-700 transition-colors cursor-pointer'
                        title='Markdown Help'
                        aria-label='Markdown formatting guide'
                    >
                        <HelpCircle className='w-4 h-4' />
                    </button>

                    <div className='flex-1' />

                    {/* Auto-save status */}
                    {autoSaveStatus !== 'idle' && (
                        <div className='flex items-center gap-1 text-xs text-muted-foreground mr-2'>
                            <Save className='w-3 h-3' />
                            {autoSaveStatus === 'saving' && 'Saving...'}
                            {autoSaveStatus === 'saved' && 'Saved'}
                            {autoSaveStatus === 'error' && 'Save failed'}
                        </div>
                    )}

                    {/* Tab buttons */}
                    <div className='flex gap-1'>
                        <button
                            type='button'
                            onClick={() => setShowPreview(false)}
                            className={cn(
                                'px-3 py-1 text-sm rounded transition-colors cursor-pointer',
                                !showPreview ?
                                    'bg-background text-foreground'
                                :   'text-muted-foreground hover:text-foreground',
                            )}
                        >
                            Write
                        </button>
                        <button
                            type='button'
                            onClick={() => setShowPreview(true)}
                            className={cn(
                                'px-3 py-1 text-sm rounded transition-colors cursor-pointer',
                                showPreview ?
                                    'bg-background text-foreground'
                                :   'text-muted-foreground hover:text-foreground',
                            )}
                        >
                            Preview
                        </button>
                    </div>
                </div>

                {/* Content area */}
                {showPreview ?
                    <div className='p-3 min-h-[120px] prose prose-sm dark:prose-invert max-w-none'>
                        {content ?
                            <ReactMarkdown
                                components={{
                                    // Open external links in new tab
                                    a: ({ ...props }) => (
                                        <a
                                            {...props}
                                            target='_blank'
                                            rel='noopener noreferrer'
                                            className='text-primary-500 hover:text-primary-600 underline'
                                        />
                                    ),
                                    // Code blocks
                                    code: ({
                                        className,
                                        children,
                                        ...props
                                    }) => {
                                        const isInline = !className;
                                        return isInline ?
                                                <code
                                                    className='px-1 py-0.5 rounded bg-muted text-sm font-mono'
                                                    {...props}
                                                >
                                                    {children}
                                                </code>
                                            :   <code
                                                    className={cn(
                                                        'block p-3 rounded bg-muted text-sm font-mono overflow-x-auto',
                                                        className,
                                                    )}
                                                    {...props}
                                                >
                                                    {children}
                                                </code>;
                                    },
                                }}
                            >
                                {content}
                            </ReactMarkdown>
                        :   <p className='text-muted-foreground'>
                                Nothing to preview
                            </p>
                        }
                    </div>
                :   <textarea
                        ref={textareaRef}
                        value={content}
                        onChange={e => setContent(e.target.value)}
                        onKeyDown={handleKeyDown}
                        placeholder={defaultPlaceholder}
                        maxLength={maxLength}
                        className='w-full px-3 py-2 min-h-[120px] resize-y bg-background text-foreground placeholder:text-muted-foreground focus:outline-none'
                        aria-label='Comment content'
                    />
                }

                {/* Footer */}
                <div className='flex items-center justify-between px-3 py-2 bg-muted border-t border-border text-sm'>
                    <span className='text-muted-foreground'>
                        Markdown supported. {maxLength - content.length}{' '}
                        characters remaining.
                    </span>
                    <span className='text-muted-foreground'>
                        {content.length}/{maxLength}
                    </span>
                </div>
            </div>

            {/* Action buttons */}
            <div className='flex items-center gap-2'>
                <Button
                    type='submit'
                    disabled={isEmpty || isPending}
                    loading={isPending}
                    size='sm'
                >
                    {editCommentId ?
                        'Update'
                    : parentId ?
                        'Reply'
                    :   'Comment'}
                </Button>

                {(onCancel || editCommentId) && (
                    <Button
                        type='button'
                        variant='ghost'
                        size='sm'
                        onClick={handleCancel}
                    >
                        Cancel
                    </Button>
                )}

                <span className='text-xs text-muted-foreground ml-2'>
                    Tip: Press Ctrl+Enter to submit
                </span>

                {lastSaved && (
                    <span className='text-xs text-muted-foreground ml-auto'>
                        Last saved: {lastSaved.toLocaleTimeString()}
                    </span>
                )}
            </div>

            {/* Markdown Help Modal */}
            <MarkdownHelpModal
                isOpen={showMarkdownHelp}
                onClose={() => setShowMarkdownHelp(false)}
            />

            {/* Link Input Modal */}
            <LinkInputModal
                isOpen={showLinkModal}
                onClose={() => setShowLinkModal(false)}
                onInsert={insertLink}
            />
        </form>
    );
};
