import { Trophy, Tv, Zap } from 'lucide-react';
import { useCallback, useEffect, useState } from 'react';
import { Link, useSearchParams } from 'react-router-dom';
import {
    LeaderboardSummary,
    LeaderboardTable,
} from '../components/reputation/LeaderboardTable';
import { LeaderboardSkeleton, EmptyStateWithAction } from '../components/ui';
import { useAuth } from '../context/AuthContext';
import { apiClient } from '../lib/api';
import type { LeaderboardResponse, LeaderboardType } from '../types/reputation';

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

export default function LeaderboardPage() {
    const [searchParams, setSearchParams] = useSearchParams();
    const { user } = useAuth();
    const [leaderboard, setLeaderboard] = useState<LeaderboardResponse | null>(
        null
    );
    const [streamerRankings, setStreamerRankings] = useState<StreamerRanking[]>([]);
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState<string | null>(null);

    const type = (searchParams.get('type') as LeaderboardType) || 'karma';
    const page = parseInt(searchParams.get('page') || '1', 10);
    const limit = 50;

    const fetchLeaderboard = useCallback(async () => {
        try {
            setLoading(true);
            setError(null);

            if (type === 'streamers') {
                const res = await apiClient.get<{
                    success: boolean;
                    data: StreamerRanking[];
                }>('/broadcasters/rankings?limit=100');
                setStreamerRankings(res.data?.data || []);
                setLeaderboard(null);
            } else {
                const response = await fetch(
                    `/api/v1/leaderboards/${type}?page=${page}&limit=${limit}`
                );

                // Handle non-OK responses
                if (!response.ok) {
                    let errorMessage = 'Failed to fetch leaderboard';

                    // Try to parse JSON error response
                    try {
                        const errorData = await response.json();
                        errorMessage = errorData.message || errorData.error || errorMessage;
                    } catch {
                        // If JSON parsing fails, use status text
                        errorMessage = response.statusText || errorMessage;
                    }

                    throw new Error(errorMessage);
                }

                // Ensure response is JSON
                const contentType = response.headers.get('content-type');
                if (!contentType || !contentType.includes('application/json')) {
                    throw new Error('Invalid response format from server');
                }

                const data = await response.json();
                setLeaderboard(data);
                setStreamerRankings([]);
            }
        } catch (err) {
            console.error('Leaderboard fetch error:', err);
            setError(err instanceof Error ? err.message : 'An error occurred');
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

    return (
        <div className='max-w-6xl mx-auto px-4 py-8'>
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
            <div className='flex gap-2 mb-6'>
                <button
                    onClick={() => handleTypeChange('karma')}
                    className={`flex items-center gap-2 px-6 py-3 rounded-lg font-semibold transition-colors ${
                        type === 'karma'
                            ? 'bg-purple-600 text-white'
                            : 'bg-surface text-muted-foreground hover:bg-surface-hover'
                    }`}
                >
                    <Trophy size={20} /> Karma
                </button>
                <button
                    onClick={() => handleTypeChange('engagement')}
                    className={`flex items-center gap-2 px-6 py-3 rounded-lg font-semibold transition-colors ${
                        type === 'engagement'
                            ? 'bg-purple-600 text-white'
                            : 'bg-surface text-muted-foreground hover:bg-surface-hover'
                    }`}
                >
                    <Zap size={20} /> Engagement
                </button>
                <button
                    onClick={() => handleTypeChange('streamers')}
                    className={`flex items-center gap-2 px-6 py-3 rounded-lg font-semibold transition-colors ${
                        type === 'streamers'
                            ? 'bg-purple-600 text-white'
                            : 'bg-surface text-muted-foreground hover:bg-surface-hover'
                    }`}
                >
                    <Tv size={20} /> Streamers
                </button>
            </div>

            {/* Loading State */}
            {loading && (
                <LeaderboardSkeleton />
            )}

            {/* Error State */}
            {error && (
                <EmptyStateWithAction
                    icon={
                        <svg className="w-16 h-16" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
                        </svg>
                    }
                    title="Failed to load leaderboard"
                    description={error}
                    primaryAction={{
                        label: "Try Again",
                        onClick: fetchLeaderboard
                    }}
                    secondaryAction={{
                        label: "Go Home",
                        href: "/"
                    }}
                />
            )}

            {/* Leaderboard Content */}
            {!loading && !error && type === 'streamers' && streamerRankings.length > 0 && (
                <div className='bg-surface rounded-xl border border-border overflow-hidden'>
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
                                            {streamer.engagement_score.toLocaleString()}
                                        </td>
                                        <td className='px-4 py-3 text-right text-text-secondary'>
                                            {streamer.total_clips.toLocaleString()}
                                        </td>
                                        <td className='px-4 py-3 text-right text-text-secondary'>
                                            {streamer.total_comments.toLocaleString()}
                                        </td>
                                        <td className='px-4 py-3 text-right text-text-secondary'>
                                            {streamer.follower_count.toLocaleString()}
                                        </td>
                                    </tr>
                                );
                            })}
                        </tbody>
                    </table>
                </div>
            )}

            {!loading && !error && type === 'streamers' && streamerRankings.length === 0 && (
                <div className='text-center text-muted-foreground py-12'>
                    <p>No streamer rankings available yet.</p>
                </div>
            )}

            {!loading && !error && type !== 'streamers' && leaderboard && (
                <>
                    {/* Top 3 Summary */}
                    {page === 1 && (
                        <LeaderboardSummary
                            entries={leaderboard.entries}
                            type={type}
                        />
                    )}

                    {/* Leaderboard Table */}
                    <LeaderboardTable
                        entries={leaderboard.entries}
                        type={type}
                        currentUserId={user?.id}
                    />

                    {/* Pagination */}
                    {leaderboard.entries.length === limit && (
                        <div className='flex justify-center gap-4 mt-6'>
                            <button
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
                                onClick={() => handlePageChange(page + 1)}
                                disabled={leaderboard.entries.length < limit}
                                className='px-6 py-3 bg-surface text-white rounded-lg font-semibold disabled:opacity-50 disabled:cursor-not-allowed hover:bg-surface-hover transition-colors'
                            >
                                Next
                            </button>
                        </div>
                    )}
                </>
            )}
        </div>
    );
}
