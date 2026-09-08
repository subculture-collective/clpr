import { Link } from 'react-router-dom';
import { ArrowUpRight, CheckCircle2, ShieldCheck, Sparkles } from 'lucide-react';
import { adminNavGroups } from '../../components/admin/adminNavigation';
import { SEO } from '../../components';

const priorityActions = [
    { label: 'Review moderation queue', href: '/admin/moderation', description: 'Work through the highest-priority reports and flagged content.', icon: ShieldCheck },
    { label: 'Review creator verification', href: '/admin/verification', description: 'Approve or reject pending creator applications.', icon: CheckCircle2 },
    { label: 'Check collection automation', href: '/admin/playlist-scripts', description: 'Inspect generated playlists and scheduling rules.', icon: Sparkles },
];

const dashboardGroups = adminNavGroups
    .map(group => ({ ...group, items: group.items.filter(item => item.href !== '/admin/dashboard') }))
    .filter(group => group.items.length > 0);

export function AdminDashboard() {
    const today = new Intl.DateTimeFormat('en-US', {
        weekday: 'long', month: 'long', day: 'numeric',
    }).format(new Date());

    return (
        <>
            <SEO title='Administration' noindex />
            <div className='mx-auto max-w-7xl'>
                <header className='mb-6 border-b border-border pb-5'>
                    <p className='mb-2 text-sm text-muted-foreground'>{today}</p>
                    <h1 className='text-2xl font-semibold text-text-primary'>Administration</h1>
                    <p className='mt-2 text-sm text-text-secondary'>Review community activity and manage content.</p>
                </header>

                <section className='mb-9' aria-labelledby='priority-heading'>
                    <div className='mb-4 flex items-end justify-between'>
                        <div><p className='text-xs font-bold uppercase tracking-[0.16em] text-link'>Start here</p><h2 id='priority-heading' className='mt-1 text-xl font-bold text-text-primary'>Priority workflows</h2></div>

                    </div>
                    <div className='grid grid-cols-[minmax(0,1fr)] gap-3 lg:grid-cols-3'>
                        {priorityActions.map(action => {
                            const Icon = action.icon;
                            return (
                                <Link key={action.href} to={action.href} className='group relative min-w-0 rounded-lg border border-border bg-surface p-4 transition-colors hover:bg-surface-hover'>
                                    <div className='mb-3 flex items-start justify-between'><span className='rounded-lg bg-brand/10 p-2.5 text-link'><Icon className='h-5 w-5' /></span><ArrowUpRight className='h-4 w-4 text-text-tertiary transition-transform group-hover:translate-x-0.5 group-hover:-translate-y-0.5 group-hover:text-link' /></div>
                                    <h3 className='font-semibold text-text-primary'>{action.label}</h3>
                                    <p className='mt-1.5 text-sm leading-5 text-text-secondary'>{action.description}</p>
                                </Link>
                            );
                        })}
                    </div>
                </section>

                <section aria-labelledby='workspace-heading'>
                    <div className='mb-4'><p className='text-xs font-bold uppercase tracking-[0.16em] text-link'>Directory</p><h2 id='workspace-heading' className='mt-1 text-xl font-bold text-text-primary'>Every admin workspace</h2></div>
                    <div className='grid grid-cols-[minmax(0,1fr)] gap-4 xl:grid-cols-2'>
                        {dashboardGroups.map(group => (
                            <div key={group.label} className='min-w-0 rounded-lg border border-border bg-surface p-4'>
                                <h3 className='mb-3 text-xs font-bold uppercase tracking-[0.16em] text-text-tertiary'>{group.label}</h3>
                                <div className='divide-y divide-border'>
                                    {group.items.map(item => {
                                        const Icon = item.icon;
                                        return (
                                            <Link key={item.href} to={item.href} className='group flex items-center gap-3 py-3 first:pt-1 last:pb-1'>
                                                <span className='rounded-md bg-surface-raised p-2 text-text-secondary group-hover:bg-brand/10 group-hover:text-link'><Icon className='h-4 w-4' /></span>
                                                <span className='min-w-0 flex-1'><span className='block text-sm font-medium text-text-primary'>{item.label}</span><span className='block truncate text-xs text-text-tertiary'>{item.description}</span></span>
                                                <ArrowUpRight className='h-4 w-4 text-text-tertiary group-hover:text-link' />
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
