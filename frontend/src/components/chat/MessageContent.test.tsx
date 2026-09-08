import { describe, it, expect } from 'vitest';
import { render, screen } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';
import { MessageContent } from './MessageContent';

// Helper function to render with router
const renderWithRouter = (ui: React.ReactElement) => {
  return render(<MemoryRouter>{ui}</MemoryRouter>);
};

describe('MessageContent', () => {
  it('renders plain text content', () => {
    renderWithRouter(<MessageContent content="Hello world" />);
    expect(screen.getByText('Hello world')).toBeInTheDocument();
  });

  it('preserves mentions in message text', () => {
    renderWithRouter(<MessageContent content="Hello @user123" />);
    const mention = screen.getByText('@user123');
    expect(mention).toBeInTheDocument();
  });

  it('renders URLs as links', () => {
    renderWithRouter(<MessageContent content="Check out https://example.com" />);
    const link = screen.getByRole('link');
    expect(link).toHaveAttribute('href', 'https://example.com');
    expect(link).toHaveAttribute('target', '_blank');
    expect(link).toHaveAttribute('rel', 'noopener noreferrer');
  });

  it('renders inline code blocks', () => {
    renderWithRouter(<MessageContent content="Use `console.log()` for debugging" />);
    const code = screen.getByText('console.log()');
    expect(code.tagName).toBe('CODE');
  });

  it('renders multi-line code blocks', () => {
    renderWithRouter(<MessageContent content="```const x = 1;```" />);
    const code = screen.getByText('const x = 1;');
    expect(code.tagName).toBe('CODE');
    expect(code.parentElement?.tagName).toBe('PRE');
  });

  it('handles mixed content with mentions and links', () => {
    renderWithRouter(<MessageContent content="Hey @user check https://example.com" />);
    expect(screen.getByText('@user')).toBeInTheDocument();
    expect(screen.getByRole('link')).toBeInTheDocument();
  });
});
