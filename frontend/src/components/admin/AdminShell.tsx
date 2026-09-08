import { useState } from 'react';
import type { ReactNode } from 'react';
import { NavLink, useLocation } from 'react-router-dom';
import { ChevronDown, Menu, X } from 'lucide-react';
import { adminNavGroups, adminNavItems } from './adminNavigation';
import { SEO } from '../SEO';

function itemIsActive(pathname: string, href: string) {
    if (href === '/admin/dashboard') return pathname === href;
    return pathname === href || pathname.startsWith(`${href}/`);
}

function AdminNavigation({ onNavigate }: { onNavigate?: () => void }) {
    const location = useLocation();
    const activeHref = [...adminNavItems]
        .sort((a, b) => b.href.length - a.href.length)
        .find(item => itemIsActive(location.pathname, item.href))?.href;
    return (
        <nav aria-label='Administration' className='space-y-6'>
            {adminNavGroups.map(group => (
                <div key={group.label}>
                    <p className='mb-2 px-3 text-[10px] font-bold uppercase tracking-[0.2em] text-text-tertiary'>{group.label}</p>
                    <div className='space-y-1'>
                        {group.items.map(item => {
                            const Icon = item.icon;
                            const active = item.href === activeHref;
                            return (
                                <NavLink
                                    key={item.href}
                                    to={item.href}
                                    onClick={onNavigate}
                                    aria-current={active ? 'page' : undefined}
                                    className={`group flex min-h-11 items-center gap-3 rounded-lg px-3 py-2 text-sm transition-colors ${active ? 'bg-brand text-white shadow-sm' : 'text-text-secondary hover:bg-surface-hover hover:text-text-primary'}`}
                                >
                                    <Icon className='h-4 w-4 shrink-0' strokeWidth={1.8} />
                                    <span className='truncate font-medium'>{item.label}</span>
                                </NavLink>
                            );
                        })}
                    </div>
                </div>
            ))}
        </nav>
    );
}

export function AdminShell({ children }: { children: ReactNode }) {
    const [mobileOpen, setMobileOpen] = useState(false);
    const location = useLocation();
    const current = [...adminNavItems]
        .sort((a, b) => b.href.length - a.href.length)
        .find(item => itemIsActive(location.pathname, item.href));

    return (
        <><SEO title={`${current?.label ?? 'Administration'} - Admin`} description={current?.description ?? 'clpr administration workspace.'} noindex /><div className='admin-workspace min-h-[calc(100vh-4rem)] bg-background'>
            <div className='border-b border-border bg-surface/80 backdrop-blur-xl lg:hidden'>
                <div className='flex items-center justify-between px-4 py-3'>
                    <div>
                        <p className='text-[10px] font-bold uppercase tracking-[0.2em] text-link'>clpr administration</p>
                        <p className='text-sm font-semibold text-text-primary'>{current?.label ?? 'Administration'}</p>
                    </div>
                    <button type='button' onClick={() => setMobileOpen(value => !value)} className='flex min-h-11 min-w-11 items-center justify-center rounded-lg border border-border bg-surface-raised p-2 text-text-primary' aria-expanded={mobileOpen} aria-controls={mobileOpen ? 'admin-mobile-navigation' : undefined} aria-label={mobileOpen ? 'Close admin navigation' : 'Open admin navigation'}>
                        {mobileOpen ? <X className='h-5 w-5' /> : <Menu className='h-5 w-5' />}
                    </button>
                </div>
                {mobileOpen && <div id='admin-mobile-navigation' className='max-h-[70vh] overflow-y-auto border-t border-border px-3 py-4'><AdminNavigation onNavigate={() => setMobileOpen(false)} /></div>}
            </div>

            <div className='mx-auto grid min-w-0 w-full max-w-[1600px] grid-cols-[minmax(0,1fr)] lg:grid-cols-[252px_minmax(0,1fr)]'>
                <aside className='hidden border-r border-border bg-surface/45 lg:block'>
                    <div className='sticky top-0 max-h-screen overflow-y-auto px-4 py-7'>
                        <div className='mb-7 px-3'>
                            <h2 className='font-accent text-xl font-extrabold tracking-tight text-text-primary'>Administration</h2>
                            <p className='mt-1 text-xs leading-5 text-text-secondary'>Moderation, discovery, and platform operations.</p>
                        </div>
                        <AdminNavigation />
                    </div>
                </aside>

                <section className='min-w-0'>
                    <div className='hidden items-center justify-between border-b border-border px-8 py-3 lg:flex'>
                        <div className='flex items-center gap-2 text-xs text-text-secondary'>
                            <span>Admin</span><ChevronDown className='h-3 w-3 -rotate-90' />
                            <span className='font-medium text-text-primary'>{current?.label ?? 'Workspace'}</span>
                        </div>
                        <p className='text-xs text-text-tertiary'>{current?.description}</p>
                    </div>
                    <div className='admin-content min-w-0 px-4 py-6 sm:px-6 lg:px-8 lg:py-8'>{children}</div>
                </section>
            </div>
        </div></>
    );
}
