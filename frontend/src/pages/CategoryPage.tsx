import { useState, useEffect } from 'react';
import { useParams, useSearchParams } from 'react-router-dom';
import { Gamepad2 } from 'lucide-react';
import { Container, Spinner, Button, CategoryIcon } from '../components';
import { ClipGridCard } from '../components/clip';
import { categoryApi } from '../lib/category-api';
import type { Category } from '../types/category';
import type { GameWithStats } from '../types/game';
import type { Clip } from '../types/clip';
import { Link } from 'react-router-dom';

type CategorySort = 'hot' | 'new' | 'top' | 'rising';
type CategoryTimeframe = 'hour' | 'day' | 'week' | 'month' | 'year' | 'all';

export function CategoryPage() {
    const { categorySlug } = useParams<{ categorySlug: string }>();
    const [searchParams, setSearchParams] = useSearchParams();

    const [category, setCategory] = useState<Category | null>(null);
    const [games, setGames] = useState<GameWithStats[]>([]);
    const [clips, setClips] = useState<Clip[]>([]);
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState<string | null>(null);

    const sort = (searchParams.get('sort') as CategorySort | null) || 'hot';
    const timeframe = searchParams.get('timeframe') as CategoryTimeframe | null;
    const page = parseInt(searchParams.get('page') || '1', 10);
    const [hasMore, setHasMore] = useState(false);

    useEffect(() => {
        const fetchData = async () => {
            if (!categorySlug) return;

            try {
                setLoading(true);
                setError(null);

                // Fetch category details and games in parallel
                const [categoryData, gamesData, clipsData] = await Promise.all([
                    categoryApi.getCategory(categorySlug),
                    categoryApi.getCategoryGames(categorySlug, { limit: 10 }),
                    categoryApi.getCategoryClips(categorySlug, {
                        page,
                        limit: 20,
                        sort,
                        timeframe: timeframe || undefined,
                    }),
                ]);

                setCategory(categoryData.category);
                setGames(gamesData.games || []);
                setClips(clipsData.clips || []);
                setHasMore(clipsData.has_more);
            } catch (err) {
                console.error('Failed to fetch category data:', err);
                setError('Failed to load category');
            } finally {
                setLoading(false);
            }
        };

        fetchData();
    }, [categorySlug, sort, timeframe, page]);

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

    if (error || !category) {
        return (
            <Container className='py-8'>
                <div className='text-center text-muted-foreground py-12'>
                    <p className='text-lg'>{error || 'Category not found'}</p>
                </div>
            </Container>
        );
    }

    return (
        <Container className='py-8'>
            {/* Category Header */}
            <div className='mb-8'>
                <div className='flex items-center gap-3 mb-2'>
                    <CategoryIcon icon={category.icon} size='lg' />
                    <h1 className='text-4xl font-bold'>{category.name}</h1>
                </div>
                {category.description && (
                    <p className='text-lg text-muted-foreground'>
                        {category.description}
                    </p>
                )}
            </div>

            {/* Featured Games */}
            {games.length > 0 && (
                <div className='mb-8'>
                    <h2 className='text-2xl font-bold mb-4'>Popular Games</h2>
                    <div className='grid grid-cols-2 sm:grid-cols-3 md:grid-cols-5 gap-4'>
                        {games.map(game => (
                            <Link
                                key={game.id}
                                to={`/game/${game.twitch_game_id}`}
                                className='group'
                            >
                                <div className='aspect-3/4 relative rounded-lg overflow-hidden shadow-lg transition-transform group-hover:scale-105'>
                                    {game.box_art_url ? (
                                        <img
                                            src={game.box_art_url
                                                .replace('{width}', '285')
                                                .replace('{height}', '380')}
                                            alt={game.name}
                                            className='w-full h-full object-cover'
                                        />
                                    ) : (
                                        <div className='w-full h-full bg-surface flex items-center justify-center'>
                                            <Gamepad2 size={16} strokeWidth={1.75} />
                                        </div>
                                    )}
                                    <div className='absolute bottom-0 left-0 right-0 bg-linear-to-t from-black/80 to-transparent p-2'>
                                        <h3 className='text-sm font-semibold text-white truncate'>
                                            {game.name}
                                        </h3>
                                        <p className='text-xs text-foreground'>
                                            {game.clip_count} clips
                                        </p>
                                    </div>
                                </div>
                            </Link>
                        ))}
                    </div>
                </div>
            )}

            {/* Sort and Filter Controls */}
            <div className='mb-6 flex gap-4 items-center'>
                <h2 className='text-2xl font-bold'>Recent Clips</h2>
                <div className='ml-auto flex gap-4'>
                    <div>
                        <label className='text-sm font-medium mr-2'>
                            Sort by:
                        </label>
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
            </div>

            {/* Clips Grid */}
            {clips.length === 0 ? (
                <div className='text-center text-muted-foreground py-12'>
                    <p className='text-lg'>No clips found in this category</p>
                </div>
            ) : (
                <>
                    <div className='grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6 mb-8'>
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
            )}
        </Container>
    );
}
