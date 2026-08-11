import { useMemo, useState } from 'react';
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

    return (
        <main className='min-h-[calc(100vh-4rem)] bg-background px-4 py-10 sm:py-16'>
            <SEO title='Shape your feed' description='Follow creators and choose the topics and moments you want on Clpr.' noindex />
            <div className='mx-auto max-w-5xl'>
                <div className='mb-10 flex items-center justify-between gap-6'>
                    <div>
                        <p className='mb-2 text-xs font-semibold uppercase tracking-[0.22em] text-primary-500'>{current.eyebrow}</p>
                        <h1 className='text-4xl font-black tracking-tight sm:text-6xl'>{current.title}</h1>
                        <p className='mt-3 max-w-2xl text-base text-muted-foreground sm:text-lg'>{current.detail}</p>
                    </div>
                    <div className='hidden rounded-full border border-primary-500/30 bg-primary-500/10 p-4 text-primary-500 sm:block'>
                        <Sparkles aria-hidden='true' />
                    </div>
                </div>

                <div className='mb-8 grid grid-cols-3 gap-2' aria-label='Onboarding progress'>
                    {steps.map((item, index) => (
                        <div key={item.title} className={`h-1.5 rounded-full ${index <= step ? 'bg-primary-500' : 'bg-muted'}`} />
                    ))}
                </div>

                {step === 0 && (
                    <section className='grid grid-cols-2 gap-3 sm:grid-cols-3 lg:grid-cols-6'>
                        {creatorOptions.map(creator => {
                            const selected = creators.includes(creator.broadcaster_id);
                            return (
                                <button key={creator.broadcaster_id} type='button' onClick={() => toggle(creator.broadcaster_id, creators, setCreators)}
                                    className={`group relative overflow-hidden rounded-2xl border p-3 text-left transition ${selected ? 'border-primary-500 bg-primary-500/10' : 'border-border bg-card hover:border-primary-500/50'}`}>
                                    <img src={creator.latest_clip_thumbnail || '/og-image.svg'} alt='' className='mb-3 aspect-square w-full rounded-xl object-cover' />
                                    <span className='block truncate font-bold'>{creator.broadcaster_name}</span>
                                    <span className='mt-1 block truncate text-xs text-muted-foreground'>{creator.twitch_category_name || 'Live culture'}</span>
                                    {selected && <Check className='absolute right-5 top-5 rounded-full bg-primary-500 p-1 text-white' size={24} />}
                                </button>
                            );
                        })}
                    </section>
                )}

                {step === 1 && (
                    <section className='grid gap-3 sm:grid-cols-2 lg:grid-cols-3'>
                        {(topicQuery.data?.categories || []).map(topic => {
                            const selected = topics.includes(topic.slug);
                            return <button key={topic.id} type='button' onClick={() => toggle(topic.slug, topics, setTopics)}
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
                            return <button key={tag.id} type='button' onClick={() => toggle(tag.id, tags, setTags)}
                                className={`rounded-full border px-5 py-3 font-semibold transition ${selected ? 'border-primary-500 bg-primary-500 text-white' : 'border-border bg-card hover:border-primary-500/50'}`}>
                                {tag.name}{selected && <Check className='ml-2 inline' size={16} />}
                            </button>;
                        })}
                    </section>
                )}

                {(creatorQuery.isError || topicQuery.isError || tagQuery.isError) && (
                    <p className='mt-6 rounded-xl border border-destructive/30 bg-destructive/10 p-4 text-sm text-destructive'>Some choices could not be loaded. You can continue with the selections that are available.</p>
                )}
                {mutation.isError && <p className='mt-6 text-sm text-destructive'>We could not save your feed yet. Please try again.</p>}

                <footer className='mt-10 flex items-center justify-between border-t border-border pt-6'>
                    <Button variant='ghost' disabled={step === 0 || mutation.isPending} onClick={() => setStep(value => value - 1)}>
                        <ChevronLeft size={18} /> Back
                    </Button>
                    <span className='text-sm text-muted-foreground'>{selectionCount} selected</span>
                    {step < 2 ? (
                        <Button onClick={() => setStep(value => value + 1)}>Continue <ChevronRight size={18} /></Button>
                    ) : (
                        <Button disabled={selectionCount === 0 || mutation.isPending} onClick={() => mutation.mutate({ followed_creators: creators, preferred_topics: topics, preferred_tags: tags })}>
                            {mutation.isPending ? 'Shaping your feed…' : 'Build my feed'}
                        </Button>
                    )}
                </footer>
            </div>
        </main>
    );
}
