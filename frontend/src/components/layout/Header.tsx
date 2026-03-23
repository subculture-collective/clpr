import { useState, useRef } from 'react';
import { useTranslation } from 'react-i18next';
import { Link, useNavigate, useLocation } from 'react-router-dom';
import { useAuth } from '../../context/AuthContext';
import { useClickOutside } from '../../hooks/useClickOutside';
import { useKeyboardShortcuts } from '../../hooks/useKeyboardShortcuts';
import { Button } from '../ui';
import { NotificationBell } from './NotificationBell';
import { UserMenu } from './UserMenu';
import {
    Home,
    Search,
    MessageSquare,
    Trophy,
    ListMusic,
    Sparkles,
    Star,
    ClipboardList,
    Clock,
    User,
    Settings,
    LogOut,
    Menu,
    X,
    MoreHorizontal,
    ChevronDown,
} from 'lucide-react';

export function Header() {
    const { t } = useTranslation();
    const { isAuthenticated, logout } = useAuth();
    const navigate = useNavigate();
    const location = useLocation();
    const [mobileMenuOpen, setMobileMenuOpen] = useState(false);
    const [moreMenuOpen, setMoreMenuOpen] = useState(false);
    const moreMenuRef = useRef<HTMLDivElement>(null);

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
                if (mobileMenuOpen) setMobileMenuOpen(false);
                if (moreMenuOpen) setMoreMenuOpen(false);
            },
            description: 'Close menus',
        },
    ]);

    return (
        <header className='sticky top-0 z-50 bg-background border-b border-border'>
            <div className='container mx-auto px-4'>
                <div className='flex items-center justify-between h-16'>
                    {/* Logo */}
                    <Link
                        to='/'
                        className='flex items-center gap-2 cursor-pointer focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-primary-500 focus-visible:ring-offset-2 rounded-md'
                        aria-label='clpr.tv home'
                    >
                        <svg width='20' height='20' viewBox='0 0 20 20' fill='none' aria-hidden='true'>
                            <path
                                d='M5 3.87a1 1 0 011.53-.85l9.94 6.13a1 1 0 010 1.7l-9.94 6.13A1 1 0 015 16.13V3.87z'
                                fill='rgb(var(--color-brand))'
                            />
                        </svg>
                        <span className='font-heading text-[22px] font-bold text-gradient'>clpr</span>
                        <span className='font-heading text-sm font-medium text-text-secondary'>.tv</span>
                    </Link>

                    {/* Navigation (desktop) */}
                    <nav
                        className='hidden md:flex items-center gap-1'
                        aria-label='Main navigation'
                        data-testid='main-nav'
                    >
                        <Link to='/' className={`relative ${location.pathname === '/' ? 'after:absolute after:bottom-0 after:left-2 after:right-2 after:h-0.5 after:bg-brand after:rounded-full' : ''}`}>
                            <Button variant='ghost' size='sm'>
                                <Home size={16} strokeWidth={1.75} className='mr-1.5' /> Feed
                            </Button>
                        </Link>
                        <Link to='/discover' className={`relative ${location.pathname === '/discover' ? 'after:absolute after:bottom-0 after:left-2 after:right-2 after:h-0.5 after:bg-brand after:rounded-full' : ''}`}>
                            <Button variant='ghost' size='sm'>
                                <Search size={16} strokeWidth={1.75} className='mr-1.5' /> Discover
                            </Button>
                        </Link>
                        <Link to='/forum' className={`relative ${location.pathname.startsWith('/forum') ? 'after:absolute after:bottom-0 after:left-2 after:right-2 after:h-0.5 after:bg-brand after:rounded-full' : ''}`}>
                            <Button variant='ghost' size='sm'>
                                <MessageSquare size={16} strokeWidth={1.75} className='mr-1.5' /> Forum
                            </Button>
                        </Link>

                        {/* More dropdown */}
                        <div className='relative' ref={moreMenuRef}>
                            <Button
                                variant='ghost'
                                size='sm'
                                onClick={() => setMoreMenuOpen(!moreMenuOpen)}
                                aria-expanded={moreMenuOpen}
                                aria-haspopup='true'
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
                                    role='menu'
                                >
                                    <Link
                                        to='/leaderboards'
                                        className='flex items-center gap-2 px-4 py-2 text-sm hover:bg-muted transition-colors'
                                        onClick={() => setMoreMenuOpen(false)}
                                        role='menuitem'
                                    >
                                        <Trophy size={16} strokeWidth={1.75} /> {t('nav.leaderboards')}
                                    </Link>
                                    <Link
                                        to='/playlists/discover'
                                        className='flex items-center gap-2 px-4 py-2 text-sm hover:bg-muted transition-colors'
                                        onClick={() => setMoreMenuOpen(false)}
                                        role='menuitem'
                                    >
                                        <ListMusic size={16} strokeWidth={1.75} /> Playlists
                                    </Link>
                                    {/* Watch Parties - Hidden until after launch */}
                                    {/* <Link
                                        to='/watch-parties/browse'
                                        className='block px-4 py-2 text-sm hover:bg-muted transition-colors'
                                        onClick={() => setMoreMenuOpen(false)}
                                        role='menuitem'
                                    >
                                        👥 Watch Parties
                                    </Link> */}
                                    {/* Live Feed - Hidden until after launch */}
                                    {/* {isAuthenticated && (
                                        <Link
                                            to='/discover/live'
                                            className='block px-4 py-2 text-sm hover:bg-muted transition-colors'
                                            onClick={() =>
                                                setMoreMenuOpen(false)
                                            }
                                            role='menuitem'
                                        >
                                            🔴 Live
                                        </Link>
                                    )} */}
                                </div>
                            )}
                        </div>
                    </nav>

                    {/* Right Side Actions */}
                    <div className='flex items-center gap-2'>
                        {/* User Menu or Login */}
                        {isAuthenticated ?
                            <div className='hidden md:flex items-center gap-2'>
                                <Link to='/submit'>
                                    <Button variant='primary' size='sm'>
                                        {t('nav.submit')}
                                    </Button>
                                </Link>
                                <NotificationBell />
                                <UserMenu />
                            </div>
                        :   <Link to='/login' className='hidden md:block'>
                                <Button
                                    variant='primary'
                                    size='sm'
                                    data-testid='login-button'
                                    aria-label='Login'
                                >
                                    {t('nav.login')}
                                </Button>
                            </Link>
                        }

                        {/* Mobile Menu Button */}
                        <Button
                            variant='ghost'
                            size='sm'
                            className='md:hidden'
                            onClick={() => setMobileMenuOpen(!mobileMenuOpen)}
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
                        role='navigation'
                        aria-label='Mobile navigation'
                    >
                        <nav className='flex flex-col gap-1 mb-4'>
                            <Link
                                to='/'
                                onClick={() => setMobileMenuOpen(false)}
                            >
                                <Button
                                    variant='ghost'
                                    size='sm'
                                    className='w-full justify-start'
                                    data-testid='mobile-nav-home'
                                >
                                    <Home size={16} strokeWidth={1.75} className='mr-2' /> Feed
                                </Button>
                            </Link>
                            <Link
                                to='/discover'
                                onClick={() => setMobileMenuOpen(false)}
                            >
                                <Button
                                    variant='ghost'
                                    size='sm'
                                    className='w-full justify-start'
                                >
                                    <Search size={16} strokeWidth={1.75} className='mr-2' /> Discover
                                </Button>
                            </Link>
                            <Link
                                to='/forum'
                                onClick={() => setMobileMenuOpen(false)}
                            >
                                <Button
                                    variant='ghost'
                                    size='sm'
                                    className='w-full justify-start'
                                >
                                    <MessageSquare size={16} strokeWidth={1.75} className='mr-2' /> Forum
                                </Button>
                            </Link>

                            <div className='border-t border-border my-2'></div>
                            <p className='px-3 text-xs text-muted-foreground uppercase tracking-wide'>
                                Explore
                            </p>

                            <Link
                                to='/leaderboards'
                                onClick={() => setMobileMenuOpen(false)}
                            >
                                <Button
                                    variant='ghost'
                                    size='sm'
                                    className='w-full justify-start'
                                >
                                    <Trophy size={16} strokeWidth={1.75} className='mr-2' /> {t('nav.leaderboards')}
                                </Button>
                            </Link>
                            <Link
                                to='/playlists/discover'
                                onClick={() => setMobileMenuOpen(false)}
                            >
                                <Button
                                    variant='ghost'
                                    size='sm'
                                    className='w-full justify-start'
                                >
                                    <ListMusic size={16} strokeWidth={1.75} className='mr-2' /> Playlists
                                </Button>
                            </Link>
                            {/* Watch Parties - Hidden until after launch */}
                            {/* <Link
                                to='/watch-parties/browse'
                                onClick={() => setMobileMenuOpen(false)}
                            >
                                <Button
                                    variant='ghost'
                                    size='sm'
                                    className='w-full justify-start'
                                >
                                    👥 Watch Parties
                                </Button>
                            </Link> */}
                            {/* Live Feed - Hidden until after launch */}
                            {/* {isAuthenticated && (
                                <Link
                                    to='/discover/live'
                                    onClick={() => setMobileMenuOpen(false)}
                                >
                                    <Button
                                        variant='ghost'
                                        size='sm'
                                        className='w-full justify-start'
                                    >
                                        🔴 Live
                                    </Button>
                                </Link>
                            )} */}
                        </nav>

                        {isAuthenticated ?
                            <div className='flex flex-col gap-1'>
                                <div className='border-t border-border my-2'></div>
                                <p className='px-3 text-xs text-muted-foreground uppercase tracking-wide'>
                                    Your Stuff
                                </p>

                                <Link
                                    to='/submit'
                                    onClick={() => setMobileMenuOpen(false)}
                                >
                                    <Button
                                        variant='primary'
                                        size='sm'
                                        className='w-full'
                                    >
                                        <Sparkles size={16} strokeWidth={1.75} className='mr-2' /> {t('nav.submit')}
                                    </Button>
                                </Link>
                                <Link
                                    to='/favorites'
                                    onClick={() => setMobileMenuOpen(false)}
                                >
                                    <Button
                                        variant='ghost'
                                        size='sm'
                                        className='w-full justify-start'
                                    >
                                        <Star size={16} strokeWidth={1.75} className='mr-2' /> {t('nav.favorites')}
                                    </Button>
                                </Link>
                                <Link
                                    to='/playlists'
                                    onClick={() => setMobileMenuOpen(false)}
                                >
                                    <Button
                                        variant='ghost'
                                        size='sm'
                                        className='w-full justify-start'
                                    >
                                        <ClipboardList size={16} strokeWidth={1.75} className='mr-2' /> My Playlists
                                    </Button>
                                </Link>
                                <Link
                                    to='/watch-history'
                                    onClick={() => setMobileMenuOpen(false)}
                                >
                                    <Button
                                        variant='ghost'
                                        size='sm'
                                        className='w-full justify-start'
                                    >
                                        <Clock size={16} strokeWidth={1.75} className='mr-2' /> Watch History
                                    </Button>
                                </Link>

                                <div className='border-t border-border my-2'></div>
                                <p className='px-3 text-xs text-muted-foreground uppercase tracking-wide'>
                                    Account
                                </p>

                                <Link
                                    to='/profile'
                                    onClick={() => setMobileMenuOpen(false)}
                                >
                                    <Button
                                        variant='ghost'
                                        size='sm'
                                        className='w-full justify-start'
                                    >
                                        <User size={16} strokeWidth={1.75} className='mr-2' /> {t('nav.profile')}
                                    </Button>
                                </Link>
                                <Link
                                    to='/settings'
                                    onClick={() => setMobileMenuOpen(false)}
                                >
                                    <Button
                                        variant='ghost'
                                        size='sm'
                                        className='w-full justify-start'
                                    >
                                        <Settings size={16} strokeWidth={1.75} className='mr-2' /> {t('nav.settings')}
                                    </Button>
                                </Link>
                                <Button
                                    variant='ghost'
                                    size='sm'
                                    className='w-full justify-start text-error-600'
                                    onClick={handleLogout}
                                >
                                    <LogOut size={16} strokeWidth={1.75} className='mr-2' /> {t('nav.logout')}
                                </Button>
                            </div>
                        :   <Link
                                to='/login'
                                onClick={() => setMobileMenuOpen(false)}
                            >
                                <Button
                                    variant='primary'
                                    size='sm'
                                    className='w-full'
                                    data-testid='login-button'
                                    aria-label='Login'
                                >
                                    {t('nav.login')}
                                </Button>
                            </Link>
                        }
                    </div>
                )}
            </div>
        </header>
    );
}
