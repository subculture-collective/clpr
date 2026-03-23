import { useState } from 'react';
import { Link } from 'react-router-dom';
import { useQuery } from '@tanstack/react-query';
import { Bookmark } from 'lucide-react';
import { Container, SEO } from '../components';
import { MiniFooter } from '../components/layout';
import { PlaylistCard } from '../components/playlist/PlaylistCard';
import { Button } from '@/components/ui';
import { fetchBookmarkedPlaylists } from '@/lib/playlist-api';
import type { Playlist } from '@/types/playlist';

export function BookmarkedPlaylistsPage() {
    const [page, setPage] = useState(1);

    const { data, isLoading } = useQuery({
        queryKey: ['bookmarkedPlaylists', page],
        queryFn: () => fetchBookmarkedPlaylists(page),
    });

    const playlists: Playlist[] = data?.playlists || [];
    const hasNext = data?.has_next || false;

    return (
        <>
            <SEO
                title='Bookmarked Playlists'
                description='Your bookmarked playlists'
            />
            <Container>
                <div className='py-8'>
                    {/* Header */}
                    <div className='mb-8'>
                        <div className='flex items-center gap-3 mb-2'>
                            <Bookmark className='h-7 w-7 text-foreground' />
                            <h1 className='text-3xl font-bold text-foreground'>
                                Bookmarked Playlists
                            </h1>
                        </div>
                        <p className='text-muted-foreground'>
                            Playlists you've saved for later
                        </p>
                    </div>

                    {/* Playlists Grid */}
                    {isLoading ?
                        <div className='text-center py-12 text-muted-foreground'>
                            Loading...
                        </div>
                    : playlists.length === 0 ?
                        <div className='text-center py-12 text-muted-foreground'>
                            <Bookmark className='h-12 w-12 mx-auto mb-4 opacity-50' />
                            <p>No bookmarked playlists yet</p>
                            <p className='text-sm mt-2'>
                                <Link
                                    to='/playlists/discover'
                                    className='text-primary-500 hover:underline'
                                >
                                    Discover playlists
                                </Link>{' '}
                                to find something you like!
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

                            {/* Load More */}
                            {hasNext && (
                                <div className='flex justify-center'>
                                    <Button
                                        variant='outline'
                                        onClick={() => setPage(p => p + 1)}
                                    >
                                        Load More
                                    </Button>
                                </div>
                            )}
                        </>
                    }
                </div>
                <MiniFooter />
            </Container>
        </>
    );
}
