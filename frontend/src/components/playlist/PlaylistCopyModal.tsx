import { X } from 'lucide-react';
import { PlaylistForm, type PlaylistFormValues } from './PlaylistForm';

interface PlaylistCopyModalProps {
    initialValues: PlaylistFormValues;
    onSubmit: (values: PlaylistFormValues) => Promise<void> | void;
    onClose: () => void;
    isSubmitting?: boolean;
}

export function PlaylistCopyModal({
    initialValues,
    onSubmit,
    onClose,
    isSubmitting = false,
}: PlaylistCopyModalProps) {
    return (
        <div className='fixed inset-0 bg-black/70 flex items-center justify-center z-50 p-4'>
            <div className='bg-surface rounded-xl max-w-2xl w-full border border-border shadow-xl'>
                <div className='flex items-center justify-between p-6 border-b border-border'>
                    <div>
                        <h2 className='text-xl font-bold text-white'>
                            Copy playlist
                        </h2>
                        <p className='text-sm text-muted-foreground'>
                            Customize your new playlist before saving.
                        </p>
                    </div>
                    <button
                        onClick={onClose}
                        className='text-muted-foreground hover:text-white transition-colors'
                        aria-label='Close copy modal'
                    >
                        <X className='h-5 w-5' />
                    </button>
                </div>

                <div className='p-6'>
                    <PlaylistForm
                        initialValues={initialValues}
                        onSubmit={onSubmit}
                        onCancel={onClose}
                        submitLabel='Create Copy'
                        isSubmitting={isSubmitting}
                    />
                </div>
            </div>
        </div>
    );
}
