import { useState, useRef } from 'react';
import { useTranslation } from 'react-i18next';
import { Link, useNavigate, useLocation } from 'react-router-dom';
import { useAuth } from '../../context/AuthContext';
import { useClickOutside } from '../../hooks/useClickOutside';
import { useKeyboardShortcuts } from '../../hooks/useKeyboardShortcuts';
import { Button } from '../ui/Button';
import { NotificationBell } from './NotificationBell';
import { UserMenu } from './UserMenu';
import {
    Home,
    Compass,
    Heart,
    Trophy,
    ListMusic,
    Sparkles,
    Star,
    ClipboardList,
    Clock,
    Users,
    User,
    Settings,
    LogOut,
    Menu,
    X,
    MoreHorizontal,
    ChevronDown,
    Search,
} from 'lucide-react';

export function Header() {
    const { t } = useTranslation();
    const { isAuthenticated, logout } = useAuth();
    const navigate = useNavigate();
    const location = useLocation();
    const [mobileMenuOpen, setMobileMenuOpen] = useState(false);
    const [moreMenuOpen, setMoreMenuOpen] = useState(false);
    const moreMenuRef = useRef<HTMLDivElement>(null);
    const moreButtonRef = useRef<HTMLButtonElement>(null);
    const mobileButtonRef = useRef<HTMLButtonElement>(null);

    const handleLogout = async () => {
        await logout();
        setMobileMenuOpen(false);
        navigate('/');
    };

    // Close More menu when clicking outside
    useClickOutside(moreMenuRef, () => setMoreMenuOpen(false), moreMenuOpen);

    // Keyboard shortcuts
    useKeyboardShortcuts([
        {
            key: 'Escape',
            callback: () => {
                if (mobileMenuOpen) { setMobileMenuOpen(false); mobileButtonRef.current?.focus(); }
                if (moreMenuOpen) { setMoreMenuOpen(false); moreButtonRef.current?.focus(); }
            },
            description: 'Close menus',
        },
    ]);

    return (
        <header className='sticky top-0 z-50 bg-background border-b border-border'>
            <div className='page-container'>
                <div className='flex items-center justify-between h-16'>
                    {/* Logo */}
                    <Link
                        to='/'
                        className='flex min-h-11 items-center cursor-pointer focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-primary-500 focus-visible:ring-offset-2 rounded-md'
                        aria-label='clpr.tv home'
                    >
                        <img
                            src='/clpr-logo.svg'
                            alt=''
                            aria-hidden='true'
                            className='h-9 w-auto'
                        />
                    </Link>

                    {/* Navigation (desktop) */}
                    <nav
                        className='hidden md:flex items-center gap-1'
                        aria-label='Main navigation'
                        data-testid='main-nav'
                    >
                        <Button asChild variant='ghost' size='sm'>
                            <Link to='/' aria-current={location.pathname === '/' ? 'page' : undefined} className={`relative ${location.pathname === '/' ? 'after:absolute after:bottom-0 after:left-2 after:right-2 after:h-0.5 after:bg-brand after:rounded-full' : ''}`}>
                                <Home size={16} strokeWidth={1.75} className='mr-1.5' /> Feed
                            </Link>
                        </Button>
                        <Button asChild variant='ghost' size='sm'>
                            <Link to='/discover' className='aria-[current=page]:bg-surface-hover aria-[current=page]:text-link' aria-current={location.pathname.startsWith('/discover') ? 'page' : undefined}>
                                <Compass size={16} strokeWidth={1.75} className='mr-1.5' /> Discover
                            </Link>
                        </Button>
                        <Button asChild variant='ghost' size='sm'>
                            <Link className='aria-[current=page]:bg-surface-hover aria-[current=page]:text-link' to={isAuthenticated ? '/favorites' : '/login'} state={!isAuthenticated ? { from: { pathname: '/favorites' } } : undefined} aria-current={location.pathname === '/favorites' ? 'page' : undefined}>
                                <Heart size={16} strokeWidth={1.75} className='mr-1.5' /> Saved
                            </Link>
                        </Button>
                        {/* More dropdown */}
                        <div className='relative' ref={moreMenuRef}>
                            <Button
                                variant='ghost'
                                size='sm'
                                ref={moreButtonRef}
                                onClick={() => setMoreMenuOpen(!moreMenuOpen)}
                                aria-expanded={moreMenuOpen}
                                aria-controls={moreMenuOpen ? 'more-navigation' : undefined}
                            >
                                <MoreHorizontal size={16} strokeWidth={1.75} className='mr-1.5' /> More
                                <ChevronDown
                                    size={16}
                                    strokeWidth={1.75}
                                    className={`ml-1 transition-transform ${moreMenuOpen ? 'rotate-180' : ''}`}
                                />
                            </Button>

                            {moreMenuOpen && (
                                <div
                                    className='absolute left-0 mt-1 w-48 bg-background border border-border rounded-md shadow-lg overflow-hidden z-50'
                                    id='more-navigation'
                                >
                                    <Link to='/creators' className='flex min-h-11 items-center gap-2 px-4 py-2 text-sm hover:bg-muted' onClick={() => setMoreMenuOpen(false)}>
                                        <Users size={16} strokeWidth={1.75} /> Creators
                                    </Link>
                                    <Link
                                        to='/leaderboards'
                                        className='flex min-h-11 items-center gap-2 px-4 py-2 text-sm hover:bg-muted transition-colors'
                                        onClick={() => setMoreMenuOpen(false)}
                                    >
                                        <Trophy size={16} strokeWidth={1.75} /> {t('nav.leaderboards')}
                                    </Link>
                                    <Link
                                        to='/playlists/discover'
                                        className='flex min-h-11 items-center gap-2 px-4 py-2 text-sm hover:bg-muted transition-colors'
                                        onClick={() => setMoreMenuOpen(false)}
                                    >
                                        <ListMusic size={16} strokeWidth={1.75} /> Playlists
                                    </Link>




                                </div>
                            )}
                        </div>
                    </nav>

                    {/* Right Side Actions */}
                    <div className='flex items-center gap-2'>
                        <Button asChild variant='ghost' size='sm' className='md:hidden'>
                            <Link to='/search' aria-label='Search clips and creators'>
                                <Search size={20} strokeWidth={1.75} />
                            </Link>
                        </Button>
                        {/* User Menu or Login */}
                        {isAuthenticated ?
                            <div className='hidden md:flex items-center gap-2'>
                                <Button asChild variant='primary' size='sm'>
                                    <Link to='/submit'>
                                        {t('nav.submit')}
                                    </Link>
                                </Button>
                                <NotificationBell />
                                <UserMenu />
                            </div>
                        :   <Button asChild variant='primary' size='sm' className='hidden md:inline-flex'>
                                <Link
                                    to='/login'
                                    state={{ from: location }}
                                    data-testid='login-button'
                                    aria-label='Login'
                                >
                                    {t('nav.login')}
                                </Link>
                            </Button>
                        }

                        {/* Mobile Menu Button */}
                        <Button
                            variant='ghost'
                            size='sm'
                            className='md:hidden'
                            ref={mobileButtonRef}
                            onClick={() => setMobileMenuOpen(!mobileMenuOpen)}
                            aria-controls={mobileMenuOpen ? 'mobile-navigation' : undefined}
                            aria-label={
                                mobileMenuOpen ? 'Close menu' : 'Open menu'
                            }
                            aria-expanded={mobileMenuOpen}
                            data-testid='mobile-menu-toggle'
                        >
                            {mobileMenuOpen ? <X size={20} strokeWidth={1.75} /> : <Menu size={20} strokeWidth={1.75} />}
                        </Button>
                    </div>
                </div>

                {/* Mobile Menu */}
                {mobileMenuOpen && (
                    <div
                        className='md:hidden py-4 border-t border-border'
                        id='mobile-navigation'
                    >
                        <nav className='flex flex-col gap-1 mb-4' aria-label='More mobile destinations'>
                            <Button asChild variant='ghost' size='sm' className='w-full justify-start'>
                                <Link
                                    to='/'
                                    onClick={() => setMobileMenuOpen(false)}
                                    data-testid='mobile-nav-home'
                                >
                                    <Home size={16} strokeWidth={1.75} className='mr-2' /> Feed
                                </Link>
                            </Button>
                            <Button asChild variant='ghost' size='sm' className='w-full justify-start'>
                                <Link
                                    to='/creators'
                                    onClick={() => setMobileMenuOpen(false)}
                                >
                                    <Users size={16} strokeWidth={1.75} className='mr-2' /> Creators
                                </Link>
                            </Button>

                            <div className='border-t border-border my-2'></div>
                            <p className='px-3 text-xs text-muted-foreground uppercase tracking-wide'>
                                Explore
                            </p>

                            <Button asChild variant='ghost' size='sm' className='w-full justify-start'>
                                <Link
                                    to='/leaderboards'
                                    onClick={() => setMobileMenuOpen(false)}
                                >
                                    <Trophy size={16} strokeWidth={1.75} className='mr-2' /> {t('nav.leaderboards')}
                                </Link>
                            </Button>
                            <Button asChild variant='ghost' size='sm' className='w-full justify-start'>
                                <Link
                                    to='/playlists/discover'
                                    onClick={() => setMobileMenuOpen(false)}
                                >
                                    <ListMusic size={16} strokeWidth={1.75} className='mr-2' /> Playlists
                                </Link>
                            </Button>




                        </nav>

                        {isAuthenticated ?
                            <div className='flex flex-col gap-1'>
                                <div className='border-t border-border my-2'></div>
                                <p className='px-3 text-xs text-muted-foreground uppercase tracking-wide'>
                                    Your Stuff
                                </p>

                                <Button asChild variant='primary' size='sm' className='w-full'>
                                    <Link
                                        to='/submit'
                                        onClick={() => setMobileMenuOpen(false)}
                                    >
                                        <Sparkles size={16} strokeWidth={1.75} className='mr-2' /> {t('nav.submit')}
                                    </Link>
                                </Button>
                                <Button asChild variant='ghost' size='sm' className='w-full justify-start'>
                                    <Link
                                        to='/favorites'
                                        onClick={() => setMobileMenuOpen(false)}
                                    >
                                        <Star size={16} strokeWidth={1.75} className='mr-2' /> {t('nav.favorites')}
                                    </Link>
                                </Button>
                                <Button asChild variant='ghost' size='sm' className='w-full justify-start'>
                                    <Link
                                        to='/playlists'
                                        onClick={() => setMobileMenuOpen(false)}
                                    >
                                        <ClipboardList size={16} strokeWidth={1.75} className='mr-2' /> My Playlists
                                    </Link>
                                </Button>
                                <Button asChild variant='ghost' size='sm' className='w-full justify-start'>
                                    <Link
                                        to='/watch-history'
                                        onClick={() => setMobileMenuOpen(false)}
                                    >
                                        <Clock size={16} strokeWidth={1.75} className='mr-2' /> Watch History
                                    </Link>
                                </Button>

                                <div className='border-t border-border my-2'></div>
                                <p className='px-3 text-xs text-muted-foreground uppercase tracking-wide'>
                                    Account
                                </p>

                                <Button asChild variant='ghost' size='sm' className='w-full justify-start'>
                                    <Link
                                        to='/profile'
                                        onClick={() => setMobileMenuOpen(false)}
                                    >
                                        <User size={16} strokeWidth={1.75} className='mr-2' /> {t('nav.profile')}
                                    </Link>
                                </Button>
                                <Button asChild variant='ghost' size='sm' className='w-full justify-start'>
                                    <Link
                                        to='/settings'
                                        onClick={() => setMobileMenuOpen(false)}
                                    >
                                        <Settings size={16} strokeWidth={1.75} className='mr-2' /> {t('nav.settings')}
                                    </Link>
                                </Button>
                                <Button
                                    variant='ghost'
                                    size='sm'
                                    className='w-full justify-start text-error-600'
                                    onClick={handleLogout}
                                >
                                    <LogOut size={16} strokeWidth={1.75} className='mr-2' /> {t('nav.logout')}
                                </Button>
                            </div>
                        :   <Button asChild variant='primary' size='sm' className='w-full'>
                                <Link
                                    to='/login'
                                    state={{ from: location }}
                                    onClick={() => setMobileMenuOpen(false)}
                                    data-testid='login-button'
                                    aria-label='Login'
                                >
                                    {t('nav.login')}
                                </Link>
                            </Button>
                        }
                    </div>
                )}
            </div>
        </header>
    );
}
