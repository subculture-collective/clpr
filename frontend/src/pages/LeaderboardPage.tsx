import { Trophy, Tv, Zap } from 'lucide-react';
import type { ReactNode } from 'react';
import { useCallback, useEffect, useState } from 'react';
import { Link, useSearchParams } from 'react-router-dom';
import {
    LeaderboardSummary,
    LeaderboardTable,
} from '../components/reputation/LeaderboardTable';
import { LeaderboardSkeleton, EmptyStateWithAction } from '../components/ui';
import { SEO } from '../components';
import { useAuth } from '../context/AuthContext';
import { apiClient } from '../lib/api';
import type { LeaderboardEntry, LeaderboardResponse, LeaderboardType } from '../types/reputation';

interface StreamerRanking {
    broadcaster_id: string;
    broadcaster_name: string;
    total_clips: number;
    human_submitted_clips: number;
    total_vote_score: number;
    total_comments: number;
    engagement_score: number;
    follower_count: number;
}

const tabs: Array<{ type: LeaderboardType; label: string; icon: ReactNode }> = [
    { type: 'karma', label: 'Uppies', icon: <Trophy size={20} aria-hidden='true' /> },
    { type: 'engagement', label: 'Engagement', icon: <Zap size={20} aria-hidden='true' /> },
    { type: 'streamers', label: 'Creators', icon: <Tv size={20} aria-hidden='true' /> },
];

const emptyCopy: Record<LeaderboardType, string> = {
    karma: 'Uppies rankings fill in as people vote on clips and comments.',
    engagement: 'Engagement rankings fill in as people vote, comment, and submit clips.',
    streamers: 'Creator rankings fill in as their clips get votes and comments on clpr.',
};

function LeaderboardEmptyState({ type }: { type: LeaderboardType }) {
    return (
        <EmptyStateWithAction
            title='No rankings yet'
            description={emptyCopy[type]}
            primaryAction={{ label: 'Browse clips', href: '/' }}
        />
    );
}

/**
 * The API lists every member, ordered by score. Members who have not earned
 * any points yet would all tie at zero in arbitrary order, so only positive
 * scores are ranked.
 */
function rankedEntriesOf(leaderboard: LeaderboardResponse | null): LeaderboardEntry[] {
    return (leaderboard?.entries ?? []).filter(entry => entry.score > 0);
}

