import { ArrowUpRight, Heart, Infinity as InfinityIcon, ShieldCheck, Sparkles } from 'lucide-react';
import { SEO } from '../components';

const accessPromises = [
    {
        icon: InfinityIcon,
        title: 'No account tiers',
        description: 'Every signed-in account gets the same product features. There is no paid upgrade path.',
    },
    {
        icon: Sparkles,
        title: 'No feature paywalls',
        description: 'Discovery, favorites, collections, and creator tools are part of clpr—not add-ons.',
    },
    {
        icon: ShieldCheck,
        title: 'Community safeguards only',
        description: 'Operational limits exist only for security, spam prevention, and service reliability.',
    },
];

export default function SupportPage() {
    return (
        <>
            <SEO
                title='Support clpr'
                description='clpr is free to use. If it matters to you, help sustain the project through Subcult on Patreon.'
                canonicalUrl='/support'
            />
            <div className='relative overflow-hidden bg-background px-4 py-14 sm:px-6 sm:py-20 lg:px-8'>
                <div aria-hidden='true' className='absolute left-1/2 top-0 h-80 w-[52rem] -translate-x-1/2 rounded-full bg-brand/10 blur-3xl' />
                <div className='relative mx-auto max-w-6xl'>
                    <section className='grid items-end gap-10 border-b border-border pb-12 lg:grid-cols-[minmax(0,1.35fr)_minmax(18rem,0.65fr)]'>
                        <div>
                            <p className='mb-4 text-xs font-bold uppercase tracking-[0.2em] text-violet-300'>Community supported</p>
                            <h1 className='max-w-4xl font-accent text-4xl font-extrabold leading-[0.98] tracking-tight text-text-primary sm:text-6xl'>
                                clpr is for the culture,<br />not a customer tier.
                            </h1>
                            <p className='mt-6 max-w-2xl text-base leading-7 text-text-secondary sm:text-lg'>
                                Accounts are free and features are not held behind a subscription. If clpr is useful to you, Patreon is an optional way to help Subcult keep it running and growing.
                            </p>
                        </div>
                        <a
                            href='https://patreon.com/subcult'
                            target='_blank'
                            rel='noopener noreferrer'
                            className='group flex items-center justify-between rounded-2xl border border-brand/40 bg-brand p-6 text-white shadow-[0_20px_60px_-28px_rgba(145,70,255,0.85)] transition-transform hover:-translate-y-1'
                        >
                            <span>
                                <span className='mb-2 flex items-center gap-2 text-xs font-bold uppercase tracking-[0.17em] text-white'><Heart className='h-4 w-4 fill-current' /> Keep it moving</span>
                                <span className='block text-xl font-bold'>Support Subcult on Patreon</span>
                                <span className='mt-1 block text-sm text-white'>Optional support. No features attached.</span>
                            </span>
                            <ArrowUpRight className='h-6 w-6 shrink-0 transition-transform group-hover:translate-x-1 group-hover:-translate-y-1' />
                        </a>
                    </section>

                    <section className='grid gap-px overflow-hidden rounded-2xl border border-border bg-border mt-10 md:grid-cols-3'>
                        {accessPromises.map(item => {
                            const Icon = item.icon;
                            return (
                                <article key={item.title} className='bg-surface p-6 sm:p-8'>
                                    <span className='mb-8 inline-flex rounded-xl bg-brand/10 p-3 text-brand'><Icon className='h-5 w-5' /></span>
                                    <h2 className='text-lg font-bold text-text-primary'>{item.title}</h2>
                                    <p className='mt-2 text-sm leading-6 text-text-secondary'>{item.description}</p>
                                </article>
                            );
                        })}
                    </section>

                    <p className='mx-auto mt-10 max-w-2xl text-center text-sm leading-6 text-text-secondary'>
                        Patreon support is handled by Patreon under its own terms. Supporting does not change your clpr account, permissions, limits, or ranking.
                    </p>
                </div>
            </div>
        </>
    );
}
