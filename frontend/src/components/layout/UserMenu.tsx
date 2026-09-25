import { Avatar } from '../ui/Avatar';
import { useState, useRef, useEffect } from 'react';
import { Link, useNavigate } from 'react-router-dom';
import { useAuth } from '../../context/AuthContext';
import { useClickOutside } from '../../hooks/useClickOutside';
import { useMenuKeyboard } from '../../hooks/useMenuKeyboard';
import { UserRoleBadge } from '../user';
import type { UserRole } from '../../lib/roles';
import {
    User,
    Settings,
    Star,
    ListMusic,
    Clock,
    ClipboardList,
    Upload,
    Shield,
    LogOut,
    ChevronDown,
} from 'lucide-react';

export function UserMenu() {
    const { user, logout, isModeratorOrAdmin } = useAuth();
    const [isOpen, setIsOpen] = useState(false);
    const containerRef = useRef<HTMLDivElement>(null);
    const buttonRef = useRef<HTMLButtonElement>(null);
    const navigate = useNavigate();

    // Calculate menu item count
    const menuItemCount = isModeratorOrAdmin ? 9 : 8; // Profile, Settings, Favorites, Playlists, Watch History, Queue, Submissions, (Admin), Logout
    const { menuRef } = useMenuKeyboard(
        isOpen,
        () => setIsOpen(false),
        menuItemCount,
    );
    const prevIsOpenRef = useRef<boolean>(false);

    // Close menu when clicking outside
    useClickOutside(containerRef, () => setIsOpen(false), isOpen);

    // Return focus to button when menu transitions from open to closed
    useEffect(() => {
        if (prevIsOpenRef.current && !isOpen && buttonRef.current) {
            buttonRef.current.focus();
        }
        prevIsOpenRef.current = isOpen;
    }, [isOpen]);

    const handleLogout = async () => {
        await logout();
        setIsOpen(false);
        navigate('/');
    };

    if (!user) return null;

    return (
        <div className='relative' ref={containerRef}>
            {/* Avatar Button */}
            <button
                ref={buttonRef}
                onClick={() => setIsOpen(!isOpen)}
                className='flex min-h-11 items-center gap-2 px-2 py-1 hover:bg-muted transition-colors cursor-pointer focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-focus-ring focus-visible:ring-offset-2 focus-visible:ring-offset-background'
                aria-expanded={isOpen}
                aria-haspopup='true'
                aria-label='User menu'
                data-testid='user-menu'
            >
                <Avatar
                    src={user.avatar_url}
                    alt=''
                    fallback={user.username}
                    size='sm'
                    frameClassName='bg-primary-400 text-background'
                />
                <span className='hidden md:inline text-sm font-medium'>
                    {user.username}
                </span>
                <ChevronDown
                    size={16}
                    strokeWidth={1.75}
                    className={`transition-transform ${isOpen ? 'rotate-180' : ''}`}
                />
            </button>

            {/* Dropdown Menu */}
            {isOpen && (
                <div
                    ref={menuRef}
                    className='absolute right-0 mt-2 w-56 bg-popover border border-line-strong tally-bar overflow-hidden z-50'
                    role='menu'
                    aria-orientation='vertical'
                >
                    {/* User Info */}
                    <div className='px-4 py-3 border-b border-border'>
                        <p className='font-semibold'>{user.display_name}</p>
                        <p className='text-sm text-muted-foreground'>
                            @{user.username}
                        </p>
                        <div className='flex items-center gap-2 mt-2'>
                            <p className='text-xs text-muted-foreground'>
                                {user.karma_points} uppies
                            </p>
                            <span className='text-xs text-muted-foreground'>
                                •
                            </span>
                            <UserRoleBadge
                                role={user.role as UserRole}
                                size='sm'
                            />
                        </div>
                    </div>

                    {/* Menu Items */}
                    <div className='py-1'>
                        <Link
                            to='/profile'
                            className='flex items-center gap-2 px-4 py-2 text-sm hover:bg-muted transition-colors cursor-pointer focus-visible:outline-none focus-visible:bg-muted'
                            onClick={() => setIsOpen(false)}
                            role='menuitem'
                            tabIndex={-1}
                        >
                            <User size={16} strokeWidth={1.75} /> Profile
                        </Link>
                        <Link
                            to='/settings'
                            className='flex items-center gap-2 px-4 py-2 text-sm hover:bg-muted transition-colors cursor-pointer focus-visible:outline-none focus-visible:bg-muted'
                            onClick={() => setIsOpen(false)}
                            role='menuitem'
                            tabIndex={-1}
                        >
                            <Settings size={16} strokeWidth={1.75} /> Settings
                        </Link>
                        <Link
                            to='/favorites'
                            className='flex items-center gap-2 px-4 py-2 text-sm hover:bg-muted transition-colors cursor-pointer focus-visible:outline-none focus-visible:bg-muted'
                            onClick={() => setIsOpen(false)}
                            role='menuitem'
                            tabIndex={-1}
                        >
                            <Star size={16} strokeWidth={1.75} /> Favorites
                        </Link>
                        <Link
                            to='/playlists'
                            className='flex items-center gap-2 px-4 py-2 text-sm hover:bg-muted transition-colors cursor-pointer focus-visible:outline-none focus-visible:bg-muted'
                            onClick={() => setIsOpen(false)}
                            role='menuitem'
                            tabIndex={-1}
                        >
                            <ListMusic size={16} strokeWidth={1.75} /> Playlists
                        </Link>
                        <Link
                            to='/watch-history'
                            className='flex items-center gap-2 px-4 py-2 text-sm hover:bg-muted transition-colors cursor-pointer focus-visible:outline-none focus-visible:bg-muted'
                            onClick={() => setIsOpen(false)}
                            role='menuitem'
                            tabIndex={-1}
                        >
                            <Clock size={16} strokeWidth={1.75} /> Watch History
                        </Link>
                        <Link
                            to='/queue'
                            className='flex items-center gap-2 px-4 py-2 text-sm hover:bg-muted transition-colors cursor-pointer focus-visible:outline-none focus-visible:bg-muted'
                            onClick={() => setIsOpen(false)}
                            role='menuitem'
                            tabIndex={-1}
                        >
                            <ClipboardList size={16} strokeWidth={1.75} /> My Queue
                        </Link>
                        <Link
                            to='/submissions'
                            className='flex items-center gap-2 px-4 py-2 text-sm hover:bg-muted transition-colors cursor-pointer focus-visible:outline-none focus-visible:bg-muted'
                            onClick={() => setIsOpen(false)}
                            role='menuitem'
                            tabIndex={-1}
                        >
                            <Upload size={16} strokeWidth={1.75} /> My Submissions
                        </Link>

                        {isModeratorOrAdmin && (
                            <>
                                <div className='border-t border-border my-1'></div>
                                <Link
                                    to='/admin/dashboard'
                                    className='flex items-center gap-2 px-4 py-2 text-sm hover:bg-muted transition-colors text-link cursor-pointer focus-visible:outline-none focus-visible:bg-muted'
                                    onClick={() => setIsOpen(false)}
                                    role='menuitem'
                                    tabIndex={-1}
                                >
                                    <Shield size={16} strokeWidth={1.75} /> Admin Panel
                                </Link>
                            </>
                        )}

                        <div className='border-t border-border my-1'></div>

                        <button
                            onClick={handleLogout}
                            className='flex items-center gap-2 w-full text-left px-4 py-2 text-sm hover:bg-muted transition-colors text-error-400 cursor-pointer focus-visible:outline-none focus-visible:bg-muted'
                            role='menuitem'
                            tabIndex={-1}
                        >
                            <LogOut size={16} strokeWidth={1.75} /> Logout
                        </button>
                    </div>
                </div>
            )}
        </div>
    );
}
