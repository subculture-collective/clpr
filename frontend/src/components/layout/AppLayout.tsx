import { lazy, Suspense, useEffect } from 'react';
import { Outlet, useLocation, useNavigationType } from 'react-router-dom';
import { SkipLink } from '../ui/SkipLink';
import { Footer } from './Footer';
import { Header } from './Header';
import { CategoriesNav } from './CategoriesNav';
import { OfflineIndicator } from '../OfflineIndicator';
import { useOfflineCacheInit } from '@/hooks/useOfflineCache';
import { useSyncManager } from '@/hooks/useSyncManager';
import { useAuth } from '@/context/AuthContext';

const QueueWidget = lazy(() =>
    import('../queue/QueueWidget').then((module) => ({ default: module.QueueWidget })),
);

const scrollPositions = new Map<string, number>();

export function AppLayout() {
    const location = useLocation();
    const navigationType = useNavigationType();
    const { isAuthenticated } = useAuth();

    // Initialize offline cache on app start
    useOfflineCacheInit();

    // Initialize sync manager
    useSyncManager();

    // New routes start at the top; browser back/forward restores the prior position.
    useEffect(() => {
        document.body.style.overflow = '';
        const targetPosition = navigationType === 'POP'
            ? scrollPositions.get(location.key) ?? 0
            : 0;
        try {
            if (typeof window.scrollTo === 'function') {
                window.scrollTo(0, targetPosition);
            }
        } catch {
            // no-op in environments where scrollTo is not implemented
        }

        return () => {
            scrollPositions.set(location.key, window.scrollY);
        };
    }, [location.key, navigationType]);

    return (
        <div className='min-h-screen flex flex-col bg-background text-foreground transition-theme'>
            <SkipLink targetId='main-content' label='Skip to main content' />
            <Header />
            <CategoriesNav />
            <main id='main-content' className='flex-1' tabIndex={-1}>
                <Outlet />
            </main>
            <Footer />
            <OfflineIndicator />
            {isAuthenticated && (
                <Suspense fallback={null}>
                    <QueueWidget />
                </Suspense>
            )}
        </div>
    );
}
