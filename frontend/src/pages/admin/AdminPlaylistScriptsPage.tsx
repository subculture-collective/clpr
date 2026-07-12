import { useState } from 'react';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { Link } from 'react-router-dom';
import { Pencil, Trash2, Zap, ExternalLink } from 'lucide-react';
import {
    Container,
    Card,
    CardHeader,
    CardBody,
    Button,
    Spinner,
    Badge,
    ConfirmModal,
} from '@/components';
import { playlistScriptApi } from '@/lib/playlist-script-api';
import { STRATEGY_META, SCHEDULE_LABELS } from '@/lib/playlist-script-utils';
import {
    PlaylistScriptForm,
    type PlaylistScriptFormValues,
} from '@/components/admin/PlaylistScriptForm';
import { scriptToFormValues } from '@/lib/playlist-script-form';
import type {
    PlaylistScript,
    CreatePlaylistScriptRequest,
    UpdatePlaylistScriptRequest,
    PlaylistScriptSchedule,
} from '@/types/playlistScript';
import { useToast } from '@/context/ToastContext';

const SCHEDULE_BADGE_VARIANT: Record<
    PlaylistScriptSchedule,
    'default' | 'error' | 'info' | 'success' | 'primary'
> = {
    manual: 'default',
    hourly: 'error',
    daily: 'info',
    weekly: 'success',
    monthly: 'primary',
};

function formToCreateRequest(
    form: PlaylistScriptFormValues,
): CreatePlaylistScriptRequest {
    return {
        name: form.name,
        description: form.description || undefined,
        sort: form.sort,
        timeframe: form.timeframe,
        clip_limit: form.clip_limit,
        visibility: form.visibility,
        is_active: form.is_active,
        schedule: form.schedule,
        strategy: form.strategy,
        game_id: form.game_id || undefined,
        game_ids: form.game_ids.length > 0 ? form.game_ids : undefined,
        broadcaster_id: form.broadcaster_id || undefined,
        tag: form.tag || undefined,
        exclude_tags:
            form.exclude_tags.length > 0 ? form.exclude_tags : undefined,
        language: form.language || undefined,
        min_vote_score:
            form.min_vote_score ? Number(form.min_vote_score) : undefined,
        min_view_count:
            form.min_view_count ? Number(form.min_view_count) : undefined,
        exclude_nsfw: form.exclude_nsfw,
        top_10k_streamers: form.top_10k_streamers,
        seed_clip_id: form.seed_clip_id || undefined,
        retention_days: form.retention_days,
        title_template: form.title_template || undefined,
    };
}

function formToUpdateRequest(
    form: PlaylistScriptFormValues,
): UpdatePlaylistScriptRequest {
    return formToCreateRequest(form);
}

