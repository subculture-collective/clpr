import { useEffect, useRef } from 'react';
import { useNavigate, Link } from 'react-router-dom';
import { trackCheckoutCancelled } from '../lib/paywall-analytics';
import { useAuth } from '../hooks/useAuth';

export default function SubscriptionCancelPage() {
    const navigate = useNavigate();
    const { user } = useAuth();
    const hasTrackedCancellation = useRef(false);

    useEffect(() => {
        // Track checkout cancellation only once
        if (!hasTrackedCancellation.current) {
            trackCheckoutCancelled({
                userId: user?.id,
            });
            hasTrackedCancellation.current = true;
        }
    }, [user?.id]);

    return (
        <div className='min-h-screen bg-background flex items-center justify-center px-4'>
            <div className='max-w-md w-full text-center'>
                <div className='mb-8'>
                    <svg
                        className='h-20 w-20 text-muted-foreground mx-auto'
                        fill='none'
                        stroke='currentColor'
                        viewBox='0 0 24 24'
                    >
                        <path
                            strokeLinecap='round'
                            strokeLinejoin='round'
                            strokeWidth={2}
                            d='M10 14l2-2m0 0l2-2m-2 2l-2-2m2 2l2 2m7-2a9 9 0 11-18 0 9 9 0 0118 0z'
                        />
                    </svg>
                </div>

                <h1 className='text-3xl font-bold text-white mb-4'>
                    Subscription Canceled
                </h1>

                <p className='text-muted-foreground mb-8'>
                    You've canceled the checkout process. No charges have been
                    made to your account. You can subscribe anytime you're
                    ready.
                </p>

                <div className='space-y-4'>
                    <button
                        onClick={() => navigate('/pricing')}
                        className='w-full py-3 px-6 rounded-md bg-purple-600 text-white font-medium hover:bg-purple-700 transition-colors'
                    >
                        View Plans Again
                    </button>

                    <button
                        onClick={() => navigate('/')}
                        className='w-full py-3 px-6 rounded-md bg-surface text-foreground font-medium hover:bg-surface-hover transition-colors'
                    >
                        Back to Home
                    </button>
                </div>

                <p className='text-sm text-muted-foreground mt-8'>
                    Have questions?{' '}
                    <Link
                        to='/support'
                        className='text-purple-400 hover:text-purple-300'
                    >
                        Contact our support team
                    </Link>
                </p>
            </div>
        </div>
    );
}
