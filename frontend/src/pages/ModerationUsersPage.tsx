import { useState } from 'react';
import { useQuery, useQueryClient } from '@tanstack/react-query';
import {
    Container,
    Card,
    CardHeader,
    CardBody,
    Button,
    Input,
} from '../components';
import { Search, Ban } from 'lucide-react';
import { TwitchModerationActions } from '../components/moderation/TwitchModerationActions';
import { useAuth } from '../context/AuthContext';
import axios from 'axios';

interface User {
    id: string;
    username: string;
    display_name: string;
    email: string;
    avatar_url: string;
    role: string;
    twitch_id?: string;
    is_banned_on_twitch?: boolean;
    created_at: string;
}

interface UsersResponse {
    users: User[];
    total: number;
    page: number;
    per_page: number;
}

export function ModerationUsersPage() {
    const [search, setSearch] = useState('');
    const [page, setPage] = useState(1);
    const [actionsLocked, setActionsLocked] = useState(false);
    const { user: currentUser } = useAuth();

    const queryClient = useQueryClient();
    const perPage = 25;

    // Fetch users
    const { data, isLoading, error } = useQuery<UsersResponse>({
        queryKey: ['moderation-users', search, page],
        queryFn: async () => {
            const params = new URLSearchParams({
                page: page.toString(),
                per_page: perPage.toString(),
            });
            if (search) params.append('search', search);

            const response = await axios.get(
                `/api/v1/admin/users?${params.toString()}`
            );
            return response.data;
        },
    });

    const handleActionSuccess = () => {
        // Refresh user list after ban/unban action
        queryClient.invalidateQueries({ queryKey: ['moderation-users'] });
    };

    // This global user list has no channel-scoped Twitch permission evidence.
    // Keep Twitch actions unavailable until a channel supplies explicit permissions.
    const isBroadcaster = false;
    const isTwitchModerator = false;

    const users = data?.users || [];
    const totalPages = data?.total ? Math.ceil(data.total / perPage) : 0;

    return (
        <Container className='py-8'>
            <div className='max-w-6xl mx-auto'>
                <div className='mb-6'>
                    <h1 className='text-3xl font-bold mb-2'>
                        Moderation - Users
                    </h1>
                    <p className='text-muted-foreground'>
                        Manage Twitch bans and timeouts for users
                    </p>
                </div>

                {/* Search */}
                <Card className='mb-6'>
                    <CardBody>
                        <div className='flex gap-4'>
                            <div className='flex-1'>
                                <div className='relative'>
                                    <Search className='absolute left-3 top-1/2 transform -translate-y-1/2 h-4 w-4 text-muted-foreground' />
                                    <Input
                                        type='text'
                                        placeholder='Search users...'
                                        value={search}
                                        onChange={e => {
                                            setSearch(e.target.value);
                                            setPage(1);
                                        }}
                                        className='pl-10'
                                    />
                                </div>
                            </div>
                        </div>
                    </CardBody>
                </Card>

                {/* Error state */}
                {error && (
                    <Card className='mb-6 border-red-500'>
                        <CardBody>
                            <p className='text-red-600'>
                                Failed to load users. Please try again.
                            </p>
                        </CardBody>
                    </Card>
                )}

                {/* Loading state */}
                {isLoading && (
                    <Card>
                        <CardBody>
                            <p className='text-center text-muted-foreground'>
                                Loading users...
                            </p>
                        </CardBody>
                    </Card>
                )}

                {/* Users list */}
                {!isLoading && users.length > 0 && (
                    <Card>
                        <CardHeader>
                            <h2 className='text-xl font-semibold'>
                                Users ({data?.total || 0})
                            </h2>
                        </CardHeader>
                        <CardBody>
                            <div className='space-y-4'>
                                {users.map(user => (
                                    <div
                                        key={user.id}
                                        className='flex items-center justify-between p-4 border border-border rounded-lg'
                                    >
                                        <div className='flex items-center gap-4'>
                                            <img
                                                src={
                                                    user.avatar_url ||
                                                    '/default-avatar.png'
                                                }
                                                alt={user.username}
                                                className='w-12 h-12 rounded-full'
                                            />
                                            <div>
                                                <div className='flex items-center gap-2'>
                                                    <h3 className='font-semibold'>
                                                        {user.display_name ||
                                                            user.username}
                                                    </h3>
                                                    {user.is_banned_on_twitch && (
                                                        <span className='inline-flex items-center gap-1 px-2 py-0.5 text-xs font-medium bg-red-100 dark:bg-red-900/20 text-red-800 dark:text-red-300'>
                                                            <Ban className='h-3 w-3' />
                                                            Banned on Twitch
                                                        </span>
                                                    )}
                                                </div>
                                                <p className='text-sm text-muted-foreground'>
                                                    @{user.username} •{' '}
                                                    {user.email}
                                                </p>
                                                {user.twitch_id && (
                                                    <p className='text-xs text-muted-foreground'>
                                                        Twitch ID:{' '}
                                                        {user.twitch_id}
                                                    </p>
                                                )}
                                            </div>
                                        </div>

                                        <div className='flex items-center gap-2'>
                                            {user.twitch_id &&
                                                currentUser?.twitch_id && (
                                                    <TwitchModerationActions
                                                        broadcasterID={
                                                            user.twitch_id
                                                        }
                                                        userID={user.twitch_id}
                                                        username={user.username}
                                                        isBanned={
                                                            user.is_banned_on_twitch
                                                        }
                                                        isBroadcaster={
                                                            isBroadcaster
                                                        }
                                                        isTwitchModerator={
                                                            isTwitchModerator
                                                        }
                                                        onSuccess={
                                                            handleActionSuccess
                                                        }
                                                        disableTriggers={
                                                            actionsLocked
                                                        }
                                                        onModalOpen={() =>
                                                            setActionsLocked(
                                                                true
                                                            )
                                                        }
                                                        onModalClose={() =>
                                                            setActionsLocked(
                                                                false
                                                            )
                                                        }
                                                    />
                                                )}
                                        </div>
                                    </div>
                                ))}
                            </div>
                        </CardBody>
                    </Card>
                )}

                {/* Empty state */}
                {!isLoading && users.length === 0 && (
                    <Card>
                        <CardBody>
                            <p className='text-center text-muted-foreground py-8'>
                                No users found.
                            </p>
                        </CardBody>
                    </Card>
                )}

                {/* Pagination */}
                {totalPages > 1 && (
                    <div className='mt-6 flex justify-center gap-2'>
                        <Button
                            variant='outline'
                            onClick={() => setPage(Math.max(1, page - 1))}
                            disabled={page === 1}
                        >
                            Previous
                        </Button>
                        <span className='px-4 py-2'>
                            Page {page} of {totalPages}
                        </span>
                        <Button
                            variant='outline'
                            onClick={() =>
                                setPage(Math.min(totalPages, page + 1))
                            }
                            disabled={page === totalPages}
                        >
                            Next
                        </Button>
                    </div>
                )}
            </div>
        </Container>
    );
}
