import { Compass, Heart, Home, User } from 'lucide-react';
import { Link, useLocation } from 'react-router-dom';
import { useAuth } from '@/context/AuthContext';
import { cn } from '@/lib/utils';

const navItems = [
  { label: 'Feed', to: '/', icon: Home, protected: false },
  { label: 'Discover', to: '/discover', icon: Compass, protected: false },
  { label: 'Saved', to: '/favorites', icon: Heart, protected: true },
  { label: 'Profile', to: '/profile', icon: User, protected: true },
] as const;

export function MobileBottomNav() {
  const { isAuthenticated } = useAuth();
  const location = useLocation();

  return (
    <nav
      className='fixed inset-x-0 bottom-0 z-40 grid grid-cols-4 border-t border-neutral-800 bg-neutral-950 px-2 pb-[env(safe-area-inset-bottom)] md:hidden'
      aria-label='Primary mobile navigation'
    >
      {navItems.map(item => {
        const destination = item.protected && !isAuthenticated ? '/login' : item.to;
        const active = item.to === '/' ? location.pathname === '/' : location.pathname.startsWith(item.to);
        const Icon = item.icon;
        return (
          <Link
            key={item.label}
            to={destination}
            state={item.protected && !isAuthenticated ? { from: location } : undefined}
            className={cn(
              'relative flex min-h-16 flex-col items-center justify-center gap-1 rounded-lg text-[11px] font-semibold transition-colors',
              active ? 'text-violet-300' : 'text-neutral-300 hover:text-white',
            )}
            aria-current={active ? 'page' : undefined}
          >
            <Icon size={21} strokeWidth={active ? 2.25 : 1.75} aria-hidden='true' />
            <span>{item.label}</span>
          </Link>
        );
      })}
    </nav>
  );
}
