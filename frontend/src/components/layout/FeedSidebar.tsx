import { Link } from 'react-router-dom';
import { useQuery } from '@tanstack/react-query';
import { useFeaturedPlaylists, usePlaylists } from '@/hooks/usePlaylist';
import { useTags } from '@/hooks/useTags';
import { useDiscoveryLists } from '@/hooks/useDiscoveryLists';
import { useQueueCount } from '@/hooks/useQueue';
import { useIsAuthenticated } from '@/hooks';
import { gameApi } from '@/lib/game-api';
import { categoryApi } from '@/lib/category-api';
import { apiClient } from '@/lib/api';
import {
    ChevronRight,
    ListMusic,
    Tag,
    Gamepad2,
    Sparkles,
    Bookmark,
    List,
    Flame,
    Users,
    MessageSquare,
} from 'lucide-react';

function SidebarSection({
    title,
    viewAllHref,
    icon: Icon,
    children,
}: {
    title: string;
    viewAllHref?: string;
    icon: React.ElementType;
    children: React.ReactNode;
}) {
    return (
        <div className='bg-surface rounded-lg p-4'>
            <div className='flex items-center justify-between mb-3'>
                <div className='flex items-center gap-2 text-[13px] font-semibold text-text-secondary uppercase tracking-wide'>
                    <Icon className='h-3.5 w-3.5' />
                    {title}
                </div>
                {viewAllHref && (
                    <Link
                        to={viewAllHref}
                        className='text-[11px] text-text-tertiary hover:text-brand transition-colors flex items-center gap-0.5 cursor-pointer'
                    >
                        All
                        <ChevronRight className='h-3 w-3' />
                    </Link>
                )}
            </div>
            {children}
        </div>
    );
}

function SidebarLink({
    to,
    children,
    meta,
}: {
    to: string;
    children: React.ReactNode;
    meta?: React.ReactNode;
}) {
    return (
        <Link
            to={to}
            className='flex items-center justify-between py-1.5 text-[13px] text-text-primary hover:text-brand transition-colors cursor-pointer'
        >
            <span className='truncate'>{children}</span>
            {meta && (
                <span className='text-[11px] text-text-tertiary shrink-0 ml-2'>
                    {meta}
                </span>
            )}
        </Link>
    );
}

