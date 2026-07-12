import { describe, expect, it, vi } from 'vitest';
import { hasActiveSubscription, isProUser, type Subscription } from './subscription-api';

const base: Subscription = {
    id: 'sub-1',
    user_id: 'user-1',
    stripe_customer_id: 'cus-1',
    status: 'active',
    tier: 'pro',
    cancel_at_period_end: false,
    created_at: '2026-01-01T00:00:00Z',
    updated_at: '2026-01-01T00:00:00Z',
};

describe('subscription entitlement policy', () => {
    it('matches backend grace-period entitlement and expiry', () => {
        vi.useFakeTimers();
        vi.setSystemTime(new Date('2026-07-12T12:00:00Z'));

        const inGrace = { ...base, status: 'past_due' as const, grace_period_end: '2026-07-13T12:00:00Z' };
        const expired = { ...inGrace, grace_period_end: '2026-07-11T12:00:00Z' };
        expect(hasActiveSubscription(inGrace)).toBe(true);
        expect(isProUser(inGrace)).toBe(true);
        expect(hasActiveSubscription(expired)).toBe(false);
        expect(isProUser(expired)).toBe(false);

        vi.useRealTimers();
    });

    it('never grants Pro to a free tier even with active status', () => {
        expect(isProUser({ ...base, tier: 'free' })).toBe(false);
    });
});
