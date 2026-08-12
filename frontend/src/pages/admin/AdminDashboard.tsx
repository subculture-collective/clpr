import { Link } from 'react-router-dom';
import { ArrowUpRight, CheckCircle2, Clock3, ShieldCheck, Sparkles } from 'lucide-react';
import { adminNavGroups } from '../../components/admin/adminNavigation';
import { SEO } from '../../components';

const priorityActions = [
    { label: 'Review moderation queue', href: '/admin/moderation', description: 'Work through the highest-priority reports and flagged content.', icon: ShieldCheck },
    { label: 'Review creator verification', href: '/admin/verification', description: 'Approve or reject pending creator applications.', icon: CheckCircle2 },
    { label: 'Check collection automation', href: '/admin/playlist-scripts', description: 'Inspect generated playlists and scheduling rules.', icon: Sparkles },
];

export function AdminDashboard() {
    const today = new Intl.DateTimeFormat('en-US', {
        weekday: 'long', month: 'long', day: 'numeric',
    }).format(new Date());

    return (
        <>
            <SEO title='Admin Control Room' noindex />
            <div className='mx-auto max-w-7xl'>
                <header className='relative mb-8 overflow-hidden rounded-2xl border border-border bg-surface px-6 py-7 sm:px-8 sm:py-9'>
                    <div aria-hidden='true' className='absolute -right-20 -top-28 h-72 w-72 rounded-full bg-brand/10 blur-3xl' />
                    <div className='relative max-w-3xl'>
                        <div className='mb-4 flex items-center gap-2 text-xs font-bold uppercase tracking-[0.18em] text-brand'>
                            <span className='h-1.5 w-1.5 rounded-full bg-emerald-400' />
                            {today}
                        </div>
                        <h1 className='font-accent text-3xl font-extrabold tracking-tight text-text-primary sm:text-4xl'>Keep clpr moving.</h1>
                        <p className='mt-3 max-w-2xl text-sm leading-6 text-text-secondary sm:text-base'>One workspace for community safety, creator discovery, content operations, and platform health.</p>
                    </div>
                </header>

                <section className='mb-9' aria-labelledby='priority-heading'>
                    <div className='mb-4 flex items-end justify-between'>
                        <div><p className='text-xs font-bold uppercase tracking-[0.16em] text-brand'>Start here</p><h2 id='priority-heading' className='mt-1 text-xl font-bold text-text-primary'>Priority workflows</h2></div>
                        <span className='hidden items-center gap-1.5 text-xs text-text-tertiary sm:flex'><Clock3 className='h-3.5 w-3.5' /> Live workspace</span>
                    </div>
                    <div className='grid gap-4 lg:grid-cols-3'>
                        {priorityActions.map(action => {
                            const Icon = action.icon;
                            return (
                                <Link key={action.href} to={action.href} className='group relative overflow-hidden rounded-xl border border-border bg-surface p-5 transition-all hover:-translate-y-0.5 hover:border-brand/60 hover:shadow-lg'>
                                    <div className='mb-5 flex items-start justify-between'><span className='rounded-lg bg-brand/10 p-2.5 text-brand'><Icon className='h-5 w-5' /></span><ArrowUpRight className='h-4 w-4 text-text-tertiary transition-transform group-hover:translate-x-0.5 group-hover:-translate-y-0.5 group-hover:text-brand' /></div>
                                    <h3 className='font-semibold text-text-primary'>{action.label}</h3>
                                    <p className='mt-1.5 text-sm leading-5 text-text-secondary'>{action.description}</p>
                                </Link>
                            );
                        })}
                    </div>
                </section>

                <section aria-labelledby='workspace-heading'>
                    <div className='mb-4'><p className='text-xs font-bold uppercase tracking-[0.16em] text-brand'>Directory</p><h2 id='workspace-heading' className='mt-1 text-xl font-bold text-text-primary'>Every admin workspace</h2></div>
                    <div className='grid gap-4 xl:grid-cols-2'>
                        {adminNavGroups.filter(group => group.label !== 'Overview').map(group => (
                            <div key={group.label} className='rounded-xl border border-border bg-surface p-5'>
                                <h3 className='mb-3 text-xs font-bold uppercase tracking-[0.16em] text-text-tertiary'>{group.label}</h3>
                                <div className='divide-y divide-border'>
                                    {group.items.map(item => {
                                        const Icon = item.icon;
                                        return (
                                            <Link key={item.href} to={item.href} className='group flex items-center gap-3 py-3 first:pt-1 last:pb-1'>
                                                <span className='rounded-md bg-surface-raised p-2 text-text-secondary group-hover:bg-brand/10 group-hover:text-brand'><Icon className='h-4 w-4' /></span>
                                                <span className='min-w-0 flex-1'><span className='block text-sm font-medium text-text-primary'>{item.label}</span><span className='block truncate text-xs text-text-tertiary'>{item.description}</span></span>
                                                <ArrowUpRight className='h-4 w-4 text-text-tertiary group-hover:text-brand' />
                                            </Link>
                                        );
                                    })}
                                </div>
                            </div>
                        ))}
                    </div>
                </section>
            </div>
        </>
    );
}
