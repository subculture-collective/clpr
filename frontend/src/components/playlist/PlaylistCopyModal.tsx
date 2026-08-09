import { Modal } from '@/components/ui/Modal';
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
        <Modal open onClose={onClose} title='Copy playlist' size='xl'>
                <p className='mb-4 text-sm text-muted-foreground'>
                    Customize your new playlist before saving.
                </p>
                    <PlaylistForm
                        initialValues={initialValues}
                        onSubmit={onSubmit}
                        onCancel={onClose}
                        submitLabel='Create Copy'
                        isSubmitting={isSubmitting}
                    />
        </Modal>
    );
}
