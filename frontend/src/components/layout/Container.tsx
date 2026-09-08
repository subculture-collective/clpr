import React from 'react';
import { cn } from '@/lib/utils';

export interface ContainerProps extends React.HTMLAttributes<HTMLDivElement> {
    /**
     * Maximum width of the container
     * @default 'default' (1440px)
     */
    maxWidth?: 'sm' | 'md' | 'lg' | 'xl' | '2xl' | 'full' | 'default';
    /**
     * Whether to center the container
     * @default true
     */
    center?: boolean;
    children: React.ReactNode;
}

const maxWidthClasses = {
    sm: 'max-w-[640px]',
    md: 'max-w-[768px]',
    lg: 'max-w-[1024px]',
    xl: 'max-w-[1280px]',
    '2xl': 'max-w-[1536px]',
    full: 'max-w-full',
    default: 'max-w-[1440px]',
};

/**
 * Container component provides consistent max-width wrapper with responsive padding
 */
export const Container = React.forwardRef<HTMLDivElement, ContainerProps>(
    (
        { className, maxWidth = 'default', center = true, children, ...props },
        ref,
    ) => {
        return (
            <div
                ref={ref}
                className={cn(
                    'w-full min-w-0 px-4 sm:px-6 lg:px-8',
                    maxWidthClasses[maxWidth],
                    center && 'mx-auto',
                    className,
                )}
                {...props}
            >
                {children}
            </div>
        );
    },
);

Container.displayName = 'Container';
