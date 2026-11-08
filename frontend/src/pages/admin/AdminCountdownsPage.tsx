import { useState, type FormEvent } from 'react';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { Alert, Button, Card, CardBody, CardHeader, Container, SEO, Spinner } from '../../components';
import { Input } from '../../components/ui';
import {
    countdownControlsApi,
    type CountdownExclusion,
    type CountdownRemovalTask,
} from '../../lib/countdown-controls-api';
import { getErrorMessage } from '../../lib/error-utils';

const selectClass = 'min-h-[44px] w-full rounded-lg border border-border bg-background px-3 py-2 text-foreground';
const reasons: CountdownExclusion['reason'][] = ['request', 'dmca', 'takedown'];

function formatDay(value: string) {
    return new Date(value).toLocaleDateString('en-US', { year: 'numeric', month: 'short', day: 'numeric' });
}

function RemovalRow({ task, onResolve, busy, now }: {
    task: CountdownRemovalTask;
    now: number;
    onResolve: (resolution: 'removed' | 'edited' | 'kept', note: string) => void;
    busy: boolean;
}) {
    const [resolution, setResolution] = useState<'removed' | 'edited' | 'kept'>('edited');
    const [note, setNote] = useState('');
    const overdue = new Date(task.due_at).getTime() < now;
    return (
        <li className='rounded-lg border border-border p-4' data-testid='removal-task'>
            <div className='flex flex-wrap items-baseline justify-between gap-2'>
                <p className='font-semibold text-foreground'>
                    {task.clip_title} <span className='font-normal text-muted-foreground'>· {task.broadcaster_name}</span>
                </p>
                <span className={overdue ? 'text-sm font-semibold text-error-500' : 'text-sm text-muted-foreground'}>
                    {overdue ? 'Overdue since ' : 'Due '}{formatDay(task.due_at)}
                </span>
            </div>
            <p className='mt-1 text-sm text-muted-foreground'>
                In the {task.kind === 'short' ? 'Short' : 'countdown'} of {task.series_slug}, {task.edition_date}:{' '}
                <a
                    className='text-link underline'
                    href={`https://studio.youtube.com/video/${encodeURIComponent(task.youtube_video_id)}/edit`}
                    target='_blank'
                    rel='noopener noreferrer'
                >
                    open in YouTube Studio
                </a>
                . Excluded ({task.exclusion_reason}, {task.exclusion_source === 'self_service' ? 'by the broadcaster' : 'by an admin'}) on{' '}
                {formatDay(task.excluded_at)}.
            </p>
            <div className='mt-3 grid gap-2 sm:grid-cols-[10rem_1fr_auto]'>
                <select
                    aria-label='Resolution'
                    className={selectClass}
                    value={resolution}
                    onChange={event => setResolution(event.target.value as typeof resolution)}
                >
                    <option value='edited'>Clip cut from the video</option>
                    <option value='removed'>Video removed</option>
                    <option value='kept'>Kept (note why)</option>
                </select>
                <Input aria-label='Note' placeholder='Note (optional)' value={note} maxLength={500} onChange={event => setNote(event.target.value)} />
                <Button
                    disabled={busy || (resolution === 'kept' && !note.trim())}
                    onClick={() => onResolve(resolution, note.trim())}
                >
                    Mark done
                </Button>
            </div>
        </li>
    );
}