export function AdminPlaylistScriptsPage() {
    const { showToast } = useToast();
    const queryClient = useQueryClient();
    const [editingScript, setEditingScript] = useState<PlaylistScript | null>(
        null,
    );
    const [confirmDelete, setConfirmDelete] = useState<string | null>(null);
    const [confirmGenerate, setConfirmGenerate] = useState<string | null>(null);

    const { data: scripts, isLoading } = useQuery({
        queryKey: ['admin', 'playlist-scripts'],
        queryFn: () => playlistScriptApi.listScripts(),
    });

    const createMutation = useMutation({
        mutationFn: playlistScriptApi.createScript,
        onSuccess: () => {
            queryClient.invalidateQueries({
                queryKey: ['admin', 'playlist-scripts'],
            });
            showToast('Playlist script created', 'success');
        },
        onError: () => showToast('Failed to create script', 'error'),
    });

    const updateMutation = useMutation({
        mutationFn: ({
            id,
            data,
        }: {
            id: string;
            data: UpdatePlaylistScriptRequest;
        }) => playlistScriptApi.updateScript(id, data),
        onSuccess: () => {
            queryClient.invalidateQueries({
                queryKey: ['admin', 'playlist-scripts'],
            });
            showToast('Playlist script updated', 'success');
            setEditingScript(null);
        },
        onError: () => showToast('Failed to update script', 'error'),
    });

    const deleteMutation = useMutation({
        mutationFn: (id: string) => playlistScriptApi.deleteScript(id),
        onSuccess: () => {
            queryClient.invalidateQueries({
                queryKey: ['admin', 'playlist-scripts'],
            });
            showToast('Playlist script deleted', 'success');
        },
        onError: () => showToast('Failed to delete script', 'error'),
    });

    const generateMutation = useMutation({
        mutationFn: (id: string) => playlistScriptApi.generatePlaylist(id),
        onSuccess: data => {
            queryClient.invalidateQueries({
                queryKey: ['admin', 'playlist-scripts'],
            });
            showToast('Playlist generated', 'success');
            if (data?.id) {
                window.open(`/playlists/${data.id}`, '_blank');
            }
        },
        onError: () => showToast('Failed to generate playlist', 'error'),
    });

    const handleSubmit = (values: PlaylistScriptFormValues) => {
        if (editingScript) {
            updateMutation.mutate({
                id: editingScript.id,
                data: formToUpdateRequest(values),
            });
        } else {
            createMutation.mutate(formToCreateRequest(values));
        }
    };

    if (isLoading) {
        return (
            <Container className='py-8 flex justify-center'>
                <Spinner size='xl' />
            </Container>
        );
    }

    return (
        <Container className='py-8 space-y-6'>
            <div>
                <h1 className='text-3xl font-bold mb-2'>Playlist Scripts</h1>
                <p className='text-muted-foreground'>
                    Automate playlists like daily top 10s, trending picks, and
                    curated collections.
                </p>
            </div>

            {/* Form */}
            <PlaylistScriptForm
                key={editingScript?.id ?? 'create'}
                initialValues={
                    editingScript ?
                        scriptToFormValues(editingScript)
                    :   undefined
                }
                onSubmit={handleSubmit}
                onCancel={
                    editingScript ? () => setEditingScript(null) : undefined
                }
                isEditing={!!editingScript}
                isLoading={createMutation.isPending || updateMutation.isPending}
            />

            {/* Scripts Table */}
            <Card>
                <CardHeader>
                    <h2 className='text-xl font-semibold'>Existing Scripts</h2>
                </CardHeader>
                <CardBody>
                    {!scripts || scripts.length === 0 ?
                        <div className='text-center py-12 text-muted-foreground'>
                            <p className='text-lg'>No scripts yet</p>
                            <p className='text-sm mt-2'>
                                Create your first script above.
                            </p>
                        </div>
                    :   <div className='overflow-x-auto'>
                            <table className='w-full'>
                                <thead className='border-b border-border'>
                                    <tr>
                                        <th className='text-left py-3 px-4 font-semibold'>
                                            Name
                                        </th>
                                        <th className='text-left py-3 px-4 font-semibold'>
                                            Strategy
                                        </th>
                                        <th className='text-left py-3 px-4 font-semibold'>
                                            Schedule
                                        </th>
                                        <th className='text-left py-3 px-4 font-semibold'>
                                            Rule
                                        </th>
                                        <th className='text-left py-3 px-4 font-semibold'>
                                            Status
                                        </th>
                                        <th className='text-left py-3 px-4 font-semibold'>
                                            Last Run
                                        </th>
                                        <th className='text-right py-3 px-4 font-semibold'>
                                            Actions
                                        </th>
                                    </tr>
                                </thead>
                                <tbody>
                                    {scripts.map(script => {
                                        const strategyMeta =
                                            STRATEGY_META[script.strategy] ??
                                            STRATEGY_META.standard;
                                        const scheduleKey = (script.schedule ||
                                            'manual') as PlaylistScriptSchedule;

                                        return (
                                            <tr
                                                key={script.id}
                                                className='border-b border-border hover:bg-accent transition-colors'
                                            >
                                                <td className='py-3 px-4'>
                                                    <div className='font-medium'>
                                                        {script.name}
                                                    </div>
                                                    {script.description && (
                                                        <div className='text-sm text-muted-foreground line-clamp-1'>
                                                            {script.description}
                                                        </div>
                                                    )}
                                                </td>
                                                <td className='py-3 px-4'>
                                                    <Badge
                                                        variant={
                                                            (
                                                                strategyMeta.group ===
                                                                'twitch'
                                                            ) ?
                                                                'primary'
                                                            :   'secondary'
                                                        }
                                                        size='sm'
                                                    >
                                                        {strategyMeta.label}
                                                    </Badge>
                                                </td>
                                                <td className='py-3 px-4'>
                                                    <Badge
                                                        variant={
                                                            SCHEDULE_BADGE_VARIANT[
                                                                scheduleKey
                                                            ] ?? 'default'
                                                        }
                                                        size='sm'
                                                    >
                                                        {SCHEDULE_LABELS[
                                                            scheduleKey
                                                        ] ?? scheduleKey}
                                                    </Badge>
                                                </td>
                                                <td className='py-3 px-4 text-sm'>
                                                    {script.sort} ·{' '}
                                                    {script.timeframe ||
                                                        'all time'}{' '}
                                                    · {script.clip_limit} clips
                                                </td>
                                                <td className='py-3 px-4'>
                                                    <Badge
                                                        variant={
                                                            script.is_active ?
                                                                'success'
                                                            :   'default'
                                                        }
                                                        size='sm'
                                                    >
                                                        {script.is_active ?
                                                            'Active'
                                                        :   'Inactive'}
                                                    </Badge>
                                                </td>
                                                <td className='py-3 px-4 text-sm text-muted-foreground'>
                                                    <div>
                                                        {script.last_run_at ?
                                                            new Date(
                                                                script.last_run_at,
                                                            ).toLocaleString()
                                                        :   'Never'}
                                                    </div>
                                                    {script.last_generated_playlist_id && (
                                                        <Link
                                                            to={`/playlists/${script.last_generated_playlist_id}`}
                                                            className='inline-flex items-center gap-1 text-xs text-primary-500 hover:underline mt-0.5'
                                                        >
                                                            View playlist{' '}
                                                            <ExternalLink className='w-3 h-3' />
                                                        </Link>
                                                    )}
                                                </td>
                                                <td className='py-3 px-4'>
                                                    <div className='flex justify-end gap-2'>
                                                        <Button
                                                            variant='ghost'
                                                            size='sm'
                                                            onClick={() =>
                                                                setConfirmGenerate(
                                                                    script.id,
                                                                )
                                                            }
                                                            title='Generate playlist'
                                                        >
                                                            <Zap className='w-4 h-4' />
                                                        </Button>
                                                        <Button
                                                            variant='ghost'
                                                            size='sm'
                                                            onClick={() =>
                                                                setEditingScript(
                                                                    script,
                                                                )
                                                            }
                                                            title='Edit script'
                                                        >
                                                            <Pencil className='w-4 h-4' />
                                                        </Button>
                                                        <Button
                                                            variant='ghost'
                                                            size='sm'
                                                            onClick={() =>
                                                                setConfirmDelete(
                                                                    script.id,
                                                                )
                                                            }
                                                            title='Delete script'
                                                        >
                                                            <Trash2 className='w-4 h-4' />
                                                        </Button>
                                                    </div>
                                                </td>
                                            </tr>
                                        );
                                    })}
                                </tbody>
                            </table>
                        </div>
                    }
                </CardBody>
            </Card>

            {/* Confirm Delete Modal */}
            <ConfirmModal
                open={!!confirmDelete}
                onClose={() => setConfirmDelete(null)}
                onConfirm={() => {
                    if (confirmDelete) deleteMutation.mutate(confirmDelete);
                }}
                title='Delete Script'
                message='This will permanently delete this playlist script. Generated playlists will not be affected.'
                confirmText='Delete'
                variant='danger'
            />

            {/* Confirm Generate Modal */}
            <ConfirmModal
                open={!!confirmGenerate}
                onClose={() => setConfirmGenerate(null)}
                onConfirm={() => {
                    if (confirmGenerate)
                        generateMutation.mutate(confirmGenerate);
                }}
                title='Generate Playlist'
                message='This will create a new playlist from this script right now. Continue?'
                confirmText='Generate'
            />
        </Container>
    );
}
