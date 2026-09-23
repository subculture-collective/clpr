import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { ScrollToTop } from './ScrollToTop';

describe('ScrollToTop', () => {
  beforeEach(() => {
    // Reset scroll position before each test
    window.scrollY = 0;
    vi.clearAllMocks();
  });

  it('should not render when scrollY is below threshold', () => {
    window.scrollY = 400;
    const { container } = render(<ScrollToTop threshold={500} />);
    expect(container.firstChild).toBeNull();
  });

  it('should render when scrollY is above threshold', async () => {
    window.scrollY = 600;
    render(<ScrollToTop threshold={500} />);
    
    // Trigger scroll event
    fireEvent.scroll(window);
    
    await waitFor(() => {
      const button = screen.getByRole('button', { name: /scroll to top/i });
      expect(button).toBeInTheDocument();
    });
  });

  it('should have proper accessibility attributes', async () => {
    window.scrollY = 600;
    render(<ScrollToTop threshold={500} />);
    
    fireEvent.scroll(window);
    
    await waitFor(() => {
      const button = screen.getByRole('button', { name: /scroll to top/i });
      expect(button).toHaveAttribute('aria-label', 'Scroll to top');
      expect(button).toHaveAttribute('title', 'Scroll to top');
    });
  });

  it('should scroll to top when clicked', async () => {
    const scrollToSpy = vi.spyOn(window, 'scrollTo');
    window.scrollY = 600;
    
    render(<ScrollToTop threshold={500} />);
    fireEvent.scroll(window);
    
    await waitFor(() => {
      const button = screen.getByRole('button', { name: /scroll to top/i });
      fireEvent.click(button);
    });
    
    expect(scrollToSpy).toHaveBeenCalledWith({ top: 0, behavior: 'smooth' });
  });

  it('should apply custom className', async () => {
    window.scrollY = 600;
    const customClass = 'custom-class';
    
    render(<ScrollToTop threshold={500} className={customClass} />);
    fireEvent.scroll(window);
    
    await waitFor(() => {
      const button = screen.getByRole('button', { name: /scroll to top/i });
      expect(button).toHaveClass(customClass);
    });
  });

  it('should use the tally violet fill with ink text', async () => {
    window.scrollY = 600;
    
    render(<ScrollToTop threshold={500} />);
    fireEvent.scroll(window);
    
    await waitFor(() => {
      const button = screen.getByRole('button', { name: /scroll to top/i });
      expect(button.className).toContain('bg-primary-400');
      expect(button.className).toContain('hover:bg-primary-300');
      expect(button.className).toContain('text-background');
    });
  });

  it('should clean up event listener on unmount', () => {
    const removeEventListenerSpy = vi.spyOn(window, 'removeEventListener');
    window.scrollY = 600;
    
    const { unmount } = render(<ScrollToTop threshold={500} />);
    unmount();
    
    expect(removeEventListenerSpy).toHaveBeenCalledWith('scroll', expect.any(Function));
  });
});
