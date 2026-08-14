import { describe, it, expect } from 'vitest';
import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { SkipLink } from './SkipLink';

describe('SkipLink', () => {
  it('renders with correct label', () => {
    render(
      <>
        <SkipLink targetId="main" label="Skip to content" />
        <main id="main">Main content</main>
      </>
    );
    
    const skipLink = screen.getByText('Skip to content');
    expect(skipLink).toBeInTheDocument();
  });

  it('has correct href attribute', () => {
    render(
      <>
        <SkipLink targetId="main-content" label="Skip to main" />
        <main id="main-content">Main content</main>
      </>
    );
    
    const skipLink = screen.getByText('Skip to main');
    expect(skipLink).toHaveAttribute('href', '#main-content');
  });

  it('focuses target element when clicked', async () => {
    const user = userEvent.setup();
    
    render(
      <>
        <SkipLink targetId="target" label="Skip" />
        <div id="target" tabIndex={-1}>Target</div>
      </>
    );
    
    const skipLink = screen.getByText('Skip');
    const target = screen.getByText('Target');
    
    await user.click(skipLink);
    
    expect(target).toHaveFocus();
  });

  it('focuses target element when activated with Enter', async () => {
    const user = userEvent.setup();

    render(
      <>
        <SkipLink targetId="target" label="Skip" />
        <div id="target" tabIndex={-1}>Target</div>
      </>
    );

    await user.tab();
    await user.keyboard('{Enter}');

    expect(screen.getByText('Target')).toHaveFocus();
  });

  it('is hidden by default but visible on focus', () => {
    render(
      <>
        <SkipLink targetId="main" label="Skip to content" />
        <main id="main">Main content</main>
      </>
    );
    
    const skipLink = screen.getByText('Skip to content');
    
    // Check that it has sr-only class which hides it visually
    expect(skipLink).toHaveClass('sr-only');
    // Check that it has focus:not-sr-only which makes it visible on focus
    expect(skipLink).toHaveClass('focus:not-sr-only');
  });
});
