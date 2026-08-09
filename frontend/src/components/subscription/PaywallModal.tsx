import React, { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import { useAuth } from '../../hooks/useAuth';
import { createCheckoutSession } from '../../lib/subscription-api';
import {
  trackPaywallView,
  trackUpgradeClick,
  trackCheckoutInitiated,
  trackPaywallDismissed,
  trackBillingPeriodChange,
} from '../../lib/paywall-analytics';
import { PRICING, PRO_FEATURES, calculateYearlyMonthlyPrice, calculateSavingsPercent } from '../../constants/pricing';
import { Modal } from '../ui/Modal';

export interface PaywallModalProps {
  /** Whether the modal is open */
  isOpen: boolean;
  /** Callback to close the modal */
  onClose: () => void;
  /** Feature name that triggered the paywall */
  featureName?: string;
  /** Optional custom title */
  title?: string;
  /** Optional custom description */
  description?: string;
  /** Callback when upgrade is initiated */
  onUpgradeClick?: () => void;
  /** Checkout price IDs; injectable for tests and runtime configuration. */
  priceIds?: Partial<Record<'monthly' | 'yearly', string>>;
  /** Redirect boundary used after checkout creation. */
  onCheckoutRedirect?: (url: string) => void;
}

const DEFAULT_PRICE_IDS = {
  monthly: import.meta.env.VITE_STRIPE_PRO_MONTHLY_PRICE_ID || '',
  yearly: import.meta.env.VITE_STRIPE_PRO_YEARLY_PRICE_ID || '',
};

/**
 * Modal component that displays paywall with plan comparison and upgrade options
 *
 * @example
 * ```tsx
 * const [showPaywall, setShowPaywall] = useState(false);
 *
 * <PaywallModal
 *   isOpen={showPaywall}
 *   onClose={() => setShowPaywall(false)}
 *   featureName="Collections"
 * />
 * ```
 */
export function PaywallModal({
  isOpen,
  onClose,
  featureName,
  title,
  description,
  onUpgradeClick,
  priceIds,
  onCheckoutRedirect = url => {
    window.location.href = url;
  },
}: PaywallModalProps): React.ReactElement | null {
  const { user } = useAuth();
  const navigate = useNavigate();
  const [isLoading, setIsLoading] = useState<string | null>(null);
  const [billingPeriod, setBillingPeriod] = useState<'monthly' | 'yearly'>('yearly');

  // Track paywall view when modal opens
  useEffect(() => {
    if (isOpen) {
      trackPaywallView({
        feature: featureName,
        userId: user?.id,
      });
    }
  }, [isOpen, featureName, user?.id]);

  if (!isOpen) return null;

  const handleSubscribe = async (period: 'monthly' | 'yearly') => {
    if (!user) {
      onClose();
      navigate('/login?redirect=/pricing');
      return;
    }

    // Track upgrade button click
    trackUpgradeClick({
      feature: featureName,
      billingPeriod: period,
      userId: user.id,
    });

    if (onUpgradeClick) {
      onUpgradeClick();
    }

    setIsLoading(period);

    try {
      const priceId = priceIds?.[period] ?? DEFAULT_PRICE_IDS[period];
      if (!priceId) {
        setIsLoading(null);
        alert('Subscription not configured. Please contact support.');
        return;
      }

      // Track checkout initiation
      trackCheckoutInitiated({
        feature: featureName,
        billingPeriod: period,
        userId: user.id,
      });

      const response = await createCheckoutSession(priceId);
      onCheckoutRedirect(response.session_url);
    } catch (error) {
      console.error('Failed to create checkout session:', error);
      alert('Failed to start checkout. Please try again.');
      setIsLoading(null);
    }
  };

  const handleViewPricing = () => {
    trackPaywallDismissed({
      feature: featureName,
      userId: user?.id,
      metadata: { action: 'view_pricing' },
    });
    onClose();
    navigate('/pricing');
  };

  const handleClose = () => {
    trackPaywallDismissed({
      feature: featureName,
      userId: user?.id,
      metadata: { action: 'close_button' },
    });
    onClose();
  };

  const handleBillingPeriodChange = (period: 'monthly' | 'yearly') => {
    trackBillingPeriodChange({
      feature: featureName,
      billingPeriod: period,
      userId: user?.id,
    });
    setBillingPeriod(period);
  };

  const monthlyPrice = PRICING.monthly;
  const yearlyPrice = PRICING.yearly;
  const yearlyMonthlyPrice = calculateYearlyMonthlyPrice(yearlyPrice);
  const savingsPercent = calculateSavingsPercent(monthlyPrice, yearlyPrice);

  const modalTitle = title || `${featureName ? `${featureName} is` : 'This feature is'} a Pro Feature`;
  const modalDescription = description || 'Upgrade to Clipper Pro to unlock this feature and many more.';

  return (
    <Modal
      open={isOpen}
      onClose={handleClose}
      title={modalTitle}
      size="2xl"
      className="bg-gray-900 border border-gray-800"
    >
          {/* Header illustration and description */}
          <div className="text-center pt-8 pb-6 px-6 border-b border-gray-800">
            <div className="mb-4">
              <svg
                className="w-16 h-16 mx-auto text-purple-500"
                fill="none"
                stroke="currentColor"
                viewBox="0 0 24 24"
              >
                <path
                  strokeLinecap="round"
                  strokeLinejoin="round"
                  strokeWidth={2}
                  d="M5 3v4M3 5h4M6 17v4m-2-2h4m5-16l2.286 6.857L21 12l-5.714 2.143L13 21l-2.286-6.857L5 12l5.714-2.143L13 3z"
                />
              </svg>
            </div>
            <p className="text-gray-400 text-lg">{modalDescription}</p>
          </div>

          <div className="p-6 space-y-6">
            {/* Billing toggle */}
            <div className="flex justify-center">
              <div className="bg-gray-800 rounded-lg p-1 inline-flex">
                <button
                  type="button"
                  onClick={() => handleBillingPeriodChange('monthly')}
                  aria-pressed={billingPeriod === 'monthly'}
                  className={`min-h-[44px] px-6 py-2 rounded-md text-sm font-medium transition-colors motion-reduce:transition-none focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-purple-400 ${
                    billingPeriod === 'monthly'
                      ? 'bg-purple-600 text-white'
                      : 'text-gray-400 hover:text-white'
                  }`}
                >
                  Monthly
                </button>
                <button
                  type="button"
                  onClick={() => handleBillingPeriodChange('yearly')}
                  aria-pressed={billingPeriod === 'yearly'}
                  className={`min-h-[44px] px-6 py-2 rounded-md text-sm font-medium transition-colors motion-reduce:transition-none focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-purple-400 ${
                    billingPeriod === 'yearly'
                      ? 'bg-purple-600 text-white'
                      : 'text-gray-400 hover:text-white'
                  }`}
                >
                  Yearly
                  <span className="ml-2 text-xs bg-green-500 text-white px-2 py-0.5 rounded">
                    Save {savingsPercent}%
                  </span>
                </button>
              </div>
            </div>

            {/* Plan comparison */}
            <div className="grid md:grid-cols-2 gap-4">
              {/* Current (Free) Plan */}
              <div className="bg-gray-800 rounded-lg p-6 border border-gray-700">
                <div className="mb-4">
                  <h3 className="text-xl font-bold text-white mb-1">Free</h3>
                  <p className="text-gray-400 text-sm">Your current plan</p>
                </div>
                <div className="mb-6">
                  <span className="text-3xl font-bold text-white">$0</span>
                  <span className="text-gray-400">/month</span>
                </div>
                <div className="text-sm text-gray-400 space-y-2">
                  <div>✓ Browse all clips</div>
                  <div>✓ Basic search</div>
                  <div>✓ Vote & comment</div>
                  <div>✓ 50 favorites</div>
                  <div>✗ Advanced features</div>
                </div>
              </div>

              {/* Pro Plan */}
              <div className="bg-gradient-to-br from-purple-600 to-indigo-600 rounded-lg p-6 border-2 border-purple-400 relative">
                <div className="absolute top-0 right-0 bg-yellow-400 text-gray-900 text-xs font-bold px-3 py-1 rounded-bl-lg rounded-tr-lg">
                  RECOMMENDED
                </div>
                <div className="mb-4">
                  <h3 className="text-xl font-bold text-white mb-1">Pro</h3>
                  <p className="text-purple-100 text-sm">Everything you need</p>
                </div>
                <div className="mb-6">
                  {billingPeriod === 'monthly' ? (
                    <>
                      <span className="text-3xl font-bold text-white">${monthlyPrice}</span>
                      <span className="text-purple-100">/month</span>
                    </>
                  ) : (
                    <>
                      <span className="text-3xl font-bold text-white">${yearlyMonthlyPrice}</span>
                      <span className="text-purple-100">/month</span>
                      <div className="text-sm text-purple-100 mt-1">
                        Billed ${yearlyPrice}/year
                      </div>
                    </>
                  )}
                </div>
                <button
                  type="button"
                  onClick={() => handleSubscribe(billingPeriod)}
                  disabled={isLoading !== null}
                  aria-busy={isLoading !== null || undefined}
                  className="w-full min-h-[44px] py-3 px-6 rounded-md bg-white text-purple-600 font-bold hover:bg-gray-100 transition-colors motion-reduce:transition-none focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-white focus-visible:ring-offset-2 disabled:opacity-50 disabled:cursor-not-allowed mb-4"
                >
                  {isLoading === billingPeriod ? 'Processing...' : 'Upgrade to Pro'}
                </button>
              </div>
            </div>

            {/* Features grid */}
            <div>
              <h4 className="text-lg font-semibold text-white mb-4 text-center">
                Everything in Pro
              </h4>
              <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
                {PRO_FEATURES.map((feature, index) => (
                  <div
                    key={index}
                    className="flex items-center gap-2 text-sm text-gray-300"
                  >
                    <span className="text-xl">{feature.icon}</span>
                    <span>{feature.text}</span>
                  </div>
                ))}
              </div>
            </div>

            {/* Footer */}
            <div className="text-center pt-4 border-t border-gray-800">
              <p className="text-sm text-gray-400 mb-2">
                Cancel anytime • No hidden fees • Secure payment with Stripe
              </p>
              <button
                type="button"
                onClick={handleViewPricing}
                className="min-h-[44px] px-3 text-sm text-purple-400 hover:text-purple-300 transition-colors motion-reduce:transition-none focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-purple-400 rounded"
              >
                View full pricing details →
              </button>
            </div>
          </div>
    </Modal>
  );
}