export default function LeaderboardPage() {
    const [searchParams, setSearchParams] = useSearchParams();
    const { user } = useAuth();
    const [leaderboard, setLeaderboard] = useState<LeaderboardResponse | null>(
        null
    );
    const [streamerRankings, setStreamerRankings] = useState<StreamerRanking[]>([]);
    const [loading, setLoading] = useState(true);
    const [failed, setFailed] = useState(false);

    const requestedType = searchParams.get('type');
    const type: LeaderboardType = tabs.some(tab => tab.type === requestedType)
        ? (requestedType as LeaderboardType)
        : 'karma';
    const page = Math.max(1, parseInt(searchParams.get('page') || '1', 10) || 1);
    const limit = 50;

    const fetchLeaderboard = useCallback(async () => {
        try {
            setLoading(true);
            setFailed(false);

            if (type === 'streamers') {
                const res = await apiClient.get<{
                    success: boolean;
                    data: StreamerRanking[] | null;
                }>('/broadcasters/rankings?limit=100');
                setStreamerRankings((res.data?.data || []).filter(item => item.engagement_score > 0));
                setLeaderboard(null);
            } else {
                const response = await fetch(
                    `/api/v1/leaderboards/${type}?page=${page}&limit=${limit}`
                );
                const contentType = response.headers.get('content-type');
                if (!response.ok || !contentType?.includes('application/json')) {
                    throw new Error(`Leaderboard request failed with status ${response.status}`);
                }

                const data = await response.json() as LeaderboardResponse;
                setLeaderboard(data);
                setStreamerRankings([]);
            }
        } catch (err) {
            console.error('Leaderboard fetch error:', err);
            setFailed(true);
        } finally {
            setLoading(false);
        }
    }, [type, page]);

    useEffect(() => {
        fetchLeaderboard();
    }, [fetchLeaderboard]);

    const handleTypeChange = (newType: LeaderboardType) => {
        setSearchParams({ type: newType, page: '1' });
    };

    const handlePageChange = (newPage: number) => {
        setSearchParams({ type, page: newPage.toString() });
    };

    const rankedEntries = rankedEntriesOf(leaderboard);
    const returnedEntries = leaderboard?.entries ?? [];
    // Entries arrive ordered by score, so a zero on this page means every later page is zeros too.
    const hasNextPage =
        returnedEntries.length === limit &&
        rankedEntries.length === returnedEntries.length;
    const showPagination = page > 1 || hasNextPage;
    const displayValue = (value: number) => value > 0 ? value.toLocaleString() : '—';

    return (
        <><SEO title='Leaderboards' description='Community rankings for clpr contributors, clips, and creators.' canonicalUrl='/leaderboards' /><div className='max-w-6xl mx-auto px-4 py-8'>
            {/* Header */}
            <div className='mb-8'>
                <h1 className='text-4xl font-bold text-white mb-2'>
                    Leaderboards
                </h1>
                <p className='text-muted-foreground'>
                    Top contributors in the clpr community
                </p>
            </div>

            {/* Type Selector */}
            <div className='flex flex-wrap gap-2 mb-6' role='group' aria-label='Leaderboard type'>
                {tabs.map(tab => (
                    <button
                        key={tab.type}
                        type='button'
                        onClick={() => handleTypeChange(tab.type)}
                        aria-pressed={type === tab.type}
                        className={`flex min-h-11 items-center gap-2 px-4 py-2.5 rounded-lg font-semibold transition-colors sm:px-6 sm:py-3 ${
                            type === tab.type
                                ? 'bg-primary-400 text-background'
                                : 'bg-surface text-muted-foreground hover:bg-surface-hover'
                        }`}
                    >
                        {tab.icon} {tab.label}
                    </button>
                ))}
            </div>

            {/* Loading State */}
            {loading && (
                <LeaderboardSkeleton />
            )}

            {/* Error State */}
            {!loading && failed && (
                <EmptyStateWithAction
                    icon={
                        <svg className="w-16 h-16" fill="none" stroke="currentColor" viewBox="0 0 24 24" aria-hidden='true'>
                            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
                        </svg>
                    }
                    title="We couldn't load the leaderboard"
                    description='Check your connection and try again.'
                    primaryAction={{
                        label: "Try again",
                        onClick: fetchLeaderboard
                    }}
                    secondaryAction={{
                        label: "Go home",
                        href: "/"
                    }}
                />
            )}

            {/* Leaderboard Content */}
            {!loading && !failed && type === 'streamers' && streamerRankings.length > 0 && (
                <div className='bg-surface rounded-xl border border-border overflow-x-auto'>
                    <table className='w-full'>
                        <thead>
                            <tr className='border-b border-border text-text-secondary text-sm'>
                                <th className='px-4 py-3 text-left font-semibold'>Rank</th>
                                <th className='px-4 py-3 text-left font-semibold'>Streamer</th>
                                <th className='px-4 py-3 text-right font-semibold'>Engagement Score</th>
                                <th className='px-4 py-3 text-right font-semibold'>Clips</th>
                                <th className='px-4 py-3 text-right font-semibold'>Comments</th>
                                <th className='px-4 py-3 text-right font-semibold'>Followers</th>
                            </tr>
                        </thead>
                        <tbody>
                            {streamerRankings.map((streamer, idx) => {
                                const rank = idx + 1;
                                const rankBadge =
                                    rank <= 3 ? (
                                        <span className={`font-accent font-extrabold text-lg ${
                                            rank === 1 ? 'text-yellow-400' :
                                            rank === 2 ? 'text-neutral-300' :
                                            'text-amber-600'
                                        }`}>{rank}</span>
                                    ) : `#${rank}`;
                                return (
                                    <tr
                                        key={streamer.broadcaster_id}
                                        className='border-b border-border last:border-b-0 hover:bg-surface-hover transition-colors'
                                    >
                                        <td className='px-4 py-3 text-text-primary font-semibold text-lg'>
                                            {rankBadge}
                                        </td>
                                        <td className='px-4 py-3'>
                                            <Link
                                                to={`/broadcaster/${streamer.broadcaster_id}`}
                                                className='text-text-primary font-medium hover:text-purple-400 transition-colors'
                                            >
                                                {streamer.broadcaster_name}
                                            </Link>
                                        </td>
                                        <td className='px-4 py-3 text-right text-text-primary font-semibold'>
                                            {displayValue(streamer.engagement_score)}
                                        </td>
                                        <td className='px-4 py-3 text-right text-text-secondary'>
                                            {displayValue(streamer.total_clips)}
                                        </td>
                                        <td className='px-4 py-3 text-right text-text-secondary'>
                                            {displayValue(streamer.total_comments)}
                                        </td>
                                        <td className='px-4 py-3 text-right text-text-secondary'>
                                            {displayValue(streamer.follower_count)}
                                        </td>
                                    </tr>
                                );
                            })}
                        </tbody>
                    </table>
                </div>
            )}

            {!loading && !failed && type === 'streamers' && streamerRankings.length === 0 && (
                <LeaderboardEmptyState type={type} />
            )}

            {!loading && !failed && type !== 'streamers' && leaderboard && (
                <>
                    {rankedEntries.length === 0 ? (
                        page === 1 ? (
                            <LeaderboardEmptyState type={type} />
                        ) : (
                            <EmptyStateWithAction
                                title='No more rankings'
                                description='Everyone with points is on an earlier page.'
                                primaryAction={{ label: 'Back to page 1', onClick: () => handlePageChange(1) }}
                            />
                        )
                    ) : (
                        <>
                            {/* Top 3 Summary */}
                            {page === 1 && (
                                <LeaderboardSummary
                                    entries={rankedEntries}
                                    type={type}
                                />
                            )}

                            {/* Leaderboard Table */}
                            <LeaderboardTable
                                entries={rankedEntries}
                                type={type}
                                currentUserId={user?.id}
                            />
                        </>
                    )}

                    {/* Pagination */}
                    {rankedEntries.length > 0 && showPagination && (
                        <nav className='flex flex-wrap justify-center gap-4 mt-6' aria-label='Leaderboard pages'>
                            <button
                                type='button'
                                onClick={() => handlePageChange(page - 1)}
                                disabled={page === 1}
                                className='px-6 py-3 bg-surface text-white rounded-lg font-semibold disabled:opacity-50 disabled:cursor-not-allowed hover:bg-surface-hover transition-colors'
                            >
                                Previous
                            </button>
                            <div className='px-6 py-3 bg-background text-white rounded-lg font-semibold'>
                                Page {page}
                            </div>
                            <button
                                type='button'
                                onClick={() => handlePageChange(page + 1)}
                                disabled={!hasNextPage}
                                className='px-6 py-3 bg-surface text-white rounded-lg font-semibold disabled:opacity-50 disabled:cursor-not-allowed hover:bg-surface-hover transition-colors'
                            >
                                Next
                            </button>
                        </nav>
                    )}
                </>
            )}
        </div></>
    );
}
