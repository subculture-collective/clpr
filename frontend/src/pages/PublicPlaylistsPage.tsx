import { Container, ScrollToTop } from '../components';
import { MiniFooter } from '../components/layout';
import { SEO } from '../components';
import { PlaylistCard } from '../components/playlist';
import { usePublicPlaylists } from '@/hooks/usePlaylist';
import { useState } from 'react';
import { Button } from '@/components/ui';
import { useNavigate } from 'react-router-dom';
import { Plus } from 'lucide-react';

export function PublicPlaylistsPage() {
    const [page, setPage] = useState(1);
    const { data, isLoading } = usePublicPlaylists(page, 20);
    const navigate = useNavigate();

    const playlists = data?.data || [];
    const meta = data?.meta;

    return (
        <>
            <SEO
                title='Discover Playlists'
                description='Browse and discover public playlists created by the community'
            />
            <Container>
                <div className='py-8'>
                    {/* Header */}
                    <div className='mb-8 flex items-center justify-between'>
                        <div className='flex-1'>
                            <h1 className='text-3xl font-bold text-foreground mb-2'>
                                Discover Playlists
                            </h1>
                            <p className='text-muted-foreground'>
                                Browse public playlists created by the community
                            </p>
                        </div>
                        <Button
                            onClick={() => navigate('/playlists/new')}
                            className='flex items-center gap-2'
                        >
                            <Plus className='h-4 w-4' />
                            Create Playlist
                        </Button>
                    </div>

                    {/* Playlists Grid */}
                    {isLoading ?
                        <div className='text-center py-12 text-muted-foreground'>
                            Loading...
                        </div>
                    : playlists.length === 0 ?
                        <div className='text-center py-12 text-muted-foreground'>
                            <p>No public playlists found.</p>
                            <p className='text-sm mt-2'>
                                Check back later for community playlists!
                            </p>
                        </div>
                    :   <>
                            <div className='grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-6 mb-8'>
                                {playlists.map(playlist => (
                                    <PlaylistCard
                                        key={playlist.id}
                                        playlist={playlist}
                                    />
                                ))}
                            </div>

                            {/* Pagination */}
                            {meta && meta.total_pages > 1 && (
                                <div className='flex items-center justify-center gap-4'>
                                    <Button
                                        variant='outline'
                                        onClick={() =>
                                            setPage(p => Math.max(1, p - 1))
                                        }
                                        disabled={!meta.has_prev}
                                    >
                                        Previous
                                    </Button>
                                    <span className='text-muted-foreground'>
                                        Page {meta.page} of {meta.total_pages}
                                    </span>
                                    <Button
                                        variant='outline'
                                        onClick={() => setPage(p => p + 1)}
                                        disabled={!meta.has_next}
                                    >
                                        Next
                                    </Button>
                                </div>
                            )}
                        </>
                    }
                </div>
                <MiniFooter />
            </Container>
            <ScrollToTop />
        </>
    );
}
