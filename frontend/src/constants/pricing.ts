/**
 * Pricing constants for Clipper Pro subscription
 */

export const PRICING = {
  monthly: 9.99,
  yearly: 99.99,
} as const;

/**
 * Calculate effective monthly price for yearly plan
 */
export const calculateYearlyMonthlyPrice = (yearlyPrice: number): string =>
  (yearlyPrice / 12).toFixed(2);

/**
 * Calculate savings percentage for yearly plan
 */
export const calculateSavingsPercent = (monthlyPrice: number, yearlyPrice: number): number =>
  Math.round(((monthlyPrice * 12 - yearlyPrice) / (monthlyPrice * 12)) * 100);

/**
 * Pro feature list with icons
 */
export const PRO_FEATURES = [
  { icon: 'ban', text: 'Ad-free browsing' },
  { icon: 'star', text: 'Unlimited favorites' },
  { icon: 'folder', text: 'Custom collections' },
  { icon: 'search', text: 'Advanced search & filters' },
  { icon: 'refresh-cw', text: 'Cross-device sync' },
  { icon: 'bar-chart-3', text: 'Export your data' },
  { icon: 'target', text: '5x higher rate limits' },
  { icon: 'mail', text: 'Priority support' },
] as const;

/**
 * Pro feature list for pricing pages (longer descriptions)
 */
export const PRO_FEATURES_DETAILED = [
  'Ad-free browsing experience',
  'Advanced search and filtering',
  'Favorite clips sync across devices',
  'Priority support',
  'Early access to new features',
  'Custom collections and playlists',
  'Export your data',
  '5x higher rate limits',
] as const;