/** Countdown exclusions and the takedown task list (#75). */
export function AdminCountdownsPage() {
    const queryClient = useQueryClient();
    const [now] = useState(() => Date.now());
    const [target, setTarget] = useState<'broadcaster' | 'clip'>('broadcaster');
    const [targetId, setTargetId] = useState('');
    const [reason, setReason] = useState<CountdownExclusion['reason']>('request');
    const [note, setNote] = useState('');

    const removals = useQuery({ queryKey: ['countdown-removals'], queryFn: countdownControlsApi.listRemovals });
    const exclusions = useQuery({ queryKey: ['countdown-exclusions'], queryFn: countdownControlsApi.listExclusions });
    const refresh = () =>
        Promise.all([
            queryClient.invalidateQueries({ queryKey: ['countdown-removals'] }),
            queryClient.invalidateQueries({ queryKey: ['countdown-exclusions'] }),
        ]);

    const create = useMutation({
        mutationFn: () =>
            countdownControlsApi.createExclusion({
                ...(target === 'broadcaster' ? { broadcaster_id: targetId.trim() } : { clip_id: targetId.trim() }),
                reason,
                ...(note.trim() ? { note: note.trim() } : {}),
            }),
        onSuccess: async () => {
            setTargetId('');
            setNote('');
            await refresh();
        },
    });
    const remove = useMutation({ mutationFn: countdownControlsApi.deleteExclusion, onSuccess: refresh });
    const resolve = useMutation({ mutationFn: countdownControlsApi.resolveRemoval, onSuccess: refresh });

    const submit = (event: FormEvent) => {
        event.preventDefault();
        if (targetId.trim()) create.mutate();
    };

    return (
        <>
            <SEO title='Countdowns' noindex />
            <Container className='py-8'>
                <h1 className='text-3xl font-bold text-foreground mb-2'>Countdowns</h1>
                <p className='text-muted-foreground mb-8'>
                    Broadcasters and clips left out of countdowns, and the uploaded videos that still show them.
                </p>

                <Card className='mb-6'>
                    <CardHeader>
                        <h2 className='text-xl font-semibold'>Removal tasks</h2>
                    </CardHeader>
                    <CardBody>
                        {removals.isLoading ? (
                            <Spinner />
                        ) : removals.isError ? (
                            <Alert variant='error'>Removal tasks could not be loaded.</Alert>
                        ) : (removals.data ?? []).length === 0 ? (
                            <p className='text-muted-foreground'>No uploaded video shows an excluded clip.</p>
                        ) : (
                            <ul className='space-y-3'>
                                {removals.data!.map(task => (
                                    <RemovalRow
                                        key={`${task.render_id}-${task.clip_id}`}
                                        task={task}
                                        now={now}
                                        busy={resolve.isPending}
                                        onResolve={(resolution, taskNote) =>
                                            resolve.mutate({
                                                render_id: task.render_id,
                                                clip_id: task.clip_id,
                                                resolution,
                                                ...(taskNote ? { note: taskNote } : {}),
                                            })
                                        }
                                    />
                                ))}
                            </ul>
                        )}
                        {resolve.error && <Alert variant='error' className='mt-3'>{getErrorMessage(resolve.error, 'The task could not be resolved.')}</Alert>}
                    </CardBody>
                </Card>

                <Card>
                    <CardHeader>
                        <h2 className='text-xl font-semibold'>Exclusions</h2>
                    </CardHeader>
                    <CardBody>
                        <form onSubmit={submit} className='mb-6 grid gap-3 md:grid-cols-[10rem_1fr_9rem]' aria-label='Add an exclusion'>
                            <select aria-label='Exclude' className={selectClass} value={target} onChange={event => setTarget(event.target.value as typeof target)}>
                                <option value='broadcaster'>Broadcaster</option>
                                <option value='clip'>One clip</option>
                            </select>
                            <Input
                                aria-label={target === 'broadcaster' ? 'Twitch broadcaster ID' : 'clpr clip ID'}
                                placeholder={target === 'broadcaster' ? 'Twitch broadcaster ID, e.g. 121059319' : 'clpr clip ID'}
                                value={targetId}
                                onChange={event => setTargetId(event.target.value)}
                            />
                            <select aria-label='Reason' className={selectClass} value={reason} onChange={event => setReason(event.target.value as typeof reason)}>
                                {reasons.map(r => <option key={r} value={r}>{r === 'dmca' ? 'DMCA' : r[0].toUpperCase() + r.slice(1)}</option>)}
                            </select>
                            <Input aria-label='Exclusion note' className='md:col-span-2' placeholder='Note, e.g. asked by email on …' value={note} maxLength={500} onChange={event => setNote(event.target.value)} />
                            <Button type='submit' disabled={!targetId.trim() || create.isPending}>Exclude</Button>
                        </form>
                        {create.error && <Alert variant='error' className='mb-4'>{getErrorMessage(create.error, 'The exclusion could not be added.')}</Alert>}
                        {exclusions.isLoading ? (
                            <Spinner />
                        ) : (exclusions.data ?? []).length === 0 ? (
                            <p className='text-muted-foreground'>Nothing is excluded.</p>
                        ) : (
                            <ul className='divide-y divide-border'>
                                {exclusions.data!.map(x => (
                                    <li key={x.id} className='flex flex-wrap items-center justify-between gap-3 py-3' data-testid='exclusion'>
                                        <div>
                                            <p className='font-medium text-foreground'>
                                                {x.broadcaster_id ? `Broadcaster ${x.broadcaster_id}` : `Clip ${x.clip_id}`}
                                            </p>
                                            <p className='text-sm text-muted-foreground'>
                                                {x.reason}, {x.source === 'self_service' ? 'by the broadcaster' : 'by an admin'}, {formatDay(x.created_at)}
                                                {x.note ? ` · ${x.note}` : ''}
                                            </p>
                                        </div>
                                        <Button variant='outline' size='sm' disabled={remove.isPending} onClick={() => remove.mutate(x.id)}>
                                            Lift
                                        </Button>
                                    </li>
                                ))}
                            </ul>
                        )}
                    </CardBody>
                </Card>
            </Container>
        </>
    );
}
