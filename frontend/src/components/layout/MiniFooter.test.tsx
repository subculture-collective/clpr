import { describe, it, expect } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import { BrowserRouter } from 'react-router-dom';
import { MiniFooter } from './MiniFooter';

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
    
    // Check GitHub link
    const githubLink = screen.getByRole('link', { name: /github/i });
    expect(githubLink).toHaveAttribute('href', 'https://git.subcult.tv/subculture-collective/clpr');
    expect(githubLink).toHaveAttribute('target', '_blank');
    expect(githubLink).toHaveAttribute('rel', 'noopener noreferrer');
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

  it('should be positioned at bottom-left', () => {
    const { container } = render(<MiniFooter />, { wrapper: RouterWrapper });
    
    const wrapper = container.firstChild as HTMLElement;
    expect(wrapper.className).toContain('fixed');
    expect(wrapper.className).toContain('bottom-4');
    expect(wrapper.className).toContain('left-4');
  });

  it('should have proper z-index for layering', () => {
    const { container } = render(<MiniFooter />, { wrapper: RouterWrapper });
    
    const wrapper = container.firstChild as HTMLElement;
    expect(wrapper.className).toContain('z-40');
  });
});
