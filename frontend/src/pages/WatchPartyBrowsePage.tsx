import { useEffect, useState, useRef } from 'react';
import { useNavigate } from 'react-router-dom';
import { Container, ScrollToTop, SEO } from '../components';
import { MiniFooter } from '../components/layout';
import { Button } from '../components/ui/Button';
import {
    getPublicWatchParties,
    getTrendingWatchParties,
    joinWatchParty,
} from '../lib/watch-party-api';
import { useToast } from '../context/ToastContext';
import type { WatchParty } from '../types/watchParty';
import { Users, Play, TrendingUp, Plus } from 'lucide-react';

export function WatchPartyBrowsePage() {
    const navigate = useNavigate();
    const { showToast } = useToast();
    const inviteCodeInputRef = useRef<HTMLInputElement>(null);

    const [trendingParties, setTrendingParties] = useState<WatchParty[]>([]);
    const [publicParties, setPublicParties] = useState<WatchParty[]>([]);
    const [loading, setLoading] = useState(true);
    const [joiningParty, setJoiningParty] = useState<string | null>(null);

    useEffect(() => {
        const loadParties = async () => {
            try {
                setLoading(true);
                const [trending, publicData] = await Promise.all([
                    getTrendingWatchParties(10),
                    getPublicWatchParties(20, 0),
                ]);
                setTrendingParties(trending);
                setPublicParties(publicData.parties ?? []);
            } catch (err) {
                console.error('Error loading parties:', err);
                showToast('Failed to load watch parties', 'error');
            } finally {
                setLoading(false);
            }
        };

        loadParties();
    }, [showToast]);

    const handleJoinParty = async (inviteCode: string) => {
        try {
            setJoiningParty(inviteCode);
            const party = await joinWatchParty(inviteCode);
            showToast('Joined watch party!', 'success');
            navigate(`/watch-parties/${party.id}`);
        } catch (err) {
            console.error('Error joining party:', err);
            showToast(
                err instanceof Error ?
                    err.message
                :   'Failed to join watch party',
                'error',
            );
        } finally {
            setJoiningParty(null);
        }
    };

    const handleCreateParty = () => {
        navigate('/watch-parties/create');
    };

    const PartyCard = ({ party }: { party: WatchParty }) => {
        const participantCount =
            party.active_participant_count || party.participants?.length || 0;
        const isJoining = joiningParty === party.invite_code;

        return (
            <div className='bg-surface-secondary rounded-lg p-4 hover:bg-surface-tertiary transition-colors'>
                <div className='flex items-start justify-between mb-3'>
                    <div className='flex-1'>
                        <h3 className='text-lg font-semibold text-content-primary mb-1'>
                            {party.title}
                        </h3>
                        <div className='flex items-center gap-3 text-sm text-content-secondary'>
                            <div className='flex items-center gap-1'>
                                <Users className='w-4 h-4' />
                                <span>{participantCount} watching</span>
                            </div>
                            {party.is_playing && (
                                <div className='flex items-center gap-1 text-success-600'>
                                    <Play className='w-4 h-4' />
                                    <span>Playing</span>
                                </div>
                            )}
                        </div>
                    </div>
                    <Button
                        size='sm'
                        onClick={() => handleJoinParty(party.invite_code)}
                        disabled={
                            isJoining ||
                            participantCount >= party.max_participants
                        }
                        isLoading={isJoining}
                    >
                        {participantCount >= party.max_participants ?
                            'Full'
                        :   'Join'}
                    </Button>
                </div>

                {party.playlist_id && (
                    <p className='text-xs text-content-secondary'>
                        Playlist party
                    </p>
                )}
            </div>
        );
    };

    if (loading) {
        return (
            <>
                <SEO
                    title='Browse Watch Parties'
                    description='Discover and join watch parties'
                />
                <Container>
                    <div className='py-8 flex items-center justify-center'>
                        <div className='text-center'>
                            <div className='animate-spin rounded-full h-12 w-12 border-b-2 border-primary-500 mx-auto'></div>
                            <p className='mt-4 text-content-secondary'>
                                Loading watch parties...
                            </p>
                        </div>
                    </div>
                </Container>
            </>
        );
    }

    return (
        <>
            <SEO
                title='Browse Watch Parties'
                description='Discover and join watch parties with friends and community'
            />
            <Container>
                <div className='py-8'>
                    {/* Header */}
                    <div className='flex items-center justify-between mb-6'>
                        <div>
                            <h1 className='text-3xl font-bold text-content-primary mb-2'>
                                Watch Parties
                            </h1>
                            <p className='text-content-secondary'>
                                Watch clips together with friends and community
                            </p>
                        </div>
                        <Button onClick={handleCreateParty}>
                            <Plus className='w-4 h-4 mr-2' />
                            Create Party
                        </Button>
                    </div>

                    {/* Trending parties */}
                    {trendingParties.length > 0 && (
                        <section className='mb-8'>
                            <div className='flex items-center gap-2 mb-4'>
                                <TrendingUp className='w-5 h-5 text-primary-500' />
                                <h2 className='text-xl font-semibold text-content-primary'>
                                    Trending Now
                                </h2>
                            </div>
                            <div className='grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-6'>
                                {trendingParties.map(party => (
                                    <PartyCard key={party.id} party={party} />
                                ))}
                            </div>
                        </section>
                    )}

                    {/* All public parties */}
                    <section>
                        <h2 className='text-xl font-semibold text-content-primary mb-4'>
                            Public Watch Parties
                        </h2>
                        {publicParties.length > 0 ?
                            <div className='grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-6'>
                                {publicParties.map(party => (
                                    <PartyCard key={party.id} party={party} />
                                ))}
                            </div>
                        :   <div className='text-center py-12 bg-surface-secondary rounded-lg'>
                                <p className='text-content-secondary mb-4'>
                                    No public watch parties available
                                </p>
                                <Button onClick={handleCreateParty}>
                                    <Plus className='w-4 h-4 mr-2' />
                                    Create the First Party
                                </Button>
                            </div>
                        }
                    </section>

                    {/* Quick actions */}
                    <div className='mt-8 p-6 bg-surface-secondary rounded-lg'>
                        <h3 className='text-lg font-semibold text-content-primary mb-3'>
                            Join with Invite Code
                        </h3>
                        <p className='text-sm text-content-secondary mb-4'>
                            Have an invite code? Enter it below to join a
                            private watch party.
                        </p>
                        <div className='flex gap-2'>
                            <input
                                ref={inviteCodeInputRef}
                                type='text'
                                placeholder='Enter invite code'
                                className='flex-1 px-4 py-2 bg-surface-primary border border-divider rounded-lg text-content-primary placeholder-content-tertiary focus:outline-none focus:border-primary-500'
                                onKeyPress={e => {
                                    if (e.key === 'Enter') {
                                        const code = (
                                            e.target as HTMLInputElement
                                        ).value.trim();
                                        if (code) {
                                            handleJoinParty(code);
                                        }
                                    }
                                }}
                            />
                            <Button
                                onClick={() => {
                                    const code =
                                        inviteCodeInputRef.current?.value.trim();
                                    if (code) {
                                        handleJoinParty(code);
                                    }
                                }}
                            >
                                Join
                            </Button>
                        </div>
                    </div>
                </div>
                <MiniFooter />
            </Container>
            <ScrollToTop />
        </>
    );
}
