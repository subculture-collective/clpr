import { useMemo, useState, useRef } from 'react';
import { useMutation, useQuery } from '@tanstack/react-query';
import { Check, ChevronLeft, ChevronRight, Sparkles } from 'lucide-react';
import { useNavigate } from 'react-router-dom';
import { SEO } from '../components/SEO';
import { Button } from '../components/ui/Button';
import { categoryApi } from '../lib/category-api';
import {
    fetchCreatorDiscovery,
    followBroadcaster,
} from '../lib/broadcaster-api';
import {
    completeCreatorFirstOnboarding,
    type CreatorFirstOnboardingRequest,
} from '../lib/recommendation-api';
import { tagApi } from '../lib/tag-api';

const steps = [
    { eyebrow: 'Step 1 of 3', title: 'Follow creators', detail: 'Start with people you already want more from.' },
    { eyebrow: 'Step 2 of 3', title: 'Choose topics', detail: 'Tell us which corners of live culture pull you in.' },
    { eyebrow: 'Step 3 of 3', title: 'Pick your moments', detail: 'Choose the energy you want clips to bring.' },
];

export function OnboardingPage() {
    const navigate = useNavigate();
    const headingRef = useRef<HTMLHeadingElement>(null);
    const [step, setStep] = useState(0);
    const [creators, setCreators] = useState<string[]>([]);
    const [topics, setTopics] = useState<string[]>([]);
    const [tags, setTags] = useState<string[]>([]);

    const creatorQuery = useQuery({
        queryKey: ['onboarding-creators'],
        queryFn: () => fetchCreatorDiscovery(12),
    });
    const topicQuery = useQuery({
        queryKey: ['onboarding-topics'],
        queryFn: () => categoryApi.listCategories({ featured: true }),
    });
    const tagQuery = useQuery({
        queryKey: ['onboarding-tags'],
        queryFn: () => tagApi.listTags({ sort: 'popularity', limit: 24 }),
    });

    const creatorOptions = useMemo(() => {
        const rails = creatorQuery.data;
        if (!rails) return [];
        const unique = new Map<string, (typeof rails.trending)[number]>();
        [...rails.trending, ...rails.rising, ...rails.new].forEach(creator =>
            unique.set(creator.broadcaster_id, creator),
        );
        return [...unique.values()].slice(0, 18);
    }, [creatorQuery.data]);

    const mutation = useMutation({
        mutationFn: async (request: CreatorFirstOnboardingRequest) => {
            await Promise.all(
                request.followed_creators.map(creatorID =>
                    followBroadcaster(creatorID),
                ),
            );
            return completeCreatorFirstOnboarding(request);
        },
        onSuccess: () => navigate('/', { replace: true }),
    });

    const toggle = (value: string, values: string[], setter: (next: string[]) => void) => {
        setter(values.includes(value) ? values.filter(item => item !== value) : [...values, value]);
    };
    const selectionCount = creators.length + topics.length + tags.length;
    const current = steps[step];
    const currentQuery = [creatorQuery, topicQuery, tagQuery][step];
    const optionCount = [creatorOptions.length, topicQuery.data?.categories?.length ?? 0, tagQuery.data?.tags?.length ?? 0][step];
    const changeStep = (next: number) => {
        setStep(next);
        requestAnimationFrame(() => headingRef.current?.focus());
    };

    return (
        <section
            aria-labelledby='onboarding-heading'
            className='min-h-[calc(100vh-4rem)] bg-background px-4 py-10 sm:py-16'
        >
            <SEO title='Shape your feed' description='Follow creators and choose the topics and moments you want on Clpr.' noindex />
            <div className='mx-auto max-w-5xl'>
                <div className='mb-10 flex items-center justify-between gap-6'>
                    <div>
                        <p className='mb-2 text-xs font-semibold uppercase tracking-[0.22em] text-link'>{current.eyebrow}</p>
                        <h1 ref={headingRef} tabIndex={-1} id='onboarding-heading' className='text-4xl font-black tracking-tight sm:text-6xl'>{current.title}</h1>
                        <p className='mt-3 max-w-2xl text-base text-muted-foreground sm:text-lg'>{current.detail}</p>
                    </div>
                    <div className='hidden rounded-full border border-primary-500/30 bg-primary-500/10 p-4 text-link sm:block'>
                        <Sparkles aria-hidden='true' />
                    </div>
                </div>

                <div className='mb-8 grid grid-cols-3 gap-2' role='progressbar' aria-label='Onboarding progress' aria-valuemin={1} aria-valuemax={3} aria-valuenow={step + 1} aria-valuetext={current.eyebrow}>
                    {steps.map((item, index) => (
                        <div key={item.title} className={`h-1.5 rounded-full ${index <= step ? 'bg-primary-500' : 'bg-muted'}`} />
                    ))}
                </div>

                {step === 0 && (
                    <section className='grid grid-cols-2 gap-3 sm:grid-cols-3 lg:grid-cols-6'>
                        {creatorOptions.map(creator => {
                            const selected = creators.includes(creator.broadcaster_id);
                            return (
                                <button key={creator.broadcaster_id} type='button' aria-pressed={selected} disabled={mutation.isPending} onClick={() => toggle(creator.broadcaster_id, creators, setCreators)}
                                    className={`group relative overflow-hidden rounded-2xl border p-3 text-left transition ${selected ? 'border-primary-500 bg-primary-500/10' : 'border-border bg-card hover:border-primary-500/50'}`}>
                                    <img src={creator.latest_clip_thumbnail || '/og-image.svg'} alt='' className='mb-3 aspect-square w-full rounded-xl object-cover' />
                                    <span className='block truncate font-bold'>{creator.broadcaster_name}</span>
                                    <span className='mt-1 block truncate text-xs text-muted-foreground'>{creator.twitch_category_name || 'Live culture'}</span>
                                    {selected && <Check className='absolute right-5 top-5 rounded-full bg-primary-400 p-1 text-background' size={24} />}
                                </button>
                            );
                        })}
                    </section>
                )}

                {step === 1 && (
                    <section className='grid gap-3 sm:grid-cols-2 lg:grid-cols-3'>
                        {(topicQuery.data?.categories || []).map(topic => {
                            const selected = topics.includes(topic.slug);
                            return <button key={topic.id} type='button' aria-pressed={selected} disabled={mutation.isPending} onClick={() => toggle(topic.slug, topics, setTopics)}
                                className={`rounded-2xl border p-5 text-left transition ${selected ? 'border-primary-500 bg-primary-500/10' : 'border-border bg-card hover:border-primary-500/50'}`}>
                                <span className='flex items-center justify-between text-lg font-bold'>{topic.name}{selected && <Check size={20} />}</span>
                                {topic.description && <span className='mt-2 block text-sm text-muted-foreground'>{topic.description}</span>}
                            </button>;
                        })}
                    </section>
                )}

                {step === 2 && (
                    <section className='flex flex-wrap gap-3'>
                        {(tagQuery.data?.tags || []).map(tag => {
                            const selected = tags.includes(tag.id);
                            return <button key={tag.id} type='button' aria-pressed={selected} disabled={mutation.isPending} onClick={() => toggle(tag.id, tags, setTags)}
                                className={`rounded-full border px-5 py-3 font-semibold transition ${selected ? 'border-primary-500 bg-primary-400 text-background' : 'border-border bg-card hover:border-primary-500/50'}`}>
                                {tag.name}{selected && <Check className='ml-2 inline' size={16} />}
                            </button>;
                        })}
                    </section>
                )}

                {!currentQuery.isLoading && !currentQuery.isError && optionCount === 0 && <p role='status' className='py-6 text-muted-foreground'>No choices are available for this step. You can continue or skip for now.</p>}
                {currentQuery.isLoading && <p role='status' className='py-6 text-muted-foreground'>Loading choices…</p>}
                {currentQuery.isError && (
                    <div role='alert' className='mt-6 rounded-lg border border-border p-4'>
                        <p>These choices could not be loaded. You can retry or skip this step.</p>
                        <Button variant='outline' className='mt-3' disabled={currentQuery.isFetching} onClick={() => currentQuery.refetch()}>Try again</Button>
                    </div>
                )}
                {mutation.isError && <p role='alert' className='mt-6 text-sm text-error-400'>We could not save your feed yet. Your selections are still here. Please try again.</p>}
                <p className='mt-6 text-sm text-muted-foreground'>All choices are optional. You can change your interests later.</p>

                <footer className='mt-6 flex flex-wrap items-center justify-between gap-3 border-t border-border pt-6'>
                    <Button variant='ghost' disabled={step === 0 || mutation.isPending} onClick={() => changeStep(step - 1)}>
                        <ChevronLeft size={18} /> Back
                    </Button>
                    <span className='text-sm text-muted-foreground'>{selectionCount} selected</span>
                    {step < 2 ? (
                        <Button onClick={() => changeStep(step + 1)}>Continue <ChevronRight size={18} /></Button>
                    ) : (
                        <Button disabled={mutation.isPending} onClick={() => selectionCount === 0 ? navigate('/', { replace: true }) : mutation.mutate({ followed_creators: creators, preferred_topics: topics, preferred_tags: tags })}>
                            {mutation.isPending ? 'Shaping your feed…' : selectionCount === 0 ? 'Browse clips' : 'Build my feed'}
                        </Button>
                    )}
                </footer>
                <Button variant='ghost' className='mt-3' disabled={mutation.isPending} onClick={() => navigate('/', { replace: true })}>Skip for now</Button>
            </div>
        </section>
    );
}
