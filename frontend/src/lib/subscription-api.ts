import { apiClient } from './api';

export interface Subscription {
  id: string;
  user_id: string;
  stripe_customer_id: string;
  stripe_subscription_id?: string;
  stripe_price_id?: string;
  status: 'inactive' | 'active' | 'trialing' | 'past_due' | 'canceled' | 'unpaid';
  tier: 'free' | 'pro';
  current_period_start?: string;
  current_period_end?: string;
  cancel_at_period_end: boolean;
  canceled_at?: string;
  trial_start?: string;
  trial_end?: string;
  grace_period_end?: string;
  created_at: string;
  updated_at: string;
}

export interface CreateCheckoutSessionRequest {
  price_id: string;
}

export interface CreateCheckoutSessionResponse {
  session_id: string;
  session_url: string;
}

export interface CreatePortalSessionResponse {
  portal_url: string;
}

export interface Invoice {
  id: string;
  number?: string;
  status: string;
  amount_due: number;
  amount_paid: number;
  currency: string;
  created: number;
  period_start: number;
  period_end: number;
  invoice_pdf?: string;
  hosted_invoice_url?: string;
}

/**
 * Get the current user's subscription
 */
export async function getSubscription(): Promise<Subscription | null> {
  try {
    const response = await apiClient.get('/subscriptions/me');
    return response.data;
  } catch (error: unknown) {
    if (error && typeof error === 'object' && 'response' in error) {
      const axiosError = error as { response?: { status?: number } };
      if (axiosError.response?.status === 404) {
        return null; // No subscription found
      }
    }
    throw error;
  }
}

/**
 * Create a Stripe Checkout session for subscription
 */
export async function createCheckoutSession(
  priceId: string
): Promise<CreateCheckoutSessionResponse> {
  const response = await apiClient.post(
    '/subscriptions/checkout',
    { price_id: priceId } as CreateCheckoutSessionRequest
  );
  return response.data;
}

/**
 * Create a Stripe Customer Portal session
 */
export async function createPortalSession(): Promise<CreatePortalSessionResponse> {
  const response = await apiClient.post('/subscriptions/portal', {});
  return response.data;
}

/**
 * Cancel a subscription
 */
export async function cancelSubscription(immediate: boolean = false): Promise<void> {
  await apiClient.post('/subscriptions/cancel', { immediate });
}

/**
 * Reactivate a subscription scheduled for cancellation
 */
export async function reactivateSubscription(): Promise<void> {
  await apiClient.post('/subscriptions/reactivate', {});
}

/**
 * Get user's invoices
 */
export async function getInvoices(limit: number = 10): Promise<Invoice[]> {
  const response = await apiClient.get('/subscriptions/invoices', {
    params: { limit },
  });
  return response.data;
}

/**
 * Change subscription plan
 */
export async function changeSubscriptionPlan(priceId: string): Promise<void> {
  await apiClient.post('/subscriptions/change-plan', { price_id: priceId });
}

/**
 * Check if user has an active subscription
 */
export function hasActiveSubscription(subscription: Subscription | null): boolean {
  if (subscription === null) return false;
  if (subscription.status === 'active' || subscription.status === 'trialing') return true;
  return (subscription.status === 'past_due' || subscription.status === 'unpaid') &&
    !!subscription.grace_period_end && new Date(subscription.grace_period_end).getTime() > Date.now();
}

/**
 * Check if user is a Pro subscriber
 */
export function isProUser(subscription: Subscription | null): boolean {
  return (
    hasActiveSubscription(subscription) &&
    subscription?.tier === 'pro'
  );
}
