import { useState, useEffect } from 'react';
import { useParams, useSearchParams } from 'react-router-dom';
import { Container, Spinner, Button } from '../components';
import { ClipGridCard } from '../components/clip';
import { gameApi } from '../lib/game-api';
import type { GameWithStats } from '../types/game';
import type { Clip } from '../types/clip';
import { useAuth } from '../context/AuthContext';
import { useToast } from '../context/ToastContext';

type GameSort = 'hot' | 'new' | 'top' | 'rising';
type GameTimeframe = 'hour' | 'day' | 'week' | 'month' | 'year' | 'all';

export function GamePage() {
    const { gameId } = useParams<{ gameId: string }>();
    const [searchParams, setSearchParams] = useSearchParams();
    const { user } = useAuth();
    const { showToast } = useToast();

    const [game, setGame] = useState<GameWithStats | null>(null);
    const [clips, setClips] = useState<Clip[]>([]);
    const [loading, setLoading] = useState(true);
    const [clipsLoading, setClipsLoading] = useState(false);
    const [following, setFollowing] = useState(false);
    const [error, setError] = useState<string | null>(null);

    const sort = (searchParams.get('sort') as GameSort | null) || 'hot';
    const timeframe = searchParams.get('timeframe') as GameTimeframe | null;
    const page = parseInt(searchParams.get('page') || '1', 10);
    const [hasMore, setHasMore] = useState(false);

    useEffect(() => {
        const fetchGame = async () => {
            if (!gameId) return;

            try {
                setLoading(true);
                setError(null);
                const data = await gameApi.getGame(gameId);
                setGame(data.game);
                setFollowing(data.game.is_following);
            } catch (err) {
                console.error('Failed to fetch game:', err);
                setError('Failed to load game details');
            } finally {
                setLoading(false);
            }
        };

        fetchGame();
    }, [gameId]);

    useEffect(() => {
        const fetchClips = async () => {
            if (!gameId) return;

            try {
                setClipsLoading(true);
                const params: {
                    page: number;
                    limit: number;
                    sort: GameSort;
                    timeframe?: GameTimeframe;
                } = { page, limit: 20, sort };
                if (timeframe) params.timeframe = timeframe;

                const data = await gameApi.getGameClips(gameId, params);
                setClips(data.clips || []);
                setHasMore(data.has_more);
            } catch (err) {
                console.error('Failed to fetch clips:', err);
            } finally {
                setClipsLoading(false);
            }
        };

        fetchClips();
    }, [gameId, sort, timeframe, page]);

    const handleFollow = async () => {
        if (!user) {
            showToast('Please sign in to follow games', 'error');
            return;
        }

        if (!gameId) return;

        try {
            if (following) {
                await gameApi.unfollowGame(gameId);
                setFollowing(false);
                showToast('Unfollowed game', 'success');
                if (game) {
                    setGame({
                        ...game,
                        follower_count: game.follower_count - 1,
                    });
                }
            } else {
                await gameApi.followGame(gameId);
                setFollowing(true);
                showToast('Following game', 'success');
                if (game) {
                    setGame({
                        ...game,
                        follower_count: game.follower_count + 1,
                    });
                }
            }
        } catch (err) {
            console.error('Failed to toggle follow:', err);
            showToast('Failed to update follow status', 'error');
        }
    };

    const handleSortChange = (newSort: string) => {
        setSearchParams({ sort: newSort, ...(timeframe && { timeframe }) });
    };

    const handleTimeframeChange = (newTimeframe: string) => {
        if (newTimeframe) {
            setSearchParams({ sort, timeframe: newTimeframe });
        } else {
            setSearchParams({ sort });
        }
    };

    if (loading) {
        return (
            <Container className='py-8'>
                <div className='flex items-center justify-center min-h-[400px]'>
                    <Spinner size='xl' />
                </div>
            </Container>
        );
    }

    if (error || !game) {
        return (
            <Container className='py-8'>
                <div className='text-center text-muted-foreground py-12'>
                    <p className='text-lg'>{error || 'Game not found'}</p>
                </div>
            </Container>
        );
    }

    return (
        <Container className='py-8'>
            {/* Game Header */}
            <div className='mb-8'>
                <div className='flex items-start gap-6'>
                    {game.box_art_url && (
                        <img
                            src={game.box_art_url
                                .replace('{width}', '285')
                                .replace('{height}', '380')}
                            alt={game.name}
                            className='w-32 h-auto rounded-lg shadow-lg'
                        />
                    )}
                    <div className='flex-1'>
                        <h1 className='text-4xl font-bold mb-2'>{game.name}</h1>
                        <div className='flex items-center gap-4 text-sm text-muted-foreground mb-4'>
                            <span>{game.clip_count} clips</span>
                            <span>•</span>
                            <span>{game.follower_count} followers</span>
                        </div>
                        {user && (
                            <Button
                                onClick={handleFollow}
                                variant={following ? 'outline' : 'primary'}
                            >
                                {following ? 'Following' : 'Follow'}
                            </Button>
                        )}
                    </div>
                </div>
            </div>

            {/* Sort and Filter Controls */}
            <div className='mb-6 flex gap-4 items-center'>
                <div>
                    <label className='text-sm font-medium mr-2'>Sort by:</label>
                    <select
                        value={sort}
                        onChange={e => handleSortChange(e.target.value)}
                        className='px-3 py-1.5 border rounded-md bg-background'
                    >
                        <option value='hot'>Hot</option>
                        <option value='new'>New</option>
                        <option value='top'>Top</option>
                        <option value='rising'>Rising</option>
                    </select>
                </div>
                {(sort === 'top' || sort === 'rising') && (
                    <div>
                        <label className='text-sm font-medium mr-2'>
                            Timeframe:
                        </label>
                        <select
                            value={timeframe ?? ''}
                            onChange={e =>
                                handleTimeframeChange(e.target.value)
                            }
                            className='px-3 py-1.5 border rounded-md bg-background'
                        >
                            <option value=''>All Time</option>
                            <option value='hour'>Past Hour</option>
                            <option value='day'>Past Day</option>
                            <option value='week'>Past Week</option>
                            <option value='month'>Past Month</option>
                            <option value='year'>Past Year</option>
                        </select>
                    </div>
                )}
            </div>

            {/* Clips Grid */}
            {clipsLoading ?
                <div className='flex items-center justify-center min-h-[400px]'>
                    <Spinner size='lg' />
                </div>
            : clips.length === 0 ?
                <div className='text-center text-muted-foreground py-12'>
                    <p className='text-lg'>No clips found for this game</p>
                </div>
            :   <>
                    <div className='grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-6 mb-8'>
                        {clips.map(clip => (
                            <ClipGridCard key={clip.id} clip={clip} />
                        ))}
                    </div>

                    {/* Pagination */}
                    {hasMore && (
                        <div className='flex justify-center mt-8'>
                            <Button
                                onClick={() =>
                                    setSearchParams({
                                        sort,
                                        ...(timeframe && { timeframe }),
                                        page: String(page + 1),
                                    })
                                }
                                variant='outline'
                            >
                                Load More
                            </Button>
                        </div>
                    )}
                </>
            }
        </Container>
    );
}
