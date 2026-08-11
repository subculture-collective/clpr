import { useState, type FormEvent } from 'react';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { Button, Card, CardBody, CardHeader, Container, SEO, Spinner } from '../../components';
import { Input } from '../../components/ui';
import { categoryApi } from '../../lib/category-api';
import { topicApi } from '../../lib/topic-api';
import { useToast } from '../../context/ToastContext';

const selectClass = 'min-h-[44px] w-full rounded-lg border border-border bg-background px-3 py-2 text-foreground';

export function AdminTopicsPage() {
    const [clipId, setClipId] = useState('');
    const [selectedTopicIds, setSelectedTopicIds] = useState<string[]>([]);
    const [sourceTopicId, setSourceTopicId] = useState('');
    const [targetTopicId, setTargetTopicId] = useState('');
    const [splitName, setSplitName] = useState('');
    const [splitSlug, setSplitSlug] = useState('');
    const [splitClipIds, setSplitClipIds] = useState('');
    const { showToast } = useToast();
    const queryClient = useQueryClient();

    const { data, isLoading } = useQuery({
        queryKey: ['topics', 'admin'],
        queryFn: () => categoryApi.listCategories({ type: 'topic' }),
    });
    const topics = data?.categories ?? [];

    const replaceMutation = useMutation({
        mutationFn: () => topicApi.replaceClipTopics(clipId.trim(), selectedTopicIds),
        onSuccess: () => showToast('Clip topics updated', 'success'),
        onError: () => showToast('Failed to update clip topics', 'error'),
    });
    const mergeMutation = useMutation({
        mutationFn: () => topicApi.mergeTopics(sourceTopicId, targetTopicId),
        onSuccess: async () => {
            await queryClient.invalidateQueries({ queryKey: ['topics'] });
            showToast('Topics merged', 'success');
            setSourceTopicId('');
            setTargetTopicId('');
        },
        onError: () => showToast('Failed to merge topics', 'error'),
    });
    const splitMutation = useMutation({
        mutationFn: () => topicApi.splitTopic(sourceTopicId, {
            name: splitName.trim(),
            slug: splitSlug.trim(),
            clip_ids: splitClipIds.split(/[\s,]+/).map(value => value.trim()).filter(Boolean),
        }),
        onSuccess: async () => {
            await queryClient.invalidateQueries({ queryKey: ['topics'] });
            showToast('Topic split created', 'success');
            setSplitName('');
            setSplitSlug('');
            setSplitClipIds('');
        },
        onError: () => showToast('Failed to split topic', 'error'),
    });

    const submitReplace = (event: FormEvent) => {
        event.preventDefault();
        if (clipId.trim()) replaceMutation.mutate();
    };

    if (isLoading) {
        return <Container className='flex justify-center py-12'><Spinner size='xl' /></Container>;
    }

    return (
        <Container className='py-4 xs:py-6 md:py-8'>
            <SEO title='Topic Moderation' noindex />
            <div className='mb-8'>
                <h1 className='text-3xl font-bold text-text-primary'>Topic Moderation</h1>
                <p className='mt-2 text-text-secondary'>Correct clip topics and safely merge or split the topic taxonomy.</p>
            </div>

            <Card className='mb-6'>
                <CardHeader><h2 className='text-xl font-semibold'>Correct a clip</h2></CardHeader>
                <CardBody>
                    <form onSubmit={submitReplace} className='space-y-4'>
                        <Input label='Clip ID' value={clipId} onChange={event => setClipId(event.target.value)} required />
                        <fieldset>
                            <legend className='mb-2 text-sm font-medium'>Topics (up to five)</legend>
                            <div className='grid gap-2 sm:grid-cols-2 lg:grid-cols-3'>
                                {topics.map(topic => (
                                    <label key={topic.id} className='flex min-h-11 items-center gap-2 rounded border border-border px-3 py-2'>
                                        <input
                                            type='checkbox'
                                            checked={selectedTopicIds.includes(topic.id)}
                                            disabled={!selectedTopicIds.includes(topic.id) && selectedTopicIds.length >= 5}
                                            onChange={event => setSelectedTopicIds(current => event.target.checked ? [...current, topic.id] : current.filter(id => id !== topic.id))}
                                        />
                                        {topic.name}
                                    </label>
                                ))}
                            </div>
                        </fieldset>
                        <Button type='submit' disabled={!clipId.trim() || replaceMutation.isPending}>Save clip topics</Button>
                    </form>
                </CardBody>
            </Card>

            <div className='grid gap-6 lg:grid-cols-2'>
                <Card>
                    <CardHeader><h2 className='text-xl font-semibold'>Merge topics</h2></CardHeader>
                    <CardBody className='space-y-4'>
                        <label className='block text-sm font-medium'>Source topic
                            <select className={`${selectClass} mt-1`} value={sourceTopicId} onChange={event => setSourceTopicId(event.target.value)}>
                                <option value=''>Choose source</option>
                                {topics.map(topic => <option key={topic.id} value={topic.id}>{topic.name}</option>)}
                            </select>
                        </label>
                        <label className='block text-sm font-medium'>Target topic
                            <select className={`${selectClass} mt-1`} value={targetTopicId} onChange={event => setTargetTopicId(event.target.value)}>
                                <option value=''>Choose target</option>
                                {topics.filter(topic => topic.id !== sourceTopicId).map(topic => <option key={topic.id} value={topic.id}>{topic.name}</option>)}
                            </select>
                        </label>
                        <Button variant='danger' disabled={!sourceTopicId || !targetTopicId || mergeMutation.isPending} onClick={() => mergeMutation.mutate()}>Merge source into target</Button>
                    </CardBody>
                </Card>

                <Card>
                    <CardHeader><h2 className='text-xl font-semibold'>Split topic</h2></CardHeader>
                    <CardBody className='space-y-4'>
                        <p className='text-sm text-text-secondary'>The source is selected in the merge panel.</p>
                        <Input label='New topic name' value={splitName} onChange={event => setSplitName(event.target.value)} />
                        <Input label='New topic slug' value={splitSlug} onChange={event => setSplitSlug(event.target.value)} pattern='[a-z0-9]+(?:-[a-z0-9]+)*' />
                        <label className='block text-sm font-medium'>Clip IDs
                            <textarea className={`${selectClass} mt-1 min-h-28`} value={splitClipIds} onChange={event => setSplitClipIds(event.target.value)} placeholder='Comma or newline separated UUIDs' />
                        </label>
                        <Button disabled={!sourceTopicId || !splitName.trim() || !splitSlug.trim() || !splitClipIds.trim() || splitMutation.isPending} onClick={() => splitMutation.mutate()}>Create split topic</Button>
                    </CardBody>
                </Card>
            </div>
        </Container>
    );
}

export default AdminTopicsPage;
