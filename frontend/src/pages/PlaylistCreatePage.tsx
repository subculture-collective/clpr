import { Link, useNavigate } from 'react-router-dom';
import { ArrowLeft } from 'lucide-react';
import { Container, SEO } from '@/components';
import { PlaylistForm } from '@/components/playlist/PlaylistForm';
import type { PlaylistFormValues } from '@/components/playlist/PlaylistForm';
import { useCreatePlaylist } from '@/hooks/usePlaylist';
import { useToast } from '@/hooks';

export function PlaylistCreatePage() {
    const navigate = useNavigate();
    const toast = useToast();
    const createMutation = useCreatePlaylist();

    const handleSubmit = async (values: PlaylistFormValues) => {
        try {
            const playlist = await createMutation.mutateAsync({
                title: values.title,
                description: values.description || undefined,
                visibility: values.visibility,
                cover_url: values.cover_url || undefined,
            });
            toast.success('Playlist created');
            navigate(`/playlists/${playlist.id}`);
        } catch {
            toast.error('Failed to create playlist');
        }
    };

    return (
        <>
            <SEO
                title='Create Playlist'
                description='Create a new playlist to organize your favorite clips'
            />
            <Container className='py-6'>
                <div className='max-w-3xl mx-auto'>
                    <Link
                        to='/playlists'
                        className='inline-flex items-center gap-2 text-muted-foreground hover:text-white mb-6 transition-colors'
                    >
                        <ArrowLeft className='w-4 h-4' />
                        <span>Back to Playlists</span>
                    </Link>

                    <h1 className='text-3xl font-bold text-white mb-6'>
                        Create a new playlist
                    </h1>

                    <div className='bg-background border border-border rounded-xl p-6'>
                        <PlaylistForm
                            onSubmit={handleSubmit}
                            isSubmitting={createMutation.isPending}
                            submitLabel='Create Playlist'
                            onCancel={() => navigate('/playlists')}
                        />
                    </div>
                </div>
            </Container>
        </>
    );
}
