import { describe, it, expect, beforeEach, afterEach } from 'vitest';
import { renderHook, act } from '@testing-library/react';
import { useQualityPreference } from './useQualityPreference';

describe('useQualityPreference', () => {
  const STORAGE_KEY = 'clpr_video_quality';

  beforeEach(() => {
    // Clear localStorage before each test
    localStorage.clear();
  });

  afterEach(() => {
    // Clean up after each test
    localStorage.clear();
  });

  describe('Initial State', () => {
    it('should default to auto quality when no preference is stored', () => {
      const { result } = renderHook(() => useQualityPreference());
      
      expect(result.current.quality).toBe('auto');
    });

    it('should load stored quality preference from localStorage', () => {
      localStorage.setItem(STORAGE_KEY, '1080p');
      
      const { result } = renderHook(() => useQualityPreference());
      
      expect(result.current.quality).toBe('1080p');
    });

    it('should default to auto if stored value is invalid', () => {
      localStorage.setItem(STORAGE_KEY, 'invalid-quality');
      
      const { result } = renderHook(() => useQualityPreference());
      
      expect(result.current.quality).toBe('auto');
    });
  });

  describe('Setting Quality', () => {
    it('should update quality when setQuality is called', () => {
      const { result } = renderHook(() => useQualityPreference());
      
      act(() => {
        result.current.setQuality('720p');
      });
      
      expect(result.current.quality).toBe('720p');
    });

    it('should persist quality to localStorage', () => {
      const { result } = renderHook(() => useQualityPreference());
      
      act(() => {
        result.current.setQuality('1080p');
      });
      
      expect(localStorage.getItem(STORAGE_KEY)).toBe('1080p');
    });

    it('should update localStorage when quality changes', () => {
      const { result } = renderHook(() => useQualityPreference());
      
      act(() => {
        result.current.setQuality('720p');
      });
      expect(localStorage.getItem(STORAGE_KEY)).toBe('720p');
      
      act(() => {
        result.current.setQuality('4K');
      });
      expect(localStorage.getItem(STORAGE_KEY)).toBe('4K');
    });
  });

  describe('Valid Quality Values', () => {
    const validQualities = ['480p', '720p', '1080p', '2K', '4K', 'auto'];

    validQualities.forEach((quality) => {
      it(`should accept ${quality} as valid quality`, () => {
        localStorage.setItem(STORAGE_KEY, quality);
        
        const { result } = renderHook(() => useQualityPreference());
        
        expect(result.current.quality).toBe(quality);
      });
    });
  });

  describe('Persistence', () => {
    it('should persist quality across hook remounts', () => {
      const { result, unmount } = renderHook(() => useQualityPreference());
      
      act(() => {
        result.current.setQuality('2K');
      });
      
      unmount();
      
      const { result: result2 } = renderHook(() => useQualityPreference());
      expect(result2.current.quality).toBe('2K');
    });

    it('should restore quality from localStorage on mount', () => {
      localStorage.setItem(STORAGE_KEY, '4K');
      
      const { result } = renderHook(() => useQualityPreference());
      
      expect(result.current.quality).toBe('4K');
    });
  });
});
