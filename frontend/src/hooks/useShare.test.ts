import { renderHook, act } from '@testing-library/react';
import { describe, it, expect, vi, beforeEach } from 'vitest';
import { useShare } from './useShare';
import { ToastProvider } from '@/context/ToastContext';
import { allowTestConsole } from '@/test/setup';

// Mock navigator.share and navigator.clipboard
const mockShare = vi.fn();
const mockCanShare = vi.fn();
const mockWriteText = vi.fn();

describe('useShare', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    
    // Reset navigator mocks
    Object.defineProperty(navigator, 'share', {
      writable: true,
      configurable: true,
      value: mockShare,
    });
    
    Object.defineProperty(navigator, 'canShare', {
      writable: true,
      configurable: true,
      value: mockCanShare,
    });
    
    Object.defineProperty(navigator, 'clipboard', {
      writable: true,
      configurable: true,
      value: {
        writeText: mockWriteText,
      },
    });
  });

  it('should use Web Share API when available', async () => {
    mockCanShare.mockReturnValue(true);
    mockShare.mockResolvedValue(undefined);

    const { result } = renderHook(() => useShare(), {
      wrapper: ToastProvider,
    });

    const shareData = {
      title: 'Test Title',
      text: 'Test Text',
      url: 'https://example.com',
    };

    let shareResult;
    await act(async () => {
      shareResult = await result.current.share(shareData);
    });

    expect(mockCanShare).toHaveBeenCalledWith(shareData);
    expect(mockShare).toHaveBeenCalledWith(shareData);
    expect(shareResult).toEqual({ success: true, method: 'native' });
  });

  it('should fallback to clipboard when Web Share API is not available', async () => {
    mockCanShare.mockReturnValue(false);
    mockWriteText.mockResolvedValue(undefined);

    const { result } = renderHook(() => useShare(), {
      wrapper: ToastProvider,
    });

    const shareData = {
      title: 'Test Title',
      text: 'Test Text',
      url: 'https://example.com',
    };

    let shareResult;
    await act(async () => {
      shareResult = await result.current.share(shareData);
    });

    expect(mockWriteText).toHaveBeenCalledWith(shareData.url);
    expect(shareResult).toEqual({ success: true, method: 'clipboard' });
  });

  it('should handle user cancellation gracefully', async () => {
    mockCanShare.mockReturnValue(true);
    const abortError = new Error('User cancelled');
    abortError.name = 'AbortError';
    mockShare.mockRejectedValue(abortError);

    const { result } = renderHook(() => useShare(), {
      wrapper: ToastProvider,
    });

    const shareData = {
      title: 'Test Title',
      url: 'https://example.com',
    };

    let shareResult;
    await act(async () => {
      shareResult = await result.current.share(shareData);
    });

    expect(shareResult).toEqual({ success: false, cancelled: true });
  });

  it('should handle share errors', async () => {
    allowTestConsole('error', /Error sharing:.*Share failed/);
    mockCanShare.mockReturnValue(true);
    const error = new Error('Share failed');
    mockShare.mockRejectedValue(error);

    const { result } = renderHook(() => useShare(), {
      wrapper: ToastProvider,
    });

    const shareData = {
      title: 'Test Title',
      url: 'https://example.com',
    };

    let shareResult;
    await act(async () => {
      shareResult = await result.current.share(shareData);
    });

    expect(shareResult).toEqual({ success: false, error });
  });

  it('should handle clipboard write errors', async () => {
    allowTestConsole('error', /Error sharing:.*Clipboard write failed/);
    mockCanShare.mockReturnValue(false);
    const error = new Error('Clipboard write failed');
    mockWriteText.mockRejectedValue(error);

    const { result } = renderHook(() => useShare(), {
      wrapper: ToastProvider,
    });

    const shareData = {
      title: 'Test Title',
      url: 'https://example.com',
    };

    let shareResult;
    await act(async () => {
      shareResult = await result.current.share(shareData);
    });

    expect(shareResult).toEqual({ success: false, error });
  });
});
