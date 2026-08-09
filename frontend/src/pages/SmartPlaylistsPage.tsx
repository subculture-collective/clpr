import { useState } from 'react';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { Link } from 'react-router-dom';
import {
    Plus,
    Pencil,
    Trash2,
    Zap,
    ExternalLink,
    Sparkles,
} from 'lucide-react';
import {
    Container,
    Card,
    CardBody,
    Button,
    Spinner,
    Badge,
    ConfirmModal,
} from '@/components';
import { SEO } from '@/components/SEO';
import { playlistScriptApi } from '@/lib/playlist-script-api';
import { SCHEDULE_LABELS } from '@/lib/playlist-script-utils';
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

function formToRequest(
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
        strategy: 'standard',
        retention_days: form.retention_days,
        title_template: form.title_template || undefined,
    };
}

export function SmartPlaylistsPage() {
    const { showToast } = useToast();
    const queryClient = useQueryClient();
    const [showForm, setShowForm] = useState(false);
    const [editingScript, setEditingScript] = useState<PlaylistScript | null>(
        null,
    );
    const [confirmDelete, setConfirmDelete] = useState<string | null>(null);
    const [confirmGenerate, setConfirmGenerate] = useState<string | null>(null);

    const { data: scripts, isLoading } = useQuery({
        queryKey: ['my-playlist-scripts'],
        queryFn: () => playlistScriptApi.listMyScripts(),
    });

    const createMutation = useMutation({
        mutationFn: playlistScriptApi.createMyScript,
        onSuccess: () => {
            queryClient.invalidateQueries({
                queryKey: ['my-playlist-scripts'],
            });
            showToast('Smart playlist created', 'success');
            setShowForm(false);
        },
        onError: () => showToast('Failed to create smart playlist', 'error'),
    });

    const updateMutation = useMutation({
        mutationFn: ({
            id,
            data,
        }: {
            id: string;
            data: UpdatePlaylistScriptRequest;
        }) => playlistScriptApi.updateMyScript(id, data),
        onSuccess: () => {
            queryClient.invalidateQueries({
                queryKey: ['my-playlist-scripts'],
            });
            showToast('Smart playlist updated', 'success');
            setEditingScript(null);
        },
        onError: () => showToast('Failed to update smart playlist', 'error'),
    });

    const deleteMutation = useMutation({
        mutationFn: (id: string) => playlistScriptApi.deleteMyScript(id),
        onSuccess: () => {
            queryClient.invalidateQueries({
                queryKey: ['my-playlist-scripts'],
            });
            showToast('Smart playlist deleted', 'success');
        },
        onError: () => showToast('Failed to delete smart playlist', 'error'),
    });

    const generateMutation = useMutation({
        mutationFn: (id: string) => playlistScriptApi.generateMyPlaylist(id),
        onSuccess: data => {
            queryClient.invalidateQueries({
                queryKey: ['my-playlist-scripts'],
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
                data: formToRequest(values) as UpdatePlaylistScriptRequest,
            });
        } else {
            createMutation.mutate(formToRequest(values));
        }
    };

    const handleCancel = () => {
        setShowForm(false);
        setEditingScript(null);
    };

    const isFormVisible = showForm || !!editingScript;

    if (isLoading) {
        return (
            <Container className='py-8 flex justify-center'>
                <Spinner size='xl' />
            </Container>
        );
    }

    return (
        <>
            <SEO
                title='Smart Playlists'
                description='Create auto-refreshing playlists that update on a schedule'
            />
            <Container className='py-8 space-y-6'>
                <div className='flex justify-between items-start'>
                    <div>
                        <h1 className='text-3xl font-bold mb-2 flex items-center gap-2'>
                            <Sparkles className='w-7 h-7 text-primary-500' />
                            Smart Playlists
                        </h1>
                        <p className='text-muted-foreground'>
                            Create playlists that auto-refresh with the best
                            clips based on your rules.
                        </p>
                    </div>
                    {!isFormVisible && (
                        <Button onClick={() => setShowForm(true)}>
                            <Plus className='w-4 h-4 mr-2' />
                            New Smart Playlist
                        </Button>
                    )}
                </div>

                {isFormVisible && (
                    <PlaylistScriptForm
                        key={editingScript?.id ?? 'create'}
                        initialValues={
                            editingScript ?
                                scriptToFormValues(editingScript)
                            :   undefined
                        }
                        onSubmit={handleSubmit}
                        onCancel={handleCancel}
                        isEditing={!!editingScript}
                        isLoading={
                            createMutation.isPending || updateMutation.isPending
                        }
                        standardOnly
                    />
                )}

                {/* Scripts List */}
                {!scripts || scripts.length === 0 ?
                    !isFormVisible && (
                        <Card>
                            <CardBody>
                                <div className='text-center py-12'>
                                    <Sparkles className='w-12 h-12 text-muted-foreground mx-auto mb-4' />
                                    <p className='text-lg font-medium mb-2'>
                                        No smart playlists yet
                                    </p>
                                    <p className='text-muted-foreground mb-6'>
                                        Create a smart playlist that
                                        automatically fills with clips matching
                                        your criteria.
                                    </p>
                                    <Button onClick={() => setShowForm(true)}>
                                        <Plus className='w-4 h-4 mr-2' />
                                        Create Your First Smart Playlist
                                    </Button>
                                </div>
                            </CardBody>
                        </Card>
                    )
                :   <div className='grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4'>
                        {scripts.map(script => {
                            const scheduleKey = (script.schedule ||
                                'manual') as PlaylistScriptSchedule;

                            return (
                                <Card key={script.id} hover>
                                    <CardBody className='space-y-3'>
                                        <div className='flex items-start justify-between'>
                                            <div>
                                                <h3 className='font-semibold text-lg'>
                                                    {script.name}
                                                </h3>
                                                {script.description && (
                                                    <p className='text-sm text-muted-foreground line-clamp-2 mt-1'>
                                                        {script.description}
                                                    </p>
                                                )}
                                            </div>
                                            <Badge
                                                variant={
                                                    script.is_active ? 'success'
                                                    :   'default'
                                                }
                                                size='sm'
                                            >
                                                {script.is_active ?
                                                    'Active'
                                                :   'Paused'}
                                            </Badge>
                                        </div>

                                        <div className='flex flex-wrap gap-2 text-sm'>
                                            <Badge variant='info' size='sm'>
                                                {script.sort} ·{' '}
                                                {script.timeframe || 'all time'}
                                            </Badge>
                                            <Badge variant='default' size='sm'>
                                                {script.clip_limit} clips
                                            </Badge>
                                            <Badge
                                                variant='secondary'
                                                size='sm'
                                            >
                                                {SCHEDULE_LABELS[scheduleKey]}
                                            </Badge>
                                        </div>

                                        {script.last_run_at && (
                                            <p className='text-xs text-muted-foreground'>
                                                Last updated:{' '}
                                                {new Date(
                                                    script.last_run_at,
                                                ).toLocaleString()}
                                            </p>
                                        )}

                                        <div className='flex items-center gap-2 pt-2 border-t border-border'>
                                            {script.last_generated_playlist_id && (
                                                <Link
                                                    to={`/playlists/${script.last_generated_playlist_id}`}
                                                    className='inline-flex items-center gap-1 text-sm text-primary-500 hover:underline'
                                                >
                                                    View playlist{' '}
                                                    <ExternalLink className='w-3 h-3' />
                                                </Link>
                                            )}
                                            <div className='ml-auto flex gap-1'>
                                                <Button
                                                    variant='ghost'
                                                    size='sm'
                                                    onClick={() =>
                                                        setConfirmGenerate(
                                                            script.id,
                                                        )
                                                    }
                                                    title='Generate now'
                                                >
                                                    <Zap className='w-4 h-4' />
                                                </Button>
                                                <Button
                                                    variant='ghost'
                                                    size='sm'
                                                    onClick={() => {
                                                        setShowForm(false);
                                                        setEditingScript(
                                                            script,
                                                        );
                                                    }}
                                                    title='Edit'
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
                                                    title='Delete'
                                                >
                                                    <Trash2 className='w-4 h-4' />
                                                </Button>
                                            </div>
                                        </div>
                                    </CardBody>
                                </Card>
                            );
                        })}
                    </div>
                }

                <ConfirmModal
                    open={!!confirmDelete}
                    onClose={() => setConfirmDelete(null)}
                    onConfirm={() => {
                        if (confirmDelete) deleteMutation.mutate(confirmDelete);
                    }}
                    title='Delete Smart Playlist'
                    message='This will delete this smart playlist rule. Previously generated playlists will not be affected.'
                    confirmText='Delete'
                    variant='danger'
                />

                <ConfirmModal
                    open={!!confirmGenerate}
                    onClose={() => setConfirmGenerate(null)}
                    onConfirm={() => {
                        if (confirmGenerate)
                            generateMutation.mutate(confirmGenerate);
                    }}
                    title='Generate Playlist'
                    message='This will create a new playlist from this rule right now. Continue?'
                    confirmText='Generate'
                />
            </Container>
        </>
    );
}
