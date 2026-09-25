import { describe, it, expect, vi, afterEach } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import { useRef } from 'react';
import { BrowserRouter } from 'react-router-dom';
import { MiniFooter } from './MiniFooter';
import { useRegisterTwitchPlayer } from '@/hooks/useTwitchPlayerLayer';

function Player() {
  const ref = useRef<HTMLDivElement>(null);
  useRegisterTwitchPlayer(ref, true);
  return <div ref={ref} data-testid='player' />;
}

// The player spans the viewport's left side, where the fixed button sits.
function stubPlayerUnderButton(playerLeft: number) {
  vi.spyOn(HTMLElement.prototype, 'getBoundingClientRect').mockImplementation(function (this: HTMLElement) {
    const rect = this.dataset.testid === 'player'
      ? { left: playerLeft, top: 100, right: playerLeft + 800, bottom: 700, width: 800, height: 600 }
      : { left: 32, top: 599, right: 88, bottom: 655, width: 56, height: 56 };
    return rect as DOMRect;
  });
}

// Wrapper component for router context
const RouterWrapper = ({ children }: { children: React.ReactNode }) => (
  <BrowserRouter>{children}</BrowserRouter>
);

describe('MiniFooter', () => {
  it('should render collapsed by default', () => {
    render(<MiniFooter />, { wrapper: RouterWrapper });
    
    const expandButton = screen.getByRole('button', { name: /show footer links/i });
    expect(expandButton).toBeInTheDocument();
    
    // Should not show the expanded content
    expect(screen.queryByText(/quick links/i)).not.toBeInTheDocument();
  });

  it('should expand when clicked', () => {
    render(<MiniFooter />, { wrapper: RouterWrapper });
    
    const expandButton = screen.getByRole('button', { name: /show footer links/i });
    fireEvent.click(expandButton);
    
    // Should show the expanded content
    expect(screen.getByText(/quick links/i)).toBeInTheDocument();
  });

  it('should collapse when close button is clicked', () => {
    render(<MiniFooter />, { wrapper: RouterWrapper });
    
    // Expand first
    const expandButton = screen.getByRole('button', { name: /show footer links/i });
    fireEvent.click(expandButton);
    
    // Then collapse
    const closeButton = screen.getByRole('button', { name: /close footer links/i });
    fireEvent.click(closeButton);
    
    // Should not show the expanded content
    expect(screen.queryByText(/quick links/i)).not.toBeInTheDocument();
  });

  it('should display all footer sections when expanded', () => {
    render(<MiniFooter />, { wrapper: RouterWrapper });
    
    // Expand
    const expandButton = screen.getByRole('button', { name: /show footer links/i });
    fireEvent.click(expandButton);
    
    // Check for section headers using role
    expect(screen.getByRole('heading', { name: /^about$/i })).toBeInTheDocument();
    expect(screen.getByRole('heading', { name: /legal/i })).toBeInTheDocument();
    expect(screen.getByRole('heading', { name: /community/i })).toBeInTheDocument();
  });

  it('should have internal navigation links', () => {
    render(<MiniFooter />, { wrapper: RouterWrapper });
    
    // Expand
    const expandButton = screen.getByRole('button', { name: /show footer links/i });
    fireEvent.click(expandButton);
    
    // Check for internal links
    expect(screen.getByRole('link', { name: /about clpr/i })).toHaveAttribute('href', '/about');
    expect(screen.getByRole('link', { name: /privacy policy/i })).toHaveAttribute('href', '/privacy');
    expect(screen.getByRole('link', { name: /terms of service/i })).toHaveAttribute('href', '/terms');
  });

  it('should have external links with proper attributes', () => {
    render(<MiniFooter />, { wrapper: RouterWrapper });
    
    // Expand
    const expandButton = screen.getByRole('button', { name: /show footer links/i });
    fireEvent.click(expandButton);
    
    const patreonLink = screen.getByRole('link', { name: /patreon/i });
    expect(patreonLink).toHaveAttribute('href', 'https://patreon.com/subcult');
    expect(patreonLink).toHaveAttribute('target', '_blank');
    expect(patreonLink).toHaveAttribute('rel', 'noopener noreferrer');
  });

  it('should have proper accessibility attributes on buttons', () => {
    render(<MiniFooter />, { wrapper: RouterWrapper });
    
    const expandButton = screen.getByRole('button', { name: /show footer links/i });
    expect(expandButton).toHaveAttribute('aria-label', 'Show footer links');
    expect(expandButton).toHaveAttribute('title', 'Footer links');
  });

  it('should collapse when internal link is clicked', () => {
    render(<MiniFooter />, { wrapper: RouterWrapper });
    
    // Expand
    const expandButton = screen.getByRole('button', { name: /show footer links/i });
    fireEvent.click(expandButton);
    
    // Click an internal link
    const aboutLink = screen.getByRole('link', { name: /about clpr/i });
    fireEvent.click(aboutLink);
    
    // Should collapse (Quick Links text should disappear)
    expect(screen.queryByText(/quick links/i)).not.toBeInTheDocument();
  });

  describe('around Twitch players', () => {
    afterEach(() => vi.restoreAllMocks());

    it('steps aside while it would cover a player', () => {
      stubPlayerUnderButton(0);
      render(<><Player /><MiniFooter /></>, { wrapper: RouterWrapper });
      const button = screen.getByRole('button', { name: /show footer links/i, hidden: true });
      expect(button.parentElement).toHaveClass('invisible');
    });

    it('stays visible when the player is elsewhere', () => {
      stubPlayerUnderButton(400);
      render(<><Player /><MiniFooter /></>, { wrapper: RouterWrapper });
      const button = screen.getByRole('button', { name: /show footer links/i });
      expect(button.parentElement).not.toHaveClass('invisible');
    });
  });
});
