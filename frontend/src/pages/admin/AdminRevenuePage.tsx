import { ArrowUpRight, CircleDollarSign, ExternalLink, HeartHandshake, Infinity as InfinityIcon } from 'lucide-react';
import { Link } from 'react-router-dom';
import { SEO } from '../../components';

const principles = [
    { icon: InfinityIcon, label: 'Account access', value: 'Open', detail: 'No paid tiers or feature quotas' },
    { icon: CircleDollarSign, label: 'clpr billing', value: 'Disabled', detail: 'No new Stripe checkout or plan changes' },
    { icon: HeartHandshake, label: 'Funding path', value: 'Patreon', detail: 'Optional support through Subcult' },
];

const AdminRevenuePage = () => (
    <>
        <SEO title='Community Support - Admin' noindex />
        <div className='mx-auto max-w-6xl'>
            <header className='mb-8 border-b border-border pb-7'>
                <p className='mb-2 text-xs font-bold uppercase tracking-[0.18em] text-brand'>Funding model</p>
                <h1 className='text-3xl font-bold text-text-primary'>Community Support</h1>
                <p className='mt-2 max-w-2xl text-sm leading-6 text-text-secondary'>
                    clpr account access is not monetized. People who want to sustain the project are directed to Subcult on Patreon without receiving account advantages.
                </p>
            </header>

            <section className='grid gap-4 md:grid-cols-3' aria-label='Funding status'>
                {principles.map(item => {
                    const Icon = item.icon;
                    return (
                        <article key={item.label} className='rounded-xl border border-border bg-surface p-5'>
                            <div className='mb-6 flex items-center justify-between'>
                                <span className='rounded-lg bg-brand/10 p-2.5 text-brand'><Icon className='h-5 w-5' /></span>
                                <span className='h-2 w-2 rounded-full bg-emerald-400' />
                            </div>
                            <p className='text-xs font-bold uppercase tracking-[0.14em] text-text-tertiary'>{item.label}</p>
                            <p className='mt-1 text-2xl font-bold text-text-primary'>{item.value}</p>
                            <p className='mt-2 text-sm text-text-secondary'>{item.detail}</p>
                        </article>
                    );
                })}
            </section>

            <section className='mt-8 grid gap-6 rounded-2xl border border-border bg-surface p-6 sm:p-8 lg:grid-cols-[1fr_auto] lg:items-center'>
                <div>
                    <h2 className='text-xl font-bold text-text-primary'>Public support destination</h2>
                    <p className='mt-2 max-w-2xl text-sm leading-6 text-text-secondary'>
                        The public support page explains that Patreon is optional and does not affect permissions, limits, ranking, or access.
                    </p>
                    <div className='mt-5 flex flex-wrap gap-3'>
                        <Link to='/support' className='inline-flex items-center gap-2 rounded-lg border border-border bg-surface-raised px-4 py-2.5 text-sm font-semibold text-text-primary hover:border-brand/50'>
                            View support page <ArrowUpRight className='h-4 w-4' />
                        </Link>
                        <a href='https://patreon.com/subcult' target='_blank' rel='noopener noreferrer' className='inline-flex items-center gap-2 rounded-lg bg-brand px-4 py-2.5 text-sm font-semibold text-white hover:bg-brand-hover'>
                            Open Patreon <ExternalLink className='h-4 w-4' />
                        </a>
                    </div>
                </div>
                <div className='rounded-xl border border-border bg-background px-5 py-4 text-sm text-text-secondary'>
                    <p className='font-semibold text-text-primary'>Operational note</p>
                    <p className='mt-1'>Patreon metrics are managed on Patreon.</p>
                </div>
            </section>
        </div>
    </>
);

export default AdminRevenuePage;
