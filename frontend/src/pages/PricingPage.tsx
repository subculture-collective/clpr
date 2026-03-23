import { useState, useEffect } from 'react';
import { useNavigate, Link } from 'react-router-dom';
import { useAuth } from '../context/AuthContext';
import { createCheckoutSession } from '../lib/subscription-api';
import { SEO } from '../components';
import { trackEvent, PremiumEvents } from '../lib/telemetry';
import {
    trackPricingPageView,
    trackBillingPeriodChange,
    trackUpgradeClick,
    trackCheckoutInitiated,
} from '../lib/paywall-analytics';
import {
    PRICING,
    PRO_FEATURES_DETAILED,
    calculateYearlyMonthlyPrice,
    calculateSavingsPercent,
} from '../constants/pricing';

const PRICE_IDS = {
    monthly: import.meta.env.VITE_STRIPE_PRO_MONTHLY_PRICE_ID || '',
    yearly: import.meta.env.VITE_STRIPE_PRO_YEARLY_PRICE_ID || '',
};

export default function PricingPage() {
    const { user } = useAuth();
    const navigate = useNavigate();
    const [isLoading, setIsLoading] = useState<string | null>(null);
    const [billingPeriod, setBillingPeriod] = useState<'monthly' | 'yearly'>(
        'monthly'
    );

    // Track pricing page view
    useEffect(() => {
        trackEvent(PremiumEvents.PRICING_PAGE_VIEWED, {
            source: 'pricing_page',
            user_authenticated: !!user,
            user_id: user?.id,
        });
        // Keep legacy tracking for backwards compatibility
        trackPricingPageView({
            userId: user?.id,
        });
    }, [user]);

    const handleBillingPeriodChange = (period: 'monthly' | 'yearly') => {
        trackEvent(PremiumEvents.PRICING_TIER_CLICKED, {
            billing_period: period,
        });
        // Keep legacy tracking for backwards compatibility
        trackBillingPeriodChange({
            billingPeriod: period,
            userId: user?.id,
            metadata: { source: 'pricing_page' },
        });
        setBillingPeriod(period);
    };

    const handleSubscribe = async (period: 'monthly' | 'yearly') => {
        if (!user) {
            // Redirect to login
            navigate('/login?redirect=/pricing');
            return;
        }

        // Track upgrade click
        trackUpgradeClick({
            billingPeriod: period,
            userId: user.id,
            metadata: { source: 'pricing_page' },
        });

        setIsLoading(period);

        try {
            const priceId = PRICE_IDS[period];
            if (!priceId) {
                alert('Subscription not configured. Please contact support.');
                return;
            }

            // Track checkout initiation
            trackCheckoutInitiated({
                billingPeriod: period,
                userId: user.id,
                metadata: { source: 'pricing_page' },
            });

            const response = await createCheckoutSession(priceId);

            // Redirect to Stripe Checkout
            window.location.href = response.session_url;
        } catch (error) {
            console.error('Failed to create checkout session:', error);
            alert('Failed to start checkout. Please try again.');
        } finally {
            setIsLoading(null);
        }
    };

    const monthlyPrice = PRICING.monthly;
    const yearlyPrice = PRICING.yearly;
    const yearlyMonthlyPrice = calculateYearlyMonthlyPrice(yearlyPrice);
    const savingsPercent = calculateSavingsPercent(monthlyPrice, yearlyPrice);

    return (
        <>
            <SEO
                title='Pricing - clpr Pro'
                description='Upgrade to clpr Pro for an ad-free experience, advanced search and filtering, priority support, and exclusive features. Choose monthly or yearly billing.'
                canonicalUrl='/pricing'
            />
            <div className='min-h-screen bg-background py-12 px-4 sm:px-6 lg:px-8'>
                <div className='max-w-7xl mx-auto'>
                    {/* Header */}
                    <div className='text-center mb-12'>
                        <h1 className='text-4xl font-bold text-white mb-4'>
                            Upgrade to clpr Pro
                        </h1>
                        <p className='text-xl text-muted-foreground'>
                            Get the most out of clpr with exclusive Pro features
                        </p>
                    </div>

                    {/* Billing period toggle */}
                    <div className='flex justify-center mb-12'>
                        <div className='bg-surface rounded-lg p-1 inline-flex'>
                            <button
                                onClick={() =>
                                    handleBillingPeriodChange('monthly')
                                }
                                className={`px-6 py-2 rounded-md text-sm font-medium transition-colors ${
                                    billingPeriod === 'monthly'
                                        ? 'bg-purple-600 text-white'
                                        : 'text-muted-foreground hover:text-white'
                                }`}
                            >
                                Monthly
                            </button>
                            <button
                                onClick={() =>
                                    handleBillingPeriodChange('yearly')
                                }
                                className={`px-6 py-2 rounded-md text-sm font-medium transition-colors ${
                                    billingPeriod === 'yearly'
                                        ? 'bg-purple-600 text-white'
                                        : 'text-muted-foreground hover:text-white'
                                }`}
                            >
                                Yearly
                                <span className='ml-2 text-xs bg-green-500 text-white px-2 py-0.5 rounded'>
                                    Save {savingsPercent}%
                                </span>
                            </button>
                        </div>
                    </div>

                    {/* Pricing cards */}
                    <div className='grid md:grid-cols-2 gap-8 max-w-4xl mx-auto'>
                        {/* Free Plan */}
                        <div className='bg-surface rounded-lg p-8 border border-border'>
                            <div className='mb-6'>
                                <h2 className='text-2xl font-bold text-white mb-2'>
                                    Free
                                </h2>
                                <p className='text-muted-foreground'>
                                    Perfect for casual users
                                </p>
                            </div>

                            <div className='mb-6'>
                                <span className='text-4xl font-bold text-white'>
                                    $0
                                </span>
                                <span className='text-muted-foreground'>/month</span>
                            </div>

                            <ul className='space-y-3 mb-8'>
                                <li className='flex items-start'>
                                    <svg
                                        className='h-5 w-5 text-green-500 mr-3 shrink-0 mt-0.5'
                                        fill='none'
                                        stroke='currentColor'
                                        viewBox='0 0 24 24'
                                    >
                                        <path
                                            strokeLinecap='round'
                                            strokeLinejoin='round'
                                            strokeWidth={2}
                                            d='M5 13l4 4L19 7'
                                        />
                                    </svg>
                                    <span className='text-foreground'>
                                        Browse all clips
                                    </span>
                                </li>
                                <li className='flex items-start'>
                                    <svg
                                        className='h-5 w-5 text-green-500 mr-3 shrink-0 mt-0.5'
                                        fill='none'
                                        stroke='currentColor'
                                        viewBox='0 0 24 24'
                                    >
                                        <path
                                            strokeLinecap='round'
                                            strokeLinejoin='round'
                                            strokeWidth={2}
                                            d='M5 13l4 4L19 7'
                                        />
                                    </svg>
                                    <span className='text-foreground'>
                                        Basic search
                                    </span>
                                </li>
                                <li className='flex items-start'>
                                    <svg
                                        className='h-5 w-5 text-green-500 mr-3 shrink-0 mt-0.5'
                                        fill='none'
                                        stroke='currentColor'
                                        viewBox='0 0 24 24'
                                    >
                                        <path
                                            strokeLinecap='round'
                                            strokeLinejoin='round'
                                            strokeWidth={2}
                                            d='M5 13l4 4L19 7'
                                        />
                                    </svg>
                                    <span className='text-foreground'>
                                        Vote and comment
                                    </span>
                                </li>
                                <li className='flex items-start'>
                                    <svg
                                        className='h-5 w-5 text-green-500 mr-3 shrink-0 mt-0.5'
                                        fill='none'
                                        stroke='currentColor'
                                        viewBox='0 0 24 24'
                                    >
                                        <path
                                            strokeLinecap='round'
                                            strokeLinejoin='round'
                                            strokeWidth={2}
                                            d='M5 13l4 4L19 7'
                                        />
                                    </svg>
                                    <span className='text-foreground'>
                                        Create favorites
                                    </span>
                                </li>
                            </ul>

                            <button
                                disabled
                                className='w-full py-3 px-6 rounded-md bg-surface text-muted-foreground font-medium cursor-not-allowed'
                            >
                                Current Plan
                            </button>
                        </div>

                        {/* Pro Plan */}
                        <div className='bg-linear-to-br from-purple-600 to-indigo-600 rounded-lg p-8 border-2 border-purple-400 shadow-xl relative'>
                            <div className='absolute top-0 right-0 bg-yellow-400 text-background text-xs font-bold px-3 py-1 rounded-bl-lg rounded-tr-lg'>
                                POPULAR
                            </div>

                            <div className='mb-6'>
                                <h2 className='text-2xl font-bold text-white mb-2'>
                                    Pro
                                </h2>
                                <p className='text-purple-100'>
                                    For power users and enthusiasts
                                </p>
                            </div>

                            <div className='mb-6'>
                                {billingPeriod === 'monthly' ? (
                                    <>
                                        <span className='text-4xl font-bold text-white'>
                                            ${monthlyPrice}
                                        </span>
                                        <span className='text-purple-100'>
                                            /month
                                        </span>
                                    </>
                                ) : (
                                    <>
                                        <span className='text-4xl font-bold text-white'>
                                            ${yearlyMonthlyPrice}
                                        </span>
                                        <span className='text-purple-100'>
                                            /month
                                        </span>
                                        <div className='text-sm text-purple-100 mt-1'>
                                            Billed ${yearlyPrice}/year
                                        </div>
                                    </>
                                )}
                            </div>

                            <ul className='space-y-3 mb-8'>
                                {PRO_FEATURES_DETAILED.map((feature, index) => (
                                    <li
                                        key={index}
                                        className='flex items-start'
                                    >
                                        <svg
                                            className='h-5 w-5 text-yellow-400 mr-3 shrink-0 mt-0.5'
                                            fill='none'
                                            stroke='currentColor'
                                            viewBox='0 0 24 24'
                                        >
                                            <path
                                                strokeLinecap='round'
                                                strokeLinejoin='round'
                                                strokeWidth={2}
                                                d='M5 13l4 4L19 7'
                                            />
                                        </svg>
                                        <span className='text-white'>
                                            {feature}
                                        </span>
                                    </li>
                                ))}
                            </ul>

                            <button
                                onClick={() => handleSubscribe(billingPeriod)}
                                disabled={isLoading !== null}
                                className='w-full py-3 px-6 rounded-md bg-white text-purple-600 font-bold hover:bg-surface-raised transition-colors disabled:opacity-50 disabled:cursor-not-allowed'
                            >
                                {isLoading === billingPeriod
                                    ? 'Processing...'
                                    : 'Subscribe Now'}
                            </button>
                        </div>
                    </div>

                    {/* FAQ or additional info */}
                    <div className='mt-16 text-center'>
                        <p className='text-muted-foreground mb-4'>
                            Cancel anytime. No hidden fees. Secure payment with
                            Stripe.
                        </p>
                        <p className='text-sm text-muted-foreground'>
                            Have questions?{' '}
                            <Link
                                to='/support'
                                className='text-purple-400 hover:text-purple-300'
                            >
                                Contact support
                            </Link>
                        </p>
                    </div>
                </div>
            </div>
        </>
    );
}
