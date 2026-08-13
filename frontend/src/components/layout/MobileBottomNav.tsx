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
      className='fixed inset-x-0 bottom-0 z-40 grid grid-cols-4 border-t border-border bg-background/95 px-2 pb-[env(safe-area-inset-bottom)] backdrop-blur-xl md:hidden'
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
              active ? 'text-primary-400' : 'text-muted-foreground hover:text-foreground',
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
