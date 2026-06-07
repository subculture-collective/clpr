import { describe, expect, it, beforeEach, afterEach, vi } from 'vitest';
import {
  isValidDeepLink,
  parseDeepLink,
  handleDeepLink,
  generateDeepLink,
  isOpenedViaDeepLink,
  getShareTargetData,
  DEEP_LINK_ROUTES,
} from './deep-linking';

describe('Deep Linking Utilities', () => {
  const baseUrl = 'https://clpr.example.com';

  describe('isValidDeepLink', () => {
    it('should return true for valid clip detail link', () => {
      expect(isValidDeepLink(`${baseUrl}/clip/abc123`)).toBe(true);
    });

    it('should return true for valid profile link', () => {
      expect(isValidDeepLink(`${baseUrl}/profile`)).toBe(true);
    });

    it('should return true for valid search link', () => {
      expect(isValidDeepLink(`${baseUrl}/search`)).toBe(true);
    });

    it('should return true for valid submit link', () => {
      expect(isValidDeepLink(`${baseUrl}/submit`)).toBe(true);
    });

    it('should return true for valid game link', () => {
      expect(isValidDeepLink(`${baseUrl}/game/valorant`)).toBe(true);
    });

    it('should return true for valid creator link', () => {
      expect(isValidDeepLink(`${baseUrl}/creator/shroud`)).toBe(true);
    });

    it('should return true for valid creator analytics link', () => {
      expect(isValidDeepLink(`${baseUrl}/creator/shroud/analytics`)).toBe(true);
    });

    it('should return true for valid tag link', () => {
      expect(isValidDeepLink(`${baseUrl}/tag/funny`)).toBe(true);
    });

    it('should return true for valid feed links', () => {
      expect(isValidDeepLink(`${baseUrl}/discover`)).toBe(true);
      expect(isValidDeepLink(`${baseUrl}/new`)).toBe(true);
      expect(isValidDeepLink(`${baseUrl}/top`)).toBe(true);
      expect(isValidDeepLink(`${baseUrl}/rising`)).toBe(true);
    });

    it('should return false for invalid links', () => {
      expect(isValidDeepLink(`${baseUrl}/invalid-route`)).toBe(false);
      expect(isValidDeepLink(`${baseUrl}/admin/dashboard`)).toBe(false);
    });

    it('should return false for malformed URLs', () => {
      expect(isValidDeepLink('not-a-url')).toBe(false);
      expect(isValidDeepLink('')).toBe(false);
    });
  });

  describe('parseDeepLink', () => {
    it('should parse clip detail link correctly', () => {
      expect(parseDeepLink(`${baseUrl}/clip/abc123`)).toBe('/clip/abc123');
    });

    it('should parse profile link correctly', () => {
      expect(parseDeepLink(`${baseUrl}/profile`)).toBe('/profile');
    });

    it('should parse search link with query params', () => {
      expect(parseDeepLink(`${baseUrl}/search?q=valorant`)).toBe('/search?q=valorant');
    });

    it('should parse game link correctly', () => {
      expect(parseDeepLink(`${baseUrl}/game/valorant`)).toBe('/game/valorant');
    });

    it('should parse creator link correctly', () => {
      expect(parseDeepLink(`${baseUrl}/creator/shroud`)).toBe('/creator/shroud');
    });

    it('should parse tag link correctly', () => {
      expect(parseDeepLink(`${baseUrl}/tag/funny`)).toBe('/tag/funny');
    });

    it('should return null for invalid links', () => {
      expect(parseDeepLink(`${baseUrl}/invalid-route`)).toBeNull();
    });

    it('should return null for malformed URLs', () => {
      expect(parseDeepLink('not-a-url')).toBeNull();
    });
  });

  describe('handleDeepLink', () => {
    it('should handle valid deep links', () => {
      expect(handleDeepLink(`${baseUrl}/clip/abc123`)).toBe('/clip/abc123');
      expect(handleDeepLink(`${baseUrl}/profile`)).toBe('/profile');
      expect(handleDeepLink(`${baseUrl}/search`)).toBe('/search');
    });

    it('should return null for invalid deep links', () => {
      expect(handleDeepLink(`${baseUrl}/invalid`)).toBeNull();
      expect(handleDeepLink('not-a-url')).toBeNull();
    });
  });

  describe('generateDeepLink', () => {
    beforeEach(() => {
      // Mock window.location.origin
      vi.stubGlobal('location', { origin: baseUrl });
    });

    afterEach(() => {
      vi.unstubAllGlobals();
    });

    it('should generate deep link from path', () => {
      expect(generateDeepLink('/clip/abc123')).toBe(`${baseUrl}/clip/abc123`);
      expect(generateDeepLink('/profile')).toBe(`${baseUrl}/profile`);
    });

    it('should use custom base URL when provided', () => {
      const customBase = 'https://custom.example.com';
      expect(generateDeepLink('/clip/abc123', customBase)).toBe(`${customBase}/clip/abc123`);
    });
  });

  describe('isOpenedViaDeepLink', () => {
    beforeEach(() => {
      // Reset window.location
      vi.stubGlobal('location', { 
        search: '', 
        origin: baseUrl 
      });
      // Reset document.referrer
      Object.defineProperty(document, 'referrer', {
        value: '',
        writable: true,
        configurable: true,
      });
    });

    afterEach(() => {
      vi.unstubAllGlobals();
    });

    it('should return true when opened via share target on root path', () => {
      vi.stubGlobal('location', { 
        search: '?url=https://example.com',
        pathname: '/',
        origin: baseUrl 
      });
      expect(isOpenedViaDeepLink()).toBe(true);
    });

    it('should return true when opened via share target on submit path', () => {
      vi.stubGlobal('location', { 
        search: '?url=https://example.com',
        pathname: '/submit',
        origin: baseUrl 
      });
      expect(isOpenedViaDeepLink()).toBe(true);
    });

    it('should return false when share params on non-share-target path', () => {
      vi.stubGlobal('location', { 
        search: '?url=https://example.com',
        pathname: '/search',
        origin: baseUrl 
      });
      expect(isOpenedViaDeepLink()).toBe(false);
    });

    it('should return true when referrer is external', () => {
      Object.defineProperty(document, 'referrer', {
        value: 'https://external.com',
        writable: true,
        configurable: true,
      });
      expect(isOpenedViaDeepLink()).toBe(true);
    });

    it('should return false when opened normally', () => {
      expect(isOpenedViaDeepLink()).toBe(false);
    });
  });

  describe('getShareTargetData', () => {
    beforeEach(() => {
      vi.stubGlobal('location', { search: '' });
    });

    afterEach(() => {
      vi.unstubAllGlobals();
    });

    it('should extract share target data from URL params', () => {
      vi.stubGlobal('location', { 
        search: '?url=https://twitch.tv/clip&title=Awesome%20Clip&text=Check%20this%20out' 
      });
      
      const data = getShareTargetData();
      expect(data).toEqual({
        url: 'https://twitch.tv/clip',
        title: 'Awesome Clip',
        text: 'Check this out',
      });
    });

    it('should handle partial share target data', () => {
      vi.stubGlobal('location', { 
        search: '?url=https://twitch.tv/clip' 
      });
      
      const data = getShareTargetData();
      expect(data).toEqual({
        url: 'https://twitch.tv/clip',
        title: undefined,
        text: undefined,
      });
    });

    it('should return null when no share target data', () => {
      expect(getShareTargetData()).toBeNull();
    });
  });

  describe('DEEP_LINK_ROUTES', () => {
    it('should have all expected routes', () => {
      const descriptions = DEEP_LINK_ROUTES.map(route => route.description);
      
      expect(descriptions).toContain('Clip detail page');
      expect(descriptions).toContain('User profile page');
      expect(descriptions).toContain('Search page');
      expect(descriptions).toContain('Submit clip page');
      expect(descriptions).toContain('Game page');
      expect(descriptions).toContain('Creator page');
      expect(descriptions).toContain('Tag page');
    });

    it('should have valid regex patterns', () => {
      // Test that each pattern can match and extract data correctly
      const testCases = [
        { path: '/clip/abc123', description: 'Clip detail page' },
        { path: '/profile', description: 'User profile page' },
        { path: '/game/valorant', description: 'Game page' },
        { path: '/creator/shroud', description: 'Creator page' },
        { path: '/tag/funny', description: 'Tag page' },
      ];

      testCases.forEach(({ path, description }) => {
        const route = DEEP_LINK_ROUTES.find(r => r.description === description);
        expect(route).toBeDefined();
        expect(route!.pattern.test(path)).toBe(true);
      });
    });
  });
});