export function FeedSidebar() {
    const isAuthenticated = useIsAuthenticated();

    // Data hooks
    const { data: featuredResponse } = useFeaturedPlaylists(1, 5);
    const featuredPlaylists = featuredResponse?.data ?? [];

    const { data: tagsResponse } = useTags({
        sort: 'popularity',
        limit: 10,
    });
    const tags = tagsResponse?.tags ?? [];

    const { data: gamesResponse } = useQuery({
        queryKey: ['games', 'trending', 5],
        queryFn: () => gameApi.getTrendingGames({ limit: 5 }),
    });
    const trendingGames = gamesResponse?.games ?? [];

    const { data: discoveryResponse } = useDiscoveryLists({ limit: 5 });
    const discoveryLists = (discoveryResponse as { lists?: Array<{ id: string; title: string; slug?: string; clip_count?: number }> })?.lists ?? [];

    const { data: topicsResponse } = useQuery({
        queryKey: ['categories', 'topic'],
        queryFn: () => categoryApi.listCategories({ type: 'topic' }),
    });
    const topics = (topicsResponse as { categories?: Array<{ id: string; name: string; slug: string; icon?: string }> })?.categories ?? [];

    const { data: rankingsResponse } = useQuery({
        queryKey: ['broadcasters', 'rankings', 5],
        queryFn: async () => {
            const res = await apiClient.get<{ success: boolean; data: Array<{ broadcaster_id: string; broadcaster_name: string; engagement_score: number }> }>('/broadcasters/rankings?limit=5');
            return res.data;
        },
    });
    const topStreamers = rankingsResponse?.data ?? [];

    const { data: queueCount } = useQueueCount(isAuthenticated);

    const { data: myPlaylistsResponse } = usePlaylists(1, 5);
    const myPlaylists = isAuthenticated ? (myPlaylistsResponse?.data ?? []) : [];

    return (
        <>
            {/* Trending Playlists */}
            {featuredPlaylists.length > 0 && (
                <SidebarSection
                    title='Trending'
                    icon={Sparkles}
                    viewAllHref='/discover/lists'
                >
                    <div className='space-y-0.5'>
                        {featuredPlaylists.map(playlist => (
                            <SidebarLink
                                key={playlist.id}
                                to={`/playlists/${playlist.id}`}
                                meta={
                                    playlist.clip_count !== undefined
                                        ? `${playlist.clip_count} clips`
                                        : undefined
                                }
                            >
                                {playlist.title}
                            </SidebarLink>
                        ))}
                    </div>
                </SidebarSection>
            )}

            {/* Popular Tags */}
            {tags.length > 0 && (
                <SidebarSection title='Tags' icon={Tag} viewAllHref='/tags'>
                    <div className='flex flex-wrap gap-1.5'>
                        {tags.map(tag => (
                            <Link
                                key={tag.id}
                                to={`/tag/${tag.slug}`}
                                className='inline-flex items-center gap-1 px-2 py-1 text-[11px] font-medium bg-surface-raised hover:bg-surface-hover rounded-full text-text-secondary hover:text-text-primary transition-colors cursor-pointer'
                            >
                                {tag.name}
                                <span className='text-text-tertiary'>
                                    {tag.usage_count}
                                </span>
                            </Link>
                        ))}
                    </div>
                </SidebarSection>
            )}

            {/* Top Games */}
            {trendingGames.length > 0 && (
                <SidebarSection
                    title='Games'
                    icon={Gamepad2}
                    viewAllHref='/games'
                >
                    <div className='space-y-0.5'>
                        {trendingGames.map(game => (
                            <SidebarLink
                                key={game.id}
                                to={`/game/${game.id}`}
                                meta={`${game.recent_clip_count} clips`}
                            >
                                {game.name}
                            </SidebarLink>
                        ))}
                    </div>
                </SidebarSection>
            )}

            {/* Topics (custom categories) */}
            {topics.length > 0 && (
                <SidebarSection title='Topics' icon={Flame} viewAllHref='/categories'>
                    <div className='flex flex-wrap gap-1.5'>
                        {topics.map(topic => (
                            <Link
                                key={topic.id}
                                to={`/category/${topic.slug}`}
                                className='inline-flex items-center gap-1 px-2.5 py-1 text-[12px] font-medium bg-surface-raised hover:bg-surface-hover rounded text-text-secondary hover:text-text-primary transition-colors cursor-pointer'
                            >
                                {topic.name}
                            </Link>
                        ))}
                    </div>
                </SidebarSection>
            )}

            {/* Top Streamers */}
            {topStreamers.length > 0 && (
                <SidebarSection
                    title='Streamers'
                    icon={Users}
                    viewAllHref='/leaderboard'
                >
                    <div className='space-y-0.5'>
                        {topStreamers.map((streamer, idx) => (
                            <SidebarLink
                                key={streamer.broadcaster_id}
                                to={`/broadcaster/${streamer.broadcaster_id}`}
                                meta={`#${idx + 1}`}
                            >
                                {streamer.broadcaster_name}
                            </SidebarLink>
                        ))}
                    </div>
                </SidebarSection>
            )}

            {/* Forum */}
            <SidebarSection title='Forum' icon={MessageSquare} viewAllHref='/forum'>
                <div className='space-y-0.5'>
                    <SidebarLink to='/forum'>Browse Discussions</SidebarLink>
                    <SidebarLink to='/forum/new'>Start a Thread</SidebarLink>
                </div>
            </SidebarSection>

            {/* Discovery Lists */}
            {discoveryLists.length > 0 && (
                <SidebarSection
                    title='Discovery'
                    icon={List}
                    viewAllHref='/discover'
                >
                    <div className='space-y-0.5'>
                        {discoveryLists.map(list => (
                            <SidebarLink
                                key={list.id}
                                to={`/discover/${list.slug || list.id}`}
                                meta={
                                    list.clip_count !== undefined
                                        ? `${list.clip_count} clips`
                                        : undefined
                                }
                            >
                                {list.title}
                            </SidebarLink>
                        ))}
                    </div>
                </SidebarSection>
            )}

            {/* Your Queue (authenticated) */}
            {isAuthenticated && queueCount !== undefined && queueCount > 0 && (
                <SidebarSection title='Queue' icon={ListMusic}>
                    <Link
                        to='/queue'
                        className='flex items-center justify-between py-1.5 text-[13px] text-text-primary hover:text-brand transition-colors cursor-pointer'
                    >
                        <span>
                            {queueCount} {queueCount === 1 ? 'clip' : 'clips'}{' '}
                            saved
                        </span>
                        <ChevronRight className='h-3.5 w-3.5 text-text-tertiary' />
                    </Link>
                </SidebarSection>
            )}

            {/* Your Playlists (authenticated) */}
            {isAuthenticated && myPlaylists.length > 0 && (
                <SidebarSection
                    title='My Playlists'
                    icon={Bookmark}
                    viewAllHref='/playlists'
                >
                    <div className='space-y-0.5'>
                        {myPlaylists.map(playlist => (
                            <SidebarLink
                                key={playlist.id}
                                to={`/playlists/${playlist.id}`}
                                meta={
                                    playlist.clip_count !== undefined
                                        ? `${playlist.clip_count}`
                                        : undefined
                                }
                            >
                                {playlist.title}
                            </SidebarLink>
                        ))}
                    </div>
                </SidebarSection>
            )}
        </>
    );
}
