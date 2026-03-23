import { useState, useEffect } from 'react';
import { Button } from '@/components/ui';
import { cn } from '@/lib/utils';

export interface PlaylistFormValues {
    title: string;
    description: string;
    visibility: 'private' | 'public' | 'unlisted';
    cover_url: string;
}

interface PlaylistFormProps {
    initialValues?: Partial<PlaylistFormValues>;
    onSubmit: (values: PlaylistFormValues) => Promise<void> | void;
    onCancel?: () => void;
    submitLabel?: string;
    isSubmitting?: boolean;
    showCancel?: boolean;
    className?: string;
}

const defaultValues: PlaylistFormValues = {
    title: '',
    description: '',
    visibility: 'private',
    cover_url: '',
};

export function PlaylistForm({
    initialValues,
    onSubmit,
    onCancel,
    submitLabel = 'Create Playlist',
    isSubmitting = false,
    showCancel = true,
    className,
}: PlaylistFormProps) {
    const [values, setValues] = useState<PlaylistFormValues>({
        ...defaultValues,
        ...initialValues,
    });

    useEffect(() => {
        if (!initialValues) return;
        setValues(prev => ({ ...prev, ...initialValues }));
    }, [initialValues]);

    const handleChange = (field: keyof PlaylistFormValues, value: string) => {
        setValues(prev => ({ ...prev, [field]: value }));
    };

    const handleSubmit = (event: React.FormEvent) => {
        event.preventDefault();
        onSubmit({
            ...values,
            title: values.title.trim(),
            description: values.description.trim(),
            cover_url: values.cover_url.trim(),
        });
    };

    const titleLength = values.title.length;
    const descriptionLength = values.description.length;

    return (
        <form onSubmit={handleSubmit} className={cn('space-y-6', className)}>
            <div>
                <label
                    htmlFor='title'
                    className='block text-sm font-medium text-foreground mb-2'
                >
                    Title <span className='text-red-500'>*</span>
                </label>
                <input
                    id='title'
                    type='text'
                    value={values.title}
                    onChange={e => handleChange('title', e.target.value)}
                    placeholder='My favorite clips'
                    maxLength={100}
                    className={cn(
                        'w-full px-4 py-3 bg-surface border border-border',
                        'text-white placeholder-muted-foreground rounded-lg',
                        'focus:outline-none focus:ring-2 focus:ring-purple-500 focus:border-transparent',
                    )}
                    required
                />
                <p className='mt-1 text-xs text-text-secondary'>
                    {titleLength}/100 characters
                </p>
            </div>

            <div>
                <label
                    htmlFor='description'
                    className='block text-sm font-medium text-foreground mb-2'
                >
                    Description
                </label>
                <textarea
                    id='description'
                    value={values.description}
                    onChange={e => handleChange('description', e.target.value)}
                    placeholder='What is this playlist about?'
                    rows={4}
                    maxLength={500}
                    className={cn(
                        'w-full px-4 py-3 bg-surface border border-border',
                        'text-white placeholder-muted-foreground rounded-lg resize-none',
                        'focus:outline-none focus:ring-2 focus:ring-purple-500 focus:border-transparent',
                    )}
                />
                <p className='mt-1 text-xs text-text-secondary'>
                    {descriptionLength}/500 characters
                </p>
            </div>

            <div>
                <label
                    htmlFor='cover-url'
                    className='block text-sm font-medium text-foreground mb-2'
                >
                    Cover Image URL
                </label>
                <input
                    id='cover-url'
                    type='url'
                    value={values.cover_url}
                    onChange={e => handleChange('cover_url', e.target.value)}
                    placeholder='https://example.com/cover.jpg'
                    className={cn(
                        'w-full px-4 py-3 bg-surface border border-border',
                        'text-white placeholder-muted-foreground rounded-lg',
                        'focus:outline-none focus:ring-2 focus:ring-purple-500 focus:border-transparent',
                    )}
                />
                <p className='mt-1 text-xs text-text-secondary'>
                    Optional. Use a direct image URL.
                </p>
            </div>

            <div>
                <label
                    htmlFor='visibility'
                    className='block text-sm font-medium text-foreground mb-2'
                >
                    Visibility
                </label>
                <select
                    id='visibility'
                    value={values.visibility}
                    onChange={e =>
                        handleChange(
                            'visibility',
                            e.target.value as PlaylistFormValues['visibility'],
                        )
                    }
                    className={cn(
                        'w-full px-4 py-3 bg-surface border border-border',
                        'text-white rounded-lg',
                        'focus:outline-none focus:ring-2 focus:ring-purple-500 focus:border-transparent',
                    )}
                >
                    <option value='private'>
                        Private - Only you can see it
                    </option>
                    <option value='unlisted'>
                        Unlisted - Anyone with link
                    </option>
                    <option value='public'>
                        Public - Discoverable by everyone
                    </option>
                </select>
            </div>

            <div className='flex gap-3'>
                {showCancel && (
                    <Button
                        type='button'
                        variant='secondary'
                        onClick={onCancel}
                    >
                        Cancel
                    </Button>
                )}
                <Button
                    type='submit'
                    disabled={!values.title.trim() || isSubmitting}
                    variant='primary'
                >
                    {isSubmitting ? 'Saving…' : submitLabel}
                </Button>
            </div>
        </form>
    );
}
