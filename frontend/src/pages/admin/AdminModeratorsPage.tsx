import { useState } from 'react';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { Button, Container, Input, Spinner } from '../../components';
import { platformModeratorApi } from '../../lib/platform-moderator-api';

export function AdminModeratorsPage() {
    const queryClient = useQueryClient();
    const [userId, setUserId] = useState('');
    const [reason, setReason] = useState('');
    const [error, setError] = useState<string | null>(null);
    const query = useQuery({
        queryKey: ['platform-moderators'],
        queryFn: () => platformModeratorApi.list(),
    });
    const refresh = () =>
        queryClient.invalidateQueries({ queryKey: ['platform-moderators'] });
    const add = useMutation({
        mutationFn: () =>
            platformModeratorApi.add(userId.trim(), reason.trim()),
        onSuccess: () => {
            setUserId('');
            setReason('');
            setError(null);
            void refresh();
        },
        onError: () => setError('Could not add the platform moderator.'),
    });
    const update = useMutation({
        mutationFn: (id: string) =>
            platformModeratorApi.update(
                id,
                'Platform moderator access reviewed',
            ),
        onSuccess: () => {
            setError(null);
            void refresh();
        },
        onError: () => setError('Could not update the platform moderator.'),
    });
    const revoke = useMutation({
        mutationFn: (id: string) =>
            platformModeratorApi.revoke(
                id,
                'Platform moderator access revoked',
            ),
        onSuccess: () => {
            setError(null);
            void refresh();
        },
        onError: () => setError('Could not revoke the platform moderator.'),
    });

    return (
        <Container className="py-8">
            <h1 className="text-3xl font-bold">Platform Moderators</h1>
            <p className="mt-2 text-muted-foreground">
                Manage site-wide moderator access independently from chat
                channel roles.
            </p>
            <form
                className="mt-6 grid gap-3 rounded-lg border border-border p-4 md:grid-cols-[1fr_1fr_auto]"
                onSubmit={(event) => {
                    event.preventDefault();
                    setError(null);
                    add.mutate();
                }}
            >
                <Input
                    aria-label="User ID"
                    placeholder="User UUID"
                    value={userId}
                    onChange={(event) => setUserId(event.target.value)}
                    required
                />
                <Input
                    aria-label="Reason"
                    placeholder="Reason (optional)"
                    value={reason}
                    onChange={(event) => setReason(event.target.value)}
                />
                <Button type="submit" disabled={add.isPending}>
                    Add moderator
                </Button>
            </form>
            {error && (
                <div
                    role="alert"
                    className="mt-4 rounded border border-error-500 p-3 text-error-500"
                >
                    {error}
                </div>
            )}
            {query.isLoading ? (
                <div className="flex min-h-40 items-center justify-center">
                    <Spinner />
                </div>
            ) : query.isError ? (
                <div
                    role="alert"
                    className="mt-6 rounded border border-error-500 p-4"
                >
                    Moderators could not be loaded.{' '}
                    <button
                        className="min-h-11 underline"
                        onClick={() => void query.refetch()}
                    >
                        Retry
                    </button>
                </div>
            ) : query.data?.items.length === 0 ? (
                <p className="mt-8 text-muted-foreground">
                    No platform moderators are configured.
                </p>
            ) : (
                <ul className="mt-6 divide-y divide-border rounded-lg border border-border">
                    {query.data?.items.map((moderator) => (
                        <li
                            key={moderator.id}
                            className="flex flex-wrap items-center justify-between gap-3 p-4"
                        >
                            <div>
                                <strong>
                                    {moderator.display_name ||
                                        moderator.username}
                                </strong>
                                <div className="text-sm text-muted-foreground">
                                    {moderator.username} · {moderator.id}
                                </div>
                            </div>
                            <div className="flex gap-2">
                                <Button
                                    variant="secondary"
                                    onClick={() => update.mutate(moderator.id)}
                                >
                                    Confirm access
                                </Button>
                                <Button
                                    variant="danger"
                                    onClick={() => revoke.mutate(moderator.id)}
                                >
                                    Revoke
                                </Button>
                            </div>
                        </li>
                    ))}
                </ul>
            )}
        </Container>
    );
}
