import { useState, useEffect } from 'react';
import { useParams, useSearchParams } from 'react-router-dom';
import { Container, Spinner, Button, ResourceUnavailable } from '../components';
import { ClipGridCard } from '../components/clip';
import { twitchCategoryApi } from '../lib/game-api';
import type { GameWithStats } from '../types/game';
import type { Clip } from '../types/clip';
import { useAuth } from '../context/AuthContext';
import { useToast } from '../context/ToastContext';
import { isNotFoundError } from '../lib/error-utils';

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
    const [loadFailed, setLoadFailed] = useState(false);
    // The category has no details record. Clips may still be tracked for it.
    const [gameMissing, setGameMissing] = useState(false);
    const [clipsLoaded, setClipsLoaded] = useState(false);
    const [reloadCount, setReloadCount] = useState(0);

    const sort = (searchParams.get('sort') as GameSort | null) || 'hot';
    const timeframe = searchParams.get('timeframe') as GameTimeframe | null;
    const page = parseInt(searchParams.get('page') || '1', 10);
    const [hasMore, setHasMore] = useState(false);

    useEffect(() => {
        const fetchGame = async () => {
            if (!gameId) return;

            try {
                setLoading(true);
                setLoadFailed(false);
                setGameMissing(false);
                const data = await twitchCategoryApi.getGame(gameId);
                setGame(data.game);
                setFollowing(data.game.is_following);
            } catch (err) {
                setGame(null);
                if (isNotFoundError(err)) {
                    setGameMissing(true);
                } else {
                    console.error('Failed to fetch game:', err);
                    setLoadFailed(true);
                }
            } finally {
                setLoading(false);
            }
        };

        fetchGame();
    }, [gameId, reloadCount]);

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

                const data = await twitchCategoryApi.getGameClips(gameId, params);
                setClips(data.clips || []);
                setHasMore(data.has_more);
            } catch (err) {
                setClips([]);
                setHasMore(false);
                if (!isNotFoundError(err)) console.error('Failed to fetch clips:', err);
            } finally {
                setClipsLoading(false);
                setClipsLoaded(true);
            }
        };

        fetchClips();
    }, [gameId, sort, timeframe, page, reloadCount]);

    const handleFollow = async () => {
        if (!user) {
            showToast('Please sign in to follow Twitch categories', 'error');
            return;
        }

        if (!gameId) return;

        try {
            if (following) {
                await twitchCategoryApi.unfollowGame(gameId);
                setFollowing(false);
                showToast('Unfollowed Twitch category', 'success');
                if (game) {
                    setGame({
                        ...game,
                        follower_count: game.follower_count - 1,
                    });
                }
            } else {
                await twitchCategoryApi.followGame(gameId);
                setFollowing(true);
                showToast('Following Twitch category', 'success');
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

    // A missing category is only "not found" when no clips are tracked for it.
    const decidingMissing = gameMissing && (!clipsLoaded || clipsLoading) && clips.length === 0;

    if (loading || decidingMissing) {
        return (
            <Container className='py-8'>
                <div className='flex items-center justify-center min-h-[400px]'>
                    <Spinner size='xl' />
                </div>
            </Container>
        );
    }

    if (loadFailed) {
        return (
            <Container className='py-8'>
                <ResourceUnavailable
                    kind='error'
                    title="We couldn't load this Twitch category"
                    description='Check your connection and try again.'
                    onRetry={() => setReloadCount(count => count + 1)}
                    links={[{ label: 'Back to the feed', href: '/' }]}
                />
            </Container>
        );
    }

    if (!game && clips.length === 0) {
        return (
            <Container className='py-8'>
                <ResourceUnavailable
                    kind='not-found'
                    title="This Twitch category isn't here"
                    description="clpr hasn't tracked any clips from this Twitch category yet."
                    links={[
                        { label: 'Browse topics', href: '/topics' },
                        { label: 'Search clips', href: '/search' },
                        { label: 'Back to the feed', href: '/' },
                    ]}
                />
            </Container>
        );
    }

    return (
        <Container className='py-8'>
            {/* Twitch Category Header */}
            {!game ?
                <div className='mb-8'>
                    <p className='text-xs font-semibold uppercase tracking-wider text-primary-400 mb-1'>
                        Twitch Category
                    </p>
                    <h1 className='text-4xl font-bold mb-2'>Tracked clips</h1>
                    <p className='text-sm text-muted-foreground'>
                        Details for this Twitch category aren't available yet.
                        These are the clips clpr has tracked from it.
                    </p>
                </div>
            :   <div className='mb-8'>
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
                        <p className='text-xs font-semibold uppercase tracking-wider text-primary-400 mb-1'>
                            Twitch Category
                        </p>
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
            }

            {/* Sort and Filter Controls */}
            <div className='mb-6 flex flex-wrap gap-4 items-center'>
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
                    <p className='text-lg'>No clips found for this Twitch category</p>
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
