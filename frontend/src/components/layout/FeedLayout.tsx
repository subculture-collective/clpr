import React from 'react';
import { cn } from '@/lib/utils';
import { Container } from './Container';

interface FeedLayoutProps {
    children: React.ReactNode;
    sidebar?: React.ReactNode;
    className?: string;
}

export function FeedLayout({ children, sidebar, className }: FeedLayoutProps) {
    return (
        <Container className={cn('px-0 xs:px-0 sm:px-0 py-3 md:px-4 md:py-8', className)}>
            <div className='grid grid-cols-1 lg:grid-cols-[1fr_300px] gap-6 lg:gap-8'>
                {/* Main content */}
                <div className='min-w-0'>{children}</div>

                {/* Sidebar — hidden on mobile, sticky on desktop */}
                {sidebar && (
                    <aside className='hidden lg:block'>
                        <div className='sticky top-[72px] pt-4 max-h-[calc(100vh-88px)] overflow-y-auto space-y-4 scrollbar-thin'>
                            {sidebar}
                        </div>
                    </aside>
                )}
            </div>
        </Container>
    );
}
